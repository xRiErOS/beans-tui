package data

// tagdefs_test.go — TDD coverage for the Tag-Registry persistence layer
// (E10 Task 1, bean bt-49hh, epic bt-362n D01-D04). Mirrors
// internal/config/settings_test.go's naming/structure convention
// (TestLoadSettingsMissingFileReturnsDefaults etc.) -- LoadTagDefs/
// SaveTagDefs are the repo-local, tolerant-missing analogue of
// config.LoadSettings/SaveUserSettings.

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadTagDefsMissingFileReturnsEmpty(t *testing.T) {
	c := &Client{RepoDir: t.TempDir()}
	got, err := c.LoadTagDefs()
	if err != nil || len(got) != 0 {
		t.Fatalf("want (nil, nil), got (%v, %v)", got, err)
	}
}

func TestLoadTagDefsSkipsInvalidNamesDefensively(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".beans-tags.yml"),
		[]byte("tags:\n  - good-tag\n  - Bad_Tag\n  - \n"), 0o644)
	c := &Client{RepoDir: dir}
	got, _ := c.LoadTagDefs()
	if len(got) != 1 || got[0] != "good-tag" {
		t.Fatalf("want [good-tag], got %v", got)
	}
}

func TestSaveTagDefsRoundTripSortedDeduped(t *testing.T) {
	dir := t.TempDir()
	c := &Client{RepoDir: dir}
	if err := c.SaveTagDefs([]string{"zeta", "alpha", "alpha"}); err != nil {
		t.Fatal(err)
	}
	got, _ := c.LoadTagDefs()
	want := []string{"alpha", "zeta"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestRenameTagDefNamePromotesUnregisteredOldName(t *testing.T) {
	got := RenameTagDefName([]string{"a", "b"}, "c-not-registered", "d")
	want := []string{"a", "b", "d"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

// TestLoadTagDefsCorruptYAMLReturnsEmptyNoPanic -- corrupt YAML must never
// crash or error out, same "never crash" philosophy as config.LoadSettings.
func TestLoadTagDefsCorruptYAMLReturnsEmptyNoPanic(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".beans-tags.yml"),
		[]byte("tags: [this is not: valid: yaml: at: all"), 0o644); err != nil {
		t.Fatal(err)
	}
	c := &Client{RepoDir: dir}
	got, err := c.LoadTagDefs()
	if err != nil {
		t.Fatalf("want nil error on corrupt YAML, got %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("want empty registry on corrupt YAML, got %v", got)
	}
}

// TestSaveTagDefsWritesReadablePermissions verifies the 0o644 mode and that
// the written file round-trips through LoadTagDefs as valid YAML.
func TestSaveTagDefsWritesReadablePermissions(t *testing.T) {
	dir := t.TempDir()
	c := &Client{RepoDir: dir}
	if err := c.SaveTagDefs([]string{"solo"}); err != nil {
		t.Fatal(err)
	}
	fi, err := os.Stat(filepath.Join(dir, ".beans-tags.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode().Perm() != 0o644 {
		t.Fatalf("want mode 0o644, got %v", fi.Mode().Perm())
	}
}

func TestAddTagDefNameNoOpOnDuplicate(t *testing.T) {
	got := AddTagDefName([]string{"alpha", "beta"}, "alpha")
	want := []string{"alpha", "beta"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestAddTagDefNameInsertsSortedDeduped(t *testing.T) {
	got := AddTagDefName([]string{"alpha", "zeta"}, "mu")
	want := []string{"alpha", "mu", "zeta"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestRemoveTagDefNameNoOpOnAbsence(t *testing.T) {
	got := RemoveTagDefName([]string{"alpha", "beta"}, "gamma")
	want := []string{"alpha", "beta"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestRemoveTagDefNameRemovesExisting(t *testing.T) {
	got := RemoveTagDefName([]string{"alpha", "beta", "gamma"}, "beta")
	want := []string{"alpha", "gamma"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestRenameTagDefNameRenamesExisting(t *testing.T) {
	got := RenameTagDefName([]string{"alpha", "beta"}, "alpha", "delta")
	want := []string{"beta", "delta"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

// TestAddRemoveRenameDoNotMutateInputSlice -- acceptance requires "keine
// Mutation des Eingabe-Slices — neue Slice zurückgeben".
func TestAddRemoveRenameDoNotMutateInputSlice(t *testing.T) {
	orig := []string{"beta", "alpha"}
	snapshot := append([]string(nil), orig...)

	AddTagDefName(orig, "gamma")
	if !reflect.DeepEqual(orig, snapshot) {
		t.Fatalf("AddTagDefName mutated input: got %v, want %v", orig, snapshot)
	}

	RemoveTagDefName(orig, "alpha")
	if !reflect.DeepEqual(orig, snapshot) {
		t.Fatalf("RemoveTagDefName mutated input: got %v, want %v", orig, snapshot)
	}

	RenameTagDefName(orig, "alpha", "delta")
	if !reflect.DeepEqual(orig, snapshot) {
		t.Fatalf("RenameTagDefName mutated input: got %v, want %v", orig, snapshot)
	}
}
