---
# bt-49hh
title: T1 — Tag-Registry Persistenzschicht (internal/data/tagdefs.go)
status: completed
type: task
priority: normal
created_at: 2026-07-16T15:44:20Z
updated_at: 2026-07-16T15:48:13Z
parent: bt-362n
---

T1 — Persistenzschicht `internal/data/tagdefs.go` (Epic `bt-362n`, D01-D04).
Unabhängig (kein `blocked_by`) — reiner Neu-Datei-Scope, keine bestehende
Datei wird angefasst. Fundament für T2-T6.

## Ziel

Repo-lokale Tag-Registry: laden/speichern von `.beans-tags.yml` im
Repo-Root (`client.RepoDir`, D01), tolerant-missing (D02), plus reine
Slice-Helfer für Add/Remove/Rename (TDD-freundlich, kein I/O).

## Betroffene Dateien/Symbole (neu)

- `internal/data/tagdefs.go` (neu):
  - `const tagDefsFileName = ".beans-tags.yml"`
  - `type tagDefsFile struct { Tags []string `+"`"+`yaml:"tags,omitempty"`+"`"+` }`
  - `func (c *Client) LoadTagDefs() ([]string, error)` — liest
    `filepath.Join(c.RepoDir, tagDefsFileName)`; fehlende Datei → `(nil,
    nil)`; korrupte YAML → `(nil, nil)` (NIE Fehler nach außen, mirrort
    `config.LoadSettings`); gültige Datei → sortierte, deduplizierte,
    NUR `data.ValidTagName`-gültige Namen (defensive Filterung gegen
    eine von Hand kaputt editierte Datei — ungültige Zeilen werden
    still übersprungen, nicht als Fehler eskaliert, gleiche
    „nie crashen"-Philosophie wie das restliche Laden).
  - `func (c *Client) SaveTagDefs(tags []string) error` — sortiert +
    dedupliziert vor dem Schreiben (deterministische Diffs),
    `os.WriteFile` mit `0o644`, KEIN Verzeichnis-Create nötig (Repo-Root
    existiert per Definition, anders als `~/.config/beans-tui`).
  - `func AddTagDefName(defs []string, name string) []string` — dedupe +
    sortiert einfügen, No-Op wenn bereits enthalten.
  - `func RemoveTagDefName(defs []string, name string) []string` —
    entfernt (falls vorhanden), No-Op sonst.
  - `func RenameTagDefName(defs []string, old, new string) []string` —
    entfernt `old`, fügt `new` ein (dedupe), No-Op auf `old`, wenn nicht
    vorhanden (dann nur ein reines Add von `new`, falls `new` noch
    fehlt — Rename auf einen NICHT-registrierten Namen ist zulässig,
    z. B. Promotion eines freien Tags während des Rename-Flows).

## TDD (RED zuerst, `internal/data/tagdefs_test.go`, neu)

Mirrort `internal/config/settings_test.go`s Namens-/Struktur-Konvention
(`TestLoadSettingsMissingFileReturnsDefaults` etc.). RED-Zitate (erste
Assertion, die vor der Implementierung failt):

```go
func TestLoadTagDefsMissingFileReturnsEmpty(t *testing.T) {
    c := &Client{RepoDir: t.TempDir()}
    got, err := c.LoadTagDefs()
    if err != nil || len(got) != 0 {
        t.Fatalf("want (nil, nil), got (%v, %v)", got, err)
    }
}

func TestLoadTagDefsSkipsInvalidNamesDefensively(t *testing.T) {
    dir := t.TempDir()
    os.WriteFile(filepath.Join(dir, ".beans-tags.yml"),
        []byte("tags:\n  - good-tag\n  - Bad_Tag\n  - \n"), 0o644)
    c := &Client{RepoDir: dir}
    got, _ := c.LoadTagDefs()
    if len(got) != 1 || got[0] != "good-tag" {
        t.Fatalf("want [good-tag], got %v", got)
    }
}

func TestSaveTagDefsRoundTripSortedDeduped(t *testing.T) {
    dir := t.TempDir()
    c := &Client{RepoDir: dir}
    if err := c.SaveTagDefs([]string{"zeta", "alpha", "alpha"}); err != nil {
        t.Fatal(err)
    }
    got, _ := c.LoadTagDefs()
    want := []string{"alpha", "zeta"}
    if !reflect.DeepEqual(got, want) {
        t.Fatalf("want %v, got %v", want, got)
    }
}

func TestRenameTagDefNamePromotesUnregisteredOldName(t *testing.T) {
    got := RenameTagDefName([]string{"a", "b"}, "c-not-registered", "d")
    want := []string{"a", "b", "d"}
    if !reflect.DeepEqual(got, want) {
        t.Fatalf("want %v, got %v", want, got)
    }
}
```

Weitere Tests (kein RED-Zitat nötig, aber Pflicht): `AddTagDefName`
No-Op-bei-Duplikat, `RemoveTagDefName` No-Op-bei-Abwesenheit, korrupte
YAML → leere Liste kein Panic, `SaveTagDefs` schreibt `0o644` und
gültiges YAML (Round-Trip mit `LoadTagDefs`).

## Golden-Strategie

GEGENBELEG (kein `-update`): reiner Neu-Datei-Scope, kann die 3
Basis-Goldens (Tree/Backlog/Chrome) strukturell nicht berühren — nach
`command go build -o bin/bt .` trotzdem `command go test ./internal/tui/
-run "TestTreeGolden|TestBacklogGolden|TestChromeGolden"` grün laufen
lassen und im Commit-Body als „unverändert, Gegenbeleg" festhalten.

## tmux-Smoke

Nicht anwendbar (keine UI-Änderung in diesem Task) — stattdessen: nach
dem Build `bin/bt` in diesem Repo einmal bei 120 UND bei 80 Spalten
starten, verifizieren dass die TUI unverändert hochfährt (keine neue
Datei `.beans-tags.yml` wird beim bloßen Start angelegt — `LoadTagDefs`
hat noch keinen Aufrufer, T1 verdrahtet nichts in die UI). Danach
`git status --porcelain` leer (keine Test-Artefakte).

## Akzeptanz-Checkliste

- [x] `.beans-tags.yml`-Dateiname exakt wie D01
- [x] `LoadTagDefs` tolerant-missing UND tolerant-korrupt (nie Fehler zurück)
- [x] ungültige Namen defensiv gefiltert
- [x] `SaveTagDefs` sortiert+dedupliziert
- [x] `AddTagDefName`/`RemoveTagDefName`/`RenameTagDefName` rein (kein I/O,
  keine Mutation des Eingabe-Slices — neue Slice zurückgeben)
- [x] alle RED-Tests zuerst rot, dann grün
- [x] `gofmt`/`go vet` clean
- [x] voller Lauf grün
- [x] Commit `feat(data): Tag-Registry-Persistenz (.beans-tags.yml)`

## Summary

`internal/data/tagdefs.go` (neu) liefert die Persistenzschicht der
Tag-Registry: `Client.LoadTagDefs()`/`Client.SaveTagDefs([]string)`
(repo-lokal, `.beans-tags.yml` via `RepoDir`, D01/D02/D04) plus die drei
reinen Slice-Helfer `AddTagDefName`/`RemoveTagDefName`/`RenameTagDefName`
(kein I/O, geben immer eine neue, sortierte, deduplizierte Slice zurück —
Eingabe-Slice bleibt unangetastet, verifiziert per eigenem Test). Load ist
tolerant-missing UND tolerant-korrupt (mirrort `config.LoadSettings`, nie
ein Fehler nach außen) und filtert defensiv jeden Namen, der
`data.ValidTagName` nicht besteht. Reiner Neu-Datei-Scope — kein
bestehender Code berührt, `LoadTagDefs` hat noch keinen Aufrufer (T2
verdrahtet das erst).

## Test-Output (RED→GREEN wörtlich)

RED (vor Implementierung, `command go test ./internal/data/... -run
"TestLoadTagDefs|TestSaveTagDefs|TestRenameTagDefName|TestAddTagDefName|TestRemoveTagDefName|TestAddRemoveRenameDoNotMutateInputSlice"
-v -count=1`):

```
# beans-tui/internal/data [beans-tui/internal/data.test]
internal/data/tagdefs_test.go:19:16: c.LoadTagDefs undefined (type *Client has no field or method LoadTagDefs)
internal/data/tagdefs_test.go:30:14: c.LoadTagDefs undefined (type *Client has no field or method LoadTagDefs)
internal/data/tagdefs_test.go:39:14: c.SaveTagDefs undefined (type *Client has no field or method SaveTagDefs)
internal/data/tagdefs_test.go:42:14: c.LoadTagDefs undefined (type *Client has no field or method LoadTagDefs)
internal/data/tagdefs_test.go:50:9: undefined: RenameTagDefName
internal/data/tagdefs_test.go:66:16: c.LoadTagDefs undefined (type *Client has no field or method LoadTagDefs)
internal/data/tagdefs_test.go:80:14: c.SaveTagDefs undefined (type *Client has no field or method SaveTagDefs)
internal/data/tagdefs_test.go:93:9: undefined: AddTagDefName
internal/data/tagdefs_test.go:101:9: undefined: AddTagDefName
internal/data/tagdefs_test.go:109:9: undefined: RemoveTagDefName
internal/data/tagdefs_test.go:109:9: too many errors
FAIL	beans-tui/internal/data [build failed]
FAIL
```

GREEN (nach Implementierung, gleiches `-run`-Filter, 12/12 Tests):

```
=== RUN   TestLoadTagDefsMissingFileReturnsEmpty
--- PASS: TestLoadTagDefsMissingFileReturnsEmpty (0.00s)
=== RUN   TestLoadTagDefsSkipsInvalidNamesDefensively
--- PASS: TestLoadTagDefsSkipsInvalidNamesDefensively (0.00s)
=== RUN   TestSaveTagDefsRoundTripSortedDeduped
--- PASS: TestSaveTagDefsRoundTripSortedDeduped (0.00s)
=== RUN   TestRenameTagDefNamePromotesUnregisteredOldName
--- PASS: TestRenameTagDefNamePromotesUnregisteredOldName (0.00s)
=== RUN   TestLoadTagDefsCorruptYAMLReturnsEmptyNoPanic
--- PASS: TestLoadTagDefsCorruptYAMLReturnsEmptyNoPanic (0.00s)
=== RUN   TestSaveTagDefsWritesReadablePermissions
--- PASS: TestSaveTagDefsWritesReadablePermissions (0.00s)
=== RUN   TestAddTagDefNameNoOpOnDuplicate
--- PASS: TestAddTagDefNameNoOpOnDuplicate (0.00s)
=== RUN   TestAddTagDefNameInsertsSortedDeduped
--- PASS: TestAddTagDefNameInsertsSortedDeduped (0.00s)
=== RUN   TestRemoveTagDefNameNoOpOnAbsence
--- PASS: TestRemoveTagDefNameNoOpOnAbsence (0.00s)
=== RUN   TestRemoveTagDefNameRemovesExisting
--- PASS: TestRemoveTagDefNameRemovesExisting (0.00s)
=== RUN   TestRenameTagDefNameRenamesExisting
--- PASS: TestRenameTagDefNameRenamesExisting (0.00s)
=== RUN   TestAddRemoveRenameDoNotMutateInputSlice
--- PASS: TestAddRemoveRenameDoNotMutateInputSlice (0.00s)
PASS
ok  	beans-tui/internal/data	0.474s
```

Gates: `command gofmt -l internal/data/tagdefs.go internal/data/tagdefs_test.go`
leer · `command go vet ./...` leer · `command go test ./... -short -count=1`
grün (alle Pakete ok) · `command go test ./... -count=1` grün (140s Voll-Lauf,
inkl. `internal/tui` 138.6s, exit 0).

## Golden-Gegenbeleg

`command go build -o bin/bt .` erfolgreich, danach `command go test
./internal/tui/... -run "TestTreeGolden|TestBacklogGolden|TestChromeGolden"
-count=2` → alle 5 Testfunktionen (inkl. `*Deterministic`-Varianten) PASS,
beide Wiederholungen. `git diff --stat -- internal/tui/testdata/` leer —
Basis-Goldens unverändert (wie erwartet, reiner Neu-Datei-Scope).

## Smoke

tmux-Smoke statt regulärem UI-Smoke (Task-Vorgabe: „keine UI-Änderung,
`LoadTagDefs` hat noch keinen Aufrufer"): `bin/bt` in diesem Repo je einmal
bei 120×40 UND 80×40 gestartet (tmux-Sessions `bt49hhsmoke120`/
`bt49hhsmoke80`) — Startup in beiden Breiten unverändert (Command-Center-
Chrome, Tree/Detail-Panes identisch zum bekannten Bild), sauberer Exit via
`q`→Enter (Quit-Confirm), danach kein `bt`-Prozess mehr aktiv. `ls
.beans-tags.yml` nach beiden Läufen: „No such file or directory" — bestätigt
D02/den Task-Claim, dass der bloße Start KEINE neue Registry-Datei anlegt
(`LoadTagDefs` unverdrahtet). `git status --porcelain` danach nur die zwei
neuen Quelldateien (`internal/data/tagdefs.go`, `tagdefs_test.go`), keine
Test-Artefakte.

## Deviations/ERRATA

Keine. Zusätzlich zu den 4 im Bean-Body geforderten RED-Zitat-Tests wurden
6 weitere Tests ergänzt (Akzeptanz-Checkliste verlangt sie explizit, aber
ohne eigenes RED-Zitat: `TestLoadTagDefsCorruptYAMLReturnsEmptyNoPanic`,
`TestSaveTagDefsWritesReadablePermissions`, `TestAddTagDefNameNoOpOnDuplicate`,
`TestAddTagDefNameInsertsSortedDeduped`, `TestRemoveTagDefNameNoOpOnAbsence`,
`TestRemoveTagDefNameRemovesExisting`, `TestRenameTagDefNameRenamesExisting`,
`TestAddRemoveRenameDoNotMutateInputSlice`) — reine Ergänzung, keine
Abweichung vom Spezifizierten.

## Notes for T2+T6

- `Client.LoadTagDefs()`/`Client.SaveTagDefs([]string)` sind bereit zum
  Verdrahten — T2 ruft `LoadTagDefs` bei jedem `openTagManagementPage()`
  frisch auf (D03), T6 bei jedem `openTagPicker()`.
- `AddTagDefName`/`RemoveTagDefName`/`RenameTagDefName` sind rein und
  slice-safe (kopieren immer) — T3/T4/T5 können sie direkt vor einem
  `SaveTagDefs`-Call verketten, ohne den Aufrufer-State zu mutieren.
- `data.ValidTagName` (bestehend, `tags.go`) ist die einzige
  Namens-Grammatik-Quelle — T3s Create-Validierung sollte sie wörtlich
  wiederverwenden (D11 verlangt das explizit), nicht duplizieren.
- `.beans-tags.yml` existiert in diesem Repo NICHT (nach Task-Ende sauber
  entfernt) — T2/T3/T4/T6s eigene tmux-Smokes starten alle bei einer
  leeren/fehlenden Registry (LoadTagDefs liefert `(nil, nil)`).
