---
# bt-7k7q
title: E6 T3 — validation.md + README-Finalisierung
status: completed
type: task
priority: normal
created_at: 2026-07-15T14:00:59Z
updated_at: 2026-07-15T19:46:07Z
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

[x] validation.md Kopf: Kontext, Quellen, git rev-parse --short HEAD, Live-Testfunktions-
    Count (aus T1 Step 1).
[x] US-Tabelle 14 Zeilen (ID | Titel | Status PASS/GAP | Evidenz-Anker | Kommentar), Werte
    aus T1-/T2-Commit-Bodies übernommen, jede Zeile mit konkretem Anker (Testname/Golden/
    Smoke-Verweis mit Datum/Commit).
[x] D-Codes-Tabelle (superseded durch Nachtrag unten: 8 Zeilen D01-D08 statt der hier
    ursprünglich skizzierten 5/D01-D05 — Format Dxx|Hintergrund|Entscheidung(leer)|
    Empfehlung|Status weiterhin CLAUDE.md-Pflichtformat, 🟡 Unklar durchgängig).
[x] Hygiene-Log-Abschnitt: bt-aq5s B01 + bt-gzcu I01 Vorher/Nachher + Testverweis (beide
    in T1/T2 tatsächlich aktualisiert, 🟣/🟡→🟢).
[x] Smoke-Belege-Anhang: volle tmux-Tabellen aus T1 US-02-Cross-View + T1 US-07-Konflikt +
    T2 US-14-Repo-Wechsel (Format wie bt-7dfj Smoke-Matrix) + I02-Nachtrag-Anhang (3
    frische Post-E7-Captures).
[x] README.md Status-Abschnitt: "E6 ist fertig" statt "ist offen", Verweis auf
    validation.md ergänzt.
[x] README.md Known Issues (superseded durch Nachtrag: statt VQA-I01/VQA-I02 jetzt
    Querverweis auf validation.md D01 (neue Tag-Sichtbarkeits-Bullet) + D04/D05/D06
    (bestehende I01/I02/D01-Bullets, je um Verweis ergänzt) — Known Issues NICHT
    dupliziert, nur referenziert.
[x] Commit docs(plan): validation.md + README E6-Stand → tatsächlicher Commit:
    docs: validation.md — v1-US-Abnahme (13 PASS, 1 PARTIAL) + README (2cf85dd).


## Nachträge aus Evidence-Review (2026-07-15, EVIDENCE_SOLID)

- I01: In validation.md bei US-04 die Automatiktests NAMENTLICH zitieren (TestPaletteActionsBeanContextFirst, TestDispatchPaletteBeanJumpsCursorAndSwitchesToBrowse) statt nur 'grün'.
- I02 (low-medium): US-12-Beleg-Kette dünner als Rest — 2-3 frische Post-E7-Captures ergänzen (Filter-Overlay, Palette, Settings-Form) und in validation.md referenzieren; VQA-Screenshots als 'vor E7' kennzeichnen.
- B01/bt-gdkx (US-08 Tags-Sichtbarkeit) als eigenen D-Punkt in der Entscheidungstabelle führen (Q05 beim PO offen: Meta-Zeile / Kopfblock / Tree-Suffix / Kombination — Supervisor-Empfehlung: Meta-Zeile + Tree-Suffix). US-08 in der Matrix als PARTIAL mit Verweis.



## Summary

`docs/plans/v1-port/validation.md` angelegt (committet, 2cf85dd — nicht git-ignored,
das offizielle v1-Abnahme-Artefakt): Kopf (Commit e716a3a Basis, 400 Live-Testfunktionen,
Methodik-Absatz mit Verweis auf die git-ignored Rohbeleg-Dateien e6-t1-evidence.md/
e6-t2-evidence.md), 14-zeilige US-Matrix (13 PASS + US-08 PARTIAL, jede Zeile mit
konkretem Test-/Golden-/Smoke-Anker, US-04 zitiert TestPaletteActionsBeanContextFirst/
TestDispatchPaletteBeanJumpsCursorAndSwitchesToBrowse namentlich wie im I01-Nachtrag
gefordert), Bugs-Tabelle (B01/bt-gdkx), Hygiene-Log (bt-aq5s B01, bt-gzcu I01 — beide
Vorher/Nachher + Testverweis), Smoke-Belege-Anhang (volle US-02/US-07/US-14-Tabellen +
neuer I02-Nachtrag-Unterabschnitt mit 3 frischen `tmux capture-pane -e`-Captures gegen
/tmp/bt-scratch-a: Filter-Overlay, Command-Palette, Settings-Form — schließt die laut
Evidence-Review dünne US-12-Beleg-Kette), und eine 8-zeilige Entscheidungstabelle
(D01-D08, ganz am Ende laut CLAUDE.md-Konvention) statt der ursprünglich im Plan
skizzierten 5 (D01 Tags-Sichtbarkeit/bt-gdkx mit Optionen a-d + Supervisor-Empfehlung
Meta-Zeile+Tree-Suffix, D02 Sort-Indikator, D03 esc-Detail-Fokus, D04 Header-Truncation,
D05 Overlay-Footer-Restatement, D06 Footer-Umfang, D07 Upstream-ETag, D08
Tag-Management-Page-Scope) — alle drei Nachträge aus dem Evidence-Review vollständig
eingearbeitet. bt-aq5s/bt-gzcu-Bodies auf weitere offene Vermerke geprüft: keine
zusätzlichen offenen Punkte über die bereits bekannten D02/D03/D07 hinaus gefunden.

VQA-I01/VQA-I02 (bean bt-zk9p, vor-E7-Supervisor-QA) bewusst NICHT als eigene D-Punkte
geführt (nicht Teil der 8 vom Evidence-Review vorgegebenen D-Codes) — stattdessen in der
US-12-Matrixzeile referenziert: VQA-I01 durch PF-11 entschärft/inhaltlich in D04
aufgegangen, VQA-I02 bleibt offen ohne eigenen D-Punkt (transparent vermerkt, nicht
stillschweigend fallengelassen).

README.md: Status-Abschnitt auf E6-fertig mit validation.md-Verweis; Known Issues um
eine neue Tag-Sichtbarkeits-Bullet (D01) ergänzt und die drei bestehenden E7-Bullets
(I01/I02/D01) um Querverweise auf validation.md D04/D05/D06 ergänzt — keine Duplizierung,
nur Referenzierung wie beauftragt.

Sanity: `command go vet ./...` leer, `command gofmt -l .` leer, `command go test ./...
-short` grün (alle Pakete, keine Code-Änderung in diesem Task). Working tree nach Commit
clean.

Commit: `docs: validation.md — v1-US-Abnahme (13 PASS, 1 PARTIAL) + README` (2cf85dd,
Refs bt-7k7q).
