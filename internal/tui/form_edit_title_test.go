package tui

// form_edit_title_test.go — TDD coverage for the Title-Edit-Form (`e`, E3
// Task 5, bean bt-sl45, design decision h): a single-field huh form, prefilled
// with the bean's current title, nonEmpty-validated, that fires
// mutateCmd(SetTitle) DIRECTLY on completion -- no Confirm-Gate (Port devd
// isCreateKind's own exclusion: only "create"-kind forms get gated).

import (
	"strings"
	"testing"

	"beans-tui/internal/data"
	"github.com/charmbracelet/huh"
)

// TestEditTitleFormPrefilledAndNonEmptyValidated guards both halves of
// buildEditTitleForm: the current title is bound as the field's initial
// value (no round-trip needed to observe it), and nonEmpty blocks an empty
// title from ever reaching huh.StateCompleted (Port devd forms_shared.go
// nonEmpty).
func TestEditTitleFormPrefilledAndNonEmptyValidated(t *testing.T) {
	f := buildEditTitleForm("Original Title")
	// GetString only reflects huh's f.results after a commit round-trip
	// (driveFormBudget's own doc-stamp, form_create_bean_test.go) -- settle
	// the single field WITHOUT editing it to observe the prefill survived.
	f = driveForm(f, enterMsg())
	if got := f.GetString("title"); got != "Original Title" {
		t.Fatalf("GetString(title) after settling unedited = %q, want prefill %q", got, "Original Title")
	}

	empty := buildEditTitleForm("")
	empty = driveForm(empty, enterMsg())
	if empty.State == huh.StateCompleted {
		t.Fatal("an empty title reached StateCompleted, want the nonEmpty validator to block it")
	}
}

// TestEditTitleSubmitFiresSetTitleDirectlyNoConfirm guards design decision
// h's core contract: completing the single field fires mutateCmd(SetTitle)
// IMMEDIATELY via submitForm's "editTitle" case -- no overlayCreateConfirm
// detour. Drives the underlying huh.Form to StateCompleted directly (Port
// devd's own single-field editField precedent) rather than through
// model.Update, so submitForm's returned Cmd can be inspected BEFORE it
// fires (mirrors box_picker_parent_test.go/box_confirm_create_test.go's
// nonexistent-RepoDir dispatch-proof pattern) -- the model-level chase
// (driveModel) would otherwise auto-execute the mutateCmd AND its
// applyMutationResult reload tail, whose LATER error would overwrite m.err
// and mask the one under test.
func TestEditTitleSubmitFiresSetTitleDirectlyNoConfirm(t *testing.T) {
	f := buildEditTitleForm("Original Title")
	f = driveForm(f, enterMsg())
	if f.State != huh.StateCompleted {
		t.Fatalf("setup: form.State = %v, want StateCompleted after the single field's enter", f.State)
	}

	m := fixtureModel(t, fixtureBeans())
	m.client = &data.Client{RepoDir: "/nonexistent-bt-e3-t5-scratch-dir"}
	m.form = f
	m.formKind = "editTitle"
	m.mutTarget = "tk-1"

	tm, cmd := m.submitForm()
	nm, ok := tm.(model)
	if !ok {
		t.Fatalf("submitForm() did not return a model, got %T", tm)
	}
	if nm.form != nil {
		t.Fatal("form still open after submitForm, want nil (editTitle: no Confirm-Gate)")
	}
	if nm.overlay != overlayNone {
		t.Fatalf("overlay = %v, want overlayNone (design decision h: no Confirm-Gate for edits)", nm.overlay)
	}
	if cmd == nil {
		t.Fatal("submitForm(editTitle) must fire a Cmd (mutateCmd(SetTitle), no Confirm parking)")
	}
	msg := cmd()
	mdm, ok := msg.(mutationDoneMsg)
	if !ok {
		t.Fatalf("cmd() = %T, want mutationDoneMsg", msg)
	}
	if mdm.err == nil || !strings.Contains(mdm.err.Error(), "beans update") {
		t.Fatalf("mutationDoneMsg.err = %v, want an error containing %q (proves SetTitle dispatched)", mdm.err, "beans update")
	}
}

// TestKeyNodeActionEOpensEditTitleForm guards keyNodeAction's Editor
// dispatch (design decision h): "e" opens the Title-Edit-Form prefilled from
// the focused bean, captured on m.mutTarget (reused from the value-menu/
// picker convention) -- distinct from "ctrl+e" (editor_test.go), which must
// NOT open a form.
func TestKeyNodeActionEOpensEditTitleForm(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m = focusBean(m, "tk-1")

	m = step(t, m, runeMsg('e'))
	if m.form == nil || m.formKind != "editTitle" {
		t.Fatalf("e did not open the edit-title form (form=%v formKind=%q)", m.form, m.formKind)
	}
	if m.mutTarget != "tk-1" {
		t.Fatalf("mutTarget = %q, want tk-1", m.mutTarget)
	}
	f := driveForm(m.form, enterMsg()) // settle unedited -- observe the prefill (see the Prefilled test above)
	if got := f.GetString("title"); got != "Task One" {
		t.Fatalf("GetString(title) = %q, want the focused bean's current title %q", got, "Task One")
	}
}
