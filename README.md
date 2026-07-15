# beans-tui (`bt`)

PO-Cockpit-TUI für beans-Repos — Port der DevDash-TUI (`dd`) auf das
[beans](https://github.com/hmans/beans)-Framework. Design/Architektur:
[`docs/plans/v1-port/design-spec.md`](docs/plans/v1-port/design-spec.md).

## Status

E1 (Foundation) und E2 (Browse & Detail) sind fertig: read-only Tree über den
beans-Datenlayer (Milestones → Epics → Tasks) mit Live-Reload via
fsnotify-Watcher, Quit-Confirm (E1); Master-Detail-Fokus mit
Detail-Accordion (Meta/Body/Beziehungen/Historie, Beziehungs-Sprung), lokale
Live-Suche + Bleve ab 3 Zeichen, Facetten-Filter (Status/Type/Priority/Tag,
geteilt über Tree UND Backlog) und die Backlog-View mit Sort-Toggle (E2).
E3–E6 (Mutationen, Command-Center & Review-Cockpit, Polish, Validierung)
sind offen — offener Stand: `beans list --ready`.

## Voraussetzungen

- beans-CLI ≥ 0.4.2 im `PATH`
- Go 1.26+

## Installation

```sh
make build          # → bin/bt
# oder
command go install .
```

## Start

```sh
bt          # sucht .beans.yml aufwärts vom cwd
bt <pfad>   # explizites Repo
```

## Keybindings (Stand E2)

| Taste | Aktion |
|---|---|
| `↑`/`i`, `↓`/`k` | Cursor (Tree/Backlog) bzw. Section-/Feld-Cursor im Detail-Fokus |
| `→`/`l` | Knoten expandieren; im Detail-Fokus: Beziehungen-Section → Feld-Ebene rein |
| `←`/`j` | Knoten einklappen; im Detail-Fokus: Feld-Ebene raus, danach Detail-Fokus verlassen |
| `enter` | Öffnen/Bestätigen; im Detail-Fokus auf einer Beziehung: dorthin springen |
| `tab` | Fokus-Tausch Tree/Backlog ↔ Detail-Accordion |
| `1`–`4` | Detail-Accordion: direkter Section-Sprung (Meta/Body/Beziehungen/Historie) |
| `/` | Suche — lokaler Live-Filter (Titel), ab 3 Zeichen zusätzlich Bleve (Titel+Body) |
| `f` | Facetten-Filter öffnen (Status/Type/Priority/Tag) |
| `X` | Facetten-Filter zurücksetzen (Tree — s. Bug-Hinweis in E2-Abschluss) |
| `b` | Backlog-View (parentlose+ready beans, geteilter Such-/Filter-Zustand) |
| `S` | Backlog: Sort-Toggle, zyklisch status → priority → created → updated |
| `ctrl+r` | Daten neu laden |
| `q` | Quit (mit Confirm) |
| `ctrl+c` | Sofort-Quit |

## Entwicklung

TDD (`superpowers:test-driven-development`), Ausführung `make test`
(`command go test ./...`). Konventionen (Build immer `command go …`,
Datei-Namensschema, Theme-Token, Commit-/Review-Flow) → [`CLAUDE.md`](CLAUDE.md).
