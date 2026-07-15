---
# bt-mne6
title: E5 T4 — Maus (Wheel/Klick/Doppelklick)
status: completed
type: task
priority: normal
created_at: 2026-07-15T09:04:24Z
updated_at: 2026-07-15T11:06:53Z
parent: bt-5h4d
---

Ziel: Maus (Wheel-Scroll, Klick setzt Cursor, Doppelklick expand/collapse --
Port devd `update.go`-Mausteil, T8-Opus-Review-Note: `windowStart`
(view_browse_repo.go:423) für Click-Y->Row wiederverwenden, NICHT neu
erfinden). Design decision f: kein separates `m.scroll`-Feld existiert (T8
Chrome()/ChromeOpts.Scroll ist toter Code, nur von chrome_test.go
konsumiert) -- Wheel bewegt daher den JEWEILIGEN View-Cursor (treeCursor-
Äquivalent via m.cursorID/keyTree-up-down, m.backlogList, m.reviewCursor),
nicht einen Scroll-Offset. Toast-Klick-Dismiss (Task 1) MUSS vor jedem
regulären Mouse-Dispatch geprüft werden.

Plan: docs/plans/v1-port/epic-E5-plan.md »Task 4«.

## Akzeptanz
- [x] internal/tui/update.go Update(): neuer `case tea.MouseMsg: return
      m.handleMouse(msg)` -- VOR dem `if m.form != nil { return
      m.updateForm(msg) }`-Fallback (Port devd Cross-Feature-Fix, Toast-
      Klick darf nicht von einem offenen Formular geschluckt werden)
- [x] internal/tui/mouse.go (NEU): handleMouse(msg tea.MouseMsg) -- Toast-
      Hit-Test zuerst (m.toastHit -> dismissToast), dann Overlay-Guard
      (Formulare/Picker/Menüs/Palette/Suche/Filter ignorieren Maus,
      Precedent devd), dann Wheel (Up/Down bewegt den aktiven View-Cursor
      um 1/3 Zeilen -- gleiche Richtung wie i/k), dann Linksklick-Dispatch
      je m.view (viewBrowseRepo/viewBacklog/viewReviewCockpit)
- [x] internal/tui/view_browse_repo.go: treeClickRow(m model, nodes
      []treeNode, msg tea.MouseMsg) (idx int, ok bool) -- rekonstruiert
      lw/rw/bodyH IDENTISCH zu viewBrowseRepo()s eigener Render-Formel
      (Golden-Rule-Drift-Schutz, Kommentar-Pflicht analog windowStart),
      -1 Offset für die Suchkopfzeile, windowStart(len(nodes), bodyH-3,
      cursorPos) für die Fensterung; Doppelklick (500ms-Fenster, Port devd
      doubleClickInterval -- NEUE Model-Felder lastClickIdx/lastClickAt)
      togglet expand/collapse auf demselben Knoten, Einzelklick setzt nur
      Cursor (devd D03-Semantik: Einzelklick auf offenen Knoten toggelt
      NICHT)
- [x] internal/tui/view_browse_backlog.go: backlogClickRow -- analoge
      Click-Y->Row-Funktion für backlogList (kein Doppelklick-Sinn, flache
      Liste)
- [x] internal/tui/view_review_cockpit.go: reviewClickRow -- analog für
      reviewFlat/reviewCursor
- [x] internal/tui/app.go: Run() bleibt bei tea.WithMouseCellMotion()
      (bereits gesetzt seit T8) -- KEINE Änderung nötig, nur Doku-Update
      im Kommentar (Mausteil jetzt tatsächlich verdrahtet)
- [x] mouse_test.go (NEU): TestWheelUpDownMovesTreeCursor,
      TestWheelMovesBacklogCursor, TestWheelMovesReviewCursor,
      TestClickSetsTreeCursor, TestDoubleClickTogglesExpand,
      TestSingleClickOnOpenNodeDoesNotCollapse, TestMouseIgnoredWhileFormOpen,
      TestMouseIgnoredWhileOverlayOpen, TestToastClickDismissesEvenWithFormOpen
      (Cross-Feature-Fix-Regression, Port devd's eigener Grund für die
      Reihenfolge)
- [x] `command go test ./... -short` grün, gofmt/vet leer, Goldens 2x grün
      (Maus berührt keinen Default-Render-Pfad)
- [x] Commit `feat(tui): Maus (Wheel/Klick/Doppelklick, Toast-Dismiss-Vorrang)`

## Summary

internal/tui/mouse.go (NEU, ~190 Zeilen): `handleMouse`/`wheelMove`/
`mouseTreeClick`/`mouseBacklogClick`/`mouseReviewClick` + die geteilte
`clickPaneGeometry(w, h int, head, localKeys string) (bodyH, lw, rw,
originX, originY int)` -- letztere ist jetzt die EINE Quelle für Pane-
Geometrie, genutzt sowohl von den drei View-Funktionen (`viewBrowseRepo`/
`viewBacklog`/`viewReviewCockpit`) ALS AUCH von den drei `*ClickRow`-
Funktionen (Golden-Rule-Drift-Schutz -- ein Abweichen der Render-Formel
kann die Klick-Formel nicht mehr unbemerkt zurücklassen, weil beide
dieselbe Funktion aufrufen). Toast-Hit-Test zuerst (unconditional),
dann Overlay-Guard (`m.form`/`m.overlay`/`paletteOpen`/`filterOpen`/
`searchActive`/`helpOpen`/`confirmQuit`), dann Wheel (`wheelMove`, ±1
je Tick, dispatcht auf die drei `*CursorMove`-Helfer), dann Linksklick
je `m.view`.

`browseRepoChrome`/`backlogChrome`/`reviewCockpitChrome` (je View-Datei,
NEU): extrahieren Head/Footer-String-Bau aus den drei View-Funktionen --
reiner Extract (keine Verhaltensänderung, goldenverifiziert), einzige
Quelle für die drei `*ClickRow`-Funktionen UND die drei View-Funktionen
selbst. `treeCursorMove`/`backlogCursorMove`/`reviewCursorMove` (je View-
Datei, NEU): aus `keyTree`/`keyBacklog`/`keyReviewCockpit`s eigenen Up/
Down-Cases faktoriert -- Tastatur UND Wheel teilen sich jetzt exakt
dieselbe Klemm-Logik (kein Duplikat).

`treeClickRow`/`backlogClickRow`/`reviewClickRow` (je View-Datei, NEU):
Klick-Y -> Row-Index, rekonstruieren `bodyH`/`lw`/`originX`/`originY`
über `clickPaneGeometry`. `reviewClickRow` (Sonderfall): läuft
`rs.groups`/`rs.rework` in EXAKT derselben Reihenfolge wie
`reviewQueueRows` erneut ab (eigener `reviewRowRef`-Zähltyp, KEIN
Rendering) um Header-/Separator-Zeilen (kein Klick-Ziel) von Bean-Zeilen
zu unterscheiden -- `reviewQueueRows` selbst bleibt unangetastet
(Golden-Risiko null).

internal/tui/types.go: `lastClickIdx int`/`lastClickAt time.Time`/
`clock func() time.Time` (Doppelklick-State + Test-Injection).
internal/tui/update.go: `doubleClickInterval`-Konstante + `now()`-
Methode (Port devd verbatim), `case tea.MouseMsg: return
m.handleMouse(msg)` an der bereits in T1 dokumentierten Stelle --
beans-tui's `Update()` hat (anders als devd) KEINEN Top-Level-`if
m.form != nil`-Kurzschluss VOR dem eigenen switch, daher deckt die
switch-Case-Platzierung Toast-Klick-Vorrang bei offenem Formular
automatisch ab (kein zweiter Guard nötig, siehe TestToastClickDismisses
EvenWithFormOpen).

## Test-Output

RED bewiesen: mouse_test.go zuerst mit den 9 Akzeptanz-Tests + Helfern
(`wheelMsg`/`screenLines`/`leftPaneClickAt`/`treeClickAt`) gegen den
unveränderten Stand -> Compile-Fail (`undefined: handleMouse`,
`undefined: clickPaneGeometry`, `undefined: treeClickRow`). Nach
Implementierung (mouse.go, types.go/update.go-Felder, treeClickRow/
backlogClickRow/reviewClickRow + die drei Chrome-/CursorMove-Extracts)
-> alle 9 grün (`go test ./internal/tui/ -run
'TestWheel|TestClickSetsTreeCursor|TestDoubleClickTogglesExpand|
TestSingleClickOnOpenNodeDoesNotCollapse|TestMouseIgnoredWhile|
TestToastClickDismissesEvenWithFormOpen' -v`):
TestWheelUpDownMovesTreeCursor, TestWheelMovesBacklogCursor,
TestWheelMovesReviewCursor, TestClickSetsTreeCursor,
TestDoubleClickTogglesExpand, TestSingleClickOnOpenNodeDoesNotCollapse,
TestMouseIgnoredWhileFormOpen, TestMouseIgnoredWhileOverlayOpen,
TestToastClickDismissesEvenWithFormOpen.

Mutation-Verifikation (Vertrauensnachweis, nicht Teil des dauerhaften
Testcodes): Overlay-Guard testweise mit `if false && (...)` deaktiviert
-> TestMouseIgnoredWhileFormOpen/…WhileOverlayOpen schlagen korrekt fehl;
Doppelklick-Bedingung testweise auf reinen Einzelklick-Toggle verkürzt
-> TestDoubleClickTogglesExpand/TestSingleClickOnOpenNodeDoesNotCollapse
schlagen korrekt fehl; `originY`-Formel in `clickPaneGeometry` testweise
um +1 verschoben -> TestClickSetsTreeCursor/…DoubleClick…/…SingleClick…
schlagen korrekt fehl (beweist: die Klick-Tests sind render-geerdet
gegen die ECHTE View()-Ausgabe, nicht zirkulär gegen dieselbe Formel).
Jede Mutation danach exakt zurückgesetzt (`diff` gegen Backup
bytegleich bestätigt) vor dem finalen Lauf.

`go test ./... -short` grün (alle 4 Packages). `go vet ./...` leer.
`gofmt -l .` leer. Voller Lauf `go test ./... -count=1` GRÜN
(internal/tui 126-127s über zwei Läufe, internal/data ~2.1s, internal/
theme ~0.6-0.8s, cmd ~0.3-1.0s). Alle 7 Goldens (TestChromeGolden/
TestTreeGolden/TestTreeGoldenDeterministic/TestBacklogGolden/
TestBacklogGoldenDeterministic/TestReviewCockpitGolden/
TestReviewCockpitGoldenDeterministic) `-count=2` grün,
`git status --porcelain internal/tui/testdata/` leer (byte-identisch --
die drei View-Funktionen sind zwar intern refaktoriert (Chrome-/
Geometrie-Extract), aber beweisbar output-identisch).

## Smoke

ECHTE tmux-Maus (SGR-Mausprotokoll, `tmux send-keys -l` mit rohen
Escape-Sequenzen `\x1b[<Cb;Cx;CyM`), 100x30, Scratch-Repo (`beans init`
+ 1 Milestone -> 1 Epic -> 2 Tasks + 1 Bug, Task Two mit Body + Tag
to-review, plus 2 parentlose Backlog-Beans), Binary bin/bt frisch
gebaut. Kein Update-Level-Ersatz nötig -- tmux-Maus war praktikabel:

1. **Wheel (Tree):** Wheel-Down bewegt den Cursor-Balken von Milestone
   auf die nächste Root-Zeile (Backlog Task), zweites Wheel-Down auf
   Backlog Task Two, zweimal Wheel-Up zurück auf Milestone (geklemmt).
2. **Klick setzt Cursor + expandiert (Tree):** Klick auf die
   zugeklappte Milestone-Zeile -> Marker `▸`->`▾`, Epic-Kind erscheint
   sofort (devd-Semantik: Einzelklick auf zugeklapptem Knoten expandiert).
3. **Doppelklick kollabiert (Tree):** zwei `tmux send-keys`-Klicks auf
   dieselbe (jetzt offene) Milestone-Zeile ohne Pause dazwischen ->
   kollabiert (`▾`->`▸`, Epic-Zeile verschwindet) -- ECHTE Wanduhrzeit,
   keine injizierte Clock.
4. **Einzelklick auf offenem Knoten kollabiert NICHT:** Klick, >1s
   Pause, zweiter Klick auf derselben Zeile -> bleibt offen (devd D03,
   bestätigt mit echter Zeitverzögerung statt der Unit-Test-Clock).
5. **Toast-Klick-Dismiss:** `y` auf der Milestone -> Toast "Kopiert:
   ...4p1f" oben rechts, Klick auf die Toast-Box -> verschwindet.
6. **Backlog (`b`):** Wheel-Down bewegt Cursor von Backlog Task auf
   Backlog Task Two, Klick zurück auf Backlog Task -> Cursor folgt.
7. **Review-Cockpit (`R`):** Queue zeigt "1 of 1" + Epic-Header-Zeile +
   die eine to-review-Bean-Zeile (Task Two). Klick auf die Epic-Header-
   Zeile -> No-op (Cursor bleibt auf der Bean-Zeile, kein Crash) --
   bestätigt den Header-Skip in `reviewClickRow`. Wheel-Down bei nur
   einem Eintrag -> geklemmt, kein Crash.
8. `q` zurück nach Browse, `y` erneut erfolgreich (App unversehrt nach
   allen Mausereignissen). Kein Panic/Goroutine-Dump in der gesamten
   Sitzung (per Grep gegen den kompletten Pane-Capture geprüft).

NICHT live gesmoked (bewusst, siehe Deviations): Toast-Klick-Dismiss
BEI OFFENEM FORMULAR (Cross-Feature-Fix) -- über echte Tastatur nicht
praktikabel erreichbar (das offene huh-Formular fängt `y` als
Texteingabe ab, kein natürlicher Weg, einen Toast während eines
Formulars auszulösen). Stattdessen Update-Level abgedeckt:
TestToastClickDismissesEvenWithFormOpen geht über den vollen
`m.Update()`-Dispatcher (nicht nur `handleMouse`) und ist
mutationsverifiziert (siehe Test-Output).

## Deviations/ERRATA

- **Wheel-Schrittweite ±1 statt "1/3 Zeilen"** -- die Akzeptanz-
  Checkliste (Bean-Body) sagt wörtlich "bewegt den aktiven View-Cursor
  um 1/3 Zeilen", der Plan (epic-E5-plan.md »Task 4« Step 4.3 + design
  decision f) sagt dagegen unmissverständlich: Wheel folgt derselben
  Mechanik wie ein einzelner i/k-Tastendruck (kein `m.scroll`-Feld,
  kein Multiplikator). "1/3" ist mit hoher Wahrscheinlichkeit ein
  Formatierungsartefakt (verschmolzene Listen-Nummer "3." aus dem Plan-
  Step-Text mit dem folgenden Wort) -- devds eigener Tree-Wheel-
  Precedent (`update.go:476-487`) bewegt ebenfalls nur ±1 (der ×3-
  Multiplikator gilt dort NUR für `m.scroll`, das hier gar nicht
  existiert). Entscheidung: Plan + Design-Entscheidung f als
  maßgeblich behandelt (Quelle #3 laut Auftrag, vor der Bean-Prosa
  gelesen), ±1 implementiert. Kein Bug, aber PFLICHT-Deviation-Eintrag
  wegen widersprüchlicher Quellen.
- **`clickPaneGeometry` UND `browseRepoChrome`/`backlogChrome`/
  `reviewCockpitChrome` sind zusätzlicher, ÜBER den Plan-Wortlaut
  hinausgehender Extract** -- der Plan verlangt nur, dass
  `treeClickRow` "die Formel exakt rekonstruiert" (impliziert:
  Duplikat mit Kommentar-Pflicht, wie `windowStart` selbst schon
  vorexerziert). Gewählt wurde stattdessen ein ECHTER Extract (eine
  Funktion, von Render UND Klick genutzt) statt einer dritten
  Kopie -- konsequenter gegen Golden-Rule-Drift, aber ein zusätzlicher
  Diff-Fußabdruck in allen drei View-Dateien. Abgesichert durch alle 7
  Goldens `-count=2` byte-identisch (reiner Extract, keine
  Verhaltensänderung) -- kein Bug, aber PFLICHT-Deviation-Eintrag, weil
  über den wörtlichen Plan-Umfang hinausgehend.
- **`reviewRowRef` ist ein privater, zweiter Zähl-Durchlauf über
  `rs.groups`/`rs.rework`**, NICHT in `reviewQueueRows` selbst
  zurückgeführt (die hätte man ebenfalls extrahieren können). Bewusst
  NICHT gemacht: `reviewQueueRows` ist golden-getestet und
  Text-Rendering-lastig -- ein Refactor dort hätte Golden-Risiko ohne
  Klick-Nutzen gehabt. Der Zähl-Durchlauf in `reviewClickRow` ist als
  Drift-Risiko im Funktions-Kommentar explizit dokumentiert
  (Kommentar-Pflicht analog windowStart).
- Kein neues Golden (Maus berührt keinen Default-Render-Pfad,
  bewiesen).

## Notes for T5 (Settings)

- **Keine neuen config.yaml/state.json-relevanten Felder aus T4.**
  `lastClickIdx`/`lastClickAt`/`clock` sind reiner In-Memory-Session-
  State (Doppelklick-Fenster), kein Kandidat für Persistenz -- anders
  als z.B. `m.backlogSort` (E2 T5) oder ein künftiges Fenster-Layout.
- **`doubleClickInterval` (500ms) ist eine Konstante, kein
  Settings-Feld.** Falls T5 eine "Maus-Empfindlichkeit"-Einstellung
  vorsieht (design-spec.md prüfen), wäre dies der Injektionspunkt --
  aktuell hart kodiert, Port-Parität mit devd.
- **Editor-Präzedenz (Settings > $VISUAL > $EDITOR > vi):** T4 hat KEINE
  Editor-Berührung (kein Config-relevanter Fund hier) -- der Editor-
  Suspend-Pfad (`editorTarget`/`editorETag`, E3 T5) ist mausunabhängig,
  keine neue Interaktion zwischen T4 und dem $EDITOR-Pfad.
- **`clickPaneGeometry`/`browseRepoChrome`/`backlogChrome`/
  `reviewCockpitChrome` sind die neue Single Source für Pane-Geometrie**
  -- falls T6 (Lobby V1 + Repo-Picker) eine vierte View mit eigenem
  Master-Detail-Layout einführt, dort denselben Extract-Precedent
  fortsetzen (eigene `<view>Chrome`-Funktion + `clickPaneGeometry`-
  Aufruf), statt eine vierte unabhängige Geometrie-Kopie.
- **`m.view == viewLobby` fehlt noch im Overlay-Guard** (`mouse.go`
  `handleMouse`, Kommentar dort verweist bereits hierher) -- die Lobby-
  View existiert vor T6 nicht (`viewID`-Enum, types.go, hat aktuell nur
  `viewBrowseRepo`/`viewBacklog`/`viewReviewCockpit`). T6 MUSS beim
  Anlegen von `viewLobby` diesen Guard-Zweig ergänzen (`|| m.view ==
  viewLobby`), sonst würde ein Mausklick in der Lobby (falls sie
  irgendeine tastaturgesteuerte Overlay-Fläche nutzt) fälschlich an
  Browse/Backlog/Review-Cockpit-Dispatch durchgereicht -- aktuell
  harmlos (kein 4. `switch m.view`-Case in `handleMouse`/`wheelMove`
  -> No-op), aber die Guard-Lücke sollte trotzdem geschlossen werden,
  sobald die View existiert.
