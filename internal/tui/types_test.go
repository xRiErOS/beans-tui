package tui

// types_test.go — TDD coverage for types.go's own free-standing helpers
// (cloneBoolMap already has no dedicated unit test of its own -- every
// existing regression lives at its CALL sites instead, e.g.
// TestSetExpandedDoesNotMutateSharedMapAcrossModelCopies, update_test.go).
// cloneStringSlice (F01 History-Stack, E9 Task 8, bean bt-1vbp) gets a
// direct one here since it is a small, pure, independently-testable helper
// with no natural call-site test to piggyback on yet at RED time.

import "testing"

// TestCloneStringSlice guards the I01 Copy-on-Write convention for
// navBack/navForward's element type (types.go doc-stamp on
// fullscreen/navBack/navForward): the returned slice must carry the same
// contents as src, but NOT share its backing array -- mutating the clone
// must never leak back into src (mirrors cloneBoolMap's own rationale,
// applied to []string instead of map[string]bool).
func TestCloneStringSlice(t *testing.T) {
	src := []string{"a", "b", "c"}
	out := cloneStringSlice(src)

	if len(out) != len(src) {
		t.Fatalf("cloneStringSlice(%v) len = %d, want %d", src, len(out), len(src))
	}
	for i := range src {
		if out[i] != src[i] {
			t.Fatalf("cloneStringSlice(%v)[%d] = %q, want %q", src, i, out[i], src[i])
		}
	}

	out[0] = "mutated"
	if src[0] == "mutated" {
		t.Fatal("cloneStringSlice must not share src's backing array -- mutating the clone leaked into src")
	}
}

// TestCloneStringSliceNil guards the nil/empty edge -- append(clone(nil), x)
// is the exact pattern the History-Push uses on a fresh navBack/navForward
// (e.g. the very first Relations-Sprung in a session), so cloning nil must
// not panic and must yield a usable, zero-length slice.
func TestCloneStringSliceNil(t *testing.T) {
	out := cloneStringSlice(nil)
	if len(out) != 0 {
		t.Fatalf("cloneStringSlice(nil) len = %d, want 0", len(out))
	}
	out = append(out, "x")
	if len(out) != 1 || out[0] != "x" {
		t.Fatalf("cloneStringSlice(nil) result not appendable, got %v", out)
	}
}
