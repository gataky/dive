package ui

import (
	"fmt"
	"strings"

	"github.com/gataky/dive/internal/autocomplete"
	"github.com/gataky/dive/internal/render"
	"github.com/gataky/dive/internal/ui/theme"
	"github.com/mattn/go-runewidth"
)

// formatSuggestions renders the suggestions bar: candidates separated by
// two spaces, the selected one highlighted, windowed so the selection is
// always visible, with "… +N more" when candidates overflow the width.
// selected must be -1 (no selection) or in [0, len(suggs)); out-of-range
// values render a degraded (near-empty) bar.
func formatSuggestions(suggs []autocomplete.Suggestion, selected, width int, th *theme.Theme) string {
	if len(suggs) == 0 {
		return ""
	}
	start := 0
	if selected >= 0 && fitCount(suggs, 0, width) <= selected {
		start = selected
	}
	n := fitCount(suggs, start, width)

	var b strings.Builder
	if start > 0 {
		b.WriteString("… ")
	}
	for i := start; i < start+n && i < len(suggs); i++ {
		if i > start {
			b.WriteString("  ")
		}
		if i == selected {
			b.WriteString("[" + th.SuggestHighlight + "] " + render.EscapeContent(suggs[i].Display) + " [-:-]")
		} else {
			b.WriteString(render.EscapeContent(suggs[i].Display))
		}
	}
	if rest := len(suggs) - (start + n); rest > 0 {
		fmt.Fprintf(&b, "  … +%d more", rest)
	}
	return b.String()
}

// fitCount returns how many suggestions starting at start fit in width
// display cells, reserving 12 cells for the overflow markers plus 2 for
// the selected candidate's highlight padding. The reservation assumes the
// "+N more" count is at most 2 digits; longer counts clip gracefully.
// Always >= 1 so at least the selected candidate renders.
func fitCount(suggs []autocomplete.Suggestion, start, width int) int {
	budget := width - 14
	used, n := 0, 0
	for i := start; i < len(suggs); i++ {
		cells := runewidth.StringWidth(suggs[i].Display)
		if n > 0 {
			cells += 2
		}
		if used+cells > budget && n > 0 {
			break
		}
		used += cells
		n++
	}
	return n
}
