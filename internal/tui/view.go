package tui

// view.go — the shared chrome layer: breadcrumb/footer/status-line rendering,
// masterDetailWidths(), the outer RoundedBorder frame and the four-zone hull
// (header row `> repo: Titel` + global shortcuts right; scrollable body;
// footer hints; status line). Ported from devd (~/Obsidian/tools/Developer
// Dashboard/apps/cli-go/internal/tui/view.go).
//
// Decoupling from devd: every devd chrome function is a `model` method (m
// model) and reaches into m.project/m.cfg/m.width/... . beans-tui has no
// model yet (T8 builds the App-Shell) and the chrome layer must not import
// the data layer, so every function here takes plain strings/ints instead —
// Chrome/ChromeOpts bundle the render inputs a future model will supply.
// devd's metaGrid/metaStrip/issueMetaPairs/detailTitle (Detail-view header
// slots, api.Issue-coupled) and screenTitle (project-prefix titles) are NOT
// ported: out of the "four-zone hull" this task scopes (no optional info-
// grid zone), and/or view-specific — they land with the view that needs them.

import (
	"fmt"
	"strings"

	"beans-tui/internal/theme"
	keybind "github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// breadcrumb renders header Zone 1 (devd wireframe): left `> repo: Title`
// (chevron+repo Peach, title Mauve bold), right = pre-rendered global
// shortcuts hint. title="" → only `> repo`. Port-adaptation: devd used the
// project slug ("> dd: Title"); beans-tui uses the repo name instead
// (design-spec.md §7/implementation-plan Task 7: "Breadcrumb-Format `> repo:
// Titel`").
func breadcrumb(repo, title, globalHint string, width int) string {
	left := theme.Chevron.Render("> " + repo)
	if title != "" {
		left += theme.Chevron.Render(":") + " " + theme.Header.Render(title)
	}
	right := theme.Muted.Render(globalHint)
	if lipgloss.Width(left)+lipgloss.Width(right)+1 > width { // narrow: stack instead of overflow
		return ansi.Truncate(left, width, "…") + "\n" + ansi.Truncate(right, width, "…")
	}
	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + right
}

// wrapText ANSI-safely wraps s to w cells: word-wrap first, then hard breaks
// for over-long tokens (URLs/paths) that would otherwise overflow (devd
// DD2-60 review finding). Preserves existing line breaks.
func wrapText(s string, w int) string {
	if w < 1 {
		w = 1
	}
	return ansi.Hardwrap(ansi.Wordwrap(s, w, ""), w, true)
}

// statusBar renders footer Zone 4 (devd wireframe): scroll indicator
// (Accent) + a critical error (Red), right-aligned. Empty → empty line.
func statusBar(indicator, errNote string, width int) string {
	var rparts []string
	if indicator != "" {
		rparts = append(rparts, theme.Accent.Render(indicator))
	}
	if errNote != "" {
		rparts = append(rparts, lipgloss.NewStyle().Foreground(theme.Red).Render(errNote))
	}
	right := strings.Join(rparts, "  ")
	if right == "" {
		return ""
	}
	if lipgloss.Width(right)+1 > width {
		return ansi.Truncate(right, width, "…")
	}
	gap := width - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}
	return strings.Repeat(" ", gap) + right
}

// renderBindings renders a binding list as a "key:desc" hint, joined by two
// spaces. Bindings without a Help().Key are skipped (devd DD2-175: footer
// hints are derived from the keymap, single source, can never drift from the
// real bindings).
func renderBindings(bs []keybind.Binding) string {
	parts := make([]string, 0, len(bs))
	for _, b := range bs {
		h := b.Help()
		if h.Key != "" {
			parts = append(parts, h.Key+":"+h.Desc)
		}
	}
	return strings.Join(parts, "  ")
}

// footer renders footer Zone 3 (local + global hints, already joined into
// hint) dimmed and wrapped to width — narrow terminals wrap instead of
// overflowing into the pane columns (mirrors breadcrumb/chrome).
func footer(hint string, width int) string {
	return theme.Dim.Render(wrapText(hint, width))
}

// truncate ANSI-safely truncates s to w cells (never cuts an escape
// sequence) — critical since lines carry color; rune-slicing would destroy
// sequences.
func truncate(s string, w int) string {
	if w < 1 {
		return ""
	}
	return ansi.Truncate(s, w, "…")
}

// masterDetailWidths splits the master-detail pane width as a true 1fr:2fr:
// the list on the left gets a third (w/3), the detail on the right gets the
// rest (rw = w - lw - 4, 2 border columns per pane). treeWidthFloor (devd:
// cfg.Layout.TreeWidth) is only a minimum for readability on narrow
// terminals, capped at w*2/5 — not a fixed width on wide terminals (devd
// DD2-91 rework).
func masterDetailWidths(w, treeWidthFloor int) (lw, rw int) {
	lw = w / 3 // 1fr
	if treeWidthFloor > 0 && lw < treeWidthFloor {
		lw = treeWidthFloor
	}
	if lw < 24 {
		lw = 24
	}
	if cap := w * 2 / 5; lw > cap {
		lw = cap
	}
	rw = w - lw - 4
	if rw < 20 {
		rw = 20
	}
	return
}

// outerBorder wraps content in the app's outer frame (RoundedBorder,
// Overlay-colored) — a no-op when bordered is false. width is the full
// (outer) width; content must already fill the inner width/height, i.e. NO
// Height() here (Golden Rule #1) — the frame grows naturally around it.
func outerBorder(content string, width int, bordered bool) string {
	if !bordered {
		return content
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Overlay).
		Width(width).
		Render(content)
}

// scrollView windows content to height lines from offset (clamped) and pads
// with blank lines to height so the footer stays glued to the bottom.
// Returns a scroll indicator too (empty when everything fits).
func scrollView(content string, height, offset int) (string, string) {
	if height < 1 {
		height = 1
	}
	lines := strings.Split(content, "\n")
	total := len(lines)
	maxOff := total - height
	if maxOff < 0 {
		maxOff = 0
	}
	if offset > maxOff {
		offset = maxOff
	}
	if offset < 0 {
		offset = 0
	}
	end := offset + height
	if end > total {
		end = total
	}
	win := append([]string{}, lines[offset:end]...)
	for len(win) < height {
		win = append(win, "")
	}
	ind := ""
	if total > height {
		ind = fmt.Sprintf("L %d–%d/%d", offset+1, end, total)
		switch {
		case offset == 0:
			ind += " ↓"
		case end >= total:
			ind = "↑ " + ind
		default:
			ind = "↑ " + ind + " ↓"
		}
	}
	return strings.Join(win, "\n"), ind
}

// ChromeOpts bundles the render inputs for Chrome — the plain-data
// equivalent of the fields a future model (T8) will hold.
type ChromeOpts struct {
	Width, Height int
	Bordered      bool
	Repo          string // breadcrumb "> repo"
	Title         string // breadcrumb ": Title" (skipped when empty)
	GlobalHint    string // pre-rendered "key:desc  key:desc  ..." right of breadcrumb
	Body          string // unwrapped body text; wrapped+scrolled internally
	Scroll        int    // scroll offset into the wrapped body
	FooterHint    string // pre-rendered local+global "key:desc  ..." hint
	ErrNote       string // critical error, right-aligned in the status line
	fallbackAvail int    // test hook: override the body-height fallback (0 = default 18)
}

// Chrome is the shared screen chrome (devd DD2-48): breadcrumb header,
// height-filling scroll body, footer hints + status line — wrapped in the
// outer RoundedBorder frame when Bordered is set.
func Chrome(o ChromeOpts) string {
	innerW := o.Width
	innerH := o.Height
	if o.Bordered {
		innerW -= 2
		innerH -= 2
	}
	head := breadcrumb(o.Repo, o.Title, o.GlobalHint, innerW)
	wrapped := lipgloss.NewStyle().Padding(0, 1).Render(wrapText(o.Body, innerW-2))
	localKeys := footer(o.FooterHint, innerW)
	div := theme.Dim.Render(strings.Repeat("─", innerW))
	footH := lipgloss.Height(localKeys) + 2             // + status line + divider above the footer
	avail := innerH - lipgloss.Height(head) - footH - 1 // - divider under the top bar
	if avail < 4 {
		if o.fallbackAvail > 0 {
			avail = o.fallbackAvail
		} else {
			avail = 18 // height unknown (init/tests) → generous fallback
		}
	}
	win, ind := scrollView(wrapped, avail, o.Scroll)
	status := statusBar(ind, o.ErrNote, innerW)
	content := head + "\n" + div + "\n" + win + "\n" + div + "\n" + localKeys + "\n" + status
	// outerBorder's width param is the CONTENT width the border wraps around
	// (lipgloss's Border() adds 2 columns on top of Width()) — pass innerW
	// (already reduced above), not o.Width, or the frame ends up o.Width+2
	// wide instead of o.Width.
	return outerBorder(content, innerW, o.Bordered)
}
