package tui

// view_review_cockpit.go — V6 Review-Cockpit (design-spec.md §6 V6, E4 Task
// 3, bean bt-hxyo): the PO's Merge-Gate view. Queue = idx.WithTag("to-review")
// (E1 Task 3) grouped by nearest EpicAncestor (data package, design decision
// c), a flat "(kein Epic)" bucket ALWAYS last; a second, flat "Rework"
// section (idx.WithTag("rework")) for PO awareness of already-rejected beans
// awaiting agent nacharbeit. Master-Detail layout mirrors viewBrowseRepo/
// viewBacklog's algebra exactly (masterDetailWidths/renderPane/outerBorder/
// composeOverlays, E1/E2) -- NO new pane math. Verdikt-actions (a/x/o) land
// in Task 4 (bean bt-yy6w); this task ships navigation + the read-only
// detail preview only (design decision i: OWN reviewAccOpen, no Tree/
// Backlog focus-machine reuse).
//
// Port references: EpicAncestor's cycle-guarded walk mirrors
// expandAncestorsOf (update.go) / CollectDescendants (data/hierarchy.go) --
// NOT devd (devd has no Epic-Ancestor search, Sprints are flat). The
// Master-Detail layout IDEA (Queue left with a Verdikt indicator, read-only
// Accordion preview right) references devd view_review_sprint.go's
// viewReviewSprint/reviewMasterPane/reviewDetailPane -- LAYOUT PATTERN ONLY,
// no Sprint-entity coupling (beans-tui has none, design-spec.md §4). The
// "(keine offenen Reviews)" empty-state placeholder mirrors devd
// view_navigate_reviews.go:20's own text pattern.

import (
	"fmt"
	"strings"

	"beans-tui/internal/data"
	"beans-tui/internal/theme"
	keybind "github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// reviewGroup is one epic-anchored (or the single "(kein Epic)") bucket of
// the to-review queue. epic == nil marks the "(kein Epic)" bucket -- see
// reviewQueue's own doc comment for when that fires.
type reviewGroup struct {
	epic  *data.Bean // nil for the "(kein Epic)" bucket
	beans []*data.Bean
}

// reviewQueue groups idx.WithTag("to-review") by data.EpicAncestor (design
// decision c): one reviewGroup per DISTINCT epic that has at least one
// to-review descendant (the epics themselves canonically sorted via
// data.SortBeans, each group's own members ALSO canonically sorted), plus a
// "(kein Epic)" bucket (epic == nil) for every to-review bean with no epic
// ancestor -- ALWAYS rendered last, and only when non-empty (there is no
// group for an epic with zero to-review children, symmetric with there being
// no "(kein Epic)" group when every to-review bean resolves to an epic).
// Returns nil when idx is nil or nothing carries the to-review tag.
func reviewQueue(idx *data.Index) []reviewGroup {
	if idx == nil {
		return nil
	}
	toReview := idx.WithTag("to-review")
	if len(toReview) == 0 {
		return nil
	}

	byEpic := map[string][]*data.Bean{}
	seenEpic := map[string]bool{}
	var epics []*data.Bean
	var noEpic []*data.Bean

	for _, b := range toReview {
		epic, ok := data.EpicAncestor(idx, b)
		if !ok {
			noEpic = append(noEpic, b)
			continue
		}
		byEpic[epic.ID] = append(byEpic[epic.ID], b)
		if !seenEpic[epic.ID] {
			seenEpic[epic.ID] = true
			epics = append(epics, epic)
		}
	}

	data.SortBeans(epics)

	groups := make([]reviewGroup, 0, len(epics)+1)
	for _, epic := range epics {
		members := byEpic[epic.ID]
		data.SortBeans(members)
		groups = append(groups, reviewGroup{epic: epic, beans: members})
	}
	if len(noEpic) > 0 {
		data.SortBeans(noEpic)
		groups = append(groups, reviewGroup{epic: nil, beans: noEpic})
	}
	return groups
}

// reviewRework returns idx.WithTag("rework"), flat, canonically sorted --
// deliberately NOT epic-grouped (design decision c: a secondary awareness
// list for beans already rejected and awaiting agent nacharbeit, not the
// primary reviewed queue). idx.WithTag already sorts via data.SortBeans, so
// this is a thin, nil-safe pass-through.
func reviewRework(idx *data.Index) []*data.Bean {
	if idx == nil {
		return nil
	}
	return idx.WithTag("rework")
}

// reviewFlat is the Cockpit's cursor index space (types.go's reviewCursor
// doc-stamp): every to-review bean in reviewQueue's group order, THEN every
// reviewRework bean. Group headers are a render-time-only concern
// (reviewQueueRows below) -- never part of this index space, so up/down can
// never land on a non-actionable header row.
func reviewFlat(idx *data.Index) []*data.Bean {
	var flat []*data.Bean
	for _, g := range reviewQueue(idx) {
		flat = append(flat, g.beans...)
	}
	flat = append(flat, reviewRework(idx)...)
	return flat
}

// reviewSummaryLine renders the Cockpit's Summary-Zeile (design-spec.md §6
// V6, design decision j): "x of n" -- a 1-based cursor position within the
// LIVE to-review-only flat (n = the to-review portion of reviewFlat, NOT a
// Pass/Reject tally -- there is no review_status field to tally, design
// decision j) -- while the cursor sits on a to-review item; "Rework: <n>
// offen" instead while the cursor sits on a Rework item (no meaningful "x of
// n" for the awareness-only section). idx is read directly on every call
// (LIVE derivation, no cached n): an agent tagging more beans to-review while
// the PO is mid-review simply grows n on the next render, never a stale
// count.
func reviewSummaryLine(idx *data.Index, cursor int) string {
	toReviewCount := 0
	for _, g := range reviewQueue(idx) {
		toReviewCount += len(g.beans)
	}
	if cursor < toReviewCount {
		return fmt.Sprintf("%d of %d", cursor+1, toReviewCount)
	}
	return fmt.Sprintf("Rework: %d offen", len(reviewRework(idx)))
}

// reviewDot renders the Verdikt-Dot: Peach = pending (to-review, unverdikted
// by construction -- design decision j, a "passed" item leaves every visible
// section entirely), Red = rework (rejected, awaiting agent nacharbeit).
// Reduced to these 2 states vs. devd keys_review.go's 3 (Grün/Rot/Peach) --
// there is no "passed" dot here, a passed bean simply isn't in the Cockpit
// anymore (design decision c/j).
func reviewDot(inRework bool) string {
	if inRework {
		return lipgloss.NewStyle().Foreground(theme.Red).Render("●")
	}
	return lipgloss.NewStyle().Foreground(theme.Peach).Render("●")
}

// reviewQueueRows renders every visible Cockpit row (epic headers muted,
// "(kein Epic)"/"── Rework ──" separators dim, bean rows dot+relationRow) to
// a single windowed slice -- the SAME D08 cursor treatment as treeRows/
// backlogRows (view_browse_repo.go/view_browse_backlog.go): a leading `▌`
// bar plus the whole row accent-tinted at the cursorFlat position (which
// indexes into reviewFlat's bean-only space, NOT this function's row slice --
// header/separator rows are render-only and never receive the cursor
// treatment). Windowed around the CURSOR ROW (not cursorFlat directly, since
// headers shift row indices) via the existing windowAround (E1 Task 8).
func reviewQueueRows(idx *data.Index, cursorFlat int, focused bool, bodyH int) []string {
	var rows []string
	flatIdx := 0
	cursorRow := 0

	appendBeanRow := func(b *data.Bean, inRework bool) {
		text := reviewDot(inRework) + " " + relationRow(b)
		if flatIdx == cursorFlat {
			cursorRow = len(rows)
			plain := ansi.Strip(text)
			if focused {
				rows = append(rows, theme.Accent.Render("▌"+plain))
			} else {
				rows = append(rows, theme.Dim.Render("▌"+plain))
			}
		} else {
			rows = append(rows, " "+text)
		}
		flatIdx++
	}

	for _, g := range reviewQueue(idx) {
		if g.epic != nil {
			rows = append(rows, theme.Muted.Render(relationRow(g.epic)))
		} else {
			rows = append(rows, theme.Dim.Render("(kein Epic)"))
		}
		for _, b := range g.beans {
			appendBeanRow(b, false)
		}
	}

	if rework := reviewRework(idx); len(rework) > 0 {
		rows = append(rows, theme.Dim.Render("── Rework ──"))
		for _, b := range rework {
			appendBeanRow(b, true)
		}
	}

	return windowAround(rows, bodyH, cursorRow)
}

// renderReviewDetailPane renders the read-only Detail-Accordion preview for
// the Cockpit's cursored bean, using the Cockpit's OWN reviewAccOpen digit-
// jump cursor (design decision i) -- deliberately NOT the shared
// renderBeanAccordionPane (view_browse_repo.go), which hard-reads
// m.accOpen/m.secCursor/m.fieldCursor: those belong to the Tree/Backlog's
// two-level detailFocus machine, which the always-read-only Cockpit preview
// does not have (no field-level relation-jump here). Reuses the same
// underlying primitives (beanSections/renderAccordion/renderPane, E1/E2) --
// only the field-coupling is deliberately NOT shared, per design decision i's
// own rationale (mirrors I01's copy-on-write doctrine, applied to shared vs.
// own fields instead of maps).
func (m model) renderReviewDetailPane(b *data.Bean, w, h int) string {
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
		secs := beanSections(m.idx, b, bodyW)
		activeIdx := m.reviewAccOpen - 1
		if activeIdx < 0 {
			activeIdx = 0
		}
		acc := renderAccordion(secs, m.reviewAccOpen, accW, true, activeIdx, 0)
		rows = strings.Split(acc, "\n")
	} else {
		rows = append(rows, theme.Dim.Render("(no selection)"))
	}
	return renderPane(pane{title: "Detail", rows: rows}, w, h, true)
}

// openReviewCockpit enters the Cockpit (`R`, keyTree/keyBacklog below):
// always resets reviewCursor/reviewAccOpen to 0 -- a stale cursor/open-
// section from a PREVIOUS Cockpit visit (possibly against a now-different
// flat shape) must never leak into this one, same precedent as `tab`'s own
// enterDetailFocus-equivalent reset (handleKey, update.go).
func (m model) openReviewCockpit() (tea.Model, tea.Cmd) {
	m.view = viewReviewCockpit
	m.reviewCursor = 0
	m.reviewAccOpen = 0
	return m, nil
}

// keyReviewCockpit fully captures the Review-Cockpit (design decision h) --
// a/x/o land in Task 4 (bean bt-yy6w); this task wires navigation only
// (up/down, n/p aliases, digit accordion-jump, esc/q back to Browse).
// Unmatched keys (including a/x/o for now) are a silent no-op -- the SAME
// "handled-but-stub" convention as E3 Task 1's stub keys, not a "coming
// soon" text (throwaway work, Task 4 follows immediately in the epic
// sequence).
func (m model) keyReviewCockpit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	flat := reviewFlat(m.idx)

	switch {
	case keybind.Matches(msg, keys.Back), msg.String() == "q":
		m.view = viewBrowseRepo
		return m, nil
	}

	if s := msg.String(); len(s) == 1 && s[0] >= '1' && s[0] <= '4' {
		d := int(s[0] - '0')
		if m.reviewAccOpen == d {
			m.reviewAccOpen = 0
		} else {
			m.reviewAccOpen = d
		}
		return m, nil
	}

	switch navKey(msg.String()) {
	case "up":
		if m.reviewCursor > 0 {
			m.reviewCursor--
		}
		return m, nil
	case "down":
		if m.reviewCursor < len(flat)-1 {
			m.reviewCursor++
		}
		return m, nil
	}

	switch msg.String() {
	case "n": // explicit next (design-spec §7), alias of down
		if m.reviewCursor < len(flat)-1 {
			m.reviewCursor++
		}
	case "p": // explicit prev, alias of up
		if m.reviewCursor > 0 {
			m.reviewCursor--
		}
	}
	return m, nil // a/x/o -- Task 4 (bean bt-yy6w)
}

// viewReviewCockpit renders the two-pane master-detail Review-Cockpit --
// mirrors viewBrowseRepo's/viewBacklog's algebra exactly (Golden Rule #1: no
// Height() on a bordered style) so the frame always fills width x height,
// just with reviewQueueRows on the left instead of the Tree's/Backlog's own
// row source, and renderReviewDetailPane (design decision i) on the right
// instead of the shared renderBeanAccordionPane.
func (m model) viewReviewCockpit() string {
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
	head := breadcrumb(m.repoLabel(), "Review-Cockpit", globalHint, innerW)

	localHint := renderBindings([]keybind.Binding{keys.Up, keys.Down, keys.Enter, keys.Back, keys.Section}) + "  n/p:next/prev"
	localKeys := footer(localHint, innerW)

	div := theme.Dim.Render(strings.Repeat("─", innerW))
	indicator := ""
	if m.watchUnavailable {
		indicator = "watch unavailable — ctrl+r für manuelles Reload"
	}
	status := statusBar(indicator, m.err, innerW)

	footH := lipgloss.Height(localKeys) + 2
	avail := innerH - lipgloss.Height(head) - footH - 1
	if avail < 4 {
		avail = 18 // height unknown (init/tests) -> generous fallback, mirrors Chrome()/viewBrowseRepo
	}
	bodyH := avail - 2 // both panes add their own border (+2, Golden Rule #1)
	if bodyH < 1 {
		bodyH = 1
	}

	lw, rw := masterDetailWidths(innerW, 24)
	flat := reviewFlat(m.idx)

	var listBox, detailBox string
	if len(flat) == 0 {
		rows := []string{theme.Dim.Render("(keine offenen Reviews)")}
		listBox = renderPane(pane{title: "Review-Queue", rows: rows}, lw, bodyH, true)
		detailBox = m.renderReviewDetailPane(nil, rw, bodyH)
	} else {
		// Defensive clamp: reviewCursor is rendered against the LIVE flat,
		// which may have shrunk (e.g. a watch-reload dropped a bean) since
		// the last keypress resynced it -- keyReviewCockpit itself always
		// clamps on ITS OWN reviewFlat snapshot, but a render can land
		// in-between two keystrokes against a narrower one.
		cursor := m.reviewCursor
		if cursor >= len(flat) {
			cursor = len(flat) - 1
		}
		if cursor < 0 {
			cursor = 0
		}
		// Same 1-header-row budget trade as the Tree's search head / Backlog
		// (E2 Task 3/5): the summary line costs 1 line of the list pane's
		// bodyH-2 content budget, so the actual rows window to bodyH-3.
		summaryLine := truncate(reviewSummaryLine(m.idx, cursor), lw-2)
		rows := append([]string{summaryLine}, reviewQueueRows(m.idx, cursor, true, bodyH-3)...)
		listBox = renderPane(pane{title: "Review-Queue", rows: rows}, lw, bodyH, true)
		detailBox = m.renderReviewDetailPane(flat[cursor], rw, bodyH)
	}
	body := lipgloss.JoinHorizontal(lipgloss.Top, listBox, detailBox)

	content := head + "\n" + div + "\n" + body + "\n" + div + "\n" + localKeys + "\n" + status
	out := outerBorder(content, innerW, true)

	return m.composeOverlays(out, w, h)
}
