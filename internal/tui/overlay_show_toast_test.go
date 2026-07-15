package tui

// overlay_show_toast_test.go — E5 Task 1 (bean bt-6dts): State/Update tests
// for the corner Toast (lifecycle, sticky, debounce, click, E3-conflict
// carry-over). Mirrors devd overlay_show_toast_test.go's coverage shape
// (~/Obsidian/tools/DeveloperDashboard/apps/cli-go/internal/tui/
// overlay_show_toast_test.go), scoped to beans-tui's own model/toastTarget
// shape (design decision a: toastTarget{view viewID} only).
//
// showToast()'s non-sticky return is a tea.Tick-wrapped Cmd that blocks for
// real (toastDuration(kind), 3-8s) if ever invoked -- NEVER call cmd()
// against it here (same rationale as devd's own doc comment on this file);
// only its non-nilness/the seq/generation bookkeeping is asserted.

import (
	"errors"
	"testing"

	"beans-tui/internal/data"
	tea "github.com/charmbracelet/bubbletea"
)

// --- showToast: single slot, generation bookkeeping ---

// TestShowToastSingleSlotReplacesPrevious guards the ONE-slot contract
// (design decision a): a second showToast call outside the debounce window
// replaces the first toast outright and bumps the generation (seq).
func TestShowToastSingleSlotReplacesPrevious(t *testing.T) {
	m := model{width: 90, height: 24}
	m, cmd1 := m.showToast(toastInfo, "first", "", nil, false)
	if cmd1 == nil {
		t.Fatal("first (non-sticky) showToast must return an auto-dismiss Cmd")
	}
	firstSeq := m.toast.seq

	// Push setAt outside the debounce window so the second call is treated
	// as a genuinely NEW toast, not an in-place update (TestShowToast
	// DebounceWindowUpdatesInPlace below covers the <300ms case).
	m.toast.setAt = m.toast.setAt.Add(-2 * toastDebounceWindow)

	m, cmd2 := m.showToast(toastWarn, "second", "", nil, false)
	if cmd2 == nil {
		t.Fatal("a toast outside the debounce window must return a new auto-dismiss Cmd")
	}
	if m.toast.title != "second" || m.toast.kind != toastWarn {
		t.Fatalf("toast = %+v, want the second call's content to win (single slot, no stacking)", m.toast)
	}
	if m.toast.seq == firstSeq {
		t.Fatalf("seq = %d, want it incremented past %d (new generation outside the debounce window)", m.toast.seq, firstSeq)
	}
}

// TestShowToastDebounceWindowUpdatesInPlace guards the 300ms debounce
// (design decision a): a second showToast call WITHIN the window updates the
// existing slot's content in place, WITHOUT bumping the generation or
// returning a second timer Cmd (one timer, not stacked).
func TestShowToastDebounceWindowUpdatesInPlace(t *testing.T) {
	m := model{width: 90, height: 24}
	m, cmd1 := m.showToast(toastInfo, "first", "", nil, false)
	if cmd1 == nil {
		t.Fatal("setup: first showToast must return an auto-dismiss Cmd")
	}
	firstSeq := m.toast.seq

	m, cmd2 := m.showToast(toastError, "second", "ctx", nil, false)
	if cmd2 != nil {
		t.Fatal("within the debounce window, a second showToast must NOT return a new timer Cmd")
	}
	if m.toast.seq != firstSeq {
		t.Fatalf("seq changed within the debounce window: %d -> %d, want unchanged (in-place update, not a new generation)", firstSeq, m.toast.seq)
	}
	if m.toast.title != "second" || m.toast.kind != toastError || m.toast.context != "ctx" {
		t.Fatalf("toast = %+v, want content/kind/context updated in place to the second call's values", m.toast)
	}
}

// --- Sticky ---

// TestStickyToastNoAutoDismiss guards design decision a's sticky contract
// (E3-I01 PFLICHT): sticky=true never returns an auto-dismiss Cmd.
func TestStickyToastNoAutoDismiss(t *testing.T) {
	m := model{width: 90, height: 24}
	m, cmd := m.showToast(toastError, "conflict", "", nil, true)
	if cmd != nil {
		t.Fatal("sticky toast must not return an auto-dismiss Cmd")
	}
	if m.toast == nil || !m.toast.sticky {
		t.Fatalf("toast = %+v, want sticky=true", m.toast)
	}
}

// TestClearToastUnlessSticky guards the reload-clobber guard: sticky
// survives, non-sticky is cleared.
func TestClearToastUnlessSticky(t *testing.T) {
	m := model{width: 90, height: 24}
	m, _ = m.showToast(toastError, "conflict", "", nil, true)
	m = m.clearToastUnlessSticky()
	if m.toast == nil {
		t.Fatal("a sticky toast must survive clearToastUnlessSticky")
	}
	m.toast.sticky = false
	m = m.clearToastUnlessSticky()
	if m.toast != nil {
		t.Fatal("a non-sticky toast must be cleared by clearToastUnlessSticky")
	}
}

// --- Geometry / hit-test ---

// TestToastGeometryTopRight guards the top-right placement contract
// (toastGeometry, shared between renderToast and the future Task-4 click
// hit-test): the box's right edge sits flush against m.width, y is 0.
func TestToastGeometryTopRight(t *testing.T) {
	m := model{width: 100, height: 24}
	m, _ = m.showToast(toastInfo, "x", "", nil, false)
	x, y, w, _ := m.toastGeometry()
	if y != 0 {
		t.Errorf("y = %d, want 0 (top edge)", y)
	}
	if x+w != m.width {
		t.Errorf("x+w = %d, want m.width %d (flush right edge)", x+w, m.width)
	}
	if w < 32 || w > 40 {
		t.Errorf("w = %d, outside the 32-40 target span", w)
	}
}

// TestToastHitTest guards toastHit against toastGeometry's own box: a point
// inside (including the corners) hits, a point far outside misses.
func TestToastHitTest(t *testing.T) {
	m := model{width: 100, height: 24}
	m, _ = m.showToast(toastInfo, "x", "", nil, false)
	x, y, w, h := m.toastGeometry()

	if !m.toastHit(x, y) {
		t.Error("toastHit at the box origin should be true")
	}
	if !m.toastHit(x+w-1, y+h-1) {
		t.Error("toastHit at the box's bottom-right corner should be true")
	}
	if m.toastHit(0, m.height-1) {
		t.Error("toastHit far outside the box (bottom-left corner) should be false")
	}
}

// --- Click / dismiss ---

// TestDismissToastJumpsToTarget guards dismissToast's target-adopt half: a
// set target.view is applied to m.view before the toast is cleared.
func TestDismissToastJumpsToTarget(t *testing.T) {
	m := model{width: 90, height: 24, view: viewBrowseRepo}
	m, _ = m.showToast(toastInfo, "moved", "", &toastTarget{view: viewBacklog}, false)

	tm, _ := m.dismissToast()
	nm, ok := tm.(model)
	if !ok {
		t.Fatalf("dismissToast did not return a model, got %T", tm)
	}
	if nm.toast != nil {
		t.Error("dismissToast must clear the toast")
	}
	if nm.view != viewBacklog {
		t.Errorf("view = %v, want viewBacklog (target adopted)", nm.view)
	}
}

// --- toastExpiredMsg dispatch (Update()'s new case, Task 1 Step 3) ---

// TestToastExpiredMsgClearsOnlyMatchingGeneration guards Update()'s
// toastExpiredMsg case (handleToastExpired): a stale tick (an old
// generation's seq) must not clear a toast a newer showToast call has since
// replaced; the current generation's own tick does.
func TestToastExpiredMsgClearsOnlyMatchingGeneration(t *testing.T) {
	m := model{width: 90, height: 24}
	m, _ = m.showToast(toastInfo, "first", "", nil, false)
	staleSeq := m.toast.seq
	m.toast.setAt = m.toast.setAt.Add(-2 * toastDebounceWindow)
	m, _ = m.showToast(toastWarn, "second", "", nil, false)

	tm, _ := m.Update(toastExpiredMsg{seq: staleSeq})
	nm := tm.(model)
	if nm.toast == nil || nm.toast.title != "second" {
		t.Fatalf("a stale generation's tick cleared/changed the current toast: %+v", nm.toast)
	}

	tm2, _ := nm.Update(toastExpiredMsg{seq: nm.toast.seq})
	nm2 := tm2.(model)
	if nm2.toast != nil {
		t.Errorf("toast = %+v, want cleared by the CURRENT generation's tick", nm2.toast)
	}
}

// TestToastSeqStaysMonotonicAcrossReset guards the T6b-Review Prelude I01
// fix (bean bt-ggt2): showToast's seq must be drawn from a model-wide
// counter that is NEVER reset by m.toast=nil (dismissToast,
// handleToastExpired, applyRepoSwitched all assign it directly) -- otherwise
// a still-in-flight toastTimeout tick from the toast that existed BEFORE the
// reset carries a seq that COLLIDES with the FIRST toast shown AFTER it
// (both would compute seq=1 under the old "m.toast.seq+1, m.toast==nil ->
// 1" scheme), and handleToastExpired's seq-equality check would then
// dismiss the wrong (new, unrelated) toast.
func TestToastSeqStaysMonotonicAcrossReset(t *testing.T) {
	m := model{width: 90, height: 24}
	m, _ = m.showToast(toastInfo, "first", "", nil, false)
	staleSeq := m.toast.seq

	// Simulate a toast-slot reset that is NOT a debounced update -- exactly
	// what dismissToast/handleToastExpired/applyRepoSwitched all do (m.toast
	// = nil), independent of which one triggered it.
	m.toast = nil

	m, _ = m.showToast(toastWarn, "second", "", nil, false)
	if m.toast.seq == staleSeq {
		t.Fatalf("second toast's seq (%d) collides with the pre-reset toast's seq (%d) -- want a NEW, never-reused generation", m.toast.seq, staleSeq)
	}

	// The stale tick from the FIRST (pre-reset) toast must not dismiss the
	// second, unrelated toast that would occupy the same OLD seq value under
	// the buggy per-slot counter.
	tm, _ := m.Update(toastExpiredMsg{seq: staleSeq})
	nm := tm.(model)
	if nm.toast == nil || nm.toast.title != "second" {
		t.Fatalf("a stale pre-reset tick (seq=%d) cleared/changed the post-reset toast: %+v", staleSeq, nm.toast)
	}
}

// --- E3 carry-over (E3-I01 PFLICHT, bean bt-5h4d body) ---

// TestConflictToastIsStickyAndSurvivesReload is the E3-Übernahme-Kern-Test
// (Task 1 Step 6): a data.ErrConflict-classified mutation result sets a
// sticky toast (applyMutationResult) that survives a subsequent reload
// (beansLoadedMsg) untouched -- applyLoaded never assigns m.toast, so there
// is nothing to clobber it (design decision a).
func TestConflictToastIsStickyAndSurvivesReload(t *testing.T) {
	fakeBeansConflict(t)
	m := fixtureModel(t, fixtureBeans())
	m.client = &data.Client{RepoDir: t.TempDir()}
	m = focusBean(m, "tk-2")
	m = step(t, m, runeMsg('s'))
	if m.overlay != overlayValueMenu {
		t.Fatal("setup: s did not open the value menu")
	}

	tm, cmd := m.Update(keyMsg(tea.KeyEnter))
	nm, ok := tm.(model)
	if !ok {
		t.Fatalf("Update(enter) did not return a model, got %T", tm)
	}
	if cmd == nil {
		t.Fatal("setup: expected a mutation Cmd")
	}
	msg := cmd()
	mdm, ok := msg.(mutationDoneMsg)
	if !ok {
		t.Fatalf("cmd() = %T, want mutationDoneMsg", msg)
	}
	if !errors.Is(mdm.err, data.ErrConflict) {
		t.Fatalf("setup: mutationDoneMsg.err = %v, want a data.ErrConflict-classified error", mdm.err)
	}

	nm2 := step(t, nm, mdm)
	if nm2.toast == nil || !nm2.toast.sticky {
		t.Fatalf("toast = %+v, want a sticky toast after an ErrConflict mutation result", nm2.toast)
	}
	if nm2.toast.kind != toastError {
		t.Errorf("toast.kind = %v, want toastError", nm2.toast.kind)
	}

	// A reload lands (watchMsg's own follow-up, or ctrl+r) -- the sticky
	// toast must survive it untouched, same "readable until the PO acts"
	// contract as the status line's own "Conflict" text (m.err, unaffected
	// by this reload either).
	nm3 := step(t, nm2, beansLoadedMsg{beans: fixtureBeans()})
	if nm3.toast == nil {
		t.Fatal("sticky toast must survive a reload (beansLoadedMsg)")
	}
	if !nm3.toast.sticky {
		t.Error("toast.sticky flipped to false across the reload, want unchanged")
	}
}
