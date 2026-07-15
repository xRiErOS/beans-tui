---
# bt-upyz
title: E6 T4 — Release-Hygiene
status: in-progress
type: task
priority: normal
created_at: 2026-07-15T14:01:12Z
updated_at: 2026-07-15T19:47:06Z
parent: bt-zk9p
blocked_by:
    - bt-7k7q
---

Ziel: Demo-Nachweis bt im lean-stack-Repo, Milestone bt-apmy + Epic bt-zk9p auf Tag
to-review (PO-Gate, NICHT completed), E6-Task-beans (T1-T4) auf completed, CHANGELOG-
Entscheidung dokumentiert, Supervisor-Handover-Notiz für D07/KC-Update und
lean-stack-Verweis-bean lean-stack-5jqa (NICHT selbst ausgeführt — bewusst Supervisor-Sache
nach PO-Freigabe).

Plan: docs/plans/v1-port/epic-E6-plan.md »Task 4«.

## Akzeptanz

[ ] /Users/erik/Obsidian/tools/beans-tui/
    beans-tui-repository/bin/bt — tmux-Beleg (capture-pane) im Commit-Body: Tree zeigt
    lean-stack-Beans, Direktstart ohne Lobby (.beans.yml vorhanden).
[ ] CHANGELOG-Entscheidung dokumentiert (kein bestehendes CHANGELOG.md, kein Nutzer-
    Wunsch bekannt → YAGNI, ausgelassen — Begründung im Task-Body, nicht stillschweigend).
[ ] beans show bt-zk9p --json | jq .tags geprüft, dann beans update bt-zk9p --tag to-review.
[ ] beans update bt-apmy --tag to-review (Milestone, PO-Gate).
[ ] Task-beans bt-wm4w/bt-9yvh/bt-7k7q/dieser Task auf completed (agent-abschließbar) —
    bt-zk9p und bt-apmy NICHT completed.
[ ] Supervisor-Handover-Notiz im Bean-/Commit-Body vermerkt: (a) KC-Konzept
    po-immersion-beans-via-obsidian-bases-no-custom-tui via /okf um Supersede-Hinweis (D07,
    design-spec.md §14) ergänzen; (b) lean-stack-Verweis-bean lean-stack-5jqa schließen
    (dessen eigene Akzeptanz: "v1 abgenommen -> diesen bean schließen"). Beide EXPLIZIT
    nicht von diesem Task ausgeführt.
[ ] docs/SSTD.md: Pointer-Update nur falls nötig (voraussichtlich unverändert).
[ ] Commit docs(release): E6-Abschluss — Demo, Milestone to-review, Handover-Hinweise.
