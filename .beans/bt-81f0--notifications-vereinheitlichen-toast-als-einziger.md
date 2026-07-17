---
# bt-81f0
title: 'Notifications vereinheitlichen: Toast als einziger Kanal'
status: in-progress
type: feature
priority: normal
created_at: 2026-07-17T06:27:18Z
updated_at: 2026-07-17T07:38:15Z
parent: bt-5uzr
---

NB aus PO-Review E11 (2026-07-17, Runde 3), PO verbatim:

"Wir haben 2 Systeme, um Notifications zu senden: 1) oben rechts das Toast und
2) unten rechts bspw der Error 'beans list: exit status 1: Error: querying
beans: syntax Error' Das ist nicht ok. Der kanonische Weg der Notifications ist
der Toast oben rechts. Dadurch wird auch unten die reservierte Zeile für die
Notifications frei."

Interpretation: PO-Design-Entscheidung — Toast (oben rechts) ist der EINZIGE
Notification-Kanal. Die untere Fehler-/Notification-Zeile entfällt; alle
bisherigen Konsumenten (z. B. Datenlayer-Fehler wie "beans list: exit status 1")
routen auf den Toast (Fehler-Variante toastErr, ohne Auto-Dismiss bei Fehlern?
— Planner klärt Persistenz-Verhalten). Frei werdende Zeile geht an den Content.

Akzeptanz-Entwurf:
- [ ] Inventar aller Nutzer der unteren Notification-/Error-Zeile (grep Render-Pfad)
- [ ] alle auf Toast umgeleitet (Fehler deutlich, nicht stumm verkürzt)
- [ ] untere reservierte Zeile entfernt, Layout-Höhe angepasst (Goldens!)
- [ ] Fehler bleiben lesbar bei langen Meldungen (Wrap/Truncation-Konzept im Toast)


## Plan-Konkretisierung E12 (2026-07-17)

Plan: `docs/plans/v1-port/epic-E12-plan.md` §„Item 2: Notifications
vereinheitlichen — Toast als einziger Kanal". Reihenfolge: Rang 2 (zusammen
mit `bt-9ipw` PO-mandiert zuerst, nach `bt-9ipw` wegen gemeinsamer
Datei-Nachbarschaft `box_picker_tag.go`).

**Root Cause:** zwei Feedback-Kanäle — Toast (`overlay_show_toast.go`) und
`m.err string`, gerendert über `statusBar(indicator, errNote, width)`
(`view.go:70-83`). PO-Zitiertes Beispiel ("beans list: exit status 1: Error:
querying beans: syntax Error") ist `applyBleveResult`s Fehlerpfad
(`update.go:852-856`).

**Vollständiges Inventar (11 Dateien):** 10 Stellen in `update.go`
(`applyMutationResult`, `applyBeanRawLoaded`, `applyCreateDone`,
`applyEditorFinished` ×2, `applyBleveResult`, `applyLoaded`) sind bereits
Dual-Write (Toast + `m.err`) — reine Redundanz, Toast bleibt, Rendering-
Anbindung von `m.err` entfällt. SIEBEN Stellen sind HEUTE NUR `m.err`, KEIN
Toast (würden bei blinder Zeilen-Entfernung komplett stumm):
`box_confirm_delete.go:130`, `box_confirm_create.go:48/66/91/96`,
`box_menu_value.go:191`, `box_picker_parent.go:142`,
`box_picker_blocking.go:161`, `box_picker_tag.go:425`,
`overlay_palette.go:232`.

**D02 (Scope-Begrenzung):** `m.err`-FELD bleibt bestehen (43 Testassertions
über 16 Testdateien referenzieren es direkt) — nur die Rendering-Anbindung
(`ChromeOpts.ErrNote`, `view.go:255`, plus die drei `statusBar(indicator,
m.err, innerW)`-Aufrufe in `view_browse_repo.go:954`,
`view_tag_management.go:272`, `view_browse_backlog.go:286`) entfällt.

**Offene Frage Q1 (Plan):** `statusBar` rendert auch den Scroll-Indikator
(Chrome/Lobby/Help) — reicht Kappen NUR der Fehler-Anbindung (Zeile bleibt
als Indikator-Slot), oder wird eine echte Zeilen-Entfernung erwartet?
Planner-Empfehlung: erstere Variante, außer PO bestätigt den größeren Umbau.

**Akzeptanz:**
- [ ] Alle 10 Dual-Write-Stellen: nur Rendering-Anbindung entfernt, Toast
      + `m.err`-Zuweisung bleiben (Tests laufen weiter)
- [ ] Alle 7 bisher-stummen Stellen bekommen einen neuen
      `showToast(toastError, ..., "", nil, false)`-Aufruf
- [ ] Kein bestehender `.err`-Assertion-Test bricht
- [ ] Golden-Regen falls Layout-Höhe sich ändert:
      `go test ./internal/tui/ -run TestChromeGolden -update`,
      `-run TestTreeGolden -update`, `-run TestBacklogGolden -update`
      (Diff sichten vor Commit)
- [ ] tmux-Smoke: Bleve-Syntax-Fehler zeigt NUR noch Toast


## Prelude (2026-07-17, Supervisor, aus bt-9ipw-Abschluss + Q1-Entscheid)

- **Zeilendrift:** `box_picker_tag.go` wurde durch bt-9ipw konsolidiert — die `m.err = "Bean no longer exists — selection discarded"`-Stelle liegt jetzt bei ~Zeile 409 (nicht 425 wie im E12-Plan-Inventar). Alte Zeilennummern des Plans generell gegen Ist-Code prüfen.
- **Q1-Annahme (Supervisor-Entscheid, PO-Antwort ausstehend):** Planner-Empfehlung Variante 1 umsetzen — NUR die Fehler-Anbindung von `statusBar`/`ErrNote` kappen, die Zeile bleibt als Scroll-Indikator-Slot bestehen. KEINE dynamische Footer-Höhe. Als dokumentierte Annahme im Abschluss festhalten; PO kann am Review-Gate widersprechen.
