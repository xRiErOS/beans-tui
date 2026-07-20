---
# bt-p78f
title: 'Body-Ueberschriften: Anker-Leiste + Pencil-Abschnitts-Edit'
status: todo
type: feature
priority: normal
created_at: 2026-07-20T09:24:15Z
updated_at: 2026-07-20T09:24:15Z
parent: bt-vy1q
---

PO-Befunde 2026-07-20 (#15 + #16): Der Body soll seine **Markdown-Ueberschriften als
Struktur** nutzbar machen. Zwei zusammengehoerige Faehigkeiten.

## #15 — Pencil vor Ueberschriften, Klick oeffnet nur diesen Abschnitt
Vor jeder Ueberschrift im Body ein Stift-Symbol. Klick darauf oeffnet **nur diese
Ueberschrift samt Abschnitt** im Editor — fokussiertes Bearbeiten statt des ganzen Bodies.

## #16 — Anker-Leiste der Ueberschriften oben im Body-Panel
```
╭─ Body ──────────────────────────────────────────────────────────── (e) ────╮
│[Ziel] [Definition of Done] [Herkunft] [Grundlagen] [Bau-Bloecke]           │
│                                                                            │
│   ## Ziel                                                                  │
```
Springt zur jeweiligen Ueberschrift.

## Gemeinsame Grundlage — deshalb ein bean
Beide brauchen dasselbe: **den Body in Ueberschriften-Abschnitte zerlegen** (Offset,
Ebene, Titel, Bereichsende). Das ist das eigentliche Stueck Arbeit; Pencil und Ankerleiste
sind zwei Oberflaechen darauf. Getrennt gebaut entstuenden zwei Parser.

## Fallstricke
- **Rueckschreiben ist der riskante Teil.** Nur einen Abschnitt zu bearbeiten heisst, ihn
  praezise wieder in den Gesamt-Body einzusetzen — inklusive korrektem ETag-Handling
  (beans-CLI `update --if-match`). Ein Fehler hier ueberschreibt fremde Aenderungen.
  Vorhandenen Mutationspfad benutzen, keinen zweiten bauen.
- **Codebloecke:** ` ``` `-Zeilen, die mit `#` beginnen, sind KEINE Ueberschriften. Ein
  naiver Zeilen-Parser zerschneidet Code.
- **Scroll-Kopplung:** Ankersprung muss ueber `adjustBoxFormScroll` laufen (bt-ze10), sonst
  zweiter Scroll-Pfad.
- **Ankerleiste bei 80 Spalten:** viele/lange Ueberschriften passen nicht in eine Zeile.
  Verhalten festlegen (kuerzen? scrollen? umbrechen?) und im 80-Spalten-Smoke pruefen.
- **Klickzonen:** Pencil und Anker brauchen Trefferzonen, die mit dem Scroll-Offset
  mitwandern. Genau das ist beim vorhandenen `detailBoxFormClickRow` bereits kaputt
  (siehe bean zum Klick-Offset) — erst dort reparieren, sonst wird der Fehler vererbt.

## Empfehlung zum Zuschnitt
In zwei Schritten bauen: (1) Abschnitts-Parser + Ankerleiste (nur lesend, risikoarm),
(2) Pencil + abschnittsweises Zurueckschreiben (mutierend, ETag-kritisch).

## Akzeptanz
- [ ] Ein Parser fuer beide Oberflaechen
- [ ] Codebloecke werden nicht als Ueberschriften missdeutet
- [ ] Ankersprung ueber den vorhandenen Scroll-Mutationspunkt
- [ ] Abschnitts-Edit schreibt korrekt zurueck, ETag-Konflikt getestet
- [ ] Verhalten der Ankerleiste bei 80 Spalten definiert und gesmoked
- [ ] voller Testlauf gruen
