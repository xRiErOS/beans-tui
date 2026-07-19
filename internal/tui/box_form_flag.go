package tui

// box_form_flag.go — env gate for the experimental jira-style box-form
// detail render (docs/plans/jira-style-experiment/, S2b). Mirrors
// theme.asciiIcons()'s env pattern (internal/theme/theme.go) 1:1: read
// per-call (no caching), so tests can flip it via t.Setenv without any
// package-state reset dance. Default OFF -> every existing golden/behavior
// stays byte-identical; opt in via BT_BOXFORM=1|true|yes|on.

import (
	"os"
	"strings"
)

// boxFormEnabled reports whether the experimental jira-style box-form detail
// render is active (env BT_BOXFORM=1|true|yes|on). Default off — existing
// goldens/behavior unchanged. Read per-call (no caching), like asciiIcons().
func boxFormEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("BT_BOXFORM"))) {
	case "1", "true", "yes", "on":
		return true
	}
	return false
}
