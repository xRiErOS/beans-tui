---
# bt-r92i
title: T2 — Tag-Management-Page Grundgerüst (read-only)
status: completed
type: task
priority: normal
created_at: 2026-07-16T15:44:24Z
updated_at: 2026-07-16T16:45:02Z
parent: bt-362n
blocked_by:
    - bt-49hh
---

T2 — Page-Grundgerüst: `viewTagManagement`, read-only Liste (Epic
`bt-362n`, D05-D09). `blocked_by` T1 (`internal/data/tagdefs.go` muss
existieren). Liefert die Page als NAVIGIERBARES, aber noch reines
Lese-Feature — Create/Delete/Rename folgen in T3/T4/T5.

## Ziel

Neuer Top-Level-View, erreichbar über das Command-Center („go to tags"),
zeigt die UNION aus definierten + frei-in-Verwendung-Tags (D09) als
Einzel-Pane-Liste mit Verwendungszähler. Kein Master-Detail (D08),
`enter` ist ein dokumentierter No-Op (reserviert für ein späteres
Drilldown-Fast-Follow).

## Betroffene Dateien/Symbole

- `internal/tui/types.go`:
  - `viewID`-Enum (`const ( viewBrowseRepo ... )`) wächst um
    `viewTagManagement` (neuer, letzter Wert — NIEMALS bestehende Werte
    umsortieren, das würde jeden persistierten/getesteten Vergleich
    gegen den iota-Wert brechen, obwohl aktuell nichts iota numerisch
    persistiert — Vorsichtsmaßnahme, mirrort den Kommentar-Stil der
    bestehenden drei Werte).
  - `model` bekommt: `tagMgmtRows []tagRegistryRow`,
    `tagMgmtCursor listState` (Cursor über die gerenderte Zeilenliste,
    reuse `listState` wie `backlogList`).
- `internal/tui/view_tag_management.go` (neu):
  - `type tagRegistryRow struct { name string; count int; defined bool }`
  - `func tagRegistryRows(idx *data.Index, defs []string) []tagRegistryRow`
    — D09-Sortierung: definierte Gruppe zuerst (alpha), dann
    undefinierte-in-Verwendung (Count absteigend, dann alpha als
    Tie-Break, mirrort `collectTagCounts`s eigenen Tie-Break). Nutzt
    intern dieselbe Zähl-Logik wie `collectTagCounts`
    (`box_picker_tag.go`) — Planner-Entscheidung: `collectTagCounts`
    NICHT umbenennen/verschieben (T6 erweitert es ohnehin um den
    `defined`-Parameter, danach kann `tagRegistryRows` diese ERWEITERTE
    Fassung direkt aufrufen statt eine zweite Zähl-Implementierung zu
    pflegen — s. T2s eigene ERRATUM-Notiz unten, falls T2 vor T6 landet).
  - `func (m model) openTagManagementPage() (tea.Model, tea.Cmd)` — lädt
    Registry frisch (`m.client.LoadTagDefs()`, D03, synchron), baut
    `tagMgmtRows` via `tagRegistryRows(m.idx, defs)`, setzt `m.view =
    viewTagManagement`, resettet `tagMgmtCursor` (`setLen`).
  - `func (m model) tagManagementChrome(innerW int) (head, localKeys string)`
    — mirrort `backlogChrome`: `breadcrumb(m.repoLabel(), "Tags", "",
    innerW)` (D07: GlobalHint LEER, nicht `renderBindings(globalBindings())`)
    + `footer(renderBindings(tagManagementLocalBindings()), innerW)`.
  - `func tagManagementLocalBindings() []keybind.Binding` — in T2 vorerst
    `{keys.Up, keys.Down, keys.Back}` (T3/T4/T5 hängen ihre eigenen
    Bindings anschließend an, s. deren Task-Bodies — EIN gemeinsamer
    Funktionsrumpf, kein Duplikat je Task).
  - `func (m model) viewTagManagement() string` — mirrort `viewBacklog()`s
    Grundgerüst (innerW/innerH-Berechnung, `Chrome`-Bausteine,
    `renderPane` für die Zeilenliste, `outerBorder`, `composeOverlays`
    am Ende) — EIN Pane, keine `JoinHorizontal`. Jede Zeile:
    Marker-Spalte (D10-Konvention vorgezogen: reserviert, hier noch ohne
    Picker-Bezug — s. Akzeptanz) + Name + rechtsbündiger Count.
  - `func (m model) keyTagManagement(msg tea.KeyMsg) (tea.Model, tea.Cmd)`
    — Up/Down bewegen `tagMgmtCursor` (reuse `navKey`), `enter` = handled
    No-Op (D08, mit Kommentar der das Fast-Follow referenziert), `esc`
    (`keys.Back`) → `m.view = viewBrowseRepo` (D03/D06-Pendant: EINE
    Ebene zurück, mirrort Lobbys `esc`).
- `internal/tui/update.go` (`handleKey`): neuer Check `if m.view ==
  viewTagManagement { return m.keyTagManagement(msg) }` — EXAKT an
  derselben Stelle wie der bestehende `if m.view == viewLobby { return
  m.keyLobby(msg) }`-Block (Zeile ~879, D06: full-capture, VOR den
  ctrl+k/?/p-Bare-Matches).
- `internal/tui/overlay_palette.go`: `paletteActions` bekommt einen
  neuen globalen Eintrag `paletteItem{kind: paletteKindAction, actionID:
  "go_tags", label: "go to tags"}` (gruppiert neben `"go to repo
  picker"`/`"go to settings"`, gleiche Sektion — Reihenfolge:
  Planner-Entscheidung, direkt VOR `"settings"` als letztem Eintrag,
  mirrort wie `"repo_picker"` vor `"settings"` einsortiert wurde);
  `dispatchPalette` bekommt den `case "go_tags": return
  m.openTagManagementPage()`.
- `internal/tui/view_browse_repo.go` (`View()`-Dispatcher, Zeile ~700):
  neuer `case viewTagManagement: return m.viewTagManagement()`.

## TDD (RED zuerst)

```go
// view_tag_management_test.go (neu)
func TestTagRegistryRowsDefinedFirstAlphaThenFreeByCountDesc(t *testing.T) {
    idx := data.NewIndex([]data.Bean{
        {ID: "b1", Tags: []string{"zzz-free"}},
        {ID: "b2", Tags: []string{"zzz-free"}},
        {ID: "b3", Tags: []string{"aaa-free"}},
    })
    rows := tagRegistryRows(idx, []string{"defined-b", "defined-a"})
    want := []string{"defined-a", "defined-b", "zzz-free", "aaa-free"}
    var got []string
    for _, r := range rows {
        got = append(got, r.name)
    }
    if !reflect.DeepEqual(got, want) {
        t.Fatalf("want %v, got %v", want, got)
    }
}

func TestTagRegistryRowsIncludesUnusedDefinedTagWithZeroCount(t *testing.T) {
    idx := data.NewIndex(nil)
    rows := tagRegistryRows(idx, []string{"unused"})
    if len(rows) != 1 || rows[0].count != 0 || !rows[0].defined {
        t.Fatalf("want one zero-count defined row, got %+v", rows)
    }
}
```

```go
// update_test.go oder view_tag_management_test.go
func TestKeyTagManagementEscReturnsToBrowse(t *testing.T) {
    m := newModel(nil, "")
    m.view = viewTagManagement
    nm, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
    if nm.(model).view != viewBrowseRepo {
        t.Fatalf("want viewBrowseRepo, got %v", nm.(model).view)
    }
}

func TestHandleKeyOnTagManagementViewDoesNotLeakToNodeAction(t *testing.T) {
    // Regression guard for D06's own rationale: `d` while on the tag
    // page must never open the bean Delete-Confirm.
    m := newModel(nil, "")
    m.view = viewTagManagement
    nm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
    if nm.(model).overlay != overlayNone {
        t.Fatalf("want no overlay opened, got %v", nm.(model).overlay)
    }
}
```

## Golden-Strategie

GEGENBELEG für Tree/Backlog/Chrome (dieser Task fügt einen NEUEN
Render-Pfad hinzu, berührt aber keinen bestehenden — nach `command go
build -o bin/bt .`: `command go test ./internal/tui/ -run
"TestTreeGolden|TestBacklogGolden|TestChromeGolden"` OHNE `-update`, MUSS
grün bleiben, im Commit-Body explizit als „unverändert" festhalten). Ob
`viewTagManagement()` eine EIGENE Golden-Suite bekommt, ist
Implementer-Entscheidung (mirrort E9 T7/T8s Freiheit) — mindestens ein
paar `View()`-Snapshot-Tests sind sinnvoll, da dies der EINZIGE
Render-Pfad ist, den dieser Task neu einführt.

## tmux-Smoke (120 UND 80 Spalten)

`bin/bt` in diesem Repo starten, `ctrl+k` → „tags" tippen → enter → Page
öffnet, Liste zeigt aktuell (noch leere Registry, D02: keine
`.beans-tags.yml` in diesem Repo) NUR die undefinierten/in-Verwendung-Tags
(z. B. `to-review`, falls ein Bean das Tag trägt). `esc` → zurück zu
Browse. Wiederholen bei 80 Spalten — Zeilen dürfen nicht umbrechen/
abschneiden ohne `…`-Truncation. Danach `git status --porcelain` leer
(kein `.beans-tags.yml` durch den bloßen Page-Besuch angelegt — T2 ist
rein lesend).

## Akzeptanz-Checkliste

- [x] Page über Palette „go to tags" erreichbar
- [x] Liste zeigt Union definiert+frei (D09-Sortierung)
- [x] Verwendungszähler korrekt
- [x] `enter` dokumentierter No-Op
- [x] `esc` → Browse
- [x] `handleKey`-Leak-Test (D06-Regression) grün
- [x] GlobalHint im Header leer (D07)
- [x] kein neuer `.beans-tags.yml` durch bloßes Öffnen
- [x] Goldens Gegenbeleg grün
- [x] tmux-Smoke 120+80 belegt
- [x] voller Lauf grün
- [x] Commit `feat(tui): E10 Tag-Management-Page — Grundgerüst`

## PRELUDE aus T1-Review (2026-07-16, F03, low)

Als erster eigener Commit (test-only): `TestLoadTagDefsSkipsInvalidNamesDefensively` (internal/data/tagdefs_test.go:25-34) — der dritte YAML-Eintrag (leerer Scalar) wird von yaml.v3 schon VOR dem Unmarshal verworfen (empirisch len(f.Tags)==2, nicht 3); der Test beweist damit nur die Bad_Tag-Filterung, nicht den Leer-String-Fall. Test um einen echten Leer-String ergänzen, der bis zum Filter durchreicht (z.B. YAML-Zeile `- ""`).

DONE (2026-07-16): erledigt als eigener Commit `test(data): Leer-String-Fall Registry-Filter` (44 Zeichen, VOR dem T2-Feature-Commit) — empirisch verifiziert (throwaway yaml.v3-Probe): der bare-blank-Scalar (`  - `) ergab `len(f.Tags)==2`, der explizite `- ""` ergibt `len(f.Tags)==3` mit `tags[2]==""`, der bestehende `ValidTagName`-Filter in `LoadTagDefs` schneidet diesen Fall bereits korrekt heraus (kein Produktionscode geändert).

## Summary

`internal/tui/view_tag_management.go` (neu) liefert das read-only
Grundgerüst der Tag-Management-Page (D05-D09): `tagRegistryRow`/
`tagRegistryRows` (D09-Union-Algebra: definierte Tags zuerst alpha-sortiert,
danach freie/in-Verwendung-Tags Count-absteigend+alpha-Tie-Break, ein Tag
das BEIDES ist erscheint genau einmal mit seinem echten Count),
`openTagManagementPage` (D02/D03: frisches, tolerant-fehlendes
`LoadTagDefs()` bei jedem Page-Open, nil-`client`-sicher für Test-Fixtures),
`tagManagementChrome`/`tagManagementLocalBindings` (D07: GlobalHint leer),
`viewTagManagement` (D08: Einzel-Pane, PF-12-Marker-Spalte + rechtsbündiger
Count via `pickerRowFill`), `keyTagManagement` (D06: Up/Down/Enter-No-Op/
Esc). Wired: `types.go` (`viewTagManagement`-viewID + `tagMgmtRows`/
`tagMgmtCursor`-Felder), `update.go` (`handleKey`-Full-Capture-Checkpoint
EXAKT an Lobbys Stelle), `view_browse_repo.go` (`View()`-Dispatcher-Case),
`overlay_palette.go` (`"go_tags"`-Action, direkt vor `"settings"`).

**B01 (gefunden im eigenen tmux-Smoke, gefixt, Regressionstest ergänzt):**
`viewTagManagement()` übergab `renderPane`/`tagManagementRows` zunächst
`innerW` statt `innerW-2` als Pane-Content-Breite — `renderPane`s `w`-Param
ist eine CONTENT-Breite, die eigene `RoundedBorder` addiert +2 weitere
Spalten (exakt dieselbe Buchhaltung wie `masterDetailWidths`s
`lw+rw+4==innerW` für die zwei Split-Panes, hier für EIN Pane also
`innerW-2`). Ungefixt erzeugte das eine zu breite Content-Zeile, die
`outerBorder`s äußeres `.Width()` still in ZWEI Zeilen umbrach (dieselbe
Bug-Klasse, die `view_browse_repo.go`s eigenes F01-Vollbild `paneW :=
innerW - 2` bereits dokumentiert) — sichtbar als aufgebrochene
Rahmen-Zeile + eine ~70 statt 40 Zeilen lange Ausgabe. Fix: `paneW :=
innerW - 2`, an `renderPane` UND `tagManagementRows` durchgereicht.
Regressionstest `TestViewTagManagementLinesFitExactlyNoWrap` (120+80
Spalten) pinnt seitdem „jede Zeile exakt `m.width` breit, exakt `m.height`
Zeilen" — per RED/GREEN-Probe verifiziert (Fix temporär zurückgedreht,
Test schlug mit „73 statt 40 Zeilen" fehl, dann wiederhergestellt).

## Test-Output (RED→GREEN wörtlich)

RED (Implementierung testweise entfernt, `command go test ./internal/tui/
-run "TagRegistryRows|TagManagement|GoTags|FiveGoToEntries" -v -count=1`):

```
# beans-tui/internal/tui [beans-tui/internal/tui.test]
internal/tui/types.go:587:18: undefined: tagRegistryRow
internal/tui/overlay_palette.go:259:13: m.openTagManagementPage undefined (type model has no field or method openTagManagementPage)
internal/tui/update.go:895:12: m.keyTagManagement undefined (type model has no field or method keyTagManagement)
internal/tui/view_browse_repo.go:712:11: m.viewTagManagement undefined (type model has no field or method viewTagManagement)
internal/tui/view_tag_management_test.go:28:10: undefined: tagRegistryRows
internal/tui/view_tag_management_test.go:43:10: undefined: tagRegistryRows
internal/tui/view_tag_management_test.go:59:10: undefined: tagRegistryRows
internal/tui/view_tag_management_test.go:74:10: undefined: tagRegistryRows
internal/tui/view_tag_management_test.go:84:10: undefined: tagRegistryRows
internal/tui/view_tag_management_test.go:94:10: undefined: tagRegistryRows
internal/tui/view_tag_management_test.go:94:10: too many errors
FAIL	beans-tui/internal/tui [build failed]
FAIL
```

GREEN (Implementierung wiederhergestellt, gleiches `-run`-Filter plus die
D06-Bestandstests, 24/24 grün):

```
=== RUN   TestPaletteActionsNoFocusedBeanOmitsNodeActions
--- PASS: TestPaletteActionsNoFocusedBeanOmitsNodeActions (0.00s)
=== RUN   TestPalFilteredActionsFuzzyGoMatchesAllFiveGoToEntries
--- PASS: TestPalFilteredActionsFuzzyGoMatchesAllFiveGoToEntries (0.00s)
=== RUN   TestTagRegistryRowsDefinedFirstAlphaThenFreeByCountDesc
--- PASS: TestTagRegistryRowsDefinedFirstAlphaThenFreeByCountDesc (0.00s)
=== RUN   TestTagRegistryRowsIncludesUnusedDefinedTagWithZeroCount
--- PASS: TestTagRegistryRowsIncludesUnusedDefinedTagWithZeroCount (0.00s)
=== RUN   TestTagRegistryRowsDefinedTagKeepsRealUsageCount
--- PASS: TestTagRegistryRowsDefinedTagKeepsRealUsageCount (0.00s)
=== RUN   TestTagRegistryRowsDedupesDuplicateDefNames
--- PASS: TestTagRegistryRowsDedupesDuplicateDefNames (0.00s)
=== RUN   TestTagRegistryRowsEmptyEverywhereReturnsEmpty
--- PASS: TestTagRegistryRowsEmptyEverywhereReturnsEmpty (0.00s)
=== RUN   TestTagRegistryRowsNilIndexNoPanic
--- PASS: TestTagRegistryRowsNilIndexNoPanic (0.00s)
=== RUN   TestKeyTagManagementEscReturnsToBrowse
--- PASS: TestKeyTagManagementEscReturnsToBrowse (0.00s)
=== RUN   TestHandleKeyOnTagManagementViewDoesNotLeakToNodeAction
--- PASS: TestHandleKeyOnTagManagementViewDoesNotLeakToNodeAction (0.00s)
=== RUN   TestHandleKeyOnTagManagementViewDoesNotLeakHelpOrPalette
--- PASS: TestHandleKeyOnTagManagementViewDoesNotLeakHelpOrPalette (0.00s)
=== RUN   TestKeyTagManagementUpDownMovesCursor
--- PASS: TestKeyTagManagementUpDownMovesCursor (0.00s)
=== RUN   TestKeyTagManagementEnterIsHandledNoOp
--- PASS: TestKeyTagManagementEnterIsHandledNoOp (0.00s)
=== RUN   TestOpenTagManagementPageNilClientBuildsFromIdxOnly
--- PASS: TestOpenTagManagementPageNilClientBuildsFromIdxOnly (0.00s)
=== RUN   TestOpenTagManagementPageResetsStaleCursor
--- PASS: TestOpenTagManagementPageResetsStaleCursor (0.00s)
=== RUN   TestTagManagementChromeGlobalHintEmpty
--- PASS: TestTagManagementChromeGlobalHintEmpty (0.00s)
=== RUN   TestTagManagementChromeFooterListsUpDownBack
--- PASS: TestTagManagementChromeFooterListsUpDownBack (0.00s)
=== RUN   TestViewTagManagementRendersDefinedAndFreeRows
--- PASS: TestViewTagManagementRendersDefinedAndFreeRows (0.00s)
=== RUN   TestViewTagManagementNoNewRegistryFileOnRender
--- PASS: TestViewTagManagementNoNewRegistryFileOnRender (0.00s)
=== RUN   TestViewDispatcherRoutesTagManagementView
--- PASS: TestViewDispatcherRoutesTagManagementView (0.00s)
=== RUN   TestPaletteActionsIncludesGoToTagsBeforeSettings
--- PASS: TestPaletteActionsIncludesGoToTagsBeforeSettings (0.00s)
=== RUN   TestDispatchPaletteGoTagsOpensTagManagementPage
--- PASS: TestDispatchPaletteGoTagsOpensTagManagementPage (0.00s)
PASS
ok  	beans-tui/internal/tui	0.353s
```

B01-Regressionstest eigenes RED/GREEN (separat, `paneW := innerW` testweise
zurückgedreht → `command go test ./internal/tui/ -run
TestViewTagManagementLinesFitExactlyNoWrap -v`):

```
RED: view_tag_management_test.go:314: width=120: viewTagManagement() produced 73 lines, want exactly 40 (height)
--- FAIL: TestViewTagManagementLinesFitExactlyNoWrap (0.00s)
GREEN (nach Wiederherstellung von `paneW := innerW - 2`):
--- PASS: TestViewTagManagementLinesFitExactlyNoWrap (0.00s)
```

Gates: `command gofmt -l .` leer · `command go vet ./...` leer ·
`command go test ./... -short -count=1` grün (alle Pakete ok) ·
`command go test ./... -count=1` grün (143s Voll-Lauf, `internal/tui`
138.7s, exit 0).

## Golden-Gegenbeleg

`command go build -o bin/bt .` erfolgreich, danach `command go test
./internal/tui/ -run "TestTreeGolden|TestBacklogGolden|TestChromeGolden"
-count=2 -v` → alle 5 Testfunktionen (inkl. `*Deterministic`-Varianten)
PASS, beide Wiederholungen. `git diff --stat -- internal/tui/testdata/`
leer — Basis-Goldens unverändert. Keine eigene Golden-Suite für
`viewTagManagement()` angelegt (Implementer-Entscheidung, mirrort E9
T7/T8s Freiheit) — stattdessen mehrere Render-/verhaltens-geerdete Tests
(`TestViewTagManagementRendersDefinedAndFreeRows`,
`TestViewDispatcherRoutesTagManagementView`,
`TestViewTagManagementLinesFitExactlyNoWrap`).

## Smoke

tmux, `TERM=xterm-256color`, Breite 120 UND 80 (Sessions `bt120r92i`/
`bt80r92i`, `bin/bt` in diesem Repo): `ctrl+k` → „tags" tippen → Cursor auf
„go to tags" → `enter` → Page öffnet. Registry ist leer (kein
`.beans-tags.yml` in diesem Repo, D02) — Liste zeigt NUR die 3 freien,
tatsächlich verwendeten Tags dieses Repos, korrekt D09-sortiert
(Count absteigend, Alpha-Tie-Break):

```
▌  to-review                                                                                                    10  ││
   rejected                                                                                                      1  ││
   smoke                                                                                                         1  ││
```

Navigation (↓↓) bewegt den `▌`-Cursor sichtbar durch die Liste (verifiziert
bei 120 UND 80 Spalten). Full-Capture-Beleg (120 Spalten): `d`, `s`, `t`,
`a`, `r`, `c`, `e` NACHEINANDER gedrückt — Bildschirm danach byte-identisch
zum Zustand davor (keine Overlays, kein Formular, kein Delete-Confirm,
Cursor unverändert auf `smoke`); zusätzlich `?` (kein Help-Overlay) und
`ctrl+k` (kein Command-Center) geprüft — beide ebenfalls wirkungslos
(mirrort Lobbys eigenen Full-Capture-Präzedenzfall). `enter` auf einer
Zeile: dokumentierter No-Op (Bildschirm unverändert). `esc` → zurück zu
Browse (Breadcrumb wechselt zurück zu „Browse", GlobalHint mit den 4
globalen Bindings erscheint wieder). Wiederholt bei 80 Spalten (Session
`bt80r92i`) — identisches Verhalten, keine Zeile umgebrochen/abgeschnitten
(Tag-Namen `to-review`/`rejected`/`smoke` sind kurz genug, um bei 80
Spalten unverändert zu passen — Truncation-Pfad selbst nicht live
provoziert, aber per `truncate(r.name, nameW)`-Code-Pfad UND dem breiteren
`TestViewTagManagementLinesFitExactlyNoWrap`-Regressionstest mit einem
längeren, synthetischen Tag-Namen bei width=80 abgedeckt). Beide Sessions
sauber beendet (`q`→`enter`, kein `bt`-Prozess mehr aktiv). `git status
--porcelain` nach BEIDEN Läufen: keine `.beans-tags.yml` (D02/T2-ist-rein-
lesend bestätigt), nur die erwarteten Quelldateiänderungen.

## Deviations/ERRATA

- **B01** (siehe Summary oben): `paneW := innerW - 2` statt `innerW` —
  gefunden+gefixt+regressionsgetestet im eigenen tmux-Smoke, kein
  PO-sichtbares Verhalten betroffen (Fix vor jedem Commit).
- Palette-Bestandstests angepasst (nicht im Bean-Body vorgegeben, aber
  durch die neue Action zwingend): `TestPaletteActionsNoFocusedBeanOmitsNodeActions`
  (`wantGlobal` um `"go_tags"` ergänzt) und
  `TestPalFilteredActionsFuzzyGoMatchesAllFourGoToEntries` →
  `...AllFiveGoToEntries` umbenannt (die neue „go to tags"-Aktion matcht
  denselben `"go"`-Fuzzy-Query wie die bestehenden 4 „go to …"-Einträge).
- Zusätzlich zu den 4 im Bean-Body zitierten RED-Tests wurden 17 weitere
  Tests ergänzt (Marker-Dedupe/Nil-Index/Leer-Fall für `tagRegistryRows`,
  Up/Down/Enter-No-Op für `keyTagManagement`, D06-Leak-Erweiterung auf
  `?`/`ctrl+k`, `openTagManagementPage`-Nil-Client/Cursor-Reset,
  `tagManagementChrome`-GlobalHint/Footer, Render-/Dispatcher-/
  Palette-Wiring-Tests, der B01-Regressionstest) — reine Ergänzung, keine
  Abweichung vom Spezifizierten.
- Marker-Glyph (D10-Konvention vorgezogen, PF-12) als Implementer-
  Entscheidung festgelegt: `✓` (U+2713, EAW=Neutral, Green eingefärbt) für
  definierte Tags, gleich breiter Leerraum für freie — kein PO-Wortlaut
  hat einen konkreten Glyph vorgegeben; T6 (bt-pqq3) kann diesen Glyph bei
  Bedarf wiederverwenden oder bewusst überschreiben.

## Notes for T3+T4

- `tagManagementLocalBindings()` (`view_tag_management.go`) ist EIN
  gemeinsamer Funktionsrumpf `{keys.Up, keys.Down, keys.Back}` — T3
  (Create) hängt sein eigenes Binding (`keys.NewTag`-Analogon oder ein
  neues) HIER an, nicht in einer zweiten Liste. T4 (Delete) analog.
- `keyTagManagement` hat aktuell GENAU 3 Cases (`up`/`down` via `navKey`,
  `keys.Back`, `keys.Enter`-No-Op) plus den finalen Catch-All-`return m,
  nil`. T3/T4/T5 fügen ihre eigenen `case`s VOR diesem Catch-All ein (z.B.
  `keys.Create`/`keys.Delete`/eine Rename-Taste) — der Catch-All bleibt
  der D06-Schutz gegen jedes nicht explizit behandelte Zeichen.
- `m.tagMgmtRows []tagRegistryRow` + `m.tagMgmtCursor listState` (types.go)
  sind die geteilten Felder — T3/T4/T5 lesen/schreiben dieselben, KEINE
  neuen Parallel-Felder (z.B. für einen Create-/Delete-/Rename-Zielnamen:
  `mutTarget` (bereits vorhanden, E3-Konvention) reicht, mirrort jeden
  anderen Node-Action-Picker in dieser Codebase).
- `tagRegistryRows(idx, defs)` ist eine REINE Funktion (kein `model`-
  Receiver) — T3/T4/T5 rufen sie nach jeder lokalen Registry-Mutation
  erneut auf, um `tagMgmtRows` neu aufzubauen (kein manuelles
  Row-Patchen), mirrort wie `openTagManagementPage` sie selbst nutzt.
- D10-ERRATUM-Notiz für T6 (bt-pqq3, siehe `tagRegistryRows`s eigener
  Doc-Kommentar): sobald T6 `collectTagCounts` um einen
  `defined map[string]bool`-Parameter erweitert, KÖNNTE `tagRegistryRows`
  diese erweiterte Fassung direkt aufrufen statt die eigene, separate
  Zähl-Schleife zu pflegen — nicht in T2 vorgezogen (disjunkter
  Datei-Scope, T6 lief parallel, hatte beim T2-Abschluss noch nicht
  gelandet).
- Marker-Glyph `tagManagementMarkerGlyph = "✓"` / `tagManagementMarkerStyle`
  (Green) liegen in `view_tag_management.go` — T6 kann sie direkt
  importieren/wiederverwenden für die Suggest-Mode-Marker-Spalte im
  Tag-Picker, statt einen zweiten Glyph zu erfinden (PF-12-Konsistenz).

## Review-Findings Runde 1 (2026-07-16, T2-Review, Verdict CHANGES_REQUIRED)

- **F01 (medium, view_tag_management_test.go:117-124):** `TestHandleKeyOnTagManagementViewDoesNotLeakToNodeAction` nutzt `newModel(nil, "")` — Model OHNE fokussiertes Bean. keyNodeActions Zweig (update.go:645-648) ist bei focusedBean()==nil ohnehin ein silent No-Op → der Test bleibt auch GRÜN, wenn der D06-Guard (update.go:894-896) komplett entfernt wird (Reviewer-Mutation belegt; mit echtem Bean öffnet `d` dann overlay=6). Produktionscode ist korrekt — nur der Test beweist es nicht. FIX: Test auf `fixtureModel(t, fixtureBeans())` mit gesetztem Cursor/cursorID auf ein echtes Bean umstellen (wie die ?/ctrl+k-Companion-Tests), sodass Guard-Entfernung den Test rot macht.
- **F02 (low, mouse.go:158-159):** Full-Capture-Guard in handleMouse listet `m.view == viewTagManagement` NICHT, obwohl der Doc-Kommentar darüber Vollständigkeit beansprucht. Aktuell folgenlos (switch ohne default schützt indirekt), aber latente Falle für T3-T6. FIX: View in die OR-Kette aufnehmen (Defense-in-Depth) + falls sinnvoll Testabdeckung.

Fix-Runde beim selben Implementer; Re-Review beim selben Reviewer. Nach Fix: Verifikation, dass die Reviewer-Mutation (D06-Guard raus) den neuen Test ROT macht — als RED-Beleg zitieren.

## Fix-Runde 1 (2026-07-16, selber Implementer)

Commit `test(tui): T2-Review R1 — F01 Leak-Test, F02 Maus` (Refs: bt-r92i).

**F01 (medium) — GEFIXT.** `TestHandleKeyOnTagManagementViewDoesNotLeakToNodeAction`
(view_tag_management_test.go) umgestellt auf `fixtureModel(t, fixtureBeans())`
+ `focusBean(m, "tk-2")` (Cursor/cursorID/expanded auf echtes Bean, Helper aus
box_menu_value_test.go) + `m.view = viewTagManagement`. Zusätzlich ein
Setup-Sanity-Guard im Test (`m.focusedBean() == nil` → Fatal), damit der Test
nie wieder still zur No-Op-Variante degradieren kann, plus eine zweite
Assertion (`view` bleibt `viewTagManagement`). Doc-Kommentar dokumentiert die
F01-Historie (warum die Bean-Body-Sketch-Variante beweislos war).

RED-Beleg (Reviewer-Mutation exakt nachgestellt: D06-Guard update.go:894-896
auskommentiert, `command go test ./internal/tui/ -run
TestHandleKeyOnTagManagementViewDoesNotLeakToNodeAction -v -count=1`):

```
=== RUN   TestHandleKeyOnTagManagementViewDoesNotLeakToNodeAction
    view_tag_management_test.go:145: want no overlay opened, got 6
--- FAIL: TestHandleKeyOnTagManagementViewDoesNotLeakToNodeAction (0.00s)
FAIL
FAIL	beans-tui/internal/tui	0.439s
```

(`overlay=6` == `overlayDeleteConfirm` — exakt der vom Reviewer belegte Leak.)
Guard danach wiederhergestellt, `git diff internal/tui/update.go` gegen HEAD
leer (byte-identisch), Test GREEN:

```
=== RUN   TestHandleKeyOnTagManagementViewDoesNotLeakToNodeAction
--- PASS: TestHandleKeyOnTagManagementViewDoesNotLeakToNodeAction (0.00s)
```

**F02 (low) — GEFIXT.** `m.view == viewTagManagement` in die
Full-Capture-OR-Kette in `handleMouse` (mouse.go) aufgenommen, mit
Doc-Kommentar (Defense-in-Depth: heute wirkungsgleich, da wheelMove- und
Left-Click-Switch keinen viewTagManagement-Case und keinen default haben —
aber die Vollständigkeits-Zusage des Kommentars gilt wieder, und künftige
T3-T6-Änderungen an einem der Switches können keine Streu-Klicks mehr gegen
die fremde Page-Geometrie routen). Testabdeckung ergänzt:
`TestHandleMouseIgnoredOnTagManagementPage` (mouse_test.go, spiegelt
`TestHandleMouseIgnoredWhenFullscreenActive`s Muster: Wheel + Left-Click auf
der offenen Page → Tree-Cursor, tagMgmtCursor, view, overlay alle
unverändert). Ehrliche Scope-Notiz im Test-Doc-Kommentar: der Test pinnt den
VERTRAG (Maus = No-Op auf der Page), nicht die konkrete Guard-Zeile — eine
Entfernung NUR des F02-Guards macht ihn heute nicht rot (die Switch-ohne-
default-Struktur schützt indirekt weiter), er schlägt erst an, wenn eine
künftige Änderung Maus-Events tatsächlich gegen die Page routet.

Gates Fix-Runde 1: `command gofmt -l .` leer · `command go vet ./...` leer ·
voller Lauf `command go test ./... -count=1` grün (145s, internal/tui 140.1s,
exit 0) · Goldens-Gegenbeleg `-run "TestTreeGolden|TestBacklogGolden|
TestChromeGolden" -count=2` grün, `git diff --stat -- internal/tui/testdata/`
leer.
