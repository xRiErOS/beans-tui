---
# bt-4ep2
title: 'T3 Suche /: lokaler Live-Filter + Bleve ab 3 Zeichen (async)'
status: completed
type: task
priority: high
created_at: 2026-07-14T21:57:22Z
updated_at: 2026-07-14T23:10:36Z
parent: bt-aq5s
---

Ziel: Suche `/` — lokaler Live-Filter (Titel-Substring, sofort) + Bleve-Modus ab
3 Zeichen (`beans list -S <query> --json --full` async, staleness-sicher gegen
schnelles Tippen). esc-Kaskade: Suche abbrechen -> Query+Filter leeren (kein
Lobby-Fallback in E2, Single-Repo-Start).

Plan: docs/plans/v1-port/epic-E2-plan.md »Task 3«.

## Akzeptanz
- [x] data.Client.Search(query) — beans list --json --full --search <query>
      (Test gegen Fixture-Repo)
- [x] flattenTreeFiltered(idx, expanded, match) — Vorfahren von Treffern bleiben
      Kontext, kollabierte Vorfahren verstecken Treffer weiterhin
      (DD2-178-Parität, Port-Ref view_browse_project.go:215-238)
- [x] Live-Tipp-Test: jedes Zeichen aktualisiert m.searchQuery + Cursor-Reset;
      enter committet, esc bricht ab+leert
- [x] Bleve-Cmd feuert ab 3 Zeichen, Ergebnis-Msg trägt die Ziel-Query — veraltete
      Antworten (Query != aktuelle Query) werden verworfen
- [x] go test ./... grün



## Summary

Implementiert 1:1 nach Plan (epic-E2-plan.md Task 3), TDD durchgehend (jeder
Failing-Test vor Implementierung verifiziert). data.Client.Search (client.go)
+ flattenTreeFiltered/filteredBeanNode/subtreeHasMatch (view_browse_repo.go,
DD2-178-Parität: kollabierter Vorfahre mit Treffer im Teilbaum rendert als
EINE Kontext-Zeile, kein Auto-Expand) + keySearchInput/dispatchBleveIfDue/
maybeBleveCmd (update.go) + searchBleveResultMsg/searchCmd (messages.go) +
treeSearchLine-Kopfzeile (view_browse_repo.go, kostet 1 Zeile vom Tree-Pane-
Budget, bodyH-2 -> bodyH-3). beanMatchesSearch matcht Titel UND ID
(Task-Vorgabe, über das bloße Plan-Snippet hinaus). Single-Key-Shortcuts
inaktiv bei fokussierter Eingabe: searchActive wird in handleKey VOR dem
ctrl+c/q/tab-Switch abgefangen (Konsistenz mit dem bestehenden
confirmQuit-Präzedenzfall, der ctrl+c ebenfalls schluckt).

Deviation (dokumentiert, keine stille Abweichung): Der Task-Prompt nannte
zusätzlich einen ~200ms-Debounce; Plan-Text UND Bean-Akzeptanz beschreiben
explizit NUR den Staleness-Guard ("async, staleness-sicher" /
"Staleness-Guard statt Debounce-Timer", Commit-Rationale im Plan) und der
Plan-eigene Test (TestSearchBleveFiresOnlyAtThreeOrMoreChars) ist auf
synchrones Feuern ab 3 Zeichen ausgelegt. Umgesetzt nach Plan+Bean
(doppelt/authoritativ abgesichert): jede qualifizierende Tastatureingabe
(>=3 Zeichen, Query geändert) feuert ihren eigenen beans-CLI-Subprozess,
nur die zur AKTUELLEN Query passende Antwort wird übernommen
(applyBleveResult). Kein Timer/Generation-Counter gebaut.

go.mod/go.sum: github.com/atotto/clipboard als transitive Dependency von
bubbles/textinput ergänzt (go mod tidy, keine Netzwerk-Fetches nötig, war
im Modul-Cache).

Golden: tree.golden neu generiert (Suchkopfzeile + /:Search im Footer-Hint
verschieben das Zeilenbudget), 2x ohne -update grün (Determinismus
verifiziert).

tmux-Smoke (bin/bt im eigenen Repo, echte beans-Daten): / -> "Watcher"
getippt -> Live-Filter zeigt zunächst nur die kollabierte Milestone-Zeile
(bt-apmy, Treffer im Teilbaum verborgen, DD2-178). enter committet (Query
rot). Rechts-Pfeil expandiert Milestone -> Epic (bt-blsy) -> beide Treffer
sichtbar: bt-6yr8 "T5 Datenlayer: Debounced fsnotify-Watcher" und bt-7jr8
"T8 App-Shell ... Watcher-Verdrahtung". esc leert Query+Filter, volle
Tree-Ansicht kehrt zurück, Kopfzeile zurück auf "⌕ / search"-Hinweis.

Notes für T4: beanMatches (Task 4) AND-kombiniert beanMatchesSearch (hier)
mit beanMatchesFacets; treeActive() ersetzt treeSearchActive() als
visibleNodes()-Weiche. flattenTreeFiltered/filteredBeanNode/subtreeHasMatch
bleiben unverändert wiederverwendbar (match-Prädikat ist bereits generisch
func(*data.Bean) bool).
