---
# bt-6dts
title: E5 T1 — Toast-System (Port overlay_show_toast.go) + Fehler-Toasts
status: completed
type: task
priority: normal
created_at: 2026-07-15T09:04:24Z
updated_at: 2026-07-15T09:38:08Z
parent: bt-5h4d
---

Ziel: Toast-System (Port devd `overlay_show_toast.go`) als EIN Slot (kein
Stack), Auto-Dismiss kind-spezifisch (Error 8s/Warn 3s/Info 5s), sticky=true
für ErrConflict (E3-Übernahme, E3-I01: PFLICHT, kein 1-Frame-Flash mehr) --
sticky übersteht Reload-Zyklen bis Replace/Klick. Dual-Write: m.err (Chrome
Red-Slot) bleibt bestehen (Rückwärtskompatibilität ggü. E1-E4-Tests), Toast
kommt ZUSÄTZLICH an jeder bestehenden `m.err = ...`-Stelle.

Plan: docs/plans/v1-port/epic-E5-plan.md »Task 1«.

## Akzeptanz
- [x] internal/tui/overlay_show_toast.go (NEU): toastKind (info/warn/error),
      toastTarget{view viewID}, toastState{kind/title/context/target/seq/
      sticky/setAt}, toastDebounceWindow (300ms), toastDuration(kind),
      showToast/clearToastUnlessSticky/toastKindColor/toastBox/
      toastGeometry/renderToast/toastHit/dismissToast -- Port devd VERBATIM,
      angepasst an beans-tui's eigene Chrome/outerBorder-Geometrie (kein
      m.termWidth()-Methodenaufruf -- m.width/m.height direkt, da
      view_browse_repo.go schon fertig kompositierte Frames auf m.width/
      m.height baut)
- [x] internal/tui/messages.go: toastExpiredMsg{seq int}, toastTimeout(seq,
      kind) tea.Cmd (Port devd messages.go VERBATIM)
- [x] internal/tui/types.go: model.toast *toastState (neues Feld)
- [x] internal/tui/update.go: `case toastExpiredMsg: return
      m.handleToastExpired(msg)` im Update-Dispatcher; JEDE bestehende
      `m.err = ...`-Zuweisung (applyMutationResult/applyBleveResult/
      applyPaletteBleveResult/applyEditorFinished/applyLoaded/
      createInFlightNote-Guard) bekommt zusätzlich einen showToast-Aufruf
      (Kind error/warn passend zur Schwere, sticky NUR beim
      data.ErrConflict-Zweig in applyMutationResult) -- gebatched via
      tea.Batch mit dem bestehenden Reload-/Load-Cmd
- [x] internal/tui/view_browse_repo.go: `View()`-Dispatcher wrappt das
      Ergebnis JEDER der 3 Sub-Views mit `m.renderToast(out)` NACH
      composeOverlays (Toast schwebt über ALLEM, auch Modals -- Port devd
      view.go's `View()`/`viewComposite()`-Trennung)
- [x] internal/tui/update.go Update(): tea.MouseMsg-Vorrang für Toast-Klick-
      Dismiss VOR einer etwaigen `m.form != nil`-Kurzschluss-Rückgabe (Port
      devd Cross-Feature-Fix, DD2-272/273) -- Grundlage für Task 4 (Maus),
      hier schon als Kommentar/Stub-Case vorbereitet
- [x] overlay_show_toast_test.go: TestShowToastSingleSlotReplacesPrevious,
      TestShowToastDebounceWindowUpdatesInPlace, TestStickyToastNoAutoDismiss,
      TestClearToastUnlessSticky, TestToastGeometryTopRight,
      TestToastHitTest, TestDismissToastJumpsToTarget,
      TestConflictToastIsStickyAndSurvivesReload (E3-Übernahme-Kern-Test)
      + TestToastExpiredMsgClearsOnlyMatchingGeneration (Zusatz, deckt den
      Update()-Dispatch-Case selbst ab, nicht nur showToast/dismissToast)
- [x] `command go test ./... -short` grün, gofmt/vet leer (+ volle Suite 2x,
      -race, 7 Goldens 2x, `go build -o bin/bt .` -- alles grün)
- [x] Commit `feat(tui): Toast-System (Port overlay_show_toast.go) + Konflikt-sticky`

## Summary

internal/tui/overlay_show_toast.go (NEU, ~230 Zeilen): Port devd
`overlay_show_toast.go` strukturell VERBATIM -- toastKind/toastTarget/
toastState/toastDebounceWindow/toastDuration/showToast/
clearToastUnlessSticky/toastKindColor/toastBox/toastGeometry/renderToast/
toastHit/dismissToast, EIN Slot (`m.toast *toastState`), Auto-Dismiss
kind-spezifisch (Error 8s/Warn 3s/Info 5s via `messages.go`s
toastTimeout), 300ms-Debounce (zweiter showToast-Call <300ms aktualisiert
in-place statt neue Generation+Timer). Einzige strukturelle Abweichung:
`toastTarget{view viewID}` (devd hat milestoneID/sprintID/issueID -- kein
Äquivalent in beans-tuis EINER Entität `data.Bean`). Geometrie/Rendering
(`toastGeometry`/`renderToast`) sind 1:1-Port OHNE Anpassung -- devd
rechnet nach dessen eigenem B01-Fix bereits direkt gegen `m.width`/
`m.height` (der fertig kompositierte Frame), exakt wie beans-tuis drei
View-Funktionen ihn selbst bauen (`outerBorder(...,true)`), keine
`m.termWidth()`-Diskrepanz zu kompensieren.

internal/tui/messages.go: `toastExpiredMsg{seq int}` + `toastTimeout(seq,
kind) tea.Cmd` (tea.Tick-Wrapper), Port VERBATIM, neuer `"time"`-Import.

internal/tui/types.go: `model.toast *toastState`, dokumentiert als
EINZIGE Schreibstelle showToast/dismissToast/handleToastExpired (Grep-
Audit-Kommentar) -- kein anderer Call-Site setzt `m.toast` direkt.

internal/tui/update.go: `case toastExpiredMsg: return
m.handleToastExpired(msg)` im Update()-Dispatcher (neue kleine Methode,
löscht `m.toast` nur bei Generation-Match). Dual-Write-Audit (Step 5)
gebatched via `tea.Batch(toastCmd, <bestehender Cmd>)` an GENAU den 6 in
Bean-Akzeptanz + Task-1-Files-Liste benannten Stellen (siehe Deviations
unten zum Scope): `applyMutationResult` (Konflikt-Zweig -> `toastError`
sticky=true, Titel gekürzt auf "Konflikt: Bean extern geändert" [ohne "--
neu geladen"-Suffix, damit die Toast-Box nicht truncaten muss],
Recovery-Tempfile-Pfad als zweite Zeile wenn vorhanden; sonstiger
Fehler-Zweig -> `toastError` non-sticky, Titel = `m.err` VERBATIM),
`applyBleveResult`/`applyPaletteBleveResult` (Signatur auf `(tea.Model,
tea.Cmd)` erweitert, err-Zweig -> `toastError` non-sticky), `applyLoaded`
(gleiche Signatur-Erweiterung, Lade-Fehler -> `toastError` non-sticky),
`applyEditorFinished` (Prozess-Fehler -> `toastError`, "Bean nicht mehr
vorhanden"-Guard -> `toastWarn`, beide non-sticky), `keyNodeAction`s
`pendingCreate != nil`-Guard (`createInFlightNote` -> `toastWarn`
non-sticky). `m.err` selbst UNVERÄNDERT (Dual-Write, keine einzige
bestehende `m.err = ...`-Zeile inhaltlich angefasst).

internal/tui/view_browse_repo.go: `View()` baut jetzt `var out string`,
switcht wie vorher, wrappt das Ergebnis am Ende mit `m.renderToast(out)`
-- AUSSERHALB von `composeOverlays` (design decision a Punkt 2), damit der
Toast über confirmQuit/Formularen/Overlays schwebt, nicht nur über der
Basis-View.

## Test-Output

RED zuerst bewiesen: `overlay_show_toast.go` temporär aus dem Baum entfernt
(+ die 4 Modify-Dateien via `git stash`), `go test ./internal/tui/ -run
'TestShowToast|TestConflictToastIsStickyAndSurvivesReload'` -> Compile-Fail
(`m.showToast undefined`, `m.toast undefined`, `undefined: toastInfo` etc.,
11 Fehler). Danach Implementierung wiederhergestellt (`git stash pop`) --
GREEN: alle 9 neuen Toast-Tests einzeln grün.

`command go test ./... -count=1` 2x grün (cmd/data/theme/tui, tui ~128s je
Lauf). `command go test ./... -race -count=1` grün (keine Data-Races).
`command gofmt -l .` leer, `command go vet ./...` leer, `command go build
-o bin/bt .` ok. Alle 7 Goldens (`TestChromeGolden`, `TestTreeGolden`,
`TestTreeGoldenDeterministic`, `TestBacklogGolden`,
`TestBacklogGoldenDeterministic`, `TestReviewCockpitGolden`,
`TestReviewCockpitGoldenDeterministic`) mit `-count=2` grün, BYTE-IDENTISCH
(kein `-update` nötig, `git status` auf `testdata/` leer) -- bestätigt: der
Toast ist bei `m.toast == nil` (jeder bestehende Test-Fixture-Zustand)
strukturell inert, `renderToast` gibt `base` unverändert zurück.

Ein bestehender Test musste mechanisch angepasst werden:
`TestEditorFinishedTargetVanishedSurfacesError` (`editor_test.go`) prüfte
bisher `cmd == nil` für den "Bean nicht mehr vorhanden"-Zweig -- durch den
jetzt zusätzlich gefeuerten `toastWarn` ist `cmd` legitim NICHT mehr nil
(der Toast-Auto-Dismiss-`tea.Tick`). Assertion umgestellt auf `cmd != nil`
+ `nm.toast.kind == toastWarn && nm.toast.title == nm.err` (beweist
strukturell, dass der non-nil Cmd der Toast-Timer ist, keine Mutation --
`applyEditorFinished`s vanished-target-Zweig erreicht `mutateCmd` gar
nicht). >20 andere E1-E4-Tests, die gegen `m.err`-Stringinhalte
assertieren, blieben UNVERÄNDERT grün (Dual-Write-Versprechen der Plan-
Design-Entscheidung a eingehalten).

## Smoke (tmux, Scratch-Repo `bt-e5-t1-smoke`, 1 Milestone + 2 Tasks)

1. **Sticky-Konflikt:** Fake-`beans`-Wrapper (nur `update` abgefangen,
   liefert CONFLICT-JSON, alles andere -> echtes `beans`) vorne auf PATH.
   `s` auf einem Task -> Status "in-progress" gewählt -> `enter`. Toast
   erscheint oben rechts: "● Konflikt: Bean extern geändert" (rot). Nach
   9s Warten (> toastError-Dauer 8s) UND einem expliziten `ctrl+r`-Reload
   ist der Toast UNVERÄNDERT sichtbar -- Sticky-Kontrakt (E3-I01) doppelt
   bestätigt (übersteht sowohl den Auto-Dismiss-Timer als auch einen
   echten Reload-Zyklus).
2. **Non-sticky Auto-Dismiss:** analoger Fake-Wrapper, aber
   VALIDATION_ERROR (non-conflict) statt CONFLICT. Toast "● beans:
   VALIDATION_ERROR: bean…" erscheint. Nach 9s Warten OHNE weitere
   Interaktion ist der Toast verschwunden, Rahmen wieder volle Breite --
   Auto-Dismiss-Timer (tea.Tick) verifiziert end-to-end im echten
   Bubbletea-Runtime, nicht nur strukturell wie in den Unit-Tests.
3. **Erfolgspfad ohne Toast-Müll:** frische Session, NUR Leseaktionen
   (Cursor, Value-Menü öffnen+`esc` abbrechen, `ctrl+r`) gegen das
   UNVERÄNDERTE echte `beans`-Binary -- kein "●" im gesamten Pane-Dump,
   Rahmen durchgehend intakt.

## Deviations

- **Scope der Dual-Write-Stellen (KEIN Bug, Klarstellung):** die
  Design-Entscheidung-a-Prosa im Epic-Plan sagt allgemein "Jede
  bestehende `m.err = ...`-Stelle bekommt ... showToast" -- ein
  Codebase-weiter Grep (`grep -rn "m\.err = "`) findet aber 20
  Zuweisungen über 8 Dateien, während Task 1s eigene Files-Liste, Step 5
  UND dieser Beans eigene Akzeptanz-Liste alle drei UNABHÄNGIG voneinander
  exakt dieselben 6 Stellen benennen -- ALLE in `update.go`
  (`applyMutationResult` x2, `applyBleveResult`, `applyPaletteBleveResult`,
  `applyEditorFinished` x2, `keyNodeAction`s `createInFlightNote`-Guard).
  `box_confirm_delete.go`/`box_confirm_create.go`/`box_picker_blocking.go`/
  `box_picker_tag.go`/`box_menu_value.go`/`box_picker_parent.go`/
  `overlay_palette.go`/`view_review_cockpit.go` haben je 1-2 weitere
  `m.err = ...`-Stellen (durchweg "Bean nicht mehr vorhanden"/
  `createInFlightNote`-Varianten), die NICHT in Task 1s Modify-Liste
  stehen und hier bewusst NICHT angefasst wurden -- Entscheidung zugunsten
  der drei übereinstimmenden konkreten Quellen (Files-Liste/Step 5/
  Akzeptanz) gegen die allgemeinere Prosa. Kein Funktionsverlust: diese
  Stellen behalten ihr bestehendes `m.err`-Verhalten unverändert (Status
  quo ante), sie zeigen nur noch KEINEN zusätzlichen Toast. Empfehlung
  für einen Folge-Task (T2-Zuschnitt oder eigener Bugfix-Task), falls die
  vollständige Prosa-Abdeckung gewünscht ist -- mechanisch identisches
  Muster, ca. 8 weitere Call-Sites.
- `TestEditorFinishedTargetVanishedSurfacesError`-Anpassung (siehe
  Test-Output oben) -- mechanisch, keine Verhaltensänderung des
  eigentlichen Kontrakts ("kein doomed Mutation-Cmd"), nur dessen
  Beweisführung umgestellt (structural statt `cmd == nil`).
- Kein Golden-Update nötig (Goldens 2x byte-identisch, siehe Test-Output).
- Keine weiteren Abweichungen vom Plan-Pseudocode (Steps 1-11 1:1
  übernommen inkl. Testnamen/Signaturen).

## Notes für T2 (Help-Overlay `?`, bt-wpn9)

- `composeOverlays` (`view_browse_repo.go:582`) bleibt für T2s
  `m.helpOpen`-Case der richtige Ort (VOR `m.confirmQuit`, design decision
  a Punkt 2 bestätigt das nochmal explizit: der Toast liegt AUSSERHALB
  von `composeOverlays`, nicht als weiterer Case darin -- T2s Help-Box
  reiht sich stattdessen ganz normal ALS Case in `composeOverlays` ein,
  wie im Epic-Plan vorgesehen).
- Das Mouse-Vorrang-Stub in `update.go`s `Update()` (VOR `case
  tea.KeyMsg`) ist jetzt platziert und dokumentiert -- Task 4 muss dort
  nur noch `case tea.MouseMsg: return m.handleMouse(msg)` einsetzen, keine
  neue Recherche zur Positionierung nötig.
- `toastHit`/`dismissToast` sind fertig fürs Task-4-Wiring (`handleMouse`s
  allererster Klick-Check) -- unverändert wie im Plan vorgesehen, keine
  Anpassung in T1 nötig gewesen.
