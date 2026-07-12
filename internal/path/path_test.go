package path

import (
	"reflect"
	"testing"
)

func TestSplit(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []Segment
	}{
		{"empty", "", nil},
		{"single key", "users", []Segment{{"users", KindKey}}},
		{"nested keys", "a.b.c", []Segment{{"a", KindKey}, {"b", KindKey}, {"c", KindKey}}},
		{"trailing dot yields empty segment", "users.", []Segment{{"users", KindKey}, {"", KindKey}}},
		{"array index", "users.0.name", []Segment{{"users", KindKey}, {"0", KindIndex}, {"name", KindKey}}},
		{"negative index", "users.-1", []Segment{{"users", KindKey}, {"-1", KindIndex}}},
		{"escaped dot stays in key", `fav\.movie.title`, []Segment{{`fav\.movie`, KindKey}, {"title", KindKey}}},
		{"count segment", "users.#", []Segment{{"users", KindKey}, {"#", KindAdvanced}}},
		{"query with inner dots", `users.#(a.b=="x")#.name`, []Segment{{"users", KindKey}, {`#(a.b=="x")#`, KindAdvanced}, {"name", KindKey}}},
		{"modifier", "users.@reverse", []Segment{{"users", KindKey}, {"@reverse", KindAdvanced}}},
		{"wildcard", "data.*.value", []Segment{{"data", KindKey}, {"*", KindAdvanced}, {"value", KindKey}}},
		{"colon key not special", "users.0:3", []Segment{{"users", KindKey}, {"0:3", KindKey}}},
		{"pipe", "users.#.age|@reverse", []Segment{{"users", KindKey}, {"#", KindAdvanced}, {"age|@reverse", KindAdvanced}}},
		{"multipath", "{name,age}", []Segment{{"{name,age}", KindAdvanced}}},
		{"unclosed query", `users.#(a.b=="x"`, []Segment{{"users", KindKey}, {`#(a.b=="x"`, KindAdvanced}}},
		{"trailing backslash", `trailing\`, []Segment{{`trailing\`, KindKey}}},
		{"unicode keys", "café.名前", []Segment{{"café", KindKey}, {"名前", KindKey}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Split(tt.in); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Split(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestJoin(t *testing.T) {
	paths := []string{"", "users", "a.b.c", `fav\.movie.title`, `users.#(a.b=="x")#.name`, "users."}
	for _, p := range paths {
		if got := Join(Split(p)); got != p {
			t.Errorf("Join(Split(%q)) = %q, want round-trip", p, got)
		}
	}
}

func TestJoinPrefix(t *testing.T) {
	segs := Split("a.b.c")
	if got := Join(segs[:2]); got != "a.b" {
		t.Errorf("Join(segs[:2]) = %q, want %q", got, "a.b")
	}
	if got := Join(segs[:0]); got != "" {
		t.Errorf("Join(segs[:0]) = %q, want empty", got)
	}
}

func TestEscapeKey(t *testing.T) {
	tests := []struct{ in, want string }{
		{"name", "name"},
		{"fav.movie", `fav\.movie`},
		{"a*b", `a\*b`},
		{"a?b", `a\?b`},
		{"a|b", `a\|b`},
		{`back\slash`, `back\\slash`},
		{"og:title", "og:title"},
		{"@type", `\@type`},
		{"#tag", `\#tag`},
		{"{brace", `\{brace`},
		{"[bracket", `\[bracket`},
		{"a(b)", "a(b)"},
	}
	for _, tt := range tests {
		if got := EscapeKey(tt.in); got != tt.want {
			t.Errorf("EscapeKey(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestFanoutRun(t *testing.T) {
	tests := []struct {
		name         string
		in           string
		wantRunLen   int
		wantAnchored bool
	}{
		{"empty", "", 0, false},
		{"single key", "users", 1, false},
		{"count tail", "users.#", 0, true},
		{"key after count", "users.#.name", 1, true},
		{"key and index after count", "users.#.name.0", 2, true},
		{"query tail", `users.#(age>20)#`, 0, true},
		{"plain keys and index only", "users.0.name", 3, false},
		{"modifier tail", "users.#.name.@reverse", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runLen, anchored := FanoutRun(Split(tt.in))
			if runLen != tt.wantRunLen || anchored != tt.wantAnchored {
				t.Errorf("FanoutRun(Split(%q)) = (%d, %v), want (%d, %v)",
					tt.in, runLen, anchored, tt.wantRunLen, tt.wantAnchored)
			}
		})
	}
}

// TestEscapeKeyClassifyRoundTrip asserts that any raw object key, once
// escaped, splits back into exactly one plain-key segment. Purely numeric
// keys ("0", "-1") are the exception: numeric object keys and array indices
// are indistinguishable in a path, so they classify as KindIndex.
func TestEscapeKeyClassifyRoundTrip(t *testing.T) {
	keys := []string{
		"name", "og:title", "@type", "#tag", "{brace", "[bracket",
		"a(b)", "fav.movie", "a*b", "a?b", "a|b", `back\slash`, "café",
		"0", "-1",
	}
	for _, k := range keys {
		segs := Split(EscapeKey(k))
		if len(segs) != 1 {
			t.Errorf("Split(EscapeKey(%q)) = %v, want exactly one segment", k, segs)
			continue
		}
		want := KindKey
		if k == "0" || k == "-1" {
			want = KindIndex
		}
		if segs[0].Kind != want {
			t.Errorf("Split(EscapeKey(%q))[0].Kind = %v, want %v", k, segs[0].Kind, want)
		}
	}
}
