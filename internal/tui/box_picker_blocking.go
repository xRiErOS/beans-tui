package tui

// box_picker_blocking.go — the Blocking-Picker (`B`, E3 Task 3, bean
// bt-p1uz): toggle-multi-select over EVERY bean except the focused one
// itself (design decision g/f: deliberately NO cycle exclusion, port-parity
// with beans-src blockingpicker.go, which has none either -- a blocking
// cycle is a logical PO mistake, not a render hazard, unlike the
// Parent-Picker's tree-flatten concern, box_picker_parent.go/data/
// hierarchy.go). Pending-Diff pattern ported verbatim from box_picker_tag.go
// (E3 Task 2): space toggles pending, enter diffs pending against original
// and fires ONE combined data.SetBlocking mutateCmd (mirrors SetTags' single-
// etag-no-cascade rationale), esc discards.

import (
	"sort"
	"strings"

	"beans-tui/internal/data"
	"beans-tui/internal/theme"
	keybind "github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// buildBlockingItems lists every bean except selfID, rendered via
// relationRow (view_detail_bean.go:66-68, same glyph order as every other
// bean row in the app). idx.ByID is a Go map, so a baseline ID sort runs
// first before the single-source data.SortBeans pass -- same two-pass
// determinism fix box_filter_facets.go's tagFilterOptions documents (ties on
// Status/Priority/Type/Title alike would otherwise be map-walk-order-
// dependent).
func buildBlockingItems(idx *data.Index, selfID string) []pickerItem {
	if idx == nil {
		return nil
	}
	all := make([]*data.Bean, 0, len(idx.ByID))
	for _, cand := range idx.ByID {
		if cand.ID == selfID {
			continue
		}
		all = append(all, cand)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].ID < all[j].ID })
	data.SortBeans(all)

	items := make([]pickerItem, len(all))
	for i, cand := range all {
		items[i] = pickerItem{id: cand.ID, label: relationRow(cand)}
	}
	return items
}

// openBlockingPicker opens the Blocking-Picker on the focused bean (the
// focusedBean()!=nil guard lives in the caller, keyNodeAction).
// blockOriginal/blockPending are seeded as two INDEPENDENT maps (same
// wholesale-replace convention as openTagPicker) from the focused bean's
// CURRENT Blocking field. mutTarget captures the bean ID only, never the
// etag (design decision d).
func (m model) openBlockingPicker() model {
	b := m.focusedBean()
	if b == nil {
		return m
	}
	m.mutTarget = b.ID
	m.blockItems = buildBlockingItems(m.idx, b.ID)

	orig := make(map[string]bool, len(b.Blocking))
	for _, id := range b.Blocking {
		orig[id] = true
	}
	pending := make(map[string]bool, len(orig))
	for id := range orig {
		pending[id] = true
	}
	m.blockOriginal = orig
	m.blockPending = pending

	m.menu = listState{}
	m.menu.setLen(len(m.blockItems))

	m.overlay = overlayBlockingPicker
	return m
}

// keyBlockingPicker drives the open Blocking-Picker: up/down move the cursor
// (navKey), space/x (keys.Toggle) toggles the cursored row's pending state,
// enter diffs pending against original (applyBlockingPickerDiff), esc
// discards everything and closes.
func (m model) keyBlockingPicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
	case keybind.Matches(msg, keys.Toggle):
		return m.toggleBlockPending(), nil
	case keybind.Matches(msg, keys.Enter):
		return m.applyBlockingPickerDiff()
	}
	return m, nil
}

// toggleBlockPending flips the cursored row's membership in m.blockPending.
// I01 (types.go doc-stamp): clones via cloneBoolMap before writing -- same
// convention as toggleTagPending.
func (m model) toggleBlockPending() model {
	if m.menu.cursor < 0 || m.menu.cursor >= len(m.blockItems) {
		return m
	}
	id := m.blockItems[m.menu.cursor].id
	out := cloneBoolMap(m.blockPending)
	if out[id] {
		delete(out, id)
	} else {
		out[id] = true
	}
	m.blockPending = out
	return m
}

// applyBlockingPickerDiff computes blockPending's diff against
// blockOriginal and fires it as ONE combined data.SetBlocking mutateCmd
// (mirrors applyTagPickerDiff's rationale verbatim -- N sequential
// AddBlocking/RemoveBlocking calls against ONE etag would be a conflict
// cascade: the first call rotates the etag on disk, every subsequent call
// then sees a stale etag and fails ErrConflict) against a FRESH etag
// (m.beanETag, design decision d). No pending changes at all -> no Cmd. add/
// remove are sorted before dispatch for a deterministic CLI-invocation
// order.
func (m model) applyBlockingPickerDiff() (tea.Model, tea.Cmd) {
	m.overlay = overlayNone
	id := m.mutTarget

	var add, remove []string
	for target := range m.blockPending {
		if !m.blockOriginal[target] {
			add = append(add, target)
		}
	}
	for target := range m.blockOriginal {
		if !m.blockPending[target] {
			remove = append(remove, target)
		}
	}
	if len(add) == 0 && len(remove) == 0 {
		return m, nil
	}
	sort.Strings(add)
	sort.Strings(remove)

	etag, ok := m.beanETag(id)
	if !ok {
		m.err = "Bean no longer exists — selection discarded"
		return m, nil
	}
	client := m.client
	return m, mutateCmd(func() error { return client.SetBlocking(id, add, remove, etag) })
}

// blockingDot is the ●/○ pending indicator (port beans-src
// blockingpicker.go's blockingIndicator, Red/Muted) -- rendered separately
// from the row's own relationRow theming, prefixed onto the row text before
// the shared D08 cursor treatment strips/re-tints the whole line.
func blockingDot(pending bool) string {
	if pending {
		return lipgloss.NewStyle().Foreground(theme.Red).Render("● ")
	}
	return theme.Muted.Render("○ ")
}

// blockingPickerBox renders the floating Blocking-Picker overlay -- ●/○
// indicator + relationRow-styled row, D08 cursor treatment (ansi.Strip +
// Accent-wrap, see parentPickerBox's doc comment for the full rationale)
// since each row already carries its own theme colors. windowAround
// (parentPickerRowBudget, box_picker_parent.go) keeps a repo with many beans
// from overflowing the modal.
func (m model) blockingPickerBox() string {
	var b strings.Builder
	b.WriteString(theme.Muted.Render("space/x:toggle  enter:save  esc:discard") + "\n")

	rows := make([]string, len(m.blockItems))
	for i, it := range m.blockItems {
		text := blockingDot(m.blockPending[it.id]) + it.label
		if i == m.menu.cursor {
			plain := ansi.Strip(text)
			rows[i] = theme.Accent.Render("▌" + plain)
		} else {
			rows[i] = "  " + text
		}
	}
	rows = windowAround(rows, parentPickerRowBudget, m.menu.cursor)
	b.WriteString(strings.Join(rows, "\n"))
	if len(rows) > 0 {
		b.WriteString("\n")
	}
	if len(m.blockItems) == 0 {
		b.WriteString(theme.Muted.Render("(no other beans in repo)") + "\n")
	}
	return modalPanel("Blocking", b.String(), "", clampModalWidth(48, m.width), theme.Mauve)
}
