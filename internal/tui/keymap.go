package tui

// keymap.go — the central, typed single source of every bt keybinding, ported
// from the devd-TUI (~/Obsidian/tools/DeveloperDashboard/apps/cli-go/internal/
// tui/keymap.go). Scope for this port (design-spec.md §7, implementation-plan
// »E1 Task 7«): the global + node-focused bindings needed by the chrome layer
// and the App-Shell (T8). devd also carries Sort/Tags(-Manager)/Rename/Review-
// verdict/User-Story bindings — out of scope here per design-spec §9 (Tag-
// Manager-CRUD and Docs/Notes views don't exist in beans-tui) or belong to a
// later epic (Review-Cockpit, E4); they get added to this single source when
// their view lands, not speculatively now (YAGNI).
//
// Port-adaptation vs. devd: `B` (Blocking-Picker) is new — beans has no devd
// equivalent (blocking/blocked_by relation, design-spec §4). `Editor` gains a
// bare `e` alongside `ctrl+e` (design-spec §7: "e/ctrl+e Editor").

import keybind "github.com/charmbracelet/bubbles/key"

// keyMap is the central, typed single source of every bt keybinding (mirrors
// devd DD2-47): navKey() derives the direction-cross normalization from
// Up/Down/Left/Right, the in-app help (?) and any future external doc
// generate from Help(), and re-binding a key changes exactly one place.
type keyMap struct {
	// Direction cross (evaluated by navKey — the up/down/left/right truth
	// lives here). Arrows are always an alias, per every binding below.
	Up    keybind.Binding
	Down  keybind.Binding
	Left  keybind.Binding
	Right keybind.Binding

	// Activation / return / global.
	Enter   keybind.Binding // enter — open/confirm
	Back    keybind.Binding // esc — back
	Quit    keybind.Binding // q / ctrl+c — quit (confirm)
	Help    keybind.Binding // ? — help overlay
	Palette keybind.Binding // ctrl+k / K — Command-Center
	Picker  keybind.Binding // p — repo-picker
	Reviews keybind.Binding // R — Review-Cockpit
	Backlog keybind.Binding // b — Backlog
	Search  keybind.Binding // / — search
	Filter  keybind.Binding // f — facet filter
	Yank    keybind.Binding // y — copy context (OSC52 + native)
	Refresh keybind.Binding // ctrl+r — manual data reload
	Section keybind.Binding // 1…9 — accordion section jump

	FilterClear keybind.Binding // X — reset filters
	Toggle      keybind.Binding // space/x — toggle facet checkbox (E2 Task 4, bean bt-9ldr)
	Sort        keybind.Binding // S — cycle Backlog sort mode (E2 Task 5, bean bt-gzu6)

	// Node-focused (act on the focused tree/list node).
	Status    keybind.Binding // s — status menu (all node types)
	Assign    keybind.Binding // a — parent assignment
	Blocking  keybind.Binding // B — blocking picker
	Create    keybind.Binding // c — create
	Delete    keybind.Binding // d — delete (confirm, no cascade -- E3 Task 6, bean bt-ppzb)
	TagAssign keybind.Binding // t — tag picker
	Editor    keybind.Binding // e / ctrl+e — edit body in $EDITOR
}

// newKeyMap returns the currently active keybinding set. The direction cross
// uses the jkli layout ported 1:1 from devd (DD2-34: i=up, j=left, k=down,
// l=right — inverted-T around the right index finger); arrow keys remain a
// second binding in every direction so arrow-only users are never broken.
// Every handler that routes through navKey() therefore supports both
// automatically.
func newKeyMap() keyMap {
	return keyMap{
		Up:    keybind.NewBinding(keybind.WithKeys("up", "i"), keybind.WithHelp("↑/i", "up")),
		Down:  keybind.NewBinding(keybind.WithKeys("down", "k"), keybind.WithHelp("↓/k", "down")),
		Left:  keybind.NewBinding(keybind.WithKeys("left", "j"), keybind.WithHelp("←/j", "back/out")),
		Right: keybind.NewBinding(keybind.WithKeys("right", "l"), keybind.WithHelp("→/l", "in/expand")),

		Enter:   keybind.NewBinding(keybind.WithKeys("enter"), keybind.WithHelp("enter", "open/confirm")),
		Back:    keybind.NewBinding(keybind.WithKeys("esc"), keybind.WithHelp("esc", "back")),
		Quit:    keybind.NewBinding(keybind.WithKeys("q", "ctrl+c"), keybind.WithHelp("q", "quit")),
		Help:    keybind.NewBinding(keybind.WithKeys("?"), keybind.WithHelp("?", "help")),
		Palette: keybind.NewBinding(keybind.WithKeys("ctrl+k", "K"), keybind.WithHelp("ctrl+k", "Command-Center")),
		Picker:  keybind.NewBinding(keybind.WithKeys("p"), keybind.WithHelp("p", "Repo-Picker")),
		Reviews: keybind.NewBinding(keybind.WithKeys("R"), keybind.WithHelp("R", "Review-Cockpit")),
		Backlog: keybind.NewBinding(keybind.WithKeys("b"), keybind.WithHelp("b", "Backlog")),
		Search:  keybind.NewBinding(keybind.WithKeys("/"), keybind.WithHelp("/", "Search")),
		Filter:  keybind.NewBinding(keybind.WithKeys("f"), keybind.WithHelp("f", "Filter")),
		Yank:    keybind.NewBinding(keybind.WithKeys("y"), keybind.WithHelp("y", "Copy context")),
		Refresh: keybind.NewBinding(keybind.WithKeys("ctrl+r"), keybind.WithHelp("ctrl+r", "Reload data")),
		Section: keybind.NewBinding(keybind.WithKeys("1", "2", "3", "4", "5", "6", "7", "8", "9"), keybind.WithHelp("1…9", "Section")),

		FilterClear: keybind.NewBinding(keybind.WithKeys("X"), keybind.WithHelp("X", "Clear filters")),
		Toggle:      keybind.NewBinding(keybind.WithKeys(" ", "x"), keybind.WithHelp("space/x", "Toggle facet")),
		Sort:        keybind.NewBinding(keybind.WithKeys("S"), keybind.WithHelp("S", "Sort")),

		Status:    keybind.NewBinding(keybind.WithKeys("s"), keybind.WithHelp("s", "Status menu")),
		Assign:    keybind.NewBinding(keybind.WithKeys("a"), keybind.WithHelp("a", "Assign parent")),
		Blocking:  keybind.NewBinding(keybind.WithKeys("B"), keybind.WithHelp("B", "Blocking picker")),
		Create:    keybind.NewBinding(keybind.WithKeys("c"), keybind.WithHelp("c", "Create")),
		Delete:    keybind.NewBinding(keybind.WithKeys("d"), keybind.WithHelp("d", "Delete")),
		TagAssign: keybind.NewBinding(keybind.WithKeys("t"), keybind.WithHelp("t", "Assign tags")),
		Editor:    keybind.NewBinding(keybind.WithKeys("e", "ctrl+e"), keybind.WithHelp("e", "Edit in $EDITOR")),
	}
}

// keys is the process-wide active keymap (single source).
var keys = newKeyMap()

// helpGroup is a named block of the shortcut overview (devd DD2-31/DD2-5
// pattern): the in-app help (?) and any external shortcut doc generate from
// this, so display and doc can never drift from the real bindings.
type helpGroup struct {
	title    string
	bindings []keybind.Binding
}

// helpGroups returns the keymap grouped for the in-app help overlay (added in
// a later task — the help view itself is out of scope here).
func (k keyMap) helpGroups() []helpGroup {
	return []helpGroup{
		{"Navigation", []keybind.Binding{k.Up, k.Down, k.Left, k.Right, k.Enter, k.Back, k.Section}},
		{"Views & Global", []keybind.Binding{k.Backlog, k.Reviews, k.Picker, k.Search, k.Filter, k.FilterClear, k.Refresh, k.Palette, k.Help, k.Quit}},
		{"Actions", []keybind.Binding{k.Status, k.Assign, k.TagAssign, k.Blocking, k.Create, k.Delete, k.Editor, k.Yank, k.Toggle, k.Sort}},
	}
}

// bindHas reports whether key k is part of binding b (exact match).
func bindHas(b keybind.Binding, k string) bool {
	for _, v := range b.Keys() {
		if v == k {
			return true
		}
	}
	return false
}

// navKey normalizes a raw key to a canonical direction ("up"/"down"/"left"/
// "right") using the central keymap. Keys with no direction meaning pass
// through unchanged. Every handler that routes through navKey therefore
// supports arrow keys and the jkli layout equally, and follows any future
// re-binding (DD2-34-style) without an edit at the call site.
func navKey(k string) string {
	switch {
	case bindHas(keys.Up, k):
		return "up"
	case bindHas(keys.Down, k):
		return "down"
	case bindHas(keys.Left, k):
		return "left"
	case bindHas(keys.Right, k):
		return "right"
	}
	return k
}
