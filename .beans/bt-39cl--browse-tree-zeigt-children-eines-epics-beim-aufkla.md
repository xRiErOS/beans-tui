---
# bt-39cl
title: Browse-Tree zeigt Children eines Epics beim Aufklappen nicht an
status: todo
type: bug
priority: high
created_at: 2026-07-16T20:20:40Z
updated_at: 2026-07-16T20:47:11Z
parent: bt-tct9
---

PO-Nebenbefund, US-Review Runde 3 (2026-07-16): PO hatte ein Bean (Kurzform vom
PO 'a2ca' genannt, Typ Epic, mit Children) im Browse-Tree links selektiert und
wollte aufklappen -- Children erscheinen NICHT.

STATUS: exakte betroffene Bean-ID noch nicht verifiziert (Repo-Suche nach 'a2ca'
ergab keinen Treffer -- evtl. PO-Kurzform/Sichtstand weicht ab). ERSTER SCHRITT
der Nacharbeit: mit PO das betroffene Bean identifizieren (Live-Screenshot oder
exakte ID), danach reproduzieren.

Prioritaet high, weil es die Grundfunktion Tree-Expand betreffen koennte (nicht
nur Kosmetik) -- falls reproduzierbar, Reichweite pruefen (nur dieses Epic oder
generell?).

Quelle: bt-tct9 US-Review Runde 3.



## Planner-Konkretisierung (2026-07-16)

**T1 (PFLICHT, KEIN Fix ohne diesen Schritt):** Investigator-Auftrag statt
Fix — die vom PO genannte Bean-ID ("a2ca") ist nicht verifiziert
(Repo-Suche ergebnislos, `state.json` verweist auf ein bereits
aufgeräumtes Test-Repo `/tmp/other-repo`). Auftrag: `flattenTree`/
`appendBeanNode` (view_browse_repo.go:40-120) UND `expandAncestorsOf`
(update.go:1426) auf generelle Bugs bei Epics-mit-Children prüfen,
UNABHÄNGIG von der konkreten ID — z. B. mit einem frisch angelegten
Test-Epic + 2-3 Children im aktuellen Repo (`bt-apmy` selbst hat bereits
11 Children und wird laut bt-b0w0s eigenem Smoke-Beleg korrekt
aufgeklappt — als Gegenprobe nutzbar: unterscheidet sich das Verhalten bei
einem ANDEREN Epic?). Zu prüfen: `m.expanded[id]`-Toggle (`setExpanded`,
update.go:1622) wird korrekt gesetzt beim Aufklappen · `appendBeanNode`s
Kind-Filter (Status/Archiv/Suche/Facet-aktiv?) blendet Children unter
bestimmten Filterzuständen versehentlich aus · Cursor-Sprung nach
Aufklappen (`treeCursorMove`) zeigt Kinder evtl. tatsächlich an, aber
außerhalb des sichtbaren Fensters (`windowAround`, leicht mit "zeigt nicht
an" verwechselbar).

**Ergebnis-Optionen für T1:**
- (a) Bug reproduziert mit einem generischen Test-Epic → exakter
  Repro-Pfad dokumentiert (Tasten-Sequenz, Filter-/Such-Zustand,
  Terminalgröße) → EIGENES Fix-Task-bean danach anlegen (nicht in diesem
  T1).
- (b) NICHT reproduzierbar (mirrort bt-tct9s eigenen B02-Präzedenzfall,
  epic-E9-plan.md Zeile 19-23) → bean mit Investigations-Ergebnis
  dokumentieren, PO-Retest mit exakter ID/Screenshot anfragen, KEIN
  weiterer Fix-Task in dieser Runde.

**Priorität high bleibt bestehen** (Tree-Expand ist Grundfunktion), aber
der ERSTE Schritt ist Repro, kein Blind-Fix.
