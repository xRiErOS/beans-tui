package data

import (
	"sort"
	"strings"
)

// Index is an immutable, in-memory snapshot over a set of beans, built once
// per load/reload from Client.List() output. Views (tree, backlog, tag
// filters) consume only Index methods -- there is no incremental update
// path (YAGNI): a reload just builds a fresh Index.
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
var (
	statusOrder = map[string]int{
		"in-progress": 0,
		"todo":        1,
		"draft":       2,
		"completed":   3,
		"scrapped":    4,
	}

	priorityOrder = map[string]int{
		"critical": 0,
		"high":     1,
		"normal":   2,
		"low":      3,
		"deferred": 4,
	}

	typeOrder = map[string]int{
		"milestone": 0,
		"epic":      1,
		"bug":       2,
		"feature":   3,
		"task":      4,
	}
)

// sortBeans sorts in place by Status -> Priority -> Type -> Title
// (case-insensitive), matching beans upstream ordering. This is the single
// place sort order is defined; every Index method reuses it. Empty
// priority is treated as "normal" (beans 0.4.2 default).
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
