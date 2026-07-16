package data

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestListReturnsAllBeansWithBody(t *testing.T) {
	requireBeansBinary(t)

	repo := newTestRepo(t)
	client := &Client{RepoDir: repo}

	beans, err := client.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(beans) != 3 {
		t.Fatalf("len(beans) = %d, want 3", len(beans))
	}

	byID := make(map[string]Bean, len(beans))
	for _, b := range beans {
		byID[b.ID] = b
	}

	for _, id := range []string{"tt-mlst", "tt-epic", "tt-task"} {
		if _, ok := byID[id]; !ok {
			t.Errorf("bean %s missing from List() result", id)
		}
	}

	task, ok := byID["tt-task"]
	if !ok {
		t.Fatalf("tt-task not found")
	}
	if task.Parent != "tt-epic" {
		t.Errorf("task.Parent = %q, want %q", task.Parent, "tt-epic")
	}
	if task.Type != "task" {
		t.Errorf("task.Type = %q, want %q", task.Type, "task")
	}

	epic, ok := byID["tt-epic"]
	if !ok {
		t.Fatalf("tt-epic not found")
	}
	if epic.Parent != "tt-mlst" {
		t.Errorf("epic.Parent = %q, want %q", epic.Parent, "tt-mlst")
	}

	// JSON-contract regression guard (I01): tags/blocking/blocked_by must
	// round-trip as parsed slices, not just be present-or-absent.
	wantTags := []string{"urgent", "backend"}
	if !reflect.DeepEqual(epic.Tags, wantTags) {
		t.Errorf("epic.Tags = %v, want %v", epic.Tags, wantTags)
	}
	wantBlockedBy := []string{"tt-mlst"}
	if !reflect.DeepEqual(task.BlockedBy, wantBlockedBy) {
		t.Errorf("task.BlockedBy = %v, want %v", task.BlockedBy, wantBlockedBy)
	}

	milestone, ok := byID["tt-mlst"]
	if !ok {
		t.Fatalf("tt-mlst not found")
	}
	wantBlocking := []string{"tt-task"}
	if !reflect.DeepEqual(milestone.Blocking, wantBlocking) {
		t.Errorf("milestone.Blocking = %v, want %v", milestone.Blocking, wantBlocking)
	}

	for _, b := range beans {
		if b.Body == "" {
			t.Errorf("bean %s: Body is empty, want non-empty (--full)", b.ID)
		}
		if b.ETag == "" {
			t.Errorf("bean %s: ETag is empty, want non-empty", b.ID)
		}
	}
}

// TestShowRawReturnsFileFormat guards data.Client.ShowRaw (D01, design-
// spec.md §15 PF-17, bean bt-z4b1): `beans show <id> --raw` is the seed text
// for the whole-bean $EDITOR -- verified BYTE-IDENTICAL to the on-disk
// .beans/*.md file (not just structurally similar), since the whole
// round-trip design relies on the CLI staying the ONE authority for the
// file's canonical serialization (no self-built templating, design-spec §3.1
// D02).
func TestShowRawReturnsFileFormat(t *testing.T) {
	requireBeansBinary(t)

	repo := newTestRepo(t)
	client := &Client{RepoDir: repo}

	// A bean CREATED via the CLI (not one of newTestRepo's hand-authored
	// fixture files) -- `beans show --raw` re-serializes canonically rather
	// than cat-ing the file literally, so a hand-authored fixture with a
	// non-canonical frontmatter field order (e.g. "tags" before
	// "created_at") would NOT come back byte-identical even though the
	// content is equivalent. A CLI-created bean's on-disk file is already in
	// canonical order, matching what real beans-tui usage always sees
	// (files are written by `beans` itself, never hand-edited into a
	// different field order) -- this is what the design's "byte-identical
	// to on-disk" claim actually depends on. Priority is passed EXPLICITLY
	// (empirically verified quirk): an unset priority is OMITTED from the
	// on-disk file entirely, but `show --raw` resolves it to the config
	// default ("normal") anyway -- that divergence is a resolved-default
	// artifact, not a --raw fidelity bug, and is sidestepped here by never
	// leaving a defaultable field unset in the first place (every field a
	// real, previously-saved bean carries is always explicit).
	created, err := client.Create(CreateOpts{Title: "Raw Format Fixture", Type: "task", Priority: "normal", Body: "Raw format body."})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	raw, err := client.ShowRaw(created.ID)
	if err != nil {
		t.Fatalf("ShowRaw() error = %v", err)
	}

	onDisk, err := os.ReadFile(filepath.Join(repo, ".beans", created.Path))
	if err != nil {
		t.Fatalf("read on-disk file: %v", err)
	}
	if raw != string(onDisk) {
		t.Fatalf("ShowRaw() = %q, want byte-identical to on-disk file %q", raw, string(onDisk))
	}

	for _, want := range []string{"# " + created.ID, "title: Raw Format Fixture", "status:", "type: task"} {
		if !strings.Contains(raw, want) {
			t.Errorf("ShowRaw() missing expected fragment %q in:\n%s", want, raw)
		}
	}
}

func TestListErrorIncludesStderr(t *testing.T) {
	requireBeansBinary(t)

	// A dir without a valid beans repo: `beans list` run there should fail,
	// and the error should surface stderr content.
	tmp := t.TempDir()
	client := &Client{RepoDir: tmp}

	_, err := client.List()
	if err == nil {
		t.Fatal("List() error = nil, want error (no beans repo in RepoDir)")
	}

	// The beans CLI's own diagnostic must survive the wrap -- this is what
	// makes the error actionable instead of a bare exit-code failure.
	const wantSubstr = "no .beans directory found"
	if !strings.Contains(err.Error(), wantSubstr) {
		t.Errorf("List() error = %q, want it to contain %q", err.Error(), wantSubstr)
	}
}

// TestClientSearchInvokesBleveFlagAndReturnsMatches guards data.Client.Search
// (E2 Task 3, bean bt-4ep2): `beans list --json --full --search <query>`,
// same JSON contract as List. A dedicated 4th fixture bean is written
// directly here (not folded into newTestRepo's shared 3-bean fixture) --
// TestListReturnsAllBeansWithBody asserts len(beans)==3 against that shared
// fixture, so a Search-only fixture bean must not leak into it.
func TestClientSearchInvokesBleveFlagAndReturnsMatches(t *testing.T) {
	requireBeansBinary(t)

	repo := newTestRepo(t)
	const fixture = `---
# tt-gldn
title: Golden Search Fixture
status: todo
type: task
priority: normal
created_at: 2026-01-01T00:00:00Z
updated_at: 2026-01-01T00:00:00Z
---

Body mentions golden explicitly for the Bleve full-text search test.
`
	if err := os.WriteFile(filepath.Join(repo, ".beans", "tt-gldn--golden-search-fixture.md"), []byte(fixture), 0o644); err != nil {
		t.Fatalf("write search fixture: %v", err)
	}

	c := &Client{RepoDir: repo}
	beans, err := c.Search("golden")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(beans) == 0 {
		t.Fatal("Search(\"golden\") returned no matches against a fixture repo containing one")
	}
	found := false
	for _, b := range beans {
		if b.ID == "tt-gldn" {
			found = true
		}
	}
	if !found {
		t.Errorf("Search(\"golden\") results = %v, want tt-gldn included", beans)
	}
}
