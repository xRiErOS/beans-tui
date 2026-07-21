package tui

// line_window_test.go — bean bt-f68z, PO-Befund #13. Guards the two invariants
// the wrapping change hangs on: the hanging indent itself, and the fact that a
// LINE-counting window still bounds height once rows are taller than one line
// (the parentPickerRowBudget mistake, recorded in the bean).

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestHangingWrapIndentsContinuationToTitleColumn(t *testing.T) {
	head := "  ▸ i B eq67 "
	lines := hangingWrap(head, lipgloss.Width(head), "God-Delete/Purge crasht ORM nullt child FK statt Cascade", 44)
	if len(lines) < 2 {
		t.Fatalf("expected the title to wrap, got %d line(s): %q", len(lines), lines)
	}
	if !strings.HasPrefix(lines[0], head) {
		t.Errorf("first line lost its head: %q", lines[0])
	}
	pad := strings.Repeat(" ", lipgloss.Width(head))
	for i, l := range lines[1:] {
		if !strings.HasPrefix(l, pad) {
			t.Errorf("continuation line %d is not hanging-indented to the title column: %q", i+1, l)
		}
		if strings.HasPrefix(l, pad+" ") {
			t.Errorf("continuation line %d is over-indented (leading space in the title column): %q", i+1, l)
		}
	}
}

func TestHangingWrapNeverExceedsWidth(t *testing.T) {
	head := "  ▾ t M elir "
	for _, w := range []int{20, 28, 31, 40, 60, 79} {
		lines := hangingWrap(head, lipgloss.Width(head), "Canary-Instanz sproutling-test — Staged-Rollout & Dogfood auf NAS", w)
		for i, l := range lines {
			if got := lipgloss.Width(l); got > w {
				t.Errorf("width=%d line %d is %d cells wide: %q", w, i, got, l)
			}
		}
	}
}

func TestHangingWrapCapsLineCount(t *testing.T) {
	head := "x "
	long := strings.Repeat("wort ", 200)
	lines := hangingWrap(head, lipgloss.Width(head), long, 30)
	if len(lines) > maxRowLines {
		t.Fatalf("a single row grew to %d lines, cap is %d — one bean must never crowd out the pane", len(lines), maxRowLines)
	}
	if !strings.Contains(lines[len(lines)-1], "…") {
		t.Errorf("capped row must stay VISIBLY truncated, last line: %q", lines[len(lines)-1])
	}
}

func TestHangingWrapFallsBackToSingleLineWhenTooNarrow(t *testing.T) {
	head := "  ▸ i B eq67 "
	lines := hangingWrap(head, lipgloss.Width(head), "some title here", 16)
	if len(lines) != 1 {
		t.Fatalf("below the minimum wrap budget the row must stay single-line, got %d: %q", len(lines), lines)
	}
	if lipgloss.Width(lines[0]) > 16 {
		t.Errorf("fallback line overflows: %q", lines[0])
	}
}

// blocks builds n elements of the given heights, each line tagged with its
// element index so the mapping can be asserted from the row text itself.
func testBlocks(heights ...int) [][]string {
	out := make([][]string, len(heights))
	for i, h := range heights {
		lines := make([]string, h)
		for j := range lines {
			lines[j] = string(rune('A'+i)) + string(rune('0'+j))
		}
		out[i] = lines
	}
	return out
}

func TestBlockWindowNeverExceedsHeight(t *testing.T) {
	// The parentPickerRowBudget regression in list form: 40 elements, many of
	// them multi-line. An ELEMENT cap would happily emit 3x the lines.
	heights := make([]int, 40)
	for i := range heights {
		heights[i] = 1 + i%3
	}
	b := testBlocks(heights...)
	for _, h := range []int{1, 2, 5, 12, 18, 23} {
		for _, cur := range []int{0, 1, 17, 38, 39} {
			rows, elems := blockWindow(b, h, cur)
			if len(rows) > h {
				t.Fatalf("height=%d cursor=%d produced %d lines — a line window MUST bound height", h, cur, len(rows))
			}
			if len(rows) != len(elems) {
				t.Fatalf("rows/elem mapping out of sync: %d vs %d", len(rows), len(elems))
			}
		}
	}
}

func TestBlockWindowKeepsCursorVisible(t *testing.T) {
	heights := make([]int, 40)
	for i := range heights {
		heights[i] = 1 + i%3
	}
	b := testBlocks(heights...)
	for _, h := range []int{4, 9, 18, 23} {
		for cur := 0; cur < len(b); cur++ {
			_, elems := blockWindow(b, h, cur)
			found := false
			for _, e := range elems {
				if e == cur {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("height=%d cursor=%d: the cursor element is not in the window %v", h, cur, elems)
			}
		}
	}
}

func TestBlockWindowMappingMatchesRenderedRows(t *testing.T) {
	// The mouse-hit contract: rowElem[i] must name the element row i was
	// rendered FROM, for every row, at every cursor position.
	b := testBlocks(1, 3, 1, 2, 3, 1, 2)
	rows, elems := blockWindow(b, 8, 4)
	for i, r := range rows {
		wantPrefix := string(rune('A' + elems[i]))
		if !strings.HasPrefix(r, wantPrefix) {
			t.Errorf("row %d %q was mapped to element %d (%q) — mapping is wrong", i, r, elems[i], wantPrefix)
		}
	}
	if _, ok := blockWindowElemAt(elems, len(rows)); ok {
		t.Errorf("a row past the last rendered line must not resolve to an element")
	}
	if _, ok := blockWindowElemAt(elems, -1); ok {
		t.Errorf("a negative row must not resolve to an element")
	}
}

func TestBlockWindowShowsEverythingWhenItFits(t *testing.T) {
	b := testBlocks(1, 2, 1)
	rows, elems := blockWindow(b, 20, 1)
	if len(rows) != 4 {
		t.Fatalf("expected all 4 lines, got %d", len(rows))
	}
	if elems[0] != 0 || elems[1] != 1 || elems[2] != 1 || elems[3] != 2 {
		t.Errorf("mapping wrong for the fits-entirely case: %v", elems)
	}
}

func TestBlockWindowDeterministic(t *testing.T) {
	b := testBlocks(2, 1, 3, 1, 2, 2, 1, 3)
	for i := 0; i < 5; i++ {
		r1, e1 := blockWindow(b, 7, 5)
		r2, e2 := blockWindow(b, 7, 5)
		if strings.Join(r1, "|") != strings.Join(r2, "|") {
			t.Fatalf("blockWindow is not deterministic: %v vs %v", r1, r2)
		}
		if len(e1) != len(e2) {
			t.Fatalf("mapping length differs between calls")
		}
	}
}
