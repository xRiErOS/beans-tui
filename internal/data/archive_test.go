package data

// archive_test.go — E5 Task 7 (bean bt-ggt2, epic bt-5h4d): TDD coverage for
// Bean.IsArchived() (bean.go) and the empirical claim design decision e
// rests on -- List() (beans list --json --full) already reads .beans/
// archive/ without any code change here, so a real `beans archive` run must
// leave an archived bean fully visible in List(), just with its Path
// prefixed "archive/".

import (
	"os/exec"
	"strings"
	"testing"
)

// TestBeanIsArchivedDetectsPathPrefix is the pure, no-binary-needed half:
// IsArchived() is a plain string-prefix check against an already-loaded
// Path value.
func TestBeanIsArchivedDetectsPathPrefix(t *testing.T) {
	cases := []struct {
		name string
		path string
		want bool
	}{
		{"root-level, not archived", "tt-mlst--test-milestone.md", false},
		{"archived", "archive/tt-mlst--test-milestone.md", true},
		{"empty path", "", false},
		{"archive as a mid-path segment only, not a prefix", "sub/archive/x.md", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			b := Bean{Path: c.path}
			if got := b.IsArchived(); got != c.want {
				t.Errorf("Bean{Path: %q}.IsArchived() = %v, want %v", c.path, got, c.want)
			}
		})
	}
}

// TestListIncludesArchivedBeans is the empirical proof behind design
// decision e (epic-E5-plan.md »Task 7«): Client.List() needs NO code change
// to see archived beans -- Core.loadFromDisk's WalkDir already covers
// .beans/archive/ (it is not a dot-prefixed directory). Real beans binary,
// real `beans archive` invocation (requireBeansBinary-guarded).
func TestListIncludesArchivedBeans(t *testing.T) {
	requireBeansBinary(t)

	repo := newTestRepo(t)
	client := &Client{RepoDir: repo}

	before, err := client.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(before) != 3 {
		t.Fatalf("len(before) = %d, want 3 (newTestRepo's fixture beans)", len(before))
	}
	task := findBean(t, before, "tt-task")
	if task.IsArchived() {
		t.Fatalf("setup: tt-task already reports IsArchived() = true before any archive run, Path=%q", task.Path)
	}

	if err := client.SetStatus(task.ID, "completed", task.ETag); err != nil {
		t.Fatalf("SetStatus(tt-task, completed) error = %v", err)
	}

	cmd := exec.Command("beans", "archive")
	cmd.Dir = repo
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("beans archive: %v\n%s", err, out)
	}

	after, err := client.List()
	if err != nil {
		t.Fatalf("List() after archive error = %v", err)
	}
	if len(after) != 3 {
		t.Fatalf("len(after) = %d, want 3 -- an archived bean must remain visible in List(), design decision e (\"archived beans remain visible in all queries\")", len(after))
	}

	archived := findBean(t, after, "tt-task")
	if archived.Status != "completed" {
		t.Errorf("Status = %q after archive, want completed (unchanged, beans archive only moves the file)", archived.Status)
	}
	if !strings.HasPrefix(archived.Path, "archive/") {
		t.Errorf("Path = %q after archive, want an \"archive/\" prefix", archived.Path)
	}
	if !archived.IsArchived() {
		t.Error("IsArchived() = false for a bean List() reports under archive/, want true")
	}

	// Sibling beans untouched by the archive run stay non-archived.
	epic := findBean(t, after, "tt-epic")
	if epic.IsArchived() {
		t.Errorf("tt-epic (never archived) reports IsArchived() = true, Path=%q", epic.Path)
	}
}
