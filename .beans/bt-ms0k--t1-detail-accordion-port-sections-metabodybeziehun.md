---
# bt-ms0k
title: 'T1 Detail-Accordion-Port: Sections Meta/Body/Beziehungen/Historie + kanonischer Sort (I03-Export)'
status: todo
type: task
priority: high
created_at: 2026-07-14T21:57:13Z
updated_at: 2026-07-14T21:57:13Z
parent: bt-aq5s
---

Ziel: Detail-Accordion Meta/Body/Beziehungen/Historie portieren (devd accordion.go +
view_detail_issue.go als Muster), read-only (Edit-Felder sind E3-Scope). Body-Section
rendert via glamour (Port devd editor.go:102-126 glowRender). Beziehungen-Section
(Parent/Children/Blocking/BlockedBy) braucht kanonische Sortierung für Blocking/
BlockedBy (Children ist über idx.Children bereits sortiert) -> I03-Vorbereitung:
data.SortBeans/StatusRank/PriorityRank exportieren.

Plan: docs/plans/v1-port/epic-E2-plan.md »Task 1«.

## Akzeptanz
- [ ] data.SortBeans/StatusRank/PriorityRank exportiert + getestet (index_test.go)
- [ ] accordion.go: renderAccordion (Port devd accordion.go:309-373, ohne Edit-Felder),
      exklusiv offene Section, Ziffern-Header [1]..[4]
- [ ] view_detail_bean.go: beanSections(idx, bean, bodyW) — IMMER 4 Sections
      (Meta/Body/Beziehungen/Historie), Body glamour-gerendert (leer -> Platzhalter),
      Beziehungen mit relationField (jump-only, kein Edit), dangling Blocking/
      BlockedBy-IDs als "(unresolved: id)" nicht sprungfähig
- [ ] go test ./... grün, gofmt/vet leer
