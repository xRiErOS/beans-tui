# Epos E10 — Tag-Management-Page (zentrale Tag-Definition)

Liefert das Feature aus `bt-6oyy` (PO-Wunsch 2026-07-15, im Zuge E7-Feedback;
Scope/Reihenfolge revidiert 2026-07-16, PO-Entscheid: Kette startet direkt nach
E9/`bt-tct9`, eigenes Epos, eigener Planner): eine im Command-Center erreichbare
Page zur zentralen Tag-Verwaltung (definieren, umbenennen mit Propagation über
alle Beans, entfernen) plus Suggest-Mode-Integration in den bestehenden
Tag-Picker (`t`, `box_picker_tag.go`) — definierte Tags werden dort priorisiert
angeboten, freie Tags bleiben weiterhin uneingeschränkt erlaubt (kein strict
mode, PO-Vorgabe wörtlich). Persistenz repo-lokal (exakter Dateiname/Format war
PO-seitig ausdrücklich dem Planner überlassen — s. D01 unten).

Details/Herleitung/volle Design-Entscheidungen (D01-D15) + offene Fragen
(Q01-Q03): Epic-bean `bt-362n` (vollständige Auftragsquelle, DRY — hier nicht
dupliziert). Quellen: Feature-bean `bt-6oyy` (PO-Entscheide 2026-07-16) ·
design-spec.md §4 (Entity-Mapping „Tags"), §15 PF-12 (reservierte
Marker-Spalte, Layout-Shift-Verbot), §15 PF-16 B14/bean `bt-ntoz` (v1-
Minimal-Lösung „create tag", die diese Page ablöst/ergänzt) · `epic-E9-plan.md`
(Muster für Aufbau/Detailtiefe, Golden-/tmux-Konventionen) · `docs/
LESSONS-LEARNED.md` (alle Einträge gelesen; besonders E8 #7 „Planner
verifiziert jeden Fund vor Task-Schnitt gegen Ist-Code" — hier: Persistenzort/
CLI-Verben/Verwendungszähler-Quelle empirisch verifiziert, s. Epic-Body
„Empirische Verifikation"; E9 #5 „jeder neue UI-Modus braucht eine
Exit-Pfad-Inventur" — hier direkt Grundlage von D06) · Referenz-Klon
`~/Obsidian/tools/lean-stack/beans-src` (`pkg/beancore/core.go` Scan-Verhalten)
· Code-Stand 2026-07-16 (HEAD `096a578`).

## Task-Übersicht

| Task | bean | Inhalt | Codes | blocked_by |
|---|---|---|---|---|
| T1 | `bt-49hh` | Persistenzschicht `internal/data/tagdefs.go` | D01-D04 | — |
| T2 | `bt-r92i` | Page-Grundgerüst `viewTagManagement` (read-only Liste) | D05-D09 | T1 |
| T3 | `bt-604w` | Tag-Definition anlegen (Create + geteilter Input-Submodus) | D11, D14 | T2 |
| T4 | `bt-1lsu` | Tag-Definition entfernen (Delete, registry-only) | D12, D15 | T2 |
| T5 | `bt-y9my` | Tag-Definition umbenennen + Propagation über alle Beans | D13, D14 | T3 |
| T6 | `bt-pqq3` | Tag-Picker Suggest-Mode (`collectTagCounts` erweitert) | D10 | T1 |
| T7 | `bt-sohl` | Abschluss (Voll-Validierung, design-spec.md §16, Epic to-review) | — | T2, T3, T4, T5, T6 |

**Reihenfolge-Begründung:** T1 ist das einzige echte Fundament (neue Datei,
kein bestehender Code berührt) — ALLES andere hängt direkt oder transitiv
daran. T2 (Page-Grundgerüst: neuer `viewTagManagement`, Palette-Eintrag,
`handleKey`-Dispatch, read-only Liste) braucht `LoadTagDefs`/`SaveTagDefs`
aus T1, ist aber sonst unabhängig von jeder Mutation. T3 (Create) und T4
(Delete) sind BEIDE nur von T2 abhängig und laufen PARALLEL — disjunkte
Funktionen im selben neuen File `view_tag_management.go` (mirrort E9s
T1/T4-Präzedenzfall „disjunkte Funktionen, keine Abhängigkeit trotz
derselben Datei"): T3 führt den page-lokalen Freitext-Input-Submodus ein
(`tagMgmtInputActive`/`tagMgmtInputMode`/`tagMgmtInput`), T4 einen
unabhängigen Delete-Confirm-Bool (`tagMgmtDeleteConfirm`/
`tagMgmtDeleteTarget`) — kein geteilter State zwischen beiden. T5 (Rename)
`blocked_by` T3, NICHT nur T2, weil Rename denselben Freitext-Input-Submodus
WIEDERVERWENDET (D14, vorbefüllt mit dem alten Namen) statt einen zweiten,
parallelen Eingabemechanismus zu erfinden — mirrort E9s T5←T4-Abhängigkeit
(geteilter `hangingIndentWrap`-Helfer, hier: geteilter Input-Submodus). T6
(Tag-Picker Suggest-Mode) hängt NUR an T1 (disjunkter Datei-Scope,
`box_picker_tag.go` statt `view_tag_management.go`) und kann daher komplett
PARALLEL zu T2-T5 laufen — maximale Parallelität, mirrort E9s Fan-out-
Philosophie „viele Tasks hängen nur an der frühesten fundamentalen Task,
nicht an jeder vorherigen in einer langen Kette". T7 (Abschluss)
`blocked_by` die fünf „Blatt"-Tasks T2/T3/T4/T5/T6 — deckt T1 transitiv ab,
da T2 UND T6 beide direkt daran hängen (identisches Muster zu
`epic-E9-plan.md`s T9).

**Golden-Strategie** (mirrort E9s Konvention): KEINER dieser sieben Tasks
ändert die bestehenden Tree-/Backlog-/Chrome-Basis-Goldens — die neue Page
ist ein komplett NEUER Render-Pfad (kein bestehender wird umgebaut), und T6
ändert nur Sortierung/eine Marker-Spalte INNERHALB eines Overlays (Overlays
sind laut design-spec.md §11 ohnehin nicht Teil der 3 Basis-Goldens). Jeder
Task trägt daher einen GEGENBELEG-Schritt (nach `command go build -o bin/bt
.`: `command go test ./internal/tui/ -run
"TestTreeGolden|TestBacklogGolden|TestChromeGolden"` OHNE `-update`, MUSS
grün bleiben, „unverändert" ist im Commit-Body explizit zu benennen — auch
das ist eine gültige, zu belegende Aussage). Ob `viewTagManagement()` eine
EIGENE Golden-Suite bekommt (View()-Snapshots des neuen Panes), ist
Implementer-Entscheidung ab T2 (mirrort E9 T7/T8s Freiheit bei einem
komplett neuen Render-Pfad) — mindestens ein paar gezielte
`View()`-Assertions sind sinnvoll, da dies der einzige neue Render-Pfad der
ganzen Kette ist.

**tmux-Smoke-Pflicht (Breite 120 UND 80, jeder Task):** anders als E9 (wo
nur breiten-ändernde Tasks einen expliziten 80-Spalten-Smoke trugen) verlangt
DIESER Plan-Auftrag für JEDEN Task beide Breiten — jede Task-bean trägt einen
eigenen 120+80-Smoke-Schritt (s. jeweilige Task-bean „tmux-Smoke"-Abschnitt).
T5 (Rename-Propagation) mutiert echte Beans und läuft daher bewusst gegen ein
SEPARATES Scratch-Repo, nicht gegen dieses Dogfooding-Repo (mirrort
`beans-src/CLAUDE.md`s eigene „Mutation → separates Test-Projekt"-Regel).

**Temporäre Testdateien:** T3/T4/T6 legen im Zuge ihres tmux-Smokes
vorübergehend eine `.beans-tags.yml` in DIESEM Repo an (Live-Verifikation
gegen die eigene Dogfooding-Instanz) — jede Task-bean verlangt explizit, sie
danach restlos zu entfernen (`git status --porcelain` leer am Task-Ende).
T5s Scratch-Repo liegt komplett AUSSERHALB dieses Repos und wird nach dem
Smoke gelöscht.

---

## Task 1: Persistenzschicht `internal/data/tagdefs.go` (`bt-49hh`)

Deckt **D01-D04** (Persistenzort/-format, Lade-Semantik, Reload/Liveness,
Paket-Zuordnung). Unabhängig (kein `blocked_by`) — reiner Neu-Datei-Scope.
Vollständige Spezifikation, Code-Sketches und TDD-Schritte: bean `bt-49hh`
(self-contained). Kurzfassung:

Neue Datei `.beans-tags.yml` im Repo-Root (Sibling zu `.beans.yml`, über
`client.RepoDir` — keine neue Discovery-Logik nötig), NICHT in `.beans/`
(empirisch verifiziert: beans toleriert Fremddateien dort zwar klaglos —
Live-Probe UND Referenz-Klon-Code, `pkg/beancore/core.go:128-179` —, aber
das Repo-Root ist der von beans' eigener Autorität komplett entkoppelte,
sauberere Ort, mirrort `.beans.yml` selbst). Format: flache, sortierte
`tags: [...]`-Liste, kein Farb-/Beschreibungsfeld (YAGNI). Neue Methoden auf
`*data.Client`: `LoadTagDefs()`/`SaveTagDefs([]string)` — tolerant-missing
wie `config.LoadSettings()` (fehlend/korrupt → leere Registry, nie ein
Fehler nach außen), plus reine Helfer `AddTagDefName`/`RemoveTagDefName`/
`RenameTagDefName`.

**Akzeptanz-Checkliste:** Dateiname/-format wie D01 · Load tolerant-missing
UND tolerant-korrupt · Save sortiert+dedupliziert · reine Helfer ohne I/O ·
alle RED-Tests zuerst rot · voller Lauf grün.

---

## Task 2: Page-Grundgerüst `viewTagManagement` (read-only) (`bt-r92i`)

Deckt **D05-D09** (Einstiegspunkt, Dispatch-/Capture-Modell, Chrome, Layout,
Zeilen-Menge & Sortierung). `blocked_by` `bt-49hh` (T1). Vollständige
Spezifikation, Code-Sketches und TDD-Schritte: bean `bt-r92i`
(self-contained). Kurzfassung:

Neuer `viewTagManagement` (viewID), erreichbar AUSSCHLIESSLICH über das
Command-Center („go to tags", mirrort den „go to settings"-Präzedenzfall —
auch das Settings-Form hat keine eigene Taste; Tastenraum bleibt knapp).
**Kritische Architekturentscheidung (D06):** die Page ist ein
FULL-CAPTURE-Zustand, dispatcht an DERSELBEN `handleKey`-Stelle wie
`viewLobby` (früh, VOR ctrl+k/?/p) — NICHT über die generische
`keyNodeAction`/Tree/Backlog-Kette. Grund (verifiziert gegen `update.go:
1019-1029`): `focusedBean()`s `default`-Zweig fällt ohne eigenen `viewID`-
Case auf den TREE-Cursor zurück — liefe die Page nicht full-capture, würden
globale Node-Action-Tasten (s/t/a/r/c/d/e) still gegen ein STALE,
unzusammenhängendes Bean feuern, während der PO die Tag-Registry ansieht
(exakt die Fehlerklasse aus LESSONS-LEARNED E9 #5 „Exit-Pfad-Inventur").
Chrome reuse (Backlog-Stil Einzel-Pane, nicht Lobbys Bespoke-Zentrierung),
`GlobalHint` bewusst LEER (D07 — die 4 globalen Tasten funktionieren
während Full-Capture nicht, sie anzuzeigen wäre irreführend). Layout:
Einzel-Pane-Liste, KEIN Master-Detail-Drilldown (D08, dokumentierter,
günstig nachrüstbarer YAGNI-Cut — `idx.WithTag` existiert bereits). Zeilen:
Union aus definierten Tags (Registry, auch Count 0) und freien
in-Verwendung-Tags — zwei Gruppen, Definiert (alpha) zuerst, dann
Undefiniert (Count absteigend, D09).

**Akzeptanz-Checkliste:** Page über Palette erreichbar · Union-Liste
korrekt sortiert · Verwendungszähler korrekt · `enter` dokumentierter No-Op
· `esc` → Browse · Regressionstest belegt KEIN Leak globaler Node-Action-
Tasten in diese View (D06) · GlobalHint leer · Goldens Gegenbeleg grün ·
tmux-Smoke 120+80 · voller Lauf grün.

---

## Task 3: Tag-Definition anlegen (Create) (`bt-604w`)

Deckt **D11** (Create-Semantik) + **D14** (geteilter Freitext-Input-
Submodus, hier EINGEFÜHRT). `blocked_by` `bt-r92i` (T2). Vollständige
Spezifikation, Code-Sketches und TDD-Schritte: bean `bt-604w`
(self-contained). Kurzfassung:

`n` (reuse `keys.NewTag`, gleiche Bedeutung wie im Tag-Picker, D14) öffnet
einen page-lokalen Freitext-Input (`tagMgmtInputActive`/`tagMgmtInputMode`/
`tagMgmtInput`, mirrort `box_picker_tag.go`s eigenen `tagInputActive`-
Präzedenzfall 1:1). Validierung gegen `data.ValidTagName` + Dedupe gegen die
Registry; bei Erfolg `SaveTagDefs` (async `saveTagDefsCmd`/
`tagDefsSavedMsg`, neuer, aber minimaler Cmd-Typ — Konsistenz mit jedem
anderen state-ändernden Aufruf, auch wenn der eigentliche I/O synchron
schnell wäre). Eine neue Definition berührt KEIN Bean (D11) — reiner
Registry-Akt.

**Akzeptanz-Checkliste:** `n` öffnet Input · gültiger Name legt Definition
an, Liste aktualisiert sofort · ungültiger/doppelter Name → Inline-Fehler,
kein Submit · Create berührt kein Bean (Regressionstest: `m.idx`
unverändert) · `esc` verwirft nur den Submodus · Goldens Gegenbeleg grün ·
tmux-Smoke 120+80, Testdatei danach entfernt · voller Lauf grün.

---

## Task 4: Tag-Definition entfernen (Delete) (`bt-1lsu`)

Deckt **D12** (Delete-Semantik) + **D15** (page-lokaler Confirm, kein
`overlayID`-Case). `blocked_by` `bt-r92i` (T2), PARALLEL zu T3 (disjunkte
Funktionen). Vollständige Spezifikation, Code-Sketches und TDD-Schritte:
bean `bt-1lsu` (self-contained). Kurzfassung:

`d` (reuse `keys.Delete`) öffnet einen Confirm mit LIVE-Verwendungszähler
(mirrort `box_confirm_delete.go`s Kinder-/Links-Zähler-Warnung) über ein
page-lokales `tagMgmtDeleteConfirm`/`tagMgmtDeleteTarget`-Paar (D15, mirrort
`m.confirmQuit`s eigenen „bewusst nicht ins overlayID-Enum"-Präzedenzfall).
**REGISTRY-ONLY** (D12): Beans, die den Tag tragen, behalten ihn — er wird
wieder ein freier Tag. Kein `SetTags`/Bean-Mutation-Call in diesem Pfad, nur
`SaveTagDefs` mit dem entfernten Namen. `d` auf einer freien (undefinierten)
Zeile ist ein No-Op (nichts zu löschen).

**Akzeptanz-Checkliste:** `d` auf definierter Zeile öffnet Confirm mit
korrektem Live-Count · `d` auf freier Zeile No-Op · `enter` entfernt NUR die
Definition, Beans behalten den Tag (Regressionstest) · `esc`/`n` bricht ohne
Save ab · Goldens Gegenbeleg grün · tmux-Smoke 120+80, Testdatei danach
entfernt · voller Lauf grün.

---

## Task 5: Tag umbenennen + Propagation über alle Beans (`bt-y9my`)

Deckt **D13** (Rename-Semantik & Propagation) + reused **D14** (Input-
Submodus). `blocked_by` `bt-604w` (T3, teilt dessen Input-Submodus).
Größtes Einzelstück dieser Runde. Vollständige Spezifikation, Code-Sketches
und TDD-Schritte: bean `bt-y9my` (self-contained). Kurzfassung:

`e` (NEUES `keys.RenameTag`-Binding, gebunden auf `"e"` — gleiche Taste wie
das GLOBALE `keys.Editor`, aber disjunkter, sich gegenseitig ausschließender
Anzeigekontext: `viewTagManagement` ist full-capture, D06, `keyNodeAction`s
`e`-Case wird während dieser View nie erreicht — mirrort den bereits
etablierten Backlog-`S`-Präzedenzfall „gleiche Taste, andere View-lokale
Bedeutung") öffnet den T3-Input-Submodus, vorbefüllt mit dem alten Namen
(Dedupe-Check lässt den eigenen alten Namen durch). Bei Submit: **Registry-
Rename ZUERST** (reiner lokaler Datei-Op, praktisch nie fehlschlagend,
D13), UNABHÄNGIG vom Bean-Sweep. **Bean-Sweep:** neuer `renameTagCmd`
iteriert `idx.WithTag(alt)` und feuert je EINEN kombinierten
`data.Client.SetTags(id, add=[neu], remove=[alt], etag)`-Call PRO Bean (die
BESTEHENDE Methode, kein neuer Client-Call nötig) — **CONTINUE-ON-ERROR**
(kein Abbruch bei einem stale ETag, da beans KEINE Cross-Bean-Transaktion
kennt, verifiziert gegen `beans update --help`: kein Bulk-/Rename-Verb).
Ergebnis (renamed/failed-Zahlen, erste Fehlermeldung) als EIN Toast (neuer
`tagRenameDoneMsg`-Typ — die zweite bewusste Ausnahme vom geteilten
`mutationDoneMsg`-Tail nach `createDoneMsg`); `m.idx` lädt danach IMMER neu
(mirrort `applyMutationResult`s „always reload after").

**Akzeptanz-Checkliste:** `e` auf definierter Zeile öffnet vorbefüllten
Rename-Input · `e` auf freier Zeile No-Op · Registry benennt sofort um ·
Sweep läuft asynchron, continue-on-error (Regressionstest mit simuliertem
Teilfehlschlag, Test belegt dass NACHFOLGENDE Beans trotz eines
Fehlschlags weiter versucht werden) · Toast zeigt renamed/failed · `m.idx`
lädt danach neu · `RenameTag` in `helpGroups()`, Drift-Guard grün · Goldens
Gegenbeleg grün · tmux-Smoke 120+80 in SEPARATEM Scratch-Repo, Scratch
danach gelöscht · voller Lauf grün.

---

## Task 6: Tag-Picker Suggest-Mode (`bt-pqq3`)

Deckt **D10** (Suggest-Mode-Sortierung + PF-12-konforme Marker-Spalte).
`blocked_by` `bt-49hh` (T1) NUR — disjunkter Datei-Scope zu T2-T5, läuft
komplett parallel dazu. Vollständige Spezifikation, Code-Sketches und
TDD-Schritte: bean `bt-pqq3` (self-contained). Kurzfassung:

`collectTagCounts` (`box_picker_tag.go`) bekommt einen `defined
map[string]bool`-Parameter, Sortierung wächst um „defined" als NEUEN
PRIMÄREN Schlüssel (definierte Tags zuerst), der bestehende
Count-absteigend/alpha-Tie-Break bleibt sekundär je Gruppe. `openTagPicker`
lädt die Registry frisch (D03) vor dem Zählen. Visuelle Unterscheidung: eine
IMMER reservierte Marker-Spalte (PF-12-Konvention, design-spec.md §15 —
„kein bedingtes Präfix, das das Layout verschiebt", dieselbe Regel, die
`accordion.go`/`fieldStrip` bereits zweimal fixen musste) — definierte Tags
ein gefülltes Glyph, freie Tags gleich breiter Leerraum. Freie Tags bleiben
uneingeschränkt wählbar (kein strict mode, Regressionstest Pflicht). Der
BESTEHENDE Freitext-Neu-Tag-Pfad (`n`, B14-Erbe) bleibt UNVERÄNDERT — er
schreibt weiterhin NICHT in die Registry (Q03 im Epic-Body, bewusst nicht
Teil dieses Tasks).

**Akzeptanz-Checkliste:** `collectTagCounts` sortiert defined-first ·
Marker-Spalte IMMER reserviert (PF-12-Test grün, Layout-Breiten-Vergleich
zwischen defined/undefined-Zeile) · Registry frisch bei jedem Open · freie
Tags weiterhin uneingeschränkt wählbar · bestehender Diff-/Mutation-
Test-Korpus unverändert grün · Goldens Gegenbeleg grün · tmux-Smoke 120+80,
Testdatei danach entfernt · voller Lauf grün.

---

## Task 7: Abschluss (`bt-sohl`)

**blocked_by:** T2, T3, T4, T5, T6 (deckt T1 transitiv ab). Keine
Feature-Code-Änderungen erwartet — reine Validierung + Doku + beans-Pflege
(Muster `epic-E9-plan.md` Task 9). Vollständige Spezifikation: bean
`bt-sohl`. Kurzfassung:

1. Voller Regressionslauf (Build, `-race`, `-short` 2×, VOLL 2×, alle
   Golden-Funktionen inkl. jeder in T2/T6 evtl. neu angelegten Suite mit
   `-count=2`, gofmt/vet leer) — Beleg im bean-Body unter „Voll-Gate-Beleg".
2. Design-Spec-Konsistenz: neuer Abschnitt „§16 — Tag-Management-Page"
   (design-spec.md) dokumentiert D01-D15 in Spec-Form; §4s Entity-Mapping-
   Zeile „Tag-Manager-CRUD entfällt" UND §9s „OUT (v1): Tag-Manager-CRUD"
   werden als SUPERSEDED markiert (mirrort PF-14s Umgang mit der entfernten
   Review-Cockpit-Zeile — nicht stillschweigend löschen).
3. Q01-Q03 (Epic-Body `bt-362n`) im Report explizit an den PO weiterreichen
   — nicht blockierend, Epic-Review-Ritual.
4. `bt-6oyy`-Body-Verweis auf das Epic verifizieren (vom Planner bereits
   gesetzt, hier nur Konsistenz-Check).
5. Epic-Ritual: `bt-362n` bekommt Tag `to-review` (Agent setzt NIE
   `completed`); T1-T6 auf `completed` verifizieren.
6. `docs/SSTD.md` — Pointer-Update nur falls nötig (prüfen + dokumentieren).
7. Commit `docs(release): E10-Abschluss — Epic to-review, T1-T6-Status,
   Design-Spec §16`.

**Akzeptanz-Checkliste:** voller Lauf grün · `bt-362n` trägt `to-review`,
nicht `completed` · T1-T6 alle `completed` (Lücken explizit benannt) ·
design-spec.md §16 neu + §4/§9 als superseded markiert · Q01-Q03 an den PO
weitergereicht · `bt-6oyy`-Konsistenz geprüft · `docs/SSTD.md`-Konsistenz
geprüft · kein unentdeckter Golden-Drift · `git status --porcelain` leer.

---

## Selbst-Review (Plan gegen alle D/Q-Punkte aus dem Epic-Body `bt-362n`)

- **Jeder D-Punkt genau einem Task zugeordnet:** D01-D04→T1 (Persistenz) ·
  D05-D09→T2 (Page-Grundgerüst) · D11+D14(Einführung)→T3 (Create) ·
  D12+D15→T4 (Delete) · D13+D14(Reuse)→T5 (Rename) · D10→T6 (Picker
  Suggest-Mode). Kein Punkt doppelt vergeben, keiner vergessen — 15
  D-Punkte, 6 Umsetzungs-Tasks, jeder Task deckt einen klar disjunkten
  Teil ab.
- **Q01-Q03 als Nicht-Ziele geführt:** nirgends als eigener Task-Cut
  gebaut — Q01 (destruktiver Delete-Modus) bleibt bei D12s Registry-only-
  Default, Q02 (Stale-Definition-Marker) bleibt bei D09s schlichter
  Count-0-Anzeige, Q03 (B14-Picker-„n" soll auch registrieren) bleibt bei
  T6s expliziter „B14 unverändert"-Klausel — alle drei explizit im
  Epic-Body benannt UND im Abschluss-Task (T7 Schritt 3) an den PO
  weitergereicht, keine stille Annahme.
- **D06-Sicherheitsnetz widerspruchsfrei zum bestehenden Full-Capture-
  Modell (Lobby-Präzedenzfall):** die Page dispatcht an EXAKT derselben
  `handleKey`-Stelle wie `viewLobby` (früh, vor ctrl+k/?/p) — keine neue,
  dritte Capture-Rangordnung erfunden, nur der bestehende Lobby-Rang ein
  zweites Mal besetzt. T2 trägt einen expliziten Regressionstest gegen das
  konkrete Leck-Risiko (D06s eigene Rationale: `d` während der Tag-Page
  darf NIE den Bean-Delete-Confirm öffnen).
- **`blocked_by` korrekt gesetzt (verifiziert via `beans list --parent
  bt-362n --json`):** T2←T1 (`bt-r92i`←`bt-49hh`), T3←T2 (`bt-604w`←
  `bt-r92i`), T4←T2 (`bt-1lsu`←`bt-r92i`), T5←T3 (`bt-y9my`←`bt-604w`),
  T6←T1 (`bt-pqq3`←`bt-49hh`), T7←{T2,T3,T4,T5,T6} (`bt-sohl`←{`bt-r92i`,
  `bt-604w`,`bt-1lsu`,`bt-y9my`,`bt-pqq3`}) — alle sechs Kanten bestätigt,
  keine fehlt, keine überzählig.
- **Jedes bean self-contained:** jede Task-bean trägt File-Referenzen,
  Code-Sketches, TDD-Schritte mit konkreten Testnamen (RED-Zitate),
  Golden-Erwartung, tmux-Smoke-Vorgabe (120+80), Commit-Vorschlag und eine
  eigene Akzeptanz-Checkliste — ein frischer Agent kann ohne Rückfrage an
  das Epic-bean `bt-362n` starten (dessen Body als Detailquelle zitiert,
  nicht dupliziert).
- **Golden-Update-Pflicht erfüllt:** alle sechs Umsetzungs-Tasks tragen
  einen expliziten Gegenbeleg-Schritt für die 3 Basis-Goldens (keiner
  davon regeneriert sie — konsistent mit dem Befund, dass kein bestehender
  Render-Pfad umgebaut wird); T2/T6 lassen eine EIGENE neue Golden-Suite
  als Implementer-Wahl offen.
- **tmux-Smoke-Pflicht (120+80) erfüllt:** jede der sieben Task-beans
  trägt einen eigenen, expliziten 120+80-Smoke-Schritt (Plan-Auftrag
  verlangt dies für JEDEN Task, nicht nur breiten-ändernde — anders als
  E9s schmalere CLAUDE.md-Pflicht). T5 grenzt sich bewusst ab (separates
  Scratch-Repo statt Dogfooding-Repo, da es echte Bean-Mutationen
  auslöst).
- **Temporäre Test-Artefakte:** T3/T4/T6 (Live-`.beans-tags.yml` in diesem
  Repo) und T5 (Scratch-Repo) tragen je eine explizite
  Restlos-Entfernen-Pflicht in ihrer Akzeptanz-Checkliste; T7 verifiziert
  `git status --porcelain` am Ende leer als eigenen Prüfpunkt.
- **Deviation vom Richtwert „5-7 Tasks" dokumentiert:** dieser Plan
  schneidet 6 Implementierungs-Tasks + Abschluss (7 gesamt) — innerhalb
  des Richtwerts, am unteren Ende (E9 lag bei 8+1). Begründung: die
  Persistenz- (T1) und Picker-Suggest-Mode-Arbeit (T6) sind disjunkt genug
  von der eigentlichen Page (T2-T5), um als eigene, klar geschnittene
  Tasks zu stehen, aber die Page selbst braucht keinen künstlichen
  Zusatz-Split über Create/Delete/Rename hinaus (drei CRUD-Operationen,
  drei Tasks — kein viertes Stück übrig, das einen eigenen Task
  rechtfertigen würde).
