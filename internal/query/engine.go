package query

import (
	"encoding/json"
	"fmt"

	"github.com/tidwall/gjson"
)

// QueryResult represents the result of a JSON path query
type QueryResult struct {
	Value   string // The resulting value from the query
	IsValid bool   // Whether the path was valid
	Error   string // Error message if path is invalid
}

// Engine handles JSON querying with gjson and maintains state
type Engine struct {
	jsonData       string
	lastValidPath  string
	lastValidValue string
}

// NewEngine creates a new query engine with the provided JSON data
func NewEngine(jsonData string) *Engine {
	return &Engine{
		jsonData:       jsonData,
		lastValidPath:  "",
		lastValidValue: jsonData, // Initially, empty path returns the whole document
	}
}

// Query executes a gjson path query on the JSON data
func (e *Engine) Query(path string) QueryResult {
	// Handle empty path - return the entire JSON document
	if path == "" {
		// Pretty print the entire JSON
		prettyJSON, err := prettyPrintJSON(e.jsonData)
		if err != nil {
			return QueryResult{
				Value:   e.jsonData,
				IsValid: true,
				Error:   "",
			}
		}
		e.lastValidPath = ""
		e.lastValidValue = prettyJSON
		return QueryResult{
			Value:   prettyJSON,
			IsValid: true,
			Error:   "",
		}
	}

	// Execute the gjson query
	result := gjson.Get(e.jsonData, path)

	// Check if the path exists
	if !result.Exists() {
		// Path is invalid, return last valid result with error message
		return QueryResult{
			Value:   e.lastValidValue,
			IsValid: false,
			Error:   fmt.Sprintf("Invalid path: '%s' does not exist", path),
		}
	}

	// Path is valid, update last valid state
	e.lastValidPath = path

	// Get the value as a string
	valueStr := result.String()

	// If the result is a JSON object or array, pretty print it
	if result.IsObject() || result.IsArray() {
		prettyJSON, err := prettyPrintJSON(result.Raw)
		if err == nil {
			valueStr = prettyJSON
		} else {
			valueStr = result.Raw
		}
	}

	e.lastValidValue = valueStr

	return QueryResult{
		Value:   valueStr,
		IsValid: true,
		Error:   "",
	}
}

// GetLastValidPath returns the last valid path that was queried
func (e *Engine) GetLastValidPath() string {
	return e.lastValidPath
}

// GetLastValidValue returns the last valid result value
func (e *Engine) GetLastValidValue() string {
	return e.lastValidValue
}

// prettyPrintJSON formats JSON with indentation
func prettyPrintJSON(jsonStr string) (string, error) {
	var obj any
	err := json.Unmarshal([]byte(jsonStr), &obj)
	if err != nil {
		return "", err
	}

	prettyBytes, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return "", err
	}

	return string(prettyBytes), nil
}
