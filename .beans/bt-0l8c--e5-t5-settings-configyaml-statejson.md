---
# bt-0l8c
title: E5 T5 — Settings (config.yaml + state.json)
status: todo
type: task
created_at: 2026-07-15T09:04:38Z
updated_at: 2026-07-15T09:04:38Z
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
- [ ] `command go get gopkg.in/yaml.v3`
- [ ] internal/config/settings.go (NEU, Package): Settings{Repos []string,
      Editor string, Theme.Accent string, Layout.TreeWidth int},
      DefaultSettings(), LoadSettings()/SaveUserSettings() -- Port devd
      settings.go MINUS ModalWidth/StartProject/Keybindings (design-spec
      nennt nur repos/editor/accent/treewidth), EIN Pfad-Layer (kein
      lokaler CWD-Override -- devds zweite Schicht ist YAGNI hier)
- [ ] internal/config/state.go (NEU): State{LastRepo string}, Load()/Save()/
      SetLastRepo() -- Port devd state.go, EIN Verzeichnis mit settings.go
- [ ] internal/tui/editor.go: package-level `var configuredEditor string`
      (leer default), editorBinary() prüft configuredEditor ZUERST (Port
      devd-Mechanik), dann die bestehende $VISUAL/$EDITOR/vi-Kaskade
      unverändert
- [ ] internal/tui/box_form_settings.go (NEU): huh-Form (Port devd
      form_edit_settings.go-Muster, forms_shared.go-Infra wiederverwenden)
      -- Felder repos (Textarea, ein Pfad je Zeile)/editor/accent (Hex)/
      tree_width; formKind "settings"; submitForm-Case ruft
      config.SaveUserSettings + setzt configuredEditor + theme.SetAccent
      LIVE (kein Neustart nötig, Port devd DD2-221-Prinzip)
- [ ] internal/tui/overlay_palette.go paletteActions: neuer globaler Eintrag
      `{actionID: "settings", label: "settings: öffnen"}` (KEIN dedizierter
      Key -- design-spec §7 kennt keinen -- ausschließlich über Command-
      Center erreichbar)
- [ ] internal/theme/theme.go: SetAccent(hex string) (falls noch nicht
      vorhanden -- prüfen) für Live-Akzentfarben-Override
- [ ] internal/tui/app.go Run() / cmd/tui.go: LoadSettings() beim Start,
      configuredEditor + theme.SetAccent(Settings.Theme.Accent) VOR
      tea.NewProgram, Settings an newModel durchreichen (m.settings-Feld
      ODER gezielte Felder -- Entscheidung bei Task-Start)
- [ ] settings_test.go (config package, Port devd settings_test.go-Muster):
      TestParseSettingsYAML, TestMergeSettingsOnlySetFieldsWin,
      TestValidateSettingsClampsTreeWidth, TestValidateSettingsRejects
      InvalidAccentHex, TestSaveUserSettingsReadModifyWrite
- [ ] state_test.go: TestStateLoadMissingFileReturnsZero,
      TestStateSaveAndLoadRoundtrip, TestSetLastRepoPreservesOtherFields
- [ ] box_form_settings_test.go: TestSettingsFormPrefillsCurrentValues,
      TestSettingsFormSubmitSavesAndAppliesLive,
      TestEditorBinaryPrefersConfiguredEditorOverEnv (Präzedenz-Test,
      design decision c)
- [ ] `command go test ./... -short` grün, gofmt/vet leer
- [ ] Commit `feat(config): Settings (config.yaml+state.json) + Editor-Präzedenz`
