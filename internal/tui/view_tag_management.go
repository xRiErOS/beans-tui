package tui

// view_tag_management.go — the Tag-Management page (E10 Task 2, bean
// bt-r92i, epic bt-362n D05-D09): a read-only, single-pane, FULL-CAPTURE top
// level view listing the Union of registry-defined tags (T1's
// internal/data/tagdefs.go, always shown even at count 0) and free/
// undefined tags currently in use on at least one bean. Create/Delete/
// Rename land in T3/T4/T5 -- this task ships ONLY the Grundgerüst: routing
// (D05, via the Command-Center "go to tags", mirrors the "go to settings"
// precedent -- no dedicated keybinding), the D06 full-capture dispatch, D07
// Chrome (empty GlobalHint), D08 single-pane layout (enter is a documented
// handled no-op, reserved for a future Master-Detail drilldown --
// idx.WithTag(tag) already exists), and D09's Union/sort algebra.
//
// Render structurally mirrors viewBacklog() (view_browse_backlog.go): same
// innerW/innerH algebra, same clickPaneGeometry-sourced bodyH (Golden-Rule-
// Drift-Schutz, that helper's own doc-stamp), same renderPane wrapping --
// just ONE pane instead of a JoinHorizontal master-detail pair (D08). Each
// row carries a PF-12-style ALWAYS-reserved marker column (design-spec.md
// §15: "kein bedingtes Präfix, das das Layout verschiebt") distinguishing a
// defined row from a free row -- pre-borrowing the SAME reserved-gutter
// convention D10/T6 (bean bt-pqq3) will apply to the Tag-Picker's own
// suggest-mode rows, though this page has no Picker relationship of its own
// (bean bt-r92i's own "D10-Konvention vorgezogen ... hier noch ohne
// Picker-Bezug" wording).

import (
	"sort"
	"strconv"
	"strings"

	"beans-tui/internal/data"
	"beans-tui/internal/theme"
	keybind "github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// tagRegistryRow is one row of the Tag-Management page's D09 Union list.
type tagRegistryRow struct {
	name    string
	count   int
	defined bool
}

// tagRegistryRows builds the D09 Union: every registry-defined tag in defs
// FIRST (alpha-sorted, deduped defensively -- LoadTagDefs already
// sorts/dedupes, T1, but this function must not crash or double-list if it
// is ever handed a raw, un-sanitized slice), carrying its REAL usage count
// (0 for an unused defined tag, T1's own "auch mit Count 0" wording) -- THEN
// every tag currently in use on at least one bean that is NOT in defs
// ("frei"/undefined), count-descending then alpha tie-break (mirrors
// collectTagCounts' own tie-break, box_picker_tag.go). A tag that is BOTH
// defined AND in use appears exactly ONCE, in the Defined group, with its
// real count -- never a duplicate row in the Free group.
//
// PRELUDE (T3, bean bt-604w, T6-Review F01): the COUNTING/UNION pass is now
// SHARED with collectTagCounts (box_picker_tag.go, T6, D10) instead of a
// second, separately-maintained loop over idx.ByID -- resolves the
// D10-ERRATUM-Notiz T2 (bean bt-r92i) itself forward-flagged ("once T6
// extends collectTagCounts with its own defined map[string]bool parameter,
// THIS function could call that extended version directly"). collectTagCounts
// already returns the identical defined+free Union, globally sorted (defined
// primary, count-desc/alpha secondary/tertiary, sortTagCountsDefinedFirst) --
// but D09's own grouping contract diverges from that global sort in exactly
// ONE place: the Defined group here sorts ALPHA ONLY (a predictable,
// index-like read), not count-desc like collectTagCounts' Picker-facing
// sort. So this function still owns its OWN re-split + re-sort of the
// Defined half (a stable partition of collectTagCounts' already-sorted
// output, re-sorted by name) -- the Free half's relative order coming out of
// that same partition is ALREADY exactly D09's own count-desc/alpha
// contract (collectTagCounts' secondary/tertiary keys), so it is reused
// AS-IS, no second sort. Page GROUPING (D09) stays Page logic; only the
// underlying COUNT/UNION pass is now shared (this task's own PRELUDE
// mandate: "Zählung geteilt, Page-Gruppierung bleibt Page-Logik").
func tagRegistryRows(idx *data.Index, defs []string) []tagRegistryRow {
	defSet := make(map[string]bool, len(defs))
	for _, name := range defs {
		defSet[name] = true
	}
	counts := collectTagCounts(idx, defSet)

	definedRows := make([]tagRegistryRow, 0, len(defSet))
	freeRows := make([]tagRegistryRow, 0, len(counts))
	for _, c := range counts {
		row := tagRegistryRow{name: c.tag, count: c.count, defined: c.defined}
		if c.defined {
			definedRows = append(definedRows, row)
		} else {
			freeRows = append(freeRows, row)
		}
	}
	sort.Slice(definedRows, func(i, j int) bool { return definedRows[i].name < definedRows[j].name })

	out := make([]tagRegistryRow, 0, len(definedRows)+len(freeRows))
	out = append(out, definedRows...)
	out = append(out, freeRows...)
	return out
}

// openTagManagementPage transitions into the Tag-Management page (D05
// entry): loads the registry FRESH from disk (D03, mirrors openLobby's own
// "reload from disk on every open" convention) -- synchronous, no tea.Cmd
// (D02: a local file read is fast enough for a direct call, same contract
// LoadTagDefs itself documents). A nil m.client (pre-load / test fixture)
// degrades to an empty registry rather than panicking -- mirrors D02's own
// tolerant-missing philosophy one layer up. Resets tagMgmtCursor (mirrors
// openLobby's repoList reset) so a stale cursor from a PREVIOUS, longer page
// visit never survives into a shorter re-open.
func (m model) openTagManagementPage() (tea.Model, tea.Cmd) {
	m.view = viewTagManagement
	var defs []string
	if m.client != nil {
		defs, _ = m.client.LoadTagDefs() // D02: LoadTagDefs never returns a non-nil error
	}
	m.tagMgmtRows = tagRegistryRows(m.idx, defs)
	m.tagMgmtCursor = listState{}
	m.tagMgmtCursor.setLen(len(m.tagMgmtRows))
	return m, nil
}

// tagManagementChrome builds viewTagManagement's own head/localKeys strings
// -- mirrors backlogChrome's factoring-out rationale (view_browse_backlog.go
// doc-stamp), even though this task wires no mouse-click geometry consumer
// for it yet (T3-T5 may add row-level mouse targets later, out of scope
// here). D07: GlobalHint is LEFT EMPTY (not renderBindings(globalBindings()))
// -- none of the 4 global keys (Palette/Picker/Help/Quit) function while
// this full-capture page holds input (D06), so showing them would be a
// misleading, non-functional promise (the exact bug class D06/PF-11 already
// fixed once for stale footer hints, epic bt-362n body's own D07 citation).
func (m model) tagManagementChrome(innerW int) (head, localKeys string) {
	head = breadcrumb(m.repoLabel(), "Tags", "", innerW)
	localKeys = footer(renderBindings(tagManagementLocalBindings()), innerW)
	return
}

// tagManagementLocalBindings is the page's own Footer Zone 3 local set.
// T2 vorerst: Up/Down/Back only (read-only Grundgerüst) -- T3 (Create)/T4
// (Delete)/T5 (Rename) each append their OWN binding to this SAME shared
// function body as their overlays land (bean bt-r92i's own wording: "EIN
// gemeinsamer Funktionsrumpf, kein Duplikat je Task"), never a
// second, parallel bindings list.
func tagManagementLocalBindings() []keybind.Binding {
	return []keybind.Binding{keys.Up, keys.Down, keys.Back}
}

// tagManagementMarkerGlyph is the PF-12 reserved-gutter glyph (design-
// spec.md §15: "kein bedingtes Präfix, das das Layout verschiebt") marking a
// REGISTRY-DEFINED row -- a free/undefined row renders the SAME-WIDTH blank
// instead of omitting the column outright, so neither group's name column
// ever shifts relative to the other. U+2713 CHECK MARK is EAW=Neutral
// (Narrow, same class as theme.priorityGlyph's "‼"/"!" -- already
// column-safe without needing a BT_ASCII_ICONS fallback variant, theme.go's
// own doc-stamp on that pair).
const tagManagementMarkerGlyph = "✓"

// tagManagementMarkerStyle colors the defined-marker glyph Green (a distinct
// hue from theme.Accent/theme.Dim, which the cursor-bar treatment below
// already claims for "this row has the cursor" -- reusing Accent here would
// conflate "defined" with "selected").
var tagManagementMarkerStyle = lipgloss.NewStyle().Foreground(theme.Green)

// tagManagementRowText renders one row's plain content to exactly w cells:
// the PF-12 marker column + a space + the (truncated if needed) name on the
// left, the usage count right-aligned via pickerRowFill (Port view_lobby.go
// repoPickerBody's own row-fill idiom for a right-aligned metric column) --
// same nameW-budgeting shape as repoPickerBody's own cursor+metric+nameW
// calc, so a narrow (80-column) terminal truncates the name with "…" instead
// of wrapping or overflowing (tmux-smoke acceptance wording).
func tagManagementRowText(r tagRegistryRow, w int) string {
	marker := " "
	if r.defined {
		marker = tagManagementMarkerStyle.Render(tagManagementMarkerGlyph)
	}
	count := theme.Muted.Render(strconv.Itoa(r.count))
	nameW := w - lipgloss.Width(marker) - 1 - lipgloss.Width(count) - 1
	if nameW < 1 {
		nameW = 1
	}
	left := marker + " " + truncate(r.name, nameW)
	return pickerRowFill(left, count, w)
}

// tagManagementRows renders every row using the SAME inline `▌`-cursor-bar
// convention as backlogRows/treeRows (view_browse_backlog.go/
// view_browse_repo.go) -- D08's own "structurally closer to Backlog than to
// a floating overlay menu" call, NOT the Tag-Picker's ▸+Header style
// (box_picker_tag.go's Picker-Stil-Divergenz doc-stamp, types.go). paneW is
// renderPane's OWN outer width param (the caller passes the same value it
// hands to renderPane) -- content is budgeted to paneW-3 (paneW-2 is
// renderPane's own truncate budget, doc-stamp there; -1 more reserves this
// function's own 1-column cursor-bar prefix), so the fully-assembled row
// ("▌"+content or " "+content) fits renderPane's truncate call as a
// practical no-op instead of relying on it to silently cut a too-long line.
func (m model) tagManagementRows(rows []tagRegistryRow, focused bool, paneW, bodyH int) []string {
	contentW := paneW - 3
	if contentW < 1 {
		contentW = 1
	}
	pos := m.tagMgmtCursor.cursor
	out := make([]string, len(rows))
	for i, r := range rows {
		text := tagManagementRowText(r, contentW)
		if i == pos {
			plain := ansi.Strip(text)
			if focused {
				out[i] = theme.Accent.Render("▌" + plain)
			} else {
				out[i] = theme.Dim.Render("▌" + plain)
			}
		} else {
			out[i] = " " + text
		}
	}
	return windowAround(out, bodyH, pos)
}

// viewTagManagement renders the Tag-Management page -- mirrors viewBacklog's
// own algebra (view_browse_backlog.go) exactly (innerW/innerH via the
// established w-2/h-2 convention, clickPaneGeometry-sourced bodyH, Golden
// Rule #1 via renderPane), just ONE renderPane call instead of a
// JoinHorizontal master-detail pair (D08 -- no lw/rw split).
func (m model) viewTagManagement() string {
	w, h := m.width, m.height
	if w <= 0 {
		w = 80
	}
	if h <= 0 {
		h = 24
	}
	innerW := w - 2

	head, localKeys := m.tagManagementChrome(innerW)

	div := theme.Dim.Render(strings.Repeat("─", innerW))
	indicator := ""
	if m.watchUnavailable {
		indicator = "watch unavailable — ctrl+r for manual reload"
	}
	status := statusBar(indicator, m.err, innerW)

	// Same shared geometry source viewBrowseRepo/viewBacklog use (mouse.go
	// clickPaneGeometry) -- lw/rw are the SPLIT master-detail widths neither
	// of THIS page's single pane needs (D08), only bodyH is consumed.
	bodyH, _, _, _, _ := clickPaneGeometry(w, h, head, localKeys, m.settings.Layout.TreeWidth)

	// paneW is the SINGLE pane's own CONTENT-width param passed to
	// renderPane/tagManagementRows -- mirrors viewBrowseRepo's own Vollbild
	// single-pane precedent (view_browse_repo.go, F01, bean bt-13l7's own
	// doc-stamp there): renderPane's `w` argument is a CONTENT width, its OWN
	// RoundedBorder adds +2 more on top (masterDetailWidths' lw+rw+4==innerW
	// is the identical accounting for the TWO Split panes; here there is
	// only ONE, so its content width is innerW-2, not innerW itself --
	// passing innerW unadjusted double-counts the border and overflows the
	// outer frame by 2 columns, exactly the class of bug view_browse_repo.go
	// line ~798 already documents "caught live via tmux smoke" -- caught the
	// SAME way here, this task's own tmux-smoke run).
	paneW := innerW - 2

	listRows := m.tagManagementRows(m.tagMgmtRows, true, paneW, bodyH)
	if len(m.tagMgmtRows) == 0 {
		listRows = []string{theme.Muted.Render("(no tags)")}
	}
	listBox := renderPane(pane{rows: listRows}, paneW, bodyH, true)

	content := head + "\n" + div + "\n" + listBox + "\n" + div + "\n" + localKeys + "\n" + status
	out := outerBorder(content, innerW, true)

	return m.composeOverlays(out, w, h)
}

// keyTagManagement drives the open Tag-Management page (D06 full-capture,
// dispatched from handleKey at the SAME checkpoint as keyLobby -- see that
// check's own doc-stamp in update.go). up/down move tagMgmtCursor (reuse
// navKey, same convention as every other list cursor in this codebase);
// enter is a documented HANDLED no-op (D08 -- reserved for a future
// Master-Detail drilldown fast-follow, idx.WithTag(tag) already exists);
// esc (keys.Back) returns to Browse (D03/D06-Pendant: exactly ONE level
// back, mirrors keyLobby's own esc-with-a-live-client case). Every OTHER key
// (including the global s/t/a/r/c/d/e node-action set AND `?`/ctrl+k/`p`) is
// silently swallowed -- the D06 regression this page exists to prevent.
func (m model) keyTagManagement(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch navKey(msg.String()) {
	case "up":
		m.tagMgmtCursor.move(-1)
		return m, nil
	case "down":
		m.tagMgmtCursor.move(1)
		return m, nil
	}

	switch {
	case keybind.Matches(msg, keys.Back):
		m.view = viewBrowseRepo
		return m, nil
	case keybind.Matches(msg, keys.Enter):
		// D08: handled no-op -- a future Master-Detail drilldown fast-follow
		// ("Beans zu diesem Tag") would land here, idx.WithTag(tag) already
		// exists (epic bt-362n body's own empirical-verification note).
		return m, nil
	}
	return m, nil
}
