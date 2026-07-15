package tui

// view_lobby.go — E5 Task 6 (bean bt-zhwl, epic bt-5h4d, design-spec.md §6
// V1): the Lobby/Repo-Picker. Port devd view_home.go
// (~/Obsidian/tools/DeveloperDashboard/apps/cli-go/internal/tui/
// view_home.go) STRUCTURALLY -- centerInto/pickerRowFill are VERBATIM string
// utilities (no devd-API coupling to strip), homeLogoBlock/repoPickerBody/
// viewLobby are beans-tui-native (a "beans" banner instead of "DevDashboard",
// beans-tui's own config.Settings.Repos instead of devd's api.Project list,
// an Offen/Gesamt bean-count metric instead of devd's Sprint/Backlog count).
//
// KEIN Overlay: viewLobby is a full top-level view (design decision, bean's
// own body text) -- a repo switch needs a complete client/watcher restart,
// which is not something a floating modal over the PREVIOUS repo's state
// should represent. handleKey's capture order (update.go) reflects this:
// m.view == viewLobby is checked as its OWN full-capture state, positioned
// AHEAD of the keys.Picker bare-match check (see that check's own doc-stamp
// for the full ERRATUM-vs-bt-0l8c-note rationale) so the Lobby's own
// repoQuery filter can accept "p" as an ordinary typed character.

import (
	"fmt"
	"strings"

	"beans-tui/internal/config"
	"beans-tui/internal/theme"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// homeLogoLines is a pure-ASCII "beans" banner (figlet "Standard"-style),
// deliberately narrow EAW-neutral glyphs only (no East-Asian-Ambiguous
// drift, same rationale as devd's own homeLogoLines doc-stamp) -- terminal-
// independent, stable width.
var homeLogoLines = []string{
	` _                           `,
	`| |__   ___  __ _ _ __  ___  `,
	`| '_ \ / _ \/ _` + "`" + ` | '_ \/ __| `,
	`| |_) |  __/ (_| | | | \__ \ `,
	`|_.__/ \___|\__,_|_| |_|___/ `,
}

// centerInto sets s exactly centered in a field of width w (left+right
// padded -> the resulting line is exactly w cells wide). ANSI-safe width via
// lipgloss.Width. Port devd view_home.go's centerInto VERBATIM.
func centerInto(s string, w int) string {
	sw := lipgloss.Width(s)
	if sw >= w {
		return s
	}
	left := (w - sw) / 2
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", w-sw-left)
}

// pickerRowFill right-aligns the right column to w cells (left is never
// truncated here -- beans-tui's own repo rows stay short enough that this
// never matters in practice, unlike devd's Project-name column). Port devd
// view_home.go's pickerRowFill VERBATIM.
func pickerRowFill(left, right string, w int) string {
	gap := w - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + right
}

// homeLogoBlock renders the Mauve "beans" banner as UNIFORM-width lines
// (each padded to the banner's own max width, so the block stays internally
// flush -- Port devd homeLogoBlock's own padding rationale). A narrow
// terminal falls back to a plain Header-styled "beans-tui" title instead of
// the full banner.
func homeLogoBlock(width int) []string {
	lw := 0
	for _, l := range homeLogoLines {
		if x := lipgloss.Width(l); x > lw {
			lw = x
		}
	}
	if width < lw+2 {
		return []string{theme.Header.Render("beans-tui")}
	}
	style := lipgloss.NewStyle().Foreground(theme.Mauve)
	out := make([]string, len(homeLogoLines))
	for i, l := range homeLogoLines {
		padded := l + strings.Repeat(" ", lw-lipgloss.Width(l))
		out[i] = style.Render(padded)
	}
	return out
}

// filteredRepos returns m.settings.Repos narrowed by a case-insensitive
// substring match against m.repoQuery -- an empty query returns the full
// list unchanged (mirrors palFilteredBeans' own "empty query" contract,
// overlay_palette.go, just without ITS additional "empty means no results"
// twist: the Lobby's whole PURPOSE is showing the configured list, so an
// empty filter must show everything, not nothing).
func (m model) filteredRepos() []string {
	q := strings.ToLower(strings.TrimSpace(m.repoQuery))
	if q == "" {
		return m.settings.Repos
	}
	var out []string
	for _, r := range m.settings.Repos {
		if strings.Contains(strings.ToLower(r), q) {
			out = append(out, r)
		}
	}
	return out
}

// repoMetricLabel renders repo's Offen/Gesamt column: "…" while
// repoMetricsMsg hasn't arrived yet (repoMetricsBatchCmd, messages.go, is
// always async -- design note on cost/latency), "err" on a per-repo load
// failure (a single broken repo must not blank the whole column, see
// repoMetric's own doc-stamp), "open/total" otherwise.
func (m model) repoMetricLabel(repo string) string {
	rm, ok := m.repoMetrics[repo]
	if !ok || !rm.loaded {
		return "…"
	}
	if rm.err != nil {
		return "err"
	}
	return fmt.Sprintf("%d/%d", rm.open, rm.total)
}

// repoPickerWidth caps the Lobby's own table width -- wide, but capped, so
// the centered block never sprawls across a very wide terminal (Port devd
// pickerBodyWidth's own rationale, view_home.go) -- but NEVER wider than
// width itself (B01, same doc-stamp as viewLobby's own -- found alongside
// it during the tmux smoke test): devd's own floor (30) assumes a terminal
// wide enough to always afford it, which does not hold once beans-tui's
// caller passes the ALREADY border-budgeted innerW (viewLobby, below) at a
// genuinely narrow terminal. Handing repoPickerBody a table width WIDER
// than the outer frame's own innerW made every row overflow it -- and
// outerBorder's lipgloss.Width()-styled Render silently WORD-WRAPS any
// line wider than its own Width() instead of erroring, turning a width bug
// into extra, unbudgeted LINES (the actual symptom that was visible live:
// a frame taller than m.height). The floor is still tried first (nicer
// layout on most terminals), but width is now the hard ceiling.
func repoPickerWidth(width int) int {
	w := width - 8
	if w > 72 {
		w = 72
	}
	if w < 30 {
		w = 30
	}
	if w > width {
		w = width
	}
	if w < 1 {
		w = 1
	}
	return w
}

// repoPickerBody renders the search line + header + filtered repo rows as
// an aligned table (Port devd projectPickerBody's own structure,
// view_home.go) -- w is the table's content width.
func (m model) repoPickerBody(w int) string {
	ti := m.repoSearch
	if m.repoQuery != "" {
		ti.TextStyle = lipgloss.NewStyle().Foreground(theme.Red)
	} else {
		ti.TextStyle = lipgloss.NewStyle().Foreground(theme.Text)
	}
	var b strings.Builder
	b.WriteString(theme.Muted.Render("⌕ ") + ti.View() + "\n\n")
	b.WriteString(pickerRowFill(theme.Dim.Render("  ◦ Repo"), theme.Dim.Render("offen/gesamt"), w) + "\n")
	b.WriteString(theme.Dim.Render(strings.Repeat("─", w)) + "\n")

	filtered := m.filteredRepos()
	for i, r := range filtered {
		cursor := "  "
		nameStyle := lipgloss.NewStyle().Foreground(theme.Text)
		if i == m.repoList.cursor {
			cursor = theme.Accent.Render("▸ ")
			nameStyle = theme.Header
		}
		metric := theme.Muted.Render(m.repoMetricLabel(r))
		nameW := w - lipgloss.Width(cursor) - lipgloss.Width(metric) - 1
		if nameW < 8 {
			nameW = 8
		}
		left := cursor + nameStyle.Render(truncate(r, nameW))
		b.WriteString(pickerRowFill(left, metric, w) + "\n")
	}
	if len(filtered) == 0 && len(m.settings.Repos) > 0 {
		b.WriteString(theme.Muted.Render("  (no matches)") + "\n")
	} else if len(m.settings.Repos) == 0 {
		b.WriteString(theme.Muted.Render("  (keine Repos in config.yaml -- ctrl+k -> settings)") + "\n")
	}
	return b.String()
}

// viewLobby renders the Lobby: banner + subtitle + repo table as ONE
// centered block (Port devd viewHome's own centering algebra, view_home.go)
// -- every line centered into the block's own content width (the widest of
// banner/table), placed via lipgloss.Place, wrapped in the app's outer
// frame. composeOverlays still runs at the end (same precedent as every
// other view -- confirmQuit must still be able to float over the Lobby,
// e.g. q/esc from a Lobby with no live repo yet).
func (m model) viewLobby() string {
	w, h := m.width, m.height
	if w <= 0 {
		w = 80
	}
	if h <= 0 {
		h = 24
	}
	// B01 (found during this task's own tmux smoke test, step (e)):
	// content must be budgeted to innerW x innerH BEFORE outerBorder wraps
	// it -- outerBorder's own doc comment (view.go) is explicit that its
	// width param is "the CONTENT width the border wraps around ... pass
	// innerW ... or the frame ends up o.Width+2 wide instead of o.Width",
	// and the SAME +2 arithmetic applies to height (RoundedBorder adds a
	// top+bottom row). The first cut of this function passed the FULL w/h
	// straight into both lipgloss.Place and outerBorder -- the rendered
	// frame ended up h+2 rows tall / w+2 cols wide, silently overflowing
	// the terminal's real m.height/m.width by 2 in each dimension. Nothing
	// caught this in unit tests (View() output is just a string, never
	// diffed against m.height/m.width there) -- only became visible live,
	// in tmux, as renderToast's top-right Toast (which DOES compose
	// directly against the real m.width/m.height, this file's own
	// overlay_show_toast.go doc-stamp) appearing pushed almost entirely
	// off the top of the screen, just its bottom border edge visible.
	// Every OTHER view (viewBrowseRepo/viewBacklog/viewReviewCockpit)
	// already gets this right via the identical innerW := w - 2 pattern --
	// this now matches that established convention exactly, no new
	// arithmetic invented.
	innerW := w - 2
	innerH := h - 2
	if innerW < 1 {
		innerW = 1
	}
	if innerH < 1 {
		innerH = 1
	}

	lines := append([]string{}, homeLogoBlock(innerW)...)
	lines = append(lines, "", theme.Muted.Render("Repo-Picker"), "")
	lines = append(lines, strings.Split(m.repoPickerBody(repoPickerWidth(innerW)), "\n")...)

	cw := 0
	for _, l := range lines {
		if x := lipgloss.Width(l); x > cw {
			cw = x
		}
	}
	if cw > innerW {
		cw = innerW
	}
	for i := range lines {
		lines[i] = centerInto(lines[i], cw)
	}
	body := strings.Join(lines, "\n")

	// truncate BEFORE centerInto (never after -- B01 doc-stamp above): the
	// hint's own reserved budget is exactly ONE line (bodyH below already
	// subtracts it), and unlike the table rows (already width-controlled
	// via repoPickerWidth's own fix), this literal string has no natural
	// truncation point of its own -- a narrow terminal must cut it, not
	// silently word-wrap it into a second, unbudgeted line.
	hintText := truncate("i/k:↑↓  enter:open  type:filter  esc/q:"+lobbyBackHint(m), innerW)
	hint := theme.Muted.Render(centerInto(hintText, innerW))
	bodyH := innerH - 1 // innerH total inner lines, minus the 1 reserved for hint below
	if bodyH < 1 {
		bodyH = 1
	}
	placed := lipgloss.Place(innerW, bodyH, lipgloss.Center, lipgloss.Center, body)
	out := outerBorder(placed+"\n"+hint, innerW, true)

	return m.composeOverlays(out, w, h)
}

// lobbyBackHint disambiguates the Lobby's own esc/q footer hint between its
// two entry paths (keyLobby's own doc-stamp has the full rationale): "back"
// when a live repo already exists to return to (p from a running session),
// "quit" when the Lobby IS the very first screen (design decision d, no
// repo resolved yet).
func lobbyBackHint(m model) string {
	if m.client != nil {
		return "back"
	}
	return "quit"
}

// openLobby transitions into the Lobby (E5 Task 6, design decision h):
// reachable both as the Lobby's OWN startup entry (app.go Run/Init, design
// decision d) and via keys.Picker from any live view (update.go handleKey).
// Reloads Settings from disk FIRST (bean bt-zhwl's own acceptance wording:
// "lädt Settings.Repos neu, falls seit Start geändert") -- Run() only ever
// loads config.yaml ONCE at process start, so without this reload here, a
// PO who edits config.yaml's repos list (by hand, or via the Settings-Form,
// E5 Task 5) mid-session would keep seeing the STALE list every time they
// reopen the Lobby for the rest of the process lifetime (caught live during
// this task's own tmux smoke test, step (e) -- not a hypothetical). Mirrors
// Run()'s own "corrupt/missing config.yaml -> Defaults, never crash"
// fallback (app.go) rather than introducing a second, divergent error
// policy. Resets the filter/cursor (mirrors devd's openProjPick/keyHome-
// into-fresh-state precedent) and (re)dispatches the async metrics batch --
// a stale metrics map entry for a since-removed repo is harmless
// (repoPickerBody only ever reads entries for repos CURRENTLY in
// filteredRepos()) and a newly added repo simply shows "…" until its own
// repoMetricsMsg arrives.
func (m model) openLobby() (tea.Model, tea.Cmd) {
	m.view = viewLobby
	if fresh, err := config.LoadSettings(); err == nil {
		m.settings = fresh
	} else {
		m.settings = config.DefaultSettings()
	}
	m.repoQuery = ""
	m.repoSearch.SetValue("")
	m.repoSearch.Focus() // Port devd openProjPick/keyHome's own "always focused" Lobby-search precedent
	m.repoList = listState{}
	m.repoList.setLen(len(m.settings.Repos))
	return m, repoMetricsBatchCmd(m.settings.Repos)
}

// keyLobby drives the Lobby (Port devd keyHome's own structure,
// view_home.go): nav + selection dispatches switchRepoCmd (the watcher-
// lifecycle switch, messages.go), esc/q returns to the live repo if one
// already exists or quit-confirms otherwise (US-01 parity doc-stamp,
// lobbyBackHint above), any other key filters repoQuery live.
//
// notify is wired here (not messages.go) as
// func(){ activeProgram.Send(watchMsg{}) } -- activeProgram (app.go) is
// guaranteed non-nil by the time a keypress can reach this handler (Run()
// sets it immediately after tea.NewProgram, strictly before p.Run() is ever
// called, and no key can be dispatched before that).
func (m model) keyLobby(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch navKey(msg.String()) {
	case "up":
		m.repoList.move(-1)
		return m, nil
	case "down":
		m.repoList.setLen(len(m.filteredRepos()))
		m.repoList.move(1)
		return m, nil
	}
	switch msg.String() {
	case "enter":
		filtered := m.filteredRepos()
		if m.repoList.cursor < 0 || m.repoList.cursor >= len(filtered) {
			return m, nil
		}
		target := filtered[m.repoList.cursor]
		notify := func() { activeProgram.Send(watchMsg{}) }
		return m, switchRepoCmd(m.watchStop, target, notify)
	case "esc", "q", "ctrl+c":
		if m.client != nil {
			// A live repo already exists (p from a running session) -- go
			// back to it instead of force-quitting a working session
			// (design decision, bean bt-zhwl Step 9: devd's own keyHome has
			// no equivalent case since devd's Home is its ONLY entry point;
			// beans-tui's Lobby is reachable a SECOND way).
			m.view = viewBrowseRepo
			return m, nil
		}
		return m.requestQuit()
	}

	// Port devd projFilterType's own Focus()-before-Update() precedent
	// (view_home.go): bubbles' textinput.Model.Update ignores key input
	// entirely while unfocused, and the Lobby -- unlike the Tree's `/`
	// search -- has no separate "not yet typing" mode (every non-nav/
	// non-enter/non-esc key IS the filter, mirrors devd's own keyHome/
	// projFilterType structure). Idempotent, safe to call every keystroke.
	m.repoSearch.Focus()
	var cmd tea.Cmd
	m.repoSearch, cmd = m.repoSearch.Update(msg)
	prev := m.repoQuery
	m.repoQuery = strings.TrimSpace(m.repoSearch.Value())
	if m.repoQuery != prev {
		m.repoList.cursor = 0
		m.repoList.setLen(len(m.filteredRepos()))
	}
	return m, cmd
}
