---
# bt-4ep2
title: 'T3 Suche /: lokaler Live-Filter + Bleve ab 3 Zeichen (async)'
status: todo
type: task
priority: high
created_at: 2026-07-14T21:57:22Z
updated_at: 2026-07-14T21:57:22Z
parent: bt-aq5s
---

Ziel: Suche `/` — lokaler Live-Filter (Titel-Substring, sofort) + Bleve-Modus ab
3 Zeichen (`beans list -S <query> --json --full` async, staleness-sicher gegen
schnelles Tippen). esc-Kaskade: Suche abbrechen -> Query+Filter leeren (kein
Lobby-Fallback in E2, Single-Repo-Start).

Plan: docs/plans/v1-port/epic-E2-plan.md »Task 3«.

## Akzeptanz
- [ ] data.Client.Search(query) — beans list --json --full --search <query>
      (Test gegen Fixture-Repo)
- [ ] flattenTreeFiltered(idx, expanded, match) — Vorfahren von Treffern bleiben
      Kontext, kollabierte Vorfahren verstecken Treffer weiterhin
      (DD2-178-Parität, Port-Ref view_browse_project.go:215-238)
- [ ] Live-Tipp-Test: jedes Zeichen aktualisiert m.searchQuery + Cursor-Reset;
      enter committet, esc bricht ab+leert
- [ ] Bleve-Cmd feuert ab 3 Zeichen, Ergebnis-Msg trägt die Ziel-Query — veraltete
      Antworten (Query != aktuelle Query) werden verworfen
- [ ] go test ./... grün
