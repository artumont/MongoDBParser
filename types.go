package mongoparser

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Represents metadata about a setup script
type ScriptMetadata struct {
	Name         string    `json:"name"`
	Description  string    `json:"description,omitempty"`
	Version      string    `json:"version,omitempty"`
	Author       string    `json:"author,omitempty"`
	Dependencies []string  `json:"dependencies,omitempty"`
	ExecutedAt   time.Time `json:"executed_at"`
	Status       string    `json:"status"`
	Error        string    `json:"error,omitempty"`
}

// Represents a discovered script
type ScriptInfo struct {
	Name         string
	Path         string
	Content      string
	Metadata     *ScriptMetadata
	Dependencies []string
}

// Represents the result of script execution
type ScriptResult struct {
	Success bool
	Output  interface{}
	Error   error
}

// Represents a MongoDB operation parsed from JavaScript
type MongoOperation struct {
	Type         string                           `json:"type"`
	Collection   string                           `json:"collection"`
	Operation    string                           `json:"operation"`
	Arguments    []bson.M                         `json:"arguments,omitempty"`
	IndexSpec    interface{}                      `json:"index_spec,omitempty"` // Can be bson.M or bson.D
	IndexOptions *options.IndexOptions            `json:"index_options,omitempty"`
	Validator    interface{}                      `json:"validator,omitempty"` // Can be bson.M or map[string]interface{}
	CollOptions  *options.CreateCollectionOptions `json:"coll_options,omitempty"`
}
