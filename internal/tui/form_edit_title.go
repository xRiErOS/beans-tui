package tui

// form_edit_title.go — the single-field Title-Edit-Form (`e`, E3 Task 5,
// bean bt-sl45, design decision h): the SAME huh-Form-Hosting infra Task 4
// built (styleForm/formChrome), but with exactly ONE field and NO
// Confirm-Gate -- edits fire mutateCmd(SetTitle) DIRECTLY on
// huh.StateCompleted (submitForm's "editTitle" case, box_confirm_create.go
// -- Port devd isCreateKind's own exclusion: only "create"-kind forms get
// gated, edits never do). Port devd forms_shared.go:332-360
// (buildEditFieldForm), reduced to the one field this task needs (Title,
// Input, nonEmpty-required) -- beans-tui has no generic multi-field
// editField concept yet.

import (
	"beans-tui/internal/data"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

// buildEditTitleForm constructs the keyed single-field form, pre-filled with
// the bean's current title (huh.Input.Value(&v), field_input.go), nonEmpty-
// required (Port devd forms_shared.go nonEmpty).
func buildEditTitleForm(title string) *huh.Form {
	v := title
	field := huh.NewInput().Key("title").Title("Title").Value(&v).Validate(nonEmpty)
	return huh.NewForm(huh.NewGroup(field))
}

// openEditTitleForm opens the Title-Edit-Form on b (`e`, keyNodeAction
// dispatch, update.go -- the focusedBean()!=nil guard lives in the caller).
// m.mutTarget captures WHICH bean the form acts on, reusing the value-menu/
// picker convention (one node-action target at a time; forms and overlays
// are mutually exclusive capture states, design decision a2, so there is no
// collision) -- the ETag is deliberately NOT captured here (design decision
// d): submitForm's "editTitle" case re-reads it fresh via m.beanETag.
func (m model) openEditTitleForm(b *data.Bean) (tea.Model, tea.Cmd) {
	m.mutTarget = b.ID
	m.formKind = "editTitle"
	m.form = m.styleForm(buildEditTitleForm(b.Title))
	return m, m.form.Init()
}
