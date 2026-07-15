---
# bt-6dts
title: E5 T1 — Toast-System (Port overlay_show_toast.go) + Fehler-Toasts
status: todo
type: task
created_at: 2026-07-15T09:04:24Z
updated_at: 2026-07-15T09:04:24Z
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
- [ ] internal/tui/overlay_show_toast.go (NEU): toastKind (info/warn/error),
      toastTarget{view viewID}, toastState{kind/title/context/target/seq/
      sticky/setAt}, toastDebounceWindow (300ms), toastDuration(kind),
      showToast/clearToastUnlessSticky/toastKindColor/toastBox/
      toastGeometry/renderToast/toastHit/dismissToast -- Port devd VERBATIM,
      angepasst an beans-tui's eigene Chrome/outerBorder-Geometrie (kein
      m.termWidth()-Methodenaufruf -- m.width/m.height direkt, da
      view_browse_repo.go schon fertig kompositierte Frames auf m.width/
      m.height baut)
- [ ] internal/tui/messages.go: toastExpiredMsg{seq int}, toastTimeout(seq,
      kind) tea.Cmd (Port devd messages.go VERBATIM)
- [ ] internal/tui/types.go: model.toast *toastState (neues Feld)
- [ ] internal/tui/update.go: `case toastExpiredMsg: return
      m.handleToastExpired(msg)` im Update-Dispatcher; JEDE bestehende
      `m.err = ...`-Zuweisung (applyMutationResult/applyBleveResult/
      applyPaletteBleveResult/applyEditorFinished/applyLoaded/
      createInFlightNote-Guard) bekommt zusätzlich einen showToast-Aufruf
      (Kind error/warn passend zur Schwere, sticky NUR beim
      data.ErrConflict-Zweig in applyMutationResult) -- gebatched via
      tea.Batch mit dem bestehenden Reload-/Load-Cmd
- [ ] internal/tui/view_browse_repo.go: `View()`-Dispatcher wrappt das
      Ergebnis JEDER der 3 Sub-Views mit `m.renderToast(out)` NACH
      composeOverlays (Toast schwebt über ALLEM, auch Modals -- Port devd
      view.go's `View()`/`viewComposite()`-Trennung)
- [ ] internal/tui/update.go Update(): tea.MouseMsg-Vorrang für Toast-Klick-
      Dismiss VOR einer etwaigen `m.form != nil`-Kurzschluss-Rückgabe (Port
      devd Cross-Feature-Fix, DD2-272/273) -- Grundlage für Task 4 (Maus),
      hier schon als Kommentar/Stub-Case vorbereitet
- [ ] overlay_show_toast_test.go: TestShowToastSingleSlotReplacesPrevious,
      TestShowToastDebounceWindowUpdatesInPlace, TestStickyToastNoAutoDismiss,
      TestClearToastUnlessSticky, TestToastGeometryTopRight,
      TestToastHitTest, TestDismissToastJumpsToTarget,
      TestConflictToastIsStickyAndSurvivesReload (E3-Übernahme-Kern-Test)
- [ ] `command go test ./... -short` grün, gofmt/vet leer
- [ ] Commit `feat(tui): Toast-System (Port overlay_show_toast.go) + Konflikt-sticky`
