---
# bt-lg68
title: created_at/modified_at doppelt in 1 META und 4 HISTORY
status: completed
type: bug
priority: normal
created_at: 2026-07-16T20:20:40Z
updated_at: 2026-07-16T21:19:58Z
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


## Summary

Root Cause bestätigt: `metaFields()`/`metaFieldLabels` (view_detail_bean.go)
trugen zwei `readonly`-Einträge (created_at/updated_at), gerendert von
`metaSectionBody` — `historieSectionBody` rendert Created/Updated zusätzlich
und unabhängig. Fix: beide Einträge aus `metaFields()`/`metaFieldLabels`
entfernt, META schrumpft 7→5 Felder (title/status/type/priority/tags).
Padding-Breite (`metaFieldLabelWidth`) jetzt aus `metaFieldLabels` berechnet
statt hart-kodiertem `12`-Literal (12→10, da `"priority:"` neuer längster
Eintrag). `activateDetailField`s `"readonly"`-Switch-Case (update.go) als
toter Code mitentfernt (kein Producer mehr — default-Zweig verhält sich für
`beanID==""` identisch), inkl. zugehöriger Doc-Kommentare (accordion.go
`relationField.kind`).

## Test-Output (RED→GREEN)

RED (vor dem Fix, Testdatei bereits auf 5-Felder-Erwartung umgestellt):
```
--- FAIL: TestMetaFieldsFiveEntriesWithKinds (0.00s)
    view_detail_bean_test.go:66: metaFields returned 7 entries, want 5
      (bt-lg68: created_at/updated_at removed, HISTORY is now the sole source)
FAIL
```

GREEN (nach dem Fix): `TestMetaFieldsFiveEntriesWithKinds`,
`TestMetaSectionBodyShowsSelectedFieldMarker`,
`TestMetaSectionBodyShowsAllFiveLabelsAndValues` etc. alle PASS.

Voller Lauf (`go test ./... -count=1`), ×2 für Snapshot-Stabilität:
- Lauf 1: `ok beans-tui/internal/tui 139.972s` (alle Pakete grün)
- Lauf 2: `ok beans-tui/internal/tui 139.620s` (alle Pakete grün)
- Lauf 3 (nach Kommentar-Nacharbeit in mouse.go/update.go/accordion.go):
  `ok beans-tui/internal/tui 138.927s`

`go build -o bin/bt .`, `go vet ./...`, `gofmt -l` — alle sauber.

## Golden-Diffs

- `tree.golden`: META-Block schrumpft von 7 auf 5 Zeilen (created_at:/
  updated_at:-Zeilen weg), Label-Padding 12→10 Spalten (title:/status:/
  type:/priority:/tags: enger), 2 Leerzeilen als Füllung am Sektionsende
  (fixe Accordion-Box-Höhe unverändert).
- `backlog.golden`: identisches Muster (gleicher META-Renderer, andere
  Fixture-Werte: "Second backlog task"/status d/type B/prio ‼/tag backend).
- `chrome.golden`: unverändert (kein META-Inhalt in diesem Snapshot).

## Smoke

tmux (100×30), echtes `./bin/bt .`-Binary, bean `bt-6oyy`:
- Nach `tab`+`right`: META zeigt exakt 5 Felder (title/status/type/
  priority/tags), `▶`-Marker auf title.
- `down`×4 durch status→type→priority→tags: Marker wandert korrekt Zeile
  für Zeile, ein zusätzlicher `down` auf tags bleibt geklammert (kein
  Absturz, kein Off-by-one).
- `esc` zurück auf Sektionsebene (kein Marker mehr sichtbar), `4` springt
  zu `[4] HISTORY`: zeigt `Created: 2026-07-15 14:11` / `Updated:
  2026-07-16 15:47` / `ETag: 83047018ae2fe62a` — unverändert vorhanden,
  META zeigt kein created_at/updated_at mehr.
- Ehrliche Abgrenzung: Maus-Klick-Pfad (`detailClickRow`/`mouseDetailClick`)
  NICHT separat live gesmoked, nur Unit-Level (bereits grün) — Klick-
  Geometrie ist rein generisch (`len(s.fields)`, `lipgloss.Height`), keine
  Zeilenzahl-Literale betroffen.

## Deviations/ERRATA

- Plan-Zeilenangaben (`metaFields()` 106-124, `metaFieldLabels` 92,
  `historieSectionBody` 432-445) stimmten exakt mit Ist-Code überein, keine
  Abweichung.
- Zusätzlich zum Plan-Scope entfernt (Ripple-Check Punkt 4 explizit
  gefordert, "falls ja, mitentfernen"): `activateDetailField`s `"readonly"`-
  Case (update.go) — wurde durch die Feld-Entfernung tatsächlich toter Code
  (kein Producer mehr). Zwei zugehörige Tests entfernt:
  `TestKeyDetailFocusEnterOnCreatedAtFieldNoop`,
  `TestKeyDetailFocusEnterOnUpdatedAtFieldNoop` (Feldindizes 5/6 existieren
  nicht mehr), sowie `TestActivateDetailFieldReadonlyIsNoop` (testete exakt
  den entfernten Switch-Case).
- Kein `accordion.go`-Verhaltenscode geändert (nur Doc-Kommentar zu
  `relationField.kind`) — B04/`bt-98cb`s Scope (voriges Segment kollabiert
  nicht) bleibt unberührt.

## Notes for next task

`bt-98cb` arbeitet als Nächstes in derselben Datei-Familie
(`accordion.go`/`update.go`/`view_browse_repo.go`). Relevant für dessen
Kontext: `metaFieldLabelWidth` ist jetzt selbst-berechnet (nicht mehr
`12`-hart-kodiert) — falls `bt-98cb` weitere Meta-Feld-Änderungen anfasst,
Padding bleibt automatisch korrekt. `relationField.kind == "readonly"`
existiert nicht mehr als Wert (Doc-Kommentar in accordion.go entsprechend
aktualisiert) — falls `bt-98cb`s Code irgendwo noch `"readonly"` erwartet,
das ist jetzt tote Referenz.
