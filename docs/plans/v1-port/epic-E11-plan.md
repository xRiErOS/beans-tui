# Epos E11 — Nacharbeit R3+E10-Fast-Follows

Liefert die Aufräumrunde aus dem PO-Review-Zyklus 2026-07-16 (21 User-Stories live
durchgesprochen, `bt-tct9`/E9 + `bt-362n`/E10): 20 accepted, 1 rejected (US-05, Relations-
Scroll fehlt), plus 6 neue Follow-Up-beans aus Nebenbefunden. Sieben offene Items, KEIN
neues Epic-bean — alle sieben hängen bereits an ihren jeweiligen Ur-Epics (`bt-tct9`/E9 für
die Detail-Accordion-Cluster, `bt-362n`/E10 für die Tag-Themen), dieser Plan bündelt sie
NUR zur Ausführungsreihenfolge.

Quellen: `bt-b0w0` (Review-Rejection-Sektion + NB-2), `bt-98cb`, `bt-lg68`, `bt-39cl`
(alle Parent `bt-tct9`) · `bt-ct3k`, `bt-idm1`, `bt-9ipw` (alle Parent `bt-362n`) ·
`design-spec.md` §16 (Tag-Management-Architektur) + §F01 (Vollbild/Accordion-
Zustandsmodell) · Code-Stand 2026-07-16.

## Item-Übersicht

| # | bean | Typ | Parent | Inhalt | Datei-Familie | blocked_by |
|---|---|---|---|---|---|---|
| 1 | `bt-39cl` | Bug (high) | `bt-tct9` | Browse-Tree zeigt Epic-Children nicht — Investigator-first (ID unverifiziert) | `view_browse_repo.go` (`flattenTree`/`appendBeanNode`), `update.go` (`expandAncestorsOf`) | — |
| 2 | `bt-lg68` | Bug | `bt-tct9` | created_at/updated_at doppelt in META+HISTORY | `view_detail_bean.go` (`metaFields`, `metaFieldLabels`) | — |
| 3 | `bt-98cb` | Bug | `bt-tct9` | Accordion: voriges Segment kollabiert nicht — Repro-first (Code-Lesen fand keinen Bug, evtl. PF-1-by-design) | `accordion.go`, `update.go`, `mouse.go` | — |
| 4 | `bt-b0w0` (NB-2) | Task (reopen) | `bt-tct9` | Relations-Liste scrollt nicht (US-05-Reopen-Bedingung) | `view_browse_repo.go` (`renderAccordionPane`, `windowAround`) | — |
| 5 | `bt-ct3k` | Bug | `bt-362n` | Kein Feedback bei e/d auf freier Tag-Zeile | `view_tag_management.go` (`openTagMgmtRename`, `openTagMgmtDeleteConfirm`) | — |
| 6 | `bt-idm1` | Feature | `bt-362n` | `n` auf freier Zeile registriert (Adopt statt Blank-Create) | `view_tag_management.go` (`keyTagManagement`) | `bt-ct3k` |
| 7 | `bt-9ipw` | Feature | `bt-362n` | Tag-Picker Typeahead (Filter/Nav/Auto-Create) | `box_picker_tag.go` (`keyTagInput`, `tagInputBox`) | — |

**Reihenfolge-Begründung:**

- **#1 (`bt-39cl`) zuerst:** einzige `priority: high`, komplett unabhängig (eigene
  Dateien, kein Overlap mit irgendeinem anderen Item), Investigator-Ergebnis kann parallel
  zu allem anderen laufen. KEINE Fix-Task direkt — der PO-genannte Bean-Kurzname "a2ca" ist
  nicht verifiziert (Repo-Suche ergebnislos, `state.json` zeigt einen aufgeräumten
  Test-Repo-Pfad `/tmp/other-repo`); erster Schritt ist Repro mit einem generischen
  Test-Epic, NICHT die konkrete ID.
- **#2→#3→#4 (Detail-Accordion-Cluster, `bt-lg68`→`bt-98cb`→`bt-b0w0`):** alle drei
  berühren dieselbe Render-Familie (`view_detail_bean.go`/`accordion.go`/
  `view_browse_repo.go`), aus Dateikonflikt-Vorsicht empfohlen NACHEINANDER in einer
  Session zu bearbeiten (kein hartes `blocked_by`, da die konkreten Funktionen NICHT
  überlappen — `metaFields` vs. `renderAccordion`/`accOpen`-Setter vs.
  `renderAccordionPane`/`windowAround` — aber dieselbe Datei-Nachbarschaft). Reihenfolge
  innerhalb des Clusters: `bt-lg68` zuerst (kleinster, eindeutigster Fix, keine
  Ambiguität) → `bt-98cb` (braucht zuerst Live-Repro, siehe unten — Ergebnis könnte "kein
  Bug" sein) → `bt-b0w0`/NB-2 (Scroll-Fix baut auf demselben Render-Pfad auf, den `bt-98cb`
  ggf. anfasst — sauberer nach dessen Klärung).
- **#5→#6 (`bt-ct3k`→`bt-idm1`), hartes `blocked_by`:** beide ändern
  `view_tag_management.go` — `bt-idm1` trägt `blocked_by bt-ct3k` (in beans gesetzt), um
  Merge-Konflikte in derselben Datei zu vermeiden (identisches Muster zu `epic-E9-plan.md`s
  T5←T4). `bt-ct3k` zuerst, weil kleiner/eindeutiger (reiner Toast-Zusatz an zwei
  bestehenden No-Op-Zweigen) und weil `bt-idm1`s Toast-Konvention (Erfolgs-Feedback bei
  `n`-Adopt) explizit an `bt-ct3k`s Feedback-Konsistenz anschließt (Planner-Begründung im
  bean selbst: "ein neuer stiller Erfolgspfad wäre ein Rückschritt" ggü. dem, was `bt-ct3k`
  gerade behebt).
- **#7 (`bt-9ipw`) parallelisierbar, aber empfohlen zuletzt:** eigene Datei
  (`box_picker_tag.go`), KEIN Datei-Overlap mit #5/#6 (`view_tag_management.go`) — echte
  Nebenläufigkeit möglich. Empfehlung dennoch NACH #5/#6, weil dieselbe Feature-Familie
  (Tag-Registry-Konventionen: `collectTagCounts`, `tagManagementMarkerGlyph`,
  `saveTagDefsCmd`) — ein Implementer, der #5/#6 gerade gebaut hat, trägt den Kontext
  bereits frisch mit. Größtes Einzelstück dieser Runde — Bean-Entwurf selbst schlägt eine
  Drei-Schritt-Zerlegung vor (Filter-Logik → Navigation → Enter-Dispatch+Rendering),
  sequentiell IN der Task, kein eigenes Sub-bean pro Schritt (Umfang bleibt innerhalb der
  Ein-Task-Faustregel).

**Kein neues Parent-Epic:** alle sieben beans bleiben an ihren jeweiligen Ur-Epics
(`bt-tct9`/`bt-362n`) — das ist bereits so (siehe `beans list`-Baum), dieser Plan ist reine
Ausführungs-Choreografie, kein neuer Struktur-Knoten.

---

## Item 1: Browse-Tree zeigt Epic-Children nicht (`bt-39cl`)

**Investigator-first, KEIN Fix ohne Repro.** Die vom PO genannte Bean-ID ("a2ca") ist nicht
verifizierbar — Repo-Suche ergebnislos, `state.json` verweist auf ein bereits
aufgeräumtes Test-Repo (`/tmp/other-repo`). T1: Investigator prüft `flattenTree`/
`appendBeanNode` (`view_browse_repo.go:40-120`) und `expandAncestorsOf` (`update.go:1426`)
auf generelle Bugs bei Epics-mit-Children, unabhängig von der konkreten ID — z. B. mit
einem frisch angelegten Test-Epic + 2-3 Children im aktuellen Repo (Gegenprobe: `bt-apmy`
selbst, 11 Children, klappt laut `bt-b0w0`s Smoke-Beleg bereits korrekt auf — unterscheidet
sich das Verhalten bei einem ANDEREN Epic?). Kandidaten: `m.expanded[id]`-Toggle
(`setExpanded`), Kind-Filter bei aktivem Such-/Facet-/Archiv-Zustand, oder ein
Cursor-außerhalb-des-sichtbaren-Fensters-Fall (`windowAround`) statt eines echten
"zeigt-nicht-an"-Bugs.

Ergebnis-Optionen: (a) reproduziert → exakter Repro-Pfad dokumentiert, EIGENES Fix-Task-
bean danach; (b) nicht reproduzierbar (mirrort `bt-tct9`s eigenen B02-Präzedenzfall,
`epic-E9-plan.md` Zeile 19-23) → Investigations-Ergebnis dokumentiert, PO-Retest mit
exakter ID/Screenshot angefragt, kein weiterer Fix-Task in dieser Runde.

**Akzeptanz:** Repro-Ergebnis (a oder b) im bean dokumentiert · bei (a) Repro-Pfad exakt
reproduzierbar von einem frischen Agenten · bei (b) PO-Rückfrage formuliert.

---

## Item 2: created_at/modified_at doppelt in META und HISTORY (`bt-lg68`)

Root Cause exakt lokalisiert: `metaFields()` (`view_detail_bean.go:106-124`) trägt zwei
`readonly`-Einträge (`fmtTime(b.CreatedAt)`/`fmtTime(b.UpdatedAt)`, Zeilen 121-122),
gerendert über `metaFieldLabels` (Zeile 92, Einträge 6/7). `historieSectionBody` (Zeile
432-445) rendert Created/Updated zusätzlich — daher die Dopplung. Fix: beide Einträge aus
`metaFields()`/`metaFieldLabels` entfernen, META schrumpft von 7 auf 5 Felder
(title/status/type/priority/tags), HISTORY bleibt alleinige Quelle.

Ripple-Check PFLICHT: `keyDetailFocus`s Feld-Navigation (Klammerung an `len(fields)`),
`metaSectionBody`s Padding-Berechnung (bisher `"created_at:"` als längstes Label — nach
Entfernung wird `"priority:"` neuer längster Eintrag, Padding-Breite neu verifizieren,
nicht hart-kodiert lassen), jeder bestehende Test mit `len(metaFields(b))==7` oder
Feld-Index 5/6.

**Akzeptanz:** META zeigt nur noch 5 Felder · HISTORY unverändert · Pfeiltasten-Navigation
in META funktioniert für 5 Felder ohne Off-by-one · betroffene Tests angepasst · Golden-
Regen/-Gegenbeleg (META ist Teil von Tree/Backlog/Chrome-Goldens).

---

## Item 3: Detail-Accordion — voriges Segment kollabiert nicht (`bt-98cb`)

**Repro-first.** Code-Review von `renderAccordion` (`accordion.go:69`, `isOpen := n == open
|| n == 1`) und allen `accOpen`-Settern (`update.go` 1210/1220/1228, `mouse.go:467`,
`view_fullscreen.go` 45/102/120) fand KEINEN Pfad, der zwei NICHT-Meta-Sektionen
gleichzeitig offen hält — alle setzen exklusiv `accOpen = <neue Sektion>`. PF-1
(accordion.go-Doc) legt aber fest, dass META (Sektion 1) IMMER zusätzlich zur aktiven
Sektion offen bleibt — BY DESIGN. Möglich, dass der PO-Befund genau das beschreibt und
fälschlich als Regression gelesen wurde.

Erster Schritt: Live-Repro mit `bin/bt` (tmux), Segmentwechsel in beiden Richtungen
(Tastatur UND Maus, Tree UND Backlog UND Vollbild), Pane-Capture-Beleg WELCHE Sektionen
offen bleiben. "Meta + aktive Sektion" → kein Bug, PF-1-by-design, PO-Rückmeldung einholen,
bean schließen. ">2 Sektionen offen" → echte Regression, Repro-Pfad exakt dokumentieren,
dann Fix in denselben Kandidaten-Stellen.

**Akzeptanz:** Repro-Ergebnis dokumentiert (Bug bestätigt ODER PF-1-by-design) · bei
bestätigtem Bug: Fix + Regressionstest · PO-Antwort auf die im Abschlussbericht gestellte
Disambiguierungsfrage eingeholt.

---

## Item 4: RELATIONS-Liste scrollt nicht (`bt-b0w0`, NB-2, US-05-Reopen)

Reopen-Bedingung für die bereits `accepted` US-06/US-07 desselben Tasks — NUR NB-2 im
Scope. Root Cause: `renderAccordionPane` (`view_browse_repo.go:542-560`) übergibt den
gesamten Accordion-Output an `renderPane` (`render_shared.go:40-63`), dessen Zeilen-Cap
stumpf abschneidet, kein Scroll/Indikator — genau der von `renderDetailPane`s eigenem
Doc-Kommentar bereits antizipierte Fall ("a future accordion-pane change (e.g. scrolling)
can't drift between Tree and Backlog"). Wiederverwendbarer Baustein: `windowAround`/
`windowStart` (`view_browse_repo.go:456-482`) — dasselbe cursor-zentrierte Fenster-Prinzip,
das `treeRows` bereits nutzt; `scrollView` (`view.go:205-242`) liefert zusätzlich einen
`↑/↓`-Indikator-String (heute nur in `Chrome()`/Lobby/Help genutzt) — Implementer wählt,
Empfehlung `windowAround` (konsistent mit Tree/Backlog).

**Akzeptanz:** RELATIONS-Sektion bei vielen Einträgen zeigt Fenster um den selektierten
Eintrag statt Abschnitt · Pfeiltasten halten Cursor sichtbar (Auto-Scroll) · sichtbarer
Mehr-Einträge-Hinweis · kein Golden-Bruch bei wenigen Relations · tmux-Smoke mit
relations-reichem Bean (`bt-apmy`) PFLICHT.

---

## Item 5: Tag-Management — kein Feedback bei e/d auf freier Zeile (`bt-ct3k`)

`openTagMgmtRename` (`view_tag_management.go:412-421`) und `openTagMgmtDeleteConfirm`
(Zeile 537-548) geben bei `!row.defined` einen stillen No-Op zurück. Fix: beide rufen
`m.showToast(toastWarn, "Unregistered tag — modification not possible", "", nil, false)`
(PO-Wortlaut "unregistred tag - modification not possible" geglättet: Tippfehler
korrigiert, Em-Dash). Optionaler Context-Zusatz "n to define first" (Implementer-
Entscheidung).

**Akzeptanz:** `e`/`d` auf freier Zeile → Toast statt No-Op, Registry unverändert · `e`/`d`
auf definierter Zeile → unverändert · Tabellentest erweitert um Free-Row-Cases.

---

## Item 6: Tag-Management — `n` auf freier Zeile registriert (Adopt) (`bt-idm1`)

`blocked_by bt-ct3k` (dieselbe Datei). **Planner-Entscheidung (finalisiert die offene
Frage im Bean-Entwurf):** DIREKTES Registrieren ohne Zwischenschritt (kein vorbefüllter
Input+Enter) — PO-Wortlaut "ein Tastendruck ohne Zwischenstopp" ist eindeutig, Registrieren
ist nicht-destruktiv (kein Confirm-Bedarf wie bei Delete), ein Zwischenschritt wäre
Zeremonie ohne Informationsgewinn (Name steht schon fest). Betroffene Stelle:
`keyTagManagement`s `keys.NewTag`-Case (`view_tag_management.go:373-374`) — neue Logik
(mirrort `openTagMgmtRename`s Cursor-Row-Check): Cursor auf gültiger, freier Zeile → neue
`openTagMgmtAdopt()`-Funktion dispatcht `saveTagDefsCmd(m.client,
data.AddTagDefName(definedTagNames(m.tagMgmtRows), row.name), row.name)` direkt, mit
Erfolgs-Toast in `applyTagDefsSaved` (`update.go:472`). Sonst (keine Zeilen ODER Zeile
bereits definiert) → unverändertes Blank-Create.

**Akzeptanz:** freie Zeile + `n` → Direkt-Registrierung, kein Input-Submodus, Erfolgs-Toast
· definierte Zeile/keine Zeilen + `n` → unverändert · Cursor folgt der jetzt definierten
Zeile · Tabellentest erweitert.

---

## Item 7: Tag-Picker Typeahead — Filter/Navigation/Auto-Create (`bt-9ipw`)

Betrifft AUSSCHLIESSLICH `box_picker_tag.go` (Bean-Tag-Zuweisungs-Picker) — NICHT die
Tag-Management-Page (PO-bestätigt: dort keine Suche nötig). Größtes Einzelstück dieser
Runde, empfohlene Drei-Schritt-Zerlegung INNERHALB der einen Task (kein Sub-bean-Split):

1. **Filter-Logik:** `keyTagInput` (Zeile 270-310) filtert `m.tagItems` live per Substring
   (`strings.Contains`, kein Fuzzy-Scoring — YAGNI) bei jedem Tastendruck außer enter/esc;
   neues Feld `tagInputFiltered []tagCount`, neu berechnet nach jedem
   `m.tagInput.Update(msg)`.
2. **Navigation:** neues Cursor-Feld `tagInputSuggestCursor int` (analog `m.menu`), `up`/
   `down` bewegen ihn über `tagInputFiltered` (bislang ungenutzt in diesem Sub-Modus, keine
   Kollision mit `textinput.Model`, das einzeilig ist).
3. **Enter-Dispatch + Rendering:** `enter` mit nicht-leerem `tagInputFiltered` übernimmt
   den Tag am Cursor (`m.tagPending[tag]=true`, Copy-on-Write wie `toggleTagPending`, KEIN
   Registry-Write); leeres `tagInputFiltered` → heutiger Neuanlage-Pfad unverändert.
   `tagInputBox` (Zeile 397-405) rendert die gefilterten Vorschläge mit Cursor-Marker
   (mirrort `tagPickerBox`s `▸`-Konvention).

**Akzeptanz:** Tippen filtert sichtbar (definiert+frei, Union) · Pfeiltasten navigieren
Vorschläge · `enter` auf Treffer übernimmt existierenden Tag ohne Duplikat/Registry-Write ·
`enter` ohne Treffer legt neu an (D11 unverändert) · Tag-Management-Page unangetastet ·
Test-Suite grün.

---

## Offene Fragen für den PO

| # | Frage | Betrifft |
|---|---|---|
| Q1 | War der beobachtete Effekt bei `bt-98cb` "META bleibt zusätzlich zur aktiven Sektion offen" (PF-1, by-design — kein Bug) oder "zwei NICHT-Meta-Sektionen gleichzeitig offen" (echte Regression)? Code-Review fand keinen Pfad für Letzteres. | Item 3 |
| Q2 | Bei `bt-39cl`: welches Bean genau war gemeint (PO-Kurzform "a2ca" nicht auffindbar)? Exakte ID oder Screenshot würde die Investigation beschleunigen. | Item 1 |

## Selbst-Review

- **Jedes der 7 Items genau einem Plan-Abschnitt zugeordnet**, keines doppelt, keines
  vergessen — Quelle: der vom Supervisor übergebene 7-Item-Auftrag.
- **Kein Punkt bekommt einen Fix ohne Repro, wo die Faktenlage unklar ist:** Item 1
  (ID unverifiziert) UND Item 3 (Code-Review fand keinen Bug-Beleg) sind beide
  Investigator-/Repro-first, nicht Blind-Fix — mit expliziten Ergebnis-Optionen (Bug
  bestätigt vs. kein Bug) statt eines vorab angenommenen Fixes.
- **`blocked_by` nur wo ein echtes Datei-Konflikt-Risiko besteht:** einzige harte Kante
  Item 6←Item 5 (`bt-idm1`←`bt-ct3k`, beide `view_tag_management.go`); der Accordion-
  Cluster (Items 2-4) bekommt bewusst KEIN `blocked_by` (Funktionen überlappen nicht), nur
  eine empfohlene Sessions-Reihenfolge in der Prosa.
- **Jede Bean-Konkretisierung ist ein Anhang, kein Overwrite:** alle sieben beans über
  `--body-append` erweitert, bestehender Inhalt (inkl. Review-Historie, PO-Zitate) bleibt
  vollständig erhalten.
- **Keine Status-Änderung durch den Planner:** kein bean wurde auf `in-progress`
  zurückgesetzt oder auf `completed` gesetzt — das obliegt dem jeweiligen Implementer beim
  Start.
