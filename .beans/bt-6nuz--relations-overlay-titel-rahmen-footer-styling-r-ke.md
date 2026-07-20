---
# bt-6nuz
title: 'Relations-Overlay: Titel, Rahmen, Footer-Styling, r-Keybind'
status: completed
type: task
priority: high
created_at: 2026-07-20T09:22:58Z
updated_at: 2026-07-20T14:43:33Z
parent: bt-vy1q
---

PO-Befunde aus dem Live-Test 2026-07-20 (#6-#9). Vier Punkte, alle am selben Overlay —
zusammen erledigen.

## #6 — Keybind `r` fehlt im Footer
Der Key ist aktiv, wird aber nicht angezeigt. Vermutlich Nachwirkung von bt-fy5d
(Footer entschlacken): `r` wurde dort entfernt, weil es zur inline gebadgten Gruppe
gezaehlt wurde — hat aber KEIN Inline-Badge (das Relations-Panel hat keinen eigenen
Rahmen-Hotkey wie Status/Type/Tags). Also faelschlich mit ausgeblendet.
Pruefen: `internal/tui/footer_context.go`, die Auswahl in bt-fy5d (Commit `d81e583`).

## #7 — Overlay-Titel heisst "Blocking", muss "Relations" heissen
Konsistenz zur Detail-View-Ueberschrift. `internal/tui/box_picker_blocking.go`,
Render-Funktion `blockingPickerBox()`. Nur das ANZEIGE-Label, nicht die internen
Bezeichner umbenennen (blockItems/blockPending etc. bleiben).

## #8 — Suchfeld im Overlay ist ungerahmt
Soll als Box dargestellt werden, wie die uebrigen Felder. Primitiv vorhanden:
`dropdownBox`/`boxTopBorder`/`boxBottomBorder` (box_dropdown.go, box_detail_form.go).
Der Filter-Strip darueber nutzt sie bereits — das Suchfeld faellt optisch heraus.

## #9 — Footer-Keys im Overlay nicht konsistent gestylt
Sollen wie im Main-View unten stehen: **Key in Teal, Aktion in Subtext**. Keys, die
bereits am Feld im Rahmen stehen, NICHT zusaetzlich als Tooltip auffuehren (gleiche
Regel wie bt-fy5d im Haupt-Footer).
Styling-Quelle: `renderBindings()` in `internal/tui/view.go`. Theme-Token nur aus
`internal/theme/`.

## Kontext
Die Footer-Labels werden seit bt-z4w7 (Commit `3e3ddd9`) aus der TATSAECHLICH aktiven
Bindung abgeleitet statt hart verdrahtet — `valueMenuGroupKey`, `blockingPickerToggleHint`.
Neue Labels dort einhaengen, nicht danebenbauen.

## Akzeptanz
- [ ] `r` erscheint wieder im Browse-Footer
- [ ] Overlay-Titel lautet "Relations"
- [ ] Suchfeld gerahmt, optisch konsistent zum Filter-Strip
- [ ] Footer-Keys Teal/Subtext, keine Dopplung mit Inline-Badges
- [ ] tmux-Smoke bei 80 Spalten (Footer-Aenderung = Projektregel)
- [ ] voller Testlauf gruen


## Summary

**#6 `r` im Footer** — an EINER Stelle korrigiert, nicht als Einzel-Nachtrag.
`boxFormInlineKeys` (view_browse_repo.go) hatte `r` aufgenommen mit der Begruendung
"sein Gegenstand IST das Relations-Panel" — aber ein Panel ist kein Badge. Die
Regel ist jetzt woertlich "dieser Key wird als Inline-(x)-Badge gerendert", und
`TestBoxFormInlineKeysAllHaveAnInlineBadge` haelt sie STRUKTURELL: der Test rendert
die echte box-form-Detail-Pane und verlangt fuer jeden ausgeblendeten Key ein
literales `(x)` darin. Der naechste Fall dieser Klasse faellt damit auf, statt
still in der UI zu verschwinden.

**#7 Titel** — `modalPanel("Relations", ...)`. Nur das Anzeige-Label;
blockItems/blockPending/blockFiltered/data.SetBlocking behalten "blocking"
(Domaenenwort von Datenmodell und beans-CLI).

**#8 Suchfeld gerahmt** — neues `pickerFilter.searchBox()` ueber
boxTopBorder/boxBottomBorder, Label "Search", kein Hotkey-Badge (das Feld ist
immer fokussiert, es gaebe nichts Wahres zu bewerben). `pickerFilterChromeLines`
von filterBarHeight+2 auf +4; das Zeilenbudget zahlt es, TestPickerBoxesFitIn24Lines
beweist den 80x24-Fit.

**#9 Footer-Styling** — `pickerFilterHint` nimmt jetzt `[]keybind.Binding` und geht
durch `renderBindings` (Teal Key / Subtext Aktion). Quelle sind
`blockingPickerLocalBindings()`/`parentPickerLocalBindings()` — dieselben Accessoren,
die Footer Zone 3 rendert (bt-z4w7-Muster). Innerer Hint und aeusserer Footer sind
damit EIN einmal gebauter String und koennen nicht divergieren. Die vier Facetten-
Chords ^t/^n/^p/^g entfallen: der Chip-Strip badged sie im eigenen Rahmen,
das ist genau die Dopplung, die bt-fy5d entfernt hat. "type:filter" entfaellt,
weil das Suchfeld nun ein "Search"-Label traegt (#8).

**Vom Smoke gefunden:** der Hint brach bei 80 Spalten mitten im Wort um
("esc b" / "ack") — er stand ungewrappt im Modal-Body. Er laeuft jetzt durch
dasselbe ANSI-bewusste `wrapText`, das der Haupt-Footer nutzt (renderBindings
macht jeden Eintrag per NBSP wrap-atomar, nur " · " bricht). Zusaetzlich
picker-lokale Relabelings `pickerApplyHint`/`pickerSetHint`/`pickerDiscardHint`
(enter save/set, esc discard) statt der globalen "open/confirm"/"back" — wahrer
UND kuerzer, damit einzeilig bei 80 Spalten.

Golden: browse_boxform.golden (eine Zeile: `r Blocking` zurueck im Footer).

Commit 10fb998.

## Test-Output

command go test ./... — vollstaendig, ohne -short:

    ?   github.com/xRiErOS/beans-tui  [no test files]
    ok  github.com/xRiErOS/beans-tui/cmd  1.021s
    ?   github.com/xRiErOS/beans-tui/internal/clip  [no test files]
    ok  github.com/xRiErOS/beans-tui/internal/config  (cached)
    ok  github.com/xRiErOS/beans-tui/internal/data  (cached)
    ok  github.com/xRiErOS/beans-tui/internal/theme  (cached)
    ok  github.com/xRiErOS/beans-tui/internal/tui  151.475s

tmux-Smoke 80x24 gegen ~/Obsidian/tools/sproutling, BT_BOXFORM=1, frisches Binary
und frische Session: Titel "Relations", Suchfeld gerahmt, Hint einzeilig ohne
Wortbruch, Footer zweizeilig mit `r Blocking`. Parent-Picker gegengeprueft.

## Deviations

`pickerFilterHint` ist geteilt, die Aenderung trifft daher auch den **Parent-Picker**
("↑ up · ↓ down · enter set · esc discard"). Bewusst: die beiden Overlays
auseinanderlaufen zu lassen waere schlechter als die konsistente Behandlung, und
box_picker_filter.go existiert laut eigenem Kopfkommentar genau dafuer, dass die
zwei Picker nicht driften.


## Nachtrag 2026-07-20 (Commit `56dace0`)

Beim Merge-Smoke fiel auf, dass #7 nur halb umgesetzt war: Overlay-Titel und Detail-Panel
hiessen bereits "Relations", das **Footer-Label** zeigte weiterhin `r Blocking`.
Nachgezogen in `keymap.go` (`WithHelp("r", "Relations")`), zwei Q06-Listen-Tests und vier
Golden angepasst.

Bewusst NICHT umbenannt: der Bezeichner `Blocking`, `data.SetBlocking` und
`appendGroup("Blocking", …)` in `view_detail_bean.go` — die meinen die konkrete
Beziehungsart innerhalb des Relations-Panels, nicht das Panel selbst. Nur das
ANZEIGE-Label folgt der Ueberschrift.
