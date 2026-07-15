# beans-tui (`bt`)

PO-cockpit TUI for beans repos — a port of the DevDash TUI (`dd`) onto the
[beans](https://github.com/hmans/beans) framework. Design/architecture:
[`docs/plans/v1-port/design-spec.md`](docs/plans/v1-port/design-spec.md).

## Status

E1 (Foundation), E2 (Browse & Detail), E3 (Mutations), E4 (Command-Center) and
E5 (Polish) are done: read-only Tree over the beans data layer (Milestones →
Epics → Tasks) with live-reload via an fsnotify watcher, quit-confirm (E1);
master-detail focus with a Detail Accordion (Meta/Body/Relations/History,
relation-jump), local live search + Bleve from 3 characters, facet filter
(Status/Type/Priority/Tag, shared across Tree AND Backlog) and the Backlog
view with sort-toggle (E2); full mutation wiring — combined Status/Type/
Priority menu, Tag-/Parent-/Blocking-Picker, Create-Form (huh, confirm-gate),
title/body edit (`$EDITOR`) and delete-confirm with children-/link-warning,
throughout with ETag-conflict handling (E3); Command-Center (`ctrl+k`, fuzzy
actions + bean search mixed, context-first) (E4); Toast system (incl. sticky
conflict), Help-Overlay `?`, Yank `y` (OSC52+native, bean/epic context),
mouse (wheel/click/double-click), Settings (`~/.config/beans-tui/`), Lobby V1
+ Repo-Picker `p` (watcher lifecycle switch) and the Archive view
(completed/scrapped default-off, togglable) (E5).

E7 (PO-Feedback R1: Detail-UX + Type/Status/Priority glyphs) is done: the
Review-Cockpit was removed (PF-14, see below), Type/Status/Priority now
render as single colored letters/glyphs instead of words or shape icons
(PF-6, see the legend below), all user-facing strings are English (PF-7) with
a `verb entity` Command-Center schema (PF-8), the Detail pane now opens with
a header block (bean-id / title / `type: … status: … prio: …`) followed by an
editable Meta field list with a permanently reserved cursor gutter (PF-1,
PF-3, PF-4, PF-12), redundant pane titles were removed (PF-10), the Header/
Footer keybinding split is complete with no duplication (PF-11), and
Detail-Focus gained an enter-cascade (section → field → edit-overlay) plus
symmetric `tab`/`shift+tab` focus pairing (PF-2, PF-5, PF-13). E6
(Validierung & Release) is done: all 14 v1 user stories are validated
against `design-spec.md` §10 — 13 PASS, 1 PARTIAL (US-08, tag visibility —
see Known Issues below). Full evidence trail, the smoke-test appendix and
the table of 8 open PO decisions (D01–D08) live in
[`docs/plans/v1-port/validation.md`](docs/plans/v1-port/validation.md).

**Review happens in the chat, not in the TUI.** A former Review-Cockpit view
(`R`) was removed per PO decision (PF-14, E7 T1, 2026-07-15 — "contradicts
the lean-stack spirit and reintroduces ceremony"). The TUI shows review state
only as ordinary tag visibility: the tag trio `to-review` (agent reports
done) → `accepted`/`rejected` (PO decides in chat or via `beans update
--tag`), discoverable like any other tag via Tree/Detail/Filter/Search — no
dedicated TUI interaction for it.

## Glyph legend

Type, Status and Priority render as a single colored letter/glyph each
(redundant encoding: color AND shape/letter, for accessibility) — PF-6.

| Type | Glyph | Color |
|---|---|---|
| milestone | `M` | blue |
| epic | `E` | mauve |
| feature | `F` | mauve |
| task | `T` | sky |
| bug | `B` | red |

| Status | Glyph | Color |
|---|---|---|
| draft | `d` | blue |
| todo | `t` | green |
| in-progress | `i` | yellow |
| completed | `c` | subtext (muted) |
| scrapped | `s` | subtext (muted) |

| Priority | Glyph | Color |
|---|---|---|
| critical | `‼` | red, bold |
| high | `!` | yellow, bold |
| normal | `·` | text |
| low | `↓` | subtext (muted) |
| deferred | `→` | subtext (muted) |

Unknown/future enum values fall back to a neutral `·` glyph in text color
rather than disappearing or signalling incorrectly. Set `BT_ASCII_ICONS=1`
for an ASCII-only fallback set on terminals without EAW-neutral Unicode
support (affects Priority only — Type/Status letters are already ASCII).

## Prerequisites

- beans CLI ≥ 0.4.2 on `PATH`
- Go 1.26+

## Installation

```sh
make build          # → bin/bt
# or
command go install .
```

## Start

```sh
bt          # searches .beans.yml upward from cwd
bt <path>   # explicit repo
```

## Keybindings

The Header (top row) always shows the same 7 globally-reachable bindings, no
matter which view is active. The Footer (bottom row) is context-sensitive: it
shows the bindings local to whatever currently has full input focus — the
active view when nothing else is open, or the active overlay/form/menu's own
bindings the instant one opens (so the footer never shows a stale hint for a
key that isn't live). No binding ever appears in both places at once.

### Header (global, always visible)

| Key | Action |
|---|---|
| `ctrl+r` | Reload data |
| `ctrl+k` / `K` | Open Command-Center |
| `p` | Open repo-picker/Lobby |
| `?` | Help-Overlay |
| `esc` | Back |
| `enter` | Open/confirm |
| `q` / `ctrl+c` | Quit (confirm) / immediate quit |

### Footer — Browse (Tree + Detail-Focus)

| Key | Action |
|---|---|
| `↑`/`i`, `↓`/`k` | Cursor (Tree) / Section- or field-cursor (Detail-Focus) |
| `→`/`l` | Expand node / Detail-Focus: descend a section into its field list |
| `←`/`j` | Collapse node / Detail-Focus: leave field list, then leave Detail-Focus |
| `/` | Search — local live filter (title), from 3 characters also Bleve (title+body) |
| `s` | Status/Type/Priority menu (combined, one key for all three) |
| `c` | Create bean (huh form, confirm-gate) |
| `d` | Delete (confirm, children-/link-warning — no cascade) |
| `e` | Edit body in `$EDITOR` (`$VISUAL` → `$EDITOR` → `vi`, Settings `editor` wins over both — see Settings below) |
| `tab` | Focus toggle Tree ↔ Detail-Accordion |
| `shift+tab` | Focus back to Tree (one-way, no-op if already in Tree) |
| `1`–`4` | Detail-Focus: direct section jump (Meta/Body/Relations/History) |

`f`/`X`/`b`/`t`/`a`/`B`/`y` (Filter/Clear-Filter/Backlog/Tag-Picker/
Parent-Picker/Blocking-Picker/Yank) are reachable from Browse but are not
listed in this footer — a known, deliberately narrow footer scope (see Known
Issues below); they remain in the Help-Overlay (`?`).

**Detail-Focus enter-cascade** (`tab` is the only way IN — this does not
change): with a section highlighted, `enter` (alias: `→`/`l`) descends into
its field list; with a field highlighted, `enter` opens that field's
edit-overlay (status/type/priority → the combined Value-Menu; title → the
Title-Edit-Form; `created_at`/`updated_at` are read-only, a no-op) — or, on a
Relations field, jumps the Tree cursor to that related bean and returns focus
to Tree. `1`–`9` always jump directly to a section (Meta is always expanded
and non-collapsible — PF-1).

### Footer — Backlog (`b`)

| Key | Action |
|---|---|
| `↑`/`i`, `↓`/`k` | Cursor |
| `S` | Sort-toggle, cycles status → priority → created → updated |
| `/` | Search |
| `f` | Facet filter |
| `b` | Back to Browse |
| `s` | Status/Type/Priority menu |
| `c` | Create bean |
| `d` | Delete |
| `e` | Edit body in `$EDITOR` |
| `tab` | Focus toggle Backlog-List ↔ Detail-Accordion |
| `shift+tab` | Focus back to the list |

`enter` on a Backlog row is a deliberate no-op — `tab` is the only entry into
Detail-Focus (same as Browse).

### Footer — Filter-Menu (`f`)

`↑`/`↓` move · `space`/`x` toggle facet · `X` clear filters · `enter`/`esc`/
`f` close. Facets: Status, Type, Priority, Tags, Archive ("Show archived" —
completed/scrapped are hidden by default).

### Footer — Value-Menu / Tag-/Parent-/Blocking-Picker (opened via `s`/`t`/`a`/`B`, or the enter-cascade)

`↑`/`↓` move · `enter` apply/save · `esc` cancel/discard · Tag-Picker/
Filter-Menu additionally: `space`/`x` toggle (multi-select).

### Mouse

| Action | Behavior |
|---|---|
| Wheel | Moves the cursor of the active view (Tree/Backlog — no scroll offset, the cursor follows the render) |
| Click | Sets the cursor to the clicked row; a click on a closed, expandable Tree node expands it directly |
| Double-click | On an already-open, expandable Tree node (second click <500ms) collapses it — a single click on an open node does NOT collapse it, only moves the cursor (devd D03 semantics) |
| Click on a toast | Dismisses it immediately, takes priority over any open form/overlay |

### Review (tag trio, in chat)

No dedicated TUI view (PF-14, see above). Review happens entirely outside
the TUI:

| Step | Who | beans operation |
|---|---|---|
| Work done, review requested | Agent | Set tag `to-review`, status stays `in-progress` |
| Review visibility | PO (TUI, passive) | Beans tagged `to-review` appear like any other tag in Tree/Detail, discoverable via filter/search |
| Pass | PO (chat/CLI) | Tag `to-review` → `accepted` (`beans update --tag`) |
| Reject | PO (chat/CLI) | Tag `to-review` → `rejected`; feedback lands in chat |
| Rework done | Agent | Tag `rejected` → `to-review` |

## Settings

Configuration lives under `~/.config/beans-tui/`:

- `config.yaml` — `repos:` (list of repo paths for the Lobby), `editor:`
  (empty = `$VISUAL` → `$EDITOR` → `vi`; if set, ALWAYS wins over both
  environment variables), `theme.accent:` (hex `#rrggbb`, empty = built-in
  mauve), `layout.tree_width:` (tree-width floor, 24–60).
- `state.json` — runtime state, currently only the last-opened repo
  (persisted on every repo switch, see Lobby below).

Reachable via Command-Center (`ctrl+k` → "go to settings"): editor/accent/
tree-width take effect IMMEDIATELY on save (no restart needed), `repos`
takes effect the next time the Lobby (`p`) opens.

## Lobby + Repo-Picker (`p`)

`p` opens the Lobby from anywhere: a search field + the list of repos
configured in `config.yaml` (shows "open/total" per repo, resolved
asynchronously). `enter` switches the repo — the old fsnotify watcher is
stopped, a new one started for the new repo (file changes in the previous
repo no longer trigger a reload afterwards), the last-chosen repo lands in
`state.json`.

**Start trigger:** an explicit `bt <path>` argument always wins; otherwise
straight into the repo if `.beans.yml` is found upward from cwd (unchanged
E1 behavior); only if both fail AND `repos:` has at least 2 entries does the
Lobby open on start — with 0/1 configured repos the previous error message/
direct-start behavior stays.

## Known Issues

See [`docs/plans/v1-port/validation.md`](docs/plans/v1-port/validation.md)
for the full v1 acceptance record, including a table of 8 open PO decisions
(D01–D08) — the three below carry a direct cross-reference; the remaining
five (Backlog sort-mode indicator, `esc` in Detail-Focus, upstream ETag
drift on fresh creates, and the Tag-Management-Page scope question) are
tracked there only, not duplicated here.

- **Tag-trio/tag visibility gap (open PO decision, validation.md D01):**
  tags — including the review tag-trio `to-review`/`accepted`/`rejected` —
  are not shown passively in the Tree or in the Detail Meta list; the facet
  filter (`f`) is currently the only working path of the four the design
  spec names (Tree/Detail/Filter/Search). Filtering and live-reload both
  work correctly. See `bt-gdkx`.
- **Pickers deliberately show every valid relation target:** the Parent-
  Picker (`a`) and Blocking-Picker (`B`) do NOT filter by status/archive
  visibility — archived/`completed`/`scrapped` beans stay selectable as
  relation targets as long as they're type-/cycle-valid. Deliberate v1
  design decision (not a bug), since relations to already-finished beans
  remain legitimate (e.g. "blocked by" a completed bean). Pre-existing
  since E3.
- **Lobby repo metrics are not context-cancelled:** in-flight metric queries
  (`beans list` per configured repo) from a previous Lobby opening keep
  running when the Lobby is reopened — redundant, path-keyed subprocesses,
  no data mix-up. Accepted for v1; worth a context-cancel once many repos
  are configured.
- **I01 (medium, open PO point, E7 T7-Review; validation.md D04) — Header
  wraps/truncates at ~80 columns:** the 7-item global Header can lose
  `q:quit` on a narrow (~80-column) terminal, a common terminal size.
  Candidate fixes: wrap the Header like the Footer already does, or a
  priority-truncation order. Raised to the PO in the E7 epic review, not
  yet decided.
- **I02 (low, open PO point, E7 T7-Review; validation.md D05) — Overlay
  footer restates Enter/Back despite a visible Header:** several
  overlay-local footer hints (e.g. paletteLocalBindings/helpLocalBindings)
  repeat `enter`/`esc` even though the Header already shows them. Possibly
  deliberate reinforcement for a modal context; raised to the PO for a
  sign-off or an explicit invariant test.
- **D01 (open PO point, E7 T7-Review; validation.md D06) — Footer scope is
  deliberately narrow:** `browseRepoLocalBindings`/`backlogLocalBindings`
  intentionally leave out `f`/`X`/`b`/`t`/`a`/`B`/`y` (Filter/Clear-Filter/
  Backlog/Tag-/Parent-/Blocking-Picker/Yank) — all remain reachable and are
  documented in the Help-Overlay (`?`), just not restated in the footer.
  The PO may choose to widen this.

## Development

TDD (`superpowers:test-driven-development`), run via `make test`
(`command go test ./...`). Conventions (always build with `command go …`,
file-naming scheme, theme tokens, commit/review flow) →
[`CLAUDE.md`](CLAUDE.md).
