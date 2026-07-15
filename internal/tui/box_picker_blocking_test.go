package tui

// box_picker_blocking_test.go — TDD coverage for the Blocking-Picker (`B`,
// E3 Task 3, bean bt-p1uz): toggle-multi-select over every OTHER bean
// (design decision g: deliberately NO cycle/descendant exclusion, port-
// parity with beans-src blockingpicker.go). Pending-Diff pattern ported
// verbatim from box_picker_tag.go (T2) -- space toggles pending, enter diffs
// pending against original and fires ONE combined data.SetBlocking
// mutateCmd, esc discards.

import (
	"strings"
	"testing"

	"beans-tui/internal/data"
	tea "github.com/charmbracelet/bubbletea"
)

// --- buildBlockingItems / openBlockingPicker ---

// TestBlockingPickerExcludesOnlySelf guards design decision g directly:
// focused on ep-1 (whose own CHILDREN are tk-1/tk-2, fixtureBeans), the
// picker must offer tk-1 AND tk-2 (its own descendants!) alongside ms-1 --
// only ep-1 itself is missing. This is the exact opposite of the
// Parent-Picker's descendant exclusion (box_picker_parent_test.go), proving
// the two pickers really do use different eligibility rules.
func TestBlockingPickerExcludesOnlySelf(t *testing.T) {
	m := fixtureModel(t, fixtureBeans()) // ms-1 -> ep-1 -> tk-1, tk-2
	m = focusBean(m, "ep-1")

	m = step(t, m, runeMsg('B'))

	if m.overlay != overlayBlockingPicker {
		t.Fatalf("overlay = %v, want overlayBlockingPicker", m.overlay)
	}
	ids := pickerItemIDs(m.blockItems)
	for _, want := range []string{"ms-1", "tk-1", "tk-2"} {
		if !ids[want] {
			t.Errorf("blockItems missing %q, got %+v", want, m.blockItems)
		}
	}
	if ids["ep-1"] {
		t.Fatal("blockItems must exclude the focused bean itself (ep-1)")
	}
	if len(m.blockItems) != 3 {
		t.Fatalf("len(blockItems) = %d, want 3 (every other bean, no cycle exclusion), got %+v", len(m.blockItems), m.blockItems)
	}
}

// TestBlockingPickerSeedsPendingFromBlockingField guards the wholesale-
// replace + independence convention (mirrors
// TestOpenTagPickerSeedsPendingFromFocusedBean): bean-a's Blocking field
// (fixtureBeansWithBlocking, update_test.go) seeds BOTH blockOriginal and
// blockPending, as two independent maps.
func TestBlockingPickerSeedsPendingFromFocusedBean(t *testing.T) {
	m := fixtureModel(t, fixtureBeansWithBlocking()) // bean-a Blocking: [bean-b]
	m = focusBeanFull(m, "bean-a")

	m = step(t, m, runeMsg('B'))

	if m.overlay != overlayBlockingPicker {
		t.Fatalf("overlay = %v, want overlayBlockingPicker", m.overlay)
	}
	if m.mutTarget != "bean-a" {
		t.Fatalf("mutTarget = %q, want bean-a", m.mutTarget)
	}
	if len(m.blockOriginal) != 1 || !m.blockOriginal["bean-b"] {
		t.Fatalf("blockOriginal = %v, want {bean-b:true}", m.blockOriginal)
	}
	if len(m.blockPending) != 1 || !m.blockPending["bean-b"] {
		t.Fatalf("blockPending = %v, want {bean-b:true}", m.blockPending)
	}

	m.blockPending["scratch-independence-probe"] = true
	if m.blockOriginal["scratch-independence-probe"] {
		t.Fatal("blockOriginal aliases blockPending's backing map -- must be independent")
	}
}

// --- keyBlockingPicker: toggle ---

func TestBlockingPickerToggleFlipsPendingOnly(t *testing.T) {
	m := fixtureModel(t, fixtureBeansWithBlocking())
	m = focusBeanFull(m, "bean-a") // Blocking: [bean-b]
	m = step(t, m, runeMsg('B'))

	cursorTo := func(m model, id string) model {
		for i, it := range m.blockItems {
			if it.id == id {
				m.menu.cursor = i
			}
		}
		return m
	}

	m = cursorTo(m, "bean-b")
	m = step(t, m, runeMsg(' '))
	if m.blockPending["bean-b"] {
		t.Fatal("space did not toggle bean-b OFF in blockPending")
	}
	if !m.blockOriginal["bean-b"] {
		t.Fatal("blockOriginal must stay unchanged by a pending toggle")
	}

	m = cursorTo(m, "ep-1")
	m = step(t, m, runeMsg('x'))
	if !m.blockPending["ep-1"] {
		t.Fatal("x did not toggle ep-1 ON in blockPending")
	}
	if m.blockOriginal["ep-1"] {
		t.Fatal("blockOriginal must stay unchanged (ep-1 was never original)")
	}
}

// --- keyBlockingPicker: enter diffs via ONE SetBlocking call ---

// TestBlockingPickerEnterDiffsViaSetBlocking mirrors
// TestTagPickerEnterDiffsFireOneSetTagsCall's verification shape exactly: a
// real data.Client pointed at a nonexistent repo dir (no beans binary
// required), cmd() resolves DIRECTLY to a mutationDoneMsg (not a
// tea.BatchMsg), whose error text names the dispatched CLI subcommand.
func TestBlockingPickerEnterDiffsViaSetBlocking(t *testing.T) {
	m := fixtureModel(t, fixtureBeansWithBlocking())
	m.client = &data.Client{RepoDir: "/nonexistent-bt-e3-t3-scratch-dir"}
	m = focusBeanFull(m, "bean-a") // Blocking: [bean-b]
	m = step(t, m, runeMsg('B'))

	for i, it := range m.blockItems {
		switch it.id {
		case "bean-b": // originally present -> toggle off = remove
			m.menu.cursor = i
		}
	}
	m = step(t, m, runeMsg(' '))
	for i, it := range m.blockItems {
		if it.id == "ep-1" { // not originally present -> toggle on = add
			m.menu.cursor = i
		}
	}
	m = step(t, m, runeMsg(' '))

	tm, cmd := m.Update(keyMsg(tea.KeyEnter))
	nm, ok := tm.(model)
	if !ok {
		t.Fatalf("Update(enter) did not return a model, got %T", tm)
	}
	if nm.overlay != overlayNone {
		t.Fatalf("overlay after enter = %v, want overlayNone", nm.overlay)
	}
	if cmd == nil {
		t.Fatal("enter with pending changes must return a Cmd")
	}

	msg := cmd()
	mdm, ok := msg.(mutationDoneMsg)
	if !ok {
		t.Fatalf("cmd() = %T, want mutationDoneMsg (ONE combined SetBlocking call, not tea.Batch)", msg)
	}
	if mdm.err == nil || !strings.Contains(mdm.err.Error(), "beans update") {
		t.Fatalf("mutationDoneMsg.err = %v, want an error containing %q (proves SetBlocking dispatched)", mdm.err, "beans update")
	}
}

func TestBlockingPickerEnterNoChangesNoMutation(t *testing.T) {
	m := fixtureModel(t, fixtureBeansWithBlocking())
	m = focusBeanFull(m, "bean-a")
	m = step(t, m, runeMsg('B'))

	tm, cmd := m.Update(keyMsg(tea.KeyEnter))
	nm := tm.(model)
	if nm.overlay != overlayNone {
		t.Fatal("enter must close the picker even with no changes")
	}
	if cmd != nil {
		t.Fatal("enter with no pending changes must not fire a mutation Cmd")
	}
}

// --- keyBlockingPicker: esc discards ---

func TestBlockingPickerEscDiscards(t *testing.T) {
	m := fixtureModel(t, fixtureBeansWithBlocking())
	m = focusBeanFull(m, "bean-a")
	m = step(t, m, runeMsg('B'))
	for i, it := range m.blockItems {
		if it.id == "ep-1" {
			m.menu.cursor = i
		}
	}
	m = step(t, m, runeMsg(' ')) // toggle ep-1 on in pending

	tm, cmd := m.Update(keyMsg(tea.KeyEsc))
	nm := tm.(model)
	if nm.overlay != overlayNone {
		t.Fatal("esc did not close the blocking picker")
	}
	if cmd != nil {
		t.Fatal("esc must not fire a mutation Cmd")
	}
}
