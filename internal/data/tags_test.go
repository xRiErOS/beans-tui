package data

import "testing"

// TestValidTagName guards the tag grammar (^[a-z][a-z0-9]*(-[a-z0-9]+)*$,
// epic body bt-gzcu) against the exact valid/invalid table the plan
// specifies (epic-E3-plan.md »Task 2« Step 1).
func TestValidTagName(t *testing.T) {
	valid := []string{"a", "abc", "a1", "to-review", "a-b-c", "x9-y"}
	for _, s := range valid {
		if !ValidTagName(s) {
			t.Errorf("ValidTagName(%q) = false, want true", s)
		}
	}

	invalid := []string{"", "A", "1a", "-a", "a-", "a--b", "a_b", "über", "a b"}
	for _, s := range invalid {
		if ValidTagName(s) {
			t.Errorf("ValidTagName(%q) = true, want false", s)
		}
	}
}
