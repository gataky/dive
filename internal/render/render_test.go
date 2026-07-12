package render

import (
	"strings"
	"testing"

	"github.com/gataky/dive/internal/ui/theme"
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
	// tview.Escape turns [tag] into [tag[] so it renders literally
	if !strings.Contains(got, "[red[]") || !strings.Contains(got, "[-[]") {
		t.Errorf("document content not escaped:\n%s", got)
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
