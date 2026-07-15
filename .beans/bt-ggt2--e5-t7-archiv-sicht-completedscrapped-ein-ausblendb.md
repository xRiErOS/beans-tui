---
# bt-ggt2
title: E5 T7 — Archiv-Sicht (completed/scrapped ein-/ausblendbar)
status: todo
type: task
created_at: 2026-07-15T09:04:38Z
updated_at: 2026-07-15T09:04:38Z
parent: bt-5h4d
---

Ziel: Archiv-Sicht (completed/scrapped ein-/ausblendbar, Filter-Default:
aus). Design decision e (EMPIRISCH verifiziert, siehe epic-E5-plan.md
»Design-Entscheidungen« und beans-src/pkg/beancore/core.go:119-166 +
internal/commands/archive.go): `beans list` (Client.List, ohne Flags) liest
BEREITS `.beans/archive/` mit (loadFromDisk walks den GESAMTEN .beans-Baum
rekursiv, nur PUNKT-Verzeichnisse werden übersprungen -- `archive/` ist
KEIN Punkt-Verzeichnis). `Bean.Path` trägt für archivierte Beans das Präfix
"archive/" (Core.isArchivedPath). D.h. KEINE Datenlayer-Änderung nötig --
die aktuelle Tree/Backlog zeigt HEUTE bereits alle 27 completed-Beans dieses
Repos ungefiltert (verifiziert: `beans list --json` zählt sie mit,
data.Index hat KEINE Status-Filterung). Die Aufgabe ist also der UMGEKEHRTE
Fall: ein NEUER Default-Filter versteckt completed/scrapped, togglebar --
nicht "sichtbar machen".

Plan: docs/plans/v1-port/epic-E5-plan.md »Task 7«.

## Akzeptanz
- [ ] internal/data/bean.go: Bean.IsArchived() bool (strings.HasPrefix(Path,
      "archive/") -- billige, abgeleitete Eigenschaft, keine neue CLI-Abfrage)
- [ ] internal/tui/types.go: model.showArchived bool (default false --
      "Filter-Default: aus")
- [ ] internal/tui/box_filter_facets.go: buildFilterItems bekommt eine NEUE,
      eigenständige Zeile (facet "archive", NICHT Teil der Status-Enum-
      Schleife) "Archivierte einblenden" -- keyFilterMenu's bestehender
      Toggle-Pfad (keys.Toggle, space/x) erkennt diese Zeile speziell und
      schreibt m.showArchived statt filterStatus (kein neuer Key nötig)
- [ ] internal/tui/box_filter_facets.go beanMatchesFacets (oder eine neue
      beanMatchesArchive, AND-kombiniert in beanMatches): wenn
      !m.showArchived UND (b.Status=="completed" || b.Status=="scrapped")
      -> kein Match. WICHTIG: dies gilt NICHT als "Filter aktiv"
      (filterActive() bleibt unverändert -- der Archiv-Default ist kein
      PO-gesetzter Facet, sondern ein Programm-Default, filterSummary()
      zeigt ihn separat/gar nicht)
- [ ] Tree-Verhalten (KEINE Code-Änderung nötig, nur Test-Beweis): ein
      archiviertes Milestone/Epic OHNE nicht-archivierte Nachkommen
      verschwindet automatisch komplett aus dem Baum (flattenTreeFiltered's
      bestehende subtreeHasMatch-Logik, view_browse_repo.go:271ff) --
      Ancestor-Sichtbarkeit funktioniert bereits generisch über das match-
      Prädikat, keine Sonderbehandlung für Milestones/Epics nötig
- [ ] internal/data/index.go Backlog(): KEINE Änderung (bereits nur todo/
      draft, strukturell unbeeinflusst vom Archiv-Toggle)
- [ ] archive_test.go (data package): TestBeanIsArchivedDetectsPathPrefix,
      TestListIncludesArchivedBeans (empirischer Beleg, echtes beans-Binary,
      requireBeansBinary-Guard: newTestRepo + `beans archive` CLI-Aufruf +
      erneutes List() zeigt Bean weiterhin, Path jetzt "archive/...")
- [ ] archive_visibility_test.go (tui package):
      TestArchivedBeanHiddenFromTreeByDefault,
      TestArchivedBeanShownWhenShowArchivedToggled,
      TestArchivedOnlyMilestoneDisappearsWithNoOpenDescendant,
      TestArchiveToggleDoesNotCountAsFilterActive
- [ ] `command go test ./... -short` grün, gofmt/vet leer, Tree-/Backlog-
      Goldens 2x grün (Default-aus-Zustand identisch zu heute, da alle
      bestehenden Fixture-Beans NICHT archiviert sind)
- [ ] Commit `feat(tui): Archiv-Sicht (completed/scrapped Default-aus, togglebar)`
