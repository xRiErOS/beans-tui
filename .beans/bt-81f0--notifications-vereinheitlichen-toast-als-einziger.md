---
# bt-81f0
title: 'Notifications vereinheitlichen: Toast als einziger Kanal'
status: todo
type: feature
created_at: 2026-07-17T06:27:18Z
updated_at: 2026-07-17T06:27:18Z
parent: bt-5uzr
---

NB aus PO-Review E11 (2026-07-17, Runde 3), PO verbatim:

"Wir haben 2 Systeme, um Notifications zu senden: 1) oben rechts das Toast und
2) unten rechts bspw der Error 'beans list: exit status 1: Error: querying
beans: syntax Error' Das ist nicht ok. Der kanonische Weg der Notifications ist
der Toast oben rechts. Dadurch wird auch unten die reservierte Zeile für die
Notifications frei."

Interpretation: PO-Design-Entscheidung — Toast (oben rechts) ist der EINZIGE
Notification-Kanal. Die untere Fehler-/Notification-Zeile entfällt; alle
bisherigen Konsumenten (z. B. Datenlayer-Fehler wie "beans list: exit status 1")
routen auf den Toast (Fehler-Variante toastErr, ohne Auto-Dismiss bei Fehlern?
— Planner klärt Persistenz-Verhalten). Frei werdende Zeile geht an den Content.

Akzeptanz-Entwurf:
- [ ] Inventar aller Nutzer der unteren Notification-/Error-Zeile (grep Render-Pfad)
- [ ] alle auf Toast umgeleitet (Fehler deutlich, nicht stumm verkürzt)
- [ ] untere reservierte Zeile entfernt, Layout-Höhe angepasst (Goldens!)
- [ ] Fehler bleiben lesbar bei langen Meldungen (Wrap/Truncation-Konzept im Toast)
