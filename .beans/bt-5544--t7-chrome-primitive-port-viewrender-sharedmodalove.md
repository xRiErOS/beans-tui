---
# bt-5544
title: 'T7 Chrome-Primitive-Port: view/render_shared/modal/overlay/listState/keymap'
status: completed
type: task
priority: high
created_at: 2026-07-14T18:34:04Z
updated_at: 2026-07-14T20:39:09Z
parent: bt-blsy
blocked_by:
    - bt-8q9c
---

Plan: implementation-plan.md »E1 Task 7«. Port-Quellen: gleichnamige devd-Dateien (API-Client-Referenzen entfernen).

## Akzeptanz
- [x] Breadcrumb `> repo: Titel`, Footer-Hints, masterDetailWidths, modalBox, overlayModal, listState, zentrale Keymap (design-spec §7)
- [x] TestChromeGolden (80×24) + TestKeymapNoCtrlSQ grün

## Summary of Changes

Ported the six devd chrome-infrastructure files into `internal/tui/` (view.go,
render_shared.go, modal.go, overlay.go, list.go, keymap.go), stripped of all
devd-API/data-layer coupling. Since no model exists yet (T8 builds the
App-Shell), every chrome function takes plain strings/ints instead of a
`model` receiver — `Chrome(ChromeOpts)` bundles the render inputs a future
model will supply.

Notable adaptations vs. devd:
- Breadcrumb format `> repo: Titel` (repo name, not project slug).
- keymap.go trimmed to the design-spec §7 scope: global + node-focused
  bindings only. devd's Sort/Tags-Manager/Rename/Review-verdict/User-Story
  bindings are out of scope (Tag-Manager-CRUD and Docs/Notes views don't
  exist in beans-tui per design-spec §9, Review-Cockpit is a later epic) —
  added to this single source when their view lands, not speculatively.
- New binding `B` (Blocking-Picker) — no devd equivalent (beans-specific
  blocking/blocked_by relation).
- `Editor` binding gains a bare `e` alongside `ctrl+e` (design-spec §7).
- `tagsInline`/`tagSwatch` take plain tag names (`[]string`) instead of
  devd's `api.Tag{Color}` — beans tags carry no color field
  (`internal/data.Bean.Tags` is `[]string`), so color is hashed
  deterministically from the tag name over a small fixed Catppuccin palette
  (design-spec §8).
- devd's `issueFields`/`issueMetaPairs`/`entityMetaLine`/`metaGrid`/
  `detailTitle` (api.Issue-coupled Detail-view header primitives) were not
  ported — out of this task's "four-zone hull" scope, they land with the
  Detail view that needs them (T8+).
- Fixed a bug introduced during decoupling: `outerBorder`'s `width` param is
  the *content* width the border wraps around (lipgloss's `Border()` adds 2
  columns on top of `Width()`); `Chrome()` must pass the already-reduced
  inner width, not the full outer width, or the frame renders 2 columns too
  wide. Caught by `TestChromeNeverOverflowsWidth`.

Deps added: `github.com/charmbracelet/bubbles@v1.0.0`,
`github.com/charmbracelet/bubbletea@v1.3.10`,
`github.com/charmbracelet/x/ansi@v0.11.7` (bump from v0.8.0).

Tests: `chrome_test.go` (golden 80×24 + determinism + overflow guards),
`keymap_test.go` (reflection-driven ctrl+s/ctrl+q guard + arrow-alias
guard), `list_test.go` (listState clamp/move/reset), `primitives_test.go`
(breadcrumb, footer hints, masterDetailWidths, modalBox/rebaseBg,
menuList, placeOverlay/canvasLines). All green, `gofmt`/`go vet` clean,
golden regenerated and re-verified across two clean `go test` runs.

Concern for T8: chrome functions are pure/parameterized by design — the
App-Shell model will need to translate its state into `ChromeOpts` (and into
`breadcrumb`/`footer`/`masterDetailWidths` args directly for non-Chrome
call sites) rather than finding `model`-coupled helpers to call directly.
