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

	case watchUnavailableMsg:
		// I04: data.Watch failed to start -- surface once in the status line
		// (view_browse_repo.go) instead of silently never reacting to
		// on-disk changes; ctrl+r still reloads manually.
		m.watchUnavailable = true
		return m, nil

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
		if m.detailFocus {
			// enterDetailFocus-equivalent (devd view_detail_issue.go:236-243,
			// E2 Task 2): always re-enter the Detail-Accordion at Meta,
			// section level, field cursor 0 -- a stale cursor position from a
			// PREVIOUS detail-focus visit (possibly on a different bean, whose
			// section/field shape may differ) must never leak into this one.
			m.secCursor, m.accOpen, m.detailLevel, m.fieldCursor = 0, 1, 0, 0
		}
		return m, nil
	}

	if keybind.Matches(msg, keys.Refresh) {
		return m, loadCmd(m.client)
	}

	if m.detailFocus {
		return m.keyDetailFocus(msg)
	}
	return m.keyTree(msg)
}

// focusedBean returns the bean the Detail-Accordion currently targets,
// independent of which view is active (devd port focusedIssue,
// view_detail_issue.go:20-35) -- view-agnostic so Task 5's Backlog view can
// reuse keyDetailFocus verbatim once it adds its own case here.
func (m model) focusedBean() *data.Bean {
	switch m.view {
	default: // viewBrowseRepo (T8) -- Task 5 adds a viewBacklog case
		nodes := m.visibleNodes()
		pos := m.cursorPos(nodes)
		if pos < 0 || pos >= len(nodes) || nodes[pos].orphan {
			return nil
		}
		return nodes[pos].bean
	}
}

// keyDetailFocus drives the two-level Detail focus machine (Section cursor
// <-> Field cursor within Beziehungen; devd port view_detail_issue.go:
// 281-392). Port deviation vs. devd: no separate "header edit fields" layer
// (devd's section index 0) -- E2 has no edit-field concept (E3 scope), so
// section index 0 is Meta directly, no off-by-one vs. devd's secCursor-1.
func (m model) keyDetailFocus(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	b := m.focusedBean()
	if b == nil { // defensive guard: orphan-root cursor, no focusable bean
		m.detailFocus = false
		return m, nil
	}
	secs := beanSections(m.idx, b, 40) // width is render-time only; section COUNT is fixed (4)

	// B02 (Review-Runde 2, bean bt-2jve, Critical): clamp secCursor/
	// fieldCursor against the just-computed secs BEFORE any branch below
	// indexes into them. m.secCursor/m.fieldCursor are model state that
	// survives a beansLoadedMsg reload untouched -- a watch-reload between
	// keystrokes can shrink the focused bean's Beziehungen fields while the
	// user is still parked at field level, and secs[m.secCursor].
	// fields[m.fieldCursor] further down would then index out of range.
	// Single defensive clamp point, not sprinkled per-branch.
	if m.secCursor >= len(secs) {
		m.secCursor = len(secs) - 1
	}
	if m.secCursor < 0 {
		m.secCursor = 0
	}
	if fc := len(secs[m.secCursor].fields); m.fieldCursor >= fc {
		if fc == 0 {
			m.fieldCursor = 0
			m.detailLevel = 0 // no fields left in this section -- back to section level
		} else {
			m.fieldCursor = fc - 1
		}
	}

	if s := msg.String(); len(s) == 1 && s[0] >= '1' && s[0] <= '4' {
		m.secCursor = int(s[0]-'0') - 1
		m.accOpen = int(s[0] - '0')
		m.detailLevel = 0
		m.fieldCursor = 0
		return m, nil
	}

	switch navKey(msg.String()) {
	case "up":
		if m.detailLevel == 0 && m.secCursor > 0 {
			m.secCursor--
			m.accOpen = m.secCursor + 1
		} else if m.detailLevel == 1 && m.fieldCursor > 0 {
			m.fieldCursor--
		}
		return m, nil
	case "down":
		if m.detailLevel == 0 && m.secCursor < len(secs)-1 {
			m.secCursor++
			m.accOpen = m.secCursor + 1
		} else if m.detailLevel == 1 && m.fieldCursor < len(secs[m.secCursor].fields)-1 {
			m.fieldCursor++
		}
		return m, nil
	case "right":
		if m.detailLevel == 0 && len(secs[m.secCursor].fields) > 0 {
			m.detailLevel = 1
			m.fieldCursor = 0
		}
		return m, nil
	case "left":
		if m.detailLevel == 1 {
			m.detailLevel = 0
		} else {
			m.detailFocus = false
		}
		return m, nil
	}

	if keybind.Matches(msg, keys.Enter) && m.detailLevel == 1 {
		f := secs[m.secCursor].fields[m.fieldCursor]
		if f.beanID == "" {
			return m, nil // unresolved reference -- nothing to jump to
		}
		m.expanded = expandAncestorsOf(m.idx, m.expanded, f.beanID) // I01: clone-based
		m.cursorID = f.beanID
		m.detailFocus = false
		return m, nil
	}
	return m, nil
}

// expandAncestorsOf returns a NEW expanded map (I01 copy-on-write) with every
// ancestor of id (walking Parent up to a root) marked expanded, so a
// relation-jump target is guaranteed visible in the next visibleNodes() call.
// B01 (Review-Runde 2, bean bt-2jve, Critical): visited guards the walk --
// beans' frontmatter is hand-editable, so a Parent cycle (A -> B -> A) is a
// data error but must never hang/freeze the TUI; same defensive pattern as
// appendBeanNode's per-path ancestors map (view_browse_repo.go).
func expandAncestorsOf(idx *data.Index, expanded map[string]bool, id string) map[string]bool {
	out := cloneBoolMap(expanded)
	visited := map[string]bool{id: true}
	b, ok := idx.ByID[id]
	for ok && b.Parent != "" && !visited[b.Parent] {
		visited[b.Parent] = true
		out[b.Parent] = true
		b, ok = idx.ByID[b.Parent]
	}
	return out
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

// setExpanded sets n's expand state in m.expanded; a no-op for leaves. I01
// (bean bt-7jr8 T8-review): clones m.expanded via cloneBoolMap before writing
// -- expanded is COPY-ON-WRITE (types.go doc-stamp), never mutated in place.
func (m model) setExpanded(n treeNode, open bool) model {
	if !n.hasKids {
		return m
	}
	m.expanded = cloneBoolMap(m.expanded)
	if open {
		m.expanded[n.id] = true
	} else {
		delete(m.expanded, n.id)
	}
	return m
}
