---
# bt-ntoz
title: 'E8 — PO-Feedback R2: Tags-Zeile + Detail-UX-Feinschliff'
status: in-progress
type: epic
priority: high
tags:
    - to-review
created_at: 2026-07-15T20:18:42Z
updated_at: 2026-07-16T03:46:20Z
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


**D06 ENTSCHIEDEN (Footer-Neuspezifikation, ersetzt T7-Stand):**
1. Navigations-Keys (i/k/j/l, Pfeile) komplett RAUS aus dem Footer (intuitiv genug).
2. Reihenfolge: 'tab focus in · shift+tab focus out' zuerst, dann Aktionen: / search · s Status · c Create · d Delete · e Edit · b Backlog · t Tags · y Yank · a Parent · (r|B) Blocking.
3. Footer darf 2 Zeilen einnehmen.
4. Optik: Taste in TEAL, Aktions-Wort grau (subtext), KEIN ':' mehr — Farbtrennung ersetzt den Doppelpunkt. (Gilt konsistent auch für Header-Globals? Planner: einheitlich anwenden.)
Offene Präzisierung Q06: 'r Blocking' — Umbelegung B→r oder Tippfehler? / und f/X fehlen in der PO-Liste bewusst? (f war Supervisor-Empfehlung wegen D01-Filter-Workflow.)


**Q06 GELÖST (2026-07-15):** (1) Blocking-Key UMBELEGEN B→r (r seit PF-14 frei; Keymap+helpGroups+Drift-Guard+Doku nachziehen, B wird frei). (2) f Filter kommt MIT in den Footer. Finale Footer-Liste: tab focus in · shift+tab focus out · / search · f Filter · s Status · c Create · d Delete · e Edit · b Backlog · t Tags · y Yank · a Parent · r Blocking.

**B09:** Detail-View: inaktive ▹-Feldmarker sind WEISS statt grau — auf subtext/grau stellen (nur der aktive ▶ trägt Farbe/mauve, konsistent mit B04-Feedback-Logik).


**D02 ENTSCHIEDEN:** Backlog-Sort-Modus als dezenter Suffix in der Backlog-Suchzeile (subtext, z.B. '⌕ / search · sort prio'). Schließt das E2-Erbe.

**B10:** Detail-Fokus auf [2] BODY + 'e' öffnet fälschlich das Titel-Edit; 'enter' ist no-op (inkonsistent zu [1] META, wo enter das Overlay bringt). Fix: e/enter auf BODY → $EDITOR mit der Markdown-Datei (Body-Edit); e im Sektions-Kontext generell kontextsensitiv zur gewählten Sektion.
**B11:** Feld selektiert (z.B. status) + e/enter → Overlay zeigt ALLE editierbaren Meta-Felder statt NUR status. Fix: Feld-enter öffnet das Overlay ausschließlich für das gewählte Feld.
**B12:** 's' (Status menu) zeigt den gesamten Meta-Block; 't' dagegen korrekt nur Tags. Fix: Auftrennen — status, type, priority, tags bekommen je ein EIGENES Select-Overlay (das Sammel-Value-Menü zerlegen); Keys + Palette-Commands entsprechend (set status/set type/set priority/set tags je direkt ins Einzel-Overlay).
**B13:** ctrl+k (Command-Center) durchsucht auch alle Beans — störend. Fix: Palette zeigt NUR Commands (Bean-Treffer raus; Bean-Suche gehört zu '/'). design-spec-US-04-Anpassung nötig (bewusste Revision der E4-Entscheidung).
**B14:** Kein Weg, neue Tags zu definieren — weder Palette-Command noch (auffindbar) im t-Picker. Prüfen: box_picker_tag hat lt. Code einen 'Neuer Tag'-Modal (T3-Sweep-Fund, box_picker_tag.go:334) — kaputt oder nur nicht entdeckbar? Fix: (a) Neuanlage im t-Picker sichtbar machen (Footer-Hint!), (b) Palette-Command 'create tag'. Berührt D08/bt-6oyy (Tag-Page) — B14 ist die v1-Minimal-Lösung, Tag-Page bleibt v1.1.


**D03 ENTSCHIEDEN:** esc ist UNIVERSELL 'back' — auch in der Detail-Kaskade: Feld-Ebene → Sektions-Ebene → Fokus verlassen → (Grundzustand). PO-Begründung: die globalen Keybindings deklarierten 'esc:back' bereits — Verhalten muss dem Versprechen folgen. Konsistent mit B08-Quit-Kaskade (stufiges Verlassen). Schließt E7-T6-Review-I01. Prüfauftrag Planner: alle esc-Sites auf Einheitlichkeit (Suche, Filter, Picker, Lobby, Kaskade, Quit) — EIN mentales Modell: esc geht immer genau eine Ebene zurück.


## Grilling-Abschluss (2026-07-15, ALLE D entschieden + PO-Freigabe)

- **D05 ENTSCHIEDEN:** Overlay-Footer zeigen enter/esc — Sign-off (durch D04-Header-Kürzung keine Dopplung mehr).
- **D07 ENTSCHIEDEN:** Upstream-ETag-Issue (beans 0.4.2, Drift bei frischen Creates) NACH v1-Abnahme bei hmans/beans einreichen — Entwurf durch Agent, POST NUR MIT PO-FREIGABE (extern!). Minimal-Repro aus internal/data-Tests ableitbar.
- **D08 ENTSCHIEDEN:** Tag-Management-Page (bt-6oyy) → v1.1. B14 (create tag via Palette + t-Picker sichtbar) ist die v1-Minimal-Lösung.

**PO-Freigabe erteilt (Confirmation Gate passiert):** E8-Kette starten nach Kontext-Kompaktierung des Supervisors. Auftrag an Folge-Session/Supervisor:
1. T01: E8-Planner (Sonnet, frisch) — design-spec §15 um PF-15 (D01-Tags-Zeile) + PF-16 (B01-B14 + D02/D03/D04/D06-Revisionen) erweitern, Task-Schnitt, Task-beans; danach bewährte Implementer→Review-Kette (Muster: E7).
2. T02: validation.md §5 — D01-D08 auf 🟢/entschieden setzen, E8 als Umsetzungs-Verweis (kann der E8-Planner oder -Abschluss-Task miterledigen).
3. T04: bt-6oyy-Body um v1.1-Scope-Entscheid ergänzen.
4. T03 (NACH v1-Abnahme, nicht jetzt): Upstream-Issue-Entwurf, Post nur mit PO-Freigabe.
Prozess-Konventionen: Sonnet-Subagents, Fable nur Supervisor; Findings via body-append; Epics nie completed durch Agents (to-review).


## D07-Repro-Beleg (2026-07-16, Quelle: bt-y2iw-Review I01)

Reviewer reproduzierte den Upstream-ETag-Drift minimal: zwei `beans show` auf
UNVERÄNDERTE Datei liefern verschiedene ETags (947a8f6d… vs b554f0f6…) →
`update --if-match` schlägt fehl. beans-tui verhält sich korrekt (Conflict-Toast +
Recovery-File via writeConflictTempFile). Beleg für den T03-Issue-Entwurf
(NACH v1-Abnahme, Post nur mit PO-Freigabe).



## esc-Audit-Vervollständigung (2026-07-16, Nachtrag aus E8-T3-Review I01)

Die esc-Audit-Tabelle aus T3/bt-qbyq (Commit 6d0a9fe) deckte die 6
PO-benannten Bereiche (Suche/Filter/Picker/Lobby/Quit/Kaskade). Der
T3-Reviewer fand 5 weitere esc-Sites — ALLE bereits 'eine Ebene
zurück'-konform, kein Bug, reine Nachvollziehbarkeit:

| Site | Datei:Zeile | Verhalten | Befund |
|---|---|---|---|
| Backlog-View | view_browse_backlog.go:333 (keyBacklog) | esc → zurück zu Browse/Repo — eine Ebene | konform |
| Create-Confirm | box_confirm_create.go:147 (keyCreateConfirm) | esc/n → zurück zum Formular — eine Ebene | konform |
| Delete-Confirm | box_confirm_delete.go:123 (keyDeleteConfirm) | esc/n → Abbruch, Overlay zu — eine Ebene | konform |
| Formulare | forms_shared.go:164 (keyForm) | esc → Form verwerfen ohne Speichern — dokumentierte Design-Entscheidung | konform |
| Help-Overlay | overlay_shortcuts.go:63 (keyHelp) | esc/?/q → Help schließen — eine Ebene | konform |

ERRATUM (Zeilendrift): das Review-Prelude zitierte die Backlog-Site als
view_browse_backlog.go:307 — nach den T8-Edits liegt der esc-Case heute
auf Zeile 333 (`keys.Backlog`/`keys.Back`-Case). Alle übrigen vier
Zeilennummern stimmen exakt mit dem Ist-Code überein (verifiziert
2026-07-16, bt-6ppq).



## B06-Sign-off ausstehend (2026-07-16)

Das B06-Teal-Experiment (T2/`bt-czpf`, Commit `f910b78` `feat(tui):
EXPERIMENT — teal inactive accordion headers (B06)`) braucht den
PO-Sign-off VOR der Epic-Freigabe. Der Vorher/Nachher-Beleg liegt unter
**`docs/plans/v1-port/b06-experiment/`** (`README.md` + `before.txt` +
`after.txt`, ANSI-erhaltende `tmux capture-pane -e`-Captures gegen das
echte Binary). Kern: geschlossene Section-Header-Titel (BODY/RELATIONS/
HISTORY) wechseln von Hint-Grau `#7c7c7c` auf Teal `#8bd5ca`
(`theme.HeaderInactive`); META-Header + Meta-Label-Spalte bleiben grau.
Objektiver Beleg im Commit-Body: Teal-ANSI-Sequenzen 0→3 (exakt die drei
geschlossenen Header). Rollback = Ein-Zeilen-Änderung
(`theme.HeaderInactive` → `theme.Muted`); die D06-Header/Footer-Farben
hängen NICHT daran (`theme.BindingKey`/`BindingDesc` bewusst eigene
Tokens — Entkopplung beim Abschluss gegen den Ist-Code gegengeprüft,
hielt).


## E8-Abschluss (2026-07-16, bt-6ppq)

**8 Implementierungs-Tasks** (T1-T8) alle completed, **2 Fix-Runden**
nach Review (T4/`bt-duz7` R1: clickKey-Aliasing + zeitfensterloser
Zweitklick, Commit `6e1152e` · T5/`bt-1u0t` R1: Lobby-Exit-Keys
entkoppelt, Commit `d173927`). **Commit-Spanne:** Epic-/Plan-Kette ab
`6ea0fa6`, Implementierung `397a70f` (T1) … `6b2daa3` (T8),
Abschluss-Paket `3c084d2` (I01-Regressionstest) + `7dbab76`
(validation.md §7 + Review-Nachträge) + `1a6bb4f` (LESSONS-LEARNED).

**Voll-Validierung (2026-07-16):** Build/vet/gofmt clean · volle Suite
2× frisch grün (internal/tui 137.5s/137.2s, 462 Testfunktionen
repo-weit) · `-race` grün (140.1s, keine DATA RACE) · Goldens
`-count=2` byte-stabil, `git status testdata/` leer. Detail: bean
`bt-6ppq` „Validierungs-Output".

**Doku:** validation.md §7 „E8-Umsetzung" (Status-Tabelle D01-D08 +
B01-B14 mit Commits, US-08 → PASS pending Sichtprüfung) ·
docs/LESSONS-LEARNED.md neu (7 Einträge) · esc-Audit oben
vervollständigt (5 Zusatz-Sites, alle konform).

**Offene PO-Punkte (Epic-Review):**
1. **B06-Sign-off** — Sektion oben, Beleg `docs/plans/v1-port/b06-experiment/`.
2. **Q-PO Backlog-Footer:** braucht bei <82 Spalten 3 Zeilen (D06 sagt
   „darf 2"). Umschlagpunkt exakt 82 Spalten (Klartext 161 Zeichen >
   2×78). Optionen: (a) 3 Zeilen bei <82 akzeptieren · (b) Sort-Eintrag
   aus dem Footer streichen (Suchzeilen-Suffix trägt die Info bereits) ·
   (c) Wortkürzungen freigeben. Bewusst NICHT entschieden.
3. **US-08-Sichtprüfung** — `bt-gdkx` (Tag `to-review`, Auflösung
   dokumentiert im bean).
4. **Epic-Abnahme** — dieses bean steht auf `to-review`, Status-Wechsel
   nur durch den PO (Review-Flow §5).

Danach (separat, NICHT Teil dieses Eposs): D07/T03 Upstream-Issue-Entwurf
NACH v1-Abnahme, POST nur mit PO-Freigabe (Repro-Beleg oben).
