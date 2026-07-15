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

// --- Verdikt-Aktionen a/x/o (E4 Task 4, bean bt-yy6w) ---

// TestKeyReviewCockpitPassFiresPassReview guards "a" on a to-review item
// (reviewCursor == 0 by construction, openReviewCockpit's own reset): fires
// data.Client.PassReview against the CURSORED bean's fresh etag. Drives
// keyReviewCockpit directly (not step/Update) so the returned Cmd can be
// inspected BEFORE it fires -- mirrors form_edit_title_test.go's
// TestEditTitleSubmitFiresSetTitleDirectlyNoConfirm dispatch-proof pattern:
// fixtureModel's client points at a non-repo dir, so cmd() yields a REAL
// "beans update" CLI error, proving PassReview actually dispatched rather
// than a mocked stand-in.
func TestKeyReviewCockpitPassFiresPassReview(t *testing.T) {
	m := fixtureModel(t, reviewFixtureBeans())
	m.client = &data.Client{RepoDir: "/nonexistent-bt-e4-t4-scratch-dir"}
	nm, _ := m.openReviewCockpit()
	m = nm.(model)

	_, cmd := m.keyReviewCockpit(runeMsg('a'))
	if cmd == nil {
		t.Fatal("'a' on a to-review bean must fire a Cmd (PassReview)")
	}
	msg := cmd()
	mdm, ok := msg.(mutationDoneMsg)
	if !ok {
		t.Fatalf("cmd() = %T, want mutationDoneMsg", msg)
	}
	if mdm.err == nil || !strings.Contains(mdm.err.Error(), "beans update") {
		t.Fatalf("mutationDoneMsg.err = %v, want an error containing %q (proves PassReview dispatched)", mdm.err, "beans update")
	}
}

// TestKeyReviewCockpitPassOnReworkIsNoop guards the reviewIsRework guard on
// "a": a Rework item has already been verdicted once (rejected) -- Pass must
// not fire against it (design decision f's own "already verdicted" logic,
// mirrored onto Pass/Reject, not just Reopen).
func TestKeyReviewCockpitPassOnReworkIsNoop(t *testing.T) {
	m := fixtureModel(t, reviewFixtureBeans())
	nm, _ := m.openReviewCockpit()
	m = nm.(model)
	flat := reviewFlat(m.idx)
	m.reviewCursor = len(flat) - 1 // the single rework bean (tk-rw), always last

	nm2, cmd := m.keyReviewCockpit(runeMsg('a'))
	if cmd != nil {
		t.Fatal("'a' on a Rework item must be a no-op -- must not fire PassReview")
	}
	mm, ok := nm2.(model)
	if !ok {
		t.Fatalf("keyReviewCockpit did not return a model, got %T", nm2)
	}
	if mm.err != "" {
		t.Fatalf("err = %q, want empty (no-op, not a surfaced error)", mm.err)
	}
}

// TestKeyReviewCockpitRejectOpensCommentForm guards "x" on a to-review item:
// opens the Reject-Kommentar-Form (m.formKind == "reject"), captures the
// CURSORED bean's ID on m.mutTarget (same convention as openEditTitleForm).
func TestKeyReviewCockpitRejectOpensCommentForm(t *testing.T) {
	m := fixtureModel(t, reviewFixtureBeans())
	nm, _ := m.openReviewCockpit()
	m = nm.(model)
	want := reviewFlat(m.idx)[0].ID

	nm2, _ := m.keyReviewCockpit(runeMsg('x'))
	mm, ok := nm2.(model)
	if !ok {
		t.Fatalf("keyReviewCockpit did not return a model, got %T", nm2)
	}
	if mm.form == nil || mm.formKind != "reject" {
		t.Fatalf("'x' did not open the reject form (form=%v formKind=%q)", mm.form, mm.formKind)
	}
	if mm.mutTarget != want {
		t.Fatalf("mutTarget = %q, want %q", mm.mutTarget, want)
	}
}

// TestKeyReviewCockpitRejectOnReworkIsNoop mirrors
// TestKeyReviewCockpitPassOnReworkIsNoop for "x".
func TestKeyReviewCockpitRejectOnReworkIsNoop(t *testing.T) {
	m := fixtureModel(t, reviewFixtureBeans())
	nm, _ := m.openReviewCockpit()
	m = nm.(model)
	flat := reviewFlat(m.idx)
	m.reviewCursor = len(flat) - 1

	nm2, _ := m.keyReviewCockpit(runeMsg('x'))
	mm, ok := nm2.(model)
	if !ok {
		t.Fatalf("keyReviewCockpit did not return a model, got %T", nm2)
	}
	if mm.form != nil {
		t.Fatal("'x' on a Rework item must not open the Reject-Form")
	}
}

// TestKeyReviewCockpitReopenOnReworkFiresSetTags guards "o" on a Rework item
// (design decision f): fires data.Client.SetTags(id, [to-review], [rework],
// etag) -- the exact E3-Task-2 combined-diff wrapper, reused unchanged.
func TestKeyReviewCockpitReopenOnReworkFiresSetTags(t *testing.T) {
	m := fixtureModel(t, reviewFixtureBeans())
	m.client = &data.Client{RepoDir: "/nonexistent-bt-e4-t4-scratch-dir"}
	nm, _ := m.openReviewCockpit()
	m = nm.(model)
	flat := reviewFlat(m.idx)
	m.reviewCursor = len(flat) - 1 // the single rework bean (tk-rw), always last

	_, cmd := m.keyReviewCockpit(runeMsg('o'))
	if cmd == nil {
		t.Fatal("'o' on a Rework item must fire a Cmd (SetTags reopen)")
	}
	msg := cmd()
	mdm, ok := msg.(mutationDoneMsg)
	if !ok {
		t.Fatalf("cmd() = %T, want mutationDoneMsg", msg)
	}
	if mdm.err == nil || !strings.Contains(mdm.err.Error(), "beans update") {
		t.Fatalf("mutationDoneMsg.err = %v, want an error containing %q (proves SetTags dispatched)", mdm.err, "beans update")
	}
}

// TestKeyReviewCockpitReopenOnToReviewIsNoop guards "o" on a to-review item
// (design decision f: "bereits im Zielzustand" -- already unverdicted, there
// is nothing to reopen).
func TestKeyReviewCockpitReopenOnToReviewIsNoop(t *testing.T) {
	m := fixtureModel(t, reviewFixtureBeans())
	nm, _ := m.openReviewCockpit()
	m = nm.(model) // reviewCursor == 0 (openReviewCockpit's own reset) -> a to-review item

	nm2, cmd := m.keyReviewCockpit(runeMsg('o'))
	if cmd != nil {
		t.Fatal("'o' on a to-review item must be a no-op (design decision f: already in target state)")
	}
	mm, ok := nm2.(model)
	if !ok {
		t.Fatalf("keyReviewCockpit did not return a model, got %T", nm2)
	}
	if mm.reviewCursor != 0 {
		t.Fatalf("reviewCursor changed = %d, want unchanged 0", mm.reviewCursor)
	}
}

// --- focusedBean() Cockpit case (E4 Task 4, T3-Review "I (kein Bug)" note) ---

// TestFocusedBeanReviewCockpitCase guards that focusedBean() resolves the
// Cockpit's OWN reviewFlat/reviewCursor while m.view == viewReviewCockpit,
// not the (possibly stale/irrelevant) Tree cursor -- otherwise ctrl+k's
// Palette-Node-Actions inside the Cockpit would act on the wrong bean
// (bt-hxyo's own carried-forward observation).
func TestFocusedBeanReviewCockpitCase(t *testing.T) {
	m := fixtureModel(t, reviewFixtureBeans())
	nm, _ := m.openReviewCockpit()
	m = nm.(model)
	flat := reviewFlat(m.idx)
	m.reviewCursor = 1

	got := m.focusedBean()
	if got == nil || got.ID != flat[1].ID {
		t.Fatalf("focusedBean() in Review-Cockpit = %v, want %s (reviewFlat[reviewCursor])", got, flat[1].ID)
	}
}

// TestFocusedBeanReviewCockpitOutOfRangeReturnsNil is the defensive
// counterpart: an out-of-range reviewCursor (e.g. a render landing between
// two keystrokes against a since-shrunk queue) must not panic or return a
// stale bean.
func TestFocusedBeanReviewCockpitOutOfRangeReturnsNil(t *testing.T) {
	m := fixtureModel(t, reviewFixtureBeans())
	nm, _ := m.openReviewCockpit()
	m = nm.(model)
	m.reviewCursor = 999

	if got := m.focusedBean(); got != nil {
		t.Fatalf("focusedBean() with out-of-range reviewCursor = %v, want nil", got)
	}
}

// --- Inherited T3-Review PFLICHT test gaps (bt-hxyo's own Akzeptanz list, carried into this task) ---

// TestReviewQueueEpicItselfTaggedAppearsInKeinEpicBucket closes gap (a): a
// to-review-tagged EPIC bean is its OWN group's header (relationRow(g.epic),
// reviewQueueRows) for its to-review CHILDREN, but EpicAncestor(idx, b)
// looks at b's OWN Parent chain, never at b itself -- so the epic's own
// to-review tag places IT in the "(kein Epic)" bucket as an ordinary row
// entry (its parent is the milestone, not another epic), separate from its
// role as a group header. Guards that reviewQueue handles a bean being both
// "a group's header" and "a member of a DIFFERENT group" without crashing or
// misplacing it.
func TestReviewQueueEpicItselfTaggedAppearsInKeinEpicBucket(t *testing.T) {
	beans := []data.Bean{
		{ID: "ms-1", Title: "Milestone", Status: "todo", Type: "milestone", Priority: "normal"},
		{ID: "ep-a", Title: "Epic Alpha", Status: "in-progress", Type: "epic", Priority: "normal", Parent: "ms-1", Tags: []string{"to-review"}},
		{ID: "tk-a1", Title: "Alpha Task One", Status: "in-progress", Type: "task", Priority: "normal", Parent: "ep-a", Tags: []string{"to-review"}},
	}
	idx := data.NewIndex(beans)
	groups := reviewQueue(idx)

	if len(groups) != 2 {
		t.Fatalf("len(reviewQueue) = %d, want 2 (ep-a's own child-group + (kein Epic) containing ep-a itself)", len(groups))
	}
	var epicGroup, noEpicGroup *reviewGroup
	for i := range groups {
		if groups[i].epic != nil {
			epicGroup = &groups[i]
		} else {
			noEpicGroup = &groups[i]
		}
	}
	if epicGroup == nil || epicGroup.epic.ID != "ep-a" {
		t.Fatalf("expected an ep-a-headed group, got %+v", groups)
	}
	if len(epicGroup.beans) != 1 || epicGroup.beans[0].ID != "tk-a1" {
		t.Fatalf("ep-a group beans = %v, want [tk-a1]", epicGroup.beans)
	}
	if noEpicGroup == nil || len(noEpicGroup.beans) != 1 || noEpicGroup.beans[0].ID != "ep-a" {
		t.Fatalf("(kein Epic) bucket = %v, want [ep-a] (the epic itself, tagged to-review, has no epic ANCESTOR of its own)", noEpicGroup)
	}
}

// TestReviewCockpitViewClampsCursorWhenQueueShrinksExternally closes gap
// (b): m.reviewCursor set from a PREVIOUS render against a longer flat, then
// the live idx shrinks (simulating a watch-reload landing between
// keystrokes) WITHOUT reviewCursor being resynced -- View() must clamp
// rather than index out of range (viewReviewCockpit's own documented
// defensive-clamp comment, previously never exercised by a test).
func TestReviewCockpitViewClampsCursorWhenQueueShrinksExternally(t *testing.T) {
	m := fixtureModel(t, reviewFixtureBeans())
	m = step(t, m, tea.WindowSizeMsg{Width: 100, Height: 30})
	nm, _ := m.openReviewCockpit()
	m = nm.(model)
	m.reviewCursor = 4 // valid against reviewFixtureBeans' 5-bean flat

	shrunk := []data.Bean{
		{ID: "ms-1", Title: "Milestone", Status: "todo", Type: "milestone", Priority: "normal"},
		{ID: "tk-solo", Title: "Solo Review Task", Status: "in-progress", Type: "task", Priority: "normal", Tags: []string{"to-review"}},
	}
	m.idx = data.NewIndex(shrunk)

	out := m.View() // must not panic (index out of range)
	// "tk-solo" (the ID), not the full title -- the narrow list pane
	// truncates the title row ("Solo Review Ta…"), same truncate() budget
	// every other row in this pane already has.
	if !strings.Contains(out, "tk-solo") || !strings.Contains(out, "1 of 1") {
		t.Errorf("clamped render does not show the surviving bean at a clamped \"1 of 1\" summary, got:\n%s", out)
	}
}

// TestReviewCursorClampedThroughRealMutationReloadPipeline (B01, E4-T4-Review
// PFLICHT, bean bt-v7ti) is TestReviewCockpitViewClampsCursorWhenQueueShrinksExternally's
// pipeline counterpart: that test only proves View()'s own RENDER-LOCAL
// defensive clamp (a local `cursor` variable, viewReviewCockpit's own doc
// comment) never panics -- it never touches m.reviewCursor itself, so the
// model FIELD stayed stale/out-of-range across a real Pass. This test drives
// the actual pipeline a PO triggers by pressing "a" on the LAST queue item:
// keyReviewCockpit's own mutateCmd -> the resolved mutationDoneMsg run
// through m.Update (applyMutationResult's unconditional reload) -> the
// reload's own beansLoadedMsg (simulating the server-side Pass having
// removed that bean) run through m.Update (applyLoaded). Before the B01 fix,
// m.reviewCursor came out the other end still pointing at the pre-shrink
// last index -- reviewFocused(flat, m.reviewCursor) resolved nil despite a
// visually "valid-looking" render, and "down"/"n" (guarded by `cursor <
// len(flat)-1`, permanently false against a too-large stale cursor) never
// self-healed, only "up"/"p" would (by decrementing regardless of validity)
// -- exactly the asymmetric freeze T4-Review's B01 finding describes.
func TestReviewCursorClampedThroughRealMutationReloadPipeline(t *testing.T) {
	beans := []data.Bean{
		{ID: "ms-1", Title: "Milestone", Status: "todo", Type: "milestone", Priority: "normal"},
		{ID: "tk-1", Title: "Queue Item One", Status: "in-progress", Type: "task", Priority: "normal", Tags: []string{"to-review"}},
		{ID: "tk-2", Title: "Queue Item Two", Status: "in-progress", Type: "task", Priority: "normal", Tags: []string{"to-review"}},
		{ID: "tk-3", Title: "Queue Item Three (last)", Status: "in-progress", Type: "task", Priority: "normal", Tags: []string{"to-review"}},
	}
	m := fixtureModel(t, beans)
	m.client = &data.Client{RepoDir: "/nonexistent-bt-v7ti-t5-scratch-dir"} // dispatch-proof pattern (T4's own test convention)
	nm, _ := m.openReviewCockpit()
	m = nm.(model)

	flat := reviewFlat(m.idx)
	if len(flat) != 3 {
		t.Fatalf("setup: len(reviewFlat) = %d, want 3", len(flat))
	}
	lastID := flat[len(flat)-1].ID
	m.reviewCursor = len(flat) - 1 // parked on the LAST queue item (B01's exact trigger)

	// Step 1: "a" (Pass) on the last item -- mirrors
	// TestKeyReviewCockpitPassFiresPassReview's own dispatch-proof: fires
	// mutateCmd(PassReview); the broken client makes it fail, but
	// applyMutationResult's tail (design decision d) reloads
	// UNCONDITIONALLY regardless of success/failure -- irrelevant to this
	// test, which only cares about the cursor-clamp side effect once a
	// reload's beansLoadedMsg lands.
	nm2, cmd := m.keyReviewCockpit(runeMsg('a'))
	m = nm2.(model)
	if cmd == nil {
		t.Fatal("'a' on the last to-review item must fire a Cmd (PassReview)")
	}
	msg := cmd()
	if _, ok := msg.(mutationDoneMsg); !ok {
		t.Fatalf("cmd() = %T, want mutationDoneMsg", msg)
	}

	// Step 2: run the mutationDoneMsg through the REAL Update dispatcher --
	// applyMutationResult's own unconditional-reload tail.
	tm, reloadCmd := m.Update(msg)
	m = tm.(model)
	if reloadCmd == nil {
		t.Fatal("mutationDoneMsg must always trigger a reload Cmd (design decision d)")
	}

	// Step 3: simulate the reload's own result -- server-side the Pass
	// succeeded, tk-3 (the bean the stale cursor still points at) is GONE,
	// the boundary-shrinking case B01 names explicitly. Driven through the
	// real Update dispatcher (applyLoaded), not by hand-poking m.idx (that's
	// what the render-only sibling test above already covers).
	var shrunk []data.Bean
	for _, b := range beans {
		if b.ID == lastID {
			continue // simulates the server-side Pass having removed exactly this bean
		}
		shrunk = append(shrunk, b)
	}
	m = step(t, m, beansLoadedMsg{beans: shrunk})

	newFlat := reviewFlat(m.idx)
	if len(newFlat) != 2 {
		t.Fatalf("setup: len(reviewFlat) after shrink = %d, want 2", len(newFlat))
	}
	for _, b := range newFlat {
		if b.ID == lastID {
			t.Fatalf("newFlat still contains %s, want it removed (simulates the Pass having succeeded server-side)", lastID)
		}
	}
	if m.reviewCursor >= len(newFlat) || m.reviewCursor < 0 {
		t.Fatalf("reviewCursor = %d after boundary-shrinking Pass, want clamped into [0, %d)", m.reviewCursor, len(newFlat))
	}

	// The core B01 regression: reviewFocused (== focusedBean's Cockpit case)
	// must resolve a REAL bean, not nil, immediately after the reload -- no
	// keypress required to "walk" the cursor back into range.
	if got := reviewFocused(newFlat, m.reviewCursor); got == nil {
		t.Fatal("reviewFocused(flat, reviewCursor) = nil immediately after the reload, want the clamped bean (B01)")
	}
	if got := m.focusedBean(); got == nil {
		t.Fatal("focusedBean() = nil immediately after the reload, want the clamped bean (B01)")
	}

	// Step 4: "next keypress works" (the task spec's own regression wording)
	// -- "down"/"n" must no longer be frozen. Before the fix, down's guard
	// (`m.reviewCursor < len(flat)-1`) stayed permanently false against a
	// stale too-large cursor; post-clamp it now sits at the valid last
	// index, so "down"/"n" is a correct, harmless no-op (already at the
	// end), and "up"/"p" actually moves -- both checked, so a fix that only
	// clamps without restoring real up/down symmetry would still fail here.
	before := m.reviewCursor
	nm3, _ := m.keyReviewCockpit(runeMsg('n'))
	m = nm3.(model)
	if m.reviewCursor != before {
		t.Fatalf("'n' at the (now-clamped) last index moved reviewCursor from %d to %d, want unchanged (already at the end)", before, m.reviewCursor)
	}
	nm4, _ := m.keyReviewCockpit(runeMsg('p'))
	m = nm4.(model)
	if m.reviewCursor != before-1 {
		t.Fatalf("'p' after the clamp = %d, want %d (must actually move, not stay frozen)", m.reviewCursor, before-1)
	}
}

// TestReviewCockpitReworkOnlyQueueNoToReviewGroups closes gap (c): zero
// to-review beans, one Rework bean -- reviewQueue must be empty (no epic
// groups at all), reviewFlat must be rework-only, the summary line must
// read "Rework: n offen" even at cursor 0 (no to-review beans exist to
// count), and the render must show the "── Rework ──" section with no epic
// group headers above it.
func TestReviewCockpitReworkOnlyQueueNoToReviewGroups(t *testing.T) {
	beans := []data.Bean{
		{ID: "ms-1", Title: "Milestone", Status: "todo", Type: "milestone", Priority: "normal"},
		{ID: "ep-a", Title: "Epic Alpha", Status: "todo", Type: "epic", Priority: "normal", Parent: "ms-1"},
		{ID: "tk-rw1", Title: "Rework Only Task", Status: "in-progress", Type: "task", Priority: "normal", Parent: "ep-a", Tags: []string{"rework"}},
	}
	idx := data.NewIndex(beans)

	if groups := reviewQueue(idx); len(groups) != 0 {
		t.Fatalf("reviewQueue(idx) with no to-review beans = %v, want empty", groups)
	}
	flat := reviewFlat(idx)
	if len(flat) != 1 || flat[0].ID != "tk-rw1" {
		t.Fatalf("reviewFlat(idx) rework-only = %v, want [tk-rw1]", flat)
	}
	got := reviewSummaryLine(idx, 0)
	if !strings.Contains(got, "Rework:") || strings.Contains(got, " of ") {
		t.Fatalf("reviewSummaryLine(idx, 0) rework-only = %q, want Rework-context wording (no to-review beans exist at all)", got)
	}

	m := fixtureModel(t, beans)
	m = step(t, m, tea.WindowSizeMsg{Width: 100, Height: 30})
	nm, _ := m.openReviewCockpit()
	m = nm.(model)
	out := m.View()
	if !strings.Contains(out, "── Rework ──") || !strings.Contains(out, "Rework Only Task") {
		t.Errorf("rework-only render missing the Rework section, got:\n%s", out)
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
