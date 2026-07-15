---
# bt-qzwt
title: E3 T7 — E3-Abschluss-Ritual
status: todo
type: task
priority: high
created_at: 2026-07-15T00:28:17Z
updated_at: 2026-07-15T00:28:17Z
parent: bt-gzcu
blocked_by:
    - bt-ppzb
---

Ziel: E3-Abschluss-Ritual (implementation-plan.md »Epos-Rituale« -> Epos-
Abschluss), mirrors E2 Task 6 (bt-enrd).

Plan: docs/plans/v1-port/epic-E3-plan.md »Task 7«.

## Akzeptanz
- [ ] go test ./... -count=1 (2x hintereinander) grün, go build -o bin/bt . ok,
      gofmt -l . leer, go vet ./... leer
- [ ] Manueller Dogfooding-Smoke (tmux): s/t/a/B/c/d/e auf realen Beans im eigenen
      Repo durchgespielt, Terminal-Ausschnitt als Beleg im Commit-Body
- [ ] beans pflegen: Task-beans (T1-T6) auf completed, Epic bt-gzcu Tag
      `to-review` (NICHT -s completed -- PO-Gate, lean-stack-Prinzip "der
      ausführende Agent schließt NICHT")
- [ ] Selbst-Review im Commit-Body (Spec-Coverage V7/V8, Konsolidierung ggü. devd,
      bewusste Scope-Cuts: kein Toast/E5, kein Type-Hierarchie-Client-Check bei
      Create, kein Blocking-Zyklen-Ausschluss)
- [ ] Commit `docs: E3-Abschluss` (Refs: bt-gzcu-Task-ID)
- [ ] Skill `ce-nsp-auto` -> Handover-Prompt für E4 (Command-Center & Review-
      Cockpit, bean-Suche via `beans list --json`)
