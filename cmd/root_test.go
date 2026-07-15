package cmd

import "testing"

func TestRootStartsTUIByDefault(t *testing.T) {
	cmd := NewRootCmd()
	if cmd.Use != "bt [repo-path]" {
		t.Fatalf("Use = %q", cmd.Use)
	}
	// tui subcommand present
	sub, _, err := cmd.Find([]string{"tui"})
	if err != nil || sub.Name() != "tui" {
		t.Fatalf("tui subcommand missing: %v", err)
	}
}
