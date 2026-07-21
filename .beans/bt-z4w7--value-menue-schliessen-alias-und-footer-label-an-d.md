---
# bt-z4w7
title: 'Value-Menue: Schliessen-Alias und Footer-Label an die Gruppe binden (B7)'
status: completed
type: bug
priority: low
created_at: 2026-07-20T07:26:50Z
updated_at: 2026-07-20T09:14:56Z
parent: bt-vy1q
---

**B7, gefunden in S4 (2026-07-20).** Das Value-Menue kann fuer die Gruppen `status`/`type`/`priority` geoeffnet werden (`s`/`o`/`u`). Aber:
- `keyValueMenu`s Schliessen-Zweig (`internal/tui/box_menu_value.go` ~Z.165) matcht nur `keys.Back` und `keys.Status` → ein per `o`/`u` geoeffnetes Menue schliesst auch mit `s`
- Der Footer-Hint (`box_menu_value.go` ~Z.243, `"esc/s:cancel"`) zeigt immer `s`, egal welche Gruppe offen ist

`esc` funktioniert in allen Faellen ⇒ nichts ist kaputt, nur Label-/Verhaltens-Mismatch.

## Hinweis
Die urspruengliche Design-Entscheidung a3 formulierte woertlich "esc/s schliesst" — eine Aenderung sollte das bewusst revidieren (als neue Sektion dokumentieren, nicht stillschweigend).

## Akzeptanz
- [ ] Schliessen-Match akzeptiert die Taste der jeweils offenen Gruppe (s/o/u) plus esc
- [ ] Footer zeigt die Taste der offenen Gruppe
- [ ] Entscheidung a3 als revidiert dokumentiert
- [ ] Tests fuer alle drei Gruppen, voller `command go test ./...` gruen


## Prelude aus dem Merge-Smoke (2026-07-20)

Gleiche Fehlerklasse, jetzt auch im **Blocking-Picker**: dessen Footer zeigt
`space/x Toggle facet`, waehrend der Blocking-Toggle in bt-a3a8 bewusst auf **space-only**
verengt wurde (`x` haette im neuen Suchfeld den Buchstaben untippbar gemacht — Design-
Entscheidung D6 dort). Der `x`-Teil des Labels gilt nur noch fuer das Facetten-Menue,
das kein Eingabefeld hat.

Beleg: tmux 80x30 gegen sproutling, Picker offen mit Suchtext.

Zusammen mit dem bereits notierten Value-Menue-Mismatch erledigen — beides ist derselbe
Befund: **das Footer-Label ist hart verdrahtet statt aus der real aktiven Bindung
abgeleitet.** Der Fix sollte an der Ursache ansetzen, nicht zwei Labels einzeln korrigieren.

## Summary

An der Ursache gefixt, nicht an den zwei Labels: der Fehler war, dass Label
und Handler **zwei getrennte Quellen** hatten (beide per Hand nebeneinander
gepflegt). Jeder kontextabhaengige Footer-Key kommt jetzt aus EINER
Accessor-Funktion, die der Handler ebenfalls matcht — Divergenz ist damit
konstruktiv ausgeschlossen.

- `valueMenuGroupKey(group)` (footer_context.go) — gelesen von allen DREI
  Flaechen: Handler-Match (`keyValueMenu`), Footer Zone 3
  (`valueMenuLocalBindings(group)`), Inline-Hint (`valueMenuBox`).
  `m.valueMenuGroup()` liefert den Kontext aus `menuItems[0].group`.
- `blockingPickerToggleHint` (space-only, "Toggle blocking") — von
  `keyBlockingPicker` gematcht UND von `blockingPickerLocalBindings`
  gerendert. Loest das geliehene `keys.TagToggle` ab, das key-technisch
  passte, aber die falsche Domaene beschrieb.
- `overlayLocalBindings` ist jetzt eine model-Methode (der Value-Menue-Satz
  haengt von der offenen Gruppe ab, die nur das model kennt).

## Dritter Fall — vom Guard gefunden

`TestPickerFooterKeysAreReservedNotTyped` (der Klassen-Guard) hat sofort
eine dritte Instanz derselben Ursache aufgedeckt, die im bean nicht stand:
**alle drei Such-Picker** bewarben `↑/i up` / `↓/k down`, obwohl die
Handler bewusst auf rohe `tea.KeyUp/KeyDown` schalten, damit `i`/`k` im
Suchfeld tippbar bleiben — der Footer schickte den PO also ins Suchfeld.
Behoben via `pickerNavUpHint`/`pickerNavDownHint`. Genau dafuer war der
Guard gedacht.

## Entscheidung a3 revidiert

`docs/plans/jira-style-experiment/design-spec.md` §5.1 (neu): a3s woertliches
"esc/`s` schliesst" war korrekt, solange `s` der einzige Oeffner war; S4 hat
mit `o`/`u` zwei weitere ergaenzt, ohne den Schliessen-Zweig mitzuziehen.
Neu gilt: **esc + der Key, der geoeffnet hat**. Bewusst als Revision
dokumentiert, nicht stillschweigend gepatcht.

## Test-Output

Voller Lauf `command go test ./...` gruen:

    ok  github.com/xRiErOS/beans-tui/cmd            0.409s
    ok  github.com/xRiErOS/beans-tui/internal/config (cached)
    ok  github.com/xRiErOS/beans-tui/internal/data   (cached)
    ok  github.com/xRiErOS/beans-tui/internal/theme  (cached)
    ok  github.com/xRiErOS/beans-tui/internal/tui  150.257s

Pflicht-Smoke tmux 80x30 gegen sproutling:
- `u` → Footer `u Priority`, Inline `esc/u:cancel`
- `o` → Footer `o Type`
- `s` → Footer `s Status`, Inline `esc/s:cancel`
- u-geoeffnetes Menue schliesst NICHT auf `s`, aber auf `u` und auf `esc`
- Blocking-Picker-Footer beginnt mit `↑ up` (kein `/i` mehr)
- kein Overflow, kein Wrap-Bug bei 80 Spalten

Commit 3e3ddd9.

## Deviations

Ein bestehender Test (`TestContextualLocalHintOverlayBlockingPickerShowsToggle`)
war selbst Teil des Bugs — er verriegelte `space/x Toggle facet`. Umgeschrieben
auf die korrekte Erwartung, mit Begruendung im Test-Kommentar.

Keine Golden-Datei beruehrt.
