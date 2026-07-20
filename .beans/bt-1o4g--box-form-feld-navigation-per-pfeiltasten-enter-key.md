---
# bt-1o4g
title: 'Box-Form: Feld-Navigation per Pfeiltasten + Enter (keyboard-first)'
status: todo
type: feature
priority: high
created_at: 2026-07-20T07:31:22Z
updated_at: 2026-07-20T07:31:22Z
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
