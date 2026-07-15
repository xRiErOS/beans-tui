package tui

// overlay_palette.go — the Command-Center (`ctrl+k`/`K`, E4 Task 1, bean
// bt-jpgn, design-spec.md §6 V5): a floating fuzzy-filtered action palette,
// openable from ANY view. T1 ships the action half only (design decision
// b) -- Task 2 (bean bt-yo60) mixes matching beans in below the actions
// (palFilteredBeans), Task 4 (bean bt-yy6w) appends the Review-Cockpit jump
// action once that view exists (devd overlay_palette.go's own "Reviews
// (T17)/Memory (T18) werden hier ergänzt" precedent, port reference at the
// top of epic-E4-plan.md).
//
// Port references: fuzzyMatch is a verbatim port (fuzzy.go, design decision
// a). paletteActions' "verb: label" wording convention and dispatchPalette's
// switch-over-an-id-string shape port devd overlay_palette.go:20-49/144-190
// STRUCTURALLY only -- the concrete action list and every handler body are
// beans-tui-native (design decision b). paletteBox's "> query" + separator +
// menuList render mirrors devd overlay_palette.go:192-207 1:1 (modalPanel
// standing in for devd's inline box chrome).

import (
	"strings"

	"beans-tui/internal/theme"
	tea "github.com/charmbracelet/bubbletea"
)

// paletteItemKind distinguishes the Command-Center's two candidate pools
// (design decision b).
type paletteItemKind int

const (
	paletteKindAction paletteItemKind = iota
	paletteKindBean                   // T2 populates this kind; the type exists here upfront
	// (mirrors E3 Task 1's "declare the full overlayID enum upfront"
	// precedent, types.go) so T2 is a pure addition, never a signature
	// change.
)

// paletteItem is one row of the Command-Center's combined, already-filtered
// result list.
type paletteItem struct {
	kind     paletteItemKind
	actionID string // kind == paletteKindAction
	label    string // pre-rendered row text for both kinds
}

// paletteActions returns the context-aware action list (design decision b):
// focused-bean node actions FIRST (only when m.focusedBean() != nil), global
// actions after. T4 appends "go to: review cockpit" once that view exists
// (devd overlay_palette.go's own "T17/T18 ergänzen hier" precedent).
func paletteActions(m model) []paletteItem {
	var items []paletteItem
	if b := m.focusedBean(); b != nil {
		items = append(items,
			paletteItem{kind: paletteKindAction, actionID: "status", label: "status: setzen"},
			paletteItem{kind: paletteKindAction, actionID: "tags", label: "tags: zuweisen"},
			paletteItem{kind: paletteKindAction, actionID: "parent", label: "parent: zuweisen"},
			paletteItem{kind: paletteKindAction, actionID: "blocking", label: "blocking: zuweisen"},
			paletteItem{kind: paletteKindAction, actionID: "edit_title", label: "titel: bearbeiten"},
			paletteItem{kind: paletteKindAction, actionID: "delete", label: "bean: löschen"},
		)
	}
	items = append(items,
		paletteItem{kind: paletteKindAction, actionID: "create", label: "create: bean"},
		paletteItem{kind: paletteKindAction, actionID: "go_backlog", label: "go to: backlog"},
		paletteItem{kind: paletteKindAction, actionID: "go_browse", label: "go to: browse"},
		paletteItem{kind: paletteKindAction, actionID: "filter", label: "filter: facetten"},
		paletteItem{kind: paletteKindAction, actionID: "search", label: "search: beans"},
		paletteItem{kind: paletteKindAction, actionID: "reload", label: "reload: daten"},
		// T4 appends "go_review" ("go to: review cockpit") here.
	)
	return items
}

// palFiltered combines both candidate pools (T1: actions only; T2 adds the
// bean half below the actions, design decision b's ordering) filtered
// against m.palQuery.
func (m model) palFiltered() []paletteItem {
	var out []paletteItem
	for _, it := range paletteActions(m) {
		if fuzzyMatch(m.palQuery, it.label) {
			out = append(out, it)
		}
	}
	// T2 appends palFilteredBeans(m) here.
	return out
}

// openPalette opens the Command-Center with an empty filter (design decision
// h: reachable from ANY view via keys.Palette, checked in handleKey ahead of
// the E4 Task 3 Review-Cockpit capture block so it also works from inside
// the Cockpit).
func (m model) openPalette() (tea.Model, tea.Cmd) {
	m.paletteOpen = true
	m.palQuery = ""
	m.palList = listState{}
	m.palList.setLen(len(m.palFiltered()))
	return m, nil
}

// keyPalette drives the open Command-Center: every rune/backspace edits
// palQuery (rebuilding palList's length every keystroke, T2 additionally
// dispatches a Bleve search when due); up/down move the cursor; enter
// dispatches the cursored item; esc closes without side effects.
func (m model) keyPalette(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc, tea.KeyCtrlC:
		m.paletteOpen = false
		return m, nil
	case tea.KeyUp:
		m.palList.move(-1)
		return m, nil
	case tea.KeyDown:
		m.palList.move(1)
		return m, nil
	case tea.KeyEnter:
		items := m.palFiltered()
		if m.palList.cursor >= 0 && m.palList.cursor < len(items) {
			return m.dispatchPalette(items[m.palList.cursor])
		}
		return m, nil
	case tea.KeyBackspace:
		if len(m.palQuery) > 0 {
			r := []rune(m.palQuery)
			m.palQuery = string(r[:len(r)-1])
			m.palList.setLen(len(m.palFiltered()))
		}
		return m, nil
	case tea.KeyRunes, tea.KeySpace:
		m.palQuery += string(msg.Runes)
		m.palList.setLen(len(m.palFiltered()))
		return m, nil
	}
	return m, nil
}

// dispatchPalette closes the palette and routes the selected item to the
// matching handler -- action IDs mirror the existing single-key dispatch 1:1
// (status -> m.openValueMenu(), etc.) so the Palette is a genuine second
// entry point to the SAME handlers, never a parallel implementation (US-04's
// "jede Aktion über die Command-Palette erreichbar").
func (m model) dispatchPalette(it paletteItem) (tea.Model, tea.Cmd) {
	m.paletteOpen = false
	switch it.kind {
	case paletteKindBean: // T2
		return m, nil
	case paletteKindAction:
		switch it.actionID {
		case "status":
			return m.openValueMenu(), nil
		case "tags":
			return m.openTagPicker(), nil
		case "parent":
			return m.openParentPicker(), nil
		case "blocking":
			return m.openBlockingPicker(), nil
		case "edit_title":
			if b := m.focusedBean(); b != nil {
				return m.openEditTitleForm(b)
			}
		case "delete":
			return m.openDeleteConfirm(), nil
		case "create":
			// F1 (Review-Runde 2, Async-Gap-Clobbering, Finding 1b): the
			// Command-Center is a genuine second entry point to the SAME
			// handlers (dispatchPalette's doc-stamp) -- it needs its OWN copy
			// of the same single-create guard keyNodeAction's Create case
			// (update.go) and submitForm's "create" case (box_confirm_
			// create.go) already enforce, since neither of those call sites
			// runs on this path (types.go doc-stamp: THREE guarded call
			// sites now, not two).
			if m.pendingCreate != nil {
				m.err = createInFlightNote
				return m, nil
			}
			return m.openCreateForm()
		case "go_backlog":
			m.view = viewBacklog
			m.backlogList.setLen(len(m.backlogVisible()))
			return m, nil
		case "go_browse":
			m.view = viewBrowseRepo
			return m, nil
		case "filter":
			return m.openFilterMenu()
		case "search":
			return m.openSearchInput()
		case "reload":
			return m, loadCmd(m.client)
			// "go_review" -- T4
		}
	}
	return m, nil
}

// paletteBox renders the floating Command-Center -- actions render PLAIN +
// theme.Header on select (mirrors box_menu_value.go/box_picker_tag.go: their
// row text carries no per-cell theming of its own); T2's bean rows instead
// use the D08 ansi.Strip+Accent-wrap convention (mirrors box_picker_parent.go)
// since relationRow output is ALREADY themed -- same split-styling rationale
// types.go's "Picker-Stil-Divergenz" doc-stamp already documents for the E3
// overlays, extended here to a THIRD file for the same reason.
func (m model) paletteBox() string {
	items := m.palFiltered()
	var b strings.Builder
	b.WriteString(theme.Accent.Render("> ") + m.palQuery + "▏\n")
	b.WriteString(theme.Dim.Render(strings.Repeat("─", 44)) + "\n")
	if len(items) == 0 {
		b.WriteString(theme.Dim.Render("(no matches)") + "\n")
	}
	b.WriteString(menuList(len(items), m.palList.cursor, func(i int, sel bool) string {
		label := items[i].label
		if sel {
			label = theme.Header.Render(label)
		}
		return label
	}))
	return modalPanel("Command-Center", b.String(), "type: filter   ↑↓: select   enter: run   esc: close", clampModalWidth(48, m.width), theme.Mauve)
}
