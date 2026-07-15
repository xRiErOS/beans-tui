# Epos E4 — Command-Center & Review-Cockpit (voll granular)

> Geschrieben beim Epos-Start-Ritual (implementation-plan.md »Epos-Rituale«). Struktur
> identisch zu E1/E2/E3: je Task Files/TDD-Steps/Port-Referenzen/Commit. Quelle der
> Wahrheit: `design-spec.md` §5 (Review-Tag-Konvention), §6 V5/V6, §7 (Keymap) und der
> Epic-bean-Body `bt-tfqi`. Port-Quellen:
> **devd** (`~/Obsidian/tools/DeveloperDashboard/apps/cli-go/internal/tui/`):
> `overlay_palette.go` (Command-Center, fuzzyMatch), `view_navigate_reviews.go` +
> `view_review_sprint.go` (Review-Master-Detail-Layout-Muster, NUR Layout — devd hat
> Sprint als eigene Entität, beans-tui nicht, s. Design-Entscheidung c), `keys_review.go`
> (Verdikt-Dispatch-Muster). Task-beans: `bt-jpgn`(T1) `bt-yo60`(T2) `bt-hxyo`(T3)
> `bt-yy6w`(T4) `bt-v7ti`(T5), parent `bt-tfqi`.

**Liefert:** V5 Command-Center (`ctrl+k`/`K`: fuzzy-Palette, Aktionen + Bean-Treffer
gemischt, kontextabhängig zuerst) und V6 Review-Cockpit (`R`: Queue `to-review`
gruppiert nach Epic, Verdikt-Dots, Master-Detail, `a` pass/`x` reject/`o` reopen,
Summary „x of n") — US-04/US-08. Der Datenlayer ist für beide Views bereits fast
komplett (`internal/data/index.go: WithTag` seit T3-data, `internal/data/mutations.go`
mit `SetTags`/`SetStatus`/`AppendBody` aus E3) — E4 ergänzt nur zwei neue,
kombinierte Mutations-Wrapper (`PassReview`/`RejectReview`, Design-Entscheidung d) und
einen neuen reinen Index-Walk (`EpicAncestor`, Design-Entscheidung c).

**NICHT E4-Scope:** Yank (`y`, OSC52 + nativ) — `internal/clip` existiert noch nicht
(E5), Review-Stand-Yank bleibt in der Cockpit-View bis E5 unbelegt (Design-Entscheidung
g). Repo-Picker (`p`) — Lobby ist E5. Help-Overlay (`?`) — E5; die Palette referenziert
daher KEINE „help"-Aktion (keine Aktion, die auf nichts Existierendes zeigt). Toast —
weiterhin Statuszeilen-Interim (E3-Konvention, unverändert).

**Reihenfolge (Task-bean `blocked_by`, streng sequentiell):** T1 → T2 → T3 → T4 → T5.
T1 baut die Command-Center-Infra (Capture-Order-Erweiterung, `paletteOpen`-State,
Fuzzy-Matcher) inklusive des vollständigen, aber noch bean-suchlosen Aktions-Dispatch;
T2 ergänzt die Bean-Suche (eigenständig testbar, CLI-abhängig). T3 baut die
Review-Queue-Datenableitung + eine rein lesbare Cockpit-Skeleton-View (inkl.
Capture-Order-Erweiterung für `m.view == viewReviewCockpit`); T4 verdrahtet die
Verdikt-Mutationen UND ergänzt die Palette (T1) um die jetzt existierenden
Review-Cockpit-Sprungziele — exakt das gleiche inkrementelle Muster, das devds eigener
`overlay_palette.go`-Kommentar bereits vorschreibt („Reviews (T17) und Memory (T18)
werden hier ergänzt, sobald die Views existieren").

---

## Design-Entscheidungen (vor Task 1 festgezurrt)

**(a) Fuzzy-Matching: eigene Subsequence-Implementierung, KEIN `sahilm/fuzzy`.**
Verifiziert: es gibt keine Fuzzy-Lib im `go.mod` (Handover-Gotcha bestätigt). devds
eigenes `overlay_palette.go:fuzzyMatch` (Zeilen 60-71: rune-basiert, case-insensitiv,
Subsequence-Match, leere Query matcht alles) wird VERBATIM nach `internal/tui/fuzzy.go`
portiert — kein neuer Dependency, ausreichend für eine Aktionsliste (~10 Einträge) und
ein Repo mit üblicherweise niedriger dreistelliger Bean-Zahl. `sahilm/fuzzy` (vik-Muster
aus dem NSP-Gotcha) wäre eine graduierte Score-Function (bessere Ranking-Qualität bei
sehr großen Listen) — für v1 YAGNI, dokumentierter Scope-Cut. Bei Bedarf später
austauschbar, da `fuzzyMatch(query, target string) bool` die einzige Berührungsfläche
ist (ein Funktionssignatur-Tausch, kein struktureller Umbau).

**(b) Palette-Ergebnis-Modell: EIN Eingabefeld (`m.palQuery`) treibt ZWEI
Kandidaten-Pools, Aktionen IMMER vor Beans.** `design-spec.md` §6 V5: „fuzzy Aktionen
… + Bean-Suche (Bleve) in einem". Ein `paletteItem{kind, actionID, bean, label}` fasst
beide Pools:
- **Aktionen** (`paletteActions(m)`, T1): eine KONTEXTABHÄNGIGE Liste — wenn
  `m.focusedBean() != nil`, werden zuerst die Node-Aktionen des fokussierten Beans
  vorangestellt (`status: setzen`, `tags: zuweisen`, `parent: zuweisen`,
  `blocking: zuweisen`, `titel: bearbeiten`, `bean: löschen` — Wortwahl `verb: label`
  spiegelt devds DD2-185-Konvention, Zeile 33 der Quelle: „create: issue"/„go to:
  …"), DANACH die globalen View-/Aktions-Einträge (`create: bean`, `go to: backlog`,
  `go to: browse`, `filter: facetten`, `search: beans`, `reload: daten`; T4 ergänzt
  `go to: review cockpit`, sobald die View existiert — devd-Präzedenz, s. Datei-Kopfzitat
  oben). Jeder Eintrag wird gegen `m.palQuery` per `fuzzyMatch` gefiltert.
- **Beans** (`palFilteredBeans(m)`, T2): NUR wenn `strings.TrimSpace(m.palQuery) != ""`
  (ein leeres Query listet NICHT das ganze Repo — anders als `/`s Tree-Filter, der auf
  einer bereits sichtbaren Baumstruktur operiert; die Palette ist ein kompaktes Modal).
  Matching reused wörtlich `m.beanMatchesSearch`s Bifurkation (view_browse_repo.go:
  211-220): unter 3 Zeichen / solange kein Bleve-Treffer für die AKTUELLE Query vorliegt
  → lokaler Title+ID-Substring; ab 3 Zeichen, sobald `m.palBleveFor == m.palQuery` →
  Bleve-Ergebnis (UNION lokaler ID-Match, gleiche Begründung wie beanMatchesSearch:
  Bleve indiziert Title+Body, nicht zwingend eine ID-Teilzeichenkette). Gedeckelt auf
  `paletteBeanResultCap = 20` Treffer (Design-Entscheidung, verhindert eine Modal-Flut
  bei einer breiten Query wie „e"), sortiert via `data.SortBeans` (kanonische
  Determinismus-Garantie, I03-Konvention).
- **Reihenfolge:** gematchte Aktionen ZUERST (in ihrer festen Listenreihenfolge, s.o.),
  dann gematchte Beans (kanonisch sortiert). KEINE Score-basierte Verzahnung — devds
  eigenes `fuzzyMatch` ist binär (Treffer/kein Treffer), ein gemeinsames Scoring wäre
  Overengineering ggü. dem Port-Quellcode; „Aktionen zuerst" IST bereits die
  „kontextabhängig zuerst"-Eigenschaft, die die Handover-Vorgabe verlangt (die
  Kommandopalette priorisiert Aktionsdiskoverability, `/` bleibt die dedizierte
  Bean-Suche).
- **Eigenständiges Query-Feld, KEIN Reuse von `m.searchQuery`:** ein geteiltes Feld
  würde eine aktive Tree-/Backlog-Suche beim Öffnen der Palette stillschweigend
  überschreiben und beim Schließen der Palette einen Fremd-Query im Tree hinterlassen —
  `m.palQuery`/`m.palBleveIDs`/`m.palBleveFor`/`m.palBleveLoading` sind EIGENE,
  Palette-skalierte Kopien des exakt gleichen Staleness-Guard-Tripletts, das `/` bereits
  für sich selbst etabliert hat (types.go-Konvention, wörtlich wiederholt, NIE geteilt —
  gleiche Vorsicht wie bei `tagInput` vs. `searchInput`, zwei unabhängige
  `textinput.Model`-Instanzen desselben Musters).

**(c) Review-Queue-Ableitung: `idx.WithTag("to-review")` (data, E1 vorhanden) +
NEUER reiner Walk `data.EpicAncestor` + TUI-seitige Gruppierung.** Kein neues
persistentes Feld, keine Sprint-Entität (lean-stack: „Sprint ≡ Epos ≡ bean", D02
lean-stack-key-decisions) — Gruppierung läuft rein über den bestehenden `Parent`-Baum:
`EpicAncestor(idx, b)` läuft `b.Parent` aufwärts (zyklen-geschützt, exakt das gleiche
`visited`-Muster wie `expandAncestorsOf`/`CollectDescendants`) bis der nächste Bean vom
`Type == "epic"` gefunden wird. Kein Treffer (parentloser to-review-Bean, oder ein
direkt unter einem Milestone hängender Bean OHNE Epic dazwischen) → Gruppe
`(kein Epic)`, IMMER als letzte Gruppe gerendert (gleiche „Sonderfall zuletzt"-Konvention
wie der `orphanRootID`-Wurzelknoten im Tree). Gruppen mit echtem Epic sind selbst
kanonisch sortiert (`data.SortBeans` auf die Liste der Epic-Beans, dann je Gruppe
`data.SortBeans` auf ihre to-review-Mitglieder) — deterministisch, keine Map-Walk-Reste
(gleiche Lehre wie `tagFilterOptions`s ERRATUM).

**Rework-Sektion (Design-Erweiterung, s. (f) unten für die Begründung):** die Cockpit
zeigt UNTER den to-review-Gruppen einen zweiten, FLACHEN (nicht epic-gruppierten)
Abschnitt „Rework" (`idx.WithTag("rework")`, `data.SortBeans`) — Sichtbarkeits-Fenster
für Beans, die der PO bereits abgelehnt hat und auf die Agent-Nacharbeit warten;
`o` (Design-Entscheidung f) operiert NUR auf dieser Sektion.

**(d) Verdikt-Mutationen: je EIN kombinierter `beans update`-Aufruf, verifiziert gegen
`beans update --help`.** Verifiziert (dieser Ritual-Lauf, `command beans update --help`):
`-s/--status`, `--remove-tag`/`--tag` (beide `stringArray`) und `--body-append` sind
UNABHÄNGIGE Flags desselben `update`-Kommandos, beliebig kombinierbar — exakt die gleiche
Erkenntnis, die E3 Task 2/3 bereits zu `SetTags`/`SetBlocking` geführt hat (N
sequenzielle Mutationen auf EINEM ETag wären eine Konflikt-Kaskade, da der jeweils
erste Aufruf den ETag auf der Platte rotiert und unsere Wrapper das zurückgegebene Bean
verwerfen). Zwei NEUE `internal/data/mutations.go`-Methoden (Task 4):

```go
// PassReview marks a to-review bean accepted in ONE beans-update call (design
// decision d, bean bt-yy6w): status -> completed AND the to-review tag removed,
// atomically against a single etag. Mirrors design-spec.md §5's Pass row
// verbatim -- no body change on Pass (unlike Reject).
func (c *Client) PassReview(id, etag string) error {
    return c.update(id, etag, "--status", "completed", "--remove-tag", "to-review")
}

// RejectReview swaps to-review -> rework AND appends a dated "## Review
// <date>" body section, in ONE beans-update call (design decision d) --
// combining --remove-tag/--tag/--body-append avoids the same etag-cascade
// risk SetTags (E3 Task 2) already documented. date is caller-supplied
// (view_review_cockpit.go passes time.Now().Format("2006-01-02")) rather
// than computed here, purely for deterministic tests (client_mut_test.go
// pins the exact section text without touching the clock).
func (c *Client) RejectReview(id, comment, date, etag string) error {
    section := "\n## Review " + date + "\n\n" + comment + "\n"
    return c.update(id, etag, "--remove-tag", "to-review", "--tag", "rework", "--body-append", section)
}
```

**Reopen** (`o`, Rework → to-review) braucht KEINE neue Methode — reine Wiederverwendung
von E3 Task 2s bereits generischem `data.SetTags(id, []string{"to-review"},
[]string{"rework"}, etag)` (EIN kombinierter Aufruf, gleiche Signatur, kein neuer Code).

**(e) Reject-Kommentar-Form: `huh.NewText`, PFLICHTFELD, KEIN Confirm-Gate, KEIN
eingefrorener ETag.** Wiederverwendung der kompletten E3-Task-4-Form-Infra
(`forms_shared.go`: `styleForm`/`formChrome`/`paletteFormTheme`/`keyForm`/`updateForm`)
— NEUER `formKind` „reject" (`internal/tui/form_reject_review.go`,
`buildRejectReviewForm()`, EIN Feld „comment", `nonEmpty`-validiert: eine
Ablehnung ohne Begründung entwertet genau das Feedback, das
`## Review <datum>`-Abschnitte dem Agenten liefern sollen — anders als der
Create-Form (Design-Entscheidung e, E3) gibt es HIER kein optionales Body-Feld, das
Kommentar-Feld IST der Body-Beitrag). `submitForm`s bestehender `formKind`-Switch
(`box_confirm_create.go`) bekommt einen dritten Case, DIREKT feuernd wie „editTitle"
(E3 Task 5) — KEIN Confirm-Gate (Design-Entscheidung a2/e-Präzedenz: nur „create" ist
gated, „editTitle" und jetzt „reject" feuern sofort). ETag wird bei Submit FRISCH aus
`m.beanETag(id)` gelesen (Design-Entscheidung d aus E3 gilt hier UNVERÄNDERT, anders als
`ctrl+e`s Editor-Suspend/F2-Ausnahme — ein huh-Text-Formular ist Sekunden, nie Minuten
lang offen, das Lost-Update-Risiko des langlebigen `$EDITOR`-Suspends besteht hier
nicht).

**(f) Reopen-Semantik: `o` = „Rework manuell zurück in die Queue", NICHT in
`design-spec.md` §5s Tabelle explizit definiert — hier bewusst hergeleitet.** §5s
Tabelle hat FÜNF Zeilen (Agent tag to-review → PO sieht Queue → PO Pass/Reject → Agent
tag rework→to-review) und KEINE eigene Reopen-Zeile; §7/Cockpit-Doku listen `o` aber
ausdrücklich als View-lokalen Override. Da die Queue selbst rein `WithTag("to-review")`
ist (Design-Entscheidung c), kann `o` nicht „einen Verdikt auf dem GERADE sichtbaren
Queue-Item zurücksetzen" bedeuten — jedes sichtbare to-review-Item ist per Definition
bereits unverdikted. Die kohärente, nützliche Lesart (Design-Entscheidung, hier
verbindlich): `o`, gedrückt auf einem Bean der neuen Rework-Sektion (c), führt MANUELL
genau den Schritt aus, den §5s letzte Zeile sonst dem AGENTEN zuschreibt („Tag rework →
to-review") — der PO muss nicht warten/den Agenten bitten, sondern kann ein
fehlerhaft/vorschnell abgelehntes Bean selbst sofort zurück in die aktive Queue holen.
`o` auf einem to-review-Item (noch nicht verdiktet) ist ein No-op (bereits im
Zielzustand — spiegelt devds eigenen `reviewReopenable`-Guard, hier auf einen simplen
Typ-Check reduziert statt eines Status-Enums). Diese Entscheidung ist bewusst
konservativ/reversibel: eine reine Funktions-Änderung, falls der PO im Review anders
entscheidet.

**(g) Yank-Deferral: `y` bleibt in E4 UNBELEGT innerhalb der Cockpit.** `internal/clip`
existiert nicht (verifiziert: `find internal -name clip` → leer), Yank/OSC52 ist E5-Scope
(design-spec.md §12). `keyReviewCockpit` hat schlicht KEINEN `y`-Case — ein Druck auf
`y` fällt durch den Full-Capture-Handler als No-op durch (gleiche „handled-but-silent"-
Konvention wie E3 Task 1s Stub-Keys, KEIN „coming soon"-Text — Wegwerf-Arbeit,
E5 folgt unmittelbar in der Epen-Reihenfolge).

**(h) Capture-Order-Erweiterung (NEU, kritisch — löst einen echten Tastenkollisions-
Konflikt).** `design-spec.md` §7 deklariert `a`/`x`/`o` als „View-lokale Overrides" der
Review-Cockpit — aber `keyNodeAction` (E3, update.go) matched HEUTE bereits `a`
(`keys.Assign`, Parent-Picker) UNCONDITIONAL vor jedem View-Dispatch. Ohne Eingriff
würde `a` in der Cockpit fälschlich den Parent-Picker öffnen statt „Pass" zu feuern.
Lösung: ZWEI neue `if`-Blöcke in `handleKey` (update.go), eingefügt in der exakt
folgenden Reihenfolge (Kommentar im Code verweist auf diese Design-Entscheidung):

```
confirmQuit → searchActive → filterOpen → m.form!=nil → m.overlay!=overlayNone →
[T1] m.paletteOpen → keyPalette                                    (voller Capture, mirrors filterOpen)
[T1] keys.Palette gedrückt (ctrl+k/K) → openPalette                (GLOBAL — von jeder View aus erreichbar,
                                                                      auch aus der Cockpit, s.u. — deshalb VOR
                                                                      dem Cockpit-Capture-Block eingefügt)
[T3] m.view == viewReviewCockpit → keyReviewCockpit                (voller Capture NUR für diese View —
                                                                      überschreibt a/t/B/c/d/e für die Dauer
                                                                      des Cockpit-Aufenthalts vollständig;
                                                                      s/t/B/c/d/e sind in v1 in der Cockpit
                                                                      schlicht NICHT erreichbar, bewusster
                                                                      Scope-Cut — Feldbearbeitung eines
                                                                      Beans mitten im Review läuft über
                                                                      Tree/Backlog, nicht über die Cockpit)
globaler switch: ctrl+c / q / tab                                   (unverändert)
Refresh → keyNodeAction → detailFocus / keyBacklog / keyTree        (unverändert)
```

T1 fügt NUR die ersten beiden neuen Blöcke ein (Palette existiert, Review-Cockpit noch
nicht); T3 fügt den DRITTEN Block darunter ein. Die Reihenfolge stellt sicher, dass
`ctrl+k` aus JEDER View inkl. der Cockpit funktioniert (design-spec: „ctrl+k … von
überall"), während `a`/`x`/`o` NUR innerhalb der Cockpit ihre Override-Bedeutung
bekommen — außerhalb bleiben sie unverändert Assign/Toggle/unbelegt.

**(i) Review-Cockpit Detail-Pane: EIGENE Accordion-Cursor-Felder, KEIN Reuse von
`m.secCursor`/`m.accOpen`.** Die Cockpit zeigt rechts eine rein LESBARE Vorschau
(`beanSections`+`renderAccordion`, E2, digit-jump `1…4`) — OHNE das Tree/Backlog-
Zwei-Ebenen-Fokus-System (`m.detailFocus`/`m.fieldCursor`/`keyDetailFocus`s
Beziehungs-Feld-Sprung), das die Cockpit nicht braucht. Ein NEUES `reviewAccOpen int`
(0 = alles zu, 1-4 = offene Sektion, identische Toggle-Semantik wie `keyDetailFocus`s
Digit-Case) hält dieses Rendering entkoppelt — ein künftiger Tree-Fokus-Umbau kann die
Cockpit nicht versehentlich mitreißen, und umgekehrt (gleiche Vorsicht wie I01s
Copy-on-Write-Doktrin, hier auf „geteilte vs. eigene Felder" angewendet statt auf Maps).

**(j) Summary „x of n": 1-basierte Cursor-Position im LIVE to-review-Flat (NICHT ein
Pass/Reject-Tally).** design-spec §6 V6 sagt wörtlich „Summary-Zeile x of n" — schlanker
als devds eigene reviewSummary (✓/✗/∙-Zähler), UND passend zur
Entitäten-Reduktion (D05): es gibt kein `review_status`-Feld, nur die Tag-Transition
selbst, also keinen Ort, um einen Pass/Reject-Tally session-übergreifend zu verankern,
ohne neuen Client-State zu erfinden. `n` = `len(reviewToReviewFlat)` (LIVE, jeden Reload
neu abgeleitet — ein Agent, der währenddessen weitere Beans auf `to-review` tagt, lässt
`n` einfach wachsen, korrektes statt überraschendes Verhalten). `x` = 1-basierte Position
des Cursors INNERHALB dieses Flats, solange der Cursor auf einem to-review-Item steht;
steht der Cursor auf einem Rework-Item, zeigt die Zeile stattdessen `Rework: <n>
offen` statt einer „x of n"-Zählung (kein sinnvolles „x of n" für die
Awareness-Sektion). Erfordert KEINEN neuen Session-State — reiner Ableitungswert bei
jedem Render.

---

## Geteilte Infrastruktur

### Model-Felder (types.go)

```go
// Command-Center (E4 Task 1, bean bt-jpgn, design decisions a/b/h): paletteOpen
// is a full-capture floating-overlay state, same precedent as filterOpen
// (handleKey capture order, design decision h). palQuery drives BOTH candidate
// pools (actions here in T1, beans in T2) -- ONE shared input, design decision
// b. palList is the cursor over the COMBINED already-filtered result list
// (palFiltered, rebuilt every keystroke).
paletteOpen bool
palQuery    string
palList     listState

// Command-Center bean-search half (E4 Task 2, bean bt-yo60, design decision
// b): palette-SCOPED copies of the Bleve staleness-guard triplet /'s own
// searchBleveIDs/searchBleveFor/searchBleveLoading already establish
// (types.go) -- kept SEPARATE so the palette can never clobber an active
// Tree/Backlog `/` search session, or vice versa.
palBleveIDs     map[string]bool
palBleveFor     string
palBleveLoading bool

// Review-Cockpit (E4 Task 3, bean bt-hxyo, design-spec §6 V6, design
// decisions c/i): reviewCursor indexes the FLAT ordered []*data.Bean the
// Cockpit currently shows -- every to-review bean (group-then-canonical
// order) followed by every rework bean; group headers are a RENDER-time-only
// concern (view_review_cockpit.go), never part of this index space, so
// up/down can never land on a non-actionable header row. reviewAccOpen is a
// DEDICATED digit-jump cursor for the Cockpit's read-only detail preview --
// deliberately NOT the Tree/Backlog's shared accOpen/secCursor (design
// decision i: those are entangled with detailFocus's two-level machine,
// which the always-read-only Cockpit preview does not have).
reviewCursor  int
reviewAccOpen int
```

(Task 4/Reject reuses `m.form`/`m.formKind`/`m.mutTarget` verbatim — no new fields.)

### Neue Datei `internal/data/review.go` (Task 3)

```go
// EpicAncestor walks b.Parent upward (visited-guarded against a Parent
// cycle, same defensive shape as CollectDescendants/expandAncestorsOf) until
// it finds the nearest ancestor with Type == "epic". ok=false covers both "no
// parent chain at all" and "chain exists but never passes through an epic"
// (e.g. a bean parented directly under a milestone) -- both collapse into the
// SAME "(kein Epic)" bucket at the call site (design decision c), so this
// function does not need to distinguish them itself.
func EpicAncestor(idx *Index, b *Bean) (epic *Bean, ok bool)
```

### Capture-Order (update.go, `handleKey`) — Design-Entscheidung h

T1 fügt zwei `if`-Blöcke zwischen `m.overlay != overlayNone` und dem globalen
`switch msg.String()` ein; T3 fügt einen dritten darunter ein (exakte Reihenfolge s.
Design-Entscheidung h oben).

---

## Task 1: Command-Center-Infra + Aktions-Dispatch (`bt-jpgn`)

Liefert `ctrl+k`/`K` Öffnen/Schließen/Navigieren/Filtern/Dispatch für die
kontextabhängige Aktionsliste — NOCH OHNE Bean-Suche (Task 2). Legt die
Capture-Order-Erweiterung und `composeOverlays`-Verdrahtung, von der Task 3 (View-Case)
und Task 2 (Bean-Pool) konsumieren.

**Files:**
- Create: `internal/tui/fuzzy.go` (`fuzzyMatch`)
- Create: `internal/tui/overlay_palette.go` (`paletteItem`, `paletteAction`,
  `paletteActions`, `palFiltered`, `openPalette`, `keyPalette`, `dispatchPalette`,
  `paletteBox`)
- Modify: `internal/tui/types.go` (`paletteOpen`/`palQuery`/`palList`)
- Modify: `internal/tui/update.go` (`handleKey`: zwei neue Capture-Blöcke,
  Design-Entscheidung h)
- Modify: `internal/tui/view_browse_repo.go` (`composeOverlays`: neuer Palette-Case)
- Modify: `internal/tui/keymap.go` (KEIN neues Binding — `keys.Palette` existiert
  bereits seit E1 Task 7, nur der `helpGroups`-Eintrag ist schon vorhanden, keine
  Änderung nötig — Grep vorab verifiziert: `keys.Palette` ist heute komplett unbelegt)
- Test: `internal/tui/fuzzy_test.go`, `internal/tui/overlay_palette_test.go`,
  `internal/tui/update_test.go` (erweitert: Capture-Order)

**Port-Referenzen:**
- `fuzzyMatch`: devd `overlay_palette.go:60-71` VERBATIM (Design-Entscheidung a).
- Aktionsliste + Wortwahl-Konvention `verb: label`: devd `overlay_palette.go:20-49`
  (DD2-185), NUR das Wording-Muster — die konkrete Liste ist beans-tui-eigen
  (Design-Entscheidung b).
- `dispatchPalette`-Switch-Form: devd `overlay_palette.go:144-190`, strukturell
  (Switch über eine `id string`), Inhalt komplett neu.
- `paletteBox`-Render (`"> " + query`, Separator, `menuList`): devd
  `overlay_palette.go:200-216` — 1:1 Muster, `modalPanel` statt devds Inline-Box.

**Schritte:**

- [ ] **Step 1: Failing test** — `fuzzy_test.go`:

```go
func TestFuzzyMatchSubsequenceCaseInsensitive(t *testing.T) {
    cases := []struct{ query, target string; want bool }{
        {"", "anything", true},
        {"crb", "create: bean", true},
        {"CRB", "create: bean", true}, // case-insensitiv
        {"xyz", "create: bean", false},
        {"tsk1", "tsk1", true},
    }
    for _, c := range cases {
        if got := fuzzyMatch(c.query, c.target); got != c.want {
            t.Errorf("fuzzyMatch(%q, %q) = %v, want %v", c.query, c.target, got, c.want)
        }
    }
}
```

- [ ] **Step 2:** `command go test ./internal/tui/ -run TestFuzzyMatch` → FAIL
  (undefined).
- [ ] **Step 3: Implement** — `fuzzy.go`, `fuzzyMatch` Port devd verbatim (nur Paketname
  angepasst, keine Logikänderung).
- [ ] **Step 4:** PASS.
- [ ] **Step 5 (Aktionsliste + Filter): Failing tests** — `overlay_palette_test.go`:

```go
func TestPaletteActionsBeanContextFirst(t *testing.T) {
    // fixtureModel mit cursorID auf einem Task-Bean: paletteActions(m)[0].actionID
    // == "status" (erste Node-Aktion), globale Aktionen (z.B. "create") folgen NACH
    // allen Node-Aktionen.
}
func TestPaletteActionsNoFocusedBeanOmitsNodeActions(t *testing.T) {
    // cursorID auf orphanRootID (focusedBean() == nil): paletteActions(m) enthält
    // NUR globale Aktionen (create/go_backlog/go_browse/filter/search/reload).
}
func TestPalFilteredActionsFuzzyFiltered(t *testing.T) {
    // m.palQuery = "bckl": nur "go to: backlog" bleibt übrig (fuzzyMatch-Subsequence).
}
func TestPalFilteredEmptyQueryReturnsAllActionsNoBeans(t *testing.T) {
    // m.palQuery == "": alle Items sind kind==paletteKindAction (kein einziger
    // paletteKindBean-Eintrag -- die Bean-Hälfte landet erst in T2, aber der Guard
    // "leere Query -> keine Beans" wird HIER schon als Contract gepinnt, damit T2
    // ihn nicht versehentlich aufweicht).
}
```

- [ ] **Step 6:** `command go test ./internal/tui/ -run TestPalette` → FAIL.
- [ ] **Step 7: Implement** — `overlay_palette.go`:

```go
type paletteItemKind int

const (
    paletteKindAction paletteItemKind = iota
    paletteKindBean // T2 populates this kind; the type exists here upfront
                     // (mirrors E3 Task 1's "declare the full overlayID enum
                     // upfront" precedent, types.go) so T2 is a pure addition,
                     // never a signature change.
)

type paletteItem struct {
    kind     paletteItemKind
    actionID string     // kind == paletteKindAction
    bean     *data.Bean // kind == paletteKindBean (T2)
    label    string     // pre-rendered row text for both kinds
}

// paletteActions returns the context-aware action list (design decision b):
// focused-bean node actions FIRST (only when m.focusedBean() != nil), global
// actions after. T4 appends "go to: review cockpit" once that view exists
// (devd overlay_palette.go's own "T17/T18 ergänzen hier" precedent).
func paletteActions(m model) []paletteItem {
    var items []paletteItem
    if b := m.focusedBean(); b != nil {
        items = append(items,
            paletteItem{kind: paletteKindAction, actionID: "status", label: "status: setzen"},
            paletteItem{kind: paletteKindAction, actionID: "tags", label: "tags: zuweisen"},
            paletteItem{kind: paletteKindAction, actionID: "parent", label: "parent: zuweisen"},
            paletteItem{kind: paletteKindAction, actionID: "blocking", label: "blocking: zuweisen"},
            paletteItem{kind: paletteKindAction, actionID: "edit_title", label: "titel: bearbeiten"},
            paletteItem{kind: paletteKindAction, actionID: "delete", label: "bean: löschen"},
        )
    }
    items = append(items,
        paletteItem{kind: paletteKindAction, actionID: "create", label: "create: bean"},
        paletteItem{kind: paletteKindAction, actionID: "go_backlog", label: "go to: backlog"},
        paletteItem{kind: paletteKindAction, actionID: "go_browse", label: "go to: browse"},
        paletteItem{kind: paletteKindAction, actionID: "filter", label: "filter: facetten"},
        paletteItem{kind: paletteKindAction, actionID: "search", label: "search: beans"},
        paletteItem{kind: paletteKindAction, actionID: "reload", label: "reload: daten"},
        // T4 appends "go_review" ("go to: review cockpit") here.
    )
    return items
}

// palFiltered combines both candidate pools (T1: actions only; T2 adds the
// bean half below the actions, design decision b's ordering) filtered
// against m.palQuery.
func (m model) palFiltered() []paletteItem {
    var out []paletteItem
    for _, it := range paletteActions(m) {
        if fuzzyMatch(m.palQuery, it.label) {
            out = append(out, it)
        }
    }
    // T2 appends palFilteredBeans(m) here.
    return out
}

func (m model) openPalette() (tea.Model, tea.Cmd) {
    m.paletteOpen = true
    m.palQuery = ""
    m.palList = listState{}
    m.palList.setLen(len(m.palFiltered()))
    return m, nil
}

// keyPalette drives the open Command-Center: every rune/backspace edits
// palQuery (rebuilding palList's length every keystroke, T2 additionally
// dispatches a Bleve search when due); up/down move the cursor; enter
// dispatches the cursored item; esc closes without side effects.
func (m model) keyPalette(msg tea.KeyMsg) (tea.Model, tea.Cmd)

// dispatchPalette closes the palette and routes actionID to the matching
// handler -- action IDs mirror the existing single-key dispatch 1:1 (status
// -> m.openValueMenu(), etc.) so the Palette is a genuine second entry point
// to the SAME handlers, never a parallel implementation (US-04's "jede
// Aktion über die Command-Palette erreichbar").
func (m model) dispatchPalette(it paletteItem) (tea.Model, tea.Cmd) {
    m.paletteOpen = false
    switch it.kind {
    case paletteKindBean: // T2
        m.cursorID = it.bean.ID
        m.view = viewBrowseRepo
        return m, nil
    case paletteKindAction:
        switch it.actionID {
        case "status":
            return m.openValueMenu(), nil
        case "tags":
            return m.openTagPicker(), nil
        case "parent":
            return m.openParentPicker(), nil
        case "blocking":
            return m.openBlockingPicker(), nil
        case "edit_title":
            if b := m.focusedBean(); b != nil {
                return m.openEditTitleForm(b)
            }
        case "delete":
            return m.openDeleteConfirm(), nil
        case "create":
            return m.openCreateForm()
        case "go_backlog":
            m.view = viewBacklog
            m.backlogList.setLen(len(m.backlogVisible()))
            return m, nil
        case "go_browse":
            m.view = viewBrowseRepo
            return m, nil
        case "filter":
            return m.openFilterMenu()
        case "search":
            return m.openSearchInput()
        case "reload":
            return m, loadCmd(m.client)
        // "go_review" -- T4
        }
    }
    return m, nil
}

// paletteBox renders the floating Command-Center -- actions render PLAIN +
// theme.Header on select (mirrors box_menu_value.go/box_picker_tag.go: their
// row text carries no per-cell theming of its own); T2's bean rows instead
// use the D08 ansi.Strip+Accent-wrap convention (mirrors box_picker_parent.go)
// since relationRow output is ALREADY themed -- same split-styling rationale
// types.go's "Picker-Stil-Divergenz" doc-stamp already documents for the E3
// overlays, extended here to a THIRD file for the same reason.
func (m model) paletteBox() string
```

  **ERRATUM (B01, E4-T1-Review, bean bt-jpgn):** das `"create"`-Snippet oben
  (`case "create": return m.openCreateForm()`) fehlt der F1-Async-Gap-Guard
  (`m.pendingCreate != nil` → `createInFlightNote`, kein Form-Open) — der
  identische Guard, den `keyNodeAction`s Create-Case (`update.go`) und
  `submitForm`s `"create"`-Case (`box_confirm_create.go`) bereits tragen
  (types.go-Doc-Stamp). Ohne den Guard öffnet `ctrl+k` → „create: bean"
  während eines laufenden Creates ein ZWEITES Create-Form und
  kontaminiert die Single-Slot-Felder `createDraft`/`pendingCreate`. Die
  Implementierung (`overlay_palette.go`) trägt den Guard bereits als
  DRITTE gültige Stelle — dieser Plan-Snippet bleibt als historisches
  Artefakt unkorrigiert stehen, siehe stattdessen die Implementierung.

  Capture-Order (Design-Entscheidung h), `update.go` `handleKey`, eingefügt
  UNMITTELBAR nach dem bestehenden `if m.overlay != overlayNone { return
  m.keyOverlay(msg) }`-Block:

```go
// E4 Task 1 (bean bt-jpgn, design decision h): the Command-Center's own open
// state fully captures, same precedent as filterOpen/m.overlay above.
if m.paletteOpen {
    return m.keyPalette(msg)
}
// E4 Task 1 (design decision h): ctrl+k/K opens the palette from ANY view --
// checked here, ABOVE the (E4 Task 3) Review-Cockpit capture block below, so
// it also works from inside the Cockpit (design-spec §7: "von überall").
if keybind.Matches(msg, keys.Palette) {
    return m.openPalette()
}
```

  `composeOverlays` (view_browse_repo.go), neuer Case NACH dem `overlayID`-Switch,
  VOR `m.confirmQuit` (Painter's-Algorithmus-Reihenfolge unverändert, Palette wie
  Filter/Overlay/Form ein Zwischenlayer, Quit bleibt oberste Priorität):

```go
if m.paletteOpen {
    out = placeOverlay(out, m.paletteBox(), w, h)
}
```

- [ ] **Step 8:** `command go test ./internal/tui/` → PASS.
- [ ] **Step 9 (Dispatch-Roundtrip): Failing tests** — `overlay_palette_test.go`
  (erweitert):

```go
func TestKeyPaletteEnterDispatchesAndCloses(t *testing.T) {
    // palQuery "bckl", enter: m.paletteOpen == false, m.view == viewBacklog.
}
func TestKeyPaletteEscClosesNoSideEffect(t *testing.T)
func TestKeyPaletteRuneAppendsAndResyncsList(t *testing.T)
func TestKeyPaletteBackspaceTrims(t *testing.T)
func TestHandleKeyCtrlKOpensPaletteFromTree(t *testing.T)
func TestHandleKeyCtrlKUnreachableWhileFilterOpen(t *testing.T) {
    // m.filterOpen = true: ctrl+k muss NICHT die Palette öffnen (filterOpen
    // captured zuerst, Capture-Order-Contract).
}
```

- [ ] **Step 10:** `command go test ./internal/tui/` → PASS.
- [ ] **Step 11:** `command go test ./... && command gofmt -l . && command go vet ./...`
  clean; `command go build -o bin/bt .` ok. Manueller Smoke: `ctrl+k` im eigenen Repo,
  Tippen filtert Aktionen live, `enter` auf „go to: backlog" wechselt die View.
- [ ] **Step 12: Commit**

```
feat(tui): Command-Center-Infra (ctrl+k) + kontextabh. Aktions-Dispatch

Eigener Subsequence-Fuzzy-Matcher (Port devd overlay_palette.go, kein neuer
Dependency). Aktionsliste kontextabhängig: fokussierter Bean bringt seine
Node-Aktionen (status/tags/parent/blocking/edit/delete) vor die globalen
Einträge (create/go-to-backlog/browse/filter/search/reload) -- jede Aktion
dispatcht auf denselben Handler wie ihr Einzeltasten-Pendant, kein Parallel-
Pfad (US-04). Capture-Order-Erweiterung (paletteOpen, ctrl+k-Trigger) sitzt
VOR dem künftigen Review-Cockpit-Capture-Block (E4 Task 3), damit ctrl+k von
jeder View aus erreichbar bleibt.

Refs: bt-jpgn
```

---

## Task 2: Command-Center — Bean-Suche (`bt-yo60`)

Ergänzt den zweiten Kandidaten-Pool (Design-Entscheidung b): sobald `m.palQuery`
nicht leer ist, mischen sich Bean-Treffer unter die Aktionen. Reused
`beanMatchesSearch`s Bifurkations-Logik, aber auf Palette-skalierten Bleve-Feldern.

**Files:**
- Modify: `internal/tui/overlay_palette.go` (`palFilteredBeans`, `palFiltered`
  erweitert, `keyPalette` um Bleve-Dispatch, `paletteBox` um Bean-Zeilen)
- Modify: `internal/tui/types.go` (`palBleveIDs`/`palBleveFor`/`palBleveLoading`)
- Modify: `internal/tui/messages.go` (`paletteBleveResultMsg`, `paletteSearchCmd` —
  EIGENER Msg-Typ, NICHT `searchBleveResultMsg` wiederverwendet, Design-Entscheidung b)
- Modify: `internal/tui/update.go` (neuer `Update`-Case für `paletteBleveResultMsg`)
- Test: `internal/tui/overlay_palette_test.go` (erweitert)

**Port-Referenzen:**
- Bifurkations-Logik (lokal <3 Zeichen / Bleve-authoritativ ab 3): `beanMatchesSearch`,
  `view_browse_repo.go:211-220` — WIEDERVERWENDET über eine analoge, Palette-eigene
  Funktion (kein Code-Duplikat der IF-Kaskade selbst, s. Implement-Schritt).
  Staleness-Guard-Muster: `applyBleveResult`, `update.go:413-429`.

**Schritte:**

- [ ] **Step 1: Failing tests** — `overlay_palette_test.go`:

```go
func TestPalFilteredBeansEmptyQueryNone(t *testing.T) {
    // m.palQuery == "": palFiltered(m) enthält KEIN paletteKindBean-Item (Contract
    // aus T1s TestPalFilteredEmptyQueryReturnsAllActionsNoBeans bleibt gültig).
}
func TestPalFilteredBeansLocalSubstringBelowThreshold(t *testing.T) {
    // palQuery "tk" (< 3 Zeichen): lokaler ID/Title-Substring-Match, kein Bleve nötig.
}
func TestPalFilteredBeansCappedAt20(t *testing.T) {
    // 30 Fixture-Beans, alle matchend: len(bean-Items) == paletteBeanResultCap (20).
}
func TestPalFilteredBeansSortedCanonically(t *testing.T) {
    // data.SortBeans-Reihenfolge unter den zurückgegebenen Bean-Items.
}
func TestPalFilteredOrderActionsBeforeBeans(t *testing.T) {
    // palQuery matcht sowohl eine Aktion als auch einen Bean-Titel: alle
    // paletteKindAction-Items stehen VOR allen paletteKindBean-Items im Ergebnis.
}
func TestDispatchPaletteBeanJumpsCursorAndSwitchesToBrowse(t *testing.T) {
    // dispatchPalette auf einem paletteKindBean-Item: m.cursorID == bean.ID,
    // m.view == viewBrowseRepo (auch wenn zuvor m.view == viewBacklog war).
}
func TestKeyPaletteDispatchesBleveOnQueryGrowth(t *testing.T) {
    // Query wächst auf 3 Zeichen: zurückgegebenes tea.Cmd != nil (paletteSearchCmd),
    // m.palBleveLoading == true.
}
func TestApplyPaletteBleveResultDiscardsStaleQuery(t *testing.T) {
    // paletteBleveResultMsg{query: "alt", ...} nach m.palQuery == "neu": kein State-
    // Update (Staleness-Guard, gleiche Semantik wie applyBleveResult).
}
```

- [ ] **Step 2:** `command go test ./internal/tui/ -run TestPalFilteredBeans` → FAIL.
- [ ] **Step 3: Implement** — `overlay_palette.go`:

```go
const paletteBeanResultCap = 20

// palBeanMatches mirrors beanMatchesSearch's bifurcation (view_browse_repo.go)
// against the PALETTE's own Bleve fields instead of the Tree's, so opening
// ctrl+k never touches an active `/` search session (design decision b).
func (m model) palBeanMatches(b *data.Bean) bool {
    q := strings.ToLower(strings.TrimSpace(m.palQuery))
    if len(q) >= 3 && m.palBleveFor == m.palQuery {
        return m.palBleveIDs[b.ID] || strings.Contains(strings.ToLower(b.ID), q)
    }
    return strings.Contains(strings.ToLower(b.Title), q) || strings.Contains(strings.ToLower(b.ID), q)
}

// palFilteredBeans returns up to paletteBeanResultCap matching beans, empty
// query -> nil (design decision b: the palette never dumps the whole repo).
func (m model) palFilteredBeans() []paletteItem {
    if strings.TrimSpace(m.palQuery) == "" || m.idx == nil {
        return nil
    }
    var matched []*data.Bean
    for _, b := range m.idx.ByID {
        if m.palBeanMatches(b) {
            matched = append(matched, b)
        }
    }
    data.SortBeans(matched)
    if len(matched) > paletteBeanResultCap {
        matched = matched[:paletteBeanResultCap]
    }
    items := make([]paletteItem, len(matched))
    for i, b := range matched {
        items[i] = paletteItem{kind: paletteKindBean, bean: b, label: relationRow(b)}
    }
    return items
}
```

  `palFiltered` erweitert: `out = append(out, m.palFilteredBeans()...)` NACH der
  Actions-Schleife (Design-Entscheidung b: Aktionen zuerst). `keyPalette`s
  Rune-/Backspace-Zweige dispatchen zusätzlich `maybePaletteBleveCmd()` (Analog zu
  `maybeBleveCmd`, `update.go:859-865`, auf `m.palQuery`/`palBleveFor` gespiegelt).
  `messages.go`:

```go
type paletteBleveResultMsg struct {
    query string
    ids   []string
    err   error
}

func paletteSearchCmd(c *data.Client, query string) tea.Cmd {
    return func() tea.Msg {
        beans, err := c.Search(query)
        if err != nil {
            return paletteBleveResultMsg{query: query, err: err}
        }
        ids := make([]string, len(beans))
        for i, b := range beans {
            ids[i] = b.ID
        }
        return paletteBleveResultMsg{query: query, ids: ids}
    }
}
```

  `update.go`: neuer `case paletteBleveResultMsg:` Analog zu `applyBleveResult`
  (Staleness-Guard gegen `m.palQuery` statt `m.searchQuery`). `paletteBox` erweitert:
  Bean-Zeilen über den D08-Stil (`ansi.Strip` + `theme.Accent.Render("▌"+plain)`)
  gerendert (Design-Entscheidung b, Picker-Stil-Divergenz-Präzedenz), von den
  plain-Aktionszeilen optisch getrennt via einen `theme.Dim`-Separator, sobald
  beide Pools nicht-leer sind.
- [ ] **Step 4:** `command go test ./internal/tui/` → PASS.
- [ ] **Step 5:** `command go test ./... && command gofmt -l . && command go vet ./...`
  clean; `command go build -o bin/bt .` ok. Smoke: `ctrl+k`, 3+ Zeichen eines
  Bean-Titels tippen → Bean-Treffer erscheinen unter den Aktionen, `enter` springt den
  Tree-Cursor drauf.
- [ ] **Step 6: Commit**

```
feat(tui): Command-Center Bean-Suche (Bleve, palette-skaliert)

Zweiter Kandidaten-Pool im selben Eingabefeld (design-spec §6 V5), gedeckelt
auf 20 Treffer, kanonisch sortiert, IMMER nach den Aktionen. Palette-eigene
Bleve-Staleness-Guard-Kopie (palBleveIDs/-For/-Loading, eigener
paletteBleveResultMsg) -- bewusst NICHT die Tree-/Backlog-Suchfelder
wiederverwendet, damit ctrl+k eine aktive `/`-Session nie überschreibt.

Refs: bt-yo60
```

---

## Task 3: Review-Queue-Ableitung + Cockpit-Skeleton (`bt-hxyo`)

Liefert die Datenableitung (Design-Entscheidung c) und eine VOLLSTÄNDIG navigierbare,
aber noch rein lesbare Cockpit-View: Queue nach Epic gruppiert, Rework-Sektion,
Verdikt-Dots, Master-Detail, Digit-Accordion, `R`-Einstieg, `esc`/`q` zurück. Verdikt-
AKTIONEN (`a`/`x`/`o`) kommen in Task 4.

**Files:**
- Create: `internal/data/review.go` (`EpicAncestor`)
- Create: `internal/tui/view_review_cockpit.go` (`reviewGroup`, `reviewQueue`,
  `reviewRework`, `reviewFlat`, `reviewDot`, `reviewQueueRows`, `viewReviewCockpit`,
  `openReviewCockpit`, `keyReviewCockpit`, `reviewSummaryLine`)
- Modify: `internal/tui/types.go` (`viewReviewCockpit` viewID, `reviewCursor`/
  `reviewAccOpen`)
- Modify: `internal/tui/update.go` (`View()`-Switch neuer Case, `handleKey` dritter
  Capture-Block, Design-Entscheidung h)
- Modify: `internal/tui/view_browse_repo.go` (`keyTree`: neuer `keys.Reviews`-Case)
- Modify: `internal/tui/view_browse_backlog.go` (`keyBacklog`: neuer
  `keys.Reviews`-Case)
- Test: `internal/data/review_test.go`, `internal/tui/view_review_cockpit_test.go`

**Port-Referenzen:**
- Zyklen-geschützter Ancestor-Walk: eigenes `expandAncestorsOf`-Muster (update.go:
  663-680) / `CollectDescendants` (hierarchy.go) — NICHT devd (devd hat keine
  Epic-Ancestor-Suche, Sprints sind flach).
- Master-Detail-Layout-Idee (Queue links mit Verdikt-Indikator, Detail rechts als
  Accordion-Preview): devd `view_review_sprint.go:viewReviewSprint/
  reviewMasterPane/reviewDetailPane` — NUR das Layout-Muster (borderedPane,
  windowAround um den Cursor, Header+Divider+Body), KEINE Sprint-Kopplung.
- Verdikt-Dot-Farbcodierung: devd `keys_review.go:verdictDot` (Grün/Rot/Peach) —
  hier auf 2 Zustände reduziert (Peach=pending/to-review, Rot=rework, Design-
  Entscheidung c/j — kein Grün, da „passed" das Item aus JEDER sichtbaren Sektion
  entfernt, Design-Entscheidung j).

**Schritte:**

- [ ] **Step 1: Failing tests** — `internal/data/review_test.go`:

```go
func TestEpicAncestorFindsNearestEpic(t *testing.T) {
    // ms -> ep -> tsk: EpicAncestor(idx, tsk) == (ep, true).
}
func TestEpicAncestorNoneWhenMilestoneDirectParent(t *testing.T) {
    // ms -> tsk (kein Epic dazwischen): EpicAncestor(idx, tsk) == (nil, false).
}
func TestEpicAncestorNoneWhenParentless(t *testing.T) {
    // tsk ohne Parent: (nil, false).
}
func TestEpicAncestorSurvivesParentCycle(t *testing.T) {
    // a -> b -> a: terminiert, (nil, false) -- kein Hang.
}
```

- [ ] **Step 2:** `command go test ./internal/data/ -run TestEpicAncestor` → FAIL.
- [ ] **Step 3: Implement** — `review.go`, `EpicAncestor` (visited-Map-Walk, Port-Muster
  `expandAncestorsOf`).
- [ ] **Step 4:** PASS.
- [ ] **Step 5 (Queue-Ableitung): Failing tests** — `view_review_cockpit_test.go`:

```go
func TestReviewQueueGroupsByEpicCanonicalOrder(t *testing.T) {
    // 2 Epics je mit to-review-Kindern + 1 parentloser to-review-Bean:
    // reviewQueue(idx) hat 3 Gruppen, "(kein Epic)" IMMER letzte Gruppe,
    // Epic-Gruppen in data.SortBeans-Reihenfolge der Epics selbst.
}
func TestReviewQueueGroupMembersSortedCanonically(t *testing.T)
func TestReviewQueueEmptyWhenNoToReviewTag(t *testing.T)
func TestReviewReworkFlatSortedNoGrouping(t *testing.T) {
    // WithTag("rework") -- eine flache, kanonisch sortierte Liste, KEINE
    // Epic-Gruppierung (Design-Entscheidung c).
}
func TestReviewFlatOrderToReviewThenRework(t *testing.T) {
    // reviewFlat(idx): erst alle to-review-Beans (Gruppen-Reihenfolge), dann alle
    // rework-Beans -- Cursor-Indexraum aus dem Model-Feld-Kommentar (types.go).
}
```

- [ ] **Step 6:** `command go test ./internal/tui/ -run TestReview` → FAIL.
- [ ] **Step 7: Implement** — `view_review_cockpit.go`:

```go
type reviewGroup struct {
    epic  *data.Bean // nil for the "(kein Epic)" bucket
    beans []*data.Bean
}

// reviewQueue groups idx.WithTag("to-review") by EpicAncestor (design
// decision c): a real-epic group per epic bean (epics themselves canonically
// sorted, data.SortBeans), a "(kein Epic)" bucket (epic == nil) ALWAYS last.
func reviewQueue(idx *data.Index) []reviewGroup

// reviewRework returns idx.WithTag("rework"), flat, canonically sorted --
// deliberately NOT epic-grouped (design decision c: a secondary awareness
// list, not the primary reviewed queue).
func reviewRework(idx *data.Index) []*data.Bean

// reviewFlat is the Cockpit's cursor index space (types.go's reviewCursor
// doc-stamp): every to-review bean in reviewQueue's group order, THEN every
// reviewRework bean.
func reviewFlat(idx *data.Index) []*data.Bean

// reviewDot renders the Verdikt-Dot: Peach = pending (to-review, unverdikted
// by construction -- design decision j, a "passed" item leaves every visible
// section), Red = rework (rejected, awaiting agent nacharbeit).
func reviewDot(inRework bool) string
```

  `reviewQueueRows(idx, cursorFlat int, focused bool, bodyH int) []string`: baut die
  Zeilenliste (Epic-Header via `theme.Muted.Render(relationRow(epic))`, „(kein Epic)"
  via `theme.Dim`, Rework-Trenner via `theme.Dim.Render("── Rework ──")`, Bean-Zeilen
  via `reviewDot(...)+" "+relationRow(b)` mit D08-Cursor-Behandlung an der
  `cursorFlat`-Position, ansonsten identisch zu `treeRows`/`backlogRows`s Muster),
  `windowAround` am Ende.

  `openReviewCockpit() (tea.Model, tea.Cmd)`:

```go
func (m model) openReviewCockpit() (tea.Model, tea.Cmd) {
    m.view = viewReviewCockpit
    m.reviewCursor = 0
    m.reviewAccOpen = 0
    return m, nil
}
```

  `keyReviewCockpit(msg tea.KeyMsg) (tea.Model, tea.Cmd)`:

```go
// keyReviewCockpit fully captures the Review-Cockpit (design decision h) --
// a/x/o land in Task 4; this task wires navigation only, unmatched keys are a
// silent no-op (same "handled-but-stub" convention as E3 Task 1's stub keys).
func (m model) keyReviewCockpit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    flat := reviewFlat(m.idx)
    switch {
    case keybind.Matches(msg, keys.Back), msg.String() == "q":
        m.view = viewBrowseRepo
        return m, nil
    }
    if s := msg.String(); len(s) == 1 && s[0] >= '1' && s[0] <= '4' {
        d := int(s[0] - '0')
        if m.reviewAccOpen == d {
            m.reviewAccOpen = 0
        } else {
            m.reviewAccOpen = d
        }
        return m, nil
    }
    switch navKey(msg.String()) {
    case "up":
        if m.reviewCursor > 0 {
            m.reviewCursor--
        }
        return m, nil
    case "down":
        if m.reviewCursor < len(flat)-1 {
            m.reviewCursor++
        }
        return m, nil
    }
    switch msg.String() {
    case "n": // explizites next (design-spec §7), Alias von down
        if m.reviewCursor < len(flat)-1 {
            m.reviewCursor++
        }
    case "p": // explizites prev, Alias von up
        if m.reviewCursor > 0 {
            m.reviewCursor--
        }
    }
    return m, nil // a/x/o -- Task 4
}
```

  `viewReviewCockpit() string`: mirrors `viewBrowseRepo`/`viewBacklog`s Algebra
  (`masterDetailWidths`, `renderPane`/`borderedPane`, `outerBorder`, `composeOverlays`)
  — links `reviewQueueRows`, rechts `renderBeanAccordionPane(flat[reviewCursor], rw,
  bodyH, true)` (immer `focused=true` — die Cockpit hat kein Fokus-Tausch, Design-
  Entscheidung i) mit `reviewAccOpen` statt `m.accOpen`. Footer/Head: „(keine offenen
  Reviews)"-Placeholder wenn `len(flat) == 0` (Port devd `view_navigate_reviews.go:20`
  Textmuster). `Refresh`(`ctrl+r`) bleibt VOR dem neuen Capture-Block erreichbar
  (unveränderte Position in `handleKey`, s.u.), Watch-Reload funktioniert also
  unverändert.

  `View()`-Switch (update.go) neuer Case:

```go
case viewReviewCockpit:
    return m.viewReviewCockpit()
```

  Capture-Order (Design-Entscheidung h), `handleKey`, eingefügt UNMITTELBAR NACH Task
  1s `keys.Palette`-Trigger-Block:

```go
// E4 Task 3 (bean bt-hxyo, design decision h): full capture for the Cockpit
// -- overrides a/t/B/c/d/e (keyNodeAction, below) with the view-local
// meaning design-spec §7 assigns them here (a/x land in Task 4). s/t/B/c/d/e
// are simply unreachable in v1 (bean field edits mid-review route through
// Tree/Backlog instead, a documented scope cut).
if m.view == viewReviewCockpit {
    return m.keyReviewCockpit(msg)
}
```

  `view_browse_repo.go` `keyTree`, neuer Case (VOR `keybind.Matches(msg,
  keys.Backlog)`, gleiche Position wie andere View-Wechsel-Keys):

```go
case keybind.Matches(msg, keys.Reviews):
    return m.openReviewCockpit()
```

  `view_browse_backlog.go` `keyBacklog`, identischer Case.
- [ ] **Step 8:** `command go test ./internal/tui/` → PASS.
- [ ] **Step 9 (Golden):** `TestReviewCockpitGolden` 100×30 gegen
  `testdata/review_cockpit.golden` (2 Epic-Gruppen + 1 kein-Epic-Bean + 1
  Rework-Bean, Fixture) — `-update`-Flag erzeugt, dann Re-Run ohne Flag.
- [ ] **Step 10:** `command go test ./... && command gofmt -l . && command go vet
  ./...` clean; `command go build -o bin/bt .` ok. Manueller Smoke: eigenes Repo, `R`
  drücken → Queue (leer, falls kein `to-review`-Tag aktiv — dann testweise
  `beans update <id> --tag to-review` auf einem Task setzen, Cockpit zeigt es unter
  seinem Epic), `n`/`p`/`↑↓` navigiert, Digit `2` klappt Body auf, `esc` zurück zu
  Tree, `ctrl+k` funktioniert AUCH innerhalb der Cockpit (Capture-Order-Beleg).
- [ ] **Step 11: Commit**

```
feat(tui): Review-Queue-Ableitung + Cockpit-Skeleton (R, lesbar)

EpicAncestor (zyklen-geschützter Parent-Walk, data-Paket) + reviewQueue/
reviewRework/reviewFlat (tui-Paket): Queue = WithTag(to-review) gruppiert
nach nächstem Epic-Vorfahren, "(kein Epic)" immer letzte Gruppe; eine zweite,
flache Rework-Sektion (WithTag(rework)) für PO-Sichtbarkeit (Design-
Erweiterung, Verdikt-Aktionen dafür folgen in Task 4). Master-Detail-Layout
(Queue links inkl. Verdikt-Dots, read-only Accordion-Preview rechts,
eigener reviewAccOpen statt Tree/Backlog-Fokusmaschine). Capture-Order:
Review-Cockpit überschreibt a/t/B/c/d/e für ihre Aufenthaltsdauer vollständig
(Design-Entscheidung h) -- ctrl+k (Task 1) bleibt davon unberührt erreichbar.

Refs: bt-hxyo
```

---

## Task 4: Verdikt-Aktionen — Pass/Reject/Reopen (`bt-yy6w`)

Verdrahtet `a`/`x`/`o` in die Task-3-Skeleton-Cockpit, ergänzt die zwei kombinierten
Mutations-Wrapper (Design-Entscheidung d), die Reject-Kommentar-Form (Design-
Entscheidung e) und die Palette-Erweiterung um „go to: review cockpit" (Task 1s
Vorgriffs-Kommentar).

**Files:**
- Modify: `internal/data/mutations.go` (`PassReview`, `RejectReview`)
- Create: `internal/tui/form_reject_review.go` (`buildRejectReviewForm`)
- Modify: `internal/tui/forms_shared.go` (`formTitle`-Switch: `"reject"`-Case)
- Modify: `internal/tui/box_confirm_create.go` (`submitForm`-Switch: `"reject"`-Case —
  ODER Umzug des ganzen Switches wird NICHT gemacht, nur ein Case ergänzt; der
  bestehende Datei-Ort bleibt trotz des inzwischen etwas irreführenden Dateinamens
  unverändert, konsistent mit „nur explizite Pfade selbst geänderter Dateien" statt
  einer unangeforderten Umbenennung)
- Modify: `internal/tui/view_review_cockpit.go` (`keyReviewCockpit`: `a`/`x`/`o`-Cases,
  `reviewSummaryLine`, `openRejectForm`)
- Modify: `internal/tui/overlay_palette.go` (`paletteActions`: `go_review`-Eintrag,
  `dispatchPalette`: `go_review`-Case)
- Test: `internal/data/client_mut_test.go` (erweitert), `internal/tui/
  view_review_cockpit_test.go` (erweitert), `internal/tui/form_reject_review_test.go`,
  `internal/tui/overlay_palette_test.go` (erweitert)

**Port-Referenzen:**
- Kombinierte Mutations-Semantik: E3 Task 2/3s `SetTags`/`SetBlocking`-Rationale
  (mutations.go-Doc-Stamps) — WÖRTLICH auf `PassReview`/`RejectReview` übertragen
  (Design-Entscheidung d).
- Reject-Form-Hosting: E3 Task 5s `form_edit_title.go`/`"editTitle"`-Muster (Design-
  Entscheidung e: kein Confirm-Gate, direktes Feuern).
- Reopen-Mutation: E3 Task 2s `SetTags` (Design-Entscheidung f), unverändert
  wiederverwendet.

**Schritte:**

- [ ] **Step 1: Failing tests (Datenlayer)** — `client_mut_test.go`
  (`requireBeansBinary`/`newTestRepo`-Muster wie `TestSetBodyReplacesWholeBody`):

```go
func TestPassReviewSetsCompletedAndRemovesTag(t *testing.T) {
    // newTestRepo, Bean mit Tag "to-review" + Status "in-progress"; PassReview(id, etag);
    // List/Show: Status == "completed", Tags enthält "to-review" NICHT mehr.
}
func TestRejectReviewSwapsTagAndAppendsSection(t *testing.T) {
    // RejectReview(id, "bitte X korrigieren", "2026-07-15", etag); List/Show: Tags
    // enthält "rework" statt "to-review", Body endet auf
    // "## Review 2026-07-15\n\nbitte X korrigieren\n".
}
func TestPassReviewConflictOnStaleEtag(t *testing.T) // ErrConflict via errors.Is
func TestRejectReviewConflictOnStaleEtag(t *testing.T)
```

- [ ] **Step 2:** `command go test ./internal/data/ -run TestPassReview|TestRejectReview`
  → FAIL.
- [ ] **Step 3: Implement** — `mutations.go`, `PassReview`/`RejectReview` (Design-
  Entscheidung d, Code-Snippet oben).
- [ ] **Step 4:** PASS.
- [ ] **Step 5 (Reject-Form): Failing tests** — `form_reject_review_test.go`:

```go
func TestBuildRejectReviewFormRequiresComment(t *testing.T) {
    // leerer Submit-Versuch -> huh-Validierungsfehler (nonEmpty), Form bleibt offen.
}
func TestRejectSubmitFiresRejectReviewDirectlyNoConfirm(t *testing.T) {
    // formKind "reject", ausgefülltes Comment-Feld, StateCompleted ->
    // submitForm: overlay unverändert overlayNone, m.form == nil (kein Confirm-Gate,
    // Design-Entscheidung e), zurückgegebenes Cmd triggert RejectReview via
    // mutateCmd-Fehlermeldung-Assertion (gleiche "kein Mock, echter Client-
    // Fehlertext" Verifikationsart wie E3s TestValueMenuEnterAppliesCursoredValue
    // AndCloses).
}
func TestRejectSubmitVanishedTargetSurfacesError(t *testing.T) {
    // beanETag(id) ok==false: m.err gesetzt, KEIN Cmd (gleicher Guard wie jede
    // andere E3-Overlay-Selection).
}
```

- [ ] **Step 6:** `command go test ./internal/tui/ -run TestReject` → FAIL.
- [ ] **Step 7: Implement** — `form_reject_review.go`:

```go
// buildRejectReviewForm builds the single-field ("comment") reject form --
// required (nonEmpty, design decision e: an unexplained rejection defeats the
// agent-feedback purpose the ## Review section exists for).
func buildRejectReviewForm() *huh.Form {
    return huh.NewForm(huh.NewGroup(
        huh.NewText().Key("comment").Title("Reject-Kommentar").Validate(nonEmpty),
    ))
}

func (m model) openRejectForm(b *data.Bean) (tea.Model, tea.Cmd) {
    m.mutTarget = b.ID
    m.form = m.styleForm(buildRejectReviewForm())
    m.formKind = "reject"
    return m, m.form.Init()
}
```

  `forms_shared.go` `formTitle`: `case "reject": return "Review ablehnen"`.
  `box_confirm_create.go` `submitForm`, neuer Case (mirrors `"editTitle"`s
  Direkt-Feuer-Struktur):

```go
case "reject":
    comment := m.form.GetString("comment")
    id := m.mutTarget
    m.form = nil
    m.formKind = ""
    etag, ok := m.beanETag(id)
    if !ok {
        m.err = "Bean nicht mehr vorhanden — Ablehnung verworfen"
        return m, nil
    }
    client := m.client
    date := time.Now().Format("2006-01-02")
    return m, mutateCmd(func() error { return client.RejectReview(id, comment, date, etag) })
```

- [ ] **Step 8:** `command go test ./internal/tui/` → PASS.
- [ ] **Step 9 (Cockpit-Verdrahtung): Failing tests** — `view_review_cockpit_test.go`:

```go
func TestKeyReviewCockpitPassFiresPassReview(t *testing.T) {
    // Cursor auf einem to-review-Bean, "a": zurückgegebenes Cmd feuert PassReview
    // (Fehlertext-Assertion gegen den echten Client, wie E3-Muster).
}
func TestKeyReviewCockpitRejectOpensCommentForm(t *testing.T) {
    // "x": m.form != nil, m.formKind == "reject", m.mutTarget == fokussierte ID.
}
func TestKeyReviewCockpitReopenOnReworkFiresSetTags(t *testing.T) {
    // Cursor auf einem Rework-Bean (reviewCursor jenseits der to-review-Flat-Länge),
    // "o": Cmd feuert SetTags(id, [to-review], [rework], etag).
}
func TestKeyReviewCockpitReopenOnToReviewIsNoop(t *testing.T) {
    // Cursor auf to-review-Bean, "o": kein Cmd, kein State-Change (Design-
    // Entscheidung f, "bereits im Zielzustand").
}
func TestReviewSummaryLinePositionInToReview(t *testing.T) {
    // 5 to-review-Beans, Cursor auf Index 2 (0-based): reviewSummaryLine(...) ==
    // "3 of 5".
}
func TestReviewSummaryLineReworkContext(t *testing.T) {
    // Cursor auf einem Rework-Item: reviewSummaryLine(...) enthält "Rework:" statt
    // "of".
}
```

- [ ] **Step 10:** `command go test ./internal/tui/ -run TestKeyReviewCockpit|
  TestReviewSummary` → FAIL.
- [ ] **Step 11: Implement** — `view_review_cockpit.go` `keyReviewCockpit` erweitert
  (Design-Entscheidung f/j):

```go
case msg.String() == "a": // pass
    if b := reviewFocused(flat, m.reviewCursor); b != nil && !reviewIsRework(idx, b) {
        etag, ok := m.beanETag(b.ID)
        if !ok { m.err = "Bean nicht mehr vorhanden — Verdikt verworfen"; return m, nil }
        id, client := b.ID, m.client
        return m, mutateCmd(func() error { return client.PassReview(id, etag) })
    }
case msg.String() == "x": // reject
    if b := reviewFocused(flat, m.reviewCursor); b != nil && !reviewIsRework(idx, b) {
        return m.openRejectForm(b)
    }
case msg.String() == "o": // reopen -- design decision f: only meaningful on Rework
    if b := reviewFocused(flat, m.reviewCursor); b != nil && reviewIsRework(idx, b) {
        etag, ok := m.beanETag(b.ID)
        if !ok { m.err = "Bean nicht mehr vorhanden"; return m, nil }
        id, client := b.ID, m.client
        return m, mutateCmd(func() error { return client.SetTags(id, []string{"to-review"}, []string{"rework"}, etag) })
    }
```

  (`reviewFocused`/`reviewIsRework` kleine Helfer: Index in `flat` gegen die
  `len(reviewToReviewFlat)`-Grenze prüfen, um zu wissen, ob der Cursor gerade auf
  einem to-review- oder Rework-Item steht — EINE Grenzberechnung, von `a`/`x`/`o`
  UND `reviewSummaryLine` konsumiert, kein dreifach dupliziertes Cutoff.)
  `reviewSummaryLine(idx *data.Index, cursor int) string` (Design-Entscheidung j).
  `overlay_palette.go`: `paletteActions` bekommt den `go_review`-Eintrag NACH
  `go_browse` angehängt (`{kind: paletteKindAction, actionID: "go_review", label:
  "go to: review cockpit"}`), `dispatchPalette` bekommt `case "go_review": return
  m.openReviewCockpit()`.
- [ ] **Step 12:** `command go test ./internal/tui/` → PASS.
- [ ] **Step 13:** `command go test ./... -count=1` (2×) grün, `command gofmt -l .`
  leer, `command go vet ./...` leer, `command go build -o bin/bt .` ok. Manueller
  Dogfooding-Smoke (tmux): eigenes Repo, ein Task-Bean testweise `to-review` taggen,
  `R` → Cockpit zeigt es unter seinem Epic mit Peach-Dot, „x of n" korrekt; `x`
  → Kommentar-Form → submit → `.md` zeigt `## Review <heute>`-Abschnitt + Tag
  `rework` statt `to-review`; `R` erneut → Bean jetzt in der Rework-Sektion mit
  Rot-Dot; `o` darauf → Tag zurück zu `to-review`, Bean wieder oben in seiner
  Epic-Gruppe; `a` auf einem anderen to-review-Bean → Status `completed`, Tag weg,
  Bean verschwindet aus der Queue. `ctrl+k` → „go to: review cockpit" erscheint
  und funktioniert. Terminal-Ausschnitt als Beleg in den Commit-Body.
- [ ] **Step 14: Commit**

```
feat(tui): Review-Verdikte (a pass / x reject / o reopen)

PassReview/RejectReview als je EIN kombinierter beans-update-Aufruf
(Status+Tag bzw. Tag-Swap+Body-Append, gegen `beans update --help`
verifiziert) -- gleiche Ein-Kommando-Diff-Lehre wie E3s SetTags/SetBlocking,
vermeidet eine ETag-Konflikt-Kaskade. Reject-Kommentar-Form (huh, Pflichtfeld,
kein Confirm-Gate -- Reuse der E3-Form-Infra) hängt den Kommentar als
`## Review <datum>`-Abschnitt an den Body. Reopen (o) ist auf die
Rework-Sektion beschränkt (design-spec §5 definiert es nicht explizit --
hergeleitet: PO holt ein abgelehntes Bean manuell zurück in die Queue, ohne
auf den Agenten zu warten) und wiederverwendet SetTags unverändert. Palette
(Task 1) bekommt "go to: review cockpit" ergänzt.

Refs: bt-yy6w
```

---

## Task 5: E4-Abschluss (`bt-v7ti`)

Mirrors E3 Task 7 (implementation-plan.md »Epos-Rituale« → Epos-Abschluss).

- [ ] `command go test ./... -count=1` grün (2× hintereinander), `command go build -o
  bin/bt .` ok, `command gofmt -l .` leer, `command go vet ./...` leer.
- [ ] Manueller Dogfooding-Smoke (tmux), zusätzlich zu Task 4s Beleg: `ctrl+k` aus der
  Review-Cockpit heraus öffnet die Palette (Capture-Order-Beleg, Design-Entscheidung
  h); `ctrl+k` während `filterOpen`/`m.form != nil`/Overlay offen öffnet die Palette
  NICHT (Capture-Order-Negativ-Beleg); Bean-Suche in der Palette (≥3 Zeichen) findet
  einen Titel-Treffer außerhalb des aktuell sichtbaren Tree-Ausschnitts, `enter`
  springt korrekt hin. Terminal-Ausschnitt als Beleg in den Commit-Body.
- [ ] beans pflegen: `beans update bt-jpgn -s completed` … `bt-yy6w -s completed`
  (T1-T4, agent-abschließbar), Epic `beans update bt-tfqi --tag to-review` (**NICHT**
  `-s completed` — PO-Gate, „der ausführende Agent schließt NICHT").
- [ ] Selbst-Review im Commit-Body:
  - Spec-Coverage: V5 komplett (fuzzy Aktionen+Bean-Suche gemischt, kontextabhängig
    zuerst) ✓ · V6 komplett (Queue gruppiert nach Epic, Verdikt-Dots, Summary „x of n",
    a/x/o) ✓ · US-04/US-08 ✓ · §5-Review-Tag-Konvention vollständig verdrahtet
    (Agent-Schritte unverändert, PO-Schritte alle drei über die TUI) ✓.
  - Bewusste Scope-Cuts dokumentiert: Yank → E5 (`internal/clip` existiert nicht) ·
    Repo-Picker/Lobby → E5 · Help-Overlay → E5 (Palette referenziert keine
    Help-Aktion) · s/t/B/c/d/e in der Review-Cockpit unerreichbar (Design-Entscheidung
    h, Feldbearbeitung läuft über Tree/Backlog) · keine Score-basierte Fuzzy-Rangierung
    (Design-Entscheidung a, `sahilm/fuzzy` bewusst nicht eingeführt).
  - Design-Entscheidungen ohne direkte Spec-Vorlage, explizit hergeleitet und
    begründet: Reopen-Semantik (f, §5-Tabelle hat keine Reopen-Zeile) · Summary
    „x of n" als Cursor-Position statt Pass/Reject-Tally (j, D05-Entitäten-Reduktions-
    konform) · Rework-Sichtbarkeits-Sektion (c, Erweiterung ggü. dem wörtlichen „Queue
    to-review"-Wortlaut, aber notwendig, damit `o` überhaupt einen Wirkungsort hat).
  - Konsolidierung ggü. Quellen: EIN kombinierter `beans update`-Aufruf je
    Verdikt-Mutation statt devds getrennter Tag-/Status-/Body-Calls · eigener
    Fuzzy-Matcher statt neuer Dependency · Review-Cockpit-Layout wiederverwendet
    `masterDetailWidths`/`renderPane`/`windowAround`/`renderBeanAccordionPane`
    (E1/E2) statt eigener Pane-Mathematik.
  - Capture-Order: ZWEI neue Blöcke (Palette, Review-Cockpit) an präzise begründeter
    Position (Design-Entscheidung h) — `TestHandleKeyCtrlKUnreachableWhileFilterOpen`
    + die Cockpit-internen a/x/o-Tests belegen die Kollisionsfreiheit gegenüber
    `keyNodeAction`s bestehenden Bindungen.
- [ ] Commit `docs: E4-Abschluss` (Refs: bt-v7ti).
- [ ] Skill `ce-nsp-auto` → Handover-Prompt für E5 (Polish: Toast, Help-Overlay, Yank/
  `internal/clip`, Maus, Settings+Repo-Picker+Lobby, ASCII-Fallback, Archiv-Sicht;
  Epic-bean für E5 bereits im Backlog: `bt-5h4d`).

---

## Selbst-Review (Plan gegen design-spec + implementation-plan)

- **5 Tasks, implementation-plan.md »Epos E4«s 4 inhaltlichen Punkten (Yank
  ausgenommen) zugeordnet**: Command-Center (T1 Aktionen + T2 Bean-Suche, bewusst
  gesplittet für TDD-Granularität und weil T2 CLI-abhängig ist, T1 nicht) · Review-
  Queue-View (T3, inkl. der zusätzlichen Rework-Sichtbarkeits-Sektion) · Review-
  Cockpit-Verdikte (T4) · Yank explizit als NICHT-E4-Scope dokumentiert (Design-
  Entscheidung g) statt stillschweigend ausgelassen · Abschluss (T5) — 1:1, jede
  Abweichung von der wörtlichen implementation-plan-Gliederung (Split/Rework-Sektion)
  ist oben als Design-Entscheidung explizit begründet.
- **Alle sieben Handover-Design-Punkte (a-g) + zwei zusätzliche (h/i/j) mit
  Entscheidung UND Begründung**: (a) Fuzzy-Matching ✓ · (b) Palette-Ergebnis-Modell ✓ ·
  (c) Review-Queue-Ableitung inkl. „parentlose to-review-Beans?"-Frage explizit
  beantwortet (`(kein Epic)`-Bucket) ✓ · (d) Verdikt-Mutationen GEGEN `beans update
  --help` verifiziert (nicht geraten — wörtlicher Flag-Dump im Entscheidungstext) ✓ ·
  (e) Reject-Kommentar-Form-Format ✓ · (f) Reopen-Semantik hergeleitet, da die
  §5-Tabelle sie nicht definiert ✓ · (g) Yank-Deferral ✓ · (h) Capture-Order-Erweiterung
  (NEU identifiziert: `a` kollidiert mit `keys.Assign`, ohne Fix ein echter Bug) ✓ ·
  (i) eigene Cockpit-Accordion-Felder statt Tree/Backlog-Reuse ✓ · (j) Summary-Zeilen-
  Definition ✓.
- **Handover-Gotchas adressiert**: keine Fuzzy-Lib im Repo verifiziert, eigene
  Subsequence-Implementierung gewählt (Empfehlung aus dem Handover befolgt) ✓ ·
  `WithTag('to-review')` seit T3-data wiederverwendet, kein Duplikat ✓ ·
  `focusedBean()`/`windowAround`/D08-Cursor wiederverwendet (Task 3s Queue-Rows, Task
  1/2s Palette-Bean-Zeilen) ✓ · Yank/OSC52 korrekt nach E5 verschoben (Empfehlung aus
  dem Handover befolgt, nicht als Stub gebaut) ✓ · Plan-Snippets gegen Code-Realität
  gedacht: `beans update --help` LIVE geprüft (nicht aus der Erinnerung angenommen),
  bestätigt `-s`/`--remove-tag`/`--tag`/`--body-append` sind unabhängig kombinierbar ✓.
- **Windowing/Chrome wiederverwendet, nie neu gebaut**: Review-Cockpit über
  `masterDetailWidths`/`renderPane`/`borderedPane`/`windowAround`/
  `renderBeanAccordionPane`/`relationRow` (E1/E2), Palette über `modalPanel`/
  `menuList`/`placeOverlay`/`clampModalWidth` (E1-E3) ✓.
- **data.Index nicht mutiert**: `EpicAncestor`/`reviewQueue`/`reviewRework` sind reine
  Reads über `ByID`/`Children`/`WithTag`; alle Verdikt-Mutationen laufen ausschließlich
  über `data.Client`, jede endet im bestehenden `applyMutationResult`-Unconditional-
  Reload-Pfad (E3, unverändert wiederverwendet) ✓.
- **Kein Platzhalter/TBD**: jede Task hat Dateinamen, Signaturen, Testnamen, Port-
  Referenzen (devd `file.go:NN-MM` wo zutreffend, explizit „NUR Layout-Muster, keine
  Sprint-Kopplung" wo devd strukturell nicht 1:1 übertragbar ist); alle in der
  Handover-Deliverables-Liste geforderten Design-Punkte sind mit Entscheidung UND
  Begründung festgehalten, keiner an die Implementierung delegiert.
