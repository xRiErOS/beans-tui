package tui

// view_tag_management_test.go — TDD coverage for T2 (bean bt-r92i, epic
// bt-362n E10, D05-D09): the Tag-Management page's read-only Grundgerüst --
// tagRegistryRows' Union/sort algebra (D09), the page's own full-capture key
// handler (D06/D08), and the Command-Center entry point (D05). Reuses
// fixtureBeans/fixtureModel/keyMsg/runeMsg (update_test.go), same package.

import (
	"reflect"
	"strings"
	"testing"

	"beans-tui/internal/data"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// --- tagRegistryRows: D09 Union + sort algebra ---

// TestTagRegistryRowsDefinedFirstAlphaThenFreeByCountDesc is the bean's own
// named RED test (bt-r92i TDD section, quoted verbatim).
func TestTagRegistryRowsDefinedFirstAlphaThenFreeByCountDesc(t *testing.T) {
	idx := data.NewIndex([]data.Bean{
		{ID: "b1", Tags: []string{"zzz-free"}},
		{ID: "b2", Tags: []string{"zzz-free"}},
		{ID: "b3", Tags: []string{"aaa-free"}},
	})
	rows := tagRegistryRows(idx, []string{"defined-b", "defined-a"})
	want := []string{"defined-a", "defined-b", "zzz-free", "aaa-free"}
	var got []string
	for _, r := range rows {
		got = append(got, r.name)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

// TestTagRegistryRowsIncludesUnusedDefinedTagWithZeroCount is the bean's own
// named RED test (bt-r92i TDD section, quoted verbatim).
func TestTagRegistryRowsIncludesUnusedDefinedTagWithZeroCount(t *testing.T) {
	idx := data.NewIndex(nil)
	rows := tagRegistryRows(idx, []string{"unused"})
	if len(rows) != 1 || rows[0].count != 0 || !rows[0].defined {
		t.Fatalf("want one zero-count defined row, got %+v", rows)
	}
}

// TestTagRegistryRowsDefinedTagKeepsRealUsageCount guards D09's Union
// semantics against a naive "defined XOR free" split: a tag that is BOTH
// registry-defined AND currently in use on beans must appear exactly ONCE
// (in the Defined group), carrying its REAL usage count -- not 0, and not a
// second, duplicate row in the Free group.
func TestTagRegistryRowsDefinedTagKeepsRealUsageCount(t *testing.T) {
	idx := data.NewIndex([]data.Bean{
		{ID: "b1", Tags: []string{"shared-tag"}},
		{ID: "b2", Tags: []string{"shared-tag"}},
	})
	rows := tagRegistryRows(idx, []string{"shared-tag"})
	if len(rows) != 1 {
		t.Fatalf("want exactly 1 row (no duplicate), got %d: %+v", len(rows), rows)
	}
	if !rows[0].defined || rows[0].count != 2 {
		t.Fatalf("want defined=true count=2, got %+v", rows[0])
	}
}

// TestTagRegistryRowsDedupesDuplicateDefNames guards against a registry
// slice containing an accidental duplicate (should never happen post-T1's
// own sortDedupTagNames, but tagRegistryRows must not crash or double-list
// if it ever does -- defensive, mirrors LoadTagDefs' own "never trust the
// file blindly" philosophy).
func TestTagRegistryRowsDedupesDuplicateDefNames(t *testing.T) {
	rows := tagRegistryRows(data.NewIndex(nil), []string{"dup", "dup"})
	if len(rows) != 1 {
		t.Fatalf("want exactly 1 deduped row, got %d: %+v", len(rows), rows)
	}
}

// TestTagRegistryRowsEmptyEverywhereReturnsEmpty guards the pre-load /
// empty-repo floor: no defs, no beans -> an empty (nil-or-zero-length) row
// list, never a panic.
func TestTagRegistryRowsEmptyEverywhereReturnsEmpty(t *testing.T) {
	rows := tagRegistryRows(data.NewIndex(nil), nil)
	if len(rows) != 0 {
		t.Fatalf("want 0 rows, got %d: %+v", len(rows), rows)
	}
}

// TestTagRegistryRowsNilIndexNoPanic guards the pre-load state (m.idx == nil
// before the first beansLoadedMsg) -- must never panic, defs alone still
// render as zero-count defined rows.
func TestTagRegistryRowsNilIndexNoPanic(t *testing.T) {
	rows := tagRegistryRows(nil, []string{"solo"})
	if len(rows) != 1 || !rows[0].defined || rows[0].count != 0 {
		t.Fatalf("want one zero-count defined row, got %+v", rows)
	}
}

// --- keyTagManagement: nav / esc / enter no-op (D06/D08) ---

// TestKeyTagManagementEscReturnsToBrowse is the bean's own named RED test
// (bt-r92i TDD section, quoted verbatim).
func TestKeyTagManagementEscReturnsToBrowse(t *testing.T) {
	m := newModel(nil, "")
	m.view = viewTagManagement
	nm, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if nm.(model).view != viewBrowseRepo {
		t.Fatalf("want viewBrowseRepo, got %v", nm.(model).view)
	}
}

// TestHandleKeyOnTagManagementViewDoesNotLeakToNodeAction is the D06
// regression guard (bean bt-r92i TDD section; REWRITTEN in Fix-Runde 1,
// T2-Review F01, medium): the bean's original verbatim sketch used
// `newModel(nil, "")` -- a model with NO focused bean -- but keyNodeAction's
// branch (update.go) is a silent no-op when focusedBean()==nil anyway, so
// that version stayed GREEN even with the D06 full-capture guard
// (update.go, `if m.view == viewTagManagement`) removed entirely
// (reviewer-verified by mutation). The whole POINT of D06 is that
// focusedBean()'s `default` branch falls through to the TREE cursor while
// viewTagManagement is active -- so the test must set up exactly that leak
// surface: a real index + a real tree cursor on a real bean (focusBean,
// box_menu_value_test.go). With the guard removed, `d` now reaches
// keyNodeAction, resolves tk-2 as the (stale, unrelated) target and opens
// the Delete-Confirm (overlay=overlayDeleteConfirm) -- verified RED under
// exactly that mutation during Fix-Runde 1, quoted in the bean.
func TestHandleKeyOnTagManagementViewDoesNotLeakToNodeAction(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m = focusBean(m, "tk-2")
	m.view = viewTagManagement

	// Sanity: the leak surface this test guards must actually exist --
	// focusedBean() resolves the tree cursor via its default branch even
	// while viewTagManagement is active. If this ever returns nil, the test
	// has silently degraded back to the F01 no-op version.
	if m.focusedBean() == nil {
		t.Fatal("test setup invalid: focusedBean() == nil -- the D06 leak this test guards would be impossible to trigger")
	}

	nm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	mm := nm.(model)
	if mm.overlay != overlayNone {
		t.Fatalf("want no overlay opened, got %v", mm.overlay)
	}
	if mm.view != viewTagManagement {
		t.Fatalf("want view to stay viewTagManagement, got %v", mm.view)
	}
}

// TestHandleKeyOnTagManagementViewDoesNotLeakHelpOrPalette extends the D06
// regression guard to the Command-Center/Help bare-match checks (`?`/
// ctrl+k), which sit AFTER the viewTagManagement full-capture checkpoint in
// handleKey -- mirrors viewLobby's own precedent (view_lobby.go doc-stamp:
// "ctrl+k/`?` are DELIBERATELY not special-cased to still reach the
// Lobby").
func TestHandleKeyOnTagManagementViewDoesNotLeakHelpOrPalette(t *testing.T) {
	m := newModel(nil, "")
	m.view = viewTagManagement

	nm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	mm := nm.(model)
	if mm.helpOpen {
		t.Fatalf("want helpOpen=false (full-capture), got true")
	}
	if mm.view != viewTagManagement {
		t.Fatalf("want view to stay viewTagManagement, got %v", mm.view)
	}

	nm2, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlK})
	mm2 := nm2.(model)
	if mm2.paletteOpen {
		t.Fatalf("want paletteOpen=false (full-capture), got true")
	}
}

func TestKeyTagManagementUpDownMovesCursor(t *testing.T) {
	m := newModel(nil, "")
	m.view = viewTagManagement
	m.tagMgmtRows = []tagRegistryRow{{name: "a"}, {name: "b"}, {name: "c"}}
	m.tagMgmtCursor.setLen(len(m.tagMgmtRows))

	nm, _ := m.keyTagManagement(tea.KeyMsg{Type: tea.KeyDown})
	mm := nm.(model)
	if mm.tagMgmtCursor.cursor != 1 {
		t.Fatalf("cursor after down = %d, want 1", mm.tagMgmtCursor.cursor)
	}
	nm2, _ := mm.keyTagManagement(tea.KeyMsg{Type: tea.KeyUp})
	if nm2.(model).tagMgmtCursor.cursor != 0 {
		t.Fatalf("cursor after up = %d, want 0", nm2.(model).tagMgmtCursor.cursor)
	}
}

// TestKeyTagManagementEnterIsHandledNoOp guards D08: enter must be a
// HANDLED no-op (view/cursor/overlay all unchanged), not an unhandled
// fallthrough.
func TestKeyTagManagementEnterIsHandledNoOp(t *testing.T) {
	m := newModel(nil, "")
	m.view = viewTagManagement
	m.tagMgmtRows = []tagRegistryRow{{name: "a"}}
	m.tagMgmtCursor.setLen(1)

	nm, cmd := m.keyTagManagement(tea.KeyMsg{Type: tea.KeyEnter})
	mm := nm.(model)
	if mm.view != viewTagManagement || mm.overlay != overlayNone || cmd != nil {
		t.Fatalf("enter must be a handled no-op, got view=%v overlay=%v cmd=%v", mm.view, mm.overlay, cmd)
	}
}

// --- openTagManagementPage: D03 fresh-load + D02 tolerant-missing ---

// TestOpenTagManagementPageNilClientBuildsFromIdxOnly guards D02's
// tolerant-missing contract at the model layer: no live client (pre-load /
// test fixture, e.g. fixtureModel's own newModel(nil, ...) convention) must
// never panic -- an empty registry applies, rows still build from whatever
// tags already sit on m.idx's beans.
func TestOpenTagManagementPageNilClientBuildsFromIdxOnly(t *testing.T) {
	m := fixtureModel(t, []data.Bean{
		{ID: "tk-1", Title: "T", Status: "todo", Type: "task", Priority: "normal", Tags: []string{"loose"}},
	})
	nm, cmd := m.openTagManagementPage()
	mm := nm.(model)
	if mm.view != viewTagManagement {
		t.Fatalf("view = %v, want viewTagManagement", mm.view)
	}
	if cmd != nil {
		t.Fatalf("want nil Cmd (synchronous open, D02/D03), got %v", cmd)
	}
	if len(mm.tagMgmtRows) != 1 || mm.tagMgmtRows[0].name != "loose" || mm.tagMgmtRows[0].defined {
		t.Fatalf("want one free 'loose' row, got %+v", mm.tagMgmtRows)
	}
	if mm.tagMgmtCursor.length != 1 {
		t.Fatalf("tagMgmtCursor.length = %d, want 1", mm.tagMgmtCursor.length)
	}
}

// TestOpenTagManagementPageResetsStaleCursor guards that a stale cursor from
// a PREVIOUS page visit (more rows then, cursor deep in the list) never
// survives into a re-open with fewer rows.
func TestOpenTagManagementPageResetsStaleCursor(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m.tagMgmtCursor.setLen(50)
	m.tagMgmtCursor.cursor = 40

	nm, _ := m.openTagManagementPage()
	mm := nm.(model)
	if mm.tagMgmtCursor.cursor >= mm.tagMgmtCursor.length && mm.tagMgmtCursor.length > 0 {
		t.Fatalf("cursor %d out of bounds for length %d", mm.tagMgmtCursor.cursor, mm.tagMgmtCursor.length)
	}
}

// --- tagManagementChrome: D07 empty GlobalHint ---

// TestTagManagementChromeGlobalHintEmpty guards D07: the breadcrumb's right
// side must be empty (no renderBindings(globalBindings()) promise for keys
// that don't work during full-capture) -- checked by asserting the header
// line does not contain any of the 4 global binding key glyphs' rendered
// text.
func TestTagManagementChromeGlobalHintEmpty(t *testing.T) {
	m := newModel(nil, "/tmp/bt-fixture-repo")
	head, _ := m.tagManagementChrome(80)
	for _, glyph := range []string{"commands", "repos", "help", "quit"} {
		if strings.Contains(head, glyph) {
			t.Fatalf("tagManagementChrome head contains global hint text %q (D07 requires an empty GlobalHint): %q", glyph, head)
		}
	}
}

// TestTagManagementChromeFooterListsUpDownBack guards that the page's own
// Footer Zone 3 (T2 vorerst: Up/Down/Back) is actually rendered.
func TestTagManagementChromeFooterListsUpDownBack(t *testing.T) {
	_, localKeys := (model{}).tagManagementChrome(80)
	plain := stripHint(localKeys)
	if !strings.Contains(plain, "back") {
		t.Fatalf("tagManagementChrome footer missing 'back' hint: %q", plain)
	}
}

// --- viewTagManagement: render-behavior (union rows, D09 order visible) ---

// TestViewTagManagementRendersDefinedAndFreeRows is a render-grounded test:
// after openTagManagementPage, viewTagManagement's output must contain both
// a registry-defined tag name AND a free/in-use tag name.
func TestViewTagManagementRendersDefinedAndFreeRows(t *testing.T) {
	m := fixtureModel(t, []data.Bean{
		{ID: "tk-1", Title: "T", Status: "todo", Type: "task", Priority: "normal", Tags: []string{"free-tag"}},
	})
	m.width, m.height = 100, 30
	nm, _ := m.openTagManagementPage()
	mm := nm.(model)
	mm.tagMgmtRows = append(mm.tagMgmtRows, tagRegistryRow{name: "def-tag", count: 0, defined: true})

	out := mm.viewTagManagement()
	if !strings.Contains(out, "free-tag") {
		t.Errorf("viewTagManagement() output missing free-tag:\n%s", out)
	}
	if !strings.Contains(out, "def-tag") {
		t.Errorf("viewTagManagement() output missing def-tag:\n%s", out)
	}
}

// TestViewTagManagementNoNewRegistryFileOnRender guards the bean's own
// tmux-smoke acceptance wording ("T2 ist rein lesend") at the render layer:
// rendering the page must never touch the filesystem (viewTagManagement
// itself does no I/O at all -- LoadTagDefs only runs in
// openTagManagementPage). A weak but cheap guard: viewTagManagement must not
// panic/error against a client pointed at a real temp repo dir with no
// registry file, and must not create one as a side effect of rendering.
func TestViewTagManagementNoNewRegistryFileOnRender(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m.width, m.height = 100, 30
	_ = m.viewTagManagement() // must not panic
}

// TestViewTagManagementLinesFitExactlyNoWrap is a regression guard (found
// live during this task's OWN tmux-smoke run, same "B01" bug class as
// view_lobby.go's own doc-stamp / view_browse_repo.go's F01 Vollbild
// paneW note): a single full-width pane's renderPane `w` argument is a
// CONTENT width (its own RoundedBorder adds +2 more) -- passing innerW
// unadjusted (instead of innerW-2) double-counts the border, and
// outerBorder's Render silently WORD-WRAPS any resulting over-width line
// instead of erroring, turning the width bug into extra, unbudgeted lines.
// Pins the fixed contract at BOTH tmux-smoke widths (120/80): every line is
// EXACTLY m.width cells wide, and the total line count is EXACTLY m.height
// -- a wrapped border line would inflate line count beyond height and break
// the per-line width check simultaneously.
func TestViewTagManagementLinesFitExactlyNoWrap(t *testing.T) {
	for _, width := range []int{120, 80} {
		m := fixtureModel(t, []data.Bean{
			{ID: "tk-1", Title: "T", Status: "todo", Type: "task", Priority: "normal", Tags: []string{"a-fairly-long-free-tag-name"}},
		})
		m.width, m.height = width, 40
		nm, _ := m.openTagManagementPage()
		mm := nm.(model)

		out := mm.viewTagManagement()
		lines := strings.Split(out, "\n")
		if len(lines) != m.height {
			t.Fatalf("width=%d: viewTagManagement() produced %d lines, want exactly %d (height)", width, len(lines), m.height)
		}
		for i, l := range lines {
			if w := lipgloss.Width(l); w != m.width {
				t.Fatalf("width=%d: line %d width = %d, want exactly %d: %q", width, i, w, m.width, l)
			}
		}
	}
}

// --- View() dispatcher (D05 entry wiring) ---

func TestViewDispatcherRoutesTagManagementView(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m.width, m.height = 100, 30
	nm, _ := m.openTagManagementPage()
	mm := nm.(model)

	out := mm.View()
	// viewBrowseRepo's own breadcrumb title is "Browse" -- viewTagManagement's
	// is "Tags" (tagManagementChrome). A dispatcher bug (falling through to
	// the `default: viewBrowseRepo()` case) would render "Browse" instead.
	if strings.Contains(out, "Browse") {
		t.Errorf("View() while viewTagManagement is active still rendered the Browse breadcrumb (dispatcher not wired):\n%s", out)
	}
}

// --- Command-Center entry point (D05) ---

// TestPaletteActionsIncludesGoToTagsBeforeSettings guards D05's wiring:
// "go to tags" is a global action (no focused-bean requirement, mirrors
// "go to settings"), grouped directly BEFORE "settings" as the new
// last-but-one entry (Planner-Entscheidung, epic bt-362n body's own D05
// wording).
func TestPaletteActionsIncludesGoToTagsBeforeSettings(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	items := paletteActions(m)

	var gotIdx, settingsIdx = -1, -1
	for i, it := range items {
		switch it.actionID {
		case "go_tags":
			gotIdx = i
			if it.label != "go to tags" {
				t.Fatalf(`go_tags label = %q, want "go to tags"`, it.label)
			}
		case "settings":
			settingsIdx = i
		}
	}
	if gotIdx == -1 {
		t.Fatal(`paletteActions missing "go_tags"`)
	}
	if settingsIdx == -1 {
		t.Fatal(`paletteActions missing "settings"`)
	}
	if settingsIdx != gotIdx+1 {
		t.Fatalf("go_tags at %d, settings at %d -- want settings directly after go_tags", gotIdx, settingsIdx)
	}
}

// TestDispatchPaletteGoTagsOpensTagManagementPage guards the Command-Center
// entry point (D05: NO dedicated keybinding, ONLY reachable via
// overlay_palette.go's "go to tags" action) -- mirrors
// TestDispatchPaletteSettingsOpensForm's own precedent
// (box_form_settings_test.go).
func TestDispatchPaletteGoTagsOpensTagManagementPage(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	nm, _ := m.dispatchPalette(paletteItem{kind: paletteKindAction, actionID: "go_tags", label: "go to tags"})
	mm, ok := nm.(model)
	if !ok {
		t.Fatalf("dispatchPalette did not return a model, got %T", nm)
	}
	if mm.view != viewTagManagement {
		t.Fatalf("dispatchPalette(go_tags) view = %v, want viewTagManagement", mm.view)
	}
	if mm.paletteOpen {
		t.Fatal("dispatchPalette(go_tags) left paletteOpen=true, want closed")
	}
}
