# Epos E3 — Mutationen (voll granular)

> Geschrieben beim Epos-Start-Ritual (implementation-plan.md »Epos-Rituale«). Struktur
> identisch zu E1/E2: je Task Files/TDD-Steps/Port-Referenzen/Commit. Quelle der
> Wahrheit: `design-spec.md` §6 (V7/V8) + §5 + §7 und der Epic-bean-Body `bt-gzcu`
> (PFLICHT-Item B01). Port-Quellen:
> **devd** (`~/Obsidian/tools/DeveloperDashboard/apps/cli-go/internal/tui/`):
> `forms_shared.go`, `form_create_issue.go`, `box_confirm_create.go`,
> `box_confirm_delete.go`, `editor.go`;
> **beans-upstream-TUI** (`~/Obsidian/tools/lean-stack/beans-src/internal/tui/`):
> `statuspicker.go`, `typepicker.go`, `prioritypicker.go`, `tagpicker.go`,
> `parentpicker.go`, `blockingpicker.go` + `pkg/beancore/links.go` (ValidParentTypes).
> Task-beans: `bt-dlgk`(T1) `bt-8v69`(T2) `bt-p1uz`(T3) `bt-y4ly`(T4) `bt-sl45`(T5)
> `bt-ppzb`(T6) `bt-qzwt`(T7), parent `bt-gzcu`.

**Liefert:** V7 (Create-/Edit-Forms mit huh, Confirm-Gate, `$EDITOR`) und der
Mutations-Teil von V8 (Status-/Type-/Prio-Menü, Tag-/Parent-/Blocking-Picker,
Delete-Confirm mit Kinder-Count) — US-06/US-07. Der Datenlayer ist bereits KOMPLETT
(`internal/data/mutations.go`: Create/SetStatus/SetPriority/SetType/SetTitle/AddTag/
RemoveTag/SetParent/RemoveParent/AddBlockedBy/RemoveBlockedBy/AppendBody/Delete,
alle mit `--if-match`, `ErrConflict` via `errors.Is`, JSON-Envelope-Klassifikation) —
E3 verdrahtet ihn in die UI und ergänzt nur drei fehlende dünne Wrapper (`SetBody`,
`AddBlocking`, `RemoveBlocking`, alle drei gegen `beans update --help` verifiziert).
Schließt außerdem das PFLICHT-Item aus dem E2-Abschluss: **B01** (X/FilterClear wirkt
nicht in der Backlog-View — Fix in Task 1, erster kleiner Schritt).

**NICHT E3-Scope:** Toast (E5 — Konflikt-Feedback läuft interim über die Statuszeile,
`m.err`, s. Design-Entscheidung d) · Review-Tag-Konvention §5 (E4) · Command-Center
(E4) · Settings/`config.yaml` (E5 — `$EDITOR`-Auflösung daher env-basiert, s. Task 5).

**Reihenfolge (Task-bean `blocked_by`, streng sequentiell):** T1 → T2 → T3 → T4 →
T5 → T6 → T7. T1 legt die geteilte Infrastruktur (Mutation-Cmd/Msg, keyNodeAction-
Dispatch, Overlay-Enum, Enum-Single-Source), von der jede Folge-Task abhängt; T4 legt
die huh-Form-Hosting-Infrastruktur, die T5 wiederverwendet; T6 verifiziert das
ETag-Verhalten quer über alles Vorherige.

---

## Design-Entscheidungen (vor Task 1 festgezurrt)

**(a) Picker/Menü-Overlay-Muster: `modalPanel`/`menuList`/`listState`, NICHT
`bubbles/list`.** beans-upstream baut jeden Picker als eigenes Sub-Modell auf
`bubbles/list` (mit Filtering, Delegates, WindowSize-Plumbing). beans-tui hat seit E1
T7/E2 T4 ein eigenes, schlankeres Muster: `modalBox`/`modalPanel` (modal.go) +
`menuList` + `listState` + `placeOverlay`, ein Zustand direkt auf `model` (Präzedenz:
`box_filter_facets.go`). E3 portiert die SEMANTIK der upstream-Picker (Enter-Verhalten,
Pending-Diff, Eligibility-Filter), aber auf das hauseigene Muster — kein `bubbles/list`,
keine per-Picker-Sub-Modelle. Windowing in langen Picker-Listen über das bestehende
`windowAround` (E1 Task 8), nie neu gebaut.

**(a2) EIN Overlay-Enum statt 6 neuer Bools.** Die E3-Overlays (Value-Menü, 3 Picker,
Create-Confirm, Delete-Confirm) sind strukturell wechselseitig exklusiv — sie werden
als EIN `model.overlay overlayID`-Feld (iota-Enum) geführt, nicht als 6 weitere Bools
neben `filterOpen`/`confirmQuit`. Die zwei bestehenden E2-Bools werden NICHT
retrofittet (out of scope, dokumentierte Koexistenz): `confirmQuit` und `filterOpen`
bleiben, die capture-Reihenfolge in `handleKey` regelt den Vorrang. `m.form != nil`
ist der dritte, separate Capture-Zustand (huh-Formulare sind kein Menü-Overlay).

**(a3) EIN kombiniertes Value-Menü auf `s` für Status+Type+Priority.** design-spec §7
reserviert für dieses Cluster genau EINEN Key (`s` Status-Menü; Type/Prio haben keinen
eigenen). implementation-plan »E3.1« nennt „Status-/Type-/Prio-Menüs" — aufgelöst als
EIN Menü mit drei Gruppen-Headern (exakt das `facetHead`-Muster des Filter-Menüs),
Cursor startet auf dem aktuellen Status des Beans, `(current)`-Markierung je Gruppe
(Port beans-src statuspicker.go isCurrent). Enter wendet den cursorierten Wert SOFORT
an und schließt (upstream-Enter-Semantik) — anders als das Filter-Menü, dessen enter
nur schließt.

**(b) Enum-Single-Source: `internal/data`, exportierte Value-Slices.** beans hat fixe
Enums (Status/Type/Priority je 5 Werte). Die kanonische Tier-Reihenfolge lebt bereits
in `data/index.go` (`statusOrder`/`priorityOrder`/`typeOrder`) — aber als MAPS
(Go-Maps haben keine Ordnung), und `box_filter_facets.go:47-71` dupliziert die Werte
bereits hart. Fix (Task 1): die Maps werden aus neuen geordneten Slices
(`statusValues`/`typeValues`/`priorityValues`) ABGELEITET (init-Schleife), exportiert
als `data.StatusValues()`/`TypeValues()`/`PriorityValues()` (defensive Kopie);
`buildFilterItems` und das neue Value-Menü konsumieren beide diese eine Quelle.
`internal/theme` bleibt reine Farb-/Glyph-Zuordnung (kennt Werte nur als map-Keys mit
Fallback — kein Enum-Owner).

**(c) `$EDITOR` via `tea.ExecProcess`** — Port devd `editor.go` (prepareEditor/
readEditorResult/editInEditor, Zeilen 47-96) VERBATIM: temp-Datei schreiben,
`tea.ExecProcess(cmd, callback)` suspendiert die TUI, Callback liest die Datei zurück
und liefert `editorFinishedMsg{content, changed, err}`. EINZIGE Abweichung:
`editorBinary()` löst `$VISUAL` → `$EDITOR` → Fallback `vi` auf (design-spec §7 sagt
wörtlich „$EDITOR"; devds `configuredEditor` setzt die E5-Settings-Datei voraus, die
noch nicht existiert).

**(d) ETag-Fluss: frischer ETag aus dem Index-Pointer bei Submit; Konflikt →
Statuszeile + unconditional Reload.** Der ETag wird NIE beim Öffnen eines Overlays
eingefroren, sondern bei JEDEM Submit frisch aus `m.idx.ByID[id].ETag` gelesen
(Helper `m.beanETag(id)`) — ein Watch-Reload zwischen Öffnen und Submit liefert damit
automatisch den neuesten Stand, und ein „echter" Konflikt bleibt nur das Fenster
Submit↔Disk. Jede Mutation läuft als async `mutateCmd` → `mutationDoneMsg{err}`;
der Update-Handler (`applyMutationResult`) macht IMMER `loadCmd` (Erfolg: neuen Stand
zeigen; Fehler: Stale-State auflösen) und setzt bei `errors.Is(err, data.ErrConflict)`
den Statuszeilen-Text „Konflikt: Bean extern geändert — neu geladen" (Toast ist E5;
`m.err`/Red-Slot der Statuszeile existiert seit E1, view.go statusBar).

**(e) Create-Form-Feldsatz + Confirm-Gate.** Felder (alle keyed, huh v1.0.0, Werte
nach StateCompleted per `GetString` — devd forms_shared.go-Konvention „kein
Pointer-Binding über Model-Copies"): `title` (Input, nonEmpty-Pflicht) · `type`
(Select aus `data.TypeValues()`, Default task) · `priority` (Select aus
`PriorityValues()`, Default normal) · `status` (Select aus `StatusValues()`, Default
todo) · `parent` (Input, optional, vorbelegt mit der Cursor-Bean-ID; validiert NUR
auf Existenz in `idx.ByID` oder leer — Typ-Hierarchie-Fehler laufen bewusst in den
CLI-`VALIDATION_ERROR`-Pfad/Statuszeile, kein Client-Duplikat der Server-Regel) ·
`tags` (Input, Whitespace/Komma-getrennt, jeder gegen die Tag-Regex validiert) ·
`body` (Text, optional, huh-eigener ctrl+e-Editor via `ExternalEditor(false)` AUS —
devd DD2-234-Falle: huh broadcastet das Editor-Ergebnis an alle Text-Felder).
Confirm-Gate: Port devd `box_confirm_create.go` (submitForm parkt den fertigen Cmd →
`createConfirmBox` → enter feuert, n/esc kehrt ins BEFÜLLTE Formular zurück,
Draft-Erhalt DD2-190). Nach Erfolg springt der Cursor aufs neue Bean (s. Task 4).

**(f) Parent-Picker-Zyklen-Ausschluss: Nachkommen des Beans ausgeblendet.**
`collectDescendants` (BFS, Port beans-src parentpicker.go:233-257) — aber über das
BESTEHENDE `idx.Children` (das IST schon die parent→children-Map, die upstream sich
erst baut). Ausgeschlossen: das Bean selbst, alle Nachkommen, alle Typen außerhalb
`validParentTypes` (Spiegel von beancore `ValidParentTypes`, links.go:446-459:
milestone→kein Parent, epic→[milestone], feature→[milestone,epic],
task/bug→[milestone,epic,feature]). Erste Zeile „(Kein Parent)" (Port
clearParentItem) → `RemoveParent`. Der Blocking-Picker hat KEINEN Zyklen-Ausschluss
(Port-Parität: beans-src blockingpicker.go hat auch keinen; ein Blocking-Zyklus ist
ein logischer PO-Fehler, kein Render-Risiko — bewusstes YAGNI, dokumentiert).

**(g) `B` Blocking-Picker editiert das `Blocking`-Feld des fokussierten Beans
DIREKT.** Verifiziert gegen `beans update --help`: die CLI kennt `--blocking`/
`--remove-blocking` als eigene Flags — Blocking ist beidseitig mutierbar, KEIN rein
server-berechneter Rückwärts-Index. `mutations.go` bekommt `AddBlocking`/
`RemoveBlocking` (Spiegel von AddBlockedBy/RemoveBlockedBy). Es wird also EIN Bean
mit EINEM ETag mutiert (das fokussierte), nie N fremde Beans.

**(h) `e` und `ctrl+e` teilen `keys.Editor`, verzweigen per `msg.String()`.**
keymap.go bindet beide auf EIN Binding (design-spec §7 „e/ctrl+e Editor") — `e` öffnet
das Title-Edit-Formular (einfeldrig, KEIN Confirm-Gate: Edits feuern direkt, Port devd
isCreateKind-Ausschluss), `ctrl+e` den `$EDITOR` auf dem Body. Body-Schreibweg ist
`SetBody` (voller Replace via `--body` — verifiziert: `-d/--body` = „New body";
`AppendBody`/`--body-append` ist additiv und für den Editor-Roundtrip ungeeignet).

---

## Geteilte Infrastruktur (einmal in Task 1, danach nur konsumiert)

Neue model-Felder (types.go, I01-Doc-Konvention fortgeschrieben):

```go
// E3 (bean bt-gzcu): node-action overlays -- mutually exclusive by
// construction (ONE enum field, not 6 bools; filterOpen/confirmQuit predate
// this and are deliberately not retrofitted). mutTarget is the bean the open
// overlay acts on, captured at open time; the ETag is NEVER captured -- every
// submit re-reads m.idx.ByID[mutTarget].ETag (beanETag) so watch-reloads
// between open and submit are automatically honored (design decision d).
overlay   overlayID
mutTarget string
menu      listState // shared cursor for value menu / pickers (one open at a time)

// Value-Menü (T1) / Picker (T2/T3) row lists, built at open time:
menuItems []valueMenuItem // T1
tagItems  []tagCount      // T2
tagPending, tagOriginal map[string]bool // T2 (copy-on-write NOT needed: replaced wholesale at open)
tagInput       textinput.Model // T2 Freitext-Neuanlage
tagInputActive bool
tagInputErr    string
parentItems    []pickerItem // T3 (id "" = "(Kein Parent)")
blockPending, blockOriginal map[string]bool // T3
blockItems     []pickerItem

// huh-Form-Hosting (T4/T5, Port devd forms_shared.go):
form     *huh.Form
formKind string // "create" | "editTitle"
// D01 (Plan-Hygiene, E3-T4-Review PFLICHT, closed in T5, bean bt-sl45,
// ERRATUM): this sketch originally also listed a `createConfirm bool`
// field here -- struck. Design decision a2 is unambiguous that the E3
// overlays, INCLUDING Create-Confirm, are ONE model.overlay overlayID
// enum, not "6 weitere Bools neben filterOpen/confirmQuit"; Task 4's own
// Step 4 pseudocode already agreed (`overlay = overlayCreateConfirm`,
// never `createConfirm = true`), and the real implementation
// (internal/tui/types.go) carries its own matching ERRATUM doc-stamp.
// This was a stale holdover from an earlier draft of this sketch, not a
// second, independent field.
pendingCreate tea.Cmd
createLabel   string
createDraft   *beanDraft

// Delete (T6):
delTitle    string
delChildren int

// Editor (T5): target captured before suspend (cursor may move never during
// ExecProcess, but keep it explicit)
editorTarget string
```

```go
type overlayID int

const (
    overlayNone overlayID = iota
    overlayValueMenu
    overlayTagPicker
    overlayParentPicker
    overlayBlockingPicker
    overlayCreateConfirm
    overlayDeleteConfirm
)
```

Mutation-Infra (messages.go — Msg-Typen + Cmd-Producer ONLY, Port-Konvention):

```go
// mutationDoneMsg carries any mutation's outcome. Success and failure BOTH
// trigger an unconditional reload (applyMutationResult): success must show
// the new state, an ErrConflict must resolve the stale index (design
// decision d). No per-mutation Msg types -- every setter goes through this.
type mutationDoneMsg struct{ err error }

func mutateCmd(fn func() error) tea.Cmd {
    return func() tea.Msg { return mutationDoneMsg{err: fn()} }
}

// createDoneMsg is the ONE exception (Create needs the new bean back for the
// cursor jump, design decision e).
type createDoneMsg struct {
    bean data.Bean
    err  error
}

func createCmd(c *data.Client, opts data.CreateOpts) tea.Cmd {
    return func() tea.Msg {
        b, err := c.Create(opts)
        return createDoneMsg{bean: b, err: err}
    }
}
```

Update-Handler (update.go):

```go
case mutationDoneMsg:
    return m.applyMutationResult(msg.err)
case createDoneMsg:
    // err path == applyMutationResult; success path additionally jumps the
    // cursor (see Task 4 Step 5).
```

```go
// applyMutationResult: statuszeile + IMMER reload (design decision d).
func (m model) applyMutationResult(err error) (tea.Model, tea.Cmd) {
    m.err = ""
    if err != nil {
        if errors.Is(err, data.ErrConflict) {
            m.err = "Konflikt: Bean extern geändert — neu geladen"
        } else {
            m.err = err.Error()
        }
    }
    return m, loadCmd(m.client)
}

// beanETag reads the CURRENT etag for id from the live index (never a
// captured copy -- design decision d). ok=false: bean vanished (deleted by
// an agent between open and submit) -- callers close their overlay and
// surface a status-line note instead of firing a doomed mutation.
func (m model) beanETag(id string) (etag string, ok bool)
```

keyNodeAction-Dispatch (update.go, in `handleKey` NACH dem Refresh-Check, VOR dem
`m.detailFocus`-Dispatch — Node-Aktionen müssen auch im Detail-Fokus wirken):

```go
if handled, nm, cmd := m.keyNodeAction(msg); handled {
    return nm, cmd
}
```

```go
// keyNodeAction routes the node-focused mutation keys (design-spec §7:
// s/t/a/B/c/d/e). Everything except Create requires a focused bean
// (focusedBean() != nil -- covers Tree, Backlog AND detail focus, the same
// view-agnostic dispatcher E2 Task 2 built); Create works on an empty repo.
// Returns handled=false for every other key so the caller falls through.
func (m model) keyNodeAction(msg tea.KeyMsg) (bool, tea.Model, tea.Cmd)
```

Capture-Reihenfolge in `handleKey` (neue Zeilen kursiv beschrieben): `confirmQuit` →
`searchActive` → `filterOpen` → *`m.form != nil` (keyForm, T4)* → *`m.overlay !=
overlayNone` (per-Overlay-Handler-Switch)* → globaler switch (ctrl+c/q/tab) →
Refresh → *keyNodeAction* → detailFocus → keyBacklog/keyTree.

Overlay-Compositing: `viewBrowseRepo` und `viewBacklog` duplizieren heute den
`filterOpen`/`confirmQuit`-Block — Task 1 extrahiert `m.composeOverlays(out, w, h)`
(EIN Ort, beide Views rufen ihn; jede Folge-Task ergänzt nur dort):

```go
func (m model) composeOverlays(out string, w, h int) string {
    if m.filterOpen { out = placeOverlay(out, m.treeFilterBox(), w, h) }
    switch m.overlay {
    case overlayValueMenu:      out = placeOverlay(out, m.valueMenuBox(), w, h)
    // T2/T3/T4/T6 ergänzen ihre Cases hier
    }
    if m.form != nil { out = placeOverlay(out, m.formChrome(), w, h) }
    if m.confirmQuit { out = placeOverlay(out, m.quitBox(), w, h) } // quit zuletzt (oberste Priorität)
    return out
}
```

Keymap: ALLE benötigten Bindings existieren bereits (Status/TagAssign/Assign/
Blocking/Create/Delete/Editor, keymap.go:91-97, inkl. helpGroups-Eintrag) — E3 fügt
KEIN neues Binding hinzu; `TestHelpGroupsCoverEveryBindingExactlyOnce` bleibt
unberührt grün. Footer-Hints (`localHint` in viewBrowseRepo/viewBacklog) werden in T1
um `keys.Status` und in T6 final um Create/Delete/Editor ergänzt (renderBindings —
Single-Source, driftet nie).

---

## Task 1: Status/Type/Prio-Menü + B01-Fix + Enum-Single-Source + Mutation-Infra (`bt-dlgk`)

Legt die komplette geteilte Infrastruktur (oben) und verdrahtet die ERSTE Mutation
Ende-zu-Ende: `s` → Value-Menü → enter → `SetStatus`/`SetType`/`SetPriority` →
`mutationDoneMsg` → Statuszeile + Reload. Beginnt mit dem kleinsten Schritt: B01.

**Files:**
- Modify: `internal/tui/view_browse_backlog.go` (B01: FilterClear-Case in keyBacklog)
- Modify: `internal/data/index.go` (Value-Slices als Quelle der Order-Maps + Export)
- Modify: `internal/tui/box_filter_facets.go` (buildFilterItems auf Single-Source)
- Create: `internal/tui/box_menu_value.go` (`valueMenuItem`, `openValueMenu`,
  `keyValueMenu`, `valueMenuBox`)
- Modify: `internal/tui/types.go` (overlayID-Enum + neue Felder, I01-Doc)
- Modify: `internal/tui/messages.go` (mutationDoneMsg/mutateCmd/createDoneMsg/createCmd)
- Modify: `internal/tui/update.go` (Update-Cases, applyMutationResult, beanETag,
  keyNodeAction, handleKey-Capture, composeOverlays-Aufrufe)
- Modify: `internal/tui/view_browse_repo.go` + `view_browse_backlog.go`
  (composeOverlays-Extraktion, localHint + keys.Status)
- Test: `internal/tui/view_browse_backlog_test.go` (erweitert),
  `internal/data/index_test.go` (erweitert), `internal/tui/box_menu_value_test.go`,
  `internal/tui/update_test.go` (erweitert)

**Port-Referenzen:**
- Enter-wendet-sofort-an + `(current)`-Markierung: beans-src `statuspicker.go:85-139`
  (newStatusPickerModel: selectedIndex auf aktuellem Wert) + `:158-172`
  (enter → selectedMsg, esc → close). typepicker/prioritypicker sind strukturgleich —
  hier zu EINEM Menü mit Gruppen konsolidiert (Design-Entscheidung a3).
- Gruppen-Header im Menü: eigenes `box_filter_facets.go:255-277` (treeFilterBox,
  facetHead-Muster) — 1:1 wiederverwendbares Render-Muster.
- Statuszeilen-Slot: eigenes `view.go:63-83` (statusBar, Red-Slot `m.err`).

**Schritte:**

- [ ] **Step 1 (B01): Failing test** — `view_browse_backlog_test.go`:

```go
func TestKeyBacklogFilterClearResetsFacets(t *testing.T) {
    m := fixtureModel(t, fixtureBeans())
    m.view = viewBacklog
    m.filterStatus = map[string]bool{"todo": true}
    nm := step(t, m, runeMsg('X'))
    if nm.filterActive() {
        t.Fatal("X in Backlog view must clear all facets (B01, keyTree parity)")
    }
}
```

- [ ] **Step 2:** `command go test ./internal/tui/ -run TestKeyBacklogFilterClear` →
  FAIL (X fällt in keyBacklog durch bis zum nav-switch, Facetten bleiben).
- [ ] **Step 3 (B01): Implement** — `keyBacklog` (view_browse_backlog.go:279ff)
  bekommt zwischen dem `keys.Filter`- und dem `keys.Backlog`-Case:

```go
case keybind.Matches(msg, keys.FilterClear):
    // B01 (E2-Abschluss-Übernahme, Epic-Body bt-gzcu): X-Direct-Clear wirkte
    // nur in keyTree (update.go) -- gleicher clearFacets()-Helper, danach
    // Längen-Resync gegen die JETZT breitere Sicht.
    m = m.clearFacets()
    m.backlogList.setLen(len(m.backlogVisible()))
    return m, nil
```

- [ ] **Step 4 (Enums): Failing tests** — `internal/data/index_test.go`:

```go
func TestStatusValuesCanonicalTierOrder(t *testing.T) {
    want := []string{"in-progress", "todo", "draft", "completed", "scrapped"}
    if got := StatusValues(); !reflect.DeepEqual(got, want) {
        t.Fatalf("StatusValues() = %v, want %v", got, want)
    }
}
func TestTypeValuesCanonicalTierOrder(t *testing.T) {
    want := []string{"milestone", "epic", "bug", "feature", "task"}
    if got := TypeValues(); !reflect.DeepEqual(got, want) { t.Fatalf("...") }
}
func TestPriorityValuesCanonicalTierOrder(t *testing.T) {
    want := []string{"critical", "high", "normal", "low", "deferred"}
    if got := PriorityValues(); !reflect.DeepEqual(got, want) { t.Fatalf("...") }
}
func TestValueSlicesAreDefensiveCopies(t *testing.T) {
    StatusValues()[0] = "mutated"
    if StatusValues()[0] != "in-progress" { t.Fatal("caller mutation leaked") }
}
func TestOrderMapsDerivedFromValueSlices(t *testing.T) {
    // Single-Source-Guard: die Order-Maps MÜSSEN aus den Slices abgeleitet
    // sein -- Position im Slice == Rank.
    for i, s := range StatusValues() {
        if StatusRank(s) != i { t.Fatalf("StatusRank(%q)=%d, want %d", s, StatusRank(s), i) }
    }
}
```

- [ ] **Step 5:** `command go test ./internal/data/` → FAIL.
- [ ] **Step 6 (Enums): Implement** — `index.go`: die drei Order-Maps werden durch
  geordnete Slices + init-Ableitung ersetzt (Werte/Reihenfolge UNVERÄNDERT — kein
  Golden-Drift):

```go
// statusValues/typeValues/priorityValues are the canonical, ORDERED enum
// single source (design decision b, bean bt-dlgk): beans 0.4.2's fixed enums
// in tier order. The rank maps below are DERIVED from these slices (position
// == rank) -- box_filter_facets.go's buildFilterItems and E3's value menu
// both consume the exported accessors instead of re-hardcoding the values.
var (
    statusValues   = []string{"in-progress", "todo", "draft", "completed", "scrapped"}
    typeValues     = []string{"milestone", "epic", "bug", "feature", "task"}
    priorityValues = []string{"critical", "high", "normal", "low", "deferred"}

    statusOrder   = rankMap(statusValues)
    typeOrder     = rankMap(typeValues)
    priorityOrder = rankMap(priorityValues)
)

func rankMap(vals []string) map[string]int {
    m := make(map[string]int, len(vals))
    for i, v := range vals { m[v] = i }
    return m
}

func StatusValues() []string   { return append([]string(nil), statusValues...) }
func TypeValues() []string     { return append([]string(nil), typeValues...) }
func PriorityValues() []string { return append([]string(nil), priorityValues...) }
```

  `box_filter_facets.go` buildFilterItems: die 15 hartkodierten Zeilen werden durch
  drei Schleifen über `data.StatusValues()`/`TypeValues()`/`PriorityValues()` ersetzt
  (Label = Wert für status/priority; Type-Label bleibt kapitalisiert → kleine
  `strings.Title`-freie Helper-Map ODER Labels einfach lowercase vereinheitlichen —
  ENTSCHEIDUNG: lowercase, konsistent mit status/priority; der bestehende
  Filter-Menü-Test wird angepasst, KEIN Golden betroffen, treeFilterBox hat keins).
- [ ] **Step 7:** `command go test ./internal/data/ ./internal/tui/` → PASS
  (Filter-Menü-Tests ggf. auf lowercase-Labels angepasst).
- [ ] **Step 8 (Infra+Menü): Failing tests** — `box_menu_value_test.go` +
  `update_test.go`:

```go
func TestOpenValueMenuBuildsGroupedItemsCursorOnCurrentStatus(t *testing.T) {
    // s auf tsk1 (status todo): 15 Items (5+5+5), Cursor auf Index 1
    // (todo, Gruppe status), mutTarget == "tsk1".
}
func TestValueMenuEnterAppliesCursoredValueAndCloses(t *testing.T) {
    // Cursor auf "in-progress" -> enter: overlay == overlayNone, cmd != nil;
    // cmd ausführen (kein beans-Binary: Client zeigt auf nicht-existentes
    // Repo-Dir) -> mutationDoneMsg mit err != nil kommt zurück -- der
    // DISPATCH (Set-Aufruf je Gruppe) wird über die Fehlermeldung des echten
    // data.Client verifiziert (enthält "beans update"), nicht gemockt.
}
func TestValueMenuEscClosesWithoutMutation(t *testing.T)
func TestValueMenuCurrentValueMarked(t *testing.T) // "(current)" in valueMenuBox-Output
func TestKeyNodeActionRequiresFocusedBeanExceptCreate(t *testing.T) {
    // Cursor auf orphan-Root (focusedBean nil): s/t/a/B/d/e -> handled aber
    // no-op (Statuszeilen-Hinweis), c -> öffnet (T4: bis dahin no-op-stub).
}
func TestApplyMutationResultConflictSetsStatusLineAndReloads(t *testing.T) {
    err := fmt.Errorf("%w: bean x: boom", data.ErrConflict)
    nm, cmd := m.applyMutationResult(err)
    // m.err enthält "Konflikt", cmd != nil (loadCmd)
}
func TestApplyMutationResultSuccessClearsErrAndReloads(t *testing.T)
func TestBeanETagReadsLiveIndex(t *testing.T) // reload zwischen open und submit -> neuer ETag
func TestValueMenuTargetVanishedClosesGracefully(t *testing.T) {
    // Bean zwischen open und enter aus dem Index verschwunden (Reload mit
    // gelöschtem Bean): enter -> overlay zu, m.err gesetzt, KEIN cmd.
}
```

- [ ] **Step 9:** `command go test ./internal/tui/` → FAIL (alles undefined).
- [ ] **Step 10: Implement** — types.go (Enum+Felder), messages.go (Msg/Cmd),
  update.go (Cases, applyMutationResult, beanETag, keyNodeAction, Capture-Zeile,
  composeOverlays-Extraktion in beide Views), box_menu_value.go:

```go
// valueMenuItem: eine Zeile des kombinierten s-Menüs.
type valueMenuItem struct{ group, value string } // group: "status"|"type"|"priority"

func (m model) openValueMenu() model {
    b := m.focusedBean() // Guard liegt in keyNodeAction
    m.mutTarget = b.ID
    m.menuItems = buildValueMenuItems() // data.StatusValues()+TypeValues()+PriorityValues()
    m.menu = listState{}
    m.menu.setLen(len(m.menuItems))
    m.menu.cursor = /* Index des aktuellen Status von b (Gruppe status) */
    m.overlay = overlayValueMenu
    return m
}

// keyValueMenu: up/down bewegen (navKey), enter wendet an + schließt
// (mutateCmd + SetStatus/SetType/SetPriority je group, ETag via beanETag),
// esc/s schließt ohne Mutation.
// valueMenuBox: modalPanel "Set value", Gruppen-Header wie treeFilterBox,
// "(current)"-Suffix (theme.Muted) am jeweils aktuellen Wert je Gruppe,
// Zeilen via menuList, Breite clampModalWidth(40, m.width).
```

  Dispatch in keyNodeAction: `keys.Status` → `m.openValueMenu()`. Die übrigen Keys
  (t/a/B/c/d/e) sind in T1 handled-but-stub (Statuszeilen-Hinweis „kommt in Task N"
  ODER schlicht no-op — ENTSCHEIDUNG: no-op ohne Hinweis, die Tasks folgen
  unmittelbar; ein User-facing „not yet"-Text wäre Wegwerf-Arbeit).
- [ ] **Step 11:** `command go test ./internal/tui/` → PASS.
- [ ] **Step 12:** `command go test ./... && command gofmt -l . && command go vet ./...`
  clean; `command go build -o bin/bt .` ok. Manueller Smoke: `s` auf einem Task-Bean
  im eigenen Repo, Status in-progress setzen, Tree aktualisiert sich (Watch/Reload).
- [ ] **Step 13: Commit**

```
feat(tui): Status/Type/Prio-Menü (s) + Mutation-Infra + B01 X-Fix

EIN kombiniertes Value-Menü statt drei Picker (design-spec §7 reserviert
einen Key; Gruppen-Header nach dem Filter-Menü-Muster). Erste Mutation
Ende-zu-Ende: mutateCmd/mutationDoneMsg/applyMutationResult mit
unconditional Reload, ErrConflict -> Statuszeile (Toast ist E5). ETag wird
bei Submit frisch aus dem Index gelesen, nie beim Öffnen eingefroren.
Enum-Single-Source: data.StatusValues/TypeValues/PriorityValues (Order-Maps
jetzt abgeleitet), box_filter_facets-Duplikat entfernt. B01 aus E2:
X/FilterClear wirkt jetzt auch in der Backlog-View.

Refs: bt-dlgk
```

---

## Task 2: Tag-Picker (`bt-8v69`)

`t` → Toggle-Multi-Select über alle vorhandenen Tags (Nutzungszähler, count desc →
alpha) + Freitext-Neuanlage-Zeile. Enter diff't gegen den Original-Zustand →
`AddTag`/`RemoveTag` je Änderung; esc verwirft. Erste Task, die das
Pending-Diff-Muster nutzt (T3-Blocking übernimmt es).

**Files:**
- Create: `internal/tui/box_picker_tag.go` (`tagCount`, `collectTagCounts`,
  `openTagPicker`, `keyTagPicker`, `tagPickerBox`)
- Modify: `internal/data/bean.go` oder neues `internal/data/tags.go`
  (`ValidTagName` + Regex — ENTSCHEIDUNG: `tags.go`, eigene Datei, bean.go bleibt
  reiner JSON-Contract)
- Modify: `internal/tui/types.go` (tagItems/tagPending/tagOriginal/tagInput/
  tagInputActive/tagInputErr), `update.go` (Dispatch), composeOverlays-Case
- Test: `internal/data/tags_test.go`, `internal/tui/box_picker_tag_test.go`

**Port-Referenzen:**
- Zähler + Sortierung: beans-src `tagpicker.go:64-96` (sort count desc, dann alpha).
- Pending-Toggle + Enter-Diff: beans-src `blockingpicker.go:197-245` (space toggelt
  pending, enter berechnet toAdd/toRemove gegen original, esc verwirft) — auf Tags
  übertragen. Das eigene Filter-Menü (`keyFilterMenu`) ist bewusst NICHT das Muster:
  dessen enter schließt nur (Filter wirken live); hier wird echt mutiert, also
  Confirm-Semantik.
- Text-Capture-Submodus: eigenes `keySearchInput`-Muster (update.go:414-435,
  textinput fokussiert = alle Keys gehören dem Input, enter/esc beenden).

**Schritte:**

- [ ] **Step 1: Failing tests** — `tags_test.go`:

```go
func TestValidTagName(t *testing.T) {
    valid := []string{"a", "abc", "a1", "to-review", "a-b-c", "x9-y"}
    invalid := []string{"", "A", "1a", "-a", "a-", "a--b", "a_b", "über", "a b"}
    // je Schleife gegen ValidTagName -- Regex ^[a-z][a-z0-9]*(-[a-z0-9]+)*$
    // (Epic-Body bt-gzcu)
}
```

  `box_picker_tag_test.go`:

```go
func TestCollectTagCountsSortedByCountDescThenAlpha(t *testing.T)
func TestOpenTagPickerSeedsPendingFromFocusedBean(t *testing.T)
func TestTagPickerToggleFlipsPendingOnly(t *testing.T)   // Original unberührt
func TestTagPickerEnterDiffsAddAndRemove(t *testing.T)   // +1 Tag, -1 Tag -> 2 Mutationen (tea.Batch), overlay zu
func TestTagPickerEnterNoChangesNoMutation(t *testing.T) // cmd == nil, overlay zu
func TestTagPickerEscDiscardsPending(t *testing.T)
func TestTagPickerNewTagValidatesRegex(t *testing.T)     // "Über!" -> tagInputErr gesetzt, kein Item
func TestTagPickerNewTagAddsPendingItem(t *testing.T)    // valider Name -> neue Zeile, pending=true, Input zu
```

- [ ] **Step 2:** `command go test ./internal/data/ ./internal/tui/` → FAIL.
- [ ] **Step 3: Implement** — `data/tags.go`:

```go
// tagNameRe mirrors the tag grammar the beans CLI accepts (epic bt-gzcu):
// lowercase alnum segments, single-hyphen-separated, leading letter.
var tagNameRe = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)

func ValidTagName(s string) bool { return tagNameRe.MatchString(s) }
```

  `box_picker_tag.go`: `collectTagCounts(idx)` zählt über `idx.ByID` (Determinismus:
  erst Namen sammeln, dann `sort.Slice` count desc/alpha — KEINE Map-Walk-Ordnung ins
  Ergebnis, gleiche Lehre wie tagFilterOptions' ERRATUM). `openTagPicker` seedet
  `tagOriginal`/`tagPending` aus dem fokussierten Bean (zwei UNABHÄNGIGE Maps —
  wholesale-Replace-Konvention wie `searchBleveIDs`, kein cloneBoolMap-Fall).
  `keyTagPicker`: tagInputActive-Zweig zuerst (alle Keys → tagInput; enter validiert
  → neue Zeile + pending; esc schließt nur den Input); sonst up/down (navKey), space/x
  (`keys.Toggle`) toggelt pending, „n" öffnet den Neuanlage-Input, enter diff't:

```go
var cmds []tea.Cmd
etag, ok := m.beanETag(m.mutTarget)
// ok==false: overlay zu, m.err, return (Muster T1)
for tag := range m.tagPending { if !m.tagOriginal[tag] { cmds = append(cmds, mutateCmd(func() error { return m.client.AddTag(id, tag, etag) })) } }
// ... analog RemoveTag; ACHTUNG Schleifenvariablen-Capture (tag := tag)
// MEHRERE Mutationen auf EINEM ETag: die erste gewinnt, jede weitere läuft
// in ErrConflict -> applyMutationResults Reload heilt das, ABER die
// verlorenen Diffs wären ein stiller Datenfresser. ENTSCHEIDUNG: Diffs
// werden SEQUENTIELL als EIN mutateCmd ausgeführt (ein fn, das die Liste
// abarbeitet und nach jedem Schritt den ETag aus der update-Response NICHT
// hat -- die CLI gibt das Bean samt neuem ETag zurück, aber unsere Wrapper
// verwerfen es). Da mutations.go-Setter kein Bean zurückgeben: fn macht
// AddTag/RemoveTag nacheinander und liest zwischen den Schritten den
// frischen ETag via c.Show? EXISTIERT NICHT als Wrapper. Pragmatische
// Auflösung (dokumentierter Scope-Cut): EIN kombinierter Aufruf ist mit der
// CLI möglich -- `beans update <id> --tag a --remove-tag b` nimmt BEIDE
// Flag-Familien in EINEM Kommando. mutations.go bekommt daher
// SetTags(id string, add, remove []string, etag string) error, das EINEN
// update-Aufruf mit allen --tag/--remove-tag-Flags baut (ein ETag, ein
// Kommando, atomar auf CLI-Ebene). AddTag/RemoveTag bleiben für
// Einzelfälle/E4 bestehen.
```

  **Design-Nachtrag (verbindlich für die Implementierung):** `data.SetTags(id, add,
  remove []string, etag string) error` — EIN `beans update` mit wiederholten
  `--tag`/`--remove-tag`-Flags (CLI verifiziert: beide sind `stringArray`). Test dazu
  in `client_mut_test.go` (`TestSetTagsAddsAndRemovesInOneCall`, requireBeansBinary-
  Muster gegen newTestRepo). Der Picker feuert genau EIN `mutateCmd`.
- [ ] **Step 4:** `command go test ./internal/tui/ ./internal/data/` → PASS.
- [ ] **Step 5:** Voller Gate-Lauf (`go test ./...`, gofmt, vet, build) + Smoke:
  `t` auf einem Bean, Tag togglen + neuen Tag anlegen, enter, `.md`-Datei zeigt beide
  Änderungen in EINEM Schreibvorgang.
- [ ] **Step 6: Commit**

```
feat(tui): Tag-Picker (t) — Zähler, Pending-Diff, Neuanlage, SetTags

Pending-Diff-Muster von beans-src blockingpicker portiert (space toggelt,
enter diff't, esc verwirft). Alle Tag-Änderungen als EIN beans-update
(data.SetTags, --tag/--remove-tag kombiniert) -- ein ETag, keine
Konflikt-Kaskade bei Mehrfach-Diffs. Freitext-Neuanlage regex-validiert
(data.ValidTagName).

Refs: bt-8v69
```

---

## Task 3: Parent-/Blocking-Picker (`bt-p1uz`)

`a` → Parent-Picker (Single-Select, Zyklen-Ausschluss + Typ-Hierarchie); `B` →
Blocking-Picker (Toggle-Multi-Select, Pending-Diff wie T2). Ergänzt die zwei
fehlenden mutations.go-Wrapper `AddBlocking`/`RemoveBlocking`.

**Files:**
- Modify: `internal/data/mutations.go` (AddBlocking/RemoveBlocking — und, Spiegel von
  T2s SetTags-Lehre: `SetBlocking(id, add, remove, etag)` als EIN update-Aufruf;
  Einzel-Wrapper trotzdem ergänzen für Symmetrie zur BlockedBy-Familie)
- Create: `internal/data/hierarchy.go` (`validParentTypes`, `CollectDescendants`)
- Create: `internal/tui/box_picker_parent.go`, `internal/tui/box_picker_blocking.go`
- Modify: `internal/tui/types.go` (parentItems/blockItems/blockPending/blockOriginal),
  `update.go` (Dispatch a/B), composeOverlays-Cases
- Test: `internal/data/hierarchy_test.go`, `internal/data/client_mut_test.go`
  (erweitert), `internal/tui/box_picker_parent_test.go`,
  `internal/tui/box_picker_blocking_test.go`

**Port-Referenzen:**
- Zyklen-Ausschluss: beans-src `parentpicker.go:233-257` (collectDescendants, BFS über
  parent→children) — hier über das BESTEHENDE `idx.Children` (kein Map-Neuaufbau).
- Eligibility-Filter + „(Kein Parent)"-Zeile: `parentpicker.go:95-182`
  (clearParentItem zuerst, selectedIndex auf aktuellem Parent, Typ-Filter).
- Typ-Hierarchie: beancore `links.go:446-459` (ValidParentTypes) — als eigene
  Kopie `validParentTypes` in `data/hierarchy.go` gespiegelt (mit Quell-Verweis;
  beans-src ist Referenz-Klon, kein Import-Pfad).
- Blocking-Pending-Diff: `blockingpicker.go:104-245` (original/pending-Maps,
  space-Toggle, enter-Diff, ●/○-Indikator Rot/Muted).

**Schritte:**

- [ ] **Step 1: Failing tests** — `hierarchy_test.go`:

```go
func TestValidParentTypesMirrorsBeancore(t *testing.T) {
    // milestone -> nil; epic -> [milestone]; feature -> [milestone epic];
    // task, bug -> [milestone epic feature]; unbekannt -> [milestone epic feature]
}
func TestCollectDescendantsBFS(t *testing.T) {
    // m -> e -> t1, t2; CollectDescendants(idx, "m") == {e,t1,t2};
    // CollectDescendants(idx, "t1") leer.
}
func TestCollectDescendantsSurvivesParentCycle(t *testing.T) {
    // a->b->a (Kinder-Zyklus): terminiert, kein Hang (visited-Guard wie
    // beans-src Queue-Dedup).
}
```

  `client_mut_test.go` (requireBeansBinary/newTestRepo-Muster):

```go
func TestAddRemoveBlockingRoundtrip(t *testing.T) // AddBlocking -> show zeigt blocking; RemoveBlocking -> leer
func TestSetBlockingAddsAndRemovesInOneCall(t *testing.T)
```

  `box_picker_parent_test.go`:

```go
func TestParentPickerExcludesSelfDescendantsAndInvalidTypes(t *testing.T) {
    // Epic fokussiert: nur Milestones wählbar; eigene Nachkommen + self raus.
}
func TestParentPickerClearRowFirstAndCursorOnCurrentParent(t *testing.T)
func TestParentPickerMilestoneShowsNoEligibleParents(t *testing.T) {
    // milestone: validParentTypes nil -> nur "(Kein Parent)"-Zeile.
}
func TestParentPickerEnterAppliesSetParentOrRemoveParent(t *testing.T)
func TestParentPickerEscNoMutation(t *testing.T)
```

  `box_picker_blocking_test.go`:

```go
func TestBlockingPickerExcludesOnlySelf(t *testing.T) // KEIN Zyklen-Ausschluss (Design g/f)
func TestBlockingPickerSeedsPendingFromBlockingField(t *testing.T)
func TestBlockingPickerEnterDiffsViaSetBlocking(t *testing.T)
func TestBlockingPickerEscDiscards(t *testing.T)
```

- [ ] **Step 2:** `command go test ./internal/data/ ./internal/tui/` → FAIL.
- [ ] **Step 3: Implement** — `data/hierarchy.go`:

```go
// validParentTypes mirrors beancore.ValidParentTypes (beans-src
// pkg/beancore/links.go:446-459) -- the beans CLI enforces this server-side;
// the picker pre-filters so the PO is never offered a doomed choice.
func validParentTypes(beanType string) []string { /* switch, s.o. */ }

// CollectDescendants returns all bean IDs below id, walking idx.Children
// (which IS the parent->children map beans-src rebuilds ad hoc) breadth-
// first with a visited set (hand-edited frontmatter can hold cycles).
func CollectDescendants(idx *Index, id string) map[string]bool
```

  (`validParentTypes` unexportiert + `EligibleParents(idx, b) []*Bean` als
  exportierte Fassade, die Ausschluss+Typ-Filter+Sortierung (`SortBeans`) bündelt —
  die TUI konsumiert nur die Fassade, Test der Regeln über sie.)
  `mutations.go`: AddBlocking/RemoveBlocking/SetBlocking analog der
  BlockedBy-/SetTags-Familie. `box_picker_parent.go`: `pickerItem{id, label}`,
  Zeile 0 = „(Kein Parent)" (id "" → RemoveParent; theme.Dim), Rest aus
  `data.EligibleParents`, Cursor auf aktuellem Parent (Fallback 0), Zeilen im
  `relationRow`-Stil (Status-Glyph+Typ-Icon+ID+Titel — bestehender Helper,
  view_detail_bean.go:66-68), windowAround bei Überlänge, enter → SetParent/
  RemoveParent sofort + schließen. `box_picker_blocking.go`: pending/original wie T2,
  Indikator `●` (theme.Red) / `○` (theme.Muted) vor der Zeile, enter →
  `SetBlocking`-Diff als EIN mutateCmd.
- [ ] **Step 4:** `command go test ./...` → PASS (Binary-Tests laufen lokal,
  skippen in binary-losen Umgebungen — bestehendes Muster).
- [ ] **Step 5:** gofmt/vet/build + Smoke: `a` auf einem Task → nur Milestones/Epics/
  Features angeboten, eigener Teilbaum fehlt; `B` → Toggle zweier Beans, `.md` zeigt
  `blocking:`-Liste.
- [ ] **Step 6: Commit**

```
feat(tui): Parent-/Blocking-Picker (a/B) — Zyklen-Ausschluss, SetBlocking

Parent-Picker filtert self+Nachkommen (CollectDescendants über idx.Children,
BFS mit visited-Guard) und Typen außerhalb validParentTypes (Spiegel
beancore.ValidParentTypes). Blocking-Picker bewusst OHNE Zyklen-Ausschluss
(Port-Parität beans-src; logischer PO-Fehler, kein Render-Risiko). Blocking
ist direkt mutierbar (--blocking/--remove-blocking, CLI verifiziert) --
AddBlocking/RemoveBlocking/SetBlocking ergänzt, ein Bean, ein ETag.

Refs: bt-p1uz
```

---

## Task 4: Create-Form + Confirm-Gate (`bt-y4ly`)

`c` → huh-Formular (eingebettetes Sub-Modell, KEIN form.Run()) → Confirm-Gate →
`data.Client.Create` → Cursor aufs neue Bean. Legt die Form-Hosting-Infrastruktur
(styleForm/formChrome/paletteFormTheme), die T5 wiederverwendet.

**Files:**
- Modify: `go.mod`/`go.sum` (huh v1.0.0 — Modul-Cache vorhanden:
  `~/go/pkg/mod/github.com/charmbracelet/huh@v1.0.0`, kein Netz nötig)
- Create: `internal/tui/forms_shared.go` (nonEmpty, formInnerWidth/Height, styleForm,
  paletteFormTheme, formChrome, updateForm/keyForm)
- Create: `internal/tui/form_create_bean.go` (beanDraft, buildCreateBeanForm,
  parseTagsField, draftFromForm, createOptsFromDraft)
- Create: `internal/tui/box_confirm_create.go` (submitForm, keyCreateConfirm,
  createConfirmBox)
- Modify: `internal/tui/types.go` (form/formKind/createConfirm/pendingCreate/
  createLabel/createDraft), `update.go` (Capture m.form, createDoneMsg-Case,
  Dispatch c), composeOverlays (formChrome + overlayCreateConfirm)
- Test: `internal/tui/form_create_bean_test.go`, `internal/tui/box_confirm_create_test.go`

**Port-Referenzen:**
- Form-Hosting: devd `forms_shared.go:125-172` (openForm/styleForm — hier auf 2 Kinds
  reduziert), `:167-225` (paletteFormTheme, Theme-Token auf theme.Mauve/Yellow/Base/
  Red/Text/Hint gemappt — Theme-Tokens only, kein Hex hier), `:289-307` (formChrome
  via modalPanel), `:42-47` (nonEmpty), `:100-123` (formInnerWidth/Height).
- Feld-Bau: devd `form_create_issue.go:32-53` (keyed Fields, Value-Prefill für Draft,
  `ExternalEditor(false)` auf Text-Feldern — DD2-234-Falle).
- Confirm-Gate: devd `box_confirm_create.go:33-57` (submitForm: Cmd JETZT bauen, dann
  parken), `:96-120` (keyCreateConfirm: enter feuert, n/esc → Draft-Rückkehr DD2-190),
  `:123-129` (createConfirmBox, Mauve=konstruktiv).
- Cursor-Jump-Bausteine: eigenes `expandAncestorsOf` (update.go:285-295) +
  `applyLoaded`s „exact bean still present"-Pfad (update.go:98-101).

**Schritte:**

- [ ] **Step 1:** `command go get github.com/charmbracelet/huh@v1.0.0` +
  `command go mod tidy`; verify go.mod trägt huh als direkte Dependency.
- [ ] **Step 2: Failing tests** — `form_create_bean_test.go`:

```go
func TestBuildCreateBeanFormPrefillsParentFromCursor(t *testing.T) // GetString("parent")
func TestParseTagsFieldSplitsAndValidates(t *testing.T) {
    // "a b,c" -> [a b c]; "Ü!" -> error mit Feldhinweis (huh Validate-Fehlertext)
}
func TestParentFieldValidatorAcceptsEmptyAndExistingID(t *testing.T)
func TestParentFieldValidatorRejectsUnknownID(t *testing.T)
func TestCreateOptsFromDraftMapsAllFields(t *testing.T)
```

  `box_confirm_create_test.go`:

```go
func TestSubmitFormParksCmdAndOpensConfirm(t *testing.T) {
    // formKind create + ausgefüllter Draft -> submitForm: form nil,
    // overlay==overlayCreateConfirm, pendingCreate != nil, createLabel
    // enthält Typ+Titel, createDraft gesichert.
}
func TestCreateConfirmEnterFiresPendingCmd(t *testing.T)
func TestCreateConfirmEscReturnsToFilledForm(t *testing.T) {
    // n/esc: overlay zu, form != nil, GetString("title") == Draft-Titel (DD2-190).
}
func TestCreateDoneJumpsCursorAndExpandsParentPath(t *testing.T) {
    // createDoneMsg{bean mit Parent epc1}: cursorID == neue ID,
    // expanded[epc1] == true (+ Ancestors), cmd != nil (Reload); nach
    // step(beansLoadedMsg mit neuem Bean) ist der Cursor auf dem neuen Bean.
}
func TestCreateDoneErrRoutesToApplyMutationResult(t *testing.T)
func TestFormCapturesAllKeysWhileOpen(t *testing.T) // "q" tippt, quittet nicht (searchActive-Präzedenz)
```

- [ ] **Step 3:** `command go test ./internal/tui/` → FAIL.
- [ ] **Step 4: Implement** — forms_shared.go (Port wie referenziert; updateForm als
  zentraler Nicht-Key-Router: am ENDE von `Update`s Type-Switch ein
  `if m.form != nil { return m.updateForm(msg) }`-Fallback für huh-interne Msgs
  (cursor.BlinkMsg etc.), Key-Msgs laufen über handleKey → keyForm: esc schließt
  (Formular verwerfen, Create-Kind OHNE Confirm-Umweg — Draft-Erhalt gibt es NUR über
  den n/esc-Pfad des Confirms, devd-Parität), sonst an `m.form.Update` delegieren und
  bei `huh.StateCompleted` → `submitForm()`). form_create_bean.go:

```go
// beanDraft überlebt den Form-Neuaufbau (Confirm-n/esc-Rückkehr, DD2-190).
type beanDraft struct{ title, typ, priority, status, parent, tags, body string }

func buildCreateBeanForm(idx *data.Index, d beanDraft) *huh.Form
// Selects aus data.TypeValues()/PriorityValues()/StatusValues() (Design b);
// Defaults task/normal/todo; parent-Validator gegen idx.ByID; tags-Validator
// via parseTagsField/data.ValidTagName; alle Felder keyed (title/type/
// priority/status/parent/tags/body).
```

  box_confirm_create.go: submitForm liest per GetString → beanDraft → 
  `createCmd(m.client, createOptsFromDraft(d))` parken, `overlay =
  overlayCreateConfirm`. keyCreateConfirm: enter → cmd feuern (+ overlay zu);
  n/esc → `buildCreateBeanForm(idx, *m.createDraft)` neu öffnen.
  createDoneMsg-Handler (update.go):

```go
case createDoneMsg:
    if msg.err != nil { return m.applyMutationResult(msg.err) }
    if msg.bean.Parent != "" {
        m.expanded = expandAncestorsOf(m.idx, m.expanded, msg.bean.Parent)
        exp := cloneBoolMap(m.expanded) // I01: copy-on-write
        exp[msg.bean.Parent] = true     // expandAncestorsOf expandiert Ancestors VON id, nicht id selbst
        m.expanded = exp
    }
    m.cursorID = msg.bean.ID // applyLoaded findet die ID nach dem Reload (US-10-Pfad)
    m.err = ""
    return m, loadCmd(m.client)
```

  Dispatch: keyNodeAction `keys.Create` → `m.openCreateForm()` (Parent-Prefill =
  fokussiertes Bean, sonst leer; funktioniert ohne Fokus/leeres Repo).
- [ ] **Step 5:** `command go test ./internal/tui/` → PASS.
- [ ] **Step 6:** Voller Gate-Lauf + Smoke: `c` auf einem Epic → Formular vorbefüllt
  (parent=Epic-ID), ausfüllen, enter bis Confirm, enter → neues Bean erscheint
  IM Baum unter dem Epic, Cursor sitzt darauf.
- [ ] **Step 7: Commit**

```
feat(tui): Create-Form (c, huh) + Confirm-Gate

huh v1.0.0 als eingebettetes Sub-Modell (kein form.Run()) -- Werte keyed
per GetString nach StateCompleted, kein Pointer-Binding über Model-Copies
(devd forms_shared-Konvention). Confirm-Gate parkt den fertigen createCmd;
n/esc kehrt in das BEFÜLLTE Formular zurück (Draft-Erhalt, DD2-190-Port).
Parent nur existenz-validiert -- Typ-Hierarchie bleibt CLI-Autorität
(VALIDATION_ERROR -> Statuszeile). Nach Create: Ancestor-Pfad expandiert,
Cursor auf dem neuen Bean (applyLoadeds bestehender ID-Restore-Pfad).

Refs: bt-y4ly
```

---

## Task 5: Title-Edit-Form + Body-`$EDITOR` (`bt-sl45`)

`e` → einfeldriges Title-Formular (T4-Infra, KEIN Confirm-Gate); `ctrl+e` →
`$EDITOR`-Suspend auf dem Body (tea.ExecProcess), Rückweg über neues `data.SetBody`
(voller Replace — `--body`, verifiziert; AppendBody wäre additiv/falsch).

**Files:**
- Modify: `internal/data/mutations.go` (SetBody — Spiegel von SetTitle mit `--body`)
- Create: `internal/tui/form_edit_title.go` (buildEditTitleForm + Submit-Zweig in
  submitForm/formKind "editTitle")
- Create: `internal/tui/editor.go` (editorFinishedMsg, editorBinary, prepareEditor,
  readEditorResult, editInEditor)
- Modify: `internal/tui/types.go` (editorTarget), `update.go` (editorFinishedMsg-Case,
  Editor-Dispatch in keyNodeAction per msg.String())
- Test: `internal/data/client_mut_test.go` (erweitert),
  `internal/tui/editor_test.go`, `internal/tui/form_edit_title_test.go`

**Port-Referenzen:**
- devd `editor.go:21-96` VERBATIM (editorFinishedMsg/prepareEditor/readEditorResult/
  editInEditor via tea.ExecProcess) — nur `editorBinary()` abweichend: `$VISUAL` →
  `$EDITOR` → `vi` (Design c; devd `:31-42` bindet an configuredEditor = E5-Settings).
- Einzelfeld-Form: devd `forms_shared.go:335-360` (buildEditFieldForm, Input-Zweig
  mit nonEmpty für title) — hier auf den einen Title-Fall reduziert.
- Kein Confirm für Edits: devd `box_confirm_create.go:22-28` (isCreateKind schließt
  editField aus — Edits feuern direkt).

**Schritte:**

- [ ] **Step 1: Failing test (Datenlayer)** — `client_mut_test.go`:

```go
func TestSetBodyReplacesWholeBody(t *testing.T) {
    // newTestRepo, List -> task-Bean+ETag; SetBody("neu"); Show/List: Body
    // == "neu" (alter Fixture-Body weg -- Replace, nicht Append).
}
```

- [ ] **Step 2:** `command go test ./internal/data/` → FAIL (SetBody undefined).
- [ ] **Step 3: Implement** — mutations.go:

```go
// SetBody replaces a bean's whole body (`--body` -- a FULL replace, verified
// against beans 0.4.2 --help: "New body"; --body-append is additive and
// unsuitable for the $EDITOR round-trip, bean bt-sl45).
func (c *Client) SetBody(id, body, etag string) error {
    return c.update(id, etag, "--body", body)
}
```

- [ ] **Step 4: Failing tests (TUI)** — `editor_test.go`:

```go
func TestEditorBinaryResolvesVisualThenEditorThenVi(t *testing.T) // t.Setenv-Matrix
func TestPrepareEditorWritesInitialContentToTempFile(t *testing.T) // Port devd: testbar ohne tea
func TestReadEditorResultDetectsChangeAndCleansUp(t *testing.T)    // changed-Flag + os.Remove
func TestEditorFinishedUnchangedFiresNoMutation(t *testing.T)      // changed=false -> cmd nil, kein SetBody
func TestEditorFinishedChangedFiresSetBodyWithFreshETag(t *testing.T)
func TestEditorFinishedTargetVanishedSurfacesError(t *testing.T)
```

  `form_edit_title_test.go`:

```go
func TestEditTitleFormPrefilledAndNonEmptyValidated(t *testing.T)
func TestEditTitleSubmitFiresSetTitleDirectlyNoConfirm(t *testing.T)
```

- [ ] **Step 5:** `command go test ./internal/tui/` → FAIL.
- [ ] **Step 6: Implement** — editor.go (Port, editorBinary env-Kaskade),
  keyNodeAction-Editor-Zweig:

```go
case keybind.Matches(msg, keys.Editor):
    b := m.focusedBean() // Guard bereits oben
    if msg.String() == "ctrl+e" { // Design h: ein Binding, zwei Wege
        m.editorTarget = b.ID
        return true, m, editInEditor(b.Body, ".md")
    }
    return true, m2, cmd // "e": openEditTitleForm(b)
```

  editorFinishedMsg-Case (update.go): err → m.err; !changed → no-op; sonst
  `beanETag(m.editorTarget)` → `mutateCmd(SetBody)` (ok=false → m.err, kein Cmd).
  KORRIGIERT (Review-Runde 2, F2, bean bt-sl45): dieser ursprüngliche Plan-Satz
  war falsch für den `ctrl+e`-Pfad — ein `$EDITOR`-Suspend kann lange dauern,
  ein Watch-Reload währenddessen rotiert `m.idx`s ETag unter der Hand, und ein
  frischer `beanETag`-Read bei Submit hätte den externen Edit dann still
  überschrieben (Lost-Update, kein Konflikt ausgelöst). Tatsächlich
  implementiert: der etag wird in `keyNodeAction`s `ctrl+e`-Zweig BEIM ÖFFNEN
  eingefroren (`m.editorETag = b.ETag`), `applyEditorFinished` liest bei
  Submit NUR NOCH `beanETag`s `ok`-Bool (Bean-Präsenz), nicht mehr dessen
  etag-Wert. `submitForm("editTitle")` bleibt unverändert bei "frisch bei
  Submit" (Design-Entscheidung d gilt dort unverändert weiter).
  form_edit_title.go: formKind "editTitle", submitForm-Zweig feuert
  `mutateCmd(func() error { return m.client.SetTitle(id, title, etag) })` DIREKT
  (kein Confirm — isCreateKind-Analogon: nur "create" gated).
- [ ] **Step 7:** `command go test ./internal/tui/ ./internal/data/` → PASS.
- [ ] **Step 8:** Voller Gate-Lauf + Smoke (tmux): `ctrl+e` auf einem Bean → Editor
  öffnet mit Body, ändern+speichern → Detail-Pane zeigt neuen Body; `e` → Titel
  ändern → Tree-Zeile aktualisiert.
- [ ] **Step 9: Commit**

```
feat(tui): Title-Edit (e) + Body-$EDITOR (ctrl+e) + data.SetBody

tea.ExecProcess-Suspend (devd editor.go verbatim), editorBinary löst
$VISUAL -> $EDITOR -> vi (design-spec §7 wörtlich "$EDITOR"; devds
configuredEditor setzt die E5-Settings voraus). Body-Rückweg über neues
SetBody (--body, voller Replace) -- AppendBody wäre additiv und für den
Editor-Roundtrip falsch. Edits feuern ohne Confirm-Gate (devd
isCreateKind-Parität); unveränderter Editor-Inhalt mutiert nicht.

Refs: bt-sl45
```

---

## Task 6: Delete-Confirm + ETag-Konflikt-Sweep (`bt-ppzb`)

> **ERRATUM (E3-T6-Review PFLICHT-Finding I01, bean bt-qzwt):** die Skizze
> unten (Zeilen ~1037-1124, insb. „(verwaist)"-Root / „N Kinder werden
> verwaist") ist STALE — empirisch widerlegt während T6s tmux-Smoke + einem
> isolierten `beans create`/`beans delete`-Probe (beide reproduzierbar).
> `beans delete` lässt Kinder NICHT mit einem hängenden `parent:`-Feld
> zurück; die reale beans-0.4.2-CLI entfernt das Feld aktiv — Kinder werden
> zu gewöhnlichen ROOTS (`idx.Roots()`), nicht in eine „(verwaist)"-Bucket
> einsortiert. Quelle der Wahrheit: Doc-Stamp
> `internal/tui/box_confirm_delete.go` + Regressionstest
> `internal/data/client_mut_test.go:TestDeleteClearsFormerChildrensParentField`.
> Diese Sektion wird NICHT rückwirkend umgeschrieben (Plan-Historie bleibt
> nachvollziehbar als Entwurfs-Snapshot) — nur dieser Pointer ist verbindlich.

`d` → Delete-Confirm mit Kinder-Count-WARNUNG (Semantik-Abweichung von devd:
`beans delete` kaskadiert NICHT — Kinder verwaisen und landen in der bestehenden
„(verwaist)"-Root, mutations.go-Delete-Doc; die Preview sagt „N Kinder werden
verwaist", nicht „werden mitgelöscht"). Danach der Cross-Cutting-Sweep: je
Mutations-Stelle aus T1-T5 ein ETag-Konflikt-Regressionstest.

**Files:**
- Create: `internal/tui/box_confirm_delete.go` (openDeleteConfirm, keyDeleteConfirm,
  deleteBox)
- Modify: `internal/tui/types.go` (delTitle/delChildren), `update.go` (Dispatch d),
  composeOverlays, Footer-Hints final (localHint + Create/Delete/Editor)
- Test: `internal/tui/box_confirm_delete_test.go`,
  `internal/tui/etag_conflict_test.go` (der Sweep)

**Port-Referenzen:**
- devd `box_confirm_delete.go:57-126` (keyDelete enter/esc/n, deleteBox Rot=
  destruktiv, modalBox theme.Red) — OHNE loadDeletePreview (`idx.Children[id]` ist
  synchron im Speicher, kein async Count-Load) und OHNE removeIssueFromCaches
  (beans-tui cached nicht per-View: der unconditional Reload IST die
  Cache-Invalidierung).
- Cursor nach Delete: eigener `applyLoaded`-oldPos-Fallback (update.go:85-107) —
  Wiederverwendung, KEINE neue Cursor-Logik.

**Schritte:**

- [ ] **Step 1: Failing tests** — `box_confirm_delete_test.go`:

```go
func TestOpenDeleteConfirmCountsChildrenSynchronously(t *testing.T) // epc1 -> 2 Kinder
func TestDeleteBoxWarnsChildrenWillBeOrphaned(t *testing.T) // "verwaist" im Text, NICHT "gelöscht"
func TestDeleteConfirmEnterFiresDeleteAndCloses(t *testing.T) // Delete OHNE etag (CLI hat kein --if-match für delete)
func TestDeleteConfirmEscCancels(t *testing.T)
func TestDeleteCursorClampsViaApplyLoadedOldPos(t *testing.T) {
    // Cursor auf letztem Bean, Reload ohne dieses Bean -> Cursor klemmt auf
    // den Nachbarn (bestehender oldPos-Pfad, hier nur end-to-end verifiziert).
}
```

  `etag_conflict_test.go` — der Sweep (Tabelle, subtests):

```go
// TestEtagConflictSweep simulates a stale-etag conflict for EVERY mutation
// site built in T1-T5 and asserts the ONE shared contract (design decision
// d): mutationDoneMsg{ErrConflict} -> status line contains "Konflikt",
// reload cmd fired, no panic, overlay/form state sane.
// Sites: value-menu status / value-menu type / value-menu priority /
// tag-picker SetTags / parent-picker SetParent / parent-picker RemoveParent /
// blocking-picker SetBlocking / edit-title SetTitle / editor SetBody.
// (Create braucht keinen: kein ETag; Delete: CLI-delete hat kein --if-match.)
// KORRIGIERT (Review-Runde 2, F2, bean bt-sl45): "editor SetBody" folgt NICHT
// mehr Design-Entscheidung d wie die übrigen Sites -- applyEditorFinished
// nutzt den bei ctrl+e-Open eingefrorenen m.editorETag, nicht mehr einen bei
// Submit frisch gelesenen beanETag-Wert. Der Sweep-Subtest für "editor
// SetBody" muss das etag also beim SIMULIERTEN OPEN veralten lassen, nicht
// erst beim Submit -- ein Regressionstest für genau dieses Verhalten
// existiert bereits (editor_test.go,
// TestEditorFinishedUsesEtagCapturedAtOpenNotFreshIndexRead); T6 kann ihn als
// Vorlage für den Sweep-Subtest nehmen statt das generische Site-Pattern
// unverändert zu portieren.
func TestEtagConflictSweep(t *testing.T) { /* je Site ein t.Run */ }
func TestConflictAfterWatchReloadUsesFreshETagNoConflict(t *testing.T) {
    // Positiv-Gegenprobe zu Design d: Reload zwischen Öffnen und Submit ->
    // beanETag liefert den NEUEN ETag, gar kein Konflikt erst entsteht.
    // KORRIGIERT (Review-Runde 2, F2): gilt NICHT für die editor-Site -- dort
    // ist ein Reload zwischen ctrl+e-Open und Submit jetzt GENAU der Fall,
    // der einen (gewollten, sichtbaren) Konflikt auslösen soll (siehe oben).
}
```

- [ ] **Step 2:** `command go test ./internal/tui/` → FAIL.
- [ ] **Step 3: Implement** — box_confirm_delete.go (deleteBox: Kopf Rot+Bold
  „Delete <type>", Titel, `N Kind(er) werden verwaist — bleiben unter „(verwaist)"
  erhalten` bei delChildren>0, `Irreversible.` Rot, Footer enter/esc·n; Breite
  clampModalWidth(48, m.width), Border theme.Red). keyDeleteConfirm: enter →
  `mutateCmd(func() error { return m.client.Delete(id) })` + overlay zu; esc/n zu.
  Footer-Hints: localHint beider Views final um keys.Create/Delete/Editor ergänzt.
- [ ] **Step 4:** `command go test ./internal/tui/` → PASS.
- [ ] **Step 5:** `command go test ./... -count=1` (2×) grün, gofmt/vet leer,
  `command go build -o bin/bt .` ok. Smoke: `d` auf einem Bean mit Kindern → Warnung
  zeigt Count; Löschen → Kinder erscheinen unter „(verwaist)".
- [ ] **Step 6: Commit**

```
feat(tui): Delete-Confirm (d) + ETag-Konflikt-Sweep

Kinder-Count synchron aus idx.Children (kein async Preview-Load wie devd)
und als VERWAISUNGS-Warnung formuliert: beans delete kaskadiert nicht,
Kinder landen in der bestehenden (verwaist)-Root. Cursor-Klemmung nach
Delete über applyLoadeds bestehenden oldPos-Pfad -- keine neue Logik.
Sweep-Regressionstests: jede Mutations-Stelle aus T1-T5 erfüllt den einen
Konflikt-Contract (Statuszeile + unconditional Reload).

Refs: bt-ppzb
```

---

## Task 7: E3-Abschluss (`bt-qzwt`)

Mirrors E2 Task 6 (implementation-plan.md »Epos-Rituale« → Epos-Abschluss).

- [ ] `command go test ./... -count=1` grün (2× hintereinander), `command go build
  -o bin/bt .` ok, `command gofmt -l .` leer, `command go vet ./...` leer.
- [ ] Manueller Dogfooding-Smoke (tmux): im eigenen Repo `s` Status setzen → Reload
  sichtbar; `t` Tag togglen+anlegen; `a` Parent umziehen (Nachkomme wird nicht
  angeboten); `B` Blocking setzen; `c` Bean unter Epic anlegen → Cursor springt;
  `e` Titel ändern; `ctrl+e` Body im Editor; `d` mit Kinder-Warnung. Parallel eine
  `.md` von Hand ändern (ETag altern lassen) und eine Mutation feuern → Statuszeile
  zeigt Konflikt, Reload heilt. Terminal-Ausschnitt als Beleg in den Commit-Body.
- [ ] beans pflegen: `beans update bt-dlgk -s completed` … `bt-ppzb -s completed`
  (T1-T6, agent-abschließbar), Epic `beans update bt-gzcu --tag to-review`
  (**NICHT** `-s completed` — PO-Gate, „der ausführende Agent schließt NICHT").
- [ ] Selbst-Review im Commit-Body:
  - Spec-Coverage: V7 komplett (Create/Edit, Confirm-Gate, `$EDITOR`) ✓ · V8-
    Mutationsteil (Status/Type/Prio-Menü, Tag-/Parent-/Blocking-Picker,
    Delete-Confirm) ✓ · US-06/US-07 ✓ · ETag-Konflikt-Handling überall (T6-Sweep) ✓.
  - Bewusste Scope-Cuts dokumentiert: Toast → E5 (Statuszeilen-Interim) · kein
    Client-Typ-Hierarchie-Check im Create-Parent-Feld (CLI-Autorität) · kein
    Blocking-Zyklen-Ausschluss (upstream-Parität, YAGNI) · Delete ohne --if-match
    (CLI-Signatur) · huh-Formbreite zum Open-Zeitpunkt (kein Resize-Rebuild).
  - Konsolidierung ggü. Quellen: EIN Value-Menü statt 3 upstream-Picker · EIN
    Overlay-Enum statt 6 Bools · SetTags/SetBlocking als Ein-Kommando-Diffs statt
    Mutations-Kaskaden auf einem ETag · hauseigenes modalPanel/menuList statt
    bubbles/list-Sub-Modelle.
  - Keymap: KEIN neues Binding (alle 7 existierten seit E1 T7);
    helpGroups-Drift-Guard unverändert grün.
- [ ] Commit `docs: E3-Abschluss` (Refs: bt-qzwt).
- [ ] Skill `ce-nsp-auto` → Handover-Prompt für E4 (Command-Center + Review-Cockpit,
  bean für E4 im Backlog; Review-Tag-Konvention §5 wird DORT gebraucht).

---

## Selbst-Review (Plan gegen design-spec + Bean-Body-Pflichten)

- **7 Tasks aus implementation-plan.md »Epos E3«**: Status/Type/Prio-Menüs (T1) ✓ ·
  Tag-Picker (T2) ✓ · Parent-/Blocking-Picker mit Zyklen-Ausschluss (T3) ✓ ·
  Create-Form+Confirm-Gate (T4) ✓ · Edit-Form+`$EDITOR` (T5) ✓ · Delete-Confirm+
  ETag-Handling-überall (T6) ✓ · Abschluss-Ritual (T7) ✓ — 1:1, keine Task
  verschmolzen ohne Kennzeichnung (a3: drei Menüs → EIN Menü, explizit begründet).
- **Epic-Body-PFLICHT B01**: Task 1 Step 1-3, erster Schritt des Epos, mit
  Regressionstest ✓.
- **Handover-Gotchas adressiert**: Datenlayer nur verdrahtet (3 verifizierte neue
  Wrapper: SetBody/AddBlocking+SetBlocking/SetTags — jede Ergänzung gegen
  `beans update --help` geprüft, nicht geraten) ✓ · huh v1.0.0 eingebettet, kein
  form.Run() ✓ · ErrConflict → Statuszeile, Toast explizit E5 ✓ · I01-CoW bei jedem
  `expanded`-Write (createDoneMsg-Handler cloned) ✓ · Theme-Tokens only
  (paletteFormTheme mappt auf theme.*) ✓ · keymap/helpGroups-Drift-Guard: kein neues
  Binding nötig ✓ · Goldens: bestehende bleiben byte-identisch (Enum-Refactor ändert
  Werte/Reihenfolge nicht; Filter-Menü-Label-Lowercase betrifft kein Golden — 
  treeFilterBox hat keins) ✓ · `command go` überall ✓.
- **Plan-Snippets gegen Code-Realität gedacht, nicht blind portiert** (E2-Erratum-
  Lehre): 2 im Task-Bean-Erstentwurf gefundene Fehlannahmen SOFORT korrigiert und
  in den Beans nachgezogen — (1) „Blocking sei nur server-berechneter
  Rückwärts-Index" → CLI hat `--blocking`-Flags (T3-Bean korrigiert); (2)
  „AppendBody für den Editor-Rückweg" → `--body` ist voller Replace, SetBody nötig
  (T5-Bean korrigiert). Dazu T2s SetTags-Entscheidung: N Einzel-Mutationen auf EINEM
  ETag wären eine Konflikt-Kaskade — als Ein-Kommando-Diff aufgelöst.
- **Windowing/Chrome wiederverwendet, nie neu gebaut**: alle Overlays über
  modalBox/modalPanel/menuList/placeOverlay/windowAround/listState (E1/E2) ✓.
- **data.Index nicht mutiert**: alle neuen Reads über ByID/Children; Mutationen
  ausschließlich über data.Client; jede UI-Mutation endet in unconditional Reload ✓.
- **Kein Platzhalter/TBD**: jede Task hat Dateinamen, Signaturen, Testnamen,
  Port-Zeilenangaben (devd/beans-src `file.go:NN-MM`); alle sechs geforderten
  Design-Punkte (a-f) plus drei zusätzliche (a2/a3/g/h) sind mit Entscheidung UND
  Begründung festgehalten.
