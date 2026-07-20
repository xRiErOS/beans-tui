---
# bt-mx4k
title: Palette nur noch K, ctrl+k entfernen
status: todo
type: task
priority: normal
created_at: 2026-07-20T09:30:49Z
updated_at: 2026-07-20T09:30:49Z
parent: bt-vy1q
---

**PO-Entscheidung 2026-07-20.**

Die Command-Palette hoert heute auf ZWEI Tasten:
```go
Palette: keybind.NewBinding(keybind.WithKeys("ctrl+k", "K"), keybind.WithHelp("ctrl+k", "commands")),
```
`K` ist also laengst gebunden und entspricht D07 (gross = View/Global). Nur das Hilfe-Label
zeigt weiterhin `ctrl+k`.

**Entscheidung: `ctrl+k` entfaellt, `K` bleibt als einzige Bindung.** Eine Taste, eine
Funktion. Nebeneffekt: der Header wird sechs Zeichen kuerzer, was bei 80 Spalten zaehlt.

## Betroffen
- `internal/tui/keymap.go:172` — `WithKeys("K")`, `WithHelp("K", "commands")`
- `internal/tui/view_browse_repo_test.go:34` und `view_browse_backlog_test.go:511`
  erwarten woertlich `"ctrl+k commands"` -> auf `"K commands"` anpassen
- Golden mit Header (`tree`, `backlog`, `browse_flat`, `browse_boxform`, `chrome`)
  regenerieren
- `docs/plans/jira-style-experiment/design-spec.md` — D07 als bestaetigt vermerken

## Akzeptanz
- [ ] `K` oeffnet die Palette, `ctrl+k` nicht mehr
- [ ] Header zeigt `K commands`
- [ ] Golden regeneriert und Zeile fuer Zeile geprueft
- [ ] voller Testlauf gruen
