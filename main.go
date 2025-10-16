package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	// Check if a file path was provided as an argument
	var jsonData []byte
	var err error

	if len(os.Args) > 1 {
		// File path provided as argument
		filePath := os.Args[1]
		jsonData, err = os.ReadFile(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", filePath, err)
			os.Exit(1)
		}
	} else {
		// No argument provided, check if stdin has data
		stat, err := os.Stdin.Stat()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error checking stdin: %v\n", err)
			os.Exit(1)
		}

		// Check if stdin is being piped
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			// Data is being piped to stdin
			jsonData, err = io.ReadAll(os.Stdin)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
				os.Exit(1)
			}
		} else {
			// No file provided and no piped input
			printUsage()
			os.Exit(1)
		}
	}

	// At this point we have JSON data in jsonData
	// For now, just print that we successfully read the data
	fmt.Printf("Successfully read %d bytes of JSON data\n", len(jsonData))

	// TODO: Validate JSON
	// TODO: Initialize UI
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: dive <json-file>\n")
	fmt.Fprintf(os.Stderr, "   or: cat <json-file> | dive\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "dive - Interactive JSON Viewer\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Provide a JSON file as an argument or pipe JSON data via stdin.\n")
}
