package tui

import (
	"reflect"
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
