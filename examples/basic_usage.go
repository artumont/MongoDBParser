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
	db := client.Database("example_app")

	// Create parser
	parser := mongoparser.NewParser()

	// Basic MongoDB JavaScript operations
	jsScript := `
		// Create a simple collection
		db.createCollection("users");

		// Create indexes
		db.users.createIndex({ email: 1 }, { unique: true });
		db.users.createIndex({ created_at: -1 });
		db.users.createIndex({ "profile.department": 1, status: 1 });

		// Insert a single document
		db.users.insertOne({
			name: "John Doe",
			email: "john.doe@example.com",
			age: 30,
			status: "active",
			profile: {
				department: "Engineering",
				role: "Senior Developer"
			},
			created_at: new Date()
		});

		// Insert multiple documents
		db.users.insertMany([
			{
				name: "Jane Smith",
				email: "jane.smith@example.com",
				age: 28,
				status: "active",
				profile: {
					department: "Marketing",
					role: "Marketing Manager"
				}
			},
			{
				name: "Bob Johnson",
				email: "bob.johnson@example.com",
				age: 35,
				status: "inactive",
				profile: {
					department: "Sales",
					role: "Sales Representative"
				}
			}
		]);
	`

	// Execute the script
	result := parser.ExecuteScript(ctx, db, jsScript)
	if !result.Success {
		log.Fatal("Script execution failed:", result.Error)
	}

	log.Println("✅ Basic setup completed successfully!")
	log.Printf("Results: %+v", result.Output)
}
