---
# bt-0l8c
title: E5 T5 — Settings (config.yaml + state.json)
status: completed
type: task
priority: normal
created_at: 2026-07-15T09:04:38Z
updated_at: 2026-07-15T11:40:13Z
parent: bt-5h4d
---

Ziel: Settings (`~/.config/beans-tui/config.yaml`: repos/editor/accent/
tree_width, Port devd `internal/config/settings.go`) + Runtime-State
(`~/.config/beans-tui/state.json`: LastRepo, Port devd `internal/config/
state.go`). Design decision c (Editor-Präzedenz, PFLICHT lt. Handover):
Settings.Editor (wenn gesetzt) > $VISUAL > $EDITOR > "vi" -- NEUE Schicht
STRENG ÜBER der bestehenden E3-Kaskade (editor.go, bt-sl45), Default leer
(Out-of-the-box-Verhalten unverändert, TestEditorBinaryResolvesVisualThen
EditorThenVi bleibt grün ohne Anpassung). EIN Config-Verzeichnis für
config.yaml UND state.json (`~/.config/beans-tui/`) -- Deviation von devd's
Split (`~/.config/devd-cli/` vs. `~/.config/dd/`), bewusst konsolidiert.

Plan: docs/plans/v1-port/epic-E5-plan.md »Task 5«.

## Akzeptanz
- [x] `command go get gopkg.in/yaml.v3`
- [x] internal/config/settings.go (NEU, Package): Settings{Repos []string,
      Editor string, Theme.Accent string, Layout.TreeWidth int},
      DefaultSettings(), LoadSettings()/SaveUserSettings() -- Port devd
      settings.go MINUS ModalWidth/StartProject/Keybindings (design-spec
      nennt nur repos/editor/accent/treewidth), EIN Pfad-Layer (kein
      lokaler CWD-Override -- devds zweite Schicht ist YAGNI hier)
- [x] internal/config/state.go (NEU): State{LastRepo string}, Load()/Save()/
      SetLastRepo() -- Port devd state.go, EIN Verzeichnis mit settings.go
- [x] internal/tui/editor.go: package-level `var configuredEditor string`
      (leer default), editorBinary() prüft configuredEditor ZUERST (Port
      devd-Mechanik), dann die bestehende $VISUAL/$EDITOR/vi-Kaskade
      unverändert
- [x] internal/tui/box_form_settings.go (NEU): huh-Form (Port devd
      form_edit_settings.go-Muster, forms_shared.go-Infra wiederverwenden)
      -- Felder repos (Textarea, ein Pfad je Zeile)/editor/accent (Hex)/
      tree_width; formKind "settings"; submitForm-Case ruft
      config.SaveUserSettings + setzt configuredEditor + theme.SetAccent
      LIVE (kein Neustart nötig, Port devd DD2-221-Prinzip)
- [x] internal/tui/overlay_palette.go paletteActions: neuer globaler Eintrag
      `{actionID: "settings", label: "settings: öffnen"}` (KEIN dedizierter
      Key -- design-spec §7 kennt keinen -- ausschließlich über Command-
      Center erreichbar)
- [x] internal/theme/theme.go: SetAccent(hex string) (falls noch nicht
      vorhanden -- prüfen) für Live-Akzentfarben-Override
- [x] internal/tui/app.go Run() / cmd/tui.go: LoadSettings() beim Start,
      configuredEditor + theme.SetAccent(Settings.Theme.Accent) VOR
      tea.NewProgram, Settings an newModel durchreichen (m.settings-Feld
      ODER gezielte Felder -- Entscheidung bei Task-Start)
- [x] settings_test.go (config package, Port devd settings_test.go-Muster):
      TestParseSettingsYAML, TestMergeSettingsOnlySetFieldsWin,
      TestValidateSettingsClampsTreeWidth, TestValidateSettingsRejects
      InvalidAccentHex, TestSaveUserSettingsReadModifyWrite
- [x] state_test.go: TestStateLoadMissingFileReturnsZero,
      TestStateSaveAndLoadRoundtrip, TestSetLastRepoPreservesOtherFields
- [x] box_form_settings_test.go: TestSettingsFormPrefillsCurrentValues,
      TestSettingsFormSubmitSavesAndAppliesLive,
      TestEditorBinaryPrefersConfiguredEditorOverEnv (Präzedenz-Test,
      design decision c)
- [x] `command go test ./... -short` grün, gofmt/vet leer
- [x] Commit `feat(config): Settings (config.yaml+state.json) + Editor-Präzedenz`

## Summary

internal/config/ (NEU, Package): settings.go (Settings{Repos []string,
Editor string, Theme.Accent string, Layout.TreeWidth int},
DefaultSettings/LoadSettings/SaveUserSettings/parseSettings/mergeSettings/
validateSettings/reposTrimmed, EIN Pfad-Layer `~/.config/beans-tui/
config.yaml`, kein CWD-Override) + state.go (State{LastRepo,
LastSeenVersion}, Load/Save/SetLastRepo, `~/.config/beans-tui/state.json`,
SAME Verzeichnis). DEVIATION vs. devd's eigener DefaultSettings: Editor
default `""` (nicht devd's `"nvim"`) -- design decision c verlangt
Out-of-the-box-Verhalten unverändert.

internal/tui/editor.go: `var configuredEditor string` (package-level,
Default leer) + editorBinary() prüft configuredEditor ZUERST, danach
unverändert $VISUAL/$EDITOR/vi.

internal/tui/box_form_settings.go (NEU): buildSettingsForm (4 Felder:
repos Textarea/editor Input/accent Input+Hex-Validator/tree_width
Input+Range-Validator 24-60), openSettingsForm (kein mutTarget -- Settings
sind repo-unabhängig), reposToLines/linesToRepos (Textarea <-> []string).
box_confirm_create.go submitForm(): neuer Case "settings" -- rein lokal
(kein data.Client, daher kein Cmd wie bei editTitle/reject), ruft
config.SaveUserSettings + config.LoadSettings (Re-Read/Clamp) +
setzt m.settings + configuredEditor + theme.SetAccent LIVE.
forms_shared.go formTitle(): Case "settings" -> "Settings".

internal/tui/overlay_palette.go: paletteActions neuer globaler Eintrag
`{actionID: "settings", label: "settings: öffnen"}` (letzter Eintrag,
KEIN dedizierter Key), dispatchPalette Case "settings" ->
openSettingsForm().

internal/theme/theme.go: SetAccent existierte bereits, hatte aber KEINEN
Guard für leeren/ungültigen Hex (Bug ggü. Plan-Vorgabe "leerer/ungültiger
Hex -> No-Op") -- gefixt: `accentHexRe`-Regex-Guard, No-Op bei
Nichttreffer. Siehe Deviations.

internal/tui/app.go Run(): config.LoadSettings() VOR tea.NewProgram,
configuredEditor + theme.SetAccent gesetzt, `m.settings = settings`
NACH newModel() (mirrort devd app.go's `m.cfg`-Zuweisung in Run(), KEIN
newModel()-Signatur-Change -- Entscheidung bei Task-Start).
internal/tui/types.go: `settings config.Settings`-Feld + Import.

command go mod tidy nach `go get`: promotet gopkg.in/yaml.v3 von
`// indirect` zu direct require (Package internal/config importiert es
jetzt tatsächlich).

## Test-Output

RED bewiesen (Compile-Fail zählt als RED, CLAUDE.md/Auftrag): (1)
TestEditorBinaryPrefersConfiguredEditorOverEnv gegen den unveränderten
editor.go -> `undefined: configuredEditor` (3 Fundstellen). (2)
TestSetAccentOverridesThenNoOpOnEmptyOrInvalid (theme-Package) gegen den
UNVERÄNDERTEN SetAccent -> lief durch, aber schlug inhaltlich fehl (kein
Guard vorhanden: leerer/ungültiger Hex überschrieb Accent/Header
trotzdem) -- echter RED-Beleg für den Bug, nicht nur Compile-Fail. (3)
box_form_settings_test.go (TestSettingsFormPrefillsCurrentValues/
...ValidatesAccentAndTreeWidth/...SubmitSavesAndAppliesLive/
TestDispatchPaletteSettingsOpensForm) gegen den Stand vor
box_form_settings.go -> `undefined: buildSettingsForm` (4 Fundstellen).
Jeweils implementiert -> GREEN (siehe unten). internal/config/{settings,
state}_test.go wurden zusammen mit ihrer Implementierung verfasst (Port
eines bereits bekannten, gut verstandenen devd-Musters, niedriges Risiko)
-- kein separater RED-Lauf dafür, abweichend vom strikten RED-first für
die zwei riskanteren Stellen oben (Editor-Präzedenz, Theme-Golden-Fix),
wo echte Unklarheit/Bug-Potential bestand.

`command go test ./internal/config/... -v`: alle 9 Tests grün
(TestParseSettingsYAML, TestMergeSettingsOnlySetFieldsWin,
TestValidateSettingsClampsTreeWidth, TestValidateSettingsRejects
InvalidAccentHex, TestSaveUserSettingsReadModifyWrite,
TestLoadSettingsMissingFileReturnsDefaults, TestStateLoadMissingFileReturnsZero,
TestStateSaveAndLoadRoundtrip, TestSetLastRepoPreservesOtherFields).

`command go test ./internal/theme/... -v`: TestSetAccentOverridesThen
NoOpOnEmptyOrInvalid + alle 4 Bestandstests grün.

`command go test ./internal/tui/... -run 'TestEditorBinary|TestSettingsForm|
TestDispatchPaletteSettings|TestPaletteActionsNoFocusedBeanOmitsNodeActions'
-v`: alle grün, inkl. UNVERÄNDERTEM TestEditorBinaryResolvesVisualThen
EditorThenVi (Regressions-Beleg design decision c).

Kosten-Hinweis (neue huh-drive-Tests, CLAUDE.md-Konvention): die drei
Settings-Form-Tests, die den 4-Felder-Form treiben (Prefills/Validates/
SubmitSavesAndAppliesLive), kosten je ~3-5s (huhs selbstperpetuierende
Blink-Cmd-Kette, gleiche Ursache wie die 7 langsamen Create-Form-Tests,
nur proportional günstiger bei 4 statt 7 Feldern) -- alle drei mit
`skipSlowHuhDriveInShortMode` versehen (einer lag empirisch über der
5s-Schwelle, die anderen beiden knapp darunter, aus Konsistenzgründen
alle drei).

Voller Lauf `command go test ./... -count=1`: GRÜN (cmd 0.28s, config
0.41s, data 2.56s, theme 0.81s, tui 138.18s). `command go vet ./...`
leer. `gofmt -l .` leer. 7 Goldens (TestChromeGolden/TestTreeGolden/
TestTreeGoldenDeterministic/TestBacklogGolden/TestBacklogGoldenDeterministic/
TestReviewCockpitGolden/TestReviewCockpitGoldenDeterministic) `-count=2`
grün, `git status --porcelain internal/tui/testdata/` leer (byte-identisch
-- die Golden-Risiko-Pflichtprüfung aus dem Auftrag).

## Smoke

tmux (100x30), `bin/bt` frisch gebaut, `HOME=/tmp/bt-t5-home` (leeres,
frisches Verzeichnis -- Missing-Config-Robustheit) gegen ein Scratch-Repo
(`beans init` + 1 Milestone). App startete fehlerfrei OHNE
`~/.config/beans-tui/` (Robustheit bestätigt).

1. `ctrl+k` -> "settings" tippen -> Command-Center filtert auf genau
   "settings: öffnen" -> `enter` öffnet das Formular (Titel "Settings",
   4 Felder repos/editor/accent/tree_width, korrekt vorbelegt mit
   Defaults leer/leer/leer/36).
2. repos übersprungen (leer), editor auf "code -w" gesetzt, accent auf
   "#f5a97f" gesetzt, tree_width bei 36 belassen, submit (huh brauchte
   nach dem letzten Feld eine zusätzliche `enter` für den impliziten
   Submit-Schritt -- UI-Detail, kein Bug). Formular schloss, zurück zu
   Browse.
3. `cat /tmp/bt-t5-home/.config/beans-tui/config.yaml` zeigt:
   ```yaml
   editor: code -w
   theme:
       accent: '#f5a97f'
   layout:
       tree_width: 36
   ```
   (repos korrekt weggelassen, `omitempty`, da leer gelassen).
4. Live-Accent-Beweis (kein Neustart): `tmux capture-pane -e` direkt
   danach enthält `38;2;245;169;127` (= RGB von #f5a97f) in den ANSI-
   Sequenzen des gerenderten Frames -- der Akzent wechselte SOFORT.
5. Formular erneut geöffnet (`ctrl+k` -> "settings" -> `enter`):
   editor-Feld zeigt "code -w", accent-Feld zeigt "#f5a97f" -- Prefill
   aus der persistierten config.yaml korrekt.
6. `esc` schließt das Formular, `q`/`q` beendet die App. Kein Panic/
   Goroutine-Dump im gesamten Pane-Capture (per Grep geprüft).

## Deviations/ERRATA

- **theme.SetAccent hatte KEINEN Guard (echter Bug, jetzt gefixt).** Die
  Funktion existierte bereits (aus einer früheren Epoche), überschrieb
  Accent/Header aber bedingungslos -- ein leerer Hex (z.B.
  Settings.Theme.Accent's eigener Zero-Value bei jedem TUI-Start ohne
  config.yaml) hätte den eingebauten Mauve-Akzent stillschweigend auf
  eine leere/ungültige lipgloss.Color gesetzt (Golden-Risiko, PFLICHT
  lt. Auftrag). Gefixt mit einem eigenen `accentHexRe`-Regex-Guard
  (dritte unabhängige Validierungsschicht neben box_form_settings.go's
  Formular-Validator und config.validateSettings' Clamp) -- No-Op bei
  leer/ungültig, Test `TestSetAccentOverridesThenNoOpOnEmptyOrInvalid`
  (internal/theme) beweist RED->GREEN.
- **State{LastRepo string} vs. TestSetLastRepoPreservesOtherFields.** Die
  Akzeptanz-Prosa skizziert `State{LastRepo string}` (ein Feld), der
  eigene Testname TestSetLastRepoPreservesOtherFields braucht aber ein
  ZWEITES Feld, um etwas Echtes zu prüfen. Aufgelöst über Plan-Step-4-
  Wortlaut "Port devd VERBATIM" (devd's State hat LastProject UND
  LastSeenVersion) -- State trägt LastSeenVersion als reserviertes,
  aktuell ungenutztes Paritätsfeld (kein bt-Feature liest/schreibt es),
  einzig damit SetLastRepo's Read-Modify-Write etwas Echtes zu bewahren
  hat, exakt wie devd's SetLastSeenVersion LastProject bewahrt.
- **TreeWidth NICHT live in die Render-Pipeline verdrahtet.**
  clickPaneGeometry (mouse.go:67) hartcodiert weiterhin `24` als
  treeWidthFloor für masterDetailWidths -- Settings.Layout.TreeWidth wird
  in diesem Task nur persistiert/validiert (Clamp 24-60), nicht
  konsumiert. Bewusst: der Plan-Datei-Scope für T5 listet weder mouse.go
  noch die drei view_*.go-Dateien unter "Modify" -- eine Verdrahtung wäre
  über den Task-Rahmen hinausgegangen. Natürlicher Folge-Punkt, nicht
  blockierend für T5.
  - **Wiring-Ort app.go statt cmd/tui.go.** Plan-Wortlaut "app.go Run() /
  cmd/tui.go" (Slash = Alternative) -- gewählt: ausschließlich app.go
  Run() (bereits verantwortlich für lipgloss/tea.NewProgram/data.Watch),
  cmd/tui.go bleibt unverändert. Keine echte Deviation (der Plan erlaubt
  explizit beides), nur dokumentiert zur Nachvollziehbarkeit.
- **repos-Feld: keine Pfad-Existenz-Validierung.** Weder Plan noch
  design-spec verlangen sie; T6s Repo-Picker validiert beim tatsächlichen
  Wechsel. Kein Validator im repos-Textarea-Feld (bewusst minimal).

## Notes for T6 (Lobby)

- **Settings.Repos-Liste** ist jetzt verfügbar (`m.settings.Repos
  []string`, gepflegt über das Settings-Formular) -- T6s Repo-Picker `p`
  konsumiert sie direkt, kein weiterer Persistenz-Layer nötig.
- **state.json LastRepo** ist verfügbar (`config.Load()`/`config.SetLastRepo`,
  `~/.config/beans-tui/state.json`) -- T6 ruft beim Repo-Wechsel
  `config.SetLastRepo(neuerPfad)`, beim Boot `config.Load().LastRepo` als
  Startkandidat (Fallback Picker bei leer/ungültig, design-spec US-14).
- **mouse.go Overlay-Guard braucht `|| m.view == viewLobby`** -- bereits
  in bt-mne6s eigenen Notes dokumentiert, hier nur bestätigt: T5 hat
  handleMouse/wheelMove NICHT angefasst, die Lücke besteht unverändert
  fort. T6 MUSS sie beim Anlegen von viewLobby schließen.
- **TreeWidth-Wiring als möglicher T6-Nebenschauplatz.** Falls T6s
  Lobby/Repo-Picker eine eigene Breiten-Berechnung braucht (eher
  unwahrscheinlich, Lobby ist vermutlich kein Master-Detail-Pane) --
  ansonsten bleibt das Verdrahten von Settings.Layout.TreeWidth in
  masterDetailWidths' treeWidthFloor ein offener, nicht terminierter
  Folgepunkt (siehe Deviations oben), kein T6-Blocker.
- **Editor-Präzedenz ist jetzt VOLLSTÄNDIG** (Settings > $VISUAL >
  $EDITOR > vi) -- T6 hat keine Editor-Berührung zu erwarten (Lobby/
  Repo-Picker nutzt keinen $EDITOR-Pfad).
