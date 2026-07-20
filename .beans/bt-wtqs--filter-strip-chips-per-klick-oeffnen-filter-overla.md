---
# bt-wtqs
title: 'Filter-Strip: Chips per Klick oeffnen Filter-Overlay'
status: in-progress
type: bug
priority: normal
created_at: 2026-07-20T18:05:08Z
updated_at: 2026-07-20T18:09:25Z
parent: bt-vy1q
---

# Kontext (Epos bt-vy1q — Rest siehe Epos, DRY)

PO-Befund (2026-07-20): Der **Filter-Strip** (Code `filterBar`, `box_filter_bar.go`)
reagiert NICHT auf Maus. Erwartet: Klick auf ein boxed field im Strip (Type/Status/
Priority/Tags) öffnet das Filter-Overlay (Facetten), analog zur Tastatur (`f`).
Nur mit `BT_BOXFORM=1` sichtbar.

## Ursache (Investigator verifiziert, read-only)

Ein Klick-Handler für den Strip **fehlt komplett**. `filterBarHeight` wird überall
nur als übersprungener Offset behandelt, nie als eigene Hit-Region:

| Datei:Zeile | Symbol | Rolle |
|---|---|---|
| `mouse.go:148` | `handleMouse` | globaler MouseMsg-Router |
| `mouse.go:196-201` | `switch m.view` | nur `viewBrowseRepo→mouseTreeClick`, `viewBacklog→mouseBacklogClick` — **kein Strip-Case** |
| `box_filter_bar.go:34` | `const filterBarHeight = 3` | Strip-Höhe (gridRow, 4 Chips) |
| `box_filter_bar.go:41` | `filterBar(width)` | reines Rendering |
| `view_browse_repo.go:1423-1445` | `treeClickRow` | `originY += filterBarHeight`, `bodyH -= filterBarHeight` → Strip als reservierter, nicht-klickbarer Bereich |
| `view_browse_flat.go:148-157` | `flatClickRow` | identische Reservierung |
| `mouse.go:757` | `detailBoxFormClickRow` | `originY += filterBarHeight`, Detail-Pane ebenso verschoben |
| **fehlt** | `filterBarFieldAt` / Pendant zu `boxFormFieldAt` | Geometrie Klick→Chip |

Vorbild (Detail-Pane, funktioniert): `mouse.go:885-897` `activateBoxFormTarget` →
Feld→Aktion. Genau dieses Muster fehlt für den Strip.

## Ziel-Funktion (WICHTIG — richtige Aktion)

Der Strip ist **Filtern**, nicht Bean-Mutation. Tastatur-Trigger ist `f` →
`openFilterMenu()` (`update.go:1614` → `box_filter_facets.go`). Klick MUSS
`openFilterMenu()` auslösen — **NICHT** `openValueMenu()` (das ist die Detail-Pane,
mutiert DAS bean; verwechseln = falsch).

## Grounding-first (VOR Umsetzung, ins bean appenden)

- [ ] Kann `openFilterMenu()` / das Facetten-Overlay auf eine **bestimmte Facette
      geseedet** öffnen (Klick auf „Status"-Chip → Overlay bei Status)? `box_filter_facets.go`
      lesen. Falls ja: Chip-Spalte→Facette mappen. Falls nein: Overlay oben öffnen +
      als Folge-Prelude notieren. Ergebnis als `## Grounding`-Sektion appenden.
- [ ] Spaltenlayout der 4 Chips ermitteln (`gridRow` in `box_filter_bar.go`) — wie
      clickCol → Chip-Index.

## Akzeptanz (abhakbar)

- [ ] `filterBarFieldAt(clickRow, clickCol, width) (facet string, ok bool)` (o.ä.):
      mappt einen Klick in der reservierten Strip-Region auf einen der 4 Chips.
      Region = die obersten `filterBarHeight` Zeilen unter etwaigem Header, nur wenn
      `boxFormEnabled()`.
- [ ] `handleMouse` (`mouse.go`): Strip-Hit VOR dem Pane-`switch` prüfen (Strip liegt
      über beiden Panes, volle Breite) → bei Treffer `openFilterMenu()` (ggf. geseedet),
      Klick nicht an Tree/Detail durchreichen.
- [ ] **RED zuerst:** Update-Test `tea.MouseMsg` (Klick auf Strip-Zeile) → Model-State
      zeigt Filter-Overlay offen. Tabellen-getrieben über die 4 Chip-Spalten, falls
      Seeding unterstützt. Test MUSS vor der Verdrahtung RED sein (zitieren).
- [ ] **Geometrie-Test bei 80 UND 120 Spalten** — Maus-Off-by-one lebt genau an der
      Grenzbreite (Repo-Historie bt-hd42/bt-vpvu: Klick-Offset-Bugs). Beide Breiten pinnen.
- [ ] Klick, der NICHT auf den Strip trifft, verhält sich unverändert (Regression:
      Tree-/Detail-Klick unberührt).
- [ ] Voller Lauf grün `command go test ./...` · Build `command go build -o bin/bt .` · `go vet`.
- [ ] Goldens: nur falls sich statisches Rendering ändert (unwahrscheinlich — Overlay
      ist State). Falls ja: `-update` + Diff-Prüfung.
- [ ] Commit `fix(mouse): Filter-Strip-Chips per Klick öffnen Filter-Overlay` · `Refs: <id>`.

## Empfohlener Smoke

tmux-Smoke bei 80 Spalten gegen sproutling: Strip-Chips anklicken, prüfen dass das
richtige Facetten-Overlay öffnet (Maus-Geometrie hat dieses Repo mehrfach gebissen).
Ehrlich dokumentieren: real gesmoked vs. nur Unit.


## Grounding

**Seeding: JA, unterstützt.** `openFilterMenu()` (update.go:1614) setzt `m.filterItems`/`m.filterMenu`/`m.filterTab=0` — kein eingebauter Seed-Parameter, aber `filterFacetOrder(m.filterItems)` (box_filter_facets.go:269) liefert die Facetten-Reihenfolge (status,type,priority,tag,archive) und `filterFacetRange` (box_filter_facets.go:285) das [start,end)-Zeilenfenster einer Facette. Neue Hilfsfunktion `openFilterMenuAt(facet string)` ruft `openFilterMenu()`, sucht `facet` in `filterFacetOrder`, setzt `m.filterTab` auf den gefundenen Index + `m.filterMenu.cursor` auf `start` -- exakt das Muster, das `filterMenuSwitchTab` (box_filter_facets.go:350) für Tab/Shift+Tab bereits nutzt. Fällt facet nicht gefunden (kann praktisch nicht passieren, alle 4 Chip-Facetten sind immer in buildFilterItems), bleibt Tab 0 (Status) wie bisher.

**Chip-Spaltenlayout** (box_filter_bar.go:49-54, `filterBar`'s `cells` Slice-Reihenfolge): Spalte 0=Type, 1=Status, 2=Priority, 3=Tags. Mapping Spalte→Facet-String (box_filter_facets.go's `ffItem.facet`-Vokabular): 0→"type", 1→"status", 2→"priority", 3→"tag" (NICHT "tags" — buildFilterItems verwendet den Singular "tag" als facet-Key, box_filter_facets.go:69).

Geometrie: `filterBar` rendert bei `innerW` (view_browse_repo.go:1303, volle Pane-Breite inkl. beider Sub-Panes + deren Rahmen, NICHT nur `lw+rw`), beginnend an `clickPaneGeometry`'s ROHEM `originY`/`originX` (VOR der `+= filterBarHeight`-Korrektur, die `treeClickRow`/`detailBoxFormClickRow` erst NACH dem Strip anwenden) -- der Strip besetzt exakt `[originY, originY+filterBarHeight)` x `[originX, originX+innerW)`. Spalten-Bucket via `gridColWidths(4, innerW)` + `gridColAt` (mouse.go:594-612, bereits bestehendes Muster für die Detail-Pane-Boxen).
