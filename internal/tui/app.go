package tui

// app.go — model constructor + program runner (T8 App-Shell keystone), ported
// architecture-wise from devd (~/Obsidian/tools/DeveloperDashboard/apps/
// cli-go/internal/tui/app.go): AltScreen + mouse cell-motion, dark-background
// forced (huh/lipgloss adaptive-color detection over ssh/tmux is unreliable —
// this app is always dark, see devd DD2-24).

import (
	"errors"

	"beans-tui/internal/data"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// errNilClient is Q01's (bean bt-7jr8, T8-review) surfaced load error: a nil
// data.Client is a caller-side construction bug (Run() must build one before
// newModel), but Init() must never let that turn into a nil-deref panic --
// see Init() below.
var errNilClient = errors.New("bt: nil beans client (Run() must construct a data.Client before newModel)")

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
	} else {
		// I04 (T8 Opus quality review): don't just swallow the error -- tell
		// the model so it can surface it in the status line. Sent from a
		// goroutine, NOT inline: p.Send blocks on an unbuffered channel until
		// the program's event loop is reading it, which only starts inside
		// p.Run() below -- an inline call here would deadlock before Run()
		// ever gets to execute (mirrors the B05 constraint on watchMsg's own
		// onChange callback, just for this one-shot startup-failure message).
		go p.Send(watchUnavailableMsg{})
	}

	_, runErr := p.Run()
	return runErr
}

// Init kicks off the initial async load (spinnerless, per plan scope).
func (m model) Init() tea.Cmd {
	if m.client == nil { // Q01 (bean bt-7jr8 T8-review): the nil-client invariant is
		// otherwise only enforced by convention at Run()'s call site -- this guard
		// turns a would-be nil-deref panic (inside loadCmd -> Client.List -> run)
		// into a normal, status-line-surfaced load error instead.
		return func() tea.Msg { return beansLoadedMsg{err: errNilClient} }
	}
	return loadCmd(m.client)
}
