---
# bt-1e0t
title: Backlog-Footer verliert Sort-Eintrag, S bleibt funktional (D02)
status: todo
type: task
created_at: 2026-07-16T06:45:48Z
updated_at: 2026-07-16T06:45:48Z
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

- [ ] `backlogLocalBindings()` enthält `keys.Sort` nicht mehr
- [ ] `S` bleibt funktional (Backlog-Sort-Zyklus unverändert)
- [ ] `keys.Sort` bleibt in `helpGroups()` (Help-Overlay), Drift-Guard-Tests grün
- [ ] Suchzeilen-Suffix `· sort <modus>` unverändert sichtbar
- [ ] Backlog-Footer passt bei 80 Spalten in 2 Zeilen (tmux-Smoke-Beleg im Commit-Body)
- [ ] Tree-Footer unverändert (Gegenbeleg dokumentiert)
- [ ] Goldens regeneriert (Backlog/Chrome) + Vorher/Nachher-Beschreibung im Commit-Body
- [ ] Voller Testlauf grün, gofmt/vet leer
