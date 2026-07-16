# LESSONS-LEARNED — beans-tui

Append-only-Log. Einträge werden NIE umgeschrieben oder gelöscht, nur unten
angehängt (neue Einträge ans Ende der jeweiligen Datums-Sektion bzw. eine
neue Sektion). Jeder Eintrag hat exakt drei Felder:

- **Was lief nicht rund** — der beobachtete Fehler/Reibungspunkt, mit Quelle
  (Task/bean/Review-Runde).
- **Wie gefixt** — die konkrete Korrektur (Commit/Mechanik).
- **Forward-Guard** — der Mechanismus mit Zähnen, der die Wiederholung
  verhindert: ein benannter Test, eine CLAUDE.md-Zeile oder ein
  Checklisten-Punkt. Ein Eintrag ohne Guard ist unvollständig.

## 2026-07-16 — E8-Abschluss (bt-6ppq, Quellen: E8-Fix-Runden + Reviews)

### 1. Cross-Pane-clickKey-Aliasing (T4/bt-duz7, Fix-Runde R1)

- **Was lief nicht rund:** `mouseDetailClick`s clickKey-Encoding
  (`secIdx*10+fieldIdx+1`) teilte den Zahlenraum mit den Tree-/Backlog-
  Row-Indizes im GEMEINSAMEN `m.lastClickIdx`-Feld — ein Tree-Klick auf
  Row 1 gefolgt <500ms von einem Erst-Klick aufs `title:`-Feld (Key
  ebenfalls 1) aliaste zu einem falschen Doppelklick und öffnete das
  Title-Edit-Form ohne vorherige Selektion (Reviewer live reproduziert).
- **Wie gefixt:** benannter Helfer `detailClickKey()` mit
  `detailClickKeyBase = 1000`-Offset — Detail-Keys garantiert disjunkt von
  jedem realistischen Row-Index (Commit `6e1152e`).
- **Forward-Guard:** Test
  `TestMouseDetailClickTreeClickIndexDoesNotAliasFieldClickKey`
  (mouse_test.go). Muster: **geteilte State-Felder brauchen disjunkte
  Namensräume** — wer ein bestehendes Feld für eine neue Ereignisklasse
  mitbenutzt, muss die Schlüsselräume beweisbar trennen (Offset/Prefix),
  nicht nur „unwahrscheinlich kollidierend" wählen.

### 2. PO-Wortlaut mit zwei Auslösern auf einen gefaltet (T4/bt-duz7, Fix-Runde R1)

- **Was lief nicht rund:** Der PO-Wortlaut nannte ZWEI Fall-c-Auslöser
  („Doppelklick (oder Zweitklick auf ein bereits selektiertes Feld)").
  Die Implementierung las beides als EIN zeitfenstergebundenes
  `isDouble`-Ereignis — der Zweitklick auf ein bereits selektiertes Feld
  nach >500ms öffnete nichts, obwohl die gespiegelte Enter-Kaskade
  zeitfensterlos ist.
- **Wie gefixt:** `wasSelected`-Check vor dem State-Update, feuert
  unabhängig vom Zeitfenster (`isDouble || wasSelected`, Commit `6e1152e`).
- **Forward-Guard:** Test
  `TestMouseDetailClickSecondClickOnSelectedFieldOpensOverlayOutsideWindow`
  (mouse_test.go). Checklisten-Punkt für Implementer: **Akzeptanzkriterien
  mit „identisch zu X"-Claims IMMER gegen X' tatsächliche Semantik prüfen**
  (hier: Enter-Kaskade ist zeitfensterlos, Doppelklick zeitfenstergebunden
  — „analog zur Enter-Kaskade" umfasste also beide Auslöser).

### 3. Lobby-Exit im Hauptfall unerreichbar (T5/bt-1u0t, Fix-Runde R1)

- **Was lief nicht rund:** Nach der B08-Hauptrunde war Quit-Stufe 2 im
  Hauptfall tot: `keyLobby` schickte q (und sogar ctrl+c) bei
  `client != nil` zurück zu Browse — es gab schlicht KEINEN Test, der
  `keyLobby`s esc/q/ctrl+c-Case exercierte, der Bug rutschte unbemerkt
  durch die grüne Suite (Reviewer live bestätigt).
- **Wie gefixt:** q/ctrl+c/esc in `keyLobby` entkoppelt (q → requestQuit,
  ctrl+c → sofort Quit, esc unverändert), Hint-Split `esc:back  q:quit`
  (Commit `d173927`).
- **Forward-Guard:** Tests `TestLobbyQOpensQuitConfirmWithLiveClient`,
  `TestLobbyCtrlCQuitsImmediatelyWithLiveClient` (view_lobby_test.go).
  Checklisten-Punkt: **neue Kaskaden IMMER end-to-end testen — JEDE Stufe,
  aus jedem realistischen Ausgangszustand**, nicht nur die neu gebaute
  Stufe (die Stufe-2-Tests der Hauptrunde liefen gegen den Kaltstart-Fall
  `client == nil` und verfehlten den Hauptfall).

### 4. NBSP-Wordwrap-Bug nur im Live-Smoke gefunden (T8/bt-d8kc)

- **Was lief nicht rund:** D06s Wechsel vom atomaren `:` zu einem echten
  Leerzeichen zwischen Taste und Aktionswort machte den Footer-Wordwrap
  angreifbar — bei 80 Spalten landete `S` allein am Zeilenende, `Sort` auf
  der Folgezeile. KEIN Unit-/Golden-Test fing das; erst der tmux-Smoke bei
  Grenzbreite zeigte es.
- **Wie gefixt:** `renderBindings()` glued Taste+Wort (und mehrwortige
  Descs) mit U+00A0 NBSP — `x/ansi`s Wordwrap bricht dort nie (Commit
  `6b2daa3`).
- **Forward-Guard:** Tests `TestRenderBindingsKeepsKeyAndDescTogetherAcrossWrap`,
  `TestRenderBindingsKeepsMultiWordDescTogether` (Wrap-Mechanik) +
  `TestDetailClickBacklogThreeLineFooterAt80Cols` (mouse_test.go,
  80-Spalten-Geometrie, Commit `3c084d2`) + CLAUDE.md-Zeile (Repo-Regeln):
  **Footer-/Wrap-Änderungen brauchen einen Smoke bei Grenzbreite (80
  Spalten)** — Unit-Tests bei „bequemer" Breite (100/120) sehen
  Umbruch-Bugs strukturell nicht.

### 5. Selbst-referenzieller Pin-Test (T4-Re-Review I01)

- **Was lief nicht rund:** `TestDetailClickKeyDisjointNumberSpaces`
  (mouse_test.go) prüft gegen die Konstante `detailClickKeyBase` statt
  gegen das harte Literal 1000 — gegen eine Mutation der Base selbst
  (z.B. versehentlich 10) ist der Test wertlos, er rechnet mit demselben
  Wert, den er prüfen soll.
- **Wie gefixt:** bewusst NICHT eigens gefixt (Verhaltenstest
  `TestMouseDetailClickTreeClickIndexDoesNotAliasFieldClickKey` deckt die
  eigentliche Invariante ab); Korrektur bei der nächsten Berührung von
  mouse_test.go (Literal 1000 einsetzen).
- **Forward-Guard:** Checklisten-Punkt für Test-Reviews: **Pin-Tests
  pinnen LITERALE, nie die Konstante, die sie pinnen sollen** — ein Test,
  der `want := someConst` schreibt, testet die Kopiertreue des Compilers,
  nicht den Wert.

### 6. Stille Vorfärbe-Konvention footer()/breadcrumb() (T8-Review I02)

- **Was lief nicht rund:** Seit D06 erwarten `footer()`/`breadcrumb()`
  (view.go) VORGEFÄRBTEN `renderBindings()`-Output und legen selbst keinen
  `theme.Muted`/`theme.Dim`-Wash mehr darüber (das äußere Wrap würde die
  inneren ANSI-Resets zerstören). Diese Konvention ist nur in
  Doc-Kommentaren dokumentiert — ein Aufrufer, der einen ROHEN String
  hineinreicht, bekommt kommentarlos ungestylten Text.
- **Wie gefixt:** kein Code-Fix nötig (Konvention ist konsistent
  umgesetzt); hier als dokumentierte Falle festgehalten.
- **Forward-Guard:** `chrome.golden` pinnt den Wegfall des äußeren Wraps
  (ANSI-Diff, Commit `6b2daa3`) + Doc-Kommentare an `footer()`/
  `breadcrumb()` (view.go). Checklisten-Punkt: **wer footer()/breadcrumb()
  neue Aufrufer gibt, reicht renderBindings()-Output (oder selbst
  gefärbte Strings) hinein — nie Roh-Text.**

### 7. Planner-ERRATUM B03 — Positiv-Beispiel (E8-Planning)

- **Was lief nicht rund:** GAR NICHTS — bewusst als Positiv-Beispiel
  festgehalten: der PO-Fund B03 („kinderlose Beans zeigen ein
  Expand-Dreieck") reproduzierte NICHT gegen den Ist-Code
  (`treeNodeMarker` prüfte bereits `!n.hasKids`). Der E8-Planner
  verifizierte den Fund VOR dem Task-Schnitt, deklarierte ihn als ERRATUM
  und schnitt statt eines Phantom-Fixes einen Regressionstest.
- **Wie gefixt:** kein Fix nötig — `TestTreeNodeMarkerBlankForLeaf`
  schreibt das bereits-korrekte Verhalten fest (Commit `75f1f4d`).
- **Forward-Guard:** Checklisten-Punkt für Planner: **jeden Fund vor dem
  Task-Schnitt gegen den Ist-Code verifizieren** — ein nicht
  reproduzierender Fund wird ERRATUM + Regressionstest, nie ein
  Fix-Task. (Genau so geschehen — beibehalten.)

## 2026-07-16 — E9-Abschluss (bt-tct9, Quellen: E9-Fix-Runden + Reviews)

### 1. Editor-Fehlerpfad verlor $EDITOR-Text (T2/bt-z4b1 F01, critical)

- **Was lief nicht rund:** Die Editor-Recovery-Closure prüfte nur
  `ErrConflict` — bei jedem anderen `UpdateWhole`-Fehler (live gefangen:
  CLI-VALIDATION_ERROR nach erfolgreichem Parse) wurde der komplette im
  $EDITOR verfasste Text kommentarlos verworfen. Datenverlust am
  häufigsten getippten Eingabeweg.
- **Wie gefixt:** JEDER `UpdateWhole`-Fehler schreibt jetzt via
  `writeConflictTempFile` recoverable weg; `applyMutationResult` matcht
  per `errors.As`. Der Analogfall „Ziel-Bean extern verschwunden" (F04)
  wurde als eigener Bug `bt-6bgn` identisch gesichert — alle vier
  Editor-Fehlerpfade recovery-gesichert.
- **Forward-Guard:** Tests je Fehlerpfad (editor/client). Checklisten-
  Punkt: **wo User-Eingaben an einem Fehlerpfad hängen, darf nie auf
  einen konkreten Fehlertyp gematcht werden — der Default-Arm muss die
  Daten retten.**

### 2. huh-KeyMap ersetzt Field-Konfiguration still (T3/bt-2v38 F01)

- **Was lief nicht rund:** `.ExternalEditor(false)` am huh-Text-Feld
  wirkte nur durch einen Aufrufreihenfolge-Zufall in `styleForm` —
  `huh.NewForm` + `WithKeyMap` ersetzt Field-Keymaps mit frischen
  Defaults, die Einstellung war de facto tot und der ctrl+e-Editor-Exit
  jederzeit reaktivierbar.
- **Wie gefixt:** expliziter `field.KeyBinds()`-Call nach `huh.NewForm`
  (form_edit_title.go) statt Verlass auf Reihenfolge-Nebenwirkung.
- **Forward-Guard:** Behavioral-Test, der BEIDE Bruch-Szenarien rot
  macht (Reihenfolge-Tausch und fehlender KeyBinds-Call). Merkregel:
  **huh-Field-Optionen nach jedem NewForm re-applizieren — WithKeyMap
  ist destruktiv.**

### 3. Wrap-Helfer ohne Hardwrap-Pass (T4/bt-b0w0 F01)

- **Was lief nicht rund:** `hangingIndentWrap` nutzte nur Wordwrap — ein
  103-Zeichen-Token ohne Leerzeichen erzeugte eine 110-Zellen-Zeile und
  riss das Layout; der CJK-Fall (Doppelbreiten) war der zweite Arm
  desselben Bugs.
- **Wie gefixt:** exakt das bestehende `wrapText`-Zwei-Pass-Muster
  übernommen: `ansi.Hardwrap(ansi.Wordwrap(text, w, ""), w, true)`.
- **Forward-Guard:** Tests mit spaceless-Token UND CJK-Input.
  Checklisten-Punkt: **jeder neue Wrap-Helfer folgt dem
  Zwei-Pass-Muster — Wordwrap allein bricht tokenlose Strings nicht.**

### 4. Commit-Meta-Verstöße erst am Gate gefangen (T6/bt-1e0t)

- **Was lief nicht rund:** Der Implementer committete Titel mit 67/53
  Zeichen und `Refs: bt-tct9` (Epic statt Task-bean) — Verstoß gegen
  die Commit-Regeln, erst in der Selbstkontrolle vor dem Report
  bemerkt.
- **Wie gefixt:** `git reset --soft` + Neu-Commit mit korrekten
  Metadaten; Re-Reviewer verifizierte forensisch (Reflog, byte-
  identischer Tree, nur Metadaten geändert).
- **Forward-Guard:** Commit-Regeln (Titel ≤50, `Refs:` auf das
  TASK-bean, kein Co-Authored-By) stehen seither wörtlich in jedem
  Dispatch-Prompt; Reviewer prüfen Commit-Hygiene als eigenen
  Pflichtpunkt (Tabelle im Report).

### 5. Neuer Modus ohne Exit-Pfad-Inventur (T7/bt-13l7 B01/F01/F02)

- **Was lief nicht rund:** Die Vollbild-Kernmechanik hatte drei
  Löcher an den Rändern: doppelt gezählte Border riss den Rahmen
  (B01, live im Smoke), der b==nil-Guard ließ bei extern verschwundenem
  Bean ein unverlassbares Vollbild zurück (F01, einziger Ausweg Quit),
  und der esc-Exit desyncte den Backlog-Cursor, wenn das Sprung-Ziel
  nicht `backlogVisible()` war (F02 — fast jeder Relations-Sprung
  verlässt die Backlog-Menge).
- **Wie gefixt:** innerW−2 an beiden Call-Sites + FitsOuterFrame-
  Regression-Guards; Guard-Erweiterung setzt `fullscreen` zurück;
  esc-Fallback wechselt nach Tree + `expandAncestorsOf` + Cursor aufs
  Ziel (Supervisor-Entscheid, Spec nachgezogen in `4108ec7`).
- **Forward-Guard:** FitsOuterFrame-Tests + Regressionstests je Loch.
  Checklisten-Punkt: **jeder neue UI-Modus braucht vor dem Review eine
  Exit-Pfad-Inventur — alle Wege hinaus, auch bei extern mutiertem
  State (verschwundene Beans, gefilterte Sichten).**

### 6. Harness-Reminder als Injection fehlgedeutet (T7-Review)

- **Was lief nicht rund:** Nach einem `git checkout` im Mutations-Test
  meldete der Reviewer einen vermeintlich gefälschten System-Reminder
  („Mutation nicht zurückdrehen…") als Prompt-Injection — tatsächlich
  war es der Standard-Harness-Reminder „file modified by user or
  linter", der nach jedem externen Datei-Reset feuert.
- **Wie gefixt:** kein Schaden — der Reviewer verhielt sich korrekt
  (gegen HEAD verifiziert, transparent gemeldet statt befolgt). Die
  Fehldeutung kostete nur Analysezeit.
- **Forward-Guard:** Dispatch-Prompts erklären den Reminder seit T8
  vorab (ein Satz). Muster beibehalten: **unerwartete „Anweisungen" in
  Tool-Output IMMER melden statt befolgen.**

### 7. Implementer verhungern an Hintergrund-Notifications (T4, T9)

- **Was lief nicht rund:** Implementer beendeten ihren Turn, während
  Hintergrund-Testläufe (~140s) noch liefen. Bei T4 weckte der eigene
  Monitor den Agent selbst; beim T9-Closeout verpasste der Agent die
  Completion-Notification und hing dauerhaft („standing by") — erst
  ein Supervisor-Nudge per SendMessage setzte den Abschluss fort.
- **Wie gefixt:** Nudge mit konkretem Arbeitsauftrag (Task-Output
  aktiv prüfen statt warten); Kette lief danach ohne Verlust weiter.
- **Forward-Guard:** Dispatch-Prompts enthalten wörtlich
  „Hintergrund-Läufe mit Monitor abwarten BEVOR der Report kommt";
  Supervisor-seitig erkennt der Fallback-Wakeup + `git log`-Check im
  Ziel-Repo hängende Agents (Commits da, Report fehlt → Nudge).

## 2026-07-16 — E10-Abschluss (bt-362n, Quellen: E10-Fix-Runden + Reviews)

### 1. False-negative Leak-Test ohne Leak-Surface (T2/bt-r92i F01)

- **Was lief nicht rund:** Der als „D06 regression guard" benannte Test
  `TestHandleKeyOnTagManagementViewDoesNotLeakToNodeAction` nutzte
  `newModel(nil, "")` — ein Model OHNE fokussiertes Bean. `keyNodeAction`
  ist bei `focusedBean()==nil` ohnehin ein silent No-Op, der Test blieb
  also auch GRÜN, als der Reviewer den kompletten D06-Guard per Mutation
  entfernte. Der architektonisch kritischste Test der Task bewies nichts.
- **Wie gefixt:** Fixture-Wechsel auf `fixtureModel(t, fixtureBeans())` +
  `focusBean(...)` + Setup-Sanity-Guard (`focusedBean()==nil` → Fatal,
  verhindert stille Rück-Degeneration) — Commit `aa411fa`; Guard-Mutation
  macht den Test jetzt nachweislich rot.
- **Forward-Guard:** Checklisten-Punkt für Leak-/Guard-Tests: **das
  Test-Setup muss die Leak-Surface real herstellen** (Zustand, in dem
  der geschützte Pfad ohne Guard tatsächlich feuern würde) plus
  Sanity-Assertion, die das Setup selbst pinnt. Reviewer prüfen
  Guard-Tests grundsätzlich per Guard-Entfernungs-Mutation.

### 2. Geteilter Save-Tail las Kontext aus Seiteneffekt (T4/bt-1lsu B01)

- **Was lief nicht rund:** `applyTagDefsSaved` fand den Cursor nach
  jedem `tagDefsSavedMsg` implizit über `m.tagMgmtInput.Value()` —
  sicher, solange Create der einzige Aufrufer war. T4s Delete wurde
  zweiter Aufrufer, der das Feld nie setzt; T3s esc-Abbruch leert
  `Value()` nicht → nach „Create getippt → esc → unabhängiges Delete"
  sprang der Cursor auf den stalen Text.
- **Wie gefixt:** `tagDefsSavedMsg` trägt ein explizites
  `refindName string`; jeder Aufrufer übergibt den Namen selbst
  (Create: neuer Name, Delete: Ziel, Rename: neuer Name) — Commit
  `315d21f`, Rollback-Mutation macht zwei Regressionstests rot.
- **Forward-Guard:** **Geteilte Ergebnis-Handler bekommen ihren Kontext
  als explizites Msg-Feld, nie über Model-Seiteneffekt-Reads** — jeder
  neue Aufrufer eines bestehenden tea.Cmd-Producers prüft, welche
  impliziten Model-Annahmen der Handler macht.

### 3. Bulk-Sweep gegen Resolver-Randfall (T5/bt-y9my, Datenverlust-Falle)

- **Was lief nicht rund (fast):** `SetTags` dokumentiert „same tag in
  add and remove → remove wins". T5s Rename-Sweep hätte beim Resubmit
  des UNVERÄNDERTEN Namens (den die D14-Dedupe-Exklusion bewusst
  erlaubt) je Bean `--tag X --remove-tag X` gesendet — das Tag wäre
  still von JEDEM betroffenen Bean gestrippt worden.
- **Wie gefixt:** Same-Name-Guard im Sweep-Cmd (`oldTag == newTag` →
  No-Op, kein einziger SetTags-Aufruf), vom Implementer proaktiv
  ergänzt und regression-getestet; Reviewer-Mutation bestätigte, dass
  der Test den Datenverlust gefangen hätte (Commit `6d3a9b4`).
- **Forward-Guard:** Checklisten-Punkt für jeden Bulk-Sweep über
  Mutations-APIs: **die dokumentierte Resolver-Semantik der API gegen
  Identitäts- und Randfälle (x→x, leer, nicht existent) prüfen, BEVOR
  der Sweep geschnitten wird** — und den Randfall als Test pinnen.

### 4. Upstream-Quirk: `beans create --tag` liefert stalen ETag (beans 0.4.2)

- **Was lief nicht rund:** Ein bei `beans create --tag` getaggtes Bean
  bekommt einen `list`/`show`-berichteten ETag, der mit dem intern von
  `update --if-match` geprüften NICHT übereinstimmt — jeder folgende
  `SetTags` läuft in CONFLICT, bis ein erfolgreiches `update` den ETag
  „repariert". Von T5-Implementer UND -Reviewer unabhängig live
  reproduziert (traf das Smoke-Setup; der continue-on-error-Sweep ritt
  korrekt durch alle CONFLICTs — ungeplanter D13-Beweis).
- **Wie gefixt:** Workaround im Smoke-Setup: Tags via separatem
  `beans update --tag` setzen, nie via `create --tag`. Kein
  beans-tui-Fix möglich (Upstream-Verhalten).
- **Forward-Guard:** Smoke-/Test-Setups taggen NIE über `create --tag`.
  Upstream-Issue-Kandidat für hmans/beans — **POST nur mit
  PO-Freigabe** (extern), gesammelt beim bestehenden
  T03-Upstream-ETag-Punkt.

### 5. API-529-Abbrüche per Transcript-Resume überbrückt (T5-Review, T7)

- **Was lief nicht rund:** Zwei Agents (T5-Reviewer mitten im Review,
  T7-Implementer direkt nach Dispatch) starben am serverseitigen
  „529 Overloaded"-API-Fehler.
- **Wie gefixt:** SendMessage-Resume aus dem erhaltenen Transcript —
  beide setzten ohne Arbeitsverlust fort (der Reviewer explizit mit
  „bereits Erledigtes nicht wiederholen, nur Restpunkte"). Vorher je
  ein `git log`/`git status`-Check im Ziel-Repo, um den echten Stand
  in den Resume-Prompt zu schreiben.
- **Forward-Guard:** Supervisor-Playbook-Regel: bei `failed`-Status
  IMMER erst Repo-Stand erheben (`git log --oneline`, `git status`),
  dann Resume mit präzisem Reststand-Auftrag; Dispatch-Prompts weisen
  Agents an, nach einem Resume selbst `git log` zu prüfen statt Arbeit
  zu doppeln.

## 2026-07-17 — E11-Nacharbeitsrunde (Quellen: 6 Reviews, 3 Fix-Runden, 1 CHANGES_REQUIRED)

### 1. tmux-Session-Namenskollision zwischen parallelen Agents

- **Was lief nicht rund:** Zwei parallel smokende Agents nutzten generische tmux-Session-Namen — Keystrokes des einen landeten in der Session des anderen (bt-9ipw-Smoke lief kurz gegen das falsche Repo).
- **Wie gefixt:** Re-Run mit eindeutigem Namen (`bt9ipw$$`); Vorgabe ab da in jedem Dispatch-Prompt.
- **Forward-Guard:** Prompt-Vorlagen-Pflichtzeile: tmux-Smoke IMMER mit task-eindeutigem Session-Namen (`<task>$$`).

### 2. Fixtures erbten Zustand statt ihn zu setzen (PF-1-Abhängigkeit)

- **Was lief nicht rund:** 4-5 Maus-Klick-Tests liefen nur grün, weil PF-1 META offen erzwang — `detailFocusModel`-Fixture setzte `accOpen`/`detailFocus` nie explizit. Die PF-18-Umsetzung deckte die stille Kopplung auf.
- **Wie gefixt:** Fixtures setzen jetzt den echten FocusIn-Zustand explizit (bt-98cb, Commit 66b91f4).
- **Forward-Guard:** Test-Konvention: Fixtures setzen JEDEN gelesenen Model-Zustand explizit, nie auf Render-Seiteneffekte anderer Invarianten verlassen. Mutations-Stichprobe im Review prüft das.

### 3. Neue Node-Art im Tree: Konsumenten-Sweep war unvollständig

- **Was lief nicht rund:** Platzhalter-Node (bt-39cl) war in Navigation/Maus/focusedBean abgesichert, aber `applyLoaded`s Cursor-Restore (Watcher-Reload-Pfad) konnte den Cursor auf die id-lose Platzhalter-Zeile setzen — einziges CHANGES_REQUIRED der Runde, vom Reviewer per eigenem Repro-Test gefunden, vom Implementer-Smoke strukturell unsichtbar.
- **Wie gefixt:** Guard `n.id != ""` im id-Match + skipPlaceholder im oldPos-Fallback + Regressionstests (Commit 3a79433).
- **Forward-Guard:** Checklisten-Punkt für neue Node-/Zeilen-Arten: ALLE `visibleNodes()`-/`nodes[pos]`-Konsumenten sweepen, inkl. asynchroner Pfade (applyLoaded/Watcher/Resize), nicht nur Input-Handler. Reviewer-Prompt fragt das explizit ab.

### 4. Glyph-Rescan statt struktureller Index (latent)

- **Was lief nicht rund:** `activeRelationLine` findet die Cursor-Zeile per `strings.Contains("▶")`-Rescan des gerenderten Strings — Titel mit ▶/▷ würden das Fenster falsch zentrieren (latent, kein Titel betroffen; Reviewer-Testfall belegt).
- **Wie gefixt:** Noch nicht — als bt-se4q verankert (numerischer Index aus Body-Konstruktion).
- **Forward-Guard:** Konvention: Positions-Informationen strukturell durchreichen (Rückgabewert/Feld), nie aus gerendertem Text zurückparsen. bt-se4q setzt das um.

### 5. beans-CLI-Flag-Falle: --body-append + -d kombiniert

- **Was lief nicht rund:** `beans update <id> --body-append -d -` → "accepts 1 arg(s), received 2" (zwei Body-Flags, "-" wird Positional).
- **Wie gefixt:** `--body-append -` liest selbst stdin.
- **Forward-Guard:** Dispatch-Prompt-Standardzeile (bereits in allen Vorlagen dieser Runde): "`--body-append -` liest stdin. NICHT `--body-file -`, NICHT mit `-d -` kombinieren."
