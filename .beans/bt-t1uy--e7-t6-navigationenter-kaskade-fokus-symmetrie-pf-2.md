---
# bt-t1uy
title: E7 T6 — Navigation/Enter-Kaskade + Fokus-Symmetrie (PF-2, PF-5, PF-13)
status: completed
type: task
priority: high
created_at: 2026-07-15T14:26:51Z
updated_at: 2026-07-15T17:35:47Z
parent: bt-heg9
blocked_by:
    - bt-kyj5
---

Details/Steps/Akzeptanz: docs/plans/v1-port/epic-E7-plan.md Task 6. Enter-Kaskade INNERHALB Detail-Fokus (Sektion->Feld->Edit-Overlay), tab bleibt einziger Einstieg, NEU shift+tab als deterministischer Ausstieg.

## Summary

PF-13: `tab` keeps its existing bidirectional Tree<->Detail toggle; NEW
`shift+tab` (`FocusOut`) is a deterministic one-way exit (no-op if already in
Tree, no cursor-state reset — only tab-IN resets secCursor/accOpen/
detailLevel/fieldCursor). Both are now typed `keyMap` fields (`FocusIn`/
`FocusOut`, `keymap.go`), replacing the former raw `msg.String()=="tab"`
comparison, at the SAME dispatch point in `handleKey` (still ahead of
keyNodeAction/keyDetailFocus/keyBacklog/keyTree — no new collision).

PF-5: `keyDetailFocus`'s `enter` now cascades. At section level (`detailLevel
== 0`) it is an alias for `→`/`l` — enters field level at `fieldCursor 0`,
guarded by the same "section has fields" condition. At field level
(`detailLevel == 1`) it dispatches by the field's `kind` (already tagged by
T4's `metaFields`/`relationField.kind`): `status`/`type`/`priority` open the
combined Value-Menu seeded on that group (`openValueMenu(group)`, NEW
signature — was hardcoded to `"status"`); `title` opens the same
Title-Edit-Form the `e` key opens; `readonly` (created_at/updated_at) is a
no-op; the default (`""`, Relations section) keeps the UNCHANGED E2 jump
behavior. Design decision (plan Step 7): `m.detailFocus` stays `true` for the
overlay/form dispatch cases (the overlay is a separate capture state on top —
closing it lands the user back on the same field, D02 "schnell/einfach");
only the jump case still exits detail focus (a DIFFERENT bean).

PF-2: the digit-jump range check (`keyDetailFocus`) now compares against the
existing `beanSectionCount` constant (`s[0]-'0' <= byte(beanSectionCount)`)
instead of the hardcoded `'4'` literal — same behavior today (4 sections),
robust to a future 5th section without a second edit site.

`tab` remains the ONLY way to ENTER detail focus (D01 revidiert, PO-Nachtrag
3) — `keyTree`/`keyBacklog`'s own `enter` (leaf no-op / node-with-children
expand-toggle) is untouched; verified live in the tmux smoke below.

## Files

- `internal/tui/keymap.go` — `FocusIn`/`FocusOut` fields + `newKeyMap()` +
  `helpGroups()` "Navigation" entry.
- `internal/tui/update.go` — `handleKey`: `case "tab":` replaced by
  `keybind.Matches(msg, keys.FocusIn/FocusOut)`; `keyDetailFocus`: digit
  check on `beanSectionCount`, NEW section-level enter alias, field-level
  enter now a `switch f.kind` dispatch; `keyNodeAction`'s `s`-key call site
  updated to `m.openValueMenu("status")`.
- `internal/tui/box_menu_value.go` — `openValueMenu(group string) model`
  (was `()`, hardcoded to `"status"`); NEW `currentValueForGroup(b, group)`
  helper (empty-priority-defaults-to-normal, same convention as
  `valueMenuIsCurrent`).
- `internal/tui/overlay_palette.go` — SECOND call site the plan's Step 9 did
  not name (`dispatchPalette`'s `"status"` action, line ~247) — also had to
  move to `m.openValueMenu("status")`, compiler-forced (see Deviations).
- Tests: `internal/tui/keymap_test.go`, `internal/tui/update_test.go`,
  `internal/tui/box_menu_value_test.go`.

## Test-Output

RED (confirmed before touching production code — compile failure from the
new `openValueMenu("type")` call in the not-yet-updated-signature test):

```
internal/tui/box_menu_value_test.go:74:22: too many arguments in call to m.openValueMenu
	have (string)
	want ()
FAIL	beans-tui/internal/tui [build failed]
```

GREEN after implementation:

- `command go build ./...` → clean.
- `command go test ./internal/tui/... -run 'TestFocusInFocusOutKeysBound|TestKeyShiftTab|TestKeyTabStillTogglesBothDirections|TestKeyDetailFocusEnter|TestKeyDetailFocusDigitJump|TestOpenValueMenuSeedsOnGivenGroup' -v` → all 15 new/renamed tests PASS.
- `command go test ./... -short` → PASS (fast loop, ~6s for `internal/tui`).
- `command go test ./...` (FULL run, no `-short`, MANDATORY pre-commit) →
  PASS, `beans-tui/internal/tui  135.442s` (matches the ~135s baseline this
  repo's own CLAUDE.md documents — navigation-only change, no new slow-path
  tests added).
- `command go vet ./...` → clean.
- `command gofmt -l .` → empty.
- `command go test ./internal/tui/ -run "TestTreeGolden|TestBacklogGolden|TestChromeGolden" -count=2 -v` → PASS, both runs byte-stable, **no Golden changed** (Step 11's expectation — pure key-dispatch change, no render-path touched).

## Smoke (Kaskaden-Protokoll, tmux 120x34, this repo's own real `.beans/` data, `bt-apmy`)

1. `tab` → Detail focus (▌ on `[1] META`, ▶ on `title` field marker — Meta's
   own per-row marker already reflects the reset `fieldCursor=0`).
2. `i`/`k` → sections: `k k` moved META(1)→RELATIONS(3) (`▌` bar tracked, RELATIONS'
   own field-strip appeared, confirming section-level i/k unchanged).
3. Digit `2` → BODY opened directly (glamour-rendered body visible).
4. Digit `4` → HISTORY opened directly.
5. Digit `1` → META again, then `enter` → field level (`▶ title`, unchanged
   from a `→`/`l` press — the new alias).
6. `k` (field nav) → `▶ status`, `enter` → Value-Menu opened, `Status` group,
   cursor on `▸ todo (current)` (bean's real status) — `esc` cancelled
   (no mutation fired), landed back on the SAME field (`▶ status`),
   `detailFocus` still true (Detail pane still accented).
7. `shift+tab` → deterministic exit (▌ moved back to the Tree row, Detail
   header lost its `▌`) — verified from mid-field-level state (not just
   section level).
8. `tab` back in → confirmed full reset (`▌ [1] META`, `▶ title`, i.e.
   secCursor/detailLevel/fieldCursor all back to 0 — stale field-level state
   from step 6/7 did NOT leak across re-entry).
9. `→`/`←` pair inside detail focus: `→` on META (section level) → field
   level (same as `enter`); `←` → back to section level (still detail-
   focused); `←` again (at section level) → exits to Tree (`detailFocus`
   false) — confirms the PF-13 pair is fully symmetric AND cleanly
   separated per-context, no `enter`/`tab` involved.
10. Tree-local `→`/`l` and `←`/`j`: with Tree now focused, `→` on `bt-apmy`
    (has children) expanded it (▸→▾, 6 epic children visible) — confirms
    Tree's own left/right (expand/collapse) is completely unaffected by the
    Detail-Focus left/right pair (separate code paths, separate meaning,
    exactly as design-spec §15 PF-13's Kollisionsanalyse predicted).
11. Tree `enter` on the now-expanded `bt-apmy` (node with children): `←`
    first collapsed it, then `enter` re-expanded it (▸→▾) — confirms
    `keyTree`'s own enter (expand/collapse toggle on a node-with-children,
    no-op on a leaf) is untouched; enter never became a second detail-focus
    entry point (D01 revidiert, still holds).
12. Field-level cascade repeated for `title` (→ Title-Edit-Form, same as `e`,
    cancelled via `esc` without saving), `priority` (→ Value-Menu, `Priority`
    group, cursor on `▸ high (current)`, cancelled), `created_at` (→ pure
    no-op: no overlay, no form, state unchanged) — all confirmed live.
13. `q` at the base Tree state (no detail focus, no overlay) → `Quit?`
    confirm modal appeared (`enter: quit  esc: cancel`); cancelled via
    `esc`, app still running; final quit via `q`→`enter` exited cleanly (no
    panic, `tmux has-session` reported the session gone).
14. `git status --short` after the ENTIRE smoke run showed only my own
    source-file edits — `.beans/*.md` untouched, confirming every
    Value-Menu/Title-Form excursion during the smoke was a genuine
    cancel-only dry run, no real mutation ever fired against the live repo.

## Deviations/ERRATA

- **Second `openValueMenu()` call site, not named by the plan.** The plan's
  Step 9 says "Call-Site fixen. `keyNodeAction`: `m.openValueMenu()` →
  `m.openValueMenu(\"status\")`" and design-spec §15 PF-5 says "der einzige
  Call-Site in `keyNodeAction`" — but `overlay_palette.go`'s `dispatchPalette`
  (its own `"status"` palette action, ~line 247) is a SECOND production call
  site, compiler-forced to update once `openValueMenu`'s signature gained the
  `group` parameter (same "Compiler-gesteuerte Vollständigkeit" pattern
  bt-uyzf's own Deviations section names for its 6 `pane.title` call sites).
  Fixed identically (`m.openValueMenu("status")`) — behavior-preserving,
  since the OLD `openValueMenu()` always hardcoded `"status"` regardless of
  caller, so both call sites are semantically unchanged, only the literal
  moved from inside the function to the call site. No new test needed (no
  existing test pinned this call site's group-seeding specifically; the
  existing bean-jump/create-in-flight palette tests are untouched and still
  green).
- **`TestKeyDetailFocusEnterOnRelationFieldStillJumps` reuses existing
  coverage rather than duplicating assertions.** The plan names this test
  explicitly (Step 2); the exact same invariant (relation-field enter still
  jumps + exits detail focus) was ALREADY guarded end-to-end by the
  pre-existing `TestDetailFocusEnterOnRelationJumpsCursorAndExitsToTree` /
  `TestDetailFocusEnterOnUnresolvedRelationIsNoOp` (both green, unmodified).
  Added the plan's named test anyway (pins the SAME invariant under the T6
  name, cheap and explicit) rather than skip it — no scope reduction, just
  flagging the pre-existing overlap for transparency.
- **T5's plan checkboxes were already unchecked before this task** (Task 5,
  `epic-E7-plan.md` lines 471-501) — `bt-uyzf`'s own closing commit
  (`5f2e1fd`) only touched the bean file, not the plan file, unlike T3/T4's
  pattern. Left untouched here (out of T6's scope, not something this task
  introduced) — flagging for the PO/next task to decide whether to
  retroactively flip them.

## Notes for T7 (Header/Footer-Keybinding-Split, PF-11)

Global-vs-local key inventory AFTER this task's refactor (`handleKey`'s
dispatch order, `internal/tui/update.go`):

| Scope | Keys | Dispatch point |
|---|---|---|
| Full-capture (own state machine, no other key reaches through) | search input, filter menu, huh form, node-action overlay (Value-Menu/pickers/confirms), Lobby, Palette, Help | `handleKey`'s first 8 `if`-guards, unchanged by T6 |
| **Global** (single checkpoint, reachable from Tree AND Backlog, before any view-specific dispatch) | `Palette`(ctrl+k/K), `Help`(?), `Picker`(p), `Quit`(q/ctrl+c), **`FocusIn`(tab)/`FocusOut`(shift+tab) — NEW as typed bindings, same dispatch position `tab` already held**, `Refresh`(ctrl+r), then `keyNodeAction`'s own set: `Create`(c), `Status`(s), `TagAssign`(t), `Assign`(a), `Blocking`(B), `Delete`(d), `Editor`(e/ctrl+e), `Yank`(y) | `handleKey` lines ~793-857 |
| **Detail-Focus-local** (`m.detailFocus==true`, `keyDetailFocus`) | `Up`/`Down`(i/k, arrows) = section/field nav, `Left`/`Right`(j/l, arrows) = level nav + exit-to-Tree, **`Enter` = NEW section→field alias AND field→overlay/form/jump dispatch (this task's core deliverable)**, digits `1`-`4` = section jump (now via `beanSectionCount`) | `keyDetailFocus`, entered only via the global `FocusIn` above |
| **Tree-local** (`m.detailFocus==false && m.view==viewBrowseRepo`, `keyTree`) | `Up`/`Down`/`Left`/`Right` = node nav/expand/collapse (UNCHANGED, verified in this task's smoke — no collision with the Detail-Focus pair above, separate code path), `Enter` = expand/collapse toggle (leaf no-op, UNCHANGED — `enter` never became a second detail-focus entry point), `Search`(/), `Filter`(f), `FilterClear`(X), `Backlog`(b) | `keyTree`, untouched by T6 |
| **Backlog-local** (`m.detailFocus==false && m.view==viewBacklog`, `keyBacklog`) | analogous set, `Sort`(S) additionally | `keyBacklog`, untouched by T6 |

Concrete actionable findings for T7, from this refactor:

1. **`FocusIn`/`FocusOut` exist now as real `keybind.Binding`s** (this task's
   deliverable) — ready to feed `renderBindings`. Both `browseRepoChrome`
   (`view_browse_repo.go:702-705`) and `backlogChrome`
   (`view_browse_backlog.go:206-209`) still hand-type a raw `"  tab:focus"`
   string SUFFIX (not going through `renderBindings`/a `keybind.Binding` at
   all) — `shift+tab` currently has ZERO footer visibility anywhere. This is
   exactly the gap design-spec §15 PF-13's closing paragraph names ("den
   heute hand-getippten `'  tab:focus'`-Footer-Suffix ... durch denselben
   `renderBindings`-Mechanismus ... ersetzen").
2. Both Chrome builders' CURRENT `globalHint` is only `{Refresh, Help, Quit}`
   (3 items) — design-spec §15 PF-11's target header list is 7 items
   (`Refresh, Palette, Picker, Help, Back, Enter, Quit`) — `Palette`/
   `Picker`/`Back`/`Enter` are missing from today's header entirely (PO-
   Nachtrag 9's own finding, still true post-T6, T6 did not touch Chrome
   rendering).
3. Both Chrome builders' CURRENT `localHint` includes `Refresh` and `Enter`
   — per PF-11, these move to the (now-7-item) global header list and must
   be REMOVED from `localHint` to avoid duplication (design-spec's own
   explicit call-out).
4. `FocusIn`/`FocusOut` are NOT part of design-spec's global 7-item header
   list — per PF-13's own text they stay footer/view-local (shown via the
   `renderBindings` mechanism now available), so T7 should add them to the
   NEW `localHint`-extraction functions (not to `globalBindings()`) even
   though their DISPATCH position in `handleKey` is already global-style
   (same checkpoint as `Refresh`/`Quit`) — dispatch position and display
   bucket are two different axes, don't conflate them.
5. No render/Chrome file was touched by this task — T7 starts from a clean,
   fully green baseline (goldens unchanged, confirmed above).

Refs: bt-t1uy
