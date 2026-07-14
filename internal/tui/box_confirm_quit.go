package tui

// box_confirm_quit.go — the quit-confirm modal. Ported from devd
// (~/Obsidian/tools/DeveloperDashboard/apps/cli-go/internal/tui/
// box_confirm_quit.go, DD2-49/DD2-174): `q` opens this confirm instead of
// quitting immediately; `enter` confirms, `esc` cancels. `ctrl+c` bypasses
// this entirely (handleKey routes it straight to tea.Quit, bean bt-7jr8 task
// scope: ctrl+c is the hard/immediate kill switch, q is the soft prompt).

import (
	"beans-tui/internal/theme"
	keybind "github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// requestQuit opens the quit-confirm instead of quitting immediately.
func (m model) requestQuit() (tea.Model, tea.Cmd) {
	m.confirmQuit = true
	return m, nil
}

// keyConfirmQuit drives the quit-confirm: enter quits, esc cancels.
func (m model) keyConfirmQuit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case keybind.Matches(msg, keys.Enter):
		return m, tea.Quit
	case keybind.Matches(msg, keys.Back):
		m.confirmQuit = false
		return m, nil
	}
	return m, nil
}

// quitBox renders the floating quit-confirm modal.
func (m model) quitBox() string {
	body := "\n" + theme.Dim.Render("Really quit bt.") + "\n\n"
	body += theme.Accent.Render("enter") + theme.Dim.Render(": quit   ") +
		theme.Accent.Render("esc") + theme.Dim.Render(": cancel")
	return modalPanel("Quit?", body, "", 40, theme.Mauve)
}
