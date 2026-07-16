---
# bt-lg68
title: created_at/modified_at doppelt in 1 META und 4 HISTORY
status: todo
type: bug
priority: normal
created_at: 2026-07-16T20:20:40Z
updated_at: 2026-07-16T20:47:11Z
parent: bt-tct9
---

PO-Nebenbefund, US-Review Runde 3 (2026-07-16): Detail-View zeigt created_at/
modified_at sowohl in Sektion '1 META' als auch in '4 HISTORY'. Sollen NUR in
'4 HISTORY' erscheinen (META bleibt Typ/Status/Prio/Tags).

Fundort vermutlich view_detail_bean.go Meta-Section-Renderer.
Quelle: bt-tct9 US-Review Runde 3.



## Planner-Konkretisierung (2026-07-16)

**Root Cause exakt lokalisiert:** `metaFields()` (view_detail_bean.go:106-124)
hängt zwei `{kind: "readonly", label: fmtTime(b.CreatedAt)}` /
`{kind: "readonly", label: fmtTime(b.UpdatedAt)}`-Einträge an (Zeilen
121-122) — gerendert von `metaSectionBody` (Zeile 137ff.) über
`metaFieldLabels` (Zeile 92, Einträge 6/7: `"created_at:"`,
`"updated_at:"`). `historieSectionBody` (Zeile 432-445) rendert
Created/Updated ZUSÄTZLICH, unabhängig — daher die Dopplung.

**Fix:** die beiden `readonly`-Einträge (Zeilen 121-122) UND die zwei
letzten `metaFieldLabels`-Strings (Zeile 92, `"created_at:"`/
`"updated_at:"`) aus `metaFields()`/`metaFieldLabels` entfernen — META
schrumpft von 7 auf 5 Felder (title/status/type/priority/tags). HISTORY
bleibt alleinige Quelle (unverändert, Created/Updated/ETag).

**Ripple-Check (PFLICHT, da `metaFields()`s Rückgabelänge in mehreren
Stellen als Cursor-Range verwendet wird):** `keyDetailFocus`s
Feld-Navigation (`update.go`, `secs[m.secCursor].fields`, Klammerung an
`len(fields)`) · `activateDetailField`s `kind`-Switch (readonly = No-Op,
Wegfall betrifft nur die beiden entfernten Einträge, alle verbleibenden
kind-Werte unverändert) · `metaSectionBody`s Zeilen-Schleife über
`metaFieldLabels` (Padding-Berechnung nutzt `metaFieldLabels`' gemeinsame
Breite — mit `"created_at:"` als bisher LÄNGSTEM Label entfernt, wird
`"priority:"` [9 Zeichen] neuer längster Eintrag, Padding-Breite MUSS neu
verifiziert werden, nicht hart-kodiert auf 12 belassen falls das der Fall
ist) · jeden bestehenden Test, der `len(metaFields(b))==7` oder
Feld-Index 5/6 (readonly) referenziert (`grep -rn "metaFields\|fieldIdx"
internal/tui/*_test.go` als Startpunkt).

**Akzeptanzkriterium-Zusatz:**
- META zeigt NUR NOCH title/status/type/priority/tags (5 Zeilen), kein
  created_at/updated_at mehr.
- HISTORY unverändert (Created/Updated/ETag, `historieSectionBody`).
- Pfeiltasten-Navigation innerhalb META funktioniert weiterhin für die
  verbleibenden 5 Felder (kein Off-by-one, kein Zugriff auf entfernte
  Indizes).
- Bestehende Tests, die auf 7 Meta-Felder referenzieren, angepasst statt
  gebrochen liegen gelassen.
- Golden-Regen/-Gegenbeleg PFLICHT (META-Sektion ist Teil von
  Tree/Backlog/Chrome-Goldens).
