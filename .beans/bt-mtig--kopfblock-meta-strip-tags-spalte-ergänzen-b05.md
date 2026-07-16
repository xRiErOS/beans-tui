---
# bt-mtig
title: 'Kopfblock-Meta-Strip: Tags-Spalte ergänzen (B05)'
status: in-progress
type: task
priority: normal
created_at: 2026-07-16T06:45:40Z
updated_at: 2026-07-16T07:40:03Z
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

- [ ] Kopfblock-Zeile 4 zeigt `tags: <Swatches>` für ein Bean mit Tags
- [ ] Taglos zeigt `tags: (none)`
- [ ] type/status/prio-Spalten unverändert (Padding-Breiten aus E8-B02 nicht angetastet)
- [ ] Goldens verifiziert (regeneriert ODER Gegenbeleg grün, im Commit-Body dokumentiert)
- [ ] Voller Testlauf grün, gofmt/vet leer
