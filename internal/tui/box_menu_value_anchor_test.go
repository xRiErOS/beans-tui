package tui

// box_menu_value_anchor_test.go — TDD coverage for Slice C (bt-f0y9,
// "feld-verankertes Inline-Dropdown", D09 revidiert): the value menu
// (Status/Type/Priority) opens field-anchored (placeOverlayAt +
// boxFormFieldRect) instead of centered (placeOverlay), while
// boxFormEnabled(). Every render-grounded assertion here locates a REAL
// landmark (the popup's own bottom-border corner, or its title text) in
// ansi.Strip(m.View()) and compares its position to what boxFormFieldRect
// (Slice A) + valueMenuPopupY (this slice's own pure function, unit-tested
// separately below) compute -- never a hand-typed pixel guess, and never a
// "re-render with overlay closed" baseline (footer text itself depends on
// m.overlay, footer_context.go's overlayLocalBindings -- that would silently
// compare against a DIFFERENT chrome, not just a different overlay layer).

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

// --- valueMenuPopupY (pure function) ---

func TestValueMenuPopupYFitsBelow(t *testing.T) {
	got := valueMenuPopupY(5, 3, 10, 30)
	if want := 8; got != want {
		t.Fatalf("valueMenuPopupY(fits below) = %d, want %d", got, want)
	}
}

func TestValueMenuPopupYFlipsAboveWhenBelowOverflowsAndFullyFits(t *testing.T) {
	// anchorY=20, anchorH=3 -> below=23, +fgH(10)=33 > canvasH(25): overflow.
	// above = 20-10 = 10 >= 0: flips, fits fully.
	got := valueMenuPopupY(20, 3, 10, 25)
	if want := 10; got != want {
		t.Fatalf("valueMenuPopupY(flip, full fit) = %d, want %d", got, want)
	}
}

func TestValueMenuPopupYFlipsAboveEvenWithPartialClip(t *testing.T) {
	// anchorY=10, anchorH=3 -> below=13, +fgH(12)=25 > canvasH(20): overflow.
	// above = 10-12 = -2: still flips (anchorY>0, "some room above") even
	// though the result itself would clip the popup's own top rows --
	// placeOverlayAt's pre-existing silent-overflow-drop (Slice B) handles
	// that same as it already handles every other overlay in this codebase.
	got := valueMenuPopupY(10, 3, 12, 20)
	if want := -2; got != want {
		t.Fatalf("valueMenuPopupY(flip, partial clip) = %d, want %d", got, want)
	}
}

func TestValueMenuPopupYNoRoomAboveFallsBackToBelow(t *testing.T) {
	// anchorY=0: the field sits at the very top, NO room above at all
	// (PO-Vorgabe "nur wenn oben auch kein Platz: clampen") -- falls back to
	// "below", clipped at the bottom same as any other overflow.
	got := valueMenuPopupY(0, 3, 10, 8)
	if want := 3; got != want {
		t.Fatalf("valueMenuPopupY(no room above) = %d, want %d", got, want)
	}
}

// --- boxFormFieldIndexForGroup ---

func TestBoxFormFieldIndexForGroupMatchesFieldOrder(t *testing.T) {
	cases := []struct{ group, field string }{
		{"status", "status"},
		{"type", "type"},
		{"priority", "priority"},
	}
	for _, tc := range cases {
		if got, want := boxFormFieldIndexForGroup(tc.group), fieldIdx(t, tc.field); got != want {
			t.Fatalf("boxFormFieldIndexForGroup(%q) = %d, want %d", tc.group, got, want)
		}
	}
}

func TestBoxFormFieldIndexForGroupUnknownReturnsNegativeOne(t *testing.T) {
	if got := boxFormFieldIndexForGroup("bogus"); got != -1 {
		t.Fatalf("boxFormFieldIndexForGroup(bogus) = %d, want -1", got)
	}
}

// --- Render-grounded: anchored placement, one test per trigger path ---

// anchorModel builds a BT_BOXFORM=1 fixtureModel at the given size, tk-2
// focused (Task Two, ms-1->ep-1->tk-2 ancestry expanded, box_menu_value_test.
// go's focusBean), Detail-Pane focused (tab-equivalent) -- the state every
// one of the four trigger paths needs at minimum.
func anchorModel(t *testing.T, width, height int) model {
	t.Helper()
	t.Setenv("BT_BOXFORM", "1")
	m := fixtureModel(t, fixtureBeans())
	m = step(t, m, tea.WindowSizeMsg{Width: width, Height: height})
	m = focusBean(m, "tk-2")
	m.detailFocus = true
	return m
}

// cellAt returns the rune (as a string, ANSI-safe) at the given column of an
// already ansi.Strip()'d line -- ansi.Cut(l, col, col+1), mirroring Slice A's
// own box_nav_field_test.go corner-check pattern (never strings.Index, which
// would find the FIRST occurrence of a glyph ANYWHERE on the row -- the
// background chrome (e.g. the Tree pane's own Search sub-box) can render the
// SAME "╰"/"╭" box-drawing characters at an earlier column than the popup's
// own corner).
func cellAt(l string, col int) string {
	return ansi.Cut(l, col, col+1)
}

// assertValueMenuAnchoredBelowField is the shared render-grounded assertion
// every trigger-path test below drives: m (with the value menu already open,
// via whichever trigger) must have its popup's bottom-border corner at
// EXACTLY (x, valueMenuPopupY(rect.y, rect.h, fgH, m.height)+fgH-1) -- where
// rect = boxFormFieldRect(m, mutTarget-bean, wantField)'s own (x, y, h) --
// proving composeOverlays' new anchored branch actually ran with the RIGHT
// field's geometry, not a leftover centered placeOverlay (which would put
// the corner at a completely different column/row for any non-trivial
// terminal size).
func assertValueMenuAnchoredBelowField(t *testing.T, m model, wantField string) {
	t.Helper()
	if m.overlay != overlayValueMenu {
		t.Fatalf("overlay = %v, want overlayValueMenu", m.overlay)
	}
	wantIdx := fieldIdx(t, wantField)
	if m.valueMenuAnchorField != wantIdx {
		t.Fatalf("valueMenuAnchorField = %d, want %d (%s)", m.valueMenuAnchorField, wantIdx, wantField)
	}
	b := m.idx.ByID[m.mutTarget]
	if b == nil {
		t.Fatalf("setup: mutTarget %q not in index", m.mutTarget)
	}
	rx, ry, _, rh, ok := boxFormFieldRect(m, b, wantIdx)
	if !ok {
		t.Fatalf("boxFormFieldRect(%s) ok=false", wantField)
	}

	fg := m.valueMenuBox()
	fgLines := strings.Split(fg, "\n")
	fgW, fgH := maxLineWidth(fgLines), len(fgLines)
	y := valueMenuPopupY(ry, rh, fgH, m.height)
	wantRow := y + fgH - 1

	// wantCol mirrors placeValueMenuOverlay's OWN right-edge clamp
	// (box_menu_value.go) -- rx unclamped would run the ~42-cell-wide popup
	// off an 80-column canvas for Type/Priority (anchored further right than
	// Status), so the production code deliberately shifts x left to stay
	// on-canvas. Re-deriving that same, simple, already-documented formula
	// here is not circular: it is the ONLY way to compute the CORRECT
	// expected column, the same way the "below"/"above" Y choice needed
	// valueMenuPopupY itself.
	wantCol := rx
	if fgW < m.width && wantCol+fgW > m.width {
		wantCol = m.width - fgW
	}
	if wantCol < 0 {
		wantCol = 0
	}

	lines := screenLines(m)
	if wantRow < 0 || wantRow >= len(lines) {
		t.Fatalf("m.View() (field=%s): wantRow=%d out of range (%d lines) (rect=(%d,%d,_,%d), popupY=%d, fgH=%d)", wantField, wantRow, len(lines), rx, ry, rh, y, fgH)
	}
	if got := cellAt(lines[wantRow], wantCol); got != "╰" {
		t.Fatalf("m.View() (field=%s): cell at (row=%d,col=%d) = %q, want \"╰\" (popup bottom-left corner; rect=(%d,%d,_,%d), popupY=%d, row=%q)", wantField, wantRow, wantCol, got, rx, ry, rh, y, lines[wantRow])
	}

	// Sanity: the anchored column must differ from the OLD centered column
	// for every size this file tests (never coincidentally equal) -- a
	// regression that silently fell back to centered would otherwise still
	// pass the row/col check above only by coincidence.
	centerX := (m.width - fgW) / 2
	if centerX < 0 {
		centerX = 0
	}
	if wantCol == centerX {
		t.Fatalf("m.View() (field=%s): anchored x (%d) coincides with the OLD centered x -- test fixture cannot distinguish the two, widen it", wantField, wantCol)
	}
}

func maxLineWidth(lines []string) int {
	w := 0
	for _, l := range lines {
		if lw := ansi.StringWidth(l); lw > w {
			w = lw
		}
	}
	return w
}

// TestValueMenuAnchoredViaKeyboard guards the keyboard trigger (`s`,
// update.go's keyNodeAction) at 80 AND 120 columns, tall enough (h=40) that
// the popup fits fully below Status -- the default, no-flip case.
func TestValueMenuAnchoredViaKeyboard(t *testing.T) {
	for _, width := range []int{80, 120} {
		t.Run(width2str(width), func(t *testing.T) {
			m := anchorModel(t, width, 40)
			m = step(t, m, runeMsg('s'))
			assertValueMenuAnchoredBelowField(t, m, "status")
		})
	}
}

// TestValueMenuAnchoredViaFieldCursorEnter guards the field-cursor Enter
// trigger (boxFormActivateCursor -> activateBoxFormTarget, box_nav_field.go/
// mouse.go): cursor parked on "type", enter opens Type's menu anchored to
// Type's own box, not Status'.
func TestValueMenuAnchoredViaFieldCursorEnter(t *testing.T) {
	for _, width := range []int{80, 120} {
		t.Run(width2str(width), func(t *testing.T) {
			m := anchorModel(t, width, 40)
			b := m.focusedBean()
			m.boxFormCursor = fieldIdx(t, "type")
			m.boxFormCursorBean = b.ID
			m = step(t, m, keyMsg(tea.KeyEnter))
			assertValueMenuAnchoredBelowField(t, m, "type")
		})
	}
}

// TestValueMenuAnchoredViaMouseClick guards the mouse-click trigger
// (mouseBoxFormDetailClick -> detailBoxFormClickRow -> activateBoxFormTarget,
// mouse.go): clicking Priority's own "(u)" hotkey badge anchors to Priority.
func TestValueMenuAnchoredViaMouseClick(t *testing.T) {
	for _, width := range []int{80, 120} {
		t.Run(width2str(width), func(t *testing.T) {
			m := anchorModel(t, width, 40)
			msg := boxFormClickAt(t, m, "(u)")
			tm, _ := m.handleMouse(msg)
			m2, ok := tm.(model)
			if !ok {
				t.Fatalf("handleMouse(click) did not return a model, got %T", tm)
			}
			assertValueMenuAnchoredBelowField(t, m2, "priority")
		})
	}
}

// TestValueMenuAnchoredViaPalette guards the Command-Center trigger
// (dispatchPalette, overlay_palette.go): actionID "type" opens the SAME
// m.openValueMenu("type") the `o` key does, anchored to Type.
func TestValueMenuAnchoredViaPalette(t *testing.T) {
	for _, width := range []int{80, 120} {
		t.Run(width2str(width), func(t *testing.T) {
			m := anchorModel(t, width, 40)
			tm, _ := m.dispatchPalette(paletteItem{kind: paletteKindAction, actionID: "type", label: "set type"})
			m2, ok := tm.(model)
			if !ok {
				t.Fatalf("dispatchPalette did not return a model, got %T", tm)
			}
			assertValueMenuAnchoredBelowField(t, m2, "type")
		})
	}
}

// TestValueMenuFlipsAboveFieldWhenBelowOverflows is the render-grounded
// flip-up proof (PO-Vorgabe, Slice C step 3): a short terminal (h=20) leaves
// no room for the popup below Status. assertValueMenuAnchoredBelowField's
// own bottom-corner check already re-derives the position via
// valueMenuPopupY (so it inherently covers whichever branch fires); this
// test additionally asserts the FLIP branch is the one that actually fired
// (popup Y is strictly above the field's own top, not below it).
func TestValueMenuFlipsAboveFieldWhenBelowOverflows(t *testing.T) {
	for _, width := range []int{80, 120} {
		t.Run(width2str(width), func(t *testing.T) {
			m := anchorModel(t, width, 20)
			m = step(t, m, runeMsg('s'))
			assertValueMenuAnchoredBelowField(t, m, "status")

			b := m.idx.ByID[m.mutTarget]
			rx, ry, _, rh, ok := boxFormFieldRect(m, b, fieldIdx(t, "status"))
			_ = rx
			if !ok {
				t.Fatalf("boxFormFieldRect ok=false")
			}
			fg := m.valueMenuBox()
			fgH := len(strings.Split(fg, "\n"))
			gotY := valueMenuPopupY(ry, rh, fgH, m.height)
			below := ry + rh
			if gotY >= below {
				t.Fatalf("valueMenuPopupY = %d, want < %d (below-position) -- flip did not trigger for h=%d", gotY, below, m.height)
			}
			if gotY >= ry {
				t.Fatalf("valueMenuPopupY = %d, want < field's own top (%d) -- popup is not actually above the field", gotY, ry)
			}
		})
	}
}

// TestValueMenuStaysCenteredWithoutBoxFormFlag guards the OFF-by-default
// contract: accordion mode (BT_BOXFORM unset) has no boxed fields to anchor
// to at all (D09's own scope, "nur die Einzelfeld-Menüs"), so the value menu
// must stay centered exactly as before -- its title lands at the SAME
// (centerX+2, centerY+1) position placeOverlay's own (bgW-fgW)/2,
// (bgH-fgH)/2 formula has always produced (pre-existing, unrelated to this
// slice).
func TestValueMenuStaysCenteredWithoutBoxFormFlag(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m = step(t, m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m = focusBean(m, "tk-2")

	if boxFormEnabled() {
		t.Fatal("setup: BT_BOXFORM must be unset for this test")
	}
	m = step(t, m, runeMsg('s'))
	if m.overlay != overlayValueMenu {
		t.Fatalf("overlay = %v, want overlayValueMenu", m.overlay)
	}

	fg := m.valueMenuBox()
	fgLines := strings.Split(fg, "\n")
	fgW, fgH := maxLineWidth(fgLines), len(fgLines)
	centerX := (m.width - fgW) / 2
	centerY := (m.height - fgH) / 2
	wantRow, wantCol := centerY+1, centerX+2

	title := valueMenuTitle(m.menuItems)
	lines := screenLines(m)
	if wantRow < 0 || wantRow >= len(lines) {
		t.Fatalf("setup: wantRow=%d out of range (%d lines)", wantRow, len(lines))
	}
	idx := strings.Index(lines[wantRow], title)
	if idx < 0 {
		t.Fatalf("title %q not found on row %d: %q", title, wantRow, lines[wantRow])
	}
	if gotCol := ansi.StringWidth(lines[wantRow][:idx]); gotCol != wantCol {
		t.Fatalf("title %q at col=%d, want %d (centered, BT_BOXFORM off)", title, gotCol, wantCol)
	}
}

func width2str(w int) string {
	switch w {
	case 80:
		return "w=80"
	case 120:
		return "w=120"
	default:
		return "w=other"
	}
}
