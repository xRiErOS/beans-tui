package cmd

import (
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

// runTUI startet die eigentliche TUI. Stub — die Implementierung folgt in
// späteren Tasks (Datenlayer, Theme, App-Shell).
func runTUI(path string) error {
	return nil
}
