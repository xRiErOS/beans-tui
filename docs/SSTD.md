---
uid: e347d7f5-a3ec-4c1e-8f97-f2edfc23f8aa
---
# beans-tui — Session State Transfer (Pointer-Manifest)

## Kanonischer Einstieg

- Zweck/Design: `docs/plans/v1-port/design-spec.md` (Quelle der Wahrheit für v1)
- Plan: `docs/plans/v1-port/implementation-plan.md`
- Arbeitsplan / offen / nächster Schritt: `beans list --ready` (Dogfooding: `.beans/` dieses Repos)

## Aktiver Strang (2026-07-21)

- **jira-Style-UI-Experiment GEMERGED auf `main`** (`bt-2o9a` done, Merge-Commit `ae2efe3`,
  `--no-ff`, als opt-in `BT_BOXFORM=1`, Default AUS). Epos `bt-vy1q`, Design
  `docs/plans/jira-style-experiment/`. Post-Merge Build + voller Testlauf grün.
  `bt-adkn` (Body-Blättern + Seiten-Indikator) 3/3 accepted; `bt-p78f` (TOC) scrapped.
  **Offen im Epos `bt-vy1q`:** `bt-ty48` (GIF), draft `bt-dovm` (S7 huh→Inline-Box, PO-Freigabe).
  **Released 2026-07-21:** `main` nach `origin/main` gepusht + annotated Tag **`v0.1.0`**
  (erster öffentlicher Release, `xRiErOS/beans-tui`). README-Screenshots aktualisiert
  (5 Bestand + `boxform.png` für §Experimental; isoliertes Single-Repo-vhs-Setup, kein Leak).
  Branch `experiment/jira-style-ui` voll gemerged, noch nicht gelöscht (aufräumbar per `git branch -d`).

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

