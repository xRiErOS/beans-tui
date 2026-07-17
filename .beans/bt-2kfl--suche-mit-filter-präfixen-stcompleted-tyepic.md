---
# bt-2kfl
title: Suche mit Filter-Präfixen (st:completed ty:epic)
status: completed
type: feature
priority: normal
created_at: 2026-07-17T06:21:42Z
updated_at: 2026-07-17T11:44:39Z
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

## Summary

Such-Präfixe `st:`/`ty:`/`pr:`/`tag:` in der `/`-Suche implementiert, gemäß
D02 (separater additiver Layer)/D03 (Parser bei jedem Tastendruck), final
gemäß PO-Entscheidungen Grilling 2026-07-17 — nicht neu aufgemacht.

**Neue Datei `internal/tui/search_prefix.go`** (Implementer-Wahl, hält
`view_browse_repo.go`/`box_filter_facets.go`-Diffs klein, Datei-Disziplin
Parallel-Welle):
- `parseSearchPrefixes(query) (facets map[string][]string, rest string)` —
  whitespace-Tokenizer, `<prefix>:<value>` mit `prefix` ∈ {st,ty,pr,tag}
  case-insensitiv (Kürzel→Facet-Mapping analog `facetHead` minus `archive`).
  Ungültige/unvollständige Tokens (kein Doppelpunkt, leerer Präfix, leerer
  Wert, unbekanntes Präfix) fallen unverändert in `rest`. Werte werden
  lower-cased gespeichert (Implementer-Entscheidung — Konsistenz mit den
  bereits lowercase Status/Type/Priority-Enum-Werten, siehe Deviations).
- `applySearchPrefixes()` — re-derived beide neuen Model-Felder aus
  `m.searchQuery`, einziger Schreib-Call (`keySearchInput`, alle drei
  Zweige: Tastendruck/Enter/Esc).
- `beanMatchesSearchPrefixFacets`/`searchPrefixFacetHit`/`containsFold` —
  AND über Facetten, OR innerhalb einer Facette (Membership-Semantik wie
  `beanMatchesFacets`), case-insensitiv (EqualFold).

**Geänderte Dateien:**
- `types.go`: zwei neue Felder `m.searchPrefixFacets map[string][]string`,
  `m.searchPrefixRest string` — GETRENNT von `m.filterStatus`/etc. (D02).
- `update.go`: `keySearchInput` ruft `applySearchPrefixes()` in allen drei
  Zweigen; `maybeBleveCmd`/`applyBleveResult` laufen jetzt gegen
  `m.searchPrefixRest` statt der vollen `m.searchQuery` (D03); Reset-Stellen
  (`applyRepoSwitched`, esc-cascade Rung 2 in `keyTree`) räumen die beiden
  neuen Felder mit auf.
- `view_browse_repo.go`: `beanMatchesSearch` prüft zuerst
  `beanMatchesSearchPrefixFacets`, dann den Text-Pfad gegen
  `m.searchPrefixRest` (statt `m.searchQuery`).
- `box_filter_facets.go`: NUR `filterSummary` angefasst (Datei-Disziplin) —
  lokale `union`-Closure (kein neuer Top-Level-Func) zeigt pro Facet-Zeile
  die Union aus `m.filterStatus`/etc. UND `m.searchPrefixFacets`; ohne
  getippte Präfixe byte-identisch zum Vorzustand (golden-safe).
- `messages.go`: Doc-Kommentar von `searchBleveResultMsg` auf
  `m.searchPrefixRest` aktualisiert (war `m.searchQuery`).

## Test-Output

RED (vor Implementierung, `go vet` schlug fehl):
```
vet: internal/tui/search_prefix_test.go:103:20: undefined: parseSearchPrefixes
```

GREEN (nach Implementierung, neue + angepasste Tests):
```
=== RUN   TestParseSearchPrefixes (12 Subtests) ... PASS
=== RUN   TestSearchPrefixMatchingAndsAcrossFacetsOrsWithinFacet --- PASS
=== RUN   TestSearchPrefixCombinesWithTextRest --- PASS
=== RUN   TestSearchPrefixInvalidTokenTreatedAsPlainText --- PASS
=== RUN   TestSearchPrefixDoesNotMutateFilterMenuState --- PASS
=== RUN   TestFilterSummaryShowsUnionOfMenuFacetsAndTypedPrefixes --- PASS
=== RUN   TestFilterSummaryUnchangedWithoutTypedPrefixes --- PASS
=== RUN   TestFilterSummaryClearingQueryDropsTypedFilters --- PASS
=== RUN   TestBlevePrefixExcludedFromDispatchedQuery --- PASS
=== RUN   TestBleveNotDispatchedBelowThresholdWhenOnlyPrefixesTyped --- PASS
PASS
```

Voller Lauf (`go test ./... -count=1`, ohne `-short`): **GREEN**, 148.3s
(internal/tui), Gesamtlaufzeit 2:28.74.

Golden ×2 (`go test ./internal/tui/... -run Golden -count=1`, zweimal
hintereinander): beide GREEN, `git status --short internal/tui/testdata/`
leer nach beiden Läufen — byte-identisch, keine Regen nötig (`filterSummary`
degradiert ohne getippte Präfixe exakt auf den Vorzustand).

`go vet ./...`: clean. `gofmt -l .`: clean (ein Formatierungs-Delta in
`box_filter_facets_test.go` durch den Auto-Formatter beim Speichern
korrigiert, keine funktionale Änderung).

## Smoke

tmux-Session `bt2kfl$$` (worktree-Binary `bin/bt` gegen die
Worktree-`.beans/`-Kopie):
1. `/st:completed ty:epic` getippt → Tree filtert live (0 Treffer korrekt,
   keine `completed`-Beans im Fixture-Repo), Kopf zeigt `St:completed
   Ty:epic` exakt.
2. Query committed (Enter), `f` geöffnet → Filter-Menü zeigt `[ ] completed`
   UNCHECKED trotz aktivem `St:completed` im Kopf — kein Mit-Togglen (D02
   bestätigt).
3. `/st:completed foo` getippt → Kopf zeigt `St:completed` (Text-Rest `foo`
   geht separat in die Bleve-Schwelle, nicht in die Facet-Anzeige).
4. Doppel-Esc → Query+Filter vollständig geleert, voller Tree wieder
   sichtbar.
5. `git status --short .beans/` nach dem Smoke: leer (clean).

## Deviations/ERRATA

- Q2/Q3 aus der E12-Konkretisierung sind bereits durch D02/D03 (Grilling
  2026-07-17) final beantwortet — nicht erneut aufgemacht, wie im Bean
  vorgegeben.
- **Implementer-Entscheidung (nicht in D02/D03/Plan spezifiziert):**
  Präfix-WERTE (nicht die Präfix-Keywords) werden beim Parsen lower-cased
  gespeichert. Begründung: Status/Type/Priority-Enum-Werte sind im gesamten
  Code bereits durchgängig lowercase (`data.StatusValues()` etc.,
  `buildFilterItems`); ohne diese Normalisierung hätte `St:Epic` in der
  Union-Anzeige inkonsistent neben dem Menü-eigenen `epic` gestanden. Das
  Matching selbst ist ohnehin case-insensitiv (`strings.EqualFold`) und wäre
  auch ohne die Normalisierung korrekt gewesen — der Lowercase-Schritt dient
  ausschließlich der Anzeige-Konsistenz.
- **Bestehende Tests angepasst (14 Stellen, 6 Dateien):** Tests, die
  `m.searchQuery` bisher direkt setzten (ohne echten `tea.Update`-Rundlauf),
  mussten auf einen neuen Test-Helfer `setSearchQuery(m, q)`
  (`update_test.go`) umgestellt werden, der `m.searchPrefixRest`/
  `m.searchPrefixFacets` synchron mithält — sonst hätten diese Tests (keine
  Präfix-Syntax enthalten) durch den Wechsel von `beanMatchesSearch` auf
  `m.searchPrefixRest` fälschlich 0 Treffer gesehen. Rein mechanische
  Anpassung, keine Verhaltensänderung der jeweiligen Testfälle.
- Ein neuer Test (`TestBleveNotDispatchedBelowThresholdWhenOnlyPrefixesTyped`)
  musste nach der ersten (zu strengen) Fassung angepasst werden: eine
  transiente Zwischen-Eingabe wie `"st:"` (3 Zeichen, unvollständiger Wert)
  fällt laut Spec korrekterweise als Klartext in `rest` und darf legitim
  einen Bleve-Dispatch auslösen — das ist kein Bug, sondern exakt die
  PO-Akzeptanz "ungültiges Präfix/Wert → normaler Suchtext". Der Test prüft
  jetzt den AKTUELLEN Dispatch-Entscheid (`maybeBleveCmd() == nil`) für die
  fertig getippte reine Präfix-Query, nicht das (durch einen früheren
  Zwischenzustand potenziell noch gesetzte) `searchBleveLoading`-Flag.
