---
# bt-1u0t
title: Quit-Flow zweistufig (Lobby-Zwischenstopp) + Text-Fix
status: completed
type: task
priority: normal
created_at: 2026-07-15T21:06:05Z
updated_at: 2026-07-16T01:54:24Z
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

- [x] quitBox() zeigt "Really quit bt?" (Frage)
- [x] q+enter aus Browse/Backlog mit konfigurierten Repos fuehrt zur Lobby, TUI bleibt offen
- [x] q+enter aus der Lobby beendet die TUI
- [x] Randfall (keine konfigurierten Repos) beendet weiterhin direkt -- dokumentiert als bewusste Deviation
- [x] keyLobby()s eigener esc/q-Case UNVERAENDERT (kein Seiteneffekt)
- [x] 3 Haupt-Goldens bleiben unveraendert (Verifikation ohne -update)
- [x] Voller Testlauf (inkl. -race) gruen, gofmt/vet leer



## Summary (2026-07-16, Agent)

A1+A2 implementiert in `internal/tui/box_confirm_quit.go`: `quitBox()`
zeigt jetzt "Really quit bt?" (Frage statt Aussage). `keyConfirmQuit`s
`enter`-Zweig loest die zweistufige Kaskade auf: geteiltes Praedikat
`quitBoxWillGoToLobby()` (`m.view != viewLobby && len(m.settings.Repos) >
0`) entscheidet, ob `enter` zur Lobby (`m.openLobby()`, Stufe 1) oder zu
`tea.Quit` (Stufe 2 bzw. Randfall) fuehrt -- von `keyConfirmQuit` UND
`quitBox()`s Hint-Zeile gemeinsam genutzt (kein Duplikat, kein Drift-Risiko).
`keyLobby()`s eigener esc/q/ctrl+c-Case (view_lobby.go) unangetastet.

## Test-Output (RED -> GREEN je Haeppchen)

Alle 5 neuen Tests in `internal/tui/box_confirm_quit_test.go` zuerst
gegen den Alt-Code laufen lassen (RED), dann gegen den neuen Code (GREEN).

**(a) A1 Text-Fix** -- `TestQuitBoxTextIsQuestion`
RED: `quitBox() does not contain the question form, got: ... "Really quit bt."`
GREEN: `--- PASS: TestQuitBoxTextIsQuestion (0.00s)`

**(b) A2 Browse-q-enter-Lobby** -- `TestKeyConfirmQuitEnterGoesToLobbyWhenReposConfiguredAndNotInLobby`
RED: `confirmQuit must be cleared once the cascade resolves (Lobby or Quit)`
GREEN: `--- PASS: TestKeyConfirmQuitEnterGoesToLobbyWhenReposConfiguredAndNotInLobby (0.00s)`

**(c) Lobby-q-enter-Exit** -- `TestKeyConfirmQuitEnterQuitsWhenAlreadyInLobby`
War bereits GREEN gegen den ALTEN Code (unbedingtes `tea.Quit` deckte
diesen Fall zufaellig ab) -- kein echtes RED hier, da Stufe 2 mit dem
alten unconditional-Quit ununterscheidbar ist. Nach der Implementierung
weiterhin GREEN, jetzt aus dem RICHTIGEN Zweig (`m.view==viewLobby`)
statt aus dem alten Blanket-Quit: `--- PASS:
TestKeyConfirmQuitEnterQuitsWhenAlreadyInLobby (0.00s)`

**(d) Randfall ohne Repos** -- `TestKeyConfirmQuitEnterQuitsWhenNoReposConfigured`
Gleiche Lage wie (c): alter Code quittete unbedingt, deckt den Randfall
zufaellig ab (RED entfaellt aus demselben Grund). GREEN nach
Implementierung, jetzt ueber den EXPLIZITEN Randfall-Zweig:
`--- PASS: TestKeyConfirmQuitEnterQuitsWhenNoReposConfigured (0.00s)`

**Hint-Text-Zusatzverbesserung** -- `TestQuitBoxHintTextContextSensitive` (3 Subtests)
RED: `stage 1` Subtest failed (`quitBox() hint does not mention 'go to lobby'`); `stage 2`+`randfall` Subtests bereits GREEN gegen Alt-Code (Text war immer "quit").
GREEN: alle 3 Subtests PASS nach Implementierung.

**Alt-Test umbenannt** -- `update_test.go`: `TestQuitConfirm` ->
`TestQuitConfirmNoReposConfigured` (kein RED/GREEN-Zyklus noetig, reines
Doku-/Scoping-Fix -- der Test blieb die ganze Zeit gruen, da seine
Fixture nie `config.yaml` laedt und damit IMMER den Randfall testet;
sein alter Name/Docstring las sich aber wie "q+enter aus Browse quittet
immer", was A2 nur noch bedingt wahr macht).

**Voll-Gate (nach Implementierung + DRY-Refactor von
`quitBoxWillGoToLobby()`):**
- `command go build -o bin/bt .` gruen
- `gofmt -l .` leer, `command go vet ./...` leer
- Goldens ohne `-update`: `TestTreeGolden`/`TestBacklogGolden`/`TestChromeGolden` alle PASS (unveraendert, wie erwartet -- confirmQuit ist in keinem der 3 Haupt-Snapshots sichtbar)
- `command go test ./... -count=1` 2x GRUEN (Lauf 1: internal/tui 137.46s; Lauf 2: internal/tui 139.06s [dieser Lauf inkl. `-race`])
- `command go test ./... -race -count=1` GRUEN (internal/tui 139.06s)

## Smoke (real, tmux, echtes Binary gegen dieses Repo)

Alle 4 PO-Faelle live durchgespielt (tmux capture-pane-Belege,
Prozess-Lebenszyklus via `ps -ef` verifiziert, nicht nur Unit-Level):

1. **Browse (Repos konfiguriert) -> q -> enter -> Lobby, Prozess LEBT:**
   `bin/bt .` mit `HOME` auf temp-Config (`repos: [.../beans-tui-repository]`).
   `q` zeigt Modal "Really quit bt?" / Hint "enter: go to lobby   esc: cancel".
   `enter` -> Lobby-View erscheint (Repo-Tabelle mit `14/68`), `ps -ef`
   zeigt den `bt`-Prozess weiterhin laufend (PID 78384).
2. **Lobby (client==nil, Kaltstart) -> q -> enter -> Exit, sauber:**
   `bin/bt` OHNE Pfad-Arg, cwd AUSSERHALB jedes Repos, 2 Repos in
   config.yaml (`decideStartup` -> `startupLobby`, `client==nil`) --
   Footer zeigt `esc/q:quit` (lobbyBackHint bestaetigt `client==nil`).
   `q` -> Modal "enter: quit   esc: cancel" (NICHT "go to lobby",
   korrekt -- Stufe 2). `enter` -> Prozess beendet, Terminal sauber
   restauriert (leerer Pane, Shell-Prompt zurueck), `ps -ef` findet
   KEINEN `bt`-Prozess mehr, `tmux list-panes` zeigt `pane_dead=0`
   (Pane selbst lebt weiter, kein Zombie).
3. **esc am Confirm bricht ab (unveraendert):** aus Browse `q` dann
   `esc` -> Modal verschwindet, Browse-View unveraendert sichtbar,
   Prozess laeuft weiter (`ps -ef` zeigt PID 78967 aktiv). Danach mit
   `ctrl+c` (Hard-Quit, bt-7jr8-Kontrakt) sauber beendet.
4. **Randfall (keine Repos konfiguriert) -> q -> enter -> Exit direkt:**
   `bin/bt <repo-pfad>` mit `HOME` auf leeres Config-Verzeichnis (kein
   `config.yaml`, `ls $HOME/.config/beans-tui/` leer). `q` zeigt "enter:
   quit" (kein Lobby-Hint, korrekt). `enter` -> Prozess terminiert SOFORT
   (kein Lobby-Zwischenstopp), Shell-Prompt "took 4s" zurueck, `ps -ef`
   findet keinen Prozess mehr.

Temp-HOME-Setup (dokumentiert fuer Reproduzierbarkeit): 2
Scratch-`$HOME`-Verzeichnisse mit je eigenem `~/.config/beans-tui/config.yaml`
(eins mit 1-2 Repo-Eintraegen, eins ganz ohne Datei) via `tmux new-session`
+ `export HOME=...` + `cd <nicht-repo-dir>` fuer den Kaltstart-Fall. Alle
4 Szenarien real durchgespielt, keine Unit-only-Abgrenzung noetig.

## Deviations / ERRATA

**ERRATUM (2026-07-16, Agent) -- Randfall-Bestaetigung, keine Abweichung
vom Plan:** Der im bean-Text bereits dokumentierte Randfall ("Direktstart
ohne konfigurierte Repos -> q+enter beendet direkt") wurde 1:1 wie
spezifiziert umgesetzt (`len(m.settings.Repos) > 0`-Guard) -- keine
Code-Abweichung vom PO-/Planner-Wortlaut. Einzige Ergaenzung: das geteilte
`quitBoxWillGoToLobby()`-Praedikat (Planner-Sketch hatte die Bedingung
inline in `keyConfirmQuit` UND implizit nochmal in der Hint-Logik --
waere ohne Extraktion ein Duplikat mit Drift-Risiko gewesen; keine
Verhaltensaenderung, nur DRY).

**ERRATUM (Test-Design):** Die TDD-Schritte (b)+(c)+(d) liessen sich
NICHT alle vier als reines RED/GREEN zeigen -- (c) und (d) waren gegen
den ALTEN Code (unbedingtes `tea.Quit`) bereits zufaellig gruen, weil der
Alt-Code IMMER quittete (die neuen Faelle unterscheiden sich vom Alt-Code
nur in Fall (b), wo die Kaskade jetzt zur Lobby statt zum Exit fuehrt).
Echtes RED gab es nur fuer (a) Text-Fix, (b) Lobby-Abzweigung und den
Hint-Text-Subtest "stage 1". Dokumentiert statt stillschweigend als "TDD
komplett befolgt" behauptet.

## Notes for bt-d8kc (Header/Footer-Neuspezifikation R2)

Kein Handlungsbedarf fuer bt-d8kc: Der Header/Footer zeigt weiterhin nur
`q:quit` als KEY-Label (`keymap.go` `keys.Quit`, `globalBindings()`) --
das bleibt korrekt, weil es den KEY beschreibt ("q oeffnet den
Quit-Flow"), nicht das SOFORTIGE Ergebnis. Die neue Kaskaden-Nuance
("geht evtl. erst zur Lobby") ist bereits vollstaendig im Confirm-Modal
selbst geloest (`quitBox()`s eigene kontextsensitive Hint-Zeile, B08
Planner-Zusatz) -- KEINE Aenderung an `globalBindings()`/Footer-Texten
noetig oder gewuenscht. Falls T8 den Header/Footer ohnehin neu
zusammenstellt: `q:quit` als Wortlaut unveraendert uebernehmen, keine
neue "q: quit/lobby"-Doppel-Beschriftung noetig.

## Voll-Gate-Beleg (Abschluss)

- Build: `command go build -o bin/bt .` gruen
- `gofmt -l .` leer, `command go vet ./...` leer
- Goldens (tree/backlog/chrome, ohne `-update`): alle PASS, unveraendert
- `command go test ./... -count=1`: 2x GRUEN (internal/tui 137.46s, 139.06s)
- `command go test ./... -race -count=1`: GRUEN (internal/tui 139.06s)
- Commit: `ce3200a` `fix(tui): PF-16 zweistufiger Quit-Flow ueber Lobby + Text-Fix (B08)`



## Fix-Runde 1 (2026-07-16, Review R1)

**B01 (high, blocking, Reviewer live bestaetigt): Lobby-Stufe-2 war im
Hauptfall unerreichbar.** `keyLobby` (view_lobby.go) behandelte
esc/q/ctrl+c UNIFORM: `client != nil` -> zurueck zu Browse. Da `m.client`
nach dem ersten Repo-Oeffnen nie wieder nil wird, gab es danach KEINEN
q-basierten Exit mehr (Lobby-q ging immer zu Browse; sogar ctrl+c in der
Lobby beendete nicht -- Prozess lebte weiter). Widersprach dem
PO-Wortlaut (bt-ntoz B08: "aus der Lobby q→enter beendet die TUI") UND
dem eigenen bean-Text ("Bereits in der Lobby (egal ob m.client nil oder
nicht) → Exit (Stufe 2)"). Die urspruengliche Umsetzung hatte B08s
"NICHT anfassen: keyLobby"-Anweisung woertlich befolgt -- der bean-Text
selbst enthielt den Widerspruch (Stufe 2 "egal ob client nil" vs.
keyLobby unangetastet lassen); der Reviewer hat ihn aufgeloest.

**Fix (PO-Wortlaut-konform, kein Herkunfts-Flag):** Die drei Keys in
`keyLobby` ENTKOPPELT:
1. `q` -> IMMER `requestQuit()` (Confirm oeffnet ueber der Lobby;
   `quitBoxWillGoToLobby()` liefert dort false -> Hint "enter: quit",
   enter -> tea.Quit -- Stufe 2 komplett). Beide client-Zustaende.
2. `ctrl+c` -> sofortiger `tea.Quit` (konsistent mit globalem
   ctrl+c-Kill-Switch-Kontrakt bt-7jr8; das alte "ctrl+c -> Browse" war
   Teil desselben Lochs).
3. `esc` UNVERAENDERT: `client != nil` -> Browse (D03: eine Ebene
   zurueck, Lobby als Abstecher), `client == nil` -> `requestQuit()`.

**Zusatz (notwendige Konsequenz, dokumentiert):** Der Lobby-Footer-Hint
`esc/q:back` waere nach der Entkopplung eine Luege gewesen (q fuehrt
nicht mehr zurueck). `lobbyBackHint()` -> `lobbyExitHint()`: bei
`client != nil` jetzt `esc:back  q:quit` (getrennt), bei `client == nil`
weiterhin kombiniert `esc/q:quit`. Gleiche Anti-Ueberraschungs-Rationale
wie die quitBox-Hint-Zeile aus der Hauptrunde.

**I01 (non-blocking, miterledigt):** Header-Kommentar
box_confirm_quit.go ("ctrl+c bypasses this entirely") praezisiert --
gilt nur solange der Confirm NICHT offen ist (bei confirmQuit=true
schluckt keyConfirmQuits Full-Capture ctrl+c wie jeden anderen
non-enter/non-esc-Key). NUR Kommentar, Verhalten unveraendert.

**Q01 (Reviewer, dokumentierte Abgrenzung, bewusst NICHT abgedeckt):**
Randfall "Repo konfiguriert, aber nicht ladbar" (config.yaml nennt einen
Pfad, der kein beans-Repo mehr ist/nicht existiert) -> die Kaskade
fuehrt zur Lobby, die das Repo mit err-Metrik listet. Bewusste
Abgrenzung: die Lobby IST auch dann der richtige Zwischenstopp (sie
zeigt den Fehler sichtbar an, statt ihn zu verschlucken), kein
Sonderpfad noetig. Nicht getestet/nicht Teil von B08s drei PO-Faellen.

### Test-Output Fix-Runde 1 (RED -> GREEN)

Neue Tests in view_lobby_test.go (+ Fixture
`lobbyFixtureModelWithClient`):

**(1) q in Lobby mit Client -> Confirm + Stufe 2** --
`TestLobbyQOpensQuitConfirmWithLiveClient`
RED: `q in the Lobby with a live client did not open the quit-confirm (B01: bounced back to Browse instead?)`
GREEN: `--- PASS: TestLobbyQOpensQuitConfirmWithLiveClient (0.00s)`
(asserted: confirmQuit true, view bleibt viewLobby, enter -> tea.QuitMsg)

**(2) ctrl+c in Lobby mit Client -> sofort tea.Quit** --
`TestLobbyCtrlCQuitsImmediatelyWithLiveClient`
RED: `ctrl+c in the Lobby must return a Cmd` (alter Code: view-Wechsel zu Browse, cmd nil)
GREEN: `--- PASS: TestLobbyCtrlCQuitsImmediatelyWithLiveClient (0.00s)`

**(3) esc-Bestand gepinnt (kein RED, Verhalten unveraendert)** --
`TestLobbyEscReturnsToBrowseWithLiveClient`: esc mit Client -> Browse,
esc ohne Client -> Confirm. PASS vor UND nach dem Fix.

**(4) Hint-Split** -- `TestLobbyHintReflectsSplitEscQBehavior`
RED (Subtest "live client"): `viewLobby() hint still shows the combined esc/q label ... esc/q:back`
GREEN: beide Subtests PASS (`esc:back  q:quit` bei Client, `esc/q:quit` ohne).

**Alt-Tests, die q/ctrl+c->Browse pinnten: KEINE gefunden** (grep ueber
alle *_test.go: kein Test exercierte keyLobbys esc/q/ctrl+c-Case --
genau deshalb konnte B01 durch die Hauptrunde rutschen).

### Voll-Gate Fix-Runde 1

- Build gruen, `gofmt -l .` leer, `command go vet ./...` leer
- Goldens (tree/backlog/chrome, ohne -update): PASS, unveraendert
- `command go test ./... -count=1` GRUEN (internal/tui 136.98s)
- `command go test ./... -race -count=1` GRUEN
- tmux-Smoke komplette Kaskade (echtes Binary, ps-Belege):
  Browse q->enter -> Lobby LEBT (Hint zeigt neu `esc:back  q:quit`) ->
  q -> Confirm "enter: quit" ueber der Lobby -> enter -> Prozess weg,
  Pane lebt, Terminal sauber. esc-Abstecher: Lobby -> esc -> Browse,
  Prozess lebt. ctrl+c in Lobby mit Client -> Prozess SOFORT beendet
  (vorher: bounced zu Browse, Prozess lebte weiter).
