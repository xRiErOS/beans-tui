package tui

// overlay_palette.go — the Command-Center (`ctrl+k`/`K`, E4 Task 1, bean
// bt-jpgn, design-spec.md §6 V5): a floating fuzzy-filtered action palette,
// openable from ANY view. T1 shipped the action half (paletteActions); Task
// 2 (bean bt-yo60) adds the second candidate pool -- matching beans mixed in
// BELOW the actions (palFilteredBeans, design decision b), palette-scoped
// Bleve staleness guard (palBleveIDs/palBleveFor/palBleveLoading, types.go;
// paletteBleveResultMsg/paletteSearchCmd, messages.go). Task 4 (bean
// bt-yy6w) appends the "go to: review cockpit" jump action now that the
// Review-Cockpit view exists (devd overlay_palette.go's own "Reviews
// (T17)/Memory (T18) werden hier ergänzt" precedent, port reference at the
// top of epic-E4-plan.md).
//
// Port references: fuzzyMatch is a verbatim port (fuzzy.go, design decision
// a). paletteActions' "verb: label" wording convention and dispatchPalette's
// switch-over-an-id-string shape port devd overlay_palette.go:20-49/144-190
// STRUCTURALLY only -- the concrete action list and every handler body are
// beans-tui-native (design decision b). paletteBox's "> query" + separator +
// menuList render mirrors devd overlay_palette.go:192-207 1:1 (modalPanel
// standing in for devd's inline box chrome); T2's bean rows additionally
// port box_picker_parent.go's D08 ansi.Strip+Accent-wrap convention (see
// paletteBox's own doc comment below).

import (
	"strings"

	"beans-tui/internal/data"
	"beans-tui/internal/theme"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
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
	actionID string     // kind == paletteKindAction
	bean     *data.Bean // kind == paletteKindBean (T2)
	label    string     // pre-rendered row text for both kinds
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
		paletteItem{kind: paletteKindAction, actionID: "go_review", label: "go to: review cockpit"},
		paletteItem{kind: paletteKindAction, actionID: "filter", label: "filter: facetten"},
		paletteItem{kind: paletteKindAction, actionID: "search", label: "search: beans"},
		paletteItem{kind: paletteKindAction, actionID: "reload", label: "reload: daten"},
	)
	return items
}

// palFiltered combines both candidate pools -- actions (T1) FIRST, then the
// bean half (T2, palFilteredBeans) -- filtered against m.palQuery
// (design decision b's ordering: actions always precede beans, no
// score-based interleaving).
func (m model) palFiltered() []paletteItem {
	var out []paletteItem
	for _, it := range paletteActions(m) {
		if fuzzyMatch(m.palQuery, it.label) {
			out = append(out, it)
		}
	}
	out = append(out, m.palFilteredBeans()...)
	return out
}

// paletteBeanResultCap caps palFilteredBeans' result set (design decision b:
// prevents a broad query like "e" from flooding the modal).
const paletteBeanResultCap = 20

// palBeanMatches mirrors beanMatchesSearch's bifurcation (view_browse_repo.go)
// against the PALETTE's own Bleve fields instead of the Tree/Backlog's, so
// opening ctrl+k never touches an active `/` search session (design decision
// b). Below the Bleve threshold (<3 chars), or while a palette-Bleve response
// for THIS exact query hasn't arrived yet, falls back to an immediate local
// title+ID substring match; once palBleveFor catches up to m.palQuery, the
// Bleve result set becomes authoritative -- UNIONed with the local
// ID-substring match (I01 precedent, beanMatchesSearch's own doc comment):
// data.Client.Search indexes title+body, not necessarily an arbitrary ID
// substring.
func (m model) palBeanMatches(b *data.Bean) bool {
	q := strings.ToLower(strings.TrimSpace(m.palQuery))
	if len(q) >= 3 && m.palBleveFor == m.palQuery {
		return m.palBleveIDs[b.ID] || strings.Contains(strings.ToLower(b.ID), q)
	}
	return strings.Contains(strings.ToLower(b.Title), q) || strings.Contains(strings.ToLower(b.ID), q)
}

// palFilteredBeans returns up to paletteBeanResultCap matching beans,
// canonically sorted (data.SortBeans, I03) -- an empty (or all-whitespace)
// query returns nil (design decision b: the palette never dumps the whole
// repo the way `/`'s already-visible tree filter can).
func (m model) palFilteredBeans() []paletteItem {
	if strings.TrimSpace(m.palQuery) == "" || m.idx == nil {
		return nil
	}
	var matched []*data.Bean
	for _, b := range m.idx.ByID {
		if m.palBeanMatches(b) {
			matched = append(matched, b)
		}
	}
	data.SortBeans(matched)
	if len(matched) > paletteBeanResultCap {
		matched = matched[:paletteBeanResultCap]
	}
	items := make([]paletteItem, len(matched))
	for i, b := range matched {
		items[i] = paletteItem{kind: paletteKindBean, bean: b, label: relationRow(b)}
	}
	return items
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
// palQuery (rebuilding palList's length every keystroke, additionally
// dispatching a palette-scoped Bleve search via maybePaletteBleveCmd when
// due, T2); up/down move the cursor; enter dispatches the cursored item;
// esc closes without side effects.
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
		return m.dispatchPaletteBleveIfDue()
	case tea.KeyRunes, tea.KeySpace:
		m.palQuery += string(msg.Runes)
		m.palList.setLen(len(m.palFiltered()))
		return m.dispatchPaletteBleveIfDue()
	}
	return m, nil
}

// dispatchPaletteBleveIfDue is keyPalette's shared Bleve-dispatch tail (E4
// Task 2, bean bt-yo60) -- unlike update.go's dispatchBleveIfDue, there is no
// extra Cmd to batch (the Palette manages palQuery as a plain string, no
// bubbles textinput.Model backing it), so this simply fires
// maybePaletteBleveCmd() when due, flagging palBleveLoading.
func (m model) dispatchPaletteBleveIfDue() (tea.Model, tea.Cmd) {
	cmd := m.maybePaletteBleveCmd()
	if cmd == nil {
		return m, nil
	}
	m.palBleveLoading = true
	return m, cmd
}

// dispatchPalette closes the palette and routes the selected item to the
// matching handler -- action IDs mirror the existing single-key dispatch 1:1
// (status -> m.openValueMenu(), etc.) so the Palette is a genuine second
// entry point to the SAME handlers, never a parallel implementation (US-04's
// "jede Aktion über die Command-Palette erreichbar").
func (m model) dispatchPalette(it paletteItem) (tea.Model, tea.Cmd) {
	m.paletteOpen = false
	switch it.kind {
	case paletteKindBean:
		// E4 Task 2 (bean bt-yo60): jump the tree cursor onto the matched
		// bean and switch to Browse (even from Backlog) -- expandAncestorsOf
		// (cycle-guarded, update.go) marks every ancestor expanded so the
		// jump target is guaranteed visible in the very next
		// visibleNodes() call, same call shape as keyDetailFocus's own
		// relation-jump (update.go).
		m.expanded = expandAncestorsOf(m.idx, m.expanded, it.bean.ID)
		m.cursorID = it.bean.ID
		m.view = viewBrowseRepo
		// B01 (E4 Task 2 review, bean bt-yo60): a bean-jump must leave
		// Detail-Focus, exactly like keyDetailFocus's own relation-jump
		// (update.go:702) -- otherwise arrow keys on the NEW bean still
		// drive the OLD bean's accordion instead of the tree. The
		// Detail-Accordion focus machine ints get the SAME reset
		// tab-into-detail-focus uses (handleKey, types.go's "All four
		// reset on every tab-into-detail-focus transition" doc-stamp),
		// not just a defensive clamp -- a stale secCursor/fieldCursor
		// pointing past the NEW bean's section/field shape must never
		// leak into the next detail-focus visit.
		m.detailFocus = false
		m.secCursor, m.accOpen, m.detailLevel, m.fieldCursor = 0, 1, 0, 0
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
		case "go_review":
			// E4 Task 4 (bean bt-yy6w): reuses openReviewCockpit verbatim
			// (view_review_cockpit.go) -- same reviewCursor/reviewAccOpen
			// reset `R` already performs, so a Palette-driven jump into the
			// Cockpit behaves identically to the direct keybinding.
			return m.openReviewCockpit()
		case "filter":
			return m.openFilterMenu()
		case "search":
			return m.openSearchInput()
		case "reload":
			return m, loadCmd(m.client)
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
// overlays, extended here to a THIRD file for the same reason. The two pools
// render as two SEPARATE menuList calls (not one over the combined slice) so
// each can carry its own row style while still sharing ONE cursor index
// space (m.palList.cursor over m.palFiltered()'s combined order, actions
// first) -- a theme.Dim separator is inserted between them only when BOTH
// pools are non-empty (T2).
func (m model) paletteBox() string {
	items := m.palFiltered()
	var b strings.Builder
	b.WriteString(theme.Accent.Render("> ") + m.palQuery + "▏\n")
	b.WriteString(theme.Dim.Render(strings.Repeat("─", 44)) + "\n")
	if len(items) == 0 {
		b.WriteString(theme.Dim.Render("(no matches)") + "\n")
	}

	split := len(items) // index of the first paletteKindBean item, if any
	for i, it := range items {
		if it.kind == paletteKindBean {
			split = i
			break
		}
	}
	actionItems, beanItems := items[:split], items[split:]

	b.WriteString(menuList(len(actionItems), m.palList.cursor, func(i int, sel bool) string {
		label := actionItems[i].label
		if sel {
			label = theme.Header.Render(label)
		}
		return label
	}))

	if len(actionItems) > 0 && len(beanItems) > 0 {
		b.WriteString(theme.Dim.Render(strings.Repeat("─", 44)) + "\n")
	}

	if len(beanItems) > 0 {
		b.WriteString(menuList(len(beanItems), m.palList.cursor-split, func(i int, sel bool) string {
			label := beanItems[i].label
			if sel {
				return theme.Accent.Render(ansi.Strip(label))
			}
			return label
		}))
	}

	return modalPanel("Command-Center", b.String(), "type: filter   ↑↓: select   enter: run   esc: close", clampModalWidth(48, m.width), theme.Mauve)
}
