package tui

// update.go — the Update dispatcher (Elm architecture): routes tea.Msg to
// state transitions only. No rendering here (view_browse_repo.go), no
// Cmd-only definitions here (messages.go) -- port convention from devd
// update.go.

import (
	"beans-tui/internal/data"

	keybind "github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case beansLoadedMsg:
		return m.applyLoaded(msg), nil

	case watchMsg:
		// data.Watch's onChange fired -- always a full reload, never partial
		// (design decision D02).
		return m, loadCmd(m.client)

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

// applyLoaded rebuilds the Index from a (initial or reload) beansLoadedMsg
// and restores the cursor by bean ID: the same bean keeps the same ID across
// a reload, so the cursor stays on it; a bean that vanished (deleted/
// renamed) clamps to roughly where it used to sit rather than jumping back
// to the top (US-10: reload must never lose the PO's place arbitrarily).
func (m model) applyLoaded(msg beansLoadedMsg) model {
	if msg.err != nil {
		m.err = msg.err.Error()
		return m
	}
	m.err = ""

	// Capture the cursor's position in the OLD tree before swapping idx, so
	// a vanished bean can clamp near its old spot instead of resetting to 0.
	oldPos := 0
	if m.idx != nil {
		oldPos = m.cursorPos(flattenTree(m.idx, m.expanded))
	}

	m.idx = data.NewIndex(msg.beans)
	nodes := m.visibleNodes()

	if len(nodes) == 0 {
		m.cursorID = ""
		return m
	}
	for _, n := range nodes {
		if n.id == m.cursorID {
			return m // exact bean still present -> cursor unchanged
		}
	}
	if oldPos >= len(nodes) {
		oldPos = len(nodes) - 1
	}
	m.cursorID = nodes[oldPos].id
	return m
}

// handleKey is the top-level key dispatch (devd keys.go port pattern):
// the quit-confirm modal fully captures input first; then a couple of
// view-global keys (ctrl+c hard-quit, tab focus-swap -- Q01, deliberately
// NOT part of keymap.go's Right binding); then the tree/detail handlers.
func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.confirmQuit {
		return m.keyConfirmQuit(msg)
	}

	switch msg.String() {
	case "ctrl+c": // immediate quit, no confirm (bean bt-7jr8: distinct from `q`)
		return m, tea.Quit
	case "q":
		return m.requestQuit()
	case "tab": // Q01: view-local Tree<->Detail focus swap
		m.detailFocus = !m.detailFocus
		return m, nil
	}

	if keybind.Matches(msg, keys.Refresh) {
		return m, loadCmd(m.client)
	}

	if m.detailFocus {
		// Detail-pane navigation lands in E2 (full accordion); T8's
		// placeholder preview has nothing interactive yet.
		return m, nil
	}
	return m.keyTree(msg)
}

// keyTree drives the tree: up/down move the cursor, right/left expand/
// collapse, enter toggles expand (no-op on a leaf, per task scope).
func (m model) keyTree(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	nodes := m.visibleNodes()
	if len(nodes) == 0 {
		return m, nil
	}
	pos := m.cursorPos(nodes)

	switch navKey(msg.String()) {
	case "up":
		if pos > 0 {
			pos--
		}
		m.cursorID = nodes[pos].id
		return m, nil
	case "down":
		if pos < len(nodes)-1 {
			pos++
		}
		m.cursorID = nodes[pos].id
		return m, nil
	case "right":
		return m.setExpanded(nodes[pos], true), nil
	case "left":
		return m.setExpanded(nodes[pos], false), nil
	}

	if keybind.Matches(msg, keys.Enter) {
		n := nodes[pos]
		if !n.hasKids {
			return m, nil // leaf: no-op for now (E2 opens detail focus instead)
		}
		return m.setExpanded(n, !n.open), nil
	}
	return m, nil
}

// setExpanded sets n's expand state in m.expanded; a no-op for leaves.
func (m model) setExpanded(n treeNode, open bool) model {
	if !n.hasKids {
		return m
	}
	if open {
		m.expanded[n.id] = true
	} else {
		delete(m.expanded, n.id)
	}
	return m
}
