package tui

// view_browse_repo.go — the Browse Primat-View (design-spec.md §6 V2 basis):
// a two-pane master-detail layout, left = expandable Tree (Milestones as
// roots -> Epics/Features -> leaves, plus a synthetic "(verwaist)" root for
// orphans, MANDATORY per bean bt-7jr8/T3-review Q01), right = a placeholder
// detail preview (full accordion lands in E2). Port-pattern reference: devd
// view_browse_project.go (treeNode flattening + D08 cursor-bar rendering,
// ~/Obsidian/tools/DeveloperDashboard/apps/cli-go/internal/tui/
// view_browse_project.go:382-398).

import (
	"path/filepath"
	"sort"
	"strings"

	"beans-tui/internal/data"
	"beans-tui/internal/theme"
	keybind "github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// treeNode is a single visible (already-flattened) row of the tree. Index-free
// (holds a *data.Bean pointer straight into the Index, mirroring data.Index's
// own contract: consumers must not mutate it) -- depth-first order matches
// render order 1:1, so cursor position == slice index.
type treeNode struct {
	id      string // bean ID, or orphanRootID for the synthetic "(verwaist)" root
	bean    *data.Bean
	depth   int
	hasKids bool
	open    bool
	orphan  bool // true only for the synthetic orphan-root node itself
}

// flattenTree walks idx depth-first (Roots() -> Children, expand-state from
// expanded) into the visible row list, appending the synthetic orphan root
// (+ its subtree, when expanded) last. Returns nil for a nil idx (pre-load).
func flattenTree(idx *data.Index, expanded map[string]bool) []treeNode {
	if idx == nil {
		return nil
	}

	var nodes []treeNode
	ancestors := map[string]bool{}
	for _, b := range idx.Roots() {
		nodes = appendBeanNode(nodes, idx, b, 0, expanded, ancestors)
	}

	orphans := collectOrphans(idx)
	if len(orphans) > 0 {
		open := expanded[orphanRootID]
		nodes = append(nodes, treeNode{id: orphanRootID, depth: 0, hasKids: true, open: open, orphan: true})
		if open {
			for _, b := range orphans {
				nodes = appendBeanNode(nodes, idx, b, 1, expanded, ancestors)
			}
		}
	}
	return nodes
}

// appendBeanNode appends b (and, if expanded, its children) depth-first.
// ancestors is the current DFS path -- a bean already on its own ancestor
// path is skipped (defensive cycle guard; beans' frontmatter is hand-
// editable, a parent cycle is a data error but must never hang/crash bt).
func appendBeanNode(nodes []treeNode, idx *data.Index, b *data.Bean, depth int, expanded map[string]bool, ancestors map[string]bool) []treeNode {
	if ancestors[b.ID] {
		return nodes
	}
	children := idx.Children[b.ID]
	open := expanded[b.ID]
	nodes = append(nodes, treeNode{id: b.ID, bean: b, depth: depth, hasKids: len(children) > 0, open: open})
	if open && len(children) > 0 {
		ancestors[b.ID] = true
		for _, c := range children {
			nodes = appendBeanNode(nodes, idx, c, depth+1, expanded, ancestors)
		}
		delete(ancestors, b.ID)
	}
	return nodes
}

// collectOrphans returns every bean whose Parent is set but does not resolve
// to a known bean (dangling parent -- beans-legal, `beans check` only
// reports it as broken_links). Sorted Title -> ID for determinism: idx.ByID
// iteration order is a Go map, so this must never be rendered unsorted (would
// break golden-test determinism and PO expectations alike). data.sortBeans
// itself is unexported to this package, so this uses its own simple, stable
// tie-break -- exact upstream tier order is not required for the rare orphan
// path.
func collectOrphans(idx *data.Index) []*data.Bean {
	var out []*data.Bean
	for _, b := range idx.ByID {
		if b.Parent == "" {
			continue
		}
		if _, ok := idx.ByID[b.Parent]; ok {
			continue
		}
		out = append(out, b)
	}
	sort.SliceStable(out, func(i, j int) bool {
		ti, tj := strings.ToLower(out[i].Title), strings.ToLower(out[j].Title)
		if ti != tj {
			return ti < tj
		}
		return out[i].ID < out[j].ID
	})
	return out
}

// visibleNodes flattens the model's current idx+expanded state.
func (m model) visibleNodes() []treeNode {
	return flattenTree(m.idx, m.expanded)
}

// cursorPos finds m.cursorID's index in nodes, defaulting to 0 (covers both
// the pre-load state, cursorID == "", and a cursor whose bean disappeared
// without a matching reload clamp having run yet).
func (m model) cursorPos(nodes []treeNode) int {
	for i, n := range nodes {
		if n.id == m.cursorID {
			return i
		}
	}
	return 0
}

// treeNodeMarker returns the expand marker (▾ open / ▸ closed / blank leaf).
func treeNodeMarker(n treeNode) string {
	if !n.hasKids {
		return "  "
	}
	if n.open {
		return "▾ "
	}
	return "▸ "
}

// treeRowText renders one row's plain content: indent + expand marker +
// status glyph + type icon + ID (Sapphire) + title (design-spec.md §8).
func treeRowText(n treeNode) string {
	indent := strings.Repeat("  ", n.depth)
	marker := treeNodeMarker(n)
	if n.orphan {
		return indent + marker + theme.Dim.Render("(verwaist)")
	}
	b := n.bean
	return indent + marker + theme.StatusIcon(b.Status) + " " + theme.TypeIcon(b.Type) + " " + theme.Key.Render(b.ID) + " " + b.Title
}

// treeRows renders every visible row, applying the D08 cursor treatment to
// the cursor's row only (devd view_browse_project.go:382-398): a leading `▌`
// bar plus the WHOLE row accent-tinted (own per-cell colors stripped first).
// focused=false (Detail pane has focus) freezes the cursor muted instead of
// accent -- only the focused pane's cursor is highlighted (devd D03).
func (m model) treeRows(nodes []treeNode, focused bool) []string {
	pos := m.cursorPos(nodes)
	rows := make([]string, len(nodes))
	for i, n := range nodes {
		text := treeRowText(n)
		if i == pos {
			plain := ansi.Strip(text)
			if focused {
				rows[i] = theme.Accent.Render("▌" + plain)
			} else {
				rows[i] = theme.Dim.Render("▌" + plain)
			}
		} else {
			rows[i] = " " + text
		}
	}
	return rows
}

// renderDetailPane renders the placeholder detail preview: title + a meta
// line (status/type/priority/tags) of the cursored bean, via theme tokens
// only. The full accordion (Meta/Body/Beziehungen/Historie sections) is E2
// scope (design-spec.md §6 V4) -- this is deliberately minimal.
func (m model) renderDetailPane(nodes []treeNode, w, h int, focused bool) string {
	pos := m.cursorPos(nodes)
	var rows []string
	if pos >= 0 && pos < len(nodes) && !nodes[pos].orphan && nodes[pos].bean != nil {
		b := nodes[pos].bean
		rows = append(rows, theme.Header.Render(b.Title))
		meta := theme.StatusStyle(b.Status).Render(b.Status) + "  " + theme.TypeStyle(b.Type).Render(b.Type)
		if b.Priority != "" {
			meta += "  " + theme.Priority(b.Priority)
		}
		if t := tagsInline(b.Tags); t != "" {
			meta += "  " + t
		}
		rows = append(rows, meta)
	} else {
		rows = append(rows, theme.Dim.Render("(no selection)"))
	}
	return renderPane(pane{title: "Detail", rows: rows}, w, h, focused)
}

// repoLabel is the breadcrumb `> repo` segment: the repo directory's base
// name (design-spec.md §7's "> repo: Titel" format; port-adaptation vs.
// devd's project slug, view.go's breadcrumb doc comment).
func (m model) repoLabel() string {
	if m.repoDir == "" {
		return "bt"
	}
	return filepath.Base(m.repoDir)
}

// View dispatches on viewID (devd port convention: enum + switch in view).
// T8 ships exactly one view; later epics add cases here, never branches.
func (m model) View() string {
	switch m.view {
	default:
		return m.viewBrowseRepo()
	}
}

// viewBrowseRepo renders the two-pane master-detail Browse view. Mirrors
// Chrome()'s own algebra exactly (view.go) so the frame always fills
// width x height, just with a two-pane body instead of Chrome's single
// scroll body -- Golden Rule #1 still applies (no Height() on a bordered
// style; renderPane pads/caps to its h param instead).
func (m model) viewBrowseRepo() string {
	w, h := m.width, m.height
	if w <= 0 {
		w = 80
	}
	if h <= 0 {
		h = 24
	}
	innerW := w - 2
	innerH := h - 2

	globalHint := renderBindings([]keybind.Binding{keys.Refresh, keys.Help, keys.Quit})
	head := breadcrumb(m.repoLabel(), "Browse", globalHint, innerW)

	localHint := renderBindings([]keybind.Binding{keys.Up, keys.Down, keys.Left, keys.Right, keys.Enter, keys.Refresh}) + "  tab:focus"
	localKeys := footer(localHint, innerW)

	div := theme.Dim.Render(strings.Repeat("─", innerW))
	status := statusBar("", m.err, innerW)

	footH := lipgloss.Height(localKeys) + 2             // + status line + divider above footer
	avail := innerH - lipgloss.Height(head) - footH - 1 // - divider under the top bar
	if avail < 4 {
		avail = 18 // height unknown (init/tests) -> generous fallback, mirrors Chrome()
	}
	bodyH := avail - 2 // both panes add their own border (+2, Golden Rule #1)
	if bodyH < 1 {
		bodyH = 1
	}

	lw, rw := masterDetailWidths(innerW, 24)
	nodes := m.visibleNodes()
	treeBox := renderPane(pane{title: "Tree", rows: m.treeRows(nodes, !m.detailFocus)}, lw, bodyH, !m.detailFocus)
	detailBox := m.renderDetailPane(nodes, rw, bodyH, m.detailFocus)
	body := lipgloss.JoinHorizontal(lipgloss.Top, treeBox, detailBox)

	content := head + "\n" + div + "\n" + body + "\n" + div + "\n" + localKeys + "\n" + status
	out := outerBorder(content, innerW, true)

	if m.confirmQuit {
		out = placeOverlay(out, m.quitBox(), w, h)
	}
	return out
}
