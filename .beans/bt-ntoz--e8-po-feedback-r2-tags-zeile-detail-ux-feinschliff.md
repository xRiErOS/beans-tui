---
# bt-ntoz
title: 'E8 — PO-Feedback R2: Tags-Zeile + Detail-UX-Feinschliff'
status: todo
type: epic
priority: high
created_at: 2026-07-15T20:18:42Z
updated_at: 2026-07-15T20:22:56Z
parent: bt-apmy
---

PO-Feedback-Runde 2 (2026-07-15, während D-Grilling + Live-Nutzung von bt). Umsetzung NACH Grilling-Freigabe (Confirmation Gate).

## D01-Entscheidung (aus Grilling, ENTSCHIEDEN)

Tags werden als Zeile in der Meta-Feldliste DIREKT NACH priority angezeigt (7. Feld, vor created_at). KEIN Tree-Suffix — rasches Filtern nach offenen Reviews via f-Filter (Tags-Facette) deckt den Überblick. Löst bt-gdkx (US-08 → PASS nach Fix; Kaskaden-enter auf tags → Tag-Picker analog status).

## Nebenfunde (PO verbatim strukturiert)

- **B01:** Pfeil-links wechselt aus dem Detail-Pane zurück, Pfeil-rechts wechselt aber NICHT hinein — asymmetrische Erwartung. Fix: Pfeil-links-Fokus-Exit ENTFERNEN; Fokus-Wechsel exklusiv tab/shift+tab. (Revidiert PF-13-Pfeil-Anteil: Pfeile sind Navigations-, keine Fokus-Tasten.)
- **B02:** Kopfblock-Zeile 'type: … status: … prio: …' springt beim Bean-Wechsel, weil Wertlängen variieren (epic vs feature, todo vs in-progress). Fix: feste Spaltenbreiten (Padding auf Maximal-Wortlänge je Feld: type→9/milestone, status→11/in-progress) — Ruhe in der Visualisierung, alles bleibt an seiner Stelle.
- **B03:** Kinderlose Beans zeigen im Tree ein Expand-Dreieck — Nutzer prüft unnötig auf Kinder. Fix: Dreieck nur bei Beans MIT Kindern (kinderlos: Leerraum gleicher Breite, kein Layout-Shift).
- **B04:** Nach tab-Fokus ist [1] META gewählt UND title erscheint bereits ▶-selektiert (mauve), obwohl die Feld-Ebene noch nicht betreten wurde. Fix: Feld-Marker (▶ + mauve) erst NACH explizitem Aktivieren der Feld-Ebene (enter auf Sektion); vorher ▷/neutral — Nutzer bekommt Feedback, dass die Aktivierung stattfand.
- **B05:** Accordion-Header '> [1] META  ▼' — das ▼/▸-Dreieck ist redundant (Zustand am Inhalt sichtbar) → entfernen.
- **B06 (Experiment):** Inaktive Accordion-Header sind grau = gleiche Farbe wie Meta-Label-Spalte → verwechselbar. Ausprobieren: Accordion-Header in TEAL (Catppuccin-Token). PO will das SEHEN (Screenshot-Vergleich vor Abnahme), explizit als Experiment markiert.

## Hinweise für Planner

- design-spec §15 um PF-15 (D01-Tags-Zeile) + PF-16 (B01-B06) ergänzen; PF-13-Pfeil-Revision dokumentieren.
- B02/B03/B04/B05 ändern Goldens legitim; B06 als eigener kleiner Golden-Vorher/Nachher-Vergleich für PO.
- bt-gdkx als Kind dieses Epics einhängen (D01-Fix = dessen Auflösung).


## Grilling-Nachträge (2026-07-15)

**D04 ENTSCHIEDEN (ersetzt validation.md-Empfehlung):** Header-Globals verkürzen auf genau 4: 'ctrl+k' (commands) · 'p:repos' · '?:help' · 'q:quit'. ctrl+r/esc/enter fliegen aus dem Header (bleiben im Help-Overlay + Footer-Kontexten). Passt in 80 Spalten ohne Truncation/Wrap. globalBindings() + Disjunktheits-Guard anpassen — die 3 degradierten Keys dürfen dann in lokalen Footer-Listen auftauchen wo relevant.

**B07:** Maus im Detail-Pane unvollständig: (a) Accordion-Sektionen nicht per Klick aktivierbar, (b) Meta-Feldzeilen nicht per Klick selektier-/editierbar. Fix: Klick auf Sektions-Header = Sektion aktivieren/expandieren; Klick auf Feldzeile = Feld selektieren; Doppelklick (oder Zweit-Klick auf selektiertes Feld) = Edit-Overlay (analog enter-Kaskade). Toast-Dismiss-Vorrang + Overlay-Guard-Mechanik aus E5-T4 wiederverwenden; detailClickRow analog treeClickRow (clickPaneGeometry, Kopfblock-Offset 5 Zeilen beachten!).
**B08:** Quit-Flow: (A1) Confirm-Text 'Really quit bt.' → 'Really quit bt?' (Frage). (A2) Quit-Kaskade zweistufig: q→enter führt zur LOBBY (statt Exit); aus der Lobby q→enter beendet die TUI. Randfall (Planner konservativ entscheiden + dokumentieren): Direktstart ohne konfigurierte Repos (Lobby wäre leer/sinnlos) → q→enter beendet direkt wie bisher.
