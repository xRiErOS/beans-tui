---
# bt-98cb
title: 'Detail-Accordion: voriges Segment kollabiert nicht bei Segment-Wechsel'
status: completed
type: bug
priority: normal
created_at: 2026-07-16T20:20:40Z
updated_at: 2026-07-16T21:45:43Z
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

## Implementer-Abschluss (2026-07-16, PF-18 Umsetzung)

**Summary:** PF-1s Meta-Sonderfall ("Sektion 1 nie kollabierbar") entfernt. `accordion.go:85`
`isOpen := n == open || n == 1` -> `isOpen := n == open` (mirror in `mouse.go`s
`detailClickRow`-Row-Counting-Formel, sonst Off-by-one bei Klick-Geometrie). META verhält sich
jetzt exklusiv-offen wie Sektion 2-4: geschlossen im Default-Browse (Cursor auf Bean, kein
Detail-Focus, `accOpen==0`), offen nur bei aktiver Auswahl (`accOpen==1`) via Ziffer `1`,
Pfeiltasten oder Klick auf `[1] META`-Header, schließt sofort bei Wechsel auf eine andere
Sektion.

**Akzeptanz-Checkboxen (aus PO-Antwort-Sektion oben):**
- [x] META beim Öffnen des Detail-Pane geschlossen, solange andere Sektion aktiv
- [x] META öffnet bei Auswahl (Tastatur UND Maus), schließt beim Wechsel weg
- [x] Meta-Strip unverändert (Informationsträger im Default) — `detailHeaderBlock` nicht angefasst
- [x] Goldens bewusst regeneriert, Diff je Snapshot beschrieben
- [x] design-spec §15: PF-18-Nachtrag (PF-1 revidiert, nie stilles Umschreiben)
- [x] tmux-Smoke: Sektionswechsel Tastatur+Maus, Tree+Backlog+Vollbild

**Geänderte Dateien:** `internal/tui/accordion.go` (isOpen-Fix + Doc), `internal/tui/mouse.go`
(gespiegelte isOpen-Formel + Doc; Prelude-Kommentar-Glättung separat committed),
`internal/tui/view_detail_bean.go` (Doc-Nachzug), `internal/tui/accordion_test.go` (RED->GREEN,
2 Tests umgebaut), `internal/tui/mouse_test.go` (`detailFocusModel`-Fixture + 1 Standalone-Test
ans neue Default-Verhalten angepasst), `internal/tui/testdata/{tree,backlog}.golden`
(regeneriert), `docs/plans/v1-port/design-spec.md` (§15 PF-18-Nachtrag vor §16 eingefügt).

**RED->GREEN:**
- `TestRenderAccordionExclusiveOpen`: Meta-Body darf bei `open=2` NICHT mehr rendern (vorher:
  MUSSTE rendern) — RED zitiert "PF-18: section 1 (Meta) body must NOT render when another
  section is open", GREEN nach `isOpen`-Fix.
- `TestRenderAccordionSectionOneAlwaysOpenRegardlessOfOpenParam` UMBENANNT zu
  `TestRenderAccordionSectionOneExclusiveLikeOthers`: Meta-Body nur bei `open==1`, RED für
  `open∈{0,2}` ("meta body rendered=true, want false"), GREEN danach.
- `TestDetailClickRowMapsMetaFieldClick` + 3 weitere Maus-Click-Tests fielen NACH dem
  `accordion.go`-Fix, weil `detailFocusModel` (Fixture) nie explizit `accOpen`/`detailFocus`
  setzte — funktionierte nur, weil PF-1 Meta immer offen hielt. Fixture jetzt korrekt auf
  echten FocusIn-Reset-Zustand gesetzt (`secCursor,accOpen,detailLevel,fieldCursor = 0,1,0,0`).
  `TestMouseDetailClickTreeClickIndexDoesNotAliasFieldClickKey` bekam zusätzlich ein explizites
  `m2.accOpen = metaSectionIdx + 1` vor dem Feld-Klick (realistischer erreichbarer Zustand:
  Meta war aus einem früheren Detail-Besuch offen geblieben) — Regressions-Kern des Tests
  (Klick-Key-Zahlenraum-Kollision) bleibt unverändert.

**Golden-Diffs:**
- `tree.golden`: META-Body (5 Feldzeilen title/status/type/priority/tags) unter `[1] META`
  entfernt, 5 Leerzeilen am Pane-Ende statt dessen, META-Header jetzt Teal/inaktiv statt
  Mauve/aktiv gefärbt (Default-Browse-Zustand, `accOpen==0`).
- `backlog.golden`: identisches Muster (gleiche Ursache, gleicher Diff-Typ).

**tmux-Smoke (Session `bt98cb74146`, 100x30, bin/bt frisch gebaut):**
- Default Browse (Cursor auf Bean, kein Detail-Focus): `[1] META` Header sichtbar, KEIN Body —
  bestätigt Golden-Fix live.
- `tab` (FocusIn): META öffnet automatisch als aktive Default-Sektion (`secCursor=0,accOpen=1`)
  — bestätigt "aktiv=offen bleibt legitim".
- Ziffer `3`: META kollabiert, RELATIONS öffnet exklusiv. Ziffer `1`: META öffnet wieder,
  RELATIONS kollabiert.
- `down`/`up`: identisches Exklusiv-Verhalten über Pfeiltasten (BODY öffnet, META kollabiert;
  zurück zu META via `up`).
- Maus: SGR-Klick auf `[3] RELATIONS`-Header (raw ESC-Sequenz via `tmux load-buffer`+
  `paste-buffer`, kein `-p`/Bracketed-Paste sonst schluckt bt die Bytes) öffnet RELATIONS,
  META kollabiert; Rück-Klick auf `[1] META` öffnet META, RELATIONS kollabiert — Maus UND
  Tastatur exklusiv bestätigt.
- Backlog-View (`b`): `accOpen` bleibt über View-Wechsel/`esc` aus Detail-Focus hinweg
  erhalten (kein Reset bei Back) — Default-Backlog zeigte META offen, weil vorher offen
  gelassen; Ziffer `3` schließt Meta dort identisch wie im Tree.
- Vollbild (`v` aus Backlog-Detail): erbt denselben Accordion-Zustand (RELATIONS offen, META
  zu); Ziffer `1` im Vollbild öffnet META exklusiv; `esc` verlässt Vollbild sauber zurück zu
  Backlog.
- Session sauber beendet (`q` + `tmux kill-session`), keine verwaisten Sessions zurückgelassen.

**Gates:** `go build -o bin/bt .` grün · `go test ./... -short` grün · `go test ./...` (voller
Lauf) 2x grün (138.9s / 139.2s, `-count=1` beim zweiten Lauf erzwungen) · `gofmt -l
internal/tui/` clean · `go vet ./...` clean.

**Deviations/ERRATA:** keine Zeilennummern-Abweichungen zu den Quellen-Angaben — alle
Fundstellen (`accordion.go:82`/`85`, `mouse.go:363`/`467`→`469` in der aktuellen Datei, siehe
unten) trafen wie im bean/Plan angegeben. Kleine Korrektur: `mouse.go`s `accOpen`-Setter liegt
nach dem I01-Kommentar-Fix bei Zeile 469 (nicht 467, Verschiebung durch die Prelude-Kommentar-
Glättung selbst — kein funktionaler Unterschied, reine Zeilenverschiebung).

**Notes für bt-b0w0 (Relations-Scroll, nächster Task derselben Familie):** RELATIONS-Sektion
öffnet jetzt wie jede andere Sektion nur bei aktiver Auswahl — die Scroll-Fenster-Logik
(`windowAround`) muss NICHT mehr mit einem gleichzeitig offenen META koexistieren (META ist
beim RELATIONS-Fokus immer zu), das vereinfacht den Höhenbudget-Rechenweg gegenüber der alten
PF-1-Situation (Meta-Body + Relations-Body gleichzeitig sichtbar, Höhenbudget musste beide
berücksichtigen). `renderAccordionPane`/`windowAround` (`view_browse_repo.go:456-482,542-560`)
weiterhin die empfohlenen Ansatzpunkte, jetzt mit einer Sektion weniger im Höhenbudget zu
rechnen, wenn RELATIONS selbst offen ist.
