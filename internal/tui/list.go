package tui

// listState is a minimal, testable cursor over a list of a given length.
// Ported unchanged from devd (~/Obsidian/tools/DeveloperDashboard/apps/cli-go/
// internal/tui/list.go) — no devd-API coupling to strip, pure state.
type listState struct {
	length int
	cursor int
}

// setLen updates the list length and clamps the cursor into [0, n).
func (l *listState) setLen(n int) {
	l.length = n
	if l.cursor >= n {
		l.cursor = n - 1
	}
	if l.cursor < 0 {
		l.cursor = 0
	}
}

// move shifts the cursor by d, clamped to the list bounds.
func (l *listState) move(d int) {
	if l.length == 0 {
		l.cursor = 0
		return
	}
	l.cursor += d
	if l.cursor < 0 {
		l.cursor = 0
	}
	if l.cursor >= l.length {
		l.cursor = l.length - 1
	}
}

// reset sets the cursor back to 0 (e.g. when the parent selection changes).
func (l *listState) reset() { l.cursor = 0 }
