---
# bt-2jve
title: 'T2 Master-Detail-Verdrahtung: Section/Feld-Fokus, Beziehungs-Sprung, Map-Ownership (I01), nil-Guard (Q01)'
status: todo
type: task
priority: high
created_at: 2026-07-14T21:57:17Z
updated_at: 2026-07-14T21:57:17Z
parent: bt-aq5s
blocked_by:
    - bt-ms0k
---

Ziel: Detail-Fokus-Maschine verdrahten (Section<->Feld, wie devd view_detail_issue.go
keyDetailFocus) + Beziehungs-Sprung (enter auf Parent/Child/Blocking/BlockedBy setzt
Tree-Cursor + expandiert Vorfahren-Pfad) + zwei PFLICHT-Items aus T8-Review: I01
(expanded-map Ownership: Copy-on-Write-Konvention, BEVOR E2 weitere map-Felder
addiert) und Q01 (Init() nil-client-Guard).

Plan: docs/plans/v1-port/epic-E2-plan.md »Task 2«.

## Akzeptanz
- [ ] I01: Copy-on-Write-Konvention für map[string]bool-Modellfelder dokumentiert
      (types.go) + cloneBoolMap-Helper + setExpanded refactored; Regressionstest
      beweist keine Shared-Mutation über Modell-Kopien hinweg
- [ ] Q01: Init() nil-client-Guard (Test + Fix), kein Panic bei m.client == nil
- [ ] keyDetailFocus: Ziffer-Sprung, i/k Section-Wechsel, l/j Section<->Feld-Ebene
      (nur Beziehungen hat Felder), enter auf Beziehungs-Zeile springt + verlässt
      Detail-Fokus
- [ ] focusedBean()-Dispatcher (Port devd focusedIssue, view_detail_issue.go:20-35) —
      view-agnostische Basis für T5/Backlog-Wiederverwendung
- [ ] view_browse_repo.go renderDetailPane nutzt jetzt beanSections+renderAccordion
      (T1) statt Platzhalter
- [ ] tree.golden ggf. regeneriert (-update) + Determinismus-Test weiter grün
