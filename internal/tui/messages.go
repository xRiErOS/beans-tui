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

// mutationDoneMsg carries any mutation's outcome (E3, bean bt-dlgk: the
// SHARED tail every Set*/Add*/Remove*/Delete mutation goes through -- no
// per-mutation Msg types). Success and failure BOTH trigger an unconditional
// reload (applyMutationResult, update.go): success must show the new state,
// an ErrConflict must resolve the now-stale index (design decision d).
type mutationDoneMsg struct{ err error }

// mutateCmd wraps a single mutation call (a data.Client Set*/Add*/Remove*/
// Delete method, already bound to its args via a closure) into the shared
// mutationDoneMsg Cmd -- every E3 overlay (Value-Menü T1, Tag-/Parent-/
// Blocking-Picker T2/T3, Delete T6) dispatches through this ONE producer.
func mutateCmd(fn func() error) tea.Cmd {
	return func() tea.Msg { return mutationDoneMsg{err: fn()} }
}

// createDoneMsg is the ONE exception to mutationDoneMsg (E3 Task 4, bean
// bt-y4ly): Create needs the newly minted bean back (for the post-create
// cursor jump), not just a bare error. Defined here in Task 1 alongside the
// rest of the shared mutation infra (plan »Task 1« Files list) -- Task 4
// wires the Update-dispatch case and the cursor-jump behavior once the
// Create form exists.
type createDoneMsg struct {
	bean data.Bean
	err  error
}

// createCmd runs data.Client.Create async, tagging the result as
// createDoneMsg.
func createCmd(c *data.Client, opts data.CreateOpts) tea.Cmd {
	return func() tea.Msg {
		b, err := c.Create(opts)
		return createDoneMsg{bean: b, err: err}
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

// paletteBleveResultMsg carries the result of an async data.Client.Search
// call dispatched from the Command-Center (E4 Task 2, bean bt-yo60),
// structurally IDENTICAL to searchBleveResultMsg above but kept as its OWN
// type -- applyPaletteBleveResult's staleness guard (update.go) checks
// m.palQuery, never m.searchQuery, so the Palette's Bleve half can never
// cross-talk with an active Tree/Backlog `/` search session (design decision
// b).
type paletteBleveResultMsg struct {
	query string
	ids   []string
	err   error
}

// paletteSearchCmd mirrors searchCmd exactly (data.Client.Search, ID-only
// result kept -- palFilteredBeans, overlay_palette.go, only ever needs ID
// membership), tagged as paletteBleveResultMsg instead.
func paletteSearchCmd(c *data.Client, query string) tea.Cmd {
	return func() tea.Msg {
		beans, err := c.Search(query)
		if err != nil {
			return paletteBleveResultMsg{query: query, err: err}
		}
		ids := make([]string, len(beans))
		for i, b := range beans {
			ids[i] = b.ID
		}
		return paletteBleveResultMsg{query: query, ids: ids}
	}
}
