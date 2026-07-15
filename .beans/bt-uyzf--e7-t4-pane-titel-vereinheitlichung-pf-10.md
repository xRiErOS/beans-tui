---
# bt-uyzf
title: E7 T5 — Pane-Titel-Vereinheitlichung (PF-10)
status: completed
type: task
priority: normal
created_at: 2026-07-15T14:26:51Z
updated_at: 2026-07-15T17:10:42Z
parent: bt-heg9
blocked_by:
    - bt-kyj5
---

Details/Steps/Akzeptanz: docs/plans/v1-port/epic-E7-plan.md Task 5. Entfernt redundante Pane-Titel (Tree/Backlog/Detail), Breadcrumb traegt View-Identitaet. Klick-Geometrie (clickPaneGeometry) muss mitziehen.


## Optionaler Mini-Punkt aus T4-Review (I01, trivial)

TestPalFilteredActionsFuzzyGoMatchesAllFourGoToEntries nimmt implizit an, dass kein fixtureBeans()-Titel 'go' fuzzy-matcht — 1 Kommentar-Zeile (oder kind-Assertion) ergänzen, wenn du eh in Test-Dateien bist. Kein Pflichtteil.

## Summary

PF-10: removed the pane-internal title + underline-separator (`renderPane`,
`render_shared.go`) from Tree/Backlog/Detail — the Breadcrumb ("> repo:
Browse"/"Backlog") already carried the view identity redundantly. `pane.title`
field removed entirely (compiler forced call-site cleanup at all 3 production
+ 3 test call sites, per this repo's established "Compiler-gesteuerte
Vollständigkeit" pattern, bt-kyj5 precedent). Review-Cockpit's fourth
call-site did not exist anymore (already removed in an earlier task per
`epic-E7-plan.md`'s own note: "Review-Queue-Pane existiert seit T1 nicht
mehr").

Geometry follow-through (the KRITISCH point from the task brief):
`clickPaneGeometry`'s `originY` loses exactly the 2 `+1` terms that used to
account for the pane's own title+separator lines — its OWN top border is a
DIFFERENT `+1` term and stays (`originY = 1 + head + 1 + 1`, not `1 + head +
1`; caught this off-by-one myself via the render-grounded mouse tests, see
Deviations). Beyond `originY`, the row-WINDOW height derivations
(`treeRows`/`backlogRows`'s own window-height parameter, used identically in
`treeClickRow`/`backlogClickRow`) had to widen from `bodyH-3` to `bodyH-1`:
renderPane's content budget grew by exactly the 2 lines title+separator used
to consume, so the pre-existing `[searchLine, ...treeRows]` budget math
(`bodyH-2` content budget, `bodyH-3` window height) no longer summed to the
pane's real capacity. Left un-adjusted, Tree/Backlog would have rendered 2
trailing blank lines instead of 2 more usable content rows, AND (more
seriously) `treeClickRow`/`backlogClickRow`'s own `windowRows` would have
silently drifted out of sync with what's actually rendered — this is the
exact single-source invariant the file's own doc comments already call out
("the SAME bodyH-N window height treeRows itself windows to").

## Test-Output

RED (both new tests fail against pre-change code, confirmed before touching
production code):
- `command go test ./internal/tui/... -run 'TestRenderPaneNoTitleLine' -v` →
  FAIL (renderPane still emits an empty title line + short separator ahead of
  rows[0]).
- `command go test ./internal/tui/... -run
  'TestClickPaneGeometryOriginYExcludesTitleAndSeparator' -v` → FAIL (originY
  = 6, want 3 against the then-still-old formula's expected value).

GREEN after implementation (`pane.title` removed → compiler forced 6 call-site
fixes: `view_browse_repo.go` ×2, `view_browse_backlog.go` ×1,
`primitives_test.go` ×1, `update_test.go` ×2):
`command go build ./...` clean. `command go test ./internal/tui/... -short` →
initially 5 FAILs: the 2 target tests now GREEN, but 3 pre-existing tests
broke — `TestEmptyFacetMatchViewRendersWithoutPanic` (asserted literal string
`"Tree"` was on screen — exactly the redundant title this task removes, fixed
to assert the search-head glyph instead) and
`TestDoubleClickTogglesExpand`/`TestSingleClickOnOpenNodeDoesNotCollapse`
(render-grounded click tests — these caught a REAL bug: my first `originY` cut
was too aggressive, see Deviations). After the `originY` fix: `command go test
./internal/tui/... -short` → PASS, only `TestTreeGolden`/`TestBacklogGolden`
diff (expected, regenerated below). Full run: `command go test ./...` → PASS,
136.8s (matches the ~135s baseline noted in this repo's `CLAUDE.md`).
`command go test ./... -race -short` → PASS. `command go vet ./...` → clean.
`command gofmt -l .` → empty. `command go test ./internal/tui/... -run
"TestTreeGolden|TestBacklogGolden|TestChromeGolden" -count=2 -v` → PASS, both
runs byte-stable.

## Golden-Diffs

- `tree.golden`: the `"Tree"`/`"Detail"` title line + its underline-separator
  (2 lines) removed from the top of both panes — first content row is now
  `"⌕ / search"` (Tree) / the Kopfblock ID (Detail) directly. 2 more blank
  lines appear at the pane's bottom (the freed content budget — this fixture's
  6 nodes already fit inside the OLD window too, so no new node rows become
  visible, just 2 more blank padding rows; confirms `bodyH-1` is exactly
  filling the pane's real capacity, no overflow/underflow).
- `backlog.golden`: identical pattern (`"Backlog"`/`"Detail"` title+separator
  removed, 2 more blank lines at the bottom).
- `chrome.golden`: UNCHANGED (`git diff --stat` shows no entry) — `Chrome()`
  never called `renderPane`, exactly as the plan predicted.

## Smoke

`command go build -o bin/bt .` then tmux (this repo's own real `.beans/` data
as source, 120x34):
- Browse: first pane row is `"⌕ / search"` directly, no `"Tree"` text
  anywhere on screen (verified via `grep -E "^\s*.{0,3}(Tree|Detail)\b"` over
  the full `tmux capture-pane` output → no match). Detail pane's first row is
  the bean ID (`bt-apmy`) directly, no `"Detail"` text.
- Backlog (`b` key): same pattern, breadcrumb shows `"Backlog"`, pane content
  starts at the search line.
- **Klick-Geometrie-Beleg (PF-10's critical risk point):** sent a real SGR
  mouse click (`ESC[<0;10;7M` press + `ESC[<0;10;7m` release via `tmux
  send-keys -H`) at the screen row showing the `bt-6oyy` tree node. Before:
  Detail pane showed `bt-apmy`. After: Detail pane switched to `bt-6oyy` /
  "Tag-Management-Page (zentrale Tag-Definition)" — cursor bar (`▌`) moved to
  the clicked row, confirming `originY`/`windowRows` stayed in sync with the
  real render after the 2-line shift.
- Wheel: `ESC[<64;10;7M` (wheel up) moved the cursor from `bt-6oyy` back to
  `bt-apmy` — confirmed via before/after `bt-*` ID in the Detail header.
- Doppelklick: two press+release pairs ~100-200ms apart on `bt-apmy`'s tree
  row toggled its expand marker `▸`→`▾` (children visible, e.g. `bt-blsy`)
  then `▾`→`▸` again (collapsed) — devd D03 semantics intact through the
  geometry change.

## Deviations

- **`originY` off-by-one, self-caught via the render-grounded mouse tests:**
  my first cut of `clickPaneGeometry`'s `originY` removed THREE `+1` terms
  (`1 + head + 1` — outer border + divider only) instead of the correct TWO
  (title+separator). The pane's OWN top border is a separate `+1` term from
  title/separator and must stay (`1 + head + 1 + 1`). My own new geometry
  unit test (`TestClickPaneGeometryOriginYExcludesTitleAndSeparator`) did NOT
  catch this — it was self-referentially derived from the same wrong formula.
  What DID catch it: the pre-existing render-grounded
  `TestDoubleClickTogglesExpand`/`TestSingleClickOnOpenNodeDoesNotCollapse`
  (mouse_test.go's own doc comment: "a click coordinate is found by rendering
  the REAL View() ... never hand-computed against the click formula itself
  (that would be circular)") — exactly the safety net the task brief called
  out. Fixed both the production formula and my own unit test's expectation
  together once the bug was located.
- **`bodyH-3` → `bodyH-1` window-height widening (Tree/Backlog render AND
  click-row mapping, 4 call sites):** design-spec.md §15's own "Geometrie-
  Folge" paragraph only names `originY`'s 2 lost `+1` terms explicitly; it
  does not spell out that the `treeRows`/`backlogRows` window-height
  parameter (and its `treeClickRow`/`backlogClickRow` mirror) must widen too.
  I widened it anyway — required by this task's own brief ("windowStart-
  bodyH-Ableitungen" must be pulled in sync) and by the repo's own single-
  source doc-comment invariant ("the SAME bodyH-N window height treeRows
  itself windows to"): leaving it at `bodyH-3` would have rendered 2 wasted
  blank lines per pane instead of using the freed budget, contradicting the
  task brief's own "mehr bodyH → evtl. mehr sichtbare Zeilen" expectation.
  Golden diffs above show the widened budget as 2 more blank padding lines in
  THIS fixture (not more real rows, since both fixtures already fit inside the
  old window) — a taller node list would show 2 more real rows instead.
- **`TestEmptyFacetMatchViewRendersWithoutPanic` assertion swap:** this
  pre-existing test asserted the literal string `"Tree"` appeared on screen —
  exactly the redundant title PF-10 removes. Swapped to assert `searchShield`
  (the `"⌕"` glyph, the pane's own still-rendered, non-Breadcrumb signal that
  its frame survived a zero-match render) instead, same intent (panic-guard +
  frame-survives assertion), new signal.
- **Mini-Punkt aus T4-Review (I01, optional, done since already in test
  files):** added a comment line to
  `TestPalFilteredActionsFuzzyGoMatchesAllFourGoToEntries`
  (`overlay_palette_test.go`) documenting that none of `fixtureBeans`' titles
  contain a `'g'` at all, so no bean item could ever fuzzy-match `"go"` and
  silently inflate the asserted 4-item count — no test-behavior change, purely
  documentary per the bean's own "1 Kommentar-Zeile ... kein Pflichtteil".

## Notes for T6 (Navigation/Enter-Kaskade + Fokus-Symmetrie)

- The Tree/Backlog pane's first VISIBLE row is now ALWAYS the search/filter
  head line (`treeSearchLine`) directly under the pane's own top border — no
  title/separator buffer above it anymore. Any T6 enter-cascade/digit-jump
  work that reasons about "row 0 inside the pane" should use this directly
  (`clickRow == 0` in `treeClickRow`/`backlogClickRow` already encodes this:
  "row 0 == the search head line").
- `pane.title` is GONE from the `pane` struct (not deprecated, not a no-op
  field) — if any future task needs a pane-scoped label again, it needs a
  fresh design decision (PO explicitly ruled redundant title+Breadcrumb out,
  not just "hide it conditionally").
- Detail's Kopfblock (T4, bt-kyj5) needed ZERO changes here — it lives
  entirely inside `renderAccordionPane`'s own `rows` slice, downstream of
  `renderPane`'s line-cap, exactly as bt-kyj5's own "Notes for T5" predicted.
  Confirms the Kopfblock/PF-3/PF-4 area is fully decoupled from pane-chrome
  geometry — safe ground for T6's enter-cascade/digit-jump work on the
  Meta-Feldliste without any interaction with THIS task's geometry change.
- Fokus-Symmetrie (PF-13, shift+tab / arrow-left|right pairing) and the
  enter-Kaskade (PF-2/PF-5) touch `update.go`/`keymap.go`, not any file this
  task changed — no drift risk observed, but flagging since T6 will likely
  re-derive `windowRows`/`bodyH` values in its own reasoning about which row a
  digit-jump lands on; the CURRENT values are `bodyH-1` (Tree/Backlog window
  height) and `bodyH` (Detail's full `renderPane` budget, unwindowed — Detail
  has no fixed row-window, it relies purely on `renderPane`'s line cap).
- Mouse-click regression coverage for the geometry change lived entirely in
  ALREADY-EXISTING render-grounded tests (`mouse_test.go`'s
  `leftPaneClickAt`/`treeClickAt` helpers, which locate a substring on the
  REAL rendered `View()` rather than hand-computing a Y coordinate) — this
  pattern caught my own `originY` bug when my own new unit test could not
  (see Deviations). Any T6 mouse/geometry work should keep relying on this
  pattern rather than adding hand-computed-coordinate tests.

Refs: bt-uyzf, Commit 6849bdf
