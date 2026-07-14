package data

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

// TestWatcherFiresOnceForBurst verifies the debounce contract: a burst of
// filesystem writes in quick succession collapses into exactly one
// onChange call after the quiet period, and a subsequent single write
// produces exactly one more call.
func TestWatcherFiresOnceForBurst(t *testing.T) {
	dir := t.TempDir()
	beansDir := filepath.Join(dir, ".beans")
	if err := os.MkdirAll(beansDir, 0o755); err != nil {
		t.Fatalf("mkdir .beans: %v", err)
	}

	var count int32
	changed := make(chan struct{}, 10)
	onChange := func() {
		atomic.AddInt32(&count, 1)
		changed <- struct{}{}
	}

	stop, err := Watch(dir, onChange)
	if err != nil {
		t.Fatalf("Watch() error = %v", err)
	}
	defer stop()

	// Burst: 3 quick writes to the same file.
	target := filepath.Join(beansDir, "burst.md")
	for i := 0; i < 3; i++ {
		if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
			t.Fatalf("write burst file: %v", err)
		}
	}

	select {
	case <-changed:
	case <-time.After(1 * time.Second):
		t.Fatal("onChange not called within 1s after burst")
	}

	// No second call should follow for the same burst -- wait comfortably
	// past the debounce window.
	select {
	case <-changed:
		t.Fatal("onChange called a second time for a single burst")
	case <-time.After(400 * time.Millisecond):
	}

	if got := atomic.LoadInt32(&count); got != 1 {
		t.Fatalf("onChange called %d times after burst, want 1", got)
	}

	// A fresh, single write after the quiet period is a new debounce
	// window and must produce exactly one more call.
	if err := os.WriteFile(filepath.Join(beansDir, "second.md"), []byte("y"), 0o644); err != nil {
		t.Fatalf("write second file: %v", err)
	}

	select {
	case <-changed:
	case <-time.After(1 * time.Second):
		t.Fatal("onChange not called for second write")
	}

	if got := atomic.LoadInt32(&count); got != 2 {
		t.Fatalf("onChange called %d times total, want 2", got)
	}
}

// TestWatcherStops verifies that after stop() returns, further filesystem
// events do not trigger onChange, and the watcher goroutine has shut down.
func TestWatcherStops(t *testing.T) {
	dir := t.TempDir()
	beansDir := filepath.Join(dir, ".beans")
	if err := os.MkdirAll(beansDir, 0o755); err != nil {
		t.Fatalf("mkdir .beans: %v", err)
	}

	var count int32
	stop, err := Watch(dir, func() { atomic.AddInt32(&count, 1) })
	if err != nil {
		t.Fatalf("Watch() error = %v", err)
	}

	stop()

	if err := os.WriteFile(filepath.Join(beansDir, "afterstop.md"), []byte("z"), 0o644); err != nil {
		t.Fatalf("write after stop: %v", err)
	}

	// Wait comfortably past the debounce window: no onChange should fire.
	time.Sleep(400 * time.Millisecond)

	if got := atomic.LoadInt32(&count); got != 0 {
		t.Fatalf("onChange called %d times after stop, want 0", got)
	}

	// Calling stop() a second time must not panic (idempotent shutdown).
	stop()
}
