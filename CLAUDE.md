# beans-tui

PO-Cockpit-TUI für beans-Repos — Port der DevDash-TUI (`dd`) auf das beans-Framework.
Binary `bt`, Go/bubbletea, beans als Go-Library in-process (`github.com/hmans/beans@v0.4.2`).

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
- Review-Flow-Konvention: Tag `to-review` (Agent) → PO passt (`completed`) oder rejected (Tag `rework` + `## Review <datum>`-Body-Abschnitt). Agent setzt NIE `completed`.
