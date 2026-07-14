// Package cmd implementiert die CLI-Oberfläche von bt (beans-tui): Root-Command
// und Subcommands via cobra.
package cmd

import (
	"github.com/spf13/cobra"
)

// NewRootCmd baut den Root-Command von bt. Ohne Subcommand startet er direkt
// die TUI (optionaler Pfad-Arg für den beans-Repo-Pfad, Default cwd).
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "bt [repo-pfad]",
		Short: "bt — PO-Cockpit-TUI für beans-Repos",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := ""
			if len(args) > 0 {
				path = args[0]
			}
			return runTUI(path)
		},
	}

	root.AddCommand(newTUICmd())

	return root
}

// Execute führt den Root-Command aus. Wird von main.go aufgerufen.
func Execute() error {
	return NewRootCmd().Execute()
}
