---
# bt-gzcu
title: E3 Mutationen — Forms, Picker, Editor, ETag-Handling
status: in-progress
type: epic
priority: high
tags:
    - to-review
created_at: 2026-07-14T18:33:30Z
updated_at: 2026-07-15T19:20:56Z
parent: bt-apmy
blocked_by:
    - bt-aq5s
---

Ziel: Create-/Edit-Forms (huh) mit Confirm-Gate, Status-/Type-/Prio-Menüs (nur beans-Enums), Tag-/Parent-/Blocking-Picker (Zyklen-Ausschluss), Delete-Confirm mit Kinder-Count, `$EDITOR`-Integration, ErrConflict→Toast+Reload.

Epos-Start-Ritual wie E2 (epic-E3-plan.md). Port-Quellen: forms_shared.go, form_create_*.go, box_confirm_*.go, editor.go (devd cli-go). Tag-Regex: ^[a-z][a-z0-9]*(-[a-z0-9]+)*$.


## Übernahme aus E2-Abschluss (PFLICHT bei E3-Start)
- [ ] B01: keys.FilterClear-Case in keyBacklog ergänzen (X wirkt derzeit nicht in Backlog-View) — Fix analog keyTree, Test dazu


## Erkenntnisse aus T2 (für T3-T6 + E5)
- B01 (Upstream beans 0.4.2): frisch per 'beans create' angelegte Beans melden via list/show ein ETag, das vom update-internen --if-match-Check abweicht, bis das erste erfolgreiche Update die Datei neu schrieb. TUI-Smokes in Scratch-Repos: IMMER erst Warm-up-Update via rohem CLI fahren. Nicht bt-fixbar (Upstream). Kandidat für Upstream-Issue nach v1.
- B02: Konflikt-Statuszeile wird vom Folge-Reload nach 1 Frame überschrieben (Flash) — bei E5-Toast-Einführung klären: Flash ok oder persistieren bis nächste User-Aktion.


## E3-Abschluss (Agent-Meldung, 2026-07-15)
T1-T7 completed, je Task Spec+Quality-Review, 2 Fix-Runden (T4/T5-Findings: Draft-Verlust bei CLI-Reject, Async-Gap-Clobbering, ETag-Freeze für Editor-Sessions). 3 empirische ERRATA gegen Plan/Annahmen: (1) Kinder werden bei Delete zu Roots (nicht verwaist), (2) Delete räumt fremde blocking/blocked_by-Referenzen still ab — beide jetzt im Confirm-Text gewarnt + regressionsgetestet, (3) Upstream-ETag-Drift bei frischen Creates (Warm-up-Konvention dokumentiert). End-to-End-Smoke s→t→a→B→c→e→ctrl+e→d belegt. Suite ~127s voll / ~5s -short.

### PO-Hinweise
| Code | Beschreibung | Empfehlung | Status |
|------|--------------|------------|--------|
| I01 | Konflikt-Statuszeile nach ErrConflict nur 1 Frame sichtbar (Reload überschreibt) | E5-Toast löst das — Entscheid dort: Flash ok vs. persistent | 🟢 gelöst (Toast sticky, TestConflictToastIsStickyAndSurvivesReload PASS, E6 T2) |
| I02 | Upstream-ETag-Drift bei frischen Creates | Upstream-Issue bei hmans/beans nach v1-Abnahme | 🟡 |
