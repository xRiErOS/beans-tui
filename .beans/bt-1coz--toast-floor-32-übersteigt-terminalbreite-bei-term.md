---
# bt-1coz
title: Toast-Floor 32 übersteigt Terminalbreite bei termW<32
status: completed
type: bug
priority: low
created_at: 2026-07-17T11:16:47Z
updated_at: 2026-07-18T15:05:03Z
parent: bt-5uzr
---

Reviewer-Finding B04 aus bt-0xrb-Review (2026-07-17, non-blocking, PRE-EXISTING Klasse): `toastBoxWidth(termW, contentW)` (overlay_show_toast.go:146-177) — der Floor `toastBoxMinWidth=32` überschreibt den terminal-abgeleiteten Cap `min(termW-4, 70)`. Bei termW<32 ist die Box breiter als das Terminal; live bei 30 Spalten reproduziert (Frame-Layout verschiebt sich, Header/Suchfeld aus dem Viewport). Verhalten NICHT durch bt-0xrb eingeführt — die alte fixe 36er-Konstante hatte dieselbe Klasse für termW<36; bt-0xrb verbesserte die Schwelle auf <32.

Fix-Rezept (Reviewer): capW muss gegen toastBoxMinWidth gewinnen, nicht umgekehrt — d.h. Endbreite = min(max(32, contentW), termW-4, 70), mit termW-4 als harte Obergrenze. Grenzwert-Test bei termW=30 ergänzen. Achtung devd-DD2-272-Parity-Kommentar am Floor prüfen/aktualisieren.

Akzeptanz:
- [ ] Toast nie breiter als termW-4, auch bei termW<32
- [ ] Grenzwert-Test termW=30 grün
- [ ] Bestehende Wachstums-/Clamp-Tests grün
