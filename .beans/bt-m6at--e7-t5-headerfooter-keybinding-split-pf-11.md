---
# bt-m6at
title: E7 T7 — Header/Footer-Keybinding-Split (PF-11)
status: completed
type: task
priority: normal
created_at: 2026-07-15T14:26:51Z
updated_at: 2026-07-15T18:18:34Z
parent: bt-heg9
blocked_by:
    - bt-kyj5
    - bt-t1uy
---

Details/Steps/Akzeptanz: docs/plans/v1-port/epic-E7-plan.md Task 7. Header zeigt ALLE 7 globalen Bindings, Footer wird kontextsensitiv (view-lokal vs. overlay-/form-lokal, Q04). Braucht T6s FocusIn/FocusOut.

## Summary

PF-11: Header Zone 1 (`browseRepoChrome`/`backlogChrome`, the only two Chrome-calling views left post-PF-14) now shows ALL 7 global bindings via a new single source `globalBindings()` (keymap.go) — `{Refresh, Palette, Picker, Help, Back, Enter, Quit}`, rendered `ctrl+r:reload · ctrl+k:commands · p:repos · ?:help · esc:back · enter:open/confirm · q:quit`. Refresh/Palette/Picker's `Help().Desc` were shortened ("reload"/"commands"/"repos", Planner-Entscheidung "kurz und konsistent") — `helpGroups()`/the Help-Overlay pick this up for free (Single Source, same `keybind.Binding` objects).

Footer Zone 3 now shows ONLY view-local bindings, extracted into named `browseRepoLocalBindings()`/`backlogLocalBindings()` (their previous inline `[]keybind.Binding{...}` literal, minus `Refresh`/`Enter` — now header-only — plus `FocusIn`/`FocusOut` replacing the hand-typed `"  tab:focus"` suffix; `shift+tab` is visible in a footer for the first time). New Drift-Guard `TestNoDuplicateBindingBetweenGlobalAndAnyLocalHintList` (keymap_test.go) reflects over both lists and fails if any binding ever reappears in `globalBindings()`.

Q04-Antwort (kontextsensitiver Footer): new `internal/tui/footer_context.go`, `func (m model) contextualLocalHint(viewLocal []keybind.Binding) string` — priority `m.filterOpen > m.overlay != overlayNone > m.searchActive > m.paletteOpen > m.helpOpen > view-local default`, mirroring `handleKey`'s own full-capture dispatch order (update.go). Both Chrome functions now call `footer(m.contextualLocalHint(<view>LocalBindings()), innerW)` instead of a static string. `m.form != nil` (huh Forms) and `m.confirmQuit` are deliberately NOT cases here — `formChrome`/`quitBox` already bake a complete hint into their own `modalPanel` footer argument; there is no base-view fallback to build for either.

## Files

- `internal/tui/keymap.go` — Help-text shortening (Refresh/Palette/Picker); new `globalBindings()`.
- `internal/tui/footer_context.go` (NEW) — `contextualLocalHint` + one binding-list helper per overlay/context kind.
- `internal/tui/view_browse_repo.go` — `browseRepoChrome` rewired to `globalBindings()`/`contextualLocalHint`; new `browseRepoLocalBindings()`.
- `internal/tui/view_browse_backlog.go` — `backlogChrome` rewired identically; new `backlogLocalBindings()`.
- Tests: `internal/tui/keymap_test.go` (+3), `internal/tui/footer_context_test.go` (NEW, 12 tests), `internal/tui/view_browse_repo_test.go` (NEW, 4 tests), `internal/tui/view_browse_backlog_test.go` (+4 tests).
- Goldens: `internal/tui/testdata/tree.golden`, `internal/tui/testdata/backlog.golden` (regenerated). `internal/tui/testdata/chrome.golden` confirmed UNCHANGED (see Deviations).
- Plan: `docs/plans/v1-port/epic-E7-plan.md` Task 7 — all Steps + Akzeptanz-Checkliste ticked, with inline ERRATUM/DEVIATION notes at Step 2/5/6/8 (mirrors this bean's own Deviations, not duplicated here twice).

## Test-Output

RED (confirmed before touching production code):

```
vet: internal/tui/footer_context_test.go:27:11: m.contextualLocalHint undefined (type model has no field or method contextualLocalHint)
```

GREEN after implementation:

- `command go vet ./internal/tui/...` → clean.
- `command go test ./internal/tui/ -run "TestGlobalBindings|TestNoDuplicateBinding|TestContextualLocalHint|TestBrowseRepoChrome|TestBacklogChrome|TestFocusInFocusOut|TestHelpGroupsCoverEveryBindingExactlyOnce" -v` → all 23 matched tests PASS (`TestGlobalBindingHelpTextsShortened` verified separately, its name doesn't match the "Bindings" substring in that filter — also PASS).
- `command go test ./... -short` → PASS (`beans-tui/internal/tui 5.754s`).
- `command go test ./...` (FULL run, no `-short`, MANDATORY pre-commit) → PASS, `beans-tui/internal/tui 136.062s` (matches this repo's own ~135s documented baseline — no new slow-path tests added).
- `command go vet ./...` → clean.
- `command gofmt -l .` → empty.
- `command go test ./internal/tui/ -run "TestTreeGolden$|TestBacklogGolden$|TestChromeGolden$" -count=2 -v` → PASS, all three byte-stable across both runs.
- `command go build -o bin/bt .` → clean.

## Golden-Diffs

`chrome.golden`: **unchanged**, confirmed via a plain (non-`-update`) run before touching any golden — `chrome_test.go`'s `goldenChromeOpts()` hand-writes its own `GlobalHint`/`FooterHint` strings, never calling `globalBindings()`/`browseRepoChrome`/`backlogChrome` — nothing in this task's diff reaches that test.

`tree.golden` / `backlog.golden` (both regenerated with `-update`, both width 100):

```diff
 ╭──────────────────────────────────────────────────────────────────────────────────────────────────╮
-│> bt-golden-repo: Browse                                        ctrl+r:Reload data  ?:help  q:quit│
+│> bt-golden-repo: Browse                                                                          │
+│ctrl+r:reload  ctrl+k:commands  p:repos  ?:help  esc:back  enter:open/confirm  q:quit             │
 │──────────────────────────────────────────────────────────────────────────────────────────────────│
 ...(one blank detail-pane row absorbed -- the Detail body budget shrinks by exactly the +1 header line)...
-│↑/i:up  ↓/k:down  ←/j:back/out  →/l:in/expand  enter:open/confirm  /:Search  ctrl+r:Reload data   │
-│s:Status menu  c:Create  d:Delete  e:Edit in $EDITOR  tab:focus                                   │
+│↑/i:up  ↓/k:down  ←/j:back/out  →/l:in/expand  /:Search  s:Status menu  c:Create  d:Delete  e:Edit│
+│in $EDITOR  tab:focus in/toggle  shift+tab:focus out                                              │
```

(`backlog.golden`'s diff is the structurally identical shape, `Backlog` title + its own local set instead.)

Cause of the header going 1→2 lines at width 100 (innerW 98): the OLD 3-item global hint ("ctrl+r:Reload data  ?:help  q:quit", ~35 chars) always fit next to the left breadcrumb (~24 chars) inside 98 columns; the NEW 7-item hint is 85 chars — 85+24+1 gap = 110 > 98, so `breadcrumb()`'s existing narrow-width fallback (`view.go`, already tested by `TestChromeNeverOverflowsWidth`) stacks left/right onto two lines instead of overflowing. This is EXISTING, tested machinery, not new code — `clickPaneGeometry` already derives `bodyH`/`originY` from `lipgloss.Height(head)` dynamically (mouse.go), so the extra line is absorbed correctly everywhere (mouse click mapping included), just at the cost of one row of Detail-pane content at this width. The footer's own 2-line wrap is PRE-EXISTING (already present in both OLD goldens) — the new footer content (151 chars) is actually marginally SHORTER than the old one (160 chars, computed incl. the hand-typed `"  tab:focus"` suffix), so this is not a new wrap, same wrap.

At the smoke-tested 120-column terminal the header fits on ONE line (85+~30 breadcrumb ≈ 115 ≤ 118 innerW) — the stacking is a medium-width (~90–108 col) phenomenon only. See Smoke below for the 80-column finding (I01).

## Smoke (tmux 120×34 then 80×30, this repo's own real `.beans/` data, bt-apmy)

1. Start (Browse, 120×34): header ONE line, all 7 globals visible in order (`ctrl+r:reload  ctrl+k:commands  p:repos  ?:help  esc:back  enter:open/confirm  q:quit`); footer TWO lines, view-local only (`↑/i:up  ↓/k:down  ←/j:back/out  →/l:in/expand  /:Search  s:Status menu  c:Create  d:Delete  e:Edit in $EDITOR` / `tab:focus in/toggle  shift+tab:focus out`) — **duplicate-check**: manually cross-referenced every header token against every footer token, zero overlap.
2. `f` (Filter-Menu open): footer switched to `↑/i:up  ↓/k:down  space/x:Toggle facet  ...esc:back` (visible around the modal's left/right edges) — the modal's OWN inline hint (`space/x:toggle X:clear enter/esc/f:done`) also rendered at the top of its body, confirming BOTH surfaces now agree (Q04) instead of the outer footer staying stale.
3. `esc` closed filter; `s` (Value-Menu open on bt-apmy, Status group): footer switched to `↑/i:up  ↓/k:down  enter:open/confirm  s:Status menu  esc:back` — matches `valueMenuLocalBindings`; `esc` cancelled (no mutation).
4. `c` (Create-Form open, huh): footer stayed on the BASE view-local set (`m.form != nil` deliberately not a `contextualLocalHint` case, doc-stamp above) — the form's own baked-in `"enter next/save · esc cancel"` hint is the primary surface inside the modal. `esc` aborted the form.
5. `p` (Lobby): confirmed completely untouched by this task — its own hand-typed `"i/k:↑↓  enter:open  type:filter  esc/q:back"` hint unchanged (Lobby is not one of the two Chrome-calling views, out of T7's Files list). `esc` returned to Browse.
6. `b` (Backlog): header ONE line (all 7 globals, 120 cols), footer `↑/i:up  ↓/k:down  S:Sort  /:Search  f:Filter  b:Backlog  s:Status menu  c:Create  d:Delete  e:Edit in $EDITOR` / `tab:focus in/toggle  shift+tab:focus out` — zero overlap with header, confirmed.
7. **80-column resize** (`tmux resize-window -x 80`, still Browse): header STACKS (as tree.golden's own diff predicts at ~100 cols) but ALSO gets `ansi.Truncate`'d — `ctrl+r:reload  ctrl+k:commands  p:repos  ?:help  esc:back  enter:open/confirm…` with `q:quit` cut off entirely and invisible. **I01 finding, NOT fixed in this task** (see Deviations) — `q` still functionally quits (handled at `handleKey`'s `"q"` case regardless of header visibility), this is a discoverability/visibility regression only, at a width narrower than either golden tests.
8. `q` → confirm-quit → `enter` → clean exit (`tmux has-session` reports the session gone, no panic).
9. `git status --short .beans/` after the ENTIRE smoke run: empty — every Value-Menu/Filter/Create-Form excursion above was a genuine cancel-only dry run, no live mutation ever fired.

## Deviations

| Code | Severity/Prio | Beschreibung | Empfehlung | Status |
|------|------|------|------|--------|
| D01 | — | Implementer-Auftrag paraphrasierte die lokale Footer-Liste breiter (`f/X/b/t/a/B/y` zusätzlich zu Up/Down/Left/Right/tab/shift-tab/Search/Status/Create/Delete/Editor/Yank) als epic-E7-plan.md Task 7 Step 5 ("ihre BISHERIGE Liste") + bt-t1uy's eigene Notes-for-T7 (nur 3 konkrete Baustellen: tab-Suffix, Header 3→7, Refresh/Enter raus) es fordern. **Entscheidung: der PLAN-Wortlaut wurde befolgt**, nicht die breitere Paraphrase — eine Erweiterung hätte PF-11s eigene "Footer wird kürzer"-Begründung (VQA-I01-Entschärfung) unterlaufen bzw. umgekehrt, UND ist über 3 unabhängige Primärquellen (Plan, T6-Notes, Ist-Code) hinweg NICHT belegt. `f/X/b/t/a/B/y` sind seit vor diesem Task bereits real aktive, aber im Footer nie sichtbare Bindings (pre-existing gap) — nicht durch T7 verursacht oder in dessen Files-Scope. | PO-Entscheidung, ob ein eigenständiges Folge-bean für "Footer-Vollständigkeit" gewünscht ist. | 🟣 Offen |
| D02 | — | `overlayCreateConfirm`/`overlayDeleteConfirm` (2 von 6 `overlayID`-Werten) waren vom Plan-Step-6-Text NICHT mit einem eigenen Footer-Set benannt (nur Value-Menu + Tag-/Parent-/Blocking-Picker). Gap-Fill: beide bekommen `{Enter,Back}` (deckt sich mit ihren real einzigen Tasten, `keyCreateConfirm`/`keyDeleteConfirm`). | — | 🟢 Erledigt (in diesem Task gefixt) |
| D03 | — | Tag-/Blocking-Picker-Footer bekommen zusätzlich `Toggle` (Plan-Text nannte nur `{Up,Down,Enter,Back}` für alle drei Picker gemeinsam) — `keyTagPicker`/`keyBlockingPicker` verdrahten `keys.Toggle` (space/x) real (Multi-Select), Parent-Picker NICHT (Single-Select, bleibt Toggle-frei). Auslassen hätte Q04s eigene allgemeine Formulierung ("wenn ein Form/Overlay aktiv ist ... inkl. 'space: select/toggle'") nur am namentlich genannten Beispiel (Filter-Menü) erfüllt, nicht an den beiden anderen real betroffenen Overlays. | — | 🟢 Erledigt (in diesem Task gefixt) |
| I01 | medium | Bei ≤80 Spalten wird der Header-Hint (85 Zeichen, 7 Bindings) selbst im gestapelten 2-Zeilen-Fallback zu breit und via `ansi.Truncate` beschnitten — `q:quit` verschwindet komplett aus der sichtbaren Zeile (Smoke Schritt 7). `q` funktioniert weiterhin (Dispatch unabhängig von der Anzeige), nur nicht mehr sichtbar/auffindbar. Design-spec/Plan nennen keine Kürzungs-/Responsive-Truncation-Strategie für diesen Fall (nur die bereits umgesetzte Label-Kürzung reload/commands/repos) — bewusst NICHT selbst entschieden (wäre Raten über Prioritätsreihenfolge, welche Bindings zuerst weichen). | PO-Entscheidung: (a) akzeptieren (80 Spalten ist schmaler als beide Goldens/Design-Referenzbreiten), (b) responsive Truncation analog devd, oder (c) andere Kürzungsstrategie. | 🟣 Offen |
| I02 | low | Backlog-Footer behält seine VOR-T7-Asymmetrie ggü. Tree (z.B. `Filter`/`Backlog` waren schon vor diesem Task nur im Backlog-Footer sichtbar, nicht im Tree-Footer, obwohl beide Views `f`/`b` gleich behandeln) — bewusst NICHT angeglichen (out of scope, siehe D01). | Im selben Zug wie D01 lösen, falls PO Footer-Vollständigkeit priorisiert. | 🟣 Offen |

## Notes for T8 (Abschluss)

1. **Voller Regressionslauf bereits grün** (dieser Task hat ihn schon gefahren): `command go build -o bin/bt .`, `command go test ./...` (136s, kein `-race` bisher — T8 Step 1 verlangt zusätzlich `-race`, das NICHT Teil dieses Tasks war), `command go test ./... -short` (2×), `gofmt -l .` leer, `go vet ./...` leer. T8 sollte `-race` als NEUEN Schritt behandeln, nicht als Wiederholung.
2. **E6-Blocking-Verifikation (T8 Step 2)**: unverändert von T6 übernommen, dieser Task hat `bt-wm4w`/`bt-9yvh`s `blocked_by` nicht berührt.
3. **Offene D/I-Punkte (D01/I01/I02 oben) sind KEINE Blocker für den Epic-Abschluss** — PF-11s Akzeptanzkriterien (Header exakt 7 Bindings, Disjunktheit, kontextsensitiver Footer inkl. Q04) sind alle grün. D01/I01/I02 sind bewusste Scope-Grenzen bzw. eine dokumentierte Breiten-Grenze, keine Regressionen — für den PO-Review-Durchlauf (`bt-heg9` → `to-review`) sichtbar machen, nicht stillschweigend weiterreichen.
4. Commit dieses Tasks: `7d83e5e` (`feat(tui): Header/Footer-Keybinding-Split, kontextsensitiver Footer (PF-11)`).

Refs: bt-m6at
