package tui

// box_picker_parent.go — the Parent-Picker (`a`, E3 Task 3, bean bt-p1uz):
// single-select over data.EligibleParents(idx, b) (self + descendants +
// invalid types pre-filtered server-side rule mirror, design decision f)
// plus a "(No parent)" row pinned first (port beans-src parentpicker.go's
// clearParentItem). Enter applies IMMEDIATELY (SetParent/RemoveParent) and
// closes -- the SAME immediate-apply Enter semantics as the combined value
// menu (box_menu_value.go, design decision a3), NOT the Pending-Diff pattern
// the Tag-/Blocking-Picker use (box_picker_tag.go, box_picker_blocking.go):
// a bean has exactly ONE parent, there is nothing to diff.

import (
	"strings"

	"beans-tui/internal/data"
	"beans-tui/internal/theme"
	keybind "github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

// pickerItem is one row shared by the Parent-Picker (this file) and the
// Blocking-Picker (box_picker_blocking.go). id "" is the Parent-Picker's own
// "(No parent)" clear row (RemoveParent) -- the Blocking-Picker never
// produces an empty id, every one of its rows is a real bean.
//
// T5 (bean bt-4mo9, B06): prefix/title replace the old single pre-rendered
// `label` field (relationRow(cand, "", relationRowNoWrap) -- a fully
// concatenated, NEVER-wrapped string baked at open-time). Splitting them
// lets blockingPickerBox/parentPickerBox compose their OWN leading element
// (the D08 cursor bar and, for Blocking, the pending dot) onto prefix and
// hand BOTH to hangingIndentWrap at RENDER time, against the LIVE picker
// width (wideModalWidth(m.width)) -- the actual B06 fix (PO: overlay IDs
// broke mid-word because the old single-line label was never wrap-aware).
type pickerItem struct {
	id string
	// prefix is the themed glyph+ID half of a relation row
	// (relationRowPrefix, view_detail_bean.go: status icon, type icon,
	// Key-styled ID, trailing space) -- "" for the Parent-Picker's own
	// synthetic "(No parent)" clear row, which carries no relation glyphs.
	prefix string
	// title is the wrappable text: a real bean's Title, or the clear row's
	// plain "(No parent)" string (themed at render time, mirroring the
	// pre-T5 contract).
	title string
}

// parentPickerRowBudget caps the Parent-Picker's/Blocking-Picker's row
// window (plan Step 3/5: "windowAround bei Überlänge"). Unlike the Tree/
// Backlog panes, these floating overlays are not handed an explicit body
// height by their caller (modalPanel sizes to its content, not the other way
// around) -- there is no existing "clampModalHeight" helper to mirror
// (clampModalWidth, box_filter_facets.go, is width-only), so a fixed,
// generous cap keeps a repo with many eligible/blockable beans from growing
// the modal past the terminal instead of computing one from m.height.
const parentPickerRowBudget = 14

// buildParentItems assembles the Parent-Picker's row list: "(No parent)"
// pinned first (id "", port clearParentItem), then data.EligibleParents(idx,
// b) (self/descendants/invalid-types pre-filtered, sorted) rendered via the
// existing relationRow helper (view_detail_bean.go:66-68) -- same status-
// icon+type-icon+ID+title glyph order as every other bean row in the app.
func buildParentItems(idx *data.Index, b *data.Bean) []pickerItem {
	items := []pickerItem{{id: "", title: "(No parent)"}}
	for _, cand := range data.EligibleParents(idx, b) {
		items = append(items, pickerItem{id: cand.ID, prefix: relationRowPrefix(cand), title: cand.Title})
	}
	return items
}

// openParentPicker opens the Parent-Picker on the focused bean (the
// focusedBean()!=nil guard lives in the caller, keyNodeAction). Cursor
// starts on the bean's CURRENT parent row, falling back to index 0 (the
// "(No parent)" row) for an unparented bean, or one whose parent fell out
// of eligibility since (port beans-src parentpicker.go's selectedIndex seed,
// newParentPickerModel:175-182). mutTarget captures the bean ID only, never
// the etag (design decision d) -- a watch-reload between open and the
// eventual enter is honored automatically via beanETag.
func (m model) openParentPicker() model {
	b := m.focusedBean()
	if b == nil {
		return m
	}
	m.mutTarget = b.ID
	m.parentItems = buildParentItems(m.idx, b)

	m.menu = listState{}
	m.menu.setLen(len(m.parentItems))
	for i, it := range m.parentItems {
		if it.id == b.Parent {
			m.menu.cursor = i
			break
		}
	}

	m.overlay = overlayParentPicker
	return m
}

// keyParentPicker drives the open Parent-Picker: up/down move the cursor
// (navKey), enter applies the cursored row immediately (SetParent/
// RemoveParent) and closes, esc closes without mutating anything.
func (m model) keyParentPicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch navKey(msg.String()) {
	case "up":
		m.menu.move(-1)
		return m, nil
	case "down":
		m.menu.move(1)
		return m, nil
	}
	switch {
	case keybind.Matches(msg, keys.Back):
		m.overlay = overlayNone
		return m, nil
	case keybind.Matches(msg, keys.Enter):
		return m.applyParentPickerSelection()
	}
	return m, nil
}

// applyParentPickerSelection dispatches SetParent (a real row) or
// RemoveParent (the "(No parent)" row, id "") against a FRESH etag read
// from the live index (m.beanETag, update.go, design decision d -- never a
// captured copy). Always fires, regardless of whether the cursored row
// already matches the bean's current parent -- mirrors
// applyValueMenuSelection's own unconditional-dispatch Enter semantics
// (design decision a3), not a diff. ok==false means mutTarget vanished from
// the index between open and this enter -- the overlay closes with a
// status-line note instead of firing a doomed mutation (same vanished-target
// guard as applyValueMenuSelection/applyTagPickerDiff).
func (m model) applyParentPickerSelection() (tea.Model, tea.Cmd) {
	m.overlay = overlayNone
	if m.menu.cursor < 0 || m.menu.cursor >= len(m.parentItems) {
		return m, nil
	}
	it := m.parentItems[m.menu.cursor]
	id := m.mutTarget
	etag, ok := m.beanETag(id)
	if !ok {
		m.err = "Bean no longer exists — selection discarded"
		return m, nil
	}
	client := m.client
	if it.id == "" {
		return m, mutateCmd(func() error { return client.RemoveParent(id, etag) })
	}
	parentID := it.id
	return m, mutateCmd(func() error { return client.SetParent(id, parentID, etag) })
}

// parentPickerBox renders the floating Parent-Picker overlay -- D08 cursor
// treatment (ansi.Strip + Accent-wrap) for the cursored row, same convention
// as treeRows/backlogRows (view_browse_repo.go/view_browse_backlog.go)
// rather than box_menu_value.go's/box_picker_tag.go's theme.Header.Render:
// unlike those two overlays' PLAIN value/tag strings, a real row's prefix
// here is ALREADY themed (relationRowPrefix -- status-icon/type-icon
// colors, Key-styled ID), so re-wrapping it in a second style would clash;
// stripping to plain text and re-tinting Accent is the same technique the
// Tree/Backlog panes already use for exactly this reason -- T5 (bean
// bt-4mo9, B06) narrows that treatment to the PREFIX only (cursor bar +
// glyph+ID), matching hangingIndentWrap's own contract (prefix carries
// styling, text/title stays plain unless separately styled): the pending
// dot's SHAPE (blockingDot, box_picker_blocking.go) still reads under an
// Accent override, exactly mirroring the ▷/▶ marker-only convention T4
// introduced for the RELATIONS section (view_detail_bean.go). windowAround
// (parentPickerRowBudget) keeps a long eligible-parent list from overflowing
// the modal (plan Step 3) -- ACHTUNG (B06 Architektur-Vorgabe, documented
// judgment call): rows can now be MULTI-LINE (a wrapped long title), so
// "parentPickerRowBudget=14" caps 14 windowAround SLICE ELEMENTS (i.e. 14
// eligible-parent rows), not 14 visible terminal lines -- an unusually long
// run of unusually long titles in the visible window can still grow the
// modal past 14 screen lines. Accepted per the bean's own note: most bean
// titles stay single-line even at 85% width, height (parentPickerRowBudget)
// deliberately stays untouched (PO: "Höhe passt").
func (m model) parentPickerBox() string {
	var b strings.Builder
	b.WriteString(theme.Muted.Render("enter:set  esc:cancel") + "\n")

	w := wideModalWidth(m.width)
	contW := w - 2 // modalBox's own Padding(0,1) overhead -- border adds no further inner-width cost (empirically verified: Width() already absorbs padding, Border() only adds outside it)
	if contW < 8 {
		contW = 8
	}

	rows := make([]string, len(m.parentItems))
	for i, it := range m.parentItems {
		title := it.title
		if it.id == "" {
			title = theme.Dim.Render(title)
		}
		var prefix string
		if i == m.menu.cursor {
			plain := ansi.Strip(it.prefix)
			prefix = theme.Accent.Render("▌" + plain)
			title = theme.Accent.Render(ansi.Strip(title))
		} else {
			prefix = " " + it.prefix
		}
		rows[i] = hangingIndentWrap(prefix, title, contW)
	}
	rows = windowAround(rows, parentPickerRowBudget, m.menu.cursor)
	b.WriteString(strings.Join(rows, "\n"))
	if len(rows) > 0 {
		b.WriteString("\n")
	}
	if len(m.parentItems) == 1 {
		b.WriteString(theme.Muted.Render("(no eligible parent types)") + "\n")
	}
	return modalPanel("Assign parent", b.String(), "", w, theme.Mauve)
}
