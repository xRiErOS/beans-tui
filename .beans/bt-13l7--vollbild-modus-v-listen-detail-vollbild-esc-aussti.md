---
# bt-13l7
title: Vollbild-Modus 'v' — Listen-/Detail-Vollbild, esc-Ausstieg (F01 Kernmechanik)
status: completed
type: task
priority: normal
created_at: 2026-07-16T06:45:55Z
updated_at: 2026-07-16T13:31:51Z
parent: bt-tct9
blocked_by:
    - bt-b0w0
---

E9 Task 7 — deckt F01 Kernmechanik (Zustandsmodell, Einstieg `v`, Listen→Detail-Vollbild
via `enter`, Ausstieg `esc`, Rendering, Maus-Guard) aus bean bt-tct9 — OHNE den History-
Stack (`ctrl+links`/`ctrl+rechts`, separater Folge-Task bt-tct9 Task 8, s. Notes for
<Task-8-bean-id> am Ende dieses Bodys). Quelle: design-spec.md §15 „F01 — Vollbild-
Navigation" (vollständiges Zustandsmodell dort, dieser bean-Body ist die verkürzte
Umsetzungs-Anleitung + Code-Sketches). Ist-Code: internal/tui/types.go (model-Felder),
internal/tui/update.go (handleKey, focusedBean, keyDetailFocus, activateDetailField),
internal/tui/view_browse_repo.go (viewBrowseRepo, browseRepoChrome), internal/tui/
view_browse_backlog.go (viewBacklog, backlogChrome), internal/tui/mouse.go (handleMouse),
internal/tui/keymap.go (neue keys.Fullscreen-Bindung), NEUE Datei internal/tui/
view_fullscreen.go. **blocked_by bt-tct9 Task 4 (B04):** braucht B04.2 (Relations-Einträge
per Pfeil selektierbar) — ohne B04.2 gäbe es im Vollbild-Detail keinen Weg, überhaupt einen
Relations-Eintrag für den Sprung zu selektieren.

## F01 (Kernmechanik) — 'v' Vollbild-Modus + Navigations-Pfad (ohne History-Stack)

PO-Idee, verbatim in bt-tct9: "'v' (view) öffnet das aktuell fokussierte Pane im Vollbild:
Browse+links-fokussiert → Beans-Liste Vollbild; Browse+rechts-fokussiert → Detail-View
Vollbild. In Listen-Vollbild: enter auf Bean → Detail-View-Vollbild. esc kehrt zum Browse
zurück (aus Detail-Vollbild via Relations-Sprung: zurück zum Browse mit dem AKTUELLEN Bean
selektiert). Im Detail-View: enter auf Relations-Eintrag (z.B. Children) öffnet das
Ziel-Bean im Detail-View (Sprung)."

## Zustandsmodell

Neuer, zu `m.view` ORTHOGONALER Modus (types.go):

```go
type fullscreenMode int
const (
    fullscreenNone fullscreenMode = iota
    fullscreenList                // Tree ODER Backlog-Liste (je nach m.view) — Vollbreite
    fullscreenDetail               // Detail-Accordion EINES Beans — Vollbreite
)
```

Neue Modelfelder: `fullscreen fullscreenMode` · `fullscreenBeanID string` (das in
`fullscreenDetail` gezeigte Bean — UNABHÄNGIG vom Tree-/Backlog-Cursor). `navBack`/
`navForward []string` werden HIER bereits als leere Felder deklariert (Typ `[]string`),
aber in DIESEM Task noch nicht gefüllt/gelesen (Task 8 baut die History-Logik darauf auf —
Deklaration hier vermeidet einen unnötigen Struct-Diff in Task 8).

## Einstieg (`v`)

Neue Bindung `keys.Fullscreen` (keymap.go: `keybind.NewBinding(keybind.WithKeys("v"),
keybind.WithHelp("v", "fullscreen"))` — verifiziert frei, keine Kollision). Neue Datei
`view_fullscreen.go`: `keyFullscreen(msg tea.KeyMsg) (bool, tea.Model, tea.Cmd)` (Signatur
mirrort `keyNodeAction`s handled-Flag-Muster). Dispatch-Punkt: `handleKey` (update.go),
NACH dem `FocusIn`/`FocusOut`-Block, VOR `keyNodeAction` — greift nur wenn `m.view !=
viewLobby` (alle Full-Capture-States sind an dieser Stelle bereits abgefangen, Bestandsmuster
unverändert).

```go
func (m model) keyFullscreen(msg tea.KeyMsg) (bool, tea.Model, tea.Cmd) {
    if !keybind.Matches(msg, keys.Fullscreen) || m.view == viewLobby {
        return false, m, nil
    }
    if m.fullscreen != fullscreenNone {
        return true, m, nil // v ist Einweg-Einstieg, kein Toggle -- esc ist der Ausstieg
    }
    if !m.detailFocus {
        m.fullscreen = fullscreenList
        return true, m, nil
    }
    b := m.focusedBean()
    if b == nil {
        return true, m, nil
    }
    m.fullscreen = fullscreenDetail
    m.fullscreenBeanID = b.ID
    return true, m, nil
}
```

Gilt SYMMETRISCH für Browse (Tree) UND Backlog — `m.view` bleibt beim Eintritt unverändert
(die Renderfunktion, s. u., fragt nur "welche Listen-Rows/welches Bean", nicht "welche
View").

## Listen-Vollbild → Detail-Vollbild (`enter`)

Innerhalb `keyFullscreen`, VOR dem `keys.Fullscreen`-Check (oder als zweiter Case in
derselben Funktion): wenn `m.fullscreen == fullscreenList` UND `keybind.Matches(msg,
keys.Enter)`: Bean unter dem aktuellen Tree-/Backlog-Cursor auflösen (`m.focusedBean()`
funktioniert hier bereits UNVERÄNDERT, da `m.view` nicht wechselt); bei nil (Blatt-loser
Cursor/orphan-Root) → handled No-Op; sonst `m.fullscreen = fullscreenDetail;
m.fullscreenBeanID = b.ID` + Detail-Fokus-Maschine reinitialisiert (`m.secCursor,
m.accOpen, m.detailLevel, m.fieldCursor = 0, 1, 0, 0` — identisch zu `FocusIn`s
bestehendem Reset-Muster). Auf/Ab/Rechts/Links im Listen-Vollbild bleiben SONST identisch
zum Split-Modus — `keyFullscreen` reicht bei `m.fullscreen == fullscreenList` und keinem
der oben behandelten Keys unbehandelt durch (`return false, m, nil`), sodass die
BESTEHENDEN `keyTree`/`keyBacklog`-Handler unverändert weiterlaufen (kein Duplikat der
Cursor-Bewegungs-Logik).

## Detail-Vollbild: Feld-Navigation + Relations-Sprung (setzt B04.2 voraus)

`m.focusedBean()` (update.go) bekommt einen NEUEN, VOR dem bestehenden `m.view`-Switch
geprüften Fall:

```go
func (m model) focusedBean() *data.Bean {
    if m.fullscreen == fullscreenDetail {
        if m.idx == nil { return nil }
        b, ok := m.idx.ByID[m.fullscreenBeanID]
        if !ok { return nil }
        return b
    }
    switch m.view {
    // ... unverändert
    }
}
```

Damit funktioniert `keyDetailFocus` (update.go) VERBATIM im Vollbild — `handleKey`s
bestehender `if m.detailFocus { return m.keyDetailFocus(msg) }`-Zweig muss zusätzlich bei
`m.fullscreen == fullscreenDetail` greifen (NICHT nur bei `m.detailFocus` — im Vollbild-
Detail ist `m.detailFocus`s Wahrheitswert für die Dispatch-Entscheidung irrelevant, der
Vollbild-Zustand selbst IST der Signal-Geber): `if m.detailFocus || m.fullscreen ==
fullscreenDetail { return m.keyDetailFocus(msg) }`. JEDE bestehende Feld-Kaskade (PF-5,
status/type/priority/tags-Overlays, Titel-Form) funktioniert dadurch im Vollbild
UNVERÄNDERT. Einziger NEUER Fall: `activateDetailField`s Relations-Sprung-Zweig
(`default:`-Case, `f.beanID != ""`) — innerhalb `fullscreenDetail` darf ein Sprung NICHT
`m.detailFocus = false` setzen (bestehendes Split-Modus-Verhalten):

```go
default: // "" -- Relations jump
    if f.beanID == "" { return m, nil }
    if m.fullscreen == fullscreenDetail {
        // History-Push kommt in Task 8 -- HIER NUR der Zielwechsel:
        m.fullscreenBeanID = f.beanID
        m.secCursor, m.accOpen, m.detailLevel, m.fieldCursor = 0, 1, 0, 0
        return m, nil
    }
    m.expanded = expandAncestorsOf(m.idx, m.expanded, f.beanID)
    m.cursorID = f.beanID
    m.detailFocus = false
    return m, nil
```

(Split-Modus-Zweig bewusst UNVERÄNDERT — kein PO-Wortlaut verlangt History für den Split-
Modus, s. design-spec.md.)

## Ausstieg (`esc`)

**Entscheidung (Planner, s. design-spec.md für die volle Herleitung):** `esc` verlässt das
Vollbild IMMER direkt zurück zu Browse/Backlog (Split-Modus), UNABHÄNGIG vom Eintrittspfad
(direktes `v`, `enter` aus Listen-Vollbild, N Relations-Sprünge) — NICHT schrittweise
zurück durch die Relations-Sprung-Kette (das ist `ctrl+links`/`[`s Aufgabe, Task 8). Fügt
sich in D03s bestehendes „esc = eine Ebene pro Druck"-Modell ein:

| Zustand | `esc`-Wirkung |
|---|---|
| `fullscreenDetail`, `detailLevel==1` | → `detailLevel=0` (bestehende D03-Rast, `keyDetailFocus`s bestehender `Back`-Case wirkt bereits unverändert korrekt, da er nur `m.detailLevel`/`m.detailFocus` liest) |
| `fullscreenDetail`, `detailLevel==0` | → NEU: Vollbild verlassen — `m.fullscreen = fullscreenNone`, `m.detailFocus = true`, UND (je nach `m.view`) `m.cursorID = m.fullscreenBeanID` (Tree, inkl. `m.expanded = expandAncestorsOf(...)` damit der Knoten sichtbar ist) BZW. `m.backlogList`-Cursor auf den Index von `m.fullscreenBeanID` in `backlogVisible()` gesetzt (Backlog) |
| `fullscreenList` | → NEU: `m.fullscreen = fullscreenNone` (Cursor war nie entkoppelt, keine Sync nötig) |

`keyDetailFocus`s bestehender `Back`-Case (`if keybind.Matches(msg, keys.Back) {
if m.detailLevel == 1 { m.detailLevel = 0 } else { m.detailFocus = false } ... }`) MUSS für
den `fullscreenDetail`+`detailLevel==0`-Fall erweitert werden (der bestehende `else`-Zweig
setzt nur `m.detailFocus = false`, das reicht im Vollbild NICHT — der Vollbild-Zustand
selbst muss mit verlassen werden UND die Cursor-Sync passieren). NEUER, dedizierter Zweig
in `keyDetailFocus` (nicht in `keyFullscreen` — der Back-Case lebt bereits an der
korrekten Stelle für Feld/Sektions-Ebenen-Unterscheidung):

```go
if keybind.Matches(msg, keys.Back) {
    if m.detailLevel == 1 {
        m.detailLevel = 0
        return m, nil
    }
    if m.fullscreen == fullscreenDetail {
        id := m.fullscreenBeanID
        m.fullscreen = fullscreenNone
        m.detailFocus = true
        if m.view == viewBacklog {
            vis := m.backlogVisible()
            for i, bb := range vis { if bb.ID == id { m.backlogList.cursor = i; break } }
        } else {
            m.expanded = expandAncestorsOf(m.idx, m.expanded, id)
            m.cursorID = id
        }
        return m, nil
    }
    m.detailFocus = false
    return m, nil
}
```

`fullscreenList`s `esc`: in `keyFullscreen` selbst behandelt (`if m.fullscreen ==
fullscreenList && keybind.Matches(msg, keys.Back) { m.fullscreen = fullscreenNone; return
true, m, nil }`).

## Maus im Vollbild — explizites Nicht-Ziel (v1, Scope-Cut)

`handleMouse` (mouse.go) bekommt EINEN neuen Guard direkt NACH dem bestehenden Toast-Klick-
Vorrang (bleibt unbedingt, unverändert) und VOR dem bisherigen Overlay-Guard:

```go
if m.fullscreen != fullscreenNone {
    return m, nil
}
```

Verhindert, dass ein Klick im Vollbild gegen die (dann falsche) Split-Geometrie fehl-
interpretiert wird. Wheel/Klick sind im Vollbild damit funktionslos, aber NIE falsch —
dokumentierter Nicht-Ziel-Punkt (kein PO-Wortlaut verlangt Maus-Support im Vollbild),
volle Maus-Unterstützung ist ein Fast-Follow außerhalb dieser Runde (kein eigener bean,
YAGNI bis PO-Bedarf).

## Rendering (`view_fullscreen.go`)

```go
// renderFullscreenBody is the shared single-pane renderer for BOTH
// fullscreen flavors (F01, design-spec.md §15) -- Chrome (breadcrumb/
// footer/status) stays IDENTICAL to the split view (browseRepoChrome/
// backlogChrome, unchanged callers), only the body swaps from
// JoinHorizontal(listBox, detailBox) to a single full-width pane.
func renderFullscreenBody(fs fullscreenMode, innerW, bodyH int, listRows []string, focused bool, idx *data.Index, detailBean *data.Bean, secCursor, accOpen, fieldCursor, detailLevel int) string {
    if fs == fullscreenList {
        return renderPane(pane{rows: listRows}, innerW, bodyH, true)
    }
    return renderAccordionPane(idx, detailBean, innerW, bodyH, accOpen, secCursor, fieldCursor, detailLevel, true)
}
```

`viewBrowseRepo()`/`viewBacklog()` (view_browse_repo.go/view_browse_backlog.go) prüfen
`m.fullscreen != fullscreenNone` an ihrem jeweiligen Kopf (VOR dem bestehenden
`JoinHorizontal`-Aufbau) und rufen `renderFullscreenBody` mit ihren eigenen (unverändert
berechneten) `listRows` (Tree- bzw. Backlog-Rows, `bodyH` statt `bodyH-1` — die Vollbild-
Liste hat ebenfalls die Such-/Sortierzeile als erste Zeile, GLEICHE Konvention wie im
Split-Modus) bzw. dem aufgelösten `m.fullscreenBeanID`-Bean auf. `clickPaneGeometry`
bleibt UNVERÄNDERT (liefert weiterhin `lw`/`rw` für den Split-Fall — im Vollbild-Fall wird
stattdessen `innerW` direkt als Pane-Breite verwendet, `bodyH` aus `clickPaneGeometry`
bleibt WEITERHIN die korrekte Höhen-Quelle, unverändert wiederverwendet). Footer/Breadcrumb
zeigen weiterhin den NORMALEN View-Kontext (`browseRepoChrome`/`backlogChrome`) PLUS die
neuen `keys.Fullscreen`/History-Bindings über eine neue kontextsensitive Local-Bindings-
Funktion, s. Notes for Task 8 unten (Footer-Feinschliff für die History-Keys ist explizit
Task 8s Aufgabe — DIESER Task ergänzt nur `keys.Fullscreen` selbst zu den bestehenden
`browseRepoLocalBindings()`/`backlogLocalBindings()`-Listen, damit `v` im Footer sichtbar
ist).

## TDD-Schritte

1. Failing tests (neue Datei `view_fullscreen_test.go` + Ergänzungen in `update_test.go`/
   `mouse_test.go`): `TestKeyFullscreenEntersListModeWhenTreeFocused`,
   `TestKeyFullscreenEntersDetailModeWhenDetailFocused`,
   `TestKeyFullscreenNoOpWhenAlreadyFullscreen`, `TestKeyFullscreenNoOpInLobby`,
   `TestKeyFullscreenEnterOnListEntersDetailMode`,
   `TestKeyFullscreenEnterOnEmptyListIsNoOp`, `TestFocusedBeanResolvesFullscreenBeanID`,
   `TestActivateDetailFieldJumpStaysInFullscreenDetail`,
   `TestKeyDetailFocusEscFieldLevelStepsToSectionLevelInFullscreen` (bestehende D03-Rast,
   Regressionstest im neuen Kontext), `TestKeyDetailFocusEscSectionLevelExitsFullscreenToTree`,
   `TestKeyDetailFocusEscSectionLevelExitsFullscreenToBacklogWithCursorSynced`,
   `TestKeyFullscreenEscExitsListModeToSplitView`,
   `TestHandleMouseIgnoredWhenFullscreenActive` (mirrort `TestMouseDetailClickIgnoredWhenOverlayOpen`),
   `TestRenderFullscreenBodyUsesFullPaneWidth` (Golden-artiger Breiten-Assert, kein
   `-update`-Golden nötig — direkter `lipgloss.Width`-Check gegen `innerW`).
2. `command go test ./internal/tui/... -run "Fullscreen"` → FAIL.
3. Implementieren (Reihenfolge: `types.go` [Modelfelder+Enum] → `keymap.go`
   [keys.Fullscreen] → `update.go` [focusedBean-Erweiterung, handleKey-Dispatch-Punkt,
   keyDetailFocus-Back-Case-Erweiterung, activateDetailField-Erweiterung] →
   `view_fullscreen.go` [keyFullscreen, renderFullscreenBody] → `view_browse_repo.go`/
   `view_browse_backlog.go` [Fullscreen-Zweig in viewBrowseRepo/viewBacklog,
   keys.Fullscreen in *LocalBindings] → `mouse.go` [Guard]).
4. Tests grün.
5. Golden-Regen: NEUE Golden-Fixtures für die Vollbild-Views sind PLAUSIBEL (neue
   Render-Pfade) — prüfen ob bestehende `TestTreeGolden`/`TestBacklogGolden`/
   `TestChromeGolden` unverändert bleiben (Split-Modus-Render-Pfad ist strukturell
   UNVERÄNDERT, nur ein neuer Zweig DAVOR) — falls unverändert: Gegenbeleg OHNE -update,
   im Commit-Body dokumentieren. Falls das Team eine eigene Golden-Suite für Vollbild will
   (`TestFullscreenListGolden`/`TestFullscreenDetailGolden`), als NEUE Golden-Datei(en)
   anlegen (`-update`-Lauf, Erstbeleg statt Diff, da es sie noch nicht gibt) — Implementer-
   Entscheidung, ob nötig für die Akzeptanz oder ob die Unit-Tests aus Schritt 1 ausreichen.
6. `command go test ./... -short` grün (2x), voller Lauf grün, `-race` grün, gofmt/vet leer.
7. **Smoke PFLICHT** (Kaskaden end-to-end testen, LESSONS-LEARNED Eintrag 3): real in tmux
   gegen `./bin/bt` — `v` aus Tree-Fokus → Listen-Vollbild → `enter` auf Bean →
   Detail-Vollbild → mehrere Relations-Sprünge (via `enter` auf selektierten Relations-
   Zeilen, B04.2) → `esc` (Feld→Sektion, falls auf Feld-Ebene) → `esc` (Vollbild verlassen)
   → verifizieren: Split-Ansicht zeigt das ZULETZT im Vollbild gezeigte Bean selektiert,
   NICHT das ursprüngliche. Wiederholen aus Backlog-Kontext. `tmux capture-pane -p`-Belege
   im Commit-Body.
8. Commit `feat(tui): Vollbild-Modus 'v' — Listen-/Detail-Vollbild, esc-Ausstieg (F01
   Kernmechanik)`. Footer `Refs: bt-tct9`.

## Akzeptanz-Checkliste

- [x] `v` bei Tree-/Backlog-Fokus → Listen-Vollbild; bei Detail-Fokus → Detail-Vollbild
- [x] `v` im bereits aktiven Vollbild ist No-Op (kein Toggle)
- [x] `v` in der Lobby ist No-Op
- [x] `enter` auf einem Bean im Listen-Vollbild → Detail-Vollbild für dieses Bean
- [x] Alle bestehenden Feld-Kaskaden (status/type/priority/tags/title) funktionieren
      unverändert im Detail-Vollbild
- [x] Relations-Sprung im Detail-Vollbild bleibt im Vollbild (zeigt das neue Zielbean),
      verlässt NICHT zum Split-Tree (Split-Modus-Sprungverhalten bleibt separat unverändert)
- [x] `esc` kaskadiert Feld→Sektion→Vollbild-Exit (Detail-Vollbild) bzw. direkt Exit
      (Listen-Vollbild) — IMMER zurück zu Browse/Backlog, mit dem zuletzt gezeigten Bean
      selektiert
- [x] Symmetrisch für Browse (Tree) UND Backlog
- [x] Mausklicks im Vollbild sind No-Op (kein Fehlklick gegen falsche Geometrie)
- [x] Kein `viewID`-Enum-Wert hinzugefügt (Vollbild bleibt orthogonal zu `m.view`)
- [x] tmux-Smoke-Kaskade end-to-end belegt (Commit-Body)
- [x] Voller Testlauf grün, gofmt/vet leer

## Notes for Task 8 (History-Stack)

- `m.navBack`/`m.navForward []string` sind bereits deklariert (types.go), aber leer/
  ungenutzt — Task 8 füllt sie im Relations-Sprung-Zweig (`activateDetailField`s
  `fullscreenDetail`-Case oben) UND in den neuen `ctrl+links`/`ctrl+rechts`-Handlern.
- Der `esc`-Exit-Pfad (`keyDetailFocus`s Back-Case) leert `navBack`/`navForward` NICHT
  (design-spec.md: "werden beim Verlassen NICHT geleert") — Task 8 muss das NICHT
  nachrüsten, nur beachten (kein Reset an dieser Stelle einbauen).
- Neue `keys.HistoryBack`/`keys.HistoryForward`-Bindings UND die kontextsensitive
  `fullscreenDetailLocalBindings()`-Footer-Ergänzung sind vollständig Task 8s Aufgabe,
  hier bewusst nicht vorgegriffen.


## PRELUDE (2026-07-16, aus T6-Review F01 — ZUERST erledigen, eigener Commit)

Test-Härtung, low: D02 beseitigte die letzte real-gerenderte Footer-Höhen-Divergenz —
die Click-Boundary-Mathematik (clickPaneGeometry: bodyH/footerY via footH =
lipgloss.Height(localKeys)+2) wird nur noch bei footH-Wert 2 gegen echten Chrome-Render
geprüft. Härtung: synthetischer clickPaneGeometry-Test mit echtem mehrzeiligem
(>=3-Zeilen) localKeys-String, der die Boundary-Auflösung (bodyH/footerY) asserted —
beweist die Generalität des dynamischen Mechanismus unabhängig von View-Divergenz.
Commit: `test(tui): clickPaneGeometry multi-line footH pin (T6-F01)`, `Refs: bt-13l7`.

## Summary

F01 Kernmechanik implementiert: `fullscreenMode`-Enum (`types.go`, orthogonal
zu `m.view`) + `fullscreen`/`fullscreenBeanID`-Felder (`navBack`/`navForward`
deklariert, leer, Task 8). `keys.Fullscreen` (`v`, `keymap.go`) + neue Datei
`view_fullscreen.go` (`keyFullscreen`, `renderFullscreenBody`). `handleKey`
(update.go) dispatcht `keyFullscreen` nach FocusIn/FocusOut, vor
keyNodeAction; `focusedBean()` bekommt einen vorrangigen fullscreenDetail-
Fall; `activateDetailField`s Relations-Sprung bleibt bei fullscreenDetail im
Vollbild (neuer Fall); `keyDetailFocus`s Back-Case bekommt die neue
D03-Ausstiegs-Rast (Vollbild verlassen + Cursor-Sync Tree/Backlog).
`mouse.go` bekommt den Vollbild-Guard. `viewBrowseRepo()`/`viewBacklog()`
rendern bei `m.fullscreen != fullscreenNone` einen neuen, einzelnen
Vollbreite-Pane-Zweig VOR dem bestehenden Split-Aufbau (Split-Pfad
unverändert).

## Test-Output

RED (Implementierung komplett entfernt via `git checkout --`/Datei
verschoben, NUR Testdateien behalten):
```
# beans-tui/internal/tui [beans-tui/internal/tui.test]
internal/tui/mouse_test.go:306:4: m.fullscreen undefined (type model has no field or method fullscreen)
internal/tui/mouse_test.go:306:17: undefined: fullscreenList
internal/tui/update_test.go:859:4: m.fullscreen undefined (type model has no field or method fullscreen)
internal/tui/update_test.go:859:17: undefined: fullscreenDetail
...
FAIL	beans-tui/internal/tui [build failed]
```

GREEN (Implementierung wiederhergestellt, alle 21 neuen F01-Tests):
```
--- PASS: TestHandleMouseIgnoredWhenFullscreenActive
--- PASS: TestFocusedBeanResolvesFullscreenBeanID
--- PASS: TestFocusedBeanFullscreenDetailUnknownBeanIDReturnsNil
--- PASS: TestActivateDetailFieldJumpStaysInFullscreenDetail
--- PASS: TestKeyDetailFocusRoutesToDetailFocusMachineViaFullscreenAlone
--- PASS: TestKeyDetailFocusEscFieldLevelStepsToSectionLevelInFullscreen
--- PASS: TestKeyDetailFocusEscSectionLevelExitsFullscreenToTree
--- PASS: TestKeyDetailFocusEscSectionLevelExitsFullscreenToBacklogWithCursorSynced
--- PASS: TestKeyFullscreenEntersListModeWhenTreeFocused
--- PASS: TestKeyFullscreenEntersDetailModeWhenDetailFocused
--- PASS: TestKeyFullscreenNoOpWhenAlreadyFullscreenList
--- PASS: TestKeyFullscreenNoOpWhenAlreadyFullscreenDetail
--- PASS: TestKeyFullscreenNoOpInLobby
--- PASS: TestKeyFullscreenEnterOnListEntersDetailMode
--- PASS: TestKeyFullscreenEnterOnListEntersDetailModeBacklog
--- PASS: TestKeyFullscreenEnterOnEmptyListIsNoOp
--- PASS: TestKeyFullscreenEscExitsListModeToSplitView
--- PASS: TestKeyFullscreenEscExitsListModeToSplitViewBacklog
--- PASS: TestFullscreenNeverChangesViewID
--- PASS: TestViewBrowseRepoFullscreenFitsOuterFrame
--- PASS: TestViewBacklogFullscreenFitsOuterFrame
PASS
```

Voller Lauf (2x, count=1): `go test ./... -short` grün (cmd/config/data/
theme/tui). `go test ./... -race -count=1`: grün (tui-Paket 141-142s).
`gofmt -l .` leer, `go vet ./...` leer.

## Golden-Gegenbeleg

`git diff --stat -- internal/tui/testdata/` -> leer (tree.golden/
backlog.golden/chrome.golden byte-identisch, `TestTreeGolden`/
`TestBacklogGolden`/`TestChromeGolden` grün ohne `-update`). Split-Renderpfad
in `viewBrowseRepo()`/`viewBacklog()` strukturell unverändert (neuer Zweig
NUR bei `m.fullscreen != fullscreenNone`, davor).

Entscheidung eigene Vollbild-Goldens: NEIN. Begründung: `TestViewBrowseRepo
FullscreenFitsOuterFrame`/`TestViewBacklogFullscreenFitsOuterFrame` (neu,
render-grounded auf `m.View()`-Ebene, beide Vollbild-Spielarten x beide
Views) + die 21 State-/Kaskaden-Unit-Tests + der komplette tmux-Live-Smoke
decken Rendering UND Zustandsübergänge bereits ab; ein zusätzliches
Golden-Fixture-Paar hätte hier primär Byte-Diffs bei jeder kosmetischen
Änderung gepflegt, ohne einen Fehlerfall abzudecken, den die Breiten-/
Höhen-Asserts nicht schon abdecken.

## Smoke (tmux, Scratch-Repo /tmp/bt-smoke-f01, gelöscht nach Lauf)

Fixture: Milestone -> Epic One -> Task A (blocking Task B), Epic Two ->
Task B; + 2 parentlose Backlog-Beans A/B (A blocking B).

**Tree-Kaskade** (100x30): `l`/`k`/`l` Tree expandieren, Cursor auf Task A ->
`v` (Listen-Vollbild, Einzelpane volle Breite, Cursor weiterhin Task A) ->
`enter` (Detail-Vollbild Task A, META offen) -> `3` (RELATIONS) -> `l`
(Feld-Ebene, Parent) -> `k` (Blocking/Task B) -> `enter` (Feld-Ebene ->
aktiviert -> Sprung, BLEIBT Vollbild, zeigt jetzt Task B) -> `esc`
(Sektions-Ebene -> Vollbild verlassen) -> Split zeigt Tree-Cursor auf
Task B, Epic Two automatisch expandiert, Detail-Pane zeigt Task B
fokussiert. Zusätzlich `v`/`v` (zweites `v` im Vollbild = No-Op, Inhalt
unverändert) und `v` in der Lobby (`p` -> `v` tippt "v" ins Repo-Filter,
KEIN Vollbild-Effekt -- Lobby fängt Input vollständig ab, wie erwartet).

**Backlog-Kaskade** (Backlog-Ansicht via `shift+tab`+`b`): Cursor auf
Backlog Bean A -> `v` (Listen-Vollbild, `sort status`-Suffix erhalten) ->
`enter` (Detail-Vollbild Bean A) -> `3` (RELATIONS, Blocking Bean B
vorselektiert) -> `enter` (Feld-Ebene, kein Sprung) -> `enter` (aktiviert,
Sprung -> Bean B, BLEIBT Vollbild) -> `esc` -> Backlog-Split zeigt
`backlogList.cursor` auf Bean B, Detail-Pane Bean B fokussiert.

**Bug live gefunden + gefixt:** erster Smoke-Lauf zeigte einen zerrissenen
äußeren Rahmen (Bottom-Border wickelte auf eine eigene Zeile) -- Ursache:
`renderFullscreenBody`-Aufruf reichte `innerW` statt `innerW-2` als
Pane-Content-Breite durch (Border-Doppelzählung, 2 Spalten Überlauf).
Diagnostiziert über einen `m.View()`-Zeilen/Breiten-Dump (nicht sichtbar in
den isolierten `renderFullscreenBody`-Unit-Tests, die NUR die Funktion
selbst prüften, nicht den Call-Site-Breiten-Vertrag). Gefixt (`paneW :=
innerW - 2` an beiden Call-Sites) + Regressionstest
`TestViewBrowseRepoFullscreenFitsOuterFrame`/`...Backlog...` auf
`m.View()`-Ebene ergänzt. Nach Fix: sauberer Rahmen in allen obigen
Captures, Breiten/Höhen exakt im 100x30- bzw. 80x24-Fenster.

## Deviations/ERRATA

**D01 (Implementierer-Entscheidung, kein Raten):** `keys.Fullscreen` NICHT
zu `browseRepoLocalBindings()`/`backlogLocalBindings()` (Footer Zone 3)
hinzugefügt -- abweichend vom bean-Wortlaut ("ergänzt nur keys.Fullscreen
selbst zu den bestehenden ...-Listen"). Grund: 14 statt 13 Footer-Einträge
brechen bei 80 Spalten von 2 auf 3 Zeilen um (gemessen:
`renderBindings(browseRepoLocalBindings()+Fullscreen)` = 167 Zellen Inhalt
gegen 78 Spalten -> 3 Zeilen; ohne Fullscreen 152 Zellen -> 2 Zeilen) --
exakt die Regression, die D02 (bt-1e0t, Sort-Footer-Entzug) bereits einmal
für den identischen 14-Entries-Fall gefixt hat ("D06 permits at most two").
`TestDetailClickBacklogFooterAt80Cols`s Precondition (`footH == 2` für
BEIDE Chromes bei 80 Spalten) hätte sofort gebrochen. Entscheidung: `v`
bleibt voll funktional + in `helpGroups()` (Navigation-Gruppe,
`TestHelpGroupsCoverEveryBindingExactlyOnce` grün) dokumentiert, NICHT im
Footer -- mirrort D02s eigenen, PO-bestätigten Präzedenzfall 1:1
(funktional aber Help-only, wenn das Footer-Budget überschritten würde).
Kein Code-Regressionsrisiko, aber PO-Sichtbarkeits-Trade-off, den ein
Planner/PO explizit bestätigen sollte, falls `v` doch im Footer sichtbar
sein MUSS (dann müsste ein anderer Footer-Eintrag weichen).

**B01 (im Zuge dieser Runde gefunden + gefixt, kein separater Bug-Report
nötig):** die oben beschriebene Vollbild-Breiten-Überlauf (innerW vs.
innerW-2) -- siehe Smoke-Abschnitt.

## Notes for T8 (History-Stack, bean bt-1vbp)

- `m.navBack`/`m.navForward` (`[]string`, `types.go`) sind bereits deklariert,
  leer/ungenutzt.
- Relations-Sprung-Zweig, der T8 den Push braucht: `activateDetailField`
  (`update.go`), `default:`-Case, der NEUE `if m.fullscreen ==
  fullscreenDetail { m.fullscreenBeanID = f.beanID; ... }`-Block -- History-
  Push kommt DAVOR (`m.navBack = append(clone(m.navBack),
  <bisheriges fullscreenBeanID>); m.navForward = nil`), bevor
  `m.fullscreenBeanID` überschrieben wird.
- Neue `keys.HistoryBack`/`keys.HistoryForward`-Bindings + eine
  `fullscreenDetailLocalBindings()`-Footer-Ergänzung sind vollständig T8s
  eigene Aufgabe (hier bewusst nicht vorgegriffen) -- ACHTUNG:
  `browseRepoLocalBindings()`/`backlogLocalBindings()` liegen (D01 oben)
  bereits an ihrem 2-Zeilen-Footer-Budget bei 80 Spalten, `[`/`]` müssten
  in einem EIGENEN, kontextsensitiven Footer-Set landen (nur sichtbar
  während `fullscreenDetail`), nicht in diesen beiden Listen.
- `keyDetailFocus`s Back-Case (esc-Exit-Pfad) leert `navBack`/`navForward`
  NICHT (design-spec.md: "werden beim Verlassen NICHT geleert") -- T8 muss
  hier nichts nachrüsten, nur beachten.
- `keyFullscreen` (`view_fullscreen.go`) ist der Ort für `ctrl+left`/`[`
  bzw. `ctrl+right`/`]` -- mirrort den bestehenden
  `m.fullscreen == fullscreenList`-Vorab-Check-Stil (neuer Block, NUR
  wirksam bei `m.fullscreen == fullscreenDetail`).
