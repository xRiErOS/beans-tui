---
# bt-9yvh
title: E6 T2 — US-Validierungslauf Teil 2 (US-08..14)
status: completed
type: task
priority: normal
created_at: 2026-07-15T14:00:45Z
updated_at: 2026-07-15T19:25:44Z
parent: bt-zk9p
blocked_by:
    - bt-heg9
---

Ziel: Für US-08..14 (design-spec.md §10) je einen dokumentierten Validierungs-Nachweis
erbringen. Die meisten sind laut Matrix-Entwurf bereits "vollständig" (Bestätigungslauf
reicht) — Kern-Arbeit: US-14 + US-10-Zusatz in EINEM kombinierten Zwei-Repo-Smoke
(Lobby, Watcher-Lifecycle-Switch, state.json-Persistenz), US-11 ein eigenständiger
OSC52-Beleg statt Fremdverweis auf bt-7dfj, US-12 VQA-Übernahme (KEIN neuer Screenshot-Lauf
— Supervisor hat das bereits 2026-07-15 mit 16 Screenshots erledigt).

Plan: docs/plans/v1-port/epic-E6-plan.md »Task 2«.

## Akzeptanz

[x] US-08: go test -run "TestReviewQueueGroupsByEpicCanonicalOrder|
    TestKeyReviewCockpitPassFiresPassReview|TestKeyReviewCockpitRejectOpensCommentForm|
    TestPassReviewSetsCompletedAndRemovesTag|TestRejectReviewSwapsTagAndAppendsSection|
    TestReviewSummaryLinePositionInToReview" alle PASS.
[x] US-09: go test -run "TestBacklogShowsParentlessReadyBeansFromIndex|
    TestBacklogSortCyclesThroughFourModesAndBackToStart|TestBacklogGolden" alle PASS.
    PO-Entscheid I02 (Sort-Indikator) NICHT selbst entschieden, geht unverändert in D02 (T3).
[x] US-10: go test -run "TestWatcherFiresOnceForBurst|TestReloadKeepsCursorOnID|
    TestSwitchRepoCmdStopsOldWatcherStartsNew" alle PASS; Repo-Wechsel-Reload-Fall LIVE im
    kombinierten Smoke reproduziert (eigenständiger Beleg, nicht nur Verweis auf bt-7dfj).
[x] US-11: go test -run "TestYankShowsConfirmationToast|TestYankOnOrphanRootNoop|
    TestReviewCockpitYankUsesReviewStandNotSingleBean" alle PASS; im Smoke y auf Bean UND
    y im Review-Cockpit ausgelöst, pbpaste-Inhalt nach OSC52-Capture geprüft.
[x] US-12: VQA-Befund (bean bt-zk9p Body, docs/_free-notes/vqa-2026-07-15/) mit
    Verzeichnis-Verweis übernommen — keine neue Screenshot-Session.
[x] US-13: internal/tui/view_lobby.go auf Footer/Hint-Text geprüft (grep), Befund
    (vorhanden/GAP) im kombinierten Smoke live abgelesen und dokumentiert.
[x] US-14 + kombinierter Smoke: zwei Scratch-Repos (/tmp/bt-scratch-a, /tmp/bt-scratch-b),
    eigenes HOME mit config.yaml (repos: beide Pfade) — p-Lobby zeigt offen/gesamt-Metrik
    beider Repos, enter wechselt, externe Änderung im ALTEN Repo löst KEINEN Reload mehr
    aus, externe Änderung im NEUEN Repo löst Reload <2s aus, state.json zeigt zuletzt
    aktives Repo.
[x] Hygiene bt-gzcu I01 (Konflikt-Statuszeile 1-Frame): NUR bei grünem
    TestConflictToastIsStickyAndSurvivesReload Status-Zeile 🟡→🟢 + Testverweis aktualisiert.
[x] command go test ./... -short grün, gofmt/vet leer.
[x] Commit test(tui): US-08..14 Validierungsbelege + Repo-Wechsel-Smoke (E6 Task 2) —
    Body enthält alle Tabellen/Belege für T3.

## Summary

Alle US-08..14 validiert. 6/7 PASS, 1 PARTIAL (US-08: Filter+Live-Reload
funktionieren voll, aber Tree/Detail zeigen Tags gar nicht — B-Finding
`bt-gdkx`, medium, Repro dokumentiert, NICHT gefixt). Kombinierter
Zwei-Repo-Smoke (`/tmp/bt-scratch-a`+`/tmp/bt-scratch-b`, eigenes
`HOME`-Overlay) deckt US-10/US-11/US-14 in einem durchgehenden Live-Lauf ab
(Lobby-Metrik, Watcher-Switch beide Richtungen, externe Tag-Änderung ohne
`ctrl+r`, echter OSC52/pbpaste-Yank-Beleg, `state.json`-Persistenz). US-12
per VQA-Übernahme (16 Screenshots, KEIN neuer Lauf) + 3 frische
Post-E7-Delta-Captures (Header/Footer-Split, Glyphen, Kopfblock, Help-Overlay
— bestätigt keine Cockpit-Reste, PF-14 vollständig sauber). US-13s
Matrix-GAP ("Lobby-Footer existiert noch nicht") ist bereits geschlossen,
kein Fix nötig. bt-gzcu I01 bei grünem Testbeweis aktualisiert (🟡→🟢). Volle
Belege inkl. aller Tabellen/Capture-Auszüge:
docs/_free-notes/e6-t2-evidence.md (git-ignored, Input für T3).

## Ergebnis-Übersicht

| US    | Titel                        | Ergebnis | Kurzbeleg |
|-------|-------------------------------|----------|-----------|
| US-08 | Review-Stand sehen (redef.)  | PARTIAL  | Filter+Live-Reload PASS; Tree/Detail zeigen Tags GAR NICHT (bt-gdkx, medium) |
| US-09 | Backlog priorisiert           | PASS     | 4/4 Tests grün; D02 Sort-Indikator bekannt-offen (bt-7k7q) |
| US-10 | Live-Reload                   | PASS     | 3/3 Tests grün + Live-Beleg beide Richtungen im Zwei-Repo-Smoke |
| US-11 | Yank/Clipboard                | PASS     | 2/2 Tests grün + echter OSC52/pbpaste-Beleg, Toast englisch |
| US-12 | devd-Look                     | PASS     | VQA-Übernahme + 3 frische Post-E7-Delta-Captures, keine Regression |
| US-13 | Hilfe/Footer-Hints            | PASS     | Matrix-GAP geschlossen; Help-Overlay 3-Gruppen/Englisch/esc-?-q live |
| US-14 | Repo-Wechsel                  | PASS     | Voller Zwei-Repo-Smoke: Lobby-Metrik, Watcher-Switch, state.json |

## ERRATA

- **Plan-Testnamen (US-08/US-11, Matrix-Entwurf VOR E7):**
  `TestReviewQueueGroupsByEpicCanonicalOrder`,
  `TestKeyReviewCockpitPassFiresPassReview`,
  `TestKeyReviewCockpitRejectOpensCommentForm`,
  `TestReviewSummaryLinePositionInToReview`,
  `TestReviewCockpitYankUsesReviewStandNotSingleBean` existieren nicht mehr
  (Review-Cockpit vollständig entfernt, PF-14, Commit `a25b851`). Kein FAIL —
  ersatzlos entfernt. Nur die 2 Datenlayer-Tests
  (`TestPassReviewSetsCompletedAndRemovesTag`/
  `TestRejectReviewSwapsTagAndAppendsSection`) bestehen noch (YAGNI-Entscheidung
  PF-14: Datenlayer bleibt harmlos bestehen).
- **US-13 Matrix-Lücke bereits geschlossen:** Lobby-Footer-Hint existiert
  bereits (`view_lobby.go:264-286`), kein GAP mehr, kein Fix nötig.
- **`date +%s%3N`:** GNU-date-Syntax, macOS-inkompatibel — Python-Workaround
  wie in T1 verwendet.

## Notes for T3

- B01 (bean `bt-gdkx`, US-08 Tag-Sichtbarkeit Tree/Detail) braucht PO-Entscheid:
  Fix jetzt vs. zurückstellen (Filter funktioniert als Workaround) — als
  D-Punkt in validation.md aufnehmen, nicht nur Bug-Fußnote.
- D02 (Sort-Indikator Backlog, bean `bt-7k7q`) bleibt unverändert offen, dieser
  Task hat nur den Ist-Zustand (kein Indikator) nochmal bestätigt.
- VQA-Screenshots (docs/_free-notes/vqa-2026-07-15/) sind vor PF-14 entstanden
  (Dateinamen zeigen noch Review-Cockpit) — bei T3-Verweis mit Hinweis "vor E7,
  s. Post-E7-Delta-Check E6-T2" versehen.
- Scratch-Repos /tmp/bt-scratch-a, /tmp/bt-scratch-b, /tmp/bt-scratch-home
  bleiben bestehen (nicht aufgeräumt) — T2 kann wiederverwendet oder neu
  aufgebaut werden, T1s /tmp/bt-scratch-100//tmp/bt-scratch-etag ebenfalls
  unverändert liegen gelassen.
- Volle Rohbelege: docs/_free-notes/e6-t2-evidence.md (git-ignored, direkt für
  validation.md T3 übernehmbar).
