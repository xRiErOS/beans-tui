---
# bt-dovm
title: 'S7: huh durch Inline-Box-Editing ersetzen (D09)'
status: draft
type: task
priority: normal
created_at: 2026-07-20T07:25:37Z
updated_at: 2026-07-20T18:05:19Z
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

## D09 REVIDIERT (PO 2026-07-20)

Scope dieses breiten S7-Umbaus ist **verengt** (design-spec.md § PO-Entscheidungen,
D09 REVIDIERT):

- **Create-Form bleibt huh** — der in D09 versprochene huh-Komplett-Ausbau + Wegfall der
  langsamen huh-Drive-Tests entfaellt. PO: Create ist mit huh gut.
- Nur die **endlichen Enum-Felder (Status/Type/Priority)** wandern von Center-Modal zu
  feld-verankertem Inline-Dropdown → umgesetzt im fokussierten Task **bt-f0y9**.

Damit ist die breite Praemisse dieses beans (huh redundant, Create=Box-Form, langsame
Tests weg) ueberholt. Dieses bean bleibt als Historie/Draft; die aktive Arbeit laeuft in
bt-f0y9. PO kann es am Gate scrappen.
