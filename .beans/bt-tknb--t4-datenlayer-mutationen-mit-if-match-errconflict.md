---
# bt-tknb
title: 'T4 Datenlayer: Mutationen mit --if-match (ErrConflict)'
status: todo
type: task
priority: high
created_at: 2026-07-14T18:34:04Z
updated_at: 2026-07-14T18:52:55Z
parent: bt-blsy
blocked_by:
    - bt-xmrs
---

Plan: implementation-plan.md »E1 Task 4«.

## Akzeptanz
- [ ] Create/SetStatus/SetPriority/SetType/SetTitle/AddTag/RemoveTag/SetParent/RemoveParent/AddBlockedBy/RemoveBlockedBy/AppendBody/Delete
- [ ] Updates senden --if-match; typed ErrConflict bei Stale-ETag
- [ ] Tests inkl. TestConflictOnStaleETag grün


## Übernommene Review-Findings aus T2-Quality-Review (PFLICHT in diesem Task)
- [ ] B01: TestListErrorIncludesStderr muss stderr-Inhalt wirklich asserten (strings.Contains auf stabilen Substring)
- [ ] I01: Fixture-bean mit tags/blocking/blocked_by in newTestRepo + Assertion der geparsten Slices (JSON-Vertrag als Regression-Guard)
- [ ] I02: client.go List: Unmarshal-Fehler mit Kommando-Kontext wrappen
- [ ] I03: client.go run: trailing ": " bei leerem stderr vermeiden
- [ ] I04: discover.go: Fehlermeldung mit resolved-Pfad statt raw start-Arg
