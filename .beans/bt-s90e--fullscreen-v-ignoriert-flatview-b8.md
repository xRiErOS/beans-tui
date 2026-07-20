---
# bt-s90e
title: Fullscreen (v) ignoriert flatView (B8)
status: completed
type: bug
priority: low
created_at: 2026-07-20T07:26:50Z
updated_at: 2026-07-20T09:12:28Z
parent: bt-vy1q
---

**B8, gefunden in S5 (2026-07-20).** Der Nested/Flat-Toggle `G` wirkt nur in der Browse-Split-Ansicht. Der Fullscreen-Modus (`v`, `view_fullscreen.go`) rendert immer den Tree, auch wenn `m.flatView` gesetzt ist — inkonsistent.

## Akzeptanz
- [ ] Fullscreen respektiert `m.flatView` (Flat-Liste statt Tree, wenn aktiv)
- [ ] `G` funktioniert auch im Fullscreen (oder ist dort bewusst deaktiviert — dann dokumentieren)
- [ ] Bestandsgolden unveraendert bzw. bewusst regeneriert
- [ ] Voller `command go test ./...` gruen


## Prelude aus bt-ze10 (2026-07-20) — zuerst erledigen

Der Detail-Scroll (Commit `8e5a869`) hat den **Fullscreen-Pfad bewusst ausgespart**:
`fullscreenDetail` im Box-Modus scrollt NICHT (Single-Pane-Geometrie weicht von der
Split-Pane ab; `view_fullscreen.go` stand nicht in der Datei-Liste jenes beans).
`renderFullscreenBody` uebergibt literal `0` als Scroll-Offset; Kommentare sitzen am
`keyDetailFocus`-Guard und an der Aufrufstelle.

Damit sammeln sich **zwei** Fullscreen-Luecken an derselben Stelle:
1. `v` ignoriert `flatView` (B8, dieses bean)
2. `v` ignoriert im Box-Modus den Scroll-Offset (aus ze10)

Beide in `view_fullscreen.go`. Zusammen erledigen — getrennt waere doppelte Einarbeitung
in dieselbe Geometrie.


## Summary

Beide Fullscreen-Luecken geschlossen, je ein Commit.

**1. Scroll-Luecke (Prelude aus bt-ze10) — 40e2a77**
Die Vermutung im Prelude ("Single-Pane-Geometrie weicht ab") traf zu, aber
nur zur Haelfte: die Abweichung ist **width-only**. `bodyH` leitet
boxFormScrollBounds ohnehin aus dem Chrome des AKTUELLEN Modells ab
(browseRepoChrome ist footer-context-aware), liefert im Vollbild also von
selbst den richtigen Wert. Nur `accW` war falsch: Split nutzt `rw-2`,
Vollbild braucht `innerW-4` (paneW=innerW-2, davon renderAccordionPane's
eigenes w-2). Damit:
- mouse.go: `accW`-Branch auf `m.fullscreen == fullscreenDetail`
- update.go: keyDetailFocus-Guard `&& m.fullscreen != fullscreenDetail` entfernt
- view_fullscreen.go: `boxScroll`-Parameter statt literal `0`
- beide Aufrufstellen reichen `boxFormEffectiveScroll(m, detailBean)` durch

`adjustBoxFormScroll` bleibt der EINE Mutationspunkt fuer beide Geometrien —
kein zweiter Scroll-Pfad, der driften koennte. Maus-Rad blieb bewusst
unberuehrt: handleMouse no-opt im Vollbild komplett (dokumentierter
Scope-Cut F01/bt-13l7).

**2. B8 flatView — 3d8e00f**
Der fullscreenList-Zweig in viewBrowseRepo baute seine Rows unbedingt aus
`treeRows()`. Der Dispatch konnte flat schon laenger (keyTree -> keyFlat,
focusedBean -> flatSelected) — nur der Render fehlte. Jetzt derselbe
`m.flatView`-Switch wie im Split-Pane. **Nicht** auf boxFormEnabled()
gegated, da `G` unabhaengig vom Flag existiert.

## Test-Output

Voller Lauf, uncached (`command go test ./... -count=1`):

```
?   	github.com/xRiErOS/beans-tui	[no test files]
ok  	github.com/xRiErOS/beans-tui/cmd	0.385s
?   	github.com/xRiErOS/beans-tui/internal/clip	[no test files]
ok  	github.com/xRiErOS/beans-tui/internal/config	0.717s
ok  	github.com/xRiErOS/beans-tui/internal/data	5.031s
ok  	github.com/xRiErOS/beans-tui/internal/theme	1.050s
ok  	github.com/xRiErOS/beans-tui/internal/tui	151.410s
```

Neue Tests: box_form_scroll_fullscreen_test.go (5), view_fullscreen_flat_test.go (4).
**Keine Golden-Datei angefasst** (`git diff --name-only 7dc4aa0..HEAD -- '*.golden'` leer) —
weder geteilte noch neue.

## Deviations

- **Worktree-Basis war falsch** (von `main` erzeugt, ~40 Commits hinter
  `experiment/jira-style-ui`). Vor der ersten Code-Aenderung per
  `git reset --hard experiment/jira-style-ui` korrigiert; keine eigenen
  Commits gingen verloren.
- **Keine neuen Fullscreen-Golden angelegt.** Die Akzeptanz erlaubte sie,
  aber beide Luecken sind ueber echte `tea.KeyMsg`-Roundtrips + `m.View()`-
  Assertions strukturell abgedeckt (z.B. "collapsed leaf tk-2 sichtbar" ist
  im Tree-Pfad gar nicht erzeugbar). Ein zusaetzliches Golden haette nur
  Kollisionsflaeche mit den zwei parallel laufenden Agenten geschaffen.
- **Testannahme korrigiert:** Erst angenommen, `bodyH` sei in Split und
  Vollbild identisch — ist es nicht (Footer-Kontext unterscheidet sich, 18
  vs. 19). Die Assertion wurde durch die Erkenntnis ersetzt, dass die Hoehe
  gar keinen Branch braucht.
- Kein tmux-Smoke: keine Footer-/Wrap-Aenderung (die Regel greift nicht).
- Keymap unveraendert — keine neue Bindung, `TestHelpGroupsCoverEveryBindingExactlyOnce` unberuehrt.
