package theme

import "github.com/charmbracelet/lipgloss"

// --- beans-Type Text-Icons (kein Emoji — Unicode-Geometrie, monospace-sicher) ---
//
// Port-Adaptation 3 (devd internal/theme/theme.go typeIcon/typeColor): devd kannte
// bug/feature/improvement/core; beans kennt milestone/epic/feature/task/bug — die
// beiden Sets überlappen nur in bug/feature, die übrigen sind neu gemappt. Alle
// Glyphen hier MÜSSEN East-Asian-Width = Neutral/Narrow sein (nie Ambiguous), sonst
// verrutschen Spalten in Ambiguous=Wide-Terminals (tmux/CJK) — lipgloss zählt Breite
// via clipperhouse/displaywidth, nicht go-runewidth.
var typeIcon = map[string]string{
	"milestone": "⬢", // U+2B22 BLACK HEXAGON (neutral)
	"epic":      "✦", // U+2726 BLACK FOUR POINTED STAR (neutral)
	"feature":   "✦", // U+2726 — gleicher Glyph wie epic, Unterscheidung über Farbe
	"task":      "⯅", // U+2BC5 BLACK MEDIUM UP-POINTING TRIANGLE (neutral)
	"bug":       "⯁", // U+2BC1 BLACK MEDIUM DIAMOND (neutral)
}

var typeColor = map[string]lipgloss.Color{
	"milestone": Peach,
	"epic":      Mauve,
	"feature":   Green,
	"task":      Blue,
	"bug":       Red,
}

// typeIconASCII — ASCII-Ersatz je Typ (garantiert darstellbar, EAW-Neutral), aktiv via
// BT_ASCII_ICONS. Distinkt pro Typ, damit der Typ ohne Farbe erkennbar bleibt.
var typeIconASCII = map[string]string{
	"milestone": "#",
	"epic":      "*",
	"feature":   "+",
	"task":      "^",
	"bug":       "!",
}

// typeGlyph wählt den Type-Glyph (ASCII-Fallback berücksichtigt). Unbekannte Typen
// bekommen den generischen Fallback-Glyph (fallbackGlyph, siehe theme.go).
func typeGlyph(t string) string {
	ic, ok := typeIcon[t]
	if !ok {
		return fallbackGlyph()
	}
	if asciiIcons() {
		return typeIconASCII[t]
	}
	return ic
}

// TypeIcon liefert das gefärbte Text-Icon eines beans-Typs (Fallback: Text-Farbe +
// fallbackGlyph für unbekannte Typen).
func TypeIcon(t string) string {
	col, ok := typeColor[t]
	if !ok {
		col = Text
	}
	return lipgloss.NewStyle().Foreground(col).Render(typeGlyph(t))
}

// TypeStyle färbt einen Typ-Text (Fallback: Text-Farbe für unbekannte Typen).
func TypeStyle(t string) lipgloss.Style {
	col, ok := typeColor[t]
	if !ok {
		col = Text
	}
	return lipgloss.NewStyle().Foreground(col)
}
