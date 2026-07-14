---
# bt-8q9c
title: 'T6 Theme-Port: Catppuccin-Token + beans-Enum-Mapping'
status: completed
type: task
priority: high
created_at: 2026-07-14T18:34:04Z
updated_at: 2026-07-14T20:14:30Z
parent: bt-blsy
blocked_by:
    - bt-snkb
---

Plan: implementation-plan.md »E1 Task 6«. Quelle: devd internal/theme/theme.go 1:1, nur 4 dokumentierte Abweichungen (BT_ASCII_ICONS, Status-/Type-/Prio-Mapping für beans-Enums).

## Akzeptanz
- [x] `internal/theme/{theme,icons}.go` — Token identisch devd (Hex-Werte), Glyph ◉, Type-Icons je beans-Type
- [x] Tests: StatusColorMapping (5 Status), AsciiFallback, TypeIconAllTypes grün

## Summary of Changes

- `internal/theme/theme.go`: Catppuccin-Macchiato-Palette 1:1 aus devd portiert (alle Hex-Token), Style-Konstruktoren (Header/Key/Accent/Dim/Muted/Chevron), `SetAccent`, beans-Status-Mapping (draft/todo/in-progress/completed/scrapped → Blue/Text/Yellow/Green/Red), gemeinsamer Status-Glyph `◉`/ASCII `*`, beans-Priority-Mapping (critical/high/normal/low/deferred).
- `internal/theme/icons.go`: beans-Type-Icon-Mapping (milestone ⬢ Peach, epic ✦ Mauve, feature ✦ Green, task ⯅ Blue, bug ⯁ Red) + ASCII-Fallbacks (# * + ^ !).
- Env-Var `DEVD_ASCII_ICONS` → `BT_ASCII_ICONS` (Port-Adaptation 1).
- Unbekannte Enum-Werte (Status/Type): neutral Text-Farbe + Fallback-Glyph `·`/ASCII `.` (dokumentierte Design-Entscheidung, siehe Kommentar in theme.go).
- `internal/theme/theme_test.go`: TestStatusColorMapping, TestAsciiFallback, TestTypeIconAllTypes, TestPriorityColorMapping — alle grün.
- go.mod/go.sum: `github.com/charmbracelet/lipgloss@v1.1.1-0.20250404203927-76690c660834` + `github.com/muesli/termenv@v0.16.0` (Pin identisch devd; termenv aktuell ungenutzt, vorgezogen für Task 7 Chrome-Primitiven).
- `gofmt`/`go vet`/`go build`/`go test ./...` clean.
