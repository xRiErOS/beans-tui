package tui

// box_confirm_create_test.go — TDD coverage for the Create-Form's
// Confirm-Gate (E3 Task 4, bean bt-y4ly, design decision e): submitForm's
// park-and-open, enter-fires/n-esc-returns-filled Confirm-Gate semantics
// (Draft-Erhalt, Port DD2-190), and createDoneMsg's cursor-jump +
// applyMutationResult fallback (update.go).

import (
	"fmt"
	"strings"
	"testing"

	"beans-tui/internal/data"
	tea "github.com/charmbracelet/bubbletea"
)

// driveModel is the model-level analogue of driveForm
// (form_create_bean_test.go): step() (update_test.go) discards every
// returned tea.Cmd, so it can never drive an open huh Form to
// huh.StateCompleted on its own -- keyForm/updateForm's returned Cmd is the
// SAME unresolved field-advance chain (see driveForm's doc comment for the
// full rationale, including the bounded driveFormBudget against huh's
// self-perpetuating textinput-blink timer).
func driveModel(t *testing.T, m model, msg tea.Msg) model {
	t.Helper()
	nm, _ := driveModelStep(t, m, msg, driveFormBudget)
	return nm
}

func driveModelStep(t *testing.T, m model, msg tea.Msg, budget int) (model, int) {
	t.Helper()
	if budget <= 0 {
		return m, 0
	}
	budget--
	tm, cmd := m.Update(msg)
	nm, ok := tm.(model)
	if !ok {
		t.Fatalf("Update(%T) did not return a model, got %T", msg, tm)
	}
	return driveModelCmd(t, nm, cmd, budget)
}

func driveModelCmd(t *testing.T, m model, cmd tea.Cmd, budget int) (model, int) {
	t.Helper()
	if cmd == nil || budget <= 0 {
		return m, budget
	}
	budget--
	msg := cmd()
	if msg == nil {
		return m, budget
	}
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, c := range batch {
			if budget <= 0 {
				break
			}
			m, budget = driveModelCmd(t, m, c, budget)
		}
		return m, budget
	}
	return driveModelStep(t, m, msg, budget)
}

// typeIntoModel sends s as a sequence of single-rune KeyMsgs -- mirrors a PO
// actually typing text into the focused field, consistent with this suite's
// own single-rune runeMsg convention (update_test.go).
func typeIntoModel(t *testing.T, m model, s string) model {
	t.Helper()
	for _, r := range s {
		m = driveModel(t, m, runeMsg(r))
	}
	return m
}

// advanceFieldsModel drives m forward n fields (n calls to enterMsg) -- the
// model-level counterpart of advanceFields (form_create_bean_test.go).
func advanceFieldsModel(t *testing.T, m model, n int) model {
	t.Helper()
	for i := 0; i < n; i++ {
		m = driveModel(t, m, enterMsg())
	}
	return m
}

// openFilledCreateConfirm drives a fresh Create-Form (`c`, optionally
// focused on parentID first for the Parent-Prefill) all the way through its
// 7 keyed fields (title/type/priority/status/parent/tags/body) to a filled
// Confirm-Gate -- shared setup for the Confirm-Gate tests below.
func openFilledCreateConfirm(t *testing.T, m model, parentID, title string) model {
	t.Helper()
	if parentID != "" {
		m = focusBean(m, parentID)
	}
	m = step(t, m, runeMsg('c'))
	if m.form == nil || m.formKind != "create" {
		t.Fatalf("setup: c did not open the create form (form=%v formKind=%q)", m.form, m.formKind)
	}
	m = typeIntoModel(t, m, title)
	m = advanceFieldsModel(t, m, 7) // title, type, priority, status, parent, tags, body -> submit
	if m.overlay != overlayCreateConfirm {
		t.Fatalf("setup: overlay = %v, want overlayCreateConfirm (form=%v)", m.overlay, m.form)
	}
	return m
}

// TestSubmitFormParksCmdAndOpensConfirm guards submitForm's own park-and-open
// contract (box_confirm_create.go): completing the create form nulls it,
// opens overlayCreateConfirm, parks a non-nil createCmd, and captures the
// draft (DD2-190) BEFORE the confirm.
func TestSubmitFormParksCmdAndOpensConfirm(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m = openFilledCreateConfirm(t, m, "ep-1", "New Task")

	if m.form != nil {
		t.Fatalf("form still open after the last field, want nil (submitForm must null it)")
	}
	if m.pendingCreate == nil {
		t.Fatal("pendingCreate is nil, want the parked createCmd")
	}
	if !strings.Contains(m.createLabel, "New Task") {
		t.Fatalf("createLabel = %q, want it to mention the title", m.createLabel)
	}
	if m.createDraft == nil || m.createDraft.title != "New Task" {
		t.Fatalf("createDraft = %+v, want title %q captured (DD2-190)", m.createDraft, "New Task")
	}
	if m.createDraft.parent != "ep-1" {
		t.Fatalf("createDraft.parent = %q, want ep-1 (Parent-Prefill survived to submit)", m.createDraft.parent)
	}
}

// TestCreateConfirmEnterFiresPendingCmd guards the enter branch of
// keyCreateConfirm: closes the overlay, clears the parked state, and returns
// the createCmd that actually dispatches `beans create` (verified via a real
// data.Client pointed at a nonexistent repo dir -- same no-binary-required
// pattern box_picker_parent_test.go/box_picker_tag_test.go already use).
func TestCreateConfirmEnterFiresPendingCmd(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m.client = &data.Client{RepoDir: "/nonexistent-bt-e3-t4-scratch-dir"}
	m = openFilledCreateConfirm(t, m, "ep-1", "New Task")

	tm, cmd := m.Update(keyMsg(tea.KeyEnter))
	nm, ok := tm.(model)
	if !ok {
		t.Fatalf("Update(enter) did not return a model, got %T", tm)
	}
	if nm.overlay != overlayNone {
		t.Fatalf("overlay after enter = %v, want overlayNone", nm.overlay)
	}
	if nm.pendingCreate != nil {
		t.Fatal("pendingCreate not cleared after enter")
	}
	if nm.createDraft != nil {
		t.Fatal("createDraft must be consumed (nil) after a real create fires")
	}
	if cmd == nil {
		t.Fatal("enter must fire the parked createCmd")
	}
	msg := cmd()
	cdm, ok := msg.(createDoneMsg)
	if !ok {
		t.Fatalf("cmd() = %T, want createDoneMsg", msg)
	}
	if cdm.err == nil || !strings.Contains(cdm.err.Error(), "beans create") {
		t.Fatalf("createDoneMsg.err = %v, want an error containing %q (proves Create dispatched)", cdm.err, "beans create")
	}
}

// TestCreateConfirmEscReturnsToFilledForm guards the n/esc branch
// (Draft-Erhalt, Port DD2-190): the Confirm-Gate closes WITHOUT firing a
// Cmd, and the Create-Form reopens carrying the original draft -- proven by
// advancing straight through the reopened form (no retyping) and checking
// the resulting createLabel still names the original title.
func TestCreateConfirmEscReturnsToFilledForm(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m = openFilledCreateConfirm(t, m, "ep-1", "New Task")

	tm, cmd := m.Update(keyMsg(tea.KeyEsc))
	nm, ok := tm.(model)
	if !ok {
		t.Fatalf("Update(esc) did not return a model, got %T", tm)
	}
	if nm.overlay != overlayNone {
		t.Fatalf("overlay after esc = %v, want overlayNone", nm.overlay)
	}
	if nm.form == nil {
		t.Fatal("esc did not reopen the Create-Form (Draft-Erhalt, DD2-190)")
	}
	// cmd is the reopened form's own m.form.Init() (huh's group/field
	// housekeeping, e.g. WindowSize/cursor setup) -- NOT a create dispatch;
	// pendingCreate/createLabel being cleared (below) is what proves no
	// creation actually fired.
	_ = cmd
	if nm.createDraft != nil {
		t.Fatal("createDraft must be consumed (nil) once reopened into the form")
	}
	if nm.pendingCreate != nil {
		t.Fatal("pendingCreate must be cleared, esc/n must not leave a stale create dispatch parked")
	}

	nm = advanceFieldsModel(t, nm, 7)
	if !strings.Contains(nm.createLabel, "New Task") {
		t.Fatalf("createLabel after reopen+resubmit = %q, want it to still mention %q (draft survived)", nm.createLabel, "New Task")
	}
}

// TestCreateDoneJumpsCursorAndExpandsParentPath guards applyCreateDone's
// success path (update.go): the new bean's immediate parent AND its own
// ancestors get expanded, the cursor jumps to the new bean's ID BEFORE the
// reload fires, and applyLoaded's existing "exact bean still present" path
// keeps the cursor there once the reload (a manually-fed beansLoadedMsg,
// mirroring fixtureModel's own convention) arrives.
func TestCreateDoneJumpsCursorAndExpandsParentPath(t *testing.T) {
	m := fixtureModel(t, fixtureBeans()) // ms-1 -> ep-1 -> tk-1, tk-2
	newBean := data.Bean{ID: "tk-9", Title: "New Task", Type: "task", Status: "todo", Priority: "normal", Parent: "ep-1"}

	tm, cmd := m.Update(createDoneMsg{bean: newBean})
	nm, ok := tm.(model)
	if !ok {
		t.Fatalf("Update(createDoneMsg) did not return a model, got %T", tm)
	}
	if !nm.expanded["ep-1"] {
		t.Fatalf("expanded[ep-1] = %v, want true (new bean's immediate parent must be expanded)", nm.expanded)
	}
	if !nm.expanded["ms-1"] {
		t.Fatalf("expanded[ms-1] = %v, want true (ep-1's own ancestor must be expanded too)", nm.expanded)
	}
	if nm.cursorID != "tk-9" {
		t.Fatalf("cursorID = %q, want tk-9 (jump to the new bean)", nm.cursorID)
	}
	if cmd == nil {
		t.Fatal("createDoneMsg (success) must trigger a reload Cmd")
	}

	nm = step(t, nm, beansLoadedMsg{beans: append(fixtureBeans(), newBean)})
	if nm.cursorID != "tk-9" {
		t.Fatalf("cursorID after reload = %q, want tk-9 (applyLoaded's exact-bean-present path)", nm.cursorID)
	}
}

// TestCreateDoneErrRoutesToApplyMutationResult guards the failure path: a
// createDoneMsg carrying an error must NOT touch the cursor -- it routes
// straight through the shared applyMutationResult tail (status line +
// unconditional reload), same as every other mutation.
func TestCreateDoneErrRoutesToApplyMutationResult(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	wantCursor := m.cursorID
	wantErr := fmt.Errorf("beans create: boom")

	tm, cmd := m.Update(createDoneMsg{err: wantErr})
	nm, ok := tm.(model)
	if !ok {
		t.Fatalf("Update(createDoneMsg) did not return a model, got %T", tm)
	}
	if nm.err != wantErr.Error() {
		t.Fatalf("m.err = %q, want %q", nm.err, wantErr.Error())
	}
	if nm.cursorID != wantCursor {
		t.Fatalf("cursorID = %q, want unchanged %q (a failed create must not jump the cursor)", nm.cursorID, wantCursor)
	}
	if cmd == nil {
		t.Fatal("createDoneMsg (error) must still trigger a reload Cmd (applyMutationResult)")
	}
}

// TestFormCapturesAllKeysWhileOpen guards the capture-order contract
// (handleKey doc-stamp): "q" while a Create-Form is open must type into the
// focused field, NOT request a quit -- same precedent as searchActive/
// filterOpen's full-capture behavior.
func TestFormCapturesAllKeysWhileOpen(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m = step(t, m, runeMsg('c'))
	if m.form == nil {
		t.Fatal("setup: c did not open the create form")
	}

	nm := step(t, m, runeMsg('q'))
	if nm.confirmQuit {
		t.Fatal("q while the create form is open must type 'q', not request quit (searchActive-style capture precedent)")
	}
	if nm.form == nil {
		t.Fatal("form must still be open after typing q into it")
	}
	if nm.overlay != overlayNone {
		t.Fatalf("overlay = %v, want overlayNone (q must not open anything either)", nm.overlay)
	}
}
