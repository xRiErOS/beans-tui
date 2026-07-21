---
# bt-zk9p
title: E6 Validierung & Release — alle US nachgewiesen
status: completed
type: epic
priority: high
tags:
    - accepted
created_at: 2026-07-14T18:33:30Z
updated_at: 2026-07-21T08:39:00Z
parent: bt-apmy
blocked_by:
    - bt-5h4d
---

Ziel: Jede User Story US-01…US-14 validiert (Test oder tmux-Beleg) → docs/plans/v1-port/validation.md; Lücken als bug-beans fixen; README komplett; Demo `bt` im lean-stack-Repo; Milestone-Tag to-review; KC-Konzept po-immersion (D07) via /okf aktualisieren; lean-stack-Verweis-bean pflegen.

Epos-Start-Ritual wie E2 (epic-E6-plan.md). US-Tabelle: design-spec §10.


## Visuelle QA durch Supervisor (2026-07-15, vor E6-Start)

16 Screenshots (echtes Terminal.app-Fenster, screencapture, visuell geprüft): docs/_free-notes/vqa-2026-07-15/. Ergebnis: bestanden — devd-Look (US-12) bestätigt, Overlay-Compositing sauber, Live-Reload (US-10), Yank-Toast (US-11), Lobby+Repo-Wechsel (US-14), Archiv-Default, Settings-Prefill alle visuell belegt. Kein Bug.

2 kosmetische Findings für E6-Bewertung:
- VQA-I01 (low): Footer-Keymap bricht bei 110 Spalten in 2. Zeile um ('e:Edit in / $EDITOR') — prüfen ob devd identisch; ggf. kürzen oder akzeptieren+dokumentieren.
- VQA-I02 (low): Lobby kürzt lange Repo-Pfade rechtsseitig (Basename unsichtbar bei tiefen Pfaden) — Ellipsis links wäre lesbarer; akzeptieren oder Mini-Fix, PO-Entscheid.
