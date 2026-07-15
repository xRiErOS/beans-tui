# beans-tui (`bt`)

PO-Cockpit-TUI für beans-Repos — Port der DevDash-TUI (`dd`) auf das
[beans](https://github.com/hmans/beans)-Framework. Design/Architektur:
[`docs/plans/v1-port/design-spec.md`](docs/plans/v1-port/design-spec.md).

## Status

E1 (Foundation), E2 (Browse & Detail), E3 (Mutationen) und E4 (Command-Center &
Review-Cockpit) sind fertig: read-only Tree über den beans-Datenlayer
(Milestones → Epics → Tasks) mit Live-Reload via fsnotify-Watcher, Quit-Confirm
(E1); Master-Detail-Fokus mit Detail-Accordion (Meta/Body/Beziehungen/Historie,
Beziehungs-Sprung), lokale Live-Suche + Bleve ab 3 Zeichen, Facetten-Filter
(Status/Type/Priority/Tag, geteilt über Tree UND Backlog) und die
Backlog-View mit Sort-Toggle (E2); volle Mutations-Verdrahtung — kombiniertes
Status/Type/Priority-Menü, Tag-/Parent-/Blocking-Picker, Create-Form (huh,
Confirm-Gate), Titel-/Body-Edit (`$EDITOR`) und Delete-Confirm mit
Kinder-/Verknüpfungs-Warnung, durchgehend mit ETag-Konflikt-Handling (E3);
Command-Center (`ctrl+k`, fuzzy Aktionen + Bean-Suche gemischt,
kontextabhängig zuerst) und das Review-Cockpit (`R`, PO-Merge-Gate: Queue
gruppiert nach Epic + Rework-Sichtbarkeits-Sektion, Verdikt-Aktionen
Pass/Reject/Reopen) (E4). E5–E6 (Polish, Validierung) sind offen — offener
Stand: `beans list --ready`.

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

## Keybindings (Stand E4)

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
| `X` | Facetten-Filter zurücksetzen (Tree und Backlog) |
| `b` | Backlog-View (parentlose+ready beans, geteilter Such-/Filter-Zustand) |
| `S` | Backlog: Sort-Toggle, zyklisch status → priority → created → updated |
| `s` | Status/Type/Priority-Menü (kombiniert, ein Key für alle drei) |
| `t` | Tag-Picker (Toggle-Multi-Select, Zähler, Freitext-Neuanlage) |
| `a` | Parent-Picker (Zyklen-Ausschluss + Typ-Hierarchie) |
| `B` | Blocking-Picker (Toggle-Multi-Select) |
| `c` | Bean anlegen (huh-Formular, Confirm-Gate) |
| `e` | Titel bearbeiten (Formular, direkt ohne Confirm) |
| `ctrl+e` | Body im `$EDITOR` bearbeiten (`$VISUAL` → `$EDITOR` → `vi`) |
| `d` | Löschen (Confirm, Kinder-/Verknüpfungs-Warnung — kaskadiert nicht) |
| `ctrl+r` | Daten neu laden |
| `ctrl+k`/`K` | Command-Center öffnen (fuzzy Aktionen + Bean-Suche, von überall außer aus einem offenen Overlay/Formular/Filter/Suchfeld heraus) |
| `R` | Review-Cockpit öffnen (PO-Merge-Gate, siehe unten) |
| `q` | Quit (mit Confirm) |
| `ctrl+c` | Sofort-Quit |

### Review-Cockpit (`R`)

Eigener Tasten-Scope, überschreibt s/t/a/B/c/d/e (Feldbearbeitung läuft über
Tree/Backlog, bewusster Scope-Cut) — nur die folgenden Tasten sind aktiv:

| Taste | Aktion |
|---|---|
| `↑`/`i`/`p`, `↓`/`k`/`n` | Queue-Cursor (to-review-Gruppen nach Epic, dann Rework-Sektion) |
| `1`–`4` | Detail-Accordion des cursorierten Beans: Section-Sprung |
| `a` | Pass — Status → `completed`, Tag `to-review` entfernt (nur auf to-review-Items) |
| `x` | Reject — öffnet Kommentar-Formular, danach Tag `to-review`→`rework` + `## Review <datum>`-Body-Abschnitt (nur auf to-review-Items) |
| `o` | Reopen — Tag `rework`→`to-review` (nur auf Rework-Items) |
| `esc`/`q` | Zurück zur Browse-Ansicht |
| `ctrl+k`/`K` | Command-Center öffnet trotzdem (Capture-Order-Ausnahme) |

## Entwicklung

TDD (`superpowers:test-driven-development`), Ausführung `make test`
(`command go test ./...`). Konventionen (Build immer `command go …`,
Datei-Namensschema, Theme-Token, Commit-/Review-Flow) → [`CLAUDE.md`](CLAUDE.md).
