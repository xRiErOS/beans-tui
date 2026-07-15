package tui

// form_reject_review.go — the single-field Reject-Kommentar-Form (`x` in the
// Review-Cockpit, E4 Task 4, bean bt-yy6w, design decision e): the SAME
// huh-Form-Hosting infra E3 Task 4/5 built (styleForm/formChrome), one
// REQUIRED huh.NewText field ("comment", nonEmpty-validated -- unlike
// form_edit_title.go's Title-Edit-Form, there is no prefill: an unexplained
// rejection defeats the agent-feedback purpose the "## Review <datum>"
// section exists for). No Confirm-Gate -- submitForm's "reject" case
// (box_confirm_create.go) fires data.Client.RejectReview DIRECTLY on
// huh.StateCompleted, mirroring "editTitle"s Port devd isCreateKind
// exclusion (only "create"-kind forms get gated).

import (
	"beans-tui/internal/data"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

// buildRejectReviewForm constructs the keyed single-field form -- no prefill
// (Value binding), nonEmpty-required (Port devd forms_shared.go nonEmpty,
// shared with buildEditTitleForm/buildCreateBeanForm).
func buildRejectReviewForm() *huh.Form {
	return huh.NewForm(huh.NewGroup(
		huh.NewText().Key("comment").Title("Reject-Kommentar").Validate(nonEmpty),
	))
}

// openRejectForm opens the Reject-Kommentar-Form on b (`x`, keyReviewCockpit
// dispatch, view_review_cockpit.go -- the !reviewIsRework(b) guard lives in
// the caller). m.mutTarget captures WHICH bean the form acts on, reusing the
// value-menu/picker/editTitle convention (one node-action target at a time)
// -- the ETag is deliberately NOT captured here (design decision d, same as
// openEditTitleForm): submitForm's "reject" case re-reads it fresh via
// m.beanETag at submit time, not at open time.
func (m model) openRejectForm(b *data.Bean) (tea.Model, tea.Cmd) {
	m.mutTarget = b.ID
	m.formKind = "reject"
	m.form = m.styleForm(buildRejectReviewForm())
	return m, m.form.Init()
}
