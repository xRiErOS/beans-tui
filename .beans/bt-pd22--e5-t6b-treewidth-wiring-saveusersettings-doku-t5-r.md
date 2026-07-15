---
# bt-pd22
title: E5 T6b — TreeWidth-Wiring + SaveUserSettings-Doku (T5-Review-Findings)
status: in-progress
type: task
priority: normal
created_at: 2026-07-15T11:49:11Z
updated_at: 2026-07-15T12:34:44Z
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

[ ] clickPaneGeometry nutzt Settings-TreeWidth (Fallback 24), Live-Apply im settings-Submit
[ ] 2 neue Tests grün, 7 Goldens -count=2 byte-identisch
[ ] SaveUserSettings-Doc-Kommentar präzisiert
[ ] command go test ./... voll grün, gofmt/vet leer
[ ] Commit fix(tui): TreeWidth-Setting wirkt im Render (T5-Review I01/I02)


## Prelude aus T6-Review (I01, low — Reviewer 2026-07-15)

applyRepoSwitched (update.go) resettet m.toast nicht — Toast aus altem Repo bleibt nach Repo-Wechsel kurz sichtbar. Fix hier miterledigen: m.toast=nil (bzw. Muster der übrigen Reset-Felder) in applyRepoSwitched + Regressionstest TestRepoSwitchClearsToast.
