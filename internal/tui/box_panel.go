package tui

// box_panel.go — panelBox (jira-style experiment, design-spec.md D01/D04):
// a multi-line titled panel sibling to box_dropdown.go's dropdownBox. Same
// border composition (label in the TOP border, hotkey badge in the BOTTOM
// border, focus-colored frame), but the middle is N pre-wrapped content
// lines instead of a single value+▾ row — a panel is not a dropdown, so
// there is no ▾ arrow. ADDITIVE (S2c): used by detailBoxForm's Body/
// Relations/History sections, not wired into the live accordion view.

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/xRiErOS/beans-tui/internal/theme"
)

// panelBox frames pre-rendered multi-line content as a titled box (design-spec
// D01/D04): label in top border, hotkey (if any) in bottom border, content
// lines framed by │ … │. content is already wrapped to the inner width
// (= width-4). Used for Body/Relations/History in detailBoxForm.
func panelBox(label, content, hotkey string, width int, focused bool) string {
	if width < 8 {
		width = 8
	}
	borderColor := theme.Overlay
	if focused {
		borderColor = theme.Mauve
	}
	frame := lipgloss.NewStyle().Foreground(borderColor)
	labelStyle := lipgloss.NewStyle().Foreground(theme.Subtext)

	labelText := clampVisible(label, width-6)
	labelSeg := frame.Render("─ ") + labelStyle.Render(labelText) + frame.Render(" ")
	topFill := width - 2 - (2 + lipgloss.Width(labelText) + 1)
	top := frame.Render("╭") + labelSeg + frame.Render(borderDashes(topFill)) + frame.Render("╮")

	inner := width - 4
	rawLines := strings.Split(content, "\n")
	midLines := make([]string, len(rawLines))
	for i, l := range rawLines {
		l = clampVisible(l, inner)
		pad := inner - lipgloss.Width(l)
		if pad < 0 {
			pad = 0
		}
		midLines[i] = frame.Render("│") + " " + l + strings.Repeat(" ", pad) + " " + frame.Render("│")
	}
	mid := strings.Join(midLines, "\n")

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
