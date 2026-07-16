package theme

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// TestStatusColorMapping — alle 5 beans-Status → erwartete Catppuccin-Macchiato-Hex
// UND ihr Buchstaben-Icon (PF-6, design-spec.md §15, 2026-07-15). Unbekannter Status
// fällt neutral auf Subtext + den generischen Fallback-Glyph zurück.
func TestStatusColorMapping(t *testing.T) {
	cases := []struct {
		status string
		want   lipgloss.Color
		letter string
	}{
		{"draft", Blue, "d"},
		{"todo", Green, "t"},
		{"in-progress", Yellow, "i"},
		{"completed", Subtext, "c"},
		{"scrapped", Subtext, "s"},
	}
	for _, c := range cases {
		if got := StatusColor(c.status); got != c.want {
			t.Errorf("StatusColor(%q) = %q, want %q", c.status, got, c.want)
		}
		if icon := StatusIcon(c.status); !strings.Contains(icon, c.letter) {
			t.Errorf("StatusIcon(%q) = %q, want contains letter %q", c.status, icon, c.letter)
		}
	}

	if got := StatusColor("some-unknown-status"); got != Subtext {
		t.Errorf("StatusColor(unknown) = %q, want fallback Subtext %q", got, Subtext)
	}
	if icon := StatusIcon("some-unknown-status"); !strings.Contains(icon, fallbackGlyphChar) {
		t.Errorf("StatusIcon(unknown) = %q, want fallback glyph %q", icon, fallbackGlyphChar)
	}
}

// TestAsciiFallback — BT_ASCII_ICONS ist seit PF-6 nur noch für Priorität relevant:
// Status-/Type-Buchstaben sind bereits ASCII/EAW-neutral und bleiben in BEIDEN Modi
// identisch (kein Unicode-vs-ASCII-Branch mehr). Per-Aufruf-Read via os.Getenv, daher
// mit t.Setenv pro Sub-Case steuerbar (kein Package-State-Caching).
func TestAsciiFallback(t *testing.T) {
	t.Setenv("BT_ASCII_ICONS", "1")

	if !strings.Contains(Priority("critical"), "!!") {
		t.Errorf("Priority(critical) with BT_ASCII_ICONS=1 should contain ASCII glyph %q", "!!")
	}
	if strings.Contains(Priority("critical"), "‼") {
		t.Errorf("Priority(critical) with BT_ASCII_ICONS=1 should not contain unicode glyph %q", "‼")
	}
	if !strings.Contains(StatusIcon("todo"), "t") {
		t.Errorf("StatusIcon(todo) with BT_ASCII_ICONS=1 should still contain letter %q (no-op for status)", "t")
	}
	if !strings.Contains(TypeIcon("bug"), "B") {
		t.Errorf("TypeIcon(bug) with BT_ASCII_ICONS=1 should still contain letter %q (no-op for type)", "B")
	}

	t.Setenv("BT_ASCII_ICONS", "0")

	if !strings.Contains(Priority("critical"), "‼") {
		t.Errorf("Priority(critical) without ascii mode should contain unicode glyph %q", "‼")
	}
	if strings.Contains(Priority("critical"), "!!") {
		t.Errorf("Priority(critical) without ascii mode should not contain ASCII glyph %q", "!!")
	}
}

// TestTypeIconAllTypes — alle 5 beans-Typen → erwarteter Buchstabe + Farbe (PF-6,
// design-spec.md §15, 2026-07-15). Unbekannter Typ fällt auf den generischen
// Fallback-Glyph zurück (siehe fallbackGlyph).
func TestTypeIconAllTypes(t *testing.T) {
	cases := []struct {
		typ   string
		glyph string
		color lipgloss.Color
	}{
		{"milestone", "M", Blue},
		{"epic", "E", Mauve},
		{"feature", "F", Mauve},
		{"task", "T", Sky},
		{"bug", "B", Red},
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

// TestSetAccentOverridesThenNoOpOnEmptyOrInvalid — E5 Task 5 (bean bt-0l8c):
// a valid #rrggbb hex overrides Accent/Header's foreground LIVE; an empty OR
// malformed hex is a No-Op -- the Golden-Risiko this guards (the bean's own
// PFLICHT wording): SetAccent must never blank out the built-in Mauve accent
// just because a caller passes "" (e.g. Settings.Theme.Accent's own
// "unset" zero value at every TUI start that never touched config.yaml) --
// the 7 golden snapshots render against that default and must stay
// byte-identical.
func TestSetAccentOverridesThenNoOpOnEmptyOrInvalid(t *testing.T) {
	origAccent, origHeader := Accent, Header
	t.Cleanup(func() { Accent, Header = origAccent, origHeader })

	SetAccent("#f5a97f")
	if got := Accent.GetForeground(); got != lipgloss.Color("#f5a97f") {
		t.Errorf("Accent.GetForeground() = %v, want #f5a97f", got)
	}
	if got := Header.GetForeground(); got != lipgloss.Color("#f5a97f") {
		t.Errorf("Header.GetForeground() = %v, want #f5a97f", got)
	}

	for _, invalid := range []string{"", "nope", "#zzzzzz", "#fff"} {
		Accent, Header = origAccent, origHeader // reset before each sub-case
		SetAccent(invalid)
		if got := Accent.GetForeground(); got != origAccent.GetForeground() {
			t.Errorf("SetAccent(%q): Accent.GetForeground() = %v, want unchanged %v (No-Op)", invalid, got, origAccent.GetForeground())
		}
		if got := Header.GetForeground(); got != origHeader.GetForeground() {
			t.Errorf("SetAccent(%q): Header.GetForeground() = %v, want unchanged %v (No-Op)", invalid, got, origHeader.GetForeground())
		}
	}
}

// TestPriorityColorMapping — beans-Priorität → erwartete Farbe + Glyph (PF-6,
// design-spec.md §15, 2026-07-15). Priority() liefert jetzt den Glyph statt des
// ausgeschriebenen Worts. Unbekannte Priorität fällt auf Text-Farbe + den
// generischen Fallback-Glyph zurück (deckt sich zufällig mit dem "normal"-Glyph).
func TestPriorityColorMapping(t *testing.T) {
	cases := []struct {
		priority string
		want     lipgloss.Color
		glyph    string
	}{
		{"critical", Red, "‼"},
		{"high", Yellow, "!"},
		{"normal", Text, "·"},
		{"low", Subtext, "↓"},
		{"deferred", Subtext, "→"},
	}
	for _, c := range cases {
		if got := priorityColor(c.priority); got != c.want {
			t.Errorf("priorityColor(%q) = %q, want %q", c.priority, got, c.want)
		}
		if rendered := Priority(c.priority); !strings.Contains(rendered, c.glyph) {
			t.Errorf("Priority(%q) = %q, want contains glyph %q", c.priority, rendered, c.glyph)
		}
	}

	rendered := Priority("critical")
	if strings.Contains(rendered, "critical") {
		t.Errorf("Priority(critical) = %q, want glyph only, not the word %q", rendered, "critical")
	}

	if got := priorityColor("unknown-priority"); got != Text {
		t.Errorf("priorityColor(unknown) = %q, want default Text %q", got, Text)
	}
	if rendered := Priority("unknown-priority"); !strings.Contains(rendered, fallbackGlyphChar) {
		t.Errorf("Priority(unknown) = %q, want contains fallback glyph %q", rendered, fallbackGlyphChar)
	}
}

// TestHeaderInactiveStyleIsTeal guards B06 (EXPERIMENT, design-spec.md §15
// PF-16, bean bt-ntoz/bt-czpf, 2026-07-16): the new HeaderInactive token --
// used for the Accordion's CLOSED Section-Header title (accordion.go) --
// must resolve to Teal (#8bd5ca), not the previous Hint-grey (Muted). PO
// sign-off on the experiment is still PENDING (see bean bt-czpf); a
// rejection is a one-line rollback (theme.HeaderInactive -> theme.Muted at
// accordion.go's single call site).
func TestHeaderInactiveStyleIsTeal(t *testing.T) {
	if got := HeaderInactive.GetForeground(); got != Teal {
		t.Errorf("HeaderInactive.GetForeground() = %v, want Teal %v", got, Teal)
	}
	if got := HeaderInactive.GetForeground(); got == Hint {
		t.Errorf("HeaderInactive.GetForeground() = %v, must not still be Hint-grey (Muted's color)", got)
	}
}

// TestBindingKeyDescStylesAreTealAndSubtext guards D06 (design-spec.md §15
// PF-16, bean bt-ntoz/bt-d8kc): the Header/Footer keybinding-hint tokens --
// key rendered Teal, description rendered Subtext -- and that BindingKey is
// its OWN token, independent of the B06-EXPERIMENT HeaderInactive (which
// happens to share the same Teal hex today but is a separate, still-pending
// decision -- see this file's own TestHeaderInactiveStyleIsTeal and
// theme.go's BindingKey doc comment).
func TestBindingKeyDescStylesAreTealAndSubtext(t *testing.T) {
	if got := BindingKey.GetForeground(); got != Teal {
		t.Errorf("BindingKey.GetForeground() = %v, want Teal %v", got, Teal)
	}
	if got := BindingDesc.GetForeground(); got != Subtext {
		t.Errorf("BindingDesc.GetForeground() = %v, want Subtext %v", got, Subtext)
	}
}
