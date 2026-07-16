---
# bt-9ipw
title: Tag-Picker Typeahead — Fuzzy-Suggest + Auto-Create bei No-Match
status: todo
type: feature
priority: normal
created_at: 2026-07-16T20:13:14Z
updated_at: 2026-07-16T20:47:11Z
parent: bt-362n
---

Nebenbefund aus US-Review 2026-07-16 (Runde 2, US-03 New-Tag-Overlay), PO-Zitat:

"Wenn die tag-liste lang wird (lange projekte, große projekte), dann muss ich die tags
suchen können. Für jetzt ist die Liste so ok. Aber ich denke, dass ein change besser ist,
dass ich die tags tippe, mir Vorschläge zu existierenden angezeigt werden und ich diese
mit Enter übernehme, mit Pfeiltasten (rauf/runter) navigiere und mit Enter übernehme. Ein
getippter tag, der keine Übereinstimmung hat, wird automatisch neu angelegt."

## Akzeptanzkriterien (Entwurf, Planner verfeinert vor Umsetzung)

- Tag-Picker-Input (`n`-Modus, alle drei Einstiege: `t`-Picker/Feld-`enter`/Palette
  „create tag") wird zum Typeahead: Tippen filtert die bestehende Tag-Liste
  (definiert + frei, `collectTagCounts`) live nach Substring/Fuzzy-Match.
- Pfeiltasten rauf/runter navigieren durch die gefilterten Vorschläge.
- `enter` auf einem Vorschlag übernimmt den EXISTIERENDEN Tag (kein Duplikat, kein
  Registry-Eintrag nötig falls schon definiert).
- `enter` auf getipptem Text OHNE Treffer legt den Tag NEU an (heutiges `n`-Verhalten,
  D11: reiner Registry-Akt, kein Bean automatisch betroffen — nur Zuweisung ans
  aktuell editierte Bean bleibt wie gehabt).
- Betrifft NUR den Tag-Zuweisungs-Picker am Bean (E8-B14/E10-Suggest-Mode), NICHT die
  Tag-Management-Page (`bt-362n` §16) — dort bleibt die Liste unverändert (kleine,
  überschaubare Menge, kein Such-Bedarf laut PO).

## Nicht jetzt (YAGNI bis PO bestätigt)

- Fuzzy-Scoring-Algorithmus (Bleve o.ä.) — einfacher Substring-Filter reicht vermutlich,
  Planner entscheidet bei Umsetzung.
- Kein Scope für Tag-Management-Page-Suche (siehe oben).

Quelle: PO-Chat 2026-07-16, US-Review-Session (ce-start), im Kontext von US-03 (bt-tct9).



## Planner-Konkretisierung (2026-07-16)

**Scope-Bestätigung:** betrifft AUSSCHLIESSLICH `box_picker_tag.go`
(Tag-Zuweisungs-Picker am Bean, `t`/Feld-`enter`/Palette „create tag") —
NICHT `view_tag_management.go` (Tag-Management-Page bleibt exakt wie in
E10 gebaut, keine Suche dort, PO-bestätigt). Betroffene Bestandsfunktionen:
`keyTagPicker` (Zeile 203-229, aktuell: `up`/`down` bewegen `m.menu` über
ALLE `m.tagItems`, `n`/`keys.NewTag` öffnet den GETRENNTEN Blank-Input-
Submodus `keyTagInput`/`m.tagInputActive`), `openTagInput`/`keyTagInput`
(Zeile 254-310, reiner Freitext, KEINE Filterung/Anzeige der bestehenden
Liste während der Eingabe), `tagPickerBox`/`tagInputBox` (Zeile 365-405,
zwei GETRENNTE Render-Pfade).

**Empfohlene Aufteilung in 3 Teil-Tasks (EINE Datei, sequentiell in EINER
Session, um Merge-Konflikte zu vermeiden):**

1. **Filter-Logik:** `keyTagInput` bekommt bei JEDEM Tastendruck (außer
   enter/esc, die bleiben) eine Live-Filterung von `m.tagItems` nach
   `strings.Contains(strings.ToLower(tag), strings.ToLower(query))`
   (Substring, KEIN Fuzzy-Scoring — YAGNI laut Bean-Entwurf) — neues
   Modelfeld `tagInputFiltered []tagCount` (oder Index-Slice), neu
   berechnet nach jedem `m.tagInput.Update(msg)`-Aufruf (Zeile 308).
   Reuse: `sortTagCountsDefinedFirst` (Zeile 115) bleibt die Sortierung
   INNERHALB des gefilterten Ergebnisses.
2. **Navigation:** neues Cursor-Feld (z. B. `tagInputSuggestCursor int`,
   types.go, analog `m.menu`), `up`/`down` innerhalb `keyTagInput` (heute
   unbenutzt/an `textinput.Update` durchgereicht, da `textinput.Model`
   einzeilig ist — keine Kollision) bewegen diesen Cursor über
   `tagInputFiltered`, geklemmt auf `[0, len(tagInputFiltered)-1]`.
3. **Enter-Dispatch + Rendering:** `keyTagInput`s `enter`-Case (Zeile
   277-304) verzweigt NEU: `len(tagInputFiltered) > 0` → EXISTIERENDEN Tag
   an `tagInputSuggestCursor` übernehmen (`m.tagPending[tag]=true`,
   mirrort `toggleTagPending`s Copy-on-Write, Zeile 237-250 — KEIN
   `data.AddTagDefName`, KEIN Registry-Write, nur Zuweisung ans aktuelle
   Bean, wie im Bean-Entwurf gefordert), Submodus schließt.
   `len(tagInputFiltered) == 0` (kein Treffer) → heutiger
   Neuanlage-Pfad (Zeile 279-304) UNVERÄNDERT. `tagInputBox` (Zeile
   397-405) rendert zusätzlich die gefilterten Vorschlagszeilen MIT
   Cursor-Marker (mirrort `tagPickerBox`s `▸`-Konvention, Zeile 383-386)
   unterhalb des Eingabefelds.

**Akzeptanzkriterium:**
- Tippen im `n`-Submodus filtert sichtbar (definierte + freie Tags,
  `collectTagCounts`-Union).
- Pfeiltasten navigieren durch gefilterte Vorschläge, sichtbarer Cursor.
- `enter` auf einem Vorschlag → existierender Tag übernommen, KEIN
  Duplikat, KEIN Registry-Write falls bereits definiert.
- `enter` ohne Treffer → Neuanlage wie heute (D11 unverändert: reiner
  Zuweisungs-Akt am aktuellen Bean, Registry unangetastet für undefinierte
  neue Tags).
- Tag-Management-Page (`view_tag_management.go`) UNVERÄNDERT (kein
  Cross-Scope-Leck).
- Golden-Gegenbeleg (Overlay ist nicht Teil der 3 Basis-Goldens, aber
  Test-Suite muss grün bleiben).

Kein Datei-Konflikt mit bt-ct3k/bt-idm1 (andere Datei) — unabhängig
bearbeitbar, aus Kontext-Kontinuität aber sinnvoll NACH der
Tag-Management-Nacharbeit eingeplant (dieselbe Feature-Familie, geteilte
Konventionen wie `collectTagCounts`/`tagManagementMarkerGlyph`).
