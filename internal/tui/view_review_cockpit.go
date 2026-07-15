package tui

// view_review_cockpit.go — V6 Review-Cockpit (design-spec.md §6 V6, E4 Task
// 3, bean bt-hxyo): the PO's Merge-Gate view. Queue = idx.WithTag("to-review")
// (E1 Task 3) grouped by nearest EpicAncestor (data package, design decision
// c), a flat "(kein Epic)" bucket ALWAYS last; a second, flat "Rework"
// section (idx.WithTag("rework")) for PO awareness of already-rejected beans
// awaiting agent nacharbeit. Master-Detail layout mirrors viewBrowseRepo/
// viewBacklog's algebra exactly (masterDetailWidths/renderPane/outerBorder/
// composeOverlays, E1/E2) -- NO new pane math. E4 Task 4 (bean bt-yy6w) wires
// the Verdikt-actions (a pass / x reject / o reopen, design-spec.md §5) into
// keyReviewCockpit below, plus two E4-T3-Review PFLICHT carry-overs: I01
// (reviewState, a single per-render/per-keypress idx.WithTag/EpicAncestor
// derivation instead of the 3-4x re-walk T3 left behind) and I02
// (renderReviewDetailPane now delegates to the SHARED renderAccordionPane,
// view_browse_repo.go, instead of hand-duplicating renderBeanAccordionPane's
// body).
//
// Port references: EpicAncestor's cycle-guarded walk mirrors
// expandAncestorsOf (update.go) / CollectDescendants (data/hierarchy.go) --
// NOT devd (devd has no Epic-Ancestor search, Sprints are flat). The
// Master-Detail layout IDEA (Queue left with a Verdikt indicator, read-only
// Accordion preview right) references devd view_review_sprint.go's
// viewReviewSprint/reviewMasterPane/reviewDetailPane -- LAYOUT PATTERN ONLY,
// no Sprint-entity coupling (beans-tui has none, design-spec.md §4). The
// "(keine offenen Reviews)" empty-state placeholder mirrors devd
// view_navigate_reviews.go:20's own text pattern. keyReviewCockpit's a/x/o
// dispatch structurally mirrors devd keys_review.go's Verdikt-Dispatch
// pattern (a switch over the focused item's current tag state) -- devd
// itself has 3 verdict states (Grün/Rot/Peach) against a Sprint-scoped
// review_status enum; beans-tui has 2 (to-review/rework, design decision c/j
// -- no "passed" state to render, a passed bean simply leaves every visible
// section).

import (
	"fmt"
	"strings"

	"beans-tui/internal/clip"
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

// reviewStandMarkdown renders the WHOLE Review-Cockpit queue as Markdown for
// the clipboard (E5 Task 3, bean bt-e4a6, design decision b: the Cockpit's
// `y`-override, in place of the generic per-bean beanContext, context.go) --
// one "## <Epic ID> <Epic Title>" (or "## (kein Epic)") table per
// reviewQueue group, THEN a "## Rework" table when reviewRework is
// non-empty. Reuses reviewQueue/reviewRework as-is (Plan Step 4: "KEIN neuer
// Gruppierungscode") -- both already canonically sort their members/epics.
func reviewStandMarkdown(idx *data.Index) string {
	groups := reviewQueue(idx)
	rework := reviewRework(idx)

	total := 0
	for _, g := range groups {
		total += len(g.beans)
	}

	var out strings.Builder
	fmt.Fprintf(&out, "# Review-Stand (%d to-review)\n", total)

	for _, g := range groups {
		title := "(kein Epic)"
		if g.epic != nil {
			title = g.epic.ID + " " + g.epic.Title
		}
		fmt.Fprintf(&out, "\n## %s\n\n", title)
		out.WriteString(contextBeanTable(g.beans))
	}

	if len(rework) > 0 {
		fmt.Fprintf(&out, "\n## Rework (%d)\n\n", len(rework))
		out.WriteString(contextBeanTable(rework))
	}

	return out.String()
}

// reviewFlat is the Cockpit's cursor index space (types.go's reviewCursor
// doc-stamp): every to-review bean in reviewQueue's group order, THEN every
// reviewRework bean. Group headers are a render-time-only concern
// (reviewQueueRows below) -- never part of this index space, so up/down can
// never land on a non-actionable header row. Thin idx-taking wrapper over
// newReviewState/flat (I01, below) for callers (tests, keyReviewCockpit's
// own single per-keypress derivation) that only have an *data.Index at
// hand, not an already-built reviewState.
func reviewFlat(idx *data.Index) []*data.Bean {
	return newReviewState(idx).flat()
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
// count. Thin idx-taking wrapper over reviewState.summaryLine (I01, below)
// for standalone callers (tests) -- viewReviewCockpit itself calls
// rs.summaryLine directly on its own already-derived reviewState instead of
// going through this wrapper, so a single render only walks idx once.
func reviewSummaryLine(idx *data.Index, cursor int) string {
	return newReviewState(idx).summaryLine(cursor)
}

// reviewState (I01, E4-T3-Review PFLICHT carried into E4 Task 4, bean
// bt-yy6w): the Cockpit's queue shape, derived ONCE per render or per
// keypress and threaded through as a plain value instead of re-walked.
// Before this task, reviewQueue/reviewRework were independently re-invoked
// 3-4x by a SINGLE viewReviewCockpit render (reviewFlat's own reviewQueue+
// reviewRework calls, reviewSummaryLine's own reviewQueue call,
// reviewQueueRows' own reviewQueue+reviewRework calls) -- Task 4 adds a/x/o,
// three MORE call sites that all need the same to-review/rework boundary
// (reviewFocused/reviewIsRework below), so the redundant idx.WithTag/
// EpicAncestor re-walk is fixed here rather than grown further.
// reviewQueue/reviewRework themselves stay untouched (pure, independently
// unit-tested primitives, unchanged signatures) -- reviewState is a thin
// bundle over their output, not a replacement.
type reviewState struct {
	groups   []reviewGroup // reviewQueue's own grouped result
	toReview []*data.Bean  // groups flattened, in order -- cursor space [0, len(toReview))
	rework   []*data.Bean  // reviewRework's own result -- cursor space [len(toReview), len(flat()))
}

// newReviewState is the ONE idx.WithTag/EpicAncestor derivation callers
// should perform per render or per keypress (I01) -- viewReviewCockpit and
// keyReviewCockpit each call this exactly once at their own top and thread
// the result through every helper below that used to re-derive it.
func newReviewState(idx *data.Index) reviewState {
	groups := reviewQueue(idx)
	var toReview []*data.Bean
	for _, g := range groups {
		toReview = append(toReview, g.beans...)
	}
	return reviewState{groups: groups, toReview: toReview, rework: reviewRework(idx)}
}

// flat concatenates rs's to-review and rework halves -- the SAME
// to-review-then-rework cursor index space reviewFlat's own doc comment
// establishes, just read off an already-computed reviewState instead of
// re-walking idx.
func (rs reviewState) flat() []*data.Bean {
	flat := make([]*data.Bean, 0, len(rs.toReview)+len(rs.rework))
	flat = append(flat, rs.toReview...)
	flat = append(flat, rs.rework...)
	return flat
}

// summaryLine is reviewSummaryLine's logic (design decision j), scoped to an
// already-derived reviewState.
func (rs reviewState) summaryLine(cursor int) string {
	if cursor < len(rs.toReview) {
		return fmt.Sprintf("%d of %d", cursor+1, len(rs.toReview))
	}
	return fmt.Sprintf("Rework: %d offen", len(rs.rework))
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
// Takes an already-derived reviewState (I01) instead of an *data.Index --
// the caller (viewReviewCockpit) computes it ONCE per render and reuses it
// here and for the summary line, rather than this function re-walking
// idx.WithTag/EpicAncestor on its own.
func reviewQueueRows(rs reviewState, cursorFlat int, focused bool, bodyH int) []string {
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

	for _, g := range rs.groups {
		if g.epic != nil {
			rows = append(rows, theme.Muted.Render(relationRow(g.epic)))
		} else {
			rows = append(rows, theme.Dim.Render("(kein Epic)"))
		}
		for _, b := range g.beans {
			appendBeanRow(b, false)
		}
	}

	if len(rs.rework) > 0 {
		rows = append(rows, theme.Dim.Render("── Rework ──"))
		for _, b := range rs.rework {
			appendBeanRow(b, true)
		}
	}

	return windowAround(rows, bodyH, cursorRow)
}

// renderReviewDetailPane renders the read-only Detail-Accordion preview for
// the Cockpit's cursored bean, using the Cockpit's OWN reviewAccOpen digit-
// jump cursor (design decision i) -- deliberately NOT the Tree/Backlog's
// m.accOpen/m.secCursor/m.fieldCursor two-level detailFocus machine, which
// the always-read-only Cockpit preview does not have (no field-level
// relation-jump here). Delegates the actual bodyW/accW/beanSections/
// renderAccordion/renderPane body to the SHARED renderAccordionPane (I02,
// E4-T3-Review PFLICHT carried into this task, view_browse_repo.go) --
// before this task, this function and renderBeanAccordionPane
// hand-duplicated that ~15-line body; only the STATE feeding it (this
// function's own reviewAccOpen-derived activeIdx vs. the Tree/Backlog's
// shared fields) stays deliberately un-shared, per design decision i's own
// rationale (mirrors I01's copy-on-write doctrine, applied to shared render
// body vs. independent state instead of maps).
func (m model) renderReviewDetailPane(b *data.Bean, w, h int) string {
	activeIdx := m.reviewAccOpen - 1
	if activeIdx < 0 {
		activeIdx = 0
	}
	return renderAccordionPane(m.idx, b, w, h, m.reviewAccOpen, activeIdx, 0, true)
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

// clampReviewCursor (B01, E4-T4-Review PFLICHT, closed in T5, bean bt-v7ti)
// clamps m.reviewCursor into [0, len(reviewFlat(m.idx))) against the CURRENT
// m.idx -- called from applyLoaded (update.go) right after every reload, the
// one place the flat this cursor indexes into can shrink out from under it
// (a Pass/Reject verdict removes exactly the bean the cursor was parked on).
// Deliberately unconditional (safe to call even when m.view != viewReview
// Cockpit or reviewCursor is already valid -- a no-op then, cheaper than a
// view-gated branch) so it can sit in the shared reload path without a
// special case. Distinct from viewReviewCockpit's own render-local clamp
// (that one only ever fixed a LOCAL variable for a single render, never this
// field -- see that function's "Defensive clamp" comment): this is the fix
// for the model FIELD itself, so reviewFocused/focusedBean resolve correctly
// immediately after a reload, not just the next render's highlight.
func (m model) clampReviewCursor() model {
	switch flat := reviewFlat(m.idx); {
	case len(flat) == 0:
		m.reviewCursor = 0
	case m.reviewCursor >= len(flat):
		m.reviewCursor = len(flat) - 1
	case m.reviewCursor < 0:
		m.reviewCursor = 0
	}
	return m
}

// reviewFocused bounds-checks cursor against flat (the Cockpit's own
// to-review+rework cursor index space, reviewFlat's own doc comment) and
// returns the bean there, or nil when out of range -- the ONE guard
// keyReviewCockpit's a/x/o cases (below) and focusedBean's Cockpit case
// (update.go) share instead of independent index checks.
func reviewFocused(flat []*data.Bean, cursor int) *data.Bean {
	if cursor < 0 || cursor >= len(flat) {
		return nil
	}
	return flat[cursor]
}

// reviewIsRework reports whether b currently carries the "rework" tag --
// the single source of truth Pass/Reject (must NOT fire on an
// already-rejected bean) and Reopen (must ONLY fire on one) share below.
// ERRATUM vs. the plan's own pseudocode (epic-E4-plan.md »Task 4« Step 11,
// which sketches `reviewIsRework(idx, b)`): verified against the actual
// data shape instead of implemented as written -- ANY bean pulled out of
// reviewFlat already reflects the live idx it came from (reviewRework is a
// thin idx.WithTag("rework") pass-through), so a direct tag check on b is
// both correct and doesn't need an idx argument or a second WithTag walk
// (I01 precedent: avoid a redundant re-derivation, not just move it here).
func reviewIsRework(b *data.Bean) bool {
	for _, t := range b.Tags {
		if t == "rework" {
			return true
		}
	}
	return false
}

// keyReviewCockpit fully captures the Review-Cockpit (design decision h):
// up/down, n/p aliases, digit accordion-jump, esc/q back to Browse
// (Task 3), plus a/x/o Verdikt-actions (Task 4, bean bt-yy6w,
// design-spec.md §5/§7): `a` Pass (PassReview, design decision d) and `x`
// Reject (opens the Reject-Kommentar-Form, design decision e) both act on a
// to-review item and are a no-op on a Rework item (already verdicted once,
// re-verdicting would double-process it); `o` Reopen (SetTags,
// design decision f) is the mirror -- acts ONLY on a Rework item, a no-op on
// a to-review item (already in its target state). flat is derived ONCE at
// the top (I01) and reused by every case below via reviewFocused, instead
// of each case re-deriving reviewFlat/reviewIsRework's own idx.WithTag walk.
func (m model) keyReviewCockpit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	flat := reviewFlat(m.idx) // ONE derivation for this whole keypress (I01)

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
		// E5 Task 4 (bean bt-mne6): factored into reviewCursorMove (below)
		// so the keyboard AND the wheel dispatch (mouse.go handleMouse)
		// share the exact same clamp logic instead of two independent
		// copies.
		return m.reviewCursorMove(flat, -1), nil
	case "down":
		return m.reviewCursorMove(flat, 1), nil
	}

	// E5 Task 3 (bean bt-e4a6, design decision b): the Cockpit's own `y`
	// override -- reviewStandMarkdown (the WHOLE queue) instead of
	// keyNodeAction's generic per-bean beanContext. Reachable here at all
	// because keyReviewCockpit fully captures ahead of keyNodeAction
	// (handleKey, view==viewReviewCockpit case, checked BEFORE keyNodeAction
	// is ever dispatched) -- no separate guard against keyNodeAction needed.
	if keybind.Matches(msg, keys.Yank) {
		if err := clip.Copy(reviewStandMarkdown(m.idx)); err != nil {
			return m.showToast(toastWarn, "Yank fehlgeschlagen", "", nil, false)
		}
		return m.showToast(toastInfo, "Review-Stand kopiert", "", nil, false)
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
	case "a": // pass (design-spec §5's Pass row: status -> completed, tag removed)
		if b := reviewFocused(flat, m.reviewCursor); b != nil && !reviewIsRework(b) {
			etag, ok := m.beanETag(b.ID)
			if !ok {
				m.err = "Bean nicht mehr vorhanden — Verdikt verworfen"
				return m, nil
			}
			id, client := b.ID, m.client
			return m, mutateCmd(func() error { return client.PassReview(id, etag) })
		}
	case "x": // reject -> opens the Reject-Kommentar-Form (design decision e)
		if b := reviewFocused(flat, m.reviewCursor); b != nil && !reviewIsRework(b) {
			return m.openRejectForm(b)
		}
	case "o": // reopen -- design decision f: only meaningful on a Rework item
		if b := reviewFocused(flat, m.reviewCursor); b != nil && reviewIsRework(b) {
			etag, ok := m.beanETag(b.ID)
			if !ok {
				m.err = "Bean nicht mehr vorhanden"
				return m, nil
			}
			id, client := b.ID, m.client
			return m, mutateCmd(func() error {
				return client.SetTags(id, []string{"to-review"}, []string{"rework"}, etag)
			})
		}
	}
	return m, nil
}

// reviewCursorMove shifts m.reviewCursor by delta (±1), clamped to flat's
// bounds -- factored out of keyReviewCockpit's own up/down case (E5 Task 4,
// bean bt-mne6; I01: flat is derived ONCE per keypress there and threaded
// through, this takes it as a parameter instead of re-deriving
// reviewFlat(m.idx) itself) so the keyboard AND the wheel dispatch
// (mouse.go handleMouse) share one clamp.
func (m model) reviewCursorMove(flat []*data.Bean, delta int) model {
	switch {
	case delta < 0 && m.reviewCursor > 0:
		m.reviewCursor--
	case delta > 0 && m.reviewCursor < len(flat)-1:
		m.reviewCursor++
	}
	return m
}

// reviewCockpitChrome builds viewReviewCockpit's own head/localKeys strings
// -- factored out (E5 Task 4, bean bt-mne6) so reviewClickRow (mouse.go) can
// reconstruct this view's IDENTICAL geometry via clickPaneGeometry without a
// second, independently maintained copy of this breadcrumb/footer
// construction (Golden-Rule-Drift-Schutz, mirrors browseRepoChrome's own
// rationale, view_browse_repo.go).
func (m model) reviewCockpitChrome(innerW int) (head, localKeys string) {
	globalHint := renderBindings([]keybind.Binding{keys.Refresh, keys.Help, keys.Quit})
	head = breadcrumb(m.repoLabel(), "Review-Cockpit", globalHint, innerW)
	localHint := renderBindings([]keybind.Binding{keys.Up, keys.Down, keys.Enter, keys.Back, keys.Section}) + "  n/p:next/prev"
	localKeys = footer(localHint, innerW)
	return
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

	head, localKeys := m.reviewCockpitChrome(innerW)

	div := theme.Dim.Render(strings.Repeat("─", innerW))
	indicator := ""
	if m.watchUnavailable {
		indicator = "watch unavailable — ctrl+r für manuelles Reload"
	}
	status := statusBar(indicator, m.err, innerW)

	// E5 Task 4 (bean bt-mne6): bodyH/lw/rw now come from the SAME
	// clickPaneGeometry helper reviewClickRow (below) uses -- single source
	// for the numeric pane geometry (Golden-Rule-Drift-Schutz).
	bodyH, lw, rw, _, _ := clickPaneGeometry(w, h, head, localKeys)
	// ONE derivation for this whole render (I01) -- flat/summaryLine/
	// reviewQueueRows below all read off this SAME rs instead of each
	// independently re-walking idx.WithTag/EpicAncestor.
	rs := newReviewState(m.idx)
	flat := rs.flat()

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
		summaryLine := truncate(rs.summaryLine(cursor), lw-2)
		rows := append([]string{summaryLine}, reviewQueueRows(rs, cursor, true, bodyH-3)...)
		listBox = renderPane(pane{title: "Review-Queue", rows: rows}, lw, bodyH, true)
		detailBox = m.renderReviewDetailPane(flat[cursor], rw, bodyH)
	}
	body := lipgloss.JoinHorizontal(lipgloss.Top, listBox, detailBox)

	content := head + "\n" + div + "\n" + body + "\n" + div + "\n" + localKeys + "\n" + status
	out := outerBorder(content, innerW, true)

	return m.composeOverlays(out, w, h)
}

// reviewRowRef is one row of reviewQueueRows' own row space (above) --
// either a bean row (isBean=true, flatIdx indexes rs.flat()) or a header/
// separator row (epic title, "(kein Epic)", "── Rework ──") that is never a
// click target. reviewClickRow's own private counting type -- deliberately
// NOT threaded back into reviewQueueRows itself (that function stays
// untouched, zero golden-render risk; see reviewClickRow's own doc comment).
type reviewRowRef struct {
	flatIdx int
	isBean  bool
}

// reviewClickRow maps a mouse click to a flatIdx into rs.flat() (E5 Task 4,
// bean bt-mne6, design decision f; caller: mouse.go's mouseReviewClick) --
// mirrors treeClickRow's own reconstruction (view_browse_repo.go) against
// viewReviewCockpit's OWN render formula (above, via reviewCockpitChrome +
// clickPaneGeometry). The Cockpit's row space (reviewQueueRows, above)
// interleaves epic-header/"(kein Epic)"/"── Rework ──" rows (RENDER-ONLY,
// never a click target) among the actual bean rows -- this function
// re-walks rs.groups/rs.rework in EXACTLY the same order reviewQueueRows
// does (Golden-Rule-Drift-Schutz: if that walk order ever changes, this MUST
// change with it) to recover which windowed row a click landed on, and
// whether it is a bean row at all. Row 0 is the summary line
// (rs.summaryLine) -- never a row target, same "search/summary head line"
// budget trade as Tree/Backlog. ok=false for a click outside the Queue
// pane's column span, on/above the summary line, past the last rendered
// row, on a header/separator row, or when the queue is empty (the
// "(keine offenen Reviews)" placeholder has no row target).
func reviewClickRow(m model, rs reviewState, msg tea.MouseMsg) (flatIdx int, ok bool) {
	flat := rs.flat()
	if len(flat) == 0 {
		return 0, false
	}
	cursor := m.reviewCursor
	if cursor >= len(flat) {
		cursor = len(flat) - 1
	}
	if cursor < 0 {
		cursor = 0
	}

	// Re-walk rs's row structure WITHOUT rendering text -- just enough to
	// know, per row, whether it's a bean row (and which flatIdx) or a
	// header/separator, plus the SAME cursorRow reviewQueueRows computes for
	// its own windowAround call (view_review_cockpit.go's own doc comment
	// above: "cursorRow = len(rows)" right before appending the cursored
	// row).
	var rows []reviewRowRef
	cursorRow := 0
	fi := 0
	for _, g := range rs.groups {
		rows = append(rows, reviewRowRef{isBean: false}) // epic / "(kein Epic)" header
		for range g.beans {
			if fi == cursor {
				cursorRow = len(rows)
			}
			rows = append(rows, reviewRowRef{flatIdx: fi, isBean: true})
			fi++
		}
	}
	if len(rs.rework) > 0 {
		rows = append(rows, reviewRowRef{isBean: false}) // "── Rework ──" separator
		for range rs.rework {
			if fi == cursor {
				cursorRow = len(rows)
			}
			rows = append(rows, reviewRowRef{flatIdx: fi, isBean: true})
			fi++
		}
	}

	w, h := m.width, m.height
	if w <= 0 {
		w = 80
	}
	if h <= 0 {
		h = 24
	}
	innerW := w - 2
	head, localKeys := m.reviewCockpitChrome(innerW)

	bodyH, lw, _, originX, originY := clickPaneGeometry(w, h, head, localKeys)

	if msg.X < originX || msg.X >= originX+lw {
		return 0, false
	}
	clickRow := msg.Y - originY
	if clickRow <= 0 {
		return 0, false
	}

	windowRows := bodyH - 3
	if windowRows < 0 {
		windowRows = 0
	}
	start := windowStart(len(rows), windowRows, cursorRow)
	visible := windowRows
	if len(rows)-start < visible {
		visible = len(rows) - start
	}
	rowIdx := clickRow - 1
	if rowIdx < 0 || rowIdx >= visible {
		return 0, false
	}
	r := rows[start+rowIdx]
	if !r.isBean {
		return 0, false
	}
	return r.flatIdx, true
}
