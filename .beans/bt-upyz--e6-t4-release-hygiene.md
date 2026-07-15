---
# bt-upyz
title: E6 T4 — Release-Hygiene
status: completed
type: task
priority: normal
created_at: 2026-07-15T14:01:12Z
updated_at: 2026-07-15T19:53:55Z
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

[x] /Users/erik/Obsidian/tools/beans-tui/
    beans-tui-repository/bin/bt — tmux-Beleg (capture-pane) im Commit-Body: Tree zeigt
    lean-stack-Beans, Direktstart ohne Lobby (.beans.yml vorhanden).
[x] CHANGELOG-Entscheidung dokumentiert (kein bestehendes CHANGELOG.md, kein Nutzer-
    Wunsch bekannt → YAGNI, ausgelassen — Begründung im Task-Body, nicht stillschweigend).
[x] beans show bt-zk9p --json | jq .tags geprüft, dann beans update bt-zk9p --tag to-review.
[x] beans update bt-apmy --tag to-review (Milestone, PO-Gate).
[x] Task-beans bt-wm4w/bt-9yvh/bt-7k7q/dieser Task auf completed (agent-abschließbar) —
    bt-zk9p und bt-apmy NICHT completed.
[x] Supervisor-Handover-Notiz im Bean-/Commit-Body vermerkt: (a) KC-Konzept
    po-immersion-beans-via-obsidian-bases-no-custom-tui via /okf um Supersede-Hinweis (D07,
    design-spec.md §14) ergänzen; (b) lean-stack-Verweis-bean lean-stack-5jqa schließen
    (dessen eigene Akzeptanz: "v1 abgenommen -> diesen bean schließen"). Beide EXPLIZIT
    nicht von diesem Task ausgeführt.
[x] docs/SSTD.md: Pointer-Update nur falls nötig — geprüft, Worktree-Weiche/Referenzquellen
    unverändert, keine Änderung vorgenommen.
[x] Commit docs(release): E6-Abschluss — Demo, Milestone to-review, Handover-Hinweise.

## Summary

- CHANGELOG-Entscheidung: `ls CHANGELOG.md` bestätigt kein bestehendes CHANGELOG in
  diesem Repo, kein Nutzer-Wunsch bekannt → YAGNI, bewusst ausgelassen (dokumentiert,
  nicht stillschweigend übersprungen).
- Voll-Lauf (frisch, `-count=1`): `command go test ./...` grün (`internal/tui`
  136.158s, `internal/data` 2.890s, `cmd`/`config`/`theme` je <1.5s) ·
  `command go vet ./...` clean · `gofmt -l .` leer · Goldens
  `-run Golden -count=2 -v` → 10× PASS (5 Golden-Tests × 2 Läufe, deterministisch).
- beans-Endpflege: Epic `bt-zk9p` und Milestone `bt-apmy` je Tag `to-review` gesetzt
  (Status unverändert, NICHT `completed`). Milestone-Body per `--body-append` um
  `## v1-Abnahme-Zusammenfassung (für PO)` ergänzt: Epen-Tabelle E1–E7 (alle
  `to-review`), validation.md-Kernzahlen (14 US validiert, 13× PASS/1× PARTIAL,
  8 D-Entscheide D01–D08), Verweis auf die D-Tabelle in validation.md §5 als
  Entscheidungsvorlage.
- Task-beans T1–T4 (bt-wm4w/bt-9yvh/bt-7k7q/bt-upyz) alle `completed`.
- Scratch-Aufräumen: `/tmp/bt-scratch-{100,a,b,etag,home}` entfernt — vorher geprüft
  (`tmux list-panes -a`), dass keine offene tmux-Session darauf zeigt (nur `vqa`
  und `bt-e3t2-smoke` liefen, keine davon in einem `bt-scratch-*`-Pfad); `vqa`-Session
  des Supervisors unangetastet gelassen.

## Demo-Beleg

tmux-Session `btdemo` (200×50), `bin/bt` direktgestartet gegen
`/Users/erik/Obsidian/tools/lean-stack` (`.beans.yml` dort vorhanden → kein
Lobby-Umweg):

```
> lean-stack: Browse    ctrl+r:reload  ctrl+k:commands  p:repos  ?:help  esc:back  enter:open/confirm  q:quit
⌕ / search
▌▸ t E lean-stack-fnz8 Behavior-Surface-Angleichung: Skills/Age…
   t T lean-stack-5jqa beans-tui: devd-TUI-Port auf beans (exte…
 ▸ t E lean-stack-cih3 ADR-Modell schärfen: ADR = AOS-Heartbeat…
 ▸ t E lean-stack-b59y DevDash-Ablösung: Wissen→OKF · Arbeit→be…
 ▸ t E lean-stack-qur2 docs/architecture-Konvention festschreib…
 ▸ t E lean-stack-3s1k git-Standard tools-weit festschreiben
 ▸ t E lean-stack-a2ca init-project auf lean-stack reframen + d…
 ▸ t E lean-stack-o4c4 Konzeptionelle Fragen: offene Klärungen …
 ▸ t E lean-stack-n0ly lean-stack-Repo Klasse-B-konform machen …
   t T lean-stack-jv6w ce-beans-Skill: Review-Tag-Konvention al…
   d T lean-stack-58vh smoke
```

Read-only Smoke (KEINE Mutation an lean-stack-Beans):

| Segment | PASS/FAIL | Beleg |
|---|---|---|
| Direktstart (kein Lobby) | PASS | Breadcrumb `> lean-stack: Browse` sofort, `.beans.yml` upward-discovery |
| Navigation (`k`×2) | PASS | Cursor wandert `fnz8`→`cih3`, Detail-Pane live synchron |
| Detail (`tab`, `2`) | PASS | BODY-Sektion expandiert, Glamour-Markdown-Rendering korrekt |
| Suche (`/` `beans`) | PASS | Fuzzy-Filter reduziert Tree auf Treffer mit "beans" im Titel |
| `esc` verlässt Suche | PASS | Filter zurückgesetzt, volle Liste wieder sichtbar |
| `q`→`enter` Quit-Confirm | PASS | Modal "Quit? Really quit bt.", sauberer Exit zu zsh |

Nach Exit: `git status --porcelain .beans/` im lean-stack-Repo leer — keine Mutation.
lean-stack-Bean-Count zum Zeitpunkt dieses Laufs: 83 (8 epic, 1 feature, 1 milestone,
73 task) — höher als die im Plan-Text notierten "43" (Repo ist seit Plan-Erstellung
gewachsen, kein Fehler, hier informativ festgehalten).

## Handover

Offen für den Supervisor nach PO-Freigabe von E6 (bewusst NICHT Teil dieses Tasks):

(a) KC-Konzept `po-immersion-beans-via-obsidian-bases-no-custom-tui` via `/okf` um
einen Supersede-Hinweis ergänzen (D07, design-spec.md §14).
(b) lean-stack-Verweis-bean `lean-stack-5jqa` schließen (dessen eigene Akzeptanz:
"v1 abgenommen (alle US validiert) → diesen bean schließen").
(c) Offene PO-Entscheide D01–D08 in `docs/plans/v1-port/validation.md` §5 —
8 Punkte, keiner blockiert v1 funktional, D01 (Tag-Sichtbarkeit, US-08 PARTIAL) hat
direkten User-Story-Bezug und sollte vorrangig entschieden werden.
