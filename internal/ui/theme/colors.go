package theme

import "github.com/gdamore/tcell/v2"

// Theme contains all color definitions used throughout the application
type Theme struct {
	// Border colors for different states
	BorderFocused   tcell.Color // Border color when component has focus
	BorderUnfocused tcell.Color // Border color when component doesn't have focus
	BorderValid     tcell.Color // Border color for valid input state
	BorderInvalid   tcell.Color // Border color for invalid input state

	// Background colors
	Background      tcell.Color // Default background color
	FieldBackground tcell.Color // Background for input fields

	// Text colors
	TextDefault     tcell.Color // Default text color
	TextPlaceholder tcell.Color // Placeholder text color
	TextAccent      tcell.Color // Accent text color (for headers/highlights)

	// Message colors
	ColorSuccess tcell.Color // Success message color
	ColorError   tcell.Color // Error message color

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
		BorderValid:     tcell.ColorGreen,         // Green for valid input
		BorderInvalid:   tcell.ColorRed,           // Red for invalid input

		// Background colors - use terminal defaults
		Background:      tcell.ColorDefault,
		FieldBackground: tcell.ColorDefault,

		// Text colors - use terminal defaults
		TextDefault:     tcell.ColorDefault,
		TextPlaceholder: tcell.ColorDefault,
		TextAccent:      tcell.ColorYellow, // Keep yellow for header branding

		// Message colors
		ColorSuccess: tcell.ColorGreen,
		ColorError:   tcell.ColorRed,

		// JSON syntax highlighting
		JSONKey:    "aqua",
		JSONString: "green",
		JSONNumber: "yellow",
		JSONBool:   "orange",
		JSONNull:   "gray",

		SuggestHighlight: "black:aqua",
	}
}

// GetBorderColor returns the appropriate border color based on focus state
func (t *Theme) GetBorderColor(focused bool) tcell.Color {
	if focused {
		return t.BorderFocused
	}
	return t.BorderUnfocused
}

// GetInputBorderColor returns the border color for input field considering focus and validation state
// Special states (valid/invalid) take precedence over focus state
func (t *Theme) GetInputBorderColor(focused bool, isValid bool, hasValidationState bool) tcell.Color {
	// If there's a validation state, it takes precedence
	if hasValidationState {
		if isValid {
			return t.BorderValid
		}
		return t.BorderInvalid
	}

	// Otherwise, use focus state
	return t.GetBorderColor(focused)
}
