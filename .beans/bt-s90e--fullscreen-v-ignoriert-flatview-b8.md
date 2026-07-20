---
# bt-s90e
title: Fullscreen (v) ignoriert flatView (B8)
status: todo
type: bug
priority: low
created_at: 2026-07-20T07:26:50Z
updated_at: 2026-07-20T07:26:50Z
parent: bt-vy1q
---

**B8, gefunden in S5 (2026-07-20).** Der Nested/Flat-Toggle `G` wirkt nur in der Browse-Split-Ansicht. Der Fullscreen-Modus (`v`, `view_fullscreen.go`) rendert immer den Tree, auch wenn `m.flatView` gesetzt ist — inkonsistent.

## Akzeptanz
- [ ] Fullscreen respektiert `m.flatView` (Flat-Liste statt Tree, wenn aktiv)
- [ ] `G` funktioniert auch im Fullscreen (oder ist dort bewusst deaktiviert — dann dokumentieren)
- [ ] Bestandsgolden unveraendert bzw. bewusst regeneriert
- [ ] Voller `command go test ./...` gruen
