---
# bt-idm1
title: 'Tag-Management: n auf freier Zeile registriert bestehenden Tag (Adopt statt Blank-Create)'
status: todo
type: feature
priority: normal
created_at: 2026-07-16T20:35:29Z
updated_at: 2026-07-16T20:47:17Z
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
