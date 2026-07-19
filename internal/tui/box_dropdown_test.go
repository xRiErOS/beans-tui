package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func TestDropdownBoxLayout(t *testing.T) {
	const w = 30
	out := dropdownBox("Status", "todo", "s", w, false)
	lines := strings.Split(out, "\n")
	if len(lines) != 3 {
		t.Fatalf("want 3 lines, got %d: %q", len(lines), out)
	}
	for i, ln := range lines {
		if got := lipgloss.Width(ln); got != w {
			t.Errorf("line %d width = %d, want %d: %q", i, got, w, ln)
		}
	}
	top, mid, bot := ansi.Strip(lines[0]), ansi.Strip(lines[1]), ansi.Strip(lines[2])
	if !strings.HasPrefix(top, "╭─ Status ") {
		t.Errorf("top border missing label: %q", top)
	}
	if !strings.Contains(mid, "todo") || !strings.Contains(mid, "▾") {
		t.Errorf("mid missing value/▾: %q", mid)
	}
	if !strings.Contains(bot, "(s)") || !strings.HasPrefix(bot, "╰") {
		t.Errorf("bottom border missing hotkey: %q", bot)
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
