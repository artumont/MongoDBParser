package main

import (
	"context"
	"log"
	"time"

	mongoparser "github.com/artumont/MongoDBParser"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func errorHandlingExample() {
	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer client.Disconnect(ctx)

	// Get database
	db := client.Database("error_demo")

	// Create parser
	parser := mongoparser.NewParser()

	log.Println("🧪 Testing Error Handling and Edge Cases")

	// Test 1: Valid script
	log.Println("\n1️⃣  Testing valid script...")
	validScript := `
		db.createCollection("test_collection");
		db.test_collection.createIndex({ name: 1 });
	`

	result := parser.ExecuteScript(ctx, db, validScript)
	if result.Success {
		log.Println("✅ Valid script executed successfully")
	} else {
		log.Printf("❌ Unexpected failure: %v", result.Error)
	}

	// Test 2: Script with JavaScript syntax quirks (should be handled)
	log.Println("\n2️⃣  Testing JavaScript syntax quirks...")
	jsQuirksScript := `
		// Unquoted keys, single quotes, trailing commas
		db.createCollection('products', {
			validator: {
				$jsonSchema: {
					bsonType: "object",
					required: ["name", "price"],
					properties: {
						name: { bsonType: "string" },
						price: { bsonType: "number" }, // trailing comma
					},
				},
			},
		});

		// Mixed quote styles and unquoted keys
		db.products.createIndex({ name: 1, 'category': -1, "status": 1 });
	`

	result = parser.ExecuteScript(ctx, db, jsQuirksScript)
	if result.Success {
		log.Println("✅ JavaScript syntax quirks handled successfully")
	} else {
		log.Printf("❌ Failed to handle JS syntax: %v", result.Error)
	}

	// Test 3: Empty script
	log.Println("\n3️⃣  Testing empty script...")
	emptyResult := parser.ExecuteScript(ctx, db, "")
	if emptyResult.Success {
		log.Println("✅ Empty script handled gracefully")
	} else {
		log.Printf("❌ Empty script failed: %v", emptyResult.Error)
	}

	// Test 4: Comment-only script
	log.Println("\n4️⃣  Testing comment-only script...")
	commentScript := `
		// This is just a comment
		// Another comment
		/* Multi-line comment */
	`
	commentResult := parser.ExecuteScript(ctx, db, commentScript)
	if commentResult.Success {
		log.Println("✅ Comment-only script handled gracefully")
	} else {
		log.Printf("❌ Comment-only script failed: %v", commentResult.Error)
	}

	// Test 5: Invalid JSON structure (should fail gracefully)
	log.Println("\n5️⃣  Testing invalid JSON structure...")
	invalidJsonScript := `
		db.createCollection("invalid", {
			validator: {
				$jsonSchema: {
					bsonType: "object"
					// Missing comma, invalid structure
					required: ["name"]
				}
			}
		});
	`

	result = parser.ExecuteScript(ctx, db, invalidJsonScript)
	if !result.Success {
		log.Printf("✅ Invalid JSON handled gracefully: %v", result.Error)
	} else {
		log.Println("❌ Invalid JSON should have failed")
	}

	// Test 6: Metadata parsing
	log.Println("\n6️⃣  Testing metadata parsing...")
	metadataScript := `
		// METADATA:
		// {
		//   "description": "Test script with metadata",
		//   "version": "1.0.0",
		//   "author": "Test Author",
		//   "dependencies": ["dep1", "dep2"]
		// }

		db.createCollection("metadata_test");
	`

	metadata := parser.ParseMetadata(metadataScript)
	if metadata != nil {
		log.Printf("✅ Metadata parsed successfully: %s v%s by %s",
			metadata.Description, metadata.Version, metadata.Author)
		log.Printf("   Dependencies: %v", metadata.Dependencies)
	} else {
		log.Println("❌ Failed to parse metadata")
	}

	// Test 7: Complex nested operations
	log.Println("\n7️⃣  Testing complex nested operations...")
	complexScript := `
		db.createCollection("complex", {
			validator: {
				$jsonSchema: {
					bsonType: "object",
					required: ["user", "profile"],
					properties: {
						user: {
							bsonType: "object",
							required: ["email", "name"],
							properties: {
								email: { 
									bsonType: "string",
									pattern: "^[\\w-\\.]+@([\\w-]+\\.)+[\\w-]{2,4}$"
								},
								name: { 
									bsonType: "string",
									minLength: 2,
									maxLength: 100
								},
								preferences: {
									bsonType: "object",
									properties: {
										theme: { enum: ["light", "dark"] },
										notifications: { bsonType: "bool" },
										languages: {
											bsonType: "array",
											items: { bsonType: "string" },
											maxItems: 5
										}
									}
								}
							}
						},
						profile: {
							bsonType: "object",
							properties: {
								bio: { bsonType: "string", maxLength: 500 },
								avatar_url: { bsonType: "string" },
								social_links: {
									bsonType: "array",
									items: {
										bsonType: "object",
										properties: {
											platform: { bsonType: "string" },
											url: { bsonType: "string" }
										}
									}
								}
							}
						}
					}
				}
			}
		});

		db.complex.createIndex({ "user.email": 1 }, { unique: true });
		db.complex.createIndex({ "user.name": "text", "profile.bio": "text" });
	`

	result = parser.ExecuteScript(ctx, db, complexScript)
	if result.Success {
		log.Println("✅ Complex nested operations executed successfully")
	} else {
		log.Printf("❌ Complex operations failed: %v", result.Error)
	}

	log.Println("\n🏁 Error handling tests completed!")
}

// Uncomment the line below to run this example
// func main() { errorHandlingExample() }
