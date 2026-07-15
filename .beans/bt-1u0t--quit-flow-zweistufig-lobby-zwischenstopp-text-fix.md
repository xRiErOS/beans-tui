---
# bt-1u0t
title: Quit-Flow zweistufig (Lobby-Zwischenstopp) + Text-Fix
status: todo
type: task
priority: normal
created_at: 2026-07-15T21:06:05Z
updated_at: 2026-07-15T21:06:05Z
parent: bt-ntoz
---

E8 Task 5 — deckt B08 aus bean bt-ntoz. Quelle: design-spec.md §15 PF-16. Ist-Code: internal/tui/box_confirm_quit.go, internal/tui/view_lobby.go (nur lesend, keine Aenderung dort noetig). Unabhaengig von allen anderen E8-Tasks (kein blocked_by).

## B08 — Quit-Flow

(A1) Text-Fix: quitBox() (box_confirm_quit.go) zeigt "Really quit bt." (Aussage) -- wird zu "Really quit bt?" (Frage).

(A2) Zweistufige Kaskade: keyConfirmQuit()s enter-Zweig quittet HEUTE unbedingt (`return m, tea.Quit`). NEU:
```go
case keybind.Matches(msg, keys.Enter):
    m.confirmQuit = false
    if m.view != viewLobby && len(m.settings.Repos) > 0 {
        return m.openLobby()
    }
    return m, tea.Quit
```
Deckt exakt die drei PO-Faelle ab:
1. Aus Browse/Backlog, Repos konfiguriert -> Lobby (Stufe 1), TUI beendet NICHT.
2. Bereits in der Lobby (egal ob m.client nil oder nicht -- Stufe 1 wurde bereits passiert bzw. ist irrelevant) -> Exit (Stufe 2).
3. RANDFALL (PO explizit: "konservativ entscheiden + dokumentieren"): Browse/Backlog OHNE konfigurierte Repos (len(m.settings.Repos)==0) -> Exit DIREKT wie bisher (ein Umweg ueber eine leere Lobby, die nur "(no repos in config.yaml...)" zeigen wuerde, waere sinnlos). Diese Entscheidung ist bereits von design-spec.md §15 PF-16 als Planner-Entscheidung dokumentiert -- im Commit-Body NOCHMAL explizit als ERRATUM/Deviation-Absatz festhalten (nicht nur implizit im Code).

WICHTIG -- NICHT anfassen: keyLobby()s eigener "esc"/"q"/"ctrl+c"-Case (view_lobby.go) bleibt UNVERAENDERT (wenn m.client != nil, geht's zurueck zu Browse statt zum Quit-Confirm -- ein SEPARATER, bereits bestehender Pfad, von B08 nicht erwaehnt/betroffen). B08 aendert AUSSCHLIESSLICH keyConfirmQuit()s enter-Verhalten + den quitBox()-Text.

Kleine, klar begruendete Zusatzverbesserung (ueber den Wortlaut hinaus, damit die Modal-Hint-Zeile nicht selbst zur neuen Ueberraschung wird -- lean-stack "kein Ueberraschungsverhalten"): quitBox()s Hint-Zeile "enter: quit   esc: cancel" wird kontextabhaengig -- "enter: go to lobby   esc: cancel" wenn Stufe 1 zutrifft (m.view != viewLobby && len(Repos)>0), sonst "enter: quit   esc: cancel" wie bisher. Im Commit-Body als Planner-Ergaenzung ueber B08 hinaus kennzeichnen.

## TDD-Schritte

1. Failing tests (box_confirm_quit_test.go, ggf. neu anlegen falls nicht vorhanden -- pruefen): TestQuitBoxTextIsQuestion ("Really quit bt?"); TestKeyConfirmQuitEnterGoesToLobbyWhenReposConfiguredAndNotInLobby; TestKeyConfirmQuitEnterQuitsWhenAlreadyInLobby; TestKeyConfirmQuitEnterQuitsWhenNoReposConfigured (Randfall); TestQuitBoxHintTextContextSensitive.
2. command go test ./internal/tui/... -> FAIL.
3. Implementieren: box_confirm_quit.go (keyConfirmQuit, quitBox).
4. command go test ./internal/tui/... -> PASS.
5. Golden-Check: quitBox() ist NICHT Teil der 3 Haupt-Goldens (tree/backlog/chrome rendern kein confirmQuit-Overlay) -- command go test ./internal/tui/ -run "TestTreeGolden|TestBacklogGolden|TestChromeGolden" (OHNE -update) MUSS gruen bleiben. Falls unerwartet FAIL: Ursache klaeren, nicht blind -update.
6. command go test ./... -short gruen (2x), command go test ./... -race gruen, gofmt/vet leer.
7. Manueller Beleg (tmux-Smoke, kurz): bin/bt in tmux, q aus Browse (mit >=1 konfiguriertem Repo) -> Lobby (nicht Exit) -> erneut q+enter aus Lobby -> Exit. Ergebnis im Commit-Body.
8. Commit fix(tui): PF-16 zweistufiger Quit-Flow ueber Lobby + Text-Fix (B08) -- Body dokumentiert den Randfall-Erratum-Absatz + die Hint-Text-Zusatzverbesserung explizit.

## Akzeptanz-Checkliste

- [ ] quitBox() zeigt "Really quit bt?" (Frage)
- [ ] q+enter aus Browse/Backlog mit konfigurierten Repos fuehrt zur Lobby, TUI bleibt offen
- [ ] q+enter aus der Lobby beendet die TUI
- [ ] Randfall (keine konfigurierten Repos) beendet weiterhin direkt -- dokumentiert als bewusste Deviation
- [ ] keyLobby()s eigener esc/q-Case UNVERAENDERT (kein Seiteneffekt)
- [ ] 3 Haupt-Goldens bleiben unveraendert (Verifikation ohne -update)
- [ ] Voller Testlauf (inkl. -race) gruen, gofmt/vet leer
