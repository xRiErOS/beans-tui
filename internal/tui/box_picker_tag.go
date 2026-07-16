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

// tagCount is one usage-counted row of the tag picker's menu. defined marks
// a row whose tag is registered in the repo-local Tag-Registry (T6, bean
// bt-pqq3, epic bt-362n D10 Suggest-Mode) -- as opposed to a "free" tag that
// merely sits on some bean but was never defined there. Suggest-Mode is
// display/sort-only: a free row stays exactly as togglable/savable as
// before (kein strict mode, PO-Vorgabe wörtlich).
type tagCount struct {
	tag     string
	count   int
	defined bool
}

// collectTagCounts counts each tag's usage across every bean in idx and
// marks every tag present in defined as a registry-defined row (T6, D10).
// The returned set is the UNION of every tag currently in use AND every
// registry-defined tag -- a defined tag with ZERO current usage still
// appears (count 0, defined=true), mirroring D09's "auch mit Count 0"
// wording one layer up (view_tag_management.go's tagRegistryRows): the
// Suggest-Mode picker offers every defined tag as a candidate even before it
// has ever been applied to a bean (this task's own tmux-Smoke acceptance
// wording "unbenutzt-definierte sichtbar"). defined may be nil (e.g. no
// client yet, or a missing/corrupt registry file, D02 tolerant-missing) --
// a nil map read always returns false, so every row is simply "free" in
// that case, identical to this function's pre-D10 behavior.
//
// Sort (sortTagCountsDefinedFirst): "defined" is the NEW PRIMARY key (D10 --
// defined tags always sort before free tags, regardless of count); the
// pre-existing count-desc/alpha tie-break stays SECONDARY/TERTIARY,
// unchanged WITHIN each of the two groups.
//
// Determinism note (the same ERRATUM lesson box_filter_facets.go's
// tagFilterOptions documents, but resolved more simply here): idx.ByID is a
// Go map with no defined iteration order, so the intermediate counts map
// built by ranging over it is itself order-dependent -- but the FINAL sort
// key (defined, then count desc, then tag name alpha) fully determines row
// order on its own, because tag names are unique within the counts map (no
// two rows can tie on all three). Unlike tagFilterOptions (which sorts whole
// *Bean values that CAN tie completely on every sort field), no baseline
// pre-sort is needed here -- the map-walk order that built the counts never
// leaks into the returned slice's order.
func collectTagCounts(idx *data.Index, defined map[string]bool) []tagCount {
	counts := map[string]int{}
	if idx != nil {
		for _, b := range idx.ByID {
			for _, t := range b.Tags {
				if t == "" {
					continue
				}
				counts[t]++
			}
		}
	}

	out := make([]tagCount, 0, len(counts)+len(defined))
	seen := make(map[string]bool, len(counts))
	for tag, n := range counts {
		out = append(out, tagCount{tag: tag, count: n, defined: defined[tag]})
		seen[tag] = true
	}
	for tag := range defined {
		if seen[tag] {
			continue
		}
		out = append(out, tagCount{tag: tag, count: 0, defined: true})
	}
	sortTagCountsDefinedFirst(out)
	return out
}

// sortTagCountsDefinedFirst sorts items in place: defined rows first (D10
// NEW primary key), then count descending, then tag name alpha -- the SAME
// comparator collectTagCounts and keyTagInput's new-tag insert path (both
// build/maintain a []tagCount slice that must stay consistently ordered)
// share, so the two call sites can never drift apart on what "sorted" means.
func sortTagCountsDefinedFirst(items []tagCount) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].defined != items[j].defined {
			return items[i].defined
		}
		if items[i].count != items[j].count {
			return items[i].count > items[j].count
		}
		return items[i].tag < items[j].tag
	})
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
//
// D03/D10 (T6, bean bt-pqq3, epic bt-362n): the Tag-Registry is loaded
// FRESH from disk on EVERY open (mirrors openTagManagementPage's own
// "reload from disk on every open" convention, view_tag_management.go) -- a
// nil m.client (pre-load/test fixture) degrades to an empty registry rather
// than panicking, same D02 tolerant-missing philosophy one layer up.
// tagItems is then built via collectTagCounts' Suggest-Mode union: defined
// tags sort first, but every free tag stays fully togglable regardless
// (kein strict mode, PO-Vorgabe wörtlich).
func (m model) openTagPicker() model {
	b := m.focusedBean()
	if b == nil {
		return m
	}
	m.mutTarget = b.ID

	var defs []string
	if m.client != nil {
		defs, _ = m.client.LoadTagDefs() // D02: LoadTagDefs never returns a non-nil error
	}
	defined := make(map[string]bool, len(defs))
	for _, name := range defs {
		defined[name] = true
	}
	m.tagItems = collectTagCounts(m.idx, defined)

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
			// defined: false (zero value) -- B14's free-text new-tag path stays
			// UNCHANGED (Q03 in the epic body, bt-362n: NOT part of this task,
			// no Registry-write from here).
			items = append(items, tagCount{tag: name, count: 0})
			sortTagCountsDefinedFirst(items)
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
// the free-text capture prompt when it is open. Each row's marker column
// (T6, bean bt-pqq3, D10 Suggest-Mode, PF-12 design-spec.md §15) is ALWAYS
// reserved -- a defined tag renders tagManagementMarkerGlyph (reused
// verbatim from view_tag_management.go/T2, bean bt-r92i's own explicit
// "T6 kann diesen Glyph ... wiederverwenden" hand-off -- PF-12-Konsistenz,
// one "defined" glyph across the whole app, instead of a second bespoke
// glyph), a free tag renders the SAME-WIDTH blank -- NEVER a conditional
// omission of the column itself, so neither group's checkbox/name column
// ever shifts relative to the other.
func (m model) tagPickerBox() string {
	if m.tagInputActive {
		return m.tagInputBox()
	}

	var b strings.Builder
	b.WriteString(theme.Muted.Render("space/x:toggle  n:new tag  enter:save  esc:discard") + "\n")
	for i, it := range m.tagItems {
		marker := " "
		if it.defined {
			marker = tagManagementMarkerStyle.Render(tagManagementMarkerGlyph)
		}
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
		b.WriteString(cursor + marker + " " + box + " " + label + count + "\n")
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
