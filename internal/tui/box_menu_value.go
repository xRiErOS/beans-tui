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

// buildValueMenuItems assembles ONE group's 5 rows (B11/B12, design-spec.md
// §15 PF-16, bean bt-ntoz): the menu USED to always return all 15 rows (5
// status + 5 type + 5 priority) regardless of which group the caller asked
// for -- confusing both for the `s` key (Status-only by design decision a3)
// and for the PF-5 Meta field-level enter cascade (which reaches this menu
// already correctly group-scoped via f.kind, but the combined render then
// showed the other two groups anyway). group selects exactly one of
// data.StatusValues/TypeValues/PriorityValues (design decision b -- single
// source; box_filter_facets.go's buildFilterItems consumes the exact same
// accessors); an unrecognized group returns nil (openValueMenu's caller
// contract only ever passes "status"/"type"/"priority", mirrors
// currentValueForGroup's own unrecognized-group fallback).
func buildValueMenuItems(group string) []valueMenuItem {
	var items []valueMenuItem
	switch group {
	case "status":
		for _, v := range data.StatusValues() {
			items = append(items, valueMenuItem{group: "status", value: v})
		}
	case "type":
		for _, v := range data.TypeValues() {
			items = append(items, valueMenuItem{group: "type", value: v})
		}
	case "priority":
		for _, v := range data.PriorityValues() {
			items = append(items, valueMenuItem{group: "priority", value: v})
		}
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

// currentValueForGroup returns b's current value in the given group
// (status/type/priority) -- the same empty-defaults-to-normal convention as
// valueMenuIsCurrent. Used by openValueMenu(group) to seed the cursor on
// whichever group the caller opens (T6, design-spec.md §15 PF-5: the Meta
// field-level enter cascade seeds "type"/"priority" too, not only the
// `s`-key's hardcoded "status"). Returns "" for an unrecognized group (falls
// back to valueMenuCursorFor's own index-0 default, never panics).
func currentValueForGroup(b *data.Bean, group string) string {
	switch group {
	case "status":
		return b.Status
	case "type":
		return b.Type
	case "priority":
		p := b.Priority
		if p == "" {
			p = "normal"
		}
		return p
	}
	return ""
}

// openValueMenu opens the value menu on the focused bean, filtered to and
// seeded on group's current value (the focusedBean()!=nil guard lives in the
// caller, keyNodeAction/keyDetailFocus/dispatchPalette). group is "status"/
// "type"/"priority" (T6, design-spec.md §15 PF-5 -- was hardcoded to "status"
// before the Meta field-level enter cascade needed to seed the other two
// groups as well). B11/B12 (E8 Task 6, bean bt-ntoz): group now ALSO filters
// buildValueMenuItems, not just the cursor seed -- the menu shows ONLY the
// requested group's 5 rows. mutTarget captures WHICH bean the open overlay
// acts on; the ETag is deliberately NOT captured here (design decision d, see
// beanETag in update.go) -- only the ID, so a watch-reload between open and
// submit is automatically honored.
func (m model) openValueMenu(group string) model {
	b := m.focusedBean()
	if b == nil {
		return m
	}
	m.mutTarget = b.ID
	m.menuItems = buildValueMenuItems(group)
	m.menu = listState{}
	m.menu.setLen(len(m.menuItems))
	m.menu.cursor = valueMenuCursorFor(m.menuItems, group, currentValueForGroup(b, group))
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
		m.err = "Bean no longer exists — selection discarded"
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

// valueMenuTitle picks modalPanel's title from the open menu's (now single,
// B11/B12) group -- "Set status"/"Set type"/"Set priority" instead of the old
// hard-coded "Set value", so it's unambiguous WHICH group is currently open
// (Planner addition beyond the bare B12 wording, bean bt-y2iw: "welches Menü
// ist gerade offen" must not become the new source of confusion once the
// combined 15-row menu splits into three single-group ones). Falls back to
// "Set value" for an empty menuItems slice (defensive -- openValueMenu never
// produces one for a recognized group, but valueMenuBox must not panic on a
// zero-value model either, e.g. in a render-only test).
func valueMenuTitle(items []valueMenuItem) string {
	if len(items) == 0 {
		return "Set value"
	}
	switch items[0].group {
	case "status":
		return "Set status"
	case "type":
		return "Set type"
	case "priority":
		return "Set priority"
	}
	return "Set value"
}

// valueMenuBox renders the floating (now single-group, B11/B12) value menu --
// same modalPanel + facetHead-group-header render pattern as
// box_filter_facets.go's treeFilterBox, "(current)" suffix (theme.Muted) on
// the group's current value (port beans-src statuspicker.go isCurrent). The
// per-row group-header line still renders (lastGroup bookkeeping) even though
// exactly one group is present now -- harmless, keeps the render loop
// unchanged/simple, and the header text still adds value confirming which
// group is open.
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
	return modalPanel(valueMenuTitle(m.menuItems), b.String(), "", clampModalWidth(40, m.width), theme.Mauve)
}
