---
# bt-6nuz
title: 'Relations-Overlay: Titel, Rahmen, Footer-Styling, r-Keybind'
status: todo
type: task
priority: high
created_at: 2026-07-20T09:22:58Z
updated_at: 2026-07-20T09:22:58Z
parent: bt-vy1q
---

PO-Befunde aus dem Live-Test 2026-07-20 (#6-#9). Vier Punkte, alle am selben Overlay —
zusammen erledigen.

## #6 — Keybind `r` fehlt im Footer
Der Key ist aktiv, wird aber nicht angezeigt. Vermutlich Nachwirkung von bt-fy5d
(Footer entschlacken): `r` wurde dort entfernt, weil es zur inline gebadgten Gruppe
gezaehlt wurde — hat aber KEIN Inline-Badge (das Relations-Panel hat keinen eigenen
Rahmen-Hotkey wie Status/Type/Tags). Also faelschlich mit ausgeblendet.
Pruefen: `internal/tui/footer_context.go`, die Auswahl in bt-fy5d (Commit `d81e583`).

## #7 — Overlay-Titel heisst "Blocking", muss "Relations" heissen
Konsistenz zur Detail-View-Ueberschrift. `internal/tui/box_picker_blocking.go`,
Render-Funktion `blockingPickerBox()`. Nur das ANZEIGE-Label, nicht die internen
Bezeichner umbenennen (blockItems/blockPending etc. bleiben).

## #8 — Suchfeld im Overlay ist ungerahmt
Soll als Box dargestellt werden, wie die uebrigen Felder. Primitiv vorhanden:
`dropdownBox`/`boxTopBorder`/`boxBottomBorder` (box_dropdown.go, box_detail_form.go).
Der Filter-Strip darueber nutzt sie bereits — das Suchfeld faellt optisch heraus.

## #9 — Footer-Keys im Overlay nicht konsistent gestylt
Sollen wie im Main-View unten stehen: **Key in Teal, Aktion in Subtext**. Keys, die
bereits am Feld im Rahmen stehen, NICHT zusaetzlich als Tooltip auffuehren (gleiche
Regel wie bt-fy5d im Haupt-Footer).
Styling-Quelle: `renderBindings()` in `internal/tui/view.go`. Theme-Token nur aus
`internal/theme/`.

## Kontext
Die Footer-Labels werden seit bt-z4w7 (Commit `3e3ddd9`) aus der TATSAECHLICH aktiven
Bindung abgeleitet statt hart verdrahtet — `valueMenuGroupKey`, `blockingPickerToggleHint`.
Neue Labels dort einhaengen, nicht danebenbauen.

## Akzeptanz
- [ ] `r` erscheint wieder im Browse-Footer
- [ ] Overlay-Titel lautet "Relations"
- [ ] Suchfeld gerahmt, optisch konsistent zum Filter-Strip
- [ ] Footer-Keys Teal/Subtext, keine Dopplung mit Inline-Badges
- [ ] tmux-Smoke bei 80 Spalten (Footer-Aenderung = Projektregel)
- [ ] voller Testlauf gruen
