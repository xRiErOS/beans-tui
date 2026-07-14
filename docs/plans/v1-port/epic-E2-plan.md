# Epos E2 — Browse & Detail (voll granular)

> Geschrieben beim Epos-Start-Ritual (implementation-plan.md »Epos-Rituale«). Struktur
> identisch zu E1 in `implementation-plan.md`: je Task Files/TDD-Steps/Port-Referenzen/
> Commit. Quelle der Wahrheit: `design-spec.md` §6 (V2/V3/V4) + §7 (Keymap). Port-Quellen
> (devd, `~/Obsidian/tools/DeveloperDashboard/apps/cli-go/internal/tui/`): `accordion.go`,
> `view_detail_issue.go`, `view_browse_project.go`, `view_browse_backlog.go`, `editor.go`
> (glamour). Task-beans: `bt-ms0k`(T1) `bt-2jve`(T2) `bt-4ep2`(T3) `bt-9ldr`(T4)
> `bt-gzu6`(T5) `bt-enrd`(T6), parent `bt-aq5s`.

**Liefert:** V2 Browse komplett (Master-Detail, Fokus-Tausch, Beziehungs-Navigation), V4
Detail-Accordion (Meta/Body/Beziehungen/Historie, glamour-Body), Suche `/` (lokal + Bleve),
Facetten-Filter `f`/`X` (Status/Type/Priority/Tag, geteilt Tree+Backlog), V3 Backlog
(Master-Detail, Sort-Toggle `S`). Schließt drei PFLICHT-Items aus dem T8-Opus-Review
(bean `bt-7jr8`, jetzt im Epic-Body `bt-aq5s` verlinkt): **I01** (expanded-map Ownership,
Task 2), **I03** (Orphan- vs. Tree-Sortierung, Vorbereitung Task 1 / Abschluss Task 4),
**Q01** (Init() nil-Guard, Task 2). **I02** (flattenTree/collectOrphans-Cache) ist per
Bean-Body explizit NUR eine Notiz, kein Fix — hier bewusst nicht umgesetzt (YAGNI, s.
Selbst-Review). **Maus ist NICHT E2-Scope** (E5) — Windowing (`windowStart`/`windowAround`,
E1 Task 8) wird überall NUR wiederverwendet, nie neu gebaut.

**Reihenfolge (Task-bean `blocked_by`):** T1 (keine Vorbedingung) → T2 (braucht T1s
Accordion-Typen) ↘ T3 (unabhängig, kann parallel zu T1/T2 laufen — hier sequentiell
abgearbeitet) → T4 (erweitert T3s Filter-Infra) ↘ T5 (braucht T1 Accordion + T4s
geteilte Prädikate) → T6 (Abschluss, braucht T2 + T5 fertig).

---

## Task 1: Detail-Accordion-Port (`bt-ms0k`)

Liefert die Render-Infrastruktur (keine Verdrahtung — die folgt in Task 2): exklusiv
offene, ziffern-adressierte Accordion-Sections + die 4 beans-Sections selbst. Rein
lesend; Edit-Felder/Forms sind E3-Scope (kein `detailField{editor}"`-Äquivalent hier —
das eigene `relationField` ist NUR Sprung-Ziel, nie Edit-Ziel).

**Files:**
- Create: `internal/tui/accordion.go` (`accordionSection`, `relationField`,
  `renderAccordion`, `fieldStrip`-Äquivalent, `glowRender`)
- Create: `internal/tui/view_detail_bean.go` (`beanSections`)
- Modify: `internal/data/index.go` (I03-Vorbereitung: `SortBeans`/`StatusRank`/
  `PriorityRank` exportieren)
- Modify: `go.mod`/`go.sum` (glamour-Dependency)
- Test: `internal/data/index_test.go` (erweitert), `internal/tui/accordion_test.go`,
  `internal/tui/view_detail_bean_test.go`

**Port-Referenzen:**
- `renderAccordion`: devd `accordion.go:309-355` (exklusiv offene Section, Ziffern-Header
  `> [n] Title`, D08-Balken bei aktiver Section) — OHNE `detailFocusView.field`/Edit-Teil
  (Zeilen 346-355 dort behandeln `fieldStrip` bei Feld-Ebene; hier wird `fieldStrip`
  NUR für die Beziehungen-Section gebraucht, s.u.).
- `fieldStrip`: devd `accordion.go:360-373` — Konzept (aktives Feld hervorgehoben,
  Rest muted) 1:1 übernehmbar, `detailField.label` → `relationField.label`.
- `glowRender`: devd `editor.go:102-126` — Markdown→ANSI via glamour, `notty`-Style im
  Ascii-Farbprofil (Golden-Determinismus!), Fallback auf rohen Text bei Renderer-Fehler.

**Schritte:**

- [ ] **Step 1:** `command go get github.com/charmbracelet/glamour@v1.0.0` (bereits im
  Modul-Cache, s. `~/go/pkg/mod/github.com/charmbracelet/glamour@v1.0.0` — kein
  Netzwerkzugriff nötig). Verify `go.mod` trägt `glamour v1.0.0` als direkte Dependency.

- [ ] **Step 2: Failing tests (I03-Vorbereitung)** — `internal/data/index_test.go`:

```go
func TestSortBeansExportedMatchesCanonicalTierOrder(t *testing.T) {
    beans := []*Bean{
        {ID: "b", Status: "todo", Priority: "normal", Type: "task", Title: "B"},
        {ID: "a", Status: "in-progress", Priority: "high", Type: "bug", Title: "A"},
    }
    SortBeans(beans)
    if beans[0].ID != "a" { // in-progress sorts before todo
        t.Fatalf("SortBeans order = %v, want a before b", beans)
    }
}

func TestStatusRankOrdersLifecycle(t *testing.T) {
    if !(StatusRank("in-progress") < StatusRank("todo") &&
        StatusRank("todo") < StatusRank("draft") &&
        StatusRank("draft") < StatusRank("completed") &&
        StatusRank("completed") < StatusRank("scrapped")) {
        t.Fatal("StatusRank does not reproduce the documented tier order")
    }
}

func TestPriorityRankEmptyDefaultsNormal(t *testing.T) {
    if PriorityRank("") != PriorityRank("normal") {
        t.Fatalf("PriorityRank(\"\") = %d, want == PriorityRank(normal) = %d",
            PriorityRank(""), PriorityRank("normal"))
    }
    if !(PriorityRank("critical") < PriorityRank("high") && PriorityRank("high") < PriorityRank("normal")) {
        t.Fatal("PriorityRank does not reproduce the documented tier order")
    }
}
```

- [ ] **Step 3:** `command go test ./internal/data/` → FAIL (SortBeans/StatusRank/
  PriorityRank undefined).
- [ ] **Step 4: Implement** — `index.go` ergänzt (keine Änderung an der bestehenden
  unexported `sortBeans`/`rank`-Logik, nur dünne exportierte Wrapper, Single-Source
  bleibt gewahrt):

```go
// SortBeans is the exported single-source sort (Status -> Priority -> Type ->
// Title) every consumer OUTSIDE this package must use for bean lists it
// itself assembles (e.g. resolved Blocking/BlockedBy IDs) -- I03 (bean
// bt-7jr8 T8-review): a second, ad-hoc tie-break in the tui package would
// violate the "single place sort order is defined" contract this file
// already asserts for its own callers.
func SortBeans(beans []*Bean) { sortBeans(beans) }

// StatusRank exposes the status tier position (see statusOrder) for callers
// that need to compare two beans' status without a full SortBeans call
// (e.g. the Backlog sort-toggle, E2 Task 5).
func StatusRank(status string) int { return rank(statusOrder, status) }

// PriorityRank exposes the priority tier position, empty treated as "normal"
// (mirrors sortBeans' own empty-priority handling).
func PriorityRank(priority string) int {
    if priority == "" {
        priority = "normal"
    }
    return rank(priorityOrder, priority)
}
```

- [ ] **Step 5:** `command go test ./internal/data/` → PASS.

- [ ] **Step 6: Failing tests** — `internal/tui/accordion_test.go`:
  - `TestRenderAccordionExclusiveOpen`: 3 Sections, `open=2` → nur Section 2's Body
    erscheint im Output, Section 1/3 zeigen nur Header+Chevron (`▸`).
  - `TestRenderAccordionDigitHeaderNumbering`: Header-Zeilen enthalten `[1]`..`[3]` in
    Reihenfolge (ansi-strip vor Vergleich, wie bestehende Tests in `primitives_test.go`).
  - `TestRenderAccordionEmptySectionsShowsPlaceholder`: `renderAccordion(nil, ...)` →
    `"(no detail fields)"`-artiger Hinweis (Text muss NICHT identisch mit devd sein,
    aber vorhanden).
  - `TestGlowRenderEmptyBodyReturnsEmpty`: `glowRender("", 80) == ""`.
  - `TestGlowRenderAsciiProfileProducesNoEscapeSequences`: `lipgloss.SetColorProfile(termenv.Ascii)`
    gesetzt (wie golden-Tests) → Output enthält kein `\x1b[` (notty-Style, Determinismus
    für künftige Detail-Goldens).
- [ ] **Step 7:** `command go test ./internal/tui/` → FAIL (Typen/Funktionen fehlen).
- [ ] **Step 8: Implement** — `accordion.go`:

```go
package tui

// accordion.go — exclusive-open, digit-addressed detail sections (design-spec
// §6 V4). Port of devd accordion.go's render algebra (renderAccordion,
// fieldStrip), stripped of every edit-field/detailField-editor concept (E3
// scope) -- E2's "fields" are pure navigation targets inside the
// Beziehungen section (relationField), never edit targets.

type relationField struct {
    beanID string // target bean ID; "" for an unresolved/dangling reference (not jumpable)
    label  string // pre-rendered row text (theme colors already applied)
}

// accordionSection mirrors devd's shape minus the editor-specific parts.
// fields == nil for every section except Beziehungen (Meta/Body/Historie are
// pure display, no field-level navigation in E2).
type accordionSection struct {
    title  string
    body   string
    fields []relationField
}

func renderAccordion(secs []accordionSection, open, w int, active bool, activeIdx, fieldIdx int) string {
    // ... ported from devd accordion.go:309-355, activeIdx/fieldIdx replace
    // detailFocusView{sec,field}; fieldStrip only invoked when
    // len(secs[activeIdx].fields) > 0 (i.e. only ever for Beziehungen).
}

func fieldStrip(fields []relationField, active, w int) string {
    // ... ported from devd accordion.go:360-373, detailField.label -> relationField.label
}

func glowRender(md string, width int) string {
    // ... ported verbatim from devd editor.go:102-126
}
```

- [ ] **Step 9:** `command go test ./internal/tui/ -run TestRenderAccordion` and
  `-run TestGlowRender` → PASS.

- [ ] **Step 10: Failing tests** — `internal/tui/view_detail_bean_test.go`:
  - `TestBeanSectionsAlwaysFourFixedSections`: for a minimal bean (no body, no tags, no
    relations), `beanSections` returns exactly 4 sections titled `"Meta"`, `"Body"`,
    `"Beziehungen"`, `"Historie"` in that order (unlike devd's content-gated sections —
    beans-tui's 4 are unconditional, simpler digit-jump semantics: `1..4` always mean
    the same thing regardless of bean content).
  - `TestBeanSectionsMetaRendersStatusTypePriorityTags`: Meta body contains the themed
    status/type/priority/tag strings (reuses `theme.StatusStyle`/`TypeStyle`/`Priority`/
    `tagsInline` from `render_shared.go`, no new theme code).
  - `TestBeanSectionsBodyEmptyShowsPlaceholder` / `TestBeanSectionsBodyRendersMarkdown`
    (glamour heading/bold survives as ANSI in TrueColor profile).
  - `TestBeanSectionsBeziehungenListsParentChildrenBlockingBlockedByInCanonicalOrder`:
    fixture with 2 Blocking IDs whose resolved beans have different Status
    (in-progress + todo) → resolved order matches `data.SortBeans`/`StatusRank` tier
    order, NOT insertion order (I03 applied — Children already arrives pre-sorted via
    `idx.Children`, only Blocking/BlockedBy need the explicit `data.SortBeans` call
    here since `Bean.Blocking`/`Bean.BlockedBy` are raw `[]string` IDs).
  - `TestBeanSectionsBeziehungenDanglingReferenceShowsUnresolvedNotJumpable`: a
    `BlockedBy` ID with no match in `idx.ByID` renders as a muted `"(unresolved: <id>)"`
    row with `relationField.beanID == ""` (jump-guard in Task 2 checks this).
  - `TestBeanSectionsHistorieShowsCreatedUpdatedETag`: nil `CreatedAt`/`UpdatedAt` render
    a placeholder, not a zero-time string.
- [ ] **Step 11:** FAIL → **Step 12: Implement** — `view_detail_bean.go`:

```go
package tui

// view_detail_bean.go — the 4 fixed accordion sections for a bean (design-
// spec §6 V4): Meta / Body (glamour) / Beziehungen / Historie. Always all 4
// (unlike devd's content-gated issueSections) -- E2 has no edit-field
// concept yet (E3), so there's no reason to hide an empty section; digit
// jump 1..4 stays meaningful for every bean.

func beanSections(idx *data.Index, b *data.Bean, bodyW int) []accordionSection {
    var secs []accordionSection
    secs = append(secs, accordionSection{title: "Meta", body: metaSectionBody(b)})
    secs = append(secs, accordionSection{title: "Body", body: bodySectionBody(b, bodyW)})
    rel, fields := relationsSectionBody(idx, b, bodyW)
    secs = append(secs, accordionSection{title: "Beziehungen", body: rel, fields: fields})
    secs = append(secs, accordionSection{title: "Historie", body: historieSectionBody(b)})
    return secs
}
```

  `relationsSectionBody` resolves `b.Parent` (1, via `idx.ByID`), `idx.Children[b.ID]`
  (already sorted), `b.Blocking`/`b.BlockedBy` (resolve each ID via `idx.ByID`, THEN
  `data.SortBeans` the resolved slice — I03), grouped under muted sub-headers
  (`"Parent"`/`"Children"`/`"Blocking"`/`"Blocked By"`, mirrors the `facetHead`-map
  grouping style already used in devd's filter box, port-adapted). Each resolved
  relation becomes one `relationField{beanID: rel.ID, label: <status-icon+type-icon+ID+title>}`;
  an unresolved ID becomes `relationField{beanID: "", label: theme.Dim.Render("(unresolved: "+id+")")}`.

- [ ] **Step 13:** `command go test ./internal/tui/` → PASS.
- [ ] **Step 14:** `command go test ./... && command gofmt -l . && command go vet ./...`
  → all clean.
- [ ] **Step 15: Commit**

```
feat(tui): Detail-Accordion-Port (Meta/Body/Beziehungen/Historie)

I03-Vorbereitung: data.SortBeans/StatusRank/PriorityRank exportiert (Single-
Source-Sort auch außerhalb data), genutzt für die Blocking/BlockedBy-Listen der
neuen Beziehungen-Section. Body-Section rendert via glamour (Port devd
editor.go glowRender). Reine Render-Infrastruktur -- Verdrahtung (Fokus-
Maschine, Ziffer-Tasten, Beziehungs-Sprung) folgt in Task 2.

Refs: bt-ms0k
```

---

## Task 2: Master-Detail-Verdrahtung (`bt-2jve`)

Ersetzt den E1-Platzhalter (`if m.detailFocus { return m, nil }` in `update.go`) durch
die echte zwei-Ebenen-Fokus-Maschine + Beziehungs-Sprung. Schließt **I01** und **Q01**.

**Files:**
- Modify: `internal/tui/types.go` (neue Modell-Felder + I01-Konvention-Doku)
- Modify: `internal/tui/update.go` (`setExpanded` refactor, `keyDetailFocus`,
  `focusedBean`)
- Modify: `internal/tui/app.go` (Q01: `Init()` nil-Guard)
- Modify: `internal/tui/view_browse_repo.go` (`renderDetailPane` → echter Accordion)
- Test: `internal/tui/update_test.go` (erweitert), `internal/tui/app_test.go` (neu, Q01),
  `internal/tui/testdata/tree.golden` (ggf. `-update`)

**Port-Referenzen:**
- `keyDetailFocus`: devd `view_detail_issue.go:281-392` — Ziffer-Sprung (`348-356`),
  i/k Section-Wechsel (`358-378`), l/j Ebene rein/raus (`379-390`). Port-Abweichung:
  KEIN separater "Kopf/Übersicht"-Layer (devd Index 0 = editierbare Titel/Type/Prio-
  Felder, DD2-77) — beans-tui hat keine Header-Edit-Felder in E2 (E3-Scope), also
  ist Section-Index 0 direkt Meta (kein Off-by-one wie bei devd `secCursor-1`).
- `focusedIssue`: devd `view_detail_issue.go:20-35` (Dispatcher je nach `m.view`) — hier
  `focusedBean()`: `viewBrowseRepo` → Tree-Cursor-Bean, `viewBacklog` (Task 5) →
  Backlog-Listen-Selektion. Macht die Fokus-Maschine view-agnostisch (Task 5
  wiederverwendet sie unverändert).
- `paneBorderColors`/D03-Fokus-Border-Tausch: bereits in `view_browse_repo.go`
  (`!m.detailFocus`-Parameter an `renderPane`) aus E1 vorhanden — hier nur der
  Detail-Pane-Content ausgetauscht, keine neue Border-Logik nötig.

**Schritte:**

- [ ] **Step 1: I01 — Failing regression test** (beweist die AKTUELLE Shared-Map-
  Mutation, RED gegen den heutigen Code) — `internal/tui/update_test.go`:

```go
func TestSetExpandedDoesNotMutateSharedMapAcrossModelCopies(t *testing.T) {
    base := model{expanded: map[string]bool{}}
    copy1 := base // struct copy -- map HEADER copied, backing array still shared pre-fix
    node := treeNode{id: "x", hasKids: true}
    copy2 := copy1.setExpanded(node, true)

    if copy1.expanded["x"] {
        t.Error("copy1.expanded was mutated by copy2's setExpanded call -- " +
            "expanded is a shared map, not copy-on-write (I01, bean bt-7jr8 T8-review)")
    }
    if !copy2.expanded["x"] {
        t.Error("copy2.expanded should carry the new expand state")
    }
}
```

- [ ] **Step 2:** `command go test ./internal/tui/ -run TestSetExpandedDoesNotMutate`
  → FAIL (current `setExpanded` mutates the shared backing map in place).
- [ ] **Step 3: Implement (I01 resolution: Copy-on-Write)** — add to `types.go`, right
  above the `model` struct, the doc-stamp resolving I01:

```go
// I01 (bean bt-7jr8, T8-review): every map[string]bool field on model
// (expanded, and E2's new filter facet sets) is COPY-ON-WRITE, never mutated
// in place. Rationale: model is a value-receiver Elm architecture (design-
// spec.md §3.3) -- Go map values are reference types, so mutating one
// in-place silently aliases every other struct copy holding the same map
// header (e.g. an old model variable a test kept around, or a future
// undo/diff feature). Bubbletea's own single-active-model discipline
// (Update always returns a fresh value, the caller discards the old one)
// currently hides this hazard, but it is not guaranteed to stay hidden as
// more map fields are added (E2 Task 4's filter facets) -- so every setter
// clones via cloneBoolMap before writing, closing the hazard permanently at
// negligible cost (these maps are always small: expand state + a handful of
// filter selections).
func cloneBoolMap(src map[string]bool) map[string]bool {
    out := make(map[string]bool, len(src))
    for k, v := range src {
        out[k] = v
    }
    return out
}
```

  Refactor `update.go`'s `setExpanded`:

```go
func (m model) setExpanded(n treeNode, open bool) model {
    if !n.hasKids {
        return m
    }
    m.expanded = cloneBoolMap(m.expanded)
    if open {
        m.expanded[n.id] = true
    } else {
        delete(m.expanded, n.id)
    }
    return m
}
```

- [ ] **Step 4:** `command go test ./internal/tui/` → PASS (both the new regression
  test and every existing expand/collapse test from E1).

- [ ] **Step 5: Q01 — Failing test** — `internal/tui/app_test.go` (new file):

```go
func TestInitNilClientReturnsErrorMsgInsteadOfPanicking(t *testing.T) {
    m := newModel(nil, "/tmp/does-not-matter")
    cmd := m.Init()
    if cmd == nil {
        t.Fatal("Init() with a nil client must still return a cmd (never nil -- caller expects a msg)")
    }
    msg := cmd() // must not panic
    loaded, ok := msg.(beansLoadedMsg)
    if !ok || loaded.err == nil {
        t.Fatalf("Init() with nil client should yield a beansLoadedMsg carrying an error, got %#v", msg)
    }
}
```

- [ ] **Step 6:** `command go test ./internal/tui/ -run TestInitNilClient` → FAIL
  (current `Init()` calls `loadCmd(m.client)` unconditionally; `loadCmd`'s returned
  func would nil-deref inside `c.List()` → `c.run()` → `cmd.Dir = c.RepoDir`).
- [ ] **Step 7: Implement** — `app.go`:

```go
var errNilClient = errors.New("bt: nil beans client (Run() must construct a data.Client before newModel)")

func (m model) Init() tea.Cmd {
    if m.client == nil { // Q01 (bean bt-7jr8 T8-review): the nil-client invariant is
        // otherwise only enforced by convention at Run()'s call site -- this guard
        // turns a would-be nil-deref panic (inside loadCmd -> Client.List -> run)
        // into a normal, status-line-surfaced load error instead.
        return func() tea.Msg { return beansLoadedMsg{err: errNilClient} }
    }
    return loadCmd(m.client)
}
```

- [ ] **Step 8:** `command go test ./internal/tui/` → PASS.

- [ ] **Step 9: Failing tests — Fokus-Maschine** — `internal/tui/update_test.go`:
  - `TestFocusedBeanDispatchesOnTreeCursorInBrowseView`.
  - `TestDetailFocusDigitJumpOpensMatchingSection` (press `"3"` while `detailFocus` →
    `secCursor==2 && accOpen==3`, mirrors devd digit semantics 1-based `accOpen`,
    0-based `secCursor`).
  - `TestDetailFocusUpDownMovesSectionCursorClampedAtEnds` (4 sections, down from
    section 4 stays at 4).
  - `TestDetailFocusRightEntersFieldLevelOnlyForBeziehungenSection` (right/`l` on
    Meta/Body/Historie is a no-op — no fields; on Beziehungen with ≥1 relation it sets
    `detailLevel=1, fieldCursor=0`).
  - `TestDetailFocusEnterOnRelationJumpsCursorAndExitsToTree` (fixture: bean A blocked
    by B; focus A's Beziehungen field-level, cursor on the Blocking row for B, enter →
    `m.cursorID == "B"`, `m.detailFocus == false`, and B's ancestor chain is expanded in
    the returned model so it is actually visible in `visibleNodes()`).
  - `TestDetailFocusEnterOnUnresolvedRelationIsNoOp` (relationField with `beanID == ""`
    — dangling reference — enter does nothing, stays in field-level).
  - `TestDetailFocusLeftAtFieldLevelReturnsToSectionLevel` /
    `TestDetailFocusLeftAtSectionLevelExitsDetailFocus`.
- [ ] **Step 10:** FAIL → **Step 11: Implement** — `update.go`:

```go
func (m model) focusedBean() *data.Bean {
    switch m.view {
    case viewBacklog: // wired in Task 5; nil-safe no-op until then
        return m.backlogSelected()
    default:
        nodes := m.visibleNodes()
        pos := m.cursorPos(nodes)
        if pos < 0 || pos >= len(nodes) || nodes[pos].orphan {
            return nil
        }
        return nodes[pos].bean
    }
}

func (m model) keyDetailFocus(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    b := m.focusedBean()
    if b == nil {
        m.detailFocus = false
        return m, nil
    }
    secs := beanSections(m.idx, b, 40) // width is render-time only, section COUNT is fixed (4)

    if s := msg.String(); len(s) == 1 && s[0] >= '1' && s[0] <= '4' {
        m.secCursor = int(s[0]-'0') - 1
        m.accOpen = int(s[0] - '0')
        m.detailLevel = 0
        m.fieldCursor = 0
        return m, nil
    }
    switch navKey(msg.String()) {
    case "up":
        if m.detailLevel == 0 && m.secCursor > 0 {
            m.secCursor--
            m.accOpen = m.secCursor + 1
        } else if m.detailLevel == 1 && m.fieldCursor > 0 {
            m.fieldCursor--
        }
        return m, nil
    case "down":
        if m.detailLevel == 0 && m.secCursor < len(secs)-1 {
            m.secCursor++
            m.accOpen = m.secCursor + 1
        } else if m.detailLevel == 1 && m.fieldCursor < len(secs[m.secCursor].fields)-1 {
            m.fieldCursor++
        }
        return m, nil
    case "right":
        if m.detailLevel == 0 && len(secs[m.secCursor].fields) > 0 {
            m.detailLevel = 1
            m.fieldCursor = 0
        }
        return m, nil
    case "left":
        if m.detailLevel == 1 {
            m.detailLevel = 0
        } else {
            m.detailFocus = false
        }
        return m, nil
    }
    if keybind.Matches(msg, keys.Enter) && m.detailLevel == 1 {
        f := secs[m.secCursor].fields[m.fieldCursor]
        if f.beanID == "" {
            return m, nil // unresolved reference -- nothing to jump to
        }
        m.expanded = expandAncestorsOf(m.idx, m.expanded, f.beanID) // I01: clone-based
        m.cursorID = f.beanID
        m.detailFocus = false
        return m, nil
    }
    return m, nil
}

// expandAncestorsOf returns a NEW expanded map (I01 copy-on-write) with every
// ancestor of id (walking Parent up to a root) marked expanded, so a
// relation-jump target is guaranteed visible in the next visibleNodes() call.
func expandAncestorsOf(idx *data.Index, expanded map[string]bool, id string) map[string]bool {
    out := cloneBoolMap(expanded)
    b, ok := idx.ByID[id]
    for ok && b.Parent != "" {
        out[b.Parent] = true
        b, ok = idx.ByID[b.Parent]
    }
    return out
}
```

  Add `secCursor, accOpen, detailLevel, fieldCursor int` fields to `model` in
  `types.go`. Wire `handleKey`'s existing `if m.detailFocus { ... }` stub to call
  `m.keyDetailFocus(msg)` instead of `return m, nil`. `enterDetailFocus`-equivalent:
  extend the existing `tab` case in `handleKey` to also reset
  `secCursor=0; accOpen=1; detailLevel=0; fieldCursor=0` when TURNING focus ON (mirrors
  devd `enterDetailFocus`, `view_detail_issue.go:236-243`).
- [ ] **Step 12:** `command go test ./internal/tui/` → PASS.

- [ ] **Step 13:** Wire `view_browse_repo.go`'s `renderDetailPane` to call
  `beanSections(m.idx, bean, rw-4)` + `renderAccordion(secs, m.accOpen, rw-2,
  m.detailFocus, m.secCursor, m.fieldCursor)` instead of the E1 placeholder
  (title+meta-line only). Existing `TestTreeGolden`/`TestTreeGoldenDeterministic` will
  very likely need `-update` (detail pane content changes materially) — regenerate,
  THEN re-run without `-update` to confirm determinism still holds.
- [ ] **Step 14:** `command go test ./... && command gofmt -l . && command go vet ./...`
  clean.
- [ ] **Step 15: Commit**

```
feat(tui): Master-Detail-Verdrahtung — Fokus-Maschine, Beziehungs-Sprung

I01 (bean bt-7jr8 T8-review): expanded-Map ist jetzt copy-on-write
(cloneBoolMap) statt in-place-mutiert -- Konvention in types.go dokumentiert,
gilt ab jetzt für JEDES map[string]bool-Modellfeld (Task 4 fügt die nächsten
hinzu). Q01: Init() hat einen nil-Client-Guard (kein Panic mehr, Fehler landet
in der Statuszeile). keyDetailFocus (Port devd view_detail_issue.go)
verdrahtet Ziffer-Sprung, Section<->Feld-Navigation und Beziehungs-Sprung
(enter springt + expandiert den Vorfahren-Pfad + verlässt den Detail-Fokus).
focusedBean() ist view-agnostisch (Task 5 Backlog nutzt sie unverändert).

Refs: bt-2jve
```

---

## Task 3: Suche `/` (`bt-4ep2`)

Lokaler Live-Filter (sofort, Title-Substring) + Bleve-Modus ab 3 Zeichen (async,
staleness-sicher). Baut die generische gefilterte Baum-Flachung, die Task 4 um
Facetten erweitert.

**Files:**
- Modify: `internal/data/client.go` (`Search`)
- Modify: `internal/tui/types.go` (Suchfelder)
- Modify: `internal/tui/view_browse_repo.go` (`flattenTreeFiltered`, Suchkopfzeile,
  `visibleNodes` Umschaltung)
- Modify: `internal/tui/update.go` (`keyTree`-Erweiterung: `/` öffnet, esc-Kaskade)
- Modify: `internal/tui/messages.go` (`searchBleveResultMsg`, `searchCmd`)
- Test: `internal/data/client_test.go` (erweitert), `internal/tui/search_test.go` (neu)

**Port-Referenzen:**
- `treeNodesFiltered`-Ancestor-Erhalt + DD2-178-Nuance (kollabierte Vorfahren
  verstecken Treffer weiterhin, KEIN Auto-Expand): devd `view_browse_project.go:161-241`,
  insbesondere die Kommentare `215-238` — **wichtig, wörtlich portieren**: ein manuell
  kollabierter Knoten bleibt kollabiert, auch wenn er Treffer enthält (PO-Entscheidung
  im devd-Code, nicht neu verhandeln).
- `keyTreeSearch`/Live-Filter: devd `view_browse_project.go:1073-1097`.
- `treeSearchLine`: devd `view_browse_project.go:1099-1117` (Shield-Glyph `⌕`, Rot bei
  aktivem Filter/Suche).
- esc-Kaskade: devd `view_browse_project.go:725-736` — Port-Abweichung: KEIN
  `goHome()`-Fallback (kein Lobby in E2, design-spec §12 E5) — die Kaskade endet nach
  "Suche+Filter leeren", ein drittes `esc` ist ein No-Op (dokumentierter Scope-Cut,
  kein TBD).

**Schritte:**

- [ ] **Step 1: Failing test** — `internal/data/client_test.go`:

```go
func TestClientSearchInvokesBleveFlagAndReturnsMatches(t *testing.T) {
    repo := newTestRepo(t) // existing fixture helper (client_test.go / testrepo_test.go)
    c := &Client{RepoDir: repo}
    beans, err := c.Search("golden")
    if err != nil {
        t.Fatal(err)
    }
    // fixture must contain >=1 bean with "golden" in title or body (extend
    // newTestRepo's fixture set if the existing 3 fixtures don't cover this,
    // or add a 4th fixture bean here specifically for this test).
    if len(beans) == 0 {
        t.Fatal("Search(\"golden\") returned no matches against a fixture repo containing one")
    }
}
```

- [ ] **Step 2:** `command go test ./internal/data/` → FAIL (`Search` undefined).
- [ ] **Step 3: Implement** — `client.go`:

```go
// Search runs a Bleve full-text query (title+body) via `beans list -S`, with
// the same --full/--json contract as List (E2 Task 3, design-spec.md §6 V2:
// "-S-Bleve-Modus ab 3 Zeichen").
func (c *Client) Search(query string) ([]Bean, error) {
    out, err := c.run("list", "--json", "--full", "--search", query)
    if err != nil {
        return nil, err
    }
    var beans []Bean
    if err := json.Unmarshal(out, &beans); err != nil {
        return nil, fmt.Errorf("beans list --search: parse output: %w", err)
    }
    return beans, nil
}
```

- [ ] **Step 4:** `command go test ./internal/data/` → PASS.

- [ ] **Step 5: Failing tests** — `internal/tui/search_test.go`:
  - `TestLocalSearchFiltersTreeByTitleSubstringCaseInsensitive`.
  - `TestLocalSearchPreservesAncestorPathOfMatch` (a matching leaf's expanded parent
    stays visible even though the parent's own title doesn't match).
  - `TestLocalSearchCollapsedAncestorHidesMatchingDescendant` (DD2-178 parity: parent
    NOT expanded → its matching child is absent from `visibleNodes()`, only the
    collapsed parent row itself shows, per devd `view_browse_project.go:215-238`).
  - `TestSearchTypingUpdatesQueryLiveAndResetsCursor` (KeyMsg sequence: `/`, then
    runes, each keystroke updates `m.searchQuery` and resets `m.cursorID`/position to
    the first visible match, mirrors `TestCursorMovesAndExpands`'s KeyMsg-sequence
    style from E1).
  - `TestSearchEnterCommitsQueryStaysActiveInputBlurred`.
  - `TestSearchEscWhileTypingCancelsAndClearsQuery`.
  - `TestSearchEscAfterCommitClearsQueryAndFacets` (esc-Kaskade: query committed, not
    typing → esc clears `searchQuery` + all filter maps in one step, per devd
    `view_browse_project.go:725-736` minus the `goHome()` fallback).
  - `TestSearchBleveFiresOnlyAtThreeOrMoreChars` (2 chars → no `tea.Cmd` batch item of
    the Bleve kind produced; 3 chars → one produced, tag it with the query so the test
    can assert `searchCmd`'s returned msg carries `query: "abc"`).
  - `TestSearchBleveStaleResultDiscardedWhenQueryChangedMeanwhile` (apply
    `searchBleveResultMsg{query: "abc", ids: [...]}` to a model whose CURRENT
    `searchQuery` is now `"abcd"` → Update is a no-op on the result set, no panic).
- [ ] **Step 6:** FAIL → **Step 7: Implement**:
  - `types.go`: `searchActive bool`, `searchInput textinput.Model`, `searchQuery
    string`, `searchBleveIDs map[string]bool`, `searchBleveFor string` (which query the
    current `searchBleveIDs` answers), `searchBleveLoading bool`.
  - `messages.go`: `searchBleveResultMsg{query string; ids []string; err error}`,
    `searchCmd(c *data.Client, query string) tea.Cmd` (calls `c.Search`, wraps into the
    msg, tags with `query` for staleness-guarding on arrival).
  - `view_browse_repo.go`:

```go
// treeSearchActive reports whether a local/Bleve search query is currently
// narrowing the tree (Task 4 extends this with facet state via treeActive()).
func (m model) treeSearchActive() bool {
    return strings.TrimSpace(m.searchQuery) != ""
}

// beanMatchesSearch is the search half of the combined predicate Task 4
// extends with facet criteria (AND-combined there as beanMatches).
func (m model) beanMatchesSearch(b *data.Bean) bool {
    q := strings.ToLower(strings.TrimSpace(m.searchQuery))
    if q == "" {
        return true
    }
    if len(q) >= 3 && m.searchBleveFor == m.searchQuery {
        return m.searchBleveIDs[b.ID] // authoritative once Bleve results for THIS query arrived
    }
    return strings.Contains(strings.ToLower(b.Title), q) // <3 chars, or Bleve still in flight
}

// flattenTreeFiltered mirrors flattenTree's depth-first walk but only
// includes a subtree the caller has actually expanded (devd DD2-178 parity,
// view_browse_project.go:215-238: a manually collapsed ancestor hides a
// matching descendant -- it is NOT force-opened). A collapsed node whose
// (structurally, expand-state-independent) subtree contains no match at all
// is dropped entirely; one that does contain a match anywhere below still
// renders as a single collapsed context row.
func flattenTreeFiltered(idx *data.Index, expanded map[string]bool, match func(*data.Bean) bool) []treeNode {
    // ... walk idx.Roots() via a recursive helper returning (nodes []treeNode, hit bool);
    // hit = match(b) || (expanded[b.ID] && any child hit) || (!expanded[b.ID] && subtreeHasMatch(...))
    // orphan bucket goes through the SAME predicate -- an orphan bucket with
    // zero matching orphans/cycle-beans is omitted entirely, matching ones
    // render exactly like today (I03/Task 4 still owns their sort order).
}
```

  - `update.go`: `keyTree` gains the `/`-open case (mirrors devd's `keys.Search` case,
    `view_browse_project.go:689-699`, minus `loadAllIssues` — beans-tui already holds
    the full `Index` in memory, no separate "load everything for search" step needed);
    a `keySearchInput(msg tea.KeyMsg)` handler for `m.searchActive == true` (enter
    commits + blurs, esc cancels + clears, any other key updates `searchInput` +
    live-copies its value into `searchQuery`, resets cursor); the Bleve-trigger:
    whenever the committed-or-live `searchQuery` reaches length ≥ 3 and differs from
    `searchBleveFor`, batch in `searchCmd(m.client, m.searchQuery)`.
  - `visibleNodes()` (`view_browse_repo.go`) switches: `if m.treeSearchActive() {
    return flattenTreeFiltered(m.idx, m.expanded, m.beanMatchesSearch) }; return
    flattenTree(m.idx, m.expanded)` (Task 4 replaces `m.treeSearchActive()`/
    `m.beanMatchesSearch` with the combined `m.treeActive()`/`m.beanMatches`).
- [ ] **Step 8:** `command go test ./internal/tui/` → PASS.
- [ ] **Step 9:** Regenerate `tree.golden` with `-update` IF the search header line
  changes the Tree pane's visible row budget (it adds one header row) — verify
  determinism re-run afterward.
- [ ] **Step 10:** `command go test ./... && command gofmt -l . && command go vet ./...`
  clean.
- [ ] **Step 11: Commit**

```
feat(tui): Suche / — lokaler Live-Filter + Bleve ab 3 Zeichen

data.Client.Search wrappt `beans list --search` (gleicher --full/--json-
Vertrag wie List). flattenTreeFiltered erhält Vorfahren-Pfade von Treffern
sichtbar, respektiert aber manuell kollabierte Knoten (DD2-178-Parität zu
devd -- ein Treffer unter einem zugeklappten Vorfahren bleibt verborgen,
kein Auto-Expand). Bleve-Antworten sind query-getaggt und werden verworfen,
wenn die Suche inzwischen weitergetippt wurde (Staleness-Guard statt
Debounce-Timer).

Refs: bt-4ep2
```

---

## Task 4: Facetten-Filter `f` + `X` (`bt-9ldr`)

Erweitert Task 3s gefilterte Flachung um Status/Type/Priority/Tag-Facetten (EIN
geteilter Zustand für Tree UND Backlog, design-spec §6 US-05). Schließt **I03**.

**Files:**
- Create: `internal/tui/box_filter_facets.go`
- Modify: `internal/tui/types.go` (Facet-Maps + Menü-State)
- Modify: `internal/tui/keymap.go` (`Toggle`-Binding)
- Modify: `internal/tui/view_browse_repo.go` (`treeActive`/`beanMatches` verallgemeinert,
  I03: `sortByTitleThenID` → `data.SortBeans`)
- Modify: `internal/tui/update.go` (`f`/`X` Routing, `keyFilterMenu`)
- Test: `internal/tui/box_filter_facets_test.go`, `internal/tui/update_test.go`
  (I03-Regressionscheck), `internal/tui/view_browse_repo_test.go` (falls noch nicht
  vorhanden — sonst `update_test.go` erweitert)

**Port-Referenzen:**
- Facetten-Menü-Aufbau + Render: devd `view_browse_backlog.go:89-136`
  (`buildBacklogFilterItems`/`backlogFilterActive`/`openBacklogFilter` — schlankeres
  Muster als der Tree-Filter, KEINE "Art"-Facette nötig, da beans' `Type`-Enum
  Milestone/Epic/Feature/Task/Bug bereits vollständig abdeckt) UND
  `view_browse_project.go:936-1033` (`keyTreeFilter`/`treeFilterBox` — Checkbox-
  Rendering mit `facetHead`-Gruppierung, `space`/`x` Toggle, `X` Clear).
- `ffItem{facet, value, label}`: devd `view_browse_project.go:48-54` — Struct-Form 1:1
  übernommen, aber EIN gemeinsamer Satz für Tree+Backlog (devd dupliziert
  Tree-Filter/Backlog-Filter als zwei Parallel-Implementierungen — beans-tui
  konsolidiert das bewusst, da design-spec "Filter wirkt auf Tree UND Backlog"
  explizit EINEN Zustand verlangt).

**Schritte:**

- [ ] **Step 1: I03-Abschluss — Failing test** (beweist die aktuelle, abweichende
  Orphan-Sortierung soll durch die kanonische ersetzt werden) — erweitere
  `internal/tui/update_test.go`:

```go
func TestOrphanBucketUsesCanonicalStatusPriorityTypeTitleOrder(t *testing.T) {
    beans := []data.Bean{
        {ID: "root", Title: "Root", Type: "milestone", Status: "todo"},
        {ID: "orph-z", Title: "Z orphan", Status: "todo", Type: "task", Priority: "normal", Parent: "missing"},
        {ID: "orph-a", Title: "A orphan", Status: "in-progress", Type: "task", Priority: "critical", Parent: "missing"},
    }
    m := fixtureModel(t, beans)
    m.expanded[orphanRootID] = true
    nodes := m.visibleNodes()
    // canonical order ranks in-progress before todo -> orph-a (in-progress)
    // must render before orph-z (todo), NOT alphabetically (orph-a happens to
    // also win alphabetically here on purpose -- add a second case with
    // reversed alphabetical/status order if needed to make the assertion
    // unambiguous against a title-only sort).
    var order []string
    for _, n := range nodes {
        if n.bean != nil && n.bean.Parent == "missing" {
            order = append(order, n.id)
        }
    }
    if len(order) != 2 || order[0] != "orph-a" {
        t.Fatalf("orphan order = %v, want [orph-a orph-z] (canonical status tier, not title)", order)
    }
}
```

  (Design the fixture so canonical-order and alphabetical-order DISAGREE, so the test
  is a real discriminator — e.g. swap titles so the alphabetically-first orphan has
  the LATER status tier, forcing the assertion to fail under the old
  `sortByTitleThenID` and pass only under `data.SortBeans`.)

- [ ] **Step 2:** `command go test ./internal/tui/ -run TestOrphanBucketUsesCanonical`
  → FAIL against current `sortByTitleThenID`.
- [ ] **Step 3: Implement (I03 closure)** — `view_browse_repo.go`: replace both
  `sortByTitleThenID(out)` call sites (`collectOrphans`, `collectCycleOrphans`) with
  `data.SortBeans(out)`; delete the now-dead `sortByTitleThenID` function.
- [ ] **Step 4:** `command go test ./internal/tui/` → PASS. **Check existing E1 tests**
  `TestOrphanShownUnderSyntheticRoot` / `TestCycleBeansShowUnderOrphanRoot`
  (`update_test.go`) for hardcoded multi-orphan ordering assumptions — adjust their
  fixtures/expectations to canonical order if they assumed title-only sort (they were
  written with 1 orphan each in the E1 fixtures per `tree_golden_test.go`'s
  `goldenTreeModel`, so a conflict is unlikely but MUST be verified, not assumed).

- [ ] **Step 5: Failing tests — Facet core** — `internal/tui/box_filter_facets_test.go`:
  - `TestBuildFilterItemsCoversStatusTypePriorityAndDynamicTags` (5 status values, 5
    beans-types, 5 priorities as FIXED options; tag options DYNAMIC from `idx.ByID`,
    stable insertion order, deduped).
  - `TestFacetToggleUsesCopyOnWrite` (mirrors `TestSetExpandedDoesNotMutate...` from
    Task 2: toggling a facet on one model copy must not mutate a sibling copy's map —
    same regression shape, new field).
  - `TestFilterActiveWhenAnyFacetMapNonEmpty`.
  - `TestFilterClearXResetsAllFourFacetMaps`.
  - `TestFilterMenuUpDownMovesCursorSpaceTogglesRow`.
  - `TestFilterMenuEscAndEnterAndFCloseWithoutClearing`.
  - `TestTreeCombinesSearchAndFacetsWithAnd` (a bean matching the search text but
    excluded by an active status facet is NOT visible; matches both → visible).
- [ ] **Step 6:** FAIL → **Step 7: Implement**:
  - `keymap.go`: add `Toggle keybind.Binding // space/x — toggle facet checkbox` (`
    keybind.NewBinding(keybind.WithKeys(" ", "x"), keybind.WithHelp("space/x", "Toggle facet"))`),
    add to `helpGroups()`'s "Actions" group (the existing
    `TestHelpGroupsCoverEveryBindingExactlyOnce` drift-guard from E1 enforces this —
    forgetting it fails that test, not a new one).
  - `types.go`: `filterStatus, filterType, filterPriority, filterTag map[string]bool`,
    `filterOpen bool`, `filterItems []ffItem`, `filterMenu listState` (reuse from
    `list.go`, first consumer since E1).
  - `box_filter_facets.go`:

```go
package tui

// box_filter_facets.go — the ONE shared facet-filter menu for Tree AND
// Backlog (design-spec.md §6 US-05: "Filter wirkt auf Tree UND Backlog").
// Port pattern: devd view_browse_backlog.go:89-136 (buildBacklogFilterItems
// shape) + view_browse_project.go:998-1033 (treeFilterBox render) --
// consolidated into one implementation instead of devd's two parallel ones,
// since beans-tui's filter state is model-level, not per-view.

type ffItem struct {
    facet string // "status" | "type" | "priority" | "tag"
    value string
    label string
}

func (m model) buildFilterItems() []ffItem {
    // fixed status/type/priority facets (beans enums, design-spec.md §4) +
    // dynamic tag facet collected from m.idx.ByID (stable insertion order,
    // deduped) -- mirrors devd tagFilterOptions/filterStatusOptions shape.
}

func (m model) filterActive() bool {
    return len(m.filterStatus)+len(m.filterType)+len(m.filterPriority)+len(m.filterTag) > 0
}

func (m model) beanMatchesFacets(b *data.Bean) bool {
    if len(m.filterStatus) > 0 && !m.filterStatus[b.Status] { return false }
    if len(m.filterType) > 0 && !m.filterType[b.Type] { return false }
    if len(m.filterPriority) > 0 && !m.filterPriority[b.Priority] { return false }
    if len(m.filterTag) > 0 {
        hit := false
        for _, t := range b.Tags { if m.filterTag[t] { hit = true; break } }
        if !hit { return false }
    }
    return true
}

// beanMatches is the ONE combined predicate every filtered view (Tree Task 3,
// Backlog Task 5) calls -- AND of search (Task 3) and facets (this task).
func (m model) beanMatches(b *data.Bean) bool {
    return m.beanMatchesSearch(b) && m.beanMatchesFacets(b)
}

func (m model) treeFilterBox() string { /* port devd view_browse_project.go:998-1033 */ }

func (m model) keyFilterMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    // up/down move m.filterMenu; space/x toggles the cursored ffItem's facet
    // map via cloneBoolMap (I01); X clears all four maps (fresh empty maps,
    // not nil -- keeps filterActive() cheap); esc/f/enter closes the menu.
}
```

  - `view_browse_repo.go`: rename/replace `treeSearchActive`/`beanMatchesSearch`
    call-sites in `visibleNodes()` with `m.treeActive()` (`= treeSearchActive() ||
    filterActive()`) and `m.beanMatches` (the AND-combination above) — Task 3's
    functions stay as the search-only half `beanMatches` calls into.
  - `update.go`: `keyTree` gains the `f`-open case (`m.filterItems =
    m.buildFilterItems(); m.filterMenu.setLen(len(m.filterItems)); m.filterOpen =
    true`) and routes to `keyFilterMenu` when `m.filterOpen`; `X` (already
    `keys.FilterClear`) works both inside the menu AND, per design-spec, as a direct
    top-level reset even when the menu is closed (mirrors devd's `esc`-cascade
    clearing filters too) — wire both paths to the same clear helper.
- [ ] **Step 8:** `command go test ./internal/tui/` → PASS.
- [ ] **Step 9:** Golden check: if the search/filter summary line's rendering changes
  Tree pane row budget further, regenerate `tree.golden` (`-update`) once more, verify
  determinism.
- [ ] **Step 10:** `command go test ./... && command gofmt -l . && command go vet ./...`
  clean.
- [ ] **Step 11: Commit**

```
feat(tui): Facetten-Filter f/X (Status/Type/Priority/Tag)

Ein geteilter Filter-Zustand für Tree UND Backlog (design-spec US-05) --
konsolidiert gegenüber devds zwei parallelen Tree-/Backlog-Filter-
Implementierungen in EINE (box_filter_facets.go), da beans-tui den Zustand
ohnehin model-weit hält. Facet-Maps folgen der I01-Copy-on-Write-Konvention
aus Task 2. I03 abgeschlossen: der Orphan-Bucket sortiert jetzt kanonisch
(data.SortBeans) statt über eine zweite, parallele Titel-Sortierlogik --
sortByTitleThenID entfernt.

Refs: bt-9ldr
```

---

## Task 5: Backlog-View V3 `b` (`bt-gzu6`)

Parentlose+ready beans, Master-Detail (reuse Task 1 Accordion via `focusedBean()`),
geteilter Such-/Filter-Zustand (Task 3/4), Sort-Toggle `S`.

**Files:**
- Create: `internal/tui/view_browse_backlog.go`
- Modify: `internal/tui/types.go` (`viewBacklog`, `backlogList listState`,
  `backlogSort string`)
- Modify: `internal/tui/keymap.go` (`Sort` binding `S`)
- Modify: `internal/tui/update.go` (`View()`/`Update()` dispatch, `b` open/close)
- Test: `internal/tui/view_browse_backlog_test.go`,
  `internal/tui/testdata/backlog.golden` (neu)

**Port-Referenzen:**
- Gesamtstruktur: devd `view_browse_backlog.go` (Master-Detail, Search-/Filter-Kopf,
  Windowing) — Port-Abweichungen: (1) KEIN eigenes Filter-Menü (`backlogFilterBox`
  entfällt — Task 4s `box_filter_facets.go` ist bereits geteilt); (2) Sortier-Modi
  `status/prio/created/updated` (design-spec §6 V3) statt devds `prio/title/created/
  key` — implementiert als reiner Tasten-ZYKLUS (`S` rückt eine Position weiter),
  KEIN schwebendes Sortier-Menü (design-spec §6 V8 listet keine "Sort-Overlay" —
  bewusster Vereinfachungs-Cut ggü. devds `backlogSortBox`).
- `backlogVisible`: devd `view_browse_backlog.go:57-77` — Port-Abweichung: ruft
  `m.beanMatches` (Task 4, geteilt) statt eigener Such-/Facet-Logik auf.
- `windowBlocks`/`blockWindow`: devd `view_browse_backlog.go:204-312` — NICHT portiert
  (Scope-Cut): beans-Titel sind kurz genug für die bestehende EINZEILIGE
  `windowAround`-Fensterung (E1) zu genügen; die variable-Höhen-Block-Fensterung war
  in devd durch mehrzeilige Issue-Header+Titel-Umbruch motiviert. Falls ein
  Titel im Review zu lang für eine Zeile ist: `truncate` (bestehend) greift, kein
  neuer Mechanismus (YAGNI, explizit dokumentiert statt stillschweigend übernommen).

**Schritte:**

- [ ] **Step 1: Failing tests** — `internal/tui/view_browse_backlog_test.go`:
  - `TestBacklogShowsParentlessReadyBeansFromIndex` (reuses `idx.Backlog()` from E1
    Task 3 unchanged).
  - `TestBacklogAppliesSharedSearchAndFacetFilters` (a bean excluded by an active
    Task-4 facet is absent from the backlog list too — same `m.beanMatches` call).
  - `TestBacklogSortCyclesThroughFourModesAndBackToStart` (press `S` 4×, mode sequence
    `status→priority→created→updated→status`, order changes each press verified via
    `data.StatusRank`/`PriorityRank` (Task 1) and `CreatedAt`/`UpdatedAt` comparisons;
    nil timestamps sort last, not panicking).
  - `TestBacklogMasterDetailFocusSwapMatchesD03BorderColors` (reuses the existing
    `!m.detailFocus`-parameterized `renderPane`, no new border logic).
  - `TestBacklogAccordionReusesTask1SectionsViaFocusedBean` (enter on the backlog list
    → `m.detailFocus == true`, `m.focusedBean()` returns the selected bean, NOT the
    tree cursor — proves `focusedBean()`'s Task-2 dispatcher is genuinely reused, not
    reimplemented).
  - `TestBacklogOpenFromTreeAndBackPreservesSharedFilterState` (`b` from Tree with an
    active facet filter → Backlog shows the SAME filtered view; `esc`/`b` back to Tree
    → filter still active there too).
  - `TestBacklogWindowingReusesExistingWindowAround` (many backlog items, small pane
    height, cursor near the end stays visible — behavioral proof, not a new
    implementation).
- [ ] **Step 2:** `command go test ./internal/tui/` → FAIL.
- [ ] **Step 3: Implement**:
  - `keymap.go`: `Sort keybind.Binding // S — cycle backlog sort mode`
    (`keybind.NewBinding(keybind.WithKeys("S"), keybind.WithHelp("S", "Sort"))`), added
    to `helpGroups()`'s "Actions" group (drift-guard enforces).
  - `types.go`: `viewID` gains `viewBacklog`; `backlogList listState`; `backlogSort
    string` (`""` == default == `idx.Backlog()`'s own canonical order, matching
    Task 4's `beanMatches` filter applied on top).
  - `view_browse_backlog.go`:

```go
package tui

// view_browse_backlog.go — V3 Backlog (design-spec.md §6): parentless+ready
// beans, Master-Detail, reusing Task 1's accordion (via focusedBean(), Task
// 2) and Task 3/4's shared search+facet predicate (m.beanMatches) -- NOT a
// second parallel filter implementation (devd has two, beans-tui has one).

func (m model) backlogVisible() []*data.Bean {
    var out []*data.Bean
    for _, b := range m.idx.Backlog() {
        if m.beanMatches(b) {
            out = append(out, b)
        }
    }
    sortBacklog(out, m.backlogSort) // no-op for "" (idx.Backlog() order already canonical)
    return out
}

func sortBacklog(beans []*data.Bean, mode string) {
    switch mode {
    case "priority":
        sort.SliceStable(beans, func(i, j int) bool {
            return data.PriorityRank(beans[i].Priority) < data.PriorityRank(beans[j].Priority)
        })
    case "created":
        sort.SliceStable(beans, func(i, j int) bool { return afterOrLast(beans[i].CreatedAt, beans[j].CreatedAt) })
    case "updated":
        sort.SliceStable(beans, func(i, j int) bool { return afterOrLast(beans[i].UpdatedAt, beans[j].UpdatedAt) })
    case "status":
        sort.SliceStable(beans, func(i, j int) bool {
            return data.StatusRank(beans[i].Status) < data.StatusRank(beans[j].Status)
        })
    } // "" -- leave idx.Backlog()'s own canonical order in place
}

// afterOrLast orders newest-first, nil timestamps sort last (never panics).
func afterOrLast(a, b *time.Time) bool {
    if a == nil { return false }
    if b == nil { return true }
    return a.After(*b)
}

func (m model) backlogSelected() *data.Bean {
    vis := m.backlogVisible()
    if m.backlogList.cursor < 0 || m.backlogList.cursor >= len(vis) { return nil }
    return vis[m.backlogList.cursor]
}

func (m model) viewBacklog() string {
    // ... Master-Detail: same renderPane/masterDetailWidths/windowAround/
    // outerBorder primitives as viewBrowseRepo (view_browse_repo.go), left
    // pane = backlogVisible() rows (icon+status+prio+ID+title, one line,
    // truncate — Scope-Cut vs. devd's multi-line blocks), right pane =
    // beanSections(m.idx, m.backlogSelected(), rw-4) + renderAccordion (Task 1/2).
}

func (m model) keyBacklog(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    if m.detailFocus { return m.keyDetailFocus(msg) } // Task 2, reused verbatim
    switch {
    case keybind.Matches(msg, keys.Sort):
        modes := []string{"", "priority", "created", "updated", "status"}
        // find current, advance, wrap -- "" and "status" are adjacent so the
        // cycle reads status->priority->created->updated->status to the PO
        // (idx.Backlog()'s own order IS status-tier order already).
    case keybind.Matches(msg, keys.Filter):
        m.filterItems = m.buildFilterItems() // Task 4, shared menu
        // ...
    case keybind.Matches(msg, keys.Search):
        // ... same searchInput wiring as keyTree (Task 3), shared model fields
    case keybind.Matches(msg, keys.Backlog), keybind.Matches(msg, keys.Back):
        m.view = viewBrowseRepo
    case keybind.Matches(msg, keys.Enter):
        if m.focusedBean() != nil { m.detailFocus = true; /* reset secCursor etc like Task 2's enterDetailFocus */ }
    }
    switch navKey(msg.String()) {
    case "up": m.backlogList.move(-1)
    case "down": m.backlogList.move(1)
    }
    return m, nil
}
```

  - `update.go`: `View()` switch gains `case viewBacklog: return
    m.viewBacklog()`; `handleKey` routes to `m.keyBacklog(msg)` when `m.view ==
    viewBacklog` (mirrors the existing `m.keyTree(msg)` routing for `viewBrowseRepo`);
    `keyTree`'s `b`-case (`keys.Backlog`) sets `m.view = viewBacklog;
    m.backlogList.setLen(len(m.backlogVisible()))`.
- [ ] **Step 4:** `command go test ./internal/tui/` → PASS.
- [ ] **Step 5: Golden** — new `TestBacklogGolden` (100×30, mirrors
  `tree_golden_test.go`'s exact pattern: `goldenBacklogModel` fixture, `-update` first
  pass writes `testdata/backlog.golden`, re-run confirms byte-identical, plus a
  `TestBacklogGoldenDeterministic` companion).
- [ ] **Step 6:** `command go test ./... && command gofmt -l . && command go vet ./...`
  clean, `make build` succeeds.
- [ ] **Step 7: Commit**

```
feat(tui): Backlog-View V3 (b) — Master-Detail, Sort-Toggle S

Backlog reused Task 1s Accordion (via focusedBean(), Task 2) und Task 3/4s
geteilten Such-/Filter-Zustand -- KEIN zweites Filter-Menü, KEINE zweite
Detail-Rendering-Logik. Sort-Toggle S zyklisch (status/priority/created/
updated, design-spec V3) statt eines schwebenden Sortier-Menüs (bewusste
Vereinfachung ggü. devd, kein V8-Overlay dafür in der Spec). Windowing
(E1 windowAround) wiederverwendet, devds variable-Höhen-Block-Fensterung
NICHT portiert (Scope-Cut, beans-Titel kurz genug für die bestehende
Einzeilen-Fensterung).

Refs: bt-gzu6
```

---

## Task 6: E2-Abschluss (`bt-enrd`)

Mirrors E1 Task 9 (`implementation-plan.md` »Epos-Rituale« → Epos-Abschluss).

- [ ] `command go test ./... -count=1` grün (2× hintereinander, wie E1s Abschluss-
  Nachweis), `command go build -o bin/bt .` ok, `command gofmt -l .` leer,
  `command go vet ./...` leer.
- [ ] Manueller Dogfooding-Smoke (tmux, wie T8/E1): `bin/bt` im eigenen Repo — Tree
  zeigt reale beans, `tab` in ein Bean mit Blocking-Relation, Ziffer `3` öffnet
  Beziehungen, `enter` auf einer Relation springt + Tree-Cursor folgt; `/` filtert
  live; `f` öffnet Facetten, ein Toggle schränkt Tree UND Backlog (`b`) gleich ein;
  `S` im Backlog zyklet die Sortierung. Terminal-Ausschnitt als Beleg in den
  Commit-Body (wie E1 T8/T9).
- [ ] beans pflegen: `beans update bt-ms0k -s completed`, ... `bt-gzu6 -s completed`
  (T1-T5), Epic `beans update bt-aq5s --tag to-review` (**NICHT** `-s completed` — PO-
  Gate, lean-stack-Prinzip "der ausführende Agent schließt NICHT").
- [ ] Selbst-Review (analog implementation-plan.md's Bottom-Section) im Commit-Body
  dokumentieren:
  - Spec-Coverage: V2 komplett (Master-Detail+Fokus+Beziehungs-Nav) ✓ · V3 Backlog ✓ ·
    V4 Accordion (Meta/Body/Beziehungen/Historie) ✓ · US-02/US-03/US-05/US-09 (design-
    spec §12) ✓.
  - I01/I03/Q01 geschlossen (Task 2 / Task 1+4 / Task 2) — I02 bewusst NICHT
    umgesetzt (Bean-Body: "Notiz, kein Auto-Fix"; `flattenTree`/`collectOrphans`
    bleiben Pro-Frame-Aufbau — kein Profiling-Anlass in E2).
  - Maus: bewusst NICHT E2 (design-spec §12 E5); Windowing (`windowStart`/
    `windowAround`, E1) in JEDER neuen Liste (Tree-Filter, Backlog) wiederverwendet,
    nirgends neu gebaut.
  - Konsolidierung ggü. devd: EIN Filter-Zustand/-Menü statt zwei (Task 4), EIN
    `focusedBean()`-Dispatcher statt separater Backlog-/Tree-Fokus-Logik (Task 2/5) —
    bewusste Vereinfachung, kein Feature-Verlust ggü. design-spec.
- [ ] Commit `docs: E2-Abschluss` (Refs: bt-enrd).
- [ ] Skill `ce-nsp-auto` → Handover-Prompt für E3 (Mutationen, bean `bt-gzcu`)
  erzeugen: Epos-Start-Ritual wiederholt sich (design-spec.md §7 + implementation-
  plan.md »E3« lesen, `epic-E3-plan.md` schreiben, Task-beans unter `bt-gzcu`).

---

## Selbst-Review (Plan gegen design-spec + Bean-Body-Pflichten)

- **5 Tasks aus implementation-plan.md »Epos E2«**: Accordion-Port (T1) ✓,
  Master-Detail-Verdrahtung (T2) ✓, Suche `/` (T3) ✓, Facetten-Filter `f`+`X` (T4) ✓,
  Backlog-View `b` (T5) ✓ — 1:1 abgedeckt, keine Aufgabe umbenannt/verschmolzen ohne
  Kennzeichnung.
- **Bean-`bt-aq5s`-Pflichtitems**: I01 (Task 2, Copy-on-Write-Konvention +
  Regressionstest) ✓ · I03 (Task 1 Vorbereitung: Export; Task 4 Abschluss: Anwendung
  auf Orphans) ✓ · Q01 (Task 2, nil-Guard + Test) ✓ · I02 EXPLIZIT nicht gefixt
  (Bean-Body sagt "Notiz, kein Auto-Fix" — Umsetzung wäre Scope-Creep) ✓ · Maus
  NACH Windowing (Windowing existiert seit E1, Maus bleibt E5, hier nirgends
  angefasst) ✓.
- **Windowing wiederverwendet, nicht neu gebaut**: Task 4 (Filter-gefilterter Tree)
  und Task 5 (Backlog) rufen beide das bestehende `windowAround`/`windowStart` (E1
  Task 8) auf; keine neue Fenster-Logik in diesem Plan.
- **data.Index nicht mutiert**: alle neuen Funktionen (`beanSections`,
  `flattenTreeFiltered`, `backlogVisible`) lesen nur `idx.ByID`/`idx.Children`/
  `idx.Backlog()` — Mutationen laufen weiterhin ausschließlich über `data.Client`
  (unverändert in E2, ETag-Handling ist E3-Scope).
- **keymap Single Source**: zwei neue Bindings (`Toggle` Task 4, `Sort` Task 5), beide
  in `helpGroups()` ergänzt — der bestehende `TestHelpGroupsCoverEveryBindingExactlyOnce`
  (E1 Task 7) fängt ein Vergessen automatisch ab, kein neuer Drift-Test nötig.
- **Golden-Determinismus**: jeder Task, der `View()`-Output ändert (T2 Detail-Pane,
  T3/T4 Suchkopfzeile, T5 neues Backlog-Golden), hat einen expliziten
  `-update`-Regenerierungs-Schritt + Determinismus-Re-Run — kein stiller Golden-Drift.
- **Kein Platzhalter/TBD**: jede Aufgabe hat konkrete Dateinamen, Funktionssignaturen,
  Testnamen und Port-Zeilenangaben (devd `file.go:NN-MM`); offene Design-Fragen (I01,
  I03, Sort-Menü-vs-Zyklus, Block-Windowing-Verzicht) sind mit Entscheidung + Begründung
  aufgelöst, nicht an die Implementierung delegiert.
