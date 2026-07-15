---
# bt-uyzf
title: E7 T5 — Pane-Titel-Vereinheitlichung (PF-10)
status: in-progress
type: task
priority: normal
created_at: 2026-07-15T14:26:51Z
updated_at: 2026-07-15T16:50:54Z
parent: bt-heg9
blocked_by:
    - bt-kyj5
---

Details/Steps/Akzeptanz: docs/plans/v1-port/epic-E7-plan.md Task 5. Entfernt redundante Pane-Titel (Tree/Backlog/Detail), Breadcrumb traegt View-Identitaet. Klick-Geometrie (clickPaneGeometry) muss mitziehen.


## Optionaler Mini-Punkt aus T4-Review (I01, trivial)

TestPalFilteredActionsFuzzyGoMatchesAllFourGoToEntries nimmt implizit an, dass kein fixtureBeans()-Titel 'go' fuzzy-matcht — 1 Kommentar-Zeile (oder kind-Assertion) ergänzen, wenn du eh in Test-Dateien bist. Kein Pflichtteil.
