---
# bt-pqq3
title: T6 — Tag-Picker Suggest-Mode
status: completed
type: task
priority: normal
created_at: 2026-07-16T15:44:35Z
updated_at: 2026-07-16T17:08:53Z
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

- [x] `collectTagCounts` sortiert defined-first, Tie-Break unverändert
- [x] Marker-Spalte IMMER reserviert (PF-12-Test grün)
- [x] `openTagPicker` lädt Registry frisch bei jedem Open (D03)
- [x] freie Tags bleiben uneingeschränkt wählbar (kein strict mode —
  Regressionstest: ein NICHT-definierter Tag lässt sich weiterhin
  togglen/speichern)
- [x] B14-Freitext-Pfad unverändert (Q03 nicht umgesetzt)
- [x] bestehender Diff-/Mutation-Test-Korpus grün
- [x] Goldens Gegenbeleg grün
- [x] tmux-Smoke 120+80 belegt, Testdatei entfernt
- [x] voller Lauf grün
- [x] Commit `feat(tui): E10 Tag-Picker Suggest-Mode`

## Summary

`collectTagCounts` (`box_picker_tag.go`) wächst um einen `defined
map[string]bool`-Parameter (D10): `defined` wird NEUER PRIMÄRER Sortier-
Schlüssel, der bestehende Count-absteigend/alpha-Tie-Break bleibt sekundär
je Gruppe (neuer Helfer `sortTagCountsDefinedFirst`, geteilt mit dem
Neu-Tag-Einfüge-Pfad in `keyTagInput`, damit beide Stellen nie
auseinanderdriften). `collectTagCounts` liefert jetzt die UNION aus jedem
aktuell verwendeten Tag UND jedem Registry-definierten Tag — ein definierter
Tag mit Count 0 bleibt sichtbar (mirrort D09/T2s `tagRegistryRows` eine
Ebene höher; konkrete Grundlage für die tmux-Smoke-Vorgabe
"unbenutzt-definierte sichtbar"). `openTagPicker()` lädt die Registry frisch
via `m.client.LoadTagDefs()` (D03, nil-Client degradiert tolerant zu keiner
Definition, mirrort D02/T2s `openTagManagementPage`) und baut daraus die
`defined`-Lookup-Menge. `tagPickerBox()` bekommt eine IMMER reservierte
Marker-Spalte VOR der Checkbox (PF-12) — definierte Zeilen tragen
`tagManagementMarkerGlyph`/`tagManagementMarkerStyle` (✓, Green,
wiederverwendet aus `view_tag_management.go`/T2 statt eines zweiten
Glyphs, PF-12-Konsistenz), freie Zeilen den gleich breiten Leerraum. Der
B14-Freitext-Neu-Tag-Pfad (`n`) bleibt VERHALTENSMÄSSIG unverändert (schreibt
weiterhin nicht in die Registry, Q03 nicht umgesetzt) — nur seine interne
Re-Sortierung nutzt jetzt denselben geteilten Comparator.

## Test-Output (RED→GREEN wörtlich)

RED (`command go test ./internal/tui/... -run
"TestCollectTagCounts|TestTagPickerBoxReservesMarkerColumn|TestTagPickerBoxMarkerColumnWidthStable|TestOpenTagPickerLoadsRegistryFresh|TestOpenTagPickerNilClient|TestOpenTagPickerFreeTag"
-v -count=1`, vor der Implementierung):

```
# beans-tui/internal/tui [beans-tui/internal/tui.test]
internal/tui/box_picker_tag_test.go:62:31: too many arguments in call to collectTagCounts
	have (*data.Index, nil)
	want (*data.Index)
internal/tui/box_picker_tag_test.go:94:31: too many arguments in call to collectTagCounts
	have (*data.Index, map[string]bool)
	want (*data.Index)
internal/tui/box_picker_tag_test.go:95:62: got[0].defined undefined (type tagCount has no field or method defined)
internal/tui/box_picker_tag_test.go:98:44: got[1].defined undefined (type tagCount has no field or method defined)
internal/tui/box_picker_tag_test.go:116:31: too many arguments in call to collectTagCounts
	have (*data.Index, map[string]bool)
	want (*data.Index)
internal/tui/box_picker_tag_test.go:120:46: got[0].defined undefined (type tagCount has no field or method defined)
internal/tui/box_picker_tag_test.go:123:43: got[1].defined undefined (type tagCount has no field or method defined)
internal/tui/box_picker_tag_test.go:182:29: unknown field defined in struct literal of type tagCount
internal/tui/box_picker_tag_test.go:183:36: unknown field defined in struct literal of type tagCount
internal/tui/box_picker_tag_test.go:184:30: unknown field defined in struct literal of type tagCount
internal/tui/box_picker_tag_test.go:184:30: too many errors
FAIL	beans-tui/internal/tui [build failed]
FAIL
```

GREEN (gleiches `-run`-Filter, nach Implementierung, alle 8/8 grün):

```
=== RUN   TestCollectTagCountsSortedByCountDescThenAlpha
--- PASS: TestCollectTagCountsSortedByCountDescThenAlpha (0.00s)
=== RUN   TestCollectTagCountsDefinedFirstThenCountDescThenAlpha
--- PASS: TestCollectTagCountsDefinedFirstThenCountDescThenAlpha (0.00s)
=== RUN   TestCollectTagCountsIncludesUnusedDefinedTagAtCountZero
--- PASS: TestCollectTagCountsIncludesUnusedDefinedTagAtCountZero (0.00s)
=== RUN   TestOpenTagPickerLoadsRegistryFreshMarksDefinedIncludingUnused
--- PASS: TestOpenTagPickerLoadsRegistryFreshMarksDefinedIncludingUnused (0.00s)
=== RUN   TestOpenTagPickerNilClientDegradesToNoDefinedTags
--- PASS: TestOpenTagPickerNilClientDegradesToNoDefinedTags (0.00s)
=== RUN   TestOpenTagPickerFreeTagRemainsTogglableAndSavable
--- PASS: TestOpenTagPickerFreeTagRemainsTogglableAndSavable (0.03s)
=== RUN   TestTagPickerBoxReservesMarkerColumnRegardlessOfDefined
--- PASS: TestTagPickerBoxReservesMarkerColumnRegardlessOfDefined (0.00s)
=== RUN   TestTagPickerBoxMarkerColumnWidthStableAcrossNonCursorRows
--- PASS: TestTagPickerBoxMarkerColumnWidthStableAcrossNonCursorRows (0.00s)
PASS
ok  	beans-tui/internal/tui	0.533s
```

Vollständiger Tag-Picker-Korpus (bestehend + neu, 22 Tests) grün:
`command go test ./internal/tui/... -run "TestTagPicker|TestCollectTagCounts|TestOpenTagPicker|TestOverlayCaptureSwallowsQuitKeysWhileTagPickerOpen" -v -count=1` → alle PASS.

Gates: `command gofmt -l .` leer · `command go vet ./...` leer ·
`command go test ./... -short -count=1` grün (alle Pakete ok) ·
`command go test ./... -count=1` grün (Voll-Lauf, `internal/tui` 138.712s,
exit 0).

## Golden-Beleg

`command go build -o bin/bt .` erfolgreich, danach `command go test
./internal/tui/... -run "TestTreeGolden|TestBacklogGolden|TestChromeGolden"
-count=2` → alle 5 Testfunktionen (inkl. `*Deterministic`-Varianten) PASS,
beide Wiederholungen. `git diff --stat -- internal/tui/testdata/` leer —
Basis-Goldens unverändert (Picker ist Overlay-Zustand, kein eigener
Golden-Fall laut design-spec.md §11, wie erwartet).

## Smoke

tmux, `TERM=xterm-256color`, Breite 120 UND 80 (Sessions `bt120pqq3`/
`bt80pqq3`, `bin/bt` in diesem Repo). Temporäre `.beans-tags.yml`
(3 definierte Tags: `to-review`/`smoke` — beide real in diesem Repo
verwendet — plus `archived`, absichtlich UNBENUTZT) im Repo-Root angelegt,
Fokus auf `bt-apmy` (trägt real die Tags `to-review`+`smoke`), `t` gedrückt:

```
│ Tags                                   │
│ space/x:toggle  n:new tag  enter:save  │
│  esc:discard                           │
│ ▸ ✓ [x] to-review (10)                 │
│   ✓ [x] smoke (1)                      │
│   ✓ [ ] archived (0)                   │
│     [ ] rejected (1)                   │
```

Bestätigt: definierte Tags (`to-review`/`smoke`/`archived`) priorisiert
oben, Marker `✓` sichtbar, `archived` trotz Count 0 sichtbar
("unbenutzt-definierte sichtbar"), freier Tag `rejected` darunter ohne
Marker. Freitext-Neu-Tag (`n`, Tippen von `smoke-t6-probe`, enter) legt
korrekt eine WEITERE freie (nicht markierte) Zeile ans Ende der freien
Gruppe, sofort pending=`[x]`:

```
│   ✓ [x] to-review (10)                 │
│   ✓ [x] smoke (1)                      │
│   ✓ [ ] archived (0)                   │
│     [ ] rejected (1)                   │
│ ▸   [x] smoke-t6-probe (0)             │
```

Freier Tag `rejected` per Space getoggelt (`[ ]`→`[x]`) — belegt "kein
strict mode": ein NICHT-definierter Tag bleibt uneingeschränkt wählbar.
Jede Änderung per `esc` verworfen (`git status --porcelain .beans/` nach
JEDEM Discard leer bestätigt — kein reales Bean in diesem Repo mutiert).
Bei 80 Spalten identisches Bild, Modal (`clampModalWidth(40, ...)`) bleibt
vollständig innerhalb des 80-Spalten-Rahmens, Checkbox/Name nicht
abgeschnitten. Beide Sessions sauber beendet (`q`→enter, kein `bt`-Prozess
mehr aktiv). Testdatei `.beans-tags.yml` und `bin/bt` danach entfernt,
`git status --porcelain` zeigt nur die beiden Quelldateien.

## Deviations/ERRATA

**B-ERRATUM (gefunden+gefixt+regressionsgetestet im eigenen Testlauf, kein
PO-sichtbares Verhalten betroffen, mirrort bean bt-r92i's eigenen
B01-Präzedenzfall):** die bean-eigene RED-Test-Vorlage
`TestTagPickerBoxReservesMarkerColumnRegardlessOfDefined` hatte zwei
Test-Bugs, unabhängig vom Produktionscode:

1. Der Zeilen-Filter (`strings.Contains(l, "a") || strings.Contains(l,
   "b ")`) matchte auch den Modal-Titel ("Tags", enthält "a") und die
   Hint-Zeile ("space/x:toggle ... enter:save", enthält "a" — bei
   `modalBox`-Breite 40 zusätzlich auf ZWEI physische Zeilen umgebrochen,
   keine davon enthält "["). Beide Nicht-Tag-Zeilen sortierten in `lines`
   VOR den zwei echten Tag-Zeilen, sodass `rowLines[0]`/`rowLines[1]`
   tatsächlich zwei Nicht-Tag-Zeilen waren, deren "["-Suche beide -1
   liefert — ein VAKUOSER Pass (-1 == -1), der nichts prüfte. Fix: Filter
   direkt auf `strings.Contains(l, "[")` (die tatsächlich geprüfte
   Eigenschaft: eine Checkbox-Zeile).
2. `strings.Index` liefert einen BYTE-Offset; `tagManagementMarkerGlyph`
   ("✓") und der bestehende Cursor-Glyph "▸" sind beide 3-Byte-UTF-8-
   Sequenzen, die je EINE Terminal-Zelle belegen — wie das ASCII-Leerzeichen,
   das sie ersetzen. Eine markierte+cursorierte Zeile und eine freie,
   nicht-cursorierte Zeile können damit IDENTISCHE Anzeige-Spalten, aber
   UNTERSCHIEDLICHE Byte-Offsets haben. Fix: neuer Test-Helfer
   `checkboxColStart` vergleicht über `lipgloss.Width`, nicht über rohen
   Byte-Offset — deckt sich wörtlich mit PF-12s eigenem Test-Vertrag
   (design-spec.md §15, im Epic-Body bt-362n zitiert: "`lipgloss.Width`
   ... identisch"). Zusätzlich ergänzt:
   `TestTagPickerBoxMarkerColumnWidthStableAcrossNonCursorRows` — isoliert
   NUR die Marker-Spalte (zwei NICHT-cursorierte Zeilen, eine definiert,
   eine frei), unabhängig vom Cursor-Glyph-Byte-Offset-Confound.

Sonst keine Abweichungen. Zusätzlich zu den 2 im Bean-Body zitierten
RED-Zitat-Tests wurden 5 weitere Tests ergänzt (Akzeptanz-Checkliste
verlangt sie sinngemäß, ohne eigenes RED-Zitat):
`TestCollectTagCountsIncludesUnusedDefinedTagAtCountZero`,
`TestOpenTagPickerLoadsRegistryFreshMarksDefinedIncludingUnused`,
`TestOpenTagPickerNilClientDegradesToNoDefinedTags`,
`TestOpenTagPickerFreeTagRemainsTogglableAndSavable`,
`TestTagPickerBoxMarkerColumnWidthStableAcrossNonCursorRows` — reine
Ergänzung, keine Abweichung vom Spezifizierten.

## Notes for T3

- `sortTagCountsDefinedFirst` (`box_picker_tag.go`) ist der EINE geteilte
  Comparator für `[]tagCount` (defined-first, dann count-desc, dann alpha)
  — falls T3/T4/T5 je einen ähnlichen `[]tagRegistryRow`-Sortierbedarf
  bekommen sollten, bitte NICHT diesen Helfer direkt wiederverwenden
  (anderer Typ, `tagRegistryRow` vs. `tagCount`), aber das Comparator-Muster
  spiegeln statt neu zu erfinden.
- `tagManagementMarkerGlyph`/`tagManagementMarkerStyle`
  (`view_tag_management.go`, T2) werden jetzt an ZWEI Stellen verwendet
  (Tag-Management-Page UND Tag-Picker) — jede künftige Änderung an
  Glyph/Farbe (z. B. durch T3-T5) wirkt sich auf BEIDE Stellen aus,
  PF-12-Konsistenz bewusst gewollt, nicht versehentlich gekoppelt.
- `collectTagCounts(idx, defined)` ist jetzt die EINZIGE Zähl-Funktion, die
  die D09-Union (genutzt + definiert, auch Count 0) beherrscht — T2s
  `tagRegistryRows` (`view_tag_management.go`) hat weiterhin ihre EIGENE,
  separate Zähl-Schleife (ERRATUM-Notiz aus bt-r92i, bewusst NICHT in
  diesem Task vereinheitlicht, disjunkter Datei-Scope). Ein künftiger
  Task (T7 oder Fast-Follow) KÖNNTE `tagRegistryRows` auf `collectTagCounts`
  umstellen, ist aber nicht Teil dieses Tasks.
