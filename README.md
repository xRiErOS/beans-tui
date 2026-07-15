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
Pass/Reject/Reopen) (E4). E5 (Polish) ist fertig: Toast-System (inkl.
Konflikt-sticky), Help-Overlay `?`, Yank `y` (OSC52+nativ, Bean-/Epic-Kontext
+ Review-Stand), Maus (Wheel/Klick/Doppelklick), Settings
(`~/.config/beans-tui/`), Lobby V1 + Repo-Picker `p` (Watcher-Lifecycle-
Switch) und die Archiv-Sicht (completed/scrapped default-aus, togglebar).
E6 (Validierung & Release) ist offen — offener Stand: `beans list --ready`.

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

## Keybindings (Stand E5)

| Taste | Aktion |
|---|---|
| `↑`/`i`, `↓`/`k` | Cursor (Tree/Backlog) bzw. Section-/Feld-Cursor im Detail-Fokus |
| `→`/`l` | Knoten expandieren; im Detail-Fokus: Beziehungen-Section → Feld-Ebene rein |
| `←`/`j` | Knoten einklappen; im Detail-Fokus: Feld-Ebene raus, danach Detail-Fokus verlassen |
| `enter` | Öffnen/Bestätigen; im Detail-Fokus auf einer Beziehung: dorthin springen |
| `tab` | Fokus-Tausch Tree/Backlog ↔ Detail-Accordion |
| `1`–`4` | Detail-Accordion: direkter Section-Sprung (Meta/Body/Beziehungen/Historie) |
| `/` | Suche — lokaler Live-Filter (Titel), ab 3 Zeichen zusätzlich Bleve (Titel+Body) |
| `f` | Facetten-Filter öffnen (Status/Type/Priority/Tag/Archiv, siehe unten) |
| `X` | Facetten-Filter zurücksetzen (Tree und Backlog) |
| `b` | Backlog-View (parentlose+ready beans, geteilter Such-/Filter-Zustand) |
| `S` | Backlog: Sort-Toggle, zyklisch status → priority → created → updated |
| `s` | Status/Type/Priority-Menü (kombiniert, ein Key für alle drei) |
| `t` | Tag-Picker (Toggle-Multi-Select, Zähler, Freitext-Neuanlage) |
| `a` | Parent-Picker (Zyklen-Ausschluss + Typ-Hierarchie) |
| `B` | Blocking-Picker (Toggle-Multi-Select) |
| `c` | Bean anlegen (huh-Formular, Confirm-Gate) |
| `e` | Titel bearbeiten (Formular, direkt ohne Confirm) |
| `ctrl+e` | Body im `$EDITOR` bearbeiten (`$VISUAL` → `$EDITOR` → `vi`, Settings-`editor` hat Vorrang vor beiden — siehe Settings unten) |
| `d` | Löschen (Confirm, Kinder-/Verknüpfungs-Warnung — kaskadiert nicht) |
| `y` | Yank — Bean-/Epic-Kontext (Markdown, inkl. Children-Tabelle bei Epic/Milestone) in die Zwischenablage (OSC52 + nativ), Toast bestätigt. Im Review-Cockpit stattdessen der ganze Review-Stand (siehe unten) |
| `p` | Repo-Picker/Lobby öffnen (von überall), letztes Repo persistiert (`~/.config/beans-tui/state.json`) |
| `?` | Help-Overlay (aus der Keymap generiert), `esc`/`?`/`q` schließt |
| `ctrl+r` | Daten neu laden |
| `ctrl+k`/`K` | Command-Center öffnen (fuzzy Aktionen + Bean-Suche, von überall außer aus einem offenen Overlay/Formular/Filter/Suchfeld heraus) |
| `R` | Review-Cockpit öffnen (PO-Merge-Gate, siehe unten) |
| `q` | Quit (mit Confirm) |
| `ctrl+c` | Sofort-Quit |
| Maus: Wheel | Cursor der aktiven View bewegen (Tree/Backlog/Review-Cockpit — kein Scroll-Offset, der Cursor folgt dem Render automatisch) |
| Maus: Klick | Cursor auf die geklickte Zeile setzen; bei einem expandierbaren, noch geschlossenen Tree-Knoten expandiert der Klick direkt |
| Maus: Doppelklick | Auf einem bereits offenen, expandierbaren Tree-Knoten (<500ms zweiter Klick) klappt ihn ein — ein Einzelklick auf einem offenen Knoten kollabiert NICHT, nur Cursor (devd-D03-Semantik) |
| Maus: Klick auf Toast | Dismisst den Toast sofort, hat Vorrang vor jedem offenen Formular/Overlay |

Archiv-Toggle liegt bewusst als Facetten-Zeile ("Archivierte einblenden") im
`f`-Menü statt als eigener Key — Toggle mit `space`/`x` wie jede andere
Facette. Default aus: `completed`/`scrapped` Beans sind ausgeblendet (auch
archivierte), togglebar sichtbar.

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
| `y` | Yank — der GANZE Review-Stand (alle to-review-Gruppen + Rework-Sektion als Markdown-Tabellen), nicht nur der cursorierte Bean |
| `esc`/`q` | Zurück zur Browse-Ansicht |
| `ctrl+k`/`K` | Command-Center öffnet trotzdem (Capture-Order-Ausnahme) |

## Settings

Konfiguration liegt unter `~/.config/beans-tui/`:

- `config.yaml` — `repos:` (Liste von Repo-Pfaden für die Lobby), `editor:`
  (leer = `$VISUAL` → `$EDITOR` → `vi`; gesetzt gewinnt IMMER vor beiden
  Umgebungsvariablen), `theme.accent:` (Hex `#rrggbb`, leer = eingebautes
  Mauve), `layout.tree_width:` (Baumbreite-Floor, 24–60).
- `state.json` — Laufzeit-Zustand, aktuell nur das zuletzt geöffnete Repo
  (persistiert bei jedem Repo-Wechsel, siehe Lobby unten).

Formular via Command-Center (`ctrl+k` → "settings: öffnen"): Editor/Akzent/
Baumbreite wirken beim Speichern SOFORT (kein Neustart nötig), `repos`
wirkt beim nächsten Öffnen der Lobby (`p`).

## Lobby + Repo-Picker (`p`)

`p` öffnet von überall die Lobby: Suchfeld + Liste der in `config.yaml`
konfigurierten Repos (Anzeige "offen/gesamt" je Repo, asynchron ermittelt).
`enter` wechselt das Repo — der alte fsnotify-Watcher wird gestoppt, ein
neuer für das neue Repo gestartet (Datei-Änderungen im vorherigen Repo lösen
danach KEIN Reload mehr aus), das zuletzt gewählte Repo landet in
`state.json`.

**Start-Trigger:** ein explizites `bt <pfad>`-Argument gewinnt immer;
ansonsten direkt ins Repo, wenn `.beans.yml` vom cwd aufwärts gefunden wird
(unverändertes E1-Verhalten); nur wenn beides fehlschlägt UND `repos:`
mindestens 2 Einträge hat, öffnet sich die Lobby beim Start — bei 0/1
konfigurierten Repos bleibt die bisherige Fehlermeldung/Direktstart.

## Known Issues

- **Picker zeigen bewusst alle gültigen Relationsziele:** Parent-Picker
  (`a`) und Blocking-Picker (`B`) filtern NICHT nach Status/Archiv-Sichtbarkeit
  — auch archivierte/`completed`/`scrapped` Beans bleiben als Relationsziele
  wählbar, solange sie typ-/zyklen-gültig sind. Bewusste v1-Design-
  Entscheidung (kein Fix), da Beziehungen zu bereits abgeschlossenen Beans
  legitim bleiben (z.B. "blocked by" ein fertiges Bean). Vorbestehend seit E3.
- **Lobby-Repo-Metriken laufen nicht kontext-gecancelt:** in-flight
  Metrik-Abfragen (`beans list` je konfiguriertem Repo) eines vorherigen
  Lobby-Öffnens laufen bei erneutem Öffnen weiter — redundante, pfadgekeyte
  Subprozesse, keine Datenverwechslung. Für v1 akzeptiert; bei vielen
  konfigurierten Repos später ein Kontext-Cancel nachziehen.

## Entwicklung

TDD (`superpowers:test-driven-development`), Ausführung `make test`
(`command go test ./...`). Konventionen (Build immer `command go …`,
Datei-Namensschema, Theme-Token, Commit-/Review-Flow) → [`CLAUDE.md`](CLAUDE.md).
