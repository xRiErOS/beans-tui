---
# bt-aq5s
title: E2 Browse & Detail — Tree, Accordion, Suche, Filter, Backlog
status: in-progress
type: epic
priority: high
tags:
    - to-review
created_at: 2026-07-14T18:33:30Z
updated_at: 2026-07-16T06:02:33Z
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


## E2-Abschluss (Agent-Meldung, 2026-07-15)
T1-T6 completed, je Task Spec+Quality-Review (Sonnet), 3 Fix-Runden mit echten Funden (2 Criticals in T2: Zyklus-Endlosschleife aus Plan-Snippet + fieldCursor-Panic nach Reload; Union-Fix Bleve/ID; 2 Plan-Erratums von Implementern selbst gefunden). 138 Tests grün, race-clean, 2 Goldens deterministisch. Suche/Filter/Backlog/Accordion/Beziehungs-Sprung dogfooding-belegt (tmux).

### PO-Hinweise (Entscheidungspunkte)
| Code | Beschreibung | Empfehlung | Status |
|------|--------------|------------|--------|
| I01 | esc in Detail-Fokus No-Op (nur j/tab verlassen) | belassen oder esc als 3. Exit — PO-Call | 🟡 |
| I02 | kein Sort-Modus-Indikator im Backlog | für E5/Polish vormerken | 🟡 |
| B01 | X (Filter-Reset) wirkt nicht in Backlog-View | Fix bei E3-Start (verankert in bt-gzcu) | 🟢 TestKeyBacklogFilterClearResetsFacets (E6 T1, bt-wm4w, verifiziert 2026-07-15) |
