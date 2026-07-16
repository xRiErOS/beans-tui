package tui

// view_tag_management_test.go — TDD coverage for T2 (bean bt-r92i, epic
// bt-362n E10, D05-D09): the Tag-Management page's read-only Grundgerüst --
// tagRegistryRows' Union/sort algebra (D09), the page's own full-capture key
// handler (D06/D08), and the Command-Center entry point (D05). Reuses
// fixtureBeans/fixtureModel/keyMsg/runeMsg (update_test.go), same package.

import (
	"fmt"
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

// --- T3 (bean bt-604w, epic bt-362n D11/D14): Create + shared free-text
// input sub-mode ---

// TestKeyTagMgmtInputRejectsInvalidName is the bean's own named RED test
// (bt-604w TDD section) -- quoted almost verbatim. ERRATUM vs. the bean's
// literal sketch: `m, _ = m.openTagMgmtInput(...)` does not compile
// (openTagMgmtInput returns tea.Model, an interface, into `m` which is the
// concrete `model` type -- the SAME assignability rule every other
// tea.Model-returning method in this codebase already requires an explicit
// type assertion for, e.g. TestOpenTagManagementPageNilClientBuildsFromIdxOnly
// above: `nm, cmd := m.openTagManagementPage(); mm := nm.(model)`). Fixed by
// routing through a fresh `nm`/type-assertion, same shape as every other
// test in this file -- assertions themselves are UNCHANGED from the bean's
// own wording (mirrors bt-pqq3's own precedent of fixing a test-sketch bug
// while leaving the asserted contract untouched, documented in ERRATA).
func TestKeyTagMgmtInputRejectsInvalidName(t *testing.T) {
	m := newModel(nil, "")
	m.view = viewTagManagement
	nm, _ := m.openTagMgmtInput("create", "")
	m = nm.(model)
	m.tagMgmtInput.SetValue("Not Valid!")
	nm2, _ := m.keyTagMgmtInput(tea.KeyMsg{Type: tea.KeyEnter})
	got := nm2.(model)
	if !got.tagMgmtInputActive || got.tagMgmtInputErr == "" {
		t.Fatalf("want input to stay open with an error, got active=%v err=%q",
			got.tagMgmtInputActive, got.tagMgmtInputErr)
	}
}

// TestKeyTagMgmtInputRejectsDuplicateAgainstExistingRows is the bean's own
// named RED test (bt-604w TDD section) -- same ERRATUM fix as
// TestKeyTagMgmtInputRejectsInvalidName above (fresh var + type assertion).
func TestKeyTagMgmtInputRejectsDuplicateAgainstExistingRows(t *testing.T) {
	m := newModel(nil, "")
	m.view = viewTagManagement
	m.tagMgmtRows = []tagRegistryRow{{name: "already-there", defined: true}}
	nm, _ := m.openTagMgmtInput("create", "")
	m = nm.(model)
	m.tagMgmtInput.SetValue("already-there")
	nm2, _ := m.keyTagMgmtInput(tea.KeyMsg{Type: tea.KeyEnter})
	if got := nm2.(model); !got.tagMgmtInputActive || got.tagMgmtInputErr == "" {
		t.Fatalf("want rejected duplicate, got %+v", got)
	}
}

// TestKeyTagMgmtInputRejectsDuplicateAgainstFreeRowToo pins the bean body's
// own explicit wording ("Dedupe-Check gegen die aktuelle m.tagMgmtRows-
// Namensmenge ... Dedupe prüft also gegen ALLE vorhandenen Namen" -- ALL
// existing names, not just registry-defined ones): typing the exact name of
// an already-visible FREE (undefined) row is ALSO rejected as a duplicate,
// not silently treated as a "promote this free tag" shortcut (D09's own
// promotion narrative is a documented, not-yet-built fast-follow, epic
// bt-362n Q-section -- retyping an existing free tag's name here is not that
// mechanism).
func TestKeyTagMgmtInputRejectsDuplicateAgainstFreeRowToo(t *testing.T) {
	m := newModel(nil, "")
	m.view = viewTagManagement
	m.tagMgmtRows = []tagRegistryRow{{name: "free-tag", count: 3, defined: false}}
	nm, _ := m.openTagMgmtInput("create", "")
	m = nm.(model)
	m.tagMgmtInput.SetValue("free-tag")
	nm2, _ := m.keyTagMgmtInput(tea.KeyMsg{Type: tea.KeyEnter})
	if got := nm2.(model); !got.tagMgmtInputActive || got.tagMgmtInputErr == "" {
		t.Fatalf("want rejected duplicate against a FREE row too, got %+v", got)
	}
}

// TestKeyTagManagementNewTagOpensInput guards D14's entry point: `n`
// (keys.NewTag, reused verbatim from the Tag-Picker, same binding/meaning)
// opens the shared free-text input sub-mode in "create" mode, with an EMPTY
// target (D11: Create never prefills/targets an old name -- only T5/Rename
// will).
func TestKeyTagManagementNewTagOpensInput(t *testing.T) {
	m := newModel(nil, "")
	m.view = viewTagManagement

	nm, cmd := m.keyTagManagement(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	mm := nm.(model)
	if !mm.tagMgmtInputActive {
		t.Fatal("want tagMgmtInputActive=true after 'n'")
	}
	if mm.tagMgmtInputMode != "create" {
		t.Fatalf("tagMgmtInputMode = %q, want %q", mm.tagMgmtInputMode, "create")
	}
	if mm.tagMgmtInputTarget != "" {
		t.Fatalf("tagMgmtInputTarget = %q, want empty (D11: Create never targets an old name)", mm.tagMgmtInputTarget)
	}
	if !mm.tagMgmtInput.Focused() {
		t.Fatal("want tagMgmtInput focused after opening")
	}
	if cmd == nil {
		t.Fatal("want a non-nil Cmd (textinput.Blink), mirrors openTagInput")
	}
}

// TestKeyTagMgmtInputEscDiscardsOnlySubmode guards D14/D06: esc while the
// input is active closes ONLY the input sub-mode -- the Page itself
// (m.view) and its row list stay completely untouched, mirrors keyTagInput's
// own esc contract (box_picker_tag.go doc-stamp: "only the input sub-mode
// itself is discarded").
func TestKeyTagMgmtInputEscDiscardsOnlySubmode(t *testing.T) {
	m := newModel(nil, "")
	m.view = viewTagManagement
	m.tagMgmtRows = []tagRegistryRow{{name: "existing", defined: true}}
	m = step(t, m, runeMsg('n'))
	if !m.tagMgmtInputActive {
		t.Fatal("setup: want input active after 'n'")
	}
	m = step(t, m, runeMsg('x'))
	if m.tagMgmtInput.Value() != "x" {
		t.Fatalf("setup: want 'x' typed into the input, got %q", m.tagMgmtInput.Value())
	}

	m = step(t, m, keyMsg(tea.KeyEsc))
	if m.tagMgmtInputActive {
		t.Fatal("esc must close the input sub-mode")
	}
	if m.view != viewTagManagement {
		t.Fatalf("esc in the input sub-mode must stay on the Page, view = %v", m.view)
	}
	if len(m.tagMgmtRows) != 1 || m.tagMgmtRows[0].name != "existing" {
		t.Fatalf("esc must not touch tagMgmtRows, got %+v", m.tagMgmtRows)
	}
}

// TestKeyTagMgmtInputCapturesEveryKeyNoLeak is the Full-Capture-Disziplin
// regression this task's own harness brief explicitly demands: while the
// input sub-mode is active, EVERY other key (including the global node-
// action set AND `?`/ctrl+k) belongs to the textinput, not to any outer
// handler -- mirrors TestHandleKeyOnTagManagementViewDoesNotLeakToNodeAction's
// D06 guard one layer up, here for D14's own nested capture state.
func TestKeyTagMgmtInputCapturesEveryKeyNoLeak(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m = focusBean(m, "tk-2")
	m.view = viewTagManagement
	m = step(t, m, runeMsg('n'))
	if !m.tagMgmtInputActive {
		t.Fatal("setup: want input active after 'n'")
	}

	for _, r := range "d?" {
		m = step(t, m, runeMsg(r))
	}
	m = step(t, m, keyMsg(tea.KeyCtrlK))

	if m.overlay != overlayNone {
		t.Fatalf("'d' while typing must not open the Delete-Confirm, overlay = %v", m.overlay)
	}
	if m.helpOpen {
		t.Fatal("'?' while typing must not open Help")
	}
	if m.paletteOpen {
		t.Fatal("ctrl+k while typing must not open the Command-Center")
	}
	if !m.tagMgmtInputActive || m.view != viewTagManagement {
		t.Fatalf("want to stay in the input sub-mode on the Page, active=%v view=%v", m.tagMgmtInputActive, m.view)
	}
	if got, want := m.tagMgmtInput.Value(), "d?"; got != want {
		t.Fatalf("want every captured key typed into the input, got %q, want %q", got, want)
	}
}

// TestKeyTagMgmtInputRetypingNDoesNotReopen guards the 'n'-as-a-literal-
// -character edge case: keys.NewTag's own binding key ('n') must be
// swallowed by keyTagMgmtInput's OWN default branch (textinput.Update) once
// the sub-mode is already active, not re-dispatched to
// keyTagManagement's outer `case keybind.Matches(msg, keys.NewTag)` (which
// only exists OUTSIDE the `if m.tagMgmtInputActive` guard).
func TestKeyTagMgmtInputRetypingNDoesNotReopen(t *testing.T) {
	m := newModel(nil, "")
	m.view = viewTagManagement
	m = step(t, m, runeMsg('n'))
	m = step(t, m, runeMsg('n')) // literal 'n' as part of a tag name, e.g. "new-ish"

	if m.tagMgmtInput.Value() != "n" {
		t.Fatalf("want the second 'n' typed into the input, got value=%q", m.tagMgmtInput.Value())
	}
}

// TestKeyTagMgmtInputValidSubmitFiresSaveTagDefsCmdWithAddedName guards D11's
// Create dispatch: a valid, non-duplicate name fires saveTagDefsCmd with the
// registry's CURRENT defined names (extracted from tagMgmtRows) plus the new
// one (data.AddTagDefName) -- proven by intercepting the returned Cmd
// (mirrors TestEditTitleSubmitFiresSetTitleDirectlyNoConfirm's own
// "inspect before it fires" pattern, form_edit_title_test.go) against a REAL
// t.TempDir()-backed *data.Client, doubling as the Persistenz-Roundtrip
// requirement: after cmd() runs, the on-disk .beans-tags.yml actually
// contains the new name.
func TestKeyTagMgmtInputValidSubmitFiresSaveTagDefsCmdWithAddedName(t *testing.T) {
	dir := t.TempDir()
	m := newModel(&data.Client{RepoDir: dir}, dir)
	m.view = viewTagManagement
	m.tagMgmtRows = []tagRegistryRow{{name: "old-defined", defined: true}, {name: "free-one", defined: false}}

	nm, _ := m.openTagMgmtInput("create", "")
	m = nm.(model)
	m.tagMgmtInput.SetValue("brand-new")

	tm, cmd := m.keyTagMgmtInput(tea.KeyMsg{Type: tea.KeyEnter})
	mm := tm.(model)
	if mm.tagMgmtInputErr != "" {
		t.Fatalf("want no validation error, got %q", mm.tagMgmtInputErr)
	}
	if cmd == nil {
		t.Fatal("valid submit must return a non-nil Cmd (saveTagDefsCmd)")
	}

	msg := cmd()
	tdm, ok := msg.(tagDefsSavedMsg)
	if !ok {
		t.Fatalf("cmd() = %T, want tagDefsSavedMsg", msg)
	}
	if tdm.err != nil {
		t.Fatalf("SaveTagDefs against a real t.TempDir() client failed: %v", tdm.err)
	}

	got, err := (&data.Client{RepoDir: dir}).LoadTagDefs()
	if err != nil {
		t.Fatalf("LoadTagDefs after save: %v", err)
	}
	want := []string{"brand-new", "old-defined"} // AddTagDefName, sorted -- "free-one" was never a registry entry
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("on-disk registry = %v, want %v (persistence round trip)", got, want)
	}
}

// TestApplyTagDefsSavedSuccessRefreshesRowsAndMovesCursor guards
// applyTagDefsSaved's success tail: the input sub-mode closes, tagMgmtRows
// is rebuilt from a FRESH LoadTagDefs (D03) via tagRegistryRows, and the
// cursor lands on the newly created row (mirrors applyLoaded's own
// cursor-refinding-by-ID pattern one layer up).
func TestApplyTagDefsSavedSuccessRefreshesRowsAndMovesCursor(t *testing.T) {
	dir := t.TempDir()
	client := &data.Client{RepoDir: dir}
	if err := client.SaveTagDefs([]string{"old-defined", "brand-new"}); err != nil {
		t.Fatalf("setup SaveTagDefs: %v", err)
	}

	m := newModel(client, dir)
	m.view = viewTagManagement
	m.tagMgmtInputActive = true
	m.tagMgmtInputMode = "create"
	m.tagMgmtInput.SetValue("brand-new")
	m.tagMgmtInput.Focus()
	m.tagMgmtRows = []tagRegistryRow{{name: "old-defined", defined: true}}
	m.tagMgmtCursor.setLen(1)

	nm, cmd := m.applyTagDefsSaved(tagDefsSavedMsg{err: nil})
	mm := nm.(model)
	if cmd != nil {
		t.Fatalf("success must not return a Cmd (no unconditional reload, only the Registry changed), got %v", cmd)
	}
	if mm.tagMgmtInputActive {
		t.Fatal("success must close the input sub-mode")
	}
	if mm.tagMgmtInput.Focused() {
		t.Fatal("success must blur the input")
	}

	wantNames := []string{"brand-new", "old-defined"}
	var gotNames []string
	for _, r := range mm.tagMgmtRows {
		gotNames = append(gotNames, r.name)
	}
	if !reflect.DeepEqual(gotNames, wantNames) {
		t.Fatalf("tagMgmtRows names = %v, want %v", gotNames, wantNames)
	}
	if mm.tagMgmtCursor.cursor < 0 || mm.tagMgmtCursor.cursor >= len(mm.tagMgmtRows) || mm.tagMgmtRows[mm.tagMgmtCursor.cursor].name != "brand-new" {
		t.Fatalf("cursor must land on the new row, cursor=%d rows=%+v", mm.tagMgmtCursor.cursor, mm.tagMgmtRows)
	}
}

// TestApplyTagDefsSavedErrorKeepsInputOpenShowsToast guards the error tail
// (mirrors applyMutationResult's error branch, WITHOUT its unconditional
// loadCmd reload -- "hier gibt es kein m.idx zu invalidieren, nur die
// Registry", bean bt-604w wording): the input sub-mode stays open (so the
// PO can retry the same name without retyping it) and a Toast surfaces the
// error.
func TestApplyTagDefsSavedErrorKeepsInputOpenShowsToast(t *testing.T) {
	m := newModel(&data.Client{RepoDir: "/nonexistent-bt-604w-scratch-dir"}, "/nonexistent-bt-604w-scratch-dir")
	m.view = viewTagManagement
	m.tagMgmtInputActive = true
	m.tagMgmtInputMode = "create"
	m.tagMgmtInput.SetValue("brand-new")

	nm, cmd := m.applyTagDefsSaved(tagDefsSavedMsg{err: fmt.Errorf("write .beans-tags.yml: no such file or directory")})
	mm := nm.(model)
	if !mm.tagMgmtInputActive {
		t.Fatal("an I/O failure must keep the input sub-mode open for a retry")
	}
	if cmd == nil {
		t.Fatal("want a non-nil Cmd (the error Toast)")
	}
	if mm.toast == nil || mm.toast.kind != toastError {
		t.Fatalf("want an error Toast, got %+v", mm.toast)
	}
}

// TestFullCreateFlowRefreshesPageAndTouchesNoBean is the end-to-end
// behavioral test the harness brief demands ("Page-Refresh nach Create") --
// drives the ENTIRE wiring through m.Update() (handleKey -> keyTagManagement
// -> keyTagMgmtInput -> saveTagDefsCmd -> cmd() -> Update(tagDefsSavedMsg)
// -> applyTagDefsSaved), then asserts D11's own regression requirement: a
// Create never touches m.idx (no Bean is mutated, only the Registry).
func TestFullCreateFlowRefreshesPageAndTouchesNoBean(t *testing.T) {
	dir := t.TempDir()
	m := fixtureModel(t, fixtureBeans())
	m.client = &data.Client{RepoDir: dir}
	m.repoDir = dir
	m.view = viewTagManagement
	nm, _ := m.openTagManagementPage()
	m = nm.(model)
	idxBefore := m.idx

	m = step(t, m, runeMsg('n'))
	for _, r := range "released" {
		m = step(t, m, runeMsg(r))
	}

	tm, cmd := m.Update(keyMsg(tea.KeyEnter))
	m = tm.(model)
	if cmd == nil {
		t.Fatal("enter on a valid name must return the saveTagDefsCmd Cmd")
	}
	msg := cmd()

	tm2, _ := m.Update(msg)
	m = tm2.(model)

	if m.tagMgmtInputActive {
		t.Fatal("Page must exit the input sub-mode after a successful create")
	}
	found := false
	for _, r := range m.tagMgmtRows {
		if r.name == "released" && r.defined {
			found = true
		}
	}
	if !found {
		t.Fatalf("want 'released' to appear as a defined row after Create, got %+v", m.tagMgmtRows)
	}
	if m.idx != idxBefore {
		t.Fatal("D11: Create must not touch m.idx (no Bean mutation, Registry-only)")
	}
	for _, b := range m.idx.ByID {
		for _, tag := range b.Tags {
			if tag == "released" {
				t.Fatalf("D11: Create must not touch any Bean, but %s now carries 'released'", b.ID)
			}
		}
	}
}

// TestTagManagementLocalBindingsIncludesNewTag guards the Footer Zone 3
// wiring: tagManagementLocalBindings (shared function body, T2's own
// "EIN gemeinsamer Funktionsrumpf" contract) grows keys.NewTag alongside
// Up/Down/Back.
func TestTagManagementLocalBindingsIncludesNewTag(t *testing.T) {
	bindings := tagManagementLocalBindings()
	found := false
	for _, b := range bindings {
		if b.Help().Key == keys.NewTag.Help().Key {
			found = true
		}
	}
	if !found {
		t.Fatalf("tagManagementLocalBindings missing keys.NewTag, got %+v", bindings)
	}
}

// TestViewTagManagementRendersInputSubmode guards the (bean-implicit, but
// tmux-smoke-mandatory) render contract: while the input sub-mode is active,
// viewTagManagement's output shows the live input value AND an inline error
// (when set) INSIDE the SAME full-capture page (D06/D14 -- "lebt INNERHALB
// der Full-Capture-Page"), not a floating overlay/modalPanel like the
// Tag-Picker's own tagInputBox.
func TestViewTagManagementRendersInputSubmode(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m.width, m.height = 100, 30
	nm, _ := m.openTagManagementPage()
	m = nm.(model)
	nm2, _ := m.openTagMgmtInput("create", "")
	m = nm2.(model)
	m.tagMgmtInput.SetValue("typed-so-far")
	m.tagMgmtInputErr = "invalid tag name (a-z0-9, hyphen-separated, lowercase)"

	out := m.viewTagManagement()
	if !strings.Contains(out, "typed-so-far") {
		t.Errorf("viewTagManagement() while input active missing the typed value:\n%s", out)
	}
	if !strings.Contains(out, "invalid tag name") {
		t.Errorf("viewTagManagement() while input active missing the inline error:\n%s", out)
	}
}

// TestViewTagManagementInputSubmodeNoWrapAt80 pins the tmux-smoke acceptance
// wording verbatim ("Bei 80 Spalten: Inline-Fehlertext darf nicht truncaten
// ohne …") -- an oversized error message at width=80 must be cut with an
// ellipsis (renderPane's own per-row truncate(w-2) budget, view.go's
// truncate helper), never silently wrapped/overflowed (same B01 bug class
// TestViewTagManagementLinesFitExactlyNoWrap already pins for the normal row
// list).
func TestViewTagManagementInputSubmodeNoWrapAt80(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m.width, m.height = 80, 40
	nm, _ := m.openTagManagementPage()
	m = nm.(model)
	nm2, _ := m.openTagMgmtInput("create", "")
	m = nm2.(model)
	m.tagMgmtInputErr = strings.Repeat("a very long inline validation error message ", 5)

	out := m.viewTagManagement()
	lines := strings.Split(out, "\n")
	if len(lines) != m.height {
		t.Fatalf("got %d lines, want exactly %d (height) -- an unbudgeted error line broke the pane's line count", len(lines), m.height)
	}
	for i, l := range lines {
		if w := lipgloss.Width(l); w != m.width {
			t.Fatalf("line %d width = %d, want exactly %d: %q", i, w, m.width, l)
		}
	}
}
