---
# bt-sl45
title: E3 T5 — Title-Edit-Form + Body-$EDITOR-Suspend
status: completed
type: task
priority: high
created_at: 2026-07-15T00:26:52Z
updated_at: 2026-07-15T03:31:20Z
parent: bt-gzcu
blocked_by:
    - bt-y4ly
---

Ziel: Title-Edit-Formular (`e`) + Body-$EDITOR-Suspend (`ctrl+e`) -- BEIDE unter
keys.Editor (design-spec §7: "e/ctrl+e Editor" ist EIN Binding, keymap.go hat
keine zweite). Unterscheidung per msg.String() innerhalb des gemeinsamen
keyNodeAction-Zweigs: "e" -> kleines huh-Einzelfeld-Formular (Port devd
buildEditFieldForm-Muster, NUR Title, kein Confirm-Gate -- Edits feuern direkt,
Port devd isCreateKind-Ausschluss), "ctrl+e" -> $EDITOR-Suspend auf Body (Port
devd editor.go prepareEditor/readEditorResult/editInEditor VERBATIM per
tea.ExecProcess).

editorBinary() Abweichung von devd: KEIN configuredEditor (das ist E5s
~/.config/beans-tui/config.yaml, existiert noch nicht) -- Auflösung
$VISUAL -> $EDITOR -> Fallback "vi" (POSIX-Default, portabler als devds "nvim"-
Annahme; design-spec §7 sagt wörtlich "$EDITOR").

GEKLÄRT (verifiziert via `beans update --help`): `-d/--body` ist ein VOLLER
Replace ("New body"), NICHT additiv wie `--body-append` (das bestehende
mutations.go AppendBody nutzt). Der $EDITOR-Flow braucht vollen Replace (Datei-
Inhalt komplett durch die editierte Version ersetzen) -- mutations.go bekommt
daher eine NEUE Funktion `SetBody(id, body, etag string) error` (Spiegel von
SetTitle, nutzt `--body` statt `--title`), NICHT AppendBody.

Plan: docs/plans/v1-port/epic-E3-plan.md »Task 5«.

## Akzeptanz
- [x] internal/data/mutations.go: SetBody ergänzt (Spiegel von SetTitle, `--body`
      voller Replace), Test analog client_mut_test.go
- [x] form_edit_title.go: einfeldriges huh-Formular (Title, nonEmpty-Validierung
      Port devd forms_shared.go nonEmpty), nutzt Task 4s styleForm/formChrome,
      SetTitle direkt bei Abschluss (kein Confirm-Gate)
- [x] editor.go: prepareEditor/readEditorResult/editInEditor (Port devd editor.go
      44-96 verbatim bis auf editorBinary()), editorFinishedMsg-Handler ->
      SetBody (nur wenn changed==true, unveränderter Inhalt löst keine Mutation aus)
- [x] editorBinary(): $VISUAL -> $EDITOR -> "vi", getestet ohne echten Editor-Launch
      (prepareEditor ist testbar ohne tea-Runtime, Port-Kommentar sagt das explizit)
- [x] go test ./... grün, gofmt/vet leer


## Übernommene Findings aus E3-T4-Review (PFLICHT)
- [x] B01 (Important): Draft-Verlust bei CLI-rejected Create — createDraft erst bei createDoneMsg-SUCCESS nullen (nicht bei enter im Confirm-Gate); bei Fehler Draft behalten + Weg zurück ins gefüllte Formular. Test dazu.
- [x] I01: driveFormBudget-Doc korrigieren (Budget saturiert auf 0, Korrektheit hängt an huh-Cmd-Ordering) + Assertion/loud-fail wenn Budget exakt 0 UND Werte fehlen
- [x] I02: TestFormCapturesAllKeys um ctrl+c-Fall ergänzen
- [x] D01 (Plan-Hygiene): stale 'createConfirm bool'-Zeile in epic-E3-plan.md »Geteilte Infrastruktur«-Sketch streichen (ERRATUM-Note)

## Abschluss

editor.go (NEU): editorFinishedMsg{content,changed,err} + editorBinary()
($VISUAL -> $EDITOR -> "vi", whitespace-only zählt als unset) +
prepareEditor/readEditorResult/editInEditor -- Port devd editor.go:14-96
VERBATIM bis auf editorBinary() (design decision c). glowRender NICHT
mitportiert -- existiert bereits in accordion.go (E2). form_edit_title.go
(NEU): buildEditTitleForm(title) *huh.Form (ein Input-Feld "title",
nonEmpty-Validierung) + openEditTitleForm(b) (m.mutTarget = b.ID, wieder-
verwendet aus der Value-Menü/Picker-Konvention statt eines neuen Feldes --
Formulare und Overlays sind wechselseitig exklusive Capture-Zustände, keine
Kollision). forms_shared.go: formTitle()-Switch um case "editTitle" ->
"Edit title" ergänzt. box_confirm_create.go: submitForm() um case
"editTitle" ergänzt -- liest GetString("title"), nullt Formular, liest
FRISCHEN etag via m.beanETag (Design-Entscheidung d), feuert
mutateCmd(SetTitle) DIREKT (kein Confirm-Umweg, Port devd isCreateKind-
Ausschluss: nur "create" ist gated). types.go: editorTarget-Feld (E3 Task 5)
ergänzt, dokumentiert warum "e" stattdessen mutTarget wiederverwendet.
update.go: editorFinishedMsg-Case + applyEditorFinished (err -> Statuszeile
ohne Mutation; unchanged -> stiller No-op; sonst SetBody gegen frischen etag,
vanished-target-Guard analog applyValueMenuSelection) + keyNodeAction-Editor-
Zweig (b := m.focusedBean() einmal gebunden, msg.String()=="ctrl+e" ->
editInEditor(b.Body, ".md"), sonst openEditTitleForm(b)).

Keine neuen Keybindings/keymap.go-Änderungen nötig -- keys.Editor
("e","ctrl+e") existierte bereits seit T1-Grundlage, helpGroups unverändert
grün.

PFLICHT-Findings aus T4-Review geschlossen:
- B01 (Draft-Verlust): keyCreateConfirm nullt createDraft NICHT mehr beim
  enter -- nur noch createDoneMsg (applyCreateDone) löst es auf: Erfolg
  konsumiert (nil), CLI-Fehler öffnet das BEFÜLLTE Formular erneut
  (openCreateFormWithDraft) statt über den draft-unkundigen
  applyMutationResult-Pfad zu laufen. Live gegen die echte CLI verifiziert
  (Smoke unten, nicht nur Unit-Test) -- epic mit task-Parent -> beans-CLI
  FILE_ERROR "epic beans can only have milestone as parent, not task" ->
  Formular öffnet mit Title/Type/Priority unverändert wieder.
  Tests: TestCreateConfirmEnterErrorPreservesDraftAndReopensForm (neu),
  TestCreateDoneSuccessConsumesDraft (neu), TestCreateConfirmEnterFiresPendingCmd
  (Assertion umgestellt: createDraft überlebt jetzt das enter).
- I01 (driveFormBudget-Doc + loud-fail): Doc-Kommentar um Budget-Saturierung
  (nie negativ) + Cmd-Ordering-Abhängigkeit ergänzt. Neue
  advanceFieldsExhausted/requireFieldValue-Helfer (form_create_bean_test.go)
  exponieren das Exhaustion-Signal; TestBuildCreateBeanFormPrefillsParentFromCursor
  als konkreter Schließungs-Beleg umgestellt (nicht alle T1-T4-Tests
  retrofittet -- bewusster Scope-Cut, siehe Deviations).
- I02 (ctrl+c-Testfall): TestFormCapturesAllKeysWhileOpen zu Subtests
  umgebaut ("q types into the field" unverändert, NEU "ctrl+c aborts the
  form, not the app" -- huh bindet ctrl+c selbst auf Quit/StateAborted,
  bricht das Formular ab statt die App zu beenden).
- D01 (Plan-Hygiene): stale `createConfirm bool`-Zeile im epic-E3-plan.md
  »Geteilte Infrastruktur«-Sketch mit ERRATUM-Kommentar gestrichen
  (verweist auf types.go's bereits existierenden Doc-Stamp).

Tests (neu/geändert): internal/data/client_mut_test.go
(TestSetBodyReplacesWholeBody), internal/tui/editor_test.go (NEU, 6 Tests:
editorBinary-Env-Kaskade, prepareEditor-Tempfile, readEditorResult
changed/unchanged, editorFinishedMsg unchanged/changed/vanished-target,
+ Dispatch-Test ctrl+e), internal/tui/form_edit_title_test.go (NEU: Prefill+
nonEmpty, Submit-ohne-Confirm, Dispatch-Test "e"), box_confirm_create_test.go
(B01/I02 wie oben), form_create_bean_test.go (I01 wie oben). `go test ./...`/
`-race` je 2x grün, gofmt/vet leer, Tree-Goldens 2x grün (Editor/Title-Form
berühren keinen Default-Render-Pfad).

Smoke (tmux, Scratch-Repo `bt-e3t5-smoke`, CLI-Warm-up per `beans list`
vorab, B01-Upstream-Präzedenz):
1. `e` auf dem einzigen Bean -> Edit-title-Formular mit "Smoke Task"
   vorbefüllt, Feld geleert + "Renamed via Smoke" getippt, enter ->
   Tree-Zeile UND Frontmatter (`.beans/*.md`) zeigen den neuen Titel.
2. EDITOR=<append-Skript, hängt deterministisch Text an -- kein
   interaktives vi in tmux nötig> gesetzt, `ctrl+e` -> Skript läuft,
   Body-Section im Detail-Pane zeigt den angehängten Text, Frontmatter-Body
   auf Disk bestätigt (System-Note beim Schreiben aufgefangen).
   Gegenprobe: EDITOR=true (No-Op) -> `ctrl+e` -> `updated_at` im
   Frontmatter UNVERÄNDERT (kein Mutations-Aufruf bei changed==false, live
   verifiziert, nicht nur im Unit-Test).
3. B01-Demo: `c` -> Type auf "epic" umgestellt, Parent-Feld vorbefüllt mit
   der einzigen (task-typisierten) Bean-ID im Repo -- clientseitig nur
   Existenz-validiert (Design-Entscheidung e), also gültig fürs Formular.
   Alle Felder durchlaufen -> Confirm-Gate ("Create? Bean (epic): Broken
   Parent Demo") -> enter -> beans-CLI lehnt serverseitig ab (FILE_ERROR:
   "epic beans can only have milestone as parent, not task") -> Formular
   öffnet SOFORT wieder, Title/Type/Priority unverändert -- kein
   verwaistes/kaputtes Bean im Baum, Draft real bewiesen erhalten (nicht
   nur im Unit-Test).

Deviations vom Plan:
- editorBinary()'s Whitespace-only-Guard (` `/`  ` zählt als unset) ist eine
  kleine Ergänzung ggü. dem reinen $VISUAL->$EDITOR->vi-Wortlaut, Port des
  devd-Vorbilds (strings.TrimSpace-Guard dort ebenfalls vorhanden) -- keine
  Abweichung von der Design-Entscheidung c, nur deren TrimSpace-Detail
  mitgenommen.
- I01 wurde NICHT auf alle bestehenden T1-T4-Formular-Tests retrofittet --
  advanceFields/driveForm/driveModel bleiben unverändert (Blast-Radius/
  Risiko vs. Nutzen), nur der explizit im Review genannte Prefill-Test
  (TestBuildCreateBeanFormPrefillsParentFromCursor) nutzt die neue
  loud-fail-Assertion als konkreter Schließungs-Beleg. Die Fähigkeit
  (advanceFieldsExhausted/requireFieldValue) steht für künftige Tests
  bereit.
- mutTarget (statt eines neuen Formular-Zielfeldes) für "e" wiederverwendet
  -- der Plan nennt in »Geteilte Infrastruktur« NUR editorTarget als neues
  T5-Feld; die Title-Form nutzt daher bewusst dasselbe Ein-Ziel-Feld wie
  Value-Menü/Picker (dokumentiert in form_edit_title.go/types.go).

Commit: siehe `Refs: bt-sl45` (feat(tui)).

## Notes für T6 (Delete-Confirm + ETag-Konflikt-Sweep, bt-ppzb)

keyNodeAction hat jetzt NUR noch den Delete-Stub übrig (`return true, m, nil
// stub: T6 (Delete)`) -- Editor/Status/Tag/Assign/Blocking sind alle
verdrahtet. composeOverlays braucht einen neuen `case overlayDeleteConfirm`
(overlayID existiert bereits seit T1, siehe types.go-Enum). Für den ETag-
Konflikt-Sweep: applyEditorFinished/submitForm("editTitle") lesen den etag
IMMER frisch via m.beanETag (Design-Entscheidung d) -- T6s Sweep kann diese
beiden Pfade also wie jeden anderen Mutationspfad behandeln, keine
Sonderbehandlung nötig. B01-Präzedenz (Draft-Erhalt bei CLI-Fehler) gilt
NUR für Create (hat einen Draft-Begriff) -- editTitle/SetBody haben keinen
Draft zu erhalten (Single-Field- bzw. Editor-Inhalt ist bereits weg/im
Editor-Prozess beendet), das ist kein Präzedenzfall für T6s Delete-Confirm.

## Korrektur (Review-Runde 2, F2)

Obiger Absatz ist für den `ctrl+e`-Pfad FALSCH und wurde durch die Review-
Runde-2-Fixes korrigiert: `applyEditorFinished` liest den etag NICHT mehr
frisch via `m.beanETag` bei Submit -- der `ctrl+e`-Pfad HAT eine
Sonderbehandlung. Begründung (ETag-Lost-Update): ein `$EDITOR`-Suspend kann
lange dauern; ein Watch-Reload WÄHREND der Suspend rotiert `m.idx`s
In-Memory-ETag unter der Hand, und ein frischer Read bei Submit hätte den
externen Edit dann still überschrieben (kein Konflikt wird je ausgelöst).
Fix: der etag wird jetzt in `keyNodeAction`s `ctrl+e`-Zweig (update.go) BEIM
ÖFFNEN eingefroren (`m.editorETag = b.ETag`, neues Feld, types.go), NICHT
beim Submit -- `m.beanETag(id)` wird in `applyEditorFinished` nur noch für
den `ok`-Bool (Bean-Präsenz) konsultiert, der zurückgegebene etag-Wert wird
verworfen. `submitForm("editTitle")` liest weiterhin frisch (Design-
Entscheidung d gilt dort unverändert -- das Formular-Fenster ist ein
einzelner Tastendruck, kein potenziell langes Suspend). T6s Sweep muss den
`ctrl+e`-Pfad also NICHT wie jeden anderen Mutationspfad behandeln -- er hat
jetzt einen eigenen ETag-Capture-Zeitpunkt; ein Regressionstest dafür lebt
bereits in `editor_test.go`
(`TestEditorFinishedUsesEtagCapturedAtOpenNotFreshIndexRead`). Nice-to-have
mitgeliefert: ein echter `ErrConflict` auf diesem Pfad schreibt den editierten
Body jetzt zusätzlich in ein aufbewahrtes Tempfile (`writeConflictTempFile`),
dessen Pfad in der Statuszeile landet (`conflictWithRecovery`,
`applyMutationResult`), statt den Edit kommentarlos zu verlieren.

## Review-Fixes (Runde 2)

Zwei Important-Findings aus dem E3-T5-Review geschlossen (TDD, RED vor Fix).

**1. Bugs**

| Bxx | Schwere | Beschreibung | Empfehlung | Status |
|-----|---------|--------------|------------|--------|
| B01 | high | Async-Gap-Clobbering: zwischen Confirm-Gate-Enter (`keyCreateConfirm`) und `createDoneMsg` ist `overlay=None`/`form=nil` -- ein zweites `c` oder ein anderes Overlay konnte in dieser Lücke geöffnet werden; `applyCreateDone`s Fehler-Zweig öffnete das Create-Formular dann UNBEDINGT neu, egal was der PO inzwischen tat -- Clobbering und Cross-Kontamination von `createDraft`/`pendingCreate`. | `pendingCreate` als In-Flight-Guard (bleibt bis `createDoneMsg` gesetzt, `keyNodeAction`s Create-Case + `submitForm("create")` lehnen ein zweites Create ab); Busy-Guard in `applyCreateDone` (Reopen NUR wenn `m.form==nil && m.overlay==overlayNone`, sonst Draft verwerfen + `applyMutationResult`). | 🟢 Done |
| B02 | high | ETag-Lost-Update auf `$EDITOR`-Pfad: `applyEditorFinished` las den etag bei Submit FRISCH aus `m.idx` -- ein Watch-Reload während einer langen `$EDITOR`-Session rotierte den etag unter der Hand, ein externer Edit wurde beim Speichern still überschrieben, kein Konflikt wurde je ausgelöst. | etag wird jetzt bei `ctrl+e`-Open eingefroren (`m.editorETag`, `keyNodeAction`); `applyEditorFinished` nutzt NUR noch den eingefrorenen Wert (`m.beanETag` dient nur noch dem Präsenz-Check). Nice-to-have: `ErrConflict` schreibt den editierten Body zusätzlich in ein aufbewahrtes Tempfile, Pfad landet in der Statuszeile (`conflictWithRecovery`/`writeConflictTempFile`). | 🟢 Done |

**4. Tasks**

| Txx | Prio | Aufgabe | Status |
|-----|------|---------|--------|
| T01 | high | update.go: `applyCreateDone` Busy-Guard + `pendingCreate` In-Flight-Semantik, `keyNodeAction`-Create-Guard, `createInFlightNote`. | 🟢 Done |
| T02 | high | box_confirm_create.go: `keyCreateConfirm`-Enter nullt `pendingCreate` nicht mehr; `submitForm("create")` eigener In-Flight-Guard. | 🟢 Done |
| T03 | high | types.go: `editorETag`-Feld + Doc-Stamps (`pendingCreate` Dual-Bedeutung, `editorETag`). | 🟢 Done |
| T04 | high | update.go: `keyNodeAction`s `ctrl+e`-Zweig setzt `m.editorETag = b.ETag`; `applyEditorFinished` nutzt `etag` statt frischem `m.beanETag`-Wert. | 🟢 Done |
| T05 | medium | update.go: `conflictWithRecovery`-Typ + `writeConflictTempFile`, `applyMutationResult` hängt Tempfile-Pfad an die Konflikt-Statuszeile. | 🟢 Done |
| T06 | high | Tests (RED zuerst): box_confirm_create_test.go 3 neue + 1 angepasst (Busy-Guard, In-Flight-Guard ×2, `pendingCreate`-Semantik); editor_test.go 2 neue Kern-Tests (`TestEditorFinishedUsesEtagCapturedAtOpenNotFreshIndexRead` via Fake-`beans`-Binary, `TestEditorFinishedConflictWritesRecoveryTempFileAndSurfacesPath`) + 3 bestehende angepasst. | 🟢 Done |
| T07 | low | Notes-für-T6-Korrektur in diesem Bean + surgical edit in epic-E3-plan.md (Task-5-Sektion + Task-6-Sweep-Vorschau). | 🟢 Done |

Verifikation: `command go test ./... -race -count=1 -timeout 300s` 2× grün (`beans-tui/cmd`,
`internal/data`, `internal/theme`, `internal/tui` -- alle ok, `internal/tui`
~124s wegen der bewusst langsamen no-binary-required-Pfade), `gofmt -l .`
leer, `command go vet ./...` leer, Tree-Goldens (`TestTreeGolden`,
`TestTreeGoldenDeterministic`) 2× grün.

Commit: siehe `Refs: bt-sl45` (fix(tui)).
