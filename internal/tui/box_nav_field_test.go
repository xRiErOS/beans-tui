package tui

// box_nav_field_test.go — TDD coverage for the box-form FIELD CURSOR (bean
// bt-1o4g, epic bt-vy1q, PO-Nebenbefund N8 "keyboard-first"): while
// boxFormEnabled() AND the Detail pane is focused (tab), the arrow keys must
// step between detailBoxForm's boxes (box_detail_form.go) instead of doing
// nothing, the cursored box must render with dropdownBox/panelBox's
// pre-existing focused=true Mauve frame, and enter must open the SAME editor
// the box's hotkey (and a mouse click, mouse.go) already opens.
//
// Mirrors box_form_scroll_test.go's fixture/drive pattern (boxFormScrollModel,
// step()/keyMsg()) -- the scroll interplay is part of THIS bean's contract
// ("a focused field that scrolls out of view would be a bug"), so the two
// suites deliberately share the long-Body fixture.

import (
	"reflect"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// boxFormFieldModel is boxFormScrollModel with the Detail pane focused --
// the state a real `tab` press produces (update.go's keyTree Focus branch),
// set directly here the same way box_form_scroll_test.go's own tests do.
func boxFormFieldModel(t *testing.T) model {
	t.Helper()
	m := boxFormScrollModel(t)
	m.detailFocus = true
	return m
}

// fieldIdx resolves a boxFormFieldOrder name to its index so the tests below
// read as field NAMES rather than magic numbers (and fail loudly if the order
// table is ever renamed out from under them).
func fieldIdx(t *testing.T, name string) int {
	t.Helper()
	for i, f := range boxFormFieldOrder {
		if f.name == name {
			return i
		}
	}
	t.Fatalf("boxFormFieldOrder has no field named %q", name)
	return -1
}

// --- Order table / geometry single source ---

// TestBoxFormFieldOrderMatchesClickGeometry guards the bean's own
// architecture constraint: the field ORDER must be derived from the same
// hit-test geometry the mouse already uses (detailBoxFormClickRow/gridColAt,
// mouse.go), never a second independent list that can drift. Every field's
// (row, col) must round-trip back to itself through boxFormFieldAt.
func TestBoxFormFieldOrderMatchesClickGeometry(t *testing.T) {
	const accW = 90
	for i, f := range boxFormFieldOrder {
		if f.row > boxFormRowScalarsB {
			continue // panels: unbounded height, no per-row hit-test
		}
		// Middle line of the field's 3-line box, middle column of its cell.
		widths := gridColWidths(boxFormRowCols(f.row), accW)
		col := 0
		for c := 0; c < f.col; c++ {
			col += widths[c] + detailBoxFormGap
		}
		col += widths[f.col] / 2
		row := f.row*3 + 1

		if got := boxFormFieldAt(row, col, accW); got != i {
			t.Fatalf("boxFormFieldAt(row=%d, col=%d) = %d, want %d (%s)", row, col, got, i, f.name)
		}
	}
}

// --- Rendering: the cursored box gets the Mauve frame ---

// TestDetailBoxFormRendersCursoredFieldFocused guards Ziel 3: the cursored
// field renders through dropdownBox/panelBox's pre-existing focused=true
// branch (Mauve frame) -- and NOTHING else changes, i.e. exactly the cursored
// box's own lines differ from the cursor-less render.
func TestDetailBoxFormRendersCursoredFieldFocused(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(termenv.Ascii)

	b := detailBoxFormFixture()
	idx := detailBoxFormIndex()
	const w = 90

	plain := detailBoxForm(idx, b, w, -1)
	for i, f := range boxFormFieldOrder {
		got := detailBoxForm(idx, b, w, i)
		if got == plain {
			t.Fatalf("detailBoxForm(cursor=%d, %s) is byte-identical to the cursor-less render -- the focused frame is not wired through", i, f.name)
		}

		blocks := boxFormBlocks(idx, b, w, -1)
		start, height := boxFormFieldSpan(blocks, i)
		gotLines := strings.Split(got, "\n")
		plainLines := strings.Split(plain, "\n")
		if len(gotLines) != len(plainLines) {
			t.Fatalf("cursor=%d (%s) changed the form's line count: %d vs %d", i, f.name, len(gotLines), len(plainLines))
		}
		for ln := range gotLines {
			inside := ln >= start && ln < start+height
			if !inside && gotLines[ln] != plainLines[ln] {
				t.Fatalf("cursor=%d (%s) changed line %d, which lies OUTSIDE the field's own span [%d,%d)", i, f.name, ln, start, start+height)
			}
		}
	}
}

// TestDetailBoxFormNoCursorIsUnchanged guards the inert default: cursor -1
// must render exactly what the pre-cursor code produced, i.e. every box with
// focused=false -- the flag-on goldens (browse_boxform.golden) depend on it.
func TestDetailBoxFormNoCursorIsUnchanged(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(termenv.Ascii)

	b := detailBoxFormFixture()
	idx := detailBoxFormIndex()
	want := dropdownBox("Title", b.Title, "e", 90, false)
	got := detailBoxForm(idx, b, 90, -1)
	if !strings.HasPrefix(got, want) {
		t.Fatal("detailBoxForm(cursor=-1) must start with the UNFOCUSED Title box (flag-on goldens depend on it)")
	}
}

// TestBoxFormTabRendersFieldCursorInRealView is the PO's own acceptance
// sentence, driven end to end through the REAL pipeline: pressing tab (focus
// into the Detail pane) must make a field VISIBLY focused in m.View() -- a
// Mauve frame that was not on screen a keypress earlier.
func TestBoxFormTabRendersFieldCursorInRealView(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(termenv.Ascii)

	m := boxFormScrollModel(t) // Tree-focused
	b := m.focusedBean()
	accW, _ := boxFormPaneMetrics(m, b)

	// The Title box (first field, the cursor's default) in both frame colors:
	// its value line carries the Mauve/Overlay border glyphs, so it is the
	// discriminator -- other Mauve on screen (pane focus borders, filter
	// chips) can never match this exact rendered line.
	focusedLine := strings.Split(dropdownBox("Title", b.Title, "e", accW, true), "\n")[1]
	plainLine := strings.Split(dropdownBox("Title", b.Title, "e", accW, false), "\n")[1]

	before := m.View()
	if !strings.Contains(before, plainLine) {
		t.Fatal("setup: the UNFOCUSED Title box is not on screen while the Tree has focus")
	}
	if strings.Contains(before, focusedLine) {
		t.Fatal("setup: the Title box already renders focused while the TREE has focus -- the cursor must only show inside the Detail pane")
	}

	m = step(t, m, keyMsg(tea.KeyTab))
	if !m.detailFocus {
		t.Fatal("tab did not move focus into the Detail pane")
	}
	after := m.View()
	if !strings.Contains(after, focusedLine) {
		t.Fatal("after tab the Title box does not render with the Mauve focus frame -- the field cursor is invisible (PO-Nebenbefund N8)")
	}
	if strings.Contains(after, plainLine) {
		t.Fatal("after tab the Title box still renders UNFOCUSED -- the field cursor did not take effect")
	}
}

// --- Keyboard navigation ---

// TestBoxFormArrowsWalkFieldsInOrder guards Ziel 2: down/right/left/up step
// through detailBoxForm's fixed layout -- down/up between rows (keeping the
// column where the target row has one), left/right within a grid row.
func TestBoxFormArrowsWalkFieldsInOrder(t *testing.T) {
	m := boxFormFieldModel(t)

	steps := []struct {
		key  tea.KeyType
		want string
	}{
		{tea.KeyDown, "status"}, // Title -> row A, col 0
		{tea.KeyRight, "type"},  // within row A
		{tea.KeyRight, "priority"},
		{tea.KeyRight, "priority"}, // right edge: no-op
		{tea.KeyDown, "tags"},      // row B has only 2 cols -> col clamps to 1
		{tea.KeyLeft, "parent"},
		{tea.KeyLeft, "parent"}, // left edge: no-op
		{tea.KeyUp, "status"},   // back into row A at col 0
		{tea.KeyUp, "title"},
		{tea.KeyUp, "title"}, // top edge: no-op
	}
	for i, s := range steps {
		m = step(t, m, keyMsg(s.key))
		if got := boxFormFieldOrder[m.boxFormCursor].name; got != s.want {
			t.Fatalf("step %d (%v): cursor on %q, want %q", i, s.key, got, s.want)
		}
	}
}

// TestBoxFormDownReachesPanelFields walks all the way down and asserts the
// panel fields (Body/Relations/History) are reachable and that the walk stops
// at History rather than wrapping.
func TestBoxFormDownReachesPanelFields(t *testing.T) {
	m := boxFormFieldModel(t)
	seen := map[string]bool{}
	for i := 0; i < 300; i++ {
		m = step(t, m, keyMsg(tea.KeyDown))
		seen[boxFormFieldOrder[m.boxFormCursor].name] = true
	}
	for _, name := range []string{"status", "parent", "body", "relations", "history"} {
		if !seen[name] {
			t.Fatalf("field %q was never reached by repeated `down` presses", name)
		}
	}
	if got := boxFormFieldOrder[m.boxFormCursor].name; got != "history" {
		t.Fatalf("after 300 downs the cursor sits on %q, want the LAST field %q (no wrap-around)", got, "history")
	}
}

// --- Scroll interplay (the bean's explicit "must not scroll out of view") ---

// TestBoxFormCursorStaysVisibleWhileNavigating is the headline interplay
// criterion: after EVERY down press the cursored field's own line span must
// intersect the visible window [boxFormScroll, boxFormScroll+height) -- a
// focused field that scrolled out of the pane would be a bug.
func TestBoxFormCursorStaysVisibleWhileNavigating(t *testing.T) {
	m := boxFormFieldModel(t)
	b := m.focusedBean()
	requireOverflow(t, m, b)

	accW, height := boxFormPaneMetrics(m, b)
	blocks := boxFormBlocks(m.idx, b, accW, -1)

	for i := 0; i < 300; i++ {
		m = step(t, m, keyMsg(tea.KeyDown))
		off := boxFormEffectiveScroll(m, b)
		start, h := boxFormFieldSpan(blocks, m.boxFormCursor)
		if start >= off+height || start+h <= off {
			t.Fatalf("down #%d: field %q spans [%d,%d) but the visible window is [%d,%d) -- the cursor scrolled out of view",
				i, boxFormFieldOrder[m.boxFormCursor].name, start, start+h, off, off+height)
		}
	}
}

// TestBoxFormDownScrollsThroughATallField guards the "reveal before move"
// rule: while the cursored field is TALLER than the pane, down scrolls the
// viewport through it instead of jumping the cursor onward -- otherwise a
// long Body's tail would be keyboard-unreachable (bt-ze10's own headline
// criterion, preserved through this bean's key-handling change).
func TestBoxFormDownScrollsThroughATallField(t *testing.T) {
	m := boxFormFieldModel(t)
	b := m.focusedBean()
	requireOverflow(t, m, b)

	body := fieldIdx(t, "body")
	for i := 0; i < 10 && m.boxFormCursor < body; i++ {
		m = step(t, m, keyMsg(tea.KeyDown))
	}
	if m.boxFormCursor != body {
		t.Fatalf("setup: cursor on %q, want body", boxFormFieldOrder[m.boxFormCursor].name)
	}

	before := m.boxFormScroll
	m = step(t, m, keyMsg(tea.KeyDown))
	if m.boxFormCursor != body {
		t.Fatalf("down on the tall Body field moved the cursor to %q -- it must first scroll the rest of the field into view", boxFormFieldOrder[m.boxFormCursor].name)
	}
	if m.boxFormScroll <= before {
		t.Fatalf("boxFormScroll = %d, want > %d (down inside a tall field scrolls)", m.boxFormScroll, before)
	}
	if m.boxFormScrollBean != b.ID {
		t.Fatalf("boxFormScrollBean = %q, want %q", m.boxFormScrollBean, b.ID)
	}

	up := m.boxFormScroll
	m = step(t, m, keyMsg(tea.KeyUp))
	if m.boxFormScroll >= up {
		t.Fatalf("boxFormScroll after up = %d, want < %d (up inside a tall field scrolls back)", m.boxFormScroll, up)
	}
}

// --- enter opens the field's editor (the SAME path a click takes) ---

// TestBoxFormEnterOpensFieldEditor guards Ziel 4 for every field kind: enter
// on the cursored box fires the EXACT action mouseBoxFormDetailClick fires
// for the same box (activateBoxFormTarget, mouse.go) -- one dispatch, no
// parallel path.
func TestBoxFormEnterOpensFieldEditor(t *testing.T) {
	cases := []struct {
		field string
		check func(t *testing.T, m model, beanID string)
	}{
		{"status", func(t *testing.T, m model, _ string) {
			if m.overlay != overlayValueMenu || !reflect.DeepEqual(m.menuItems, buildValueMenuItems("status")) {
				t.Fatalf("enter on Status: overlay=%v, want the status value menu", m.overlay)
			}
		}},
		{"type", func(t *testing.T, m model, _ string) {
			if m.overlay != overlayValueMenu || !reflect.DeepEqual(m.menuItems, buildValueMenuItems("type")) {
				t.Fatalf("enter on Type: overlay=%v, want the type value menu", m.overlay)
			}
		}},
		{"priority", func(t *testing.T, m model, _ string) {
			if m.overlay != overlayValueMenu || !reflect.DeepEqual(m.menuItems, buildValueMenuItems("priority")) {
				t.Fatalf("enter on Priority: overlay=%v, want the priority value menu", m.overlay)
			}
		}},
		{"parent", func(t *testing.T, m model, _ string) {
			if m.overlay != overlayParentPicker {
				t.Fatalf("enter on Parent: overlay=%v, want overlayParentPicker", m.overlay)
			}
		}},
		{"tags", func(t *testing.T, m model, _ string) {
			if m.overlay != overlayTagPicker {
				t.Fatalf("enter on Tags: overlay=%v, want overlayTagPicker", m.overlay)
			}
		}},
		{"title", func(t *testing.T, m model, id string) {
			if m.editorTarget != id {
				t.Fatalf("enter on Title: editorTarget=%q, want %q ($EDITOR on the whole bean)", m.editorTarget, id)
			}
		}},
		{"body", func(t *testing.T, m model, id string) {
			if m.editorTarget != id {
				t.Fatalf("enter on Body: editorTarget=%q, want %q", m.editorTarget, id)
			}
		}},
	}

	for _, tc := range cases {
		t.Run(tc.field, func(t *testing.T) {
			m := boxFormFieldModel(t)
			b := m.focusedBean()
			m.boxFormCursor = fieldIdx(t, tc.field)
			m.boxFormCursorBean = b.ID

			m = step(t, m, keyMsg(tea.KeyEnter))
			tc.check(t, m, b.ID)
		})
	}
}

// --- Reset on bean change ---

// TestBoxFormCursorResetsWhenSelectedBeanChanges guards "Cursor resettet bei
// Bean-Wechsel": a cursor recorded for one bean must never carry over into a
// different bean's render.
func TestBoxFormCursorResetsWhenSelectedBeanChanges(t *testing.T) {
	m := boxFormFieldModel(t)
	b := m.focusedBean()

	m.boxFormCursor = fieldIdx(t, "tags")
	m.boxFormCursorBean = b.ID

	m = focusBean(m, "tk-1")
	other := m.focusedBean()
	if other == nil || other.ID == b.ID {
		t.Fatalf("setup: expected a different focused bean, got %v", other)
	}
	if got := boxFormEffectiveCursor(m, other); got != 0 {
		t.Fatalf("boxFormEffectiveCursor(new bean) = %d, want 0 (stale cursor must not leak)", got)
	}
}

// --- Flag off: accordion mode is untouched ---

// TestBoxFormFieldNavInertWithoutFlag guards the experiment's hard contract:
// without BT_BOXFORM, the same keys must drive the pre-existing accordion
// Section/Field machine and never write the box-form cursor state.
func TestBoxFormFieldNavInertWithoutFlag(t *testing.T) {
	m := fixtureModel(t, boxFormLongBodyBeans()) // BT_BOXFORM left unset
	m = step(t, m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m = focusBean(m, "tk-2")
	m.detailFocus = true

	m = step(t, m, keyMsg(tea.KeyDown))
	m = step(t, m, keyMsg(tea.KeyRight))
	m = step(t, m, keyMsg(tea.KeyEnter))
	if m.boxFormCursor != 0 || m.boxFormCursorBean != "" {
		t.Fatalf("boxFormCursor/boxFormCursorBean = %d/%q, want 0/\"\" (accordion mode must never touch the box-form cursor)", m.boxFormCursor, m.boxFormCursorBean)
	}
}
