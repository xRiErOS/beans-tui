---
# bt-idm1
title: 'Tag-Management: n auf freier Zeile registriert bestehenden Tag (Adopt statt Blank-Create)'
status: completed
type: feature
priority: normal
created_at: 2026-07-16T20:35:29Z
updated_at: 2026-07-16T21:44:21Z
parent: bt-362n
blocked_by:
    - bt-ct3k
---

PO-Nebenbefund, US-Review Runde 8 (2026-07-16), PO-Zitat: "Ich kann bestehende
tags nicht neu registrieren. Wenn ich ein unregistrierten tag waehle und 'n'
druecke, dann sollte dieser automatisch registriert werden, um diesen dann
bearbeiten zu koennen."

## Ist-Stand (E10, D11)

'n' oeffnet IMMER den Blank-Create-Input-Submodus, unabhaengig vom Cursor --
Registrieren eines bereits existierenden freien Tags erfordert den exakt
gleichen Namen erneut abzutippen (Dedupe erlaubt das zwar, ist aber
unnoetiger Umweg gegenueber "Cursor steht schon drauf").

## Gewuenschtes Verhalten (Entwurf, Planner praezisiert vor Umsetzung)

- Cursor auf einer FREIEN Tag-Zeile + n => dieser Tag wird DIREKT registriert
  (Registry-Add mit dem Zeilennamen), OHNE erneute Texteingabe noetig.
- Cursor auf KEINER Zeile bzw. es gibt (noch) keine Zeilen => n verhaelt sich
  wie heute (Blank-Create-Input).
- Offene Planner-Frage: soll n auf einer freien Zeile trotzdem kurz den
  Input-Submodus zeigen (vorbefuellt, Bestaetigung per enter) oder sofort ganz
  ohne Zwischenschritt registrieren? PO-Wortlaut ("sollte automatisch
  registriert werden") liest sich eher wie EIN Tastendruck ohne Zwischenstopp
  -- Planner legt bei Umsetzung fest, ggf. kurze Rueckfrage.

Quelle: bt-362n US-Review Runde 8.



## Planner-Konkretisierung (2026-07-16)

**Planner-Entscheidung (offene Frage aus Bean-Entwurf, hiermit final):**
DIREKTES Registrieren OHNE Zwischenschritt (nicht vorbefüllter Input +
Enter). Begründung: PO-Wortlaut "sollte automatisch registriert werden" +
"ein Tastendruck ohne Zwischenstopp" liest sich eindeutig als
Ein-Schritt-Aktion. Gegen die Alternative (vorbefüllter Input): der
Input-Submodus existiert für FREITEXT-Eingabe (neuer Name/Rename) — hier
steht der Name schon exakt fest (der Zeilenname), ein Zwischenschritt wäre
reine Bestätigungs-Zeremonie ohne Informationsgewinn, und `d` (Delete) hat
bereits ein Confirm-Modal als etabliertes Muster für "Bestätigung nötig" —
Registrieren ist NICHT destruktiv (kein Datenverlust-Risiko wie bei
Delete), daher kein Confirm-Bedarf. Feedback via Toast (analog bt-ct3k,
Konsistenz: da bt-ct3k gerade das "stiller No-Op fühlt sich kaputt an"-
Problem behebt, wäre ein neuer stiller Erfolgspfad hier ein Rückschritt).

**Betroffene Stelle:** `keyTagManagement` (view_tag_management.go:373-374),
`case keybind.Matches(msg, keys.NewTag): return m.openTagMgmtInput("create",
"")`. Neue Logik (mirrort `openTagMgmtRename`s Cursor-Row-Check, Zeile
412-421): wenn `m.tagMgmtCursor.cursor` eine gültige Zeile referenziert UND
`!row.defined` → neue Funktion `m.openTagMgmtAdopt()` (Namensvorschlag)
dispatcht `saveTagDefsCmd(m.client, data.AddTagDefName(definedTagNames(
m.tagMgmtRows), row.name), row.name)` (identische Save-Infra wie
Create/Delete, `messages.go:374`) OHNE `openTagMgmtInput` zu betreten;
Erfolgs-Toast `showToast(toastInfo, "tag '<name>' registered", "", nil,
false)` in `applyTagDefsSaved` (update.go:472, bestehender Dispatch-Punkt
nach bestätigtem Write — dort den Erfolgstoast anhängen, mirrort dessen
bereits vorhandene Fehler-Toast-Kante). Wenn Cursor UNGÜLTIG (keine Zeilen)
ODER Zeile bereits `defined` → Fallback auf heutiges Verhalten
(`openTagMgmtInput("create", "")`).

**Akzeptanzkriterium:**
- Cursor auf freier Zeile + `n` → Tag DIREKT registriert, kein
  Input-Submodus sichtbar, Erfolgs-Toast.
- Cursor auf definierter Zeile + `n` → unverändert Blank-Create-Input
  (Registrieren ergibt hier keinen Sinn, Name existiert schon).
- Keine Zeilen (leere Liste) + `n` → unverändert Blank-Create-Input.
- Cursor-Verhalten nach Adopt: Zeile bleibt (jetzt `defined=true`,
  `refindName` = Zeilenname, mirrort Delete/Rename-Konvention).
- Test: `keyTagManagement`-Tabellentest um Free-Row-`n`-Case erweitert,
  Assertion auf `SaveTagDefs`-Aufruf-Inhalt + KEIN `tagMgmtInputActive`.

Konfliktrisiko: berührt DIESELBE Datei wie `bt-ct3k`
(view_tag_management.go) — `blocked_by bt-ct3k` gesetzt, in einer Session
NACH `bt-ct3k` bearbeiten.


## Summary

`keyTagManagement`s `keys.NewTag`-Case (`internal/tui/view_tag_management.go`) prüft jetzt die
Cursor-Zeile (mirrort `openTagMgmtRename`s Row-Check): gültige, FREIE Zeile → neue Funktion
`openTagMgmtAdopt(row)` dispatcht `saveTagDefsCmd(m.client, data.AddTagDefName(definedTagNames(
m.tagMgmtRows), row.name), row.name, "tag '<name>' registered")` DIREKT — kein Input-Submodus
(Planner-Entscheidung final). Sonst (keine Zeilen ODER Zeile bereits defined) → unverändertes
Blank-Create. `tagDefsSavedMsg` trägt neues Feld `successToast` (`internal/tui/messages.go`);
`applyTagDefsSaved` (`internal/tui/update.go`) zeigt bei nicht-leerem `successToast` einen
toastInfo-Erfolgs-Toast — Create/Rename/Delete passen `""` und bleiben unverändert still.
Cursor-Follow kommt gratis über den bestehenden refindName-Refind.

## Test-Output

RED (vor Fix): Compile-RED, `go vet`:
```
vet: internal/tui/view_tag_management_test.go:562:9: tdm.successToast undefined (type tagDefsSavedMsg has no field or method successToast)
```
(Die drei neuen Tests — `TestKeyTagManagementNewTagOnFreeRowAdoptsDirectlyNoInput`,
`TestApplyTagDefsSavedSuccessToastShowsInfoToastWhenSuccessToastSet`,
`TestFullAdoptFlowRegistersFreeRowShowsToastCursorFollows` plus
`TestKeyTagManagementNewTagOnDefinedRowStillOpensBlankCreate` — waren vor dem Fix nicht
kompilierbar/nicht erfüllbar.)

GREEN (nach Fix): alle 4 neuen Tests PASS; voller `go test ./...` ×2 (Run 2 mit `-count=1`,
internal/tui 139.7s) grün, `go vet` clean, `gofmt -l` leer.

## Smoke

Echter tmux-Smoke (`bin/bt` im Worktree, Registry `.beans-tags.yml` existierte nicht → alle
Tags frei): Tags-Page, Cursor auf freiem "to-review", `n` → Toast "tag 'to-review' registered",
Zeile bekommt ✓-Marker, KEIN Input-Submodus. `n` auf der jetzt definierten Zeile → unverändertes
Blank-Create-Input (esc). Cursor-Follow: `n` auf freiem "smoke" (Mitte der Liste) → "smoke"
wandert alpha-sortiert in die Defined-Gruppe, Cursor folgt (▌ auf "smoke"). Erzeugte
`.beans-tags.yml` vor Commit entfernt (Test-Mutation), keine Bean-Mutationen (Adopt ist
Registry-only).

## Deviations/ERRATA

- Zeilenreferenz Plan/bean `update.go:472` (applyTagDefsSaved) stimmte; `view_tag_management.go:373-374`
  (keys.NewTag-Case) real bei Z.373-379 nach bt-ct3k-Commit — funktional identisch, kein echtes Erratum.
- Signatur-Erweiterung statt neuem Msg-Typ: `saveTagDefsCmd` bekam einen 4. Parameter
  `successToast` (alle 3 Bestands-Callsites explizit `""`), der Erfolgs-Toast hängt wie geplant
  in `applyTagDefsSaved` am bestätigten Write — bean sah "Erfolgs-Toast in applyTagDefsSaved
  anhängen" vor, ließ den Transportweg offen; refindName-B01-Konvention (jede Dispatch-Site
  benennt explizit) wurde gespiegelt.


## Fix-Runde (Review-Finding, 2026-07-16)

Non-blocking Review-Finding umgesetzt: `openTagMgmtAdopt` validiert `row.name` jetzt defensiv
gegen `data.ValidTagName` VOR dem Registry-Write (Spiegelung des Create-Pfads) — ungültiger
Name (hand-editierte Bean-Datei kann Grammatik-verletzende Tags tragen) → toastWarn
"invalid tag name (a-z0-9, hyphen-separated, lowercase)", KEIN Registry-Write, bewusst KEIN
Fallback auf Blank-Create (n wurde AUF dieser Zeile gedrückt).

RED: `TestKeyTagManagementNewTagOnFreeRowInvalidNameWarnsNoSave` —
"want a toastWarn Toast for an invalid free name, got <nil>" → GREEN nach Gate.
Gates: `-short`-Lauf + voller Lauf (internal/tui 142.0s) grün, vet/gofmt clean.
