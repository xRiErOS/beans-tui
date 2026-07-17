---
# bt-d3ps
title: 'Lobby: automatische Repo-Discovery statt nur config-Registrierung'
status: completed
type: feature
priority: normal
created_at: 2026-07-17T09:48:21Z
updated_at: 2026-07-17T11:45:07Z
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

## Summary

Implemented per PO-Redefinition Grilling 2026-07-17 (final, not reopened): NO scan. Three pieces:

1. `internal/data.RepoSlug(repoDir string) string` (new file `internal/data/repo_slug.go`) — reads `<repoDir>/.beans.yml`'s `beans.prefix` (gopkg.in/yaml.v3, minimal struct), returns `strings.TrimSuffix(strings.TrimSpace(prefix), "-")` when non-empty, else `filepath.Base(repoDir)`. Deliberately does NOT read `project.name` (plan's own reasoning: bean-IDs are coupled to `beans.prefix`, and some repos, e.g. lean-stack, omit `project.name` entirely). Never errors — missing/corrupt/unreadable `.beans.yml` all fall back silently, mirroring `config.LoadSettings`'s own contract.
2. `view_lobby.go`'s `repoPickerBody` (~line 174-189, current code): row label changed from raw `r` to `data.RepoSlug(r) + " — " + r`, passed through the existing `truncate(label, nameW)` call unchanged. Design choice (documented inline): ONE combined truncate, not slug/path budgeted separately — `ansi.Truncate` cuts from the right, so the short slug (always first) survives on narrow terminals and the long path (always last) is what gets clipped.
3. `overlay_palette.go`: new `paletteActions` entry `{actionID: "register_project", label: "register project"}`, placed directly after `repo_picker` (same repo-registry neighborhood). New `dispatchPalette` case routes to new method `m.registerProject() (tea.Model, tea.Cmd)`:
   - `m.client == nil` (Palette opened from the Lobby itself) → no-op, no crash (mirrors `edit_title`'s `focusedBean()`-nil guard shape).
   - `m.client.RepoDir` already in `m.settings.Repos` → `toastInfo` "already registered", no duplicate write.
   - otherwise → `config.SaveUserSettings(append(m.settings.Repos, repoDir), m.settings.Editor, m.settings.Theme.Accent, m.settings.Layout.TreeWidth)` (existing signature, unchanged), then `m.settings.Repos` updated in-model + `toastInfo` "Registered: " + slug. Save error → `toastError` with `err.Error()`.

No separate Lobby refresh wiring needed — `openLobby()` already reloads `config.LoadSettings()` on every open (pre-existing contract).

## Test-Output

TDD, RED→GREEN per piece (all three compiled-fail first via `command go vet`/`command go test`, confirmed undefined symbol, then implemented):

- `internal/data/repo_slug_test.go` (5 new tests: prefix present, `.beans.yml` missing, prefix empty/only `project.name` present — explicitly NOT read, prefix without trailing hyphen, corrupt YAML) — RED: `undefined: RepoSlug` (5x compile errors) → GREEN after `repo_slug.go`.
- `internal/tui/view_lobby_test.go` (2 new tests: slug+path shown from a real `beans.prefix`, dir-basename fallback with no `.beans.yml`) — RED: assertion failure (`repoPickerBody() missing "lt — <dir>"`, raw path only) → GREEN after the label change. (Rendered at content-width 300, bypassing `repoPickerWidth`'s 72-cell production cap, since macOS `t.TempDir()` paths routinely exceed that cap — width-cap/truncation behavior itself covered separately by the pre-existing `TestViewLobbyFrameMatchesWidthHeight`.)
- `internal/tui/overlay_palette_test.go` (6 new tests: action present + positioned after `repo_picker`, nil-client no-op, already-registered dedup+toast, success save+toast+persisted-to-disk, dispatch wiring) — RED: `m.registerProject undefined` (vet compile error) → GREEN after the method + dispatch case + action entry.

Two PRE-EXISTING tests needed updates for the new global action count (not scope creep — direct consequence of adding a 10th global palette item): `TestPaletteActionsNoFocusedBeanOmitsNodeActions` (9→10 global actions) and the "go" fuzzy-query test, renamed `TestPalFilteredActionsFuzzyGoMatchesAllGoToEntriesPlusRegisterProject` (5→6 matches — "register project" incidentally fuzzy-matches "go" as a rune subsequence: g in "register", o in "project" — documented as an honest widening, not a scope change, in the test's own doc-stamp).

Full-repo gates, all green:
- `command go build -o bin/bt .` — OK
- `command gofmt -l .` (excl. `beans-src`) — no output (clean)
- `command go vet ./...` — no output (clean)
- `command go test ./... -short -count=1` — all packages OK (~14s `internal/tui`)
- `command go test ./... -count=1` (full, mandatory pre-commit run) — all packages OK, `internal/tui` 149.945s (~155s expected budget)

## Smoke

tmux session `btd3ps_smoke`, isolated `HOME=<scratchpad>/bt-d3ps-smoke/home` (own `~/.config/beans-tui/config.yaml`, never touched Erik's real one — verified before/after: real `~/.config/beans-tui/config.yaml` unchanged, still lists only Erik's own pre-existing repo). Two temp repos: `repo-registered` (prefix `rr-`, pre-registered in the isolated config.yaml) and `repo-unregistered` (prefix `lean-stack-`, NOT pre-registered, cwd-opened).

1. `bt` started with cwd = `repo-unregistered` → opened directly into Browse (header "repo-unregistered: Browse", cwd-resolved, no prior registration).
2. `ctrl+k` → Command-Center shows "register project" directly after "go to repo picker".
3. Typed "register" → filtered to exactly "register project" → `enter` → toast `● Registered: lean-stack` (slug correctly resolved from `beans.prefix: lean-stack-`).
4. Verified `$HOME/.config/beans-tui/config.yaml` on disk: both repos present (`repo-registered` pre-existing + `repo-unregistered` newly appended).
5. `p` → Lobby shows both rows as `slug — path`: `rr — <repo-registered path>` and `lean-stack — <repo-unregistered path>` (path truncated at terminal width, slug fully visible — per the documented truncation-priority choice).
6. `esc` back to Browse, `ctrl+k` → "register" → enter a second time → toast `● already registered` — confirmed via `config.yaml` on disk still had exactly 2 entries (no duplicate write).
7. Quit (`q` + `enter`), tmux session killed, entire isolated smoke dir (`<scratchpad>/bt-d3ps-smoke`) removed. Real `~/.config/beans-tui/config.yaml` re-verified unchanged after teardown.

## Deviations/ERRATA

- Two pre-existing palette tests updated for the new 10th global action (see Test-Output) — direct, unavoidable consequence of the new action entry, not scope creep.
- View-lobby label test uses render width 300 (not the production `repoPickerWidth` cap) to isolate label-construction from width-truncation, since `t.TempDir()` paths on macOS routinely exceed the 72-cell production cap and would make the assertion width-flaky for reasons unrelated to what it guards. Production truncation behavior is unchanged and remains covered by the pre-existing `TestViewLobbyFrameMatchesWidthHeight`.
- No golden regen needed: none of the 3 existing golden fixtures (`chrome.golden`, `backlog.golden`, `tree.golden`) render `viewLobby` — verified by running all three explicitly (`TestChromeGolden`, `TestBacklogGolden`, `TestTreeGolden`), all pass unchanged against the existing fixtures.
- No line-number drift found against the plan's `overlay_palette.go`/`view_lobby.go` citations at implementation time.
