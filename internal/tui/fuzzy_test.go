package tui

// fuzzy_test.go — TDD coverage for fuzzyMatch (E4 Task 1, bean bt-jpgn,
// design decision a): verbatim port of devd overlay_palette.go's own
// fuzzyMatch, pinned here against the exact cases the plan's Step 1
// specifies.

import "testing"

func TestFuzzyMatchSubsequenceCaseInsensitive(t *testing.T) {
	cases := []struct {
		query, target string
		want          bool
	}{
		{"", "anything", true},
		{"crb", "create: bean", true},
		{"CRB", "create: bean", true}, // case-insensitiv
		{"xyz", "create: bean", false},
		{"tsk1", "tsk1", true},
	}
	for _, c := range cases {
		if got := fuzzyMatch(c.query, c.target); got != c.want {
			t.Errorf("fuzzyMatch(%q, %q) = %v, want %v", c.query, c.target, got, c.want)
		}
	}
}
