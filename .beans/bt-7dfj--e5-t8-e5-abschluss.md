---
# bt-7dfj
title: E5 T8 — E5-Abschluss
status: todo
type: task
created_at: 2026-07-15T09:04:38Z
updated_at: 2026-07-15T09:04:38Z
parent: bt-5h4d
---

Ziel: E5-Abschluss-Ritual (implementation-plan.md »Epos-Abschluss«): Tests+
Build grün, beans-Pflege, README-Ergänzung (Settings/Lobby/Yank/Maus kurz
dokumentiert), Commit, SSTD-Pointer falls nötig, NSP-Auto-Handover für E6.

Plan: docs/plans/v1-port/epic-E5-plan.md »Task 8«.

## Akzeptanz
- [ ] `command go test ./...` grün (ohne -short, 2x hintereinander), `command
      go test ./... -race` grün, `command go build -o bin/bt .` ok, `command
      gofmt -l .` leer, `command go vet ./...` leer
- [ ] Alle Goldens (Chrome/Tree/TreeDeterministic/Backlog/
      BacklogDeterministic + etwaige neue) 2x grün
- [ ] tmux-Smoke im Scratch-Repo: Toast (Konflikt sticky bis Klick), Help-
      Overlay, Yank (Clipboard-Inhalt via `pbpaste`/OSC52-Capture
      verifiziert), Maus (Klick+Wheel+Doppelklick), Settings-Form (Editor/
      Accent live), Lobby+Repo-Wechsel (zwei Scratch-Repos), Archiv-Toggle
      -- als Beleg im Commit-Body/Bean-Body dokumentiert
- [ ] beans-Pflege: T1-T7-Task-beans auf `completed` (agent-abschließbar),
      bt-5h4d (Epic) bekommt Tag `to-review` (NICHT completed -- PO-Gate,
      implementation-plan.md »Epos-Abschluss«)
- [ ] README.md: Kurzabschnitt Settings-Pfad + Keymap-Ergänzung (p/y/?/
      Maus) falls README eine Keymap-Referenz führt (prüfen)
- [ ] docs/SSTD.md: Pointer-Update falls Worktree-Weiche/Referenzen sich
      geändert haben (voraussichtlich unverändert)
- [ ] Commit `docs: README + E5-Abschluss`
- [ ] Skill `ce-nsp-auto`: Handover-Prompt für E6 (Validierung & Release,
      bean bt-zk9p) erzeugen
