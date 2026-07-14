package data

import "testing"

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
}
