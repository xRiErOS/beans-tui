---
# bt-a3a8
title: 'Blocking-/Parent-Picker: Suche + Filter-Strip (Type/Status/Priority/Tags/Title)'
status: todo
type: feature
priority: high
created_at: 2026-07-20T07:29:58Z
updated_at: 2026-07-20T07:48:39Z
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


## Grounding 2026-07-20 (Investigator, read-only)

**Kernbefund: das Suchmuster existiert bereits im Repo.** Der Tag-Picker hat Suche +
Live-Filter; Blocking-/Parent-Picker haben sie nicht. Diese Aufgabe ist ein **Port eines
erprobten Musters**, kein Neubau.

### Vorlage (hier abschauen)
| Baustein | Ort |
|---|---|
| Sucheingabe | `box_picker_tag.go:183` — standalone `textinput.Model` (`tagInput`) |
| Live-Filter | `box_picker_tag.go:229` — `filterTagItems(items, query)`, case-insensitive substring |
| Tests dazu | `box_picker_tag_test.go` (24 Testfunktionen, u.a. textinput-Integration) |

### Zu aendernde Picker
| Picker | Datei | Konstruktor | Render | Key-Handler | State |
|---|---|---|---|---|---|
| Blocking (`r`) | `box_picker_blocking.go` | `openBlockingPicker()` :60 | `blockingPickerBox()` :198 | `keyBlockingPicker()` :90 | `types.go:379` — `blockItems`/`blockOriginal`/`blockPending` |
| Parent (`a`) | `box_picker_parent.go` | `openParentPicker()` :80 | `parentPickerBox()` :180 | `keyParentPicker()` :104 | `types.go:368` — `parentItems` |

Kandidaten-Quellen: Blocking `buildBlockingItems()` :33 (alle beans ausser self, via `m.idx.ByID`),
Parent `buildParentItems()` :64 (via `data.EligibleParents(idx, b)` — Zyklus-Ausschluss steckt
schon im data-Layer, nicht neu bauen).

Key-Routing: `update.go:794` (`keys.Blocking`) / `:791` (`keys.Assign`);
Overlay-Cases `update.go:857` / `:855`. Bindings `keymap.go:67` / `:66`.

### Filter-Strip: direkt wiederverwendbar
`scalarCell` (`box_detail_form.go:29`), `gridRow(cells, width)` (:65), `gridColWidths(n,width)` (:40),
`dropdownBox(label,value,hotkey,w,focused)` — alle width-aware und bereits durch
`box_detail_form` + `box_filter_bar` getestet. Der Strip (Type/Status/Priority/Tags/Titel) baut
darauf auf, **nichts Neues erfinden**.

Nicht wiederverwendbar: das Facetten-Overlay hinter `f` (`box_filter_facets.go:57`) ist auf seine
4 festen Facetten verdrahtet; die Command-Palette (`overlay_palette.go:222`) filtert selbst nicht.

### Betroffene Tests
`box_picker_blocking_test.go` (16 Funktionen) · `box_picker_parent_test.go` (14) ·
`box_picker_tag_test.go` (Vorlage) · indirekt `overlay_palette_test.go`, `tree_golden_test.go`,
`view_browse_repo_test.go`. Eigene Picker-Golden existieren nicht — Picker rendern in
`tree.golden`/`backlog.golden`.

### Kollisions-Hinweis fuer den Dispatch
Beruehrt `types.go` + `update.go` — dieselben Dateien wie bt-ze10 (Detail-Scroll).
**Nicht gleichzeitig im selben Working Tree bearbeiten.**
