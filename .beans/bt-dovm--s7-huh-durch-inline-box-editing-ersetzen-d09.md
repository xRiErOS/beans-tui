---
# bt-dovm
title: 'S7: huh durch Inline-Box-Editing ersetzen (D09)'
status: draft
type: task
priority: normal
created_at: 2026-07-20T07:25:37Z
updated_at: 2026-07-20T07:25:37Z
parent: bt-vy1q
---

**D09 aus der Design-Spec** — der letzte und groesste Umbau des Spikes. PO wollte das Timing selbst steuern; daher `draft`, bis freigegeben.

## Idee
Im jira-Modell IST das Detail das Edit-Formular. Wenn jedes Feld eine editierbare Box mit Inline-Popup ist, wird die separate huh-Form redundant:
- Create-Form (7 Felder, huh) → leere Box-Form gleichen Layouts
- Einzel-Picker → eigene Inline-Popups
- **Synergie:** eigene Popups sind maus-nativ → loest den huh-MouseMsg-Blocker ohne Workaround (huh faengt MouseMsg ab, `update.go` routet Maus vor `updateForm()`)
- Nebeneffekt: die sieben langsamen huh-Drive-Tests (`box_confirm_create_test.go`, je ~16-19s, `skipSlowHuhDriveInShortMode`) fallen weg

## Risiko
Grosser Eingriff in bestehende, ausgereifte Form-/Picker-Pfade. Sollte in eigene Slices zerlegt werden (nicht als ein Task umsetzen).

## Vor Beginn zu klaeren
- [ ] PO-Freigabe fuer Timing (Status draft → todo)
- [ ] Slice-Kette planen (Create-Form, Picker, huh-Entfernung, Test-Aufraeumen getrennt)
- [ ] Entscheidung: weiterhin hinter `BT_BOXFORM` gated oder als echter Ersatz?
