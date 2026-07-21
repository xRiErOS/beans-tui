---
# bt-oox1
title: 'Detail-View: bean-ID, Body-Hotkey bei langem Body, v im Footer'
status: completed
type: task
priority: normal
created_at: 2026-07-20T09:23:37Z
updated_at: 2026-07-20T10:14:39Z
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


## Summary

**#1 bean-ID** — im Rahmen-Label der Title-Box: `╭─ Title · sproutling-elir ─…─╮`.
Das Label ist der einzige Slot, der dem Titeltext keine Spalten kostet, und
boxTopBorder clampt es bereits, also degradiert eine schmale Pane statt ueberzulaufen.
**Volle ID**, wie im bean empfohlen: bt-pl5p hat die LISTEN-Zeilen gekuerzt, weil der
Slug sich dort auf jeder Zeile wiederholt — dieselbe Entscheidung hielt die
vollstaendige, kopierbare ID in der DETAIL-Pane fest. Der Gegen-Test
`TestRelationRowKeepsFullIDInDetail` ist unangetastet; der neue
`TestDetailBoxFormShowsBeanIDNextToTitle` traegt jetzt die zweite Haelfte derselben
Entscheidung und benennt sie im Kommentar, damit niemand die beiden "vereinheitlicht".

**#4 Body-Hotkey oben** — neues `boxTopBorderHotkey` (box_dropdown.go) +
`panelBoxTopHotkey` (box_panel.go); beide panelBox-Varianten teilen sich EINE
Rahmen-Komposition (`panelBoxWith`), koennen also nicht bei Padding/Clamping/Fokus
driften. Nur die Body-Panel-Zeile in detailBoxForm nutzt sie — Relations/History
haben ohnehin keinen Hotkey. Gleiches defensives Fallback wie boxBottomBorder:
passt Label+Badge nicht, entfaellt das Badge statt den Rahmen zu sprengen.

**#10 `v` im Footer** — `keys.Fullscreen` ergaenzt, aber NUR unter dem Box-Form-Flag.
Das ist eine Budget-Entscheidung, kein Scoping-Versehen: der Flag-AUS-Footer misst
bereits 152 Zellen und fuellt seine zwei Zeilen bei 80 Spalten exakt aus — jeder
weitere Eintrag kippt ihn auf drei und kostet eine Zeile Listeninhalt
(mouse_test.go pinnt die zwei Zeilen als D02-Precondition). Ohne die vier
inline-gebadgten Keys liegt der Flag-EIN-Footer bei 127 Zellen, mit Luft. Epic
bt-vy1q's stehende Auflage (alles additiv + gated, Flag-AUS byte-identisch) zeigt
in dieselbe Richtung, und der PO hat den Befund mit BT_BOXFORM=1 gesehen.

Golden regeneriert, Zeile fuer Zeile per git diff geprueft — genau drei Aenderungen:
detail_boxform.golden + browse_boxform.golden (Title-Label + `(e)` wandert vom
unteren in den oberen Body-Rahmen), browse_boxform.golden (`· v fullscreen`).

Commit 7853ee9.

## Test-Output

command go test ./... — vollstaendig, ohne -short:

    ?   github.com/xRiErOS/beans-tui  [no test files]
    ok  github.com/xRiErOS/beans-tui/cmd  0.506s
    ?   github.com/xRiErOS/beans-tui/internal/clip  [no test files]
    ok  github.com/xRiErOS/beans-tui/internal/config  (cached)
    ok  github.com/xRiErOS/beans-tui/internal/data  (cached)
    ok  github.com/xRiErOS/beans-tui/internal/theme  (cached)
    ok  github.com/xRiErOS/beans-tui/internal/tui  153.757s

tmux-Smoke 80x24 gegen ~/Obsidian/tools/sproutling, BT_BOXFORM=1, frisches Binary
und frische Session: `╭─ Title · sproutling-elir ─`, `╭─ Body ──── (e) ───╮`,
Footer weiterhin ZWEIZEILIG mit `· v fullscreen`.

## Deviations

`v` erscheint nur bei BT_BOXFORM=1 (Begruendung oben). Der Flag-AUS-Footer bleibt
byte-identisch — `TestBrowseRepoLocalBindingsUnchangedWithoutBoxForm` haelt das fest.
