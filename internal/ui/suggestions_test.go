package ui

import (
	"strings"
	"testing"

	"github.com/gataky/dive/internal/autocomplete"
	"github.com/gataky/dive/internal/ui/theme"
)

func suggs(names ...string) []autocomplete.Suggestion {
	var out []autocomplete.Suggestion
	for _, n := range names {
		out = append(out, autocomplete.Suggestion{Full: n, Display: n})
	}
	return out
}

func TestFormatSuggestionsEmpty(t *testing.T) {
	th := theme.DefaultTheme()
	if got := formatSuggestions(nil, -1, 80, th); got != "" {
		t.Errorf("empty input should give empty string, got %q", got)
	}
}

func TestFormatSuggestionsPlain(t *testing.T) {
	th := theme.DefaultTheme()
	got := formatSuggestions(suggs("name", "nationality"), -1, 80, th)
	if got != "name  nationality" {
		t.Errorf("got %q, want %q", got, "name  nationality")
	}
}

func TestFormatSuggestionsHighlight(t *testing.T) {
	th := theme.DefaultTheme()
	got := formatSuggestions(suggs("name", "nationality"), 1, 80, th)
	want := "[" + th.SuggestHighlight + "] nationality [-:-]"
	if !strings.Contains(got, want) {
		t.Errorf("selected candidate not highlighted: %q", got)
	}
	if strings.Contains(got, "] name [") {
		t.Errorf("unselected candidate should not be highlighted: %q", got)
	}
}

func TestFormatSuggestionsOverflow(t *testing.T) {
	th := theme.DefaultTheme()
	many := suggs("alpha", "bravo", "charlie", "delta", "echo", "foxtrot", "golf", "hotel")
	got := formatSuggestions(many, -1, 30, th)
	if !strings.Contains(got, "more") {
		t.Errorf("expected overflow marker in %q", got)
	}
	if strings.Contains(got, "hotel") {
		t.Errorf("overflowing candidate should not render: %q", got)
	}
}

func TestFormatSuggestionsSelectedAlwaysVisible(t *testing.T) {
	th := theme.DefaultTheme()
	many := suggs("alpha", "bravo", "charlie", "delta", "echo", "foxtrot", "golf", "hotel")
	got := formatSuggestions(many, 7, 30, th)
	if !strings.Contains(got, "hotel") {
		t.Errorf("selected candidate must be visible: %q", got)
	}
	if !strings.HasPrefix(got, "… ") {
		t.Errorf("shifted window should show leading ellipsis: %q", got)
	}
}

func TestFormatSuggestionsEscapesMarkup(t *testing.T) {
	th := theme.DefaultTheme()
	got := formatSuggestions(suggs("[:::key", "plain"), -1, 80, th)
	if strings.Contains(got, "[:::key") {
		t.Errorf("raw markup-dangerous display should not appear unescaped: %q", got)
	}
	if !strings.Contains(got, "[​:::key") {
		t.Errorf("display should be ZWSP-neutralized: %q", got)
	}
}
