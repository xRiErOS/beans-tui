---
# bt-wpn9
title: E5 T2 — Help-Overlay ? (aus Keymap generiert)
status: todo
type: task
created_at: 2026-07-15T09:04:24Z
updated_at: 2026-07-15T09:04:24Z
parent: bt-5h4d
---

Ziel: Help-Overlay `?` (Port devd `overlay_shortcuts.go`), generiert aus
`keys.helpGroups()` -- Single Source, kein Drift zur realen Keymap. KEINE
neuen keyMap-Felder nötig (`Help` existiert bereits seit E1 Task 7 und ist
bereits im Drift-Guard-Test `TestHelpGroupsCoverEveryBindingExactlyOnce`
abgedeckt) -- reine Verdrahtung von Overlay-State + Rendering + Capture.

Plan: docs/plans/v1-port/epic-E5-plan.md »Task 2«.

## Akzeptanz
- [ ] internal/tui/overlay_shortcuts.go (NEU): helpBox() -- Port devd
      VERBATIM (modalPanel-basiert, Spaltenbreite über alle Gruppen
      global bestimmt)
- [ ] internal/tui/types.go: model.helpOpen bool (neues Feld)
- [ ] internal/tui/update.go handleKey: `m.helpOpen`-Vollcapture-Block
      (Precedent filterOpen/paletteOpen) direkt VOR dem `ctrl+c/q/tab`-
      Switch; `?` (keys.Help) öffnet von ÜBERALL (wie ctrl+k, design
      decision h) -- Case oberhalb des Review-Cockpit-Capture-Blocks
- [ ] internal/tui/keyHelp(msg) (NEU, in overlay_shortcuts.go oder
      update.go): esc/?/q schließen (Port devd Footer-Hinweis "esc/?/q:
      close" wörtlich)
- [ ] internal/tui/view_browse_repo.go composeOverlays: neuer Help-Case
      NACH m.paletteOpen, VOR m.confirmQuit (Painter's-Algorithmus-Reihen-
      folge -- Quit bleibt oberste Priorität, Port devd viewComposite-
      Reihenfolge confirmQuit > helpOpen)
- [ ] overlay_shortcuts_test.go: TestHelpBoxRendersEveryGroup,
      TestKeyHelpOpensFromAnyView, TestKeyHelpEscQCloseHelp,
      TestHelpCapturesSingleKeysWhileOpen (q darf NICHT quit-confirm
      triggern während Help offen ist)
- [ ] `command go test ./... -short` grün, gofmt/vet leer, TestKeymapNoCtrlSQ
      + TestHelpGroupsCoverEveryBindingExactlyOnce weiterhin grün (Drift-
      Guard unangetastet, da keine neuen Bindings)
- [ ] Commit `feat(tui): Help-Overlay ? aus zentraler Keymap generiert`
