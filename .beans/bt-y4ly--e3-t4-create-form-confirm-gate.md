---
# bt-y4ly
title: E3 T4 — Create-Form + Confirm-Gate
status: completed
type: task
priority: high
created_at: 2026-07-15T00:26:37Z
updated_at: 2026-07-15T02:44:11Z
parent: bt-gzcu
blocked_by:
    - bt-p1uz
---

Ziel: Create-Form (`c`, huh v1.0.0) + Confirm-Gate. Feldsatz: Title(Input,required)/
Type(Select, data.TypeValues())/Priority(Select, data.PriorityValues())/
Status(Select, data.StatusValues())/Parent(Input, optional, vorbelegt aus
Tree-Cursor-Kontext, NUR Existenz-validiert)/Tags(Input, Leerzeichen/Komma-
getrennt, regex-validiert)/Body(Text, optional). Confirm-Gate (Port devd
box_confirm_create.go submitForm/createConfirmBox) vor dem echten Create;
n/esc am Confirm kehrt ins BEFÜLLTE Formular zurück (Draft-Erhalt, Port DD2-190).
Nach Erfolg: Cursor springt auf das neue Bean (Ancestor-Pfad seines Parents
expandiert via bestehendem expandAncestorsOf, cursorID gesetzt VOR dem Reload
-- applyLoadeds bestehende "exact bean still present"-Logik greift dann
unverändert).

Parent-Feld bewusst KEIN Picker (anders als Task 3s `a`): ein frisch angelegtes
Bean hat noch keine Nachkommen, Zyklen-Ausschluss ist an dieser Stelle strukturell
bedeutungslos. Typ-Hierarchie-Fehler (z.B. Task als Parent eines Epics) werden
NICHT clientseitig vorgefiltert, sondern laufen in den bestehenden
VALIDATION_ERROR-Pfad (classifyError, mutations.go) -> Statuszeile.

Bringt außerdem die geteilte Form-Hosting-Infra (m.form/m.formKind/styleForm/
formChrome/paletteFormTheme, Port devd forms_shared.go), die Task 5 fürs
Title-Edit-Formular wiederverwendet.

Plan: docs/plans/v1-port/epic-E3-plan.md »Task 4«.

## Akzeptanz
- [x] go.mod/go.sum: huh v1.0.0 direkte Dependency (Modul-Cache vorhanden,
      kein Netzwerkzugriff nötig)
- [x] forms_shared.go: styleForm/formChrome/paletteFormTheme (Theme-Mapping auf
      theme.Mauve/Yellow/Base/Red/Text/Hint, Port devd forms_shared.go:164-225)
- [x] form_create_bean.go: buildCreateBeanForm(idx, prefillParent string) *huh.Form,
      beanDraft-Struct für Draft-Erhalt
- [x] box_confirm_create.go: submitForm/createConfirmBox (Port devd 1:1-Muster,
      NUR "create" als isCreateKind -- kein zweiter Kind-Typ hier)
- [x] Nach Confirm: data.Client.Create, Cursor+Expand-Handling wie oben beschrieben,
      go test deckt den vollen Flow inkl. n/esc-Draft-Rückkehr ab
- [x] go test ./... grün, gofmt/vet leer

## Abschluss

huh v1.0.0 als eingebettetes Sub-Modell (kein form.Run()) -- Werte keyed per
GetString nach huh.StateCompleted (kein Pointer-Binding über Model-Copies).
forms_shared.go: nonEmpty/formInnerWidth/Height/styleForm/paletteFormTheme/
formChrome/updateForm/keyForm (Port devd 1:1, Theme-Tokens Mauve/Yellow/Base/
Red/Text/Hint -- alle bereits im theme-Paket vorhanden, kein neuer Token
nötig). form_create_bean.go: beanDraft/buildCreateBeanForm/parseTagsField/
parentFieldValidator/createOptsFromDraft/openCreateForm(WithDraft) --
7 keyed Felder (title/type/priority/status/parent/tags/body), Selects aus
der E3-Enum-Single-Source (data.TypeValues/PriorityValues/StatusValues),
Body mit ExternalEditor(false) (DD2-234-Falle, auf Prinzip gesetzt auch
ohne zweites Text-Feld). box_confirm_create.go: submitForm parkt den
fertigen createCmd + Draft, öffnet overlayCreateConfirm; keyCreateConfirm:
enter feuert (+ Draft verworfen), n/esc -> openCreateFormWithDraft (Draft-
Erhalt, DD2-190) -- go.mod trägt huh v1.0.0 jetzt als direkte Dependency
(go get + go mod tidy, Modul-Cache lokal vorhanden, kein Netzwerkzugriff
nötig). Nach Create: applyCreateDone (update.go) expandiert den Ancestor-
Pfad des NEUEN Beans (muss von msg.bean.Parent aus laufen, nicht von
msg.bean.ID -- m.idx ist zum Zeitpunkt des Handlers noch der ALTE Index,
der neue Bean existiert darin noch nicht) und setzt cursorID VOR dem
Reload; applyLoadeds bestehender "exact bean still present"-Pfad hält den
Cursor danach unverändert auf dem neuen Bean.

ERRATUM (dokumentiert in types.go, Konvention aus dem Task-Auftrag
übernommen): der epic-weite "Geteilte Infrastruktur"-Codeblock nennt
zusätzlich ein Modell-Feld `createConfirm bool` -- das widerspricht
Design-Entscheidung a2 (EIN overlayID-Enum statt weiterer Bools) UND Task
4s eigenem Step-4-Pseudocode (`overlay = overlayCreateConfirm`, nie
`createConfirm = true`). Kein separates Bool-Feld angelegt; der
Confirm-Gate-Zustand IST `m.overlay == overlayCreateConfirm`.

Tests: form_create_bean_test.go (Prefill/parseTagsField/
parentFieldValidator/createOptsFromDraft) + box_confirm_create_test.go
(submitForm-Park+Open, Confirm-enter/n-esc inkl. Draft-Rückkehr,
createDoneMsg Erfolg/Fehler, q-Capture bei offenem Formular) -- beide neu.
`go test ./... `/`-race` je 2x grün, gofmt/vet leer, Goldens (Chrome/Tree/
Backlog) 2x grün (Create-Form/-Confirm berühren keinen bestehenden
Default-Render-Pfad). Notiz zur Test-Infrastruktur: huh v1.0.0 lässt sich
NICHT über ein bloßes `f.Update(msg)` bis StateCompleted treiben (kein
form.Run() in Tests) -- ein eigener driveForm/driveModel-Helper führt
Cmd-Ketten inkl. tea.BatchMsg-Fan-out synchron aus. FALLE dabei gefunden
(nicht im Plan erwähnt): das Fokussieren eines Input-Feldes (huh
field_input.go Focus -> bubbles textinput.Focus) startet einen
SELBSTERHALTENDEN Cursor-Blink-Timer (tea.Tick-basiert) -- ein
ungebremster Cmd-Chase hängt sich daran auf (leer nachgewiesen). Fix:
festes Rundenbudget (driveFormBudget=6) je Cmd-Kette, das die Blink-Kette
nach kurzer Zeit hart abbricht (Feldwerte/Fokus davon unberührt).

Smoke (tmux, Scratch-Repo mit CLI-Warm-up: milestone + epic via `beans
create` vorab angelegt, B01-Präzedenz): Cursor auf Epic, `c` -> Formular
mit Type=task/Priority=normal/Status=todo vorbelegt, Parent-Feld zeigt die
Epic-ID vorbefüllt; Title+Tags getippt, alle Felder durchlaufen ->
Confirm-Gate ("Create? Bean (task): <Titel>"); `n` -> Formular öffnet
erneut MIT allen zuvor eingegebenen Werten (Title/Parent/Tags); erneut
durchlaufen -> Confirm-Gate -> `enter` -> `.md`-Datei angelegt (Parent/
Tags/Status/Priority/Type korrekt im Frontmatter), Baum zeigt das neue
Bean unter dem Epic, Cursor sitzt darauf.

Commit: siehe `Refs: bt-y4ly` (feat(tui)).

Notes für T5 (Title-Edit-Form, bt-sl45): styleForm/formChrome/
paletteFormTheme sind kind-agnostisch -- T5 ergänzt nur formTitle() um den
"editTitle"-Case + eine eigene buildEditTitleForm (einfeldrig). WICHTIG:
submitForm (box_confirm_create.go) dispatcht aktuell NUR auf
m.formKind=="create" -- T5 MUSS dort einen zweiten case "editTitle"
ergänzen, der OHNE Confirm-Gate direkt mutateCmd(SetTitle) feuert (Design-
Entscheidung h: isCreateKind-Analogon, nur "create" ist gated). keyForm
(forms_shared.go) ist bereits formKind-agnostisch (esc bricht IMMER ab,
kein Unterschied zwischen create/editTitle) -- keine Änderung dort nötig.
