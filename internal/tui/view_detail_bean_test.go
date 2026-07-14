package tui

// view_detail_bean_test.go — TDD-first tests for beanSections (E2 Task 1,
// bean bt-ms0k): the 4 fixed accordion sections (Meta/Body/Beziehungen/
// Historie) built from a *data.Bean + its *data.Index.

import (
	"strings"
	"testing"
	"time"

	"beans-tui/internal/data"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// TestBeanSectionsAlwaysFourFixedSections guards E2's simpler digit-jump
// semantics vs. devd's content-gated sections: a minimal bean (no body, no
// tags, no relations) still yields exactly 4 sections, titled and ordered
// Meta/Body/Beziehungen/Historie.
func TestBeanSectionsAlwaysFourFixedSections(t *testing.T) {
	beans := []data.Bean{
		{ID: "min-1", Title: "Minimal", Status: "todo", Type: "task", Priority: "normal"},
	}
	idx := data.NewIndex(beans)
	secs := beanSections(idx, idx.ByID["min-1"], 60)

	if len(secs) != 4 {
		t.Fatalf("beanSections returned %d sections, want 4", len(secs))
	}
	want := []string{"Meta", "Body", "Beziehungen", "Historie"}
	for i, w := range want {
		if secs[i].title != w {
			t.Errorf("section %d title = %q, want %q", i, secs[i].title, w)
		}
	}
}

// TestBeanSectionsMetaRendersStatusTypePriorityTags guards the Meta section's
// content: it must reuse render_shared.go's themed status/type/priority/tag
// rendering (no new theme code), so the raw values show up in the body.
func TestBeanSectionsMetaRendersStatusTypePriorityTags(t *testing.T) {
	beans := []data.Bean{
		{ID: "meta-1", Title: "Meta Bean", Status: "in-progress", Type: "bug", Priority: "critical", Tags: []string{"urgent"}},
	}
	idx := data.NewIndex(beans)
	secs := beanSections(idx, idx.ByID["meta-1"], 60)

	meta := secs[0].body
	if !strings.Contains(meta, "in-progress") {
		t.Errorf("Meta body missing status: %q", meta)
	}
	if !strings.Contains(meta, "bug") {
		t.Errorf("Meta body missing type: %q", meta)
	}
	if !strings.Contains(meta, "critical") {
		t.Errorf("Meta body missing priority: %q", meta)
	}
	if !strings.Contains(meta, "urgent") {
		t.Errorf("Meta body missing tag: %q", meta)
	}
}

// TestBeanSectionsBodyEmptyShowsPlaceholder guards the empty-body path: no
// Body field set -> a placeholder, not an empty string (digit-jump 2 must
// always show something).
func TestBeanSectionsBodyEmptyShowsPlaceholder(t *testing.T) {
	beans := []data.Bean{
		{ID: "nobody-1", Title: "No Body", Status: "todo", Type: "task", Priority: "normal"},
	}
	idx := data.NewIndex(beans)
	secs := beanSections(idx, idx.ByID["nobody-1"], 60)

	if strings.TrimSpace(secs[1].body) == "" {
		t.Fatal("Body section must show a placeholder for an empty bean body, got empty string")
	}
}

// TestBeanSectionsBodyRendersMarkdown guards the glamour wiring: markdown
// heading/bold syntax survives as rendered ANSI under TrueColor.
func TestBeanSectionsBodyRendersMarkdown(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(termenv.Ascii)

	beans := []data.Bean{
		{ID: "body-1", Title: "Has Body", Status: "todo", Type: "task", Priority: "normal", Body: "# Heading\n\nSome **bold** text."},
	}
	idx := data.NewIndex(beans)
	secs := beanSections(idx, idx.ByID["body-1"], 60)

	if !strings.Contains(secs[1].body, "\x1b[") {
		t.Errorf("Body section under TrueColor must contain ANSI escapes (glamour-rendered), got: %q", secs[1].body)
	}
	if !strings.Contains(secs[1].body, "Heading") {
		t.Errorf("Body section missing rendered heading text: %q", secs[1].body)
	}
}

// TestBeanSectionsBeziehungenListsParentChildrenBlockingBlockedByInCanonicalOrder
// guards I03: Blocking/BlockedBy resolved beans must render in data.SortBeans
// canonical tier order (Status -> Priority -> Type -> Title), NOT insertion
// order. Children arrives pre-sorted via idx.Children; this test forces
// Blocking to disagree with insertion order (in-progress bean listed second
// in the raw slice, must still render first).
func TestBeanSectionsBeziehungenListsParentChildrenBlockingBlockedByInCanonicalOrder(t *testing.T) {
	beans := []data.Bean{
		{ID: "rel-parent", Title: "Parent Bean", Status: "todo", Type: "epic", Priority: "normal"},
		{ID: "rel-main", Title: "Main Bean", Status: "todo", Type: "task", Priority: "normal", Parent: "rel-parent",
			Blocking: []string{"rel-block-todo", "rel-block-inprog"}},
		{ID: "rel-child", Title: "Child Bean", Status: "todo", Type: "task", Priority: "normal", Parent: "rel-main"},
		// Blocking targets: listed todo-first in Blocking slice above, but
		// in-progress must sort first per data.SortBeans canonical order.
		{ID: "rel-block-todo", Title: "Blocked Todo", Status: "todo", Type: "task", Priority: "normal"},
		{ID: "rel-block-inprog", Title: "Blocked InProgress", Status: "in-progress", Type: "task", Priority: "normal"},
	}
	idx := data.NewIndex(beans)
	secs := beanSections(idx, idx.ByID["rel-main"], 60)

	rel := secs[2]
	if rel.title != "Beziehungen" {
		t.Fatalf("section 2 title = %q, want Beziehungen", rel.title)
	}
	if !strings.Contains(rel.body, "Parent Bean") {
		t.Errorf("Beziehungen body missing Parent: %q", rel.body)
	}
	if !strings.Contains(rel.body, "Child Bean") {
		t.Errorf("Beziehungen body missing Children: %q", rel.body)
	}

	// Find the relationField order for the two Blocking targets among
	// rel.fields (Parent is 1 field, Children 1 field, then Blocking x2).
	var blockingIDs []string
	for _, f := range rel.fields {
		if f.beanID == "rel-block-todo" || f.beanID == "rel-block-inprog" {
			blockingIDs = append(blockingIDs, f.beanID)
		}
	}
	if len(blockingIDs) != 2 {
		t.Fatalf("expected 2 Blocking relationFields, got %v", blockingIDs)
	}
	if blockingIDs[0] != "rel-block-inprog" {
		t.Errorf("Blocking field order = %v, want rel-block-inprog first (canonical status tier, not insertion order)", blockingIDs)
	}
}

// TestBeanSectionsBeziehungenDanglingReferenceShowsUnresolvedNotJumpable
// guards the dangling-BlockedBy path: an ID with no match in idx.ByID renders
// as a muted "(unresolved: <id>)" row with relationField.beanID == "" (Task
// 2's jump-guard checks this).
func TestBeanSectionsBeziehungenDanglingReferenceShowsUnresolvedNotJumpable(t *testing.T) {
	beans := []data.Bean{
		{ID: "dangling-1", Title: "Dangling Bean", Status: "todo", Type: "task", Priority: "normal",
			BlockedBy: []string{"does-not-exist"}},
	}
	idx := data.NewIndex(beans)
	secs := beanSections(idx, idx.ByID["dangling-1"], 60)

	rel := secs[2]
	if !strings.Contains(rel.body, "does-not-exist") {
		t.Errorf("Beziehungen body missing the unresolved reference id: %q", rel.body)
	}
	found := false
	for _, f := range rel.fields {
		if strings.Contains(f.label, "does-not-exist") {
			found = true
			if f.beanID != "" {
				t.Errorf("unresolved relationField.beanID = %q, want empty (not jumpable)", f.beanID)
			}
		}
	}
	if !found {
		t.Fatal("expected a relationField for the unresolved BlockedBy reference")
	}
}

// TestBeanSectionsHistorieShowsCreatedUpdatedETag guards the Historie
// section: nil CreatedAt/UpdatedAt render a placeholder, never a zero-time
// string ("0001-01-01...").
func TestBeanSectionsHistorieShowsCreatedUpdatedETag(t *testing.T) {
	beans := []data.Bean{
		{ID: "hist-nil", Title: "No Timestamps", Status: "todo", Type: "task", Priority: "normal"},
	}
	idx := data.NewIndex(beans)
	secs := beanSections(idx, idx.ByID["hist-nil"], 60)

	histNil := secs[3].body
	if strings.Contains(histNil, "0001-01-01") {
		t.Errorf("Historie must not render Go's zero-time string for nil timestamps: %q", histNil)
	}

	created := time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC)
	updated := time.Date(2026, 1, 3, 11, 0, 0, 0, time.UTC)
	beansSet := []data.Bean{
		{ID: "hist-set", Title: "With Timestamps", Status: "todo", Type: "task", Priority: "normal",
			CreatedAt: &created, UpdatedAt: &updated, ETag: "abc123"},
	}
	idx2 := data.NewIndex(beansSet)
	secs2 := beanSections(idx2, idx2.ByID["hist-set"], 60)
	hist := secs2[3].body
	if !strings.Contains(hist, "2026-01-02") {
		t.Errorf("Historie missing CreatedAt: %q", hist)
	}
	if !strings.Contains(hist, "2026-01-03") {
		t.Errorf("Historie missing UpdatedAt: %q", hist)
	}
	if !strings.Contains(hist, "abc123") {
		t.Errorf("Historie missing ETag: %q", hist)
	}
}
