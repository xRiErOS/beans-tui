package config

// state_test.go — TDD coverage for State (E5 Task 5, bean bt-0l8c). See
// state.go's ERRATUM doc-stamp for why LastSeenVersion exists alongside
// LastRepo (needed for TestSetLastRepoPreservesOtherFields to guard
// anything real).

import "testing"

func TestStateLoadMissingFileReturnsZero(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	s, err := Load()
	if err != nil {
		t.Fatalf("Load() on a missing file: %v", err)
	}
	if s != (State{}) {
		t.Errorf("Load() = %+v, want zero State", s)
	}
}

func TestStateSaveAndLoadRoundtrip(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	want := State{LastRepo: "/repo/a", LastSeenVersion: "v1.2.3"}
	if err := Save(want); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got != want {
		t.Errorf("Load() = %+v, want %+v", got, want)
	}
}

func TestSetLastRepoPreservesOtherFields(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	if err := Save(State{LastRepo: "/repo/old", LastSeenVersion: "v1.0.0"}); err != nil {
		t.Fatalf("seed Save: %v", err)
	}
	if err := SetLastRepo("/repo/new"); err != nil {
		t.Fatalf("SetLastRepo: %v", err)
	}
	got, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.LastRepo != "/repo/new" {
		t.Errorf("LastRepo = %q, want /repo/new", got.LastRepo)
	}
	if got.LastSeenVersion != "v1.0.0" {
		t.Errorf("LastSeenVersion = %q, want v1.0.0 (must survive SetLastRepo's read-modify-write)", got.LastSeenVersion)
	}
}
