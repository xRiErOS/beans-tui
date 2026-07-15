---
# bt-v7ti
title: E4 T5 — E4-Abschluss-Ritual
status: in-progress
type: task
priority: normal
created_at: 2026-07-15T05:40:16Z
updated_at: 2026-07-15T08:21:53Z
parent: bt-tfqi
blocked_by:
    - bt-yy6w
---

## Übernommene Findings aus E4-T4-Review (PFLICHT)
- [ ] B01: m.reviewCursor nach boundary-shrinking Pass nie im Model geclampt (nur render-lokal) → reviewFocused nil auf sichtbar markierter Zeile, down/n eingefroren. Fix: Clamp in applyLoaded (Spiegel der cursorID-Reconciliation) ODER am Anfang von keyReviewCockpit. Regressionstest über echte mutateCmd→mutationDoneMsg→Reload-Pipeline (nicht nur View()).
- [ ] I01: Datenlayer-Test RejectReview-Kommentar mit Anführungszeichen, Backtick, Umlaut, eingebettetem \n (argv-Garantie pinnen)
- [ ] Q01 (Doc-Note): Palette-Delete im Cockpit trifft jetzt reviewed Bean — generisches Delete-Confirm reicht, kurz im bean dokumentieren
