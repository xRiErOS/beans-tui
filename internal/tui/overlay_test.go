package tui

// overlay_test.go — TDD coverage for Slice B (bt-f0y9, "feld-verankertes
// Inline-Dropdown", D09 revidiert): placeOverlayAt, the anchor-positioned
// sibling of placeOverlay (overlay.go). Unit tests against small, fixed
// bg/fg string pairs (render-grounded in spirit — the ASSERTION reads the
// actual composited output back via ansi.Strip, never hand-predicts the
// splice), not wired into composeOverlays yet (Slice C).

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// fixedBg builds a th-line, tw-wide background of distinct per-cell markers
// ("." repeated) so a spliced-in fg is trivially visible against it.
func fixedBg(tw, th int) string {
	line := strings.Repeat(".", tw)
	lines := make([]string, th)
	for i := range lines {
		lines[i] = line
	}
	return strings.Join(lines, "\n")
}

// TestPlaceOverlayAtPositionsAtExactAnchor guards the core contract: fg
// lands with its top-left cell at exactly (x, y), not centered.
func TestPlaceOverlayAtPositionsAtExactAnchor(t *testing.T) {
	bg := fixedBg(20, 10)
	fg := "AB\nCD" // 2x2 block

	out := placeOverlayAt(bg, fg, 20, 10, 5, 3)
	lines := strings.Split(ansi.Strip(out), "\n")

	if len(lines) != 10 {
		t.Fatalf("len(lines) = %d, want 10 (th preserved)", len(lines))
	}
	if got := lines[3][5:7]; got != "AB" {
		t.Fatalf("row 3 cols[5:7] = %q, want \"AB\" (top-left anchor)", got)
	}
	if got := lines[4][5:7]; got != "CD" {
		t.Fatalf("row 4 cols[5:7] = %q, want \"CD\"", got)
	}
	// Everything strictly left of the anchor column on the fg's own rows
	// must be untouched background.
	if got := lines[3][:5]; got != "....." {
		t.Fatalf("row 3 cols[:5] = %q, want untouched background", got)
	}
	// A row the fg does NOT occupy must be fully untouched background.
	if got := lines[0]; got != strings.Repeat(".", 20) {
		t.Fatalf("row 0 = %q, want untouched background (fg starts at row 3)", got)
	}
}

// TestPlaceOverlayAtBottomOverflowDropsExcessRowsSilently guards the
// documented overflow contract (bt-f0y9 Slice B grounding Q3): fg rows that
// would land past the bottom of bg are silently dropped -- the SAME
// no-clip/no-flip behavior placeOverlay itself has always had (overlay.go's
// own doc comment), not a crash and not an upward flip (explicitly out of
// this slice's scope).
func TestPlaceOverlayAtBottomOverflowDropsExcessRowsSilently(t *testing.T) {
	bg := fixedBg(10, 4)
	fg := "X1\nX2\nX3\nX4\nX5" // 5 rows, anchored at y=2 -> rows 2,3 fit, 4,5,6 overflow

	out := placeOverlayAt(bg, fg, 10, 4, 3, 2)
	lines := strings.Split(ansi.Strip(out), "\n")

	if len(lines) != 4 {
		t.Fatalf("len(lines) = %d, want 4 (bg height preserved, no growth/panic on overflow)", len(lines))
	}
	if got := lines[2][3:5]; got != "X1" {
		t.Fatalf("row 2 cols[3:5] = %q, want \"X1\"", got)
	}
	if got := lines[3][3:5]; got != "X2" {
		t.Fatalf("row 3 cols[3:5] = %q, want \"X2\"", got)
	}
	// Rows past bg's own height simply don't exist -- no panic, no growth.
}

// TestPlaceOverlayAtNegativeXClampsToZero guards the defensive x<0 clamp
// (a negative anchor column is a caller bug, not a valid "off left edge"
// request this primitive tries to interpret).
func TestPlaceOverlayAtNegativeXClampsToZero(t *testing.T) {
	bg := fixedBg(10, 3)
	fg := "ZZ"

	out := placeOverlayAt(bg, fg, 10, 3, -4, 1)
	lines := strings.Split(ansi.Strip(out), "\n")

	if got := lines[1][:2]; got != "ZZ" {
		t.Fatalf("row 1 cols[:2] = %q, want \"ZZ\" (negative x clamped to 0)", got)
	}
}

// TestPlaceOverlayAtNegativeYDropsRowsAboveTop guards the mirrored top-edge
// case: fg rows landing above row 0 are dropped the same way sub-bottom rows
// are, no crash.
func TestPlaceOverlayAtNegativeYDropsRowsAboveTop(t *testing.T) {
	bg := fixedBg(10, 3)
	fg := "R0\nR1\nR2" // anchored at y=-1 -> R0 lands at row -1 (dropped), R1 at row 0, R2 at row 1

	out := placeOverlayAt(bg, fg, 10, 3, 2, -1)
	lines := strings.Split(ansi.Strip(out), "\n")

	if len(lines) != 3 {
		t.Fatalf("len(lines) = %d, want 3", len(lines))
	}
	if got := lines[0][2:4]; got != "R1" {
		t.Fatalf("row 0 cols[2:4] = %q, want \"R1\"", got)
	}
	if got := lines[1][2:4]; got != "R2" {
		t.Fatalf("row 1 cols[2:4] = %q, want \"R2\"", got)
	}
}

// TestPlaceOverlaySharesCompositingCoreWithPlaceOverlayAt is the "Golden-
// Rule-Drift-Schutz" cross-check the bean's own text asks for (panelBox/
// panelBoxTopHotkey sharing panelBoxWith, the named precedent): placeOverlay
// centered at bgW=20/bgH=10/fgW=2/fgH=2 must produce the EXACT same output
// as placeOverlayAt anchored at that same computed (x, y) -- proving both
// variants run through one shared splice, not two independently drifting
// implementations.
func TestPlaceOverlaySharesCompositingCoreWithPlaceOverlayAt(t *testing.T) {
	bg := fixedBg(20, 10)
	fg := "AB\nCD"

	centered := placeOverlay(bg, fg, 20, 10)
	anchored := placeOverlayAt(bg, fg, 20, 10, (20-2)/2, (10-2)/2)

	if centered != anchored {
		t.Fatalf("placeOverlay and placeOverlayAt diverged for the same effective (x,y):\ncentered=%q\nanchored=%q", centered, anchored)
	}
}
