package mongoparser

import (
	"context"
	"fmt"
	"log"
	"strings"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Executes a parsed MongoDB operation
func (p *Parser) executeMongoOperation(ctx context.Context, db *mongo.Database, op MongoOperation) (interface{}, error) {
	switch op.Type {
	case "createCollection":
		return p.executeCreateCollection(ctx, db, op)
	case "createIndex":
		return p.executeCreateIndex(ctx, db, op)
	case "insert":
		return p.executeInsert(ctx, db, op)
	case "update":
		return p.executeUpdate(ctx, db, op)
	case "delete":
		return p.executeDelete(ctx, db, op)
	default:
		return nil, fmt.Errorf("unsupported operation type: %s", op.Type)
	}
}

// Executes createCollection operation
func (p *Parser) executeCreateCollection(ctx context.Context, db *mongo.Database, op MongoOperation) (interface{}, error) {
	opts := options.CreateCollection()
	if op.Validator != nil {
		opts.SetValidator(op.Validator)
	}

	err := db.CreateCollection(ctx, op.Collection, opts)
	if err != nil {
		// Check if collection already exists
		if mongo.IsDuplicateKeyError(err) || strings.Contains(err.Error(), "already exists") {
			log.Printf("Collection %s already exists, skipping", op.Collection)
			return "Collection already exists", nil
		}
		return nil, err
	}

	return fmt.Sprintf("Collection %s created successfully", op.Collection), nil
}

// Executes createIndex operation
func (p *Parser) executeCreateIndex(ctx context.Context, db *mongo.Database, op MongoOperation) (interface{}, error) {
	collection := db.Collection(op.Collection)

	indexModel := mongo.IndexModel{
		Keys: op.IndexSpec,
	}
	if op.IndexOptions != nil {
		indexModel.Options = op.IndexOptions
	}

	result, err := collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		// Check if index already exists
		if strings.Contains(err.Error(), "already exists") {
			log.Printf("Index already exists on collection %s, skipping", op.Collection)
			return "Index already exists", nil
		}
		return nil, err
	}

	return fmt.Sprintf("Index created on %s: %s", op.Collection, result), nil
}

// Executes insert operations
func (p *Parser) executeInsert(ctx context.Context, db *mongo.Database, op MongoOperation) (interface{}, error) {
	collection := db.Collection(op.Collection)

	if len(op.Arguments) == 0 {
		return nil, fmt.Errorf("no document to insert")
	}

	switch op.Operation {
	case "insertOne":
		result, err := collection.InsertOne(ctx, op.Arguments[0])
		if err != nil {
			return nil, err
		}
		return result.InsertedID, nil
	case "insertMany":
		var docs []interface{}
		for _, doc := range op.Arguments {
			docs = append(docs, doc)
		}
		result, err := collection.InsertMany(ctx, docs)
		if err != nil {
			return nil, err
		}
		return result.InsertedIDs, nil
	default:
		return nil, fmt.Errorf("unsupported insert operation: %s", op.Operation)
	}
}

// Executes update operations
func (p *Parser) executeUpdate(ctx context.Context, db *mongo.Database, op MongoOperation) (interface{}, error) {
	if len(op.Arguments) < 2 {
		return nil, fmt.Errorf("update operation requires filter and update documents")
	}

	collection := db.Collection(op.Collection)
	filter := op.Arguments[0]
	update := op.Arguments[1]

	switch op.Operation {
	case "updateOne":
		result, err := collection.UpdateOne(ctx, filter, update)
		if err != nil {
			return nil, err
		}
		return result.ModifiedCount, nil
	case "updateMany":
		result, err := collection.UpdateMany(ctx, filter, update)
		if err != nil {
			return nil, err
		}
		return result.ModifiedCount, nil
	default:
		return nil, fmt.Errorf("unsupported update operation: %s", op.Operation)
	}
}

// Executes delete operations
func (p *Parser) executeDelete(ctx context.Context, db *mongo.Database, op MongoOperation) (interface{}, error) {
	if len(op.Arguments) == 0 {
		return nil, fmt.Errorf("delete operation requires filter document")
	}

	collection := db.Collection(op.Collection)
	filter := op.Arguments[0]

	switch op.Operation {
	case "deleteOne":
		result, err := collection.DeleteOne(ctx, filter)
		if err != nil {
			return nil, err
		}
		return result.DeletedCount, nil
	case "deleteMany":
		result, err := collection.DeleteMany(ctx, filter)
		if err != nil {
			return nil, err
		}
		return result.DeletedCount, nil
	default:
		return nil, fmt.Errorf("unsupported delete operation: %s", op.Operation)
	}
}
