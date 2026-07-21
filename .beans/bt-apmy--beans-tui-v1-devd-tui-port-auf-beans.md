---
# bt-apmy
title: beans-tui v1 — devd-TUI-Port auf beans
status: completed
type: milestone
priority: high
tags:
    - smoke
    - accepted
created_at: 2026-07-14T18:32:41Z
updated_at: 2026-07-21T08:39:00Z
---

Ziel: Eigenständige TUI `bt`, die Look & Funktionalität der devd-TUI auf beans-Repos portiert — PO-Cockpit zum Ansehen, Bearbeiten und für den Review-Flow mit Agenten.

Quelle der Wahrheit: docs/plans/v1-port/design-spec.md · Plan: docs/plans/v1-port/implementation-plan.md

## Definition of Done
- [ ] Alle User Stories US-01…US-14 (design-spec §10) validiert erfüllt (Nachweise in docs/plans/v1-port/validation.md)
- [ ] `bt` startbar im beans-tui-Repo UND im lean-stack-Repo (Demo-Beleg)
- [ ] `command go test ./...` grün, README mit Install-Anleitung
- [ ] PO-Abnahme (Tag to-review gesetzt, Erik schließt)



## v1-Abnahme-Zusammenfassung (für PO)

Alle 7 Epen tragen Tag `to-review` (PO-Gate) — der Agent setzt bei
Epics/Milestones nie `completed`, das bleibt PO-Sache.

| Epic | bean | Titel | Status | Tag |
|---|---|---|---|---|
| E1 | `bt-blsy` | Foundation — Skeleton, Datenlayer, Theme, App-Shell | in-progress | to-review |
| E2 | `bt-aq5s` | Browse & Detail — Tree, Accordion, Suche, Filter, Backlog | in-progress | to-review |
| E3 | `bt-gzcu` | Mutationen — Forms, Picker, Editor, ETag-Handling | in-progress | to-review |
| E4 | `bt-tfqi` | Command-Center & Review-Cockpit | in-progress | to-review |
| E5 | `bt-5h4d` | Polish — Toast, Help, Yank, Maus, Settings, Lobby, Archiv | in-progress | to-review |
| E6 | `bt-zk9p` | Validierung & Release — alle US nachgewiesen | in-progress | to-review |
| E7 | `bt-heg9` | PO-Feedback R1 — Detail-UX + Typ/Status/Prio-Glyphen | todo | to-review |

**validation.md-Kernzahlen** (`docs/plans/v1-port/validation.md`, Stand
Commit `e716a3a`, 2026-07-15, 400 Testfunktionen):

- 14/14 User Stories validiert — **13× PASS**, **1× PARTIAL** (US-08:
  Tag-Sichtbarkeit im Tree/Detail fehlt, Filter-Facette funktioniert; Bug
  `bt-gdkx` medium, s. D01)
- **8 offene PO-Entscheidungen (D01–D08)** gesammelt zur Abnahme vorgelegt —
  keine blockiert v1 funktional, D01 hat direkten User-Story-Bezug und
  sollte vorrangig entschieden werden

**Entscheidungsvorlage:** die vollständige D-Codes-Tabelle (Hintergrund +
Empfehlung je Zeile) steht in `docs/plans/v1-port/validation.md` §5
("Entscheidungstabelle") — dient als Grundlage für die PO-Abnahme-Session.

**Demo-Beleg (Fremd-Repo):** `bt` direktstartbar gegen
`/Users/erik/Obsidian/tools/lean-stack` (`.beans.yml` vorhanden, kein
Lobby-Umweg), Tree zeigt die 83 lean-stack-Beans live — tmux-Beleg im
Commit-Body von bt-upyz (E6 Task 4).
