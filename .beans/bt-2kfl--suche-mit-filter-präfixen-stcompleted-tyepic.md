---
# bt-2kfl
title: Suche mit Filter-Präfixen (st:completed ty:epic)
status: todo
type: feature
priority: normal
created_at: 2026-07-17T06:21:42Z
updated_at: 2026-07-17T09:58:26Z
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


## Plan-Konkretisierung E12 (2026-07-17)

Plan: `docs/plans/v1-port/epic-E12-plan.md` §„Item 6: Suche mit
Filter-Präfixen". Reihenfolge: Rang 6 (vor `bt-2p9m` — offene Sync-Frage
Q2 sollte vor dem Tab-Layout-Umbau geklärt sein).

**Ist-Stand:** `beanMatchesSearch` (`view_browse_repo.go:251-260`) liest
`m.searchQuery` als reinen Freitext, KEINE Präfix-Erkennung. Facetten
(`status`/`type`/`priority`/`tag`) leben in eigenen Maps
(`box_filter_facets.go`), nur über das Filter-Menü (`f`) beschreibbar.
`beanMatches` (`box_filter_facets.go:189-191`) UND-verknüpft Suche und
Facetten bereits. `filterSummary` (`box_filter_facets.go:324-339`) rendert
bereits das vom PO zitierte Kürzel-Format ("St:completed Ty:epic") —
kanonische Quelle für die gewünschten Präfixe (`st:`/`ty:`/`pr:`/`tag:`).

**Vorgehen (Gerüst, Details nach Q2/Q3-Antwort):** neue Parse-Funktion
(z. B. `parseSearchPrefixes(query string) (facets map[string][]string, rest
string)`) tokenisiert nach `<prefix>:<value>`-Paaren (case-insensitiv,
Kürzel-Set `st`/`ty`/`pr`/`tag`, mirrort `facetHead` minus `archive`).
Ungültige Präfixe/Werte fallen in `rest` (kein Fehler, PO-Akzeptanz
explizit). `rest` geht durch den bestehenden `beanMatchesSearch`-Pfad,
geparste Präfixe je nach Q2-Antwort additiv in `beanMatchesFacets`
eingespeist ODER direkt in `m.filterStatus`/etc. geschrieben (Sync mit
`f`-Menü).

**Offene Fragen (Plan):**
- Q2: Sync-Richtung Suche↔Facetten-Overlay — togglen getippte Präfixe das
  `f`-Menü sichtbar mit, oder bleibt es ein separater additiver Layer?
- Q3: Parser bei JEDEM Tastendruck (lokal) oder erst ab Bleve-Schwellenwert
  (3 Zeichen)? Bleve indiziert nur Titel+Body, keine Facet-Felder.

Falls ohne PO-Antwort begonnen werden muss: getroffene Annahme explizit im
Bean-Summary dokumentieren.

**Akzeptanz:**
- [ ] `/st:completed ty:epic foo` filtert auf status=completed AND type=epic
      AND Text-Match "foo"
- [ ] Präfixe case-insensitiv, Reihenfolge egal
- [ ] Ungültiges Präfix/Wert → normaler Suchtext, kein Fehler
- [ ] Header-Filteranzeige (`filterSummary`) spiegelt getippte Filter (gemäß
      Q2-Antwort)
- [ ] Q2/Q3-Antworten oder dokumentierte Annahme im Bean-Summary
- [ ] Test-Suite grün, neue Tests für `parseSearchPrefixes`


## PO-Entscheidungen Grilling 2026-07-17 (D02/D03, final)

- **D02 — Sync-Richtung (ehem. Q2): Separater additiver Layer.** Präfixe bleiben Teil der Query, Parser wertet live aus; f-Menü-State (`m.filterStatus` etc.) bleibt unberührt. Filteranzeige im Tree-Kopf (`filterSummary`) zeigt die UNION aus Menü-Facetten und getippten Präfixen. Query löschen = getippte Filter weg.
- **D03 — Parser-Zeitpunkt (ehem. Q3): Bei jedem Tastendruck.** Parser trennt lokal Präfix-Paare + Rest-Text; NUR der Rest-Text geht den bestehenden Such-Pfad (Bleve ab 3 Zeichen des Rests). Präfixe erreichen Bleve nie.
