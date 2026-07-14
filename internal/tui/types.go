package tui

// types.go — the App-Shell model + viewID enum (T8, implementation-plan.md
// »E1 Task 8«). Elm value-receiver model, ported architecture-wise from devd
// (~/Obsidian/tools/DeveloperDashboard/apps/cli-go/internal/tui/types.go):
// viewID enum + a single model struct, no devd-API coupling (data layer is
// beans-native from the start, design-spec.md §3.1/§13).

import (
	"beans-tui/internal/data"

	"github.com/charmbracelet/bubbles/textinput"
)

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

// I01 (bean bt-7jr8, T8-review): every map[string]bool field on model
// (expanded, and E2's new filter facet sets) is COPY-ON-WRITE, never mutated
// in place. Rationale: model is a value-receiver Elm architecture (design-
// spec.md §3.3) -- Go map values are reference types, so mutating one
// in-place silently aliases every other struct copy holding the same map
// header (e.g. an old model variable a test kept around, or a future
// undo/diff feature). Bubbletea's own single-active-model discipline
// (Update always returns a fresh value, the caller discards the old one)
// currently hides this hazard, but it is not guaranteed to stay hidden as
// more map fields are added (E2 Task 4's filter facets) -- so every setter
// clones via cloneBoolMap before writing, closing the hazard permanently at
// negligible cost (these maps are always small: expand state + a handful of
// filter selections).
func cloneBoolMap(src map[string]bool) map[string]bool {
	out := make(map[string]bool, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

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

	// Detail-Accordion focus machine (E2 Task 2, bean bt-2jve, port devd
	// view_detail_issue.go's detailFocusView): secCursor/accOpen track the
	// section level (0-based cursor / 1-based exclusive-open, kept as two
	// fields since renderAccordion's `open` param is 1-based per its digit
	// header while every other cursor in this codebase is 0-based).
	// detailLevel 0 = section level, 1 = field level (Beziehungen only).
	// fieldCursor indexes the currently open section's fields when
	// detailLevel == 1. All four reset on every `tab`-into-detail-focus
	// transition (handleKey).
	secCursor, accOpen, detailLevel, fieldCursor int

	// Suche `/` (E2 Task 3, bean bt-4ep2, design-spec.md §6 V2): searchActive
	// is the textinput-focused typing state (single-key shortcuts inactive
	// while true, design-spec §7); searchQuery is the committed-OR-live query
	// (updated on every keystroke, mirrors devd's treeQuery -- there is no
	// separate "live preview vs. committed" split, enter only blurs the
	// input). searchBleveIDs/searchBleveFor answer the async Bleve half: once
	// searchQuery reaches >=3 chars, searchBleveFor tags which query the
	// CURRENT searchBleveIDs answers -- a searchBleveResultMsg whose query no
	// longer matches m.searchQuery on arrival is discarded (staleness guard,
	// messages.go). searchBleveIDs is always REPLACED wholesale with a fresh
	// map (never mutated in place), so it does not need the I01
	// cloneBoolMap convention above -- there is no shared-backing-array
	// hazard when every writer assigns a brand-new map value.
	searchActive       bool
	searchInput        textinput.Model
	searchQuery        string
	searchBleveIDs     map[string]bool
	searchBleveFor     string
	searchBleveLoading bool

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
	ti := textinput.New() // E2 Task 3: Tree search box (port devd app.go treeSearch)
	ti.Placeholder = "Suche (Titel/ID, ab 3 Zeichen zusätzlich Bleve)"
	ti.Prompt = ""
	ti.CharLimit = 80

	return model{
		view:        viewBrowseRepo,
		client:      client,
		repoDir:     repoDir,
		expanded:    map[string]bool{},
		searchInput: ti,
	}
}
