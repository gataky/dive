# dive - Interactive JSON Viewer

`dive` is a fast, terminal-based interactive JSON viewer that lets you explore and query JSON data using gjson path syntax with real-time autocomplete.

## Features

- 🚀 **Real-time Query Engine** - Type gjson paths and see results instantly
- 🎯 **Smart Autocomplete** - Press Tab for intelligent path suggestions
- 🎨 **Visual Feedback** - Color-coded input (green for valid paths, red for invalid)
- 📋 **Clipboard Support** - Copy results with Ctrl+C
- 💾 **Save to File** - Save query results with Ctrl+S
- ⌨️  **Keyboard Navigation** - Fully keyboard-driven interface
- 📦 **Flexible Input** - Read from files or stdin

## Installation

### Prerequisites

- Go 1.21 or higher

### From Source

```bash
git clone <repository-url>
cd dive
go build
```

The binary will be created as `./dive` in the current directory.

### Install to PATH

```bash
go install
```

## Usage

### Basic Usage

```bash
# View a JSON file
./dive data.json

# Pipe JSON from stdin
cat data.json | ./dive

# Use with curl
curl https://api.example.com/data | ./dive

# Use with jq
echo '{"users":[{"name":"Alice"}]}' | jq . | ./dive
```

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Tab` | Show autocomplete suggestions |
| `↑` / `↓` | Navigate autocomplete dropdown |
| `Enter` | Select autocomplete suggestion |
| `Esc` | Hide autocomplete dropdown |
| `Ctrl+C` | Copy current output to clipboard |
| `Ctrl+S` | Save output to file |
| `Ctrl+Q` | Quit application |

## gjson Path Syntax

`dive` uses [gjson](https://github.com/tidwall/gjson) for path queries. Here are some common patterns:

### Basic Paths

```
name                 # Get a top-level field
user.name           # Nested field access
user.address.city   # Deep nesting
```

### Array Access

```
users.0             # First element
users.0.name        # Field in first element
users.#             # Count of array elements
```

### Advanced Queries

gjson supports many more features like queries, modifiers, and more. See the [gjson syntax guide](https://github.com/tidwall/gjson/blob/master/SYNTAX.md) for complete documentation.

## Example Session

Given this JSON file (`test.json`):

```json
{
  "company": {
    "name": "Acme Corp",
    "employees": [
      {
        "id": 1,
        "name": "Alice",
        "role": "Engineer"
      },
      {
        "id": 2,
        "name": "Bob",
        "role": "Designer"
      }
    ]
  }
}
```

Run `./dive test.json` and try these paths:

- `company.name` → Returns: "Acme Corp"
- `company.employees.0.name` → Returns: "Alice"
- `company.employees.1.role` → Returns: "Designer"
- Press `Tab` after typing `company.` to see available fields

## Features in Detail

### Autocomplete

Press `Tab` at any time to see available keys at your current path level. The autocomplete system:

- Shows only valid keys from your JSON structure
- Filters suggestions based on what you've typed
- Works with nested objects and array indices
- Navigate with arrow keys, select with Enter

### Visual Feedback

The query input border color indicates path validity:
- **Green border** - Valid path with results
- **Red border** - Invalid path (last valid result is retained)

### Export Options

**Copy to Clipboard (Ctrl+C)**
- Copies the current query result to your system clipboard
- Shows confirmation message in footer

**Save to File (Ctrl+S)**
- Opens a dialog to enter filename
- Default filename: `output.json`
- Creates directories if they don't exist
- Press Enter to save, Esc to cancel

## Architecture

```
dive/
├── main.go                          # Entry point
├── internal/
│   ├── input/                       # JSON input handling
│   │   ├── reader.go
│   │   └── reader_test.go
│   ├── query/                       # gjson query engine
│   │   ├── engine.go
│   │   └── engine_test.go
│   ├── autocomplete/                # Autocomplete system
│   │   ├── suggester.go
│   │   └── suggester_test.go
│   ├── export/                      # Export functionality
│   │   ├── clipboard.go
│   │   ├── file.go
│   │   └── export_test.go
│   └── ui/                          # Terminal UI
│       ├── app.go
│       └── components.go
└── test.json                        # Sample data
```

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests for a specific package
go test ./internal/query

# Run tests with verbose output
go test -v ./...
```

### Building

```bash
# Build for current platform
go build

# Build for specific platform
GOOS=linux GOARCH=amd64 go build
GOOS=darwin GOARCH=arm64 go build
GOOS=windows GOARCH=amd64 go build
```

## Dependencies

- [tidwall/gjson](https://github.com/tidwall/gjson) - JSON path queries
- [rivo/tview](https://github.com/rivo/tview) - Terminal UI framework
- [gdamore/tcell](https://github.com/gdamore/tcell) - Terminal handling
- [atotto/clipboard](https://github.com/atotto/clipboard) - Clipboard support

## License

[Your License Here]

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Acknowledgments

- Built with [gjson](https://github.com/tidwall/gjson) by Josh Baker
- UI powered by [tview](https://github.com/rivo/tview)
