---
# bt-hd42
title: Klick trifft nach Scrollen falsches Feld (Offset fehlt)
status: in-progress
type: bug
priority: high
created_at: 2026-07-20T09:24:15Z
updated_at: 2026-07-20T09:49:37Z
parent: bt-vy1q
---

**Bug, gefunden von bt-1o4g (nicht dort verursacht).**

`detailBoxFormClickRow` ignoriert `boxFormScroll`. Nach dem Scrollen der Detail-Pane
treffen Mausklicks daher die **ungescrollten** Feldzonen — man klickt auf Status und
oeffnet Type.

## Herkunft
Entstanden aus der Kombination von bt-ze10 (Scroll-Offset eingefuehrt) und der frueheren
Maus-Slice S6 (Trefferzonen ohne Offset gerechnet). Beide fuer sich korrekt, zusammen
falsch. Von bt-1o4g weder verursacht noch verschlimmert, dort nur entdeckt.

## Warum das dringlich ist
Es ist nicht nur ein eigener Fehler, es ist ein **Fundament, auf dem gerade weitere
Features geplant werden**: Pencil-Klick auf Body-Ueberschriften und die Anker-Leiste
brauchen ebenfalls scroll-korrigierte Trefferzonen. Wird das hier nicht repariert, erben
sie den Fehler.

## Betroffen
`internal/tui/mouse.go` — `detailBoxFormClickRow`, `boxFormFieldAt`, `gridColAt`.
Der Offset liegt in `boxFormScroll`/`boxFormEffectiveScroll` (types.go, bt-ze10) bereit.

## Akzeptanz
- [ ] Klick trifft nach dem Scrollen das richtige Feld
- [ ] Test, der ERST scrollt und DANN klickt (der fehlende Fall)
- [ ] Mausrad und Tastatur-Scroll fuehren zum selben Ergebnis
- [ ] voller Testlauf gruen
