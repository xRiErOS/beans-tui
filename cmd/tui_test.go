package cmd

// tui_test.go — E5 Task 6 (bean bt-zhwl, design decision d, "Startup-
// Trigger"): decideStartup is the WHOLE decision, factored out pure
// specifically so it's testable without ever invoking tui.Run (an
// interactive AltScreen tea.Program that a `command go test` process cannot
// safely drive -- no TTY).

import (
	"errors"
	"testing"
)

func TestDecideStartupPrioritiesInOrder(t *testing.T) {
	cwdOK := error(nil)
	cwdFail := errors.New("not a beans repo")

	cases := []struct {
		name    string
		argPath string
		cwdErr  error
		repos   int
		want    startupDecision
	}{
		{"explicit path always wins, even with cwd ok and >=2 repos", "/some/repo", cwdOK, 5, startupUseArgRepo},
		{"explicit path wins even with cwd failing and 0 repos", "/some/repo", cwdFail, 0, startupUseArgRepo},
		{"no arg, cwd resolves -- US-01 unchanged, ignores repo count", "", cwdOK, 5, startupUseCwdRepo},
		{"no arg, cwd resolves, 0 repos -- still cwd, no regression", "", cwdOK, 0, startupUseCwdRepo},
		{"no arg, cwd fails, >=2 repos -- Lobby", "", cwdFail, 2, startupLobby},
		{"no arg, cwd fails, many repos -- Lobby", "", cwdFail, 5, startupLobby},
		{"no arg, cwd fails, exactly 1 repo -- pre-Lobby error, unchanged (US-01 core case)", "", cwdFail, 1, startupError},
		{"no arg, cwd fails, 0 repos -- pre-Lobby error, unchanged", "", cwdFail, 0, startupError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := decideStartup(tc.argPath, tc.cwdErr, tc.repos)
			if got != tc.want {
				t.Fatalf("decideStartup(%q, err=%v, repos=%d) = %v, want %v", tc.argPath, tc.cwdErr, tc.repos, got, tc.want)
			}
		})
	}
}
