---
# bt-gbfe
title: 'T3 Datenlayer: In-Memory-Index (Children/Roots/Backlog/Tags)'
status: todo
type: task
priority: high
created_at: 2026-07-14T18:34:04Z
updated_at: 2026-07-14T18:34:04Z
parent: bt-blsy
blocked_by:
    - bt-xmrs
---

Plan: implementation-plan.md »E1 Task 3«.

## Akzeptanz
- [ ] `internal/data/index.go`: ByID, Children (sortiert Status→Prio→Type→Titel), Roots, Backlog (parentlos, todo/draft), WithTag
- [ ] Reine Unit-Tests ohne beans-Binary grün (4 Tests laut Plan)
