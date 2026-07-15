---
# bt-e4a6
title: E5 T3 — Yank y (Bean-/Epic-Kontext + Review-Stand, OSC52+nativ)
status: completed
type: task
priority: normal
created_at: 2026-07-15T09:04:24Z
updated_at: 2026-07-15T10:34:02Z
parent: bt-5h4d
---

Ziel: Yank `y` (OSC52 + nativ, Port devd `internal/clip` VERBATIM). Design
decision b (Deviation von devd, begründet): EIN view-agnostischer
`beanContext(idx, b)`-Markdown-Generator (ID/Titel/Status/Type/Priority/
Tags/Parent/Blocking/BlockedBy aufgelöst + Body, PLUS Children-Tabelle wenn
vorhanden -- deckt sowohl "Issue" als auch "Epic/Milestone" ab, da beans
keine devd-Dreiteilung Milestone/Sprint/Issue kennt, design-spec §4) statt
devds drei separaten milestoneClip/sprintClip/backlogClip-Funktionen. `y`
wirkt IMMER auf `m.focusedBean()` (Tree/Backlog identisch) -- EINZIGE
View-lokale Ausnahme: Review-Cockpit (design-spec §7 "Review-Stand
yanken") bekommt eine eigene `reviewStandMarkdown(idx)` (Gruppen-Tabelle
to-review je Epic + Rework-Sektion, spiegelt reviewQueue/reviewGroup aus
E4 T3, bt-hxyo).

Plan: docs/plans/v1-port/epic-E5-plan.md »Task 3«.

## Akzeptanz
- [x] internal/clip/clip.go (NEU, Package): Copy(s string) error -- Port
      devd VERBATIM (OSC52 + tmux/screen-Passthrough + best-effort nativ
      pbcopy/wl-copy/xclip/xsel), `command go get
      github.com/aymanbagabas/go-osc52/v2` promoten (bereits indirect in
      go.mod/go.sum -- kein neues Supply-Chain-Risiko)
- [x] internal/tui/context.go (NEU): beanContext(idx *data.Index, b
      *data.Bean) string -- Header + optionale Children-Tabelle (ID|Type|
      Status|Prio|Title, direkte Kinder, sortBeans-Reihenfolge)
- [x] internal/tui/view_review_cockpit.go (oder eigene review_context.go):
      reviewStandMarkdown(idx *data.Index) string -- Kopf + je reviewGroup
      eine Tabelle (to-review) + Rework-Sektion, reuse reviewQueue/
      reviewRework (KEIN neuer Gruppierungs-Code)
- [x] internal/tui/update.go keyNodeAction (oder keyTree/keyBacklog direkt):
      `y` (keys.Yank) -- clip.Copy(beanContext(...)) im Tree/Backlog,
      Erfolg -> showToast(toastInfo, "Kopiert", ...) (US-11: "Toast
      bestätigt"), Fehler -> showToast(toastWarn, ...)
- [x] internal/tui/view_review_cockpit.go keyReviewCockpit: `y`-Override
      (Review-Stand statt Einzel-Bean, design-spec §7)
- [x] context_test.go: TestBeanContextLeafHasNoChildrenTable,
      TestBeanContextParentHasChildrenTable, TestBeanContextResolvesParent
      TitleAndRelations, TestReviewStandMarkdownGroupsByEpic,
      TestYankShowsConfirmationToast (US-11), TestYankOnOrphanRootNoop
- [x] `command go test ./... -short` grün, gofmt/vet leer
- [x] Commit `feat(tui): Yank y (OSC52+nativ, Bean-/Epic-Kontext + Review-Stand)`

## Summary

internal/clip/clip.go (NEU, Package, ~48 Zeilen): Port devd VERBATIM
(`~/Obsidian/tools/DeveloperDashboard/apps/cli-go/internal/clip/clip.go`) --
`Copy(s string) error`, OSC52 (`aymanbagabas/go-osc52/v2`) + tmux/screen-
Passthrough auf os.Stderr, plus best-effort nativer Copy (pbcopy/wl-copy/
xclip/xsel, erste gefundene PATH-Kandidatin). `command go mod tidy` hat
`go-osc52/v2` von `// indirect` auf direct require promotet (go.mod-Diff:
1 Zeile verschoben, kein neuer Download -- die Version v2.0.1 lag bereits
in go.sum).

internal/tui/context.go (NEU, ~130 Zeilen): `beanContext(idx, b)` -- EIN
view-agnostischer Markdown-Generator (design decision b) statt devds drei
separaten milestoneClip/sprintClip/backlogClip. Header (ID/Titel/Status/
Type/Priority/Tags optional/Parent optional/Blocking optional/Blocked By
optional, Relationen aufgelöst zu "ID Title" via idx.ByID, dangling
Referenz -> "ID (unresolved)", mirrors view_detail_bean.go's resolveSorted-
Fallback) + Body, PLUS "## Children"-Tabelle NUR wenn idx.Children[b.ID]
nicht leer ist (Epic/Milestone-Fall). contextBeanTable (ID|Type|Status|
Prio|Title) ist von reviewStandMarkdown (view_review_cockpit.go)
mitgenutzt -- kein Tabellen-Code doppelt.

internal/tui/view_review_cockpit.go: reviewStandMarkdown(idx) (NEU, ~30
Zeilen, direkt neben reviewQueue/reviewRework, reused as-is -- KEIN neuer
Gruppierungscode) -- Kopf "# Review-Stand (N to-review)", je reviewQueue-
Gruppe eine "## <Epic ID> <Epic Title>"- (oder "## (kein Epic)"-) Tabelle,
danach optional "## Rework (N)". keyReviewCockpit: `y`-Override
(keybind.Matches(msg, keys.Yank)) VOR dem generischen msg.String()-Switch
-- clip.Copy(reviewStandMarkdown(m.idx)), Erfolg -> toastInfo
"Review-Stand kopiert", Fehler -> toastWarn "Yank fehlgeschlagen". Kein
Sonderfall nötig, um keyNodeActions y zu blocken: keyReviewCockpit wird in
handleKey VOR keyNodeAction dispatcht (view==viewReviewCockpit-Zweig),
strukturell identisch zum Help-Overlay-Precedent aus T2.

internal/tui/update.go keyNodeAction: neuer Case
`keybind.Matches(msg, keys.Yank)` -- `b := m.focusedBean(); if b == nil {
return true, m, nil }` (Orphan-Root-Guard, gleiches Muster wie Status/
TagAssign/Assign/Blocking/Delete/Editor), sonst clip.Copy(beanContext(m.idx,
b)) -> showToast(toastInfo, "Kopiert: "+b.ID, ...) bzw. bei Fehler
showToast(toastWarn, "Yank fehlgeschlagen", ...).

Kein neues Golden, keine Keymap-Änderung (keys.Yank existierte bereits seit
E1/E5-T2, Drift-Guards unangetastet).

## Test-Output

RED bewiesen: context_test.go zuerst geschrieben (7 Tests) gegen den
unveränderten Stand -> Compile-Fail (`internal/tui/context_test.go:52:8:
undefined: beanContext`, `go vet ./internal/tui/...`). Nach Implementierung
(internal/clip/clip.go, internal/tui/context.go, view_review_cockpit.go
reviewStandMarkdown + y-Override, update.go keyNodeAction y-Case) -> alle 7
grün (`go test ./internal/tui/... -run
'TestBeanContext|TestReviewStandMarkdown|TestYank|TestReviewCockpitYank' -v`):
TestBeanContextLeafHasNoChildrenTable, TestBeanContextParentHasChildrenTable,
TestBeanContextResolvesParentTitleAndRelations,
TestReviewStandMarkdownGroupsByEpic, TestYankShowsConfirmationToast,
TestYankOnOrphanRootNoop, TestReviewCockpitYankUsesReviewStandNotSingleBean
(Plan Step 8, zusätzlich zu den 6 Bean-Akzeptanz-Tests).

Drift-Guards explizit einzeln grün: TestKeymapNoCtrlSQ,
TestHelpGroupsCoverEveryBindingExactlyOnce (unangetastet, keine neue
Keymap-Änderung).

`go test ./... -short` grün (alle 4 Packages). `go vet ./...` leer.
`gofmt -l .` leer. Voller Lauf `go test ./... -count=1` grün
(internal/tui 126.708s, internal/data 2.224s, internal/theme 0.783s,
cmd 0.285s). Alle 7 Goldens (TestChromeGolden/TestTreeGolden/
TestTreeGoldenDeterministic/TestBacklogGolden/TestBacklogGoldenDeterministic/
TestReviewCockpitGolden/TestReviewCockpitGoldenDeterministic) `-count=2`
grün, `git status --porcelain internal/tui/testdata/` leer (byte-identisch
-- dieser Task ändert keine Views, nur Toast-Text bei Auslösung).

## Smoke

tmux 100x30, Scratch-Repo (`beans init` + 1 Milestone -> 1 Epic -> 2 Tasks,
Task Two mit Body + Tag `to-review`), Binary `bin/bt` frisch gebaut:

1. **Tree, Epic fokussiert (Parent-Fall):** `y` auf der Epic -> Toast
   "Kopiert: bt-e4a6-smoke-pwr8" sichtbar oben rechts. Host-Terminal
   `pbpaste`: `# bt-e4a6-smoke-pwr8 — Smoke Epic` + Meta (Status/Type/
   Priority/Parent aufgelöst zu "bt-e4a6-smoke-5sox Smoke Milestone") +
   "## Children"-Tabelle mit exakt 2 Zeilen (Task One, Task Two).
2. **Tree, Leaf-Task ohne Body:** `y` auf Task One -> Toast "Kopiert:
   ...hwg3". `pbpaste`: Meta + aufgelöster Parent, KEIN "## Children".
3. **Tree, Leaf-Task mit Body + Tag:** `y` auf Task Two -> Toast "Kopiert:
   ...dwpg". `pbpaste`: "- Tags: to-review" + Body-Absatz "Body for task
   two" vorhanden, KEIN "## Children".
4. **Review-Cockpit:** `R` öffnet Cockpit ("1 of 1", Task Two als einziger
   to-review-Eintrag unter der Epic). `y` -> Toast "Review-Stand kopiert"
   (NICHT "Kopiert: ..." -- Override greift). `pbpaste`:
   `# Review-Stand (1 to-review)` + `## bt-e4a6-smoke-pwr8 Smoke Epic` +
   Tabelle mit Task Two. esc -> zurück zu Browse, unversehrt.

Alle 4 Szenarien wie erwartet. Kein Crash, kein hängender Zustand.

## Deviations/ERRATA

- **Kein clip_test.go** (Plan Step 2 selbst zieht diesen Schluss: "keine
  sinnvolle Verhaltensprüfung ohne echtes Terminal/Clipboard möglich" --
  daher direkt context_test.go als Testfokus verwendet, kein separates
  leeres/triviales clip_test.go angelegt). Kein Bug, sondern Plans eigene
  Vorwegnahme.
- **TestYankShowsConfirmationToast /
  TestReviewCockpitYankUsesReviewStandNotSingleBean rufen den ECHTEN
  clip.Copy-Pfad auf** (kein Mock/Seam) -- Precedent devds eigenes
  dd2_217_test.go: "der clip.Copy-Pfad selbst wird nicht getestet
  (Seiteneffekt OSC52/pbcopy)" bzw. TestBacklogYankKeyWired, das denselben
  echten Pfad übt. Konsequenz: `go test` schreibt bei diesen 2 Tests
  tatsächlich in die System-Zwischenablage (OSC52 an os.Stderr + best-
  effort pbcopy auf macOS) -- harmlos, aber ein Seiteneffekt während der
  Testausführung, identisch zu devds eigenem Verhalten, keine neue
  Abweichung.
- **reviewStandMarkdown lebt in view_review_cockpit.go statt eigener
  review_context.go** -- Plan-Files-Sektion selbst listet es als "Modify"
  (nicht "Create"), Bean-Akzeptanz nennt beide Optionen ("oder eigene
  review_context.go"). Gewählt: Modify, wie im Plan als Default
  vorgesehen -- kein neues File für eine ~30-Zeilen-Funktion neben 2
  bereits co-lokierten Schwestern (reviewQueue/reviewRework).
- Kein neues Golden, keine Sichtbarkeits-/Layout-Änderung -- Toast-Text
  ist die einzige sichtbare Delta, bereits vom bestehenden Toast-Rendering
  (E5 T1) abgedeckt.

## Notes for T4 (Maus)

- Toast-Geometrie/Hit-Test bereits fertig verdrahtet (E5 T1,
  overlay_show_toast.go: `toastHitTest`, `target *toastTarget`) -- T3 hat
  KEIN `target` gesetzt (`nil` an jeder der 2 neuen showToast-Sites,
  Kopiert-Toast hat kein sinnvolles Klick-Ziel). Falls T4 toastHit generisch
  für jeden Toast verdrahtet: die Yank-Toasts sind reine Bestätigungen,
  ein Klick müsste ein reines No-op/Dismiss sein (kein Navigationsziel),
  keine T4-Sonderbehandlung nötig -- `target == nil` ist bereits der
  no-navigate-Fall, den overlay_show_toast.go's eigenes `dismissToast`
  laut T2-Notes ("gesetztes target.view wird übernommen") ohnehin
  voraussetzt.
- Kein neues Maus-relevantes Overlay/keine neue Picker-Fläche durch T3 --
  Yank ist eine reine Single-Key-Aktion (kein Formular, kein Menü), daher
  keine neuen Klick-Flächen für T4 zu berücksichtigen außer dem bereits
  bekannten Toast selbst.
