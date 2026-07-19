package tui

// box_detail_form.go — detailBoxForm renders a bean's SCALAR fields as
// jira-style stacked/gridded titled boxes (docs/plans/jira-style-experiment/
// design-spec.md D01/D03, mockup §6.3), reusing S1's dropdownBox primitive
// (box_dropdown.go). ADDITIVE ONLY (S2): a new pure renderer, not wired into
// the live accordion view (view_detail_bean.go's metaSectionBody/accordion
// stay untouched) -- Body/Relations/History (multi-line panels) are a later
// slice, out of scope here.

import (
	"strings"

	"github.com/xRiErOS/beans-tui/internal/data"
	"github.com/xRiErOS/beans-tui/internal/theme"
	"github.com/charmbracelet/lipgloss"
)

// detailBoxFormGap is the single space column gap between adjacent boxes in
// a row (design-spec.md §6.3 mockup: boxes separated by one blank column).
const detailBoxFormGap = 1

// detailBoxFormPerRow picks the responsive column count for the scalar-field
// grid from the available inner pane width: 3-up on wide panes, 2-up on
// medium, 1-up (stacked) on narrow -- mirrors the mockup's own breakpoints
// (design-spec.md §6.3).
func detailBoxFormPerRow(width int) int {
	switch {
	case width >= 96:
		return 3
	case width >= 64:
		return 2
	default:
		return 1
	}
}

// detailBoxForm renders a bean's scalar fields as jira-style titled boxes
// (design-spec.md D01/D03, mockup §6.3): a full-width Title box, then a
// responsive grid of Status/Type/Priority/Parent/Tags boxes, then full-width
// Body/Relations/History panels (S2c, design-spec.md D01/D04) built via
// panelBox (box_panel.go) reusing view_detail_bean.go's own content
// renderers. idx is needed to resolve Relations (parent/children/blocking).
// width = inner pane width in cells.
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

	type scalarField struct {
		label, value, hotkey string
	}
	fields := []scalarField{
		{"Status", theme.StatusStyle(b.Status).Render(b.Status), "s"},
		{"Type", theme.TypeStyle(b.Type).Render(b.Type), "o"},
		{"Priority", theme.Priority(priority), "u"},
		{"Parent", parent, "a"},
		{"Tags", tags, "t"},
	}

	perRow := detailBoxFormPerRow(width)
	colW := (width - (perRow-1)*detailBoxFormGap) / perRow

	lines := []string{title}
	for i := 0; i < len(fields); i += perRow {
		end := i + perRow
		if end > len(fields) {
			end = len(fields)
		}
		row := fields[i:end]
		boxes := make([]string, len(row))
		for j, f := range row {
			boxes[j] = dropdownBox(f.label, f.value, f.hotkey, colW, false)
		}
		joined := boxes[0]
		for _, box := range boxes[1:] {
			joined = lipgloss.JoinHorizontal(lipgloss.Top, joined, strings.Repeat(" ", detailBoxFormGap), box)
		}
		lines = append(lines, joined)
	}

	relationsBody, _, _ := relationsSectionBody(idx, b, width-4, false, 0)
	lines = append(lines,
		panelBox("Body", bodySectionBody(b, width-4), "e", width, false),
		panelBox("Relations", relationsBody, "", width, false),
		panelBox("History", historieSectionBody(b, width-4), "", width, false),
	)

	return strings.Join(lines, "\n")
}
