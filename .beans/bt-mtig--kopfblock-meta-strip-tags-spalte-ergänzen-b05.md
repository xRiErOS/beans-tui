---
# bt-mtig
title: 'Kopfblock-Meta-Strip: Tags-Spalte ergänzen (B05)'
status: completed
type: task
priority: normal
created_at: 2026-07-16T06:45:40Z
updated_at: 2026-07-16T07:54:44Z
parent: bt-tct9
---

E9 Task 1 — deckt B05 aus bean bt-tct9 (bereits redefiniert im Epic-Body, Abschnitt "B05 REDEFINIERT"). Quelle: design-spec.md §15 PF-17 (Abschnitt B05). Ist-Code: internal/tui/view_detail_bean.go (detailHeaderBlock, tagsInline aus render_shared.go). Kein blocked_by, unabhängig von jedem anderen E9-Task (eigene, disjunkte Funktion).

## B05 — Kopfblock-Meta-Strip zeigt zusätzlich Tags

PO verbatim: "Ich sehe die Tags unter '1 META' aber nicht im Meta-Strip." Die
`type: …    status: …    prio: …`-Zeile im Detail-Pane-Kopfblock (`detailHeaderBlock`,
view_detail_bean.go) bekommt eine VIERTE Spalte `    tags: <Werte>` — PO-Mockup exakt:

    type: epic  status: in-progress  prio: !  tags: to-review

## Architektur-Vorgabe

`detailHeaderBlock(b *data.Bean, w int) string` — die bestehende `typeStatusPrio`-Zeile
(3 Spalten, feste Padding-Breiten `type`→9, `status`→11 aus E8-B02, UNVERÄNDERT) bekommt
angehängt: `theme.Muted.Render("    tags: ") + <tagsOrNone>` — `tagsInline`
(render_shared.go) liefert bereits Hash-gefärbte `● tag`-Swatches (wiederverwendet aus
metaFields, PF-15). Taglos: `theme.Dim.Render("(none)")` (dieselbe Konvention wie
metaFields' eigener taglos-Fallback). Variable Breite der vierten Spalte ist unkritisch
(Zeilenende, kein nachfolgendes gepaddetes Feld — anders als type/status/prio, die
NACHFOLGENDE Spalten haben und deshalb feste Breite brauchen).

Code-Sketch:

```go
typeStatusPrio := theme.Muted.Render("type: ") + theme.TypeStyle(b.Type).Render(typeWord) +
    theme.Muted.Render("    status: ") + theme.StatusStyle(b.Status).Render(statusWord) +
    theme.Muted.Render("    prio: ") + theme.Priority(priority) +
    theme.Muted.Render("    tags: ") + tagsOrDim(b.Tags)
```

`tagsOrDim` ist ein kleiner, neuer Helfer ODER inline-Duplikat von metaFields' eigenem
`tags := tagsInline(b.Tags); if tags == "" { tags = theme.Dim.Render("(none)") }`-Muster
— Implementer-Entscheidung; bei Duplizierung Kommentar-Verweis auf metaFields ergänzen,
damit beide Stellen nicht driften.

## TDD-Schritte

1. Failing tests (view_detail_bean_test.go): `TestDetailHeaderBlockShowsTagsColumn`
   (Bean mit Tags "to-review" → Kopfblock-Zeile 4 enthält "tags:" + "to-review"),
   `TestDetailHeaderBlockShowsNoneForTaglessBean` (Bean ohne Tags → "(none)").
2. `command go test ./internal/tui/... -run TestDetailHeaderBlock` → FAIL.
3. Implementieren.
4. `command go test ./internal/tui/... -run TestDetailHeaderBlock` → PASS.
5. Golden-Check: prüfen ob TestTreeGolden/TestBacklogGolden/TestChromeGolden ein Bean mit
   Tags im aufgeklappten Kopfblock zeigen. Falls ja: `command go test ./internal/tui/ -run
   "TestTreeGolden|TestBacklogGolden|TestChromeGolden" -update`, Vorher/Nachher-Diff im
   Commit-Body dokumentieren (Pflicht). Falls nein: Gegenbeleg-Lauf OHNE -update MUSS grün
   bleiben, explizit als "unverändert" im Commit-Body vermerken.
6. `command go test ./... -short` grün (2x), voller Lauf grün, `-race` grün, gofmt/vet leer.
7. Commit `feat(tui): Kopfblock-Meta-Strip zeigt Tags (B05)`, Footer `Refs: bt-tct9`.

## Akzeptanz-Checkliste

- [x] Kopfblock-Zeile 4 zeigt `tags: <Swatches>` für ein Bean mit Tags
- [x] Taglos zeigt `tags: (none)`
- [x] type/status/prio-Spalten unverändert (Padding-Breiten aus E8-B02 nicht angetastet)
- [x] Goldens verifiziert (regeneriert ODER Gegenbeleg grün, im Commit-Body dokumentiert)
- [x] Voller Testlauf grün, gofmt/vet leer

## Summary

B05 (design-spec.md §15 PF-17, bt-tct9 "B05 REDEFINIERT") umgesetzt:
`detailHeaderBlock` (view_detail_bean.go) rendert die `type/status/prio`-Zeile
jetzt mit einer 4. Spalte `    tags: <tagsInline(b.Tags)>` (Taglos:
`theme.Dim.Render("(none)")`, spiegelt metaFields' eigene Konvention).
type/status-Padding aus E8-B02 unangetastet — tags ist ein reiner Anhang
(variable Breite unkritisch, Zeilenende). Ganze Zeile bleibt über `truncate`
auf `w` geklemmt wie zuvor.

Nebenbefund beim Commit-Gate: `detailClickAt(t, m, "tags:")` (mouse_test.go)
fand ab jetzt zuerst das NEUE Kopfblock-"tags:" (Zeile 3, < headerBlockLines)
statt der Meta-Feldliste-Zeile — 4 Tests brachen (TestDetailClickRowMaps
MetaFieldClick, TestMouseDetailClickSingleClickSelectsField, TestMouseDetail
ClickDoubleClickOnFieldOpensOverlay, TestMouseDetailClickSecondClickOn
SelectedFieldOpensOverlayOutsideWindow). Fix: neuer Helfer
`metaTagsFieldSubstr()` (= `fmt.Sprintf("%-12s", "tags:")`, render-geerdet
gegen dieselbe Formel wie metaSectionBody's Label-Padding) disambiguiert
alle 6 betroffenen Aufrufstellen von der Kopfblock-Kollision.

## Test-Output

RED (vor Implementierung, `command go test ./internal/tui/... -run TestDetailHeaderBlock -v`):

    === RUN   TestDetailHeaderBlockShowsTagsColumn
        view_detail_bean_test.go:351: header block line 4 missing "tags:" label: "type: epic         status: in-progress    prio: !"
        view_detail_bean_test.go:354: header block line 4 missing tag value "to-review": "type: epic         status: in-progress    prio: !"
        view_detail_bean_test.go:358: header block line 4 does not contain tagsInline(b.Tags) output "● to-review" verbatim: "type: epic         status: in-progress    prio: !"
    --- FAIL: TestDetailHeaderBlockShowsTagsColumn (0.00s)
    === RUN   TestDetailHeaderBlockShowsNoneForTaglessBean
        view_detail_bean_test.go:377: header block line 4 missing taglos placeholder "tags: (none)": "type: task         status: todo           prio: ·"
    --- FAIL: TestDetailHeaderBlockShowsNoneForTaglessBean (0.00s)
    FAIL

GREEN (nach Implementierung, gleicher Lauf):

    === RUN   TestDetailHeaderBlockShowsIDTitleTypeStatusPrio
    --- PASS: TestDetailHeaderBlockShowsIDTitleTypeStatusPrio (0.00s)
    === RUN   TestDetailHeaderBlockFixedColumnWidthsNoJumpAcrossBeans
    --- PASS: TestDetailHeaderBlockFixedColumnWidthsNoJumpAcrossBeans (0.00s)
    === RUN   TestDetailHeaderBlockTruncatesToWidth
    --- PASS: TestDetailHeaderBlockTruncatesToWidth (0.00s)
    === RUN   TestDetailHeaderBlockShowsTagsColumn
    --- PASS: TestDetailHeaderBlockShowsTagsColumn (0.00s)
    === RUN   TestDetailHeaderBlockShowsNoneForTaglessBean
    --- PASS: TestDetailHeaderBlockShowsNoneForTaglessBean (0.00s)
    PASS
    ok  	beans-tui/internal/tui	0.429s

Commit-Gate:

    command go test ./... (voller Lauf)
    ok  	beans-tui/cmd	0.342s
    ok  	beans-tui/internal/config	(cached)
    ok  	beans-tui/internal/data	(cached)
    ok  	beans-tui/internal/theme	0.548s
    ok  	beans-tui/internal/tui	136.949s

    command go test ./internal/tui/ -race
    ok  	beans-tui/internal/tui	139.379s

    gofmt -l .    -> leer (nach Formatierungs-Fix in mouse_test.go)
    command go vet ./...  -> leer

## Smoke

Real in tmux gegen dieses Repo (`.beans/` echte Daten, `./bin/bt`, 160x30):
- Tags nicht-leer: bt-apmy (Milestone, tags `to-review`+`smoke`) zeigt
  `type: milestone    status: todo           prio: !    tags: ● to-review ● smoke`
  (Hash-Swatches aus echten Repo-Daten, exakt wie PO-Mockup-Form).
- Taglos: bt-tct9 (Epic, keine tags) zeigt
  `type: epic         status: in-progress    prio: !    tags: (none)`.
- type/status/prio-Spalten unverändert ggü. E8-B02 (Padding/Startspalten
  identisch in beiden Captures).

## Deviations/ERRATA

- Footer-Referenz abweichend vom bean-Text (Schritt 7 nennt `Refs: bt-tct9`)
  — Supervisor-Auftrag verlangt explizit `Refs: bt-mtig`; gemäß Supervisor-
  Weisung umgesetzt.
- Zusätzlich zu den 2 spezifizierten TDD-Tests (TestDetailHeaderBlockShows
  TagsColumn/...NoneForTaglessBean) musste `mouse_test.go` an 6 Stellen
  angepasst werden (siehe Summary) — kein Scope-Creep im Produktcode, reine
  Test-Kollisions-Behebung, notwendig für den Pflicht-Commit-Gate (voller
  Lauf grün).
- Goldens: tree.golden + backlog.golden geändert (siehe Golden-Diff im
  Commit-Body) — beide zeigten bereits vor dieser Änderung ein Bean im
  aufgeklappten Kopfblock, TDD-Schritt 5 traf also den "ja"-Zweig.

## Notes for T(n+1)

- `metaTagsFieldSubstr()` (mouse_test.go) ist jetzt der Standard-Weg, um in
  Maus-Tests gezielt die Meta-Feldliste-"tags:"-Zeile zu treffen (nicht die
  Kopfblock-Spalte) — bei künftigen Kopfblock-Erweiterungen mit weiteren
  Labels, die auch in metaFieldLabels vorkommen, dasselbe Muster
  (`fmt.Sprintf("%-12s", ...)`) wiederverwenden statt erneut zu debuggen.
- Bei sehr schmalen Panes (Golden-Testbreite, ~62 Spalten) wird die
  Tags-Spalte durch `truncate` bis zur Ellipse gekappt (`tags: …`) — das ist
  erwartetes, bereits getestetes Verhalten (TestDetailHeaderBlockTruncates
  ToWidth), kein Bug.
