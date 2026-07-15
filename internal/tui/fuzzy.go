package tui

// fuzzy.go — subsequence fuzzy matcher for the Command-Center (E4 Task 1,
// bean bt-jpgn, design decision a). VERBATIM port of devd
// overlay_palette.go:60-71 (fuzzyMatch), only the package name changes — no
// new dependency (there is no fuzzy-matching lib in go.mod), sufficient for
// an action list (~10 entries) and a repo with a low-hundreds bean count
// (Task 2's palette-scoped bean search). Swappable later behind this same
// `fuzzyMatch(query, target string) bool` signature if a graduated
// score-based matcher (e.g. sahilm/fuzzy) is ever needed.

import "strings"

// fuzzyMatch reports whether query is a (case-insensitive, rune-based)
// subsequence of target. An empty query matches everything.
func fuzzyMatch(query, target string) bool {
	q := []rune(strings.ToLower(query))
	t := []rune(strings.ToLower(target))
	i := 0
	for _, tc := range t {
		if i < len(q) && tc == q[i] {
			i++
		}
	}
	return i == len(q)
}
