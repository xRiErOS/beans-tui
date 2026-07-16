---
# bt-6oyy
title: Tag-Management-Page (zentrale Tag-Definition)
status: in-progress
type: feature
priority: normal
created_at: 2026-07-15T14:11:37Z
updated_at: 2026-07-16T15:47:02Z
---

PO-Wunsch (2026-07-15, im Zuge E7-Feedback): Tags behalten UND zentral definieren können über eine eigene Page in bt.

## Kontext

beans kennt keine Tag-Registry — Tags sind freie Frontmatter-Strings. Eine zentrale Definition braucht: (a) Persistenzort (Kandidaten: ~/.config/beans-tui/config.yaml-Sektion, .beans/-Konventionsdatei im Repo — Repo-lokal wäre teamfähig und beans-nah), (b) eigene View/Page (Liste definierter Tags + Verwendungszähler, anlegen/umbenennen/entfernen; rename müsste alle Beans mit dem Tag umschreiben — Bulk-Mutation via CLI je Bean), (c) Integration: Tag-Picker (t) bietet definierte Tags priorisiert an, freie Tags weiter erlaubt (oder strict mode — PO-Entscheid).

## Offen

- Q03 (PO): Scope — noch v1 (nach E7, vor Release) oder v1.1-Backlog? Empfehlung Supervisor: v1.1 — neues Feature, kein Port-Parität-Bestandteil; v1-Goal ist der devd-Port + PO-Feedback-Polish.
- Design-Detail (Persistenzort, strict vs. suggest) bei Task-Start klären.

Status: todo, KEIN Parent (bewusst außerhalb v1-Milestone bis Q03 entschieden).


## v1.1-Scope-Entscheid (2026-07-15, D08 aus Grilling)

D08 ENTSCHIEDEN: Tag-Management-Page ist v1.1-Scope — NICHT Teil von v1/E8.
Die v1-Minimal-Lösung für Tag-Erstellung liefert E8/B14 (bt-ntoz): 'create tag'
als Palette-Command + Neuanlage im t-Picker sichtbar (Footer-Hint). Diese Page
baut darauf auf: zentrale Verwaltung (CRUD, Farben, Umbenennen mit Propagation).


## PO-Entscheide (2026-07-16, Chat)

- **Scope/Reihenfolge (revidiert D08-Timing):** Kette für dieses Feature startet **direkt NACH der E9-Kette** (bt-tct9) — nicht erst v1.1-Backlog. Eigenes Epos + eigener Planner, sobald E9 to-review ist.
- **Persistenzort:** repo-lokal (Konventionsdatei im Repo, teamfähig, beans-nah) — exakter Dateiname/Format Planner-Entscheid.
- **Picker-Verhalten:** Suggest-Mode — definierte Tags priorisiert angeboten, freie Tags bleiben erlaubt (kein strict mode).



## Planung abgeschlossen (2026-07-16, Planner)

Eigenes Epos `bt-362n` (E10 — Tag-Management-Page) angelegt, 6 Umsetzungs-Tasks + Abschluss (`bt-49hh`/`bt-r92i`/`bt-604w`/`bt-1lsu`/`bt-y9my`/`bt-pqq3`/`bt-sohl`). Plan-Dokument: `docs/plans/tag-management/epic-E10-plan.md`. D01-D15 (Persistenzort `.beans-tags.yml` Repo-Root, Full-Capture-Dispatch, Registry-only Delete, continue-on-error Rename-Propagation etc.) + Q01-Q03 (destruktiver Delete-Modus, Stale-Definition-Marker, B14-Picker-Registrierung) im Epic-Body. Status auf `in-progress` — Epic-Kette übernimmt die Umsetzung.
