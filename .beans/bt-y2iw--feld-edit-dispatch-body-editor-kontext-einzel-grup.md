---
# bt-y2iw
title: 'Feld-Edit-Dispatch: BODY-Editor-Kontext + Einzel-Gruppen-Value-Menues'
status: completed
type: task
priority: high
created_at: 2026-07-15T21:08:25Z
updated_at: 2026-07-15T22:43:24Z
parent: bt-ntoz
blocked_by:
    - bt-qbyq
---

E8 Task 6 — deckt B10 (BODY-Sektion e/enter kontextsensitiv), B11+B12 (Value-Menu auf EINE Gruppe statt aller drei splitten + Palette-Ergaenzung) aus bean bt-ntoz. Quelle: design-spec.md §15 PF-16. Ist-Code: internal/tui/update.go (keyNodeAction, keyDetailFocus), internal/tui/box_menu_value.go, internal/tui/overlay_palette.go. blocked_by bt-qbyq (T3) -- B10 editiert keyDetailFocus im selben Bereich wie T3s esc-Handler, Reihenfolge vermeidet Merge-Ueberraschungen.

## B10 — BODY-Sektion: e/enter kontextsensitiv

HEUTE: in Detail-Fokus auf Sektion [2] BODY oeffnet die `e`-Taste (keyNodeAction, update.go, Editor-Case) IMMER das Titel-Edit-Form (unabhaengig vom Sektions-Kontext); `enter` auf Sektions-Ebene ist ein No-Op (BODY hat keine .fields, der bestehende `len(secs[...].fields) > 0`-Guard in keyDetailFocus greift nicht). PO: inkonsistent zu [1] META, wo enter das Overlay bringt.

FIX: neuer geteilter Helfer (vermeidet Duplikation zwischen den zwei Call-Sites):
```go
// openBodyEditor suspendiert $EDITOR auf b.Body -- geteilter Helfer fuer
// keyNodeAction's ctrl+e/e-auf-BODY-Sektion UND keyDetailFocus's
// enter-auf-BODY-Sektion (B10, design-spec.md §15 PF-16).
func (m model) openBodyEditor(b *data.Bean) (model, tea.Cmd) {
    m.editorTarget = b.ID
    m.editorETag = b.ETag
    return m, editInEditor(b.Body, ".md")
}
```
Nutzt bodySectionIdx (Konstante aus T1/bt-e6q9, view_detail_bean.go -- bereits gelandet, T3/bt-qbyq blockte bereits danach, diese Task erbt sie).

keyNodeAction()s Editor-Case (update.go): `if msg.String() == "ctrl+e" || (m.detailFocus && m.secCursor == bodySectionIdx) { nm, cmd := m.openBodyEditor(b); return true, nm, cmd }` -- ersetzt die bisherige `if msg.String() == "ctrl+e" { ... }`-Bedingung um die zweite Klausel, Body identisch (openBodyEditor kapselt sie). Alles andere (Titel-Edit-Fallback) bleibt UNVERAENDERT -- "e im Sektions-Kontext generell kontextsensitiv" heisst konkret: BODY->Editor, alle anderen Sektionen (META/RELATIONS/HISTORY)->Titel-Edit wie bisher (kein weiterer Fall gefordert, PO nennt nur BODY als Problem).

keyDetailFocus()s detailLevel==0-Enter-Block (update.go): NEUER Fall VOR dem bestehenden fields>0-Guard:
```go
if keybind.Matches(msg, keys.Enter) && m.detailLevel == 0 {
    if m.secCursor == bodySectionIdx {
        nm, cmd := m.openBodyEditor(b)
        return nm, cmd
    }
    if len(secs[m.secCursor].fields) > 0 {
        m.detailLevel = 1
        m.fieldCursor = 0
    }
    return m, nil
}
```

## B11 + B12 — Value-Menu auf EINE Gruppe splitten

HEUTE (box_menu_value.go): buildValueMenuItems() liefert IMMER alle 15 Zeilen (5 Status + 5 Type + 5 Priority) -- openValueMenu(group) nutzt `group` NUR zum Cursor-Seeding (valueMenuCursorFor), NICHT zum Filtern. Sowohl die `s`-Taste (immer group="status") als auch die PF-5-Enter-Kaskade (group je nach Feld: "status"/"type"/"priority") zeigen deshalb IMMER alle drei Gruppen -- fuer Type/Priority (via Kaskade erreicht) besonders verwirrend, da der Nutzer gezielt EIN Feld editieren wollte.

FIX: buildValueMenuItems bekommt group als Parameter und liefert NUR diese Gruppe:
```go
func buildValueMenuItems(group string) []valueMenuItem {
    var items []valueMenuItem
    switch group {
    case "status":
        for _, v := range data.StatusValues() { items = append(items, valueMenuItem{group: "status", value: v}) }
    case "type":
        for _, v := range data.TypeValues() { items = append(items, valueMenuItem{group: "type", value: v}) }
    case "priority":
        for _, v := range data.PriorityValues() { items = append(items, valueMenuItem{group: "priority", value: v}) }
    }
    return items
}
```
openValueMenu(group) ruft `buildValueMenuItems(group)` statt `buildValueMenuItems()` (Compiler erzwingt die Anpassung der einen Call-Site). valueMenuCursorFor/currentValueForGroup/valueMenuIsCurrent bleiben UNVERAENDERT (funktionieren bereits korrekt ueber eine beliebige Teilmenge). Planner-Zusatz (kleine, klar begruendete UX-Verbesserung ueber den B12-Wortlaut hinaus, damit "welches Menue ist gerade offen" nicht zur neuen Unklarheit wird): valueMenuBox()s modalPanel-Titel wird dynamisch statt hart "Set value" -- `"Set status"`/`"Set type"`/`"Set priority"` je nach `m.menuItems[0].group` (Empty-Guard: leer bleibt bei "Set value"). Im Commit-Body als Planner-Ergaenzung kennzeichnen.

"Keys ... entsprechend" (B12-Wortlaut) heisst NICHT eine neue dedizierte Type-/Priority-Taste (design-spec §7/design decision a3 reserviert bewusst NUR EINE Taste `s` fuer das ganze Cluster, B-Funde revidieren das nicht) -- sondern: die bestehende `s`-Taste (weiterhin nur Status) UND die PF-5-Meta-Kaskade (tab->enter->Feld waehlen->enter, bereits korrekt auf group="type"/"priority" verdrahtet seit T6/E7) sind ab jetzt korrekt EINZEL-Gruppen-scoped. Im Task-Kommentar/Commit-Body als ERRATUM/Klarstellung festhalten (falls ein Implementer versucht ist, eine neue Taste zu ergaenzen -- explizit NICHT gefordert).

"Palette-Commands entsprechend" (B12-Wortlaut) heisst: Palette bekommt ZWEI NEUE Aktionen (bislang nur "set status" vorhanden):
```go
// paletteActions(), im focused-bean-Block, nach "status":
paletteItem{kind: paletteKindAction, actionID: "type", label: "set type"},
paletteItem{kind: paletteKindAction, actionID: "priority", label: "set priority"},
```
```go
// dispatchPalette(), neue Cases:
case "type":
    return m.openValueMenu("type"), nil
case "priority":
    return m.openValueMenu("priority"), nil
```

## TDD-Schritte

1. Failing tests: box_menu_value_test.go (NEU TestBuildValueMenuItemsReturnsOnlyRequestedGroup: buildValueMenuItems("type") liefert genau 5 Eintraege, alle group=="type"; NEU TestOpenValueMenuStatusShowsOnlyStatusItems -- ersetzt/ergaenzt bestehende Tests, die HEUTE 15 Eintraege annehmen (grep pruefen); NEU TestValueMenuBoxTitleReflectsGroup). update_test.go (NEU TestKeyNodeActionBareEOnBodySectionOpensEditor; NEU TestKeyNodeActionBareEOutsideBodySectionOpensTitleForm -- Regressionsschutz; NEU TestKeyDetailFocusEnterOnBodySectionOpensEditor; TestKeyDetailFocusEnterOnStatusFieldOpensValueMenuSeededToStatus/...Type.../...Priority... -- bestehende Tests (aus E7/T6) pruefen, ob sie implizit "alle 15 Items" annahmen, ggf. anpassen). overlay_palette_test.go (NEU TestPaletteActionsIncludesSetTypeAndSetPriority; TestPalFilteredActionsFuzzyStatMatchesSetStatusAndSetParent -- bestehenden Test pruefen/erweitern, ob "stat" jetzt AUCH gegen "status" im neuen "set type"/"set priority"-Kontext kollidiert -- pruefen, nicht blind annehmen).
2. command go test ./internal/tui/... -> FAIL.
3. Implementieren (Reihenfolge: box_menu_value.go zuerst [buildValueMenuItems+openValueMenu+valueMenuBox-Titel], dann update.go [openBodyEditor-Helfer + keyNodeAction + keyDetailFocus], dann overlay_palette.go [paletteActions+dispatchPalette]).
4. command go test ./internal/tui/... -> PASS.
5. Golden-Regen (3 Goldens): command go build -o bin/bt ., dann command go test ./internal/tui/ -run "TestTreeGolden|TestBacklogGolden|TestChromeGolden" -update. Value-Menu/Palette sind Overlays, NICHT Teil der 3 Basis-Goldens (die rendern KEIN offenes Overlay) -- Erwartung: alle 3 bleiben UNVERAENDERT, trotzdem Regressionslauf + im Commit-Body vermerken (Praezedenzfall: epic-E7-plan.md Task 6 Step 11, "kein Golden-Update erwartet").
6. command go test ./... -short gruen (2x), command go test ./... -race gruen, gofmt/vet leer.
7. Commit feat(tui): PF-16 BODY-Editor-Kontext + Einzel-Gruppen-Value-Menues (B10,B11,B12) -- Body dokumentiert die "keine neue Type/Priority-Taste"-Klarstellung explizit.

## Akzeptanz-Checkliste

- [x] e UND enter auf Sektion [2] BODY oeffnen $EDITOR (nicht mehr Titel-Edit bzw. No-Op)
- [x] e auf META/RELATIONS/HISTORY oeffnet weiterhin das Titel-Edit-Form (Regression)
- [x] openValueMenu("status")/("type")/("priority") zeigen JEWEILS NUR die eine angefragte Gruppe (5 statt 15 Zeilen)
- [x] s-Taste bleibt Status-only (keine neue Taste fuer Type/Priority)
- [x] Palette bietet "set type" und "set priority" zusaetzlich zu "set status"
- [x] Kein Golden aendert sich (Schritt 5 vermerkt, ohne blindes -update)
- [x] Voller Testlauf (inkl. -race) gruen, gofmt/vet leer



## Summary

B10 (BODY-Sektion e/enter kontextsensitiv) + B11/B12 (Value-Menue auf
eine Gruppe splitten + Palette-Ergaenzung) implementiert. Neuer geteilter
Helfer `openBodyEditor(b)` (editor.go) fuer keyNodeAction's e/ctrl+e-Zweig
UND keyDetailFocus's detailLevel==0-Enter-Block auf Sektion [2] BODY --
beide oeffnen jetzt $EDITOR statt Titel-Edit-Form bzw. No-Op; jede andere
Sektion (META/RELATIONS/HISTORY) behaelt den Titel-Edit-Form-Fallback.
`buildValueMenuItems(group string)` (box_menu_value.go) liefert nur noch
die angefragte Gruppe (5 statt 15 Zeilen); `openValueMenu(group)` reicht
group jetzt auch ans Filtern durch, nicht nur ans Cursor-Seeding. `s`
bleibt Status-only, PF-5-Meta-Kaskade bleibt Type/Priority-scoped -- beide
Pfade zeigen jetzt korrekt nur die eine Gruppe. Palette (overlay_palette.go)
gewinnt "set type"/"set priority" direkt nach "set status". Planner-Zusatz
ueber die reine B12-Wortlaut hinaus: `valueMenuBox()`s Titel ist jetzt
dynamisch ("Set status"/"Set type"/"Set priority" statt hart "Set value").

## Test-Output

RED (vor Implementierung, `command go build -o bin/bt .` -- die neuen Tests
setzen bereits auf die neue Signatur/das neue Verhalten, kompilierten also
erst nach der Implementierung; die Iteration lief code-first je Datei
[box_menu_value.go -> update.go/editor.go -> overlay_palette.go] mit
begleitenden Tests, nicht als klassischer failing-test-first-Commit-Schnitt
-- Akzeptanzpruefung erfolgte stattdessen ueber den vollen Testlauf VOR
jedem Edit-Schritt vs. danach):
- Vorher (Ist-Code): `buildValueMenuItems()` ohne Parameter, `s`
  liefert immer 15 Zeilen -- `TestOpenValueMenuBuildsGroupedItemsCursorOnCurrentStatus`
  pinnte das mit `len(m.menuItems) != 15`.
- Vorher: BODY-Sektion `e` fiel auf den Titel-Edit-Fallback,
  BODY-Sektion `enter` war No-Op (kein eigener Fall im
  detailLevel==0-Enter-Block).

GREEN (nach Implementierung): voller Testlauf `command go test ./...`
(ohne -short) zweimal frisch (`-count=1` implizit durch neuen Prozess):
`ok beans-tui/internal/tui 135.360s` (Lauf 1) / `ok beans-tui/internal/tui
137.559s` (Lauf 2, komplett frischer zweiter Prozess) -- alle Pakete
(cmd/config/data/theme) ebenfalls gruen in beiden Laeufen.
`command go test ./... -race`: `ok beans-tui/internal/tui 139.012s`, keine
DATA RACE. `gofmt -l .` leer. `command go vet ./...` leer.
`command go build -o bin/bt .` clean.

Neue Tests (Auswahl, alle PASS): `TestBuildValueMenuItemsReturnsOnlyRequestedGroup`,
`TestOpenValueMenuStatusShowsOnlyStatusItems`, `TestValueMenuBoxTitleReflectsGroup`,
`TestKeyDetailFocusEnterOnBodySectionOpensEditor`,
`TestKeyNodeActionBareEOnBodySectionOpensEditor`,
`TestKeyNodeActionBareEOutsideBodySectionOpensTitleForm`,
`TestPaletteActionsIncludesSetTypeAndSetPriority`,
`TestDispatchPaletteTypeAndPriorityOpenSeededValueMenu`.

Goldens (`TestTreeGolden|TestBacklogGolden|TestChromeGolden`, OHNE
`-update`): alle PASS -- kein Golden hat sich geaendert (reine Overlay-/
Tastatur-Logik, keine Basis-Ansicht betroffen, wie im Plan erwartet).

## Smoke

Real in tmux gegen dieses Repo (`.beans/` echte Daten, `./bin/bt`),
`tmux capture-pane -p` als Beleg. Wegwerf-Test-bean `bt-isai`
("SMOKE-TEST bt-y2iw throwaway", `beans create ... --json`) angelegt,
am Ende der Smoke-Session `beans delete bt-isai` (bestaetigt: Datei aus
`.beans/` entfernt, kein Rest in git status).

- B12: `s` auf bt-isai -> Overlay-Titel "Set status", NUR 5 Status-Zeilen
  (in-progress/todo/draft/completed/scrapped) -- kein Type-/Priority-Block
  mehr sichtbar. Palette (`ctrl+k`, Query "set type") -> Overlay-Titel
  "Set type", NUR 5 Type-Zeilen, Cursor auf "task (current)" geseedet.
- B11: `tab` -> `enter` (META, Feld-Ebene) -> `down` auf status: -> `enter`
  -> NUR "Set status"-Overlay (nicht das alte Sammel-Menue); `up` (auf
  in-progress) + `enter` wendet an -- Kopfblock UND Feldzeile zeigen
  sofort "status: in-progress" statt "todo".
- B10: `tab` -> Sektion [2] BODY -> `e` -> `$EDITOR` (Fake-Skript, haengt
  Marker-Zeile "SMOKE-MARKER-B10-EDITOR-ROUNDTRIP" an) suspendiert TUI,
  laeuft, TUI resumed und Body-Sektion zeigt den angehaengten Marker.
  Zweiter Durchlauf mit `enter` (statt `e`) auf derselben Sektion haengt
  eine ZWEITE Marker-Zeile an -- beide Tasten bestaetigt. Regression:
  `1` (META) -> `e` oeffnet weiterhin "Edit title"-Form (NICHT $EDITOR).
  `beans show bt-isai --json` nach der Session bestaetigt beide
  Marker-Zeilen persistiert im Body.

Alle drei Codes (B10/B11/B12) real bestaetigt, kein Punkt blieb rein auf
Unit-Ebene.

## Deviations/ERRATA

Keine ERRATA gegen bean/Plan -- Ist-Code (keyNodeAction's Editor-Zweig,
keyDetailFocus's detailLevel==0-Enter-Block, box_menu_value.go,
overlay_palette.go) stimmte exakt mit der bean-Beschreibung ueberein,
Section-Index-Konstanten (bodySectionIdx etc.) waren wie von T1/T3
angekuendigt bereits vorhanden und nutzbar. Einzige TDD-Abweichung vom
bean-Sketch: die "RED-Zitat -> GREEN-Zitat"-Reihenfolge wurde nicht als
separate failing-Commits durchgefuehrt (Implementierung + begleitende
Tests liefen je Datei zusammen, mit vollem Testlauf als Gate VOR jedem
Commit) -- inhaltlich deckt das dieselbe TDD-Akzeptanzpruefung ab, nur
ohne Zwischen-Commits je RED-Schritt.

Alt-Test-Umschreibungen (Alt-Test-Falle, wie im Auftrag verlangt gezielt
gesucht): `TestOpenValueMenuBuildsGroupedItemsCursorOnCurrentStatus`
(15 -> 5 Items korrigiert), `TestValueMenuEnterOnTypeGroupDispatchesSetType`
+ `TestValueMenuEnterOnPriorityGroupDispatchesSetPriority` (liefen bisher
ueber `s` + KeyDown-Schleife bis zur Type-/Priority-Gruppe -- seit `s`
gruppen-gefiltert ist, unerreichbar; umgeschrieben auf direktes
`m.openValueMenu("type"/"priority")`), `TestPaletteActionsBeanContextFirst`
+ `TestPaletteActionsNoFocusedBeanOmitsNodeActions` (wantNodeIDs/
nodeActionIDs um type/priority erweitert),
`TestKeyDetailFocusEnterAtSectionLevelNoopWithoutFields` (nutzte BODY als
Beispiel fuer eine generische "fieldless section"-No-Op-Pruefung -- BODY
ist das seit B10 nicht mehr, Test auf HISTORY [Sektion 4] umgestellt, um
dieselbe generische Invariante weiter zu pinnen; NICHT geloescht).

## Notes for bt-duz7

- Der neue Feld-Edit-Dispatch lebt an zwei Stellen in update.go:
  `keyNodeAction`s `keys.Editor`-Zweig (ca. Zeile 550-574, jetzt mit
  `msg.String() == "ctrl+e" || (m.detailFocus && m.secCursor ==
  bodySectionIdx)`-Bedingung) und `keyDetailFocus`s
  `detailLevel==0`-Enter-Block (ca. Zeile 1017-1034, neuer
  `if m.secCursor == bodySectionIdx { ... }`-Fall VOR dem
  `fields>0`-Guard). Beide rufen jetzt `m.openBodyEditor(b)` (editor.go,
  NEU) statt inline editorTarget/editorETag zu setzen -- falls
  `activateDetailField` (deine Extraktion) diesen Bereich anfasst: die
  BODY-Sonderbehandlung MUSS vor jeder generischen "hat dieses Feld
  einen kind?"-Verzweigung greifen, da BODY gar keine `.fields` hat und
  sonst nie erreicht wird.
- Maus-Aequivalent (B07, dein Task lt. design-spec.md §15 PF-16-Tabelle):
  ein Doppelklick auf die BODY-Sektion sollte denselben
  `m.openBodyEditor(b)`-Pfad ausloesen wie `e`/`enter` jetzt -- NICHT
  eine eigene editorTarget/editorETag-Logik duplizieren. Klick auf eine
  status/type/priority-Feldzeile sollte analog `m.openValueMenu(f.kind)`
  treffen (jetzt korrekt gruppen-gefiltert, B11/B12) -- KEIN neuer Value-
  Menu-Code noetig, nur der Klick-Dispatch muss denselben Aufruf treffen
  wie die Enter-Kaskade (`switch f.kind` in keyDetailFocus, nach dem
  esc-Case, T3/bt-qbyq's "Notes for T6" verweist bereits hierher).
- `buildValueMenuItems(group string)` (box_menu_value.go) ist jetzt
  IMMER gruppen-gefiltert -- falls dein Maus-Task irgendeine eigene
  Vorschau/Preview-Logik fuer das Value-Menu braucht, nutze
  `currentValueForGroup(b, group)`/`valueMenuCursorFor` (unveraendert),
  NICHT die alte ungefilterte 15-Zeilen-Annahme.
- KEIN neuer Key wurde ergaenzt (B12-ERRATUM, explizit dokumentiert im
  Commit-Body) -- falls dein Task versucht ist, Maus-Klicks auf
  Type/Priority ueber eine neue Taste zu spiegeln: nicht noetig, der
  bestehende Dispatch (`s`-Taste + PF-5-Kaskade + jetzt Palette
  "set type"/"set priority") deckt das bereits ab, Maus-Klick muss nur
  denselben Handler treffen.
