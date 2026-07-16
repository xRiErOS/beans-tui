---
# bt-4mo9
title: Relations-Picker (Blocking/Parent) breiter + hängender Einzug (B06)
status: completed
type: task
priority: normal
created_at: 2026-07-16T06:45:53Z
updated_at: 2026-07-16T12:10:22Z
parent: bt-tct9
blocked_by:
    - bt-b0w0
---

E9 Task 5 — deckt B06 aus bean bt-tct9 (Live-Nachtrag 2026-07-16, Abschnitt "B06" im
Epic-Body). Quelle: design-spec.md §15 PF-17 (Abschnitt B06). Ist-Code:
internal/tui/box_picker_blocking.go (blockingPickerBox), internal/tui/box_picker_parent.go
(parentPickerBox, parentPickerRowBudget), internal/tui/box_filter_facets.go
(clampModalWidth — Vorbild/Nachbar für den neuen Breiten-Helfer). **blocked_by bt-tct9
Task 4 (B04):** dieser Task nutzt den `hangingIndentWrap`-Helfer, den Task 4 in
view_detail_bean.go einführt — Reihenfolge vermeidet einen zweiten, unabhängigen Wrap-
Helfer für praktisch dieselbe Zeilenform (Glyph+ID+Titel).

## B06 — Relations-Picker (Blocking `r`, Parent `a`) viel zu schmal

PO verbatim: "Das overlay für 'r' um die relations anzugeben ist viel zu schmal. So kann
man die beans nicht sauber lesen und bearbeiten. Höhe passt. Aber die Breite muss viel
weiter werden." Screenshot-Befund: Blocking-Picker (~48 Spalten) bricht Einträge MITTEN in
der ID um ("bt-" am Zeilenende, Rest nächste Zeile), Glyphen/IDs/Titel unlesbar
verschränkt. Interpretation (PO im bean, gilt analog für den Parent-Picker 'a' — geprüft:
identische `clampModalWidth(48, m.width)`-Aufrufe in beiden Dateien): Overlay-Breite
deutlich erhöhen (Richtung ~80-90% der Terminalbreite bzw. inhaltsbasiert), Einträge
einzeilig wo möglich; bei Umbruch hängender Einzug analog B04.3 (Meta-Spalten nie
unterwandern). **Höhe UNVERÄNDERT** (PO: "Höhe passt").

## Architektur-Vorgabe

**1. Neuer Breiten-Helfer** (box_filter_facets.go, direkt neben `clampModalWidth` — Single
Source für Modal-Breiten-Helfer, gleiche Datei):

```go
// wideModalWidth sizes a floating box relative to the terminal, unlike
// clampModalWidth (which only ever SHRINKS a fixed preference) -- B06,
// design-spec.md §15 PF-17: the Blocking-/Parent-Picker need to GROW on
// wide terminals (PO: "die Breite muss viel weiter werden"), not stay
// pinned to a fixed 48. ~85% of termW, floor 60 (never narrower than the
// old fixed value), ceiling termW-4 (same 2-column margin convention as
// clampModalWidth).
func wideModalWidth(termW int) int {
    w := termW * 85 / 100
    if w < 60 { w = 60 }
    if termW > 4 && w > termW-4 { w = termW - 4 }
    if w < 24 { w = 24 } // absolute floor, mirrors clampModalWidth's own floor
    return w
}
```

**2. Aufrufer wechseln.** `blockingPickerBox` (box_picker_blocking.go) und
`parentPickerBox` (box_picker_parent.go): `clampModalWidth(48, m.width)` →
`wideModalWidth(m.width)` (jeweils EINE Zeile, an der bestehenden `return modalPanel(...)`-
Stelle).

**3. Zeilen-Rendering: hängender Einzug (wiederverwendet aus Task 4/B04.3).**
`blockingPickerBox`s/`parentPickerBox`s bisherige Ein-Zeilen-Konkatenation
(`blockingDot(...) + it.label` bzw. `it.label`, `it.label` selbst ist bereits
`relationRow(...)`-vorformatiert) wird durch denselben `hangingIndentWrap`-Aufruf ersetzt,
den Task 4 für die RELATIONS-Sektion einführt — strukturell identische Zeilenform
(Cursor-Präfix `▌`/` ` bzw. `●`/`○` statt `▷`/`▶`, sonst dieselbe Glyph+ID+Titel-Form).
`prefix` = der D08-Cursor-Präfix (`"▌"` aktiv / `" "` inaktiv für Parent, `blockingDot(...)`
für Blocking) + Glyphen+ID-Teil von `relationRow`; `text` = der Titel-Teil. Verhindert
exakt den vom PO gezeigten Mitten-in-der-ID-Umbruch (die ID selbst ist Teil des `prefix`,
wird NIE vom `ansi.Wordwrap`-Aufruf berührt — nur `text`/der Titel wrapt).

**Höhe bleibt unverändert** — `parentPickerRowBudget = 14` (box_picker_parent.go, von
BEIDEN Pickern genutzt) wird NICHT angefasst. `windowAround(rows, parentPickerRowBudget,
m.menu.cursor)` bleibt unverändert — ACHTUNG: mit mehrzeiligen (umgebrochenen) Rows
verschiebt sich die Bedeutung von "14 Einträge" ggf. auf "14 windowAround-Slice-Elemente",
was bei Multi-Line-Rows NICHT mehr 14 sichtbare Zeilen bedeutet, sondern bis zu 14×N
Zeilen (N = Zeilen pro gewrapptem Eintrag) — falls das bei sehr vielen langen Titeln zu
einem sichtbar überlangen Modal führt, ist das ein Implementer-Judgment-Call (dokumentieren,
kein Blocker: die meisten Bean-Titel bleiben auch bei 85%-Breite einzeilig, das Problem
tritt nur bei ungewöhnlich vielen ungewöhnlich langen Titeln gleichzeitig im sichtbaren
Fenster auf).

## TDD-Schritte

1. Failing tests: `box_filter_facets_test.go` (oder neue Datei) `TestWideModalWidthScalesWithTerminal`
   (verschiedene termW-Werte → erwartete Breiten, inkl. Boden-60/Deckel-termW-4-Grenzfälle);
   `box_picker_blocking_test.go`/`box_picker_parent_test.go`:
   `TestBlockingPickerBoxUsesWideModalWidth`/`TestParentPickerBoxUsesWideModalWidth`
   (Breite im gerenderten Output ≥ alter fixer 48 bei einem breiten Terminal, z. B. 120
   Spalten → erwartete ~100 statt 48); `TestBlockingPickerBoxLongTitleWrapsWithHangingIndent`/
   `TestParentPickerBoxLongTitleWrapsWithHangingIndent` (langer Titel, ID bleibt auf Zeile 1
   vollständig intakt, Folgezeile beginnt mit `indentW` Leerzeichen, KEIN Umbruch mitten in
   der ID).
2. `command go test ./internal/tui/... -run "WideModalWidth|BlockingPickerBox|ParentPickerBox"` → FAIL.
3. Implementieren (Reihenfolge: `wideModalWidth` zuerst, dann beide Picker-Boxen).
4. Tests grün.
5. Golden-Regen: Picker-Overlays sind NICHT Teil von Tree/Backlog/Chrome-Goldens
   (Bestandskonvention, E8-plan.md: "Overlays sind nicht Teil von Tree/Backlog/Chrome-
   Goldens") — Gegenbeleg-Lauf `command go test ./internal/tui/ -run "TestTreeGolden|TestBacklogGolden|TestChromeGolden"`
   OHNE -update MUSS grün bleiben.
6. **Grenzbreiten-Smoke PFLICHT** (CLAUDE.md-Regel, jede Breiten-/Wrap-Änderung braucht
   einen tmux-Smoke bei 80 Spalten): Blocking-Picker UND Parent-Picker bei 80 Spalten real
   in tmux gegen `./bin/bt` prüfen (ein Bean mit langem Titel als Kandidat-Zeile
   sicherstellen), `tmux capture-pane -p` als Beleg im Commit-Body — verifizieren dass IDs
   nie mehr mitten im Wort umbrechen.
7. `command go test ./... -short` grün (2x), voller Lauf grün, `-race` grün, gofmt/vet leer.
8. Commit `feat(tui): Relations-Picker (Blocking/Parent) breiter + hängender Einzug (B06)`.
   Footer `Refs: bt-tct9`.

## Akzeptanz-Checkliste

- [x] `wideModalWidth` skaliert mit der Terminalbreite (≈85%, Boden 60, Deckel termW-4)
- [x] Blocking-Picker UND Parent-Picker nutzen `wideModalWidth` statt `clampModalWidth(48,…)`
- [x] Kein Eintrag bricht mehr mitten in der ID um (hängender Einzug, ID ist Teil des
      unwrappable Präfix)
- [x] Höhe (`parentPickerRowBudget = 14`) unverändert
- [x] tmux-Smoke bei 80 Spalten belegt (Commit-Body/Summary)
- [x] Keine Tree/Backlog/Chrome-Goldens betroffen (Gegenbeleg grün)
- [x] Voller Testlauf grün, gofmt/vet leer


## PRELUDE (2026-07-16, aus T4-Re-Review F05 — ZUERST erledigen, eigener Mini-Commit oder im Haupt-Commit, kein eigener Zyklus)

Doku-Nit in view_detail_bean.go:236 (Kommentar über dem Hardwrap-Aufruf in hangingIndentWrap):
"(true = preserve ANSI sequences)" ist falsch beschriftet — Hardwraps dritter Parameter ist
preserveSpace (führende Leerzeichen erhalten); ANSI-Sequenzen werden immer erhalten.
Korrektur: "(true = preserveSpace, wie wrapText)". Kein Verhaltens-Fix, reine Kommentar-Korrektur.


## Summary

wideModalWidth(termW) hinzugefuegt (box_filter_facets.go, neben clampModalWidth): ~85% termW, Boden 60, Deckel termW-4, absoluter Boden 24. blockingPickerBox/parentPickerBox wechseln von clampModalWidth(48, m.width) auf wideModalWidth(m.width). relationRowPrefix(rel) aus relationRow extrahiert (view_detail_bean.go) -- Blocking-/Parent-Picker komponieren jetzt prefix (Cursor-Bar/Dot + Glyph+ID) und text (Titel) getrennt und reichen sie an hangingIndentWrap mit der ECHTEN, live berechneten Picker-Breite durch, statt der alten relationRowNoWrap-Sentinel (entfernt, keine Aufrufer mehr). pickerItem trägt jetzt prefix/title statt eines vorgerenderten label. D08-Cursor-Behandlung (ansi.Strip + Accent-Recolor) bleibt erhalten, jetzt auf den prefix (inkl. Dot/Cursor-Bar) UND separat auf title beschränkt statt auf die alte Ein-Zeilen-Konkatenation -- Pending-Dot-SHAPE (●/○) bleibt auch unter Accent-Override lesbar. Hoehe (parentPickerRowBudget=14) unveraendert.

## Test-Output

RED (Compile-Fail, wideModalWidth undefined):
```
internal/tui/box_filter_facets_test.go:296:14: undefined: wideModalWidth
internal/tui/box_filter_facets_test.go:311:10: undefined: wideModalWidth
FAIL	beans-tui/internal/tui [build failed]
```

GREEN (Ziel-Tests):
```
--- PASS: TestWideModalWidthScalesWithTerminal (0.00s)
--- PASS: TestWideModalWidthNeverNarrowerThanOldFixedPicker (0.00s)
--- PASS: TestBlockingPickerBoxUsesWideModalWidth (0.00s)
--- PASS: TestBlockingPickerBoxLongTitleWrapsWithHangingIndent (0.00s)
--- PASS: TestParentPickerBoxUsesWideModalWidth (0.00s)
--- PASS: TestParentPickerBoxLongTitleWrapsWithHangingIndent (0.00s)
```

Voller Lauf (2x -short, 1x voll, -race, gofmt, vet) -- alle gruen:
```
go test ./... -short -count=1  -> ok (alle Pakete)
go test ./... -count=1         -> ok (alle Pakete, internal/tui 139.6s)
go test ./internal/tui/ -race -count=1 -> ok (142.5s)
gofmt -l .  -> leer
go vet ./... -> leer
```

## Golden-Gegenbeleg

```
go test ./internal/tui/ -run "TestTreeGolden|TestBacklogGolden|TestChromeGolden" -v
--- PASS: TestChromeGolden
--- PASS: TestTreeGolden
--- PASS: TestTreeGoldenDeterministic
--- PASS: TestBacklogGolden
--- PASS: TestBacklogGoldenDeterministic
PASS
```
Kein -update noetig, alle drei Basis-Goldens unveraendert gruen (Picker-Overlays sind nicht Teil dieser Goldens).

## Smoke

tmux 80 Spalten, Blocking-Picker ('r' auf bt-apmy): Overlay deutlich breiter als der alte 48er-Fixwert (spannt fast die volle 80er-Breite), alle 14 IDs (bt-blsy, bt-aq5s, bt-gzcu, bt-tfqi, bt-zk9p, bt-ntoz, bt-tct9, bt-5h4d, bt-4mo9, bt-heg9, bt-de1v [Test-Kandidat, langer Titel], bt-gdkx, bt-6oyy, bt-1e0t) vollstaendig intakt, kein Umbruch mitten in der ID. bt-de1v (Testbean, 4-zeiliger Titel) wrapt hängend eingerückt exakt unter dem Titelbeginn. Kein Layout-Bruch bei 80 Spalten, Hoehe passt ins 30-Zeilen-Fenster.

tmux 80 Spalten, Parent-Picker ('a' auf bt-6oyy, Feature): identisches Bild -- "(No parent)"-Zeile + alle eligiblen Milestones/Epics mit intakter ID, bt-de1v hängend eingerückt ueber 4 Zeilen.

tmux 160 Spalten (beide Picker): wideModalWidth(160)=136 -- die meisten Titel jetzt einzeilig (Ziel "Einträge einzeilig wo möglich" erfuellt), nur bt-de1v's extrem langer Testtitel wrapt noch (2 Zeilen), weiterhin sauber hängend eingerueckt. Deutlich sichtbare 85%-Skalierung ggue. 80 Spalten.

Testbean bt-de1v (temporaerer Kandidat mit langem Titel fuer den Smoke) nach dem Smoke wieder geloescht (beans delete bt-de1v --force) -- .beans/ danach clean, kein Rueckstand.

## Deviations/ERRATA

- **Cursor-Farb-Konvention geaendert (dokumentierter Judgment Call):** Die alte D08-Behandlung faerbte die GESAMTE Zeile (Dot+Glyphen+ID+Titel) als EINEN String-Block Accent ein (ansi.Strip + ein Accent.Render-Aufruf ueber alles). Da hangingIndentWrap prefix (gestylt) und text (roh) getrennt entgegennimmt, faerbe ich jetzt prefix und title SEPARAT Accent (zwei Render-Aufrufe statt einem) -- visuell identisch (durchgehend Accent-farbig inkl. Pending-Dot, dessen ●/○-FORM auch unter Accent-Override lesbar bleibt), aber nicht mehr byte-identisch (zwei SGR-Paare statt einem an der Naht). Da Picker-Overlays nicht golden-getestet sind (E8-Konvention), keine Golden-Auswirkung. Non-Cursor-Prefix wechselt von festen 2 Leerzeichen auf 1 Leerzeichen (Breitenparitaet mit dem 1-Zeichen-Cursor-Balken "▌", noetig damit hangingIndentWrap den Einzug korrekt und konsistent zwischen Cursor-/Nicht-Cursor-Zustand berechnet -- sonst haette ein Zeilenumbruch beim Cursor-Wechsel den Einzug um 1 Spalte verschoben).
- **contW-Herleitung empirisch verifiziert:** modalBox nutzt lipgloss `.Width(w).Border(...).Padding(0,1)` -- per go-run-Experiment verifiziert, dass Padding INNERHALB des deklarierten Width absorbiert wird (nicht davor addiert), Border dagegen AUSSERHALB addiert (+2 Gesamt). contW = w-2 (nur Padding-Abzug) ist damit die korrekte Textbreite fuer hangingIndentWrap, NICHT das an anderer Stelle etablierte "w-4"-Muster (das ist accordion-spezifisch: Cursor-Praefix-Reservierung + eigenes PaddingLeft, eine andere Geometrie).
- **relationRowNoWrap entfernt** (dead code nach dem Refactor, keine Aufrufer mehr) statt nur ungenutzt liegen gelassen -- Compiler-gesteuerte Verifikation, Muster PF-14/B13-Removal.

## Notes for T9 (Abschluss)

- design-spec.md §15 PF-17 Abschnitt B06 bleibt inhaltlich konsistent mit dem Code-Stand -- hangingIndentWrap jetzt geteilt zwischen T4 (RELATIONS-Sektion) und T5 (Blocking-/Parent-Picker), wie im Plan vorgesehen.
- Kein neuer Wrap-Helfer eingefuehrt (relationRowPrefix ist eine reine Extraktion, keine neue Wrap-Logik).
