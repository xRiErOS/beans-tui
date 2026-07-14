---
# bt-9ldr
title: 'T4 Facetten-Filter f/X: Status/Type/Priority/Tag + kanonischer Orphan-Sort (I03-Abschluss)'
status: in-progress
type: task
priority: high
created_at: 2026-07-14T21:57:28Z
updated_at: 2026-07-14T23:19:03Z
parent: bt-aq5s
blocked_by:
    - bt-4ep2
---

Ziel: Facetten-Filter `f` (Status/Type/Priority/Tag, EIN geteilter Zustand für Tree
UND Backlog) + `X` Reset. Schliesst I03 ab (Orphan-Bucket nutzt jetzt data.SortBeans
statt eigener sortByTitleThenID-Logik — Single-Source-Prinzip, das index.go für
sortBeans bereits für sich beansprucht).

Plan: docs/plans/v1-port/epic-E2-plan.md »Task 4«.

## Akzeptanz
- [ ] I03 Abschluss: sortByTitleThenID entfernt, collectOrphans/collectCycleOrphans
      nutzen data.SortBeans; bestehende Orphan-/Zyklus-Tests ggf. auf kanonische
      Reihenfolge angepasst
- [ ] box_filter_facets.go: ffItem + buildFilterItems (Status/Type/Priority/Tag —
      kein "Art"-Facet wie devd, beans-Type deckt das ab), Checkbox-Menü
      (Port-Muster view_browse_backlog.go:89-136 / view_browse_project.go:998-1033)
- [ ] Toggle-Keybinding (space/x) in keymap.go ergänzt, helpGroups-Drift-Test bleibt
      grün
- [ ] Facet-Maps nutzen die I01-Copy-on-Write-Konvention aus T2 (keine neue Ausnahme)
- [ ] Filter wirkt kombiniert mit Suche (AND) auf Tree; X leert alle Facetten;
      go test ./... grün


## Übernommene Findings aus E2-T3-Review (PFLICHT in diesem Task)
- [ ] I01: beanMatchesSearch — lokale ID-Substring-Treffer mit Bleve-Ergebnis UNIONen statt Replace (sonst Flicker: ID-Treffer verschwindet, sobald async Bleve-Antwort landet). Test dazu.
- [ ] I02 (optional): Render-Test für leeren Filter-Treffer (View() mit 0 Matches)
- [ ] Q01 (Notiz): Bleve ohne Generation-Counter — akzeptiertes Restrisiko, nur dokumentieren falls berührt
