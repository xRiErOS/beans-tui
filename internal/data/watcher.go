package data

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// watchedOps is the event mask Watch reacts to. Chmod is deliberately
// excluded -- permission-only changes are not bean-data changes.
const watchedOps = fsnotify.Create | fsnotify.Write | fsnotify.Remove | fsnotify.Rename

// debounceDuration is the quiet period after the last matching filesystem
// event before onChange fires. It is a package-level var (not a const) so
// tests can shrink it if timing ever needs tightening; production code
// should leave it at the default 150ms.
var debounceDuration = 150 * time.Millisecond

// Watch starts an fsnotify watch on <repoDir>/.beans (and
// <repoDir>/.beans/archive, watched if present at start OR created later --
// fsnotify is non-recursive, so Watch dynamically Add()s archive/ the
// moment it observes the Create event for it) and calls onChange once a
// burst of matching filesystem events (Create/Write/Remove/Rename) settles
// for debounceDuration -- a burst of N events yields exactly one onChange
// call after the quiet period, not N calls.
//
// onChange takes no arguments and carries no event payload: the consumer
// (T8) always reacts with a full reload (Client.List + NewIndex), never a
// partial/incremental update (design decision D02).
//
// B05 (MANDATORY, bean bt-7jr8): onChange runs ON THE WATCHER GOROUTINE
// ITSELF (synchronously, from inside the select loop below), so it must
// NEVER call the returned stop synchronously from within the callback --
// stop blocks until this very goroutine exits (see stoppedCh below), so
// calling it from inside onChange deadlocks the watcher forever. The T8
// consumer (internal/tui) only ever dispatches an async tea.Msg from
// onChange (tea.Program.Send, documented safe to call from any goroutine)
// and calls stop solely from its own teardown path, after the tea.Program
// has stopped running.
//
// Scope cut: the watched directory is hardcoded to ".beans" (design-spec
// D02/D06 convention). Reading a custom `beans.path` from .beans.yml is
// out of scope for this function -- callers relying on a non-default
// beans.path will not get watch coverage.
//
// Watch returns a stop func that shuts down the watcher goroutine cleanly.
// stop is idempotent, safe to call more than once, and synchronous: it
// blocks until the watcher goroutine has fully exited. After stop()
// returns, no further onChange calls will occur.
func Watch(repoDir string, onChange func()) (stop func(), err error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	beansDir := filepath.Join(repoDir, ".beans")
	if err := watcher.Add(beansDir); err != nil {
		watcher.Close()
		return nil, fmt.Errorf("watch %s: %w", beansDir, err)
	}

	archiveDir := filepath.Join(beansDir, "archive")
	if info, statErr := os.Stat(archiveDir); statErr == nil && info.IsDir() {
		if err := watcher.Add(archiveDir); err != nil {
			watcher.Close()
			return nil, fmt.Errorf("watch %s: %w", archiveDir, err)
		}
	}

	done := make(chan struct{})
	stoppedCh := make(chan struct{})

	go func() {
		defer close(stoppedCh)

		var timer *time.Timer
		var timerC <-chan time.Time

		defer func() {
			if timer != nil {
				timer.Stop()
			}
		}()

		for {
			select {
			case <-done:
				return

			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&watchedOps == 0 {
					continue
				}
				if event.Op&fsnotify.Create != 0 && filepath.Clean(event.Name) == archiveDir {
					// fsnotify is non-recursive: archive/ was not present
					// (or not yet watched) at Watch() time. Now that it
					// exists, start watching it too. Ignore the error --
					// keep watching .beans regardless (T5 review B02).
					_ = watcher.Add(archiveDir)
				}
				if timer == nil {
					timer = time.NewTimer(debounceDuration)
				} else {
					if !timer.Stop() {
						select {
						case <-timer.C:
						default:
						}
					}
					timer.Reset(debounceDuration)
				}
				timerC = timer.C

			case <-timerC:
				timerC = nil
				onChange()

			case _, ok := <-watcher.Errors:
				if !ok {
					return
				}
				// Non-fatal: keep watching on read/decode errors from the
				// underlying OS watch.
			}
		}
	}()

	var once sync.Once
	stop = func() {
		once.Do(func() {
			close(done)
			watcher.Close()
			<-stoppedCh
		})
	}

	return stop, nil
}

// StartWatch is switchRepoCmd's (internal/tui/messages.go) own call site for
// starting a fresh watcher during a repo switch (E5 Task 6, bean bt-zhwl) --
// an identically-shaped, purely pass-through wrapper around Watch (same
// params, same contract, same synchronous/idempotent stop). Kept as its own
// name (not just a second call to Watch inline at every switch-driven call
// site) so "start a NEW watch because the PO just switched repos" reads
// textually distinct from Watch's original steady-state caller (app.go's
// Run, the very first watch of a session) at a glance -- no behavioral
// difference between the two names, this is a readability seam only.
func StartWatch(repoDir string, notify func()) (stop func(), err error) {
	return Watch(repoDir, notify)
}
