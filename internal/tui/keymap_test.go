package tui

import (
	"reflect"
	"strings"
	"testing"

	keybind "github.com/charmbracelet/bubbles/key"
)

// TestKeymapNoCtrlSQ guards design-spec.md §7: "Keine ctrl+s/ctrl+q-Belegung
// (XOFF/XON)". Binding either would shadow the terminal's flow-control keys —
// on many terminal/tmux configs the sequence never reaches the app at all.
// Iterates every keybind.Binding field of keyMap via reflection so a future
// added binding is covered automatically, not just the fields known today.
func TestKeymapNoCtrlSQ(t *testing.T) {
	forbidden := map[string]bool{"ctrl+s": true, "ctrl+q": true}
	v := reflect.ValueOf(newKeyMap())
	typ := v.Type()
	found := 0
	for i := 0; i < v.NumField(); i++ {
		b, ok := v.Field(i).Interface().(keybind.Binding)
		if !ok {
			continue
		}
		found++
		for _, k := range b.Keys() {
			if forbidden[k] {
				t.Errorf("field %s binds forbidden key %q (XOFF/XON conflict)", typ.Field(i).Name, k)
			}
		}
	}
	if found == 0 {
		t.Fatal("no keybind.Binding fields found on keyMap — reflection scan is broken")
	}
}

// TestArrowsAlwaysAlias guards design-spec.md §7: "↑/i ↓/k ←/j →/l (Pfeile
// immer Alias)" — every direction binding must include its arrow key
// alongside the jkli letter, and navKey must normalize both to the same
// canonical direction.
func TestArrowsAlwaysAlias(t *testing.T) {
	dirs := []struct {
		name  string
		b     keybind.Binding
		arrow string
		want  string
	}{
		{"Up", keys.Up, "up", "up"},
		{"Down", keys.Down, "down", "down"},
		{"Left", keys.Left, "left", "left"},
		{"Right", keys.Right, "right", "right"},
	}
	for _, d := range dirs {
		if !bindHas(d.b, d.arrow) {
			t.Errorf("%s binding %v does not include arrow key %q", d.name, d.b.Keys(), d.arrow)
		}
		if got := navKey(d.arrow); got != d.want {
			t.Errorf("navKey(%q)=%q, want %q (%s)", d.arrow, got, d.want, d.name)
		}
	}
}

// TestFocusInFocusOutKeysBound guards PF-13 (design-spec.md §15, E7 T6):
// FocusIn binds "tab" (backward-compat, existing bidirectional toggle),
// FocusOut binds the NEW "shift+tab" (deterministic one-way exit) -- pins
// keymap.go's own single source so a future re-binding accident surfaces
// here instead of only as a runtime behavior regression.
func TestFocusInFocusOutKeysBound(t *testing.T) {
	k := newKeyMap()
	if !bindHas(k.FocusIn, "tab") {
		t.Errorf("FocusIn.Keys() = %v, want to contain %q", k.FocusIn.Keys(), "tab")
	}
	if !bindHas(k.FocusOut, "shift+tab") {
		t.Errorf("FocusOut.Keys() = %v, want to contain %q", k.FocusOut.Keys(), "shift+tab")
	}
}

// TestNewTagKeyBound guards B14 (design-spec.md §15 PF-16, bean bt-ntoz, E8
// Task 7, bean bt-yqdy): keys.NewTag is a real, typed keybind.Binding for the
// Tag-Picker's free-text new-tag sub-mode (box_picker_tag.go's `n`) -- a raw
// msg.String()=="n" comparison can drive keyTagPicker, but cannot be
// rendered by renderBindings() (view.go), which is what the Footer-Hint fix
// needs. Pins both the key AND the Help() text the footer will display.
func TestNewTagKeyBound(t *testing.T) {
	k := newKeyMap()
	if !bindHas(k.NewTag, "n") {
		t.Errorf("NewTag.Keys() = %v, want to contain %q", k.NewTag.Keys(), "n")
	}
	if got := k.NewTag.Help().Desc; got != "New tag" {
		t.Errorf("NewTag.Help().Desc = %q, want %q", got, "New tag")
	}
}

// TestGlobalBindingsExactSet guards PF-11 (design-spec.md §15, E7 T7, bean
// bt-m6at): globalBindings() is Header Zone 1's single source -- exactly the
// 7 header-global bindings, in the exact display order design-spec §15
// specifies (`ctrl+r:reload · ctrl+k:commands · p:repos · ?:help · esc:back
// · enter:open/confirm · q:quit`).
func TestGlobalBindingsExactSet(t *testing.T) {
	want := []keybind.Binding{keys.Refresh, keys.Palette, keys.Picker, keys.Help, keys.Back, keys.Enter, keys.Quit}
	got := globalBindings()
	if len(got) != len(want) {
		t.Fatalf("globalBindings() has %d entries, want %d", len(got), len(want))
	}
	for i := range want {
		if strings.Join(got[i].Keys(), ",") != strings.Join(want[i].Keys(), ",") {
			t.Errorf("globalBindings()[%d].Keys() = %v, want %v", i, got[i].Keys(), want[i].Keys())
		}
	}
}

// TestGlobalBindingHelpTextsShortened guards PF-11's Label-Kürzung
// (design-spec.md §15, Planner-Entscheidung "kurz und konsistent"):
// Refresh/Palette/Picker's Help().Desc shrink to single lowercase words --
// helpGroups()/the Help-Overlay reuse the SAME keybind.Binding objects
// (Single Source), so this shortening is the only edit needed anywhere.
func TestGlobalBindingHelpTextsShortened(t *testing.T) {
	cases := []struct {
		name string
		b    keybind.Binding
		want string
	}{
		{"Refresh", keys.Refresh, "reload"},
		{"Palette", keys.Palette, "commands"},
		{"Picker", keys.Picker, "repos"},
	}
	for _, c := range cases {
		if got := c.b.Help().Desc; got != c.want {
			t.Errorf("keys.%s.Help().Desc = %q, want %q", c.name, got, c.want)
		}
	}
}

// TestNoDuplicateBindingBetweenGlobalAndAnyLocalHintList is PF-11's own
// Drift-Guard (design-spec.md §15, epic-E7-plan.md Task 7 Step 1): no
// binding may appear in BOTH globalBindings() (Header Zone 1, always
// visible) AND either Chrome-calling view's own local-footer list --
// duplication is exactly what PF-11 removes. Scoped to the two VIEW-local
// lists only (browseRepoLocalBindings/backlogLocalBindings) -- the
// overlay-/search-/palette-/help-context footer sets (footer_context.go)
// are a DIFFERENT axis (Q04-Antwort) and deliberately restate Enter/Back
// for local reinforcement while a modal has full input capture (see that
// file's own doc comment for the rationale).
func TestNoDuplicateBindingBetweenGlobalAndAnyLocalHintList(t *testing.T) {
	global := map[string]bool{}
	for _, b := range globalBindings() {
		global[strings.Join(b.Keys(), ",")] = true
	}
	lists := map[string][]keybind.Binding{
		"browseRepoLocalBindings": browseRepoLocalBindings(),
		"backlogLocalBindings":    backlogLocalBindings(),
	}
	for name, list := range lists {
		for _, b := range list {
			id := strings.Join(b.Keys(), ",")
			if global[id] {
				t.Errorf("%s contains %v, which is also in globalBindings() -- duplicate header/footer binding (PF-11)", name, b.Keys())
			}
		}
	}
}

// TestHelpGroupsCoverEveryBindingExactlyOnce is a drift guard (T7 follow-up
// I02, bean bt-7jr8): reflects over every keybind.Binding field of keyMap and
// asserts helpGroups() references each one exactly once -- a future added
// keyMap field that nobody wired into a help group (or a copy-paste that
// lists one binding twice) fails loudly instead of silently drifting the
// in-app help / doc-generation source out of sync with the real bindings.
func TestHelpGroupsCoverEveryBindingExactlyOnce(t *testing.T) {
	k := newKeyMap()
	v := reflect.ValueOf(k)
	typ := v.Type()

	// identity keys a binding by its joined Keys() -- every field in this
	// keymap binds a distinct key set, so this is a safe, comparable stand-in
	// for keybind.Binding itself (which isn't usable as a map key directly).
	type coverage struct {
		field string
		count int
	}
	byIdentity := map[string]*coverage{}
	for i := 0; i < v.NumField(); i++ {
		b, ok := v.Field(i).Interface().(keybind.Binding)
		if !ok {
			continue
		}
		byIdentity[strings.Join(b.Keys(), ",")] = &coverage{field: typ.Field(i).Name}
	}
	if len(byIdentity) == 0 {
		t.Fatal("no keybind.Binding fields found on keyMap -- reflection scan is broken")
	}

	for _, g := range k.helpGroups() {
		for _, b := range g.bindings {
			id := strings.Join(b.Keys(), ",")
			c, ok := byIdentity[id]
			if !ok {
				t.Errorf("helpGroups() %q references a binding %v that is not a keyMap field", g.title, b.Keys())
				continue
			}
			c.count++
		}
	}

	for id, c := range byIdentity {
		switch {
		case c.count == 0:
			t.Errorf("keyMap field %s (%s) is not covered by any helpGroups() entry", c.field, id)
		case c.count > 1:
			t.Errorf("keyMap field %s (%s) is covered %d times by helpGroups() (want exactly once)", c.field, id, c.count)
		}
	}
}
