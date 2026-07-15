package tui

// update_test.go — TDD coverage for the App-Shell Update dispatcher (T8,
// implementation-plan.md »E1 Task 8«): cursor movement/expand/collapse,
// quit-confirm, reload-keeps-cursor-by-ID, and the MANDATORY orphan rule
// (bean bt-7jr8 / T3-review Q01). Also covers the T8 Opus quality-review
// follow-ups (bean bt-7jr8, Runde 2): tree windowing (B01), cycle-bean
// visibility (B02), and tab focus-swap accenting (I05 #1).

import (
	"fmt"
	"strings"
	"testing"

	"beans-tui/internal/data"
	"beans-tui/internal/theme"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// fixtureBeans is a small milestone -> epic -> 2 tasks hierarchy shared by
// the Update tests below.
func fixtureBeans() []data.Bean {
	return []data.Bean{
		{ID: "ms-1", Title: "Milestone One", Status: "todo", Type: "milestone", Priority: "normal"},
		{ID: "ep-1", Title: "Epic One", Status: "todo", Type: "epic", Priority: "normal", Parent: "ms-1"},
		// tk-1 sorts before tk-2 (Index.sortBeans: Status -> Priority -> Type ->
		// Title; in-progress ranks before todo) -- deliberately picked so
		// "tk-1 first, tk-2 last" holds without fighting the real sort order.
		{ID: "tk-1", Title: "Task One", Status: "in-progress", Type: "task", Priority: "high", Parent: "ep-1"},
		{ID: "tk-2", Title: "Task Two", Status: "todo", Type: "task", Priority: "normal", Parent: "ep-1"},
	}
}

// fixtureModel builds a model already past its initial load (mirrors what
// Init()+the real beansLoadedMsg round-trip produces, without a live client).
func fixtureModel(t *testing.T, beans []data.Bean) model {
	t.Helper()
	return step(t, newModel(nil, "/tmp/bt-fixture-repo"), beansLoadedMsg{beans: beans})
}

// step sends msg through m.Update and type-asserts the resulting tea.Model
// back to model, failing the test if that assertion ever breaks (it never
// should -- Update always returns the same concrete type it received).
func step(t *testing.T, m model, msg tea.Msg) model {
	t.Helper()
	tm, _ := m.Update(msg)
	mm, ok := tm.(model)
	if !ok {
		t.Fatalf("Update(%T) did not return a model, got %T", msg, tm)
	}
	return mm
}

func keyMsg(k tea.KeyType) tea.KeyMsg { return tea.KeyMsg{Type: k} }
func runeMsg(r rune) tea.KeyMsg       { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}} }

func nodeIDs(nodes []treeNode) []string {
	ids := make([]string, len(nodes))
	for i, n := range nodes {
		ids[i] = n.id
	}
	return ids
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestCursorMovesAndExpands drives a KeyMsg sequence (down/expand/collapse)
// against a fixture Index and asserts both the visible node set and the
// cursor position at each step.
func TestCursorMovesAndExpands(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())

	// Initial state: only the milestone is a root (epic/tasks are parented),
	// and the cursor defaults onto it.
	nodes := m.visibleNodes()
	if got := nodeIDs(nodes); !equalStrings(got, []string{"ms-1"}) {
		t.Fatalf("initial roots = %v, want [ms-1]", got)
	}
	if m.cursorID != "ms-1" {
		t.Fatalf("initial cursorID = %q, want ms-1", m.cursorID)
	}

	// down at the bottom of a 1-row list is a no-op (clamped).
	m = step(t, m, keyMsg(tea.KeyDown))
	if m.cursorID != "ms-1" {
		t.Fatalf("cursorID after down (no-op) = %q, want ms-1", m.cursorID)
	}

	// right: expand the milestone -> its epic becomes visible.
	m = step(t, m, keyMsg(tea.KeyRight))
	nodes = m.visibleNodes()
	if got := nodeIDs(nodes); !equalStrings(got, []string{"ms-1", "ep-1"}) {
		t.Fatalf("after expand ms-1: nodes = %v, want [ms-1 ep-1]", got)
	}

	// down: cursor moves onto the epic.
	m = step(t, m, keyMsg(tea.KeyDown))
	if m.cursorID != "ep-1" {
		t.Fatalf("cursorID after down = %q, want ep-1", m.cursorID)
	}

	// right: expand the epic -> both tasks become visible.
	m = step(t, m, keyMsg(tea.KeyRight))
	nodes = m.visibleNodes()
	if got := nodeIDs(nodes); !equalStrings(got, []string{"ms-1", "ep-1", "tk-1", "tk-2"}) {
		t.Fatalf("after expand ep-1: nodes = %v, want [ms-1 ep-1 tk-1 tk-2]", got)
	}

	// down, down: cursor lands on the last task; a further down clamps there.
	m = step(t, m, keyMsg(tea.KeyDown))
	m = step(t, m, keyMsg(tea.KeyDown))
	if m.cursorID != "tk-2" {
		t.Fatalf("cursorID after 2x down = %q, want tk-2", m.cursorID)
	}
	m = step(t, m, keyMsg(tea.KeyDown))
	if m.cursorID != "tk-2" {
		t.Fatalf("cursor moved past the last row: cursorID = %q, want tk-2 (clamped)", m.cursorID)
	}

	// left on a leaf is a no-op (leaves carry no expand state).
	m = step(t, m, keyMsg(tea.KeyLeft))
	if len(m.visibleNodes()) != 4 {
		t.Fatal("left on a leaf changed the visible node count")
	}

	// enter on a leaf is also a no-op (task scope: leaves don't open detail yet).
	m = step(t, m, keyMsg(tea.KeyEnter))
	if len(m.visibleNodes()) != 4 || m.cursorID != "tk-2" {
		t.Fatal("enter on a leaf must be a no-op")
	}

	// up, up: cursor back on the epic; left collapses it -> tasks disappear.
	m = step(t, m, keyMsg(tea.KeyUp))
	m = step(t, m, keyMsg(tea.KeyUp))
	if m.cursorID != "ep-1" {
		t.Fatalf("cursorID after 2x up = %q, want ep-1", m.cursorID)
	}
	m = step(t, m, keyMsg(tea.KeyLeft))
	nodes = m.visibleNodes()
	if got := nodeIDs(nodes); !equalStrings(got, []string{"ms-1", "ep-1"}) {
		t.Fatalf("after collapse ep-1: nodes = %v, want [ms-1 ep-1]", got)
	}
}

// TestQuitConfirm guards the quit-confirm modal: q opens it (visible in
// View()), esc cancels it (gone from View()), enter resolves to tea.Quit.
func TestQuitConfirm(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m.width, m.height = 80, 24

	m = step(t, m, runeMsg('q'))
	if !m.confirmQuit {
		t.Fatal("q did not open the quit-confirm")
	}
	if !strings.Contains(m.View(), "Quit?") {
		t.Fatal("View() does not show the quit-confirm once open")
	}

	m = step(t, m, keyMsg(tea.KeyEsc))
	if m.confirmQuit {
		t.Fatal("esc did not cancel the quit-confirm")
	}
	if strings.Contains(m.View(), "Quit?") {
		t.Fatal("View() still shows the quit-confirm after esc")
	}

	m = step(t, m, runeMsg('q'))
	_, cmd := m.Update(keyMsg(tea.KeyEnter))
	if cmd == nil {
		t.Fatal("enter on the quit-confirm must return a Cmd")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatal("enter's Cmd did not resolve to tea.QuitMsg")
	}
}

// TestCtrlCQuitsImmediately guards the task-scoped distinction from `q`:
// ctrl+c bypasses the confirm entirely.
func TestCtrlCQuitsImmediately(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	_, cmd := m.Update(keyMsg(tea.KeyCtrlC))
	if cmd == nil {
		t.Fatal("ctrl+c must return a Cmd")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatal("ctrl+c's Cmd did not resolve to tea.QuitMsg")
	}
}

// TestReloadKeepsCursorOnID guards US-10: reloading with the same beans keeps
// the cursor on the same bean ID; reloading with the cursored bean removed
// clamps to the nearest surviving row instead of resetting to the top.
func TestReloadKeepsCursorOnID(t *testing.T) {
	beans := fixtureBeans()
	m := fixtureModel(t, beans)
	m = step(t, m, keyMsg(tea.KeyRight)) // expand ms-1
	m = step(t, m, keyMsg(tea.KeyDown))  // -> ep-1
	m = step(t, m, keyMsg(tea.KeyRight)) // expand ep-1
	m = step(t, m, keyMsg(tea.KeyDown))  // -> tk-1
	m = step(t, m, keyMsg(tea.KeyDown))  // -> tk-2
	if m.cursorID != "tk-2" {
		t.Fatalf("setup: cursorID = %q, want tk-2", m.cursorID)
	}

	// Reload with the same beans (unchanged): cursor stays on tk-2.
	m = step(t, m, beansLoadedMsg{beans: beans})
	if m.cursorID != "tk-2" {
		t.Fatalf("cursor after no-op reload = %q, want tk-2 (unchanged)", m.cursorID)
	}

	// Reload with tk-2 removed: cursor must not still point at it, and must
	// clamp to a bean that is actually still visible in the new tree.
	without := []data.Bean{beans[0], beans[1], beans[2]} // ms-1, ep-1, tk-1
	m = step(t, m, beansLoadedMsg{beans: without})
	if m.cursorID == "tk-2" {
		t.Fatal("cursor still points at the removed bean")
	}
	nodes := m.visibleNodes()
	found := false
	for _, n := range nodes {
		if n.id == m.cursorID {
			found = true
		}
	}
	if !found {
		t.Fatalf("cursorID %q does not match any visible node after reload", m.cursorID)
	}
	// tk-2 sat at the last position (index 3 of 4); with only 3 nodes left
	// (ms-1, ep-1, tk-1) the clamp lands on the new last row, tk-1.
	if m.cursorID != "tk-1" {
		t.Fatalf("cursorID after removal = %q, want tk-1 (clamped to nearest surviving row)", m.cursorID)
	}
}

// errBoom is a minimal error fixture for TestBeansLoadedErrorSurfacesInStatusLine.
type errBoom struct{}

func (errBoom) Error() string { return "boom" }

// TestBeansLoadedErrorSurfacesInStatusLine guards that a failed (re)load
// renders into the status line instead of crashing/blanking the tree.
func TestBeansLoadedErrorSurfacesInStatusLine(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m.width, m.height = 80, 24

	m = step(t, m, beansLoadedMsg{err: errBoom{}})
	if m.err == "" {
		t.Fatal("m.err empty after a failed reload")
	}
	if !strings.Contains(m.View(), "boom") {
		t.Errorf("View() does not surface the load error %q", m.err)
	}
	// The tree itself must survive untouched (idx not clobbered by a failed reload).
	if len(m.visibleNodes()) == 0 {
		t.Error("tree emptied out after a failed reload -- previous Index must be kept")
	}
}

// TestOrphanShownUnderSyntheticRoot guards the MANDATORY orphan rule (bean
// bt-7jr8 / T3-review Q01): a bean with an unresolvable parent must appear
// under the synthetic "(verwaist)" root, never silently dropped, and never
// leak into the real Roots() top level (its Parent is non-empty).
func TestOrphanShownUnderSyntheticRoot(t *testing.T) {
	beans := append(fixtureBeans(), data.Bean{
		ID: "orph-1", Title: "Orphaned Task", Status: "todo", Type: "task",
		Priority: "normal", Parent: "does-not-exist",
	})
	m := fixtureModel(t, beans)

	nodes := m.visibleNodes()
	var orphanRoot *treeNode
	for i := range nodes {
		if nodes[i].orphan {
			orphanRoot = &nodes[i]
		}
	}
	if orphanRoot == nil {
		t.Fatal("orphan root node ('(verwaist)') missing from the tree -- orphan silently dropped")
	}
	if orphanRoot.id != orphanRootID {
		t.Errorf("orphan root id = %q, want the orphanRootID sentinel", orphanRoot.id)
	}

	for _, b := range m.idx.Roots() {
		if b.ID == "orph-1" {
			t.Fatal("orphaned bean leaked into the real Roots() top level")
		}
	}

	// Expand the orphan root -> the orphaned bean appears nested under it.
	m.expanded[orphanRootID] = true
	nodes = m.visibleNodes()
	found := false
	for _, n := range nodes {
		if n.id == "orph-1" {
			found = true
			if n.depth == 0 {
				t.Error("orphaned bean rendered at depth 0 -- must be nested under the orphan root")
			}
		}
	}
	if !found {
		t.Fatal("orphaned bean not visible even with the orphan root expanded")
	}
}

// TestOrphanBucketUsesCanonicalStatusPriorityTypeTitleOrder guards I03
// closure (bean bt-9ldr, E2 Task 4): the orphan bucket must sort via
// data.SortBeans (Status -> Priority -> Type -> Title, the single-source
// order Index.Children/Roots() already use), NOT the tree package's own
// title-only sortByTitleThenID -- the fixture titles are picked so
// alphabetical-title order and canonical-status order DISAGREE, making the
// assertion a real discriminator.
func TestOrphanBucketUsesCanonicalStatusPriorityTypeTitleOrder(t *testing.T) {
	beans := []data.Bean{
		{ID: "root", Title: "Root", Type: "milestone", Status: "todo"},
		{ID: "orph-z", Title: "Z orphan", Status: "todo", Type: "task", Priority: "normal", Parent: "missing"},
		{ID: "orph-a", Title: "A orphan", Status: "in-progress", Type: "task", Priority: "critical", Parent: "missing"},
	}
	m := fixtureModel(t, beans)
	m.expanded[orphanRootID] = true
	nodes := m.visibleNodes()
	// canonical order ranks in-progress before todo -> orph-a (in-progress)
	// must render before orph-z (todo); orph-a also happens to win
	// alphabetically here, but TestOrphanBucketUsesCanonicalOrderOverTitle
	// below picks titles where the two orders disagree, so this pair alone
	// is a sanity check, not the sole discriminator.
	var order []string
	for _, n := range nodes {
		if n.bean != nil && n.bean.Parent == "missing" {
			order = append(order, n.id)
		}
	}
	if len(order) != 2 || order[0] != "orph-a" {
		t.Fatalf("orphan order = %v, want [orph-a orph-z] (canonical status tier, not title)", order)
	}
}

// TestOrphanBucketUsesCanonicalOrderOverTitle is the real discriminator for
// I03: titles are picked so alphabetical order and canonical status order
// DISAGREE ("Alpha" alphabetically first but "scrapped" sorts last) -- this
// fails under the old sortByTitleThenID (title-only) and passes only under
// data.SortBeans.
//
// showArchived is set true here (E5 Task 7, bean bt-ggt2): one of the two
// fixture beans is deliberately "scrapped" to create the sort-order
// discriminator above -- since T7, that status is hidden by DEFAULT
// (beanMatchesArchive, box_filter_facets.go), which would collapse this
// test's own two-orphan setup down to one and defeat its actual point (pure
// canonical-order proof, unrelated to archive visibility). Toggling
// showArchived on keeps both orphans visible so the ordering assertion below
// still exercises what it was written to exercise.
func TestOrphanBucketUsesCanonicalOrderOverTitle(t *testing.T) {
	beans := []data.Bean{
		{ID: "orph-alpha", Title: "Alpha orphan", Status: "scrapped", Type: "task", Priority: "normal", Parent: "missing"},
		{ID: "orph-zulu", Title: "Zulu orphan", Status: "in-progress", Type: "task", Priority: "normal", Parent: "missing"},
	}
	m := fixtureModel(t, beans)
	m.showArchived = true
	m.expanded[orphanRootID] = true
	nodes := m.visibleNodes()
	var order []string
	for _, n := range nodes {
		if n.bean != nil && n.bean.Parent == "missing" {
			order = append(order, n.id)
		}
	}
	if len(order) != 2 || order[0] != "orph-zulu" {
		t.Fatalf("orphan order = %v, want [orph-zulu orph-alpha] (in-progress before scrapped, contradicting alphabetical title order)", order)
	}
}

// --- T8 Opus quality review, Runde 2 (bean bt-7jr8) ---

// TestTabTogglesDetailFocusAndPaneAccent guards Q01 end-to-end: tab flips
// m.detailFocus, AND the rendered pane border actually swaps which side
// carries the Mauve (focused) accent -- not just the model flag. Renders
// through the exact same renderPane/renderDetailPane calls viewBrowseRepo
// itself uses (I05 #1).
func TestTabTogglesDetailFocusAndPaneAccent(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(termenv.Ascii)
	mauve := fgEscape(theme.Mauve)

	m := fixtureModel(t, fixtureBeans())
	m.width, m.height = 80, 24
	if m.detailFocus {
		t.Fatal("detailFocus must start false (Tree focused by default)")
	}

	nodes := m.visibleNodes()
	treePane := renderPane(pane{title: "Tree", rows: m.treeRows(nodes, !m.detailFocus, 10)}, 30, 10, !m.detailFocus)
	detailPane := m.renderDetailPane(nodes, 30, 10, m.detailFocus)
	if !strings.Contains(treePane, mauve) {
		t.Error("Tree pane not accented while Tree has focus")
	}
	if strings.Contains(detailPane, mauve) {
		t.Error("Detail pane accented while Tree still has focus")
	}

	m = step(t, m, keyMsg(tea.KeyTab))
	if !m.detailFocus {
		t.Fatal("tab did not flip detailFocus to true")
	}

	treePane = renderPane(pane{title: "Tree", rows: m.treeRows(nodes, !m.detailFocus, 10)}, 30, 10, !m.detailFocus)
	detailPane = m.renderDetailPane(nodes, 30, 10, m.detailFocus)
	if strings.Contains(treePane, mauve) {
		t.Error("Tree pane still accented after tab moved focus to Detail")
	}
	if !strings.Contains(detailPane, mauve) {
		t.Error("Detail pane not accented after tab moved focus there")
	}
}

// TestCycleBeansShowUnderOrphanRoot guards B02: beans trapped in a pure
// parent cycle (A -> B -> A) are unreachable from any root, but their Parent
// DOES resolve (to each other) -- so collectOrphans alone would never catch
// them either. Both must still surface under the synthetic "(verwaist)"
// root, and flattening a pure cycle must return promptly (no hang).
func TestCycleBeansShowUnderOrphanRoot(t *testing.T) {
	beans := []data.Bean{
		{ID: "cyc-a", Title: "Cycle A", Status: "todo", Type: "task", Priority: "normal", Parent: "cyc-b"},
		{ID: "cyc-b", Title: "Cycle B", Status: "todo", Type: "task", Priority: "normal", Parent: "cyc-a"},
	}
	m := fixtureModel(t, beans) // must return promptly -- no hang on the cycle

	nodes := m.visibleNodes()
	foundOrphanRoot := false
	for _, n := range nodes {
		if n.orphan {
			foundOrphanRoot = true
		}
	}
	if !foundOrphanRoot {
		t.Fatal("orphan root missing even though both beans are unreachable (pure cycle)")
	}
	for _, b := range m.idx.Roots() {
		if b.ID == "cyc-a" || b.ID == "cyc-b" {
			t.Fatal("cycle bean leaked into the real Roots() top level")
		}
	}

	m.expanded[orphanRootID] = true
	nodes = m.visibleNodes()
	seen := map[string]bool{}
	for _, n := range nodes {
		seen[n.id] = true
	}
	if !seen["cyc-a"] || !seen["cyc-b"] {
		t.Fatalf("cycle beans not both visible under (verwaist): nodes=%v", nodeIDs(nodes))
	}
}

// fiftyFlatBeans returns 50 flat (unparented) tasks wnd-00..wnd-49 -- a tree
// taller than any reasonable pane, used to exercise windowing (B01).
func fiftyFlatBeans() []data.Bean {
	beans := make([]data.Bean, 50)
	for i := range beans {
		beans[i] = data.Bean{
			ID:       fmt.Sprintf("wnd-%02d", i),
			Title:    fmt.Sprintf("Window Task %02d", i),
			Status:   "todo",
			Type:     "task",
			Priority: "normal",
		}
	}
	return beans
}

// TestTreeWindowingKeepsCursorVisible guards B01: renderPane alone only
// clips rows to the pane height starting from row 0 -- it has no idea where
// the cursor is. With 50 flat root nodes and a pane far shorter than that,
// the cursored row must still be part of the rendered window (and a row far
// from the cursor must have scrolled out, or the "window" isn't limiting
// anything).
func TestTreeWindowingKeepsCursorVisible(t *testing.T) {
	beans := fiftyFlatBeans()
	m := fixtureModel(t, beans)
	m.width, m.height = 100, 21 // bodyH-2 == 10 visible rows, far less than 50 nodes
	m.cursorID = "wnd-40"

	out := m.View()
	if !strings.Contains(out, "wnd-40") {
		t.Fatalf("cursored row (wnd-40) not visible in View() with 50 nodes / ~10 visible rows -- windowing missing/broken:\n%s", out)
	}
	if strings.Contains(out, "wnd-00") {
		t.Error("row far from the cursor (wnd-00) still visible -- window did not scroll")
	}
}

// --- E2 Task 2 (bean bt-2jve): I01 (expanded-map copy-on-write), Q01 (Init()
// nil-guard covered in app_test.go), Fokus-Maschine (keyDetailFocus). ---

// TestSetExpandedDoesNotMutateSharedMapAcrossModelCopies guards I01 (bean
// bt-7jr8, T8-review): setExpanded must clone m.expanded before writing, not
// mutate the shared backing map in place -- a struct copy (Go map values are
// reference types) must never see the other copy's expand-state change.
func TestSetExpandedDoesNotMutateSharedMapAcrossModelCopies(t *testing.T) {
	base := model{expanded: map[string]bool{}}
	copy1 := base // struct copy -- map HEADER copied, backing array still shared pre-fix
	node := treeNode{id: "x", hasKids: true}
	copy2 := copy1.setExpanded(node, true)

	if copy1.expanded["x"] {
		t.Error("copy1.expanded was mutated by copy2's setExpanded call -- " +
			"expanded is a shared map, not copy-on-write (I01, bean bt-7jr8 T8-review)")
	}
	if !copy2.expanded["x"] {
		t.Error("copy2.expanded should carry the new expand state")
	}
}

// fixtureBeansWithBlocking is a milestone with two epics, bean-a (under
// ep-1, Blocking bean-b) and bean-b (under ep-2) -- used to exercise the
// Beziehungen-field jump + ancestor-expand (bean-b's ancestor ep-2 starts
// collapsed, so the jump must expand it).
func fixtureBeansWithBlocking() []data.Bean {
	return []data.Bean{
		{ID: "ms-1", Title: "Milestone", Status: "todo", Type: "milestone", Priority: "normal"},
		{ID: "ep-1", Title: "Epic One", Status: "todo", Type: "epic", Priority: "normal", Parent: "ms-1"},
		{ID: "ep-2", Title: "Epic Two", Status: "todo", Type: "epic", Priority: "normal", Parent: "ms-1"},
		{ID: "bean-a", Title: "Bean A", Status: "todo", Type: "task", Priority: "normal", Parent: "ep-1", Blocking: []string{"bean-b"}},
		{ID: "bean-b", Title: "Bean B", Status: "todo", Type: "task", Priority: "normal", Parent: "ep-2"},
	}
}

// TestFocusedBeanDispatchesOnTreeCursorInBrowseView guards focusedBean()'s
// Browse-view dispatch (devd port focusedIssue, view_detail_issue.go:20-35):
// in viewBrowseRepo it must return the currently tree-cursored bean.
func TestFocusedBeanDispatchesOnTreeCursorInBrowseView(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m.cursorID = "ms-1"
	b := m.focusedBean()
	if b == nil || b.ID != "ms-1" {
		t.Fatalf("focusedBean() = %v, want ms-1", b)
	}
}

// TestDetailFocusDigitJumpOpensMatchingSection guards the ziffer-jump 1..4:
// each digit jumps straight to its matching section (0-based secCursor,
// 1-based accOpen per renderAccordion's digit header). Table-driven (I02,
// Review-Runde 2, bean bt-2jve) -- previously only '3' had coverage; this
// folds it into the same table alongside 1/2/4.
func TestDetailFocusDigitJumpOpensMatchingSection(t *testing.T) {
	cases := []struct {
		digit         rune
		wantSecCursor int
		wantAccOpen   int
	}{
		{'1', 0, 1},
		{'2', 1, 2},
		{'3', 2, 3},
		{'4', 3, 4},
	}
	for _, tc := range cases {
		t.Run(string(tc.digit), func(t *testing.T) {
			m := fixtureModel(t, fixtureBeans())
			m.cursorID = "ms-1"
			m = step(t, m, keyMsg(tea.KeyTab))
			m = step(t, m, runeMsg(tc.digit))
			if m.secCursor != tc.wantSecCursor || m.accOpen != tc.wantAccOpen {
				t.Fatalf("digit jump '%c': secCursor=%d accOpen=%d, want %d/%d",
					tc.digit, m.secCursor, m.accOpen, tc.wantSecCursor, tc.wantAccOpen)
			}
		})
	}
}

// TestDetailFocusUpDownMovesSectionCursorClampedAtEnds guards section-level
// i/k navigation: 4 fixed sections, down/up both clamp at the ends instead
// of running past them.
func TestDetailFocusUpDownMovesSectionCursorClampedAtEnds(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m.cursorID = "ms-1"
	m = step(t, m, keyMsg(tea.KeyTab))
	for i := 0; i < 5; i++ { // 5 downs on 4 sections must clamp at index 3
		m = step(t, m, keyMsg(tea.KeyDown))
	}
	if m.secCursor != 3 {
		t.Fatalf("secCursor after 5x down = %d, want 3 (clamped)", m.secCursor)
	}
	for i := 0; i < 5; i++ {
		m = step(t, m, keyMsg(tea.KeyUp))
	}
	if m.secCursor != 0 {
		t.Fatalf("secCursor after 5x up = %d, want 0 (clamped)", m.secCursor)
	}
}

// TestDetailFocusRightEntersFieldLevelOnlyForBeziehungenSection guards that
// right/l is a no-op on a fieldless section (Meta) but enters field level on
// Beziehungen (the only section carrying relationFields in E2).
func TestDetailFocusRightEntersFieldLevelOnlyForBeziehungenSection(t *testing.T) {
	m := fixtureModel(t, fixtureBeansWithBlocking())
	m.expanded["ms-1"] = true
	m.expanded["ep-1"] = true
	m.cursorID = "bean-a"
	m = step(t, m, keyMsg(tea.KeyTab)) // secCursor=0 (Meta)
	m = step(t, m, keyMsg(tea.KeyRight))
	if m.detailLevel != 0 {
		t.Fatal("right on Meta (no fields) must stay at section level")
	}
	m = step(t, m, runeMsg('3')) // Beziehungen
	m = step(t, m, keyMsg(tea.KeyRight))
	if m.detailLevel != 1 {
		t.Fatal("right on Beziehungen (has fields) must enter field level")
	}
}

// TestDetailFocusEnterOnRelationJumpsCursorAndExitsToTree guards the
// Beziehungs-Sprung: enter on a resolved relationField moves the tree cursor
// to the target bean, exits detail focus, AND expands the target's ancestor
// chain so it is actually visible in the next visibleNodes() call.
func TestDetailFocusEnterOnRelationJumpsCursorAndExitsToTree(t *testing.T) {
	m := fixtureModel(t, fixtureBeansWithBlocking())
	m.width, m.height = 100, 30
	m.expanded["ms-1"] = true
	m.expanded["ep-1"] = true
	m.cursorID = "bean-a"

	m = step(t, m, keyMsg(tea.KeyTab))
	m = step(t, m, runeMsg('3')) // Beziehungen
	if m.secCursor != 2 || m.accOpen != 3 {
		t.Fatalf("setup: secCursor=%d accOpen=%d, want 2/3", m.secCursor, m.accOpen)
	}
	m = step(t, m, keyMsg(tea.KeyRight)) // field level: fields = [Parent ep-1, Blocking bean-b]
	if m.detailLevel != 1 {
		t.Fatal("setup: expected field level")
	}
	m = step(t, m, keyMsg(tea.KeyDown)) // fieldCursor 0 (ep-1) -> 1 (bean-b)
	m = step(t, m, keyMsg(tea.KeyEnter))

	if m.cursorID != "bean-b" {
		t.Fatalf("cursorID after enter-jump = %q, want bean-b", m.cursorID)
	}
	if m.detailFocus {
		t.Fatal("enter-jump must exit detail focus")
	}
	if !m.expanded["ep-2"] {
		t.Fatal("bean-b's parent (ep-2) must be expanded after the jump so it is visible")
	}
	found := false
	for _, n := range m.visibleNodes() {
		if n.id == "bean-b" {
			found = true
		}
	}
	if !found {
		t.Fatal("bean-b not visible in the tree after the relation-jump (ancestor expand missing)")
	}
}

// TestDetailFocusEnterOnUnresolvedRelationIsNoOp guards the jump-guard: a
// relationField with beanID == "" (dangling reference) must not move the
// cursor or exit detail focus.
func TestDetailFocusEnterOnUnresolvedRelationIsNoOp(t *testing.T) {
	beans := []data.Bean{
		{ID: "dangling-1", Title: "Dangling Bean", Status: "todo", Type: "task", Priority: "normal",
			BlockedBy: []string{"does-not-exist"}},
	}
	m := fixtureModel(t, beans)
	m.width, m.height = 100, 30
	m.cursorID = "dangling-1"

	m = step(t, m, keyMsg(tea.KeyTab))
	m = step(t, m, runeMsg('3')) // Beziehungen
	m = step(t, m, keyMsg(tea.KeyRight))
	if m.detailLevel != 1 {
		t.Fatal("setup: expected field level")
	}
	m = step(t, m, keyMsg(tea.KeyEnter))

	if !m.detailFocus {
		t.Fatal("enter on an unresolved relation must NOT exit detail focus")
	}
	if m.detailLevel != 1 {
		t.Fatal("enter on an unresolved relation must stay in field level")
	}
	if m.cursorID != "dangling-1" {
		t.Fatal("enter on an unresolved relation must not move the tree cursor")
	}
}

// TestDetailFocusLeftAtFieldLevelReturnsToSectionLevel guards left/j at
// field level: it steps back to section level without exiting detail focus.
func TestDetailFocusLeftAtFieldLevelReturnsToSectionLevel(t *testing.T) {
	m := fixtureModel(t, fixtureBeansWithBlocking())
	m.expanded["ms-1"] = true
	m.expanded["ep-1"] = true
	m.cursorID = "bean-a"

	m = step(t, m, keyMsg(tea.KeyTab))
	m = step(t, m, runeMsg('3'))
	m = step(t, m, keyMsg(tea.KeyRight))
	if m.detailLevel != 1 {
		t.Fatal("setup: expected field level")
	}
	m = step(t, m, keyMsg(tea.KeyLeft))
	if m.detailLevel != 0 {
		t.Fatal("left at field level must return to section level")
	}
	if !m.detailFocus {
		t.Fatal("left at field level must NOT exit detail focus")
	}
}

// TestDetailFocusLeftAtSectionLevelExitsDetailFocus guards left/j at section
// level: it exits detail focus entirely (back to Tree focus).
func TestDetailFocusLeftAtSectionLevelExitsDetailFocus(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m.cursorID = "ms-1"
	m = step(t, m, keyMsg(tea.KeyTab))
	if !m.detailFocus {
		t.Fatal("setup: expected detail focus on")
	}
	m = step(t, m, keyMsg(tea.KeyLeft))
	if m.detailFocus {
		t.Fatal("left at section level must exit detail focus")
	}
}

// TestKeyDetailFocusOnOrphanRootExitsGracefully is a defensive nil-safety
// test for focusedBean()'s orphan-guard: cursoring the synthetic
// "(verwaist)" root itself (not a real bean) must never panic and must exit
// detail focus rather than getting stuck.
func TestKeyDetailFocusOnOrphanRootExitsGracefully(t *testing.T) {
	beans := append(fixtureBeans(), data.Bean{
		ID: "orph-1", Title: "Orphan", Status: "todo", Type: "task", Priority: "normal", Parent: "missing",
	})
	m := fixtureModel(t, beans)
	m.cursorID = orphanRootID

	m = step(t, m, keyMsg(tea.KeyTab))
	m = step(t, m, keyMsg(tea.KeyRight)) // must not panic
	if m.detailFocus {
		t.Fatal("keyDetailFocus on an orphan-root cursor (no focusable bean) must exit detail focus")
	}
}

// TestWindowAroundStableAtEdges guards windowStart/windowAround directly
// (devd port): fits-entirely is a no-op, the window never runs past either
// edge, and the cursor stays inside the returned window at every position --
// no jitter, no out-of-range slice.
func TestWindowAroundStableAtEdges(t *testing.T) {
	rows := make([]string, 20)
	for i := range rows {
		rows[i] = fmt.Sprintf("row-%02d", i)
	}

	if got := windowAround(rows, 30, 5); len(got) != 20 {
		t.Fatalf("windowAround with height >= len(rows) must be a no-op, got %d rows", len(got))
	}

	for _, cursor := range []int{0, 1, 5, 10, 14, 15, 19} {
		win := windowAround(rows, 8, cursor)
		if len(win) != 8 {
			t.Fatalf("cursor=%d: window has %d rows, want 8", cursor, len(win))
		}
		want := fmt.Sprintf("row-%02d", cursor)
		found := false
		for _, r := range win {
			if r == want {
				found = true
			}
		}
		if !found {
			t.Errorf("cursor=%d: cursored row %q not in window %v", cursor, want, win)
		}
	}
}

// --- E2-T2 Quality-Review, Runde 2 (bean bt-2jve) ---

// TestExpandAncestorsOfHandlesParentCycle guards B01 (Critical, Review-Runde
// 2): expandAncestorsOf's ancestor walk had no visited-set, so a Parent
// cycle in the on-disk data (A -> B -> A -- beans-legal, frontmatter is
// hand-editable) hung it forever on any relation-jump into the cycle,
// freezing the whole TUI. Run with a bounded `go test -timeout` so a
// regression fails loudly (timeout) instead of hanging the suite.
func TestExpandAncestorsOfHandlesParentCycle(t *testing.T) {
	beans := []data.Bean{
		{ID: "cyc-a", Title: "Cycle A", Status: "todo", Type: "task", Priority: "normal", Parent: "cyc-b"},
		{ID: "cyc-b", Title: "Cycle B", Status: "todo", Type: "task", Priority: "normal", Parent: "cyc-a"},
	}
	idx := data.NewIndex(beans)

	out := expandAncestorsOf(idx, map[string]bool{}, "cyc-a")
	if !out["cyc-b"] {
		t.Error("expandAncestorsOf did not mark cyc-a's immediate parent (cyc-b) expanded")
	}
}

// TestFieldCursorClampsAfterReloadShrinksRelations guards B02 (Critical,
// Review-Runde 2): m.secCursor/m.fieldCursor are model state that survives a
// beansLoadedMsg reload untouched -- if a watch-reload shrinks the focused
// bean's Beziehungen fields while the user is parked at field level,
// keyDetailFocus's `secs[m.secCursor].fields[m.fieldCursor]` indexed out of
// range and panicked. Repro: drill into field 1 of 2, reload down to 1
// relation, press enter -- must not panic, must clamp to a sane state.
func TestFieldCursorClampsAfterReloadShrinksRelations(t *testing.T) {
	beans := []data.Bean{
		{ID: "root-1", Title: "Root", Status: "todo", Type: "task", Priority: "normal",
			Blocking: []string{"blk-1", "blk-2"}},
		{ID: "blk-1", Title: "Blocker One", Status: "todo", Type: "task", Priority: "normal"},
		{ID: "blk-2", Title: "Blocker Two", Status: "todo", Type: "task", Priority: "normal"},
	}
	m := fixtureModel(t, beans)
	m.width, m.height = 100, 30
	m.cursorID = "root-1"

	m = step(t, m, keyMsg(tea.KeyTab))
	m = step(t, m, runeMsg('3'))         // Beziehungen
	m = step(t, m, keyMsg(tea.KeyRight)) // field level, fieldCursor 0 (blk-1)
	if m.detailLevel != 1 {
		t.Fatal("setup: expected field level")
	}
	m = step(t, m, keyMsg(tea.KeyDown)) // fieldCursor 1 (blk-2) -- field 2 of 2
	if m.fieldCursor != 1 {
		t.Fatalf("setup: fieldCursor = %d, want 1", m.fieldCursor)
	}

	// A watch-reload shrinks root-1's Blocking down to 1 entry while the
	// user is still parked at field index 1 -- fieldCursor is now stale
	// relative to the freshly computed secs.
	shrunk := []data.Bean{
		{ID: "root-1", Title: "Root", Status: "todo", Type: "task", Priority: "normal",
			Blocking: []string{"blk-1"}},
		{ID: "blk-1", Title: "Blocker One", Status: "todo", Type: "task", Priority: "normal"},
	}
	m = step(t, m, beansLoadedMsg{beans: shrunk})

	// A key press right after the reload (no jump side-effect: "up" only
	// decrements if fieldCursor > 0) must clamp the stale fieldCursor into
	// range for the shrunk section without panicking -- direct clamp
	// evidence, still focused on root-1.
	m = step(t, m, keyMsg(tea.KeyUp))
	if m.fieldCursor != 0 {
		t.Fatalf("fieldCursor after reload+clamp = %d, want 0 (only 1 field left)", m.fieldCursor)
	}

	// Enter must not panic either, even though fieldCursor was stale
	// relative to the pre-reload (2-field) section shape: clamped to the
	// sole remaining (resolved) relation, it now behaves like a normal
	// relation-jump -- itself sane B02 behavior, not a special case.
	m = step(t, m, keyMsg(tea.KeyEnter))
	if m.fieldCursor < 0 {
		t.Fatalf("fieldCursor went negative: %d", m.fieldCursor)
	}
}

// TestTabReentryResetsStaleDetailFocusState guards I01 (Important,
// Review-Runde 2): re-entering detail focus after having drilled into a deep
// section/field state must never leak that state into the new visit -- tab
// always re-enters the accordion at Meta/section level/field 0 (Q01,
// handleKey's tab case, update.go). Previously untested end-to-end.
func TestTabReentryResetsStaleDetailFocusState(t *testing.T) {
	m := fixtureModel(t, fixtureBeansWithBlocking())
	m.expanded["ms-1"] = true
	m.expanded["ep-1"] = true
	m.cursorID = "bean-a"

	m = step(t, m, keyMsg(tea.KeyTab)) // tab in
	m = step(t, m, runeMsg('3'))       // section 3 (Beziehungen)
	m = step(t, m, keyMsg(tea.KeyRight))
	if m.secCursor != 2 || m.accOpen != 3 || m.detailLevel != 1 {
		t.Fatalf("setup: secCursor=%d accOpen=%d detailLevel=%d, want 2/3/1",
			m.secCursor, m.accOpen, m.detailLevel)
	}

	m = step(t, m, keyMsg(tea.KeyTab)) // tab out
	if m.detailFocus {
		t.Fatal("setup: tab did not exit detail focus")
	}
	m = step(t, m, keyMsg(tea.KeyTab)) // tab back in
	if !m.detailFocus {
		t.Fatal("setup: tab did not re-enter detail focus")
	}
	if m.secCursor != 0 || m.accOpen != 1 || m.detailLevel != 0 || m.fieldCursor != 0 {
		t.Fatalf("stale detail-focus state leaked across tab re-entry: secCursor=%d accOpen=%d detailLevel=%d fieldCursor=%d, want 0/1/0/0",
			m.secCursor, m.accOpen, m.detailLevel, m.fieldCursor)
	}
}

// TestRepoSwitchClearsToast is the T6-Review Prelude I01 regression guard
// (T6b, bean bt-pd22): a Toast surfaced against the OLD repo -- sticky or
// not -- must not linger visible after a successful repo switch.
// applyRepoSwitched clears it UNCONDITIONALLY (update.go), unlike
// applyLoaded's clearToastUnlessSticky (which protects a sticky toast
// across a SAME-repo background reload) -- a repo switch is a bigger
// session discontinuity, so even the one sticky ErrConflict toast kind must
// not survive it. sticky=true here is the worst case (the ONE toast kind
// that otherwise survives every OTHER reload path, showToast's own doc
// comment).
func TestRepoSwitchClearsToast(t *testing.T) {
	t.Setenv("HOME", t.TempDir()) // applyRepoSwitched's best-effort config.SetLastRepo must not touch the real $HOME

	m := fixtureModel(t, fixtureBeans())
	m, _ = m.showToast(toastError, "Konflikt: Bean extern geändert", "", nil, true)
	if m.toast == nil {
		t.Fatal("setup: showToast did not set m.toast")
	}

	newRepo := t.TempDir()
	tm, _ := m.applyRepoSwitched(repoSwitchedMsg{
		client:  &data.Client{RepoDir: newRepo},
		repoDir: newRepo,
		beans:   []data.Bean{},
	})
	m2, ok := tm.(model)
	if !ok {
		t.Fatalf("applyRepoSwitched did not return a model, got %T", tm)
	}
	if m2.toast != nil {
		t.Fatalf("m.toast = %+v after a successful repo switch, want nil (even a sticky toast from the OLD repo must clear)", m2.toast)
	}
}
