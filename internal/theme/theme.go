// Package theme hält die geteilte Catppuccin-Palette + Status-/Priority-Styles für die
// beans-tui. Port von Eriks devd-TUI (~/Obsidian/tools/DeveloperDashboard/apps/cli-go/
// internal/theme/theme.go) — Farb-Tokens 1:1 übernommen, Enum-Mapping auf beans-Status/
// -Priority angepasst (docs/plans/v1-port/implementation-plan.md »E1 Task 6«).
// Macchiato-Variante, TrueColor-Hex (Akzent = Mauve).
package theme

import (
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// asciiIcons — Opt-in ASCII-Ersatz-Glyphen für Terminals/Fonts, die die (EAW-neutrale,
// aber font-seitig schwach abgedeckte) Unicode-Geometrie nicht darstellen. ASCII ist
// garantiert verfügbar UND EAW=Neutral (kein Spalten-Drift). Aktiv via
// BT_ASCII_ICONS=1|true|yes|on (Port-Adaptation 1: devd DEVD_ASCII_ICONS → BT_ASCII_ICONS).
// Pro Aufruf gelesen (Render ist nicht perf-kritisch) → auch in Tests via t.Setenv
// setz-/rücksetzbar, kein Package-State-Caching.
func asciiIcons() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("BT_ASCII_ICONS"))) {
	case "1", "true", "yes", "on":
		return true
	}
	return false
}

// Catppuccin Macchiato — TrueColor-Hex. 1:1 aus devd internal/theme/theme.go übernommen.
var (
	Rosewater = lipgloss.Color("#f4dbd6")
	Flamingo  = lipgloss.Color("#f0c6c6")
	Pink      = lipgloss.Color("#f5bde6")
	Mauve     = lipgloss.Color("#c6a0f6") // Akzent
	Red       = lipgloss.Color("#ed8796")
	Maroon    = lipgloss.Color("#ee99a0")
	Peach     = lipgloss.Color("#f5a97f")
	Yellow    = lipgloss.Color("#eed49f")
	Green     = lipgloss.Color("#a6da95")
	Teal      = lipgloss.Color("#8bd5ca")
	Sky       = lipgloss.Color("#91d7e3")
	Sapphire  = lipgloss.Color("#7dc4e4")
	Blue      = lipgloss.Color("#8aadf4")
	Lavender  = lipgloss.Color("#b7bdf8")

	Text    = lipgloss.Color("#cad3f5")
	Subtext = lipgloss.Color("#a5adcb")
	Overlay = lipgloss.Color("#8087a2") // Input-/Feld-Border
	Surface = lipgloss.Color("#494d64")
	Base    = lipgloss.Color("#24273a") // App-/Body-Hintergrund
	Mantle  = lipgloss.Color("#1e2030") // Header-/Accordion-Header-BG
	Crust   = lipgloss.Color("#181926") // Form-Rahmen-BG (Modal-Backdrop)

	// Hint = Hinweis/Erklärung (Labels, Sub-Label, Shortcuts, Placeholder, inaktive
	// Tabs). Zwei-Klassen-Text-Regel: muted ggü. echter Info. Bewusst ≠ Overlay
	// (#8087a2 = Feld-Border).
	Hint = lipgloss.Color("#7c7c7c")
	// Select = Interaktions-/Auswahl-Signal (aktiver Auswahl-Button), laut
	// (Latte-Peach). Bewusst ≠ Peach #f5a97f (struktureller Akzent).
	Select = lipgloss.Color("#fe640b")

	Header = lipgloss.NewStyle().Bold(true).Foreground(Mauve)
	Key    = lipgloss.NewStyle().Foreground(Sapphire) // IDs/Keys = Sapphire
	Accent = lipgloss.NewStyle().Foreground(Mauve)

	Dim = lipgloss.NewStyle().Foreground(Overlay)
	// Muted = Hinweis/Erklärung: Shortcuts, Sub-Label, Placeholder. Bewusst Hint
	// #7c7c7c, nicht Overlay (= Feld-Border) — Zwei-Klassen-Text-Prinzip.
	Muted = lipgloss.NewStyle().Foreground(Hint)
	// Chevron = struktureller Marker `>`/`o`, Peach.
	Chevron = lipgloss.NewStyle().Foreground(Peach)
)

// SetAccent überschreibt den Akzentstil (Cursor/Header) mit einer User-Farbe (z.B. aus
// einer künftigen Theme-Config). hex muss bereits validiert sein (#rrggbb). Globale
// Theme-Mutation — nur einmal beim TUI-Start aufrufen.
func SetAccent(hex string) {
	c := lipgloss.Color(hex)
	Accent = lipgloss.NewStyle().Foreground(c)
	Header = lipgloss.NewStyle().Bold(true).Foreground(c)
}

// --- beans-Status: draft/todo/in-progress/completed/scrapped (Port-Adaptation 2) ---

var statusColor = map[string]lipgloss.Color{
	"draft":       Blue,
	"todo":        Text,
	"in-progress": Yellow,
	"completed":   Green,
	"scrapped":    Red,
}

// StatusColor liefert die Farbe für einen beans-Status (Fallback: Text — unbekannte
// Enum-Werte neutral statt falsch-signalisierend; siehe fallbackGlyph für den
// zugehörigen Glyph-Fallback).
func StatusColor(status string) lipgloss.Color {
	if col, ok := statusColor[status]; ok {
		return col
	}
	return Text
}

// StatusStyle liefert den lipgloss-Style für einen Status (Default: Text via StatusColor).
func StatusStyle(status string) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(StatusColor(status))
}

// statusGlyph — EIN gemeinsamer Glyph für ALLE Status (devd-Prinzip DD2-176): die
// Bedeutung trägt allein die Farbe (StatusColor/StatusStyle), nicht die Form.
const statusGlyph = "◉"      // U+25C9 FISHEYE (EAW=Neutral)
const statusGlyphASCII = "*" // ASCII-Ersatz (garantiert darstellbar, EAW=Neutral)

// statusIconGlyph wählt den Status-Glyph (ASCII-Fallback berücksichtigt). Unbekannte
// Status bekommen den generischen Fallback-Glyph statt eines eigenen Zeichens.
func statusIconGlyph(status string) string {
	if _, ok := statusColor[status]; !ok {
		return fallbackGlyph()
	}
	if asciiIcons() {
		return statusGlyphASCII
	}
	return statusGlyph
}

// StatusIcon liefert den einheitlichen, statusgefärbten Status-Glyph.
func StatusIcon(status string) string {
	return StatusStyle(status).Render(statusIconGlyph(status))
}

// --- Fallback für unbekannte Enum-Werte (Status/Type) ---
//
// Dokumentierte Design-Entscheidung (Task-6-Vorgabe): unbekannte beans-Enum-Werte
// (z.B. ein künftig hinzugefügter Status/Type) rendern neutral statt zu verschwinden
// oder falsch zu signalisieren — Farbe Text, Glyph „·" (U+00B7 MIDDLE DOT). Hinweis:
// U+00B7 ist laut Unicode East-Asian-Width-Tabelle als Ambiguous klassifiziert (anders
// als die übrigen hier verwendeten Glyphen, die bewusst EAW=Neutral/Narrow sind, um
// Spalten-Drift in Ambiguous=Wide-Terminals zu vermeiden). Für den seltenen
// Fallback-Pfad (unbekannter Enum-Wert) wird das in Kauf genommen — exakt der in der
// Aufgabenstellung vorgegebene Glyph.
const fallbackGlyphChar = "·"      // U+00B7 MIDDLE DOT
const fallbackGlyphCharASCII = "." // ASCII-Ersatz

func fallbackGlyph() string {
	if asciiIcons() {
		return fallbackGlyphCharASCII
	}
	return fallbackGlyphChar
}

// --- beans-Priorität: critical/high/normal/low/deferred (Port-Adaptation 4) ---

// priorityColor: critical/high dringend (Rot), normal neutral (Text), low entspannt
// (Grün), deferred zurückgestellt (Hint). Unbekannte Werte neutral (Text).
func priorityColor(p string) lipgloss.Color {
	switch p {
	case "critical", "high":
		return Red
	case "low":
		return Green
	case "deferred":
		return Hint
	default: // "normal" und unbekannte Werte
		return Text
	}
}

// Priority rendert eine beans-Priorität gefärbt (critical/high fett zur Hervorhebung).
func Priority(p string) string {
	st := lipgloss.NewStyle().Foreground(priorityColor(p))
	if p == "critical" || p == "high" {
		st = st.Bold(true)
	}
	return st.Render(p)
}
