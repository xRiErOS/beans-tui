package tui

// form_reject_review_test.go — TDD coverage for the Reject-Kommentar-Form
// (`x` in the Review-Cockpit, E4 Task 4, bean bt-yy6w, design decision e): a
// single required huh.Text field, no prefill, no Confirm-Gate -- fires
// mutateCmd(RejectReview) DIRECTLY on huh.StateCompleted. Mirrors
// form_edit_title_test.go's structure (buildX/openX/submitForm coverage).

import (
	"strings"
	"testing"

	"beans-tui/internal/data"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

// TestBuildRejectReviewFormRequiresComment guards nonEmpty on the "comment"
// field: an empty submit attempt must not reach huh.StateCompleted (Port
// devd forms_shared.go nonEmpty, same validator form_create_bean.go/
// form_edit_title.go already share).
func TestBuildRejectReviewFormRequiresComment(t *testing.T) {
	f := buildRejectReviewForm()
	f = driveForm(f, enterMsg()) // submit attempt WITHOUT ever typing anything
	if f.State == huh.StateCompleted {
		t.Fatal("an empty comment reached StateCompleted, want the nonEmpty validator to block it")
	}
}

// TestRejectSubmitFiresRejectReviewDirectlyNoConfirm guards design decision
// e's core contract: completing the single field fires
// mutateCmd(RejectReview) IMMEDIATELY via submitForm's "reject" case -- no
// overlayCreateConfirm detour. Types a comment via runes, then settles the
// field with enter, mirroring TestEditTitleSubmitFiresSetTitleDirectlyNoConfirm's
// drive-then-inspect-before-firing pattern (form_edit_title_test.go).
func TestRejectSubmitFiresRejectReviewDirectlyNoConfirm(t *testing.T) {
	f := buildRejectReviewForm()
	// Drive Init()'s own Cmd chain FIRST: unlike buildEditTitleForm's
	// prefilled Input (set at construction time via .Value(&v), read
	// without ever needing focus), huh's Text field only accepts typed rune
	// input into its underlying bubbles textarea once Focus() has actually
	// fired -- which Group.Init() only does when the group is
	// `active` (Form.Init()'s own group-0-active bootstrap). Skipping this
	// leaves every typed rune silently dropped (empirically confirmed:
	// GetString stayed "" without this call) while the unconditional
	// enter/Next/Submit validation path (field_text.go) still ran and kept
	// correctly blocking an empty value -- which is exactly why
	// TestBuildRejectReviewFormRequiresComment above never needed this.
	f, _ = driveFormCmd(f, f.Init(), driveFormBudget)
	f = driveForm(f, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("bitte X korrigieren")})
	f = driveForm(f, enterMsg())
	if f.State != huh.StateCompleted {
		t.Fatalf("setup: form.State = %v, want StateCompleted after the single field's enter", f.State)
	}
	if got := f.GetString("comment"); got != "bitte X korrigieren" {
		t.Fatalf("GetString(comment) = %q, want the typed text", got)
	}

	m := fixtureModel(t, reviewFixtureBeans())
	m.client = &data.Client{RepoDir: "/nonexistent-bt-e4-t4-scratch-dir"}
	m.form = f
	m.formKind = "reject"
	m.mutTarget = "tk-a2"

	tm, cmd := m.submitForm()
	nm, ok := tm.(model)
	if !ok {
		t.Fatalf("submitForm() did not return a model, got %T", tm)
	}
	if nm.form != nil {
		t.Fatal("form still open after submitForm, want nil (reject: no Confirm-Gate)")
	}
	if nm.overlay != overlayNone {
		t.Fatalf("overlay = %v, want overlayNone (design decision e: no Confirm-Gate for reject)", nm.overlay)
	}
	if cmd == nil {
		t.Fatal("submitForm(reject) must fire a Cmd (mutateCmd(RejectReview), no Confirm parking)")
	}
	msg := cmd()
	mdm, ok := msg.(mutationDoneMsg)
	if !ok {
		t.Fatalf("cmd() = %T, want mutationDoneMsg", msg)
	}
	if mdm.err == nil || !strings.Contains(mdm.err.Error(), "beans update") {
		t.Fatalf("mutationDoneMsg.err = %v, want an error containing %q (proves RejectReview dispatched)", mdm.err, "beans update")
	}
}

// TestRejectSubmitVanishedTargetSurfacesError guards the same "bean deleted
// between form-open and submit" guard editTitle's submitForm case already
// has: m.beanETag(id) ok==false -> m.err set, NO Cmd fired.
func TestRejectSubmitVanishedTargetSurfacesError(t *testing.T) {
	f := buildRejectReviewForm()
	f, _ = driveFormCmd(f, f.Init(), driveFormBudget) // see TestRejectSubmitFiresRejectReviewDirectlyNoConfirm's doc comment
	f = driveForm(f, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("bitte X korrigieren")})
	f = driveForm(f, enterMsg())
	if f.State != huh.StateCompleted {
		t.Fatalf("setup: form.State = %v, want StateCompleted", f.State)
	}

	m := fixtureModel(t, reviewFixtureBeans())
	m.form = f
	m.formKind = "reject"
	m.mutTarget = "does-not-exist"

	tm, cmd := m.submitForm()
	nm, ok := tm.(model)
	if !ok {
		t.Fatalf("submitForm() did not return a model, got %T", tm)
	}
	if cmd != nil {
		t.Fatal("submitForm(reject) on a vanished target must not fire a Cmd")
	}
	if nm.err == "" {
		t.Fatal("submitForm(reject) on a vanished target must surface m.err")
	}
}

// TestOpenRejectFormSetsMutTargetAndFormKind guards openRejectForm directly
// (view_review_cockpit.go's keyReviewCockpit "x" case calls this).
func TestOpenRejectFormSetsMutTargetAndFormKind(t *testing.T) {
	m := fixtureModel(t, reviewFixtureBeans())
	b := m.idx.ByID["tk-a2"]

	nm, _ := m.openRejectForm(b)
	mm, ok := nm.(model)
	if !ok {
		t.Fatalf("openRejectForm did not return a model, got %T", nm)
	}
	if mm.form == nil || mm.formKind != "reject" {
		t.Fatalf("openRejectForm did not open the form (form=%v formKind=%q)", mm.form, mm.formKind)
	}
	if mm.mutTarget != "tk-a2" {
		t.Fatalf("mutTarget = %q, want tk-a2", mm.mutTarget)
	}
}
