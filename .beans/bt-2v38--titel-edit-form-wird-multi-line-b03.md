---
# bt-2v38
title: Titel-Edit-Form wird multi-line (B03)
status: completed
type: task
priority: normal
created_at: 2026-07-16T06:45:45Z
updated_at: 2026-07-16T09:47:18Z
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

- [x] Titel-Edit-Form nutzt `huh.NewText()` mit `.Lines(3)`, nicht mehr `huh.NewInput()`
- [x] `.ExternalEditor(false)` gesetzt (keine Kollision mit D01s app-weitem `e`/`ctrl+e`)
- [x] Ein langer Titel bricht sichtbar mehrzeilig um, statt horizontal zu scrollen/
      abzuschneiden
- [x] `nonEmpty`-Validierung funktioniert unverändert (leerer Titel wird abgelehnt)
- [x] Submit liefert den vollen, unveränderten String (kein Datenverlust durch den
      Feldtyp-Wechsel)
- [x] Kein Golden ändert sich (Gegenbeleg grün)
- [x] Voller Testlauf grün, gofmt/vet leer


## PRELUDE (2026-07-16, aus bt-6bgn-Review F05 — ZUERST erledigen, eigener Commit)

Ein-Zeilen-Test-Härtung, low, kein eigener Review-Zyklus: in
`internal/tui/editor_test.go`, `TestEditorFinishedTargetVanishedSurfacesError` ist die
Dual-Write-Assertion `strings.HasPrefix(nm.err, nm.toast.title)` für einen leeren
Toast-Title trivially true. Härtung: zusätzlich `nm.toast.title == ""` in die
Fail-Bedingung aufnehmen (oder Title gegen die Note-Konstante pinnen). Als eigener
kleiner Commit (`test(tui): harden vanished-target dual-write pin (F05)`,
`Refs: bt-2v38`) VOR der eigentlichen B03-Arbeit.



## Summary

`buildEditTitleForm` (`internal/tui/form_edit_title.go`) tauscht `huh.NewInput()` gegen
`huh.NewText().Key("title").Title("Title").Lines(3).ExternalEditor(false).Value(&v).Validate(nonEmpty)`
— exakt der Code-Sketch aus der Architektur-Vorgabe. `.ExternalEditor(false)` verhindert die
Kollision mit D01s app-weitem `e`/`ctrl+e`-Ganz-Bean-Editor. `submitForm`s `"editTitle"`-Case
(`box_confirm_create.go`, `m.form.GetString("title")`) blieb unverändert — `huh.Text.GetValue()`
liefert weiterhin einen reinen `string`, gegengeprüft via Test. Zwei neue Tests in
`form_edit_title_test.go`: `TestBuildEditTitleFormUsesMultiLineText` (Feldtyp-Assertion
`*huh.Text` + Zeilen-Delta-Vergleich gegen ein äquivalentes `*huh.Input` als Beweis für
`.Lines(3)`, da `huh.Text` keinen exportierten Height-Getter hat) und
`TestOpenEditTitleFormRoundTripsLongTitleUnchanged` (>60-Zeichen-Titel, unbearbeiteter Submit,
`GetString("title")` exakt gleich + kein eingebettetes `\n`, plus echter `submitForm()`-Aufruf
gegen einen nicht-existenten `RepoDir` als Dispatch-Beweis, Konvention dieser Testdatei).

PRELUDE (F05, aus bt-6bgn-Review) zuerst erledigt: `TestEditorFinishedTargetVanishedSurfacesError`
in `editor_test.go` härtet die Dual-Write-Assertion um einen expliziten
`nm.toast.title == ""`-Check, da `strings.HasPrefix(nm.err, nm.toast.title)` für einen leeren
Toast-Title trivially true wäre.

## Test-Output

RED (vor Implementierung, `command go test ./internal/tui/... -run "EditTitleForm" -v`):

```
=== RUN   TestBuildEditTitleFormUsesMultiLineText
    form_edit_title_test.go:108: field type = *huh.Input, want *huh.Text (B03: huh.NewInput -> huh.NewText)
--- FAIL: TestBuildEditTitleFormUsesMultiLineText (0.00s)
...
FAIL
FAIL	beans-tui/internal/tui	0.471s
FAIL
```

GREEN (nach Implementierung, dieselbe -run-Filterung):

```
=== RUN   TestEditTitleFormPrefilledAndNonEmptyValidated
--- PASS: TestEditTitleFormPrefilledAndNonEmptyValidated (0.00s)
=== RUN   TestBuildEditTitleFormUsesMultiLineText
--- PASS: TestBuildEditTitleFormUsesMultiLineText (0.00s)
=== RUN   TestOpenEditTitleFormRoundTripsLongTitleUnchanged
--- PASS: TestOpenEditTitleFormRoundTripsLongTitleUnchanged (0.00s)
=== RUN   TestKeyNodeActionEDoesNotOpenEditTitleFormAnymore
--- PASS: TestKeyNodeActionEDoesNotOpenEditTitleFormAnymore (0.00s)
=== RUN   TestKeyDetailFocusEnterOnTitleFieldOpensEditTitleForm
--- PASS: TestKeyDetailFocusEnterOnTitleFieldOpensEditTitleForm (0.00s)
=== RUN   TestActivateDetailFieldTitleOpensEditTitleForm
--- PASS: TestActivateDetailFieldTitleOpensEditTitleForm (0.00s)
PASS
ok  	beans-tui/internal/tui	0.439s
```

Golden-Gegenbeleg (ohne `-update`, `command go test ./internal/tui/ -run "TestTreeGolden|TestBacklogGolden|TestChromeGolden"`): `PASS` alle 5 Subtests (Tree/TreeDeterministic/Backlog/BacklogDeterministic/Chrome), kein Golden geändert.

Commit-Gate: `command go test ./...` → alle Packages `ok` (`internal/tui` 137.5s); `command go test ./internal/tui/ -race` → `ok` (140.0s); `gofmt -l .` → leer; `command go vet ./...` → leer.

## Smoke

tmux (100×30), `bin/bt` in diesem Repo, Bean `bt-apmy` (realer Titel ~40 Zeichen). `Tab` → `Enter`
→ `Enter` (title-Feld, Index 0) öffnet das "Edit title"-Modal: sichtbar 3 reservierte Textarea-
Zeilen (statt vormals 1 Input-Zeile). Zusätzlicher langer Text getippt → Titel bricht sichtbar
über alle 3 Zeilen um (kein horizontales Scrollen/Abschneiden). `esc` bricht ab — `git status
--short .beans/` bestätigt: keine Mutation am Repo (Draft korrekt verworfen, kein versehentlicher
Save gegen echte Daten).

## Deviations/ERRATA

Keine. Umsetzung deckt sich 1:1 mit dem Code-Sketch aus der Architektur-Vorgabe; keine
Design-Lücke, keine Abweichung von den TDD-Schritten.

## Notes for T(n+1)

- `huh.Text` hat keinen exportierten Height-Getter — falls künftige Tasks `.Lines(N)` an anderer
  Stelle prüfen wollen: der robuste Weg ist der Zeilen-Delta-Vergleich gegen ein äquivalentes
  `*huh.Input` (siehe `TestBuildEditTitleFormUsesMultiLineText`), nicht Reflection auf huhs
  unexportiertes `textarea`-Feld.
- `huh.Text`s Submit-Binding ist weiterhin plain `enter` (NewLine liegt separat auf
  `alt+enter`/`ctrl+j`, `keymap.go` verifiziert) — bestehende `driveForm(f, enterMsg())`-Aufrufe
  in diesem Testfile funktionieren dadurch unverändert, keine Testinfra-Anpassung nötig.
- Diese Repo-Testdatei hat keinen PATH-gestubbten `beans`-Binary zum Abfangen echter CLI-Args —
  Konvention ist der `strings.Contains(err, "beans update")`-Dispatch-Beweis gegen einen
  nicht-existenten `RepoDir`. Für künftige Tasks, die exakte CLI-Argumente prüfen wollen, wäre
  ein PATH-Stub ein separates, größeres Infra-Stück (hier bewusst nicht gebaut, außerhalb B03-Scope).
