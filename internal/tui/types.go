package tui

// types.go — the App-Shell model + viewID enum (T8, implementation-plan.md
// »E1 Task 8«). Elm value-receiver model, ported architecture-wise from devd
// (~/Obsidian/tools/DeveloperDashboard/apps/cli-go/internal/tui/types.go):
// viewID enum + a single model struct, no devd-API coupling (data layer is
// beans-native from the start, design-spec.md §3.1/§13).

import (
	"beans-tui/internal/data"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
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

// Picker-Stil-Divergenz (E3-T3-Review PFLICHT, carried into bean bt-ppzb/E3
// Task 6, closes the review's "Doc-Note" finding): the overlays above render
// their cursored row in TWO different styles, both deliberate, neither a
// drift bug. box_menu_value.go (valueMenuBox) and box_picker_tag.go
// (tagPickerBox) use theme.Header.Render plus a leading "▸ " glyph on an
// otherwise PLAIN value/tag string. box_picker_parent.go (parentPickerBox)
// and box_picker_blocking.go (blockingPickerBox) instead use the D08 cursor
// treatment shared with the Tree/Backlog panes (ansi.Strip + a whole-line
// theme.Accent.Render("▌"+plain) wrap) -- because THEIR rows are
// relationRow-rendered (already carrying their own status-icon/type-icon
// colors), so a second Header-wrap would clash with the existing per-cell
// theming instead of highlighting it. Full rationale lives on
// parentPickerBox's own doc comment (box_picker_parent.go); this note exists
// so a reader scanning the overlay set here doesn't mistake the split for an
// unnoticed inconsistency.

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

	// Parent-Picker `a` (E3 Task 3, bean bt-p1uz, box_picker_parent.go):
	// parentItems is the row list built fresh at open time
	// (buildParentItems) -- index 0 is ALWAYS the "(Kein Parent)" clear row
	// (id "", port beans-src parentpicker.go's clearParentItem), the rest is
	// data.EligibleParents(idx, b) (self/descendants/invalid-types
	// pre-filtered, design decision f). Single-select, immediate-apply Enter
	// semantics (like the value menu, design decision a3) -- no
	// pending/original diff state here, a bean has exactly one parent.
	parentItems []pickerItem

	// Blocking-Picker `B` (E3 Task 3, bean bt-p1uz, box_picker_blocking.go):
	// blockItems lists every OTHER bean (self excluded, deliberately NO
	// descendant/cycle exclusion -- design decision g: port-parity with
	// beans-src blockingpicker.go, which has none either). blockOriginal/
	// blockPending are two INDEPENDENT maps seeded from the focused bean's
	// CURRENT Blocking field (wholesale-replace convention, mirrors
	// tagOriginal/tagPending) -- every TOGGLE against blockPending during the
	// picker's lifetime goes through cloneBoolMap (I01), same convention as
	// toggleTagPending.
	blockItems    []pickerItem
	blockOriginal map[string]bool
	blockPending  map[string]bool

	// huh-Form-Hosting (E3 Task 4, bean bt-y4ly, Port devd forms_shared.go):
	// form is the embedded huh sub-model (nil when no form is open, the
	// THIRD separate capture state alongside filterOpen/m.overlay, design
	// decision a2 -- huh forms are not a menu overlay); formKind is "create"
	// (T4; T5 adds "editTitle") and drives both submitForm's dispatch
	// (box_confirm_create.go) and formTitle's label (forms_shared.go).
	// pendingCreate/createLabel/createDraft back the Create-Form's
	// Confirm-Gate: pendingCreate is the ALREADY-BUILT createCmd, parked by
	// submitForm until the Confirm-Gate's enter fires it; createLabel is the
	// modal's preview text; createDraft survives an n/esc Confirm-Gate
	// bounce so the Create-Form reopens FILLED instead of losing the PO's
	// work (Draft-Erhalt, port DD2-190, box_confirm_create.go).
	//
	// F1 (Review-Runde 2, Async-Gap-Clobbering): pendingCreate ALSO now
	// doubles as the "a create is in flight" guard -- keyCreateConfirm's
	// enter (box_confirm_create.go) deliberately no longer nulls it when
	// firing the parked Cmd (a deviation from the field's original "parked,
	// not yet dispatched" meaning); it stays non-nil for the WHOLE async gap
	// until createDoneMsg resolves (applyCreateDone below clears it on
	// EITHER outcome). keyNodeAction's Create case, submitForm's "create"
	// case AND dispatchPalette's "create" case (overlay_palette.go, E4 Task
	// 1, bean bt-jpgn -- the Command-Center is a genuine second entry point
	// to the SAME handlers) -- exactly THREE guarded call sites -- all check
	// `pendingCreate != nil` and refuse to start a SECOND create while one
	// is still in flight -- without this, the gap between the Confirm-Gate's
	// enter (overlay -> None, form -> nil) and createDoneMsg arriving is a
	// window where a second `c` (or a palette "create: bean") would cross-
	// contaminate these very same single-slot fields.
	//
	// ERRATUM vs. epic-E3-plan.md's epic-level "Geteilte Infrastruktur"
	// sketch (which additionally lists a `createConfirm bool` field
	// alongside these): design decision a2 is unambiguous that the E3
	// overlays -- INCLUDING Create-Confirm -- are ONE model.overlay
	// overlayID enum, not "6 weitere Bools neben filterOpen/confirmQuit";
	// Task 4's own Step 4 pseudocode agrees verbatim (`overlay =
	// overlayCreateConfirm`, never `createConfirm = true`). No separate bool
	// is added here -- the Confirm-Gate's open/closed state IS
	// `m.overlay == overlayCreateConfirm` (overlayID enum above), the sketch
	// mention is a stale holdover from an earlier draft.
	form          *huh.Form
	formKind      string
	pendingCreate tea.Cmd
	createLabel   string
	createDraft   *beanDraft

	// Editor (E3 Task 5, bean bt-sl45, design decisions c/h): editorTarget
	// captures WHICH bean's body a ctrl+e $EDITOR-Suspend acts on, set
	// immediately before tea.ExecProcess fires (the cursor cannot move
	// DURING a suspend, but capturing explicitly avoids relying on that --
	// same rationale as mutTarget above). The Title-Edit-Form ("e") reuses
	// mutTarget instead, mirroring the value-menu/picker convention (one
	// node-action target at a time) -- forms and the editor suspend are
	// mutually exclusive capture states (m.form != nil vs. a fired
	// editInEditor Cmd), so there is no collision between the two fields.
	//
	// editorETag (F2, Review-Runde 2 fix): the target bean's ETag, captured
	// in the SAME keyNodeAction branch as editorTarget, at $EDITOR-OPEN time
	// -- a deliberate DEVIATION from design decision d's "always read fresh
	// at submit" rule (m.beanETag(id), update.go). That rule is correct for
	// every other E3 overlay (open<->submit is a single keystroke), but
	// wrong for a potentially long-lived $EDITOR session: a watch-reload
	// landing WHILE the PO is still typing would silently rotate m.idx's
	// in-memory ETag out from under them, and a fresh re-read at submit
	// would then let SetBody sail through against a DIFFERENT bean state
	// than the one the PO's edit is actually based on -- a silent lost
	// update, no conflict ever raised. Freezing the etag at open instead
	// means that same external change now surfaces as a genuine, visible
	// ErrConflict (applyEditorFinished, update.go) -- optimistic-lock
	// semantics over the WHOLE session, not just the Submit<->Disk instant.
	editorTarget string
	editorETag   string

	// Delete-Confirm `d` (E3 Task 6, bean bt-ppzb, box_confirm_delete.go):
	// delTitle/delChildren are captured at OPEN time (openDeleteConfirm) for
	// the modal's own render -- mutTarget (shared with every other overlay,
	// above) already carries the bean ID enter acts on, so no separate delID
	// field is needed. delChildren is idx.Children[id]'s DIRECT count only
	// (idx.Children is already in memory -- Port devd box_confirm_delete.go
	// MINUS its loadDeletePreview, no async count-load) -- deliberately NOT
	// data.CollectDescendants's full recursive descendant walk
	// (hierarchy.go/box_picker_parent.go's cycle-exclusion machinery):
	// `beans delete` does not cascade (mutations.go Delete's own doc-stamp),
	// so only the bean's IMMEDIATE children are ever affected -- a
	// grandchild's own parent (the direct child) stays intact either way.
	// ERRATUM (empirically verified, box_confirm_delete.go's own doc-stamp
	// has the full story): the affected direct children do NOT become
	// "(verwaist)"-bucket orphans as originally assumed (epic-E3-plan.md) --
	// beans 0.4.2's `delete` clears their `parent:` field outright, so they
	// become ordinary parentless ROOTS instead. deleteBox's copy reflects
	// the CORRECTED outcome, never devd's own "will also be deleted" cascade
	// wording either way.
	delTitle    string
	delChildren int
	// delLinks (Q01, E3-T6-Review PFLICHT finding, bean bt-qzwt) is the
	// DISTINCT count of OTHER beans that reference the delete target via
	// Blocking or BlockedBy, captured synchronously at open time (same
	// idx-in-memory convention as delChildren, box_confirm_delete.go's
	// countLinkedBeans). Empirically verified (scratch-repo probe, mirrored
	// in internal/data/client_mut_test.go's
	// TestDeleteClearsOtherBeansBlockedByReference/...BlockingReference):
	// `beans delete` silently clears these references in the OTHER bean's
	// frontmatter too -- the exact same "actively rewrites the file, no CLI
	// warning" behavior the parent-field ERRATUM above already documents,
	// just for a different link family. deleteBox's copy warns about this
	// the same way it warns about children losing their parent.
	delLinks int

	// watchUnavailable is set once (I04, T8 Opus quality review) when
	// data.Watch failed to start in app.go's Run: the App-Shell still works
	// (ctrl+r reloads manually), it just never reacts to on-disk changes on
	// its own -- this must be surfaced in the status line, never silently
	// degraded.
	watchUnavailable bool

	// Command-Center (E4 Task 1, bean bt-jpgn, design decisions a/b/h):
	// paletteOpen is a full-capture floating-overlay state, same precedent
	// as filterOpen (handleKey capture order, design decision h). palQuery
	// drives BOTH candidate pools (actions T1, beans T2) -- ONE shared
	// input, design decision b. palList is the cursor over the COMBINED
	// already-filtered result list (palFiltered, rebuilt every keystroke).
	paletteOpen bool
	palQuery    string
	palList     listState

	// Command-Center bean-search half (E4 Task 2, bean bt-yo60, design
	// decision b): palette-SCOPED copies of the Bleve staleness-guard
	// triplet `/`'s own searchBleveIDs/searchBleveFor/searchBleveLoading
	// already establish above -- kept SEPARATE (own fields, own Msg type
	// paletteBleveResultMsg, messages.go) so opening ctrl+k can never
	// clobber an active Tree/Backlog `/` search session, or vice versa.
	// Same wholesale-replace convention as searchBleveIDs: always REPLACED
	// with a fresh map on a fresh (non-stale) result, never mutated in
	// place -- no cloneBoolMap (I01) needed.
	palBleveIDs     map[string]bool
	palBleveFor     string
	palBleveLoading bool
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
