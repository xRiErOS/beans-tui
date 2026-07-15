---
# bt-7dfj
title: E5 T8 — E5-Abschluss
status: in-progress
type: task
priority: normal
created_at: 2026-07-15T09:04:38Z
updated_at: 2026-07-15T13:22:25Z
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


## Review-Finding aus T4 (I01, low — Reviewer 2026-07-15)

reviewClickRow (mouse.go) dupliziert die Zeilenstruktur von reviewQueueRows als privaten Zähl-Walk (bewusster Trade-off: Golden-Risiko null). Walk- und Render-Reihenfolge sind nur per Kommentar gekoppelt, nicht compile-time. Empfehlung für T8: Table-Test ergänzen, der beide Walks gegen dieselbe Fixture vergleicht (Zeilenindex je Bean identisch), statt Kommentar-Vertrauen.


## Notiz aus T6-Review (I02, low, kein v1-Handlungsbedarf)

In-flight repoMetricsCmd-Aufrufe eines früheren Lobby-Open laufen bei erneutem Öffnen weiter (kein Cancel-Guard) — redundante beans-list-Subprozesse, pfadgekeyt, keine Datenverwechslung. Bewusst akzeptiert für v1; bei vielen Repos später Kontext-Cancel. Nur im Abschluss-/Known-Issues-Text erwähnen.


## Notiz aus T7-Review (I01, low, vorbestehend seit E3)

Parent-Picker (EligibleParents, hierarchy.go) + Blocking-Picker (buildBlockingItems) filtern nicht nach showArchived/Status — archivierte/completed/scrapped Beans bleiben als Relationsziele wählbar. Keine Regression, Plan schweigt. In T8 entscheiden + dokumentieren: entweder bewusste Design-Entscheidung ('Picker zeigen alle gültigen Relationsziele') im README/Known-Issues festhalten ODER Mini-Fix (Status-Filter im Picker). Empfehlung: dokumentieren, kein Fix (v1-Scope).
