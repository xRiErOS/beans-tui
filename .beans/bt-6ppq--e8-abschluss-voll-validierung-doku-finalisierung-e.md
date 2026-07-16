---
# bt-6ppq
title: E8-Abschluss (Voll-Validierung, Doku-Finalisierung, Epic to-review)
status: completed
type: task
priority: high
created_at: 2026-07-15T21:10:37Z
updated_at: 2026-07-16T03:47:42Z
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

- [x] Voller Testlauf (inkl. -race, 2x ohne -short, Golden-Funktionen mit -count=2) gruen, gofmt/vet leer
- [x] bt-gdkx traegt Tag to-review mit Body-Verweis auf den Fix (bleibt selbst NICHT completed)
- [x] bt-ntoz traegt Tag to-review, ist NICHT completed
- [x] T1-T8 alle completed (verifiziert)
- [x] B06-Sign-off-Verweis im Epic-Bean bt-ntoz auffindbar
- [x] validation.md-Konsistenz verifiziert (kein Widerspruch Doku vs. Code)


## Prelude aus E8-T3-Review (2026-07-16, Quelle: bt-qbyq-Review APPROVED, I01 non-blocking)

esc-Audit-Tabelle (Commit 6d0a9fe-Body + bt-qbyq) deckt nur die 6 PO-benannten Bereiche.
Reviewer fand 5 weitere esc-Sites — ALLE bereits 'eine Ebene zurück'-konform, kein Bug:
view_browse_backlog.go:307 (Backlog→Repo) · box_confirm_create.go:147 (Confirm→Form) ·
box_confirm_delete.go:123 (Abbruch) · forms_shared.go:164 (Form verwerfen, dokumentierte
Design-Entscheidung) · overlay_shortcuts.go:63 (Help schließen).
Auftrag Doku-Finalisierung: Audit-Tabelle im Epic-bean bt-ntoz (oder design-spec-Anhang)
um diese 5 Sites ergänzen — reine Nachvollziehbarkeit.


## Prelude aus E8-T6-Review (2026-07-16, non-blocking, Quelle: bt-y2iw-Review)

- I02: bt-y2iw-TDD-Schritt forderte dokumentierte Prüfung 'stat'-Fuzzy vs. 'set type'/'set priority' — Prüfung fehlte im bean-Body. Reviewer-Analyse: kein 'a' in beiden → Subsequence matcht nie, Ergebnis korrekt. Bei Doku-Finalisierung als Satz in bt-y2iw nachtragen.


## Prelude aus E8-T4-Re-Review (2026-07-16, non-blocking, Quelle: bt-duz7-Review R2 APPROVED)

- I01 (low): TestDetailClickKeyDisjointNumberSpaces (mouse_test.go) ist selbst-referenziell — prüft gegen die Konstante detailClickKeyBase statt gegen hartes Literal 1000; gegen Base-Mutation wertlos (Verhaltenstest TestMouseDetailClickTreeClickIndexDoesNotAliasFieldClickKey deckt ab). Bei nächster Berührung von mouse_test.go Literal einsetzen — Abschluss-Task: NICHT eigens fixen, nur in LESSONS-LEARNED als Muster 'selbst-referenzieller Pin-Test' aufnehmen.
- I02 (dokumentarisch): bewusste Asymmetrie BODY-Doppelklick (zeitfenstergebunden) vs. enter-auf-BODY (zeitfensterlos) — PO-Wortlaut scopet zeitfensterlosen Zweitklick auf Felder. Steht im Doc-Kommentar mouseDetailClick.


## Prelude aus E8-T7-Review (2026-07-16, non-blocking, Quelle: bt-yqdy-Review APPROVED)

- I01 (Doku): bt-yqdy-Summary referenziert nicht-existente Datei 'search.go' — /-Suche-Logik liegt in view_browse_repo.go/update.go/messages.go/types.go, Tests in search_test.go. Bei Doku-Finalisierung im bean-Body korrigieren.


## Prelude aus E8-T8-Review (2026-07-16, Quelle: bt-d8kc-Review APPROVED)

- Q-PO (offen, an PO im Chat): Backlog-Footer braucht bei 80 Spalten 3 Zeilen (D06 erlaubt 2). Reviewer-Messung: Klartext 161 Zeichen > 2×78; Umschlagpunkt exakt 82 Spalten. Spec-konformer Fix ohne PO unmöglich (Q06-Liste wortgesperrt, Sort-Entzug gegen Planner-ERRATUM). Optionen für PO: (a) 3 Zeilen bei <82 Spalten akzeptieren, (b) Sort-Footer-Eintrag streichen (Suffix in Suchzeile trägt die Info bereits), (c) Wortkürzungen freigeben. NICHT selbst entscheiden — im Abschluss-Report als offene Frage ausweisen.
- I01 (low): mouse_test.go clickPaneGeometry-Tests nur Width:100 — 3-Zeilen-Footer-Fall (80 Sp. Backlog) ohne automatisierte Abdeckung. Als kleinen Regressionstest im Abschluss-Task ergänzen (footH-Dynamik pinnen).
- I02 (low, Doku): footer()/breadcrumb() erwarten vorgefärbten renderBindings()-Output (kein Dim-Wash mehr) — stille Konvention, nur Doc-Kommentar. In LESSONS-LEARNED als Falle notieren.



## Summary (2026-07-16, E8-Abschluss)

Alle 6 Arbeitspakete abgearbeitet: (1) Voll-Validierung gruen (Detail
unten). (2) I01-Regressionstest
`TestDetailClickBacklogThreeLineFooterAt80Cols` (mouse_test.go, Commit
`3c084d2`) — pinnt die footH-Dynamik am realen 3-Zeilen-Backlog-Footer
bei 80 Spalten via render-geerdetem Boundary-Paar (letzte Body-Zeile
trifft, Pane-Bottom-Border/Footer-Zeile nicht); RED via gezielter
Mutation des detailClickRow-Chrome-Branches bewiesen. (3)
Doku-Finalisierung (Commit `7dbab76`): esc-Audit-Vervollstaendigung im
Epic (5 Zusatz-Sites, alle konform, Zeilendrift-ERRATUM 307→333) ·
'stat'-Fuzzy-Pruefbeleg an bt-y2iw · search.go-Korrektur an bt-yqdy
(ERRATUM: Auftrag ordnete sie bt-y2iw zu, der Fehler steht in bt-yqdys
Summary) · validation.md §7 "E8-Umsetzung" (D01-D08 + B01-B14 mit
Commits) + US-08 auf "PASS (pending PO-Sichtpruefung, bt-gdkx)". (4)
docs/LESSONS-LEARNED.md angelegt (7 Eintraege, jeder mit
Forward-Guard; Commit `1a6bb4f`, inkl. neuer CLAUDE.md-Guard-Zeile
"Footer-/Wrap-Aenderungen → Smoke bei 80 Spalten"). (5) beans-Pflege:
bt-gdkx Aufloesung + Tag to-review (Status unveraendert, live
verifiziert: `▷ tags: ● to-review` im Detail-Pane von bt-apmy) · Epic
bt-ntoz Tag to-review + Sektionen "B06-Sign-off ausstehend" und
"E8-Abschluss" · T1-T8 alle completed verifiziert. (6) Abschluss:
Checkboxen, Pflicht-Sektionen, completed, finaler Commit.

SSTD (Step 6): geprueft, unveraendert — docs/SSTD.md verweist auf
design-spec.md/implementation-plan.md/`beans list --ready`, alle
Referenzen weiterhin gueltig, kein Update noetig.

## Validierungs-Output (Voll-Gate-Beleg, 2026-07-16)

- `command go build -o bin/bt .` — clean
- `command go vet ./...` — leer · `gofmt -l .` — leer
- `command go test ./... -count=1` ZWEIMAL frisch: Lauf 1 `ok
  beans-tui/internal/tui 137.484s` (total 2:17.84), Lauf 2 `ok
  beans-tui/internal/tui 137.163s` (total 2:17.54) — alle Pakete
  (cmd/config/data/theme/tui) beide Male gruen; 462 Testfunktionen
  repo-weit (390 in internal/tui, inkl. neuem I01-Test)
- Goldens byte-stabil zwischen den Laeufen: tree/backlog/chrome.golden
  `diff -q` IDENTICAL gegen Vor-Lauf-Kopie, `git status
  internal/tui/testdata/` leer
- `command go test ./... -race -count=1` — gruen (`ok
  beans-tui/internal/tui 140.050s`, total 2:20.70), KEINE DATA RACE
- Golden-Funktionen `-count=2`: TestChromeGolden/TestTreeGolden/
  TestTreeGoldenDeterministic/TestBacklogGolden/
  TestBacklogGoldenDeterministic — alle 2x PASS
- Nach dem I01-Test ein weiterer voller Lauf vor dessen Commit: `ok
  beans-tui/internal/tui 137.399s` (total 2:17.77), gruen

I01-Test RED→GREEN: Mutation (detailClickRow-Chrome-Branch auf
browseRepoChrome gezwungen) → `mouse_test.go:825: click below the pane
(Y=23) must not resolve to a Detail hit (footH under-counts the 3-line
footer)` FAIL → Revert (mouse.go byte-identisch, leerer git diff) →
PASS.

## Deviations/ERRATA

1. **ERRATUM (Zeilendrift, T3-Prelude):** esc-Site
   view_browse_backlog.go:307 liegt nach den T8-Edits auf Zeile 333
   (keys.Backlog/keys.Back-Case) — Site konform, nur Zeilennummer
   gedriftet; die 4 uebrigen Prelude-Zeilennummern stimmen exakt.
2. **ERRATUM (Zuordnung, Auftrags-Paket 3b):** die search.go-Korrektur
   war als "bt-y2iw-Korrektur" gruppiert — der Fehler steht in bt-yqdys
   Summary (so benennt es auch das T7-Prelude). An bt-yqdy geroutet,
   in dessen Nachtrag dokumentiert.
3. **Layout-Detail (I01-Test):** die Status-Zeile rendert UNTER dem
   Footer, nicht dazwischen — die Relation ist footerY = originY +
   bodyH + 2 (Border+Divider), nicht +3; per Render-Probe verifiziert,
   Test entsprechend geerdet.
4. **Erste RED-Mutation verworfen:** footH-Hartkodierung INNERHALB von
   clickPaneGeometry ist als RED-Beweis untauglich — View und Klick-Pfad
   teilen die Funktion (Single Source), beide verschieben sich
   konsistent. Der Beweis nutzt stattdessen die einzig real
   desynchronisierbare Stelle (detailClickRow-Chrome-Branch) — genau
   die Bug-Klasse, die der Test pinnt.
5. **Zusatz ueber den Wortlaut hinaus (begruendet):** CLAUDE.md-Zeile
   "Footer-/Wrap-Aenderungen → tmux-Smoke bei 80 Spalten" ergaenzt —
   LESSONS-LEARNED-Eintrag 4 verlangt einen Guard mit Zaehnen; ohne die
   Zeile waere der zitierte CLAUDE.md-Guard fiktiv gewesen.

## Handover

- **PO-Review-Paket (Epic bt-ntoz, Tag to-review):** (1) B06-Sign-off —
  Beleg docs/plans/v1-port/b06-experiment/; Rollback = 1 Zeile,
  D06-Farben entkoppelt (gegengeprueft). (2) Q-PO Backlog-Footer:
  3 Zeilen bei <82 Spalten — Optionen (a) akzeptieren / (b)
  Sort-Eintrag streichen (Suchzeilen-Suffix traegt die Info) / (c)
  Wortkuerzungen; bewusst NICHT entschieden. (3) US-08-Sichtpruefung
  bt-gdkx (Tag to-review; live vorverifiziert). (4) Epic-Abnahme.
- **Nach v1-Abnahme:** D07/T03 Upstream-ETag-Issue-Entwurf (hmans/
  beans), POST nur mit PO-Freigabe — Minimal-Repro-Beleg im Epic-bean
  ("D07-Repro-Beleg").
- **Kleinkram fuer die naechste mouse_test.go-Beruehrung:**
  TestDetailClickKeyDisjointNumberSpaces auf Literal 1000 umstellen
  (selbst-referenzieller Pin-Test, LESSONS-LEARNED Eintrag 5).
- **Einstiegspunkte:** validation.md §7 (Umsetzungsstand) ·
  docs/LESSONS-LEARNED.md (Muster/Guards) · Epic-bean bt-ntoz
  (E8-Abschluss-Sektion mit Commit-Spanne).
