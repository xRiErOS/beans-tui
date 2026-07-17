---
# bt-2p9m
title: Filter-Menü als Querformat mit Tab-Kategorien
status: completed
type: feature
priority: normal
created_at: 2026-07-17T06:21:42Z
updated_at: 2026-07-17T08:32:43Z
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

## Summary (2026-07-17)

Facet-Filter-Menü (`f`) von einer langen vertikalen Gesamtliste auf ein
Querformat mit Tab-Kategorien umgebaut (Status/Type/Priority/Tags/Archive,
`facetHead`-Reihenfolge). PO-Klickpfad (Plan-Konkretisierung E12 Item 7)
1:1 umgesetzt:

- `f` öffnet mit Tab 0 (Status) aktiv, erstes Element vorselektiert
  (`openFilterMenu` setzt `m.filterTab=0`, `listState{}` liefert Cursor 0).
- `tab`/`shift+tab` (`keys.FocusIn`/`FocusOut`, wiederverwendet) wechseln
  die Kategorie, Cursor springt auf deren erstes Element
  (`filterMenuSwitchTab`), wrapt an beiden Enden.
- `↑`/`↓` navigieren NUR innerhalb der aktiven Kategorie
  (`filterMenuMoveCursor`, geclampt auf `[start,end)` der aktiven
  Facet-Range).
- `space`/`x` (`toggleFacet`) und `enter`/`esc`/`f` (Apply/Close)
  UNVERÄNDERT — reine Rendering-/Navigations-Umstellung, keine
  Facet-Semantik-Änderung (`beanMatchesFacets`/`filterSummary` unangetastet).

Neuer State: `model.filterTab int` (types.go) — getrennt von
`m.filterMenu.cursor` (bleibt GLOBALER Index in `m.filterItems`, unverändert
für `toggleFacet`). Neue reine Helfer `filterFacetOrder`/`filterFacetRange`
(box_filter_facets.go) gruppieren `m.filterItems` nach `facet`-Feld, ohne
neuen Datenzustand — Facet mit 0 Zeilen (z. B. keine Tags geladen) erzeugt
KEINEN Phantom-Tab.

## Test-Output

RED (vor Implementierung, Compile-Fehler durch fehlende Helper):
```
vet: internal/tui/box_filter_facets_test.go:268:14: undefined: filterFacetRange
```

GREEN (nach Implementierung, gezielter Lauf):
```
--- PASS: TestFilterMenuOpensWithFirstTabActiveAndFirstElementSelected (0.00s)
--- PASS: TestFilterMenuTabSwitchesCategoryAndSelectsFirstElement (0.00s)
--- PASS: TestFilterMenuArrowsStayWithinActiveTabBounds (0.00s)
--- PASS: TestFilterMenuTabDoesNotLeakToGlobalFocusToggle (0.00s)
--- PASS: TestFilterMenuSpaceTogglesCorrectFacetAfterTabSwitch (0.00s)
--- PASS: TestFilterMenuEnterAppliesUnchangedAcrossTabs (0.00s)
--- PASS: TestTreeFilterBoxShowsTabBarAndOnlyActiveFacetRows (0.00s)
```
(alle bestehenden `TestFilterMenu*`/`TestBuildFilterItems*`/
`TestContextualLocalHint*`/`TestHelpGroupsCoverEveryBindingExactlyOnce`
weiterhin grün, keine Regression.)

Voller Lauf vor Commit: `command go test ./... -count=1` — **PASS**,
`beans-tui/internal/tui 154.8s` (~155s, Baseline ~150s eingehalten).
`command go vet ./...` clean, `gofmt -l` clean (nach Auto-Format von
`types.go`). Golden-Gegenbeleg: `TestChromeGolden`/`TestTreeGolden`/
`TestBacklogGolden` unverändert PASS — Filter-Overlay ist nicht Teil der
3 Basis-Goldens, keine Regen nötig.

## Smoke

**Normalbreite (120×40):** kompletter PO-Klickpfad live durchgespielt —
`f` öffnet Querformat mit `[Status]` aktiv/erstes Element vorselektiert →
`tab` wechselt zu `[Type]`, erstes Element (`milestone`) selektiert →
`↓↓` navigiert innerhalb Type bis `bug` → `space` togglet `[x] bug` →
5× weiteres `↓` clampt bei `task` (letzte Type-Zeile, kein Übertritt in
Priority) → `enter` wendet an + schließt, Tree-Kopf zeigt `Ty:bug`,
Baum korrekt gefiltert. Filter danach mit `f`→`X`→`esc` geleert.

**80 Spalten (Grenzbreite):** derselbe Klickpfad wiederholt — Tab-Leiste
+ beide Hint-Zeilen rendern sauber ohne Wortumbruch. `tab` → Type,
`↓`+`space` togglet `epic`. `shift+tab` zweimal: Type→Status→**Archive**
(Wrap-Around über beide Enden bestätigt, `[Archive]` aktiv). `enter`
wendet an, Tree-Kopf zeigt `Ty:epic`. Filter danach geleert, Session
sauber beendet (`tmux kill-session`).

`.beans/` nach beiden Smokes clean (`git status --short .beans/` leer).

## Deviations/ERRATA

**Regression live im tmux-Smoke gefunden, NICHT von Unit-Tests erfasst**
(gleiche Fehlerklasse wie LESSONS-LEARNED Eintrag 4, aber ein ANDERER
Wrap-Pfad): die erste Fassung der Hint-Zeile war EIN langer String
("space/x:toggle  tab/shift+tab:category  X:clear  enter/esc/f:done",
65 Zeichen) und verließ sich auf `modalBox`s lipgloss-`Width`-Auto-Wrap.
Unter `go test` (kein TTY, NoColor-Profil, keine ANSI-Bytes) wrapte das
sauber an Wortgrenzen — im echten tmux (TrueColor-Profil, `rebaseBg`
schreibt Reset-Codes um) hat derselbe Wrap "X:clear" MITTEN im Wort
gesplittet (`X:cl` / `ear`), obwohl der Token gar kein internes
Leerzeichen enthält — NBSP hätte hier nichts geschützt (keine Leerstelle
zum Schützen vorhanden). Fix: die Hint-Zeile wird jetzt an ZWEI fest
gesetzten Stellen selbst umgebrochen (`"\n"` hart im Code), beide Zeilen
(23/40 Zeichen) bleiben komfortabel unter dem 44-Zeichen-Content-Budget
bei Breite 46 — kein Auto-Wrap mehr nötig, Bug dadurch strukturell
ausgeschlossen statt nur kaschiert. Verifiziert per Vergleichs-Probe
(Direktaufruf `treeFilterBox()` ohne TTY vs. echter tmux-Capture) und im
finalen 80-Spalten-Smoke oben — sauber auf beiden Breiten.

Kein weiteres Abweichen vom Plan (epic-E12-plan.md Item 7) — alle
Akzeptanz-Punkte erfüllt, Facet-Semantik unverändert, Golden-Regen nicht
nötig.

## Notes

Breite `treeFilterBox`-Modal 40→46 (`clampModalWidth(46, m.width)`,
Floor 24 unverändert) — Querformat-Tab-Leiste braucht mehr Platz als die
alte reine Checkbox-Spalte, aber der Boxinhalt ist jetzt viel KÜRZER
(max. 5 Zeilen statt der alten ~27-zeiligen Gesamtliste), erfüllt die
PO-Absicht "Querformat" (breiter, kürzer) strukturell.
