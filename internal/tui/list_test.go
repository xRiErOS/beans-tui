package tui

import "testing"

// TestListStateSetLenClampsCursor guards the clamp behavior of setLen: the
// cursor must never point outside [0, n).
func TestListStateSetLenClampsCursor(t *testing.T) {
	l := &listState{cursor: 5}
	l.setLen(3)
	if l.cursor != 2 {
		t.Errorf("cursor=%d, want 2 (clamped to len-1)", l.cursor)
	}

	l2 := &listState{cursor: -1}
	l2.setLen(3)
	if l2.cursor != 0 {
		t.Errorf("cursor=%d, want 0 (clamped up from negative)", l2.cursor)
	}

	l3 := &listState{cursor: 0}
	l3.setLen(0)
	if l3.cursor != 0 {
		t.Errorf("cursor=%d, want 0 on empty list", l3.cursor)
	}
}

// TestListStateMoveClamps guards move()'s bounds: it never leaves [0, length).
func TestListStateMoveClamps(t *testing.T) {
	l := &listState{length: 5}
	l.move(2)
	if l.cursor != 2 {
		t.Errorf("cursor=%d, want 2", l.cursor)
	}
	l.move(-10)
	if l.cursor != 0 {
		t.Errorf("cursor=%d, want 0 (clamped at floor)", l.cursor)
	}
	l.move(10)
	if l.cursor != 4 {
		t.Errorf("cursor=%d, want 4 (clamped at length-1)", l.cursor)
	}
}

// TestListStateMoveOnEmptyList guards the length==0 special case: move must
// pin the cursor to 0 rather than letting it drift negative/positive.
func TestListStateMoveOnEmptyList(t *testing.T) {
	l := &listState{length: 0, cursor: 3}
	l.move(1)
	if l.cursor != 0 {
		t.Errorf("cursor=%d, want 0 on empty list", l.cursor)
	}
}

// TestListStateReset guards reset(): always back to cursor 0.
func TestListStateReset(t *testing.T) {
	l := &listState{length: 5, cursor: 3}
	l.reset()
	if l.cursor != 0 {
		t.Errorf("cursor=%d, want 0 after reset", l.cursor)
	}
}
