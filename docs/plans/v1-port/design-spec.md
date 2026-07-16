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
| Tags | beans-Tags (implizit, ohne Farbregister) | Tag-Picker mit Nutzungszählern; Farbe per Hash aus fester Palette; ~~Tag-Manager-CRUD entfällt (Tags entstehen/sterben implizit)~~ **Superseded (E10, §16, 2026-07-16):** Tag-Manager-CRUD existiert jetzt als eigene Page (v1.1, Registry `.beans-tags.yml`); Tags entstehen/sterben weiterhin implizit, die Registry ist eine zusätzliche optionale Definitionsschicht |
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

**OUT (v1, bewusst):** ~~Tag-Manager-CRUD (Tags implizit)~~ **(Superseded E10, §16,
2026-07-16 — Tag-Management-Page in v1.1 nachgeliefert)** · Memory/Docs/Notes/ToDos-Views
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
Filter-Menü > Overlay > Suche > Palette > Help > View-Normalzustand — eine EIGENE,
für die Footer-Anzeige gewählte Reihenfolge, KEIN Abbild von `handleKey`s
Full-Capture-Dispatch-Reihenfolge (`update.go`), die `m.searchActive` VOR
`m.filterOpen` prüft (umgekehrte Reihenfolge). Folgenlos, da die Capture-Zustände
sich gegenseitig ausschließen (nie gleichzeitig aktiv) — die beiden Reihenfolgen
dürfen aber nicht verwechselt werden (T7-Review I03, bean bt-dsog).

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

**PF-15 — Tags als 7. Meta-Feld** (PO-Feedback-Runde 2, 2026-07-15, bean
`bt-ntoz`, D01 ENTSCHIEDEN). Löst US-08s einzige verbliebene Lücke
(validation.md §2 B01/bean `bt-gdkx`: Tags weder im Tree noch im
Detail-META sichtbar, nur die Filter-Facette funktionierte). Die
PF-4-Meta-Feldliste (§15 oben) wächst von 6 auf 7 Zeilen — `tags:` wird
DIREKT NACH `priority` und VOR `created_at` eingefügt (PO-Entscheid
verbatim: KEIN Tree-Suffix — rasches Filtern nach z.B. offenen Reviews via
`f`-Filter/Tags-Facette deckt den Überblick bereits ab):

```
[1] META
▷ title:      xxxx
▷ status:     xxxx
▷ type:       xxxx
▷ priority:   xxxx
▷ tags:       xxxx
▷ created_at: xxxx
▷ updated_at: xxxx
```

Wert-Rendering: die bislang tote `tagsInline`/`tagSwatch`-Helferfunktion
(`render_shared.go`, seit dem PF-1/PF-4-Meta-Redesign ohne Aufrufer —
validation.md B01 wertet das als Indiz für ein ERSATZLOSES statt bewusstes
Wegfallen) wird hier wiederbelebt: `tagsInline(b.Tags)` liefert die
Hash-gefärbten `● tag`-Swatches; ein taglos Bean zeigt einen
`theme.Dim`-Platzhalter `(none)` statt Leerzeile. `kind: "tags"` ist ein
NEUER `relationField.kind`-Wert — die Enter-Kaskade (PF-5, §15 oben)
behandelt ihn analog `status`/`type`/`priority`: `enter` auf der
`tags:`-Zeile öffnet den bestehenden Tag-Picker (`m.openTagPicker()`,
funktional identisch zur `t`-Taste) statt des Value-Menüs; `m.detailFocus`
bleibt dabei — wie bei den anderen drei constrained-Feldern — `true`
(Overlay legt sich als eigener Capture-State darüber, D02-Leitprinzip
„schnell/einfach", PF-5 oben). Schließt bean `bt-gdkx` (US-08-Redefinition,
design-spec §10) inhaltlich — der Implementer-Task referenziert `bt-gdkx`,
schließt es aber NICHT selbst (PO-Gate, Review-Flow §5).

**PF-16 — PO-Feedback-Runde 2 Sammelposten** (Grilling 2026-07-15, bean
`bt-ntoz`, B01-B14 + Revisionen D02-D06, D08 — vollständige Herleitung/
Zitate NUR in `bt-ntoz`, hier bewusst kompakt; Umsetzung
`epic-E8-plan.md`):

| Code | Fix (kompakt) | Betroffen |
|---|---|---|
| B01 | Pfeil-links verlässt Detail-Fokus nicht mehr (asymmetrisch zu Pfeil-rechts, das nie hineinführt) — Fokus-Wechsel exklusiv `tab`/`shift+tab`. **Revidiert PF-13** (unten). | `update.go` (`keyDetailFocus`) |
| B02 | Kopfblock `type:…status:…prio:…` springt bei Bean-Wechsel — feste Spaltenbreiten (`type`→9, `status`→11, Wortlänge `milestone`/`in-progress`). | `view_detail_bean.go` (`detailHeaderBlock`) |
| B03 | Kinderlose Beans zeigen ein Expand-Dreieck. **Verifiziert bereits korrekt** (`treeNodeMarker`, `view_browse_repo.go:401-409`, blanks bereits für `!hasKids`) — kein Code-Fix, nur Regressionstest ergänzt. | `view_browse_repo.go` (Test only) |
| B04 | `title:` erscheint sofort ▶-selektiert nach `tab`, bevor die Feld-Ebene betreten wurde — ▶-Marker erst ab `m.detailLevel==1`. | `view_detail_bean.go`/`accordion.go`/`view_browse_repo.go` (Signatur-Erweiterung `detailLevel` durchgereicht) |
| B05 | Redundantes `▾`/`▸`-Chevron im Accordion-Header entfernt (Zustand am Inhalt sichtbar). | `accordion.go` (`renderAccordion`) |
| B06 | EXPERIMENT: inaktive Accordion-Header-Farbe Grau→Teal (Verwechslungsgefahr mit Meta-Label-Spalte). PO-Sign-off per Vorher/Nachher-Vergleich VOR Abnahme. | `accordion.go`, `theme.go` (neuer Token) |
| B07 | Maus im Detail-Pane: Sektions-Header UND Meta-Feldzeilen bislang nicht klickbar. Klick auf Header = aktivieren/expandieren; Klick auf Feld = selektieren; Doppelklick = Edit-Overlay (analog Enter-Kaskade PF-5). | `mouse.go` (neu `detailClickRow`/`mouseDetailClick`), `view_browse_repo.go`/`view_browse_backlog.go` (Dispatch) |
| B08 | Quit-Text `„Really quit bt."`→`„Really quit bt?"`. Quit-Kaskade zweistufig: `q`→`enter` aus Browse/Backlog (mit konfigurierten Repos) führt zur Lobby statt zum Exit; aus der Lobby (oder ohne konfigurierte Repos, Randfall) beendet `q`→`enter` die TUI. | `box_confirm_quit.go` |
| B09 | Inaktive `▷`-Feldmarker der Meta-Feldliste waren unstilisiert (weiß) — auf `theme.Muted` (subtext/grau) gestellt, nur `▶` trägt Mauve. | `view_detail_bean.go` (`metaSectionBody`) |
| B10 | `e`/`enter` auf Sektion `[2] BODY` war inkonsistent (`e` öffnete fälschlich Titel-Edit, `enter` war No-Op) — beide öffnen jetzt `$EDITOR` auf dem Body, kontextsensitiv zur gewählten Sektion. | `update.go` (`keyNodeAction`, `keyDetailFocus`), neuer Helfer `openBodyEditor` |
| B11, B12 | Kombiniertes Value-Menü (`s`, `box_menu_value.go`) zeigte IMMER alle 15 Zeilen (Status+Type+Priority) statt nur der über die Enter-Kaskade/`s`-Taste gewählten Gruppe — `buildValueMenuItems(group)` liefert künftig nur noch die eine angefragte Gruppe; Palette gewinnt `set type`/`set priority` (bislang nur `set status`). | `box_menu_value.go`, `overlay_palette.go` |
| B13 | Command-Center mischte Bean-Treffer unter die Commands — entfernt (Bean-Suche gehört exklusiv zu `/`). **Revidiert US-04** (§10) und die E4-Design-Entscheidung „Aktionen + Bean-Treffer gemischt". | `overlay_palette.go`, `types.go`, `messages.go` (kompletter Bleve-Palette-Unterbau entfernt, Compiler-gesteuert wie PF-14/T1) |
| B14 | Tag-Neuanlage war unentdeckbar (existierte bereits als `n`-Taste im Tag-Picker, aber nirgends im Footer sichtbar, kein Palette-Command). Neue `keys.NewTag`-Bindung (`n`) im Tag-Picker-Footer sichtbar; Palette-Command `create tag`. Tag-Management-**Page** (`bt-6oyy`) bleibt v1.1 (D08). | `keymap.go`, `box_picker_tag.go`, `footer_context.go`, `overlay_palette.go` |
| D02 | Backlog-Sort-Modus (E2-Erbe, kein sichtbarer Indikator) — dezenter Subtext-Suffix in der Backlog-Suchzeile, z. B. `⌕ / search · sort prio` (Tree unverändert, kein Suffix). | `view_browse_repo.go` (`treeSearchLine`, neuer Parameter), `view_browse_backlog.go` |
| D03 | `esc` war in der Detail-Kaskade ein No-Op (E2-Erbe) — jetzt universelles „eine Ebene zurück": Feld-Ebene→Sektions-Ebene→Fokus verlassen. Prüfauftrag „alle esc-Sites einheitlich" (Suche/Filter/Picker/Lobby/Quit) als Audit-Tabelle im Umsetzungs-Task belegt, nicht blind neu gebaut — alle fünf bereits konform. | `update.go` (`keyDetailFocus`) |
| D04 | Header-Globals auf genau 4 gekürzt: `ctrl+k` · `p:repos` · `?:help` · `q:quit` — `ctrl+r`/`esc`/`enter` fliegen aus dem Header (bleiben im Help-Overlay dokumentiert). | `keymap.go` (`globalBindings()`) |
| D05 | Overlay-Footer zeigen weiterhin `enter`/`esc` — durch D04 keine Dopplung mehr mit dem Header. Sign-off, KEIN Code-Fix. | — |
| D06 + Q06 | Footer-Neuspezifikation (ersetzt den T7/PF-11-Stand): Navigations-Keys raus, Reihenfolge `tab focus in · shift+tab focus out · / search · f Filter · s Status · c Create · d Delete · e Edit · b Backlog · t Tags · y Yank · a Parent · r Blocking`, Taste TEAL/Aktionswort subtext-grau, KEIN `:` mehr (Farbtrennung ersetzt den Doppelpunkt, gilt einheitlich auch für die 4 Header-Globals aus D04). `Blocking` wird von `B` auf `r` umbelegt (`r` seit PF-14 frei, `B` wird frei). Backlogs bestehender `Sort`(`S`)-Footer-Eintrag bleibt zusätzlich erhalten (Backlog-exklusiv, von Q06s Liste nicht berührt — Planner-Entscheidung, kein Entzug ohne PO-Anweisung). | `keymap.go`, `view_browse_repo.go` (`browseRepoLocalBindings`), `view_browse_backlog.go` (`backlogLocalBindings`) |
| D08 | Tag-Management-Page (`bt-6oyy`) → v1.1, B14 ist die v1-Minimal-Lösung. Bereits als Body-Nachtrag auf `bt-6oyy` dokumentiert (Grilling-Abschluss 2026-07-15) — kein weiterer Schritt nötig. | — |

**PF-13-Pfeil-Revision (B01):** PF-13 (oben) beschrieb `left` bei
`detailLevel==0` als gültigen, dokumentierten Rückweg zum Tree
(„Rückweg bis zum Tree"). B01 nimmt das zurück: Pfeiltasten sind ab sofort
REIN Navigation (Sektion/Feld wechseln), NIE Fokus-Wechsel — einzig
`tab`/`shift+tab` (weiterhin PF-13) ändern `m.detailFocus`. Grund (PO
verbatim): Pfeil-rechts führte nie in den Detail-Fokus hinein, Pfeil-links
aber heraus — asymmetrisch, „für Nutzer murks" (dieselbe Formulierung, die
PF-13 selbst ursprünglich motivierte). `left` bei `detailLevel==0` wird zum
No-Op; der Rückweg zum Tree läuft danach über `shift+tab` (deterministisch,
PF-13) oder die neue `esc`-Kaskade (D03, oben).

**US-04-Revision (B13, design-spec §10):** US-04s Akzeptanzkriterium
„Aktionen + Bean-Treffer gemischt, kontextabhängige Einträge zuerst" wird
per B13 bewusst zurückgenommen — das Command-Center zeigt AUSSCHLIESSLICH
Commands, Bean-Treffer entfallen ersatzlos (Bean-Suche ist exklusiv `/`s
Aufgabe, design-spec §6 V2/V3). Dies revidiert die ursprüngliche E4-Design-
Entscheidung (§6 V5, „Bean-Suche (Bleve) in einem" mit den Aktionen) — neuer
US-04-Wortlaut: „ctrl+k, fuzzy NUR über Commands, enter dispatcht,
kontextabhängige Einträge (fokussiertes Bean) zuerst." E6/US-04-Validierung
muss künftig gegen diesen neuen Wortlaut laufen, nicht den alten.

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
Realisierungsplan (PF-1…PF-14): `docs/plans/v1-port/epic-E7-plan.md`.

Stand 2026-07-15 (Grilling R2): PF-15/PF-16 (D01-D06, D08, B01-B14)
vollständig entschieden, PO-Freigabe erteilt (bean `bt-ntoz`). Realisierungsplan:
`docs/plans/v1-port/epic-E8-plan.md`.

## PF-17 — PO-Feedback-Runde 3 (2026-07-16, Live-Test v1.0/E8, bean `bt-tct9`)

Sammelposten analog PF-16: D01/D02/B03/B04/B05/B06/F01, vollständige Zitate/
Herleitung nur in `bt-tct9`, hier die verbindlichen Design-Entscheidungen.
Realisierungsplan: `docs/plans/v1-port/epic-E9-plan.md`.

**D01 — Edit-Modell: `e` wird der Ganz-Bean-`$EDITOR`, `enter` bleibt reine
Feld-Kaskade (REVIDIERT/supersedet E8-B10, `PF-16`-Tabelle B10-Zeile).**
PO verbatim: „'enter' öffnet im details-view die forms für das edit eines
Feldes, 'e' öffnet egal an welcher Stelle das gesamte bean in \$EDITOR."
Zwei getrennte Zuständigkeiten, ab jetzt ohne Überschneidung:

- `enter` ist ausschließlich die PF-5-Feld-Kaskade (Section→Feld→Overlay/
  Form/Jump) — unverändert. Auf `[2] BODY` war das seit E8-B10 ein
  Sonderfall (`enter` öffnete dort `$EDITOR` auf dem Body) — **B10-Revision:**
  dieser Sonderfall entfällt ersatzlos, `enter` auf `[2] BODY` wird wieder
  der generische No-Op (die Sektion trägt keine `.fields`, exakt der
  Vor-E8-Zustand) — PO's neues Mentalmodell reserviert „$EDITOR öffnen"
  ausschließlich für `e`, Body-Inhalt ist über `e` (als Teil der Gesamtdatei)
  ohnehin erreichbar. `openBodyEditor` (`editor.go`) wird durch
  `openBeanEditor` ersetzt (nicht daneben behalten — EIN Editor-Helfer).
- `e`/`ctrl+e` (eine `keys.Editor`-Bindung, unverändert im Keymap) öffnen ab
  sofort UNBEDINGT denselben `openBeanEditor(b)`-Pfad — die bisherige
  Verzweigung in `keyNodeAction` (`update.go`: `ctrl+e` immer Body,
  `e` kontextsensitiv Titel-Form vs. Body) entfällt vollständig zugunsten
  eines einzigen unconditional calls. „ctrl+e-Sonderpfad vereinheitlichen"
  (PO/D01-Wortlaut) heißt konkret: es gibt nach dieser Revision NUR NOCH
  einen Pfad, nicht mehr zwei konvergente. Erreichbar von JEDER Stelle
  (Tree/Backlog/Detail, jede Sektion/Feld-Ebene, auch ohne aktiven
  Detail-Fokus) — `keyNodeAction` wird bereits heute VOR dem
  `detailFocus`/`keyBacklog`/`keyTree`-Dispatch geprüft (`handleKey`,
  `update.go`) und wirkt auf `m.focusedBean()` (view-agnostisch), das
  erfüllt „egal an welcher Stelle" strukturell bereits — keine
  Dispatch-Order-Änderung nötig, nur die Handler-Logik selbst vereinfacht
  sich. `e`/`ctrl+e` öffnen NIE MEHR das Titel-Edit-Form (`openEditTitleForm`
  bleibt nur noch über `enter` auf dem `title:`-Feld erreichbar,
  `activateDetailField`s `"title"`-Case unverändert).

**Technisches Design „Ganz-Bean-Editor" (kein bestehender CLI-Weg, neu
zu bauen):** `beans show <id> --raw` liefert bereits EXAKT das reale
Datei-Format (verifiziert, `beans show bt-tct9 --raw` == on-disk
`.beans/bt-tct9--*.md` Byte für Byte): YAML-Frontmatter zwischen zwei
`---`-Zeilen (mit `# <id>`-Kommentarzeile, `title/status/type/priority/
tags?/created_at/updated_at/parent?/blocking?/blocked_by?`) gefolgt vom
Markdown-Body. Seed-Text für `$EDITOR` ist also `beans show <id> --raw`
UNVERÄNDERT (kein selbstgebautes Markdown-Templating nötig — „beans bleibt
die eine Autorität", design-spec.md §3.1 D02, gilt auch hier: die
CANONICAL-Serialisierung kommt aus dem CLI, nicht aus einer eigenen
Re-Implementierung des Dateiformats).

- Neu `data.Client.ShowRaw(id string) (string, error)` (`internal/data/
  client.go`): `c.run("show", id, "--raw")`, gibt den rohen String zurück
  (kein JSON-Envelope bei `--raw`, kein `classifyError`-Pfad nötig — ein
  reiner Read).
- `openBeanEditor(b *data.Bean) (model, tea.Cmd)` (`editor.go`, ersetzt
  `openBodyEditor`) ruft `ShowRaw` synchron NICHT direkt (Subprocess-Latenz
  ~20-50ms, design-spec §3.1) — analog zu jedem anderen Mutation-Cmd läuft
  der Read als eigener `tea.Cmd` (neuer `beanRawLoadedMsg`), dessen Ergebnis
  ERST den `tea.ExecProcess`-Suspend triggert (zwei Cmd-Hops statt einem,
  analog zum bestehenden Zwei-Schritt-Muster Read→Suspend, kein
  synchroner Subprocess-Call im Update-Pfad). `m.editorTarget`/
  `m.editorETag` werden wie bisher beim ÖFFNEN eingefroren (F2-Konvention,
  `applyEditorFinished`s Lost-Update-Rationale unverändert gültig — ein
  potenziell langlebiger `$EDITOR`-Suspend braucht das gefrorene ETag, nicht
  ein frisches bei Submit).
- Rückweg (Submit): `editorFinishedMsg{content,changed,err}` bleibt der
  TYP (unverändert) — `applyEditorFinished` (`update.go`) wird umgebaut: bei
  `changed==true` wird `content` NICHT mehr direkt als `SetBody`-Argument
  verwendet, sondern zuerst geparst:
  1. Split am ZWEITEN `---`-Trenner (erste Zeile ist der erste `---`) in
     Frontmatter-Block + Body (Body = Rest NACH dem zweiten `---`, exakt
     EIN führender Newline abgeschnitten, spiegelt `--raw`s eigenes Format).
  2. `gopkg.in/yaml.v3` (bereits Projekt-Dependency, `internal/config/
     settings.go`) unmarshalt den Frontmatter-Block in einen neuen,
     schlanken Typ `rawBeanFrontmatter{Title, Status, Type, Priority string;
     Tags, Blocking, BlockedBy []string; Parent string}` — die `# <id>`-Zeile
     ist ein YAML-Kommentar, wird von yaml.v3 automatisch übersprungen.
     `created_at`/`updated_at`/die ID selbst werden BEWUSST NICHT geparst
     (kein `beans update`-Flag existiert dafür — s. „Bekannte Grenze" unten).
  3. Diff gegen den bei `$EDITOR`-Open eingefrorenen Bean-Snapshot
     (`m.editorSnapshot *data.Bean`, NEUES Modelfeld neben `editorTarget`/
     `editorETag` — der volle Bean-Wert zum Öffnungszeitpunkt, nicht nur
     ID+ETag, weil der Diff jedes Feld einzeln braucht).
  4. EIN kombinierter `beans update`-Call trägt jede geänderte Eigenschaft
     (mirrort SetTags/SetBlocking's Single-Etag-No-Cascade-Konvention,
     `mutations.go`) — neue `data.Client.UpdateWhole(id string, diff
     WholeEditDiff, etag string) error` (`internal/data/mutations.go`), mit
     `WholeEditDiff{Title, Status, Type, Priority *string; TagsAdd,
     TagsRemove, BlockingAdd, BlockingRemove, BlockedByAdd, BlockedByRemove
     []string; ParentChanged bool; Parent string; Body *string}` — baut die
     `update`-Argumentliste konditional (nur geänderte Felder tragen Flags),
     EIN `c.update(id, etag, args...)`-Aufruf, kein Kaskaden-Risiko.
     `ParentChanged`+`Parent==""` mapt auf `--remove-parent` (mirrort
     `RemoveParent`), sonst `--parent <val>`.
  5. Fehlerfall (Parse-Fehler ODER CLI-`VALIDATION_ERROR`, z.B. ein von Hand
     getippter ungültiger `status:`-Wert): das PO-Risiko „constrained Edits
     — no 422" (design-spec §1 Leitbild) gilt hier NICHT mehr uneingeschränkt
     — der Ganz-Bean-Editor ist bewusst UNconstrained (Freitext-YAML), das
     ist der ganze Punkt von D01/B01 (PO will MEHR Zugriff als die
     Overlays bieten). Rejection wird daher NICHT verhindert, sondern
     RECOVERABLE gemacht: mirrort `writeConflictTempFile`s bestehende
     Konvention (`update.go`) — der editierte Rohtext wird in eine
     GEHALTENE Tempdatei geschrieben, ihr Pfad reitet im Toast/Status-Zeilen-
     Text mit, nichts geht verloren. Ein echter ETag-Konflikt (paralleler
     externer Write) läuft weiterhin durch den bestehenden
     `conflictWithRecovery`-Pfad (unverändert).
- **Bekannte Grenze (zu dokumentieren, kein Bug):** `created_at`/
  `updated_at`/die `# <id>`-Kopfzeile sind im Editor-Text SICHTBAR (Teil von
  `--raw`), aber NICHT editierbar wirksam — `beans update` kennt dafür keine
  Flags, Änderungen daran werden beim Parse zwar erkannt, aber beim Bauen
  der `WholeEditDiff` stillschweigend verworfen (kein Flag existiert). Muss
  im PO-Report als ERRATUM/Deviation stehen, kein Implementierungsauftrag,
  ein CLI-seitiges Feature-Gap.

**D02 — Backlog-Footer verliert `Sort`, Taste bleibt (PO bestätigt, Option
b + Präzisierung).** `backlogLocalBindings()` (`view_browse_backlog.go`)
verliert das angehängte `keys.Sort` — reduziert auf exakt
`browseRepoLocalBindings()` (keine eigene Backlog-Ergänzung mehr). `S`
bleibt funktional (`keyBacklog`s `Sort`-Case unverändert), taucht aber NUR
noch in `helpGroups()` (`keymap.go`, dort bereits gelistet, keine Änderung
nötig) auf. Suchzeilen-Suffix `· sort <modus>` (E8/D02,
`treeSearchLine`) bleibt unverändert die sichtbare Laufzeit-Anzeige. Q06s
Kern-Footer-Liste (`browseRepoLocalBindings`) bleibt wortgesperrt,
unangetastet von diesem Punkt.

**B03 — Titel-Edit-Form wird multi-line.** `buildEditTitleForm`
(`form_edit_title.go`) tauscht `huh.NewInput()` gegen `huh.NewText().
Lines(3).ExternalEditor(false)` (huh v1.0.0, `field_text.go` verifiziert:
`NewText` ist eine Textarea mit Wordwrap out-of-the-box).
`.ExternalEditor(false)` ist PFLICHT: `huh.Text` bringt einen EIGENEN,
huh-internen `ctrl+e`-Editor-Suspend-Mechanismus mit (Default `true`) — der
würde mit D01s app-weitem `e`/`ctrl+e`-Ganz-Bean-Editor kollidieren
(zwei verschiedene Editor-Sessions auf zwei verschiedene Inhalte, gleiche
Taste, nur weil GENAU in diesem einen Formular fokussiert) — deaktiviert,
damit `keyForm`s eigene Tastatur-Semantik die einzige bleibt. `.Lines(3)`
ist eine Planner-Schätzung (PO nannte keine Zeilenzahl) — 3 Zeilen decken
die längsten bislang beobachteten Bean-Titel ab, ohne das Formular-Modal
unnötig zu strecken (`formInnerHeight`, `forms_shared.go`, deckelt ohnehin).
`nonEmpty`-Validator bleibt unverändert. **Ergänzung (Supervisor-Entscheid,
T3-Review Fix-Runde 1, bean `bt-2v38`, F02):** `huh.Text` bringt (anders als
das vorherige `huh.Input`) einen eigenen NewLine-Tastenweg
(`alt+enter`/`ctrl+j`) mit, über den ein echter `\n` in den Titel gelangen
kann — strukturell unmöglich mit `huh.Input`, also eine ECHTE neue
Verhaltensfläche durch B03s Feldtyp-Wechsel, vom ursprünglichen Planner-Text
hier nicht antizipiert. Gefixt über einen erweiterten `singleLineTitle`-
Validator (`nonEmpty` PLUS `strings.Contains(s, "\n")` → Ablehnung "title
must be single-line", keine stille Normalisierung) statt einer strukturellen
Verhinderung — bewusst auf dieses eine Formular gescoped, das Create-Form
bleibt `huh.Input` und damit weiterhin strukturell `\n`-frei.

**B04 — RELATIONS-Sektion: Dopplung raus, Pfeil-selektierbar, hängender
Einzug (drei Punkte, EINE zusammenhängende Änderung an
`relationsSectionBody`).**

1. Die separate `Fields:`-Zeile (`fieldStrip`, `accordion.go`) entfällt für
   die RELATIONS-Sektion (deren einziger verbleibender Aufrufer — `Meta`
   war laut `renderAccordion`s eigenem `i != 0`-Gate schon immer
   ausgeschlossen). `fieldStrip` selbst wird entfernt (Compiler-gesteuert,
   Muster PF-14/B13-Removal) — nach diesem Task gibt es keinen Aufrufer
   mehr.
2. Ersatz: `relationsSectionBody` bekommt die GLEICHE ▷/▶-Cursor-Konvention
   wie `metaSectionBody` (PF-4) — Signatur wächst um `(active bool, fieldIdx
   int)`, jede gerenderte Zeile (Parent/Children/Blocking/Blocked-By, in der
   bestehenden `fields`-Akkumulationsreihenfolge — die `fields`-Slice
   entsteht schon heute in exakt der Zeilen-Reihenfolge, `appendGroup`
   hängt `lines`/`fields` im Lockschritt an) trägt einen eigenen `▷ `/`▶ `-
   Marker VOR dem bisherigen `relationRow`-Text. Die Pfeiltasten-Navigation
   selbst (`keyDetailFocus`s `up`/`down` bei `detailLevel==1`, `fieldCursor`
   über `secs[m.secCursor].fields`) ändert KEINEN Code — sie zeigte immer
   schon auf dieselbe `fields`-Slice, nur die vorherige Fehlvorstellung war,
   dass nur die (jetzt entfernte) `Fields:`-Strip diese Slice visualisierte.
   Mit dem Fix zeigt die EINZIGE verbleibende Visualisierung (die Zeilen
   selbst) den Cursor — PO-Befund „nur Fields wählbar" löst sich auf, weil
   es danach nur noch EINE Repräsentation gibt.
3. Hängender Einzug (PO-Mockup, Zitat oben in `bt-tct9`): neuer Helfer
   `hangingIndentWrap(prefix, text string, w int) string` (`view_detail_
   bean.go` oder `render_shared.go`) — `prefix` ist die bereits
   VOLL-gestylte Glyphen+ID-Zeile (`▷ `/`▶ ` + `theme.StatusIcon` +
   `theme.TypeIcon` + `theme.Key.Render(id)` + Leerzeichen), ihre SICHTBARE
   Breite (`lipgloss.Width(ansi.Strip(prefix))`) bestimmt den Einzug der
   Folgezeilen — PRO ZEILE individuell (keine globale Tabellen-Spalte über
   alle Zeilen hinweg, da IDs unterschiedlich lang sein können,
   `<prefix>-<n-chars>`, design-spec.md §4-Nachbarschaft) — genau das PO-
   Mockup („die Übersicht ist gewahrt", nicht „alle Titel beginnen an
   exakt derselben Spalte"). Wortumbruch selbst via `ansi.Wordwrap`
   (gleiche Primitive wie `wrapText`, `view.go`), NICHT die blanket-
   `wrapText`-Anwendung auf den GESAMTEN verketteten Gruppen-Block
   (`relationsSectionBody`s bisheriges `wrapText(strings.Join(groups,
   "\n\n"), bodyW)` am Ende) — DAS war der eigentliche B04.3-Bug: der Titel
   einer Zeile lief über die Zeilenbreite hinaus und wickelte OHNE Bezug
   zum eigenen Präfix zurück auf Spalte 0. Jede Relation-Zeile wird jetzt
   VOR dem Join bereits fertig eingezogen umgebrochen; `wrapText` bleibt für
   die (unveränderten) Muted-Subheader-Zeilen (`Parent`/`Children`/…)
   zuständig.

`relationsSectionBody`s Aufrufer (`beanSections`, `view_detail_bean.go`)
gibt die zusätzlichen zwei Parameter durch (`focused && activeIdx ==
relationsSectionIdx`, `fieldIdx` — exakt das gleiche Muster wie
`metaSectionBody`s eigener Aufruf in derselben Funktion).

**B05 — Kopfblock-Meta-Strip zeigt zusätzlich Tags (bereits redefiniert,
`bt-tct9` „B05 REDEFINIERT"-Abschnitt, HIER erstmals formal in die
Design-Spec übernommen).** `detailHeaderBlock` (`view_detail_bean.go`)
erweitert die `type: …    status: …    prio: …`-Zeile um eine vierte Spalte
`    tags: <tagsInline(b.Tags)>` (PO-Mockup exakt: `type: epic  status:
in-progress  prio: !  tags: to-review`) — `tagsInline` (`render_shared.go`,
seit PF-15 kein toter Code mehr) wird HIER ein zweiter Aufrufer neben
`metaFields`. Variable Breite der Tags-Spalte ist unkritisch (Zeilenende,
kein nachfolgendes gepaddetes Feld) — B02s (E8) feste `type`/`status`-
Padding-Konvention bleibt für die ERSTEN drei Spalten unverändert, nur die
vierte kommt neu hinzu. Taglos: `theme.Dim.Render("(none)")`, mirrort
`metaFields`s eigene Konvention.

**B06 — Relations-Picker (Blocking `r`, Parent `a`) viel zu schmal
(PO-Fund, Screenshot: IDs brechen mitten im Wort um).** PO verbatim: „Das
overlay für 'r' ... So kann man die beans nicht sauber lesen und
bearbeiten. Höhe passt. Aber die Breite muss viel weiter werden." Beide
Picker rufen heute `clampModalWidth(48, m.width)` — eine reine
NACH-UNTEN-Clamp-Funktion (`min(pref, termW-4)`, nie nach oben), 48 ist
also faktisch die feste Breite auf jedem Terminal ≥52 Spalten. Neuer
Helfer `wideModalWidth(termW int) int` (`box_filter_facets.go`, direkt
neben `clampModalWidth`, gleiche Datei = Single Source für Modal-Breiten-
Helfer): `≈85% von termW`, Boden 60 (nie schmaler als der alte fixe Wert),
Deckel `termW-4` (dieselbe 2-Spalten-Rand-Konvention wie
`clampModalWidth`). `blockingPickerBox`/`parentPickerBox`
(`box_picker_blocking.go`/`box_picker_parent.go`) wechseln von
`clampModalWidth(48, m.width)` auf `wideModalWidth(m.width)`. Zeilen-
Rendering: dieselbe `hangingIndentWrap`-Logik wie B04.3 (Planner-
Entscheidung, EIN Wrap-Helfer für beide Stellen — Parent-/Blocking-Picker-
Zeilen sind strukturell identisch zu Relations-Zeilen: `▌`/` `-Cursor-
Präfix statt `▷`/`▶`, sonst dieselbe Glyphen+ID+Titel-Form) statt der
bisherigen Ein-Zeilen-`it.label`-Konkatenation — verhindert exakt den vom
PO gezeigten Mitten-in-der-ID-Umbruch. Höhe (`parentPickerRowBudget = 14`)
bleibt unverändert (PO: „Höhe passt").

## F01 — Vollbild-Navigation (`v`), Zustandsmodell + Navigations-Pfad

PO-Idee, verbatim in `bt-tct9` übernommen. Größtes Einzelstück dieser
Runde — Zustandsmodell hier vollständig ausformuliert (Folgefragen
antizipiert), Umsetzung in zwei Tasks (`epic-E9-plan.md`).

### Zustandsmodell

Neuer, zu `m.view` (Browse/Backlog/Lobby, unverändert) ORTHOGONALER
Modus — Vollbild ändert NICHT, welcher `viewID` aktiv ist, nur WIE dessen
Master-Detail-Paar gerendert wird:

```go
type fullscreenMode int
const (
    fullscreenNone fullscreenMode = iota
    fullscreenList                // linkes Pane (Tree ODER Backlog-Liste, je nach m.view) — Vollbreite
    fullscreenDetail               // Detail-Accordion eines EINEN Beans — Vollbreite
)
```

Neue Modelfelder (`types.go`): `fullscreen fullscreenMode` ·
`fullscreenBeanID string` (das in `fullscreenDetail` gezeigte Bean —
UNABHÄNGIG vom Tree-/Backlog-Cursor, s. u.) · `navBack, navForward
[]string` (Bean-ID-Stacks, s. History-Stack unten).

### Einstieg (`v`)

Neue Bindung `keys.Fullscreen` (`v`, frei — verifiziert gegen
`keymap.go`, keine Kollision). Dispatch-Punkt: `handleKey` (`update.go`),
NACH dem `FocusIn`/`FocusOut`-Block (gleiche Fokus-/Ansichts-Modus-
Familie), VOR `keyNodeAction` — greift NUR wenn `m.view != viewLobby` UND
kein Full-Capture-State aktiv ist (identische Guard-Kette wie jeder andere
Top-Level-Key: Suche/Filter/Form/Overlay/Palette/Help bereits vorher
abgefangen, Bestandsmuster unverändert). Verhalten:

- `m.fullscreen == fullscreenNone && !m.detailFocus` → `m.fullscreen =
  fullscreenList` (Browse/Backlog+links-fokussiert → Beans-Liste Vollbild —
  PO-Wortlaut).
- `m.fullscreen == fullscreenNone && m.detailFocus` → `b :=
  m.focusedBean(); if b == nil { no-op }` sonst `m.fullscreen =
  fullscreenDetail; m.fullscreenBeanID = b.ID` (Browse/Backlog+rechts-
  fokussiert → Detail-View Vollbild). `m.detailFocus`/`m.secCursor`/
  `m.fieldCursor`/`m.detailLevel` bleiben UNVERÄNDERT (dieselbe
  Detail-Fokus-Maschine wird im Vollbild weiterverwendet, s. u.) — kein
  Reset beim Eintritt.
- `m.fullscreen != fullscreenNone` (bereits im Vollbild) → No-Op
  (Planner-Entscheidung: `v` ist ein EINWEG-Einstieg, kein Toggle — PO
  definiert `esc` explizit als den Ausstiegsweg, s. u.; ein zweites `v`
  spekulativ als Ausstieg zu belegen würde eine vom PO nicht verlangte
  zweite Bedeutung einführen, genau die Klasse Überraschungsverhalten,
  die PF-13/B01 bereits einmal beheben musste).

Gilt SYMMETRISCH für Browse (Tree) UND Backlog (Planner-Entscheidung,
Begründung: `m.view` bleibt beim Eintritt unverändert, die Vollbild-
Renderfunktion (unten) fragt nur „welche Listen-Rows/welches Bean", nicht
„welche View" — eine Tree-only-Beschränkung wäre eine künstliche, vom
PO-Wortlaut nicht geforderte Asymmetrie und würde `renderBeanAccordionPane`s
bereits etablierte View-Agnostik konterkarieren).

### Listen-Vollbild → Detail-Vollbild (`enter`)

`enter` auf einem Bean IM Listen-Vollbild (`m.fullscreen ==
fullscreenList`) verhält sich WIE der Sprung, den `activateDetailField`s
Relations-Case heute für einen Jump macht, nur als neuer Einstiegspunkt:
`b := <Bean unter dem aktuellen Tree-/Backlog-Cursor>; if b == nil ||
(Tree-Node ist Blatt-loser Cursor) { no-op }` sonst `m.fullscreen =
fullscreenDetail; m.fullscreenBeanID = b.ID` + Detail-Fokus-Maschine auf
Meta/Section-Ebene reinitialisiert (`m.secCursor, m.accOpen,
m.detailLevel, m.fieldCursor = 0, 1, 0, 0` — identisch zu `FocusIn`s
bestehendem Reset-Muster, `handleKey`). Dispatch-Ort: `keyFullscreen`
(neue Datei `view_fullscreen.go`) prüft `m.fullscreen` VOR den bestehenden
`keyTree`/`keyBacklog`-Cursor-Bewegungen — Auf/Ab/Rechts/Links im
Listen-Vollbild bleiben SONST identisch zum Split-Modus (dieselben
`treeCursorMove`/`backlogCursorMove`-Helfer, wiederverwendet, keine
Duplikat-Navigation).

### Detail-Vollbild: Feld-Navigation + Relations-Sprung (`enter`, setzt B04.2 voraus)

Innerhalb `fullscreenDetail` wird `keyDetailFocus` (`update.go`) VERBATIM
wiederverwendet — `m.focusedBean()` bekommt dafür einen NEUEN, VOR dem
bestehenden `m.view`-Switch geprüften Fall: `if m.fullscreen ==
fullscreenDetail { return idx.ByID[m.fullscreenBeanID] }` (view-agnostischer
Dispatcher, gleiches Prinzip wie der bestehende Tree/Backlog-Switch selbst).
Damit funktioniert JEDE bestehende Feld-Kaskade (PF-5, status/type/
priority/tags-Overlays, Titel-Form) im Vollbild UNVERÄNDERT. Der einzige
NEUE Fall ist `activateDetailField`s Relations-Sprung-Zweig (`default:`-Case,
`f.beanID != ""`): innerhalb `fullscreenDetail` darf ein Sprung NICHT
`m.detailFocus = false` setzen (das bestehende Verhalten, das im Split-
Modus zurück zum Tree/Backlog springt) — stattdessen: History-Push (unten)
+ `m.fullscreenBeanID = f.beanID` (bleibt in `fullscreenDetail`, zeigt das
NEUE Ziel-Bean, Detail-Fokus-Maschine reinitialisiert wie beim Listen-
Vollbild-Einstieg oben). Voraussetzung B04.2 (Relations-Einträge selbst
per Pfeil erreichbar, s. oben) — ohne B04.2 gäbe es im Vollbild keinen
Weg, überhaupt einen Relations-Eintrag zu selektieren, `blocked_by` in
`epic-E9-plan.md` entsprechend gesetzt.

### Ausstieg (`esc`)

**Entscheidung (Planner, PO-Wortlaut wörtlich genommen):** `esc` verlässt
das Vollbild IMMER direkt zurück zu Browse/Backlog (Split-Modus) — NICHT
schrittweise zurück durch die Relations-Sprung-Kette (das ist exklusiv
`ctrl+links`/`[`s Aufgabe, s. History-Stack unten). Begründung: PO-Zitat
„esc kehrt zum Browse zurück (aus Detail-Vollbild via Relations-Sprung:
zurück zum Browse mit dem AKTUELLEN Bean selektiert)" — der Klammerzusatz
ist ein Beispiel-Fall (der am häufigsten auftretende: nach N
Relations-Sprüngen), keine Sonderregel NUR für diesen Fall; die
Hauptaussage „esc kehrt zum Browse zurück" gilt uniform, unabhängig vom
Eintrittspfad (direktes `v`, `enter` aus Listen-Vollbild, N
Relations-Sprünge) — ALLES landet auf demselben, einfachen Exit. Eine
pfadabhängige Fallunterscheidung („esc geht zurück zum Listen-Vollbild,
WENN von dort betreten, sonst zu Browse") wäre eine vom PO nicht verlangte
dritte esc-Zielsemantik und exakt die Klasse impliziter Zustands-
Abhängigkeit, die B01/PF-13 bereits einmal als „für Nutzer murks" verworfen
hat.

Konkret fügt sich das NAHTLOS in D03s bestehendes „esc = eine Ebene pro
Druck"-Modell (PF-16) ein, OHNE dessen Kontrakt zu brechen — Vollbild trägt
GENAU SO VIELE Ebenen bei, wie es Rasten hat:

| Zustand | `esc` |
|---|---|
| `fullscreenDetail`, `detailLevel==1` (Feld-Ebene) | → `detailLevel=0` (Sektions-Ebene) — bestehende D03-Rast, unverändert |
| `fullscreenDetail`, `detailLevel==0` (Sektions-Ebene) | → Vollbild verlassen: `m.cursorID = m.fullscreenBeanID` (Tree) BZW. `m.backlogList`-Cursor auf `m.fullscreenBeanID` gesetzt (Backlog, je nach `m.view`) + `m.fullscreen = fullscreenNone` + `m.detailFocus = true` (Split-Detail bleibt fokussiert, PO: „mit dem AKTUELLEN Bean selektiert" impliziert die Detailansicht bleibt sichtbar, nicht Rücksprung auf den Tree-Fokus) — **Ausnahme (Supervisor-Entscheid, T7-Review Fix-Runde 1, bean `bt-13l7`, F02):** ist `m.view == viewBacklog` UND das zuletzt gezeigte Bean NICHT `backlogVisible()` (z. B. weil es einen Parent hat oder Status weder `todo` noch `draft`), wechselt der Exit stattdessen auf `m.view = viewBrowseRepo` + `expandAncestorsOf(...)` + `m.cursorID = m.fullscreenBeanID` (Tree kann JEDES Bean zeigen, Backlog nicht) — erfüllt „mit dem AKTUELLEN Bean selektiert" maximal statt den Backlog-Cursor still auf ein anderes Bean stehen zu lassen. Nur im backlogVisible-Fall bleibt der oben beschriebene reine Backlog-Sync. |
| `fullscreenDetail`, `m.fullscreenBeanID` zeigt auf ein extern verschwundenes Bean (Watch-Reload/Repo-Wechsel/Delete) | → **Ergänzung (T7-Review Fix-Runde 1, bean `bt-13l7`, F01):** der bestehende `b==nil`-Guard in `keyDetailFocus` setzt zusätzlich `m.fullscreen = fullscreenNone`, statt nur `m.detailFocus = false` — ohne diesen Zusatz wäre das Vollbild sonst eine Sackgasse (jede Detail-Taste bliebe wirkungslos, einziger Ausweg wäre Quit). Vom ursprünglichen Planner-Text nicht antizipierter Randfall, keine Änderung der drei Haupt-Zeilen dieser Tabelle. |
| `fullscreenList` | → Vollbild verlassen: `m.fullscreen = fullscreenNone` (Cursor war nie entkoppelt, keine Sync nötig — Split-Ansicht zeigt exakt denselben Stand) |

**Revidiert (Supervisor-Entscheid, E9 Task 8, bean `bt-1vbp`, 2026-07-16):**
entgegen dem ursprünglichen Planner-Text hier („NICHT geleert", s. History
darunter) werden `navBack`/`navForward` bei JEDEM Vollbild-Exit GELEERT —
an allen drei Choke-Points (`keyDetailFocus`s `b==nil`-Guard, `keyDetailFocus`s
Back-Case beim Sektions-Ebenen-Exit aus `fullscreenDetail`, `keyFullscreen`s
`fullscreenList`-Exit, jeweils `update.go`/`view_fullscreen.go`). Grund:
History-Leak-Vermeidung zwischen unabhängigen Vollbild-Sessions — ein
erneutes `v` (auf demselben ODER einem anderen Bean) darf keine STALE
History aus einer bereits verlassenen Session mitschleppen. Innerhalb EINER
laufenden `fullscreenDetail`-Session bleibt die unten beschriebene Push-/
Kappungs-Semantik (frischer Sprung kappt `navForward`) unverändert gültig.

### History-Stack (`ctrl+links`/`ctrl+rechts`, Fallback `[`/`]`)

**Scope-Entscheidung (Planner, explizit zu bestätigen):** der Stack trackt
AUSSCHLIESSLICH Relations-Sprünge INNERHALB `fullscreenDetail` (der vom PO
beschriebene „Navigations-Pfad" entsteht literal nur dort — B04.2 s.o.).
Der bestehende Split-Detail-Sprung (`activateDetailField`s Relations-Case
AUSSERHALB von Vollbild, E2-Ära, springt zum Tree UND verlässt
`detailFocus`) bleibt UNVERÄNDERT und speist die History NICHT — ein
dokumentierter, bewusster Schnitt (kein PO-Wortlaut verlangt History für
den Split-Modus, und der Split-Modus-Sprung wechselt ohnehin die Cursor-
Repräsentation selbst, nicht nur den angezeigten Bean).

- **Push (bei jedem Relations-Sprung INNERHALB `fullscreenDetail`):**
  `m.navBack = append(clone(m.navBack), <bisheriges m.fullscreenBeanID>)`,
  `m.navForward = nil` (frischer Sprung kappt die Vorwärts-Historie — Standard-
  Browser-Semantik, dasselbe Modell wie jeder Web-Browser-Back-Button).
- **`ctrl+links` / `[` (neu `keys.HistoryBack`):** No-Op wenn `navBack`
  leer ODER `m.fullscreen != fullscreenDetail`. Sonst: letztes Element aus
  `navBack` poppen, `m.fullscreenBeanID` (VOR dem Pop) auf `navForward`
  pushen, `m.fullscreenBeanID` = gepopptes Element, Detail-Fokus-Maschine
  auf Sektions-Ebene reinitialisiert (wie jeder andere Ziel-Wechsel oben).
- **`ctrl+rechts` / `]` (neu `keys.HistoryForward`):** symmetrisch
  (`navForward` poppen, aktuelles Bean auf `navBack`, Ziel = gepopptes
  Element).
- Beide NUR wirksam wenn `m.fullscreen == fullscreenDetail` — in
  `fullscreenList`/Split-Modus sind sie unbelegte No-Ops (keine Kollision,
  `ctrl+links`/`ctrl+rechts` sind heute nirgends gebunden, verifiziert
  gegen `keymap.go`).
- **Ergänzung (Supervisor-Entscheid, E9 Task 8, bean `bt-1vbp`, D02 — Lücke
  weder im Bean noch ursprünglich hier spezifiziert):** ein per `navBack`/
  `navForward` referenziertes Bean kann extern verschwinden (Watch-Reload,
  Repo-Wechsel, paralleler Agent-Delete — dieselbe Trigger-Klasse wie der
  `b==nil`-Guard oben). Back/Forward LOOPEN über solche toten Einträge
  hinweg (Existenz-Check je Pop gegen `idx.ByID`) statt darauf zu landen —
  ein leerer ODER vollständig toter Stack ist ein sauberer No-Op, kein
  Trap (dieselbe „No-Op mit sauberem Zustand"-Linie wie der `b==nil`-Guard).

**Terminal-Verfügbarkeit (PO-Implementierungshinweis, geprüft):**
`bubbletea` v1.3.10 dekodiert `ctrl+left`/`ctrl+right` bereits nativ
(`key.go`: `KeyCtrlLeft`/`KeyCtrlRight`, aus den Standard-xterm-CSI-
Sequenzen `\x1b[1;5D`/`\x1b[1;5C` sowie deren `urxvt`/Alt-Varianten) — die
Bindung FUNKTIONIERT in den meisten modernen Terminal-Emulatoren
(iTerm2, Alacritty, Kitty, WezTerm, xterm mit `modifyOtherKeys`) direkt.
Bekanntes Risiko (PO-Hinweis bestätigt): ältere Terminals, manche
tmux-Konfigurationen ohne `xterm-keys on` sowie SSH-Ketten mit
abweichendem `TERM` können die Sequenz verschlucken oder unverändert als
bloßes `left`/`right` durchreichen. **Fallback (Planner-Entscheidung):**
`keys.HistoryBack`/`keys.HistoryForward` binden JEWEILS ZWEI Tasten
(`keybind.WithKeys("ctrl+left", "[")` / `keybind.WithKeys("ctrl+right",
"]")`) — `[`/`]` sind im gesamten Keymap unbelegt (verifiziert), terminal-
unabhängig garantiert zustellbar (einfache Druckzeichen, keine
Modifier-Sequenz) und ein in TUI-Tools verbreitetes Back/Forward-
Idiom (u. a. `k9s`, `lazygit`-artige Tools) — kein exotisches Neu-
Idiom. Sichtbar gemacht (PO: „im Footer/Help ausweisen"): NEUE
`fullscreenDetailLocalBindings()` (`footer_context.go`-Konvention,
kontextsensitiver Footer) zeigt `[`/`]` (gerendert über
`renderBindings`, das ohnehin BEIDE Tasten einer Bindung im Kurz-Label
zusammenfasst) NUR während `m.fullscreen == fullscreenDetail` — im
Split-Modus (wo die Tasten wirkungslos sind) tauchen sie nicht auf,
zusätzlich vollständig in `helpGroups()` (Navigation-Gruppe) dokumentiert.

### Maus im Vollbild — explizites Nicht-Ziel (v1, dokumentierter Scope-Cut)

Der PO-Wortlaut beschreibt F01 ausschließlich über Tastatur-Flows (`v`/
`enter`/`esc`/`ctrl+arrow`) — volle Klick-Unterstützung im Vollbild (neue
Geometrie-Berechnung analog `clickPaneGeometry`, aber für eine EINZELNE
Vollbreite-Pane statt des Splits) ist ein separates, nicht angefragtes
Stück Arbeit und wird NICHT in dieser Runde gebaut (YAGNI, kein PO-Bedarf
geäußert). Sicherheitsnetz statt Vollimplementierung: `handleMouse`
(`mouse.go`) bekommt EINEN neuen Guard direkt NACH dem bestehenden
Toast-Klick-Vorrang (der bleibt unbedingt, unverändert) und VOR dem
bisherigen Overlay-Guard — `if m.fullscreen != fullscreenNone { return m,
nil }` — verhindert, dass ein Klick im Vollbild gegen die (dann falsche)
Split-Geometrie fehl-interpretiert wird und eine falsche Zeile/Sektion
trifft. Wheel/Klick sind im Vollbild damit funktionslos, aber NIE
falsch — ein bewusster, dokumentierter Nicht-Ziel-Punkt, kein stiller
Bug. Vollständige Maus-Unterstützung ist ein Fast-Follow-Kandidat
außerhalb dieser Runde (kein bean angelegt, YAGNI bis PO-Bedarf).

### Rendering

Neue Datei `view_fullscreen.go`: `keyFullscreen(msg) (bool, tea.Model,
tea.Cmd)` (Dispatch, `v`/`enter`-im-Listen-Vollbild/`esc`/History-Keys,
handled-Flag analog `keyNodeAction`s Signatur) + `renderFullscreenBody(w,
h int, head, localKeys string, listRows []string, detailBean *data.Bean)
string` — EIN gemeinsamer Renderer für beide Vollbild-Spielarten:
`viewBrowseRepo()`/`viewBacklog()` prüfen `m.fullscreen != fullscreenNone`
an ihrem jeweiligen Kopf (VOR dem bestehenden `JoinHorizontal(listBox,
detailBox)`-Aufbau) und reichen entweder ihre eigenen (unverändert
berechneten) `listRows` ODER `nil` (im `fullscreenDetail`-Fall, dann zieht
der Renderer `detailBean` über `renderBeanAccordionPane` mit VOLLER
`innerW` statt der Split-`rw`) durch — Chrome (Breadcrumb/Footer/
Statuszeile) bleibt IDENTISCH zum jeweiligen View (`browseRepoChrome`/
`backlogChrome`, unverändert), nur der Body-Teil wird eine einzelne
Vollbreite-Pane statt zwei nebeneinander. Kein neuer `viewID`-Enum-Wert
(bestätigt gegen `types.go`: `m.view` bleibt `viewBrowseRepo`/`viewBacklog`
— Vollbild ist orthogonal, s. Zustandsmodell oben).

## 16. Tag-Management (E10) — zentrale Tag-Definition (v1.1, bean `bt-6oyy`, Epic `bt-362n`)

Nachgeliefert 2026-07-16 (Tasks T1-T6, Epic `bt-362n`). Superseded damit §4s
„Tag-Manager-CRUD entfällt" und §9s „OUT (v1): Tag-Manager-CRUD" (dort
markiert, nicht gelöscht — PF-14-Muster). Dieser Abschnitt dokumentiert
AUSSCHLIESSLICH die umgesetzten Entscheide D01-D15 (Epic-Body `bt-362n`)
plus die dokumentierten Deviations der T-beans (`bt-49hh`/`bt-r92i`/
`bt-604w`/`bt-1lsu`/`bt-y9my`/`bt-pqq3`) — keine neuen Design-Änderungen.

### Registry (Persistenz)

- **Datei:** `.beans-tags.yml` im Repo-Root (Sibling zu `.beans.yml`, via
  `client.RepoDir`, keine eigene Discovery). Bewusst NICHT in `.beans/` —
  komplett entkoppelt von beans' eigenem Scan/Autorität (D01).
- **Format:** minimal `tags: [<name>, …]`, alphabetisch sortiert
  gespeichert (deterministische Diffs). Kein Farb-/Beschreibungsfeld —
  Tag-Farbe bleibt Hash-aus-Name (§4/§8) (D01).
- **Laden:** tolerant-missing/tolerant-corrupt — fehlende Datei → leere
  Registry, korrupte YAML → leere Registry, nie crashen; synchroner
  `os.ReadFile` (mirrort `config.LoadSettings`) (D02).
- **Liveness:** frisch von Platte bei jedem Page-Open (`openTagManagementPage`)
  bzw. Picker-Open (`openTagPicker`) — KEINE fsnotify-Erweiterung; externe
  Änderungen werden erst beim nächsten Open sichtbar (dokumentierter
  Trade-off, identisch zur Lobby) (D03, D10).
- **Schreiben:** direktes `os.WriteFile`, NICHT atomar (kein temp+rename) —
  akzeptierter, dokumentierter Trade-off (mirrort `internal/config`
  settings/state); D02s tolerant-corrupt-Load degradiert einen Crash
  mitten im Write zu leerer Registry statt Panic (T1-Review F02).
- **API:** `internal/data/tagdefs.go` — `Client.LoadTagDefs()`/
  `Client.SaveTagDefs([]string)` + pure, slice-safe Helfer `AddTagDefName`/
  `RemoveTagDefName`/`RenameTagDefName` (D04). Namens-Validierung
  (`data.ValidTagName`) liegt bewusst an der EINGABEGRENZE, nicht in der
  Persistenzschicht (T1-F04-Merkposten, Epic-Body).

### Page (`viewTagManagement`)

- **Einstieg:** ausschließlich Command-Center „go to tags" (mirrort „go to
  settings" — keine eigene globale Taste, Tastenraum knapp) (D05).
- **Capture:** FULL-CAPTURE wie die Lobby, geprüft an derselben
  `handleKey`-Stelle (früh, vor `ctrl+k`/`?`/`p`) — globale
  Node-Action-Tasten können nie gegen ein stale Bean feuern; `esc` ist der
  Rückweg (D06).
- **Chrome:** `Chrome()`/`renderPane()`-Einzel-Listen-Pane (Backlog-Stil,
  nicht Lobbys Banner-Zentrierung); `GlobalHint` bewusst LEER — keine der
  globalen Tasten funktioniert unter Capture, sie anzuzeigen wäre ein
  nicht-funktionales Versprechen (D07). Footer-Zone-3: `↑/↓ · n · d · e ·
  esc`.
- **Layout:** Einzel-Pane-Liste, KEIN Master-Detail-Drilldown; `enter` ist
  dokumentierter handled No-Op, reserviert für den Fast-Follow „Beans zu
  diesem Tag" (`idx.WithTag` existiert bereits) (D08).
- **Zeilenmodell:** UNION aus (a) definierten Tags (Registry, auch Count 0)
  und (b) freien (unregistrierten) Tags in Verwendung. Sortierung:
  Definiert alphabetisch zuerst, dann Frei count-absteigend. Jede Zeile
  trägt `defined bool`; definierte Zeilen tragen Marker `✓` (Green) in
  einer IMMER reservierten Marker-Spalte (PF-12) (D09). Glyph `✓` ist
  Implementer-Entscheid T2 — der ursprüngliche Plan-Text nannte `●`
  (T6-Review F02, dokumentierte Abweichung).
- **Stale-Grenze:** `applyLoaded` baut `tagMgmtRows` nicht reaktiv neu —
  eine bereits OFFENE Page zeigt nach Rename/externem Reload alte Counts
  bis zum nächsten Page-Open (kein Bug: D03 deckt nur „frisch bei jedem
  Open"; dokumentierter Fast-Follow-Kandidat, T5-Deviations `bt-y9my`).

### Create (`n`)

- Page-lokaler Freitext-Input-Submodus (`tagMgmtInputActive`/
  `tagMgmtInputMode` „create"|„rename", EIN dauerhaftes `textinput.Model`,
  mirrort den Tag-Picker-Input); `esc` verwirft NUR den Submodus (D14).
- Validierung `data.ValidTagName` wörtlich an der Eingabegrenze; Dedupe
  gegen ALLE Zeilennamen (definierte UND freie — T3-Deviation `bt-604w`:
  dem literaleren Bean-Wortlaut gefolgt statt D11s engerem
  „Registry-Einträge"); Fehlertext neutral `name already in use: X`
  (T5-F01). Ein Fehler lässt den Submodus für Retry offen.
- Eine neue Definition BERÜHRT KEIN Bean (kein Bean bekommt den Tag
  automatisch) — reiner Registry-Akt, ab sofort im Picker
  sichtbar/priorisiert (D11).
- Submit dispatcht `saveTagDefsCmd` mit EXPLIZITEM `refindName` (jede
  Dispatch-Site benennt ihr Cursor-Ziel selbst, nie impliziter
  Input-Feld-Read — T4-Fix-Runde B01, `bt-1lsu`); der Submodus schließt
  erst nach BESTÄTIGTEM Write (`applyTagDefsSaved`).

### Delete (`d`)

- REGISTRY-ONLY (D12): Beans, die den Tag tragen, BEHALTEN ihn — er wird
  wieder ein freier Tag, im Suggest-Mode weiter erlaubt, nur unpriorisiert.
  Kein destruktiver Strip-Modus (→ Q01).
- Page-lokales Bool+Ziel-Paar `tagMgmtDeleteConfirm`/`tagMgmtDeleteTarget`,
  KEIN neuer `overlayID`-Case (mirrort `confirmQuit`) (D15).
- Confirm zeigt den LIVE-Verwendungszähler (render-zeitige Auflösung aus
  `tagMgmtRows`, kein drittes Feld — T4-Deviation), Singular/Plural
  korrekt („Still used by 1 bean"). `d` auf freier Zeile = No-Op.
- Cursor nach Delete: `refindName` = gelöschtes Ziel — der Cursor FOLGT
  einem noch benutzten Tag in die Frei-Gruppe; ein unbenutzter Tag
  verschwindet, der Cursor bleibt stehen (T4-Fix-Runde B01).

### Rename (`e`, Propagation über alle Beans)

- NEUES Binding `keys.RenameTag` auf `e` — kollisionsfrei zum globalen
  `keys.Editor` (disjunkter Full-Capture-Kontext, Backlog-`S`-Präzedenz).
  Nur auf definierten Zeilen (No-Op auf freien — erst per `n` definieren).
- Nutzt DENSELBEN Input-Submodus, vorbefüllt mit dem alten Namen; der
  Dedupe-Check exkludiert den eigenen alten Namen (`tagMgmtInputTarget`) —
  Re-Confirm des unveränderten Namens wird nie als Selbst-Duplikat
  abgelehnt (D13, D14).
- **Registry-Rename ZUERST** (lokaler Datei-Op), UNABHÄNGIG vom Ausgang
  des Bean-Sweeps. **Sweep:** async `renameTagCmd`, best-effort
  CONTINUE-ON-ERROR über `idx.WithTag(alt)` — je betroffenem Bean EIN
  `SetTags(id, add=[neu], remove=[alt], etag)`-CLI-Aufruf (bestehende
  Methode; die beans-CLI kennt keinen Bulk-/Rename-Verb und keine
  Cross-Bean-Transaktion — ein stales ETag auf Bean K bricht K+1..N nicht
  ab). Ergebnis (renamed/failed, erste Fehlermeldung) als EIN Toast;
  danach `m.idx`-Reload (D13).
- **Same-Name-Guard** (T5-Deviation `bt-y9my`, Datenverlust-Schutz):
  `alt == neu` → No-Op OHNE jeden `SetTags`-Aufruf (dessen dokumentierter
  „remove gewinnt"-Resolver hätte den Tag sonst still von JEDEM
  betroffenen Bean entfernt). Zusätzlich `idx == nil`-Guard.
- Rename AUF den Namen eines existierenden freien Tags wird per Dedupe
  abgelehnt (= Merge-Semantik, bewusst nicht in v1.1 → Q04).

### Tag-Picker Suggest-Mode (`t`)

- `collectTagCounts` erweitert um `defined map[string]bool`; Sortierung:
  „defined" als neuer PRIMÄRER Schlüssel (definierte Tags zuerst), der
  bestehende Count-desc/alpha-Tie-Break bleibt SEKUNDÄR je Gruppe.
  Registry frisch bei `openTagPicker()` (D10).
- Marker-Spalte IMMER reserviert (PF-12): definiert `✓` (geteilte
  Konstante `tagManagementMarkerGlyph` aus der Page), frei gleich breiter
  Leerraum — Breiten-Stabilität über `lipgloss.Width` getestet, nicht über
  Byte-Offsets (3-Byte-UTF-8-Glyphs, T6-ERRATUM `bt-pqq3`).
- Freie Tags bleiben voll erlaubt, toggle- und speicherbar — KEIN strict
  mode; die bestehende Picker-`n`-Neuanlage (B14/E8) bleibt UNVERÄNDERT
  (→ Q03).

### Offen (PO) — Q01-Q04, nicht blockierend

| # | Frage | v1.1-Stand |
|---|---|---|
| Q01 | Destruktiver Delete-Modus (Tag zusätzlich von jedem Bean strippen, GitHub-Label-Semantik)? | Registry-only (D12) |
| Q02 | Definierte Tags mit Count 0 als „Aufräum-Kandidat" markieren? | schlichte Count-0-Anzeige (D09) |
| Q03 | Picker-`n` (B14/E8) soll künftig auch die Registry befüllen? | B14 unverändert (T6) |
| Q04 | Rename auf existierenden freien Namen = Merge erlauben? | per Dedupe abgelehnt; `data.RenameTagDefName` ließe es datenseitig zu |
