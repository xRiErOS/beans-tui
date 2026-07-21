---
# bt-8d35
title: 'Fokus-Modell: tab navigiert in der Region, esc verlaesst sie'
status: completed
type: feature
priority: high
created_at: 2026-07-20T09:30:49Z
updated_at: 2026-07-20T10:05:12Z
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


## Summary

Umgesetzt in a789c62 (Branch experiment/jira-style-ui), vollstaendig
boxFormEnabled()-gated.

**Detail-Region (Flow #2)**
- `keyDetailFocus` (update.go): up/down scrollen jetzt in BEIDEN Geometrien
  ueber `adjustBoxFormScroll` — die Split-vs-Vollbild-Fallunterscheidung aus
  der Merge-Aufloesung ist ersatzlos weg, samt Kommentarblock.
- tab/shift+tab treiben den Feldcursor: neue Richtungen "next"/"prev" in
  `boxFormNav` (box_nav_field.go), linear ueber `boxFormFieldOrder` mit
  Wrap an beiden Enden. Der gemeinsame Tail (scroll-into-view + Ownership)
  ist unveraendert, damit die Scroll-Mitnahme mit dem Cursor mitwandert.
- `handleKey`: neuer Guard `inFocusedDetailRegion` routet tab/shift+tab zu
  `keyDetailFocus`, solange die Detail-Region den Fokus hat. shift+tab ist
  damit NICHT mehr der Exit — esc ist der einzige (keyDetailFocus Back-Case).
- tab-in aus dem Tree setzt `boxFormCursor/boxFormCursorBean` zurueck →
  Einstieg immer auf Title (nutzt boxFormEffectiveCursors abgeleiteten Reset,
  keine zweite Reset-Regel).
- `boxFormFieldOrder`/`boxFormFieldAt`/`activateBoxFormTarget` unveraendert.
  `boxFormMoveField` + die vier Grid-Richtungen bleiben stehen (kein
  Voll-Rueckbau), sind aber aktuell an keine Taste gebunden.

**Filter-Strip (Flow #3)**
- `keyFilterMenu`: enter schliesst nicht mehr, sondern wendet den cursorten
  Wert an und HAELT den Fokus (neuer, gated Case vor dem Close-Case). esc/f
  verlassen die Region, tab/shift+tab wechseln die Kategorie (bestand schon),
  space toggelt, X leert (bestand schon).

**Footer (bt-z4w7-konform, aus der aktiven Bindung abgeleitet)**
- `boxFormFieldNext`/`boxFormFieldPrev` (box_nav_field.go) sind die
  Bindungen, die der Handler matched UND die der Footer zeigt.
  `boxFormRegionLabels` (footer_context.go) tauscht sie im Zone-3-Trichter
  ein, sobald die Detail-Region fokussiert ist: "tab next field · shift+tab
  prev field" statt "focus in/out".
- `filterStripApplyHint` ("enter apply") analog fuer den Strip.

**Reichweite (alle Ansichten geprueft)**
- Browse: geaendert (siehe oben).
- Backlog: teilt keyDetailFocus + Footer-Funnel → gleiches Verhalten, per
  Test gepinnt (TestBoxFormTabWalksFieldsInTheBacklogToo).
- Vollbild-Detail: bewusst AUSGENOMMEN — renderFullscreenBody rendert
  fieldCursor -1, es gibt dort keinen sichtbaren Cursor zu bewegen. tab
  behaelt seine globale Bedeutung; up/down scrollen (wie bisher).
- Ein separates Review-Cockpit als eigene View existiert nicht mehr (in
  view_detail_bean.go aufgegangen) — nichts zusaetzlich anzupassen.
- Flag AUS: unveraendert, per Test gepinnt (TestTabStillSwapsPanesWithout
  BoxForm, TestFooterKeepsFocusLabelsWithoutBoxForm).

## Test-Output

Voller Lauf (ohne -short), nach dem finalen Stand:

    ?   github.com/xRiErOS/beans-tui        [no test files]
    ok  github.com/xRiErOS/beans-tui/cmd    0.921s
    ?   github.com/xRiErOS/beans-tui/internal/clip  [no test files]
    ok  github.com/xRiErOS/beans-tui/internal/config    0.303s
    ok  github.com/xRiErOS/beans-tui/internal/data  5.422s
    ok  github.com/xRiErOS/beans-tui/internal/theme 0.507s
    ok  github.com/xRiErOS/beans-tui/internal/tui   150.832s

Neue Tests: TestBoxFormTabWalksFieldsWithWrap, TestBoxFormShiftTabWalks
FieldsBackwardsWithWrap, TestBoxFormArrowsScrollInsteadOfMovingTheCursor,
TestBoxFormEscLeavesTheDetailRegion, TestBoxFormTabFromTreeEntersAtTheFirst
Field, TestBoxFormTabWalksFieldsInTheBacklogToo, TestTabStillSwapsPanes
WithoutBoxForm, TestBoxFormFullscreenTabDoesNotMoveTheFieldCursor,
TestFilterStripEnterAppliesAndHoldsFocus, TestFilterStripEscLeavesTheRegion,
TestFooterNamesTabAsFieldNavInsideTheDetailRegion, TestFooterKeepsFocus
LabelsWithoutBoxForm.

Angepasst: TestBoxFormCursorStaysVisibleWhileNavigating und TestBoxFormDown
ScrollsThroughATallField (gelten unveraendert weiter, nur von tab getrieben
statt von den Pfeilen). TestBoxFormArrowsWalkFieldsInOrder →
TestBoxFormGridMoveWalksTheLayout (prueft die Grid-Geometrie jetzt eine
Ebene unter dem Keymap, da die Pfeile sie nicht mehr treiben).
TestBoxFormDownReachesPanelFields entfaellt (seine "kein Wrap"-Aussage ist
durch die PO-Entscheidung ueberholt, ersetzt durch die Wrap-Tests).

**tmux-Smoke, 80 Spalten, echtes Repo sproutling, frisches Binary +
frische Session:** kein Overflow, kein Wrap-Bug. Footer wechselt beim
tab-in korrekt auf "tab next field · shift+tab prev field" und umbricht
sauber; tab zieht den Viewport zum Body nach; Pfeile scrollen ohne den
Cursor zu bewegen; esc verlaesst die Region; im Strip zeigt der Footer
"enter apply", enter wendet an und haelt den Fokus, X leert, esc verlaesst.

## Deviations

1. **History bleibt im tab-Zyklus.** Die PO-Reihenfolge endet bei
   Relations, `boxFormFieldOrder` hat als 9. Eintrag History. Da die
   Feldtabelle laut Auftrag unveraendert bleibt und keine zweite Liste
   angelegt werden darf, laeuft tab ueber alle neun Felder. History ist ein
   echtes, klick- und rendermaessig existierendes Feld — es zu ueberspringen
   haette genau die zweite Liste erzwungen.
2. **Vollbild-Detail ausgenommen** (Begruendung oben). Dort bleibt tab die
   globale Geste.
3. **Alles gated hinter BT_BOXFORM.** Der Epic-Constraint ("alles additiv +
   gated, Bestandsgolden byte-identisch") schlaegt die Formulierung
   "beruehrt jede Ansicht": ohne Flag aendert sich nichts.
4. **enter im Strip == toggeln des cursorten Werts.** Der Strip hat keine
   separat aufklappbaren Dropdowns; die Werte des aktiven Facets stehen
   bereits offen. "Anwenden" ist dort also das Setzen des cursorten Werts —
   mit gehaltenem Fokus, wie verlangt. space behaelt seine Toggle-Rolle.
5. **`boxFormMoveField` + Grid-Richtungen bleiben unbenutzt stehen** ("kein
   Voll-Rueckbau"), weiterhin testabgedeckt, aber an keine Taste gebunden.

## Golden

Keine Golden-Datei regeneriert — kein geteiltes Golden angefasst.
