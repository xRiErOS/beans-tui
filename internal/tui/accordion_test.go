package tui

// accordion_test.go — TDD-first tests for the Detail-Accordion port (E2 Task
// 1, bean bt-ms0k): renderAccordion's exclusive-open/digit-numbering algebra
// (port of devd accordion.go:309-373, minus the edit-field concept) and
// glowRender (port of devd editor.go:102-126).

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
)

// TestRenderAccordionExclusiveOpen guards the exclusive-open contract: with 3
// sections and open=2, only section 2's body renders -- sections 1/3 show
// only their header (+ chevron), never their body.
func TestRenderAccordionExclusiveOpen(t *testing.T) {
	secs := []accordionSection{
		{title: "One", body: "body-one-content"},
		{title: "Two", body: "body-two-content"},
		{title: "Three", body: "body-three-content"},
	}
	out := renderAccordion(secs, 2, 60, false, 0, 0)

	if strings.Contains(out, "body-one-content") {
		t.Error("closed section 1's body must not render")
	}
	if !strings.Contains(out, "body-two-content") {
		t.Error("open section 2's body must render")
	}
	if strings.Contains(out, "body-three-content") {
		t.Error("closed section 3's body must not render")
	}
	if !strings.Contains(ansi.Strip(out), "▸") {
		t.Error("closed sections must show the closed chevron ▸")
	}
	if !strings.Contains(ansi.Strip(out), "▾") {
		t.Error("open section must show the open chevron ▾")
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
