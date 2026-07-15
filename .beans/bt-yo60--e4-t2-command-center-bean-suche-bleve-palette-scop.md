---
# bt-yo60
title: 'E4 T2 — Command-Center: Bean-Suche (Bleve, palette-scoped)'
status: completed
type: task
priority: normal
created_at: 2026-07-15T05:40:07Z
updated_at: 2026-07-15T06:38:05Z
parent: bt-tfqi
blocked_by:
    - bt-jpgn
---

Ziel: zweiter Kandidaten-Pool im Command-Center (`ctrl+k`/`K`, design-spec.md
§6 V5) -- sobald `m.palQuery` nicht leer ist, mischen sich Bean-Treffer UNTER
die Aktionen (design decision b: Aktionen IMMER zuerst, gedeckelt auf 20,
kanonisch sortiert). Reused `beanMatchesSearch`s <3-Zeichen/Bleve-ab-3-
Bifurkation strukturell, aber auf palette-EIGENEN Bleve-Feldern
(palBleveIDs/palBleveFor/palBleveLoading, eigener paletteBleveResultMsg) --
NIE die Tree-/Backlog-Suchfelder wiederverwendet.

Plan: docs/plans/v1-port/epic-E4-plan.md »Task 2«. Blocked-by bt-jpgn (T1,
completed) -- Korrektur I01 aus bt-jpgn gelesen: `paletteItem` hatte VOR
diesem Task nur 3 Felder (kind/actionID/label), kein `bean`-Feld -- hier neu
hinzugefügt, nicht bloß befüllt.

## Akzeptanz

- [x] internal/tui/overlay_palette.go: `paletteItem.bean` (neues Feld,
      I01-Korrektur berücksichtigt), `paletteBeanResultCap` (20),
      `palBeanMatches` (Bifurkations-Port von `beanMatchesSearch`, auf
      `palBleveFor`/`palBleveIDs`), `palFilteredBeans` (Cap + kanonische
      Sortierung via `data.SortBeans`), `palFiltered` erweitert (Beans NACH
      Aktionen angehängt), `keyPalette`s Rune-/Backspace-Zweige um
      `dispatchPaletteBleveIfDue()` (neuer Bleve-Dispatch-Tail) ergänzt,
      `dispatchPalette`s `paletteKindBean`-Case gefüllt (Cursor-Jump +
      `expandAncestorsOf`), `paletteBox` um Bean-Zeilen (D08
      ansi.Strip+Accent-Stil, Separator NUR wenn beide Pools nicht-leer)
      erweitert
- [x] internal/tui/types.go: `palBleveIDs`/`palBleveFor`/`palBleveLoading`
      (palette-EIGENE Kopien, NICHT `searchBleveIDs`/... wiederverwendet)
- [x] internal/tui/messages.go: `paletteBleveResultMsg`/`paletteSearchCmd`
      (eigener Msg-Typ, strukturell identisch zu `searchBleveResultMsg`/
      `searchCmd`, aber getrennt)
- [x] internal/tui/update.go: neuer `Update()`-Case für
      `paletteBleveResultMsg` -> `applyPaletteBleveResult` (Staleness-Guard
      gegen `m.palQuery`, Analog `applyBleveResult`), `maybePaletteBleveCmd`
      (Analog `maybeBleveCmd`, Threshold-Check gegen `palBleveFor`)
- [x] internal/tui/overlay_palette_test.go erweitert: 8 Pflicht-Tests aus dem
      Plan-Pseudocode (Step 1) 1:1 übernommen, plus 2 Zusatztests
      (`TestApplyPaletteBleveResultAppliedWhenQueryStillCurrent`,
      `TestPaletteSearchCmdTagsResultWithQuery` -- positive Gegenstücke zu
      den bestehenden `/`-Bleve-Tests in search_test.go, gleiche
      Test-Hygiene)
- [x] T1-Contract `TestPalFilteredEmptyQueryReturnsAllActionsNoBeans`
      UNVERÄNDERT grün -- nicht aufgeweicht (per Auftrag verifiziert, s.
      Verifikation)
- [x] `command go test ./...` (2x, ohne -short), `-race`, `gofmt -l .`,
      `go vet ./...`, Goldens (Chrome/Tree/TreeDeterministic/Backlog/
      BacklogDeterministic) 2x -- alle grün
- [x] `command go build -o bin/bt .` ok
- [x] tmux-Smoke im frischen Scratch-Repo (Details unten)

## Summary

`palBeanMatches` portiert `beanMatchesSearch`s (view_browse_repo.go)
Bifurkation strukturell 1:1, aber gegen `m.palBleveFor`/`m.palBleveIDs`
statt `m.searchBleveFor`/`m.searchBleveIDs`: unter 3 Zeichen, oder solange
kein Bleve-Treffer für die AKTUELLE Palette-Query vorliegt, lokaler
Title+ID-Substring (case-insensitive); ab 3 Zeichen, sobald `palBleveFor ==
palQuery`, wird Bleve authoritativ -- UNIONed mit dem lokalen ID-Substring
(I01-Precedent aus E2 Task 3, wörtlich übernommen: Bleve indiziert Title+
Body, nicht zwingend eine ID-Teilzeichenkette). `palFilteredBeans` iteriert
`m.idx.ByID` (ALLE Beans, unabhängig vom Tree-Expand-Zustand -- ein
kollabierter Epic verbirgt seine Kinder nur im Tree, nicht im Index),
sortiert kanonisch (`data.SortBeans`, I03) und deckelt auf
`paletteBeanResultCap = 20`.

`palFiltered` hängt `palFilteredBeans(m)` NACH der bestehenden
Aktionen-Schleife an (design decision b: Aktionen immer zuerst, keine
Score-Verzahnung) -- der kombinierte `m.palList.cursor`-Indexraum bleibt EIN
Feld über beide Pools, wie von T1 bereits vorgesehen.

`keyPalette`s Rune-/Backspace-Zweige rufen jetzt zusätzlich
`dispatchPaletteBleveIfDue()` (neuer Tail, analog `update.go`s
`dispatchBleveIfDue`, aber ohne zusätzlichen Textinput-Cmd zu batchen -- die
Palette führt `palQuery` als reinen String, kein `bubbles/textinput.Model`).
`maybePaletteBleveCmd`/`applyPaletteBleveResult` sind wörtliche Kopien von
`maybeBleveCmd`/`applyBleveResult`, nur auf die Palette-Felder gespiegelt --
komplett eigenständiger Staleness-Guard, keine Cross-Talk-Möglichkeit mit
einer aktiven `/`-Session.

`dispatchPalette`s `paletteKindBean`-Case: `expandAncestorsOf(m.idx,
m.expanded, it.bean.ID)` (zyklen-geschützt, cloneBoolMap-CoW) + `cursorID =
it.bean.ID` + `view = viewBrowseRepo` -- exakt der gleiche 3-Zeilen-
Aufruf wie `keyDetailFocus`s bestehender Relation-Jump (update.go:664-672),
funktioniert daher auch aus der Backlog-View heraus (Test pinnt das
explizit).

`paletteBox`: die kombinierte Item-Liste wird am ersten `paletteKindBean`-
Index gesplittet (`split`), Aktionen und Beans laufen durch ZWEI separate
`menuList`-Aufrufe (Bean-Aufruf mit `m.palList.cursor - split` als eigener
Cursor-Offset) -- dadurch behält jeder Pool seinen eigenen Zeilenstil
(Aktionen: `theme.Header.Render` bei Selektion, Beans: D08
`ansi.Strip`+`theme.Accent.Render`, da `relationRow`-Output bereits
themed ist, box_picker_parent.go-Präzedenz) UNTER Beibehaltung EINES
gemeinsamen Cursor-Indexraums über `m.palFiltered()`. Ein
`theme.Dim`-Separator trennt beide Blöcke nur, wenn BEIDE nicht-leer sind.

## Verifikation

`command go test ./internal/tui/ -run 'TestPal...'` (alle neuen + alle
bestehenden T1-Palette-Tests) grün, inkl. explizit erneut geprüftem
T1-Contract `TestPalFilteredEmptyQueryReturnsAllActionsNoBeans`.
`command go test ./... -timeout 600s` 2x grün (Lauf 2 aus dem Go-Test-Cache,
gleiches Vorgehen wie bt-jpgns eigene Verifikation: der `-race`-Lauf gilt
als der real erneut ausgeführte zweite Durchlauf). `command go test ./...
-race -timeout 600s` grün (~125s, tui-Paket). `command gofmt -l .` leer.
`command go vet ./...` leer. Goldens (TestChromeGolden/TestTreeGolden/
TestTreeGoldenDeterministic/TestBacklogGolden/
TestBacklogGoldenDeterministic) 2x mit `-count=1` (erzwungen, nicht
gecached) grün -- Bean-Suche berührt keinen Default-Render-Pfad.
`command go build -o bin/bt .` ok.

## Smoke (tmux, frischer Scratch-Repo `bt-smoke-e4t2`: 1 Epic "Payments
## Overhaul" + 2 Kinder-Tasks ("Refactor Checkout Flow",
## "Add Retry Logic To Webhook") + 1 unrelated Task)

1. `ctrl+k` auf dem fokussierten Epic -> Palette öffnet mit den 6
   Node-Aktionen zuerst (T1-Verhalten unverändert).
2. Query "checkout" (voller Token, >=3 Zeichen) getippt -> Bean-Treffer
   "Refactor Checkout Flow" erscheint UNTER den (hier leeren, da "checkout"
   fuzzy gegen keine Aktion matcht) Aktionen, D08-Akzent-Stil, Cursor
   automatisch auf dem einzigen Treffer -> `enter` -> Palette schließt, Tree-
   Cursor springt auf "Refactor Checkout Flow", Epic "Payments Overhaul"
   automatisch expandiert (▾ statt ▸) -- Cursor+Ancestor-Expand-Kontrakt
   bestätigt.
3. Lokaler Sofort-Match vs. async Bleve-Merge explizit gezeigt: Query "we"
   (2 Zeichen, unter Bleve-Threshold) -> lokaler Title-Substring matcht
   sofort "Add Retry Logic To Webhook". 3. Zeichen "b" getippt (Query "web",
   erreicht Threshold, `paletteSearchCmd` dispatcht) -> UNMITTELBAR danach
   (vor Settle) zeigt die Palette den Treffer NOCH (lokaler Fallback, da
   `palBleveFor != palQuery`); nach Settle (Bleve-Antwort gelandet)
   verschwindet der Treffer -- `beans list --search web` liefert `[]`, da
   Bleve "web" als EXACT-TERM-Match behandelt (kein Prefix/Wildcard ohne
   `*`/`~`) und "webhook" ein eigenes Token ist. Direkt gegen die beans-CLI
   verifiziert (`beans list --json --search web` -> `[]`, `--search
   webhook`/`--search "web*"` -> Treffer) -- KEIN Bug, sondern dieselbe
   Exact-Term-Bleve-Semantik, die `/`s bestehender `beanMatchesSearch` (E2
   Task 3) bereits hat; hier zum ersten Mal über die Palette sichtbar
   gemacht, s. Deviations.

## Deviations

- **Keine Plan-Abweichung.** Alle 8 im Plan vorgeschriebenen Tests (Step 1)
  1:1 implementiert, plus 2 Zusatztests (positive Bleve-Apply/Cmd-Tag-
  Gegenstücke, gleiche Hygiene wie search_test.go).
- **Beobachtung (kein Bug, dokumentiert für spätere Sessions):** Bleve-Such-
  Queries ohne `*`/`~` sind EXACT-TERM-Matches gegen den Bleve-Tokenizer --
  ein Teilwort wie "web" matcht das Token "webhook" NICHT (anders als der
  lokale <3-Zeichen-Substring-Fallback, der IMMER Substring-basiert bleibt).
  Das ist bestehende `data.Client.Search`-Semantik (E2 Task 3, unverändert),
  hier im Smoke zum ersten Mal für die Palette demonstriert -- kein neues
  Verhalten, kein Fix nötig, aber notierenswert für PO-Erwartungsmanagement
  (Bean-Suche in der Palette ab 3 Zeichen ist "Bleve-Modus", kein Fuzzy-
  Substring-Modus, design-spec §6 V2 wörtlich).

## Notes für T3 (Review-Cockpit, bean bt-hxyo)

Die Capture-Order-Kette (design decision h) steht bereits vollständig für
den dritten Block bereit: `m.paletteOpen` -> `keyPalette` und
`keys.Palette` -> `openPalette` sitzen in `update.go`s `handleKey` VOR der
Stelle, an der T3 seinen `m.view == viewReviewCockpit` -> `keyReviewCockpit`
Capture-Block einfügt (direkt darunter, wie im Plan-Kommentar Zeile
201-216 vorgeschrieben) -- unverändert seit T1, von T2 nicht berührt.
`dispatchPalette`s `paletteKindBean`-Case (Cursor-Jump + `expandAncestorsOf`)
ist ein wiederverwendbares Muster, falls T4s künftige "go to: review
cockpit"-Aktion einen ähnlichen Jump braucht (dort aber vermutlich ohne
Tree-Expand, da die Cockpit ihre eigene Cursor-Repräsentation
`reviewCursor` hat, design decision i -- kein Blindübernahme-Kandidat, nur
strukturelle Analogie). Keine offenen Fragen/Blocker für T3.

Commit: siehe `Refs: bt-yo60` (feat(tui)).
