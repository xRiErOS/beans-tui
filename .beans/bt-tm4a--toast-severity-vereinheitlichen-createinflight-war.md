---
# bt-tm4a
title: 'Toast-Severity vereinheitlichen: createInFlight warn vs error'
status: todo
type: task
priority: low
created_at: 2026-07-17T08:08:31Z
updated_at: 2026-07-17T08:08:31Z
parent: bt-5uzr
---

Reviewer-Finding aus bt-81f0-Review (2026-07-17, non-blocking): Die Meldung `createInFlightNote` erscheint an zwei Guards derselben Aktion in ZWEI Severities — `update.go:735` nutzt `toastWarn`, die bt-81f0-neuen Stellen (`box_confirm_create.go:48`, `overlay_palette.go:231`) nutzen `toastError` (bindender Task-Rahmen schrieb toastError wörtlich vor). Nutzer sieht dieselbe Meldung mal gelb, mal rot.

Aufgabe: eine Severity wählen (Empfehlung: `toastWarn` — in-flight-Guard ist kein Fehler) und alle drei Stellen angleichen. Kleiner Fix, Tests anpassen falls Severity asserted wird.

Akzeptanz:
- [ ] Alle createInFlightNote-Toasts nutzen dieselbe Severity
- [ ] Begründung der Wahl im Abschluss dokumentiert
- [ ] Test-Suite grün
