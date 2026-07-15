package tui

// view_browse_repo_test.go — TDD coverage for viewBrowseRepo's own Chrome
// (browseRepoChrome, PF-11 Header/Footer-Keybinding-Split, design-spec.md
// §15, epic-E7-plan.md Task 7, bean bt-m6at): the Header now carries ALL 7
// global bindings (Nachtrag 9), the Footer drops Refresh/Enter (now
// header-only) and gains FocusIn/FocusOut as real bindings (replacing the
// former hand-typed "  tab:focus" suffix, PF-13), and swaps to a context's
// own bindings while a Filter-Menu/Overlay/Search/Palette/Help is open
// (Q04-Antwort).

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestBrowseRepoChromeHeaderShowsAllSevenGlobals(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	head, _ := m.browseRepoChrome(200) // wide enough to never trigger breadcrumb's narrow-stack fallback
	plain := ansi.Strip(head)
	for _, want := range []string{"ctrl+r:reload", "ctrl+k:commands", "p:repos", "?:help", "esc:back", "enter:open/confirm", "q:quit"} {
		if !strings.Contains(plain, want) {
			t.Errorf("browseRepoChrome header = %q, want it to contain %q", plain, want)
		}
	}
}

func TestBrowseRepoChromeFooterOmitsRefreshAndEnter(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	_, localKeys := m.browseRepoChrome(200)
	plain := ansi.Strip(localKeys)
	if strings.Contains(plain, "ctrl+r:") {
		t.Errorf("browseRepoChrome footer = %q, must NOT contain ctrl+r (now header-only, PF-11)", plain)
	}
	if strings.Contains(plain, "enter:") {
		t.Errorf("browseRepoChrome footer = %q, must NOT contain enter: (now header-only, PF-11)", plain)
	}
}

func TestBrowseRepoChromeFooterShowsFocusInFocusOut(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	_, localKeys := m.browseRepoChrome(200)
	plain := ansi.Strip(localKeys)
	if !strings.Contains(plain, "tab:focus in/toggle") {
		t.Errorf("browseRepoChrome footer = %q, want FocusIn's real hint (not a hand-typed suffix)", plain)
	}
	if !strings.Contains(plain, "shift+tab:focus out") {
		t.Errorf("browseRepoChrome footer = %q, want shift+tab visible for the first time (PF-13)", plain)
	}
}

func TestBrowseRepoChromeFooterIsContextSensitiveOnFilterOpen(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m.filterOpen = true
	_, localKeys := m.browseRepoChrome(200)
	plain := ansi.Strip(localKeys)
	if !strings.Contains(plain, "space/x:Toggle facet") {
		t.Errorf("browseRepoChrome footer while filterOpen = %q, want the Toggle hint (Q04)", plain)
	}
	if strings.Contains(plain, "c:Create") {
		t.Errorf("browseRepoChrome footer while filterOpen = %q, must not leak the (now irrelevant) view-local Create hint", plain)
	}
}
