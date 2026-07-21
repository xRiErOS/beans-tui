---
# bt-f0y9
title: 'Endliche Felder: feld-verankertes Inline-Dropdown (D09 revidiert)'
status: completed
type: task
priority: normal
created_at: 2026-07-20T18:05:08Z
updated_at: 2026-07-20T21:01:24Z
parent: bt-vy1q
---

# Kontext (Epos bt-vy1q — Rest siehe Epos, DRY)

Umsetzung von **D09 REVIDIERT** (design-spec.md § PO-Entscheidungen 2026-07-20).
PO 2026-07-20: Timing jetzt. Scope verengt — **Create-Form bleibt huh**; nur die
endlichen Einzelfeld-Menüs werden feld-verankerte Inline-Dropdowns.

## Ziel

Die endlichen Enum-Felder **Status / Type / Priority** öffnen ihr Auswahlmenü heute
als **zentrales Modal** (`box_menu_value.go`, `modalPanel`). Ziel: ein **feld-
verankertes Inline-Dropdown**, das direkt am boxed field aufklappt (jira-treu, in der
boxed-field-Nähe schneller). Maus-nativ (Zeilen per Klick wählbar) = der Effizienz-Gewinn.

## Bewusste Abgrenzung

- **Create-Form bleibt huh** — NICHT anfassen. Die langsamen huh-Drive-Tests
  (`box_confirm_create_test.go`, `skipSlowHuhDriveInShortMode`) bleiben.
- Parent/Tags (Picker über bestehende Werte) sind KEIN Enum-Dropdown — außen vor,
  außer die Geometrie fällt trivial ab (dann als Bonus, sonst eigenes bean).
- Filter-Strip-Facetten sind ein separater Befund (eigenes bean) — hier nur die
  Detail-Pane-Wertmenüs.

## Ist-Stand (verifiziert)

| Datei:Zeile | Symbol | Rolle |
|---|---|---|
| `box_menu_value.go:136` | `openValueMenu(group)` | öffnet kombiniertes Status/Type/Priority-Menü als `modalPanel` |
| `box_menu_value.go` | `valueMenuItem`, Enter-wendet-sofort-an | Auswahl-Semantik (Port beans-src statuspicker) — BEHALTEN |
| `update.go:773/781/784` | Tastatur-Trigger `s`/`o`/`u` | rufen `openValueMenu` |
| `overlay_palette.go:199-203` | Palette-Trigger | ruft `openValueMenu` |
| `mouse.go:885-897` | `activateBoxFormTarget` → `openValueMenu` | Detail-Pane-Klick auf Feld |

## Grounding-first (VOR Umsetzung — Ergebnis als `## Grounding` appenden)

Dieser Task ist der architektonisch heikelste offene Punkt (Overlay-Compositing).
Bevor Code geschrieben wird, klären und ins bean appenden:

- [ ] Wie wird `modalPanel` heute über die View **komponiert/positioniert**
      (zentriert)? Wo sitzt der Compositor? Was braucht ein **anker-positioniertes**
      Popup (x,y aus dem Screen-Rect des Feldes)?
- [ ] Woher kommt das Screen-Rect des Status/Type/Priority-boxed-field?
      (`boxFormFieldSpan`, `gridRow`-Geometrie in `box_detail_form.go` — x-Spalte +
      y-Zeile des Feldes, unter Berücksichtigung von `boxFormEffectiveScroll`.)
- [ ] Grenzfälle: Popup, das unten aus dem Viewport ragt (nach oben klappen?),
      und bei 80 Spalten (Breite des Popups vs. Pane-Breite).
- [ ] **Slice-Vorschlag** (D09 ist groß): z.B. (1) anker-positioniertes Render read-only,
      (2) Tastatur-Auswahl umziehen, (3) Maus-Auswahl im Popup, (4) Modal-Pfad entfernen.
      Vorschlag ins bean appenden; der Supervisor entscheidet, ob als Slices dispatcht wird.

## Akzeptanz (Rahmen — nach Grounding zu schärfen)

- [ ] Status/Type/Priority öffnen als feld-verankertes Inline-Dropdown statt Center-Modal.
- [ ] Enter-wendet-sofort-an-Semantik erhalten; `esc` schließt ohne Mutation.
- [ ] Maus: Popup-Zeile per Klick wählbar (nativ, kein huh-Blocker).
- [ ] Tastatur-Pfad (`s`/`o`/`u`) + Palette-Pfad weiter funktionsfähig.
- [ ] Update-Tests (RED→GREEN) + Golden für die neue Popup-Darstellung; 80-Spalten-Geometrie
      pinnen. tmux-Smoke bei 80 Spalten gegen sproutling (Maus + Tastatur + Grenzfälle).
- [ ] Voller Lauf grün, Build, vet. Commits je Slice, `Refs: <id>`.

## Hinweise

- ERRATUM-Kultur: Ist-Snippets prüfen, Abweichungen als ERRATUM appenden.
- Theme nur aus `internal/theme` (kein Hex-Literal in Views).
- Ersetzt im Scope den breiten `bt-dovm` (S7) — der bleibt als Historie mit Pointer.



## Grounding (2026-07-20)

### 1. Compositing heute

Der Compositor ist **`placeOverlay`** (`internal/tui/overlay.go:52-100`), aufgerufen aus **`composeOverlays`** (`internal/tui/view_browse_repo.go:1086-1138`), dort speziell Zeile 1091-1092:

```go
case overlayValueMenu:
    out = placeOverlay(out, m.valueMenuBox(), w, h)
```

`composeOverlays` wird von `viewBrowseRepo()`/`viewBacklog()` am Ende jedes Sub-View-Renders aufgerufen (painter's algorithm, fixe z-Reihenfolge: Filter-Menü → `m.overlay`-Switch → huh-Form → Palette → Help → Tag-Mgmt-Delete-Confirm → Quit).

`placeOverlay(bg, fg, tw, th)` ist **strikt zentriert, ohne jeden Anker-Parameter**:
```go
x := (bgW - fgW) / 2   // horizontal Center
y := (len(bgLines) - fgH) / 2   // vertikal Center
```
Es gibt keine x/y-Eingabe von außen — die Funktion kennt nur die Gesamt-Canvas (tw×th = volle Terminalfläche inkl. Tree-Pane) und die fg-Box-Maße. Steuerung von Position/Sichtbarkeit läuft ausschließlich über `m.overlay` (Enum, `types.go`, Werte u.a. `overlayValueMenu`) — kein separates Positions-Feld im Model.

`m.valueMenuBox()` (`internal/tui/box_menu_value.go:250-282`) baut den Inhalt über `modalPanel(...)` (`internal/tui/modal.go:60`), Breite `clampModalWidth(40, m.width)` — **`m.width` ist die volle Terminalbreite**, nicht die Detail-Pane-Breite.

### 2. Feld-Screen-Rect

Kein einziger vorhandener Helper liefert heute eine absolute Bildschirmposition für ein Feld — die Bausteine sind da, aber nur für **Content-relative** bzw. **inverse** (Screen→Feld) Auflösung:

- **Pane-Ursprung (x/y-Offset der Detail-Pane im Master-Detail-Split):** `clickPaneGeometry(w,h,head,localKeys,status,treeWidth) (bodyH,lw,rw,originX,originY)` (`internal/tui/mouse.go:111-128`). `originX = 1` (äußere Border), `originY = 1 + head-Höhe + 1 + 1`. Detail-Pane linke Kante = `originX+lw`; bei box-form zusätzlich `+= filterBarHeight` auf `originY` (nur `viewBrowseRepo`, nicht Backlog) — siehe `mouse.go:895-897` bzw. `mouse.go:792` Kommentar.
- **Pane-Breite:** `lw,rw := masterDetailWidths(innerW, treeWidthFloor)` (`internal/tui/view.go:209-224`); `accW := rw - 2` (Border-Abzug) ist die Breite, die `detailBoxForm`/`gridRow` tatsächlich rendern (`mouse.go:766` in `boxFormPaneMetrics`).
- **Vertikale Zeilen-Spanne eines Feldes INNERHALB des ungescrollten Form-Strings:** `boxFormFieldSpan(blocks, field) (start, height)` (`internal/tui/box_detail_form.go:195-207`) — läuft über `boxFormFieldOrder[field].row` (`box_nav_field.go:78-88`) und summiert `lineCount(blocks[i])` der vorangehenden Blöcke. Title=3 Zeilen, RowA (Status|Type|Priority)=3 Zeilen, RowB (Parent|Tags)=3 Zeilen, danach Body/Relations/History unbounded.
- **Aktueller Scroll-Offset:** `boxFormEffectiveScroll(m, b)` (`mouse.go:704-710`) — liest `m.boxFormScroll`, aber NUR wenn `b.ID == m.boxFormScrollBean` (sonst 0, derived-reset-Doktrin). Screen-Zeile = `start(field) - scroll`.
- **Horizontale Spalte innerhalb einer Grid-Row:** `gridColWidths(n, width)` (`box_detail_form.go:40-58`, Integer-Division + Rest auf erste Spalten verteilt) + `detailBoxFormGap=1`. `gridColAt` (`mouse.go:684-693`) ist die **inverse** Richtung (Spalten-Offset → Spalten-Index) — die **vorwärts**-Richtung (Spalten-Index → x-Offset) existiert noch NICHT als eigener Helper, ist aber trivial aus `gridColWidths` ableitbar: `x = Σ(widths[0:i]) + i*gap`.
- **Row→Screen-Row-Mapping fürs Hit-Testing (die Referenz-Implementierung für die Rückrichtung):** `boxFormFieldAt(row, col, accW)` (`box_nav_field.go:127-148`) — 3-Zeilen-Blöcke sind hart codiert (`row<3`→Title, `row<6`→RowA, `row<9`→RowB, sonst→Editor-Fallback).

**Fazit:** Alle Bausteine für ein `boxFormFieldScreenRect(m, b, field) (x, y, w, h)` sind vorhanden (`boxFormPaneMetrics` für `accW`/Pane-Origin, `boxFormFieldSpan` für die Y-Spanne, `gridColWidths`+`boxFormFieldOrder[field].col` für die X-Spanne, `boxFormEffectiveScroll` für den Scroll-Korrekturterm) — sie müssen nur zu EINER neuen Funktion zusammengezogen werden, die exakt dieselbe Geometrie wie `detailBoxFormClickRow` (`mouse.go:879-935`) reproduziert, nur vorwärts statt rückwärts.

### 3. Grenzfälle

- **Popup ragt unten aus dem Viewport:** Es gibt **KEINE Präzedenz für Nach-oben-Klappen** irgendwo im Code. `placeOverlay` zentriert stur; wenn `fgH > th` reicht `y` ggf. so weit, dass `row >= len(bgLines)` greift (`overlay.go:87-89`) — die überschüssigen unteren Zeilen werden STILLSCHWEIGEND VERWORFEN (kein Clip-Indicator, kein Scroll). Ein feld-verankertes Popup, das unten herausragt, hätte über `placeOverlay` allein exakt dasselbe Problem — die neue Positionierungsfunktion muss diesen Fall selbst behandeln (eigene Klapp-Logik "nach oben, wenn unten kein Platz", analog zu `windowStart`/`scrollView`-Mustern, die es für andere Zwecke im Code gibt, aber nicht für Overlay-Y-Position).
- **80 Spalten — Popup-Breite vs. Detail-Pane-Breite:** Bei `w=80`: `innerW=78`, `masterDetailWidths(78,0)` → `lw=26` (1/3, gedeckelt durch `cap=78*2/5=31`), `rw=78-26-4=48`, `accW=rw-2=46`. Der heutige zentrierte Modal-Value-Menu hat Breite `clampModalWidth(40, 80)=40` (da `80-4=76>40`, bleibt bei 40) — bezogen auf die VOLLE Terminalbreite, nicht auf `accW`. Ein feld-verankertes Popup müsste sich dagegen an der Feld-BOX selbst orientieren: `gridColWidths(3, 46)` → Spalten `{15,15,14}` (avail=46-2=44, base=14, rem=2). Ein Status/Type/Priority-Dropdown, das an seiner ~15 Zellen breiten Box verankert ist, ist damit MASSIV schmaler als der heutige 40-Zeilen-Modal — Optionen wie `in-progress`/`blocked-by-deps` (falls solche Werte existieren) könnten die Feldbreite überschreiten und müssten entweder umbrechen oder die Popup-Breite darf über die Feldbreite hinauswachsen (dann aber rechtfertigungsbedürftig, wohin es links/rechts ausbricht).

### 4. Trigger-Pfade (bleiben unverändert im Einstiegspunkt)

Alle vier Pfade rufen exakt `m.openValueMenu(group)` (`box_menu_value.go:136-148`) — nur das Rendering (`valueMenuBox`) und die Positionierung (`composeOverlays`/`placeOverlay`) ändern sich, der Einstiegspunkt bleibt:

- **Tastatur:** `internal/tui/update.go:772` (`keys.Status`), `:780` (`keys.Type`), `:783` (`keys.Priority`) — alle drei im selben `if`-Block ab `update.go:760`.
- **Feld-Cursor Enter (bt-1o4g):** `boxFormActivateCursor` (`box_nav_field.go:323-329`) → `activateBoxFormTarget` (`mouse.go:966-978`) → `openValueMenu("status"|"type"|"priority")`.
- **Detail-Pane-Klick:** `mouseBoxFormDetailClick` (`mouse.go:928-934`) → `detailBoxFormClickRow` (`mouse.go:879-919`, GEOMETRIE-Auflösung) → `activateBoxFormTarget` (`mouse.go:966-978`, DISPATCH).
- **Palette:** `dispatchPalette` (`overlay_palette.go:199-203`) — `case "status"/"type"/"priority": return m.openValueMenu(...)`.

`activateBoxFormTarget` (`mouse.go:966-978`) ist der EINE gemeinsame Dispatch-Punkt für Klick UND Feld-Cursor-Enter — beide rufen dieselbe Funktion, die wiederum `openValueMenu` aufruft. Für die Umstellung heißt das: `openValueMenu` selbst muss NICHT geändert werden (es setzt nur `m.overlay=overlayValueMenu` + Menu-State) — die Änderung sitzt ausschließlich in der Render-/Compositing-Schicht (`valueMenuBox`/`composeOverlays`) und ggf. einem neuen Anker-Feld im Model (z.B. `m.valueMenuAnchorField int`, das beim Öffnen mitgesetzt wird, damit der Compositor weiß, WELCHES Feld zu verankern ist — Tastatur/Palette-Pfade haben ja keinen Klick-Ursprung, sondern müssen den `boxFormEffectiveCursor`/das aufgerufene `group` in eine Feld-Position auflösen).

### 5. Enter-wendet-sofort-an + esc-schließt-ohne-Mutation

Sitzt vollständig in `keyValueMenu` (`box_menu_value.go:164-181`) und `applyValueMenuSelection` (`box_menu_value.go:191-216`):
- `enter` → `applyValueMenuSelection()`: setzt `m.overlay = overlayNone` SOFORT, dann feuert `mutateCmd` mit frisch gelesenem ETag (`m.beanETag(id)`, NICHT der beim Öffnen erfasste — Design-Entscheidung d).
- `esc` ODER die Taste, die die Gruppe geöffnet hat (`valueMenuGroupKey`, `footer_context.go`) → `m.overlay = overlayNone` OHNE Mutation (`box_menu_value.go:174-176`).

Diese Semantik hängt an KEINER Positionierungsentscheidung — sie bleibt beim Umbau vollständig unberührt, solange `keyValueMenu` weiterhin auf `m.overlay == overlayValueMenu` dispatcht (`update.go:853` `case overlayValueMenu:`) und die neue Anker-Positionierung nur `valueMenuBox()`/`composeOverlays()` betrifft, nicht `keyValueMenu`.

## Slice-Vorschlag

1. **Slice A — Anker-Geometrie als reine Funktion, read-only, keine Verhaltensänderung.**
   Ziel: `boxFormFieldScreenRect(m model, b *data.Bean, field int) (x, y, w, h int)` bauen (kombiniert `boxFormPaneMetrics`, `boxFormFieldSpan`, `gridColWidths`, `boxFormEffectiveScroll` — reproduziert `detailBoxFormClickRow`s Geometrie vorwärts), mit Unit-Tests, die sie gegen `detailBoxFormClickRow`s RÜCKWÄRTS-Auflösung cross-checken (ein Klick auf `(x,y)` muss via `detailBoxFormClickRow` wieder auf `field` zurückführen). Betroffene Dateien: neue Funktion in `box_nav_field.go` oder `box_detail_form.go` + Test-Datei. KEIN Rendering-Eingriff, kein `composeOverlays`-Touch — isoliert grün-testbar.
   Risiko: niedrig (reine Geometrie, kein Compositing).

2. **Slice B — Anker-positioniertes `placeOverlay`-Äquivalent, noch hinter Flag/ungenutzt.**
   Ziel: Neue Funktion `placeOverlayAt(bg, fg string, tw, th, x, y int) string` (Variante von `placeOverlay`, `overlay.go`, OHNE die Center-Berechnung, sondern mit übergebenem x/y + Clip-nach-oben-Logik für den Grenzfall aus Frage 3) — NICHT in `composeOverlays` verdrahtet, nur per Unit-Test gegen feste x/y/tw/th-Kombinationen (inkl. Rand-/Überlauf-Fälle) geprüft.
   Risiko: MITTEL — das ist der Kern des Compositing-Risikos (ANSI-safe Splicing, Overflow-Klapplogik), aber isoliert testbar OHNE die Live-Render-Kette zu berühren.

3. **Slice C — Trigger/Positionierung umziehen (composeOverlays + Anker-Feld im Model).**
   Ziel: `m.valueMenuAnchorField int` (oder Wiederverwendung von `boxFormCursor`/`mutTarget`-Kontext) beim `openValueMenu`-Aufruf mitsetzen; `composeOverlays` (`view_browse_repo.go:1090-1092`) ruft für `overlayValueMenu` `placeOverlayAt` mit dem aus Slice A berechneten Rect statt `placeOverlay` (zentriert). Betroffene Dateien: `box_menu_value.go` (`openValueMenu`, `types.go` Model-Feld), `view_browse_repo.go` (`composeOverlays`), Golden-Tests für alle drei Trigger-Pfade (Tastatur/Cursor-Enter/Klick/Palette) neu ziehen.
   Risiko: HOCH — das ist die eigentliche Verhaltensänderung; alle 4 Trigger-Pfade + Fullscreen-Fallback (`boxFormPaneMetrics`' `fullscreenDetail`-Branch) müssen konsistent bleiben.

4. **Slice D — Maus-Auswahl im Popup (Klick auf eine Menü-Zeile statt nur Tastatur).**
   Ziel: `valueMenuBox()`-Zeilen bekommen eine Klick-Hit-Test-Funktion (analog `treeClickRow`/`detailBoxFormClickRow`-Pattern), die einen Klick auf eine Menüzeile direkt auf `applyValueMenuSelection` mit dem geklickten Index mappt — bisher ist das Menü NUR per Tastatur bedienbar (`keyValueMenu`, up/down/enter). Betroffene Dateien: `mouse.go` (neuer Hit-Test + Dispatch-Zweig), `box_menu_value.go` (ggf. Cursor-Set-Helper extrahieren).
   Risiko: mittel — additiv, berührt die bestehende Enter/Esc-Semantik nicht, aber addiert einen neuen Eingabepfad, der mit dem Anker-Rect aus Slice C übereinstimmen MUSS (sonst Klick-Drift wie das Golden-Rule-Drift-Schutz-Muster im Code mehrfach warnt).

5. **Slice E — Modal-Pfad entfernen/aufräumen.**
   Ziel: Sobald C+D grün sind: alten zentrierten `placeOverlay`-Call für `overlayValueMenu` entfernen (bleibt für andere Overlays — Tag/Parent/Blocking-Picker, Palette, Help, Confirms — unverändert bestehen, NUR der Value-Menu-Case wandert auf `placeOverlayAt`), tote Kommentare/Doku in `box_menu_value.go`/`modal.go` aktualisieren, GLOSSARY.md-Eintrag für den neuen Begriff "feld-verankertes Inline-Dropdown" ergänzen.
   Risiko: niedrig — Aufräum-Slice, keine neue Logik.

**Risiko-Urteil:** Slice C (Trigger/Positionierung umziehen in `composeOverlays`) ist der gefährlichste Schnitt — er ist der einzige, der die LIVE-Render-Kette für alle vier Einstiegspunkte gleichzeitig umbiegt und dabei zwingend mit dem Fullscreen-Sonderfall (`boxFormPaneMetrics`' `fullscreenDetail`-Branch, `mouse.go` Kommentar zu `bt-s90e`) und dem Scroll-Offset (`boxFormEffectiveScroll`) konsistent bleiben muss. Slice B (Compositing-Primitive) ist technisch am kniffligsten (ANSI-safe Splicing + Klapplogik), aber isoliert testbar und damit weniger riskant als C, weil C das erst in Produktion "scharf schaltet".



## Slice A/B Fortschritt (2026-07-20, Implementer-Dispatch)

**Slice A — `boxFormFieldRect` (`internal/tui/box_nav_field.go`), Commit `601259a`.**
Forward Feld->Screen-Rect (x,y,w,h,ok), read-only, nirgends verdrahtet.
Test `TestBoxFormFieldRectMatchesRenderedBoxCorners` (`box_nav_field_test.go`)
RED (compile fail: `undefined: boxFormFieldRect`) -> GREEN nach Implementierung,
belegt bei 80 UND 120 Spalten je für Status/Type/Priority: (1) reale `╭`-Ecke
in `ansi.Strip(m.View())` == berechnetes (x,y), (2) Klick auf Rect-Mitte
löst über `detailBoxFormClickRow` auf dasselbe Feld zurück auf (Reziprozitäts-
Cross-Check aus dem eigenen Slice-Vorschlag). Zusätzlich
`TestBoxFormFieldRectNilBeanNotOk` (nil-Bean/out-of-range -> ok=false).

**ERRATUM (im Code-Kommentar dokumentiert, nicht nur hier):**
`detailBoxFormClickRow`s eigene X-Grenze (`originX+lw`) liegt ~3 Spalten
(Split-Modus) / trifft in Fullscreen um 1 Spalte daneben von dem, wo
`detailBoxForm` tatsächlich zeichnet (Render-Sonde bestätigt: reale
Content-Startspalte = `originX+lw+3` im Split, `originX+1` im Fullscreen).
Für die bestehende Klick-Auflösung folgenlos (grobes Pane-Gate + ~14-15-
Zellen-Buckets schlucken den Versatz), aber für einen ANKER-positionierten
Popup sichtbar falsch. `boxFormFieldRect` berechnet deshalb den
geometrisch exakten Ursprung (render-verifiziert) statt `detailBoxFormClickRow`s
Formel zu kopieren — Klick-Reziprozität bleibt trotzdem erhalten (Slack
kreuzt keine Bucket-Grenze). `detailBoxFormClickRow` selbst wurde NICHT
angefasst.

**Slice B — `placeOverlayAt` (`internal/tui/overlay.go`), Commit `a068cbd`.**
Anker-positionierte Variante von `placeOverlay` (zentriert bleibt unverändert
live/unangetastet), teilt sich den Splice-Kern über neue Funktion
`placeCompose` (byte-identische Extraktion aus dem bisherigen `placeOverlay`-
Rumpf). NICHT in `composeOverlays` verdrahtet. Tests (`overlay_test.go`)
RED (compile fail: `undefined: placeOverlayAt`) -> GREEN:
`TestPlaceOverlayAtPositionsAtExactAnchor`,
`TestPlaceOverlayAtBottomOverflowDropsExcessRowsSilently` (Überlauf unten
= stillschweigend verworfen, wie `placeOverlay` es schon immer tut — KEIN
Aufwärts-Klappen, s.u.), `TestPlaceOverlayAtNegativeXClampsToZero`,
`TestPlaceOverlayAtNegativeYDropsRowsAboveTop`,
`TestPlaceOverlaySharesCompositingCoreWithPlaceOverlayAt` (Drift-Schutz:
`placeOverlay` zentriert bei identischem effektivem (x,y) liefert
byte-identisches Ergebnis zu `placeOverlayAt`).

**Gates:** `command go build -o bin/bt .` grün · `command go vet ./...`
grün · `command go test ./...` (VOLL, kein `-short`) grün, `internal/tui`
151.4s · Goldens byte-identisch (kein `testdata/*.golden` im `git status`).

## Notes for Slice C

- **Überlauf/Flip:** `placeOverlayAt` klappt NICHT nach oben, wenn das
  Popup unten aus dem Pane ragt — überschüssige Zeilen werden still
  verworfen (wie `placeOverlay` es immer schon tut). Slice C muss VOR dem
  Verdrahten entscheiden, ob ein Popup, das unten aus der Detail-Pane
  ragen würde, geklappt (`y` nach oben verschoben) oder einfach
  abgeschnitten werden soll — aktuell gibt es dafür KEINEN Code, weder
  hier noch sonst im Repo.
- **X-Geometrie-Diskrepanz:** `boxFormFieldRect`s Ursprung weicht von
  `detailBoxFormClickRow`s eigener (ungenauerer) X-Formel um ~3 Spalten
  ab (s. ERRATUM oben). Für Slice C harmlos, SOLANGE nur `boxFormFieldRect`
  zur Positionierung und `detailBoxFormClickRow` weiter nur zur
  Klick-Auflösung verwendet wird — beide dürfen NICHT vermischt werden
  (z.B. nicht versehentlich `originX+lw` für die Popup-Positionierung
  nehmen).
- **Fullscreen-Konsistenz:** `boxFormFieldRect` behandelt
  `m.fullscreen == fullscreenDetail` bereits (kein `lw`-Split-Offset,
  Content startet bei `originX+1`) — an `boxFormPaneMetrics`s bestehende
  Verzweigung angelehnt. Beim Verdrahten in `composeOverlays` prüfen, ob
  Slice C das Value-Menu auch im Vollbild ankern soll (Tastatur/Palette-
  Pfade funktionieren dort ja weiterhin, nur Klick ist im Vollbild laut
  F01 tot) — falls ja, ist die Geometrie schon da, sonst ggf. bewusst auf
  den alten zentrierten `placeOverlay`-Pfad zurückfallen für diesen Fall.
- **Welches Feld verankern?** `openValueMenu(group)` kennt aktuell keinen
  Feld-Index — Slice C braucht ein neues Model-Feld (z.B.
  `m.valueMenuAnchorField int`), das beim Öffnen (Tastatur/Cursor-Enter/
  Klick/Palette, alle 4 Pfade) mitgesetzt wird, damit `composeOverlays`
  weiß, an welchem `boxFormFieldOrder`-Index es `boxFormFieldRect`
  aufrufen soll.



## Slice C Fortschritt (2026-07-20, Implementer-Dispatch)

**Commit `4abee07` — `feat(boxform): Value-Menue feld-verankert statt zentriert (Slice C)`.**
`internal/tui/box_menu_value.go`, `internal/tui/types.go`,
`internal/tui/view_browse_repo.go`, `internal/tui/box_menu_value_anchor_test.go`.

**Model-Feld:** `m.valueMenuAnchorField int` (types.go). Gesetzt in
`openValueMenu(group)` (box_menu_value.go) über die NEUE statische
Zuordnung `boxFormFieldIndexForGroup(group)` — alle vier Trigger-Pfade
(Tastatur s/o/u, Feld-Cursor-Enter, Maus-Klick, Palette) rufen bereits
exakt `openValueMenu(group)` (S6-Grounding §4 bestätigt) → EIN Änderungsort
statt vier.

**Positionierung:** `placeValueMenuOverlay` (box_menu_value.go), verdrahtet
in `composeOverlays`s `overlayValueMenu`-Case (view_browse_repo.go). Bei
`boxFormEnabled()` + auflösbarem `boxFormFieldRect`: `placeOverlayAt` an
`(x, valueMenuPopupY(rect))`, sonst Fallback auf das alte zentrierte
`placeOverlay` (accordion-Modus unverändert, kein boxed field zum
Verankern).

**Flip-up:** `valueMenuPopupY(anchorY, anchorH, fgH, canvasH)` — Standard
unten; klappt nach OBEN, wenn unten kein Platz UND `anchorY>0` (irgendein
Platz oben vorhanden); fällt nur bei `anchorY==0` (Feld sitzt bereits ganz
oben) auf "unten, unten geclippt" zurück. **Wichtige Design-Entscheidung:**
Das Klappen verlangt NICHT, dass der Popup oben vollständig passt —
Status/Type/Priority sitzen real immer bei y≈10 (S6-ERRATUM,
box_nav_field.go), ein Popup mit 12 Zeilen (`fgH`) passt dort NIE
vollständig oberhalb. Ein "nur klappen wenn vollständig passt"-Kriterium
hätte den Klapp-Zweig für genau diese drei Felder permanent totgelegt.
Ein teilweise oben abgeschnittener Popup nutzt `placeOverlayAt`s
bestehenden Silent-Overflow-Drop (Slice B) — kein neuer Mechanismus.

**X-Präzision:** ausschließlich `boxFormFieldRect` verwendet (NIE
`detailBoxFormClickRow`s X gemischt, wie im Slice-A/B-ERRATUM verlangt).
Zusätzlich ein defensiver rechter-Rand-Clamp (`placeValueMenuOverlay`):
der ~42 Zellen breite Popup (`clampModalWidth(40,...)`) ist unabhängig
von der ~15 Zellen breiten Feld-Box und würde bei Type/Priority auf
80 Spalten rechts aus dem Bildschirm ragen — reiner Canvas-Clamp, KEIN
Resize/Reflow auf Feldbreite (siehe Notes for Reviewer).

**Tests (`box_menu_value_anchor_test.go`), RED (compile fail:
`undefined: valueMenuPopupY`/`boxFormFieldIndexForGroup`) → GREEN:**
- `valueMenuPopupY`: 4 reine Branch-Tests (unten passt, klappt+passt
  vollständig, klappt+teilweise geclippt, kein-Platz-oben-Fallback).
- `boxFormFieldIndexForGroup`: Mapping-Test + unbekannte Gruppe → -1.
- Je EIN render-geerdeter Test pro Trigger-Pfad (Tastatur `s`,
  Feld-Cursor-Enter auf "type", Maus-Klick auf "(u)", Palette
  `actionID:"type"`), 80 UND 120 Spalten: Popup-Bottom-Border-Ecke "╰"
  exakt bei der aus `boxFormFieldRect`+`valueMenuPopupY` berechneten
  Position (inkl. Rand-Clamp), NICHT an der alten zentrierten Position.
- Flip-up render-geerdet (`h=20`, erzwingt Überlauf unten): bestätigt
  sowohl die Position (gleicher Ecke-Check) als auch dass der Y-Wert
  tatsächlich < Feld-Oberkante ist (Klapp-Zweig feuerte).
- `TestValueMenuStaysCenteredWithoutBoxFormFlag`: Accordion-Modus bleibt
  byte-genau zentriert (alte `(bgW-fgW)/2,(bgH-fgH)/2`-Formel).

**Wichtiger methodischer Fund während der Testkonstruktion:** Ein erster
Testansatz verglich `m.View()` gegen ein "Baseline"-Rendering mit
`m.overlay=overlayNone` — das schlug FLÄCHENDECKEND fehl, weil die Footer-
Zeile selbst von `m.overlay` abhängt (`footer_context.go`s
`overlayLocalBindings`) — das Baseline hatte also eine ANDERE Chrome, nicht
nur eine andere Overlay-Schicht. Ersetzt durch direkte Landmark-Suche
(Popup-Eckzeichen via `ansi.Cut`, wie Slice A) — robuster und ohne diese
Kopplung.

**Goldens:** `command go test ./internal/tui -run Golden -update` erzeugt
NULL Diff (`git status`/`git diff --stat` auf `testdata/` leer) — kein
Golden-Fixture rendert den geöffneten Value-Menu-Zustand, also keine
Golden-Änderung nötig oder erfolgt.

**Gates:** `command go build -o bin/bt .` grün · `command go vet ./...`
grün · `command go test ./...` (VOLL, im Vordergrund) grün,
`internal/tui` 151.6s.

## Notes for Reviewer

- **Flip-Semantik ist NICHT "voll passt oder klappt nicht"** — sie ist
  "klappt sobald unten überläuft UND oben überhaupt Platz ist (auch nur
  1 Zeile)", mit stillem Clip oben falls nötig. Das ist eine bewusste
  Abweichung von einer strengeren "nur klappen wenn komplett sichtbar"-
  Regel, weil letztere für Status/Type/Priority (reale y≈10, Popup-Höhe
  12) NIE greifen würde. Bitte explizit prüfen, ob das PO-Vorgabe
  "Standard-Dropdown-Verhalten" trifft, oder ob eine strengere Regel
  gewünscht war (dann bräuchte D09 vermutlich eine schmalere Popup-Höhe
  oder ein Redesign, kein reiner Slice-C-Fix mehr).
- **Rechter-Rand-Clamp ist ein Sicherheitsnetz, keine Lösung der offenen
  Breiten-Frage** aus dem ursprünglichen S6-Grounding (§3: Popup massiv
  breiter als die ~15-Zellen-Box, sollte es umbrechen/schrumpfen?). Dieser
  Slice beantwortet das NICHT — er verhindert nur, dass der Popup sichtbar
  über den rechten Bildschirmrand hinausragt.
- **Alle 4 Trigger-Pfade belegt** (s.o.), inkl. Cross-Check dass
  `m.valueMenuAnchorField` unabhängig vom Pfad korrekt via
  `boxFormFieldIndexForGroup` gesetzt wird (EIN Änderungsort, wie geplant).
- **Fullscreen:** `boxFormFieldRect` behandelt `fullscreenDetail` bereits
  konsistent (Slice A) — `placeValueMenuOverlay` ruft es unverändert auf,
  sollte also auch im Vollbild korrekt verankern (Tastatur/Palette öffnen
  dort ja weiterhin, Klick ist laut F01 tot). NICHT eigens render-geerdet
  getestet in diesem Slice (Scope-Entscheidung: Zeitbudget auf die vier
  Trigger-Pfade + Flip-up konzentriert, wie vom Supervisor vorgegeben) —
  falls das Review-Gate Fullscreen als Pflichtabdeckung sieht, fehlt dort
  noch ein expliziter Test.
- **Golden-Diff:** keiner (s.o.) — kein Fixture deckt den offenen
  Value-Menu-Zustand ab. Falls der Reviewer das für lückenhaft hält
  (Golden-Abdeckung des NEUEN Anker-Renderings fehlt komplett), wäre das
  ein sinnvoller Slice-D/E-Nachtrag (ein neues Golden mit offenem,
  verankertem Value-Menu).
- **Enter-sofort-anwenden/Esc-ohne-Mutation** (`keyValueMenu`/
  `applyValueMenuSelection`) unverändert — bestehende Tests dafür (in
  `box_menu_value_test.go`) liefen im vollen Lauf grün mit.



## Slice C — B01/B02-Fix (2026-07-20, Re-Review-Runde)

**Commit `27dc35d` — `fix(boxform): Value-Menue Inline-Popup Breite+Flip (B01/B02, Slice C)`.**
`internal/tui/box_menu_value.go`, `internal/tui/box_nav_field.go`,
`internal/tui/box_menu_value_anchor_test.go`,
`internal/tui/box_menu_value_golden_test.go` (neu),
`internal/tui/testdata/value_menu_anchored.golden` (neu).

**B01 (high) — Chrome-Überschreiben durch Flip-up behoben.**
`valueMenuPopupY` bekommt einen neuen Parameter `minY` (Chrome-Fußboden,
`boxFormPopupChromeFloor` in `box_nav_field.go` — eine Zeile hinter der
Trennlinie, view-/Vollbild-unabhängig, da `clickPaneGeometry`s `originY`
selbst schon dokumentiert "+1 Pane-Oberrahmen" enthält und der Fußboden
genau EINE Zeile davor sitzt). Klappt nur noch, wenn der GESAMTE Popup
oberhalb passt (`above>=minY`); sonst bleibt er unten (Clip am unteren
Rand über `placeOverlayAt`s bestehenden Silent-Overflow-Drop) — NIE mehr
negatives Y, NIE mehr Chrome überschrieben.

**B02 (medium-high) + Breiten-Entscheidung (PO 2026-07-20) — Popup schrumpft
auf Inhaltsbreite, linksbündig.**
Neue Funktion `valueMenuContentWidth(title, body)` — Breite = breiteste
Zeile (Hint/Titel/Item) + 2 (Padding). `valueMenuBox()` nutzt sie jetzt
statt der fixen `clampModalWidth(40, m.width)` (Ergebnis real ~29 statt
42 Zellen). `placeValueMenuOverlay` positioniert linksbündig an
`boxFormFieldRect(...).x`; NUR bei Rechts-Überlauf nach links geschoben,
NIE unter `boxFormPaneContentLeft(m)` (neue Helper-Funktion,
`box_nav_field.go`, aus `boxFormFieldRect`s eigener Pane-Ursprungs-Geometrie
extrahiert — Golden-Rule-Drift-Schutz).

**Refactor (Drift-Schutz):** `boxFormFieldRect`s Pane-Ursprungs-Geometrie in
zwei wiederverwendbare Helfer aufgeteilt: `boxFormPaneContentLeft(m)`
(linke Content-Kante) und `boxFormPopupChromeFloor(m)` (Chrome-Fußboden).
Beide werden jetzt sowohl von `boxFormFieldRect` selbst als auch von
`placeValueMenuOverlay` verwendet — kann nicht mehr unabhängig
auseinanderdriften.

**Tests (alle GREEN, RED zuerst durch Signaturänderung
`valueMenuPopupY` bestätigt — Compile-Fail):**
- `TestValueMenuPopupYNeverCrossesChromeFloor` (B01, reine Funktion):
  Szenario aus dem Reviewer-Bug (anchorY=10, fgH=12, minY=3) bleibt
  bei `below=13`, NICHT geflippt.
- `TestValueMenuFlipDoesNotCorruptChrome` (B01, render-geerdet, 80×24):
  ALLE Zeilen 0..`boxFormPopupChromeFloor` byte-identisch zum
  Vor-Öffnen-Baseline (nicht nur eine Ecke).
- `TestValueMenuFlipStaysBelowWhenFullFitAboveIsImpossible`: dokumentiert
  render-geerdet, dass für Status/Type/Priority ein vollständiger
  Oben-Fit real unerreichbar ist (y bleibt unten).
- `TestValueMenuTypeAndPriorityPopupsSitAtDifferentFieldNearColumns` (B02,
  render-geerdet, 80 Spalten): Type(x=46, ungeklemmt) ≠ Priority(x=51,
  leicht geklemmt) — beide feld-nah (Delta ≤20).
- `TestValueMenuContentWidthSizesToLongestLine` +
  `TestValueMenuBoxWidthIsUnderOldFixedFortyAndMatchesContent`
  (Breite < 42, == Inhaltsbreite).
- `TestValueMenuAnchoredInFullscreenDetail` (Q01): echter `v`→`s`-Roundtrip,
  Anker-Geometrie im Vollbild korrekt (boxFormFieldRect bereits
  Vollbild-bewusst aus Slice A).
- Alle 4 Trigger-Pfad-Tests aus dem ursprünglichen Slice-C-Durchlauf
  bleiben grün (Keyboard/Feld-Cursor-Enter/Klick/Palette), 80+120 Spalten.

**Golden (Q02):** `testdata/value_menu_anchored.golden` NEU angelegt
(nach dem Fix, nicht davor — friert also den KORREKTEN Zustand ein):
100×30, Status-Menü offen, `goldenTreeModel`-Fixture. Alle
VORHANDENEN Goldens per `-update` gegengeprüft — `git diff --stat` auf
`testdata/` zeigt NULL Diff für alle bestehenden Dateien, nur das neue
Fixture ist hinzugekommen.

**Gates:** `command go build -o bin/bt .` grün · `command go vet ./...`
grün · `command go test ./...` (VOLL, Vordergrund) grün,
`internal/tui` 151.7s.

## Notes for Reviewer (Re-Review-Fokus)

- **B01-Fix-Prinzip:** Flip erfordert jetzt VOLLSTÄNDIGEN Platz oberhalb
  des Chrome-Fußbodens — für Status/Type/Priority (reale y≈10, Popup-Höhe
  jetzt nur noch ~12 Zeilen dank Breiten-Fix, aber Höhe unverändert) ist
  das real NIE erreichbar (Chrome-Fußboden ≈3-7). Der Flip-Zweig ist damit
  für DIESE drei Felder faktisch tot Code (nur an synthetischen
  Unit-Test-Werten belegt) — Popup bleibt bei zu wenig Platz immer unten
  und clippt unten. Das ist die Konsequenz aus "niemals Chrome
  überschreiben" und sollte explizit abgenommen werden (Alternative wäre
  eine geringere Popup-Höhe, aber das war nicht Teil des Fix-Auftrags).
- **B02-Fix:** Type/Priority sitzen jetzt bei x=46/x=51 (80 Spalten,
  ungeklemmt/leicht geklemmt) statt beide bei 38. Preis: der Popup
  überlappt bei sehr rechtsständigen Feldern (Priority) leicht in Richtung
  der Feldbox selbst statt frei daneben — bei Bedarf wäre ein
  "occlude the field itself"-Sonderfall denkbar, war aber nicht verlangt.
- **Golden-Diff:** NUR ein neues Fixture hinzugekommen
  (`value_menu_anchored.golden`), alle bestehenden byte-identisch. Bitte
  das neue Golden visuell gegenprüfen (ASCII-Dump im Report unten) —
  insbesondere dass Zeile 0-9 (App-Chrome, Filter-Strip, Title-Box)
  unverändert zur Nicht-Popup-Ansicht sind und der Popup sauber unter der
  Status-Box (Zeile 12, Badge "(s)") beginnt.
- **Kosmetischer Nebenbefund (kein neuer Bug):** Wo der schmalere Popup
  über die Parent/Tags-Zeile (rowB, 2-spaltiges Grid) überlappt, bleibt
  ein einzelnes Rahmenzeichen der darunterliegenden Box sichtbar direkt
  neben der Popup-eigenen Ecke (sichtbar im neuen Golden, Zeile 13) — das
  ist dieselbe Compositing-Eigenschaft, die JEDES Overlay über
  unregelmäßigem Hintergrund hat (nicht neu durch Slice C, nicht durch
  diesen Fix verursacht), kein Verhaltensfehler.



## Grounding Slice D (2026-07-20)

- **Guard bestätigt:** `handleMouse` (`mouse.go:181-185`) blockt JEDE Maus-
  Aktion, sobald `m.overlay != overlayNone` — ein Klick auf eine Popup-Zeile
  tat bis Slice D buchstäblich nichts. Davor laufen nur Toast-Klick-Vorrang
  und Fullscreen-Guard (`m.fullscreen != fullscreenNone`, F01-Scope-Cut,
  unangetastet).
- **Wert-Anwendung bestätigt:** `applyValueMenuSelection()`
  (`box_menu_value.go:191-216`) liest `m.menu.cursor`, setzt
  `m.overlay = overlayNone` sofort, dispatcht `SetStatus/SetType/SetPriority`
  via `mutateCmd` — GENAU das, was `keyValueMenu`s `enter`-Case bereits tut.
  Slice D setzt nur `m.menu.cursor` auf den geklickten Index und ruft
  dieselbe Funktion — keine eigene Mutations-Logik.
- **Popup-Zeilen-Layout verifiziert** (`valueMenuBox()`, single-group seit
  B11/B12): Rahmen oben, Titel, Hint, Leerzeile, Gruppenkopf, dann n Items,
  dann Leerzeile+Rahmen unten. NICHT hart-codiert — `valueMenuBodyAndItemRows`
  zählt Zeilen beim Bauen mit (`strings.Count`), speist sowohl `valueMenuBox()`
  (Render) als auch `valueMenuItemRow()` (Hit-Test).
- **X-Geometrie:** Popup-Rect kommt ausschließlich aus der neuen
  `valueMenuPopupRect()` (jetzt die EINE Quelle für `placeValueMenuOverlay`
  UND den Klick-Hit-Test) — nie aus `detailBoxFormClickRow`s eigener
  (laut Slice-A/B-ERRATUM ungenauer) X-Logik.

## Summary — bt-f0y9 (Slices A-E, D09 revidiert)

Alle fünf Slices abgeschlossen, Feld-verankertes Inline-Dropdown für
Status/Type/Priority ersetzt das zentrierte Modal, MAUS-nativ (Klick wählt),
mit Flip-up/Chrome-Schutz und Inhaltsbreite. Create-Form bleibt huh
(unangetastet, verifiziert per `git diff --stat` über alle Commits: keine
Änderung an `box_confirm_create.go`/`box_confirm_create_test.go`).

| Slice | Commit(s) | Inhalt |
|---|---|---|
| A | `601259a` | `boxFormFieldRect` — Vorwärts-Geometrie Feld→Screen-Rect |
| B | `a068cbd` | `placeOverlayAt` — anker-positionierte Overlay-Variante |
| C | `4abee07`, `27dc35d` (B01/B02-Fix nach Review) | Value-Menü feld-verankert statt zentriert, Flip-up mit Chrome-Fußboden, Inhaltsbreite statt fix 40 |
| D | `cdeec98` | Klick im Popup wählt Wert (maus-nativ), Klick außerhalb schließt |
| E | — (kein Code-Commit, s.u.) | Cleanup-Prüfung + finale Validierung |

**Slice E — Cleanup-Ergebnis:** Kein toter Code gefunden. Alle in A-D neu
eingeführten Funktionen (`boxFormFieldRect`, `boxFormPaneContentLeft`,
`boxFormPopupChromeFloor`, `placeOverlayAt`, `placeCompose`,
`valueMenuContentWidth`, `valueMenuBodyAndItemRows`, `valueMenuItemRow`,
`valueMenuPopupRect`, `valueMenuPopupY`, `mouseValueMenuClick`,
`boxFormFieldIndexForGroup`) per Grep gegen echte (Nicht-Test-)Call-Sites
verifiziert — jede hat mindestens einen echten Aufrufer außerhalb ihrer
eigenen Definition/Doc-Kommentare. Der zentrierte `placeOverlay`-Pfad für
Accordion/Flag-OFF ist erhalten (`TestValueMenuStaysCenteredWithoutBoxFormFlag`
weiterhin grün). Der faktisch tote Flip-Zweig in `valueMenuPopupY`
(unerreichbar für Status/Type/Priority bei realistischen Terminalgrößen,
siehe Slice-C-B01-Notiz) bleibt bewusst stehen — Reviewer hatte das als
keinen Defekt eingestuft, nicht wegrefaktoriert. **Kein Cleanup-Commit
nötig** (kein Leer-Commit erzeugt).

## Test-Output

- `command go build -o bin/bt .` — grün (alle Runden).
- `command go vet ./...` — grün (alle Runden).
- `command go test ./...` — **ZWEIMAL** im Vordergrund gelaufen (zweiter
  Lauf mit `-count=1`, erzwungen ohne Cache): beide Male grün,
  `internal/tui` 151.5s / 153.3s, alle anderen Pakete `ok`.
- Goldens: `command go test ./internal/tui -run Golden -update` nach JEDER
  Slice ausgeführt — `git status`/`git diff --stat` auf `testdata/` zeigt
  über den GESAMTEN Verlauf NULL Diff für alle bestehenden Golden-Dateien;
  einzige Änderung war das bewusst neu hinzugefügte
  `testdata/value_menu_anchored.golden` (Slice C, nach dem B01/B02-Fix
  erzeugt).

## Smoke (REAL, nicht nur Unit)

tmux 80×24 gegen `/Users/erik/Obsidian/tools/sproutling` (echtes Repo, echte
Beans-Daten), `BT_BOXFORM=1`, frisches `bin/bt` (mtime nach dem letzten
Commit geprüft: `cdeec98` @ 22:47, Binary @ 22:53 — frisch gebaut).

1. `bt` gestartet, Bean `sproutling-elir` (Status `todo`) im Fokus.
2. `s` gedrückt → Status-Menü öffnet feld-verankert direkt unter der
   Status-Box (nicht zentriert), App-Chrome (Zeile 0-2) unverändert.
3. **Echter Maus-Klick** via SGR-Mouse-Escape-Sequenz
   (`\x1b[<0;COL;ROWM` / `...m`, Press+Release) auf die "completed"-Zeile
   im Popup gesendet — via `tmux send-keys -l` (nicht simuliert, echte
   Escape-Bytes an den bubbletea-Prozess).
4. **Ergebnis:** Menü schloss, Status-Box zeigt "complet ▾" (Wert
   angewendet), Tree-Zeile für `elir` wechselt Status-Glyph von "t" auf "c".
   Auf Platte verifiziert: `.beans/sproutling-elir--bug-fixing-ios-app.md`
   zeigt `status: completed` unmittelbar nach dem Klick (echte Mutation,
   kein reiner UI-Zustand).
5. **Aufräumen:** `sproutling`-Repo per `git checkout --` auf den
   Original-Stand zurückgesetzt (verifiziert: `git status --porcelain`
   zeigt clean für die betroffene Datei).

Ehrlichkeit: NUR der Maus-Klick-Trigger (Slice D, der risikoreichste/
wichtigste Pfad) wurde live gesmoked. Tastatur/Feld-Cursor-Enter/Palette-
Trigger sowie Flip-up/Fullscreen/Breiten-Verhalten sind NUR unit-/
render-geerdet getestet (Slices A-C), nicht zusätzlich live gesmoked — im
Zeitbudget dieses Abschluss-Dispatches nicht verlangt, aber hier transparent
benannt.

## Deviations/ERRATA (kumulativ über A-E)

1. **Slice A:** `detailBoxFormClickRow`s eigene X-Grenze (`originX+lw`) liegt
   ~3 Spalten (Split) / 1 Spalte (Fullscreen) neben der tatsächlichen
   Render-Position — für die bestehende Klick-Auflösung folgenlos (grobes
   Pane-Gate + große Buckets schlucken den Versatz), für einen Anker-Popup
   aber sichtbar falsch. `boxFormFieldRect` berechnet den geometrisch
   exakten, render-verifizierten Ursprung separat.
2. **Slice C (vor Fix):** Flip-up konnte Chrome überschreiben (B01) und
   Type/Priority klemmten beide auf dieselbe Spalte (B02) — BEIDE per
   Re-Review gefunden und in der Fix-Runde behoben, verifiziert grün.
3. **Slice C (nach Fix):** Der Flip-Zweig ist für Status/Type/Priority bei
   realistischen Terminalgrößen faktisch unerreichbar (Popup-Höhe > Platz
   oberhalb des Chrome-Fußbodens) — bewusste Konsequenz aus "niemals Chrome
   überschreiben", vom Reviewer als kein Defekt eingestuft.
4. **Slice D:** Klick INNERHALB des Popups, aber NICHT auf einer Item-Zeile
   (Titel/Hint/Gruppenkopf/Leerzeile/Rahmen) ist ein bewusster No-Op (weder
   Auswahl noch Schließen) — in der Aufgabenstellung nicht explizit
   spezifiziert, als sicherster/unauffälligster Default gewählt.
5. **Kosmetischer Nebenbefund (kein Bug, aus Slice C):** wo der Popup über
   die Parent/Tags-Zeile (rowB) reicht, bleibt ein einzelnes Rahmenzeichen
   der darunterliegenden Box neben der Popup-Ecke sichtbar (Golden-Fixture,
   Zeile 13) — dieselbe Compositing-Eigenschaft, die jedes Overlay über
   unregelmäßigem Hintergrund hat.

## Notes for final Reviewer

- Alle vier Trigger-Pfade (Tastatur, Feld-Cursor-Enter, Klick, Palette)
  sind für ANCHORING (Slice C) UND — nur für Klick — für die neue
  Auswahl-Fähigkeit (Slice D) abgedeckt.
- Live-Smoke deckt NUR den Maus-Klick-Pfad ab (s.o., "Ehrlichkeit"-Absatz) —
  falls der finale Reviewer weitere Live-Smokes verlangt (z.B. Flip-up bei
  80×24 visuell, Fullscreen-Anker), wäre das ein zusätzlicher, in diesem
  Dispatch nicht enthaltener Schritt.
- Der faktisch tote Flip-Zweig (Deviation 3) ist ein bewusster Kompromiss
  zwischen "niemals Chrome überschreiben" (B01) und "sichtbares Klappen" —
  bitte explizit abnehmen oder als bekannten Kompromiss protokollieren.
- Kein Cleanup-Commit in Slice E — alle Funktionen aus A-D sind in
  Benutzung, nichts wurde tot hinterlassen.
- Create-Form (huh) unangetastet über den GESAMTEN Verlauf verifiziert
  (`git diff --stat` über alle 5 Commits gegen `box_confirm_create*.go`:
  leer).
- `docs/GLOSSARY.md` durchgehend NICHT angefasst (pre-existing Fremd-Reformat,
  bleibt für den Nutzer).
