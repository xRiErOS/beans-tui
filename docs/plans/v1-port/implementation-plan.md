# beans-tui v1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development
> (recommended) or superpowers:executing-plans to implement this plan task-by-task.
> Steps use checkbox (`- [ ]`) syntax for tracking.
> **Subagent-Modell-Regel:** alle Implementierungs-Subagents mit `model: sonnet`
> (Opus nur in begründeten Ausnahmen). Fable = Supervisor.

**Goal:** Eigenständige TUI `bt` (Go/bubbletea), die Look & Funktionalität der devd-TUI auf
beans-Repos portiert — PO-Cockpit: browsen, bearbeiten, Review-Flow mit Agenten fahren.

**Architecture:** Elm-Architektur + Overlay-Compositing 1:1 aus devd-TUI
(`~/Obsidian/tools/DeveloperDashboard/apps/cli-go`, Eriks Code — Port erlaubt).
Datenlayer NEU: beans-CLI-Subprocess (`--json`, `--if-match`, `-S`) + In-Memory-Index +
eigener fsnotify-Watcher (D02). Quelle der Wahrheit: `design-spec.md` daneben.

**Tech Stack:** Go 1.26 · bubbletea v1.3.10 · bubbles v1.0 · lipgloss v1.1.1-0.20250404203927 ·
huh v1.0 · glamour v1.0 · charmbracelet/x/ansi · muesli/termenv · fsnotify v1.9 · cobra ·
go-osc52/v2 · atotto/clipboard · gopkg.in/yaml.v3

**Konventionen (verbindlich):**
- Build/Test IMMER `command go …` (lokales go-Shadowing wie bei devd-cli meiden).
- Datei-Namen `internal/tui/`: `<art>_<verb>_<entität>.go`; Theme-Token nur aus `internal/theme`.
- Commits: Conventional Commits, `Refs: <bean-id>` Footer, KEIN Co-Authored-By (tools-Standard).
- Nach jedem Task: Tests grün + Commit. Nach jedem Epos: Epos-Abschluss-Ritual (unten).

---

## Epos-Rituale (gilt für E1–E6)

**Epos-Start (E2–E6):** (1) `design-spec.md` + dieses Dokument + `docs/SSTD.md` lesen;
(2) voll-granularen Epos-Plan schreiben nach `docs/plans/v1-port/epic-EN-plan.md`
(gleiche TDD-Struktur wie E1 unten — Files/Steps/Tests/Commits, kein Placeholder);
(3) zugehöriges Epic-bean auf `in-progress` setzen (`beans update <id> -s in-progress`).

**Epos-Abschluss:** (1) `command go test ./...` grün + `command go build -o bin/bt .` ok;
(2) beans pflegen: erledigte Task-beans `completed` setzen (Implementierungs-Tasks sind
agent-abschließbar; NUR Epic-/Milestone-beans bekommen Tag `to-review` statt completed —
PO-Gate); (3) committen; (4) SSTD-Pointer aktualisieren falls nötig;
(5) via Skill `ce-nsp-auto` Handover-Prompt für Folge-Agent erzeugen (frischer Kontext).

---

## Epos E1 — Foundation (voll granular)

Liefert: lauffähiges `bt` mit Datenlayer (Read+Write+Watch), Theme, Chrome-Primitiven,
App-Shell und read-only Tree im eigenen Repo (Dogfooding).

### Task 1: Modul-Skeleton + cobra + Makefile

**Files:**
- Create: `go.mod` (Modul `beans-tui`), `main.go`, `cmd/root.go`, `cmd/tui.go`, `Makefile`
- Test: `cmd/root_test.go`

- [ ] **Step 1:** `command go mod init beans-tui && command go get github.com/spf13/cobra@v1.10.2`
- [ ] **Step 2: Failing test** — `cmd/root_test.go`:

```go
package cmd

import "testing"

func TestRootStartsTUIByDefault(t *testing.T) {
	cmd := NewRootCmd()
	if cmd.Use != "bt [repo-pfad]" {
		t.Fatalf("Use = %q", cmd.Use)
	}
	// tui-Subcommand vorhanden
	sub, _, err := cmd.Find([]string{"tui"})
	if err != nil || sub.Name() != "tui" {
		t.Fatalf("tui subcommand fehlt: %v", err)
	}
}
```

- [ ] **Step 3:** Run `command go test ./cmd/` → FAIL (NewRootCmd undefined)
- [ ] **Step 4: Implement** — `cmd/root.go`: `NewRootCmd()` (Use `bt [repo-pfad]`,
  RunE startet TUI mit optionalem Pfad-Arg), `Execute()`; `cmd/tui.go`: explizites
  `tui`-Subcommand (gleicher RunE); `main.go` ruft `cmd.Execute()`. TUI-Start vorerst
  Stub `runTUI(path string) error { return nil }` in `cmd/tui.go`.
- [ ] **Step 5:** `command go test ./cmd/` → PASS
- [ ] **Step 6:** `Makefile`: `build: ; command go build -o bin/bt .` · `test: ; command go test ./...` · `clean: ; rm -f bin/bt`. Verify `make build` erzeugt `bin/bt`.
- [ ] **Step 7:** Commit `feat(cmd): bt-Skeleton mit cobra (root + tui)`

### Task 2: Datenlayer — Repo-Discovery + Read (`internal/data`)

**Files:**
- Create: `internal/data/bean.go` (Typ), `internal/data/client.go` (CLI-Wrapper), `internal/data/discover.go`
- Test: `internal/data/client_test.go`, `internal/data/discover_test.go`

**Bean-Typ (Vertrag, überall wiederverwendet):**

```go
package data

import "time"

type Bean struct {
	ID        string     `json:"id"`
	Slug      string     `json:"slug"`
	Path      string     `json:"path"`
	Title     string     `json:"title"`
	Status    string     `json:"status"`   // draft|todo|in-progress|completed|scrapped
	Type      string     `json:"type"`     // milestone|epic|feature|task|bug
	Priority  string     `json:"priority"` // critical|high|normal|low|deferred
	Tags      []string   `json:"tags"`
	Parent    string     `json:"parent"`
	Blocking  []string   `json:"blocking"`
	BlockedBy []string   `json:"blocked_by"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
	Body      string     `json:"body"` // nur bei --full
	ETag      string     `json:"etag"`
}
```

- [ ] **Step 1: Failing tests** — Fixture-Helper `newTestRepo(t)` erzeugt tmp-Dir mit
  `.beans.yml` (prefix `tt-`) + 3 Fixture-Beans (milestone/epic/task, task mit
  `parent: <epic-id>`) als echte `.md`-Dateien; Guard `requireBeansBinary(t)`
  (skip wenn `beans` nicht im PATH). Tests: `TestDiscoverFindsConfigUpward` (Discovery aus
  Unterverzeichnis), `TestListReturnsAllBeansWithBody` (Client.List liefert 3 beans, Body
  gefüllt, ETag nicht leer).
- [ ] **Step 2:** `command go test ./internal/data/` → FAIL
- [ ] **Step 3: Implement** — `discover.go`: `FindRepo(start string) (string, error)`
  (aufwärts bis `.beans.yml`); `client.go`:

```go
type Client struct{ RepoDir string }

func (c *Client) List() ([]Bean, error) {
	out, err := c.run("list", "--json", "--full")
	if err != nil { return nil, err }
	var beans []Bean
	return beans, json.Unmarshal(out, &beans)
}

func (c *Client) run(args ...string) ([]byte, error) {
	cmd := exec.Command("beans", args...)
	cmd.Dir = c.RepoDir
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("beans %s: %w: %s", args[0], err, strings.TrimSpace(stderr.String()))
	}
	return stdout.Bytes(), nil
}
```

- [ ] **Step 4:** Tests → PASS. **Hinweis:** JSON-Feldnamen gegen echten Output verifizieren
  (`beans list --json --full | head`), Typ ggf. anpassen (z.B. fehlendes `slug`).
- [ ] **Step 5:** Commit `feat(data): CLI-Client mit Repo-Discovery und List`

### Task 3: Datenlayer — Index + Tree-Ableitung

**Files:**
- Create: `internal/data/index.go`
- Test: `internal/data/index_test.go`

Index über geladene beans: `ByID map[string]*Bean`, `Children map[string][]*Bean`
(Parent invertiert, sortiert Status→Priority→Type→Titel wie beans upstream),
`Roots(types ...string)` (Milestones + parentlose Epics/Rest für Tree V2),
`Backlog()` (parentlose non-milestone/epic, status todo/draft), `WithTag(tag string)`.

- [ ] **Step 1: Failing tests** — reine Unit-Tests ohne Binary (Fixture-Slice):
  `TestIndexChildrenSorted`, `TestRootsMilestonesFirst`, `TestBacklogExcludesParented`,
  `TestWithTagToReview`.
- [ ] **Step 2:** FAIL → **Step 3: Implement** → **Step 4:** PASS
- [ ] **Step 5:** Commit `feat(data): In-Memory-Index mit Tree/Backlog/Tag-Ableitungen`

### Task 4: Datenlayer — Mutationen mit `--if-match`

**Files:**
- Modify: `internal/data/client.go`
- Test: `internal/data/client_mut_test.go`

API: `Create(opts CreateOpts) (Bean, error)` · `SetStatus(id, status, etag)` ·
`SetPriority/SetType/SetTitle` · `AddTag/RemoveTag` · `SetParent/RemoveParent` ·
`AddBlockedBy/RemoveBlockedBy` · `AppendBody(id, text, etag)` · `Delete(id)`.
Alle Updates über `beans update <id> … --if-match <etag> --json`; Create über
`beans create <titel> -t … --json`; typed error `ErrConflict` bei ETag-Mismatch
(stderr-Match auf if-match/conflict).

- [ ] **Step 1: Failing tests** gegen `newTestRepo(t)`: `TestSetStatusRoundtrip`,
  `TestCreateReturnsNewBean`, `TestConflictOnStaleETag` (bean extern via zweitem
  `beans update` ändern → alter ETag → ErrConflict), `TestAppendBodyAddsSection`.
- [ ] **Step 2:** FAIL → **Step 3: Implement** → **Step 4:** PASS
- [ ] **Step 5:** Commit `feat(data): Mutationen mit optimistic locking (--if-match)`

### Task 5: Datenlayer — fsnotify-Watcher

**Files:**
- Create: `internal/data/watcher.go`
- Test: `internal/data/watcher_test.go`

`Watch(repoDir string, onChange func()) (stop func(), err error)` — fsnotify auf
`.beans/` + `.beans/archive/` (falls existiert), Debounce 150 ms (Timer-Reset),
reagiert auf Create/Write/Remove/Rename.

- [ ] **Step 1: Failing test** — `TestWatcherFiresOnceForBurst`: 3 schnelle Datei-Writes
  → genau 1 onChange innerhalb 1 s (channel + Zähler).
- [ ] **Step 2:** FAIL → **Step 3: Implement** (`command go get github.com/fsnotify/fsnotify@v1.9.0`) → **Step 4:** PASS
- [ ] **Step 5:** Commit `feat(data): Debounced fsnotify-Watcher für .beans/`

### Task 6: Theme-Port (`internal/theme`)

**Files:**
- Create: `internal/theme/theme.go`, `internal/theme/icons.go`
- Test: `internal/theme/theme_test.go`
- Quelle: devd `internal/theme/theme.go` (Catppuccin-Token 1:1 übernehmen)

Port-Anpassungen (einzige Abweichungen von der Quelle):
(1) Env `DEVD_ASCII_ICONS` → `BT_ASCII_ICONS`;
(2) Status-Mapping beans: draft→Blue `#8aadf4` · todo→Text `#cad3f5` · in-progress→Yellow
`#eed49f` · completed→Green `#a6da95` · scrapped→Red `#ed8796`, Glyph `◉` (ASCII `*`);
(3) Type-Icons: milestone `⬢` Peach · epic `✦` Mauve · feature `✦` Green · task `⯅` Blue ·
bug `⯁` Red (ASCII `#`/`*`/`+`/`^`/`!`);
(4) Priority: critical/high→Red bold · normal→Text · low/deferred→Green/Hint.

- [ ] **Step 1: Failing tests** — `TestStatusColorMapping` (alle 5 beans-Status → erwartete
  Hex), `TestAsciiFallback` (Env gesetzt → `*`), `TestTypeIconAllTypes`.
- [ ] **Step 2:** FAIL → **Step 3: Port+Anpassen** → **Step 4:** PASS
- [ ] **Step 5:** Commit `feat(theme): Catppuccin-Macchiato-Port mit beans-Enum-Mapping`

### Task 7: Chrome-Primitive-Port (`internal/tui` Infrastruktur)

**Files:**
- Create: `internal/tui/view.go` (chrome/breadcrumb/footer/masterDetailWidths),
  `internal/tui/render_shared.go`, `internal/tui/modal.go`, `internal/tui/overlay.go`,
  `internal/tui/list.go` (listState), `internal/tui/keymap.go`
- Test: `internal/tui/chrome_test.go` (golden), `internal/tui/keymap_test.go`
- Quelle: gleichnamige devd-Dateien; Breadcrumb-Format `> repo: Titel`, Keymap nach
  design-spec §7 (devd-Belegung, `B` für Blocking-Picker, `p` Repo-Picker)

- [ ] **Step 1: Failing golden test** — `TestChromeGolden`: 80×24-Frame mit Header/Body/
  Footer gegen `testdata/chrome.golden` (Update-Flag `-update`); `TestKeymapNoCtrlSQ`
  (Guard: keine ctrl+s/ctrl+q-Bindings).
- [ ] **Step 2:** FAIL → **Step 3: Port** (API-Client-Referenzen entfernen, theme-Import
  auf `beans-tui/internal/theme`) → **Step 4:** PASS (golden mit `-update` erzeugen,
  dann Re-Run ohne Flag)
- [ ] **Step 5:** Commit `feat(tui): Chrome-Primitive (view/modal/overlay/keymap) portiert`

### Task 8: App-Shell + read-only Tree (V2-Basis)

**Files:**
- Create: `internal/tui/types.go` (model, viewID-Enum), `internal/tui/app.go` (Init/Program),
  `internal/tui/update.go`, `internal/tui/messages.go` (loadedMsg/errorMsg + Cmd-Producer),
  `internal/tui/view_browse_repo.go` (Tree links, Platzhalter-Detail rechts),
  `internal/tui/box_confirm_quit.go`
- Modify: `cmd/tui.go` (runTUI: FindRepo → Client → tea.Program mit AltScreen+Mouse)
- Test: `internal/tui/update_test.go`, `internal/tui/tree_golden_test.go`

Verhalten: Start lädt beans async (spinnerlos ok), Tree zeigt Milestones→Epics→Tasks
(expand `→/l`, collapse `←/j`, Cursor `↑↓/i k`, Status-Glyph+Farbe+ID+Titel je Zeile),
`q` → Quit-Confirm, `ctrl+r` Reload, Watcher-Event → Reload mit Cursor-Restore (per ID).

- [ ] **Step 1: Failing update tests** — `TestCursorMovesAndExpands` (KeyMsg-Sequenz gegen
  Fixture-Index), `TestQuitConfirm` (q → confirm sichtbar, esc → weg),
  `TestReloadKeepsCursorOnID`.
- [ ] **Step 2:** FAIL → **Step 3: Implement** (Port-Muster `view_browse_project.go`,
  Datenpfad: `data.Client`+`data.Index` statt HTTP) → **Step 4:** PASS
- [ ] **Step 4b: Golden** — `TestTreeGolden` 100×30 gegen `testdata/tree.golden`.
- [ ] **Step 5:** `make build && ./bin/bt` im beans-tui-Repo — manueller Smoke: eigene
  Entwicklungs-beans sichtbar (Dogfooding). Screenshot/Terminal-Dump als Beleg in Commit-Body.
- [ ] **Step 6:** Commit `feat(tui): App-Shell + read-only Tree gegen beans-Datenlayer`

### Task 9: E1-Abschluss

- [ ] `command go test ./...` grün, `make build` ok, README.md (Kurz: Zweck, Install, Start)
- [ ] beans pflegen (Task-beans completed; E1-Epic Tag `to-review`)
- [ ] Commit `docs: README + E1-Abschluss` → Epos-Abschluss-Ritual (ce-nsp-auto)

---

## Epos E2 — Browse & Detail (Detail-Plan bei Epos-Start)

Tasks (je Task TDD wie E1; Detail-Plan `epic-E2-plan.md` beim Start schreiben):
1. **Detail-Accordion-Port** (`accordion.go` + `view_detail_bean.go`): Sections Meta/Body
   (glamour)/Beziehungen/Historie; Ziffern `1…9`; Zwei-Ebenen-Fokus. Quelle: devd
   `accordion.go`, `view_detail_issue.go`.
2. **Master-Detail-Verdrahtung**: Tree-Cursor treibt Detail-Pane; Fokus-Tausch `tab`/`→`
   mit Border-Farbwechsel (D03-Muster); Beziehungs-Navigation (enter auf Parent/Child/
   Blocking springt).
3. **Suche `/`**: Live-Filter im Tree-Kopf (lokal auf Index) + `-S`-Bleve-Modus ab 3 Zeichen
   (async); esc-Kaskade (Suche→Filter→Lobby-Äquivalent) wie devd.
4. **Facetten-Filter `f`** + `X` Reset: Checkbox-Menü Status/Type/Priority/Tag
   (Port `view_browse_backlog.go:89-136`-Muster), wirkt auf Tree UND Backlog.
5. **Backlog-View V3** (`b`): parentlose+ready beans, Sort-Toggle `S`
   (status/prio/created/updated), Master-Detail. Quelle: `view_browse_backlog.go`.
6. E2-Abschluss: Tests+build, beans-Pflege, Golden-Suite erweitert, ce-nsp-auto.

## Epos E3 — Mutationen (Detail-Plan bei Epos-Start)

1. **Status-/Type-/Prio-Menüs** (`s`/`t`-Kontext): schwebende Menüs, nur beans-Enums
   (Port `types.go:416-473`-Transitions-Muster → hier statuslos-frei, aber Enum-constrained).
2. **Tag-Picker** (`t`) mit Nutzungszählern + Freitext-Neuanlage (validiert
   `^[a-z][a-z0-9]*(-[a-z0-9]+)*$`).
3. **Parent-/Blocking-Picker** (`a`/`B`): Auswahl-Liste mit Typ-Hierarchie-Filter +
   Zyklen-Ausschluss (Nachkommen ausblenden — Port parentpicker-Muster aus beans-upstream-TUI-Analyse).
4. **Create-Form** (`c`, huh): Titel/Typ/Prio/Status/Parent/Tags/Body + Confirm-Gate;
   danach Cursor auf neuem Bean. Quelle: `forms_shared.go`, `form_create_*.go`, `box_confirm_create.go`.
5. **Edit-Form** (`e`) + `ctrl+e` Body im `$EDITOR` (tea.ExecProcess-Suspend, Port `editor.go`).
6. **Delete-Confirm** (`d`) mit Kinder-Count-Preview; ETag-Konflikt-Handling überall
   (ErrConflict → Toast + Reload).
7. E3-Abschluss-Ritual.

## Epos E4 — Command-Center & Review-Cockpit (Detail-Plan bei Epos-Start)

1. **Command-Center** (`ctrl+k`/`K`): fuzzy Palette (Port `overlay_palette.go`) — Aktionen
   (Create/Views/Status…) + Bean-Treffer gemischt, kontextabhängige Einträge zuerst.
2. **Review-Queue-View** (`R`): beans mit Tag `to-review`, gruppiert nach Epic
   (Port `view_navigate_reviews.go`).
3. **Review-Cockpit**: Master-Detail mit Verdikt-Dots, Summary „x of n", `a` pass
   (SetStatus completed + RemoveTag), `x` reject (huh-Kommentar-Form → AppendBody
   `## Review <YYYY-MM-DD>` + Tag-Swap to-review→rework), `o` reopen, `n/p` next/prev.
   Quelle: `view_review_sprint.go`, `keys_review.go`.
4. **Yank Review-Stand** (`y`): Markdown-Zusammenfassung ins Clipboard.
5. E4-Abschluss-Ritual.

## Epos E5 — Polish (Detail-Plan bei Epos-Start)

1. **Toast-System** (Port `overlay_show_toast.go`) + Fehler-Toasts aus allen async Cmds.
2. **Help-Overlay** `?` aus zentraler Keymap generiert (Port `overlay_shortcuts.go`).
3. **Yank Bean/Epic-Kontext** (`y`, OSC52 + nativ, Port `internal/clip` + `context.go`).
4. **Maus**: Wheel-Scroll, Klick setzt Cursor, Doppelklick collapsed (Port `update.go`-Mausteil).
5. **Settings** `~/.config/beans-tui/config.yaml` (repos/editor/accent/treewidth) +
   Runtime-State `state.json` (LastRepo) — Port `internal/config`.
6. **Lobby V1 + Repo-Picker `p`**: ASCII-Logo `beans`, Repo-Liste aus Config, Live-Filter;
   Direkt-Start bei cwd-Repo. Quelle: `view_home.go`.
7. **Archiv-Sicht**: completed/scrapped ein-/ausblenden (Filter-Default: aus).
8. E5-Abschluss-Ritual.

## Epos E6 — Validierung & Release (Detail-Plan bei Epos-Start)

1. Je User Story US-01…US-14 (design-spec §10): Validierungs-Schritt ausführen
   (Test zeigen oder tmux-Session-Beleg), Ergebnis in `docs/plans/v1-port/validation.md`
   dokumentieren (Tabelle US | Nachweis | Status).
2. Offene Lücken fixen (Bugs aus Validierung als beans erfassen, abarbeiten).
3. README vervollständigen (Install `command go install .`, Keymap-Referenz aus
   `docs/shortcuts.md` generiert).
4. Demo-Nachweis: `bt` im lean-stack-Repo gestartet (US-01/US-14-Beleg).
5. Milestone-bean Tag `to-review` (PO-Abnahme), lean-stack-Verweis-bean aktualisieren,
   KC-Konzept `po-immersion-beans-via-obsidian-bases-no-custom-tui` via `/okf` um
   Supersede-Hinweis ergänzen (D07).

---

## Self-Review (Plan gegen Spec)

- Spec-Coverage: V1–V8 → E1(V2-Basis)/E2(V2,V3,V4)/E3(V7,V8-Teile)/E4(V5,V6)/E5(V1,V8-Rest,
  Settings) ✓ · US-01…14 → E-Zuordnung in design-spec §12 ✓ · Review-Konvention §5 → E4.3 ✓ ·
  Live-Reload → E1.5+E1.8 ✓ · ETag → E1.4+E3.6 ✓ · Archiv → E5.7 ✓
- Placeholder: E2–E6 sind bewusst Roadmap-Ebene mit Detail-Plan-Pflicht bei Epos-Start
  (KC-Konvention aus devd-TUI-Plan: „Jede Folge-Phase bekommt bei Start ihren eigenen
  voll-granularen Plan") — kein versteckter TBD in E1.
- Typ-Konsistenz: `data.Bean`/`data.Client`/`data.Index` in T2–T5 und E2–E4 einheitlich ✓
