---
# bt-qbyq
title: 'Detail-Focus: Pfeil-Symmetrie-Fix + esc-Rueckwaerts-Kaskade'
status: completed
type: task
priority: high
created_at: 2026-07-15T21:07:40Z
updated_at: 2026-07-15T22:06:19Z
parent: bt-ntoz
blocked_by:
    - bt-e6q9
---

E8 Task 3 — deckt B01 (Pfeil-links-Fokus-Exit entfernen) und D03 (esc als universelles "eine Ebene zurueck", inkl. Detail-Kaskade + Audit aller esc-Sites) aus bean bt-ntoz. Quelle: design-spec.md §15 PF-16 ("PF-13-Pfeil-Revision"). Ist-Code: internal/tui/update.go (keyDetailFocus). blocked_by bt-e6q9 (T1) -- T1 aendert bereits keyDetailFocus's beanSections()-Aufrufzeile (detailLevel-Parameter), diese Task editiert denselben Funktionskoerper an anderer Stelle (left-case + neuer esc-Handler) -- Reihenfolge vermeidet Merge-Ueberraschungen in derselben Funktion.

## B01 — Pfeil-Symmetrie

keyDetailFocus()s "left"-Case (update.go, ca. Zeile 973-979) HEUTE:
```go
case "left":
    if m.detailLevel == 1 {
        m.detailLevel = 0
    } else {
        m.detailFocus = false
    }
    return m, nil
```
Der `else`-Zweig (Fokus-Ebene verlassen bei detailLevel==0) macht Pfeil-links zu einer FOKUS-Taste -- asymmetrisch, weil Pfeil-rechts NIE in den Detail-Fokus hineinfuehrt (nur tab tut das, PF-5/D01-revidiert, unveraendert). PO: "fuer Nutzer murks". FIX: den else-Zweig komplett entfernen:
```go
case "left":
    if m.detailLevel == 1 {
        m.detailLevel = 0
    }
    return m, nil
```
Pfeiltasten sind danach REIN Sektion/Feld-Navigation, NIE Fokus-Wechsel -- exklusiv tab/shift+tab (PF-13, unveraendert) aendern m.detailFocus. Dies REVIDIERT einen Satz aus design-spec.md §15 PF-13 ("Rueckweg bis zum Tree ... left bei detailLevel==0 -> detailFocus=false") -- design-spec ist bereits per PF-16/PF-13-Pfeil-Revision (dieser Plan) aktualisiert, hier nur der Code-Nachvollzug.

## D03 — esc als universelle Rueckwaerts-Taste (Detail-Kaskade)

Mit B01s Fix wird `esc` (keys.Back) in der Detail-Kaskade zum EINZIGEN Kandidaten fuer einen mehrstufigen Rueckweg (Pfeile machen es nicht mehr, siehe oben) -- HEUTE hat keyDetailFocus() GAR KEINEN esc-Case (faellt durch alle switches/ifs, landet beim abschliessenden `return m, nil` -- No-Op, validation.md D03 bestaetigt das als Ist-Zustand). FIX: neuer esc-Handler VOR den bestehenden Enter-Kaskaden-Ifs (Reihenfolge nach den bestehenden if-Bloecken ist unerheblich, solange VOR dem finalen `return m, nil`):
```go
if keybind.Matches(msg, keys.Back) {
    if m.detailLevel == 1 {
        m.detailLevel = 0
    } else {
        m.detailFocus = false
    }
    return m, nil
}
```
Ergebnis: EIN mentales Modell -- esc geht immer GENAU eine Ebene zurueck (Feld-Ebene -> Sektions-Ebene bei erstem esc, Sektions-Ebene -> Fokus verlassen bei zweitem esc). Beachte: dies ist STRUKTURELL das GENAUE GEGENTEIL von B01s Loeschung (B01 entfernt eine Zwei-Stufen-Kaskade AUS den Pfeiltasten, D03 fuegt SIE ALS esc-Handler wieder ein) -- nicht verwechseln, beide Aenderungen sind in dieser einen Task korrekt gebuendelt, weil sie beide Teil DESSELBEN Navigationsmodell-Wechsels sind (Pfeile=reine Navigation, tab/shift+tab=Fokus, esc=Rueckwaerts-Kaskade).

## D03 — Audit-Pflicht (Planner-Prüfauftrag, ALS DOKUMENTATION im Task zu belegen, nicht nur im Code)

PO-Auftrag woertlich: "alle esc-Sites auf Einheitlichkeit (Suche, Filter, Picker, Lobby, Kaskade, Quit) -- EIN mentales Modell: esc geht immer genau eine Ebene zurueck." Vor dem Commit eine kurze Audit-Tabelle im Commit-Body (oder Bean-Kommentar) mit File:Line-Beleg pro Standort erstellen, die belegt, dass JEDES esc bereits (oder ab dieser Task) "genau eine Ebene zurueck" bedeutet:
- Suche (m.searchActive): esc schliesst die Suche (keySearchInput, update.go) -- eine Ebene (Suche -> Basis). Bereits konform, nur verifizieren + zitieren.
- Filter (m.filterOpen): esc schliesst das Filter-Menue (keyFilterMenu, box_filter_facets.go) -- eine Ebene. Bereits konform, verifizieren + zitieren.
- Picker (Tag-/Parent-/Blocking-/Value-Menu, m.overlay): esc schliesst das jeweilige Overlay OHNE zu mutieren -- eine Ebene. Tag-Picker hat sogar bereits eine VERSCHACHTELTE Zwei-Stufen-Kaskade als Praezedenzfall (tagInputActive-Suchmodus -> esc schliesst NUR den Input, ein zweites esc schliesst danach den ganzen Picker, box_picker_tag.go keyTagInput/keyTagPicker) -- als POSITIVES Vorbild fuer "genau eine Ebene" im Audit zitieren.
- Lobby (m.view==viewLobby): esc geht zurueck zu Browse WENN ein Client existiert, sonst requestQuit() (view_lobby.go keyLobby) -- als "Lobby ist bereits der Grundzustand, esc bedeutet dort 'aufgeben'" einordnen, KEINE Aenderung noetig (aus Q06/D03-Wortlaut geht kein expliziter Aenderungswunsch fuer die Lobby hervor).
- Quit (m.confirmQuit): esc bricht ab (keyConfirmQuit, box_confirm_quit.go) -- eine Ebene (Confirm -> vorheriger View). Bereits konform.
- Kaskade (Detail-Fokus): DAS ist die einzige tatsaechliche Luecke -- durch diese Task geschlossen (s.o.).
Ergebnis: 5 von 6 Bereichen waren bereits konform (nur verifizieren + im Audit dokumentieren, NICHT anfassen), nur die Detail-Kaskade brauchte den neuen esc-Handler.

## TDD-Schritte

1. Failing tests (update_test.go): TestKeyDetailFocusLeftArrowNoLongerExitsFocusAtSectionLevel (m.detailFocus bleibt true nach left bei detailLevel==0); TestKeyDetailFocusLeftArrowStillExitsFieldLevelToSectionLevel (Regressionsschutz, detailLevel 1->0 bleibt); NEU TestKeyDetailFocusEscAtFieldLevelGoesToSectionLevel; NEU TestKeyDetailFocusEscAtSectionLevelExitsDetailFocus; NEU TestKeyDetailFocusEscNoopOutsideDetailFocus (Regressionsschutz -- falls relevant, esc ausserhalb detailFocus wird von einem ANDEREN Handler behandelt, hier nur pruefen dass keyDetailFocus selbst korrekt reagiert wenn aufgerufen).
2. command go test ./internal/tui/... -> FAIL.
3. Implementieren: update.go keyDetailFocus (left-case kuerzen, esc-Handler ergaenzen).
4. command go test ./internal/tui/... -> PASS.
5. Golden-Check (kein Golden-Update erwartet, reine Tastatur-Logik-Aenderung): command go test ./internal/tui/ -run "TestTreeGolden|TestBacklogGolden|TestChromeGolden" OHNE -update -> MUSS gruen bleiben. Falls unerwartet FAIL: Ursache klaeren.
6. command go test ./... -short gruen (2x), command go test ./... -race gruen, gofmt/vet leer.
7. Commit feat(tui): PF-16 Pfeil-Symmetrie-Fix + esc-Rueckwaerts-Kaskade (B01,D03) -- Body enthaelt die Audit-Tabelle (5 der 6 Bereiche bereits konform, Fundstellen zitiert).

## Akzeptanz-Checkliste

- [x] Pfeil-links bei detailLevel==0 ist No-Op (verlaesst Detail-Fokus NICHT mehr)
- [x] Pfeil-links bei detailLevel==1 bleibt unveraendert (Feld-Ebene -> Sektions-Ebene)
- [x] esc bei detailLevel==1 geht auf detailLevel==0 (Feld -> Sektion)
- [x] esc bei detailLevel==0 verlaesst Detail-Fokus (m.detailFocus=false)
- [x] Audit-Tabelle (Suche/Filter/Picker/Lobby/Quit) mit File:Line-Belegen im Commit-Body
- [x] Kein Golden aendert sich (Schritt 5 gruen ohne -update)
- [x] Voller Testlauf (inkl. -race) gruen, gofmt/vet leer


## Prelude aus E8-T1-Review (2026-07-16, Quelle: bt-e6q9-Review APPROVED)

- Non-blocking Finding: Section-Index-Konstanten (metaSectionIdx/bodySectionIdx/relationsSectionIdx/historySectionIdx, view_detail_bean.go) sind definiert aber noch UNGENUTZT — in diesem Task konsequent verwenden statt Magic-Numbers.
- Neue Signaturen (siehe auch '## Notes for T3' in bt-e6q9): beanSections(idx, b, bodyW, focused, activeIdx, fieldIdx, detailLevel) · renderAccordionPane(idx, b, w, h, open, secCursor, fieldCursor, detailLevel, focused). renderBeanAccordionPane unverändert.
- metaFields-Reihenfolge verschoben: tags an Index 4, created_at/updated_at jetzt 5/6 — Feld-Index-Annahmen prüfen.


## Summary

B01 (Pfeil-links verlaesst Detail-Fokus nicht mehr) und D03 (esc als
universelle Rueckwaerts-Kaskade in der Detail-Kaskade) implementiert.
keyDetailFocus's `left`-Case verlor den `else`-Zweig (`m.detailFocus =
false`) -- Pfeiltasten sind jetzt reine Section/Feld-Navigation, nie
Fokus-Wechsel. Im Gegenzug bekam keyDetailFocus einen neuen esc-Case:
Feld-Ebene -> Sektions-Ebene (1. esc), Sektions-Ebene -> Fokus-Exit
(2. esc). Audit aller esc-Sites (Suche/Filter/Picker/Lobby/Quit/
Kaskade) zeigt 5 von 6 Bereichen bereits konform -- die Detail-Kaskade
war die einzige Luecke, jetzt geschlossen.

## esc-Site-Audit-Tabelle

| Site | Datei:Zeile | Vorher | Nachher |
|---|---|---|---|
| Suche | update.go:1197 (keySearchInput) | esc schliesst Suche + leert Query — eine Ebene | unveraendert, bereits konform |
| Filter | box_filter_facets.go:278 (keyFilterMenu) | esc schliesst Filter-Menue — eine Ebene | unveraendert, bereits konform |
| Picker | box_picker_tag.go:158/212, box_picker_parent.go:100, box_picker_blocking.go:100, box_menu_value.go:151, overlay_palette.go:169 | esc schliesst Overlay ohne Mutation — eine Ebene; Tag-Picker hat bereits eine verschachtelte Zwei-Stufen-Kaskade (tagInputActive-Submodus) | unveraendert, bereits konform (positives Vorbild fuer D03) |
| Lobby | view_lobby.go:351 (keyLobby) | esc geht zurueck zu Browse wenn Client existiert, sonst requestQuit | unveraendert -- Lobby ist bereits Grundzustand ("esc = aufgeben"), kein PO-Wortlaut fuer Aenderung |
| Quit | box_confirm_quit.go:27 (keyConfirmQuit) | esc bricht ab — eine Ebene | unveraendert (B08/bt-1u0t aendert Quit-Verhalten selbst, hier nicht angefasst) |
| Kaskade (Detail-Fokus) | update.go:997 (keyDetailFocus) | GAR KEIN esc-Case -- fiel durch, No-Op (E2-Luecke) | NEU: zweistufig -- Feld->Sektion (1. esc), Sektion->Fokus-Exit (2. esc) |

## Test-Output

RED (vor Implementierung, `command go test ./internal/tui/... -run
'TestDetailFocusLeftAtSectionLevelIsNoOp|TestKeyDetailFocusEscAtFieldLevelGoesToSectionLevel|TestKeyDetailFocusEscAtSectionLevelExitsDetailFocus|TestKeyDetailFocusEscNoopOutsideDetailFocus'`):
```
--- FAIL: TestDetailFocusLeftAtSectionLevelIsNoOp
    update_test.go:748: left at section level must NOT exit detail focus anymore (B01)
--- FAIL: TestKeyDetailFocusEscAtFieldLevelGoesToSectionLevel
    update_test.go:778: esc at field level must return to section level
--- FAIL: TestKeyDetailFocusEscAtSectionLevelExitsDetailFocus
    update_test.go:799: esc at section level must exit detail focus
--- PASS: TestKeyDetailFocusEscNoopOutsideDetailFocus  (bereits gruen vor Fix -- routing-only, kein Verhalten in keyDetailFocus betroffen)
```

GREEN (nach Implementierung, gleicher Testfilter): alle 4 PASS.

Voll-Lauf (ohne `-short`, zweimal fuer Snapshot-Stabilitaet, nach
jedem der beiden Commits separat gruen):
`ok beans-tui/internal/tui 136.3s` / `136.4s` (cmd/config/data/theme
ebenfalls ok). `-race` VOLL (ohne `-short`): `ok beans-tui/internal/tui
139.8s`, keine DATA RACE. `gofmt -l .` leer, `go vet ./...` leer.
Goldens (`TestTreeGolden|TestBacklogGolden|TestChromeGolden`, OHNE
`-update`): alle PASS -- kein Golden hat sich geaendert (reine
Tastatur-Logik-Aenderung, wie erwartet).

## Umgeschriebene Alt-Tests

- `TestDetailFocusLeftAtSectionLevelExitsDetailFocus` (pinnte das ALTE
  B01-Verhalten: left bei detailLevel==0 verlaesst Detail-Fokus) ->
  umbenannt+invertiert zu `TestDetailFocusLeftAtSectionLevelIsNoOp`
  (left bei detailLevel==0 ist jetzt No-Op).

Keine weiteren Alt-Tests gefunden, die altes esc-Verhalten (No-Op in
der Detail-Kaskade) explizit pinnten -- es gab schlicht keinen
esc-Test fuer keyDetailFocus, da esc dort zuvor gar keinen Case hatte
(durchgefallen zum abschliessenden `return m, nil`).

## Smoke

Real in tmux gegen dieses Repo (`.beans/` echte Daten, `./bin/bt`),
`tmux capture-pane -p` als Beleg:

- tab -> [1] META aktiv (Accent-Bar), title zeigt `▷` (KEIN `▶`, B04
  aus T1 bleibt intakt).
- **left bei Sektions-Ebene** (B01): `▌> [1] META` bleibt bestehen --
  Detail-Fokus NICHT verlassen (vorher: sofortiger Exit zurueck zum
  Tree).
- enter -> Feld-Ebene, title zeigt `▶`.
- **esc (1x)** (D03 Rung 1): title zurueck auf `▷`, `▌> [1] META`
  bleibt aktiv -- Feld->Sektion, Fokus bleibt.
- **esc (2x)** (D03 Rung 2): `▌`-Bar verschwindet von `[1] META` --
  Detail-Fokus verlassen, zurueck zum Tree.
- **esc (3x)**: No-Op (Tree-eigene Rung 3, keine Suche/Facetten aktiv)
  -- unveraendert, ausserhalb dieses Tasks.
- `/`-Suche: Query getippt ("tag"), esc schliesst + leert Query in
  einem Schritt -- unveraendert.
- `f`-Filter: Menue geoeffnet, esc schliesst sauber -- unveraendert.
- `t`-Tag-Picker: geoeffnet, `n` -> "New tag"-Input-Submodus, esc (1x)
  schliesst NUR den Input (Picker bleibt offen, "Tags"-Titel wieder
  sichtbar), esc (2x) schliesst den ganzen Picker -- bestehende
  Zwei-Stufen-Kaskade unveraendert, als Vorbild fuer D03 bestaetigt.

Alle sechs Audit-Sites real bestaetigt (kein Punkt blieb rein auf
Unit-Ebene).

## Deviations/ERRATA

Keine ERRATA gegen bean/Plan -- Ist-Code (keyDetailFocus, Zeilen/
Struktur) stimmte exakt mit der bean-Beschreibung ueberein. Einzige
Abweichung von der bean's eigenen TDD-Schritte-Empfehlung (Schritt 7:
EIN kombinierter Commit "B01,D03"): auf explizite Anweisung des
Supervisors wurden B01 und D03 in ZWEI thematische Commits getrennt
(`fix(tui)` fuer B01, `feat(tui)` fuer D03) statt einem -- inhaltlich
identisch, nur Commit-Granularitaet abweichend vom bean-Sketch.

## Notes for T6 (bt-y2iw)

- keyDetailFocus's esc-Case (update.go:997) liegt VOR dem PF-5
  Enter-Kaskaden-Block (section-level enter alias, field-level
  kind-switch) -- T6s neuer BODY-Editor-Fall (`e`/`enter` auf
  Sektion [2] BODY) haengt am bestehenden Enter-Block, NICHT am
  esc-Case; keine Kollision, nur Reihenfolge-Hinweis.
- Die Kaskaden-/Level-Mechanik ist jetzt EINDEUTIG: `m.detailLevel`
  (0=Sektion, 1=Feld) plus `m.detailFocus` (bool) sind die einzigen
  zwei Zustandsachsen. esc dekrementiert IMMER genau eine Achse pro
  Druck (detailLevel 1->0, DANN erst beim naechsten Druck
  detailFocus true->false) -- T6s neue Value-Menu-Gruppierung
  (`buildValueMenuItems(group)`) und der BODY-Editor-Dispatch aendern
  NICHTS an dieser Zwei-Achsen-Mechanik, sie haengen sich nur an den
  bestehenden Enter-Kaskaden-`switch f.kind` (update.go, nach dem
  esc-Case) an.
- `activateDetailField`-Extraktion (T4/bt-duz7, nicht T6) wird
  denselben `switch f.kind`-Block anfassen, der jetzt direkt NACH dem
  neuen esc-Case liegt (update.go ~Zeile 1002+) -- keine funktionale
  Ueberschneidung mit B01/D03, nur raeumliche Naehe im File.
