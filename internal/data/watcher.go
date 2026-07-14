package data

import (
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

// Watch starts an fsnotify watch on <repoDir>/.beans (and, if it exists,
// <repoDir>/.beans/archive) and calls onChange once a burst of matching
// filesystem events (Create/Write/Remove/Rename) settles for
// debounceDuration -- a burst of N events yields exactly one onChange
// call after the quiet period, not N calls.
//
// onChange takes no arguments and carries no event payload: the consumer
// (T8) always reacts with a full reload (Client.List + NewIndex), never a
// partial/incremental update (design decision D02).
//
// Scope cut: the watched directory is hardcoded to ".beans" (design-spec
// D02/D06 convention). Reading a custom `beans.path` from .beans.yml is
// out of scope for this function -- callers relying on a non-default
// beans.path will not get watch coverage.
//
// Watch returns a stop func that shuts down the watcher goroutine cleanly.
// stop is idempotent and safe to call more than once. After stop() returns,
// no further onChange calls will occur.
func Watch(repoDir string, onChange func()) (stop func(), err error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	beansDir := filepath.Join(repoDir, ".beans")
	if err := watcher.Add(beansDir); err != nil {
		watcher.Close()
		return nil, err
	}

	archiveDir := filepath.Join(beansDir, "archive")
	if info, statErr := os.Stat(archiveDir); statErr == nil && info.IsDir() {
		if err := watcher.Add(archiveDir); err != nil {
			watcher.Close()
			return nil, err
		}
	}

	done := make(chan struct{})

	go func() {
		var timer *time.Timer
		var timerC <-chan time.Time

		defer func() {
			if timer != nil {
				timer.Stop()
			}
		}()

		for {
			// Prioritize shutdown: if done is already closed, exit before
			// considering any newly-fired timer or fs event.
			select {
			case <-done:
				return
			default:
			}

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
		})
	}

	return stop, nil
}
