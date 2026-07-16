---
# bt-czpf
title: 'Accordion-Header: Chevron entfernen + Teal-Experiment'
status: in-progress
type: task
priority: normal
created_at: 2026-07-15T21:05:38Z
updated_at: 2026-07-16T00:27:18Z
parent: bt-ntoz
---

E8 Task 2 — deckt B05 (redundantes Chevron entfernen), B06 (Teal-Experiment fuer inaktive Header) aus bean bt-ntoz. Quelle: design-spec.md §15 PF-16. Ist-Code: internal/tui/accordion.go (renderAccordion), internal/theme/theme.go. Unabhaengig von allen anderen E8-Tasks (kein blocked_by), da nur accordion.go + theme.go beruehrt werden (kein Ueberlapp mit view_detail_bean.go/update.go, die T1/T3/T6 anfassen).

## B05 — Redundantes Chevron entfernen

renderAccordion (accordion.go) haengt heute an jeden Sektions-Header ein zusaetzliches `hint`-Suffix ("  ▾" offen / "  ▸" zu) NACH dem title an -- redundant, der Zustand ist am Sektionsinhalt selbst sichtbar (offen = Body sichtbar, geschlossen = kein Body). Fix: hint-Variable und ihre beiden isOpen-Zweige komplett entfernen, `header := marker+title` (ohne +hint) in beiden activeSec-Zweigen (aktiv/inaktiv).

## B06 — EXPERIMENT: inaktive Header-Farbe Grau->Teal

Inaktive Sektions-Header-Titel (theme.Muted = Hint-Grau #7c7c7c) sind farblich mit der Meta-Label-Spalte (ebenfalls theme.Muted, z.B. "status:") verwechselbar. PO will testweise Teal (theme.Teal, #8bd5ca, bereits in theme.go definiert -- KEIN neuer Hex-Wert noetig) fuer den INAKTIVEN Header-Titel sehen, EXPLIZIT als Experiment markiert -- PO entscheidet erst nach Screenshot-/Golden-Vergleich, ob es bleibt.

Repo-Regel beachten (CLAUDE.md tools-weit: "Theme-Token nur aus internal/theme/ -- keine Hex-Literale in Views"): KEIN inline lipgloss.NewStyle().Foreground(theme.Teal) in accordion.go -- stattdessen NEUEN benannten Style-Token in theme.go ergaenzen, z.B. `HeaderInactive = lipgloss.NewStyle().Foreground(Teal)` (neben dem bestehenden Header-Token), von accordion.go als `theme.HeaderInactive.Render(s.title)` konsumiert (ersetzt den bisherigen theme.Muted.Render(s.title)-Zweig fuer den INAKTIVEN Header -- die Meta-Label-Spalte selbst bleibt bei theme.Muted, NUR der Accordion-Section-Header wechselt).

## TDD-Schritte

1. Failing test: accordion_test.go NEU TestRenderAccordionInactiveHeaderNoChevronSuffix (Header-String enthaelt kein "▾"/"▸" mehr, egal ob offen/zu) + NEU TestRenderAccordionInactiveHeaderUsesTealNotMuted (Farb-Assertion gegen theme.HeaderInactive statt theme.Muted). theme_test.go NEU TestHeaderInactiveStyleIsTeal.
2. command go test ./internal/tui/... ./internal/theme/... -> FAIL.
3. Implementieren: theme.go (HeaderInactive-Token), accordion.go (hint entfernen, inaktiver Zweig auf theme.HeaderInactive umgestellt).
4. command go test ./internal/tui/... ./internal/theme/... -> PASS.
5. Golden-Regen (3 Goldens): command go build -o bin/bt ., dann command go test ./internal/tui/ -run "TestTreeGolden|TestBacklogGolden|TestChromeGolden" -update. ALLE DREI aendern sich voraussichtlich (Accordion-Header ist Teil von Tree- UND Backlog-Detail-Pane; chrome.golden pruefen, vermutlich unveraendert da Chrome() keine Accordion-Sektionen rendert -- explizit vermerken).
6. PFLICHT vor Abnahme (B06 ist EXPERIMENT): Vorher/Nachher-Beleg fuer den PO -- entweder ein tmux capture-pane-Screenshot (Detail-Pane mit >=2 Sektionen, eine aktiv eine inaktiv) ODER ein expliziter Golden-Diff-Ausschnitt (git diff internal/tui/testdata/tree.golden, die Header-Zeilen-Hunks) im Commit-Body/Task-Kommentar. Epic bt-ntoz NICHT auf completed setzen (ohnehin PO-Gate) -- aber dieser Teil-Fund braucht ZUSAETZLICH einen expliziten Hinweis im Abschluss-Task (T9), dass B06 PO-Sign-off vor v1-Freigabe braucht (koennte gegen theme.Muted zurückgerollt werden, falls PO das Experiment ablehnt -- EIN-Zeilen-Rollback dank des neuen Tokens).
7. command go test ./... -short gruen (2x), command go test ./... -race gruen, gofmt/vet leer.
8. Commit feat(tui): PF-16 Accordion-Header Chevron entfernt + Teal-Experiment (B05,B06) -- Body markiert B06 EXPLIZIT als "Experiment, PO-Sign-off ausstehend".

## Akzeptanz-Checkliste

- [ ] Kein "▾"/"▸"-Suffix mehr in irgendeinem Sektions-Header (offen oder zu)
- [ ] Inaktive Sektions-Header nutzen theme.HeaderInactive (Teal), nicht mehr theme.Muted
- [ ] Meta-Label-Spalte (z.B. "status:") bleibt UNVERAENDERT theme.Muted (B06 betrifft NUR Accordion-Section-Header, nicht die Meta-Feldliste)
- [ ] Vorher/Nachher-Beleg (Screenshot ODER Golden-Diff-Ausschnitt) im Commit-Body dokumentiert
- [ ] Goldens regeneriert, Vorher/Nachher je Datei vermerkt
- [ ] Voller Testlauf (inkl. -race) gruen, gofmt/vet leer
