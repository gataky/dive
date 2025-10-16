package query

import (
	"strings"
	"testing"
)

func TestNewEngine(t *testing.T) {
	jsonData := `{"name": "test"}`
	engine := NewEngine(jsonData)

	if engine == nil {
		t.Fatal("Expected engine to be created, got nil")
	}

	if engine.jsonData != jsonData {
		t.Errorf("Expected jsonData to be %s, got %s", jsonData, engine.jsonData)
	}

	if engine.lastValidPath != "" {
		t.Errorf("Expected lastValidPath to be empty, got %s", engine.lastValidPath)
	}
}

func TestQueryEmptyPath(t *testing.T) {
	jsonData := `{"name": "Alice", "age": 30}`
	engine := NewEngine(jsonData)

	result := engine.Query("")

	if !result.IsValid {
		t.Error("Expected empty path to be valid")
	}

	if result.Error != "" {
		t.Errorf("Expected no error, got: %s", result.Error)
	}

	// Should return pretty-printed entire JSON
	if !strings.Contains(result.Value, "Alice") {
		t.Error("Expected result to contain the JSON data")
	}
}

func TestQueryValidPath(t *testing.T) {
	jsonData := `{"name": "Alice", "age": 30, "city": "NYC"}`
	engine := NewEngine(jsonData)

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"simple string", "name", "Alice"},
		{"simple number", "age", "30"},
		{"another string", "city", "NYC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.Query(tt.path)

			if !result.IsValid {
				t.Errorf("Expected path '%s' to be valid", tt.path)
			}

			if result.Error != "" {
				t.Errorf("Expected no error, got: %s", result.Error)
			}

			if result.Value != tt.expected {
				t.Errorf("Expected value '%s', got '%s'", tt.expected, result.Value)
			}

			// Check that last valid path and value were updated
			if engine.GetLastValidPath() != tt.path {
				t.Errorf("Expected lastValidPath to be '%s', got '%s'", tt.path, engine.GetLastValidPath())
			}

			if engine.GetLastValidValue() != tt.expected {
				t.Errorf("Expected lastValidValue to be '%s', got '%s'", tt.expected, engine.GetLastValidValue())
			}
		})
	}
}

func TestQueryNestedPath(t *testing.T) {
	jsonData := `{
		"user": {
			"name": "Bob",
			"address": {
				"street": "123 Main St",
				"city": "Boston"
			}
		}
	}`
	engine := NewEngine(jsonData)

	result := engine.Query("user.name")
	if !result.IsValid {
		t.Error("Expected nested path to be valid")
	}
	if result.Value != "Bob" {
		t.Errorf("Expected 'Bob', got '%s'", result.Value)
	}

	result = engine.Query("user.address.city")
	if !result.IsValid {
		t.Error("Expected deep nested path to be valid")
	}
	if result.Value != "Boston" {
		t.Errorf("Expected 'Boston', got '%s'", result.Value)
	}
}

func TestQueryArrayPath(t *testing.T) {
	jsonData := `{
		"users": [
			{"name": "Alice", "age": 25},
			{"name": "Bob", "age": 30},
			{"name": "Charlie", "age": 35}
		]
	}`
	engine := NewEngine(jsonData)

	result := engine.Query("users.0.name")
	if !result.IsValid {
		t.Error("Expected array index path to be valid")
	}
	if result.Value != "Alice" {
		t.Errorf("Expected 'Alice', got '%s'", result.Value)
	}

	result = engine.Query("users.1.age")
	if !result.IsValid {
		t.Error("Expected array index path to be valid")
	}
	if result.Value != "30" {
		t.Errorf("Expected '30', got '%s'", result.Value)
	}

	result = engine.Query("users.2.name")
	if !result.IsValid {
		t.Error("Expected array index path to be valid")
	}
	if result.Value != "Charlie" {
		t.Errorf("Expected 'Charlie', got '%s'", result.Value)
	}
}

func TestQueryInvalidPath(t *testing.T) {
	jsonData := `{"name": "Alice", "age": 30}`
	engine := NewEngine(jsonData)

	// First, query a valid path to set lastValidValue
	validResult := engine.Query("name")
	if !validResult.IsValid {
		t.Fatal("Expected valid path to succeed")
	}

	// Now query an invalid path
	result := engine.Query("nonexistent")

	if result.IsValid {
		t.Error("Expected invalid path to return IsValid=false")
	}

	if result.Error == "" {
		t.Error("Expected error message for invalid path")
	}

	if !strings.Contains(result.Error, "nonexistent") {
		t.Errorf("Expected error to mention the invalid path, got: %s", result.Error)
	}

	// Should return last valid value
	if result.Value != validResult.Value {
		t.Errorf("Expected last valid value '%s', got '%s'", validResult.Value, result.Value)
	}

	// Last valid path should not change
	if engine.GetLastValidPath() != "name" {
		t.Errorf("Expected lastValidPath to remain 'name', got '%s'", engine.GetLastValidPath())
	}
}

func TestQueryInvalidPathWithoutPriorValidPath(t *testing.T) {
	jsonData := `{"name": "Alice", "age": 30}`
	engine := NewEngine(jsonData)

	// Query an invalid path without any prior valid path
	result := engine.Query("nonexistent")

	if result.IsValid {
		t.Error("Expected invalid path to return IsValid=false")
	}

	if result.Error == "" {
		t.Error("Expected error message for invalid path")
	}

	// Should return the initial lastValidValue (pretty-printed JSON)
	if !strings.Contains(result.Value, "Alice") {
		t.Error("Expected result to contain initial JSON data")
	}
}

func TestQueryObjectResult(t *testing.T) {
	jsonData := `{
		"user": {
			"name": "Alice",
			"age": 30
		}
	}`
	engine := NewEngine(jsonData)

	result := engine.Query("user")

	if !result.IsValid {
		t.Error("Expected path to be valid")
	}

	// Should return pretty-printed JSON object
	if !strings.Contains(result.Value, "Alice") || !strings.Contains(result.Value, "30") {
		t.Errorf("Expected result to contain user object data, got: %s", result.Value)
	}
}

func TestQueryArrayResult(t *testing.T) {
	jsonData := `{
		"numbers": [1, 2, 3, 4, 5]
	}`
	engine := NewEngine(jsonData)

	result := engine.Query("numbers")

	if !result.IsValid {
		t.Error("Expected path to be valid")
	}

	// Should return array
	if !strings.Contains(result.Value, "1") || !strings.Contains(result.Value, "5") {
		t.Errorf("Expected result to contain array data, got: %s", result.Value)
	}
}

func TestQueryWithSpecialCharacters(t *testing.T) {
	jsonData := `{"field-name": "value", "field.with.dots": "value2"}`
	engine := NewEngine(jsonData)

	// gjson handles special characters with escaping
	result := engine.Query("field-name")
	if !result.IsValid {
		t.Error("Expected path with dash to be valid")
	}
	if result.Value != "value" {
		t.Errorf("Expected 'value', got '%s'", result.Value)
	}
}

func TestQueryStateTracking(t *testing.T) {
	jsonData := `{"a": 1, "b": 2, "c": 3}`
	engine := NewEngine(jsonData)

	// Query sequence: valid -> valid -> invalid
	engine.Query("a")
	if engine.GetLastValidPath() != "a" || engine.GetLastValidValue() != "1" {
		t.Error("State not updated correctly after first query")
	}

	engine.Query("b")
	if engine.GetLastValidPath() != "b" || engine.GetLastValidValue() != "2" {
		t.Error("State not updated correctly after second query")
	}

	result := engine.Query("invalid")
	if result.IsValid {
		t.Error("Expected invalid path")
	}
	// State should remain at last valid
	if engine.GetLastValidPath() != "b" || engine.GetLastValidValue() != "2" {
		t.Error("State should not change for invalid query")
	}
}
