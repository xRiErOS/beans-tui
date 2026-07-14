---
# bt-gbfe
title: 'T3 Datenlayer: In-Memory-Index (Children/Roots/Backlog/Tags)'
status: completed
type: task
priority: high
created_at: 2026-07-14T18:34:04Z
updated_at: 2026-07-14T18:55:25Z
parent: bt-blsy
blocked_by:
    - bt-xmrs
---

Plan: implementation-plan.md »E1 Task 3«.

## Akzeptanz
- [x] `internal/data/index.go`: ByID, Children (sortiert Status→Prio→Type→Titel), Roots, Backlog (parentlos, todo/draft), WithTag
- [x] Reine Unit-Tests ohne beans-Binary grün (4 Tests laut Plan)

## Summary of Changes

- `internal/data/index.go`: `Index` mit `ByID map[string]*Bean`, `Children map[string][]*Bean`
  (Parent invertiert), `NewIndex(beans []Bean) *Index`, `Roots(types ...string) []*Bean`,
  `Backlog() []*Bean`, `WithTag(tag string) []*Bean`. Sortierlogik einmalig in
  unexported `sortBeans`/`rank` (Status→Priority→Type→Title, case-insensitiv;
  leere Priority default "normal"; unbekannte Werte sortieren zuletzt).
- `internal/data/index_test.go`: `TestIndexChildrenSorted`, `TestRootsMilestonesFirst`
  (inkl. Filter- und Parent-Exclusion-Subtests), `TestBacklogExcludesParented`,
  `TestWithTagToReview` — alle mit voller Order-Assertion (nicht nur Membership),
  reine Unit-Tests ohne beans-Binary.
- TDD: Tests zuerst geschrieben (`undefined: NewIndex`, Build-Fail), dann Implementierung,
  danach grün. `gofmt -l .` clean, `go vet ./...` clean, `go test ./...` (ganzes Repo) grün.
