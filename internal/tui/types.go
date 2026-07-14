package tui

// types.go — the App-Shell model + viewID enum (T8, implementation-plan.md
// »E1 Task 8«). Elm value-receiver model, ported architecture-wise from devd
// (~/Obsidian/tools/DeveloperDashboard/apps/cli-go/internal/tui/types.go):
// viewID enum + a single model struct, no devd-API coupling (data layer is
// beans-native from the start, design-spec.md §3.1/§13).

import "beans-tui/internal/data"

// viewID enumerates the top-level screens. T8 ships exactly one (the Browse
// Primat-View, design-spec.md §6 V2); later epics add Backlog/Detail/
// Command-Center/Review/Forms/Overlays as siblings — the switch in View()
// (view_browse_repo.go) grows one case per view, never a branch here.
type viewID int

const (
	viewBrowseRepo viewID = iota
)

// orphanRootID is the synthetic node ID for the "(verwaist)" root that
// collects every bean whose parent field does not resolve to a known bean
// (MANDATORY orphan rule, bean bt-7jr8 / T3-review Q01: dangling parents are
// beans-legal — `.md` is freely editable and `beans check` only reports
// broken_links — so the tree must show them, never silently drop them). The
// leading NUL byte guarantees this can never collide with a real beans ID
// (beans IDs are `<prefix>-<n-chars>`, never containing \x00).
const orphanRootID = "\x00orphans"

// model is the App-Shell state. T8 wires the read-only Tree (design-spec.md
// §6 V2 basis) + async load/reload/watch; mutation state (forms, pickers,
// menus) lands in E2/E3 as new fields on this same struct (devd port
// convention: one model, one viewID switch, no per-view sub-models).
type model struct {
	view viewID

	width, height int

	client  *data.Client
	repoDir string

	idx *data.Index
	err string // rendered in the status line; "" = no error to show

	expanded map[string]bool // node ID (bean ID or orphanRootID) -> expanded
	cursorID string          // currently selected node's ID (bean ID or orphanRootID)

	// detailFocus is the Tree<->Detail focus toggle (bean bt-7jr8 Q01):
	// view-local `tab` handling, deliberately NOT routed through keymap.Right
	// (which stays the tree's expand key).
	detailFocus bool

	confirmQuit bool

	// watchUnavailable is set once (I04, T8 Opus quality review) when
	// data.Watch failed to start in app.go's Run: the App-Shell still works
	// (ctrl+r reloads manually), it just never reacts to on-disk changes on
	// its own -- this must be surfaced in the status line, never silently
	// degraded.
	watchUnavailable bool
}

// newModel builds the initial (pre-load) App-Shell state.
func newModel(client *data.Client, repoDir string) model {
	return model{
		view:     viewBrowseRepo,
		client:   client,
		repoDir:  repoDir,
		expanded: map[string]bool{},
	}
}
