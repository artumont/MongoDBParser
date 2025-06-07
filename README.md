# 🍃 MongoDB JavaScript Parser

A powerful Go library for parsing and executing MongoDB JavaScript operations. Transform MongoDB shell scripts into native Go operations with automatic syntax conversion and type safety. Originally made for [project kyber](https://github.com/ZentriLinkInc) but now available as a standalone library (important: this library might not work out of the box as it was originally designed for a specific project and may require adjustments to fit your use case).

## ✨ Features

- 🔄 Parse complex multi-line MongoDB JavaScript operations
- 🎯 Automatic JavaScript-to-JSON conversion with unquoted key support
- 🧹 Smart trailing comma removal for JSON compliance
- 🔢 Intelligent numeric type conversion (string numbers → proper types)
- 📊 Support for collection creation with complex validators
- 🗂️ Index creation with proper field ordering using bson.D
- 📝 Script metadata parsing from comments
- 🔍 Migration tracking with MongoDB schema validation
- 🚀 High-performance parsing with error recovery
- 💪 Type-safe operations with flexible BSON support

## 🛠️ Technologies

- Go 1.21+
- MongoDB Go Driver
- BSON document handling
- JSON parsing with JavaScript syntax support

## 📋 Prerequisites

- Go 1.21 or later
- MongoDB Go Driver (`go.mongodb.org/mongo-driver`)

## 📦 Installation

```bash
go get github.com/artumont/MongoDBParser
```

## 🚀 Usage

### Basic Usage

```go
package main

import (
    "context"
    "log"
    
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "github.com/artumont/MongoDBParser"
)

func main() {
    // Connect to MongoDB
    client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
    if err != nil {
        log.Fatal(err)
    }
    defer client.Disconnect(context.TODO())
    
    db := client.Database("myapp")
    
    // Create parser
    parser := mongoparser.NewParser()
    
    // Execute MongoDB JavaScript
    jsScript := `
    db.createCollection("users", {
        validator: {
            $jsonSchema: {
                bsonType: "object",
                required: ["name", "email"],
                properties: {
                    name: { bsonType: "string" },
                    email: { bsonType: "string" }
                }
            }
        }
    });
    
    db.users.createIndex({ email: 1 });
    db.users.createIndex({ created_at: -1 });
    `
    
    result := parser.ExecuteScript(context.TODO(), db, jsScript)
    if !result.Success {
        log.Fatal("Script execution failed:", result.Error)
    }
    
    log.Println("Script executed successfully!")
}
```

### Advanced Usage with Metadata

```go
// Script with metadata
jsWithMetadata := `
// METADATA:
// {
//   "description": "User collection setup",
//   "version": "1.0.0",
//   "author": "artumont",
//   "dependencies": ["base_setup"]
// }

db.createCollection("users", {
    validator: {
        $jsonSchema: {
            bsonType: "object",
            required: ["name", "email", "created_at"],
            properties: {
                name: {
                    bsonType: "string",
                    description: "User's full name"
                },
                email: {
                    bsonType: "string",
                    pattern: "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$"
                },
                created_at: {
                    bsonType: "date",
                    description: "Account creation timestamp"
                }
            }
        }
    }
});

// Create indexes with proper numeric types
db.users.createIndex({ email: 1 }, { unique: true });
db.users.createIndex({ created_at: -1 });
db.users.createIndex({ "profile.department": 1, status: 1 });
`

parser := mongoparser.NewParser()

// Parse metadata
metadata := parser.ParseMetadata(jsWithMetadata)
if metadata != nil {
    log.Printf("Script: %s v%s by %s", metadata.Description, metadata.Version, metadata.Author)
}

// Execute script
result := parser.ExecuteScript(ctx, db, jsWithMetadata)
```

### Supported Operations

#### Collection Operations

```javascript
// Collection creation with validators
db.createCollection("products", {
    validator: {
        $jsonSchema: {
            bsonType: "object",
            required: ["name", "price"],
            properties: {
                name: { bsonType: "string" },
                price: { bsonType: "number", minimum: 0 },
                tags: { 
                    bsonType: "array",
                    items: { bsonType: "string" }
                }
            }
        }
    }
});
```

#### Index Operations

```javascript
// Simple indexes
db.products.createIndex({ name: 1 });
db.products.createIndex({ price: -1 });

// Compound indexes  
db.products.createIndex({ category: 1, price: -1 });

// Indexes with options
db.products.createIndex(
    { name: "text" }, 
    { name: "product_text_search" }
);
```

#### Document Operations

```javascript
// Insert operations
db.products.insertOne({
    name: "Laptop",
    price: 999.99,
    category: "Electronics"
});

// Update operations
db.products.updateMany(
    { category: "Electronics" },
    { $set: { updated_at: new Date() } }
);

// Delete operations
db.products.deleteOne({ _id: ObjectId("...") });
```

## 🎯 Key Features

### JavaScript Syntax Support

The parser automatically handles common JavaScript syntax issues:

```javascript
// ✅ Unquoted keys (converted automatically)
db.users.createIndex({ user_id: 1, status: -1 });

// ✅ Trailing commas (removed automatically)  
db.createCollection("test", {
    validator: {
        $jsonSchema: {
            bsonType: "object",
            properties: {
                name: { bsonType: "string" },
                age: { bsonType: "number" }, // <- trailing comma handled
            },
        },
    },
});

// ✅ Single quotes (converted to double quotes)
db.users.createIndex({ 'email': 1 });
```

### Type Safety

```go
// Automatic type conversion for index values
"1"  → int(1)      // Ascending index
"-1" → int(-1)     // Descending index  
"2dsphere" → "2dsphere"  // Geospatial index (kept as string)
```

### Error Recovery

```go
result := parser.ExecuteScript(ctx, db, script)
if !result.Success {
    log.Printf("Execution failed: %v", result.Error)
    // Continue with other operations
}
```

## 📁 Package Structure

```javascript
mongoparser/
├── parser.go      # Main parser logic and operation handlers
├── types.go       # Type definitions and structures  
├── utils.go       # Utility functions for JavaScript/JSON conversion
├── executor.go    # MongoDB operation execution logic
└── README.md      # This file
```

## 🧪 Testing

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test -run TestParseCreateCollection ./...
```

## 🔧 Configuration

### Parser Options

```go
parser := mongoparser.NewParser()

// Parse metadata from script comments
metadata := parser.ParseMetadata(scriptContent)

// Execute with context and timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result := parser.ExecuteScript(ctx, db, scriptContent)
```

### Supported MongoDB Operations

| Operation | Support | Notes |
|-----------|---------|-------|
| `createCollection` | ✅ | With validator support |
| `createIndex` | ✅ | All index types, options |
| `insertOne` | ✅ | Single document insert |
| `insertMany` | ✅ | Batch document insert |
| `updateOne` | ✅ | Single document update |
| `updateMany` | ✅ | Multiple document update |
| `deleteOne` | ✅ | Single document delete |
| `deleteMany` | ✅ | Multiple document delete |

## 🐛 Error Handling

The parser provides detailed error information:

```go
result := parser.ExecuteScript(ctx, db, script)
if !result.Success {
    switch {
    case strings.Contains(result.Error.Error(), "parse"):
        log.Println("JavaScript parsing error:", result.Error)
    case strings.Contains(result.Error.Error(), "execute"):
        log.Println("MongoDB execution error:", result.Error)
    default:
        log.Println("Unknown error:", result.Error)
    }
}
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Commit your changes: `git commit -m 'Add amazing feature'`
4. Push to the branch: `git push origin feature/amazing-feature`
5. Open a pull request

## 📝 Examples

The `examples/` directory contains comprehensive examples demonstrating various use cases for the MongoDB JavaScript Parser. Each example shows complete, working code that you can run in your own projects.

### Available Examples

Most of these examples where generated with the help of AI and may not work out of the box, but they should give you a good starting point for using the parser effectively:

- **[basic_usage.go](examples/basic_usage.go)** - Fundamental usage with basic operations
- **[complex_schema.go](examples/complex_schema.go)** - Advanced schema validation and e-commerce models  
- **[crud_operations.go](examples/crud_operations.go)** - Complete CRUD operations demonstration
- **[error_handling.go](examples/error_handling.go)** - Error handling and edge cases

See the [examples README](examples/README.md) for detailed documentation and usage instructions.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
