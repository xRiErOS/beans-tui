---
# bt-39cl
title: Browse-Tree zeigt Children eines Epics beim Aufklappen nicht an
status: in-progress
type: bug
priority: high
created_at: 2026-07-16T20:20:40Z
updated_at: 2026-07-16T21:05:13Z
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


## Investigation 2026-07-16

**Ergebnis: (a) REPRODUZIERT** — generisch, nicht an die konkrete PO-ID "a2ca" gebunden.

**Root Cause:** Zusammenspiel aus dem Archiv-Default (`m.showArchived == false`,
`box_filter_facets.go:172` `beanMatchesArchive`) und dem gefilterten Tree-Flatten
(`view_browse_repo.go:299-330` `filteredBeanNode`). `visibleNodes()`
(`view_browse_repo.go:199-204`) geht IMMER über `flattenTreeFiltered`, sobald
`!m.showArchived` — also im Programm-Default, nicht nur bei aktiver Suche/Facette.

In `filteredBeanNode`: `hasKids` wird aus der UNGEFILTERTEN Kinderzahl gesetzt
(`children := idx.Children[b.ID]`, Zeile 303/311/328), der Aufklapp-Marker
(`treeNodeMarker`, Zeile 401-409) zeigt also IMMER ▸/▾ basierend auf der rohen
Struktur. Beim Aufklappen (offener Zweig, Zeile 314-328) werden aber nur
Kinder aufgenommen, die `match(b)` (= `m.beanMatches`, AND aus Search+Facets+
Archive) erfüllen. Sind ALLE direkten Kinder eines Epics `completed`/`scrapped`
und `m.showArchived` steht auf `false` (Default, auch nach Repo-Wechsel:
`update.go:247`), matcht keines — `anyChildHit` bleibt `false`,
`childNodes` bleibt leer. Der Epic-Knoten selbst rendert weiter (self matcht,
da sein eigener Status z. B. `in-progress` ist), inkl. `open: true` und
Marker ▾ — aber OHNE eine einzige Kind-Zeile darunter. Exakt der PO-Befund:
Epic hat Children, Aufklappen zeigt sie nicht.

**Repro-Pfad (Wegwerf-Repo, generisch):**
1. Repo: `repro-39cl` (`/tmp/.../scratchpad/repro-39cl`) — Epic `repro-39cl-9y6h`
   (status in-progress) mit 3 Children: 2× `completed`, 1× `scrapped`.
   Kontroll-Epic `repro-39cl-700c` ohne Children.
2. `bin/bt` im Repo-cwd starten, Cursor auf `repro-39cl-9y6h` (bereits an Pos. 1).
3. `l` (Right/expand) drücken → Marker wechselt ▸→▾, aber KEINE Kind-Zeilen
   erscheinen (Capture 1, 80×24).
4. `f` (Filter) → runter zu "Archive" → `space` (Show archived: ON) → `esc` →
   `l` erneut → jetzt erscheinen alle 3 Children eingerückt unter dem Epic
   (Capture 2, 120×40) — bestätigt den Archiv-Default als Ursache.
5. Getestet bei 80×24 UND 120×40 — Terminalbreite/-höhe ist NICHT der Faktor
   (kein Fenster-Scroll-Problem, `windowAround`-Hypothese verworfen für diesen Fall).

**Capture 1 (80×24, showArchived=false, nach `l`):**
```
▌▾ i E repro-39cl-9y6h Repro…
   i E repro-39cl-700c Repro…
```
(keine Kind-Zeilen zwischen Epic und dem zweiten Epic)

**Capture 2 (120×40, showArchived=true, nach erneutem `l`):**
```
▌▾ i E repro-39cl-9y6h Repro Epic mit a…
     c T repro-39cl-l4i7 Child 1 comple…
     c T repro-39cl-xldd Child 2 comple…
     s T repro-39cl-lyrn Child 3 scrapp…
   i E repro-39cl-700c Repro Epic ohne …
```

**Geprüfte Hypothesen:**

| Hypothese | geprüft wie | Ergebnis |
|---|---|---|
| `setExpanded`/`m.expanded[id]`-Toggle fehlerhaft | Code gelesen (`update.go:1622-1633`) + Marker-Wechsel ▸→▾ live beobachtet | Verworfen — Toggle funktioniert korrekt, Marker flippt |
| Cursor/Fenster-Scroll (`windowAround`) versteckt Children | 80×24 UND 120×40 getestet, beide zeigen 0 Zeilen bei showArchived=false | Verworfen — kein Scroll-Effekt, Zeilen fehlen strukturell |
| Archiv-Default filtert Children, Aufklapp-Marker bleibt inkonsistent zum Inhalt | Live-Toggle Filter→Archive→Show archived: Children erscheinen sofort | **BESTÄTIGT — Root Cause** |
| Verhalten unterschiedlich je Epic (bt-apmy Gegenprobe) | bt-apmys 11 Children vermutlich überwiegend NICHT alle archiviert → matched Default-Filter, daher unauffällig im Smoke-Beleg | Konsistent mit Root-Cause-These, nicht separat live gegengeprüft (kein Zugriff auf Original-Repo-Daten nötig, Ursache ist generisch) |

**Root-Cause-Datei:Zeile:** `internal/tui/view_browse_repo.go:303-328`
(`filteredBeanNode`, `hasKids` aus ungefilterter Kinderzahl + `open`-Zweig
schließt archivierte Kinder aus) im Zusammenspiel mit
`internal/tui/box_filter_facets.go:172` (`beanMatchesArchive`).

**Empfehlung:** Fix-Task JA. Skizze: entweder (1) Marker/`hasKids` in
`filteredBeanNode` auf die GEFILTERTE Kinderzahl umstellen, wenn `open &&
len(childNodes)==0` (Epic zeigt dann korrekt KEINEN Marker bzw. einen
"nichts sichtbar"-Hinweis, wenn alle Kinder durch den Archiv-Default
verdeckt sind), oder (2) beim Aufklappen eines Knotens, dessen sichtbare
Kinderzahl 0 ist trotz `hasKids`, einen Inline-Hinweis rendern ("3 archiviert,
versteckt — Show archived togglen"). (1) ist die konsistentere Lösung
(Marker lügt sonst weiterhin), (2) ist PO-freundlicher (erklärt WARUM). Sollte
diskutiert und in einem eigenen bean (Parent `bt-tct9`, wie die anderen
E11-Items) angelegt werden — nicht Teil dieses Investigator-Auftrags.


## PO-Antwort Q2 + Bestätigung (2026-07-16)

PO: "ich habe nur die automatische id angegeben, kein screenshot. Das war das bean,
welches ich zum testen verwendete." Aufgelöst: "a2ca" = `lean-stack-a2ca` — Epic im
lean-stack-Repo (PO-Demo lief per Commit b1212e0 gegen
~/Obsidian/tools/lean-stack). Dessen 4 Children sind ALLE completed → deckt sich
exakt mit dem Investigation-Root-Cause (alle Direkt-Children archiviert +
showArchived=false-Default ⇒ Marker ▾, keine Kind-Zeilen). Investigation damit
doppelt bestätigt (generisches Repro-Repo + Original-Fall).


## D01 entschieden: Platzhalter-Zeile (PO, 2026-07-16)

PO: "B passt" — Variante (b): Beim Aufklappen eines Epics, dessen sichtbare
Kinderzahl 0 ist (alle Direkt-Children archiviert, showArchived=false), rendert
der Tree eine Platzhalter-Zeile im Kind-Einzug, Stil gedimmt/Hint-Ton, Muster:
`N archiviert — f→Archive`. Marker bleibt ▾ (Epic HAT Children — Marker lügt
damit nicht mehr, die Zeile erklärt den Zustand). Verworfen: (a) Marker
unterdrücken (versteckt Existenz der Children), (c) archivierte Children bei
Expand zeigen (bricht Filter-Konsistenz).

Fix-Ort: `filteredBeanNode` (`view_browse_repo.go:303-328`) — Zweig `open &&
hasKids && anyChildHit==false` → Platzhalter-Node statt nichts. Platzhalter ist
NICHT selektierbar (Cursor überspringt) ODER selektierbar mit No-Op — Implementer
prüft, was `flattenTree`/Cursor-Logik sauberer trägt, Entscheidung im ERRATUM/
Notes dokumentieren. Zählwert N = Anzahl der weggefilterten Direkt-Children.

Abweichung vom Plan-Wortlaut (Item 1: "eigenes Fix-Task-bean danach"): Fix läuft
IN diesem Bug-bean — Investigation + Fix gehören zusammen, ein separates bean
wäre Kontext-Kopie ohne Gewinn.

Akzeptanz:
- [ ] Repro-Szenario (Epic, alle Children archiviert): Expand zeigt Platzhalter-Zeile mit korrektem N
- [ ] showArchived=true: unverändert, kein Platzhalter
- [ ] Epic mit Mix (1 offen, 2 archiviert): offene Children sichtbar, KEIN Platzhalter (nur Voll-Verdeckungs-Fall) — ODER Implementer begründet abweichend, falls Mix-Hinweis konsistenter
- [ ] Tabellentest für filteredBeanNode-Fälle + Golden falls Tree-Render betroffen
- [ ] tmux-Smoke im Repro-Repo (Investigation-Setup) beider Fälle
