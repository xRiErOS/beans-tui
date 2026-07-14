package cmd

import "testing"

func TestRootStartsTUIByDefault(t *testing.T) {
	cmd := NewRootCmd()
	if cmd.Use != "bt [repo-pfad]" {
		t.Fatalf("Use = %q", cmd.Use)
	}
	// tui-Subcommand vorhanden
	sub, _, err := cmd.Find([]string{"tui"})
	if err != nil || sub.Name() != "tui" {
		t.Fatalf("tui subcommand fehlt: %v", err)
	}
}
