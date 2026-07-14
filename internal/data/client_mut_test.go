package data

import (
	"errors"
	"strings"
	"testing"
)

// findBean is a small test helper -- every mutation test below needs to
// pull one specific bean (with its current ETag) out of a List() result.
func findBean(t *testing.T, beans []Bean, id string) Bean {
	t.Helper()
	for _, b := range beans {
		if b.ID == id {
			return b
		}
	}
	t.Fatalf("bean %s not found in List() result", id)
	return Bean{}
}

func TestSetStatusRoundtrip(t *testing.T) {
	requireBeansBinary(t)

	repo := newTestRepo(t)
	client := &Client{RepoDir: repo}

	beans, err := client.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	task := findBean(t, beans, "tt-task")

	if err := client.SetStatus(task.ID, "completed", task.ETag); err != nil {
		t.Fatalf("SetStatus() error = %v", err)
	}

	after, err := client.List()
	if err != nil {
		t.Fatalf("List() after SetStatus error = %v", err)
	}
	updated := findBean(t, after, "tt-task")
	if updated.Status != "completed" {
		t.Errorf("Status = %q, want %q", updated.Status, "completed")
	}
}

func TestCreateReturnsNewBean(t *testing.T) {
	requireBeansBinary(t)

	repo := newTestRepo(t)
	client := &Client{RepoDir: repo}

	created, err := client.Create(CreateOpts{
		Title:  "New Feature Bean",
		Type:   "feature",
		Status: "todo",
		Tags:   []string{"alpha"},
		Body:   "Created by TestCreateReturnsNewBean.",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.ID == "" {
		t.Error("Create() bean.ID is empty")
	}
	if created.ETag == "" {
		t.Error("Create() bean.ETag is empty")
	}
	if created.Title != "New Feature Bean" {
		t.Errorf("Create() bean.Title = %q, want %q", created.Title, "New Feature Bean")
	}
	if created.Type != "feature" {
		t.Errorf("Create() bean.Type = %q, want %q", created.Type, "feature")
	}

	beans, err := client.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	found := findBean(t, beans, created.ID)
	if found.Title != "New Feature Bean" {
		t.Errorf("List() bean.Title = %q, want %q", found.Title, "New Feature Bean")
	}
}

func TestConflictOnStaleETag(t *testing.T) {
	requireBeansBinary(t)

	repo := newTestRepo(t)
	client := &Client{RepoDir: repo}

	beans, err := client.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	task := findBean(t, beans, "tt-task")
	staleETag := task.ETag

	// Change the bean externally (a second, successful update), which
	// rotates the ETag out from under the stale copy above.
	if err := client.SetStatus(task.ID, "in-progress", staleETag); err != nil {
		t.Fatalf("first SetStatus() error = %v", err)
	}

	err = client.SetPriority(task.ID, "high", staleETag)
	if err == nil {
		t.Fatal("SetPriority() with stale ETag: error = nil, want ErrConflict")
	}
	if !errors.Is(err, ErrConflict) {
		t.Errorf("SetPriority() error = %v, want errors.Is(err, ErrConflict)", err)
	}
}

func TestAppendBodyAddsSection(t *testing.T) {
	requireBeansBinary(t)

	repo := newTestRepo(t)
	client := &Client{RepoDir: repo}

	beans, err := client.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	task := findBean(t, beans, "tt-task")

	const appended = "## Notes\nAppended via TestAppendBodyAddsSection."
	if err := client.AppendBody(task.ID, "\n"+appended, task.ETag); err != nil {
		t.Fatalf("AppendBody() error = %v", err)
	}

	after, err := client.List()
	if err != nil {
		t.Fatalf("List() after AppendBody error = %v", err)
	}
	updated := findBean(t, after, "tt-task")
	if !strings.Contains(updated.Body, "Appended via TestAppendBodyAddsSection.") {
		t.Errorf("Body = %q, want it to contain the appended section", updated.Body)
	}
}

func TestDeleteRemovesBean(t *testing.T) {
	requireBeansBinary(t)

	repo := newTestRepo(t)
	client := &Client{RepoDir: repo}

	if err := client.Delete("tt-task"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	beans, err := client.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	for _, b := range beans {
		if b.ID == "tt-task" {
			t.Error("tt-task still present in List() after Delete()")
		}
	}
}
