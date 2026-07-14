---
# bt-gzu6
title: 'T5 Backlog-View V3 (b): Master-Detail, geteilter Filter/Suche, Sort-Toggle S'
status: completed
type: task
priority: high
created_at: 2026-07-14T21:57:33Z
updated_at: 2026-07-14T23:54:17Z
parent: bt-aq5s
blocked_by:
    - bt-ms0k
    - bt-9ldr
---

Ziel: Backlog-View V3 (`b`) — parentlose+ready beans (idx.Backlog()), Master-Detail
(Wiederverwendung Task-1-Accordion via focusedBean()-Dispatcher aus T2), geteilter
Such-/Filter-Zustand (T3/T4), Sort-Toggle `S` (status/prio/created/updated,
client-seitig, kein beans-CLI-Resort nötig). Windowing wird wiederverwendet
(windowAround, E1) — kein Neubau.

Plan: docs/plans/v1-port/epic-E2-plan.md »Task 5«.

## Akzeptanz
- [x] view_browse_backlog.go: viewBacklog + Render (Master-Detail, D03
      Border-Fokus-Tausch), reuse renderPane/masterDetailWidths/windowAround
- [x] backlogVisible(): idx.Backlog() + geteilte Suche/Filter-Prädikate (KEIN
      Duplikat zu Tree — eine gemeinsame beanMatches()-Funktion)
- [x] Sort-Toggle S zyklisch status->prio->created->updated->status
      (data.StatusRank/PriorityRank aus T1, CreatedAt/UpdatedAt nil-sicher)
- [x] `b` öffnet aus Tree, esc/b zurück — geteilter Filter/Suche-Zustand bleibt
      beim Zurückwechseln erhalten
- [x] backlog.golden neu + Determinismus-Test
- [x] go test ./... grün, make build ok

## Summary

view_browse_backlog.go (neu): backlogVisible() (idx.Backlog() + m.beanMatches,
KEIN zweiter Filter — Task 3/4 unverändert wiederverwendet), sortBacklog +
nextBacklogSort (Sort-Toggle-Zyklus), backlogSelected(), backlogRowText/
backlogRows (D08-Cursor-Balken, windowAround-Wiederverwendung), renderBacklogDetail
Pane (Task-1-Accordion via beanSections/renderAccordion, gleiche Breiten-Algebra
wie renderDetailPane), viewBacklog() (Master-Detail-Chrome, identischer Aufbau zu
viewBrowseRepo), keyBacklog() (up/down/Enter/S/‌/f/b/esc). focusedBean() (update.go)
bekam den viewBacklog-Case (`return m.backlogSelected()`) — der von T2 offen
gelassene Seam funktioniert genau wie geplant, keyDetailFocus wird unverändert
wiederverwendet. keyTree gewann den `b`-Case (view=viewBacklog + backlogList.setLen).
`/`und `f` wurden aus keyTree in zwei geteilte Helper extrahiert (openSearchInput/
openFilterMenu, update.go) — keyBacklog ruft dieselben Funktionen, keine zweite
Search-/Filter-Öffnungs-Logik.

ERRATUM gefunden und korrigiert (nicht wörtlich portiert): der Plan-Snippet für
den Sort-Zyklus (`modes := []string{"", "priority", "created", "updated", "status"}`,
"find current, advance, wrap") hätte fünf STRING-Zustände durchlaufen, obwohl ""
und "status" optisch IDENTISCH rendern (idx.Backlog()'s canonical order IST bereits
status-tier order, sortBacklog's ""-Case ist ein No-op genau deshalb). Das
5-Elemente-Array wrapt von "status" (Index 4) zurück zu "" (Index 0) statt zu
"priority" (Index 1) — ein toter 5. Tastendruck pro Umlauf (Zustand ändert sich,
Anzeige nicht), im Widerspruch zur eigenen Akzeptanz-Formulierung ("S zyklisch
status->prio->created->updated->status", 4 Modi). Fix: backlogSortModes ist ein
kanonisches 4-Element-Array (view_browse_backlog.go), nextBacklogSort behandelt ""
nur als ALIAS für "status" beim Nachschlagen der aktuellen Position — sobald einmal
zyklisiert wurde, wird "" nie wieder besucht (reiner Perioden-4-Zyklus, jeder
Tastendruck sichtbar). Test TestBacklogSortCyclesThroughFourModesAndBackToStart
deckt exakt das ab (inkl. 5. Tastendruck als Regressions-Beweis: muss "priority"
sein, nicht "").

Zweite Abweichung (Bugfix, kein Scope-Wechsel): backlogList (listState,
index-basiert) kann veralten, weil `/`- und `f`-Tastendrücke über die GETEILTEN
keySearchInput/keyFilterMenu-Handler laufen (update.go/box_filter_facets.go), die
NICHTS von backlogList wissen (sie kennen nur den Tree-cursorID-Mechanismus).
keyBacklog resynct backlogList.setLen() deshalb bei JEDEM Tastendruck, den es
selbst behandelt, bevor up/down bewegt wird — schließt die Lücke, ohne die
geteilten Handler anzufassen (Rendering selbst hängt nur von .cursor ab, nie von
.length, also kein Render-Bug, nur ein potenziell falscher move()-Bound ohne den
Resync).

View()-Dispatch (view_browse_repo.go) bekam den `case viewBacklog` — Abweichung
vom Plan-Text, der das fälschlich unter "Modify: update.go" führte: View() lebt in
diesem Repo tatsächlich in view_browse_repo.go, nicht update.go (Plan-Text-
Ungenauigkeit, keine Logik-Frage — 3-Zeilen-Korrektur an der TATSÄCHLICHEN
Fundstelle).

renderBacklogDetailPane dupliziert bewusst ~15 Zeilen aus renderDetailPane
(view_browse_repo.go) statt zu konsolidieren — Plan-Files-Liste für Task 5 nennt
view_browse_repo.go nicht als Modify-Ziel, die Konsolidierungs-Pflicht aus dem Plan
bezieht sich explizit nur auf das Filter-Menü (Task 4) und focusedBean() (Task 2),
nicht auf Pane-Rendering. Akzeptierte Kosten der Scope-Grenze.

Tests: 12 neue Tests in view_browse_backlog_test.go (backlogVisible inkl. nil-idx-
Guard, geteilte Suche/Facetten, Sort-Zyklus inkl. Erratum-Regressionstest,
Sort-Reihenfolgen inkl. nil-Timestamp-Sicherheit, D03-Fokus-Tausch, focusedBean-
Wiederverwendung, b-open/esc-Rückkehr mit geteiltem Filter-Zustand, Windowing) +
TestBacklogGolden/TestBacklogGoldenDeterministic (neues backlog.golden, 100x30,
2x ohne -update grün). go test ./... 2x -count=1 grün, zusätzlich 1x -race grün,
gofmt -l . leer, go vet ./... leer, make build ok (bin/bt).

tmux-Smoke (Scratch-Repo mit 3 Backlog-fähigen + 1 completed Bean, da das
dogfooding-Repo selbst 0 Backlog-Beans hat — alles ist unter dem einen Milestone
geparkt): `b` aus Tree öffnet Backlog (Breadcrumb "Backlog", 3 Zeilen, completed
Bean korrekt ausgeblendet). S 1x -> Priority-Sortierung (critical-Bug zuerst). S 2x
-> Created (Tie auf Sekunden-Ebene -> stabiler Fallback auf kanonische Baseline,
korrekt). S 4x -> zurück auf Status-Reihenfolge (== Ausgangszustand, Zyklus
bestätigt). Enter öffnet Detail-Fokus auf dem SELECTED Backlog-Eintrag (nicht auf
einem Tree-Rest); Ziffer 3 springt zu Beziehungen ("(no relations)" für die
relationslosen Fixture-Beans). j (links) verlässt Detail-Fokus zurück auf
Listen-Ebene; esc von dort zurück zu Tree (Breadcrumb wieder "Browse"). Einzige
Beobachtung (kein Task-5-Bug): esc INNERHALB von Detail-Fokus ist ein No-op (nur
j/tab verlassen ihn) — das ist bestehendes, unverändertes Verhalten aus T2s
keyDetailFocus (dort schon so, hier nur wiederverwendet), siehe Notiz unten.

Notiz für T6 (E2-Abschluss):
- I01 (Improvement, kein Blocker): esc innerhalb von Detail-Fokus (m.detailFocus
  == true) ist ein No-op in keyDetailFocus (update.go, aus T2) — nur j/left oder
  tab verlassen den Detail-Fokus. Gilt identisch für Tree UND Backlog (gleiche
  Funktion). Falls das beim E2-Gesamt-Review als Lücke auffällt: gehört zu T2s
  Scope, nicht T5s — hier nur dokumentiert, nicht angefasst.
- I02 (Improvement, kein Blocker): kein sichtbarer Sort-Modus-Indikator im UI
  (z.B. Pane-Titel "Backlog [sort: priority]") — devd hat das (backlogFilterSummary
  hängt "Sort:<mode>" an, view_browse_backlog.go:189-199 dort), design-spec.md V3/
  US-09 verlangt es nicht explizit und der Plan cuttet bewusst das schwebende
  Sort-Menü inkl. dessen Anzeige. Bewusst NICHT nachgerüstet (Scope-Treue), aber
  als PO-Entscheidungspunkt hier vermerkt, falls beim Dogfooding vermisst.
- Alle Seams aus T2/T3/T4 (focusedBean, beanMatches, treeActive, keyDetailFocus,
  openSearchInput/openFilterMenu) wurden UNVERÄNDERT wiederverwendet, keine
  Parallel-Implementierung entstanden — E2-Abschluss-Selbst-Review-Punkt
  "Konsolidierung ggü. devd" ist für Task 5 vollständig erfüllt.
