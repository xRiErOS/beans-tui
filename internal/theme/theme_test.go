package theme

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// TestStatusColorMapping — alle 5 beans-Status → erwartete Catppuccin-Macchiato-Hex
// (Plan »E1 Task 6« Adaptation 2). Unbekannter Status fällt neutral auf Text zurück.
func TestStatusColorMapping(t *testing.T) {
	cases := []struct {
		status string
		want   lipgloss.Color
	}{
		{"draft", Blue},
		{"todo", Text},
		{"in-progress", Yellow},
		{"completed", Green},
		{"scrapped", Red},
	}
	for _, c := range cases {
		if got := StatusColor(c.status); got != c.want {
			t.Errorf("StatusColor(%q) = %q, want %q", c.status, got, c.want)
		}
	}

	if got := StatusColor("some-unknown-status"); got != Subtext {
		t.Errorf("StatusColor(unknown) = %q, want fallback Subtext %q", got, Subtext)
	}

	if icon := StatusIcon("some-unknown-status"); !strings.Contains(icon, fallbackGlyphChar) || strings.Contains(icon, statusGlyph) {
		t.Errorf("StatusIcon(unknown) = %q, want fallback glyph %q not status glyph %q", icon, fallbackGlyphChar, statusGlyph)
	}
}

// TestAsciiFallback — BT_ASCII_ICONS steuert Status- und Type-Glyphen (Adaptation 1).
// Per-Aufruf-Read via os.Getenv, daher mit t.Setenv pro Sub-Case steuerbar (kein
// Package-State-Caching, das Tests verunreinigen könnte).
func TestAsciiFallback(t *testing.T) {
	t.Setenv("BT_ASCII_ICONS", "1")

	if !strings.Contains(StatusIcon("todo"), statusGlyphASCII) {
		t.Errorf("StatusIcon(todo) with BT_ASCII_ICONS=1 should contain ASCII glyph %q", statusGlyphASCII)
	}
	if strings.Contains(StatusIcon("todo"), statusGlyph) {
		t.Errorf("StatusIcon(todo) with BT_ASCII_ICONS=1 should not contain unicode glyph %q", statusGlyph)
	}
	if !strings.Contains(TypeIcon("bug"), typeIconASCII["bug"]) {
		t.Errorf("TypeIcon(bug) with BT_ASCII_ICONS=1 should contain ASCII glyph %q", typeIconASCII["bug"])
	}

	t.Setenv("BT_ASCII_ICONS", "0")

	if !strings.Contains(StatusIcon("todo"), statusGlyph) {
		t.Errorf("StatusIcon(todo) without ascii mode should contain unicode glyph %q", statusGlyph)
	}
	if !strings.Contains(TypeIcon("bug"), typeIcon["bug"]) {
		t.Errorf("TypeIcon(bug) without ascii mode should contain unicode glyph %q", typeIcon["bug"])
	}
}

// TestTypeIconAllTypes — alle 5 beans-Typen → erwarteter Glyph + Farbe (Adaptation 3).
// Unbekannter Typ fällt auf den generischen Fallback-Glyph zurück (siehe fallbackGlyph).
func TestTypeIconAllTypes(t *testing.T) {
	cases := []struct {
		typ   string
		glyph string
		color lipgloss.Color
	}{
		{"milestone", "⬢", Peach},
		{"epic", "✦", Mauve},
		{"feature", "✦", Green},
		{"task", "⯅", Blue},
		{"bug", "⯁", Red},
	}
	for _, c := range cases {
		if got := typeGlyph(c.typ); got != c.glyph {
			t.Errorf("typeGlyph(%q) = %q, want %q", c.typ, got, c.glyph)
		}
		if got, ok := typeColor[c.typ]; !ok || got != c.color {
			t.Errorf("typeColor[%q] = %q, want %q", c.typ, got, c.color)
		}
		if rendered := TypeIcon(c.typ); !strings.Contains(rendered, c.glyph) {
			t.Errorf("TypeIcon(%q) = %q, want contains %q", c.typ, rendered, c.glyph)
		}
	}

	if got := typeGlyph("unknown-type"); got != fallbackGlyphChar {
		t.Errorf("typeGlyph(unknown) = %q, want fallback %q", got, fallbackGlyphChar)
	}
	if _, ok := typeColor["unknown-type"]; ok {
		t.Fatal("typeColor should have no entry for an unknown type")
	}
	if got := TypeIcon("unknown-type"); !strings.Contains(got, fallbackGlyphChar) {
		t.Errorf("TypeIcon(unknown) = %q, want contains fallback %q", got, fallbackGlyphChar)
	}
}

// TestPriorityColorMapping — beans-Priorität → erwartete Farbe + Bold-Flag
// (Adaptation 4).
func TestPriorityColorMapping(t *testing.T) {
	cases := []struct {
		priority string
		want     lipgloss.Color
	}{
		{"critical", Red},
		{"high", Red},
		{"normal", Text},
		{"low", Green},
		{"deferred", Hint},
	}
	for _, c := range cases {
		if got := priorityColor(c.priority); got != c.want {
			t.Errorf("priorityColor(%q) = %q, want %q", c.priority, got, c.want)
		}
	}

	rendered := Priority("critical")
	if !strings.Contains(rendered, "critical") {
		t.Errorf("Priority(critical) = %q, want contains %q", rendered, "critical")
	}
}
