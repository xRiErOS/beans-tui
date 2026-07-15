# Epos E5 — Polish (voll granular)

Liefert: Toast-System (mit sticky Konflikt-Meldung, E3-Übernahme-Pflicht),
Help-Overlay `?`, Yank `y` (Bean-/Epic-Kontext + Review-Stand, OSC52+nativ),
Maus (Wheel/Klick/Doppelklick), Settings (`~/.config/beans-tui/config.yaml`
+ `state.json`), Lobby V1 + Repo-Picker `p` (inkl. Watcher-Lifecycle-Switch),
Archiv-Sicht (completed/scrapped default-aus, togglebar). Quelle:
`design-spec.md` §6 V1/V8, §7, §9, `implementation-plan.md` »Epos E5«, bean
`bt-5h4d` Body (Toast-Konflikt-Pflicht aus E3, E3-I01).

Referenz-Quellen (devd, `~/Obsidian/tools/DeveloperDashboard/apps/cli-go`):
`internal/tui/overlay_show_toast.go`, `overlay_shortcuts.go`, `context.go`,
`internal/clip/clip.go`, `view_home.go`, `internal/config/{settings,state}.go`,
`update.go` (Mausteil, Zeilen 459-598/880-950).

---

## Design-Entscheidungen (vor Task 1 festgezurrt)

### a) Toast-Architektur

**Ein Slot** (kein Stack, `toastState`), Port devd `overlay_show_toast.go`
strukturell VERBATIM (`toastKind`/`toastState`/`toastTarget`/
`toastDebounceWindow` 300ms/`toastDuration` Error=8s·Warn=3s·Info=5s/
`showToast`/`clearToastUnlessSticky`/`toastBox`/`toastGeometry`/
`renderToast`/`toastHit`/`dismissToast`). Zwei Deviationen ggü. devd, beide
notwendig weil beans-tui's Chrome-Architektur anders liegt (kein einzelnes
`m.termWidth()`, keine `viewComposite()`-Zwischenschicht, siehe unten):

1. **Geometrie:** devd's `toastGeometry`/`renderToast` rechnen gegen
   `m.width`/`m.height` (den fertig kompositierten Frame INKL. äußerem
   Rahmen, B01-Fix-Kommentar in der Quelle). beans-tui's drei View-Funktionen
   (`viewBrowseRepo`/`viewBacklog`/`viewReviewCockpit`) bauen JEDE ihren
   eigenen `w, h := m.width, m.height`-Frame inkl. `outerBorder(...,true)` —
   identisches Prinzip, also 1:1 gegen `m.width`/`m.height` kompositierbar,
   keine Methode nötig.
2. **Reihenfolge/Ort:** devd trennt `View()` (nur `renderToast(viewComposite())`)
   von `viewComposite()` (alle modalen Overlays via `placeOverlay`, `confirmQuit`
   topmost). beans-tui hat KEINE zentrale `viewComposite`-Funktion — jede der
   drei View-Funktionen endet selbst mit `return m.composeOverlays(out, w, h)`
   (`composeOverlays` in `view_browse_repo.go:582`, von allen drei Views
   aufgerufen). **Entscheidung:** der TOP-LEVEL `View()`-Dispatcher
   (`view_browse_repo.go:618`) wird umgebaut, um das Ergebnis JEDER
   Sub-View-Funktion mit `m.renderToast(out)` zu wrappen — NICHT
   `composeOverlays` selbst (Toast schwebt über composeOverlays' GESAMTEM
   Stack inkl. `confirmQuit`, exakt devd's Anspruch "über ALLEM").

**Sticky-Regel (E3-Übernahme, PFLICHT, bean-Body bt-5h4d):** `data.ErrConflict`
setzt `sticky=true` in `applyMutationResult`. Da `m.toast` NUR von
`showToast`/`dismissToast`/`handleToastExpired` beschrieben wird (Grep-Audit
als Teil von Task 1 Step 4: kein anderer Call-Site darf `m.toast = ...`
setzen), übersteht ein sticky Toast jeden Reload automatisch — `applyLoaded`/
`watchMsg`-Handling fasst `m.toast` nie an. Kein `clearToastUnlessSticky`-
Aufruf an einer Reload-Stelle nötig (anders als devd, wo mehrere Reload-Pfade
das explizit brauchten) — hier reicht die Abwesenheit einer Schreiboperation.

**Dual-Write mit `m.err` (Deviation, begründet):** devd hat den alten
Footer-Status komplett durch Toast ABGELÖST (DD2-272-Kommentar: "Löst den
alten Footer-Status-Toast ab"). beans-tui macht das NICHT — `m.err` bleibt
UNVERÄNDERT bestehen (Chrome-Statusbar Red-Slot, `view_browse_repo.go`s
eigener `status := statusBar(indicator, m.err, innerW)`), weil >20 bestehende
Tests aus E1-E4 (u.a. `etag_conflict_test.go`, jeder Mutation-Test) direkt
gegen `m.err`-Stringinhalte assertieren. Ein Ersatz wäre ein Multi-Epic-
Refactor außerhalb E5-Scope. **Jede** bestehende `m.err = ...`-Stelle bekommt
stattdessen ZUSÄTZLICH einen `showToast(...)`-Aufruf gleicher Schwere
(Kind error für harte Fehler, warn für Hinweise wie "Bean nicht mehr
vorhanden"/`createInFlightNote`). Nur der ErrConflict-Zweig ist sticky, alle
anderen non-sticky (Standard-Dismiss-Timer).

**Klick-Dismiss vs. Maus (Reihenfolge!):** Toast (Task 1) MUSS vor Maus
(Task 4) landen — `handleMouse`s allererste Prüfung ist der Toast-Hit-Test
(Port devd `update.go:466`), und Update()s Dispatch-Reihenfolge (Toast-Klick
VOR einem offenen-Formular-Kurzschluss, Port devd's Cross-Feature-Fix
DD2-272/273) wird in Task 1 bereits vorbereitet (Stub-Case), in Task 4 fertig
verdrahtet. Deckt sich mit der im Handover geforderten Abarbeitungsreihenfolge.

### b) Yank-Kontext-Format

devd trennt `milestoneClip`/`sprintClip`/`backlogClip` (drei Entitäten, drei
Funktionen), weil devd eine starre Milestone→Sprint→Issue-Dreiteilung hat.
beans-tui kennt nur EINE Entität (`data.Bean`, `type` unterscheidet
milestone/epic/feature/task/bug, design-spec §4) — **Entscheidung:** EIN
view-agnostischer `beanContext(idx *data.Index, b *data.Bean) string`
(`internal/tui/context.go`, NEU):

```
# <ID> — <Titel>
- Status: <status>
- Type: <type>
- Priority: <priority>
- Tags: <tag1, tag2>              (nur wenn vorhanden)
- Parent: <parentID> — <parentTitel>   (nur wenn vorhanden)
- Blocking: <id> — <titel>, ...   (nur wenn vorhanden)
- Blocked by: <id> — <titel>, ... (nur wenn vorhanden)

<Body, unverändert Markdown>       (nur wenn nicht leer)

## Children (<n>)                  (nur wenn idx.Children[b.ID] nicht leer)
| ID | Type | Status | Prio | Title |
|----|------|--------|------|-------|
...
```

Deckt sowohl "Issue" (Blätter: nur Header+Body) als auch "Epic/Milestone"
(Header+Children-Tabelle, Children-Zeilen via bestehendem `sortBeans`) in
EINER Funktion ab — kein Parallelcode wie devd. `y` wirkt IMMER auf
`m.focusedBean()` (identisch in Tree UND Backlog, gleicher Resolver wie
`keyNodeAction`, design-spec §7 macht hier KEINEN Unterschied zwischen den
Views). **Einzige View-lokale Ausnahme** (design-spec §7 wörtlich: "Review
... y Review-Stand yanken"): Review-Cockpit überschreibt `y` mit
`reviewStandMarkdown(idx)` — Kopf + je `reviewGroup` (aus E4 T3,
`view_review_cockpit.go:64` `reviewQueue`) eine Tabelle der to-review-Beans,
danach eine Rework-Sektion (`reviewRework`). Kein neuer Gruppierungscode —
reine Serialisierung der E4-Datenstrukturen.

**Erfolgsbestätigung (US-11 PFLICHT: "Toast bestätigt"):** jeder `y`-Erfolg
feuert `showToast(toastInfo, "Kopiert: <ID/Kontext>", "", nil, false)`.

### c) Settings-Schema + Editor-Präzedenz

Schema (NUR was design-spec §9/Handover nennen — `repos`/`editor`/`accent`/
`treewidth`, KEIN `modal_width`/`start_project`/`keybindings` wie devd, YAGNI):

```yaml
repos:
  - /Users/erik/Obsidian/tools/beans-tui/beans-tui-repository
  - /Users/erik/Obsidian/tools/lean-stack
editor: "code -w"
theme:
  accent: "#f5a97f"
layout:
  tree_width: 32
```

EIN Verzeichnis `~/.config/beans-tui/` für `config.yaml` UND `state.json`
(Deviation von devd, das `~/.config/devd-cli/` vs. `~/.config/dd/` SPLITTET —
eine unbegründete historische Inkonsistenz in devd selbst, hier bewusst
konsolidiert). KEIN lokaler CWD-Override (devd's zweite Config-Schicht) —
YAGNI, ein Nutzer mit mehreren beans-Repos braucht keinen Pfad-lokalen
Override, `repos:` deckt das ab.

**Editor-Präzedenz (PFLICHT lt. Handover — "muss sich mit editorBinary()
vertragen"):** `Settings.Editor` (wenn gesetzt) **>** `$VISUAL` **>**
`$EDITOR` **>** `"vi"`-Fallback. Konkret: `editor.go` bekommt eine neue
package-level `var configuredEditor string` (Default `""`, Port-Mechanik
identisch zu devd's `configuredEditor`-Var, ABER ohne devd's hartkodierten
`"nvim"`-Default — leer bleibt leer). `editorBinary()` prüft
`configuredEditor` ZUERST (`strings.TrimSpace(configuredEditor) != ""` →
`strings.Fields(configuredEditor)`), erst danach die BESTEHENDE E3-Kaskade
unverändert. Da `configuredEditor` nur in `app.go Run()`/`cmd/tui.go` (nach
`LoadSettings()`) gesetzt wird und in JEDEM bestehenden Test unangetastet
`""` bleibt, bleibt `TestEditorBinaryResolvesVisualThenEditorThenVi`
(E3, `editor_test.go`) **unverändert grün** — reine Additiv-Erweiterung,
kein Umbau.

### d) Lobby-Trigger

Lobby (`viewLobby`, neuer 4. `viewID`-Case — KEIN Overlay-Bool: ein
Repo-Wechsel braucht einen kompletten Client-/Watcher-Neustart, das ist kein
schwebendes Modal, sondern ein Elm-typischer View-Wechsel, Port-Konvention
aus `types.go`s eigenem Doc-Kommentar "later epics add ... as siblings").

**Start-Trigger** (`cmd/tui.go RunE`, design-spec §3.2 wörtlich: "Ohne
Config oder mit cwd-Treffer: direkt ins Repo — Lobby übersprungen"):

1. `bt <pfad>`-Argument gegeben → direkt dieses Repo (Lobby übersprungen,
   US-01 unverändert).
2. Kein Argument, `data.FindRepo(cwd)` erfolgreich → direkt dieses Repo
   (Lobby übersprungen — bestehendes E1-E4-Verhalten, KEINE Regression).
3. Kein Argument, `FindRepo` schlägt fehl, `Settings.Repos` hat **≥2**
   Einträge → Lobby.
4. Kein Argument, `FindRepo` schlägt fehl, `Settings.Repos` hat **0 oder 1**
   Eintrag → bestehende Fehlermeldung (0) bzw. dieses eine Repo direkt (1) —
   Lobby für einen einzigen konfigurierten Repo wäre reine Reibung.

**Laufzeit-Trigger:** `p` (`keys.Picker`, seit E1 Task 7 im Keymap, Drift-
Guard bereits grün) öffnet `viewLobby` von ÜBERALL (design decision h-
Precedent wie `ctrl+k`/`?`) — lädt `Settings.Repos` frisch (falls seit
Programmstart per Hand editiert).

### e) Archiv-Daten-Pfad (EMPIRISCH geklärt)

**Befund (verifiziert gegen `beans-src/pkg/beancore/core.go:119-166` +
`internal/commands/archive.go` + Live-Test in diesem Repo):**
`Core.loadFromDisk()` (das `beans list` letztlich bedient) läuft
`filepath.WalkDir(c.root, ...)` über den **gesamten** `.beans`-Baum
rekursiv — übersprungen werden NUR Punkt-Präfix-Verzeichnisse
(`.worktrees/`, `.conversations/`). `archive/` ist KEIN Punkt-Verzeichnis
→ wird mitgelesen. `internal/commands/archive.go`s eigene Doku bestätigt
das wörtlich: *"Archived beans are preserved for project memory and remain
visible in all queries."* Live-Beleg in DIESEM Repo: `beans list --json`
liefert bereits heute alle 27 `completed`-Beans ungefiltert (verifiziert per
Python-Zählung über den echten JSON-Output) — `data.Index` hat AKTUELL
keinerlei Status-Filterung (`Roots`/`Children` in `index.go` filtern nicht
nach Status).

**Konsequenz:** `data.Client.List()` (→ `beans list --json --full`) braucht
**keine** Code-Änderung, um archivierte Beans zu lesen — das tut es
bereits. `Bean.Path` trägt für archivierte Beans das Präfix `"archive/"`
(`Core.isArchivedPath`, `core.go:826`) — billig auswertbar ohne neue
CLI-Anfrage. **Die eigentliche Task-7-Arbeit ist der UMGEKEHRTE Fall:** ein
NEUER Default-Filter versteckt `completed`/`scrapped` (archiviert ODER
nicht — die Unterscheidung "archiviert" vs. "completed-aber-noch-in-.beans/"
ist für die Sichtbarkeits-Policy irrelevant, beide sind "fertig, aus dem
Weg"), togglebar sichtbar via der bestehenden `f`-Facetten-Menü-Infrastruktur
(`box_filter_facets.go`, EINE neue, eigenständige Menüzeile "Archivierte
einblenden", KEIN neuer Key — `keys.Toggle`, space/x, bereits vorhanden).
`beanMatches` (die EINE geteilte Tree+Backlog-Prädikat-Funktion,
`box_filter_facets.go:155`) bekommt ein drittes AND-Glied. Tree-Ancestor-
Sichtbarkeit (ein archiviertes Milestone OHNE sichtbare Nachkommen
verschwindet automatisch) funktioniert BEREITS generisch über
`flattenTreeFiltered`s bestehende `subtreeHasMatch`-Logik
(`view_browse_repo.go:271ff`) — keine Sonderbehandlung nötig.

### f) Maus-Mapping

Kein `m.scroll`-Feld existiert (`Chrome()`/`ChromeOpts.Scroll` in `view.go`
ist TOTER Code seit T8 — nur `chrome_test.go` konsumiert ihn, keine der drei
echten View-Funktionen ruft `Chrome()` auf, sie bauen ihr eigenes Layout
inline, verifiziert per Grep). **Entscheidung:** Wheel bewegt den
View-eigenen CURSOR (nicht einen Scroll-Offset) — `windowAround`/
`windowStart` (`view_browse_repo.go:416-435`, bereits seit T8 vorhanden,
"devd windowAround/windowStart port ... itself shared there by render + the
mouse-click Y->row mapping", exakt das T8-Opus-Review-Note aus dem Handover)
folgt dem Cursor automatisch beim Rendern — kein Doppelmechanismus.

**Click-Y→Row-Geometrie** (Tree, `viewBrowseRepo`): eine neue
`treeClickRow(m model, nodes []treeNode, msg tea.MouseMsg) (idx int, ok
bool)` MUSS `lw, rw := masterDetailWidths(innerW, 24)` und `bodyH` IDENTISCH
zu `viewBrowseRepo()`s eigener Render-Formel (`view_browse_repo.go:659-682`)
rekonstruieren — sonst driftet Klick- gegen Render-Geometrie (derselbe Grund,
warum `windowStart` als gemeinsame Funktion existiert). Row-Mapping: erste
Zeile der Tree-Pane ist die Suchkopfzeile (`treeSearchLine`, KEIN Node-Ziel),
ab Zeile 2 `nodeWindowIdx := clickRow-1`, echter Node-Index =
`windowStart(len(nodes), bodyH-3, m.cursorPos(nodes)) + nodeWindowIdx`
(`bodyH-3`, exakt der Wert, den `m.treeRows` selbst an `windowAround`
übergibt, `view_browse_repo.go:680`). Doppelklick (Port devd
`doubleClickInterval` 500ms, NEUE Felder `lastClickIdx`/`lastClickAt`
+ `now()`-Methode) togglet expand/collapse auf demselben Node — devd-D03-
Semantik: Einzelklick auf einen BEREITS offenen Node togglet NICHT, nur
Cursor. Backlog/Review-Cockpit bekommen je eine analoge, aber EIGENE
`*ClickRow`-Funktion (verschiedene Cursor-Typen: `backlogList`
listState-Index vs. `reviewCursor` int in `reviewFlat`) — kein
Daten-Overreach in eine gemeinsame Funktion, mirror des `focusedBean()`-
Precedents (view-spezifische Resolver, gemeinsamer Aufrufer).

### g) Task-Schnitt + Reihenfolge

Implementation-plan.md »Epos E5« listet bereits die Reihenfolge 1-7 +
Abschluss — sie erfüllt die Handover-Vorgabe "Toast zuerst" bereits
strukturell (Toast=1, Maus=4, Maus braucht Toast's `toastHit`). Übernommen
1:1 als Task 1-8 unten, keine Umsortierung nötig:

| Task | bean | Inhalt |
|------|------|--------|
| 1 | `bt-6dts` | Toast-System |
| 2 | `bt-wpn9` | Help-Overlay `?` |
| 3 | `bt-e4a6` | Yank `y` |
| 4 | `bt-mne6` | Maus |
| 5 | `bt-0l8c` | Settings |
| 6 | `bt-zhwl` | Lobby + Repo-Picker `p` |
| 7 | `bt-ggt2` | Archiv-Sicht |
| 8 | `bt-7dfj` | E5-Abschluss |

---

## Geteilte Infrastruktur

### Model-Felder (types.go), pro Task ergänzt

```go
// Task 1 (Toast)
toast *toastState

// Task 2 (Help)
helpOpen bool

// Task 4 (Maus)
lastClickIdx  int
lastClickAt   time.Time
clock         func() time.Time // test-injizierbar, Port devd m.clock/m.now()

// Task 5 (Settings) — Entscheidung bei Task-Start: eigenes Feld vs. gezielte
// Felder; Vorschlag: settings config.Settings (ganzer Struct, einfachste
// Variante, mirror app.go/m.cfg-Precedent aus devd)
settings config.Settings

// Task 6 (Lobby)
view viewID // +1 neuer Case viewLobby
repoQuery  string
repoSearch textinput.Model
repoList   listState
watchStop  func() // aktuell laufender Watcher-Stop, für Repo-Wechsel

// Task 7 (Archiv)
showArchived bool // default false
```

### Neue Pakete/Dateien (Überblick, Details je Task)

- `internal/clip/clip.go` (Task 3, Port devd VERBATIM)
- `internal/config/settings.go`, `internal/config/state.go` (Task 5, Port
  devd, reduziertes Schema)
- `internal/tui/overlay_show_toast.go`, `context.go`, `overlay_shortcuts.go`,
  `mouse.go`, `view_lobby.go`, `box_form_settings.go` (neu)

### Capture-Order-Erweiterung (`handleKey`, update.go) — Reihenfolge nach Task 6

```
m.confirmQuit               (unverändert, oberste Priorität)
m.searchActive               (unverändert)
m.filterOpen                  (unverändert)
m.form != nil                  (unverändert)
m.overlay != overlayNone        (unverändert)
m.helpOpen                       (NEU, Task 2 — Precedent filterOpen)
keys.Help ("?")                   (NEU, Task 2 — von überall, design decision h)
m.paletteOpen                      (unverändert)
keys.Palette                        (unverändert)
keys.Picker ("p")                    (NEU, Task 6 — von überall, design decision h)
m.view == viewLobby                   (NEU, Task 6 — volle Capture wie Review-Cockpit)
m.view == viewReviewCockpit             (unverändert)
... (Rest unverändert)
```

`composeOverlays` (`view_browse_repo.go:582`) — neue Cases in
Painter's-Algorithmus-Reihenfolge (spät = oben):

```
m.filterOpen → overlayID-Switch → m.form → m.paletteOpen
  → m.helpOpen   (NEU, Task 2 — VOR confirmQuit)
  → m.confirmQuit (unverändert, bleibt topmost unter den Modalen)
```

Toast (`renderToast`) liegt AUSSERHALB von `composeOverlays` — siehe design
decision a, Punkt 2 (Top-Level-`View()`-Wrap, über composeOverlays' GESAMTEM
Ergebnis inkl. confirmQuit).

---

## Task 1: Toast-System (`bt-6dts`)

**Files:**
- Create: `internal/tui/overlay_show_toast.go`, `internal/tui/overlay_show_toast_test.go`
- Modify: `internal/tui/messages.go` (toastExpiredMsg/toastTimeout), `internal/tui/types.go`
  (model.toast), `internal/tui/update.go` (Dispatch-Case + jede `m.err = ...`-
  Stelle bekommt showToast), `internal/tui/view_browse_repo.go` (View()-Wrap)

- [ ] **Step 1: Failing tests** — `overlay_show_toast_test.go`:
  `TestShowToastSingleSlotReplacesPrevious` (zweiter `showToast`-Aufruf
  ersetzt den ersten, `seq` inkrementiert), `TestShowToastDebounceWindow
  UpdatesInPlace` (zwei Aufrufe <300ms auseinander → gleiche `seq`, Inhalt
  aktualisiert), `TestStickyToastNoAutoDismiss` (sticky=true →
  zurückgegebener `tea.Cmd` ist nil), `TestClearToastUnlessSticky` (sticky
  bleibt, non-sticky wird `nil`), `TestToastGeometryTopRight` (x+w == m.width,
  y==0), `TestToastHitTest` (Punkt innerhalb/außerhalb der Box),
  `TestDismissToastJumpsToTarget` (gesetztes `target.view` wird übernommen).
- [ ] **Step 2:** `command go test ./internal/tui/ -run TestShowToast` → FAIL
  (Typen/Funktionen fehlen).
- [ ] **Step 3: Implement** — Port devd `overlay_show_toast.go` (siehe design
  decision a für die zwei Geometrie-Deviationen). `messages.go`: `type
  toastExpiredMsg struct{ seq int }`, `func toastTimeout(seq int, kind
  toastKind) tea.Cmd { return tea.Tick(toastDuration(kind), func(time.Time)
  tea.Msg { return toastExpiredMsg{seq} }) }`. `types.go`: `toast *toastState`
  Feld. `update.go` Update(): `case toastExpiredMsg: return
  m.handleToastExpired(msg)` (neue Funktion: `if m.toast != nil && msg.seq ==
  m.toast.seq { m.toast = nil }; return m, nil`).
- [ ] **Step 4:** Tests → PASS.
- [ ] **Step 5: Dual-Write-Audit** — jede bestehende `m.err = ...`-Zuweisung
  (grep `m.err = ` in update.go: `applyMutationResult` non-conflict UND
  conflict Zweig, `applyBleveResult`, `applyPaletteBleveResult`,
  `applyEditorFinished` err/gone-Zweige, `applyLoaded`, `createInFlightNote`-
  Guard) bekommt eine gebatchte `showToast`-Ergänzung: `m, toastCmd :=
  m.showToast(kind, title, "", nil, sticky); return m, tea.Batch(toastCmd,
  reloadCmd)` (reloadCmd = der bestehende Rückgabewert der Funktion). Kind
  `toastError` für harte Fehler, `toastWarn` für Hinweise. `sticky=true` NUR
  im `errors.Is(err, data.ErrConflict)`-Zweig von `applyMutationResult`.
- [ ] **Step 6: Failing regression test** —
  `TestConflictToastIsStickyAndSurvivesReload` (fakeBeansConflict-Muster wie
  `editor_test.go`: Mutation → ErrConflict → `m.toast.sticky == true` →
  `beansLoadedMsg` (simulierter Reload) angewendet → `m.toast` weiterhin
  gesetzt).
- [ ] **Step 7:** FAIL → Implement (falls Step 5 unvollständig) → PASS.
- [ ] **Step 8: View()-Wrap** — `view_browse_repo.go` `View()`:
  ```go
  func (m model) View() string {
      var out string
      switch m.view {
      case viewBacklog:
          out = m.viewBacklog()
      case viewReviewCockpit:
          out = m.viewReviewCockpit()
      default:
          out = m.viewBrowseRepo()
      }
      return m.renderToast(out)
  }
  ```
- [ ] **Step 9: Mouse-Vorbereitung (Stub)** — `update.go` Update(): Kommentar-
  Stub an der Stelle, wo Task 4 `case tea.MouseMsg: return m.handleMouse(msg)`
  einfügen wird (VOR einem etwaigen `m.form != nil`-Kurzschluss) — noch KEIN
  Code, nur die Platzierungs-Dokumentation, damit Task 4 nicht neu recherchieren
  muss.
- [ ] **Step 10:** `command go test ./... -short` grün, `gofmt -l .`/`go vet
  ./...` leer.
- [ ] **Step 11:** Commit `feat(tui): Toast-System (Port overlay_show_toast.go) + Konflikt-sticky`

---

## Task 2: Help-Overlay `?` (`bt-wpn9`)

**Files:**
- Create: `internal/tui/overlay_shortcuts.go`, `internal/tui/overlay_shortcuts_test.go`
- Modify: `internal/tui/types.go` (model.helpOpen), `internal/tui/update.go`
  (Capture-Block + `?`-Dispatch), `internal/tui/view_browse_repo.go`
  (composeOverlays-Case)

- [ ] **Step 1: Failing tests** — `TestHelpBoxRendersEveryGroup` (alle 3
  `keys.helpGroups()`-Titel erscheinen im gerenderten Output),
  `TestKeyHelpOpensFromAnyView` (`m.view = viewBacklog`, `?` gedrückt →
  `m.helpOpen == true`), `TestKeyHelpEscQCloseHelp` (esc UND q UND `?`
  schließen, Footer-Text "esc/?/q: close" wörtlich geprüft),
  `TestHelpCapturesSingleKeysWhileOpen` (`q` bei offenem Help schließt NUR
  Help, `m.confirmQuit` bleibt false).
- [ ] **Step 2:** FAIL (helpBox/helpOpen/keyHelp fehlen).
- [ ] **Step 3: Implement** — `overlay_shortcuts.go`: `helpBox()` Port devd
  VERBATIM (Spaltenbreite `keyW` global über alle Gruppen, `modalPanel`-
  basiert, Titel "Keyboard shortcuts", Footer "esc/?/q: close"). `keyHelp
  (msg tea.KeyMsg) (tea.Model, tea.Cmd)`: `switch msg.String() { case "esc",
  "?", "q": m.helpOpen = false; return m, nil }; return m, nil` (alles
  andere: no-op, volle Capture). `update.go` handleKey: `if m.helpOpen {
  return m.keyHelp(msg) }` direkt vor dem `if keybind.Matches(msg,
  keys.Palette)`-Block (design decision h: gleiche Ebene wie `ctrl+k`), plus
  `if keybind.Matches(msg, keys.Help) { m.helpOpen = true; return m, nil }`
  darunter.
- [ ] **Step 4:** Tests → PASS.
- [ ] **Step 5: composeOverlays** — neuer Case nach `m.paletteOpen`, vor
  `m.confirmQuit`: `if m.helpOpen { out = placeOverlay(out, m.helpBox(), w,
  h) }`.
- [ ] **Step 6: Drift-Guard-Regression** — `command go test ./internal/tui/
  -run TestHelpGroupsCoverEveryBindingExactlyOnce` UND `TestKeymapNoCtrlSQ`
  explizit laufen lassen (keine neuen `keyMap`-Felder in diesem Task,
  Guard MUSS unverändert grün bleiben — reine Bestätigung, kein neuer Code).
- [ ] **Step 7:** `command go test ./... -short` grün, gofmt/vet leer.
- [ ] **Step 8:** Commit `feat(tui): Help-Overlay ? aus zentraler Keymap generiert`

---

## Task 3: Yank `y` (`bt-e4a6`)

**Files:**
- Create: `internal/clip/clip.go`, `internal/clip/clip_test.go` (falls
  sinnvoll testbar — OSC52 schreibt auf stderr, nativ ist best-effort;
  Testfokus auf die REINE String-Formatierung, nicht den echten Clipboard-
  Schreibvorgang), `internal/tui/context.go`, `internal/tui/context_test.go`
- Modify: `internal/tui/view_review_cockpit.go` (reviewStandMarkdown +
  `y`-Override), `internal/tui/update.go` (keyNodeAction `y`-Case)

- [ ] **Step 1:** `command go get github.com/aymanbagabas/go-osc52/v2`
  (bereits `// indirect` in go.mod — Befehl promotet auf direct require,
  KEIN neuer Dependency-Download).
- [ ] **Step 2: Failing tests** — `clip_test.go`: keine sinnvolle
  Verhaltensprüfung ohne echtes Terminal/Clipboard möglich; stattdessen
  `context_test.go` (der eigentliche Test-Schwerpunkt):
  `TestBeanContextLeafHasNoChildrenTable` (Task-Bean ohne Kinder → Output
  enthält KEIN "## Children"), `TestBeanContextParentHasChildrenTable`
  (Epic mit 2 Kindern → Tabelle mit exakt 2 Zeilen, `sortBeans`-Reihenfolge),
  `TestBeanContextResolvesParentTitleAndRelations` (Bean mit Parent+Blocking
  → aufgelöste Titel, nicht nur IDs), `TestReviewStandMarkdownGroupsByEpic`
  (2 to-review-Beans unter verschiedenen Epics → 2 Gruppen-Überschriften +
  Rework-Sektion bei vorhandenen rework-Tags).
- [ ] **Step 3:** FAIL (`beanContext`/`reviewStandMarkdown` fehlen).
- [ ] **Step 4: Implement** — `internal/clip/clip.go`: Port devd VERBATIM
  (`Copy(s string) error`, OSC52 + tmux/screen-Detection + `nativeCopy`
  best-effort pbcopy/wl-copy/xclip/xsel). `context.go`: `beanContext(idx
  *data.Index, b *data.Bean) string` — Format wie design decision b oben.
  `view_review_cockpit.go`: `reviewStandMarkdown(idx *data.Index) string` —
  reuse `reviewQueue(idx)`/`reviewRework(idx)` (E4 T3, KEIN neuer
  Gruppierungscode).
- [ ] **Step 5:** Tests → PASS.
- [ ] **Step 6: Dispatch** — `update.go` `keyNodeAction`: neuer Case
  `keybind.Matches(msg, keys.Yank)` — außerhalb Review-Cockpit:
  `clip.Copy(beanContext(m.idx, b))`, Erfolg → `showToast(toastInfo,
  "Kopiert: "+b.ID, "", nil, false)`, Fehler → `showToast(toastWarn, "Yank
  fehlgeschlagen", "", nil, false)`. Orphan-Root-Cursor (`b == nil`) → no-op
  (kein Toast, kein Crash).
- [ ] **Step 7: Review-Override** — `view_review_cockpit.go` `keyReview
  Cockpit`: `y`-Case VOR dem generischen `keyNodeAction`-Pfad (Capture-Order,
  design decision h — Cockpit überschreibt bereits s/t/B/c/d/e, `y` reiht
  sich ein) — `clip.Copy(reviewStandMarkdown(m.idx))`.
- [ ] **Step 8: Failing tests für Step 6/7** —
  `TestYankShowsConfirmationToast` (US-11-Kern), `TestYankOnOrphanRootNoop`,
  `TestReviewCockpitYankUsesReviewStandNotSingleBean` (Cockpit-Override
  greift, nicht `beanContext`).
- [ ] **Step 9:** FAIL → PASS.
- [ ] **Step 10:** `command go test ./... -short` grün, gofmt/vet leer.
- [ ] **Step 11:** Commit `feat(tui): Yank y (OSC52+nativ, Bean-/Epic-Kontext + Review-Stand)`

---

## Task 4: Maus (`bt-mne6`)

**Files:**
- Create: `internal/tui/mouse.go`, `internal/tui/mouse_test.go`
- Modify: `internal/tui/types.go` (lastClickIdx/lastClickAt/clock),
  `internal/tui/update.go` (Update() MouseMsg-Case + `now()`-Methode),
  `internal/tui/view_browse_repo.go` (treeClickRow), `internal/tui/
  view_browse_backlog.go` (backlogClickRow), `internal/tui/
  view_review_cockpit.go` (reviewClickRow)

- [ ] **Step 1: Failing tests** — `mouse_test.go`:
  `TestWheelUpDownMovesTreeCursor`, `TestWheelMovesBacklogCursor`,
  `TestWheelMovesReviewCursor`, `TestClickSetsTreeCursor`,
  `TestDoubleClickTogglesExpand` (zwei `tea.MouseMsg` <500ms auseinander auf
  demselben Node), `TestSingleClickOnOpenNodeDoesNotCollapse` (devd-D03-
  Semantik), `TestMouseIgnoredWhileFormOpen` (Overlay-Guard — ABER Toast-
  Klick MUSS trotzdem durchgehen, siehe nächster Test),
  `TestMouseIgnoredWhileOverlayOpen`,
  `TestToastClickDismissesEvenWithFormOpen` (Cross-Feature-Fix-Regression,
  Kernpunkt design decision a).
- [ ] **Step 2:** `command go test ./internal/tui/ -run TestWheel` → FAIL.
- [ ] **Step 3: Implement `now()`** — `update.go`: `const
  doubleClickInterval = 500 * time.Millisecond`; `func (m model) now()
  time.Time { if m.clock != nil { return m.clock() }; return time.Now() }`
  (Port devd VERBATIM, testbar über `m.clock`-Injection).
- [ ] **Step 4: Implement `mouse.go`** — `handleMouse(msg tea.MouseMsg)
  (tea.Model, tea.Cmd)`:
  1. Toast-Hit-Test ZUERST (unabhängig von jedem Overlay-Zustand):
     `if msg.Button == tea.MouseButtonLeft && msg.Action ==
     tea.MouseActionPress && m.toastHit(msg.X, msg.Y) { return
     m.dismissToast() }`.
  2. Overlay-Guard: `if m.form != nil || m.overlay != overlayNone ||
     m.paletteOpen || m.filterOpen || m.searchActive || m.helpOpen ||
     m.confirmQuit || m.view == viewLobby { return m, nil }`.
  3. Wheel: `tea.MouseButtonWheelUp`/`WheelDown` → je `m.view` den
     passenden Cursor-Move-Helper aufrufen (Tree: dieselbe Logik wie
     `keyTree`s up/down, Backlog: `keyBacklog`s, Review: `keyReviewCockpit`s
     — als kleine gemeinsam aufrufbare Helfer faktorisieren, falls noch
     nicht vorhanden, KEIN Code-Duplikat).
  4. Linksklick: `switch m.view { case viewBrowseRepo: ... case
     viewBacklog: ... case viewReviewCockpit: ... }`.
- [ ] **Step 5: Implement `treeClickRow`** (`view_browse_repo.go`) — Formel
  exakt wie design decision f oben (lw/rw/bodyH-Rekonstruktion,
  Suchkopfzeilen-Offset -1, `windowStart(len(nodes), bodyH-3, cursorPos)`).
  Doppelklick-Logik: `isDouble := idx == m.lastClickIdx && m.now().Sub
  (m.lastClickAt) < doubleClickInterval`; `m.lastClickIdx, m.lastClickAt =
  idx, m.now()`; danach devd-D03: `if n.expand { switch { case n.open &&
  isDouble: collapse; case !n.open: expand } }` (Einzelklick auf offenen
  Node: nur Cursor, kein Toggle).
- [ ] **Step 6: Implement `backlogClickRow`/`reviewClickRow`** — analoge,
  aber eigenständige Funktionen (kein Doppelklick-Bedarf bei Backlog/Review,
  flache Listen, nur Cursor-Set).
- [ ] **Step 7: Wire Update()** — `case tea.MouseMsg: return
  m.handleMouse(msg)` — Position: VOR einem etwaigen `if m.form != nil`
  early-return (falls ein solcher inzwischen existiert; aktuell hat
  beans-tui's `Update()` KEINEN solchen Top-Level-Kurzschluss vor dem
  switch, `m.form != nil` wird nur NACH dem switch für nicht-gematchte
  Message-Typen geprüft — der `tea.MouseMsg`-Case im switch selbst deckt
  das also automatisch ab; **Verifikation als eigener Test**
  `TestToastClickDismissesEvenWithFormOpen` schließt diese Annahme ab,
  nicht nur Doku).
- [ ] **Step 8:** Tests → PASS.
- [ ] **Step 9:** `command go test ./... -short` grün, gofmt/vet leer,
  Goldens (`TestChromeGolden`/`TestTreeGolden`/`TestTreeGoldenDeterministic`/
  `TestBacklogGolden`/`TestBacklogGoldenDeterministic`) 2x grün (Maus
  berührt keinen Default-Render-Pfad).
- [ ] **Step 10:** Commit `feat(tui): Maus (Wheel/Klick/Doppelklick, Toast-Dismiss-Vorrang)`

---

## Task 5: Settings (`bt-0l8c`)

**Files:**
- Create: `internal/config/settings.go`, `internal/config/settings_test.go`,
  `internal/config/state.go`, `internal/config/state_test.go`,
  `internal/tui/box_form_settings.go`, `internal/tui/box_form_settings_test.go`
- Modify: `internal/tui/editor.go` (configuredEditor-Var), `internal/theme/
  theme.go` (SetAccent, falls fehlend — prüfen), `internal/tui/app.go`
  (LoadSettings beim Start), `cmd/tui.go` (Settings durchreichen),
  `internal/tui/overlay_palette.go` (paletteActions "settings: öffnen")

- [ ] **Step 1:** `command go get gopkg.in/yaml.v3`.
- [ ] **Step 2: Failing tests** — `settings_test.go` (Port devd
  `settings_test.go`-Muster, reduziertes Schema): `TestParseSettingsYAML`,
  `TestMergeSettingsOnlySetFieldsWin`, `TestValidateSettingsClampsTreeWidth`
  (Range wie devd: min 24/max 60, aus design-spec §8 "Baumbreite" implizit
  via `masterDetailWidths`s `treeWidthFloor`-Parameter), `TestValidate
  SettingsRejectsInvalidAccentHex`, `TestSaveUserSettingsReadModifyWrite`.
  `state_test.go`: `TestStateLoadMissingFileReturnsZero`,
  `TestStateSaveAndLoadRoundtrip`, `TestSetLastRepoPreservesOtherFields`.
- [ ] **Step 3:** FAIL.
- [ ] **Step 4: Implement `settings.go`** — Port devd `settings.go` MINUS
  `ModalWidth`/`StartProject`/`Keybindings` (design-spec nennt nur
  repos/editor/accent/treewidth), PLUS `Repos []string`. EIN Pfad
  `~/.config/beans-tui/config.yaml` (kein CWD-Override-Layer, design
  decision c). `state.go`: Port devd VERBATIM, `LastProject` → `LastRepo`
  umbenannt, gleiches Verzeichnis `~/.config/beans-tui/state.json`.
- [ ] **Step 5:** Tests → PASS.
- [ ] **Step 6: Editor-Präzedenz** — `editor.go`: `var configuredEditor
  string` (package-level, Default `""`); `editorBinary()` bekommt am Anfang:
  `if ed := strings.TrimSpace(configuredEditor); ed != "" { return
  strings.Fields(ed) }` — VOR der bestehenden `$VISUAL`/`$EDITOR`-Schleife.
- [ ] **Step 7: Failing test** —
  `TestEditorBinaryPrefersConfiguredEditorOverEnv` (configuredEditor="code
  -w", VISUAL/EDITOR gesetzt → configuredEditor gewinnt);
  `TestEditorBinaryResolvesVisualThenEditorThenVi` (bestehender E3-Test,
  UNVERÄNDERT, muss weiter grün sein — configuredEditor bleibt in diesem
  Test "").
- [ ] **Step 8:** FAIL → PASS (inkl. Regressionsbeleg für den unveränderten Test).
- [ ] **Step 9: Settings-Form** — `box_form_settings.go`: huh-Form (Port
  devd `form_edit_settings.go`-Muster, `forms_shared.go`-Infra: `styleForm`/
  `formBoxWidth`/`formInnerWidth`/`formInnerHeight` wiederverwenden). Felder:
  `repos` (Textarea, ein Pfad je Zeile), `editor` (Input, Platzhalter "leer =
  $VISUAL/$EDITOR/vi"), `accent` (Input, Hex, Validator wie
  `hexColor`-Regex), `tree_width` (Input, numerisch, Validator Bereich
  24-60). `formKind = "settings"`. Submit-Case (`forms_shared.go`
  `submitForm`): `config.SaveUserSettings(...)`, `configuredEditor =
  form.editor`, `theme.SetAccent(form.accent)` (LIVE, kein Neustart, Port
  devd DD2-221-Prinzip).
- [ ] **Step 10: theme.SetAccent prüfen/implementieren** — falls
  `internal/theme/theme.go` noch keine Laufzeit-Override-Funktion für den
  Mauve-Akzent hat (aktuell `theme.Mauve`/`theme.Accent` vermutlich
  `const`/package-level `var` fixer Hex-Werte, VERIFIZIEREN bei Task-Start):
  `SetAccent(hex string)` überschreibt den Akzent-Farbwert zur Laufzeit
  (leerer/ungültiger Hex → No-Op, Built-in bleibt).
- [ ] **Step 11: Palette-Eintrag** — `overlay_palette.go` `paletteActions`:
  neuer globaler Eintrag `{actionID: "settings", label: "settings:
  öffnen"}`, `dispatchPalette`-Case öffnet das Settings-Form.
- [ ] **Step 12: Wiring** — `app.go` `Run()` (oder `cmd/tui.go` `RunE` davor):
  `settings, _ := config.LoadSettings()`; `configuredEditor =
  settings.Editor`; `theme.SetAccent(settings.Theme.Accent)`; `settings` an
  `newModel` durchreichen (`m.settings`-Feld).
- [ ] **Step 13: Failing tests für Step 9-12** —
  `TestSettingsFormPrefillsCurrentValues`,
  `TestSettingsFormSubmitSavesAndAppliesLive` (nach Submit:
  `configuredEditor` geändert, `config.LoadSettings()` liefert neuen Wert
  zurück).
- [ ] **Step 14:** FAIL → PASS.
- [ ] **Step 15:** `command go test ./... -short` grün, gofmt/vet leer.
- [ ] **Step 16:** Commit `feat(config): Settings (config.yaml+state.json) + Editor-Präzedenz`

---

## Task 6: Lobby V1 + Repo-Picker `p` (`bt-zhwl`)

**Files:**
- Create: `internal/tui/view_lobby.go`, `internal/tui/view_lobby_test.go`,
  `internal/tui/switch_repo_test.go`
- Modify: `internal/tui/types.go` (viewLobby + repoQuery/repoSearch/
  repoList/watchStop), `internal/tui/messages.go` (repoSwitchedMsg,
  switchRepoCmd), `internal/data/watcher.go` (startWatch-Wrapper),
  `internal/tui/app.go` (activeProgram-Var, watchStop-Wiring),
  `internal/tui/update.go` (repoSwitchedMsg-Case, `p`-Dispatch),
  `internal/config/state.go`-Nutzung (SetLastRepo), `cmd/tui.go`
  (Start-Trigger-Logik design decision d), `internal/tui/overlay_palette.go`
  ("repo: wechseln")

- [ ] **Step 1: Failing tests (Kern-Watcher-Lifecycle, isoliert von Bubbletea)**
  — `internal/data/watcher.go`-Erweiterung: `startWatch(repoDir string,
  notify func()) (stop func(), err error)` (dünner Wrapper, identische
  Signatur zu `Watch`, EIGENER Name nur zur Trennung Produktions-Wiring vs.
  Testbarkeit). Test (in `internal/tui`, da `switchRepoCmd` dort lebt):
  `TestSwitchRepoCmdStopsOldWatcherStartsNew` — zwei `newTestRepo`-artige
  Fixtures (eigener Helper `newTestRepoTUI(t)` falls `internal/data`s
  `newTestRepo` nicht paketübergreifend exportiert ist — PRÜFEN, ggf.
  exportieren als `data.NewTestRepoForTests` oder duplizieren, Entscheidung
  bei Task-Start), fake `notify func()` zählt Aufrufe. Datei-Touch im NEUEN
  Repo NACH `switchRepoCmd`-Ausführung → `notify` gefeuert. Datei-Touch im
  ALTEN Repo NACH dem Switch → `notify` NICHT gefeuert (alter Watcher tot).
- [ ] **Step 2:** FAIL (`switchRepoCmd`/`repoSwitchedMsg` fehlen).
- [ ] **Step 3: Implement `switchRepoCmd`** — `messages.go`:
  ```go
  type repoSwitchedMsg struct {
      client    *data.Client
      repoDir   string
      beans     []data.Bean
      watchStop func()
      err       error
  }

  func switchRepoCmd(oldStop func(), newRepoDir string, notify func()) tea.Cmd {
      return func() tea.Msg {
          if oldStop != nil {
              oldStop()
          }
          client := &data.Client{RepoDir: newRepoDir}
          beans, err := client.List()
          var newStop func()
          if err == nil {
              newStop, _ = data.Watch(newRepoDir, notify)
          }
          return repoSwitchedMsg{client: client, repoDir: newRepoDir, beans: beans, watchStop: newStop, err: err}
      }
  }
  ```
  KEIN `tea.Program`-Bezug in dieser Funktion — `notify` ist injiziert,
  testbar ohne echten Bubbletea-Runtime (genau der Punkt von Step 1).
- [ ] **Step 4:** Tests → PASS.
- [ ] **Step 5: app.go-Wiring** — `var activeProgram *tea.Program`
  (package-level); in `Run()`: `p := tea.NewProgram(...)`; `activeProgram =
  p` NACH der Konstruktion, VOR `data.Watch(...)`/`p.Run()`. Initialer
  `data.Watch`-Aufruf liefert seinen `stop` jetzt zusätzlich an `m.watchStop`
  (nicht mehr NUR lokale `Run()`-Variable — `newModel` bekommt den stop-func
  erst NACH dem ersten `Init()`, also über eine `initialWatchStoppedMsg`
  ODER direktes Feld-Setzen vor `p.Run()`: EINFACHSTE Variante, da `m` schon
  vor `tea.NewProgram` gebaut wird — `m.watchStop = stop` VOR
  `tea.NewProgram(m, ...)` aufrufen, Reihenfolge in `Run()` entsprechend
  umstellen: Watch VOR NewProgram, `p.Send`-Closure bekommt `p` erst NACH
  NewProgram — PRÜFEN, ob das die B05-Synchronität (kein `p.Send` vor
  `p.Run()`) verletzt; falls ja: `watchStop` bleibt vorerst `nil` im Modell
  bis zum ersten `repoSwitchedMsg`, und die INITIALE Watch-Stop-Funktion wird
  weiterhin nur in `Run()`s eigenem `defer stop()` gehalten (zwei getrennte
  Stop-Quellen: initial vs. jeder Switch) — Entscheidung final bei
  Task-Start anhand eines Kompilier-/Test-Versuchs treffen, hier beide
  Optionen dokumentiert, damit keine Sackgasse entsteht).
- [ ] **Step 6: view_lobby.go** — `homeLogoBlock()` (ASCII "beans"-Banner,
  EAW-neutral wie devd's `homeLogoLines`, eigener Text statt "DevDashboard"),
  `centerInto`/`pickerRowFill` Port devd VERBATIM (reine String-Utilities,
  keine devd-API-Kopplung), `repoPickerBody(w int) string` (Suchzeile +
  gefilterte `m.settings.Repos`-Liste, Metrik "Offen/Gesamt" — EINFACHSTE
  Variante ohne Latenz-Explosion: NUR beim Lobby-Öffnen einmalig pro
  konfiguriertem Repo `&data.Client{RepoDir: r}; c.List()` synchron NICHT
  blockierend genug für viele Repos — daher als `tea.Cmd`/`tea.Batch` pro
  Repo async, Ergebnis in einer neuen `repoMetricsMsg`-Message gesammelt;
  bei 2-5 konfigurierten Repos unkritisch), `viewLobby()` (Chrome +
  zentrierter Block, Port `viewHome` VERBATIM-Struktur).
- [ ] **Step 7: types.go** — `viewLobby viewID` (4. Case), `repoQuery
  string`, `repoSearch textinput.Model`, `repoList listState`, `watchStop
  func()`.
- [ ] **Step 8: update.go** — `case repoSwitchedMsg:` (neue `apply
  RepoSwitched`-Funktion): `m.client = msg.client`, `m.repoDir =
  msg.repoDir`, `m.watchStop = msg.watchStop`, `m.idx =
  data.NewIndex(msg.beans)`, `m.view = viewBrowseRepo`, Cursor/Expand-State
  zurückgesetzt (Port devd `selectProject`-Reset-Muster:
  `m.expanded = map[string]bool{}`, `m.cursorID = ""`,
  `m.detailFocus = false`), `config.SetLastRepo(msg.repoDir)`
  (Read-Modify-Write, Port devd DD2-273-Kommentar). `handleKey`: `if
  keybind.Matches(msg, keys.Picker) { m.view = viewLobby;
  m.repoQuery = ""; m.repoSearch.SetValue(""); m.repoList.cursor = 0;
  return m, nil }` VOR dem Review-Cockpit-Capture-Block (design decision h);
  `if m.view == viewLobby { return m.keyLobby(msg) }` als weiterer
  Vollcapture-Block (Precedent Review-Cockpit).
- [ ] **Step 9: keyLobby** (`view_lobby.go`) — Nav (up/down auf
  `repoList`), enter → `switchRepoCmd(m.watchStop, selectedRepo, func(){
  activeProgram.Send(watchMsg{}) })`, esc/q → zurück zu `viewBrowseRepo`
  falls bereits ein Repo geladen war, sonst quit-confirm (US-01-Parität mit
  devd `keyHome`), Tippen filtert `m.repoQuery`.
- [ ] **Step 10: Start-Trigger** — `cmd/tui.go` `RunE`: Logik aus design
  decision d (4 Fälle), `LoadSettings()` vor der Repo-Auflösung.
- [ ] **Step 11: Palette-Eintrag** — `overlay_palette.go`: `{actionID:
  "repo_picker", label: "repo: wechseln"}`.
- [ ] **Step 12: Failing tests für Step 6-11** —
  `TestLobbyShowsConfiguredRepos`, `TestLobbyFilterNarrowsBySearch`,
  `TestLobbySelectSwitchesRepoAndView`, `TestPickerKeyOpensLobbyFromAnyView`,
  `TestNoLobbyOnSingleRepoCwdMatch` (design decision d Kernfall — E1-E4-
  Verhalten bleibt intakt, KEINE Regression bei bestehendem Single-Repo-
  Start).
- [ ] **Step 13:** FAIL → PASS.
- [ ] **Step 14:** `command go test ./... -short` grün, gofmt/vet leer.
- [ ] **Step 15:** Commit `feat(tui): Lobby V1 + Repo-Picker p (Watcher-Lifecycle-Switch)`

---

## Task 7: Archiv-Sicht (`bt-ggt2`)

**Files:**
- Modify: `internal/data/bean.go` (IsArchived), `internal/tui/types.go`
  (showArchived), `internal/tui/box_filter_facets.go` (neue Menüzeile +
  Prädikat)
- Create: `internal/data/archive_test.go` (falls nicht bereits als
  `bean_test.go`-Erweiterung sinnvoller), `internal/tui/
  archive_visibility_test.go`

- [ ] **Step 1: Failing tests (empirischer Beleg, data package)** —
  `TestBeanIsArchivedDetectsPathPrefix` (`Bean{Path: "archive/x.md"}.
  IsArchived() == true`, `Bean{Path: "x.md"}.IsArchived() == false`),
  `TestListIncludesArchivedBeans` (echtes beans-Binary,
  `requireBeansBinary`-Guard: `newTestRepo(t)` + ein Bean auf `completed`
  setzen + `beans archive` CLI-Aufruf im Repo-Verzeichnis + erneutes
  `client.List()` → Bean weiterhin vorhanden, `Path` jetzt mit
  `"archive/"`-Präfix — empirischer Beweis für design decision e).
- [ ] **Step 2:** `command go test ./internal/data/ -run TestListIncludes
  Archived` → FAIL (`IsArchived` fehlt; der List-Teil selbst funktioniert
  schon heute, das ist der Beweis-Test, kein Implementierungs-Test).
- [ ] **Step 3: Implement** — `bean.go`: `func (b Bean) IsArchived() bool {
  return strings.HasPrefix(b.Path, "archive/") }`.
- [ ] **Step 4:** Tests → PASS.
- [ ] **Step 5: Failing tests (tui package)** —
  `TestArchivedBeanHiddenFromTreeByDefault` (Fixture mit einem `status:
  completed`-Bean → `m.beanMatches(b) == false` bei `m.showArchived ==
  false`), `TestArchivedBeanShownWhenShowArchivedToggled`
  (`m.showArchived = true` → `beanMatches == true`, sofern sonst nichts
  filtert), `TestArchivedOnlyMilestoneDisappearsWithNoOpenDescendant`
  (Milestone `completed` + alle Kinder `completed` → Tree zeigt GAR NICHTS
  von diesem Teilbaum bei `showArchived == false`, Beleg für die
  automatische Ancestor-Logik aus design decision e),
  `TestArchiveToggleDoesNotCountAsFilterActive` (`m.showArchived` beeinflusst
  NICHT `filterActive()` — der Statuszeilen-Hinweis "Filter aktiv" bleibt
  unabhängig).
- [ ] **Step 6:** FAIL.
- [ ] **Step 7: Implement** — `types.go`: `showArchived bool` (Default
  `false`, kein Zutun in `newModel` nötig — Go zero-value). `box_filter_
  facets.go`: `buildFilterItems` — neue, EIGENSTÄNDIGE Zeile (NICHT Teil der
  `data.StatusValues()`-Schleife) `ffItem{facet: "archive", value: "show",
  label: "Archivierte einblenden"}`. `keyFilterMenu`s bestehender
  `keys.Toggle`-Pfad: `if item.facet == "archive" { m.showArchived =
  !m.showArchived; return m, nil }` VOR dem generischen
  `filterStatus`/`filterType`/...-Toggle-Zweig. Neue `beanMatchesArchive
  (b *data.Bean) bool`: `if m.showArchived { return true }; return
  b.Status != "completed" && b.Status != "scrapped"`. `beanMatches`
  (`box_filter_facets.go:155`): drittes AND-Glied `m.beanMatchesArchive(b)
  && ...` neben `beanMatchesSearch`/`beanMatchesFacets`.
- [ ] **Step 8:** Tests → PASS.
- [ ] **Step 9:** `command go test ./... -short` grün, gofmt/vet leer,
  Goldens (`TestTreeGolden`/`TestBacklogGolden` u.a.) 2x grün — Default-aus-
  Zustand identisch zu heute, DA alle bestehenden Fixture-/Golden-Beans
  NICHT `completed`/`scrapped` sind (verifizieren, sonst Golden-Update
  nötig — falls doch: `-update`-Flag, neuer Golden-Snapshot, im Commit-Body
  begründen).
- [ ] **Step 10:** Commit `feat(tui): Archiv-Sicht (completed/scrapped Default-aus, togglebar)`

---

## Task 8: E5-Abschluss (`bt-7dfj`)

- [ ] `command go test ./...` (2x, ohne -short) grün, `command go test
  ./... -race` grün, `command go build -o bin/bt .` ok, `command gofmt -l .`
  leer, `command go vet ./...` leer.
- [ ] Alle Goldens 2x grün.
- [ ] tmux-Smoke im Scratch-Repo (mehrere Fixture-Beans + 2 Scratch-Repos
  für den Lobby-Wechsel): Toast (Konflikt-Fall sticky bis Klick — E3-I01-
  Regressionsbeweis), Help-Overlay (`?`/esc/q), Yank (Clipboard-Inhalt via
  `pbpaste` nach OSC52-Capture verifiziert, sowohl Bean- als auch Review-
  Stand-Kontext), Maus (Klick/Wheel/Doppelklick), Settings-Form (Editor/
  Accent live ohne Neustart), Lobby+Repo-Wechsel (Watcher reagiert im NEUEN
  Repo, NICHT mehr im alten), Archiv-Toggle — als Beleg im Commit-Body/
  Bean-Body dokumentiert (Format wie bt-jpgn/bt-hxyo Abschluss-Bodies).
- [ ] beans-Pflege: `bt-6dts`/`bt-wpn9`/`bt-e4a6`/`bt-mne6`/`bt-0l8c`/
  `bt-zhwl`/`bt-ggt2` auf `completed` (agent-abschließbar), `bt-5h4d`
  (Epic) bekommt Tag `to-review` (NICHT completed — PO-Gate,
  implementation-plan.md »Epos-Abschluss«).
- [ ] README.md: Kurzabschnitt Settings-Pfad (`~/.config/beans-tui/`) +
  Keymap-Ergänzung (p/y/?/Maus) falls README eine Keymap-Referenz führt
  (prüfen, ggf. neu anlegen).
- [ ] docs/SSTD.md: Pointer-Update falls nötig (voraussichtlich
  unverändert — Worktree-Weiche/Referenzen betreffen E5 nicht).
- [ ] Commit `docs: README + E5-Abschluss`.
- [ ] Skill `ce-nsp-auto`: Handover-Prompt für E6 (Validierung & Release,
  bean `bt-zk9p`) erzeugen.

---

## Selbst-Review (Plan gegen design-spec + implementation-plan + bean-Body)

- **Scope-Coverage:** design-spec §9 IN-Liste E5-relevante Punkte — Yank
  (OSC52+nativ) ✓ Task 3 · `$EDITOR`-Integration bleibt E3, NUR Präzedenz-
  Erweiterung ✓ Task 5 · Toast ✓ Task 1 · Help-Overlay ✓ Task 2 · Maus
  (Wheel, Klick-Cursor — Doppelklick zusätzlich, Handover-Pflicht) ✓ Task 4
  · Settings (repos/editor/accent/Baumbreite) ✓ Task 5 · ASCII-Icon-
  Fallback (bereits T6/`BT_ASCII_ICONS`, KEINE E5-Arbeit nötig — korrekt
  nicht in Task-Liste) · Archiv-Sicht ✓ Task 7. V1 (Lobby) ✓ Task 6, V8-Rest
  (Toast/Help aus §6-Matrix) ✓ Task 1/2.
- **US-Zuordnung (design-spec §12):** US-10 (Live-Reload, bereits E1
  T5/T8 — E5 fügt NUR den Watcher-Lifecycle-SWITCH für Task 6 hinzu, keine
  Neuimplementierung des Grundmechanismus) · US-11 (Yank) ✓ Task 3 · US-13
  (Help) ✓ Task 2 · US-14 (Repo-Wechsel) ✓ Task 6.
- **E3-Übernahme-Pflicht (bean-Body bt-5h4d) erfüllt:** Toast sticky für
  ErrConflict ✓ Task 1 design decision a + Step 6 Regressionstest;
  Recovery-Tempfile-Pfad (`applyMutationResult`s `conflictWithRecovery`-
  Zweig, bereits seit E3/F2 vorhanden) bleibt UNVERÄNDERT lesbar, da der
  Dual-Write (`m.err` UNVERÄNDERT + Toast zusätzlich) den bestehenden
  `TestEditorFinishedConflictWritesRecoveryTempFileAndSurfacesPath`-Test
  (E3, `editor_test.go`) nicht anfasst.
- **Placeholder-Check:** kein Task unten Step-1-Ebene ohne konkrete
  Testnamen/Funktionssignaturen — Task 6 (Lobby) hat als EINZIGER Punkt eine
  bewusst offen gelassene Implementierungs-ENTSCHEIDUNG (Step 5, initiale
  vs. Switch-Watcher-Stop-Verdrahtung) mit BEIDEN Optionen ausformuliert,
  keine echte Lücke — das ist Architektur-Unsicherheit, die erst beim
  tatsächlichen `tea.Program`-Lifecycle-Test entschieden werden kann, kein
  fehlendes Nachdenken.
- **Datei-Namenskonvention** (`<art>_<verb>_<entität>.go`) eingehalten:
  `overlay_show_toast.go`/`overlay_shortcuts.go` (Port-Namen 1:1 von devd
  übernommen, passen ins Schema), `view_lobby.go`, `box_form_settings.go`,
  `mouse.go` (Ausnahme — kein `<art>_<verb>`-Schema, da es KEIN Overlay/View/
  Box ist, sondern eine Dispatcher-Datei parallel zu `update.go`; devd hat
  den Mausteil direkt in `update.go` — hier bewusst ausgelagert für
  Lesbarkeit, gleiche Begründung wie `editor.go`/`context.go`, die auch
  keinem strikten `<art>_<verb>_<entität>`-Muster folgen).
- **Typ-Konsistenz:** `data.Bean`/`data.Client`/`data.Index` unverändert
  konsumiert (Task 7 fügt NUR eine Methode `IsArchived()` hinzu, kein neues
  Feld) ✓ · `toastState`/`toastKind`/`toastTarget` (Task 1) und
  `config.Settings`/`config.State` (Task 5) sind NEUE, in sich geschlossene
  Typen ohne Kollision mit bestehenden Namen (verifiziert: kein `Settings`/
  `State`-Symbol existiert bereits im `tui`-Package).
