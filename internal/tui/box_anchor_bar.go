package tui

// box_anchor_bar.go — the Body anchor bar (bean bt-p78f #16, epic bt-vy1q): a
// single-line chip strip of the Body's Markdown headings, parked at the top of
// the Body panel. Clicking a chip jumps the viewport to that heading. Two pure
// helpers here; the render/click wiring lives at the detailBoxForm/mouse call
// sites.
//
//   renderedHeadingLines maps each parsed section (raw-Markdown offsets) to its
//   line in the GLAMOUR-rendered body -- the scroll offset indexes rendered
//   lines, the parser indexes raw ones (PO-Entscheidung D03: locate by title
//   match, no body-render rewrite). Robust across glamour's notty style
//   ("  ## Ziel  ", hashes kept) and its dark style ("Ziel", hashes dropped):
//   strip a leading run of '#'/space, then compare to the title.
//
//   anchorBar builds the chip row + per-chip click spans, truncating with an
//   ellipsis when the headings overflow the width (PO-Entscheidung D04) -- the
//   dropped chips stay reachable via the body text itself, the bar is only the
//   quick jump.

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/xRiErOS/beans-tui/internal/data"
	"github.com/xRiErOS/beans-tui/internal/theme"
)

// anchorChip is one rendered chip's click span: [start,end) are RUNE columns in
// the ANSI-stripped anchor row, sec is the index into the section list it jumps
// to.
type anchorChip struct {
	sec        int
	start, end int
}

// headingText strips a rendered line to its bare heading text: drop ANSI, trim,
// then drop a leading run of '#' and spaces (glamour notty keeps the hashes,
// dark drops them). A non-heading prose line reduces to its own trimmed text,
// which will simply not equal any title.
func headingText(renderedLine string) string {
	s := strings.TrimSpace(ansi.Strip(renderedLine))
	s = strings.TrimLeft(s, "#")
	return strings.TrimSpace(s)
}

// renderedHeadingLines returns, for each section (in order), the 0-based line
// index within rendered where its heading appears, or -1 if it cannot be
// located. Scans with a forward-only cursor so repeated titles map in document
// order and a later heading never matches a line above an earlier one.
func renderedHeadingLines(rendered string, secs []bodySection) []int {
	lines := strings.Split(rendered, "\n")
	out := make([]int, len(secs))
	cursor := 0
	for i, s := range secs {
		out[i] = -1
		for j := cursor; j < len(lines); j++ {
			if headingText(lines[j]) == s.title {
				out[i] = j
				cursor = j + 1
				break
			}
		}
	}
	return out
}

// anchorBar renders the heading chip strip to fit width visible columns and
// returns it with the click span of every chip that made it in. Chips read
// "[Title]" separated by a space; when the next chip would not fit, an ellipsis
// is appended and the rest are dropped (D04). A single title wider than the
// whole strip is truncated with an inner ellipsis. Empty (row "", no chips)
// when there are no sections.
func anchorBar(secs []bodySection, width int) (string, []anchorChip) {
	if len(secs) == 0 || width < 3 {
		return "", nil
	}
	const ellipsis = "…"
	chipStyle := lipgloss.NewStyle().Foreground(theme.Subtext)

	var b strings.Builder
	var chips []anchorChip
	col := 0 // visible rune column of the next write

	for i, s := range secs {
		sep := ""
		if i > 0 {
			sep = " "
		}
		label := "[" + s.title + "]"
		need := len([]rune(sep)) + len([]rune(label))

		// Reserve room for a trailing ellipsis unless this is the last chip.
		reserve := 0
		if i < len(secs)-1 {
			reserve = len([]rune(ellipsis)) + 1
		}
		if col+need+reserve > width {
			// This chip (and the rest) don't fit -> ellipsis, stop.
			if col+len([]rune(ellipsis)) <= width {
				b.WriteString(chipStyle.Render(ellipsis))
			}
			break
		}

		b.WriteString(sep)
		col += len([]rune(sep))
		start := col
		b.WriteString(chipStyle.Render(label))
		col += len([]rune(label))
		chips = append(chips, anchorChip{sec: i, start: start, end: col})
	}
	return b.String(), chips
}

// anchorBarHit returns the section index whose chip span contains rune column
// x, or -1 if x lands in a gap, on the ellipsis, or past the row.
func anchorBarHit(chips []anchorChip, x int) int {
	for _, c := range chips {
		if x >= c.start && x < c.end {
			return c.sec
		}
	}
	return -1
}

// boxFormBodyPanelContent builds the Body panel's inner content for the box
// form: the anchor bar as line 0 (when the Body has locatable headings) then
// the glamour-rendered body. inner is the panel's inner width (= box width-4).
// Returns the chip spans and each section's line index within the RENDERED body
// (without the prepended bar) so a click can jump to it. When there are no
// headings the content is just the body and chips/headingLines are nil -- the
// render then stays byte-identical to the pre-bt-p78f box form.
func boxFormBodyPanelContent(b *data.Bean, inner int) (content string, chips []anchorChip, headingLines []int) {
	body := bodySectionBody(b, inner)
	secs := parseBodySections(b.Body)
	bar, ch := anchorBar(secs, inner)
	if bar == "" {
		return body, nil, nil
	}
	return bar + "\n" + body, ch, renderedHeadingLines(body, secs)
}

// boxFormBodyContent is the render-site view of boxFormBodyPanelContent: just
// the Body panel's inner content string (anchor bar + body, or body alone).
func boxFormBodyContent(b *data.Bean, inner int) string {
	content, _, _ := boxFormBodyPanelContent(b, inner)
	return content
}

// boxFormAnchorRow returns the form content line of the anchor bar itself (the
// row a click must land on to hit a chip), given detailBoxForm's block slice.
// The anchor bar is line 0 of the Body panel's inner content, i.e. one line
// below the Body block's top border, which itself follows the Title/rowA/rowB
// blocks. Returns -1 if the block slice is too short.
func boxFormAnchorRow(blocks []string) int {
	if len(blocks) <= boxFormRowBody {
		return -1
	}
	start := 0
	for i := 0; i < boxFormRowBody; i++ {
		start += lineCount(blocks[i])
	}
	return start + 1 // +1 for the Body panel's top border
}

// boxFormAnchorTargetLine returns the form content line to scroll to so that
// section sec's heading sits at the top of the viewport, or -1 if it cannot be
// located. Built from the SAME block slice boxFormAnchorRow uses: the body's
// rendered content begins one line below the anchor bar, so a heading at
// headingLines[sec] within the body lands at anchorRow + 1 + that offset.
func boxFormAnchorTargetLine(blocks []string, headingLines []int, sec int) int {
	if sec < 0 || sec >= len(headingLines) || headingLines[sec] < 0 {
		return -1
	}
	anchorRow := boxFormAnchorRow(blocks)
	if anchorRow < 0 {
		return -1
	}
	return anchorRow + 1 + headingLines[sec]
}
