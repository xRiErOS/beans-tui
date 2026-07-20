---
# bt-f0y9
title: 'Endliche Felder: feld-verankertes Inline-Dropdown (D09 revidiert)'
status: todo
type: task
created_at: 2026-07-20T18:05:08Z
updated_at: 2026-07-20T18:05:08Z
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
