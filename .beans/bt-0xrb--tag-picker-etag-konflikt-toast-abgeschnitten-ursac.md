---
# bt-0xrb
title: 'Tag-Picker: ETag-Konflikt-Toast abgeschnitten, Ursache unklar'
status: todo
type: bug
priority: normal
created_at: 2026-07-17T09:43:51Z
updated_at: 2026-07-17T09:52:16Z
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


## PO-Bestätigung Review R2 (2026-07-17, US-02)

PO wörtlich: „NB: Auch hier: Die Fehlermeldung sollte vollständig angezeigt werden >> Toast muss dynamisch größer werden. Haben wir bei US-01 schon angemerkt."

Konkretisierte Anforderung an den UX-Teil dieses beans: Toast wächst dynamisch (Breite/Mehrzeiler) bis die Meldung VOLLSTÄNDIG lesbar ist — kein Truncate mit `...`. Gilt für alle Toast-Severities, nicht nur Conflict.


## Diagnose-Ergebnis (ce-diagnose, 2026-07-17)

**Root Cause: Upstream-Bug in beans/beancore 0.4.2 — NICHT beans-tui.** Zwei inkonsistente ETag-Quellen:
- `beans list --json` liefert `hash(Bean.Render())` der IN-MEMORY-Repräsentation, in die der Loader beim Laden stille Defaults füllt (`Priority "" → "normal"`, beans-src/pkg/beancore/core.go:205-219).
- `beans update --if-match` validiert gegen `hash(rohe Datei-Bytes)` (core.go:560-576).

Divergiert deterministisch bei jedem bean, dessen Datei kein `priority:`-Feld trägt (bulk-importiert, nie durch beans-Writer gelaufen) — z. B. `lean-stack-n0ly` (list-etag 54ff1e8f… vs. raw 4c728326…). Kein Watcher-Refresh kann das heilen (content-basiert, kein Timing). Self-Heal: erste erfolgreiche Mutation schreibt die Datei kanonisch → deshalb funktionierte `lean-stack-o4c4` (schon einmal mutiert). 9 weitere beans im lean-stack-Repo tragen dieselbe latente Klasse. beans-tui-Mutationspfad (box_picker_tag.go:401-431, update.go:681-698, data/mutations.go) verifiziert korrekt.

**B01a (beans-tui, unabhängig fixbar):** overlay_show_toast.go:146/156 — toastBoxWidth=36 → Titel-Budget 30 Zeichen, „Conflict: bean changed externally" (34) wird geclippt. Zusätzlich übergibt der plain-ErrConflict-Zweig (update.go:290/303) `toastCtx=""` — keine bean-ID/ETag-Details im Toast, anders als der Editor-Conflict-Zweig.

**Fix-Optionen:**
1. B01a in beans-tui: toastCtx für plain-ErrConflict befüllen (mirrort update.go:290-303) + PO-Anforderung „Toast wächst dynamisch" umsetzen; Regressionstests overlay_show_toast_test.go/etag_conflict_test.go.
2. Upstream-Issue bei hmans/beans filen (Update-Validierung muss dieselbe Repräsentation hashen wie list/ETag()).
3. Optional Mitigation in bt (Konflikt-Sonderfall erkennen, handlungsleitende Meldung) — Diagnose-Empfehlung: eher nicht (Komplexität vs. self-healing Quirk).
4. Sofort-Heilung Daten-Seite: betroffene lean-stack-beans einmal kanonisch durchschreiben (trivial, außerhalb bt).
