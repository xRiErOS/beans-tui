package tui

// box_dropdown.go — das eine wiederverwendete jira-Style Dropdown-Widget
// (docs/plans/jira-style-experiment/design-spec.md, D01): Label im oberen
// Rahmen, Wert + ▾ innen, Hotkey unten-rechts im Rahmen. lipgloss kann keinen
// Text-im-Rahmen -> Rahmenzeilen manuell aus Box-Zeichen komponiert. Breiten-
// und ANSI-korrekt über lipgloss.Width. Reused von Detail-Pane (S2) +
// Filter-Leiste (S3).

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/xRiErOS/beans-tui/internal/theme"
)

func clampVisible(s string, n int) string {
	if n < 0 {
		n = 0
	}
	return ansi.Truncate(s, n, "")
}

func borderDashes(n int) string {
	if n < 1 {
		return ""
	}
	return strings.Repeat("─", n)
}

// dropdownBox rendert das 3-Zeilen-Widget in exakt width Zellen Breite.
// focused = Mauve-Rahmen, sonst Overlay.
func dropdownBox(label, value, hotkey string, width int, focused bool) string {
	if width < 8 {
		width = 8
	}
	borderColor := theme.Overlay
	if focused {
		borderColor = theme.Mauve
	}
	frame := lipgloss.NewStyle().Foreground(borderColor)

	labelSeg := "─ " + clampVisible(label, width-6) + " "
	topFill := width - 2 - lipgloss.Width(labelSeg)
	top := frame.Render("╭") + frame.Render(labelSeg) + frame.Render(borderDashes(topFill)) + frame.Render("╮")

	arrow := theme.Chevron.Render("▾")
	inner := width - 6
	val := clampVisible(value, inner)
	pad := inner - lipgloss.Width(val)
	if pad < 0 {
		pad = 0
	}
	mid := frame.Render("│") + " " + val + strings.Repeat(" ", pad) + " " + arrow + " " + frame.Render("│")

	var bot string
	if hotkey == "" {
		bot = frame.Render("╰") + frame.Render(borderDashes(width-2)) + frame.Render("╯")
	} else {
		badge := theme.BindingKey.Render("(" + hotkey + ")")
		badgeSeg := " " + badge + " "
		right := frame.Render(borderDashes(3))
		fill := width - 2 - lipgloss.Width(badgeSeg) - 3
		if fill < 1 {
			fill = 1
		}
		bot = frame.Render("╰") + frame.Render(borderDashes(fill)) + badgeSeg + right + frame.Render("╯")
	}

	return top + "\n" + mid + "\n" + bot
}
