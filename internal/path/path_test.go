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
		{"slice", "users.0:3", []Segment{{"users", KindKey}, {"0:3", KindAdvanced}}},
		{"pipe", "users.#.age|@reverse", []Segment{{"users", KindKey}, {"#", KindAdvanced}, {"age|@reverse", KindAdvanced}}},
		{"multipath", "{name,age}", []Segment{{"{name,age}", KindAdvanced}}},
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
	}
	for _, tt := range tests {
		if got := EscapeKey(tt.in); got != tt.want {
			t.Errorf("EscapeKey(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
