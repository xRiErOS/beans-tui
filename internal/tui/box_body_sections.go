package tui

// box_body_sections.go — parseBodySections splits a bean's raw Markdown Body
// into heading-delimited sections (bean bt-p78f, epic bt-vy1q). It is the ONE
// parser both surfaces share: the anchor bar (#16) renders one chip per
// section; the section pencil (#15) opens exactly one section's byte span in
// the editor and splices the result back. Deliberately raw-string / byte-offset
// based (not glamour's rendered output) -- the offsets must index the SAME
// string the write-back path mutates, or a section edit would corrupt the Body.
//
// Fenced code blocks (``` or ~~~) are tracked so a `#` line inside them is NOT
// mistaken for a heading (bean Fallstrick: "ein naiver Zeilen-Parser
// zerschneidet Code"). A heading is 1..6 `#` immediately followed by a space.

import "strings"

// bodySection is one heading and the byte span it governs within the Body.
// [start,end) covers from the heading line to the next heading of level <= this
// one (deeper subsections are swallowed), or to end-of-body for the last.
type bodySection struct {
	title string // heading text, trimmed (no leading #'s, no surrounding space)
	level int    // 1..6
	line  int    // 0-based line index of the heading line (for scroll-to-anchor)
	start int    // byte offset of the heading line's first byte
	end   int    // byte offset one past the section's last byte
}

// headingLevel returns the heading level (1..6) of a single line, or 0 if the
// line is not an ATX heading: 1..6 leading '#' followed by a space. Seven or
// more hashes, or a '#' with no following space ("#tag"), are not headings.
func headingLevel(line string) int {
	n := 0
	for n < len(line) && line[n] == '#' {
		n++
	}
	if n < 1 || n > 6 {
		return 0
	}
	if n >= len(line) || line[n] != ' ' {
		return 0
	}
	return n
}

// fenceMarker reports whether the line opens/closes a fenced code block, i.e.
// its first non-space run is 3+ backticks or 3+ tildes (CommonMark fences).
func fenceMarker(line string) bool {
	s := strings.TrimLeft(line, " ")
	for _, q := range []byte{'`', '~'} {
		n := 0
		for n < len(s) && s[n] == q {
			n++
		}
		if n >= 3 {
			return true
		}
	}
	return false
}

// parseBodySections walks body line by line, tracking fenced-code state, and
// returns its heading sections in document order (empty slice if none). Byte
// offsets index body directly, so body[s.start:s.end] is the exact, splice-safe
// text of section s.
func parseBodySections(body string) []bodySection {
	var secs []bodySection
	inFence := false
	offset := 0 // byte offset of the current line's first byte

	// Split keeping track of byte offsets. strings.Split drops the separators,
	// so recompute each line's length + 1 for the '\n' as we go. The final line
	// (no trailing '\n') is handled by clamping the last section's end below.
	lines := strings.Split(body, "\n")
	for i, line := range lines {
		lineStart := offset
		offset += len(line) + 1 // +1 for the '\n' Split removed

		if fenceMarker(line) {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		if lvl := headingLevel(line); lvl > 0 {
			secs = append(secs, bodySection{
				title: strings.TrimSpace(line[lvl:]),
				level: lvl,
				line:  i,
				start: lineStart,
			})
		}
	}

	// Second pass: each section ends where the next heading of level <= its own
	// begins; the tail runs to end-of-body. Clamp to len(body) (the running
	// offset overshoots by 1 past the final line when body has no trailing \n).
	for i := range secs {
		end := len(body)
		for j := i + 1; j < len(secs); j++ {
			if secs[j].level <= secs[i].level {
				end = secs[j].start
				break
			}
		}
		secs[i].end = end
	}
	return secs
}
