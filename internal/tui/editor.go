package tui

// editor.go — $EDITOR-Suspend for the Body field (`ctrl+e`, E3 Task 5, bean
// bt-sl45, design decisions c/h): tea.ExecProcess suspends the running
// Bubble Tea program, hands the terminal to the user's editor on a temp file
// seeded with the bean's current body, and resumes once the editor exits.
// Port devd editor.go:14-96 (editorFinishedMsg/prepareEditor/
// readEditorResult/editInEditor) VERBATIM.
//
// editorBinary()'s cascade (E5 Task 5, bean bt-0l8c, updates the E3-era
// comment this file used to carry here): configuredEditor (Settings, E5's
// ~/.config/beans-tui/config.yaml -- now wired, see configuredEditor's own
// doc-stamp below) -> $VISUAL -> $EDITOR -> a bare "vi" fallback
// (design-spec §7 says literally "$EDITOR"; "vi" is the POSIX default,
// portable everywhere -- devd's "nvim" assumption is not).
//
// glowRender (devd editor.go:98-126) is NOT ported here -- beans-tui already
// has its own copy in accordion.go (E2, view_detail_bean.go's Body-section
// render path); this file only carries the Editor-Suspend half.

import (
	"os"
	"os/exec"
	"strings"

	"beans-tui/internal/data"
	tea "github.com/charmbracelet/bubbletea"
)

// editorFinishedMsg carries an $EDITOR-Suspend session's result back into the
// Update loop (Port devd editor.go:21-25). The caller (keyNodeAction's
// ctrl+e branch, update.go) remembers WHICH bean is being edited via
// m.editorTarget, captured BEFORE the suspend fires -- content/changed/err
// alone carry no bean identity.
type editorFinishedMsg struct {
	content string
	changed bool
	err     error
}

// configuredEditor is E5's Settings override (design decision c, PFLICHT:
// Settings.Editor > $VISUAL > $EDITOR > vi -- bean bt-0l8c, epic bt-5h4d).
// Package-level like devd's own configuredEditor: set once at TUI-start
// (app.go Run(), config.LoadSettings) and LIVE-updated by the Settings-Form's
// submit (box_form_settings.go, Port devd DD2-221). Default "" -- a NEW
// layer STRICTLY above the pre-existing $VISUAL/$EDITOR/vi cascade below, so
// an empty value (out-of-the-box, config.yaml never touched) falls straight
// through unchanged: TestEditorBinaryResolvesVisualThenEditorThenVi remains
// green without modification.
var configuredEditor string

// editorBinary resolves the editor command to launch (design decision c):
// configuredEditor (Settings, E5) -> $VISUAL -> $EDITOR -> a bare "vi"
// fallback (POSIX default -- portable everywhere, unlike devd's "nvim"
// assumption; design-spec §7 says literally "$EDITOR"). A value may carry
// arguments (e.g. "code -w"), split on whitespace like devd's own
// editorBinary.
func editorBinary() []string {
	if ed := strings.TrimSpace(configuredEditor); ed != "" {
		return strings.Fields(ed)
	}
	for _, envVar := range []string{"VISUAL", "EDITOR"} {
		if ed := strings.TrimSpace(os.Getenv(envVar)); ed != "" {
			return strings.Fields(ed)
		}
	}
	return []string{"vi"}
}

// prepareEditor writes initial into a temp file and builds the exec.Cmd that
// opens the resolved editor on it. Factored out for testability (no tea
// runtime needed, Port devd editor.go:44-68 comment). The caller owns
// os.Remove(path).
func prepareEditor(initial, suffix string) (path string, cmd *exec.Cmd, err error) {
	if suffix == "" {
		suffix = ".md"
	}
	f, err := os.CreateTemp("", "beans-tui-*"+suffix)
	if err != nil {
		return "", nil, err
	}
	path = f.Name()
	if _, err = f.WriteString(initial); err != nil {
		f.Close()
		os.Remove(path)
		return "", nil, err
	}
	if err = f.Close(); err != nil {
		os.Remove(path)
		return "", nil, err
	}
	bin := editorBinary()
	cmd = exec.Command(bin[0], append(bin[1:], path)...)
	return path, cmd, nil
}

// readEditorResult reads the (possibly edited) file back, compares against
// initial, and cleans up the temp file. runErr is the editor process's own
// exit error (Port devd editor.go:70-83).
func readEditorResult(path, initial string, runErr error) editorFinishedMsg {
	defer os.Remove(path)
	if runErr != nil {
		return editorFinishedMsg{err: runErr}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return editorFinishedMsg{err: err}
	}
	content := string(data)
	return editorFinishedMsg{content: content, changed: content != initial}
}

// editInEditor opens initial in the resolved editor via a tea.ExecProcess
// suspend, returning an editorFinishedMsg once the editor exits (Port devd
// editor.go:85-96). suffix controls the temp file's extension (".md" for a
// bean's body).
func editInEditor(initial, suffix string) tea.Cmd {
	path, cmd, err := prepareEditor(initial, suffix)
	if err != nil {
		return func() tea.Msg { return editorFinishedMsg{err: err} }
	}
	return tea.ExecProcess(cmd, func(runErr error) tea.Msg {
		return readEditorResult(path, initial, runErr)
	})
}

// openBodyEditor suspends into $EDITOR on b's Body (B10, design-spec.md §15
// PF-16, bean bt-ntoz, E8 Task 6) -- the ONE shared helper for BOTH call
// sites that need to open the Body in $EDITOR: keyNodeAction's "ctrl+e"/"e"-
// on-BODY-section branch and keyDetailFocus's "enter"-on-BODY-section branch
// (update.go). Factored out specifically so B10's second call site (the new
// enter-on-BODY case) does not duplicate the etag-capture dance ctrl+e
// already established (F2, Review-Runde 2: the etag is captured HERE, at
// open time -- never a fresh m.beanETag(id) re-read later, see
// applyEditorFinished's own doc-stamp for the full lost-update rationale).
func (m model) openBodyEditor(b *data.Bean) (model, tea.Cmd) {
	m.editorTarget = b.ID
	m.editorETag = b.ETag
	return m, editInEditor(b.Body, ".md")
}
