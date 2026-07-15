---
# bt-ggt2
title: E5 T7 — Archiv-Sicht (completed/scrapped ein-/ausblendbar)
status: completed
type: task
priority: normal
created_at: 2026-07-15T09:04:38Z
updated_at: 2026-07-15T13:17:48Z
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
- [x] internal/data/bean.go: Bean.IsArchived() bool (strings.HasPrefix(Path,
      "archive/") -- billige, abgeleitete Eigenschaft, keine neue CLI-Abfrage)
- [x] internal/tui/types.go: model.showArchived bool (default false --
      "Filter-Default: aus")
- [x] internal/tui/box_filter_facets.go: buildFilterItems bekommt eine NEUE,
      eigenständige Zeile (facet "archive", NICHT Teil der Status-Enum-
      Schleife) "Archivierte einblenden" -- keyFilterMenu's bestehender
      Toggle-Pfad (keys.Toggle, space/x) erkennt diese Zeile speziell und
      schreibt m.showArchived statt filterStatus (kein neuer Key nötig)
- [x] internal/tui/box_filter_facets.go beanMatchesFacets (oder eine neue
      beanMatchesArchive, AND-kombiniert in beanMatches): wenn
      !m.showArchived UND (b.Status=="completed" || b.Status=="scrapped")
      -> kein Match. WICHTIG: dies gilt NICHT als "Filter aktiv"
      (filterActive() bleibt unverändert -- der Archiv-Default ist kein
      PO-gesetzter Facet, sondern ein Programm-Default, filterSummary()
      zeigt ihn separat/gar nicht)
- [x] Tree-Verhalten (KEINE Code-Änderung nötig, nur Test-Beweis): ein
      archiviertes Milestone/Epic OHNE nicht-archivierte Nachkommen
      verschwindet automatisch komplett aus dem Baum (flattenTreeFiltered's
      bestehende subtreeHasMatch-Logik, view_browse_repo.go:271ff) --
      Ancestor-Sichtbarkeit funktioniert bereits generisch über das match-
      Prädikat, keine Sonderbehandlung für Milestones/Epics nötig
- [x] internal/data/index.go Backlog(): KEINE Änderung (bereits nur todo/
      draft, strukturell unbeeinflusst vom Archiv-Toggle)
- [x] archive_test.go (data package): TestBeanIsArchivedDetectsPathPrefix,
      TestListIncludesArchivedBeans (empirischer Beleg, echtes beans-Binary,
      requireBeansBinary-Guard: newTestRepo + `beans archive` CLI-Aufruf +
      erneutes List() zeigt Bean weiterhin, Path jetzt "archive/...")
- [x] archive_visibility_test.go (tui package):
      TestArchivedBeanHiddenFromTreeByDefault,
      TestArchivedBeanShownWhenShowArchivedToggled,
      TestArchivedOnlyMilestoneDisappearsWithNoOpenDescendant,
      TestArchiveToggleDoesNotCountAsFilterActive
- [x] `command go test ./... -short` grün, gofmt/vet leer, Tree-/Backlog-
      Goldens 2x grün (Default-aus-Zustand identisch zu heute, da alle
      bestehenden Fixture-Beans NICHT archiviert sind)
- [x] Commit `feat(tui): Archiv-Sicht (completed/scrapped Default-aus, togglebar)`


## Prelude aus T6b-Review (PFLICHT zuerst, Reviewer 2026-07-15)

**I01 (low, vorbestehend seit T1) -- ERLEDIGT (Commit a125118):** Toast-seq ist nicht global monoton — showToast (overlay_show_toast.go:106-109) berechnete seq nur relativ zu m.toast; nach nil-Reset (dismissToast ODER applyRepoSwitched) startete der nächste Toast wieder bei seq=1. Ein noch laufender, nicht-cancelbarer toastTimeout-Tick des alten Toasts liefert toastExpiredMsg{seq:1} aus und killt den NEUEN seq=1-Toast vorzeitig (auch sticky — handleToastExpired prüft nur msg.seq==m.toast.seq). Fix: seq aus modellweitem, NIE zurückgesetztem Zähler (model.toastSeqCounter) statt m.toast.seq+1. RED-Test TestToastSeqStaysMonotonicAcrossReset (stale toastExpiredMsg nach Reset+neuem Toast dismisst NICHT) -- RED vor dem Fix (Seq-Kollision beobachtet), GREEN danach.

**I02 (low) -- ERLEDIGT (Commit a125118):** types.go:400-404-Doc behauptete 'ONLY ever written by showToast/dismissToast/handleToastExpired' — seit T6b schreibt applyRepoSwitched ebenfalls m.toast=nil. Kommentar um vierte sanktionierte Schreibstelle ergänzt (konsolidiert mit dem I01-Fix, neuer toastSeqCounter-Doc-Block direkt daneben).

## Summary

Archiv-Sicht implementiert: completed/scrapped standardmäßig ausgeblendet
(Tree UND Backlog-Praedikat, letzteres strukturell bereits unbeeinflusst),
togglebar über eine neue, eigenständige Zeile "Archivierte einblenden" im
f-Menü (facet "archive", gleicher space/x-Pfad, kein neuer Key). Kernfund
während der Umsetzung: `visibleNodes()` musste zusätzlich zu
`m.treeActive()` auch `!m.showArchived` als Routing-Bedingung bekommen --
sonst hätte der Default NIE gegriffen, sobald weder Suche noch Facetten
aktiv sind (der Normalfall/Alltagsfall), weil `flattenTree` (der
unfilterte Pfad) `beanMatches` komplett ignoriert. `filterActive()`/
`treeActive()` selbst bleiben bewusst unverändert (kein PO-Facet, design
decision e) -- die rote "Filter aktiv"-Kopfzeile springt beim Archiv-
Toggle NICHT an, live im Smoke verifiziert.

Geänderte/neue Dateien:
- `internal/data/bean.go` (+`Bean.IsArchived()`)
- `internal/data/archive_test.go` (neu)
- `internal/tui/types.go` (+`model.showArchived`, +`toastSeqCounter`
  Prelude-Feld)
- `internal/tui/box_filter_facets.go` (+facet "archive": facetHead,
  buildFilterItems, facetOn, toggleFacet, `beanMatchesArchive`,
  `beanMatches` AND-Erweiterung)
- `internal/tui/view_browse_repo.go` (`visibleNodes()`-Routing-Fix)
- `internal/tui/update.go` (`applyRepoSwitched`: `showArchived`-Reset)
- `internal/tui/update_test.go` (`TestOrphanBucketUsesCanonicalOrderOverTitle`
  auf `showArchived=true` umgestellt, siehe Deviations)
- `internal/tui/archive_visibility_test.go` (neu, 9 Tests)
- `internal/tui/testdata/tree.golden` (legitimes Update, siehe Deviations)
- `internal/tui/overlay_show_toast.go`, `overlay_show_toast_test.go`
  (Prelude, eigener Commit a125118)

## Test-Output

Prelude RED -> GREEN (`command go test ./internal/tui/ -run
TestToastSeqStaysMonotonicAcrossReset -v`):

    RED:  overlay_show_toast_test.go:214: second toast's seq (1) collides
          with the pre-reset toast's seq (1) -- want a NEW, never-reused
          generation
    GREEN: --- PASS: TestToastSeqStaysMonotonicAcrossReset (0.00s)

Alle bestehenden Toast-Tests unverändert grün danach (kein Test assertete
einen harten seq-Startwert).

Haupttask RED (vor Implementierung, `command go test ./internal/tui/ -run
'TestArchive|TestBuildFilterItemsIncludesArchiveRow|...' -v`):

    --- FAIL: TestArchivedBeanHiddenFromTreeByDefault
    --- FAIL: TestArchivedBeanShownWhenShowArchivedToggled (facet "archive"
        fehlt in filterItems)
    --- FAIL: TestBuildFilterItemsIncludesArchiveRow
    --- FAIL: TestArchivedOnlyMilestoneDisappearsWithNoOpenDescendant
    --- FAIL: TestArchivedBeanExcludedFromBleveHitByDefault
    --- FAIL: TestRepoSwitchResetsShowArchived
    (TestArchiveToggleDoesNotCountAsFilterActive/
     TestFilterClearXDoesNotResetArchiveToggle/
     TestBacklogUnaffectedByArchiveToggle bereits GREEN vor Implementierung
     -- unverändertes Verhalten war schon korrekt)

GREEN nach Implementierung: alle 9 Tests in archive_visibility_test.go
PASS, plus `TestOrphanBucketUsesCanonicalOrderOverTitle` (nach Anpassung,
siehe Deviations) und die 2 neuen data-Package-Tests
(`TestBeanIsArchivedDetectsPathPrefix`, `TestListIncludesArchivedBeans`).

Voller Lauf (ohne -short), 2x, plus vet/gofmt:

    $ command go vet ./...      # leer
    $ gofmt -l .                 # leer
    $ command go test ./... -count=1
    ok      beans-tui/cmd            0.469s
    ok      beans-tui/internal/config 0.968s
    ok      beans-tui/internal/data   2.517s
    ok      beans-tui/internal/theme  1.229s
    ok      beans-tui/internal/tui    140.715s
    (zweiter Lauf direkt danach, cached ok -- identisches Ergebnis, s.o.
    Prelude-Commit-Lauf ebenfalls 139.5s grün vor den T7-Änderungen)

Goldens, `-count=2`, byte-identisch:

    $ command go test ./internal/tui/ -run 'TestChromeGolden|TestTreeGolden|
    TestTreeGoldenDeterministic|TestBacklogGolden|
    TestBacklogGoldenDeterministic|TestReviewCockpitGolden|
    TestReviewCockpitGoldenDeterministic' -count=2 -v
    --- PASS (7 Tests) x2
    $ git status --porcelain internal/tui/testdata/
    (nur tree.golden als geändert markiert, s. Deviations -- nach Update
    erneut -count=2 grün)

## Smoke

tmux (100x30), `bin/bt` frisch gebaut, `HOME=/tmp/bt-t7-home` (isoliert),
Scratch-Repo `bt-t7-scratch` (6 Beans: Milestone+Kind offen, Kind
completed, 2 Standalone offen/completed, 1 Standalone scrapped;
`beans archive` real ausgeführt -> 3 davon jetzt unter `.beans/archive/`)
plus zweites Scratch-Repo `bt-t7-scratch2` für den Repo-Wechsel-Schritt.

**(a) Default versteckt completed/scrapped, auch archive/-Dateien:** Tree
zeigt initial nur `t7-rubw` (Milestone, offen) und `t7-fj8l` (Standalone,
offen) -- Milestone expandiert zeigt nur `t7-fpgr` (offenes Kind), das
completed Kind `t7-bwpl` bleibt versteckt. Kopfzeile zeigt weiterhin die
gemutete `⌕ / search`-Idle-Hint, NICHT die rote "Filter aktiv"-Zeile.
PASS.

**(b) f-Menü -> "Archivierte einblenden" (eigene "Archiv"-Sektion,
Checkbox anfangs `[ ]`) -> Toggle:** nach space auf der Zeile `[x]`,
esc schließt -- Tree zeigt jetzt ALLE 6 Beans inkl. `t7-bwpl`
(completed, archiviert), `t7-9mja` (completed, archiviert), `t7-q7ru`
(scrapped, archiviert). Kopfzeile weiterhin gemutet (kein "Filter
aktiv"). PASS.

**(c) Toggle zurück:** zweiter space-Druck auf derselben Zeile -> `[ ]`,
Tree zeigt wieder nur die 3 offenen Beans. PASS.

**(d) Suche respektiert den Archiv-Default:** `/Scrapped` bei
showArchived=false -> 0 Treffer (der scrapped Bean matcht per Titel, ist
aber archiv-gefiltert). showArchived=true (via f-Menü) -> derselbe Query
findet `t7-q7ru Scrapped Standalone` sofort. PASS -- bestätigt die
dokumentierte Entscheidung (beanMatches AND-Kombination, keine
Sonderbehandlung in applyBleveResult nötig).

**(e) Repo-Wechsel (p) -> Toggle zurück auf Default:** showArchived=true
gesetzt in `bt-t7-scratch`, dann `p` (Lobby öffnet von überall, zeigt
beide Repos mit korrekter offen/gesamt-Metrik 3/6 und 1/2), Auswahl von
`bt-t7-scratch2` -> Browse zeigt nur den einen offenen Bean
(`t8-meim`), das archivierte `Repo2 Done Task` bleibt versteckt. f-Menü
im neuen Repo zeigt die Archiv-Checkbox wieder `[ ]` (nicht `[x]`) --
showArchived korrekt auf false zurückgesetzt, kein Leak über den
Repo-Wechsel. PASS.

Kein Panic/Goroutine-Dump in der gesamten Sitzung (Pane-Capture geprüft,
`grep -i panic` leer).

## Deviations/ERRATA

- **`visibleNodes()`-Routing-Bedingung erweitert (`m.treeActive() ||
  !m.showArchived`), NICHT `treeActive()` selbst.** Der Akzeptanz-Text
  deutet nahe, dass `beanMatches`s neues drittes AND-Glied allein
  genügt -- tatsächlich hätte das den Default nie greifen lassen: ohne
  aktive Suche/Facetten ruft `visibleNodes()` den UNFILTERTEN
  `flattenTree()`-Pfad auf, der `beanMatches` komplett ignoriert.
  `treeActive()` selbst durfte NICHT erweitert werden (bewusst
  geprüft) -- es speist auch `treeSearchLine`s "Filter aktiv"-Rot-Logik,
  eine Erweiterung dort hätte die Kopfzeile beim reinen Archiv-Default
  fälschlich rot gefärbt (Widerspruch zum "kein PO-Facet"-Akzeptanz-
  Text). Fix lebt daher ausschließlich in `visibleNodes()`s eigener
  Bedingung, `filterActive()`/`treeActive()` bleiben byte-für-byte
  unverändert (verifiziert per `TestArchiveToggleDoesNotCountAsFilterActive`).
- **`tree.golden` legitim aktualisiert.** `goldenTreeModel`
  (tree_golden_test.go) enthält eine Fixture-Bean `gld-tsk2` mit
  Status `completed`, exponiert als Kind eines expandierten Epics --
  nach dem neuen Default verschwindet genau diese eine Zeile aus dem
  Tree. Diff verifiziert: NUR diese Zeile entfernt (durch eine
  Leerzeile ersetzt, Zeilenzahl bleibt 30), Chrome/Layout/restliche
  Zeilen byte-identisch. `-update` bewusst ausgeführt, `-count=2`
  danach erneut byte-identisch grün.
- **`TestOrphanBucketUsesCanonicalOrderOverTitle` (update_test.go) auf
  `m.showArchived = true` umgestellt.** Der Test nutzt einen `scrapped`-
  Bean gezielt, um kanonische Sortierreihenfolge gegen alphabetische
  Titel-Reihenfolge zu diskriminieren (I03-Regression) -- mit dem neuen
  Archiv-Default wäre dieser Bean jetzt standardmäßig unsichtbar und
  hätte den Test auf einen einzelnen verbleibenden Orphan kollabiert,
  was NICHT der eigentliche Testzweck ist (reine Sortierreihenfolge,
  unabhängig von Archiv-Sichtbarkeit). `showArchived=true` hält beide
  Orphans sichtbar, Testzweck bleibt intakt.
- **Suche/Bleve: keine Code-Änderung, nur Test+Doku.** Plan-Task-7-Text
  ließ die Suche-Frage offen ("prüfe ob..."); Entscheidung dokumentiert
  in `beanMatches`s eigenem Kommentar (box_filter_facets.go) und live im
  Smoke (d) verifiziert: `beanMatches` kombiniert Suche/Facetten/Archiv
  per AND, ein archivierter Bleve-Treffer wird also automatisch
  ausgeblendet, keine Sonderbehandlung in `applyBleveResult` nötig.
- **Archiv-Facet-Zeile bekommt eine eigene "Archiv"-Sektion im
  Filter-Menü** (`facetHead["archive"] = "Archiv"`), statt kopflos am
  Ende der Tag-Liste zu hängen -- Akzeptanz-Text sagt nur "eigenständige
  Zeile", keine Sektions-Vorgabe; ohne eigenen `facetHead`-Eintrag hätte
  `treeFilterBox()`s Gruppierungs-Logik eine leere Kopfzeile gerendert
  (facetHead-Map-Zugriff auf einen fehlenden Key liefert `""`). Live im
  Smoke (b) verifiziert: eigene "Archiv"-Überschrift erscheint sauber.

## Notes for T8

- E5-Abschluss-Bean `bt-7dfj` listet die Prelude/T6b-Review-Punkte
  bereits separat (I01 aus T4-Review zu `reviewClickRow`, I02 aus
  T6-Review zu `repoMetricsCmd`-Cancel-Guard) -- unverändert, keine
  neuen Funde aus T7 zu ergänzen.
- **Kein neuer Known-Issue aus T7.** Die einzige während der Arbeit
  gefundene Lücke (`visibleNodes()`s Routing) wurde noch in diesem Task
  gefixt, nicht als Folgearbeit verschoben.
- T8s Volllauf sollte den bereits hier bestätigten `-count=2`-Goldens-
  Stand vorfinden (`tree.golden` ist bereits aktuell) -- keine weitere
  Golden-Migration in T8 zu erwarten, nur die finale Bestätigung.
