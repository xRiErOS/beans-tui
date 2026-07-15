package tui

// overlay_palette_test.go — TDD coverage for the Command-Center (`ctrl+k`/
// `K`, E4 Task 1, bean bt-jpgn). Reuses fixtureBeans/fixtureModel/step/
// keyMsg/runeMsg (update_test.go) and focusBean (box_menu_value_test.go),
// same package. Task 2 extends this file with the bean-search half (design
// decision b) -- every test here pins the action-only contract that T2 must
// not weaken.

import (
	"fmt"
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
	wantGlobal := []string{"create", "go_backlog", "go_browse", "filter", "search", "reload", "repo_picker", "settings"}
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

// TestPalFilteredActionsFuzzyGoMatchesAllFourGoToEntries guards T3-review
// I01 (bean bt-kyj5 Prelude): the 4 "go to <entity>" entries (backlog/
// browse/repo picker/settings) share PF-8's UNCHANGED "go to X" shape -- a
// plain "go" query must still fuzzy-match all 4 (not fewer, per the
// Prelude's "genau die 4 'go to'-Einträge" requirement), in declaration
// order, with no other action or bean leaking in. (T5-mini, bean bt-uyzf,
// optional T4-review I01): none of fixtureBeans' titles ("Milestone One"/
// "Epic One"/"Task One"/"Task Two") contain a 'g' at all, so no bean item
// could ever fuzzy-match "go" and silently inflate the 4-item count above --
// the exact-length assertion below implicitly guards that absence too.
func TestPalFilteredActionsFuzzyGoMatchesAllFourGoToEntries(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m.palQuery = "go"

	items := m.palFiltered()
	wantIDs := []string{"go_backlog", "go_browse", "repo_picker", "settings"}
	if len(items) != len(wantIDs) {
		t.Fatalf("len(palFiltered) = %d, want %d (%v) for query %q", len(items), len(wantIDs), wantIDs, m.palQuery)
	}
	for i, want := range wantIDs {
		if items[i].actionID != want {
			t.Fatalf("palFiltered[%d].actionID = %q, want %q", i, items[i].actionID, want)
		}
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

// --- palFilteredBeans: bean-search half (E4 Task 2, bean bt-yo60, design decision b) ---

// TestPalFilteredBeansEmptyQueryNone guards that T2 does not weaken T1's
// TestPalFilteredEmptyQueryReturnsAllActionsNoBeans contract: an empty query
// yields zero paletteKindBean items.
func TestPalFilteredBeansEmptyQueryNone(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m.palQuery = ""

	items := m.palFilteredBeans()
	if len(items) != 0 {
		t.Fatalf("palFilteredBeans with empty query = %d items, want 0", len(items))
	}
}

// TestPalFilteredBeansLocalSubstringBelowThreshold guards the <3-char local
// fallback (mirrors beanMatchesSearch's own bifurcation, view_browse_repo.go):
// "tk" ID-substring-matches both tk-1/tk-2 without any Bleve round-trip.
func TestPalFilteredBeansLocalSubstringBelowThreshold(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m.palQuery = "tk"

	items := m.palFilteredBeans()
	if len(items) != 2 {
		t.Fatalf("len(palFilteredBeans) = %d, want 2 (tk-1, tk-2 ID substring match)", len(items))
	}
	for _, it := range items {
		if it.kind != paletteKindBean {
			t.Fatalf("palFilteredBeans returned a non-bean item: %+v", it)
		}
		if it.bean == nil {
			t.Fatal("palFilteredBeans item has a nil bean")
		}
	}
}

// fixtureManyBeans builds n flat, parentless beans sharing the title token
// "manybean" -- used by the cap/sort tests below, which need a matching set
// larger than paletteBeanResultCap.
func fixtureManyBeans(n int) []data.Bean {
	beans := make([]data.Bean, n)
	for i := 0; i < n; i++ {
		beans[i] = data.Bean{
			ID:       fmt.Sprintf("mb-%02d", i),
			Title:    fmt.Sprintf("Manybean %02d", i),
			Status:   "todo",
			Type:     "task",
			Priority: "normal",
		}
	}
	return beans
}

// TestPalFilteredBeansCappedAt20 guards the paletteBeanResultCap ceiling
// (design decision b: prevents a broad query like "e" from flooding the
// modal) -- 30 matching fixture beans, only paletteBeanResultCap survive.
func TestPalFilteredBeansCappedAt20(t *testing.T) {
	m := fixtureModel(t, fixtureManyBeans(30))
	m.palQuery = "manybean"

	items := m.palFilteredBeans()
	if len(items) != paletteBeanResultCap {
		t.Fatalf("len(palFilteredBeans) = %d, want %d (cap)", len(items), paletteBeanResultCap)
	}
}

// TestPalFilteredBeansSortedCanonically guards that the bean pool is
// canonically ordered (data.SortBeans, I03) -- not left in map-iteration
// (nondeterministic) order.
func TestPalFilteredBeansSortedCanonically(t *testing.T) {
	m := fixtureModel(t, fixtureManyBeans(6))
	m.palQuery = "manybean"

	items := m.palFilteredBeans()
	if len(items) != 6 {
		t.Fatalf("setup: len(palFilteredBeans) = %d, want 6", len(items))
	}
	got := make([]*data.Bean, len(items))
	for i, it := range items {
		got[i] = it.bean
	}
	want := append([]*data.Bean{}, got...)
	data.SortBeans(want)
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("palFilteredBeans order[%d] = %s, want %s (data.SortBeans order)", i, got[i].ID, want[i].ID)
		}
	}
}

// TestPalFilteredOrderActionsBeforeBeans guards that a query matching BOTH an
// action AND a bean returns every paletteKindAction item before any
// paletteKindBean item (design decision b: "Aktionen zuerst", NO score-based
// interleaving).
func TestPalFilteredOrderActionsBeforeBeans(t *testing.T) {
	beans := append(fixtureBeans(), data.Bean{
		ID: "bckl-1", Title: "Something Unrelated", Status: "todo", Type: "task", Priority: "normal",
	})
	m := fixtureModel(t, beans)
	// "bckl" fuzzy-subsequence-matches the "go to backlog" action label
	// (TestPalFilteredActionsFuzzyFiltered precedent above) AND ID-substring-
	// matches bckl-1.
	m.palQuery = "bckl"

	items := m.palFiltered()
	sawBean := false
	for _, it := range items {
		if it.kind == paletteKindBean {
			sawBean = true
			continue
		}
		if sawBean {
			t.Fatalf("action item %+v found AFTER a bean item -- actions must always precede beans (design decision b)", it)
		}
	}
	if !sawBean {
		t.Fatal("test setup invalid: no bean item matched the query")
	}
}

// TestDispatchPaletteBeanJumpsCursorAndSwitchesToBrowse guards dispatchPalette
// on a paletteKindBean item: cursor jumps to the bean, view switches to
// viewBrowseRepo (even from viewBacklog), and the bean's ancestors expand so
// the jump target is actually visible in the next visibleNodes() call
// (expandAncestorsOf, same call shape as keyDetailFocus's relation-jump,
// update.go).
func TestDispatchPaletteBeanJumpsCursorAndSwitchesToBrowse(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m.view = viewBacklog

	target := m.idx.ByID["tk-2"] // parent ep-1
	nm, _ := m.dispatchPalette(paletteItem{kind: paletteKindBean, bean: target, label: relationRow(target)})
	mm, ok := nm.(model)
	if !ok {
		t.Fatalf("dispatchPalette did not return a model, got %T", nm)
	}

	if mm.paletteOpen {
		t.Fatal("dispatchPalette must close the palette")
	}
	if mm.cursorID != "tk-2" {
		t.Fatalf("cursorID = %q, want tk-2", mm.cursorID)
	}
	if mm.view != viewBrowseRepo {
		t.Fatalf("view = %v, want viewBrowseRepo", mm.view)
	}
	if !mm.expanded["ep-1"] {
		t.Fatal("dispatchPalette must expand tk-2's ancestors (ep-1) so the jump target is visible in the tree")
	}
}

// TestDispatchPaletteBeanJumpResetsDetailFocus guards B01 (E4 Task 2 review):
// a bean-jump from Detail-Focus must reset BOTH detailFocus AND the
// Detail-Accordion focus machine ints (secCursor/accOpen/detailLevel/
// fieldCursor) -- same reset shape as tab-into-detail-focus (types.go's own
// "All four reset on every tab-into-detail-focus transition" doc-stamp) --
// otherwise arrow keys on the NEW bean manipulate a stale accordion position
// left over from whatever bean the palette was opened FROM, instead of
// driving the tree (empirically confirmed by reviewer, precedent:
// keyDetailFocus's own relation-jump, update.go:702, resets detailFocus on
// the same jump-and-leave-detail-focus shape).
func TestDispatchPaletteBeanJumpResetsDetailFocus(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m.detailFocus = true
	m.secCursor = 2
	m.detailLevel = 1
	m.fieldCursor = 3

	target := m.idx.ByID["tk-2"]
	nm, _ := m.dispatchPalette(paletteItem{kind: paletteKindBean, bean: target, label: relationRow(target)})
	mm, ok := nm.(model)
	if !ok {
		t.Fatalf("dispatchPalette did not return a model, got %T", nm)
	}

	if mm.detailFocus {
		t.Fatal("dispatchPalette bean-jump did not reset detailFocus -- arrow keys would still drive the accordion, not the tree")
	}
	if mm.secCursor != 0 || mm.accOpen != 1 || mm.detailLevel != 0 || mm.fieldCursor != 0 {
		t.Fatalf("focus-machine ints not reset: secCursor=%d accOpen=%d detailLevel=%d fieldCursor=%d, want 0,1,0,0",
			mm.secCursor, mm.accOpen, mm.detailLevel, mm.fieldCursor)
	}
	if mm.view != viewBrowseRepo {
		t.Fatalf("view = %v, want viewBrowseRepo", mm.view)
	}
	if mm.cursorID != "tk-2" {
		t.Fatalf("cursorID = %q, want tk-2", mm.cursorID)
	}
}

// --- Palette-scoped Bleve half (palBleveIDs/palBleveFor/palBleveLoading) ---

// TestKeyPaletteDispatchesBleveOnQueryGrowth guards that keyPalette's
// rune-typing path dispatches a paletteSearchCmd once palQuery reaches the
// Bleve threshold (mirrors TestSearchBleveFiresOnlyAtThreeOrMoreChars,
// search_test.go, but routed through keyPalette directly instead of
// Update()/keySearchInput).
func TestKeyPaletteDispatchesBleveOnQueryGrowth(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	nm, _ := m.openPalette()
	m = nm.(model)

	nm, _ = m.keyPalette(runeMsg('a'))
	m = nm.(model)
	nm, _ = m.keyPalette(runeMsg('b'))
	m = nm.(model)
	if m.palBleveLoading {
		t.Fatal("2 chars must not dispatch a palette Bleve search yet")
	}

	nm, cmd := m.keyPalette(runeMsg('c'))
	mm, ok := nm.(model)
	if !ok {
		t.Fatalf("keyPalette did not return a model, got %T", nm)
	}
	if cmd == nil {
		t.Fatal("the keystroke reaching 3 chars must dispatch a palette Bleve search (paletteSearchCmd)")
	}
	if !mm.palBleveLoading {
		t.Fatal("palBleveLoading must be set once the palette Bleve search is dispatched")
	}
}

// TestApplyPaletteBleveResultDiscardsStaleQuery guards the staleness guard
// (mirrors TestSearchBleveStaleResultDiscardedWhenQueryChangedMeanwhile,
// search_test.go): a paletteBleveResultMsg tagged for a query that no longer
// matches m.palQuery must be a no-op.
func TestApplyPaletteBleveResultDiscardsStaleQuery(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m.palQuery = "abcd" // query has already moved on past "abc"

	m = step(t, m, paletteBleveResultMsg{query: "abc", ids: []string{"tk-1"}})

	if m.palBleveFor != "" {
		t.Fatalf("stale result must not update palBleveFor, got %q", m.palBleveFor)
	}
	if m.palBleveIDs != nil {
		t.Fatalf("stale result must not update palBleveIDs, got %v", m.palBleveIDs)
	}
}

// TestApplyPaletteBleveResultAppliedWhenQueryStillCurrent is the positive
// counterpart (mirrors TestSearchBleveResultAppliedWhenQueryStillCurrent).
func TestApplyPaletteBleveResultAppliedWhenQueryStillCurrent(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m.palQuery = "abc"
	m.palBleveLoading = true

	m = step(t, m, paletteBleveResultMsg{query: "abc", ids: []string{"tk-1"}})

	if m.palBleveFor != "abc" {
		t.Fatalf("palBleveFor = %q, want \"abc\"", m.palBleveFor)
	}
	if !m.palBleveIDs["tk-1"] {
		t.Fatal("palBleveIDs must contain tk-1 from the applied result")
	}
	if m.palBleveLoading {
		t.Fatal("applying a (non-stale) result must clear palBleveLoading")
	}
}

// TestPaletteSearchCmdTagsResultWithQuery guards paletteSearchCmd's message
// tagging directly (mirrors TestSearchCmdTagsResultWithQuery, search_test.go
// -- no beans binary required either way, data.Client.Search tags the query
// on both its success and its error path).
func TestPaletteSearchCmdTagsResultWithQuery(t *testing.T) {
	c := &data.Client{RepoDir: t.TempDir()}
	msg := paletteSearchCmd(c, "abc")()
	res, ok := msg.(paletteBleveResultMsg)
	if !ok {
		t.Fatalf("paletteSearchCmd()() = %T, want paletteBleveResultMsg", msg)
	}
	if res.query != "abc" {
		t.Errorf("paletteBleveResultMsg.query = %q, want \"abc\"", res.query)
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
