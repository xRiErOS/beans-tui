---
# bt-heg9
title: 'E7 — PO-Feedback R1: Detail-UX + Typ/Status/Prio-Glyphen'
status: todo
type: epic
priority: high
created_at: 2026-07-15T13:56:25Z
updated_at: 2026-07-15T14:03:50Z
parent: bt-apmy
---

PO-Feedback aus visueller QA (2026-07-15, Screenshots docs/_free-notes/vqa-2026-07-15/). Spec-ändernd — design-spec.md ist entsprechend zu erweitern (Quelle der Wahrheit). Implementierung VOR den E6-Validierungs-Tasks (E6 validiert sonst veralteten Stand).

## PO-Anforderungen (verbatim strukturiert)

1. **Meta immer offen:** Accordion-Abschnitt [1] Meta ist immer aufgeklappt (Default + nicht kollabierbar oder Default-offen — Planner entscheidet mit devd-Blick).
2. **Zifferntasten:** 1–9 expandieren/wechseln die Accordion-Sektionen direkt (im Browse-Detail; Cockpit-Footer listet '1…9:Section' bereits — Verhalten vereinheitlichen).
3. **Header rechts:** Inhalt noch offen — WARTET AUF PO (Q01, Codeblock war leer). Nicht raten.
4. **Meta-Layout neu:**
```
bean-id
NAME

type: xxxx    status: xxxx    prio: xxxx

[1] META
▷ title:      xxxx
▷ status:     xxxx
▷ type:       xxxx
▷ priority:   xxxx
▷ created_at: xxxx
▷ updated_at: xxxx

[2] BODY
[3] RELATIONS
[4] HISTORY
```
▷ = nicht selektiert, ▶ = selektiert (Akzentfarbe/Primärfarbe).
5. **Edit-Modus rechtes Panel:** enter (zusätzlich zu tab) wechselt in Detail-Fokus; Pfeiltasten/i-k navigieren Sektionen; enter auf Sektion → Feld-Navigation (▷-Cursor); enter auf Feld → passendes Edit-Overlay (z.B. status → Status-Menü). e bleibt $EDITOR (Default-Interpretation, Q02-Anteil).
6. **Glyphen-Ersatz (ASCII-Robustheit):** Typ als kapitalisierter Buchstabe mit Farbe: M=Milestone blue · E=Epic mauve · F=Feature mauve · T=Task sky · B=Bug red. Status als Kleinbuchstabe: c=completed subtext · s=scrapped subtext · t=todo green · i=in-progress yellow (draft: Planner-Vorschlag, vermutlich d subtext/blue). Priority: ‼/!=critical/high (red/yellow) · →=deferred · ↓=low · ·=normal.

## Offene PO-Fragen

- Q01: Inhalt Header rechts (Punkt 3) — Codeblock war leer.
- Q02: Farb-/Zeichen-Detail: critical vs. high exakt (‼ red / ! yellow angenommen), draft-Status-Zeichen+Farbe.

## Auswirkung E6

E6-Validierungs-Tasks erst nach E7 ausführen (blocked_by setzen sobald E6-Tasks existieren). US-12-Kriterium (devd-Look) wird durch dieses Feedback präzisiert, validation.md validiert gegen den NEUEN Stand.


## PO-Antworten (2026-07-15, Q01/Q02 GELÖST)

- **Q01 (Header rechts / PF-3):** Inhalt =
```
beans-id
beans-title

type: xxxx    status: xxxx    prio: xxxx
```
→ identisch mit dem Kopfblock aus PF-4 — PF-3 und PF-4-Kopf sind EIN Feature: Kopfblock oben im Detail-Panel. Kein separater App-Header-Umbau.
- **Q02 (Farben):** bestätigt: ‼ critical=red · ! high=yellow · · normal · ↓ low=subtext · → deferred=subtext · d=draft blue.
- **D01:** enter = Detail-View betreten — bestätigt.
- **D02:** Leitprinzip: Elemente möglichst schnell/einfach wählbar (Ziffern 1-9 + flüssige enter-Kaskade).
- **D03:** E6 nach E7 — bestätigt (blocked_by gesetzt: bt-wm4w/bt-9yvh ← bt-heg9).
