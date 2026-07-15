---
# bt-yy6w
title: 'E4 T4 — Verdikt-Aktionen: Pass/Reject/Reopen + Palette-Review-Eintrag'
status: in-progress
type: task
priority: normal
created_at: 2026-07-15T05:40:13Z
updated_at: 2026-07-15T07:31:13Z
parent: bt-tfqi
blocked_by:
    - bt-hxyo
---

## Übernommene Findings aus E4-T3-Review (PFLICHT)
- [ ] I01: reviewQueue/reviewFlat/reviewRework werden 3-4× pro Render/Keypress neu abgeleitet — einmal pro Aufruf berechnen und als Parameter durchreichen BEVOR T4 weitere Call-Sites ergänzt
- [ ] I02: renderReviewDetailPane vs renderBeanAccordionPane ~15 Zeilen dupliziert — kleiner privater Helper (open/secCursor/fieldCursor/focused als Parameter), Kopplung bleibt draußen
- [ ] Testlücken: (a) to-review-getaggtes Epic selbst (Gruppen-Header + eigener Eintrag), (b) Render-Clamp-Pfad bei extern geschrumpfter Queue, (c) rework-only Queue
- [ ] focusedBean(): Cockpit-Case ergänzen (reviewFlat[reviewCursor]) — Palette-Node-Actions wirken sonst auf den falschen Bean (T3-Note)
