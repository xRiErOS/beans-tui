---
# bt-6oyy
title: Tag-Management-Page (zentrale Tag-Definition)
status: todo
type: feature
created_at: 2026-07-15T14:11:37Z
updated_at: 2026-07-15T14:11:37Z
---

PO-Wunsch (2026-07-15, im Zuge E7-Feedback): Tags behalten UND zentral definieren können über eine eigene Page in bt.

## Kontext

beans kennt keine Tag-Registry — Tags sind freie Frontmatter-Strings. Eine zentrale Definition braucht: (a) Persistenzort (Kandidaten: ~/.config/beans-tui/config.yaml-Sektion, .beans/-Konventionsdatei im Repo — Repo-lokal wäre teamfähig und beans-nah), (b) eigene View/Page (Liste definierter Tags + Verwendungszähler, anlegen/umbenennen/entfernen; rename müsste alle Beans mit dem Tag umschreiben — Bulk-Mutation via CLI je Bean), (c) Integration: Tag-Picker (t) bietet definierte Tags priorisiert an, freie Tags weiter erlaubt (oder strict mode — PO-Entscheid).

## Offen

- Q03 (PO): Scope — noch v1 (nach E7, vor Release) oder v1.1-Backlog? Empfehlung Supervisor: v1.1 — neues Feature, kein Port-Parität-Bestandteil; v1-Goal ist der devd-Port + PO-Feedback-Polish.
- Design-Detail (Persistenzort, strict vs. suggest) bei Task-Start klären.

Status: todo, KEIN Parent (bewusst außerhalb v1-Milestone bis Q03 entschieden).
