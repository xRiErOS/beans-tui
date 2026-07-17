# Epos E12 — Nacharbeitsrunde (US-07-Reopen + E11-Review-Nebenbefunde)

Liefert die zweite Nacharbeitsrunde nach E11: ein Reopen (`bt-9ipw`, US-07 vom PO
2026-07-17 abgelehnt) plus sechs Follow-Up-beans aus dem E11-Review — drei neu
gesammelt unter `bt-5uzr` (2026-07-17), drei ältere offene Reviewer-/Investigator-
Funde. Sieben Items, KEIN neues Parent-Epic — jedes bean bleibt an seinem
bestehenden Parent (`bt-362n` für `bt-9ipw`, `bt-5uzr` für `bt-2p9m`/`bt-2kfl`/
`bt-81f0`, `bt-tct9` für `bt-se4q`/`bt-gdkx`, `bt-l8e7` steht parentlos), dieser
Plan bündelt nur die Ausführungsreihenfolge.

Quellen: `bt-9ipw` (vollständiger Body inkl. „Review 2026-07-17 — US-07
REJECTED"-Fix-Prelude) · `bt-5uzr` + Kinder `bt-2p9m`/`bt-2kfl`/`bt-81f0` (PO-
Zitate verbatim, 2026-07-17) · `bt-se4q` (Reviewer-Finding aus `bt-b0w0`-Review)
· `bt-l8e7` (Reviewer-Finding aus `bt-9ipw`-Review) · `bt-gdkx` (US-08-Historie,
zuletzt „Klarstellung 2026-07-16", tag `rejected`) · `bt-tct9` (B05/Q02-Kontext,
Zeilen 64-234) · `docs/plans/v1-port/epic-E11-plan.md` (Stilvorlage) ·
`design-spec.md` §16 (Tag-Management) · Ist-Code-Stand 2026-07-17 (siehe je
Item).

## Item-Übersicht

| # | bean | Typ | Parent | Inhalt | Datei-Familie | blocked_by |
|---|---|---|---|---|---|---|
| 1 | `bt-9ipw` | Feature (reopen) | `bt-362n` | Tag-Picker (`t`) braucht ein IMMER sichtbares Suchfeld, das die Liste live filtert — US-07-Reject | `box_picker_tag.go` (`keyTagPicker`, `tagPickerBox`, `keyTagInput`, `tagInputBox`) | — |
| 2 | `bt-81f0` | Feature | `bt-5uzr` | Notifications vereinheitlichen: Toast als einziger Kanal, untere Status-Zeile entfällt als sichtbarer Notification-Pfad | `view.go` (`statusBar`, `Chrome`), + jede `m.err =`-Schreibstelle (11 Dateien, Inventar unten) | — |
| 3 | `bt-l8e7` | Bug (low) | — | Lobby-Suche: `navKey` verschluckt `i`/`k` vor dem Textinput | `view_lobby.go` (`keyLobby`) | — |
| 4 | `bt-se4q` | Bug (medium, latent) | `bt-tct9` | Relations-Fenster: Cursor-Zentrierung per Glyph-Rescan (`▶`) statt numerischem Index | `view_browse_repo.go` (`windowRelationsSection`, `activeRelationLine`), `view_detail_bean.go` (`relationsSectionBody`) | — |
| 5 | `bt-gdkx` | Bug (verification-only) | `bt-tct9` | US-08 Tags-Sichtbarkeit — Kernfunktion laut Code/Historie bereits erfüllt (D01/B05), einzig offener Rest ist Q02 (schmale Terminals) in `bt-tct9` | `view_detail_bean.go` (`detailHeaderBlock`, nur Lese-Verifikation) | — |
| 6 | `bt-2kfl` | Feature | `bt-5uzr` | Suche mit Filter-Präfixen (`st:completed ty:epic <text>`) | `view_browse_repo.go` (`beanMatchesSearch`, `treeSearchLine`, neue Parse-Funktion), `box_filter_facets.go` (Facet-Schreibzugriff) | — |
| 7 | `bt-2p9m` | Feature | `bt-5uzr` | Filter-Menü als Querformat mit Tab-Kategorien | `box_filter_facets.go` (`keyFilterMenu`, `treeFilterBox`) | — |

**Kein neues Parent-Epic:** alle sieben beans bleiben an ihren jeweiligen
Ur-Epics/Ur-Beans — dieser Plan ist reine Ausführungs-Choreografie.

## Reihenfolge-Begründung

- **#1 (`bt-9ipw`) und #2 (`bt-81f0`) zuerst — PO-Vorgabe, bindend.**
  Reihenfolge UNTEREINANDER (Planner-Entscheidung, PO nannte keine): `bt-9ipw`
  vor `bt-81f0`. Begründung: `bt-9ipw` ist ein Reopen einer bereits laufenden
  Review-Kette (US-07 REJECTED, höchste Dringlichkeit — ein Feature gilt laut
  PO als „unbenutzt ohne Sichtbarkeit der Eingabe"), komplett in EINER Datei
  (`box_picker_tag.go`) lösbar, kein Fächer-Umbau. `bt-81f0` ist dagegen
  repo-weit (11 Dateien, siehe Item-Detail) — u. a. berührt es
  `box_picker_tag.go:425` (die `applyTagPickerDiff`-Fehlermeldung), aber eine
  ANDERE Funktion als `bt-9ipw`s Scope (`keyTagPicker`/`tagPickerBox`/
  `keyTagInput`/`tagInputBox`) — kein echter Zeilen-Konflikt, aber dieselbe
  Datei-Nachbarschaft; `bt-9ipw` zuerst hält diese Nachbarschaft klein. Ferner
  bauen mehrere spätere Items (#4, #5, #6) auf `view_browse_repo.go`/
  `view_detail_bean.go`-Stellen auf, die `bt-81f0`s Statuszeilen-Entfernung
  mit-berührt (Aufruf-Stelle `statusBar(indicator, m.err, innerW)`,
  `view_browse_repo.go:954`) — diese Items sollen gegen den BEREITS
  bereinigten Stand arbeiten, nicht gegen einen, der später nochmal
  umgeschrieben wird.
- **#3 (`bt-l8e7`) direkt danach, nicht Teil der PO-Vorgabe, aber
  Planner-Empfehlung:** kleinster, isoliertester Fix dieser Runde (EINE
  Datei, `view_lobby.go`, keine Überschneidung mit irgendeinem anderen Item),
  UND der Fix-Pattern (rohe `tea.KeyUp`/`tea.KeyDown`-Interception statt
  `navKey()`-Aliasing) ist exakt das, was `bt-9ipw` gerade in
  `box_picker_tag.go` gebaut hat (`keyTagInput`, dort bereits mit Test
  `TestTagInputArrowKeysDoNotLeakIntoTypedText` abgesichert) — ein
  Implementer, der `bt-9ipw` gerade fertiggestellt hat, trägt den Kontext
  bereits frisch mit (mirrort `epic-E11-plan.md`s Begründung für Item 7).
  Verletzt NICHT die PO-Vorgabe „9ipw und 81f0 zuerst" — beide bleiben
  Rang 1/2, `bt-l8e7` ist Rang 3, vor dem Rest der Runde.
- **#4→#5 (`bt-se4q`→`bt-gdkx`), empfohlene Cluster-Reihenfolge, KEIN hartes
  `blocked_by`:** beide berühren `view_detail_bean.go`, aber unterschiedliche
  Funktionen (`relationsSectionBody` vs. `detailHeaderBlock`) — aus
  Dateikonflikt-Vorsicht empfohlen NACHEINANDER in einer Session (mirrort
  `epic-E11-plan.md`s Items 2-4-Cluster-Muster), kein echter
  Zeilen-Konflikt. `bt-se4q` zuerst: ein echter (wenn auch aktuell latenter)
  Bug mit klarem Fix-Rezept vom Reviewer. `bt-gdkx` danach: reine
  Verifikation ohne erwarteten Code-Change (siehe Item-Detail) — inhaltlich
  eher eine Bestandsaufnahme als ein Fix, daher NIEDRIGES Konflikt-Risiko,
  aber gleiche Datei-Familie wie `bt-se4q`, deshalb im Cluster belassen statt
  irgendwo isoliert eingeschoben.
- **#6 (`bt-2kfl`) vor #7 (`bt-2p9m`), beide unabhängig ausführbar (VERSCHIEDENE
  Dateien: `view_browse_repo.go`/`box_filter_facets.go`-Lesezugriff vs.
  `box_filter_facets.go`-Schreibzugriff), Planner-Empfehlung dennoch
  sequenziell statt parallel:** beide manipulieren dieselben vier
  Facet-State-Maps (`m.filterStatus`/`m.filterType`/`m.filterPriority`/
  `m.filterTag`) über unterschiedliche Eintrittspunkte (Such-Präfix-Parser
  vs. Tab-Filter-Menü) — kein Zeilen-Konflikt, aber `bt-2kfl`s eigener
  Bean-Entwurf nennt explizit eine offene Frage („Sync-Richtung
  Suche↔Facetten-Overlay") — sinnvoller VOR `bt-2p9m` zu klären, damit der
  Such-Präfix-Parser nicht gegen ein sich währenddessen änderndes
  Tab-Layout entwickelt wird. `bt-2kfl` ist außerdem `view_browse_repo.go`-
  lastig (gleiche Datei-Familie wie #4), `bt-2p9m` ist komplett isoliert
  (nur `box_filter_facets.go`) — die größte, am saubersten abgegrenzte
  Einzelaufgabe dieser Runde kommt bewusst zuletzt.

## Item 1: Tag-Picker (`t`) braucht sichtbares Suchfeld (`bt-9ipw`, US-07-Reopen)

**Root Cause.** `box_picker_tag.go` hat heute ZWEI getrennte Overlay-Zustände:
der Haupt-Picker (`tagPickerBox`, `keyTagPicker`, Zeile 203-229/444-473) ist ein
reines Multi-Select (space/x togglet, kein Textinput, kein Filter) — Tippen
tut dort schlicht NICHTS (kein `case` in `keyTagPicker`s `switch`). Erst der
separate `n`-Submodus (`tagInputActive`, `openTagInput`/`keyTagInput`/
`tagInputBox`, Zeile 254-389/475-511, aus `bt-9ipw`s EIGENER erster
Umsetzungsrunde) hat ein sichtbares `textinput.Model` (`m.tagInput.View()`,
Zeile 489) UND Live-Filterung (`filterTagItems`, Zeile 277-289) UND
Pfeiltasten-Navigation über die gefilterten Vorschläge
(`tagInputSuggestCursor`). PO-Zitat („Ich sehe, dass der Selektor wandert,
wenn ich Tippe. Aber ich habe kein visuellen Hinweis darauf, WAS ich
getippt habe.") beschreibt exakt diese Lücke am Haupt-Picker: der PO tippt
direkt im `t`-Picker (kein sichtbares Suchfeld dort, nur Navigation über
`m.menu`), NICHT den bereits vorhandenen `n`-Submodus mit eigenem Feld.

**Vorgehen (D01, Planner-Entscheidung, löst die PO-Empfehlung „ggf.
Konsolidierung beider Modi in einen" aus dem eigenen Fix-Prelude auf):**
Repro-first PFLICHT (Fix-Prelude fordert das explizit) — beide Einstiege
(`t` direkt tippen vs. `t`→`n`) im tmux prüfen, um zu bestätigen, dass der
PO tatsächlich im Haupt-Picker getippt hat, nicht im `n`-Submodus. Danach:
**EIN konsolidierter Modus statt zwei getrennter** — der Haupt-Picker
bekommt ein IMMER sichtbares, dauerhaft fokussiertes Suchfeld (dasselbe
`m.tagInput`-Widget, dieselbe `filterTagItems`-Funktion, wiederverwendet aus
der bestehenden Implementierung) direkt in `tagPickerBox`, oberhalb der
Checkbox-Liste. Tippen filtert die angezeigte Liste live (Substring, wie
heute in `filterTagItems`), Pfeiltasten navigieren weiterhin den Cursor über
die (jetzt gefilterte) Liste, space/x togglet weiterhin Multi-Select
(unverändert — das ist der zentrale Unterschied zum bisherigen `n`-Submodus,
dessen `enter` sofort einen einzelnen Tag übernahm UND schloss). `n`/kein
Treffer bleibt der Eskalationspfad zur Neuanlage (heutiges
`keyTagInput`-Enter-ohne-Treffer-Verhalten, D11 unverändert: reiner
Zuweisungs-Akt, kein Registry-Write für undefinierte Tags). Der separate
`tagInputActive`-Zwei-Stufen-Einstieg (erst `t`, dann extra `n` um überhaupt
tippen zu können) entfällt — Implementer entscheidet, ob `tagInputBox` als
eigene Funktion verschwindet (in `tagPickerBox` aufgeht) oder als
Compat-Layer für den reinen Neuanlage-Fall bestehen bleibt.

**Akzeptanz:**
- [ ] Repro-Ergebnis dokumentiert (welcher Einstieg zeigte das Problem)
- [ ] `t` öffnet den Picker MIT sofort sichtbarem, fokussiertem Suchfeld (kein
      zweiter Tastendruck nötig)
- [ ] Tippen filtert die Liste sichtbar, Eingabe ist als Text im Feld zu sehen
- [ ] Pfeiltasten navigieren die gefilterte Liste, space/x toggelt Multi-Select
      unverändert (mehrere Tags gleichzeitig anhakbar)
- [ ] `enter` ohne Treffer legt neu an (D11 unverändert), Zuweisung ans
      aktuelle Bean bleibt bestehen
- [ ] Tag-Management-Page (`view_tag_management.go`) UNANGETASTET
- [ ] Test-Suite grün, Golden-Gegenbeleg falls Overlay-Breite sich ändert
- [ ] tmux-Smoke: Tippen zeigt Text im Feld, Filterung sichtbar, PO-Wortlaut
      exakt nachvollzogen

## Item 2: Notifications vereinheitlichen — Toast als einziger Kanal (`bt-81f0`)

**Root Cause.** Zwei parallele Feedback-Kanäle: (a) `m.showToast(...)`
(oben rechts, `overlay_show_toast.go`) und (b) `m.err string` (unten rechts,
gerendert über `statusBar(indicator, errNote, width)`, `view.go:70-83`, PO
selbst zitiert exakt diesen Fall: „beans list: exit status 1: Error:
querying beans: syntax Error" — das ist `applyBleveResult`s Fehlerpfad,
`update.go:852-856`). Vollständiges Inventar jeder `m.err =`-Schreibstelle
(11 Dateien):

| Datei:Zeile | Kontext | Toast bereits vorhanden? |
|---|---|---|
| `update.go:280/303`, `:305/324`, `:418/427`, `:547/552`, `:616/618`, `:634/641`, `:648/653`, `:735/740`, `:853/855`, `:874/876` | `applyMutationResult`, `applyBeanRawLoaded`, `applyCreateDone`, `applyEditorFinished` (2×), `applyBleveResult`, `applyLoaded` | JA — bereits Dual-Write (Toast + `m.err`), reine Redundanz |
| `box_confirm_delete.go:130` | Bean während Delete-Confirm verschwunden | NEIN — nur `m.err` |
| `box_confirm_create.go:48/66/91/96` | Create-in-flight-Guard, Settings-Save/-Reload-Fehler | NEIN — nur `m.err` |
| `box_menu_value.go:191` | Bean während Value-Menu verschwunden | NEIN — nur `m.err` |
| `box_picker_parent.go:142` | Bean während Parent-Picker verschwunden | NEIN — nur `m.err` |
| `box_picker_blocking.go:161` | Bean während Blocking-Picker verschwunden | NEIN — nur `m.err` |
| `box_picker_tag.go:425` | Bean während Tag-Picker-Diff verschwunden | NEIN — nur `m.err` |
| `overlay_palette.go:232` | Create-in-flight-Guard (Palette-Pfad) | NEIN — nur `m.err` |

Die "NEIN"-Zeilen sind der eigentliche Bug hinter dem PO-Befund: diese sieben
Stellen zeigen HEUTE nur die untere Zeile, GAR KEINEN Toast — eine blinde
Entfernung der Zeile ohne Gegenmaßnahme würde sie komplett stumm schalten.

**D02 (Planner-Entscheidung, Scope-Begrenzung):** `m.err` als Feld
(`types.go:145`) bleibt BESTEHEN — 43 Testassertions über 16 Testdateien
referenzieren es direkt (`grep -rn ".err ==\|.err !=" internal/tui/*_test.go`
zählt 43 Treffer); eine vollständige Feld-Entfernung wäre ein Test-Rewrite
weit über den PO-Auftrag hinaus (YAGNI). Stattdessen: `m.err` bleibt der
interne "was ist zuletzt schiefgelaufen"-Zustand (Tests dürfen weiter darauf
prüfen), verliert aber JEDE Rendering-Anbindung — `ErrNote`
(`ChromeOpts`, `view.go:255`) und die drei direkten `statusBar(indicator,
m.err, innerW)`-Aufrufe (`view_browse_repo.go:954`,
`view_tag_management.go:272`, `view_browse_backlog.go:286`) hören auf,
`m.err` zu lesen. Jede der sieben "NEIN"-Stellen bekommt einen NEUEN
`showToast(toastError, ..., "", nil, false)`-Aufruf (Titel = der bisherige
`m.err`-Text) direkt neben der bestehenden `m.err =`-Zeile, mirrort exakt das
Muster, das `update.go` an den zehn "JA"-Stellen schon zeigt.

**Offene Layout-Frage (Q1, siehe unten):** `statusBar` rendert HEUTE nicht
nur `errNote`, sondern AUCH den Scroll-Indikator (`ind`, z. B. bei Chrome/
Lobby/Help — `view.go:283`, `renderer`-Kommentar Zeile 71). Die reservierte
Zeile kann also nicht ersatzlos verschwinden, ohne den Scroll-Indikator zu
verlieren — Planner-Empfehlung siehe Q1.

**Akzeptanz:**
- [ ] Inventar oben vollständig abgearbeitet: alle zehn Dual-Write-Stellen
      verlieren NUR die Rendering-Anbindung (Toast bleibt, `m.err`-Zuweisung
      darf bleiben, da Tests darauf prüfen)
- [ ] Alle sieben bisher-stummen Stellen bekommen einen neuen `showToast`-Ruf
      (kein Fehler wird leiser als vorher)
- [ ] `ErrNote`/die drei `statusBar(..., m.err, ...)`-Aufrufe verlieren die
      `m.err`-Anbindung (Signatur-Entscheidung gemäß Q1-Antwort)
- [ ] Kein bestehender Test, der auf `m.err`-Werte prüft, bricht (Feld bleibt)
- [ ] Golden-Regen falls Layout-Höhe sich ändert:
      `go test ./internal/tui/ -run TestChromeGolden -update`,
      `-run TestTreeGolden -update`, `-run TestBacklogGolden -update`
      (Diff vor dem Commit sichten, nicht blind akzeptieren)
- [ ] tmux-Smoke: ein Bleve-Syntax-Fehler (PO-Repro: `/` + ungültige Bleve-
      Query) zeigt NUR noch den Toast, keine zweite Meldung unten rechts

## Item 3: Lobby-Suche verschluckt `i`/`k` (`bt-l8e7`)

**Root Cause.** `keyLobby` (`view_lobby.go:350-359`) prüft
`navKey(msg.String())` (Zeile 351) VOR der Textinput-Weiterleitung
(Zeile 400-408). `navKey` (`keymap.go:232-244`) aliast `keys.Up`/`keys.Down`
auf Buchstaben inkl. `i`/`k` (vim-Stil) — ein Repo-Query, der mit `i` oder
`k` beginnt, bewegt den Cursor statt ins Suchfeld zu gehen. Exaktes Vorbild
für den Fix bereits im selben Repo umgesetzt: `keyTagInput`
(`box_picker_tag.go:313-335`, aus `bt-9ipw`) fängt bewusst die ROHEN
`tea.KeyUp`/`tea.KeyDown`-`KeyType`s ab statt über `navKey()`s
Buchstaben-Alias-Tabelle zu gehen — exakt dieses Muster verhindert dort, dass
getippte "i"/"k" verschluckt werden.

**Vorgehen:** `keyLobby`s `switch navKey(msg.String())`-Block (Zeile
351-359) durch eine `switch msg.Type`-Prüfung auf `tea.KeyUp`/
`tea.KeyDown` ersetzen (mirrort `keyTagInput` 1:1) — jede andere Taste
(inkl. literal getippter "i"/"k") fällt weiter zur bestehenden
`m.repoSearch.Update(msg)`-Zeile (400-408) durch.

**Akzeptanz:**
- [ ] Query "ide" (oder jedes andere mit `i`/`k` beginnende Wort) landet
      VOLLSTÄNDIG im Suchfeld, kein verschlucktes Zeichen
- [ ] Pfeiltasten (`↑`/`↓`) navigieren die Repo-Liste weiterhin unverändert
- [ ] Regressionstest analog `TestTagInputArrowKeysDoNotLeakIntoTypedText`
      (neuer Name z. B. `TestLobbyQueryDoesNotSwallowArrowAliasLetters`)
- [ ] Test-Suite grün

## Item 4: Relations-Fenster — numerischer Cursor-Index statt Glyph-Rescan (`bt-se4q`)

**Root Cause.** `activeRelationLine` (`view_browse_repo.go:721-728`) sucht
die aktive Zeile per `strings.Contains(l, "▶")` im BEREITS gerenderten
String — unverankert, erste Fundstelle gewinnt. Reviewer-Beleg (`bt-b0w0`-
Review): eine gerenderte Zeile mit `▷ … Titel mit ▶ …` VOR der echten
aktiven Zeile zentriert das Fenster falsch. Aktuell NICHT scharf (kein
Bean-Titel im Repo enthält `▶`/`▷`, grep-verifiziert vom Reviewer), aber
latent. Die eigentliche Information — WELCHE Zeile aktiv ist — ist zum
Konstruktionszeitpunkt in `relationsSectionBody`
(`view_detail_bean.go:399-460`) bereits als exakter Integer bekannt: die
Funktion führt bereits einen laufenden Zeilenzähler `gi` (Parent=0,
Children=1..n, Blocking=n+1..m, Blocked By=m+1..k) UND weiß pro Zeile, ob
sie via `hangingIndentWrap` auf mehrere Display-Zeilen umbricht — genau die
zwei Informationen, die `activeRelationLine`s nachträglicher String-Scan
NICHT rekonstruieren kann, sondern nur über den (fehleranfälligen)
Glyphen-Fund errät.

**Vorgehen:** `relationsSectionBody` um einen dritten Rückgabewert
erweitern (z. B. `activeLine int`, Default 0), der die Display-Zeilen-Position
der aktiven Zeile MITZÄHLT während die Gruppen aufgebaut werden (inkl.
Subheader-Zeilen "Parent"/"Children"/... und ggf. mehrzeiliger
`hangingIndentWrap`-Ausgabe pro Zeile). Der neue Wert läuft durch
`beanSections` (`view_detail_bean.go:71-89`, `accordionSection`-Struct
braucht ein neues Feld oder einen Parallel-Return) bis zu
`renderAccordionPane`/`windowRelationsSection`
(`view_browse_repo.go:629-668`), wo `activeRelationLine(lines)` durch den
durchgereichten numerischen Wert ersetzt wird. Implementer wählt die genaue
Durchreichungs-Mechanik (drittes Struct-Feld vs. Parallel-Slice) — Vorgabe
ist NUR: kein String-Rescan des gerenderten Outputs mehr für diesen Zweck.

I01/I02 (Reviewer, sekundär, gleiche Funktion) mitnehmen: I01 (Fenster zählt
Display-Zeilen post-Wrap, eine nicht-aktive gewrappte Zeile kann am
Fensterrand ohne Fortsetzungshinweis abgeschnitten werden), I02 (`winH==1`:
"Children"-Subheader fällt aus dem Fenster trotz mathematisch korrekter
Top-Position) — je bewerten: fixen ODER begründet als Won't-fix dokumentieren
(kein Zwang zu beidem).

**Akzeptanz:**
- [ ] `activeRelationLine`/Glyph-Rescan durch numerischen, aus der Body-
      Konstruktion durchgereichten Index ersetzt
- [ ] Regressionstest: ein Relations-Eintrag mit `▶`/`▷` IM TITEL zentriert
      das Fenster weiterhin korrekt auf die tatsächlich aktive Zeile (nicht
      auf den Titel-Glyphen)
- [ ] I01/I02 bewertet, Ergebnis (Fix oder Won't-fix + Begründung) dokumentiert
- [ ] Bestehende NB-2/`bt-b0w0`-Akzeptanz (Fenster/Auto-Scroll/Indikator)
      bleibt unverändert grün
- [ ] tmux-Smoke mit relations-reichem Bean (`bt-apmy`)

## Item 5: US-08 Tags-Sichtbarkeit — Verifikation, kein erwarteter Fix (`bt-gdkx`)

**Kein Fix ohne neuen Befund — Verifikations-Task.** Historie zeigt: die
Kernanforderung ("Tags sichtbar in Tree/Detail") ist bereits ZWEIFACH gelöst
— (1) D01 (Epic `bt-ntoz`, Commit 397a70f): Tags als 7.-dann-5.-Meta-Feld
(nach `bt-lg68`s Kürzung auf 5 Felder, `metaFields`, `view_detail_bean.go`);
(2) B05 (Epic `bt-tct9`, redefiniert 2026-07-16): Tags ZUSÄTZLICH im
Kopfblock-Meta-Strip (`detailHeaderBlock`, `view_detail_bean.go:193-227`,
Zeile 219 `theme.Muted.Render("    tags: ") + tags`). US-04 (Kopfblock-Tags)
wurde vom PO in Runde 2 explizit AKZEPTIERT (`bt-tct9` Zeile 216). Der EINZIGE
noch offene Rest ist Q02 (`bt-tct9` Zeile 165-178, Reviewer-Finding
`bt-mtig`): bei ~100 Spalten Gesamtbreite (Split-View, `accW≈61`) füllt die
Kopfblock-Zeile bereits 59/61 Zeichen OHNE Tag-Inhalt — `truncate(
typeStatusPrio, w)` (`view_detail_bean.go:224`) schneidet die Tag-SPALTE
dann hart ab (Label `tags:` sichtbar, Wert meist nicht). Erst ab breiteren
Terminals (~160 Spalten) erscheint das PO-Mockup-Bild vollständig. Dieser
Reviewer-Fund ist BEREITS als offene PO-Frage in `bt-tct9` dokumentiert und
per PO-Entscheid „bis zur Antwort keine Nacharbeit" (Zeile 179) gesperrt.

**Vorgehen:** Nur Bestandsaufnahme, kein Code-Change in diesem bean.
Implementer liest `detailHeaderBlock` live (nicht nur Code, auch ein
tmux-Smoke bei Standard-Split-View-Breite) und bestätigt den oben
dokumentierten Zustand, verlinkt explizit `bt-tct9`s offene Q02 als den
tatsächlichen verbleibenden Rest. KEINE Status-/Tag-Änderung an `bt-gdkx`
selbst (Supervisor-Sache) — der Body-Append dokumentiert nur den
Verifikations-Befund und empfiehlt dem PO, `bt-gdkx` gemeinsam mit `bt-tct9`s
Q02 zu schließen, sobald diese beantwortet ist (bean sagt das bereits selbst:
„Dieses bean wird mit E9-B05 gemeinsam geschlossen").

**Akzeptanz:**
- [ ] Live-Verifikation (tmux, `bin/bt`, Standard-Split-View) bestätigt: Tags
      sichtbar in META (5. Feld) UND im Kopfblock (Label mindestens, Wert bei
      ausreichender Breite)
- [ ] Q02-Zusammenhang (`bt-tct9` Zeile 165-178) im Body-Append explizit
      referenziert, kein doppeltes/drittes Tracking derselben Frage
- [ ] KEIN Code-Change, KEIN Status-/Tag-Wechsel an diesem bean

## Item 6: Suche mit Filter-Präfixen (`bt-2kfl`)

**Root Cause / Ist-Stand.** `beanMatchesSearch` (`view_browse_repo.go:
251-260`) liest `m.searchQuery` heute als reinen Freitext (lokaler
Substring-Fallback unter 3 Zeichen, Bleve-Ergebnis ab 3 Zeichen via
`m.searchBleveIDs`) — KEINE Präfix-Erkennung. Facetten (`status`/`type`/
`priority`/`tag`) leben in eigenen Maps (`m.filterStatus` etc.,
`box_filter_facets.go`), ausschließlich über das Filter-Menü (`f`,
`toggleFacet`) beschreibbar. `beanMatches` (`box_filter_facets.go:189-191`)
UND-verknüpft Suche und Facetten bereits — die Infrastruktur zum Kombinieren
existiert, es fehlt nur der Parser, der `/`-Eingabetext in Facet-Writes
UND Rest-Suchtext aufteilt. `filterSummary` (`box_filter_facets.go:
324-339`) rendert bereits exakt das vom PO zitierte Format
("St:completed Ty:epic") für den Tree-Kopf (`treeSearchLine`,
`view_browse_repo.go`) — die Kürzel-Konvention (`St:`/`Ty:`/`Pr:`/`Tags:`)
ist also bereits die kanonische Quelle für die vom PO gewünschten
Präfixe (`st:`/`ty:`/`pr:`/`tag:`, case-insensitiv laut PO-Akzeptanz).

**Offene Design-Fragen (siehe Q2/Q3 unten) — NICHT vom Planner
vorentschieden, da PO-Wortlaut selbst zwei ungeklärte Punkte benennt.**
Implementer beginnt mit den Q2/Q3-Antworten (falls vor Umsetzungsstart
eingeholt) oder dokumentiert die getroffene Annahme explizit im Bean-Summary,
falls ohne PO-Antwort begonnen werden muss.

**Vorgehen (Gerüst, Details Implementer-Wahl nach Q2/Q3):** neue Parse-
Funktion (z. B. `parseSearchPrefixes(query string) (facets map[string][]string,
rest string)`) tokenisiert `m.searchQuery` bei jedem Tastendruck (oder bei
Commit — abhängig von Q3) nach `<prefix>:<value>`-Paaren (case-insensitiv,
Kürzel-Set `st`/`ty`/`pr`/`tag`, mirrort `facetHead`s vier Facetten MINUS
`archive`, das hat laut PO-Zitat keinen Präfix). Ungültige Präfixe/Werte
fallen in `rest` (kein Fehler, PO-Akzeptanz explizit). `rest` geht weiter
durch den bestehenden `beanMatchesSearch`-Pfad, die geparsten Präfixe
werden — je nach Q2-Antwort — entweder direkt in `beanMatchesFacets`
eingespeist (Parser-Resultat additiv zu den Menü-Facetten) oder schreiben
`m.filterStatus`/`m.filterType`/etc. direkt (dann sync mit dem `f`-Menü,
`facetOn`/`treeFilterBox` zeigen die getippten Filter live an — das ist die
vom PO gewünschte „St:completed Ty:epic"-Kopfzeilen-Spiegelung).

**Akzeptanz:**
- [ ] `/st:completed ty:epic foo` filtert auf `status=completed AND
      type=epic AND` Text-Match `foo`
- [ ] Präfixe case-insensitiv, Reihenfolge egal
- [ ] Ungültiges Präfix/Wert → als normaler Suchtext behandelt, kein Fehler
- [ ] Header-Filteranzeige (`filterSummary`) spiegelt die getippten Filter
      (eine Wahrheit mit Facetten-State, gemäß Q2-Antwort)
- [ ] Q2/Q3-Antworten (oder dokumentierte Annahme) im Bean-Summary festgehalten
- [ ] Test-Suite grün, ggf. neue Tests für `parseSearchPrefixes`

## Item 7: Filter-Menü als Querformat mit Tab-Kategorien (`bt-2p9m`)

**Root Cause / Ist-Stand.** `treeFilterBox` (`box_filter_facets.go:
296-318`) rendert HEUTE eine EINZIGE vertikale Liste über alle fünf Facetten
(`Status`/`Type`/`Priority`/`Tags`/`Archive`, `facetHead`-Reihenfolge,
Zeile 34-40) mit EINEM `m.filterMenu`-Cursor über die volle, geflachte
`m.filterItems`-Liste (`keyFilterMenu`, Zeile 268-292) — bei vielen Tags
entsprechend lang. PO will stattdessen Tabs je Kategorie mit eigenem
Pfeiltasten-Fokus. **Kein Tastenkonflikt:** `tab`/`shift+tab` sind zwar
global an `keys.FocusIn`/`keys.FocusOut` gebunden (`keymap.go:123-124`),
aber `handleKey` prüft `m.filterOpen` (Full-Capture, `update.go:946`) VOR
der `FocusIn`/`FocusOut`-Prüfung (Zeile 1061/1073) — `keyFilterMenu` darf
`tab`/`shift+tab` also gefahrlos für den Kategoriewechsel belegen, exakt wie
`m.searchActive`/`m.tagInputActive` denselben Full-Capture-Vorrang bereits
für ihre eigenen Zwecke nutzen.

**Vorgehen:** `buildFilterItems` liefert bereits fünf klar abgegrenzte
Facet-Gruppen (`items = append(items, ffItem{"status", ...})` etc.,
Zeile 57-81) — Implementer gruppiert `m.filterItems` NACH `facet`-Feld in
Tab-Reihen (kein neuer Datenzustand nötig, nur Rendering + Navigation
umgestellt). Neuer/erweiterter State: ein Tab-Cursor (welche Facet-Gruppe
aktiv) getrennt vom bestehenden `m.filterMenu`-Zeilencursor (der dann NUR
noch innerhalb der aktiven Gruppe läuft, nicht mehr über alle Items). PO-
Klickpfad wörtlich: `f` öffnet mit erstem Tab aktiv (erstes Element
vorselektiert) → `tab`/`shift+tab` wechselt Tab (Fokus springt auf dessen
erstes Element) → `↑`/`↓` navigiert innerhalb des Tabs → `space` togglet
(bestehendes `toggleFacet`, UNVERÄNDERT) → `enter` wendet an/schließt
(bestehendes Verhalten, UNVERÄNDERT). Breiten-Verhalten bei 80 Spalten
beachten (NBSP-Wordwrap-Falle, `docs/LESSONS-LEARNED.md` Eintrag 4,
`docs/SSTD.md`-Pointer) — Querformat mit 5 Tabs braucht ggf. mehr
Terminalbreite als die alte Liste, `clampModalWidth`/`wideModalWidth`
(Zeile 369-399) als bestehende Bausteine prüfen statt neu zu erfinden.

**Akzeptanz:**
- [ ] `f` öffnet Overlay im Querformat, erster Tab aktiv, erstes Element
      vorselektiert
- [ ] `tab`/`shift+tab` wechseln Kategorie-Tabs, kein Konflikt mit globalem
      `FocusIn`/`FocusOut` (Full-Capture bestätigt via Test)
- [ ] Fokus-Tab: `↑`/`↓` navigiert NUR innerhalb der aktiven Kategorie
- [ ] `space`/`x` togglet Kriterium (bestehende Semantik unverändert),
      `enter` wendet an
- [ ] bestehende Filter-Semantik (`beanMatchesFacets`, `filterSummary`)
      unverändert — nur Overlay-Rendering/Navigation betroffen
- [ ] tmux-Smoke bei 80 Spalten (Grenzbreite, NBSP-Falle)
- [ ] Golden-Gegenbeleg falls Overlay-Breite/-Höhe golden-relevant wird

## Offene Fragen für den PO

| # | Frage | Betrifft |
|---|---|---|
| Q1 | `statusBar` rendert HEUTE nicht nur `errNote`, sondern auch den Scroll-Indikator (Chrome/Lobby/Help). Reicht es, NUR die Fehler-Anbindung zu kappen (Zeile bleibt als reiner Scroll-Indikator-Slot bestehen, PO-Wortlaut „reservierte Zeile frei" wird dann nur BEI FEHLENDEM Indikator wörtlich erfüllt), oder wird eine ECHTE Zeilen-Entfernung erwartet (dynamische Footer-Höhe, höheres Jitter-/Golden-Bruch-Risiko)? Planner-Empfehlung: erste Variante (kleinerer, risikoärmerer Schnitt) — außer PO bestätigt explizit den größeren Umbau. | Item 2 (`bt-81f0`) |
| Q2 | Sync-Richtung Suche↔Facetten-Overlay: sollen getippte Präfixe (`st:completed`) das `f`-Menü sichtbar mit-togglen (ein gemeinsamer State), oder bleibt die Such-Präfix-Filterung ein separater, additiver Layer, der NUR die Ergebnisliste beeinflusst, ohne das Menü zu verändern? | Item 6 (`bt-2kfl`) |
| Q3 | Ab welcher Zeichenzahl/wann greift der Präfix-Parser — bei JEDEM Tastendruck (lokal, wie der bestehende <3-Zeichen-Fallback) oder erst ab dem Bleve-Schwellenwert (3 Zeichen)? Berührt, ob Präfixe im async-Bleve-Pfad überhaupt ankommen (Bleve indiziert Titel+Body, keine Facet-Felder laut `client.go`-Doc). | Item 6 (`bt-2kfl`) |

## Selbst-Review

- **Deckung:** alle sieben beans (`bt-9ipw`, `bt-81f0`, `bt-l8e7`, `bt-se4q`,
  `bt-gdkx`, `bt-2kfl`, `bt-2p9m`) je genau einem der sieben Plan-Abschnitte
  zugeordnet, keines doppelt, keines vergessen — Quelle: der vom Supervisor
  übergebene 7-bean-Scope.
- **Struktur:** kein hartes `blocked_by` in dieser Runde gesetzt — jede
  Datei-Nachbarschaft (Item 1↔2 `box_picker_tag.go`, Item 4↔5
  `view_detail_bean.go`, Item 6↔4 `view_browse_repo.go`) wurde geprüft und
  betrifft nachweislich UNTERSCHIEDLICHE Funktionen ohne Zeilen-Overlap
  (mirrort `epic-E11-plan.md`s Parsimonie — dort nur EIN echtes
  `blocked_by` unter sieben Items, hier: keines, weil kein Fall die dortige
  „identische Funktion wird von zwei Tasks erweitert"-Schwelle erreicht).
  Reihenfolge-Zwänge stattdessen als geordnete Item-Nummern + Prosa-
  Begründung kodiert.
- **bean-Qualität:** jeder Abschnitt nennt Root-Cause mit `datei:zeile`,
  konkretes Vorgehen, abhakbare Akzeptanz-Checkboxen mit Artefakten (Test-
  Namen, Kommando-Wortlaut für Golden-Regen) — ein frischer Agent kann ohne
  Rückfrage in JEDEM Item starten. Item 1 und Item 6 markieren explizit, wo
  eine offene PO-Frage die letzte Design-Entscheidung noch offen lässt
  (Item 1 via Repro-first-Pflicht aus dem Fix-Prelude selbst, Item 6 via
  Q2/Q3), damit der Implementer nicht blind eine Annahme baut, die dem
  Review wieder um die Ohren fliegt.
- **Keine Anti-Patterns:** kein Punkt bekommt einen Blind-Fix ohne
  Code-Beleg — Item 5 (`bt-gdkx`) ist bewusst als Verifikations-Task ohne
  erwarteten Code-Change ausgewiesen (Historie zeigt die Kernanforderung
  bereits zweifach gelöst), statt einen dritten Fix-Versuch gegen ein
  bereits gelöstes Problem zu starten.
- **Keine Status-/Tag-Änderungen:** dieser Plan selbst ändert an keinem der
  sieben beans Status oder Tags — jeder Body-Append (separater Schritt) trägt
  nur eine datierte `## Plan-Konkretisierung E12 (2026-07-17)`-Sektion an,
  additiv, kein Body-Overwrite. `bt-gdkx`s bestehendes Tag `rejected` und
  `bt-362n`/`bt-tct9`s Epic-Tags (`rejected`/`accepted`) bleiben unverändert.
