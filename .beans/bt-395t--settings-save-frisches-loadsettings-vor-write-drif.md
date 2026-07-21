---
# bt-395t
title: 'Settings-Save: frisches LoadSettings vor Write (Drift-Fenster)'
status: completed
type: task
priority: low
created_at: 2026-07-17T11:55:08Z
updated_at: 2026-07-18T15:05:06Z
parent: bt-5uzr
---

Reviewer-Beobachtungen aus bt-d3ps-Review (2026-07-17, non-blocking, GEERBTES Verhalten — kein neuer Bug):

1. **Drift-Fenster bei Settings-Writes:** `registerProject` (overlay_palette.go) und der Settings-Formular-Pfad schreiben `SaveUserSettings(...)` auf Basis von `m.settings` (geladen bei App-Start/Lobby-Öffnen). Externe config.yaml-Änderungen zwischen letztem Load und Write gehen verloren. Fix-Idee (Reviewer): vor dem Append/Write ein frisches `config.LoadSettings()` einholen (billiger Datei-Read) statt Model-Cache zu vertrauen — an BEIDEN Schreibstellen.
2. **Slug-Truncation ungetestet:** Lobby-Label `slug — Pfad` — dass bei knapper Breite der Slug (nicht der Pfad) erhalten bleibt, sichert nur die Code-Konstruktion, kein expliziter Test. Kleiner Inhalts-Test bei schmaler Breite ergänzen.

Akzeptanz:
- [ ] Beide Settings-Schreibstellen lesen frisch vor Write; Test mit extern geänderter config zwischen Load und Save
- [ ] Truncation-Test: schmale Breite → Slug sichtbar
- [ ] Test-Suite grün
