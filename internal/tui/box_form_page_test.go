package tui

// box_form_page_test.go — TDD coverage for bean bt-adkn (jira-style-ui
// experiment, epic bt-vy1q): seitenweises Body-Blaettern in der Box-Form.
// pgup/pgdn scroll the Detail pane by one FULL page (== the pane's own
// visible line budget) instead of the line-at-a-time up/down of bt-ze10.
//
// Load-bearing (SSTD C): paging MUST run through the SAME adjustBoxFormScroll
// mutation point the wheel and up/down use -- no second scroll path, so
// line-wise and page-wise scrolling can never drift on the reset/clamp rules.
// These tests assert exactly that: the page delta is boxFormScrollBounds' own
// reported height, applied via adjustBoxFormScroll, clamped identically.
//
// Reuses box_form_scroll_test.go's fixture (boxFormScrollModel /
// boxFormLongBodyBeans / requireOverflow): tk-2 carries an 80-line Body that
// reliably overflows the Detail pane by more than 2x, so ONE page down never
// itself hits the clamp ceiling (the assertion below depends on that).

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

// dotRun extracts the first contiguous run of page-indicator glyphs
// (pageDotFilled/pageDotEmpty) from s, or "" if none -- lets the render tests
// below assert on the indicator without depending on its exact frame column.
func dotRun(s string) string {
	start := strings.IndexAny(s, pageDotFilled+pageDotEmpty)
	if start < 0 {
		return ""
	}
	run := []rune{}
	for _, r := range s[start:] {
		if string(r) == pageDotFilled || string(r) == pageDotEmpty {
			run = append(run, r)
			continue
		}
		break
	}
	return string(run)
}

// TestBoxFormPageDownUpScrollsByPage guards the headline Akzeptanz: while the
// Detail pane is focused and boxFormEnabled(), pgdn moves boxFormScroll down by
// exactly one page (the pane's visible line budget) and pgup moves it back --
// both through adjustBoxFormScroll, which records the owning bean.
func TestBoxFormPageDownUpScrollsByPage(t *testing.T) {
	m := boxFormScrollModel(t)
	b := m.focusedBean()
	total, height := requireOverflow(t, m, b)
	if total-height < height {
		t.Fatalf("fixture must overflow by >= one full page (total=%d, height=%d) for a page-down to land un-clamped", total, height)
	}

	m.detailFocus = true
	m = step(t, m, keyMsg(tea.KeyPgDown))
	if m.boxFormScroll != height {
		t.Fatalf("boxFormScroll after one pgdn = %d, want exactly one page (%d)", m.boxFormScroll, height)
	}
	if m.boxFormScrollBean != b.ID {
		t.Fatalf("boxFormScrollBean = %q, want %q (paging must record its owning bean via adjustBoxFormScroll)", m.boxFormScrollBean, b.ID)
	}

	m = step(t, m, keyMsg(tea.KeyPgUp))
	if m.boxFormScroll != 0 {
		t.Fatalf("boxFormScroll after pgup back = %d, want 0 (one page up returns to the top)", m.boxFormScroll)
	}
}

// TestBoxFormPageDownClampsAtEnd guards that page-wise scrolling honours the
// SAME clamp ceiling as line-wise/wheel scrolling: repeated pgdn pins at
// total-height and never overshoots.
func TestBoxFormPageDownClampsAtEnd(t *testing.T) {
	m := boxFormScrollModel(t)
	b := m.focusedBean()
	total, height := requireOverflow(t, m, b)
	wantMax := total - height

	m.detailFocus = true
	for i := 0; i < 50; i++ {
		m = step(t, m, keyMsg(tea.KeyPgDown))
	}
	if m.boxFormScroll != wantMax {
		t.Fatalf("boxFormScroll after many pgdn = %d, want clamped ceiling %d", m.boxFormScroll, wantMax)
	}
}

// TestBoxFormPageInertWithoutFlag guards the OFF-by-default contract: without
// BT_BOXFORM, pgdn/pgup must never touch box-form scroll state (the whole
// paging path is experiment-gated, epic bt-vy1q "alles additiv + gated").
func TestBoxFormPageInertWithoutFlag(t *testing.T) {
	m := fixtureModel(t, boxFormLongBodyBeans()) // BT_BOXFORM left unset
	m = step(t, m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m = focusBean(m, "tk-2")
	m.detailFocus = true

	m = step(t, m, keyMsg(tea.KeyPgDown))
	if m.boxFormScroll != 0 || m.boxFormScrollBean != "" {
		t.Fatalf("boxFormScroll/boxFormScrollBean = %d/%q, want 0/\"\" (flag off: paging must be inert)", m.boxFormScroll, m.boxFormScrollBean)
	}
}

// --- Page indicator (part b) ---

// TestBoxFormPageIndexMapsOffsetToPage guards the pure page-arithmetic: count
// is ceil(total/height), the current page is offset/height, and the CLAMPED
// ceiling (offset == total-height) always maps to the LAST page even when
// total is not a whole multiple of height -- so paging to the end lights the
// last dot (bean bt-adkn).
func TestBoxFormPageIndexMapsOffsetToPage(t *testing.T) {
	cases := []struct {
		total, height, offset int
		wantPage, wantCount   int
	}{
		{total: 10, height: 20, offset: 0, wantPage: 0, wantCount: 1}, // fits -> single page
		{total: 102, height: 18, offset: 0, wantPage: 0, wantCount: 6},
		{total: 102, height: 18, offset: 18, wantPage: 1, wantCount: 6},
		{total: 102, height: 18, offset: 72, wantPage: 4, wantCount: 6},
		{total: 102, height: 18, offset: 84, wantPage: 5, wantCount: 6}, // maxOff -> last page (not 4)
		{total: 36, height: 18, offset: 18, wantPage: 1, wantCount: 2},  // exact multiple
	}
	for _, c := range cases {
		page, count := boxFormPageIndex(c.total, c.height, c.offset)
		if page != c.wantPage || count != c.wantCount {
			t.Errorf("boxFormPageIndex(%d,%d,%d) = (%d,%d), want (%d,%d)",
				c.total, c.height, c.offset, page, count, c.wantPage, c.wantCount)
		}
	}
}

// TestBoxFormPageIndicatorTracksPage is the headline Akzeptanz through the REAL
// render pipeline: a long Body shows a dot row (one dot per page) on the pane's
// fixed bottom frame, the FIRST dot filled at the top -- and after paging to
// the end the LAST dot is filled, while the row stays visible throughout
// (bean bt-adkn "Seiten-Indikator sichtbar, auch waehrend des Blaetterns").
func TestBoxFormPageIndicatorTracksPage(t *testing.T) {
	m := boxFormScrollModel(t)
	b := m.focusedBean()
	total, height := requireOverflow(t, m, b)
	_, wantCount := boxFormPageIndex(total, height, 0)
	if wantCount < 2 {
		t.Fatalf("fixture must span >= 2 pages (count=%d) for the indicator to be meaningful", wantCount)
	}

	top := dotRun(ansi.Strip(m.View()))
	if len([]rune(top)) != wantCount {
		t.Fatalf("dot run at top = %q (%d dots), want %d dots (one per page)", top, len([]rune(top)), wantCount)
	}
	if strings.Count(top, pageDotFilled) != 1 {
		t.Fatalf("dot run at top = %q, want exactly one filled dot", top)
	}
	if !strings.HasPrefix(top, pageDotFilled) {
		t.Fatalf("dot run at top = %q, want the FIRST dot filled (page 0)", top)
	}

	end := m.adjustBoxFormScroll(b, total) // oversized -> clamps to the last page
	endRun := dotRun(ansi.Strip(end.View()))
	if len([]rune(endRun)) != wantCount {
		t.Fatalf("dot run at end = %q (%d dots), want %d (indicator must stay visible while paging)", endRun, len([]rune(endRun)), wantCount)
	}
	if !strings.HasSuffix(endRun, pageDotFilled) {
		t.Fatalf("dot run at end = %q, want the LAST dot filled (paged to the end)", endRun)
	}
}

// TestBoxFormPageIndicatorAbsentWhenFits guards the no-overflow case: a bean
// whose box-form fits the pane shows NO dot row (count==1) -- so short beans in
// box-form mode stay byte-identical to their pre-bt-adkn render (no golden
// drift for the fits case).
func TestBoxFormPageIndicatorAbsentWhenFits(t *testing.T) {
	t.Setenv("BT_BOXFORM", "1")
	m := fixtureModel(t, fixtureBeans()) // default fixture bodies are short
	m = step(t, m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m = focusBean(m, "tk-1")

	b := m.focusedBean()
	total, height := boxFormScrollBounds(m, b)
	if total > height {
		t.Skipf("fixture tk-1 unexpectedly overflows (total=%d height=%d) -- this test needs a fitting bean", total, height)
	}
	if run := dotRun(ansi.Strip(m.View())); run != "" {
		t.Fatalf("dot run = %q, want none (a fitting box-form must not show a page indicator)", run)
	}
}
