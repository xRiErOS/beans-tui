---
# bt-mne6
title: E5 T4 — Maus (Wheel/Klick/Doppelklick)
status: todo
type: task
created_at: 2026-07-15T09:04:24Z
updated_at: 2026-07-15T09:04:24Z
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
- [ ] internal/tui/update.go Update(): neuer `case tea.MouseMsg: return
      m.handleMouse(msg)` -- VOR dem `if m.form != nil { return
      m.updateForm(msg) }`-Fallback (Port devd Cross-Feature-Fix, Toast-
      Klick darf nicht von einem offenen Formular geschluckt werden)
- [ ] internal/tui/mouse.go (NEU): handleMouse(msg tea.MouseMsg) -- Toast-
      Hit-Test zuerst (m.toastHit -> dismissToast), dann Overlay-Guard
      (Formulare/Picker/Menüs/Palette/Suche/Filter ignorieren Maus,
      Precedent devd), dann Wheel (Up/Down bewegt den aktiven View-Cursor
      um 1/3 Zeilen -- gleiche Richtung wie i/k), dann Linksklick-Dispatch
      je m.view (viewBrowseRepo/viewBacklog/viewReviewCockpit)
- [ ] internal/tui/view_browse_repo.go: treeClickRow(m model, nodes
      []treeNode, msg tea.MouseMsg) (idx int, ok bool) -- rekonstruiert
      lw/rw/bodyH IDENTISCH zu viewBrowseRepo()s eigener Render-Formel
      (Golden-Rule-Drift-Schutz, Kommentar-Pflicht analog windowStart),
      -1 Offset für die Suchkopfzeile, windowStart(len(nodes), bodyH-3,
      cursorPos) für die Fensterung; Doppelklick (500ms-Fenster, Port devd
      doubleClickInterval -- NEUE Model-Felder lastClickIdx/lastClickAt)
      togglet expand/collapse auf demselben Knoten, Einzelklick setzt nur
      Cursor (devd D03-Semantik: Einzelklick auf offenen Knoten toggelt
      NICHT)
- [ ] internal/tui/view_browse_backlog.go: backlogClickRow -- analoge
      Click-Y->Row-Funktion für backlogList (kein Doppelklick-Sinn, flache
      Liste)
- [ ] internal/tui/view_review_cockpit.go: reviewClickRow -- analog für
      reviewFlat/reviewCursor
- [ ] internal/tui/app.go: Run() bleibt bei tea.WithMouseCellMotion()
      (bereits gesetzt seit T8) -- KEINE Änderung nötig, nur Doku-Update
      im Kommentar (Mausteil jetzt tatsächlich verdrahtet)
- [ ] mouse_test.go (NEU): TestWheelUpDownMovesTreeCursor,
      TestWheelMovesBacklogCursor, TestWheelMovesReviewCursor,
      TestClickSetsTreeCursor, TestDoubleClickTogglesExpand,
      TestSingleClickOnOpenNodeDoesNotCollapse, TestMouseIgnoredWhileFormOpen,
      TestMouseIgnoredWhileOverlayOpen, TestToastClickDismissesEvenWithFormOpen
      (Cross-Feature-Fix-Regression, Port devd's eigener Grund für die
      Reihenfolge)
- [ ] `command go test ./... -short` grün, gofmt/vet leer, Goldens 2x grün
      (Maus berührt keinen Default-Render-Pfad)
- [ ] Commit `feat(tui): Maus (Wheel/Klick/Doppelklick, Toast-Dismiss-Vorrang)`
