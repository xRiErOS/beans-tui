---
# bt-wm4w
title: E6 T1 — US-Validierungslauf Teil 1 (US-01..07)
status: in-progress
type: task
priority: normal
created_at: 2026-07-15T14:00:26Z
updated_at: 2026-07-15T18:51:27Z
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

[ ] Baseline: command go test ./... (2x ohne -short) grün, -race grün, go build -o bin/bt . ok,
    gofmt -l . leer, go vet ./... leer. Live-Testfunktions-Count notiert
    (grep -rhoP "^func Test\w+" internal/ cmd/ | sort -u | wc -l).
[ ] internal/data/testrepo_test.go: newTestRepoN(t, n int) string (1 Milestone + 10 Epics
    + n-11 Tasks, gleiches Frontmatter-Schema wie newTestRepo).
[ ] internal/data/performance_test.go: TestListPerformanceAt100Beans — Client.List() gegen
    newTestRepoN(t, 100) unter time.Since()-Messung, t.Logf der Dauer, Fatalf bei >2s.
[ ] Real-Terminal-Wanduhr-Beleg: /tmp/bt-scratch-100 über echtes beans-CLI (init + 1
    Milestone + 10 Epics + 89 Tasks), tmux capture-pane-Polling-Schleife misst wall_ms
    zwischen Start und erstem sichtbaren Tree-Inhalt — Wert <2000ms dokumentiert.
[ ] .beans.yml-Discovery-Fehlermeldung ("kein Repo gefunden") einmal real zitiert.
[ ] US-02: EIN tmux-Durchlauf Tree→Detail(1-4)→Backlog(b)→Palette(ctrl+k)→Review-Cockpit(R)
    →Help(?)→Settings→Lobby(p), Segment-für-Segment PASS/FAIL-Tabelle.
[ ] US-03: Markdown-Fixture (Codeblock/Liste/Link) via beans update --body-file auf einen
    Bean im Smoke geschrieben, Glamour-Rendering im Detail-Accordion visuell geprüft.
[ ] US-04: ctrl+k-Palette auf neue T6/T7-Einträge (Repo-Picker/Archiv-Toggle) geprüft,
    Befund exakt dokumentiert (auch falls architektonisch nicht als Palette-Item vorgesehen).
[ ] US-05: Bleve-Feldsyntax (z.B. status:in-progress) einmal real gegen dieses Repo
    ausprobiert, Ergebnis notiert.
[ ] US-06: c-Formular → Confirm → Cursor-auf-neuem-Bean visuell bestätigt (capture-pane).
[ ] US-07: zwei tmux-Panes gegen dasselbe Scratch-Repo — Pane A ctrl+e (ETag eingefroren),
    Pane B rohes beans update (bumpt ETag), Pane A speichert → echter ErrConflict, Toast
    sticky + Recovery-Tempfile-Pfad belegt.
[ ] Hygiene bt-aq5s B01 (Filter-Reset Backlog): NUR bei grünem
    TestKeyBacklogFilterClearResetsFacets Status-Zeile 🟣→🟢 + Testverweis aktualisiert
    (beans update bt-aq5s --body-replace-old/--body-replace-new).
[ ] command go test ./... -short grün, gofmt/vet leer.
[ ] Commit test(data): US-01..07 Validierungsbelege (E6 Task 1) — Body enthält alle
    Messwerte/Tabellen für T3.
