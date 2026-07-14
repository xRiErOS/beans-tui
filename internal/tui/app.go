package tui

// app.go — model constructor + program runner (T8 App-Shell keystone), ported
// architecture-wise from devd (~/Obsidian/tools/DeveloperDashboard/apps/
// cli-go/internal/tui/app.go): AltScreen + mouse cell-motion, dark-background
// forced (huh/lipgloss adaptive-color detection over ssh/tmux is unreliable —
// this app is always dark, see devd DD2-24).

import (
	"beans-tui/internal/data"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Run starts the bt TUI against repoDir (an already-resolved beans repo, see
// data.FindRepo) using client for all reads. AltScreen + mouse (wheel/click)
// are enabled from the start (design-spec.md §9: "Maus (Wheel, Klick-
// Cursor)" is in v1 scope even though T8 doesn't yet wire click handling).
func Run(client *data.Client, repoDir string) error {
	lipgloss.SetHasDarkBackground(true)

	m := newModel(client, repoDir)
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())

	// B05 (MANDATORY, bean bt-7jr8): data.Watch's onChange callback runs on
	// the watcher's own goroutine and must NEVER call stop() synchronously
	// from inside it (that would deadlock -- stop() blocks until the watcher
	// goroutine exits, which is the very goroutine invoking onChange). This
	// consumer only ever hands the running program an async tea.Msg via
	// p.Send -- documented safe to call from any goroutine -- and never
	// touches stop from onChange. stop itself is only ever called from the
	// teardown path below, after p.Run() has returned.
	stop, err := data.Watch(repoDir, func() {
		p.Send(watchMsg{})
	})
	if err == nil {
		defer stop()
	}

	_, runErr := p.Run()
	return runErr
}

// Init kicks off the initial async load (spinnerless, per plan scope).
func (m model) Init() tea.Cmd {
	return loadCmd(m.client)
}
