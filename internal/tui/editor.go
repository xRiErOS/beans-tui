package tui

// editor.go — $EDITOR-Suspend for the Body field (`ctrl+e`, E3 Task 5, bean
// bt-sl45, design decisions c/h): tea.ExecProcess suspends the running
// Bubble Tea program, hands the terminal to the user's editor on a temp file
// seeded with the bean's current body, and resumes once the editor exits.
// Port devd editor.go:14-96 (editorFinishedMsg/prepareEditor/
// readEditorResult/editInEditor) VERBATIM -- the ONE deviation is
// editorBinary(): devd resolves configuredEditor (a TUI-wide setting from
// devd's config.yaml, an E5-equivalent scope that does not exist yet here),
// while beans-tui resolves $VISUAL -> $EDITOR -> a bare "vi" fallback
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

// editorBinary resolves the editor command to launch (design decision c):
// $VISUAL -> $EDITOR -> a bare "vi" fallback (POSIX default -- portable
// everywhere, unlike devd's "nvim" assumption; design-spec §7 says literally
// "$EDITOR"). A value may carry arguments (e.g. "code -w"), split on
// whitespace like devd's own editorBinary. No configuredEditor here: that is
// E5's ~/.config/beans-tui/config.yaml, which does not exist yet.
func editorBinary() []string {
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
