---
# bt-v7ti
title: E4 T5 — E4-Abschluss-Ritual
status: completed
type: task
priority: normal
created_at: 2026-07-15T05:40:16Z
updated_at: 2026-07-15T08:48:38Z
parent: bt-tfqi
blocked_by:
    - bt-yy6w
---

## Übernommene Findings aus E4-T4-Review (PFLICHT)
- [x] B01: m.reviewCursor nach boundary-shrinking Pass nie im Model geclampt (nur render-lokal) → reviewFocused nil auf sichtbar markierter Zeile, down/n eingefroren. Fix: Clamp in applyLoaded (Spiegel der cursorID-Reconciliation) ODER am Anfang von keyReviewCockpit. Regressionstest über echte mutateCmd→mutationDoneMsg→Reload-Pipeline (nicht nur View()).
- [x] I01: Datenlayer-Test RejectReview-Kommentar mit Anführungszeichen, Backtick, Umlaut, eingebettetem \n (argv-Garantie pinnen)
- [x] Q01 (Doc-Note): Palette-Delete im Cockpit trifft jetzt reviewed Bean — generisches Delete-Confirm reicht, kurz im bean dokumentieren

## Akzeptanz

- [x] B01 RED zuerst: neuer Test `TestReviewCursorClampedThroughRealMutationReloadPipeline` (internal/tui/view_review_cockpit_test.go) reproduziert den Fund gegen die UNGEFIXTE Cockpit-Logik (`reviewCursor = 2` blieb nach dem Schrumpfen stehen, valider Bereich war `[0,2)`) — echte Pipeline: `keyReviewCockpit('a')` → `cmd()` → `mutationDoneMsg` → `m.Update` (`applyMutationResult`) → `beansLoadedMsg` (simuliert den serverseitig erfolgreichen Pass) → `m.Update` (`applyLoaded`).
- [x] B01 Fix (GRÜN): neue Methode `clampReviewCursor` (internal/tui/view_review_cockpit.go, direkt nach `openReviewCockpit`) klemmt `m.reviewCursor` gegen `reviewFlat(m.idx)`; aufgerufen aus `applyLoaded` (internal/tui/update.go) als gemeinsamer Tail nach der bestehenden `cursorID`-Reconciliation — Spiegel-Ort wie im Fund vorgeschlagen, fixt das Modell-Feld SOFORT beim Reload statt erst beim nächsten Tastendruck.
- [x] I01: neuer Test `TestRejectReviewCommentSpecialCharactersSurviveArgv` (internal/data/client_mut_test.go) — Kommentar mit Anführungszeichen/Backtick/Apostroph/eingebettetem `\n`/Umlauten, End-to-End gegen die echte `beans`-Binary. War von Anfang an GRÜN (kein Produktionscode-Fix nötig) — `client.go`s `run` nutzt `exec.Command("beans", args...)`, also strukturell nie eine Shell, die argv-Garantie war schon vorher korrekt, jetzt EMPIRISCH gepinnt statt nur angenommen.
- [x] Q01: Doc-Note in `focusedBean()` (internal/tui/update.go) ergänzt — Palette-Delete (`paletteActions`' "bean: löschen") trifft im Cockpit jetzt den gecursorten Review-Bean; `openDeleteConfirm` (box_confirm_delete.go) baut seine Kinder-/Verknüpfungs-Warnung bereits rein aus `m.focusedBean()`/`m.idx`, keine Tree/Backlog-Annahme eingebacken — bestätigt: generisches Delete-Confirm reicht, kein Cockpit-Sonderfall nötig. Live im Smoke bestätigt (s.u.).
- [x] Manueller Dogfooding-Smoke (tmux, frischer Scratch-Repo) — Details s. Smoke-Abschnitt unten, inkl. explizitem B01-Szenario (Pass des LETZTEN Queue-Eintrags, Navigation danach bewiesen funktionsfähig).
- [x] `command go test ./... -count=1` 2× hintereinander grün, `command go test ./... -race -count=1` grün, `command gofmt -l .` leer, `command go vet ./...` leer, `command go build -o bin/bt .` ok, alle 7 Goldens (`-run Golden`) 2× grün.
- [x] README.md: Status-Absatz E4 als fertig markiert, Keybindings-Tabelle "Stand E4" + `ctrl+k`/`R`-Zeilen + neue Review-Cockpit-Subsektion (↑/↓/n/p, 1-4, a/x/o, esc/q, ctrl+k-Ausnahme).
- [x] beans-Pflege: T1-T4 (bt-jpgn/bt-yo60/bt-hxyo/bt-yy6w) per `beans show` verifiziert — bereits `completed` (durch die jeweils eigene Abschluss-Runde), keine Änderung nötig (verify-only, wie vom PO abgegrenzt).
- [ ] Epic `bt-tfqi` (Status/Tag `to-review`) — NICHT in diesem Task-Scope (PO/Supervisor-Gate, s. Notes unten).

## Summary

B01-Fix bewusst in `applyLoaded` statt am Kopf von `keyReviewCockpit` platziert (der Fund erlaubte beides): `applyLoaded` ist die EINE Stelle, an der `m.idx` (und damit die Größe von `reviewFlat`) sich durch einen Reload ändern kann — der bestehende `cursorID`-Reconciliation-Block dort klemmt das Tree-Pendant bereits nach genau demselben Prinzip ("clamp auf len-1 statt auf 0 zurückspringen"). Ein Fix nur am Kopf von `keyReviewCockpit` hätte das Problem erst beim NÄCHSTEN Tastendruck behoben — `focusedBean()`/`reviewFocused()`-Aufrufe dazwischen (z. B. ctrl+k Palette direkt nach dem Pass) hätten weiterhin `nil` gesehen. `clampReviewCursor` selbst ist unconditional (kein View-Gate) und dadurch sicher aus dem gemeinsamen Reload-Pfad aufrufbar, ohne Sonderfall für `m.view != viewReviewCockpit`.

Regressionstest fährt die ECHTE Pipeline (keine Model-Feld-Manipulation von Hand): `keyReviewCockpit('a')` liefert den `mutateCmd`, `cmd()` liefert `mutationDoneMsg`, `m.Update(msg)` durchläuft `applyMutationResult`s unconditionalen Reload-Tail, ein `beansLoadedMsg` mit dem entfernten letzten Bean simuliert den erfolgreichen Server-Roundtrip, `m.Update` durchläuft `applyLoaded`. Anschließend geprüft: (1) `reviewCursor` liegt sofort im validen Bereich, (2) `reviewFocused`/`focusedBean()` lösen sofort einen echten Bean auf (nicht `nil`), (3) `n` an der (jetzt korrekten) letzten Position ist ein sauberer No-op, (4) `p` bewegt den Cursor tatsächlich — schließt explizit die im Fund beschriebene Asymmetrie (down/n eingefroren, up/p "heilt" zufällig) als Regression.

Der bestehende `TestReviewCockpitViewClampsCursorWhenQueueShrinksExternally` (T3-Review-Erbe) bleibt unverändert bestehen — er beweist weiterhin ausschließlich `View()`s renderlokale Klemme (eine lokale `cursor`-Variable) und dient jetzt als dokumentierter Kontrast zum neuen Pipeline-Test, nicht als Ersatz dafür.

I01 brauchte keinen Produktionscode: `internal/data/client.go`s `run` ruft `exec.Command("beans", args...)` — Go interpretiert das NIE über eine Shell, jedes Element von `args` bleibt ein eigenes argv-Element unabhängig von Anführungszeichen/Backticks/Newlines darin. Der neue Test pinnt das empirisch End-to-End (Kommentar → CLI-Aufruf → On-Disk-Datei → `List()`-Rückweg), exakt nach dem Muster von `TestRejectReviewSwapsTagAndAppendsSection`, nur mit einem bewusst garstigen Kommentar-String.

Q01 brauchte ebenfalls keinen Produktionscode: `openDeleteConfirm` liest ausschließlich `m.focusedBean()`/`m.idx` — view-agnostisch von Haus aus. Seit T4s `focusedBean()`-Erweiterung um den `viewReviewCockpit`-Case (Palette-Node-Actions sollen den gecursorten Review-Bean statt eines stale Tree-Beans treffen) trifft Palette-Delete im Cockpit jetzt konsequent denselben Bean — verifiziert (kein Kinder-/Verknüpfungs-Sonderfall fehlt) und dokumentiert direkt an der Stelle, die diesen Seiteneffekt verursacht (`focusedBean()`-Doc-Kommentar, update.go).

## Test-Output

`command go test ./... -count=1` (Lauf 1): cmd 0.53s · internal/data 2.35s · internal/theme 0.21s · internal/tui 125.84s — alles ok.
`command go test ./... -count=1` (Lauf 2): cmd 0.86s · internal/data 2.30s · internal/theme 0.30s · internal/tui 125.96s — alles ok.
`command go test ./... -short -count=1`: cmd 0.28s · internal/data 2.26s · internal/theme 0.89s · internal/tui 7.76s (Gesamt 8.08s) — bestätigt die im Repo-CLAUDE.md dokumentierte ~121s→~3-5s-Ersparnis weiterhin (7.76s, leicht über der dokumentierten Spanne, Rechnerlast-abhängig, keine Regression).
`command go test ./... -race -count=1`: cmd 1.47s · internal/data 3.42s · internal/theme 1.97s · internal/tui 128.60s — alles ok, keine Data-Races.
`command gofmt -l .`: leer. `command go vet ./...`: leer.
`command go build -o bin/bt .`: ok (bin/bt, 16.6 MB).
`command go test ./internal/tui/ -run Golden -v -count=1` 2× hintereinander: TestChromeGolden, TestTreeGolden, TestTreeGoldenDeterministic, TestBacklogGolden, TestBacklogGoldenDeterministic, TestReviewCockpitGolden, TestReviewCockpitGoldenDeterministic — alle 7/7 PASS in beiden Läufen, kein `-update` nötig (B01-Fix ist rendering-neutral, Klemme wirkt nur auf das Model-Feld, nicht auf View()-Ausgabe).
`command go test ./internal/tui/ -run TestReviewCursorClampedThroughRealMutationReloadPipeline -v`: RED (vor dem Fix) → `reviewCursor = 2 after boundary-shrinking Pass, want clamped into [0, 2)`; GREEN (nach dem Fix) → PASS.
`command go test ./internal/data/ -run TestRejectReviewCommentSpecialCharactersSurviveArgv -v`: PASS (sofort grün, kein Fix nötig).

## Smoke (tmux, frischer Scratch-Repo `bt-smoke`)

Fixtures (echte `beans create`/`beans update`-Aufrufe gegen einen frischen `beans init`-Scratch-Repo, Prefix `bt-smoke-`): 1 Milestone, Epic Alpha + Epic Beta (beide Kinder des Milestones), 4 to-review-Tasks (Alpha Task One/Two, Beta Task One, ein parentloses "Parentless Review Task" → "(kein Epic)"-Bucket), 1 rework-Task ("Rework Task", Kind von Epic Alpha).

1. `bt .` startet in Browse. `ctrl+k` → Command-Center öffnet (Capture-Order-Beleg positiv), Aktionsliste kontextabhängig (kein Bean fokussiert → keine Node-Actions oben). Query "Beta" (≥3 Zeichen) findet "Epic Beta" UND "Beta Task One" — beide zu diesem Zeitpunkt AUSSERHALB des sichtbaren (eingeklappten) Tree-Ausschnitts. `enter` auf "Beta Task One" → springt korrekt hin, Ahnen (Milestone, Epic Beta) automatisch expandiert, Cursor exakt auf dem Ziel-Bean.
2. `f` öffnet den Facetten-Filter, danach `ctrl+k` → Command-Center öffnet NICHT (Capture-Order-Negativ-Beleg, Terminal-Ausschnitt zeigt den Filter-Overlay unverändert). `esc` schließt den Filter.
3. `R` → Review-Cockpit öffnet: Gruppen "Epic Alpha" (Alpha Task One/Two) / "Epic Beta" (Beta Task One) / "(kein Epic)" (Parentless Review Task) / "── Rework ──" (Rework Task), Summary "1 of 4", Cursor auf "Alpha Task One".
4. `ctrl+k` INNERHALB des Cockpits → Command-Center öffnet TROTZDEM (Capture-Order-Ausnahme, Design-Entscheidung h) — Aktionsliste enthält jetzt "bean: löschen" (Q01 live: würde jetzt "Alpha Task One" treffen, den gecursorten Review-Bean, nicht mehr einen stale Tree-Bean). `esc` schließt die Palette, zurück im Cockpit.
5. `n` ×10 (No-op-sicher über die Queue-Grenze hinaus) seekt robust zum WIRKLICH letzten Flat-Eintrag: "Rework Task" (Summary "Rework: 1 offen"). `o` (Reopen) → EIN Fehlschlag durch einen bereits bekannten, dokumentierten beans-0.4.2-CLI-Quirk (ETag-Divergenz zwischen `list`/`show` und dem internen `update --if-match`-Check bei Beans, deren `tags:`-Block beim allerersten Schreiben — hier: `beans create --tag rework` — entsteht; identisches Muster wie `internal/data/testrepo_test.go`s eigener Doku-Kommentar zu diesem Upstream-Verhalten, bt-tknb). Workaround (kein beans-tui-Code betroffen): alle Fixture-Beans einmal über einen echten `beans update --tag ...`-Aufruf normalisiert (danach ist der ETag konsistent, exakt wie im testrepo_test.go-Kommentar beschrieben). `ctrl+r` neu geladen, `o` erneut → diesmal erfolgreich: "Rework Task" wandert zurück in die "Epic Alpha"-Gruppe, KEINE Rework-Sektion mehr, Summary "5 of 5", Cursor (unverändert 4) landet — reiner Zufall der Positions-Arithmetik, kein Klemmen nötig, da die Flat-Länge gleich blieb (1 Rework raus, 1 to-review-Slot rein) — exakt auf "Parentless Review Task", dem jetzt WIRKLICH letzten Eintrag.
6. **B01-Szenario:** `a` (Pass) auf "Parentless Review Task" (letzter Queue-Eintrag) → Reload → Bean komplett aus der Queue verschwunden (status completed, Tag entfernt, per `beans show` auf der Platte verifiziert), Summary sofort "4 of 4", Cursor SOFORT (ohne weiteren Tastendruck) korrekt auf "Beta Task One" (den neuen letzten Eintrag) hervorgehoben — kein "nil"/leerer Fokus, keine optische Karteileiche.
7. `n` → No-op, bleibt "4 of 4" (korrektes Verhalten an der Grenze, NICHT eingefroren-falsch). `p` → bewegt sich zu "3 of 4" ("Rework Task") — Navigation NACH dem Grenz-Pass beweisbar funktionsfähig, exakt die im Fund verlangte Regression widerlegt.
8. `x` auf "Rework Task" → Reject-Kommentar-Form "Review ablehnen" öffnet, Freitext "Bitte Grenzfälle nochmal prüfen" getippt, `enter` → Reload, Summary "3 of 3", Bean erscheint wieder unter "── Rework ──". `beans show` bestätigt: Tag `rework`, Body endet auf `## Review 2026-07-15\n\nBitte Grenzfälle nochmal prüfen`.
9. `esc` → zurück zu Browse. `q` → Quit-Confirm-Modal ("Quit? Really quit bt.") korrekt angezeigt, `enter` → sauberer Exit.

Kompletter E4-Flow (Palette Aktionen+Bean-Suche+Capture-Order pos./neg., Cockpit-Navigation, alle drei Verdikte a/x/o inkl. explizitem B01-Grenzfall, Quit) end-to-end mit der echten `beans`-Binary belegt.

## Deviations

- Kein Produktionscode-Fix für I01/Q01 nötig (nur Test bzw. Doku, s. Summary) — bewusst kein "erfundener" Fix für einen bereits korrekten Zustand.
- Smoke-Workaround für einen VORBESTEHENDEN, bereits dokumentierten beans-0.4.2-CLI-ETag-Quirk (s. Smoke Schritt 5) — kein beans-tui-Bug, keine Code-Änderung, nur Fixture-Hygiene im Scratch-Repo.
- B01-Fix-Ort: `applyLoaded` gewählt (nicht `keyReviewCockpit`-Kopf) — Begründung s. Summary, beide Optionen waren im Fund gleichwertig zugelassen ("ODER").

## Notes für Epic-Abschluss (bt-tfqi, PO/Supervisor-Gate — NICHT von diesem Task berührt)

- T1-T5 (bt-jpgn/bt-yo60/bt-hxyo/bt-yy6w/bt-v7ti) sind alle `completed`. E4 (Command-Center & Review-Cockpit) ist funktional UND smoke-belegt fertig — inkl. des B01-Nachzüglers aus der T4-Review.
- Offen für den PO/Supervisor: Epic `bt-tfqi` selbst auf `--tag to-review` setzen (NICHT `-s completed` — Epics/Milestones bleiben PO-Gate, Agent schließt sie nie, Review-Flow-Konvention aus CLAUDE.md).
- E5 (Polish: Toast, Help-Overlay, Yank/`internal/clip`, Maus, Settings+Repo-Picker+Lobby, ASCII-Fallback, Archiv-Sicht) ist der nächste Epic, bean `bt-5h4d` bereits im Backlog.

Commit: siehe Refs: bt-v7ti (Test/Fix + docs).
