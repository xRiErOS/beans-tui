---
# bt-2p9m
title: Filter-Menü als Querformat mit Tab-Kategorien
status: todo
type: feature
priority: normal
created_at: 2026-07-17T06:21:42Z
updated_at: 2026-07-17T06:46:47Z
parent: bt-5uzr
---

NB aus PO-Review E11 (2026-07-17, Runde 2), PO verbatim:

"Das Filter-Menü ist sehr umfangreich. Wenn ich als Nutzer zu einem konkreten
Filter gehen möchte, dann muss ich sehr lange mit navigieren. Besser wäre, wenn
das Filter-Menü querformat wird und die Filterkategorien in Tabs dargestellt
werden. Klickpfad: 1) f für Filter, erster tab aktiviert 2) mit tab/shift-tab
andere Filter wählen 3) im fokussierten Tab ist das erste Element immer aktiv
und ich kann mit Pfeil-rauf/runter direkt die Auswahl navigieren 4) mit
leertaste filter-kriterium wählen, 5) mit enter den Filter anwenden."

Interpretation: Facetten-Overlay (f, box_filter_facets.go) von vertikaler
Gesamtliste auf horizontales Tab-Layout umbauen — eine Filterkategorie
(Status/Type/Priority/Tags/Archive) je Tab. Interaktionsmodell komplett vom PO
spezifiziert (5 Schritte oben). Akzeptanz-Entwurf:
- [ ] f öffnet Overlay im Querformat, erster Tab aktiv
- [ ] tab/shift-tab wechseln Kategorie-Tabs
- [ ] Fokus-Tab: erstes Element aktiv, Pfeil rauf/runter navigiert direkt
- [ ] Leertaste toggelt Kriterium, enter wendet an
- [ ] bestehende Filter-Semantik (Facetten-Logik) unverändert

Planner verfeinert vor Umsetzung (Breiten-Verhalten 80 Spalten beachten,
NBSP-Wordwrap-Falle, LL Eintrag 4).


## Plan-Konkretisierung E12 (2026-07-17)

Plan: `docs/plans/v1-port/epic-E12-plan.md` §„Item 7: Filter-Menü als
Querformat mit Tab-Kategorien". Reihenfolge: Rang 7 (nach `bt-2kfl` —
isolierteste, größte Einzelaufgabe dieser Runde, bewusst zuletzt).

**Ist-Stand:** `treeFilterBox` (`box_filter_facets.go:296-318`) rendert
heute EINE vertikale Liste über alle fünf Facetten (`Status`/`Type`/
`Priority`/`Tags`/`Archive`, `facetHead`-Reihenfolge, Zeile 34-40) mit EINEM
`m.filterMenu`-Cursor über die volle geflachte `m.filterItems`-Liste
(`keyFilterMenu`, Zeile 268-292).

**Kein Tastenkonflikt (verifiziert):** `tab`/`shift+tab` sind global an
`keys.FocusIn`/`keys.FocusOut` gebunden (`keymap.go:123-124`), aber
`handleKey` prüft `m.filterOpen` (Full-Capture, `update.go:946`) VOR der
`FocusIn`/`FocusOut`-Prüfung (Zeile 1061/1073) — `keyFilterMenu` darf
`tab`/`shift+tab` gefahrlos für den Kategoriewechsel belegen, mirrort
`m.searchActive`/`m.tagInputActive`s eigenen Full-Capture-Vorrang.

**Vorgehen:** `buildFilterItems` liefert bereits fünf klar abgegrenzte
Facet-Gruppen (Zeile 57-81) — `m.filterItems` nach `facet`-Feld in
Tab-Reihen gruppieren (kein neuer Datenzustand, nur Rendering + Navigation
umgestellt). Neuer/erweiterter State: Tab-Cursor (aktive Facet-Gruppe)
getrennt vom bestehenden `m.filterMenu`-Zeilencursor (läuft dann NUR
innerhalb der aktiven Gruppe). PO-Klickpfad wörtlich: `f` öffnet mit erstem
Tab aktiv (erstes Element vorselektiert) → `tab`/`shift+tab` wechselt Tab
(Fokus springt auf dessen erstes Element) → `↑`/`↓` navigiert innerhalb des
Tabs → `space` togglet (bestehendes `toggleFacet`, UNVERÄNDERT) → `enter`
wendet an/schließt (UNVERÄNDERT). Breiten-Verhalten bei 80 Spalten beachten
(NBSP-Wordwrap-Falle, `docs/LESSONS-LEARNED.md` Eintrag 4) —
`clampModalWidth`/`wideModalWidth` (Zeile 369-399) als bestehende Bausteine
prüfen statt neu erfinden.

**Akzeptanz:**
- [ ] `f` öffnet Overlay im Querformat, erster Tab aktiv, erstes Element
      vorselektiert
- [ ] `tab`/`shift+tab` wechseln Kategorie-Tabs, kein Konflikt mit globalem
      FocusIn/FocusOut (Test bestätigt Full-Capture-Vorrang)
- [ ] Fokus-Tab: ↑/↓ navigiert NUR innerhalb der aktiven Kategorie
- [ ] space/x togglet Kriterium (Semantik unverändert), enter wendet an
- [ ] bestehende Filter-Semantik (`beanMatchesFacets`, `filterSummary`)
      unverändert — nur Overlay-Rendering/Navigation betroffen
- [ ] tmux-Smoke bei 80 Spalten (Grenzbreite, NBSP-Falle)
- [ ] Golden-Gegenbeleg falls Overlay-Breite/-Höhe golden-relevant wird
