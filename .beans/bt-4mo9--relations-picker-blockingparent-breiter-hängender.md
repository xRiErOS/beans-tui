---
# bt-4mo9
title: Relations-Picker (Blocking/Parent) breiter + hängender Einzug (B06)
status: in-progress
type: task
priority: normal
created_at: 2026-07-16T06:45:53Z
updated_at: 2026-07-16T11:39:25Z
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

- [ ] `wideModalWidth` skaliert mit der Terminalbreite (≈85%, Boden 60, Deckel termW-4)
- [ ] Blocking-Picker UND Parent-Picker nutzen `wideModalWidth` statt `clampModalWidth(48,…)`
- [ ] Kein Eintrag bricht mehr mitten in der ID um (hängender Einzug, ID ist Teil des
      unwrappable Präfix)
- [ ] Höhe (`parentPickerRowBudget = 14`) unverändert
- [ ] tmux-Smoke bei 80 Spalten belegt (Commit-Body)
- [ ] Keine Tree/Backlog/Chrome-Goldens betroffen (Gegenbeleg grün)
- [ ] Voller Testlauf grün, gofmt/vet leer


## PRELUDE (2026-07-16, aus T4-Re-Review F05 — ZUERST erledigen, eigener Mini-Commit oder im Haupt-Commit, kein eigener Zyklus)

Doku-Nit in view_detail_bean.go:236 (Kommentar über dem Hardwrap-Aufruf in hangingIndentWrap):
"(true = preserve ANSI sequences)" ist falsch beschriftet — Hardwraps dritter Parameter ist
preserveSpace (führende Leerzeichen erhalten); ANSI-Sequenzen werden immer erhalten.
Korrektur: "(true = preserveSpace, wie wrapText)". Kein Verhaltens-Fix, reine Kommentar-Korrektur.
