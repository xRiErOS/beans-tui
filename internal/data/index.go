package data

import (
	"sort"
	"strings"
)

// Index is an in-memory snapshot over a set of beans, built once per
// load/reload from Client.List() output -- there is no incremental update
// path (YAGNI): a reload just builds a fresh Index.
//
// ByID and Children are exported for cheap read access. Consumers MUST NOT
// mutate the maps, their slices, or the underlying *Bean values -- this is
// not enforced by the type system, only by convention. All ordering must go
// through sortBeans; do not re-sort or append to Children entries directly.
type Index struct {
	// ByID looks up a bean by its short ID.
	ByID map[string]*Bean
	// Children maps a parent ID to its direct children, sorted
	// Status -> Priority -> Type -> Title (see sortBeans).
	Children map[string][]*Bean

	beans []*Bean
}

// NewIndex builds an Index over beans. Callers must not mutate the slice
// (or its elements) afterward -- Index holds pointers into it.
func NewIndex(beans []Bean) *Index {
	idx := &Index{
		ByID:     make(map[string]*Bean, len(beans)),
		Children: make(map[string][]*Bean),
		beans:    make([]*Bean, 0, len(beans)),
	}

	for i := range beans {
		b := &beans[i]
		idx.ByID[b.ID] = b
		idx.beans = append(idx.beans, b)

		if b.Parent != "" {
			idx.Children[b.Parent] = append(idx.Children[b.Parent], b)
		}
	}

	for parent := range idx.Children {
		sortBeans(idx.Children[parent])
	}

	return idx
}

// Roots returns top-level beans (no parent), optionally restricted to the
// given types, sorted Status -> Priority -> Type -> Title. Used by the tree
// view to seed milestones plus parentless epics/rest.
func (idx *Index) Roots(types ...string) []*Bean {
	var allowed map[string]bool
	if len(types) > 0 {
		allowed = make(map[string]bool, len(types))
		for _, t := range types {
			allowed[t] = true
		}
	}

	var roots []*Bean
	for _, b := range idx.beans {
		if b.Parent != "" {
			continue
		}
		if allowed != nil && !allowed[b.Type] {
			continue
		}
		roots = append(roots, b)
	}

	sortBeans(roots)
	return roots
}

// Backlog returns parentless beans that are neither milestone nor epic,
// with status todo or draft -- the "unscheduled work" view.
func (idx *Index) Backlog() []*Bean {
	var backlog []*Bean
	for _, b := range idx.beans {
		if b.Parent != "" {
			continue
		}
		if b.Type == "milestone" || b.Type == "epic" {
			continue
		}
		if b.Status != "todo" && b.Status != "draft" {
			continue
		}
		backlog = append(backlog, b)
	}

	sortBeans(backlog)
	return backlog
}

// WithTag returns all beans carrying tag, sorted Status -> Priority -> Type
// -> Title.
func (idx *Index) WithTag(tag string) []*Bean {
	var tagged []*Bean
	for _, b := range idx.beans {
		for _, t := range b.Tags {
			if t == tag {
				tagged = append(tagged, b)
				break
			}
		}
	}

	sortBeans(tagged)
	return tagged
}

// Sort orders mirror beans 0.4.2's config defaults (implementation-plan.md
// »E1 Task 3«). Values not listed here sort last within their tier.
//
// Scope cut: these are hardcoded beans-0.4.2 defaults only. This does not
// read custom status/priority/type names from .beans.yml, and upstream's
// manual fractional-order tier is deliberately not implemented here (out of
// scope per plan Task 3).
//
// statusValues/typeValues/priorityValues are the canonical, ORDERED enum
// single source (design decision b, E3 Task 1, bean bt-dlgk): beans 0.4.2's
// fixed enums in tier order. The rank maps below are DERIVED from these
// slices (position == rank), not maintained as a second, independently
// hand-authored order -- box_filter_facets.go's buildFilterItems and E3's
// combined value menu (box_menu_value.go) both consume the exported
// accessors (StatusValues/TypeValues/PriorityValues) instead of
// re-hardcoding the values a second time.
var (
	statusValues   = []string{"in-progress", "todo", "draft", "completed", "scrapped"}
	typeValues     = []string{"milestone", "epic", "bug", "feature", "task"}
	priorityValues = []string{"critical", "high", "normal", "low", "deferred"}

	statusOrder   = rankMap(statusValues)
	typeOrder     = rankMap(typeValues)
	priorityOrder = rankMap(priorityValues)
)

// rankMap builds a value->position lookup from an ordered slice (position ==
// rank) -- the single place statusOrder/typeOrder/priorityOrder derive their
// ranking from their respective ordered slice above.
func rankMap(vals []string) map[string]int {
	m := make(map[string]int, len(vals))
	for i, v := range vals {
		m[v] = i
	}
	return m
}

// StatusValues returns the canonical beans status enum in tier order. A
// defensive copy -- callers must not be able to mutate the package-level
// slice (and thus every future caller's view of it) through the returned
// value.
func StatusValues() []string { return append([]string(nil), statusValues...) }

// TypeValues returns the canonical beans type enum in tier order (defensive
// copy, see StatusValues).
func TypeValues() []string { return append([]string(nil), typeValues...) }

// PriorityValues returns the canonical beans priority enum in tier order
// (defensive copy, see StatusValues).
func PriorityValues() []string { return append([]string(nil), priorityValues...) }

// sortBeans sorts in place by Status -> Priority -> Type -> Title
// (case-insensitive), matching beans upstream ordering. This is the single
// place sort order is defined; every Index method reuses it. Empty
// priority is treated as "normal" (beans 0.4.2 default).
//
// Deliberate deviation: sort.SliceStable is used for a deterministic order
// on full ties (same status/priority/type/title), rather than sort.Slice.
// Upstream beans sorts with a non-stable sort -- this is not guaranteed to
// produce byte-for-byte parity with upstream ordering on ties.
func sortBeans(beans []*Bean) {
	sort.SliceStable(beans, func(i, j int) bool {
		a, b := beans[i], beans[j]

		if ra, rb := rank(statusOrder, a.Status), rank(statusOrder, b.Status); ra != rb {
			return ra < rb
		}

		pa, pb := a.Priority, b.Priority
		if pa == "" {
			pa = "normal"
		}
		if pb == "" {
			pb = "normal"
		}
		if ra, rb := rank(priorityOrder, pa), rank(priorityOrder, pb); ra != rb {
			return ra < rb
		}

		if ra, rb := rank(typeOrder, a.Type), rank(typeOrder, b.Type); ra != rb {
			return ra < rb
		}

		return strings.ToLower(a.Title) < strings.ToLower(b.Title)
	})
}

// rank returns the sort position of val in order, or len(order) (sorts
// last) if val is not a recognized value.
func rank(order map[string]int, val string) int {
	if r, ok := order[val]; ok {
		return r
	}
	return len(order)
}

// SortBeans is the exported single-source sort (Status -> Priority -> Type ->
// Title) every consumer OUTSIDE this package must use for bean lists it
// itself assembles (e.g. resolved Blocking/BlockedBy IDs) -- I03 (bean
// bt-7jr8 T8-review): a second, ad-hoc tie-break in the tui package would
// violate the "single place sort order is defined" contract this file
// already asserts for its own callers.
func SortBeans(beans []*Bean) { sortBeans(beans) }

// StatusRank exposes the status tier position (see statusOrder) for callers
// that need to compare two beans' status without a full SortBeans call
// (e.g. the Backlog sort-toggle, E2 Task 5).
func StatusRank(status string) int { return rank(statusOrder, status) }

// PriorityRank exposes the priority tier position, empty treated as "normal"
// (mirrors sortBeans' own empty-priority handling).
func PriorityRank(priority string) int {
	if priority == "" {
		priority = "normal"
	}
	return rank(priorityOrder, priority)
}
