# beans-tui

PO-Cockpit-TUI für beans-Repos — Port der DevDash-TUI (`dd`) auf das beans-Framework.
Binary `bt`, Go/bubbletea. Datenlayer: beans-CLI-Subprocess (`beans list --json --full`,
`update --if-match`, `-S` Bleve-Suche) + eigener fsnotify-Watcher — beans-Binary bleibt die
eine Autorität; v0.4.2 exponiert keine importierbaren Packages (alles `internal/`).

## Pointer

- Design (Quelle der Wahrheit): `docs/plans/v1-port/design-spec.md`
- Plan: `docs/plans/v1-port/implementation-plan.md`
- Stand/Weichen: `docs/SSTD.md` · Arbeit: `.beans/` (`beans list --ready`)
- Referenz-Code (Port-Basis): `~/Obsidian/tools/DeveloperDashboard/apps/cli-go` (Eriks devd-TUI)

## Regeln

- **Worktree-Weiche: main-direkt** — autonome Commits auf `main`, PO-Gate = Abnahme, nicht Commit.
- Build IMMER `command go build -o bin/bt .` (lokales `go`-Shadowing wie devd-cli meiden).
- TDD: Update-Tests (`tea.KeyMsg`) + Golden-View-Snapshots (`testdata/*.golden`).
- Datei-Namenskonvention `internal/tui/`: `<art>_<verb>_<entität>.go` (`view_`/`form_`/`box_`/`picker_`/`overlay_`); Infrastruktur ohne Präfix.
- Keymap Single Source: `internal/tui/keymap.go` — Help-Overlay wird daraus generiert.
- Theme-Token nur aus `internal/theme/` (Catppuccin Macchiato, TrueColor) — keine Hex-Literale in Views.
- Review-Flow-Konvention (gilt für **Epic-/Milestone-beans**): Tag `to-review` (Agent) → PO passt (`completed`) oder rejected (Tag `rework` + `## Review <datum>`-Body-Abschnitt). Agent setzt bei Epics/Milestones NIE `completed`. **Implementierungs-Task-beans** sind dagegen agent-abschließbar (completed nach grünen Tests + Review-Durchlauf) — Plan-Ritual in docs/plans/v1-port/implementation-plan.md.
