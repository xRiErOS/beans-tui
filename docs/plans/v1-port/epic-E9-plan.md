# Epos E9 — PO-Feedback R3: Edit-Modell, Relations-Layout, Vollbild-Navigation

Liefert: das D01-Edit-Modell (`e` = Ganz-Bean-`$EDITOR` überall, `enter` bleibt reine
Feld-Kaskade — supersedet E8-B10), die RELATIONS-Sektion-Überarbeitung (B04: Dopplung
raus, Pfeil-selektierbar, hängender Einzug), zwei kleine UX-Fixes (B03 Titel-Multi-line,
B05 Kopfblock-Tags-Spalte), einen Footer-Feinschliff (D02), die Relations-Picker-
Breitenkorrektur (B06, Live-Nachtrag) und das größte Einzelstück dieser Runde: den
Vollbild-Modus `v` samt History-Stack (F01) — aus der PO-Feedback-Runde 3 (2026-07-16,
Live-Test von v1.0/E8) entstanden. Details/Zitate/Herleitung: `design-spec.md` §15 PF-17
(neuer Abschnitt „F01 — Vollbild-Navigation"), Epic-bean `bt-tct9` (vollständige
Auftragsquelle, D01-D03 PO-bestätigt, B01-B06, F01 verbatim strukturiert).

Quellen: `design-spec.md` §15 PF-17 (D01/D02/B03/B04/B05/B06 + eigener „F01"-Abschnitt) ·
Epic `bt-tct9` (alle Funde + Entscheidungen im Body, inkl. Live-Nachtrag B06) ·
`epic-E8-plan.md` (Muster für Aufbau/Detailtiefe) · `docs/LESSONS-LEARNED.md` (7 Guards
aus E8, hier v. a. Eintrag 3 „Kaskaden end-to-end testen" [F01-esc/History], Eintrag 4
„Grenzbreiten-Smoke bei 80 Spalten" [D02/B06] beachtet) · Code-Stand 2026-07-16.

Nicht in diesem Epos: **B02** (Tag-Neuanlage) — vom Investigator als NICHT reproduzierbar
eingestuft (alle drei New-Tag-Pfade live verifiziert funktionsfähig, `bt-tct9`
„B02 Investigations-Ergebnis"), PO-Retest mit frischem `bin/bt` angefragt; kein Task-bean
angelegt. Reaktivierung nur mit exaktem Reproduktionsablauf (welcher Pfad, welche
Eingabe) als neuer Bug, nicht in dieser Runde.

## Task-Übersicht

| Task | bean | Inhalt | Codes | blocked_by |
|---|---|---|---|---|
| T1 | `bt-mtig` | Kopfblock-Meta-Strip: Tags-Spalte | B05 | — |
| T2 | `bt-z4b1` | Edit-Modell: Ganz-Bean-`$EDITOR` via `e`/`ctrl+e` | D01, B01 | — |
| T3 | `bt-2v38` | Titel-Edit-Form multi-line | B03 | — |
| T4 | `bt-b0w0` | RELATIONS: Dopplung raus, Pfeil-selektierbar, hängender Einzug | B04 | — |
| T5 | `bt-4mo9` | Relations-Picker (Blocking/Parent) breiter | B06 | T4 |
| T6 | `bt-1e0t` | Backlog-Footer verliert Sort-Eintrag | D02 | — |
| T7 | `bt-13l7` | Vollbild-Modus `v` — Kernmechanik (Listen-/Detail-Vollbild, esc) | F01 (Kern) | T4 |
| T8 | `bt-1vbp` | Vollbild-Detail History-Stack (`ctrl+arrow` + `[`/`]`-Fallback) | F01 (History) | T7 |
| T9 | `bt-7pk2` | Abschluss (Voll-Validierung, Epic to-review) | — | T1, T2, T3, T5, T6, T8 |

**Reihenfolge-Begründung:** fünf Tasks (T1/T2/T3/T4/T6) haben keine Abhängigkeiten — sie
berühren disjunkte Dateibereiche (T1: `view_detail_bean.go::detailHeaderBlock`; T2:
`internal/data/{client,mutations}.go` + `editor.go` + `update.go`s Editor-Dispatch; T3:
`form_edit_title.go`; T4: `view_detail_bean.go::relationsSectionBody` +
`accordion.go::renderAccordion` — DISJUNKT von T1s `detailHeaderBlock` trotz derselben
Datei, mirrort E8s eigenes T1/T8-Präzedenzfall „disjunkte Funktionen, keine
Abhängigkeit"; T6: `view_browse_backlog.go::backlogLocalBindings`) und können in
beliebiger Reihenfolge landen. T5 (B06) folgt auf T4 (B04), weil beide denselben neuen
`hangingIndentWrap`-Helfer nutzen — T4 führt ihn ein, T5 verhindert damit einen zweiten,
unabhängigen Wrap-Helfer für strukturell dieselbe Glyph+ID+Titel-Zeilenform (Planner-
Entscheidung: B06 wurde bewusst NICHT in T4 hineingezogen, obwohl beide denselben Helfer
teilen — unterschiedliche Datei-Scopes, `view_detail_bean.go`/`accordion.go` vs.
`box_picker_blocking.go`/`box_picker_parent.go`/`box_filter_facets.go`, kleinerer,
review-freundlicherer Diff je Task). T7 (F01 Kernmechanik) `blocked_by` T4, weil F01 laut
PO-Wortlaut explizit „B04.2 voraussetzt" (Relations-Einträge müssen per Pfeil selektierbar
sein, bevor ein Relations-Sprung im Vollbild-Detail überhaupt einen Eintrag zum Springen
hat). T8 (F01 History-Stack) `blocked_by` T7 — baut direkt auf dessen
Vollbild-Zustandsmodell und dem Relations-Sprung-Zweig auf (der History-Push sitzt in
GENAU der Stelle, die T7 baut). T9 (Abschluss) `blocked_by` die sechs „Blatt"-Tasks des
Abhängigkeitsgraphen (T1/T2/T3/T5/T6/T8) — deckt transitiv auch T4/T7 ab, da T5 und T8
bereits an ihnen hängen (identisches Muster zu `epic-E8-plan.md`s T9).

**Golden-Strategie** (mirrort E8s Konvention): T1 (Kopfblock), T4 (RELATIONS-Body), T6
(Backlog-Footer) ändern sichtbaren Render-Output in Tree/Backlog/Chrome-Goldens
potenziell — nach JEDEM: `command go build -o bin/bt .`, dann `command go test
./internal/tui/ -run "TestTreeGolden|TestBacklogGolden|TestChromeGolden" -update`.
`git diff --stat internal/tui/testdata/` danach ansehen, JEDE geänderte Datei bekommt
eine Vorher/Nachher-Beschreibung im Commit-Body (Pflicht) — auch „unverändert" ist eine
gültige, explizit zu nennende Aussage. T2 (D01), T3 (B03), T5 (B06) ändern reine
Tastatur-/Formular-/Overlay-Logik OHNE Render-Output-Unterschied in den 3 Basis-Goldens
(Overlays/Formulare sind nicht Teil von Tree/Backlog/Chrome-Goldens) — ihr jeweiliger
Golden-Schritt ist ein GEGENBELEG (OHNE `-update`, MUSS grün bleiben), kein Regen. T7/T8
(F01) sind ein NEUER Render-Pfad (Vollbild) — ob eine eigene Vollbild-Golden-Suite
angelegt wird, ist Implementer-Entscheidung (im jeweiligen Task-bean als Option
vorgezeichnet); die bestehenden 3 Basis-Goldens bleiben in JEDEM Fall unverändert
(Split-Modus-Renderpfad ist strukturell unangetastet, Vollbild ist ein neuer Zweig davor)
— Gegenbeleg dafür in beiden Tasks Pflicht.

**Grenzbreiten-Smoke-Pflicht** (CLAUDE.md-Regel, LESSONS-LEARNED Eintrag 4): T6 (Backlog-
Footer schrumpft) und T5 (Picker-Breite wächst) ändern beide sichtbare Breiten/Umbrüche —
beide Task-beans tragen einen expliziten tmux-80-Spalten-Smoke-Schritt.

---

## Task 1: Kopfblock-Meta-Strip — Tags-Spalte (`bt-mtig`)

Deckt **B05** (bereits redefiniert im Epic-Body, Abschnitt „B05 REDEFINIERT"). Unabhängig
(kein `blocked_by`) — reiner `view_detail_bean.go::detailHeaderBlock`-Scope. Vollständige
Spezifikation, Code-Sketch und TDD-Schritte: bean `bt-mtig` (self-contained). Kurzfassung:
die bestehende `type: …    status: …    prio: …`-Kopfblock-Zeile bekommt eine vierte
Spalte `    tags: <tagsInline(b.Tags)>` (PO-Mockup: `type: epic  status: in-progress
prio: !  tags: to-review`), taglos `(none)` — `tagsInline` (render_shared.go, seit PF-15
kein toter Code mehr) bekommt hier einen zweiten Aufrufer.

**Akzeptanz-Checkliste:** Kopfblock zeigt Tags-Spalte · taglos zeigt `(none)` ·
type/status/prio-Padding aus E8-B02 unangetastet · Goldens verifiziert · voller Lauf grün.

---

## Task 2: Edit-Modell — Ganz-Bean-`$EDITOR` via `e`/`ctrl+e` (`bt-z4b1`)

Deckt **D01** (supersedet E8-B10) und **B01** (dessen Ist-Beschreibung — kein separater
Fix, D01 IST das Soll-Verhalten). Unabhängig (kein `blocked_by`). Vollständige
Spezifikation, Code-Sketches (neuer `data.Client.ShowRaw`/`UpdateWhole`, `rawBeanFrontmatter`-
Parsing via `gopkg.in/yaml.v3`, `openBeanEditor`, `applyEditorFinished`-Umbau) und
TDD-Schritte: bean `bt-z4b1` (self-contained, umfangreichster Sketch dieser Runde).
Kurzfassung:

`enter` bleibt AUSSCHLIESSLICH die PF-5-Feld-Kaskade — der E8-B10-Sonderfall „`enter` auf
`[2] BODY` öffnet `$EDITOR`" entfällt ersatzlos (**B10-Revision**), `keyDetailFocus`s
entsprechender Zweig (update.go) fällt zurück auf den generischen No-Op. `e`/`ctrl+e`
(eine `keys.Editor`-Bindung, unverändert) öffnen ab sofort UNBEDINGT denselben neuen
`openBeanEditor(b)`-Pfad — egal welche Sektion/Feld-Ebene, auch ohne aktiven Detail-Fokus,
aus Tree/Backlog/Detail. `keyNodeAction`s bisherige e/ctrl+e-Verzweigung (Body vs.
Titel-Form) entfällt zugunsten EINES unconditional Calls; `e`/`ctrl+e` öffnen nie mehr das
Titel-Edit-Form.

Seed-Text: `beans show <id> --raw` (verifiziert byte-identisch zum on-disk `.beans/*.md`-
Format — kein selbstgebautes Templating, die CLI bleibt die eine Autorität). Rückweg:
`parseRawBean` (neu, `gopkg.in/yaml.v3`, bereits Projekt-Dependency) splittet Frontmatter/
Body, Diff gegen den bei `$EDITOR`-Open eingefrorenen Snapshot (neues Modelfeld
`editorSnapshot`), EIN kombinierter `data.Client.UpdateWhole`-Call (mirrort SetTags'/
SetBlocking's Single-Etag-No-Cascade-Konvention). Validation-/Parse-Fehler: Recovery-
Tempfile (mirrort `writeConflictTempFile`), kein Datenverlust. **Bekannte Grenze**
(Dokumentationspflicht, kein Bug): `created_at`/`updated_at`/die ID-Kopfzeile sind im
Editor-Text sichtbar, aber nicht wirksam editierbar (kein `beans update`-Flag existiert
dafür) — muss im Commit-Body als ERRATUM/Deviation stehen.

**Akzeptanz-Checkliste:** `e`/`ctrl+e` überall identisch (Tree/Backlog/Detail, jede
Ebene) · nie mehr Titel-Form · `enter` auf BODY wieder No-Op · Seed = `beans show --raw` ·
nur geänderte Felder im kombinierten Update-Call · Recovery bei Validation-/Parse-Fehler ·
ETag-Konflikt weiterhin über bestehenden `conflictWithRecovery`-Pfad · „Bekannte Grenze"
dokumentiert · kein Golden ändert sich · voller Lauf grün.

---

## Task 3: Titel-Edit-Form wird multi-line (`bt-2v38`)

Deckt **B03**. Unabhängig (kein `blocked_by`) — reiner `form_edit_title.go`-Scope.
Vollständige Spezifikation: bean `bt-2v38`. Kurzfassung: `buildEditTitleForm` tauscht
`huh.NewInput()` gegen `huh.NewText().Lines(3).ExternalEditor(false)` (huh v1.0.0 bringt
eine Multi-Line-Textarea-Komponente nativ). `.ExternalEditor(false)` ist PFLICHT —
`huh.Text` hat einen eigenen `ctrl+e`-Editor-Suspend-Mechanismus (Default `true`), der mit
D01s app-weitem Ganz-Bean-Editor kollidieren würde. `.Lines(3)` ist eine Planner-Schätzung
(PO nannte keine Zeilenzahl). Betrifft NUR das Formular hinter `enter` auf dem
`title:`-Feld, nicht D01s Ganz-Bean-Editor (dort ist der Titel als Teil der Frontmatter
sowieso frei editierbar).

**Akzeptanz-Checkliste:** `huh.NewText()` mit `.Lines(3)` · `.ExternalEditor(false)`
gesetzt · langer Titel bricht sichtbar um · `nonEmpty`-Validierung funktioniert · Submit
liefert vollen String verlustfrei · kein Golden ändert sich · voller Lauf grün.

---

## Task 4: RELATIONS-Sektion — Dopplung raus, Pfeil-selektierbar, hängender Einzug (`bt-b0w0`)

Deckt **B04** (alle drei Teilpunkte, EINE zusammenhängende Änderung). Unabhängig (kein
`blocked_by`). Vollständige Spezifikation, Code-Sketches (neuer `hangingIndentWrap`-
Helfer) und TDD-Schritte: bean `bt-b0w0` (self-contained). Kurzfassung:

1. Die separate `Fields:`-Zeile (`fieldStrip`, accordion.go) entfällt — RELATIONS war ihr
   einziger verbleibender Aufrufer, `fieldStrip` wird komplett gelöscht (Compiler-
   gesteuerte Verifikation, Muster PF-14/B13-Removal).
2. Ersatz: `relationsSectionBody` bekommt dieselbe ▷/▶-Cursor-Konvention wie
   `metaSectionBody` (PF-4) — Signatur wächst um `(active bool, fieldIdx int)`, jede
   Zeile trägt einen eigenen Marker. Die Pfeiltasten-Navigation selbst
   (`keyDetailFocus`s bestehende `fields`-Iteration) ändert KEINEN Code — sie zeigte
   immer schon auf dieselbe Slice, nur die einzige VISUALISIERUNG war bisher die
   (jetzt entfernte) Strip statt der Zeilen selbst.
3. Neuer Helfer `hangingIndentWrap(prefix, text string, w int) string`: Folgezeilen eines
   umbrechenden Titels richten sich unter dem (PRO ZEILE individuell breiten, da
   ID-Länge variiert) Präfix-Ende aus statt auf Spalte 0 zurückzufallen — löst den
   PO-Mockup-Bug (`relationsSectionBody`s bisheriges Blanket-`wrapText` über den
   gesamten verketteten Block war die Ursache).

Wiederverwendet von Task 5 (B06) für die Relations-Picker.

**Akzeptanz-Checkliste:** keine „Fields:"-Zeile mehr · Relations-Einträge tragen eigene
▷/▶-Marker · Pfeiltasten selektieren sichtbar die echten Zeilen · hängender Einzug
verifiziert (Meta-Spalten nie unterwandert) · `fieldStrip` vollständig entfernt · `enter`
auf selektierter Zeile springt weiterhin korrekt · Goldens verifiziert · voller Lauf grün.

---

## Task 5: Relations-Picker (Blocking/Parent) breiter (`bt-4mo9`)

Deckt **B06** (Live-Nachtrag 2026-07-16, PO-Fund während dieser Planungssession).
`blocked_by` `bt-b0w0` (T4, teilt dessen `hangingIndentWrap`-Helfer). Vollständige
Spezifikation: bean `bt-4mo9`. Kurzfassung: neuer `wideModalWidth(termW int) int`
(box_filter_facets.go, neben `clampModalWidth` — dessen Boden-only-Clamp NIE nach oben
skaliert, `wideModalWidth` dagegen ≈85% der Terminalbreite, Boden 60, Deckel `termW-4`).
`blockingPickerBox`/`parentPickerBox` wechseln von `clampModalWidth(48, m.width)` auf
`wideModalWidth(m.width)`. Zeilen-Rendering nutzt denselben `hangingIndentWrap` wie T4
(strukturell identische Glyph+ID+Titel-Zeilenform) statt der bisherigen Ein-Zeilen-
Konkatenation — verhindert den vom PO gezeigten Mitten-in-der-ID-Umbruch. Höhe
(`parentPickerRowBudget = 14`) bleibt unverändert (PO: „Höhe passt").

**Akzeptanz-Checkliste:** `wideModalWidth` skaliert mit Terminalbreite · beide Picker
nutzen ihn · kein Umbruch mehr mitten in der ID · Höhe unverändert · tmux-Smoke bei 80
Spalten belegt · keine Basis-Goldens betroffen · voller Lauf grün.

---

## Task 6: Backlog-Footer verliert Sort-Eintrag (`bt-1e0t`)

Deckt **D02** (PO-bestätigt, Option b + Präzisierung). Unabhängig (kein `blocked_by`) —
reiner `view_browse_backlog.go::backlogLocalBindings`-Scope. Vollständige Spezifikation:
bean `bt-1e0t`. Kurzfassung: `backlogLocalBindings()` gibt künftig unverändert
`browseRepoLocalBindings()` zurück (kein angehängtes `keys.Sort` mehr). `S` bleibt
funktional (`keyBacklog`s Case unverändert), bleibt in `helpGroups()` gelistet
(unverändert), Suchzeilen-Suffix `· sort <modus>` (E8/D02) bleibt die sichtbare
Laufzeit-Anzeige.

**Akzeptanz-Checkliste:** `keys.Sort` nicht mehr in `backlogLocalBindings()` · `S`
funktional · in `helpGroups()` weiterhin gelistet, Drift-Guard grün · Suchzeilen-Suffix
unverändert · Backlog-Footer passt bei 80 Spalten in 2 Zeilen (tmux-Smoke-Beleg) ·
Tree-Footer unverändert · Goldens regeneriert · voller Lauf grün.

---

## Task 7: Vollbild-Modus `v` — Kernmechanik (`bt-13l7`)

Deckt **F01** (Kernmechanik: Zustandsmodell, Einstieg `v`, Listen→Detail-Vollbild via
`enter`, Ausstieg `esc`, Rendering, Maus-Guard — OHNE History-Stack). `blocked_by`
`bt-b0w0` (T4, B04.2 — Relations-Einträge müssen selektierbar sein, bevor ein Sprung im
Vollbild-Detail möglich ist). Vollständige Spezifikation, Zustandsmodell und Code-Sketches:
design-spec.md §15 „F01 — Vollbild-Navigation" + bean `bt-13l7` (self-contained).
Kurzfassung:

Neuer, zu `m.view` ORTHOGONALER `fullscreenMode`-Enum (`fullscreenNone`/`fullscreenList`/
`fullscreenDetail`) + Modelfelder `fullscreen`/`fullscreenBeanID` (`navBack`/`navForward`
bereits deklariert, von T8 gefüllt). `v` (neue `keys.Fullscreen`-Bindung, verifiziert
frei) liest `m.detailFocus`: false → `fullscreenList` (Tree/Backlog-Liste vollbreit),
true → `fullscreenDetail` (Detail-Accordion des fokussierten Beans vollbreit) — kein
Toggle (`v` im bereits aktiven Vollbild ist No-Op, `esc` ist der einzige Ausstieg). `enter`
auf einem Bean im Listen-Vollbild wechselt zu `fullscreenDetail`. `m.focusedBean()`
bekommt einen neuen, vorrangigen Fullscreen-Fall — dadurch funktioniert `keyDetailFocus`
(alle Feld-Kaskaden, PF-5) VERBATIM im Vollbild-Detail, nur `activateDetailField`s
Relations-Sprung-Zweig bekommt einen neuen Fall: bleibt im Vollbild-Detail (zeigt das neue
Ziel-Bean) statt zum Split-Tree zu springen. `esc` verlässt das Vollbild IMMER direkt
zurück zu Browse/Backlog (Split-Modus) — NICHT schrittweise durch die Sprung-Kette (das
ist History-Job, T8) —, fügt sich als weitere Rast in D03s bestehendes „eine Ebene pro
Druck"-Modell ein; beim Verlassen wird der Tree-/Backlog-Cursor auf das zuletzt gezeigte
Bean synchronisiert (PO: „mit dem AKTUELLEN Bean selektiert"). Maus im Vollbild:
dokumentierter Scope-Cut (Guard in `handleMouse`, No-Op statt Fehlklick gegen falsche
Geometrie) — kein PO-Wortlaut verlangt Klick-Support, Fast-Follow außerhalb dieser Runde.

**F01-Kernentscheidungen (für den Supervisor-Check):**
- Zustandsmodell: `fullscreenMode` orthogonal zu `viewID` — KEIN neuer `viewID`-Wert.
- `v`: No-Op wenn bereits Vollbild (Einweg-Einstieg, kein Toggle).
- `enter` im Listen-Vollbild → Detail-Vollbild; Relations-Sprung im Detail-Vollbild bleibt
  im Vollbild (Split-Modus-Sprung bleibt separat unverändert).
- `esc`: IMMER direkter Exit zu Browse/Backlog (nie schrittweise durch die Sprung-Kette),
  mit Cursor-Sync auf das zuletzt gezeigte Bean — fügt sich in D03 als zusätzliche Rast ein.
- Maus im Vollbild: dokumentierter Nicht-Ziel-Punkt (v1), Guard statt Fehlklick.

**Akzeptanz-Checkliste:** `v` öffnet korrekt Listen- bzw. Detail-Vollbild je nach Fokus ·
No-Op im bereits aktiven Vollbild und in der Lobby · `enter` im Listen-Vollbild wechselt zu
Detail-Vollbild · alle Feld-Kaskaden funktionieren unverändert im Vollbild-Detail ·
Relations-Sprung bleibt im Vollbild · `esc`-Kaskade korrekt (Feld→Sektion→Exit bzw. direkt
Exit) mit Cursor-Sync · symmetrisch Tree/Backlog · Maus-Klicks im Vollbild sind No-Op ·
kein neuer `viewID` · tmux-Smoke-Kaskade end-to-end belegt · voller Lauf grün.

---

## Task 8: Vollbild-Detail History-Stack (`bt-1vbp`)

Deckt **F01** (History-Stack-Teil: `ctrl+links`/`ctrl+rechts`, Fallback `[`/`]`,
Footer/Help-Sichtbarkeit). `blocked_by` `bt-13l7` (T7, baut direkt auf dessen
Vollbild-Zustandsmodell + Relations-Sprung-Zweig auf). Vollständige Spezifikation:
design-spec.md §15 „History-Stack" + bean `bt-1vbp` (self-contained). Kurzfassung:

Scope: der Stack trackt AUSSCHLIESSLICH Relations-Sprünge INNERHALB `fullscreenDetail`
(der bestehende Split-Detail-Sprung bleibt unverändert, speist die History nicht — kein
PO-Wortlaut verlangt sie dort). Push bei jedem Sprung (`navBack` += voriges Bean,
`navForward` = nil — Standard-Browser-Semantik). `ctrl+links`/`[` (neu
`keys.HistoryBack`) und `ctrl+rechts`/`]` (neu `keys.HistoryForward`) poppen/tauschen
symmetrisch zwischen `navBack`/`navForward`, wirksam NUR in `fullscreenDetail`.
**Terminal-Verfügbarkeit (PO-Implementierungshinweis, geprüft):** `bubbletea` v1.3.10
dekodiert `ctrl+left`/`ctrl+right` bereits nativ (Standard-xterm-CSI-Sequenzen) — läuft in
den meisten modernen Terminals direkt; bekanntes Risiko (ältere Terminals, tmux ohne
`xterm-keys on`, abweichendes `TERM` in SSH-Ketten) wird durch `[`/`]` als zweite,
terminal-unabhängig garantiert zustellbare Taste je Richtung abgefangen (verifiziert
unbelegt im gesamten Keymap). Sichtbarkeit: neue kontextsensitive
`fullscreenDetailLocalBindings()`-Footer-Ergänzung (mirrort `footer_context.go`s
bestehende Konvention) zeigt `[`/`]` NUR während `fullscreenDetail` aktiv ist, zusätzlich
vollständig in `helpGroups()` dokumentiert.

**Akzeptanz-Checkliste:** Sprung pusht/leert Stacks korrekt · Back/Forward No-Op bei
leerem Stack bzw. außerhalb `fullscreenDetail` · N-Sprünge→N-Back→N-Forward landet exakt
am Ausgangsbean · `[`/`]` funktionieren unabhängig davon, ob `ctrl+arrow` im Terminal
ankommt (Smoke-Beleg beider Pfade einzeln) · History-Keys nur im Detail-Vollbild-Footer
sichtbar, vollständig im Help dokumentiert · `navBack`/`navForward` Copy-on-Write · voller
Lauf grün.

---

## Task 9: Abschluss (`bt-7pk2`)

**blocked_by:** T1, T2, T3, T5, T6, T8 (deckt transitiv T4/T7 ab). Keine Code-Änderungen
erwartet — reine Validierung + Doku + beans-Pflege (Muster `epic-E8-plan.md` Task 9).
Vollständige Spezifikation: bean `bt-7pk2`. Kurzfassung:

1. Voller Regressionslauf (Build, `-race`, `-short` 2×, VOLL 2×, alle Golden-Funktionen
   mit `-count=2`, gofmt/vet leer) — Beleg im bean-Body unter „Voll-Gate-Beleg".
2. Design-Spec-Konsistenz: §15 PF-17 gegen den tatsächlichen Code-Stand nach T1-T8
   gegenprüfen (insbesondere: `hangingIndentWrap` geteilt zwischen T4/T5, `fieldStrip`
   vollständig entfernt, `openBodyEditor` durch `openBeanEditor` ersetzt statt daneben
   behalten, `fullscreenMode` orthogonal zu `viewID`).
3. „Bekannte Grenze" (D01, T2: created_at/updated_at nicht editierbar) im Commit-Body von
   T2 verifizieren — falls dort nicht dokumentiert, hier nachtragen.
4. Epic-Ritual: `bt-tct9` bekommt Tag `to-review` (Agent setzt NIE `completed`); T1-T8 auf
   `completed` verifizieren.
5. `docs/SSTD.md` — Pointer-Update nur falls nötig (prüfen + dokumentieren).
6. Commit `docs(release): E9-Abschluss — Epic to-review, T1-T8-Status, Design-Spec-
   Konsistenz-Beleg`.

**Akzeptanz-Checkliste:** voller Lauf grün · `bt-tct9` trägt `to-review`, nicht
`completed` · T1-T8 alle `completed` (Lücken explizit benannt) · design-spec.md §15 PF-17
verifiziert konsistent · „Bekannte Grenze" dokumentiert · `docs/SSTD.md`-Konsistenz
geprüft · kein unentdeckter Golden-Drift.

---

## Selbst-Review (Plan gegen alle D/B/F-Punkte aus `bt-tct9`)

- **Jeder D/B/F-Punkt genau einem Task zugeordnet:** D01→T2 · B01→T2 (Ist-Beschreibung zu
  D01, kein separater Fix) · B03→T3 · B04→T4 (alle drei Teilpunkte) · B05→T1 (redefinierte
  Fassung) · B06→T5 (Live-Nachtrag) · D02→T6 · F01→T7 (Kernmechanik) + T8
  (History-Stack) · B02→ EXPLIZIT AUSGESCHLOSSEN (Investigator: nicht reproduzierbar,
  PO-Retest angefragt, kein Task-bean). Kein Punkt doppelt vergeben, keiner vergessen.
- **B02 als Nicht-Ziel geführt:** nirgends als Task-bean angelegt, nur als expliziter
  Ausschluss in diesem Plan-Kopf UND in der Task-Übersicht-Präambel dokumentiert — mit
  Reaktivierungs-Bedingung (exakter Reproduktionsablauf).
- **F01-esc-Semantik widerspruchsfrei zum D03-Modell (E8):** `esc` verlässt Vollbild als
  EINE zusätzliche Rast im bestehenden „eine Ebene pro Druck"-Modell (design-spec.md §15
  PF-16 D03) — Feld→Sektion (bestehende Rast, unverändert) → Vollbild-Exit (NEUE Rast,
  T7). Keine pfadabhängige Sonderregel (direkter `v`-Einstieg vs. `enter`-aus-Listen-
  Vollbild vs. N-Relations-Sprünge landen alle auf demselben Exit) — bewusste
  Planner-Entscheidung gegen eine dritte, vom PO nicht verlangte esc-Zielsemantik
  (design-spec.md §15, Abschnitt „Ausstieg (esc)", vollständige Begründung).
- **`blocked_by` korrekt gesetzt (verifiziert via `beans list --parent bt-tct9 --json`):**
  T5←T4 (`bt-4mo9`←`bt-b0w0`), T7←T4 (`bt-13l7`←`bt-b0w0`), T8←T7 (`bt-1vbp`←`bt-13l7`),
  T9←{T1,T2,T3,T5,T6,T8} (`bt-7pk2`←{`bt-mtig`,`bt-z4b1`,`bt-2v38`,`bt-4mo9`,`bt-1e0t`,
  `bt-1vbp`}) — alle sechs Kanten bestätigt, keine fehlt, keine überzählig.
- **Jedes bean self-contained:** jede Task-bean trägt File-Referenzen, Code-Sketches für
  die mechanischen Änderungen, TDD-Schritt-Listen mit konkreten Testnamen, Golden-
  Erwartung, Commit-Vorschlag und eine eigene Akzeptanz-Checkliste — ein frischer Agent
  kann ohne Rückfrage an das Epic-bean `bt-tct9` starten (dessen Body als Detailquelle
  zitiert, nicht dupliziert; design-spec.md §15 PF-17 als Design-Quelle zitiert).
- **Golden-Update-Pflicht erfüllt:** T1/T4/T6 haben je einen expliziten Golden-Regen-
  Schritt mit Vorher/Nachher-Pflicht; T2/T3/T5 haben je einen expliziten Gegenbeleg-
  Schritt; T7/T8 lassen die Wahl (neue Vollbild-Golden-Suite ODER Unit-Test-Abdeckung),
  verlangen aber in JEDEM Fall einen Gegenbeleg für die 3 Basis-Goldens.
- **Kaskaden-end-to-end-Pflicht (LESSONS-LEARNED Eintrag 3) erfüllt:** T7 und T8 tragen
  je einen expliziten tmux-Smoke-Schritt, der die VOLLE Kaskade (v→enter→Sprünge→esc bzw.
  Sprünge→Back/Forward) aus REALISTISCHEM Ausgangszustand (Tree UND Backlog) durchspielt,
  nicht nur die jeweils neu gebaute Stufe isoliert.
- **Grenzbreiten-Smoke-Pflicht (LESSONS-LEARNED Eintrag 4) erfüllt:** T5 und T6 tragen je
  einen expliziten 80-Spalten-tmux-Smoke-Schritt.
- **Deviation vom Richtwert „5-7 Tasks" dokumentiert:** dieser Plan schneidet 8
  Implementierungs-Tasks + Abschluss (9 gesamt) — zwei Punkte über dem oberen Richtwert-
  Ende. Begründung: F01 wurde (PO-explizit erlaubt: „ggf. in 2 Tasks splitten") in Kern +
  History getrennt (unterschiedliche Fertigstellungs-/Risiko-Profile, T8 ist optional
  klein genug für einen eigenen Review-Zyklus); B06 kam als Live-Nachtrag WÄHREND dieser
  Planungssession hinzu und wurde bewusst NICHT in T4 hineingezogen (disjunkte
  Datei-Scopes, kleinerer Diff je Task) — beide Splits sind dokumentierte,
  Review-getriebene Entscheidungen, kein unreflektiertes Aufblähen.
