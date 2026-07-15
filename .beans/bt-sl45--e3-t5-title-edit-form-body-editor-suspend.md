---
# bt-sl45
title: E3 T5 — Title-Edit-Form + Body-$EDITOR-Suspend
status: todo
type: task
priority: high
created_at: 2026-07-15T00:26:52Z
updated_at: 2026-07-15T00:27:51Z
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
- [ ] internal/data/mutations.go: SetBody ergänzt (Spiegel von SetTitle, `--body`
      voller Replace), Test analog client_mut_test.go
- [ ] form_edit_title.go: einfeldriges huh-Formular (Title, nonEmpty-Validierung
      Port devd forms_shared.go nonEmpty), nutzt Task 4s styleForm/formChrome,
      SetTitle direkt bei Abschluss (kein Confirm-Gate)
- [ ] editor.go: prepareEditor/readEditorResult/editInEditor (Port devd editor.go
      44-96 verbatim bis auf editorBinary()), editorFinishedMsg-Handler ->
      SetBody (nur wenn changed==true, unveränderter Inhalt löst keine Mutation aus)
- [ ] editorBinary(): $VISUAL -> $EDITOR -> "vi", getestet ohne echten Editor-Launch
      (prepareEditor ist testbar ohne tea-Runtime, Port-Kommentar sagt das explizit)
- [ ] go test ./... grün, gofmt/vet leer
