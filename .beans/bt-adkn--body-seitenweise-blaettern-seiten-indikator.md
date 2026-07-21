---
# bt-adkn
title: Body seitenweise blaettern + Seiten-Indikator
status: completed
type: feature
priority: normal
created_at: 2026-07-20T09:23:37Z
updated_at: 2026-07-21T06:23:40Z
parent: bt-vy1q
---

PO-Befund 2026-07-20 (#5): Der Body soll **seitenweise blaetterbar** sein, mit einem
Seiten-Indikator aus hell-/dunkelgrauen Punkten unten rechts im Rahmen.

## ACHTUNG — die vorgeschlagenen Keys sind belegt
Der PO schlug `ctrl+k` / `ctrl+i` vor. Beides geht nicht:
- **`ctrl+i` IST Tab.** Im Terminal identisch (0x09) — nicht unterscheidbar. Kollidiert
  frontal mit dem gewuenschten Tab-Feldwechsel (#2, bean fuer das Fokus-Modell).
- **`ctrl+k` ist die Command-Palette.** Steht so im Header (`ctrl+k commands`).

Ausserdem verbietet `TestKeymapNoCtrlSQ` bereits `ctrl+s`/`ctrl+q` (XOFF/XON) — es gibt
also Praezedenz dafuer, Terminal-Kollisionen hart auszuschliessen.

**Ersatz muss mit dem PO abgestimmt werden.** Kandidaten: `pgup`/`pgdn` (semantisch exakt
"blaettern", kollisionsfrei), `ctrl+f`/`ctrl+b` (vi/less-Konvention), oder `[`/`]`.
Empfehlung: `pgup`/`pgdn` als Primaerbindung, weil Blaettern genau ihre Bedeutung ist.

## Seiten-Indikator
Punkte unten rechts im Body-Rahmen, gefuellt = aktuelle Seite. Rahmen-Render:
`boxBottomBorder` (box_dropdown.go/box_detail_form.go) — dort sitzt schon das Hotkey-Badge
rechts, der Indikator muss sich den Platz mit ihm teilen.
**Beachte #4:** dort wird der Body-Hotkey moeglicherweise nach OBEN verlegt, weil der
untere Rahmen bei langem Body wegscrollt. Dann gilt dasselbe fuer den Indikator — ein
Seiten-Anzeiger, den man beim Blaettern nicht sieht, ist sinnlos. **Beide zusammen denken.**
Theme-Token nur aus `internal/theme/` (hell/dunkel = vorhandene Grau-Token, keine Hex).

## Verhaeltnis zum vorhandenen Scroll
bt-ze10 hat zeilenweises Scrollen gebaut (`adjustBoxFormScroll`, geclamped, plus Mausrad);
bt-1o4g laesst die Pfeiltasten Felder ansteuern und den Scroll mitziehen. Seitenweises
Blaettern kommt HINZU und muss durch denselben Mutationspunkt laufen
(`adjustBoxFormScroll`), sonst driften die Pfade auseinander.

## Akzeptanz
- [ ] Ersatz-Keybinding mit dem PO abgestimmt (ctrl+i/ctrl+k sind ausgeschlossen)
- [ ] Blaettern laeuft ueber `adjustBoxFormScroll`, kein zweiter Scroll-Pfad
- [ ] Seiten-Indikator sichtbar, auch waehrend des Blaetterns
- [ ] voller Testlauf gruen


## PO-Entscheidung 2026-07-20: `pgup` / `pgdn`

Blaettern liegt auf `pgup`/`pgdn` — bedeutet genau das, kollidiert mit nichts.
`ctrl+i` (== Tab) und `ctrl+k` (Palette) sind ausgeschlossen; siehe bt-mx4k, wo `ctrl+k`
ganz entfaellt, und bt-8d35 zur Tab-Belegung.

## Summary (2026-07-21)

Seitenweises Body-Blaettern (pgup/pgdn) + Seiten-Indikator gebaut, beide hinter BT_BOXFORM.

**Teil a — Paging:** pgup/pgdn im `keyDetailFocus` (update.go), Delta = sichtbares Zeilen-Budget (`boxFormScrollBounds` height), durch DENSELBEN `adjustBoxFormScroll` wie up/down + Mausrad — ein Mutationspunkt, kein zweiter Scroll-Pfad (Load-bearing C erfuellt). Bindings `boxFormPageUp/Down` als standalone vars in box_nav_field.go.

**Teil b — Indikator:** `●○`-Punktreihe unten rechts im **aeusseren** renderPane-Rahmen (`overlayPaneBottomBadge`, render_shared.go), NICHT im Body-Panel-Rahmen. `boxFormPageIndex`/`boxFormPageBadge` (box_form_page.go): count=ceil(total/h), Clamp-Ceiling→letzte Seite. Grau-Token Subtext(hell)/Surface(dunkel), keine Hex. >count Seiten als Cells → numerischer `n/N`-Fallback.

## Akzeptanz
- [x] Ersatz-Keybinding PO-abgestimmt: pgup/pgdn (ctrl+i==Tab / ctrl+k==Palette ausgeschlossen)
- [x] Blaettern laeuft ueber adjustBoxFormScroll, kein zweiter Scroll-Pfad
- [x] Seiten-Indikator sichtbar, auch waehrend des Blaetterns (fixe Aussen-Chrome)
- [x] voller Testlauf gruen

## Test-Output (RED→GREEN)
RED: `boxFormScroll after one pgdn = 0, want exactly one page (18)` + `undefined: boxFormPageIndex/pageDotFilled`.
GREEN: `ok internal/tui 151.296s`, gesamter `go test ./...` exit 0. Neue Tests: TestBoxFormPageDownUpScrollsByPage, PageDownClampsAtEnd, PageInertWithoutFlag, PageIndexMapsOffsetToPage, PageIndicatorTracksPage, PageIndicatorAbsentWhenFits.

## Smoke (80 Spalten, tmux, reales Repo)
Initial `●○○…` (Seite 0), nach 2×PageDown `○○●○…` (Seite 2) — Indikator auf fixem Rahmen, beim Blaettern sichtbar, kein Wrap, Rahmen schliesst sauber.

## Deviations/ERRATA
- **Binding-Ort:** SSTD sagte 'keymap.go Single Source', doch der verifizierte Code-Praezedenzfall (Sibling boxFormFieldNext/Prev) legt box-form-REGION-lokale Bindings bewusst als standalone vars in box_nav_field.go ab — sonst schlaegt TestHelpGroupsCoverEveryBindingExactlyOnce an (reflektiert nur keyMap-Felder). Dem Code gefolgt (CLAUDE.md: bei Konflikt dem Beobachteten trauen).
- Goldens browse_boxform + value_menu_anchored regeneriert: einzige Aenderung = `●○` im Detail-Pane-Unterrahmen (stripped diff verifiziert, keine Struktur-Drift). Flag-AUS-Goldens byte-identisch.

## Notes for T(n+1) (bt-p78f #4)
Body-Hotkey liegt bereits oben (bt-oox1, panelBoxTopHotkey). Der Seiten-Indikator sitzt auf dem AEUSSEREN renderPane-Rahmen (fix), NICHT auf boxBottomBorder — bt-p78f muss ihn daher NICHT mit verschieben. Falls bt-p78f am aeusseren Unterrahmen etwas aendert: overlayPaneBottomBadge teilt sich diese Zeile.
