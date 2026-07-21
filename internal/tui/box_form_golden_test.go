package tui

// box_form_golden_test.go — golden snapshot for the Browse Primat-View with
// the experimental jira-style box-form detail render switched ON (S2b,
// BT_BOXFORM env gate, box_form_flag.go). Mirrors tree_golden_test.go's
// TestTreeGolden EXACTLY (same fixture via goldenTreeModel, same 100x30
// frame, same TrueColor/-update pattern) -- the ONLY difference is
// t.Setenv("BT_BOXFORM", "1") before rendering, so this test proves the
// flag's ON state without touching tree.golden (which stays the flag-OFF
// snapshot, byte-identical, per this slice's contract).

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// TestBrowseBoxFormGolden renders the SAME fixture TestTreeGolden uses, but
// with BT_BOXFORM=1, and compares it against testdata/browse_boxform.golden.
// Regenerate with: go test ./internal/tui/ -run TestBrowseBoxFormGolden -update
func TestBrowseBoxFormGolden(t *testing.T) {
	t.Setenv("BT_BOXFORM", "1")
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(termenv.Ascii)

	m := goldenTreeModel(t)
	out := m.View()

	if h := lipgloss.Height(out); h != 30 {
		t.Errorf("browse boxform view height=%d, want 30 (full terminal height)", h)
	}
	for i, ln := range strings.Split(out, "\n") {
		if w := lipgloss.Width(ln); w > 100 {
			t.Errorf("line %d overflows width (%d > 100): %q", i, w, ln)
		}
	}

	path := filepath.Join("testdata", "browse_boxform.golden")
	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(out), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("golden missing (%s) — regenerate with -update: %v", path, err)
	}
	if out != string(want) {
		t.Errorf("browse boxform view output differs from golden %q (frame/width/truncation?).\n--- got ---\n%s\n--- want ---\n%s", path, out, string(want))
	}
}

// TestBrowseBoxFormDefaultOffMatchesTree guards the flag's OFF-by-default
// contract at the golden level: WITHOUT BT_BOXFORM set, the exact same
// fixture must still render byte-identically to tree.golden -- i.e. this new
// code path is fully inert unless explicitly opted into.
func TestBrowseBoxFormDefaultOffMatchesTree(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(termenv.Ascii)

	m := goldenTreeModel(t)
	out := m.View()

	want, err := os.ReadFile(filepath.Join("testdata", "tree.golden"))
	if err != nil {
		t.Fatalf("tree.golden missing: %v", err)
	}
	if out != string(want) {
		t.Errorf("default (BT_BOXFORM unset) browse view diverged from tree.golden -- flag must be fully inert when unset")
	}
}
