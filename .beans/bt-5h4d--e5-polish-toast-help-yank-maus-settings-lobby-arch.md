---
# bt-5h4d
title: E5 Polish — Toast, Help, Yank, Maus, Settings, Lobby, Archiv
status: completed
type: epic
priority: normal
tags:
    - to-review
created_at: 2026-07-14T18:33:30Z
updated_at: 2026-07-18T15:04:06Z
parent: bt-apmy
blocked_by:
    - bt-tfqi
---

Ziel: Toast-System, `?`-Help aus Keymap generiert, Yank (OSC52+nativ), Maus (Wheel/Klick/Doppelklick), Settings (~/.config/beans-tui/config.yaml + state.json LastRepo), Lobby mit Repo-Picker `p`, Archiv-Sicht (completed/scrapped togglebar).

Epos-Start-Ritual wie E2 (epic-E5-plan.md). Port-Quellen: overlay_show_toast.go, overlay_shortcuts.go, internal/clip, context.go, view_home.go, internal/config (devd cli-go).


## Übernahme aus E3 (für E5-Planung)
- [ ] Toast-Design muss Konflikt-Fall abdecken: ErrConflict-Meldung persistent (sticky) statt 1-Frame-Flash (E3-I01); Recovery-Tempfile-Pfad (deine Fassung: /tmp/…) muss lesbar bleiben
