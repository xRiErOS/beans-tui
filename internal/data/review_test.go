package data

// review_test.go — TDD coverage for EpicAncestor (E4 Task 3, bean bt-hxyo,
// epic-E4-plan.md »Task 3« Step 1).

import (
	"testing"
	"time"
)

func TestEpicAncestorFindsNearestEpic(t *testing.T) {
	beans := []Bean{
		{ID: "ms", Title: "Milestone", Type: "milestone"},
		{ID: "ep", Title: "Epic", Type: "epic", Parent: "ms"},
		{ID: "tsk", Title: "Task", Type: "task", Parent: "ep"},
	}
	idx := NewIndex(beans)

	epic, ok := EpicAncestor(idx, idx.ByID["tsk"])
	if !ok {
		t.Fatal("EpicAncestor(tsk) ok = false, want true")
	}
	if epic.ID != "ep" {
		t.Errorf("EpicAncestor(tsk) = %q, want ep", epic.ID)
	}
}

func TestEpicAncestorNoneWhenMilestoneDirectParent(t *testing.T) {
	beans := []Bean{
		{ID: "ms", Title: "Milestone", Type: "milestone"},
		{ID: "tsk", Title: "Task", Type: "task", Parent: "ms"},
	}
	idx := NewIndex(beans)

	epic, ok := EpicAncestor(idx, idx.ByID["tsk"])
	if ok || epic != nil {
		t.Errorf("EpicAncestor(tsk, milestone-direct-parent) = (%v, %v), want (nil, false)", epic, ok)
	}
}

func TestEpicAncestorNoneWhenParentless(t *testing.T) {
	beans := []Bean{
		{ID: "tsk", Title: "Task", Type: "task"},
	}
	idx := NewIndex(beans)

	epic, ok := EpicAncestor(idx, idx.ByID["tsk"])
	if ok || epic != nil {
		t.Errorf("EpicAncestor(tsk, parentless) = (%v, %v), want (nil, false)", epic, ok)
	}
}

// TestEpicAncestorSurvivesParentCycle guards the defensive visited guard: a
// hand-edited Parent-Zyklus (a's parent is b, b's parent is a) must terminate,
// never hang -- run on a goroutine with a timeout so a regression (missing
// visited guard) fails the test instead of hanging the whole suite forever
// (mirrors hierarchy_test.go's TestCollectDescendantsSurvivesParentCycle).
func TestEpicAncestorSurvivesParentCycle(t *testing.T) {
	beans := []Bean{
		{ID: "a", Title: "A", Type: "task", Parent: "b"},
		{ID: "b", Title: "B", Type: "task", Parent: "a"},
	}
	idx := NewIndex(beans)

	type result struct {
		epic *Bean
		ok   bool
	}
	done := make(chan result, 1)
	go func() {
		epic, ok := EpicAncestor(idx, idx.ByID["a"])
		done <- result{epic, ok}
	}()

	select {
	case r := <-done:
		if r.ok || r.epic != nil {
			t.Errorf("EpicAncestor on a parent cycle = (%v, %v), want (nil, false)", r.epic, r.ok)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("EpicAncestor hung on a parent cycle (missing visited guard)")
	}
}

// TestEpicAncestorNoneWhenDanglingParent guards a dangling (unresolved)
// Parent ID interrupting the walk before any epic is found -- beans-legal
// (frontmatter is hand-editable), must resolve to the "(kein Epic)" bucket
// rather than panicking on a missing idx.ByID entry.
func TestEpicAncestorNoneWhenDanglingParent(t *testing.T) {
	beans := []Bean{
		{ID: "tsk", Title: "Task", Type: "task", Parent: "gone"},
	}
	idx := NewIndex(beans)

	epic, ok := EpicAncestor(idx, idx.ByID["tsk"])
	if ok || epic != nil {
		t.Errorf("EpicAncestor(tsk, dangling parent) = (%v, %v), want (nil, false)", epic, ok)
	}
}
