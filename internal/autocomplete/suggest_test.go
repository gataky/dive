package autocomplete

import (
	"reflect"
	"testing"

	"github.com/tidwall/gjson"
)

const doc = `{
	"users": [
		{"name": "Alice", "nationality": "US", "age": 28},
		{"name": "Bob"}, {"name": "C"}, {"name": "D"}, {"name": "E"},
		{"name": "F"}, {"name": "G"}, {"name": "H"}, {"name": "I"},
		{"name": "J"}, {"name": "K"}, {"name": "L"}
	],
	"fav.movie": "Dune",
	"company": {"name": "Acme"}
}`

func fulls(s []Suggestion) []string {
	var out []string
	for _, x := range s {
		out = append(out, x.Full)
	}
	return out
}

func TestSuggestTopLevel(t *testing.T) {
	got := Suggest(doc, "")
	want := []string{"users", `fav\.movie`, "company"}
	if !reflect.DeepEqual(fulls(got), want) {
		t.Errorf("Suggest(\"\") = %v, want %v", fulls(got), want)
	}
}

func TestSuggestPrefixFilter(t *testing.T) {
	got := Suggest(doc, "users.0.na")
	want := []string{"users.0.name", "users.0.nationality"}
	if !reflect.DeepEqual(fulls(got), want) {
		t.Errorf("got %v, want %v", fulls(got), want)
	}
	if got[0].Display != "name" || got[1].Display != "nationality" {
		t.Errorf("Display labels = %q, %q; want key names", got[0].Display, got[1].Display)
	}
}

func TestSuggestTrailingDot(t *testing.T) {
	got := Suggest(doc, "company.")
	want := []string{"company.name"}
	if !reflect.DeepEqual(fulls(got), want) {
		t.Errorf("got %v, want %v", fulls(got), want)
	}
}

func TestSuggestArrayIndicesPast10(t *testing.T) {
	got := Suggest(doc, "users.")
	if len(got) != 12 {
		t.Fatalf("got %d suggestions, want 12", len(got))
	}
	if got[10].Display != "10" || got[11].Display != "11" {
		t.Errorf("indices past 9 wrong: %q, %q", got[10].Display, got[11].Display)
	}
	// index prefix filtering
	got = Suggest(doc, "users.1")
	want := []string{"users.1", "users.10", "users.11"}
	if !reflect.DeepEqual(fulls(got), want) {
		t.Errorf("got %v, want %v", fulls(got), want)
	}
}

func TestSuggestionsResolve(t *testing.T) {
	// every suggestion must be a path gjson can actually resolve
	for _, s := range Suggest(doc, "") {
		if !gjson.Get(doc, s.Full).Exists() {
			t.Errorf("suggestion %q does not resolve", s.Full)
		}
	}
}

func TestSuggestInvalidParent(t *testing.T) {
	if got := Suggest(doc, "zzz.a"); got != nil {
		t.Errorf("invalid parent should give nil, got %v", got)
	}
}

func TestSuggestAdvancedSegmentDegrades(t *testing.T) {
	if got := Suggest(doc, "users.#(age>2"); got != nil {
		t.Errorf("advanced partial should give nil, got %v", got)
	}
	// but a *parent* that is an advanced query still enumerates its result
	got := Suggest(doc, "users.#(age>20)#.")
	if len(got) == 0 {
		t.Error("expected index suggestions under advanced parent")
	}
}

func TestSuggestPrimitiveParent(t *testing.T) {
	if got := Suggest(doc, "company.name."); got != nil {
		t.Errorf("primitive parent should give nil, got %v", got)
	}
}
