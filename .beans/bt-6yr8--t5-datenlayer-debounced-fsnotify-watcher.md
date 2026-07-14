---
# bt-6yr8
title: 'T5 Datenlayer: Debounced fsnotify-Watcher'
status: todo
type: task
priority: high
created_at: 2026-07-14T18:34:04Z
updated_at: 2026-07-14T18:34:04Z
parent: bt-blsy
blocked_by:
    - bt-snkb
---

Plan: implementation-plan.md »E1 Task 5«.

## Akzeptanz
- [ ] `internal/data/watcher.go` Watch(repoDir, onChange) mit stop-Func, Debounce 150 ms, .beans/ + archive/
- [ ] TestWatcherFiresOnceForBurst grün
