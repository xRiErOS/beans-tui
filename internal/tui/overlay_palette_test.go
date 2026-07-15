package tui

// overlay_palette_test.go — TDD coverage for the Command-Center (`ctrl+k`/
// `K`, E4 Task 1, bean bt-jpgn). Reuses fixtureBeans/fixtureModel/step/
// keyMsg/runeMsg (update_test.go) and focusBean (box_menu_value_test.go),
// same package. Task 2 extends this file with the bean-search half (design
// decision b) -- every test here pins the action-only contract that T2 must
// not weaken.

import (
	"testing"

	"beans-tui/internal/data"
	tea "github.com/charmbracelet/bubbletea"
)

// --- paletteActions: context-aware ordering (design decision b) ---

// TestPaletteActionsBeanContextFirst guards that a focused bean's node
// actions (status/tags/parent/blocking/edit_title/delete) come BEFORE the
// global actions (create/go_backlog/...).
func TestPaletteActionsBeanContextFirst(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m = focusBean(m, "tk-2") // task bean -- focusedBean() != nil

	items := paletteActions(m)
	if len(items) == 0 {
		t.Fatal("paletteActions returned no items")
	}
	if items[0].actionID != "status" {
		t.Fatalf("items[0].actionID = %q, want %q (first node action)", items[0].actionID, "status")
	}

	wantNodeIDs := []string{"status", "tags", "parent", "blocking", "edit_title", "delete"}
	for i, want := range wantNodeIDs {
		if items[i].actionID != want {
			t.Fatalf("items[%d].actionID = %q, want %q", i, items[i].actionID, want)
		}
	}
	// The global "create" action must appear strictly AFTER every node action.
	createIdx := -1
	for i, it := range items {
		if it.actionID == "create" {
			createIdx = i
			break
		}
	}
	if createIdx == -1 {
		t.Fatal(`"create" action missing from paletteActions`)
	}
	if createIdx < len(wantNodeIDs) {
		t.Fatalf("create action at index %d, want it after all %d node actions", createIdx, len(wantNodeIDs))
	}
}

// TestPaletteActionsNoFocusedBeanOmitsNodeActions guards that the orphan
// root (focusedBean() == nil) yields ONLY global actions -- no node action
// IDs (status/tags/parent/blocking/edit_title/delete) leak in.
func TestPaletteActionsNoFocusedBeanOmitsNodeActions(t *testing.T) {
	beans := append(fixtureBeans(), fixtureOrphanBean())
	m := fixtureModel(t, beans)
	m.cursorID = orphanRootID

	if m.focusedBean() != nil {
		t.Fatal("test setup invalid: focusedBean() != nil on the orphan root")
	}

	items := paletteActions(m)
	nodeActionIDs := map[string]bool{
		"status": true, "tags": true, "parent": true,
		"blocking": true, "edit_title": true, "delete": true,
	}
	for _, it := range items {
		if nodeActionIDs[it.actionID] {
			t.Fatalf("paletteActions leaked node action %q with no focused bean", it.actionID)
		}
	}
	wantGlobal := []string{"create", "go_backlog", "go_browse", "filter", "search", "reload"}
	if len(items) != len(wantGlobal) {
		t.Fatalf("len(items) = %d, want %d (%v)", len(items), len(wantGlobal), wantGlobal)
	}
	for i, want := range wantGlobal {
		if items[i].actionID != want {
			t.Fatalf("items[%d].actionID = %q, want %q", i, items[i].actionID, want)
		}
	}
}

// fixtureOrphanBean is a bean with a dangling parent, shared by every test
// here that needs the synthetic orphan root to exist (visibleNodes only
// surfaces orphanRootID once at least one such bean is present, port
// precedent: update_test.go's TestOrphanShownUnderSyntheticRoot).
func fixtureOrphanBean() data.Bean {
	return data.Bean{
		ID: "orph-1", Title: "Orphaned Task", Status: "todo", Type: "task",
		Priority: "normal", Parent: "does-not-exist",
	}
}

// --- palFiltered: fuzzy filter + empty-query contract ---

// TestPalFilteredActionsFuzzyFiltered guards that palFiltered fuzzy-filters
// the action label list against m.palQuery.
func TestPalFilteredActionsFuzzyFiltered(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m.palQuery = "bckl"

	items := m.palFiltered()
	if len(items) != 1 {
		t.Fatalf("len(palFiltered) = %d, want 1 (only 'go to: backlog' matches %q)", len(items), m.palQuery)
	}
	if items[0].actionID != "go_backlog" {
		t.Fatalf("palFiltered[0].actionID = %q, want go_backlog", items[0].actionID)
	}
}

// TestPalFilteredEmptyQueryReturnsAllActionsNoBeans pins the T1 contract T2
// must not weaken: an empty query returns every action and NO bean items.
func TestPalFilteredEmptyQueryReturnsAllActionsNoBeans(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m.palQuery = ""

	items := m.palFiltered()
	if len(items) != len(paletteActions(m)) {
		t.Fatalf("len(palFiltered) = %d, want %d (all actions, no beans)", len(items), len(paletteActions(m)))
	}
	for _, it := range items {
		if it.kind != paletteKindAction {
			t.Fatalf("palFiltered with empty query contains a non-action item: %+v", it)
		}
	}
}

// --- openPalette / keyPalette: open/nav/filter/dispatch/close ---

func TestOpenPaletteResetsQueryAndSeedsList(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m.palQuery = "stale"

	nm, _ := m.openPalette()
	mm := nm.(model)
	if !mm.paletteOpen {
		t.Fatal("openPalette did not set paletteOpen")
	}
	if mm.palQuery != "" {
		t.Fatalf("palQuery = %q, want empty on open", mm.palQuery)
	}
	if mm.palList.length != len(mm.palFiltered()) {
		t.Fatalf("palList.length = %d, want %d", mm.palList.length, len(mm.palFiltered()))
	}
}

func TestKeyPaletteEnterDispatchesAndCloses(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	nm, _ := m.openPalette()
	m = nm.(model)

	m.palQuery = "bckl"
	m.palList.setLen(len(m.palFiltered()))

	nm2, _ := m.keyPalette(keyMsg(tea.KeyEnter))
	mm := nm2.(model)
	if mm.paletteOpen {
		t.Fatal("keyPalette enter did not close the palette")
	}
	if mm.view != viewBacklog {
		t.Fatalf("view = %v, want viewBacklog (go_backlog dispatched)", mm.view)
	}
}

func TestKeyPaletteEscClosesNoSideEffect(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	nm, _ := m.openPalette()
	m = nm.(model)
	m.palQuery = "bckl"

	nm2, _ := m.keyPalette(keyMsg(tea.KeyEsc))
	mm := nm2.(model)
	if mm.paletteOpen {
		t.Fatal("esc did not close the palette")
	}
	if mm.view != viewBrowseRepo {
		t.Fatalf("view = %v, want viewBrowseRepo unchanged (esc must not dispatch)", mm.view)
	}
}

func TestKeyPaletteRuneAppendsAndResyncsList(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	nm, _ := m.openPalette()
	m = nm.(model)

	nm2, _ := m.keyPalette(runeMsg('c'))
	mm := nm2.(model)
	if mm.palQuery != "c" {
		t.Fatalf("palQuery = %q, want %q", mm.palQuery, "c")
	}
	if mm.palList.length != len(mm.palFiltered()) {
		t.Fatalf("palList.length = %d, want %d (resynced)", mm.palList.length, len(mm.palFiltered()))
	}
}

func TestKeyPaletteBackspaceTrims(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	nm, _ := m.openPalette()
	m = nm.(model)
	m.palQuery = "cr"
	m.palList.setLen(len(m.palFiltered()))

	nm2, _ := m.keyPalette(keyMsg(tea.KeyBackspace))
	mm := nm2.(model)
	if mm.palQuery != "c" {
		t.Fatalf("palQuery = %q, want %q", mm.palQuery, "c")
	}
}

// --- Capture-Order (design decision h) ---

func TestHandleKeyCtrlKOpensPaletteFromTree(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m = step(t, m, tea.KeyMsg{Type: tea.KeyCtrlK})
	if !m.paletteOpen {
		t.Fatal("ctrl+k from the tree did not open the palette")
	}
}

func TestHandleKeyCtrlKUnreachableWhileFilterOpen(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m.filterOpen = true
	m = step(t, m, tea.KeyMsg{Type: tea.KeyCtrlK})
	if m.paletteOpen {
		t.Fatal("ctrl+k opened the palette while filterOpen -- capture order violated (design decision h)")
	}
	if !m.filterOpen {
		t.Fatal("filterOpen was cleared -- ctrl+k must have been swallowed by keyFilterMenu, not routed to the palette")
	}
}

func TestHandleKeyCtrlKOpensPaletteEvenWithOverlayOpen(t *testing.T) {
	// Sanity: an E3 node-action overlay (e.g. Value-Menu) captures BEFORE
	// the palette check -- ctrl+k must NOT reach openPalette while overlay
	// != overlayNone (same capture-order contract as filterOpen above).
	m := fixtureModel(t, fixtureBeans())
	m = focusBean(m, "tk-2")
	m = step(t, m, runeMsg('s')) // opens the value menu overlay
	if m.overlay == overlayNone {
		t.Fatal("test setup invalid: value menu did not open")
	}
	m = step(t, m, tea.KeyMsg{Type: tea.KeyCtrlK})
	if m.paletteOpen {
		t.Fatal("ctrl+k opened the palette while an overlay was open -- capture order violated")
	}
}

// --- F1 in-flight guard (Review-Runde 2, Async-Gap-Clobbering, Finding 1b) ---

// TestDispatchPaletteCreateIgnoredWhileCreateInFlight guards dispatchPalette's
// "create" case OWN copy of the F1 in-flight guard -- a THIRD call site of
// the same single-create invariant keyNodeAction's Create case (update.go)
// and submitForm's "create" case (box_confirm_create.go) already enforce
// (types.go doc-stamp). The Command-Center is a genuine second entry point
// to the SAME handlers (dispatchPalette's own doc-stamp), so its "create"
// case must refuse a second Create-Form exactly like the `c` key does while
// m.pendingCreate != nil -- otherwise ctrl+k -> "create: bean" would
// cross-contaminate the single createDraft/pendingCreate slots the same way
// Finding 1b originally closed for the `c` key.
func TestDispatchPaletteCreateIgnoredWhileCreateInFlight(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m.pendingCreate = func() tea.Msg { return nil } // simulate an earlier create already in flight
	m.paletteOpen = true

	nm, cmd := m.dispatchPalette(paletteItem{kind: paletteKindAction, actionID: "create", label: "create: bean"})
	mm, ok := nm.(model)
	if !ok {
		t.Fatalf("dispatchPalette did not return a model, got %T", nm)
	}
	_ = cmd
	if mm.form != nil {
		t.Fatal("palette create while a create is in flight must not open a second Create-Form")
	}
	if mm.overlay == overlayCreateConfirm {
		t.Fatal("palette create while a create is in flight must not open a second Confirm-Gate")
	}
	if mm.err != createInFlightNote {
		t.Fatalf("err = %q, want createInFlightNote %q", mm.err, createInFlightNote)
	}
	if mm.pendingCreate == nil {
		t.Fatal("pendingCreate must remain set -- the original in-flight create must not be forgotten")
	}
}
