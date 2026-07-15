---
# bt-ppzb
title: E3 T6 — Delete-Confirm + ETag-Konflikt-Regression-Sweep
status: completed
type: task
priority: high
created_at: 2026-07-15T00:28:07Z
updated_at: 2026-07-15T04:47:56Z
parent: bt-gzcu
blocked_by:
    - bt-sl45
---

Ziel: Delete-Confirm (`d`) mit Kinder-Count-Preview + ETag-Konflikt-Regression-
Sweep (Cross-Cutting-Verifikation über ALLE Mutations-Stellen aus Task 1-6, nicht
nur diese Task).

Kinder-Count-Preview ist SYNCHRON (idx.Children[id] ist schon im Speicher, KEIN
async Load wie devd box_confirm_delete.go loadDeletePreview). Semantik-Abweichung
von devd bewusst dokumentieren: `beans delete` kaskadiert NICHT (mutations.go
Delete-Doc-Kommentar: "--json skips ... any reference/child warnings", KEINE
Cascade-Delete-Funktion vorhanden) -- Kinder werden beim Löschen des Parents
NICHT mitgelöscht, sondern zu Waisen (dangling Parent), die die bestehende
"(verwaist)"-Root-Mechanik bereits korrekt rendert. Die Preview ist also eine
WARNUNG ("N Kinder werden verwaist"), kein "wird mitgelöscht"-Cascade-Count wie
bei devd.

ETag-Konflikt-Sweep: für JEDE in Task 1-6 gebaute Mutations-Stelle (Status/Type/
Prio-Menü, Tag-Picker, Parent-Picker, Blocking-Picker, Create, Title-Edit, Body-
$EDITOR, Delete) ein Regressionstest, der einen veralteten ETag simuliert (Reload
zwischen Öffnen und Submit) und verifiziert: ErrConflict -> Statuszeile +
unconditional Reload (Task 1s applyMutationResult), NIE ein Absturz/Silent-Drop.

Plan: docs/plans/v1-port/epic-E3-plan.md »Task 6«.

## ERRATUM (während T6 empirisch widerlegt, siehe Deviations)

Die obige "Kinder werden zu Waisen ... (verwaist)-Root" Annahme (Ziel-Absatz +
ursprünglicher Akzeptanz-Wortlaut) ist FALSCH -- gegen die reale beans-0.4.2-CLI
verifiziert (tmux-Smoke + isolierte `beans create`/`beans delete`-Probe, beide
reproduzierbar): `beans delete` räumt das `parent:`-Feld der direkten Kinder
komplett AUS, statt es dangling stehen zu lassen. Die Kinder werden zu
gewöhnlichen, elternlosen WURZEL-Beans (idx.Roots()) -- landen NICHT im
synthetischen "(verwaist)"-Bucket. deleteBox-Text korrigiert: "N Kind(er)
verlieren den Parent — werden zu eigenen Wurzeln". Volle Doc-Kette:
box_confirm_delete.go (Datei-Header), types.go (delChildren-Feld),
mutations.go (Delete()), Regressionstest
internal/data/client_mut_test.go:TestDeleteClearsFormerChildrensParentField.

## Akzeptanz
- [x] box_confirm_delete.go: deleteBox zeigt Typ+Titel+Kinder-Count (idx.Children,
      synchron, KEIN Loading-State), Rot=destruktiv (Port modalBox theme.Red wie
      quitBox mit theme.Mauve)
- [x] enter löscht (Client.Delete, KEIN --if-match auf delete -- CLI-Signatur hat
      kein --if-match für delete, siehe mutations.go Delete()), esc/n bricht ab
- [x] Nach Löschen: Cursor-Klemmung wie applyLoaded's bestehende oldPos-Fallback-
      Logik (E1 Task 8/E2) -- KEINE neue Cursor-Logik, Wiederverwendung verifiziert
- [x] ETag-Konflikt-Regressionstest je Mutations-Stelle aus Task 1-6 (Tabelle im
      Plan, mind. 8 Fälle), alle grün -- 9 Sites (value-menu status/type/priority,
      tag-picker SetTags, parent-picker SetParent/RemoveParent, blocking-picker
      SetBlocking, edit-title SetTitle, editor SetBody) + Gegenprobe
      TestConflictAfterWatchReloadUsesFreshETagNoConflict
- [x] go test ./... -count=1 (2x) grün, gofmt/vet leer, go build -o bin/bt . ok


## Übernommene Findings aus E3-T3-Review (in Sweep aufnehmen)
- [x] Vanished-mutTarget-Regressionstests für Parent- UND Blocking-Picker (Muster TestValueMenuTargetVanished…)
- [x] Test: Blocking-Zyklus via CLI → VALIDATION_ERROR-Shape gepinnt (bisher nur Kommentar)
- [x] Test: current-parent-out-of-eligibility → Cursor-Fallback 0
- [x] Doc-Note Picker-Stil-Divergenz (Header/▸ vs Accent/▌) in types.go o.ä.

## NEU (Kosten-Finding, außerhalb ursprünglicher Akzeptanz)
- [x] `internal/tui`-Testsuite war auf ~121-123s gewachsen (7 huh-drive-Tests in
      box_confirm_create_test.go, je ~16-19s wegen huhs selbst-perpetuierender
      Blink-Tick-Cmds über echte tea.Update-Roundtrips). testing.Short()-Guards
      (skipSlowHuhDriveInShortMode) NUR auf diese 7 -- `go test ./... -short`
      jetzt ~5s statt ~121s. Dokumentiert in CLAUDE.md ("Schneller Lauf").

## Summary

box_confirm_delete.go (NEU): openDeleteConfirm (delTitle/delChildren aus
idx.Children[id], synchron, DIREKTE Kinder nur -- NICHT
data.CollectDescendants) + keyDeleteConfirm (enter -> mutateCmd(Client.Delete),
KEIN --if-match; esc/n Abbruch) + deleteBox (Rot/theme.Red, Typ aus
m.idx.ByID[m.mutTarget] zur Render-Zeit). update.go: keyNodeAction-Delete-Zweig
ersetzt den T5-Stub, keyOverlay-Case overlayDeleteConfirm. view_browse_repo.go:
composeOverlays-Case + Footer-Hints final um keys.Create/Delete/Editor ergänzt
(view_browse_backlog.go analog). keymap.go: irreführendes "Delete (cascade)"
korrigiert zu "Delete" (kaskadiert nie, siehe ERRATUM).

etag_conflict_test.go (NEU): TestEtagConflictSweep (9 Subtests je Mutations-
Stelle aus T1-T5, gemeinsamer assertConflictSweep-Helper: ErrConflict ->
Statuszeile "Konflikt") + TestConflictAfterWatchReloadUsesFreshETagNoConflict
(Gegenprobe Design-Entscheidung d, gilt NICHT für den editor-Pfad). Editor/
Edit-Title-Subtests nutzen bewusst den FORM-Level-Chase (driveForm) statt des
teuren Model-Level-Keystroke-Drives, um keine neuen langsamen Tests
einzuführen.

Geerbte T3-Findings: TestParentPickerEnterTargetVanishedClosesGracefully +
TestBlockingPickerEnterTargetVanishedClosesGracefully (Muster
TestValueMenuTargetVanishedClosesGracefully). TestAddBlockingCyclePins
ValidationErrorShape (internal/data/client_mut_test.go, gegen echte CLI,
JSON-Shape empirisch verifiziert: code VALIDATION_ERROR, NICHT ErrConflict).
TestParentPickerCurrentParentOutOfEligibilityCursorFallsBackToZero (bereits
korrektes Verhalten, jetzt gepinnt). Picker-Stil-Divergenz-Doc-Note in
types.go bei overlayID (Header/▸ vs. D08 Accent/▌, Verweis auf
parentPickerBox's volle Begründung).

ERRATUM (Kern-Finding dieser Task): tmux-Smoke deckte auf, dass die
Plan-Annahme "Kinder werden zu (verwaist)-Waisen" falsch ist -- beans 0.4.2
räumt das parent-Feld der direkten Kinder beim Delete aus, sie werden zu
normalen Wurzel-Beans. deleteBox-Text, alle betroffenen Doc-Kommentare
(box_confirm_delete.go, types.go, mutations.go Delete()) und die
Test-Assertions korrigiert; neuer Regressionstest
TestDeleteClearsFormerChildrensParentField pinnt das reale Verhalten.

Tests (neu/geändert): internal/tui/box_confirm_delete_test.go (NEU, 10 Tests),
internal/tui/etag_conflict_test.go (NEU, Sweep + Gegenprobe),
internal/tui/box_picker_parent_test.go (+2), internal/tui/box_picker_
blocking_test.go (+1), internal/data/client_mut_test.go (+2:
TestAddBlockingCyclePinsValidationErrorShape,
TestDeleteClearsFormerChildrensParentField), internal/tui/box_confirm_
create_test.go (7x testing.Short()-Guard). go test ./... -count=1 2x grün,
-race grün, gofmt/vet leer, Goldens (tree.golden/backlog.golden) 2x grün
neu generiert (Footer-Hint-Erweiterung, einzeilige Diffs).

Smoke (tmux, Scratch-Repo, Warm-up per beans list vorab): d auf Epic mit 2
Kindern -> Confirm zeigt "2 Kind(er) verlieren den Parent — werden zu eigenen
Wurzeln" -> enter -> Datei weg, Kinder als Wurzel-Beans sichtbar (parent-Feld
auf Disk bestätigt entfernt). d auf Blatt -> "Delete task", keine
Kinder-Zeile -> enter -> Datei weg, Cursor auf Nachbar geklemmt (bestehender
oldPos-Pfad, live verifiziert).

Deviations vom Plan:
- ERRATUM oben (Kinder-Wurzel statt Kinder-Waise) -- deleteBox-Copy weicht vom
  Plan-Wortlaut ("N Kinder werden verwaist") ab, aus empirischer Notwendigkeit.
- keymap.go's "Delete (cascade)"-Label korrigiert (nicht im Plan-Dateien-Scope
  gelistet, aber direkt widersprüchlich zur jetzt sichtbaren Footer-Zeile --
  kleiner, begründeter Fix im selben Commit).
- testing.Short()-Guards sind ein NEUER Task-Bean-Auftrag (Kosten-Finding),
  nicht Teil der ursprünglichen bt-ppzb-Akzeptanz -- oben als eigener
  Abschnitt geführt statt in die bestehenden Boxen gemischt.

## Notes für T7 (Abschluss, bt-qzwt)

E3 vollständig verdrahtet: V7 (Create/Edit/$EDITOR) + V8-Mutationsteil
(Status/Type/Prio, Tag/Parent/Blocking, Delete) + ETag-Konflikt-Handling
überall (T6-Sweep, 9 Sites + Gegenprobe). Kein offener Stub mehr in
keyNodeAction. composeOverlays trägt alle 6 overlayID-Cases. Footer-Hints in
beiden Views vollständig (Create/Delete/Editor final ergänzt). Kein neues
Keybinding nötig (keys.Delete existierte seit T1) -- helpGroups-Drift-Guard
unverändert grün. Für den T7-Smoke-Durchlauf: das Delete-Copy-ERRATUM ist
bereits live verifiziert (siehe Smoke oben), T7 muss das NICHT erneut
gegenprüfen, nur den Rest des kombinierten Flows (s/t/a/B/c/e/d nacheinander).
