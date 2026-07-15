---
# bt-y4ly
title: E3 T4 — Create-Form + Confirm-Gate
status: in-progress
type: task
priority: high
created_at: 2026-07-15T00:26:37Z
updated_at: 2026-07-15T02:00:20Z
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
- [ ] go.mod/go.sum: huh v1.0.0 direkte Dependency (Modul-Cache vorhanden,
      kein Netzwerkzugriff nötig)
- [ ] forms_shared.go: styleForm/formChrome/paletteFormTheme (Theme-Mapping auf
      theme.Mauve/Yellow/Base/Red/Text/Hint, Port devd forms_shared.go:164-225)
- [ ] form_create_bean.go: buildCreateBeanForm(idx, prefillParent string) *huh.Form,
      beanDraft-Struct für Draft-Erhalt
- [ ] box_confirm_create.go: submitForm/createConfirmBox (Port devd 1:1-Muster,
      NUR "create" als isCreateKind -- kein zweiter Kind-Typ hier)
- [ ] Nach Confirm: data.Client.Create, Cursor+Expand-Handling wie oben beschrieben,
      go test deckt den vollen Flow inkl. n/esc-Draft-Rückkehr ab
- [ ] go test ./... grün, gofmt/vet leer
