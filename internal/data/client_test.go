package data

import (
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
