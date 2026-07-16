---
# bt-1e0t
title: Backlog-Footer verliert Sort-Eintrag, S bleibt funktional (D02)
status: completed
type: task
priority: normal
created_at: 2026-07-16T06:45:48Z
updated_at: 2026-07-16T12:48:11Z
parent: bt-tct9
---

E9 Task 6 — deckt D02 aus bean bt-tct9 (PO-bestätigt, "D02 BESTÄTIGT (Option b +
Präzisierung)" im Epic-Body). Quelle: design-spec.md §15 PF-17 (Abschnitt D02). Ist-Code:
internal/tui/view_browse_backlog.go (backlogLocalBindings), internal/tui/keymap.go
(helpGroups, keys.Sort — bereits vorhanden, unverändert). Kein blocked_by — unabhängig,
eigene Datei/Funktion.

## D02 — Backlog-Footer verliert `Sort`-Eintrag, Taste bleibt funktional

PO-bestätigt (Option b + Präzisierung): "'S Sort' fliegt aus dem Backlog-Footer; die
S-Taste bleibt funktional, wird aber NUR im Help-Overlay ('?') dokumentiert. Suchzeilen-
Suffix '· sort <modus>' bleibt die sichtbare Zustandsanzeige." Footer passt danach in 2
Zeilen bei 80 Spalten (statt bisher potenziell 3).

## Architektur-Vorgabe

`backlogLocalBindings()` (view_browse_backlog.go) verliert das angehängte `keys.Sort`:

```go
// VORHER:
func backlogLocalBindings() []keybind.Binding {
    return append(append([]keybind.Binding{}, browseRepoLocalBindings()...), keys.Sort)
}

// NACHHER:
func backlogLocalBindings() []keybind.Binding {
    return browseRepoLocalBindings()
}
```

`S` bleibt funktional (`keyBacklog`s `keys.Sort`-Case, view_browse_backlog.go, UNVERÄNDERT)
— taucht aber nur noch in `helpGroups()` auf (keymap.go, dort BEREITS gelistet unter
"Actions", KEINE Änderung an `helpGroups()` nötig — die Drift-Guard-Tests
`TestHelpGroupsCoverEveryBindingExactlyOnce`/`TestNoDuplicateBindingBetweenGlobalAndAnyLocalHintList`
müssen weiterhin grün bleiben: `keys.Sort` bleibt in `helpGroups()`, verschwindet nur aus
`backlogLocalBindings()` — kein Widerspruch, das Drift-Guard prüft NUR dass jedes Binding
IRGENDWO auftaucht, nicht dass es überall auftaucht). Suchzeilen-Suffix `· sort <modus>`
(`treeSearchLine`, view_browse_repo.go — bereits aus E8/D02 vorhanden) bleibt UNVERÄNDERT
die sichtbare Laufzeit-Anzeige des aktiven Sort-Modus.

## TDD-Schritte

1. Failing test (view_browse_backlog_test.go oder chrome_test.go, je nach bestehendem
   Test-Ort für `backlogLocalBindings`): `TestBacklogLocalBindingsOmitsSort` (prüft dass
   `keys.Sort` NICHT mehr in `backlogLocalBindings()` enthalten ist). Bestehenden Test, der
   das GEGENTEIL prüfte (falls vorhanden, z. B. ein Test der `keys.Sort` explizit als
   letztes Element erwartet), FINDEN und aktualisieren statt einen widersprüchlichen Test
   stehen zu lassen.
2. `command go test ./internal/tui/... -run "BacklogLocalBindings|Sort"` → FAIL.
3. Implementieren (eine Zeile).
4. `command go test ./internal/tui/... -run "BacklogLocalBindings|Sort"` → PASS.
5. Golden-Regen: Backlog-Footer ist Teil von `TestBacklogGolden`/`TestChromeGolden` (Footer-
   Zeile ändert sich sichtbar, ein Eintrag weniger) — `command go test ./internal/tui/ -run
   "TestBacklogGolden|TestChromeGolden" -update`, Vorher/Nachher-Diff PFLICHT im Commit-
   Body. `TestTreeGolden` sollte UNVERÄNDERT bleiben (Tree-Footer war nie betroffen) —
   Gegenbeleg dafür explizit mitlaufen lassen und im Commit-Body als "unverändert"
   vermerken.
6. **Grenzbreiten-Smoke PFLICHT** (CLAUDE.md-Regel, Footer-/Wrap-Änderungen brauchen einen
   tmux-Smoke bei 80 Spalten — LESSONS-LEARNED Eintrag 4, NBSP-Wordwrap-Falle): Backlog-
   Footer bei 80 Spalten real in tmux gegen `./bin/bt` prüfen, `tmux capture-pane -p` als
   Beleg im Commit-Body — verifizieren dass der Footer jetzt tatsächlich in 2 statt 3
   Zeilen passt (PO-Aussage "Footer passt dann in 2 Zeilen bei 80 Spalten" empirisch
   bestätigen, nicht nur annehmen).
7. `command go test ./... -short` grün (2x), voller Lauf grün, `-race` grün, gofmt/vet leer.
8. Commit `feat(tui): Backlog-Footer verliert Sort-Eintrag, S bleibt funktional (D02)`.
   Footer `Refs: bt-tct9`.

## Akzeptanz-Checkliste

- [x] `backlogLocalBindings()` enthält `keys.Sort` nicht mehr
- [x] `S` bleibt funktional (Backlog-Sort-Zyklus unverändert)
- [x] `keys.Sort` bleibt in `helpGroups()` (Help-Overlay), Drift-Guard-Tests grün
- [x] Suchzeilen-Suffix `· sort <modus>` unverändert sichtbar
- [x] Backlog-Footer passt bei 80 Spalten in 2 Zeilen (tmux-Smoke-Beleg im Commit-Body)
- [x] Tree-Footer unverändert (Gegenbeleg dokumentiert)
- [x] Goldens regeneriert (Backlog/Chrome) + Vorher/Nachher-Beschreibung im Commit-Body
- [x] Voller Testlauf grün, gofmt/vet leer


## PRELUDE (2026-07-16, aus T5-Review F02/F03 — ZUERST erledigen, eigener Commit)

Zwei Test-Härtungen aus dem T5-Review (bt-4mo9), beide low, kein eigener Review-Zyklus:
1. **F03 (wertvollste):** Jitter-Parity-Regressionstest für Blocking- UND Parent-Picker —
   dieselbe gewrappte lange Titel-Zeile einmal als Cursor-Row, einmal als Non-Cursor-Row
   rendern, Einzug der Folgezeile MUSS identisch sein (Mutations-Beweis D des Reviewers:
   Pad-Revert 1→2 Spaces ließ die volle Suite grün — diese Lücke schließen).
2. **F02:** TestBlockingPickerBoxUsesWideModalWidth/TestParentPickerBoxUsesWideModalWidth
   sind selbstreferenziell (wantW := wideModalWidth(m.width)+2). Je einen unabhängigen
   Literal-Erwartungswert ergänzen (z.B. bei termW=120 → wantW 104).
Commit: `test(tui): Picker-Jitter-Parity + Literal-Breiten-Pins (T5-F02/F03)`, `Refs: bt-1e0t`.


## Summary

D02 umgesetzt: `backlogLocalBindings()` gibt jetzt unverändert `browseRepoLocalBindings()`
zurück, kein angehängtes `keys.Sort` mehr. `S` bleibt funktional (`keyBacklog`s Sort-Case
unverändert), taucht nur noch im Help-Overlay auf (`helpGroups()`, keymap.go — unverändert,
Sort war dort bereits gelistet). Suchzeilen-Suffix `· sort <modus>` unverändert. PRELUDE
(F02/F03 aus dem T5-Review, bean bt-4mo9) zuerst erledigt, eigener Commit.

Zwei Folgeänderungen waren durch D02 zwingend mitgezogen (nicht im TDD-Schritt-Text
explizit genannt, aber direkte Konsequenz der Architektur-Vorgabe):
- `TestBacklogChromeFooterMatchesQ06ListPlusSort` → umbenannt `...MatchesQ06List`, want-
  String verliert `· S Sort` (jetzt byte-identisch mit `TestBrowseRepoChromeFooterMatchesQ06List`).
- `TestDetailClickBacklogThreeLineFooterAt80Cols` (mouse_test.go) → umbenannt
  `...FooterAt80Cols`: dieser Test baute seine Unterscheidungskraft explizit auf der
  3-vs-2-Zeilen-Divergenz zwischen Backlog- und Browse-Footer bei 80 Spalten auf
  (bt-d8kc-Ära). D02 macht beide Footer-Listen identisch — die Divergenz existiert nicht
  mehr. Test bleibt als Grenzbreiten-Pin für den dynamischen footH-Mechanismus (footerY ==
  originY+bodyH+2, click/border-Auflösung) bestehen, Verlust der Chrome-Unterscheidungskraft
  ist im Kommentar dokumentiert statt stillschweigend übergangen.

## Test-Output

RED (vor Implementierung, `TestBacklogLocalBindingsOmitsSort`):

    view_browse_backlog_test.go:584: backlogLocalBindings() unexpectedly includes
    keys.Sort -- D02 moved Sort to Help-overlay-only (bean bt-1e0t)
    --- FAIL: TestBacklogLocalBindingsOmitsSort (0.00s)

GREEN (nach Implementierung, gleicher Test):

    --- PASS: TestBacklogLocalBindingsOmitsSort (0.00s)

Mitgezogene Tests (nach Implementierung aktualisiert, dann grün):
`TestBacklogChromeFooterMatchesQ06List`, `TestDetailClickBacklogFooterAt80Cols`,
`TestHelpGroupsCoverEveryBindingExactlyOnce`, `TestNoDuplicateBindingBetweenGlobalAndAnyLocalHintList`
— alle PASS.

Voller Lauf (Commit-Gate):

    go test ./... -short -count=1   -> ok (2x, alle Pakete)
    go test ./... -count=1          -> ok (alle Pakete, internal/tui 139.4s)
    go test ./internal/tui/ -race -count=1 -> ok (141.1s)
    gofmt -l .                      -> leer
    go vet ./...                    -> leer

## Smoke

tmux 80 Spalten, Backlog-View (`b`): Footer rendert exakt 2 Zeilen —
`tab focus in · shift+tab focus out · / search · f Filter · s Status · c Create` /
`· d Delete · e Edit · b Backlog · t Tags · y Yank · a Parent · r Blocking` — kein
`S Sort` mehr, kein Umbruch auf eine dritte Zeile (vorher potenziell 3 Zeilen, PO-Aussage
empirisch bestätigt). `S` gedrückt: Suchzeilen-Suffix wechselt sichtbar `· sort status` →
`· sort prio` (Sort-Zyklus funktional unverändert). `?` (Help-Overlay, Fenster auf 45
Zeilen vergrößert um die volle Liste zu sehen): `S Sort` unter "Actions" weiterhin
gelistet.

## Deviations/ERRATA

- Commit-Titel-Format korrigiert: die Prelude- und D02-Commit-Titel wurden initial mit
  67 bzw. 53 Zeichen committet (Bean-Text schlug `test(tui): Picker-Jitter-Parity +
  Literal-Breiten-Pins (T5-F02/F03)` wörtlich vor, das überschreitet die harte
  ≤50-Zeichen-Vorgabe dieser Session) — per `git reset --soft` (kein Push erfolgt, keine
  Historie extern sichtbar) neu committet als `test(tui): Picker-Jitter + Breiten-Pins
  (F02/F03)` (49 Zeichen) und `feat(tui): Backlog-Footer verliert Sort (D02)` (45
  Zeichen), Refs-Footer beider Commits auf `bt-1e0t` korrigiert (Bean-Text schlug
  `Refs: bt-tct9` vor, Session-Vorgabe verlangt `bt-1e0t`). Kein Diff auf Produktionscode/
  Testcode durch diesen Reset, nur Commit-Metadaten.
- F02/F03 (PRELUDE) manuell mutations-verifiziert: temporärer Pad-Revert (1→2 Leerzeichen)
  in `box_picker_blocking.go`/`box_picker_parent.go` ließ beide neuen Jitter-Parity-Tests
  sichtbar fehlschlagen (Indent-Differenz 1 Spalte), danach zurückgesetzt — kein Diff auf
  Produktionscode im Prelude-Commit, reiner Test-Zuwachs.

## Notes for T(n+1)

- Beide Footer-Listen (`backlogLocalBindings`/`browseRepoLocalBindings`) sind jetzt
  identisch — ein zukünftiger Task, der die beiden wieder divergieren lässt (z. B. ein
  neues Backlog-exklusives Binding), sollte `TestBacklogChromeFooterMatchesQ06List` und
  `TestDetailClickBacklogFooterAt80Cols` gegenlesen, deren Kommentare die jetzige Parität
  explizit dokumentieren.
