package autocomplete

import (
	"strings"

	"github.com/tidwall/gjson"
)

// GetSuggestions returns autocomplete suggestions for a given path
func GetSuggestions(jsonData string, currentPath string) []string {
	// Handle empty path - suggest top-level keys
	if currentPath == "" {
		return getTopLevelKeys(jsonData)
	}

	// Parse the current path to determine the base path and incomplete segment
	basePath, incomplete := parsePathForAutocomplete(currentPath)

	// Query the JSON at the base path level
	var result gjson.Result
	if basePath == "" {
		// We're at the top level
		result = gjson.Parse(jsonData)
	} else {
		result = gjson.Get(jsonData, basePath)
	}

	// Check if the result exists and is an object or array
	if !result.Exists() {
		return []string{}
	}

	// Extract available keys at the current level
	keys := extractKeys(result)

	// Filter suggestions based on what the user has typed so far (prefix matching)
	suggestions := filterByPrefix(keys, incomplete)

	// Build full paths for suggestions
	fullSuggestions := make([]string, 0, len(suggestions))
	for _, key := range suggestions {
		if basePath == "" {
			fullSuggestions = append(fullSuggestions, key)
		} else {
			fullSuggestions = append(fullSuggestions, basePath+"."+key)
		}
	}

	return fullSuggestions
}

// parsePathForAutocomplete splits the path into base path and incomplete segment
// For example: "user.addr" -> ("user", "addr")
//              "user." -> ("user", "")
//              "user" -> ("", "user")
func parsePathForAutocomplete(path string) (basePath, incomplete string) {
	// Find the last dot
	lastDot := strings.LastIndex(path, ".")

	if lastDot == -1 {
		// No dot found, entire path is incomplete at top level
		return "", path
	}

	// Check if the dot is at the end
	if lastDot == len(path)-1 {
		// Path ends with dot, base is everything before, incomplete is empty
		return path[:lastDot], ""
	}

	// Split at the last dot
	return path[:lastDot], path[lastDot+1:]
}

// getTopLevelKeys returns all top-level keys from the JSON
func getTopLevelKeys(jsonData string) []string {
	result := gjson.Parse(jsonData)
	return extractKeys(result)
}

// extractKeys extracts keys from a gjson result (object or array indices)
func extractKeys(result gjson.Result) []string {
	keys := []string{}

	if result.IsObject() {
		// Extract object keys
		result.ForEach(func(key, value gjson.Result) bool {
			keys = append(keys, key.String())
			return true // continue iteration
		})
	} else if result.IsArray() {
		// For arrays, suggest indices as strings
		arr := result.Array()
		for i := range arr {
			keys = append(keys, string(rune('0'+i)))
		}
	}

	return keys
}

// filterByPrefix filters keys that start with the given prefix
func filterByPrefix(keys []string, prefix string) []string {
	if prefix == "" {
		return keys
	}

	filtered := []string{}
	for _, key := range keys {
		if strings.HasPrefix(key, prefix) {
			filtered = append(filtered, key)
		}
	}

	return filtered
}
