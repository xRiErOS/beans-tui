package tui

// view_detail_bean_test.go — TDD-first tests for beanSections (E2 Task 1,
// bean bt-ms0k): the 4 fixed accordion sections built from a *data.Bean + its
// *data.Index. E7 Task 4 (bean bt-kyj5, PF-1/PF-3/PF-4/PF-12) rebuilds this
// around the new Meta-Feldliste + Kopfblock: section titles go uppercase-
// English (META/BODY/RELATIONS/HISTORY, PF-7's remainder), beanSections
// grows 3 render-state params (focused, activeIdx, fieldIdx -- Meta's body
// now needs to know whether IT is the active section to place the ▶ marker),
// and Meta's body becomes a 6-entry cursor-navigable field list instead of a
// 4-line status/type/priority/tags summary (tags are no longer part of Meta
// at all -- PF-4 fixes the field set to title/status/type/priority/
// created_at/updated_at).

import (
	"strings"
	"testing"
	"time"

	"beans-tui/internal/data"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
)

// TestBeanSectionsAlwaysFourFixedSections guards E2's simpler digit-jump
// semantics vs. devd's content-gated sections: a minimal bean (no body, no
// tags, no relations) still yields exactly 4 sections, titled and ordered
// META/BODY/RELATIONS/HISTORY (PF-7, uppercase English -- T3 deliberately
// left these 4 titles to this task, see bt-w9o8 "Notes for T4").
func TestBeanSectionsAlwaysFourFixedSections(t *testing.T) {
	beans := []data.Bean{
		{ID: "min-1", Title: "Minimal", Status: "todo", Type: "task", Priority: "normal"},
	}
	idx := data.NewIndex(beans)
	secs := beanSections(idx, idx.ByID["min-1"], 60, false, 0, 0)

	if len(secs) != 4 {
		t.Fatalf("beanSections returned %d sections, want 4", len(secs))
	}
	want := []string{"META", "BODY", "RELATIONS", "HISTORY"}
	for i, w := range want {
		if secs[i].title != w {
			t.Errorf("section %d title = %q, want %q", i, secs[i].title, w)
		}
	}
}

// TestMetaFieldsSixEntriesWithKinds guards metaFields' structural contract
// (PF-4): exactly 6 kind-tagged entries in a fixed order (title/status/type/
// priority/created_at/updated_at), every entry's beanID empty (Meta fields
// are never Beziehungen-style jump targets -- kind drives T6's future enter-
// dispatch instead). Supersedes the old
// TestBeanSectionsMetaRendersStatusTypePriorityTags (Tags is no longer part
// of Meta at all under PF-4's fixed 6-field layout).
func TestMetaFieldsSixEntriesWithKinds(t *testing.T) {
	b := &data.Bean{ID: "meta-1", Title: "Meta Bean", Status: "in-progress", Type: "bug", Priority: "critical"}
	fields := metaFields(b)
	if len(fields) != 6 {
		t.Fatalf("metaFields returned %d entries, want 6", len(fields))
	}
	wantKinds := []string{"title", "status", "type", "priority", "readonly", "readonly"}
	for i, want := range wantKinds {
		if fields[i].kind != want {
			t.Errorf("fields[%d].kind = %q, want %q", i, fields[i].kind, want)
		}
		if fields[i].beanID != "" {
			t.Errorf("fields[%d].beanID = %q, want empty (Meta fields are never jump targets)", i, fields[i].beanID)
		}
	}
	if fields[0].label != "Meta Bean" {
		t.Errorf("fields[0] (title) label = %q, want %q", fields[0].label, "Meta Bean")
	}
	// status/type/priority reuse T2's GLYPH output (PF-6 StatusIcon/TypeIcon/
	// Priority), NOT the word -- design-spec.md §15 PF-4 explicit ("nutzt T2s
	// Glyph-Output"), unlike the Kopfblock's type/status (see
	// TestDetailHeaderBlockShowsIDTitleTypeStatusPrio, which uses the WORD).
	if got := ansi.Strip(fields[1].label); got != "i" {
		t.Errorf("fields[1] (status) label = %q, want glyph \"i\" (PF-6 StatusIcon(\"in-progress\"))", got)
	}
	if got := ansi.Strip(fields[2].label); got != "B" {
		t.Errorf("fields[2] (type) label = %q, want glyph \"B\" (PF-6 TypeIcon(\"bug\"))", got)
	}
	if got := ansi.Strip(fields[3].label); got != "‼" {
		t.Errorf("fields[3] (priority) label = %q, want glyph \"‼\" (PF-6 Priority(\"critical\"))", got)
	}
}

// TestMetaFieldsPriorityDefaultsToNormalWhenEmpty guards the same "" ->
// "normal" default the old metaSectionBody used to apply (b.Priority is
// beans-legal empty on hand-edited frontmatter).
func TestMetaFieldsPriorityDefaultsToNormalWhenEmpty(t *testing.T) {
	b := &data.Bean{ID: "meta-noprio", Title: "No Priority", Status: "todo", Type: "task"}
	fields := metaFields(b)
	if got := ansi.Strip(fields[3].label); got != "·" {
		t.Errorf("fields[3] (priority) label = %q, want the \"normal\" glyph \"·\" for empty Priority", got)
	}
}

// TestMetaFieldsReadonlyTimestampsUseFmtTime guards created_at/updated_at:
// nil renders the same muted placeholder historieSectionBody uses (shared
// fmtTime helper, not a re-implementation), a set timestamp renders
// formatted.
func TestMetaFieldsReadonlyTimestampsUseFmtTime(t *testing.T) {
	created := time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC)
	b := &data.Bean{ID: "meta-ts", Title: "T", Status: "todo", Type: "task", Priority: "normal", CreatedAt: &created}
	fields := metaFields(b)
	if !strings.Contains(fields[4].label, "2026-01-02") {
		t.Errorf("fields[4] (created_at) label = %q, want formatted CreatedAt", fields[4].label)
	}
	if !strings.Contains(ansi.Strip(fields[5].label), "unknown") {
		t.Errorf("fields[5] (updated_at) label = %q, want the nil-timestamp placeholder", fields[5].label)
	}
}

// TestMetaSectionBodyShowsSelectedFieldMarker guards PF-4's ▷/▶ cursor +
// PF-12's gutter contract at the same time: exactly one ▶ at the active
// fieldIdx's row when the Meta section itself is active, ▷ on all 6 rows
// (never an omitted marker) when it is not.
func TestMetaSectionBodyShowsSelectedFieldMarker(t *testing.T) {
	b := &data.Bean{ID: "meta-mark", Title: "Marked", Status: "todo", Type: "task", Priority: "high"}

	active := metaSectionBody(b, 80, true, 2) // priority row (index 2)
	lines := strings.Split(active, "\n")
	if len(lines) != 6 {
		t.Fatalf("metaSectionBody active=true has %d lines, want 6", len(lines))
	}
	if n := strings.Count(ansi.Strip(active), "▶"); n != 1 {
		t.Fatalf("active metaSectionBody has %d ▶ markers, want exactly 1", n)
	}
	if !strings.Contains(ansi.Strip(lines[2]), "▶") {
		t.Errorf("▶ marker not on the priority row (index 2): %q", ansi.Strip(lines[2]))
	}
	if n := strings.Count(ansi.Strip(active), "▷"); n != 5 {
		t.Errorf("active metaSectionBody has %d ▷ markers, want exactly 5 (the other rows): %q", n, ansi.Strip(active))
	}

	inactive := metaSectionBody(b, 80, false, 2)
	if strings.Contains(ansi.Strip(inactive), "▶") {
		t.Error("inactive metaSectionBody must show no ▶ marker anywhere")
	}
	if n := strings.Count(ansi.Strip(inactive), "▷"); n != 6 {
		t.Errorf("inactive metaSectionBody has %d ▷ markers, want exactly 6 (every row): %q", n, ansi.Strip(inactive))
	}
}

// TestMetaSectionBodyShowsAllSixLabelsAndValues guards the row content
// itself (label prefixes + values), not just the marker: PF-4's mockup shows
// title/status/type/priority/created_at/updated_at in that exact order.
func TestMetaSectionBodyShowsAllSixLabelsAndValues(t *testing.T) {
	created := time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC)
	b := &data.Bean{ID: "meta-full", Title: "Full Bean", Status: "todo", Type: "task", Priority: "normal", CreatedAt: &created}
	body := ansi.Strip(metaSectionBody(b, 80, false, 0))

	wantLabels := []string{"title:", "status:", "type:", "priority:", "created_at:", "updated_at:"}
	lastIdx := -1
	for _, label := range wantLabels {
		idx := strings.Index(body, label)
		if idx < 0 {
			t.Fatalf("metaSectionBody missing label %q: %q", label, body)
		}
		if idx <= lastIdx {
			t.Errorf("label %q out of order (idx %d, previous %d)", label, idx, lastIdx)
		}
		lastIdx = idx
	}
	if !strings.Contains(body, "Full Bean") {
		t.Errorf("metaSectionBody missing title value: %q", body)
	}
	if !strings.Contains(body, "2026-01-02") {
		t.Errorf("metaSectionBody missing created_at value: %q", body)
	}
}

// TestDetailHeaderBlockShowsIDTitleTypeStatusPrio guards the NEW Kopfblock
// (PF-3+PF-4, verschmolzen laut PO-Antwort Q01): bean-id / NAME / blank /
// "type: X    status: Y    prio: Z" / trailing blank. type/status render as
// the WORD (theme.TypeStyle/StatusStyle + the raw word, mirrors the OLD
// metaSectionBody's word rendering) -- prio renders as the GLYPH (theme.
// Priority has no word-producing counterpart since PF-6, epic-E7-plan.md
// Task 4 signature block is explicit: "...+Priority", not "+PriorityStyle").
func TestDetailHeaderBlockShowsIDTitleTypeStatusPrio(t *testing.T) {
	b := &data.Bean{ID: "hdr-1", Title: "Header Bean", Status: "in-progress", Type: "bug", Priority: "high"}
	out := detailHeaderBlock(b, 60)
	stripped := ansi.Strip(out)

	if !strings.Contains(stripped, "hdr-1") {
		t.Errorf("header block missing bean ID: %q", stripped)
	}
	if !strings.Contains(stripped, "Header Bean") {
		t.Errorf("header block missing title: %q", stripped)
	}
	if !strings.Contains(stripped, "type: bug") {
		t.Errorf("header block missing type as WORD (\"type: bug\"): %q", stripped)
	}
	if !strings.Contains(stripped, "status: in-progress") {
		t.Errorf("header block missing status as WORD (\"status: in-progress\"): %q", stripped)
	}
	if !strings.Contains(stripped, "prio: !") {
		t.Errorf("header block missing prio as GLYPH (\"prio: !\", high): %q", stripped)
	}

	lines := strings.Split(out, "\n")
	if len(lines) < 5 {
		t.Fatalf("header block has %d lines, want at least 5 (ID/Title/blank/type-status-prio/trailing blank)", len(lines))
	}
	if lines[len(lines)-1] != "" {
		t.Errorf("header block must end with a trailing blank line, got %q as last line", lines[len(lines)-1])
	}
}

// TestDetailHeaderBlockTruncatesToWidth guards that every content line is
// clamped to w (no App-Chrome wrap concept here, single-line-per-field
// mirrors the ID/type-status-prio mockup) -- an overlong title must not blow
// out the accordion pane's bordered width.
func TestDetailHeaderBlockTruncatesToWidth(t *testing.T) {
	b := &data.Bean{ID: "hdr-long", Title: strings.Repeat("very-long-title-word ", 10), Status: "todo", Type: "task", Priority: "normal"}
	const w = 24
	for _, line := range strings.Split(ansi.Strip(detailHeaderBlock(b, w)), "\n") {
		if width := lipgloss.Width(line); width > w {
			t.Errorf("header block line exceeds w=%d (got %d): %q", w, width, line)
		}
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
	secs := beanSections(idx, idx.ByID["nobody-1"], 60, false, 0, 0)

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
	secs := beanSections(idx, idx.ByID["body-1"], 60, false, 0, 0)

	if !strings.Contains(secs[1].body, "\x1b[") {
		t.Errorf("Body section under TrueColor must contain ANSI escapes (glamour-rendered), got: %q", secs[1].body)
	}
	if !strings.Contains(secs[1].body, "Heading") {
		t.Errorf("Body section missing rendered heading text: %q", secs[1].body)
	}
}

// TestBeanSectionsRelationsListsParentChildrenBlockingBlockedByInCanonicalOrder
// guards I03: Blocking/BlockedBy resolved beans must render in data.SortBeans
// canonical tier order (Status -> Priority -> Type -> Title), NOT insertion
// order. Children arrives pre-sorted via idx.Children; this test forces
// Blocking to disagree with insertion order (in-progress bean listed second
// in the raw slice, must still render first). Renamed from ...Beziehungen...
// (PF-7, section title RELATIONS).
func TestBeanSectionsRelationsListsParentChildrenBlockingBlockedByInCanonicalOrder(t *testing.T) {
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
	secs := beanSections(idx, idx.ByID["rel-main"], 60, false, 0, 0)

	rel := secs[2]
	if rel.title != "RELATIONS" {
		t.Fatalf("section 2 title = %q, want RELATIONS", rel.title)
	}
	if !strings.Contains(rel.body, "Parent Bean") {
		t.Errorf("Relations body missing Parent: %q", rel.body)
	}
	if !strings.Contains(rel.body, "Child Bean") {
		t.Errorf("Relations body missing Children: %q", rel.body)
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

// TestBeanSectionsRelationsDanglingReferenceShowsUnresolvedNotJumpable
// guards the dangling-BlockedBy path: an ID with no match in idx.ByID renders
// as a muted "(unresolved: <id>)" row with relationField.beanID == "" (Task
// 2's jump-guard checks this). Renamed from ...Beziehungen... (PF-7).
func TestBeanSectionsRelationsDanglingReferenceShowsUnresolvedNotJumpable(t *testing.T) {
	beans := []data.Bean{
		{ID: "dangling-1", Title: "Dangling Bean", Status: "todo", Type: "task", Priority: "normal",
			BlockedBy: []string{"does-not-exist"}},
	}
	idx := data.NewIndex(beans)
	secs := beanSections(idx, idx.ByID["dangling-1"], 60, false, 0, 0)

	rel := secs[2]
	if !strings.Contains(rel.body, "does-not-exist") {
		t.Errorf("Relations body missing the unresolved reference id: %q", rel.body)
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

// TestBeanSectionsMetaWrapsLongTitleToBodyWidth guards B01 (E2-T1-quality-
// review finding, MANDATORY carry-over)'s wrapping requirement against the
// PF-4 Meta field list: Tags is no longer part of Meta at all (the old test
// exercised long tag names), so the overflow risk now sits on the title row
// -- an arbitrarily long bean title must still wrap inside bodyW, not
// overflow the Detail pane's bordered width.
func TestBeanSectionsMetaWrapsLongTitleToBodyWidth(t *testing.T) {
	beans := []data.Bean{
		{ID: "wrap-meta", Title: strings.Repeat("a-very-long-title-word ", 6), Status: "todo", Type: "task", Priority: "normal"},
	}
	idx := data.NewIndex(beans)
	const bodyW = 20
	secs := beanSections(idx, idx.ByID["wrap-meta"], bodyW, false, 0, 0)

	for _, line := range strings.Split(ansi.Strip(secs[0].body), "\n") {
		if w := lipgloss.Width(line); w > bodyW {
			t.Errorf("Meta line exceeds bodyW=%d (got %d): %q", bodyW, w, line)
		}
	}
}

// TestBeanSectionsHistorieShowsCreatedUpdatedETag guards the History
// section: nil CreatedAt/UpdatedAt render a placeholder, never a zero-time
// string ("0001-01-01..."). Section title assertion updated to HISTORY
// (PF-7).
func TestBeanSectionsHistorieShowsCreatedUpdatedETag(t *testing.T) {
	beans := []data.Bean{
		{ID: "hist-nil", Title: "No Timestamps", Status: "todo", Type: "task", Priority: "normal"},
	}
	idx := data.NewIndex(beans)
	secs := beanSections(idx, idx.ByID["hist-nil"], 60, false, 0, 0)

	if secs[3].title != "HISTORY" {
		t.Fatalf("section 3 title = %q, want HISTORY", secs[3].title)
	}

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
	secs2 := beanSections(idx2, idx2.ByID["hist-set"], 60, false, 0, 0)
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

// TestBeanSectionsHistorieWrapsLongETagToBodyWidth guards B01 (E2-T1-quality-
// review finding, MANDATORY carry-over into Task 2): historieSectionBody must
// wrap via wrapText -- otherwise a long ETag overflows the Detail pane's
// bordered width.
func TestBeanSectionsHistorieWrapsLongETagToBodyWidth(t *testing.T) {
	beans := []data.Bean{
		{ID: "wrap-hist", Title: "Wrap Hist", Status: "todo", Type: "task", Priority: "normal",
			ETag: strings.Repeat("etagchunk-", 5)},
	}
	idx := data.NewIndex(beans)
	const bodyW = 20
	secs := beanSections(idx, idx.ByID["wrap-hist"], bodyW, false, 0, 0)

	for _, line := range strings.Split(ansi.Strip(secs[3].body), "\n") {
		if w := lipgloss.Width(line); w > bodyW {
			t.Errorf("Historie line exceeds bodyW=%d (got %d): %q", bodyW, w, line)
		}
	}
}
