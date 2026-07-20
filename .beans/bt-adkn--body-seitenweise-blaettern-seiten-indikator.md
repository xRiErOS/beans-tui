---
# bt-adkn
title: Body seitenweise blaettern + Seiten-Indikator
status: todo
type: feature
priority: normal
created_at: 2026-07-20T09:23:37Z
updated_at: 2026-07-20T09:31:02Z
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
