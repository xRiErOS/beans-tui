---
# bt-fy5d
title: 'Footer entschlacken: inline gezeigte Keys nicht doppeln'
status: todo
type: task
priority: normal
created_at: 2026-07-20T07:25:11Z
updated_at: 2026-07-20T07:25:11Z
parent: bt-vy1q
---

**Nebenbefund N2 (PO, 2026-07-20).** Im Box-Modus stehen die Feld-Hotkeys bereits salient IM Box-Rahmen (`(e) (s) (o) (u) (a) (t)`). Der Footer wiederholt sie trotzdem (`s Status · e Edit · t Tags · a Parent · r Blocking`). Redundanz — bei 80 Spalten kostet der Footer dadurch 3 Zeilen.

Beleg: `~/Obsidian/Vault/lean-stack/beans-tui/beans-tui-boxform-narrow.gif` (Footer 3-zeilig, alle Keys doppelt).

## Ziel
Wenn `boxFormEnabled()`: die Keys, die das Detail bereits inline als Badge zeigt, aus der Footer-Liste entfernen. Alles andere (tab/shift+tab, `/`, `f`, `c`, `d`, `b`, `y`, `G`) bleibt.

## Betroffen
- `internal/tui/footer_context.go` — die `*LocalBindings()`-Sets (Single Source der Footer-Zeile)
- ggf. `internal/tui/view.go` `renderBindings`
- `internal/tui/box_form_flag.go` — `boxFormEnabled()`

## Akzeptanz
- [ ] Bei `BT_BOXFORM=1` zeigt der Footer NICHT mehr `s Status`, `e Edit`, `t Tags`, `a Parent`, `r Blocking` (die inline sichtbaren)
- [ ] Bei Flag AUS ist der Footer unveraendert (Bestandsgolden byte-identisch!)
- [ ] `browse_boxform.golden` regeneriert, Footer dort kuerzer
- [ ] Test: Footer-Inhalt unter Flag AN vs AUS (assert die entfernten Keys fehlen/erscheinen)
- [ ] Kein Drift-Guard-Bruch (Keys bleiben in `helpGroups()`, nur die FOOTER-Anzeige aendert sich)
- [ ] Voller `command go test ./...` gruen
