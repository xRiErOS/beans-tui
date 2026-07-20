---
# bt-s90e
title: Fullscreen (v) ignoriert flatView (B8)
status: todo
type: bug
priority: low
created_at: 2026-07-20T07:26:50Z
updated_at: 2026-07-20T07:51:05Z
parent: bt-vy1q
---

**B8, gefunden in S5 (2026-07-20).** Der Nested/Flat-Toggle `G` wirkt nur in der Browse-Split-Ansicht. Der Fullscreen-Modus (`v`, `view_fullscreen.go`) rendert immer den Tree, auch wenn `m.flatView` gesetzt ist — inkonsistent.

## Akzeptanz
- [ ] Fullscreen respektiert `m.flatView` (Flat-Liste statt Tree, wenn aktiv)
- [ ] `G` funktioniert auch im Fullscreen (oder ist dort bewusst deaktiviert — dann dokumentieren)
- [ ] Bestandsgolden unveraendert bzw. bewusst regeneriert
- [ ] Voller `command go test ./...` gruen


## Prelude aus bt-ze10 (2026-07-20) — zuerst erledigen

Der Detail-Scroll (Commit `8e5a869`) hat den **Fullscreen-Pfad bewusst ausgespart**:
`fullscreenDetail` im Box-Modus scrollt NICHT (Single-Pane-Geometrie weicht von der
Split-Pane ab; `view_fullscreen.go` stand nicht in der Datei-Liste jenes beans).
`renderFullscreenBody` uebergibt literal `0` als Scroll-Offset; Kommentare sitzen am
`keyDetailFocus`-Guard und an der Aufrufstelle.

Damit sammeln sich **zwei** Fullscreen-Luecken an derselben Stelle:
1. `v` ignoriert `flatView` (B8, dieses bean)
2. `v` ignoriert im Box-Modus den Scroll-Offset (aus ze10)

Beide in `view_fullscreen.go`. Zusammen erledigen — getrennt waere doppelte Einarbeitung
in dieselbe Geometrie.
