// Package config persists beans-tui's TUI-wide configuration
// (~/.config/beans-tui/config.yaml) and lightweight runtime state
// (~/.config/beans-tui/state.json) -- ONE config directory for BOTH files
// (E5 Task 5, bean bt-0l8c, design decision c), a deliberate deviation from
// devd's split (~/.config/devd-cli/config.yaml vs. ~/.config/dd/state.json,
// Port devd internal/config/{settings,state}.go).
package config

// settings.go — Port devd internal/config/settings.go, reduced to the
// design-spec's own field set (§7/§9: "Settings (~/.config/beans-tui/
// config.yaml: repos, editor, Akzentfarbe, Baumbreite)") -- MINUS devd's
// ModalWidth/StartProject/Keybindings (not part of beans-tui's design-spec,
// YAGNI here per the bean's own acceptance checklist), PLUS Repos []string
// (design-spec §6 US-14: "Liste konfigurierter Repos", T6's Lobby/Repo-Picker
// consumes it). EIN Pfad-Layer (~/.config/beans-tui/config.yaml) -- unlike
// devd's second CWD-override layer (its own settingsPaths()), which is
// deliberately NOT ported (YAGNI, design decision c).
//
// DEVIATION vs. devd's DefaultSettings: Editor defaults to "" here, NOT
// devd's "nvim" -- design decision c is explicit that Settings.Editor's
// default must leave editorBinary()'s pre-existing $VISUAL/$EDITOR/vi
// cascade (editor.go, bean bt-sl45) byte-for-byte unchanged out of the box.
// A "nvim" default would silently override that cascade for every PO who
// has never touched config.yaml.
//
// Example ~/.config/beans-tui/config.yaml:
//
//	repos:
//	  - /Users/po/work/repo-a
//	  - /Users/po/work/repo-b
//	editor: "code -w"          # empty (default) = $VISUAL/$EDITOR/vi cascade
//	theme:
//	  accent: "#f5a97f"        # empty (default) = built-in Mauve
//	layout:
//	  tree_width: 40           # tree column width floor, 24-60

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Settings is beans-tui's TUI configuration (YAML). Zero values mean "not
// set" and fall back to DefaultSettings()/the previous layer on merge.
type Settings struct {
	Repos  []string       `yaml:"repos,omitempty"`
	Editor string         `yaml:"editor,omitempty"`
	Theme  ThemeSettings  `yaml:"theme"`
	Layout LayoutSettings `yaml:"layout"`
}

// ThemeSettings carries the Accent override (design-spec §9 "Akzentfarbe").
type ThemeSettings struct {
	Accent string `yaml:"accent"` // Hex (#rrggbb) or "" = built-in Mauve
}

// LayoutSettings carries the tree-column width floor (design-spec §9
// "Baumbreite"). NOT wired into the live render yet (clickPaneGeometry,
// mouse.go, still hardcodes the "24" floor every View function shares) --
// this task only persists/validates the value; wiring it into
// masterDetailWidths' treeWidthFloor parameter is a follow-up, deliberately
// out of scope here (not in this task's Modify-file list, epic-E5-plan.md
// »Task 5«).
type LayoutSettings struct {
	TreeWidth int `yaml:"tree_width"`
}

// Defaults -- Single Source of the built-in values (= current behavior).
const (
	defTreeWidth = 36
	minTreeWidth = 24
	maxTreeWidth = 60
)

var hexColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// DefaultSettings returns the built-in defaults (apply without a config
// file). Editor/Theme.Accent/Repos default to zero values -- see the
// DEVIATION doc-stamp above for why Editor is "" here, not devd's "nvim".
func DefaultSettings() Settings {
	return Settings{Layout: LayoutSettings{TreeWidth: defTreeWidth}}
}

// configDir resolves ~/.config/beans-tui -- the ONE directory config.yaml
// AND state.json (state.go) both live in (design decision c). Tests inject
// a fake HOME via t.Setenv("HOME", t.TempDir()) -- os.UserHomeDir() reads
// $HOME directly on the platforms bt targets (darwin/linux), so this needs
// no separate package-level override var (unlike configuredEditor, which
// has no environment-variable equivalent to piggyback on).
func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "beans-tui"), nil
}

// settingsPath returns ~/.config/beans-tui/config.yaml.
func settingsPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// LoadSettings loads Defaults -> user config (merged) and validates/clamps
// the result. A missing or unreadable/malformed config.yaml is NOT an error
// (Port devd's own "fehlende Dateien sind kein Fehler" contract) -- Defaults
// apply instead, so a fresh ~/.config/beans-tui (or no $HOME at all) never
// blocks TUI startup. err is only ever non-nil for a genuinely unexpected
// os.UserHomeDir failure.
func LoadSettings() (Settings, error) {
	s := DefaultSettings()
	path, err := settingsPath()
	if err != nil {
		return validateSettings(s), nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return validateSettings(s), nil // missing -> defaults
	}
	over, err := parseSettings(data)
	if err != nil {
		return validateSettings(s), nil // corrupt YAML -> defaults, never crash
	}
	return validateSettings(mergeSettings(s, over)), nil
}

// SaveUserSettings writes repos/editor/theme.accent/layout.tree_width into
// the user config (design decision c: ONE path, no CWD-override layer).
// Read-modify-write (Port devd SaveUserSettings's own rationale): an
// existing file is read first so a config.yaml field this task's form does
// not cover (future schema growth) survives the write instead of being
// clobbered back to its zero value. Directory created if missing.
func SaveUserSettings(repos []string, editor, accent string, treeWidth int) error {
	path, err := settingsPath()
	if err != nil {
		return err
	}
	var s Settings
	if data, rerr := os.ReadFile(path); rerr == nil {
		_ = yaml.Unmarshal(data, &s) // corrupt YAML -> start from zero rather than fail the save
	}
	s.Repos = reposTrimmed(repos)
	s.Editor = editor
	s.Theme.Accent = accent
	s.Layout.TreeWidth = treeWidth
	out, err := yaml.Marshal(s)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, out, 0o644)
}

// parseSettings decodes YAML into Settings (pure, FS-free -- testable).
func parseSettings(data []byte) (Settings, error) {
	var s Settings
	if err := yaml.Unmarshal(data, &s); err != nil {
		return Settings{}, err
	}
	return s, nil
}

// mergeSettings layers over onto base: only SET (non-zero) fields win. An
// empty Repos slice counts as "not set" (zero value), same as an empty
// Editor/Accent string or a zero TreeWidth.
func mergeSettings(base, over Settings) Settings {
	if len(over.Repos) > 0 {
		base.Repos = over.Repos
	}
	if over.Editor != "" {
		base.Editor = over.Editor
	}
	if over.Theme.Accent != "" {
		base.Theme.Accent = over.Theme.Accent
	}
	if over.Layout.TreeWidth != 0 {
		base.Layout.TreeWidth = over.Layout.TreeWidth
	}
	return base
}

// validateSettings clamps TreeWidth into [minTreeWidth,maxTreeWidth] and
// discards an invalid Accent (not #rrggbb) -> built-in Mauve stays.
func validateSettings(s Settings) Settings {
	s.Layout.TreeWidth = clampInt(s.Layout.TreeWidth, minTreeWidth, maxTreeWidth, defTreeWidth)
	if s.Theme.Accent != "" && !hexColor.MatchString(s.Theme.Accent) {
		s.Theme.Accent = ""
	}
	return s
}

// clampInt keeps v in [lo,hi]; 0/negative (= "not set") -> def.
func clampInt(v, lo, hi, def int) int {
	if v <= 0 {
		return def
	}
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// reposTrimmed strips surrounding whitespace from every repo path and drops
// empty entries -- shared by SaveUserSettings' caller (box_form_settings.go's
// linesToRepos) and available here for any future direct config.Settings{}
// construction that wants the same normalization.
func reposTrimmed(repos []string) []string {
	out := make([]string, 0, len(repos))
	for _, r := range repos {
		if r = strings.TrimSpace(r); r != "" {
			out = append(out, r)
		}
	}
	return out
}
