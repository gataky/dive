# Task List: Interactive JSON Viewer CLI Tool

Based on PRD: `0001-prd-interactive-json-viewer.md`

## Current State Assessment

This is a greenfield project with no existing Go codebase. The project directory contains only the `.rules` and `tasks` folders. We will be building the entire application from scratch following Go best practices and idiomatic patterns.

## Relevant Files

- `main.go` - Entry point for the CLI application, handles command-line arguments
- `go.mod` - Go module definition with dependencies (gjson, tview, clipboard)
- `internal/input/reader.go` - Handles JSON input from files and stdin
- `internal/input/reader_test.go` - Unit tests for input reader
- `internal/ui/app.go` - Main tview application setup and layout management
- `internal/ui/components.go` - Individual UI components (header, footer, input field, output panel)
- `internal/ui/components_test.go` - Unit tests for UI components
- `internal/query/engine.go` - gjson query engine with path validation
- `internal/query/engine_test.go` - Unit tests for query engine
- `internal/autocomplete/suggester.go` - Autocomplete logic for generating key suggestions
- `internal/autocomplete/suggester_test.go` - Unit tests for autocomplete suggester
- `internal/export/clipboard.go` - Clipboard copy functionality
- `internal/export/file.go` - Save to file functionality
- `internal/export/export_test.go` - Unit tests for export features
- `internal/formatter/json.go` - JSON pretty-printing and syntax highlighting
- `internal/formatter/json_test.go` - Unit tests for JSON formatter

### Notes

- Unit tests should be placed alongside the code files they are testing (e.g., `engine.go` and `engine_test.go` in the same directory).
- Use `go test ./...` to run all tests in the project.
- Use `go test ./internal/query` to run tests for a specific package.

## Tasks

- [x] 1.0 Project Setup & Dependencies
  - [x] 1.1 Initialize Go module with `go mod init` (choose appropriate module path)
  - [x] 1.2 Add tidwall/gjson dependency (`go get github.com/tidwall/gjson`)
  - [x] 1.3 Add rivo/tview dependency (`go get github.com/rivo/tview`)
  - [x] 1.4 Add gdamore/tcell dependency for tview (`go get github.com/gdamore/tcell/v2`)
  - [x] 1.5 Add atotto/clipboard dependency (`go get github.com/atotto/clipboard`)
  - [x] 1.6 Create project directory structure (`internal/input`, `internal/ui`, `internal/query`, `internal/autocomplete`, `internal/export`, `internal/formatter`)
  - [x] 1.7 Create main.go with basic CLI argument parsing (check for file path arg or stdin)

- [x] 2.0 JSON Input Handler & Core Data Model
  - [x] 2.1 Implement `internal/input/reader.go` with `ReadFromFile(path string)` function
  - [x] 2.2 Implement `ReadFromStdin()` function in reader.go
  - [x] 2.3 Add JSON validation using `json.Valid()` from standard library
  - [x] 2.4 Return raw JSON as string and any errors
  - [x] 2.5 Write unit tests in `reader_test.go` for file reading, stdin reading, and invalid JSON handling
  - [x] 2.6 Handle edge cases: empty files, non-existent files, invalid JSON with clear error messages

- [x] 3.0 Terminal UI Layout with tview
  - [x] 3.1 Create `internal/ui/app.go` with `NewApp()` function that initializes tview application
  - [x] 3.2 Implement header component in `components.go` - display app name "dive - Interactive JSON Viewer" and brief instructions
  - [x] 3.3 Implement path input field component using `tview.InputField` with label "Path: "
  - [x] 3.4 Implement output panel component using `tview.TextView` with scrolling enabled and word wrap
  - [x] 3.5 Implement footer component showing keybindings: "Tab: Autocomplete | Ctrl+C: Copy | Ctrl+S: Save | Ctrl+Q: Quit"
  - [x] 3.6 Use `tview.Flex` layout to arrange components vertically (header, input, output, footer)
  - [x] 3.7 Configure panel borders and titles using tview's box styling
  - [x] 3.8 Set up application focus management so input field has initial focus

- [x] 4.0 Path Query Engine & Real-time Updates
  - [x] 4.1 Create `internal/query/engine.go` with `Query(jsonData string, path string)` function
  - [x] 4.2 Implement gjson integration using `gjson.Get(jsonData, path)` to execute queries
  - [x] 4.3 Add path validation logic - check if `gjson.Get()` returns `.Exists()` as true
  - [x] 4.4 Implement state tracking for last valid path and last valid result
  - [x] 4.5 Create `QueryResult` struct with fields: `Value string`, `IsValid bool`, `Error string`
  - [x] 4.6 Handle invalid paths by returning last valid result with error message
  - [x] 4.7 Wire up input field's `SetChangedFunc()` callback to call query engine on each keystroke
  - [x] 4.8 Update output panel with query results in real-time
  - [x] 4.9 Implement visual feedback - change input field label color to red when path is invalid
  - [x] 4.10 Restore normal color when path becomes valid again
  - [x] 4.11 Write unit tests in `engine_test.go` for valid paths, invalid paths, edge cases (empty path, special characters)

- [x] 5.0 Autocomplete System
  - [x] 5.1 Create `internal/autocomplete/suggester.go` with `GetSuggestions(jsonData string, currentPath string)` function
  - [x] 5.2 Parse the current path to determine the base path (everything before the last incomplete segment)
  - [x] 5.3 Use gjson to query the JSON at the base path level
  - [x] 5.4 Extract available keys at the current level from the gjson result
  - [x] 5.5 Filter suggestions based on what the user has typed so far (prefix matching)
  - [x] 5.6 Return a slice of suggestion strings
  - [x] 5.7 Create autocomplete dropdown UI using `tview.List` component
  - [x] 5.8 Show dropdown below input field when suggestions are available
  - [x] 5.9 Handle Tab key press in input field to show autocomplete dropdown
  - [x] 5.10 Handle arrow keys (Up/Down) to navigate dropdown selections
  - [x] 5.11 Handle Enter key to select a suggestion and update the path input
  - [x] 5.12 Hide dropdown when Escape is pressed or when no suggestions are available
  - [x] 5.13 Write unit tests in `suggester_test.go` for various path scenarios and JSON structures

- [x] 6.0 Export Features (Clipboard & File Save)
  - [x] 6.1 Create `internal/export/clipboard.go` with `CopyToClipboard(content string)` function
  - [x] 6.2 Implement clipboard copy using atotto/clipboard library
  - [x] 6.3 Handle clipboard errors gracefully (show error message if clipboard unavailable)
  - [x] 6.4 Create `internal/export/file.go` with `SaveToFile(content string, filepath string)` function
  - [x] 6.5 Implement file writing with proper error handling and file permissions
  - [x] 6.6 Create a modal dialog using `tview.Modal` or `tview.InputField` to prompt for filename when saving
  - [x] 6.7 Wire up Ctrl+C keybinding to copy current output to clipboard
  - [x] 6.8 Wire up Ctrl+S keybinding to open save dialog
  - [x] 6.9 Show success/error messages in footer or as temporary notification
  - [x] 6.10 Wire up Ctrl+Q keybinding to quit the application
  - [x] 6.11 Write unit tests in `export_test.go` for file saving (use temp files)

### Additional Tasks (Nice to Have)

- [ ] 7.0 JSON Syntax Highlighting
  - [ ] 7.1 Create `internal/formatter/json.go` with `FormatWithColors(jsonString string)` function
  - [ ] 7.2 Parse JSON and apply tview color tags for different token types
  - [ ] 7.3 Use tcell colors: cyan for keys, green for strings, yellow for numbers, magenta for booleans, gray for null
  - [ ] 7.4 Handle nested structures with proper indentation
  - [ ] 7.5 Apply formatted output to the output panel
  - [ ] 7.6 Write unit tests in `json_test.go` for various JSON structures

- [ ] 8.0 Polish & Error Handling
  - [ ] 8.1 Add comprehensive error messages for all failure scenarios
  - [ ] 8.2 Implement graceful handling of Ctrl+C (SIGINT) to ensure clean exit
  - [ ] 8.3 Add usage instructions when no file is provided and stdin is empty
  - [ ] 8.4 Test with various JSON file sizes to ensure performance is acceptable
  - [ ] 8.5 Add README.md with installation instructions, usage examples, and screenshots
  - [ ] 8.6 Consider adding version flag (-v, --version) to CLI
