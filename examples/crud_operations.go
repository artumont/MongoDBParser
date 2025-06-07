package main

import (
	"context"
	"log"
	"time"

	mongoparser "github.com/artumont/MongoDBParser"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func crudExample() {
	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer client.Disconnect(ctx)

	// Get database
	db := client.Database("crud_demo")

	// Create parser
	parser := mongoparser.NewParser()

	// CRUD operations demonstration
	jsScript := `
		// METADATA:
		// {
		//   "description": "CRUD operations demonstration with data manipulation",
		//   "version": "1.0.0",
		//   "author": "MongoDB Parser",
		//   "dependencies": []
		// }

		// Create a simple users collection
		db.createCollection("users");

		// Insert single user
		db.users.insertOne({
			name: "Alice Johnson",
			email: "alice.johnson@example.com",
			age: 28,
			department: "Engineering",
			salary: 75000,
			skills: ["JavaScript", "Python", "Go"],
			active: true,
			created_at: new Date()
		});

		// Insert multiple users
		db.users.insertMany([
			{
				name: "Bob Smith",
				email: "bob.smith@example.com",
				age: 35,
				department: "Marketing",
				salary: 65000,
				skills: ["Analytics", "SEO", "Content Marketing"],
				active: true,
				created_at: new Date()
			},
			{
				name: "Carol Wilson",
				email: "carol.wilson@example.com",
				age: 42,
				department: "HR",
				salary: 70000,
				skills: ["Recruitment", "Training", "Management"],
				active: false,
				created_at: new Date()
			},
			{
				name: "David Brown",
				email: "david.brown@example.com",
				age: 29,
				department: "Engineering",
				salary: 80000,
				skills: ["Java", "Kubernetes", "Docker"],
				active: true,
				created_at: new Date()
			}
		]);

		// Update operations
		db.users.updateOne(
			{ email: "alice.johnson@example.com" },
			{
				$set: {
					salary: 78000,
					updated_at: new Date()
				},
				$push: {
					skills: "React"
				}
			}
		);

		// Update multiple users
		db.users.updateMany(
			{ department: "Engineering" },
			{
				$inc: { salary: 2000 },
				$set: { updated_at: new Date() }
			}
		);

		// Update with upsert behavior (would create if not exists)
		db.users.updateOne(
			{ email: "eve.davis@example.com" },
			{
				$set: {
					name: "Eve Davis",
					email: "eve.davis@example.com",
					age: 31,
					department: "Design",
					salary: 72000,
					skills: ["UI/UX", "Figma", "Prototyping"],
					active: true,
					created_at: new Date()
				}
			}
		);

		// Delete operations
		db.users.deleteOne({ active: false });

		// Delete multiple inactive users
		db.users.deleteMany({ active: false });
	`

	// Execute the script
	result := parser.ExecuteScript(ctx, db, jsScript)
	if !result.Success {
		log.Fatal("❌ CRUD operations failed:", result.Error)
	}

	log.Println("✅ CRUD operations completed successfully!")
	log.Printf("📊 Results: %+v", result.Output)
}

// Uncomment the line below to run this example
// func main() { crudExample() }
