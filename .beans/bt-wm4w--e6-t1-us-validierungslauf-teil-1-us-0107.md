---
# bt-wm4w
title: E6 T1 — US-Validierungslauf Teil 1 (US-01..07)
status: completed
type: task
priority: normal
created_at: 2026-07-15T14:00:26Z
updated_at: 2026-07-15T19:12:55Z
parent: bt-zk9p
blocked_by:
    - bt-heg9
---

Ziel: Für US-01..07 (design-spec.md §10) je einen dokumentierten Validierungs-Nachweis
erbringen — automatisierte Testbestätigung PLUS Smoke, wo der Matrix-Entwurf eine Lücke
benennt. Kern: US-01 Performance <2s @100 beans MESSBAR machen (Test + tmux-Wanduhr),
US-02 EIN durchgehender Tastatur-Only-Cross-View-Smoke, US-07 ein echter Zwei-Prozess-
ETag-Konflikt (bisher nur simuliert).

Plan: docs/plans/v1-port/epic-E6-plan.md »Task 1«.

## Akzeptanz

[x] Baseline: command go test ./... (2x ohne -short) grün, -race grün, go build -o bin/bt . ok,
    gofmt -l . leer, go vet ./... leer. Live-Testfunktions-Count notiert
    (grep -rhoP "^func Test\w+" internal/ cmd/ | sort -u | wc -l).
[x] internal/data/testrepo_test.go: newTestRepoN(t, n int) string (1 Milestone + 10 Epics
    + n-11 Tasks, gleiches Frontmatter-Schema wie newTestRepo).
[x] internal/data/performance_test.go: TestListPerformanceAt100Beans — Client.List() gegen
    newTestRepoN(t, 100) unter time.Since()-Messung, t.Logf der Dauer, Fatalf bei >2s.
[x] Real-Terminal-Wanduhr-Beleg: /tmp/bt-scratch-100 über echtes beans-CLI (init + 1
    Milestone + 10 Epics + 89 Tasks), tmux capture-pane-Polling-Schleife misst wall_ms
    zwischen Start und erstem sichtbaren Tree-Inhalt — Wert <2000ms dokumentiert.
[x] .beans.yml-Discovery-Fehlermeldung ("kein Repo gefunden") einmal real zitiert.
[x] US-02: EIN tmux-Durchlauf Tree→Detail(1-4)→Backlog(b)→Palette(ctrl+k)→Review-Cockpit(R)
    →Help(?)→Settings→Lobby(p), Segment-für-Segment PASS/FAIL-Tabelle. (ERRATUM: Review-
    Cockpit existiert nicht mehr, PF-14/E7 — Segment gestrichen, s. Evidence-Datei)
[x] US-03: Markdown-Fixture (Codeblock/Liste/Link) via beans update --body-file auf einen
    Bean im Smoke geschrieben, Glamour-Rendering im Detail-Accordion visuell geprüft.
[x] US-04: ctrl+k-Palette auf neue T6/T7-Einträge (Repo-Picker/Archiv-Toggle) geprüft,
    Befund exakt dokumentiert (auch falls architektonisch nicht als Palette-Item vorgesehen).
[x] US-05: Bleve-Feldsyntax (z.B. status:in-progress) einmal real gegen dieses Repo
    ausprobiert, Ergebnis notiert.
[x] US-06: c-Formular → Confirm → Cursor-auf-neuem-Bean visuell bestätigt (capture-pane).
[x] US-07: zwei tmux-Panes gegen dasselbe Scratch-Repo — Pane A ctrl+e (ETag eingefroren),
    Pane B rohes beans update (bumpt ETag), Pane A speichert → echter ErrConflict, Toast
    sticky + Recovery-Tempfile-Pfad belegt.
[x] Hygiene bt-aq5s B01 (Filter-Reset Backlog): NUR bei grünem
    TestKeyBacklogFilterClearResetsFacets Status-Zeile 🟣→🟢 + Testverweis aktualisiert
    (beans update bt-aq5s --body-replace-old/--body-replace-new).
[x] command go test ./... -short grün, gofmt/vet leer.
[x] Commit test(data): US-01..07 Validierungsbelege (E6 Task 1) — Body enthält alle
    Messwerte/Tabellen für T3.

## Summary

Alle US-01..07 validiert, ausschließlich PASS-Ergebnisse (kein FAIL/PARTIAL). Neuer
Testcode: `internal/data/testrepo_test.go::newTestRepoN` + `internal/data/performance_test.go::TestListPerformanceAt100Beans`
(automatisierter Performance-Beleg). Zwei unabhängige US-01-Belege (Go-Test ~35ms,
tmux-Wanduhr ~415ms) weit unter 2s-Kriterium. EIN großer tmux-Durchlauf
(read-only) deckt US-02/03(teilweise)/04/05 gegen dieses Repo ab; Mutationen
(US-03-Markdown-Write, US-06-Create, US-07-Edits+ETag-Konflikt) liefen bewusst
gegen `/tmp/bt-scratch-100`/`/tmp/bt-scratch-etag` statt gegen dieses Repo (Schutz
der dogfooding-Daten, deckt sich mit Supervisor-Vorgabe „echte Mutationen im
Scratch-Repo"). US-07 liefert erstmals einen ECHTEN Zwei-Prozess-ETag-Konflikt
(nicht simuliert): Toast, Sticky-Verhalten, Recovery-Tempfile alle live verifiziert.
Ein Bleve-Scope-Fund (US-05: `status:`-Feldsyntax von der beans-CLI selbst nicht
unterstützt, nur `slug`/`title`/`body`) dokumentiert, kein Bug. Volle Belege inkl.
aller Tabellen/Capture-Auszüge: `docs/_free-notes/e6-t1-evidence.md` (git-ignored,
Input für T3). Keine Bugs gefunden. bt-aq5s B01 als gelöst bestätigt (Hygiene).

## Ergebnis-Übersicht

| US | Titel | Ergebnis | Kurzbeleg |
|---|---|---|---|
| US-01 | `bt` starten, sofort Projektbaum | PASS | List() 32-36ms (3x); tmux-Wanduhr 411-419ms (3x); Discovery+Fehlermeldung zitiert |
| US-02 | Navigation ohne Maus | PASS | 1 tmux-Durchlauf, alle Segmente PASS (Review-Cockpit-Segment entfällt, PF-14) |
| US-03 | Bean-Details vollständig | PASS | Glamour-Stichprobe (Codeblock/Liste/Link) korrekt gerendert |
| US-04 | Command-Palette | PASS | Repo-Picker-Action bestätigt; Archiv-Toggle korrekt KEINE Palette-Action (lebt im Filter-Facetten-Overlay) |
| US-05 | Suche/Filter | PASS | lokaler Substring + `title:`-Bleve-Feldsyntax PASS; `status:`-Feldsyntax 0 Treffer (beans-CLI-Scope-Fund, kein Bug) |
| US-06 | Bean anlegen | PASS | Confirm-Gate + Cursor-auf-neuem-Bean visuell + CLI-Frontmatter verifiziert |
| US-07 | Feld-Edit constrained | PASS | Status/Priority/Title/Tags via neue Enter-Kaskade (PF-5) verifiziert; ECHTER Zwei-Prozess-ETag-Konflikt (Toast sticky, Recovery-Tempfile-Inhalt exakt geprüft) |

## ERRATA

- **Review-Cockpit (Plan vs. Ist, PF-14):** `epic-E6-plan.md`/Matrix-Entwurf
  (vor-E7) nennen "Review-Cockpit (`R`)" als US-02-Segment — existiert nicht mehr
  (PF-14, §15: vollständig entfernt). Segment ersatzlos gestrichen, Settings
  (nur via Palette) stattdessen geprüft.
- **`date +%s%3N` (Plan-Skript, Step 3):** GNU-`date`-Syntax, auf macOS/BSD-`date`
  nicht unterstützt (`bad math expression`, kein `%N`). Ersetzt durch
  `python3 -c "import time; print(int(time.time()*1000))"`. Relevant für T2, falls
  weitere Wanduhr-Messungen folgen.
- **Mutations-Repo-Wahl:** Plan Step 5 empfiehlt EINEN gebündelten tmux-Durchlauf
  gegen DIESES Repo inkl. aller Mutationen (US-03/06/07). Task-Auftrag verlangt
  „echte Mutationen im Scratch-Repo". Entscheidung: gesplittet — Lese-Navigation
  gegen dieses Repo, Mutationen gegen `/tmp/bt-scratch-100`/`/tmp/bt-scratch-etag`.
  Schützt die dogfooding-`.beans/`-Daten, kein Kriterium wurde dadurch verwässert.

## Notes for T2

- Review-Cockpit-ERRATUM gilt genauso für T2s US-08-Bestätigungslauf — falls
  irgendwo noch "Review-Cockpit"-View erwartet wird: existiert nicht mehr.
- `date +%s%3N`-ERRATUM: bei eigenen Wanduhr-Messungen (US-14-Smoke) den
  Python-Workaround verwenden statt GNU-`date`-Syntax.
- `/tmp/bt-scratch-100` (100-Bean-Fixture inkl. Markdown-/Create-/Edit-Mutationen
  aus T1) und `/tmp/bt-scratch-etag` bleiben bestehen (nicht aufgeräumt) — T2 kann
  wiederverwenden oder neu aufbauen.
- Volle Rohbelege (Tabellen, Capture-Auszüge, exakte Toast-Texte) in
  `docs/_free-notes/e6-t1-evidence.md` (git-ignored) — direkt für `validation.md`
  (T3) übernehmbar.
