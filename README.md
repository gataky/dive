# dive - Interactive JSON Viewer

`dive` is a fast, terminal-based interactive JSON viewer that lets you explore and query JSON data using gjson path syntax with real-time autocomplete.

## Features

- 🚀 **Real-time Query Engine** - Type gjson paths and see results instantly
- 🎯 **Live Autocomplete** - Always-visible suggestions with Tab cycling
- 🎨 **Visual Feedback** - Output title tracks the matched path; red border marks unmatched input
- 📋 **Clipboard Support** - Copy results with Ctrl+Y
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
| `Tab` / `Shift+Tab` | Cycle through completions (back to your typed text at the end) |
| `Esc` | Cancel completion cycling / leave output panel |
| `F1` | Toggle gjson syntax help |
| `Ctrl+O` | Focus output panel for scrolling |
| `Ctrl+Y` | Copy current output to clipboard |
| `Ctrl+S` | Save output to file |
| `Ctrl+C` / `Ctrl+Q` | Quit application |

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
- Type `company.` and the suggestions bar shows the available fields; press `Tab` to cycle through them

## Features in Detail

### Autocomplete

The suggestions bar under the query box is always visible and updates as you
type, showing the keys (or array indices) available at your current path
level. Press `Tab` to cycle through the candidates — your originally typed
text stays in the cycle, so tabbing past the last candidate brings it back.
`Shift+Tab` cycles backward and `Esc` cancels the cycle. Completions under
array-query results (like `users.#(age>30)#`) are offered in gjson's pipe
form (`users.#(age>30)#|0`), which is the syntax gjson requires there.

### Visual Feedback

The output panel always shows the deepest part of your path that matches the
document — a typo never blanks the view. Its title names the path being
shown, and when part of your path doesn't match (e.g. `users.0.nmae`), the
title shows where matching stopped (`users.0 ✗ nmae`) and the query border
turns red. Output JSON is syntax-highlighted.

### Export Options

**Copy to Clipboard (Ctrl+Y)**
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
│   ├── path/                        # gjson path segmentation
│   ├── query/                       # deepest-valid-ancestor resolution
│   ├── autocomplete/                # suggestions + Tab-cycle state
│   ├── render/                      # pretty-printing and colorizing
│   ├── export/                      # clipboard / file export
│   └── ui/                          # tview layout and key bindings
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

## Known Limitations

- Very large documents (multiple MB) can show noticeable keystroke latency,
  since the resolved value is re-rendered on every keystroke.
- Keys that contain a literal dot (like `fav.movie`) are suggested in
  escaped form (`fav\.movie`); typing the unescaped dot parses as a path
  separator, so suggestions pause until the escaped form is used.
- gjson has no array-slicing syntax (`users.0:3` is not a gjson feature).

## Dependencies

- [tidwall/gjson](https://github.com/tidwall/gjson) - JSON path queries
- [rivo/tview](https://github.com/rivo/tview) - Terminal UI framework
- [gdamore/tcell](https://github.com/gdamore/tcell) - Terminal handling
- [atotto/clipboard](https://github.com/atotto/clipboard) - Clipboard support
- [mattn/go-runewidth](https://github.com/mattn/go-runewidth) - Display width measurement
- [tidwall/pretty](https://github.com/tidwall/pretty) - JSON formatting

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Acknowledgments

- Built with [gjson](https://github.com/tidwall/gjson) by Josh Baker
- UI powered by [tview](https://github.com/rivo/tview)
