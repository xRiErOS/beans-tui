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
	viewBacklog           // V3 Backlog (E2 Task 5, bean bt-gzu6, design-spec.md §6 V3)
)

// orphanRootID is the synthetic node ID for the "(verwaist)" root that
// collects every bean whose parent field does not resolve to a known bean
// (MANDATORY orphan rule, bean bt-7jr8 / T3-review Q01: dangling parents are
// beans-legal — `.md` is freely editable and `beans check` only reports
// broken_links — so the tree must show them, never silently drop them). The
// leading NUL byte guarantees this can never collide with a real beans ID
// (beans IDs are `<prefix>-<n-chars>`, never containing \x00).
const orphanRootID = "\x00orphans"

// overlayID enumerates E3's node-action overlays (Value-Menü, Tag-/Parent-/
// Blocking-Picker, Create-/Delete-Confirm) -- mutually exclusive by
// construction (design decision a2, bean bt-dlgk): ONE model.overlay field
// (iota-enum) instead of six more bool fields alongside filterOpen/
// confirmQuit. The two E2 bools are NOT retrofitted into this enum
// (documented, deliberate coexistence) -- handleKey's capture order settles
// precedence between them. m.form != nil (T4) is a third, separate capture
// state (huh forms are not a menu overlay). T1 wires overlayValueMenu only;
// T2/T3/T4/T6 add their cases to keyOverlay/composeOverlays as their
// overlays land -- the full closed set is declared here upfront so a new
// overlay is always a new `case`, never a new bool.
type overlayID int

const (
	overlayNone overlayID = iota
	overlayValueMenu
	overlayTagPicker
	overlayParentPicker
	overlayBlockingPicker
	overlayCreateConfirm
	overlayDeleteConfirm
)

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

	// Facetten-Filter `f`/`X` (E2 Task 4, bean bt-9ldr, design-spec.md §6
	// US-05): ONE shared filter state for Tree (this task) AND Backlog (Task
	// 5 reuses it unchanged, same view-agnostic pattern as focusedBean()).
	// filterStatus/filterType/filterPriority/filterTag are COPY-ON-WRITE
	// (I01 doc-stamp above) -- every toggle clones via cloneBoolMap before
	// writing. filterItems is the flattened menu row list built fresh each
	// time the menu opens (buildFilterItems, box_filter_facets.go);
	// filterMenu is its cursor (listState, first non-test consumer since
	// E1). filterOpen mirrors searchActive's "floating overlay fully
	// captures input" precedent (handleKey).
	filterStatus, filterType, filterPriority, filterTag map[string]bool
	filterOpen                                          bool
	filterItems                                         []ffItem
	filterMenu                                          listState

	// Backlog `b` (E2 Task 5, bean bt-gzu6, design-spec.md §6 V3): backlogList
	// is the Backlog master pane's cursor (index-based, listState -- unlike
	// the Tree's bean-ID cursorID, the Backlog is a flat, non-hierarchical
	// list, so an index is sufficient and mirrors devd's own backlogCursor
	// shape). Kept in sync (setLen) by keyBacklog on every keypress it
	// handles and by the `b`-open case in keyTree, since backlogVisible()'s
	// LENGTH can change out from under it (search/filter routes through the
	// SHARED keySearchInput/keyFilterMenu handlers, which know nothing about
	// backlogList -- view_browse_backlog.go's doc comment has the full
	// rationale). backlogSort is the active Sort-Toggle `S` mode: ""
	// (default) leaves idx.Backlog()'s own canonical order in place;
	// "status"/"priority"/"created"/"updated" are the four cycle stops
	// (nextBacklogSort, view_browse_backlog.go).
	backlogList listState
	backlogSort string

	confirmQuit bool

	// E3 (bean bt-dlgk): node-action overlays -- mutually exclusive by
	// construction (ONE enum field, not 6 bools; filterOpen/confirmQuit
	// predate this and are deliberately not retrofitted, see overlayID's
	// doc-stamp above). mutTarget is the bean ID the open overlay acts on,
	// captured at open time; the ETag is NEVER captured -- every submit
	// re-reads m.idx.ByID[mutTarget].ETag (beanETag, update.go) so
	// watch-reloads between open and submit are automatically honored
	// (design decision d). menu is the shared cursor for the value menu
	// (T1) / future pickers (T2/T3) -- one open at a time, so one field
	// suffices. menuItems is the value-menu's row list, built fresh at
	// open time (openValueMenu, box_menu_value.go).
	overlay   overlayID
	mutTarget string
	menu      listState
	menuItems []valueMenuItem

	// Tag-Picker `t` (E3 Task 2, bean bt-8v69, box_picker_tag.go): tagItems
	// is the usage-counted, deterministically sorted row list (count desc,
	// then alpha -- collectTagCounts), built fresh at open time.
	// tagOriginal/tagPending are two INDEPENDENT maps seeded from the
	// focused bean's current tags (wholesale-replace convention, mirrors
	// searchBleveIDs -- NOT the I01 cloneBoolMap-on-every-write pattern for
	// SEEDING, since both are fresh at open). Every TOGGLE against
	// tagPending during the picker's lifetime still goes through
	// cloneBoolMap before writing, though (I01, same convention as
	// toggleFacet) -- the map is long-lived for the whole overlay session,
	// not a throwaway per-keystroke value, so the same aliasing hazard the
	// I01 doc-stamp describes applies. tagInput/tagInputActive/tagInputErr
	// are the free-text new-tag sub-mode (`n`), mirroring searchInput's
	// "one persistent textinput.Model, reset+focused on open" convention
	// (openSearchInput).
	tagItems       []tagCount
	tagOriginal    map[string]bool
	tagPending     map[string]bool
	tagInput       textinput.Model
	tagInputActive bool
	tagInputErr    string

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

	tagIn := textinput.New() // E3 Task 2: Tag-Picker free-text new-tag input
	tagIn.Placeholder = "neuer Tag (a-z0-9, Bindestrich-getrennt)"
	tagIn.Prompt = ""
	tagIn.CharLimit = 40

	return model{
		view:        viewBrowseRepo,
		client:      client,
		repoDir:     repoDir,
		expanded:    map[string]bool{},
		searchInput: ti,
		tagInput:    tagIn,
	}
}
