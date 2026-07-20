package tui

// box_detail_form.go — detailBoxForm renders a bean's SCALAR fields as
// jira-style stacked/gridded titled boxes (docs/plans/jira-style-experiment/
// design-spec.md D01/D03, mockup §6.3), reusing S1's dropdownBox primitive
// (box_dropdown.go). ADDITIVE ONLY (S2): a new pure renderer, not wired into
// the live accordion view (view_detail_bean.go's metaSectionBody/accordion
// stay untouched) -- Body/Relations/History (multi-line panels) are a later
// slice, out of scope here.
//
// S2e (D12, review B1): the responsive perRow collapse (3-up/2-up/1-up) is
// gone. The scalar grid is now FIXED: Row A = Status|Type|Priority (3 cols),
// Row B = Parent|Tags (2 cols) -- always, at any width. Columns shrink with
// the pane (dropdownBox floors at 8 cells) but never collapse to 1-up.

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/xRiErOS/beans-tui/internal/data"
	"github.com/xRiErOS/beans-tui/internal/theme"
)

// detailBoxFormGap is the single space column gap between adjacent boxes in
// a row (design-spec.md §6.3 mockup: boxes separated by one blank column).
const detailBoxFormGap = 1

// scalarCell is one column of a gridRow: a dropdownBox's label/value/hotkey.
type scalarCell struct {
	label, value, hotkey string
}

// gridRow renders the given scalar cells as one horizontal row occupying
// EXACTLY `width` cells (S2e Change 3, review B1): len(cells) columns, one-
// space gaps between them, and the integer-division remainder spread across
// the first columns so the widths sum to width. Each cell is a dropdownBox
// at its column width.
func gridRow(cells []scalarCell, width int) string {
	n := len(cells)
	if n == 0 {
		return ""
	}
	gap := detailBoxFormGap
	avail := width - (n-1)*gap
	base := avail / n
	rem := avail % n

	boxes := make([]string, n)
	for i, c := range cells {
		colW := base
		if i < rem {
			colW++
		}
		boxes[i] = dropdownBox(c.label, c.value, c.hotkey, colW, false)
	}

	gapCol := " \n \n "
	joined := boxes[0]
	for _, box := range boxes[1:] {
		joined = lipgloss.JoinHorizontal(lipgloss.Top, joined, gapCol, box)
	}
	return joined
}

// detailBoxForm renders a bean's scalar fields as jira-style titled boxes
// (design-spec.md D01/D03, mockup §6.3; D12: fixed grid, no responsive
// collapse): a full-width Title box, then a FIXED Status|Type|Priority row
// (3 cols) and a FIXED Parent|Tags row (2 cols), then full-width Body/
// Relations/History panels (S2c, design-spec.md D01/D04) built via panelBox
// (box_panel.go) reusing view_detail_bean.go's own content renderers. idx is
// needed to resolve Relations (parent/children/blocking). width = inner pane
// width in cells.
func detailBoxForm(idx *data.Index, b *data.Bean, width int) string {
	title := dropdownBox("Title", b.Title, "e", width, false)

	priority := b.Priority
	if priority == "" {
		priority = "normal"
	}
	parent := b.Parent
	if parent == "" {
		parent = theme.Dim.Render("—")
	}
	tags := tagsInline(b.Tags)
	if tags == "" {
		tags = theme.Dim.Render("—")
	}

	rowA := gridRow([]scalarCell{
		{"Status", theme.StatusStyle(b.Status).Render(b.Status), "s"},
		{"Type", theme.TypeStyle(b.Type).Render(b.Type), "o"},
		{"Priority", theme.Priority(priority), "u"},
	}, width)
	rowB := gridRow([]scalarCell{
		{"Parent", parent, "a"},
		{"Tags", tags, "t"},
	}, width)

	relationsBody, _, _ := relationsSectionBody(idx, b, width-4, false, 0)

	lines := []string{
		title,
		rowA,
		rowB,
		panelBox("Body", bodySectionBody(b, width-4), "e", width, false),
		panelBox("Relations", relationsBody, "", width, false),
		panelBox("History", historieSectionBody(b, width-4), "", width, false),
	}

	return strings.Join(lines, "\n")
}
