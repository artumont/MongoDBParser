package main

import (
	"context"
	"log"
	"time"

	mongoparser "github.com/artumont/MongoDBParser"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer client.Disconnect(ctx)

	// Get database
	db := client.Database("ecommerce_app")

	// Create parser
	parser := mongoparser.NewParser()

	// Complex e-commerce schema with advanced validators
	jsScript := `
		// METADATA:
		// {
		//   "description": "E-commerce product and order collections with complex validators",
		//   "version": "2.1.0",
		//   "author": "MongoDB Parser Example",
		//   "dependencies": ["base_setup"]
		// }

		// Create products collection with comprehensive validation
		db.createCollection("products", {
			validator: {
				$jsonSchema: {
					bsonType: "object",
					required: ["name", "price", "category", "sku", "status"],
					properties: {
						name: {
							bsonType: "string",
							minLength: 3,
							maxLength: 200,
							description: "Product name must be 3-200 characters"
						},
						description: {
							bsonType: "string",
							maxLength: 2000,
							description: "Product description"
						},
						price: {
							bsonType: "number",
							minimum: 0,
							exclusiveMinimum: true,
							description: "Price must be positive"
						},
						category: {
							enum: ["Electronics", "Clothing", "Books", "Home", "Sports", "Beauty"],
							description: "Must be one of the predefined categories"
						},
						sku: {
							bsonType: "string",
							pattern: "^[A-Z]{3}-[0-9]{6}$",
							description: "SKU format: XXX-123456"
						},
						status: {
							enum: ["active", "inactive", "discontinued"],
							description: "Product status"
						},
						dimensions: {
							bsonType: "object",
							properties: {
								length: { bsonType: "number", minimum: 0 },
								width: { bsonType: "number", minimum: 0 },
								height: { bsonType: "number", minimum: 0 },
								weight: { bsonType: "number", minimum: 0 }
							}
						},
						tags: {
							bsonType: "array",
							items: {
								bsonType: "string",
								minLength: 2,
								maxLength: 30
							},
							maxItems: 10
						},
						inventory: {
							bsonType: "object",
							required: ["quantity", "warehouse"],
							properties: {
								quantity: {
									bsonType: "int",
									minimum: 0
								},
								warehouse: {
									bsonType: "string",
									enum: ["US-EAST", "US-WEST", "EU-CENTRAL", "ASIA-PACIFIC"]
								},
								reorderPoint: {
									bsonType: "int",
									minimum: 0
								}
							}
						},
						created_at: {
							bsonType: "date",
							description: "Product creation timestamp"
						},
						updated_at: {
							bsonType: "date",
							description: "Last update timestamp"
						}
					}
				}
			}
		});

		// Create comprehensive indexes for products
		db.products.createIndex({ sku: 1 }, { unique: true, name: "unique_sku" });
		db.products.createIndex({ category: 1, status: 1 }, { name: "category_status" });
		db.products.createIndex({ price: 1 }, { name: "price_index" });
		db.products.createIndex({ name: "text", description: "text", tags: "text" }, { 
			name: "product_search", 
			weights: { name: 10, description: 5, tags: 1 }
		});
		db.products.createIndex({ "inventory.warehouse": 1, "inventory.quantity": 1 }, { name: "inventory_lookup" });
		db.products.createIndex({ created_at: -1 }, { name: "recent_products" });

		// Create orders collection with complex validation
		db.createCollection("orders", {
			validator: {
				$jsonSchema: {
					bsonType: "object",
					required: ["customer_id", "items", "status", "total_amount", "created_at"],
					properties: {
						customer_id: {
							bsonType: "objectId",
							description: "Reference to customer"
						},
						order_number: {
							bsonType: "string",
							pattern: "^ORD-[0-9]{8}$",
							description: "Order number format: ORD-12345678"
						},
						items: {
							bsonType: "array",
							minItems: 1,
							maxItems: 50,
							items: {
								bsonType: "object",
								required: ["product_id", "quantity", "price"],
								properties: {
									product_id: { bsonType: "objectId" },
									quantity: { bsonType: "int", minimum: 1 },
									price: { bsonType: "number", minimum: 0 },
									discount: { bsonType: "number", minimum: 0, maximum: 1 }
								}
							}
						},
						status: {
							enum: ["pending", "confirmed", "processing", "shipped", "delivered", "cancelled"],
							description: "Order status"
						},
						total_amount: {
							bsonType: "number",
							minimum: 0,
							description: "Total order amount"
						},
						shipping_address: {
							bsonType: "object",
							required: ["street", "city", "country", "postal_code"],
							properties: {
								street: { bsonType: "string", minLength: 5 },
								city: { bsonType: "string", minLength: 2 },
								state: { bsonType: "string" },
								country: { bsonType: "string", minLength: 2 },
								postal_code: { bsonType: "string", minLength: 3 }
							}
						},
						payment_method: {
							enum: ["credit_card", "debit_card", "paypal", "bank_transfer", "cash_on_delivery"]
						},
						created_at: { bsonType: "date" },
						updated_at: { bsonType: "date" }
					}
				}
			}
		});

		// Create indexes for orders
		db.orders.createIndex({ customer_id: 1, created_at: -1 }, { name: "customer_orders" });
		db.orders.createIndex({ status: 1, created_at: -1 }, { name: "status_timeline" });
		db.orders.createIndex({ order_number: 1 }, { unique: true, name: "unique_order_number" });
		db.orders.createIndex({ "items.product_id": 1 }, { name: "product_orders" });
		db.orders.createIndex({ total_amount: -1 }, { name: "high_value_orders" });
	`

	// Parse metadata first
	metadata := parser.ParseMetadata(jsScript)
	if metadata != nil {
		log.Printf("📋 Setting up: %s v%s", metadata.Description, metadata.Version)
		log.Printf("👤 Author: %s", metadata.Author)
	}

	// Execute the script
	result := parser.ExecuteScript(ctx, db, jsScript)
	if !result.Success {
		log.Fatal("❌ Script execution failed:", result.Error)
	}

	log.Println("✅ Complex e-commerce schema setup completed successfully!")
	log.Printf("📊 Results: %+v", result.Output)

	// Verify collections were created
	collections, err := db.ListCollectionNames(ctx, nil)
	if err != nil {
		log.Printf("Warning: Could not list collections: %v", err)
	} else {
		log.Printf("📁 Collections created: %v", collections)
	}
}
