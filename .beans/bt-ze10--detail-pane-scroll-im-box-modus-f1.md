---
# bt-ze10
title: Detail-Pane Scroll im Box-Modus (F1)
status: completed
type: task
priority: high
created_at: 2026-07-20T07:25:11Z
updated_at: 2026-07-20T07:50:03Z
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
- [x] Langer Body: Detail-Pane scrollt, Relations/History erreichbar
- [x] Offset clamped (kein Scroll ueber Anfang/Ende hinaus)
- [x] Nur im Box-Modus aktiv; Accordion-Verhalten unveraendert
- [x] Bestandsgolden byte-identisch (Flag AUS)
- [x] Tests: Scroll-Offset via `tea.KeyMsg` + `tea.MouseMsg` (Wheel), Clamping an beiden Enden
- [x] Voller `command go test ./...` gruen

## Summary

Added a clamped scroll viewport for the box-form Detail-Pane (BT_BOXFORM=1 only). detailBoxForm's rendered string (Title + scalar grid + Body/Relations/History) is now windowed via scrollView (the same primitive windowRelationsSection already uses for the accordion's Relations section) instead of being silently cut off by renderPane's line cap.

- `internal/tui/types.go`: two new model fields, `boxFormScroll` (offset) and `boxFormScrollBean` (owning bean ID). The owning-bean field enables a *derived* reset instead of hunting down every cursor-move/click/jump call site that can change the selected bean: `boxFormEffectiveScroll` (mouse.go) returns 0 whenever the currently focused bean's ID no longer matches `boxFormScrollBean`.
- `internal/tui/mouse.go`: `boxFormEffectiveScroll`, `boxFormScrollBounds` (reconstructs the Detail pane's real content-height budget via the existing `clickPaneGeometry` helper, minus the filter-bar row when applicable — matches exactly what render time uses), `adjustBoxFormScroll` (single mutation point shared by keyboard and wheel), and `boxFormWheelHit` (X/Y bounds test, mirrors `detailBoxFormClickRow`'s geometry). `wheelMove` now takes the `tea.MouseMsg` and scrolls the box-form pane when the wheel lands inside its bounds and `boxFormEnabled()`; otherwise falls through to the pre-existing cursor-move behavior unchanged.
- `internal/tui/update.go`: `keyDetailFocus`'s up/down are repurposed to call `adjustBoxFormScroll` when `boxFormEnabled()` and NOT inside `fullscreenDetail` (fullscreen box-form scrolling is an explicit, documented scope-cut — its geometry differs from the split pane's and isn't part of this task's file list). No new keybindings — reuses `keys.Up`/`keys.Down`, so `helpGroups()` needed no change.
- `internal/tui/box_detail_form.go`: `clampBoxFormScroll(offset, total, height)` — the same clamp algebra `scrollView` applies internally, duplicated (not shared) because callers need the clamped int back, not a rendered string.
- `internal/tui/view_browse_repo.go` / `view_fullscreen.go`: `renderAccordionPane` gained a `boxScroll int` parameter. The box-form branch now does `win, _ := scrollView(form, h, boxScroll)` before splitting into rows. At `boxScroll==0` with content that already fits, this is byte-identical to the pre-existing output (verified: `browse_boxform.golden` needed no regeneration). `renderFullscreenBody` passes a literal `0` (fullscreen scope-cut, documented at the call site).

Scope cut (documented in code): fullscreenDetail's box-form view does not scroll in this slice — its own single-pane geometry differs from the split Detail pane's, and view_fullscreen.go was not in the bean's own "Betroffen" file list.

## Test-Output

`command go test ./...` (full suite, no `-short`), run twice — once before writing new tests (baseline green) and once after:

```
ok  	github.com/xRiErOS/beans-tui/cmd	(cached)
ok  	github.com/xRiErOS/beans-tui/internal/config	(cached)
ok  	github.com/xRiErOS/beans-tui/internal/data	(cached)
ok  	github.com/xRiErOS/beans-tui/internal/theme	(cached)
ok  	github.com/xRiErOS/beans-tui/internal/tui	150.138s
```

New test file `internal/tui/box_form_scroll_test.go` (7 tests): offset changes on up/down only when Detail-focused (and is provably inert when Tree-focused), clamping at both ends, mouse wheel over the Detail pane scrolls (and is inert over the Tree pane), offset resets when the selected bean changes (direct + "next adjust doesn't compound the stale value" check), a long-body bean's previously-cut-off tail becomes visible via the real `m.View()` render pipeline after scrolling, and accordion mode (flag off) never touches the new fields.

No goldens changed — `browse_boxform.golden` and `detail_boxform.golden` stayed byte-identical (confirmed by the full suite passing without `-update`).
