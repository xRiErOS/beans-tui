---
# bt-9yvh
title: E6 T2 — US-Validierungslauf Teil 2 (US-08..14)
status: in-progress
type: task
priority: normal
created_at: 2026-07-15T14:00:45Z
updated_at: 2026-07-15T19:14:28Z
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

[ ] US-08: go test -run "TestReviewQueueGroupsByEpicCanonicalOrder|
    TestKeyReviewCockpitPassFiresPassReview|TestKeyReviewCockpitRejectOpensCommentForm|
    TestPassReviewSetsCompletedAndRemovesTag|TestRejectReviewSwapsTagAndAppendsSection|
    TestReviewSummaryLinePositionInToReview" alle PASS.
[ ] US-09: go test -run "TestBacklogShowsParentlessReadyBeansFromIndex|
    TestBacklogSortCyclesThroughFourModesAndBackToStart|TestBacklogGolden" alle PASS.
    PO-Entscheid I02 (Sort-Indikator) NICHT selbst entschieden, geht unverändert in D02 (T3).
[ ] US-10: go test -run "TestWatcherFiresOnceForBurst|TestReloadKeepsCursorOnID|
    TestSwitchRepoCmdStopsOldWatcherStartsNew" alle PASS; Repo-Wechsel-Reload-Fall LIVE im
    kombinierten Smoke reproduziert (eigenständiger Beleg, nicht nur Verweis auf bt-7dfj).
[ ] US-11: go test -run "TestYankShowsConfirmationToast|TestYankOnOrphanRootNoop|
    TestReviewCockpitYankUsesReviewStandNotSingleBean" alle PASS; im Smoke y auf Bean UND
    y im Review-Cockpit ausgelöst, pbpaste-Inhalt nach OSC52-Capture geprüft.
[ ] US-12: VQA-Befund (bean bt-zk9p Body, docs/_free-notes/vqa-2026-07-15/) mit
    Verzeichnis-Verweis übernommen — keine neue Screenshot-Session.
[ ] US-13: internal/tui/view_lobby.go auf Footer/Hint-Text geprüft (grep), Befund
    (vorhanden/GAP) im kombinierten Smoke live abgelesen und dokumentiert.
[ ] US-14 + kombinierter Smoke: zwei Scratch-Repos (/tmp/bt-scratch-a, /tmp/bt-scratch-b),
    eigenes HOME mit config.yaml (repos: beide Pfade) — p-Lobby zeigt offen/gesamt-Metrik
    beider Repos, enter wechselt, externe Änderung im ALTEN Repo löst KEINEN Reload mehr
    aus, externe Änderung im NEUEN Repo löst Reload <2s aus, state.json zeigt zuletzt
    aktives Repo.
[ ] Hygiene bt-gzcu I01 (Konflikt-Statuszeile 1-Frame): NUR bei grünem
    TestConflictToastIsStickyAndSurvivesReload Status-Zeile 🟡→🟢 + Testverweis aktualisiert.
[ ] command go test ./... -short grün, gofmt/vet leer.
[ ] Commit test(tui): US-08..14 Validierungsbelege + Repo-Wechsel-Smoke (E6 Task 2) —
    Body enthält alle Tabellen/Belege für T3.
