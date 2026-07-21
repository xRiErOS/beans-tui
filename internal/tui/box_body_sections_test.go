package tui

// box_body_sections_test.go — TDD for the ONE Markdown-heading parser both
// bt-p78f surfaces share (anchor bar #16 + section pencil #15, epic bt-vy1q).
// The parser is the real work; the two UIs are thin views on its output. Two
// hard requirements drive these tests: (1) a `#` line INSIDE a fenced code
// block is NOT a heading (a naive line parser slices code), (2) each section's
// byte span [start,end) must be exact so section-wise write-back can splice the
// edited text back into the whole Body without disturbing its neighbours.

import (
	"strings"
	"testing"
)

func TestParseBodySectionsBasic(t *testing.T) {
	body := "intro line\n## Ziel\nwas wir wollen\n## Definition of Done\nfertig wenn\n### Detail\ntief\n"
	secs := parseBodySections(body)
	if len(secs) != 3 {
		t.Fatalf("got %d sections, want 3: %+v", len(secs), secs)
	}
	want := []struct {
		title string
		level int
	}{
		{"Ziel", 2},
		{"Definition of Done", 2},
		{"Detail", 3},
	}
	for i, w := range want {
		if secs[i].title != w.title || secs[i].level != w.level {
			t.Errorf("section %d = (%q, L%d), want (%q, L%d)", i, secs[i].title, secs[i].level, w.title, w.level)
		}
	}
	// The span of each section must round-trip: body[start:end] begins with the
	// heading line and covers up to the next heading of level <= its own.
	for i, s := range secs {
		if s.start < 0 || s.end > len(body) || s.start >= s.end {
			t.Fatalf("section %d span [%d,%d) out of range for body len %d", i, s.start, s.end, len(body))
		}
		if !strings.HasPrefix(body[s.start:s.end], strings.Repeat("#", s.level)+" ") {
			t.Errorf("section %d span does not start at its heading: %q", i, body[s.start:s.end])
		}
	}
	// "Ziel" (L2) ends where "Definition of Done" (L2) begins.
	if body[secs[0].end-1] != '\n' && secs[0].end != secs[1].start {
		t.Errorf("Ziel end=%d should meet Definition start=%d", secs[0].end, secs[1].start)
	}
	if secs[0].end != secs[1].start {
		t.Errorf("Ziel.end (%d) must equal DoD.start (%d) -- L2 section ends at next L<=2 heading", secs[0].end, secs[1].start)
	}
	// "Definition of Done" (L2) INCLUDES its "### Detail" (L3) subsection, so it
	// runs to end-of-body, not to the L3 heading.
	if secs[1].end != len(body) {
		t.Errorf("DoD.end = %d, want %d (a section swallows deeper subsections)", secs[1].end, len(body))
	}
}

func TestParseBodySectionsIgnoresFencedCode(t *testing.T) {
	body := "## Real\ntext\n```sh\n# not a heading\n## also not\n```\n## After\ndone\n"
	secs := parseBodySections(body)
	if len(secs) != 2 {
		t.Fatalf("got %d sections, want 2 (the two inside the ``` fence must be ignored): %+v", len(secs), secs)
	}
	if secs[0].title != "Real" || secs[1].title != "After" {
		t.Errorf("titles = %q,%q, want Real,After", secs[0].title, secs[1].title)
	}
	// The fenced "# not a heading" lines must fall INSIDE Real's span, not split it.
	if !strings.Contains(body[secs[0].start:secs[0].end], "# not a heading") {
		t.Errorf("Real's span must contain the fenced hashes, got %q", body[secs[0].start:secs[0].end])
	}
}

func TestParseBodySectionsTildeFenceAndNoHeadings(t *testing.T) {
	if secs := parseBodySections("just prose\nno headings here\n"); len(secs) != 0 {
		t.Errorf("body with no headings -> %d sections, want 0", len(secs))
	}
	// ~~~ fences count too.
	body := "~~~\n# fenced\n~~~\n# Real\n"
	secs := parseBodySections(body)
	if len(secs) != 1 || secs[0].title != "Real" {
		t.Fatalf("tilde fence not honoured: %+v", secs)
	}
}

func TestParseBodySectionsRejectsNonHeadings(t *testing.T) {
	// "#tag" (no space) and "#### too deep is fine to L6, but 7 hashes" edge:
	// require 1..6 hashes followed by a space.
	body := "#nospace\n####### sevenhashes\n# Good\n"
	secs := parseBodySections(body)
	if len(secs) != 1 || secs[0].title != "Good" || secs[0].level != 1 {
		t.Fatalf("only '# Good' is a valid heading, got %+v", secs)
	}
}
