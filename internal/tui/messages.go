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

// searchBleveResultMsg carries the result of an async data.Client.Search
// call (E2 Task 3, bean bt-4ep2), tagged with the query it answers. Update
// (applyBleveResult, update.go) discards it if m.searchQuery has moved on in
// the meantime -- staleness guard chosen over a debounce timer (E2 Task 3
// commit rationale, keySearchInput/dispatchBleveIfDue doc comments): every
// qualifying (>=3 char) keystroke dispatches its own beans-CLI subprocess,
// but only the response matching the CURRENT query is ever applied.
type searchBleveResultMsg struct {
	query string
	ids   []string
	err   error
}

// searchCmd runs an async Bleve full-text search via data.Client.Search,
// tagging the result with query (design-spec.md §6 V2: "-S-Bleve-Modus ab 3
// Zeichen"). Only the resolved bean IDs are kept -- beanMatchesSearch
// (view_browse_repo.go) only ever needs ID membership, not the full Bean.
func searchCmd(c *data.Client, query string) tea.Cmd {
	return func() tea.Msg {
		beans, err := c.Search(query)
		if err != nil {
			return searchBleveResultMsg{query: query, err: err}
		}
		ids := make([]string, len(beans))
		for i, b := range beans {
			ids[i] = b.ID
		}
		return searchBleveResultMsg{query: query, ids: ids}
	}
}
