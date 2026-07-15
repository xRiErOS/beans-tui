package tui

// view_lobby_test.go — E5 Task 6 (bean bt-zhwl): TDD coverage for the Lobby
// V1 view + Repo-Picker `p` + the design-decision-d startup-trigger
// invariant at the model layer.

import (
	"strings"
	"testing"

	"beans-tui/internal/config"
	"beans-tui/internal/data"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// lobbyFixtureModel builds a model already positioned in the Lobby, with
// Settings.Repos set and repoList/repoSearch primed the way openLobby()
// itself would prime them (mirrors fixtureModel's own "already past its
// initial transition" convention, update_test.go). openLobby() reloads
// Settings from disk (bean bt-zhwl's own "lädt Settings.Repos neu, falls
// seit Start geändert" acceptance wording -- caught missing during this
// task's own tmux smoke test) -- so repos must be seeded via a REAL,
// isolated config.yaml (t.Setenv("HOME", ...) + config.SaveUserSettings),
// not just a direct model-field assignment that openLobby's own reload
// would immediately clobber.
func lobbyFixtureModel(t *testing.T, repos []string) model {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	if err := config.SaveUserSettings(repos, "", "", 0); err != nil {
		t.Fatalf("SaveUserSettings: %v", err)
	}
	m := newModel(nil, "/tmp/bt-fixture-repo")
	nm, _ := m.openLobby()
	return nm.(model)
}

func TestLobbyShowsConfiguredRepos(t *testing.T) {
	m := lobbyFixtureModel(t, []string{"/tmp/repo-alpha", "/tmp/repo-beta"})
	out := m.viewLobby()
	if !strings.Contains(out, "/tmp/repo-alpha") {
		t.Errorf("viewLobby() output missing repo-alpha:\n%s", out)
	}
	if !strings.Contains(out, "/tmp/repo-beta") {
		t.Errorf("viewLobby() output missing repo-beta:\n%s", out)
	}
}

func TestLobbyFilterNarrowsBySearch(t *testing.T) {
	m := lobbyFixtureModel(t, []string{"/tmp/repo-alpha", "/tmp/repo-beta"})

	for _, r := range "beta" {
		nm, _ := m.keyLobby(runeMsg(r))
		m = nm.(model)
	}
	if m.repoQuery != "beta" {
		t.Fatalf("repoQuery = %q, want %q", m.repoQuery, "beta")
	}
	filtered := m.filteredRepos()
	if len(filtered) != 1 || filtered[0] != "/tmp/repo-beta" {
		t.Fatalf("filteredRepos() = %v, want exactly [/tmp/repo-beta]", filtered)
	}

	out := m.viewLobby()
	if strings.Contains(out, "repo-alpha") {
		t.Errorf("viewLobby() still shows repo-alpha after filtering to %q:\n%s", m.repoQuery, out)
	}
	if !strings.Contains(out, "repo-beta") {
		t.Errorf("viewLobby() missing repo-beta after filtering to %q:\n%s", m.repoQuery, out)
	}
}

func TestLobbySelectSwitchesRepoAndView(t *testing.T) {
	fakeBeansOnPath(t, "#!/bin/sh\necho '[]'\n")
	repoA := newTestRepoTUI(t)
	repoB := newTestRepoTUI(t)

	m := lobbyFixtureModel(t, []string{repoA, repoB})
	m.repoList.cursor = 1 // repoB

	nm, cmd := m.keyLobby(tea.KeyMsg{Type: tea.KeyEnter})
	m = nm.(model)
	if cmd == nil {
		t.Fatal("keyLobby(enter) returned a nil cmd, want switchRepoCmd's tea.Cmd")
	}

	msg := cmd()
	rs, ok := msg.(repoSwitchedMsg)
	if !ok {
		t.Fatalf("cmd() returned %T, want repoSwitchedMsg", msg)
	}
	if rs.err != nil {
		t.Fatalf("repoSwitchedMsg.err = %v, want nil", rs.err)
	}
	if rs.repoDir != repoB {
		t.Fatalf("repoSwitchedMsg.repoDir = %q, want %q (repoB, cursor 1)", rs.repoDir, repoB)
	}
	t.Cleanup(func() {
		if rs.watchStop != nil {
			rs.watchStop()
		}
	})

	final := step(t, m, msg)
	if final.view != viewBrowseRepo {
		t.Fatalf("view after repoSwitchedMsg = %v, want viewBrowseRepo", final.view)
	}
	if final.client == nil {
		t.Fatal("client is nil after a successful repo switch")
	}
	if final.repoDir != repoB {
		t.Fatalf("repoDir = %q, want %q", final.repoDir, repoB)
	}
	if final.idx == nil {
		t.Fatal("idx is nil after a successful repo switch")
	}
	if final.watchStop == nil {
		t.Fatal("watchStop is nil after a successful repo switch with a working watcher")
	}
	t.Cleanup(final.watchStop)
}

func TestPickerKeyOpensLobbyFromAnyView(t *testing.T) {
	beans := fixtureBeans()
	// openLobby() reloads Settings from disk (its own doc-stamp) -- isolate
	// HOME so this test never depends on whatever (if anything) happens to
	// exist at the real dev machine's ~/.config/beans-tui/config.yaml.
	t.Setenv("HOME", t.TempDir())

	t.Run("from Backlog", func(t *testing.T) {
		m := fixtureModel(t, beans)
		m.view = viewBacklog
		out := step(t, m, runeMsg('p'))
		if out.view != viewLobby {
			t.Fatalf("view after 'p' from Backlog = %v, want viewLobby", out.view)
		}
	})

	t.Run("from Browse", func(t *testing.T) {
		m := fixtureModel(t, beans)
		m.view = viewBrowseRepo
		out := step(t, m, runeMsg('p'))
		if out.view != viewLobby {
			t.Fatalf("view after 'p' from Browse = %v, want viewLobby", out.view)
		}
	})

	// Deliberate exception (update.go handleKey's own doc-stamp on the
	// keys.Picker match check): the Review-Cockpit already binds the bare
	// "p" key to its own pre-existing design-spec §7 "explicit prev"
	// meaning -- the global Picker must NOT shadow it, unlike ctrl+k/`?`
	// which have no such collision.
	t.Run("NOT from Review-Cockpit (p is already bound to prev there)", func(t *testing.T) {
		m := fixtureModel(t, beans)
		nm, _ := m.openReviewCockpit()
		m = nm.(model)
		if m.view != viewReviewCockpit {
			t.Fatal("test setup invalid: openReviewCockpit() did not switch view")
		}
		out := step(t, m, runeMsg('p'))
		if out.view == viewLobby {
			t.Fatal("'p' opened the Lobby from inside the Review-Cockpit -- it must stay bound to the Cockpit's own 'prev' action")
		}
	})
}

// TestViewLobbyFrameMatchesWidthHeight is the regression guard for the
// height/width overflow bug found live during this task's own tmux smoke
// test (viewLobby's own doc-stamp, "B01", has the full story: the first cut
// handed lipgloss.Place/outerBorder the FULL w/h instead of budgeting
// innerW/innerH first, so the rendered frame silently overflowed m.height/
// m.width by 2 in each dimension -- invisible to every OTHER test here since
// none diffed the rendered string's own dimensions against the input).
// Mirrors TestChromeNeverOverflowsWidth's own pattern (chrome_test.go) --
// every OTHER view already has an equivalent guard, the Lobby is the fourth.
func TestViewLobbyFrameMatchesWidthHeight(t *testing.T) {
	m := lobbyFixtureModel(t, []string{"/tmp/repo-alpha", "/tmp/a-much-longer-repo-path-beta"})
	for _, dims := range []struct{ w, h int }{{30, 24}, {60, 20}, {100, 30}, {200, 50}} {
		m.width, m.height = dims.w, dims.h
		out := m.viewLobby()
		if got := lipgloss.Height(out); got != dims.h {
			t.Errorf("w=%d h=%d: viewLobby() height = %d, want %d", dims.w, dims.h, got, dims.h)
		}
		for i, ln := range strings.Split(out, "\n") {
			if lw := lipgloss.Width(ln); lw > dims.w {
				t.Errorf("w=%d h=%d: line %d overflows (%d > %d)", dims.w, dims.h, i, lw, dims.w)
			}
		}
	}
}

// TestNoLobbyOnSingleRepoCwdMatch guards the MODEL-layer half of design
// decision d's own invariant (the cmd/tui.go trigger decision itself is
// covered independently by cmd/tui_test.go's TestDecideStartupPrioritiesInOrder,
// which cannot invoke tui.Run -- an interactive AltScreen program -- from a
// `command go test` process): newModel/Init NEVER auto-derive viewLobby from
// Settings.Repos' size themselves -- that decision is made EXCLUSIVELY by
// Run()'s own `client == nil` branch (app.go), one layer above. A model
// built with a REAL (non-nil) client -- the E1-E4 "already resolved a repo"
// shape, regardless of how many repos Settings.Repos lists -- must stay on
// viewBrowseRepo, exactly like every pre-Task-6 session.
func TestNoLobbyOnSingleRepoCwdMatch(t *testing.T) {
	m := newModel(&data.Client{RepoDir: "/tmp/bt-fixture-repo"}, "/tmp/bt-fixture-repo")
	m.settings.Repos = []string{"/tmp/repo-a", "/tmp/repo-b", "/tmp/repo-c"} // >=2, would trigger Lobby at the cmd/tui.go layer if cwd had failed to resolve
	if m.view != viewBrowseRepo {
		t.Fatalf("newModel with a non-nil client started in view %v, want viewBrowseRepo -- newModel must never auto-derive the Lobby from Settings.Repos", m.view)
	}
	cmd := m.Init()
	if cmd == nil {
		t.Fatal("Init() returned a nil cmd")
	}
	msg := cmd()
	if _, ok := msg.(beansLoadedMsg); !ok {
		t.Fatalf("Init() with a non-nil client + view=viewBrowseRepo returned %T, want beansLoadedMsg (not the Lobby's repoMetricsMsg path)", msg)
	}
}
