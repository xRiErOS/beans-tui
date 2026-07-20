package tui

// box_panel_test.go — S2c: unit tests for panelBox (box_panel.go), mirroring
// box_dropdown_test.go's own dropdownBox coverage pattern.

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// TestPanelBoxMultilineContent asserts multi-line content is framed one
// content-line per │ … │ row, every rendered line's width equals the
// requested box width, and the label appears in the top border.
func TestPanelBoxMultilineContent(t *testing.T) {
	const w = 30
	out := panelBox("Body", "first line\nsecond line\nthird", "", w, false)
	lines := strings.Split(out, "\n")
	if len(lines) != 5 { // top + 3 content + bottom
		t.Fatalf("want 5 lines (top+3 content+bottom), got %d: %q", len(lines), out)
	}
	for i, ln := range lines {
		if got := lipgloss.Width(ln); got != w {
			t.Errorf("line %d width = %d, want %d: %q", i, got, w, ln)
		}
	}
	top := ansi.Strip(lines[0])
	if !strings.HasPrefix(top, "╭─ Body ") {
		t.Errorf("top border missing label: %q", top)
	}
	for i, want := range []string{"first line", "second line", "third"} {
		got := ansi.Strip(lines[i+1])
		if !strings.Contains(got, want) {
			t.Errorf("content line %d missing %q: %q", i, want, got)
		}
	}
	bot := ansi.Strip(lines[4])
	if !strings.HasPrefix(bot, "╰") {
		t.Errorf("bottom border malformed: %q", bot)
	}
}

// TestPanelBoxEmptyContent asserts empty content still renders exactly one
// blank inner line (never a zero-height panel).
func TestPanelBoxEmptyContent(t *testing.T) {
	const w = 20
	out := panelBox("Relations", "", "", w, false)
	lines := strings.Split(out, "\n")
	if len(lines) != 3 { // top + 1 blank content + bottom
		t.Fatalf("want 3 lines (top+1 blank+bottom), got %d: %q", len(lines), out)
	}
	for i, ln := range lines {
		if got := lipgloss.Width(ln); got != w {
			t.Errorf("line %d width = %d, want %d: %q", i, got, w, ln)
		}
	}
	mid := ansi.Strip(lines[1])
	if !strings.HasPrefix(mid, "│") || !strings.HasSuffix(mid, "│") {
		t.Errorf("blank content line not framed: %q", mid)
	}
}

// TestPanelBoxHotkeyInBottomBorder asserts a non-empty hotkey renders as a
// "(x)" badge in the bottom border, and its absence renders no parens.
func TestPanelBoxHotkeyInBottomBorder(t *testing.T) {
	const w = 24
	withKey := panelBox("Body", "content", "e", w, false)
	lines := strings.Split(withKey, "\n")
	bot := ansi.Strip(lines[len(lines)-1])
	if !strings.Contains(bot, "(e)") || !strings.HasPrefix(bot, "╰") {
		t.Errorf("bottom border missing hotkey badge: %q", bot)
	}

	noKey := panelBox("Body", "content", "", w, false)
	lines2 := strings.Split(noKey, "\n")
	bot2 := ansi.Strip(lines2[len(lines2)-1])
	if strings.Contains(bot2, "(") {
		t.Errorf("empty hotkey must not render parens: %q", bot2)
	}
}

// TestPanelBoxHotkeyOverflowClamp asserts B2/B3: at a narrow width with a
// long hotkey, the bottom border never overflows `width` -- if the badge
// can't fit, it drops to a plain dash line instead of exceeding width.
func TestPanelBoxHotkeyOverflowClamp(t *testing.T) {
	for _, w := range []int{8, 9, 10, 12} {
		out := panelBox("Body", "content", "shift+tab", w, false)
		for i, ln := range strings.Split(out, "\n") {
			if got := lipgloss.Width(ln); got != w {
				t.Errorf("width %d: line %d width = %d, want %d (must never overflow): %q", w, i, got, w, ln)
			}
		}
	}
}
