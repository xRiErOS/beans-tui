package data

// review.go — Review-Queue-Ableitung, Datenlayer-Hälfte (E4 Task 3, bean
// bt-hxyo, design decision c): EpicAncestor is the sole new primitive here --
// a pure, read-only Parent-chain walk. The TUI-side grouping (reviewQueue/
// reviewRework/reviewFlat) used to live in internal/tui/view_review_cockpit.go,
// consuming this + idx.WithTag("to-review"/"rework") (E1 Task 3) -- that view
// was removed by E7 T1 (PF-14, bean bt-wmtb, 2026-07-15: Review-Cockpit
// replaced by a Chat-driven Tag-Trio, design-spec.md §5). EpicAncestor itself
// stays (YAGNI, same rationale as data.Client.PassReview/RejectReview,
// mutations.go): harmless, pure, no TUI coupling to carry, no caller left.

// EpicAncestor walks b.Parent upward (visited-guarded against a Parent
// cycle, same defensive shape as CollectDescendants/expandAncestorsOf --
// hierarchy.go / internal/tui/update.go's expandAncestorsOf) until it finds
// the nearest ancestor with Type == "epic". ok=false covers all three of "no
// parent at all", "a parent chain that never resolves through to an epic"
// (e.g. a bean parented directly under a milestone), AND "a chain that hits a
// dangling (unresolved) parent ID before ever reaching an epic" -- design
// decision c collapses every one of these into the SAME "(kein Epic)" bucket
// at the call site, so this function does not need to distinguish them
// itself, only report whether an epic was actually found.
func EpicAncestor(idx *Index, b *Bean) (epic *Bean, ok bool) {
	if idx == nil || b == nil {
		return nil, false
	}
	visited := map[string]bool{b.ID: true}
	cur := b
	for cur.Parent != "" && !visited[cur.Parent] {
		visited[cur.Parent] = true
		parent, exists := idx.ByID[cur.Parent]
		if !exists {
			return nil, false // dangling parent -- beans-legal, never hangs, just no epic found
		}
		if parent.Type == "epic" {
			return parent, true
		}
		cur = parent
	}
	return nil, false
}
