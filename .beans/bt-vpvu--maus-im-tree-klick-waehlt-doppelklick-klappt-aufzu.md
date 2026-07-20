---
# bt-vpvu
title: 'Maus im Tree: Klick waehlt, Doppelklick klappt auf/zu'
status: todo
type: bug
priority: high
created_at: 2026-07-20T09:22:58Z
updated_at: 2026-07-20T09:22:58Z
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
