---
# bt-apmy
title: beans-tui v1 — devd-TUI-Port auf beans
status: todo
type: milestone
priority: high
created_at: 2026-07-14T18:32:41Z
updated_at: 2026-07-14T18:32:41Z
---

Ziel: Eigenständige TUI `bt`, die Look & Funktionalität der devd-TUI auf beans-Repos portiert — PO-Cockpit zum Ansehen, Bearbeiten und für den Review-Flow mit Agenten.

Quelle der Wahrheit: docs/plans/v1-port/design-spec.md · Plan: docs/plans/v1-port/implementation-plan.md

## Definition of Done
- [ ] Alle User Stories US-01…US-14 (design-spec §10) validiert erfüllt (Nachweise in docs/plans/v1-port/validation.md)
- [ ] `bt` startbar im beans-tui-Repo UND im lean-stack-Repo (Demo-Beleg)
- [ ] `command go test ./...` grün, README mit Install-Anleitung
- [ ] PO-Abnahme (Tag to-review gesetzt, Erik schließt)
