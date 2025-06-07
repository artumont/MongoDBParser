package mongoparser

import (
	"testing"
)

func TestNewParser(t *testing.T) {
	parser := NewParser()
	if parser == nil {
		t.Fatal("NewParser() returned nil")
	}
}

func TestParseMetadata(t *testing.T) {
	parser := NewParser()

	// Test valid metadata
	script := `
		// METADATA:
		// {
		//   "description": "Test script",
		//   "version": "1.0.0",
		//   "author": "Test Author"
		// }

		db.createCollection("test");
	`

	metadata := parser.ParseMetadata(script)
	if metadata == nil {
		t.Fatal("ParseMetadata() returned nil for valid metadata")
	}

	if metadata.Description != "Test script" {
		t.Errorf("Expected description 'Test script', got '%s'", metadata.Description)
	}

	if metadata.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", metadata.Version)
	}

	if metadata.Author != "Test Author" {
		t.Errorf("Expected author 'Test Author', got '%s'", metadata.Author)
	}
}

func TestParseMetadataNoMetadata(t *testing.T) {
	parser := NewParser()

	// Test script without metadata
	script := `
		db.createCollection("test");
	`

	metadata := parser.ParseMetadata(script)
	if metadata != nil {
		t.Error("ParseMetadata() should return nil for script without metadata")
	}
}
