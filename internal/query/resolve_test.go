package query

import "testing"

const doc = `{
	"users": [
		{"name": "Alice", "age": 28, "tags": ["a", "b"]},
		{"name": "Bob", "age": 35}
	],
	"company": {"name": "Acme", "founded": 2010}
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
