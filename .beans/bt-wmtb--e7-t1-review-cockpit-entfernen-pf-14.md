---
# bt-wmtb
title: E7 T1 — Review-Cockpit entfernen (PF-14)
status: completed
type: task
priority: high
created_at: 2026-07-15T14:35:04Z
updated_at: 2026-07-15T15:05:55Z
parent: bt-heg9
---

Removal-Task (PF-14). Details/Steps/Akzeptanz: docs/plans/v1-port/epic-E7-plan.md Task 1. Loescht viewReviewCockpit + form_reject_review.go + 2 Goldens, behaelt PassReview/RejectReview im Datenlayer (YAGNI). MUSS zuerst laufen (alle Folge-Tasks profitieren von reduziertem Scope).


## Summary

Review-Cockpit (PF-14, PO-Nachtrag 8) vollständig aus beans-tui entfernt.
Gelöscht: `view_review_cockpit.go`(+`_test.go`), `form_reject_review.go`
(+`_test.go`), `testdata/review_cockpit.golden` (5 Dateien). Compiler-getrieben
nachgezogen: `keys.Reviews` (`R`) + `helpGroups()`-Eintrag (keymap.go),
`viewReviewCockpit`-Enum + `reviewCursor`/`reviewAccOpen`-Modellfelder
(types.go), alle `keyTree`/`handleKey`/`focusedBean`-Verweise + die zwei
`applyRepoSwitched`-Reset-Zeilen (update.go), 3 `viewReviewCockpit`-Stellen +
`mouseReviewClick` (mouse.go), Palette-Eintrag "go to review cockpit"
(overlay_palette.go), toter `rejectReview`-`submitForm`-Zweig inkl. dadurch
ungenutztem `time`-Import (box_confirm_create.go). Zusätzlich (ERRATUM-Sweep)
totes `case "reject":` in `formTitle()` entfernt (forms_shared.go) — dessen
einziger Setter (`openRejectForm`) war bereits mit `view_review_cockpit.go`
gelöscht. Kommentar-Hygiene (Step 6, optional) mitgezogen: `context.go`,
`overlay_show_toast.go`, `view_lobby.go`, `internal/data/review.go`,
`internal/data/mutations.go` — stale Verweise auf die gelöschte Datei/View
korrigiert bzw. mit Removal-Datum versehen. README: Review-Cockpit-Abschnitt +
R/a/x/o-Zeilen raus, neuer kurzer Tag-Trio-Abschnitt ("Review läuft im Chat").

YAGNI BELASSEN (Plan-Entscheid, design-spec §15 PF-14, explizit im
Supervisor-Prompt bestätigt): `data.Client.PassReview`/`RejectReview`
(mutations.go) UND `data.EpicAncestor` (internal/data/review.go) — beide
jetzt aufruferlos, aber harmlos/CLI-nah/keine TUI-Kopplung. Hinweis: der
epic-E7-plan.md-Wortlaut für Task 1 nennt im "Files (bewusst NICHT löschen)"-
Abschnitt nur `PassReview`/`RejectReview` explizit — `EpicAncestor` wird dort
nicht genannt (Plan-Lücke, Detail unter Deviations).

## Test-Output

RED→GREEN: nach Löschung der 5 Dateien (Step 2) `command go build ./...`
zeigte 10 Compile-Fehler (mouse.go, overlay_palette.go, view_browse_backlog.go,
view_browse_repo.go, update.go) — iterativ behoben, dann `command go build
./...` sauber. Test-Compile (`command go vet ./internal/tui/...`) zeigte 1
weiteren Fehler (`context_test.go: undefined: reviewFixtureBeans`) — Sweep
`grep -rln -E "ReviewCockpit|reviewCursor|reviewAccOpen|viewReviewCockpit|
openReviewCockpit|keys\.Reviews|reviewFlat|reviewClickRow|newReviewState|
reviewFocused|reviewFixtureBeans|mouseReviewClick|clampReviewCursor|
keyReviewCockpit|reviewStandMarkdown|reviewState|reviewGroup|reviewRework|
rejectReview|form_reject_review|go_review|go to: review|go to review"
internal/tui/*_test.go` bestätigte exakt die vom Plan vorhergesagten 4 Dateien
(context_test.go, overlay_palette_test.go, mouse_test.go, view_lobby_test.go)
— Cockpit-spezifische Testfälle entfernt (Datei bleibt, andere Tests bleiben
unangetastet), danach `command go vet ./...` sauber.

Voller Lauf: `command go test ./...` (ohne `-short`) → PASS, 136.5s
(internal/tui, inkl. der 7 langsamen huh-drive-Tests). `command go test
./internal/tui/ -race` → PASS, 139.4s. `command gofmt -l .` → 1 Treffer
(mouse.go, trailing blank line nach Entfernen von `mouseReviewClick`) →
`gofmt -w` behoben, danach leer. `command go vet ./...` → leer.

Goldens: SHA1 vor/nach Regen identisch —
`backlog.golden 628c413e922cbacf4359ac7734fe1a691a53a235`,
`chrome.golden db2d3398633bc47fcaa753e438e79ff37815430c`,
`tree.golden da65ba397ee75360635b253e6ae3000e6ef1d836` (alle drei
byte-identisch, `git diff --stat internal/tui/testdata/` nach `-update`
leer). Ursache verifiziert: `browseRepoChrome`/`backlogChrome` bauen ihre
Header/Footer-Strings aus expliziten `keybind.Binding`-Listen, die
`keys.Reviews` nie enthielten — die Cockpit-Tastenbelegung tauchte nur im
Help-Overlay (`helpGroups()`) und in einem hart codierten, vom echten Keymap
entkoppelten Test-Fixture-String (`chrome_test.go` `goldenChromeOpts()`,
`"...R:reviews..."`) auf, nicht in einem gerenderten Realpfad.

## Smoke

tmux (Scratch-Repo, 3 Beans: Milestone/Epic/Task mit Tag `to-review`):
`R` → Browse bleibt aktiv, kein Crash (No-Op bestätigt). `ctrl+k` → Command-
Center zeigt 14 Einträge, keinen Cockpit-Eintrag ("go to review cockpit"
fehlt, "go to: backlog"/"go to: browse" direkt benachbart). `?` → Help-
Overlay zeigt kein `R` in Navigation/Views & Global/Actions (Views & Global:
b/p/​/​/f/X/ctrl+r/ctrl+k/?/q — 9 Einträge, keine Lücke/Artefakt). `tab` auf
den Task mit Tag `to-review` → Meta-Section zeigt `Tags: ● to-review` —
Tag-Sichtbarkeit im Detail-Panel unverändert funktional.

## Deviations

- Commit-Typ `refactor(tui)!:` statt des im Plan-Dokument (epic-E7-plan.md
  Step 9) vorgeschlagenen `feat(tui)!:` — folgt der EXPLIZITEN
  Supervisor-Vorgabe im Auftrags-Prompt (`refactor(tui)!: Review-Cockpit
  entfernt (PF-14) — Review via Tag-Trio im Chat`). Plan-Wortlaut und
  Supervisor-Prompt divergieren hier; Supervisor-Prompt wurde befolgt
  (spezifischer, jüngerer Anker). Commit-Hash `a25b851`.
- `data.EpicAncestor` (internal/data/review.go) zusätzlich zu
  `PassReview`/`RejectReview` unter YAGNI belassen, obwohl der
  epic-E7-plan.md-Wortlaut in Task 1s "Files (bewusst NICHT löschen)" nur
  `PassReview`/`RejectReview` nennt. Grund: der Supervisor-Prompt selbst
  nennt explizit "PassReview/RejectReview + EpicAncestor in internal/data
  BLEIBEN" — als jüngere, spezifischere Anweisung befolgt; gleiche
  Risikolage wie die beiden Mutations-Funktionen (aufruferlos seit diesem
  Task, pure/read-only, kein TUI-Coupling). Plan-Lücke zur Kenntnisnahme
  für Planner/PO — design-spec.md §15 PF-14 nennt ebenfalls nur
  PassReview/RejectReview, nicht EpicAncestor.
- Kommentar-Hygiene (Step 6, "optional, kein Akzeptanzkriterium") vollständig
  durchgeführt statt nur "bei Gelegenheit": `context.go` (2 Stellen),
  `overlay_show_toast.go` (2 Stellen), `view_lobby.go` (1 Stelle),
  `internal/data/review.go` (Datei-Doc-Kommentar), `internal/data/
  mutations.go` (RejectReview-Doc-Kommentar) korrigiert. `renderAccordionPane`s
  Doc-Kommentar (`view_browse_repo.go:528-539`, referenziert das gelöschte
  `renderReviewDetailPane`) BEWUSST NICHT angefasst — der Plan weist diesen
  exakten Kommentar/diese Funktionssignatur explizit T4 (bt-kyj5) zu ("DER
  EINZIGE verbleibende Aufrufer seit T1" — T4 baut die Funktion ohnehin um,
  Kopfblock + neue beanSections-Parameter).
- `internal/tui/keymap.go`s Datei-Kopf-Kommentar (Zeilen 7-10, "later epic
  (Review-Cockpit, E4)") NICHT angefasst — rein historische Ist-Stand-Notiz
  von T8 (E1), vom Plan nicht explizit genannt (Step 6 nennt context.go/
  view_browse_repo.go/view_browse_backlog.go/view_lobby.go namentlich),
  bleibt sachlich korrekt als Beschreibung des damaligen Scope-Stands.
- ERRATUM-Fund (nicht im Plan vorausgesehen): `forms_shared.go`s
  `formTitle()` hatte ein `case "reject": return "Review ablehnen"` OHNE
  jeden verbleibenden Setter — entfernt, im Grep-Sweep-Abschnitt unten
  dokumentiert.

## Grep-Sweep (ERRATUM-Kultur, Pflicht)

`grep -rn -i "review" internal/ cmd/ | grep -v _test` lief NACH dem
Compiler-Durchlauf UND nach der Kommentar-Hygiene erneut. Rest-Treffer
bewertet:
- Überwiegende Mehrheit (~30 Treffer, `types.go`/`update.go`/
  `box_confirm_delete.go`/`view_browse_repo.go`/`overlay_palette.go`/
  `mouse.go`/`data/watcher.go`/`data/mutations.go`/`data/index.go`) sind
  Code-REVIEW-Runden-Referenzen ("Review-Runde 2", "T5-Review", "T8-review",
  "E3-T4-Review PFLICHT", …) oder `preview`-Substring-Treffer (`loadDeletePreview`,
  "modal's preview text") — KEIN Cockpit-Feature-Bezug, unangetastet.
- `internal/tui/view_browse_repo.go:528-539` (`renderAccordionPane`-Doc,
  referenziert `renderReviewDetailPane`/`reviewAccOpen`) — bewusst T4
  überlassen (s. Deviations).
- `internal/tui/testdata/chrome.golden:2` — literaler Testfixture-String
  (`chrome_test.go` `goldenChromeOpts()`), byte-identisch bestätigt (s.
  Test-Output), kein Nacharbeits-Bedarf.
- `internal/data/mutations.go` `PassReview`/`RejectReview`-Funktionskörper
  ("## Review <date>" ist der von RejectReview generierte Markdown-Header,
  Teil des bewusst belassenen YAGNI-Codes) — unangetastet, korrekt.
- Alle sonstigen direkten Feature-Treffer (keymap.go/types.go/update.go/
  mouse.go/overlay_palette.go/box_confirm_create.go/forms_shared.go/
  view_browse_backlog.go/view_lobby.go/context.go/overlay_show_toast.go/
  data/review.go) → entfernt bzw. korrigiert (s. Summary).

## Notes for T2 (Glyphen)

Beim Grep-Sweep gesehene Stellen, an denen Typ-/Status-/Prio-Glyphen zentral
leben (T2 = `bt-2af1`, Theme/Glyphen-Umstellung PF-6):
- `internal/theme/theme.go` — `statusColor`-Map, `statusGlyph`/
  `statusGlyphASCII`-Konstanten (T2 ersetzt durch `statusLetter`-Map),
  `StatusIcon()`, `priorityColor`, `Priority()`.
- `internal/theme/icons.go` — `typeIcon`-Map, `typeIconASCII` (T2: entfällt
  komplett), `typeColor`-Map, `TypeIcon()`/`TypeStyle()`.
- `internal/theme/theme_test.go` — `TestStatusColorMapping`,
  `TestAsciiFallback`, `TestTypeIconAllTypes`, `TestPriorityColorMapping`
  (alle vier laut epic-E7-plan.md Task 2 umzuschreiben).
- Call-Sites, die `theme.StatusIcon`/`theme.TypeIcon`/`theme.Priority`
  konsumieren (Render-Pfade, die T2s Golden-Diffs erzeugen werden):
  `internal/tui/view_browse_repo.go` (Tree-Zeilen-Rendering),
  `internal/tui/view_browse_backlog.go` (Backlog-Zeilen), `internal/tui/
  view_detail_bean.go`/`accordion.go` (Meta-Sektion/Beziehungen-Feldliste —
  T4 baut `metaFields()` hier NEU und konsumiert direkt T2s Glyph-Output,
  laut epic-E7-plan.md Task 4 Step 4 "nutzt T2s Glyph-Output").
- KEINE Cockpit-Reste mehr im Weg: alle drei Theme-Konsumenten, die früher
  auch von `view_review_cockpit.go` (`reviewQueueRows`, Verdikt-Dots)
  genutzt wurden, sind jetzt ausschließlich Tree/Backlog/Meta — T2 muss
  keinen vierten (gelöschten) Call-Site mehr berücksichtigen.

Refs: bt-wmtb
