---
# bt-aq5s
title: E2 Browse & Detail — Tree, Accordion, Suche, Filter, Backlog
status: todo
type: epic
priority: high
created_at: 2026-07-14T18:33:30Z
updated_at: 2026-07-14T18:33:30Z
parent: bt-apmy
blocked_by:
    - bt-blsy
---

Ziel: V2 Browse komplett (Master-Detail, Fokus-Tausch), V4 Detail-Accordion (glamour-Body, Beziehungs-Navigation), Suche `/` (lokal + Bleve via `-S`), Facetten-Filter `f`, V3 Backlog.

Epos-Start-Ritual: design-spec §6/§7 + implementation-plan »E2« lesen, voll-granularen Plan nach docs/plans/v1-port/epic-E2-plan.md schreiben (TDD wie E1), dann Task-beans anlegen (parent = dieses Epos). Port-Quellen: accordion.go, view_detail_issue.go, view_browse_project.go, view_browse_backlog.go (devd cli-go).
