package tui

// box_confirm_create.go — the Create-Form's Confirm-Gate (E3 Task 4, bean
// bt-y4ly, design decision e, Port devd box_confirm_create.go's 1:1
// pattern): after the Create-Form reaches huh.StateCompleted, a y/n-style
// Confirm modal opens BEFORE the real data.Client.Create fires. enter fires
// the already-built createCmd; n/esc returns to the BEFILLED form
// (Draft-Erhalt, Port DD2-190) instead of discarding the PO's work.
// beans-tui has only ONE gated form kind so far ("create") -- unlike devd's
// multi-kind isCreateKind, submitForm below dispatches on m.formKind
// directly.

import (
	"strings"
	"time"

	"beans-tui/internal/theme"
	keybind "github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// submitForm is called once the open form reaches huh.StateCompleted
// (updateForm, forms_shared.go). It captures the form's values into a
// beanDraft BEFORE nulling m.form (draftFromForm needs the still-open form),
// parks the resulting createCmd, and opens the Confirm-Gate
// (m.overlay = overlayCreateConfirm, design decision a2 -- see types.go's
// ERRATUM doc-stamp on why there is no separate createConfirm bool).
// "editTitle" (E3 Task 5, bean bt-sl45, design decision h) is NOT a
// create-kind: it fires mutateCmd(SetTitle) DIRECTLY against a FRESH etag
// (design decision d, m.beanETag) -- no Confirm-Gate detour (Port devd
// isCreateKind's own exclusion, only "create" is gated).
func (m model) submitForm() (tea.Model, tea.Cmd) {
	switch m.formKind {
	case "create":
		// F1 (Review-Runde 2, Async-Gap-Clobbering, Finding 1b): a SECOND
		// layer of the same single-create guard keyNodeAction's Create case
		// already enforces (update.go) -- under normal UI flow that guard
		// prevents a second Create-Form from ever opening while pendingCreate
		// is parked-or-in-flight, so this branch should be unreachable in
		// practice; kept anyway as belt-and-braces against the SAME
		// createDraft/pendingCreate slots getting cross-contaminated by a
		// second submit, cheap to keep in sync since both sites gate on the
		// same field.
		if m.pendingCreate != nil {
			m.form = nil
			m.formKind = ""
			m.err = createInFlightNote
			return m, nil
		}
		d := draftFromForm(m.form)
		m.createDraft = &d
		m.createLabel = createConfirmLabel(d)
		m.pendingCreate = createCmd(m.client, createOptsFromDraft(d))
		m.form = nil
		m.formKind = ""
		m.overlay = overlayCreateConfirm
		return m, nil
	case "editTitle":
		title := m.form.GetString("title")
		id := m.mutTarget
		m.form = nil
		m.formKind = ""
		etag, ok := m.beanETag(id)
		if !ok {
			m.err = "Bean nicht mehr vorhanden — Titel-Edit verworfen"
			return m, nil
		}
		client := m.client
		return m, mutateCmd(func() error { return client.SetTitle(id, title, etag) })
	case "reject":
		// E4 Task 4 (bean bt-yy6w, design decision e): mirrors "editTitle"s
		// direct-fire shape exactly -- no Confirm-Gate, ETag re-read fresh
		// at submit time (design decision d).
		comment := m.form.GetString("comment")
		id := m.mutTarget
		m.form = nil
		m.formKind = ""
		etag, ok := m.beanETag(id)
		if !ok {
			m.err = "Bean nicht mehr vorhanden — Ablehnung verworfen"
			return m, nil
		}
		client := m.client
		date := time.Now().Format("2006-01-02")
		return m, mutateCmd(func() error { return client.RejectReview(id, comment, date, etag) })
	}
	m.form = nil
	m.formKind = ""
	return m, nil
}

// createConfirmLabel describes the pending bean for the Confirm modal (Type
// + Title, Port devd createConfirmLabel's "issue" branch).
func createConfirmLabel(d beanDraft) string {
	typ := d.typ
	if typ == "" {
		typ = "task"
	}
	return "Bean (" + typ + "): " + strings.TrimSpace(d.title)
}

// keyCreateConfirm drives the open Confirm-Gate (overlayCreateConfirm,
// keyOverlay dispatch, update.go): enter fires the parked createCmd and
// closes; n/esc discards the CONFIRM without discarding the WORK -- it
// reopens the Create-Form pre-filled from createDraft (Draft-Erhalt, Port
// DD2-190).
func (m model) keyCreateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case keybind.Matches(msg, keys.Enter):
		cmd := m.pendingCreate
		m.overlay = overlayNone
		// F1 (Review-Runde 2, Async-Gap-Clobbering, Finding 1b): pendingCreate
		// is deliberately NOT nulled here anymore (a behavior change vs. the
		// original T4 version) -- it now doubles as the "a create is in
		// flight" guard for the WHOLE async gap between this enter and
		// createDoneMsg actually arriving (types.go doc-stamp), so
		// keyNodeAction's Create case / submitForm's "create" case can
		// refuse a second `c` during that window. Only applyCreateDone
		// (update.go) clears it, once createDoneMsg resolves either outcome.
		m.createLabel = ""
		// B01 (E3-T4-Review PFLICHT, closed in T5, bean bt-sl45):
		// createDraft is deliberately NOT nulled here anymore -- a
		// CLI-rejected create (e.g. VALIDATION_ERROR) must not lose the
		// PO's filled-in draft. Only createDoneMsg (applyCreateDone,
		// update.go) resolves it now, on EITHER outcome: success consumes
		// it (no reopen needed), a failure reopens the FILLED form from it
		// instead of routing through the draft-agnostic
		// applyMutationResult tail.
		return m, cmd
	case keybind.Matches(msg, keys.Back), msg.String() == "n":
		m.overlay = overlayNone
		m.pendingCreate = nil
		m.createLabel = ""
		if m.createDraft != nil {
			d := *m.createDraft
			m.createDraft = nil
			return m.openCreateFormWithDraft(d)
		}
		return m, nil
	}
	return m, nil
}

// createConfirmBox renders the Confirm-Gate modal (Mauve = constructive, not
// destructive -- Port devd createConfirmBox).
func (m model) createConfirmBox() string {
	var b strings.Builder
	b.WriteString(theme.Header.Render("Create?") + "\n\n")
	b.WriteString(theme.Accent.Render(m.createLabel) + "\n\n")
	b.WriteString(theme.Dim.Render("enter: create   esc/n: back"))
	return modalBox(b.String(), clampModalWidth(54, m.width), theme.Mauve)
}
