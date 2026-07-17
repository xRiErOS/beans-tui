---
# bt-2kfl
title: Suche mit Filter-Präfixen (st:completed ty:epic)
status: todo
type: feature
created_at: 2026-07-17T06:21:42Z
updated_at: 2026-07-17T06:21:42Z
parent: bt-5uzr
---

NB/Idea aus PO-Review E11 (2026-07-17, Runde 2), PO verbatim:

"Wenn ich Filter anwende, dann wird mir bspw. oben angezeigt 'St:completed
Ty:epic' Es wäre super, wenn diese Filter auch durch das tippen aktiviert
werden. Also ich drücke '/' (suche aktiv) und tippe 'st:completed Ty:epic &
bean das ich suche' Das führt dann dazu, dass ich nur solche beans durchsuche,
die completed AND epuc [epic] sind und die string-search erfüllen."

Interpretation: Such-Input (/) parst Filter-Präfixe (st:/ty:/pr:/tag: —
Kürzel-Set analog Header-Anzeige) und kombiniert sie AND-verknüpft mit der
verbleibenden String-Suche. Präfixe wirken wie die entsprechenden Facetten-
Filter. Akzeptanz-Entwurf:
- [ ] "/st:completed ty:epic foo" filtert auf status=completed AND type=epic AND match "foo"
- [ ] Präfixe case-insensitiv, Reihenfolge egal
- [ ] ungültiges Präfix/Wert → als normaler Suchtext behandeln (kein Fehler)
- [ ] Header-Filteranzeige spiegelt getippte Filter (eine Wahrheit mit Facetten-State)

Offene Planner-Fragen: Sync-Richtung Suche↔Facetten-Overlay (setzt Tippen die
Facetten-Auswahl?), Bleve- vs. lokaler Filter-Pfad ab 3 Zeichen.
