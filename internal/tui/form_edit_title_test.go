package tui

// form_edit_title_test.go — TDD coverage for the Title-Edit-Form (E3 Task 5,
// bean bt-sl45, design decision h): a single-field huh form, prefilled with
// the bean's current title, nonEmpty-validated, that fires
// mutateCmd(SetTitle) DIRECTLY on completion -- no Confirm-Gate (Port devd
// isCreateKind's own exclusion: only "create"-kind forms get gated). Once
// reachable via "e"; D01 (design-spec.md §15 PF-17, bean bt-z4b1) moved "e"
// to the whole-bean $EDITOR exclusively -- this form is now reachable ONLY
// via enter on the title: field (activateDetailField, update.go).

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

// TestBuildEditTitleFormUsesMultiLineText guards B03's core swap (design-
// spec.md §15 PF-17, bean bt-2v38): the title field itself must now be a
// *huh.Text (multi-line textarea, huh v1.0.0 field_text.go), not the
// single-line *huh.Input E3 Task 5 originally wired. huh.Text exposes no
// getter to read .Lines(3) back off the field directly, so the
// CONFIGURATION is checked via its own documented, version-independent
// render contract instead (bubbles textarea.go:1182-1199, "Always show at
// least `m.Height` lines"): a Text field's View() ALWAYS reserves exactly
// its configured height's worth of rows regardless of content -- strictly
// MORE rows than an equivalent *huh.Input's fixed single row renders for
// the SAME title/width. That row-count delta IS the observable proof
// .Lines(3) took effect, not just the bare type swap.
func TestBuildEditTitleFormUsesMultiLineText(t *testing.T) {
	f := buildEditTitleForm("Original Title")

	field := f.GetFocusedField()
	textField, ok := field.(*huh.Text)
	if !ok {
		t.Fatalf("field type = %T, want *huh.Text (B03: huh.NewInput -> huh.NewText)", field)
	}
	textField.WithWidth(40)
	textRows := strings.Count(textField.View(), "\n") + 1

	v := "Original Title"
	inputField := huh.NewInput().Key("title").Title("Title").Value(&v).Validate(nonEmpty)
	inputField.WithWidth(40)
	inputRows := strings.Count(inputField.View(), "\n") + 1

	if textRows <= inputRows {
		t.Fatalf("multi-line field rendered %d row(s), want MORE than the single-line Input's %d row(s) for the same title/width (Lines(3) must reserve extra height)", textRows, inputRows)
	}
}

// TestOpenEditTitleFormRoundTripsLongTitleUnchanged guards B03's data-
// integrity contract: a title long enough to need the new multi-line wrap
// (.Lines(3), > 60 chars) must survive an UNEDITED submit byte-for-byte --
// no truncation, no embedded newline snuck in by the Input->Text field-type
// swap. Two halves, both a REAL huh.Form/model tea.Update round-trip (Port
// devd-style, ANALOG TestEditTitleSubmitFiresSetTitleDirectlyNoConfirm
// above): (1) the exact string huh hands back via GetString("title") -- the
// SAME untransformed read submitForm's "editTitle" case performs
// (box_confirm_create.go, `title := m.form.GetString("title")`), checked
// directly since this suite has no PATH-stubbed "beans" binary to intercept
// the mutation's real CLI args (every sibling test in this package settles
// for the same "beans update" dispatch-proof instead); (2) submitForm
// itself still fires the SAME mutateCmd with no Confirm-Gate detour.
func TestOpenEditTitleFormRoundTripsLongTitleUnchanged(t *testing.T) {
	long := "A bean title that runs well past sixty characters so the new multi-line Text field must wrap it instead of truncating or scrolling"
	if len(long) <= 60 {
		t.Fatalf("setup: fixture title is %d chars, want > 60", len(long))
	}

	f := buildEditTitleForm(long)
	f = driveForm(f, enterMsg())
	if f.State != huh.StateCompleted {
		t.Fatalf("setup: form.State = %v, want StateCompleted after the single field's enter", f.State)
	}
	if got := f.GetString("title"); got != long {
		t.Fatalf("GetString(title) after an unedited submit = %q (%d chars), want the full unchanged title (%d chars) -- B03 must not truncate", got, len(got), len(long))
	}
	if strings.Contains(f.GetString("title"), "\n") {
		t.Fatal("GetString(title) contains a newline -- B03's multi-line Text field must not leak its internal line-wrap into the stored value")
	}

	m := fixtureModel(t, fixtureBeans())
	m.client = &data.Client{RepoDir: "/nonexistent-bt-2v38-scratch-dir"}
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
	if cmd == nil {
		t.Fatal("submitForm(editTitle) must fire a Cmd (mutateCmd(SetTitle))")
	}
	msg := cmd()
	mdm, ok := msg.(mutationDoneMsg)
	if !ok {
		t.Fatalf("cmd() = %T, want mutationDoneMsg", msg)
	}
	if mdm.err == nil || !strings.Contains(mdm.err.Error(), "beans update") {
		t.Fatalf("mutationDoneMsg.err = %v, want an error containing %q (proves SetTitle dispatched, same GetString(title) verified full/unchanged above)", mdm.err, "beans update")
	}
}

// TestKeyNodeActionEDoesNotOpenEditTitleFormAnymore is D01's regression
// guard for this file's own scope (design-spec.md §15 PF-17, bean bt-z4b1,
// SUPERSEDES design decision h's original "e opens Title-Edit-Form"
// contract, E3 Task 5): "e" now ALWAYS opens the whole-bean $EDITOR
// (TestKeyNodeActionEditorAlwaysOpensBeanEditor, update_test.go) and NEVER
// the Title-Edit-Form anymore -- that form is reachable ONLY via enter on
// the title: field now (TestKeyDetailFocusEnterOnTitleFieldOpensEditTitleForm,
// update_test.go, activateDetailField's "title" case, unchanged by D01).
func TestKeyNodeActionEDoesNotOpenEditTitleFormAnymore(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m = focusBean(m, "tk-1")

	m = step(t, m, runeMsg('e'))
	if m.form != nil {
		t.Fatalf("e opened a form (form=%v formKind=%q), want NO form -- e now ALWAYS opens the whole-bean $EDITOR (D01)", m.form, m.formKind)
	}
	if m.editorTarget != "tk-1" {
		t.Fatalf("editorTarget = %q, want tk-1 (e must open the whole-bean $EDITOR instead)", m.editorTarget)
	}
}
