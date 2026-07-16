package tui

// box_filter_facets.go — the ONE shared facet-filter menu for Tree (this
// task) AND Backlog (E2 Task 5, bean bt-gzu6, reuses everything here
// unchanged). design-spec.md §6 US-05: "Filter wirkt auf Tree UND Backlog".
// Port pattern: devd view_browse_backlog.go:89-136 (buildBacklogFilterItems
// shape) + view_browse_project.go:936-1033 (keyTreeFilter/treeFilterBox) --
// consolidated into ONE implementation instead of devd's two parallel ones
// (Tree-filter vs. Backlog-filter), since beans-tui's filter state already
// lives model-wide, not per-view. No "Art" facet (devd's Milestone/Sprint/
// Issue split) -- beans' Type enum (milestone/epic/feature/task/bug) already
// covers that distinction, so Status/Type/Priority/Tag is the complete set.

import (
	"sort"
	"strings"

	"beans-tui/internal/data"
	"beans-tui/internal/theme"
	keybind "github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// ffItem is one facet-menu row (port devd view_browse_project.go:48-54,
// struct shape 1:1). facet is one of "status"/"type"/"priority"/"tag".
type ffItem struct {
	facet string
	value string
	label string
}

// facetHead groups the filter menu's rows under a muted sub-header per facet
// (port devd view_browse_project.go:1003's facetHead map).
var facetHead = map[string]string{
	"status":   "Status",
	"type":     "Type",
	"priority": "Priority",
	"tag":      "Tags",
	"archive":  "Archive",
}

// buildFilterItems assembles the menu row list: fixed Status/Type/Priority
// facets (the beans enums, design-spec.md §4 -- 5 values each) plus the
// dynamic Tag facet collected from the currently loaded beans.
//
// Enum single source (design decision b, E3 Task 1, bean bt-dlgk): the
// former 15 hardcoded rows are now three loops over data.StatusValues()/
// TypeValues()/PriorityValues() -- the SAME canonical tier order
// data/index.go's statusOrder/typeOrder/priorityOrder derive from, so the
// menu still reads top-to-bottom in "most urgent first" order, and this menu
// can never drift out of sync with box_menu_value.go's combined value menu
// (E3), which consumes the exact same accessors. Type labels are lowercase
// now (previously "Milestone"/"Epic"/... capitalized) -- a deliberate
// consistency fix with status/priority's already-lowercase labels; no golden
// test covers the filter box (box_filter_facets_test.go only asserts facet
// counts/values, not label casing).
func (m model) buildFilterItems() []ffItem {
	var items []ffItem
	for _, v := range data.StatusValues() {
		items = append(items, ffItem{"status", v, v})
	}
	for _, v := range data.TypeValues() {
		items = append(items, ffItem{"type", v, v})
	}
	for _, v := range data.PriorityValues() {
		items = append(items, ffItem{"priority", v, v})
	}
	for _, tag := range m.tagFilterOptions() {
		items = append(items, ffItem{"tag", tag, tag})
	}
	// E5 Task 7 (bean bt-ggt2, design decision e): ONE standalone row, NOT
	// part of the Status-enum loop above (data.StatusValues() has no
	// "archive" pseudo-status) -- toggles m.showArchived (a program DEFAULT,
	// not a PO-set facet, filterActive()'s own doc-stamp) via the SAME
	// space/x toggle path every other row uses (facetOn/toggleFacet below),
	// no new key. value "show" is a placeholder, never map-looked-up:
	// facetOn/toggleFacet special-case facet=="archive" against
	// m.showArchived directly.
	items = append(items, ffItem{"archive", "show", "Show archived"})
	return items
}

// tagFilterOptions collects distinct tag names across every loaded bean,
// deduped, in a DETERMINISTIC order.
//
// ERRATUM vs. the plan's own buildFilterItems sketch ("tag options DYNAMIC
// from idx.ByID, stable insertion order"): idx.ByID is a Go map, which has
// NO defined iteration order at all -- ranging over it directly (as devd's
// tagFilterOptions does over its own api.Issue SLICE, a genuinely
// insertion-ordered source) would make the tag facet's row order flicker
// nondeterministically between calls, breaking any golden test that ever
// renders the filter box.
//
// Fix, two passes: (1) sort the collected beans by ID first -- a
// deterministic BASELINE order that does not depend on the map walk: (2)
// THEN run data.SortBeans (the same single-source canonical order I03
// already established for the orphan bucket above) on top of that baseline.
// data.SortBeans uses sort.SliceStable, so beans that tie completely under
// its own criteria (same Status/Priority/Type/Title -- a real, observed case
// in this package's own tests, see TestTagFilterOptionsDeterministicAcross
// Calls) preserve the ID-sorted baseline instead of silently inheriting
// whatever order that call's map walk happened to produce. A single
// data.SortBeans pass alone is NOT sufficient for full determinism on ties.
func (m model) tagFilterOptions() []string {
	if m.idx == nil {
		return nil
	}
	all := make([]*data.Bean, 0, len(m.idx.ByID))
	for _, b := range m.idx.ByID {
		all = append(all, b)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].ID < all[j].ID })
	data.SortBeans(all)

	seen := map[string]bool{}
	var out []string
	for _, b := range all {
		for _, t := range b.Tags {
			if t != "" && !seen[t] {
				seen[t] = true
				out = append(out, t)
			}
		}
	}
	return out
}

// filterActive reports whether any facet is currently narrowing the view.
func (m model) filterActive() bool {
	return len(m.filterStatus)+len(m.filterType)+len(m.filterPriority)+len(m.filterTag) > 0
}

// beanMatchesFacets is the facet half of the combined filter predicate
// (beanMatches AND-combines this with beanMatchesSearch, view_browse_repo.go
// / search_test.go, E2 Task 3). An empty facet map imposes no constraint
// (matches everything); a non-empty one requires membership.
func (m model) beanMatchesFacets(b *data.Bean) bool {
	if len(m.filterStatus) > 0 && !m.filterStatus[b.Status] {
		return false
	}
	if len(m.filterType) > 0 && !m.filterType[b.Type] {
		return false
	}
	if len(m.filterPriority) > 0 && !m.filterPriority[b.Priority] {
		return false
	}
	if len(m.filterTag) > 0 {
		hit := false
		for _, t := range b.Tags {
			if m.filterTag[t] {
				hit = true
				break
			}
		}
		if !hit {
			return false
		}
	}
	return true
}

// beanMatchesArchive is the archive-visibility half of the combined filter
// predicate (E5 Task 7, bean bt-ggt2, design decision e): completed/scrapped
// beans -- archived (Bean.IsArchived(), data/bean.go) or merely
// completed-but-still-in-.beans/, the distinction is irrelevant to
// visibility policy, epic-E5-plan.md »Task 7« -- are hidden unless
// m.showArchived is toggled on. A PROGRAM DEFAULT, not a PO-set facet:
// filterActive() deliberately does NOT consult m.showArchived (doc-stamp
// there), so this predicate lives entirely outside the four
// filterStatus/filterType/filterPriority/filterTag maps beanMatchesFacets
// checks.
func (m model) beanMatchesArchive(b *data.Bean) bool {
	if m.showArchived {
		return true
	}
	return b.Status != "completed" && b.Status != "scrapped"
}

// beanMatches is the ONE combined predicate every filtered view (Tree here,
// Backlog in E2 Task 5) calls -- AND of search (Task 3's beanMatchesSearch),
// facets (beanMatchesFacets), and archive-visibility (beanMatchesArchive,
// E5 Task 7). Search is NOT special-cased against the archive default: an
// async Bleve hit for an archived bean is excluded here exactly like any
// facet-excluded hit -- the same AND-combination status/type/priority
// already impose on a Bleve result set (beanMatchesSearch's own doc
// comment applies unchanged), no separate wiring needed in
// applyBleveResult (documented Task-7 decision, "Suche respektiert den
// Archiv-Default").
func (m model) beanMatches(b *data.Bean) bool {
	return m.beanMatchesSearch(b) && m.beanMatchesFacets(b) && m.beanMatchesArchive(b)
}

// treeActive generalizes Task 3's treeSearchActive() to "search OR facets are
// narrowing the tree" -- visibleNodes() (view_browse_repo.go) switches to
// the filtered flattening whenever this is true.
func (m model) treeActive() bool {
	return m.treeSearchActive() || m.filterActive()
}

// facetOn reports whether item's value is currently set in its facet map --
// the single lookup treeFilterBox's checkbox rendering and the menu-nav
// tests both need.
func (m model) facetOn(it ffItem) bool {
	switch it.facet {
	case "status":
		return m.filterStatus[it.value]
	case "type":
		return m.filterType[it.value]
	case "priority":
		return m.filterPriority[it.value]
	case "tag":
		return m.filterTag[it.value]
	case "archive":
		return m.showArchived
	}
	return false
}

// toggleFacet flips it's membership in its facet map. I01 (bean bt-7jr8
// T8-review, types.go doc-stamp): clones the TARGET map via cloneBoolMap
// before writing -- copy-on-write, same convention as setExpanded
// (update.go) and expandAncestorsOf.
func (m model) toggleFacet(it ffItem) model {
	toggle := func(mp map[string]bool) map[string]bool {
		out := cloneBoolMap(mp)
		if out[it.value] {
			delete(out, it.value)
		} else {
			out[it.value] = true
		}
		return out
	}
	switch it.facet {
	case "status":
		m.filterStatus = toggle(m.filterStatus)
	case "type":
		m.filterType = toggle(m.filterType)
	case "priority":
		m.filterPriority = toggle(m.filterPriority)
	case "tag":
		m.filterTag = toggle(m.filterTag)
	case "archive":
		// E5 Task 7: a plain bool, not a facet map -- no cloneBoolMap
		// aliasing hazard (model is a value receiver, the bool is copied by
		// value like every other scalar field).
		m.showArchived = !m.showArchived
	}
	return m
}

// clearFacets resets all four facet maps to FRESH EMPTY maps, not nil --
// keeps filterActive()/len() checks cheap and the map identity consistently
// non-nil after any clear, mirroring devd's own X-clear (view_browse_project
// .go:955-960, `map[treeKind]bool{}` literals, not `nil`).
func (m model) clearFacets() model {
	m.filterStatus = map[string]bool{}
	m.filterType = map[string]bool{}
	m.filterPriority = map[string]bool{}
	m.filterTag = map[string]bool{}
	return m
}

// keyFilterMenu drives the floating facet-filter menu (port devd
// keyTreeFilter, view_browse_project.go:936-996): up/down move the cursor,
// space/x toggles the cursored row's facet (copy-on-write), X clears all
// four facet maps WITHOUT closing the menu (devd parity), enter/esc/f close
// the menu without touching filter state.
func (m model) keyFilterMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch navKey(msg.String()) {
	case "up":
		m.filterMenu.move(-1)
		return m, nil
	case "down":
		m.filterMenu.move(1)
		return m, nil
	}
	switch {
	case keybind.Matches(msg, keys.Back), keybind.Matches(msg, keys.Filter), keybind.Matches(msg, keys.Enter):
		m.filterOpen = false
		return m.resetCursorToFirstVisible(), nil
	case keybind.Matches(msg, keys.FilterClear):
		m = m.clearFacets()
		return m.resetCursorToFirstVisible(), nil
	case keybind.Matches(msg, keys.Toggle):
		if m.filterMenu.cursor < 0 || m.filterMenu.cursor >= len(m.filterItems) {
			return m, nil
		}
		m = m.toggleFacet(m.filterItems[m.filterMenu.cursor])
		return m.resetCursorToFirstVisible(), nil
	}
	return m, nil
}

// treeFilterBox renders the floating facet-filter menu (port devd
// treeFilterBox, view_browse_project.go:998-1033).
func (m model) treeFilterBox() string {
	var b strings.Builder
	b.WriteString(theme.Muted.Render("space/x:toggle  X:clear  enter/esc/f:done") + "\n")
	lastFacet := ""
	for i, it := range m.filterItems {
		if it.facet != lastFacet {
			b.WriteString("\n" + theme.Dim.Render(facetHead[it.facet]) + "\n")
			lastFacet = it.facet
		}
		box := theme.Dim.Render("[ ]")
		if m.facetOn(it) {
			box = theme.Accent.Render("[x]")
		}
		cursor := "  "
		label := it.label
		if i == m.filterMenu.cursor {
			cursor = theme.Accent.Render("▸ ")
			label = theme.Header.Render(label)
		}
		b.WriteString(cursor + box + " " + label + "\n")
	}
	return modalPanel("Filter", b.String(), "", clampModalWidth(40, m.width), theme.Mauve)
}

// filterSummary summarizes the active facets for the Tree head row
// (treeSearchLine, view_browse_repo.go) -- port devd filterSummary
// (view_browse_project.go:1035-1060), generalized to beans-tui's 4-facet
// shared state (no "Art" facet here, see the file doc comment above).
func (m model) filterSummary() string {
	var parts []string
	if len(m.filterStatus) > 0 {
		parts = append(parts, "St:"+joinFilterKeys(m.filterStatus))
	}
	if len(m.filterType) > 0 {
		parts = append(parts, "Ty:"+joinFilterKeys(m.filterType))
	}
	if len(m.filterPriority) > 0 {
		parts = append(parts, "Pr:"+joinFilterKeys(m.filterPriority))
	}
	if len(m.filterTag) > 0 {
		parts = append(parts, "Tags:"+joinFilterKeys(m.filterTag))
	}
	return strings.Join(parts, " ")
}

// joinFilterKeys returns a facet map's set keys, comma-joined.
//
// ERRATUM vs. devd's own joinFilterKeys (view_browse_project.go:1062-1071):
// that port source ranges over the map directly with no sort step -- Go map
// iteration order is randomized per-process, so the SAME filter state would
// render its summary keys in a different order from one View() call to the
// next (well, actually stable within one run since Go re-randomizes only
// across process starts, but still nondeterministic across test runs / a
// real hazard for any future golden test covering the filter-active head
// row). Fixed here with sort.Strings -- same display contract, deterministic
// order.
func joinFilterKeys(mp map[string]bool) string {
	var out []string
	for k, v := range mp {
		if v {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return strings.Join(out, ",")
}

// clampModalWidth clamps a floating box's preferred width to the terminal
// width (port devd forms_shared.go clampModalWidth, DD2-55): shrinks on
// narrow terminals (termW-4 leaves 2 cols of margin/border on each side),
// floor 24. First shared-modal-width consumer in beans-tui -- the existing
// quit-confirm box (box_confirm_quit.go) still hardcodes a bare 40; not
// retrofitted here, out of this task's scope.
func clampModalWidth(pref, termW int) int {
	w := pref
	if termW > 4 && termW-4 < w {
		w = termW - 4
	}
	if w < 24 {
		w = 24
	}
	return w
}

// wideModalWidth sizes a floating box relative to the terminal, unlike
// clampModalWidth (which only ever SHRINKS a fixed preference) -- B06,
// design-spec.md §15 PF-17: the Blocking-/Parent-Picker need to GROW on
// wide terminals (PO: "die Breite muss viel weiter werden"), not stay
// pinned to a fixed 48. ~85% of termW, floor 60 (never narrower than the
// old fixed value), ceiling termW-4 (same 2-column margin convention as
// clampModalWidth).
func wideModalWidth(termW int) int {
	w := termW * 85 / 100
	if w < 60 {
		w = 60
	}
	if termW > 4 && w > termW-4 {
		w = termW - 4
	}
	if w < 24 {
		w = 24 // absolute floor, mirrors clampModalWidth's own floor
	}
	return w
}
