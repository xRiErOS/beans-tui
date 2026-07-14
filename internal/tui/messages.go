package tui

// messages.go — tea.Msg types + tea.Cmd producers ONLY (port convention from
// devd messages.go: no dispatch/rendering lives here, see update.go/
// view_browse_repo.go).

import (
	"beans-tui/internal/data"

	tea "github.com/charmbracelet/bubbletea"
)

// beansLoadedMsg carries the result of an (initial or reload) data.Client.List
// call. err is non-nil on failure -- Update renders it into the status line
// rather than treating it as fatal: the App-Shell must survive a transient
// beans-CLI failure (e.g. a mid-edit malformed frontmatter file) without
// crashing, per the load error handling called out in the task brief.
type beansLoadedMsg struct {
	beans []data.Bean
	err   error
}

// watchMsg signals that data.Watch's debounced onChange fired. Update reacts
// with a full async reload (loadCmd), never a partial/incremental update
// (design decision D02, mirrored in data.Watch's doc comment) -- and NEVER
// synchronously, since onChange itself must not block (see app.go/watcher.go
// B05 doc contract).
type watchMsg struct{}

// watchUnavailableMsg signals that data.Watch failed to start at all (app.go
// Run) -- distinct from watchMsg (a live watcher firing). Sent exactly once,
// asynchronously (app.go: a goroutine, since the unbuffered tea.Program.msgs
// channel would otherwise deadlock the caller if sent before p.Run() starts
// consuming it -- same B05-style constraint as watchMsg, just for the
// startup-failure path instead of the steady-state one). I04 (T8 Opus
// quality review): must surface in the status line, never a silent degrade.
type watchUnavailableMsg struct{}

// loadCmd (re)loads all beans via the CLI client, async -- the sole read path
// for both the initial Init() load and every subsequent reload (ctrl+r,
// watchMsg).
func loadCmd(c *data.Client) tea.Cmd {
	return func() tea.Msg {
		beans, err := c.List()
		return beansLoadedMsg{beans: beans, err: err}
	}
}
