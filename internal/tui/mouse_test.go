package tui

// mouse_test.go — TDD coverage for E5 Task 4 (bean bt-mne6, epic bt-5h4d,
// design decision f): Maus (Wheel/Klick/Doppelklick). Mirrors devd
// mouse_test.go's render-grounded click pattern (~/Obsidian/tools/Developer
// Dashboard/apps/cli-go/internal/tui/mouse_test.go): a click coordinate is
// found by rendering the REAL View() and locating a known substring on
// screen, never hand-computed against the click formula itself (that would
// be circular -- a formula bug would pass its own test). Doppelklick tests
// inject a fixed m.clock (mirrors devd's own TestMouseDoubleClickCollapses
// OpenNode/TestMouseSingleClickOnOpenNodeStaysOpen pattern) for a
// deterministic <500ms window instead of a real sleep.

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// wheelMsg builds a wheel tea.MouseMsg (X/Y are irrelevant to wheel dispatch,
// handleMouse's wheel branch only reads Button).
func wheelMsg(b tea.MouseButton) tea.MouseMsg {
	return tea.MouseMsg{Button: b, Action: tea.MouseActionPress}
}

// screenLines renders m's FULL View() and splits it into ANSI-stripped
// lines -- real screen geometry (outer border, panes, everything), the same
// ground truth clickAt/leftPaneClickAt below search over.
func screenLines(m model) []string {
	return strings.Split(ansi.Strip(m.View()), "\n")
}

// leftPaneClickAt finds substr's first occurrence WITHIN the left (list)
// pane's column span (boundary = originX+lw, from clickPaneGeometry,
// mouse.go) of m's rendered View() and returns a real left-click
// tea.MouseMsg at that screen coordinate. Restricting to the left pane
// matters: the right Detail pane can render the SAME text (e.g. a bean
// title also shown in its own Meta section, or listed under a parent's
// Beziehungen section) -- a plain "first match anywhere" search could
// silently click the wrong pane. head/localKeys must be the SAME strings
// the caller's own view renders with (browseRepoChrome/backlogChrome), so
// the boundary matches the real render exactly.
func leftPaneClickAt(t *testing.T, m model, head, localKeys, substr string) tea.MouseMsg {
	t.Helper()
	_, lw, _, originX, _ := clickPaneGeometry(m.width, m.height, head, localKeys, m.settings.Layout.TreeWidth)
	boundary := originX + lw
	for y, l := range screenLines(m) {
		i := strings.Index(l, substr)
		if i < 0 {
			continue
		}
		col := ansi.StringWidth(l[:i])
		if col >= boundary {
			continue // leftmost match is already in the right Detail pane
		}
		return tea.MouseMsg{Button: tea.MouseButtonLeft, Action: tea.MouseActionPress, X: col, Y: y}
	}
	t.Fatalf("substr %q not found in the left pane of the rendered View()", substr)
	return tea.MouseMsg{}
}

// treeClickAt is leftPaneClickAt scoped to the Browse/Tree view's own
// chrome (browseRepoChrome, view_browse_repo.go).
func treeClickAt(t *testing.T, m model, substr string) tea.MouseMsg {
	t.Helper()
	head, localKeys := m.browseRepoChrome(m.width - 2)
	return leftPaneClickAt(t, m, head, localKeys, substr)
}

// rightPaneClickAt mirrors leftPaneClickAt but scoped to the RIGHT Detail
// pane's own column span (X >= originX+lw) -- B07 (design-spec.md §15
// PF-16, bean bt-duz7)'s render-grounded click helper for the
// mouseDetailClick dispatch path.
func rightPaneClickAt(t *testing.T, m model, head, localKeys, substr string) tea.MouseMsg {
	t.Helper()
	_, lw, _, originX, _ := clickPaneGeometry(m.width, m.height, head, localKeys, m.settings.Layout.TreeWidth)
	boundary := originX + lw
	for y, l := range screenLines(m) {
		i := strings.Index(l, substr)
		if i < 0 {
			continue
		}
		col := ansi.StringWidth(l[:i])
		if col < boundary {
			continue // leftmost match is in the left pane
		}
		return tea.MouseMsg{Button: tea.MouseButtonLeft, Action: tea.MouseActionPress, X: col, Y: y}
	}
	t.Fatalf("substr %q not found in the right Detail pane of the rendered View()", substr)
	return tea.MouseMsg{}
}

// detailClickAt is rightPaneClickAt scoped to the Browse/Tree view's own
// chrome (browseRepoChrome).
func detailClickAt(t *testing.T, m model, substr string) tea.MouseMsg {
	t.Helper()
	head, localKeys := m.browseRepoChrome(m.width - 2)
	return rightPaneClickAt(t, m, head, localKeys, substr)
}

// backlogDetailClickAt is rightPaneClickAt scoped to the Backlog view's own
// chrome (backlogChrome) -- guards detailClickRow's m.view branch.
func backlogDetailClickAt(t *testing.T, m model, substr string) tea.MouseMsg {
	t.Helper()
	head, localKeys := m.backlogChrome(m.width - 2)
	return rightPaneClickAt(t, m, head, localKeys, substr)
}

// detailFocusModel builds a fixtureModel with tk-2 focused, ancestors
// expanded (focusBean, box_menu_value_test.go), and a real terminal size --
// shared setup for the detailClickRow/mouseDetailClick tests below (B07).
func detailFocusModel(t *testing.T) model {
	t.Helper()
	m := fixtureModel(t, fixtureBeans())
	m = step(t, m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m = focusBean(m, "tk-2")
	return m
}

// --- Wheel: moves the active view's cursor (design decision f) ---

func TestWheelUpDownMovesTreeCursor(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m = step(t, m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m.expanded["ms-1"] = true
	m.expanded["ep-1"] = true // nodes: ms-1(0) ep-1(1) tk-1(2) tk-2(3)
	m.cursorID = "ms-1"

	tm, _ := m.handleMouse(wheelMsg(tea.MouseButtonWheelDown))
	m2, ok := tm.(model)
	if !ok {
		t.Fatalf("handleMouse(wheel down) did not return a model, got %T", tm)
	}
	if m2.cursorID != "ep-1" {
		t.Fatalf("wheel down: cursorID = %q, want ep-1", m2.cursorID)
	}

	tm2, _ := m2.handleMouse(wheelMsg(tea.MouseButtonWheelUp))
	m3 := tm2.(model)
	if m3.cursorID != "ms-1" {
		t.Fatalf("wheel up: cursorID = %q, want ms-1", m3.cursorID)
	}
}

func TestWheelMovesBacklogCursor(t *testing.T) {
	m := fixtureModel(t, backlogBeans())
	m = step(t, m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m.view = viewBacklog
	vis := m.backlogVisible()
	if len(vis) < 2 {
		t.Fatalf("setup: need >=2 backlog-visible beans, got %d", len(vis))
	}
	m.backlogList.setLen(len(vis))
	m.backlogList.cursor = 0

	tm, _ := m.handleMouse(wheelMsg(tea.MouseButtonWheelDown))
	m2, ok := tm.(model)
	if !ok {
		t.Fatalf("handleMouse(wheel down) did not return a model, got %T", tm)
	}
	if m2.backlogList.cursor != 1 {
		t.Fatalf("wheel down: backlogList.cursor = %d, want 1", m2.backlogList.cursor)
	}

	tm2, _ := m2.handleMouse(wheelMsg(tea.MouseButtonWheelUp))
	m3 := tm2.(model)
	if m3.backlogList.cursor != 0 {
		t.Fatalf("wheel up: backlogList.cursor = %d, want 0", m3.backlogList.cursor)
	}
}

// --- Click: sets the Tree cursor, render-grounded ---

func TestClickSetsTreeCursor(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m = step(t, m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m.expanded["ms-1"] = true
	m.expanded["ep-1"] = true
	m.cursorID = "ms-1"

	msg := treeClickAt(t, m, "Task Two")
	tm, _ := m.handleMouse(msg)
	m2, ok := tm.(model)
	if !ok {
		t.Fatalf("handleMouse(click) did not return a model, got %T", tm)
	}
	if m2.cursorID != "tk-2" {
		t.Fatalf("click on \"Task Two\" row: cursorID = %q, want tk-2", m2.cursorID)
	}
}

// --- Doppelklick: devd D03 semantics (design decision f) ---

// TestDoubleClickTogglesExpand: a SECOND click on the SAME node within
// doubleClickInterval collapses an already-open, expandable node.
func TestDoubleClickTogglesExpand(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m = step(t, m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m.expanded["ms-1"] = true
	m.expanded["ep-1"] = true // Epic One open, expandable (2 children)
	m.cursorID = "ms-1"
	fixed := time.Unix(1000, 0)
	m.clock = func() time.Time { return fixed }

	msg := treeClickAt(t, m, "Epic One")
	tm, _ := m.handleMouse(msg) // 1st click: registers, does not collapse
	m2, ok := tm.(model)
	if !ok {
		t.Fatalf("handleMouse(click) did not return a model, got %T", tm)
	}
	if !m2.expanded["ep-1"] {
		t.Fatalf("setup: first click on an open node must not collapse it")
	}

	tm2, _ := m2.handleMouse(msg) // 2nd click, same fixed time/node -> double
	m3 := tm2.(model)
	if m3.expanded["ep-1"] {
		t.Fatalf("double click on an open, expandable node must collapse it (devd D03)")
	}
}

// TestSingleClickOnOpenNodeDoesNotCollapse: an ISOLATED single click on an
// already-open node only moves the cursor -- it must NOT toggle (devd D03,
// distinct from the double-click case above: no second click ever arrives).
func TestSingleClickOnOpenNodeDoesNotCollapse(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m = step(t, m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m.expanded["ms-1"] = true
	m.expanded["ep-1"] = true
	m.cursorID = "ms-1"
	fixed := time.Unix(2000, 0)
	m.clock = func() time.Time { return fixed }

	msg := treeClickAt(t, m, "Epic One")
	tm, _ := m.handleMouse(msg)
	m2, ok := tm.(model)
	if !ok {
		t.Fatalf("handleMouse(click) did not return a model, got %T", tm)
	}
	if !m2.expanded["ep-1"] {
		t.Fatal("a single click on an already-open node must NOT collapse it (devd D03)")
	}
	if m2.cursorID != "ep-1" {
		t.Fatalf("cursorID = %q, want ep-1 (click still moves the cursor)", m2.cursorID)
	}
}

// --- Overlay guard: mouse ignored while a form/overlay/palette/search/
// filter/help/quit-confirm fully captures input (devd precedent) ---

func TestMouseIgnoredWhileFormOpen(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m = step(t, m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m.expanded["ms-1"] = true
	m.cursorID = "ms-1"

	nm, _ := m.openCreateForm()
	m, ok := nm.(model)
	if !ok || m.form == nil {
		t.Fatal("setup: openCreateForm did not set m.form")
	}

	tm, _ := m.handleMouse(wheelMsg(tea.MouseButtonWheelDown))
	m2, ok := tm.(model)
	if !ok {
		t.Fatalf("handleMouse(wheel) did not return a model, got %T", tm)
	}
	if m2.cursorID != "ms-1" {
		t.Fatalf("wheel while a form is open moved the cursor: cursorID = %q, want unchanged (ms-1)", m2.cursorID)
	}
}

func TestMouseIgnoredWhileOverlayOpen(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m = step(t, m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m.expanded["ms-1"] = true
	m.cursorID = "ms-1"
	m.overlay = overlayValueMenu

	tm, _ := m.handleMouse(wheelMsg(tea.MouseButtonWheelDown))
	m2, ok := tm.(model)
	if !ok {
		t.Fatalf("handleMouse(wheel) did not return a model, got %T", tm)
	}
	if m2.cursorID != "ms-1" {
		t.Fatalf("wheel while an overlay is open moved the cursor: cursorID = %q, want unchanged (ms-1)", m2.cursorID)
	}
}

// TestToastClickDismissesEvenWithFormOpen is the Cross-Feature-Fix
// regression guard (design decision a, Port devd DD2-272/273): a Toast
// click-dismiss must reach the PO even while a form is open. Goes through
// the FULL m.Update() dispatcher (not handleMouse directly) -- this is the
// test update.go's own placement comment calls out as the actual
// verification that the tea.MouseMsg case runs ahead of the `if m.form !=
// nil` fallback, not just documentation of the assumption.
func TestToastClickDismissesEvenWithFormOpen(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m = step(t, m, tea.WindowSizeMsg{Width: 100, Height: 30})

	nm, _ := m.openCreateForm()
	m, ok := nm.(model)
	if !ok || m.form == nil {
		t.Fatal("setup: openCreateForm did not set m.form")
	}

	m, _ = m.showToast(toastInfo, "x", "", nil, false)
	x, y, _, _ := m.toastGeometry()

	tm, _ := m.Update(tea.MouseMsg{Button: tea.MouseButtonLeft, Action: tea.MouseActionPress, X: x, Y: y})
	m2, ok := tm.(model)
	if !ok {
		t.Fatalf("Update(mouse) did not return a model, got %T", tm)
	}
	if m2.toast != nil {
		t.Fatal("a click on the toast must dismiss it even while a form is open (Cross-Feature-Fix)")
	}
	if m2.form == nil {
		t.Error("the form must stay open -- only the toast is dismissed by this click")
	}
}

// --- TreeWidth wiring (T6b, bean bt-pd22, T5-Review I01): clickPaneGeometry
// resolves treeWidth (the caller's m.settings.Layout.TreeWidth) into
// masterDetailWidths' floor via treeWidthFloor (mouse.go) instead of the
// hardcoded "24" every caller passed before this task. w=72/h=30 is chosen
// so innerW=70's own w/3=23 floor-1fr sits BELOW both 24 (the fallback) and
// 28 (the configured value below) while w*2/5=28's cap does NOT clip 28 --
// i.e. a width where the floor's own value is what determines lw, not the
// unrelated 1fr/cap arithmetic masterDetailWidths (view.go) also applies.

// TestTreeWidthZeroFallsBackToDefault guards the 0/unset fallback: every
// existing golden fixture/test model's m.settings is the config.Settings{}
// zero value (TreeWidth 0) until app.go's Run() or a Settings-Form submit
// populates it -- clickPaneGeometry(treeWidth=0) must still resolve to 24,
// the SAME value this file hardcoded before T6b, so the seven Golden
// snapshot tests stay byte-identical.
func TestTreeWidthZeroFallsBackToDefault(t *testing.T) {
	_, lw, _, _, _ := clickPaneGeometry(72, 30, "head", "footer", 0)
	if lw != 24 {
		t.Fatalf("clickPaneGeometry(treeWidth=0): lw = %d, want 24 (fallback, byte-identical to the pre-T6b hardcoded floor)", lw)
	}
}

// TestTreeWidthFromSettingsAffectsGeometry guards the actual wiring: a
// configured m.settings.Layout.TreeWidth must widen the left pane beyond
// the 0/unset fallback -- the whole point of T6b (T5-Review I01: before
// this task, the Settings-Form's tree_width field was persisted/validated
// but had NO effect on the render, a silent no-op for the PO).
func TestTreeWidthFromSettingsAffectsGeometry(t *testing.T) {
	_, lwDefault, _, _, _ := clickPaneGeometry(72, 30, "head", "footer", 0)
	_, lwConfigured, _, _, _ := clickPaneGeometry(72, 30, "head", "footer", 28)
	if lwConfigured != 28 {
		t.Fatalf("clickPaneGeometry(treeWidth=28): lw = %d, want 28", lwConfigured)
	}
	if lwConfigured <= lwDefault {
		t.Fatalf("configured treeWidth (28) did not widen the left pane vs. the default fallback (lw=%d): got lw=%d", lwDefault, lwConfigured)
	}
}

// TestClickPaneGeometryOriginYExcludesTitleAndSeparator guards PF-10
// (design-spec.md §15, epic-E7-plan.md »Task 5«, bean bt-uyzf): renderPane no
// longer draws a title line + underline-separator ahead of a pane's rows, so
// clickPaneGeometry's originY loses the two `+1` terms that used to account
// for them -- originY is now outer-top-border(1) + head height + divider(1)
// + the pane's OWN top border(1) (a DIFFERENT line, still there), with row 0
// of the caller's own windowed-content index space starting immediately
// after.
func TestClickPaneGeometryOriginYExcludesTitleAndSeparator(t *testing.T) {
	head := "head"
	_, _, _, _, originY := clickPaneGeometry(80, 24, head, "footer", 0)
	want := 1 + lipgloss.Height(head) + 1 + 1
	if originY != want {
		t.Fatalf("originY = %d, want %d (outer border(1) + head(%d) + divider(1) + pane's own top border(1) -- PF-10 drops only the pane's title+separator lines)", originY, want, lipgloss.Height(head))
	}
}

// --- B07 (design-spec.md §15 PF-16, bean bt-duz7, E8 Task 4): Maus im
// Detail-Pane -- Sektionen + Meta-Felder klickbar ---

// TestDetailClickRowAccountsForFiveLineHeaderOffset guards the Kopfblock
// offset (bean bt-duz7 PO-Wortlaut, "Kopfblock-Offset 5 Zeilen beachten!"):
// a click on the Kopfblock's own ID line (row 0 of the pane's content,
// ABOVE the 5-line offset) must resolve to NO Section-/Feld-Treffer.
func TestDetailClickRowAccountsForFiveLineHeaderOffset(t *testing.T) {
	m := detailFocusModel(t)
	b := m.focusedBean()

	msg := detailClickAt(t, m, "tk-2") // Kopfblock row 0 (ID line)
	_, _, ok := detailClickRow(m, b, msg)
	if ok {
		t.Fatal("a click on the Kopfblock's ID line must not resolve to a Section-/Feld-Treffer")
	}
}

// TestDetailClickRowMapsSectionHeaderClick guards the Section-Header
// mapping: a click on Section [3] RELATIONS's own header row resolves to
// secIdx=2 (relationsSectionIdx), fieldIdx=-1.
func TestDetailClickRowMapsSectionHeaderClick(t *testing.T) {
	m := detailFocusModel(t)
	b := m.focusedBean()

	msg := detailClickAt(t, m, "[3]") // "> [3] RELATIONS" header marker
	secIdx, fieldIdx, ok := detailClickRow(m, b, msg)
	if !ok {
		t.Fatal("click on the RELATIONS header must resolve")
	}
	if secIdx != relationsSectionIdx {
		t.Fatalf("secIdx = %d, want %d (relationsSectionIdx)", secIdx, relationsSectionIdx)
	}
	if fieldIdx != -1 {
		t.Fatalf("fieldIdx = %d, want -1 (Section-Header hit)", fieldIdx)
	}
}

// metaTagsFieldSubstr disambiguates a click-target search for the Meta
// field list's "tags:" row from the Kopfblock's OWN "tags:" column (B05,
// bean bt-mtig, design-spec.md §15 PF-17): detailHeaderBlock now renders
// "tags: " (single space after the colon) ABOVE the Accordion, so a bare
// "tags:" search finds the Kopfblock first (row < headerBlockLines, never a
// Section-/Feld-Treffer -- see TestDetailClickRowAccountsForFiveLineHeader
// Offset). metaFieldLabels' 12-wide padding (metaSectionBody's
// `fmt.Sprintf("%-12s", ...)`) is unique to the Meta row -- computed here
// via the SAME formula, not a hand-typed literal that could silently drift
// from it.
func metaTagsFieldSubstr() string {
	return fmt.Sprintf("%-12s", "tags:")
}

// TestDetailClickRowMapsMetaFieldClick guards the Meta field-row mapping
// (T1/bt-e6q9's 7-line field list): a click on the "tags:" row resolves to
// secIdx=metaSectionIdx, fieldIdx=4 (metaFields' fixed order: title/status/
// type/priority/tags/created_at/updated_at).
func TestDetailClickRowMapsMetaFieldClick(t *testing.T) {
	m := detailFocusModel(t)
	b := m.focusedBean()

	msg := detailClickAt(t, m, metaTagsFieldSubstr())
	secIdx, fieldIdx, ok := detailClickRow(m, b, msg)
	if !ok {
		t.Fatal("click on the tags: row must resolve")
	}
	if secIdx != metaSectionIdx {
		t.Fatalf("secIdx = %d, want %d (metaSectionIdx)", secIdx, metaSectionIdx)
	}
	if fieldIdx != 4 {
		t.Fatalf("fieldIdx = %d, want 4 (tags, metaFields' fixed order)", fieldIdx)
	}
}

// TestDetailClickRowNoOffByOneWhenRelationsSectionActiveWithFields guards a
// coupling this file's own detailClickRow doc comment (above) flags: B04
// (design-spec.md §15 PF-17, bean bt-b0w0) removes fieldStrip's row from
// renderAccordion's ACTUAL output for the RELATIONS section (its only
// remaining caller) -- detailClickRow's row-counting walk must drop the
// matching skip-row block too, or every section rendered AFTER an
// active+open RELATIONS section (here: HISTORY) resolves one row too low
// and silently returns ok=false. tk-2 (detailFocusModel) has Parent=ep-1,
// so RELATIONS carries >=1 field -- exactly the precondition the old
// fieldStrip-row skip block required (activeSec && i != 0 && len(s.fields)
// > 0).
func TestDetailClickRowNoOffByOneWhenRelationsSectionActiveWithFields(t *testing.T) {
	m := detailFocusModel(t)
	m.detailFocus = true
	m.secCursor = relationsSectionIdx
	m.accOpen = relationsSectionIdx + 1 // opens RELATIONS (PF-1 keeps Meta open too)

	b := m.focusedBean()
	msg := detailClickAt(t, m, "[4]") // "> [4] HISTORY" header marker
	secIdx, fieldIdx, ok := detailClickRow(m, b, msg)
	if !ok {
		t.Fatal("click on the HISTORY header must resolve (row-count off-by-one regression, B04)")
	}
	if secIdx != historySectionIdx {
		t.Fatalf("secIdx = %d, want %d (historySectionIdx)", secIdx, historySectionIdx)
	}
	if fieldIdx != -1 {
		t.Fatalf("fieldIdx = %d, want -1 (Section-Header hit)", fieldIdx)
	}
}

// TestDetailClickRowWrapContinuationLineResolvesToRelationsSectionHit
// guards F04 (bt-b0w0 Review-Findings, Fix-Runde 1): the combination
// "RELATIONS active+open x multi-line wrapping entry" -- B04.3's
// hangingIndentWrap makes a single Relations entry occupy SEVERAL rendered
// rows, and detailClickRow counts section-body height via lipgloss.Height
// (s.body), so both halves must hold: (a) a click on a wrap CONTINUATION
// line inside the Relations body resolves to the section-level hit
// (secIdx=relationsSectionIdx, fieldIdx=-1 -- v1: Relations rows are not
// individually clickable), and (b) the section AFTER the multi-line
// RELATIONS body (HISTORY) resolves without any row shift. The existing
// off-by-one regression test above runs a short, non-wrapping fixture and
// cannot see a wrap-height miscount.
func TestDetailClickRowWrapContinuationLineResolvesToRelationsSectionHit(t *testing.T) {
	beans := fixtureBeans()
	for i := range beans {
		if beans[i].ID == "ep-1" {
			// tk-2's Parent row must wrap: unique WRAPTOKEN lands on a
			// continuation line (render-grounded lookup below fails loudly
			// if it does not render at all).
			beans[i].Title = "Epic One with an intentionally very long title that must wrap across the detail pane WRAPTOKEN end"
		}
	}
	m := fixtureModel(t, beans)
	m = step(t, m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m = focusBean(m, "tk-2")
	m.detailFocus = true
	m.secCursor = relationsSectionIdx
	m.accOpen = relationsSectionIdx + 1
	b := m.focusedBean()

	// Setup sanity: WRAPTOKEN must sit on a CONTINUATION line, not the
	// row's FIRST line -- otherwise this test silently stops covering the
	// wrap-continuation click path. A continuation line carries only
	// hanging indent + title words; the row's Meta prefix (glyphs + the
	// "ep-1" ID) lives exclusively on the first line, so the token's line
	// must NOT contain the ID. ANSI-stripped first (screenLines returns
	// styled lines).
	var onContinuation bool
	for _, raw := range screenLines(m) {
		l := ansi.Strip(raw)
		if strings.Contains(l, "WRAPTOKEN") && !strings.Contains(l, "ep-1") {
			onContinuation = true
		}
	}
	if !onContinuation {
		t.Fatal("setup: WRAPTOKEN must render on a wrap continuation line (a line without the row's own ep-1 ID prefix)")
	}

	msg := detailClickAt(t, m, "WRAPTOKEN")
	secIdx, fieldIdx, ok := detailClickRow(m, b, msg)
	if !ok {
		t.Fatal("click on a Relations wrap continuation line must resolve")
	}
	if secIdx != relationsSectionIdx {
		t.Fatalf("secIdx = %d, want %d (relationsSectionIdx)", secIdx, relationsSectionIdx)
	}
	if fieldIdx != -1 {
		t.Fatalf("fieldIdx = %d, want -1 (v1: section-level hit, Relations rows not individually clickable)", fieldIdx)
	}

	msg2 := detailClickAt(t, m, "[4]") // "> [4] HISTORY" header AFTER the multi-line Relations body
	secIdx2, fieldIdx2, ok2 := detailClickRow(m, b, msg2)
	if !ok2 {
		t.Fatal("click on the HISTORY header below a multi-line RELATIONS body must resolve")
	}
	if secIdx2 != historySectionIdx {
		t.Fatalf("secIdx = %d, want %d (historySectionIdx -- no row shift from the wrapped Relations rows)", secIdx2, historySectionIdx)
	}
	if fieldIdx2 != -1 {
		t.Fatalf("fieldIdx = %d, want -1 (Section-Header hit)", fieldIdx2)
	}
}

// TestMouseDetailClickSectionHeaderActivatesAndExpands guards Fall a (PO-
// Wortlaut, bean bt-duz7): a click on a Section-Header activates AND
// expands that section -- m.detailFocus=true, m.secCursor/m.accOpen set,
// m.detailLevel reset to 0 (section level, not a stale field-level
// carry-over).
func TestMouseDetailClickSectionHeaderActivatesAndExpands(t *testing.T) {
	m := detailFocusModel(t)
	m.detailLevel = 1 // pre-existing field-level state must reset

	msg := detailClickAt(t, m, "[3]")
	tm, _ := m.handleMouse(msg)
	nm, ok := tm.(model)
	if !ok {
		t.Fatalf("handleMouse did not return a model, got %T", tm)
	}
	if !nm.detailFocus {
		t.Fatal("a Section-Header click must set detailFocus=true")
	}
	if nm.secCursor != relationsSectionIdx {
		t.Fatalf("secCursor = %d, want %d (relationsSectionIdx)", nm.secCursor, relationsSectionIdx)
	}
	if nm.accOpen != relationsSectionIdx+1 {
		t.Fatalf("accOpen = %d, want %d", nm.accOpen, relationsSectionIdx+1)
	}
	if nm.detailLevel != 0 {
		t.Fatal("a Section-Header click must reset detailLevel to 0 (section level)")
	}
}

// TestMouseDetailClickSingleClickSelectsField guards Fall b: a single click
// on a Meta field row SELECTS the field (detailLevel=1, fieldCursor set)
// WITHOUT opening any overlay.
func TestMouseDetailClickSingleClickSelectsField(t *testing.T) {
	m := detailFocusModel(t)

	msg := detailClickAt(t, m, metaTagsFieldSubstr())
	tm, _ := m.handleMouse(msg)
	nm, ok := tm.(model)
	if !ok {
		t.Fatalf("handleMouse did not return a model, got %T", tm)
	}
	if !nm.detailFocus || nm.secCursor != metaSectionIdx {
		t.Fatalf("detailFocus=%v secCursor=%d, want true/%d", nm.detailFocus, nm.secCursor, metaSectionIdx)
	}
	if nm.detailLevel != 1 || nm.fieldCursor != 4 {
		t.Fatalf("detailLevel=%d fieldCursor=%d, want 1/4 (tags)", nm.detailLevel, nm.fieldCursor)
	}
	if nm.overlay != overlayNone || nm.form != nil {
		t.Fatal("a single click on a Meta field must NOT open any overlay/form")
	}
}

// TestMouseDetailClickDoubleClickOnFieldOpensOverlay guards Fall c: a
// double click on an already-selected Meta field opens its Overlay --
// identical to the Enter-Kaskade (activateDetailField). Mirrors
// TestDoubleClickTogglesExpand's fixed-clock pattern (above).
func TestMouseDetailClickDoubleClickOnFieldOpensOverlay(t *testing.T) {
	m := detailFocusModel(t)
	fixed := time.Unix(3000, 0)
	m.clock = func() time.Time { return fixed }

	msg := detailClickAt(t, m, metaTagsFieldSubstr())
	tm, _ := m.handleMouse(msg) // 1st click: selects, no overlay
	m2, ok := tm.(model)
	if !ok {
		t.Fatalf("handleMouse did not return a model, got %T", tm)
	}
	if m2.overlay != overlayNone {
		t.Fatal("setup: first click must not open an overlay")
	}

	msg2 := detailClickAt(t, m2, metaTagsFieldSubstr()) // re-render, same field is now selected
	tm2, _ := m2.handleMouse(msg2)                      // 2nd click, same fixed time -> double
	m3, ok := tm2.(model)
	if !ok {
		t.Fatalf("handleMouse did not return a model, got %T", tm2)
	}
	if m3.overlay != overlayTagPicker {
		t.Fatalf("overlay = %v, want overlayTagPicker (double click on tags field, identical to the Enter-Kaskade)", m3.overlay)
	}
	if m3.mutTarget != "tk-2" {
		t.Fatalf("mutTarget = %q, want tk-2", m3.mutTarget)
	}
}

// TestMouseDetailClickDoubleClickOnBodySectionIsNoOpAgain is the B10-
// Revision regression test's mouse counterpart (D01, design-spec.md §15
// PF-17, bean bt-z4b1, mirrors TestKeyDetailFocusEnterOnBodyIsNoOpAgain,
// update_test.go): a double click on BODY's OWN Section-Header used to open
// $EDITOR (bt-y2iw's "Notes for bt-duz7", E8 Task 6/B10) -- that exception
// is REVERTED ersatzlos, same as keyDetailFocus's enter-on-BODY branch, so
// $EDITOR-opening is reserved EXCLUSIVELY for e/ctrl+e (D01). A double click
// on BODY's header is now a plain no-op, exactly like every other section
// header.
func TestMouseDetailClickDoubleClickOnBodySectionIsNoOpAgain(t *testing.T) {
	beans := fixtureBeans()
	for i := range beans {
		if beans[i].ID == "tk-2" {
			beans[i].ETag = "tk-2-etag"
		}
	}
	m := fixtureModel(t, beans)
	m = step(t, m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m = focusBean(m, "tk-2")
	fixed := time.Unix(4000, 0)
	m.clock = func() time.Time { return fixed }

	msg := detailClickAt(t, m, "[2]") // "> [2] BODY" header marker
	tm, _ := m.handleMouse(msg)       // 1st click: selects Section [2]
	m2, ok := tm.(model)
	if !ok {
		t.Fatalf("handleMouse did not return a model, got %T", tm)
	}
	if m2.secCursor != bodySectionIdx || m2.editorTarget != "" {
		t.Fatal("setup: first click must only select Section [2], not open $EDITOR")
	}

	msg2 := detailClickAt(t, m2, "[2]")
	tm2, cmd := m2.handleMouse(msg2) // 2nd click, same fixed time -> double
	m3, ok := tm2.(model)
	if !ok {
		t.Fatalf("handleMouse did not return a model, got %T", tm2)
	}
	if m3.editorTarget != "" || m3.editorETag != "" || m3.editorSnapshot != nil {
		t.Fatal("double click on BODY's header must NEVER open $EDITOR anymore (B10-Revision, D01)")
	}
	if m3.secCursor != bodySectionIdx || m3.detailLevel != 0 {
		t.Fatal("double click on BODY's header must stay a plain section-select no-op")
	}
	if cmd != nil {
		t.Fatal("double click on BODY's header must return a nil Cmd (pure no-op)")
	}
}

// TestDetailClickKeyDisjointNumberSpaces pins the extracted detailClickKey
// helper directly (I01, Fix-Runde 1): every possible Detail-Pane key --
// section headers (fieldIdx -1) and the 7 Meta fields -- must sit at or
// above detailClickKeyBase (disjoint from any realistic Tree/Backlog row
// index) and all 4+7 keys must be pairwise distinct.
func TestDetailClickKeyDisjointNumberSpaces(t *testing.T) {
	seen := map[int]string{}
	check := func(name string, key int) {
		t.Helper()
		if key < detailClickKeyBase {
			t.Fatalf("%s: detailClickKey = %d, want >= %d (must never alias a Tree/Backlog row index, B01)", name, key, detailClickKeyBase)
		}
		if prev, dup := seen[key]; dup {
			t.Fatalf("%s: detailClickKey = %d collides with %s", name, key, prev)
		}
		seen[key] = name
	}
	for sec := 0; sec < beanSectionCount; sec++ {
		check(fmt.Sprintf("section %d header", sec), detailClickKey(sec, -1))
	}
	for fi := 0; fi < 7; fi++ { // metaFields' fixed 7 entries
		check(fmt.Sprintf("meta field %d", fi), detailClickKey(metaSectionIdx, fi))
	}
}

// TestMouseDetailClickTreeClickIndexDoesNotAliasFieldClickKey guards B01
// (Fix-Runde 1, bean bt-duz7): the Detail-Pane clickKey lives in the SAME
// shared m.lastClickIdx int the Tree/Backlog row indices use -- the key must
// therefore occupy a GUARANTEED disjoint number space (detailClickKeyBase
// offset, mouse.go). Reviewer-reproduced bug: a Tree click on row index 1
// ("Epic One" here) followed <500ms by a FIRST click on the Detail pane's
// "title:" field (whose UNOFFSET key was 0*10+0+1 == 1) aliased into a
// false double click and opened the Title-Edit-Form without any prior
// selection. A first click on a not-yet-selected field must ONLY select.
func TestMouseDetailClickTreeClickIndexDoesNotAliasFieldClickKey(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m = step(t, m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m.expanded["ms-1"] = true
	m.expanded["ep-1"] = true // nodes: ms-1(0) ep-1(1) tk-1(2) tk-2(3)
	m.cursorID = "ms-1"
	fixed := time.Unix(5000, 0)
	m.clock = func() time.Time { return fixed }

	// 1st click: Tree row "Epic One" (visibleNodes index 1 == the title:
	// field's OLD, unoffset clickKey).
	msg := treeClickAt(t, m, "Epic One")
	tm, _ := m.handleMouse(msg)
	m2, ok := tm.(model)
	if !ok {
		t.Fatalf("handleMouse did not return a model, got %T", tm)
	}
	if m2.cursorID != "ep-1" || m2.lastClickIdx != 1 {
		t.Fatalf("setup: cursorID=%q lastClickIdx=%d, want ep-1/1 (tree row index)", m2.cursorID, m2.lastClickIdx)
	}

	// 2nd click, same fixed time (<500ms): FIRST click on the Detail pane's
	// title: row -- must ONLY select the field, never open a form/overlay.
	msg2 := detailClickAt(t, m2, "title:")
	tm2, _ := m2.handleMouse(msg2)
	m3, ok := tm2.(model)
	if !ok {
		t.Fatalf("handleMouse did not return a model, got %T", tm2)
	}
	if m3.form != nil {
		t.Fatal("a Tree-row click must never alias a Detail-field clickKey: first click on title: opened the Title-Edit-Form (B01)")
	}
	if m3.overlay != overlayNone {
		t.Fatalf("overlay = %v, want overlayNone (first field click selects only)", m3.overlay)
	}
	if !m3.detailFocus || m3.detailLevel != 1 || m3.fieldCursor != 0 {
		t.Fatalf("detailFocus=%v detailLevel=%d fieldCursor=%d, want true/1/0 (title selected)", m3.detailFocus, m3.detailLevel, m3.fieldCursor)
	}
}

// TestMouseDetailClickSecondClickOnSelectedFieldOpensOverlayOutsideWindow
// guards B02 (Fix-Runde 1, bean bt-duz7): the PO wording names TWO triggers
// -- "Doppelklick (oder Zweitklick auf ein bereits selektiertes Feld)". The
// second trigger must fire INDEPENDENT of the 500ms double-click window
// (the Enter-Kaskade it mirrors is windowless): select field X, wait
// >doubleClickInterval (controlled clock, no real sleep), click X again ->
// its overlay MUST open.
func TestMouseDetailClickSecondClickOnSelectedFieldOpensOverlayOutsideWindow(t *testing.T) {
	m := detailFocusModel(t)
	current := time.Unix(6000, 0)
	m.clock = func() time.Time { return current }

	msg := detailClickAt(t, m, metaTagsFieldSubstr())
	tm, _ := m.handleMouse(msg) // 1st click: selects tags (fieldCursor=4)
	m2, ok := tm.(model)
	if !ok {
		t.Fatalf("handleMouse did not return a model, got %T", tm)
	}
	if m2.overlay != overlayNone || m2.fieldCursor != 4 || m2.detailLevel != 1 {
		t.Fatalf("setup: first click must select tags only (overlay=%v fieldCursor=%d detailLevel=%d)", m2.overlay, m2.fieldCursor, m2.detailLevel)
	}

	current = current.Add(800 * time.Millisecond) // OUTSIDE doubleClickInterval (500ms)

	msg2 := detailClickAt(t, m2, metaTagsFieldSubstr())
	tm2, _ := m2.handleMouse(msg2)
	m3, ok := tm2.(model)
	if !ok {
		t.Fatalf("handleMouse did not return a model, got %T", tm2)
	}
	if m3.overlay != overlayTagPicker {
		t.Fatalf("overlay = %v, want overlayTagPicker -- a second click on an ALREADY-SELECTED field must open its overlay regardless of the double-click window (B02, PO wording)", m3.overlay)
	}
	if m3.mutTarget != "tk-2" {
		t.Fatalf("mutTarget = %q, want tk-2", m3.mutTarget)
	}
}

// TestMouseDetailClickIgnoredWhenOverlayOpen is the Overlay-Guard
// regression check (E5-T4 precedent, verified NOT rebuilt per bt-duz7
// Architektur-Vorgabe #4): a click that would otherwise hit the Detail pane
// is swallowed while an overlay is open, mirrors TestMouseIgnoredWhile
// OverlayOpen (above) but for a Detail-Pane-shaped coordinate.
func TestMouseDetailClickIgnoredWhenOverlayOpen(t *testing.T) {
	m := detailFocusModel(t)
	msg := detailClickAt(t, m, "[3]")
	m.overlay = overlayValueMenu

	tm, _ := m.handleMouse(msg)
	nm, ok := tm.(model)
	if !ok {
		t.Fatalf("handleMouse did not return a model, got %T", tm)
	}
	if nm.secCursor == relationsSectionIdx {
		t.Fatal("a click while an overlay is open must not reach mouseDetailClick's dispatch")
	}
	if nm.overlay != overlayValueMenu {
		t.Fatal("the pre-existing overlay must stay open")
	}
}

// TestMouseDetailClickReachableFromBacklogView guards the Backlog wiring
// (mouseBacklogClick's own delegation, mirrors mouseTreeClick's) AND
// detailClickRow's m.view branch (backlogChrome instead of
// browseRepoChrome) -- a click on the Detail pane while viewBacklog is
// active must reach the SAME mouseDetailClick dispatch.
func TestMouseDetailClickReachableFromBacklogView(t *testing.T) {
	m := fixtureModel(t, backlogBeans())
	m = step(t, m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m.view = viewBacklog
	vis := m.backlogVisible()
	if len(vis) == 0 {
		t.Fatal("setup: need at least 1 backlog-visible bean")
	}
	m.backlogList.setLen(len(vis))
	m.backlogList.cursor = 0

	msg := backlogDetailClickAt(t, m, "[3]") // RELATIONS header
	tm, _ := m.handleMouse(msg)
	nm, ok := tm.(model)
	if !ok {
		t.Fatalf("handleMouse did not return a model, got %T", tm)
	}
	if !nm.detailFocus || nm.secCursor != relationsSectionIdx {
		t.Fatalf("detailFocus=%v secCursor=%d, want true/%d (Backlog view Detail-Pane click)", nm.detailFocus, nm.secCursor, relationsSectionIdx)
	}
}

// TestDetailClickBacklogThreeLineFooterAt80Cols pins clickPaneGeometry's
// dynamic footH (lipgloss.Height(localKeys)+2, mouse.go) against the ONE
// real situation where it differs between the two chromes: the Backlog view
// at 80 columns, whose Q06 footer list + Sort wraps to THREE lines while the
// Browse footer stays at two (bt-d8kc Deviations, "Backlog-Footer 3 Zeilen
// bei 80 Spalten"). Every other clickPaneGeometry/detail-click test runs at
// Width 100, where both footers are 2 lines -- a detailClickRow that picked
// the WRONG chrome (browseRepoChrome instead of backlogChrome) or hardcoded
// a 2-line footer would have passed all of them (E8-T8-Review I01, bean
// bt-6ppq). The pin is the click/render boundary pair: the LAST real body
// row must hit, the pane's bottom-border row (one below) must not -- off by
// one in footH breaks exactly one of the two.
func TestDetailClickBacklogThreeLineFooterAt80Cols(t *testing.T) {
	m := fixtureModel(t, backlogBeans())
	m = step(t, m, tea.WindowSizeMsg{Width: 80, Height: 30})
	m.view = viewBacklog
	vis := m.backlogVisible()
	if len(vis) == 0 {
		t.Fatal("setup: need at least 1 backlog-visible bean")
	}
	m.backlogList.setLen(len(vis))
	m.backlogList.cursor = 0
	// Long body (30 paragraphs — glamour preserves paragraph breaks, so the
	// section body is guaranteed taller than the pane) + BODY section open:
	// the accordion content extends past the pane bottom, so a click below
	// the last body row has real content to (wrongly) resolve to if footH
	// under-counts the 3-line footer.
	vis[0].Body = strings.Repeat("body filler line\n\n", 30)
	m.accOpen = bodySectionIdx + 1

	// Preconditions (fixture reality, not the behavior under test): the
	// Backlog footer at 80 cols spans 3 lines, the Browse footer only 2 --
	// exactly the divergence that gives this test its discriminating power.
	// If a future footer respec changes either height, update the fixture
	// (e.g. a narrower width) so the 3-vs-2 divergence is preserved.
	head, localKeys := m.backlogChrome(m.width - 2)
	if got := lipgloss.Height(localKeys); got != 3 {
		t.Fatalf("precondition: backlog footer at 80 cols = %d lines, want 3", got)
	}
	if _, browseKeys := m.browseRepoChrome(m.width - 2); lipgloss.Height(browseKeys) != 2 {
		t.Fatalf("precondition: browse footer at 80 cols = %d lines, want 2 (divergence vs backlog is what this test exercises)", lipgloss.Height(browseKeys))
	}

	bodyH, lw, _, originX, originY := clickPaneGeometry(m.width, m.height, head, localKeys, m.settings.Layout.TreeWidth)

	// Render-grounded anchor: locate the footer's first line (Q06 order
	// starts with "tab focus in"; renderBindings glues key/desc with NBSP)
	// in the REAL View() and assert the geometry accounts for all three
	// footer lines: pane bottom border + divider sit between the last body
	// row and the footer (the status line renders BELOW the footer), so
	// footerY == originY + bodyH + 2.
	lines := screenLines(m)
	footerY := -1
	for y, l := range lines {
		if strings.Contains(strings.ReplaceAll(l, nbsp, " "), "focus in") {
			footerY = y
			break
		}
	}
	if footerY < 0 {
		t.Fatal("footer first line (\"focus in\") not found in the rendered View()")
	}
	if want := originY + bodyH + 2; footerY != want {
		t.Fatalf("footH dynamic broken: footer first line at row %d, want originY+bodyH+2 = %d", footerY, want)
	}
	borderY := footerY - 2 // pane bottom border (footer -1 divider, -2 border)
	if !strings.Contains(lines[borderY], "╰") {
		t.Fatalf("row %d (footerY-3) is not the pane bottom border: %q", borderY, lines[borderY])
	}

	x := originX + lw + 2 // inside the right Detail pane

	// Last real body row (borderY-1) must resolve: BODY section body hit ->
	// single click activates the section (detailFocus true, secCursor BODY).
	click := tea.MouseMsg{Button: tea.MouseButtonLeft, Action: tea.MouseActionPress, X: x, Y: borderY - 1}
	tm, _ := m.handleMouse(click)
	nm := tm.(model)
	if !nm.detailFocus || nm.secCursor != bodySectionIdx {
		t.Fatalf("click on the LAST body row (Y=%d): detailFocus=%v secCursor=%d, want true/%d (footH over-counts the 3-line footer?)", borderY-1, nm.detailFocus, nm.secCursor, bodySectionIdx)
	}

	// The pane's bottom-border row and the footer row itself must NOT
	// resolve to any Detail hit (clickRow >= bodyH) -- with a 2-line footH
	// assumption the border row falls back INSIDE bodyH and lands in the
	// open BODY body (the bug class this test pins).
	for _, y := range []int{borderY, footerY} {
		click := tea.MouseMsg{Button: tea.MouseButtonLeft, Action: tea.MouseActionPress, X: x, Y: y}
		tm, _ := m.handleMouse(click)
		nm := tm.(model)
		if nm.detailFocus {
			t.Fatalf("click below the pane (Y=%d) must not resolve to a Detail hit (footH under-counts the 3-line footer)", y)
		}
	}
}
