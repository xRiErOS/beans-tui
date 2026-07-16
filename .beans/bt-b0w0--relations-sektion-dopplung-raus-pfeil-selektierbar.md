---
# bt-b0w0
title: 'RELATIONS-Sektion: Dopplung raus, Pfeil-selektierbar, hängender Einzug (B04)'
status: completed
type: task
priority: normal
created_at: 2026-07-16T06:45:47Z
updated_at: 2026-07-16T11:04:44Z
parent: bt-tct9
---

E9 Task 4 — deckt B04 (alle drei Teilpunkte) aus bean bt-tct9. Quelle: design-spec.md §15
PF-17 (Abschnitt B04). Ist-Code: internal/tui/accordion.go (renderAccordion, fieldStrip),
internal/tui/view_detail_bean.go (relationsSectionBody, relationRow, beanSections,
metaSectionBody als Vorbild-Muster), internal/tui/update.go (keyDetailFocus's Pfeil-
Navigation über secs[...].fields — LIEST bereits die richtige Slice, keine Änderung dort
nötig). Kein blocked_by. WICHTIG: bt-tct9 Task 5 (B06, Relations-Picker-Breite) UND
bt-tct9 Task 7 (F01a, Vollbild — braucht B04.2 für den Relations-Sprung im Vollbild-Detail)
hängen `blocked_by` an DIESEM Task — beide führen erst NACH diesem Task aus.

## B04 — RELATIONS-Sektion: Dopplung raus, Pfeil-selektierbar, hängender Einzug

Drei zusammenhängende Punkte, EINE Änderung an `relationsSectionBody` + `renderAccordion`:

1. Die separate `Fields:`-Zeile (`fieldStrip`, accordion.go) ist redundant zur Parent/
   Children/Blocked-By-Darstellung darunter — raus.
2. Pfeiltasten selektieren HEUTE nur die (jetzt entfernte) `Fields:`-Strip — die Einträge
   in Parent/Children/Blocked By selbst sind nicht per Pfeil erreichbar/wählbar. Nach dem
   Fix: dieselbe Selektion muss an den ECHTEN Zeilen sichtbar sein.
3. Layout: Bean-Titel müssen mit HÄNGENDEM EINZUG umbrechen, sodass die Meta-Spalten
   (type-Glyph, Status, ID) nie vom Titeltext unterwandert werden. PO-Mockup:

       t M bt-apmy Hier steht ein langer Titel eines beans
                   der so umbricht, dass die Uebersicht gewahrt
                   ist
       c T bt-2jve Hier steht ein kurzer Titel eines beans

## Architektur-Vorgabe

**1. `fieldStrip` entfernen.** `renderAccordion`s (accordion.go) `if activeSec && i != 0 &&
len(s.fields) > 0 { b.WriteString(fieldStrip(...)) }`-Zweig entfällt — RELATIONS
(`relationsSectionIdx`) war der EINZIGE verbleibende Aufrufer (Meta ist per `i != 0`
strukturell bereits ausgeschlossen; Body/History haben keine `.fields`). `fieldStrip`
selbst wird komplett gelöscht (Compiler-gesteuerte Verifikation, kein toter Code — Muster
design-spec.md §15 PF-14/B13-Removal).

**2. Ersatz: `relationsSectionBody` bekommt dieselbe ▷/▶-Cursor-Konvention wie
`metaSectionBody`.** Signatur wächst um zwei Parameter, mirrort `metaSectionBody(b
*data.Bean, bodyW int, active bool, fieldIdx int)`:

```go
// VORHER:
func relationsSectionBody(idx *data.Index, b *data.Bean, bodyW int) (string, []relationField)

// NACHHER:
func relationsSectionBody(idx *data.Index, b *data.Bean, bodyW int, active bool, fieldIdx int) (string, []relationField)
```

Jede gerenderte Zeile (Parent/Children/Blocking/Blocked-By, in der BESTEHENDEN `fields`-
Akkumulationsreihenfolge — `appendGroup` hängt `lines`/`fields` bereits heute im
Lockschritt an, KEINE Änderung an dieser Reihenfolge-Logik nötig) trägt einen eigenen
`▷ `/`▶ `-Marker VOR dem bisherigen `relationRow`-Text — exakt wie `metaSectionBody`s
Zeilen-Schleife:

```go
marker := theme.Muted.Render("▷ ")
if active && fieldIdx == globalRowIndex { marker = theme.Accent.Render("▶ ") }
```

`globalRowIndex` zählt über ALLE Gruppen hinweg fortlaufend (Parent=0, Children=1..n,
Blocking=n+1..m, Blocked-By=m+1..k) — identisch zum Index, den `fields[]` bereits heute
in genau dieser Reihenfolge trägt (`fields[i]` UND die i-te sichtbare Zeile sind
bereits 1:1 verknüpft, nur bisher OHNE eigene visuelle Markierung). Die Pfeiltasten-
Navigation selbst (`keyDetailFocus`s `up`/`down` bei `detailLevel==1`, `fieldCursor` über
`secs[m.secCursor].fields`) ändert KEINEN Code — sie zeigte immer schon auf dieselbe
`fields`-Slice; der PO-Befund "nur Fields wählbar" löst sich auf, weil es danach nur noch
EINE Repräsentation gibt (die Zeilen selbst), nicht mehr zwei (Strip + Zeilen).

**3. Hängender Einzug.** Neuer Helfer (view_detail_bean.go, neben `relationRow`):

```go
// hangingIndentWrap wraps text to width w, with the FIRST line prefixed by
// prefix (already-styled; its VISIBLE width via lipgloss.Width(ansi.Strip(
// prefix)) becomes the indent) and every CONTINUATION line left-padded with
// that many spaces instead of the wrap restarting at column 0 (B04.3,
// design-spec.md §15 PF-17) -- indent is computed PER ROW (not a shared
// table column across all rows: bean IDs have variable prefix length,
// "<prefix>-<n-chars>").
func hangingIndentWrap(prefix, text string, w int) string {
    indentW := lipgloss.Width(ansi.Strip(prefix))
    contW := w - indentW
    if contW < 8 { contW = 8 } // never collapse to nothing on narrow terminals
    wrapped := ansi.Wordwrap(text, contW, "")
    lines := strings.Split(wrapped, "\n")
    var b strings.Builder
    for i, line := range lines {
        if i == 0 {
            b.WriteString(prefix + line)
        } else {
            b.WriteString("\n" + strings.Repeat(" ", indentW) + line)
        }
    }
    return b.String()
}
```

`relationRow` (heute: `theme.StatusIcon(rel.Status) + " " + theme.TypeIcon(rel.Type) + " "
+ theme.Key.Render(rel.ID) + " " + rel.Title`, bare Konkatenation, KEIN Wrap) wird für die
RELATIONS-Sektion (UND Task 5/B06s Picker, die dieselbe Zeilenform brauchen) durch einen
`▷`/`▶`-Marker + `hangingIndentWrap`-Aufruf ersetzt: `prefix = marker +
theme.StatusIcon(...) + " " + theme.TypeIcon(...) + " " + theme.Key.Render(id) + " "`,
`text = rel.Title`. `relationsSectionBody`s bisheriges Schluss-`wrapText(strings.Join(
groups, "\n\n"), bodyW)` entfällt für die Zeilen selbst (die sind jetzt VOR dem Join
bereits fertig eingezogen umgebrochen) — `wrapText` bleibt NUR noch für die (unveränderten)
Muted-Subheader-Zeilen (`Parent`/`Children`/`Blocking`/`Blocked By`) relevant, falls
überhaupt (die sind kurze feste Strings, wrappen praktisch nie).

`beanSections` (view_detail_bean.go) reicht die zwei neuen Parameter durch: `focused &&
activeIdx == relationsSectionIdx`, `fieldIdx` — exakt das gleiche Muster wie
`metaSectionBody`s eigener Aufruf in derselben Funktion (Zeile direkt darüber).

## TDD-Schritte

1. Failing tests (view_detail_bean_test.go): `TestRelationsSectionBodyShowsCursorMarkerOnActiveRow`
   (active=true, fieldIdx=1 → Zeile 1 trägt `▶`, alle anderen `▷`); `TestHangingIndentWrapAlignsContinuationUnderTitleStart`
   (langer Titel, prüft dass Folgezeilen exakt `indentW` Leerzeichen tragen);
   `TestRelationsSectionBodyNoLongerRendersFieldsStripLine` (Regressionstest: kein "Fields:"
   im Output mehr); accordion_test.go: `TestRenderAccordionOmitsFieldStripForRelations`
   (Regressionstest, ersetzt einen ggf. bestehenden fieldStrip-Test — prüfen ob ein
   bestehender `TestFieldStrip*`-Test jetzt gelöscht statt anpasst werden muss, da die
   Funktion selbst entfällt). update_test.go: `TestKeyDetailFocusArrowNavigatesRelationsRows`
   (Regressionstest: bestätigt dass die BESTEHENDE Pfeil-Logik bereits korrekt auf die
   neue Darstellung wirkt, kein Verhaltens-, nur Render-Fix).
2. `command go test ./internal/tui/... -run "Relations|HangingIndent|FieldStrip"` → FAIL
   (bzw. Compile-Fehler nach Löschen von `fieldStrip`, falls Implementierung vor Tests
   passiert — TDD-Reihenfolge: Tests zuerst schreiben GEGEN die neue Signatur, dann
   implementieren, dann grün).
3. Implementieren (Reihenfolge: `hangingIndentWrap` zuerst [isoliert testbar], dann
   `relationsSectionBody`-Umbau, dann `beanSections`-Signaturkette, dann `fieldStrip`-
   Removal in `renderAccordion`).
4. Tests grün.
5. Golden-Regen: RELATIONS-Sektion ist Teil des Detail-Pane-Renderpfads — prüfen ob
   TestTreeGolden/TestBacklogGolden/TestChromeGolden ein Bean mit sichtbaren Relations
   (Parent/Children/Blocking/Blocked-By) im aufgeklappten Zustand zeigen. Falls ja:
   `command go test ./internal/tui/ -run "TestTreeGolden|TestBacklogGolden|TestChromeGolden" -update`,
   Vorher/Nachher-Diff PFLICHT im Commit-Body (jede geänderte Datei einzeln beschrieben,
   auch "unverändert" ist eine gültige, explizit zu nennende Aussage — Muster E8-plan.md).
6. `command go test ./... -short` grün (2x), voller Lauf grün, `-race` grün, gofmt/vet leer.
7. Commit `feat(tui): RELATIONS-Sektion — Fields-Dopplung raus, Pfeil-selektierbar,
   hängender Einzug (B04)`. Footer `Refs: bt-tct9`.

## Akzeptanz-Checkliste

- [x] Keine separate "Fields:"-Zeile mehr in RELATIONS
- [x] Parent/Children/Blocking/Blocked-By-Einträge tragen eigene `▷`/`▶`-Marker
- [x] Pfeiltasten (bereits bestehende Logik) selektieren sichtbar die echten Zeilen, nicht
      mehr eine unsichtbare Strip
- [x] Lange Titel brechen mit hängendem Einzug um — Meta-Spalten (Glyph/Status/ID) bleiben
      auf der ersten Zeile unangetastet, Folgezeilen richten sich unter dem Titel-Start aus
- [x] `fieldStrip` vollständig entfernt (kein toter Code, Compiler-verifiziert)
- [x] `enter` auf einer selektierten Relations-Zeile springt weiterhin korrekt
      (`activateDetailField`s Default-Case unverändert)
- [x] Goldens verifiziert (regeneriert ODER Gegenbeleg grün, im Commit-Body dokumentiert)
- [x] Voller Testlauf grün, gofmt/vet leer



## Summary

`accordion.go`s `fieldStrip` ist komplett gelöscht (Compiler-verifiziert: `renderAccordion`s
`if activeSec && i != 0 && len(s.fields) > 0 { fieldStrip(...) }`-Zweig entfernt, kein Aufrufer
bleibt übrig). `relationsSectionBody` (`view_detail_bean.go`) wächst um `(active bool, fieldIdx
int)`, mirrort `metaSectionBody`s Signatur exakt (PF-4). Neuer Helfer `relationRowMarker(active,
fieldIdx, rowIdx)` liefert pro Zeile `▶`(Accent)/`▷`(Muted) — `rowIdx` ist EIN über Parent(0)/
Children(1..n)/Blocking(n+1..m)/Blocked-By(m+1..k) laufender Zähler, threaded durch
`resolveSorted`/`beanListRow` (beide wachsen um `startIdx, active, fieldIdx, w` → `nextIdx`).
`relationRow` wächst um `(marker string, w int)` und wrappt via dem neuen
`hangingIndentWrap(prefix, text string, w int) string` (Ablage: `view_detail_bean.go`, neben
`relationRow`) statt der alten bare-Konkatenation — `relationsSectionBody`s Schluss-`wrapText`
über den gesamten verketteten Block entfällt ersatzlos (das WAR der PO-Mockup-Bug: eine Zeile
kannte beim Reflow nicht mehr den eigenen Präfix und fiel auf Spalte 0 zurück).

Zwei Bestands-Aufrufer von `relationRow` (Parent-/Blocking-Picker, `box_picker_parent.go`/
`box_picker_blocking.go`) rufen jetzt `relationRow(cand, "", relationRowNoWrap)` — neue Konstante
`relationRowNoWrap = 1 << 20` hält ihr Rendering byte-identisch zum Vor-B04-Stand (T5/bt-4mo9,
`blocked_by` dieser Task, schaltet sie später auf echtes Wrapping um).

`mouse.go`s `detailClickRow` verlor den GLEICHEN fieldStrip-Zeilen-Skip (`activeSec && i != 0 &&
len(s.fields) > 0 { row++ }`) — sonst hätte jede Sektion NACH einer aktiven+offenen RELATIONS-
Sektion um eine Zeile verschoben aufgelöst (eigener Test fängt die Regression, siehe Test-Output).
Diese Kopplung stand NICHT explizit in der Bean-Spezifikation, ist aber durch `mouse.go`s eigenen
Doc-Kommentar ("EXACT SAME ... fieldStrip-shown conditionals renderAccordion uses") belegt und
wurde als notwendige Konsequenz der Compiler-gesteuerten `fieldStrip`-Entfernung mitgezogen.

`relationsSectionBody`s `active`-Gate ist bewusst `focused && activeIdx == relationsSectionIdx`
OHNE zusätzliches `detailLevel == 1`-Gate (anders als `metaSectionBody`s eigenes) — das alte
`fieldStrip` zeigte seinen aktiven Feld-Marker bereits, sobald die SEKTION aktiv war, unabhängig
vom `detailLevel`. Diese Timing wurde 1:1 beibehalten (Pfeiltasten-Navigation ändert keinen Code,
nur die Visualisierung wechselt).

## Test-Output

RED 1 (hangingIndentWrap, vor Implementierung, `command go test ./internal/tui/ -run
"HangingIndent" -v`):

```
# beans-tui/internal/tui [beans-tui/internal/tui.test]
internal/tui/view_detail_bean_test.go:592:9: undefined: hangingIndentWrap
internal/tui/view_detail_bean_test.go:606:9: undefined: hangingIndentWrap
internal/tui/view_detail_bean_test.go:633:14: undefined: hangingIndentWrap
internal/tui/view_detail_bean_test.go:634:13: undefined: hangingIndentWrap
internal/tui/view_detail_bean_test.go:662:9: undefined: hangingIndentWrap
FAIL	beans-tui/internal/tui [build failed]
```

RED 2 (fieldStrip-Löschung, Compiler-gesteuert, `command go vet ./...` nach dem Löschen der
Funktion, accordion_test.go noch nicht angepasst):

```
# beans-tui/internal/tui
# [beans-tui/internal/tui]
vet: internal/tui/accordion_test.go:165:26: undefined: fieldStrip
```

RED 3 (mouse.go-Kopplung, `command go test ./internal/tui/ -run
"TestDetailClickRowNoOffByOneWhenRelationsSectionActiveWithFields" -v`, VOR dem mouse.go-Fix):

```
=== RUN   TestDetailClickRowNoOffByOneWhenRelationsSectionActiveWithFields
    mouse_test.go:480: secIdx = 2, want 3 (historySectionIdx)
--- FAIL: TestDetailClickRowNoOffByOneWhenRelationsSectionActiveWithFields (0.00s)
FAIL
```

GREEN (alle drei Stränge, `command go test ./internal/tui/ -run
"HangingIndent|RelationsSectionBody|FieldStrip|RenderAccordion|TestDetailClickRowNoOffByOne|
TestKeyDetailFocusArrowNavigatesRelationsRows" -v`): alle 16 Subtests PASS, u.a.
`TestHangingIndentWrapShortTitleNoWrap`, `TestHangingIndentWrapLongTitleAlignsContinuationUnder
PrefixEnd`, `TestHangingIndentWrapVaryingPrefixWidths`,
`TestHangingIndentWrapNeverCollapsesBelowMinimumContinuationWidth`,
`TestRelationsSectionBodyShowsCursorMarkerOnActiveRow`,
`TestRelationsSectionBodyNoLongerRendersFieldsStripLine`,
`TestRelationsSectionBodyLongTitleAlignsContinuationUnderTitleStartNotColumnZero`,
`TestRenderAccordionOmitsFieldStripForRelations`,
`TestDetailClickRowNoOffByOneWhenRelationsSectionActiveWithFields`,
`TestKeyDetailFocusArrowNavigatesRelationsRows`.

Golden-Gegenbeleg (`command go test ./internal/tui/ -run "TestTreeGolden|TestBacklogGolden|
TestChromeGolden" -update` gefolgt von `git diff --stat internal/tui/testdata/`): KEINE Datei
geändert (leerer Diff) — RELATIONS ist in allen 3 Goldens geschlossen (nur Header sichtbar, kein
aktiver Feld-Cursor), daher betrifft weder die fieldStrip-Entfernung noch relationsSectionBodys
Zeilen-Umbau den sichtbaren Golden-Output. Danach 2x ohne `-update`: byte-stabil (`ok`/cached).

Commit-Gate: `command go test ./...` → alle Packages `ok` (`internal/tui` 138.777s); `command go
test ./internal/tui/ -race` → `ok` (141.112s); `gofmt -l .` → leer; `command go vet ./...` → leer;
`command go test ./... -short` 2x → `ok`/cached.

## Smoke

tmux (100×30, dann zusätzlich 70×40 zur Umbruch-Stresstest), `bin/bt` in diesem Repo, Bean
`bt-tct9` (Parent=`bt-apmy`, 11 Children, mehrere mit langen Titeln, u.a. `bt-b0w0` selbst).
`Tab` → `3` (RELATIONS öffnet) → sichtbar: KEINE `Fields:`-Zeile mehr; `▶`-Marker liegt sofort auf
der Parent-Zeile (`bt-apmy`, kein `detailLevel`-Gate nötig, siehe Summary); der lange Parent-Titel
"beans-tui v1 — devd-TUI-Port auf beans" bricht um, Folgezeile eingerückt unter dem Titel-Start
(NICHT Spalte 0), Meta-Spalten (`t M bt-apmy`) auf Zeile 1 unangetastet. Children-Gruppe zeigt 4
Zeilen mit eigenem `▷`-Marker, mehrere mit mehrzeilig umbrechenden Titeln (`bt-b0w0`s eigener
Titel dreizeilig), alle konsistent hängend eingerückt. `l` (rechts, Feld-Ebene) + `k` (runter,
dieses Repo hat `k`=Down/`i`=Up) bewegt `▶` sichtbar von Parent → erstem Child (`bt-b0w0`) → zweitem
Child (`bt-gdkx`) — Live-Beleg der Pfeiltasten-Kaskade auf den echten Zeilen. `enter` auf der
selektierten `bt-gdkx`-Zeile: Tree-Cursor UND Detail-Pane springen zu `bt-gdkx`, `detailFocus`
verlässt (unverändertes E2-Verhalten). Bei 70 Spalten (bt-b0w0 fokussiert, RELATIONS offen):
`bt-b0w0`s eigener Parent (`bt-tct9`, langer Titel) bricht dreizeilig, jede Folgezeile weiterhin
korrekt eingerückt (kein Spalte-0-Fallback trotz engem Terminal). `grep -c "Fields:"` über das
gesamte Pane-Capture: 0.

## Deviations/ERRATA

Eine implizite, aber notwendige Ergänzung über die Bean-Spezifikation hinaus: `mouse.go`s
`detailClickRow` musste denselben fieldStrip-Zeilen-Skip verlieren wie `accordion.go`s
`renderAccordion` (siehe Summary) — sonst wäre jede Sektion nach einer aktiven+offenen RELATIONS-
Sektion beim Maus-Klick um eine Zeile verschoben aufgelöst worden. Kein Akzeptanzkriterium der
Bean nennt `mouse.go` explizit, aber die Kopplung war durch `mouse.go`s eigenen Doc-Kommentar
("EXACT SAME ... fieldStrip-shown conditionals") bereits dokumentiert und wurde durch einen echten
RED/GREEN-Regressionstest (`TestDetailClickRowNoOffByOneWhenRelationsSectionActiveWithFields`)
belegt, nicht nur behauptet. Ansonsten keine Abweichung von Architektur-Vorgabe/TDD-Schritten.

## Notes for T(n+1)

- **hangingIndentWrap** (`internal/tui/view_detail_bean.go`, neben `relationRow`): Signatur
  `func hangingIndentWrap(prefix, text string, w int) string` — `prefix` ist bereits VOLL
  gestylt (ANSI), seine `lipgloss.Width(ansi.Strip(prefix))`-Breite wird der Einzug jeder
  Folgezeile; `text` bleibt unstyled (roher Titel). Kontinuierungsbreite floort bei 8
  (`if contW < 8 { contW = 8 }`), Einzug selbst bleibt IMMER der volle (ungeclampte)
  Präfix-Breite. T5/bt-4mo9 kann ihn direkt importieren/aufrufen — keine Anpassung nötig, exportiert
  bereits paket-intern (`tui`-Package, kein neues Package).
- **relationRow** wuchs um `(marker string, w int)` — für T5s Picker-Zeilen bitte NICHT
  `relationRowNoWrap` (`1 << 20`, extra für die Alt-Kompatibilität der beiden Picker gedacht)
  weiterverwenden, sondern die ECHTE Picker-Breite (`wideModalWidth(...)` laut T5-Spec) durchreichen
  — sonst bleibt der B06-Umbruch-Bug bestehen, den T5 eigentlich beheben soll.
- **relationRowMarker(active bool, fieldIdx, rowIdx int) string** (`view_detail_bean.go`): kleiner
  Helfer, falls T5s Picker eine ähnliche `▷`/`▶`-Selektion braucht — bislang nur von
  `relationsSectionBody`/`resolveSorted`/`beanListRow` aufgerufen, aber generisch genug für Wiederverwendung.
- `relationField.label` ist für RESOLVED Relations-Zeilen (Parent/Children/Blocking/Blocked-By,
  NICHT die unresolved-Variante) seit dieser Task praktisch totes Datenfeld (nur noch
  `TestBeanSectionsRelationsDanglingReferenceShowsUnresolvedNotJumpable` prüft es, und nur für den
  unresolved-Zweig) — bewusst NICHT entfernt (out of scope, würde `resolveSorted`/`beanListRow`s
  Rückgabe-Contract unnötig aufbrechen), aber ein Kandidat für eine spätere Aufräum-Runde (I-Code,
  kein Bug).
