package tui

// box_menu_value_test.go — TDD coverage for the combined Status/Type/
// Priority value menu (`s`, E3 Task 1, bean bt-dlgk, design decision a3).
// Reuses fixtureBeans/fixtureModel/step/keyMsg/runeMsg from update_test.go
// (same package).

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"beans-tui/internal/data"
	tea "github.com/charmbracelet/bubbletea"
)

// focusBean expands every ancestor fixtureBeans' tk-1/tk-2 need (ms-1 ->
// ep-1 -> tk-*, both start collapsed) so id actually appears in
// visibleNodes() and cursorPos can find it -- mirrors the same
// ancestor-expand dance update_test.go's detail-focus tests already use
// (e.g. TestDetailFocusRightEntersFieldLevelOnlyForBeziehungenSection).
// Setting m.cursorID alone is not enough: cursorPos falls back to index 0
// (ms-1) for a cursorID absent from the current visibleNodes() slice.
func focusBean(m model, id string) model {
	m.expanded = map[string]bool{"ms-1": true, "ep-1": true}
	m.cursorID = id
	return m
}

// --- openValueMenu / buildValueMenuItems ---

// TestOpenValueMenuBuildsGroupedItemsCursorOnCurrentStatus guards `s` on a
// focused bean: 15 rows (5 status + 5 type + 5 priority), mutTarget captures
// the bean's ID, and the cursor seeds onto the bean's CURRENT status within
// the status group (port beans-src statuspicker.go's selectedIndex seed).
func TestOpenValueMenuBuildsGroupedItemsCursorOnCurrentStatus(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m = focusBean(m, "tk-2") // status=todo, type=task, priority=normal (fixtureBeans, ms-1 -> ep-1 -> tk-2)

	m = step(t, m, runeMsg('s'))

	if len(m.menuItems) != 15 {
		t.Fatalf("menuItems len = %d, want 15 (5 status + 5 type + 5 priority)", len(m.menuItems))
	}
	if m.mutTarget != "tk-2" {
		t.Fatalf("mutTarget = %q, want tk-2", m.mutTarget)
	}
	if m.overlay != overlayValueMenu {
		t.Fatalf("overlay = %v, want overlayValueMenu", m.overlay)
	}
	want := valueMenuCursorFor(m.menuItems, "status", "todo")
	if m.menu.cursor != want {
		t.Fatalf("menu.cursor = %d, want %d (todo within the status group)", m.menu.cursor, want)
	}
	if got := m.menuItems[m.menu.cursor]; got.group != "status" || got.value != "todo" {
		t.Fatalf("cursored item = %+v, want {status todo}", got)
	}
}

// --- keyValueMenu: enter applies + closes ---

// TestValueMenuEnterAppliesCursoredValueAndCloses guards the immediate-apply
// Enter semantics (design decision a3): moving the cursor onto "in-progress"
// (status group) and pressing enter closes the overlay and dispatches
// SetStatus via a mutateCmd. The dispatch itself is verified through the
// REAL data.Client's error message (pointed at a non-existent repo dir, no
// beans binary required) rather than a mock -- the error text starting with
// "beans update" proves SetStatus (not SetType/SetPriority) actually ran.
func TestValueMenuEnterAppliesCursoredValueAndCloses(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m.client = &data.Client{RepoDir: "/nonexistent-bt-e3-t1-scratch-dir"}
	m = focusBean(m, "tk-2") // status=todo, type=task, priority=normal (fixtureBeans, ms-1 -> ep-1 -> tk-2)

	m = step(t, m, runeMsg('s'))
	// "in-progress" sits BEFORE "todo" in tier order (data.StatusValues()),
	// so it cannot be reached by driving KeyDown from the seeded cursor --
	// set the cursor directly via the same lookup openValueMenu itself uses.
	m.menu.cursor = valueMenuCursorFor(m.menuItems, "status", "in-progress")

	tm, cmd := m.Update(keyMsg(tea.KeyEnter))
	nm, ok := tm.(model)
	if !ok {
		t.Fatalf("Update(enter) did not return a model, got %T", tm)
	}
	if nm.overlay != overlayNone {
		t.Fatalf("overlay after enter = %v, want overlayNone", nm.overlay)
	}
	if cmd == nil {
		t.Fatal("enter must return a Cmd (mutateCmd)")
	}

	msg := cmd()
	mdm, ok := msg.(mutationDoneMsg)
	if !ok {
		t.Fatalf("cmd() = %T, want mutationDoneMsg", msg)
	}
	if mdm.err == nil || !strings.Contains(mdm.err.Error(), "beans update") {
		t.Fatalf("mutationDoneMsg.err = %v, want an error containing %q (proves SetStatus dispatched)", mdm.err, "beans update")
	}
}

// TestValueMenuEnterOnTypeGroupDispatchesSetType mirrors the status-group
// test for the type group, guarding that the group->Set* dispatch itself is
// keyed correctly (not just "some mutation fired").
func TestValueMenuEnterOnTypeGroupDispatchesSetType(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m.client = &data.Client{RepoDir: "/nonexistent-bt-e3-t1-scratch-dir"}
	m = focusBean(m, "tk-2") // status=todo, type=task, priority=normal (fixtureBeans, ms-1 -> ep-1 -> tk-2)

	m = step(t, m, runeMsg('s'))
	for m.menuItems[m.menu.cursor].group != "type" {
		m = step(t, m, keyMsg(tea.KeyDown))
	}

	_, cmd := m.Update(keyMsg(tea.KeyEnter))
	if cmd == nil {
		t.Fatal("enter must return a Cmd")
	}
	mdm := cmd().(mutationDoneMsg)
	if mdm.err == nil || !strings.Contains(mdm.err.Error(), "beans update") {
		t.Fatalf("mutationDoneMsg.err = %v, want an error containing %q", mdm.err, "beans update")
	}
}

// TestValueMenuEnterOnPriorityGroupDispatchesSetPriority mirrors the above
// for the priority group.
func TestValueMenuEnterOnPriorityGroupDispatchesSetPriority(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m.client = &data.Client{RepoDir: "/nonexistent-bt-e3-t1-scratch-dir"}
	m = focusBean(m, "tk-2") // status=todo, type=task, priority=normal (fixtureBeans, ms-1 -> ep-1 -> tk-2)

	m = step(t, m, runeMsg('s'))
	for m.menuItems[m.menu.cursor].group != "priority" {
		m = step(t, m, keyMsg(tea.KeyDown))
	}

	_, cmd := m.Update(keyMsg(tea.KeyEnter))
	if cmd == nil {
		t.Fatal("enter must return a Cmd")
	}
	mdm := cmd().(mutationDoneMsg)
	if mdm.err == nil || !strings.Contains(mdm.err.Error(), "beans update") {
		t.Fatalf("mutationDoneMsg.err = %v, want an error containing %q", mdm.err, "beans update")
	}
}

// --- keyValueMenu: esc/s close without mutation ---

// TestValueMenuEscClosesWithoutMutation guards esc: overlay closes, no Cmd.
func TestValueMenuEscClosesWithoutMutation(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m = focusBean(m, "tk-2") // status=todo, type=task, priority=normal (fixtureBeans, ms-1 -> ep-1 -> tk-2)
	m = step(t, m, runeMsg('s'))

	tm, cmd := m.Update(keyMsg(tea.KeyEsc))
	nm := tm.(model)
	if nm.overlay != overlayNone {
		t.Fatal("esc did not close the value menu")
	}
	if cmd != nil {
		t.Fatal("esc must not fire a mutation Cmd")
	}
}

// TestValueMenuSReclosesWithoutMutation guards `s` itself as the second
// close key (design decision a3: "esc/s schließt ohne Mutation") -- mirrors
// keyFilterMenu's own multi-key-closes-without-clearing precedent.
func TestValueMenuSReclosesWithoutMutation(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m = focusBean(m, "tk-2") // status=todo, type=task, priority=normal (fixtureBeans, ms-1 -> ep-1 -> tk-2)
	m = step(t, m, runeMsg('s'))
	if m.overlay != overlayValueMenu {
		t.Fatal("setup: s did not open the value menu")
	}

	tm, cmd := m.Update(runeMsg('s'))
	nm := tm.(model)
	if nm.overlay != overlayNone {
		t.Fatal("s did not close the already-open value menu")
	}
	if cmd != nil {
		t.Fatal("s-close must not fire a mutation Cmd")
	}
}

// --- valueMenuBox rendering ---

// TestValueMenuCurrentValueMarked guards the "(current)" marker (port
// beans-src statuspicker.go isCurrent) on the focused bean's actual
// status/type/priority values.
func TestValueMenuCurrentValueMarked(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m = focusBean(m, "tk-2") // status=todo, type=task, priority=normal (fixtureBeans, ms-1 -> ep-1 -> tk-2)
	m = step(t, m, runeMsg('s'))

	out := m.valueMenuBox()
	if !strings.Contains(out, "(current)") {
		t.Fatalf("valueMenuBox() missing a (current) marker:\n%s", out)
	}
}

// --- keyNodeAction: focused-bean guard ---

// TestKeyNodeActionRequiresFocusedBeanExceptCreate guards that every T1
// node-action key except Create (s/t/a/B/d/e) is a handled-but-silent no-op
// with no focused bean (an empty repo / the orphan-root cursor), while
// Create (c) is handled regardless -- it works on an empty repo (T4).
func TestKeyNodeActionRequiresFocusedBeanExceptCreate(t *testing.T) {
	m := fixtureModel(t, nil) // no beans -> focusedBean() == nil
	if m.focusedBean() != nil {
		t.Fatal("setup: expected focusedBean() == nil with zero beans loaded")
	}

	for _, k := range []tea.KeyMsg{runeMsg('s'), runeMsg('t'), runeMsg('a'), runeMsg('B'), runeMsg('d'), runeMsg('e')} {
		handled, nm, cmd := m.keyNodeAction(k)
		if !handled {
			t.Fatalf("key %v: handled = false, want true (handled-but-silent no-op)", k)
		}
		if cmd != nil {
			t.Fatalf("key %v: cmd != nil, want nil (no-op, no focused bean)", k)
		}
		mm, ok := nm.(model)
		if !ok {
			t.Fatalf("key %v: keyNodeAction did not return a model, got %T", k, nm)
		}
		if mm.overlay != overlayNone {
			t.Fatalf("key %v: overlay = %v, want overlayNone (no focused bean -- must not open)", k, mm.overlay)
		}
	}

	handled, _, _ := m.keyNodeAction(runeMsg('c'))
	if !handled {
		t.Fatal("c (Create) must be handled even with no focused bean")
	}
}

// TestKeyNodeActionStatusOpensMenuWithFocusedBean is the positive-path
// counterpart: `s` with a real focused bean opens the value menu (already
// covered end-to-end above, but this pins keyNodeAction's own dispatch
// directly, independent of handleKey's capture order).
func TestKeyNodeActionStatusOpensMenuWithFocusedBean(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m = focusBean(m, "tk-2") // status=todo, type=task, priority=normal (fixtureBeans, ms-1 -> ep-1 -> tk-2)

	handled, nm, cmd := m.keyNodeAction(runeMsg('s'))
	if !handled {
		t.Fatal("s must be handled")
	}
	if cmd != nil {
		t.Fatal("s (open) must not itself fire a Cmd")
	}
	mm := nm.(model)
	if mm.overlay != overlayValueMenu {
		t.Fatalf("overlay = %v, want overlayValueMenu", mm.overlay)
	}
}

// --- applyMutationResult ---

// TestApplyMutationResultConflictSetsStatusLineAndReloads guards the
// ErrConflict branch: m.err carries a "Conflict" note (Toast is E5;
// the status line is the interim channel) and a reload Cmd fires regardless.
func TestApplyMutationResultConflictSetsStatusLineAndReloads(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	err := fmt.Errorf("%w: bean x: boom", data.ErrConflict)

	tm, cmd := m.applyMutationResult(err)
	nm := tm.(model)
	if !strings.Contains(nm.err, "Conflict") {
		t.Fatalf("m.err = %q, want it to contain %q", nm.err, "Conflict")
	}
	if cmd == nil {
		t.Fatal("cmd == nil, want a reload Cmd (loadCmd) even on conflict")
	}
}

// TestApplyMutationResultSuccessClearsErrAndReloads guards the success path:
// a stale m.err from a previous failure is cleared and a reload still fires
// (design decision d: unconditional reload, success or failure alike).
func TestApplyMutationResultSuccessClearsErrAndReloads(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m.err = "stale error from a previous mutation"

	tm, cmd := m.applyMutationResult(nil)
	nm := tm.(model)
	if nm.err != "" {
		t.Fatalf("m.err = %q, want empty (cleared on success)", nm.err)
	}
	if cmd == nil {
		t.Fatal("cmd == nil, want a reload Cmd (loadCmd) on success")
	}
}

// TestApplyMutationResultNonConflictErrorSurfacesRawMessage guards a
// non-ErrConflict failure (e.g. VALIDATION_ERROR): the raw error text
// surfaces verbatim, not the "Conflict" wording.
func TestApplyMutationResultNonConflictErrorSurfacesRawMessage(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	tm, cmd := m.applyMutationResult(errors.New("beans update: VALIDATION_ERROR: bad status"))
	nm := tm.(model)
	if strings.Contains(nm.err, "Conflict") {
		t.Fatalf("m.err = %q, must not use the Conflict wording for a non-ErrConflict failure", nm.err)
	}
	if !strings.Contains(nm.err, "VALIDATION_ERROR") {
		t.Fatalf("m.err = %q, want it to surface the raw error text", nm.err)
	}
	if cmd == nil {
		t.Fatal("cmd == nil, want a reload Cmd")
	}
}

// --- beanETag ---

// TestBeanETagReadsLiveIndex guards design decision d: beanETag always reads
// the CURRENT index, so a reload between two calls returns the fresh value,
// never a captured/stale copy.
func TestBeanETagReadsLiveIndex(t *testing.T) {
	beans := fixtureBeans()
	m := fixtureModel(t, beans)

	etag1, ok := m.beanETag("tk-1")
	if !ok {
		t.Fatal("beanETag(tk-1) ok = false, want true")
	}
	if etag1 != m.idx.ByID["tk-1"].ETag {
		t.Fatalf("beanETag(tk-1) = %q, want %q (live index value)", etag1, m.idx.ByID["tk-1"].ETag)
	}

	changed := append([]data.Bean(nil), beans...)
	for i := range changed {
		if changed[i].ID == "tk-1" {
			changed[i].ETag = "rotated-etag-after-reload"
		}
	}
	m = step(t, m, beansLoadedMsg{beans: changed})

	etag2, ok := m.beanETag("tk-1")
	if !ok || etag2 != "rotated-etag-after-reload" {
		t.Fatalf("beanETag(tk-1) after reload = %q, ok=%v, want %q/true", etag2, ok, "rotated-etag-after-reload")
	}
}

// TestBeanETagVanishedReturnsNotOK guards the vanished-bean path: an ID no
// longer in the index reports ok=false, never a stale/zero ETag treated as
// valid.
func TestBeanETagVanishedReturnsNotOK(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	if _, ok := m.beanETag("does-not-exist"); ok {
		t.Fatal("beanETag for an unknown ID: ok = true, want false")
	}

	m.idx = nil
	if _, ok := m.beanETag("tk-1"); ok {
		t.Fatal("beanETag with a nil index: ok = true, want false")
	}
}

// --- vanished-target guard on enter ---

// TestValueMenuTargetVanishedClosesGracefully guards the vanished-mutTarget
// path (design decision d): the focused bean disappears from the index
// (external delete + reload) between open and enter -- enter must close the
// overlay and set a status-line note, WITHOUT firing a doomed mutation Cmd.
func TestValueMenuTargetVanishedClosesGracefully(t *testing.T) {
	beans := fixtureBeans()
	m := fixtureModel(t, beans)
	m = focusBean(m, "tk-2") // status=todo, type=task, priority=normal (fixtureBeans, ms-1 -> ep-1 -> tk-2)
	m = step(t, m, runeMsg('s'))
	if m.overlay != overlayValueMenu {
		t.Fatal("setup: s did not open the value menu")
	}

	// tk-2 vanishes externally; a reload lands while the menu is still open
	// (m.overlay survives a reload untouched, applyLoaded never touches it).
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
