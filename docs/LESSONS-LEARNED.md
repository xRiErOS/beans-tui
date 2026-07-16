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
