---
# bt-1vbp
title: Vollbild-Detail History-Stack — ctrl+arrow + [/]-Fallback (F01 History)
status: todo
type: task
created_at: 2026-07-16T06:45:59Z
updated_at: 2026-07-16T06:45:59Z
parent: bt-tct9
blocked_by:
    - bt-13l7
---

E9 Task 8 — deckt F01s History-Stack-Teil (`ctrl+links`/`ctrl+rechts`, Fallback-Tasten,
Footer/Help-Sichtbarkeit) aus bean bt-tct9. Quelle: design-spec.md §15 „F01 —
Vollbild-Navigation", Abschnitt „History-Stack". Ist-Code: internal/tui/types.go
(`navBack`/`navForward`, bereits deklariert von Task 7), internal/tui/update.go
(`activateDetailField`s `fullscreenDetail`-Zweig, von Task 7 gebaut), internal/tui/
view_fullscreen.go (`keyFullscreen`, von Task 7 gebaut), internal/tui/keymap.go,
internal/tui/footer_context.go. **blocked_by bt-tct9 Task 7 (F01 Kernmechanik):** baut
direkt auf dessen Vollbild-Zustandsmodell/Relations-Sprung-Zweig auf — ohne Task 7 gibt es
keinen Sprung-Pfad, den die History tracken könnte.

## F01 (History-Stack) — `ctrl+links`/`ctrl+rechts` Navigations-Pfad zurück/vor

PO verbatim: "ctrl+links / ctrl+rechts: Navigations-Pfad zurück/vor (History-Stack)."
Implementierungs-Hinweis (PO): "ctrl+arrow wird von manchen Terminals/tmux geschluckt —
Planner: Verfügbarkeit prüfen, ggf. Fallback-Keys dokumentieren + im Footer/Help
ausweisen."

## Scope-Entscheidung (bereits im design-spec.md getroffen, hier nur angewendet)

Der Stack trackt AUSSCHLIESSLICH Relations-Sprünge INNERHALB `fullscreenDetail` (Task 7s
`activateDetailField`-Erweiterung) — der bestehende Split-Detail-Sprung (E2-Ära, springt
zum Tree UND verlässt `detailFocus`) bleibt UNVERÄNDERT und speist die History NICHT.

## Terminal-Verfügbarkeit (geprüft, s. design-spec.md für Details)

`bubbletea` v1.3.10 dekodiert `ctrl+left`/`ctrl+right` bereits nativ (`key.go`:
`KeyCtrlLeft`/`KeyCtrlRight`, Standard-xterm-CSI-Sequenzen `\x1b[1;5D`/`\x1b[1;5C` +
`urxvt`/Alt-Varianten) — funktioniert in den meisten modernen Terminal-Emulatoren direkt.
Bekanntes Risiko: ältere Terminals/tmux ohne `xterm-keys on`/abweichendes `TERM` in
SSH-Ketten können die Sequenz verschlucken. **Fallback:** `[`/`]` (verifiziert unbelegt im
gesamten Keymap) als ZWEITE, terminal-unabhängig garantiert zustellbare Taste je Richtung.

## Architektur-Vorgabe

**1. Neue Keymap-Bindings** (keymap.go):

```go
HistoryBack:    keybind.NewBinding(keybind.WithKeys("ctrl+left", "["), keybind.WithHelp("[", "history back")),
HistoryForward: keybind.NewBinding(keybind.WithKeys("ctrl+right", "]"), keybind.WithHelp("]", "history fwd")),
```

(Feldnamen `HistoryBack`/`HistoryForward` auf dem `keyMap`-Struct, `newKeyMap()` ergänzt
beide.) `helpGroups()`s "Navigation"-Gruppe ergänzt beide Bindings (Drift-Guard-Pflicht,
`TestHelpGroupsCoverEveryBindingExactlyOnce` deckt sie automatisch ab, sobald sie dort
gelistet sind).

**2. Push (in Task 7s `activateDetailField`-`fullscreenDetail`-Zweig, JETZT ergänzt):**

```go
default:
    if f.beanID == "" { return m, nil }
    if m.fullscreen == fullscreenDetail {
        m.navBack = append(cloneStringSlice(m.navBack), m.fullscreenBeanID) // NEU
        m.navForward = nil                                                   // NEU
        m.fullscreenBeanID = f.beanID
        m.secCursor, m.accOpen, m.detailLevel, m.fieldCursor = 0, 1, 0, 0
        return m, nil
    }
    // ... Split-Modus unverändert
```

`cloneStringSlice` — NEUER kleiner Helfer (types.go, mirrort `cloneBoolMap`s I01-Copy-on-
Write-Konvention für `[]string`-Felder: `out := append([]string(nil), src...)`), da
`navBack`/`navForward` genau wie `expanded`/die Filter-Maps NIEMALS in-place mutiert werden
dürfen (Elm-Value-Receiver-Semantik, I01-Präzedenzfall, types.go).

**3. `ctrl+links`/`[` (HistoryBack) und `ctrl+rechts`/`]` (HistoryForward)** — neuer Case
in `keyFullscreen` (view_fullscreen.go, Task 7s Datei), NUR wirksam wenn `m.fullscreen ==
fullscreenDetail`:

```go
case keybind.Matches(msg, keys.HistoryBack):
    if m.fullscreen != fullscreenDetail || len(m.navBack) == 0 {
        return true, m, nil
    }
    n := len(m.navBack)
    target := m.navBack[n-1]
    m.navBack = m.navBack[:n-1]
    m.navForward = append(cloneStringSlice(m.navForward), m.fullscreenBeanID)
    m.fullscreenBeanID = target
    m.secCursor, m.accOpen, m.detailLevel, m.fieldCursor = 0, 1, 0, 0
    return true, m, nil
case keybind.Matches(msg, keys.HistoryForward):
    if m.fullscreen != fullscreenDetail || len(m.navForward) == 0 {
        return true, m, nil
    }
    n := len(m.navForward)
    target := m.navForward[n-1]
    m.navForward = m.navForward[:n-1]
    m.navBack = append(cloneStringSlice(m.navBack), m.fullscreenBeanID)
    m.fullscreenBeanID = target
    m.secCursor, m.accOpen, m.detailLevel, m.fieldCursor = 0, 1, 0, 0
    return true, m, nil
```

(Platzierung: als weitere `case`-Zweige in `keyFullscreen`s bestehendem Dispatch, NEBEN
dem `keys.Fullscreen`/`enter`-im-Listen-Vollbild-Handling aus Task 7 — dieselbe Funktion,
kein neuer Dispatch-Einstiegspunkt in `handleKey` nötig.)

**4. Sichtbarkeit (PO: "im Footer/Help ausweisen").** Neue
`fullscreenDetailLocalBindings()` (footer_context.go-Konvention, kontextsensitiver
Footer): `[]keybind.Binding{keys.HistoryBack, keys.HistoryForward, keys.Back}` — NUR
gerendert, wenn `m.fullscreen == fullscreenDetail` (analog `contextualLocalHint`s
bestehenden Prioritäts-Cases: Filter-Menü > Overlay > Suche > Palette > Help >
view-lokaler Default — NEUER Case `m.fullscreen == fullscreenDetail` fügt sich VOR dem
finalen `viewLocal`-Fallback ein, da Vollbild ein weiterer Capture-artiger Zustand ist,
strukturell analog zu den bestehenden). `fullscreenListLocalBindings()` (analog, zeigt
`keys.Back` + `keys.Enter`, KEINE History-Keys — dort wirkungslos). Beide vollständig in
`helpGroups()`s Navigation-Gruppe dokumentiert (bereits durch Schritt 1 erledigt).

## TDD-Schritte

1. Failing tests: `TestActivateDetailFieldJumpPushesHistoryInFullscreen`,
   `TestActivateDetailFieldJumpClearsForwardHistoryOnNewJump`,
   `TestHistoryBackNoOpWhenStackEmpty`, `TestHistoryBackNoOpOutsideFullscreenDetail`,
   `TestHistoryBackPopsAndPushesForward`, `TestHistoryForwardPopsAndPushesBack`,
   `TestHistoryBackForwardRoundTripReturnsToOriginalBean` (mehrere Sprünge, dann N×Back,
   dann N×Forward → identischer Endzustand wie vor dem ersten Back), `TestCloneStringSlice`
   (I01-Copy-on-Write-Verifikation, mirrort `cloneBoolMap`-Test-Stil),
   `TestFullscreenDetailLocalBindingsShowsHistoryKeys`,
   `TestFullscreenListLocalBindingsOmitsHistoryKeys`,
   `TestHelpGroupsIncludeHistoryBindings` (Drift-Guard-Erweiterung).
2. `command go test ./internal/tui/... -run "History"` → FAIL.
3. Implementieren (Reihenfolge: `keymap.go` [Bindings+helpGroups] → `types.go`
   [cloneStringSlice] → `update.go` [Push im activateDetailField-Zweig] →
   `view_fullscreen.go` [Back/Forward-Handler] → `footer_context.go`
   [fullscreenDetailLocalBindings/fullscreenListLocalBindings + contextualLocalHint-Case]).
4. Tests grün.
5. Golden-Check: reine Tastatur-/State-Logik + ein neuer Footer-Kontext-Fall — falls
   Chrome-Goldens den Vollbild-Footer NICHT abdecken (Task 7 hat ggf. keine eigene
   Vollbild-Golden-Suite angelegt), Gegenbeleg-Lauf OHNE -update MUSS grün bleiben; falls
   Task 7 eine Vollbild-Golden-Suite angelegt hat, DIESE ggf. um einen Detail-Vollbild-mit-
   sichtbarem-History-Footer-Fall ergänzen (`-update`, Vorher/Nachher im Commit-Body).
6. `command go test ./... -short` grün (2x), voller Lauf grün, `-race` grün, gofmt/vet leer.
7. **Terminal-Verfügbarkeits-Smoke PFLICHT** (PO-Implementierungshinweis, nicht optional):
   real in tmux gegen `./bin/bt` — `ctrl+left`/`ctrl+right` UND `[`/`]` je EINZELN testen
   (mehrere Relations-Sprünge im Detail-Vollbild, dann History-Back/-Forward, Endzustand
   verifizieren). Falls `ctrl+left`/`ctrl+right` in der genutzten tmux-Session NICHT
   ankommen (bekanntes Risiko): explizit im Commit-Body dokumentieren (welche
   tmux-Version/`TERM`/`xterm-keys`-Einstellung), UND bestätigen dass `[`/`]` als Fallback
   funktioniert — das ist der eigentliche Beleg, den der PO-Hinweis verlangt.
8. Commit `feat(tui): Vollbild-Detail History-Stack — ctrl+arrow + [/]-Fallback (F01
   History)`. Footer `Refs: bt-tct9`.

## Akzeptanz-Checkliste

- [ ] Relations-Sprung im Detail-Vollbild pusht das VORHERIGE Bean auf `navBack`, leert
      `navForward`
- [ ] `ctrl+links`/`[` (HistoryBack): No-Op bei leerem Stack ODER außerhalb
      `fullscreenDetail`; sonst korrekter Pop/Push-Tausch
- [ ] `ctrl+rechts`/`]` (HistoryForward): symmetrisch
- [ ] Mehrere Sprünge → N×Back → N×Forward landet exakt beim Ausgangsbean
- [ ] `[`/`]` funktionieren als Fallback UNABHÄNGIG davon, ob `ctrl+left`/`ctrl+right` im
      genutzten Terminal ankommen (tmux-Smoke-Beleg im Commit-Body, beide Pfade einzeln
      geprüft)
- [ ] History-Keys sichtbar NUR im Detail-Vollbild-Footer (nicht im Listen-Vollbild, nicht
      im Split-Modus) UND vollständig in `helpGroups()` dokumentiert
- [ ] `navBack`/`navForward` sind Copy-on-Write (kein In-Place-Mutieren, `cloneStringSlice`
      durchgängig verwendet)
- [ ] Voller Testlauf grün, gofmt/vet leer
