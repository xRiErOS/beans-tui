---
# bt-0xrb
title: 'Tag-Picker: ETag-Konflikt-Toast abgeschnitten, Ursache unklar'
status: in-progress
type: bug
priority: normal
created_at: 2026-07-17T09:43:51Z
updated_at: 2026-07-17T10:53:21Z
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


## PO-Entscheid Grilling 2026-07-17 (D04)

1. Upstream-Issue bei hmans/beans wird gefiled (Entwurf → PO-Freigabe → Absenden).
2. Die 10 betroffenen lean-stack-beans werden sofort geheilt (einmalige kanonische Mutation; außerhalb dieses beans).
3. KEINE Konflikt-Sonderfall-Mitigation in beans-tui. Scope dieses beans damit final: B01a — Toast wächst dynamisch bis Meldung vollständig (alle Severities), plain-ErrConflict-Zweig befüllt toastCtx mit handlungsleitenden Details (mirrort update.go:290-303-Zweig). Regressionstests overlay_show_toast_test.go + etag_conflict_test.go.


## Plan-Konkretisierung E13 (2026-07-17)

Plan: `docs/plans/v1-port/epic-E13-plan.md` §„Item 1: Toast wächst dynamisch
+ toastCtx im plain-ErrConflict-Zweig". Reihenfolge: Rang 1 (Toast-Familie,
vor `bt-tm4a`, sequenziell in EINEM Worktree).

**Root Cause (file:line, verifiziert gegen Ist-Code 2026-07-17):**
- `overlay_show_toast.go:146` `const toastBoxWidth = 36` — fixe Breite.
- `overlay_show_toast.go:156` `ansi.Truncate(t.title, innerW-2, "…")` —
  innerW-2=30, „Conflict: bean changed externally" (34 Zeichen) wird
  abgeschnitten.
- `overlay_show_toast.go:160` dieselbe Truncate-Logik für `context`.
- `update.go:275-327` (`applyMutationResult`): plain-ErrConflict-Zweig
  (Zeile 279-303, kein `*conflictWithRecovery`-Match) lässt `toastCtx = ""`
  (Zeile 290) — `err.Error()` (Bean-ID + CLI-Detail aus
  `internal/data/mutations.go:63/75`) wird verworfen.

**Vorgehen (Kurzfassung, Details im Plan):** `toastBoxWidth` durch
content-getriebene Breite ersetzen (geklemmt `[32, min(m.width-4, 70)]`,
Cap=70 Planner-Entscheidung), `ansi.Truncate("…")` durch Wordwrap ersetzen
(alle drei `toastKind`-Severities), plain-ErrConflict-Zweig `toastCtx` mit
`err.Error()` vorbelegen (vor dem `errors.As(&cr)`-Override).

**Akzeptanz (abhakbar, siehe Plan für Volltext):**
- [ ] `toastBoxWidth` content-getrieben, geklemmt `[32, min(m.width-4, 70)]`
- [ ] Kein `ansi.Truncate("…")` mehr für Titel/Context — Wordwrap statt
      Abschneiden
- [ ] Gilt für alle drei `toastKind`-Severities
- [ ] Plain-ErrConflict-Zweig setzt nicht-leeren `toastCtx` aus `err.Error()`
- [ ] Recovery-Pfad-Verhalten (`toastCtx = "Version saved: …"`) unverändert
- [ ] Test-Suite grün, neue Tests in `overlay_show_toast_test.go` +
      `etag_conflict_test.go`
- [ ] tmux-Smoke: Conflict-Repro (`lean-stack-n0ly` via `t`-Picker) zeigt
      vollständige Meldung + Detailzeile, kein Abschneiden

Scope bleibt final D04-gebunden (siehe eigene „PO-Entscheid Grilling
2026-07-17"-Sektion oben): NUR B01a, keine Konflikt-Sonderfall-Mitigation.
