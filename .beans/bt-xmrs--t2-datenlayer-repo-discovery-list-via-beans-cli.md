---
# bt-xmrs
title: 'T2 Datenlayer: Repo-Discovery + List via beans-CLI'
status: todo
type: task
priority: high
created_at: 2026-07-14T18:34:04Z
updated_at: 2026-07-14T18:34:04Z
parent: bt-blsy
blocked_by:
    - bt-snkb
---

Plan: implementation-plan.md »E1 Task 2« (Bean-Typ-Vertrag + Client-Code dort).

## Akzeptanz
- [ ] `internal/data/bean.go` Bean-Struct (JSON-Felder gegen echten `beans list --json --full`-Output verifiziert)
- [ ] `internal/data/discover.go` FindRepo (aufwärts bis .beans.yml)
- [ ] `internal/data/client.go` Client.List via Subprocess, Fehler mit stderr
- [ ] Tests mit tmp-Fixture-Repo (newTestRepo, requireBeansBinary-Guard) grün
