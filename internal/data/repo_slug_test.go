package data

import (
	"os"
	"path/filepath"
	"testing"
)

// repo_slug_test.go — bt-d3ps (epic-E13-plan.md Item 4, PO-Redefinition
// Grilling 2026-07-17): TDD coverage for RepoSlug's four documented cases
// (bean acceptance: "beans.prefix (trimmed) primär, Dir-Basename Fallback
// bei fehlendem/unlesbarem .beans.yml").

// TestRepoSlugUsesPrefixFromBeansYML: primary source, .beans.yml present
// with a trailing-hyphen prefix (newTestRepo's own fixture, prefix "tt-").
func TestRepoSlugUsesPrefixFromBeansYML(t *testing.T) {
	dir := newTestRepo(t)
	if got := RepoSlug(dir); got != "tt" {
		t.Errorf("RepoSlug(%q) = %q, want %q", dir, got, "tt")
	}
}

// TestRepoSlugFallsBackToDirBaseWhenBeansYMLMissing: no .beans.yml at all
// (e.g. a repo dir not yet resolved via FindRepo, or a stale config.yaml
// entry) -- must not error, falls back to filepath.Base.
func TestRepoSlugFallsBackToDirBaseWhenBeansYMLMissing(t *testing.T) {
	dir := t.TempDir()
	want := filepath.Base(dir)
	if got := RepoSlug(dir); got != want {
		t.Errorf("RepoSlug(%q) = %q, want dir basename %q", dir, got, want)
	}
}

// TestRepoSlugFallsBackToDirBaseWhenPrefixEmpty: .beans.yml exists but
// carries no beans.prefix key (or an empty one) -- project.name is
// deliberately NOT read (plan's own "Begründung gegen project.name").
func TestRepoSlugFallsBackToDirBaseWhenPrefixEmpty(t *testing.T) {
	dir := t.TempDir()
	yml := "project:\n    name: no-prefix-repo\nbeans:\n    path: .beans\n"
	if err := os.WriteFile(filepath.Join(dir, ".beans.yml"), []byte(yml), 0o644); err != nil {
		t.Fatalf("write .beans.yml: %v", err)
	}
	want := filepath.Base(dir)
	if got := RepoSlug(dir); got != want {
		t.Errorf("RepoSlug(%q) = %q, want dir basename %q (project.name must NOT be read)", dir, got, want)
	}
}

// TestRepoSlugHandlesPrefixWithoutTrailingHyphen: a prefix with no trailing
// "-" (unusual but not invalid .beans.yml) must pass through unchanged, not
// be mangled by TrimSuffix.
func TestRepoSlugHandlesPrefixWithoutTrailingHyphen(t *testing.T) {
	dir := t.TempDir()
	yml := "beans:\n    prefix: xy\n"
	if err := os.WriteFile(filepath.Join(dir, ".beans.yml"), []byte(yml), 0o644); err != nil {
		t.Fatalf("write .beans.yml: %v", err)
	}
	if got := RepoSlug(dir); got != "xy" {
		t.Errorf("RepoSlug(%q) = %q, want %q", dir, got, "xy")
	}
}

// TestRepoSlugCorruptYAMLFallsBackToDirBase: unreadable/corrupt YAML must
// never crash or propagate an error -- same "missing file is not a failure"
// contract as config.LoadSettings (settings.go).
func TestRepoSlugCorruptYAMLFallsBackToDirBase(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".beans.yml"), []byte("not: [valid: yaml"), 0o644); err != nil {
		t.Fatalf("write .beans.yml: %v", err)
	}
	want := filepath.Base(dir)
	if got := RepoSlug(dir); got != want {
		t.Errorf("RepoSlug(%q) = %q, want dir basename %q on corrupt YAML", dir, got, want)
	}
}
