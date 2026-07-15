package tui

// editor_test.go — TDD coverage for the $EDITOR-Suspend on the Body field
// (`ctrl+e`, E3 Task 5, bean bt-sl45): editorBinary's env-cascade (design
// decision c), prepareEditor/readEditorResult's tempfile round-trip (Port
// devd editor.go, testable without a tea runtime per its own doc-comment),
// and the editorFinishedMsg -> SetBody dispatch (changed-only, vanished-
// target guard, F2 Review-Runde 2: etag CAPTURED AT OPEN, never a fresh
// m.beanETag(id) re-read at submit -- see applyEditorFinished's doc-stamp,
// update.go, for the full lost-update rationale).

import (
	"errors"
	"os"
	"strings"
	"testing"

	"beans-tui/internal/data"
	tea "github.com/charmbracelet/bubbletea"
)

// fakeBeansOnPath installs an executable "beans" shell script (script,
// verbatim) on PATH ahead of the real binary, for the duration of the
// calling test (t.Setenv reverts automatically). Used by the F2 tests below
// to observe values (e.g. the --if-match etag) that cross into a real
// subprocess -- something no amount of in-process struct inspection can see
// once data.Client has handed them to exec.Command.
func fakeBeansOnPath(t *testing.T, script string) {
	t.Helper()
	dir := t.TempDir()
	path := dir + "/beans"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake beans script: %v", err)
	}
	t.Setenv("PATH", dir+":"+os.Getenv("PATH"))
}

// fakeBeansEchoIfMatch installs a fake "beans" that always fails, echoing
// the --if-match value it received to stderr in a form classifyError's
// conflictSubstring fallback (mutations.go) recognizes as a conflict --
// letting a test assert on the ACTUAL etag string SetBody dispatched, not
// just that a mutation fired at all.
func fakeBeansEchoIfMatch(t *testing.T) {
	t.Helper()
	fakeBeansOnPath(t, `#!/bin/sh
prev=""
etag=""
for arg in "$@"; do
  if [ "$prev" = "--if-match" ]; then
    etag="$arg"
  fi
  prev="$arg"
done
echo "etag mismatch: received-if-match=$etag" 1>&2
exit 1
`)
}

// fakeBeansConflict installs a fake "beans" that always fails with a real
// CONFLICT JSON error envelope on stdout (the shape classifyError's PRIMARY
// branch parses, mutations.go) -- for tests that only need a genuine
// data.ErrConflict-classified error, not the specific etag value.
func fakeBeansConflict(t *testing.T) {
	t.Helper()
	fakeBeansOnPath(t, `#!/bin/sh
echo '{"success":false,"error":"etag mismatch: stale","code":"CONFLICT"}'
exit 1
`)
}

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

// TestEditorBinaryPrefersConfiguredEditorOverEnv guards design decision c's
// TOP layer (E5 Task 5, bean bt-0l8c): a non-empty configuredEditor
// (Settings.Editor, box_form_settings.go/app.go's config.LoadSettings)
// STRICTLY wins over $VISUAL/$EDITOR, even when both env vars are set --
// Settings > $VISUAL > $EDITOR > vi (the bean's own "design decision c,
// PFLICHT" wording). configuredEditor is reset via t.Cleanup so this test
// can never leak into TestEditorBinaryResolvesVisualThenEditorThenVi (or any
// other test in this package) regardless of run order -- that E3 test stays
// UNCHANGED and green with configuredEditor at its zero-value default "".
func TestEditorBinaryPrefersConfiguredEditorOverEnv(t *testing.T) {
	orig := configuredEditor
	t.Cleanup(func() { configuredEditor = orig })

	configuredEditor = "code -w"
	t.Setenv("VISUAL", "vim")
	t.Setenv("EDITOR", "nano")

	got := editorBinary()
	want := []string{"code", "-w"}
	if len(got) != len(want) {
		t.Fatalf("editorBinary() = %v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("editorBinary() = %v, want %v", got, want)
		}
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
	m.editorETag = "captured-etag"

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
	if nm.editorETag != "" {
		t.Errorf("editorETag = %q, want cleared after handling (F2)", nm.editorETag)
	}
}

// TestEditorFinishedChangedFiresSetBodyWithCapturedETag guards the changed
// path: SetBody must fire against the etag CAPTURED AT $EDITOR-OPEN TIME
// (m.editorETag, F2 Review-Runde 2), verified the same no-binary-required
// way box_picker_parent_test.go/box_confirm_create_test.go already use --
// the dispatched mutationDoneMsg's error text names the CLI subcommand. The
// actual "captured, not fresh" claim is proven precisely by
// TestEditorFinishedUsesEtagCapturedAtOpenNotFreshIndexRead below; this test
// only guards the dispatch + field-clearing contract.
func TestEditorFinishedChangedFiresSetBodyWithCapturedETag(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m.client = &data.Client{RepoDir: "/nonexistent-bt-e3-t5-scratch-dir"}
	m.editorTarget = "tk-1"
	m.editorETag = "captured-etag"

	tm, cmd := m.Update(editorFinishedMsg{content: "new body", changed: true})
	nm, ok := tm.(model)
	if !ok {
		t.Fatalf("Update(editorFinishedMsg) did not return a model, got %T", tm)
	}
	if nm.editorTarget != "" {
		t.Errorf("editorTarget = %q, want cleared after handling", nm.editorTarget)
	}
	if nm.editorETag != "" {
		t.Errorf("editorETag = %q, want cleared after handling (F2)", nm.editorETag)
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

// TestEditorFinishedUsesEtagCapturedAtOpenNotFreshIndexRead is F2's core
// regression test (Review-Runde 2, Finding 2, ETag-Lost-Update): a
// watch-reload landing WHILE the $EDITOR session is still open rotates
// tk-1's etag in the LIVE index (m.idx) -- design decision d's "always read
// fresh at submit" would silently pick up that NEW etag here and let
// SetBody sail through with no conflict ever raised, discarding whatever
// changed on disk during the session. Freezing the etag at open
// (keyNodeAction's ctrl+e branch, update.go) must mean applyEditorFinished
// still dispatches the OLD (open-time) etag even after the reload. Proven
// via a fake "beans" binary (fakeBeansEchoIfMatch) that echoes the actual
// --if-match value it received back into the error text -- something no
// amount of in-process struct inspection can observe once the value has
// crossed into a real subprocess.
func TestEditorFinishedUsesEtagCapturedAtOpenNotFreshIndexRead(t *testing.T) {
	fakeBeansEchoIfMatch(t)

	openBeans := fixtureBeans()
	for i := range openBeans {
		if openBeans[i].ID == "tk-1" {
			openBeans[i].ETag = "etag-open"
		}
	}
	m := fixtureModel(t, openBeans)
	m.client = &data.Client{RepoDir: t.TempDir()}
	m = focusBean(m, "tk-1")

	// ctrl+e captures editorETag from the CURRENT (open-time) index.
	tm, _ := m.Update(keyMsg(tea.KeyCtrlE))
	nm, ok := tm.(model)
	if !ok {
		t.Fatalf("Update(ctrl+e) did not return a model, got %T", tm)
	}
	if nm.editorETag != "etag-open" {
		t.Fatalf("editorETag = %q, want %q (captured at open)", nm.editorETag, "etag-open")
	}

	// A watch-reload lands WHILE the $EDITOR session is still open, rotating
	// tk-1's etag in the live index -- the exact async window F2 guards.
	reloadBeans := fixtureBeans()
	for i := range reloadBeans {
		if reloadBeans[i].ID == "tk-1" {
			reloadBeans[i].ETag = "etag-after-reload"
		}
	}
	nm = step(t, nm, beansLoadedMsg{beans: reloadBeans})
	if nm.editorETag != "etag-open" {
		t.Fatalf("editorETag = %q, want unchanged by an unrelated reload", nm.editorETag)
	}

	tm2, cmd := nm.Update(editorFinishedMsg{content: "new body", changed: true})
	if _, ok := tm2.(model); !ok {
		t.Fatalf("Update(editorFinishedMsg) did not return a model, got %T", tm2)
	}
	if cmd == nil {
		t.Fatal("changed editor content must fire a mutation Cmd")
	}
	msg := cmd()
	mdm, ok := msg.(mutationDoneMsg)
	if !ok {
		t.Fatalf("cmd() = %T, want mutationDoneMsg", msg)
	}
	if mdm.err == nil {
		t.Fatal("setup: the fake beans script always exits 1")
	}
	if !strings.Contains(mdm.err.Error(), "received-if-match=etag-open") {
		t.Fatalf("SetBody dispatched --if-match: %v, want it to contain the OPEN-time etag %q (F2: must NOT re-read m.idx fresh at submit)", mdm.err, "etag-open")
	}
	if strings.Contains(mdm.err.Error(), "etag-after-reload") {
		t.Fatalf("SetBody dispatched the RELOADED etag: %v, want the one captured at open (F2 regression)", mdm.err)
	}
}

// TestEditorFinishedConflictWritesRecoveryTempFileAndSurfacesPath guards
// F2's nice-to-have (Review-Runde 2): on a genuine ErrConflict, the PO's
// just-$EDITOR-edited content would otherwise be silently discarded by
// applyMutationResult's unconditional reload -- it must instead be persisted
// to a KEPT tempfile first, with the file's path riding along in the
// status-line text applyMutationResult (update.go) produces.
func TestEditorFinishedConflictWritesRecoveryTempFileAndSurfacesPath(t *testing.T) {
	fakeBeansConflict(t)

	m := fixtureModel(t, fixtureBeans())
	m.client = &data.Client{RepoDir: t.TempDir()}
	m.editorTarget = "tk-1"
	m.editorETag = "etag-open"

	const editedBody = "edited body that must not be lost"
	tm, cmd := m.Update(editorFinishedMsg{content: editedBody, changed: true})
	nm, ok := tm.(model)
	if !ok {
		t.Fatalf("Update(editorFinishedMsg) did not return a model, got %T", tm)
	}
	if cmd == nil {
		t.Fatal("changed editor content must fire a mutation Cmd")
	}
	msg := cmd()
	mdm, ok := msg.(mutationDoneMsg)
	if !ok {
		t.Fatalf("cmd() = %T, want mutationDoneMsg", msg)
	}
	if !errors.Is(mdm.err, data.ErrConflict) {
		t.Fatalf("mutationDoneMsg.err = %v, want errors.Is(_, data.ErrConflict)", mdm.err)
	}

	fm := step(t, nm, mdm)
	if !strings.Contains(fm.err, "Conflict") {
		t.Fatalf("status line = %q, want it to mention the conflict", fm.err)
	}
	const marker = "your version: "
	idx := strings.Index(fm.err, marker)
	if idx < 0 {
		t.Fatalf("status line = %q, want it to carry the recovery tempfile's path (%q)", fm.err, marker)
	}
	path := strings.TrimSpace(fm.err[idx+len(marker):])
	content, rerr := os.ReadFile(path)
	if rerr != nil {
		t.Fatalf("ReadFile(%q) error = %v (recovery tempfile missing)", path, rerr)
	}
	if string(content) != editedBody {
		t.Fatalf("recovery tempfile content = %q, want %q", content, editedBody)
	}
}

// TestEditorFinishedTargetVanishedSurfacesError guards the vanished-target
// guard (mirrors beanETag's every other caller): the bean editorTarget
// names is no longer in the index -- no doomed mutation Cmd, a status-line
// note instead. m.beanETag(id) is still consulted for this presence check
// (F2: only its returned etag VALUE is now ignored in favor of
// m.editorETag, not the ok bool itself, applyEditorFinished's doc-stamp).
//
// E5 Task 1 (bean bt-6dts) update: this branch now ALSO fires a (non-sticky)
// warn Toast alongside m.err (design decision a, Dual-Write) -- so cmd is no
// longer nil, it carries the Toast's auto-dismiss tea.Tick. That Cmd is
// NEVER invoked here (it would really sleep toastDuration(toastWarn), 3s --
// same rationale as devd's own overlay_show_toast_test.go doc comment): the
// "no doomed mutation" guarantee is instead proven structurally --
// applyEditorFinished's vanished-target branch returns before ever reaching
// mutateCmd/client.SetBody, and m.toast's own kind/title pin down that the
// non-nil Cmd is the Toast timer, not a mutation dispatch.
func TestEditorFinishedTargetVanishedSurfacesError(t *testing.T) {
	m := fixtureModel(t, fixtureBeans())
	m.editorTarget = "does-not-exist"
	m.editorETag = "captured-etag"

	tm, cmd := m.Update(editorFinishedMsg{content: "new body", changed: true})
	nm, ok := tm.(model)
	if !ok {
		t.Fatalf("Update(editorFinishedMsg) did not return a model, got %T", tm)
	}
	if cmd == nil {
		t.Fatal("a vanished target must still fire the warn-Toast's auto-dismiss Cmd (E5 Task 1)")
	}
	if nm.err == "" {
		t.Fatal("a vanished target must surface a status-line error")
	}
	if nm.toast == nil || nm.toast.kind != toastWarn || nm.toast.title != nm.err {
		t.Fatalf("toast = %+v, want a non-nil toastWarn mirroring m.err %q (E5 Task 1 Dual-Write)", nm.toast, nm.err)
	}
	if nm.editorTarget != "" {
		t.Errorf("editorTarget = %q, want cleared even on the vanished-target path", nm.editorTarget)
	}
	if nm.editorETag != "" {
		t.Errorf("editorETag = %q, want cleared even on the vanished-target path (F2)", nm.editorETag)
	}
}

// TestKeyNodeActionCtrlEStartsEditorSuspend guards keyNodeAction's Editor
// dispatch (design decision h): ctrl+e records m.editorTarget AND
// m.editorETag (F2 Review-Runde 2: the etag captured HERE, at open time, is
// what applyEditorFinished must use later, not a fresh re-read) and returns
// a non-nil Cmd (the ExecProcess-wrapped suspend) -- it must NOT open a form
// (that's "e"'s job, form_edit_title_test.go).
func TestKeyNodeActionCtrlEStartsEditorSuspend(t *testing.T) {
	beans := fixtureBeans()
	for i := range beans {
		if beans[i].ID == "tk-1" {
			beans[i].ETag = "tk-1-etag"
		}
	}
	m := fixtureModel(t, beans)
	m = focusBean(m, "tk-1")

	tm, cmd := m.Update(keyMsg(tea.KeyCtrlE))
	nm, ok := tm.(model)
	if !ok {
		t.Fatalf("Update(ctrl+e) did not return a model, got %T", tm)
	}
	if nm.editorTarget != "tk-1" {
		t.Fatalf("editorTarget = %q, want tk-1", nm.editorTarget)
	}
	if nm.editorETag != "tk-1-etag" {
		t.Fatalf("editorETag = %q, want %q (captured at open, F2)", nm.editorETag, "tk-1-etag")
	}
	if nm.form != nil {
		t.Fatal("ctrl+e must NOT open a form (that is e's job)")
	}
	if cmd == nil {
		t.Fatal("ctrl+e must return a Cmd (the ExecProcess-wrapped editor suspend)")
	}
}
