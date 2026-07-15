---
# bt-pd22
title: E5 T6b — TreeWidth-Wiring + SaveUserSettings-Doku (T5-Review-Findings)
status: completed
type: task
priority: normal
created_at: 2026-07-15T11:49:11Z
updated_at: 2026-07-15T12:48:13Z
parent: bt-5h4d
blocking:
    - bt-7dfj
---

Ziel: Die zwei T5-Review-Findings schließen (Reviewer 2026-07-15, bean bt-0l8c APPROVED mit I01/I02).

## I01 (medium) — TreeWidth ohne Render-Wirkung

Settings.Layout.TreeWidth wird persistiert+validiert (internal/config/settings.go), aber clickPaneGeometry (internal/tui/mouse.go:67) hardcodet weiter 24. Stiller No-Op für die PO. Task-5-Plan-Scope hat das bewusst ausgeklammert, Task 8 schließt es NICHT — diese Lücke hier schließen:

- clickPaneGeometry (Single-Source aus T4!) bezieht treeWidth aus m.settings.Layout.TreeWidth (Fallback 24 wenn 0/ungesetzt, Clamp-Range aus validateSettings respektieren)
- Live-Apply: submitForm-"settings"-Case wendet TreeWidth ohne Neustart an (analog Accent)
- Tests: TestTreeWidthFromSettingsAffectsGeometry, TestTreeWidthZeroFallsBackToDefault; Goldens byte-identisch (Default 24 unverändert!)

## I02 (low) — SaveUserSettings-Doku überzeichnet

Doc-Kommentar behauptet Erhalt fremder config.yaml-Keys; tatsächlich überlebt ein unbekannter YAML-Key den getypten Unmarshal/Marshal-Roundtrip NICHT (Reviewer experimentell verifiziert; devd-Original identisch). Kommentar präzisieren: "bewahrt zukünftige Settings-Struct-Felder, NICHT beliebige fremde YAML-Keys". Kein yaml.Node-Umbau (YAGNI).

## Akzeptanz

[x] clickPaneGeometry nutzt Settings-TreeWidth (Fallback 24), Live-Apply im settings-Submit
[x] 2 neue Tests grün, 7 Goldens -count=2 byte-identisch
[x] SaveUserSettings-Doc-Kommentar präzisiert
[x] command go test ./... voll grün, gofmt/vet leer
[x] Commit fix(tui): TreeWidth-Setting wirkt im Render (T5-Review I01/I02, T6-Review I01)


## Prelude aus T6-Review (I01, low — Reviewer 2026-07-15)

applyRepoSwitched (update.go) resettet m.toast nicht — Toast aus altem Repo bleibt nach Repo-Wechsel kurz sichtbar. Fix hier miterledigen: m.toast=nil (bzw. Muster der übrigen Reset-Felder) in applyRepoSwitched + Regressionstest TestRepoSwitchClearsToast.



## Summary

Alle 3 Findings geschlossen. clickPaneGeometry (internal/tui/mouse.go) nimmt neu einen `treeWidth int`-Parameter (jeder der sechs Call-Sites in view_browse_repo.go/view_review_cockpit.go/view_browse_backlog.go/mouse_test.go übergibt `m.settings.Layout.TreeWidth`); neuer Helper `treeWidthFloor` löst 0/ungesetzt -> 24 auf (Fallback identisch zum alten hardcoded Wert) und clampt sonst in [24,60] (konsistent zu config.validateSettings). Live-Apply ergab sich automatisch aus der bestehenden `m.settings = settings`-Zuweisung in submitForm's "settings"-Case (box_confirm_create.go) — TreeWidth wird anders als Accent direkt aus dem Model gelesen, kein zusätzlicher Apply-Call nötig. SaveUserSettings-Doc-Kommentar (internal/config/settings.go) präzisiert: bewahrt zukünftige Settings-Struct-Felder, nicht beliebige fremde YAML-Keys (kein Code-Umbau). applyRepoSwitched (internal/tui/update.go) setzt `m.toast = nil` unconditional (auch sticky) bei erfolgreichem Repo-Wechsel — bewusst NICHT über clearToastUnlessSticky (das schützt nur SAME-repo Reloads).

## Test-Output

RED (vor Implementierung, git-checkout der jeweiligen Produktivdateien + go test):
- TestRepoSwitchClearsToast: FAIL — `m.toast = &{...} after a successful repo switch, want nil`
- TestTreeWidthZeroFallsBackToDefault / TestTreeWidthFromSettingsAffectsGeometry: Compile-FAIL — `too many arguments in call to clickPaneGeometry, have (int,int,string,string,int) want (int,int,string,string)` (Signatur-Erweiterung ist der eigentliche Fix)

GREEN (nach Implementierung):
- `command go test ./internal/tui/... -run TestRepoSwitchClearsToast -v` → PASS
- `command go test ./internal/tui/... -run 'TestTreeWidth' -v` → TestTreeWidthZeroFallsBackToDefault PASS, TestTreeWidthFromSettingsAffectsGeometry PASS
- `command go test ./...` (voller Lauf, ohne -short) → alle Pakete ok (internal/tui 139.6s)
- `command go vet ./...` → leer
- `gofmt -l .` → leer
- `command go build -o bin/bt .` → OK
- 7 Goldens `-count=2`: TestChromeGolden/TestTreeGolden/TestTreeGoldenDeterministic/TestBacklogGolden/TestBacklogGoldenDeterministic/TestReviewCockpitGolden/TestReviewCockpitGoldenDeterministic → alle 14 Läufe PASS, byte-identisch (testdata/*.golden unverändert lt. git status)

## Smoke

tmux, Scratch-Repo (`beans init` + 2 Tasks), Temp-HOME (`~/.config/beans-tui/config.yaml` frisch):
1. TUI Start, ctrl+k → "settings" → Enter → Settings-Form öffnet (tree_width prefilled "36", DefaultSettings)
2. tree_width auf "34" gesetzt, enter → Form schließt, KEIN Neustart nötig. Terminal 90x40: linke Pane Border-zu-Border = 36 Spalten = lw 34 + 2 Border (sofortige Wirkung bestätigt)
3. `q` → enter (Quit-Confirm) → Prozess beendet; `~/.config/beans-tui/config.yaml` enthält `layout: tree_width: 34` (persistiert)
4. TUI neu gestartet (selbe Temp-HOME) → linke Pane wieder 36 Spalten breit (Persistenz über Neustart bestätigt)
5. Settings-Form erneut geöffnet, tree_width komplett geleert, Enter → Form BLOCKT Submit ("expected a number", huh-Validator validateTreeWidth, box_form_settings.go — PRE-EXISTING, nicht Teil dieses Tasks) → 0/leer ist über die Live-UI nicht erreichbar, der 24-Fallback ist ausschließlich über Pre-LoadSettings-Zero-Value-Modelle (Goldens/Tests) beobachtbar, dort abgedeckt durch TestTreeWidthZeroFallsBackToDefault
6. tree_width auf "24" (untere Clamp-Grenze) gesetzt, submit, Terminal auf 66x40 verkleinert (schmal genug, dass der Floor tatsächlich bindet) → linke Pane Border-zu-Border = 26 Spalten = lw 24 + 2 Border (Clamp-Untergrenze live bestätigt)

## Deviations

- Live-Apply brauchte KEINE Code-Änderung in box_confirm_create.go — `m.settings = settings` existierte bereits aus T5 (bt-0l8c); TreeWidth wird pro Render direkt aus m.settings gelesen (anders als Accent, das globalen theme.SetAccent-State braucht), daher automatisch live.
- Der im Bean skizzierte Smoke-Schritt "tree_width leeren/0 → Default 24" ist über die laufende UI NICHT erreichbar: validateTreeWidth (box_form_settings.go, pre-existing, unverändert) blockt eine leere Eingabe vor StateCompleted. Der 0-Fallback (treeWidthFloor) ist nur für Pre-LoadSettings-Zero-Value-Modelle relevant (jedes golden/fixture Model vor app.go Run()) — dort per TestTreeWidthZeroFallsBackToDefault abgedeckt; im Smoke stattdessen die untere Clamp-Grenze (24) live verifiziert.
- LayoutSettings-Doc-Kommentar (internal/config/settings.go) zusätzlich aktualisiert (war als "NOT wired into the live render yet" stale nach diesem Fix) — kleine Doku-Korrektur über den Bean-Scope hinaus, kein Code-Umbau.
