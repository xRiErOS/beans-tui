---
# bt-hxyo
title: E4 T3 — Review-Queue-Derivation + Cockpit-View (read-only, gruppiert)
status: completed
type: task
priority: normal
created_at: 2026-07-15T05:40:09Z
updated_at: 2026-07-15T07:15:54Z
parent: bt-tfqi
blocked_by:
    - bt-yo60
---

Ziel: Review-Queue-Ableitung (`data.EpicAncestor`, reine, zyklen-geschützte
Parent-Walk) + eine VOLLSTÄNDIG navigierbare, aber noch rein lesbare
Review-Cockpit-View (`R`): Queue `to-review` gruppiert nach nächstem
Epic-Ancestor, `(kein Epic)`-Bucket zuletzt, flache Rework-Sektion,
Verdikt-Dots (Peach/Rot), Master-Detail (Queue links, Accordion-Preview
rechts via eigenes reviewAccOpen), Summary-Zeile "x of n" / "Rework: n
offen", `n`/`p`/↑↓-Navigation, Digit-Accordion-Sprung, `esc`/`q` zurück.
Verdikt-AKTIONEN (`a`/`x`/`o`) sind bewusst NICHT dieser Task -- Task 4
(bt-yy6w) ergänzt sie in den bereits vorbereiteten Capture-Block.

Plan: docs/plans/v1-port/epic-E4-plan.md »Task 3«.

## Akzeptanz
- [x] internal/data/review.go: EpicAncestor(idx, b) (zyklen-geschützter
      Parent-Walk, visited-Map, Port-Muster expandAncestorsOf/
      CollectDescendants) + review_test.go (5 Tests: findet nächsten Epic,
      kein Epic bei Milestone-Direct-Parent, parentlos, Zyklus terminiert,
      dangling Parent terminiert)
- [x] internal/tui/view_review_cockpit.go (NEU): reviewGroup, reviewQueue
      (Epic-Gruppierung, design decision c), reviewRework (flach,
      WithTag("rework")), reviewFlat (Cursor-Indexraum), reviewSummaryLine
      (design decision j), reviewDot (Peach/Rot), reviewQueueRows
      (D08-Cursor, windowAround), renderReviewDetailPane (EIGENES
      reviewAccOpen statt m.accOpen/m.secCursor/m.fieldCursor, design
      decision i -- bewusst NICHT die geteilte renderBeanAccordionPane, da
      die dort hart auf m.accOpen/m.secCursor/m.fieldCursor liest),
      viewReviewCockpit, openReviewCockpit, keyReviewCockpit (Navigation
      only, a/x/o-Fälle als Kommentar für T4 markiert)
- [x] internal/tui/types.go: viewReviewCockpit-viewID, reviewCursor/
      reviewAccOpen
- [x] internal/tui/update.go: View()-Switch neuer Case; handleKey dritter
      Capture-Block (design decision h) UNMITTELBAR nach dem
      keys.Palette-Trigger (T1) -- ctrl+k bleibt dadurch AUCH aus der
      Cockpit heraus erreichbar, verifiziert per Test + tmux-Smoke
- [x] internal/tui/view_browse_repo.go keyTree: neuer keys.Reviews-Case
      (`R` öffnet die Cockpit aus dem Tree)
- [x] internal/tui/view_browse_backlog.go keyBacklog: identischer
      keys.Reviews-Case (aus dem Backlog)
- [x] keymap.go: KEINE Änderung nötig -- keys.Reviews (`R`) existierte
      bereits seit E1 Task 7 (unbelegt), helpGroups führt sie bereits
      (Views & Global-Gruppe) -- Grep vorab verifiziert
- [x] Neues Golden testdata/review_cockpit.golden (100x30, 2 Epic-Gruppen +
      1 (kein Epic)-Bean + 1 Rework-Bean-Fixture) -- 2x hintereinander grün
      (TestReviewCockpitGolden + TestReviewCockpitGoldenDeterministic)
- [x] go test ./... (2x), -race, gofmt -l ., go vet ./... -- alle grün
- [x] tmux-Smoke im Scratch-Repo (Details unten)

## Summary

internal/data/review.go (NEU): EpicAncestor(idx, b) läuft b.Parent aufwärts
(visited-Map, Zyklen-/dangling-Parent-sicher) bis Type=="epic" gefunden
wird oder die Kette endet. ok=false deckt alle drei Fälle ab (parentlos,
Kette ohne Epic, dangling Parent) -- design decision c braucht diese
Unterscheidung am Call-Site nicht.

internal/tui/view_review_cockpit.go (NEU): reviewQueue(idx) gruppiert
idx.WithTag("to-review") über EpicAncestor -- pro Epic MIT mindestens
einem to-review-Kind eine reviewGroup (Epics selbst UND Gruppen-Mitglieder
je data.SortBeans-sortiert), "(kein Epic)"-Bucket (epic==nil) nur wenn
nicht-leer, immer zuletzt. reviewRework ist ein dünner Pass-Through auf
idx.WithTag("rework") (bereits sortiert). reviewFlat verkettet beide zum
Cursor-Indexraum. reviewSummaryLine zählt die to-review-Beans LIVE (kein
gecachter Tally, design decision j) -- "x of n" vor der Rework-Grenze,
sonst "Rework: n offen". reviewQueueRows rendert Header (Epic gemuted /
"(kein Epic)" gedimmt / "── Rework ──"-Trenner) + Bean-Zeilen
(reviewDot+relationRow, D08-Cursor-Bar an der cursorFlat-Position) und
windowt ums Cursor-ROW (nicht cursorFlat direkt, da Header die Zeilen-
Indizes verschieben). renderReviewDetailPane ist eine bewusst EIGENSTÄNDIGE
Variante von renderBeanAccordionPane (view_browse_repo.go) -- letztere
liest m.accOpen/m.secCursor/m.fieldCursor hart vom Receiver, design
decision i verlangt aber eigene, entkoppelte Felder; eine Signatur-
Erweiterung der geteilten Funktion hätte view_browse_repo.go/
view_browse_backlog.go über den von Task 3 dokumentierten Datei-Scope
hinaus angefasst. viewReviewCockpit spiegelt viewBrowseRepo/viewBacklogs
Algebra 1:1 (masterDetailWidths/renderPane/outerBorder/composeOverlays,
KEINE neue Pane-Mathematik). keyReviewCockpit: esc/q -> Browse, Digit 1-4
togglet reviewAccOpen (identische Toggle-Semantik wie keyDetailFocus),
↑↓/n/p bewegen reviewCursor (an beiden Enden geklemmt), a/x/o sind No-op
mit Kommentarverweis auf Task 4 (bt-yy6w).

Capture-Order (design decision h): der dritte Block
(`if m.view == viewReviewCockpit { return m.keyReviewCockpit(msg) }`) sitzt
UNMITTELBAR nach T1s `keys.Palette`-Check, VOR dem globalen ctrl+c/q/tab-
Switch -- dadurch bleibt ctrl+k aus der Cockpit heraus erreichbar (Test +
Smoke verifiziert), während `a` (Assign) innerhalb der Cockpit NICHT mehr
zu keyNodeAction durchsickert (Test TestKeyReviewCockpitAssignDoesNotOpen
ParentPicker) -- der von design decision h beschriebene Tastenkollisions-
Fix ist damit für Task 3 wirksam, auch ohne dass a/x/o selbst schon etwas
tun.

## Test-Output

`command go build ./...` clean. `command go test ./... -count=1` grün (2x
hintereinander, cmd/data/theme/tui alle ok, tui ~123s). `command go test
./... -race -count=1` grün. `command gofmt -l .` leer. `command go vet
./...` leer. `command go build -o bin/bt .` ok. Golden
(TestReviewCockpitGolden + TestReviewCockpitGoldenDeterministic) mit
`-count=2` grün. 5 neue Tests in internal/data/review_test.go, 20 neue
Tests in internal/tui/view_review_cockpit_test.go.

## Smoke (tmux, Scratch-Repo, 1 Milestone + 2 Epics + 3 to-review-Beans
## (2 mit Epic-Parent, 1 parentlos) + 1 Rework-Bean)

1. Browse -> `R` -> Cockpit zeigt "Epic Alpha"/"Epic Beta"-Gruppen (je 1
   to-review-Kind), "(kein Epic)"-Bucket mit dem parentlosen Bean, "──
   Rework ──"-Trenner mit dem Rework-Bean -- exakte Gruppierung + Reihen-
   folge wie spezifiziert. Summary "1 of 3" bei Cursor auf dem ersten
   to-review-Item.
2. `n`/`n` bewegt den Cursor bis zum letzten to-review-Item (Summary
   "3 of 3"), Detail-Pane folgt sichtbar (Digit `1` öffnet die Meta-Sektion
   des JEWEILS fokussierten Beans -- Tags-Zeile wechselt mit).
3. Weiteres `n` bewegt auf den Rework-Bean -- Summary wechselt zu "Rework:
   1 offen" (kein "x of n" mehr, wie design decision j vorschreibt).
   `esc` -> zurück zu Browse.
4. `R` erneut (Cursor-Reset verifiziert: wieder "1 of 3"), `ctrl+k` INNERHALB
   der Cockpit -> Command-Center öffnet sich ALS Overlay über der Cockpit
   (Capture-Order-Beleg, design decision h). `esc` schließt die Palette,
   Cockpit bleibt sichtbar. `a` (Assign) danach -> No-op, KEIN Parent-
   Picker öffnet (Capture-Order-Negativ-Beleg).

## Deviations

- Design decision h's Prosa ("Refresh(ctrl+r) bleibt VOR dem neuen
  Capture-Block erreichbar") vs. die von derselben Design-Entscheidung
  vorgegebene Einfüge-Position (Cockpit-Block sitzt VOR dem globalen Switch,
  also auch vor dem ctrl+r-Check) widersprechen sich wörtlich genommen.
  Aufgelöst zugunsten der EXISTIERENDEN Codebase-Präzedenz: searchActive/
  filterOpen/m.form/m.overlay blockieren ctrl+r schon HEUTE genauso (jede
  volle Capture-State tut das, update.go:580 wird erst NACH all diesen
  Checks erreicht) -- die Cockpit reiht sich unauffällig in dasselbe,
  bereits akzeptierte Verhalten ein, kein Sonderfall, keine Regression.
  ctrl+r bleibt in JEDER anderen (Nicht-Full-Capture-)Situation unverändert
  erreichbar, exakt wie die Prosa es für die "anderen" Views beschreibt.
  Kein Code-Fix nötig, nur hier dokumentiert (ERRATUM-Präzedenz).
- reviewSummaryLine wurde -- wie im Task-3-Dateien-Scope bereits explizit
  gelistet -- vollständig in T3 gebaut (nicht erst in T4 nachgezogen): sie
  ist eine reine Ableitungsfunktion ohne jede Verdikt-Kopplung, gehört
  damit sauber zum "read-only Skeleton"-Scope dieser Task. T4 muss sie NICHT
  mehr anfassen, nur ggf. re-verifizieren, dass a/x/o den Cursor korrekt
  weiterschieben.
- Keine weiteren Abweichungen vom Plan-Pseudocode (Task 3 Steps 1-11 1:1
  übernommen inkl. Signaturen/Testnamen).

## Beobachtung für T4 (Verdikt-Aktionen, bt-yy6w)

- **I (kein Bug, Palette-Kontext in der Cockpit):** `m.focusedBean()`
  (update.go) hat KEINEN `viewReviewCockpit`-Case -- fällt in den
  `default`-Zweig (Tree-cursorID-Lookup) statt den aktuell fokussierten
  Review-Bean (`reviewFlat(idx)[reviewCursor]`) zu liefern. Per tmux-Smoke
  verifiziert: `ctrl+k` INNERHALB der Cockpit zeigt Node-Aktionen für den
  zuletzt im TREE fokussierten Bean, nicht für den gerade im Review
  gecursorten. T3s Scope (design decision h) verlangt das nicht explizit,
  und Task 1/2s Palette-Design predates die Cockpit -- aber T4 ergänzt
  laut Plan ohnehin "go to: review cockpit" zur Palette; falls T4/T5 auch
  node-Aktionen AUS der Cockpit heraus über die Palette anbieten wollen,
  bräuchte focusedBean() einen `viewReviewCockpit`-Case
  (`reviewFocused(reviewFlat(m.idx), m.reviewCursor)`). Nicht in T3 gefixt
  (Scope: read-only Skeleton, keine Mutations-/Palette-Erweiterung hier).
- reviewFocused/reviewIsRework (laut Plan-Pseudocode für a/x/o-Cutoff)
  existieren noch NICHT -- T4 muss sie neu bauen (kleiner Helper: Index in
  `flat` gegen `len(reviewQueue-Summe)` prüfen, s. Plan »Task 4« Step 11
  Kommentar).

Commit: siehe `Refs: bt-hxyo` (feat(tui)).
