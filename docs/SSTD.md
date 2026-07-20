---
uid: e347d7f5-a3ec-4c1e-8f97-f2edfc23f8aa
---
# beans-tui — Session State Transfer (Pointer-Manifest)

## Kanonischer Einstieg

- Zweck/Design: `docs/plans/v1-port/design-spec.md` (Quelle der Wahrheit für v1)
- Plan: `docs/plans/v1-port/implementation-plan.md`
- Arbeitsplan / offen / nächster Schritt: `beans list --ready` (Dogfooding: `.beans/` dieses Repos)

## Festlegungen

- **Worktree-Weiche: main-direkt.** Solo-Repo, sequentielle Agent-Kette (NSP-Auto-Handover
  je Epos) — autonome Commits direkt auf `main`, kein Worktree-Zwang.
- **Subagents: Sonnet** (Fable nur Supervisor), Opus nur in Ausnahmen.
- Referenz-Quellen: devd-TUI `~/Obsidian/tools/DeveloperDashboard/apps/cli-go` ·
  KC-Bundle `devdash-tui` · beans-CLI 0.4.2 (Datenlayer via Subprocess, D02).

## Wichtige Dateien 

- **Glossar:** Für die Verbesserung der gemeinsamen Sprache: `docs/GLOSSARY.md`
- **Lessons Learned:** Damit sich Fehler nicht erneut ergeben: `/docs/LESSONS-LEARNED.md`

