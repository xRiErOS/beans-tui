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

// TestBoxFormGridMoveWalksTheLayout guards the GRID geometry of the field
// table (boxFormMoveField): up/down between rows keeping the column where the
// target row has one, left/right within a grid row, edges are no-ops.
//
// bean bt-8d35 (PO-Entscheidung 3): this is no longer driven by the ARROW
// KEYS -- those went back to plain viewport scrolling and tab/shift+tab took
// over the field cursor (linear order + wrap, see TestBoxFormTabWalksFields
// WithWrap below). The grid mover itself is deliberately NOT torn out
// ("kein Voll-Rueckbau", bt-8d35's own wording) and stays covered here at the
// boxFormNav level, one rung below the keymap.
func TestBoxFormGridMoveWalksTheLayout(t *testing.T) {
	m := boxFormFieldModel(t)
	b := m.focusedBean()

	steps := []struct {
		dir  string
		want string
	}{
		{"down", "status"}, // Title -> row A, col 0
		{"right", "type"},  // within row A
		{"right", "priority"},
		{"right", "priority"}, // right edge: no-op
		{"down", "tags"},      // row B has only 2 cols -> col clamps to 1
		{"left", "parent"},
		{"left", "parent"}, // left edge: no-op
		{"up", "status"},   // back into row A at col 0
		{"up", "title"},
		{"up", "title"}, // top edge: no-op
	}
	for i, s := range steps {
		m = m.boxFormNav(b, s.dir)
		if got := boxFormFieldOrder[m.boxFormCursor].name; got != s.want {
			t.Fatalf("step %d (%s): cursor on %q, want %q", i, s.dir, got, s.want)
		}
	}
}

// --- bean bt-8d35: tab/shift+tab own the field cursor, arrows scroll ---

// TestBoxFormTabWalksFieldsWithWrap is the PO's Flow #2 verbatim: tab steps
// through detailBoxForm's fields in RENDER order and WRAPS at the end instead
// of falling back into the Tree ("sonst verliert man den Fokus versehentlich
// beim Durchsteppen").
func TestBoxFormTabWalksFieldsWithWrap(t *testing.T) {
	m := boxFormFieldModel(t)

	want := make([]string, 0, len(boxFormFieldOrder))
	for _, f := range boxFormFieldOrder[1:] {
		want = append(want, f.name)
	}
	want = append(want, boxFormFieldOrder[0].name) // wrap back to Title

	for i, w := range want {
		m = step(t, m, keyMsg(tea.KeyTab))
		if got := boxFormFieldOrder[m.boxFormCursor].name; got != w {
			t.Fatalf("tab #%d: cursor on %q, want %q", i+1, got, w)
		}
		if !m.detailFocus {
			t.Fatalf("tab #%d left the Detail region -- tab must move WITHIN it (bt-8d35)", i+1)
		}
	}
}

// TestBoxFormShiftTabWalksFieldsBackwardsWithWrap is Flow #2's other half:
// shift+tab is the reverse step, and from the FIRST field it wraps to the
// last rather than exiting the region (that is esc's job now).
func TestBoxFormShiftTabWalksFieldsBackwardsWithWrap(t *testing.T) {
	m := boxFormFieldModel(t)

	m = step(t, m, keyMsg(tea.KeyShiftTab))
	if got, want := boxFormFieldOrder[m.boxFormCursor].name, boxFormFieldOrder[len(boxFormFieldOrder)-1].name; got != want {
		t.Fatalf("shift+tab on the first field: cursor on %q, want the LAST field %q (wrap)", got, want)
	}
	if !m.detailFocus {
		t.Fatal("shift+tab left the Detail region -- only esc may do that (bt-8d35)")
	}

	m = step(t, m, keyMsg(tea.KeyShiftTab))
	if got, want := boxFormFieldOrder[m.boxFormCursor].name, boxFormFieldOrder[len(boxFormFieldOrder)-2].name; got != want {
		t.Fatalf("second shift+tab: cursor on %q, want %q", got, want)
	}
}

// TestBoxFormArrowsScrollInsteadOfMovingTheCursor is PO-Entscheidung 3:
// the arrows scroll the WHOLE view again (bt-ze10's behaviour) and leave the
// field cursor exactly where tab put it.
func TestBoxFormArrowsScrollInsteadOfMovingTheCursor(t *testing.T) {
	m := boxFormFieldModel(t)
	b := m.focusedBean()
	requireOverflow(t, m, b)

	before := m.boxFormCursor
	m = step(t, m, keyMsg(tea.KeyDown))
	if m.boxFormCursor != before {
		t.Fatalf("down moved the field cursor to %q -- the arrows must only scroll (bt-8d35)", boxFormFieldOrder[m.boxFormCursor].name)
	}
	if m.boxFormScroll != 1 {
		t.Fatalf("boxFormScroll after one down = %d, want 1 (plain ±1 viewport scroll)", m.boxFormScroll)
	}
	m = step(t, m, keyMsg(tea.KeyUp))
	if m.boxFormScroll != 0 {
		t.Fatalf("boxFormScroll after up = %d, want 0", m.boxFormScroll)
	}
}

// TestBoxFormEscLeavesTheDetailRegion pins the other half of bt-8d35's rule
// -- "esc VERLAESST die Region" -- now that shift+tab no longer does.
func TestBoxFormEscLeavesTheDetailRegion(t *testing.T) {
	m := boxFormFieldModel(t)
	m = step(t, m, keyMsg(tea.KeyEsc))
	if m.detailFocus {
		t.Fatal("esc did not leave the Detail region")
	}
}

// TestBoxFormTabFromTreeEntersAtTheFirstField guards Flow #2's entry step:
// tab out of the Tree focuses the Detail region AT TITLE, even when a stale
// cursor from a previous visit to the same bean is still on the model.
func TestBoxFormTabFromTreeEntersAtTheFirstField(t *testing.T) {
	m := boxFormScrollModel(t) // Tree-focused
	b := m.focusedBean()
	m.boxFormCursor, m.boxFormCursorBean = fieldIdx(t, "tags"), b.ID

	m = step(t, m, keyMsg(tea.KeyTab))
	if !m.detailFocus {
		t.Fatal("tab did not focus the Detail region")
	}
	if got := boxFormEffectiveCursor(m, b); got != 0 {
		t.Fatalf("tab-in landed on field %q, want the first field %q", boxFormFieldOrder[got].name, boxFormFieldOrder[0].name)
	}
}

// TestBoxFormTabWalksFieldsInTheBacklogToo is bt-8d35's "Reichweite" check:
// tab is a GLOBAL binding, so the Fokus-Modell must land identically in every
// view that renders the box form -- the Backlog shares the very same Detail
// pane and routes through the same keyDetailFocus (handleKey's detailFocus
// dispatch), so this pins that it really does, rather than only Browse.
func TestBoxFormTabWalksFieldsInTheBacklogToo(t *testing.T) {
	t.Setenv("BT_BOXFORM", "1")
	m := fixtureModel(t, backlogBeans())
	m = step(t, m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m.view = viewBacklog

	m = step(t, m, keyMsg(tea.KeyTab)) // into the Detail region
	if !m.detailFocus {
		t.Fatal("setup: tab did not focus the Backlog's Detail pane")
	}
	if got := boxFormEffectiveCursor(m, m.focusedBean()); got != 0 {
		t.Fatalf("tab-in landed on field %q, want Title", boxFormFieldOrder[got].name)
	}

	m = step(t, m, keyMsg(tea.KeyTab))
	if got := boxFormFieldOrder[m.boxFormCursor].name; got != boxFormFieldOrder[1].name {
		t.Fatalf("second tab in the Backlog: cursor on %q, want %q", got, boxFormFieldOrder[1].name)
	}
	if !m.detailFocus {
		t.Fatal("tab swapped the Backlog's panes instead of moving within the Detail region")
	}
}

// TestTabStillSwapsPanesWithoutBoxForm is the flag-OFF regression pin: the
// whole Fokus-Modell is experiment-gated (epic bt-vy1q), so without
// BT_BOXFORM tab remains PF-13's bidirectional pane toggle and shift+tab
// remains the deterministic exit.
func TestTabStillSwapsPanesWithoutBoxForm(t *testing.T) {
	t.Setenv("BT_BOXFORM", "")
	m := fixtureModel(t, fixtureBeans())
	m = step(t, m, tea.WindowSizeMsg{Width: 100, Height: 30})

	m = step(t, m, keyMsg(tea.KeyTab))
	if !m.detailFocus {
		t.Fatal("flag OFF: tab must still focus the Detail pane")
	}
	m = step(t, m, keyMsg(tea.KeyTab))
	if m.detailFocus {
		t.Fatal("flag OFF: a second tab must toggle focus back to the Tree (PF-13)")
	}
	m = step(t, m, keyMsg(tea.KeyTab))
	m = step(t, m, keyMsg(tea.KeyShiftTab))
	if m.detailFocus {
		t.Fatal("flag OFF: shift+tab must still be the deterministic exit (PF-13)")
	}
}

// --- Scroll interplay (the bean's explicit "must not scroll out of view") ---

// TestBoxFormCursorStaysVisibleWhileNavigating is the headline interplay
// criterion: after EVERY tab press the cursored field's own line span must
// intersect the visible window [boxFormScroll, boxFormScroll+height) -- a
// focused field that scrolled out of the pane would be a bug.
//
// bean bt-8d35: unchanged in substance, driven by tab instead of down (the
// Scroll-Mitnahme moved with the field cursor onto tab, "Springt tab auf ein
// Feld ausserhalb des sichtbaren Bereichs, muss der Viewport nachziehen").
func TestBoxFormCursorStaysVisibleWhileNavigating(t *testing.T) {
	m := boxFormFieldModel(t)
	b := m.focusedBean()
	requireOverflow(t, m, b)

	accW, height := boxFormPaneMetrics(m, b)
	blocks := boxFormBlocks(m.idx, b, accW, -1)

	for i := 0; i < 300; i++ {
		m = step(t, m, keyMsg(tea.KeyTab))
		off := boxFormEffectiveScroll(m, b)
		start, h := boxFormFieldSpan(blocks, m.boxFormCursor)
		if start >= off+height || start+h <= off {
			t.Fatalf("tab #%d: field %q spans [%d,%d) but the visible window is [%d,%d) -- the cursor scrolled out of view",
				i, boxFormFieldOrder[m.boxFormCursor].name, start, start+h, off, off+height)
		}
	}
}

// TestBoxFormDownScrollsThroughATallField guards bt-ze10's own headline
// criterion: while the cursored field is TALLER than the pane, down scrolls
// the viewport through it instead of jumping the cursor onward -- otherwise a
// long Body's tail would be keyboard-unreachable.
//
// bean bt-8d35: the cursor is walked onto the Body with tab (the arrows no
// longer move it); the assertions themselves are unchanged, they hold now
// because the arrows scroll again (PO-Entscheidung 3).
func TestBoxFormDownScrollsThroughATallField(t *testing.T) {
	m := boxFormFieldModel(t)
	b := m.focusedBean()
	requireOverflow(t, m, b)

	body := fieldIdx(t, "body")
	for i := 0; i < 10 && m.boxFormCursor < body; i++ {
		m = step(t, m, keyMsg(tea.KeyTab))
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
