---
# bt-zhwl
title: E5 T6 — Lobby V1 + Repo-Picker p
status: completed
type: task
priority: normal
created_at: 2026-07-15T09:04:38Z
updated_at: 2026-07-15T12:26:47Z
parent: bt-5h4d
---

Ziel: Lobby V1 (ASCII-Logo `beans` + Repo-Picker, Port devd `view_home.go`)
+ globaler Repo-Wechsel `p`. Design decision d (Trigger): Lobby erscheint
NUR beim Start, wenn Settings.Repos >=2 Einträge hat UND kein `bt <pfad>`-
Arg gegeben wurde UND data.FindRepo(cwd) fehlschlägt (design-spec §3.2:
"Ohne Config oder mit cwd-Treffer: direkt ins Repo") -- sonst Direkt-Start
wie bisher (E1-E4 unverändert). `p` (keys.Picker, bereits im Keymap seit E1
T7 + Drift-Guard-abgedeckt) wechselt `viewID` zu `viewLobby` als 4. View
(Elm-Konvention: ein weiterer Case im View()-Switch, KEIN Overlay-Bool --
Lobby braucht einen kompletten Client-/Watcher-Neustart, das ist kein
schwebendes Modal). Architektur-Kern: Repo-Wechsel muss den laufenden
fsnotify-Watcher SAUBER stoppen+neu starten (app.go besitzt aktuell den
stop-func nur als lokale Variable, unsichtbar für Update()) -- gelöst über
eine testbare, tea.Program-entkoppelte switchRepoCmd-Kern-Funktion
(injizierbares notify func(), Produktion bindet p.Send).

Plan: docs/plans/v1-port/epic-E5-plan.md »Task 6«.

## Akzeptanz
- [x] internal/tui/types.go: viewLobby viewID (4. Case), model.repoQuery/
      repoSearch (textinput)/repoList (listState)/watchStop func() (NEU --
      hält den AKTUELL laufenden Watcher-Stop, damit ein Repo-Wechsel ihn
      aufrufen kann)
- [x] internal/tui/view_lobby.go (NEU): homeLogoBlock (ASCII "beans"-Banner,
      EAW-neutral, Port devd view_home.go-Muster), repoPickerBody(w)
      (Suchzeile + gefilterte Settings.Repos-Liste + Offen/Gesamt-Metrik
      via einem leichten `beans list --json -s todo -s in-progress -s draft
      --quiet | wc -l`-artigen Zähl-Aufruf ODER client.List()+Index einmalig
      pro Repo -- Entscheidung + Kosten/Latenz-Abwägung bei Task-Start
      dokumentieren), viewLobby() (Chrome + zentrierter Block, Port
      centerInto/pickerRowFill VERBATIM)
- [x] internal/data (neue Datei watch_lifecycle.go ODER Erweiterung
      watcher.go): startWatch(repoDir string, notify func()) (stop func(),
      err error) -- dünner, testbarer Wrapper um data.Watch (identische
      Signatur, EIGENER Name nur zur Trennung Produktions-Wiring vs.
      Testbarkeit)
- [x] internal/tui/messages.go: repoSwitchedMsg{client *data.Client,
      repoDir string, beans []data.Bean, watchStop func(), err error};
      switchRepoCmd(oldStop func(), newRepoDir string, notify func())
      tea.Cmd -- ruft oldStop() (falls != nil), baut neuen data.Client,
      lädt List(), startet neuen Watch, gibt repoSwitchedMsg zurück (KEIN
      tea.Program-Bezug in dieser Funktion selbst -- notify ist injiziert,
      testbar ohne echten Bubbletea-Runtime)
- [x] internal/tui/app.go Run(): package-level `var activeProgram
      *tea.Program` (oder Closure-Var) gesetzt direkt nach
      tea.NewProgram(...), initialer data.Watch-Aufruf liefert seinen
      stop-func jetzt an m.watchStop (Modell trägt ihn, nicht mehr NUR
      lokale Run()-Variable) -- switchRepoCmd's notify-Parameter wird mit
      `func(){ activeProgram.Send(watchMsg{}) }` gefüllt
- [x] internal/tui/update.go: `case repoSwitchedMsg:` -- m.client/m.repoDir/
      m.watchStop/idx (via loadCmd oder direkt msg.beans) neu setzen,
      m.view = viewBrowseRepo, Cursor/Expand-State zurücksetzen (Port devd
      selectProject-Reset-Muster); handleKey: `keys.Picker` (`p`) von
      ÜBERALL (design decision h, wie ctrl+k/?) -> öffnet viewLobby (lädt
      Settings.Repos neu, falls seit Start geändert)
- [x] internal/config/state.go Nutzung: beim Repo-Wechsel SetLastRepo(newDir)
      (Read-Modify-Write, Port devd DD2-273-Musterkommentar: andere
      State-Felder bleiben erhalten)
- [x] cmd/tui.go RunE: Startup-Trigger-Logik (design decision d) -- Settings
      laden, `bt <pfad>`-Arg > data.FindRepo(cwd)-Erfolg > (>=2 Repos in
      Settings -> Lobby) > (1 Repo/kein Repo -> Fehlermeldung wie bisher,
      US-01 unverändert)
- [x] overlay_palette.go paletteActions: neuer globaler Eintrag `{actionID:
      "repo_picker", label: "repo: wechseln"}`
- [x] view_lobby_test.go: TestLobbyShowsConfiguredRepos,
      TestLobbyFilterNarrowsBySearch, TestLobbySelectSwitchesRepoAndView,
      TestPickerKeyOpensLobbyFromAnyView, TestNoLobbyOnSingleRepoCwdMatch
      (design decision d Kernfall -- E1-E4-Verhalten bleibt intakt)
- [x] switch_repo_test.go (internal/tui oder internal/data): TestSwitchRepo
      CmdStopsOldWatcherStartsNew (fake notify func(), zwei newTestRepo(t)-
      Instanzen, Datei-Touch im NEUEN Repo -> notify gefeuert; Datei-Touch
      im ALTEN Repo NACH Switch -> notify NICHT gefeuert, alter Watcher tot)
- [x] `command go test ./... -short` grün, gofmt/vet leer
- [x] Commit `feat(tui): Lobby V1 + Repo-Picker p (Watcher-Lifecycle-Switch)`

## Summary

Lobby V1 (`viewLobby`, 4. top-level View) + globaler Repo-Picker `p`
implementiert, inkl. der Kernschwierigkeit: ein voll getesteter,
tea.Program-entkoppelter Watcher-Lifecycle-Switch (`switchRepoCmd`,
messages.go). Neue Dateien: `internal/tui/view_lobby.go` (+Test),
`internal/tui/switch_repo_test.go`, `cmd/tui_test.go`. Geänderte Dateien:
`internal/data/watcher.go` (`StartWatch`-Wrapper), `internal/tui/types.go`
(viewLobby + Lobby-Felder inkl. `repoMetrics`), `internal/tui/messages.go`
(`initialWatchMsg`, `repoSwitchedMsg`/`switchRepoCmd`, `repoMetricsMsg`/
`repoMetricsCmd`/`repoMetricsBatchCmd`), `internal/tui/app.go`
(`activeProgram`, `Run()`-Signatur +`settings`-Param, Lobby-first-Start,
finaler-Watcher-Stop bei Quit), `internal/tui/update.go` (Capture-Order +
`applyRepoSwitched`/`applyRepoMetrics`), `internal/tui/mouse.go`
(Overlay-Guard `|| m.view == viewLobby`), `internal/tui/view_browse_repo.go`
(View()-Switch-Case), `internal/tui/overlay_palette.go` ("repo: wechseln"),
`cmd/tui.go` (`decideStartup`, design decision d).

## Test-Output

RED (vor Implementierung, Kern-Watcher-Lifecycle-Test gegen fehlende Typen):
```
$ command go test ./internal/tui/ -run TestSwitchRepoCmd
# beans-tui/internal/tui [beans-tui/internal/tui.test]
./switch_repo_test.go:XX:XX: undefined: switchRepoCmd
./switch_repo_test.go:XX:XX: undefined: repoSwitchedMsg
FAIL
```
(analog für `view_lobby_test.go` gegen `viewLobby`/`openLobby`/`keyLobby`
vor deren Implementierung.)

GREEN, `-race`, Kern-Watcher-Lifecycle + Lobby (Pflicht-Scope):
```
$ command go test ./internal/data/ ./internal/tui/ -race -run 'Watch|Repo|Lobby' -v
=== RUN   TestWatcherFiresOnceForBurst          --- PASS
=== RUN   TestWatcherStops                      --- PASS
=== RUN   TestWatcherStopDuringBurst             --- PASS
=== RUN   TestWatcherPicksUpLateArchiveDir       --- PASS
ok      beans-tui/internal/data        1.885s
=== RUN   TestDeleteLastBeanInRepoClearsCursorGracefully           --- PASS
=== RUN   TestConflictAfterWatchReloadUsesFreshETagNoConflict      --- PASS
=== RUN   TestSwitchRepoCmdStopsOldWatcherStartsNew                --- PASS (0.91s)
=== RUN   TestSwitchRepoCmdKeepsOldWatcherAliveOnValidationFailure --- PASS (0.33s)
=== RUN   TestLobbyShowsConfiguredRepos                            --- PASS
=== RUN   TestLobbyFilterNarrowsBySearch                           --- PASS
=== RUN   TestLobbySelectSwitchesRepoAndView                       --- PASS
=== RUN   TestPickerKeyOpensLobbyFromAnyView (+3 subtests)         --- PASS
=== RUN   TestViewLobbyFrameMatchesWidthHeight                     --- PASS
=== RUN   TestNoLobbyOnSingleRepoCwdMatch                          --- PASS
ok      beans-tui/internal/tui        3.300s
```

Goldens, `-count=2`, byte-identical:
```
$ command go test ./internal/tui/ -run 'TestChromeGolden|TestTreeGolden|TestTreeGoldenDeterministic|TestBacklogGolden|TestBacklogGoldenDeterministic|TestReviewCockpitGolden|TestReviewCockpitGoldenDeterministic' -count=2
ok      beans-tui/internal/tui        0.287s
$ git status --porcelain internal/tui/testdata/
(leer)
```

`decideStartup` (design decision d, cmd/tui.go), pur+isoliert getestet
(kein `tui.Run`-Aufruf -- interaktives AltScreen-Programm, in `go test`
nicht sicher fahrbar):
```
$ command go test ./cmd/... -run TestDecideStartup -v
--- PASS: TestDecideStartupPrioritiesInOrder (8 Subtests, alle PASS)
```

Voller Lauf (ohne `-short`), `vet`, `gofmt`:
```
$ command go vet ./...      # leer
$ gofmt -l .                 # leer
$ command go test ./...
ok      beans-tui/cmd            0.290s
ok      beans-tui/internal/config (cached)
ok      beans-tui/internal/data   (cached)
ok      beans-tui/internal/theme  (cached)
ok      beans-tui/internal/tui    139.205s
```

Regressions gefunden+gefixt während der Arbeit (siehe Deviations):
`TestPaletteActionsNoFocusedBeanOmitsNodeActions` (erwartete Liste um
`repo_picker` ergänzt) und `TestKeyReviewCockpitNavigationMovesCursor`
(Capture-Order-Fix für `p` im Review-Cockpit) -- beide grün nach Fix.

## Smoke

tmux (100x30), `bin/bt` frisch gebaut, `HOME=/tmp/bt-t6-home` (isoliert),
zwei Scratch-beans-Repos (`repo-alpha`: 2 Tasks, `repo-beta`: 1 Task),
`config.yaml` mit beiden Pfaden.

**(a) Start ohne Arg außerhalb eines Repos, >=2 Repos konfiguriert:**
Lobby erscheint sofort -- "beans"-ASCII-Banner, "Repo-Picker"-Untertitel,
beide Repos gelistet mit korrekter Offen/Gesamt-Metrik (2/2, 1/1). PASS.

**(b) Auswahl -> Browse lädt Repo:** `enter` auf repo-alpha -> Browse zeigt
`> repo-alpha: Browse` + beide Alpha-Tasks im Tree. `cat state.json` zeigt
`{"last_repo": "/tmp/bt-t6-scratch/repo-alpha"}`. PASS.

**(c) `p` aus Browse -> Lobby -> anderes Repo -> Wechsel + Watcher folgt:**
`p` aus der laufenden repo-alpha-Session öffnet die Lobby (Footer korrekt
`esc/q:back`, nicht `quit` -- lobbyBackHint). Cursor auf repo-beta, `enter`
-> Browse zeigt `> repo-beta: Browse` + den Beta-Task. `beans create` (real,
außerhalb der TUI) in repo-beta -> neuer Task erscheint im Tree OHNE
manuelles `ctrl+r` (Watcher ist dem Wechsel gefolgt). PASS.

**(d) Start IM Repo-Verzeichnis -> keine Lobby (US-01 unverändert):** `bt`
aus `repo-alpha/` gestartet (trotz 2 konfigurierter Repos) -> direkt Browse,
keine Lobby. PASS.

**(e) Wechsel auf kaputtes Verzeichnis -> Fehler-Toast, Session lebt:**
config.yaml auf `[repo-beta, does-not-exist]` editiert (Datei, während TUI
lief), `p` aus der laufenden repo-alpha-Session -> Lobby zeigt SOFORT die
FRISCH geladene Liste (nicht die beim Start geladene) inkl. `err`-Metrik für
den kaputten Pfad -- fand den fehlenden Settings-Reload-Bug live, siehe
Deviations. Auswahl von `does-not-exist` -> roter Toast oben rechts
"Repo-Wechsel fehlgeschlagen: …", Lobby bleibt offen, KEIN Crash. `esc`
zurück -> repo-alpha-Session unverändert live (Tree/Detail intakt). PASS.

Kein Panic/Goroutine-Dump in der gesamten Sitzung (Pane-Capture geprüft).

## Deviations/ERRATA

- **Reihenfolge validate-then-stop statt stop-then-validate (bewusste
  Abweichung vom Plan-Sketch, Task 6 Step 3).** Der Plan-Pseudocode ruft
  `oldStop()` unbedingt zuerst; implementiert wurde stattdessen: NEUEN
  Client validieren (`client.List()` muss erfolgreich sein) -- ERST DANN
  `oldStop()`. Grund: die Bean-eigene "ACHTUNG"-Anmerkung verlangt bewusste
  Entscheidung genau hier -- stop-then-validate hätte einen gesunden alten
  Watcher für einen fehlschlagenden Wechsel geopfert (Session ohne JEDEN
  Watcher zurückgelassen). Regressionstest:
  `TestSwitchRepoCmdKeepsOldWatcherAliveOnValidationFailure`.
- **`p` NICHT von überall -- Ausnahme Review-Cockpit.** bt-0l8c's
  Notes-für-T6 skizzierten `keys.Picker`-Match VOR
  `m.view==viewReviewCockpit` (wie ctrl+k/`?`). Beim Testen gegen den
  echten `keyReviewCockpit`-Code gefunden: die Cockpit bindet `p` bereits
  auf "explicit prev" (design-spec §7, VOR diesem Task existent) --
  `TestKeyReviewCockpitNavigationMovesCursor` schlug fehl. Aufgelöst:
  `keys.Picker`-Match jetzt NACH dem Cockpit-Capture-Block (Cockpits
  eigenes `p` gewinnt dort), Lobby bleibt aus der Cockpit trotzdem über
  `ctrl+k -> "repo: wechseln"` erreichbar (Command-Center, zweiter
  Eintrittspunkt zum selben Handler). Test:
  `TestPickerKeyOpensLobbyFromAnyView` (Subtest deckt den Cockpit-Fall
  explizit als Negativ-Fall ab).
- **`m.view==viewLobby` VOR `keys.Picker`-Match, nicht danach (zweite
  Abweichung von bt-0l8c's Notiz).** Grund: die Lobby-eigene
  `repoQuery`-Suche muss "p" als normales Zeichen tippen können (reale
  Repo-Pfade enthalten "p" häufig, z.B. dieses Repo selbst via
  "rePository"). Mit `keys.Picker`-Match zuerst hätte JEDE "p"-Eingabe die
  Lobby zurückgesetzt statt gefiltert. Mirrort das bereits im Code
  etablierte Muster (m.helpOpen-vs-keys.Palette-ERRATUM).
- **`openLobby()` lädt Settings NEU von Disk (live im Smoke-Test (e)
  gefunden, nicht in der Akzeptanz-Checkliste als eigene Box, aber deren
  Prosa-Text "lädt Settings.Repos neu, falls seit Start geändert" verlangt
  es explizit).** Erste Implementierung vergaß den Reload -- die Lobby
  zeigte beim Wiederöffnen die beim Prozessstart geladene, STALE Liste.
  Gefixt: `config.LoadSettings()` bei jedem `openLobby()`-Aufruf (mirrort
  Run()s eigenes "korrupte/fehlende config.yaml -> Defaults"-Fallback).
- **Layout-Bug (B01, live im selben Smoke-Schritt (e) gefunden): `viewLobby`
  überlief `m.height`/`m.width` um je 2 Zeilen/Spalten.** Erste
  Implementierung übergab das VOLLE `w`/`h` an `lipgloss.Place`/
  `outerBorder`, statt (wie `viewBrowseRepo`/`Chrome()` es bereits
  etabliert) `innerW := w-2`/`innerH := h-2` VOR dem Border-Wrap zu
  budgetieren. Sichtbar wurde es NICHT in Unit-Tests (kein Test verglich
  `View()`-Output-Dimensionen gegen `m.height`/`m.width`), sondern live in
  tmux: der Toast (top-right, komponiert direkt gegen die echten
  `m.width`/`m.height`) erschien fast komplett über den Bildschirmrand
  hinausgeschoben, nur die untere Rahmenkante sichtbar. Zweiter, verwandter
  Fund: `repoPickerWidth`s Floor (30) ignorierte das tatsächlich verfügbare
  `innerW` -- bei schmalen Terminals erzeugte das Zeilen, die breiter als
  der Rahmen waren, was lipgloss' `.Width()`-Style STILLSCHWEIGEND umbricht
  (zusätzliche, unbudgetierte Zeilen -- die eigentliche Ursache des
  Höhen-Überlaufs). Beide gefixt (`innerW`/`innerH`-Budgetierung wie in
  `viewBrowseRepo`, `repoPickerWidth` deckelt jetzt hart gegen `width`,
  Hint-Zeile wird VOR dem Zentrieren `truncate()`t statt zu umbrechen).
  Neuer Regressionstest (Pflicht-Lücke geschlossen, mirrort
  `TestChromeNeverOverflowsWidth`): `TestViewLobbyFrameMatchesWidthHeight`
  (4 Breiten/Höhen, inkl. schmal 30x24).
- **`internal/data.StartWatch`** ist ein reiner Pass-Through zu `Watch`
  (identisches Verhalten) -- die Akzeptanz-Box verlangt den Wrapper
  explizit ("EIGENER Name nur zur Trennung Produktions-Wiring vs.
  Testbarkeit"), er trägt aber keine eigene Logik. Kein Bug, nur zur
  Nachvollziehbarkeit dokumentiert.
- **`tui.Run`-Signatur erweitert um `settings config.Settings`** (Plan
  listet nur "activeProgram-Var, watchStop-Wiring" für app.go). Grund:
  design decision d's Trigger-Logik (cmd/tui.go) braucht `settings.Repos`
  VOR der Entscheidung, ob `Run` überhaupt mit `client==nil` aufgerufen
  wird -- ohne Signaturänderung hätte `config.LoadSettings()` zweimal
  unabhängig aufgerufen werden müssen (einmal in cmd/tui.go, einmal in
  Run()), mit potenziell divergentem Fehler-Fallback. Einziger Call-Site
  (`cmd/tui.go`), keine Breaking-Change-Fläche sonst.
- **Kein eigenes Golden für die Lobby.** Plan/Akzeptanz fordern keins;
  `TestViewLobbyFrameMatchesWidthHeight` deckt das Dimensions-Risiko
  stattdessen ab (kein Byte-Vergleich nötig, da die Lobby -- anders als
  Tree/Backlog/Review-Cockpit -- keine mit E1-E4 geteilte Render-Pipeline
  berührt).

## Design-Notes

**Metrik-Entscheidung (Offen/Gesamt pro Repo):** async, EIN unabhängiger
`tea.Cmd` (`repoMetricsCmd`) pro konfiguriertem Repo, gebündelt via
`tea.Batch` (`repoMetricsBatchCmd`) -- NICHT ein synchroner N-Subprozess-
Loop beim Lobby-Öffnen (das hätte die Lobby für die Dauer von N
`beans list`-Aufrufen eingefroren, "Latenz-Gift" laut Auftrag). bubbletea
führt jeden Batch-Cmd auf einer eigenen Goroutine aus -- N Repos kosten die
Wall-Clock-Zeit EINES Aufrufs, nicht N sequenzieller. Jedes Repo liefert
sein eigenes `repoMetricsMsg` (nicht ein gesammeltes Gesamt-Ergebnis), damit
ein langsames/kaputtes Repo die anderen nicht blockiert -- `repoMetrics` ist
eine `map[string]repoMetric` (COPY-ON-WRITE, I01-Konvention), fehlender
Eintrag = "…" (lädt noch), `err != nil` = "err" (nur DIESES Repo betroffen,
Rest bleibt lesbar). Bei 2-5 Repos unkritisch (Bean-eigene Einschätzung,
im Smoke-Test mit 2 Repos bestätigt: Metrik erscheint quasi sofort).

**Stop-Reihenfolge (validate-then-stop):** siehe Deviations oben -- die
zentrale, bewusste Entscheidung des Tasks. `switchRepoCmd` ruft
`client.List()` auf dem NEUEN Repo zuerst; nur bei Erfolg wird `oldStop()`
aufgerufen, danach `data.StartWatch` für das neue Repo. Ein fehlschlagender
Wechsel hinterlässt den ALTEN Watcher vollständig unangetastet -- die
laufende Session verliert nie ihre Live-Reload-Fähigkeit für einen
gescheiterten Wechselversuch. B05 (watcher.go) bleibt gewahrt: der
synchrone `oldStop()`-Aufruf läuft NICHT auf der Watcher-eigenen
onChange-Goroutine, sondern auf `switchRepoCmd`s eigener `tea.Cmd`-
Goroutine (vom bubbletea-Runtime gestartet) -- exakt der Grund, warum dieser
Code als `tea.Cmd` statt inline im Tastatur-Handler lebt (ein blockierender
`oldStop()` inline hätte die gesamte UI für die Dauer des
Watcher-Goroutine-Shutdowns eingefroren).

**Watcher-Stop bei Quit:** ersetzt Run()s alte `defer stop()`-Variable
(kannte nur den INITIALEN Watcher) durch eine Auswertung des von `p.Run()`
zurückgegebenen finalen Modells (`fm.watchStop`) -- damit wird IMMER der
zuletzt aktive Watcher gestoppt, egal ob die Session nie gewechselt hat
oder mehrfach. Der initiale Watcher wird per `initialWatchMsg` (async,
`go p.Send(...)`, gleiches B05-Muster wie `watchUnavailableMsg`) ins Modell
eingespeist, da `p` selbst erst nach `tea.NewProgram(...)` existiert (die
notify-Closure des initialen Watchs bindet weiterhin direkt an `p`, nicht
an `activeProgram` -- kein Race möglich, siehe app.go-Kommentar).

## Notes for T7 (Archiv)

- **archive/ wird beim Repo-Wechsel automatisch mitgewatcht.**
  `data.StartWatch`/`data.Watch` fügt `archive/` bereits beim initialen
  `Watch()`-Aufruf hinzu (falls zum Zeitpunkt vorhanden) bzw. dynamisch bei
  dessen erstem `Create`-Event -- dieses Verhalten ist repo-UNABHÄNGIG in
  `Watch()` selbst verankert, `switchRepoCmd` startet für JEDES neue Repo
  einen frischen `Watch()`-Aufruf mit der immer gleichen Logik. T7 muss
  hier NICHTS zusätzlich verdrahten -- der Repo-Wechsel und ein späterer
  `showArchived`-Filter sind komplett orthogonal (der Filter ist reine
  View-/Index-Logik über bereits geladene `[]data.Bean`, der Wechsel
  betrifft nur, WELCHES Repo geladen wird).
- **`applyRepoSwitched` setzt `m.filterOpen`/Facetten zurück (via
  `clearFacets()`), aber KEINEN eigenen `showArchived`-Reset** -- da dieses
  Feld heute noch nicht existiert. Wenn T7 `m.showArchived bool` (oder
  ähnlich) einführt, gehört es in dieselbe Reset-Liste in
  `applyRepoSwitched` (update.go) wie `filterOpen`/`searchQuery` -- sonst
  würde "Archiv anzeigen" für Repo A unsichtbar ins frisch geladene Repo B
  durchschlagen (gleiche Kategorie Bug wie der bereits gefundene
  Settings-Reload-Fehler, siehe Deviations).
- **Lobby-Metrik zählt NUR offene Beans aus `client.List()`** (== `.beans/`,
  ohne `archive/` -- `beans list --json --full` liest laut bestehendem
  Datenlayer-Kontrakt kein Archiv). Falls T7 eine "N archiviert"-Zusatz-
  Metrik für die Lobby will, ist das ein NEUER `repoMetricsCmd`-Zweig
  (eigener `beans`-Aufruf oder ein zusätzliches Feld in `data.Client.List`),
  kein Umbau der bestehenden Async-Batch-Architektur.
