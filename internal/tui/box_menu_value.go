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

	"github.com/xRiErOS/beans-tui/internal/data"
	"github.com/xRiErOS/beans-tui/internal/theme"
	keybind "github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
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
	// Slice C (bt-f0y9, D09 revidiert): seed the anchor field HERE, the ONE
	// place all four trigger paths (keyboard s/o/u, field-cursor Enter, mouse
	// click, Palette) funnel through (S6 grounding §4) -- so this is the only
	// call site that needs touching, not four. Read by composeOverlays'
	// placeValueMenuOverlay (view_browse_repo.go) ONLY while boxFormEnabled().
	m.valueMenuAnchorField = boxFormFieldIndexForGroup(group)
	m.overlay = overlayValueMenu
	return m
}

// boxFormFieldIndexForGroup maps a value-menu group ("status"/"type"/
// "priority") to its boxFormFieldOrder (box_nav_field.go) index -- a STATIC
// lookup (the group<->field position never varies per-bean or per-render,
// unlike boxFormEffectiveCursor's per-session state), used by openValueMenu
// (above) to seed m.valueMenuAnchorField regardless of which of the four
// trigger paths opened the menu. -1 for an unrecognized group (defensive
// only -- openValueMenu's own callers never pass one); placeValueMenuOverlay
// falls back to the pre-existing centered placeOverlay whenever
// boxFormFieldRect can't resolve the resulting index (same as a nil focused
// bean).
func boxFormFieldIndexForGroup(group string) int {
	var target string
	switch group {
	case "status":
		target = boxFormTargetStatus
	case "type":
		target = boxFormTargetType
	case "priority":
		target = boxFormTargetPriority
	default:
		return -1
	}
	for i, f := range boxFormFieldOrder {
		if f.target == target {
			return i
		}
	}
	return -1
}

// keyValueMenu drives the open value menu: up/down move the cursor
// (navKey), enter applies the cursored value and closes (immediate-apply
// Enter semantics, design decision a3), esc and the group's OWN key close
// WITHOUT mutating anything (mirrors keyFilterMenu's esc/enter/f
// close-without-clearing precedent, box_filter_facets.go).
//
// a3-NACHTRAG (bean bt-z4w7, B7): decision a3 was written as "esc/s
// schliesst" back when `s` was the only key that could open this menu. S4
// added o=Type and u=Priority, and the literal `s` became wrong twice over
// -- an o-opened menu answered to a key its own footer had no business
// showing. The close alias is now the key that OPENED the menu, resolved
// through valueMenuGroupKey (footer_context.go) -- the SAME accessor the
// footer and the inline hint read, so the three can no longer disagree.
// This deliberately revises a3 rather than quietly patching a label.
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
	case keybind.Matches(msg, keys.Back), keybind.Matches(msg, valueMenuGroupKey(m.valueMenuGroup())):
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
		// bt-81f0: see box_confirm_delete.go's identical guard comment.
		var toastCmd tea.Cmd
		m, toastCmd = m.showToast(toastError, m.err, "", nil, false)
		return m, toastCmd
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
	// bean bt-z4w7: the cancel key is READ OFF the binding keyValueMenu
	// actually matches (valueMenuGroupKey, footer_context.go) instead of the
	// hardcoded "s" this line used to carry -- a Priority menu opened with
	// `u` said "esc/s:cancel" while the outer Footer Zone 3 said the same
	// wrong thing, both describing a group they were not showing.
	b.WriteString(theme.Muted.Render("enter:apply  esc/"+valueMenuGroupKey(m.valueMenuGroup()).Help().Key+":cancel") + "\n")

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

// valueMenuPopupY picks the value-menu popup's Y position (Slice C, bt-f0y9
// "feld-verankertes Inline-Dropdown", D09 revidiert) given its triggering
// field's own rect (anchorY, anchorH -- boxFormFieldRect, box_nav_field.go)
// and the popup's own height (fgH) against the canvas height (canvasH):
// conventional dropdown behavior, BELOW the field by default (PO-Vorgabe).
//
// If below would run past the bottom of the canvas, it flips ABOVE the
// field instead -- as long as the field itself isn't already sitting at
// canvas row 0 (anchorY == 0, "kein Platz" at all to flip into). The flipped
// position is NOT required to fully fit above (Status/Type/Priority sit only
// ~10 rows down in every real render, see this slice's own ERRATUM in
// box_nav_field.go -- a popup taller than that can never fully fit above
// them, so requiring a full fit would make the flip branch permanently dead
// code for this feature's own fields): a partially-clipped-at-the-top popup
// still beats one silently staying centered or bottom-clipped, and
// placeOverlayAt's own silent-overflow-drop (Slice B, placeCompose) already
// handles the clip the exact same way it handles every other overlay in
// this codebase -- no new overflow mechanism introduced here.
//
// anchorY == 0 is the ONLY case that skips the flip (PO-Wortlaut "nur wenn
// oben auch kein Platz: clampen") -- falls through to "below", clipped at
// the bottom like any other overflow.
func valueMenuPopupY(anchorY, anchorH, fgH, canvasH int) int {
	below := anchorY + anchorH
	if below+fgH <= canvasH {
		return below
	}
	if anchorY > 0 {
		return anchorY - fgH
	}
	return below
}

// placeValueMenuOverlay composes the value-menu overlay onto out (Slice C):
// while boxFormEnabled() AND the anchor field's rect resolves
// (boxFormFieldRect, box_nav_field.go, seeded via m.mutTarget/
// m.valueMenuAnchorField at open time, openValueMenu above), the menu is
// placed directly BELOW its triggering boxed field (or flipped above it,
// valueMenuPopupY) via placeOverlayAt (Slice B) instead of the pre-existing
// CENTERED placeOverlay. Falls back to placeOverlay in every other case:
// boxFormEnabled() off (accordion mode has no boxed fields to anchor to at
// all -- D09's own scope, "nur die endlichen Einzelfeld-Menüs"), or a rect
// that fails to resolve (nil mutTarget bean, out-of-range anchor field --
// defensive; should not happen in practice since openValueMenu always seeds
// m.valueMenuAnchorField from the static boxFormFieldIndexForGroup lookup in
// the SAME call that sets m.overlay).
//
// x is clamped so the popup never renders past the RIGHT edge of the canvas
// (defensive only -- the popup's own width, clampModalWidth(40, m.width), is
// unrelated to the ~15-cell field it anchors to and can run wider than the
// remaining space toward the right edge of an 80-column pane; this clamp
// keeps the render on-canvas, it does NOT resize/reflow the popup to the
// field's own width -- that "should the popup ever shrink or reflow
// horizontally" question is explicitly OUT of this slice's scope, see
// bt-f0y9's own S6 grounding §3 and this slice's "## Notes for Reviewer").
func (m model) placeValueMenuOverlay(out string, w, h int) string {
	fg := m.valueMenuBox()
	if !boxFormEnabled() {
		return placeOverlay(out, fg, w, h)
	}
	b := m.idx.ByID[m.mutTarget] // nil-safe: a nil map / missing key both yield nil
	rx, ry, _, rh, ok := boxFormFieldRect(m, b, m.valueMenuAnchorField)
	if !ok {
		return placeOverlay(out, fg, w, h)
	}

	fgLines := strings.Split(fg, "\n")
	fgW, fgH := 0, len(fgLines)
	for _, l := range fgLines {
		if lw := ansi.StringWidth(l); lw > fgW {
			fgW = lw
		}
	}

	x := rx
	if fgW < w && x+fgW > w {
		x = w - fgW
	}
	if x < 0 {
		x = 0
	}
	y := valueMenuPopupY(ry, rh, fgH, h)

	return placeOverlayAt(out, fg, w, h, x, y)
}
