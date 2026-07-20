# beans-tui

PO-Cockpit-TUI für beans-Repos — Port der DevDash-TUI (`dd`) auf das beans-Framework.
Binary `bt`, Go/bubbletea. Datenlayer: beans-CLI-Subprocess (`beans list --json --full`,
`update --if-match`, `-S` Bleve-Suche) + eigener fsnotify-Watcher — beans-Binary bleibt die
eine Autorität; v0.4.2 exponiert keine importierbaren Packages (alles `internal/`).

## Pointer

- **Sprache/Begriffe: `docs/GLOSSARY.md`** — verbindlich (boxed field, Box-Titel, Box-Badge, Region, Golden, Smoke …). Vor UI-Arbeit lesen; neue Elemente dort nachtragen.
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
- **Footer-/Wrap-Änderungen brauchen einen tmux-Smoke bei Grenzbreite (80 Spalten)** — Unit-Tests bei 100/120 Spalten sehen Umbruch-Bugs strukturell nicht (NBSP-Wordwrap-Falle, E8/T8; Details docs/LESSONS-LEARNED.md Eintrag 4).
- **Parallele Sub-Agenten (bewährt, 2026-07-20):** (1) `isolation: worktree` legt den Worktree **immer von `main`** an, nie vom Arbeitsbranch — Agenten im Prompt anweisen, VOR der ersten Code-Änderung `git merge-base --is-ancestor <arbeitsbranch> HEAD` zu prüfen und sonst `git reset --hard <arbeitsbranch>`. (2) **Nach jedem Merge bauen UND testen** — ein konfliktfreier Textmerge kann semantisch falsch sein (Signaturänderung in einem Test des anderen Strangs). (3) **Golden-Konflikte** paralleler Stränge NICHT von Hand lösen, sondern EINMAL auf dem gemergten Code regenerieren, dann per `git diff` prüfen. (4) `.claude/` ist gitignored, weil Agent-Worktrees dort landen.
- **Vault-Symlink-Nebenwirkung:** `docs/` ist in den Obsidian-Vault gesymlinkt — Obsidian stempelt beim Öffnen `uid`-Frontmatter in Repo-Dokus und kann beim Aufräumen Dateien (z.B. `docs/screenshots/*`) vom Dateisystem entfernen. Vor Commits `git status` prüfen; Gelöschtes via `git checkout --` zurückholen.
- **Schneller Lauf:** `command go test ./... -short` — überspringt die sieben teuersten huh-drive-Tests in `internal/tui/box_confirm_create_test.go` (7-Felder-Create-Form-Drive über echte `tea.Update`-Roundtrips, je ~16-19s wegen huhs selbst-perpetuierender Blink-Tick-Cmds — `skipSlowHuhDriveInShortMode`, E3 Task 6/bean bt-ppzb). Bringt `internal/tui` von ~121s auf ~3-5s. Vor jedem Commit bleibt der VOLLE Lauf (ohne `-short`) Pflicht — `-short` ist nur der lokale Iterationsloop.

## Status-Quellen (via /ce-start, 2026-07-15)

| Quelle | Pfad/ID | Notiz |
|---|---|---|
| beans | .beans/ | `beans list --ready` |
| SSTD | docs/SSTD.md | Pointer-Manifest |
| Plans | docs/plans/v1-port/ | design-spec + implementation-plan + epic-E2..E5-Pläne |
| git | Repo (main-direkt) | — |
