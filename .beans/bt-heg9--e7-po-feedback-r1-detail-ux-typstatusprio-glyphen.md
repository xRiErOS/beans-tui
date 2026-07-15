---
# bt-heg9
title: 'E7 — PO-Feedback R1: Detail-UX + Typ/Status/Prio-Glyphen'
status: todo
type: epic
priority: high
created_at: 2026-07-15T13:56:25Z
updated_at: 2026-07-15T14:19:49Z
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


## PO-Nachtrag 2 (2026-07-15): Sprache + Command-Center-Schema

**PF-7 — UI-Sprache durchgängig ENGLISCH.** Aktuell gemischt (deutsch: 'Archivierte einblenden', 'status: setzen', 'titel: bearbeiten', 'bean: löschen', 'Kopiert: …', 'Review-Stand kopiert', 'keine offenen Reviews', Accordion 'Beziehungen'/'Historie', Lobby 'Repo filtern (Pfad)', …). ALLE nutzerseitigen Strings → Englisch (Accordion: META/BODY/RELATIONS/HISTORY — deckt sich mit PF-4-Layout). Betrifft auch Toasts, Filter-Menü, Cockpit-Leerzustand, Fehlertexte, Footer-Hints.

**PF-8 — Command-Center-Einträge nach Schema 'verb entity', OHNE Doppelpunkt:**
- set tags · set status · set priority · set title · set parent · set blocking
- go to backlog · go to browse · go to review cockpit · go to settings
- reload data · create bean · delete bean · search beans · filter facets (analog: 'filter beans'? — Planner: konsistentes verb entity wählen) · switch repo (bzw. 'go to repo picker' — Planner entscheidet konsistent, PO-Beispiele sind maßgeblich)

PO-Beispiele verbatim: 'set tags', 'set status', 'go to backlog', 'reload data', 'go to settings', 'set title'.


## PO-Nachtrag 3 (2026-07-15)

- **D01 REVIDIERT / PF-5 präzisiert:** KEIN enter als Detail-Fokus-Einstieg — der bestehende tab-Mechanismus gefällt dem PO und BLEIBT der Einstieg. Die enter-Kaskade gilt nur INNERHALB des Detail-Fokus: enter auf Accordion-Sektion → Feld-Navigation (▷/▶ mit i/k bzw. Pfeilen), erneutes enter auf Feld → Edit-Overlay (z.B. status → Status-Menü). Zifferntasten 1-9 (PF-2) unverändert gewünscht.
- **PO-Validierung (E6-Evidenz):** Filter-Logik in Backlog/Tree vom PO explizit als 'exzellent' abgenommen → als PO-Statement in US-05-Validierung übernehmen.
- **Tags:** bleiben (bestätigt). NEU: zentrale Tag-Definition über eigene Page → separates Feature-bean (bt-Verweis folgt), Scope-Entscheid v1 vs. v1.1 offen (Q03 an PO).


## PO-Nachtrag 4 (2026-07-15): Pane-Titel-Dopplung

**PF-10 — Redundante Pane-Titel entfernen.** Breadcrumb '> repo-b: Browse' bzw. '> repo-b: Backlog' trägt die View-Info bereits — der Pane-Titel ('Tree' / 'Backlog' + Unterstreichung) im linken Pane ist Dopplung und entfällt. Erste Zeile im Pane ist dann direkt die Suchzeile ('⌕ / search'). PO wörtlich: 'Es genügt, wenn es in den Breadcrumbs > repo-b: backlog angezeigt wird. Dann die Suche - sonst ist es obsolet.'

Konsistenz-Prüfauftrag an Planner: gleiche Logik für Review-Cockpit ('Review-Queue'-Titel vs. Breadcrumb 'Review-Cockpit') und Detail-Pane ('Detail'-Titel wird durch PF-3/PF-4-Kopfblock ohnehin ersetzt) — einheitlich entscheiden und in design-spec festhalten. Screenshots-Referenz: PO-Anhang (Browse vs. Backlog, repo-b).


## PO-Nachtrag 5 (2026-07-15): Keybinding-Split Header/Footer

**PF-11 — Keine Keybinding-Dopplung.** Header oben rechts zeigt die GLOBALEN Bindings: 'ctrl+r:Reload · ?:help · q:quit · esc:back · enter:open/confirm'. Footer zeigt AUSSCHLIESSLICH die view-lokalen Bindings der aktiven View — globale dort nicht duplizieren. (Nebeneffekt: entschärft VQA-I01 Footer-Umbruch, da Footer kürzer wird.)

**Q04 (offen, PO):** 'Leertaste für das Auswählen (fehlt noch m.E.)' — Semantik unklar: (a) Multi-Select von Beans im Tree/Backlog (Markieren für Bulk-Aktionen — wäre NEUES Feature), oder (b) die bestehende space/x-Toggle-Auswahl in Filter-Menü/Pickern (existiert — dann nur Footer-Sichtbarkeit herstellen)? Nicht raten — bis PO-Antwort nur PF-11 umsetzen.


## PO-Antwort Q04 (2026-07-15, GELÖST)

Q04 = Variante (b): Gemeint ist die BESTEHENDE Space-Auswahl in Forms/Overlays (aufgefallen beim Facetten-Filter: space/x toggle). KEIN Multi-Select-Feature. Anforderung → Teil von PF-11: Wenn ein Form/Overlay aktiv ist, zeigt die lokale Keybinding-Zeile (Footer) dessen lokale Bindings — inkl. 'space: select/toggle'. Kontextsensitiver Footer: View-lokal im Normalzustand, Overlay-/Form-lokal wenn eines offen ist.
