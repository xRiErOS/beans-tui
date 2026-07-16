---
# bt-7pk2
title: E9-Abschluss (Voll-Validierung, Epic to-review)
status: in-progress
type: task
priority: normal
created_at: 2026-07-16T06:46:07Z
updated_at: 2026-07-16T14:41:50Z
parent: bt-tct9
blocked_by:
    - bt-mtig
    - bt-z4b1
    - bt-2v38
    - bt-4mo9
    - bt-1e0t
    - bt-1vbp
    - bt-6bgn
---

E9 Task 9 — Abschluss. blocked_by: alle acht Implementierungs-Tasks dieses Eposs (deckt
transitiv B04→B06 und F01-Kernmechanik→F01-History mit ab). Keine Code-Änderungen erwartet
— reine Validierung + Doku + beans-Pflege (Muster epic-E8-plan.md Task 9/bean bt-6ppq).

## Schritte

1. Voller Regressionslauf (Build, `-race`, `-short` 2×, VOLL 2×, alle Golden-Funktionen mit
   `-count=2`, `gofmt -l .`/`command go vet ./...` leer) — Beleg im bean-Body unter
   „Voll-Gate-Beleg" (Muster bt-6ppq).
2. Design-Spec-Konsistenz: `design-spec.md` §15 PF-17 (bereits vom Planner geschrieben)
   gegen den TATSÄCHLICHEN Code-Stand nach T1-T8 gegenprüfen — insbesondere: `hangingIndentWrap`
   existiert wie beschrieben und wird von B04 UND B06 geteilt (nicht dupliziert);
   `fieldStrip` ist tatsächlich vollständig entfernt (kein toter Code); `openBodyEditor`
   existiert nicht mehr (durch `openBeanEditor` ersetzt, nicht daneben); `fullscreenMode`
   ist tatsächlich orthogonal zu `viewID` (kein neuer `viewID`-Wert hinzugekommen).
3. Bekannte-Grenze-Dokumentation (D01, Task 2): verifizieren dass der Commit-Body von
   Task 2 die "created_at/updated_at nicht editierbar"-Einschränkung tatsächlich trägt —
   falls nicht, hier nachtragen (in DIESEM Abschluss-bean, nicht rückwirkend in Task 2s
   bean, Append-only-Prinzip).
4. Epic-Ritual: `bt-tct9` bekommt Tag `to-review` (Agent setzt NIE `completed` — PO-Gate,
   design-spec.md §5 Review-Flow-Konvention). T1-T8 auf `completed` verifizieren (nicht
   selbst auf completed setzen, falls einzelne Implementer-Tasks das noch nicht getan
   haben — als Lücke im Abschluss-Bericht vermerken, nicht stillschweigend nachholen).
5. `docs/SSTD.md` — Pointer-Update nur falls nötig (prüfen + dokumentieren, Muster bt-6ppq
   Schritt 6).
6. `docs/plans/v1-port/epic-E9-plan.md` — Status-Kopfzeile/Task-Tabelle gegen den
   tatsächlichen Abschluss-Stand aktualisieren (analog wie epic-E8-plan.md nach Abschluss
   aussehen würde — falls dort kein expliziter "Status"-Header existiert, keinen neu
   erfinden, nur die Task-Tabelle mit dem finalen bean-Status abgleichen falls gewünscht).
7. Commit `docs(release): E9-Abschluss — Epic to-review, T1-T8-Status, Design-Spec-
   Konsistenz-Beleg`.

## Akzeptanz-Checkliste

- [ ] Voller Lauf grün (Build, -short 2×, VOLL 2×, -race, Golden -count=2, gofmt/vet leer)
- [ ] `bt-tct9` trägt `to-review`, NICHT `completed`
- [ ] T1-T8 alle `completed` (oder Lücken explizit im Abschluss-Bericht benannt)
- [ ] design-spec.md §15 PF-17 verifiziert konsistent zum tatsächlichen Code-Stand
- [ ] "Bekannte Grenze" (D01, created_at/updated_at) im Commit-Body von Task 2 ODER hier
      nachgetragen
- [ ] `docs/SSTD.md`-Konsistenz geprüft (Update nur falls nötig)
- [ ] Kein Golden-Drift unentdeckt (letzter Gegenbeleg-/Regenlauf grün, im Commit-Body
      referenziert)

## PRELUDE aus T8-Review (2026-07-16, Reviewer bt-1vbp, Verdict APPROVED)

Zwei optionale low-Findings — als erster eigener Commit erledigen (test-only, kein Produktionscode):

- **T8-F01 (low, view_fullscreen.go:94,112):** `if m.idx == nil { break }` in beiden History-Loops ohne dedizierte Testabdeckung. Regressionstest ergänzen: `m.idx = nil` bei `fullscreen == fullscreenDetail` + gefülltem navBack → HistoryBack muss No-Op sein (kein Panic, kein stiller Stack-Verlust ohne Navigation prüfen — Verhalten wie implementiert festschreiben).
- **T8-F02 (low, view_fullscreen_test.go TestHistoryBackForwardRoundTripReturnsToOriginalBean):** Roundtrip-Test prüft am Ende nur Stack-LÄNGEN + fullscreenBeanID. Zusätzliche Assertion auf exakte finale Slice-INHALTE von navBack/navForward (reflect.DeepEqual) ergänzen.

Zudem für die Design-Spec-Konsistenzprüfung (Kernauftrag T9): design-spec.md §15 F01-History sagt noch 'navBack/navForward werden beim Verlassen NICHT geleert' — implementiert ist per Supervisor-Entscheid das GEGENTEIL (jeder Vollbild-Exit leert beide Stacks, an allen 3 Choke-Points verifiziert). Spec-Wortlaut auf implementierten Stand bringen, Supervisor-Entscheid als Änderungsgrund vermerken.
