---
# bt-2v38
title: Titel-Edit-Form wird multi-line (B03)
status: todo
type: task
created_at: 2026-07-16T06:45:45Z
updated_at: 2026-07-16T06:45:45Z
parent: bt-tct9
---

E9 Task 3 — deckt B03 aus bean bt-tct9. Quelle: design-spec.md §15 PF-17 (Abschnitt B03).
Ist-Code: internal/tui/form_edit_title.go (buildEditTitleForm, openEditTitleForm),
internal/tui/forms_shared.go (styleForm/formInnerHeight als Kontext). Kein blocked_by —
unabhängig von jedem anderen E9-Task (eigene Datei, keine Überschneidung).

## B03 — Titel-Edit-Form wird multi-line

PO: Titel-Edit-Form ist single-line; bei langen Titeln muss sie multi-line umbrechen
(huh-Form: Input→Text oder Wrap). Betrifft NUR das Formular hinter `enter` auf dem
`title:`-Feld (`activateDetailField`s `"title"`-Case, unverändert) — NICHT den neuen
Ganz-Bean-Editor aus D01 (separater Task, bt-tct9 Task 2), der Titel als Teil der
Frontmatter im `$EDITOR` sowieso frei editierbar macht.

## Architektur-Vorgabe

huh v1.0.0 (`go.mod`, verifiziert gegen `field_text.go`) bringt bereits eine Multi-Line-
Textarea-Komponente: `huh.NewText()`. `buildEditTitleForm` (form_edit_title.go) tauscht:

```go
// VORHER:
field := huh.NewInput().Key("title").Title("Title").Value(&v).Validate(nonEmpty)

// NACHHER:
field := huh.NewText().Key("title").Title("Title").Lines(3).
    ExternalEditor(false).Value(&v).Validate(nonEmpty)
```

`.ExternalEditor(false)` ist PFLICHT (nicht optional): `huh.Text` bringt einen EIGENEN,
huh-internen `ctrl+e`-Editor-Suspend-Mechanismus mit (Default `true`, `field_text.go`
verifiziert: `externalEditor: true` im Konstruktor) — der würde mit D01s app-weitem
`e`/`ctrl+e`-Ganz-Bean-Editor kollidieren (zwei verschiedene Editor-Sessions auf zwei
verschiedene Inhalte, dieselbe Taste, nur weil GENAU in diesem einen Formular fokussiert).
Deaktiviert, damit `keyForm`s eigene Tastatur-Semantik (forms_shared.go) die einzige
bleibt. `.Lines(3)` ist eine Planner-Schätzung (PO nannte keine Zeilenzahl) — deckt die
längsten bislang beobachteten Bean-Titel ab, ohne das Formular-Modal unnötig zu strecken
(`formInnerHeight`, forms_shared.go, deckelt ohnehin auf max. 20 Zeilen). `nonEmpty`-
Validator bleibt unverändert (Signatur ist identisch bei `huh.Text` wie bei `huh.Input` —
`Validate(func(string) error)`).

Prüfen (Implementer-Verantwortung, kein Code-Vorgriff hier): `huh.Text.GetValue()`/die
`Value(&v)`-Bindung liefert weiterhin einen reinen `string` (kein `[]string`/Zeilen-Slice)
— `submitForm`s `"editTitle"`-Case (box_confirm_create.go, `m.form.GetString("title")`)
sollte dadurch UNVERÄNDERT funktionieren, aber explizit gegenprüfen (Feldtyp-Wechsel kann
den `GetString`-Zugriff theoretisch brechen, falls huh intern anders keyed).

## TDD-Schritte

1. Failing test (form_edit_title_test.go): `TestBuildEditTitleFormUsesMultiLineText`
   (prüft Feldtyp/`.Lines(3)`-Konfiguration, nicht nur Verhalten — mirrort bestehende
   Formular-Konfigurationstests in diesem Repo, falls ein Muster existiert, sonst neu);
   `TestOpenEditTitleFormRoundTripsLongTitleUnchanged` (Drive-Test via echten `tea.Update`-
   Roundtrip, ANALOG den bestehenden `box_confirm_create_test.go`-Drive-Tests — Titel mit
   >60 Zeichen, Submit → `submitForm`s editTitle-Case liefert den vollen, unveränderten
   String zurück, kein Abschneiden/kein Zeilenumbruch-Zeichen im Ergebnis).
2. `command go test ./internal/tui/... -run "EditTitleForm"` → FAIL.
3. Implementieren.
4. `command go test ./internal/tui/... -run "EditTitleForm"` → PASS.
5. Golden-Check: das Formular selbst ist kein Golden-Ziel (Overlays sind nicht Teil von
   Tree/Backlog/Chrome-Goldens laut bestehender Konvention, E8-plan.md) — Gegenbeleg-Lauf
   `command go test ./internal/tui/ -run "TestTreeGolden|TestBacklogGolden|TestChromeGolden"`
   OHNE -update MUSS grün bleiben.
6. `command go test ./... -short` grün (2x, dieser Drive-Test zählt ggf. zu den 7 teuren
   huh-Drive-Tests aus `skipSlowHuhDriveInShortMode` — bei Bedarf denselben Skip-Mechanismus
   anwenden, CLAUDE.md-Konvention), voller Lauf grün (OHNE -short, Pflicht vor Commit),
   `-race` grün, gofmt/vet leer.
7. Commit `feat(tui): Titel-Edit-Form wird multi-line (B03)`. Footer `Refs: bt-tct9`.

## Akzeptanz-Checkliste

- [ ] Titel-Edit-Form nutzt `huh.NewText()` mit `.Lines(3)`, nicht mehr `huh.NewInput()`
- [ ] `.ExternalEditor(false)` gesetzt (keine Kollision mit D01s app-weitem `e`/`ctrl+e`)
- [ ] Ein langer Titel bricht sichtbar mehrzeilig um, statt horizontal zu scrollen/
      abzuschneiden
- [ ] `nonEmpty`-Validierung funktioniert unverändert (leerer Titel wird abgelehnt)
- [ ] Submit liefert den vollen, unveränderten String (kein Datenverlust durch den
      Feldtyp-Wechsel)
- [ ] Kein Golden ändert sich (Gegenbeleg grün)
- [ ] Voller Testlauf grün, gofmt/vet leer
