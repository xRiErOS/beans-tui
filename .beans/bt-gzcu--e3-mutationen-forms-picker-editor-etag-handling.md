---
# bt-gzcu
title: E3 Mutationen — Forms, Picker, Editor, ETag-Handling
status: todo
type: epic
priority: high
created_at: 2026-07-14T18:33:30Z
updated_at: 2026-07-14T18:33:30Z
parent: bt-apmy
blocked_by:
    - bt-aq5s
---

Ziel: Create-/Edit-Forms (huh) mit Confirm-Gate, Status-/Type-/Prio-Menüs (nur beans-Enums), Tag-/Parent-/Blocking-Picker (Zyklen-Ausschluss), Delete-Confirm mit Kinder-Count, `$EDITOR`-Integration, ErrConflict→Toast+Reload.

Epos-Start-Ritual wie E2 (epic-E3-plan.md). Port-Quellen: forms_shared.go, form_create_*.go, box_confirm_*.go, editor.go (devd cli-go). Tag-Regex: ^[a-z][a-z0-9]*(-[a-z0-9]+)*$.
