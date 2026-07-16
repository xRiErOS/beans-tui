---
# bt-6bgn
title: 'Editor-Recovery: Vanished-Target verwirft Editor-Text (F04)'
status: in-progress
type: bug
priority: low
created_at: 2026-07-16T09:10:40Z
updated_at: 2026-07-16T09:10:51Z
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
