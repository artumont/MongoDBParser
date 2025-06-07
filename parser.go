package mongoparser

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Handles parsing and execution of MongoDB JavaScript operations
type Parser struct{}

// Creates a new MongoDB JavaScript parser
func NewParser() *Parser {
	return &Parser{}
}

// Extracts metadata from script comments
func (p *Parser) ParseMetadata(content string) *ScriptMetadata {
	lines := strings.Split(content, "\n")
	var metadataLines []string

	// Look for JSON metadata in comments at the start of the file
	inMetadata := false
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "// METADATA:") {
			inMetadata = true
			continue
		}

		if inMetadata {
			if strings.HasPrefix(line, "//") {
				// Remove comment prefix and add to metadata
				metadataLine := strings.TrimPrefix(line, "//")
				metadataLine = strings.TrimSpace(metadataLine)
				if metadataLine != "" {
					metadataLines = append(metadataLines, metadataLine)
				}
			} else {
				// End of metadata section
				break
			}
		}
	}

	if len(metadataLines) == 0 {
		return nil
	}

	// Try to parse as JSON
	jsonStr := strings.Join(metadataLines, "")
	var metadata ScriptMetadata
	if err := json.Unmarshal([]byte(jsonStr), &metadata); err != nil {
		log.Printf("Warning: failed to parse script metadata: %v", err)
		return nil
	}

	return &metadata
}

// Executes JavaScript content by parsing and converting to Go MongoDB operations
func (p *Parser) ExecuteScript(ctx context.Context, db *mongo.Database, jsContent string) ScriptResult {
	if len(strings.TrimSpace(jsContent)) == 0 {
		return ScriptResult{
			Success: true,
			Output:  "Script is empty, skipped",
		}
	}

	operations, err := p.parseJavaScriptOperations(jsContent)
	if err != nil {
		return ScriptResult{
			Success: false,
			Error:   fmt.Errorf("failed to parse JavaScript operations: %w", err),
		}
	}

	var results []interface{}
	for _, op := range operations {
		result, err := p.executeMongoOperation(ctx, db, op)
		if err != nil {
			return ScriptResult{
				Success: false,
				Error:   fmt.Errorf("failed to execute operation %s on %s: %w", op.Operation, op.Collection, err),
			}
		}
		results = append(results, result)
	}

	return ScriptResult{
		Success: true,
		Output:  results,
	}
}

// Parses JavaScript MongoDB operations and converts them to Go operations
func (p *Parser) parseJavaScriptOperations(jsContent string) ([]MongoOperation, error) {
	var operations []MongoOperation

	// First, split the content into complete statements that may span multiple lines
	statements := p.splitIntoStatements(jsContent)

	for _, statement := range statements {
		statement = strings.TrimSpace(statement)
		if statement == "" || strings.HasPrefix(statement, "//") {
			continue
		}

		// Parse db.collection.operation() patterns
		if strings.HasPrefix(statement, "db.") && strings.Contains(statement, "(") {
			op, err := p.parseMongoStatement(statement)
			if err != nil {
				log.Printf("Warning: failed to parse statement '%s': %v", statement, err)
				continue
			}
			if op != nil {
				operations = append(operations, *op)
			}
		}
	}

	return operations, nil
}

// Parses createIndex operation
func (p *Parser) parseCreateIndex(collection, argsString string) (*MongoOperation, error) {
	op := &MongoOperation{
		Type:       "createIndex",
		Collection: collection,
		Operation:  "createIndex",
	}

	// Parse index specification and options using splitArguments
	args := p.splitArguments(argsString)
	if len(args) > 0 {
		indexSpecStr := strings.TrimSpace(args[0])

		// Convert to bson.D for proper index specification
		var indexSpecMap map[string]interface{}
		if err := p.parseJSONLikeString(indexSpecStr, &indexSpecMap); err != nil {
			return nil, fmt.Errorf("failed to parse index specification: %w", err)
		}

		// Convert map to bson.D to preserve field order for indexes
		var indexSpec bson.D
		for key, value := range indexSpecMap {
			// Ensure numeric values are properly typed
			if numValue, err := p.convertToNumber(value); err == nil {
				indexSpec = append(indexSpec, bson.E{Key: key, Value: numValue})
			} else {
				indexSpec = append(indexSpec, bson.E{Key: key, Value: value})
			}
		}
		op.IndexSpec = indexSpec

		// Parse index options if provided
		if len(args) > 1 {
			var indexOptions map[string]interface{}
			if err := p.parseJSONLikeString(strings.TrimSpace(args[1]), &indexOptions); err != nil {
				log.Printf("Warning: failed to parse index options: %v", err)
			} else {
				opts := options.Index()
				if unique, ok := indexOptions["unique"]; ok {
					if uniqueBool, ok := unique.(bool); ok {
						opts.SetUnique(uniqueBool)
					}
				}
				if name, ok := indexOptions["name"]; ok {
					if nameStr, ok := name.(string); ok {
						opts.SetName(nameStr)
					}
				}
				op.IndexOptions = opts
			}
		}
	}

	return op, nil
}

// Attempts to convert a value to the appropriate numeric type
func (p *Parser) convertToNumber(value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case string:
		// Try to parse as int first
		if strings.Contains(v, ".") {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return f, nil
			}
		} else {
			if i, err := strconv.Atoi(v); err == nil {
				return i, nil
			}
		}
		return nil, fmt.Errorf("not a number")
	case float64:
		// Check if it's actually an integer
		if v == float64(int(v)) {
			return int(v), nil
		}
		return v, nil
	case int, int32, int64:
		return v, nil
	default:
		return nil, fmt.Errorf("not a number")
	}
}

// Parses insert operations
func (p *Parser) parseInsert(collection, operation, argsString string) (*MongoOperation, error) {
	op := &MongoOperation{
		Type:       "insert",
		Collection: collection,
		Operation:  operation,
	}

	var document bson.M
	if err := p.parseJSONLikeString(argsString, &document); err != nil {
		return nil, fmt.Errorf("failed to parse insert document: %w", err)
	}

	op.Arguments = []bson.M{document}
	return op, nil
}

// Parses update operations
func (p *Parser) parseUpdate(collection, operation, argsString string) (*MongoOperation, error) {
	op := &MongoOperation{
		Type:       "update",
		Collection: collection,
		Operation:  operation,
	}

	// Parse filter and update document
	args := p.splitArguments(argsString)
	if len(args) < 2 {
		return nil, fmt.Errorf("update operation requires at least 2 arguments")
	}

	var filter, update bson.M
	if err := p.parseJSONLikeString(args[0], &filter); err != nil {
		return nil, fmt.Errorf("failed to parse update filter: %w", err)
	}
	if err := p.parseJSONLikeString(args[1], &update); err != nil {
		return nil, fmt.Errorf("failed to parse update document: %w", err)
	}

	op.Arguments = []bson.M{filter, update}
	return op, nil
}

// Parses delete operations
func (p *Parser) parseDelete(collection, operation, argsString string) (*MongoOperation, error) {
	op := &MongoOperation{
		Type:       "delete",
		Collection: collection,
		Operation:  operation,
	}

	var filter bson.M
	if err := p.parseJSONLikeString(argsString, &filter); err != nil {
		return nil, fmt.Errorf("failed to parse delete filter: %w", err)
	}

	op.Arguments = []bson.M{filter}
	return op, nil
}

// Splits JavaScript content into complete statements
func (p *Parser) splitIntoStatements(jsContent string) []string {
	var statements []string
	var current strings.Builder
	braceLevel := 0
	inQuotes := false
	var quoteChar rune

	lines := strings.Split(jsContent, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		// Add this line to current statement
		if current.Len() > 0 {
			current.WriteRune(' ')
		}
		current.WriteString(line)

		// Count braces and quotes to determine when statement ends
		for _, char := range line {
			switch char {
			case '"', '\'':
				if !inQuotes {
					inQuotes = true
					quoteChar = char
				} else if char == quoteChar {
					inQuotes = false
				}
			case '{':
				if !inQuotes {
					braceLevel++
				}
			case '}':
				if !inQuotes {
					braceLevel--
				}
			}
		}

		// If statement ends with semicolon and braces are balanced, it's complete
		if strings.HasSuffix(line, ";") && braceLevel == 0 && !inQuotes {
			statements = append(statements, current.String())
			current.Reset()
		}
	}

	// Add any remaining content as a statement
	if current.Len() > 0 {
		statements = append(statements, current.String())
	}

	return statements
}

// Parses a complete MongoDB JavaScript statement
func (p *Parser) parseMongoStatement(statement string) (*MongoOperation, error) {
	// Remove trailing semicolon and whitespace
	statement = strings.TrimSuffix(strings.TrimSpace(statement), ";")

	// Handle db.createCollection() operations
	if strings.HasPrefix(statement, "db.createCollection(") {
		return p.parseDbCreateCollection(statement)
	}

	// Handle db.collection.operation() patterns
	if !strings.HasPrefix(statement, "db.") {
		return nil, fmt.Errorf("invalid MongoDB operation format")
	}

	// Find the second dot to separate collection from operation
	firstDot := strings.Index(statement, ".")
	if firstDot == -1 || firstDot != 2 { // "db" should be followed by dot at position 2
		return nil, fmt.Errorf("invalid MongoDB operation format")
	}

	secondDot := strings.Index(statement[firstDot+1:], ".")
	if secondDot == -1 {
		return nil, fmt.Errorf("invalid MongoDB operation format")
	}
	secondDot += firstDot + 1

	collection := statement[firstDot+1 : secondDot]
	operationPart := statement[secondDot+1:]

	// Extract operation name and arguments
	parenIndex := strings.Index(operationPart, "(")
	if parenIndex == -1 {
		return nil, fmt.Errorf("no opening parenthesis found")
	}

	operation := operationPart[:parenIndex]

	// Find matching closing parenthesis
	openCount := 0
	closeIndex := -1
	for i, char := range operationPart[parenIndex:] {
		if char == '(' {
			openCount++
		} else if char == ')' {
			openCount--
			if openCount == 0 {
				closeIndex = parenIndex + i
				break
			}
		}
	}

	if closeIndex == -1 {
		return nil, fmt.Errorf("no matching closing parenthesis found")
	}

	argsString := operationPart[parenIndex+1 : closeIndex]

	// Parse arguments based on operation type
	switch operation {
	case "createIndex":
		return p.parseCreateIndex(collection, argsString)
	case "insertOne", "insertMany":
		return p.parseInsert(collection, operation, argsString)
	case "updateOne", "updateMany":
		return p.parseUpdate(collection, operation, argsString)
	case "deleteOne", "deleteMany":
		return p.parseDelete(collection, operation, argsString)
	default:
		log.Printf("Warning: unsupported operation '%s' for collection '%s'", operation, collection)
		return nil, nil
	}
}

// Handles db.createCollection() operations
func (p *Parser) parseDbCreateCollection(statement string) (*MongoOperation, error) {
	// Extract arguments from db.createCollection(collectionName, options)
	parenStart := strings.Index(statement, "(")
	parenEnd := strings.LastIndex(statement, ")")
	if parenStart == -1 || parenEnd == -1 {
		return nil, fmt.Errorf("invalid createCollection syntax")
	}

	argsString := statement[parenStart+1 : parenEnd]
	args := p.splitArguments(argsString)
	if len(args) == 0 {
		return nil, fmt.Errorf("createCollection requires collection name")
	}

	// Extract collection name (remove quotes)
	collectionName := strings.Trim(args[0], `"'`)

	op := &MongoOperation{
		Type:       "createCollection",
		Collection: collectionName,
		Operation:  "createCollection",
	}

	// Parse options if provided
	if len(args) > 1 {
		var options map[string]interface{}
		if err := p.parseJSONLikeString(args[1], &options); err != nil {
			log.Printf("Warning: failed to parse createCollection options: %v", err)
		} else {
			if validator, ok := options["validator"]; ok {
				if validatorMap, ok := validator.(map[string]interface{}); ok {
					op.Validator = validatorMap
				}
			}
		}
	}

	return op, nil
}
