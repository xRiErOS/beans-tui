---
# bt-2kfl
title: Suche mit Filter-Präfixen (st:completed ty:epic)
status: todo
type: feature
priority: normal
created_at: 2026-07-17T06:21:42Z
updated_at: 2026-07-17T10:08:09Z
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


## Plan-Konkretisierung E13 (2026-07-17)

Plan: `docs/plans/v1-port/epic-E13-plan.md` §„Item 3: Suche mit
Filter-Präfixen". Reihenfolge: Parallel-Welle (mit `bt-d3ps`/`bt-nxuk`,
disjunkte Dateien), NACH der Toast-Familie (`bt-0xrb`/`bt-tm4a`).

D02/D03 (siehe eigene „PO-Entscheidungen Grilling 2026-07-17"-Sektion oben)
bleiben final, NICHT neu aufgemacht: additiver Layer (`f`-Menü unberührt),
Parser bei jedem Tastendruck, nur Rest-Text erreicht Bleve.

**ERRATUM (Zeilen-Drift seit der ursprünglichen Diagnose):**
`box_filter_facets.go`s `filterSummary` liegt jetzt bei Zeile **500-517**
(nicht mehr 324-339) — `bt-2p9m`s Querformat-Umbau (Commits `b150c9f`/
`e3b1701`, nach dieser Diagnose gemerged) hat das File umsortiert.
`facetHead` liegt jetzt bei Zeile **32-38**. `view_browse_repo.go:251-260`
(`beanMatchesSearch`) ist UNVERÄNDERT, kein Drift dort.

**Vorgehen (Kurzfassung, Details im Plan):**
1. Neue Funktion `parseSearchPrefixes(query string) (facets
   map[string][]string, rest string)` — Kürzel-Set `st`/`ty`/`pr`/`tag`,
   case-insensitiv, ungültige Präfixe/Werte fallen in `rest`.
2. `keySearchInput` (`update.go:1588-1607`): pro Tastendruck parsen, Ergebnis
   in NEUEN Model-Feldern (`m.searchPrefixFacets`/`m.searchPrefixRest`)
   ablegen, GETRENNT von `m.filterStatus`/etc. `dispatchBleveIfDue` dispatcht
   gegen `rest`, nicht die volle Query.
3. `beanMatchesSearch` (`view_browse_repo.go:251-260`): Substring-Abgleich
   gegen `m.searchPrefixRest`, zusätzliche AND-Bedingung gegen
   `m.searchPrefixFacets` (Membership-Semantik wie `beanMatchesFacets`).
4. `filterSummary` (`box_filter_facets.go:500-517`): UNION-Anzeige aus
   `m.filterStatus`/etc. UND `m.searchPrefixFacets` je Facet-Zeile.

**Akzeptanz (siehe Plan für Volltext, unverändert aus Bean-Entwurf
übernommen):**
- [ ] `/st:completed ty:epic foo` filtert AND-verknüpft korrekt
- [ ] Präfixe case-insensitiv, Reihenfolge egal
- [ ] Ungültiges Präfix/Wert → normaler Suchtext, kein Fehler
- [ ] `filterSummary` spiegelt UNION aus Menü-Facetten + getippten Präfixen
- [ ] Parser bei jedem Tastendruck, nur `rest` erreicht Bleve
- [ ] `m.filterStatus`/etc. bleibt von getippten Präfixen unberührt
- [ ] Test-Suite grün, neue Tests für `parseSearchPrefixes` + Union-Anzeige
- [ ] tmux-Smoke: `/st:completed ty:epic` → Tree-Kopf zeigt
      `St:completed Ty:epic`, `f`-Menü bleibt beim Öffnen leer
