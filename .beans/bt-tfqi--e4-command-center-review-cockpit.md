---
# bt-tfqi
title: E4 Command-Center & Review-Cockpit
status: completed
type: epic
priority: high
tags:
    - to-review
created_at: 2026-07-14T18:33:30Z
updated_at: 2026-07-18T15:04:23Z
parent: bt-apmy
blocked_by:
    - bt-gzcu
---

Ziel: `ctrl+k`-Palette (Aktionen + Bean-Treffer, kontextabhängig) und Review-Flow: Queue-View `R` (Tag to-review, gruppiert nach Epic), Cockpit mit `a` pass (completed + Tag weg), `x` reject (Kommentar-Form → `## Review <datum>`-Body-Abschnitt + Tag-Swap to-review→rework), `o` reopen, Yank Review-Stand.

Epos-Start-Ritual wie E2 (epic-E4-plan.md). Review-Konvention: design-spec §5. Port-Quellen: overlay_palette.go, view_navigate_reviews.go, view_review_sprint.go, keys_review.go.


## E4-Abschluss (Agent-Meldung, 2026-07-15)
T1-T5 completed, je Task Spec+Quality-Review, 4 Fix-Runden (in-flight-Guard Palette, detailFocus-Reset, reviewCursor-Clamp via applyLoaded, argv-Garantie gepinnt). Command-Center (fuzzy, Aktionen+Bean-Suche, kontextabhängig) + Review-Cockpit (Queue gruppiert nach Epic, Verdikte a/x/o mit Ein-Call-Mutationen) komplett. US-04+US-08-Smoke belegt inkl. Boundary-Fall (Pass des letzten Eintrags). 7 Goldens deterministisch, Suite ~126s voll / ~8s -short.
