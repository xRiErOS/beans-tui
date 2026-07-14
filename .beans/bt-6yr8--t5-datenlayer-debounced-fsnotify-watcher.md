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

## Review-Fixes (Runde 2)

T5-Quality-Review-Findings B01/B02/I01/I02 behoben, TDD (neuer Test zuerst rot verifiziert):

- **B01 (high, synchroner Stop)**: `stop()` wartete nicht auf das Goroutine-Ende → der
  `timerC`-Select-Branch konnte `onChange()` noch nach Rückkehr von `stop()` auslösen
  (Gleichzeitigkeits-Race, `done` in dem Branch nie erneut geprüft). Fix: Goroutine bekommt
  `stoppedCh := make(chan struct{})` mit `defer close(stoppedCh)`; `stop()` ist jetzt
  `close(done); watcher.Close(); <-stoppedCh` (innerhalb `sync.Once`) — nach Rückkehr ist die
  Goroutine beweisbar beendet, Doku-Garantie "kein `onChange` nach `stop()`" ist damit wasserdicht.
  Der bisherige "Shutdown-priorisieren"-Pre-Check am Loop-Anfang war zusätzliche Komplexität ohne
  echten Nutzen (löste die Race nicht, da Select bei zwei bereiten Channels zufällig wählt) und
  wurde entfernt.
- **B02 (medium, archive/ nach Start angelegt)**: fsnotify ist nicht rekursiv — `.beans/archive`,
  das erst nach `Watch()`-Start angelegt wird, wurde nie beobachtet. Fix: im Event-Loop wird bei
  einem `Create`-Event, dessen (via `filepath.Clean` normalisierter) Pfad `archiveDir` entspricht,
  `watcher.Add(archiveDir)` nachgeholt (Fehler ignoriert, `.beans`-Watch bleibt in jedem Fall
  bestehen). Doku-Kommentar aktualisiert: archive/ wird abgedeckt, ob bei Start vorhanden ODER
  später angelegt.
- **I01 (minor)**: beide `watcher.Add`-Fehler jetzt gewrapped: `fmt.Errorf("watch %s: %w", dir, err)`.
- **I02 (minor)**: Tests überschreiben `debounceDuration` jetzt auf 20ms (`withTestDebounce` +
  `t.Cleanup`-Restore), Wartefenster entsprechend verschärft (Callback-Erwartung ≤500ms,
  Quiet-Window ~5× Debounce). Neu: `TestWatcherStopDuringBurst` (Schreiben + sofortiges `stop()`
  mitten in der Debounce-Phase → deterministisch 0 Callbacks danach, dank synchronem Stop) und
  `TestWatcherPicksUpLateArchiveDir` (Watch ohne archive/ gestartet, archive/ nach Start angelegt,
  Settle abwarten, dann Schreiben INNERHALB archive/ → zweiter `onChange`, beweist dynamisches
  Re-Add). Beide Tests vorab gegen den ungefixten Code verifiziert: `TestWatcherPicksUpLateArchiveDir`
  schlug fehl (fehlendes Re-Add), `TestWatcherStopDuringBurst` deckte die B01-Race sogar direkt als
  `-race`-Data-Race auf (Goroutine liest `debounceDuration` noch, während `t.Cleanup` es nach
  `stop()`-Rückkehr zurücksetzt) — härtester Beleg für B01.

Verifiziert: `go test ./internal/data/ -run Watcher -race -count=3` grün (alle 4 Tests × 3 Runs),
volle `go test ./...` grün, `gofmt -l .` clean, `go vet ./...` clean.
