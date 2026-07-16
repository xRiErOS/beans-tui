---
# bt-1vbp
title: Vollbild-Detail History-Stack — ctrl+arrow + [/]-Fallback (F01 History)
status: completed
type: task
priority: normal
created_at: 2026-07-16T06:45:59Z
updated_at: 2026-07-16T14:31:11Z
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

- [x] Relations-Sprung im Detail-Vollbild pusht das VORHERIGE Bean auf `navBack`, leert
      `navForward`
- [x] `ctrl+links`/`[` (HistoryBack): No-Op bei leerem Stack ODER außerhalb
      `fullscreenDetail`; sonst korrekter Pop/Push-Tausch
- [x] `ctrl+rechts`/`]` (HistoryForward): symmetrisch
- [x] Mehrere Sprünge → N×Back → N×Forward landet exakt beim Ausgangsbean
- [x] `[`/`]` funktionieren als Fallback UNABHÄNGIG davon, ob `ctrl+left`/`ctrl+right` im
      genutzten Terminal ankommen (tmux-Smoke-Beleg im Commit-Body, beide Pfade einzeln
      geprüft)
- [x] History-Keys sichtbar NUR im Detail-Vollbild-Footer (nicht im Listen-Vollbild, nicht
      im Split-Modus) UND vollständig in `helpGroups()` dokumentiert
- [x] `navBack`/`navForward` sind Copy-on-Write (kein In-Place-Mutieren, `cloneStringSlice`
      durchgängig verwendet)
- [x] Voller Testlauf grün, gofmt/vet leer


## Summary

History-Stack (`ctrl+left`/`ctrl+right`, Fallback `[`/`]`) für Detail-Vollbild
implementiert: `keys.HistoryBack`/`HistoryForward` (keymap.go), Push in
`activateDetailField`s fullscreenDetail-Zweig (update.go), Pop/Push-Handler in
`keyFullscreen` (view_fullscreen.go), Footer-Sichtbarkeit via neue
`fullscreenDetailLocalBindings()`/`fullscreenListLocalBindings()`
(footer_context.go) + `helpGroups()`-Eintrag. `cloneStringSlice` (types.go)
als `[]string`-Pendant zu `cloneBoolMap` für die I01-Copy-on-Write-Konvention.

Zwei vom Supervisor entschiedene Abweichungen vom ursprünglichen
design-spec.md §15-Text (in types.go's navBack/navForward-Doc-Stamp verankert):

1. **Stack-Leeren bei Vollbild-Exit** (Deviation vs. "NICHT geleert"): ALLE
   drei Vollbild-Exit-Choke-Points (keyDetailFocus b==nil-Guard,
   keyDetailFocus Back-Case section-level esc-exit, keyFullscreen
   fullscreenList esc-exit) leeren jetzt `navBack`/`navForward` — verhindert
   History-Leak zwischen unabhängigen Vollbild-Sessions.
2. **Verschwundenes History-Bean** (F01-Analogie zum b==nil-Guard-Fix):
   Back/Forward LOOPEN über tote Einträge hinweg (idx.ByID-Check je Pop)
   statt auf ihnen zu landen — ein leerer ODER vollständig toter Stack ist
   ein sauberer No-Op, kein Trap. Getestet:
   `TestHistoryBackSkipsVanishedEntry`,
   `TestHistoryBackAllEntriesVanishedIsCleanNoOp`,
   `TestHistoryForwardSkipsVanishedEntry`.

## Test-Output

RED (Compile-Fehler, referenzierte Symbole existierten noch nicht):
```
internal/tui/footer_context_test.go:231:34: undefined: fullscreenDetailLocalBindings
internal/tui/footer_context_test.go:245:34: undefined: fullscreenListLocalBindings
internal/tui/types_test.go:21:9: undefined: cloneStringSlice
internal/tui/types_test.go:43:9: undefined: cloneStringSlice
FAIL	beans-tui/internal/tui [build failed]
```

GREEN (Ziel-Testlauf nach Implementierung, alle History/Fullscreen/
CloneStringSlice-Tests, `-v`): alle PASS (inkl. `TestHistoryBackPopsAndPushesForward`
Subtests ctrl+left + `[`-Fallback einzeln, `TestHistoryForwardPopsAndPushesBack`
Subtests ctrl+right + `]`-Fallback einzeln, `TestHistoryBackForwardRoundTripReturnsToOriginalBean`,
`TestHistoryBackNoOpOutsideFullscreenDetail`/`TestHistoryForwardNoOpOutsideFullscreenDetail`
je 3 Subtests Split-Detail/Listen-Vollbild/Tree-Backlog, esc-Exit-Clear-Tests,
b==nil-Guard-Clear-Test, Vanished-Entry-Skip-Tests, `TestHelpGroupsIncludeHistoryBindings`,
`TestHistoryBindingsUnbelegtElsewhere`).

Voller Lauf: `command go test ./... -short -count=1` — 2x grün
(`beans-tui/internal/tui ok`, alle Packages ok). `command go test ./... -count=1`
(voll, kein -short): grün, `beans-tui/internal/tui 138.345s`.
`command go test ./... -race -count=1`: grün, `beans-tui/internal/tui 142.187s`.
`command gofmt -l .`: leer. `command go vet ./...`: leer.

## Golden-Gegenbeleg

`command go test ./internal/tui/... -run "TestTreeGolden|TestBacklogGolden|TestChromeGolden" -v -count=2`
— alle PASS (2x), ohne `-update`. `git diff --stat -- internal/tui/testdata/`
leer (keine Golden-Drift) — Footer-Ergänzung ist fullscreenDetail-only,
Basis-Goldens (Split-Modus) unberührt.

## Smoke

tmux 3.7b, `TERM=xterm-256color` (macOS lokale Session, kein SSH). Live gegen
`./bin/bt` im Dogfooding-Repo selbst (`bt-tui`-Alias-Ziel).

Kaskade: `v` (List-Vollbild) → `enter` auf bt-apmy → Detail-Vollbild →
Footer zeigt `[ history back · ] history fwd · esc back` sofort. 3 Relations-
Sprünge (bt-apmy → bt-tct9 → bt-gdkx → bt-tct9). `ctrl+left` **funktionierte
nativ** in dieser Session (kein xterm-keys-Problem beobachtet) — 3x
`ctrl+left` lief exakt bt-gdkx → bt-tct9 → bt-apmy (Ausgangsbean) zurück,
4. `ctrl+left` war korrekter No-Op (blieb bt-apmy). 3x `ctrl+right` lief
exakt bt-tct9 → bt-gdkx → bt-tct9 wieder vor (identischer Endzustand).
Danach frischer Sprung bt-tct9 → bt-7pk2 (verifiziert: kappt alte
Forward-History), `[`-Fallback einzeln getestet (bt-7pk2 → bt-tct9 zurück),
`]`-Fallback einzeln getestet (bt-tct9 → bt-7pk2 vor), zweites `]` No-Op.
`esc` verließ das Vollbild korrekt zum Split-Tree mit bt-7pk2 selektiert +
Ahnen expandiert (bt-apmy/bt-tct9 aufgeklappt). Erneuter Eintritt in
Detail-Vollbild auf demselben Bean + `[` bestätigte: Stack war durch den
esc-Exit geleert (No-Op, kein Rücksprung in die alte Session-History).
`?` (Help-Overlay) listet `[ history back` / `] history fwd` in der
Navigation-Gruppe direkt nach `v fullscreen`.

Beide Tastenpfade (ctrl+arrow UND `[`/`]`) damit live UNABHÄNGIG
voneinander bestätigt — PO-Implementierungshinweis erfüllt.

## Akzeptanz-Checkliste

Alle 7 Punkte ✓ (im bean-Body oben abgehakt).

## Deviations/ERRATA

D01 (Supervisor-Entscheid, nicht Planner/PO-Wortlaut): design-spec.md §15
sagt wörtlich "navBack/navForward werden beim Verlassen NICHT geleert" —
bt-13l7s "Notes for T8" zitierte denselben Satz als "T8 muss hier nichts
nachrüsten". Der Supervisor hat das für DIESEN Task explizit überschrieben:
JEDER Vollbild-Exit leert jetzt beide Stacks. design-spec.md §15 selbst
NICHT geändert (bean-Body ist die kanonische Spec-Quelle, kein Docs-Sync in
diesem Task) — Diskrepanz ist hiermit im Commit/bean dokumentiert, damit
T9 (Design-Spec-Konsistenz-Check) sie nicht als unentdeckten Drift meldet.

D02 (Supervisor-Entscheid, Gap im bean/design-spec): verschwundenes
History-Bean war weder im bean noch in design-spec.md spezifiziert. Gewählt:
Skip-Loop (nicht simpler One-Shot-Pop wie im Code-Sketch) — verhindert, dass
EIN totes Zwischenglied Back/Forward dauerhaft blockiert, obwohl gültige
Historie weiter hinten im Stack liegt (F01-Analogie zum b==nil-Guard).

I01 (kosmetisch, kein Fix nötig): view_fullscreen.go's Datei-Doc-Kommentar-
Zeile ("file's own dispatch (handleKey's checkpoint, update.go) --
renderFullscreenBody is") ist nach dem Edit >80 Spalten lang — gofmt/vet
unberührt (Kommentare werden nicht umbrochen), rein optisch.

## Notes for T9

- design-spec.md §15 "History-Stack"-Abschnitt (Zeilen ~1090-1094, "werden
  beim Verlassen NICHT geleert") ist jetzt INKONSISTENT mit dem tatsächlichen
  Code-Stand (D01 oben) — T9s Design-Spec-Konsistenz-Check MUSS diese Stelle
  aktualisieren oder als bekannte/akzeptierte Abweichung vermerken.
- Alle 3 Vollbild-Exit-Choke-Points, die navBack/navForward leeren:
  `keyDetailFocus`s b==nil-Guard, `keyDetailFocus`s Back-Case (section-level
  esc-exit), `keyFullscreen`s fullscreenList esc-exit (update.go bzw.
  view_fullscreen.go).
- `bt-1vbp` ist laut bt-tct9-Epic-Body Kind von bt-tct9 UND laut Notes selbst
  auch als Kind-Bean im RELATIONS-Baum von bt-tct9 sichtbar (im Smoke real
  angetroffen) — kein Bug, nur Beobachtung für T9s Bean-Baum-Review.
