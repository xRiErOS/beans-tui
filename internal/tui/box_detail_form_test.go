package tui

// box_detail_form_test.go — S2 (jira-style experiment, design-spec.md D01/
// D03, mockup §6.3): tests for detailBoxForm, the jira-style scalar-fields
// grid built from S1's dropdownBox primitive. Structural (width-agnostic
// content) + responsive (perRow column count) + golden snapshot, mirroring
// tree_golden_test.go's own -update flag pattern (chrome_test.go's shared
// package-level `update` flag).

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xRiErOS/beans-tui/internal/data"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
)

// detailBoxFormFixture is the fixed bean fixture shared by all three tests
// below (structural/responsive/golden) — same ID/field shape as tree_golden_
// test.go's gld-tsk1 (Parent="gld-epic") so a future cross-golden comparison
// stays apples-to-apples, but Tags nil here (Parent test coverage already
// exercises tags via the tree golden; this fixture instead covers the "no
// tags" Dim-placeholder path of detailBoxForm's own Tags box).
func detailBoxFormFixture() *data.Bean {
	return &data.Bean{
		ID:       "gld-tsk1",
		Title:    "First golden task",
		Status:   "todo",
		Type:     "task",
		Priority: "",
		Parent:   "gld-epic",
		Tags:     nil,
	}
}

// TestDetailBoxFormStructure asserts (width-agnostic, ansi.Strip'd) that all
// 5 scalar fields plus Title render with their labels/values/hotkey badges,
// and that no produced line ever overflows the requested width (100).
func TestDetailBoxFormStructure(t *testing.T) {
	b := detailBoxFormFixture()
	out := detailBoxForm(b, 100)
	plain := ansi.Strip(out)

	for _, want := range []string{
		"Title", "First golden task",
		"Status", "todo",
		"Type", "task",
		"Priority",
		"Parent", "gld-epic",
		"Tags",
		"(e)", "(s)", "(o)", "(u)", "(a)", "(t)",
	} {
		if !strings.Contains(plain, want) {
			t.Errorf("output missing %q\n--- got ---\n%s", want, plain)
		}
	}

	for i, ln := range strings.Split(out, "\n") {
		if w := lipgloss.Width(ln); w > 100 {
			t.Errorf("line %d overflows width (%d > 100): %q", i, w, ln)
		}
	}
}

// TestDetailBoxFormResponsiveNarrow asserts the perRow=1 narrow-width path
// (width 50 < 64): every scalar box lands on its own row, so the rendered
// output has strictly more lines than the wide (perRow=3, width 100) layout,
// and no line exceeds the narrower width.
func TestDetailBoxFormResponsiveNarrow(t *testing.T) {
	b := detailBoxFormFixture()
	wide := detailBoxForm(b, 100)
	narrow := detailBoxForm(b, 50)

	wideLines := strings.Count(wide, "\n") + 1
	narrowLines := strings.Count(narrow, "\n") + 1
	if narrowLines <= wideLines {
		t.Errorf("narrow (width 50) layout has %d lines, want more than wide (width 100)'s %d lines", narrowLines, wideLines)
	}

	for i, ln := range strings.Split(narrow, "\n") {
		if w := lipgloss.Width(ln); w > 50 {
			t.Errorf("line %d overflows width (%d > 50): %q", i, w, ln)
		}
	}
}

// TestDetailBoxFormGolden renders detailBoxForm(fixture, 100) at TrueColor
// and compares it against testdata/detail_boxform.golden. Mirrors tree_
// golden_test.go's pattern exactly (shared `update` flag from chrome_test.go).
// Regenerate with: go test ./internal/tui/ -run TestDetailBoxFormGolden -update
func TestDetailBoxFormGolden(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(termenv.Ascii)

	b := detailBoxFormFixture()
	out := detailBoxForm(b, 100)

	path := filepath.Join("testdata", "detail_boxform.golden")
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
		t.Errorf("detailBoxForm output differs from golden %q.\n--- got ---\n%s\n--- want ---\n%s", path, out, string(want))
	}
}
