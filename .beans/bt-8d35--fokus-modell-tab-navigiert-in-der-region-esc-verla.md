---
# bt-8d35
title: 'Fokus-Modell: tab navigiert in der Region, esc verlaesst sie'
status: todo
type: feature
priority: high
created_at: 2026-07-20T09:30:49Z
updated_at: 2026-07-20T09:47:31Z
parent: bt-vy1q
---

PO-Befunde 2026-07-20 (#2 + #3). Zwei Wuensche, die zusammen EIN Muster ergeben — deshalb
ein bean. Der PO hat ausdruecklich verlangt, den Flow vorher durchzudenken und konkret zu
definieren; das Ergebnis steht unten und ist vor der Umsetzung freizugeben.

## Das gemeinsame Muster
> **tab/shift-tab bewegen INNERHALB der fokussierten Region. esc VERLAESST die Region.**

Regionen: Tree (links) · Detail (rechts) · Filter-Strip (oben).
Damit hoert tab auf, zwischen Panes zu springen — es wird zur Navigation im Fokus.

## Flow #2 — Detail-View
Betreten: `tab` aus dem Tree -> Fokus auf das erste Feld (Title).
Bewegen: `tab` weiter, `shift+tab` zurueck, in dieser Reihenfolge:
`Title > Status > Type > Priority > Parent > Tags > Body > Relations`.
Am Ende umbrechen (wrap), NICHT in den Tree zurueckfallen — sonst verliert man den Fokus
versehentlich beim Durchsteppen.
Oeffnen: `enter` auf dem fokussierten Feld (bereits gebaut, bt-1o4g).
Verlassen: `esc` -> Tree fokussiert (heute schon so).

## Flow #3 — Filter-Strip
Betreten: `f` -> Fokus auf das erste Chip.
Bewegen: `tab` / `shift+tab` durch Type/Status/Priority/Tags.
Oeffnen: Dropdown aufklappen.
Waehlen: Pfeiltasten durch die Werte.
Umschalten: einzelnen Wert an/aus (Mehrfachauswahl).
Anwenden: `enter` -> Filter wirkt, **Fokus bleibt im Strip** (ausdruecklicher PO-Wunsch).
Zuruecksetzen: Filter leeren.
Verlassen: `esc` -> Tree fokussiert.

## OFFENE PUNKTE — vor der Umsetzung klaeren
1. **Oeffnen: `space` oder `enter`?** Der PO schlug fuer den Strip `space` zum Aufklappen und
   `enter` zum Anwenden vor. Im Detail oeffnet aber bereits `enter` (bt-1o4g).
   Zwei Bedeutungen fuer dieselbe Geste in zwei Regionen = Inkonsistenz.
   **Vorschlag:** `enter` oeffnet ueberall; `space` schaltet einen Wert in einer offenen
   Mehrfachauswahl um (so macht es der Tag-Picker heute schon). Dann bleibt `enter` zum
   Anwenden frei und die Regionen verhalten sich gleich.
2. **Zuruecksetzen: `ctrl+x` oder `X`?** Der PO schlug `ctrl+x` vor; es GIBT bereits
   `FilterClear` auf `X` (D07: gross = View/Global).
   **Vorschlag:** `X` wiederverwenden — es sei denn, ein Strip-Feld nimmt Texteingaben
   entgegen (Tags?). Dann frisst das Eingabefeld Buchstaben und es MUSS ein Ctrl-Akkord
   sein — genau die Falle, die in bt-a3a8 zu den Ctrl-Akkorden im Picker gefuehrt hat.
   Also: haengt davon ab, ob der Strip tippbare Felder bekommt. Zuerst entscheiden.
3. **Was machen die Pfeiltasten im Detail dann noch?** Heute bewegen sie den Feld-Cursor
   (bt-1o4g, reveal-then-move mit Scroll-Mitnahme). Uebernimmt `tab` die Feldnavigation,
   waeren die Pfeile frei fuer reines Viewport-Scrollen — was #4/#5 (langer Body) entgegen
   kommt und das Modell vereinfacht.
   **Aber:** das nimmt einen frisch gebauten, getesteten Mechanismus wieder zurueck.
   Bewusst entscheiden, nicht nebenbei.

## Fallstricke
- `tab` ist heute der Pane-Wechsel. Diese Aenderung beruehrt **jede** Ansicht, nicht nur
  Browse — Backlog, Review-Cockpit, Fullscreen. Ueberall pruefen.
- `tab` == `ctrl+i` im Terminal. Es ist NICHT moeglich, beide zu unterscheiden. Jede
  kuenftige `ctrl+i`-Idee ist damit tot (siehe bean zum Body-Blaettern).
- Der Drift-Guard `TestHelpGroupsCoverEveryBindingExactlyOnce` verlangt, dass jede Bindung
  in genau einer helpGroups()-Gruppe liegt.
- Footer-Labels werden seit bt-z4w7 aus der aktiven Bindung ABGELEITET — neue
  kontextabhaengige Labels dort einhaengen, nicht danebenbauen.

## Akzeptanz
- [ ] Die drei offenen Punkte oben sind vom PO entschieden
- [ ] tab/shift+tab wandern im Detail durch die Felder, mit Umbruch
- [ ] `f` fokussiert den Strip, tab wandert darin, esc verlaesst ihn
- [ ] `enter` im Strip wendet an und HAELT den Fokus
- [ ] Verhalten in Backlog/Review/Fullscreen geprueft, nicht nur in Browse
- [ ] tmux-Smoke bei 80 Spalten
- [ ] voller Testlauf gruen


## PO-Entscheidungen 2026-07-20 — die drei offenen Punkte sind geklaert

**1. Oeffnen: `enter`, ueberall.** Konsistenz schlaegt Bequemlichkeit. `space` oeffnet NICHT.
Im offenen Mehrfachauswahl-Menue schaltet `space` einen Wert um (wie im Tag-Picker heute).

**2. Zuruecksetzen: `X`.** Kein `ctrl+x`. Die vorhandene `FilterClear`-Bindung (D07,
gross = View/Global) wird wiederverwendet.
**Folge-Constraint:** damit darf der Filter-Strip **keine tippbaren Felder** bekommen —
ein Texteingabefeld wuerde das `X` verschlucken und den Ctrl-Akkord erzwingen. Wenn spaeter
eine Titel-/Freitextsuche in den Strip soll, ist diese Entscheidung neu zu bewerten.

**3. Pfeiltasten im Detail scrollen die GESAMTE Ansicht.** Sie bewegen NICHT mehr den
Feld-Cursor — das uebernimmt `tab`/`shift+tab`.

### Was das fuer bt-1o4g bedeutet (wichtig, kein Voll-Rueckbau)
bt-1o4g (Commit `75d791a`) hat zwei Dinge gebaut:
- die Pfeiltasten-Bindung auf `boxFormNav` -> **wird ersetzt**, Pfeile scrollen kuenftig
  wieder ueber `adjustBoxFormScroll` (das ist exakt der Zustand von bt-ze10)
- die **Scroll-Mitnahme** ("reveal-then-move", fokussiertes Feld darf nie ausserhalb des
  Viewports liegen) -> **BLEIBT und wird an `tab` gehaengt**. Springt `tab` auf ein Feld
  ausserhalb des sichtbaren Bereichs, muss der Viewport nachziehen. Ohne das waere die
  Tab-Navigation blind.
- `boxFormFieldOrder` / `boxFormFieldAt` / `activateBoxFormTarget` -> **bleiben unveraendert**.
  Die Feld-Reihenfolge, die `tab` abwandert, ist dieselbe Tabelle, die Maus und Render
  benutzen. Keine zweite Liste anlegen.

Der Test `TestBoxFormCursorStaysVisibleWhileNavigating` gilt weiter — nur getrieben von
`tab` statt von den Pfeilen. `TestBoxFormDownScrollsThroughATallField` (exakt-plusminus-1
am Pfeil) gilt ebenfalls weiter, denn die Pfeile scrollen ja wieder.

## Status
Draft aufgehoben — umsetzbar.
