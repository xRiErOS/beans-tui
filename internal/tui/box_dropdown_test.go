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
