// Package clip kopiert Text in die System-Zwischenablage. Primärweg ist OSC52
// (funktioniert lokal, über SSH und in tmux/screen — wichtig, da oft remote
// entwickelt wird). Zusätzlich best-effort nativ (pbcopy/wl-copy/xclip/xsel),
// damit lokal direkt die System-Zwischenablage gesetzt ist.
//
// Port devd VERBATIM (~/Obsidian/tools/DeveloperDashboard/apps/cli-go/
// internal/clip/clip.go, E5 Task 3, bean bt-e4a6, epic-E5-plan.md »Task 3«
// Step 4) -- no beans-tui-specific changes, same package/API shape as the
// devd source this was copied from.
package clip

import (
	"os"
	"os/exec"
	"strings"

	osc52 "github.com/aymanbagabas/go-osc52/v2"
)

// Copy schreibt s in die Zwischenablage via OSC52 (+ nativ best-effort).
func Copy(s string) error {
	seq := osc52.New(s)
	switch {
	case os.Getenv("TMUX") != "":
		seq = seq.Tmux()
	case strings.HasPrefix(os.Getenv("TERM"), "screen"):
		seq = seq.Screen()
	}
	// stderr, nicht stdout — Bubble-Tea rendert auf stdout; OSC52 erzeugt keine
	// sichtbare Ausgabe und erreicht das Terminal über dieselbe PTY.
	_, err := seq.WriteTo(os.Stderr)
	nativeCopy(s) // best-effort, Fehler ignoriert
	return err
}

var nativeCandidates = [][]string{
	{"pbcopy"},
	{"wl-copy"},
	{"xclip", "-selection", "clipboard"},
	{"xsel", "--clipboard", "--input"},
}

func nativeCopy(s string) {
	for _, c := range nativeCandidates {
		if _, err := exec.LookPath(c[0]); err != nil {
			continue
		}
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Stdin = strings.NewReader(s)
		_ = cmd.Run()
		return
	}
}
