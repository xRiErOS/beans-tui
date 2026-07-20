---
# bt-mx4k
title: Palette nur noch K, ctrl+k entfernen
status: completed
type: task
priority: normal
created_at: 2026-07-20T09:30:49Z
updated_at: 2026-07-20T10:13:48Z
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


## Summary

`ctrl+k` entfernt, `K` ist die einzige Palette-Bindung (keymap.go:172).
Header sechs Zeichen kuerzer: `ctrl+k commands` -> `K commands`.
D07 in docs/plans/jira-style-experiment/design-spec.md als PO-bestaetigt vermerkt.

Fuenf Full-Capture-Guards (view_tag_management_test.go x4, overlay_shortcuts_test.go x1)
sondierten mit `ctrl+k`. Ein unbebundener Key beweist nichts — sie sondieren jetzt `K`,
womit sie wieder eine lebende Bindung testen (in den Tipp-Modi landet `K` als
literales Zeichen im Input, die erwarteten Werte sind entsprechend `d?K`).

Golden regeneriert (nur die Header-Zeile, Zeile fuer Zeile per git diff geprueft):
tree.golden, backlog.golden, browse_flat.golden, browse_boxform.golden.

Commit 862d8ee.

## Test-Output

command go test ./... — vollstaendig, ohne -short:

    ?   github.com/xRiErOS/beans-tui  [no test files]
    ok  github.com/xRiErOS/beans-tui/cmd  1.831s
    ?   github.com/xRiErOS/beans-tui/internal/clip  [no test files]
    ok  github.com/xRiErOS/beans-tui/internal/config  1.050s
    ok  github.com/xRiErOS/beans-tui/internal/data  3.907s
    ok  github.com/xRiErOS/beans-tui/internal/theme  0.702s
    ok  github.com/xRiErOS/beans-tui/internal/tui  152.444s

## Deviations

Keine. `chrome_test.go:27` enthaelt weiterhin den Text "ctrl+k:cmd" — das ist ein
handgeschriebener Fixture-String als Eingabe fuer `breadcrumb()`, keine Keymap-Aussage
und keine gerenderte UI; unangetastet gelassen (kein Refactoring ueber den Auftrag hinaus).
