package config

// settings_test.go — TDD coverage for Settings (E5 Task 5, bean bt-0l8c),
// Port devd internal/config/settings_test.go's pattern, reduced to bt's own
// schema (repos/editor/theme.accent/layout.tree_width).

import (
	"os"
	"reflect"
	"testing"
)

func TestParseSettingsYAML(t *testing.T) {
	in := []byte("repos:\n  - /repo/a\n  - /repo/b\neditor: \"code -w\"\ntheme:\n  accent: \"#f5a97f\"\nlayout:\n  tree_width: 40\n")
	s, err := parseSettings(in)
	if err != nil {
		t.Fatalf("parseSettings: %v", err)
	}
	wantRepos := []string{"/repo/a", "/repo/b"}
	if !reflect.DeepEqual(s.Repos, wantRepos) {
		t.Errorf("Repos = %v, want %v", s.Repos, wantRepos)
	}
	if s.Editor != "code -w" {
		t.Errorf("Editor = %q, want %q", s.Editor, "code -w")
	}
	if s.Theme.Accent != "#f5a97f" {
		t.Errorf("Theme.Accent = %q, want %q", s.Theme.Accent, "#f5a97f")
	}
	if s.Layout.TreeWidth != 40 {
		t.Errorf("Layout.TreeWidth = %d, want 40", s.Layout.TreeWidth)
	}
}

func TestMergeSettingsOnlySetFieldsWin(t *testing.T) {
	base := DefaultSettings() // TreeWidth=36, everything else zero
	over := Settings{
		Editor: "nano",
		Layout: LayoutSettings{TreeWidth: 40}, // Theme.Accent "" = unset -> base stays
	}
	got := mergeSettings(base, over)
	if got.Editor != "nano" {
		t.Errorf("Editor = %q, want %q (override)", got.Editor, "nano")
	}
	if got.Layout.TreeWidth != 40 {
		t.Errorf("Layout.TreeWidth = %d, want 40 (override)", got.Layout.TreeWidth)
	}
	if got.Theme.Accent != "" {
		t.Errorf("Theme.Accent = %q, want %q (unset -> base)", got.Theme.Accent, "")
	}
	if len(got.Repos) != 0 {
		t.Errorf("Repos = %v, want empty (unset override -> base)", got.Repos)
	}

	over2 := Settings{Repos: []string{"/repo/x"}}
	got2 := mergeSettings(base, over2)
	if !reflect.DeepEqual(got2.Repos, []string{"/repo/x"}) {
		t.Errorf("Repos = %v, want [/repo/x] (override)", got2.Repos)
	}
}

func TestValidateSettingsClampsTreeWidth(t *testing.T) {
	cases := []struct {
		in   int
		want int
	}{
		{5, minTreeWidth},
		{999, maxTreeWidth},
		{0, defTreeWidth},
		{-1, defTreeWidth},
		{30, 30}, // within range -> unchanged
	}
	for _, c := range cases {
		got := validateSettings(Settings{Layout: LayoutSettings{TreeWidth: c.in}})
		if got.Layout.TreeWidth != c.want {
			t.Errorf("validateSettings(TreeWidth=%d).Layout.TreeWidth = %d, want %d", c.in, got.Layout.TreeWidth, c.want)
		}
	}
}

func TestValidateSettingsRejectsInvalidAccentHex(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"nope", ""},
		{"#zzzzzz", ""},
		{"", ""},
		{"#A1B2C3", "#A1B2C3"}, // valid -> kept verbatim
	}
	for _, c := range cases {
		got := validateSettings(Settings{Theme: ThemeSettings{Accent: c.in}})
		if got.Theme.Accent != c.want {
			t.Errorf("validateSettings(Accent=%q).Theme.Accent = %q, want %q", c.in, got.Theme.Accent, c.want)
		}
	}
}

func TestSaveUserSettingsReadModifyWrite(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	if err := SaveUserSettings([]string{"/repo/a"}, "vim", "#f5a97f", 40); err != nil {
		t.Fatalf("first SaveUserSettings: %v", err)
	}
	// A second save MODIFIES the existing file (read-modify-write) rather
	// than starting a fresh Settings{} each time.
	if err := SaveUserSettings([]string{"/repo/a", "/repo/b"}, "nano", "#abc123", 50); err != nil {
		t.Fatalf("second SaveUserSettings (modify): %v", err)
	}

	path, err := settingsPath()
	if err != nil {
		t.Fatalf("settingsPath: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("config.yaml not written at %s: %v", path, err)
	}

	got, err := LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings: %v", err)
	}
	wantRepos := []string{"/repo/a", "/repo/b"}
	if !reflect.DeepEqual(got.Repos, wantRepos) {
		t.Errorf("Repos = %v, want %v (second save's values)", got.Repos, wantRepos)
	}
	if got.Editor != "nano" {
		t.Errorf("Editor = %q, want %q", got.Editor, "nano")
	}
	if got.Theme.Accent != "#abc123" {
		t.Errorf("Theme.Accent = %q, want %q", got.Theme.Accent, "#abc123")
	}
	if got.Layout.TreeWidth != 50 {
		t.Errorf("Layout.TreeWidth = %d, want 50", got.Layout.TreeWidth)
	}
}

// TestLoadSettingsMissingFileReturnsDefaults guards the "Missing-Config-
// Robustheit" contract (bean bt-0l8c smoke step): a fresh $HOME with no
// ~/.config/beans-tui at all must never error -- Defaults apply silently.
func TestLoadSettingsMissingFileReturnsDefaults(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	got, err := LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings on a fresh HOME: %v", err)
	}
	want := DefaultSettings()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("LoadSettings() = %+v, want DefaultSettings() %+v", got, want)
	}
}
