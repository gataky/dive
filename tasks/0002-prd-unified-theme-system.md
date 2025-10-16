# PRD: Unified Theme System for Consistent Visual Style

## Introduction/Overview

Currently, the `dive` application has inconsistent visual styling across its components. Border colors vary (green input, yellow dropdown, blue header, aqua output), backgrounds are handled inconsistently, and there's no clear visual indication when navigating between focusable components. This PRD addresses the need for a unified, consistent visual style that uses terminal defaults as the base, making the application feel like a natural extension of the terminal while maintaining professional polish and clear focus indication.

**Problem:** Inconsistent colors and lack of focus indication make the interface feel unpolished and can confuse users about which component currently has focus.

**Goal:** Create a centralized theme system with consistent visual styling across all components, using terminal defaults for colors to maintain terminal integration, while providing clear focus indication through border color changes.

## Goals

1. **Visual Consistency:** All components use consistent, coordinated colors from a centralized theme
2. **Terminal Integration:** Default colors match terminal defaults (ColorDefault) so the app feels integrated
3. **Clear Focus Indication:** Users can immediately identify which component has focus via border color
4. **Maintainability:** Colors defined in one place, easy to modify
5. **Professional Polish:** The application looks cohesive and well-designed

## User Stories

1. **As a user**, I want the application to blend seamlessly with my terminal so that it doesn't feel jarring or out of place when I launch it.

2. **As a user**, I want to clearly see which component has focus so that I know where my keyboard input will go.

3. **As a keyboard-only user**, I want visual feedback when I tab between components so that I can navigate confidently without using a mouse.

4. **As a developer**, I want all color definitions in one place so that I can easily adjust the visual style without hunting through multiple files.

5. **As a user with a custom terminal theme**, I want the application to respect my terminal's color scheme so that it matches my workflow aesthetic.

## Functional Requirements

### 1. Theme Package Structure

1.1. Create a new package `internal/ui/theme` with a `colors.go` file
1.2. Define a `Theme` struct containing all color values used in the application
1.3. Theme must include colors for:
   - Border colors (focused and unfocused states)
   - Background colors for all components
   - Text colors
   - Placeholder text colors
   - Header/footer styling colors

### 2. Default Color Scheme

2.1. All default colors must use `tcell.ColorDefault` to respect terminal defaults
2.2. Exception: Focus indication border color should be distinct (e.g., a subtle highlight color that works in most terminal themes)
2.3. Unfocused borders should use a muted color that doesn't compete visually
2.4. Current state indicators (valid/invalid path) should override focus colors when needed

### 3. Component Consistency

3.1. **Input Field:**
   - Unfocused: Muted border color, default background, default text color
   - Focused: Distinct border color to indicate active state
   - Invalid path: Red border (overrides focus color)
   - Valid path: Green border (overrides focus color) when focused, muted when unfocused

3.2. **Autocomplete Dropdown:**
   - Unfocused: Same muted border as other unfocused components
   - Focused: Same focus border color as input field
   - Background: Terminal default
   - Selected item: Use terminal's selection/highlight colors

3.3. **Output Panel:**
   - Unfocused: Same muted border as other components
   - Focused: Same focus border color (for scrolling)
   - Background: Terminal default
   - Text: Terminal default color

3.4. **Header:**
   - Border: Muted color (never focused)
   - Background: Terminal default
   - Text: Can use accent colors for title/branding

3.5. **Footer:**
   - No border
   - Background: Terminal default
   - Text: Use subtle colors for keybinding hints

### 4. Focus Management Integration

4.1. Components must update their border color when receiving focus
4.2. Components must revert to unfocused border color when losing focus
4.3. Special states (invalid/valid path) must take precedence over focus colors
4.4. When modal dialogs are open (save dialog), they should use focus colors

### 5. Color Constants

5.1. Define named constants for semantic colors:
   - `ColorFocused` - Border color for focused component
   - `ColorUnfocused` - Border color for unfocused components
   - `ColorValid` - Border color for valid input
   - `ColorInvalid` - Border color for invalid input
   - `ColorAccent` - For headers/highlights
   - `ColorDefault` - Terminal default for backgrounds/text

5.2. All color assignments in components must use these constants

### 6. Component Refactoring

6.1. Update `internal/ui/components.go` to use theme colors
6.2. Update `internal/ui/app.go` to apply theme colors to all components
6.3. Add theme instance to `App` struct
6.4. Wire up focus change handlers to update border colors

### 7. Focus Handling

7.1. Track currently focused component in `App` struct
7.2. When focus changes:
   - Remove focus styling from previously focused component
   - Apply focus styling to newly focused component
7.3. Components that can receive focus: Input field, autocomplete dropdown, output panel
7.4. Save dialog modal should use focus colors when displayed

## Non-Goals (Out of Scope)

1. **User Configuration:** No config file or command-line flags for theme customization (just developer-adjustable constants)
2. **Multiple Themes:** No built-in theme variants (dark, light, etc.)
3. **Runtime Theme Switching:** Colors set at initialization only
4. **Color Picker UI:** No interactive color selection
5. **Syntax Highlighting Colors:** JSON output syntax highlighting is separate concern
6. **Per-Component Customization:** All components follow the unified theme

## Design Considerations

### Current State
- Input field: Green border (changes to red when invalid)
- Autocomplete dropdown: Yellow border
- Header: Blue border
- Output panel: Aqua border
- Footer: No border
- No visual distinction between focused/unfocused states

### Proposed State
- All unfocused borders: Subtle gray/dim color
- Focused component: Brighter/accent border color
- Input field special states: Red (invalid), Green (valid) - override focus color
- All backgrounds: Terminal default (ColorDefault)
- All text: Terminal default unless semantic meaning (errors, accents)

### Visual Hierarchy
1. Focus indication is primary (which component is active)
2. State indication is secondary (valid/invalid path)
3. Component separation is tertiary (borders just for clarity)

## Technical Considerations

### Implementation Approach

1. **Create Theme Package:**
   ```go
   package theme

   type Theme struct {
       BorderFocused   tcell.Color
       BorderUnfocused tcell.Color
       BorderValid     tcell.Color
       BorderInvalid   tcell.Color
       // ... other colors
   }

   func DefaultTheme() *Theme {
       return &Theme{
           BorderFocused:   tcell.ColorDimGray,
           BorderUnfocused: tcell.ColorDarkSlateGray,
           // ...
       }
   }
   ```

2. **Component Updates:**
   - Pass theme instance to component creation functions
   - Components store reference to theme
   - Implement `SetFocused(bool)` methods to update styling

3. **Focus Tracking:**
   - App struct tracks current focused component
   - Focus change handler updates both old and new components
   - Special handling for modal overlays

### Dependencies
- Existing tcell color system
- Current tview component structure
- Existing focus management in tview

### Testing Strategy
- Manual testing with different terminal emulators
- Visual verification of focus changes
- Test with light and dark terminal themes
- Verify ColorDefault respects terminal settings

## Success Metrics

1. **Consistency:** All components use colors from the theme package (no hardcoded colors in components)
2. **Focus Clarity:** Users can identify focused component within 1 second of looking at the screen
3. **Terminal Integration:** Application colors match terminal colors when using defaults
4. **Code Quality:** All color constants defined in one location (`internal/ui/theme/colors.go`)
5. **Polish:** Application receives positive feedback on visual consistency in user testing

## Decisions

1. **Focus indication:** Border color only (no background changes)
   - **Decision:** Use `tcell.ColorDimGray` for focused component borders

2. **Save modal dialog:** Adopt the unified theme
   - **Decision:** Use `ColorFocused` for consistency with other components

3. **Header accent colors:** Keep existing yellow title for branding/personality
   - **Decision:** Maintain current header styling with yellow accent

4. **Error message colors:** Use consistent color system
   - **Decision:** Error messages in footer use `ColorInvalid` (same red as invalid paths)

## Implementation Notes

### File Structure
```
internal/ui/
├── theme/
│   ├── colors.go      # Theme struct and color constants
│   └── theme_test.go  # (optional) Theme tests
├── app.go             # Updated to use theme
└── components.go      # Updated to use theme
```

### Migration Strategy
1. Create theme package with all color constants
2. Update component creation functions to accept theme
3. Update App to instantiate and use theme
4. Add focus tracking to App
5. Implement focus change handlers
6. Test with various terminal themes
7. Remove hardcoded colors from all components

### Validation
- Run application in different terminal emulators (iTerm2, Terminal.app, Alacritty, etc.)
- Test with light and dark terminal themes
- Verify focus indication is clear
- Confirm terminal default colors are respected
