---
# bt-b0w0
title: 'RELATIONS-Sektion: Dopplung raus, Pfeil-selektierbar, hängender Einzug (B04)'
status: in-progress
type: task
priority: normal
created_at: 2026-07-16T06:45:47Z
updated_at: 2026-07-16T10:34:28Z
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

- [ ] Keine separate "Fields:"-Zeile mehr in RELATIONS
- [ ] Parent/Children/Blocking/Blocked-By-Einträge tragen eigene `▷`/`▶`-Marker
- [ ] Pfeiltasten (bereits bestehende Logik) selektieren sichtbar die echten Zeilen, nicht
      mehr eine unsichtbare Strip
- [ ] Lange Titel brechen mit hängendem Einzug um — Meta-Spalten (Glyph/Status/ID) bleiben
      auf der ersten Zeile unangetastet, Folgezeilen richten sich unter dem Titel-Start aus
- [ ] `fieldStrip` vollständig entfernt (kein toter Code, Compiler-verifiziert)
- [ ] `enter` auf einer selektierten Relations-Zeile springt weiterhin korrekt
      (`activateDetailField`s Default-Case unverändert)
- [ ] Goldens verifiziert (regeneriert ODER Gegenbeleg grün, im Commit-Body dokumentiert)
- [ ] Voller Testlauf grün, gofmt/vet leer
