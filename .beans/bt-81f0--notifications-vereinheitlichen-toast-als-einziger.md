---
# bt-81f0
title: 'Notifications vereinheitlichen: Toast als einziger Kanal'
status: completed
type: feature
priority: normal
created_at: 2026-07-17T06:27:18Z
updated_at: 2026-07-17T07:58:56Z
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

## Summary (2026-07-17, Implementer)

Toast ist jetzt der EINZIGE sichtbare Notification-Kanal. `m.err` bleibt als
Feld bestehen (43 Testassertions), verliert aber jede Rendering-Anbindung:
`statusBar` hat kein `errNote`-Parameter mehr, `ChromeOpts.ErrNote` ist
entfernt, die drei direkten `statusBar(indicator, m.err, innerW)`-Aufrufe
(view_browse_repo.go/view_tag_management.go/view_browse_backlog.go) lesen
`m.err` nicht mehr. Die Statuszeile bleibt als reiner Scroll-Indikator-Slot
bestehen (Q1-Annahme, kein Layout-Umbau, keine dynamische Footer-Höhe).

Alle Nur-`m.err`-Schreibstellen (7 Dateien, 10 individuelle `m.err =`-Zeilen)
bekamen einen neuen `showToast(toastError, ...)`-Aufruf direkt neben der
bestehenden `m.err =`-Zuweisung:
- box_confirm_delete.go:130
- box_confirm_create.go:48, :66, :91, :96 (vier Stellen in einer Datei)
- box_menu_value.go:191
- box_picker_parent.go:142
- box_picker_blocking.go:161
- box_picker_tag.go:424
- overlay_palette.go:231

Die zehn Dual-Write-Stellen in update.go (bereits Toast+m.err) sind
unverändert — sie verlieren nichts, ihre m.err-Zuweisung darf bleiben.

## Test-Output (RED → GREEN)

RED (vor Implementierung, 9 Tests, u.a.):
```
--- FAIL: TestSubmitFormCreateIgnoredWhilePendingCreateInFlight
    submitForm dropping the second create must also show a Toast (m.err lost its rendering, bt-81f0)
--- FAIL: TestDeleteConfirmEnterTargetVanishedClosesGracefully
    enter on a vanished target must also show a Toast (m.err lost its rendering, bt-81f0)
--- FAIL: TestValueMenuTargetVanishedClosesGracefully
--- FAIL: TestBlockingPickerEnterTargetVanishedClosesGracefully
--- FAIL: TestParentPickerEnterTargetVanishedClosesGracefully
--- FAIL: TestTagPickerEnterTargetVanishedClosesGracefully
--- FAIL: TestEditTitleSubmitTargetVanishedShowsToast (neu)
--- FAIL: TestSettingsFormSubmitSaveErrorShowsToast (neu)
--- FAIL: TestDispatchPaletteCreateIgnoredWhileCreateInFlight
--- FAIL: TestViewBrowseRepoNeverRendersMErr (neu, Render-Seite)
--- FAIL: TestViewBacklogNeverRendersMErr (neu)
--- FAIL: TestViewTagManagementNeverRendersMErr (neu)
```

Nach Implementierung: alle GREEN. Nebenbefund während GREEN-Lauf: sechs der
o.g. Tests hatten eine STALE Assertion `cmd != nil -> Fatal` ("no doomed
mutation") — showToast() gibt bei nicht-sticky Toasts jetzt IMMER einen
Cmd zurück (den Auto-Dismiss-Tick, toastTimeout). Assertion umgedreht zu
`cmd == nil -> Fatal` mit Kommentar (Cmd ist strukturell garantiert NIE eine
Mutation, da der Guard vor jedem mutateCmd(...)-Aufbau returned) — NICHT
per cmd() ausgeführt (toastError-Dauer 8s hätte den Testlauf blockiert).

Voller Lauf (Gate, ohne -short):
```
ok  	beans-tui	[no test files]
ok  	beans-tui/cmd	0.455s
ok  	beans-tui/internal/config	0.966s
ok  	beans-tui/internal/data	2.790s
ok  	beans-tui/internal/theme	1.232s
ok  	beans-tui/internal/tui	155.388s
```
Dauer gesamt ~2:36 min. `go vet ./...` clean, `gofmt -l .` leer,
`go build -o bin/bt .` clean.

## Golden-Diffs

Keine Regeneration nötig. `TestChromeGolden`/`TestTreeGolden`/
`TestBacklogGolden` liefen unverändert GREEN gegen die bestehenden
Snapshots — alle drei Fixtures haben `m.err`/ErrNote nie gesetzt (leer),
der frühere errNote-Zweig war in ihnen also schon vorher der No-Op-Pfad.
Q1 bestätigt: keine Layout-Höhen-Änderung, daher kein Golden-Bruch erwartet
und keiner beobachtet.

## Smoke (tmux, Session bt81f0smoke, 100x30, gegen dieses Repo)

1. PO-Repro exakt nachgestellt: `/` + `tag:` (ungültige Bleve-Query) + Enter
   -> NUR Toast oben rechts ("● beans list: exit status 1: Er…"), Status-
   zeile unten rechts LEER (kein zweiter Fehlertext). Pane-Capture bestätigt
   leere Statuszeile vor dem unteren Rahmen.
2. Real getriggert (nicht nur Unit-Level): scratch bean `bt-8hiz` angelegt
   (`beans create`), im TUI Status-Menü (`s`) auf bt-8hiz geöffnet, bean
   EXTERN per `beans delete bt-8hiz` gelöscht (fsnotify-Reload während
   Overlay offen), Enter gedrückt -> box_menu_value.go:191's Guard feuert,
   Toast "● Bean no longer exists — selec…" oben rechts, Statuszeile unten
   weiterhin LEER. Das IST einer der sieben Ex-stummen Stellen, live
   getriggert, keine Mock-Konstruktion.
3. Die übrigen sechs Ex-stummen Stellen: Unit-Level (TDD-Tests oben) statt
   Live-Trigger — box_confirm_create.go's Settings-Save-Fehler (deterministisch
   via HOME-Kollisionsdatei erzwungen) und der create-in-flight-Guard (async
   Zeitfenster) sind live praktisch nicht reproduzierbar innerhalb der
   Smoke-Zeit; ehrlich abgegrenzt statt vorgetäuscht.
4. Scratch-Artefakt bt-8hiz vollständig entfernt (`beans delete`), keine
   Repo-Verschmutzung. tmux-Session `bt81f0smoke` sauber beendet
   (`tmux kill-session`).
5. Footer-Höhe unverändert (Q1) -> kein zusätzlicher 80-Spalten-Sonderfall
   nötig (CLAUDE.md-Regel greift nur bei sichtbarer Footer-Änderung).

## Deviations/ERRATA

1. **Zeilendrift bestätigt** (Prelude-Warnung zutreffend): box_picker_tag.go
   Ist-Zeile 424 (Plan: 425, Prelude-Schätzung: ~409) — beide Plan-Quellen
   danebengelegen, eigener Grep-Sweep war maßgeblich. overlay_palette.go
   Ist-Zeile 231 (Plan: 232) — off-by-one, gleiche Ursache (Datei seit
   Plan-Erstellung um eine Zeile verschoben).
2. **"Sieben Stellen" = 7 Dateien, nicht 7 Zeilen**: box_confirm_create.go
   bündelt 4 individuelle `m.err =`-Zeilen (48/66/91/96) unter einem
   Tabellen-Eintrag — alle 4 wurden individuell behandelt (nicht nur die
   erste), wie das Bean-Prompt-Inventar es explizit auflistet.
3. **toastError statt toastWarn bei den beiden `createInFlightNote`-Stellen**
   (box_confirm_create.go:48, overlay_palette.go:231): bt-81f0s bindender
   Rahmen schreibt `showToast(toastError, ...)` für ALLE sieben Stellen
   wörtlich vor. update.go:735s eigene (bereits bestehende) Toast-Kopie
   derselben Meldung nutzt jedoch bewusst `toastWarn` ("Hinweis, kein harter
   Fehler", eigener Doc-Kommentar dort). Diese Inkonsistenz (gleiche
   Meldung, zwei Farben je nach Aufrufstelle) wurde NICHT stillschweigend
   aufgelöst, sondern die bindende Vorgabe wörtlich befolgt und die
   Abweichung hier dokumentiert — reine Farb-/Severity-Frage, keine
   Scope-Änderung, daher kein STOPP.
4. **Reload-Settings-Fehlerpfad (box_confirm_create.go:96) nur strukturell
   abgesichert, kein dedizierter Test**: `config.LoadSettings()` gibt
   praktisch nur bei einem echten `os.UserHomeDir()`-Fehler einen Fehler
   zurück — dieser Zustand ist im selben Testlauf nicht isolierbar (die
   vorausgehende `SaveUserSettings`-Zeile würde bei fehlendem/leerem HOME
   bereits selbst fehlschlagen und den ANDEREN Zweig treffen). showToast-
   Aufruf ist trotzdem gesetzt (Akzeptanz "kein Fehler wird leiser als
   vorher"), aber unit-technisch unreachable belassen — ehrlich als
   Unit-Level-Lücke dokumentiert statt eines künstlichen/invasiven Test-Hooks.
5. **Q1-Annahme final bestätigt**: Supervisor-Entscheid (Variante 1, nur
   Fehler-Anbindung kappen) umgesetzt. Kein Golden-Bruch, keine Footer-
   Höhen-Änderung — PO kann am Review-Gate widersprechen, aber der
   risikoärmere Schnitt hat sich als korrekt erwiesen (0 Goldens mussten
   regeneriert werden).

## Notes for bt-l8e7 (Lobby-Suche verschluckt i/k, view_lobby.go)

Keine Datei-Überlappung mit dieser Änderung. Angefasste Dateien hier:
box_confirm_create.go, box_confirm_delete.go, box_menu_value.go,
box_picker_parent.go, box_picker_blocking.go, box_picker_tag.go,
overlay_palette.go, view.go, view_browse_repo.go, view_tag_management.go,
view_browse_backlog.go (+ zugehörige _test.go, + neue
notifications_toast_only_test.go). bt-l8e7 arbeitet in view_lobby.go —
KEINE dieser Dateien wird dort berührt. Einziger gemeinsamer Nenner:
`showToast`/`toastError`-Konvention (overlay_show_toast.go, unverändert) —
falls bt-l8e7 ebenfalls einen Fehlerpfad braucht, folgt dasselbe Muster
(`m, toastCmd = m.showToast(toastError, <text>, "", nil, false)`).
