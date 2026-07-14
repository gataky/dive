package theme

import "github.com/gdamore/tcell/v2"

// Theme contains all color definitions used throughout the application
type Theme struct {
	// Border colors for different states
	BorderFocused   tcell.Color // Border color when component has focus
	BorderUnfocused tcell.Color // Border color when component doesn't have focus
	BorderInvalid   tcell.Color // Border color for invalid input state

	// Background colors
	Background      tcell.Color // Default background color
	FieldBackground tcell.Color // Background for input fields

	// Text colors
	TextDefault tcell.Color // Default text color

	// JSON syntax highlighting (tview markup color names)
	JSONKey    string // object keys
	JSONString string // string values
	JSONNumber string // numbers
	JSONBool   string // true/false
	JSONNull   string // null

	// SuggestHighlight is the fg:bg tview markup for the selected suggestion.
	SuggestHighlight string
}

// DefaultTheme returns the default theme with terminal-friendly colors
func DefaultTheme() *Theme {
	return &Theme{
		// Border colors
		BorderFocused:   tcell.ColorDimGray,       // Subtle highlight for focused component
		BorderUnfocused: tcell.ColorDarkSlateGray, // Muted border for unfocused components
		BorderInvalid:   tcell.ColorRed,           // Red for invalid input

		// Background colors - use terminal defaults
		Background:      tcell.ColorDefault,
		FieldBackground: tcell.ColorDefault,

		// Text colors - use terminal defaults
		TextDefault: tcell.ColorDefault,

		// JSON syntax highlighting
		JSONKey:    "aqua",
		JSONString: "green",
		JSONNumber: "yellow",
		JSONBool:   "orange",
		JSONNull:   "gray",

		SuggestHighlight: "black:aqua",
	}
}
