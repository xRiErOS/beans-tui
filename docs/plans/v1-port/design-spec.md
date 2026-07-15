# beans-tui — Design-Spec (Port DevDash-Cockpit → beans)

> Brainstorming-Deliverable (superpowers, autonomer Modus). Quelle der Wahrheit für den
> Realisierungs-Plan (`implementation-plan.md` daneben). Stand 2026-07-14.

## 1. Kontext & Ziel

Die DevDash-TUI (`~/Obsidian/tools/DeveloperDashboard/apps/cli-go`, Go/bubbletea, „dd") ist
Eriks PO-Cockpit für das DevDashboard. Ihre Anforderungen sind im Knowledge Catalogue
dokumentiert (Bundle `devdash-tui`: Capability-Inventar, Screen-Mockups, Entity-Feldmodelle,
Hotkey-Map D04). DevDash ist Klasse-C-Legacy; neue Arbeit läuft über **beans** (lean-stack).

**Ziel:** Eine eigenständige TUI `bt`, die Visualisierung und Funktionalität der devd-TUI so
vollständig wie möglich auf beans überträgt — PO-Cockpit für jedes beans-Repo: Beans ansehen,
bearbeiten, Review-Flow fahren, besser mit Agenten interagieren.

**Leitbild (aus PO-Spec devdash-tui):** „Obsidian-Command-Palette. Fuzzy-Eintippen →
deskriptiver Vorschlag → Drill-down." Keyboard-first, PO-exklusive Aktionen (Schließen von
Beans = Abnahme) laufen über die TUI.

## 2. Quellen (Recherche-Grundlage)

| Quelle | Inhalt |
|---|---|
| `DeveloperDashboard/apps/cli-go` | Referenz-Implementierung: 14 Views, Keymap (jkli + Leader), Catppuccin-Macchiato-Theme, Accordion/Master-Detail/Overlay-Primitive, huh-Forms, Review-Cockpit |
| KC-Bundle `devdash-tui` | Anforderungen: MVP-Capability-Inventar, 5 Screen-Mockups, Rollenteilung Agent↔PO, NFRs (keyboard-only, konfliktfreie Hotkeys, constrained Edits — „no 422") |
| `beans-src` + Binary 0.4.2 | Datenmodell (Bean: title/status/type/priority/tags/parent/blocking/blocked_by/order/body), öffentliche `pkg/`-Library (beancore mit fsnotify-Watcher, beangraph CoreResolver, config mit Farb-Enums), Apache-2.0 |

## 3. Architektur

### 3.1 Stack & Datenlayer

- **Go 1.26 + bubbletea/bubbles/lipgloss/huh/glamour** — identisch zur devd-TUI. Der
  UI-Code der devd-TUI (Eriks eigener Code) wird als Port-Basis übernommen; der Datenlayer
  wird ausgetauscht.
- **Datenlayer: beans-CLI als Subprocess.** Verifiziert: beans v0.4.2 (letztes Release =
  installiertes Binary) exponiert KEINE importierbaren `pkg/`-Packages (alles `internal/`;
  die `pkg/`-Struktur existiert nur im ungetaggten Dev-Klon). Daher: Reads via
  `beans list --json --full` (+ In-Memory-Index in der TUI), Volltext via `beans list -S`
  (Bleve), Mutationen via `beans create/update/delete --json` mit `--if-match`
  (optimistic locking), Fallback für berechnete Felder: `beans graphql` (CLI, ohne Server).
  Vorteil: das beans-Binary bleibt die eine Autorität (lean-stack-Prinzip), TUI ist
  versions-entkoppelt; Subprocess-Latenz (~20-50 ms) für TUI unkritisch.
- **Konsistenz:** ETag aus `--json`-Output; Formular-Saves senden `--if-match`
  (Konflikt → Toast + Reload).
- **Live-Reload:** eigener fsnotify-Watcher auf `.beans/` (+ `archive/`), Debounce ~150 ms
  → Reload über CLI. Externe Änderungen (Agent, `beans update` in anderem Terminal)
  erscheinen ohne Neustart. Kern-Feature für „besser mit Agenten interagieren".

### 3.2 Projekt / Repo

```
tools/beans-tui/
  beans-tui-worktrees/     (Konvention; Weiche: main-direkt, s. SSTD)
  beans-tui-repository/    Modul `beans-tui`, Binary `bt`
    main.go  cmd/  internal/tui/  internal/theme/  internal/config/  internal/clip/
    docs/ (plans/, SSTD.md)   .beans/ (Dogfooding: eigene Entwicklung als beans)
```

- **Start:** `bt` in beliebigem beans-Repo — `.beans.yml`-Discovery aufwärts vom cwd
  (wie beans-CLI). `bt <pfad>` öffnet explizites Repo. Build `make build` → `bin/bt`,
  Install `command go install .`.
- **Repo-Picker (Lobby):** Liste konfigurierter Repos aus `~/.config/beans-tui/config.yaml`
  (`repos:`-Liste). Ohne Config oder mit cwd-Treffer: direkt ins Repo (Lobby übersprungen).
  Ersetzt devd-Projekt-Picker funktional 1:1.
- **Dogfooding:** Die Entwicklung von beans-tui wird in `.beans/` dieses Repos getrackt —
  die TUI zeigt ab E2 ihre eigene Entwicklung.

### 3.3 Übernommene Architektur-Patterns (devd-TUI)

Elm (Value-Receiver-Model, `Update`-Dispatcher, reine `View()`) · `viewID`-Enum +
Rückkehr-Pointer statt Stack · Overlay-Compositing (`viewComposite`, ANSI-Splicing,
Painter's-Algorithmus) · Datei-Namenskonvention `<art>_<verb>_<entität>.go` ·
Message/Command-Trennung (`messages.go` nur Msg-Typen + Cmd-Producer) · zentrale Keymap
(Single Source → Help-Overlay + `docs/shortcuts.md`) · Zwei-Ebenen-Fokus (Section↔Feld,
Accordion) · `listState`-Primitiv · Settings zweischichtig (User-YAML + Runtime-State-JSON).

## 4. Entity-Mapping devd → beans

| devd | beans | Behandlung |
|---|---|---|
| Project | beans-Repo (`.beans.yml`, `project.name`) | Repo-Picker statt Projekt-Picker |
| Milestone | bean `type: milestone` | 1:1 (Tree-Wurzel) |
| Sprint | bean `type: epic` | 1:1 — lean-stack: „Sprint ≡ Epos ≡ bean" |
| Issue (bug/feature/improvement/core) | bean `type: bug/feature/task` | Tree-Blätter; `improvement/core` → `task` |
| Backlog (Issues ohne Sprint) | beans ohne `parent` bzw. `--ready` | Backlog-View |
| Subtasks | Markdown-Checkboxen im Body | Anzeige im Detail; Toggle v2 |
| DoD (Milestone) | Markdown-Checkliste im Body | wie Subtasks |
| Dependencies | `blocking` / `blocked_by` | 1:1, Picker vorhanden (Upstream-Muster) |
| Tags | beans-Tags (implizit, ohne Farbregister) | Tag-Picker mit Nutzungszählern; Farbe per Hash aus fester Palette; Tag-Manager-CRUD entfällt (Tags entstehen/sterben implizit) |
| Status-Lifecycle Issues (new→refined→planned→…) | beans-Status `draft·todo·in-progress·completed·scrapped` | Status-Menü zeigt nur beans-Enum (constrained, „no 422"-Prinzip analog) |
| Review-Flow (`to_review`, Verdicts) | **Tag-Konvention** (s. § 5) | Kern-Flow bleibt erhalten |
| User Stories je Issue | `## User Stories`- / `## Akzeptanz`-Checkboxen im Body | Review-Modal rendert Body-Abschnitt |
| project_memories | — entfällt | lean-stack-Autorität: Wissen → OKF, nicht beans |
| Docs / User-Notes | — entfällt | Docs leben im Repo/OKF |
| ToDos | — entfällt | beans-Tasks decken das ab |
| SSTD-Slots | — entfällt | SSTD ist Datei im Repo |

**Status-Farb-/Glyph-Mapping** (devd-Look, beans-Semantik): `draft` Blue · `todo` Text ·
`in-progress` Yellow · `completed` Green · `scrapped` Red — Glyph `◉` einheitlich
(Bedeutung = Farbe, DD2-176), Type-Icons: milestone `⬢` Peach · epic `✦` Mauve ·
feature `+`/`✦` Green · task `⯅` Blue · bug `⯁` Red. ASCII-Fallback via `BT_ASCII_ICONS`.
**Superseded (PF-6, §15, 2026-07-15):** Typ/Status werden nicht mehr über geometrische
Icons bzw. den gemeinsamen `◉`-Glyph unterschieden, sondern über farbige Buchstaben —
verbindliche Tabelle siehe §15.

## 5. Review-Flow-Konvention (Agent↔PO)

**Superseded (PF-14, §15, 2026-07-15 — Review-Cockpit ENTFERNT):** die Original-
Konvention dieses Abschnitts (Zeilen unten, als historischer Kontext belassen) sah
ein eigenes TUI-View (V6, §6) für den Review-Flow vor. PO-Entscheidung (Nachtrag 8):
„widerspricht dem lean-stack-Wesen und schafft wieder Zeremonie" — Review findet
künftig direkt im Chat statt, die TUI zeigt nur noch Tag-Sichtbarkeit (Tree/Detail/
Filter), keine eigene Review-Interaktion mehr. **Aktuelle, verbindliche Konvention:**

| Schritt | Wer | beans-Operation |
|---|---|---|
| Arbeit fertig, Abnahme angefragt | Agent | Tag `to-review` setzen, Status bleibt `in-progress` |
| Review-Sichtbarkeit | PO (TUI, passiv) | beans mit Tag `to-review` erscheinen wie jeder andere Tag im Tree/Detail, auffindbar via Filter/Suche — KEINE eigene TUI-Interaktion, kein View |
| Pass | PO (Chat/CLI) | Tag `to-review` → `accepted` (Chat-gesteuert, `beans update --tag`) |
| Reject | PO (Chat/CLI) | Tag `to-review` → `rejected`; Feedback landet im Chat bzw. als Body-Abschnitt (Chat-Konvention, nicht TUI) |
| Rework fertig | Agent | Tag `rejected` → `to-review` |

Tag-Trio (kebab-case, ersetzt die frühere Zwei-Tag-Konvention `to-review`/`rework`):
**`to-review`** (Agent meldet fertig) · **`accepted`** (PO nimmt an) · **`rejected`**
(PO weist zurück → Agent-Rework). Die TUI trägt hier keine Autorität mehr — sie ist
reines Sichtbarkeits-Fenster auf die drei Tags, identisch zu jedem anderen Tag.

<details>
<summary>Historischer Kontext: ursprüngliche TUI-Cockpit-Konvention (E4, bis E7/PF-14)</summary>

beans kennt keinen Review-Status. Ursprüngliche Konvention (lean-stack-konform: „der
ausführende Agent schließt NICHT", PO-Autorität am Gate):

| Schritt | Wer | beans-Operation |
|---|---|---|
| Arbeit fertig, Abnahme angefragt | Agent | Tag `to-review` setzen, Status bleibt `in-progress` |
| Review-Queue ansehen | PO (TUI) | Review-Cockpit: alle beans mit Tag `to-review`, gruppiert nach Epic |
| Pass | PO (TUI) | Status → `completed`, Tag `to-review` entfernen |
| Reject | PO (TUI) | Tag `to-review` → `rework`; Kommentar wird als `## Review <datum>`-Abschnitt an Body angehängt |
| Rework fertig | Agent | Tag `rework` → `to-review` |

Die TUI war damit das Merge-Gate-Cockpit: Agenten liefern in `.beans/`, der PO nimmt in
der TUI ab — Live-Reload macht Agent-Fortschritt in Echtzeit sichtbar. Per PF-14
(E7) entfernt — s. oben für den aktuellen Stand.

</details>

## 6. Views (Port-Matrix)

| # | beans-tui-View | devd-Vorlage | Inhalt |
|---|---|---|---|
| V1 | Lobby | `viewHome` | ASCII-Logo `beans` + Repo-Picker (Suchfeld + Tabelle Name·Pfad·Offen/Gesamt); entfällt bei Einzel-Repo-Start |
| V2 | Browse (Primat-View) | `viewBrowseProject` | Master-Detail: links Tree Milestone→Epic→Task (expandierbar, Such-/Filterkopf), rechts Detail-Accordion; Fokus-Tausch mit Border-Farbwechsel |
| V3 | Backlog | `viewBrowseBacklog` | Flache Liste parentloser/ready beans, Master-Detail, Sort (status/prio/created/updated) |
| V4 | Detail-Accordion | `accordion.go` + `viewDetailIssue` | Sections: Meta (Status/Type/Prio/Tags) · Body (glamour) · Beziehungen (Parent/Children/Blocking/BlockedBy) · Historie (created/updated/ETag); Ziffern-Sprung `1…9` |
| V5 | Command-Center | `viewCommandCenter` + `overlay_palette` | `ctrl+k`: fuzzy Aktionen (Create, View-Wechsel, Status…) + Bean-Suche (Bleve) in einem |
| V6 | ~~Review-Cockpit~~ **ENTFERNT (PF-14, §15)** | `viewReviewSprint` + `viewNavigateReviews` | War: Queue `to-review` (links, Verdikt-Dots), Detail rechts, Summary-Zeile „x of n"; `a` pass · `x` reject+Kommentar · `o` reopen. Review läuft jetzt im Chat, TUI zeigt nur noch Tag-Sichtbarkeit (§5) |
| V7 | Create/Edit-Forms | `form_create_*`/`form_edit_*` (huh) | Titel/Typ/Prio/Status/Parent/Tags/Body; nur gültige Werte angeboten; Confirm-Gate vor Create; `ctrl+e` Body im `$EDITOR` |
| V8 | Overlays | `box_*`/`picker_*`/`overlay_*` | Status-/Type-/Prio-Menü, Tag-/Parent-/Blocking-Picker, Delete-Confirm (mit Kinder-Count), Quit-Confirm, Toast, Help (`?`) |

## 7. Keymap (D04-Prinzipien, devd-Belegung übernommen)

Global: `↑/i ↓/k ←/j →/l` Richtungskreuz (Pfeile immer Alias) · `enter` open/confirm ·
`esc` back · `q`/`ctrl+c` quit-Confirm · `?` Help · `ctrl+k`/`K` Command-Center ·
`p` Repo-Picker · `R` Review-Cockpit · `b` Backlog · `/` Suche · `f` Filter · `X` Filter-Reset ·
`ctrl+r` Reload · `1…9` Accordion-Sprung · `y` Yank.
Node-fokussiert: `s` Status-Menü · `t` Tag-Picker · `a` Parent-Zuweisung (Assign) ·
`B` Blocking-Picker · `c` Create · `d` Delete-Confirm · `e`/`ctrl+e` Editor.
Review (View-lokale Overrides, wie devd): `a` pass · `x` reject · `o` reopen · `y` Review-Stand
yanken · `n/p` next/prev (Repo-Picker im Review via Palette).
Keine `ctrl+s`/`ctrl+q`-Belegung (XOFF/XON). Single-Keys inaktiv bei fokussierter Texteingabe.
**Erweitert (PF-5, §15, 2026-07-15, REVIDIERT durch PO-Nachtrag 3):** `enter` erhält
INNERHALB eines bereits aktiven Detail-Fokus (Einstieg bleibt `tab`, unverändert)
zusätzliche Bedeutung (Feld-Navigation, Edit-Overlay-Kaskade) — kein neues
`keyMap`-Feld, reine Verhaltenserweiterung des bestehenden `Enter`-Bindings, siehe §15.

## 8. Visual Design

Catppuccin Macchiato 1:1 aus `theme.go` der devd-TUI (TrueColor-Hex, Single-Source-Token):
Base `#24273a` · Mantle `#1e2030` · Mauve-Akzent `#c6a0f6` (aktiver Border/Cursor/Header) ·
Peach `#f5a97f` Chevrons · Sapphire `#7dc4e4` IDs · Hint `#7c7c7c` Zwei-Klassen-Text ·
Overlay `#8087a2` inaktive Borders. RoundedBorder-Außenrahmen, vierzonige Hülle
(Breadcrumb-Header `> repo: Titel` + Shortcuts rechts · optionales Info-Grid · Master-Detail-
Panes 1fr:2fr · Footer lokale Hints + Statuszeile). Modale zentriert schwebend, Toast oben
rechts. `lipgloss.SetHasDarkBackground(true)`. huh-Theme mit Mauve/Yellow-Akzent.
beans-Farb-Enums aus `pkg/config` werden auf die Catppuccin-Token gemappt (nicht die
16-Farben-ANSI-Namen von beans upstream).

## 9. Scope

**IN (v1):** V1–V8 komplett · Suche (Bleve über Library) · Facetten-Filter (Status/Type/
Prio/Tag) · Mutationen (Create/Edit/Status/Tags/Parent/Blocking/Delete) · Review-Flow §5 ·
Live-Reload · Yank (OSC52 + nativ) · `$EDITOR`-Integration · Toast · Help-Overlay ·
Maus (Wheel, Klick-Cursor) · Settings (`~/.config/beans-tui/config.yaml`: repos, editor,
Akzentfarbe, Baumbreite) · ASCII-Icon-Fallback · Archiv-Sicht (completed/scrapped ein-/
ausblendbar; `.beans/archive/` wird mitgelesen).

**OUT (v1, bewusst):** Tag-Manager-CRUD (Tags implizit) · Memory/Docs/Notes/ToDos-Views
(Entitäten entfallen) · Tutorial · Release-Notes-Overlay · Subtask-Checkbox-Toggle im
Detail (v2) · Reorder via fractional index (v2) · GraphQL-Server/Subscriptions ·
Multi-Select-Bulk-Ops (v2) · Agent-Session-Features des beans-Web-UI (Chat, Worktrees,
Terminals) — die TUI interagiert mit Agenten über Daten (beans + Tags), nicht über
Prozess-Steuerung.

## 10. User Stories (Validierungs-Grundlage)

Abnahme v1 = alle Stories validiert erfüllt (Nachweis je Story im Validierungs-Epic).

| ID | Story | Akzeptanzkriterien (Kern) |
|---|---|---|
| US-01 | Als PO starte ich `bt` in einem beans-Repo und sehe sofort den Projektbaum. | Start < 2 s bei 100 beans; `.beans.yml`-Discovery aufwärts; ohne Repo: klare Fehlermeldung + Hinweis |
| US-02 | Als PO navigiere ich vollständig ohne Maus. | jkli+Pfeile, enter/esc, Fokus-Tausch Tree↔Detail sichtbar (Border), `1…9`-Accordion, alle Aktionen erreichbar |
| US-03 | Als PO sehe ich zu jedem Bean alle Details. | Titel, Status/Type/Prio/Tags farbcodiert, Body als gerendertes Markdown, Parent/Children/Blocking/BlockedBy navigierbar (enter springt) |
| US-04 | Als PO erreiche ich jede Aktion über die Command-Palette. | `ctrl+k`, fuzzy, Aktionen + Bean-Treffer gemischt, enter dispatcht, kontextabhängige Einträge zuerst |
| US-05 | Als PO finde und filtere ich beans. | `/` Live-Suche (Titel+Body, Bleve-Syntax), `f` Facetten Status/Type/Prio/Tag, `X` Reset, Filter wirken auf Tree UND Backlog |
| US-06 | Als PO lege ich ein Bean ohne Terminal-Wechsel an. | `c`/Palette → huh-Form (Titel Pflicht, Typ/Prio/Status/Parent/Tags), nur gültige Enum-Werte, Confirm-Gate, danach Cursor auf neuem Bean |
| US-07 | Als PO bearbeite ich jedes Bean-Feld constrained. | `s` Status-Menü (nur beans-Enum), Prio/Type-Menü, `t` Tags, `a` Parent (Zyklen ausgeschlossen), `B` Blocking, Body via `$EDITOR`, ETag-Konflikt → Toast+Reload |
| US-08 | Als PO sehe ich den Review-Stand meiner Agenten. **(Redefiniert PF-14, §15 — war „fahre ich den Review-Flow mit Agenten" via eigenem Cockpit-View)** | beans mit Tag `to-review`/`accepted`/`rejected` sind über Tree/Detail/Filter/Suche auffindbar wie jeder andere Tag (KEIN eigenes View, KEINE TUI-Interaktion); Abnahme (`to-review`→`accepted`/`rejected`) erfolgt im Chat bzw. via `beans update --tag` (CLI), nicht in der TUI |
| US-09 | Als PO sehe ich den Backlog priorisiert. | `b`: parentlose+ready beans, Sortierung umschaltbar, Master-Detail |
| US-10 | Als PO sehe ich Agent-Änderungen live. | Externe `.beans/`-Änderung (CLI/Agent) erscheint < 1 s ohne Neustart, Cursor-Position bleibt stabil |
| US-11 | Als PO übergebe ich Kontext an Agenten per Clipboard. | `y` kopiert Bean/Epic-Kontext als Markdown (ID, Titel, Status, Body) via OSC52; Toast bestätigt |
| US-12 | Als PO erlebe ich den devd-Look. | Catppuccin-Token §8, vierzonige Hülle, Master-Detail-Borders mit Fokus-Farbe, Accordion mit Chevron/Ziffern, Statusglyph `◉` (superseded PF-6/§15: Typ-/Status-Buchstaben statt Icon/`◉`) |
| US-13 | Als PO sehe ich jederzeit meine Möglichkeiten. | `?`-Overlay aus zentraler Keymap generiert; Footer zeigt lokale Hints je View |
| US-14 | Als PO wechsle ich zwischen mehreren beans-Repos. | `p`-Picker aus Config-`repos:`; letztes Repo persistiert (state.json); `bt <pfad>` direkt |

## 11. Teststrategie (devd-Konvention übernommen)

- **TDD verpflichtend** (Anforderung aus KC-Plan): Logik zuerst als Test.
- **Update-Tests:** `tea.KeyMsg`-Sequenzen gegen Model-State (kein Terminal nötig).
- **Golden-View-Snapshots:** `View()`-Output gegen `testdata/*.golden` (Update via `-update`-Flag); TrueColor-Guard-Test gegen Farb-Misalignment.
- **Datenlayer-Tests:** gegen tmp-Verzeichnis mit generierten `.beans/`-Fixtures (Library in-process — kein Mock-HTTP nötig).
- **Abnahme:** je User Story ein dokumentierter Validierungs-Schritt (Test oder belegtes manuelles Skript via tmux/teatest).

## 12. Epen (Umsetzungs-Wellen, je Epos NSP-Auto-Handover)

| Epos | Inhalt | Stories |
|---|---|---|
| E1 Foundation | Modul, Makefile, beans-Library-Anbindung (Spike: Core laden/mutieren/watchen), Theme+Chrome+Keymap-Port, App-Shell, Read-only-Tree | US-01, US-12 (Basis) |
| E2 Browse & Detail | Tree komplett (expand/lazy), Detail-Accordion, glamour-Body, Beziehungs-Navigation, Suche+Filter, Backlog | US-02, US-03, US-05, US-09 |
| E3 Mutationen | Create-/Edit-Forms (huh), Status-/Type-/Prio-Menüs, Tag-/Parent-/Blocking-Picker, Delete-Confirm, `$EDITOR`, ETag-Handling | US-06, US-07 |
| E4 Palette & Review | Command-Center, Review-Cockpit inkl. Tag-Konvention + Body-Feedback (Review-Cockpit-Teil per PF-14/E7 wieder entfernt, s. §5/§15) | US-04, US-08 |
| E5 Polish | Live-Reload, Yank/OSC52, Toast, Help, Maus, Settings+Repo-Picker+Lobby, ASCII-Fallback, Archiv-Sicht | US-10, US-11, US-13, US-14 |
| E6 Validierung & Release | Alle US validiert (Nachweise), README, Install-Doku, `bt` in lean-stack demo-startbar | Abnahme aller US |

## 13. Risiken

| Risiko | Behandlung |
|---|---|
| beans-CLI-Flags/JSON-Shape ändern sich bei Upgrade | Version-Guard beim Start (`beans version`, getestet gegen 0.4.x); Datenlayer als einzige CLI-Berührungsfläche (ein Package) |
| devd-Code-Port zieht API-Client-Reste mit | Port-Reihenfolge: Theme/Primitive zuerst, Views einzeln, Datenlayer von Anfang an beans-nativ (kein Adapter auf devd-Typen) |
| Subprocess-Latenz spürbar bei Massen-Mutationen | Reads gebündelt (ein `list --json --full` je Reload), Mutationen einzeln + optimistisches UI-Update vor Reload |
| Watcher-Races (eigene Mutation triggert Reload) | Debounce ~150 ms; eigene Writes markieren+ignorieren, Cursor-Restore testen (US-10) |

## 14. Entscheidungen

| Code | Hintergrund | Entscheidung | Status |
|---|---|---|---|
| D01 | Stack-Wahl | Go/bubbletea, devd-TUI-Code als Port-Basis, kein Neuaufbau in Ink/TS | 🟢 |
| D02 | Datenzugriff | beans-CLI-Subprocess (`--json`/`--if-match`/`-S`), eigener fsnotify-Watcher — kein Library-Import möglich (v0.4.2 nur `internal/`), kein Fork, kein HTTP. Revidiert nach Modul-Verifikation | 🟢 |
| D03 | Projekt-Ort | Eigenes Repo `tools/beans-tui/beans-tui-repository`, Binary `bt`, main-direkt (Worktree-Weiche, Solo+sequentielle Agent-Kette) | 🟢 |
| D04 | Review-Abbildung | Tag-Konvention `to-review`/`rework` + Body-Feedback-Abschnitte; PO schließt (completed) exklusiv via TUI | 🟢 |
| D05 | Entitäten-Reduktion | Memories/Docs/Notes/ToDos entfallen (Autoritäts-Trennung lean-stack: Wissen→OKF, Docs→Repo) | 🟢 |
| D06 | Tracking | Dogfooding: Entwicklung in `.beans/` dieses Repos; lean-stack erhält Verweis-bean | 🟢 |
| D07 | Vorherige „no custom TUI"-Entscheidung (`po-immersion-beans-via-obsidian-bases-no-custom-tui`, KC agent-memory) | Durch expliziten User-Auftrag superseded — KC-Konzept nach Abschluss via `/okf` aktualisieren | 🟢 |

## 15. PO-Feedback R1 (2026-07-15)

PO-Feedback aus visueller QA (Screenshots `docs/_free-notes/vqa-2026-07-15/`, Epic
`bt-heg9`). Spec-ändernd: die folgenden 6 Punkte (PF-1…PF-6) sind ab sofort
verbindliche Design-Entscheidungen — implementiert VOR den E6-Validierungs-Tasks
(`bt-wm4w`/`bt-9yvh`, `blocked_by: bt-heg9`), damit E6 gegen den neuen Stand
validiert statt gegen einen veralteten. Q01 (Header-Inhalt) und Q02
(Farb-/Zeichen-Details) waren zunächst offen, sind aber durch die PO-Antworten vom
2026-07-15 (bean `bt-heg9`, Abschnitt „PO-Antworten") vollständig gelöst — keine
offenen Q-Marker mehr in diesem Abschnitt.

**PF-1 — Meta-Sektion `[1]` nicht kollabierbar (Planner-Entscheidung, devd-Blick).**
Von den zwei durch die PO offengelassenen Interpretationen (`nicht kollabierbar` vs.
`Default-offen`) wird **nicht kollabierbar** gewählt: `renderAccordion`
(`accordion.go`) rendert Sektion 1 (Meta) IMMER mit Body, unabhängig vom
`open`-Parameter (`isOpen := n == open || n == 1`), während Sektionen 2-4
(Body/Beziehungen/Historie) weiterhin exklusiv-offen bleiben (max. eine zusätzliche
Sektion gleichzeitig sichtbar). Begründung: devds Referenzarchitektur
(`view_detail_issue.go`, `detailTitle + Meta-Strip`) zeigt Titel/Typ/Priorität als
permanent sichtbaren, nicht-kollabierbaren Banner oberhalb der eigentlichen
Accordion-Sektionen — genau dieses Prinzip (immer sichtbar, unabhängig von der
Sektionsnavigation) erfüllt die PO-Formulierung „immer offen" wörtlich;
„Default-offen" würde beim Navigieren zu Sektion 2-4 wieder kollabieren (das ist
bereits der Status quo — `accOpen` startet bei 1) und die Anforderung damit NICHT
über den Status quo hinaus verbessern. Ziffer `1` bleibt ein gültiger Cursor-Sprung
(`secCursor`/`accOpen = 1`) für die Feld-Navigation innerhalb Meta, ist aber für die
Sichtbarkeit selbst redundant.

**PF-2 — Zifferntasten 1-9 vereinheitlicht (Browse-Detail ↔ Review-Cockpit).**
`keyReviewCockpit`s bisheriges Toggle-Verhalten (zweites Drücken derselben Ziffer
setzt `reviewAccOpen` auf 0 = alles zu) entfällt — ersetzt durch denselben reinen
Direktsprung wie `keyDetailFocus` im Browse-Detail (`accOpen = d`, kein Toggle-zu-0).
Mit PF-1 (Meta nicht kollabierbar) ist `accOpen`/`reviewAccOpen == 0` ohnehin ein
legitimer, sinnvoller Ruhezustand („nur Meta sichtbar, keine der Sektionen 2-4
zusätzlich offen") — kein Sonderfall mehr nötig. Zusätzlich wird der
Ziffern-Bereichs-Check (bisher an zwei Stellen unabhängig hart auf `'4'` kodiert:
`update.go::keyDetailFocus` und `view_review_cockpit.go::keyReviewCockpit`) auf eine
gemeinsame Konstante `beanSectionCount = 4` (`view_detail_bean.go`, neben
`beanSections`) umgestellt — beide Stellen vergleichen künftig gegen dieselbe
Konstante statt gegen zwei unabhängige Literale (Robustheit für eine künftige 5.
Sektion: nur EINE Stelle muss nachgezogen werden). `keyReviewCockpit` berechnet
`beanSections` nicht selbst (kein Bean/Index dafür nötig, geringerer Umbau als
`len(secs)` dort zu ermitteln) — die Konstante ist der einfachere, gleichwertige
Weg zum selben Robustheitsziel. Digits 5-9 bleiben (aktuell 4 Sektionen)
wirkungslose No-Ops, wie schon heute implizit der Fall.

**PF-3 + PF-4 — Detail-Kopfblock + Meta-Feldliste** (verschmolzen laut PO-Antwort
Q01, 2026-07-15: „identisch mit dem Kopfblock aus PF-4 — EIN Feature"). Layout des
gesamten Detail-Panels (gilt für Tree/Backlog UND Review-Cockpit, da beide über
`renderAccordionPane`, `view_browse_repo.go`, denselben Render-Pfad teilen):

```
bean-id
NAME

type: xxxx    status: xxxx    prio: xxxx

[1] META
▷ title:      xxxx
▷ status:     xxxx
▷ type:       xxxx
▷ priority:   xxxx
▷ created_at: xxxx
▷ updated_at: xxxx

[2] BODY
[3] RELATIONS
[4] HISTORY
```

`▷` = nicht selektiert, `▶` = selektiert (Akzentfarbe Mauve). Die oberen 4 Zeilen
(bean-id / NAME / Leerzeile / `type:…status:…prio:…`) sind ein NEUER, nicht
nummerierter, nicht-interaktiver Kopfblock oberhalb der Accordion-Sektionen — kein
App-Chrome-Umbau (PO-Antwort Q01 explizit: „Kein separater App-Header-Umbau",
betrifft NUR das Detail-Panel, nicht die vierzonige Hülle aus §8).
`[1] META` bleibt Sektion 1 (keine Off-by-one-Verschiebung ggü. dem bestehenden
E2-Schema) — ihr Body wandelt sich von der bisherigen 4-zeiligen Klartext-Darstellung
(`metaSectionBody`) zu einer 6-zeiligen, cursor-navigierbaren Feldliste: title /
status / type / priority / created_at / updated_at, mit Cursor-Marker `▷`/`▶` je
Zeile — `▶` nur wenn Meta die fokussierte Sektion ist (`activeIdx == 0 && focused`)
UND die Zeile dem aktiven `fieldIdx` entspricht. `created_at`/`updated_at` sind NICHT
editierbar (system-verwaltet), aber trotzdem cursor-adressierbar (PO-Mockup zeigt
`▷` auch dort) — `enter` darauf ist ein No-Op (analog zum bestehenden Muster für
unresolved relationFields, `beanID == ""`).

**PF-5 — Enter-Kaskade INNERHALB des Detail-Fokus** (REVIDIERT durch PO-Nachtrag 3,
2026-07-15 — D01 zurückgenommen: „KEIN enter als Detail-Fokus-Einstieg — der
bestehende tab-Mechanismus gefällt dem PO und BLEIBT der Einstieg"). `tab` bleibt der
EINZIGE Weg, den Detail-Fokus zu betreten — `keyTree`s heutiges `enter`-Verhalten
(Blatt-Knoten: No-Op; Knoten mit Kindern: expand/collapse-Toggle) bleibt
**komplett unverändert**. **ERRATUM/Präzisierung (T6-Review B01, bean `bt-t1uy`):**
die frühere Fassung dieses Absatzes behauptete dasselbe pauschal auch für
`keyBacklog` — FALSCH: `keyBacklog`s `enter` trug bis T6 noch das
Vor-D01-Revisions-Verhalten (Detail-Fokus-Einstieg + Cursor-Reset). Korrektur:
Backlog-`enter` ist ein handled **No-Op** (die flache Liste kennt kein
Expand-Konzept — das Analogon zu `keyTree`s Blatt-No-Op); der Einstieg ist auch dort
exklusiv `tab`. Die Kaskade gilt ausschließlich INNERHALB eines
bereits aktiven Detail-Fokus (`m.detailFocus == true`, via `tab` erreicht):

| Ebene | Taste | Wirkung |
|---|---|---|
| Detail-Fokus, Sektions-Ebene | Pfeile/i-k | Sektion wechseln (`secCursor`/`accOpen`, unverändert) |
| Detail-Fokus, Sektions-Ebene | `enter` (NEU) | Feld-Navigation der aktiven Sektion betreten (`detailLevel = 1`) — Alias zum bestehenden `→`/`l`, dieselbe Bedingung (nur wenn die Sektion Felder hat) |
| Detail-Fokus, Feld-Ebene | `enter` (Verhalten erweitert) | passendes Edit-Overlay für das aktive Feld öffnen statt wie bisher IMMER ein Beziehungs-Jump |

Da die gesamte Kaskade innerhalb von `keyDetailFocus` (`update.go`) verortet ist —
einem Handler, der ausschließlich erreicht wird, wenn `m.detailFocus == true`, selbst
hinter jedem anderen Full-Capture-Gate (Form/Overlay/Search/Filter/Palette/Help/
Lobby/ReviewCockpit) in `handleKey` — gibt es KEINE Kollision mit Tree/Backlog/
Cockpit-eigenen `enter`-Bedeutungen: die Zustandsräume sind exklusiv (entweder
`m.detailFocus` oder Tree/Backlog-Dispatch, nie beides). Kein neuer `keyMap`-Eintrag
nötig (reine Verhaltenserweiterung des bestehenden `Enter`-Bindings, `helpGroups()`
unverändert, Drift-Guard unberührt).

Feld→Overlay-Zuordnung: `status`/`type`/`priority` → der bestehende kombinierte
Value-Menu-Overlay (`box_menu_value.go`), NEU seedbar auf die passende Gruppe
(`openValueMenu(group string)`, bisher hart auf `"status"` — der einzige Call-Site in
`keyNodeAction` wird auf `m.openValueMenu("status")` angepasst); `title` →
bestehendes `openEditTitleForm` (dieselbe Form wie die `e`-Taste, unverändert);
`created_at`/`updated_at` → No-Op (s. PF-4); Beziehungen-Felder (Sektion 3) →
unverändertes Jump-Verhalten (E2, `relationField.beanID`). `e`/`ctrl+e` bleiben
UNVERÄNDERT (Titel-Formular bzw. `$EDITOR` auf Body) — explizit außerhalb dieses
Scopes („e bleibt $EDITOR", PO-Formulierung). Leitprinzip (PO-Antwort D02): Auswahl
soll „möglichst schnell/einfach" sein — die Kaskade braucht in jedem Fall ≤ 2
Tastendrücke von „Sektion sichtbar" bis „Edit-Overlay offen" (`enter`→`enter`, ggf. +
Pfeil zur Feldwahl), keine zusätzliche Zeremonie/Bestätigung.

**PF-6 — Glyphen-Ersatz (Typ/Status/Priorität)**, Q02 vollständig bestätigt
(2026-07-15).

Typ (Großbuchstabe):

| Typ | Glyph | Farbe (alt → neu) |
|---|---|---|
| milestone | `M` | Peach → **Blue** |
| epic | `E` | Mauve (unverändert) |
| feature | `F` | Green → **Mauve** |
| task | `T` | Blue → **Sky** |
| bug | `B` | Red (unverändert) |

Status (Kleinbuchstabe):

| Status | Glyph | Farbe (alt → neu) |
|---|---|---|
| draft | `d` | Blue (unverändert, Q02 bestätigt) |
| todo | `t` | Text → **Green** |
| in-progress | `i` | Yellow (unverändert) |
| completed | `c` | Green → **Subtext** |
| scrapped | `s` | Red → **Subtext** |

Priorität (Glyph statt Wort, Q02 bestätigt):

| Priorität | Glyph | ASCII-Fallback | Farbe (alt → neu) |
|---|---|---|---|
| critical | `‼` | `!!` | Red (unverändert) |
| high | `!` | `!` | Red → **Yellow** |
| normal | `·` | `.` | Text (unverändert — nicht explizit in Q02, neutraler Default beibehalten) |
| low | `↓` | `v` | Green → **Subtext** |
| deferred | `→` | `>` | Hint → **Subtext** |

Unbekannte Enum-Werte (Status/Typ) bleiben unverändert beim generischen Fallback
(`·`/`.`, Farbe Text) — von PF-6 nicht betroffen. `BT_ASCII_ICONS` bleibt für Typ/
Status wirkungslos (Buchstaben sind bereits ASCII/EAW-Neutral) — nur für Priorität
weiterhin relevant. Der bisherige DD2-176-Grundsatz („ein gemeinsamer Glyph für alle
Status, Bedeutung ausschließlich über Farbe") wird durch PF-6 EXPLIZIT aufgehoben
(PO-Direktive) — Status/Typ tragen Bedeutung jetzt redundant über Buchstabe UND
Farbe; das verbessert nebenbei Barrierefreiheit (Epic/Feature waren zuvor NUR über
Farbe unterscheidbar, jetzt zusätzlich über den Buchstaben).

**PF-7 — UI-Sprache durchgängig Englisch** (PO-Nachtrag 2, 2026-07-15). Alle
nutzerseitigen String-Literale in `internal/tui/*.go` (Produktionscode, nicht Tests)
wechseln von Deutsch/gemischt auf Englisch — Toasts, Fehlertexte, Overlay-/Menü-
Labels, Filter-Menü, Cockpit-Leerzustand, Lobby, Footer-Hints, Formular-Titel.
Accordion-Sektionstitel `Meta`/`Body`/`Beziehungen`/`Historie` → `META`/`BODY`/
`RELATIONS`/`HISTORY` (Großschreibung + Übersetzung) sind Teil derselben PO-Vorgabe,
werden aber vom Meta-Layout-Task (PF-1/PF-3+PF-4, dieselbe Funktion
`beanSections`/`view_detail_bean.go`) MITERLEDIGT statt doppelt angefasst — der
String-Sweep-Task deckt den REST ab (siehe `epic-E7-plan.md` Task-Aufteilung für die
Begründung dieses Schnitts).

**PF-8 — Command-Center-Schema `verb entity`, ohne Doppelpunkt** (PO-Nachtrag 2,
verbindliche PO-Beispiele: `set tags`, `set status`, `go to backlog`, `reload data`,
`go to settings`, `set title`). Vollständige Remap-Tabelle für
`overlay_palette.go`s Action-Labels:

| Alt (Deutsch) | Neu (`verb entity`) |
|---|---|
| `status: setzen` | `set status` |
| `tags: zuweisen` | `set tags` |
| `parent: zuweisen` | `set parent` |
| `blocking: zuweisen` | `set blocking` |
| `titel: bearbeiten` | `set title` |
| `bean: löschen` | `delete bean` |
| `create: bean` | `create bean` |
| `go to: backlog` | `go to backlog` |
| `go to: browse` | `go to browse` |
| `go to: review cockpit` | `go to review cockpit` |
| `filter: facetten` | `filter facets` |
| `search: beans` | `search beans` |
| `reload: daten` | `reload data` |
| `repo: wechseln` | `go to repo picker` |
| `settings: öffnen` | `go to settings` |

Zwei Planner-Entscheidungen (PO ließ bewusst offen, „konsistentes verb entity
wählen"): `filter: facetten` → **`filter facets`** (nicht „filter beans" — trifft die
tatsächliche Aktion präziser, die Facetten filtert, nicht die Bean-Liste als Ganzes);
`repo: wechseln` → **`go to repo picker`** (konsistent mit den drei bestehenden
`go to <view>`-Einträgen statt eines Einzelfalls `switch repo`). Fuzzy-Matching
(`palFilteredActions`) bleibt funktional unverändert (reiner Label-Text-Tausch, keine
Matching-Logik-Änderung) — Regressionstest: `TestPalFilteredActionsFuzzyFiltered`
muss mit den NEUEN Labels weiterhin Wortanfänge treffen (z.B. „stat" → `set status`).

**PF-10 — Redundante Pane-Titel entfernen** (PO-Nachtrag 4, PO wörtlich: „Es genügt,
wenn es in den Breadcrumbs `> repo-b: backlog` angezeigt wird. Dann die Suche - sonst
ist es obsolet."). Der Breadcrumb (Chrome-Zeile 1) trägt die View-Identität bereits —
der Pane-interne Titel + Unterstreichungszeile (`renderPane`, `render_shared.go`:
`title`-Feld des `pane`-Structs, aktuell `"Tree"`/`"Backlog"`/`"Review-Queue"`/
`"Detail"`, vier Call-Sites in `view_browse_repo.go`/`view_browse_backlog.go`/
`view_review_cockpit.go`) ist Dopplung und entfällt — **einheitlich für alle vier**
(Planner-Entscheidung zum Konsistenz-Prüfauftrag: `Detail`s Titel ist ohnehin durch
den PF-3/PF-4-Kopfblock ersetzt, `Review-Queue` dupliziert exakt wie `Tree`/`Backlog`
den Breadcrumb — dieselbe Regel „Breadcrumb = View-Identität, Pane-Titel nur bei
echter Zusatzinfo" trifft auf alle vier gleichermaßen zu, kein Grund für eine
Ausnahme). `renderPane` verliert die Titel+Trennlinien-Zeilen komplett (`pane.title`
wird entweder entfernt oder zum No-Op) — die erste sichtbare Zeile ist dann direkt
der Suchkopf (Tree/Backlog) bzw. die Summary-Zeile (Review-Cockpit) bzw. der PF-3/
PF-4-Kopfblock (Detail). **Geometrie-Folge:** `clickPaneGeometry` (`mouse.go`)
kodiert die alten 2 Zeilen explizit in `originY` (Doc-Kommentar: „the pane's OWN top
border(1) + its title line(1) + its separator line(1)") — die Formel verliert genau
diese beiden `+1`-Terme; `treeClickRow`/`backlogClickRow`/`reviewClickRow` und ihre
Tests (`mouse_test.go`) müssen entsprechend neu vermessen werden (Detail hat keinen
Click-Row-Konsumenten, dort ist es reine Render-Vereinfachung ohne Geometriefolge).

**PF-11 — Keybinding-Split Header/Footer, keine Dopplung** (PO-Nachtrag 5, erweitert
PO-Nachtrag 9). Header Zone 1 (`breadcrumb`, rechter Teil) zeigt künftig ALLE
globalen Bindings (Nachtrag 9: „ALLE globalen Bindings erscheinen oben rechts"):
`ctrl+r:reload · ctrl+k:commands · p:repos · ?:help · esc:back · enter:open/confirm ·
q:quit` (7 statt bisher 3 — `esc`/`enter`/`ctrl+k`/`p` fehlten heute komplett im
Header). Label-Kürzungen (Planner-Entscheidung, PO: „Wortwahl deine Entscheidung,
kurz und konsistent"): `keys.Refresh` Help-Text `"Reload data"`→`"reload"`,
`keys.Palette` `"Command-Center"`→`"commands"`, `keys.Picker`
`"Repo-Picker"`→`"repos"` (einheitlich kurz wie die bestehenden `"help"`/`"quit"`/
`"back"`) — da `helpGroups()`/Help-Overlay dieselben `keybind.Binding`-Objekte
wiederverwendet (Single Source), zieht die Kürzung dort automatisch mit, keine
zweite Textquelle. Footer Zone 3 zeigt AUSSCHLIESSLICH die view-lokalen Bindings der
aktiven View, UND wird zusätzlich kontextsensitiv für offene Overlays/Forms/Filter
(Q04, s. u.). Betrifft nur noch ZWEI `*Chrome()`-Baufunktionen
(`browseRepoChrome`/`backlogChrome` — `reviewCockpitChrome` entfällt komplett durch
PF-14). Beide verlieren `keys.Refresh`/`keys.Enter` (`browseRepoChrome` dupliziert
heute `Refresh`+`Enter`, `backlogChrome` dupliziert `Enter`) aus ihrer jeweiligen
`localHint`-Liste. Single Source: neue `globalBindings() []keybind.Binding`
(`keymap.go`, Rückgabe `{Refresh, Palette, Picker, Help, Back, Enter, Quit}`, exakte
Reihenfolge wie der Header-Text oben) ersetzt die heute inline duplizierte
`renderBindings([]keybind.Binding{keys.Refresh, keys.Help, keys.Quit})`-Zeile in
beiden Chrome-Baufunktionen. Neuer Drift-Guard (`keymap_test.go`, Reflection wie
`TestHelpGroupsCoverEveryBindingExactlyOnce`): kein Binding erscheint gleichzeitig in
`globalBindings()` UND einer der (jetzt zwei) view-lokalen Listen — dafür werden die
`localHint`-Bindinglisten aus den Chrome-Baufunktionen in eigene benannte
`[]keybind.Binding`-Funktionen extrahiert (testbar per Reflection/Keys()-Identität,
nicht per fragilem String-Vergleich am gerenderten Footer-Text). Nebeneffekt (PO
notiert es selbst): entschärft das VQA-I01-Footer-Umbruch-Finding bei 110 Spalten, da
der Footer kürzer wird — kein separater Fix nötig. **Kontextsensitiver Footer
(Q04-Antwort, PO-Nachtrag 5):** sobald ein Overlay/Form/Filter-Menü aktiv ist, zeigt
der Footer dessen lokale Bindings statt der (dann irrelevanten) View-Bindings —
insbesondere `keys.Toggle` (Help-Text „Toggle facet", deckt PO Q04s „space:
select/toggle" ab) beim offenen Filter-Menü, dem konkreten Fall, an dem PO das
Fehlen bemerkte. Priorität bei mehreren gleichzeitig denkbaren Zuständen:
Filter-Menü > Overlay > Suche > Palette > Help > View-Normalzustand (deckt sich mit
`handleKey`s bestehender Full-Capture-Reihenfolge, `update.go`).

**PF-14 — Review-Cockpit vollständig entfernt (Feature-Removal)** (PO-Nachtrag 8, PO
wörtlich: „widerspricht dem lean-stack-Wesen und schafft wieder Zeremonie. Das
Review-Cockpit ist Zauberei on-top und bitte raus nehmen. Das Review möchte ich in
Zukunft direkt im Chat machen."). Ersetzt V6 (§6) ersatzlos — kein Nachfolge-View.
Die Tag-Konvention bleibt (§5, oben, umgeschrieben), nur die TUI-Interaktion dafür
entfällt. **Removal-Scope** (vollständig, PO-Nachtrag 8 verbatim strukturiert):
`viewReviewCockpit` (`view_review_cockpit.go` komplett) inkl.
`reviewState`/`reviewQueue`/`reviewGroup`/`reviewRework`-Derivation,
`reviewCursor`/`reviewAccOpen`-Modellfelder (`types.go`), `clampReviewCursor`;
Keybinding `R` + Cockpit-lokale Keys (`a`/`x`/`o`, `n`/`p`-Override) aus
`keymap.go`/`helpGroups()` (Drift-Guard-Test zieht automatisch mit);
`reviewStandMarkdown` + der Cockpit-`y`-Override (Yank-Review-Stand entfällt);
`reviewClickRow` + zugehöriger Lock-Test; Palette-Eintrag „go to review cockpit";
Reject-Form (`form_reject_review.go`); 2 Cockpit-Goldens. **Datenlayer-Entscheidung
(YAGNI, Planner):** `PassReview`/`RejectReview`-Mutationsfunktionen im Datenlayer
(`internal/data/`) BLEIBEN — harmlos, CLI-nah, keine TUI-Kopplung, kein Grund für
Removal-Risiko an einer Stelle, die niemand mehr aufruft; nur die TUI-Verdrahtung
(View, Keybindings, Palette-Eintrag) verschwindet. E6-Auswirkung: US-08-Validierung
läuft gegen die NEUE Definition (§10, oben). **Reihenfolge:** dieser Removal-Task
läuft ALS ERSTES im E7-Plan (`epic-E7-plan.md`) — alle Glyphen-/Sprach-/Footer-/
Klick-Geometrie-Umbauten müssen das Cockpit dann nicht mehr mitziehen (spart Arbeit
in jedem Folge-Task: ein View, zwei Goldens, mehrere Keybindings und String-Literale
weniger zu berücksichtigen).

**PF-12 — Kein Layout-Shift bei Selektion im Detail-Pane** (PO-Nachtrag 6). Der
Platz für den Select-/Fokus-Marker ist eine IMMER reservierte Gutter-Spalte (Leerraum
gleicher Breite, wenn nicht selektiert) — gilt für Accordion-Sektionsköpfe UND die
PF-4-Meta-Feldliste. Betrifft ZWEI bestehende Stellen in `accordion.go`, die diese
Regel heute verletzen (bedingtes Voranstellen statt reservierter Spalte): (1)
`renderAccordion`s aktive-Sektion-Styling hängt `▌` NUR bei `activeSec` an (inaktive
Header haben KEIN Prefix-Zeichen, `w`-Truncation ist zwischen den beiden Zweigen
unterschiedlich: `w` inaktiv vs. `w-1` aktiv) — Fix: BEIDE Zweige reservieren
IMMER 1 Spalte (`" "` inaktiv, `"▌"` aktiv), truncate konsistent auf `w-1`. (2)
`fieldStrip`s Feld-Marker (Beziehungen-Sektion) hat dieselbe Asymmetrie
(`"▌"+label` aktiv vs. nacktes `label` inaktiv) — PF-12s allgemeine Formulierung
„alle markierbaren Zeilen" deckt strukturell dieselbe Wurzelursache, im selben
Aufwasch mitgefixt. Die NEUE Meta-Feldliste (PF-4) ist von Anfang an mit fixem
`▷ `/`▶ `-Prefix zu bauen (keine bedingte Auslassung) — erfüllt PF-12 automatisch,
wenn korrekt implementiert. Testpflicht: Assertion/Golden, dass die
`lipgloss.Width(ansi.Strip(zeile))` einer NICHT-aktiven Zeile identisch ist,
unabhängig davon, welche ANDERE Zeile gerade aktiv ist (zwei Renders mit
unterschiedlichem `activeIdx`/`fieldIdx`, dieselbe inaktive Zeile verglichen).

**PF-13 — Fokus-Wechsel-Symmetrie `tab`/`shift+tab`** (PO-Nachtrag 7, PO wörtlich:
„für Nutzer murks"; Leitlinie: „vorhersagbare Paare, kein Überraschungsverhalten").
**Devd-Referenz geprüft** (Kollisionscheck-Auftrag): devd kennt KEIN `shift+tab` —
dort ist `tab` lediglich ein dritter Alias auf die `Right`-Bindung
(`WithKeys("right","l","tab")`, devd `keymap.go:76`), keine eigene
Fokus-Wechsel-Bindung. beans-tuis `tab` ist strukturell etwas ANDERES (ein globaler
Toggle `m.detailFocus = !m.detailFocus`, `update.go` Zeile ~883) — devds Muster lässt
sich nicht 1:1 übertragen, PO-Wunsch geht bewusst über das devd-Vorbild hinaus.
**Ist-Zustand** (verifiziert, nicht vermutet): `tab` toggled bereits HEUTE in beide
Richtungen (Tree→Detail UND Detail→Tree, dieselbe Taste) — der fehlende Teil ist
ausschließlich eine EIGENE, deterministische `shift+tab`-Rückwärts-Taste; Pfeile
funktionieren bereits korrekt getrennt (im Tree: `→`=Node expandieren, `←`=Node
einklappen — NICHT Fokus-Wechsel; in Detail-Fokus: `→`=Sektion→Feld,
`←`=Feld→Sektion bzw. bei Sektions-Ebene ganz zurück zu Tree) — **keine
Verhaltensänderung an den Pfeiltasten nötig**, nur Verifikation+Dokumentation
(dieser Absatz IST der geforderte Kollisionsanalyse-Nachweis). **Entscheidung**
(Planner, geringstes Regressions-Risiko): `tab` BEHÄLT sein bestehendes
Toggle-Verhalten (kein Bruch für bestehende Nutzer/Tests, erfüllt „tab = Fokus
vorwärts" weiterhin) — NEU: `shift+tab` wird als eigene, deterministische
`m.detailFocus = false`-Taste ergänzt (No-Op wenn bereits im Tree). Damit
funktionieren BEIDE PO-geforderten Richtungen ohne eine bestehende Fähigkeit zu
entfernen. Formalisiert als ZWEI neue `keyMap`-Felder (`keymap.go`): `FocusIn`
(`tab`, Help „focus in/toggle") und `FocusOut` (`shift+tab`, Help „focus out") —
ersetzt den heutigen rohen `msg.String()=="tab"`-Vergleich (`update.go`) durch
`keybind.Matches`, UND den heute hand-getippten `"  tab:focus"`-Footer-Suffix
(`browseRepoChrome`/`backlogChrome`) durch denselben `renderBindings`-Mechanismus
wie jede andere Bindung (DD2-175-Konformität, schließt an PF-11 an). Beide neue
Felder gehören in die „Navigation"-`helpGroup` (Drift-Guard-Pflicht). **Sequenz-
Folge:** der Navigations-Task (führt `FocusIn`/`FocusOut` ein) muss vor dem
Header/Footer-Task (PF-11, baut den kontextsensitiven Footer, der diese Felder mit
anzeigt) laufen — `blocked_by` in `epic-E7-plan.md`s Task-Übersicht entsprechend
gesetzt.

### Offene, NICHT umgesetzte Punkte

- **Q03 (PO-Nachtrag 3):** zentrale Tag-Definition über eigene Page — bereits als
  eigenständiges Feature-bean `bt-6oyy` außerhalb von E7 angelegt, Scope-Entscheid v1
  vs. v1.1 liegt beim PO. NICHT Teil dieses Eposs.

Q04 (PO-Nachtrag 5) war zunächst offen, ist aber vollständig gelöst (s.
PO-Antworten-Tabelle unten, „Q04-Antwort") — hier absichtlich nicht mehr als offener
Punkt geführt.

### PO-Antworten/Nachträge (2026-07-15) — Referenz

| Frage/Punkt | Antwort | Übernommen in |
|---|---|---|
| Q01 (Header rechts) | Identisch mit PF-4-Kopf, kein App-Chrome-Umbau | PF-3+PF-4 |
| Q02 (Farben) | Vollständig bestätigt (Tabellen oben) | PF-6 |
| D01 | ~~`enter` = Detail-View betreten~~ **ZURÜCKGENOMMEN** (Nachtrag 3): `tab` bleibt einziger Einstieg | PF-5 (revidiert) |
| D02 | Leitprinzip: Auswahl schnell/einfach (wenig Tasten) | PF-5 (Kaskaden-Constraint) |
| D03 | E6 nach E7 (`blocked_by` bereits gesetzt: `bt-wm4w`, `bt-9yvh` ← `bt-heg9`) | Task-Übersicht `epic-E7-plan.md` |
| Nachtrag 2 | PF-7 (Sprache Englisch), PF-8 (Command-Center-Schema) | PF-7, PF-8 |
| Nachtrag 3 | D01 revidiert; Filter-Logik PO-validiert („exzellent", → E6/US-05-Evidenz); Tags bleiben, Tag-Page als `bt-6oyy` ausgelagert (Q03) | PF-5 (revidiert), E6-Hinweis in `epic-E7-plan.md` |
| Nachtrag 4 | PF-10 (Pane-Titel-Dopplung) | PF-10 |
| Nachtrag 5 | PF-11 (Header/Footer-Keybinding-Split); Q04 zunächst offen | PF-11 |
| Q04-Antwort | Bestehende space/x-Toggle in Forms/Overlays (kein Multi-Select) → Footer wird kontextsensitiv (view-lokal vs. overlay-/form-lokal) | PF-11 |
| Nachtrag 6 | PF-12 (kein Layout-Shift bei Selektion) | PF-12 |
| Nachtrag 7 | PF-13 (`tab`/`shift+tab`-Symmetrie) | PF-13 |
| Nachtrag 8 | PF-14 (Review-Cockpit vollständig entfernt) | PF-14, §5, §6, §10 |
| PF-14-Präzisierung | Tag-Trio `to-review`/`accepted`/`rejected` ersetzt `to-review`/`rework` | §5 |
| Nachtrag 9 | Header-Globals vollständig (`p`, `ctrl+k` ergänzt) | PF-11 |

Stand 2026-07-15: PF-1…PF-14 vollständig entschieden, keine offenen Q-Marker mehr
(Q03 bewusst außerhalb E7, Q04 gelöst).
Realisierungsplan: `docs/plans/v1-port/epic-E7-plan.md`.
