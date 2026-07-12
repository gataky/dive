package query

import "testing"

const doc = `{
	"users": [
		{"name": "Alice", "age": 28, "tags": ["a", "b"]},
		{"name": "Bob", "age": 35}
	],
	"company": {"name": "Acme", "founded": 2010},
	"empty": [],
	"fav.movie": "Dune"
}`

func TestResolve(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		wantResolved string
		wantSuffix   string
		wantValue    string // Result.String() spot check, "" to skip
	}{
		{"empty path is whole doc", "", "", "", ""},
		{"full match", "company.name", "company.name", "", "Acme"},
		{"array index", "users.1.name", "users.1.name", "", "Bob"},
		{"typo on leaf", "company.nmae", "company", "nmae", ""},
		{"typo mid-path", "users.9.name", "users", "9.name", ""},
		{"nothing matches", "zzz.yyy", "", "zzz.yyy", ""},
		{"trailing dot not unmatched", "company.", "company", "", ""},
		{"bare trailing dot", ".", "", "", ""},
		{"count query", "users.#", "users.#", "", "2"},
		{"filter query", `users.#(age>30)#.name`, `users.#(age>30)#.name`, "", ""},
		{"typo after query", `users.#(age>30)#.nmae`, `users.#(age>30)#`, "nmae", ""},
		{"genuine empty array resolves", "empty", "empty", "", ""},
		{"fan-out typo over #", "users.#.nmae", "users.#", "nmae", ""},
		{"empty query result resolves", `users.#(age>90)#`, `users.#(age>90)#`, "", ""},
		{"deep fan-out typo", "users.#.name.nmae", "users.#.name", "nmae", ""},
		{"deep fan-out typo nested", "users.#.tags.zzz", "users.#.tags", "zzz", ""},
		{"deep fan-out typo through index", "users.#.tags.0.zzz", "users.#.tags.0", "zzz", ""},
		{"fan-out out-of-range index", "users.#.tags.9", "users.#.tags", "9", ""},
		{"escaped key resolves", `fav\.movie`, `fav\.movie`, "", "Dune"},
		{"escaped key typo", `fav\.novie`, "", `fav\.novie`, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Resolve(doc, tt.path)
			if got.ResolvedPath != tt.wantResolved {
				t.Errorf("ResolvedPath = %q, want %q", got.ResolvedPath, tt.wantResolved)
			}
			if got.UnmatchedSuffix != tt.wantSuffix {
				t.Errorf("UnmatchedSuffix = %q, want %q", got.UnmatchedSuffix, tt.wantSuffix)
			}
			if tt.wantValue != "" && got.Result.String() != tt.wantValue {
				t.Errorf("Result = %q, want %q", got.Result.String(), tt.wantValue)
			}
			if !got.Result.Exists() {
				t.Error("Result must always exist (never blank)")
			}
		})
	}
}

// TestResolveEmptyDocument pins the behavior for empty (invalid) input:
// the fallback Result does not exist. Resolve documents that data must be
// valid JSON; callers validate input at startup.
func TestResolveEmptyDocument(t *testing.T) {
	got := Resolve("", "foo")
	if got.ResolvedPath != "" {
		t.Errorf("ResolvedPath = %q, want %q", got.ResolvedPath, "")
	}
	if got.UnmatchedSuffix != "foo" {
		t.Errorf("UnmatchedSuffix = %q, want %q", got.UnmatchedSuffix, "foo")
	}
	if got.Result.Exists() {
		t.Error("Result.Exists() = true, want false for empty document")
	}
}
