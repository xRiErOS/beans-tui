package tui

// box_picker_tag.go — the Tag-Picker (`t`, E3 Task 2, bean bt-8v69):
// usage-counted toggle-multi-select over every tag currently in use, plus a
// free-text new-tag entry sub-mode. Port pattern (design decision, plan
// »Task 2«): counter + sort semantics from beans-src tagpicker.go:64-96
// (sort count desc, then alpha); the Pending-Diff/Enter-confirms/Esc-
// discards semantics from beans-src blockingpicker.go:197-245 (space
// toggles pending, enter diffs pending against original, esc discards) --
// this is the FIRST beans-tui overlay carrying real mutation state across
// multiple keystrokes before a single confirm, so it is deliberately NOT
// box_filter_facets.go's keyFilterMenu pattern (whose enter only closes,
// since facets act live and there is nothing to "confirm"). The free-text
// capture sub-mode mirrors keySearchInput's "one persistent textinput.Model,
// reset+focused on open, every key but enter/esc belongs to the input"
// convention (update.go:520-549).
//
// ERRATUM vs. the plan's own EARLIER Step-1 test sketch ("+1 Tag, -1 Tag ->
// 2 Mutationen (tea.Batch)"): superseded by the plan's own "Design-Nachtrag"
// immediately below it (verbindlich für die Implementierung) -- N sequential
// AddTag/RemoveTag mutateCmds against ONE etag would be a conflict cascade
// (the first call rotates the etag on disk, every subsequent call then sees
// a stale etag and fails ErrConflict). applyTagPickerDiff below fires
// exactly ONE mutateCmd wrapping data.SetTags (mutations.go), which builds
// ONE `beans update` invocation carrying every added/removed tag as
// repeated --tag/--remove-tag flags -- one etag, no cascade.

import (
	"fmt"
	"sort"
	"strings"

	"beans-tui/internal/data"
	"beans-tui/internal/theme"
	keybind "github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// tagCount is one usage-counted row of the tag picker's menu.
type tagCount struct {
	tag   string
	count int
}

// collectTagCounts counts each tag's usage across every bean in idx.
//
// Determinism note (the same ERRATUM lesson box_filter_facets.go's
// tagFilterOptions documents, but resolved more simply here): idx.ByID is a
// Go map with no defined iteration order, so the intermediate counts map
// built by ranging over it is itself order-dependent -- but the FINAL sort
// key (count desc, then tag name alpha) fully determines row order on its
// own, because tag names are unique within the counts map (no two rows can
// tie on BOTH count and name). Unlike tagFilterOptions (which sorts whole
// *Bean values that CAN tie completely on every sort field), no baseline
// pre-sort is needed here -- the map-walk order that built the counts never
// leaks into the returned slice's order.
func collectTagCounts(idx *data.Index) []tagCount {
	if idx == nil {
		return nil
	}
	counts := map[string]int{}
	for _, b := range idx.ByID {
		for _, t := range b.Tags {
			if t == "" {
				continue
			}
			counts[t]++
		}
	}
	out := make([]tagCount, 0, len(counts))
	for tag, n := range counts {
		out = append(out, tagCount{tag: tag, count: n})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].count != out[j].count {
			return out[i].count > out[j].count
		}
		return out[i].tag < out[j].tag
	})
	return out
}

// tagItemIndex returns items' index for tag, or -1 if not present -- shared
// by the toggle handler and the new-tag submit's dedupe check.
func tagItemIndex(items []tagCount, tag string) int {
	for i, it := range items {
		if it.tag == tag {
			return i
		}
	}
	return -1
}

// openTagPicker opens the Tag-Picker on the focused bean (the
// focusedBean()!=nil guard lives in the caller, keyNodeAction -- same
// convention as openValueMenu). tagOriginal/tagPending are seeded as two
// INDEPENDENT maps (design decision, plan »Task 2«) from the focused bean's
// CURRENT tags -- wholesale-replace at open time, mirroring searchBleveIDs'
// "always a brand-new map value" convention (types.go doc-stamp). mutTarget
// captures the bean ID only, never the etag (design decision d) -- a
// watch-reload between open and the eventual enter is honored automatically
// via beanETag.
func (m model) openTagPicker() model {
	b := m.focusedBean()
	if b == nil {
		return m
	}
	m.mutTarget = b.ID
	m.tagItems = collectTagCounts(m.idx)

	orig := make(map[string]bool, len(b.Tags))
	for _, t := range b.Tags {
		orig[t] = true
	}
	pending := make(map[string]bool, len(orig))
	for t := range orig {
		pending[t] = true
	}
	m.tagOriginal = orig
	m.tagPending = pending

	m.menu = listState{}
	m.menu.setLen(len(m.tagItems))

	m.tagInputActive = false
	m.tagInputErr = ""
	m.tagInput.SetValue("")

	m.overlay = overlayTagPicker
	return m
}

// keyTagPicker drives the open Tag-Picker. The free-text new-tag sub-mode
// (m.tagInputActive) is checked FIRST and fully captures every key except
// enter/esc (same precedent as keySearchInput, update.go) -- otherwise
// typing e.g. "x" into a new tag name would toggle the cursored row instead
// of appearing in the input. Outside that sub-mode: up/down move the cursor
// (navKey), space/x (keys.Toggle) toggles the cursored row's pending state,
// "n" opens the new-tag input, enter diffs pending against original (see
// applyTagPickerDiff), esc discards everything and closes.
func (m model) keyTagPicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.tagInputActive {
		return m.keyTagInput(msg)
	}

	switch navKey(msg.String()) {
	case "up":
		m.menu.move(-1)
		return m, nil
	case "down":
		m.menu.move(1)
		return m, nil
	}

	switch {
	case keybind.Matches(msg, keys.Back):
		m.overlay = overlayNone
		return m, nil
	case keybind.Matches(msg, keys.Toggle):
		return m.toggleTagPending(), nil
	case keybind.Matches(msg, keys.NewTag):
		return m.openTagInput()
	case keybind.Matches(msg, keys.Enter):
		return m.applyTagPickerDiff()
	}
	return m, nil
}

// toggleTagPending flips the cursored row's membership in m.tagPending.
// I01 (types.go doc-stamp, bean bt-7jr8): clones via cloneBoolMap before
// writing -- tagPending is COPY-ON-WRITE for every toggle during the
// picker's (potentially multi-keystroke) open session, same convention as
// toggleFacet (box_filter_facets.go), even though the map itself is
// wholesale-replaced fresh on each OPEN (openTagPicker's doc-stamp).
func (m model) toggleTagPending() model {
	if m.menu.cursor < 0 || m.menu.cursor >= len(m.tagItems) {
		return m
	}
	tag := m.tagItems[m.menu.cursor].tag
	out := cloneBoolMap(m.tagPending)
	if out[tag] {
		delete(out, tag)
	} else {
		out[tag] = true
	}
	m.tagPending = out
	return m
}

// openTagInput enters the free-text new-tag capture sub-mode (port
// openSearchInput's convention: reset value, focus, blink).
func (m model) openTagInput() (tea.Model, tea.Cmd) {
	m.tagInputActive = true
	m.tagInputErr = ""
	m.tagInput.SetValue("")
	m.tagInput.Focus()
	return m, textinput.Blink
}

// keyTagInput drives the free-text new-tag input. enter validates against
// data.ValidTagName: invalid -> tagInputErr is set and the input STAYS open
// for a retry (no submit, plan »Task 2« Step 1: "invalider Name -> Inline-
// Fehler, kein Submit"); valid -> the tag is added to tagItems (deduping
// against an existing row, e.g. a name the picker's own count list already
// carries) as pending=true, and the input closes. esc closes the input
// WITHOUT touching the outer picker's pending/original state (only the
// input sub-mode itself is discarded).
func (m model) keyTagInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.tagInputActive = false
		m.tagInput.Blur()
		m.tagInputErr = ""
		return m, nil
	case tea.KeyEnter:
		name := strings.TrimSpace(m.tagInput.Value())
		if !data.ValidTagName(name) {
			m.tagInputErr = "invalid tag name (a-z0-9, hyphen-separated, lowercase)"
			return m, nil
		}
		m.tagInputErr = ""
		m.tagInputActive = false
		m.tagInput.Blur()
		m.tagInput.SetValue("")

		if tagItemIndex(m.tagItems, name) < 0 {
			items := append([]tagCount(nil), m.tagItems...)
			items = append(items, tagCount{tag: name, count: 0})
			sort.Slice(items, func(i, j int) bool {
				if items[i].count != items[j].count {
					return items[i].count > items[j].count
				}
				return items[i].tag < items[j].tag
			})
			m.tagItems = items
			m.menu.setLen(len(m.tagItems))
		}
		out := cloneBoolMap(m.tagPending)
		out[name] = true
		m.tagPending = out
		if idx := tagItemIndex(m.tagItems, name); idx >= 0 {
			m.menu.cursor = idx
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.tagInput, cmd = m.tagInput.Update(msg)
	return m, cmd
}

// applyTagPickerDiff computes tagPending's diff against tagOriginal and
// fires it as ONE combined data.SetTags mutateCmd (design decision, see
// this file's doc comment) against a FRESH etag read from the live index
// (m.beanETag, design decision d -- never a captured copy). No pending
// changes at all -> no Cmd (mirrors applyValueMenuSelection's vanished-
// target guard: ok==false means mutTarget disappeared from the index
// between open and this enter -- the overlay still closes, with a
// status-line note, but no doomed mutation fires). add/remove are sorted
// before dispatch so the underlying CLI invocation's flag order (and thus
// any test/log asserting on it) is deterministic despite the map-iteration
// order the diff loops below walk in.
func (m model) applyTagPickerDiff() (tea.Model, tea.Cmd) {
	m.overlay = overlayNone
	id := m.mutTarget

	var add, remove []string
	for tag := range m.tagPending {
		if !m.tagOriginal[tag] {
			add = append(add, tag)
		}
	}
	for tag := range m.tagOriginal {
		if !m.tagPending[tag] {
			remove = append(remove, tag)
		}
	}
	if len(add) == 0 && len(remove) == 0 {
		return m, nil
	}
	sort.Strings(add)
	sort.Strings(remove)

	etag, ok := m.beanETag(id)
	if !ok {
		m.err = "Bean no longer exists — selection discarded"
		return m, nil
	}
	client := m.client
	return m, mutateCmd(func() error { return client.SetTags(id, add, remove, etag) })
}

// tagPickerBox renders the floating Tag-Picker overlay -- the checkbox +
// usage-count row list (same modalPanel + [x]/[ ] convention as
// treeFilterBox, box_filter_facets.go) when the new-tag input is closed, or
// the free-text capture prompt when it is open.
func (m model) tagPickerBox() string {
	if m.tagInputActive {
		return m.tagInputBox()
	}

	var b strings.Builder
	b.WriteString(theme.Muted.Render("space/x:toggle  n:new tag  enter:save  esc:discard") + "\n")
	for i, it := range m.tagItems {
		box := theme.Dim.Render("[ ]")
		if m.tagPending[it.tag] {
			box = theme.Accent.Render("[x]")
		}
		cursor := "  "
		label := it.tag
		if i == m.menu.cursor {
			cursor = theme.Accent.Render("▸ ")
			label = theme.Header.Render(label)
		}
		count := theme.Muted.Render(fmt.Sprintf(" (%d)", it.count))
		b.WriteString(cursor + box + " " + label + count + "\n")
	}
	if len(m.tagItems) == 0 {
		b.WriteString(theme.Muted.Render("(no tags in repo)") + "\n")
	}
	return modalPanel("Tags", b.String(), "", clampModalWidth(40, m.width), theme.Mauve)
}

// tagInputBox renders the free-text new-tag capture prompt.
func (m model) tagInputBox() string {
	var b strings.Builder
	b.WriteString(theme.Muted.Render("enter:create  esc:cancel") + "\n\n")
	b.WriteString(m.tagInput.View() + "\n")
	if m.tagInputErr != "" {
		b.WriteString("\n" + lipgloss.NewStyle().Foreground(theme.Red).Render(m.tagInputErr) + "\n")
	}
	return modalPanel("New tag", b.String(), "", clampModalWidth(40, m.width), theme.Mauve)
}
