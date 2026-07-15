---
# bt-ntoz
title: 'E8 — PO-Feedback R2: Tags-Zeile + Detail-UX-Feinschliff'
status: todo
type: epic
priority: high
created_at: 2026-07-15T20:18:42Z
updated_at: 2026-07-15T20:18:42Z
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
