---
# bt-z4w7
title: 'Value-Menue: Schliessen-Alias und Footer-Label an die Gruppe binden (B7)'
status: todo
type: bug
priority: low
created_at: 2026-07-20T07:26:50Z
updated_at: 2026-07-20T07:26:50Z
parent: bt-vy1q
---

**B7, gefunden in S4 (2026-07-20).** Das Value-Menue kann fuer die Gruppen `status`/`type`/`priority` geoeffnet werden (`s`/`o`/`u`). Aber:
- `keyValueMenu`s Schliessen-Zweig (`internal/tui/box_menu_value.go` ~Z.165) matcht nur `keys.Back` und `keys.Status` → ein per `o`/`u` geoeffnetes Menue schliesst auch mit `s`
- Der Footer-Hint (`box_menu_value.go` ~Z.243, `"esc/s:cancel"`) zeigt immer `s`, egal welche Gruppe offen ist

`esc` funktioniert in allen Faellen ⇒ nichts ist kaputt, nur Label-/Verhaltens-Mismatch.

## Hinweis
Die urspruengliche Design-Entscheidung a3 formulierte woertlich "esc/s schliesst" — eine Aenderung sollte das bewusst revidieren (als neue Sektion dokumentieren, nicht stillschweigend).

## Akzeptanz
- [ ] Schliessen-Match akzeptiert die Taste der jeweils offenen Gruppe (s/o/u) plus esc
- [ ] Footer zeigt die Taste der offenen Gruppe
- [ ] Entscheidung a3 als revidiert dokumentiert
- [ ] Tests fuer alle drei Gruppen, voller `command go test ./...` gruen
