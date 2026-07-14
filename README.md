# beans-tui (`bt`)

PO-Cockpit-TUI für beans-Repos — Port der DevDash-TUI (`dd`) auf das
[beans](https://github.com/hmans/beans)-Framework. Design/Architektur:
[`docs/plans/v1-port/design-spec.md`](docs/plans/v1-port/design-spec.md).

## Status

E1 (Foundation) ist fertig: read-only Tree über den beans-Datenlayer
(Milestones → Epics → Tasks), Live-Reload via fsnotify-Watcher, Quit-Confirm.
E2–E6 (Browse & Detail, Suche/Filter, Review-Cockpit, Mutationen, Polish)
sind in Arbeit — offener Stand: `beans list --ready`.

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

## Keybindings (Stand E1)

| Taste | Aktion |
|---|---|
| `↑`/`i`, `↓`/`k` | Cursor |
| `→`/`l` | Knoten expandieren |
| `←`/`j` | Knoten einklappen |
| `tab` | Fokus-Tausch Tree ↔ Detail |
| `ctrl+r` | Daten neu laden |
| `q` | Quit (mit Confirm) |
| `ctrl+c` | Sofort-Quit |

## Entwicklung

TDD (`superpowers:test-driven-development`), Ausführung `make test`
(`command go test ./...`). Konventionen (Build immer `command go …`,
Datei-Namensschema, Theme-Token, Commit-/Review-Flow) → [`CLAUDE.md`](CLAUDE.md).
