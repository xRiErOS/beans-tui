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

func TestCreateTitleWithLeadingDash(t *testing.T) {
	requireBeansBinary(t)

	repo := newTestRepo(t)
	client := &Client{RepoDir: repo}

	const title = "- fix bug"
	created, err := client.Create(CreateOpts{Title: title, Type: "task"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.Title != title {
		t.Errorf("Create() bean.Title = %q, want %q", created.Title, title)
	}

	beans, err := client.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	found := findBean(t, beans, created.ID)
	if found.Title != title {
		t.Errorf("List() bean.Title = %q, want %q", found.Title, title)
	}
}

func TestValidationErrorContainingEtagMismatchIsNotConflict(t *testing.T) {
	requireBeansBinary(t)

	repo := newTestRepo(t)
	client := &Client{RepoDir: repo}

	beans, err := client.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	task := findBean(t, beans, "tt-task")

	// The literal string "etag mismatch" here is user-supplied (an invalid
	// --type value), not a real conflict -- the CLI reports this as a
	// VALIDATION_ERROR, not CONFLICT. A naive strings.Contains(err.Error(),
	// "etag mismatch") check would misclassify this as ErrConflict (B02).
	err = client.SetType(task.ID, "etag mismatch", task.ETag)
	if err == nil {
		t.Fatal("SetType() with invalid type: error = nil, want error")
	}
	if errors.Is(err, ErrConflict) {
		t.Errorf("SetType() error = %v, want NOT errors.Is(err, ErrConflict)", err)
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

// TestSetTagsAddsAndRemovesInOneCall guards the E3 Task 2 design decision
// (bean bt-8v69): a SINGLE SetTags call combining --tag/--remove-tag adds
// AND removes tags against ONE etag.
//
// Uses tt-task, not tt-epic: testrepo_test.go's newTestRepo doc comment
// records a real beans 0.4.2 CLI bug where a bean with HAND-AUTHORED
// "tags:" frontmatter reports an ETag from list/show that diverges from
// update --if-match's internal conflict check, until the file has been
// rewritten once by beans itself. tt-task ships with no hand-authored tags
// (deliberately, per that same doc comment), so a first AddTag "seed" call
// establishes a tag to remove without ever touching the divergent path --
// the combined SetTags call below is then the thing under test.
func TestSetTagsAddsAndRemovesInOneCall(t *testing.T) {
	requireBeansBinary(t)

	repo := newTestRepo(t)
	client := &Client{RepoDir: repo}

	beans, err := client.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	task := findBean(t, beans, "tt-task")

	if err := client.AddTag(task.ID, "seed", task.ETag); err != nil {
		t.Fatalf("seed AddTag() error = %v", err)
	}

	beans, err = client.List()
	if err != nil {
		t.Fatalf("List() after seed error = %v", err)
	}
	task = findBean(t, beans, "tt-task")

	if err := client.SetTags(task.ID, []string{"alpha", "beta"}, []string{"seed"}, task.ETag); err != nil {
		t.Fatalf("SetTags() error = %v", err)
	}

	after, err := client.List()
	if err != nil {
		t.Fatalf("List() after SetTags error = %v", err)
	}
	updated := findBean(t, after, "tt-task")

	want := map[string]bool{"alpha": true, "beta": true}
	for _, tg := range updated.Tags {
		if tg == "seed" {
			t.Fatal("SetTags() left the removed tag \"seed\" in place")
		}
		if !want[tg] {
			t.Errorf("SetTags() left unexpected tag %q, want only %v", tg, want)
			continue
		}
		delete(want, tg)
	}
	if len(want) != 0 {
		t.Errorf("SetTags() missing tags %v, got %v", want, updated.Tags)
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
