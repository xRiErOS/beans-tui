---
# bt-hd42
title: Klick trifft nach Scrollen falsches Feld (Offset fehlt)
status: completed
type: bug
priority: high
created_at: 2026-07-20T09:24:15Z
updated_at: 2026-07-20T09:55:38Z
parent: bt-vy1q
---

**Bug, gefunden von bt-1o4g (nicht dort verursacht).**

`detailBoxFormClickRow` ignoriert `boxFormScroll`. Nach dem Scrollen der Detail-Pane
treffen Mausklicks daher die **ungescrollten** Feldzonen — man klickt auf Status und
oeffnet Type.

## Herkunft
Entstanden aus der Kombination von bt-ze10 (Scroll-Offset eingefuehrt) und der frueheren
Maus-Slice S6 (Trefferzonen ohne Offset gerechnet). Beide fuer sich korrekt, zusammen
falsch. Von bt-1o4g weder verursacht noch verschlimmert, dort nur entdeckt.

## Warum das dringlich ist
Es ist nicht nur ein eigener Fehler, es ist ein **Fundament, auf dem gerade weitere
Features geplant werden**: Pencil-Klick auf Body-Ueberschriften und die Anker-Leiste
brauchen ebenfalls scroll-korrigierte Trefferzonen. Wird das hier nicht repariert, erben
sie den Fehler.

## Betroffen
`internal/tui/mouse.go` — `detailBoxFormClickRow`, `boxFormFieldAt`, `gridColAt`.
Der Offset liegt in `boxFormScroll`/`boxFormEffectiveScroll` (types.go, bt-ze10) bereit.

## Akzeptanz
- [ ] Klick trifft nach dem Scrollen das richtige Feld
- [ ] Test, der ERST scrollt und DANN klickt (der fehlende Fall)
- [ ] Mausrad und Tastatur-Scroll fuehren zum selben Ergebnis
- [ ] voller Testlauf gruen


## Summary

`detailBoxFormClickRow` (internal/tui/mouse.go) mappte eine **Screen**-Zeile auf
`boxFormFieldAt`s **Content**-Zeilen-Tabelle. Beide fallen nur bei Scroll-Offset 0
zusammen — genau dort sassen alle bestehenden Klick-Tests, deshalb war die ganze
Fehlerklasse unsichtbar.

Fix (eine Zeile): `boxFormFieldAt(clickRow + boxFormEffectiveScroll(m, m.focusedBean()), ...)`.
Gelesen wird bewusst ueber `boxFormEffectiveScroll`, nicht ueber `m.boxFormScroll` —
derselbe Lesepfad wie der Render (renderBeanAccordionPane), damit bt-ze10s
abgeleiteter Per-Bean-Reset mitgilt. Ein Roh-Lesen haette denselben Bug in
Richtung Bean-Wechsel neu aufgemacht.

`gridColAt`/`boxFormFieldAt` brauchten keine Aenderung: der Fehler war rein
vertikal, die Spalten-Bucketierung ist offset-unabhaengig.

### Vollbild-Klickpfad: nicht erreichbar (geprueft, nicht angenommen)

`handleMouse` (mouse.go:161) gibt fuer `m.fullscreen != fullscreenNone`
unbedingt `return m, nil` zurueck — vor Wheel- UND Klick-Dispatch (F01, bean
bt-13l7, dokumentierter v1-Scope-Cut "Maus im Vollbild"). Jeder Aufrufer von
`detailBoxFormClickRow` liegt hinter diesem Guard. Ein `accW`-Vollbild-Zweig
waere Code fuer einen toten Pfad — stattdessen als Doc-Kommentar festgehalten,
inkl. Hinweis, dass beim Aufheben von F01 `boxFormPaneMetrics`' `innerW-4`-Zweig
in derselben Aenderung nachzuziehen ist.

## Test-Output

Neue Datei `internal/tui/mouse_boxform_scroll_test.go` — vier Tests, alle ueber
echte `tea.MouseMsg`/`tea.KeyMsg`-Roundtrips und render-geerdete Koordinaten
(`boxFormClickAt` sucht das Badge in der ECHTEN `View()`).

RED (vor dem Fix):

    === RUN   TestBoxFormClickOnStatusBoxAfterScrollOpensStatus
        overlay after clicking the Status box at scroll offset 3 = 0, want overlayValueMenu
    --- FAIL: TestBoxFormClickOnStatusBoxAfterScrollOpensStatus (0.01s)
    === RUN   TestBoxFormClickOnTagsBoxAfterScrollOpensTagPicker
        overlay after clicking the Tags box at scroll offset 3 = 1, want overlayTagPicker
    --- FAIL: TestBoxFormClickOnTagsBoxAfterScrollOpensTagPicker (0.01s)
    === RUN   TestBoxFormClickUnaffectedAtZeroScroll
    --- PASS: TestBoxFormClickUnaffectedAtZeroScroll (0.00s)
    === RUN   TestBoxFormClickSameAfterWheelAndKeyboardScroll
    --- PASS: TestBoxFormClickSameAfterWheelAndKeyboardScroll (0.04s)
    FAIL

`overlay = 1` im Tags-Fall ist woertlich das gemeldete Symptom: Klick auf Tags
oeffnete das Value-Menue statt des Tag-Pickers.

GREEN (nach dem Fix), inkl. der drei Bestands-Klick-Tests:

    --- PASS: TestBoxFormClickOnStatusBoxAfterScrollOpensStatus (0.02s)
    --- PASS: TestBoxFormClickOnTagsBoxAfterScrollOpensTagPicker (0.02s)
    --- PASS: TestBoxFormClickUnaffectedAtZeroScroll (0.01s)
    --- PASS: TestBoxFormClickSameAfterWheelAndKeyboardScroll (0.08s)
    --- PASS: TestBoxFormClickOnStatusBoxOpensValueMenuSeededToStatus (0.00s)
    --- PASS: TestBoxFormClickOnTypeBoxOpensValueMenuSeededToType (0.00s)
    --- PASS: TestBoxFormClickOnTagsBoxOpensTagPicker (0.00s)
    PASS

Voller Lauf (`command go test ./...`, ohne `-short`):

    ?   	github.com/xRiErOS/beans-tui	[no test files]
    ok  	github.com/xRiErOS/beans-tui/cmd	0.466s
    ?   	github.com/xRiErOS/beans-tui/internal/clip	[no test files]
    ok  	github.com/xRiErOS/beans-tui/internal/config	1.850s
    ok  	github.com/xRiErOS/beans-tui/internal/data	4.817s
    ok  	github.com/xRiErOS/beans-tui/internal/theme	1.480s
    ok  	github.com/xRiErOS/beans-tui/internal/tui	151.778s

Kein Golden beruehrt (Trefferzonen sind unsichtbar) — Diff: `internal/tui/mouse.go`
+ die neue Testdatei.

## Deviations

1. **Worktree-Basis war falsch** (`main`, ~45 Commits hinter dem Stand; der
   reparierte Code existierte dort nicht). Vor der ersten Aenderung per
   `git reset --hard experiment/jira-style-ui` korrigiert.
2. **Tastatur/Wheel-Aequivalenz als Invarianten-Guard, nicht als RED-Test.**
   `boxFormNav` koppelt Scroll an den Feld-Cursor und snappt ein hohes Feld in
   den View — ein *beliebiger* Ziel-Offset ist per Tastatur nicht ansteuerbar.
   Der Test nimmt daher den Offset, den die Tastatur erreicht, spielt exakt so
   viele Wheel-Ticks nach und vergleicht den Hit-Test ueber die GANZE Pane
   (jede Zeile x fuenf Spalten). Er war schon vor dem Fix gruen — beide Pfade
   schreiben dasselbe Feld-Paar, waren also konsistent falsch. Er pinnt die
   Eigenschaft fuer die Zukunft (zweiter Offset / vergessener Bean-Owner).
3. **Kein `accW`-Vollbild-Zweig** — begruendet oben (toter Pfad).
