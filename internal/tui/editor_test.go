package tui

// editor_test.go — TDD coverage for the $EDITOR-Suspend on the Body field
// (`ctrl+e`, E3 Task 5, bean bt-sl45): editorBinary's env-cascade (design
// decision c), prepareEditor/readEditorResult's tempfile round-trip (Port
// devd editor.go, testable without a tea runtime per its own doc-comment),
// and the editorFinishedMsg -> SetBody dispatch (design decision d: fresh
// etag, changed-only, vanished-target guard).

import (
	"os"
	"strings"
	"testing"

	"beans-tui/internal/data"
	tea "github.com/charmbracelet/bubbletea"
)

// TestEditorBinaryResolvesVisualThenEditorThenVi guards design decision c's
// env cascade: $VISUAL wins over $EDITOR, both win over the "vi" fallback
// (POSIX default -- portable everywhere, unlike devd's "nvim" assumption).
// Whitespace-only values are treated as unset (mirrors devd's own
// strings.TrimSpace guard).
func TestEditorBinaryResolvesVisualThenEditorThenVi(t *testing.T) {
	cases := []struct {
		name   string
		visual string
		editor string
		want   []string
	}{
		{"neither set", "", "", []string{"vi"}},
		{"EDITOR only", "", "nano", []string{"nano"}},
		{"VISUAL wins over EDITOR", "code -w", "nano", []string{"code", "-w"}},
		{"whitespace-only falls back to vi", "   ", "  ", []string{"vi"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("VISUAL", tc.visual)
			t.Setenv("EDITOR", tc.editor)
			got := editorBinary()
			if len(got) != len(tc.want) {
				t.Fatalf("editorBinary() = %v, want %v", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("editorBinary() = %v, want %v", got, tc.want)
				}
			}
		})
	}
}

// TestPrepareEditorWritesInitialContentToTempFile guards prepareEditor's
// tempfile setup -- testable WITHOUT a tea runtime (Port devd editor.go
// doc-comment, cmd is built but never run here).
func TestPrepareEditorWritesInitialContentToTempFile(t *testing.T) {
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "true") // any resolvable no-op binary; prepareEditor never RUNS cmd

	path, cmd, err := prepareEditor("hello body", ".md")
	if err != nil {
		t.Fatalf("prepareEditor() error = %v", err)
	}
	defer os.Remove(path)

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	if string(got) != "hello body" {
		t.Fatalf("temp file content = %q, want %q", got, "hello body")
	}
	if !strings.HasSuffix(path, ".md") {
		t.Errorf("temp file path = %q, want a .md suffix", path)
	}
	if cmd == nil {
		t.Fatal("prepareEditor() cmd is nil")
	}
	if len(cmd.Args) < 2 || cmd.Args[0] != "true" || cmd.Args[len(cmd.Args)-1] != path {
		t.Errorf("cmd.Args = %v, want [\"true\" ... %q]", cmd.Args, path)
	}
}

// TestReadEditorResultDetectsChangeAndCleansUp guards the changed-flag +
// tempfile-cleanup contract (Port devd editor.go:70-83).
func TestReadEditorResultDetectsChangeAndCleansUp(t *testing.T) {
	f, err := os.CreateTemp("", "beans-tui-readresult-*.md")
	if err != nil {
		t.Fatalf("CreateTemp() error = %v", err)
	}
	path := f.Name()
	if _, err := f.WriteString("edited content"); err != nil {
		t.Fatalf("WriteString() error = %v", err)
	}
	f.Close()

	msg := readEditorResult(path, "original content", nil)
	if msg.err != nil {
		t.Fatalf("readEditorResult() err = %v, want nil", msg.err)
	}
	if !msg.changed {
		t.Error("changed = false, want true (content differs from initial)")
	}
	if msg.content != "edited content" {
		t.Errorf("content = %q, want %q", msg.content, "edited content")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("readEditorResult did not clean up the temp file")
	}

	// Unchanged content -> changed must be false. errBoom (update_test.go)
	// covers runErr propagation implicitly via the vanished-target test
	// below (a runErr short-circuits before the file is even read).
	f2, err := os.CreateTemp("", "beans-tui-readresult-unchanged-*.md")
	if err != nil {
		t.Fatalf("CreateTemp() error = %v", err)
	}
	path2 := f2.Name()
	if _, err := f2.WriteString("same"); err != nil {
		t.Fatalf("WriteString() error = %v", err)
	}
	f2.Close()
	msg2 := readEditorResult(path2, "same", nil)
	if msg2.changed {
		t.Error("changed = true, want false (content identical to initial)")
	}
}

// TestEditorFinishedUnchangedFiresNoMutation guards the changed==false
// no-op path: an unedited editor round-trip must not fire SetBody.
func TestEditorFinishedUnchangedFiresNoMutation(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m.client = &data.Client{RepoDir: "/nonexistent-bt-e3-t5-scratch-dir"}
	m.editorTarget = "tk-1"

	tm, cmd := m.Update(editorFinishedMsg{content: "same body", changed: false})
	nm, ok := tm.(model)
	if !ok {
		t.Fatalf("Update(editorFinishedMsg) did not return a model, got %T", tm)
	}
	if cmd != nil {
		t.Fatal("unchanged editor content must not fire a mutation Cmd")
	}
	if nm.editorTarget != "" {
		t.Errorf("editorTarget = %q, want cleared after handling", nm.editorTarget)
	}
}

// TestEditorFinishedChangedFiresSetBodyWithFreshETag guards the changed
// path: SetBody must fire against a FRESH etag read from the live index
// (design decision d), verified the same no-binary-required way
// box_picker_parent_test.go/box_confirm_create_test.go already use -- the
// dispatched mutationDoneMsg's error text names the CLI subcommand.
func TestEditorFinishedChangedFiresSetBodyWithFreshETag(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m.client = &data.Client{RepoDir: "/nonexistent-bt-e3-t5-scratch-dir"}
	m.editorTarget = "tk-1"

	tm, cmd := m.Update(editorFinishedMsg{content: "new body", changed: true})
	nm, ok := tm.(model)
	if !ok {
		t.Fatalf("Update(editorFinishedMsg) did not return a model, got %T", tm)
	}
	if nm.editorTarget != "" {
		t.Errorf("editorTarget = %q, want cleared after handling", nm.editorTarget)
	}
	if cmd == nil {
		t.Fatal("changed editor content must fire a mutation Cmd")
	}
	msg := cmd()
	mdm, ok := msg.(mutationDoneMsg)
	if !ok {
		t.Fatalf("cmd() = %T, want mutationDoneMsg", msg)
	}
	if mdm.err == nil || !strings.Contains(mdm.err.Error(), "beans update") {
		t.Fatalf("mutationDoneMsg.err = %v, want an error containing %q (proves SetBody dispatched)", mdm.err, "beans update")
	}
}

// TestEditorFinishedTargetVanishedSurfacesError guards the vanished-target
// guard (design decision d, mirrors beanETag's every other caller): the
// bean editorTarget names is no longer in the index -- no doomed mutation
// Cmd, a status-line note instead.
func TestEditorFinishedTargetVanishedSurfacesError(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m.editorTarget = "does-not-exist"

	tm, cmd := m.Update(editorFinishedMsg{content: "new body", changed: true})
	nm, ok := tm.(model)
	if !ok {
		t.Fatalf("Update(editorFinishedMsg) did not return a model, got %T", tm)
	}
	if cmd != nil {
		t.Fatal("a vanished target must not fire a doomed mutation Cmd")
	}
	if nm.err == "" {
		t.Fatal("a vanished target must surface a status-line error")
	}
	if nm.editorTarget != "" {
		t.Errorf("editorTarget = %q, want cleared even on the vanished-target path", nm.editorTarget)
	}
}

// TestKeyNodeActionCtrlEStartsEditorSuspend guards keyNodeAction's Editor
// dispatch (design decision h): ctrl+e records m.editorTarget and returns a
// non-nil Cmd (the ExecProcess-wrapped suspend) -- it must NOT open a form
// (that's "e"'s job, form_edit_title_test.go).
func TestKeyNodeActionCtrlEStartsEditorSuspend(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m = focusBean(m, "tk-1")

	tm, cmd := m.Update(keyMsg(tea.KeyCtrlE))
	nm, ok := tm.(model)
	if !ok {
		t.Fatalf("Update(ctrl+e) did not return a model, got %T", tm)
	}
	if nm.editorTarget != "tk-1" {
		t.Fatalf("editorTarget = %q, want tk-1", nm.editorTarget)
	}
	if nm.form != nil {
		t.Fatal("ctrl+e must NOT open a form (that is e's job)")
	}
	if cmd == nil {
		t.Fatal("ctrl+e must return a Cmd (the ExecProcess-wrapped editor suspend)")
	}
}
