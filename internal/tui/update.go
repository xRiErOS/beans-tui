package tui

// update.go — the Update dispatcher (Elm architecture): routes tea.Msg to
// state transitions only. No rendering here (view_browse_repo.go), no
// Cmd-only definitions here (messages.go) -- port convention from devd
// update.go.

import (
	"errors"
	"strings"

	"beans-tui/internal/data"

	keybind "github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
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

	case searchBleveResultMsg:
		return m.applyBleveResult(msg), nil

	case mutationDoneMsg:
		// E3 (bean bt-dlgk): every Set*/Add*/Remove*/Delete mutation goes
		// through this ONE case -- applyMutationResult below is the shared
		// status-line + unconditional-reload tail (design decision d).
		return m.applyMutationResult(msg.err)

	case createDoneMsg:
		// E3 Task 4 (bean bt-y4ly): the ONE exception to mutationDoneMsg --
		// a success additionally jumps the cursor (applyCreateDone below), a
		// failure routes through the same applyMutationResult tail as every
		// other mutation.
		return m.applyCreateDone(msg)

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	// E3 Task 4 (bean bt-y4ly): huh-internal Msgs (cursor blink, viewport
	// ticks, its own nextField/nextGroup chain) that matched none of the
	// cases above -- the form's own Init()/Update() Cmds keep firing these
	// for as long as it's open. Key-Msgs never reach here: tea.KeyMsg has
	// its own case above, routed through handleKey -> keyForm when
	// m.form != nil (capture order doc-stamp, handleKey below).
	if m.form != nil {
		return m.updateForm(msg)
	}
	return m, nil
}

// applyMutationResult is the shared tail every E3 mutation's mutationDoneMsg
// runs through (design decision d, bean bt-dlgk): ALWAYS reloads
// (loadCmd), success or failure alike -- a success must show the new state,
// a failure (including a genuine ETag conflict) must resolve the now-stale
// index rather than leave the UI showing a value that silently failed to
// apply. errors.Is(err, data.ErrConflict) gets its own status-line wording;
// Toast is E5 scope, m.err/the status line's Red slot (view.go statusBar)
// is the interim feedback channel.
func (m model) applyMutationResult(err error) (tea.Model, tea.Cmd) {
	m.err = ""
	if err != nil {
		if errors.Is(err, data.ErrConflict) {
			m.err = "Konflikt: Bean extern geändert — neu geladen"
		} else {
			m.err = err.Error()
		}
	}
	return m, loadCmd(m.client)
}

// applyCreateDone handles a completed Create (E3 Task 4, bean bt-y4ly): a
// failure routes through the shared applyMutationResult tail exactly like
// every other mutation (status line + reload). A success additionally jumps
// the cursor onto the freshly minted bean and expands its parent's ancestor
// path BEFORE the reload fires: m.idx here is still the OLD index (the new
// bean does not exist in it yet), so expandAncestorsOf must walk UP FROM
// msg.bean.Parent (an EXISTING bean) rather than from msg.bean.ID itself (a
// lookup that would silently no-op against the old index) -- and since
// expandAncestorsOf only marks an id's ANCESTORS, not the id itself, the
// immediate parent needs its own explicit exp[...]=true so the new bean
// (its child) is revealed once flattenTree runs against it. applyLoaded's
// existing "exact bean still present" cursor-restore path (above) then finds
// msg.bean.ID unchanged after the beansLoadedMsg arrives (US-10-parity, no
// new cursor logic). A parentless new bean (msg.bean.Parent == "") needs no
// expansion at all -- idx.Roots() already surfaces every parentless bean.
func (m model) applyCreateDone(msg createDoneMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		return m.applyMutationResult(msg.err)
	}
	if msg.bean.Parent != "" {
		m.expanded = expandAncestorsOf(m.idx, m.expanded, msg.bean.Parent)
		exp := cloneBoolMap(m.expanded) // I01: copy-on-write
		exp[msg.bean.Parent] = true     // expandAncestorsOf expands ancestors OF id, not id itself
		m.expanded = exp
	}
	m.cursorID = msg.bean.ID
	m.err = ""
	return m, loadCmd(m.client)
}

// beanETag reads id's CURRENT etag straight from the live index (design
// decision d, bean bt-dlgk) -- NEVER a copy captured when an overlay opened,
// so a watch-reload between open and submit is automatically honored and a
// "real" ETag conflict is only ever the narrow Submit<->Disk race window.
// ok=false means id is no longer in the index (deleted by another agent, or
// this reload's own applyLoaded already dropped it) -- callers must close
// their overlay and surface a status-line note instead of firing a doomed
// mutation against a bean that no longer exists.
func (m model) beanETag(id string) (etag string, ok bool) {
	if m.idx == nil {
		return "", false
	}
	b, ok := m.idx.ByID[id]
	if !ok {
		return "", false
	}
	return b.ETag, true
}

// keyNodeAction routes the node-focused mutation keys (design-spec §7:
// s/t/a/B/c/d/e -- Status/TagAssign/Assign/Blocking/Create/Delete/Editor).
// Placed in handleKey directly after the Refresh check and BEFORE the
// m.detailFocus dispatch, since node actions must act on m.focusedBean()
// regardless of which pane currently has focus (focusedBean() already
// covers Tree, Backlog AND detail focus, the same view-agnostic dispatcher
// E2 Task 2 built) -- routing through keyDetailFocus/keyTree/keyBacklog
// first would either shadow these keys entirely or require duplicating the
// dispatch in three places. Create is the ONE key that works without a
// focused bean (an empty repo can still be seeded); every other key here is
// a handled-but-silent no-op with no focused bean (design decision, plan
// »Task 1« Step 10: a user-facing "not yet" text would be throwaway work,
// the remaining keys land in the immediately following tasks T2-T6).
// Returns handled=false for every other key so callers fall through to the
// view-specific dispatch.
func (m model) keyNodeAction(msg tea.KeyMsg) (bool, tea.Model, tea.Cmd) {
	switch {
	case keybind.Matches(msg, keys.Create):
		// E3 Task 4 (bean bt-y4ly): Create works WITHOUT a focused bean (an
		// empty repo can still be seeded) -- the ONLY key here that skips
		// the focusedBean() guard below.
		nm, cmd := m.openCreateForm()
		return true, nm, cmd
	case keybind.Matches(msg, keys.Status),
		keybind.Matches(msg, keys.TagAssign),
		keybind.Matches(msg, keys.Assign),
		keybind.Matches(msg, keys.Blocking),
		keybind.Matches(msg, keys.Delete),
		keybind.Matches(msg, keys.Editor):
		if m.focusedBean() == nil {
			return true, m, nil // handled, silent no-op -- no node to act on
		}
		if keybind.Matches(msg, keys.Status) {
			return true, m.openValueMenu(), nil
		}
		if keybind.Matches(msg, keys.TagAssign) {
			return true, m.openTagPicker(), nil
		}
		if keybind.Matches(msg, keys.Assign) {
			return true, m.openParentPicker(), nil
		}
		if keybind.Matches(msg, keys.Blocking) {
			return true, m.openBlockingPicker(), nil
		}
		return true, m, nil // stub: T5 (Editor), T6 (Delete)
	}
	return false, m, nil
}

// keyOverlay routes to the currently open node-action overlay's own key
// handler (design decision a2: exactly one overlayID is active at a time).
// Checked in handleKey ahead of the global ctrl+c/q/tab switch (same
// full-capture precedent as m.searchActive/m.filterOpen just above it) so
// e.g. "q" cannot leak through to a quit-request while a menu/picker is
// open. T1 wires overlayValueMenu; T2/T3/T4/T6 add their cases as their
// overlays land.
func (m model) keyOverlay(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.overlay {
	case overlayValueMenu:
		return m.keyValueMenu(msg)
	case overlayTagPicker:
		return m.keyTagPicker(msg)
	case overlayParentPicker:
		return m.keyParentPicker(msg)
	case overlayBlockingPicker:
		return m.keyBlockingPicker(msg)
	case overlayCreateConfirm:
		return m.keyCreateConfirm(msg)
	}
	return m, nil
}

// applyBleveResult applies an async Bleve search result (E2 Task 3, bean
// bt-4ep2) -- discarded if the query has moved on since the request was
// dispatched (staleness guard, chosen over a debounce timer: every
// qualifying keystroke dispatches its own beans-CLI subprocess, but only the
// response tagged for the CURRENT searchQuery is ever applied; see
// messages.go's searchBleveResultMsg doc comment for the full rationale).
func (m model) applyBleveResult(msg searchBleveResultMsg) model {
	if msg.query != m.searchQuery {
		return m // stale -- searchQuery has moved on since this request was sent
	}
	m.searchBleveLoading = false
	if msg.err != nil {
		m.err = msg.err.Error()
		return m
	}
	ids := make(map[string]bool, len(msg.ids))
	for _, id := range msg.ids {
		ids[id] = true
	}
	m.searchBleveIDs = ids
	m.searchBleveFor = msg.query
	return m
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

	// E2 Task 3 (bean bt-4ep2, design-spec.md §7: "Single-Keys inaktiv bei
	// fokussierter Texteingabe"): the search input captures EVERY key except
	// enter/esc while focused -- typing "q" or "tab" must produce that
	// character in the search box, not quit/focus-swap. Checked before the
	// ctrl+c/q/tab switch below, mirroring how m.confirmQuit's modal already
	// captures input ahead of that same switch (keyConfirmQuit does not
	// special-case ctrl+c either -- same existing precedent, applied here for
	// consistency rather than introducing a second, partial-capture pattern).
	if m.searchActive {
		return m.keySearchInput(msg)
	}

	// E2 Task 4 (bean bt-9ldr): the facet-filter menu is a floating overlay
	// that fully captures input, same precedent as m.searchActive just above
	// (and m.confirmQuit at the very top of this function) -- single-key
	// shortcuts like `q`/`tab` must not leak through to the Tree underneath
	// while the menu is open.
	if m.filterOpen {
		return m.keyFilterMenu(msg)
	}

	// E3 Task 4 (bean bt-y4ly): an open huh Form fully captures input too,
	// same precedent as filterOpen/searchActive/confirmQuit above -- checked
	// BEFORE the node-action overlay switch below since m.form != nil is its
	// own separate capture state, not one of the overlayID cases (design
	// decision a2: "huh forms are not a menu overlay").
	if m.form != nil {
		return m.keyForm(msg)
	}

	// E3 (bean bt-dlgk): the open node-action overlay (Value-Menü T1, Tag-/
	// Parent-/Blocking-Picker T2/T3, Create-/Delete-Confirm T4/T6) fully
	// captures input, same precedent as m.filterOpen/m.searchActive just
	// above -- checked ahead of the ctrl+c/q/tab switch so those single keys
	// cannot leak through while a menu/picker is open.
	if m.overlay != overlayNone {
		return m.keyOverlay(msg)
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

	// E3 (bean bt-dlgk): node-focused mutation keys (s/t/a/B/c/d/e) --
	// checked BEFORE the detailFocus/keyBacklog/keyTree dispatch below so
	// they act on m.focusedBean() regardless of which pane has focus (see
	// keyNodeAction's own doc comment for the full rationale).
	if handled, nm, cmd := m.keyNodeAction(msg); handled {
		return nm, cmd
	}

	if m.detailFocus {
		return m.keyDetailFocus(msg)
	}
	if m.view == viewBacklog {
		return m.keyBacklog(msg)
	}
	return m.keyTree(msg)
}

// focusedBean returns the bean the Detail-Accordion currently targets,
// independent of which view is active (devd port focusedIssue,
// view_detail_issue.go:20-35) -- view-agnostic so Task 5's Backlog view
// (viewBacklog case below, bean bt-gzu6) reuses keyDetailFocus verbatim.
func (m model) focusedBean() *data.Bean {
	switch m.view {
	case viewBacklog: // E2 Task 5: the Backlog list's own selection, NOT the (possibly stale/irrelevant) tree cursor
		return m.backlogSelected()
	default: // viewBrowseRepo (T8)
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

// openSearchInput enters the search-input capture state (E2 Task 3, bean
// bt-4ep2): pre-loads the input with the CURRENT query (re-opening a
// committed search resumes editing it, mirrors devd's SetValue(m.treeQuery)).
// Factored out of keyTree (E2 Task 5, bean bt-gzu6): keyBacklog
// (view_browse_backlog.go) shares this verbatim -- the search head/live
// filter is ONE piece of shared model state, not a per-view concept.
func (m model) openSearchInput() (tea.Model, tea.Cmd) {
	m.searchActive = true
	m.searchInput.SetValue(m.searchQuery)
	m.searchInput.CursorEnd()
	m.searchInput.Focus()
	return m, textinput.Blink
}

// openFilterMenu opens the shared facet-filter menu (E2 Task 4, bean
// bt-9ldr), port devd openTreeFilter minus the loadAllIssues call. Factored
// out of keyTree (E2 Task 5): keyBacklog (view_browse_backlog.go) shares this
// verbatim -- ONE filter menu for Tree AND Backlog (design-spec.md US-05,
// box_filter_facets.go doc comment).
func (m model) openFilterMenu() (tea.Model, tea.Cmd) {
	m.filterItems = m.buildFilterItems()
	m.filterMenu = listState{}
	m.filterMenu.setLen(len(m.filterItems))
	m.filterOpen = true
	return m, nil
}

// keyTree drives the tree: up/down move the cursor, right/left expand/
// collapse, enter toggles expand (no-op on a leaf, per task scope). `/`, `f`,
// `X`, and the esc-cascade's search/filter-clearing rung (E2 Task 3+4) are
// checked FIRST, ahead of the len(nodes)==0 short-circuit below -- opening
// the search box/filter menu or clearing an active query/facet must work
// even on an empty/pre-load tree.
func (m model) keyTree(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case keybind.Matches(msg, keys.Search):
		// Port devd keyTree's Search case (view_browse_project.go:689-699),
		// minus loadAllIssues -- beans-tui already holds the full Index in
		// memory (no separate "load everything for search" step needed).
		return m.openSearchInput()
	case keybind.Matches(msg, keys.Filter):
		// E2 Task 4 (bean bt-9ldr), port devd openTreeFilter
		// (view_browse_project.go:865-874) minus the loadAllIssues call --
		// beans-tui already holds the full Index in memory.
		return m.openFilterMenu()
	case keybind.Matches(msg, keys.Backlog):
		// E2 Task 5 (bean bt-gzu6): `b` opens the Backlog view -- search/
		// filter state carries over unchanged (ONE shared m.beanMatches
		// predicate, Task 3/4), only the master list's data source and
		// cursor representation (backlogList) differ from the Tree.
		m.view = viewBacklog
		m.backlogList.setLen(len(m.backlogVisible()))
		return m, nil
	case keybind.Matches(msg, keys.FilterClear):
		// X as a direct top-level reset even with the menu closed (design-
		// spec.md, mirrors devd's esc-cascade also clearing filters below) --
		// wired to the SAME clearFacets() helper keyFilterMenu's own X-case
		// uses (box_filter_facets.go).
		m = m.clearFacets()
		return m.resetCursorToFirstVisible(), nil
	case keybind.Matches(msg, keys.Back):
		if m.treeActive() { // esc-cascade Rung 2: committed query AND/OR active facets -> clear both in one step (Task 4 generalizes Task 3's search-only rung, devd parity view_browse_project.go:725-736)
			m.searchQuery = ""
			m.searchBleveIDs = nil
			m.searchBleveFor = ""
			m = m.clearFacets()
			return m.resetCursorToFirstVisible(), nil
		}
		// Rung 3 (existing behavior): no-op -- no Lobby fallback in E2
		// (design-spec.md §12 E5, plan Task 3 Port-Referenzen: the cascade
		// ends here, a documented Scope-Cut rather than a TBD).
		return m, nil
	}

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

// keySearchInput drives the active search textinput (E2 Task 3, bean
// bt-4ep2, port devd keyTreeSearch view_browse_project.go:1073-1097): every
// keystroke updates m.searchQuery LIVE (immediate local filter, not just on
// commit) and resets the cursor to the freshly filtered list's first row
// (mirrors devd's unconditional treeCursor=0). enter commits (blurs the
// input; the filter/query itself is NOT cleared). esc cancels AND clears the
// query entirely (esc-cascade Rung 1 -- distinct from enter). Both also
// dispatch an async Bleve search once the query is due (maybeBleveCmd).
func (m model) keySearchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		m.searchActive = false
		m.searchInput.Blur()
		m.searchQuery = strings.TrimSpace(m.searchInput.Value())
		m = m.resetCursorToFirstVisible()
		return m.dispatchBleveIfDue(nil)
	case tea.KeyEsc:
		m.searchActive = false
		m.searchInput.Blur()
		m.searchInput.SetValue("")
		m.searchQuery = ""
		return m.resetCursorToFirstVisible(), nil
	}

	var inputCmd tea.Cmd
	m.searchInput, inputCmd = m.searchInput.Update(msg)
	m.searchQuery = strings.TrimSpace(m.searchInput.Value())
	m = m.resetCursorToFirstVisible()
	return m.dispatchBleveIfDue(inputCmd)
}

// resetCursorToFirstVisible parks the cursor on the (newly filtered) node
// list's first row -- called on every search keystroke/clear, mirrors devd's
// unconditional `m.treeCursor = 0` (view_browse_project.go:1073-1097): a
// bean ID, not an index, is beans-tui's own cursor representation (T8), so
// the equivalent reset is "first visible node's ID" rather than a bare 0.
func (m model) resetCursorToFirstVisible() model {
	nodes := m.visibleNodes()
	if len(nodes) == 0 {
		m.cursorID = ""
		return m
	}
	m.cursorID = nodes[0].id
	return m
}

// dispatchBleveIfDue is keySearchInput's shared tail: batches an extra Cmd
// (the textinput's own Update cmd, or nil) together with maybeBleveCmd()'s
// dispatch when one is due, flagging searchBleveLoading so the search head
// (treeSearchLine) can surface it.
func (m model) dispatchBleveIfDue(extra tea.Cmd) (tea.Model, tea.Cmd) {
	bleve := m.maybeBleveCmd()
	if bleve == nil {
		return m, extra
	}
	m.searchBleveLoading = true
	return m, tea.Batch(extra, bleve)
}

// maybeBleveCmd returns a searchCmd dispatch when the current searchQuery
// has reached the Bleve threshold (>=3 chars, design-spec.md §6 V2) and
// differs from the query the current searchBleveIDs answer (searchBleveFor)
// -- nil below the threshold, or when nothing has changed since the last
// dispatch. Deliberately NOT debounced (E2 Task 3 commit rationale, bean
// bt-4ep2): the plan's own design (and the bean's Akzeptanz criteria) choose
// a staleness guard on the RESPONSE (applyBleveResult) over delaying the
// REQUEST with a timer -- every qualifying keystroke may dispatch its own
// beans-CLI subprocess, but only the freshest response is ever applied.
func (m model) maybeBleveCmd() tea.Cmd {
	q := m.searchQuery
	if len(q) < 3 || q == m.searchBleveFor {
		return nil
	}
	return searchCmd(m.client, q)
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
