---
# bt-zhwl
title: E5 T6 — Lobby V1 + Repo-Picker p
status: todo
type: task
created_at: 2026-07-15T09:04:38Z
updated_at: 2026-07-15T09:04:38Z
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
- [ ] internal/tui/types.go: viewLobby viewID (4. Case), model.repoQuery/
      repoSearch (textinput)/repoList (listState)/watchStop func() (NEU --
      hält den AKTUELL laufenden Watcher-Stop, damit ein Repo-Wechsel ihn
      aufrufen kann)
- [ ] internal/tui/view_lobby.go (NEU): homeLogoBlock (ASCII "beans"-Banner,
      EAW-neutral, Port devd view_home.go-Muster), repoPickerBody(w)
      (Suchzeile + gefilterte Settings.Repos-Liste + Offen/Gesamt-Metrik
      via einem leichten `beans list --json -s todo -s in-progress -s draft
      --quiet | wc -l`-artigen Zähl-Aufruf ODER client.List()+Index einmalig
      pro Repo -- Entscheidung + Kosten/Latenz-Abwägung bei Task-Start
      dokumentieren), viewLobby() (Chrome + zentrierter Block, Port
      centerInto/pickerRowFill VERBATIM)
- [ ] internal/data (neue Datei watch_lifecycle.go ODER Erweiterung
      watcher.go): startWatch(repoDir string, notify func()) (stop func(),
      err error) -- dünner, testbarer Wrapper um data.Watch (identische
      Signatur, EIGENER Name nur zur Trennung Produktions-Wiring vs.
      Testbarkeit)
- [ ] internal/tui/messages.go: repoSwitchedMsg{client *data.Client,
      repoDir string, beans []data.Bean, watchStop func(), err error};
      switchRepoCmd(oldStop func(), newRepoDir string, notify func())
      tea.Cmd -- ruft oldStop() (falls != nil), baut neuen data.Client,
      lädt List(), startet neuen Watch, gibt repoSwitchedMsg zurück (KEIN
      tea.Program-Bezug in dieser Funktion selbst -- notify ist injiziert,
      testbar ohne echten Bubbletea-Runtime)
- [ ] internal/tui/app.go Run(): package-level `var activeProgram
      *tea.Program` (oder Closure-Var) gesetzt direkt nach
      tea.NewProgram(...), initialer data.Watch-Aufruf liefert seinen
      stop-func jetzt an m.watchStop (Modell trägt ihn, nicht mehr NUR
      lokale Run()-Variable) -- switchRepoCmd's notify-Parameter wird mit
      `func(){ activeProgram.Send(watchMsg{}) }` gefüllt
- [ ] internal/tui/update.go: `case repoSwitchedMsg:` -- m.client/m.repoDir/
      m.watchStop/idx (via loadCmd oder direkt msg.beans) neu setzen,
      m.view = viewBrowseRepo, Cursor/Expand-State zurücksetzen (Port devd
      selectProject-Reset-Muster); handleKey: `keys.Picker` (`p`) von
      ÜBERALL (design decision h, wie ctrl+k/?) -> öffnet viewLobby (lädt
      Settings.Repos neu, falls seit Start geändert)
- [ ] internal/config/state.go Nutzung: beim Repo-Wechsel SetLastRepo(newDir)
      (Read-Modify-Write, Port devd DD2-273-Musterkommentar: andere
      State-Felder bleiben erhalten)
- [ ] cmd/tui.go RunE: Startup-Trigger-Logik (design decision d) -- Settings
      laden, `bt <pfad>`-Arg > data.FindRepo(cwd)-Erfolg > (>=2 Repos in
      Settings -> Lobby) > (1 Repo/kein Repo -> Fehlermeldung wie bisher,
      US-01 unverändert)
- [ ] overlay_palette.go paletteActions: neuer globaler Eintrag `{actionID:
      "repo_picker", label: "repo: wechseln"}`
- [ ] view_lobby_test.go: TestLobbyShowsConfiguredRepos,
      TestLobbyFilterNarrowsBySearch, TestLobbySelectSwitchesRepoAndView,
      TestPickerKeyOpensLobbyFromAnyView, TestNoLobbyOnSingleRepoCwdMatch
      (design decision d Kernfall -- E1-E4-Verhalten bleibt intakt)
- [ ] switch_repo_test.go (internal/tui oder internal/data): TestSwitchRepo
      CmdStopsOldWatcherStartsNew (fake notify func(), zwei newTestRepo(t)-
      Instanzen, Datei-Touch im NEUEN Repo -> notify gefeuert; Datei-Touch
      im ALTEN Repo NACH Switch -> notify NICHT gefeuert, alter Watcher tot)
- [ ] `command go test ./... -short` grün, gofmt/vet leer
- [ ] Commit `feat(tui): Lobby V1 + Repo-Picker p (Watcher-Lifecycle-Switch)`
