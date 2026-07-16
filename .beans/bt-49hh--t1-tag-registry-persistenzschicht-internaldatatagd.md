---
# bt-49hh
title: T1 — Tag-Registry Persistenzschicht (internal/data/tagdefs.go)
status: in-progress
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

- `.beans-tags.yml`-Dateiname exakt wie D01 · `LoadTagDefs`
  tolerant-missing UND tolerant-korrupt (nie Fehler zurück) ·
  ungültige Namen defensiv gefiltert · `SaveTagDefs` sortiert+dedupliziert
  · `AddTagDefName`/`RemoveTagDefName`/`RenameTagDefName` rein (kein I/O,
  keine Mutation des Eingabe-Slices — neue Slice zurückgeben) · alle
  RED-Tests zuerst rot, dann grün · `gofmt`/`go vet` clean · voller Lauf
  grün · Commit `feat(data): Tag-Registry-Persistenz (.beans-tags.yml)`.
