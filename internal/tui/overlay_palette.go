package tui

// overlay_palette.go — the Command-Center (`ctrl+k`/`K`, E4 Task 1, bean
// bt-jpgn, design-spec.md §6 V5): a floating fuzzy-filtered action palette,
// openable from ANY view. T1 shipped the action half (paletteActions); E4
// Task 2 (bean bt-yo60) had added a second candidate pool -- matching beans
// mixed in BELOW the actions, plus a palette-scoped Bleve staleness guard --
// removed again by B13 (design-spec.md §15 PF-16/"US-04-Revision", bean
// bt-ntoz, E8 Task 7, bean bt-yqdy): the Command-Center shows ONLY commands
// now, bean search is exclusively `/`'s job (design-spec §6 V2/V3). Task 4
// (bean bt-yy6w) had appended a "go to: review cockpit" jump action (devd
// overlay_palette.go's own "Reviews (T17)/Memory (T18) werden hier ergänzt"
// precedent) -- removed again by E7 T1 (PF-14, bean bt-wmtb): the
// Review-Cockpit view no longer exists.
//
// Port references: fuzzyMatch is a verbatim port (fuzzy.go, design decision
// a). paletteActions' "verb entity" wording convention (E7 T3, PF-8 --
// formerly "verb: label" with a colon) and dispatchPalette's
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

// paletteItemKind distinguishes Command-Center row kinds. B13 (design-spec.md
// §15 PF-16, bean bt-ntoz, E8 Task 7) removed the palette's former second
// pool (paletteKindBean) -- paletteKindAction is the only value left, kept as
// a typed field (rather than collapsed away) so a future genuinely NEW row
// kind is a pure addition again, not a signature change (same precedent this
// type originally documented for T2).
type paletteItemKind int

const (
	paletteKindAction paletteItemKind = iota
)

// paletteItem is one row of the Command-Center's filtered action list.
type paletteItem struct {
	kind     paletteItemKind
	actionID string
	label    string // pre-rendered row text
}

// paletteActions returns the context-aware action list (design decision b):
// focused-bean node actions FIRST (only when m.focusedBean() != nil), global
// actions after.
func paletteActions(m model) []paletteItem {
	var items []paletteItem
	if b := m.focusedBean(); b != nil {
		items = append(items,
			paletteItem{kind: paletteKindAction, actionID: "status", label: "set status"},
			// B12 (design-spec.md §15 PF-16, bean bt-ntoz, E8 Task 6): the
			// combined value menu split into three single-group overlays --
			// "set type"/"set priority" are the Palette's new second entry
			// points to the SAME m.openValueMenu(group) handler `s` and the
			// PF-5 Meta field-level enter cascade already use. This is NOT a
			// new keybinding (design-spec §7/decision a3 still reserves
			// exactly ONE key, `s`, for the whole cluster -- explicit
			// clarification per the bean's own ERRATUM note, in case a future
			// reader is tempted to add a dedicated Type/Priority key).
			paletteItem{kind: paletteKindAction, actionID: "type", label: "set type"},
			paletteItem{kind: paletteKindAction, actionID: "priority", label: "set priority"},
			paletteItem{kind: paletteKindAction, actionID: "tags", label: "set tags"},
			// B14 (design-spec.md §15 PF-16, bean bt-ntoz, E8 Task 7, bean
			// bt-yqdy): "create tag" is a Palette-only entry point to the SAME
			// openTagPicker().openTagInput() new-tag sub-mode the Tag-Picker's
			// own `n` key already opens (box_picker_tag.go) -- grouped
			// directly after "set tags" (both tag-related). Node action, since
			// its dispatchPalette handler needs a focused bean's ID as the
			// mutation target -- see that handler's own guard doc-stamp.
			paletteItem{kind: paletteKindAction, actionID: "create_tag", label: "create tag"},
			paletteItem{kind: paletteKindAction, actionID: "parent", label: "set parent"},
			paletteItem{kind: paletteKindAction, actionID: "blocking", label: "set blocking"},
			paletteItem{kind: paletteKindAction, actionID: "edit_title", label: "set title"},
			paletteItem{kind: paletteKindAction, actionID: "delete", label: "delete bean"},
		)
	}
	items = append(items,
		paletteItem{kind: paletteKindAction, actionID: "create", label: "create bean"},
		paletteItem{kind: paletteKindAction, actionID: "go_backlog", label: "go to backlog"},
		paletteItem{kind: paletteKindAction, actionID: "go_browse", label: "go to browse"},
		paletteItem{kind: paletteKindAction, actionID: "filter", label: "filter facets"},
		paletteItem{kind: paletteKindAction, actionID: "search", label: "search beans"},
		paletteItem{kind: paletteKindAction, actionID: "reload", label: "reload data"},
		// E5 Task 6 (bean bt-zhwl): a second entry point to the SAME
		// keys.Picker dispatch (openLobby, view_lobby.go) -- mirrors every
		// other single-key binding's own Command-Center mirror above.
		paletteItem{kind: paletteKindAction, actionID: "repo_picker", label: "go to repo picker"},
		// E10 Task 2 (bean bt-r92i, epic bt-362n D05): the Tag-Management
		// page has NO dedicated keybinding either (D05 mirrors the "go to
		// settings" precedent immediately below -- Tastenraum bleibt knapp) --
		// grouped directly BEFORE "settings" as the new last-but-one entry
		// (Planner-Entscheidung, mirrors how "repo_picker" itself already
		// sits directly before "settings").
		paletteItem{kind: paletteKindAction, actionID: "go_tags", label: "go to tags"},
		// E5 Task 5 (bean bt-0l8c): the Settings-Form has NO dedicated
		// keybinding (design-spec §7 knows none) -- reachable exclusively
		// through the Command-Center, appended last.
		paletteItem{kind: paletteKindAction, actionID: "settings", label: "go to settings"},
	)
	return items
}

// palFiltered fuzzy-filters the action pool against m.palQuery. B13
// (design-spec.md §15 PF-16/"US-04-Revision", bean bt-ntoz, E8 Task 7, bean
// bt-yqdy) removed the palette's former second pool (matching beans mixed in
// below the actions, palFilteredBeans) -- the Command-Center shows ONLY
// commands now, bean search is exclusively `/`'s job.
func (m model) palFiltered() []paletteItem {
	var out []paletteItem
	for _, it := range paletteActions(m) {
		if fuzzyMatch(m.palQuery, it.label) {
			out = append(out, it)
		}
	}
	return out
}

// openPalette opens the Command-Center with an empty filter (design decision
// h: reachable from ANY view via keys.Palette).
func (m model) openPalette() (tea.Model, tea.Cmd) {
	m.paletteOpen = true
	m.palQuery = ""
	m.palList = listState{}
	m.palList.setLen(len(m.palFiltered()))
	return m, nil
}

// keyPalette drives the open Command-Center: every rune/backspace edits
// palQuery, rebuilding palList's length every keystroke; up/down move the
// cursor; enter dispatches the cursored item; esc closes without side
// effects. B13 (design-spec.md §15 PF-16/"US-04-Revision", bean bt-ntoz,
// E8 Task 7, bean bt-yqdy) removed the rune/backspace branches' former
// palette-scoped Bleve dispatch tail (dispatchPaletteBleveIfDue) -- there is
// no longer a bean-search half to keep fresh, so both branches now just
// resync palList and return.
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
// (status -> m.openValueMenu("status"), etc.) so the Palette is a genuine second
// entry point to the SAME handlers, never a parallel implementation (US-04's
// "jede Aktion über die Command-Palette erreichbar").
func (m model) dispatchPalette(it paletteItem) (tea.Model, tea.Cmd) {
	m.paletteOpen = false
	switch it.kind {
	case paletteKindAction:
		switch it.actionID {
		case "status":
			return m.openValueMenu("status"), nil
		case "type":
			return m.openValueMenu("type"), nil
		case "priority":
			return m.openValueMenu("priority"), nil
		case "tags":
			return m.openTagPicker(), nil
		case "create_tag":
			// B14 (design-spec.md §15 PF-16, bean bt-ntoz, E8 Task 7, bean
			// bt-yqdy): opens the Tag-Picker AND its free-text new-tag
			// sub-mode in one step -- m.openTagPicker() (box_picker_tag.go)
			// is ITSELF already a no-op-safe method (returns m unchanged,
			// overlay stays overlayNone, when focusedBean()==nil), but
			// chaining .openTagInput() straight onto that unconditionally
			// sets m.tagInputActive=true regardless -- a latent, unreachable
			// state (tagInputActive true while overlay==overlayNone, no
			// picker actually open). This guard is therefore MANDATORY here
			// (unlike the plain "tags" case above, which safely no-ops on
			// its own via openTagPicker's internal guard): it runs BEFORE
			// the openTagPicker().openTagInput() chain, not relying on the
			// chain's own internal no-op to cover the .openTagInput() half.
			if m.focusedBean() == nil {
				return m, nil
			}
			return m.openTagPicker().openTagInput()
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
		case "repo_picker":
			// E5 Task 6 (bean bt-zhwl): identical to keys.Picker's own
			// dispatch (handleKey, update.go) -- the Command-Center is a
			// genuine second entry point to the SAME handler, never a
			// parallel implementation (dispatchPalette's own doc-stamp).
			return m.openLobby()
		case "go_tags":
			// E10 Task 2 (bean bt-r92i, epic bt-362n D05): the Tag-Management
			// page's ONLY entry point (no dedicated keybinding, same
			// "Command-Center only" shape as "settings" just below).
			return m.openTagManagementPage()
		case "settings":
			// E5 Task 5 (bean bt-0l8c): opens the Settings-Form
			// (box_form_settings.go) -- same open-form shape as edit_title,
			// no mutTarget (settings are repo-independent).
			return m.openSettingsForm()
		}
	}
	return m, nil
}

// paletteBox renders the floating Command-Center -- actions render PLAIN +
// theme.Header on select (mirrors box_menu_value.go/box_picker_tag.go: their
// row text carries no per-cell theming of its own). B13 (design-spec.md §15
// PF-16/"US-04-Revision", bean bt-ntoz, E8 Task 7, bean bt-yqdy) removed the
// former SECOND menuList call + separator this doc comment used to describe
// (E4 Task 2's bean-result rows, split/beanItems) -- ONE pool, ONE menuList
// call, no split needed anymore.
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
