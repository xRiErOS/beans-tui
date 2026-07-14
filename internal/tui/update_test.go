package tui

// update_test.go — TDD coverage for the App-Shell Update dispatcher (T8,
// implementation-plan.md »E1 Task 8«): cursor movement/expand/collapse,
// quit-confirm, reload-keeps-cursor-by-ID, and the MANDATORY orphan rule
// (bean bt-7jr8 / T3-review Q01).

import (
	"strings"
	"testing"

	"beans-tui/internal/data"
	tea "github.com/charmbracelet/bubbletea"
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
