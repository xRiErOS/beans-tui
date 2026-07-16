---
# bt-tct9
title: 'E9 — PO-Feedback R3: Edit-Modell, Relations-Layout, Vollbild-Navigation'
status: in-progress
type: epic
priority: high
tags:
    - to-review
created_at: 2026-07-16T06:21:08Z
updated_at: 2026-07-16T20:59:50Z
parent: bt-apmy
---

PO-Feedback-Runde 3 (2026-07-16, nach Live-Test von v1.0/E8). Nummerierung R3-lokal.

## Entscheidungen (PO, ENTSCHIEDEN)

**D01 (Edit-Modell, supersedet E8-B10-Kontextsensitivität):** PO verbatim: "'enter' öffnet im details-view die forms für das edit eines Feldes, 'e' öffnet egal an welcher Stelle das gesamte bean in \$EDITOR." Interpretation: enter = Feld-Edit (Kaskade wie gehabt, B11-konform); e = IMMER ganzes Bean als Markdown-Datei in \$EDITOR — unabhängig von Sektion/Feld/Fokus, auch aus Tree/Backlog (PO: "egal an welcher Stelle"). Footer-Label 'e Edit' bleibt akkurat. E8-B10 (e kontextsensitiv zur Sektion) ist damit REVIDIERT; ctrl+e-Sonderpfad prüfen/vereinheitlichen.

**D02 (aus Q01 der E8-Abnahme, Interpretation "Guter Vorschlag" = Option b):** 'S Sort'-Eintrag aus dem Backlog-Footer streichen — Suchzeilen-Suffix '· sort <modus>' trägt die Info bereits. Footer passt dann in 2 Zeilen bei 80 Spalten. (Falls PO Option a/c meinte: trivial korrigierbar, im Report gespiegelt.)

**D03 (B06 Teal): ACCEPTED** — PO: "Ich habe es Live angeschaut und ist super." Sign-off erteilt, kein Rollback.

## Bugs/Nacharbeit (PO verbatim strukturiert)

**B01 (Ist-Beschreibung zu D01):** "Ich kann weiterhin nur mit enter status,type,priority,tags bearbeiten" — Feld-Edit nur via enter; Soll-Verhalten definiert D01 (e = Ganz-Bean-Editor).

**B02:** "Ich kann im tags overlay keinen eigenen tag ergänzen." — E8-B14 baute 'n New tag' (t-Picker + Footer-Hint + Palette 'create tag'); PO findet es nicht/es funktioniert nicht. URSACHE OFFEN — Investigator-Ergebnis abwarten (Kandidaten: via enter-auf-tags-Feld geöffneter Picker zeigt n-Modus nicht; stale Binary; Hint unsichtbar).

**B03:** Titel-Edit-Form ist single-line; bei langen Titeln muss sie multi-line umbrechen (huh-Form: Input→Text oder Wrap).

**B04 (RELATIONS-Sektion, Screenshot im Chat 2026-07-16):**
1. 'Fields:'-Zeile ist redundant zur Parent/Children/Blocked-By-Darstellung darunter (Dopplung raus).
2. Pfeiltasten selektieren nur 'Fields' — die Einträge in Parent/Children/Blocked By sind nicht per Pfeil erreichbar/wählbar.
3. Layout: Bean-Titel müssen mit HÄNGENDEM EINZUG umbrechen, sodass die Meta-Spalten (type-Glyph, Status, ID) nie vom Titeltext unterwandert werden. PO-Mockup:
   t M bt-apmy Hier steht ein langer Titel eines beans
               der so umbricht, dass die Uebersicht gewahrt
               ist
   c T bt-2jve Hier steht ein kurzer Titel eines beans

**B05 (= D02-Reject der E8-Abnahme, US-08/bt-gdkx reopened):** "Kein tag sichtbar -> Nacharbeit erforderlich." Widerspruch zum Abschluss-Smoke (live '▷ tags: ● to-review' verifiziert, Commit 397a70f) — URSACHE OFFEN, Investigator läuft. bt-gdkx (Tag rejected) hängt als Kind unter diesem Epic.

## Feature (PO-Idee, verbatim übernommen)

**F01 — 'v' Vollbild-Modus + Navigations-Pfad:**
- 'v' (view) öffnet das aktuell fokussierte Pane im Vollbild: Browse+links-fokussiert → Beans-Liste Vollbild; Browse+rechts-fokussiert → Detail-View Vollbild.
- In Listen-Vollbild: enter auf Bean → Detail-View-Vollbild.
- esc kehrt zum Browse zurück (aus Detail-Vollbild via Relations-Sprung: zurück zum Browse mit dem AKTUELLEN Bean selektiert).
- Im Detail-View: enter auf Relations-Eintrag (z.B. Children) öffnet das Ziel-Bean im Detail-View (Sprung). Setzt B04.2 voraus (Relations-Einträge selektierbar).
- ctrl+links / ctrl+rechts: Navigations-Pfad zurück/vor (History-Stack).
- Implementierungs-Hinweis: ctrl+arrow wird von manchen Terminals/tmux geschluckt — Planner: Verfügbarkeit prüfen, ggf. Fallback-Keys dokumentieren + im Footer/Help ausweisen.

## Prozess

Muster E8: Planner → Task-beans → sequentielle Implementer→Review-Kette. Reihenfolge-Empfehlung: Ursachen-Klärung B02/B05 zuerst (Investigator-Findings einarbeiten), dann D01-Edit-Umbau, B03, B04, D02-Footer, F01 zuletzt (größtes Stück, setzt B04.2 voraus).


## B05 REDEFINIERT (2026-07-16, PO-Klarstellung im Chat)

Kein Bug — US-08/D01 funktioniert (PO verbatim: "Ich sehe die Tags unter '1 META' aber
nicht im Meta-Strip"). Neue Anforderung: Tags ZUSÄTZLICH im Kopfblock-Meta-Strip
(Zeile unter dem Titel) anzeigen:

    type: epic  status: in-progress  prio: !  tags: to-review

Fix-Ort: detailHeaderBlock (view_detail_bean.go, E8-B02-Padding beachten — tags als
letzte Spalte, variable Breite unkritisch da Zeilenende; tagsInline wiederverwenden).
Investigator-Auftrag auf B02 reduziert.


## B02 Investigations-Ergebnis (2026-07-16, Investigator, HEAD 192f51a)

NICHT reproduzierbar — alle drei New-Tag-Pfade live verifiziert funktionsfähig:
(a) t-Picker: Footer 'n New tag' + Box-Hint, n öffnet Input (Placeholder 'New tag (a-z0-9, hyphen-separated)');
(b) Feld-enter auf tags: identischer Overlay (ein openTagPicker-Pfad, keine Divergenz);
(c) ctrl+k 'create tag' → direkt New-Tag-Input.
Randnotiz: 'create tag' erscheint in der Palette NUR bei fokussiertem Bean (overlay_palette.go:61).
Hypothesen für PO-Befund: stale Binary ODER Discoverability ODER Validierungs-Ablehnung
(nur a-z0-9/Hyphen — Großbuchstaben/Leerzeichen werden abgelehnt).
KEIN Code-Task — PO-Retest mit frischem bin/bt angefragt; falls Retest fehlschlägt,
mit exaktem Ablauf (welcher Pfad, welche Eingabe) als Bug reaktivieren.


## PO-Bestätigungen (2026-07-16, Chat)

- **D02 BESTÄTIGT (Option b + Präzisierung):** 'S Sort' fliegt aus dem Backlog-Footer;
  die S-Taste bleibt funktional, wird aber NUR im Help-Overlay ('?') dokumentiert.
  Suchzeilen-Suffix '· sort <modus>' bleibt die sichtbare Zustandsanzeige.
- **D01 BESTÄTIGT:** "egal an welcher stelle öffne ich mit 'e' das bean im $EDITOR" —
  e = Ganz-Bean-Editor überall (Tree, Backlog, Detail-View, jede Sektion/Feld-Ebene).


## B06 (2026-07-16, PO nannte es B05 — hier B06, B05 ist Kopfblock-Tags)

PO verbatim: "Das overlay für 'r' um die relations anzugeben ist viel zu schmal. So kann
man die beans nicht sauber lesen und bearbeiten. Höhe passt. Aber die Breite muss viel
weiter werden."

Screenshot-Befund: Blocking-Picker (~48 Spalten) bricht Einträge mitten in der ID um
('bt-' am Zeilenende, Rest nächste Zeile), Glyphen/IDs/Titel unlesbar verschränkt.

Interpretation (gilt analog für den Parent-Picker 'a' — vermutlich gleiche Box-Breite,
prüfen): Overlay-Breite deutlich erhöhen (Richtung ~80-90% der Terminalbreite bzw.
inhaltsbasiert), Einträge einzeilig wo möglich; bei Umbruch hängender Einzug analog
B04.3-Mockup (Meta-Spalten nie unterwandern). Höhe unverändert.


## B02 GESCHLOSSEN (2026-07-16, PO-Retest)

PO verbatim: "Retest gemacht: New Tag funktioniert jetzt mit frischem Binary."
Ursache war stale Binary (Investigator-Hypothese a bestätigt). Kein Code-Task.
Gegenmittel etabliert: zsh-Aliase bt-tui / bt-tui-build (Build zeigt Commit an).


## Review-Finding T1/F01 (2026-07-16, Reviewer bt-mtig) — offene PO-Frage Q02

B05 umgesetzt und APPROVED, aber Breiten-Nachrechnung des Reviewers: bei ~100 Spalten
Gesamtbreite (Split-View, accW≈61) belegt die Kopfblock-Zeile ohne Tag-Inhalt bereits 59
von 61 Zeichen — die neue Spalte zeigt dort nur `tags: …` ohne Klartext. Erst ab deutlich
breiteren Terminals (~160) erscheint das PO-Mockup-Bild vollständig. Mechanik konsistent
(bestehendes truncate), getestet, kein Bug — aber PO-Ziel "Tags sehen" bei Normalbreite
nur teilweise erfüllt (Label ja, Wert meist nicht).

**Q02 (PO):** Akzeptierte Grenze — oder Nacharbeit gewünscht (Optionen: Tags-Vorrang vor
Ellipse auf der Zeile, zweizeiliger Umbruch statt Hart-Truncate)? Bis zur Antwort keine
Nacharbeit; Task bt-mtig bleibt completed/APPROVED.


## T5-Review-Nachtrag (2026-07-16): F01 dokumentierter Nicht-Fix

T5-Reviewer-Finding F01 (low): wideModalWidths absoluter Boden 24 kann bei termW∈[5,27]
Terminal- und Deckelbreite überschreiten (kosmetischer Renderbruch in placeOverlay).
BEWUSST NICHT gefixt: mirrort exakt das bereits produktive clampModalWidth-Verhalten
(8 Aufrufstellen), praktisch unerreichbar (keine realen <28-Spalten-Terminals; kein Picker
vor erstem WindowSizeMsg öffenbar). Falls Minimal-Terminalbreiten je relevant werden:
gemeinsames Ticket für clampModalWidth+wideModalWidth. F02/F03 (Test-Härtungen) → Prelude T6.


## T7-Deviation D01 (2026-07-16) — offene PO-Frage Q03

F01-Kernmechanik umgesetzt, aber: `keys.Fullscreen` (v) wurde NICHT in die
Browse-/Backlog-Footer aufgenommen (bean-Wortlaut hätte es verlangt). Messung: 14 Einträge
kippen den Footer bei 80 Spalten von 2 auf 3 Zeilen (167 > 152 Zellen) — exakt die
Regression, die D02 (Sort-Entfernung) gerade behoben hat; D06 erlaubt max 2 Zeilen.
`v` ist voll funktional und in helpGroups() ('?') dokumentiert — mirrort den
PO-bestätigten D02-Präzedenzfall (Taste funktional, nur im Help).

**Q03 (PO):** v Help-only akzeptiert — oder Footer-Sichtbarkeit gewünscht (dann müsste
ein anderer Eintrag weichen, PO-Wahl welcher)? Bis zur Antwort gilt Help-only.


## US-Review 2026-07-16 (PO, Runde 1)

- US-01 (enter=Feld-Edit-Kaskade): accepted
- US-02 (e=Ganzes-Bean-im-$EDITOR überall): accepted


## US-Review 2026-07-16 (PO, Runde 2)

- US-03 (n=Neuer-Tag im Tag-Overlay): accepted
- US-04 (Tags im Kopfblock-Meta-Strip): accepted


## US-Review 2026-07-16 (PO, Runde 3)

- US-05 (Relations: Fields-Zeile raus): REJECTED — Feedback im Task-bean bt-b0w0 (Scroll-Blocker NB-2)
- US-06 (Relations-Einträge per Pfeil selektierbar): accepted
- US-07 (hängender Einzug bei langen Titeln): accepted

## Nebenbefunde (PO, Runde 3, 2026-07-16)

- NB-1: Meta-Accordion kollabiert das vorherige Segment nicht, wenn ein anderes
  Segment ausgewählt wird (mehrere Segmente bleiben gleichzeitig offen).
- NB-2: siehe bt-b0w0 (Relations-Liste braucht Scroll bei vielen Einträgen).
- NB-3: created_at/modified_at erscheinen doppelt (in '1 META' UND '4 HISTORY')
  — gehören nur nach '4 HISTORY'.
- NB-4: Bean 'a2ca' (Typ Epic, hat Children) zeigt seine Children im
  Browse-Tree links beim Aufklappen NICHT an. Exakte ID im Repo nicht
  gefunden (PO-Kurzform evtl. abweichend) — Investigator muss beim
  Nacharbeits-Start zuerst das betroffene Bean identifizieren.


## US-Review 2026-07-16 (PO, Runde 4)

- US-08 (v=Listen-Vollbild): accepted
- US-09 (v=Detail-Vollbild): accepted
- US-10 (enter im Listen-Vollbild -> Detail-Vollbild): accepted


## US-Review 2026-07-16 (PO, Runde 5)

- US-11 (Relations-Sprung im Detail-Vollbild): accepted
- US-12 (esc -> zurueck zu Browse/Backlog mit aktuellem Bean): accepted
- US-13 (History Back/Forward ctrl+links/rechts, [/]): accepted


## US-Review 2026-07-16 (PO, Runde 6)

- US-14 (Blocking/Parent-Picker breiter): accepted
- US-15 (Backlog-Footer ohne S-Sort-Eintrag, Help-only): accepted


## PF-18 — META-Accordion default geschlossen (PO, 2026-07-16)

PO verbatim (Antwort auf Q1/bt-98cb): "Meta kann geschlossen sein als default, da
die relevanten Informationen im meta-strip sitzen. Erst wenn ich in das detail-pane
gehe und meta wähle, klappt es auf."

Interpretation: REVIDIERT PF-1 (design-spec §15, "META Sektion 1 immer offen",
accordion.go:82 `isOpen := n == open || n == 1`). Neu: META verhält sich wie jede
andere Sektion — exklusives Accordion, offen NUR wenn aktiv gewählt. Der
ursprüngliche bt-98cb-Befund ("voriges Segment kollabiert nicht") war genau der
PF-1-Effekt — kein Regression-Bug, sondern Design-Änderungswunsch. Umsetzung in
bt-98cb (aus Repro-first wird Design-Change-Task). Spec-§15-Nachtrag PF-18 erfolgt
im Zuge der Umsetzung.
