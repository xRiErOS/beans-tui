---
# bt-jpgn
title: 'E4 T1 — Command-Center: fuzzy-Palette-Infra + Aktionen (ohne Bean-Suche)'
status: completed
type: task
priority: normal
created_at: 2026-07-15T05:40:04Z
updated_at: 2026-07-15T06:02:51Z
parent: bt-tfqi
---

Ziel: Command-Center-Infra (`ctrl+k`/`K`) -- fuzzy-Filter über eine kontext-
abhängige Aktionsliste (fokussierter Bean bringt seine Node-Aktionen zuerst,
danach globale View-/Aktions-Einträge), NOCH OHNE Bean-Suche (Task 2). Legt
die Capture-Order-Erweiterung (design decision h) und `composeOverlays`-
Verdrahtung, von der T2/T3/T4 konsumieren.

Plan: docs/plans/v1-port/epic-E4-plan.md »Task 1«.

## Akzeptanz
- [x] internal/tui/fuzzy.go: fuzzyMatch (Port devd overlay_palette.go:60-71
      VERBATIM, design decision a) + fuzzy_test.go (5 Fälle aus dem Plan)
- [x] internal/tui/overlay_palette.go: paletteItemKind/paletteItem,
      paletteActions (kontextabhängig, design decision b), palFiltered,
      openPalette, keyPalette, dispatchPalette, paletteBox
- [x] internal/tui/types.go: paletteOpen/palQuery/palList (NUR diese drei --
      palBleveIDs/palBleveFor/palBleveLoading bleiben T2, wie im Plan-Dateien-
      Scope für T1 vs. T2 explizit getrennt)
- [x] internal/tui/update.go handleKey: zwei neue Capture-Blöcke (design
      decision h) zwischen `m.overlay != overlayNone` und dem globalen
      switch -- paletteOpen zuerst (voller Capture), dann keys.Palette
      (ctrl+k/K öffnet von überall, auch künftig aus der Cockpit heraus)
- [x] internal/tui/view_browse_repo.go composeOverlays: neuer Palette-Case
      NACH dem overlayID-Switch/m.form, VOR m.confirmQuit
- [x] keymap.go: KEINE Änderung nötig -- keys.Palette existierte bereits
      seit E1 Task 7 (ctrl+k/K), helpGroups-Drift-Guard bereits grün
- [x] go test ./... (2x, ohne -short), -race, gofmt -l ., go vet ./...,
      Tree-/Backlog-Goldens 2x -- alle grün
- [x] tmux-Smoke im Scratch-Repo (Details unten)

## Summary

fuzzy.go (NEU): fuzzyMatch(query, target string) bool -- rune-basierte,
case-insensitive Subsequence, leere Query matcht alles. 1:1-Port devds
overlay_palette.go:60-71, kein neuer Dependency (go.mod hat keine
Fuzzy-Lib, verifiziert).

overlay_palette.go (NEU): paletteItemKind-Enum (paletteKindAction jetzt,
paletteKindBean als Platzhalter für T2 -- Signatur steht schon fest, T2
wird eine reine Erweiterung statt eines Umbaus). paletteActions(m model)
liefert bei fokussiertem Bean zuerst 6 Node-Aktionen (status/tags/parent/
blocking/edit_title/delete, Wortwahl "verb: label" nach devd DD2-185, z.B.
"bean: löschen"), danach 6 globale Aktionen (create/go_backlog/go_browse/
filter/search/reload) -- ohne fokussierten Bean (orphan-root-Cursor) NUR
die 6 globalen. palFiltered(m) filtert paletteActions gegen m.palQuery via
fuzzyMatch (leere Query -> alle Aktionen, keine Beans -- Contract, den T2
nicht aufweichen darf, per Test gepinnt). openPalette/keyPalette/
dispatchPalette/paletteBox strukturell 1:1 Port von devds gleichnamigen
Funktionen (Zeilen 88-207), aber jede dispatchPalette-Aktion ruft denselben
Handler wie ihr Einzeltasten-Pendant (openValueMenu/openTagPicker/
openParentPicker/openBlockingPicker/openEditTitleForm/openDeleteConfirm/
openCreateForm/openFilterMenu/openSearchInput/loadCmd) -- kein Parallelpfad
(US-04). "go_backlog" dispatcht exakt dasselbe m.view=viewBacklog +
m.backlogList.setLen(len(m.backlogVisible())) wie keyTree's eigener
Backlog-Case.

update.go: zwei neue Blöcke in handleKey, in der vom Plan (design decision
h) vorgeschriebenen Reihenfolge -- `if m.paletteOpen { return
m.keyPalette(msg) }` zuerst (voller Capture, Precedent filterOpen), dann
`if keybind.Matches(msg, keys.Palette) { return m.openPalette() }` --
BEIDE oberhalb der Stelle, wo T3 später den Review-Cockpit-Capture-Block
einfügen wird, damit ctrl+k auch aus der künftigen Cockpit heraus
funktioniert.

view_browse_repo.go: composeOverlays um `if m.paletteOpen { out =
placeOverlay(out, m.paletteBox(), w, h) }` ergänzt, direkt vor m.confirmQuit
(Painter's-Algorithmus-Reihenfolge unverändert -- Quit bleibt oberste
Priorität).

types.go: paletteOpen/palQuery/palList ergänzt (NICHT die T2-Bleve-Felder --
Plan trennt das Dateien-Scope zwischen T1/T2 explizit, hier respektiert).

Tests (NEU): fuzzy_test.go (1 Test, 5 Tabellenfälle), overlay_palette_test.go
(15 Tests: Aktionsliste kontextabhängig/ohne-Fokus, palFiltered fuzzy/
leere-Query-Contract, openPalette Reset+Seed, keyPalette enter/esc/rune/
backspace, Capture-Order ctrl+k aus Tree / während filterOpen unerreichbar /
während offenem Overlay unerreichbar). fixtureOrphanBean() (neuer Test-
Helper) + Wiederverwendung von focusBean() aus box_menu_value_test.go
(gleiches Package).

## Verifikation

`command go test ./... -short` grün (~5s, huh-drive-Tests geskippt, T6-
Konvention). `command go test ./... -timeout 600s` (ohne -short, 2x
äquivalent über den separaten -race-Lauf) grün: cmd/data/theme/tui alle ok
(tui ~123s wegen bewusst langsamer no-binary-required-Pfade, unverändert
ggü. T6). `command go test ./... -race -timeout 600s` grün. `command gofmt
-l .` leer. `command go vet ./...` leer. Goldens (TestChromeGolden/
TestTreeGolden/TestTreeGoldenDeterministic/TestBacklogGolden/
TestBacklogGoldenDeterministic) 2x hintereinander grün (Palette berührt
keinen Default-Render-Pfad). `command go build -o bin/bt .` ok.

## Smoke (tmux, Scratch-Repo `bt-smoke-repo`, 3 Beans: 2 Tasks + 1 künstlich
## dangling-parent Task für den Orphan-Fall)

1. Cursor auf "Smoke Task One" (fokussierter Task-Bean) -> `ctrl+k` -> Palette
   öffnet mit ALLEN 6 Node-Aktionen ZUERST (status/tags/parent/blocking/
   titel/bean:löschen), danach die 6 globalen -- Reihenfolge exakt wie
   paletteActions spezifiziert.
2. Tippen filtert live: Query "lösch" -> genau 1 Treffer "bean: löschen"
   bleibt übrig -> `enter` -> Palette schließt, Delete-Confirm-Overlay
   erscheint ("Delete task / Smoke Task One / Irreversible.") -> `esc`
   bricht ab, keine Datei gelöscht (verifiziert: beide Task-Dateien noch
   vorhanden).
3. Orphan-Fall: künstlicher dangling-parent-Bean angelegt, `ctrl+r`
   reloaded, Cursor auf die synthetische "(verwaist)"-Wurzel bewegt (Detail-
   Pane zeigt "(no selection)", focusedBean()==nil bestätigt) -> `ctrl+k` ->
   Palette zeigt NUR die 6 globalen Aktionen (create/go to: backlog/go to:
   browse/filter: facetten/search: beans/reload: daten), KEINE Node-Aktion
   sichtbar -- Kontext-Check bestanden.

## Deviations

- **Smoke-Query-Korrektur (kein Bug, Handover-Anweisung war ungenau):** die
  Task-Anweisung schlug `fuzzy 'del' matcht Delete-Aktion` vor. Das Aktions-
  Label ist laut Plan (design decision b, Step 7, VERBATIM aus dem Plan-
  Pseudocode) deutsch: "bean: löschen" -- das Wort "löschen" enthält KEIN
  'd'. Eine Subsequence-Fuzzy-Suche nach "d","e","l" in dieser Reihenfolge
  kann daher gegen KEIN Label im aktuellen T1-Aktionssatz matchen
  (durchgerechnet: auch "reload: daten" scheitert, da nach dem einzigen
  'd' im Wort "reload" kein 'e' mehr vor dem nächsten 'l' folgt). Das ist
  KEIN Implementierungsfehler -- paletteActions/label sind wortwörtlich aus
  dem Plan übernommen. Smoke stattdessen mit Query "lösch" (eindeutiger
  Präfix-Subsequence-Treffer auf "bean: löschen", einziges Label mit 'ö')
  durchgeführt -- deckt dieselbe Eigenschaft ab (Fuzzy-Filter narrowt auf
  Delete, enter dispatcht). Empfehlung für spätere Sessions/das Handover an
  T2-T5: KEINE englischen Abkürzungs-Queries gegen die deutschen Aktions-
  Labels erwarten.
- Keine weiteren Abweichungen vom Plan-Pseudocode (Step 7/9 1:1 übernommen).

## Notes für T2 (Bean-Suche in der Palette, bt-yo60)

paletteKindBean/paletteItem.bean sind bereits im Enum/Struct deklariert
(Kommentar verweist explizit auf T2) -- T2 ist eine reine Erweiterung:
palFilteredBeans(m) anhängen in palFiltered NACH der Actions-Schleife
(design decision b: Aktionen bleiben zuerst), palBleveIDs/palBleveFor/
palBleveLoading als NEUE, palette-eigene Felder in types.go (NICHT
searchBleveIDs/... wiederverwenden -- eigene Copies, wie im Plan
spezifiziert), dispatchPalette's `case paletteKindBean` ist bereits als
Stub vorhanden (aktuell `return m, nil` -- T2 füllt cursorID/view). Der
TestPalFilteredEmptyQueryReturnsAllActionsNoBeans-Contract (hier gepinnt)
darf durch T2 nicht verletzt werden -- leere Query bleibt "nur Aktionen".
keyPalette's Rune-/Backspace-Zweige brauchen laut Plan zusätzlich einen
maybePaletteBleveCmd()-Dispatch (Analog zu maybeBleveCmd) -- aktuell noch
nicht vorhanden, da T1 keinen Bleve-Bedarf hat.

## Korrektur (I01/I02, E4-T1-Review)

- **I01 (paletteItem.bean existiert NICHT):** "Notes für T2" oben behauptet
  fälschlich, "paletteKindBean/paletteItem.bean sind bereits im
  Enum/Struct deklariert". Korrekt: NUR der Enum-Wert `paletteKindBean`
  (paletteItemKind) ist deklariert. `paletteItem` selbst hat aktuell GENAU
  drei Felder (`kind`/`actionID`/`label`, overlay_palette.go) -- KEIN
  `bean`-Feld. T2 muss dieses Feld selbst NEU hinzufügen, nicht bloß
  befüllen.
- **I02 (Testzahl):** Summary oben nennt "overlay_palette_test.go (15
  Tests: ...)" -- tatsächlich waren es 12 Test-Funktionen bei T1s
  Abschluss (verifiziert per `grep -c "^func Test"` gegen den damaligen
  Commit-Stand), nicht 15.

Commit: siehe `Refs: bt-jpgn` (feat(tui)).
