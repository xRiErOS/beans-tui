---
# bt-a3a8
title: 'Blocking-/Parent-Picker: Suche + Filter-Strip (Type/Status/Priority/Tags/Title)'
status: todo
type: feature
priority: high
created_at: 2026-07-20T07:29:58Z
updated_at: 2026-07-20T07:29:58Z
parent: bt-vy1q
---

**Nebenbefund N7 (PO, 2026-07-20).** Der Blocking-Picker (`r`) listet Kandidaten ohne Suche und ohne Filter. Bei realistischen Repos (sproutling: 114 beans, beans-tui: 123) ist er dadurch **kaum nutzbar** — man scrollt blind. Gleiches Problem sehr wahrscheinlich beim Parent-Picker (`a`).

Der Tag-Picker (`t`) hat bereits ein immer-fokussiertes Suchfeld — dort gibt es den Praezedenzfall inkl. Fallstricken.

## Ziel
Das Picker-Overlay bekommt:
1. **Textsuche** (Titel/ID) — tippen filtert die Kandidatenliste live
2. **Filter-Strip** mit Type · Status · Priority · Tags — optisch derselbe Chip-Look wie die Browse-Filter-Leiste

## Wiederverwenden (nicht neu bauen)
- `internal/tui/box_filter_bar.go` — `filterBar(width)` rendert bereits genau so einen Chip-Strip via `gridRow`/`dropdownBox`
- `internal/tui/box_filter_facets.go` — `beanMatchesFacets` / `beanMatchesSearch` / `beanMatches` sind die vorhandenen Praedikate
- `internal/tui/box_picker_tag.go` — Praezedenzfall fuer ein immer-fokussiertes Suchfeld IM Overlay
- `wideModalWidth(m.width)` — die Pickers dimensionieren sich schon so

## Betroffen
- `internal/tui/box_picker_blocking.go` (Key `r`)
- `internal/tui/box_picker_parent.go` (Key `a`) — gleiche Behandlung, sonst bleibt die Inkonsistenz

## WICHTIG — Design-Fallstricke
- **Picker-lokaler Filter-State**, NICHT der globale Browse-Filter. Der Picker darf `m.filterStatus/Type/Priority/Tag` NIEMALS mutieren, nur die Praedikate mit eigenem State nutzen. Sonst veraendert das Oeffnen eines Pickers die Browse-Ansicht.
- **Tasten-Kollision im Suchfeld** (Lehre aus dem Tag-Picker, bean bt-9ipw/D01): waehrend ein Suchfeld fokussiert ist, muessen normale Buchstaben TIPPBAR bleiben — `x`/`s`/`t` duerfen dort nicht als Toggle/Aktion feuern. Nur `space`/esc/enter bzw. explizit reservierte Tasten steuern.
- **Nicht hinter `BT_BOXFORM` gaten**: das ist ein genereller Usability-Fix, kein Experiment-Look. (Der Chip-Strip stammt zwar aus dem Experiment, die Funktion ist aber allgemein noetig.) Falls das Widget nur unter dem Flag existiert → Abhaengigkeit im Bean vermerken und mit PO klaeren.
- Overlay-Hoehe: Strip + Suchzeile kosten Zeilen — Kandidatenliste entsprechend kuerzen, nicht ueberlaufen lassen.

## Akzeptanz
- [ ] `r` (Blocking): Tippen filtert Kandidaten nach Titel/ID live
- [ ] `r`: Filter-Strip Type/Status/Priority/Tags sichtbar und wirksam
- [ ] `a` (Parent): dieselbe Behandlung
- [ ] Picker-Filter mutiert den globalen Browse-Filter NICHT (Test!)
- [ ] Buchstaben bleiben im Suchfeld tippbar (keine Aktions-Kollision)
- [ ] Overlay laeuft bei 80 Spalten nicht ueber (tmux-Smoke)
- [ ] Tests: Filterung, Auswahl nach Filterung trifft den richtigen Bean, globaler Filter unveraendert
- [ ] Voller `command go test ./...` gruen
