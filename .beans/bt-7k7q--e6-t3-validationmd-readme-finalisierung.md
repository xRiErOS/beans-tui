---
# bt-7k7q
title: E6 T3 — validation.md + README-Finalisierung
status: todo
type: task
created_at: 2026-07-15T14:00:59Z
updated_at: 2026-07-15T14:00:59Z
parent: bt-zk9p
blocked_by:
    - bt-wm4w
    - bt-9yvh
---

Ziel: docs/plans/v1-port/validation.md anlegen — Kopf/Quellen/Stand, US-Tabelle (14 Zeilen,
Evidenz aus T1/T2, NICHT blind aus dem Matrix-Entwurf kopiert), D-Codes-Tabelle (5 offene
PO-Entscheide: D01 esc-Detail-Fokus, D02 Sort-Indikator Backlog, D03 Upstream-ETag-Issue,
D04 VQA-I01 Footer-Wrap, D05 VQA-I02 Lobby-Pfad-Ellipsis — je mit Empfehlung, KEINE
Implementierung), Hygiene-Log, Smoke-Belege-Anhang. README.md Status-Abschnitt + Known
Issues aktualisieren.

Plan: docs/plans/v1-port/epic-E6-plan.md »Task 3«.

## Akzeptanz

[ ] validation.md Kopf: Kontext, Quellen, git rev-parse --short HEAD, Live-Testfunktions-
    Count (aus T1 Step 1).
[ ] US-Tabelle 14 Zeilen (ID | Titel | Status PASS/GAP | Evidenz-Anker | Kommentar), Werte
    aus T1-/T2-Commit-Bodies übernommen, jede Zeile mit konkretem Anker (Testname/Golden/
    Smoke-Verweis mit Datum/Commit).
[ ] D-Codes-Tabelle exakt 5 Zeilen (D01-D05 wie oben), Format Dxx|Hintergrund|Entscheidung
    (leer)|Empfehlung|Status (🟡 Unklar) — CLAUDE.md-Pflichtformat.
[ ] Hygiene-Log-Abschnitt: bt-aq5s B01 + bt-gzcu I01 Vorher/Nachher + Testverweis (nur
    falls in T1/T2 tatsächlich aktualisiert).
[ ] Smoke-Belege-Anhang: volle tmux-Tabellen aus T1 US-02-Cross-View + T1 US-07-Konflikt +
    T2 US-14-Repo-Wechsel (Format wie bt-7dfj Smoke-Matrix).
[ ] README.md Status-Abschnitt: "E6 ist fertig" statt "ist offen", Verweis auf
    validation.md ergänzt.
[ ] README.md Known Issues: VQA-I01/VQA-I02 ergänzt als "PO-Entscheid ausstehend, siehe
    validation.md D04/D05" (NICHT als gelöst markiert).
[ ] Commit docs(plan): validation.md + README E6-Stand.
