---
# bt-8q9c
title: 'T6 Theme-Port: Catppuccin-Token + beans-Enum-Mapping'
status: in-progress
type: task
priority: high
created_at: 2026-07-14T18:34:04Z
updated_at: 2026-07-14T20:09:02Z
parent: bt-blsy
blocked_by:
    - bt-snkb
---

Plan: implementation-plan.md »E1 Task 6«. Quelle: devd internal/theme/theme.go 1:1, nur 4 dokumentierte Abweichungen (BT_ASCII_ICONS, Status-/Type-/Prio-Mapping für beans-Enums).

## Akzeptanz
- [ ] `internal/theme/{theme,icons}.go` — Token identisch devd (Hex-Werte), Glyph ◉, Type-Icons je beans-Type
- [ ] Tests: StatusColorMapping (5 Status), AsciiFallback, TypeIconAllTypes grün
