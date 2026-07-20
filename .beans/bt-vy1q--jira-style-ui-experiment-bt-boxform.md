---
# bt-vy1q
title: jira-Style-UI Experiment (BT_BOXFORM)
status: in-progress
type: epic
priority: high
created_at: 2026-07-20T07:24:38Z
updated_at: 2026-07-20T07:24:38Z
---

Spike: ist eine flachere, salientere jira-artige Darstellung eine Verbesserung fuer beans-tui?
Vorbild: jiratui.sh (Python/Textual — nur Design-Vorbild, kein Code-Transfer).

## Kontext fuer alle Kinder (DRY)
- **Branch:** `experiment/jira-style-ui` (main-direkt-Repo; Spike, kein Merge-Zwang)
- **Feature-Flag:** `BT_BOXFORM=1` — Default AUS. Alles additiv + gated; Bestandsgolden muessen byte-identisch bleiben.
- **Quelle der Wahrheit:** `docs/plans/jira-style-experiment/design-spec.md` (Entscheidungen D01-D12, Catppuccin-Farb-Salienzkarte, Mockups, Risiken I01-I04)
- **Plan:** `docs/plans/jira-style-experiment/implementation-plan.md` (Slice-Roadmap S1-S8)
- **Resume-Anker/Status:** `docs/plans/jira-style-experiment/STATE.md` — bei Wiederaufnahme ZUERST lesen

## Harte Constraints
- Voller Testlauf `command go test ./...` (OHNE `-short`) gruen vor jedem Commit
- Keine Hex-Literale in Views — nur Tokens aus `internal/theme/`
- Keymap Single Source: `internal/tui/keymap.go`; jede neue Bindung MUSS in `helpGroups()` (Drift-Guard-Test)
- Case-Konvention D07: **klein** = Feld-Aktion (`s t a r e o u`), **gross** = View/Global (`S X K G`)
- Build: `command go build -o bin/bt .`

## Erledigt (S1-S6, Belege in git)
- S1 `2b08977` dropdownBox-Primitiv (Label+Hotkey IM Rahmen, lipgloss kann das nicht nativ)
- S2 `f76b6e9` detailBoxForm Scalars + R1 (Label=Subtext)
- S2c `d62edf0` panelBox multi-line + Body/Relations/History
- S2b `987ce06` Live-Wiring gated in `renderAccordionPane`
- S2e `e27c5f2` feste 3|2-Gruppierung (D12) + Geometrie-Fixes B1-B5
- S3 `db52457` persistente Filter-Leiste (Chips, aktiv=Peach)
- S4 `5dc3d01` editierbare Dropdowns: Keys `o`=Type, `u`=Priority
- S5 `bad6c18` Nested/Flat-Toggle `G`
- S6 `cf00b72` Maus: B6-Offset, Klick-oeffnet-Feld-Editor, Flat-Klick

## Validiert
- VHS-GIF gegen sproutling: `~/Obsidian/Vault/lean-stack/beans-tui/beans-tui-boxform.gif`
- Clamping-GIF schmal: `beans-tui-boxform-narrow.gif`
- 80-Spalten-tmux-Smoke bestanden: kein Overflow, kein Wrap-Bug, Labels clampen graziös

## Offen
Siehe Kinder-beans. Merge auf main erst nach PO-Freigabe.
