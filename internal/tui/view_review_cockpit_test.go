package tui

// view_review_cockpit_test.go — TDD coverage for the Review-Queue-Ableitung +
// Cockpit-Skeleton (E4 Task 3, bean bt-hxyo, epic-E4-plan.md »Task 3«).
// Verdikt-actions (a/x/o) land in Task 4 -- this file covers navigation,
// grouping, and the read-only render only.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"beans-tui/internal/data"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// reviewFixtureBeans builds 2 Epics with to-review children, 1 parentless
// to-review bean ((kein Epic) bucket), and 1 rework bean -- the exact shape
// TestReviewQueueGroupsByEpicCanonicalOrder / the golden fixture need.
func reviewFixtureBeans() []data.Bean {
	return []data.Bean{
		{ID: "ms-1", Title: "Milestone", Status: "todo", Type: "milestone", Priority: "normal"},
		{ID: "ep-a", Title: "Epic Alpha", Status: "todo", Type: "epic", Priority: "normal", Parent: "ms-1"},
		{ID: "ep-b", Title: "Epic Beta", Status: "todo", Type: "epic", Priority: "normal", Parent: "ms-1"},
		{ID: "tk-a1", Title: "Alpha Task One", Status: "in-progress", Type: "task", Priority: "normal", Parent: "ep-a", Tags: []string{"to-review"}},
		{ID: "tk-a2", Title: "Alpha Task Two", Status: "in-progress", Type: "task", Priority: "high", Parent: "ep-a", Tags: []string{"to-review"}},
		{ID: "tk-b1", Title: "Beta Task One", Status: "in-progress", Type: "task", Priority: "normal", Parent: "ep-b", Tags: []string{"to-review"}},
		{ID: "tk-none", Title: "Parentless Review Task", Status: "in-progress", Type: "task", Priority: "normal", Tags: []string{"to-review"}},
		{ID: "tk-rw", Title: "Rework Task", Status: "in-progress", Type: "task", Priority: "normal", Parent: "ep-a", Tags: []string{"rework"}},
	}
}

func TestReviewQueueGroupsByEpicCanonicalOrder(t *testing.T) {
	idx := data.NewIndex(reviewFixtureBeans())
	groups := reviewQueue(idx)

	if len(groups) != 3 {
		t.Fatalf("len(reviewQueue(idx)) = %d, want 3 (ep-a, ep-b, (kein Epic))", len(groups))
	}
	// "(kein Epic)" is ALWAYS the last group (design decision c).
	last := groups[len(groups)-1]
	if last.epic != nil {
		t.Fatalf("last group epic = %v, want nil ((kein Epic) bucket)", last.epic)
	}
	if len(last.beans) != 1 || last.beans[0].ID != "tk-none" {
		t.Fatalf("(kein Epic) bucket = %v, want [tk-none]", last.beans)
	}
	for _, g := range groups[:len(groups)-1] {
		if g.epic == nil {
			t.Fatalf("non-last group has nil epic: %v", g)
		}
	}
	// Epic groups themselves are canonically sorted (data.SortBeans on the
	// epic beans) -- both ep-a/ep-b share status/priority/type, so title
	// ("Epic Alpha" < "Epic Beta") decides the order.
	if groups[0].epic.ID != "ep-a" || groups[1].epic.ID != "ep-b" {
		t.Fatalf("epic group order = [%s, %s], want [ep-a, ep-b]", groups[0].epic.ID, groups[1].epic.ID)
	}
}

func TestReviewQueueGroupMembersSortedCanonically(t *testing.T) {
	idx := data.NewIndex(reviewFixtureBeans())
	groups := reviewQueue(idx)

	var epA reviewGroup
	for _, g := range groups {
		if g.epic != nil && g.epic.ID == "ep-a" {
			epA = g
		}
	}
	if len(epA.beans) != 2 {
		t.Fatalf("ep-a group has %d beans, want 2", len(epA.beans))
	}
	// tk-a2 (priority high) sorts before tk-a1 (priority normal) --
	// data.SortBeans' Status->Priority->Type->Title tiering.
	if epA.beans[0].ID != "tk-a2" || epA.beans[1].ID != "tk-a1" {
		t.Fatalf("ep-a group order = [%s, %s], want [tk-a2, tk-a1] (priority tiering)", epA.beans[0].ID, epA.beans[1].ID)
	}
}

func TestReviewQueueEmptyWhenNoToReviewTag(t *testing.T) {
	beans := []data.Bean{
		{ID: "ms-1", Title: "Milestone", Type: "milestone"},
		{ID: "ep-a", Title: "Epic", Type: "epic", Parent: "ms-1"},
		{ID: "tk-1", Title: "Task", Type: "task", Parent: "ep-a"},
	}
	idx := data.NewIndex(beans)
	if got := reviewQueue(idx); len(got) != 0 {
		t.Fatalf("reviewQueue(idx) with no to-review tag = %v, want empty", got)
	}
}

func TestReviewReworkFlatSortedNoGrouping(t *testing.T) {
	idx := data.NewIndex(reviewFixtureBeans())
	rework := reviewRework(idx)
	if len(rework) != 1 || rework[0].ID != "tk-rw" {
		t.Fatalf("reviewRework(idx) = %v, want [tk-rw]", rework)
	}
}

func TestReviewFlatOrderToReviewThenRework(t *testing.T) {
	idx := data.NewIndex(reviewFixtureBeans())
	flat := reviewFlat(idx)

	if len(flat) != 5 { // 3 to-review in ep-a/ep-b + 1 (kein Epic) + 1 rework
		t.Fatalf("len(reviewFlat(idx)) = %d, want 5", len(flat))
	}
	// The last entry must be the rework bean -- every to-review bean precedes
	// every rework bean (types.go's reviewCursor doc-stamp: cursor index space
	// is "to-review flat, THEN rework flat").
	if flat[len(flat)-1].ID != "tk-rw" {
		t.Fatalf("last flat entry = %s, want tk-rw (rework must trail every to-review bean)", flat[len(flat)-1].ID)
	}
	for _, b := range flat[:len(flat)-1] {
		found := false
		for _, t2 := range b.Tags {
			if t2 == "to-review" {
				found = true
			}
		}
		if !found {
			t.Fatalf("flat entry %s before the rework tail is not tagged to-review: %v", b.ID, b.Tags)
		}
	}
}

func TestReviewSummaryLinePositionInToReview(t *testing.T) {
	idx := data.NewIndex(reviewFixtureBeans())
	// 4 to-review beans total (tk-a2, tk-a1, tk-b1, tk-none in flat order),
	// cursor on index 2 (0-based) -> "3 of 4".
	got := reviewSummaryLine(idx, 2)
	if got != "3 of 4" {
		t.Errorf("reviewSummaryLine(idx, 2) = %q, want %q", got, "3 of 4")
	}
}

func TestReviewSummaryLineReworkContext(t *testing.T) {
	idx := data.NewIndex(reviewFixtureBeans())
	flat := reviewFlat(idx)
	reworkIdx := len(flat) - 1 // the single rework bean, always last
	got := reviewSummaryLine(idx, reworkIdx)
	if !strings.Contains(got, "Rework:") {
		t.Errorf("reviewSummaryLine(idx, rework-cursor) = %q, want it to contain %q", got, "Rework:")
	}
	if strings.Contains(got, " of ") {
		t.Errorf("reviewSummaryLine(idx, rework-cursor) = %q, must not contain the to-review %q wording", got, " of ")
	}
}

func TestOpenReviewCockpitResetsCursorAndAccOpen(t *testing.T) {
	m := fixtureModel(t, reviewFixtureBeans())
	m.reviewCursor = 3
	m.reviewAccOpen = 2
	nm, _ := m.openReviewCockpit()
	mm := nm.(model)
	if mm.view != viewReviewCockpit {
		t.Fatalf("view = %v, want viewReviewCockpit", mm.view)
	}
	if mm.reviewCursor != 0 || mm.reviewAccOpen != 0 {
		t.Errorf("reviewCursor/reviewAccOpen = %d/%d, want 0/0", mm.reviewCursor, mm.reviewAccOpen)
	}
}

func TestKeyTreeReviewsOpensCockpit(t *testing.T) {
	m := fixtureModel(t, reviewFixtureBeans())
	m = step(t, m, runeMsg('R'))
	if m.view != viewReviewCockpit {
		t.Fatalf("view after R from Tree = %v, want viewReviewCockpit", m.view)
	}
}

func TestKeyBacklogReviewsOpensCockpit(t *testing.T) {
	m := fixtureModel(t, reviewFixtureBeans())
	m.view = viewBacklog
	m = step(t, m, runeMsg('R'))
	if m.view != viewReviewCockpit {
		t.Fatalf("view after R from Backlog = %v, want viewReviewCockpit", m.view)
	}
}

func TestKeyReviewCockpitBackReturnsToBrowse(t *testing.T) {
	m := fixtureModel(t, reviewFixtureBeans())
	nm, _ := m.openReviewCockpit()
	m = nm.(model)
	m = step(t, m, keyMsg(tea.KeyEsc))
	if m.view != viewBrowseRepo {
		t.Fatalf("view after esc in Cockpit = %v, want viewBrowseRepo", m.view)
	}
}

func TestKeyReviewCockpitQReturnsToBrowse(t *testing.T) {
	m := fixtureModel(t, reviewFixtureBeans())
	nm, _ := m.openReviewCockpit()
	m = nm.(model)
	m = step(t, m, runeMsg('q'))
	if m.view != viewBrowseRepo {
		t.Fatalf("view after q in Cockpit = %v, want viewBrowseRepo", m.view)
	}
}

func TestKeyReviewCockpitNavigationMovesCursor(t *testing.T) {
	m := fixtureModel(t, reviewFixtureBeans())
	nm, _ := m.openReviewCockpit()
	m = nm.(model)

	m = step(t, m, keyMsg(tea.KeyDown))
	if m.reviewCursor != 1 {
		t.Fatalf("reviewCursor after down = %d, want 1", m.reviewCursor)
	}
	m = step(t, m, runeMsg('n')) // explicit next alias
	if m.reviewCursor != 2 {
		t.Fatalf("reviewCursor after n = %d, want 2", m.reviewCursor)
	}
	m = step(t, m, runeMsg('p')) // explicit prev alias
	if m.reviewCursor != 1 {
		t.Fatalf("reviewCursor after p = %d, want 1", m.reviewCursor)
	}
	m = step(t, m, keyMsg(tea.KeyUp))
	if m.reviewCursor != 0 {
		t.Fatalf("reviewCursor after up = %d, want 0", m.reviewCursor)
	}
	// Clamped at 0, no negative.
	m = step(t, m, keyMsg(tea.KeyUp))
	if m.reviewCursor != 0 {
		t.Fatalf("reviewCursor after up-at-top = %d, want clamped 0", m.reviewCursor)
	}
}

func TestKeyReviewCockpitNavigationClampsAtBottom(t *testing.T) {
	m := fixtureModel(t, reviewFixtureBeans())
	nm, _ := m.openReviewCockpit()
	m = nm.(model)
	flat := reviewFlat(m.idx)
	for i := 0; i < len(flat)+3; i++ { // overshoot on purpose
		m = step(t, m, keyMsg(tea.KeyDown))
	}
	if m.reviewCursor != len(flat)-1 {
		t.Fatalf("reviewCursor after overshoot down = %d, want clamped %d", m.reviewCursor, len(flat)-1)
	}
}

func TestKeyReviewCockpitDigitTogglesAccOpen(t *testing.T) {
	m := fixtureModel(t, reviewFixtureBeans())
	nm, _ := m.openReviewCockpit()
	m = nm.(model)

	m = step(t, m, runeMsg('2'))
	if m.reviewAccOpen != 2 {
		t.Fatalf("reviewAccOpen after '2' = %d, want 2", m.reviewAccOpen)
	}
	m = step(t, m, runeMsg('2')) // pressing the SAME digit again closes it
	if m.reviewAccOpen != 0 {
		t.Fatalf("reviewAccOpen after second '2' = %d, want 0 (toggle-closed)", m.reviewAccOpen)
	}
}

func TestHandleKeyCtrlKOpensPaletteFromReviewCockpit(t *testing.T) {
	// Capture-Order-Beleg (design decision h): ctrl+k must reach openPalette
	// from WITHIN the Cockpit too -- keys.Palette is checked ABOVE the
	// viewReviewCockpit capture block in handleKey.
	m := fixtureModel(t, reviewFixtureBeans())
	nm, _ := m.openReviewCockpit()
	m = nm.(model)
	m = step(t, m, tea.KeyMsg{Type: tea.KeyCtrlK})
	if !m.paletteOpen {
		t.Fatal("ctrl+k from inside the Review-Cockpit did not open the palette")
	}
}

func TestKeyReviewCockpitAssignDoesNotOpenParentPicker(t *testing.T) {
	// Capture-Order-Beleg (design decision h): "a" must NOT fall through to
	// keyNodeAction's Assign case while inside the Cockpit -- Task 4 wires "a"
	// as Pass, but even before that lands, "a" must be swallowed here rather
	// than leak to the parent-picker.
	m := fixtureModel(t, reviewFixtureBeans())
	nm, _ := m.openReviewCockpit()
	m = nm.(model)
	m = step(t, m, runeMsg('a'))
	if m.overlay != overlayNone {
		t.Fatalf("overlay = %v after 'a' in Review-Cockpit, want overlayNone (a must not leak to keyNodeAction's Assign case)", m.overlay)
	}
	if m.view != viewReviewCockpit {
		t.Fatalf("view = %v after 'a' in Review-Cockpit, want unchanged viewReviewCockpit", m.view)
	}
}

// goldenReviewCockpitModel builds the deterministic fixture for
// TestReviewCockpitGolden: 2 Epic groups + 1 (kein Epic) bean + 1 Rework
// bean -- every row kind the Cockpit renders is on screen at once.
func goldenReviewCockpitModel(t *testing.T) model {
	t.Helper()
	m := newModel(nil, "/tmp/bt-golden-repo")
	m = step(t, m, beansLoadedMsg{beans: reviewFixtureBeans()})
	m = step(t, m, tea.WindowSizeMsg{Width: 100, Height: 30})
	nm, _ := m.openReviewCockpit()
	m = nm.(model)
	m.reviewCursor = 1 // parked on a real bean row, not row 0, for render coverage
	return m
}

// TestReviewCockpitGolden renders a full 100x30 frame of the Review-Cockpit
// and compares it against testdata/review_cockpit.golden.
// Regenerate with: go test ./internal/tui/ -run TestReviewCockpitGolden -update
func TestReviewCockpitGolden(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(termenv.Ascii)

	m := goldenReviewCockpitModel(t)
	out := m.View()

	if h := lipgloss.Height(out); h != 30 {
		t.Errorf("review cockpit view height=%d, want 30 (full terminal height)", h)
	}
	for i, ln := range strings.Split(out, "\n") {
		if w := lipgloss.Width(ln); w > 100 {
			t.Errorf("line %d overflows width (%d > 100): %q", i, w, ln)
		}
	}

	path := filepath.Join("testdata", "review_cockpit.golden")
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
		t.Errorf("review cockpit view output differs from golden %q (frame/width/truncation?).\n--- got ---\n%s\n--- want ---\n%s", path, out, string(want))
	}
}

// TestReviewCockpitGoldenDeterministic guards that View() is a pure function
// of the model for the Cockpit too (mirrors tree_golden_test.go's own
// TestTreeGoldenDeterministic).
func TestReviewCockpitGoldenDeterministic(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(termenv.Ascii)

	m := goldenReviewCockpitModel(t)
	a := m.View()
	b := m.View()
	if a != b {
		t.Error("Review-Cockpit View() is not deterministic across repeated calls with identical model state")
	}
}

// TestReviewCockpitEmptyPlaceholder guards the "(keine offenen Reviews)"
// placeholder (Port devd view_navigate_reviews.go:20 Textmuster) when no bean
// carries to-review/rework at all.
func TestReviewCockpitEmptyPlaceholder(t *testing.T) {
	beans := []data.Bean{
		{ID: "ms-1", Title: "Milestone", Type: "milestone"},
	}
	m := fixtureModel(t, beans)
	m = step(t, m, tea.WindowSizeMsg{Width: 100, Height: 30})
	nm, _ := m.openReviewCockpit()
	m = nm.(model)
	out := m.View()
	if !strings.Contains(out, "keine offenen Reviews") {
		t.Errorf("empty Review-Cockpit render does not contain the placeholder text, got:\n%s", out)
	}
}
