package tui

// overlay_palette_test.go — TDD coverage for the Command-Center (`ctrl+k`/
// `K`, E4 Task 1, bean bt-jpgn). Reuses fixtureBeans/fixtureModel/step/
// keyMsg/runeMsg (update_test.go) and focusBean (box_menu_value_test.go),
// same package. E4 Task 2 (bean bt-yo60) had extended this file with a
// bean-search half (design decision b) -- removed again by B13 (design-
// spec.md §15 PF-16/"US-04-Revision", bean bt-ntoz, E8 Task 7, bean
// bt-yqdy): every test here now pins the action-only contract, permanently
// (not just until a hypothetical future T2).

import (
	"testing"

	"beans-tui/internal/data"
	tea "github.com/charmbracelet/bubbletea"
)

// --- paletteActions: context-aware ordering (design decision b) ---

// TestPaletteActionsBeanContextFirst guards that a focused bean's node
// actions (status/type/priority/tags/create_tag/parent/blocking/edit_title/
// delete) come BEFORE the global actions (create/go_backlog/...). type/
// priority (B12, design-spec.md §15 PF-16, bean bt-ntoz, E8 Task 6) sit
// directly after status -- the bean's own wording ("paletteActions(), im
// focused-bean-Block, nach 'status':"). create_tag (B14, design-spec.md §15
// PF-16, bean bt-ntoz, E8 Task 7, bean bt-yqdy) sits directly after tags --
// same "grouped with the related action" precedent.
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

	wantNodeIDs := []string{"status", "type", "priority", "tags", "create_tag", "parent", "blocking", "edit_title", "delete"}
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

// TestPaletteActionsIncludesSetTypeAndSetPriority guards B12's Palette
// addition (design-spec.md §15 PF-16, bean bt-ntoz, E8 Task 6): the
// Command-Center gains "set type"/"set priority" labels alongside the
// pre-existing "set status" -- NOT a new keybinding (design decision a3
// still reserves exactly ONE key, `s`, for the whole cluster; these are
// additional Palette-only entry points to the SAME m.openValueMenu(group)
// handler).
func TestPaletteActionsIncludesSetTypeAndSetPriority(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m = focusBean(m, "tk-2")

	items := paletteActions(m)
	wantLabels := map[string]string{"status": "set status", "type": "set type", "priority": "set priority"}
	for actionID, label := range wantLabels {
		found := false
		for _, it := range items {
			if it.actionID == actionID {
				found = true
				if it.label != label {
					t.Fatalf("action %q label = %q, want %q", actionID, it.label, label)
				}
			}
		}
		if !found {
			t.Fatalf("paletteActions missing action %q (%q)", actionID, label)
		}
	}
}

// TestDispatchPaletteTypeAndPriorityOpenSeededValueMenu guards
// dispatchPalette's new "type"/"priority" cases (B12): each opens the value
// menu filtered AND seeded on that group, exactly mirroring the pre-existing
// "status" case (m.openValueMenu("status")) -- a genuine second entry point
// to the SAME handler, never a parallel implementation (dispatchPalette's own
// doc-stamp).
func TestDispatchPaletteTypeAndPriorityOpenSeededValueMenu(t *testing.T) {
	cases := []struct {
		actionID string
		want     string
	}{
		{"type", "task"},       // tk-2's Type (fixtureBeans)
		{"priority", "normal"}, // tk-2's Priority (fixtureBeans)
	}
	for _, tc := range cases {
		t.Run(tc.actionID, func(t *testing.T) {
			m := fixtureModel(t, fixtureBeans())
			m = focusBean(m, "tk-2")
			m.paletteOpen = true

			nm, _ := m.dispatchPalette(paletteItem{kind: paletteKindAction, actionID: tc.actionID, label: "set " + tc.actionID})
			mm, ok := nm.(model)
			if !ok {
				t.Fatalf("dispatchPalette did not return a model, got %T", nm)
			}
			if mm.overlay != overlayValueMenu {
				t.Fatalf("overlay = %v, want overlayValueMenu", mm.overlay)
			}
			if len(mm.menuItems) != 5 {
				t.Fatalf("menuItems len = %d, want 5 (%s group only, B11/B12)", len(mm.menuItems), tc.actionID)
			}
			for _, it := range mm.menuItems {
				if it.group != tc.actionID {
					t.Fatalf("menuItems leaked group %q, want only %q", it.group, tc.actionID)
				}
			}
			got := mm.menuItems[mm.menu.cursor]
			if got.value != tc.want {
				t.Fatalf("cursored value = %q, want %q (tk-2's current %s)", got.value, tc.want, tc.actionID)
			}
		})
	}
}

// TestPaletteActionsIncludesCreateTag guards B14's Palette addition
// (design-spec.md §15 PF-16, bean bt-ntoz, E8 Task 7, bean bt-yqdy): the
// Command-Center gains a "create tag" node action, alongside the pre-
// existing "set tags" -- both dispatch into box_picker_tag.go, "create tag"
// additionally opening its free-text new-tag sub-mode (dispatchPalette).
func TestPaletteActionsIncludesCreateTag(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m = focusBean(m, "tk-2")

	items := paletteActions(m)
	found := false
	for _, it := range items {
		if it.actionID == "create_tag" {
			found = true
			if it.label != "create tag" {
				t.Fatalf("action %q label = %q, want %q", "create_tag", it.label, "create tag")
			}
		}
	}
	if !found {
		t.Fatal(`paletteActions missing action "create_tag" ("create tag")`)
	}
}

// TestPaletteActionsNoFocusedBeanOmitsNodeActions guards that the orphan
// root (focusedBean() == nil) yields ONLY global actions -- no node action
// IDs (status/type/priority/tags/create_tag/parent/blocking/edit_title/
// delete) leak in.
func TestPaletteActionsNoFocusedBeanOmitsNodeActions(t *testing.T) {
	beans := append(fixtureBeans(), fixtureOrphanBean())
	m := fixtureModel(t, beans)
	m.cursorID = orphanRootID

	if m.focusedBean() != nil {
		t.Fatal("test setup invalid: focusedBean() != nil on the orphan root")
	}

	items := paletteActions(m)
	nodeActionIDs := map[string]bool{
		"status": true, "type": true, "priority": true, "tags": true, "create_tag": true, "parent": true,
		"blocking": true, "edit_title": true, "delete": true,
	}
	for _, it := range items {
		if nodeActionIDs[it.actionID] {
			t.Fatalf("paletteActions leaked node action %q with no focused bean", it.actionID)
		}
	}
	wantGlobal := []string{"create", "go_backlog", "go_browse", "filter", "search", "reload", "repo_picker", "go_tags", "settings"}
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
		t.Fatalf("len(palFiltered) = %d, want 1 (only 'go to backlog' matches %q)", len(items), m.palQuery)
	}
	if items[0].actionID != "go_backlog" {
		t.Fatalf("palFiltered[0].actionID = %q, want go_backlog", items[0].actionID)
	}
}

// TestPalFilteredActionsFuzzyStatMatchesSetStatusAndSetParent guards T3-
// review I01 (bean bt-kyj5 Prelude, carried into E7 T4): Plan-Step 4 of the
// English-i18n task (bt-w9o8) claimed a fuzzy-regression test for the 5
// word-reversed PF-8 labels ("status: setzen" -> "set status" etc.) but only
// ever re-verified "bckl" -> "go to backlog", whose word order PF-8 left
// UNCHANGED (lowest-risk case, doesn't exercise the reversal at all). This
// test exercises a query against the ACTUALLY reversed "set ..." labels.
// "stat" is verified (not assumed) to be a rune SUBSEQUENCE -- not a
// contiguous substring, fuzzyMatch is a subsequence matcher (fuzzy.go) -- of
// BOTH "set status" (s-t-a-t contiguous) AND "set parent" (s-t-a-...-t,
// spanning "seT stAT" via the trailing "t" of "parent"): 2 matches, not the
// single match the Prelude note's phrasing suggested was likely.
func TestPalFilteredActionsFuzzyStatMatchesSetStatusAndSetParent(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m = focusBean(m, "tk-2") // node actions ("set status"/"set parent") only exist with a focused bean
	m.palQuery = "stat"

	items := m.palFiltered()
	wantIDs := []string{"status", "parent"}
	if len(items) != len(wantIDs) {
		t.Fatalf("len(palFiltered) = %d, want %d (%v) for query %q", len(items), len(wantIDs), wantIDs, m.palQuery)
	}
	for i, want := range wantIDs {
		if items[i].actionID != want {
			t.Fatalf("palFiltered[%d].actionID = %q, want %q", i, items[i].actionID, want)
		}
	}
}

// TestPalFilteredActionsFuzzyGoMatchesAllFiveGoToEntries guards T3-review
// I01 (bean bt-kyj5 Prelude): the "go to <entity>" entries (backlog/browse/
// repo picker/settings) share PF-8's UNCHANGED "go to X" shape -- a plain
// "go" query must still fuzzy-match every one of them (not fewer), in
// declaration order, with no other action or bean leaking in. E10 Task 2
// (bean bt-r92i) added a 5th such entry ("go to tags", go_tags) -- widening
// this guard from 4 to 5 (ERRATUM vs. the original Prelude's "genau die 4
// 'go to'-Einträge" wording, which predates this task and is superseded by
// the new entry, not violated by it). (T5-mini, bean bt-uyzf, optional
// T4-review I01): none of fixtureBeans' titles ("Milestone One"/"Epic One"/
// "Task One"/"Task Two") contain a 'g' at all, so no bean item could ever
// fuzzy-match "go" and silently inflate the count above -- the exact-length
// assertion below implicitly guards that absence too.
func TestPalFilteredActionsFuzzyGoMatchesAllFiveGoToEntries(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m.palQuery = "go"

	items := m.palFiltered()
	wantIDs := []string{"go_backlog", "go_browse", "repo_picker", "go_tags", "settings"}
	if len(items) != len(wantIDs) {
		t.Fatalf("len(palFiltered) = %d, want %d (%v) for query %q", len(items), len(wantIDs), wantIDs, m.palQuery)
	}
	for i, want := range wantIDs {
		if items[i].actionID != want {
			t.Fatalf("palFiltered[%d].actionID = %q, want %q", i, items[i].actionID, want)
		}
	}
}

// TestPalFilteredEmptyQueryReturnsAllActions pins the T1 contract (B13,
// design-spec.md §15 PF-16/"US-04-Revision", made permanent -- the palette's
// former second, bean-result pool no longer exists at all, so "no bean
// items" is no longer a query-dependent qualifier worth naming in the test):
// an empty query returns every action.
func TestPalFilteredEmptyQueryReturnsAllActions(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m.palQuery = ""

	items := m.palFiltered()
	if len(items) != len(paletteActions(m)) {
		t.Fatalf("len(palFiltered) = %d, want %d (all actions)", len(items), len(paletteActions(m)))
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
// m.pendingCreate != nil -- otherwise ctrl+k -> "create bean" would
// cross-contaminate the single createDraft/pendingCreate slots the same way
// Finding 1b originally closed for the `c` key.
func TestDispatchPaletteCreateIgnoredWhileCreateInFlight(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m.pendingCreate = func() tea.Msg { return nil } // simulate an earlier create already in flight
	m.paletteOpen = true

	nm, cmd := m.dispatchPalette(paletteItem{kind: paletteKindAction, actionID: "create", label: "create bean"})
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

// --- create_tag dispatch guard (B14, design-spec.md §15 PF-16, bean bt-ntoz, E8 Task 7, bean bt-yqdy) ---

// TestDispatchPaletteCreateTagOpensTagPickerReadyToType guards the happy
// path post-bt-9ipw-consolidation (D01, epic-E12-plan.md »Item 1«): with a
// focused bean, "create tag" opens the Tag-Picker (overlayTagPicker) with
// its search field ALREADY focused and ready to type -- since D01 merged
// the former separate free-text new-tag sub-mode into the Haupt-Picker
// itself, there is no second "input mode" left to enter; opening the picker
// IS the ready-to-create state.
func TestDispatchPaletteCreateTagOpensTagPickerReadyToType(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m = focusBean(m, "tk-2")
	m.paletteOpen = true

	nm, _ := m.dispatchPalette(paletteItem{kind: paletteKindAction, actionID: "create_tag", label: "create tag"})
	mm, ok := nm.(model)
	if !ok {
		t.Fatalf("dispatchPalette did not return a model, got %T", nm)
	}
	if mm.paletteOpen {
		t.Fatal("dispatchPalette must close the palette")
	}
	if mm.overlay != overlayTagPicker {
		t.Fatalf("overlay = %v, want overlayTagPicker", mm.overlay)
	}
	if !mm.tagInput.Focused() {
		t.Fatal("create_tag must land on the Tag-Picker's search field already focused (D01, no second gate)")
	}
}

// TestDispatchPaletteCreateTagNoFocusedBeanNoOp guards the PFLICHT-Guard
// (bean bt-yqdy's own wording): without a focused bean, "create tag" must be
// a clean no-op -- relying on openTagPicker()'s own internal no-op guard
// (returns m unchanged, overlay untouched, when focusedBean()==nil), the
// SAME guard the plain "tags" case above relies on -- post-D01 the two
// actions are the identical handler, so there is no separate chain left to
// guard here.
func TestDispatchPaletteCreateTagNoFocusedBeanNoOp(t *testing.T) {
	beans := append(fixtureBeans(), fixtureOrphanBean())
	m := fixtureModel(t, beans)
	m.cursorID = orphanRootID
	m.paletteOpen = true

	if m.focusedBean() != nil {
		t.Fatal("test setup invalid: focusedBean() != nil on the orphan root")
	}

	nm, _ := m.dispatchPalette(paletteItem{kind: paletteKindAction, actionID: "create_tag", label: "create tag"})
	mm, ok := nm.(model)
	if !ok {
		t.Fatalf("dispatchPalette did not return a model, got %T", nm)
	}
	if mm.overlay != overlayNone {
		t.Fatalf("overlay = %v, want overlayNone (no focused bean -- must not open the Tag-Picker)", mm.overlay)
	}
}
