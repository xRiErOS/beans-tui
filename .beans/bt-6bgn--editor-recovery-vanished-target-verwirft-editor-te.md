---
# bt-6bgn
title: 'Editor-Recovery: Vanished-Target verwirft Editor-Text (F04)'
status: completed
type: bug
priority: low
created_at: 2026-07-16T09:10:40Z
updated_at: 2026-07-16T09:19:04Z
parent: bt-tct9
---

Review-Fund aus T2-Re-Review (bt-z4b1, 2026-07-16, Reviewer) — nicht blockierend, aber
gleiche Datenverlust-Klasse wie F01.

## Fund

`internal/tui/update.go`, `applyEditorFinished`, Vanished-Target-Guard: „Bean no longer
exists — editor edit discarded" verwirft den vollen $EDITOR-Text OHNE Recovery-Tempfile,
wenn das Bean während der Editor-Session extern gelöscht wurde. Pikant: OHNE den Guard
würde der CLI-Call fehlschlagen und der neue F01-Wrap (Commit 34c62c8) würde recovern —
der Guard verhindert die Recovery aktuell aktiv. Nicht vom Spec-Fehlerfall gedeckt
(PF-17 nennt nur Parse-Fehler + CLI-VALIDATION_ERROR), E3-Erbe, schmales Fenster.

## Fix-Rezept (Reviewer)

Im Vanished-Target-Zweig ebenfalls `writeConflictTempFile(msg.content)` schreiben und
den Pfad an die Warn-Meldung hängen — drei Zeilen, gleiche Konvention wie F01-Fix.
TDD: Test analog TestApplyEditorFinishedRecoversTempfileOnCLIValidationError, aber mit
verschwundenem Ziel-Bean; Tempfile-Existenz UND -Inhalt prüfen, RED-Zitat Pflicht.

## Akzeptanz

- [ ] Vanished-Target-Zweig schreibt Recovery-Tempfile, Pfad in der Warn-Meldung
- [ ] Neuer Test RED→GREEN belegt
- [ ] Bestehende Vanished-Target-Semantik (Warn statt Fehler) sonst unverändert
- [ ] Kein Golden ändert sich (Gegenbeleg)
- [ ] Volles Gate grün (voller Lauf, -race, gofmt/vet leer)


## Akzeptanz (final)

- [x] Vanished-Target-Zweig schreibt Recovery-Tempfile, Pfad in der Warn-Meldung
- [x] Neuer Test RED→GREEN belegt
- [x] Bestehende Vanished-Target-Semantik (Warn statt Fehler) sonst unverändert
- [x] Kein Golden ändert sich (Gegenbeleg)
- [x] Volles Gate grün (voller Lauf, -race, gofmt/vet leer)

## Summary

F04 behoben (Commit e1cc888): der Vanished-Target-Guard in applyEditorFinished (update.go) schreibt jetzt writeConflictTempFile(msg.content) und hängt den Pfad an die Warn-Meldung (m.err: "… — your version: <pfad>"), Toast-ctx trägt "Version saved: <pfad>" — exakt die F01-Konvention. Warn-Semantik unverändert (toastWarn, kein Fehler); der frühere toast.title==m.err-Spiegel wird zur Prefix-Relation (Title bleibt die nackte Note, m.err trägt zusätzlich den Pfad) — derselbe Title/ctx-Split wie Konflikt-Branch und F01-else-Zweig.

## Test-Output

RED (wörtlich): `editor_test.go:1019: status line = "Bean no longer exists — editor edit discarded", want it to carry the recovery tempfile's path ("your version: ") -- F04: a vanished target must not lose the PO's edits` → `--- FAIL: TestApplyEditorFinishedVanishedTargetRecoversTempfile (0.00s)`.

GREEN: `--- PASS: TestApplyEditorFinishedVanishedTargetRecoversTempfile (0.00s)` + `--- PASS: TestEditorFinishedTargetVanishedSurfacesError (0.00s)` (Warn-Semantik-Pin, Assertion auf Prefix-Relation angepasst). Volles Gate: `command go test ./...` grün (`ok beans-tui/internal/tui 137.559s`), `command go test ./internal/tui/ -race -count=1` grün (`ok 139.708s`), `gofmt -l .` leer, `command go vet ./...` leer. Golden-Gegenbeleg ohne -update: TestChromeGolden/TestTreeGolden/TestBacklogGolden (+Deterministic-Varianten) alle PASS.

Der neue Test prüft Tempfile-Existenz UND -Inhalt gegen den vollen Rohtext (analog TestApplyEditorFinishedRecoversTempfileOnCLIValidationError, aber mit verschwundenem Ziel-Bean — kein Fake-CLI nötig, der Guard feuert vor jedem CLI-Call).

## Smoke

Kein manueller Smoke (reiner Fehlerpfad-Fix ohne Render-Änderung, Golden-Gegenbeleg grün). Verhalten end-to-end über echtes model.Update() getestet: editorFinishedMsg gegen einen Index ohne das Ziel-Bean → Tempfile auf Platte gelesen und Inhalt byte-genau verifiziert.

## Deviations/ERRATA

1. **Dual-Write-Kontrakt angepasst (bewusst, Teil des Fixes):** TestEditorFinishedTargetVanishedSurfacesError pinnte toast.title == m.err; nach F04 trägt m.err zusätzlich den Tempfile-Pfad, der Toast-Title bleibt die nackte Note → Assertion auf strings.HasPrefix(m.err, toast.title) umgestellt. Kein Verhaltensverlust — der Pfad ist im Toast über ctx ("Version saved: …") weiterhin sichtbar, konsistent mit Konflikt-Branch und F01.
2. **writeConflictTempFile best-effort unverändert:** schlägt das Tempfile-Schreiben selbst fehl, bleibt die nackte Warn-Meldung ohne Pfad (kein Maskieren des eigentlichen Zustands) — gleiche Konvention wie im Parse-Fehler-Zweig.

## Notes

- Damit sind alle vier $EDITOR-Fehlerpfade recovery-gesichert: Parse-Fehler, CLI-VALIDATION_ERROR (F01), ETag-Konflikt, Vanished Target (F04). Der einzige nicht-recovernde Pfad bleibt msg.err (Editor-Prozess selbst fehlgeschlagen) — dort existiert kein verlässlicher Content, readEditorResult liefert bei runErr keinen (editor.go, short-circuit vor dem File-Read).
