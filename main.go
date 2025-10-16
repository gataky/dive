package main

import (
	"fmt"
	"os"

	"github.com/gataky/dive/internal/input"
	"github.com/gataky/dive/internal/ui"
)

func main() {
	// Read JSON data from file or stdin
	var jsonData string
	var err error

	if len(os.Args) > 1 {
		// File path provided as argument
		filePath := os.Args[1]
		jsonData, err = input.ReadFromFile(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}
	} else {
		// No argument provided, try reading from stdin
		jsonData, err = input.ReadFromStdin()
		if err != nil {
			// No piped input
			printUsage()
			os.Exit(1)
		}
	}

	// Validate that we have non-empty JSON data
	if jsonData == "" {
		fmt.Fprintf(os.Stderr, "Error: empty JSON data\n")
		os.Exit(1)
	}

	// Initialize and run the UI
	app := ui.NewApp(jsonData)
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: dive <json-file>\n")
	fmt.Fprintf(os.Stderr, "   or: cat <json-file> | dive\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "dive - Interactive JSON Viewer\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Provide a JSON file as an argument or pipe JSON data via stdin.\n")
}
