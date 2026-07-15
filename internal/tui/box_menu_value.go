package tui

// box_menu_value.go — the combined Status/Type/Priority value menu (`s`, E3
// Task 1, bean bt-dlgk, design decision a3): design-spec.md §7 reserves
// exactly ONE key for this cluster (no separate Type/Priority key), so all
// three beans enums render as ONE modalPanel with three group headers (the
// SAME facetHead render pattern box_filter_facets.go's treeFilterBox already
// established) instead of three separate picker overlays -- beans-src's
// statuspicker.go/typepicker.go/prioritypicker.go are structurally
// identical, consolidated here per design decision a3. Enter applies the
// cursored value IMMEDIATELY and closes (port beans-src statuspicker.go
// Enter-semantics, statuspicker.go:158-172) -- unlike the filter menu, whose
// enter only closes without mutating anything (filter facets act live).

import (
	"strings"

	"beans-tui/internal/data"
	"beans-tui/internal/theme"
	keybind "github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// valueMenuItem is one row of the combined s-menu. group is "status",
// "type", or "priority" -- value is one of that group's canonical enum
// members (data.StatusValues/TypeValues/PriorityValues, the E3 enum
// single-source, design decision b).
type valueMenuItem struct{ group, value string }

// valueMenuGroupHead labels each group's sub-header -- same facetHead-style
// render pattern as box_filter_facets.go's treeFilterBox.
var valueMenuGroupHead = map[string]string{
	"status":   "Status",
	"type":     "Type",
	"priority": "Priority",
}

// buildValueMenuItems assembles the combined menu's 15 rows (5 status + 5
// type + 5 priority), each group in the SAME canonical tier order the rest
// of the app already sorts by (data.StatusValues/TypeValues/PriorityValues,
// design decision b -- single source; box_filter_facets.go's
// buildFilterItems consumes the exact same accessors).
func buildValueMenuItems() []valueMenuItem {
	var items []valueMenuItem
	for _, v := range data.StatusValues() {
		items = append(items, valueMenuItem{group: "status", value: v})
	}
	for _, v := range data.TypeValues() {
		items = append(items, valueMenuItem{group: "type", value: v})
	}
	for _, v := range data.PriorityValues() {
		items = append(items, valueMenuItem{group: "priority", value: v})
	}
	return items
}

// valueMenuCursorFor finds group/value's index in items -- openValueMenu
// uses this to seed the menu's initial cursor on the focused bean's CURRENT
// status (port beans-src statuspicker.go newStatusPickerModel: selectedIndex
// seeded from the current value, statuspicker.go:85-139). Falls back to 0
// for an unrecognized value (never panics on stale/foreign data).
func valueMenuCursorFor(items []valueMenuItem, group, value string) int {
	for i, it := range items {
		if it.group == group && it.value == value {
			return i
		}
	}
	return 0
}

// valueMenuIsCurrent reports whether it is b's current value in its own
// group -- drives the "(current)" marker (valueMenuBox). Priority mirrors
// data.PriorityRank's own empty-defaults-to-normal convention.
func valueMenuIsCurrent(b *data.Bean, it valueMenuItem) bool {
	switch it.group {
	case "status":
		return b.Status == it.value
	case "type":
		return b.Type == it.value
	case "priority":
		p := b.Priority
		if p == "" {
			p = "normal"
		}
		return p == it.value
	}
	return false
}

// openValueMenu opens the combined value menu on the focused bean (the
// focusedBean()!=nil guard lives in the caller, keyNodeAction). mutTarget
// captures WHICH bean the open overlay acts on; the ETag is deliberately NOT
// captured here (design decision d, see beanETag in update.go) -- only the
// ID, so a watch-reload between open and submit is automatically honored.
func (m model) openValueMenu() model {
	b := m.focusedBean()
	if b == nil {
		return m
	}
	m.mutTarget = b.ID
	m.menuItems = buildValueMenuItems()
	m.menu = listState{}
	m.menu.setLen(len(m.menuItems))
	m.menu.cursor = valueMenuCursorFor(m.menuItems, "status", b.Status)
	m.overlay = overlayValueMenu
	return m
}

// keyValueMenu drives the open value menu: up/down move the cursor
// (navKey), enter applies the cursored value and closes (immediate-apply
// Enter semantics, design decision a3), esc/s close WITHOUT mutating
// anything (mirrors keyFilterMenu's esc/enter/f close-without-clearing
// precedent, box_filter_facets.go).
func (m model) keyValueMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch navKey(msg.String()) {
	case "up":
		m.menu.move(-1)
		return m, nil
	case "down":
		m.menu.move(1)
		return m, nil
	}
	switch {
	case keybind.Matches(msg, keys.Back), keybind.Matches(msg, keys.Status):
		m.overlay = overlayNone
		return m, nil
	case keybind.Matches(msg, keys.Enter):
		return m.applyValueMenuSelection()
	}
	return m, nil
}

// applyValueMenuSelection dispatches the cursored row's Set* mutation
// (SetStatus/SetType/SetPriority, keyed by group) against a FRESH etag read
// from the live index (m.beanETag, update.go, design decision d -- never a
// captured copy). ok==false means mutTarget vanished from the index between
// open and this enter (e.g. a watch-reload after an external delete) -- the
// overlay closes with a status-line note instead of firing a doomed
// mutation; no reload Cmd is fired either, since the vanished-bean fact was
// itself already learned from the CURRENT (already reloaded) index.
func (m model) applyValueMenuSelection() (tea.Model, tea.Cmd) {
	m.overlay = overlayNone
	if m.menu.cursor < 0 || m.menu.cursor >= len(m.menuItems) {
		return m, nil
	}
	it := m.menuItems[m.menu.cursor]
	id := m.mutTarget
	etag, ok := m.beanETag(id)
	if !ok {
		m.err = "Bean nicht mehr vorhanden — Auswahl verworfen"
		return m, nil
	}
	client := m.client
	switch it.group {
	case "status":
		return m, mutateCmd(func() error { return client.SetStatus(id, it.value, etag) })
	case "type":
		return m, mutateCmd(func() error { return client.SetType(id, it.value, etag) })
	case "priority":
		return m, mutateCmd(func() error { return client.SetPriority(id, it.value, etag) })
	}
	return m, nil
}

// valueMenuBox renders the floating combined value menu -- same modalPanel +
// facetHead-group-header render pattern as box_filter_facets.go's
// treeFilterBox, "(current)" suffix (theme.Muted) on each group's current
// value (port beans-src statuspicker.go isCurrent).
func (m model) valueMenuBox() string {
	var b strings.Builder
	b.WriteString(theme.Muted.Render("enter:apply  esc/s:cancel") + "\n")

	var cur *data.Bean
	if m.idx != nil {
		cur = m.idx.ByID[m.mutTarget]
	}

	lastGroup := ""
	for i, it := range m.menuItems {
		if it.group != lastGroup {
			b.WriteString("\n" + theme.Dim.Render(valueMenuGroupHead[it.group]) + "\n")
			lastGroup = it.group
		}
		cursor := "  "
		label := it.value
		if i == m.menu.cursor {
			cursor = theme.Accent.Render("▸ ")
			label = theme.Header.Render(label)
		}
		if cur != nil && valueMenuIsCurrent(cur, it) {
			label += " " + theme.Muted.Render("(current)")
		}
		b.WriteString(cursor + label + "\n")
	}
	return modalPanel("Set value", b.String(), "", clampModalWidth(40, m.width), theme.Mauve)
}
