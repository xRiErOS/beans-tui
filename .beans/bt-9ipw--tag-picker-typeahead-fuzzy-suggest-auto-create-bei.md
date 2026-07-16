---
# bt-9ipw
title: Tag-Picker Typeahead — Fuzzy-Suggest + Auto-Create bei No-Match
status: completed
type: feature
priority: normal
created_at: 2026-07-16T20:13:14Z
updated_at: 2026-07-16T21:36:06Z
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


## Summary (Implementer, 2026-07-16)

Drei-Schritt-Zerlegung (Filter/Navigation/Enter-Dispatch+Rendering) in EINER Session
umgesetzt, ausschließlich `box_picker_tag.go` (+ `types.go` Feld-Deklarationen,
+ `box_picker_tag_test.go`). `view_tag_management.go` unangetastet.

- Neue Felder: `tagInputFiltered []tagCount`, `tagInputSuggestCursor int` (types.go).
- Neue Funktion `filterTagItems(items []tagCount, query string) []tagCount` —
  case-insensitive `strings.Contains`, kein Fuzzy (YAGNI). Leerer Query matcht alles
  (mirrort `filteredRepos()`s Kontrakt, view_lobby.go).
- `openTagInput` seedet `tagInputFiltered` sofort mit der vollen `tagItems`-Liste
  (Query "") statt einer leeren Vorschlagsliste beim Öffnen.
- `keyTagInput`: `tea.KeyUp`/`tea.KeyDown` (RAW KeyType, NICHT `navKey`s
  Buchstaben-Alias-Tabelle — `keys.Up`/`keys.Down` binden zusätzlich "i"/"k", das
  hätte im Freitext-Feld getippte "i"/"k" verschluckt) bewegen den Cursor, geklemmt
  auf `[0, len(tagInputFiltered)-1]`. `enter` verzweigt: nicht-leeres
  `tagInputFiltered` → existierenden Tag am Cursor übernehmen (Copy-on-Write wie
  `toggleTagPending`, KEIN Registry-Write, KEINE neue `tagItems`-Zeile); leeres
  `tagInputFiltered` → heutiger Neuanlage-Pfad UNVERÄNDERT (D11). Jede andere Taste
  geht an `m.tagInput.Update`, `tagInputFiltered`/`tagInputSuggestCursor` werden NUR
  neu berechnet, wenn sich der Input-Wert dadurch TATSÄCHLICH geändert hat (mirrort
  `keyLobby`s prev/current-Guard, view_lobby.go) — eine wertneutrale Taste (z. B.
  Cursor-Bewegung im Textfeld) reißt die gerade navigierte Vorschlags-Cursor-Position
  nicht zurück auf 0.
- `tagInputBox`: Hinweiszeile ist jetzt dynamisch ("up/down:navigate enter:select
  esc:cancel" bei Treffern, sonst unverändert "enter:create esc:cancel"), darunter
  die gefilterte Vorschlagsliste mit `▸`-Cursor + Marker-Spalte (mirrort
  `tagPickerBox`s Zeilenformat 1:1).

## Test-Output (RED→GREEN je Schritt)

RED (vor Implementierung, `go test ./internal/tui/... -run TestTagInput -v`):
```
=== RUN   TestTagInputOpensWithFilteredSeededFromFullTagList
    tagInputFiltered len = 0, want 3 (full tagItems on empty query)
--- FAIL
=== RUN   TestTagInputFiltersLiveBySubstringCaseInsensitive
    tagInputFiltered = [], want exactly [urgent]
--- FAIL
=== RUN   TestTagInputNavigationMovesCursorClampedToFilteredBounds
    cursor after two downs = 0, want 2
--- FAIL
=== RUN   TestTagInputEnterOnSuggestionAssignsExistingTagNoDuplicateNoRegistryWrite
    setup: tagInputFiltered = [], want exactly [backend]
--- FAIL
=== RUN   TestTagInputBoxRendersFilteredSuggestionsWithCursorMarker
    want all 3 suggestion rows rendered
--- FAIL
FAIL
```
(`TestTagInputArrowKeysDoNotLeakIntoTypedText` und
`TestTagInputEnterWithNoMatchStillFallsBackToUnchangedCreatePath` waren bereits vor
der Implementierung GREEN — reine Regressions-/Kollisions-Guards gegen bestehendes
Verhalten, kein neuer Code nötig für ihr Bestehen.)

GREEN (nach Implementierung, gleicher Testlauf, 8 neue + 20 bestehende Tests im File):
```
--- PASS: TestTagInputOpensWithFilteredSeededFromFullTagList (0.00s)
--- PASS: TestTagInputFiltersLiveBySubstringCaseInsensitive (0.00s)
--- PASS: TestTagInputNavigationMovesCursorClampedToFilteredBounds (0.00s)
--- PASS: TestTagInputArrowKeysDoNotLeakIntoTypedText (0.00s)
--- PASS: TestTagInputEnterOnSuggestionAssignsExistingTagNoDuplicateNoRegistryWrite (0.00s)
--- PASS: TestTagInputEnterWithNoMatchStillFallsBackToUnchangedCreatePath (0.00s)
--- PASS: TestTagInputBoxRendersFilteredSuggestionsWithCursorMarker (0.00s)
PASS
ok  	beans-tui/internal/tui	0.522s
```

Voller Lauf (`go test ./...`, KEIN `-short`) zweimal in Folge grün, keine Golden-Diffs:
```
ok  	beans-tui/cmd	0.687s
ok  	beans-tui/internal/config	1.202s
ok  	beans-tui/internal/data	3.089s
ok  	beans-tui/internal/theme	0.332s
ok  	beans-tui/internal/tui	139.105s
```
(Lauf 2: identisch, exit 0.) `gofmt -l` leer, `go vet ./...` sauber.

## Smoke (echt gesmoked via tmux, bin/bt, worktree feat-typeahead)

1. Bean `bt-6oyy` fokussiert, `t` → Tag-Picker öffnet mit `to-review (11)`,
   `rejected (1)`, `smoke (1)`.
2. `n` → New-Tag-Overlay öffnet mit ALLEN 3 Tags als Vorschläge (Cursor auf
   `to-review`), Hint "up/down:navigate enter:select esc:cancel".
3. Tippen `re` → Liste filtert sichtbar auf `to-review (11)` + `rejected (1)`
   (`smoke` verschwindet) — Case-insensitive Substring bestätigt.
4. `↓` → Cursor bewegt sich sichtbar auf `rejected`.
5. `enter` → Overlay schließt zurück in den Tag-Picker, `rejected` jetzt `[x]`
   angehakt, Zeilenzahl weiterhin 3 (KEIN Duplikat).
6. `n` erneut, Tippen `smoke-brandnew` (kein Treffer) → Vorschlagsliste leer,
   Hint fällt zurück auf "enter:create esc:cancel".
7. `enter` → Tag-Picker zeigt neue Zeile `smoke-brandnew (0)`, sofort `[x]`
   angehakt — heutiger Neuanlage-Pfad unverändert bestätigt.
8. `esc` verwirft alle Pending-Changes (kein Save), `q` → Quit-Confirm-Dialog
   erscheint korrekt.
9. `git status --short .beans/` vor UND nach dem Smoke: leer — keine
   dauerhaften Test-Mutationen.

Ein Zwischenfall während des ERSTEN Smoke-Versuchs: eine generische
tmux-Session namens `btsmoke` wurde von einem PARALLEL laufenden Agent (dasselbe
beans-tui-Projekt, vermutlich Default-`main-direkt`-Konvention gegen
`beans-tui-repository`) unter demselben Namen neu angelegt und hat meine
Folge-Keystrokes abgefangen (sichtbar am Repo-Titel-Wechsel auf
"beans-tui-repository" und späterem Prozessende). Kein Bug im eigenen Code — die
ERSTEN beiden Captures (Tag-Picker offen, New-Tag-Overlay mit korrekter
Vorschlagsliste) bewiesen die eigene Implementierung bereits vor dem Zwischenfall.
Wiederholt mit kollisionssicherem Sessionnamen (`bt9ipw$$`), vollständig grün, siehe
oben.

## Deviations/ERRATA

Keine Zeilenreferenz-Abweichungen vom Plan (`keyTagInput` Z.270-310,
`tagInputBox` Z.397-405 stimmten exakt). Eine bewusste Design-Entscheidung über
den Plan-Wortlaut hinaus: Cursor-Navigation nutzt den RAW `tea.KeyUp`/
`tea.KeyDown`-KeyType statt `navKey()` (das im übrigen Code Buchstaben-Aliase wie
"i"/"k" für hoch/runter bindet) — im Freitext-Erfassungsfeld müssen "i"/"k" als
literale, tippbare Zeichen erhalten bleiben (z. B. Tag-Name "risk"). Eigens per Test
abgesichert (`TestTagInputArrowKeysDoNotLeakIntoTypedText`). Dynamische Hint-Zeile
in `tagInputBox` ("enter:select" vs. "enter:create") war nicht explizit gefordert,
aber direkte Konsequenz aus dem verzweigten enter-Verhalten — ohne sie wäre der Hint
irreführend.


### Nachtrag Fix-Runde (Review-Findings, 2026-07-16)

Review APPROVED mit zwei non-blocking Findings, beide behoben (Commit 9a42af6):

1. **Test-Lücke (medium):** Die Mutation `tagInputFiltered[tagInputSuggestCursor]`
   → `[0]` überlebte die Suite (einziger Enter-auf-Vorschlag-Test hatte genau 1
   Treffer, Cursor trivial 0). Neuer Test
   `TestTagInputEnterSelectsCursoredSuggestionNotFirst`: 3 Treffer, Cursor per Down
   auf Index 1, enter MUSS den gecursorten Tag wählen. RED-Beweis gegen die
   temporär eingespielte `[0]`-Mutation:
   `enter must assign the CURSORED suggestion "backend", tagPending = map[urgent:true]`
   → FAIL. Mutation zurückgenommen → PASS.
2. **Doc-Kommentar (low):** keyLobby-Begründung in `types.go` +
   `box_picker_tag.go` korrigiert — keyLobby routet navKey() VOR dem
   Textinput-Update, das i/k-Verschlucken dort ist ein BESTEHENDER Bug
   (bean bt-l8e7 im Haupt-Repo), keine bewusste Abwägung; der raw-KeyType-
   Intercept hier ist das korrekte Muster.

Gates: voller Lauf ohne `-short` grün (`ok beans-tui/internal/tui 139.135s`),
gofmt leer, go vet sauber, working tree clean.
