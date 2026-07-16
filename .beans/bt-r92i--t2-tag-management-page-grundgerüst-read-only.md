---
# bt-r92i
title: T2 — Tag-Management-Page Grundgerüst (read-only)
status: in-progress
type: task
priority: normal
created_at: 2026-07-16T15:44:24Z
updated_at: 2026-07-16T16:01:31Z
parent: bt-362n
blocked_by:
    - bt-49hh
---

T2 — Page-Grundgerüst: `viewTagManagement`, read-only Liste (Epic
`bt-362n`, D05-D09). `blocked_by` T1 (`internal/data/tagdefs.go` muss
existieren). Liefert die Page als NAVIGIERBARES, aber noch reines
Lese-Feature — Create/Delete/Rename folgen in T3/T4/T5.

## Ziel

Neuer Top-Level-View, erreichbar über das Command-Center („go to tags"),
zeigt die UNION aus definierten + frei-in-Verwendung-Tags (D09) als
Einzel-Pane-Liste mit Verwendungszähler. Kein Master-Detail (D08),
`enter` ist ein dokumentierter No-Op (reserviert für ein späteres
Drilldown-Fast-Follow).

## Betroffene Dateien/Symbole

- `internal/tui/types.go`:
  - `viewID`-Enum (`const ( viewBrowseRepo ... )`) wächst um
    `viewTagManagement` (neuer, letzter Wert — NIEMALS bestehende Werte
    umsortieren, das würde jeden persistierten/getesteten Vergleich
    gegen den iota-Wert brechen, obwohl aktuell nichts iota numerisch
    persistiert — Vorsichtsmaßnahme, mirrort den Kommentar-Stil der
    bestehenden drei Werte).
  - `model` bekommt: `tagMgmtRows []tagRegistryRow`,
    `tagMgmtCursor listState` (Cursor über die gerenderte Zeilenliste,
    reuse `listState` wie `backlogList`).
- `internal/tui/view_tag_management.go` (neu):
  - `type tagRegistryRow struct { name string; count int; defined bool }`
  - `func tagRegistryRows(idx *data.Index, defs []string) []tagRegistryRow`
    — D09-Sortierung: definierte Gruppe zuerst (alpha), dann
    undefinierte-in-Verwendung (Count absteigend, dann alpha als
    Tie-Break, mirrort `collectTagCounts`s eigenen Tie-Break). Nutzt
    intern dieselbe Zähl-Logik wie `collectTagCounts`
    (`box_picker_tag.go`) — Planner-Entscheidung: `collectTagCounts`
    NICHT umbenennen/verschieben (T6 erweitert es ohnehin um den
    `defined`-Parameter, danach kann `tagRegistryRows` diese ERWEITERTE
    Fassung direkt aufrufen statt eine zweite Zähl-Implementierung zu
    pflegen — s. T2s eigene ERRATUM-Notiz unten, falls T2 vor T6 landet).
  - `func (m model) openTagManagementPage() (tea.Model, tea.Cmd)` — lädt
    Registry frisch (`m.client.LoadTagDefs()`, D03, synchron), baut
    `tagMgmtRows` via `tagRegistryRows(m.idx, defs)`, setzt `m.view =
    viewTagManagement`, resettet `tagMgmtCursor` (`setLen`).
  - `func (m model) tagManagementChrome(innerW int) (head, localKeys string)`
    — mirrort `backlogChrome`: `breadcrumb(m.repoLabel(), "Tags", "",
    innerW)` (D07: GlobalHint LEER, nicht `renderBindings(globalBindings())`)
    + `footer(renderBindings(tagManagementLocalBindings()), innerW)`.
  - `func tagManagementLocalBindings() []keybind.Binding` — in T2 vorerst
    `{keys.Up, keys.Down, keys.Back}` (T3/T4/T5 hängen ihre eigenen
    Bindings anschließend an, s. deren Task-Bodies — EIN gemeinsamer
    Funktionsrumpf, kein Duplikat je Task).
  - `func (m model) viewTagManagement() string` — mirrort `viewBacklog()`s
    Grundgerüst (innerW/innerH-Berechnung, `Chrome`-Bausteine,
    `renderPane` für die Zeilenliste, `outerBorder`, `composeOverlays`
    am Ende) — EIN Pane, keine `JoinHorizontal`. Jede Zeile:
    Marker-Spalte (D10-Konvention vorgezogen: reserviert, hier noch ohne
    Picker-Bezug — s. Akzeptanz) + Name + rechtsbündiger Count.
  - `func (m model) keyTagManagement(msg tea.KeyMsg) (tea.Model, tea.Cmd)`
    — Up/Down bewegen `tagMgmtCursor` (reuse `navKey`), `enter` = handled
    No-Op (D08, mit Kommentar der das Fast-Follow referenziert), `esc`
    (`keys.Back`) → `m.view = viewBrowseRepo` (D03/D06-Pendant: EINE
    Ebene zurück, mirrort Lobbys `esc`).
- `internal/tui/update.go` (`handleKey`): neuer Check `if m.view ==
  viewTagManagement { return m.keyTagManagement(msg) }` — EXAKT an
  derselben Stelle wie der bestehende `if m.view == viewLobby { return
  m.keyLobby(msg) }`-Block (Zeile ~879, D06: full-capture, VOR den
  ctrl+k/?/p-Bare-Matches).
- `internal/tui/overlay_palette.go`: `paletteActions` bekommt einen
  neuen globalen Eintrag `paletteItem{kind: paletteKindAction, actionID:
  "go_tags", label: "go to tags"}` (gruppiert neben `"go to repo
  picker"`/`"go to settings"`, gleiche Sektion — Reihenfolge:
  Planner-Entscheidung, direkt VOR `"settings"` als letztem Eintrag,
  mirrort wie `"repo_picker"` vor `"settings"` einsortiert wurde);
  `dispatchPalette` bekommt den `case "go_tags": return
  m.openTagManagementPage()`.
- `internal/tui/view_browse_repo.go` (`View()`-Dispatcher, Zeile ~700):
  neuer `case viewTagManagement: return m.viewTagManagement()`.

## TDD (RED zuerst)

```go
// view_tag_management_test.go (neu)
func TestTagRegistryRowsDefinedFirstAlphaThenFreeByCountDesc(t *testing.T) {
    idx := data.NewIndex([]data.Bean{
        {ID: "b1", Tags: []string{"zzz-free"}},
        {ID: "b2", Tags: []string{"zzz-free"}},
        {ID: "b3", Tags: []string{"aaa-free"}},
    })
    rows := tagRegistryRows(idx, []string{"defined-b", "defined-a"})
    want := []string{"defined-a", "defined-b", "zzz-free", "aaa-free"}
    var got []string
    for _, r := range rows {
        got = append(got, r.name)
    }
    if !reflect.DeepEqual(got, want) {
        t.Fatalf("want %v, got %v", want, got)
    }
}

func TestTagRegistryRowsIncludesUnusedDefinedTagWithZeroCount(t *testing.T) {
    idx := data.NewIndex(nil)
    rows := tagRegistryRows(idx, []string{"unused"})
    if len(rows) != 1 || rows[0].count != 0 || !rows[0].defined {
        t.Fatalf("want one zero-count defined row, got %+v", rows)
    }
}
```

```go
// update_test.go oder view_tag_management_test.go
func TestKeyTagManagementEscReturnsToBrowse(t *testing.T) {
    m := newModel(nil, "")
    m.view = viewTagManagement
    nm, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
    if nm.(model).view != viewBrowseRepo {
        t.Fatalf("want viewBrowseRepo, got %v", nm.(model).view)
    }
}

func TestHandleKeyOnTagManagementViewDoesNotLeakToNodeAction(t *testing.T) {
    // Regression guard for D06's own rationale: `d` while on the tag
    // page must never open the bean Delete-Confirm.
    m := newModel(nil, "")
    m.view = viewTagManagement
    nm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
    if nm.(model).overlay != overlayNone {
        t.Fatalf("want no overlay opened, got %v", nm.(model).overlay)
    }
}
```

## Golden-Strategie

GEGENBELEG für Tree/Backlog/Chrome (dieser Task fügt einen NEUEN
Render-Pfad hinzu, berührt aber keinen bestehenden — nach `command go
build -o bin/bt .`: `command go test ./internal/tui/ -run
"TestTreeGolden|TestBacklogGolden|TestChromeGolden"` OHNE `-update`, MUSS
grün bleiben, im Commit-Body explizit als „unverändert" festhalten). Ob
`viewTagManagement()` eine EIGENE Golden-Suite bekommt, ist
Implementer-Entscheidung (mirrort E9 T7/T8s Freiheit) — mindestens ein
paar `View()`-Snapshot-Tests sind sinnvoll, da dies der EINZIGE
Render-Pfad ist, den dieser Task neu einführt.

## tmux-Smoke (120 UND 80 Spalten)

`bin/bt` in diesem Repo starten, `ctrl+k` → „tags" tippen → enter → Page
öffnet, Liste zeigt aktuell (noch leere Registry, D02: keine
`.beans-tags.yml` in diesem Repo) NUR die undefinierten/in-Verwendung-Tags
(z. B. `to-review`, falls ein Bean das Tag trägt). `esc` → zurück zu
Browse. Wiederholen bei 80 Spalten — Zeilen dürfen nicht umbrechen/
abschneiden ohne `…`-Truncation. Danach `git status --porcelain` leer
(kein `.beans-tags.yml` durch den bloßen Page-Besuch angelegt — T2 ist
rein lesend).

## Akzeptanz-Checkliste

- Page über Palette „go to tags" erreichbar · Liste zeigt Union
  definiert+frei (D09-Sortierung) · Verwendungszähler korrekt ·
  `enter` dokumentierter No-Op · `esc` → Browse · `handleKey`-Leak-Test
  (D06-Regression) grün · GlobalHint im Header leer (D07) · kein neuer
  `.beans-tags.yml` durch bloßes Öffnen · Goldens Gegenbeleg grün ·
  tmux-Smoke 120+80 belegt · voller Lauf grün · Commit `feat(tui): E10
  Tag-Management-Page — Grundgerüst`.

## PRELUDE aus T1-Review (2026-07-16, F03, low)

Als erster eigener Commit (test-only): `TestLoadTagDefsSkipsInvalidNamesDefensively` (internal/data/tagdefs_test.go:25-34) — der dritte YAML-Eintrag (leerer Scalar) wird von yaml.v3 schon VOR dem Unmarshal verworfen (empirisch len(f.Tags)==2, nicht 3); der Test beweist damit nur die Bad_Tag-Filterung, nicht den Leer-String-Fall. Test um einen echten Leer-String ergänzen, der bis zum Filter durchreicht (z.B. YAML-Zeile `- ""`).
