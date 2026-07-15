---
# bt-gzcu
title: E3 Mutationen — Forms, Picker, Editor, ETag-Handling
status: in-progress
type: epic
priority: high
created_at: 2026-07-14T18:33:30Z
updated_at: 2026-07-15T01:32:31Z
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
