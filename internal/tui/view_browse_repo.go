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
	"strings"

	"beans-tui/internal/data"
	"beans-tui/internal/theme"
	keybind "github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
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
	cycles := collectCycleOrphans(idx, orphans)
	if len(orphans) > 0 || len(cycles) > 0 {
		open := expanded[orphanRootID]
		nodes = append(nodes, treeNode{id: orphanRootID, depth: 0, hasKids: true, open: open, orphan: true})
		if open {
			for _, b := range orphans {
				nodes = appendBeanNode(nodes, idx, b, 1, expanded, ancestors)
			}
			// B02 (T8 Opus review, bean bt-7jr8): cycle-trapped beans render as
			// FLAT depth-1 rows -- never recursed into via appendBeanNode. Every
			// bean whose whole Parent-chain loops back on itself (A -> B -> A)
			// is, by construction, only reachable downward from another member
			// of the very same cycle (see collectCycleOrphans), so recursing
			// would either re-hit the cycle guard for nothing or -- for a bean
			// merely hanging off a cycle member -- render it a second time here
			// AND nested under its (also-unreachable) parent. A flat row keeps
			// the fix simple and duplicate-free: every such bean gets exactly
			// one visible row, just without its undefined/circular nesting
			// reconstructed.
			for _, b := range cycles {
				nodes = append(nodes, treeNode{id: b.ID, bean: b, depth: 1, hasKids: false})
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
// reports it as broken_links). Sorted via data.SortBeans (Status -> Priority
// -> Type -> Title) for determinism: idx.ByID iteration order is a Go map,
// so this must never be rendered unsorted (would break golden-test
// determinism and PO expectations alike). I03 (bean bt-7jr8 T8-review,
// closed E2 Task 4/bean bt-9ldr): this used to run its own ad-hoc
// title-only tie-break (sortByTitleThenID) because data.sortBeans was
// unexported at the time -- now that data.SortBeans is exported (E2 Task 1),
// the orphan bucket uses the SAME single-source order as every other bean
// list in the app instead of a second, parallel sort definition.
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
	data.SortBeans(out)
	return out
}

// reachableIDs returns every bean ID reachable via idx.Children, descending
// from roots. Structural, INDEPENDENT of the current expand/collapse UI
// state on purpose: collapsing a node in the tree must never make its
// subtree "disappear" from cycle detection -- only from the currently
// rendered rows (flattenTree's own expanded lookups handle that separately).
// A bean already visited stops the walk -- guards a cycle attached under a
// legitimately-reachable bean the same way appendBeanNode's per-path
// ancestors guard does, just as a global set here since only the full
// reachable SET is needed, not render order.
func reachableIDs(idx *data.Index, roots []*data.Bean) map[string]bool {
	seen := map[string]bool{}
	var walk func(b *data.Bean)
	walk = func(b *data.Bean) {
		if seen[b.ID] {
			return
		}
		seen[b.ID] = true
		for _, c := range idx.Children[b.ID] {
			walk(c)
		}
	}
	for _, b := range roots {
		walk(b)
	}
	return seen
}

// collectCycleOrphans returns every bean reachable from NEITHER a real root
// NOR a dangling-parent orphan's subtree -- beans trapped in a pure parent
// cycle (A -> B -> A), or hanging off one. Their own Parent DOES resolve (so
// collectOrphans misses them) and they can never be a Root() (Parent is
// non-empty), so without this sweep they are silently invisible. B02 (T8
// Opus quality review, bean bt-7jr8): beans-legal state -- frontmatter is
// hand-editable -- must never be dropped; swept into the same synthetic
// "(verwaist)" root as dangling-parent orphans instead.
func collectCycleOrphans(idx *data.Index, orphans []*data.Bean) []*data.Bean {
	anchors := append(append([]*data.Bean{}, idx.Roots()...), orphans...)
	reachable := reachableIDs(idx, anchors)

	var out []*data.Bean
	for id, b := range idx.ByID {
		if !reachable[id] {
			out = append(out, b)
		}
	}
	data.SortBeans(out)
	return out
}

// visibleNodes flattens the model's current idx+expanded state -- switches
// to the filtered flattening (flattenTreeFiltered) whenever search OR facets
// are narrowing the tree (m.treeActive(), box_filter_facets.go, E2 Task 4).
// m.beanMatches AND-combines Task 3's beanMatchesSearch with Task 4's
// beanMatchesFacets; both stay in place as the individual halves that
// combined predicate calls into (Backlog, E2 Task 5, reuses beanMatches
// unchanged).
func (m model) visibleNodes() []treeNode {
	if m.treeActive() {
		return flattenTreeFiltered(m.idx, m.expanded, m.beanMatches)
	}
	return flattenTree(m.idx, m.expanded)
}

// treeSearchActive reports whether a local/Bleve search query is currently
// narrowing the tree (E2 Task 3; Task 4 extends this with facet state via a
// combined treeActive()).
func (m model) treeSearchActive() bool {
	return strings.TrimSpace(m.searchQuery) != ""
}

// beanMatchesSearch is the search half of the combined filter predicate
// (Task 4 AND-combines this with facet criteria as beanMatches,
// box_filter_facets.go). Below the Bleve threshold (<3 chars), or while a
// Bleve response for THIS exact query hasn't arrived yet, it falls back to
// an immediate local title+ID substring match (case-insensitive) -- once
// searchBleveFor catches up to the current query, the Bleve result set
// becomes authoritative (design-spec.md §6 V2: "Bleve-Modus ab 3 Zeichen").
//
// I01 (E2-T3-Review finding, PFLICHT carried into bean bt-9ldr/Task 4): the
// authoritative-Bleve branch UNIONs the local ID-substring match back in
// rather than replacing it outright. Rationale: data.Client.Search runs a
// title+body full-text query (client.go doc comment) -- it does not
// necessarily index an arbitrary ID substring as a token, so a bean that
// only matched the query by ID would otherwise flicker OUT of the tree the
// instant the async Bleve response for the same query lands, even though
// the user's query still substring-matches its ID exactly as it did before.
func (m model) beanMatchesSearch(b *data.Bean) bool {
	q := strings.ToLower(strings.TrimSpace(m.searchQuery))
	if q == "" {
		return true
	}
	if len(q) >= 3 && m.searchBleveFor == m.searchQuery {
		return m.searchBleveIDs[b.ID] || strings.Contains(strings.ToLower(b.ID), q)
	}
	return strings.Contains(strings.ToLower(b.Title), q) || strings.Contains(strings.ToLower(b.ID), q)
}

// flattenTreeFiltered mirrors flattenTree's depth-first walk but only shows
// a subtree the caller has actually expanded (devd DD2-178 parity,
// view_browse_project.go:215-238, wörtlich portiert per plan Port-Referenzen:
// a manually collapsed ancestor stays collapsed even if a descendant
// matches -- it renders as ONE collapsed context row, never force-expanded).
// Every real root, the orphan bucket, AND the cycle-bean sweep go through
// the SAME predicate: an orphan/cycle bucket with zero matches is omitted
// entirely, matching entries render exactly like the unfiltered tree.
func flattenTreeFiltered(idx *data.Index, expanded map[string]bool, match func(*data.Bean) bool) []treeNode {
	if idx == nil {
		return nil
	}

	var nodes []treeNode
	ancestors := map[string]bool{}
	for _, b := range idx.Roots() {
		if ns, hit := filteredBeanNode(idx, b, 0, expanded, ancestors, match); hit {
			nodes = append(nodes, ns...)
		}
	}

	orphans := collectOrphans(idx)
	cycles := collectCycleOrphans(idx, orphans)
	var orphanNodes []treeNode
	for _, b := range orphans {
		if ns, hit := filteredBeanNode(idx, b, 1, expanded, ancestors, match); hit {
			orphanNodes = append(orphanNodes, ns...)
		}
	}
	var cycleHits []*data.Bean
	for _, b := range cycles {
		if match(b) {
			cycleHits = append(cycleHits, b)
		}
	}

	if len(orphanNodes) > 0 || len(cycleHits) > 0 {
		open := expanded[orphanRootID]
		nodes = append(nodes, treeNode{id: orphanRootID, depth: 0, hasKids: true, open: open, orphan: true})
		if open {
			nodes = append(nodes, orphanNodes...)
			for _, b := range cycleHits { // cycle beans render flat (depth 1), same as flattenTree
				nodes = append(nodes, treeNode{id: b.ID, bean: b, depth: 1, hasKids: false})
			}
		}
	}
	return nodes
}

// filteredBeanNode is flattenTreeFiltered's recursive per-bean step. hit
// reports whether b or anything reachable below it matches: match(b) itself,
// OR (if b is expanded) any expanded-visible child hit, OR (if b is
// collapsed) a structural, expand-state-INDEPENDENT scan of the whole
// subtree below (subtreeHasMatch) -- so a collapsed ancestor still renders
// as ONE collapsed context row when a match exists somewhere beneath it,
// without ever being force-expanded (devd DD2-178 parity, see
// flattenTreeFiltered's own doc comment). ancestors guards a Parent-cycle
// exactly like appendBeanNode's own defensive guard (only threaded through
// the EXPANDED/recursing branch, mirroring appendBeanNode).
func filteredBeanNode(idx *data.Index, b *data.Bean, depth int, expanded map[string]bool, ancestors map[string]bool, match func(*data.Bean) bool) ([]treeNode, bool) {
	if ancestors[b.ID] {
		return nil, false
	}
	children := idx.Children[b.ID]
	open := expanded[b.ID]
	self := match(b)

	if !open {
		if !self && !subtreeHasMatch(idx, children, b.ID, match) {
			return nil, false
		}
		return []treeNode{{id: b.ID, bean: b, depth: depth, hasKids: len(children) > 0, open: false}}, true
	}

	ancestors[b.ID] = true
	var childNodes []treeNode
	anyChildHit := false
	for _, c := range children {
		if ns, hit := filteredBeanNode(idx, c, depth+1, expanded, ancestors, match); hit {
			anyChildHit = true
			childNodes = append(childNodes, ns...)
		}
	}
	delete(ancestors, b.ID)

	if !self && !anyChildHit {
		return nil, false
	}
	nodes := append([]treeNode{{id: b.ID, bean: b, depth: depth, hasKids: len(children) > 0, open: true}}, childNodes...)
	return nodes, true
}

// subtreeHasMatch is a structural (expand-state-INDEPENDENT) scan used only
// for a COLLAPSED node (filteredBeanNode): whether match hits anywhere below
// it, so it still renders as a single collapsed context row without being
// force-expanded (devd DD2-178, view_browse_project.go:215-238's documented
// PO decision -- "a manually collapsed node stays collapsed even if it
// contains a match"). skip excludes the collapsed node itself (already
// checked by the caller as `self`); its own local visited set guards a
// Parent-cycle reachable under a collapsed node, independent of the caller's
// DFS-path ancestors set (different concern: this walks unconditionally
// downward, ignoring expand state entirely).
func subtreeHasMatch(idx *data.Index, beans []*data.Bean, skip string, match func(*data.Bean) bool) bool {
	seen := map[string]bool{skip: true}
	var walk func(b *data.Bean) bool
	walk = func(b *data.Bean) bool {
		if seen[b.ID] {
			return false
		}
		seen[b.ID] = true
		if match(b) {
			return true
		}
		for _, c := range idx.Children[b.ID] {
			if walk(c) {
				return true
			}
		}
		return false
	}
	for _, b := range beans {
		if walk(b) {
			return true
		}
	}
	return false
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

// treeCursorMove shifts the Tree cursor by delta (±1), clamped at both ends
// -- factored out of keyTree's own up/down case (update.go, E5 Task 4, bean
// bt-mne6) so the keyboard AND the wheel dispatch (mouse.go handleMouse,
// design decision f: Wheel moves the View-eigene CURSOR) share the exact
// same clamp logic instead of two independent copies.
func (m model) treeCursorMove(nodes []treeNode, delta int) model {
	if len(nodes) == 0 {
		return m
	}
	pos := m.cursorPos(nodes) + delta
	if pos < 0 {
		pos = 0
	}
	if pos > len(nodes)-1 {
		pos = len(nodes) - 1
	}
	m.cursorID = nodes[pos].id
	return m
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
// accent -- only the focused pane's cursor is highlighted (devd D03). bodyH
// is the pane's available ROW height (renderPane's h minus its own
// title+separator lines -- Golden Rule #1 still holds: windowing trims the
// ROWS handed to renderPane, it never forces a Height() on the bordered
// style itself). B01 (T8 Opus quality review): rows are windowed around the
// cursor (devd windowAround/windowStart port, view_browse_project.go:
// 647-670) so a tree taller than the pane never hides the cursor below the
// fold.
func (m model) treeRows(nodes []treeNode, focused bool, bodyH int) []string {
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
	return windowAround(rows, bodyH, pos)
}

// windowStart computes the window start index so cursor stays visible --
// centered when there's room, clamped at both edges (devd port:
// view_browse_project.go:649-661, itself shared there by render + the
// mouse-click Y->row mapping). Deterministic in cursor+height alone (no
// hidden state), which is what keeps it jitter-free: the same (n, height,
// cursor) always yields the same window, it never "remembers" a previous
// scroll offset.
func windowStart(n, height, cursor int) int {
	if height <= 0 || n <= height {
		return 0
	}
	start := cursor - height/2
	if start < 0 {
		start = 0
	}
	if start+height > n {
		start = n - height
	}
	return start
}

// windowAround windows rows to height entries so the cursor stays visible
// (devd port: view_browse_project.go:663-670). A no-op when everything
// already fits (including height<=0, an init/edge-case fallback -- renderPane
// itself caps to its own h regardless of how many rows it's handed).
func windowAround(rows []string, height, cursor int) []string {
	if height <= 0 || len(rows) <= height {
		return rows
	}
	start := windowStart(len(rows), height, cursor)
	return rows[start : start+height]
}

// renderDetailPane renders the Detail-Accordion (Meta/Body/Beziehungen/
// Historie, design-spec.md §6 V4) for the cursored bean -- E2 Task 2 wiring,
// replacing T8's title+meta-line placeholder. Resolves the tree cursor's
// bean (or nil for "no selection") and hands off to the shared
// renderBeanAccordionPane (below) -- E2-T6 fast-follow (T5-review): this and
// renderBacklogDetailPane (view_browse_backlog.go) used to hand-duplicate
// the ~20-line bodyW/accW/beanSections/renderAccordion/renderPane body;
// extracted once both call sites existed so a future accordion-pane change
// (e.g. scrolling) can't drift between Tree and Backlog.
func (m model) renderDetailPane(nodes []treeNode, w, h int, focused bool) string {
	pos := m.cursorPos(nodes)
	var b *data.Bean
	if pos >= 0 && pos < len(nodes) && !nodes[pos].orphan {
		b = nodes[pos].bean
	}
	return m.renderBeanAccordionPane(b, w, h, focused)
}

// renderBeanAccordionPane renders the Detail-Accordion pane (Meta/Body/
// Beziehungen/Historie, Task 1) for a single already-resolved bean, or the
// "(no selection)" placeholder for nil -- the shared body of
// renderDetailPane (Tree, above) and renderBacklogDetailPane (Backlog,
// view_browse_backlog.go), which differ only in HOW they resolve a bean from
// their own cursor shape (treeNode slice + orphan-guard vs. backlogList
// index into a []*data.Bean). w/w-2/w-4 mirror renderPane's own
// truncate(rows[i], w-2) content budget (render_shared.go) and
// renderAccordion's own PaddingLeft(2)-eats-into-Width(w) contract
// (accordion.go doc comment) -- bodyW = w-4 is NOT an arbitrary number.
//
// Scrolling when a section's content exceeds the pane height: no separate
// scroll-offset state is added here (out of plan scope) -- rows are simply
// split and handed to renderPane, whose existing Golden-Rule-#1 line cap
// (render_shared.go: `for i := 0; ... && len(lines) < h`) already prevents
// any overflow past the pane's border, the same mechanism the Tree pane
// relies on. A true scroll-with-indicator (mirroring Chrome()'s scrollView)
// is not needed yet: exclusive-open sections keep the open section's content
// near the top, and digit-jump/1-4 always re-opens from row 0.
func (m model) renderBeanAccordionPane(b *data.Bean, w, h int, focused bool) string {
	return renderAccordionPane(m.idx, b, w, h, m.accOpen, m.secCursor, m.fieldCursor, focused)
}

// renderAccordionPane (I02, E4-T3-Review PFLICHT carried into E4 Task 4,
// bean bt-yy6w) is the shared body renderBeanAccordionPane (above) and
// renderReviewDetailPane (Review-Cockpit, view_review_cockpit.go) both used
// to hand-duplicate (~15 lines: bodyW/accW clamps, beanSections,
// renderAccordion, renderPane). open/secCursor/fieldCursor/focused are
// PARAMETERS rather than read off the model, so this stays decoupled from
// BOTH call sites' own state shape: the Tree/Backlog's shared m.accOpen/
// m.secCursor/m.fieldCursor two-level detailFocus machine vs. the Cockpit's
// own single reviewAccOpen digit-jump cursor (design decision i) -- WHICH
// state backs the accordion stays entirely at each call site, mirroring
// I01's copy-on-write doctrine applied to "shared render body, independent
// state" instead of maps.
func renderAccordionPane(idx *data.Index, b *data.Bean, w, h, open, secCursor, fieldCursor int, focused bool) string {
	var rows []string
	if b != nil {
		bodyW := w - 4
		if bodyW < 1 {
			bodyW = 1
		}
		accW := w - 2
		if accW < 1 {
			accW = 1
		}
		secs := beanSections(idx, b, bodyW)
		acc := renderAccordion(secs, open, accW, focused, secCursor, fieldCursor)
		rows = strings.Split(acc, "\n")
	} else {
		rows = append(rows, theme.Dim.Render("(no selection)"))
	}
	return renderPane(pane{title: "Detail", rows: rows}, w, h, focused)
}

// searchShield is the search head row's glyph (U+2315 TELEPHONE RECORDER,
// port devd treeSearchLine view_browse_project.go:1099-1117 -- devd's own
// comment there notes it replaced an ambiguous-width lookalike, DD2-53; kept
// verbatim here for the same EAW-neutral reason).
const searchShield = "⌕"

// treeSearchLine renders the Tree pane's persistent search head row (design-
// spec.md §6 V2 "Such-/Filterkopf", port devd treeSearchLine
// view_browse_project.go:1099-1117): the live textinput while typing, the
// committed query AND/OR the active-facet summary in Red once either is
// active (DD2-53 "Filter aktiv" signal, E2 Task 4/bean bt-9ldr: extended to
// cover facets alongside search), or a muted hint when neither is active.
// The idle-hint text is deliberately UNCHANGED from Task 3 (no "f filter"
// addition) so the existing tree.golden fixture (no search/filter state)
// keeps rendering byte-identical -- only the active-state branches grew.
func (m model) treeSearchLine(w int) string {
	if m.searchActive {
		line := searchShield + " " + m.searchInput.View()
		if fs := m.filterSummary(); fs != "" {
			line += "  " + fs
		}
		return truncate(line, w)
	}
	if m.treeActive() {
		var parts []string
		if m.treeSearchActive() {
			q := m.searchQuery
			if m.searchBleveLoading {
				q += " …"
			}
			parts = append(parts, q)
		}
		if fs := m.filterSummary(); fs != "" {
			parts = append(parts, fs)
		}
		line := searchShield + " " + strings.Join(parts, "  ")
		return truncate(lipgloss.NewStyle().Foreground(theme.Red).Render(line), w)
	}
	return truncate(theme.Muted.Render(searchShield+" / search"), w)
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

// composeOverlays layers every floating overlay onto out in a fixed
// z-priority order (painter's algorithm, placeOverlay) -- E3 Task 1
// extraction (bean bt-dlgk): viewBrowseRepo and viewBacklog used to
// duplicate the filterOpen/confirmQuit block; a new overlay is now wired in
// exactly ONE place instead of two. Order: filter menu first, then the
// node-action overlay (design decision a2 -- EXACTLY one overlayID active
// at a time), then huh forms (T4/T5, m.form != nil is its own separate
// capture state, not part of the overlayID enum), quit-confirm LAST (it can
// interrupt anything, same as its pre-extraction position in both views).
func (m model) composeOverlays(out string, w, h int) string {
	if m.filterOpen {
		out = placeOverlay(out, m.treeFilterBox(), w, h)
	}
	switch m.overlay {
	case overlayValueMenu:
		out = placeOverlay(out, m.valueMenuBox(), w, h)
	case overlayTagPicker:
		out = placeOverlay(out, m.tagPickerBox(), w, h)
	case overlayParentPicker:
		out = placeOverlay(out, m.parentPickerBox(), w, h)
	case overlayBlockingPicker:
		out = placeOverlay(out, m.blockingPickerBox(), w, h)
	case overlayCreateConfirm:
		out = placeOverlay(out, m.createConfirmBox(), w, h)
	case overlayDeleteConfirm:
		out = placeOverlay(out, m.deleteBox(), w, h)
	}
	if m.form != nil {
		out = placeOverlay(out, m.formChrome(), w, h)
	}
	// E4 Task 1 (bean bt-jpgn): the Command-Center -- a floating overlay
	// like the ones above, painted BEFORE m.confirmQuit (quit stays the
	// topmost layer, unchanged precedent).
	if m.paletteOpen {
		out = placeOverlay(out, m.paletteBox(), w, h)
	}
	// E5 Task 2 (bean bt-wpn9): the Help-Overlay -- painted AFTER the
	// Command-Center, BEFORE m.confirmQuit (quit stays the topmost layer
	// under the modals, unchanged precedent, Painter's-Algorithmus-
	// Reihenfolge: spät = oben).
	if m.helpOpen {
		out = placeOverlay(out, m.helpBox(), w, h)
	}
	if m.confirmQuit {
		out = placeOverlay(out, m.quitBox(), w, h)
	}
	return out
}

// View dispatches on viewID (devd port convention: enum + switch in view).
// E2 Task 5 (bean bt-gzu6) adds the first sibling case (viewBacklog); later
// epics grow this switch further, never branches.
//
// E5 Task 1 (bean bt-6dts, design decision a, point 2): every sub-view's
// result is wrapped in m.renderToast(...) HERE, at the top level -- NOT
// inside composeOverlays (called by each of the three sub-views above).
// This is beans-tui's own deviation from devd's View()/viewComposite()
// split (there is no separate viewComposite() layer here, each sub-view
// already ends in its own `return m.composeOverlays(out, w, h)`) -- the
// Toast must float over composeOverlays' ENTIRE stack, including
// confirmQuit (devd's own "über ALLEM" contract, overlay_show_toast.go),
// so it cannot be just another composeOverlays case.
func (m model) View() string {
	var out string
	switch m.view {
	case viewBacklog:
		out = m.viewBacklog()
	case viewReviewCockpit:
		out = m.viewReviewCockpit()
	default:
		out = m.viewBrowseRepo()
	}
	return m.renderToast(out)
}

// browseRepoChrome builds viewBrowseRepo's own head/localKeys strings --
// factored out (E5 Task 4, bean bt-mne6) so treeClickRow (mouse.go) can
// reconstruct this view's IDENTICAL geometry via clickPaneGeometry without a
// second, independently maintained copy of this breadcrumb/footer
// construction (Golden-Rule-Drift-Schutz: one source instead of two that
// could drift apart, mirrors windowStart's own shared-geometry rationale).
func (m model) browseRepoChrome(innerW int) (head, localKeys string) {
	globalHint := renderBindings([]keybind.Binding{keys.Refresh, keys.Help, keys.Quit})
	head = breadcrumb(m.repoLabel(), "Browse", globalHint, innerW)
	localHint := renderBindings([]keybind.Binding{keys.Up, keys.Down, keys.Left, keys.Right, keys.Enter, keys.Search, keys.Refresh, keys.Status, keys.Create, keys.Delete, keys.Editor}) + "  tab:focus"
	localKeys = footer(localHint, innerW)
	return
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

	head, localKeys := m.browseRepoChrome(innerW)

	div := theme.Dim.Render(strings.Repeat("─", innerW))
	// I04 (T8 Opus quality review): a failed data.Watch start (no live
	// reload) must never degrade silently -- it goes into the same
	// indicator slot Chrome() uses for the scroll position (Accent, non-
	// critical), leaving m.err/the Red slot reserved for real load failures.
	indicator := ""
	if m.watchUnavailable {
		indicator = "watch unavailable — ctrl+r für manuelles Reload"
	}
	status := statusBar(indicator, m.err, innerW)

	// E5 Task 4 (bean bt-mne6): bodyH/lw/rw now come from the SAME
	// clickPaneGeometry helper treeClickRow (below) uses to map a click back
	// to a row -- single source for the numeric pane geometry, not just the
	// head/localKeys strings above (Golden-Rule-Drift-Schutz).
	bodyH, lw, rw, _, _ := clickPaneGeometry(w, h, head, localKeys)
	nodes := m.visibleNodes()
	// E2 Task 3 (bean bt-4ep2): the search head row is prepended to the Tree
	// pane's rows, costing 1 line of its bodyH-2 content budget -- the actual
	// tree rows window to bodyH-3 (one less than T8/E1's bodyH-2) so the
	// combined [searchLine, ...treeRows] slice still fits renderPane's own
	// Golden-Rule-#1 line cap.
	searchLine := m.treeSearchLine(lw - 2)
	treeRowsWithHead := append([]string{searchLine}, m.treeRows(nodes, !m.detailFocus, bodyH-3)...)
	treeBox := renderPane(pane{title: "Tree", rows: treeRowsWithHead}, lw, bodyH, !m.detailFocus)
	detailBox := m.renderDetailPane(nodes, rw, bodyH, m.detailFocus)
	body := lipgloss.JoinHorizontal(lipgloss.Top, treeBox, detailBox)

	content := head + "\n" + div + "\n" + body + "\n" + div + "\n" + localKeys + "\n" + status
	out := outerBorder(content, innerW, true)

	return m.composeOverlays(out, w, h)
}

// treeClickRow maps a mouse click to a Tree node index (E5 Task 4, bean
// bt-mne6, design decision f; caller: mouse.go's mouseTreeClick) --
// bodyH/lw/originX/originY are RECONSTRUCTED identically to viewBrowseRepo's
// own render formula (above) via the shared browseRepoChrome +
// clickPaneGeometry helpers (Golden-Rule-Drift-Schutz, Kommentar-Pflicht
// analog windowStart's own doc comment): if viewBrowseRepo's algebra ever
// changes, it changes HERE too automatically (single source), so a click can
// never silently land on the wrong row through independent drift. Row 0
// (clickRow==0, right below the pane's title+separator) is the search head
// line (treeSearchLine) -- never a node target. Row 1+ maps via
// windowStart(len(nodes), bodyH-3, cursorPos) + (clickRow-1), the SAME
// bodyH-3 window height treeRows itself windows to (treeRowsWithHead's own
// budget trade, above). ok=false for a click outside the Tree pane's column
// span, on/above the search line, or past the last actually-rendered row.
func treeClickRow(m model, nodes []treeNode, msg tea.MouseMsg) (idx int, ok bool) {
	w, h := m.width, m.height
	if w <= 0 {
		w = 80
	}
	if h <= 0 {
		h = 24
	}
	innerW := w - 2
	head, localKeys := m.browseRepoChrome(innerW)

	bodyH, lw, _, originX, originY := clickPaneGeometry(w, h, head, localKeys)

	if msg.X < originX || msg.X >= originX+lw {
		return 0, false // right Detail pane, or off-screen -- no Tree target
	}
	clickRow := msg.Y - originY
	if clickRow <= 0 {
		return 0, false // above the pane, or row 0 == the search head line
	}

	windowRows := bodyH - 3
	if windowRows < 0 {
		windowRows = 0
	}
	pos := m.cursorPos(nodes)
	start := windowStart(len(nodes), windowRows, pos)
	visible := windowRows
	if len(nodes)-start < visible {
		visible = len(nodes) - start
	}
	nodeWindowIdx := clickRow - 1
	if nodeWindowIdx < 0 || nodeWindowIdx >= visible {
		return 0, false
	}
	return start + nodeWindowIdx, true
}
