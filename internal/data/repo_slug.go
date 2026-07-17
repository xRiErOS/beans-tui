package data

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// repo_slug.go — bt-d3ps (epic-E13-plan.md Item 4, PO-Redefinition Grilling
// 2026-07-17, replaces the earlier discovery-scan design entirely -- NO
// scan, NO discovery roots, NO find-persistence): the Lobby (view_lobby.go's
// repoPickerBody) needs a short, stable label per config-registered repo.
// RepoSlug supplies it.

// repoSlugConfig is the MINIMAL .beans.yml shape RepoSlug needs -- just
// beans.prefix, not a full schema mirror (no other beans-tui code parses
// .beans.yml's content today, only FindRepo's bare-existence check,
// discover.go's own doc-stamp).
type repoSlugConfig struct {
	Beans struct {
		Prefix string `yaml:"prefix"`
	} `yaml:"beans"`
}

// RepoSlug resolves a short, cross-repo-stable identifier for repoDir:
// <repoDir>/.beans.yml's beans.prefix (trimmed of a trailing "-"), if
// present. Bean-IDs (bt-d3ps, lean-stack-58j0, ... -- visible everywhere in
// the UI) are coupled to this prefix, NOT to .beans.yml's optional
// project.name (which some repos omit entirely, e.g. lean-stack's own --
// epic-E13-plan.md Item 4's own "Begründung gegen project.name"), so
// project.name is deliberately never read here.
//
// Falls back to filepath.Base(repoDir) whenever .beans.yml is missing,
// unreadable, corrupt, or its prefix is empty -- mirrors config.LoadSettings'
// "a missing/corrupt file is not a failure" contract (settings.go): RepoSlug
// never errors, so repoPickerBody (a pure render path) never needs an error
// branch of its own.
func RepoSlug(repoDir string) string {
	raw, err := os.ReadFile(filepath.Join(repoDir, configFileName))
	if err == nil {
		var cfg repoSlugConfig
		if yaml.Unmarshal(raw, &cfg) == nil {
			if prefix := strings.TrimSuffix(strings.TrimSpace(cfg.Beans.Prefix), "-"); prefix != "" {
				return prefix
			}
		}
	}
	return filepath.Base(repoDir)
}
