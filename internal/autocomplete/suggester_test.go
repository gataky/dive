package autocomplete

import (
	"reflect"
	"testing"
)

func TestGetSuggestionsEmptyPath(t *testing.T) {
	jsonData := `{"name": "Alice", "age": 30, "city": "NYC"}`
	suggestions := GetSuggestions(jsonData, "")

	expected := []string{"name", "age", "city"}
	if !reflect.DeepEqual(suggestions, expected) {
		t.Errorf("Expected %v, got %v", expected, suggestions)
	}
}

func TestGetSuggestionsTopLevelPrefix(t *testing.T) {
	jsonData := `{"name": "Alice", "age": 30, "city": "NYC", "country": "USA"}`
	suggestions := GetSuggestions(jsonData, "c")

	expected := []string{"city", "country"}
	if !reflect.DeepEqual(suggestions, expected) {
		t.Errorf("Expected %v, got %v", expected, suggestions)
	}
}

func TestGetSuggestionsNestedObject(t *testing.T) {
	jsonData := `{
		"user": {
			"name": "Bob",
			"age": 25,
			"address": {
				"street": "123 Main St",
				"city": "Boston"
			}
		}
	}`

	suggestions := GetSuggestions(jsonData, "user.")
	expected := []string{"user.name", "user.age", "user.address"}
	if !reflect.DeepEqual(suggestions, expected) {
		t.Errorf("Expected %v, got %v", expected, suggestions)
	}
}

func TestGetSuggestionsNestedObjectWithPrefix(t *testing.T) {
	jsonData := `{
		"user": {
			"name": "Bob",
			"age": 25,
			"address": {
				"street": "123 Main St",
				"city": "Boston"
			}
		}
	}`

	suggestions := GetSuggestions(jsonData, "user.a")
	expected := []string{"user.age", "user.address"}
	if !reflect.DeepEqual(suggestions, expected) {
		t.Errorf("Expected %v, got %v", expected, suggestions)
	}
}

func TestGetSuggestionsDeepNested(t *testing.T) {
	jsonData := `{
		"user": {
			"address": {
				"street": "123 Main St",
				"city": "Boston",
				"zip": "02101"
			}
		}
	}`

	suggestions := GetSuggestions(jsonData, "user.address.")
	expected := []string{"user.address.street", "user.address.city", "user.address.zip"}
	if !reflect.DeepEqual(suggestions, expected) {
		t.Errorf("Expected %v, got %v", expected, suggestions)
	}
}

func TestGetSuggestionsArrayIndices(t *testing.T) {
	jsonData := `{
		"users": [
			{"name": "Alice"},
			{"name": "Bob"},
			{"name": "Charlie"}
		]
	}`

	suggestions := GetSuggestions(jsonData, "users.")
	expected := []string{"users.0", "users.1", "users.2"}
	if !reflect.DeepEqual(suggestions, expected) {
		t.Errorf("Expected %v, got %v", expected, suggestions)
	}
}

func TestGetSuggestionsArrayElement(t *testing.T) {
	jsonData := `{
		"users": [
			{"name": "Alice", "age": 25, "active": true},
			{"name": "Bob", "age": 30, "active": false}
		]
	}`

	suggestions := GetSuggestions(jsonData, "users.0.")
	expected := []string{"users.0.name", "users.0.age", "users.0.active"}
	if !reflect.DeepEqual(suggestions, expected) {
		t.Errorf("Expected %v, got %v", expected, suggestions)
	}
}

func TestGetSuggestionsInvalidPath(t *testing.T) {
	jsonData := `{"name": "Alice", "age": 30}`
	suggestions := GetSuggestions(jsonData, "nonexistent.")

	if len(suggestions) != 0 {
		t.Errorf("Expected empty suggestions for invalid path, got %v", suggestions)
	}
}

func TestGetSuggestionsNoMatch(t *testing.T) {
	jsonData := `{"name": "Alice", "age": 30, "city": "NYC"}`
	suggestions := GetSuggestions(jsonData, "xyz")

	if len(suggestions) != 0 {
		t.Errorf("Expected empty suggestions for no prefix match, got %v", suggestions)
	}
}

func TestGetSuggestionsComplexJSON(t *testing.T) {
	jsonData := `{
		"company": {
			"name": "Acme Corp",
			"employees": [
				{
					"id": 1,
					"name": "Alice",
					"department": "Engineering"
				},
				{
					"id": 2,
					"name": "Bob",
					"department": "Sales"
				}
			],
			"location": {
				"city": "San Francisco",
				"state": "CA"
			}
		}
	}`

	// Test top-level
	suggestions := GetSuggestions(jsonData, "")
	if len(suggestions) != 1 || suggestions[0] != "company" {
		t.Errorf("Expected [company], got %v", suggestions)
	}

	// Test nested object
	suggestions = GetSuggestions(jsonData, "company.")
	expected := []string{"company.name", "company.employees", "company.location"}
	if !reflect.DeepEqual(suggestions, expected) {
		t.Errorf("Expected %v, got %v", expected, suggestions)
	}

	// Test array element
	suggestions = GetSuggestions(jsonData, "company.employees.0.")
	expected = []string{"company.employees.0.id", "company.employees.0.name", "company.employees.0.department"}
	if !reflect.DeepEqual(suggestions, expected) {
		t.Errorf("Expected %v, got %v", expected, suggestions)
	}

	// Test deep nested
	suggestions = GetSuggestions(jsonData, "company.location.")
	expected = []string{"company.location.city", "company.location.state"}
	if !reflect.DeepEqual(suggestions, expected) {
		t.Errorf("Expected %v, got %v", expected, suggestions)
	}
}

func TestParsePathForAutocomplete(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		wantBase   string
		wantIncomplete string
	}{
		{"empty path", "", "", ""},
		{"top-level incomplete", "user", "", "user"},
		{"dot at end", "user.", "user", ""},
		{"nested incomplete", "user.addr", "user", "addr"},
		{"deep nested dot", "user.address.", "user.address", ""},
		{"deep nested incomplete", "user.address.city", "user.address", "city"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base, incomplete := parsePathForAutocomplete(tt.path)
			if base != tt.wantBase || incomplete != tt.wantIncomplete {
				t.Errorf("parsePathForAutocomplete(%q) = (%q, %q), want (%q, %q)",
					tt.path, base, incomplete, tt.wantBase, tt.wantIncomplete)
			}
		})
	}
}

func TestGetSuggestionsSpecialCharacters(t *testing.T) {
	jsonData := `{"field-name": "value", "field_name": "value2", "fieldName": "value3"}`

	suggestions := GetSuggestions(jsonData, "")
	expected := []string{"field-name", "field_name", "fieldName"}
	if !reflect.DeepEqual(suggestions, expected) {
		t.Errorf("Expected %v, got %v", expected, suggestions)
	}

	suggestions = GetSuggestions(jsonData, "field")
	if !reflect.DeepEqual(suggestions, expected) {
		t.Errorf("Expected %v for prefix 'field', got %v", expected, suggestions)
	}
}
