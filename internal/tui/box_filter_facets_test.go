package tui

// box_filter_facets_test.go — TDD coverage for Facetten-Filter `f`/`X` (E2
// Task 4, bean bt-9ldr, design-spec.md §6 US-05): buildFilterItems (fixed
// Status/Type/Priority + dynamic Tag facets), the I01 copy-on-write
// convention applied to the new facet maps (mirrors
// TestSetExpandedDoesNotMutateSharedMapAcrossModelCopies, update_test.go),
// the floating filter-menu key machine (keyFilterMenu), and the combined
// search+facet AND predicate (beanMatches). Reuses fixtureBeans/
// fixtureModel/step/keyMsg/runeMsg/nodeIDs/equalStrings from update_test.go
// (same package).

import (
	"strings"
	"testing"

	"beans-tui/internal/data"
	tea "github.com/charmbracelet/bubbletea"
)

// --- buildFilterItems / tagFilterOptions ---

// TestBuildFilterItemsCoversStatusTypePriorityAndDynamicTags guards the
// fixed facet counts (5 status values, 5 beans types, 5 priorities -- the
// full beans enums, design-spec.md §4) plus the dynamic, deduped Tag facet
// collected from the currently loaded beans.
func TestBuildFilterItemsCoversStatusTypePriorityAndDynamicTags(t *testing.T) {
	beans := []data.Bean{
		{ID: "b1", Title: "One", Status: "todo", Type: "task", Priority: "normal", Tags: []string{"backend", "urgent"}},
		{ID: "b2", Title: "Two", Status: "in-progress", Type: "bug", Priority: "high", Tags: []string{"backend"}},
	}
	m := fixtureModel(t, beans)
	items := m.buildFilterItems()

	count := func(facet string) int {
		n := 0
		for _, it := range items {
			if it.facet == facet {
				n++
			}
		}
		return n
	}
	if n := count("status"); n != 5 {
		t.Errorf("status facet count = %d, want 5", n)
	}
	if n := count("type"); n != 5 {
		t.Errorf("type facet count = %d, want 5", n)
	}
	if n := count("priority"); n != 5 {
		t.Errorf("priority facet count = %d, want 5", n)
	}

	gotTags := map[string]bool{}
	tagCount := 0
	for _, it := range items {
		if it.facet == "tag" {
			gotTags[it.value] = true
			tagCount++
		}
	}
	if tagCount != 2 {
		t.Errorf("tag facet count = %d, want 2 (deduped backend+urgent), items=%v", tagCount, items)
	}
	if !gotTags["backend"] || !gotTags["urgent"] {
		t.Errorf("tag facets = %v, want backend+urgent present", gotTags)
	}
}

// TestTagFilterOptionsDeterministicAcrossCalls guards the fix for the plan's
// own ERRATUM (box_filter_facets.go doc comment): idx.ByID is a Go map with
// no defined iteration order, so a naive "range idx.ByID" walk would make
// the tag facet order flicker between calls. Repeated calls against the SAME
// model must return byte-identical order.
func TestTagFilterOptionsDeterministicAcrossCalls(t *testing.T) {
	beans := make([]data.Bean, 0, 10)
	for i := 0; i < 10; i++ {
		beans = append(beans, data.Bean{
			ID: string(rune('a' + i)), Title: "Bean", Status: "todo", Type: "task", Priority: "normal",
			Tags: []string{string(rune('z' - i))},
		})
	}
	m := fixtureModel(t, beans)

	first := m.tagFilterOptions()
	for i := 0; i < 5; i++ {
		got := m.tagFilterOptions()
		if !equalStrings(got, first) {
			t.Fatalf("tagFilterOptions() call %d = %v, want %v (must be deterministic)", i, got, first)
		}
	}
}

// --- I01 copy-on-write (facet maps) ---

// TestFacetToggleUsesCopyOnWrite mirrors
// TestSetExpandedDoesNotMutateSharedMapAcrossModelCopies (update_test.go):
// toggling a facet on one model copy must not mutate a sibling copy's map --
// same I01 regression shape, applied to the new filter facet fields.
func TestFacetToggleUsesCopyOnWrite(t *testing.T) {
	base := model{filterStatus: map[string]bool{}}
	copy1 := base // struct copy -- map HEADER copied, backing array still shared pre-fix
	copy2 := copy1.toggleFacet(ffItem{facet: "status", value: "todo"})

	if copy1.filterStatus["todo"] {
		t.Error("copy1.filterStatus was mutated by copy2's toggleFacet call -- facet maps must be copy-on-write (I01)")
	}
	if !copy2.filterStatus["todo"] {
		t.Error("copy2.filterStatus should carry the new toggle state")
	}
}

// --- filterActive / clearFacets ---

func TestFilterActiveWhenAnyFacetMapNonEmpty(t *testing.T) {
	m := model{}
	if m.filterActive() {
		t.Fatal("filterActive() true with every facet map empty/nil")
	}
	m.filterTag = map[string]bool{"backend": true}
	if !m.filterActive() {
		t.Fatal("filterActive() false with a non-empty facet map")
	}
}

// TestFilterClearXResetsAllFourFacetMaps guards X as a direct top-level
// reset (menu closed) -- design-spec.md: "X leert alle Facetten".
func TestFilterClearXResetsAllFourFacetMaps(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m.filterStatus = map[string]bool{"todo": true}
	m.filterType = map[string]bool{"task": true}
	m.filterPriority = map[string]bool{"high": true}
	m.filterTag = map[string]bool{"backend": true}

	m = step(t, m, runeMsg('X'))

	if m.filterActive() {
		t.Fatalf("filterActive() still true after X: status=%v type=%v priority=%v tag=%v",
			m.filterStatus, m.filterType, m.filterPriority, m.filterTag)
	}
}

// --- Filter-menu key machine (keyFilterMenu / `f` open) ---

// TestFilterMenuUpDownMovesCursorSpaceTogglesRow guards menu navigation +
// checkbox toggling end-to-end through Update.
func TestFilterMenuUpDownMovesCursorSpaceTogglesRow(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m = step(t, m, runeMsg('f'))
	if !m.filterOpen {
		t.Fatal("f did not open the filter menu")
	}
	if len(m.filterItems) == 0 {
		t.Fatal("filterItems empty after opening the filter menu")
	}
	if m.filterMenu.cursor != 0 {
		t.Fatalf("filterMenu.cursor after open = %d, want 0", m.filterMenu.cursor)
	}

	m = step(t, m, keyMsg(tea.KeyDown))
	if m.filterMenu.cursor != 1 {
		t.Fatalf("filterMenu.cursor after down = %d, want 1", m.filterMenu.cursor)
	}

	it := m.filterItems[1]
	m = step(t, m, keyMsg(tea.KeySpace))
	if !m.facetOn(it) {
		t.Fatalf("facet %+v not toggled on after space", it)
	}
	m = step(t, m, keyMsg(tea.KeySpace))
	if m.facetOn(it) {
		t.Fatalf("facet %+v still on after second space toggle", it)
	}

	// "x" is the Toggle binding's second key (alongside space) -- verify it
	// toggles too, distinct from the uppercase "X" FilterClear binding.
	m = step(t, m, runeMsg('x'))
	if !m.facetOn(it) {
		t.Fatalf("facet %+v not toggled on by lowercase x", it)
	}
}

// TestFilterMenuEscAndEnterAndFCloseWithoutClearing guards that all three
// close keys (esc/enter/f) leave the toggled facet state untouched -- only
// X (a separate binding) clears.
func TestFilterMenuEscAndEnterAndFCloseWithoutClearing(t *testing.T) {
	closeKeys := []tea.KeyMsg{keyMsg(tea.KeyEsc), keyMsg(tea.KeyEnter), runeMsg('f')}
	for _, closeKey := range closeKeys {
		m := fixtureModel(t, fixtureBeans())
		m = step(t, m, runeMsg('f'))
		it := m.filterItems[0]
		m = step(t, m, keyMsg(tea.KeySpace))
		if !m.facetOn(it) {
			t.Fatalf("setup: facet %+v not toggled on", it)
		}

		m = step(t, m, closeKey)
		if m.filterOpen {
			t.Fatalf("close key %v did not close the filter menu", closeKey)
		}
		if !m.facetOn(it) {
			t.Fatalf("close key %v cleared the toggled facet -- closing must not clear filter state", closeKey)
		}
	}
}

// TestFilterMenuXClearsWithoutClosing guards devd parity: X inside the open
// menu clears every facet but leaves the menu open (view_browse_project.go's
// keyTreeFilter FilterClear case does not touch treeFilterOpen).
func TestFilterMenuXClearsWithoutClosing(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m = step(t, m, runeMsg('f'))
	it := m.filterItems[0]
	m = step(t, m, keyMsg(tea.KeySpace))
	if !m.facetOn(it) {
		t.Fatal("setup: facet not toggled on")
	}

	m = step(t, m, runeMsg('X'))
	if !m.filterOpen {
		t.Fatal("X inside the menu closed it -- devd parity requires it to stay open")
	}
	if m.filterActive() {
		t.Fatal("X inside the menu did not clear the facet")
	}
}

// --- Combined search + facet AND (beanMatches / visibleNodes) ---

// TestTreeCombinesSearchAndFacetsWithAnd guards that an active search AND an
// active facet combine with AND semantics: a bean matching the search text
// but excluded by the facet must not be visible; one matching both is.
func TestTreeCombinesSearchAndFacetsWithAnd(t *testing.T) {
	beans := []data.Bean{
		{ID: "a", Title: "Alpha Task", Status: "todo", Type: "task", Priority: "normal"},
		{ID: "b", Title: "Alpha Bug", Status: "todo", Type: "bug", Priority: "normal"},
	}
	m := fixtureModel(t, beans)
	m.searchQuery = "alpha"                     // matches both by title
	m.filterType = map[string]bool{"bug": true} // excludes "a"

	nodes := m.visibleNodes()
	if got := nodeIDs(nodes); !equalStrings(got, []string{"b"}) {
		t.Fatalf("combined search+facet nodes = %v, want [b] (AND-combined)", got)
	}
}

// TestBeanMatchesFacetsEmptyMapsImposeNoConstraint guards the base case:
// every facet map empty -> everything matches.
func TestBeanMatchesFacetsEmptyMapsImposeNoConstraint(t *testing.T) {
	m := model{}
	b := &data.Bean{ID: "x", Status: "todo", Type: "task", Priority: "normal"}
	if !m.beanMatchesFacets(b) {
		t.Fatal("beanMatchesFacets with all facet maps empty must match everything")
	}
}

// TestBeanMatchesFacetsTagRequiresAnyOverlap guards the Tag facet's OR-
// within-facet semantics (a bean matches if ANY of its tags is selected).
func TestBeanMatchesFacetsTagRequiresAnyOverlap(t *testing.T) {
	m := model{filterTag: map[string]bool{"backend": true}}
	hit := &data.Bean{ID: "a", Tags: []string{"frontend", "backend"}}
	miss := &data.Bean{ID: "b", Tags: []string{"frontend"}}
	if !m.beanMatchesFacets(hit) {
		t.Error("bean with a matching tag among several must match the tag facet")
	}
	if m.beanMatchesFacets(miss) {
		t.Error("bean with no matching tag must not match the tag facet")
	}
}

// --- I02 (optional, E2-T3-Review finding carried into bean bt-9ldr): empty
// filter-match render must not panic. ---

// TestEmptyFacetMatchViewRendersWithoutPanic guards that a facet filter
// matching zero beans renders cleanly -- the Tree pane simply shows no rows
// below its (still-rendered) search/filter head row. The real assertion is
// implicit: if View()/visibleNodes() panicked, this test would fail with a
// stack trace rather than a plain t.Fatal.
func TestEmptyFacetMatchViewRendersWithoutPanic(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m.width, m.height = 100, 30
	m.filterStatus = map[string]bool{"scrapped": true} // no fixture bean has this status

	if n := len(m.visibleNodes()); n != 0 {
		t.Fatalf("visibleNodes() = %d nodes, want 0 (no fixture bean is scrapped)", n)
	}
	out := m.View()
	if !strings.Contains(out, "Tree") {
		t.Error("View() with zero filtered matches lost the Tree pane frame entirely")
	}
}
