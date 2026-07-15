---
# bt-y2iw
title: 'Feld-Edit-Dispatch: BODY-Editor-Kontext + Einzel-Gruppen-Value-Menues'
status: todo
type: task
priority: high
created_at: 2026-07-15T21:08:25Z
updated_at: 2026-07-15T21:08:25Z
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

- [ ] e UND enter auf Sektion [2] BODY oeffnen $EDITOR (nicht mehr Titel-Edit bzw. No-Op)
- [ ] e auf META/RELATIONS/HISTORY oeffnet weiterhin das Titel-Edit-Form (Regression)
- [ ] openValueMenu("status")/("type")/("priority") zeigen JEWEILS NUR die eine angefragte Gruppe (5 statt 15 Zeilen)
- [ ] s-Taste bleibt Status-only (keine neue Taste fuer Type/Priority)
- [ ] Palette bietet "set type" und "set priority" zusaetzlich zu "set status"
- [ ] Kein Golden aendert sich (Schritt 5 vermerkt, ohne blindes -update)
- [ ] Voller Testlauf (inkl. -race) gruen, gofmt/vet leer
