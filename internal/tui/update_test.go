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
