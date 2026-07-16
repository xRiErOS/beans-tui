---
# bt-d8kc
title: Header/Footer-Neuspezifikation R2 + Backlog-Sort-Indikator
status: in-progress
type: task
priority: high
created_at: 2026-07-15T21:06:57Z
updated_at: 2026-07-16T02:00:51Z
parent: bt-ntoz
---

E8 Task 8 — deckt D02 (Backlog-Sort-Suffix), D04 (Header auf 4 Globals gekuerzt), D06+Q06 (Footer-Neuspezifikation inkl. Blocking-Remap B->r) aus bean bt-ntoz. Quelle: design-spec.md §15 PF-16. Ist-Code: internal/tui/keymap.go (globalBindings, keyMap.Blocking), internal/tui/view_browse_repo.go (browseRepoLocalBindings, treeSearchLine), internal/tui/view_browse_backlog.go (backlogLocalBindings). Unabhaengig von T1/T3/T4/T6/T7 (kein Ueberlapp mit view_detail_bean.go/update.go's keyDetailFocus) -- kein blocked_by.

## D04 — Header auf 4 Globals

globalBindings() (keymap.go) liefert HEUTE 7 Bindings ({Refresh, Palette, Picker, Help, Back, Enter, Quit}). NEU: exakt 4, in dieser Reihenfolge: {Palette, Picker, Help, Quit} -- ctrl+r(Refresh)/esc(Back)/enter(Enter) fliegen aus dem Header (bleiben ausschliesslich im Help-Overlay dokumentiert -- helpGroups() bleibt UNVERAENDERT, die 3 Bindings sind dort bereits gelistet). Header-Text danach: `ctrl+k:commands · p:repos · ?:help · q:quit`.

## D06 + Q06 — Footer-Neuspezifikation (ersetzt den kompletten T7/PF-11-Stand)

browseRepoLocalBindings()/backlogLocalBindings() (view_browse_repo.go/view_browse_backlog.go) HEUTE: Up/Down/Left/Right + eine Teilmenge der Aktionen. NEU (Q06 GELOEST, finale Liste, PO verbatim):

`tab focus in · shift+tab focus out · / search · f Filter · s Status · c Create · d Delete · e Edit · b Backlog · t Tags · y Yank · a Parent · r Blocking`

Konkret:
- Navigations-Keys (Up/Down/Left/Right) KOMPLETT raus aus BEIDEN Listen ("intuitiv genug", PO-Begruendung).
- Reihenfolge: FocusIn/FocusOut ZUERST, dann die Aktionen in exakt obiger Reihenfolge.
- browseRepoLocalBindings() NEU: {FocusIn, FocusOut, Search, Filter, Status, Create, Delete, Editor, Backlog, TagAssign, Yank, Assign, Blocking} (Backlog=`b`-Taste zum Wechsel IN den Backlog, aus der Tree-View sinnvoll -- war vorher NICHT in dieser Liste).
- backlogLocalBindings() NEU: {FocusIn, FocusOut, Search, Filter, Status, Create, Delete, Editor, Backlog, TagAssign, Yank, Assign, Blocking} PLUS Sort ans Ende angehaengt (siehe Planner-Entscheidung unten) -- Backlog=`b` fuehrt HIER zurueck zu Browse (bestehendes Verhalten unveraendert, nur jetzt auch angezeigt).
- Planner-Entscheidung (dokumentieren, ERRATUM ggue. wortwoertlicher Q06-Liste): Sort (`S`) ist NICHT Teil der Q06-Liste (die Liste ist fuer Browse+Backlog gemeinsam formuliert, Sort existiert nur in Backlog). Sort bleibt TROTZDEM im Backlog-Footer erhalten (Backlog-exklusiver Zusatz, kein Entzug einer heute sichtbaren, haeufig genutzten Funktion ohne explizite PO-Anweisung) -- NICHT einfach stillschweigend weglassen.
- X (FilterClear) bleibt wie bisher NICHT in der Liste (war schon vorher nicht drin, Q06-Liste bestaetigt das implizit durch Auslassung -- keine Aenderung).

## Q06 — Blocking-Remap B->r

keys.Blocking (keymap.go) wechselt von "B" auf "r" (r seit PF-14/Review-Cockpit-Removal frei, verifiziert -- kein anderer keyMap-Eintrag nutzt "r"; B wird durch den Remap frei, aktuell von NICHTS sonst referenziert -- grep bestaetigt). newKeyMap(): `Blocking: keybind.NewBinding(keybind.WithKeys("r"), keybind.WithHelp("r", "Blocking picker"))`. helpGroups() referenziert dasselbe keyMap-Feld (Single Source) -- automatisch aktualisiert, KEIN separater Eintrag noetig. Footer-Label gemaess Q06-Liste: "r Blocking".

## D06 — Optik: Taste TEAL, Aktionswort subtext-grau, kein ":"

Betrifft renderBindings() (view.go) -- pruefen ob diese Funktion HEUTE bereits Taste/Label unterschiedlich stylt oder einen ":" hart einbaut. Falls ":" hart kodiert ist: entfernen, Farbtrennung uebernimmt die Funktion des Trenners (Taste = theme.Teal ODER ein bestehendes Akzent-Token -- klaeren welches Teal-Aequivalent fuer Tasten schon in Verwendung ist, ODER denselben theme.HeaderInactive-Token aus T2/bt-czpf falls diese Task zuerst gelandet ist [kein blocked_by hier, aber pruefen ob T2 bereits gemergt wurde -- falls ja, Token wiederverwenden statt zu duplizieren; falls nein, einen eigenen Teal-Style in theme.go definieren, T2 raeumt dann ggf. spaeter zusammen]). GILT EINHEITLICH auch fuer die 4 Header-Globals (D04) -- renderBindings() ist die GEMEINSAME Funktion fuer Header UND Footer (globalBindings() und die *LocalBindings()-Funktionen werden beide durch renderBindings() gerendert, view.go) -- EINE Aenderung deckt beide Zonen ab.

## Backlog-Sort-Indikator (D02)

treeSearchLine(w int) (view_browse_repo.go) ist HEUTE von Tree UND Backlog gemeinsam genutzt (Backlog ruft dieselbe Funktion fuer seine Suchkopf-Zeile). NEU: Signatur `treeSearchLine(w int, sortSuffix string) string` -- sortSuffix wird in ALLEN DREI bestehenden Render-Zweigen (searchActive / treeActive / idle) angehaengt, WENN nicht leer (z.B. `line += "  · " + sortSuffix`), IMMER truncate-geschuetzt (Budget bleibt w). Tree-Call-Site (viewBrowseRepo) uebergibt "" (unveraendertes Verhalten, tree.golden bleibt fuer die Suchzeile byte-identisch -- die bestehende "idle-hint bleibt unveraendert"-Garantie im Doc-Kommentar bleibt gueltig, da leerer Suffix ein No-Op ist). Backlog-Call-Site (viewBacklog) uebergibt `"sort " + backlogSortDisplayLabel(m.backlogSort)` -- neue kleine Helferfunktion, die "" auf "status" mapped (dieselbe Alias-Konvention wie nextBacklogSort) und "priority" auf "prio" abkuerzt (PO-Beispiel woertlich: "⌕ / search · sort prio"), alle anderen Modi (status/created/updated) unveraendert als Wort. Farbe: theme.Muted (dezent/subtext, PO-Wortlaut "dezenter Suffix").

## TDD-Schritte

1. Failing tests: keymap_test.go (TestGlobalBindingsExactSet -> want-Slice auf {Palette,Picker,Help,Quit} aendern; TestNoDuplicateBindingBetweenGlobalAndAnyLocalHintList bleibt strukturell, deckt automatisch die neuen Listen ab; NEU TestBlockingKeyIsR: keys.Blocking.Keys() enthaelt "r" nicht "B"). view_browse_repo_test.go/view_browse_backlog_test.go (Assertions auf die neuen *LocalBindings()-Listen, NEU TestBacklogLocalBindingsIncludesSort). view_detail_bean... NEIN, treeSearchLine-Tests liegen vermutlich in view_browse_repo_test.go (grep pruefen) -- NEU TestTreeSearchLineEmptySuffixUnchanged (Tree-Golden-Aequivalenz), NEU TestBacklogSortSuffixShowsAbbreviatedPriority.
2. command go test ./internal/tui/... -> FAIL.
3. Implementieren: keymap.go zuerst (globalBindings + Blocking-Remap), dann view.go (renderBindings Optik falls noetig), dann view_browse_repo.go (*LocalBindings + treeSearchLine-Signatur + Tree-Call-Site), dann view_browse_backlog.go (*LocalBindings + Backlog-Call-Site + backlogSortDisplayLabel).
4. command go test ./internal/tui/... -> PASS.
5. Golden-Regen (3 Goldens): command go build -o bin/bt ., dann command go test ./internal/tui/ -run "TestTreeGolden|TestBacklogGolden|TestChromeGolden" -update. ALLE DREI aendern sich voraussichtlich (Header UND Footer sind Teil jeder View, chrome.golden rendert eine generische Chrome() mit STATISCHEN ChromeOpts -- pruefen ob chrome_test.go ueberhaupt globalBindings()/*LocalBindings() beruehrt, ERRATUM-Praezedenzfall aus epic-E7-plan.md Task 7 Step 8: chrome.golden blieb dort unveraendert, da chrome_test.go eigene Literal-ChromeOpts nutzt, KEIN Aufruf von globalBindings()/browseRepoChrome() -- falls hier identisch: chrome.golden bleibt unveraendert, explizit vermerken statt blind -update).
6. command go test ./... -short gruen (2x), command go test ./... -race gruen, gofmt/vet leer.
7. Commit feat(tui): PF-16 Header/Footer-Neuspezifikation R2 (D04,D06,Q06) + Backlog-Sort-Indikator (D02) -- Body dokumentiert die Sort-bleibt-erhalten-Planner-Entscheidung (ERRATUM ggue. woertlicher Q06-Liste) + die Blocking-Remap-Kollisionspruefung (B frei geworden, r vorher frei verifiziert).

## Akzeptanz-Checkliste

- [ ] Header zeigt exakt 4 Bindings: ctrl+k · p · ? · q (Refresh/Back/Enter draussen)
- [ ] Footer (Browse+Backlog) zeigt exakt die Q06-Liste in der vorgegebenen Reihenfolge (FocusIn/FocusOut zuerst)
- [ ] Backlog-Footer behaelt zusaetzlich Sort (dokumentierte Planner-Entscheidung)
- [ ] keys.Blocking ist jetzt an "r" gebunden, "B" ist frei (kein anderer Verwender)
- [ ] Taste/Aktionswort farblich getrennt statt ":" -- gilt fuer Header UND Footer identisch
- [ ] Backlog-Suchzeile zeigt "· sort <label>"-Suffix, Tree-Suchzeile unveraendert (leerer Suffix)
- [ ] Drift-Guard (TestNoDuplicateBindingBetweenGlobalAndAnyLocalHintList) gruen mit den neuen Listen
- [ ] Goldens regeneriert, Vorher/Nachher (inkl. chrome.golden-ERRATUM-Praezedenzfall-Pruefung) je Datei vermerkt
- [ ] Voller Testlauf (inkl. -race) gruen, gofmt/vet leer
