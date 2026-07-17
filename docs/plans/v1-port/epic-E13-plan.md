# Epos E13 — Nacharbeitsrunde (Grilling 2026-07-17, 5 beans unter bt-5uzr)

Liefert die dritte Nacharbeitsrunde: fünf beans, alle mit demselben Parent
(`bt-5uzr`), alle mit finalen PO-Entscheidungen aus dem Grilling-Termin
2026-07-17 (Sektionen „PO-Entscheid Grilling"/„PO-Redefinition Grilling" je
bean, verbatim übernommen — keine davon wird hier neu aufgemacht). KEIN neues
Parent-Epic — dieser Plan bündelt nur die Ausführungsreihenfolge/
Parallelisierung. Zwei Items (`bt-0xrb`/`bt-tm4a`) bilden eine Toast-Familie
und werden sequenziell bearbeitet; drei Items (`bt-2kfl`/`bt-d3ps`/`bt-nxuk`)
berühren disjunkte Datei-Familien und sind worktree-parallelisierbar.

Quellen: `beans show bt-0xrb`/`bt-tm4a`/`bt-2kfl`/`bt-d3ps`/`bt-nxuk`
(vollständige Bodies inkl. Grilling-Sektionen, 2026-07-17) ·
`docs/plans/v1-port/epic-E12-plan.md` (Stilvorlage) · Ist-Code-Stand
2026-07-17 nach E12-Merge (git log: `b7a1ca6` bt-2p9m Querformat-Merge zuletzt
relevant für Item 3/5) · `docs/LESSONS-LEARNED.md` (NBSP-Wordwrap-Falle,
Modal-Hint-Zeilen).

## Item-Übersicht

| # | bean | Typ | Inhalt | Datei-Familie | blocked_by |
|---|---|---|---|---|---|
| 1 | `bt-0xrb` | Bug | Toast wächst dynamisch (alle Severities) bis Meldung vollständig lesbar; plain-ErrConflict-Zweig befüllt `toastCtx` | `overlay_show_toast.go` (`toastBox`/`toastGeometry`/`toastBoxWidth`), `update.go` (`applyMutationResult`, Zeile 275-327) | — |
| 2 | `bt-tm4a` | Task | `createInFlightNote`-Toast-Severity vereinheitlichen auf `toastWarn` (3 Stellen) | `update.go:735` (bleibt), `box_confirm_create.go:56`, `overlay_palette.go:240` + zugehörige Tests | empfohlen NACH #1 (Datei-Nachbarschaft `update.go`, siehe Reihenfolge-Begründung) |
| 3 | `bt-2kfl` | Feature | Such-Präfixe `st:`/`ty:`/`pr:`/`tag:` — additiver Layer (D02), Parser je Tastendruck (D03) | `view_browse_repo.go` (`beanMatchesSearch`, `update.go`s `keySearchInput`), `box_filter_facets.go` (`filterSummary`, lesend+Union-Erweiterung) | — |
| 4 | `bt-d3ps` | Feature | Lobby zeigt `slug — Pfad`; Palette-Befehl „register project" | `view_lobby.go` (`repoPickerBody`), `overlay_palette.go` (`paletteActions`/`dispatchPalette`, NEUER Case), `internal/config` (`SaveUserSettings`, gelesen), NEU `internal/data` (`RepoSlug`) | — |
| 5 | `bt-nxuk` | Task | Footer-Hint bei offenem Filter-Menü zeigt „tab/shift+tab category" statt globalem FocusIn/FocusOut-Label | `footer_context.go` (`filterMenuLocalBindings`) | — |

**Parallelisierungs-Welle (Worktree):** #3 (`bt-2kfl`), #4 (`bt-d3ps`), #5
(`bt-nxuk`) berühren disjunkte Datei-Familien — kein gemeinsames File
zwischen den dreien (`view_browse_repo.go`/`box_filter_facets.go` vs.
`view_lobby.go`/`overlay_palette.go`/`internal/config` vs.
`footer_context.go`) — parallel in drei Worktrees bearbeitbar. #1
(`bt-0xrb`) und #2 (`bt-tm4a`) bilden die Toast-Familie und werden
SEQUENZIELL in EINEM Worktree/EINER Session bearbeitet (Reihenfolge:
`bt-0xrb` → `bt-tm4a`).

## Reihenfolge-Begründung

- **Toast-Familie (#1→#2) zuerst, sequenziell, VOR der Parallel-Welle —
  Planner-Empfehlung, kein hartes PO-Datum, aber sinnvoll aus
  Datei-Nachbarschafts-Gründen:** `bt-0xrb` ändert `overlay_show_toast.go`s
  Kern-Rendering (`toastBox`/`toastGeometry`, Breiten-/Wrap-Logik für ALLE
  Severities) UND `update.go`s Conflict-Branch (Zeile 275-327). `bt-tm4a`
  ändert `update.go:735`/`box_confirm_create.go:56`/`overlay_palette.go:240`
  — DIESELBEN Dateien wie `bt-0xrb` (wenn auch andere Zeilen/Funktionen:
  `applyMutationResult`s Conflict-Zweig vs. die drei
  `createInFlightNote`-Guards) UND ruft `showToast(...)` auf, dessen
  Signatur/Rendering `bt-0xrb` gerade umbaut. `bt-0xrb` zuerst hält
  `bt-tm4a`s Severity-Fix gegen den bereits fertigen Toast-Rendering-Stand,
  statt gegen einen, der währenddessen nochmal verändert wird (mirrort
  `epic-E12-plan.md`s Item-1→2-Begründung, „dieselbe Datei-Nachbarschaft").
  Zusätzlicher Datei-Nachbar außerhalb der Toast-Familie: `bt-d3ps` (#4)
  fügt EBENFALLS einen neuen `dispatchPalette`-Case in `overlay_palette.go`
  hinzu (andere Funktion/anderer Case als `bt-tm4a`s
  `createInFlightNote`-Guard, aber dieselbe Datei) — kein hartes
  `blocked_by`, aber die Toast-Familie VOR der Parallel-Welle abzuschließen
  hält `overlay_palette.go` für `bt-d3ps` auf einem bereits stabilen Stand
  (kein Merge-Zittern zwischen zwei Worktree-Branches, die dieselbe Datei
  anfassen).
- **#3/#4/#5 parallel, keine Reihenfolge untereinander vorgegeben (PO nannte
  keine, Datei-Familien beweisbar disjunkt):** `bt-2kfl` ist
  `view_browse_repo.go`/`box_filter_facets.go`-lastig (Suchpfad), `bt-d3ps`
  ist `view_lobby.go`/`overlay_palette.go`/`internal/config`-lastig
  (Lobby+Register-Befehl), `bt-nxuk` ist `footer_context.go`-isoliert (eine
  einzelne Funktion, `filterMenuLocalBindings`). `bt-2kfl` und `bt-nxuk`
  berühren beide `box_filter_facets.go`-NAHE Konzepte (Filter-Menü), aber
  `bt-2kfl` liest `box_filter_facets.go` nur (Union-Anzeige in
  `filterSummary`, KEIN Schreibzugriff auf `m.filterStatus`/etc. — D02),
  `bt-nxuk` schreibt ausschließlich `footer_context.go` — kein
  Datei-Overlap. `bt-d3ps`s neuer `overlay_palette.go`-Case liegt in einer
  ANDEREN Funktion/einem anderen `switch`-Case als jeder Toast-Familie-Case
  — sauber parallel zu #1/#2, solange die Toast-Familie zuerst LANDET (siehe
  oben).

## Item 1: Toast wächst dynamisch + toastCtx im plain-ErrConflict-Zweig (`bt-0xrb`)

**Scope-Bindung (D04, PO-Entscheid Grilling 2026-07-17, bean-Sektion
„PO-Entscheid Grilling 2026-07-17"):** final NUR B01a — keine
Konflikt-Sonderfall-Mitigation in beans-tui, kein Daten-Heilungs-Schritt
(Diagnose-Item 2/4, außerhalb dieses beans), kein Upstream-Warten
(Diagnose-Item 1, außerhalb dieses beans). Die Diagnose-Sektion („Diagnose-
Ergebnis (ce-diagnose, 2026-07-17)" im bean) bleibt Kontext, NICHT
nochmals aufgemacht.

**Root Cause.**

- `overlay_show_toast.go:146`: `const toastBoxWidth = 36` — fixe Breite,
  keine Content-Anpassung.
- `overlay_show_toast.go:156`: `title := ansi.Truncate(t.title, innerW-2,
  "…")` mit `innerW := toastBoxWidth - 4` (Zeile 154) = 32, minus 2 für
  Dot+Leerzeichen = 30 nutzbare Zeichen für den Titel. „Conflict: bean
  changed externally" ist 34 Zeichen lang — wird bei 30 hart abgeschnitten
  (`…`).
- `overlay_show_toast.go:160`: dieselbe `ansi.Truncate(t.context, innerW,
  "…")`-Logik für die Context-Zeile — betrifft JEDE Toast-Instanz mit langem
  Titel/Context, nicht nur Conflict (PO-Anforderung „gilt für alle
  Severities" adressiert genau das).
- `update.go:275-327` (`applyMutationResult`): der Conflict-Zweig
  (`errors.Is(err, data.ErrConflict)`, Zeile 279-303) setzt `toastCtx` NUR,
  wenn `errors.As(err, &cr)` auf `*conflictWithRecovery` matcht (Zeile
  291-295, Editor-Conflict-Pfad mit Recovery-Tempfile). Der PLAIN-ErrConflict-
  Fall (kein Recovery-Tempfile, z. B. der `t`-Picker-Tagging-Konflikt aus dem
  bean-Repro) lässt `toastCtx = ""` (Zeile 290) unverändert bis zum
  `showToast`-Aufruf (Zeile 303) — obwohl `err` selbst (aus
  `internal/data/mutations.go:63/75`, `fmt.Errorf("%w: bean %s: %s",
  ErrConflict, id, ...)`) die Bean-ID UND die CLI-Fehlermeldung bereits
  trägt. Diese Information wird komplett verworfen; der Toast-Titel bleibt
  der generische Hardcoded-String „Conflict: bean changed externally" ohne
  jeden handlungsleitenden Zusatz.

**Vorgehen.**

1. `toastBoxWidth` (Konstante) durch eine content-getriebene Breitenberechnung
   in `toastBox()`/`toastGeometry()` ersetzen: Breite = längste Zeile
   (`dot + " " + title` bzw. `context`, via `ansi.StringWidth`), geklemmt in
   `[32, min(m.width-4, cap)]`. **Planner-Entscheidung (Cap-Wert, PO nannte
   keine Zahl):** `cap = 70` — orientiert an bestehenden Modal-Breiten-
   Konventionen (`clampModalWidth(46, ...)`, `box_filter_facets.go`), breit
   genug für die längsten heute vorkommenden Toast-Titel ohne auf jedem
   Terminal fast die volle Breite zu beanspruchen. Floor bleibt 32 (devd
   DD2-272 Parity, unverändert).
2. `ansi.Truncate(..., "…")`-Aufrufe für Titel UND Context entfernen. Wenn
   Inhalt auch bei der geklemmten Maximalbreite nicht in EINE Zeile passt:
   Wordwrap statt Truncate — bestehende `wrapText()`/`x/ansi`-Wordwrap-
   Bausteine (`view.go`, dort bereits für den Footer genutzt) wiederverwenden,
   keine neue Wrap-Implementierung. `toastGeometry()`s `h = len(lines)`
   (Zeile 187) adaptiert sich bereits automatisch an mehr Zeilen — keine
   separate Höhen-Verdrahtung nötig.
3. Fix lebt in `toastBox()`/`toastGeometry()` (dem gemeinsamen Rendering-Pfad
   ALLER drei `toastKind`-Werte) — NICHT als Conflict-Sonderfall, PO-Wortlaut
   „gilt für alle Toast-Severities" ist bindend.
4. `update.go`s plain-ErrConflict-Zweig (Zeile 279-303): `toastCtx` VOR dem
   `errors.As(&cr)`-Check auf `err.Error()` vorbelegen (z. B. `toastCtx :=
   err.Error()`), sodass der `errors.As`-Treffer (Editor-Conflict-Pfad) ihn
   weiterhin mit „Version saved: …" überschreibt (Zeile 294, Verhalten
   unverändert), der plain-Pfad aber NICHT mehr leer bleibt — mirrort exakt
   den Recovery-Zweig-Zeilen 290-303, wie D04 fordert.
5. Regressionstests: `overlay_show_toast_test.go` — neuer Test, dass ein
   langer Titel/Context NICHT truncated wird und die Box wächst (Breite
   und/oder Zeilenzahl); `etag_conflict_test.go` — neuer/erweiterter Test,
   dass der plain-ErrConflict-Pfad (kein `*conflictWithRecovery`-Match) einen
   NICHT-leeren `toastCtx` erzeugt.

**Akzeptanz:**
- [ ] `toastBoxWidth`s fixe 36 durch content-getriebene Breite ersetzt,
      geklemmt `[32, min(m.width-4, 70)]`
- [ ] Titel/Context nicht mehr per `ansi.Truncate("…")` abgeschnitten —
      Wordwrap auf mehrere Zeilen, wenn Inhalt die Maximalbreite überschreitet
- [ ] Gilt für alle drei `toastKind`-Severities (Info/Warn/Error), nicht nur
      Conflict
- [ ] Plain-ErrConflict-Zweig (`update.go`, kein `*conflictWithRecovery`-
      Match) setzt einen nicht-leeren `toastCtx` mit Bean-ID/CLI-Detail aus
      `err.Error()`
- [ ] Bestehendes Recovery-Pfad-Verhalten (`toastCtx = "Version saved: …"`)
      unverändert
- [ ] Test-Suite grün, neue Tests in `overlay_show_toast_test.go` UND
      `etag_conflict_test.go`
- [ ] Golden-Gegenbeleg falls Toast-Geometrie golden-relevant wird
- [ ] tmux-Smoke: `lean-stack-n0ly` (oder jedes andere Bean mit stalem ETag)
      via `t`-Picker taggen — vollständige Conflict-Meldung + Detailzeile
      sichtbar, kein Abschneiden

## Item 2: `createInFlightNote`-Toast-Severity vereinheitlichen (`bt-tm4a`)

**Root Cause.** Dieselbe Meldung (`createInFlightNote`, `update.go:704`,
„Creation already in progress — please wait") erscheint an DREI Guards in
ZWEI Severities:

- `update.go:735` → `m.showToast(toastWarn, createInFlightNote, "", nil,
  false)` — bereits korrekt (Ziel-Severity dieses Items).
- `box_confirm_create.go:56` → `m.showToast(toastError, createInFlightNote,
  "", nil, false)`.
- `overlay_palette.go:240` → `m.showToast(toastError, createInFlightNote, "",
  nil, false)`.

Beide `toastError`-Stellen tragen ein explizites ERRATUM-Code-Kommentar
(`box_confirm_create.go:49-53`, `overlay_palette.go:233-237`), das die
Diskrepanz bereits dokumentiert und auf den bindenden `bt-81f0`-Task-Rahmen
zurückführt (der bei seiner Einführung wörtlich `toastError` vorschrieb, ohne
`update.go:735`s bereits bestehenden `toastWarn` zu kennen/zu reconcilen).
Zwei Tests fixieren den aktuellen (inkonsistenten) Zustand explizit:
`box_confirm_create_test.go:415-417` und `overlay_palette_test.go:462-463`
assertieren je `toast.kind != toastError` mit dem Kommentar „bt-81f0
bindender Rahmen".

**Vorgehen.**

1. `box_confirm_create.go:56` und `overlay_palette.go:240` von `toastError`
   auf `toastWarn` ändern (`update.go:735` bleibt unverändert — bereits Ziel-
   Severity).
2. Die drei ERRATUM-Kommentare (`box_confirm_create.go:49-53`,
   `overlay_palette.go:233-237`, plus `update.go:737`s eigener Verweis auf
   die Diskrepanz) entfernen bzw. auf den jetzt EINHEITLICHEN Zustand
   umschreiben — sie dokumentieren sonst eine bereits behobene Abweichung und
   verwirren einen künftigen Leser.
3. Begründung (Akzeptanz-Pflicht, Bean-Vorgabe): in-flight-Guard ist kein
   Fehler, sondern ein Hinweis „bitte kurz warten" (`createInFlightNote`s
   eigener Wortlaut) — `toastWarn` (Gelb) passt der Nutzer-Intention besser
   als `toastError` (Rot), das für tatsächliche Fehler reserviert bleibt.
4. Testanpassung: `box_confirm_create_test.go:415-417` und
   `overlay_palette_test.go:462-463` — Assertion von `!= toastError` auf
   `== toastWarn` (bzw. `!= toastWarn` als Negativ-Check) drehen, Kommentar-
   Verweis auf „bt-81f0 bindender Rahmen"/ERRATUM entfernen.

**Akzeptanz:**
- [ ] Alle drei `createInFlightNote`-Toasts (`update.go:735`,
      `box_confirm_create.go:56`, `overlay_palette.go:240`) nutzen `toastWarn`
- [ ] ERRATUM-Kommentare an allen drei Stellen entfernt/aktualisiert
- [ ] Begründung (toastWarn statt toastError) im Bean-Abschluss dokumentiert
- [ ] `box_confirm_create_test.go`/`overlay_palette_test.go`-Assertions auf
      `toastWarn` angepasst
- [ ] Test-Suite grün

## Item 3: Suche mit Filter-Präfixen (`bt-2kfl`)

**PO-Entscheidungen final (D02/D03, bean-Sektion „PO-Entscheidungen Grilling
2026-07-17"), NICHT neu aufgemacht:**
- **D02 — Separater additiver Layer:** getippte Präfixe verändern
  `m.filterStatus`/`m.filterType`/`m.filterPriority`/`m.filterTag` NICHT —
  das `f`-Menü bleibt unberührt. `filterSummary` zeigt die UNION aus
  Menü-Facetten und getippten Präfixen. Query löschen = getippte Filter weg.
- **D03 — Parser bei jedem Tastendruck:** Parser trennt lokal Präfix-Paare +
  Rest-Text; NUR der Rest-Text geht den bestehenden Such-Pfad (Bleve ab 3
  Zeichen des Rests). Präfixe erreichen Bleve nie.

**Ist-Stand (ERRATUM vs. bean-Zeilenangaben — Code hat sich seit der
Diagnose durch E12 NICHT verschoben für `view_browse_repo.go`, WOHL aber für
`box_filter_facets.go`):**
- `view_browse_repo.go:251-260` (`beanMatchesSearch`) — Zeilen decken sich
  exakt mit der bean-Angabe, kein Drift.
- `box_filter_facets.go`: `filterSummary` liegt inzwischen bei Zeile
  **500-517** (NICHT mehr 324-339 wie im bean zitiert) — `bt-2p9m`s
  Querformat-Umbau (Commit `b150c9f`/`e3b1701`, nach der bean-Diagnose vom
  2026-07-17 gemerged) hat das File umsortiert. `facetHead`
  (Zeile 34-40 laut bean) liegt jetzt bei Zeile **32-38**. Inhalt/Vertrag
  beider Funktionen unverändert, nur die Zeilennummern verschoben.
- `update.go:1588-1607` (`keySearchInput`): JEDER Tastendruck aktualisiert
  `m.searchQuery` bereits live (Zeile 1606) und dispatcht
  `dispatchBleveIfDue` (Zeile 1607) — das ist der D03-Hook-Punkt, kein neuer
  Erfassungspunkt nötig.
- `beanMatches` (`box_filter_facets.go:189-191`) UND-verknüpft bereits
  `beanMatchesSearch`/`beanMatchesFacets`/`beanMatchesArchive` — Infrastruktur
  zum Kombinieren existiert unverändert.

**Vorgehen.**

1. Neue reine Parse-Funktion (z. B. `view_browse_repo.go` oder ein neues
   `search_prefix.go`, Implementer-Wahl): `parseSearchPrefixes(query string)
   (facets map[string][]string, rest string)`. Tokenisiert whitespace-
   getrennt; jedes Token der Form `<prefix>:<value>` mit `prefix` ∈
   `{st,ty,pr,tag}` (case-insensitiv, mirrort `facetHead` MINUS `archive` —
   PO-Zitat „das hat keinen Präfix") wird in `facets["status"|"type"|
   "priority"|"tag"]` einsortiert (Kürzel→Facet-Name-Mapping analog
   `facetHead`s Werte); alles andere (ungültiges Präfix, Wert ohne Doppelpunkt,
   leerer Wert) fällt unverändert in `rest` (space-gejoint, Reihenfolge
   erhalten) — PO-Akzeptanz „kein Fehler" explizit.
2. `keySearchInput` (`update.go:1588-1607`): nach jedem
   `m.searchQuery = strings.TrimSpace(...)`-Update `facets, rest :=
   parseSearchPrefixes(m.searchQuery)` aufrufen, Ergebnis in ZWEI NEUEN
   Model-Feldern ablegen (`types.go`, Implementer-Namenswahl, z. B.
   `m.searchPrefixFacets map[string][]string` und `m.searchPrefixRest
   string`) — GETRENNT von `m.filterStatus`/etc. (D02: additiver Layer, kein
   gemeinsamer State). `dispatchBleveIfDue` muss ab jetzt gegen `rest`
   dispatchen, NICHT gegen die volle `m.searchQuery` (D03: Präfixe erreichen
   Bleve nie).
3. `beanMatchesSearch` (`view_browse_repo.go:251-260`): den lokalen/Bleve-
   Substring-Abgleich gegen `m.searchPrefixRest` statt `m.searchQuery` laufen
   lassen; zusätzlich (neue AND-Bedingung, gleiche Funktion oder ein neuer,
   von `beanMatches` mit-aufgerufener Predicate-Baustein) `m.searchPrefixFacets`
   gegen `b.Status`/`b.Type`/`b.Priority`/`b.Tags` prüfen — Membership-
   Semantik pro Facet identisch zu `beanMatchesFacets` (mehrere Werte
   derselben Facet = OR, verschiedene Facets = AND), damit `st:completed
   ty:epic foo` sich exakt wie in der PO-Akzeptanz verhält.
4. `filterSummary` (`box_filter_facets.go:504-517`, NICHT 324-339, siehe
   ERRATUM oben): pro Facet-Zeile (`St:`/`Ty:`/`Pr:`/`Tags:`) die UNION aus
   `m.filterStatus`/etc. Keys UND `m.searchPrefixFacets`s entsprechenden
   Keys anzeigen (dedupliziert, `joinFilterKeys`-Muster wiederverwenden oder
   erweitern) — D02s „eine Wahrheit mit Facetten-State" wörtlich als
   Anzeige-Union, NICHT als Schreib-Merge in `m.filterStatus` selbst.

**Akzeptanz** (Bean-Vorgabe, unverändert übernommen):
- [ ] `/st:completed ty:epic foo` filtert auf `status=completed AND
      type=epic AND` Text-Match `foo`
- [ ] Präfixe case-insensitiv, Reihenfolge egal
- [ ] Ungültiges Präfix/Wert → als normaler Suchtext behandelt, kein Fehler
- [ ] Header-Filteranzeige (`filterSummary`) spiegelt die UNION aus
      Menü-Facetten und getippten Präfixen (D02)
- [ ] Parser läuft bei jedem Tastendruck, NUR `rest` erreicht Bleve (D03)
- [ ] `m.filterStatus`/etc. (das `f`-Menü) bleibt von getippten Präfixen
      UNBERÜHRT (D02, additiver Layer — kein Sync-Schreibzugriff)
- [ ] Test-Suite grün, neue Tests für `parseSearchPrefixes` UND die
      Union-Anzeige in `filterSummary`
- [ ] tmux-Smoke: `/st:completed ty:epic` tippen, Tree-Kopf zeigt
      `St:completed Ty:epic`, `f`-Menü bleibt beim Öffnen leer (kein
      versehentliches Mit-Togglen)

## Item 4: Lobby `slug — Pfad` + Palette-Befehl „register project" (`bt-d3ps`)

**PO-Redefinition final (bean-Sektion „PO-Redefinition Grilling 2026-07-17
(ersetzt Discovery-Scan-Ansatz)"), NICHT neu aufgemacht:** KEIN Scan. Lobby
zeigt weiterhin exakt die config-Repos, Darstellung je Eintrag `slug — Pfad`.
NEU: Palette-Befehl „register project" registriert das aktuell geöffnete
Repo im zentralen Register. Eriks eigene `config.yaml`-Ergänzung ist ein
Supervisor-Schritt AUSSERHALB dieses beans.

**Ist-Stand / Root Cause (fehlender Baustein).** `view_lobby.go:174-189`
(`repoPickerBody`) rendert je Zeile ausschließlich den rohen Pfad `r`
(`m.settings.Repos`, `[]string`, `internal/config/settings.go:49`) — kein
Slug-Konzept existiert im Code. `.beans.yml` (Repo-Root, z. B.
`beans-tui-repository/.beans.yml`) trägt bereits `beans.prefix` (z. B.
`"bt-"`) UND optional `project.name` — beans-tui liest diese Datei bisher
NIRGENDS selbst (nur `internal/data/discover.go`s `FindRepo` sucht ihre
BLOSSE Existenz als Repo-Marker, parst den Inhalt nicht). `overlay_palette.go`
hat KEINEN Case, der `config.yaml` schreibt — jede Repo-Registrierung läuft
heute ausschließlich über den manuellen Settings-Form-Textarea-Pfad
(`box_form_settings.go`, `linesToRepos`/`SaveUserSettings`).

**Vorgehen.**

1. **Neuer Slug-Resolver (Planner-Entscheidung, Quelle nicht vom PO
   spezifiziert):** `internal/data` (liegt bereits bei `.beans.yml`-Wissen,
   `discover.go`) bekommt `func RepoSlug(repoDir string) string` — liest
   `<repoDir>/.beans.yml`, parst minimal `{Beans struct{Prefix string
   `+"`yaml:\"prefix\"`"+`}}` (gleiches `gopkg.in/yaml.v3`, bereits
   `internal/config`-Dependency, keine neue). Rückgabe:
   `strings.TrimSuffix(prefix, "-")` falls `prefix != ""`, sonst
   `filepath.Base(repoDir)` als Fallback (Datei fehlt/unlesbar/`prefix`
   leer). **Begründung gegen `project.name` als Primärquelle:** Bean-IDs
   (`bt-0xrb`, `lean-stack-58j0`, überall im UI sichtbar) sind an
   `beans.prefix` gekoppelt, NICHT an `project.name` — `project.name` fehlt
   sogar im lean-stack-eigenen `.beans.yml` (nur `beans-tui`s eigenes trägt
   es) und wäre damit keine verlässliche repo-übergreifende Quelle.
2. `view_lobby.go:174-189` (`repoPickerBody`): Zeilen-Label von `r` auf
   `data.RepoSlug(r) + " — " + r` ändern, bestehendes `nameW`-Truncate-Budget
   (Zeile 183-187) unverändert weiterverwenden — Implementer entscheidet, ob
   bei zu knappem `nameW` der Slug oder der Pfad zuerst gekürzt wird.
3. **Palette-Befehl:** `paletteActions` (`overlay_palette.go:83-103`, globaler
   Block) neuer Eintrag `paletteItem{kind: paletteKindAction, actionID:
   "register_project", label: "register project"}` — Platzierung neben
   `repo_picker`/`settings` (gleiche Nachbarschaft, mirrort bestehende
   Gruppierungs-Konvention dieser Liste). `dispatchPalette`
   (`overlay_palette.go:183-264`) neuer Case `"register_project"` →
   neue Methode `m.registerProject() (tea.Model, tea.Cmd)`:
   - `m.client == nil` (Palette aus der Lobby selbst geöffnet, kein Repo
     live) → no-op (mirrort `edit_title`s `if b := m.focusedBean(); b !=
     nil`-Gating-Muster, Zeile ~217-220) statt Crash.
   - Bereits registriert (`m.client.RepoDir` ∈ `m.settings.Repos`) →
     `toastInfo`-Hinweis „already registered", kein Doppel-Eintrag.
   - Sonst: `config.SaveUserSettings(append(m.settings.Repos,
     m.client.RepoDir), m.settings.Editor, m.settings.Theme.Accent,
     m.settings.Layout.TreeWidth)` — bestehende Signatur
     (`internal/config/settings.go:147`) unverändert wiederverwendet, kein
     neuer Schreibpfad. Erfolg: `m.settings.Repos` im Model ebenfalls
     aktualisieren + `toastInfo`-Bestätigung („Registered: " + Slug).
     Fehler: `toastError` mit `err.Error()`.
   - Lobby-Liste selbst braucht KEINE separate Refresh-Verdrahtung —
     `openLobby()` (`view_lobby.go:321-334`) lädt `config.LoadSettings()`
     bereits bei JEDEM Öffnen neu ein (eigener Doc-Kommentar: „lädt
     Settings.Repos neu, falls seit Start geändert").
4. Supervisor-Schritt (Eriks persönliche `config.yaml` um bestehende Projekte
   ergänzen) bleibt EXPLIZIT außerhalb dieses beans/dieses Plans.

**Akzeptanz:**
- [ ] Lobby-Zeile zeigt `slug — Pfad` statt nur `Pfad`, für jeden
      config-Repo-Eintrag
- [ ] `RepoSlug()`: `beans.prefix` (trimmed) als Primärquelle, Dir-Basename
      als Fallback bei fehlendem/unlesbarem `.beans.yml`
- [ ] Palette-Befehl „register project" registriert `m.client.RepoDir` im
      zentralen Register (`config.yaml`), dedupliziert gegen bereits
      registrierte Pfade
- [ ] Erfolgreiche Registrierung zeigt Toast-Bestätigung; Lobby zeigt den
      neuen Eintrag beim nächsten Öffnen (bestehender `openLobby`-Reload)
- [ ] `m.client == nil` → Aktion no-op, kein Crash
- [ ] KEIN automatischer Scan, KEINE Scan-Wurzeln, KEINE Fund-Persistenz
      (PO-Redefinition final)
- [ ] Test-Suite grün, neue Tests für `RepoSlug()` UND `registerProject()`
- [ ] tmux-Smoke: lean-stack-Repo öffnen, „register project" ausführen,
      Lobby zeigt neuen Eintrag mit korrektem Slug (`lean-stack`, aus
      `beans.prefix`)

## Item 5: Footer-Hint bei offenem Filter-Menü zeigt falsches Label (`bt-nxuk`)

**Root Cause.** `footer_context.go:43-45` (`filterMenuLocalBindings`) gibt
`keys.FocusIn`/`keys.FocusOut` DIREKT zurück — dieselben globalen
`keybind.Binding`-Werte, die `keymap.go:136-137` mit `WithHelp("tab", "focus
in")`/`WithHelp("shift+tab", "focus out")` definieren. Die Datei-eigene
Doc-Notiz (Zeile 33-42) dokumentiert bereits, dass `tab`/`shift+tab` hier
„a DIFFERENT, filter-menu-local meaning" haben (Kategorie-Wechsel statt
Tree↔Detail-Fokus) — aber die WIEDERVERWENDETE `Binding` trägt weiterhin das
GLOBALE Label. `renderBindings` (`view.go:128-137`) rendert `h.Key`/`h.Desc`
direkt aus dem `Binding`, also erscheint in Footer Zone 3
(`contextualLocalHint`, `footer_context.go:185-208`, Fall `m.filterOpen`,
Zeile 187-188) wörtlich „tab focus in · shift+tab focus out" — obwohl
`keyFilterMenu` (`box_filter_facets.go:380-397`) `tab`/`shift+tab` tatsächlich
an `filterMenuSwitchTab` (Kategorie-Wechsel) bindet, UND das Modal-interne
Hint (`treeFilterBox`, `box_filter_facets.go:474-479`, Zeile 476) bereits
korrekt „tab/shift+tab:category" zeigt. Fußzeile und Modal-Hint widersprechen
sich für Nutzer, die nur die Fußzeile lesen.

**Vorgehen.** EIN neues, eigenständiges `keybind.Binding` definieren (NICHT
als `keyMap`-Struct-Feld — bleibt lokal, z. B. package-level `var` oder
Inline-Literal in `footer_context.go`) mit `keybind.WithKeys("tab",
"shift+tab")`, `keybind.WithHelp("tab/shift+tab", "category")` — mirrort
`treeFilterBox`s eigenen Wortlaut „tab/shift+tab:category" exakt. Die beiden
`keys.FocusIn, keys.FocusOut`-Einträge in `filterMenuLocalBindings()`
(`footer_context.go:44`) durch dieses EINE kombinierte Binding ersetzen.

**Verifiziert gegen beide Drift-Guards (Code gelesen,
`keymap_test.go:198-274`):**
- `TestNoDuplicateBindingBetweenGlobalAndAnyLocalHintList`
  (`keymap_test.go:208-225`) scoped NUR `browseRepoLocalBindings`/
  `backlogLocalBindings` (Zeile 213-216) — `filterMenuLocalBindings` liegt
  außerhalb seines Scans, bleibt unberührt.
- `TestHelpGroupsCoverEveryBindingExactlyOnce` (`keymap_test.go:233-274`)
  reflektiert AUSSCHLIESSLICH über `keyMap`-Struct-Felder (`v :=
  reflect.ValueOf(k)`, Zeile 235). Ein lokales, NICHT-`keyMap`-Binding-
  Literal ist für diese Reflection unsichtbar — kein neues Struct-Feld
  nötig, kein Risiko, den Guard zu verletzen.

**Akzeptanz:**
- [ ] Footer Zone 3 zeigt bei offenem Filter-Menü (`m.filterOpen`)
      „tab/shift+tab category" statt „tab focus in · shift+tab focus out"
- [ ] Globales `tab`/`shift+tab` (FocusIn/FocusOut) außerhalb des
      Filter-Menüs UNVERÄNDERT
- [ ] `TestNoDuplicateBindingBetweenGlobalAndAnyLocalHintList` grün
- [ ] `TestHelpGroupsCoverEveryBindingExactlyOnce` grün
- [ ] Test-Suite grün, neuer Test (z. B.
      `TestFilterMenuFooterHintShowsCategoryLabel`)

## Offene Fragen für den PO

Keine. Alle fünf beans tragen finale, datierte PO-Entscheidungen aus dem
Grilling-Termin 2026-07-17 (D04 bei `bt-0xrb`, D02/D03 bei `bt-2kfl`,
PO-Redefinition bei `bt-d3ps`; `bt-tm4a`/`bt-nxuk` sind reine, unstrittige
Reviewer-Follow-ups ohne offene Design-Achse). Die einzigen in diesem Plan
getroffenen Entscheidungen sind Implementierungsdetails ohne Produktachse
(Toast-Maximalbreite 70 Zeichen, `RepoSlug()`-Quellenpriorität
`beans.prefix` vor Dir-Basename) — beide unten in „Design-Entscheidungen"
begründet, keine PO-Rückfrage nötig.

## Selbst-Review

- **Deckung:** alle fünf beans (`bt-0xrb`, `bt-tm4a`, `bt-2kfl`, `bt-d3ps`,
  `bt-nxuk`) je genau einem der fünf Plan-Abschnitte zugeordnet, keines
  doppelt, keines vergessen — Quelle: der vom Supervisor übergebene
  5-bean-Scope, alle mit Parent `bt-5uzr` bestätigt (`beans show` je bean).
- **Struktur:** ein empfohlenes (nicht hartes) Reihenfolge-Constraint
  (Toast-Familie #1→#2 vor der Parallel-Welle #3/#4/#5) aus geteilter
  Datei-Nachbarschaft (`update.go`, `overlay_palette.go`) — kein
  `blocked_by` in der Item-Tabelle als HART markiert, da kein Zeilen-
  Overlap nachweisbar ist (nur Datei-Nachbarschaft, mirrort
  `epic-E12-plan.md`s eigene Parsimonie-Konvention). #3/#4/#5 sind
  nachweislich disjunkt (keine gemeinsame Datei) und damit
  worktree-parallelisierbar, wie vom Supervisor vorgegeben.
- **bean-Qualität:** jeder Abschnitt nennt Root-Cause mit `datei:zeile`
  (gegen den TATSÄCHLICHEN Ist-Code verifiziert, nicht nur gegen die
  bean-Zitate — ein Zeilen-Drift bei `bt-2kfl`s `box_filter_facets.go`
  wurde gefunden und als ERRATUM markiert, siehe Item 3), konkretes
  Vorgehen, abhakbare Akzeptanz-Checkboxen mit Artefakten (Testnamen,
  tmux-Smoke-Schritte) — ein frischer Agent kann ohne Rückfrage in JEDEM
  Item starten.
- **Keine Anti-Patterns:** keine PO-Entscheidung neu aufgemacht (D04 bei
  `bt-0xrb`, D02/D03 bei `bt-2kfl`, Redefinition bei `bt-d3ps` je wörtlich
  referenziert, nicht neu verhandelt); die zwei Implementierungs-
  Entscheidungen dieses Plans (Toast-Cap 70, `RepoSlug()`-Quelle) sind
  reine Implementierungsdetails ohne Produktachse, explizit als
  Planner-Entscheidung markiert statt stillschweigend als PO-Vorgabe
  ausgegeben.
- **Keine Status-/Tag-Änderungen:** dieser Plan selbst ändert an keinem der
  fünf beans Status oder Tags — jeder Body-Append (separater Schritt) trägt
  nur eine datierte `## Plan-Konkretisierung E13 (2026-07-17)`-Sektion an,
  additiv, kein Body-Overwrite. Keine Epic-Tag-Änderung an `bt-5uzr`.
