---
# bt-gzu6
title: 'T5 Backlog-View V3 (b): Master-Detail, geteilter Filter/Suche, Sort-Toggle S'
status: todo
type: task
priority: high
created_at: 2026-07-14T21:57:33Z
updated_at: 2026-07-14T21:57:33Z
parent: bt-aq5s
blocked_by:
    - bt-ms0k
    - bt-9ldr
---

Ziel: Backlog-View V3 (`b`) — parentlose+ready beans (idx.Backlog()), Master-Detail
(Wiederverwendung Task-1-Accordion via focusedBean()-Dispatcher aus T2), geteilter
Such-/Filter-Zustand (T3/T4), Sort-Toggle `S` (status/prio/created/updated,
client-seitig, kein beans-CLI-Resort nötig). Windowing wird wiederverwendet
(windowAround, E1) — kein Neubau.

Plan: docs/plans/v1-port/epic-E2-plan.md »Task 5«.

## Akzeptanz
- [ ] view_browse_backlog.go: viewBacklog + Render (Master-Detail, D03
      Border-Fokus-Tausch), reuse renderPane/masterDetailWidths/windowAround
- [ ] backlogVisible(): idx.Backlog() + geteilte Suche/Filter-Prädikate (KEIN
      Duplikat zu Tree — eine gemeinsame beanMatches()-Funktion)
- [ ] Sort-Toggle S zyklisch status->prio->created->updated->status
      (data.StatusRank/PriorityRank aus T1, CreatedAt/UpdatedAt nil-sicher)
- [ ] `b` öffnet aus Tree, esc/b zurück — geteilter Filter/Suche-Zustand bleibt
      beim Zurückwechseln erhalten
- [ ] backlog.golden neu + Determinismus-Test
- [ ] go test ./... grün, make build ok
