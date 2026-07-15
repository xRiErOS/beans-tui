package tui

// mouse.go — E5 Task 4 (bean bt-mne6, epic bt-5h4d, design decision f): Maus
// (Wheel/Klick/Doppelklick). Port devd update.go's Mausteil (handleMouse,
// ~/Obsidian/tools/DeveloperDashboard/apps/cli-go/internal/tui/update.go:
// 459-524, mouseTreeClick/treeClickIdx:887-950) -- TWO deliberate deviations
// vs. devd's own handleMouse, both consequences of design decision f
// (view_browse_repo.go's own doc comment, "kein m.scroll-Feld existiert"):
//
//  1. Wheel moves the active view's CURSOR (m.cursorID/m.backlogList.cursor/
//     m.reviewCursor via the three *CursorMove helpers, view_browse_repo.go/
//     view_browse_backlog.go/view_review_cockpit.go) instead of devd's
//     `m.scroll -= 3`/`+= 3` non-Tree branch -- beans-tui has no scroll-
//     offset field to move; windowAround/windowStart already follow the
//     cursor automatically on render.
//  2. Left-click dispatch is a SINGLE `switch m.view` over three sibling
//     views (viewBrowseRepo/viewBacklog/viewReviewCockpit), not devd's
//     Tree-vs-everything-else split (devd's non-Tree views are all Chrome()-
//     scrolled detail screens; beans-tui's three views are all master-detail
//     panes with their own *ClickRow row-mapping, view_*.go).
//
// Toast-Klick-Vorrang (design decision a, overlay_show_toast.go's own doc
// comment): handleMouse's FIRST check is m.toastHit/dismissToast,
// unconditional on any overlay/form guard below it -- Port devd
// update.go:463-468 verbatim, Cross-Feature-Fix precedent (devd DD2-272/
// 273): a Toast is not modal, so its click-dismiss must reach the PO even
// while a form/overlay is open. TestToastClickDismissesEvenWithFormOpen
// (mouse_test.go) is the regression guard for this exact ordering.

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// clickPaneGeometry recomputes the numeric frame geometry every one of the
// three View functions (viewBrowseRepo/viewBacklog/viewReviewCockpit,
// view_browse_repo.go/view_browse_backlog.go/view_review_cockpit.go) build
// ahead of their own two-pane body -- bodyH/lw/rw plus the screen origin
// (originX/originY) of the pane's FIRST windowed content row (title +
// separator + the search-or-summary head row already consumed). Single
// source for BOTH halves: the three View functions themselves (each now
// calls this instead of an independently maintained avail/bodyH/lw/rw copy)
// AND treeClickRow/backlogClickRow/reviewClickRow (Golden-Rule-Drift-Schutz,
// mirrors windowStart's own shared-geometry rationale, view_browse_repo.go).
// head/localKeys are the caller's own EXACT breadcrumb()/footer() strings
// (browseRepoChrome/backlogChrome/reviewCockpitChrome), so headH always
// matches that view's real head height (narrow-terminal 2-line breadcrumb
// wrap included).
func clickPaneGeometry(w, h int, head, localKeys string) (bodyH, lw, rw, originX, originY int) {
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
	lw, rw = masterDetailWidths(innerW, 24)
	// originY: outer top border(1) + head line(s) + divider(1) + the pane's
	// OWN top border(1) + its title line(1) + its separator line(1) -- the
	// row after this is row 0 of the caller's own windowed-content index
	// space (the search/summary head line).
	originY = 1 + lipgloss.Height(head) + 1 + 1 + 1 + 1
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
	// Fehlklick-Fokus), devd precedent update.go:470-474. m.view == viewLobby
	// is NOT checked here yet -- that view does not exist until Task 6
	// (Lobby V1), see Deviations/Notes for T5/T6.
	if m.form != nil || m.overlay != overlayNone || m.paletteOpen || m.filterOpen ||
		m.searchActive || m.helpOpen || m.confirmQuit {
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
		case viewReviewCockpit:
			return m.mouseReviewClick(msg)
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
	case viewReviewCockpit:
		return m.reviewCursorMove(reviewFlat(m.idx), delta)
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
		return m, nil
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
		return m, nil
	}
	m.backlogList.setLen(len(vis))
	m.backlogList.cursor = idx
	return m, nil
}

// mouseReviewClick dispatches a Review-Cockpit left-click: resolves the
// clicked row via reviewClickRow (view_review_cockpit.go) and sets
// m.reviewCursor -- a header/separator row miss (ok=false) is a no-op, same
// as clicking outside a pane entirely.
func (m model) mouseReviewClick(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	rs := newReviewState(m.idx)
	idx, ok := reviewClickRow(m, rs, msg)
	if !ok {
		return m, nil
	}
	m.reviewCursor = idx
	return m, nil
}
