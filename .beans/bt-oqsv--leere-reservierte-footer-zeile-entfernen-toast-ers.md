---
# bt-oqsv
title: Leere reservierte Footer-Zeile entfernen (Toast ersetzt sie)
status: todo
type: task
priority: normal
created_at: 2026-07-20T07:26:22Z
updated_at: 2026-07-20T07:26:22Z
parent: bt-vy1q
---

**Nebenbefund N6 (PO, 2026-07-20).** Der Footer haelt eine leere Zeile vor — historisch reservierter Platz fuer Notifications. Das Toast-System (`overlay_show_toast.go`) uebernimmt das laengst, die Reservierung ist tot und kostet eine Bildschirmzeile.

Beleg: in `internal/tui/testdata/tree.golden` und `browse_boxform.golden` ist die vorletzte Zeile innerhalb des Rahmens leer.

## Ziel
Die reservierte Leerzeile entfernen, die Zeile geht an den Body/die Panes zurueck.

## Betroffen
- `internal/tui/view.go` — Footer-/Chrome-Komposition (wo die Zeile reserviert wird)
- ggf. die Hoehen-Rechnung (`bodyH`), damit der Rahmen NICHT waechst, sondern die Panes eine Zeile mehr bekommen

## Vorsicht
- Beruehrt praktisch ALLE Goldens (`tree`, `chrome`, `backlog`, `browse_boxform`, `browse_flat`) → bewusst regenerieren und im Commit begruenden
- Pruefen, dass Toasts weiterhin korrekt ueberlagern und nichts verdecken/springt
- Vorher verifizieren, dass die Zeile wirklich Notification-Reservierung ist und nicht Wrap-Reserve fuer den zweizeiligen Footer bei schmalen Terminals (bei 80 Spalten ist der Footer 2-3 Zeilen!)

## Akzeptanz
- [ ] Ursprung der Leerzeile im Code belegt (Datei:Zeile) vor der Aenderung
- [ ] Zeile entfernt, Panes gewinnen eine Zeile, Rahmen bleibt gleich hoch
- [ ] Toast-Verhalten unveraendert (Test/Smoke)
- [ ] 80-Spalten-Smoke: mehrzeiliger Footer bricht nichts
- [ ] Goldens regeneriert, voller `command go test ./...` gruen
