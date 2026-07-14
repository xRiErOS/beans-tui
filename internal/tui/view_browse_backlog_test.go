package tui

// view_browse_backlog_test.go — TDD coverage for Backlog-View V3 `b` (E2 Task
// 5, bean bt-gzu6, design-spec.md §6 V3): idx.Backlog() reuse, the shared
// search+facet predicate (m.beanMatches, Task 3/4), the Sort-Toggle `S`
// cycle, Master-Detail focus reuse (focusedBean(), Task 2's Accordion, Task
// 1), and windowing reuse (E1 windowAround). Reuses fixtureBeans/
// fixtureModel/step/keyMsg/runeMsg/nodeIDs/equalStrings from update_test.go
// (same package).

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"beans-tui/internal/data"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// backlogBeans is a small parentless fixture set covering both Backlog
// membership (parentless, non-milestone/epic, status todo/draft) and
// non-membership (parented, milestone/epic type, or completed/in-progress
// status) -- mirrors idx.Backlog()'s own filter (data/index.go).
func backlogBeans() []data.Bean {
	return []data.Bean{
		{ID: "bk-mlst", Title: "A Milestone", Status: "todo", Type: "milestone", Priority: "normal"},
		{ID: "bk-tsk1", Title: "Backlog Task One", Status: "todo", Type: "task", Priority: "normal"},
		{ID: "bk-tsk2", Title: "Backlog Task Two", Status: "draft", Type: "bug", Priority: "high"},
		{ID: "bk-done", Title: "Done Task", Status: "completed", Type: "task", Priority: "normal"},
		{ID: "bk-child", Title: "Parented Task", Status: "todo", Type: "task", Priority: "normal", Parent: "bk-mlst"},
	}
}

// --- backlogVisible / idx.Backlog() reuse ---

// TestBacklogShowsParentlessReadyBeansFromIndex guards that backlogVisible
// reuses idx.Backlog() (E1 Task 3) unchanged: only parentless, non-milestone/
// epic, todo/draft beans appear.
func TestBacklogShowsParentlessReadyBeansFromIndex(t *testing.T) {
	m := fixtureModel(t, backlogBeans())
	vis := m.backlogVisible()

	var got []string
	for _, b := range vis {
		got = append(got, b.ID)
	}
	want := []string{"bk-tsk1", "bk-tsk2"} // milestone/completed/parented excluded
	if !equalStrings(got, want) {
		t.Fatalf("backlogVisible() ids = %v, want %v", got, want)
	}
}

// TestBacklogVisibleNilIndexReturnsNil guards the pre-load state (m.idx ==
// nil, e.g. Backlog opened before the first beansLoadedMsg arrives) never
// panics.
func TestBacklogVisibleNilIndexReturnsNil(t *testing.T) {
	m := newModel(nil, "/tmp/bt-fixture-repo")
	if vis := m.backlogVisible(); vis != nil {
		t.Fatalf("backlogVisible() on nil idx = %v, want nil", vis)
	}
}

// --- shared search + facet filters ---

// TestBacklogAppliesSharedSearchAndFacetFilters guards that backlogVisible
// calls the SAME m.beanMatches predicate the Tree uses (Task 3/4) -- no
// second, parallel filter implementation.
func TestBacklogAppliesSharedSearchAndFacetFilters(t *testing.T) {
	m := fixtureModel(t, backlogBeans())
	m.searchQuery = "One" // matches only "Backlog Task One"

	vis := m.backlogVisible()
	if len(vis) != 1 || vis[0].ID != "bk-tsk1" {
		t.Fatalf("backlogVisible() with searchQuery=One = %v, want [bk-tsk1]", vis)
	}

	m.searchQuery = ""
	m.filterStatus = map[string]bool{"draft": true}
	vis = m.backlogVisible()
	if len(vis) != 1 || vis[0].ID != "bk-tsk2" {
		t.Fatalf("backlogVisible() with filterStatus=draft = %v, want [bk-tsk2]", vis)
	}
}

// --- Sort-Toggle S ---

// TestBacklogSortCyclesThroughFourModesAndBackToStart presses S 4 times: the
// mode sequence must read status->priority->created->updated->status (bean
// bt-gzu6 Akzeptanz + plan Port-Referenzen), landing back on a visually
// identical order to the untouched default ("" -- idx.Backlog()'s own
// canonical order IS already status-tier order).
//
// ERRATUM vs. the plan's own sketch: the plan's `modes := []string{"",
// "priority", "created", "updated", "status"}` cycled through FIVE distinct
// string states ("" and "status" both members) -- wrapping from "status"
// (index 4) back to "" (index 0) instead of back to "priority" (index 1).
// Since "" and "status" render IDENTICALLY (idx.Backlog() canonical order
// already IS status-tier order), that 5-state cycle produces a DEAD 5th
// keypress every lap (state changes "status"->"", but the rendered order
// does not) -- contradicting the bean's own "4 modes" Akzeptanz wording.
// Fixed here with a 4-element canonical mode list (nextBacklogSort,
// view_browse_backlog.go) that treats "" as an ALIAS for "status" when
// finding the current position, never re-entering "" once cycling starts.
func TestBacklogSortCyclesThroughFourModesAndBackToStart(t *testing.T) {
	beans := []data.Bean{
		{ID: "bk-a", Title: "Zeta", Status: "todo", Type: "task", Priority: "low"},
		{ID: "bk-b", Title: "Alpha", Status: "in-progress", Type: "task", Priority: "critical"},
	}
	m := fixtureModel(t, beans)
	m.view = viewBacklog

	wantSeq := []string{"priority", "created", "updated", "status"}
	for i, want := range wantSeq {
		m = step(t, m, runeMsg('S'))
		if m.backlogSort != want {
			t.Fatalf("press %d: backlogSort = %q, want %q", i+1, m.backlogSort, want)
		}
	}

	// A 5th press must cycle back to "priority" -- NOT re-visit "" (proves
	// the fix: a pure 4-state cycle, no dead keypress).
	m = step(t, m, runeMsg('S'))
	if m.backlogSort != "priority" {
		t.Fatalf("press 5: backlogSort = %q, want %q (period-4 cycle, no dead state)", m.backlogSort, "priority")
	}
}

// TestBacklogSortOrdersChangeVisibly guards that each sort mode actually
// reorders the visible list using data.StatusRank/PriorityRank (Task 1) and
// CreatedAt/UpdatedAt, nil timestamps sorting last without panicking.
func TestBacklogSortOrdersChangeVisibly(t *testing.T) {
	older := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	newer := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	beans := []data.Bean{
		{ID: "bk-a", Title: "A", Status: "todo", Type: "task", Priority: "low", CreatedAt: &older, UpdatedAt: nil},
		{ID: "bk-b", Title: "B", Status: "todo", Type: "task", Priority: "critical", CreatedAt: &newer, UpdatedAt: &newer},
	}
	m := fixtureModel(t, beans)

	m.backlogSort = "priority"
	vis := m.backlogVisible()
	if vis[0].ID != "bk-b" { // critical ranks before low
		t.Fatalf("priority sort first = %s, want bk-b (critical)", vis[0].ID)
	}

	m.backlogSort = "created"
	vis = m.backlogVisible()
	if vis[0].ID != "bk-b" { // newer CreatedAt first
		t.Fatalf("created sort first = %s, want bk-b (newer)", vis[0].ID)
	}

	m.backlogSort = "updated"
	vis = m.backlogVisible() // bk-a has nil UpdatedAt -- must sort last, not panic
	if vis[0].ID != "bk-b" {
		t.Fatalf("updated sort first = %s, want bk-b (nil UpdatedAt sorts last)", vis[0].ID)
	}
}

// --- Master-Detail focus reuse ---

// TestBacklogMasterDetailFocusSwapMatchesD03BorderColors guards that tab
// while m.view == viewBacklog toggles m.detailFocus exactly like the Tree
// view (existing renderPane focus-border machinery, no new border logic).
func TestBacklogMasterDetailFocusSwapMatchesD03BorderColors(t *testing.T) {
	m := fixtureModel(t, backlogBeans())
	m.view = viewBacklog

	if m.detailFocus {
		t.Fatal("detailFocus should start false")
	}
	m = step(t, m, keyMsg(tea.KeyTab))
	if !m.detailFocus {
		t.Fatal("tab should have set detailFocus = true")
	}
	m = step(t, m, keyMsg(tea.KeyTab))
	if m.detailFocus {
		t.Fatal("tab should have set detailFocus = false")
	}
}

// TestBacklogAccordionReusesTask1SectionsViaFocusedBean guards that entering
// detail focus from the Backlog list drives m.focusedBean() to the SELECTED
// backlog bean, not the (irrelevant, possibly stale) tree cursor -- proves
// focusedBean()'s Task-2 dispatcher is genuinely reused, not reimplemented.
func TestBacklogAccordionReusesTask1SectionsViaFocusedBean(t *testing.T) {
	m := fixtureModel(t, backlogBeans())
	m.view = viewBacklog
	m.backlogList.setLen(len(m.backlogVisible()))
	m.backlogList.move(1) // move onto the 2nd backlog row (bk-tsk2)

	want := m.backlogVisible()[m.backlogList.cursor]
	got := m.focusedBean()
	if got == nil || got.ID != want.ID {
		t.Fatalf("focusedBean() = %v, want %s (the Backlog list's selected bean)", got, want.ID)
	}

	m = step(t, m, keyMsg(tea.KeyEnter))
	if !m.detailFocus {
		t.Fatal("enter on a focusable backlog row must enter detail focus")
	}
}

// --- b open/close, shared filter state preserved ---

// TestBacklogOpenFromTreeAndBackPreservesSharedFilterState guards that `b`
// from Tree with an active facet filter shows the SAME filtered view in
// Backlog, and esc/b back to Tree leaves the filter active there too --
// ONE shared filter state (Task 4), never reset by a view switch.
func TestBacklogOpenFromTreeAndBackPreservesSharedFilterState(t *testing.T) {
	m := fixtureModel(t, backlogBeans())
	m.filterStatus = map[string]bool{"draft": true}

	m = step(t, m, runeMsg('b'))
	if m.view != viewBacklog {
		t.Fatalf("b should open Backlog, view = %v", m.view)
	}
	vis := m.backlogVisible()
	if len(vis) != 1 || vis[0].ID != "bk-tsk2" {
		t.Fatalf("Backlog with filterStatus=draft carried over = %v, want [bk-tsk2]", vis)
	}

	m = step(t, m, runeMsg('b'))
	if m.view != viewBrowseRepo {
		t.Fatalf("b again should return to Tree, view = %v", m.view)
	}
	if !m.filterStatus["draft"] {
		t.Fatal("filterStatus must still be active back in Tree -- shared state, not reset by the view switch")
	}
}

// TestBacklogEscReturnsToTree guards the esc half of the "b open, esc/b
// back" contract (design-spec.md, bean bt-gzu6 Akzeptanz).
func TestBacklogEscReturnsToTree(t *testing.T) {
	m := fixtureModel(t, backlogBeans())
	m = step(t, m, runeMsg('b'))
	if m.view != viewBacklog {
		t.Fatalf("b should open Backlog, view = %v", m.view)
	}
	m = step(t, m, keyMsg(tea.KeyEsc))
	if m.view != viewBrowseRepo {
		t.Fatalf("esc should return to Tree, view = %v", m.view)
	}
}

// --- windowing reuse ---

// TestBacklogWindowingReusesExistingWindowAround guards that a Backlog
// longer than the pane's row budget keeps the cursor visible via the
// existing windowAround/windowStart (E1 Task 8) -- no new fenestration
// mechanism.
func TestBacklogWindowingReusesExistingWindowAround(t *testing.T) {
	var beans []data.Bean
	for i := 0; i < 40; i++ {
		beans = append(beans, data.Bean{
			ID: "bk-w" + string(rune('a'+i)), Title: "Row", Status: "todo", Type: "task", Priority: "normal",
		})
	}
	m := fixtureModel(t, beans)
	m.view = viewBacklog
	m.width, m.height = 100, 20 // small enough that 40 rows won't all fit
	m.backlogList.setLen(len(m.backlogVisible()))
	m.backlogList.cursor = 35 // near the end

	out := m.viewBacklog()
	if !strings.Contains(out, "bk-w"+string(rune('a'+35))) {
		t.Fatalf("cursor row (index 35) must stay visible via windowAround, output:\n%s", out)
	}
}

// --- Golden ---

// goldenBacklogModel builds the deterministic fixture rendered by
// TestBacklogGolden: two Backlog-eligible beans (todo task, draft bug -- both
// row kinds on screen at once) plus a milestone AND a parented task (both
// excluded from idx.Backlog(), proving the view really is the backlog, not
// the tree), cursor parked on the second row, Backlog view active.
func goldenBacklogModel(t *testing.T) model {
	t.Helper()
	beans := []data.Bean{
		{ID: "gbk-mlst", Title: "Golden Milestone", Status: "in-progress", Type: "milestone", Priority: "high"},
		{ID: "gbk-tsk1", Title: "First backlog task", Status: "todo", Type: "task", Priority: "normal"},
		{ID: "gbk-tsk2", Title: "Second backlog task", Status: "draft", Type: "bug", Priority: "critical", Tags: []string{"backend"}},
		{ID: "gbk-child", Title: "Parented, excluded", Status: "todo", Type: "task", Priority: "normal", Parent: "gbk-mlst"},
	}

	m := newModel(nil, "/tmp/bt-golden-repo")
	m = step(t, m, beansLoadedMsg{beans: beans})
	m = step(t, m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m.view = viewBacklog
	m.backlogList.setLen(len(m.backlogVisible()))
	m.backlogList.cursor = 1 // gbk-tsk2
	return m
}

// TestBacklogGolden renders a full 100x30 frame of the Backlog view and
// compares it against testdata/backlog.golden.
// Regenerate with: go test ./internal/tui/ -run TestBacklogGolden -update
func TestBacklogGolden(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(termenv.Ascii)

	m := goldenBacklogModel(t)
	out := m.View()

	if h := lipgloss.Height(out); h != 30 {
		t.Errorf("backlog view height=%d, want 30 (full terminal height)", h)
	}
	for i, ln := range strings.Split(out, "\n") {
		if w := lipgloss.Width(ln); w > 100 {
			t.Errorf("line %d overflows width (%d > 100): %q", i, w, ln)
		}
	}

	path := filepath.Join("testdata", "backlog.golden")
	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(out), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("golden missing (%s) — regenerate with -update: %v", path, err)
	}
	if out != string(want) {
		t.Errorf("backlog view output differs from golden %q (frame/width/truncation?).\n--- got ---\n%s\n--- want ---\n%s", path, out, string(want))
	}
}

// TestBacklogGoldenDeterministic guards that View() is a pure function of
// the model for the Backlog view too: repeated calls on identical state must
// byte-for-byte agree.
func TestBacklogGoldenDeterministic(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(termenv.Ascii)

	m := goldenBacklogModel(t)
	a := m.View()
	b := m.View()
	if a != b {
		t.Error("View() is not deterministic across repeated calls with identical Backlog model state")
	}
}
