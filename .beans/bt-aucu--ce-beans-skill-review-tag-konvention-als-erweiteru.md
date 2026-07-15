---
# bt-aucu
title: 'ce-beans-Skill: Review-Tag-Konvention als Erweiterung dokumentieren'
status: todo
type: task
created_at: 2026-07-15T14:28:42Z
updated_at: 2026-07-15T14:28:42Z
---

PO-Side-ToDo aus beans-tui-Session (2026-07-15): Die Review-Tag-Konvention gehört in den ce-beans-Skill (lokale Erweiterung, ~/.claude/skills/ce-beans/SKILL.md bzw. echtes Verzeichnis in ~/.agents/skills).

## Inhalt der Erweiterung

Tag-Trio für Chat-Review (kein Tool-Feature, reine Konvention):
- to-review — Agent meldet Epic/Milestone fertig (Task-beans bleiben agent-abschließbar)
- accepted — PO nimmt im Chat an
- rejected — PO weist zurück, Begründung im Chat/Bean-Body → Agent-Rework, danach wieder to-review

Kontext: Review-Cockpit wurde aus beans-tui entfernt (PF-14, bean beans-tui:bt-heg9) — Review passiert im Chat, Tags steuern nur den Zustand. lean-stack-Prinzip: Signal statt Zeremonie.

## Akzeptanz

[ ] ce-beans-Skill um Review-Tag-Sektion ergänzt (kurz, im Stil des Skills)
[ ] Abgleich mit design-spec §5 in beans-tui (nach E7-Planner-Update) — Formulierungen konsistent
[ ] Drift-Net beachten (kein Rename, nur Inhalts-Edit — pre-commit sollte durchgehen)
