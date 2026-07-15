package tui

// mouse.go — E5 Task 4 (bean bt-mne6, epic bt-5h4d, design decision f): Maus
// (Wheel/Klick/Doppelklick). Port devd update.go's Mausteil (handleMouse,
// ~/Obsidian/tools/DeveloperDashboard/apps/cli-go/internal/tui/update.go:
// 459-524, mouseTreeClick/treeClickIdx:887-950) -- TWO deliberate deviations
// vs. devd's own handleMouse, both consequences of design decision f
// (view_browse_repo.go's own doc comment, "kein m.scroll-Feld existiert"):
//
//  1. Wheel moves the active view's CURSOR (m.cursorID/m.backlogList.cursor
//     via the two *CursorMove helpers, view_browse_repo.go/
//     view_browse_backlog.go) instead of devd's `m.scroll -= 3`/`+= 3`
//     non-Tree branch -- beans-tui has no scroll-offset field to move;
//     windowAround/windowStart already follow the cursor automatically on
//     render.
//  2. Left-click dispatch is a SINGLE `switch m.view` over two sibling
//     views (viewBrowseRepo/viewBacklog), not devd's Tree-vs-everything-else
//     split (devd's non-Tree views are all Chrome()-scrolled detail screens;
//     beans-tui's two views are both master-detail panes with their own
//     *ClickRow row-mapping, view_*.go).
//
// Toast-Klick-Vorrang (design decision a, overlay_show_toast.go's own doc
// comment): handleMouse's FIRST check is m.toastHit/dismissToast,
// unconditional on any overlay/form guard below it -- Port devd
// update.go:463-468 verbatim, Cross-Feature-Fix precedent (devd DD2-272/
// 273): a Toast is not modal, so its click-dismiss must reach the PO even
// while a form/overlay is open. TestToastClickDismissesEvenWithFormOpen
// (mouse_test.go) is the regression guard for this exact ordering.

import (
	"beans-tui/internal/data"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// minTreeWidthFloor/maxTreeWidthFloor mirror config.minTreeWidth/
// config.maxTreeWidth (internal/config/settings.go) -- unexported there, so
// duplicated here as literal values rather than imported (T6b, bean
// bt-pd22, T5-Review I01): the two packages' clamp ranges must agree, but
// there is no shared exported constant to single-source them from without
// widening config's public surface for two int literals.
const (
	minTreeWidthFloor = 24
	maxTreeWidthFloor = 60
)

// treeWidthFloor resolves configured (the caller's m.settings.Layout.
// TreeWidth, clickPaneGeometry's own treeWidth param below) into the
// numeric floor masterDetailWidths (view.go) receives. 0/unset -- every
// existing golden fixture/test model built via newModel/fixtureModel, whose
// m.settings is the config.Settings{} zero value until app.go's Run() (or a
// Settings-Form submit) populates it (types.go's own "settings" field
// doc-stamp) -- falls back to 24, the SAME value this file hardcoded before
// T6b, so the Golden snapshot tests (tree/chrome/backlog) stay
// byte-identical. A configured value is additionally
// clamped into [minTreeWidthFloor,maxTreeWidthFloor] -- consistent with
// config.validateSettings' own [24,60] clamp range: belt-and-braces here,
// since a LIVE model's m.settings already arrives pre-clamped via
// config.LoadSettings (app.go Run()) / the settings-Submit's own re-read
// (config.LoadSettings, box_confirm_create.go submitForm "settings" case).
func treeWidthFloor(configured int) int {
	if configured <= 0 {
		return minTreeWidthFloor
	}
	if configured < minTreeWidthFloor {
		return minTreeWidthFloor
	}
	if configured > maxTreeWidthFloor {
		return maxTreeWidthFloor
	}
	return configured
}

// clickPaneGeometry recomputes the numeric frame geometry every one of the
// two View functions (viewBrowseRepo/viewBacklog, view_browse_repo.go/
// view_browse_backlog.go) build ahead of their own two-pane body -- bodyH/
// lw/rw plus the screen origin (originX/originY) of the pane's FIRST
// windowed content row (the search-or-summary head row already consumed --
// PF-10, bean bt-uyzf, removed the pane's own title+separator lines, so
// nothing else precedes it now). Single source for BOTH halves: the two View functions
// themselves (each now calls this instead of an independently maintained
// avail/bodyH/lw/rw copy) AND treeClickRow/backlogClickRow (Golden-Rule-
// Drift-Schutz, mirrors windowStart's own shared-geometry rationale,
// view_browse_repo.go). head/localKeys are the caller's own EXACT
// breadcrumb()/footer() strings (browseRepoChrome/backlogChrome), so headH
// always matches that view's real head height (narrow-terminal 2-line
// breadcrumb wrap included). treeWidth is the caller's OWN
// m.settings.Layout.TreeWidth (T6b, bean bt-pd22, T5-Review I01 -- BEFORE
// this task, every caller passed a hardcoded "24" straight into
// masterDetailWidths below, a silent no-op for the Settings-Form's own
// tree_width field) -- resolved via treeWidthFloor just above before
// reaching masterDetailWidths, so every one of this function's View-function
// callers AND its *ClickRow callers picks up a configured Baumbreite
// consistently, the same Single-Source guarantee this doc comment's opening
// paragraph already establishes for bodyH/lw/rw/originX/originY themselves.
func clickPaneGeometry(w, h int, head, localKeys string, treeWidth int) (bodyH, lw, rw, originX, originY int) {
	if w <= 0 {
		w = 80
	}
	if h <= 0 {
		h = 24
	}
	innerW := w - 2
	innerH := h - 2
	footH := lipgloss.Height(localKeys) + 2             // + status line + divider above footer
	avail := innerH - lipgloss.Height(head) - footH - 1 // - divider under the top bar
	if avail < 4 {
		avail = 18 // height unknown (init/tests) -> generous fallback, mirrors Chrome()
	}
	bodyH = avail - 2 // both panes add their own border (+2, Golden Rule #1)
	if bodyH < 1 {
		bodyH = 1
	}
	lw, rw = masterDetailWidths(innerW, treeWidthFloor(treeWidth))
	// originY: outer top border(1) + head line(s) + divider(1) + the pane's
	// OWN top border(1) -- its title line + separator line no longer exist
	// (PF-10, design-spec.md §15, epic-E7-plan.md »Task 5«, bean bt-uyzf:
	// renderPane dropped both, the Breadcrumb alone carries view identity
	// now; the pane's own top border is a DIFFERENT line and stays). The row
	// after this is row 0 of the caller's own windowed-content index space
	// (the search/summary head line).
	originY = 1 + lipgloss.Height(head) + 1 + 1
	originX = 1 // outer border's left column
	return
}

// handleMouse dispatches every tea.MouseMsg (Update()'s own case, update.go
// -- placed ahead of the `if m.form != nil` fallback per that case's own
// placement comment). Toast-Klick-Vorrang first (design decision a, this
// file's own doc comment), THEN the overlay guard (forms/pickers/menus/
// palette/search/filter/help/quit-confirm ignore the mouse entirely, devd
// precedent update.go:470-474), THEN wheel, THEN left-click dispatch.
func (m model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress && m.toastHit(msg.X, msg.Y) {
		return m.dismissToast()
	}

	// Modale/Overlays sind tastaturgesteuert -- Maus ignorieren (kein
	// Fehlklick-Fokus), devd precedent update.go:470-474. m.view ==
	// viewLobby closes the gap bt-mne6's own Notes-für-T6 flagged (T4 could
	// not add this guard since viewLobby did not exist yet) -- E5 Task 6
	// (bean bt-zhwl): the Lobby is keyboard-only in V1 (design-spec.md §9
	// scope cut, no click-to-select row mapping), so a click anywhere on it
	// is a no-op, same as every other full-capture state in this list.
	if m.form != nil || m.overlay != overlayNone || m.paletteOpen || m.filterOpen ||
		m.searchActive || m.helpOpen || m.confirmQuit || m.view == viewLobby {
		return m, nil
	}

	switch msg.Button {
	case tea.MouseButtonWheelUp:
		return m.wheelMove(-1), nil
	case tea.MouseButtonWheelDown:
		return m.wheelMove(1), nil
	case tea.MouseButtonLeft:
		if msg.Action != tea.MouseActionPress {
			return m, nil
		}
		switch m.view {
		case viewBrowseRepo:
			return m.mouseTreeClick(msg)
		case viewBacklog:
			return m.mouseBacklogClick(msg)
		}
	}
	return m, nil
}

// wheelMove dispatches a single wheel tick (delta = ±1) to the active
// view's own cursor-move helper -- design decision f: Wheel bewegt den
// View-eigene CURSOR (nicht einen Scroll-Offset), same ±1 step as a single
// i/k keypress (devd's own Tree-wheel precedent, update.go:476-487 -- devd's
// NON-Tree ×3 m.scroll branch has no beans-tui equivalent, this file's own
// doc comment).
func (m model) wheelMove(delta int) model {
	switch m.view {
	case viewBrowseRepo:
		return m.treeCursorMove(m.visibleNodes(), delta)
	case viewBacklog:
		return m.backlogCursorMove(m.backlogVisible(), delta)
	}
	return m
}

// mouseTreeClick dispatches a Tree left-click (design decision f): resolves
// the clicked row via treeClickRow (view_browse_repo.go, pure geometry, no
// side effect), sets the cursor, then applies devd-D03's Doppelklick
// semantics (Port devd mouseTreeClick verbatim, update.go:922-950) -- a
// SECOND click on the SAME node within doubleClickInterval collapses an
// OPEN node; a single click on a CLOSED expandable node expands it; a
// single click on an ALREADY-open node only moves the cursor (no toggle,
// devd D03).
func (m model) mouseTreeClick(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	nodes := m.visibleNodes()
	idx, ok := treeClickRow(m, nodes, msg)
	if !ok {
		// B07 (design-spec.md §15 PF-16, bean bt-duz7): treeClickRow
		// returns ok=false BOTH for a click outside the Tree pane's own
		// column (right Detail pane) AND for one inside it that misses
		// every rendered row (above the search line, or past the last
		// node) -- delegating unconditionally to mouseDetailClick is safe
		// for the latter case too: detailClickRow (mouse.go, below)
		// independently re-checks msg.X against the Detail pane's OWN
		// column span and rejects (ok=false, no-op) a click that is still
		// inside the Tree's column, so no cross-pane leakage occurs.
		// Implementer decision (bean bt-duz7 Architektur-Vorgabe #4,
		// "saubersten Diff-Groesse"): the smallest possible diff over a
		// second, independently-computed X-bounds pre-check here.
		return m.mouseDetailClick(msg)
	}
	n := nodes[idx]
	m.cursorID = n.id

	// Doppelklick = zweiter Klick auf DENSELBEN Node-Index innerhalb des
	// Zeitfensters. lastClickAt=zero (Zero-Value) ⇒ riesiges Delta ⇒ erster
	// Klick nie Doppelklick (Port devd mouseTreeClick verbatim).
	now := m.now()
	isDouble := idx == m.lastClickIdx && now.Sub(m.lastClickAt) < doubleClickInterval
	m.lastClickIdx = idx
	m.lastClickAt = now

	if n.hasKids {
		switch {
		case n.open && isDouble:
			return m.setExpanded(n, false), nil
		case !n.open:
			return m.setExpanded(n, true), nil
		}
	}
	return m, nil
}

// mouseBacklogClick dispatches a Backlog left-click: resolves the clicked
// row via backlogClickRow (view_browse_backlog.go) and sets the cursor --
// no Doppelklick semantics (flat list, plan epic-E5-plan.md »Task 4« Step
// 6: "kein Doppelklick-Bedarf bei Backlog").
func (m model) mouseBacklogClick(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	vis := m.backlogVisible()
	idx, ok := backlogClickRow(m, vis, msg)
	if !ok {
		// B07: mirrors mouseTreeClick's own delegation above (same
		// rationale -- backlogClickRow's ok=false covers BOTH "wrong pane"
		// and "in-pane but no row hit", detailClickRow independently
		// re-validates the X bounds either way).
		return m.mouseDetailClick(msg)
	}
	m.backlogList.setLen(len(vis))
	m.backlogList.cursor = idx
	return m, nil
}

// detailClickRow maps a Detail-Pane click to either a Section-Header hit
// (fieldIdx == -1) or a Meta field-row hit (secIdx == metaSectionIdx,
// fieldIdx >= 0) -- B07 (design-spec.md §15 PF-16, bean bt-duz7), ANALOG
// treeClickRow/backlogClickRow above: reconstructs the pane's geometry via
// the SAME clickPaneGeometry + browseRepoChrome/backlogChrome helpers those
// two use (Golden-Rule-Drift-Schutz) -- picks the caller's OWN chrome via
// m.view, since the Detail pane renders identically from BOTH viewBrowseRepo
// and viewBacklog (renderBeanAccordionPane, view_browse_repo.go), just
// alongside a different left pane/footer.
//
// The row->section/field walk below is GROUNDED against the real render
// pipeline, not a second, independently-maintained height formula: it reuses
// the EXACT SAME state (m.detailFocus/m.secCursor/m.accOpen/m.fieldCursor/
// m.detailLevel) and the EXACT SAME beanSections() call renderAccordionPane
// itself makes (view_browse_repo.go), then walks secs with the IDENTICAL
// isOpen/activeSec/fieldStrip-shown conditionals renderAccordion (accordion.
// go) uses to build its own line-by-line output. Section-body line counts
// are measured via lipgloss.Height on the SAME s.body string renderAccordion
// hands to boxStyle.Render -- PaddingLeft alone never changes a line count,
// so this is byte-equivalent without reconstructing boxStyle's own
// Width/Padding here. A change to either function's algorithm is caught by
// the render-grounded tests in mouse_test.go, which locate click coordinates
// by searching the REAL m.View() output (mirrors treeClickAt/
// leftPaneClickAt's own pattern) -- never by hand-deriving a row number (the
// "selbst-referenzielle Geometrie-Test" trap E7 already hit once, this
// task's own Architektur-Vorgabe #2 warns against repeating it).
//
// v1 simplification (bean bt-duz7 Architektur-Vorgabe #2, no PO-Wortlaut
// requires more): ONLY Meta (metaSectionIdx) has a fixed, direct
// Zeile->Feldindex mapping (Zeile 0 = title ... Zeile 6 = updated_at, per
// metaFields' fixed order) -- a click landing inside an OPEN Body/Relations/
// History section's body (or Relations' own fieldStrip row) resolves to
// that section's OWN header hit (fieldIdx == -1), never a field index.
func detailClickRow(m model, b *data.Bean, msg tea.MouseMsg) (secIdx, fieldIdx int, ok bool) {
	w, h := m.width, m.height
	if w <= 0 {
		w = 80
	}
	if h <= 0 {
		h = 24
	}
	innerW := w - 2

	var head, localKeys string
	if m.view == viewBacklog {
		head, localKeys = m.backlogChrome(innerW)
	} else {
		head, localKeys = m.browseRepoChrome(innerW)
	}

	bodyH, lw, rw, originX, originY := clickPaneGeometry(w, h, head, localKeys, m.settings.Layout.TreeWidth)

	if msg.X < originX+lw || msg.X >= originX+lw+rw {
		return 0, 0, false // left pane, or off-screen -- no Detail target
	}
	clickRow := msg.Y - originY
	if clickRow < 0 || clickRow >= bodyH {
		return 0, 0, false // above the pane, or past renderPane's own Golden-Rule-#1 line cap
	}
	if b == nil {
		return 0, 0, false // "(no selection)" placeholder -- nothing to hit
	}

	// detailHeaderBlock (view_detail_bean.go) ALWAYS renders 5 fixed rows
	// (ID/Title/blank/type-status-prio/blank) ahead of the Accordion --
	// bt-duz7's own PO-Wortlaut ("Kopfblock-Offset 5 Zeilen beachten!").
	const headerBlockLines = 5
	if clickRow < headerBlockLines {
		return 0, 0, false // Kopfblock -- never a Section-/Feld-Treffer
	}
	accordionRow := clickRow - headerBlockLines

	bodyW := rw - 4
	if bodyW < 1 {
		bodyW = 1
	}
	secs := beanSections(m.idx, b, bodyW, m.detailFocus, m.secCursor, m.fieldCursor, m.detailLevel)

	row := 0
	for i, s := range secs {
		n := i + 1
		isOpen := n == m.accOpen || n == 1 // PF-1: Meta always open, mirrors renderAccordion
		if accordionRow == row {
			return i, -1, true // this section's own header row
		}
		row++
		if !isOpen {
			continue
		}
		activeSec := m.detailFocus && m.secCursor == i
		if activeSec && i != 0 && len(s.fields) > 0 {
			if accordionRow == row {
				return i, -1, true // fieldStrip row -- v1: section-level hit
			}
			row++
		}
		bodyLines := lipgloss.Height(s.body)
		if accordionRow >= row && accordionRow < row+bodyLines {
			if i == metaSectionIdx {
				if fi := accordionRow - row; fi >= 0 && fi < len(s.fields) {
					return i, fi, true
				}
			}
			return i, -1, true // Body/Relations/History body, or an out-of-range Meta row -- section-level hit
		}
		row += bodyLines
	}
	return 0, 0, false // past the last rendered section (renderPane's own line cap already excluded above, defensive)
}

// mouseDetailClick dispatches a Detail-Pane left-click (B07, design-spec.md
// §15 PF-16, bean bt-duz7): resolves the clicked row via detailClickRow
// (pure geometry, no side effect, above) then applies the SAME
// lastClickIdx/lastClickAt Doppelklick machinery mouseTreeClick already uses
// (Architektur-Vorgabe #3, "gleiche Felder auf model wiederverwenden, KEIN
// zweites Zeitfenster-Paar"). clickKey folds (secIdx, fieldIdx) into ONE int
// so a Meta field click and a Section-header click can never alias each
// other (secIdx*10+fieldIdx+1: section hits land on the multiples of 10 --
// 0/10/20/30, Meta field hits on 1..7); a genuine collision against a
// Tree/Backlog row index sharing the SAME lastClickIdx int in the SAME
// 500ms window remains a theoretical, explicitly-accepted edge case (the
// bean's own "reuse, no new pair" directive), not a new bug introduced here.
//
// A single click ALWAYS selects (Fall a/b: activates+expands the section,
// additionally enters field level for a Meta field) and NEVER opens an
// overlay. A double click (or an immediately-repeated click on an
// already-selected target, same window -- the two are the SAME event,
// mirrors mouseTreeClick's own isDouble, not a second mechanism) on a Meta
// field additionally calls activateDetailField (Fall c, identical to the
// Enter-Kaskade, update.go). Per bt-y2iw's "Notes for bt-duz7" (BODY's
// enter-cascade equivalent has no field to double-click), a double click on
// the BODY section's OWN header instead opens $EDITOR via the SAME
// openBodyEditor helper keyDetailFocus's enter-on-BODY branch uses
// (update.go) -- never a duplicated editorTarget/editorETag assignment.
func (m model) mouseDetailClick(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	b := m.focusedBean()
	if b == nil {
		return m, nil
	}
	secIdx, fieldIdx, ok := detailClickRow(m, b, msg)
	if !ok {
		return m, nil
	}

	clickKey := secIdx*10 + fieldIdx + 1
	now := m.now()
	isDouble := clickKey == m.lastClickIdx && now.Sub(m.lastClickAt) < doubleClickInterval
	m.lastClickIdx = clickKey
	m.lastClickAt = now

	m.detailFocus = true
	m.secCursor = secIdx
	m.accOpen = secIdx + 1

	if fieldIdx < 0 {
		m.detailLevel = 0
		m.fieldCursor = 0
		if isDouble && secIdx == bodySectionIdx {
			return m.openBodyEditor(b)
		}
		return m, nil
	}

	// detailClickRow only ever returns fieldIdx >= 0 nested under
	// i == metaSectionIdx (its own doc comment above) -- metaFields(b) is
	// therefore the SAME fixed 7-entry slice s.fields already was.
	m.detailLevel = 1
	m.fieldCursor = fieldIdx

	if isDouble {
		fields := metaFields(b)
		if fieldIdx < len(fields) {
			return m.activateDetailField(b, fields[fieldIdx])
		}
	}
	return m, nil
}
