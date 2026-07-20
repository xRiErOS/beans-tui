---
# bt-ze10
title: Detail-Pane Scroll im Box-Modus (F1)
status: in-progress
type: task
priority: high
created_at: 2026-07-20T07:25:11Z
updated_at: 2026-07-20T07:27:46Z
parent: bt-vy1q
---

**F1 aus dem Spike-Review.** Die Box-Form rendert Title + 3|2-Grid + Body/Relations/History als einen hohen String. Ist er hoeher als die Detail-Pane, wird unten abgeschnitten — Body mittendrin, Relations/History unsichtbar. Es gibt KEIN Scrolling im Box-Modus (das Accordion hatte Klapp-/Scroll-Mechanik).

Durch die kompakte 3|2-Gruppierung (D12) passt der Normalfall inzwischen; ein LANGER Body ueberlaeuft weiterhin.

## Ziel
Scrollbarer Viewport fuer die Box-Form-Detail-Pane. Offene Design-Weiche D10 (mit PO klaeren falls unklar): Viewport-Scroll vs. nur-Fullscreen vs. kollabierbare Panels. **Vorschlag: Viewport-Scroll** — Offset-State, geclamped, bedienbar wenn Detail fokussiert (tab) via up/down + Mausrad ueber der Pane.

## Betroffen
- `internal/tui/view_browse_repo.go` — `renderAccordionPane` (dort sitzt der `boxFormEnabled()`-Zweig, ~Z.655)
- `internal/tui/box_detail_form.go` — `detailBoxForm`
- `internal/tui/types.go` — Scroll-Offset-Feld
- `internal/tui/update.go` — up/down wenn Detail fokussiert
- `internal/tui/mouse.go` — Mausrad ueber der Detail-Pane

## Akzeptanz
- [ ] Langer Body: Detail-Pane scrollt, Relations/History erreichbar
- [ ] Offset clamped (kein Scroll ueber Anfang/Ende hinaus)
- [ ] Nur im Box-Modus aktiv; Accordion-Verhalten unveraendert
- [ ] Bestandsgolden byte-identisch (Flag AUS)
- [ ] Tests: Scroll-Offset via `tea.KeyMsg` + `tea.MouseMsg` (Wheel), Clamping an beiden Enden
- [ ] Voller `command go test ./...` gruen
