---
# bt-tm4a
title: 'Toast-Severity vereinheitlichen: createInFlight warn vs error'
status: todo
type: task
priority: low
created_at: 2026-07-17T08:08:31Z
updated_at: 2026-07-17T10:07:32Z
parent: bt-5uzr
---

Reviewer-Finding aus bt-81f0-Review (2026-07-17, non-blocking): Die Meldung `createInFlightNote` erscheint an zwei Guards derselben Aktion in ZWEI Severities — `update.go:735` nutzt `toastWarn`, die bt-81f0-neuen Stellen (`box_confirm_create.go:48`, `overlay_palette.go:231`) nutzen `toastError` (bindender Task-Rahmen schrieb toastError wörtlich vor). Nutzer sieht dieselbe Meldung mal gelb, mal rot.

Aufgabe: eine Severity wählen (Empfehlung: `toastWarn` — in-flight-Guard ist kein Fehler) und alle drei Stellen angleichen. Kleiner Fix, Tests anpassen falls Severity asserted wird.

Akzeptanz:
- [ ] Alle createInFlightNote-Toasts nutzen dieselbe Severity
- [ ] Begründung der Wahl im Abschluss dokumentiert
- [ ] Test-Suite grün


## Plan-Konkretisierung E13 (2026-07-17)

Plan: `docs/plans/v1-port/epic-E13-plan.md` §„Item 2:
createInFlightNote-Toast-Severity vereinheitlichen". Reihenfolge: Rang 2
(Toast-Familie, NACH `bt-0xrb`, gleiches Worktree/gleiche Session —
Datei-Nachbarschaft `update.go`/`overlay_palette.go`).

**Root Cause (file:line, verifiziert gegen Ist-Code 2026-07-17):**
- `update.go:735` → `toastWarn` (bereits korrekt, Ziel-Severity).
- `box_confirm_create.go:56` → `toastError` (Ziel: ändern zu `toastWarn`).
- `overlay_palette.go:240` → `toastError` (Ziel: ändern zu `toastWarn`).
- ERRATUM-Kommentare an allen drei Stellen (`box_confirm_create.go:49-53`,
  `overlay_palette.go:233-237`, `update.go:737`) dokumentieren die
  Diskrepanz bereits selbst.
- Tests, die den aktuellen (inkonsistenten) Zustand fixieren:
  `box_confirm_create_test.go:415-417`, `overlay_palette_test.go:462-463`
  (beide `toast.kind != toastError`, Kommentar „bt-81f0 bindender Rahmen").

**Vorgehen:** `box_confirm_create.go:56` + `overlay_palette.go:240` auf
`toastWarn` ändern; alle drei ERRATUM-Kommentare entfernen/aktualisieren;
beide Test-Assertions auf `toastWarn` drehen.

**Akzeptanz:**
- [ ] Alle drei `createInFlightNote`-Toasts nutzen `toastWarn`
- [ ] ERRATUM-Kommentare entfernt/aktualisiert (kein Verweis auf offene
      Diskrepanz mehr)
- [ ] Begründung (toastWarn statt toastError: in-flight-Guard ist kein
      Fehler) im Bean-Abschluss dokumentiert
- [ ] `box_confirm_create_test.go`/`overlay_palette_test.go`-Assertions
      angepasst
- [ ] Test-Suite grün
