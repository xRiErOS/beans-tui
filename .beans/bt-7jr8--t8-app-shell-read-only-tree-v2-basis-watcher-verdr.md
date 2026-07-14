---
# bt-7jr8
title: T8 App-Shell + read-only Tree (V2-Basis) + Watcher-Verdrahtung
status: todo
type: task
priority: high
created_at: 2026-07-14T18:34:04Z
updated_at: 2026-07-14T20:09:02Z
parent: bt-blsy
blocked_by:
    - bt-gbfe
    - bt-6yr8
    - bt-5544
---

Plan: implementation-plan.md »E1 Task 8«. Port-Muster: view_browse_project.go, Datenpfad data.Client+data.Index.

## Akzeptanz
- [ ] model/viewID/update-Dispatcher, async Load, Tree Milestones→Epics→Tasks (expand/collapse/cursor, Glyph+Farbe+ID+Titel)
- [ ] q→Quit-Confirm, ctrl+r Reload, Watcher-Event→Reload mit Cursor-Restore (per ID)
- [ ] Update-Tests (3 laut Plan) + TestTreeGolden grün
- [ ] cmd/tui.go: FindRepo→Client→tea.Program (AltScreen+Mouse); manueller Dogfooding-Smoke belegt


## Ergänzung aus T3-Review (Q01, entschieden)
beans erlaubt dangling parents (.md frei editierbar; `beans check` meldet broken_links nur). Der Tree MUSS Orphans sichtbar machen: beans mit nicht-auflösbarem parent erscheinen unter synthetischem Root-Knoten „(verwaist)" statt still zu verschwinden.


## Ergänzung aus T5-Re-Review (B05, PFLICHT in diesem Task)
- [ ] Doc-Zeile an Watch() ergänzen: onChange darf stop() NIEMALS synchron (aus dem Callback heraus) aufrufen — Deadlock. T8-Consumer nutzt async Dispatch (tea-Msg), stop() nur im Teardown-Pfad.
