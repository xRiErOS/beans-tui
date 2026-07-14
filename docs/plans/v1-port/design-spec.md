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
- **Datenlayer: beans als Go-Library, in-process.** `github.com/hmans/beans@v0.4.2`
  (= installierte Binary-Version) via `pkg/beancore` (Core: Laden, Mutationen, fsnotify-Watch
  + Subscribe) und `pkg/bean`/`pkg/config` (Modell, Enums, Farben, Sortierung).
  Kein HTTP, kein Subprocess, kein Fork. GraphQL-Layer (`pkg/beangraph`) nur falls Filter-
  Wiederverwendung lohnt — primär direkt gegen `beancore.Core`.
- **Konsistenz:** Schreiboperationen über dieselben Core-Pfade wie die beans-CLI; ETag/
  optimistic concurrency (`if-match`) wird bei Formular-Saves genutzt (Konflikt → Toast + Reload).
- **Live-Reload:** `Core.StartWatching()` + `Subscribe()` — externe Änderungen (Agent,
  `beans update` in anderem Terminal) erscheinen ohne Neustart. Kern-Feature für
  „besser mit Agenten interagieren".

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

## 5. Review-Flow-Konvention (Agent↔PO)

beans kennt keinen Review-Status. Konvention (lean-stack-konform: „der ausführende Agent
schließt NICHT", PO-Autorität am Gate):

| Schritt | Wer | beans-Operation |
|---|---|---|
| Arbeit fertig, Abnahme angefragt | Agent | Tag `to-review` setzen, Status bleibt `in-progress` |
| Review-Queue ansehen | PO (TUI) | Review-Cockpit: alle beans mit Tag `to-review`, gruppiert nach Epic |
| Pass | PO (TUI) | Status → `completed`, Tag `to-review` entfernen |
| Reject | PO (TUI) | Tag `to-review` → `rework`; Kommentar wird als `## Review <datum>`-Abschnitt an Body angehängt |
| Rework fertig | Agent | Tag `rework` → `to-review` |

Die TUI ist damit das Merge-Gate-Cockpit: Agenten liefern in `.beans/`, der PO nimmt in der
TUI ab — Live-Reload macht Agent-Fortschritt in Echtzeit sichtbar.

## 6. Views (Port-Matrix)

| # | beans-tui-View | devd-Vorlage | Inhalt |
|---|---|---|---|
| V1 | Lobby | `viewHome` | ASCII-Logo `beans` + Repo-Picker (Suchfeld + Tabelle Name·Pfad·Offen/Gesamt); entfällt bei Einzel-Repo-Start |
| V2 | Browse (Primat-View) | `viewBrowseProject` | Master-Detail: links Tree Milestone→Epic→Task (expandierbar, Such-/Filterkopf), rechts Detail-Accordion; Fokus-Tausch mit Border-Farbwechsel |
| V3 | Backlog | `viewBrowseBacklog` | Flache Liste parentloser/ready beans, Master-Detail, Sort (status/prio/created/updated) |
| V4 | Detail-Accordion | `accordion.go` + `viewDetailIssue` | Sections: Meta (Status/Type/Prio/Tags) · Body (glamour) · Beziehungen (Parent/Children/Blocking/BlockedBy) · Historie (created/updated/ETag); Ziffern-Sprung `1…9` |
| V5 | Command-Center | `viewCommandCenter` + `overlay_palette` | `ctrl+k`: fuzzy Aktionen (Create, View-Wechsel, Status…) + Bean-Suche (Bleve) in einem |
| V6 | Review-Cockpit | `viewReviewSprint` + `viewNavigateReviews` | Queue `to-review` (links, Verdikt-Dots), Detail rechts, Summary-Zeile „x of n"; `a` pass · `x` reject+Kommentar · `o` reopen |
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
| US-08 | Als PO fahre ich den Review-Flow mit Agenten. | `R`-Cockpit listet `to-review`-beans je Epic; `a` → completed+Tag weg; `x` → Kommentar-Form, Tag `rework`, Feedback-Abschnitt im Body; Summary „x of n" |
| US-09 | Als PO sehe ich den Backlog priorisiert. | `b`: parentlose+ready beans, Sortierung umschaltbar, Master-Detail |
| US-10 | Als PO sehe ich Agent-Änderungen live. | Externe `.beans/`-Änderung (CLI/Agent) erscheint < 1 s ohne Neustart, Cursor-Position bleibt stabil |
| US-11 | Als PO übergebe ich Kontext an Agenten per Clipboard. | `y` kopiert Bean/Epic-Kontext als Markdown (ID, Titel, Status, Body) via OSC52; Toast bestätigt |
| US-12 | Als PO erlebe ich den devd-Look. | Catppuccin-Token §8, vierzonige Hülle, Master-Detail-Borders mit Fokus-Farbe, Accordion mit Chevron/Ziffern, Statusglyph `◉` |
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
| E4 Palette & Review | Command-Center, Review-Cockpit inkl. Tag-Konvention + Body-Feedback | US-04, US-08 |
| E5 Polish | Live-Reload, Yank/OSC52, Toast, Help, Maus, Settings+Repo-Picker+Lobby, ASCII-Fallback, Archiv-Sicht | US-10, US-11, US-13, US-14 |
| E6 Validierung & Release | Alle US validiert (Nachweise), README, Install-Doku, `bt` in lean-stack demo-startbar | Abnahme aller US |

## 13. Risiken

| Risiko | Behandlung |
|---|---|
| `pkg/`-API von beans v0.4.2 weicht vom beans-src-Klon ab (Klon neuer als Tag) | E1-Spike gegen das gepinnte Modul, nicht gegen den Klon; bei API-Lücken Feature lokal nachbauen statt upgraden |
| devd-Code-Port zieht API-Client-Reste mit | Port-Reihenfolge: Theme/Primitive zuerst, Views einzeln, Datenlayer von Anfang an beans-nativ (kein Adapter auf devd-Typen) |
| Bleve-Suche über Library ohne Server verfügbar? | E1-Spike verifiziert; Fallback: eigene Substring-/Fuzzy-Suche über geladene beans (im Speicher, unkritisch bei <1000 beans) |
| ETag/Watcher-Races (eigene Mutation triggert Reload) | Debounce vorhanden (100 ms); eigene Writes markieren+ignorieren, Cursor-Restore testen (US-10) |

## 14. Entscheidungen

| Code | Hintergrund | Entscheidung | Status |
|---|---|---|---|
| D01 | Stack-Wahl | Go/bubbletea, devd-TUI-Code als Port-Basis, kein Neuaufbau in Ink/TS | 🟢 |
| D02 | Datenzugriff | beans als Go-Library in-process (`pkg/beancore` u.a., Pin v0.4.2) — kein Fork, kein Subprocess, kein HTTP | 🟢 |
| D03 | Projekt-Ort | Eigenes Repo `tools/beans-tui/beans-tui-repository`, Binary `bt`, main-direkt (Worktree-Weiche, Solo+sequentielle Agent-Kette) | 🟢 |
| D04 | Review-Abbildung | Tag-Konvention `to-review`/`rework` + Body-Feedback-Abschnitte; PO schließt (completed) exklusiv via TUI | 🟢 |
| D05 | Entitäten-Reduktion | Memories/Docs/Notes/ToDos entfallen (Autoritäts-Trennung lean-stack: Wissen→OKF, Docs→Repo) | 🟢 |
| D06 | Tracking | Dogfooding: Entwicklung in `.beans/` dieses Repos; lean-stack erhält Verweis-bean | 🟢 |
| D07 | Vorherige „no custom TUI"-Entscheidung (`po-immersion-beans-via-obsidian-bases-no-custom-tui`, KC agent-memory) | Durch expliziten User-Auftrag superseded — KC-Konzept nach Abschluss via `/okf` aktualisieren | 🟢 |
