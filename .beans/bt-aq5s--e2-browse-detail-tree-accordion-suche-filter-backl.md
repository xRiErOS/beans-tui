---
# bt-aq5s
title: E2 Browse & Detail — Tree, Accordion, Suche, Filter, Backlog
status: in-progress
type: epic
priority: high
created_at: 2026-07-14T18:33:30Z
updated_at: 2026-07-14T22:02:11Z
parent: bt-apmy
blocked_by:
    - bt-blsy
---

Ziel: V2 Browse komplett (Master-Detail, Fokus-Tausch), V4 Detail-Accordion (glamour-Body, Beziehungs-Navigation), Suche `/` (lokal + Bleve via `-S`), Facetten-Filter `f`, V3 Backlog.

Epos-Start-Ritual: design-spec §6/§7 + implementation-plan »E2« lesen, voll-granularen Plan nach docs/plans/v1-port/epic-E2-plan.md schreiben (TDD wie E1), dann Task-beans anlegen (parent = dieses Epos). Port-Quellen: accordion.go, view_detail_issue.go, view_browse_project.go, view_browse_backlog.go (devd cli-go).


## Ergänzungen aus T8-Opus-Review (PFLICHT in E2)
- [ ] I01: expanded-map Ownership-Konvention festlegen (Doc-Stempel oder Copy-on-Write in setExpanded) BEVOR weitere map-Felder (Filter/Form/Picker) dazukommen
- [ ] I02: flattenTree/collectOrphans wird pro Frame gebaut — nur bei Profiling-Bedarf cachen (Notiz, kein Auto-Fix)
- [ ] I03: Orphan-Sortierung (Title→ID) vs. Tree-Sortierung (Status→Prio→Type→Titel) — kanonischen Sort exportieren oder bewusst belassen
- [ ] Q01: Init() nil-client-Guard erwägen (Invariante liegt an Run()-Boundary)
- [ ] Maus-Task erst NACH Windowing (windowStart wird für Click-Y→Row gebraucht)
