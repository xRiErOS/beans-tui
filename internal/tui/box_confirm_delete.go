package tui

// box_confirm_delete.go — the Delete-Confirm modal (`d`, E3 Task 6, bean
// bt-ppzb): Port devd box_confirm_delete.go:57-126 (keyDelete enter/esc/n,
// deleteBox Rot=destruktiv, modalBox theme.Red) -- MINUS devd's
// loadDeletePreview (idx.Children[id] is already synchronous in memory here,
// no async count-load like devd's cascade-delete preview needs) and MINUS
// devd's removeIssueFromCaches (beans-tui does not cache per-view: the
// unconditional reload every mutationDoneMsg already triggers, design
// decision d, IS the cache invalidation).
//
// Semantic deviation from devd (bean bt-ppzb Ziel, mutations.go Delete's own
// doc-stamp): `beans delete` does NOT cascade -- there is no server-side
// cascade-delete, `--json` only skips the confirmation prompt and any
// reference/child WARNINGS, it does not delete anything beyond the one bean.
// deleteBox's copy therefore never claims children get deleted (devd's own
// "will also be deleted" cascade wording does not apply here).
//
// ERRATUM (empirically verified against beans 0.4.2 during this task's tmux
// smoke test + an isolated `beans create`/`beans delete` probe, both
// reproducible): the epic-E3-plan.md/bean bt-ppzb's ORIGINAL assumption --
// that a deleted bean's direct children survive with a now-DANGLING Parent
// field and surface under the synthetic "(verwaist)" root -- is WRONG. The
// real beans 0.4.2 `delete` command actively rewrites every direct child's
// frontmatter, REMOVING the `parent:` field entirely rather than leaving it
// dangling. A deleted bean's direct children therefore become genuine,
// ordinary ROOT beans (idx.Roots(), view_browse_repo.go) -- NOT
// "(verwaist)"-bucket orphans; they render at the top level exactly like any
// other parentless bean, not flagged or quarantined in any way. delChildren
// (below) is still an accurate PRE-delete count (computed from the CURRENT
// in-memory idx.Children before the mutation fires) -- only the POST-delete
// outcome the modal's copy describes was corrected. Grandchildren are
// unaffected either way: their own parent (the direct child, untouched by
// this rewrite) stays intact, so only DIRECT children are ever in play here
// (idx.Children[id], never data.CollectDescendants's full recursive walk).
// Regression: internal/data/client_mut_test.go's
// TestDeleteClearsFormerChildrensParentField.

import (
	"fmt"
	"strings"

	"beans-tui/internal/theme"
	keybind "github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// openDeleteConfirm opens the Delete-Confirm overlay on the focused bean (the
// focusedBean()!=nil guard mirrors every other T1-T3 open* helper --
// keyNodeAction's own guard already covers the real dispatch path, this one
// makes openDeleteConfirm safe to call directly in tests too). mutTarget
// captures the bean ID (the shared field every other overlay already uses);
// delTitle/delChildren are captured HERE, at open time, purely for the
// modal's own render -- delChildren is idx.Children[id]'s DIRECT count only
// (types.go doc-stamp: NOT data.CollectDescendants's full recursive walk,
// since `beans delete` does not cascade -- only direct children orphan).
func (m model) openDeleteConfirm() model {
	b := m.focusedBean()
	if b == nil {
		return m
	}
	m.mutTarget = b.ID
	m.delTitle = b.Title
	m.delChildren = 0
	if m.idx != nil {
		m.delChildren = len(m.idx.Children[b.ID])
	}
	m.overlay = overlayDeleteConfirm
	return m
}

// keyDeleteConfirm drives the open Delete-Confirm: enter fires
// data.Client.Delete (NO --if-match -- the beans CLI's `delete` subcommand
// takes no etag at all, mutations.go Delete's own doc-stamp; m.beanETag(id)
// is still consulted here, but ONLY for its ok bool, the same vanished-
// target guard shape as every other overlay's applyXSelection/
// applyEditorFinished), esc/n cancels without mutating anything (Port devd
// keyDelete's own esc/n cancel, DD2-174). Cursor-clamping after a successful
// delete needs NO new logic here: applyMutationResult's unconditional reload
// (design decision d) runs applyLoaded's EXISTING oldPos-fallback path
// (update.go, T8/E1) exactly like any other reload that drops the cursored
// bean.
func (m model) keyDeleteConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case keybind.Matches(msg, keys.Back), msg.String() == "n":
		m.overlay = overlayNone
		return m, nil
	case keybind.Matches(msg, keys.Enter):
		m.overlay = overlayNone
		id := m.mutTarget
		if _, ok := m.beanETag(id); !ok {
			m.err = "Bean nicht mehr vorhanden — Löschung verworfen"
			return m, nil
		}
		client := m.client
		return m, mutateCmd(func() error { return client.Delete(id) })
	}
	return m, nil
}

// deleteBox renders the floating Delete-Confirm modal -- Red border/header
// (destructive, Port devd's theme.Red convention, mirrors quitBox's own
// Mauve-for-non-destructive counterpart, box_confirm_quit.go). typ resolves
// from the LIVE index at render time (mirrors valueMenuBox's own
// m.idx.ByID[m.mutTarget] lookup) rather than a cached field, since it is
// only needed for display, not for dispatch.
func (m model) deleteBox() string {
	typ := "bean"
	if m.idx != nil {
		if cur, ok := m.idx.ByID[m.mutTarget]; ok {
			typ = cur.Type
		}
	}

	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Foreground(theme.Red).Bold(true).Render("Delete "+typ) + "\n")
	b.WriteString(theme.Header.Render(m.delTitle) + "\n\n")
	if m.delChildren > 0 {
		b.WriteString(fmt.Sprintf("%d Kind(er) verlieren den Parent — werden zu eigenen Wurzeln\n", m.delChildren))
	}
	b.WriteString("\n" + lipgloss.NewStyle().Foreground(theme.Red).Render("Irreversible.") + "\n")
	b.WriteString("\n" + theme.Dim.Render("enter: delete permanently   esc/n: cancel"))
	return modalBox(b.String(), clampModalWidth(48, m.width), theme.Red)
}
