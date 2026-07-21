package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func TestDropdownBoxLayout(t *testing.T) {
	for _, w := range []int{30, 24} {
		out := dropdownBox("Status", "todo", "s", w, false)
		lines := strings.Split(out, "\n")
		if len(lines) != 3 {
			t.Fatalf("width %d: want 3 lines, got %d: %q", w, len(lines), out)
		}
		for i, ln := range lines {
			if got := lipgloss.Width(ln); got != w {
				t.Errorf("width %d: line %d width = %d, want %d: %q", w, i, got, w, ln)
			}
		}
		top, mid, bot := ansi.Strip(lines[0]), ansi.Strip(lines[1]), ansi.Strip(lines[2])
		if !strings.HasPrefix(top, "╭─ Status ") {
			t.Errorf("width %d: top border missing label: %q", w, top)
		}
		if !strings.Contains(mid, "todo") || !strings.Contains(mid, "▾") {
			t.Errorf("width %d: mid missing value/▾: %q", w, mid)
		}
		if !strings.Contains(bot, "(s)") || !strings.HasPrefix(bot, "╰") {
			t.Errorf("width %d: bottom border missing hotkey: %q", w, bot)
		}
	}
}

func TestDropdownBoxNoHotkey(t *testing.T) {
	out := dropdownBox("Type", "task", "", 24, false)
	lines := strings.Split(out, "\n")
	bot := ansi.Strip(lines[2])
	if strings.Contains(bot, "(") {
		t.Errorf("empty hotkey must not render parens: %q", bot)
	}
	if lipgloss.Width(lines[2]) != 24 {
		t.Errorf("bottom width drift without hotkey")
	}
}

// TestDropdownBoxHotkeyOverflowClamp asserts B2/B3: at a narrow width with a
// long hotkey, the bottom border never overflows `width` -- if the badge
// can't fit, it drops to a plain dash line instead of exceeding width.
func TestDropdownBoxHotkeyOverflowClamp(t *testing.T) {
	for _, w := range []int{8, 9, 10, 12} {
		out := dropdownBox("Status", "todo", "shift+tab", w, false)
		lines := strings.Split(out, "\n")
		for i, ln := range lines {
			if got := lipgloss.Width(ln); got != w {
				t.Errorf("width %d: line %d width = %d, want %d (must never overflow): %q", w, i, got, w, ln)
			}
		}
	}
}

// TestBoxTopBorderBadgesStableIndicatorPosition guards bean bt-adkn US-02
// (PO-Reject 2026-07-21: "das (e) immer an der gleichen Stelle belassen und
// dafuer die Punkte verschieben. Das fuehrt zu einer stabilen praesentation").
// The hotkey badge (e) must sit at a FIXED column regardless of the indicator's
// width, and the indicator must be LEFT-anchored right after the label (its
// left edge fixed), so paging -- which changes the indicator's content/width --
// never shifts (e) or the indicator's start. Both are asserted by rendering the
// same header with a narrow and a wide badge and comparing columns.
func TestBoxTopBorderBadgesStableIndicatorPosition(t *testing.T) {
	frame := lipgloss.NewStyle()
	const w = 60
	narrow := ansi.Strip(boxTopBorderBadges("Body", "1/9", "e", w, frame))
	wide := ansi.Strip(boxTopBorderBadges("Body", "10/99", "e", w, frame))

	// (e) is parked at a fixed distance from the RIGHT corner in both.
	eNarrow := strings.LastIndex(narrow, "(e)")
	eWide := strings.LastIndex(wide, "(e)")
	if eNarrow < 0 || eWide < 0 {
		t.Fatalf("(e) missing: narrow=%q wide=%q", narrow, wide)
	}
	if len(narrow)-eNarrow != len(wide)-eWide {
		t.Fatalf("(e) not right-anchored: distance from right = %d (narrow) vs %d (wide) -- (e) must stay put while the indicator width changes",
			len(narrow)-eNarrow, len(wide)-eWide)
	}

	// The indicator is LEFT-anchored: its first character starts at the SAME
	// column in both, right after the "Body" label -- not floating against (e).
	iNarrow := strings.IndexAny(narrow, "0123456789")
	iWide := strings.IndexAny(wide, "0123456789")
	if iNarrow != iWide {
		t.Fatalf("indicator not left-anchored: starts at col %d (narrow) vs %d (wide) -- its left edge must be fixed", iNarrow, iWide)
	}
	// And it sits left of (e), not right of it.
	if iNarrow >= eNarrow {
		t.Fatalf("indicator at col %d is not left of (e) at col %d", iNarrow, eNarrow)
	}
}
