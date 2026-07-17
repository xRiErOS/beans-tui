package cmd

import (
	"github.com/xRiErOS/beans-tui/internal/config"
	"github.com/xRiErOS/beans-tui/internal/data"
	"github.com/xRiErOS/beans-tui/internal/tui"

	"github.com/spf13/cobra"
)

// newTUICmd baut den expliziten `tui`-Subcommand (gleiches Verhalten wie der
// Root-Command ohne Subcommand).
func newTUICmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tui [repo-path]",
		Short: "Starts the TUI explicitly",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := ""
			if len(args) > 0 {
				path = args[0]
			}
			return runTUI(path)
		},
	}
}

// startupDecision is design decision d's own trigger logic (E5 Task 6, bean
// bt-zhwl, "Startup-Trigger"), factored out as a PURE function (no I/O)
// specifically so it is unit-testable without ever invoking tui.Run --
// tui.Run launches an interactive AltScreen tea.Program, which a `command go
// test` process cannot safely drive (no TTY). decideStartup below is the
// ENTIRE decision; runTUI just executes whichever branch it names.
type startupDecision int

const (
	startupUseArgRepo startupDecision = iota // an explicit path arg resolved -- use it
	startupUseCwdRepo                        // no arg, but data.FindRepo(cwd) resolved -- use it (US-01 unchanged)
	startupLobby                             // no arg, cwd unresolved, >=2 configured repos -- Lobby
	startupError                             // no arg, cwd unresolved, 0/1 configured repos -- pre-Lobby error, unchanged
)

// decideStartup implements design decision d's exact priority order
// (matching the bean's own wording verbatim): an explicit path arg ALWAYS
// wins over everything (a PO who named a repo gets that repo, full stop --
// Lobby never second-guesses an explicit choice); absent that, cwd
// resolving wins next (US-01: standing inside/under a repo already answers
// "which repo", regardless of how many entries Settings.Repos lists -- E1-
// E4 behavior stays byte-for-byte unchanged for this, the single most
// common invocation); only once BOTH are unresolved does Settings.Repos'
// size decide: >=2 -> Lobby (an actual choice to make), 0/1 -> the
// pre-Lobby error behavior, UNCHANGED (a Lobby over a single repo, or zero,
// has no value-add over the plain FindRepo error a PO already understands).
func decideStartup(argPath string, cwdFindErr error, configuredRepoCount int) startupDecision {
	if argPath != "" {
		return startupUseArgRepo
	}
	if cwdFindErr == nil {
		return startupUseCwdRepo
	}
	if configuredRepoCount >= 2 {
		return startupLobby
	}
	return startupError
}

// runTUI startet die eigentliche TUI: FindRepo (Discovery aufwärts vom
// gegebenen Pfad, Default cwd) -> Client -> tui.Run (AltScreen+Mouse,
// implementation-plan.md »E1 Task 8«) -- E5 Task 6 (bean bt-zhwl) inserts
// design decision d's Lobby-first branch (decideStartup above) ahead of the
// pre-Lobby error return.
func runTUI(path string) error {
	settings, err := config.LoadSettings()
	if err != nil {
		settings = config.DefaultSettings()
	}

	// An explicit path arg is resolved (and, per decideStartup, ALWAYS
	// wins) without ever needing data.FindRepo(cwd) at all -- skip that
	// walk entirely in this branch rather than computing it unconditionally
	// just to feed a decision that never looks at it. path != "" here is
	// EXACTLY decideStartup's own first condition (its own doc comment) --
	// duplicated as a plain bool check rather than a dummy decideStartup("",
	// ...) call, so this branch reads as what it is (an arg-given check),
	// not a partially-applied decision call.
	if path != "" {
		repoDir, ferr := data.FindRepo(path)
		if ferr != nil {
			return ferr
		}
		return tui.Run(&data.Client{RepoDir: repoDir}, repoDir, settings)
	}

	cwdRepoDir, cwdErr := data.FindRepo(".")
	switch decideStartup("", cwdErr, len(settings.Repos)) {
	case startupUseCwdRepo:
		return tui.Run(&data.Client{RepoDir: cwdRepoDir}, cwdRepoDir, settings)
	case startupLobby:
		return tui.Run(nil, "", settings) // design decision d: Lobby-first (client == nil, tui.Run's own contract)
	default: // startupError
		return cwdErr
	}
}
