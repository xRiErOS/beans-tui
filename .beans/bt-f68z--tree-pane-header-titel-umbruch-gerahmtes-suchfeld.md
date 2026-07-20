---
# bt-f68z
title: 'Tree-Pane: Header, Titel-Umbruch, gerahmtes Suchfeld'
status: todo
type: feature
priority: high
created_at: 2026-07-20T09:22:58Z
updated_at: 2026-07-20T09:22:58Z
parent: bt-vy1q
---

PO-Befunde aus dem Live-Test 2026-07-20 (#11-#13). Drei Punkte an der linken Pane —
zusammen erledigen, sie teilen sich Renderer und Golden.

## #11 — Suchfeld zu stark gemutet
`⌕ / search` verschwindet optisch. Soll ebenfalls **boxed** dargestellt werden, wie die
uebrigen Felder. Primitive: `dropdownBox`/`boxTopBorder`/`boxBottomBorder`.

## #12 — Tree-View hat keine Spaltenueberschriften
- **Master-Detail (schmal):** Header auf die Keybindings kuerzen — `S` = Status, `Title`.
- **Vollbild (`v`, breit):** Header ausschreiben.
Also breitenabhaengig, nicht zwei getrennte Implementierungen.

## #13 — Titel brechen nicht um, sondern werden abgeschnitten
Gewuenscht ist Umbruch mit **haengendem Einzug**, ausgerichtet auf den Titelbeginn:

```
  i B eq67 God-Delete/Purge crasht: ORM nullt
           child-FK statt Cascade (alert_configs
           NOT NULL)
▶ t M elir Bug Fixing - iOS-App
▶ t M 4x21 Canary-Instanz sproutling-test —
           Staged-Rollout & Dogfood auf NAS
```

**Achtung, Folgewirkung:** Zeilen sind dann nicht mehr 1:1 zu beans. Das betrifft
- `treeClickRow` / `flatClickRow` (Maus-Trefferzonen) — muessen Zeilen auf beans abbilden
- `windowStart`/Scroll-Rechnung — zaehlt Zeilen, nicht Elemente
- Cursor-Bewegung up/down — soll ueber beans springen, nicht ueber Zeilen

Praezedenzfall im Repo: `parentPickerRowBudget` hatte exakt diesen Fehler (Elemente statt
Zeilen gezaehlt), gefunden im 80x24-Smoke von bt-a3a8. Nicht wiederholen.

## Betroffen
`internal/tui/view_browse_repo.go` (`treeRowText`, Tree-Render), `view_browse_backlog.go`
(`backlogRowText`), `view_browse_flat.go`, `mouse.go`. Golden: `tree`, `backlog`,
`browse_flat`, `browse_boxform`, `chrome`.

## Akzeptanz
- [ ] Suchfeld gerahmt
- [ ] Header im Split gekuerzt, im Vollbild ausgeschrieben
- [ ] Titel brechen mit haengendem Einzug um
- [ ] Klick trifft nach Umbruch weiterhin das richtige bean (Test mit umbrechenden Titeln)
- [ ] up/down springen ueber beans, nicht ueber Zeilen
- [ ] tmux-Smoke bei 80 Spalten
- [ ] voller Testlauf gruen
