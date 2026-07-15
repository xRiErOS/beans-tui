package tui

// box_confirm_delete_test.go — TDD coverage for the Delete-Confirm modal
// (`d`, E3 Task 6, bean bt-ppzb): Kinder-Count captured SYNCHRONOUSLY from
// idx.Children (no async preview-load like devd box_confirm_delete.go), the
// orphaning wording (NOT a cascade-delete claim), enter/esc dispatch, and the
// cursor-clamp-after-delete end-to-end proof (reusing applyLoaded's EXISTING
// oldPos fallback, no new cursor logic).

import (
	"strings"
	"testing"

	"beans-tui/internal/data"
	tea "github.com/charmbracelet/bubbletea"
)

// --- openDeleteConfirm ---

// TestOpenDeleteConfirmCountsChildrenSynchronously guards `d` on a bean with
// children: delChildren is idx.Children[id]'s DIRECT count (ep-1 has two
// children, tk-1/tk-2, fixtureBeans), read straight from the in-memory
// index -- no loading state, no async Cmd (types.go/box_confirm_delete.go
// doc-stamps: deliberately NOT data.CollectDescendants's recursive walk,
// since only DIRECT children orphan when their parent is deleted).
func TestOpenDeleteConfirmCountsChildrenSynchronously(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m = focusBean(m, "ep-1")

	tm, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	nm, ok := tm.(model)
	if !ok {
		t.Fatalf("Update('d') did not return a model, got %T", tm)
	}
	if cmd != nil {
		t.Fatal("opening the delete confirm must not itself fire a Cmd")
	}
	if nm.overlay != overlayDeleteConfirm {
		t.Fatalf("overlay = %v, want overlayDeleteConfirm", nm.overlay)
	}
	if nm.mutTarget != "ep-1" {
		t.Fatalf("mutTarget = %q, want ep-1", nm.mutTarget)
	}
	if nm.delTitle != "Epic One" {
		t.Fatalf("delTitle = %q, want %q", nm.delTitle, "Epic One")
	}
	if nm.delChildren != 2 {
		t.Fatalf("delChildren = %d, want 2 (tk-1, tk-2)", nm.delChildren)
	}
}

// TestOpenDeleteConfirmLeafHasZeroChildren guards the other half: a leaf
// (tk-1, no children) opens with delChildren == 0, so deleteBox's orphan
// warning line does not render at all (TestDeleteBoxLeafOmitsOrphanWarning
// below).
func TestOpenDeleteConfirmLeafHasZeroChildren(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m = focusBean(m, "tk-1")

	m = step(t, m, runeMsg('d'))
	if m.overlay != overlayDeleteConfirm {
		t.Fatalf("overlay = %v, want overlayDeleteConfirm", m.overlay)
	}
	if m.delChildren != 0 {
		t.Fatalf("delChildren = %d, want 0 (tk-1 is a leaf)", m.delChildren)
	}
}

// --- deleteBox ---

// TestDeleteBoxWarnsChildrenLoseParentBecomeRoots guards the semantic
// deviation from devd (bean bt-ppzb Ziel): `beans delete` does NOT cascade
// -- the modal must never claim children get deleted too ("gelöscht"). The
// exact wording ("verlieren den Parent — werden zu eigenen Wurzeln") is
// itself an ERRATUM correction (box_confirm_delete.go's own doc-stamp): the
// original "become '(verwaist)'-bucket orphans" assumption was empirically
// WRONG -- beans 0.4.2 clears the parent reference outright rather than
// leaving it dangling, so this test also guards against the STALE "verwaist"
// wording reappearing.
func TestDeleteBoxWarnsChildrenLoseParentBecomeRoots(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m = focusBean(m, "ep-1")
	m = step(t, m, runeMsg('d'))

	box := m.deleteBox()
	if !strings.Contains(box, "verlieren den Parent") {
		t.Fatalf("deleteBox() = %q, want it to mention the children losing their parent (not \"verwaist\" -- ERRATUM, beans clears the reference outright)", box)
	}
	if strings.Contains(box, "verwaist") {
		t.Fatalf("deleteBox() = %q, must NOT use the stale \"verwaist\" wording -- beans 0.4.2 clears the parent reference, it does not leave a dangling one", box)
	}
	if strings.Contains(box, "gelöscht") {
		t.Fatalf("deleteBox() = %q, must NOT claim children get deleted (\"gelöscht\") -- beans delete does not cascade", box)
	}
	if !strings.Contains(box, "2") {
		t.Fatalf("deleteBox() = %q, want the child count (2) rendered", box)
	}
	if !strings.Contains(box, "Epic One") {
		t.Fatalf("deleteBox() = %q, want the bean's title rendered", box)
	}
	if !strings.Contains(box, "Irreversible") {
		t.Fatalf("deleteBox() = %q, want the Irreversible warning", box)
	}
}

// TestDeleteBoxLeafOmitsChildrenWarning guards the delChildren==0 branch: a
// leaf's confirm modal must not render a children-count line at all.
func TestDeleteBoxLeafOmitsChildrenWarning(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m = focusBean(m, "tk-1")
	m = step(t, m, runeMsg('d'))

	box := m.deleteBox()
	if strings.Contains(box, "verlieren den Parent") {
		t.Fatalf("deleteBox() for a leaf = %q, must not mention children losing a parent (delChildren == 0)", box)
	}
	if !strings.Contains(box, "Task One") {
		t.Fatalf("deleteBox() = %q, want the bean's title rendered", box)
	}
}

// --- keyDeleteConfirm ---

// TestDeleteConfirmEnterFiresDeleteAndCloses guards enter: closes the
// overlay and returns a Cmd whose resolved mutationDoneMsg's error names
// "beans delete" (proves data.Client.Delete dispatched, NO --if-match --
// mutations.go's Delete signature takes no etag at all), verified the same
// no-binary-required way box_picker_parent_test.go/box_confirm_create_test.go
// already use.
func TestDeleteConfirmEnterFiresDeleteAndCloses(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m.client = &data.Client{RepoDir: "/nonexistent-bt-e3-t6-scratch-dir"}
	m = focusBean(m, "tk-1")
	m = step(t, m, runeMsg('d'))

	tm, cmd := m.Update(keyMsg(tea.KeyEnter))
	nm, ok := tm.(model)
	if !ok {
		t.Fatalf("Update(enter) did not return a model, got %T", tm)
	}
	if nm.overlay != overlayNone {
		t.Fatalf("overlay after enter = %v, want overlayNone", nm.overlay)
	}
	if cmd == nil {
		t.Fatal("enter must return the delete mutateCmd")
	}
	msg := cmd()
	mdm, ok := msg.(mutationDoneMsg)
	if !ok {
		t.Fatalf("cmd() = %T, want mutationDoneMsg", msg)
	}
	if mdm.err == nil || !strings.Contains(mdm.err.Error(), "beans delete") {
		t.Fatalf("mutationDoneMsg.err = %v, want an error containing %q (proves Delete dispatched)", mdm.err, "beans delete")
	}
}

// TestDeleteConfirmEscCancels guards esc: closes the overlay without firing
// any Cmd.
func TestDeleteConfirmEscCancels(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m = focusBean(m, "tk-1")
	m = step(t, m, runeMsg('d'))

	tm, cmd := m.Update(keyMsg(tea.KeyEsc))
	nm := tm.(model)
	if nm.overlay != overlayNone {
		t.Fatal("esc did not close the delete confirm")
	}
	if cmd != nil {
		t.Fatal("esc must not fire a mutation Cmd")
	}
}

// TestDeleteConfirmNCancels guards the devd-parity "n" shorthand (Port
// box_confirm_delete.go's own esc/n cancel, DD2-174) alongside keys.Back.
func TestDeleteConfirmNCancels(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m = focusBean(m, "tk-1")
	m = step(t, m, runeMsg('d'))

	tm, cmd := m.Update(runeMsg('n'))
	nm := tm.(model)
	if nm.overlay != overlayNone {
		t.Fatal("n did not close the delete confirm")
	}
	if cmd != nil {
		t.Fatal("n must not fire a mutation Cmd")
	}
}

// --- vanished-target guard on enter (design decision d parity) ---

// TestDeleteConfirmEnterTargetVanishedClosesGracefully mirrors every other
// overlay's vanished-mutTarget regression (TestValueMenuTargetVanishedClose
// sGracefully, TestTagPickerEnterTargetVanishedClosesGracefully): the focused
// bean disappears (external delete + reload) between open and enter -- enter
// must close the overlay and set a status-line note, WITHOUT firing a doomed
// second Delete.
func TestDeleteConfirmEnterTargetVanishedClosesGracefully(t *testing.T) {
	beans := fixtureBeans()
	m := fixtureModel(t, beans)
	m = focusBean(m, "tk-2")
	m = step(t, m, runeMsg('d'))
	if m.overlay != overlayDeleteConfirm {
		t.Fatal("setup: d did not open the delete confirm")
	}

	remaining := []data.Bean{beans[0], beans[1], beans[2]} // ms-1, ep-1, tk-1 -- no tk-2
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

// --- cursor clamp after delete (reuse, no new logic) ---

// TestDeleteCursorClampsViaApplyLoadedOldPos guards the plan's explicit
// "keine neue Cursor-Logik" requirement: cursor parked on the LAST visible
// node (tk-2), delete it, reload without it -- the cursor must clamp near
// its old spot (applyLoaded's EXISTING oldPos fallback, update.go/T8) rather
// than resetting to the top. Mirrors the real flow WITHOUT executing the
// actual mutateCmd (that dispatch is covered by
// TestDeleteConfirmEnterFiresDeleteAndCloses above) -- a successful Delete's
// mutationDoneMsg always carries err == nil, which is all applyMutationResult
// needs to fire its unconditional reload.
func TestDeleteCursorClampsViaApplyLoadedOldPos(t *testing.T) {
	beans := fixtureBeans() // ms-1 -> ep-1 -> tk-1, tk-2
	m := fixtureModel(t, beans)
	m.expanded = map[string]bool{"ms-1": true, "ep-1": true}
	m.cursorID = "tk-2"

	m = step(t, m, runeMsg('d'))
	if m.overlay != overlayDeleteConfirm {
		t.Fatal("setup: d did not open the delete confirm")
	}

	tm, cmd := m.Update(keyMsg(tea.KeyEnter))
	nm := tm.(model)
	if nm.overlay != overlayNone {
		t.Fatal("enter did not close the delete confirm")
	}
	if cmd == nil {
		t.Fatal("enter must return the delete mutateCmd")
	}

	nm = step(t, nm, mutationDoneMsg{err: nil})
	remaining := []data.Bean{beans[0], beans[1], beans[2]} // ms-1, ep-1, tk-1 -- tk-2 deleted
	nm = step(t, nm, beansLoadedMsg{beans: remaining})

	if nm.cursorID == "tk-2" {
		t.Fatal("cursor must clamp off the deleted bean")
	}
	if nm.cursorID != "tk-1" {
		t.Fatalf("cursorID after delete-clamp = %q, want tk-1 (neighbor clamp)", nm.cursorID)
	}
}

// --- keyNodeAction dispatch ---

// TestKeyNodeActionDNoFocusedBeanIsSilentNoOp guards keyNodeAction's shared
// focusedBean()==nil guard (update.go) also covers Delete: no node to act on
// on an empty/pre-load repo, handled but silent.
func TestKeyNodeActionDNoFocusedBeanIsSilentNoOp(t *testing.T) {
	m := newModel(nil, "/tmp/bt-fixture-repo") // pre-load: idx == nil, no focused bean

	tm, cmd := m.Update(runeMsg('d'))
	nm, ok := tm.(model)
	if !ok {
		t.Fatalf("Update('d') did not return a model, got %T", tm)
	}
	if nm.overlay != overlayNone {
		t.Fatalf("overlay = %v, want overlayNone (no focused bean to delete)", nm.overlay)
	}
	if cmd != nil {
		t.Fatal("d with no focused bean must not fire a Cmd")
	}
}
