---
# bt-98cb
title: 'Detail-Accordion: voriges Segment kollabiert nicht bei Segment-Wechsel'
status: todo
type: bug
priority: normal
created_at: 2026-07-16T20:20:40Z
updated_at: 2026-07-16T21:29:12Z
parent: bt-tct9
---

PO-Nebenbefund, US-Review Runde 3 (2026-07-16): Meta-Accordion-Segmente (1 META/2
BODY/3 RELATIONS/4 HISTORY) klappen beim Wechsel auf ein anderes Segment nicht
mehr ein — mehrere bleiben gleichzeitig offen. Erwartet: Auswahl eines Segments
kollabiert die anderen (Standard-Accordion-Semantik).

Fundort vermutlich accordion.go/view_detail_bean.go (E2-Ära, m.accOpen).
Quelle: bt-tct9 US-Review Runde 3.



## Planner-Konkretisierung (2026-07-16)

**Befund bei Code-Review:** Die aktuelle Logik in `renderAccordion`
(accordion.go:69-141, `isOpen := n == open || n == 1`) und alle bekannten
`accOpen`-Setter (`update.go`: Zeilen 238/1053/1210/1220/1228/1409
Reset/digit-jump/up-down; `view_fullscreen.go`: 45/102/120;
`mouse.go:467` `mouseDetailClick`) setzen bei JEDEM Segmentwechsel exklusiv
`m.accOpen = <neue Sektion>+1` — beim Code-Lesen konnte KEIN Pfad gefunden
werden, der zwei NICHT-Meta-Sektionen gleichzeitig offen lässt. PF-1
(accordion.go-Doc) legt explizit fest, dass Sektion 1 (META) IMMER
ZUSÄTZLICH zur aktiven Sektion offen bleibt — das ist BY DESIGN, kein Bug.
Möglich, dass der PO-Befund genau DAS beschreibt (Meta bleibt neben der
neuen Sektion offen) und fälschlich als Regression gelesen wurde, ODER es
gibt einen Pfad, der beim reinen Code-Lesen nicht auffiel (z. B. ein
Render-Caching-Artefakt, oder ein Spezialfall bei sehr schnellen
Tastendrücken/Doppelklicks).

**Erster Schritt (PFLICHT vor jedem Fix):** Live-Repro mit `bin/bt` (tmux),
Segmentwechsel 1→2→3→4 in beiden Richtungen (Tastatur UND Maus, Tree UND
Backlog UND Vollbild), mit Pane-Capture-Beleg, WELCHE Sektionen sichtbar
offen bleiben. Falls das beobachtete Verhalten "Meta + aktive Sektion offen"
ist: bean als "kein Bug, PF-1 by-design" schließen (PO-Rückmeldung
einholen) — KEIN Fix nötig. Falls tatsächlich >2 Sektionen gleichzeitig
offen bleiben (echte Regression): Repro-Pfad exakt dokumentieren
(Tasten-Sequenz, Maus vs. Tastatur, welche View), dann erst Fix.

**Betroffene Kandidaten-Stellen (für die Untersuchung):**
`accordion.go::renderAccordion` (Zeile 69, `isOpen`-Berechnung) ·
`update.go` Zeilen 1208-1230 (`keyDetailFocus`, digit-jump + up/down) ·
`mouse.go:340-467` (`detailClickRow`/`mouseDetailClick`) ·
`view_fullscreen.go` (Reset-Stellen 45/102/120, falls der Bug spezifisch im
Vollbild-Modus auftritt).

**Offene Frage an PO** (siehe Planner-Abschlussbericht): war der beobachtete
Effekt "Meta bleibt zusätzlich offen" (PF-1, by-design) oder "zwei
NICHT-Meta-Sektionen gleichzeitig offen" (echte Regression)?


## PO-Antwort Q1 → Task-Neudefinition (2026-07-16, PF-18)

Repro-Frage beantwortet: Beobachtet war der PF-1-by-design-Effekt (META bleibt
zusätzlich offen). PO revidiert das Design (PF-18, siehe Epic bt-tct9): META
default GESCHLOSSEN — relevante Infos sitzen im Meta-Strip; META öffnet erst bei
aktiver Auswahl im Detail-Pane.

Kein Live-Repro mehr nötig. Neue Aufgabe: PF-1-Sonderfall entfernen —
`accordion.go:82` `isOpen := n == open || n == 1` → `isOpen := n == open`; alle
Stellen prüfen, die PF-1 kommentieren/annehmen (accordion.go Doc, update.go-Setter,
Goldens, Drift-Guard-Tests, design-spec §15 PF-18-Nachtrag).

Akzeptanz (ersetzt Repro-Akzeptanz):
- [ ] META beim Öffnen des Detail-Pane geschlossen, solange andere Sektion aktiv
- [ ] META öffnet bei Auswahl (Tastatur UND Maus), schließt beim Wechsel weg
- [ ] Meta-Strip unverändert (Informationsträger im Default)
- [ ] Goldens bewusst regeneriert, Diff je Snapshot beschrieben
- [ ] design-spec §15: PF-18-Nachtrag (PF-1 revidiert, nie stilles Umschreiben)
- [ ] tmux-Smoke: Sektionswechsel Tastatur+Maus, Tree+Backlog+Vollbild


## Prelude aus bt-lg68-Review (2026-07-16, non-blocking)

I01: `mouse.go:308` — Kommentar grammatikalisch verschachtelt (Klammer schließt
spät), rein kosmetisch. Beim ohnehin anstehenden Anfassen der Datei-Familie
glätten. Quelle: Reviewer bt-lg68, APPROVED-Run.
