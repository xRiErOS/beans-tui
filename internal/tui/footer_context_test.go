package tui

// footer_context_test.go — TDD coverage for PF-11's context-sensitive
// footer (Q04-Antwort, design-spec.md §15, epic-E7-plan.md Task 7 Step 6,
// bean bt-m6at): footer_context.go's contextualLocalHint swaps the base
// view's own footer to the ACTIVE capture state's bindings the instant one
// fully captures input (handleKey's own precedent, update.go) -- the
// view-local bindings underneath are then non-functional and would mislead.

import (
	"strings"
	"testing"

	keybind "github.com/charmbracelet/bubbles/key"
)

// viewLocalStub is a deliberately distinct sentinel binding list --
// contextualLocalHint's tests assert EITHER this sentinel shows (default
// case) OR it does NOT show (every context-capture case), so a bug that
// falls through to the wrong branch is caught either way.
func viewLocalStub() []keybind.Binding {
	return []keybind.Binding{keybind.NewBinding(keybind.WithKeys("z"), keybind.WithHelp("z", "stub-view-local"))}
}

func TestContextualLocalHintDefaultUsesViewLocal(t *testing.T) {
	m := model{}
	got := m.contextualLocalHint(viewLocalStub())
	got = stripHint(got)
	if !strings.Contains(got, "z stub-view-local") {
		t.Errorf("default state: contextualLocalHint = %q, want it to contain the viewLocal stub hint", got)
	}
}

func TestContextualLocalHintFilterOpen(t *testing.T) {
	m := model{filterOpen: true}
	got := m.contextualLocalHint(viewLocalStub())
	got = stripHint(got)
	if strings.Contains(got, "stub-view-local") {
		t.Errorf("filterOpen: contextualLocalHint = %q, must NOT contain the (irrelevant) viewLocal stub", got)
	}
	if !strings.Contains(got, "space/x Toggle facet") {
		t.Errorf("filterOpen: contextualLocalHint = %q, want it to contain the Toggle hint (Q04)", got)
	}
}

func TestContextualLocalHintOverlayValueMenu(t *testing.T) {
	m := model{overlay: overlayValueMenu}
	got := m.contextualLocalHint(viewLocalStub())
	got = stripHint(got)
	if strings.Contains(got, "stub-view-local") {
		t.Errorf("overlayValueMenu: contextualLocalHint = %q, must not contain the viewLocal stub", got)
	}
	if !strings.Contains(got, "s Status") {
		t.Errorf("overlayValueMenu: contextualLocalHint = %q, want it to contain keys.Status (real close-alias, keyValueMenu)", got)
	}
}

// TestContextualLocalHintOverlayTagPickerShowsToggle guards a deliberate
// DEVIATION from epic-E7-plan.md Task 7 Step 6's literal text (which lumps
// Tag-/Parent-/Blocking-Picker into one Toggle-free {Up,Down,Enter,Back}
// set): keyTagPicker (box_picker_tag.go) really wires keys.Toggle
// (space/x, multi-select tag membership) -- omitting it here would hide a
// real, working key AND leave Q04's own general wording ("wenn ein
// Form/Overlay aktiv ist ... inkl. space: select/toggle") only half
// addressed. See this task's Deviations section / footer_context.go.
func TestContextualLocalHintOverlayTagPickerShowsToggle(t *testing.T) {
	m := model{overlay: overlayTagPicker}
	got := m.contextualLocalHint(viewLocalStub())
	got = stripHint(got)
	if !strings.Contains(got, "space/x Toggle facet") {
		t.Errorf("overlayTagPicker: contextualLocalHint = %q, want the Toggle hint", got)
	}
}

// TestTagPickerLocalBindingsIncludesNewTag guards B14 (design-spec.md §15
// PF-16, bean bt-ntoz, E8 Task 7, bean bt-yqdy): the Tag-Picker's free-text
// new-tag sub-mode (`n`, box_picker_tag.go's keyTagPicker) was NOT broken --
// only undiscoverable, since the OUTER Footer Zone 3 (this file) never
// listed it, even though the picker's OWN inline hint line already did
// (tagPickerBox's doc comment). tagPickerLocalBindings() must now include
// keys.NewTag alongside its existing Up/Down/Toggle/Enter/Back set.
func TestTagPickerLocalBindingsIncludesNewTag(t *testing.T) {
	found := false
	for _, b := range tagPickerLocalBindings() {
		if strings.Join(b.Keys(), ",") == strings.Join(keys.NewTag.Keys(), ",") {
			found = true
		}
	}
	if !found {
		t.Fatalf("tagPickerLocalBindings() = %v, want it to include keys.NewTag %v", tagPickerLocalBindings(), keys.NewTag.Keys())
	}
}

// TestContextualLocalHintOverlayTagPickerShowsNewTag is the same guard at
// the Footer Zone 3 rendering layer (contextualLocalHint) -- the PO-facing
// surface B14 actually fixes ("n" was invisible in the outer footer while
// the Tag-Picker was open).
func TestContextualLocalHintOverlayTagPickerShowsNewTag(t *testing.T) {
	m := model{overlay: overlayTagPicker}
	got := m.contextualLocalHint(viewLocalStub())
	got = stripHint(got)
	if !strings.Contains(got, "n New tag") {
		t.Errorf("overlayTagPicker: contextualLocalHint = %q, want the New-Tag hint (%q)", got, "n New tag")
	}
}

func TestContextualLocalHintOverlayParentPickerOmitsToggle(t *testing.T) {
	m := model{overlay: overlayParentPicker}
	got := m.contextualLocalHint(viewLocalStub())
	got = stripHint(got)
	if strings.Contains(got, "Toggle") {
		t.Errorf("overlayParentPicker: contextualLocalHint = %q, must NOT contain Toggle (keyParentPicker is single-select, no space/x case)", got)
	}
}

// TestContextualLocalHintOverlayBlockingPickerShowsToggle mirrors the
// Tag-Picker deviation above -- keyBlockingPicker (box_picker_blocking.go)
// also wires keys.Toggle (multi-select blocking-relation membership).
func TestContextualLocalHintOverlayBlockingPickerShowsToggle(t *testing.T) {
	m := model{overlay: overlayBlockingPicker}
	got := m.contextualLocalHint(viewLocalStub())
	got = stripHint(got)
	if !strings.Contains(got, "space/x Toggle facet") {
		t.Errorf("overlayBlockingPicker: contextualLocalHint = %q, want the Toggle hint", got)
	}
}

// TestContextualLocalHintOverlayCreateConfirm/DeleteConfirm guard a plan
// GAP-FILL: epic-E7-plan.md Task 7 Step 6 only names Value-Menu and
// Tag-/Parent-/Blocking-Picker's overlay-specific sets -- overlayCreateConfirm/
// overlayDeleteConfirm are two more `m.overlay != overlayNone` cases the
// plan's own switch-priority text never assigns a set to. Both confirm
// gates (box_confirm_create.go/box_confirm_delete.go) really only answer to
// Enter/Back, so that is the fallback here too.
//
// D05 VERIFICATION (design-spec.md §15 PF-16, bean bt-ntoz, cited in
// bt-d8kc's own PO-Wortlaut): "Overlay-Footer zeigen enter/esc" -- these two
// tests are the concrete Sign-off evidence that D04's Header-Global-Kürzung
// (which degrades keys.Enter/keys.Back out of globalBindings()) is a No-Op
// for the Create-/Delete-Confirm-Gate's OWN footer: confirmGateLocalBindings
// (footer_context.go) was never derived FROM globalBindings() -- it is its
// own independent {Enter, Back} literal -- so removing Enter/Back from the
// header cannot silently remove them here too. Only the render OPTIC
// changed (no more ":" -- D06), never the binding SET.
func TestContextualLocalHintOverlayCreateConfirm(t *testing.T) {
	m := model{overlay: overlayCreateConfirm}
	got := m.contextualLocalHint(viewLocalStub())
	got = stripHint(got)
	if strings.Contains(got, "stub-view-local") {
		t.Errorf("overlayCreateConfirm: contextualLocalHint = %q, must not contain the viewLocal stub", got)
	}
	if !strings.Contains(got, "enter open/confirm") || !strings.Contains(got, "esc back") {
		t.Errorf("overlayCreateConfirm: contextualLocalHint = %q, want enter+back", got)
	}
}

func TestContextualLocalHintOverlayDeleteConfirm(t *testing.T) {
	m := model{overlay: overlayDeleteConfirm}
	got := m.contextualLocalHint(viewLocalStub())
	got = stripHint(got)
	if strings.Contains(got, "stub-view-local") {
		t.Errorf("overlayDeleteConfirm: contextualLocalHint = %q, must not contain the viewLocal stub", got)
	}
	if !strings.Contains(got, "enter open/confirm") || !strings.Contains(got, "esc back") {
		t.Errorf("overlayDeleteConfirm: contextualLocalHint = %q, want enter+back", got)
	}
}

func TestContextualLocalHintSearchActive(t *testing.T) {
	m := model{searchActive: true}
	got := m.contextualLocalHint(viewLocalStub())
	got = stripHint(got)
	if strings.Contains(got, "stub-view-local") {
		t.Errorf("searchActive: contextualLocalHint = %q, must not contain the viewLocal stub", got)
	}
}

func TestContextualLocalHintPaletteOpen(t *testing.T) {
	m := model{paletteOpen: true}
	got := m.contextualLocalHint(viewLocalStub())
	got = stripHint(got)
	if strings.Contains(got, "stub-view-local") {
		t.Errorf("paletteOpen: contextualLocalHint = %q, must not contain the viewLocal stub", got)
	}
}

func TestContextualLocalHintHelpOpen(t *testing.T) {
	m := model{helpOpen: true}
	got := m.contextualLocalHint(viewLocalStub())
	got = stripHint(got)
	if got != "esc back" {
		t.Errorf("helpOpen: contextualLocalHint = %q, want exactly %q", got, "esc back")
	}
}

// TestContextualLocalHintPriority guards the priority chain design-spec.md
// §15 PF-11 spells out (mirrors handleKey's own full-capture dispatch
// order, update.go): Filter-Menü > Overlay > Suche > Palette > Help >
// View-Normalzustand. Only adjacent pairs whose rendered hints actually
// DIFFER are usable as a discriminator here (searchActive and paletteOpen
// both render the identical {Enter,Back} pair by design, so that specific
// adjacency isn't separately re-verifiable via string content -- covered
// instead by the individual single-state tests above).
func TestContextualLocalHintPriority(t *testing.T) {
	cases := []struct {
		name string
		m    model
		want string // a substring only the winning branch's hint contains
	}{
		{"filter beats overlay", model{filterOpen: true, overlay: overlayValueMenu}, "X Clear filters"},
		{"overlay beats search", model{overlay: overlayValueMenu, searchActive: true}, "s Status"},
		{"palette beats help", model{paletteOpen: true, helpOpen: true}, "enter open/confirm"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := stripHint(c.m.contextualLocalHint(viewLocalStub()))
			if !strings.Contains(got, c.want) {
				t.Errorf("%s: contextualLocalHint = %q, want it to contain %q", c.name, got, c.want)
			}
		})
	}
}
