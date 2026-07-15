---
# bt-duz7
title: Maus im Detail-Pane (Sektionen + Meta-Felder klickbar)
status: completed
type: task
priority: normal
created_at: 2026-07-15T21:09:20Z
updated_at: 2026-07-15T23:25:07Z
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

- [x] Klick auf einen Sektions-Header aktiviert/expandiert die Sektion (m.detailFocus=true, m.secCursor/m.accOpen gesetzt)
- [x] Klick auf eine Meta-Feldzeile selektiert das Feld (m.detailLevel=1, m.fieldCursor gesetzt), OHNE Overlay zu oeffnen
- [x] Doppelklick auf ein Meta-Feld (bzw. Zweitklick auf bereits selektiertes Feld) oeffnet dessen Overlay -- identisch zur Enter-Kaskade
- [x] Kopfblock-Offset (5 Zeilen) korrekt beruecksichtigt -- Klicks auf ID/Titel/type-status-prio-Zeilen sind kein Sektions-/Feld-Treffer
- [x] Toast-Dismiss-Vorrang + Overlay-Guard wirken unveraendert fuer den neuen Dispatch-Pfad (verifiziert, nicht neu gebaut)
- [x] activateDetailField() ist die EINZIGE Stelle, die Feld-Kind auf Overlay dispatcht (keyDetailFocus UND mouseDetailClick rufen denselben Helfer)
- [x] Kein Golden aendert sich
- [x] Voller Testlauf (inkl. -race) gruen, gofmt/vet leer



## Summary

B07 implemented: (a) Klick auf einen Sektions-Header aktiviert/expandiert
die Sektion; (b) Klick auf eine Meta-Feldzeile selektiert das Feld, OHNE
Overlay zu oeffnen; (c) Doppelklick (bzw. Zweitklick auf ein bereits
selektiertes Feld) oeffnet dessen Edit-Overlay -- identisch zur
Enter-Kaskade. `activateDetailField(b, f)` (update.go) wurde EXAKT wie im
Architektur-Vorgabe-Sketch aus `keyDetailFocus`s bisherigem
`switch f.kind`-Block extrahiert -- `keyDetailFocus` ruft jetzt nur noch
`return m.activateDetailField(b, f)` an der bisherigen Stelle.
`detailClickRow` (mouse.go, NEU) bildet einen Klick im Detail-Pane auf
entweder einen Sektions-Index (Header-Zeile, fieldIdx=-1) oder einen
Meta-Feld-Index (secIdx=metaSectionIdx, fieldIdx>=0) ab -- render-geerdet
gegen den ECHTEN `beanSections`/`renderAccordion`-Zustand (m.detailFocus/
m.secCursor/m.accOpen/m.fieldCursor/m.detailLevel), Kopfblock-Offset (5
Zeilen, `detailHeaderBlock`) explizit beruecksichtigt. `mouseDetailClick`
(mouse.go, NEU) dispatcht: Einzelklick selektiert (Sektion oder Feld),
Doppelklick auf einem Meta-Feld ruft `activateDetailField`; ein Doppelklick
auf der BODY-Sektion (die kein Feld hat) oeffnet stattdessen `$EDITOR` ueber
denselben `openBodyEditor`-Helfer, den `e`/`enter` bereits nutzen (bt-y2iw's
"Notes for bt-duz7"). Doppelklick-Erkennung nutzt DIESELBEN
`m.lastClickIdx`/`m.lastClickAt`-Felder wie `mouseTreeClick` (kein zweites
Zeitfenster-Paar) -- ein `clickKey := secIdx*10+fieldIdx+1`-Encoding trennt
Sektions- von Feld-Treffer innerhalb des Detail-Pane selbst. Wiring:
`mouseTreeClick`/`mouseBacklogClick` delegieren an `mouseDetailClick`, wenn
ihre eigene `treeClickRow`/`backlogClickRow`-Aufloesung `ok=false`
liefert (kleinster Diff, `detailClickRow` validiert die X-Bounds
unabhaengig erneut, siehe Deviations). Toast-Dismiss-Vorrang +
Overlay-Guard (`handleMouse`) sind UNVERAENDERT und wirken bereits fuer
den neuen Dispatch-Pfad (verifiziert per Test + Live-Smoke, nicht neu
gebaut).

## Test-Output

RED (activateDetailField, verifiziert per gezieltem Revert -- Extraktion,
kein neues Verhalten, alle 8 neuen `TestActivateDetailField*`-Tests waren
sofort GREEN nach der Extraktion, siehe Deviations):
`command go build -o bin/bt .` vor der Extraktion kompilierte bereits
(Ist-Code), die neuen Tests riefen direkt `m.activateDetailField(...)` auf
-- vor dem Hinzufuegen der Methode waere das ein Compile-Fehler
(`undefined: activateDetailField`) gewesen; nach Hinzufuegen sofort PASS
(alle 8 Faelle: status/type/priority/tags/title/readonly/jump/
jump-unresolved).

RED (detailClickRow/mouseDetailClick, echter Revert-Beleg):
```
$ git stash push -- internal/tui/mouse.go
$ command go test ./internal/tui/... -run "TestDetailClickRow|TestMouseDetailClick" -v
internal/tui/mouse_test.go:394:14: undefined: detailClickRow
internal/tui/mouse_test.go:408:26: undefined: detailClickRow
internal/tui/mouse_test.go:429:26: undefined: detailClickRow
FAIL	beans-tui/internal/tui [build failed]
$ git stash pop
```

GREEN (nach Implementierung, alle 9 neuen Detail-Pane-Maus-Tests):
```
$ command go test ./internal/tui/... -run "TestDetailClickRow|TestMouseDetailClick" -v
--- PASS: TestDetailClickRowAccountsForFiveLineHeaderOffset (0.00s)
--- PASS: TestDetailClickRowMapsSectionHeaderClick (0.00s)
--- PASS: TestDetailClickRowMapsMetaFieldClick (0.00s)
--- PASS: TestMouseDetailClickSectionHeaderActivatesAndExpands (0.00s)
--- PASS: TestMouseDetailClickSingleClickSelectsField (0.00s)
--- PASS: TestMouseDetailClickDoubleClickOnFieldOpensOverlay (0.00s)
--- PASS: TestMouseDetailClickDoubleClickOnBodySectionOpensEditor (0.00s)
--- PASS: TestMouseDetailClickIgnoredWhenOverlayOpen (0.00s)
--- PASS: TestMouseDetailClickReachableFromBacklogView (0.00s)
PASS
```

Golden-Gegenbeleg (OHNE `-update`, MUSS gruen bleiben -- reine Maus-/
Tastatur-Logik, kein Render-Output-Unterschied): `command go test
./internal/tui/ -run "TestTreeGolden|TestBacklogGolden|TestChromeGolden" -v`
-> alle 5 PASS (TestChromeGolden/TestTreeGolden/
TestTreeGoldenDeterministic/TestBacklogGolden/
TestBacklogGoldenDeterministic). `git status --short internal/tui/testdata/`
leer -- kein Golden geaendert.

Voll-Lauf (ohne `-short`), ZWEI frische Prozesse: `command go test ./...`
-> `ok beans-tui/internal/tui 136.051s` (Lauf 1); `command go test ./...
-count=1` -> `ok beans-tui/internal/tui 136.803s` (Lauf 2, alle Pakete
cmd/config/data/theme ebenfalls PASS in beiden Laeufen). `command go test
./... -race` -> `ok beans-tui/internal/tui 139.674s`, keine DATA RACE.
`gofmt -l .` leer. `command go vet ./...` leer. `command go build -o bin/bt
.` clean.

## Smoke

Real in tmux gegen dieses Repo (`.beans/` echte Daten, `./bin/bt`,
`EDITOR=true` fuer den Editor-Klick-Pfad -- keine echte Mutation), tmux
100x30, Maus-Events per xterm-SGR-Escape-Sequenzen ueber `tmux send-keys -l`
injiziert (`\x1b[<0;COL;ROWM`), `tmux capture-pane -p` als Beleg:

- Fall a (Sektions-Klick): Klick auf `> [3] RELATIONS`-Header (real
  gerenderte Spalte/Zeile via Python-Scan des capture-pane-Outputs
  bestimmt, nicht angenommen) -> Sektion sofort `▌`-aktiv + expandiert
  (`▾`, Fields-Strip + Children-Body sichtbar). META klappte dabei
  automatisch weiterhin mit (PF-1, Meta immer offen).
- Fall b (Feld-Klick): Klick auf `tags:`-Zeile -> META reaktiviert (`▌`),
  `▶`-Marker auf `tags:` (Feld selektiert), RELATIONS wieder zu -- KEIN
  Overlay geoeffnet.
- Fall c (Doppelklick Feld): zwei `tags:`-Klicks als EIN `send-keys -l`
  ohne Pause (< 500ms) -> Tag-Picker-Overlay ("Tags", "▸ [x] to-review
  (8)") oeffnete sich sofort; `esc` schloss sauber, KEINE Mutation
  (`git status --short .beans/` leer danach).
- Fall c (BODY-Sonderfall): zwei `[2] BODY`-Klicks als EIN `send-keys -l`
  -> Sektion aktiviert+expandiert UND `$EDITOR` (=`true`) suspendierte/
  resumte sofort (TUI blieb responsiv, Body-Markdown korrekt gerendert
  danach) -- `git status --short .beans/` bestaetigt KEINE Mutation
  (EDITOR=true schreibt nichts in die temp-Datei zurueck).
- Overlay-Guard (live): `s` oeffnete das Status-Value-Menu-Overlay, ein
  anschliessender Klick auf dieselbe `[3] RELATIONS`-Koordinate (jetzt
  HINTER dem Overlay) veraenderte NICHTS (Screenshot vor/nach identisch,
  Overlay blieb offen) -- der neue Dispatch-Pfad sitzt bestaetigt HINTER
  dem bestehenden Guard, kein Fehlklick-Durchgriff.
- Toast-Dismiss-Vorrang: nicht separat live nachgestellt (der Code-Pfad
  ist strukturell UNVERAENDERT und liegt in `handleMouse` VOR jedem
  Dispatch, inkl. dem neuen -- die bestehende Regression
  `TestToastClickDismissesEvenWithFormOpen` deckt die Reihenfolge bereits
  ab und lief gruen im Voll-Lauf).

Alle vier PO-Faelle (a/b/c-Feld/c-Body) UND der Overlay-Guard real in tmux
gegen das echte Binary bestaetigt -- keiner blieb rein auf Unit-/
Render-Ebene. Relations-Feld-Klick-innerhalb-der-Sektion (v1-Vereinfachung,
siehe Deviations) wurde NICHT separat live erprobt -- ist unit-getestet
(`TestDetailClickRowMapsSectionHeaderClick` deckt den Fall bereits: ein
Klick waehrend RELATIONS offen ist bleibt ein Sektions-Treffer).

## Deviations/ERRATA

Keine ERRATA gegen bean/Plan -- Ist-Code (Section-Index-Konstanten,
`beanSections`/`renderAccordion`-Signaturen, `openValueMenu`/
`openTagPicker`/`openBodyEditor`/`openEditTitleForm`) stimmte exakt mit den
bean-Sketches ueberein.

Dokumentierte Scope-/Implementer-Entscheidungen (wie im Auftrag
vorgesehen):

1. **Relations-Feld-Klick v1-Vereinfachung** (Architektur-Vorgabe #2,
   explizit im bean vorgezeichnet): ein Klick INNERHALB einer offenen
   RELATIONS/BODY/HISTORY-Sektion (Body-Zeilen oder deren fieldStrip-Zeile)
   resolved auf einen Sektions-Treffer (fieldIdx=-1), NICHT auf ein
   einzelnes Beziehungs-Feld -- nur META hat eine feste, direkte
   Zeile->Feldindex-Abbildung. Kein PO-Wortlaut verlangt mehr fuer v1.
2. **Wiring-Diff-Groesse** (Architektur-Vorgabe #4, "Implementierer-
   Entscheidung"): `mouseTreeClick`/`mouseBacklogClick` delegieren
   UNBEDINGT an `mouseDetailClick`, sobald ihre eigene `*ClickRow`-
   Aufloesung `ok=false` liefert (statt zuerst separat die X-Spalte zu
   pruefen) -- `detailClickRow` validiert die Detail-Pane-X-Bounds
   unabhaengig selbst und liefert `ok=false` fuer jeden Klick, der noch
   in der Tree-/Backlog-Spalte liegt (z.B. unterhalb der letzten
   sichtbaren Zeile) -- keine Cross-Pane-Fehlzuordnung moeglich, kleinster
   Diff (2 Zeilen je Funktion).
3. **Doppelklick-clickKey-Encoding**: `secIdx*10+fieldIdx+1` trennt
   Sektions-Treffer (Vielfache von 10: 0/10/20/30) von Meta-Feld-Treffern
   (1..7) innerhalb DESSELBEN geteilten `m.lastClickIdx`-Feldes (bean-
   Vorgabe "gleiche Felder wiederverwenden, kein zweites Zeitfenster-
   Paar"). Ein Kollisions-Restrisiko gegen einen Tree-/Backlog-Row-Index
   im selben 500ms-Fenster bleibt (z.B. Klick auf Tree-Zeile 10, dann
   Klick auf einen RELATIONS-Header innerhalb 500ms) -- akzeptiert als
   direkte Konsequenz der bean-eigenen Wiederverwendungs-Vorgabe, nicht
   neu eingefuehrt, dokumentiert statt stillschweigend uebernommen.
4. **"Doppelklick" == "Zweitklick auf bereits selektiertes Feld"**: als
   EIN Ereignis behandelt (dieselbe `isDouble`-Pruefung wie
   `mouseTreeClick`), nicht als zwei separate Mechanismen -- die bean-
   Formulierung "(bzw. Zweitklick...)" wurde als Umschreibung desselben
   Zeitfenster-Treffers gelesen, nicht als zusaetzlicher, zeitfenster-
   loser Zustands-Check. Kein zweites Zeitfenster-Paar noetig (matcht
   Architektur-Vorgabe #3 woertlich).

## Notes for bt-yqdy

- Palette-/Overlay-Guard-Beruehrungspunkte: `mouseDetailClick` (mouse.go)
  sitzt HINTER demselben `handleMouse`-Guard-Block wie jeder andere
  Maus-Dispatch (`m.form != nil || m.overlay != overlayNone ||
  m.paletteOpen || ...`) -- WENN bt-yqdy's Command-Center-Aenderungen
  (B13/B14) diesen Guard-Block selbst anfassen, betrifft das automatisch
  auch den neuen Detail-Pane-Klick-Pfad (keine separate Anpassung noetig,
  aber bei einer Guard-Aenderung gegen `TestMouseDetailClickIgnoredWhen
  OverlayOpen` (mouse_test.go) gegenpruefen).
- `activateDetailField(b, f)` (update.go, NEU) ist jetzt die EINZIGE
  Stelle, die ein Meta-/Relations-Feld auf ein Overlay dispatcht -- falls
  bt-yqdy's `create tag`-Palette-Command (B14) einen Tag-Picker OHNE
  fokussiertes Feld oeffnen will, ist das ein SEPARATER Pfad
  (`m.openTagPicker().openTagInput()` direkt, kein Feld-Kontext) und
  beruehrt `activateDetailField` nicht.
- Kein neuer Key/keine neue Keybinding wurde ergaenzt (nur Maus-Dispatch)
  -- `helpGroups()`/Drift-Guard unveraendert, keine Beruehrung mit bt-yqdy's
  `keys.NewTag`-Ergaenzung.
