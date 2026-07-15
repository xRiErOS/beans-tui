# Epos E7 — PO-Feedback R1: Detail-UX + Typ/Status/Prio-Glyphen (voll granular)

Liefert: die 14 PO-Design-Entscheidungen PF-1…PF-14 (design-spec.md §15) als
Code — Entfernung des Review-Cockpits (Review läuft künftig im Chat),
Glyphen-/Farb-Umstellung (Typ/Status/Priorität), vollständige Englisch-
Übersetzung aller nutzerseitigen Strings + Command-Center-Schema, ein neuer
Detail-Pane-Kopfblock + editierbare Meta-Feldliste mit stabiler
Gutter-Spalte, Entfernung redundanter Pane-Titel, ein vollständiger
Header/Footer-Keybinding-Split ohne Dopplung (inkl. `p`/`ctrl+k` im
Header), eine Enter-Kaskade innerhalb des Detail-Fokus plus symmetrisches
`tab`/`shift+tab`, und den Epic-Abschluss. Aus 8 PO-Nachträgen + 2
Präzisierungen entstanden — Details/Zitate/Herleitung: design-spec.md §15.

Quellen: `design-spec.md` §5 (Review-Flow, umgeschrieben), §15 (PF-1…PF-14,
verbindlich) · Epic `bt-heg9` (alle Nachträge + Antworten im Body) ·
Code-Stand 2026-07-15 (Dateien einzeln je Task referenziert) · devd-Referenz
`~/Obsidian/tools/DeveloperDashboard/apps/cli-go` (je Task zitiert, wo
relevant).

Nicht in diesem Epos: Q03 (zentrale Tag-Page, bereits als eigenständiges
Feature-bean `bt-6oyy` ausgelagert) — nicht referenzieren, nicht anfassen.

## Task-Übersicht

| Task | bean | Inhalt | PF-Refs | blocked_by |
|---|---|---|---|---|
| T1 | `bt-wmtb` | Review-Cockpit entfernen | PF-14 | — |
| T2 | `bt-2af1` | Theme/Glyphen-Umstellung | PF-6 | T1 |
| T3 | `bt-w9o8` | UI-Sprache Englisch + Command-Center-Schema | PF-7, PF-8 | T2 |
| T4 | `bt-kyj5` | Meta-Layout + Kopfblock + Gutter-Stabilität | PF-1, PF-3, PF-4, PF-12 | T2, T3 |
| T5 | `bt-uyzf` | Pane-Titel-Vereinheitlichung | PF-10 | T4 |
| T6 | `bt-t1uy` | Navigation/Enter-Kaskade + Fokus-Symmetrie | PF-2, PF-5, PF-13 | T4 |
| T7 | `bt-m6at` | Header/Footer-Keybinding-Split | PF-11 | T4, T6 |
| T8 | `bt-dsog` | Abschluss (Epic to-review, E6-Bestätigung) | — | T5, T6, T7 |

**Reihenfolge-Begründung:** T1 (Removal) IMMER ZUERST (PO-Vorgabe,
Nachtrag 8: „Removal FRÜH einplanen … dann müssen Glyphen-/Footer-/
String-Umbauten das Cockpit nicht mehr mitziehen") — spart in T2/T3/T5/T7
jeweils ein View, zwei Goldens, mehrere Keybindings/String-Literale. T2
(Glyphen) vor T3 (Strings) — teilen dieselben Goldens, adjacent statt
verstreut regeneriert. T4 braucht BEIDE (Meta-Feldliste zeigt
Priority-Glyphen aus T2; Meta übernimmt selbst 4 der Accordion-
Sektionstitel aus PF-7, T3s Scope-Abgrenzung). T5/T6 bauen beide auf T4s
fertigem Meta-Layout auf, sind sonst gegenseitig unabhängig. T7 (PF-11)
zeigt `tab`/`shift+tab` im Footer und braucht daher T6s neue
`FocusIn`/`FocusOut`-Felder: `T7 blocked_by T6`. T8 zum Schluss.

**Golden-Strategie (gilt für T2/T3/T4/T5/T7):** jeder dieser Tasks ändert
sichtbaren Render-Output. Nach JEDEM: `command go build -o bin/bt .`, dann
`command go test ./internal/tui/ -run TestTreeGolden -update`, `-run
TestBacklogGolden -update`, `-run TestChromeGolden -update` (NUR NOCH DREI
Goldens seit T1 — `review_cockpit.golden` ist mit dem View gelöscht).
`git diff --stat internal/tui/testdata/` danach ansehen, JEDE geänderte
Datei bekommt eine Vorher/Nachher-Beschreibung im Commit-Body (Pflicht) —
auch „unverändert" ist eine gültige, explizit zu nennende Aussage.

---

## Task 1: Review-Cockpit entfernen (PF-14) (`bt-wmtb`)

**Files (Delete):**
- `internal/tui/view_review_cockpit.go`
- `internal/tui/view_review_cockpit_test.go`
- `internal/tui/form_reject_review.go`
- `internal/tui/form_reject_review_test.go`
- `internal/tui/testdata/review_cockpit.golden`

**Files (Modify — Compiler-gesteuerte Vollständigkeit, s. Step 3):**
`internal/tui/keymap.go`, `internal/tui/types.go`, `internal/tui/update.go`,
`internal/tui/mouse.go`, `internal/tui/box_confirm_create.go`,
`internal/tui/overlay_palette.go`, ggf. Kommentar-Bereinigung in
`internal/tui/view_lobby.go`, `internal/tui/view_browse_repo.go`,
`internal/tui/view_browse_backlog.go`, `internal/tui/context.go` (optional,
kein Compile-Zwang, nur stale Doc-Kommentare).

**Files (bewusst NICHT löschen — YAGNI-Entscheidung, dokumentieren):**
`internal/data/mutations.go` — `PassReview`/`RejectReview` (Zeilen ~316,
329) bleiben: harmlos, CLI-nah, keine TUI-Kopplung mehr nach diesem Task,
kein Risiko in einer Funktion, die niemand mehr aufruft. Datenlayer-Test-
Coverage dafür (falls vorhanden) bleibt ebenfalls unangetastet.

**Bereits VERIFIZIERT (Recherche zu diesem Plan, nicht mehr zu prüfen):**
- `keys.Reviews` (`R`) ist das EINZIGE zentrale `keyMap`-Feld für den
  Cockpit — `a`/`x`/`o`/`n`/`p` innerhalb des Cockpits sind rohe
  `msg.String()`-Vergleiche LOKAL in `keyReviewCockpit`
  (`view_review_cockpit.go`), keine eigenen `keyMap`-Felder — verschwinden
  automatisch mit der Datei, kein separater `keymap.go`-Eintrag dafür nötig
  außer `Reviews` selbst.
- `box_confirm_create.go:87` enthält `client.RejectReview(...)` im
  `submitForm`-Dispatch (vermutlich `case "rejectReview":`) — dieser Zweig
  wird nach Löschung von `openRejectForm` (`view_review_cockpit.go`)
  unerreichbar (toter Code) und muss mit entfernt werden.
- `mouse.go` hat 3 `case viewReviewCockpit:`-Stellen (`handleMouse`-Dispatch
  Zeile ~160, `wheelMove`/wheel-Dispatch Zeile ~179, plus
  `mouseReviewClick`-Funktion Zeile ~240) — alle entfallen mit dem
  `viewReviewCockpit`-Enum-Wert.
- `types.go:29` (`viewReviewCockpit`-Konstante), `types.go:397-398`
  (`reviewCursor`/`reviewAccOpen`-Felder + der Doc-Kommentar-Block
  `types.go:387-396` darüber) entfallen.

- [x] **Step 1: Baseline.** `command go test ./... -short` grün (Ausgangs-
  zustand bestätigen, bevor irgendetwas gelöscht wird).
- [x] **Step 2: Dateien löschen.** Die 5 Dateien aus „Files (Delete)" oben.
- [x] **Step 3: Compiler-gesteuerte Vollständigkeit.** `command go build
  ./...` — JEDER Compile-Fehler zeigt eine verbliebene Referenz. Iterativ
  beheben:
  - `keymap.go`: `Reviews`-Feld aus `keyMap`-Struct + `newKeyMap()` +
    `helpGroups()`s `"Views & Global"`-Gruppe entfernen.
  - `types.go`: `viewReviewCockpit`-Konstante, `reviewCursor`/
    `reviewAccOpen`-Felder + zugehöriger Doc-Kommentar entfernen.
  - `update.go`: `keybind.Matches(msg, keys.Reviews)`-Case (`keyTree`,
    Zeile ~1114) entfernen; `handleKey`s `if m.view == viewReviewCockpit {
    return m.keyReviewCockpit(msg) }`-Capture-Block entfernen;
    `focusedBean()`s `case viewReviewCockpit:` entfernen; jeden weiteren vom
    Compiler markierten Verweis (Kommentare mit Code-Bezug wie
    `openReviewCockpit`-Aufrufstellen) entfernen/bereinigen.
  - `box_confirm_create.go`: den `rejectReview`-`submitForm`-Zweig (Zeile
    ~87 + zugehöriger `case`) entfernen.
  - `mouse.go`: alle 3 `viewReviewCockpit`-Stellen + `mouseReviewClick`
    entfernen.
  - `overlay_palette.go`: den `"go to review cockpit"`-Action-Eintrag (Zeile
    ~75, `case "go_review":`-Dispatch) entfernen.
  - Wiederholen bis `command go build ./...` sauber durchläuft.
- [x] **Step 4: Test-Dateien bereinigen (kein Compile-Zwang, manuell).**
  `grep -rln "ReviewCockpit\|reviewCursor\|reviewAccOpen\|viewReviewCockpit"
  internal/tui/*_test.go` — für jede gefundene Datei (u.a.
  `context_test.go`, `mouse_test.go`, `overlay_palette_test.go`,
  `view_lobby_test.go`) die Cockpit-spezifischen Testfälle/Assertions
  entfernen (NICHT die ganze Datei — diese testen auch andere Dinge).
- [x] **Step 5:** `command go test ./... -short` → FAIL erwartet an Stellen,
  die noch Cockpit-Testfälle referenzieren — iterieren bis grün.
- [x] **Step 6: Kommentar-Hygiene (optional, kein Akzeptanzkriterium).**
  `grep -rln "eview" internal/tui/*.go | grep -v _test.go` — verbliebene
  Doc-Kommentare (z.B. `context.go:12-13,122-123`,
  `view_browse_repo.go`/`view_browse_backlog.go`/`view_lobby.go`s narrative
  Erwähnungen) bei Gelegenheit mitziehen, kein Blocker.
- [x] **Step 7: Golden-Regen.** `command go test ./internal/tui/ -run
  "TestTreeGolden|TestBacklogGolden|TestChromeGolden" -update` — erwartet:
  alle DREI byte-identisch zum Vorzustand (Cockpit-Removal betrifft NICHT
  ihre Render-Pfade) — falls NICHT identisch, Ursache klären (unerwarteter
  Nebeneffekt), nicht blind akzeptieren.
- [x] **Step 8:** `command go test ./... -race` UND `-short` (2x) grün,
  `command gofmt -l .` leer, `command go vet ./...` leer.
- [x] **Step 9:** Commit `feat(tui)!: PF-14 Review-Cockpit entfernen —
  Review läuft künftig im Chat` — Body zitiert PO-Begründung (Nachtrag 8)
  + Tag-Trio-Umstellung (design-spec §5) + YAGNI-Entscheidung
  (`PassReview`/`RejectReview` bleiben im Datenlayer). Umgesetzt als
  `refactor(tui)!:` statt `feat(tui)!:` (Removal, kein Feature-Zugewinn —
  explizite Vorgabe des Supervisor-Prompts, Commit `a25b851`).

**Akzeptanz-Checkliste:**
- [x] `viewReviewCockpit` (View, Konstante, Keybinding `R`, Modellfelder)
  vollständig aus dem Produktionscode entfernt — `command go build ./...`
  sauber, kein toter Cockpit-Code übrig
- [x] `internal/data/mutations.go`s `PassReview`/`RejectReview` bewusst
  BELASSEN (YAGNI, im Commit-Body begründet)
- [x] 2 Cockpit-Goldens gelöscht, verbleibende 3 Goldens UNVERÄNDERT
  (Regressionsbeleg)
- [x] Voller Testlauf (inkl. `-race`) grün, gofmt/vet leer

---

## Task 2: Theme/Glyphen-Umstellung (PF-6) (`bt-2af1`)

**Files:**
- Modify: `internal/theme/theme.go`
- Modify: `internal/theme/icons.go`
- Modify: `internal/theme/theme_test.go`
- Golden (Regen, s. Golden-Strategie): `internal/tui/testdata/tree.golden`,
  `backlog.golden`, `chrome.golden` (Regressionscheck)

**Referenz-Tabellen:** design-spec.md §15 PF-6 (verbindlich).

- [ ] **Step 1: Failing tests (TDD).** `theme_test.go` auf die NEUEN
  Erwartungen umschreiben:
  - `TestStatusColorMapping`: Tabelle → `draft` Blue / `todo` Green /
    `in-progress` Yellow / `completed` Subtext / `scrapped` Subtext.
    Unknown-Status→Subtext-Fallback-Assertion bleibt. Icon-Inhalts-Checks
    auf Buchstaben-Inhalt umstellen (`StatusIcon("draft")` enthält `"d"`).
  - `TestAsciiFallback`: Status/Type-ASCII-vs-Unicode-Branch-Assertions
    entfernen (Buchstaben identisch in beiden Modi) — NEU: Priority-ASCII-
    Assertions (`Priority("critical")` unter `t.Setenv("BT_ASCII_ICONS","1")`
    enthält `"!!"`, ohne enthält `"‼"`).
  - `TestTypeIconAllTypes`: Tabelle → milestone `M` Blue / epic `E` Mauve /
    feature `F` Mauve / task `T` Sky / bug `B` Red.
  - `TestPriorityColorMapping`: Tabelle → critical Red / high Yellow /
    normal Text / low Subtext / deferred Subtext. `Priority("critical")`
    enthält `"‼"`, NICHT mehr das Wort `"critical"`.
- [ ] **Step 2:** `command go test ./internal/theme/...` → FAIL.
- [ ] **Step 3: Implement `theme.go`.** `statusColor`-Map auf neue Farben;
  NEUE `statusLetter`-Map (`d`/`t`/`i`/`c`/`s`); `StatusIcon` liefert
  `StatusStyle(status).Render(statusLetter[status])` +
  `fallbackGlyph()`-Fallback (unverändert). `statusGlyph`/
  `statusGlyphASCII`-Konstanten entfernen — der DD2-176-zitierende
  Doc-Kommentar wird ERSETZT (PF-6 hebt diesen Grundsatz explizit auf,
  Begründung aus design-spec §15 PF-6 übernehmen).
- [ ] **Step 4: Implement `icons.go`.** `typeIcon`-Map → Buchstaben
  (`M`/`E`/`F`/`T`/`B`); `typeColor`-Map → neue Farben. `typeIconASCII`
  entfällt komplett (Buchstaben bereits ASCII/EAW-Neutral). `TypeIcon`/
  `TypeStyle`-Signaturen UNVERÄNDERT (Drop-in an den 3 Call-Sites).
- [ ] **Step 5: Priorität-Glyphen in `theme.go`.** NEUE Maps
  `priorityGlyph`/`priorityGlyphASCII` (critical `‼`/`!!`, high `!`/`!`,
  normal `·`/`.`, low `↓`/`v`, deferred `→`/`>`). `priorityColor`
  umschreiben (critical Red, high **Yellow**, normal Text, low
  **Subtext**, deferred **Subtext**). `Priority(p string) string` liefert
  künftig den Glyph statt des Worts, `Bold(true)` für critical/high bleibt.
- [ ] **Step 6:** `command go test ./internal/theme/...` → PASS.
- [ ] **Step 7: Golden-Regen** (s. Golden-Strategie). Erwartete Änderungen:
  `tree.golden`/`backlog.golden` (Status-/Type-Icons jeder Zeile).
  `chrome.golden` voraussichtlich UNVERÄNDERT (Chrome rendert keine
  Bean-Zeilen) — trotzdem Regressionslauf, im Commit-Body als
  „unverändert" vermerken.
- [ ] **Step 8:** `command go test ./... -short` grün, gofmt/vet leer.
- [ ] **Step 9:** Commit `feat(theme): PF-6 Glyphen-Ersatz
  (Typ/Status/Priorität)`.

**Akzeptanz-Checkliste:**
- [ ] Alle 4 `theme_test.go`-Funktionen grün gegen die NEUEN Tabellen
- [ ] `StatusIcon`/`TypeIcon` liefern Buchstaben, `Priority()` liefert Glyph
- [ ] Goldens regeneriert + Vorher/Nachher je Datei (`chrome.golden`
  explizit als „unverändert" oder verändert vermerkt)

---

## Task 3: UI-Sprache Englisch + Command-Center-Schema (PF-7, PF-8) (`bt-w9o8`)

**Scope-Abgrenzung:** die 4 Accordion-Sektionstitel `Meta`/`Body`/
`Beziehungen`/`Historie` → `META`/`BODY`/`RELATIONS`/`HISTORY` sind
PF-7-Scope, werden aber von **T4** erledigt (überarbeitet
`beanSections`/`view_detail_bean.go` ohnehin). Review-Cockpit-Strings
entfallen komplett — T1 hat die Datei bereits gelöscht.

**Files (verifizierte String-Literale, Produktionscode):**

| Datei | Alt (Deutsch) | Neu (Englisch) |
|---|---|---|
| `box_confirm_create.go:67` | `Bean nicht mehr vorhanden — Titel-Edit verworfen` | `Bean no longer exists — title edit discarded` |
| `box_confirm_create.go:82` | `Bean nicht mehr vorhanden — Ablehnung verworfen` | (entfällt — T1 hat den Reject-Zweig gelöscht; falls beim Nachziehen noch vorhanden, ebenfalls entfernen statt übersetzen) |
| `box_confirm_create.go:108` | `Einstellungen speichern fehlgeschlagen: ` | `Failed to save settings: ` |
| `box_confirm_create.go:113` | `Einstellungen neu laden fehlgeschlagen: ` | `Failed to reload settings: ` |
| `box_confirm_delete.go:130` | `Bean nicht mehr vorhanden — Löschung verworfen` | `Bean no longer exists — deletion discarded` |
| `box_confirm_delete.go:164` | `1 Kind verliert den Parent — wird zur eigenen Wurzel` | `1 child loses its parent — becomes its own root` |
| `box_confirm_delete.go:166` | `%d Kinder verlieren den Parent — werden zu eigenen Wurzeln` | `%d children lose their parent — become their own roots` |
| `box_confirm_delete.go:175` | `1 Bean verliert die Blocking-/Blocked-by-Verknüpfung zu diesem Bean` | `1 bean loses its blocking/blocked-by link to this bean` |
| `box_confirm_delete.go:177` | `%d Beans verlieren die Blocking-/Blocked-by-Verknüpfung zu diesem Bean` | `%d beans lose their blocking/blocked-by link to this bean` |
| `box_filter_facets.go:39` | `"Archiv"` (facetHead-Map-Wert) | `"Archive"` |
| `box_filter_facets.go:79` | `Archivierte einblenden` | `Show archived` |
| `box_menu_value.go:150` | `Bean nicht mehr vorhanden — Auswahl verworfen` | `Bean no longer exists — selection discarded` |
| `box_picker_blocking.go:161` | `Bean nicht mehr vorhanden — Auswahl verworfen` | `Bean no longer exists — selection discarded` |
| `box_picker_blocking.go:205` | `(keine anderen Beans im Repo)` | `(no other beans in repo)` |
| `box_picker_parent.go:51` | `label: "(Kein Parent)"` | `label: "(No parent)"` |
| `box_picker_parent.go:128` | `Bean nicht mehr vorhanden — Auswahl verworfen` | `Bean no longer exists — selection discarded` |
| `box_picker_parent.go:152` | `enter:setzen  esc:abbrechen` | `enter:set  esc:cancel` |
| `box_picker_parent.go:173` | `(keine zulässigen Eltern-Typen)` | `(no eligible parent types)` |
| `box_picker_parent.go:175` | `Parent zuweisen` (Modal-Titel) | `Assign parent` |
| `box_picker_tag.go:220` | `ungültiger Tag-Name (a-z0-9, Bindestrich-getrennt, Kleinbuchstaben)` | `invalid tag name (a-z0-9, hyphen-separated, lowercase)` |
| `box_picker_tag.go:288` | `Bean nicht mehr vorhanden — Auswahl verworfen` | `Bean no longer exists — selection discarded` |
| `box_picker_tag.go:321` | `(keine Tags im Repo)` | `(no tags in repo)` |
| `box_picker_tag.go:329` | `enter:anlegen  esc:abbrechen` | `enter:create  esc:cancel` |
| `types.go:539` | `Suche (Titel/ID, ab 3 Zeichen zusätzlich Bleve)` | `Search (title/ID, 3+ chars also searches Bleve)` |
| `types.go:549` | `Repo filtern (Pfad)` | `Filter repos (path)` |
| `update.go:261` | `Konflikt: Bean extern geändert — neu geladen` | `Conflict: bean changed externally — reloaded` |
| `update.go:284` | `Konflikt: Bean extern geändert` | `Conflict: bean changed externally` |
| `update.go:486` | `Erstellung läuft bereits — bitte warten` | `Creation already in progress — please wait` |
| `update.go:594` | `"Kopiert: "+b.ID` | `"Copied: "+b.ID` |
| `view_browse_backlog.go:233` | `watch unavailable — ctrl+r für manuelles Reload` | `watch unavailable — ctrl+r for manual reload` |
| `view_browse_repo.go:728` | `watch unavailable — ctrl+r für manuelles Reload` | `watch unavailable — ctrl+r for manual reload` |
| `view_lobby.go:192` | `(keine Repos in config.yaml -- ctrl+k -> settings)` | `(no repos in config.yaml -- ctrl+k -> settings)` |
| `overlay_palette.go` (14 Zeilen, `"go to review cockpit"` bereits von T1 entfernt) | s. PF-8-Tabelle unten | s. PF-8-Tabelle unten |

**PF-8-Remap (`overlay_palette.go`, design-spec §15 PF-8, `go to review
cockpit` entfällt — T1 hat den Eintrag bereits gelöscht):** `status:
setzen`→`set status` · `tags: zuweisen`→`set tags` · `parent:
zuweisen`→`set parent` · `blocking: zuweisen`→`set blocking` · `titel:
bearbeiten`→`set title` · `bean: löschen`→`delete bean` · `create:
bean`→`create bean` · `go to: backlog`→`go to backlog` · `go to:
browse`→`go to browse` · `filter: facetten`→`filter facets` · `search:
beans`→`search beans` · `reload: daten`→`reload data` · `repo:
wechseln`→`go to repo picker` · `settings: öffnen`→`go to settings`.

**NICHT ändern (verifiziert: reine Doc-Kommentare, kein Code-String):**
`update.go:870,929` (narrative Zitate), `messages.go:111` (Doc-Begriff,
kein UI-String). Optional, kein Akzeptanzkriterium.

- [ ] **Step 1: Vollständigkeits-Sweep VOR der ersten Änderung.**
  `grep -rnoE '"[^"]*[äöüßÄÖÜ][^"]*"' internal/tui/*.go | grep -v
  _test.go` UND `grep -rnoE '"[^"]*(gültig|löschen|zuweisen|bearbeiten|
  setzen|wechseln|öffnen|verworfen|nicht mehr|einblenden|abbrechen|keine?
  )[^"]*"' internal/tui/*.go | grep -v _test.go` — jeden Treffer gegen die
  Tabelle oben abgleichen: übersetzt, begründet ausgenommen, oder T4
  zugeordnet (Meta-Titel). Kein Treffer unbegründet übrig.
- [ ] **Step 2: Alle Strings aus der Tabelle ersetzen** (13 Dateien).
  `overlay_palette.go` nach PF-8-Schema.
- [ ] **Step 3: Test-Sweep.** Für jeden übersetzten String: `grep -rn
  "<alter deutscher Teilstring>" internal/tui/*_test.go` — Assertionen auf
  Englisch umschreiben. Insbesondere `overlay_palette_test.go`
  (`TestPaletteActionsBeanContextFirst`,
  `TestPalFilteredOrderActionsBeforeBeans`,
  `TestPaletteActionsGoReviewAfterGoBrowse` — DIESER Test referenziert
  vermutlich den bereits von T1 entfernten Cockpit-Eintrag, prüfen ob T1
  ihn schon angepasst hat, sonst hier nachziehen), `box_confirm_delete_test.go`,
  `box_picker_*_test.go`.
- [ ] **Step 4: Fuzzy-Match-Regression (PF-8, design-spec §15 explizit
  benannt).** `TestPalFilteredActionsFuzzyFiltered` gezielt lesen: Query-
  Strings gegen die NEUEN Labels prüfen (z.B. „stat" → `set status`
  weiterhin Treffer; „go" jetzt gegen 3 statt 4 `go to ...`-Einträge, Anzahl
  in Rejevanten Tests anpassen). Matching-Logik selbst NICHT ändern.
- [ ] **Step 5:** `command go test ./internal/tui/...` → PASS.
- [ ] **Step 6: Golden-Regen** (s. Golden-Strategie, 3 Goldens). Vorher/
  Nachher je tatsächlich geänderter Datei.
- [ ] **Step 7: Manueller Vollständigkeits-Beleg (tmux-Smoke).** `bin/bt`
  in tmux, Sequenz: Tree → Detail (`tab`) → esc → Backlog (`b`) →
  Filter-Menü (`f`) → esc → Command-Center (`ctrl+k`, Label-Liste lesen) →
  esc → Create-Form (`c`) → esc → Delete-Confirm (`d`, esc, NICHT
  bestätigen) → Parent-Picker (`a`) → esc → Tag-Picker (`t`) → esc →
  Blocking-Picker (`B`) → esc → Lobby (`p`) → Settings (`ctrl+k` → „go to
  settings") → esc. Jeden Screen per `capture-pane` auf verbliebenes
  Deutsch sichten. Ergebnis (PASS je Screen) in den Commit-Body.
- [ ] **Step 8:** `command go test ./... -short` grün, gofmt/vet leer.
- [ ] **Step 9:** Commit `feat(i18n): PF-7 UI-Sprache Englisch + PF-8
  Command-Center-Schema`.

**Akzeptanz-Checkliste:**
- [ ] Vollständigkeits-Sweep (Step 1) zeigt für JEDEN Treffer eine
  Zuordnung
- [ ] 14 Palette-Labels exakt nach PF-8-Tabelle (Cockpit-Eintrag bereits
  von T1 entfernt)
- [ ] `TestPalFilteredActionsFuzzyFiltered` grün
- [ ] tmux-Smoke über alle 12 verbleibenden Screens/Overlays, kein
  Deutsch mehr
- [ ] Goldens regeneriert + Vorher/Nachher je Datei

---

## Task 4: Meta-Layout + Kopfblock + Gutter-Stabilität (PF-1, PF-3, PF-4, PF-12) (`bt-kyj5`)

**Files:**
- Modify: `internal/tui/view_detail_bean.go` (`beanSections`-Signatur,
  `metaSectionBody`-Rewrite, NEU `metaFields`, NEU `detailHeaderBlock`, NEU
  `beanSectionCount`-Konstante [für T6 vorbereitet], Sektionstitel →
  `META`/`BODY`/`RELATIONS`/`HISTORY` [PF-7-Anteil, hier miterledigt])
- Modify: `internal/tui/accordion.go` (`relationField` +`kind string`-Feld
  + Doc-Kommentar-Korrektur; `renderAccordion`: PF-1-immer-offen +
  PF-12-Gutter-Fix für Sektionsköpfe + `fieldStrip`-Skip für Sektion 0;
  `fieldStrip`: PF-12-Gutter-Fix)
- Modify: `internal/tui/view_browse_repo.go` (`renderAccordionPane`:
  Kopfblock voranstellen, neue `beanSections`-Parameter durchreichen — DER
  EINZIGE verbleibende Aufrufer seit T1, `renderReviewDetailPane` existiert
  nicht mehr)
- Test: `internal/tui/view_detail_bean_test.go`, `internal/tui/accordion_test.go`

**Signatur-Änderungen (exakt):**
```go
// accordion.go
type relationField struct {
    beanID string
    label  string
    kind   string // "" = jump (Beziehungen, E2-Verhalten unverändert) |
                   // "status"|"type"|"priority" = Value-Menu seeded auf Gruppe |
                   // "title" = Title-Edit-Form | "readonly" = No-Op (created_at/updated_at)
}
```
Doc-Kommentar an `relationField` korrigieren (der Satz „never an edit
target -- edit fields are E3 scope" ist ab diesem Task falsch).

```go
// view_detail_bean.go
const beanSectionCount = 4 // Meta/Body/Beziehungen/Historie — Single Source für T6

func beanSections(idx *data.Index, b *data.Bean, bodyW int, focused bool, activeIdx, fieldIdx int) []accordionSection
func metaFields(b *data.Bean) []relationField                                  // NEU, 6 Einträge, kind-getaggt
func metaSectionBody(b *data.Bean, bodyW int, active bool, fieldIdx int) string // Signatur erweitert
func detailHeaderBlock(b *data.Bean, w int) string                             // NEU, Kopfblock
```

- [ ] **Step 1: Failing tests.**
  - `TestBeanSectionsAlwaysFourFixedSections`: Titel-Assertions auf
    `META`/`BODY`/`RELATIONS`/`HISTORY`.
  - `TestBeanSectionsMetaRendersStatusTypePriorityTags` → umbauen zu
    `TestMetaFieldsSixEntriesWithKinds`: `metaFields(b)` liefert genau 6
    Einträge (title/status/type/priority/created_at/updated_at), `kind` je
    Eintrag korrekt.
  - NEU `TestMetaSectionBodyShowsSelectedFieldMarker`: `metaSectionBody(b,
    80, true, 2)` (aktiv, fieldIdx=2) enthält `▶` genau einmal an der
    priority-Zeile, sonst `▷`. `metaSectionBody(b, 80, false, 2)` zeigt
    nirgends `▶`.
  - NEU `TestDetailHeaderBlockShowsIDTitleTypeStatusPrio`.
  - `TestBeanSectionsBeziehungen...` → umbenennen zu `...Relations...`,
    Titel-Assertion `RELATIONS`.
  - `TestBeanSectionsHistorie...` → Titel-Assertion `HISTORY`.
  - `accordion_test.go` `TestRenderAccordionExclusiveOpen` umbauen: Section
    1s Body erscheint JETZT IMMER (PF-1), Section 3 bleibt Header-only.
  - NEU `TestRenderAccordionSectionOneAlwaysOpenRegardlessOfOpenParam`.
  - NEU `TestRenderAccordionHeaderGutterWidthStableAcrossActiveState`
    (PF-12): zwei Renders mit unterschiedlichem `activeIdx` — eine
    NICHT-aktive Header-Zeile hat identische Breite in beiden.
  - NEU `TestFieldStripGutterWidthStable` (PF-12, analog für
    Beziehungen-Feldliste).
- [ ] **Step 2:** `command go test ./internal/tui/...` → FAIL.
- [ ] **Step 3: Implement `accordion.go`.**
  - `relationField` +`kind`-Feld, Doc-Kommentar korrigieren.
  - `renderAccordion`: `isOpen := n == open || n == 1` (PF-1). Gutter-Fix
    (PF-12): inaktiv `" " + truncate(marker+title+hint, w-1)`, aktiv
    `theme.Accent.Render("▌" + truncate(ansi.Strip(marker+title+hint),
    w-1))` (BEIDE Zweige jetzt `w-1`). `fieldStrip`-Aufruf NUR wenn `i !=
    0`.
  - `fieldStrip`: Gutter-Fix — inaktiv `" " + theme.Muted.Render(f.label)`,
    aktiv unverändert.
- [ ] **Step 4: Implement `view_detail_bean.go`.**
  - Sektionstitel-Konstanten `META`/`BODY`/`RELATIONS`/`HISTORY`.
  - `metaFields(b *data.Bean) []relationField`: 6 Einträge, `label` je
    Eintrag formatiert (title=`b.Title`, status/type/priority via
    bestehende `theme.*`-Helfer [nutzt T2s Glyph-Output], created_at/
    updated_at via `fmtTime`-Helfer [ggf. aus `historieSectionBody`
    extrahiert]), `kind` wie oben, `beanID` immer `""`.
  - `metaSectionBody(b, bodyW, active, fieldIdx int) string`: 6 Zeilen,
    `▷ `/`▶ `-Prefix (▶ nur bei `active && fieldIdx==zeilenindex`), via
    `wrapText` wie bisher.
  - `detailHeaderBlock(b *data.Bean, w int) string`: `b.ID` [`theme.Key`] /
    `b.Title` [`theme.Header`] / Leerzeile / `"type: "+TypeStyle+"
    status: "+StatusStyle+"    prio: "+Priority` [`theme.Muted`-Labels],
    je Zeile `truncate(..., w)`, trailing Leerzeile.
  - `beanSections(idx, b, bodyW, focused bool, activeIdx, fieldIdx int)
    []accordionSection`: Meta nutzt `fields: metaFields(b)`, `body:
    metaSectionBody(b, bodyW, focused && activeIdx==0, fieldIdx)`.
- [ ] **Step 5: Implement `view_browse_repo.go` (`renderAccordionPane`).**
  Vor `beanSections`-Aufruf: `rows = append(rows,
  strings.Split(detailHeaderBlock(b, accW), "\n")...)`. `beanSections`-Aufruf
  auf neue Signatur: `beanSections(idx, b, bodyW, focused, secCursor,
  fieldCursor)`.
- [ ] **Step 6:** `command go test ./internal/tui/...` → PASS.
- [ ] **Step 7: Golden-Regen** (3 Goldens, s. Golden-Strategie). Beide
  ändern sich stark (Kopfblock + neue Meta-Feldliste + Section-1-immer-
  offen). Vorher/Nachher explizit beschreiben.
- [ ] **Step 8:** `command go test ./... -short` grün, gofmt/vet leer.
- [ ] **Step 9:** Commit `feat(tui): PF-1/PF-3/PF-4/PF-12 Meta-Layout +
  Kopfblock + Gutter-Stabilität`.

**Akzeptanz-Checkliste:**
- [ ] Kopfblock (ID/Titel/type-status-prio) rendert über der Accordion in
  Tree UND Backlog
- [ ] Meta-Sektion zeigt IMMER ihren Body, unabhängig von `open`
- [ ] Meta-Feldliste: 6 Zeilen, `▷`/`▶`-Cursor funktional, `created_at`/
  `updated_at` NICHT editierbar (kind `readonly`)
- [ ] PF-12-Gutter-Tests grün (Sektionsköpfe UND `fieldStrip` UND
  Meta-Feldliste)
- [ ] Sektionstitel `META`/`BODY`/`RELATIONS`/`HISTORY`
- [ ] Goldens regeneriert + Vorher/Nachher je Datei

---

## Task 5: Pane-Titel-Vereinheitlichung (PF-10) (`bt-uyzf`)

**Files:**
- Modify: `internal/tui/render_shared.go` (`renderPane`: Titel+Trennlinie
  entfernen, `pane.title`-Feld entfernen)
- Modify: `internal/tui/mouse.go` (`clickPaneGeometry`: `originY`-Formel um
  die 2 entfallenen Zeilen kürzen)
- Modify: `internal/tui/view_browse_repo.go` (2 `renderPane`-Call-Sites:
  Tree UND Detail, `title:` entfernen)
- Modify: `internal/tui/view_browse_backlog.go` (1 Call-Site)
- Test: `internal/tui/render_shared_test.go`, `internal/tui/mouse_test.go`

**Konsistenz-Entscheidung (Planner):** BEIDE verbleibenden benannten Panes
(Tree/Backlog) UND Detail entfallen einheitlich — dieselbe Regel
„Breadcrumb = View-Identität, Pane-Titel nur bei echter Zusatzinfo" (Review-
Queue-Pane existiert seit T1 nicht mehr, ursprünglich der vierte
Kandidat).

- [ ] **Step 1: Failing tests.**
  - `render_shared_test.go`: bestehende `renderPane`-Tests auf die neue
    Zeilenzahl umstellen (keine Titel+Trennlinie mehr). NEU
    `TestRenderPaneNoTitleLine`.
  - `mouse_test.go`: `clickPaneGeometry`-Assertions auf neue (um 2
    kleinere) `originY` umstellen (`originY = 1 + lipgloss.Height(head) +
    1 + 1` statt `... + 1 + 1 + 1 + 1`). `TestTreeClickRow*`/
    `TestBacklogClickRow*` (per `grep -n "^func Test.*ClickRow"
    internal/tui/*_test.go` ermitteln) — Klick-Y-Koordinaten um 2 nach
    oben verschieben.
- [ ] **Step 2:** `command go test ./internal/tui/...` → FAIL.
- [ ] **Step 3: Implement `render_shared.go`.** `renderPane`: die beiden
  Titel/Trennlinien-`append`-Zeilen entfernen. `pane.title`-Feld entfernen
  (Compiler erzwingt Call-Site-Bereinigung).
- [ ] **Step 4: Call-Sites bereinigen.** 2 Stellen `view_browse_repo.go`
  (Tree, Detail), 1 Stelle `view_browse_backlog.go`.
- [ ] **Step 5: Implement `mouse.go` (`clickPaneGeometry`).** `originY`-
  Formel um die beiden `+1`-Terme kürzen. Doc-Kommentar korrigieren.
- [ ] **Step 6:** `command go test ./internal/tui/...` → PASS.
- [ ] **Step 7: Golden-Regen** (3 Goldens, `chrome.golden` NICHT betroffen
  — Chrome() nutzt `renderPane` nicht). Erwartete Änderung: Tree/Backlog-
  Panes um 2 Zeilen gewachsen, erste sichtbare Zeile ist Such-/Filterkopf.
- [ ] **Step 8:** `command go test ./... -short` grün, gofmt/vet leer.
- [ ] **Step 9:** Commit `refactor(tui): PF-10 redundante Pane-Titel
  entfernen`.

**Akzeptanz-Checkliste:**
- [ ] Kein Pane zeigt mehr Titel + Trennlinie (Tree, Backlog, Detail)
- [ ] `clickPaneGeometry`/`*ClickRow`-Tests grün mit neuen Koordinaten
- [ ] Goldens regeneriert + Vorher/Nachher je Datei

---

## Task 6: Navigation/Enter-Kaskade + Fokus-Symmetrie (PF-2, PF-5, PF-13) (`bt-t1uy`)

**Files:**
- Modify: `internal/tui/keymap.go` (NEU `FocusIn`/`FocusOut`-Felder,
  `helpGroups()`-Eintrag)
- Modify: `internal/tui/update.go` (`handleKey`: `case "tab":` →
  `keybind.Matches`-Checks für `FocusIn`/`FocusOut`; `keyDetailFocus`:
  Ziffern-Check auf `beanSectionCount`-Konstante [T4] umgestellt,
  Enter-Kaskade NEU)
- Modify: `internal/tui/box_menu_value.go` (`openValueMenu(group string)`
  statt `openValueMenu()`, NEU `currentValueForGroup`-Helfer)
- Modify: `internal/tui/update.go` (`keyNodeAction`: Call-Site
  `m.openValueMenu()` → `m.openValueMenu("status")`)
- Test: `internal/tui/update_test.go`, `internal/tui/keymap_test.go`,
  `internal/tui/box_menu_value_test.go`

**Scope-Hinweis:** PF-2 (Ziffern-Vereinheitlichung) war ursprünglich „Browse-
Detail UND Review-Cockpit angleichen" — seit T1 (PF-14) existiert kein
Cockpit mehr, PF-2 reduziert sich auf die reine `beanSectionCount`-Robustheit
in `keyDetailFocus` (kein zweiter Ort mehr zum Angleichen).

**Kollisionsanalyse (PFLICHT, vollständig — s. auch design-spec.md §15
PF-13):** `tab`/`shift+tab` sind HEUTE nicht als `keyMap`-Felder modelliert
(roher `msg.String()=="tab"`-Vergleich in `handleKey`) — nach diesem Task
sind sie es. Der bestehende `case "tab":`-Zweig sitzt in `handleKey` NACH
allen Full-Capture-Guards (`searchActive`/`filterOpen`/`form`/`overlay`/
`lobby`/`palette`/`help`/`Picker`-Match) — die neuen `FocusIn`/
`FocusOut`-Checks ersetzen ihn AN DERSELBEN STELLE, kein neues
Kollisionsrisiko. Die Enter-Kaskade (PF-5) lebt komplett innerhalb
`keyDetailFocus` (nur erreicht wenn `m.detailFocus==true`, selbst hinter
allen Guards) — keine Kollision mit `keyTree`/`keyBacklog`s eigenem
`enter`-Verhalten (unverändert: Blatt-Knoten No-Op, Knoten-mit-Kindern
Toggle-Expand). Pfeiltasten (`←`/`→`): VERIFIZIERT unverändert korrekt
getrennt — im Tree bereits heute reine Node-Navigation (`setExpanded`), in
Detail-Fokus bereits heute reine Section↔Field-Navigation inkl. Rückweg bis
zum Tree (`left` bei `detailLevel==0` → `detailFocus=false`) — keine
Code-Änderung an Pfeiltasten nötig, nur diese Verifikation als Nachweis.

- [ ] **Step 1: Failing tests — `keymap_test.go`.**
  `TestHelpGroupsCoverEveryBindingExactlyOnce` (bestehend) deckt die 2
  neuen Felder automatisch ab, sobald `helpGroups()` sie listet. NEU
  `TestFocusInFocusOutKeysBound`: `keys.FocusIn.Keys()` enthält `"tab"`,
  `keys.FocusOut.Keys()` enthält `"shift+tab"`.
- [ ] **Step 2: Failing tests — `update_test.go`.**
  - `TestKeyShiftTabExitsDetailFocus`: `m.detailFocus=true` → `shift+tab` →
    `m.detailFocus==false`, restlicher Cursor-Stand UNVERÄNDERT.
  - `TestKeyShiftTabNoopWhenNotInDetailFocus`.
  - `TestKeyTabStillTogglesBothDirections` (Regressionsschutz).
  - `TestKeyDetailFocusEnterAtSectionLevelEntersFieldLevel`: `detailLevel=0`,
    Sektion mit Feldern → `keys.Enter` → `detailLevel==1`, `fieldCursor==0`.
  - `TestKeyDetailFocusEnterAtSectionLevelNoopWithoutFields`.
  - `TestKeyDetailFocusEnterOnStatusFieldOpensValueMenuSeededToStatus`: Meta
    aktiv, `fieldCursor` auf status → `keys.Enter` →
    `m.overlay==overlayValueMenu`, Cursor auf status-Gruppe,
    `m.detailFocus` bleibt `true`.
  - `...OnTypeField...SeededToType` / `...OnPriorityField...SeededToPriority`:
    analog.
  - `TestKeyDetailFocusEnterOnTitleFieldOpensEditTitleForm`.
  - `TestKeyDetailFocusEnterOnCreatedAtFieldNoop` / `...UpdatedAtField...`.
  - `TestKeyDetailFocusEnterOnRelationFieldStillJumps` (Regressionsschutz,
    unverändertes E2-Verhalten).
  - `TestKeyDetailFocusDigitJumpUsesBeanSectionCount`: Ziffer `5` → No-Op.
- [ ] **Step 3: Failing tests — `box_menu_value_test.go`.**
  - `TestOpenValueMenuSeedsOnGivenGroup`: `m.openValueMenu("type")` mit
    Bean `Type: "bug"` → Cursor auf `type`/`bug`-Zeile.
  - Bestehenden `s`-Key-Test auf `m.openValueMenu("status")`-Call prüfen
    (Regressionsschutz, Verhalten bei `s`-Taste identisch).
- [ ] **Step 4:** `command go test ./internal/tui/...` → FAIL.
- [ ] **Step 5: Implement `keymap.go`.** `FocusIn keybind.Binding // tab —
  focus Tree<->Detail (toggle, backward-compat)`, `FocusOut
  keybind.Binding // shift+tab — deterministic focus back to Tree`.
  `newKeyMap()`: `FocusIn: NewBinding(WithKeys("tab"), WithHelp("tab","focus
  in/toggle"))`, `FocusOut: NewBinding(WithKeys("shift+tab"),
  WithHelp("shift+tab","focus out"))`. `helpGroups()`: beide in
  `"Navigation"`.
- [ ] **Step 6: Implement `update.go` (`handleKey`).** `case "tab":`-Zweig
  entfernen. Direkt danach: `if keybind.Matches(msg, keys.FocusIn) { ...
  bisherige Toggle-Logik unverändert ...; return m, nil }`, `if
  keybind.Matches(msg, keys.FocusOut) { m.detailFocus = false; return m,
  nil }`.
- [ ] **Step 7: Implement `keyDetailFocus`.** Ziffern-Check: `s[0] >= '1'
  && s[0]-'0' <= byte(beanSectionCount)`. NEU vor dem bestehenden
  `detailLevel==1`-Enter-Block: `if keybind.Matches(msg, keys.Enter) &&
  m.detailLevel == 0 { if len(secs[m.secCursor].fields) > 0 { m.detailLevel
  = 1; m.fieldCursor = 0 }; return m, nil }`. Bestehenden
  `detailLevel==1`-Enter-Block umbauen:
  ```go
  if keybind.Matches(msg, keys.Enter) && m.detailLevel == 1 {
      f := secs[m.secCursor].fields[m.fieldCursor]
      switch f.kind {
      case "status", "type", "priority":
          return m.openValueMenu(f.kind), nil
      case "title":
          return m.openEditTitleForm(b)
      case "readonly":
          return m, nil
      default: // "" — jump, E2-Verhalten unverändert
          if f.beanID == "" {
              return m, nil
          }
          m.expanded = expandAncestorsOf(m.idx, m.expanded, f.beanID)
          m.cursorID = f.beanID
          m.detailFocus = false
          return m, nil
      }
  }
  ```
  **Design-Entscheidung (dokumentieren):** bei `status`/`type`/`priority`/
  `title` bleibt `m.detailFocus` UNVERÄNDERT (`true`) — das Overlay/Form
  legt sich als eigener Capture-State darüber; nach Schließen landet der
  Nutzer wieder am selben Feld (PO-Leitprinzip D02 „schnell/einfach"). NUR
  beim `jump`-Fall bleibt `detailFocus=false` (Wechsel zu einem ANDEREN
  Bean).
- [ ] **Step 8: Implement `box_menu_value.go`.** `openValueMenu(group
  string) model`. NEU `currentValueForGroup(b *data.Bean, group string)
  string` (status/type/priority-Switch, priority mit `""→"normal"`-Default).
  Cursor-Seeding: `valueMenuCursorFor(m.menuItems, group,
  currentValueForGroup(b, group))`.
- [ ] **Step 9: Call-Site fixen.** `keyNodeAction`: `m.openValueMenu()` →
  `m.openValueMenu("status")`.
- [ ] **Step 10:** `command go test ./internal/tui/...` → PASS.
- [ ] **Step 11: Golden-Check (kein Golden-Update erwartet).** `command go
  test ./internal/tui/ -run
  "TestTreeGolden|TestBacklogGolden|TestChromeGolden"` OHNE `-update` →
  MUSS grün bleiben. Falls unerwartet FAIL: Ursache klären, nicht blind
  `-update`.
- [ ] **Step 12:** `command go test ./... -short` grün, gofmt/vet leer.
- [ ] **Step 13:** Commit `feat(tui): PF-2/PF-5/PF-13 Enter-Kaskade +
  Ziffern-Robustheit + Fokus-Symmetrie`.

**Akzeptanz-Checkliste:**
- [ ] `shift+tab` verlässt Detail-Fokus deterministisch, `tab` behält sein
  bestehendes Toggle-Verhalten
- [ ] `enter` auf Sektions-Ebene → Feld-Navigation; `enter` auf Feld →
  passendes Edit-Overlay bzw. No-Op bzw. unverändertes Jump-Verhalten
- [ ] `m.detailFocus` bleibt bei Overlay-Dispatch erhalten, nur bei Jump
  auf `false`
- [ ] Ziffern-Bereichs-Check nutzt `beanSectionCount`
- [ ] Kollisionsanalyse vollständig, keine neue Tasten-Kollision
- [ ] Kein Golden ändert sich (Step 11 grün ohne `-update`)

---

## Task 7: Header/Footer-Keybinding-Split (PF-11) (`bt-m6at`)

**Voraussetzung:** T6 muss gelandet sein (`FocusIn`/`FocusOut` müssen
existieren, damit der Footer sie anzeigen kann).

**Files:**
- Modify: `internal/tui/keymap.go` (NEU `globalBindings()
  []keybind.Binding`; Help-Text-Kürzung `Refresh`→`"reload"`,
  `Palette`→`"commands"`, `Picker`→`"repos"`)
- Modify: `internal/tui/view_browse_repo.go` (`browseRepoChrome`: Header
  nutzt `globalBindings()`, `localHint` verliert `Refresh`/`Enter`, gewinnt
  `FocusIn`/`FocusOut` über `renderBindings` statt hand-getipptem
  `"  tab:focus"`-Suffix)
- Modify: `internal/tui/view_browse_backlog.go` (`backlogChrome`: Header
  `globalBindings()`, `localHint` verliert `Enter`, gewinnt
  `FocusIn`/`FocusOut`)
- Modify: `internal/tui/view.go` ODER neue Datei
  `internal/tui/footer_context.go` (NEU: kontextsensitive Footer-Auswahl,
  s. Step 6)
- Test: `internal/tui/keymap_test.go` (NEU Drift-Guard-Disjunktheit),
  `internal/tui/chrome_test.go`, view-spezifische Chrome-Tests

**Scope-Hinweis:** seit T1 (PF-14) gibt es nur noch ZWEI `*Chrome()`-
Baufunktionen (`browseRepoChrome`/`backlogChrome`) — `reviewCockpitChrome`
existiert nicht mehr.

**Global-Set (design-spec.md §15 PF-11, erweitert Nachtrag 9 — ALLE
globalen Bindings, nicht nur ein Teil):** `{Refresh, Palette, Picker, Help,
Back, Enter, Quit}` — Header zeigt künftig `ctrl+r:reload · ctrl+k:commands
· p:repos · ?:help · esc:back · enter:open/confirm · q:quit` (7 statt
bisher 3 — `esc`/`enter`/`ctrl+k`/`p` fehlen heute komplett im Header).

**Kontextsensitiver Footer (Q04-Antwort, PO-Nachtrag 5/Q04):** View-lokal im
Normalzustand; sobald ein Overlay/Form/Filter aktiv ist, zeigt der Footer
DESSEN lokale Bindings statt der (dann irrelevanten) View-Bindings —
insbesondere `keys.Toggle` (`space/x:Toggle facet`) beim offenen
Filter-Menü, das Q04 auslöste.

- [ ] **Step 1: Failing tests — `keymap_test.go`.**
  - NEU `TestGlobalBindingsExactSet`: `globalBindings()` liefert exakt
    `{Refresh, Palette, Picker, Help, Back, Enter, Quit}` in dieser
    Reihenfolge.
  - NEU `TestGlobalBindingHelpTextsShortened`: `keys.Refresh.Help().Desc ==
    "reload"`, `keys.Palette.Help().Desc == "commands"`,
    `keys.Picker.Help().Desc == "repos"`.
  - NEU `TestNoDuplicateBindingBetweenGlobalAndAnyLocalHintList`
    (Drift-Guard, Reflection): für `browseRepoLocalBindings()`/
    `backlogLocalBindings()` (neu benannte Funktionen, s. Step 5): kein
    `Keys()`-Wert überschneidet sich mit `globalBindings()`s `Keys()`.
- [ ] **Step 2: Failing tests — Chrome-Funktionen.** Bestehende Tests (per
  `grep -n "func Test.*Chrome" internal/tui/*_test.go` ermitteln) auf die
  neuen Header-/Footer-Inhalte umstellen.
- [ ] **Step 3:** `command go test ./internal/tui/...` → FAIL.
- [ ] **Step 4: Implement `keymap.go`.** Help-Texte kürzen (s. oben). `func
  globalBindings() []keybind.Binding { return []keybind.Binding{keys.Refresh,
  keys.Palette, keys.Picker, keys.Help, keys.Back, keys.Enter, keys.Quit} }`.
- [ ] **Step 5: Implement die 2 Chrome-Funktionen.** Jede extrahiert ihre
  bisherige inline `[]keybind.Binding{...}`-Liste in eine benannte Funktion
  (`browseRepoLocalBindings()`, `backlogLocalBindings()`) — OHNE
  `Refresh`/`Enter`, MIT `FocusIn`/`FocusOut` statt hand-getipptem
  `tab:focus`. `head = breadcrumb(m.repoLabel(), "<Titel>",
  renderBindings(globalBindings()), innerW)` ersetzt die bisherige
  2× duplizierte Zeile.
- [ ] **Step 6: Implement kontextsensitiven Footer.** Neue Funktion
  (`view.go` oder `footer_context.go`): `func (m model) contextualLocalHint(viewLocal
  []keybind.Binding) string` — Switch-Priorität: `m.filterOpen` →
  `renderBindings([]keybind.Binding{keys.Up, keys.Down, keys.Toggle,
  keys.FilterClear, keys.Enter, keys.Back})`; `m.overlay != overlayNone` →
  overlay-spezifisches Set (Value-Menu: `{Up,Down,Enter,Status,Back}`;
  Tag-/Parent-/Blocking-Picker: `{Up,Down,Enter,Back}`); `m.searchActive`
  → `{Enter,Back}`; `m.paletteOpen` → `{Enter,Back}`; `m.helpOpen` →
  `{Back}`; sonst → `renderBindings(viewLocal)`. Beide Chrome-Funktionen
  rufen diese mit ihrer eigenen `*LocalBindings()`-Liste als `viewLocal`
  auf.
- [ ] **Step 7:** `command go test ./internal/tui/...` → PASS.
- [ ] **Step 8: Golden-Regen** (3 Goldens). ALLE betroffen (Header+Footer
  Teil jeder View). Vorher/Nachher je Datei.
- [ ] **Step 9:** `command go test ./... -short` grün, gofmt/vet leer.
- [ ] **Step 10:** Commit `refactor(tui): PF-11 Header/Footer-Keybinding-
  Split (vollständige Global-Liste) + kontextsensitiver Footer (Q04)`.

**Akzeptanz-Checkliste:**
- [ ] Header zeigt exakt `{Refresh, Palette, Picker, Help, Back, Enter,
  Quit}` (7 Bindings), EINE Quelle (`globalBindings()`)
- [ ] Kein Binding erscheint gleichzeitig in Header UND einer View-lokalen
  Footer-Liste (Drift-Guard-Test grün)
- [ ] Footer wechselt kontextsensitiv bei offenem Filter-Menü/Overlay/Form
  (inkl. `space:Toggle facet`, Q04)
- [ ] Goldens regeneriert + Vorher/Nachher je Datei

---

## Task 8: Abschluss (`bt-dsog`)

**Files:** keine Code-Änderungen erwartet; beans-Metadaten (`bt-heg9`)

**blocked_by:** T5, T6, T7

- [ ] **Step 1: Voller Regressionslauf.** `command go build -o bin/bt .`,
  `command go test ./... -race`, `command go test ./... -short` (2x),
  `command gofmt -l .` leer, `command go vet ./...` leer.
- [ ] **Step 2: E6-Blocking bestätigen (Verifikation, keine Aktion nötig).**
  `beans show bt-wm4w --json | jq .blocked_by` und `beans show bt-9yvh
  --json | jq .blocked_by` → beide enthalten `bt-heg9` (bereits vom PO/D03
  gesetzt) — im Commit-Body bestätigen.
- [ ] **Step 3: US-08/US-12-Hinweis für E6.** Kurzer Vermerk im
  Commit-Body: die E6-Validierung von US-08 (jetzt „Review-Signal sichtbar"
  statt „Cockpit fahren", PF-14) UND US-12 (design-spec.md §10, „devd-
  Look") müssen gegen den NEUEN Stand (PF-1…PF-14) laufen —
  `validation.md` (E6 Task 3) referenziert `design-spec.md §5`/§15 als
  Quelle.
- [ ] **Step 4: Epic-Ritual.** `beans update bt-heg9 --tag to-review`
  (Epic — PO-Gate, Agent setzt NIE `completed`).
- [ ] **Step 5: Task-beans dieses Eposs (T1-T7) auf `completed`**
  (agent-abschließbar, NICHT das Epic selbst).
- [ ] **Step 6:** `docs/SSTD.md` — Pointer-Update nur falls sich
  Referenzen geändert haben (voraussichtlich unverändert).
- [ ] **Step 7:** Commit `docs(release): E7-Abschluss — Epic to-review,
  E6-Blocking bestätigt`.

**Akzeptanz-Checkliste:**
- [ ] Voller Testlauf (inkl. `-race`) grün, gofmt/vet leer
- [ ] `bt-heg9` trägt Tag `to-review`, ist NICHT `completed`
- [ ] `bt-wm4w`/`bt-9yvh` bestätigt `blocked_by: [bt-heg9]`
- [ ] T1-T7 alle `completed`

---

## Selbst-Review (Plan gegen alle 8 PO-Nachträge + design-spec §15)

- **Alle 14 PF-Punkte einem Task zugeordnet:** PF-14 → T1 · PF-6 → T2 ·
  PF-7/PF-8 → T3 (minus 4 Accordion-Titel → T4, Cockpit-Strings entfallen
  durch T1) · PF-1/PF-3/PF-4/PF-12 → T4 · PF-10 → T5 · PF-2/PF-5/PF-13 → T6
  · PF-11 → T7 (inkl. Nachtrag 9 Erweiterung). Kein PF-Punkt doppelt
  vergeben, keiner vergessen.
- **PF-14-Reihenfolge-Vorgabe erfüllt:** Removal läuft als T1, ALLE
  Folge-Tasks (T2/T3/T5/T7) haben dadurch reduzierten Scope (ein View,
  zwei Goldens, mehrere Keybindings/Strings weniger) — in jedem Task
  explizit als „Scope-Hinweis" vermerkt, nicht stillschweigend
  vorausgesetzt.
- **D01-Revision (PO-Nachtrag 3) vollständig nachgezogen:** T6 trägt die
  REVIDIERTE PF-5-Fassung (`tab` bleibt einziger Einstieg, Kaskade nur
  innerhalb).
- **Q03/Q04/Tag-Trio korrekt behandelt:** Q03 (`bt-6oyy`) explizit
  außerhalb E7. Q04 in T7 eingearbeitet (kontextsensitiver Footer). Das
  Tag-Trio (`to-review`/`accepted`/`rejected`, PF-14-Präzisierung) ist reine
  design-spec-§5-Doku (keine TUI-Interaktion, kein eigener Task nötig —
  Tags sind bereits generisch sichtbar, keine Code-Änderung erforderlich).
- **Golden-Update-Pflicht erfüllt:** T2/T3/T4/T5/T7 haben je einen
  expliziten Golden-Regen-Schritt MIT Vorher/Nachher-Pflicht (jetzt nur
  noch 3 Goldens statt 4, seit T1); T1 UND T6 haben je einen expliziten
  Gegenbeleg-Schritt (Goldens bleiben unverändert, verifiziert statt
  angenommen).
- **Keymap-Drift-Guard-Pflicht erfüllt:** T6s neue `FocusIn`/`FocusOut`-
  Felder sind in `helpGroups()` verdrahtet — bestehender Guard deckt sie
  automatisch ab. T7 fügt einen NEUEN Disjunktheits-Test hinzu
  (Header-vs-Footer, jetzt nur noch 2 Local-Listen statt 3). T1 entfernt
  `keys.Reviews` — Compiler + bestehende Drift-Guard-Tests zeigen
  automatisch, falls `helpGroups()` danach eine Lücke hätte (hat sie
  nicht, da `Reviews` selbst mitentfernt wird).
- **Kollisionsanalyse (enter/1-9/tab/shift+tab) vollständig:** T6s
  Task-Kopf enthält die vollständige Analyse (Capture-Order-Beweis,
  State-Exklusivität, Pfeiltasten-Verifikation) — jeder Punkt anhand des
  tatsächlichen Codes belegt.
- **PF-14-Removal-Vollständigkeit:** T1 nutzt eine Compiler-gesteuerte
  Löschstrategie (Datei löschen → `go build` → jeden Fehler beheben →
  wiederholen) statt einer von Hand erstellten, potenziell unvollständigen
  Liste — robuster als ein reiner Grep-Sweep (Lehre aus T3s eigenem
  PF-7-Sweep, wo ein erster Grep-Durchlauf nachweislich Treffer verpasst
  hatte, s. `box_filter_facets.go`/`update.go:594`-Funde während der
  Planung).
- **Datenlayer-YAGNI-Entscheidung dokumentiert:** `PassReview`/
  `RejectReview` bleiben (T1, explizit begründet), keine stillschweigende
  Vollständig-Löschung über die TUI-Verdrahtung hinaus.
- **E6-Blocking:** bereits vom PO gesetzt (D03, bestätigt `bt-wm4w`/
  `bt-9yvh` ← `bt-heg9`) — T8 verifiziert nur, ändert nichts.
