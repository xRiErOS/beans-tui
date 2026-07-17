---
# bt-tm4a
title: 'Toast-Severity vereinheitlichen: createInFlight warn vs error'
status: completed
type: task
priority: low
created_at: 2026-07-17T08:08:31Z
updated_at: 2026-07-17T11:22:44Z
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

## Summary

Alle drei `createInFlightNote`-Toasts vereinheitlicht auf `toastWarn`:
`update.go:747` blieb unverändert (bereits Ziel-Severity),
`box_confirm_create.go:56` und `overlay_palette.go:240` von `toastError`
auf `toastWarn` geändert. Begründung (Akzeptanz-Pflicht): der in-flight-
Guard ist ein Hinweis ("Creation already in progress — please wait"),
kein Fehler — `toastWarn` (Gelb) passt der Nutzer-Intention, `toastError`
(Rot) bleibt echten Fehlern vorbehalten (z.B. dem CLI-Rejected-Create-Pfad,
box_confirm_create_test.go:267, unberührt). Alle drei ERRATUM-Kommentare
(box_confirm_create.go, overlay_palette.go, beide Pin-Tests) entfernt bzw.
auf den jetzt einheitlichen Zustand umgeschrieben.

## Test-Output

RED (vor Severity-Swap, Pin-Tests bereits auf toastWarn gedreht):
```
--- FAIL: TestSubmitFormCreateIgnoredWhilePendingCreateInFlight (23.96s)
    box_confirm_create_test.go:416: toast.kind = 2, want toastWarn (bt-tm4a: unified in-flight-guard severity)
--- FAIL: TestDispatchPaletteCreateIgnoredWhileCreateInFlight (0.00s)
    overlay_palette_test.go:464: toast.kind = 2, want toastWarn (bt-tm4a: unified in-flight-guard severity)
```

GREEN (nach Severity-Swap):
```
--- PASS: TestSubmitFormCreateIgnoredWhilePendingCreateInFlight (18.96s)
--- PASS: TestDispatchPaletteCreateIgnoredWhileCreateInFlight (0.00s)
```

Voller Lauf (`go test ./... -count=1`, ohne -short): GREEN, alle Pakete ok,
internal/tui 150.594s, Gesamtdauer 2:31 (innerhalb erwarteter ~155s). Build
(`go build -o bin/bt .`), `go vet ./...`, `gofmt -l` alle sauber.

## Deviations/ERRATA

- Plan/Bean nannten einen ERRATUM-Kommentar bei `update.go:737`, der die
  Diskrepanz dokumentiert. Im Ist-Code (verifiziert) trägt update.go an
  dieser Stelle (jetzt Zeile 743-745, Zeilendrift durch bt-0xrb) NUR die
  ursprüngliche Design-Begründung ("E5 Task 1 (bean bt-6dts): Toast additiv
  neben m.err ... toastWarn, nicht toastError") — KEIN ERRATUM-Wortlaut,
  KEIN Verweis auf eine Diskrepanz. update.go bleibt daher unverändert
  (Plan-Vorgabe "update.go:735 bleibt unverändert" erfüllt); nichts zu
  entfernen/umschreiben dort.
- Zeilendrift gegen Plan/Bean-Vorlage (bt-0xrb-Merge, wie vom Supervisor
  vorgewarnt): update.go-Guard jetzt bei Zeile 747 statt 735,
  ERRATUM-Referenz-Zeile bei ~743-745 statt 737. box_confirm_create.go:56
  und overlay_palette.go:240 stimmten exakt mit der Planvorlage überein
  (keine Drift dort).
- box_confirm_create_test.go:267 (CLI-Rejected-Create-Reopen-Pfad, B01)
  wurde geprüft und bewusst NICHT angefasst — anderer Fehlerpfad (echter
  CLI-Fehler, nicht der in-flight-Guard), toastError dort korrekt.

## Notes für Parallel-Welle

Nächste Tasks bt-2kfl/bt-d3ps/bt-nxuk laufen in eigenen Worktrees.
bt-d3ps fügt voraussichtlich einen neuen `dispatchPalette`-Case
("register project") in overlay_palette.go hinzu — Ist-Stand nach diesem
Commit: overlay_palette.go hat 303 Zeilen, letzter case-Block bei
Zeile 268 (`case "settings":`), switch-Ende bei ~270-280. Die case-Labels
in diesem Task (`case "create":` Zeile 221) haben sich durch den
Kommentar-Umbau NICHT verschoben (Netto-Zeilenänderung in
overlay_palette.go: -0, reine Kommentar-Umschreibung gleicher Zeilenzahl
minus 1 Zeile durch kürzeren Kommentarblock — siehe git diff --stat:
14 Zeilen geändert bei diesem File). bt-d3ps sollte vor dem Patchen
trotzdem den Ist-Code erneut zählen (ERRATUM-Kultur), da parallele
Worktree-Arbeit an bt-0xrb-Nachbarn (update.go) schon einmal zu Drift
führte.
