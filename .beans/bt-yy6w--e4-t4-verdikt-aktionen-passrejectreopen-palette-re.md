---
# bt-yy6w
title: 'E4 T4 — Verdikt-Aktionen: Pass/Reject/Reopen + Palette-Review-Eintrag'
status: completed
type: task
priority: normal
created_at: 2026-07-15T05:40:13Z
updated_at: 2026-07-15T08:05:08Z
parent: bt-tfqi
blocked_by:
    - bt-hxyo
---

## Übernommene Findings aus E4-T3-Review (PFLICHT)
- [x] I01: reviewQueue/reviewFlat/reviewRework werden 3-4× pro Render/Keypress neu abgeleitet — einmal pro Aufruf berechnen und als Parameter durchreichen BEVOR T4 weitere Call-Sites ergänzt
- [x] I02: renderReviewDetailPane vs renderBeanAccordionPane ~15 Zeilen dupliziert — kleiner privater Helper (open/secCursor/fieldCursor/focused als Parameter), Kopplung bleibt draußen
- [x] Testlücken: (a) to-review-getaggtes Epic selbst (Gruppen-Header + eigener Eintrag), (b) Render-Clamp-Pfad bei extern geschrumpfter Queue, (c) rework-only Queue
- [x] focusedBean(): Cockpit-Case ergänzt (reviewFlat[reviewCursor]) — Palette-Node-Actions wirken sonst auf den falschen Bean (T3-Note)

## Akzeptanz

- [x] internal/data/mutations.go: PassReview (--status completed --remove-tag to-review, EIN Call), RejectReview (--remove-tag to-review --tag rework --body-append "## Review <datum>", EIN Call) — gegen `beans update --help` verifiziert (Design-Entscheidung d)
- [x] internal/tui/form_reject_review.go (NEU): buildRejectReviewForm (huh.NewText, Key "comment", nonEmpty-Pflichtfeld, KEIN Prefill), openRejectForm (mutTarget, formKind "reject", KEIN ETag-Capture bei Open — frisch bei Submit, Design-Entscheidung d)
- [x] internal/tui/forms_shared.go: formTitle-Case "reject" → "Review ablehnen"
- [x] internal/tui/box_confirm_create.go: submitForm-Case "reject" (direktes Feuern wie "editTitle", KEIN Confirm-Gate, time-Import ergänzt für time.Now().Format)
- [x] internal/tui/view_review_cockpit.go: keyReviewCockpit a/x/o verdrahtet (a=Pass nur auf to-review, x=Reject-Form nur auf to-review, o=Reopen/SetTags nur auf Rework — alle No-op im falschen Zustand); reviewFocused/reviewIsRework (neue kleine Helfer, ERRATUM vs. Plan-Pseudocode: reviewIsRework(b) statt reviewIsRework(idx,b) — direkter Tag-Check statt zweitem idx-Walk); reviewSummaryLine bereits in T3 fertig, unverändert wiederverwendet
- [x] I01-Refactor: reviewState (groups/toReview/rework) + newReviewState — EINE idx.WithTag/EpicAncestor-Ableitung je Render (viewReviewCockpit) bzw. je Keypress (keyReviewCockpit), reviewFlat/reviewSummaryLine als dünne idx-Wrapper für Tests erhalten (Signaturen unverändert), reviewQueueRows nimmt jetzt reviewState statt *data.Index
- [x] I02-Refactor: neue paketweite renderAccordionPane(idx, b, w, h, open, secCursor, fieldCursor, focused) in view_browse_repo.go — renderBeanAccordionPane UND renderReviewDetailPane delegieren jetzt beide dorthin, Rendering byte-identisch (Goldens unverändert grün ohne -update)
- [x] internal/tui/update.go: focusedBean() neuer viewReviewCockpit-Case (reviewFocused(reviewFlat(m.idx), m.reviewCursor))
- [x] internal/tui/overlay_palette.go: paletteActions "go_review" NACH "go_browse" angehängt, dispatchPalette-Case → m.openReviewCockpit() (reused verbatim, gleicher Cursor/AccOpen-Reset wie `R`)
- [x] Tests: internal/data/client_mut_test.go (4 neue: Pass/Reject Roundtrip + je 1 Conflict-Test), internal/tui/form_reject_review_test.go (NEU, 4 Tests), internal/tui/view_review_cockpit_test.go erweitert (10 neue: a/x/o inkl. Rework-Noop-Gegenstücke, focusedBean Cockpit-Case ×2, 3 ererbte Testlücken a/b/c), internal/tui/overlay_palette_test.go erweitert (2 neue go_review-Tests + 1 bestehender Test an neue Aktionsliste angepasst)
- [x] command go test ./... (2×), -race, gofmt -l ., go vet ./..., Goldens (Chrome/Tree/TreeDeterministic/Backlog/BacklogDeterministic/ReviewCockpit/ReviewCockpitDeterministic) 2× mit -count=1 — alle grün
- [x] command go build -o bin/bt . ok
- [x] tmux-Smoke im frischen Scratch-Repo (Details unten) — kompletter US-08-Flow belegt

## Summary

PassReview/RejectReview (mutations.go) als je EIN kombinierter `beans update`-Call, exakt wie im Plan vorgeschrieben und gegen `beans update --help` verifiziert (--status/--remove-tag bzw. --remove-tag/--tag/--body-append sind unabhängige, frei kombinierbare Flags). RejectReview baut den Body-Abschnitt `"\n## Review <date>\n\n" + comment + "\n"` — ERRATUM (empirisch): die reale beans-0.4.2-CLI kappt beim Zurückschreiben genau EIN abschließendes Newline, `updated.Body` endet daher NICHT auf dem literalen `"...\n"` des Plans; Tests prüfen den Abschnitt daher per `strings.Contains` (gleiche Konvention wie das bestehende `TestAppendBodyAddsSection`), die Produktionslogik selbst ist unverändert korrekt (das schreibende `--body-append`-Argument braucht sein eigenes Trailing-Newline für einen sauberen Absatzabschluss, unabhängig davon, was die CLI beim Zurücklesen normalisiert).

Reject-Kommentar-Form (form_reject_review.go) reused die komplette E3-Formular-Infra (styleForm/formChrome/keyForm/updateForm) — EIN Pflichtfeld (huh.NewText, nonEmpty), KEIN Prefill (anders als editTitle: eine Ablehnung ohne eigene Formulierung entwertet den Zweck des Kommentars), KEIN Confirm-Gate (submitForm "reject"-Case feuert RejectReview direkt, ETag frisch bei Submit gelesen — Design-Entscheidung d/e unverändert). ERRATUM (empirisch, Testinfra): huh.Text-Felder akzeptieren getippte Runen erst NACH `Focus()`, das nur `Group.Init()` bei `group.active` setzt — anders als die bestehenden Input-Feld-Tests (editTitle/create), die nie tippen mussten (nur Prefill via `.Value(&v)` gelesen), musste `form_reject_review_test.go` als ERSTER Test in diesem Repo, der echten Tastatur-Input in ein Feld treibt, `f.Init()`s Cmd-Kette VOR dem Tippen selbst durch `driveFormCmd` jagen — dokumentiert direkt im Testcode, kein Confirm-Gate-Bug, reine Test-Harness-Lücke.

keyReviewCockpit a/x/o: `a`/`x` feuern nur auf einem Bean OHNE "rework"-Tag (Pass/Reject dürfen ein bereits verdiktetes Rework-Item nicht doppelt verarbeiten), `o` nur auf einem Bean MIT "rework"-Tag (Design-Entscheidung f: Reopen ist die manuelle Umkehrung des Agent-Schritts "rework → to-review", ein bereits im Zielzustand befindliches to-review-Item ist ein No-op). `reviewIsRework(b)` weicht vom Plan-Pseudocode `reviewIsRework(idx, b)` ab (ERRATUM, dokumentiert im Code): ein direkter Tag-Check auf dem bereits aus dem live-`idx` gezogenen Bean ist äquivalent und braucht keinen zweiten `idx.WithTag`-Walk — passt zur I01-Lehre (Redundanz vermeiden, nicht nur verschieben).

I01-Fix (ererbtes PFLICHT aus T3-Review): neuer `reviewState`-Typ (groups/toReview/rework) + `newReviewState(idx)` als EINE `idx.WithTag`/`EpicAncestor`-Ableitung pro `viewReviewCockpit`-Render bzw. pro `keyReviewCockpit`-Keypress — vorher liefen pro Render bis zu 3× `reviewQueue` und 2× `reviewRework` unabhängig. `reviewFlat`/`reviewSummaryLine` bleiben als dünne `*data.Index`-Wrapper für bestehende Tests/Call-Sites erhalten (Signaturen unverändert), `reviewQueueRows` nimmt jetzt `reviewState` statt `*data.Index` entgegen (kein Test hing direkt an dieser Signatur). `keyReviewCockpit`s a/x/o nutzen dieselbe bereits am Funktionskopf abgeleitete `flat`-Variable — keine zusätzliche Re-Derivation durch die drei neuen Call-Sites.

I02-Fix (ererbtes PFLICHT aus T3-Review): neue paketweite `renderAccordionPane(idx, b, w, h, open, secCursor, fieldCursor, focused)` (view_browse_repo.go) fasst den ~15-zeiligen bodyW/accW/beanSections/renderAccordion/renderPane-Körper, den `renderBeanAccordionPane` und `renderReviewDetailPane` vorher hand-dupliziert hatten. Beide Aufrufer bleiben dünne Wrapper, die NUR ihren eigenen State (Tree/Backlog: m.accOpen/m.secCursor/m.fieldCursor; Cockpit: reviewAccOpen-abgeleitetes activeIdx) als Parameter durchreichen — Design-Entscheidung i (eigene Felder) bleibt vollständig gewahrt, nur der Rendering-Körper ist jetzt geteilt. Alle Goldens (inkl. ReviewCockpit) blieben ohne `-update` grün — Beleg, dass das Refactoring byte-identisch ist.

focusedBean() (update.go) bekam den in bt-hxyos eigener "I (kein Bug)"-Notiz vorhergesagten `viewReviewCockpit`-Case: `reviewFocused(reviewFlat(m.idx), m.reviewCursor)` statt des Default-Tree-Zweigs — Palette-Node-Aktionen (ctrl+k aus der Cockpit heraus) wirken jetzt korrekt auf den gecursorten Review-Bean.

Palette (overlay_palette.go): "go to: review cockpit" NACH "go to: browse" angehängt (Plan-Vorgabe wörtlich), `dispatchPalette`-Case ruft `m.openReviewCockpit()` unverändert wieder — identisches Cursor/AccOpen-Reset-Verhalten wie die direkte `R`-Taste.

## Test-Output

command go build ./... clean. command go test ./internal/data/ -run 'TestPassReview|TestRejectReview' -v: 4/4 grün. command go test ./... -count=1 (2× hintereinander): cmd/data/theme/tui alle ok (tui ~127s je Lauf). command go test ./... -race -count=1: grün (~129s). command gofmt -l . leer. command go vet ./... leer. Goldens (Chrome/Tree/TreeDeterministic/Backlog/BacklogDeterministic/ReviewCockpit/ReviewCockpitDeterministic) 2× mit -count=1 (erzwungen, nicht gecached): alle grün, KEIN -update nötig (I01/I02-Refactor ist rendering-neutral). command go build -o bin/bt . ok. 21 neue/erweiterte Tests in internal/tui (form_reject_review_test.go NEU + view_review_cockpit_test.go + overlay_palette_test.go), 4 neue in internal/data/client_mut_test.go.

## Smoke (tmux, frischer Scratch-Repo bt-smoke-e4t4: 1 Epic "Payments Overhaul" + 2 Kinder-Tasks, beide via CLI auf to-review getaggt — Agent-Simulation)

1. `bt` startet in Browse. `R` → Cockpit zeigt "Payments Overhaul"-Gruppe mit beiden Kindern (Peach-Dots), Summary "1 of 2", Cursor auf "Add Retry Logic To Webhook".
2. `a` (Pass) auf dem gecursorten Bean → Reload, Bean verschwindet aus der Queue, Summary "1 of 1" (nur noch "Refactor Checkout Flow" übrig). `cat` der `.md`-Datei: `status: completed`, kein `to-review`-Tag mehr.
3. `x` auf "Refactor Checkout Flow" → Kommentar-Form "Review ablehnen" öffnet, Freitext "Bitte Edge-Case für 429-Antwort ergänzen" getippt (im Terminal sichtbar), enter → Reload, Queue-Summary wechselt zu "Rework: 1 offen", Bean erscheint unter "── Rework ──". `cat`: Tag `rework` statt `to-review`, Body endet auf `## Review 2026-07-15\n\nBitte Edge-Case für 429-Antwort ergänzen`.
4. `o` auf dem Rework-Bean → Reload, Bean zurück in der "Payments Overhaul"-Gruppe, Summary wieder "1 of 1". `cat`: Tag zurück auf `to-review`, der `## Review`-Abschnitt bleibt im Body erhalten (Reopen tilgt das Feedback nicht, nur den Verdikt).
5. `esc`/`q` zurück zu Browse, `ctrl+k` → Command-Center öffnet, "go to: review cockpit" erscheint direkt nach "go to: browse" in der Aktionsliste; Query "review cockpit" filtert auf genau diesen Eintrag, enter → Cockpit öffnet erneut (Cursor/AccOpen frisch reset). Terminal-Ausschnitte oben als Beleg — DER Abnahme-Smoke für US-08 vollständig erbracht.

## Deviations

- **ERRATUM (RejectReview-Body-Trailing-Newline):** s. Summary — reine Test-Assertion-Anpassung (Contains statt exaktem Suffix), Produktionscode unverändert korrekt.
- **ERRATUM (Plan-Pseudocode reviewFocused/reviewIsRework):** Plan skizziert `reviewFocused(flat, cursor)` (übernommen 1:1) und `reviewIsRework(idx, b)` (abgewichen zu `reviewIsRework(b)` — direkter Tag-Check, kein idx-Argument nötig, s. Summary).
- **Scope-Erweiterung über die Plan-Files-Liste hinaus:** `internal/tui/view_browse_repo.go` wurde zusätzlich geändert (I02-Fix, `renderAccordionPane` extrahiert) — nicht im Plan-»Task 4«-Files-Abschnitt gelistet, aber zwingend, um den ererbten I02-PFLICHT-Fund aus der T3-Review zu schließen (der Fund selbst benennt explizit `renderBeanAccordionPane` als Duplikations-Partner).
- **Test-Harness-Lücke (huh.Text-Focus):** s. Summary — dokumentiert im Testcode selbst, kein Produktionsbug.
- Keine weiteren Abweichungen vom Plan-Pseudocode (Task 4 Steps 1-14 inhaltlich 1:1 übernommen inkl. Mutations-Signaturen).

## Notes für T5 (Abschluss, bt-v7ti)

- Keine offenen Findings aus dieser Review-Runde — I01/I02 sind jetzt geschlossen, keine weitere Re-Derivation-Altlast für künftige Cockpit-Erweiterungen.
- E5-Scope-Cuts bleiben unverändert unbelegt (Design-Entscheidungen g/Palette-Doku): `y` (Yank) in der Cockpit weiterhin No-op, Toast weiterhin Statuszeilen-Interim (Konflikt-Flash-Problem bekannt, E5).
- PO-Review-Kern-Flow (US-08) ist mit diesem Task vollständig abgeschlossen und smoke-belegt — T5 ist reine Abschluss-Zeremonie (Full-Suite/Build/Doku), keine neue Funktionalität erwartet.

Commit: siehe Refs: bt-yy6w (feat(tui)).
