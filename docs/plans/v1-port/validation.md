# beans-tui (`bt`) — v1 Validation

**Stand:** commit `e716a3a21666a9296e3f5c0ac79e571a88e3c03e` (`e716a3a`), 2026-07-15
**Live-Testfunktions-Count:** 400 (`grep -rhoP "^func Test\w+" internal/ cmd/ | sort -u | wc -l`)

## Methodik

Zwei unabhängige Validierungsläufe gegen `design-spec.md` §10 (14 User
Stories) plus ein Evidence-Review:

- **E6 Task 1** (bean `bt-wm4w`, US-01…07) — Performance-Messung (Test +
  tmux-Wanduhr), ein durchgehender Tastatur-Only-Cross-View-Smoke, echter
  Zwei-Prozess-ETag-Konflikt.
- **E6 Task 2** (bean `bt-9yvh`, US-08…14) — Bestätigungsläufe der
  zitierten Testfunktionen plus ein kombinierter Zwei-Repo-Smoke
  (Lobby/Watcher-Switch/Yank/`state.json`).
- **Evidence-Review** (2026-07-15, Status `EVIDENCE_SOLID`) — drei
  Nachträge auf `bt-7k7q` (US-04-Testnamen namentlich, US-12-Beleg-Kette
  verstärkt, B01/bt-gdkx als eigener D-Punkt), in diesem Dokument
  eingearbeitet.

Die vollen Rohbelege (Messwerte, Capture-Auszüge, exakte Toast-/Fehler-Texte)
liegen in `docs/_free-notes/e6-t1-evidence.md` und
`docs/_free-notes/e6-t2-evidence.md` (git-ignored, Scratch-Input für dieses
Dokument — nicht Teil des Repos-Commits). Dieses Dokument ist das
konsolidierte, committete v1-Abnahme-Artefakt.

**Quellen:** `design-spec.md` §10 (US-Kriterien), §5 (Review-Flow), §15 (E7
PO-Feedback) · `epic-E6-plan.md` »Task 3« · bean `bt-7k7q` (Auftrag +
Evidence-Review-Nachträge) · `bt-wm4w`/`bt-9yvh` (Rohbeleg-Träger) ·
`bt-gdkx` (B01) · `bt-heg9` (E7-Entscheidungstabelle D01/I01/I02) ·
`bt-aq5s`/`bt-gzcu` (E2/E3-Erbe) · `bt-6oyy` (Tag-Management-Page, Q03).

---

## 1. User-Story-Matrix (14/14 validiert — 13× PASS, 1× PARTIAL)

| US | Titel | Ergebnis | Evidenz-Anker | Anmerkung |
|---|---|---|---|---|
| US-01 | `bt` starten, sofort Projektbaum | PASS | `TestListPerformanceAt100Beans` (internal/data/performance_test.go, 32.57/36.16/35.46 ms über 3 Läufe) + tmux-Wanduhr gegen `/tmp/bt-scratch-100` (100 beans, 411/411/419 ms über 3 Läufe) + `TestDiscoverFindsConfigUpward`/`TestDiscoverErrorsWhenNoConfigFound` | Zwei unabhängige Belege weit unter dem 2000ms-Kriterium; Fehlermeldung real zitiert: `Error: no .beans.yml found above /tmp` (Exit 1) |
| US-02 | Navigation ohne Maus | PASS | tmux-Smoke (Session `btsmoke`, 200×55, E6 T1, gegen dieses Repo) — 15 Segmente, alle PASS (volle Tabelle: Anhang §4.1) | Review-Cockpit-Segment aus dem Plan-Vorbild ersatzlos gestrichen (PF-14/E7 — existiert nicht mehr), Settings (nur via Palette) ersetzt es im Durchlauf |
| US-03 | Bean-Details vollständig | PASS | Glamour-Stichprobe `docs/_free-notes/e6-markdown-smoke.md` (Codeblock/Liste/Link) via `beans update --body-file` geschrieben, Detail-Accordion `[2] BODY` visuell korrekt gerendert | Farbcodierung (Glyphen) + Relations-Navigation im US-02-Durchlauf mitbestätigt (bereits vollständig testabgedeckt, kein Neu-Test nötig) |
| US-04 | Command-Palette | PASS | `TestPaletteActionsBeanContextFirst`, `TestDispatchPaletteBeanJumpsCursorAndSwitchesToBrowse` (internal/tui/overlay_palette_test.go, beide PASS, re-verifiziert 2026-07-15) + Live: `paletteItem{actionID:"repo_picker"}` (overlay_palette.go:81) im US-02-Durchlauf bestätigt | Archiv-Toggle bewusst KEINE Palette-Aktion (lebt im Filter-Facetten-Overlay, box_filter_facets.go:79, E5 T7 design decision e) — kein Bug, korrekt fehlende Erweiterung |
| US-05 | Suche/Filter | PASS | Live-Bleve-Stichprobe im US-02-Durchlauf: lokaler Substring "Foundation" (Treffer `bt-blsy`), Feldsyntax `title:Foundation` (1 Treffer) | `status:in-progress` liefert 0 Treffer — beans-CLI selbst unterstützt nur `slug:`/`title:`/`body:` im Volltextindex (kein TUI-Bug, Client reicht korrekt durch); Status-Filterung läuft bewusst über den separaten Facetten-Filter. PO hat die Filter-Logik zusätzlich als "exzellent" abgenommen (bt-heg9, PO-Nachtrag 3) |
| US-06 | Bean anlegen | PASS | Confirm-Gate + Cursor-Verifikation im Scratch-Repo `/tmp/bt-scratch-100` (`bt-scratch-100-wer2`, CLI-Frontmatter-Verifikation) | Parent korrekt vom Cursor-Bean vorbelegt (`TestBuildCreateBeanFormPrefillsParentFromCursor`-Verhalten live bestätigt), Formular zeigte ausschließlich gültige Enum-Werte |
| US-07 | Feld-Edit constrained | PASS | Feld-Edit-Kaskade (PF-5) live an `bt-scratch-100-wer2` (status/priority/title/tags) + echter Zwei-Prozess-ETag-Konflikt (`/tmp/bt-scratch-etag`, ETag `d03f384b36f3bb30`→`bb36f22d412c30a3`, volle Tabelle: Anhang §4.2) | Erster ECHTER (nicht simulierter) Zwei-Prozess-Konflikt-Beleg des Projekts; Toast, Sticky-Verhalten (`TestConflictToastIsStickyAndSurvivesReload` live reproduziert) und Recovery-Tempfile-Inhalt exakt geprüft |
| US-08 | Review-Stand sehen (redefiniert PF-14) | **PARTIAL** | `TestPassReviewSetsCompletedAndRemovesTag`, `TestRejectReviewSwapsTagAndAppendsSection` (PASS) + Live-Check aller 4 Spec-Flächen (`/tmp/bt-scratch-a`) | Filter (`f`) + Live-Reload funktionieren vollständig; Tree UND Detail-META zeigen Tags aktuell GAR NICHT → **B01/bt-gdkx** (medium, s. Bugs-Tabelle), PO-Entscheid **D01**. Suche (`tag:`-Feldsyntax) 0 Treffer — gleiche Scope-Ursache wie US-05, kein neuer Bug. **Fix in E8 (D01), bt-gdkx** — Umsetzung bean `bt-e6q9` (E8 Task 1), Re-Validierung dieser Zeile nach E8-Abnahme fällig |
| US-09 | Backlog priorisiert | PASS | `TestBacklogShowsParentlessReadyBeansFromIndex`, `TestBacklogSortCyclesThroughFourModesAndBackToStart`, `TestBacklogGolden`, `TestBacklogGoldenDeterministic` — alle PASS | Kein Sort-Modus-Indikator sichtbar (unverändert offen seit E2, **D02**) — kein funktionaler Mangel, reine Diskoverability-Frage |
| US-10 | Live-Reload | PASS | `TestWatcherFiresOnceForBurst`, `TestReloadKeepsCursorOnID`, `TestSwitchRepoCmdStopsOldWatcherStartsNew` (alle PASS) + Live-Beleg beide Richtungen im Zwei-Repo-Smoke (Anhang §4.3) | Eigenständiger Beleg dieses Tasks, kein Fremdverweis auf `bt-7dfj` mehr |
| US-11 | Yank/Clipboard | PASS | `TestYankShowsConfirmationToast`, `TestYankOnOrphanRootNoop` (PASS) + Live-OSC52/`pbpaste`-Beleg im Zwei-Repo-Smoke: Toast `● Copied: bt-scratch-b-q9do`, `pbpaste` liefert den exakten Markdown-Kontext | `TestReviewCockpitYankUsesReviewStandNotSingleBean` aus dem Plan-Vorbild existiert nicht mehr (Review-Cockpit-Yank-Override mit PF-14 entfernt) — kein FAIL, ersatzlos weg |
| US-12 | devd-Look | PASS | VQA-Übernahme (Supervisor, bean `bt-zk9p`, 16 Screenshots `docs/_free-notes/vqa-2026-07-15/`, **vor E7/PF-14** entstanden) + 3 Post-E7-Delta-Captures (E6 T2: Browse/Lobby/Help) + 3 frische Captures dieses Tasks (E6 T3, `docs/_free-notes/e6-i02-filter-overlay.txt`/`e6-i02-palette.txt`/`e6-i02-settings-form.txt`, `tmux capture-pane -e`, 2026-07-15) | VQA-Screenshots sind ausdrücklich "vor E7" gekennzeichnet (Dateinamen zeigen noch das inzwischen entfernte Review-Cockpit). Beide Delta-Check-Runden bestätigen PF-3/4/6/10/11 korrekt im Live-Build, keine Regression, keine Cockpit-Reste. Zwei kosmetische VQA-Findings aus der Supervisor-QA (`bt-zk9p`): VQA-I01 (Footer-Umbruch @110 Spalten) ist durch PF-11 entschärft, sein Kernproblem lebt am Header weiter (**D04**); VQA-I02 (Lobby-Pfad-Kürzung rechtsseitig) bleibt unverändert offen, aber ohne eigenen D-Punkt in diesem Dokument (Scope hier: die 8 D-Punkte aus dem Evidence-Review) |
| US-13 | Hilfe/Footer-Hints | PASS | Help-Overlay live (3 Gruppen Navigation/Views & Global/Actions, aus `keymap.go` generiert, kein `R`-Binding mehr) + `view_lobby.go:264-286` Footer-Hint live abgelesen (`i/k:↑↓  enter:open  type:filter  esc/q:back`/`quit`) | Matrix-Entwurf-Lücke ("Lobby-Footer existiert noch nicht") ist bereits geschlossen, kein Fix nötig |
| US-14 | Repo-Wechsel | PASS | Kombinierter Zwei-Repo-Smoke (`/tmp/bt-scratch-a` + `/tmp/bt-scratch-b`, eigenes `HOME`, volle Tabelle: Anhang §4.3) — Lobby-Metrik, Watcher-Switch beide Richtungen, `state.json` `{"last_repo": "/tmp/bt-scratch-a"}` | Erledigt zugleich US-10s Live-Bestätigung; `TestDecideStartupPrioritiesInOrder` deckt die Trigger-Matrix testseitig vollständig ab |

---

## 2. Bugs

| B | Schwere | Beschreibung | Repro | Bean | Status |
|---|---|---|---|---|---|
| B01 | medium | US-08: Tags (Tag-Trio `to-review`/`accepted`/`rejected` UND generische Tags) sind weder im Tree noch im Detail-META sichtbar — von den 4 im Spec genannten Flächen (Tree/Detail/Filter/Suche) funktioniert nur die Filter-Facette. Toter Code `tagsInline`/`tagSwatch` (internal/tui/render_shared.go:91-110) ohne Aufrufer deutet auf ersatzloses Wegfallen beim PF-1/PF-4-Meta-Redesign hin, nicht bewusstes Weglassen | `/tmp/bt-scratch-a`: Bean mit `--tag to-review` anlegen, `bt` starten, Tree inspizieren (kein Indikator), Detail öffnen (META hat kein Tags-Feld), `f` → Tags-Facette togglen (funktioniert) | `bt-gdkx` | 🟣 Offen — kein Fix in dieser Validierung (Validierung sammelt Evidenz, ändert keinen Code), PO-Entscheid s. **D01** |

---

## 3. Hygiene-Log

Zwei laut Matrix-Entwurf veraltete Bean-Markierungen wurden bei grünem
Testbeweis auf den aktuellen Stand korrigiert (reine Doku-Korrektur an
bereits erledigter Arbeit, kein Code-Fix):

| Bean | Punkt | Vorher | Nachher | Testverweis |
|---|---|---|---|---|
| `bt-aq5s` | B01 — "X (Filter-Reset) wirkt nicht in Backlog-View" | 🟣 Offen | 🟢 Gelöst | `TestKeyBacklogFilterClearResetsFacets` PASS (E6 T1, verifiziert 2026-07-15) |
| `bt-gzcu` | I01 — "Konflikt-Statuszeile nach ErrConflict nur 1 Frame sichtbar" | 🟡 Unklar | 🟢 Gelöst | `TestConflictToastIsStickyAndSurvivesReload` PASS (E6 T2, verifiziert 2026-07-15) |

---

## 4. Smoke-Belege-Anhang

### 4.1 US-02 — Cross-View-Durchlauf (E6 T1, tmux Session `btsmoke`, 200×55, gegen dieses Repo, rein Tastatur)

| Segment | PASS/FAIL | Beleg |
|---|---|---|
| Start → Tree | PASS | Breadcrumb `> beans-tui-repository: Browse`, Header zeigt alle 7 globalen Bindings (PF-11) |
| `tab` → Detail-Fokus | PASS | Cursor-Marker springt von Tree-Zeile auf `[1] META`-Zeile im Detail-Pane; `▶` erscheint auf `title:`-Feld |
| Ziffern-Sprung `2`/`3`/`4`/`1` | PASS | BODY (Glamour-Body), RELATIONS (Children-Liste), HISTORY (Created/Updated/ETag), zurück zu META — alle 4 Sektionen korrekt adressiert |
| Enter-Kaskade (PF-5) — Sektion→Feld | PASS | `1`→`enter` (Meta-Sektion) → Feldcursor `▶` auf `title:`; `down` → `▶` wandert auf `status:` |
| Enter-Kaskade — Feld→Edit-Overlay | PASS | `enter` auf `status:`-Feld öffnet "Set value"-Overlay, Cursor korrekt auf "todo (current)" vorpositioniert |
| `esc` schließt Overlay ohne Mutation | PASS | Nach Abbruch: `status: todo` unverändert (dogfooding-Repo unangetastet) |
| `shift+tab` → Fokus zurück zu Tree | PASS | Cursor-Marker zurück auf Tree-Zeile, Detail-Pane ohne Marker |
| `/` Suche (lokal) | PASS | s. US-05 |
| `b` → Backlog | PASS | Breadcrumb wechselt zu `> beans-tui-repository: Backlog`, zeigt parentloses ready-Bean `bt-6oyy` |
| `ctrl+k` → Command-Center | PASS | Palette öffnet mit "verb entity"-Labels (PF-8), kein Doppelpunkt, Englisch |
| Fuzzy-Filter "stat" | PASS | Filtert auf `set status` + `set parent` (Subsequence-Fuzzy-Match) |
| `?` → Help-Overlay | PASS | 3 Gruppen (Navigation/Views & Global/Actions), **kein** `R`-Binding mehr (PF-14 bestätigt) |
| Settings (nur via Palette) | PASS | `ctrl+k` → "settings" → Formular mit `repos`/`editor`/`accent`/`tree_width`; `esc` bricht ohne Speichern ab |
| `p` → Lobby | PASS | ASCII-Logo, "(no repos in config.yaml -- ctrl+k -> settings)" (leere Config, korrektes Leerzustand-Verhalten) |
| `esc` aus Lobby → zurück zu Browse | PASS | Kehrt zur Tree-View zurück |
| `q` → Quit-Confirm → `enter` → Exit | PASS | Modal "Quit? Really quit bt.", sauberer Prozess-Exit |

### 4.2 US-07 — Echter Zwei-Prozess-ETag-Konflikt (E6 T1, `/tmp/bt-scratch-etag`)

| Schritt | PASS/FAIL | Beleg |
|---|---|---|
| Pane A: `bin/bt`, Cursor auf Task, `ctrl+e` (ETag beim Öffnen eingefroren) | PASS | Fake-`$EDITOR` (4s-Sleep) suspendiert `bt` deterministisch (`tea.ExecProcess`-Kontrakt) |
| Pane B: `beans update <id> -s in-progress` während A suspendiert | PASS | ETag-Bump extern verifiziert: `d03f384b36f3bb30` → `bb36f22d412c30a3` |
| Pane A resumed, speichert mit eingefrorenem ETag | PASS | Echter `ErrConflict`, kein simulierter |
| Toast (sticky, oben rechts) | PASS | `Conflict: bean changed externally` / `Version saved: /var/folders/…/beans-tui-conflict-2219101204.md` |
| Reload nach Konflikt | PASS | Tree/Detail zeigen sofort `status: in-progress` (Pane B's externe Änderung) |
| Recovery-Tempfile | PASS | Datei existiert, Inhalt exakt der von Pane A angehängte Text — nichts verloren |
| Sticky-Verhalten über `ctrl+r` | PASS | Toast bleibt nach manuellem Reload sichtbar (`TestConflictToastIsStickyAndSurvivesReload` live reproduziert) |

### 4.3 US-14 — Repo-Wechsel-Smoke (E6 T2, `/tmp/bt-scratch-a` + `/tmp/bt-scratch-b`, eigenes `HOME`)

| Feature | Ergebnis | Beleg |
|---|---|---|
| Lobby-Start (2 Repos konfiguriert, Nicht-Repo-cwd) | PASS | Lobby direkt (kein Repo auto-resolved), Tabelle zeigt `/tmp/bt-scratch-a 4/4` · `/tmp/bt-scratch-b 1/1` |
| Repo-Wechsel `enter` → B | PASS | Breadcrumb → `bt-scratch-b: Browse`, Watcher-Switch (`TestSwitchRepoCmdStopsOldWatcherStartsNew`) |
| Externe Änderung im INAKTIVEN Repo (A) | PASS | Kein Reload in `bt` (Tree blieb bei 1 Eintrag) |
| Externe Änderung im AKTIVEN Repo (B) | PASS | Reload binnen ~0.01s (Poll-Schleife), neuer Task sofort im Tree sichtbar |
| `y` auf neuem Bean | PASS | Toast `Copied: bt-scratch-b-q9do`, `pbpaste` liefert exakten Markdown-Kontext |
| Lobby erneut (`p`) nach Mutationen | PASS | Metrik aktualisiert: `/tmp/bt-scratch-a 5/5` · `/tmp/bt-scratch-b 2/2` |
| Tag-Trio-Filter in Repo A (`f`→Tags→`to-review`) | PASS | Tree filtert auf "Agent Work To Review" (US-08-Beleg) |
| Externe Tag-Änderung (`--tag to-review`) während Filter aktiv | PASS | Filter zeigt sofort 2 Treffer statt 1, ohne `ctrl+r` |
| Quit-Confirm | PASS | Modal "Quit? Really quit bt.", `enter: quit`/`esc: cancel`, sauberer Exit |
| `state.json` nach Exit | PASS | `{"last_repo": "/tmp/bt-scratch-a"}` — korrekt zuletzt aktives Repo |

### 4.4 US-12/I02-Nachtrag — Frische Post-E7-Captures (E6 T3, 2026-07-15)

Schließt die im Evidence-Review als dünn markierte US-12-Beleg-Kette
(3 zusätzliche `tmux capture-pane -e`-Captures gegen `/tmp/bt-scratch-a`,
Textform, git-ignored):

| Capture | Datei | Beleg |
|---|---|---|
| Filter-Overlay (`f`) | `docs/_free-notes/e6-i02-filter-overlay.txt` | Overlay-lokaler Footer `space/x:toggle  X:clear  enter/esc/f:done`, Status/Type/Priority-Facetten mit `[ ]`-Checkboxen — Englisch (PF-7) |
| Command-Palette (`ctrl+k`) | `docs/_free-notes/e6-i02-palette.txt` | "verb entity"-Labels (`set status`, `set tags`, `go to backlog`, `filter facets`, `reload data`, …), kein Doppelpunkt (PF-8) |
| Settings-Form (Palette → "go to settings") | `docs/_free-notes/e6-i02-settings-form.txt` | Formular mit `repos`/`editor`/`accent`-Feldern, huh-Rendering intakt nach E7-Umbauten |

---

## 5. Entscheidungstabelle (D-Codes, ZULETZT)

Alle acht D-Punkte wurden in der PO-Feedback-Runde 2 (Grilling 2026-07-15,
bean `bt-ntoz`) entschieden — Umsetzung: `epic-E8-plan.md` (design-spec.md
§15 PF-15/PF-16 ist die aktualisierte Quelle der Wahrheit). Mehrere
Entscheidungen weichen bewusst von der ursprünglich hier notierten
„Empfehlung" ab (PO-Grilling hat andere Prioritäten gesetzt, z. B. D03/D04/
D06) — die „Entscheidung"-Spalte gibt den TATSÄCHLICHEN PO-Beschluss
wieder, die „Empfehlung"-Spalte bleibt als historischer Kontext stehen.

| Dxx | Hintergrund | Entscheidung | Empfehlung | Status |
|---|---|---|---|---|
| D01 | US-08-Validierung (bean `bt-gdkx`): Tags sind weder im Tree noch im Detail-META sichtbar — nur die Filter-Facette funktioniert. Toter Code (`tagsInline`/`tagSwatch`) deutet auf ersatzloses Wegfallen beim PF-1/PF-4-Meta-Redesign hin. Optionen: (a) eigene Tag-Zeile in der Meta-Feldliste, (b) Tag-Zeile im Kopfblock, (c) Tree-Row-Glyph-Suffix, (d) Kombination | Meta-Feldliste bekommt eine 7. Zeile `tags:` (nach `priority`, vor `created_at`) — KEIN Tree-Suffix (PO: rasches Filtern via `f`-Tags-Facette deckt den Überblick bereits ab). Umsetzung: E8 / bt-ntoz (Task 1, bean `bt-e6q9`) | Meta-Zeile + Tree-Suffix kombiniert (schließt beide fehlenden Flächen, ohne die schlanke 6-Feld-Meta-Liste selbst umzubauen) | 🟢 Entschieden |
| D02 | Backlog-Sortierung (`S`) cycled durch 4 Modi (status/priority/created/updated), aber es gibt keine sichtbare Anzeige des aktiven Modus (E2-Erbe, `bt-aq5s` I02, in E6 T2/US-09 erneut bestätigt — Ist-Zustand unverändert) | Dezenter Subtext-Suffix in der Backlog-Suchzeile (z. B. `⌕ / search · sort prio`), Tree bleibt unverändert. Umsetzung: E8 / bt-ntoz (Task 8, bean `bt-d8kc`) | Für Post-v1/Polish vormerken — kein funktionaler Mangel (Sortierung selbst funktioniert), reine Diskoverability-Frage | 🟢 Entschieden |
| D03 | `esc` in Detail-Fokus ist No-Op — nur `j`/`←` bzw. `tab`/`shift+tab` verlassen den Fokus (E2-Erbe, `bt-aq5s` I01 — geprüft: Bean-Body zeigt weiterhin 🟡, PF-13/tab-shift+tab-Symmetrie hat das nicht miterledigt) | `esc` wird universelles „eine Ebene zurück" — auch in der Detail-Kaskade (Feld→Sektion→Fokus verlassen). Pfeiltasten verlieren im GEGENZUG ihre Fokus-Exit-Nebenwirkung (B01, revidiert PF-13). Umsetzung: E8 / bt-ntoz (Task 3, bean `bt-qbyq`) | Belassen — `esc` ist app-weit für "Overlay/Form ohne Speichern schließen" reserviert; ein dritter Exit-Weg würde diese Konvention durchbrechen, `tab`/`shift+tab` decken den Rückweg seit PF-13 bereits symmetrisch ab | 🟢 Entschieden |
| D04 | 80-Spalten-Terminal (gängiges Maß) truncatet den 7-Globals-Header, `q:quit` kann abgeschnitten werden (E7 T7-Review I01, bean `bt-heg9`) | Header auf exakt 4 Globals gekürzt (`ctrl+k`/`p`/`?`/`q`) statt Umbruch — `ctrl+r`/`esc`/`enter` bleiben im Help-Overlay dokumentiert. Umsetzung: E8 / bt-ntoz (Task 8, bean `bt-d8kc`) | Header analog zum Footer umbrechen (Konsistenz, kein Informationsverlust) statt Prioritäts-Truncation-Reihenfolge | 🟢 Entschieden |
| D05 | Overlay-lokale Footer-Hints (z.B. Palette/Help) wiederholen `enter`/`esc`, obwohl der Header sie bereits global zeigt (E7 T7-Review I02, bean `bt-heg9`) | Sign-off, kein Code-Fix — durch D04s Header-Kürzung ohnehin keine Dopplung mehr. Umsetzung: E8 / bt-ntoz (kein separater Task, D04-Nebeneffekt) | Sign-off — Redundanz in einem Overlay-/Modal-Kontext ist visuell verstärkend, kein funktionaler Schaden | 🟢 Entschieden |
| D06 | Footer in Browse/Backlog lässt `f`/`X`/`b`/`t`/`a`/`B`/`y` absichtlich weg — alle bleiben über das Help-Overlay (`?`) erreichbar/dokumentiert (E7 T7-Review D01, bean `bt-heg9`) | Footer-Neuspezifikation (ersetzt den T7-Stand): `f`/`b`/`t`/`a`/`y`/Blocking kommen zurück in den Footer (Reihenfolge PO-verbatim, Taste teal/Aktionswort subtext-grau statt `:`), Navigations-Keys raus, Blocking-Taste `B`→`r` umbelegt. Umsetzung: E8 / bt-ntoz (Task 8, bean `bt-d8kc`, Q06-Präzisierung) | Belassen (schlank) — Footer wäre sonst auf 80 Spalten noch enger (siehe D04); Help-Overlay deckt Vollständigkeit bereits ab | 🟢 Entschieden |
| D07 | Upstream beans 0.4.2: frisch per `beans create` angelegte Beans melden via `list`/`show` ein ETag, das vom `update`-internen `--if-match`-Check abweicht, bis das erste erfolgreiche Update die Datei neu geschrieben hat (E3-Erbe, `bt-gzcu` I02). Nicht bt-fixbar (liegt im Upstream-Paket `hmans/beans`) | Upstream-Issue bei `hmans/beans` NACH v1-Abnahme einreichen (Entwurf durch Agent, POST NUR mit PO-Freigabe) — explizit NICHT Teil von E8. Bis dahin bleibt die Warm-up-Konvention die dokumentierte Workaround | Upstream-Issue bei `hmans/beans` NACH v1-Abnahme einreichen; bis dahin bleibt die Warm-up-Konvention (erst ein rohes CLI-Update, dann TUI-Smoke) die dokumentierte Workaround | 🟢 Entschieden |
| D08 | PO-Wunsch (E7-Nachtrag, bean `bt-6oyy`): Tags zentral über eine eigene Page definierbar machen (Persistenzort/View/Tag-Picker-Integration offen). Scope-Frage: noch Teil von v1 oder v1.1-Backlog? | v1.1 — B14 (E8 / bt-ntoz) liefert die v1-Minimal-Lösung (Tag-Neuanlage sichtbar über Palette + Tag-Picker-Footer-Hint), die Page selbst bleibt v1.1-Backlog. Bereits als Body-Nachtrag auf `bt-6oyy` dokumentiert | v1.1 — neues Feature ohne Port-Paritäts-Bezug; v1-Ziel ist der devd-Port + PO-Feedback-Polish, keine neuen Flächen | 🟢 Entschieden |

---

## 6. Fazit

v1-Ziel — vollständiger, validierter Port der DevDash-TUI auf beans plus
E7-PO-Feedback-Polish — ist **ehrlich zu 13/14 User Stories voll erreicht**.
Die verbliebene Lücke (US-08, Tag-Sichtbarkeit im Tree/Detail) ist real,
aber eng begrenzt: der funktional wichtigste Pfad (Filter, Live-Reload) läuft
einwandfrei, der PO hat lediglich keinen beiläufigen Sichtprüfungs-Weg beim
normalen Tree-Browsen. Keiner der 14 Nachweise beruht auf Behauptung — jede
Zeile trägt einen konkreten Test-, Golden- oder Smoke-Anker. Alle acht
D-Punkte (D01–D08, §5) sind inzwischen entschieden (PO-Grilling-Runde 2,
bean `bt-ntoz`) und Umsetzung läuft über `epic-E8-plan.md`; D01
(Tags-Sichtbarkeit) ist die einzige mit direktem User-Story-Bezug (US-08 →
PARTIAL) — nach E8-Abnahme ist eine Re-Validierung dieser Zeile fällig.
