---
# bt-6yr8
title: 'T5 Datenlayer: Debounced fsnotify-Watcher'
status: completed
type: task
priority: high
created_at: 2026-07-14T18:34:04Z
updated_at: 2026-07-14T19:47:45Z
parent: bt-blsy
blocked_by:
    - bt-snkb
---

Plan: implementation-plan.md »E1 Task 5«.

## Akzeptanz
- [x] `internal/data/watcher.go` Watch(repoDir, onChange) mit stop-Func, Debounce 150 ms, .beans/ + archive/
- [x] TestWatcherFiresOnceForBurst grün

## Summary of Changes

- `internal/data/watcher.go`: `Watch(repoDir string, onChange func()) (stop func(), err error)`.
  fsnotify watch on `<repoDir>/.beans` and `<repoDir>/.beans/archive` (if it exists), reacts to
  Create/Write/Remove/Rename (Chmod ignored), 150ms debounce with timer-reset so a burst
  collapses into one `onChange`. `stop()` is idempotent, closes cleanly (no goroutine leak, no
  `onChange` after `stop()` returns). Debounce duration is a package-level var for testability,
  default 150ms. Scope cut documented: `.beans` dir name is hardcoded (D02/D06); custom
  `beans.path` from `.beans.yml` is out of scope.
- `internal/data/watcher_test.go`: `TestWatcherFiresOnceForBurst` (3 quick writes → exactly 1
  onChange within 1s, then a single follow-up write → exactly 1 more) and `TestWatcherStops`
  (after `stop()`, further writes produce no `onChange`; `stop()` called twice does not panic).
- `go.mod`/`go.sum`: added `github.com/fsnotify/fsnotify v1.9.0` (+ transitive `golang.org/x/sys`).

Verified: TDD red→green (compile-fail before `Watch` existed, confirmed pass after), `go test
./internal/data/ -run Watcher -race -count=5` green, `go build ./...`, `go vet ./...`, `gofmt -l .`
clean, full `go test ./...` green.
