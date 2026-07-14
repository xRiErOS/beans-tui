package data

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverFindsConfigUpward(t *testing.T) {
	repo := newTestRepo(t)

	sub := filepath.Join(repo, "a", "b", "c")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatalf("mkdir sub: %v", err)
	}

	found, err := FindRepo(sub)
	if err != nil {
		t.Fatalf("FindRepo() error = %v", err)
	}
	if found != repo {
		t.Errorf("FindRepo() = %q, want %q", found, repo)
	}
}

func TestDiscoverErrorsWhenNoConfigFound(t *testing.T) {
	tmp := t.TempDir()

	if _, err := FindRepo(tmp); err == nil {
		t.Fatal("FindRepo() error = nil, want error")
	}
}
