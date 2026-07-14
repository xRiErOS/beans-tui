---
# bt-2jve
title: 'T2 Master-Detail-Verdrahtung: Section/Feld-Fokus, Beziehungs-Sprung, Map-Ownership (I01), nil-Guard (Q01)'
status: in-progress
type: task
priority: high
tags:
    - to-review
created_at: 2026-07-14T21:57:17Z
updated_at: 2026-07-14T22:36:23Z
parent: bt-aq5s
blocked_by:
    - bt-ms0k
---

Ziel: Detail-Fokus-Maschine verdrahten (Section<->Feld, wie devd view_detail_issue.go
keyDetailFocus) + Beziehungs-Sprung (enter auf Parent/Child/Blocking/BlockedBy setzt
Tree-Cursor + expandiert Vorfahren-Pfad) + zwei PFLICHT-Items aus T8-Review: I01
(expanded-map Ownership: Copy-on-Write-Konvention, BEVOR E2 weitere map-Felder
addiert) und Q01 (Init() nil-client-Guard).

Plan: docs/plans/v1-port/epic-E2-plan.md »Task 2«.

## Akzeptanz
- [x] I01: Copy-on-Write-Konvention für map[string]bool-Modellfelder dokumentiert
      (types.go) + cloneBoolMap-Helper + setExpanded refactored; Regressionstest
      beweist keine Shared-Mutation über Modell-Kopien hinweg
- [x] Q01: Init() nil-client-Guard (Test + Fix), kein Panic bei m.client == nil
- [x] keyDetailFocus: Ziffer-Sprung, i/k Section-Wechsel, l/j Section<->Feld-Ebene
      (nur Beziehungen hat Felder), enter auf Beziehungs-Zeile springt + verlässt
      Detail-Fokus
- [x] focusedBean()-Dispatcher (Port devd focusedIssue, view_detail_issue.go:20-35) —
      view-agnostische Basis für T5/Backlog-Wiederverwendung
- [x] view_browse_repo.go renderDetailPane nutzt jetzt beanSections+renderAccordion
      (T1) statt Platzhalter
- [x] tree.golden ggf. regeneriert (-update) + Determinismus-Test weiter grün


## Übernommene Findings aus E2-T1-Quality-Review (PFLICHT in diesem Task)
- [x] B01: metaSectionBody + historieSectionBody bekommen bodyW und wrappen via wrapText (Muster relationsSectionBody) — sonst Pane-Overflow bei langen Tags/ETag
- [x] I02: glowRender baut Renderer pro Aufruf — bewusste Entscheidung beim Verdrahten: Kosten akzeptieren (upstream-gleich) ODER Cache keyed (bean.ID, ETag, width) im Model. Entscheidung im Code dokumentiert.
- [x] I03 (optional): glowRender-Fallback ungewrappt — als Restrisiko kommentiert (nicht gewrappt, s. accordion.go glowRender-Doku)

## Summary

Fokus-Maschine (secCursor/accOpen/detailLevel/fieldCursor, types.go) + keyDetailFocus
(update.go, Port devd view_detail_issue.go:281-392) verdrahtet: Ziffer-Sprung `1`-`4`
öffnet die passende Section exklusiv, i/k bewegt den Section-Cursor (an beiden Enden
geklemmt), l/j wechselt Section<->Feld-Ebene (nur Beziehungen hat Felder), enter auf
einem aufgelösten relationField springt (`m.cursorID`), expandiert den kompletten
Vorfahren-Pfad des Ziels (`expandAncestorsOf`, I01-CoW-basiert) und verlässt den
Detail-Fokus; enter auf einer unresolved-Referenz (`beanID == ""`) ist ein No-Op.
`focusedBean()` ist view-agnostisch (switch mit `default`-Case auf den aktuellen
Tree-Cursor) — Task 5 (Backlog) ergänzt dort nur einen weiteren `case`.

I01: `cloneBoolMap`-Helper + Doku-Stamp in types.go (direkt über `model`) —
`setExpanded` klont jetzt vor jedem Write, `expandAncestorsOf` ebenso. Regressionstest
`TestSetExpandedDoesNotMutateSharedMapAcrossModelCopies` beweist keine Shared-Mutation
mehr über Modell-Kopien hinweg.

Q01: `Init()` hat einen nil-Client-Guard (app.go) — liefert bei `m.client == nil` einen
`beansLoadedMsg{err: errNilClient}` statt in `loadCmd -> Client.List -> Client.run`
(`cmd.Dir = c.RepoDir` auf nil-Pointer) zu paniken. Vorher am realen Panic verifiziert
(Stacktrace bis `client.go:41`), danach grün.

B01 (E2-T1-Review, MANDATORY): `metaSectionBody`/`historieSectionBody` (view_detail_bean.go)
bekommen jetzt `bodyW` und wrappen via `wrapText` (Muster `relationsSectionBody`) — vorher
RED-Test bewiesen Overflow bei langen Tags/ETag (bodyW=20, Zeile 88/56 Zeichen), jetzt grün.

I02/I03 (E2-T1-Review): Entscheidung direkt im `glowRender`-Doc-Kommentar (accordion.go)
dokumentiert — Kosten (neuer Renderer pro Aufruf, jetzt live im View()-Pfad statt nur in
Tests) akzeptiert, kein Cache (YAGNI ohne Profiling-Beleg); der unwrapped Error-Fallback
ist als akzeptiertes Restrisiko kommentiert (beide Fehlerpfade praktisch unerreichbar).

`renderDetailPane` (view_browse_repo.go) ruft jetzt `beanSections(m.idx, b, w-4)` +
`renderAccordion(secs, m.accOpen, w-2, focused, m.secCursor, m.fieldCursor)` statt des
E1-Platzhalters. Scrolling bei zu langem Section-Inhalt: keine eigene Scroll-State
hinzugefügt (out of plan scope) — `renderPane`s bestehender Zeilen-Cap (Golden Rule #1)
verhindert Overflow über den Rahmen hinaus, dasselbe Prinzip wie beim Tree.

`tree.golden` einmalig regeneriert (Detail-Pane zeigt jetzt den Accordion statt
Titel+Meta-Zeile), Determinismus danach erneut mit `-run TestTreeGolden -count=2`
verifiziert.

Tests (neu/erweitert): `update_test.go` (I01-Regression, `TestFocusedBean...`,
9x `TestDetailFocus...`, `TestKeyDetailFocusOnOrphanRootExitsGracefully`),
`app_test.go` (neu, Q01), `view_detail_bean_test.go` (2x B01-Wrap-Regression).
`go test ./... -count=1`, `gofmt -l .`, `go vet ./...` alle clean.

Manueller Smoke (tmux, bin `bt` gegen dieses Repo, `.beans/` dogfooding-Daten):
Tree voll kollabiert -> tab (Detail-Fokus) -> `3` (Beziehungen öffnet) -> `l` (Feld-Ebene)
-> `k` auf ein Children-Feld (`bt-aq5s`) -> `enter`: Tree-Cursor springt zu `bt-aq5s`,
Vorfahre `bt-apmy` (vorher kollabiert) wird automatisch expandiert, `bt-aq5s` ist mit
Cursor-Balken sichtbar, Detail-Fokus verlassen (Tree wieder fokussiert), Detail-Pane
zeigt jetzt `bt-aq5s`s eigene Beziehungen. Zweiter Lauf mit echter Blocked-By-Relation
(`bt-aq5s` blocked_by `bt-blsy`) bestätigt denselben Sprung für Blocking/BlockedBy-Felder.

Commit: 63f0f7aac7c13b465a1d822e3b3b9ed4b5352be1 (main, direkt — Worktree-Weiche
main-direkt für dieses Repo).

Notizen für T3/T4: Task 3 (Suche `/`) und Task 4 (Facetten-Filter) bauen auf derselben
Browse-View (`view_browse_repo.go`, `visibleNodes()`) auf; die hier eingeführten
Fokus-Machine-Felder (`secCursor`/`accOpen`/`detailLevel`/`fieldCursor`) sind orthogonal
dazu und werden von T3/T4 nicht angefasst. `focusedBean()`s `default`-Case ist bewusst
offen für einen künftigen `case viewBacklog` (Task 5) gelassen.
