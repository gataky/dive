# dive Rework — Design

**Date:** 2026-07-11
**Status:** Approved pending final spec review

## Purpose

dive is a TUI for exploring JSON: a query box where you type a gjson path, a
view that live-filters the document as you type, and next-level completions
that make the structure discoverable. A mistyped key must never blank the
view — the display always shows something anchored to what you typed, so the
mistake is easy to see.

The current implementation has the right concept but flawed execution:
autocomplete is Tab-triggered with focus-hopping into a dropdown, array-index
suggestions break at index 10 (`string(rune('0'+i))`), accepting a suggestion
dead-ends (no trailing continuation), and showing/hiding the dropdown rebuilds
the whole flex layout, causing visible jank. This rework replaces the
interaction and rendering while keeping the libraries (tview, tcell, gjson)
and the healthy packages (`input`, `export`, `theme`, help content).

## Decisions (from brainstorming)

- **Completions:** always-visible suggestions bar that updates live on every
  keystroke; no Tab needed to see them.
- **Tab cycling:** Tab cycles the last path segment through the candidates,
  with the originally typed text as hidden index 0. Example: `na` → Tab →
  `name` → Tab → `nationality` → Tab → back to `na`. Shift+Tab cycles
  backward.
- **Invalid path:** output shows the deepest valid ancestor of the typed path
  (e.g. `users.0.nmae` shows `users.0`), with the divergence made explicit.
- **gjson scope:** full gjson syntax passes through to the query (wildcards,
  `#`, `#(...)` queries, `@modifiers`); completions and ancestor resolution
  apply to plain key/index segments and gracefully back off on advanced
  segments.
- **Output:** JSON syntax highlighting using theme colors.
- **Kept features:** copy to clipboard (moved to Ctrl+Y), save to file
  (Ctrl+S), gjson help panel (F1), output-panel scrolling (Ctrl+O).
- **Approach:** rework the UI layer around a fixed layout and central state;
  keep module, libraries, and the `input`/`export`/`theme` packages.

## Layout

One vertical flex, created once at startup and never rebuilt:

```
┌ Query ────────────────────────────────────────┐
│ users.0.na▌                                   │   3 rows, bordered input
└───────────────────────────────────────────────┘
  name   nationality                                1 row suggestions bar
┌ users.0 ──────────────────────────────────────┐
│ {                                              │   flex-grow output panel,
│   "name": "Ada", ...                           │   title = resolved path
└───────────────────────────────────────────────┘
 Tab: Complete | F1: Help | Ctrl+Y: Copy | ...      1 row footer
```

- **Suggestions bar:** a single always-present row (TextView, not List).
  Display-only — focus never leaves the input field. Renders candidates
  horizontally; during a Tab cycle the selected candidate is highlighted.
  Overflow renders what fits plus a `… +N more` marker.
- **Output panel title** shows the path actually displayed (deepest valid
  ancestor). On a typo the title reads e.g. `users.0 ✗ nmae`.
- **Input border** turns red when the typed path has an unmatched suffix
  (tview's InputField cannot color a substring, so the border + title carry
  the signal), and returns to the normal/focused color when the full path
  matches.
- **Save dialog** is a `tview.Pages` overlay, not a root swap.
- **Help panel** (F1) is a horizontal split; visibility is toggled on
  structures built once — no `Clear()`-and-rebuild anywhere.

## Interaction

- Every keystroke runs one pipeline: suggest → resolve → render → update
  title/border/suggestions.
- **Tab cycle:** state = (base text, candidates, index). Tab advances index
  (mod len+1, index 0 = base text); Shift+Tab goes backward. Each state
  live-updates the output so the user sees each candidate's value. Typing any
  character or moving the cursor commits the currently shown text and ends
  the cycle. Escape during a cycle reverts to the base text.
- **Keys:** Tab / Shift+Tab complete-cycle · F1 help · Ctrl+O focus output
  (Esc returns to input) · Ctrl+Y copy · Ctrl+S save · Ctrl+C / Ctrl+Q quit.

## Architecture

Same module (`github.com/gataky/dive`); dependency versions bumped to latest
compatible releases.

- **`internal/path` (new)** — pure path logic. Split a gjson path into
  segments respecting `\.` escapes and advanced segments (`#`, `#(...)`,
  `@modifier`, `*`); join segments back with proper escaping; classify each
  segment as plain key, array index, or advanced.
- **`internal/query` (reworked)** — stateless. `Resolve(data, path)` returns
  `{ResolvedPath, UnmatchedSuffix, Result}`: try the full path via
  `gjson.Get`; while it doesn't exist, drop trailing segments until a prefix
  exists (empty prefix = whole document). No cached "last valid value" — the
  ancestor is recomputed from the document every keystroke. Full gjson syntax
  passes straight through.
- **`internal/autocomplete` (reworked)** — `Suggest(data, text)` splits input
  into parent path + partial last segment, enumerates keys at the parent
  (object keys; `strconv.Itoa` indices for arrays, fixing the ≥10 bug),
  prefix-filters by the partial, and returns candidates as full replacement
  texts with dots in keys escaped as `\.`. Returns nothing when the parent is
  invalid or ends in an advanced segment (graceful degrade). A pure
  `CycleState` type owns Tab-cycle transitions so it is unit-testable without
  a terminal.
- **`internal/render` (new)** — colorizer that walks the `gjson.Result` tree
  and emits pretty-printed JSON with tview color tags from the theme (keys,
  strings, numbers, booleans, null), escaping `[` as `[[` so document content
  cannot corrupt markup. Word wrap off; lines wider than the panel are
  clipped (tview's TextView does not scroll horizontally), which keeps the
  indentation structure readable.
- **`internal/ui` (thinned)** — wiring only: fixed layout construction, the
  `onTextChanged` pipeline, key handling, footer messages.
- **Unchanged:** `internal/input`, `internal/export`, `internal/ui/theme`,
  help content.

## Error handling

- Invalid JSON at startup: exit with a clear stderr message before the TUI
  starts.
- Path with unmatched suffix: deepest-valid-ancestor display as above; the
  view never blanks.
- No suggestions available: empty suggestions bar (no error).
- Clipboard/save failures: existing temporary footer message behavior.

## Testing

Table-driven unit tests:

- `path`: splitting with `\.` escapes, advanced segments, trailing dots;
  join/escape round-trips.
- `query.Resolve`: typos at each depth, array indices, advanced segments,
  empty path, full-path match.
- `autocomplete`: trailing dot, arrays with ≥10 elements, keys containing
  dots, prefix filtering, `CycleState` transitions (forward, backward, wrap,
  commit, revert).
- `render`: colorization of each JSON type, `[` escaping.

Existing `input`/`export` tests are kept. Manual TUI smoke test against
`test.json`.
