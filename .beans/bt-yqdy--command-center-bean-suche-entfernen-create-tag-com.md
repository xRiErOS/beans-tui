---
# bt-yqdy
title: 'Command-Center: Bean-Suche entfernen + create-tag-Command'
status: completed
type: task
priority: normal
created_at: 2026-07-15T21:10:08Z
updated_at: 2026-07-16T00:14:57Z
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

- [x] Command-Center zeigt bei JEDER Query AUSSCHLIESSLICH Commands, nie Bean-Treffer
- [x] paletteKindBean/palFilteredBeans/palBeanMatches/palBleve*-Unterbau vollstaendig entfernt, `command go build ./...` sauber
- [x] `/`s eigene Bleve-Suche (searchBleveIDs/searchBleveFor) unveraendert funktional (Regressionstest)
- [x] Tag-Picker-Footer zeigt "n:New tag" (vorher fehlend)
- [x] Command-Center bietet "create tag", oeffnet direkt den Neuanlage-Input im Tag-Picker
- [x] "create tag" ohne fokussiertes Bean ist ein sauberer No-Op (kein latenter tagInputActive-State)
- [x] Goldens regeneriert/verifiziert (voraussichtlich alle 3 unveraendert, im Commit-Body vermerkt)
- [x] Voller Testlauf (inkl. -race) gruen, gofmt/vet leer

## Summary

B13 implemented: `overlay_palette.go`'s former bean-search half
(`paletteKindBean`, `palFilteredBeans`/`palBeanMatches`,
`paletteBeanResultCap`) removed; `palFiltered()` filters ONLY the action
pool now; `paletteBox()` renders ONE `menuList` call (no split/separator);
`keyPalette()`'s rune/backspace branches resync `palList` and return
directly (no Bleve dispatch, `dispatchPaletteBleveIfDue` removed).
`types.go` (`palBleveIDs`/`palBleveFor`/`palBleveLoading`),
`messages.go` (`paletteBleveResultMsg`/`paletteSearchCmd`), `update.go`
(`applyPaletteBleveResult`/`maybePaletteBleveCmd` + the `Update()` case)
removed compiler-steered (production code first, `go build ./...` clean
with zero further prod-code edits, THEN the now-obsolete
`overlay_palette_test.go` sections). `/`'s own, fully separate Bleve
search (`searchBleveIDs`/`searchBleveFor`, `search.go`) untouched --
regression-verified by the full suite. `view_lobby.go`'s stray doc-comment
reference to the now-removed `palFilteredBeans` updated (comment only, no
behavior change).

B14 implemented: (a) `keys.NewTag` (keymap.go, `n`, wired into
`helpGroups()`'s Actions group) makes the Tag-Picker's pre-existing
free-text new-tag sub-mode footer-renderable; `keyTagPicker`
(box_picker_tag.go) switched from `msg.String() == "n"` to
`keybind.Matches(msg, keys.NewTag)`; `tagPickerLocalBindings()`
(footer_context.go) now includes it, so the OUTER Footer Zone 3 shows
"n:New tag" (verified missing before, confirmed present after -- see
Smoke). (b) `paletteActions()` gains a `create_tag` node action ("create
tag", placed directly after "set tags" -- both tag-related, an
implementer ordering decision the bean/plan left open); `dispatchPalette`'s
`"create_tag"` case opens the Tag-Picker AND its new-tag input in one step
(`m.openTagPicker().openTagInput()`), gated by a MANDATORY
`focusedBean()==nil` guard -- `openTagPicker()` is itself already
no-op-safe on a missing focused bean, but chaining `.openTagInput()`
straight onto that would unconditionally set `tagInputActive=true` with
no picker actually open (a latent, unreachable state) -- the guard
short-circuits before that chain runs.

## Test-Output

RED (B13, echter Compiler-Beleg -- production code deleted first, tests
still referenced the removed symbols):
```
$ command go vet ./...
vet: internal/tui/overlay_palette_test.go:395:13: m.palFilteredBeans undefined (type model has no field or method palFilteredBeans)
```
GREEN after removing/renaming the obsolete test sections
(`TestPalFilteredBeansEmptyQueryNone`, `TestPalFilteredBeansLocalSubstring
BelowThreshold`, `fixtureManyBeans`, `TestPalFilteredBeansCappedAt20`,
`TestPalFilteredBeansSortedCanonically`, `TestPalFilteredOrderActionsBeforeBeans`,
`TestDispatchPaletteBeanJumpsCursorAndSwitchesToBrowse`,
`TestDispatchPaletteBeanJumpResetsDetailFocus`,
`TestKeyPaletteDispatchesBleveOnQueryGrowth`,
`TestApplyPaletteBleveResultDiscardsStaleQuery`,
`TestApplyPaletteBleveResultAppliedWhenQueryStillCurrent`,
`TestPaletteSearchCmdTagsResultWithQuery` all deleted;
`TestPalFilteredEmptyQueryReturnsAllActionsNoBeans` renamed to
`TestPalFilteredEmptyQueryReturnsAllActions`):
```
$ command go vet ./...
(clean)
$ command go test ./internal/tui/... -run "TestPalette|TestPalFiltered|TestOpenPalette|TestKeyPalette|TestHandleKeyCtrlK|TestDispatchPalette" -v
... (24 tests) ...
PASS
ok  	beans-tui/internal/tui	0.504s
```

RED (B14, `keys.NewTag` footer-hint, echter Compiler-Beleg -- tests written
before the field existed):
```
$ command go vet ./...
vet: internal/tui/footer_context_test.go:81:55: keys.NewTag undefined (type keyMap has no field or method NewTag)
```
GREEN after adding `keys.NewTag` + wiring `tagPickerLocalBindings()` +
`helpGroups()`:
```
$ command go test ./internal/tui/... -run "TestNewTagKeyBound|TestTagPickerLocalBindingsIncludesNewTag|TestContextualLocalHintOverlayTagPickerShowsNewTag|TestHelpGroupsCoverEveryBindingExactlyOnce|TestGlobalBindingsExactSet|TestNoDuplicateBindingBetweenGlobalAndAnyLocalHintList|TestTagPickerNewTag" -v
--- PASS: TestTagPickerNewTagValidatesRegex
--- PASS: TestTagPickerNewTagAddsPendingItem
--- PASS: TestTagPickerLocalBindingsIncludesNewTag
--- PASS: TestContextualLocalHintOverlayTagPickerShowsNewTag
--- PASS: TestNewTagKeyBound
--- PASS: TestGlobalBindingsExactSet
--- PASS: TestNoDuplicateBindingBetweenGlobalAndAnyLocalHintList
--- PASS: TestHelpGroupsCoverEveryBindingExactlyOnce
PASS
```

RED (B14, `dispatchPalette`'s `"create_tag"` case, echter Revert-Beleg --
the case temporarily removed via a targeted Edit, then restored):
```
$ command go test ./internal/tui/... -run "TestDispatchPaletteCreateTag" -v
=== RUN   TestDispatchPaletteCreateTagOpensTagPickerInputMode
    overlay_palette_test.go:475: overlay = 0, want overlayTagPicker
--- FAIL: TestDispatchPaletteCreateTagOpensTagPickerInputMode
=== RUN   TestDispatchPaletteCreateTagNoFocusedBeanNoOp
--- PASS: TestDispatchPaletteCreateTagNoFocusedBeanNoOp
FAIL
```
(the guard test alone stayed green even against the reverted code, since
its own focusedBean()==nil short-circuit is unaffected by the removed
case -- confirming it exercises a DIFFERENT code path than the
happy-path test, not a tautology)
GREEN after restoring the case:
```
$ command go test ./internal/tui/... -run "TestDispatchPaletteCreateTag" -v
--- PASS: TestDispatchPaletteCreateTagOpensTagPickerInputMode
--- PASS: TestDispatchPaletteCreateTagNoFocusedBeanNoOp
PASS
```

Golden-Gegenbeleg (OHNE `-update`, MUSS gruen bleiben -- Palette/Tag-Picker
sind Overlays, kein Teil der 3 Basis-Goldens):
```
$ command go test ./internal/tui/ -run "TestTreeGolden|TestBacklogGolden|TestChromeGolden" -v
--- PASS: TestChromeGolden
--- PASS: TestTreeGolden
--- PASS: TestTreeGoldenDeterministic
--- PASS: TestBacklogGolden
--- PASS: TestBacklogGoldenDeterministic
PASS
```
`git status --short internal/tui/testdata/` leer -- kein Golden geaendert.

Voll-Lauf (362 Testfunktionen in `internal/tui`, `go test -list` gezaehlt),
DREI frische Prozesse: `command go test ./...` -> `ok beans-tui/internal/tui
136.020s` (Lauf 1); `command go test ./... -count=1` -> `ok
beans-tui/internal/tui 136.816s` (Lauf 2, alle Pakete cmd/config/data/theme
ebenfalls PASS in beiden Laeufen); `command go test ./... -race` -> `ok
beans-tui/internal/tui 139.622s`, keine DATA RACE. `gofmt -l .` leer.
`command go vet ./...` leer. `command go build -o bin/bt .` clean.

## Smoke

Real in tmux gegen dieses Repo (`.beans/` echte Daten, `./bin/bt`),
tmux 120x40, `tmux send-keys`/`capture-pane -p` als Beleg:

- ctrl+k, `apmy` eingetippt (ID-Substring des fokussierten Beans `bt-apmy`,
  haette VORHER einen Bean-Treffer gezeigt) -> "(no matches)", KEINE
  Bean-Zeile.
- ctrl+k, `devd` eingetippt (Titel-Wort-Substring, "beans-tui v1 — devd-TUI-
  Port auf beans") -> ebenfalls "(no matches)", KEINE Bean-Zeile.
- ctrl+k, `create` eingetippt -> "▸ create tag" (cursor-selektiert, node
  action zuerst) UND "create bean" (global action) beide sichtbar, richtige
  Reihenfolge.
- enter auf "create tag" -> "New tag"-Modal oeffnet SOFORT (kein
  Zwischenschritt über den Tag-Picker-Listenview) mit "enter:create
  esc:cancel" + leerem Input "new tag (a-z0-9, hyphen-separated)".
  Neuanlage NICHT durchgezogen (esc/esc zum Verwerfen) -- `git status
  --short .beans/` davor UND danach leer, keine Mutation.
- `t` gedrueckt (Tag-Picker öffnen) -> AEUSSERE Footer-Zeile (Zone 3, unten
  im Terminal, NICHT die Picker-eigene Inline-Hint-Zeile im Modal) zeigt
  jetzt "↑/i:up  ↓/k:down  space/x:Toggle facet  n:New tag  enter:open/
  confirm  esc:back" -- "n:New tag" bestaetigt sichtbar, vorher fehlend.
  Die Picker-eigene Inline-Hint-Zeile ("space/x:toggle  n:new tag
  enter:save  esc:discard") war unveraendert bereits vorher da (T3-Fund,
  nicht Teil dieses Fixes).
- Session sauber beendet, `git status --short .beans/` final leer -- keine
  Mutation ueber die gesamte Smoke-Session.

## Deviations/ERRATA

Keine ERRATA gegen bean/Plan -- Ist-Code (`paletteActions`/
`dispatchPalette`-Signaturen, `openTagPicker`/`openTagInput`-Verhalten,
`tagPickerLocalBindings`/`helpGroups`-Struktur) stimmte exakt mit den
bean-Sketches ueberein; das B14-ERRATUM war bereits vom Planner verifiziert
(bean-Text) und traf zu.

Implementer-Entscheidungen (wie im Auftrag vorgesehen):

1. **`create_tag`-Platzierung in `paletteActions()`**: direkt nach "set
   tags" (im focused-bean-Block) -- der Plan spezifiziert NUR den Block,
   nicht die exakte Position darin; "gruppiert mit der verwandten Aktion"
   ist die niedrigste-Risiko-Lesart.
2. **`paletteItemKind`-Typ NICHT entfernt**, obwohl nach B13 nur noch EIN
   Wert (`paletteKindAction`) existiert -- der bean-Text nennt explizit nur
   die `paletteKindBean`-KONSTANTE zur Entfernung, nicht den Typ selbst;
   beibehalten als Vorbereitung fuer eine zukuenftige echte dritte Art,
   analog der urspruenglichen Typ-Begruendung (T1/T2-Praezedenzfall).
3. **`TestPalFilteredEmptyQueryReturnsAllActionsNoBeans` umbenannt** (nicht
   dupliziert) zu `TestPalFilteredEmptyQueryReturnsAllActions` -- der
   Qualifier "NoBeans" ist nach B13 keine query-abhaengige Eigenschaft mehr
   (gilt jetzt IMMER, nicht nur bei leerer Query), die bean-TDD-Schritte
   nannten genau diese Test-Pruefung+Umbenennungs-Option explizit.
4. **Zwei zusaetzliche `dispatchPalette`-Tests** (`TestDispatchPaletteCreate
   TagOpensTagPickerInputMode`, `TestDispatchPaletteCreateTagNoFocusedBean
   NoOp`) ueber die im TDD-Schritte-Abschnitt namentlich genannte
   `TestPaletteActionsIncludesCreateTag` hinaus ergaenzt -- die
   Akzeptanz-Checkliste selbst verlangt sowohl das Oeffnen-Verhalten als
   auch den No-Op-Guard, beide brauchten eigene Testabdeckung.

## Notes for bt-d8kc

- Der neue Footer-Hint lebt in `tagPickerLocalBindings()`
  (`footer_context.go`) -- ein OVERLAY-lokaler Hint (Footer Zone 3's
  `contextualLocalHint`-Dispatch ueber `overlayLocalBindings(overlayTagPicker)`),
  NICHT der globale Footer/`globalBindings()`/`browseRepoLocalBindings()`/
  `backlogLocalBindings()`, die bt-d8kc neu aufbaut. Kein Konflikt zu
  erwarten -- bt-d8kc's D06-Neuspezifikation betrifft explizit nur die
  VIEW-lokalen Listen (Browse/Backlog), nicht die Overlay-lokalen Sets in
  `footer_context.go`.
- `keys.NewTag` ist ein NEUES `keyMap`-Feld (keymap.go) -- bereits in
  `helpGroups()`'s "Actions"-Gruppe verdrahtet (Drift-Guard
  `TestHelpGroupsCoverEveryBindingExactlyOnce` deckt es ab, verifiziert
  gruen). bt-d8kc's `globalBindings()`-Kuerzung auf 4 Items und der
  `Blocking`-Remap (`B`->`r`) beruehren `keys.NewTag` nicht -- `n` ist in
  keinem der beiden betroffenen Sets (Header-Globals, Footer-Q06-Liste)
  vertreten, keine Kollision.
- `create_tag` (Palette) ist ein neuer `dispatchPalette`-Case, der
  `openTagPicker().openTagInput()` direkt aufruft -- KEIN neuer Key/keine
  neue globale Keybinding, beruehrt bt-d8kc's Footer-/Header-Arbeit nicht.
