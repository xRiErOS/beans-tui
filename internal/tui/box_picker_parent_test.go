package tui

// box_picker_parent_test.go — TDD coverage for the Parent-Picker (`a`, E3
// Task 3, bean bt-p1uz): single-select over data.EligibleParents (self/
// descendants/invalid-types pre-filtered, design decision f) plus a
// "(Kein Parent)" clear row pinned first. Immediate-apply Enter semantics
// (like box_menu_value.go's value menu, design decision a3) -- NOT the
// Pending-Diff pattern box_picker_tag.go/box_picker_blocking.go use, since a
// bean has exactly one parent, nothing to diff.

import (
	"strings"
	"testing"

	"beans-tui/internal/data"
	tea "github.com/charmbracelet/bubbletea"
)

// fixtureBeansForParentPicker is a deliberately deeper/wider hierarchy than
// fixtureBeans() (update_test.go): it needs a "feature" type (absent from
// fixtureBeans) AND a hand-edited structural oddity -- ms-2, a MILESTONE
// nested under ep-1 -- to make the descendant-exclusion half of
// TestParentPickerExcludesSelfDescendantsAndInvalidTypes meaningful: ms-2's
// TYPE (milestone) would otherwise be perfectly eligible for an epic's
// parent, so only CollectDescendants (not the type filter) can be excluding
// it. Beans-legal per this app's own philosophy (hand-editable frontmatter,
// tolerated the same way collectCycleOrphans/appendBeanNode's cycle guards
// already tolerate structurally odd hierarchies, view_browse_repo.go).
func fixtureBeansForParentPicker() []data.Bean {
	return []data.Bean{
		{ID: "ms-1", Title: "Milestone One", Status: "todo", Type: "milestone", Priority: "normal"},
		{ID: "ep-1", Title: "Epic One", Status: "todo", Type: "epic", Priority: "normal", Parent: "ms-1"},
		{ID: "ms-2", Title: "Milestone Two", Status: "todo", Type: "milestone", Priority: "normal", Parent: "ep-1"},
		{ID: "ft-1", Title: "Feature One", Status: "todo", Type: "feature", Priority: "normal", Parent: "ep-1"},
		{ID: "tk-1", Title: "Task One", Status: "todo", Type: "task", Priority: "normal", Parent: "ft-1"},
	}
}

// focusBeanFull expands EVERY ancestor of id (walking Parent up through
// m.idx) so id is visible in m.visibleNodes() -- a generic version of
// box_menu_value_test.go's focusBean, which hardcodes fixtureBeans' fixed
// ms-1/ep-1 ancestor set. This file's fixture has a deeper/wider hierarchy,
// so each test walks its own bean's real ancestor chain instead.
func focusBeanFull(m model, id string) model {
	expanded := map[string]bool{}
	cur := id
	for {
		b, ok := m.idx.ByID[cur]
		if !ok || b.Parent == "" {
			break
		}
		expanded[b.Parent] = true
		cur = b.Parent
	}
	m.expanded = expanded
	m.cursorID = id
	return m
}

// pickerItemIDs is a small assertion helper -- the set of ids in items.
func pickerItemIDs(items []pickerItem) map[string]bool {
	out := make(map[string]bool, len(items))
	for _, it := range items {
		out[it.id] = true
	}
	return out
}

// --- openParentPicker ---

// TestParentPickerExcludesSelfDescendantsAndInvalidTypes guards the
// combined eligibility rule (design decision f): focused on ep-1 (an epic,
// validParentTypes -> [milestone] only), the picker must offer ms-1 and
// NOTHING else -- ep-1 itself (self), ms-2 (a descendant, DESPITE being a
// type-eligible milestone), ft-1 and tk-1 (both descendants AND wrong type)
// are all excluded.
func TestParentPickerExcludesSelfDescendantsAndInvalidTypes(t *testing.T) {
	m := fixtureModel(t, fixtureBeansForParentPicker())
	m = focusBeanFull(m, "ep-1")

	m = step(t, m, runeMsg('a'))

	if m.overlay != overlayParentPicker {
		t.Fatalf("overlay = %v, want overlayParentPicker", m.overlay)
	}
	ids := pickerItemIDs(m.parentItems)
	if !ids["ms-1"] {
		t.Fatalf("parentItems missing ms-1 (eligible: type milestone, not a descendant), got %+v", m.parentItems)
	}
	for _, excluded := range []string{"ep-1", "ms-2", "ft-1", "tk-1"} {
		if ids[excluded] {
			t.Errorf("parentItems must exclude %q, got %+v", excluded, m.parentItems)
		}
	}
	// self + descendants + invalid types all excluded -> only the clear row
	// ("") plus ms-1 remain.
	if len(m.parentItems) != 2 {
		t.Fatalf("len(parentItems) = %d, want 2 (clear row + ms-1), got %+v", len(m.parentItems), m.parentItems)
	}
}

// TestParentPickerMilestoneShowsNoEligibleParents guards the nil-
// validParentTypes short-circuit: a milestone can never take a parent, so
// only the "(Kein Parent)" clear row remains.
func TestParentPickerMilestoneShowsNoEligibleParents(t *testing.T) {
	m := fixtureModel(t, fixtureBeansForParentPicker())
	m = focusBeanFull(m, "ms-1")

	m = step(t, m, runeMsg('a'))

	if len(m.parentItems) != 1 {
		t.Fatalf("len(parentItems) = %d, want 1 (clear row only), got %+v", len(m.parentItems), m.parentItems)
	}
	if m.parentItems[0].id != "" {
		t.Fatalf("parentItems[0].id = %q, want \"\" (clear row)", m.parentItems[0].id)
	}
}

// TestParentPickerClearRowFirstAndCursorOnCurrentParent guards two upstream
// invariants at once (port beans-src parentpicker.go clearParentItem +
// newParentPickerModel's selectedIndex seed): the clear row is ALWAYS index
// 0, and the cursor starts on the focused bean's CURRENT parent's row.
func TestParentPickerClearRowFirstAndCursorOnCurrentParent(t *testing.T) {
	m := fixtureModel(t, fixtureBeansForParentPicker())
	m = focusBeanFull(m, "tk-1") // parent: ft-1

	m = step(t, m, runeMsg('a'))

	if len(m.parentItems) == 0 || m.parentItems[0].id != "" {
		t.Fatalf("parentItems[0] = %+v, want the \"(Kein Parent)\" clear row first", m.parentItems)
	}
	wantIdx := -1
	for i, it := range m.parentItems {
		if it.id == "ft-1" {
			wantIdx = i
		}
	}
	if wantIdx < 0 {
		t.Fatalf("ft-1 (current parent) not found in parentItems %+v", m.parentItems)
	}
	if m.menu.cursor != wantIdx {
		t.Fatalf("menu.cursor = %d, want %d (ft-1's row)", m.menu.cursor, wantIdx)
	}
}

// --- keyParentPicker: enter applies immediately ---

// TestParentPickerEnterAppliesSetParentOrRemoveParent covers both branches
// of the immediate-apply Enter (design decision a3 parity): cursoring onto a
// real bean row fires SetParent, cursoring onto the clear row fires
// RemoveParent. Verified the same way box_picker_tag.go's
// TestTagPickerEnterDiffsFireOneSetTagsCall is: a real data.Client pointed
// at a nonexistent repo dir, no beans binary required -- cmd() resolves
// DIRECTLY to a mutationDoneMsg whose error text names the dispatched CLI
// subcommand.
func TestParentPickerEnterAppliesSetParentOrRemoveParent(t *testing.T) {
	t.Run("SetParent", func(t *testing.T) {
		m := fixtureModel(t, fixtureBeansForParentPicker())
		m.client = &data.Client{RepoDir: "/nonexistent-bt-e3-t3-scratch-dir"}
		m = focusBeanFull(m, "ep-1")
		m = step(t, m, runeMsg('a'))

		for i, it := range m.parentItems {
			if it.id == "ms-1" {
				m.menu.cursor = i
			}
		}

		tm, cmd := m.Update(keyMsg(tea.KeyEnter))
		nm, ok := tm.(model)
		if !ok {
			t.Fatalf("Update(enter) did not return a model, got %T", tm)
		}
		if nm.overlay != overlayNone {
			t.Fatalf("overlay after enter = %v, want overlayNone", nm.overlay)
		}
		if cmd == nil {
			t.Fatal("enter on a real row must return a Cmd")
		}
		msg := cmd()
		mdm, ok := msg.(mutationDoneMsg)
		if !ok {
			t.Fatalf("cmd() = %T, want mutationDoneMsg", msg)
		}
		if mdm.err == nil || !strings.Contains(mdm.err.Error(), "beans update") {
			t.Fatalf("mutationDoneMsg.err = %v, want an error containing %q (proves SetParent dispatched)", mdm.err, "beans update")
		}
	})

	t.Run("RemoveParent", func(t *testing.T) {
		m := fixtureModel(t, fixtureBeansForParentPicker())
		m.client = &data.Client{RepoDir: "/nonexistent-bt-e3-t3-scratch-dir"}
		m = focusBeanFull(m, "ep-1")
		m = step(t, m, runeMsg('a'))
		m.menu.cursor = 0 // the clear row

		tm, cmd := m.Update(keyMsg(tea.KeyEnter))
		nm := tm.(model)
		if nm.overlay != overlayNone {
			t.Fatalf("overlay after enter = %v, want overlayNone", nm.overlay)
		}
		if cmd == nil {
			t.Fatal("enter on the clear row must return a Cmd")
		}
		msg := cmd()
		mdm, ok := msg.(mutationDoneMsg)
		if !ok {
			t.Fatalf("cmd() = %T, want mutationDoneMsg", msg)
		}
		if mdm.err == nil || !strings.Contains(mdm.err.Error(), "beans update") {
			t.Fatalf("mutationDoneMsg.err = %v, want an error containing %q (proves RemoveParent dispatched)", mdm.err, "beans update")
		}
	})
}

// --- keyParentPicker: esc discards ---

func TestParentPickerEscNoMutation(t *testing.T) {
	m := fixtureModel(t, fixtureBeansForParentPicker())
	m = focusBeanFull(m, "ep-1")
	m = step(t, m, runeMsg('a'))

	tm, cmd := m.Update(keyMsg(tea.KeyEsc))
	nm := tm.(model)
	if nm.overlay != overlayNone {
		t.Fatal("esc did not close the parent picker")
	}
	if cmd != nil {
		t.Fatal("esc must not fire a mutation Cmd")
	}
}

// --- vanished-target guard on enter (E3-T3-Review PFLICHT, carried into
// bean bt-ppzb/E3 Task 6: "Vanished-mutTarget-Regressionstests für Parent-
// UND Blocking-Picker (Muster TestValueMenuTargetVanished...)") ---

// TestParentPickerEnterTargetVanishedClosesGracefully mirrors
// TestValueMenuTargetVanishedClosesGracefully/
// TestTagPickerEnterTargetVanishedClosesGracefully (design decision d): the
// focused bean disappears (external delete + reload) between open and enter
// -- enter must close the overlay and set a status-line note, WITHOUT firing
// a doomed SetParent/RemoveParent.
func TestParentPickerEnterTargetVanishedClosesGracefully(t *testing.T) {
	beans := fixtureBeansForParentPicker()
	m := fixtureModel(t, beans)
	m = focusBeanFull(m, "ep-1")
	m = step(t, m, runeMsg('a'))
	if m.overlay != overlayParentPicker {
		t.Fatal("setup: a did not open the parent picker")
	}

	// ep-1 vanishes externally; a reload lands while the picker is still open
	// (m.overlay survives a reload untouched, applyLoaded never touches it).
	var remaining []data.Bean
	for _, b := range beans {
		if b.ID != "ep-1" {
			remaining = append(remaining, b)
		}
	}
	m = step(t, m, beansLoadedMsg{beans: remaining})

	tm, cmd := m.Update(keyMsg(tea.KeyEnter))
	nm := tm.(model)
	if nm.overlay != overlayNone {
		t.Fatal("enter on a vanished target must close the overlay")
	}
	if nm.err == "" {
		t.Fatal("enter on a vanished target must set a status-line note (m.err)")
	}
	if cmd != nil {
		t.Fatal("enter on a vanished target must not fire a Cmd (no doomed mutation)")
	}
}

// --- current-parent-out-of-eligibility cursor fallback (E3-T3-Review
// PFLICHT, carried into bean bt-ppzb/E3 Task 6: "Test: current-parent-out-
// of-eligibility -> Cursor-Fallback 0") ---

// TestParentPickerCurrentParentOutOfEligibilityCursorFallsBackToZero guards
// openParentPicker's cursor seed (box_picker_parent.go) for a bean whose
// CURRENT Parent field is set but points at a bean that is beans-legal
// (hand-editable frontmatter) yet not a legal parent TYPE for it --
// tk-1.Parent is hand-edited to tk-9, ANOTHER task (validParentTypes("task")
// = [milestone, epic, feature] -- a task can never legally parent a task),
// so tk-9 never appears in parentItems at all. The seed loop
// (openParentPicker) must fall back to its listState{} zero value (index 0,
// the "(Kein Parent)" clear row) rather than leaving the cursor on a stale
// position -- this is already the loop's own documented behavior (no
// explicit "not found" branch needed), pinned here as an end-to-end
// regression.
func TestParentPickerCurrentParentOutOfEligibilityCursorFallsBackToZero(t *testing.T) {
	beans := append(fixtureBeansForParentPicker(),
		data.Bean{ID: "tk-9", Title: "Task Nine", Status: "todo", Type: "task", Priority: "normal", Parent: "ft-1"},
	)
	for i := range beans {
		if beans[i].ID == "tk-1" {
			beans[i].Parent = "tk-9" // out-of-eligibility: a task can't legally parent a task
		}
	}
	m := fixtureModel(t, beans)
	m = focusBeanFull(m, "tk-1")

	m = step(t, m, runeMsg('a'))

	ids := pickerItemIDs(m.parentItems)
	if ids["tk-9"] {
		t.Fatalf("parentItems must exclude tk-9 (wrong type for a task's parent), got %+v", m.parentItems)
	}
	if m.menu.cursor != 0 {
		t.Fatalf("menu.cursor = %d, want 0 (fallback: current parent tk-9 is not in parentItems)", m.menu.cursor)
	}
}
