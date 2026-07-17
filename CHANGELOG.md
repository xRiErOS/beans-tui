# Changelog

This project doesn't cut versioned releases yet — this log tracks completed
epics instead. Each epic is a bean (`bt-*`); the full PO-acceptance record
(evidence, decisions, open points) lives in
[`docs/plans/v1-port/validation.md`](docs/plans/v1-port/validation.md) and
the per-epic plans under `docs/plans/v1-port/epic-*.md`.

## E8 — PO-Feedback R2: tags row + detail-UX refinements

Resolved the four open PO decisions from E7's review (`validation.md` §5,
D01/D04/D05/D06): tags now render as a 5th Meta field (`tags:`, after
`priority:`) with `enter` opening the Tag-Picker directly; the Header was
cut to exactly 4 always-visible globals (`ctrl+k`, `p`, `?`, `q`) instead of
wrapping at narrow widths, with `ctrl+r`/`esc`/`enter` staying reachable via
the Help-Overlay (`?`); the Browse/Backlog footer was rebuilt to include
`f`/`b`/`t`/`a`/`y`/Blocking (remapped from `B` to `r`) instead of hiding
them behind Help-only discovery.

## E7 — PO-Feedback R1: Detail-UX + Type/Status/Priority glyphs

The Review-Cockpit view (`R`) was removed — PO decision: "contradicts the
lean-stack spirit and reintroduces ceremony". Review now happens entirely
outside the TUI (see the Review-workflow note in `README.md`). Type/Status/
Priority now render as single colored letters/glyphs instead of words or
shape icons (see the Glyph legend in `README.md`). All user-facing strings
were normalized to English, the Command-Center adopted a `verb entity`
schema. The Detail pane gained a header block (bean-id / title / `type: …
status: … prio: …`) above an editable Meta field list with a permanently
reserved cursor gutter. Detail-Focus gained an enter-cascade (section →
field → edit-overlay) with symmetric `tab`/`shift+tab` focus pairing.

## E6 — Validation & release

All 14 v1 user stories validated against `design-spec.md` §10 — 13 PASS, 1
PARTIAL at the time (US-08, tag visibility — resolved in E8 above). Full
evidence trail and smoke-test appendix: `validation.md`.

## E1 – E5 — Foundation through Polish

- **E1 Foundation** — read-only Tree over the beans data layer (Milestones
  → Epics → Tasks), live-reload via an fsnotify watcher, quit-confirm.
- **E2 Browse & Detail** — master-detail focus with a Detail Accordion
  (Meta/Body/Relations/History, relation-jump), local live search + Bleve
  from 3 characters, facet filter (Status/Type/Priority/Tag, shared across
  Tree and Backlog), the Backlog view with sort-toggle.
- **E3 Mutations** — combined Status/Type/Priority menu, Tag-/Parent-/
  Blocking-Picker, Create-Form (huh, confirm-gate), title/body edit
  (`$EDITOR`), delete-confirm with children-/link-warning (no cascade),
  throughout with ETag-conflict handling.
- **E4 Command-Center** — `ctrl+k`, fuzzy actions + bean search mixed,
  context-first.
- **E5 Polish** — Toast system (incl. sticky conflict), Help-Overlay (`?`),
  Yank (`y`, OSC52+native, bean/epic context), mouse (wheel/click/
  double-click), Settings (`~/.config/beans-tui/`), Lobby v1 + Repo-Picker
  (`p`, watcher lifecycle switch), Archive view (completed/scrapped
  default-off, togglable).
