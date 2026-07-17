---
# bt-9ipw
title: Tag-Picker Typeahead — Fuzzy-Suggest + Auto-Create bei No-Match
status: completed
type: feature
priority: normal
created_at: 2026-07-16T20:13:14Z
updated_at: 2026-07-17T07:34:43Z
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


## Review 2026-07-17 — US-07 REJECTED

PO verbatim: "Ich sehe, dass der Selektor wandert, wenn ich Tippe. Aber ich
habe kein visuellen Hinweis darauf, WAS ich getippt habe. Daher sollte das
tags-overlay ein Suchfeld haben, in welchem ich sehe, was ich getippt habe.
Ferner soll die angezeigte Liste durch meine Suche gefiltert werden, damit ich
nur passende Elemente angezeigt bekomme."

Fix-Prelude:
- Schwere: medium (Feature unbenutzbar ohne Sichtbarkeit der Eingabe, Kern-UX)
- Fundort: internal/tui/box_picker_tag.go — tagInputBox/tagPickerBox; klären,
  in WELCHEM Sub-Modus der PO getippt hat: Typeahead lebt im n-Input-Submodus
  (tagInputBox rendert textinput + Vorschläge), aber der Befund ("Selektor
  wandert beim Tippen, kein Eingabe-Echo") klingt nach dem Haupt-Picker (t),
  wo Tippen evtl. als Navigation/Filter wirkt ohne sichtbares Suchfeld.
  Repro-first: beide Einstiege (t-Picker direkt tippen vs. t→n) im tmux prüfen.
- Fix-Rezept (Soll): Tag-Overlay bekommt ein IMMER sichtbares Suchfeld (Echo
  der Eingabe) direkt im Picker (t) — Tippen filtert die angezeigte Liste
  live; Pfeiltasten navigieren die gefilterte Liste; enter übernimmt; kein
  Treffer → Neuanlage-Pfad. Der n-Submodus-Typeahead (df249d7) liefert die
  Filter-/Cursor-Bausteine — ggf. Konsolidierung beider Modi in einen.
- Quelle: PO-Review E11 Runde 3, 2026-07-17.


## Plan-Konkretisierung E12 (2026-07-17)

Plan: `docs/plans/v1-port/epic-E12-plan.md` §„Item 1: Tag-Picker (`t`)
braucht sichtbares Suchfeld". Reihenfolge: Rang 1 (zusammen mit `bt-81f0`
PO-mandiert zuerst).

**Root Cause:** `box_picker_tag.go` hat zwei getrennte Overlay-Zustände —
der Haupt-Picker (`tagPickerBox`/`keyTagPicker`, Zeile 203-229/444-473) hat
KEIN Textinput, Tippen tut dort nichts. Das sichtbare Suchfeld +
Live-Filterung existiert nur im separaten `n`-Submodus
(`tagInputActive`/`openTagInput`/`keyTagInput`/`tagInputBox`, Zeile
254-389/475-511, aus der ERSTEN `bt-9ipw`-Runde). PO-Zitat beschreibt exakt
diese Lücke am Haupt-Picker.

**Vorgehen (D01 im Plan):** Repro-first PFLICHT (beide Einstiege `t` direkt
vs. `t`→`n` im tmux prüfen). Danach EIN konsolidierter Modus: Haupt-Picker
bekommt ein IMMER sichtbares, fokussiertes Suchfeld (dasselbe
`m.tagInput`-Widget + `filterTagItems`), Tippen filtert live, Pfeiltasten
navigieren die gefilterte Liste, space/x togglet Multi-Select UNVERÄNDERT
(zentraler Unterschied zum alten `n`-Submodus, dessen `enter` sofort
schloss). `n`/kein Treffer bleibt Eskalationspfad zur Neuanlage (D11
unverändert). Der zweistufige `tagInputActive`-Einstieg entfällt.

**Akzeptanz:**
- [ ] Repro-Ergebnis dokumentiert (welcher Einstieg zeigte das Problem)
- [ ] `t` öffnet Picker MIT sofort sichtbarem, fokussiertem Suchfeld
- [ ] Tippen filtert sichtbar, Eingabetext ist im Feld zu sehen
- [ ] Pfeiltasten navigieren gefilterte Liste, space/x toggelt Multi-Select
      unverändert (mehrere Tags gleichzeitig anhakbar)
- [ ] `enter` ohne Treffer legt neu an (D11 unverändert)
- [ ] `view_tag_management.go` UNANGETASTET
- [ ] Test-Suite grün, Golden-Gegenbeleg falls Overlay-Breite sich ändert
- [ ] tmux-Smoke: Tippen zeigt Text im Feld, Filterung sichtbar

## Summary (Implementer, 2026-07-17, US-07-Reopen Fix)

Konsolidiert `box_picker_tag.go`s zwei getrennte Overlay-Zustände (Haupt-Picker
ohne Textinput + separater `n`-Submodus mit dem einzigen sichtbaren Suchfeld)
zu EINEM Modus (D01, epic-E12-plan.md »Item 1«): `t` öffnet den Picker jetzt
mit einem sofort sichtbaren, fokussierten Suchfeld (`m.tagInput`). Tippen
filtert `m.tagInputFiltered` live (bestehende `filterTagItems`, case-
insensitive Substring, YAGNI-Fuzzy weiterhin nicht). Pfeiltasten (raw
`tea.KeyUp`/`tea.KeyDown`, NICHT `navKey`) navigieren die gefilterte Liste;
space/x (`keybind.Matches(msg, keys.Toggle)`) togglet weiterhin Multi-Select
am Cursor — UNVERÄNDERT möglich während das Feld fokussiert ist (zentrale
Akzeptanz dieses Reopens). `enter` ist jetzt kontextabhängig: mit mindestens
einem Treffer (inkl. leerer Query = volle Liste) unverändertes
`applyTagPickerDiff` (Save+Close); ohne Treffer bei nicht-leerem Query ->
Neuanlage-Pfad (D11 unverändert: reine Zuweisung, kein Registry-Write), der
Query danach zurücksetzt statt den Picker zu schließen ("weiter geht's").

Der zweistufige `t`→`n`-Einstieg (`tagInputActive`/`openTagInput`/
`keyTagInput`/`tagInputBox`) ist komplett entfernt — vollständig in
`tagPickerBox`/`keyTagPicker` aufgegangen (Implementer-Entscheidung aus dem
Plan-Text: kein Compat-Layer). `openTagPicker()` liefert jetzt
`(tea.Model, tea.Cmd)` statt nur `model` (mirrort `openSearchInput`) und
gibt `textinput.Blink` zurück, damit der jetzt immer fokussierte Cursor
tatsächlich blinkt — betroffene Call-Sites angepasst: `update.go` (zwei
Stellen: `keyNodeAction`s `t`-Case, `activateDetailField`s `"tags"`-Case)
und `overlay_palette.go` (`"tags"`-Case). Die Palette-Aktion `"create_tag"`
war früher ein SEPARATER Chain-Aufruf (`openTagPicker().openTagInput()`
+ eigener `focusedBean()`-Guard) — post-D01 ist sie inhaltlich identisch mit
`"tags"` (beide landen im selben, sofort tippbaren Picker), daher auf
`return m.openTagPicker()` vereinfacht.

`footer_context.go`s `tagPickerLocalBindings()` verliert `keys.NewTag`
("n" ist jetzt ein normaler tippbarer Buchstabe im Suchfeld, kein
Picker-Kommando mehr) — `keys.NewTag` selbst bleibt im globalen Keymap
bestehen, da `view_tag_management.go`s eigene "n"-Bindung (Tag-Management-
Page, UNANGETASTET laut Auftrag) ihn weiterhin nutzt.

## Test-Output (RED -> GREEN)

RED (Compile-Fail, `go vet ./...` nach der Implementierung, VOR der
Test-Anpassung):

```
# beans-tui/internal/tui
vet: internal/tui/box_picker_tag_test.go:365:8: m.tagInputActive undefined (type model has no field or method tagInputActive)
```

(Der Produktivcode kompilierte bereits fehlerfrei — Repro-first + Design galt
also erst als bewiesen, nachdem die Tests nachgezogen waren. Nach dem
Test-Rewrite: RED-Fehler behoben, `go vet ./...` sauber.)

Zielgerichteter Testlauf nach dem vollständigen Rewrite (Tag-Picker +
Footer + Palette + Keymap):

```
go test ./internal/tui/... -run 'TagPicker|TagInput|Palette|FooterContext|Keymap|NewTag|Contextual' -v
...
ok  	beans-tui/internal/tui	4.149s
```

Alle 8 alten Typeahead-spezifischen Tests (`TestTagInputOpensWithFiltered...`,
`TestTagInputFiltersLive...`, `TestTagInputNavigation...`,
`TestTagInputArrowKeysDoNotLeak...`, `TestTagInputEnterOnSuggestion...`,
`TestTagInputEnterSelectsCursored...`, `TestTagInputEnterWithNoMatch...`,
`TestTagInputBoxRenders...`) sowie `TestTagPickerEscFromInputKeepsPickerOpen...`
und die beiden `create_tag`-Palette-Tests wurden umbenannt/neu geschrieben, da
ihre Prämisse (getrennter `n`-Submodus) durch D01 wegfällt — Details siehe
Deviations unten.

Voller Lauf (`go test ./...`, KEIN `-short`) ZWEIMAL in Folge grün (zweiter
Lauf mit `-count=1` erzwungen, nicht aus dem Cache):

```
Lauf 1: ok  	beans-tui/internal/tui	142.378s   (+ cmd/config/data/theme grün)
Lauf 2: ok  	beans-tui/internal/tui	143.159s   (-count=1, kein Cache-Treffer)
```

`gofmt -l .` leer, `go vet ./...` sauber, `command go build -o bin/bt .`
sauber.

## Smoke (echt gesmoked via tmux, bin/bt, main-direkt-Repo)

Repro-first (Schritt 2, VOR jeder Code-Änderung, gegen den alten Stand):
tmux-Session `bt9ipwR2$$`/`btsmoke$$` (Namenskollision mit einer generischen
`bt9ipw$$`-Session eines anderen parallelen Agents beobachtet — sofort mit
eindeutigerem Namen neu aufgesetzt, mirrort E11-Lessons-Learned Eintrag 1).
Ergebnis: `t` (Haupt-Picker) + Tippen "urg" -> KEIN sichtbares Feedback
(kein Feld, keine Filterung) — bestätigt den PO-Befund exakt. `t`→`n`
(alter Submodus) + Tippen "urg" -> Feld UND Text sichtbar ("urg" im
`New tag`-Panel). Root Cause damit am Haupt-Picker (nicht am Submodus)
bestätigt, wie im Plan diagnostiziert.

Nach der Implementierung, neuer Build, echtes `bin/bt` gegen dieses Repo
(Bean `bt-apmy` als Test-Target, tags vorher `{to-review, smoke}`):

1. `t` -> Picker öffnet MIT sofort sichtbarem, fokussiertem Suchfeld
   (Platzhaltertext "type to filter or create...").
2. Tippen "rev" -> Feld zeigt "rev", Liste filtert live auf "to-review"
   (rejected/accepted/smoke verschwinden) — PO-Wortlaut exakt widerlegt.
3. Backspace ×3, Tippen "e" -> alle 4 Zeilen wieder sichtbar (alle
   enthalten "e") — Live-Filter bestätigt in beide Richtungen.
4. Backspace, Pfeil runter ×2 -> Cursor sichtbar auf "accepted" (volle,
   ungefilterte Liste nach Leerung).
5. space auf "accepted" -> `[x]`; Pfeil hoch, "x" auf "rejected" -> `[x]`
   — Multi-Select über BEIDE Toggle-Tasten (space UND x) bestätigt,
   während das Suchfeld fokussiert bleibt (kein Leck in den Feldtext).
6. Tippen "zzz-brandnew" (kein Treffer) -> Hint wechselt auf
   "no match — enter:create new tag", Zeile "(no match)".
7. enter -> "zzz-brandnew (0)" neu angelegt, sofort `[x]`, Query geleert,
   Picker BLEIBT offen (D01: "weiter geht's", kein Auto-Close mehr).
8. enter (jetzt leere Query = Treffer-Zustand) -> Picker schließt, Detail-
   Panel zeigt `tags: to-review, smoke, accepted, rejected, zzz-brandnew`
   — Save+Close-Pfad bestätigt, echte `beans update`-Mutation griff.
9. Rückgängig gemacht (Picker erneut geöffnet, die drei Zusätze
   abgewählt, enter) -> `git diff` zeigt wieder exakt `{to-review, smoke}`
   (nur `updated_at` geändert -> per `git checkout --` zurückgesetzt,
   `git status --short .beans/` danach leer).
10. esc-Test: toggle + esc -> Picker schließt OHNE Mutation, Tags
    unverändert (`git status --short .beans/` weiterhin leer).
11. `q` im Picker -> KEIN Quit, `q` landet als Text im Suchfeld (bestätigt
    Capture-Order + die neue "q ist jetzt Filtertext"-Semantik).

Real gesmoked: Suchfeld-Sichtbarkeit/Echo, Live-Filter (Tippen+Backspace),
Multi-Select (space+x) während Filterung, Neuanlage-Pfad, Save+Close-Pfad,
esc-Discard, Capture-Order. NICHT separat im tmux gesmoked (nur Unit-Level):
`tagPickerBox`s exakte Spalten-/Marker-Rendering-Details (PF-12,
`TestTagPickerBoxReservesMarkerColumnRegardlessOfDefined` etc.) und die
Vanished-Target-/nil-Client-Randfälle — beide rein datengetrieben, tmux hätte
keinen zusätzlichen Beweiswert geliefert.

## Commits

- `7eb736c` — `fix(tui): consolidate Tag-Picker into one always-searchable mode` (Refs: bt-9ipw)

## Deviations/ERRATA

- **Enter-Semantik bei Treffer nicht explizit im Bean-/Plan-Text benannt:**
  Plan/D01 sagt nur "enter ohne Treffer -> Neuanlage-Pfad", lässt aber offen,
  was `enter` bei EINEM Treffer (oder leerer Query) tun soll. Implementer-
  Entscheidung (konsistent mit "`space/x` togglet Multi-Select unverändert
  (zentraler Unterschied zum alten n-Submodus, dessen enter sofort schloss)"):
  `enter` behält bei jedem Nicht-Leer-Zustand von `tagInputFiltered` seine
  ALTE, prä-Typeahead-Bedeutung (Save+Close via `applyTagPickerDiff`) —
  Tag-AUSWAHL läuft ausschließlich über space/x, nicht mehr über enter. Das
  ist die einzige Lesart, die den zitierten Kontrastsatz überhaupt erfüllbar
  macht (sonst gäbe es keinen Unterschied zum alten Submodus-Verhalten).
- **Space/x werden vor dem Suchfeld abgefangen (space und "x" sind damit NICHT
  als Zeichen in den Suchtext tippbar):** notwendige Konsequenz aus der
  Akzeptanz "space/x togglet Multi-Select unverändert" bei gleichzeitig
  immer-fokussiertem Feld — beide Tasten können nicht gleichzeitig "Toggle"
  UND "Text" bedeuten. `data.ValidTagName` verbietet Leerzeichen ohnehin
  (kein Verlust dort); "x" als Tag-Namensbestandteil (z. B. "nginx", "box")
  ist dadurch beim Tippen NICHT eingebbar — dieselbe bewusste Abwägung wie
  Up/Down vs. i/k, nur auf die explizit geforderten Toggle-Tasten übertragen.
  Nicht eigenständig als Frage eskaliert, da es die direkte, einzig
  konsistente Umsetzung der bindenden Akzeptanzkriterien ist (keine
  Spec-Änderung, sondern deren Anwendung).
- **`openTagPicker()`-Signaturänderung** (`model` -> `(tea.Model, tea.Cmd)`,
  liefert `textinput.Blink`): nicht explizit im Plan gefordert, aber
  notwendig, damit der jetzt immer fokussierte Cursor tatsächlich blinkt
  (mirrort `openSearchInput`s Konvention) — drei Call-Sites entsprechend
  angepasst (siehe Summary). Rein additiv, kein bestehender Aufrufer verlor
  Funktionalität.
- **Palette-Aktion `"create_tag"` vereinfacht auf `"tags"`-Äquivalent:** war
  vor D01 ein eigener Chain-Aufruf mit eigenem Guard; nach der Konsolidierung
  inhaltlich identisch — Implementer-Entscheidung, kein Scope-Creep (rein
  Compile-Notwendigkeit durch das Entfernen von `openTagInput()`).
- **Test-Umbenennungen/-Ersetzungen statt reiner Löschung:** die 8
  Typeahead-Submodus-Tests + 2 `create_tag`-Palette-Tests wurden NICHT
  ersatzlos gestrichen, sondern auf die neue Semantik ummodelliert (space/x
  statt enter für Auswahl, kein `n`-Gate mehr) — Testabdeckung bleibt
  äquivalent oder höher (zusätzlich: `TestTagPickerEnterWithFilteredMatchSavesAndClosesPicker`
  als expliziter neuer Test für den "enter bei Treffer"-Pfad, der vorher nur
  implizit über den leeren-Query-Fall abgedeckt war).
- **`TestOverlayCaptureSwallowsQuitKeysWhileTagPickerOpen`** musste inhaltlich
  angepasst werden (nicht nur mechanisch): "q" liefert jetzt legitim einen
  nicht-nil Cmd (textinput-eigener Blink-Restart), das ist kein Quit-Leck.
  Test prüft jetzt explizit `cmd().(tea.QuitMsg)` statt pauschal `cmd != nil`.

## Notes for bt-81f0

Nächster Task laut `epic-E12-plan.md` (Item 2, PO-mandiert Rang 2, direkt
nach diesem hier): Notifications vereinheitlichen (Toast als einziger Kanal).
Berührt laut Plan U. a. `box_picker_tag.go:425`
(`applyTagPickerDiff`s `m.err = "Bean no longer exists — selection discarded"`)
— das ist jetzt Zeile **409** in der konsolidierten Datei (Zeilennummern haben
sich durch diesen Task verschoben, ERRATUM-Hinweis für den nächsten
Implementer: nicht blind auf alte Zeilenangaben verlassen, `applyTagPickerDiff`
selbst ist inhaltlich UNVERÄNDERT). Ansonsten keine Überschneidung: `bt-81f0`
arbeitet an einer anderen Funktion (`m.err`-Rendering-Anbindung entfernen,
`showToast` ergänzen) in derselben Datei-Nachbarschaft — kein Zeilen-Konflikt
erwartet, aber dieselbe Datei, also nacheinander statt parallel bearbeiten
(wie im Plan bereits vorgesehen).

## Fix-Runde R1 (2026-07-17, Review CHANGES_REQUIRED)

### B01 (high, Pflicht) — "x" nie tippbar im Suchfeld: BEHOBEN

Root Cause bestätigt: `keybind.Matches(msg, keys.Toggle)` in `keyTagPicker`
fing "x" vor `m.tagInput.Update(msg)` ab (`keys.Toggle` bindet " " UND "x").

**ERRATUM / D01-Nachtrag (Supervisor-Entscheid):** Toggle im Tag-Picker ist
bewusst auf das SPACE-Zeichen verengt — die Plan-Formulierung "space/x
togglet Multi-Select unverändert" gilt für dieses Overlay nur noch für
space. Begründung: Tippbarkeit des häufigen Buchstabens "x" (nginx/linux/
unix/box/proxy) schlägt Alias-Redundanz; space ist gefahrlos reservierbar,
da `data.ValidTagName` nie ein Leerzeichen zulässt — analog zur i/k-RAW-
KeyType-Lösung. Filter-Menü und Blocking-Picker behalten space/x
unverändert.

Umsetzung: neues Keymap-Binding `keys.TagToggle` (space-only, `keymap.go`
bleibt Single Source, mirrort das dokumentierte RenameTag-Präzedenz-Muster
"separates Binding statt verengter Reuse", Drift-Guard
`TestHelpGroupsCoverEveryBindingExactlyOnce` deckt es ab — in helpGroups
Actions ergänzt). `keyTagPicker` dispatcht `keybind.Matches(msg,
keys.TagToggle)`; "x" fällt zum Textinput durch. Inline-Hint
(`tagPickerBox`) jetzt "space:toggle"; Footer Zone 3
(`tagPickerLocalBindings`) zeigt `keys.TagToggle` ("space Toggle tag")
statt des irreführenden "space/x Toggle facet"-Labels.

Regressionstest `TestTagPickerTypedXStaysLiteralNotToggle` — RED-Beweis
gegen den unfixierten Stand:

    === RUN   TestTagPickerTypedXStaysLiteralNotToggle
        box_picker_tag_test.go:614: tagInput.Value() = "linu", want "linux"
        -- 'x' must stay a literal, typeable character (B01)
    --- FAIL

Nach Fix: PASS (Value == "linux", tagPending unverändert — keine
Toggle-Nebenwirkung).

### I01 (medium, Pflicht) — Render-Ebene ungeschützt: BEHOBEN

Neuer Test `TestTagPickerBoxRendersSearchField`: ANSI-gestrippter
`tagPickerBox()`-Output MUSS das Suchfeld enthalten (Platzhalter "type to
filter or create" bei leerem Query, getippter Wert "rev" nach Eingabe).

Mutations-Selbstbeweis durchgeführt: `b.WriteString(m.tagInput.View() +
"\n")` lokal entfernt →

    --- FAIL: TestTagPickerBoxRendersSearchField
        tagPickerBox() = "...│ Tags │...│ ▸   [x] urgent (2) │..." ,
        want the search field's placeholder rendered while the query is empty

→ Zeile byte-identisch wiederhergestellt (git diff des Files danach leer
gegen den Zwischenstand), Test PASS.

### I02 (low, optional) — Palette-Doppeleintrag "set tags"/"create tag": WON'T-FIX (bewusst)

Beide Einträge dispatchen seit der D01-Konsolidierung identisch auf
`m.openTagPicker()`. Bewusst NICHT zusammengeführt: "create tag" ist ein
PO-sichtbarer Palette-Eintrag aus B14 (bean bt-yqdy, PO-abgenommen in E8) —
ihn zu entfernen wäre eine Änderung der abgenommenen Command-Center-
Oberfläche ohne PO-Freigabe und über den Reopen-Auftrag hinaus. Der
Eintrag bleibt als benanntes Alias nützlich (Nutzer, der "create" tippt,
findet ihn; "set tags" matcht auf diese Fuzzy-Query nicht). Follow-up-
Kandidat fürs nächste PO-Review: entweder PO segnet die Zusammenführung ab
oder der Alias bleibt dauerhaft. Kein stiller Zustand mehr — hiermit
dokumentiert (Doc-Kommentare in overlay_palette.go benennen die Identität
der beiden Dispatches bereits explizit).

### Gates R1

Voller Lauf `go test ./... -count=1` (ohne -short) grün, `go vet` sauber,
`gofmt -l` leer, `command go build -o bin/bt .` sauber. tmux-Pflicht-Check
(Session `bt9ipwR1fix$$`, echtes bin/bt): "linux" vollständig tippbar und
im Feld sichtbar (vorher: "linu" + ungewolltes Toggle), space togglet
weiterhin sichtbar am Cursor ([x]-Wechsel), space landet NICHT als Zeichen
im Feld, Hint zeigt "space:toggle". esc-Discard, keine bean-Mutationen
(`git status --short .beans/` leer).
