---
# bt-d8kc
title: Header/Footer-Neuspezifikation R2 + Backlog-Sort-Indikator
status: completed
type: task
priority: high
created_at: 2026-07-15T21:06:57Z
updated_at: 2026-07-16T03:02:13Z
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

- [x] Header zeigt exakt 4 Bindings: ctrl+k · p · ? · q (Refresh/Back/Enter draussen)
- [x] Footer (Browse+Backlog) zeigt exakt die Q06-Liste in der vorgegebenen Reihenfolge (FocusIn/FocusOut zuerst)
- [x] Backlog-Footer behaelt zusaetzlich Sort (dokumentierte Planner-Entscheidung)
- [x] keys.Blocking ist jetzt an "r" gebunden, "B" ist frei (kein anderer Verwender)
- [x] Taste/Aktionswort farblich getrennt statt ":" -- gilt fuer Header UND Footer identisch
- [x] Backlog-Suchzeile zeigt "· sort <label>"-Suffix, Tree-Suchzeile unveraendert (leerer Suffix)
- [x] Drift-Guard (TestNoDuplicateBindingBetweenGlobalAndAnyLocalHintList) gruen mit den neuen Listen
- [x] Goldens regeneriert, Vorher/Nachher (inkl. chrome.golden-ERRATUM-Praezedenzfall-Pruefung) je Datei vermerkt
- [x] Voller Testlauf (inkl. -race) gruen, gofmt/vet leer


## Summary

D02/D04/D06/Q06 vollstaendig umgesetzt in EINEM Commit (Begruendung im
Commit-Body, siehe unten):

- **D04**: `globalBindings()` (keymap.go) auf exakt 4 Bindings gekuerzt --
  `Palette, Picker, Help, Quit`. `ctrl+r`/`esc`/`enter` fliegen aus dem
  Header, bleiben in `helpGroups()` (Help-Overlay) unveraendert dokumentiert.
- **Q06 Blocking-Remap**: `keys.Blocking` `B`->`r` (r seit PF-14 frei,
  Reflection-Scan verifiziert kollisionsfrei). Mehrere `Help().Desc`-Texte
  auf Q06s PO-verbatim Wortlaut gekuerzt (Single Source, zieht automatisch
  durchs Help-Overlay durch).
- **D06+Q06 Footer**: `browseRepoLocalBindings()`/`backlogLocalBindings()`
  komplett neu -- Navigation raus, FocusIn/FocusOut zuerst, dann Q06s exakte
  Reihenfolge. Backlog behaelt zusaetzlich `Sort` (dokumentierte
  Planner-Entscheidung/ERRATUM, siehe unten).
- **D06 Optik**: `renderBindings()` (view.go) -- Taste `theme.BindingKey`
  (Teal), Wort `theme.BindingDesc` (Subtext), kein `:` mehr. EIGENE
  Theme-Tokens statt `theme.HeaderInactive` (B06-Experiment) wiederzuverwenden
  -- Begruendung: Farbwahl-Entkopplung, siehe Deviations unten.
- **D02**: `treeSearchLine(w, sortSuffix)` -- Tree `""` (No-Op), Backlog
  `"sort "+backlogSortDisplayLabel(m.backlogSort)` (PO-Beispiel woertlich
  `"⌕ / search · sort prio"`).
- **Lobby-Optik** (bt-1u0t Notes-fuer-bt-d8kc): `lobbyExitHint()` an D06
  angeglichen (kein `:` mehr), NUR dieser Abschnitt -- Verhalten unveraendert.
- **t-Picker-Hint** (bt-yqdy Notes-fuer-bt-d8kc): `tagPickerLocalBindings()`
  unangetastet, live im Smoke verifiziert (`n New tag` weiterhin sichtbar).

## Test-Output (RED->GREEN je Haeppchen)

1. **Q06 B->r Remap** (`TestBlockingKeyIsR`, `TestGlobalBindingsExactSet`,
   `TestGlobalBindingsOmitsRefreshBackEnter`, `TestGlobalBindingsFitIn80Columns`):
   RED `globalBindings() has 7 entries, want 4` / `Blocking.Keys() = [B], want
   to contain "r"` -> GREEN nach `keymap.go`-Edit. Downstream `runeMsg('B')`
   in `box_picker_blocking_test.go`(x7)/`etag_conflict_test.go`(x1)/
   `box_menu_value_test.go`(x1) auf `runeMsg('r')` umgestellt (RED:
   `overlay = 0, want overlayBlockingPicker` -> GREEN).
2. **D06 renderBindings-Optik** (`TestRenderBindingsSkipsHelplessAndColorSeparatesNoColon`,
   `TestRenderBindingsColorsKeyTealDescSubtext`): RED
   `ansi.Strip(renderBindings(bs))="enter:open  esc:back", want "enter open ·
   esc back"` -> GREEN nach `view.go` renderBindings/footer/breadcrumb-Edit.
3. **Q06 Footer-Listen** (`TestBrowseRepoChromeFooterMatchesQ06List`,
   `TestBacklogChromeFooterMatchesQ06ListPlusSort`,
   `TestBrowseRepoLocalBindingsOmitsNavigation`,
   `TestBacklogLocalBindingsIncludesSort`): RED (alte Nav-lastige Liste)
   -> GREEN nach `browseRepoLocalBindings()`/`backlogLocalBindings()`-Rebuild.
4. **D02 Sort-Suffix** (`TestBacklogSortDisplayLabel`,
   `TestTreeSearchLineEmptySuffixUnchanged`,
   `TestBacklogSortSuffixShowsAbbreviatedPriority`,
   `TestTreeSearchLineSortSuffixAppendsIn{ActiveSearch,TreeActive}Branch`):
   RED (compile error `undefined: backlogSortDisplayLabel`) -> GREEN nach
   `treeSearchLine`-Signaturaenderung + Helper.
5. **Lobby-Optik** (`TestLobbyExitHintNoColon`,
   `TestLobbyHintReflectsSplitEscQBehavior`): RED `lobbyExitHint(with client) =
   "esc:back  q:quit", must not contain ':'` -> GREEN nach `view_lobby.go`-Edit.
6. **NBSP-Regression-Fix** (gefunden im tmux-Smoke NACH #2-5, eigener
   RED->GREEN-Zyklus): `TestRenderBindingsKeepsKeyAndDescTogetherAcrossWrap`
   RED `wrapText separated desc "Sort" from its key "S" onto the PREVIOUS
   line` -> GREEN nach `nbsp`-Const + `strings.ReplaceAll`-Fix in
   `renderBindings()`. `TestRenderBindingsKeepsMultiWordDescTogether` analog
   fuer mehrwoertige Descs ("focus in").
7. Alle `footer_context_test.go`/`view_browse_*_test.go`-Assertions, die
   String-Content pruefen, auf `stripHint()` (neuer Test-Helper,
   `update_test.go`: `ansi.Strip` + NBSP->Space) bzw. das no-colon-Format
   umgestellt -- systematisch (Alt-Test-Falle wie im Auftrag erwartet).

**Voller Lauf**: `command go build -o bin/bt .` gruen · `command go vet
./...` leer · `gofmt -l .` leer · `command go test ./... -count=1` 2x GRUEN
(`internal/tui` ~137s je Lauf) · `command go test ./... -race -count=1`
GRUEN (~140s) · Goldens (`TestTreeGolden`/`TestBacklogGolden`/
`TestChromeGolden`) `-count=2` byte-stabil.

## Golden-Diffs

- **tree.golden**: (a) Header 2 Zeilen (alter 7-Item-Umbruch) -> 1 Zeile
  (neuer 4-Item-Header passt); (b) dadurch 1 zusaetzliche Body-Zeile
  (bodyH waechst um die freigewordene Header-Zeile); (c) Footer komplett
  ersetzt -- Nav-Keys raus, Q06-Liste, kein `:` mehr, Teal/Subtext (ANSI
  im Rohbyte, im ANSI-gestrippten Diff nicht sichtbar).
- **backlog.golden**: wie tree.golden PLUS (d) Suchzeile zeigt
  `· sort status` (Default-Sort-Alias) NEU; (e) Footer zusaetzlich
  `· S Sort` am Ende (Backlog-exklusiv, ERRATUM-dokumentiert).
- **chrome.golden**: NUR ANSI-Diff, kein Text-Diff -- `breadcrumb()`s
  GlobalHint-Literal und `footer()`s FooterHint-Literal (chrome_test.go
  eigene `ChromeOpts`-Fixtures, `KEIN` Aufruf von `globalBindings()`/
  `renderBindings()`) verlieren ihre pauschale `theme.Muted`/`theme.Dim`-
  Umwicklung (D06: das aeussere Wrap wuerde die inneren Resets der
  jetzt vor-eingefaerbten Header/Footer-Strings zunichte machen -- siehe
  `footer()`s eigener Doc-Kommentar). Legitim betroffen (Header!), wie im
  Auftrag vorab angekuendigt -- KEIN chrome.golden-ERRATUM-Praezedenzfall
  wie in epic-E7-plan.md Task 7 Step 8 (dort blieb chrome.golden
  unveraendert, hier explizit NICHT, weil footer()/breadcrumb() selbst
  editiert wurden, nicht nur die Aufrufer).

## Smoke (tmux, real -- nicht nur Unit-Level)

80 Spalten (`tmux new-session -x 80 -y 24`):
- Header EIN Zeile, kein Wrap: `ctrl+k commands · p repos · ? help · q quit`,
  `ctrl+k`/`p`/`?`/`q` sichtbar Teal, Woerter Subtext.
- Footer Browse 2 Zeilen, Q06-Reihenfolge, kein `:`.
- Footer Backlog 3 Zeilen (14 Eintraege inkl. Sort -- siehe Deviations).
- Backlog-Suchzeile: `⌕ / search · sort status` (default) -> nach `S`:
  `⌕ / search · sort prio` (PO-Beispiel exakt getroffen).
- `S Sort` bleibt nach dem NBSP-Fix als EIN Block zusammen (vorher am
  80-Spalten-Umbruch getrennt beobachtet -- der Bug, der zum NBSP-Fix
  fuehrte).

120 Spalten: Header 1 Zeile, Footer Browse 2 Zeilen (13 Eintraege passen
knapp nicht auf 1 Zeile).

Interaktion: `r` oeffnet den Blocking-Picker (Modal-Titel "Blocking"
sichtbar) · `esc` schliesst, danach `B` ist ein exakter No-Op (Frame
byte-identisch zum Zustand vor `B`, per `diff` verifiziert) · `t` oeffnet
Tag-Picker, aeussere Footer-Zone-3 zeigt `n New tag` (Q06/B14 intakt) ·
`p` -> Lobby zeigt `esc back  q quit` (neue Optik, kein `:`) rechts vom
weiterhin `:`-formatierten `i/k:↑↓  enter:open  type:filter`-Praefix
(bewusst nur lobbyExitHint() angeglichen) · `?` Help-Overlay zeigt
gekuerzte Labels (`focus in`, `search`, `Status`, `Tags`, `Parent`,
`Blocking` fuer `r`) -- Single-Source-Beleg, eigener Render-Pfad
(theme.Header/theme.Dim, NICHT renderBindings) unveraendert.

## Deviations/ERRATA

- **ERRATUM (Planner-Sketch-Ungenauigkeit, bean bt-d8kc D02-Abschnitt)**:
  der Code-Sketch `line += "  · " + sortSuffix` (zwei Leerzeichen vor `·`)
  weicht vom PO-Beispiel selbst ab (`"⌕ / search · sort prio"`, EIN
  Leerzeichen je Seite). Implementiert nach dem PO-Beispiel (ein Leerzeichen),
  nicht nach dem Sketch.
- **ERRATUM (Farbwahl-Entscheidung, nicht im bean vorgegeben)**: `theme.
  BindingKey`/`theme.BindingDesc` sind NEUE, eigene Tokens statt `theme.
  HeaderInactive` (B06, bt-czpf) wiederzuverwenden, obwohl beide aktuell
  Teal sind. Begruendung: `HeaderInactive`s PO-Sign-off ist noch PENDING
  (bt-czpf Akzeptanz-Checkliste) -- ein Rollback dort (Teal->Muted) haette
  bei Token-Sharing versehentlich auch D06s bereits ENTSCHIEDENE
  Header/Footer-Farbe mitgerissen. `BindingKey`/`BindingDesc` referenzieren
  weiterhin die BESTEHENDEN `Teal`/`Subtext`-Farbwerte (keine neue Hex-
  Literale dupliziert).
- **NICHT im Plan antizipiert (live im Smoke gefunden, eigener RED/GREEN-
  Zyklus)**: D06s Wechsel von `":"` (atomar, nie umbruchbar) zu einem
  echten Leerzeichen zwischen Taste/Wort machte `footer()`s Wordwrap
  angreifbar -- eine Taste konnte allein auf einer Zeile landen, ihr
  Wort auf der naechsten (beobachtet: Backlog-Footer 80 Spalten, `S`/
  `Sort` getrennt). Fix: `renderBindings()` glued Taste+Wort (und
  mehrwoertige Descs) mit U+00A0 NBSP -- `x/ansi`s `Wordwrap` schliesst
  `nbsp` explizit von seinen Breakpoints aus (verifiziert im
  Package-Quellcode, `wrap.go:195`/`348`). Neue Regressionstests:
  `TestRenderBindingsKeepsKeyAndDescTogetherAcrossWrap`,
  `TestRenderBindingsKeepsMultiWordDescTogether`.
- **Beobachtung, nicht gefixt (kein PO-Auftrag)**: Backlog-Footer spannt
  bei 80 Spalten 3 Zeilen (13 Q06-Eintraege + Sort = 14, knapp ueber die
  2-Zeilen-Kapazitaet), Browse-Footer bleibt bei 2. D06 sagt "Footer
  DARF 2 Zeilen" -- keine explizite Ober-Grenze "NIE mehr als 2". Da (a)
  Sort bereits als dokumentierte Planner-Entscheidung erhalten bleibt und
  (b) die Geometrie (`clickPaneGeometry`) bereits dynamisch auf die
  tatsaechliche Footer-Hoehe reagiert (kein Overflow/keine Korruption,
  nur ein sauber umbrochener 3. Zeilen), keine Code-Aenderung vorgenommen
  -- an bt-6ppq zur VQA-Ruecksprache mit dem PO weitergereicht (siehe
  unten).
- **Kein separater Commit je Q06/D04+D06/D02** (Vorgehens-Empfehlung im
  Auftrag): alle vier Punkte teilen sich `renderBindings`/`footer`/
  `breadcrumb` (view.go) und alle drei Goldens. Ein `git stash`-Versuch,
  die Q06-Remap-Aenderung (`keymap.go`) isoliert zu committen, bestaetigte
  empirisch, dass der volle Testlauf auf diesem Zwischenstand ROT waere
  (`TestBrowseRepoChromeHeaderShowsAllSevenGlobals` etc. -- die
  view_*_test.go-Dateien erwarten bereits das NEUE Format). Ein einziger,
  vollstaendig gruener Commit wurde einem kuenstlich aufgespaltenen,
  zwischenzeitlich roten Verlauf vorgezogen.

## Notes for bt-6ppq (Abschluss-Task)

- **VQA-Ruecksprache noetig**: Backlog-Footer 3 Zeilen bei 80 Spalten (siehe
  Deviations oben) -- PO-Sign-off einholen, ob das akzeptabel ist oder ein
  Folge-Fix (z.B. kuerzeres Sort-Label) gewuenscht wird.
- **B06-Experiment-Kopplung pruefen**: falls bt-czpf/B06 (Accordion-Header
  Teal-Experiment) VOR bt-6ppqs Voll-Validierung noch abgelehnt wird
  (Rollback `theme.HeaderInactive` -> `theme.Muted`), betrifft das NICHT
  `theme.BindingKey`/`theme.BindingDesc` (bewusst entkoppelte Tokens,
  siehe Deviations) -- Header/Footer-Tasten bleiben Teal unabhaengig vom
  B06-Ausgang. Beim Voll-Review kurz gegenpruefen, dass diese Entkopplung
  tatsaechlich hielt (kein versehentlicher Shared-Token-Refactor
  dazwischen).
- **NBSP in Goldens**: `tree.golden`/`backlog.golden` enthalten jetzt
  rohe U+00A0-Bytes zwischen Taste und Wort (unsichtbar im Terminal,
  aber ein `grep`/Diff-Tool ohne UTF-8-Bewusstsein kann das verwirren).
  Kein Bug -- Absicht (siehe Deviations, NBSP-Fix).
- **Lobby-Hint bleibt gemischt-optisch**: `i/k:↑↓  enter:open  type:filter`
  (Alt-Format) + `esc back  q quit` (Neu-Format) auf DERSELBEN Zeile --
  bewusst so belassen (bt-1u0t Notes-fuer-bt-d8kc scopte das explizit auf
  `lobbyExitHint()` allein). Falls der PO bei der Voll-Abnahme eine
  durchgehend einheitliche Optik erwartet, waere das ein Folge-Ticket,
  kein Bug dieses Tasks.
