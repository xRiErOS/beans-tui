---
uid: e347d7f5-a3ec-4c1e-8f97-f2edfc23f8aa
---
# beans-tui — Session State Transfer (Pointer-Manifest)

## Kanonischer Einstieg

- Zweck/Design: `docs/plans/v1-port/design-spec.md` (Quelle der Wahrheit für v1)
- Plan: `docs/plans/v1-port/implementation-plan.md`
- Arbeitsplan / offen / nächster Schritt: `beans list --ready` (Dogfooding: `.beans/` dieses Repos)

## Aktiver Strang (2026-07-21)

- **jira-Style-UI-Experiment** auf Branch `experiment/jira-style-ui` (~90 Commits vor `main`),
  hinter Env-Flag `BT_BOXFORM=1`. Epos `bt-vy1q`, Design `docs/plans/jira-style-experiment/`.
  **PO hat den Spike abgenommen** — voller Merge auf `main` beschlossen, Timing offen (`bt-2o9a`).
  `bt-adkn` (Body-Blättern + Seiten-Indikator) am 2026-07-21 nach 3 Nacharbeits-Runden
  **3/3 accepted** (fokus-unabh. pgup/pgdn · sticky Body-Header · stabile Indikator-Position).
  `bt-p78f` (TOC/Ankerleiste) **scrapped** (PO, zu aufwändig). Offen: `bt-ty48` (GIF),
  `bt-2o9a` (Merge-Gate PO), draft `bt-dovm` (S7, PO-Freigabe nötig).

## Festlegungen

- **Worktree-Weiche: main-direkt** für die sequentielle Agent-Kette (autonome Commits auf
  `main`). **Für PARALLELE Sub-Agenten** dagegen Worktrees nötig (ein Tree = ein HEAD) —
  Regeln + die „Worktree-immer-von-main"-Falle in `CLAUDE.md` § Regeln.
- **Subagents: Sonnet** (Fable nur Supervisor), Opus nur in Ausnahmen.
- Referenz-Quellen: devd-TUI `~/Obsidian/tools/DeveloperDashboard/apps/cli-go` ·
  KC-Bundle `devdash-tui` · beans-CLI 0.4.2 (Datenlayer via Subprocess, D02).

## Wichtige Dateien 

- **Glossar:** Für die Verbesserung der gemeinsamen Sprache: `docs/GLOSSARY.md`
- **Lessons Learned:** Damit sich Fehler nicht erneut ergeben: `/docs/LESSONS-LEARNED.md`

