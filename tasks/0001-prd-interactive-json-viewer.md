# PRD: Interactive JSON Viewer CLI Tool

## Introduction/Overview

The Interactive JSON Viewer is a command-line tool that allows users to explore and navigate JSON data interactively. Users can type a path to navigate through JSON structures, and the display updates in real-time as they type. This tool addresses three key use cases:
- Navigating and exploring large or complex JSON files more easily
- Interactively learning and understanding JSON structure
- Debugging and inspecting API responses

The tool is built in Go, using `tidwall/gjson` for JSON querying and `tview` for the terminal UI.

## Goals

1. Provide an interactive, real-time way to explore JSON data from files or stdin
2. Support both simple dot notation and advanced gjson query syntax for flexible path specification
3. Display filtered JSON results with syntax highlighting in a clear, readable format
4. Enable users to quickly extract and save specific JSON values
5. Create an intuitive experience that helps users understand JSON structure through autocomplete suggestions

## User Stories

1. As a **developer debugging an API response**, I want to quickly navigate to specific fields in a large JSON payload so that I can inspect values without writing scripts or using complex command-line tools.

2. As a **DevOps engineer**, I want to explore configuration files interactively so that I can understand their structure and extract specific values for use in scripts.

3. As a **data analyst**, I want to filter nested JSON data visually so that I can identify the paths to the data I need for my analysis.

4. As a **junior developer**, I want autocomplete suggestions showing available keys at my current path level so that I can discover the JSON structure without referring to external documentation.

5. As a **user working with complex queries**, I want to use advanced gjson syntax (wildcards, filters, modifiers) so that I can extract arrays, apply conditions, and transform data.

6. As a **user who found the data I need**, I want to copy the filtered result to my clipboard or save it to a file so that I can use it in other tools or scripts.

## Functional Requirements

### Core Functionality

1. The system must accept JSON input from either a file path argument or stdin (pipe).
2. The system must display an interactive terminal UI with an input field for the JSON path.
3. The system must update the displayed output in real-time as the user types a path.
4. The system must support gjson dot notation for simple paths (e.g., `user.name`, `items.0.title`).
5. The system must support full gjson query syntax including wildcards (`*`), array filters (`#(...)#`), and modifiers (e.g., `@reverse`, `@flatten`).
6. The system must display only the filtered result (the value at the specified path) in the main output area.
7. The system must pretty-print the filtered JSON output with syntax highlighting for readability.

### Error Handling & Visual Feedback

8. When the user types an invalid or incomplete path, the system must continue displaying the last valid result.
9. The system must show an error message when the path is invalid or not found in the JSON.
10. The system must visually indicate invalid paths by highlighting the path input in red.
11. When a valid path is entered, the system must return the path input highlighting to normal.

### Autocomplete

12. The system must provide autocomplete suggestions as the user types.
13. Autocomplete must suggest available keys at the current path level.
14. Users must be able to select from autocomplete suggestions to complete their path.

### Export Functionality

15. The system must allow users to copy the filtered result to the system clipboard.
16. The system must allow users to save the filtered result to a file with a user-specified filename.

### User Interface

17. The interface must display the current path being typed by the user.
18. The interface must display the filtered JSON result in a scrollable view.
19. The interface must display helpful keybindings/shortcuts (e.g., for copy, save, quit).
20. The system must provide a way to exit the program cleanly (e.g., Ctrl+C, Esc, or 'q').

## Non-Goals (Out of Scope)

1. **Editing/Modifying JSON**: This tool is read-only. Users cannot edit, add, or delete JSON data.
2. **Non-JSON Formats**: The tool will only support JSON input. Formats like YAML, XML, TOML, etc. are out of scope.
3. **Large File Handling**: The tool is designed for JSON files that fit comfortably in memory (typically under 100MB). Streaming or pagination of extremely large files is not supported.
4. **Network Requests**: The tool will not fetch JSON from URLs or APIs. It only works with local files or stdin.
5. **Multiple File Viewing**: The tool processes one JSON input at a time.

## Design Considerations

### Terminal UI Layout

The `tview` interface should include:

- **Header Area**: Display the application name and brief instructions
- **Path Input Field**: A prominent input field where users type their JSON path
  - Should have clear visual distinction between valid (normal) and invalid (red) states
- **Autocomplete Dropdown**: A popup list that appears below the input field when suggestions are available
- **Output Panel**: The main display area showing the filtered JSON result
  - Should be scrollable for large results
  - Should have syntax highlighting (different colors for keys, values, strings, numbers, booleans, null)
- **Footer Area**: Display keyboard shortcuts and current status
  - Example shortcuts: `Ctrl+C: Copy`, `Ctrl+S: Save`, `Ctrl+Q: Quit`, `Tab: Autocomplete`

### Syntax Highlighting

Use distinct colors for:
- Keys (e.g., cyan)
- String values (e.g., green)
- Numbers (e.g., yellow)
- Booleans (e.g., magenta)
- Null values (e.g., gray)
- Structural characters like `{}`, `[]`, `:`, `,` (e.g., white/default)

## Technical Considerations

### Dependencies

- **Language**: Go (Golang)
- **JSON Querying**: `tidwall/gjson` - for parsing and querying JSON with flexible path syntax
- **Terminal UI**: `tview` - for building the interactive terminal interface
- **Clipboard**: Consider `atotto/clipboard` or similar for cross-platform clipboard support

### Architecture

1. **Input Handler**: Read JSON from file or stdin, validate it's valid JSON
2. **Query Engine**: Use gjson to execute path queries against the JSON data
3. **UI Manager**: Manage tview components and handle user input
4. **Autocomplete Engine**: Parse current path and JSON structure to generate key suggestions
5. **Output Formatter**: Pretty-print and apply syntax highlighting to filtered results

### Path Validation

- Use gjson's `.Get()` method to validate paths
- Track the last valid path and result
- Parse path incrementally to detect where it becomes invalid

### Performance

- For typical JSON files (<10MB), performance should be near-instantaneous
- Consider debouncing rapid keystrokes if performance degrades with very large files
- gjson is highly optimized and should handle most queries quickly

## Success Metrics

1. **Usability**: Users can successfully navigate to nested values in JSON files within seconds, without external documentation.
2. **Time Savings**: Reduce the time to extract specific JSON values by 50% compared to manual scripting or using `jq` with trial-and-error.
3. **Adoption**: Positive feedback from target user groups (developers, DevOps engineers, data analysts).
4. **Autocomplete Effectiveness**: 80% of paths are completed using autocomplete suggestions, indicating users find the feature helpful.

## Open Questions

1. **Clipboard Support**: Should we gracefully degrade if clipboard access is unavailable on a system, or make it a hard requirement?
2. **File Size Warning**: Should we warn users when loading files above a certain size threshold (e.g., 50MB)?
3. **Configuration**: Should users be able to configure syntax highlighting colors, keybindings, or other preferences (e.g., via a config file)?
4. **History**: Should the tool remember previously used paths within a session? Across sessions?
5. **Installation**: How should the tool be distributed? (Go install, homebrew, pre-built binaries, etc.)
6. **Example Keybindings**: What specific keyboard shortcuts should be used for copy, save, and quit? (e.g., should we follow common conventions like Ctrl+C for copy, or avoid conflicts with terminal signals?)
