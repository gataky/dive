package render

import (
	"strings"
	"testing"

	"github.com/gataky/dive/internal/ui/theme"
	"github.com/rivo/tview"
	"github.com/tidwall/gjson"
)

func TestPrettyPreservesKeyOrder(t *testing.T) {
	r := gjson.Parse(`{"zebra":1,"apple":2}`)
	got := Pretty(r)
	if strings.Index(got, "zebra") > strings.Index(got, "apple") {
		t.Errorf("Pretty reordered keys:\n%s", got)
	}
	if !strings.Contains(got, "  \"zebra\"") {
		t.Errorf("Pretty output not indented:\n%s", got)
	}
}

func TestPrettyPrimitive(t *testing.T) {
	r := gjson.Get(`{"name":"Ada"}`, "name")
	if got := Pretty(r); got != `"Ada"` {
		t.Errorf("Pretty(string) = %q, want %q", got, `"Ada"`)
	}
}

func TestColorizeTypes(t *testing.T) {
	th := theme.DefaultTheme()
	r := gjson.Parse(`{"s":"x","n":1.5,"b":true,"z":null,"arr":[1],"obj":{"k":2}}`)
	got := Colorize(r, th)

	for _, want := range []string{
		"[" + th.JSONKey + `]"s"[-]`,
		"[" + th.JSONString + `]"x"[-]`,
		"[" + th.JSONNumber + "]1.5[-]",
		"[" + th.JSONBool + "]true[-]",
		"[" + th.JSONNull + "]null[-]",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("Colorize missing %q in:\n%s", want, got)
		}
	}
	if !strings.Contains(got, "{") || !strings.Contains(got, "[\n") {
		t.Errorf("Colorize missing structural punctuation:\n%s", got)
	}
}

func TestColorizeEscapesMarkupInContent(t *testing.T) {
	th := theme.DefaultTheme()
	r := gjson.Parse(`{"note":"[red]danger[-]"}`)
	got := Colorize(r, th)
	// A zero-width space (U+200B) after "[" makes it an invalid tag start,
	// so the content renders literally instead of being parsed as markup.
	if !strings.Contains(got, "[​red]danger[​-]") {
		t.Errorf("document content not neutralized with ZWSP:\n%s", got)
	}
	if strings.Contains(got, "[red]danger") {
		t.Errorf("raw markup leaked into output:\n%s", got)
	}
}

// drawn runs markup through tview's real dynamic-color parser (the same one
// used when drawing) and returns the visible text, with the zero-width
// spaces we inject stripped so assertions compare plain content.
func drawn(markup string) string {
	tv := tview.NewTextView().SetDynamicColors(true)
	tv.SetText(markup)
	return strings.ReplaceAll(tv.GetText(true), "​", "")
}

// TestColorizeDrawnFaithfully is a regression test for unclosed tview tags:
// a "[:::" URL-tag prefix in document content would otherwise open a
// consume-until-"]" state and swallow everything up to our own next "]".
func TestColorizeDrawnFaithfully(t *testing.T) {
	th := theme.DefaultTheme()
	docs := []string{
		`{"a":"[:::HIDDEN","b":"visible"}`,
		`{"[:::key":"v"}`,
		`{"trail":"end["}`,
		`{"tag":"[red]x[-]"}`,
	}
	for _, doc := range docs {
		got := drawn(Colorize(gjson.Parse(doc), th))
		for _, want := range []string{"HIDDEN", "visible", "[:::", "end[", "[red]x[-]"} {
			if strings.Contains(doc, want) && !strings.Contains(got, want) {
				t.Errorf("doc %s: drawn output lost %q:\n%s", doc, want, got)
			}
		}
	}
}

func TestColorizeDeepNestingIndentation(t *testing.T) {
	th := theme.DefaultTheme()
	got := Colorize(gjson.Parse(`{"a":{"b":{"c":[1]}}}`), th)
	want := "{\n" +
		"  [" + th.JSONKey + `]"a"[-]: {` + "\n" +
		"    [" + th.JSONKey + `]"b"[-]: {` + "\n" +
		"      [" + th.JSONKey + `]"c"[-]: [` + "\n" +
		"        [" + th.JSONNumber + "]1[-]\n" +
		"      ]\n" +
		"    }\n" +
		"  }\n" +
		"}"
	if got != want {
		t.Errorf("deep nesting output mismatch\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestColorizeUnicode(t *testing.T) {
	th := theme.DefaultTheme()
	got := Colorize(gjson.Parse(`{"名前":"café"}`), th)
	if !strings.Contains(got, `"名前"`) {
		t.Errorf("unicode key missing from output:\n%s", got)
	}
	if !strings.Contains(got, `"café"`) {
		t.Errorf("unicode value missing from output:\n%s", got)
	}
}

func TestColorizeEmptyContainers(t *testing.T) {
	th := theme.DefaultTheme()
	if got := Colorize(gjson.Parse(`{}`), th); got != "{}" {
		t.Errorf("empty object = %q, want {}", got)
	}
	if got := Colorize(gjson.Parse(`[]`), th); got != "[]" {
		t.Errorf("empty array = %q, want []", got)
	}
}
