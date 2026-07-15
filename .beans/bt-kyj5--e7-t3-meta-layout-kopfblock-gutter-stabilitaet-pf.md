---
# bt-kyj5
title: E7 T4 — Meta-Layout + Kopfblock + Gutter-Stabilitaet (PF-1, PF-3, PF-4, PF-12)
status: in-progress
type: task
priority: high
created_at: 2026-07-15T14:26:51Z
updated_at: 2026-07-15T16:14:22Z
parent: bt-heg9
blocked_by:
    - bt-2af1
    - bt-w9o8
---

Details/Steps/Akzeptanz: docs/plans/v1-port/epic-E7-plan.md Task 4. Detail-Pane-Kopfblock (ID/Titel/type-status-prio) + Meta-Sektion wird 6-zeilige editierbare Feldliste, nicht-kollabierbar, Gutter-Stabilitaet (PF-12).


## Hinweis aus T2-Review (I01, für diesen Task relevant)

Die 3 Priority-Glyphen ·/↓/→ sind Unicode-Ambiguous-width (EAW-Klasse A; ‼/! sind safe). T4 schafft die ERSTEN Render-Stellen für theme.Priority() (Kopfblock prio:-Zeile + Meta-Feldliste) und regeneriert Detail-Goldens. Auf Terminals mit ambiguous=wide droht dort Layout-Shift. Umgang: PO-Schema ist verbindlich (keine Glyph-Abweichung) — aber beim Layout die Prio-Spalte so bauen, dass 1-vs-2-Zellen-Rendering keinen Umbruch erzeugt (Padding NACH dem Glyph, Glyph am Feldende oder feste Breite 2), und den Umstand im Golden-Diff-Report dokumentieren.


## Prelude aus T3-Review (I01, medium — ZUERST, eigener Commit)

Plan-Step 4 des Englisch-Tasks versprach Fuzzy-Regressionstests und wurde als [x] abgehakt, aber NICHT geliefert: TestPalFilteredActionsFuzzyFiltered testet weiter nur 'bckl'→backlog (Reihenfolge unverändert = risikoärmster Fall). Die 5 wortgedrehten Labels (set status/tags/parent/blocking/title) sind ungetestet. Nachziehen: Subtests 'stat'→'set status' (+ nur die set-Einträge?) und 'go'→genau die 4 'go to'-Einträge (backlog/browse/repo picker/settings). Commit test(tui): Fuzzy-Regression verb-entity-Labels (T3-Review I01), Refs: bt-kyj5.
