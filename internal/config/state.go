package config

// state.go — beans-tui's lightweight persisted runtime state. Port devd
// internal/config/state.go VERBATIM (epic-E5-plan.md »Task 5« Step 4) with
// ONE rename: LastProject -> LastRepo. Lives in the SAME
// ~/.config/beans-tui/ directory as config.yaml (design decision c,
// deviation from devd's split ~/.config/devd-cli/ vs. ~/.config/dd/).
//
// ERRATUM (documented, not a bug): the bean's own acceptance-checklist
// sketch abbreviates the struct to "State{LastRepo string}" -- but the SAME
// checklist also names TestSetLastRepoPreservesOtherFields, and the plan's
// own Step 4 says "Port devd VERBATIM" (devd's State carries BOTH
// LastProject AND LastSeenVersion). Read together, "VERBATIM" wins: the
// struct SHAPE is ported whole (LastSeenVersion included), not diet-ed down
// to the bean prose's one-field sketch -- otherwise the named regression
// test would have nothing of substance to guard. LastSeenVersion is carried
// as a reserved parity field: no beans-tui feature reads or writes it yet
// (devd's own DD2-273 release-notes-on-upgrade detector has no bt-side
// counterpart today), but SetLastRepo's read-modify-write needs SOME other
// field to demonstrably preserve, exactly like devd's SetLastSeenVersion
// preserves LastProject.

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// State is the persisted CLI/TUI runtime state (JSON, ~/.config/beans-tui/
// state.json).
type State struct {
	LastRepo string `json:"last_repo"`
	// LastSeenVersion mirrors devd DD2-273 -- reserved, unused today.
	LastSeenVersion string `json:"last_seen_version,omitempty"`
}

func statePath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "state.json"), nil
}

// Load reads the State. A missing file returns a zero State without error
// (Port devd Load's own contract).
func Load() (State, error) {
	path, err := statePath()
	if err != nil {
		return State{}, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return State{}, nil
	}
	if err != nil {
		return State{}, err
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return State{}, err
	}
	return s, nil
}

// Save writes the State (creates ~/.config/beans-tui if missing).
func Save(s State) error {
	path, err := statePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// SetLastRepo persists LastRepo as a read-modify-write (Port devd
// SetLastSeenVersion's own RMW pattern): other State fields (LastSeenVersion
// today, any future field tomorrow) survive instead of being reset to their
// zero value by a naive Save(State{LastRepo: repo}).
func SetLastRepo(repo string) error {
	s, _ := Load() // missing/corrupt file -> zero State, s.LastSeenVersion stays ""
	s.LastRepo = repo
	return Save(s)
}
