---
# bt-dsog
title: E7 T8 — Abschluss (Epic to-review, E6-Bestaetigung)
status: in-progress
type: task
priority: normal
created_at: 2026-07-15T14:26:51Z
updated_at: 2026-07-15T18:27:55Z
parent: bt-heg9
blocked_by:
    - bt-uyzf
    - bt-m6at
    - bt-t1uy
---

Details/Steps/Akzeptanz: docs/plans/v1-port/epic-E7-plan.md Task 8. Epic to-review setzen, E6-Blocking (bereits gesetzt) bestaetigen, Task-beans T1-T7 auf completed.


## Punkt aus T3-Review (Q01, low)

README.md-Keybinding-Tabelle + Fließtexte sind noch deutsch; PF-7-Scope war explizit nur internal/tui-Produktionscode. Bei der T8/README-Finalisierung entscheiden+umsetzen: Empfehlung KONSISTENT ENGLISCH (Tool-UI ist englisch; README beschreibt UI-Strings — gemischte Sprache wirkt ungepflegt). PO kann widersprechen (Eriks Repo-Doku ist sonst deutsch).


## Prelude aus T7-Review (Mini-Fixes ZUERST, eigener Commit)

- **B01 (docs-only):** epic-E7-plan.md Zeile ~750-752 — doppelte Golden-Checkbox-Zeile (eine [x] mit Beleg, identische [ ] als Artefakt) — Duplikat-Zeile entfernen.
- **I03 (low):** footer_context.go-Kommentar (+ design-spec-Parenthese) behauptet 'mirrors handleKey's dispatch order' — stimmt nicht (handleKey: searchActive VOR filterOpen; Footer: Filter vor Suche). Folgenlos (Staaten nie gleichzeitig), aber Claim korrigieren ODER Invarianten-Test ergänzen.

## Offene PO-Punkte für Epic-Review (mit to-review-Tag hochreichen)

Aus T7 (Reviewer-Einstufung): **I01 medium** — 80-Spalten-Header truncatet q:quit weg (gängiges Terminalmaß; Empfehlung: Header wrappen wie Footer oder Prioritäts-Truncation). **I02 low** — Overlay-Footer restaten Enter/Back trotz sichtbarem Header (gewollte Verstärkung? Sign-off oder Invarianten-Test). Plus aus T7-Implementer: **D01** — Footer-Umfang eng nach Plan (f/X/b/t/a/B/y bewusst raus; PO kann erweitern). Alle in der Epic-Review-Zusammenfassung listen, nicht still passieren lassen.
