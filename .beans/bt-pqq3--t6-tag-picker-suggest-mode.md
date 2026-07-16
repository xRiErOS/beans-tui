---
# bt-pqq3
title: T6 — Tag-Picker Suggest-Mode
status: in-progress
type: task
priority: normal
created_at: 2026-07-16T15:44:35Z
updated_at: 2026-07-16T16:50:28Z
parent: bt-362n
blocked_by:
    - bt-49hh
---

T6 — Tag-Picker Suggest-Mode (Epic `bt-362n`, D10, PF-12-konform).
`blocked_by` T1 (`internal/data/tagdefs.go`) — NUR T1, bewusst NICHT T2-T5
(disjunkter Datei-Scope, `box_picker_tag.go` vs. `view_tag_management.go`
— maximale Parallelität, mirrort E9s Fan-out-Philosophie „viele Tasks
hängen nur an der frühesten fundamentalen Task").

## Ziel

Der bestehende Tag-Picker (`t`, `box_picker_tag.go`) bietet definierte
Tags priorisiert an (Suggest-Mode, PO-Fix-Vorgabe) — freie Tags bleiben
weiterhin uneingeschränkt wählbar (kein strict mode, PO-Vorgabe wörtlich).

## Betroffene Dateien/Symbole

- `internal/tui/box_picker_tag.go`:
  - `collectTagCounts(idx *data.Index, defined map[string]bool)
    []tagCount` — Signatur wächst um `defined` (Planner-Entscheidung:
    KEIN zweiter Funktionsname/keine Overload — Go kennt kein Overload,
    ALLE bestehenden Aufrufer werden mitgezogen, aktuell nur
    `openTagPicker`; `view_tag_management.go`s `tagRegistryRows`, T2,
    ruft ab jetzt DIESE erweiterte Fassung auf statt eine zweite
    Zähl-Logik zu pflegen — T2 wurde entsprechend mit einer
    ERRATUM-Notiz vorgezeichnet, falls T2 vor T6 landet: dann zählt
    `tagRegistryRows` zunächst separat und wird in DIESEM Task auf den
    gemeinsamen Aufruf umgestellt, EIN kleiner Zusatz-Diff in
    `view_tag_management.go`).
  - `type tagCount struct { tag string; count int; defined bool }` —
    neues Feld `defined`.
  - Sortierung (`sort.Slice`, `collectTagCounts`): `defined` wird NEUER
    PRIMÄRER Schlüssel (`out[i].defined != out[j].defined:
    return out[i].defined` — defined-Zeilen zuerst), der bestehende
    Count-absteigend/alpha-Tie-Break bleibt SEKUNDÄR (unverändert
    innerhalb jeder der beiden Gruppen).
  - `openTagPicker()`: lädt die Registry frisch (`m.client.LoadTagDefs()`,
    D03) VOR dem `collectTagCounts`-Aufruf, baut eine
    `map[string]bool`-Lookup-Menge daraus.
  - `tagPickerBox()`: jede Zeile bekommt eine IMMER reservierte
    Marker-Spalte VOR dem bestehenden `[ ]`/`[x]`-Checkbox-Feld (PF-12,
    D10 — `"● "` für `it.defined`, `"  "` (zwei Leerzeichen, gleiche
    sichtbare Breite) sonst; NIE ein bedingtes Weglassen der Spalte
    selbst, nur der Glyph wechselt — mirrort PF-12s eigenen Wortlaut
    „alle markierbaren Zeilen" + die dort verlangte Test-Pflicht
    „`lipgloss.Width` einer NICHT-aktiven Zeile identisch, unabhängig
    von einer anderen aktiven Zeile" sinngemäß auf „Marker-Spalte
    identisch breit, unabhängig von `defined`" übertragen).
  - Neu-Tag-Freitext-Pfad (`keyTagInput`, B14-Erbe): **UNVERÄNDERT**
    (Q03 im Epic-Body — ausdrücklich NICHT Teil dieses Tasks, kein
    Registry-Write von hier aus).

## TDD (RED zuerst)

```go
func TestCollectTagCountsDefinedFirstThenCountDescThenAlpha(t *testing.T) {
    idx := data.NewIndex([]data.Bean{
        {ID: "b1", Tags: []string{"popular-free"}},
        {ID: "b2", Tags: []string{"popular-free"}},
        {ID: "b3", Tags: []string{"popular-free"}},
        {ID: "b4", Tags: []string{"rare-defined"}},
    })
    defined := map[string]bool{"rare-defined": true}
    got := collectTagCounts(idx, defined)
    if len(got) != 2 || got[0].tag != "rare-defined" || !got[0].defined {
        t.Fatalf("want defined tag first despite lower count, got %+v", got)
    }
    if got[1].tag != "popular-free" || got[1].defined {
        t.Fatalf("want free tag second, got %+v", got)
    }
}

func TestTagPickerBoxReservesMarkerColumnRegardlessOfDefined(t *testing.T) {
    m := newModel(nil, "")
    m.tagItems = []tagCount{{tag: "a", defined: true}, {tag: "b", defined: false}}
    m.menu.setLen(2)
    out := m.tagPickerBox()
    lines := strings.Split(ansi.Strip(out), "\n")
    // Beide Tag-Zeilen (nach der Hint-Zeile) müssen an identischer Spalte
    // beginnen -- PF-12: kein bedingtes Weglassen der Marker-Spalte.
    var rowLines []string
    for _, l := range lines {
        if strings.Contains(l, "a") || strings.Contains(l, "b ") {
            rowLines = append(rowLines, l)
        }
    }
    if len(rowLines) < 2 {
        t.Fatalf("want at least 2 tag rows rendered, got %v", rowLines)
    }
    // Position des '[' (Checkbox-Start) muss in beiden Zeilen identisch sein.
    if strings.Index(rowLines[0], "[") != strings.Index(rowLines[1], "[") {
        t.Fatalf("marker column shifted layout: %q vs %q", rowLines[0], rowLines[1])
    }
}
```

Weitere Pflicht-Tests: `openTagPicker` lädt die Registry (Fake-Client mit
vorbereiteter `.beans-tags.yml`, prüft `tagItems[i].defined` korrekt
gesetzt); bestehender `TestApplyTagPickerDiff...`-artiger Test-Korpus
bleibt GRÜN OHNE Änderung (reine Sortier-/Anzeige-Erweiterung, keine
Diff-/Mutation-Semantik-Änderung).

## Golden-Strategie

GEGENBELEG für Tree/Backlog/Chrome (unverändert). Der Tag-Picker selbst
hat KEINE eigene Golden-Suite (Overlays sind laut design-spec.md §11
nicht Teil der 3 Basis-Goldens) — Verhaltenstests (oben) sind hier die
alleinige Absicherung, kein Golden-Schritt zusätzlich nötig.

## tmux-Smoke (120 UND 80 Spalten)

In diesem Repo (rein lesend, T6 mutiert keine Beans): `.beans-tags.yml`
temporär mit 1-2 existierenden Tag-Namen aus diesem Repo anlegen (z. B.
`to-review`), `t` auf einem beliebigen Bean öffnen → die definierten
Zeilen erscheinen oben, Marker-Glyph sichtbar, Rest wie gehabt sortiert.
Bei 80 Spalten: Marker-Spalte darf die Checkbox/den Namen nicht aus dem
Modal-Rand drücken (`clampModalWidth`, unverändert). Testdatei danach
löschen.

## Akzeptanz-Checkliste

- `collectTagCounts` sortiert defined-first, Tie-Break unverändert ·
  Marker-Spalte IMMER reserviert (PF-12-Test grün) · `openTagPicker` lädt
  Registry frisch bei jedem Open (D03) · freie Tags bleiben uneingeschränkt
  wählbar (kein strict mode — Regressionstest: ein NICHT-definierter Tag
  lässt sich weiterhin togglen/speichern) · B14-Freitext-Pfad unverändert
  (Q03 nicht umgesetzt) · bestehender Diff-/Mutation-Test-Korpus grün ·
  Goldens Gegenbeleg grün · tmux-Smoke 120+80 belegt, Testdatei entfernt ·
  voller Lauf grün · Commit `feat(tui): E10 Tag-Picker Suggest-Mode`.
