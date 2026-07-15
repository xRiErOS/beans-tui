---
# bt-qbyq
title: 'Detail-Focus: Pfeil-Symmetrie-Fix + esc-Rueckwaerts-Kaskade'
status: in-progress
type: task
priority: high
created_at: 2026-07-15T21:07:40Z
updated_at: 2026-07-15T21:45:13Z
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

- [ ] Pfeil-links bei detailLevel==0 ist No-Op (verlaesst Detail-Fokus NICHT mehr)
- [ ] Pfeil-links bei detailLevel==1 bleibt unveraendert (Feld-Ebene -> Sektions-Ebene)
- [ ] esc bei detailLevel==1 geht auf detailLevel==0 (Feld -> Sektion)
- [ ] esc bei detailLevel==0 verlaesst Detail-Fokus (m.detailFocus=false)
- [ ] Audit-Tabelle (Suche/Filter/Picker/Lobby/Quit) mit File:Line-Belegen im Commit-Body
- [ ] Kein Golden aendert sich (Schritt 5 gruen ohne -update)
- [ ] Voller Testlauf (inkl. -race) gruen, gofmt/vet leer


## Prelude aus E8-T1-Review (2026-07-16, Quelle: bt-e6q9-Review APPROVED)

- Non-blocking Finding: Section-Index-Konstanten (metaSectionIdx/bodySectionIdx/relationsSectionIdx/historySectionIdx, view_detail_bean.go) sind definiert aber noch UNGENUTZT — in diesem Task konsequent verwenden statt Magic-Numbers.
- Neue Signaturen (siehe auch '## Notes for T3' in bt-e6q9): beanSections(idx, b, bodyW, focused, activeIdx, fieldIdx, detailLevel) · renderAccordionPane(idx, b, w, h, open, secCursor, fieldCursor, detailLevel, focused). renderBeanAccordionPane unverändert.
- metaFields-Reihenfolge verschoben: tags an Index 4, created_at/updated_at jetzt 5/6 — Feld-Index-Annahmen prüfen.
