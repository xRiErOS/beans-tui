package tui

// bean_id_short_test.go — bean bt-pl5p (Nebenbefund N5, epic bt-vy1q): the
// left pane's rows drop the CURRENT repo's own bean-ID prefix ("sproutling-
// btv7" -> "btv7"), the freed columns go to the title. A foreign/dangling ID
// that does NOT carry the current prefix keeps its full form (cross-repo
// references must stay unambiguous).

import (
	"strings"
	"testing"

	"github.com/xRiErOS/beans-tui/internal/data"
)

func TestShortBeanID(t *testing.T) {
	cases := []struct {
		name, id, slug, want string
	}{
		{"own repo prefix dropped", "sproutling-btv7", "sproutling", "btv7"},
		{"short prefix dropped", "bt-pl5p", "bt", "pl5p"},
		{"foreign prefix kept in full", "lean-stack-58j0", "bt", "lean-stack-58j0"},
		{"empty slug is a no-op", "bt-pl5p", "", "bt-pl5p"},
		{"prefix without separator is not a prefix", "btx-pl5p", "bt", "btx-pl5p"},
		{"suffix must stay non-empty", "bt-", "bt", "bt-"},
		{"multi-segment slug dropped", "lean-stack-58j0", "lean-stack", "58j0"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := shortBeanID(c.id, c.slug); got != c.want {
				t.Fatalf("shortBeanID(%q, %q) = %q, want %q", c.id, c.slug, got, c.want)
			}
		})
	}
}

// TestTreeRowTextDropsRepoPrefix and its Backlog twin below guard the
// Grounding's own "zwei Render-Stellen" pitfall: Tree and Flat/Backlog use
// DIFFERENT row renderers, so shortening only one would let them drift.
func TestTreeRowTextDropsRepoPrefix(t *testing.T) {
	n := treeNode{bean: &data.Bean{ID: "sproutling-btv7", Title: "Go deeper", Status: "todo", Type: "task"}}
	got := stripHint(treeRowText(n, "sproutling"))
	if strings.Contains(got, "sproutling-btv7") {
		t.Fatalf("treeRowText kept the repo prefix: %q", got)
	}
	if !strings.Contains(got, "btv7") {
		t.Fatalf("treeRowText lost the ID suffix: %q", got)
	}
}

func TestBacklogRowTextDropsRepoPrefix(t *testing.T) {
	b := &data.Bean{ID: "sproutling-btv7", Title: "Go deeper", Status: "todo", Type: "task"}
	got := stripHint(backlogRowText(b, "sproutling"))
	if strings.Contains(got, "sproutling-btv7") {
		t.Fatalf("backlogRowText kept the repo prefix: %q", got)
	}
	if !strings.Contains(got, "btv7") {
		t.Fatalf("backlogRowText lost the ID suffix: %q", got)
	}
}

// TestBacklogRowTextKeepsForeignID: a bean whose ID does not carry the
// current repo's prefix is a cross-repo/dangling reference -- shortening it
// would make it ambiguous, so it stays whole.
func TestBacklogRowTextKeepsForeignID(t *testing.T) {
	b := &data.Bean{ID: "lean-stack-58j0", Title: "Foreign", Status: "todo", Type: "task"}
	got := stripHint(backlogRowText(b, "sproutling"))
	if !strings.Contains(got, "lean-stack-58j0") {
		t.Fatalf("backlogRowText shortened a foreign ID: %q", got)
	}
}
