package tui

// view_fullscreen_test.go — TDD coverage for F01 Kernmechanik: Vollbild-Modus
// `v` (design-spec.md §15 "F01 — Vollbild-Navigation", bean bt-13l7, E9 Task
// 7). OHNE den History-Stack (Task 8, bean bt-1vbp). Reuses fixtureBeans/
// fixtureBeansWithBlocking/fixtureModel/step/keyMsg/runeMsg (update_test.go)
// and backlogBeans (view_browse_backlog_test.go) -- same package.
//
// focusedBean()/activateDetailField()/keyDetailFocus()'s own fullscreenDetail
// integration is covered in update_test.go instead (co-located with those
// functions' existing test coverage); the mouse guard is covered in
// mouse_test.go (co-located with handleMouse's existing tests). This file
// covers keyFullscreen itself + renderFullscreenBody.

import (
	"strings"
	"testing"

	"beans-tui/internal/data"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// fullscreenBacklogBeans is a small parentless, Backlog-eligible fixture
// carrying a Blocking relation between its two members -- used to exercise
// F01's Backlog symmetry (a Relations-Sprung inside fullscreenDetail while
// m.view == viewBacklog, and the esc-exit's backlogList cursor sync).
func fullscreenBacklogBeans() []data.Bean {
	return []data.Bean{
		{ID: "fb-a", Title: "Backlog A", Status: "todo", Type: "task", Priority: "normal", Blocking: []string{"fb-b"}},
		{ID: "fb-b", Title: "Backlog B", Status: "todo", Type: "task", Priority: "normal"},
	}
}

// --- Einstieg (v) ---

// TestKeyFullscreenEntersListModeWhenTreeFocused guards the PO-Wortlaut's
// first case: Browse+links-fokussiert (m.detailFocus == false) -> Beans-
// Liste Vollbild.
func TestKeyFullscreenEntersListModeWhenTreeFocused(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	if m.detailFocus {
		t.Fatal("setup: expected Tree focus (detailFocus false)")
	}

	m = step(t, m, runeMsg('v'))

	if m.fullscreen != fullscreenList {
		t.Fatalf("fullscreen = %v, want fullscreenList", m.fullscreen)
	}
}

// TestKeyFullscreenEntersDetailModeWhenDetailFocused guards the PO-Wortlaut's
// second case: Browse+rechts-fokussiert (m.detailFocus == true) -> Detail-
// View Vollbild, targeting the currently focused bean.
func TestKeyFullscreenEntersDetailModeWhenDetailFocused(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m.cursorID = "ms-1"
	m = step(t, m, keyMsg(tea.KeyTab)) // detailFocus = true

	m = step(t, m, runeMsg('v'))

	if m.fullscreen != fullscreenDetail {
		t.Fatalf("fullscreen = %v, want fullscreenDetail", m.fullscreen)
	}
	if m.fullscreenBeanID != "ms-1" {
		t.Fatalf("fullscreenBeanID = %q, want ms-1", m.fullscreenBeanID)
	}
}

// TestKeyFullscreenNoOpWhenAlreadyFullscreenList guards the Planner
// decision: v is a EINWEG-Einstieg, not a toggle -- a second v while already
// in fullscreenList changes nothing.
func TestKeyFullscreenNoOpWhenAlreadyFullscreenList(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m.fullscreen = fullscreenList

	m = step(t, m, runeMsg('v'))

	if m.fullscreen != fullscreenList {
		t.Fatalf("fullscreen = %v, want unchanged fullscreenList (v is one-way, no toggle)", m.fullscreen)
	}
}

// TestKeyFullscreenNoOpWhenAlreadyFullscreenDetail mirrors the List case for
// fullscreenDetail -- v must not re-target a DIFFERENT bean even if the
// Tree/Backlog cursor has since moved.
func TestKeyFullscreenNoOpWhenAlreadyFullscreenDetail(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m.fullscreen = fullscreenDetail
	m.fullscreenBeanID = "ms-1"
	m.cursorID = "tk-2" // deliberately a DIFFERENT bean

	m = step(t, m, runeMsg('v'))

	if m.fullscreen != fullscreenDetail || m.fullscreenBeanID != "ms-1" {
		t.Fatalf("fullscreen=%v fullscreenBeanID=%q, want unchanged fullscreenDetail/ms-1", m.fullscreen, m.fullscreenBeanID)
	}
}

// TestKeyFullscreenNoOpInLobby guards keyFullscreen's own defensive
// m.view == viewLobby guard directly (handleKey's earlier viewLobby capture
// already makes this unreachable via the full Update() dispatch -- this
// calls keyFullscreen directly to pin the function's own contract).
func TestKeyFullscreenNoOpInLobby(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m.view = viewLobby

	handled, tm, cmd := m.keyFullscreen(runeMsg('v'))

	if handled {
		t.Fatal("keyFullscreen must not claim to handle v while m.view == viewLobby")
	}
	nm, ok := tm.(model)
	if !ok {
		t.Fatalf("keyFullscreen did not return a model, got %T", tm)
	}
	if nm.fullscreen != fullscreenNone {
		t.Fatalf("fullscreen = %v, want fullscreenNone (unchanged)", nm.fullscreen)
	}
	if cmd != nil {
		t.Fatal("keyFullscreen(lobby) must not return a Cmd")
	}
}

// --- Listen-Vollbild -> Detail-Vollbild (enter) ---

// TestKeyFullscreenEnterOnListEntersDetailMode guards enter on a Tree bean
// while m.fullscreen == fullscreenList: switches to fullscreenDetail
// targeting the cursored bean AND reinitializes the Detail-Fokus-Maschine
// (Meta/section level, field cursor 0) -- a stale position from a previous
// visit must not leak in.
func TestKeyFullscreenEnterOnListEntersDetailMode(t *testing.T) {
	m := fixtureModel(t, fixtureBeansWithBlocking())
	m.cursorID = "ms-1"
	m.fullscreen = fullscreenList
	m.secCursor, m.accOpen, m.detailLevel, m.fieldCursor = 2, 3, 1, 1 // stale, must be reset

	m = step(t, m, keyMsg(tea.KeyEnter))

	if m.fullscreen != fullscreenDetail {
		t.Fatalf("fullscreen = %v, want fullscreenDetail", m.fullscreen)
	}
	if m.fullscreenBeanID != "ms-1" {
		t.Fatalf("fullscreenBeanID = %q, want ms-1", m.fullscreenBeanID)
	}
	if m.secCursor != 0 || m.accOpen != 1 || m.detailLevel != 0 || m.fieldCursor != 0 {
		t.Fatalf("Detail-Fokus-Maschine not reset: secCursor=%d accOpen=%d detailLevel=%d fieldCursor=%d, want 0/1/0/0",
			m.secCursor, m.accOpen, m.detailLevel, m.fieldCursor)
	}
}

// TestKeyFullscreenEnterOnListEntersDetailModeBacklog is
// TestKeyFullscreenEnterOnListEntersDetailMode's Backlog mirror (Akzeptanz:
// "Symmetrisch für Browse (Tree) UND Backlog").
func TestKeyFullscreenEnterOnListEntersDetailModeBacklog(t *testing.T) {
	m := fixtureModel(t, fullscreenBacklogBeans())
	m.view = viewBacklog
	vis := m.backlogVisible()
	if len(vis) == 0 {
		t.Fatal("setup: need at least 1 backlog-visible bean")
	}
	m.backlogList.setLen(len(vis))
	m.backlogList.cursor = 0
	want := vis[0].ID
	m.fullscreen = fullscreenList

	m = step(t, m, keyMsg(tea.KeyEnter))

	if m.fullscreen != fullscreenDetail {
		t.Fatalf("fullscreen = %v, want fullscreenDetail", m.fullscreen)
	}
	if m.fullscreenBeanID != want {
		t.Fatalf("fullscreenBeanID = %q, want %q", m.fullscreenBeanID, want)
	}
	if m.view != viewBacklog {
		t.Fatalf("view = %v, want unchanged viewBacklog (F01 never changes m.view)", m.view)
	}
}

// TestKeyFullscreenEnterOnEmptyListIsNoOp guards the orphan-root cursor case
// (focusedBean() returns nil): enter in fullscreenList must not crash or
// enter fullscreenDetail with an empty target. Needs a REAL dangling-parent
// bean in the fixture (mirrors TestKeyDetailFocusOnOrphanRootExitsGracefully,
// above) -- otherwise flattenTree never emits the synthetic "(orphaned)" root
// node at all, and m.cursorID=orphanRootID would silently resolve to
// cursorPos's no-match fallback (index 0, a REAL bean) instead.
func TestKeyFullscreenEnterOnEmptyListIsNoOp(t *testing.T) {
	beans := append(fixtureBeans(), data.Bean{
		ID: "orph-1", Title: "Orphan", Status: "todo", Type: "task", Priority: "normal", Parent: "missing",
	})
	m := fixtureModel(t, beans)
	m.cursorID = orphanRootID
	m.fullscreen = fullscreenList

	m = step(t, m, keyMsg(tea.KeyEnter))

	if m.fullscreen != fullscreenList {
		t.Fatalf("fullscreen = %v, want unchanged fullscreenList (orphan-root cursor has no focusable bean)", m.fullscreen)
	}
	if m.fullscreenBeanID != "" {
		t.Fatalf("fullscreenBeanID = %q, want empty", m.fullscreenBeanID)
	}
}

// --- Ausstieg (esc) aus fullscreenList ---

// TestKeyFullscreenEscExitsListModeToSplitView guards the direct
// fullscreenList exit: no cursor sync needed, the cursor was never decoupled
// while in fullscreenList.
func TestKeyFullscreenEscExitsListModeToSplitView(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m.cursorID = "ms-1"
	m.fullscreen = fullscreenList

	m = step(t, m, keyMsg(tea.KeyEsc))

	if m.fullscreen != fullscreenNone {
		t.Fatalf("fullscreen = %v, want fullscreenNone", m.fullscreen)
	}
	if m.cursorID != "ms-1" {
		t.Fatalf("cursorID = %q, want unchanged ms-1", m.cursorID)
	}
}

// TestKeyFullscreenEscExitsListModeToSplitViewBacklog is the Backlog mirror.
func TestKeyFullscreenEscExitsListModeToSplitViewBacklog(t *testing.T) {
	m := fixtureModel(t, backlogBeans())
	m.view = viewBacklog
	m.fullscreen = fullscreenList
	m.backlogList.cursor = 1

	m = step(t, m, keyMsg(tea.KeyEsc))

	if m.fullscreen != fullscreenNone {
		t.Fatalf("fullscreen = %v, want fullscreenNone", m.fullscreen)
	}
	if m.view != viewBacklog {
		t.Fatalf("view = %v, want unchanged viewBacklog", m.view)
	}
	if m.backlogList.cursor != 1 {
		t.Fatalf("backlogList.cursor = %d, want unchanged 1", m.backlogList.cursor)
	}
}

// --- Kein neuer viewID-Wert (Akzeptanz-Checkliste) ---

// TestFullscreenNeverChangesViewID guards the state-model's central
// invariant: fullscreen is orthogonal to m.view, never mutates it.
func TestFullscreenNeverChangesViewID(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m.cursorID = "ms-1"
	before := m.view

	m = step(t, m, runeMsg('v'))

	if m.view != before {
		t.Fatalf("view changed from %v to %v -- F01 must be orthogonal to m.view", before, m.view)
	}
}

// --- Rendering ---

// TestRenderFullscreenBodyListUsesFullPaneWidth is a direct, render-grounded
// width assert (no -update Golden needed): renderPane's Width(w) sets the
// CONTENT width, the RoundedBorder adds 2 more columns on top -- rendered
// total width must equal innerW+2, proving the fullscreenList branch uses
// the FULL innerW (not a Split lw/rw).
func TestRenderFullscreenBodyListUsesFullPaneWidth(t *testing.T) {
	rows := []string{"row one", "row two"}
	out := renderFullscreenBody(fullscreenList, 40, 10, rows, true, nil, nil, 0, 1, 0, 0)
	lines := strings.Split(out, "\n")
	if len(lines) == 0 {
		t.Fatal("renderFullscreenBody returned no lines")
	}
	if w := lipgloss.Width(lines[0]); w != 42 {
		t.Fatalf("renderFullscreenBody(fullscreenList) rendered width = %d, want 42 (innerW 40 + 2 border columns)", w)
	}
}

// TestRenderFullscreenBodyDetailUsesFullPaneWidth mirrors the List case for
// fullscreenDetail (renderAccordionPane's own Width(w) + border contract).
func TestRenderFullscreenBodyDetailUsesFullPaneWidth(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	b := m.idx.ByID["ms-1"]

	out := renderFullscreenBody(fullscreenDetail, 50, 12, nil, true, m.idx, b, 0, 1, 0, 0)
	lines := strings.Split(out, "\n")
	if len(lines) == 0 {
		t.Fatal("renderFullscreenBody returned no lines")
	}
	if w := lipgloss.Width(lines[0]); w != 52 {
		t.Fatalf("renderFullscreenBody(fullscreenDetail) rendered width = %d, want 52 (innerW 50 + 2 border columns)", w)
	}
}

// TestViewBrowseRepoFullscreenFitsOuterFrame is a regression guard for a REAL
// bug the tmux smoke caught live (not visible to
// TestRenderFullscreenBodyListUsesFullPaneWidth/...DetailUsesFullPaneWidth
// above, which only exercise renderFullscreenBody in ISOLATION): the call
// site originally passed innerW straight through as the single pane's
// content width, double-counting the pane's OWN RoundedBorder against the
// outer frame's border and overflowing it by 2 columns/rows -- visibly
// broke the outer frame in a real terminal (the bottom border wrapped onto
// its own line). Render-grounded at the FULL m.View() level, both
// fullscreen flavors, so a future regression at either call site (Tree OR
// Backlog) is caught the same way this one was.
func TestViewBrowseRepoFullscreenFitsOuterFrame(t *testing.T) {
	for _, fs := range []fullscreenMode{fullscreenList, fullscreenDetail} {
		m := fixtureModel(t, fixtureBeans())
		m = step(t, m, tea.WindowSizeMsg{Width: 100, Height: 30})
		m.cursorID = "ms-1"
		m.fullscreen = fs
		m.fullscreenBeanID = "ms-1"

		out := m.View()
		lines := strings.Split(out, "\n")
		if len(lines) > 30 {
			t.Fatalf("fullscreen=%v: View() produced %d lines, want <= 30 (outer frame overflow)", fs, len(lines))
		}
		for i, l := range lines {
			if w := lipgloss.Width(l); w > 100 {
				t.Fatalf("fullscreen=%v: line %d width = %d, want <= 100 (outer frame overflow): %q", fs, i, w, l)
			}
		}
	}
}

// TestViewBacklogFullscreenFitsOuterFrame is the Backlog mirror.
func TestViewBacklogFullscreenFitsOuterFrame(t *testing.T) {
	for _, fs := range []fullscreenMode{fullscreenList, fullscreenDetail} {
		m := fixtureModel(t, backlogBeans())
		m.view = viewBacklog
		m = step(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
		vis := m.backlogVisible()
		if len(vis) == 0 {
			t.Fatal("setup: need at least 1 backlog-visible bean")
		}
		m.backlogList.setLen(len(vis))
		m.backlogList.cursor = 0
		m.fullscreen = fs
		m.fullscreenBeanID = vis[0].ID

		out := m.View()
		lines := strings.Split(out, "\n")
		if len(lines) > 24 {
			t.Fatalf("fullscreen=%v: View() produced %d lines, want <= 24 (outer frame overflow)", fs, len(lines))
		}
		for i, l := range lines {
			if w := lipgloss.Width(l); w > 80 {
				t.Fatalf("fullscreen=%v: line %d width = %d, want <= 80 (outer frame overflow): %q", fs, i, w, l)
			}
		}
	}
}
