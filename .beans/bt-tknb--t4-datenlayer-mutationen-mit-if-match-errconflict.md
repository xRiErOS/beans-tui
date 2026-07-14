---
# bt-tknb
title: 'T4 Datenlayer: Mutationen mit --if-match (ErrConflict)'
status: todo
type: task
priority: high
created_at: 2026-07-14T18:34:04Z
updated_at: 2026-07-14T18:34:04Z
parent: bt-blsy
blocked_by:
    - bt-xmrs
---

Plan: implementation-plan.md »E1 Task 4«.

## Akzeptanz
- [ ] Create/SetStatus/SetPriority/SetType/SetTitle/AddTag/RemoveTag/SetParent/RemoveParent/AddBlockedBy/RemoveBlockedBy/AppendBody/Delete
- [ ] Updates senden --if-match; typed ErrConflict bei Stale-ETag
- [ ] Tests inkl. TestConflictOnStaleETag grün
