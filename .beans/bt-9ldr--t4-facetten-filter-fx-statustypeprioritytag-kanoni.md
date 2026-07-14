---
# bt-9ldr
title: 'T4 Facetten-Filter f/X: Status/Type/Priority/Tag + kanonischer Orphan-Sort (I03-Abschluss)'
status: completed
type: task
priority: high
created_at: 2026-07-14T21:57:28Z
updated_at: 2026-07-14T23:35:10Z
parent: bt-aq5s
blocked_by:
    - bt-4ep2
---

Ziel: Facetten-Filter `f` (Status/Type/Priority/Tag, EIN geteilter Zustand für Tree
UND Backlog) + `X` Reset. Schliesst I03 ab (Orphan-Bucket nutzt jetzt data.SortBeans
statt eigener sortByTitleThenID-Logik — Single-Source-Prinzip, das index.go für
sortBeans bereits für sich beansprucht).

Plan: docs/plans/v1-port/epic-E2-plan.md »Task 4«.

## Akzeptanz
- [x] I03 Abschluss: sortByTitleThenID entfernt, collectOrphans/collectCycleOrphans
      nutzen data.SortBeans; bestehende Orphan-/Zyklus-Tests ggf. auf kanonische
      Reihenfolge angepasst
- [x] box_filter_facets.go: ffItem + buildFilterItems (Status/Type/Priority/Tag —
      kein "Art"-Facet wie devd, beans-Type deckt das ab), Checkbox-Menü
      (Port-Muster view_browse_backlog.go:89-136 / view_browse_project.go:998-1033)
- [x] Toggle-Keybinding (space/x) in keymap.go ergänzt, helpGroups-Drift-Test bleibt
      grün
- [x] Facet-Maps nutzen die I01-Copy-on-Write-Konvention aus T2 (keine neue Ausnahme)
- [x] Filter wirkt kombiniert mit Suche (AND) auf Tree; X leert alle Facetten;
      go test ./... grün


## Übernommene Findings aus E2-T3-Review (PFLICHT in diesem Task)
- [x] I01: beanMatchesSearch — lokale ID-Substring-Treffer mit Bleve-Ergebnis UNIONen statt Replace (sonst Flicker: ID-Treffer verschwindet, sobald async Bleve-Antwort landet). Test dazu.
- [x] I02 (optional): Render-Test für leeren Filter-Treffer (View() mit 0 Matches)
- [x] Q01 (Notiz): Bleve ohne Generation-Counter — akzeptiertes Restrisiko, nur dokumentieren falls berührt. Nicht berührt in diesem Task (kein Generation-Counter-Code angefasst) — bleibt reine Notiz.

## Summary

I03 geschlossen: `sortByTitleThenID` entfernt, `collectOrphans`/`collectCycleOrphans`
nutzen jetzt `data.SortBeans`. Neuer Fund dabei: `data.SortBeans` allein reicht NICHT
für Determinismus bei Voll-Ties (gleicher Status/Priority/Type/Title) — der
Stable-Sort erhält dann die Reihenfolge des VORHERIGEN (Map-)Walks, die selbst
nichtdeterministisch ist. Fix in `tagFilterOptions` (dieselbe Klasse Problem beim
dynamischen Tag-Facet, s.u.): erst nach ID vorsortieren, dann `data.SortBeans`
drüber — Test `TestTagFilterOptionsDeterministicAcrossCalls` deckt das ab.

`box_filter_facets.go` (neu): `ffItem`, `buildFilterItems` (5 Status + 5 Type + 5
Priority fest, Tag dynamisch aus den geladenen Beans), `beanMatchesFacets`,
`beanMatches` (AND von Task 3s `beanMatchesSearch` + Facetten), `treeActive`
(Suche ODER Facetten aktiv), `toggleFacet`/`clearFacets` (I01 copy-on-write),
`keyFilterMenu` (up/down/space·x toggle/X clear-ohne-schließen/esc·enter·f
schließen-ohne-clear), `treeFilterBox` (schwebendes Checkbox-Menü, Port devd
view_browse_project.go:998-1033), `filterSummary`/`joinFilterKeys` (Kopfzeile),
`clampModalWidth` (Port devd forms_shared.go, erster Konsument hier).

Zwei Plan-Erratums gefunden und korrigiert (nicht wörtlich portiert):
1. Plan-Skizze für `buildFilterItems` behauptete "Tag-Optionen dynamisch aus
   idx.ByID, stabile Insertion-Order" — idx.ByID ist eine Go-Map ohne definierte
   Iterationsreihenfolge, ein direkter Walk wäre nichtdeterministisch. Fix:
   Beans zuerst nach ID vorsortieren, dann data.SortBeans (s.o.).
2. devds eigenes `joinFilterKeys` (view_browse_project.go:1062-1071) iteriert die
   Facet-Map ohne Sortierung — für die Kopfzeilen-Anzeige hier mit `sort.Strings`
   korrigiert (deterministische Anzeige-Reihenfolge).

`keyTree` (update.go): `f` öffnet das Menü (baut `filterItems` frisch,
`filterMenu.setLen`), `X` wirkt auch bei geschlossenem Menü als direkter Top-Level-
Reset (derselbe `clearFacets`-Helper wie im Menü). Die esc-Kaskade (Rung 2) wurde
von "nur Suche löschen" auf "Suche UND alle vier Facetten löschen" verallgemeinert
(`m.treeActive()` statt `m.treeSearchActive()`) — devd-Parität
(view_browse_project.go:725-736). Das schwebende Filter-Menü fängt Input
vollständig ab (`m.filterOpen`-Check in `handleKey`, direkt nach `m.searchActive`,
gleiches Präzedenzmuster).

`treeSearchLine` (view_browse_repo.go) erweitert: bei aktiver Suche UND/ODER
aktiven Facetten erscheinen Query+Filter-Summary gemeinsam in Rot in der
Tree-Kopfzeile. Der IDLE-Zustand (kein Search/Filter aktiv) blieb absichtlich
BYTE-IDENTISCH zu Task 3 — `tree.golden` brauchte deshalb KEIN `-update`
(2x ohne `-update` grün, Determinismus bestätigt).

I01 (E2-T3-Review-Fund, PFLICHT hier): `beanMatchesSearch` unioned jetzt den
lokalen ID-Substring-Treffer mit dem autoritativen Bleve-Ergebnis statt es zu
ersetzen — sonst verschwindet ein reiner ID-Treffer im Moment, in dem die
Async-Bleve-Antwort landet (Bleve indiziert Title+Body, nicht zwingend
ID-Substrings). Test: `TestSearchIDSubstringHitStaysVisibleAfterBleveResultArrives`
(search_test.go).

I02 (optional): `TestEmptyFacetMatchViewRendersWithoutPanic` deckt View() mit
0 Treffern ab (implizite Panic-Prüfung über den Testlauf selbst).

Q01: nicht berührt — keine Bleve-Generation-Counter-Änderung in diesem Task.

Tests: 6 neue Dateien-Änderungen + 2 neue Dateien, 96/96 Tests grün in
`internal/tui` (voller Lauf 3x `-count=1` + 1x `-race`, keine Flakes), `go test
./...`/`gofmt -l .`/`go vet ./...` sauber, Golden 2x ohne `-update` grün.
tmux-Smoke: `f` → Type-Facette "Epic" togglen → Tree-Kopf zeigt `Ty:epic`, Baum
zeigt nach Expand nur noch die 6 Epic-Beans unter dem Milestone-Root (Tasks
ausgeblendet); `X` setzt zurück auf `/ search`-Idle-Zustand.

Deviation vom Plan-Text (kein Bug, bewusste Life-Entscheidung): Facet-Anzeige-
Reihenfolge (Status/Type/Priority) folgt der kanonischen Tier-Order aus
data/index.go (in-progress→todo→draft→completed→scrapped etc.) statt einer
beliebigen Reihenfolge — konsistent mit dem Rest der App, nicht im Plan-Text
spezifiziert, aber naheliegend und low-risk.

Notiz für Task 5 (Backlog, bt-gzu6): `beanMatches`/`treeActive`/`filterStatus`
/`filterType`/`filterPriority`/`filterTag`/`filterOpen`/`filterItems`/
`filterMenu`/`buildFilterItems`/`treeFilterBox`/`keyFilterMenu` sind alle bereits
view-agnostisch geschrieben (kein `viewBrowseRepo`-Bezug im Code) — Task 5 kann
sie unverändert wiederverwenden, exakt wie `focusedBean()` es für die
Detail-Accordion-Verdrahtung in Task 2 vorgemacht hat. Backlog muss nur seinen
eigenen `visibleNodes()`-Äquivalent (`backlogVisible()`) durch `m.beanMatches`
filtern, keine neue Facet-Logik nötig.
