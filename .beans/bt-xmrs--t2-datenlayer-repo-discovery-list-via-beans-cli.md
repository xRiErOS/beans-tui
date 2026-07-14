---
# bt-xmrs
title: 'T2 Datenlayer: Repo-Discovery + List via beans-CLI'
status: completed
type: task
priority: high
created_at: 2026-07-14T18:34:04Z
updated_at: 2026-07-14T18:45:30Z
parent: bt-blsy
blocked_by:
    - bt-snkb
---

Plan: implementation-plan.md »E1 Task 2« (Bean-Typ-Vertrag + Client-Code dort).

## Akzeptanz
- [x] `internal/data/bean.go` Bean-Struct (JSON-Felder gegen echten `beans list --json --full`-Output verifiziert)
- [x] `internal/data/discover.go` FindRepo (aufwärts bis .beans.yml)
- [x] `internal/data/client.go` Client.List via Subprocess, Fehler mit stderr
- [x] Tests mit tmp-Fixture-Repo (newTestRepo, requireBeansBinary-Guard) grün

## Summary of Changes

- `internal/data/bean.go`: `Bean`-Struct exakt wie im Plan-Vertrag — Feldnamen/JSON-Tags
  gegen echten `beans list --json --full` + `beans show <id> --json` Output verifiziert
  (Testfixtures mit `tags`/`parent`/`blocking`/`blocked_by` in einem Scratch-Repo erzeugt).
  **Keine Abweichung** gefunden: alle Felder (`tags`, `parent`, `blocking`, `blocked_by`)
  matchen 1:1, sie fehlen im JSON nur wenn unset (von `json.Unmarshal` als Zero-Value
  sauber gehandhabt).
- `internal/data/discover.go`: `FindRepo(start string) (string, error)` — läuft aufwärts
  bis ein Verzeichnis mit `.beans.yml` gefunden wird, Fehler beim Erreichen der FS-Root.
- `internal/data/client.go`: `Client{RepoDir}` mit `List()` (ruft `beans list --json --full`)
  und privatem `run()` (stderr im Fehler-Wrap).
- Tests (TDD, erst FAIL vor Implementierung verifiziert):
  `internal/data/testrepo_test.go` (`newTestRepo`, `requireBeansBinary`),
  `internal/data/discover_test.go` (`TestDiscoverFindsConfigUpward`,
  `TestDiscoverErrorsWhenNoConfigFound`), `internal/data/client_test.go`
  (`TestListReturnsAllBeansWithBody`, `TestListErrorIncludesStderr`).
- `command go test ./...`, `command go vet ./...`, `gofmt -l .` (leer) und
  `command go build -o bin/bt .` alle grün.
