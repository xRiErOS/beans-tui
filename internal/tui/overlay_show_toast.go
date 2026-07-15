package tui

// overlay_show_toast.go — E5 Task 1 (bean bt-6dts, epic bt-5h4d): color-coded
// corner Toast overlay. Port devd overlay_show_toast.go
// (~/Obsidian/tools/DeveloperDashboard/apps/cli-go/internal/tui/
// overlay_show_toast.go) structurally VERBATIM (epic-E5-plan.md design
// decision a) -- ONE slot (no stack, no stapeln): a new toast replaces the
// old one immediately. Auto-dismiss after a kind-specific duration
// (toastTimeout, messages.go) -- except sticky=true, which survives reload
// cycles AND has NO auto-dismiss timer, until a newer toast replaces it or a
// click closes it.
//
// Deviation vs. devd (design decision a): toastTarget carries ONLY `view
// viewID` -- devd's milestoneID/sprintID/issueID triad has no beans-tui
// equivalent (ONE entity, data.Bean, design-spec §4); a future
// deeper-focus-restore producer can extend this struct when it needs to, no
// producer currently does.
//
// Rendering: no new compositing primitive -- canvasLines()/spliceLine() (the
// same primitives placeOverlay() itself uses, overlay.go) place the box in
// the top-right corner instead of placeOverlay()'s centering. modalBox
// (modal.go) stays the box-chrome source.
//
// toastGeometry/renderToast compose directly against m.width/m.height,
// UNCHANGED vs. devd's own (already B01-fixed) source: beans-tui's three
// View functions (viewBrowseRepo/viewBacklog/viewReviewCockpit) each build
// their own full frame (w, h := m.width, m.height, outerBorder(..., true))
// exactly like devd's fully-composited View()/viewComposite() split does --
// no separate termWidth()-style inner-width helper exists here to
// mis-target against (design decision a, point 1).
//
// Input: NOT modal -- never blocks the keyboard. Only a left-click on the
// Toast hit area is intercepted ahead of the regular mouse dispatch (Task 4,
// mouse.go: handleMouse) -- dismiss, or (with a set target) a view switch
// plus dismiss.

import (
	"strings"
	"time"

	"beans-tui/internal/theme"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

type toastKind int

const (
	toastInfo  toastKind = iota // Blue/Mauve — Hinweis/Erfolg
	toastWarn                   // Yellow — Warnung, nicht blockierend
	toastError                  // Red — Fehler
)

// toastTarget addresses, minimally, where a click on the toast jumps to
// (devd DD2-272 AC4 parity). Only view -- beans-tui has ONE focus resolver
// (focusedBean(), view-agnostic across Tree/Backlog/Review-Cockpit), so a
// view switch alone is enough to land the click near the right context; no
// producer needs a deeper bean/section jump yet.
type toastTarget struct {
	view viewID
}

// toastState is the ONE toast slot (no stack). seq = generation: a
// toastExpiredMsg only clears the toast when seq still matches the current
// generation (otherwise a newer toast has already replaced it).
type toastState struct {
	kind    toastKind
	title   string
	context string // second, dimmed line (optional)
	target  *toastTarget
	seq     int
	sticky  bool
	setAt   time.Time // debounce window (AC5 parity)
}

// toastDebounceWindow: a second showToast call within this span of the last
// toast replaces its content in place instead of starting a new generation +
// a new auto-dismiss tick (AC5 parity) -- ONE timer, never stacked.
const toastDebounceWindow = 300 * time.Millisecond

// toastDuration returns the kind-specific auto-dismiss duration.
func toastDuration(kind toastKind) time.Duration {
	switch kind {
	case toastError:
		return 8 * time.Second
	case toastWarn:
		return 3 * time.Second
	default:
		return 5 * time.Second
	}
}

// showToast sets the toast slot. Within the debounce window (AC5 parity) the
// existing slot is updated in place (content/kind/target), WITHOUT a new
// generation/timer -- the already-running auto-dismiss tick stays the single
// source of truth. Sticky toasts (AC2 parity) get NO auto-dismiss timer.
func (m model) showToast(kind toastKind, title, context string, target *toastTarget, sticky bool) (model, tea.Cmd) {
	now := time.Now()
	if m.toast != nil && now.Sub(m.toast.setAt) < toastDebounceWindow {
		t := *m.toast
		t.kind, t.title, t.context, t.target, t.sticky, t.setAt = kind, title, context, target, sticky, now
		m.toast = &t
		return m, nil // debounce: no new tick, the running one stays valid
	}
	seq := 1
	if m.toast != nil {
		seq = m.toast.seq + 1
	}
	m.toast = &toastState{kind: kind, title: title, context: context, target: target, seq: seq, sticky: sticky, setAt: now}
	if sticky {
		return m, nil // AC2 parity: sticky = no auto-dismiss, until replace/click
	}
	return m, toastTimeout(seq, kind)
}

// clearToastUnlessSticky clears the toast ONLY when it is not sticky --
// reload handlers must never clobber a sticky toast (E3-I01 PFLICHT: the
// conflict + recovery-tempfile path must stay readable until the PO acts).
func (m model) clearToastUnlessSticky() model {
	if m.toast != nil && !m.toast.sticky {
		m.toast = nil
	}
	return m
}

// toastKindColor maps kind -> border/accent color.
func toastKindColor(k toastKind) lipgloss.Color {
	switch k {
	case toastError:
		return theme.Red
	case toastWarn:
		return theme.Yellow
	default:
		return theme.Blue
	}
}

const toastBoxWidth = 36 // target width 32-40 cols (devd DD2-272 parity)

// toastBox renders the toast box: line 1 = colored dot + title (kind-tinted),
// line 2 = context, dimmed (ansi.Truncate, NEVER string-slicing/len()). The
// border follows the same kind color.
func (m model) toastBox() string {
	t := m.toast
	col := toastKindColor(t.kind)
	innerW := toastBoxWidth - 4 // modalBox: Border(2) + Padding(0,1)*2 = 4
	dot := lipgloss.NewStyle().Foreground(col).Render("●")
	title := ansi.Truncate(t.title, innerW-2, "…") // -2: dot + space
	line1 := dot + " " + lipgloss.NewStyle().Foreground(col).Bold(true).Render(title)
	body := line1
	if t.context != "" {
		ctx := ansi.Truncate(t.context, innerW, "…")
		body += "\n" + theme.Dim.Render(ctx)
	}
	// modalBox sets lipgloss.Width(width) on the style BEFORE the border --
	// the border then adds another +2 (left/right, 1 each) to the total
	// width. toastBoxWidth-2 compensates so the rendered box is exactly
	// toastBoxWidth wide (target 32-40, devd DD2-272 parity).
	return modalBox(body, toastBoxWidth-2, col)
}

// toastGeometry returns the toast box's placement (top-right corner) on the
// canvas -- shared geometry between rendering (renderToast) and the
// click-hit-test (mouse.go: handleMouse, Task 4), so they can never drift
// apart. Computed directly against m.width/m.height -- see this file's own
// doc comment (top) for why no termWidth()-style adaptation is needed here
// (design decision a, point 1): every one of beans-tui's View functions
// already builds its own full, bordered frame on m.width/m.height before
// composeOverlays runs.
func (m model) toastGeometry() (x, y, w, h int) {
	box := m.toastBox()
	lines := strings.Split(box, "\n")
	w = 0
	for _, l := range lines {
		if lw := ansi.StringWidth(l); lw > w {
			w = lw
		}
	}
	h = len(lines)
	tw := m.width
	x = tw - w
	if x < 0 {
		x = 0
	}
	y = 0
	return x, y, w, h
}

// renderToast lays the toast box over the finished base view (top-right).
// Reuses canvasLines()/spliceLine() (the same primitives placeOverlay() uses
// internally, overlay.go) instead of placeOverlay()'s centering -- no new
// compositing primitive, just different coordinates. base is already the
// full, finished frame (including the outer border) -- composed against
// m.width/m.height (see toastGeometry's doc comment).
func (m model) renderToast(base string) string {
	if m.toast == nil {
		return base
	}
	x, y, w, _ := m.toastGeometry()
	tw, th := m.width, m.height
	bgLines := canvasLines(base, tw, th)
	fgLines := strings.Split(m.toastBox(), "\n")
	for i, fl := range fgLines {
		row := y + i
		if row < 0 || row >= len(bgLines) {
			continue
		}
		if pad := w - ansi.StringWidth(fl); pad > 0 {
			fl += overlayPad.Render(strings.Repeat(" ", pad))
		}
		bgLines[row] = spliceLine(bgLines[row], fl, x, w)
	}
	return strings.Join(bgLines, "\n")
}

// toastHit reports whether (mx, my) lies within the currently rendered toast
// box.
func (m model) toastHit(mx, my int) bool {
	if m.toast == nil {
		return false
	}
	x, y, w, h := m.toastGeometry()
	return mx >= x && mx < x+w && my >= y && my < y+h
}

// dismissToast closes the toast (click, devd DD2-272 AC4 parity). With a set
// target it first switches to that view -- minimal addressing (target only
// carries what a future producer needs for deeper focus-restore; no current
// producer sets more than the view).
func (m model) dismissToast() (tea.Model, tea.Cmd) {
	if m.toast != nil && m.toast.target != nil {
		m.view = m.toast.target.view
	}
	m.toast = nil
	return m, nil
}
