---
# bt-heg9
title: 'E7 — PO-Feedback R1: Detail-UX + Typ/Status/Prio-Glyphen'
status: todo
type: epic
priority: high
tags:
    - to-review
created_at: 2026-07-15T13:56:25Z
updated_at: 2026-07-15T18:48:32Z
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


## PO-Nachtrag 6 (2026-07-15): Layout-Stabilität Detail-Pane

**PF-12 — Kein Layout-Shift bei Selektion.** Im rechten Detail-Pane darf sich beim Auswählen/Fokussieren NICHTS verschieben: Der Platz für den Select-Marker (▷/▶ bzw. Fokus-Cursor) ist IMMER reserviert (feste Gutter-Spalte, auch im unselektierten Zustand — dann Leerzeichen/▷ statt nichts). Gilt für alle markierbaren Zeilen: Accordion-Sektionen UND Meta-Feldliste (PF-4). Test-Anforderung: Renderbreite/Spaltenposition jeder Zeile identisch mit und ohne Selektion (Golden- oder Assertion-Test).


## PO-Nachtrag 7 (2026-07-15): Fokus-Wechsel-Symmetrie

**PF-13 — tab/shift+tab und ←/→ konsistent paaren.** Ist-Zustand in Browse: tab wechselt den Fokus (Tree↔Detail), aber shift+tab macht NICHT den Rückweg — stattdessen geht's mit arrow-left zurück. PO: 'für Nutzer murks'. Soll: BEIDE Paare vollständig und symmetrisch: tab = Fokus vorwärts, shift+tab = Fokus rückwärts; arrow-right = nach rechts (in Detail), arrow-left = nach links (zurück in Tree) — jeweils in beide Richtungen funktionsfähig und in der lokalen Keybinding-Zeile (PF-11) korrekt ausgewiesen. Kollisionscheck: arrow-left/right haben heute ggf. Zweitbelegung (j/l-Äquivalent im Tree: collapse/expand) — Planner muss die Semantik sauber trennen (Fokus-Wechsel vs. Tree-Navigation) und in design-spec festhalten; PO-Intention: vorhersagbare Paare, kein Überraschungsverhalten.


## PO-Nachtrag 8 (2026-07-15): Review-Cockpit ENTFERNEN

**PF-14 — Review-Cockpit komplett raus (Feature-Removal).** PO wörtlich: 'widerspricht dem lean-stack Wesen und schafft wieder Zeremonie. Das Review-Cockpit ist Zauberei on-top und bitte raus nehmen. Das Review möchte ich in Zukunft direkt im Chat machen.'

Scope des Removals:
- viewReviewCockpit (View, view_review_cockpit.go) inkl. reviewState/reviewQueue/reviewGroup/reviewRework-Derivation, reviewCursor/reviewAccOpen-Modelfelder, clampReviewCursor
- Keybinding R + Cockpit-lokale Keys (a/x/o, n/p-Override) aus keymap.go/helpGroups (Drift-Guard-Test zieht mit)
- reviewStandMarkdown + Cockpit-y-Override (Yank-Review-Stand entfällt)
- reviewClickRow + zugehöriger Locktest (T4-I01-Test fliegt mit)
- Palette-Eintrag 'go to review cockpit'
- Reject-Form (form_reject_review.go) + PassReview/RejectReview-Mutationen im Datenlayer: Datenlayer-Funktionen KÖNNEN bleiben (harmlos, CLI-nah), TUI-Verdrahtung raus — Planner entscheidet ob Datenlayer-Removal mit (YAGNI) oder ohne
- 2 Cockpit-Goldens raus
- design-spec: §5-Review-Flow umschreiben (Tag-Konvention bleibt: Agents setzen to-review; Sichtbarkeit über Tree/Filter/Suche; Abnahme = Chat + beans-CLI durch PO/Supervisor), US-08 entsprechend umdefinieren (nicht mehr 'Cockpit', sondern 'Review-Signal sichtbar')
- E6-Auswirkung: US-08-Validierung gegen NEUE Definition

Planner-Hinweis Reihenfolge: Removal FRÜH einplanen (idealerweise erster Task) — dann müssen Glyphen-/Footer-/String-Umbauten das Cockpit nicht mehr mitziehen (spart Arbeit in allen Folge-Tasks).


## PO-Präzisierung zu PF-14 (2026-07-15): Review-Tag-Trio

Chat-Review steuert Rework über DREI Tags (kebab-case normalisiert): **to-review** (Agent meldet fertig) · **accepted** (PO nimmt an — Chat/CLI setzt Tag um bzw. schließt ab) · **rejected** (PO weist zurück → Agent-Rework). Ersetzt die bisherige Zwei-Tag-Konvention (to-review/rework). design-spec §5 entsprechend: Tag-Trio dokumentieren, KEINE TUI-Interaktion dafür — nur Sichtbarkeit (Tags im Tree/Detail/Filter sichtbar wie jedes Tag).


## PO-Nachtrag 9 (2026-07-15): p in Header-Globals

**PF-11-Ergänzung:** Keybinding 'p' (Repo-Picker, global von überall) fehlt in der Header-Global-Liste oben rechts. Soll: 'ctrl+r:Reload · ?:help · q:quit · esc:back · enter:open/confirm · p:repos' (Label-Wortlaut Englisch, Planner wählt konsistent kurz). Generell-Regel: ALLE globalen Bindings erscheinen oben rechts — auch ctrl+k (Command-Center) prüfen, das fehlt dort heute ebenfalls.


## Epic-Review-Zusammenfassung (für PO)

Stand 2026-07-15, T8-Abschluss (bean `bt-dsog`). Alle 13 PF-Design-Punkte
(PF-1..PF-8, PF-10..PF-14 — PF-9 existiert nicht, in design-spec.md/PO-
Nachträgen nie vergeben, keine Lücke im Scope) implementiert, per Golden-
Tests + Voll-Gate (2x `go test ./...`, `-race`, Goldens `-count=2`) +
tmux-Gesamt-Smoke (Scratch-Repo, Temp-HOME) belegt.

| PF | Inhalt | Task | Status/Beleg |
|---|---|---|---|
| PF-1 | Meta-Sektion immer offen, nicht kollabierbar | T4 (`bt-kyj5`) | 🟢 Smoke: `[1] META ▾` bleibt immer expandiert, nie `▸` |
| PF-2 | Zifferntasten 1-9 springen/wechseln Accordion-Sektionen direkt | T6 (`bt-t1uy`) | 🟢 Smoke: `1`/`3` springen direkt zu META/RELATIONS, Section auto-expandiert |
| PF-3 | Kopfblock rechts (bean-id/title/type-status-prio) | T4 | 🟢 Smoke: Detail-Pane zeigt Kopfblock über der Accordion, mit PF-4 verschmolzen (PO-Antwort Q01) |
| PF-4 | Meta-Feldliste (▷/▶-Cursor, 6 Felder) | T4 | 🟢 Smoke: title/status/type/priority/created_at/updated_at mit stabilem ▷/▶-Marker |
| PF-5 | Enter-Kaskade (Sektion→Feld→Edit-Overlay/Relations-Jump), `tab` bleibt einziger Detail-Fokus-Einstieg | T6 | 🟢 Smoke: `enter` auf status öffnet Value-Menu, `enter` auf Relations-Feld springt zur verlinkten Bean |
| PF-6 | Typ/Status/Prio als 1 kapitalisierter Buchstabe/Glyph + Farbe | T2 (`bt-2af1`) | 🟢 Smoke (capture -e, alle 5 Typen/Status/5 Prioritäten Farben verifiziert): M blue·E mauve·F mauve·T sky·B red; d blue·t green·i yellow·c/s subtext; ‼red·!yellow··text·↓/→subtext |
| PF-7 | UI-Sprache durchgängig Englisch | T3 (`bt-w9o8`) | 🟢 Smoke (Filter/Palette/Create/Delete/Tags/Parent/Blocking/Settings/Help/Lobby) + README jetzt vollständig Englisch (Q01 aus T3-Review umgesetzt) |
| PF-8 | Command-Center-Schema `verb entity`, ohne Doppelpunkt | T3 | 🟢 Smoke: „set status/tags/parent/blocking/title", „delete bean", „create bean", „go to backlog/browse/repo picker/settings", „filter facets", „search beans", „reload data" |
| PF-10 | Redundante Pane-Titel entfernt (Tree/Backlog/Detail) | T5 (`bt-uyzf`) | 🟢 Smoke: erste Zeile im Pane ist direkt die Suchzeile, kein „Tree"/„Backlog"/„Detail"-Titel mehr |
| PF-11 | Header/Footer-Keybinding-Split ohne Dopplung, `p`/`ctrl+k` im Header | T7 (`bt-m6at`) | 🟢 Smoke: Header exakt 7 Globals (`ctrl+r·ctrl+k·p·?·esc·enter·q`), Footer kontextsensitiv (Filter/Value-Menu/Tag-/Parent-/Blocking-Picker je eigene Bindings), kein Overlap |
| PF-12 | Kein Layout-Shift bei Selektion (fester Gutter) | T4 | 🟢 Code-Beleg (`metaSectionBody`: Marker IMMER `▷ ` oder `▶ `, nie ausgelassen) + Smoke visuell bestätigt |
| PF-13 | `tab`/`shift+tab` und `←`/`→` symmetrisch gepaart | T6 | 🟢 Smoke: `shift+tab` springt deterministisch zurück zu Tree (Border-Fokusfarbe verifiziert), `tab` bleibt bidirektionaler Toggle |
| PF-14 | Review-Cockpit vollständig entfernt | T1 (`bt-wmtb`) | 🟢 Smoke: `R` ist no-op (kein Crash, View bleibt Browse), Help-Overlay zeigt keine Reviews-Bindings mehr |

**Zusätzlich verifiziert (T8-Scope, kein eigener PF-Punkt):** Backlog-`enter`
ist dokumentierter No-op (T6-Review B01) — Smoke bestätigt (kein
Fokuswechsel, Border bleibt bei Tree/Backlog-Liste).

### Offene PO-Punkte (Entscheidungstabelle)

| Code | Hintergrund | Entscheidung | Empfehlung | Status |
|---|---|---|---|---|
| D01 | Footer zeigt in Browse/Backlog absichtlich NICHT `f`/`X`/`b`/`t`/`a`/`B`/`y` — alle bleiben über Help-Overlay (`?`) erreichbar/dokumentiert | PO entscheidet: Footer-Umfang so belassen (schlank) oder erweitern | Belassen — Footer bliebe sonst auf 80 Spalten noch enger (siehe I01); Help-Overlay deckt Vollständigkeit bereits ab | 🟣 Offen |
| I01 | 80-Spalten-Terminal (gängiges Maß) truncatet den 7-Globals-Header, `q:quit` kann abgeschnitten werden | PO entscheidet: Header umbrechen wie Footer, oder Prioritäts-Truncation-Reihenfolge | Header analog zum Footer umbrechen (Konsistenz, kein Informationsverlust) | 🟣 Offen |
| I02 | Overlay-lokale Footer-Hints (z.B. Palette/Help) wiederholen `enter`/`esc`, obwohl der Header sie bereits zeigt | PO entscheidet: gewollte Verstärkung (Sign-off) oder als Invarianten-Test/Fix behandeln | Sign-off — Redundanz in einem Overlay-Kontext ist visuell verstärkend, kein funktionaler Schaden | 🟣 Offen |
