---
# bt-1o4g
title: 'Box-Form: Feld-Navigation per Pfeiltasten + Enter (keyboard-first)'
status: completed
type: feature
priority: high
created_at: 2026-07-20T07:31:22Z
updated_at: 2026-07-20T09:13:15Z
parent: bt-vy1q
blocked_by:
    - bt-ze10
---

**Nebenbefund N8 (PO, 2026-07-20).** Nach `tab` (Fokus ins Detail) laesst sich in der Box-Form weder mit der Maus scrollen noch mit den Pfeiltasten zwischen den Feldern navigieren. **Keyboard-first ist eine Kern-Achse dieser TUI** — ohne Feld-Cursor ist das Detail nur per globalen Hotkeys bedienbar, nicht navigierbar.

Der Maus-Scroll-Teil wird bereits von `bt-ze10` erledigt. Dieses Bean deckt die **Feld-Navigation**.

## Ausgangslage
- Das Accordion HATTE einen Feld-Cursor (`metaSectionBody`s ▶-Marker, `fieldCursor`, `keyDetailFocus` in `update.go`).
- Die Box-Form hat ihn nicht — in S2b bewusst zurueckgestellt ("field-level nav is a later concern").
- **Wichtig:** `dropdownBox(label, value, hotkey, width, focused bool)` (internal/tui/box_dropdown.go) besitzt bereits einen `focused`-Parameter, der den Rahmen auf `theme.Mauve` setzt. Er wird aktuell IMMER mit `false` aufgerufen. Die Darstellung ist also schon gebaut — sie muss nur angesteuert werden. Gleiches gilt fuer `panelBox`.

## Ziel
1. Feld-Cursor im Box-Detail, wenn die Detail-Pane fokussiert ist (`tab`).
2. Pfeiltasten (und die jkli-Aliase via `navKey`) bewegen den Cursor in der FESTEN Reihenfolge von `detailBoxForm`: Title → Status → Type → Priority → Parent → Tags → Body → Relations → History. Links/rechts innerhalb einer Grid-Zeile (Status|Type|Priority bzw. Parent|Tags), hoch/runter zwischen den Zeilen.
3. Das fokussierte Feld rendert mit `focused=true` (Mauve-Rahmen) — sichtbares Ziel.
4. `enter` auf dem fokussierten Feld oeffnet dessen Editor — dieselben Handler wie die Hotkeys: Status→`openValueMenu("status")`, Type→`"type"`, Priority→`"priority"`, Parent→`openParentPicker()`, Tags→`openTagPicker()`, Title/Body→`openBeanEditor(b)`.

## Betroffen
- `internal/tui/box_detail_form.go` — `detailBoxForm` braucht den Cursor-Index als Parameter und reicht `focused` an die richtige Box durch
- `internal/tui/types.go` — Cursor-Feld
- `internal/tui/update.go` — Pfeil-/enter-Behandlung, wenn Detail fokussiert + Box-Modus (`keyDetailFocus`-Umfeld)
- `internal/tui/view_browse_repo.go` — `boxFormEnabled()`-Zweig in `renderAccordionPane`

## Zusammenspiel mit bt-ze10 (blocked-by)
`bt-ze10` fuehrt den Scroll-Offset ein. Beim Bewegen des Cursors muss das Zielfeld **in den sichtbaren Bereich gescrollt** werden (scroll-into-view), sonst wandert der Cursor unsichtbar aus der Pane. Deshalb erst ze10, dann dieses Bean.

## Zu beachten
- Cursor zuruecksetzen, wenn der selektierte Bean wechselt
- Accordion-Modus (Flag AUS) bleibt voellig unveraendert — dessen eigene Feld-Navigation nicht anfassen
- `shift+tab`/`esc` verlassen das Detail wie bisher
- Golden: der Mauve-Fokus-Rahmen aendert `browse_boxform.golden`, wenn der Cursor beim Rendern aktiv ist — bewusst regenerieren

## Akzeptanz
- [ ] Nach `tab` ist ein Feld sichtbar fokussiert (Mauve-Rahmen)
- [ ] Pfeiltasten bewegen den Cursor ueber alle Felder inkl. links/rechts in den Grid-Zeilen
- [ ] `enter` oeffnet den Editor des fokussierten Feldes (alle 6 Feldtypen getestet)
- [ ] Cursor scrollt in den sichtbaren Bereich (Zusammenspiel mit ze10)
- [ ] Cursor resettet bei Bean-Wechsel
- [ ] Accordion-Modus unveraendert, Bestandsgolden byte-identisch
- [ ] Voller `command go test ./...` gruen


## Summary

Feld-Cursor fuer die Box-Form (BT_BOXFORM=1), Commit 75d791a.

- `internal/tui/box_nav_field.go` (neu): `boxFormFieldOrder` als EINZIGE
  Feld-Tabelle (name/target/row/col), `boxFormFieldAt` (Row/Col -> Feld),
  `boxFormMoveField` (Pfeil-Navigation), `boxFormScrollIntoView`,
  `boxFormNav` (die eine Tastatur-Mutation von Cursor UND Scroll),
  `boxFormActivateCursor`, `boxFormEffectiveCursor` (derived reset bei
  Bean-Wechsel, analog `boxFormEffectiveScroll`).
- `box_detail_form.go`: `detailBoxForm(idx, b, width, cursor)`; neu
  `boxFormBlocks` (sechs Layout-Bloecke einzeln) + `boxFormFieldSpan`
  (Zeilen-Span je Feld) + `lineCount`. `gridRow` bekommt `focusedCol`.
- `mouse.go`: `boxFormPaneMetrics` aus `boxFormScrollBounds` extrahiert
  (accW/height fuer Cursor UND Clamp aus EINER Geometrie);
  `detailBoxFormClickRow` mappt jetzt ueber `boxFormFieldAt` statt eigenem
  Inline-Switch; Dispatch als `activateBoxFormTarget` extrahiert (geteilt
  von Klick und enter).
- `update.go` keyDetailFocus: im `boxFormEnabled() && !fullscreenDetail`-
  Zweig gehen alle vier Pfeile an `boxFormNav`, enter an
  `boxFormActivateCursor`.
- `types.go`: `boxFormCursor` / `boxFormCursorBean`.
- `view_browse_repo.go` / `view_fullscreen.go`: `renderAccordionPane` hat
  einen `boxCursor`-Parameter; Split-Detail reicht den Cursor NUR bei
  `focused` durch, Fullscreen immer -1 (dort ist die Navigation
  ohnehin ausgeguertet).

### Entscheidung: Feld-Fokus vs. Scroll

Der Fokus fuehrt, das Viewport folgt. left/right bewegen den Cursor in
seiner Grid-Zeile, danach zieht scroll-into-view den Offset nach.
up/down "enthuellen erst, bewegen dann": solange das fokussierte Feld in
der gedrueckten Richtung ueber den sichtbaren Bereich hinausragt,
scrollt der Tastendruck um eine Zeile und laesst den Cursor stehen; erst
wenn die Kante sichtbar ist, springt er aufs Nachbarfeld (und wird
sichtbar gescrollt). Damit bleibt bt-ze10s Kernzusage (langer Body per
Tastatur bis zur letzten Zeile erreichbar) erhalten, obwohl up/down
nicht mehr kontextfrei scrollen -- jetzt sogar mit sichtbarem Kontext,
welches Feld man gerade durchlaeuft. Das Mausrad bleibt bewusst freier
Viewport-Scroll ohne Fokus-Semantik.

Ein Cursor ist nur sichtbar, wenn die Detail-Pane wirklich fokussiert
ist; ohne Fokus rendert die Form mit Cursor -1, also byte-identisch zu
vorher. Deshalb musste KEINE Golden-Datei angefasst werden.

## Test-Output

Voller Lauf `command go test ./... -count=1` (ohne -short):

```
?   	github.com/xRiErOS/beans-tui	[no test files]
ok  	github.com/xRiErOS/beans-tui/cmd	1.706s
?   	github.com/xRiErOS/beans-tui/internal/clip	[no test files]
ok  	github.com/xRiErOS/beans-tui/internal/config	1.203s
ok  	github.com/xRiErOS/beans-tui/internal/data	4.867s
ok  	github.com/xRiErOS/beans-tui/internal/theme	0.407s
ok  	github.com/xRiErOS/beans-tui/internal/tui	153.822s
```

Neue Tests in `box_nav_field_test.go`: Feld-Reihenfolge deckungsgleich
mit der Klick-Geometrie; Mauve-Rahmen exakt auf dem Span des cursorten
Feldes; `tab` macht ein Feld im echten `View()` sichtbar fokussiert;
Pfeil-Reihenfolge inkl. Spalten-Clamp und Kanten-No-Op; alle Panels
erreichbar ohne Wrap; Cursor bleibt bei JEDEM down im sichtbaren
Fenster; down/up scrollen durch ein zu hohes Feld; enter oeffnet fuer
alle sechs Feldtypen den richtigen Editor; Cursor-Reset bei
Bean-Wechsel; ohne Flag komplett inert.

80-Spalten-tmux-Smoke gefahren (nicht Pflicht, Footer unberuehrt):
tab -> Title mauve, down -> Status, right -> Type, enter -> "Set type"
auf `epic (current)`, weiteres down -> Body-Panel fokussiert und ins
Bild gescrollt. Kein Overflow, kein Wrap-Bug.

## Deviations

1. **bt-ze10s Tastatur-Scroll umgewidmet.** up/down sind kein blindes
   +/-1 auf `boxFormScroll` mehr (siehe Entscheidung oben).
   `TestBoxFormScrollDownUpWhenDetailFocused` wurde entsprechend
   umgeschrieben (mit Begruendungs-Kommentar im Test); die exakte
   +/-1-Zusage lebt jetzt in `TestBoxFormDownScrollsThroughATallField`
   dort, wo sie hingehoert.
2. **Fullscreen-Detail bleibt ohne Feld-Cursor** — dort ist die
   Box-Form-Navigation seit bt-ze10 ausgeguertet (andere Pane-Geometrie);
   ein Mauve-Rahmen, den nichts bewegen kann, waere schlechter als
   keiner. Gleiche Scope-Grenze wie bt-ze10.
3. **Vorbefund (nicht in diesem Bean gefixt):** `detailBoxFormClickRow`
   rechnet `boxFormScroll` nicht ein — nach dem Scrollen trifft ein
   Klick die Feld-Zone der UNGESCROLLTEN Anordnung. Stammt aus bt-ze10
   (Scroll) x S6 (Maus), von diesem Bean weder verursacht noch
   verschlimmert. Lohnt ein eigenes bean.
