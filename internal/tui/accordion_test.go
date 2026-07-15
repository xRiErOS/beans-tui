package tui

// accordion_test.go — TDD-first tests for the Detail-Accordion port (E2 Task
// 1, bean bt-ms0k): renderAccordion's exclusive-open/digit-numbering algebra
// (port of devd accordion.go:309-373, minus the edit-field concept) and
// glowRender (port of devd editor.go:102-126). E7 Task 4 (bean bt-kyj5,
// PF-1/PF-12) adds: section 1 (Meta) always renders its body regardless of
// `open` (PF-1, "nicht kollabierbar"), and both renderAccordion's header
// styling AND fieldStrip reserve a stable 1-column gutter for their
// active/inactive marker (PF-12) instead of conditionally prefixing it only
// when active.

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
)

// TestRenderAccordionExclusiveOpen guards the exclusive-open contract: with 3
// sections and open=2, section 2's body renders (the open target) AND
// section 1's body ALSO renders (PF-1: Meta is never collapsible, `isOpen :=
// n == open || n == 1`) -- section 3 stays header-only, the same as before
// PF-1.
func TestRenderAccordionExclusiveOpen(t *testing.T) {
	secs := []accordionSection{
		{title: "One", body: "body-one-content"},
		{title: "Two", body: "body-two-content"},
		{title: "Three", body: "body-three-content"},
	}
	out := renderAccordion(secs, 2, 60, false, 0, 0)

	if !strings.Contains(out, "body-one-content") {
		t.Error("PF-1: section 1's body must ALWAYS render, regardless of open")
	}
	if !strings.Contains(out, "body-two-content") {
		t.Error("open section 2's body must render")
	}
	if strings.Contains(out, "body-three-content") {
		t.Error("closed section 3's body must not render")
	}
	if !strings.Contains(ansi.Strip(out), "▸") {
		t.Error("closed section 3 must show the closed chevron ▸")
	}
	if !strings.Contains(ansi.Strip(out), "▾") {
		t.Error("open sections (1 and 2) must show the open chevron ▾")
	}
}

// TestRenderAccordionSectionOneAlwaysOpenRegardlessOfOpenParam guards PF-1
// directly: section 1's body renders no matter what `open` is set to,
// including 0 (design-spec.md §15 PF-1: `accOpen == 0` is a legitimate rest
// state post-PF-1 -- "only Meta visible, no section 2-4 additionally open"),
// a value pointing at another section, and section 1 itself.
func TestRenderAccordionSectionOneAlwaysOpenRegardlessOfOpenParam(t *testing.T) {
	secs := []accordionSection{
		{title: "Meta", body: "meta-body-content"},
		{title: "Body", body: "body-body-content"},
	}
	for _, open := range []int{0, 1, 2} {
		out := renderAccordion(secs, open, 60, false, 0, 0)
		if !strings.Contains(out, "meta-body-content") {
			t.Errorf("open=%d: section 1 (Meta) body must always render (PF-1)", open)
		}
	}
}

// TestRenderAccordionHeaderGutterWidthStableAcrossActiveState guards PF-12
// for section headers. The OLD asymmetry (design-spec.md §15 PF-12 item 1):
// the active branch prefixes "▌" (1 reserved column) then truncates to w-1,
// while the inactive branch has NO prefix and truncates to the FULL w --
// so a header's own title text starts 1 column EARLIER when inactive than
// when active (verified below: fails under the old code, col 6 vs col 7 for
// a "[1] Solo" header at w=40). The fix reserves the SAME 1 column either
// way (" " inactive, "▌" active) so the title always starts at the same
// column regardless of the row's own active state.
func TestRenderAccordionHeaderGutterWidthStableAcrossActiveState(t *testing.T) {
	secs := []accordionSection{{title: "Solo", body: "solo-body"}}
	const w = 40

	activeOut := ansi.Strip(renderAccordion(secs, 1, w, true, 0, 0))    // the only section IS active
	inactiveOut := ansi.Strip(renderAccordion(secs, 1, w, false, 0, 0)) // active=false globally -- no section active

	activeCol := cellCol(t, activeOut, "Solo")
	inactiveCol := cellCol(t, inactiveOut, "Solo")
	if activeCol != inactiveCol {
		t.Errorf("PF-12: header title starts at column %d when active vs column %d when inactive -- gutter must reserve the same 1 column either way", activeCol, inactiveCol)
	}
}

// cellCol locates substr in s (both already ANSI-stripped) and returns its
// starting terminal CELL column (lipgloss.Width of the preceding text) --
// NOT strings.Index's raw byte offset, which would over-count multi-byte
// gutter glyphs like "▌" (3 UTF-8 bytes) against a 1-byte " " and produce a
// false mismatch between two otherwise column-equal renders.
func cellCol(t *testing.T, s, substr string) int {
	t.Helper()
	idx := strings.Index(s, substr)
	if idx < 0 {
		t.Fatalf("substring %q not found in %q", substr, s)
	}
	return lipgloss.Width(s[:idx])
}

// TestFieldStripGutterWidthStable guards PF-12 for fieldStrip (design-spec.md
// §15 PF-12 item 2, "alle markierbaren Zeilen" -- same root cause as the
// header fix, fixed in the same pass): the active branch prefixes "▌" (1
// column) before the label, the inactive branch has none -- so a field's own
// label starts 1 column earlier when inactive than when active. The fix adds
// a matching " " gutter to the inactive branch.
func TestFieldStripGutterWidthStable(t *testing.T) {
	fields := []relationField{{beanID: "solo", label: "solo-field"}}

	activeOut := ansi.Strip(fieldStrip(fields, 0, 60))    // the only field IS active
	inactiveOut := ansi.Strip(fieldStrip(fields, -1, 60)) // -1 matches no field -- inactive

	activeCol := cellCol(t, activeOut, "solo-field")
	inactiveCol := cellCol(t, inactiveOut, "solo-field")
	if activeCol != inactiveCol {
		t.Errorf("PF-12: field label starts at column %d when active vs column %d when inactive -- gutter must reserve the same 1 column either way", activeCol, inactiveCol)
	}
}

// TestRenderAccordionDigitHeaderNumbering guards the ziffern-Header format:
// header lines carry [1]..[3] in order (ansi-stripped, per primitives_test.go
// convention).
func TestRenderAccordionDigitHeaderNumbering(t *testing.T) {
	secs := []accordionSection{
		{title: "One", body: "b1"},
		{title: "Two", body: "b2"},
		{title: "Three", body: "b3"},
	}
	out := ansi.Strip(renderAccordion(secs, 1, 60, false, 0, 0))

	i1 := strings.Index(out, "[1]")
	i2 := strings.Index(out, "[2]")
	i3 := strings.Index(out, "[3]")
	if i1 < 0 || i2 < 0 || i3 < 0 {
		t.Fatalf("expected [1]..[3] digit headers in output, got: %q", out)
	}
	if !(i1 < i2 && i2 < i3) {
		t.Errorf("digit headers out of order: [1]@%d [2]@%d [3]@%d", i1, i2, i3)
	}
}

// TestRenderAccordionEmptySectionsShowsPlaceholder guards the nil/empty input
// path: a placeholder hint renders instead of an empty string or a panic.
func TestRenderAccordionEmptySectionsShowsPlaceholder(t *testing.T) {
	out := renderAccordion(nil, 0, 60, false, 0, 0)
	if strings.TrimSpace(out) == "" {
		t.Fatal("renderAccordion(nil, ...) must render a placeholder hint, got empty string")
	}
}

// TestGlowRenderEmptyBodyReturnsEmpty guards glowRender's empty-input
// short-circuit (port devd editor.go:102-126).
func TestGlowRenderEmptyBodyReturnsEmpty(t *testing.T) {
	if got := glowRender("", 80); got != "" {
		t.Errorf("glowRender(\"\", 80) = %q, want empty string", got)
	}
}

// TestGlowRenderAsciiProfileProducesNoEscapeSequences guards Golden-
// Determinism for future Detail goldens: under the Ascii color profile,
// glowRender's glamour output must never contain a raw ESC[ sequence (notty
// style).
func TestGlowRenderAsciiProfileProducesNoEscapeSequences(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)

	out := glowRender("# Heading\n\nSome **bold** body text.", 80)
	if strings.Contains(out, "\x1b[") {
		t.Errorf("glowRender under Ascii profile must contain no escape sequences, got: %q", out)
	}
	if !strings.Contains(out, "Heading") {
		t.Errorf("glowRender output missing rendered content: %q", out)
	}
}
