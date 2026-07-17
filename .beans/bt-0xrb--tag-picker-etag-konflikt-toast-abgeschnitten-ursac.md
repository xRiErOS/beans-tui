---
# bt-0xrb
title: 'Tag-Picker: ETag-Konflikt-Toast abgeschnitten, Ursache unklar'
status: todo
type: bug
created_at: 2026-07-17T09:43:51Z
updated_at: 2026-07-17T09:43:51Z
parent: bt-5uzr
---

NB aus PO-Review E12 Runde 1 (2026-07-17), US-01-Einschränkung. PO wörtlich: siehe bt-362n `## Review 2026-07-17`.

Symptome:
1. Beim Taggen von `lean-stack-n0ly` (Tags smoke, smoke3) via `t`-Picker: Toast `Conflict: bean changed extern...` — Meldung abgeschnitten, erklärt nicht, WARUM das Tagging nicht klappt.
2. `lean-stack-o4c4` lässt sich mit `t` normal taggen — Konflikt ist bean-spezifisch, nicht global.

Zwei Aspekte:
- B (UX): Konflikt-Toast muss vollständig sichtbar sein (Wrap/Mehrzeiler statt Truncate) und handlungsleitend erklären (z. B. „bean wurde extern geändert — Liste aktualisiert, bitte erneut versuchen").
- B (Ursache): Warum hält bt für n0ly einen stalen ETag? Kandidaten: Upstream-Quirk beans 0.4.2 (create --tag liefert stalen ETag, LESSONS-LEARNED E10/4), Watcher-Refresh-Lücke, tatsächliche externe Änderung (lean-stack-Repo hatte uncommittete .beans-Änderungen).

Diagnose-Ergebnis wird hier angehängt (ce-diagnose, 2026-07-17).
