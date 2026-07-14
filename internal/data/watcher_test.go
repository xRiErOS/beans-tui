package data

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

// testDebounce is the shortened debounce duration used across this file's
// tests so wait margins can be tight without becoming flaky.
const testDebounce = 20 * time.Millisecond

// withTestDebounce overrides the package-level debounceDuration for the
// duration of the test and restores the original value on cleanup.
func withTestDebounce(t *testing.T, d time.Duration) {
	t.Helper()
	orig := debounceDuration
	debounceDuration = d
	t.Cleanup(func() { debounceDuration = orig })
}

// TestWatcherFiresOnceForBurst verifies the debounce contract: a burst of
// filesystem writes in quick succession collapses into exactly one
// onChange call after the quiet period, and a subsequent single write
// produces exactly one more call.
func TestWatcherFiresOnceForBurst(t *testing.T) {
	withTestDebounce(t, testDebounce)

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
	case <-time.After(500 * time.Millisecond):
		t.Fatal("onChange not called within 500ms after burst")
	}

	// No second call should follow for the same burst -- wait comfortably
	// past the debounce window (~5x the debounce duration).
	select {
	case <-changed:
		t.Fatal("onChange called a second time for a single burst")
	case <-time.After(5 * testDebounce):
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
	case <-time.After(500 * time.Millisecond):
		t.Fatal("onChange not called for second write")
	}

	if got := atomic.LoadInt32(&count); got != 2 {
		t.Fatalf("onChange called %d times total, want 2", got)
	}
}

// TestWatcherStops verifies that after stop() returns, further filesystem
// events do not trigger onChange, and the watcher goroutine has shut down.
func TestWatcherStops(t *testing.T) {
	withTestDebounce(t, testDebounce)

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

	// Wait comfortably past the debounce window (~5x): no onChange should fire.
	time.Sleep(5 * testDebounce)

	if got := atomic.LoadInt32(&count); got != 0 {
		t.Fatalf("onChange called %d times after stop, want 0", got)
	}

	// Calling stop() a second time must not panic (idempotent shutdown).
	stop()
}

// TestWatcherStopDuringBurst verifies that stop() called mid-debounce (a
// write already pending inside the quiet window) never lets onChange fire
// after stop() has returned. This is deterministic because stop() is
// synchronous: it does not return until the watcher goroutine has fully
// exited (B01 fix), so any in-flight onChange has already completed and no
// further select iteration -- and thus no further onChange -- can occur
// afterwards.
func TestWatcherStopDuringBurst(t *testing.T) {
	withTestDebounce(t, testDebounce)

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

	if err := os.WriteFile(filepath.Join(beansDir, "midburst.md"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	// Give the event a chance to reach the goroutine and start the debounce
	// timer, then stop mid-window -- well before testDebounce elapses.
	time.Sleep(testDebounce / 4)
	stop()

	// Small grace wait: by the time stop() returned above, the watcher
	// goroutine is already provably gone, so this only guards against a
	// flaky false negative rather than a real race.
	time.Sleep(5 * testDebounce)

	if got := atomic.LoadInt32(&count); got != 0 {
		t.Fatalf("onChange called %d times after stop-during-burst, want 0", got)
	}
}

// TestWatcherPicksUpLateArchiveDir verifies that .beans/archive is watched
// even when it does not exist at Watch() time: fsnotify is non-recursive,
// so the initial watcher.Add only covers .beans/ itself, but Watch() must
// dynamically Add() archive/ once it observes the Create event for it
// (B02 fix), and events written inside it afterward must still surface.
func TestWatcherPicksUpLateArchiveDir(t *testing.T) {
	withTestDebounce(t, testDebounce)

	dir := t.TempDir()
	beansDir := filepath.Join(dir, ".beans")
	if err := os.MkdirAll(beansDir, 0o755); err != nil {
		t.Fatalf("mkdir .beans: %v", err)
	}

	changed := make(chan struct{}, 10)
	stop, err := Watch(dir, func() { changed <- struct{}{} })
	if err != nil {
		t.Fatalf("Watch() error = %v", err)
	}
	defer stop()

	// archive/ does not exist yet at Watch() time -- create it now. The
	// mkdir is itself a Create event on the already-watched .beans dir and
	// settles into one onChange after the debounce window.
	archiveDir := filepath.Join(beansDir, "archive")
	if err := os.Mkdir(archiveDir, 0o755); err != nil {
		t.Fatalf("mkdir archive: %v", err)
	}

	select {
	case <-changed:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("onChange not called for late archive mkdir")
	}

	// Quiet window so the archive-mkdir burst is fully settled before the
	// next write starts a new one.
	time.Sleep(5 * testDebounce)

	// A write INSIDE the newly-created archive dir must also be observed --
	// only possible if Watch() dynamically re-Add()ed archiveDir when it
	// saw the Create event above.
	if err := os.WriteFile(filepath.Join(archiveDir, "inside.md"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write inside archive: %v", err)
	}

	select {
	case <-changed:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("onChange not called for write inside late-added archive dir")
	}
}
