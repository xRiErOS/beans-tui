---
# bt-yqdy
title: 'Command-Center: Bean-Suche entfernen + create-tag-Command'
status: in-progress
type: task
priority: normal
created_at: 2026-07-15T21:10:08Z
updated_at: 2026-07-15T23:53:56Z
parent: bt-ntoz
blocked_by:
    - bt-y2iw
---

E8 Task 7 — deckt B13 (Bean-Suche aus dem Command-Center entfernen) und B14 (Tag-Neuanlage entdeckbar machen: Footer-Hint + Palette-Command) aus bean bt-ntoz. Quelle: design-spec.md §15 PF-16 + "US-04-Revision". Ist-Code: internal/tui/overlay_palette.go, internal/tui/types.go, internal/tui/messages.go, internal/tui/update.go (paletteBleveResultMsg-Case), internal/tui/keymap.go, internal/tui/box_picker_tag.go, internal/tui/footer_context.go. blocked_by bt-y2iw (T6) -- T6 ergaenzt bereits "set type"/"set priority" in paletteActions()/dispatchPalette(), dieselbe Funktion, die diese Task fuer B13/B14 weiter anfasst -- Reihenfolge vermeidet Merge-Ueberraschungen.

## B13 — Bean-Suche aus dem Command-Center entfernen

HEUTE mischt das Command-Center (ctrl+k) Bean-Treffer unter die Commands (palFilteredBeans, Bleve-Staleness-Guard). PO: stoerend, Bean-Suche gehoert exklusiv zu `/`. FIX -- COMPILER-GESTEUERTE Entfernung (bewaehrtes Muster aus epic-E7-plan.md Task 1/PF-14: entfernen, `go build ./...`, jeden Fehler beheben, wiederholen -- robuster als ein reiner Grep-Sweep):

Zu entfernen:
- overlay_palette.go: `paletteKindBean`-Konstante, `paletteItem.bean`-Feld, `palFilteredBeans()`, `palBeanMatches()`, `paletteBeanResultCap`-Konstante, `palFiltered()`s Bean-Anhaenge-Zeile (`out = append(out, m.palFilteredBeans()...)` -- `palFiltered()` liefert danach NUR NOCH die gefilterten Actions), `dispatchPalette()`s `case paletteKindBean:`-Block, `paletteBox()`s split/beanItems-Rendering (vereinfacht auf reines Actions-Rendering, KEIN zweiter menuList-Aufruf, KEIN Separator mehr).
- types.go: `palBleveIDs`/`palBleveFor`/`palBleveLoading`-Felder (palQuery/palList BLEIBEN -- weiterhin fuer die reine Actions-Fuzzy-Filterung gebraucht).
- messages.go: `paletteBleveResultMsg`-Typ, `paletteSearchCmd()`.
- update.go: `case paletteBleveResultMsg:`-Zweig in Update(), `applyPaletteBleveResult()`, `maybePaletteBleveCmd()`.
- overlay_palette.go: `dispatchPaletteBleveIfDue()` -- keyPalette()s Backspace/Rune-Zweige rufen stattdessen direkt `return m, nil` nach dem `m.palList.setLen(...)`-Resync (kein Bleve-Dispatch mehr noetig).

NICHT anfassen: `/`s eigene Bleve-Suche (searchBleveIDs/searchBleveFor/searchBleveLoading, m.searchQuery -- KOMPLETT SEPARATER Mechanismus, palBleve* war NUR die Palette-eigene Kopie).

design-spec.md §10 US-04 ist BEREITS per PF-16/"US-04-Revision" aktualisiert (design-spec.md, dieser Plan) -- kein weiterer Doku-Schritt hier noetig, nur der Code-Nachvollzug.

## B14 — Tag-Neuanlage entdeckbar machen

VERIFIZIERT (T3-Sweep-Fund, PO-Frage aus bt-ntoz beantwortet): der "Neuer Tag"-Modus in box_picker_tag.go ist NICHT kaputt -- `keyTagPicker()` behandelt bereits `msg.String() == "n"` korrekt (oeffnet `openTagInput()`), UND `tagPickerBox()`s EIGENE Inline-Hint-Zeile (Zeile ~305) zeigt bereits "space/x:toggle  n:new tag  enter:save  esc:discard". Das Problem ist AUSSCHLIESSLICH Entdeckbarkeit: die AEUSSERE Footer-Zeile (Zone 3, footer_context.go's tagPickerLocalBindings()) zeigt "n" NICHT (die Inline-Hint ist eine SEPARATE, leicht zu uebersehende Flaeche INNERHALB des Modal-Bodys).

FIX (a) — Footer-Hint: `n` wird zu einer richtigen keybind.Binding (fuer Footer-Rendering ueber renderBindings() gebraucht -- ein roher msg.String()-Vergleich kann nicht gerendert werden):
```go
// keymap.go, keyMap-Struct:
NewTag keybind.Binding // n — new tag (Tag-Picker free-text sub-mode)
// newKeyMap():
NewTag: keybind.NewBinding(keybind.WithKeys("n"), keybind.WithHelp("n", "New tag")),
```
helpGroups() ergaenzen (Drift-Guard TestHelpGroupsCoverEveryBindingExactlyOnce erzwingt es, "Actions"-Gruppe passt). box_picker_tag.go: `case msg.String() == "n":` -> `case keybind.Matches(msg, keys.NewTag):`. footer_context.go: `tagPickerLocalBindings()` ergaenzt `keys.NewTag`:
```go
func tagPickerLocalBindings() []keybind.Binding {
    return []keybind.Binding{keys.Up, keys.Down, keys.Toggle, keys.NewTag, keys.Enter, keys.Back}
}
```

FIX (b) — Palette-Command "create tag":
```go
// paletteActions(), im focused-bean-Block:
paletteItem{kind: paletteKindAction, actionID: "create_tag", label: "create tag"},
```
```go
// dispatchPalette():
case "create_tag":
    if m.focusedBean() == nil {
        return m, nil
    }
    return m.openTagPicker().openTagInput()
```
WICHTIG: der Guard `if m.focusedBean() == nil { return m, nil }` ist PFLICHT (nicht optional wie bei "tags") -- openTagPicker() ist selbst bereits no-op-sicher bei fehlendem Bean (gibt m unveraendert zurueck, overlay bleibt overlayNone), aber das direkte Verketten von `.openTagInput()` OHNE diesen Guard wuerde `tagInputActive=true` setzen OBWOHL der Picker nie geoeffnet wurde (m.overlay bleibt overlayNone) -- ein latenter, unerreichbarer State. Diesen Unterschied im Commit-Body kurz begruenden.

Tag-Management-**Page** (bt-6oyy) bleibt bewusst v1.1 (D08, bereits als Body-Nachtrag auf bt-6oyy dokumentiert) -- B14 ist NUR die v1-Minimal-Loesung (Neuanlage sichtbar machen), KEINE zentrale Tag-Verwaltung.

## TDD-Schritte

1. Failing tests: overlay_palette_test.go (die Bean-Treffer-Tests -- TestPalFilteredBeans*, TestPalFilteredOrderActionsBeforeBeans, TestDispatchPaletteBeanJumpsCursorAndSwitchesToBrowse, TestDispatchPaletteBeanJumpResetsDetailFocus, TestKeyPaletteDispatchesBleveOnQueryGrowth, TestApplyPaletteBleveResult*, TestPaletteSearchCmdTagsResultWithQuery -- per grep "^func Test" internal/tui/overlay_palette_test.go final ermitteln -- werden GELOESCHT, nicht angepasst, da die Funktionalitaet selbst entfaellt; TestPaletteActionsBeanContextFirst/TestPalFilteredEmptyQueryReturnsAllActionsNoBeans pruefen ob sie NUR Actions testen [dann anpassen/umbenennen] oder auch Bean-Anteile [dann Bean-Teil entfernen]). NEU TestPaletteActionsIncludesCreateTag. keymap_test.go (TestHelpGroupsCoverEveryBindingExactlyOnce deckt NewTag automatisch ab, sobald helpGroups() es listet -- kein manueller Test noetig, aber verifizieren). box_picker_tag_test.go (bestehenden "n"-Test auf keybind.Matches(msg, keys.NewTag) statt raw string pruefen). footer_context_test.go (NEU TestTagPickerLocalBindingsIncludesNewTag).
2. command go test ./internal/tui/... -> FAIL (Compile-Fehler zuerst -- Dateien/Felder loeschen, `command go build ./...` iterativ bis sauber, DANN Test-Dateien bereinigen, DANN Tests gruen -- exakt das T1/PF-14-Muster, hier auf overlay_palette.go/types.go/messages.go/update.go angewendet).
3. command go build ./... -> iterieren bis sauber.
4. command go test ./internal/tui/... -> PASS.
5. Golden-Regen (3 Goldens): command go build -o bin/bt ., dann command go test ./internal/tui/ -run "TestTreeGolden|TestBacklogGolden|TestChromeGolden" -update. Palette ist ein Overlay (nicht Teil der 3 Basis-Goldens) -- Erwartung: alle 3 UNVERAENDERT, trotzdem Regressionslauf + im Commit-Body vermerken.
6. command go test ./... -short gruen (2x), command go test ./... -race gruen, gofmt/vet leer.
7. Manueller Beleg (tmux-Smoke, kurz): ctrl+k, Query eintippen die vorher Bean-Treffer gezeigt haette (z.B. ein bekannter Bean-Titel-Teilstring) -> NUR "(no matches)" oder Command-Treffer, KEINE Bean-Zeile. t-Taste -> Footer zeigt jetzt "n:New tag". ctrl+k -> "create" eintippen -> "create tag" erscheint -> enter -> Tag-Picker oeffnet DIREKT im Neuanlage-Input. Ergebnis im Commit-Body.
8. Commit refactor(tui)!: PF-16 Bean-Suche aus Command-Center entfernt (B13) + feat(tui): create-tag-Command + Footer-Hint (B14) -- ODER als EIN Commit mit klar getrennten Body-Absaetzen (Implementer-Entscheidung, beide B-Codes gehoeren zur selben Datei/demselben Funktionsbereich) -- Body zitiert design-spec.md §10 US-04-Revision + die openTagPicker().openTagInput()-Guard-Begruendung.

## Akzeptanz-Checkliste

- [ ] Command-Center zeigt bei JEDER Query AUSSCHLIESSLICH Commands, nie Bean-Treffer
- [ ] paletteKindBean/palFilteredBeans/palBeanMatches/palBleve*-Unterbau vollstaendig entfernt, `command go build ./...` sauber
- [ ] `/`s eigene Bleve-Suche (searchBleveIDs/searchBleveFor) unveraendert funktional (Regressionstest)
- [ ] Tag-Picker-Footer zeigt "n:New tag" (vorher fehlend)
- [ ] Command-Center bietet "create tag", oeffnet direkt den Neuanlage-Input im Tag-Picker
- [ ] "create tag" ohne fokussiertes Bean ist ein sauberer No-Op (kein latenter tagInputActive-State)
- [ ] Goldens regeneriert/verifiziert (voraussichtlich alle 3 unveraendert, im Commit-Body vermerkt)
- [ ] Voller Testlauf (inkl. -race) gruen, gofmt/vet leer
