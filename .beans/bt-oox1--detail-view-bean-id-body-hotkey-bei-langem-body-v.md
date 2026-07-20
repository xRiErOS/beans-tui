---
# bt-oox1
title: 'Detail-View: bean-ID, Body-Hotkey bei langem Body, v im Footer'
status: todo
type: task
priority: normal
created_at: 2026-07-20T09:23:37Z
updated_at: 2026-07-20T09:23:37Z
parent: bt-vy1q
---

Drei kleine PO-Befunde aus dem Live-Test 2026-07-20 — unabhaengig voneinander, aber je zu
klein fuer ein eigenes bean.

## #1 — bean-ID fehlt im Detail-View
Die ID soll **schlank neben dem Titel** stehen. Aktuell zeigt die Detail-Pane sie gar nicht
(links im Master-Detail).
**Zusammenhang beachten:** bt-pl5p hat den Projekt-Slug aus den Listen-IDs entfernt
(`sproutling-eq67` -> `eq67`), die Detail-Pane traegt aber bewusst die VOLLE, kopierbare ID
— das war eine ausdrueckliche Entscheidung und ist per Gegen-Test abgesichert. Beim
Umsetzen also klaeren: volle ID (kopierbar, konsistent zur bisherigen Entscheidung) oder
kurze (platzsparend)? Empfehlung: volle ID, sie ist der Kopier-Anker.

## #4 — Bei langem Body ist der Edit-Hotkey nicht sichtbar
Der Hotkey `(e)` sitzt im UNTEREN Rahmen des Body-Panels. Ist der Body laenger als die
Pane, ist der untere Rahmen weggescrollt — der Hotkey also unsichtbar.
Gewuenscht: **Ausnahme fuer Body — Hotkey oben anzeigen.**
Betroffen: `panelBox()` in `internal/tui/box_panel.go` (nutzt `boxBottomBorder`); braucht
eine Variante, die den Hotkey in den OBEREN Rahmen setzt.

## #10 — Keybind `v` fehlt im Browse-Footer
`v` (Vollbild / full-width) ist aktiv, wird aber nicht angezeigt. Ergaenzen.
Wie bei #6 (`r`) vermutlich Nachwirkung von bt-fy5d — pruefen, ob dort zu viel entfernt
wurde. **Beide zusammen betrachten**, es ist derselbe Auswahlfehler.

## Akzeptanz
- [ ] ID neben dem Titel im Detail sichtbar
- [ ] Body-Hotkey auch bei langem Body erreichbar/sichtbar
- [ ] `v` im Footer
- [ ] tmux-Smoke bei 80 Spalten (Footer-Aenderung)
- [ ] voller Testlauf gruen
