---
# bt-snkb
title: 'T1 Modul-Skeleton: go.mod, cobra (root+tui), Makefile'
status: todo
type: task
priority: high
created_at: 2026-07-14T18:34:04Z
updated_at: 2026-07-14T18:34:04Z
parent: bt-blsy
---

Plan: implementation-plan.md »E1 Task 1« (voll-granulare Steps + Testcode dort).

## Akzeptanz
- [ ] `go.mod` Modul `beans-tui`, cobra v1.10.2
- [ ] `cmd/root.go` NewRootCmd (Use `bt [repo-pfad]`), `cmd/tui.go` tui-Subcommand, runTUI-Stub
- [ ] Test `cmd/root_test.go` grün (TestRootStartsTUIByDefault)
- [ ] `Makefile` build/test/clean mit `command go`; `make build` → `bin/bt`
