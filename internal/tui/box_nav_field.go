package tui

// box_nav_field.go — the box-form FIELD CURSOR (bean bt-1o4g, epic bt-vy1q,
// PO-Nebenbefund N8): keyboard-first navigation between detailBoxForm's boxes
// (box_detail_form.go) while the Detail pane is focused (tab) and
// boxFormEnabled().
//
// THREE deliberate single-source decisions, all Golden-Rule-Drift-Schutz:
//
//  1. boxFormFieldOrder is the ONE field table. It carries each field's
//     (row, col) POSITION in detailBoxForm's own fixed layout (D12: Title /
//     Status|Type|Priority / Parent|Tags / Body / Relations / History), so the
//     render (which box gets focused=true), the navigation (which field is
//     left/right/up/down of which) and the mouse hit-test (mouse.go's
//     detailBoxFormClickRow, which now resolves a click through
//     boxFormFieldAt below) all read the SAME positions. There is no second,
//     independently-maintained order that could drift out of sync with the
//     rendered layout.
//
//  2. enter goes through activateBoxFormTarget (mouse.go) -- the identical
//     dispatch a mouse click already takes, which itself mirrors keyNodeAction's
//     s/o/u/a/t/e branches. One vocabulary (boxFormTarget*), one dispatch.
//
//  3. Feld-Fokus und Scroll (bean bt-ze10's viewport) are ONE movement, not
//     two competing ones -- see boxFormNav's own doc comment for the rule.

import (
	keybind "github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/xRiErOS/beans-tui/internal/data"
)

// boxFormFieldNext/boxFormFieldPrev are the Detail REGION's own tab/shift+tab
// (bean bt-8d35, PO's Fokus-Modell: "tab/shift+tab bewegen INNERHALB der
// fokussierten Region, esc VERLAESST die Region"). They carry the same
// physical keys as the global keys.FocusIn/keys.FocusOut pane swap, but a
// different, region-local meaning -- exactly the filterMenuCategoryHint
// precedent (footer_context.go), and safe by construction for the same
// reason: handleKey routes tab/shift+tab here FIRST while the Detail region
// has focus and boxFormEnabled(), so the global meaning never also runs.
//
// Deliberately standalone keybind.Bindings, NOT keyMap fields: keymap_test.go's
// TestHelpGroupsCoverEveryBindingExactlyOnce reflects over keyMap fields only
// and would flag these as unrouted Help entries. The label lives on the SAME
// value the handler matches (footer_context.go's boxFormRegionLabels swaps
// them into Footer Zone 3), so bt-z4w7's "the label IS the binding" rule
// holds literally here.
var (
	boxFormFieldNext = keybind.NewBinding(keybind.WithKeys("tab"), keybind.WithHelp("tab", "next field"))
	boxFormFieldPrev = keybind.NewBinding(keybind.WithKeys("shift+tab"), keybind.WithHelp("shift+tab", "prev field"))
)

// boxFormRow* index detailBoxForm's block slice (boxFormBlocks,
// box_detail_form.go) -- one entry per rendered layout row.
const (
	boxFormRowTitle = iota
	boxFormRowScalarsA
	boxFormRowScalarsB
	boxFormRowBody
	boxFormRowRelations
	boxFormRowHistory
)

// boxFormField is one navigable box: its position in detailBoxForm's fixed
// layout plus the boxFormTarget* action enter/click fires on it.
type boxFormField struct {
	name   string
	target string
	row    int
	col    int
}

// boxFormFieldOrder is the fixed navigation order, in detailBoxForm's own
// render order (bean bt-1o4g Ziel 2). Title/Body/Relations/History all carry
// boxFormTargetEditor: `e` opens the whole-bean $EDITOR regardless of section
// context (D01, update.go's keyNodeAction doc comment "egal an welcher
// Stelle"), the same collapse detailBoxFormClickRow already made for clicks.
var boxFormFieldOrder = []boxFormField{
	{"title", boxFormTargetEditor, boxFormRowTitle, 0},
	{"status", boxFormTargetStatus, boxFormRowScalarsA, 0},
	{"type", boxFormTargetType, boxFormRowScalarsA, 1},
	{"priority", boxFormTargetPriority, boxFormRowScalarsA, 2},
	{"parent", boxFormTargetParent, boxFormRowScalarsB, 0},
	{"tags", boxFormTargetTags, boxFormRowScalarsB, 1},
	{"body", boxFormTargetEditor, boxFormRowBody, 0},
	{"relations", boxFormTargetEditor, boxFormRowRelations, 0},
	{"history", boxFormTargetEditor, boxFormRowHistory, 0},
}

// boxFormRowCols reports how many columns a layout row has -- derived from
// boxFormFieldOrder itself rather than restating D12's 3|2 grouping, so the
// grid can never disagree with the table above.
func boxFormRowCols(row int) int {
	n := 0
	for _, f := range boxFormFieldOrder {
		if f.row == row {
			n++
		}
	}
	return n
}

// boxFormFieldIndex resolves a (row, col) position back to its index in
// boxFormFieldOrder, or -1 when the position carries no field.
func boxFormFieldIndex(row, col int) int {
	for i, f := range boxFormFieldOrder {
		if f.row == row && f.col == col {
			return i
		}
	}
	return -1
}

// boxFormFieldAt maps a Detail-pane content row/column (relative to the
// pane's own top-left, i.e. AFTER mouse.go's origin/filter-bar corrections)
// to a boxFormFieldOrder index, or -1 for "no field here" (the 1-col gap
// between grid boxes). accW is the pane's box width, the same value
// detailBoxForm renders at -- gridColWidths (box_detail_form.go) is the ONE
// place column boundaries are computed, shared with gridRow's real render.
//
// The three scalar rows are exactly 3 lines tall each (dropdownBox's fixed
// top/value/bottom shape), so rows 0-8 map arithmetically. Everything past
// them (Body/Relations/History panels, unbounded height) collapses onto the
// Body field: its target is boxFormTargetEditor, the SAME action all three
// panels fire, so this is not a mis-mapping -- it is the pre-existing
// detailBoxFormClickRow default branch, expressed through the field table.
func boxFormFieldAt(row, col, accW int) int {
	switch {
	case row < 0:
		return -1
	case row < 3:
		return boxFormFieldIndex(boxFormRowTitle, 0)
	case row < 6:
		c := gridColAt(gridColWidths(boxFormRowCols(boxFormRowScalarsA), accW), col)
		if c < 0 {
			return -1
		}
		return boxFormFieldIndex(boxFormRowScalarsA, c)
	case row < 9:
		c := gridColAt(gridColWidths(boxFormRowCols(boxFormRowScalarsB), accW), col)
		if c < 0 {
			return -1
		}
		return boxFormFieldIndex(boxFormRowScalarsB, c)
	default:
		return boxFormFieldIndex(boxFormRowBody, 0)
	}
}

// boxFormMoveField steps the cursor one field in dir ("up"/"down"/"left"/
// "right"), returning the new index. left/right walk the columns of the
// CURRENT row; up/down walk rows keeping the column where the target row has
// one and clamping to its last column otherwise (so `down` from Priority --
// row A col 2 -- lands on Tags, row B's last column). Edges are no-ops
// (no wrap-around): the PO's mental model is a form, not a carousel.
func boxFormMoveField(cur int, dir string) int {
	if cur < 0 || cur >= len(boxFormFieldOrder) {
		cur = 0
	}
	f := boxFormFieldOrder[cur]
	switch dir {
	case "left":
		if next := boxFormFieldIndex(f.row, f.col-1); next >= 0 {
			return next
		}
	case "right":
		if next := boxFormFieldIndex(f.row, f.col+1); next >= 0 {
			return next
		}
	case "up", "down":
		row := f.row - 1
		if dir == "down" {
			row = f.row + 1
		}
		if row < boxFormRowTitle || row > boxFormRowHistory {
			return cur
		}
		col := f.col
		if n := boxFormRowCols(row); col >= n {
			col = n - 1
		}
		if next := boxFormFieldIndex(row, col); next >= 0 {
			return next
		}
	}
	return cur
}

// boxFormEffectiveCursor returns the field cursor for b, defaulting to the
// FIRST field (Title) whenever the selection has moved to a different bean
// since the cursor was last written (bean bt-1o4g "Cursor resettet bei
// Bean-Wechsel"). Same derived-reset doctrine as boxFormEffectiveScroll
// (mouse.go): one lazy check where the value is READ beats N reset call sites
// at every cursor-move/click/jump that can change the selection. Also the
// defensive clamp point -- boxFormCursor survives watch-reloads untouched.
func boxFormEffectiveCursor(m model, b *data.Bean) int {
	if b == nil || b.ID != m.boxFormCursorBean {
		return 0
	}
	if m.boxFormCursor < 0 || m.boxFormCursor >= len(boxFormFieldOrder) {
		return 0
	}
	return m.boxFormCursor
}

// boxFormScrollIntoView returns the smallest adjustment of off that makes the
// field spanning [start, start+h) visible in a viewport of `height` lines.
// A field TALLER than the viewport can never be fully shown: any offset that
// still lands inside the field is kept (that is what lets `down` walk through
// a long Body line by line, see boxFormNav), only offsets outside it snap
// back to the field's own head/tail.
func boxFormScrollIntoView(off, start, h, height int) int {
	if height <= 0 {
		return off
	}
	if h >= height {
		if off < start {
			return start
		}
		if tail := start + h - height; off > tail {
			if tail < start {
				tail = start
			}
			return tail
		}
		return off
	}
	if start < off {
		return start
	}
	if start+h > off+height {
		return start + h - height
	}
	return off
}

// boxFormNav is the ONE keyboard mutation point for the box-form field cursor
// and, with it, the box-form scroll offset (bean bt-ze10's viewport).
//
// THE RULE (bean bt-1o4g's explicit "a focused field that scrolls out of the
// visible area would be a bug"): field focus LEADS, the viewport FOLLOWS.
// Whatever moves the cursor, the tail of this function pulls the viewport to
// the new field (scroll-into-view) before returning.
//
// Directions:
//
//   - "next"/"prev" (bean bt-8d35, the ONLY ones a key drives today:
//     tab/shift+tab) step LINEARLY through boxFormFieldOrder and WRAP at both
//     ends -- the PO's explicit "am Ende umbrechen, NICHT in den Tree
//     zurueckfallen, sonst verliert man den Fokus versehentlich beim
//     Durchsteppen".
//   - left/right move the cursor within its grid row.
//   - up/down first REVEAL, then MOVE: while the cursored field still extends
//     past the visible window in the pressed direction, the keypress scrolls
//     the viewport by one line and leaves the cursor where it is; once the
//     field's edge in that direction is on screen, the next press moves the
//     cursor to the neighbouring field.
//
// bean bt-8d35 (PO-Entscheidung 3) took the ARROW KEYS off the four
// directional cases and gave them back to plain viewport scrolling
// (adjustBoxFormScroll, bt-ze10's state) -- bt-ze10's "walk a long Body to its
// last line with the keyboard alone" criterion is served by that scroll again,
// so the reveal-then-move half is no longer on a key. The four directional
// cases are kept rather than torn out ("kein Voll-Rueckbau", bt-8d35): they
// are the grid geometry that boxFormFieldOrder's (row, col) table encodes,
// still covered by box_nav_field_test.go one rung below the keymap.
//
// The wheel path (wheelMove, mouse.go) is deliberately NOT routed through
// here: a wheel is a viewport gesture with no focus semantics.
//
// Writes both cursor and scroll ownership (boxFormCursorBean/
// boxFormScrollBean) so the derived per-bean resets stay consistent.
func (m model) boxFormNav(b *data.Bean, dir string) model {
	if b == nil {
		return m
	}
	accW, height := boxFormPaneMetrics(m, b)
	blocks := boxFormBlocks(m.idx, b, accW, -1) // cursor never changes line counts
	total := 0
	for _, blk := range blocks {
		total += lineCount(blk)
	}

	cur := boxFormEffectiveCursor(m, b)
	off := boxFormEffectiveScroll(m, b)

	switch dir {
	case "next", "prev":
		d := 1
		if dir == "prev" {
			d = -1
		}
		n := len(boxFormFieldOrder)
		cur = ((cur+d)%n + n) % n // wrap at BOTH ends (bt-8d35)
	case "left", "right":
		cur = boxFormMoveField(cur, dir)
	case "up", "down":
		start, h := boxFormFieldSpan(blocks, cur)
		switch {
		case dir == "down" && start+h > off+height:
			off++ // reveal the rest of the focused field first
		case dir == "up" && start < off:
			off--
		default:
			cur = boxFormMoveField(cur, dir)
		}
	default:
		return m
	}

	start, h := boxFormFieldSpan(blocks, cur)
	m.boxFormScroll = clampBoxFormScroll(boxFormScrollIntoView(off, start, h, height), total, height)
	m.boxFormScrollBean = b.ID
	m.boxFormCursor = cur
	m.boxFormCursorBean = b.ID
	return m
}

// boxFormActivateCursor fires the cursored field's own editor -- enter's
// half of Ziel 4, dispatched through the SAME activateBoxFormTarget
// (mouse.go) a click on that box takes, so keyboard and mouse can never open
// different things for the same field.
func (m model) boxFormActivateCursor(b *data.Bean) (tea.Model, tea.Cmd) {
	if b == nil {
		return m, nil
	}
	cur := boxFormEffectiveCursor(m, b)
	return m.activateBoxFormTarget(b, boxFormFieldOrder[cur].target)
}
