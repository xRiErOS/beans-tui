---
# bt-duz7
title: Maus im Detail-Pane (Sektionen + Meta-Felder klickbar)
status: todo
type: task
priority: normal
created_at: 2026-07-15T21:09:20Z
updated_at: 2026-07-15T21:09:20Z
parent: bt-ntoz
blocked_by:
    - bt-e6q9
    - bt-qbyq
    - bt-y2iw
---

E8 Task 4 — deckt B07 aus bean bt-ntoz. Quelle: design-spec.md §15 PF-16. Ist-Code: internal/tui/mouse.go (handleMouse, clickPaneGeometry, treeClickRow als Vorbild), internal/tui/view_browse_repo.go/view_browse_backlog.go (Dispatch-Wiring), internal/tui/view_detail_bean.go (detailHeaderBlock = 5-Zeilen-Kopfblock-Offset), internal/tui/update.go (keyDetailFocus's Feld-Kind-Switch als wiederzuverwendende Logik). blocked_by bt-e6q9 (T1, 7-Zeilen-Meta-Liste inkl. Tags MUSS stehen, bevor die Klick-zu-Feldindex-Abbildung gebaut wird -- sonst rechnet diese Task gegen die alte 6-Zeilen-Form), bt-qbyq (T3, esc-Kaskade/Pfeil-Fix im selben Funktionsbereich von keyDetailFocus), bt-y2iw (T6, BODY-Editor-Kontext + Value-Menu-Splitting -- die Feld-Kind-Switch-Logik, die dieser Task per Doppelklick wiederverwenden soll, liegt in keyDetailFocus und wird von T6 zuletzt anfasst; diese Task extrahiert sie danach in einen gemeinsamen Helfer statt sie ein drittes Mal zu duplizieren).

## B07 — Maus im Detail-Pane

HEUTE: handleMouse() (mouse.go) dispatcht Linksklicks NUR fuer viewBrowseRepo (mouseTreeClick, beschraenkt via treeClickRow auf die LINKE Tree-Spalte -- msg.X < originX+lw) und viewBacklog (mouseBacklogClick, analog). Klicks im RECHTEN Detail-Pane werden aktuell STILLSCHWEIGEND verworfen (treeClickRow/backlogClickRow geben ok=false zurueck fuer msg.X >= originX+lw, mouseTreeClick/mouseBacklogClick tun dann nichts). PO will:
(a) Klick auf einen Sektions-Header ([1] META / [2] BODY / [3] RELATIONS / [4] HISTORY) aktiviert/expandiert diese Sektion.
(b) Klick auf eine Meta-Feldzeile selektiert das Feld.
(c) Doppelklick (oder Zweitklick auf ein bereits selektiertes Feld) oeffnet dessen Edit-Overlay -- ANALOG der bestehenden Enter-Kaskade (PF-5/T6, keyDetailFocus's Feld-Kind-Switch: status/type/priority/tags -> passendes Overlay, title -> Titel-Form, readonly -> No-Op, jump -> Beziehungs-Sprung).

## Architektur-Vorgabe (verbindlich, damit die Feld-Kind-Logik NICHT dupliziert wird)

1. Extrahiere keyDetailFocus()s bestehenden `switch f.kind { ... }`-Block (update.go, der Teil INNERHALB des detailLevel==1-Enter-Ifs -- Zeilen-Fundstelle zum Zeitpunkt dieser Task per grep "case \"status\", \"type\", \"priority\":" internal/tui/update.go ermitteln, T1/T6 haben die exakte Zeilenzahl bereits verschoben) in einen NEUEN gemeinsamen Helfer:
```go
// activateDetailField dispatcht ein Meta-/Relations-Feld auf sein passendes
// Overlay/Form/Jump -- geteilte Logik zwischen keyDetailFocus's Enter-Kaskade
// (Tastatur) und mouseDetailClick's Doppelklick (Maus), B07, design-spec.md
// §15 PF-16.
func (m model) activateDetailField(b *data.Bean, f relationField) (tea.Model, tea.Cmd) {
    switch f.kind {
    case "status", "type", "priority":
        return m.openValueMenu(f.kind), nil
    case "tags":
        return m.openTagPicker(), nil
    case "title":
        return m.openEditTitleForm(b)
    case "readonly":
        return m, nil
    default:
        if f.beanID == "" {
            return m, nil
        }
        m.expanded = expandAncestorsOf(m.idx, m.expanded, f.beanID)
        m.cursorID = f.beanID
        m.detailFocus = false
        return m, nil
    }
}
```
keyDetailFocus ruft danach nur noch `return m.activateDetailField(b, f)` an der bisherigen Stelle.

2. Neue Geometrie-Funktion `detailClickRow` (mouse.go, ANALOG treeClickRow/backlogClickRow -- gleiches Rekonstruktions-Muster ueber browseRepoChrome/backlogChrome + clickPaneGeometry, KEIN zweiter unabhaengiger Geometrie-Pfad): bildet (msg.X,msg.Y) auf entweder einen Sektions-Index (Klick auf Header-Zeile) ODER einen Feld-Index INNERHALB der aktuell offenen Sektion ab (Klick auf eine Body-Zeile waehrend die Sektion offen ist). MUSS den Kopfblock-Offset beruecksichtigen: detailHeaderBlock() (view_detail_bean.go) rendert IMMER 5 Zeilen (ID/Titel/Leerzeile/type-status-prio/Leerzeile) VOR der eigentlichen Accordion -- jede Klick-Y-Koordinate im Detail-Pane muss um diese 5 Zeilen nach dem Pane-eigenen originY (clickPaneGeometry) UND vor dem Start der Accordion-Zeilen-Zaehlung verschoben werden. Innerhalb der Accordion selbst: EINE Header-Zeile pro Sektion (+1 fieldStrip-Zeile NUR wenn die Sektion offen+aktiv+fields>0 UND es NICHT Meta selbst ist, renderAccordion-Konvention, accordion.go) + Body-Zeilen wenn offen (PF-1: Meta IMMER offen). Fuer Meta-Body-Zeilen (7 Zeilen seit T1) ist die Abbildung Zeile->Feldindex direkt (Zeile 0 = title, ..., Zeile 6 = updated_at, feste Reihenfolge aus metaFields()). Fuer Relations-Body-Zeilen (variable Feldzahl, Gruppen-Header dazwischen) reicht fuer v1 EINE vereinfachte Abbildung: Klick auf die Relations-Sektion waehrend sie offen ist selektiert die Sektion (Fall a), OHNE feingranulare Feld-Adressierung ueber Gruppen-Header hinweg zu erzwingen (Klick-zu-Feld INNERHALB Relations ist NICE-TO-HAVE, kein PO-Wortlaut-Zitat verlangt es explizit -- Fall b/c sind fuer B07 primaer an der Meta-Feldliste demonstriert/verlangt, PO-Formulierung "Meta-Feldzeilen"). Diese Vereinfachung ALS ERRATUM/Scope-Entscheidung im Commit-Body dokumentieren.

3. Neuer Dispatch `mouseDetailClick(msg tea.MouseMsg) (tea.Model, tea.Cmd)` (mouse.go): loest ueber detailClickRow auf, setzt bei Sektions-Klick `m.detailFocus=true; m.secCursor=<idx>; m.accOpen=<idx>+1; m.detailLevel=0` (mirrort den bestehenden Digit-Jump-Case in keyDetailFocus, update.go), bei Meta-Feld-Klick zusaetzlich `m.detailLevel=1; m.fieldCursor=<feldidx>`. Doppelklick-Erkennung MIRRORT mouseTreeClick's bestehendes lastClickIdx/lastClickAt-Muster (doubleClickInterval, gleiche Felder auf model wiederverwenden, KEIN zweites Zeitfenster-Paar einfuehren) -- bei Doppelklick auf ein Feld (oder Zweitklick auf ein BEREITS selektiertes Feld) zusaetzlich `m.activateDetailField(b, f)` aufrufen.

4. Wiring: handleMouse()s `switch m.view { case viewBrowseRepo: return m.mouseTreeClick(msg) ... }` -- HEUTE dispatcht mouseTreeClick auch bei einem Klick RECHTS der Tree-Spalte (gibt dann nur ok=false zurueck, tut nichts). NEU: mouseTreeClick selbst (ODER handleMouse direkt, Implementer-Entscheidung anhand der saubersten Diff-Groesse) prueft ZUERST ob msg.X in der Tree-Spalte liegt (treeClickRow-Bounds) -- wenn NICHT, an mouseDetailClick delegieren (statt No-Op). Analog fuer viewBacklog/mouseBacklogClick. Toast-Dismiss-Vorrang (handleMouse's allererster Check, m.toastHit) UND der Overlay-Guard-Block (m.form!=nil || m.overlay!=overlayNone || ...) bleiben UNVERAENDERT und WIRKEN BEREITS fuer den neuen Dispatch-Pfad (er sitzt hinter demselben Guard) -- KEINE Aenderung an diesen beiden Bloecken noetig, nur verifizieren dass mouseDetailClick tatsaechlich HINTER ihnen haengt.

## TDD-Schritte

1. Failing tests (mouse_test.go): NEU TestDetailClickRowMapsSectionHeaderClick (Klick auf Header-Zeile von Sektion 3 -> secIdx=2); NEU TestDetailClickRowMapsMetaFieldClick (Klick auf die 5. Meta-Zeile [tags:, seit T1] -> fieldIdx=4); NEU TestDetailClickRowAccountsForFiveLineHeaderOffset (Klick auf Kopfblock-Zeilen selbst -> ok=false, kein Treffer); NEU TestMouseDetailClickSingleClickSelectsField (kein Overlay oeffnet sich); NEU TestMouseDetailClickDoubleClickOnFieldOpensOverlay (mirrort TestMouseTreeClick-Doppelklick-Tests); NEU TestMouseDetailClickIgnoredWhenOverlayOpen (Regressionsschutz ueber den bestehenden Guard). update_test.go: NEU TestActivateDetailFieldStatusOpensValueMenu (+ type/priority/tags/title/readonly/jump -- 7 Faelle, mirrort die bestehenden keyDetailFocus-Enter-Kaskade-Tests aus E7/T6, jetzt gegen den extrahierten Helfer).
2. command go test ./internal/tui/... -> FAIL.
3. Implementieren (Reihenfolge: update.go zuerst [activateDetailField-Extraktion, keyDetailFocus ruft ihn], dann mouse.go [detailClickRow, mouseDetailClick], dann handleMouse-Wiring).
4. command go test ./internal/tui/... -> PASS.
5. Golden-Check (kein Golden-Update erwartet -- reine Maus-/Klick-Logik, kein Render-Output-Unterschied): command go test ./internal/tui/ -run "TestTreeGolden|TestBacklogGolden|TestChromeGolden" OHNE -update -> MUSS gruen bleiben.
6. command go test ./... -short gruen (2x), command go test ./... -race gruen, gofmt/vet leer.
7. Commit feat(tui): PF-16 Maus im Detail-Pane -- Sektionen + Meta-Felder klickbar (B07) -- Body dokumentiert die Relations-Feld-Vereinfachung als ERRATUM/Scope-Entscheidung.

## Akzeptanz-Checkliste

- [ ] Klick auf einen Sektions-Header aktiviert/expandiert die Sektion (m.detailFocus=true, m.secCursor/m.accOpen gesetzt)
- [ ] Klick auf eine Meta-Feldzeile selektiert das Feld (m.detailLevel=1, m.fieldCursor gesetzt), OHNE Overlay zu oeffnen
- [ ] Doppelklick auf ein Meta-Feld (bzw. Zweitklick auf bereits selektiertes Feld) oeffnet dessen Overlay -- identisch zur Enter-Kaskade
- [ ] Kopfblock-Offset (5 Zeilen) korrekt beruecksichtigt -- Klicks auf ID/Titel/type-status-prio-Zeilen sind kein Sektions-/Feld-Treffer
- [ ] Toast-Dismiss-Vorrang + Overlay-Guard wirken unveraendert fuer den neuen Dispatch-Pfad (verifiziert, nicht neu gebaut)
- [ ] activateDetailField() ist die EINZIGE Stelle, die Feld-Kind auf Overlay dispatcht (keyDetailFocus UND mouseDetailClick rufen denselben Helfer)
- [ ] Kein Golden aendert sich
- [ ] Voller Testlauf (inkl. -race) gruen, gofmt/vet leer
