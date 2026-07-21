---
# bt-p78f
title: 'Body-Ueberschriften: Anker-Leiste + Pencil-Abschnitts-Edit'
status: scrapped
type: feature
priority: normal
created_at: 2026-07-20T09:24:15Z
updated_at: 2026-07-21T08:48:11Z
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

## Prelude — Slice 2 Grounding (2026-07-21, nach Slice 1)

Slice 1 (#16 Ankerleiste + Klick-Sprung) ist fertig, gruen, gesmoked, committed (66dacaf). Parser 53e219f. Slice 2 (#15 Pencil + Abschnitts-Rueckschreiben) NOCH OFFEN — ETag-kritisch, bewusst als eigener Schritt (Bean-Empfehlung).

**Gegroundeter Editor-/Mutationspfad (keinen zweiten bauen):**
- `openBeanEditor` (editor.go:165) friert editorTarget/editorETag/editorSnapshot bei Open ein → `showRawCmd` liest vollen Raw → Suspend $EDITOR → `editorFinishedMsg` (editor.go:51) → `applyEditorFinished` (update.go:553) → `buildWholeEditDiff` gegen Snapshot → `UpdateWhole` (nur geaenderte Felder, ETag via editorETag).
- `parseRawBean` (editor.go:209) trennt Frontmatter/Body am 2. `---`.

**Slice-2-Bauplan:**
1. Pencil-Glyph vor jeder Ueberschrift im gerenderten Body (glyph in bodySectionBody/boxFormBodyPanelContent einweben; Klickzone analog Ankerleiste via render-grounded Test, Offset=3-Konvention).
2. Klick auf Pencil → $EDITOR mit NUR `b.Body[sec.start:sec.end]` seeden. Neuer Editor-Modus (z.B. editorSectionStart/End Felder in types.go) parallel zu editorTarget.
3. Bei Finish: newBody = snap.Body[:start] + edited + snap.Body[end:]; body-only WholeEditDiff; DERSELBE UpdateWhole + editorETag. NICHT parseRawBean (Section ist kein Frontmatter-Dok, reiner Body-Text).
4. **ETag-Konflikt-Test PFLICHT** (Akzeptanz): stale editorETag → UpdateWhole-Konflikt → Recovery-Tempfile-Konvention (wie applyEditorFinished PF-17), fremde Aenderung NICHT ueberschreiben.
5. Parser `parseBodySections` (box_body_sections.go) liefert die Byte-Spans schon splice-sicher (fence-safe). renderedHeadingLines/boxFormAnchorRow fuer die Pencil-Klickzeilen wiederverwenden.

**Offene Design-Q fuer Slice 2 (PO):** Pencil-Glyph-Wahl (Stift-Unicode?) + Position (vor Heading inline vs. rechts). Nicht geraten.

## Reasons for Scrapping (2026-07-21, PO-Entscheidung D01)
PO im Review: 'Wir bauen das Experiment mit dem Inhaltsverzeichnis zurueck. Die Relevanz ist gering und die Implementierung zu aufwaendig. Daher scrapped.'
- US-03 (Ankerleiste rendert) war accepted, aber US-04 (Klick-Sprung) real kaputt und US-05 (80c-Kuerzung) macht das Inhaltsverzeichnis nutzlos, sobald gekuerzt.
- Slice 2 (Pencil + ETag-Rueckschreiben) war ETag-riskant und nie gebaut.
- Kosten/Nutzen negativ → verworfen. Code-Rollback: git revert der Commits 66dacaf (Ankerleiste) + 53e219f (Parser). bt-adkn (Paging) NICHT betroffen.
