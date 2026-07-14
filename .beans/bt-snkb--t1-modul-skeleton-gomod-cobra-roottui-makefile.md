---
# bt-snkb
title: 'T1 Modul-Skeleton: go.mod, cobra (root+tui), Makefile'
status: completed
type: task
priority: high
created_at: 2026-07-14T18:34:04Z
updated_at: 2026-07-14T18:38:04Z
parent: bt-blsy
---

Plan: implementation-plan.md »E1 Task 1« (voll-granulare Steps + Testcode dort).

## Akzeptanz
- [x] `go.mod` Modul `beans-tui`, cobra v1.10.2
- [x] `cmd/root.go` NewRootCmd (Use `bt [repo-pfad]`), `cmd/tui.go` tui-Subcommand, runTUI-Stub
- [x] Test `cmd/root_test.go` grün (TestRootStartsTUIByDefault)
- [x] `Makefile` build/test/clean mit `command go`; `make build` → `bin/bt`

## Summary of Changes

TDD-Ablauf befolgt: `cmd/root_test.go` zuerst geschrieben, Fehlschlag
(`NewRootCmd undefined`) via `command go test ./cmd/` bestätigt, danach
implementiert.

- `go.mod`/`go.sum`: Modul `beans-tui`, direkte Dependency `github.com/spf13/cobra v1.10.2`
  (`go mod tidy` gelaufen — mousetrap/pflag korrekt als indirect markiert).
- `main.go`: ruft `cmd.Execute()`.
- `cmd/root.go`: `NewRootCmd()` (Use `bt [repo-pfad]`, `Args: MaximumNArgs(1)`,
  RunE startet TUI mit optionalem Pfad-Arg via `runTUI`), `Execute()`.
- `cmd/tui.go`: expliziter `tui`-Subcommand (`newTUICmd`, gleicher RunE) +
  `runTUI(path string) error { return nil }`-Stub.
- `cmd/root_test.go`: `TestRootStartsTUIByDefault` — prüft `Use`-String und
  dass `tui`-Subcommand über `cmd.Find` auffindbar ist.
- `Makefile`: `build`/`test`/`clean`-Targets (tab-indentiert, `.PHONY`),
  alle Go-Aufrufe via `command go`. `make build` verifiziert → erzeugt `bin/bt`
  (Mach-O arm64 Binary), `bt --help`/`bt tui --help` manuell geprüft.

Verifiziert: `command go test ./...` PASS, `command go vet ./...` clean,
`command gofmt -l .` clean, `make build`/`make test`/`make clean` funktionieren.
