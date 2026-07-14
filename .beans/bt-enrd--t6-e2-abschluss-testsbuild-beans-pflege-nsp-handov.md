---
# bt-enrd
title: 'T6 E2-Abschluss: Tests+Build, beans-Pflege, NSP-Handover'
status: todo
type: task
priority: high
created_at: 2026-07-14T21:57:40Z
updated_at: 2026-07-14T21:57:40Z
parent: bt-aq5s
blocked_by:
    - bt-2jve
    - bt-gzu6
---

Ziel: E2-Abschluss-Ritual (implementation-plan.md »Epos-Rituale«): Tests+Build grün,
beans-Pflege, Commit, ce-nsp-auto-Handover für E3.

Plan: docs/plans/v1-port/epic-E2-plan.md »Task 6«.

## Akzeptanz
- [ ] command go test ./... grün, command go build -o bin/bt . ok, gofmt/vet leer
- [ ] Alle E2-Task-beans (T1-T5) auf completed; Epic bt-aq5s Tag to-review
      (NICHT completed — PO-Gate)
- [ ] Selbst-Review dokumentiert: Maus bewusst NICHT in E2 (E5-Scope), Windowing
      wiederverwendet (kein Zweitbau), I01/I03/Q01 als geschlossen referenziert
      (Task 2/4)
- [ ] Commit docs: E2-Abschluss
- [ ] ce-nsp-auto Handover-Prompt für E3 (Mutationen) erzeugt
