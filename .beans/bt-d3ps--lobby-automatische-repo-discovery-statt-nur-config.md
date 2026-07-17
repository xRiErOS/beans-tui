---
# bt-d3ps
title: 'Lobby: automatische Repo-Discovery statt nur config-Registrierung'
status: todo
type: feature
priority: normal
created_at: 2026-07-17T09:48:21Z
updated_at: 2026-07-17T10:08:09Z
parent: bt-5uzr
---

NB aus PO-Review E12 Runde 2 (2026-07-17, US-03-Kontext). PO wörtlich: „im lean-stack repo sind keine Projekte registriert. Daher muss hier eine automatische discovery ergänzt werden, damit ich von jeder beliebigen Stelle alle Projekte aufrufen kann."

Ist-Stand: Lobby-Repo-Liste speist sich ausschließlich aus registrierten Repos (config.yaml) — startet man bt in einem Repo ohne Registrierung, ist die Lobby leer/unvollständig.

Soll: Automatische Discovery von beans-Repos (Kandidaten-Mechanik zu klären: Scan definierter Wurzelverzeichnisse? `.beans.yml`-Suche? Merge mit config-Registrierung + Persistenz gefundener Repos?). Von jeder beliebigen Stelle aus müssen alle Projekte aufrufbar sein.

Offene Designfragen für Planner (vor Umsetzung PO klären):
- Discovery-Wurzel(n): fest (~/Obsidian/tools?), konfigurierbar, oder $HOME-weit mit Tiefenlimit?
- Gefundene Repos automatisch in config.yaml persistieren oder nur Session-flüchtig anzeigen?
- Performance: Scan bei Lobby-Öffnen (Latenz!) vs. Hintergrund/Cache.


## PO-Redefinition Grilling 2026-07-17 (ersetzt Discovery-Scan-Ansatz)

**KEIN automatischer Scan** (discovery_roots-Ansatz verworfen — „passt nur für mich, muss für andere Nutzende auch passen"). Stattdessen:
1. Lobby zeigt weiterhin genau die Repos aus der config (zentrales Register); Darstellung je Eintrag: `slug — Pfad`.
2. NEU: Command-Palette-Befehl `register project` — registriert das aktuell geöffnete Repo mit einem Griff im zentralen Register (config schreiben + Lobby-Liste aktualisieren).
3. Einmalig, außerhalb dieses beans: Eriks persönliche config.yaml wird um seine bestehenden Projekte ergänzt (Supervisor-Task, nicht Teil der Implementierung).

Damit obsolet: Scan-Wurzeln, Fund-Persistenz, Scan-Timing.


## Plan-Konkretisierung E13 (2026-07-17)

Plan: `docs/plans/v1-port/epic-E13-plan.md` §„Item 4: Lobby `slug — Pfad`
+ Palette-Befehl „register project"". Reihenfolge: Parallel-Welle (mit
`bt-2kfl`/`bt-nxuk`, disjunkte Dateien), NACH der Toast-Familie
(`bt-0xrb`/`bt-tm4a` — `overlay_palette.go`-Nachbarschaft, siehe Plan
Reihenfolge-Begründung).

PO-Redefinition (siehe eigene Sektion oben) bleibt final, NICHT neu
aufgemacht: KEIN Scan, Lobby zeigt weiterhin config-Repos (`slug — Pfad`),
NEUER Palette-Befehl „register project", Eriks eigene config.yaml-Ergänzung
außerhalb dieses beans.

**Root Cause / fehlender Baustein:** `view_lobby.go:174-189`
(`repoPickerBody`) rendert nur den rohen Pfad, kein Slug-Konzept existiert.
`.beans.yml` trägt bereits `beans.prefix` (ungenutzt von beans-tui).
`overlay_palette.go` hat keinen config-schreibenden Case.

**Vorgehen (Kurzfassung, Details im Plan):**
1. NEU `internal/data.RepoSlug(repoDir string) string` — liest
   `.beans.yml`, `beans.prefix` (trimmed) als Primärquelle, Dir-Basename als
   Fallback. Begründung: Bean-IDs sind an `beans.prefix` gekoppelt (überall
   im UI sichtbar), NICHT an `project.name` (fehlt z. B. bei lean-stack).
2. `view_lobby.go:174-189`: Zeilen-Label `data.RepoSlug(r) + " — " + r`.
3. `overlay_palette.go`: neuer `paletteActions`-Eintrag
   `"register_project"` (Zeile ~83-103-Nachbarschaft, neben
   `repo_picker`/`settings`) + neuer `dispatchPalette`-Case → neue Methode
   `m.registerProject()`: `m.client == nil` → no-op; bereits registriert →
   `toastInfo`; sonst `config.SaveUserSettings(append(m.settings.Repos,
   m.client.RepoDir), ...)` (bestehende Signatur, `internal/config/
   settings.go:147`) + `toastInfo`-Bestätigung. `openLobby()` lädt Settings
   bereits bei jedem Öffnen neu — keine separate Refresh-Verdrahtung nötig.

**Akzeptanz (siehe Plan für Volltext):**
- [ ] Lobby-Zeile zeigt `slug — Pfad` je config-Repo-Eintrag
- [ ] `RepoSlug()`: `beans.prefix` (trimmed) primär, Dir-Basename Fallback
- [ ] „register project" registriert `m.client.RepoDir`, dedupliziert
- [ ] Erfolgreiche Registrierung → Toast + Lobby zeigt neuen Eintrag beim
      nächsten Öffnen
- [ ] `m.client == nil` → no-op, kein Crash
- [ ] KEIN Scan, KEINE Scan-Wurzeln, KEINE Fund-Persistenz
- [ ] Test-Suite grün, neue Tests für `RepoSlug()` + `registerProject()`
- [ ] tmux-Smoke: lean-stack-Repo öffnen, „register project" ausführen,
      Lobby zeigt `lean-stack — <pfad>`
