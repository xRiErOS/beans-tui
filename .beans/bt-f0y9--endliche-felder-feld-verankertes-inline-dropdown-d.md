---
# bt-f0y9
title: 'Endliche Felder: feld-verankertes Inline-Dropdown (D09 revidiert)'
status: in-progress
type: task
priority: normal
created_at: 2026-07-20T18:05:08Z
updated_at: 2026-07-20T19:01:57Z
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
