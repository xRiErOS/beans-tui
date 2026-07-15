---
# bt-dlgk
title: E3 T1 — Status/Type/Prio-Menü + B01-Fix + Enum-Single-Source + Mutation-Infra
status: completed
type: task
priority: high
created_at: 2026-07-15T00:25:55Z
updated_at: 2026-07-15T01:04:42Z
parent: bt-gzcu
---

Ziel: Status-/Type-/Prio-Menü (`s`, EIN kombiniertes Value-Menü mit 3 Gruppen —
design-spec §7 reserviert nur EINEN Key für dieses Cluster, nicht drei) + B01-Fix
(keys.FilterClear-Case fehlt in keyBacklog, aus E2-Abschluss übernommen) +
Enum-Single-Source (data.StatusValues/TypeValues/PriorityValues, ersetzt
box_filter_facets.go's eigenes Duplikat) + geteilte Mutation-Cmd/ETag-Konflikt-Infra
(erste Mutation Ende-zu-Ende verdrahtet; jede Folge-Task nutzt dieselbe Infra statt
eigener Cmd/Msg-Typen).

Plan: docs/plans/v1-port/epic-E3-plan.md »Task 1«.

## Akzeptanz
- [x] B01: X (FilterClear) wirkt in keyBacklog wie in keyTree (Regressionstest)
- [x] data.StatusValues()/TypeValues()/PriorityValues() exportiert (kanonische
      Tier-Reihenfolge aus statusOrder/typeOrder/priorityOrder); box_filter_facets.go
      buildFilterItems auf diese Single Source umgestellt (Duplikat-Fix)
- [x] box_menu_value.go: kombiniertes Status/Type/Priority-Menü (modalBox/menuList,
      3 Gruppen wie box_filter_facets.go facetHead-Muster), enter wendet SOFORT an
      + schließt (Port beans-src statuspicker.go Enter-Semantik)
- [x] keyNodeAction-Scaffold in update.go: zentraler Dispatch für s/t/a/B/c/d/e direkt
      neben dem bestehenden Refresh-Check in handleKey, guarded auf
      focusedBean()!=nil (außer c/Create)
- [x] messages.go: mutationDoneMsg + mutationCmd-Producer, applyMutationResult
      (jede Mutation, Erfolg wie Fehler, reloaded unconditional via loadCmd;
      ErrConflict zusätzlich Statuszeilen-Hinweis, kein Toast — E5)
- [x] go test ./... grün, gofmt/vet leer

## Findings / Umsetzung

B01: `keyBacklog` (view_browse_backlog.go) bekam den fehlenden `keys.FilterClear`-Case
zwischen `keys.Filter` und `keys.Backlog` — ruft denselben `clearFacets()`-Helper wie
`keyTree` und `keyFilterMenu` auf, plus `backlogList.setLen`-Resync danach. RED bewiesen
(TestKeyBacklogFilterClearResetsFacets schlug vorher fehl), danach grün.

Enum-Single-Source: `data/index.go`s drei Order-Maps (`statusOrder`/`typeOrder`/
`priorityOrder`) sind jetzt aus geordneten Slices (`statusValues`/`typeValues`/
`priorityValues`) via `rankMap` abgeleitet (Position == Rank), exportiert über
`StatusValues()`/`TypeValues()`/`PriorityValues()` (defensive Kopien, per Test
abgesichert: `TestValueSlicesAreDefensiveCopies`). `box_filter_facets.go`s
`buildFilterItems` konsumiert dieselben Accessor statt der vorherigen 15
hartkodierten Zeilen — dabei wechseln die Type-Labels von kapitalisiert
("Milestone") zu klein ("milestone"), konsistent mit Status/Priority (kein Golden
betroffen, nur Facet-Counts getestet).

box_menu_value.go (neu): `valueMenuItem{group,value}`, `buildValueMenuItems()` (15
Zeilen: 5 Status + 5 Type + 5 Priority in kanonischer Reihenfolge),
`valueMenuCursorFor` (Cursor-Seed auf aktuellen Wert, Port statuspicker.go
selectedIndex), `openValueMenu`/`keyValueMenu`/`applyValueMenuSelection`/
`valueMenuBox`. Enter wendet sofort an + schließt (Design a3); esc/s schließen ohne
Mutation (Port keyFilterMenu-Präzedenz). "(current)"-Marker je Gruppe (Port
statuspicker.go isCurrent).

Mutation-Infra: `messages.go` bekommt `mutationDoneMsg`/`mutateCmd` (geteilter Cmd-
Producer, jede Set*/Add*/Remove*/Delete-Mutation läuft künftig darüber) sowie
`createDoneMsg`/`createCmd` (Task-4-Vorgriff laut Plan-Files-Liste — Typen/Producer
jetzt angelegt, der Update-Dispatch-Case + Cursor-Jump folgt in T4, aktuell toter
aber unschädlicher Code, kein Vet-Fund). `update.go` bekommt `applyMutationResult`
(IMMER `loadCmd`, ErrConflict → "Konflikt: Bean extern geändert — neu geladen"),
`beanETag` (liest IMMER den aktuellen `m.idx.ByID[id].ETag`, nie eine beim Öffnen
eingefrorene Kopie — Design d), `keyNodeAction` (zentraler s/t/a/B/c/d/e-Dispatch,
`c` funktioniert ohne fokussiertes Bean, alle anderen sind handled-but-silent no-op
ohne fokussiertes Bean, ENTSCHEIDUNG lt. Plan: kein User-Text, da T2-T6 unmittelbar
folgen) und `keyOverlay` (Dispatch-Switch für `m.overlay`, im Plan als "per-Overlay-
Handler-Switch" benannt, nicht explizit mit eigenem Funktionsnamen vorgegeben — hier
als separate Funktion neben `keyNodeAction` gebaut, da sie eine andere Concern ist:
Routing bei GEÖFFNETEM Overlay vs. Öffnen/Dispatch der Node-Keys).

`overlayID`-Enum (types.go) mit allen 6 künftigen Werten (design a2: EIN geschlossenes
Set, T2-T6 ergänzen nur `case`s in `keyOverlay`/`composeOverlays`, nie neue Bools).
Nur T1-Felder (`overlay`/`mutTarget`/`menu`/`menuItems`) tatsächlich auf `model`
angelegt — die restlichen (`tagItems` etc.) bleiben für T2-T4 ausständig, um keine
toten Felder vor ihrer Verwendung einzuführen.

`composeOverlays` (view_browse_repo.go) extrahiert den bisher in `viewBrowseRepo` UND
`viewBacklog` duplizierten `filterOpen`/`confirmQuit`-Compositing-Block in EINE
Methode (Reihenfolge: Filter-Menü, dann `m.overlay`-Switch, dann Quit-Confirm zuletzt)
— beide Views rufen sie jetzt auf. `localHint` in beiden Views um `keys.Status`
ergänzt (Footer zeigt "s:Status menu").

Capture-Reihenfolge in `handleKey`: `confirmQuit` → `searchActive` → `filterOpen` →
`m.overlay != overlayNone` (neu) → globaler ctrl+c/q/tab-Switch → Refresh →
`keyNodeAction` (neu) → `detailFocus` → `keyBacklog`/`keyTree` — exakt wie im Plan
vorgegeben (m.form-Capture ist T4-Scope, noch nicht vorhanden).

Golden-Regeneration: `tree.golden`/`backlog.golden` einmalig neu erzeugt (Footer-Zeile
"tab:focus" → "s:Status menu  tab:focus"), Diff auf genau diese eine Zeile geprüft,
Determinismus via zweitem `-update`-Lauf + Full-Suite-Re-Run verifiziert.

Tests (neu): `view_browse_backlog_test.go` (B01), `index_test.go` (5x Enum-Single-
Source), `box_menu_value_test.go` (17 neue Tests: openValueMenu, Enter je Gruppe
(Status/Type/Priority — Dispatch über echte data.Client-Fehlermeldung "beans update"
verifiziert statt gemockt), esc/s-Close, "(current)"-Marker, keyNodeAction-Guard,
applyMutationResult (Conflict/Success/Non-Conflict), beanETag (live + vanished),
Vanished-Target-Graceful-Close).

`go test ./... -race -count=1` grün (alle Pakete), `gofmt -l .` leer, `go vet ./...`
leer, `go build ./...` ok.

Manueller Smoke (tmux, SCRATCH-Repo `bt-e3t1-smoke-*` — NICHT dieses Repos `.beans/`):
Bean angelegt (`beans create`), `bt` gestartet, `s` → Value-Menü öffnet mit Cursor auf
`todo (current)`, `k` zweimal auf `draft`, `enter` → Overlay schließt, `.md`-Datei
zeigt `status: draft` + neuer `updated_at`/ETag (frontmatter tatsächlich verändert,
kein Mock). Menü erneut geöffnet: Cursor seedet jetzt korrekt auf `draft (current)`.
Scratch-Repo + Binary danach gelöscht.

Commit: siehe Refs unten (main, direkt — Worktree-Weiche main-direkt für dieses Repo).

Notizen für T2 (Tag-Picker, bt-8v69): wiederverwendet die hier gebaute Overlay-Infra
(`overlayID`, `keyOverlay`-Dispatch, `composeOverlays`) und Mutation-Infra
(`mutateCmd`/`mutationDoneMsg`/`applyMutationResult`/`beanETag`) unverändert — T2 fügt
nur `overlayTagPicker` als neuen `case` in `keyOverlay`/`composeOverlays` hinzu plus
die T2-eigenen Felder (`tagItems`/`tagPending`/`tagOriginal`/`tagInput*`) auf `model`,
keine neue Overlay-Maschinerie nötig.
