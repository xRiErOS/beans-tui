---
# bt-7jr8
title: T8 App-Shell + read-only Tree (V2-Basis) + Watcher-Verdrahtung
status: completed
type: task
priority: high
created_at: 2026-07-14T18:34:04Z
updated_at: 2026-07-14T21:39:08Z
parent: bt-blsy
blocked_by:
    - bt-gbfe
    - bt-6yr8
    - bt-5544
---

Plan: implementation-plan.md »E1 Task 8«. Port-Muster: view_browse_project.go, Datenpfad data.Client+data.Index.

## Akzeptanz
- [x] model/viewID/update-Dispatcher, async Load, Tree Milestones→Epics→Tasks (expand/collapse/cursor, Glyph+Farbe+ID+Titel)
- [x] q→Quit-Confirm, ctrl+r Reload, Watcher-Event→Reload mit Cursor-Restore (per ID)
- [x] Update-Tests (3 laut Plan) + TestTreeGolden grün
- [x] cmd/tui.go: FindRepo→Client→tea.Program (AltScreen+Mouse); manueller Dogfooding-Smoke belegt


## Ergänzung aus T3-Review (Q01, entschieden)
beans erlaubt dangling parents (.md frei editierbar; `beans check` meldet broken_links nur). Der Tree MUSS Orphans sichtbar machen: beans mit nicht-auflösbarem parent erscheinen unter synthetischem Root-Knoten „(verwaist)" statt still zu verschwinden.


## Ergänzung aus T5-Re-Review (B05, PFLICHT in diesem Task)
- [x] Doc-Zeile an Watch() ergänzen: onChange darf stop() NIEMALS synchron (aus dem Callback heraus) aufrufen — Deadlock. T8-Consumer nutzt async Dispatch (tea-Msg), stop() nur im Teardown-Pfad.


## Ergänzungen aus T7-Quality-Review (PFLICHT in diesem Task)
- [x] I01: Unit-Tests für renderPane (focused-Border), borderedPane, tagsInline/tagSwatch, modalPanel nachziehen (Muster primitives_test.go)
- [x] I02: Reflection-Test: helpGroups() deckt JEDES keyMap-Binding genau einmal ab (Drift-Guard)
- [x] I03: ChromeOpts.fallbackAvail: Test für avail<4-Fallback-Pfad schreiben ODER Feld entfernen (YAGNI)
- [x] Q01: tab-Fokus-Tausch Tree↔Detail hier implementieren (View-lokal, nicht in keymap.Right — bewusst aus devd-Right entfernt)

## Summary of Changes

App-Shell + read-only Tree (V2-Basis) implementiert, `bt` ist eine laufende TUI.

**Neu:** `internal/tui/types.go` (model, viewID, orphanRootID-Sentinel),
`internal/tui/app.go` (newModel-Aufruf via Run, tea.Program AltScreen+Mouse,
Watch-Verdrahtung via `p.Send`), `internal/tui/update.go` (Update-Dispatcher,
handleKey, keyTree, applyLoaded mit Cursor-Restore-per-ID),
`internal/tui/messages.go` (beansLoadedMsg, watchMsg, loadCmd — nur Msg-Typen +
Cmd-Producer), `internal/tui/view_browse_repo.go` (treeNode-Flattening inkl.
Orphan-Sammlung, D08-Cursor-Balken-Rendering, Master-Detail-Layout),
`internal/tui/box_confirm_quit.go` (Quit-Confirm-Modal).

**Geändert:** `cmd/tui.go` (runTUI: FindRepo→Client→tui.Run),
`internal/data/watcher.go` (B05-Doc-Zeile: onChange darf stop() nie synchron
aufrufen — Deadlock-Warnung + Verweis auf den async-Dispatch-Konsumenten),
`go.mod`/`go.sum` (bubbletea v1.3.10 als direkte Dependency).

**Tests:** `internal/tui/update_test.go` (TestCursorMovesAndExpands,
TestQuitConfirm, TestCtrlCQuitsImmediately, TestReloadKeepsCursorOnID,
TestBeansLoadedErrorSurfacesInStatusLine, TestOrphanShownUnderSyntheticRoot),
`internal/tui/tree_golden_test.go` (TestTreeGolden 100×30 TrueColor,
TestTreeGoldenDeterministic) + `internal/tui/testdata/tree.golden`. T7-Nachzug
in bestehenden Dateien: `primitives_test.go`
(TestRenderPaneFocusedBorderColor, TestBorderedPanePadsAndCapsToHeight,
TestTagsInlineAndSwatch, TestModalPanelIncludesHeaderBodyFooter),
`keymap_test.go` (TestHelpGroupsCoverEveryBindingExactlyOnce),
`chrome_test.go` (TestChromeFallbackAvailOverride — Feld behalten, jetzt
getestet statt YAGNI-entfernt).

**Ergebnis:** `command go test ./...` grün (alle Pakete), `command gofmt -l .`
leer, `command go vet ./...` leer, `make build`/`command go build -o bin/bt .`
erfolgreich.

**Dogfooding-Smoke (tmux, dieses Repo):** `bin/bt` im eigenen Repo gestartet
zeigt sofort die eigene Entwicklung — Root-Milestone `bt-apmy beans-tui v1 —
devd-TUI-Port auf beans` (Status todo, Glyph ◉, Sapphire-ID), Detail-Pane zeigt
Titel+Meta. `l` (expand) klappt die realen Epics auf (`bt-blsy` E1 Foundation …
`bt-5h4d` E5 Polish), `k` (down) bewegt den Cursor auf `bt-blsy` und die
Detail-Pane aktualisiert live auf "E1 Foundation — Skeleton, Datenlayer,
Theme, App-Shell" / `in-progress epic high`. `q` öffnet den Quit-Confirm
("Quit? / Really quit bt. / enter: quit esc: cancel"), `enter` beendet den
Prozess sauber (tmux-Pane terminiert). Ausschnitt:

```
╭──────────────────────────────────────────────────────────────────────╮
│> beans-tui-repository: Browse              ctrl+r:Reload data ?:help │
│────────────────────────────────────────────────────────────────────  │
│╭──────────────────────────╮╭────────────────────────────────────────╮│
││Tree                      ││Detail                                 ││
││──────                    ││────────                                ││
││▌▸ ◉ ⬢ bt-apmy beans-tui v1 — devd-T…  ││beans-tui v1 — devd-TUI-Port auf beans│
││                          ││todo  milestone  high                   ││
```

(nach `l`: Baum zeigt `bt-blsy`/`bt-aq5s`/`bt-gzcu`/`bt-tfqi`/`bt-zk9p`/`bt-5h4d`
als Kinder von `bt-apmy` — reale beans dieses Repos, keine Fixture-Daten.)

**Notes für E2:** Detail-Pane ist bewusst Platzhalter (Titel+Meta); voller
Accordion (Meta/Body/Beziehungen/Historie) + Beziehungs-Navigation folgt in
E2 Task 1/2. `detailFocus` schaltet den Fokus bereits um (Border-Farbe
wechselt), hat aber noch keine eigene Navigation (Tasten sind No-Op, bis E2
die Section/Feld-Navigation liefert).

## Review-Fixes (Runde 2, Opus-Review)

| Code | Schwere/Nutzen | Beschreibung | Status |
|------|------|------|------|
| B01 | medium | Tree-Viewport-Windowing (Cursor-Follow): `treeRows` lieferte ALLE Knoten, `renderPane` clippte nur an Pane-Höhe → Cursor konnte unterhalb des Folds unsichtbar werden. Fix: `windowStart`/`windowAround` (devd-Port `view_browse_project.go:647-670`) fenstert die Rows um den Cursor, bevor sie an `renderPane` gehen. | 🟢 |
| B02 | low | Zyklus-Beans unsichtbar: Beans in reinem Parent-Zyklus (A→B→A) sind weder Roots (Parent gesetzt) noch Orphans (Parent löst auf) → verschwanden komplett. Fix: `collectCycleOrphans` (strukturelle Erreichbarkeit ab Roots+Orphans, expand-state-unabhängig via `reachableIDs`) sammelt jede unerreichbare Bean und hängt sie flach (ohne Rekursion — Duplikat-Schutz bei angehängten Sub-Bäumen) unter „(verwaist)" an. | 🟢 |
| I04 | Nutzen: kein stiller Feature-Ausfall mehr | Silent Watch-Failure: schlug `data.Watch` fehl, degradierte die App lautlos (kein Live-Reload, keine Meldung). Fix: `watchUnavailableMsg` (async via Goroutine — inline `p.Send` vor `p.Run()` würde am unbuffered Channel deadlocken) setzt `m.watchUnavailable` → Statuszeile zeigt „watch unavailable — ctrl+r für manuelles Reload". | 🟢 |
| I05 | Nutzen: Regressionsschutz für alle 3 Findings | Test-Lücken geschlossen: (1) Tab-Fokus-Tausch (Model-Flag + Pane-Akzent-Swap via `renderPane`/`renderDetailPane`), (2) Zyklus-Input (A→B→A, kein Hang, beide Beans unter (verwaist)), (3) Windowing (50 Knoten, Pane-Höhe 10, Cursor bei 40 → in `View()` sichtbar). Alle 3 schlugen vor dem Fix fehl (Kompilierfehler ggü. alter `treeRows`-Signatur bzw. fehlender Orphan-Root) — TDD-Rot-Zustand verifiziert (`git stash` der Fix-Dateien, Tests liefen erneut). | 🟢 |

Betroffene Dateien: `internal/tui/view_browse_repo.go` (`windowStart`/`windowAround`, `collectCycleOrphans`/`reachableIDs`/`sortByTitleThenID`, `treeRows`-Signatur +`bodyH`, Statuszeile), `internal/tui/types.go` (`watchUnavailable`-Feld), `internal/tui/messages.go` (`watchUnavailableMsg`), `internal/tui/update.go` (Update-Case), `internal/tui/app.go` (async `p.Send` bei Watch-Fehler), `internal/tui/update_test.go` (4 neue Tests: `TestTabTogglesDetailFocusAndPaneAccent`, `TestCycleBeansShowUnderOrphanRoot`, `TestTreeWindowingKeepsCursorVisible`, `TestWindowAroundStableAtEdges`).

**Ergebnis:** `command go test ./... -count=1` grün (2x hintereinander), `TestTreeGolden`/`TestTreeGoldenDeterministic` grün — Golden bleibt byte-identisch (100×30-Fixture hat 6 Knoten, weit unter der gefensterten Pane-Höhe, also No-Op-Windowing wie erwartet). `command gofmt -l .` leer, `command go vet ./...` leer.
