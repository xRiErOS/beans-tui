package tui

// Direct primitive tests beyond the golden chrome frame — covers the
// acceptance list in bean bt-5544: breadcrumb, footer hints,
// masterDetailWidths, modalBox, overlayModal (placeOverlay), listState (see
// list_test.go), keymap (see keymap_test.go).

import (
	"strings"
	"testing"

	keybind "github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// TestBreadcrumbFormat guards the beans-tui breadcrumb format `> repo: Title`
// (port-adaptation vs. devd's `> slug: Title`) plus the right-aligned global
// hint, and the title-less `> repo` form.
func TestBreadcrumbFormat(t *testing.T) {
	out := breadcrumb("bt", "Browse", "q:quit", 60)
	if !strings.Contains(out, "> bt") {
		t.Errorf("breadcrumb missing '> bt': %q", out)
	}
	if !strings.Contains(out, "Browse") {
		t.Errorf("breadcrumb missing title: %q", out)
	}
	if !strings.Contains(out, "q:quit") {
		t.Errorf("breadcrumb missing global hint: %q", out)
	}

	bare := breadcrumb("bt", "", "q:quit", 60)
	if strings.Contains(bare, "> bt:") {
		t.Errorf("title-less breadcrumb must not render a trailing colon: %q", bare)
	}
}

// TestBreadcrumbNarrowStacks guards the narrow-terminal fallback: when left+
// right don't fit on one line, breadcrumb stacks them on two lines instead of
// overflowing/truncating silently.
func TestBreadcrumbNarrowStacks(t *testing.T) {
	out := breadcrumb("bt", "A Very Long Screen Title Indeed", "ctrl+k:cmd  p:project  q:quit", 20)
	if !strings.Contains(out, "\n") {
		t.Errorf("expected narrow breadcrumb to stack onto two lines, got single line: %q", out)
	}
}

// TestRenderBindingsSkipsHelplessAndJoinsWithTwoSpaces guards footer-hint
// rendering (devd DD2-175: derived from the keymap, not hand-maintained
// strings) — bindings without a Help().Key are skipped, entries join on two
// spaces.
func TestRenderBindingsSkipsHelplessAndJoinsWithTwoSpaces(t *testing.T) {
	bs := []keybind.Binding{
		keybind.NewBinding(keybind.WithKeys("enter"), keybind.WithHelp("enter", "open")),
		keybind.NewBinding(keybind.WithKeys("x")), // no Help() → must be skipped
		keybind.NewBinding(keybind.WithKeys("esc"), keybind.WithHelp("esc", "back")),
	}
	got := renderBindings(bs)
	want := "enter:open  esc:back"
	if got != want {
		t.Errorf("renderBindings=%q, want %q", got, want)
	}
}

// TestMasterDetailWidthsIsOneThirdTwoThirds guards the 1fr:2fr split on a
// wide terminal where the floor/cap don't kick in.
func TestMasterDetailWidthsIsOneThirdTwoThirds(t *testing.T) {
	lw, rw := masterDetailWidths(90, 0)
	if lw != 30 {
		t.Errorf("lw=%d, want 30 (90/3)", lw)
	}
	if rw != 90-30-4 {
		t.Errorf("rw=%d, want %d", rw, 90-30-4)
	}
}

// TestMasterDetailWidthsFloorAndCap guards the two guard rails: a
// treeWidthFloor raises a too-narrow 1fr share (readability on narrow
// terminals), but never above w*2/5 (devd DD2-91 rework: floor, not fixed).
func TestMasterDetailWidthsFloorAndCap(t *testing.T) {
	// w/3 = 20 < floor 36 → floor wins.
	lw, _ := masterDetailWidths(60, 36)
	if lw != 24 {
		// 60*2/5 = 24 caps the floor here — guards the cap-over-floor rule.
		t.Errorf("lw=%d, want 24 (cap w*2/5 overrides floor)", lw)
	}

	// On a wide terminal the floor must not force a fixed narrow column.
	lw2, _ := masterDetailWidths(300, 36)
	if lw2 != 100 {
		t.Errorf("lw2=%d, want 100 (300/3, floor irrelevant on wide terminal)", lw2)
	}
}

// TestModalBoxHasRoundedBorder guards modal.go's chrome: modalBox must carry
// a RoundedBorder around the (rebased) inner content.
func TestModalBoxHasRoundedBorder(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(termenv.Ascii)

	out := modalBox("line one\nline two", 30, lipgloss.Color("#c6a0f6"))
	if !strings.ContainsAny(out, "╭╮╰╯") {
		t.Error("modalBox output missing RoundedBorder corner glyphs")
	}
}

// TestRebaseBgReopensAfterReset guards the B02/B03 fix directly: under a
// color profile, every inner ESC[0m reset must be followed by a re-opened
// Base background sequence — otherwise cells after an inner reset (e.g. a
// theme.Header.Render call, later a huh form) fall back to the terminal's
// default background instead of the modal's Base.
func TestRebaseBgReopensAfterReset(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(termenv.Ascii)

	in := "a\x1b[0mb"
	out := rebaseBg(in)
	if out == in {
		t.Fatal("rebaseBg did not modify the string under a TrueColor profile")
	}
	if !strings.Contains(out, "\x1b[0m\x1b[") {
		t.Errorf("rebaseBg did not insert a re-open sequence right after the reset: %q", out)
	}
}

// TestRebaseBgAsciiInsertsNoRealColor guards the Ascii-profile path: termenv
// still resolves theme.Base to a (non-nil) NoColor under Ascii, so rebaseBg
// does rewrite the string — but the re-opened sequence must carry no real
// color escape (no visible color can leak into an Ascii terminal).
func TestRebaseBgAsciiInsertsNoRealColor(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	in := "a\x1b[0mb"
	out := rebaseBg(in)
	if strings.Contains(out, "\x1b[38") || strings.Contains(out, "\x1b[48") {
		t.Errorf("rebaseBg inserted a real color escape under the Ascii profile: %q", out)
	}
}

// TestMenuListMarksCursor guards menuList's ▸-cursor placement on the
// selected entry only.
func TestMenuListMarksCursor(t *testing.T) {
	out := menuList(3, 1, func(i int, sel bool) string {
		if sel {
			return "SELECTED"
		}
		return "plain"
	})
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("menuList produced %d lines, want 3", len(lines))
	}
	if !strings.Contains(lines[1], "▸") || !strings.Contains(lines[1], "SELECTED") {
		t.Errorf("line 1 (cursor) = %q, want ▸ + SELECTED", lines[1])
	}
	if strings.Contains(lines[0], "▸") || strings.Contains(lines[2], "▸") {
		t.Error("non-cursor lines must not carry the ▸ marker")
	}
}

// TestPlaceOverlayCentersAndKeepsBackgroundVisible guards overlay.go's
// painter's algorithm: the background remains visible around the overlay,
// and the overlay's own content is present in the composited output.
func TestPlaceOverlayCentersAndKeepsBackgroundVisible(t *testing.T) {
	bg := strings.Repeat("background-row-of-filler-text\n", 9)
	bg = strings.TrimSuffix(bg, "\n")
	fg := "TOP\nMID\nBOT"

	out := placeOverlay(bg, fg, 30, 9)
	if !strings.Contains(out, "TOP") || !strings.Contains(out, "MID") || !strings.Contains(out, "BOT") {
		t.Errorf("composited output missing overlay content: %q", out)
	}
	if !strings.Contains(out, "background-row-of-filler-text") {
		t.Error("composited output lost the background entirely (painter's algorithm broken)")
	}
	if got := len(strings.Split(out, "\n")); got != 9 {
		t.Errorf("composited output has %d lines, want 9 (background line count preserved)", got)
	}
}

// TestCanvasLinesPadsToRectangle guards canvasLines: every line must be
// padded to tw cells and there must be at least th lines, so the overlay
// canvas is a gapless rectangle (devd DD2-216).
func TestCanvasLinesPadsToRectangle(t *testing.T) {
	lines := canvasLines("short\nlonger-row", 20, 5)
	if len(lines) < 5 {
		t.Errorf("canvasLines produced %d lines, want >= 5", len(lines))
	}
	for i, l := range lines {
		if w := lipgloss.Width(l); w != 20 {
			t.Errorf("line %d width=%d, want 20", i, w)
		}
	}
}
