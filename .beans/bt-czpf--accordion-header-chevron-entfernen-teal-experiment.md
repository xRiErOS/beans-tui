---
# bt-czpf
title: 'Accordion-Header: Chevron entfernen + Teal-Experiment'
status: completed
type: task
priority: normal
created_at: 2026-07-15T21:05:38Z
updated_at: 2026-07-16T00:57:02Z
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

- [x] Kein "▾"/"▸"-Suffix mehr in irgendeinem Sektions-Header (offen oder zu)
- [x] Inaktive Sektions-Header nutzen theme.HeaderInactive (Teal), nicht mehr theme.Muted
- [x] Meta-Label-Spalte (z.B. "status:") bleibt UNVERAENDERT theme.Muted (B06 betrifft NUR Accordion-Section-Header, nicht die Meta-Feldliste)
- [x] Vorher/Nachher-Beleg (Screenshot ODER Golden-Diff-Ausschnitt) im Commit-Body dokumentiert
- [x] Goldens regeneriert, Vorher/Nachher je Datei vermerkt
- [x] Voller Testlauf (inkl. -race) gruen, gofmt/vet leer

## Summary (2026-07-16)

B05 (Chevron-Suffix `"  ▾"`/`"  ▸"` entfernt) und B06 (EXPERIMENT: inaktive/
geschlossene Section-Header-Titel Grau->Teal via neuem `theme.HeaderInactive`)
umgesetzt, in ZWEI getrennten Commits (B06 einzeln revertierbar). Scope exakt
wie geplant: `internal/tui/accordion.go` + `internal/theme/theme.go` (+ deren
Tests). Kein Beruehrungspunkt mit `view_detail_bean.go`/`update.go`/`mouse.go`
-- `detailClickRow`s Zeilen-basierte Geometrie (mouse.go) zaehlt Render-Zeilen,
nicht Glyphen/Spalten, daher unveraendert (Maus-Regressionssuite komplett
gruen, s. Smoke).

## Test-Output (RED->GREEN je Haeppchen)

Beide neuen Tests (B05 `TestRenderAccordionNoChevronSuffix` + B06
`TestRenderAccordionClosedHeaderUsesTealNotMuted`/`TestHeaderInactiveStyleIsTeal`)
wurden in EINER Schreibrunde angelegt, bevor irgendein Implementierungscode
geaendert war -- die erste RED-Messung ist daher ein kombinierter
Compile-Fehler (der B06-Test zieht `theme.HeaderInactive` als Import, was
den ganzen Testbinary-Build blockiert, nicht nur den B06-Test):

```
internal/theme/theme_test.go:188:12: undefined: HeaderInactive
internal/tui/accordion_test.go:91:20: undefined: theme.HeaderInactive
FAIL	beans-tui/internal/tui [build failed]
FAIL	beans-tui/internal/theme [build failed]
```

GREEN nach `theme.go`-Token + `accordion.go`-Umstellung (hint-Entfernung UND
Teal-Zweig zusammen implementiert):
```
--- PASS: TestRenderAccordionExclusiveOpen (0.00s)
--- PASS: TestRenderAccordionNoChevronSuffix (0.00s)
--- PASS: TestRenderAccordionClosedHeaderUsesTealNotMuted (0.00s)
--- PASS: TestRenderAccordionSectionOneAlwaysOpenRegardlessOfOpenParam (0.00s)
--- PASS: TestRenderAccordionHeaderGutterWidthStableAcrossActiveState (0.00s)
--- PASS: TestHeaderInactiveStyleIsTeal (0.00s)
```

Fuer die GETRENNTEN Commits (Vorgabe: B06 einzeln revertierbar) wurde die
B06-spezifische Teal-/Token-Aenderung anschliessend NOCHMAL isoliert: die
B06-Teile (theme.go-Token, accordion.go-Farbzweig, die 2 B06-Tests) temporaer
zurueckgebaut -> B05-only-Stand baute + testete GREEN eigenstaendig (inkl.
eigener Golden-Regen) -> committet -> B06-Teile wieder aufgespielt -> erneut
GREEN + eigener Golden-Regen -> committet. Damit ist JEDER der 2 Commits fuer
sich genommen ein gruener, buildfaehiger Zwischenstand (Revert-Sicherheit).

**Voll-Lauf:** `go test ./... -short` 2x gruen (2. Lauf komplett aus Cache,
byte-stabil) fuer BEIDE Commit-Zwischenstaende (B05-only vor Commit 1,
B05+B06 vor Commit 2) -- macht 4 gruene Short-Laeufe insgesamt. `go test
./...` (voll, ohne `-short`) UND `go test ./... -race` je einmal pro
Commit-Stand gruen. `gofmt -l .` leer, `go vet ./...` leer -- bei jedem der
beiden Staende.

## Golden-Diffs

- **Commit 1 (B05):** `tree.golden`/`backlog.golden` -- Chevron-Suffix
  (`"  ▾"`/`"  ▸"`) aus allen 4 Sektions-Headern entfernt, Farben unveraendert
  (Meta bleibt `124;124;124`->eigentlich Mauve/bold fuer offen, BODY/RELATIONS/
  HISTORY bleiben `38;2;124;124;124` = Hint-Grau). `chrome.golden` unberuehrt
  (keine Accordion-Sektionen dort, wie im Plan vorhergesagt).
- **Commit 2 (B06):** `tree.golden`/`backlog.golden` -- reiner Farbwechsel der
  3 geschlossenen Header (BODY/RELATIONS/HISTORY): ANSI `38;2;124;124;124`
  (Hint-Grau) -> `38;2;139;213;202` (Teal, `#8bd5ca`). META-Header (immer
  offen, PF-1) und die Meta-Label-Spalte (`title:`/`status:`/... via
  `view_detail_bean.go`) unveraendert `124;124;124`. `chrome.golden`
  unberuehrt.

## B06-Beleg

- Pfade: `docs/plans/v1-port/b06-experiment/{README.md,before.txt,after.txt}`.
- Erzeugung: `tmux capture-pane -e -p` (ANSI-erhaltend, 100x30) gegen echten
  `./bin/bt`-Lauf in DIESEM Repo (dogfooding, `.beans/` -- nur angesehen,
  nichts mutiert). `before.txt` = Build vor dem Fix (`git stash` auf
  Vorher-Stand, Binary separat gebaut, danach `git stash pop`), `after.txt` =
  aktueller Stand.
- Token/RGB: `theme.Muted` (Hint-Grau `#7c7c7c`, ANSI `124;124;124`) ->
  `theme.HeaderInactive` (NEU, Teal `#8bd5ca`, ANSI `139;213;202`) --
  ausschliesslich fuer den geschlossenen Section-Header-Titel; Meta-Label-
  Spalte + offener META-Header unveraendert.
- Objektiver Beleg (grep-Zaehlung): Chevron-Glyphen im Screen-Dump 5->1
  (verbleibender 1x = Tree-Pane-eigener Node-Expand-Marker, NICHT Accordion);
  Teal-ANSI-Sequenz 0->3 (exakt BODY+RELATIONS+HISTORY). Zusaetzlich ein
  echter SGR-Mouse-Klick (`tmux send-keys -l` mit injizierter
  `\x1b[<0;60;19M`/`m`-Sequenz) auf die BODY-Headerzeile live verifiziert --
  Sektion aktivierte+expandierte korrekt (`▌> [2] BODY` + Body-Markdown
  sichtbar), kein Chevron-Artefakt.
- Eigenbewertung (Implementer, keine PO-Freigabe): Teal hebt die
  geschlossenen Header jetzt messbar von der (weiterhin grauen) Meta-Label-
  Spalte ab -- der von B06 benannte Verwechslungsfall ist behoben. Ob der
  konkrete Farbton gefaellt, ist die eigentliche PO-Entscheidung; Rollback ist
  eine Ein-Zeilen-Aenderung.
- Status: PO-Sign-off AUSSTEHEND (Epic `bt-ntoz` bleibt `in-progress`, PO-Gate).

## Umgeschriebene Alt-Tests

- `internal/tui/accordion_test.go::TestRenderAccordionExclusiveOpen` --
  entfernte 2 Assertions, die die PRAESENZ von `"▸"`/`"▾"` forderten (galten
  vor B05). Restliche Assertions (Body-Sichtbarkeit je Sektion) unveraendert
  gueltig, Kommentar verweist auf den neuen inversen Test.

## Smoke

- **Unit-Level:** komplette Maus-Regressionssuite explizit erneut einzeln
  gegen den finalen Stand gelaufen (`TestDetailClickRowAccountsForFiveLineHeaderOffset`,
  `TestDetailClickRowMapsSectionHeaderClick`, `TestDetailClickRowMapsMetaFieldClick`,
  `TestMouseDetailClickSectionHeaderActivatesAndExpands`,
  `TestMouseDetailClickSingleClickSelectsField`,
  `TestMouseDetailClickDoubleClickOnFieldOpensOverlay`,
  `TestMouseDetailClickDoubleClickOnBodySectionOpensEditor`,
  `TestDetailClickKeyDisjointNumberSpaces`,
  `TestMouseDetailClickTreeClickIndexDoesNotAliasFieldClickKey`,
  `TestMouseDetailClickSecondClickOnSelectedFieldOpensOverlayOutsideWindow`,
  `TestMouseDetailClickIgnoredWhenOverlayOpen`,
  `TestMouseDetailClickReachableFromBacklogView`) -- alle PASS, Kaskaden-
  Geometrie (E8-T1/T3/T4) unangetastet.
- **Real:** `tmux` gegen `./bin/bt` in diesem Repo (100x30). (1) Statischer
  Vorher/Nachher-Screen-Dump (s. B06-Beleg). (2) Live-Interaktion per
  Tastatur (Digit-Jump `2`/`3` toggelt Sektionen live -- Teal-Uebergang
  korrekt bei Re-Render nach Fokuswechsel bestaetigt). (3) Live-Interaktion
  per ECHTER SGR-Mouse-Klick-Sequenz (Ziffer/Grid ausgerechnet aus dem realen
  Screen-Dump, kein hand-geratener Wert) auf die BODY-Headerzeile --
  Sektions-Aktivierung/-Expansion funktionierte, keine Chevron-Regression.

## Deviations/ERRATA

**ERRATUM 1 (2026-07-16):** Bean-Schritt 1 nennt die Testnamen
`TestRenderAccordionInactiveHeaderNoChevronSuffix` und
`TestRenderAccordionInactiveHeaderUsesTealNotMuted`. Umgesetzt als
`TestRenderAccordionNoChevronSuffix` und
`TestRenderAccordionClosedHeaderUsesTealNotMuted` -- bewusste Praezisierung:
`accordion.go` nutzt "active"/`activeSec` bereits fuer den CURSOR-Fokus-
Zustand (D08, unabhaengig von offen/zu); ein Test namens "Inactive" waere mit
diesem bestehenden Begriff kollidiert. "Closed" benennt exakt den
`isOpen==false`-Zweig, den B06 tatsaechlich aendert. Funktional identisch zum
Bean-Wortlaut, nur klarere Terminologie.

**ERRATUM 2 (2026-07-16):** Bean-Schritt 8 schlaegt EINEN gemeinsamen Commit
vor ("Commit feat(tui): ... (B05,B06)"). Der Supervisor-Auftrag fuer diesen
Task ueberschreibt das explizit ("B05 und B06 GETRENNT -- B06 ist Experiment
und muss ggf. einzeln revertierbar sein"): umgesetzt als zwei Commits
(`fix(tui): remove redundant accordion chevron suffix (B05)` gefolgt von
`feat(tui): EXPERIMENT — teal inactive accordion headers (B06)`). Kein Bug,
nur eine explizite Praezisierung/Ueberschreibung der Bean-Vorgabe.

## Notes for bt-1u0t

Keine Beruehrungspunkte -- bt-czpf aendert ausschliesslich `accordion.go` +
`theme.go` (Section-Header-Rendering); bt-1u0t (Quit-Flow) betrifft
`box_confirm_quit.go`/`keyConfirmQuit`, disjunkte Dateien, keine geteilte
Funktion, kein Merge-Risiko.
