---
# bt-blsy
title: E1 Foundation — Skeleton, Datenlayer, Theme, App-Shell
status: in-progress
type: epic
priority: high
tags:
    - to-review
created_at: 2026-07-14T18:32:41Z
updated_at: 2026-07-14T21:43:15Z
parent: bt-apmy
---

Ziel: Lauffähiges `bt` mit komplettem Datenlayer (Read/Write/Watch via beans-CLI-Subprocess, D02), Catppuccin-Theme-Port, Chrome-Primitiven und read-only Tree (Dogfooding im eigenen Repo).

Kontext für alle Kinder: implementation-plan.md »Epos E1« enthält die voll-granularen TDD-Steps je Task — dort nachschlagen. Port-Quelle: ~/Obsidian/tools/DeveloperDashboard/apps/cli-go (Eriks Code). Konventionen: CLAUDE.md (command go, Datei-Namensschema, Theme-Token nur aus internal/theme). Tests: Update-Tests + Golden-Snapshots.


## E1-Abschluss (Agent-Meldung, 2026-07-14)
Alle 9 Tasks completed, je Task Spec- + Quality-Review (Sonnet; T8 Opus), 4 Fix-Runden (T3/T4/T5/T8-Findings). TUI lauffähig: Tree mit Cursor-Follow-Windowing, Orphan-/Zyklus-Sichtbarkeit, Live-Reload (fsnotify), Quit-Confirm, Dogfooding-Smoke belegt (tmux). Tests: 4 Packages grün, Golden deterministisch. PO-Abnahme ausstehend (Tag to-review) — Agent schließt nicht (lean-stack-Gate).
