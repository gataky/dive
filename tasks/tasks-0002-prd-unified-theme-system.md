# Task List: Unified Theme System for Consistent Visual Style

Based on PRD: `0002-prd-unified-theme-system.md`

## Current State Assessment

The application currently has:
- Inconsistent border colors: Header (blue), Input (green/red), Dropdown (yellow), Output (aqua)
- No visual focus indication when navigating between components
- Hardcoded colors in component creation functions (`components.go`)
- Valid/invalid state handling exists but overrides hardcoded colors
- All backgrounds already use `tcell.ColorDefault`
- Focus management via tview exists but no visual feedback

## Relevant Files

- `internal/ui/theme/colors.go` - New theme package with color constants and Theme struct
- `internal/ui/components.go` - Update to accept and use theme colors
- `internal/ui/app.go` - Update to instantiate theme and manage focus state
- `internal/ui/app_test.go` - Unit tests for focus management (optional)

### Notes

- The theme package will centralize all color definitions
- Components will need to support dynamic border color updates for focus changes
- Special state colors (valid/invalid) take precedence over focus colors
- Terminal default (`tcell.ColorDefault`) already in use for backgrounds

## Tasks

- [x] 1.0 Create Theme Package and Color System
  - [x] 1.1 Create `internal/ui/theme` directory
  - [x] 1.2 Create `colors.go` with Theme struct containing all color fields
  - [x] 1.3 Define color constants: BorderFocused, BorderUnfocused, BorderValid, BorderInvalid, ColorAccent
  - [x] 1.4 Implement `DefaultTheme()` function returning theme with ColorDimGray for focused, muted colors for unfocused
  - [x] 1.5 Add helper methods to Theme struct for getting border colors based on focus/state

- [x] 2.0 Update Component Creation to Use Theme
  - [x] 2.1 Update `createHeader()` to accept theme parameter and use theme.BorderUnfocused
  - [x] 2.2 Update `createInputField()` to accept theme parameter and use theme.BorderUnfocused initially
  - [x] 2.3 Update `createOutputPanel()` to accept theme parameter and use theme.BorderUnfocused
  - [x] 2.4 Update `createAutocompleteDropdown()` to accept theme parameter and use theme.BorderUnfocused
  - [x] 2.5 Update `createFooter()` to use ColorDefault (no border, already correct)
  - [x] 2.6 Update `App.initComponents()` to pass theme to all component creation functions

- [x] 3.0 Implement Focus Tracking and Visual Indication
  - [x] 3.1 Add `theme` field and `focusedComponent` field to App struct
  - [x] 3.2 Initialize theme in `NewApp()` using `DefaultTheme()`
  - [x] 3.3 Create `setComponentFocus()` method to update border colors on focus change
  - [x] 3.4 Add focus handlers to input field to call `setComponentFocus()` when focused
  - [x] 3.5 Add focus handlers to autocomplete dropdown to update borders when focused
  - [x] 3.6 Add focus handlers to output panel to update borders when focused (for scrolling)
  - [x] 3.7 Update arrow key navigation to call `setComponentFocus()` when changing focus

- [x] 4.0 Integrate Theme with Special States
  - [x] 4.1 Update `setupQueryCallbacks()` to use theme.BorderValid instead of hardcoded ColorGreen
  - [x] 4.2 Update `setupQueryCallbacks()` to use theme.BorderInvalid instead of hardcoded ColorRed
  - [x] 4.3 Ensure valid/invalid state colors override focus colors in input field
  - [x] 4.4 Update `showMessage()` to use theme.BorderInvalid for error messages
  - [x] 4.5 Test that state colors take precedence over focus indication

- [x] 5.0 Update Modal Dialogs and Test Thoroughly
  - [x] 5.1 Update `showSaveDialog()` modal to use theme.BorderFocused instead of ColorYellow
  - [x] 5.2 Verify header keeps yellow accent color for branding (no change needed)
  - [ ] 5.3 Test focus indication in different terminal emulators (iTerm2, Terminal.app)
  - [ ] 5.4 Test with light and dark terminal themes
  - [ ] 5.5 Verify all components respond correctly to focus changes
  - [x] 5.6 Build and run full test suite to ensure no regressions
