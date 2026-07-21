---
# bt-vpvu
title: 'Maus im Tree: Klick waehlt, Doppelklick klappt auf/zu'
status: completed
type: bug
priority: high
created_at: 2026-07-20T09:22:58Z
updated_at: 2026-07-20T10:17:13Z
parent: bt-vy1q
---

PO-Befund 2026-07-20 (#14): **Im Tree lassen sich beans mit der Maus nicht auswaehlen.**

Gewuenscht:
1. **Einfacher Klick** — bean auswaehlen, erscheint im Detail
2. **Doppelklick** — bean aufklappen; erneuter Doppelklick — einklappen

## Stand
`treeClickRow` und `clickPaneGeometry` existieren (mouse.go) und wurden in S6 um den
Filter-Bar-Offset korrigiert. Warum die Auswahl im Live-Test trotzdem nicht funktioniert,
ist **ungeklaert — zuerst reproduzieren und die Ursache finden**, nicht blind umbauen.
Moegliche Spuren: greift der Klick nur im Box-Modus? Verschiebt der neue Header (#12) oder
der Umbruch (#13) die Zonen? Wird das Maus-Event ueberhaupt zugestellt (`handleMouse`
no-oped z.B. im Fullscreen komplett)?

## Fallstrick
Doppelklick-Erkennung braucht Zeitmessung. bubbletea liefert keine Doppelklick-Events —
Zeitstempel im Model halten. **Achtung:** in Golden-/Update-Tests darf keine Wanduhr die
Ausgabe bestimmen, sonst werden Tests flaky. Zeitquelle injizierbar halten.

## Abhaengigkeit
Sinnvoll NACH dem Tree-Umbruch-Paket umzusetzen — #13 verschiebt die Trefferzonen ohnehin.

## Akzeptanz
- [ ] Ursache des aktuellen Nicht-Funktionierens dokumentiert
- [ ] Einfachklick waehlt aus
- [ ] Doppelklick klappt auf/zu
- [ ] Tests ueber echte `tea.MouseMsg`-Roundtrips, keine Wanduhr-Abhaengigkeit
- [ ] voller Testlauf gruen


## Summary

Commit `6e14c11` (Ursachen-Fix strukturell in `a3fdb9e`/bt-f68z, weil derselbe
Hit-Test dort ohnehin neu geschrieben wurde).

### Ursache — zuerst reproduziert, dann gefixt

Live reproduziert in tmux 80x30 gegen sproutling, `BT_BOXFORM=1`, mit echten
SGR-Maus-Sequenzen. Befund: **der Klick funktioniert oben in der Liste, und
hoert auf zu funktionieren, sobald das Fenster gescrollt ist.**

Ursache: `viewBrowseRepo` zieht die Hoehe der Filter-Leiste von `bodyH` ab,
BEVOR es die Zeilen fenstert:

```go
bodyH -= lipgloss.Height(filterBarRow)   // Renderer
```

`treeClickRow` korrigierte in S6 nur `originY += filterBarHeight`, **nicht**
`bodyH`. Renderer und Hit-Test fensterten damit gegen unterschiedliche Hoehen
→ `windowStart` liefert zwei verschiedene Startindizes → der Klick landet auf
einem anderen bean, als der PO sieht (oder auf keinem, wenn der Index hinter
`visible` faellt).

Warum es "manchmal ging": bei ungescrollter Liste klemmen **beide**
`windowStart`-Aufrufe auf 0. Die Differenz existiert erst, wenn genug beans
da sind, dass gescrollt wird — also genau im echten Repo und nicht in den
Tests, die alle mit 4-bean-Fixtures oben in der Liste klicken.

`flatClickRow` hatte denselben Fehler an derselben Stelle.

**Gegenprobe per Mutation** (Lehre aus dem `footH`-Test in STATE.md): das
`bodyH -= filterBarHeight` wurde testweise wieder entfernt —
`TestScrolledTreeClickHitsEveryVisibleRow` wird rot und meldet konsistent
Off-by-one ueber alle Zeilen (`click on "Row 040" selected "gld-r039"`).
Der Test beisst also wirklich.

### Klick-Semantik

Auf PO-Wunsch geaendert (`mouseTreeClick`):

- **Einfacher Klick** — waehlt nur aus, erscheint im Detail. Kein Auf-/Zuklappen.
- **Doppelklick** — klappt auf; erneuter Doppelklick klappt zu (`!n.open`).

Das ersetzt die portierte devd-D03-Regel "Einfachklick klappt einen
geschlossenen Knoten auf". Begruendung: dadurch strukturierte sich der Baum bei
jedem blossen Hinsehen unter dem Zeiger um — das war ein Teil des Eindrucks,
die Pane sei nicht bedienbar.

**Klick-Paar wird konsumiert**: nach einem ausgeloesten Doppelklick wird
`lastClickAt` auf den Nullwert zurueckgesetzt. Sonst wuerde ein dritter
schneller Klick mit dem zweiten paaren und das Aufklappen sofort rueckgaengig
machen (Dreifachklick = "nichts passiert").

### Zeitquelle

`m.clock` (types.go) war bereits injizierbar und wird von `m.now()` gelesen —
kein Umbau noetig. Alle neuen Tests fahren eine handgesteuerte Uhr
(`fixedClock`), keine Wanduhr, keine Flakiness.

## Test-Output

Voller Lauf, ohne `-short`, im Vordergrund:

```
?   	github.com/xRiErOS/beans-tui	[no test files]
ok  	github.com/xRiErOS/beans-tui/cmd	1.128s
?   	github.com/xRiErOS/beans-tui/internal/clip	[no test files]
ok  	github.com/xRiErOS/beans-tui/internal/config	0.407s
ok  	github.com/xRiErOS/beans-tui/internal/data	4.163s
ok  	github.com/xRiErOS/beans-tui/internal/theme	1.323s
ok  	github.com/xRiErOS/beans-tui/internal/tui	151.864s
```

Neue Tests: `mouse_tree_select_test.go` (9), alle ueber echte
`tea.MouseMsg`-Roundtrips auf Koordinaten, die im tatsaechlich gerenderten
`View()` gesucht werden (`treeClickAt`) — nie aus der Klick-Formel selbst
abgeleitet, das waere zirkulaer.

**Live verifiziert** (tmux 80x30, sproutling, `BT_BOXFORM=1`, frisches Binary):
Klick in der gescrollten Liste trifft das richtige bean; Einfachklick auf 4x21
waehlt aus und laesst `▸` stehen; Doppelklick klappt auf (`▾`, Kind 49c4
erscheint); zweiter Doppelklick klappt wieder zu.

## Deviations

- **Verhaltensaenderung gegen einen niedergeschriebenen Beschluss:**
  devd-D03 "Einfachklick klappt auf" ist entfallen. Der bean fordert es
  (Punkt 1+2), aber die alte Regel stand als bewusste Port-Entscheidung in
  `mouse.go` und in `mouse_test.go`. Beide Stellen sind jetzt entsprechend
  umkommentiert. **Dem PO hiermit vorgelegt.**
- Die bestehenden Tests `TestDoubleClickTogglesExpand` und
  `TestSingleClickOnOpenNodeDoesNotCollapse` bleiben gueltig und unveraendert —
  sie decken die Haelften ab, die die Aenderung ueberlebt haben.
- Der eigentliche Ursachen-Fix liegt in Commit `a3fdb9e` (bt-f68z), nicht in
  `6e14c11`: der Hit-Test wurde dort ohnehin auf das Zeilen-Fenster
  umgestellt, und ihn zweimal anzufassen haette nur einen kaputten
  Zwischenstand erzeugt.
