---
# bt-2p9m
title: Filter-Menü als Querformat mit Tab-Kategorien
status: todo
type: feature
created_at: 2026-07-17T06:21:42Z
updated_at: 2026-07-17T06:21:42Z
parent: bt-5uzr
---

NB aus PO-Review E11 (2026-07-17, Runde 2), PO verbatim:

"Das Filter-Menü ist sehr umfangreich. Wenn ich als Nutzer zu einem konkreten
Filter gehen möchte, dann muss ich sehr lange mit navigieren. Besser wäre, wenn
das Filter-Menü querformat wird und die Filterkategorien in Tabs dargestellt
werden. Klickpfad: 1) f für Filter, erster tab aktiviert 2) mit tab/shift-tab
andere Filter wählen 3) im fokussierten Tab ist das erste Element immer aktiv
und ich kann mit Pfeil-rauf/runter direkt die Auswahl navigieren 4) mit
leertaste filter-kriterium wählen, 5) mit enter den Filter anwenden."

Interpretation: Facetten-Overlay (f, box_filter_facets.go) von vertikaler
Gesamtliste auf horizontales Tab-Layout umbauen — eine Filterkategorie
(Status/Type/Priority/Tags/Archive) je Tab. Interaktionsmodell komplett vom PO
spezifiziert (5 Schritte oben). Akzeptanz-Entwurf:
- [ ] f öffnet Overlay im Querformat, erster Tab aktiv
- [ ] tab/shift-tab wechseln Kategorie-Tabs
- [ ] Fokus-Tab: erstes Element aktiv, Pfeil rauf/runter navigiert direkt
- [ ] Leertaste toggelt Kriterium, enter wendet an
- [ ] bestehende Filter-Semantik (Facetten-Logik) unverändert

Planner verfeinert vor Umsetzung (Breiten-Verhalten 80 Spalten beachten,
NBSP-Wordwrap-Falle, LL Eintrag 4).
