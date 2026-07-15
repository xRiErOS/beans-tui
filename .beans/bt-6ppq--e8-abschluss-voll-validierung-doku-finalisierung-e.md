---
# bt-6ppq
title: E8-Abschluss (Voll-Validierung, Doku-Finalisierung, Epic to-review)
status: todo
type: task
priority: high
created_at: 2026-07-15T21:10:37Z
updated_at: 2026-07-15T21:10:37Z
parent: bt-ntoz
blocked_by:
    - bt-czpf
    - bt-duz7
    - bt-1u0t
    - bt-yqdy
    - bt-d8kc
---

E8 Task 9 (Abschluss) — letzter Task des Eposs. blocked_by ALLE fuenf Blatt-Tasks des Abhaengigkeitsgraphen (bt-czpf/T2, bt-duz7/T4, bt-1u0t/T5, bt-yqdy/T7, bt-d8kc/T8) -- deckt transitiv auch bt-e6q9/T1, bt-qbyq/T3, bt-y2iw/T6 ab (T4 und T7 haengen bereits an ihnen). Keine neuen Code-Aenderungen erwartet -- reine Validierung + Doku + beans-Pflege, mirrort epic-E7-plan.md Task 8s Muster.

## Step 1: Voller Regressionslauf

command go build -o bin/bt . ; command go test ./... -race ; command go test ./... -short (2x) ; command gofmt -l . (leer) ; command go vet ./... (leer). ZUSAETZLICH (wie E7/T8s "ERWEITERT"-Praezedenz): command go test ./... (voll, OHNE -short) 2x hintereinander gruen, alle Golden-Test-Funktionen (TestChromeGolden/TestTreeGolden/TestTreeGoldenDeterministic/TestBacklogGolden/TestBacklogGoldenDeterministic) mit -count=2 gruen. Voller Beleg im bean-Body dieses Tasks unter einem Abschnitt "Voll-Gate-Beleg".

## Step 2: US-08-Bestaetigung (bt-gdkx)

bt-gdkx (US-08: Tags nicht sichtbar in Tree/Detail) ist durch T1s D01-Umsetzung (Tags-Meta-Zeile) inhaltlich geloest. Live-Check: Bean mit Tags im Detail-Pane oeffnen, Meta-Zeile "tags:" MUSS die Tags anzeigen. bt-gdkx NICHT selbst schliessen (PO-Gate, Review-Flow §5) -- stattdessen Tag `to-review` auf bt-gdkx setzen (beans update bt-gdkx --tag to-review) MIT Body-Append, der auf den E8-Fix + diesen Abschluss-Task verweist, damit der PO beim naechsten Review sofort den Kontext hat.

## Step 3: validation.md-Hinweis

validation.md ist bereits (parallel zu diesem Epos, vom E8-Planner) auf §5 D01-D08 = entschieden aktualisiert (design-spec.md §15 PF-15/PF-16 als Quelle) -- hier NUR verifizieren, dass der Stand nach Abschluss von T1-T8 konsistent ist (kein Widerspruch zwischen "entschieden" und dem tatsaechlichen Code-Stand). Falls eine der Umsetzungen von der urspruenglichen D-Entscheidung abweicht (z.B. B06-Teal-Experiment vom PO abgelehnt wird -- siehe Step 4), validation.md NACHZIEHEN.

## Step 4: B06-Experiment-Sign-off (PFLICHT vor Epic-to-review)

B06 (Accordion-Header Teal-Experiment, T2/bt-czpf) ist EXPLIZIT als Experiment markiert -- der PO muss den Vorher/Nachher-Beleg aus T2s Commit-Body VOR der Epic-Freigabe sehen. Dieser Task fasst das NICHT eigenmaechtig zusammen/entscheidet nicht selbst -- stattdessen: den T2-Beleg (Screenshot oder Golden-Diff-Ausschnitt) im Epic-Bean bt-ntoz als eigenen Abschnitt "B06-Sign-off ausstehend" verlinken/zitieren, DAMIT der PO ihn beim to-review-Review sofort findet (nicht erst in T2s Task-bean suchen muss).

## Step 5: Epic-Ritual

beans update bt-ntoz --tag to-review (Epic — PO-Gate, Agent setzt NIE completed).
Task-beans dieses Eposs (T1-T8, bt-e6q9/bt-czpf/bt-qbyq/bt-duz7/bt-1u0t/bt-y2iw/bt-yqdy/bt-d8kc) auf completed, FALLS nicht schon durch die jeweiligen Implementer-Agents gesetzt (agent-abschliessbar nach gruenen Tests + Review-Durchlauf, Repo-CLAUDE.md-Konvention) -- hier verifizieren (beans list: alle 8 Status C), nicht blind setzen.

## Step 6: SSTD

docs/SSTD.md — Pointer-Update NUR falls sich Referenzen geaendert haben (z.B. falls design-spec.md/epic-E8-plan.md-Pfade dort erwartet werden). Pruefen, dokumentieren ob noetig oder nicht (wie E7/T8: "geprueft, unveraendert, kein Update noetig" ist eine gueltige, explizit zu nennende Aussage).

## Step 7: Commit

docs(release): E8-Abschluss — Epic to-review, US-08/bt-gdkx-Status, B06-Sign-off-Verweis. Explizite Pfade, kein git add -A.

## Akzeptanz-Checkliste

- [ ] Voller Testlauf (inkl. -race, 2x ohne -short, Golden-Funktionen mit -count=2) gruen, gofmt/vet leer
- [ ] bt-gdkx traegt Tag to-review mit Body-Verweis auf den Fix (bleibt selbst NICHT completed)
- [ ] bt-ntoz traegt Tag to-review, ist NICHT completed
- [ ] T1-T8 alle completed (verifiziert)
- [ ] B06-Sign-off-Verweis im Epic-Bean bt-ntoz auffindbar
- [ ] validation.md-Konsistenz verifiziert (kein Widerspruch Doku vs. Code)
