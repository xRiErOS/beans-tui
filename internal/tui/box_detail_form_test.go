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

// detailBoxFormIndex builds the *data.Index fixture detailBoxForm needs to
// resolve Relations (S2c) — a minimal parent/child pair around the fixture
// bean's own ID/Parent ("gld-tsk1" parented under "gld-epic"), mirroring
// view_detail_bean_test.go's own data.NewIndex(beans) construction pattern.
func detailBoxFormIndex() *data.Index {
	beans := []data.Bean{
		{ID: "gld-epic", Title: "Golden epic", Status: "todo", Type: "epic", Priority: "normal"},
		{ID: "gld-tsk1", Title: "First golden task", Status: "todo", Type: "task", Priority: "", Parent: "gld-epic"},
	}
	return data.NewIndex(beans)
}

// TestDetailBoxFormStructure asserts (width-agnostic, ansi.Strip'd) that all
// 5 scalar fields plus Title render with their labels/values/hotkey badges,
// and that EVERY produced line is exactly the requested width (100) --
// review B5/B1: overflow-only (`> width`) checks miss underflow, so this
// asserts `== width`.
func TestDetailBoxFormStructure(t *testing.T) {
	b := detailBoxFormFixture()
	idx := detailBoxFormIndex()
	const w = 100
	out := detailBoxForm(idx, b, w, -1)
	plain := ansi.Strip(out)

	for _, want := range []string{
		"Title", "First golden task",
		"Status", "todo",
		"Type", "task",
		"Priority",
		"Parent", "gld-epic",
		"Tags",
		"(e)", "(s)", "(o)", "(u)", "(a)", "(t)",
		"Body", "Relations", "History",
	} {
		if !strings.Contains(plain, want) {
			t.Errorf("output missing %q\n--- got ---\n%s", want, plain)
		}
	}

	for i, ln := range strings.Split(out, "\n") {
		if got := lipgloss.Width(ln); got != w {
			t.Errorf("line %d width = %d, want exactly %d: %q", i, got, w, ln)
		}
	}
}

// TestDetailBoxFormFixedGridNoCollapse asserts D12: the scalar grid is FIXED
// (Row A = Status|Type|Priority, Row B = Parent|Tags) at ANY width, no
// responsive perRow collapse to 1-up. At a narrow width (50) Row A must
// still render 3 columns (3 "▾" arrows / the (s)(o)(u) hotkeys on adjacent
// columns) and Row B must still render 2 columns ((a)(t) hotkeys) -- and
// every produced line is still exactly the requested width (review B1/B5).
func TestDetailBoxFormFixedGridNoCollapse(t *testing.T) {
	b := detailBoxFormFixture()
	idx := detailBoxFormIndex()
	const w = 50
	out := detailBoxForm(idx, b, w, -1)
	lines := strings.Split(out, "\n")

	for i, ln := range lines {
		if got := lipgloss.Width(ln); got != w {
			t.Errorf("line %d width = %d, want exactly %d: %q", i, got, w, ln)
		}
	}

	// Layout is fixed: Title (3 lines: top/mid/bot), then Row A (3 lines),
	// then Row B (3 lines), then the full-width panels.
	if len(lines) < 9 {
		t.Fatalf("want at least 9 lines (Title+RowA+RowB), got %d: %q", len(lines), out)
	}
	rowAMid := ansi.Strip(lines[4])
	rowABot := ansi.Strip(lines[5])
	if n := strings.Count(rowAMid, "▾"); n != 3 {
		t.Errorf("Row A mid line must have 3 ▾ arrows (no 1-up collapse), got %d: %q", n, rowAMid)
	}
	for _, hk := range []string{"(s)", "(o)", "(u)"} {
		if !strings.Contains(rowABot, hk) {
			t.Errorf("Row A bottom border missing hotkey %q (no 1-up collapse): %q", hk, rowABot)
		}
	}

	rowBMid := ansi.Strip(lines[7])
	rowBBot := ansi.Strip(lines[8])
	if n := strings.Count(rowBMid, "▾"); n != 2 {
		t.Errorf("Row B mid line must have 2 ▾ arrows (no 1-up collapse), got %d: %q", n, rowBMid)
	}
	for _, hk := range []string{"(a)", "(t)"} {
		if !strings.Contains(rowBBot, hk) {
			t.Errorf("Row B bottom border missing hotkey %q (no 1-up collapse): %q", hk, rowBBot)
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
	idx := detailBoxFormIndex()
	out := detailBoxForm(idx, b, 100, -1)

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
