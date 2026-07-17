package tui

// overlay.go — ANSI-safe overlay compositing (painter's algorithm), ported
// unchanged from devd (~/Obsidian/tools/DeveloperDashboard/apps/cli-go/
// internal/tui/overlay.go). No devd-API coupling to strip — pure string/int
// primitives already. This logic is subtle (byte-faithful line splicing on
// ANSI-escaped strings); ported verbatim, only the theme import path changed.

import (
	"strings"

	"github.com/xRiErOS/beans-tui/internal/theme"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// overlayPad is the background fill for narrow modal lines. Every modal box
// uses theme.Base as background — the padding carries the same color so no
// terminal-default stripe appears (otherwise a form footer would look
// perforated).
var overlayPad = lipgloss.NewStyle().Background(theme.Base)

// canvasLines normalizes bg into a gapless tw×th rectangle with theme.Base
// background (devd DD2-216): every cell the overlay doesn't cover itself —
// stripes left/right of the box, rows below the frame — is then Base instead
// of the terminal default (usually black). spliceLine subsequently cuts from
// a fully Base-filled bg line, so no more plain-space padding leaks through.
// Root-fix at the canvas rather than patched per consumer.
func canvasLines(bg string, tw, th int) []string {
	lines := strings.Split(bg, "\n")
	if tw <= 0 {
		return lines // width unknown (init/tests) → leave bg unchanged
	}
	n := len(lines)
	if th > n {
		n = th
	}
	out := make([]string, 0, n)
	for i := 0; i < n; i++ {
		var l string
		if i < len(lines) {
			l = lines[i]
		}
		if pad := tw - ansi.StringWidth(l); pad > 0 {
			l += overlayPad.Render(strings.Repeat(" ", pad))
		}
		out = append(out, l)
	}
	return out
}

// placeOverlay places fg centered over bg (painter's algorithm, ANSI-safe).
// bg remains visible all around — only the cells the modal occupies get
// overwritten. There's no z-index in Bubble Tea, hence line-compositing.
func placeOverlay(bg, fg string, tw, th int) string {
	bgLines := canvasLines(bg, tw, th) // devd DD2-216: fully Base-filled canvas
	fgLines := strings.Split(fg, "\n")

	fgW := 0
	for _, l := range fgLines {
		if w := ansi.StringWidth(l); w > fgW {
			fgW = w
		}
	}
	fgH := len(fgLines)

	bgW := tw
	if bgW <= 0 {
		for _, l := range bgLines {
			if w := ansi.StringWidth(l); w > bgW {
				bgW = w
			}
		}
	}

	x := (bgW - fgW) / 2
	if x < 0 {
		x = 0
	}
	y := (len(bgLines) - fgH) / 2
	if y < 0 {
		y = 0
	}

	for i, fl := range fgLines {
		row := y + i
		if row < 0 || row >= len(bgLines) {
			continue
		}
		// Pad every fg line to the uniform modal width fgW, otherwise
		// narrower lines (e.g. a form's helper line/blank lines) don't cover
		// the background and text behind bleeds through. The modal must be a
		// gapless rectangle.
		if pad := fgW - ansi.StringWidth(fl); pad > 0 {
			fl += overlayPad.Render(strings.Repeat(" ", pad))
		}
		bgLines[row] = spliceLine(bgLines[row], fl, x, fgW)
	}
	return strings.Join(bgLines, "\n")
}

// spliceLine replaces cells [x, x+fgW) of bg with fg (ANSI-safe).
func spliceLine(bg, fg string, x, fgW int) string {
	left := ansi.Truncate(bg, x, "")
	if lw := ansi.StringWidth(left); lw < x {
		left += strings.Repeat(" ", x-lw)
	}
	right := ansi.TruncateLeft(bg, x+fgW, "")
	return left + "\x1b[0m" + fg + "\x1b[0m" + right
}
