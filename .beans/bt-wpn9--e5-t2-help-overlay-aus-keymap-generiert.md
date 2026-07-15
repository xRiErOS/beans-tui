---
# bt-wpn9
title: E5 T2 — Help-Overlay ? (aus Keymap generiert)
status: completed
type: task
priority: normal
created_at: 2026-07-15T09:04:24Z
updated_at: 2026-07-15T10:11:33Z
parent: bt-5h4d
---

Ziel: Help-Overlay `?` (Port devd `overlay_shortcuts.go`), generiert aus
`keys.helpGroups()` -- Single Source, kein Drift zur realen Keymap. KEINE
neuen keyMap-Felder nötig (`Help` existiert bereits seit E1 Task 7 und ist
bereits im Drift-Guard-Test `TestHelpGroupsCoverEveryBindingExactlyOnce`
abgedeckt) -- reine Verdrahtung von Overlay-State + Rendering + Capture.

Plan: docs/plans/v1-port/epic-E5-plan.md »Task 2«.

## Akzeptanz
- [x] internal/tui/overlay_shortcuts.go (NEU): helpBox() -- Port devd
      VERBATIM (modalPanel-basiert, Spaltenbreite über alle Gruppen
      global bestimmt)
- [x] internal/tui/types.go: model.helpOpen bool (neues Feld)
- [x] internal/tui/update.go handleKey: `m.helpOpen`-Vollcapture-Block
      (Precedent filterOpen/paletteOpen) direkt VOR dem `ctrl+c/q/tab`-
      Switch; `?` (keys.Help) öffnet von ÜBERALL (wie ctrl+k, design
      decision h) -- Case oberhalb des Review-Cockpit-Capture-Blocks
- [x] internal/tui/keyHelp(msg) (NEU, in overlay_shortcuts.go oder
      update.go): esc/?/q schließen (Port devd Footer-Hinweis "esc/?/q:
      close" wörtlich)
- [x] internal/tui/view_browse_repo.go composeOverlays: neuer Help-Case
      NACH m.paletteOpen, VOR m.confirmQuit (Painter's-Algorithmus-Reihen-
      folge -- Quit bleibt oberste Priorität, Port devd viewComposite-
      Reihenfolge confirmQuit > helpOpen)
- [x] overlay_shortcuts_test.go: TestHelpBoxRendersEveryGroup,
      TestKeyHelpOpensFromAnyView, TestKeyHelpEscQCloseHelp,
      TestHelpCapturesSingleKeysWhileOpen (q darf NICHT quit-confirm
      triggern während Help offen ist)
- [x] `command go test ./... -short` grün, gofmt/vet leer, TestKeymapNoCtrlSQ
      + TestHelpGroupsCoverEveryBindingExactlyOnce weiterhin grün (Drift-
      Guard unangetastet, da keine neuen Bindings)
- [x] Commit `feat(tui): Help-Overlay ? aus zentraler Keymap generiert`


## Prelude aus E5-T1-Review (PFLICHT, vor Help-Arbeit)
- [x] B01: update.go:244 (applyCreateDone reopen-Branch) — showToast ergänzt (toastError non-sticky, Muster der 8 anderen Sites) + Mini-Test (TestCreateConfirmEnterErrorPreservesDraftAndReopensForm erweitert, RED->GREEN bewiesen)
- [x] I01: bean bt-6dts Deviations-Sektion korrigiert (surgical ## Korrektur-Append: applyLoaded ergänzt, Zählung auf 8 Stellen/6 Funktionen richtiggestellt)

## Summary

internal/tui/overlay_shortcuts.go (NEU, ~75 Zeilen): Port devd
`overlay_shortcuts.go`s `helpBox()` VERBATIM -- modalPanel-basiert
(bestehende modal.go-Infra, wie alle anderen Overlays), Tasten-Label-
Spaltenbreite GLOBAL über alle `keys.helpGroups()`-Gruppen bestimmt
(lipgloss.Width, ANSI-sicher), Titel "Keyboard shortcuts", Footer
"esc/?/q: close" wörtlich, `clampModalWidth(54, m.width)` wie devd.
devds Geschwister-Funktion `shortcutMarkdown()` (externe docs/
shortcuts.md-Generierung, DD2-5) bewusst NICHT portiert -- beans-tui
hat keine externe Shortcut-Doku (YAGNI, im Datei-Doc-Kommentar
begründet). `keyHelp(msg)`: esc/?/q schließen, ALLE anderen Keys no-op
(volle Capture) -- bewusste Deviation zu devd, wo JEDE Taste schließt
(devd TestHelpOverlayToggle: "beliebige Taste sollte Hilfe schließen");
beans-tuis eigene Akzeptanz verlangt explizit Capture-Semantik
(TestHelpCapturesSingleKeysWhileOpen: q darf quit-confirm NICHT
triggern).

internal/tui/types.go: `helpOpen bool` (Doc-Kommentar mit
filterOpen/paletteOpen-Precedent + Deviation-Hinweis).

internal/tui/update.go handleKey: `if m.helpOpen { return m.keyHelp(msg) }`
NACH `m.paletteOpen` (ERRATUM, s. Deviations), `if keybind.Matches(msg,
keys.Help) { m.helpOpen = true }` nach `keys.Palette`, VOR dem
Review-Cockpit-Capture-Block (design decision h: von überall, auch im
Cockpit -- tmux-Smoke bestätigt).

internal/tui/view_browse_repo.go composeOverlays: Help-Case nach
`m.paletteOpen`, vor `m.confirmQuit` (Painter's-Reihenfolge, Quit bleibt
topmost) -- von allen drei Views geteilt, daher überall aktiv.

Keine keymap.go-Änderung: `keys.Help` existiert seit E1 T7, Drift-Guard
`TestHelpGroupsCoverEveryBindingExactlyOnce` + `TestKeymapNoCtrlSQ`
unangetastet grün (explizit einzeln gelaufen, Step 6).

## Test-Output

RED bewiesen: overlay_shortcuts_test.go zuerst geschrieben (5 Tests:
TestHelpBoxRendersEveryGroup, TestKeyHelpOpensFromAnyView,
TestKeyHelpEscQCloseHelp, TestHelpCapturesSingleKeysWhileOpen,
TestKeyHelpUnreachableWhilePaletteOpen [Zusatz, ERRATUM-Guard]) ->
Compile-Fail (`m.helpBox undefined`). Implementierung -> alle 5 grün.

Prelude-B01 ebenfalls RED->GREEN: TestCreateConfirmEnterErrorPreserves
DraftAndReopensForm um Toast-Assertion erweitert -> FAIL ("toast =
<nil>, want a non-nil toastError") gegen unveränderten Code -> Fix in
applyCreateDone -> PASS (+ alle 7 TestCreateDone/-Confirm-Tests grün).

`command go test ./... -count=1` 2x grün (internal/tui 127.0s/126.8s),
`-race -count=1` grün (128.8s), `go test ./... -short` grün, gofmt leer,
`go vet ./...` leer, `go build -o bin/bt .` ok. Alle 7 Goldens
(`TestChromeGolden`/`TestTreeGolden`/`TestTreeGoldenDeterministic`/
`TestBacklogGolden`/`TestBacklogGoldenDeterministic`/
`TestReviewCockpitGolden`/`TestReviewCockpitGoldenDeterministic`)
`-count=2` grün, testdata/ byte-identisch (git status leer) -- Help ist
bei helpOpen=false strukturell inert. Kein neues Golden: der Plan
fordert keins für Help (Step 1-8 geprüft).

## Smoke (tmux 100x30, Scratch-Repo, 1 Milestone/1 Epic/2 Tasks, 1x to-review)

1. **Browse:** `?` -> Overlay zentriert über Tree/Detail, alle 3
   Gruppen-Titel + Bindings sichtbar, Basis-View ringsum erhalten.
   `esc` -> zu, Tree/Detail unversehrt.
2. **Capture:** `?` -> `c` getippt: KEIN Create-Form, Overlay bleibt.
   `q` getippt: Help zu, KEIN Quit-Confirm, Browse unversehrt.
3. **Backlog:** `b` -> `?` -> Overlay; zweites `?` -> zu, weiterhin
   Backlog.
4. **Review-Cockpit:** `R` -> `?` öffnet AUS dem Cockpit heraus (Beleg
   design decision h: Help-Check sitzt über dem Cockpit-Capture-Block)
   -> `esc` -> zu, Review-Queue ("1 of 1") unversehrt.

## Deviations

- **ERRATUM (Capture-Order vs. Epic-Plan-Sketch):** der "Geteilte
  Infrastruktur"-Sketch im Epic-Plan listet `m.helpOpen`/`keys.Help`
  VOR `m.paletteOpen`/`keys.Palette`. Das ist falsch herum: ein im
  offenen Command-Center getipptes `?` (legitimes Query-Zeichen) würde
  Help öffnen statt keyPalette zu erreichen. Implementiert als
  paletteOpen -> helpOpen -> keys.Palette -> keys.Help (jeder
  Vollcapture-STATE vor jedem nackten Keybind-MATCH, exakt das
  bestehende handleKey-Muster). Eigener Regressionstest
  TestKeyHelpUnreachableWhilePaletteOpen. Die Task-2-Sektion selbst
  ("direkt vor dem keybind.Matches(msg, keys.Palette)-Block") ist mit
  der Implementierung konsistent -- nur der Infrastruktur-Sketch weicht
  ab. Plan-Datei NICHT editiert (ERRATUM-Präzedenz: hier dokumentiert).
- **Kein Scroll bei Überlänge (bewusster Scope-Halt, KEIN Bug):** die
  Modal-Box ist ~37 Zeilen hoch; auf Terminals <37 Zeilen clippt
  placeOverlay unten (Footer unsichtbar, esc/?/q funktionieren
  trotzdem). Port-Treue: devds helpBox ist mit 4 Gruppen (~45 Zeilen)
  noch höher und clippt identisch -- devd hat ebenfalls keinen
  Scroll-Mechanismus. Weder Bean-Akzeptanz noch Plan-Task-2-Steps
  fordern Scroll; die "Geteilte Infrastruktur"-Feldliste sieht für
  Task 2 NUR `helpOpen bool` vor (kein Scroll-Offset-Feld). Falls
  gewünscht: Follow-up-Task (windowAround-Reuse wäre der Ansatz).
- devds `shortcutMarkdown()`/TestShortcutDoc nicht portiert (keine
  externe Shortcut-Doku in beans-tui, YAGNI -- s. Summary).
- devds "jede Taste schließt"-Semantik nicht portiert (Akzeptanz
  verlangt Vollcapture -- s. Summary).

## Notes für T3 (Yank y, bt-e4a6)

- `keys.Yank` ist bereits im Keymap + helpGroups ("y / Copy context",
  Actions-Gruppe) -- T3 braucht KEINE keymap.go-Änderung, Drift-Guard
  bleibt unangetastet (gleiche Lage wie Help in T2).
- Capture-Order-Hinweis: `y` läuft als keyNodeAction-Case (Plan Step 6)
  -- der sitzt NACH allen Vollcapture-Blöcken, ein offenes
  Help-Overlay schluckt `y` also automatisch korrekt (kein Sonderfall
  nötig). Das Review-Cockpit-Override (Plan Step 7) muss in
  keyReviewCockpit landen, das VOR keyNodeAction dispatcht wird --
  exakt wie im Plan beschrieben, keine neue Recherche nötig.
- Toast-Bestätigung ("Kopiert: <ID>") nutzt showToast(toastInfo, ...)
  -- Signatur `(m model) showToast(kind, title, context string ->
  target *toastTarget, sticky bool) (model, tea.Cmd)`
  (overlay_show_toast.go:98), Rückgabe-Muster `m, toastCmd = ...` wie
  an den jetzt 9 bestehenden Sites.
- B01-Nachtrag aus diesem Prelude: applyCreateDones Reopen-Branch ist
  jetzt die 9. showToast-Site -- falls T3s Orphan-No-op-Test gegen
  "keine Toast-Site vergessen" grept, den Branch mitzählen.
