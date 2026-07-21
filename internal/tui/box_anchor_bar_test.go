package tui

// box_anchor_bar_test.go — TDD for bt-p78f #16 Slice 1 (anchor bar, epic
// bt-vy1q). Two pure helpers: renderedHeadingLines maps each parsed section to
// its line in the GLAMOUR-rendered body by title match (PO-Entscheidung D03 --
// no body-render rewrite), and anchorBar builds the single-line chip strip with
// click spans, truncating with an ellipsis at 80 columns (PO-Entscheidung D04).

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/xRiErOS/beans-tui/internal/data"
)

func TestRenderedHeadingLinesMatchesByTitle(t *testing.T) {
	// A synthetic "rendered" body in the glamour notty shape ("  ## Title  ")
	// plus interleaved prose -- renderedHeadingLines must find each heading's
	// line by title, stripping the leading hash/space run, in document order.
	rendered := strings.Join([]string{
		"",
		"  ## Ziel                 ",
		"  was wir wollen          ",
		"",
		"  ## Definition of Done   ",
		"  fertig                  ",
		"  ### Detail              ",
	}, "\n")
	secs := []bodySection{
		{title: "Ziel", level: 2},
		{title: "Definition of Done", level: 2},
		{title: "Detail", level: 3},
	}
	got := renderedHeadingLines(rendered, secs)
	want := []int{1, 4, 6}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("section %d line = %d, want %d", i, got[i], want[i])
		}
	}
}

func TestRenderedHeadingLinesDarkStyleNoHashes(t *testing.T) {
	// Real TTY (dark) style drops the hashes -- the title stands alone. Matcher
	// must handle that too (strip a leading hash run that may be absent).
	rendered := "Ziel\nprose\nDefinition of Done\n"
	secs := []bodySection{{title: "Ziel", level: 2}, {title: "Definition of Done", level: 2}}
	got := renderedHeadingLines(rendered, secs)
	if len(got) != 2 || got[0] != 0 || got[1] != 2 {
		t.Fatalf("got %v, want [0 2]", got)
	}
}

func TestRenderedHeadingLinesUnfoundIsNegative(t *testing.T) {
	got := renderedHeadingLines("nothing here\n", []bodySection{{title: "Ghost", level: 1}})
	if len(got) != 1 || got[0] != -1 {
		t.Fatalf("got %v, want [-1] for an unlocatable heading", got)
	}
}

func TestAnchorBarChipSpansAndFit(t *testing.T) {
	secs := []bodySection{
		{title: "Ziel", level: 2},
		{title: "DoD", level: 2},
	}
	row, chips := anchorBar(secs, 40)
	if w := lipgloss.Width(row); w > 40 {
		t.Fatalf("anchor row width %d > 40", w)
	}
	if len(chips) != 2 {
		t.Fatalf("got %d chips, want 2", len(chips))
	}
	// Each chip span must point at its own title within the STRIPPED row.
	stripped := ansi.Strip(row)
	for i, c := range chips {
		if c.start < 0 || c.end > len([]rune(stripped)) || c.start >= c.end {
			t.Fatalf("chip %d span [%d,%d) out of range (row %q)", i, c.start, c.end, stripped)
		}
		seg := string([]rune(stripped)[c.start:c.end])
		if !strings.Contains(seg, secs[i].title) {
			t.Errorf("chip %d span %q does not contain title %q", i, seg, secs[i].title)
		}
		if c.sec != i {
			t.Errorf("chip %d maps to section %d, want %d", i, c.sec, i)
		}
	}
}

func TestAnchorBarTruncatesWithEllipsis(t *testing.T) {
	// Many headings, narrow width -> the row must fit and end in an ellipsis,
	// with only the chips that fit remaining clickable (D04).
	secs := []bodySection{
		{title: "Alpha", level: 2}, {title: "Bravo", level: 2},
		{title: "Charlie", level: 2}, {title: "Delta", level: 2},
		{title: "Echo", level: 2}, {title: "Foxtrot", level: 2},
	}
	row, chips := anchorBar(secs, 20)
	if w := lipgloss.Width(row); w > 20 {
		t.Fatalf("truncated row width %d > 20 (row %q)", w, ansi.Strip(row))
	}
	if !strings.Contains(ansi.Strip(row), "…") {
		t.Errorf("overflowing anchor bar must show an ellipsis, got %q", ansi.Strip(row))
	}
	if len(chips) >= len(secs) {
		t.Errorf("truncated bar kept %d chips, want fewer than %d", len(chips), len(secs))
	}
}

func TestAnchorBarHitAtColumn(t *testing.T) {
	secs := []bodySection{{title: "One", level: 2}, {title: "Two", level: 2}}
	_, chips := anchorBar(secs, 40)
	// A column inside chip 1's span resolves to section 1; a column in the gap
	// or past the end resolves to -1.
	mid := (chips[1].start + chips[1].end) / 2
	if got := anchorBarHit(chips, mid); got != 1 {
		t.Errorf("hit at col %d = %d, want section 1", mid, got)
	}
	if got := anchorBarHit(chips, 9999); got != -1 {
		t.Errorf("hit far past the row = %d, want -1", got)
	}
}

// --- Render + click-jump integration (bt-p78f #16) ---

// anchorHeadingBeans is fixtureBeans with tk-2 given a multi-section body long
// enough that a jump to a later heading yields a real, non-zero scroll offset.
func anchorHeadingBeans() []data.Bean {
	var body strings.Builder
	for _, h := range []string{"Ziel", "Mitte", "Ende"} {
		body.WriteString("## " + h + "\n")
		for i := 0; i < 12; i++ {
			body.WriteString("Absatz-Zeile fuer " + h + " Nummer\n")
		}
	}
	beans := fixtureBeans()
	for i := range beans {
		if beans[i].ID == "tk-2" {
			beans[i].Body = body.String()
		}
	}
	return beans
}

func anchorHeadingModel(t *testing.T) model {
	t.Helper()
	t.Setenv("BT_BOXFORM", "1")
	m := fixtureModel(t, anchorHeadingBeans())
	m = step(t, m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m = focusBean(m, "tk-2")
	return m
}

// TestAnchorBarRendersChips proves the anchor bar shows in the real render at
// the top of the Body panel (all three section chips visible at offset 0).
func TestAnchorBarRendersChips(t *testing.T) {
	m := anchorHeadingModel(t)
	view := ansi.Strip(m.View())
	for _, want := range []string{"[Ziel]", "[Mitte]", "[Ende]"} {
		if !strings.Contains(view, want) {
			t.Errorf("rendered View missing anchor chip %q", want)
		}
	}
}

// TestAnchorBarClickJumpsToHeading is the headline #16 Akzeptanz through the
// REAL render + mouse pipeline: clicking a chip scrolls the Body so that
// heading is at the top, via adjustBoxFormScroll (the one scroll mutation
// point), recording the owning bean.
func TestAnchorBarClickJumpsToHeading(t *testing.T) {
	m := anchorHeadingModel(t)
	b := m.focusedBean()

	// Compute the expected jump for the "Ende" section (index 2) with the SAME
	// helpers the handler uses, then clamp it as adjustBoxFormScroll would.
	accW, _ := boxFormPaneMetrics(m, b)
	blocks := boxFormBlocks(m.idx, b, accW, 0)
	_, _, hl := boxFormBodyPanelContent(b, accW-4)
	targetLine := boxFormAnchorTargetLine(blocks, hl, 2)
	if targetLine <= 0 {
		t.Fatalf("setup: 'Ende' heading target line = %d, want > 0", targetLine)
	}
	total, height := boxFormScrollBounds(m, b)
	want := clampBoxFormScroll(targetLine, total, height)
	if want == 0 {
		t.Fatalf("setup: expected clamp = 0, jump would be a no-op; make the fixture longer")
	}

	msg := boxFormClickAt(t, m, "[Ende]")
	tm, _ := m.handleMouse(msg)
	m2, ok := tm.(model)
	if !ok {
		t.Fatalf("handleMouse(anchor click) did not return a model, got %T", tm)
	}
	if m2.boxFormScroll != want {
		t.Fatalf("boxFormScroll after clicking [Ende] = %d, want %d (jump to the heading via adjustBoxFormScroll)", m2.boxFormScroll, want)
	}
	if m2.boxFormScrollBean != b.ID {
		t.Fatalf("boxFormScrollBean = %q, want %q", m2.boxFormScrollBean, b.ID)
	}
}
