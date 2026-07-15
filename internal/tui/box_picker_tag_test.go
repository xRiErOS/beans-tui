package tui

// box_picker_tag_test.go — TDD coverage for the Tag-Picker (`t`, E3 Task 2,
// bean bt-8v69): usage-counted toggle-multi-select + free-text new-tag
// entry + Pending-Diff (Port beans-src blockingpicker.go's Enter=confirm/
// esc=cancel convention -- NOT box_filter_facets.go's enter=close-without-
// mutating, since this overlay actually mutates). Reuses fixtureBeans/
// fixtureModel/step/keyMsg/runeMsg/focusBean from update_test.go/
// box_menu_value_test.go (same package).

import (
	"strings"
	"testing"

	"beans-tui/internal/data"
	tea "github.com/charmbracelet/bubbletea"
)

// fixtureBeansTagged layers tags onto fixtureBeans' fixed hierarchy so the
// tag-count/pending-diff tests have real cross-bean usage counts to work
// with: urgent appears on tk-1 AND tk-2 (count 2), backend and zeta each
// appear once (count 1, tied -- exercises collectTagCounts' alpha
// tie-break: backend < zeta).
func fixtureBeansTagged() []data.Bean {
	beans := fixtureBeans()
	for i := range beans {
		switch beans[i].ID {
		case "tk-1":
			beans[i].Tags = []string{"urgent", "backend"}
		case "tk-2":
			beans[i].Tags = []string{"urgent"}
		case "ep-1":
			beans[i].Tags = []string{"zeta"}
		}
	}
	return beans
}

// tagPickerCursorTo moves m.menu.cursor onto the row whose tag == tag,
// failing the test if no such row exists -- shared setup step for every
// toggle-driving test below.
func tagPickerCursorTo(t *testing.T, m model, tag string) model {
	t.Helper()
	for i, it := range m.tagItems {
		if it.tag == tag {
			m.menu.cursor = i
			return m
		}
	}
	t.Fatalf("tag %q not found in tagItems (%+v)", tag, m.tagItems)
	return m
}

// --- collectTagCounts ---

func TestCollectTagCountsSortedByCountDescThenAlpha(t *testing.T) {
	idx := data.NewIndex(fixtureBeansTagged())
	got := collectTagCounts(idx)

	want := []tagCount{
		{tag: "urgent", count: 2},
		{tag: "backend", count: 1},
		{tag: "zeta", count: 1},
	}
	if len(got) != len(want) {
		t.Fatalf("collectTagCounts() = %+v, want len %d", got, len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("collectTagCounts()[%d] = %+v, want %+v (full: %+v)", i, got[i], want[i], got)
		}
	}
}

// --- openTagPicker ---

func TestOpenTagPickerSeedsPendingFromFocusedBean(t *testing.T) {
	m := fixtureModel(t, fixtureBeansTagged())
	m = focusBean(m, "tk-2") // tags: urgent

	m = step(t, m, runeMsg('t'))

	if m.overlay != overlayTagPicker {
		t.Fatalf("overlay = %v, want overlayTagPicker", m.overlay)
	}
	if m.mutTarget != "tk-2" {
		t.Fatalf("mutTarget = %q, want tk-2", m.mutTarget)
	}
	if len(m.tagOriginal) != 1 || !m.tagOriginal["urgent"] {
		t.Fatalf("tagOriginal = %v, want {urgent:true}", m.tagOriginal)
	}
	if len(m.tagPending) != 1 || !m.tagPending["urgent"] {
		t.Fatalf("tagPending = %v, want {urgent:true}", m.tagPending)
	}
	if len(m.tagItems) != 3 {
		t.Fatalf("tagItems len = %d, want 3 (urgent/backend/zeta)", len(m.tagItems))
	}

	// tagOriginal/tagPending must be two INDEPENDENT maps, not one aliasing
	// the other (design decision, plan »Task 2«: "zwei UNABHÄNGIGE Maps").
	m.tagPending["scratch-independence-probe"] = true
	if m.tagOriginal["scratch-independence-probe"] {
		t.Fatal("tagOriginal aliases tagPending's backing map -- must be independent")
	}
}

// --- keyTagPicker: toggle ---

func TestTagPickerToggleFlipsPendingOnly(t *testing.T) {
	m := fixtureModel(t, fixtureBeansTagged())
	m = focusBean(m, "tk-2") // tags: urgent
	m = step(t, m, runeMsg('t'))

	m = tagPickerCursorTo(t, m, "urgent")
	m = step(t, m, runeMsg(' '))
	if m.tagPending["urgent"] {
		t.Fatal("space did not toggle urgent OFF in tagPending")
	}
	if !m.tagOriginal["urgent"] {
		t.Fatal("tagOriginal must stay unchanged by a pending toggle")
	}

	m = tagPickerCursorTo(t, m, "backend")
	m = step(t, m, runeMsg('x'))
	if !m.tagPending["backend"] {
		t.Fatal("x did not toggle backend ON in tagPending")
	}
	if m.tagOriginal["backend"] {
		t.Fatal("tagOriginal must stay unchanged (backend was never original)")
	}
}

// --- keyTagPicker: enter diffs via ONE SetTags call ---

// TestTagPickerEnterDiffsFireOneSetTagsCall guards the E3 Task 2 design
// decision (ERRATUM vs. the plan's own earlier sketch, settled in the
// "Design-Nachtrag" of epic-E3-plan.md »Task 2«): the diff (one add, one
// remove) fires as ONE combined data.SetTags call, not two separate
// AddTag/RemoveTag mutateCmds batched via tea.Batch -- N sequential
// mutations against ONE etag would be a conflict cascade. Verified two
// ways: cmd() resolves DIRECTLY to a mutationDoneMsg (a tea.Batch would
// instead resolve to a tea.BatchMsg), and the underlying error text (real
// data.Client pointed at a nonexistent repo dir, no beans binary required)
// contains "beans update", proving SetTags actually dispatched.
func TestTagPickerEnterDiffsFireOneSetTagsCall(t *testing.T) {
	m := fixtureModel(t, fixtureBeansTagged())
	m.client = &data.Client{RepoDir: "/nonexistent-bt-e3-t2-scratch-dir"}
	m = focusBean(m, "tk-2") // tags: urgent
	m = step(t, m, runeMsg('t'))

	m = tagPickerCursorTo(t, m, "urgent") // originally present -> toggle off = remove
	m = step(t, m, runeMsg(' '))
	m = tagPickerCursorTo(t, m, "backend") // not originally present -> toggle on = add
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
		t.Fatalf("cmd() = %T, want mutationDoneMsg (ONE combined SetTags call, not tea.Batch)", msg)
	}
	if mdm.err == nil || !strings.Contains(mdm.err.Error(), "beans update") {
		t.Fatalf("mutationDoneMsg.err = %v, want an error containing %q (proves SetTags dispatched)", mdm.err, "beans update")
	}
}

func TestTagPickerEnterNoChangesNoMutation(t *testing.T) {
	m := fixtureModel(t, fixtureBeansTagged())
	m = focusBean(m, "tk-2")
	m = step(t, m, runeMsg('t'))

	tm, cmd := m.Update(keyMsg(tea.KeyEnter))
	nm := tm.(model)
	if nm.overlay != overlayNone {
		t.Fatal("enter must close the picker even with no changes")
	}
	if cmd != nil {
		t.Fatal("enter with no pending changes must not fire a mutation Cmd")
	}
}

// --- keyTagPicker: esc discards ---

func TestTagPickerEscDiscardsPending(t *testing.T) {
	m := fixtureModel(t, fixtureBeansTagged())
	m = focusBean(m, "tk-2")
	m = step(t, m, runeMsg('t'))
	m = tagPickerCursorTo(t, m, "backend")
	m = step(t, m, runeMsg(' ')) // toggle backend on in pending

	tm, cmd := m.Update(keyMsg(tea.KeyEsc))
	nm := tm.(model)
	if nm.overlay != overlayNone {
		t.Fatal("esc did not close the tag picker")
	}
	if cmd != nil {
		t.Fatal("esc must not fire a mutation Cmd")
	}
}

// --- keyTagPicker: free-text new-tag sub-mode ---

func TestTagPickerNewTagValidatesRegex(t *testing.T) {
	m := fixtureModel(t, fixtureBeansTagged())
	m = focusBean(m, "tk-2")
	m = step(t, m, runeMsg('t'))
	before := len(m.tagItems)

	m = step(t, m, runeMsg('n'))
	if !m.tagInputActive {
		t.Fatal("n did not open the new-tag input")
	}
	for _, r := range "Über!" {
		m = step(t, m, runeMsg(r))
	}
	m = step(t, m, keyMsg(tea.KeyEnter))

	if m.tagInputErr == "" {
		t.Fatal("invalid new tag name must set tagInputErr")
	}
	if !m.tagInputActive {
		t.Fatal("invalid submit must keep the input open for a retry, not close it")
	}
	if len(m.tagItems) != before {
		t.Fatalf("tagItems len changed on invalid submit: %d -> %d, want no change", before, len(m.tagItems))
	}
}

func TestTagPickerNewTagAddsPendingItem(t *testing.T) {
	m := fixtureModel(t, fixtureBeansTagged())
	m = focusBean(m, "tk-2")
	m = step(t, m, runeMsg('t'))
	before := len(m.tagItems)

	m = step(t, m, runeMsg('n'))
	for _, r := range "greenfield" {
		m = step(t, m, runeMsg(r))
	}
	m = step(t, m, keyMsg(tea.KeyEnter))

	if m.tagInputErr != "" {
		t.Fatalf("valid new tag name must not set tagInputErr, got %q", m.tagInputErr)
	}
	if m.tagInputActive {
		t.Fatal("valid submit must close the new-tag input")
	}
	if len(m.tagItems) != before+1 {
		t.Fatalf("tagItems len = %d, want %d (one new row)", len(m.tagItems), before+1)
	}
	if !m.tagPending["greenfield"] {
		t.Fatal("a freshly created tag must be pending=true immediately")
	}
	found := false
	for _, it := range m.tagItems {
		if it.tag == "greenfield" {
			found = true
		}
	}
	if !found {
		t.Fatal("greenfield missing from tagItems after a valid submit")
	}
}

// --- vanished-target guard on enter (design decision d parity, T1) ---

func TestTagPickerEnterTargetVanishedClosesGracefully(t *testing.T) {
	beans := fixtureBeansTagged()
	m := fixtureModel(t, beans)
	m = focusBean(m, "tk-2")
	m = step(t, m, runeMsg('t'))
	m = tagPickerCursorTo(t, m, "backend")
	m = step(t, m, runeMsg(' ')) // ensure a pending change exists

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

// --- inherited PFLICHT findings from the E3-T2 review (bt-8v69 body),
// closed out here as part of E3 Task 3 (bean bt-p1uz) per its own bean
// body's "Übernommene Findings" section ---

// TestTagPickerEscFromInputKeepsPickerOpenPendingIntact is PFLICHT I1a: esc
// while the free-text new-tag input (m.tagInputActive) is open must close
// ONLY that input sub-mode (keyTagInput's own esc case) -- the OUTER
// Tag-Picker overlay stays open (overlayTagPicker, not overlayNone) and
// whatever was already toggled into tagPending before "n" was pressed must
// survive untouched.
func TestTagPickerEscFromInputKeepsPickerOpenPendingIntact(t *testing.T) {
	m := fixtureModel(t, fixtureBeansTagged())
	m = focusBean(m, "tk-2") // tags: urgent
	m = step(t, m, runeMsg('t'))
	m = tagPickerCursorTo(t, m, "backend")
	m = step(t, m, runeMsg(' ')) // toggle backend on in pending, BEFORE opening the input

	m = step(t, m, runeMsg('n'))
	if !m.tagInputActive {
		t.Fatal("setup: n did not open the new-tag input")
	}

	tm, cmd := m.Update(keyMsg(tea.KeyEsc))
	nm := tm.(model)
	if nm.tagInputActive {
		t.Fatal("esc must close the new-tag input sub-mode")
	}
	if nm.overlay != overlayTagPicker {
		t.Fatalf("overlay = %v, want overlayTagPicker (esc-from-input must NOT close the outer picker)", nm.overlay)
	}
	if cmd != nil {
		t.Fatal("esc-from-input must not fire a mutation Cmd")
	}
	if !nm.tagPending["urgent"] || !nm.tagPending["backend"] {
		t.Fatalf("tagPending = %v, want {urgent:true, backend:true} unchanged by esc-from-input", nm.tagPending)
	}
}

// TestTagPickerToggleOffThenOnYieldsEmptyDiffNoMutation is PFLICHT I1b: a
// tag toggled off and then back on nets out to the SAME state as
// tagOriginal -- applyTagPickerDiff's add/remove diff must come out empty,
// so enter closes the picker WITHOUT firing a Cmd, exactly like the
// no-changes-at-all case (TestTagPickerEnterNoChangesNoMutation).
func TestTagPickerToggleOffThenOnYieldsEmptyDiffNoMutation(t *testing.T) {
	m := fixtureModel(t, fixtureBeansTagged())
	m = focusBean(m, "tk-2") // tags: urgent
	m = step(t, m, runeMsg('t'))

	m = tagPickerCursorTo(t, m, "urgent")
	m = step(t, m, runeMsg(' ')) // off
	m = step(t, m, runeMsg(' ')) // back on -- net: unchanged vs. tagOriginal

	if !m.tagPending["urgent"] {
		t.Fatal("setup: urgent should be back in tagPending after toggling off then on")
	}

	tm, cmd := m.Update(keyMsg(tea.KeyEnter))
	nm := tm.(model)
	if nm.overlay != overlayNone {
		t.Fatal("enter must close the picker even when the toggle nets out to no change")
	}
	if cmd != nil {
		t.Fatal("a toggle-off-then-on net-zero diff must not fire a mutation Cmd")
	}
}

// --- capture-order guard (cheap follow-up from the T1 review, bt-8v69 body) ---

// TestOverlayCaptureSwallowsQuitKeysWhileTagPickerOpen pins handleKey's
// capture order (design decision a2/the overlay precedent m.filterOpen/
// m.searchActive already established): q/ctrl+c must never leak through to
// a quit-request while ANY node-action overlay is open. Exercised against
// the tag picker since that is what this task builds, but the guard itself
// (m.overlay != overlayNone routes to keyOverlay BEFORE the global
// ctrl+c/q switch, handleKey) is overlay-agnostic.
func TestOverlayCaptureSwallowsQuitKeysWhileTagPickerOpen(t *testing.T) {
	m := fixtureModel(t, fixtureBeansTagged())
	m = focusBean(m, "tk-2")
	m = step(t, m, runeMsg('t'))
	if m.overlay != overlayTagPicker {
		t.Fatal("setup: t did not open the tag picker")
	}

	tm, cmd := m.Update(keyMsg(tea.KeyCtrlC))
	nm := tm.(model)
	if cmd != nil {
		t.Fatal("ctrl+c while an overlay is open must not fire tea.Quit (capture-order guard)")
	}
	if nm.overlay != overlayTagPicker {
		t.Fatal("ctrl+c must not close the open overlay either")
	}

	tm, cmd = nm.Update(runeMsg('q'))
	nm = tm.(model)
	if cmd != nil {
		t.Fatal("q while an overlay is open must not fire tea.Quit (capture-order guard)")
	}
	if nm.overlay != overlayTagPicker {
		t.Fatal("q must not close the open overlay either")
	}
}
