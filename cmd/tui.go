package cmd

import (
	"beans-tui/internal/data"
	"beans-tui/internal/tui"

	"github.com/spf13/cobra"
)

// newTUICmd baut den expliziten `tui`-Subcommand (gleiches Verhalten wie der
// Root-Command ohne Subcommand).
func newTUICmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tui [repo-pfad]",
		Short: "Startet die TUI explizit",
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

// runTUI startet die eigentliche TUI: FindRepo (Discovery aufwärts vom
// gegebenen Pfad, Default cwd) -> Client -> tui.Run (AltScreen+Mouse,
// implementation-plan.md »E1 Task 8«).
func runTUI(path string) error {
	if path == "" {
		path = "."
	}
	repoDir, err := data.FindRepo(path)
	if err != nil {
		return err
	}
	client := &data.Client{RepoDir: repoDir}
	return tui.Run(client, repoDir)
}
