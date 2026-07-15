# Epos E6 — Validierung & Release (voll granular)

Liefert: `docs/plans/v1-port/validation.md` mit dokumentiertem Nachweis je
US-01…14 (design-spec §10), geschlossene Validierungslücken aus dem
Matrix-Entwurf (Smoke statt Behauptung), README-Finalisierung, `bt` real
startbar demonstriert im lean-stack-Repo, Milestone `bt-apmy` + Epic `bt-zk9p`
mit Tag `to-review` (PO-Gate — Agent schließt NICHT), eine PO-Entscheidungs-
vorlage (D-Codes-Tabelle, 5 Punkte) für die verbliebenen offenen Rufe aus
E2/E3/VQA, und eine Hygiene-Korrektur an zwei veralteten Bean-Markierungen,
die laut Recherche bereits faktisch gelöst sind. KC-Update (D07) und
lean-stack-Verweis-bean-Pflege sind NICHT Teil dieses Eposs (Supervisor-Sache
nach PO-Freigabe, s. Task 4 Schritt 5).

Quellen: `design-spec.md` §10 (US-Tabelle/AC), §12 (E6-Scope) ·
`implementation-plan.md` »Epos E6« · `docs/_free-notes/e6-validation-matrix-draft.md`
(Hauptarbeitsgrundlage, Stand a13c59a + T8-Update-Absatz) · bean `bt-zk9p`
Body (VQA-Findings VQA-I01/I02, Supervisor 2026-07-15) · bean `bt-7dfj` Body
(E6-Handover-Hinweise + Smoke-Matrix-Format-Vorbild) · bean `bt-aq5s`/`bt-gzcu`
Bodies (PO-Hinweise-Tabellen).

---

## Design-Entscheidungen (vor Task 1 festgezurrt)

### a) Validierungsmatrix-Übernahme-Strategie

Der Entwurf sagt selbst wörtlich: für US-01–US-09 und US-11–US-13 ist die
"Evidenz vorhanden"-Spalte **inhaltlich unverändert gültig** seit a13c59a;
nur US-10 und US-14 haben sich materiell geändert (jetzt vollständig statt
teilweise/in Arbeit, T6/T6b/T7/T8 haben geliefert). **Entscheidung:** T1/T2
verifizieren die im Entwurf zitierten Testnamen/Goldens per gezieltem
`command go test -run` STATT sie blind zu übernehmen (der Entwurf selbst
benutzt an mehreren Stellen "vermutlich" — das ist keine Validierung,
sondern ein Verdacht) und schließen NUR die explizit als "Validierungslücke
für E6" benannten Punkte per neuem Smoke. Kein Neu-Testen von Punkten, die
der Entwurf bereits als "vollständig" ohne Lücke einstuft (US-08 z.B. —
"Keine wesentliche") — dafür reicht ein Bestätigungslauf der genannten
Testfunktionen.

### b) Scratch-Repo-Generator für den US-01-Performance-Smoke

`newTestRepo(t)` (`internal/data/testrepo_test.go`) erzeugt fix 3
Fixture-Beans — für "100 beans" (design-spec §10 AC wörtlich) braucht es
zwei UNABHÄNGIGE Beleg-Ebenen, keine neue Debug-Instrumentierung im
Produktcode (kein Scope-Creep über Validierung hinaus):

1. **Automatisiert (Testsuite):** neuer Helfer `newTestRepoN(t *testing.T,
   n int) string` in `testrepo_test.go`, gleiches YAML-Frontmatter-Muster
   wie `newTestRepo`, generiert 1 Milestone + 10 Epics (Kinder des
   Milestones) + `n-11` Tasks gleichmäßig auf die Epics verteilt (`parent`
   gesetzt) — realistische Baumtiefe statt einer flachen Liste. Neuer Test
   `internal/data/performance_test.go::TestListPerformanceAt100Beans` misst
   `Client.List()` (= exakt der Aufruf `beans list --json --full`, den
   `cmd/tui.go::runTUI`/`loadCmd` beim Start ausführt) direkt — eine
   ehrliche Teilmessung des echten Startpfads, kein erfundener Proxy: das
   bubbletea-Rendering selbst liegt sub-Millisekunde, keine separate
   Messung nötig.
2. **Real-Terminal-Wanduhr-Beleg:** ein über das echte `beans`-CLI erzeugtes
   100-Bean-Scratch-Repo (kein Test-Tmpdir, das räumt sich automatisch weg),
   `bin/bt` darauf in tmux gestartet, Wanduhr-Differenz zwischen Send-Keys
   und erstem sichtbaren Tree-Inhalt gemessen (`date +%s%3N` vor/nach einer
   `tmux capture-pane`-Polling-Schleife). Ergänzt (1), ersetzt es nicht —
   "messbar machen, nicht behaupten" heißt zwei unabhängige Belege, nicht
   eine hübschere Behauptung.

### c) PO-Entscheidungsvorlage-Format (validation.md)

D-Codes-Tabelle (CLAUDE.md-Pflichtformat: `Dxx | Hintergrund | Entscheidung |
Empfehlung | Status`), **5 Zeilen**, deckungsgleich mit dem E6-Auftrag
("esc-in-Detail-Fokus, Sort-Indikator, Upstream-ETag, VQA-I01/I02"):

| Code | Herkunft | Kern |
|---|---|---|
| D01 | bt-aq5s I01 | `esc` in Detail-Fokus No-Op (nur `j`/`tab` verlassen) — belassen oder `esc` als 3. Exit? |
| D02 | bt-aq5s I02 | Kein Sort-Modus-Indikator im Backlog — v1-Scope behalten oder Post-v1? |
| D03 | bt-gzcu I02 | Upstream-ETag-Drift bei frischen Creates (beans 0.4.2) — Upstream-Issue bei hmans/beans nach v1-Abnahme? |
| D04 | bt-zk9p VQA-I01 | Footer-Keymap bricht bei 110 Spalten um ("e:Edit in $EDITOR") — kürzen oder akzeptieren+dokumentieren? |
| D05 | bt-zk9p VQA-I02 | Lobby kürzt lange Repo-Pfade rechtsseitig — Ellipsis links oder akzeptieren? |

Je Zeile eine **Empfehlung** (CLAUDE.md: "Empfehlung stets markieren"),
Entscheidung bleibt leer für den PO, Status durchgängig 🟡 Unklar bis PO
entscheidet. Diese 5 werden NICHT von diesem Epos implementiert (Rahmen-
bedingung, s. Auftrag).

### d) Hygiene an zwei veralteten Bean-Markierungen (KEINE PO-Entscheidungsvorlage)

Der Matrix-Entwurf stuft zwei WEITERE Fundstellen selbst als "vermutlich
bereits gelöst … Bean-Status wirkt veraltet, keine offene PO-Frage mehr"
ein — das sind **keine** echten offenen Design-Entscheidungen, sondern
Karteileichen aus E2/E3, die bei Gelegenheit (E6) bereinigt werden sollten,
sonst verschmutzen sie jede künftige Review-Queue-Ansicht dauerhaft:

- **bt-aq5s B01** ("X Filter-Reset wirkt nicht in Backlog-View", 🟣 Offen):
  `TestKeyBacklogFilterClearResetsFacets` existiert bereits und deckt genau
  das ab.
- **bt-gzcu I01** ("Konflikt-Statuszeile nur 1 Frame sichtbar", 🟡 Unklar):
  `TestConflictToastIsStickyAndSurvivesReload` (E5 T1) hat das strukturell
  gelöst (Toast sticky übersteht Reload).

**Entscheidung:** T1/T2 verifizieren den jeweils genannten Test GEZIELT
(`go test -run <Name> -v`), und NUR bei grünem Beweis wird die
Status-Markierung in der jeweiligen Bean-Body-Tabelle aktualisiert (🟣/🟡 →
🟢 + Testverweis) — reine Doku-Korrektur an bereits erledigter Arbeit, kein
Code-Fix, keine neue PO-Frage.

### e) `validation.md`-Struktur

1. Kopf (Kontext, Quellen, Stand/Commit-Hash, Testfunktions-Count LIVE
   gezählt statt der veralteten "384" aus dem Entwurf).
2. US-Tabelle (14 Zeilen: `ID | Titel | Status PASS/GAP | Evidenz-Anker |
   Kommentar`).
3. D-Codes-Tabelle (die 5 aus design decision c).
4. Hygiene-Log (die 2 aktualisierten Alt-Findings aus design decision d,
   Vorher/Nachher + Testverweis).
5. Smoke-Belege-Anhang (volle tmux-Durchlauf-Tabellen aus T1/T2, Format wie
   bt-7dfj Smoke-Matrix — PASS/FAIL je Zeile, kein Fließtext-Ersatz).

---

## Task-Übersicht

| Task | bean | Inhalt | US-Range | blocked_by |
|---|---|---|---|---|
| T1 | `bt-wm4w` | US-Validierungslauf Teil 1 | US-01…07 | — |
| T2 | `bt-9yvh` | US-Validierungslauf Teil 2 | US-08…14 | — |
| T3 | `bt-7k7q` | validation.md + README-Finalisierung | — | `bt-wm4w`, `bt-9yvh` |
| T4 | `bt-upyz` | Release-Hygiene (Demo, Milestone-Ritual, Handover) | — | `bt-7k7q` |

T1/T2 sind gegenseitig unabhängig (verschiedene US-Ranges, verschiedene
Codebereiche) — Reihenfolge zwischen ihnen ist beliebig, T3 braucht aber
BEIDE Rohbelege.

---

## Task 1: US-Validierungslauf Teil 1 (US-01…07) (`bt-wm4w`)

**Files:**
- Create: `internal/data/performance_test.go`
- Modify: `internal/data/testrepo_test.go` (+ `newTestRepoN`), ggf.
  `.beans/bt-aq5s--*.md` (Body-Hygiene, nur bei grünem Testbeweis)
- Scratch (git-ignored, kein Bundle-Schreiben): `docs/_free-notes/`,
  `/tmp/bt-scratch-100/`

- [ ] **Step 1: Baseline.** `command go test ./...` (2x, ohne `-short`)
  grün, `command go test ./... -race` grün, `command go build -o bin/bt .`
  ok, `command gofmt -l .` leer, `command go vet ./...` leer. Aktuellen
  Testfunktions-Count festhalten: `grep -rhoP "^func Test\w+" internal/
  cmd/ | sort -u | wc -l` (für validation.md — NICHT die veraltete "384"
  aus dem Entwurf übernehmen).
- [ ] **Step 2: US-01a — automatisierte Performance-Messung.**
  `internal/data/testrepo_test.go`: `newTestRepoN(t *testing.T, n int)
  string` (1 Milestone + 10 Epics + `n-11` Tasks, gleiches
  Frontmatter-Schema wie `newTestRepo`, prefix `tt-`, generierte IDs
  `tt-0001`…). `internal/data/performance_test.go`:
  `TestListPerformanceAt100Beans` — `dir := newTestRepoN(t, 100)`,
  `requireBeansBinary(t)`, `start := time.Now()`, `beans, err :=
  (&Client{RepoDir: dir}).List()`, `elapsed := time.Since(start)`,
  `t.Logf("List() @ 100 beans: %v", elapsed)`, `if elapsed > 2*time.Second
  { t.Fatalf("List() took %v, want <2s", elapsed) }`, `if len(beans) != 100
  { t.Fatalf(...) }`. Kommentar im Test: verweist auf design decision b
  (misst den echten Startpfad-Subprocess-Call, kein erfundener Proxy).
  `command go test ./internal/data/ -run TestListPerformanceAt100Beans -v`
  → PASS, geloggte Dauer notieren.
- [ ] **Step 3: US-01b — Real-Terminal-Wanduhr-Beleg.** Scratch-Repo über
  das echte CLI erzeugen (kein Test-Tmpdir):
  ```sh
  rm -rf /tmp/bt-scratch-100 && mkdir -p /tmp/bt-scratch-100
  cd /tmp/bt-scratch-100 && beans init
  m=$(beans create "Perf Milestone" -t milestone --json | jq -r .id)
  for i in $(seq 1 10); do
    eid=$(beans create "Perf Epic $i" -t epic --parent "$m" --json | jq -r .id)
    eval "epic_$i=\$eid"
  done
  for i in $(seq 1 89); do
    ep=$(( (i % 10) + 1 )); eval "p=\$epic_$ep"
    beans create "Perf Task $i" -t task --parent "$p" >/dev/null
  done
  beans list --json | jq length   # → 100 erwartet, sonst Loop-Fehler prüfen
  ```
  Zeitmessung:
  ```sh
  tmux new-session -d -s btperf -x 220 -y 50
  START=$(date +%s%3N)
  tmux send-keys -t btperf \
    "/Users/erik/Obsidian/tools/beans-tui/beans-tui-repository/bin/bt /tmp/bt-scratch-100" Enter
  while ! tmux capture-pane -t btperf -p | grep -q "Perf Milestone"; do sleep 0.05; done
  END=$(date +%s%3N)
  echo "wall_ms=$((END-START))"
  tmux send-keys -t btperf "q" Enter && sleep 0.3 && tmux send-keys -t btperf "y" Enter
  tmux kill-session -t btperf
  ```
  Akzeptanz: `wall_ms` < 2000. Wert für validation.md notieren.
- [ ] **Step 4: US-01c — Discovery + Fehlermeldung.** `.beans.yml`-Discovery
  aufwärts (bereits testabgedeckt: `TestDiscoverFindsConfigUpward`) — nur
  der Fehlermeldungs-Wortlaut ist bislang nicht zitiert:
  `cd /tmp && /Users/erik/Obsidian/tools/beans-tui/beans-tui-repository/bin/bt`
  → exakten Ausgabetext für validation.md kopieren (kein Fix erwartet,
  reiner Beleg).
- [ ] **Step 5: US-02 — großer Tastatur-Only-Cross-View-Durchlauf.** EIN
  tmux-Smoke (gegen dieses Repo, `bin/bt` ohne Argument): Tree → Detail
  (Ziffern-Sprung `1`–`4`) → Backlog (`b`) → Command-Center (`ctrl+k`) →
  Review-Cockpit (`R`) → Help (`?`) → Settings (`ctrl+k` → "settings:
  öffnen") → Lobby (`p`), rein per Tastatur, kein Maus-Einsatz. Je Segment
  PASS/FAIL + Beleg (capture-pane-Auszug) in einer Tabelle protokollieren
  (Format wie bt-7dfj Smoke-Matrix). Dieser EINE Durchlauf liefert
  zugleich Teilbelege für US-03/04/05/13 (Schritte 6–8 unten hängen sich
  hier ein, nicht separat wiederholen — Bündelung wie im Matrix-Entwurf
  empfohlen).
- [ ] **Step 6: US-03 — Glamour-Stichprobe mit komplexem Markdown.**
  Kleine Markdown-Fixture (Codeblock, Liste, Link) in
  `docs/_free-notes/e6-markdown-smoke.md` (Scratch, git-ignored) anlegen,
  per `beans update <id> --body-file docs/_free-notes/e6-markdown-smoke.md`
  auf einen Bean im laufenden Smoke (Step 5) schreiben, im Detail-Accordion
  (Section 2, Body) visuell prüfen — Codeblock/Liste/Link korrekt
  gerendert? Ergebnis in derselben Tabelle vermerken.
- [ ] **Step 7: US-04 — Palette-Regressionscheck.** Im selben Durchlauf
  `ctrl+k` öffnen, prüfen ob T6/T7 neue Einträge beigetragen haben
  (Repo-Picker-Aktion, Archiv-Toggle-Aktion). Falls nicht als
  Palette-Aktion vorhanden: `grep -n "action\|palette" internal/tui/
  overlay_palette*.go` prüfen, ob das architektonisch überhaupt vorgesehen
  ist — Ergebnis EXAKT so dokumentieren wie vorgefunden (kein
  Fake-PASS, falls die Aktion schlicht nicht existiert — dann ist das
  kein E6-Bug, sondern korrekt fehlende Erweiterung, siehe design-spec §9
  Scope).
- [ ] **Step 8: US-05 — Bleve-Query-Stichprobe.** `grep -n "func.*Search\|
  bleve" internal/data/*.go` für den echten Query-Pfad, danach im
  selben Durchlauf `/` mit einem Titel-Teilstring (≥3 Zeichen, lokal) UND
  einer Bleve-Feldsyntax (z.B. `status:in-progress`, falls die
  Client-Weiterleitung Feldsyntax durchreicht) real gegen dieses Repo
  ausprobieren. Ergebnis (Trefferzahl, Verhalten) notieren.
- [ ] **Step 9: US-06 — Cursor-auf-neuem-Bean.** Im selben Durchlauf `c` →
  Formular ausfüllen → Confirm-Gate bestätigen → per capture-pane
  verifizieren, dass der Cursor auf dem neu erzeugten Bean steht (nicht an
  alter Position).
- [ ] **Step 10: US-07 — echter Zwei-Prozess-ETag-Konflikt.** Zwei
  tmux-Panes gegen DASSELBE Scratch-Repo (`/tmp/bt-scratch-100` oder ein
  frisches Mini-Repo): Pane A `bin/bt`, Cursor auf Bean X, `ctrl+e`
  (Editor-Suspend, ETag bei Öffnen eingefroren). Pane B (rohes CLI,
  während A offen ist): `beans update <X-ID> -s in-progress` (bumpt
  ETag). Pane A: Body im `$EDITOR` ändern + speichern+schließen → echter
  `ErrConflict` erwartet. Belegen: Toast-Text ("Konflikt: Bean extern
  geändert"), Sticky-Verhalten (übersteht `ctrl+r`), Recovery-Tempfile-Pfad
  (`ls <Pfad-aus-Toast>` zeigt die Datei).
- [ ] **Step 11: Hygiene bt-aq5s B01.** `command go test ./internal/tui/
  -run TestKeyBacklogFilterClearResetsFacets -v` → PASS. NUR dann:
  `beans show bt-aq5s --json` (Body-Ausschnitt der B01-Zeile exakt
  prüfen), dann `beans update bt-aq5s --body-replace-old "<exakte
  B01-Zeile>" --body-replace-new "<gleiche Zeile mit 🟢 + Testverweis>"`.
- [ ] **Step 12:** `command go test ./... -short` grün, `gofmt -l .`/`go
  vet ./...` leer.
- [ ] **Step 13:** Commit `test(data): US-01..07 Validierungsbelege (E6
  Task 1)` — Body enthält alle Messwerte/Tabellen aus Step 2–10 (für T3
  direkt übernehmbar).

**Akzeptanz-Checkliste:**
- [ ] `TestListPerformanceAt100Beans` grün, geloggte Dauer < 2s
- [ ] tmux-Wanduhr-Messung < 2000ms dokumentiert
- [ ] US-02-Durchlauf vollständig protokolliert (PASS/FAIL je Segment)
- [ ] US-03 Markdown-Stichprobe visuell bestätigt
- [ ] US-04 Palette-Regressionscheck dokumentiert (PASS oder faktisch "keine
  Palette-Aktion vorgesehen")
- [ ] US-05 Bleve-Feldsyntax real ausprobiert, Ergebnis notiert
- [ ] US-06 Cursor-auf-neuem-Bean visuell bestätigt
- [ ] US-07 echter Zwei-Prozess-Konflikt erzeugt + belegt
- [ ] bt-aq5s B01 NUR bei grünem Testbeweis aktualisiert

---

## Task 2: US-Validierungslauf Teil 2 (US-08…14) (`bt-9yvh`)

**Files:**
- Modify: ggf. `.beans/bt-gzcu--*.md` (Body-Hygiene, nur bei grünem
  Testbeweis)
- Scratch (git-ignored): `/tmp/bt-scratch-a/`, `/tmp/bt-scratch-b/`,
  eigenes `HOME`-Overlay für `~/.config/beans-tui/`

- [ ] **Step 1: US-08 — Bestätigungslauf, keine neue Smoke-Arbeit** (Entwurf:
  "Keine wesentliche" Lücke). `command go test ./internal/tui/ -run
  "TestReviewQueueGroupsByEpicCanonicalOrder|TestKeyReviewCockpitPassFiresPassReview|TestKeyReviewCockpitRejectOpensCommentForm|TestPassReviewSetsCompletedAndRemovesTag|TestRejectReviewSwapsTagAndAppendsSection|TestReviewSummaryLinePositionInToReview"
  -v` → alle PASS.
- [ ] **Step 2: US-09 — Bestätigungslauf.** `command go test
  ./internal/tui/ -run
  "TestBacklogShowsParentlessReadyBeansFromIndex|TestBacklogSortCyclesThroughFourModesAndBackToStart|TestBacklogGolden"
  -v` → alle PASS. PO-Entscheid I02 (Sort-Indikator) NICHT selbst
  entscheiden — geht unverändert in D02 (T3).
- [ ] **Step 3: US-10 — Bestätigungslauf.** `command go test
  ./internal/tui/ ./internal/data/ -run
  "TestWatcherFiresOnceForBurst|TestReloadKeepsCursorOnID|TestSwitchRepoCmdStopsOldWatcherStartsNew"
  -v` → alle PASS. Der "Repo-Wechsel löst im alten Repo keinen Reload mehr
  aus"-Fall wird im kombinierten Smoke (Step 7) LIVE reproduziert (T8 hat
  das schon einmal belegt — E6 erzeugt einen eigenständigen, unabhängigen
  Beleg statt eines Fremdverweises auf bt-7dfj).
- [ ] **Step 4: US-11 — Bestätigungslauf + einmaliger eigenständiger
  OSC52-Beleg.** `command go test ./internal/tui/ -run
  "TestYankShowsConfirmationToast|TestYankOnOrphanRootNoop|TestReviewCockpitYankUsesReviewStandNotSingleBean"
  -v` → alle PASS. Im kombinierten Smoke (Step 7) einmal `y` auf einem
  Bean UND einmal `y` im Review-Cockpit auslösen, `pbpaste` danach prüfen
  (OSC52-Capture, wie bt-7dfj T8 es bereits vorgemacht hat — E6
  reproduziert kurz und eigenständig statt den alten Beleg zu zitieren).
- [ ] **Step 5: US-12 — VQA-Übernahme, KEINE neue Screenshot-Session.**
  Supervisor-VQA (bean `bt-zk9p` Body, 16 Screenshots
  `docs/_free-notes/vqa-2026-07-15/`) ist bereits "bestanden" mit
  explizitem devd-Look-Befund. In validation.md (T3) wird dieser Befund
  1:1 mit Verzeichnis-Verweis übernommen — keine Doppelarbeit.
- [ ] **Step 6: US-13 — Lobby-Footer-Lücke prüfen.** Matrix-Lücke: "Footer-
  Hints für die künftige Lobby-View existieren noch nicht". `grep -n
  "footer\|Footer" internal/tui/view_lobby.go` — prüfen ob ein
  Footer/Hint-Text existiert. Live im kombinierten Smoke (Step 7) ablesen
  (unterste Zeile der Lobby-View). Falls kein Footer vorhanden: als GAP in
  validation.md vermerken (KEIN Fix in diesem Task — nur bei trivialem,
  eindeutig im Scope liegendem Ein-Zeilen-Fix ohne Rückfrage ergänzen,
  sonst PO-Frage statt eigenmächtiger Erweiterung).
- [ ] **Step 7: US-14 + kombinierter Repo-Wechsel-Smoke.** Zwei
  Scratch-Repos:
  ```sh
  rm -rf /tmp/bt-scratch-a /tmp/bt-scratch-b
  mkdir -p /tmp/bt-scratch-a /tmp/bt-scratch-b
  (cd /tmp/bt-scratch-a && beans init && beans create "Repo A Task" -t task)
  (cd /tmp/bt-scratch-b && beans init && beans create "Repo B Task" -t task)
  mkdir -p /tmp/bt-scratch-home/.config/beans-tui
  cat > /tmp/bt-scratch-home/.config/beans-tui/config.yaml <<'EOF'
  repos:
    - /tmp/bt-scratch-a
    - /tmp/bt-scratch-b
  EOF
  ```
  tmux-Durchlauf mit `HOME=/tmp/bt-scratch-home`: `p` → Lobby zeigt beide
  Repos mit "offen/gesamt"-Metrik → `enter` auf Repo B → Watcher-Switch →
  von AUSSEN `beans create "External A" -t task --config
  /tmp/bt-scratch-a/.beans.yml` (löst KEINEN Reload in `bt` aus, da B
  aktiv ist) → von AUSSEN `beans create "External B" -t task --config
  /tmp/bt-scratch-b/.beans.yml` (löst Reload < 2s aus, neuer Task
  sichtbar) → Session beenden → `cat
  /tmp/bt-scratch-home/.config/beans-tui/state.json` zeigt Repo B als
  zuletzt aktiv. Dieser EINE Durchlauf erledigt zugleich Step 3's
  Live-Bestätigung — kein zweiter Smoke nötig.
- [ ] **Step 8: Hygiene bt-gzcu I01.** `command go test ./internal/tui/
  -run TestConflictToastIsStickyAndSurvivesReload -v` → PASS. NUR dann:
  `beans show bt-gzcu --json` (I01-Zeile exakt prüfen), dann `beans
  update bt-gzcu --body-replace-old "<exakte I01-Zeile>"
  --body-replace-new "<gleiche Zeile mit 🟢 + Testverweis>"`.
- [ ] **Step 9:** `command go test ./... -short` grün, `gofmt -l .`/`go vet
  ./...` leer.
- [ ] **Step 10:** Commit `test(tui): US-08..14 Validierungsbelege +
  Repo-Wechsel-Smoke (E6 Task 2)` — Body enthält alle Tabellen/Belege aus
  Step 1–7 (für T3 direkt übernehmbar).

**Akzeptanz-Checkliste:**
- [ ] US-08/09/10/11 Bestätigungsläufe alle PASS
- [ ] US-12 VQA-Übernahme in Task-Body verlinkt (kein neuer Screenshot-Lauf)
- [ ] US-13 Lobby-Footer-Befund dokumentiert (vorhanden/GAP)
- [ ] US-14-Smoke vollständig (Lobby-Metrik, Watcher-Switch beide
  Richtungen, `state.json`-Persistenz)
- [ ] bt-gzcu I01 NUR bei grünem Testbeweis aktualisiert

---

## Task 3: `validation.md` + README-Finalisierung (`bt-7k7q`)

**Files:**
- Create: `docs/plans/v1-port/validation.md`
- Modify: `README.md` (Status-Abschnitt, Known-Issues-Abschnitt, Verweis
  auf validation.md)

**blocked_by:** T1, T2 (braucht deren Rohbelege 1:1)

- [ ] **Step 1:** `validation.md` anlegen — Kopf (Kontext, Quellen,
  Commit-Hash `git rev-parse --short HEAD`, LIVE Testfunktions-Count aus
  T1 Step 1).
- [ ] **Step 2:** US-Tabelle, 14 Zeilen (`ID | Titel | Status PASS/GAP |
  Evidenz-Anker | Kommentar`) — Werte AUS T1/T2-Commit-Bodies übernehmen,
  NICHT aus dem Matrix-Entwurf blind kopieren (design decision a). Jede
  Zeile braucht mindestens einen konkreten Anker (Testname, Golden-Name,
  oder tmux-Smoke-Verweis mit Datum/Commit).
- [ ] **Step 3:** D-Codes-Tabelle einfügen (5 Zeilen, design decision c,
  exakter Wortlaut wie oben).
- [ ] **Step 4:** Hygiene-Log-Abschnitt (bt-aq5s B01 + bt-gzcu I01,
  Vorher/Nachher + Testverweis, NUR falls in T1/T2 tatsächlich
  aktualisiert).
- [ ] **Step 5:** Smoke-Belege-Anhang — volle tmux-Tabellen aus T1 Step 5
  (US-02-Cross-View) + T1 Step 10 (US-07-Konflikt) + T2 Step 7
  (US-14-Repo-Wechsel), Format wie bt-7dfj Smoke-Matrix (Feature |
  Ergebnis | Beleg).
- [ ] **Step 6:** `README.md`: "Status"-Abschnitt aktualisieren ("E6
  (Validierung & Release) ist fertig" statt "ist offen"), Verweis auf
  `docs/plans/v1-port/validation.md` ergänzen. "Known Issues"-Abschnitt um
  VQA-I01/VQA-I02 ergänzen als "PO-Entscheid ausstehend, siehe
  validation.md D04/D05" (NICHT als gelöst markieren — Entscheidung liegt
  noch offen).
- [ ] **Step 7:** Commit `docs(plan): validation.md + README E6-Stand`.

**Akzeptanz:**
- [ ] `validation.md` existiert mit allen 14 US-Zeilen, kein US ohne
  Evidenz-Anker
- [ ] D-Codes-Tabelle mit exakt 5 Zeilen, Empfehlung je Zeile ausgefüllt
- [ ] README verweist auf validation.md, Known Issues aktualisiert

---

## Task 4: Release-Hygiene (`bt-upyz`)

**Files:** keine Code-Änderungen erwartet; Modify `docs/SSTD.md` (nur bei
Bedarf), beans-Metadaten (`bt-apmy`, `bt-zk9p`, Task-beans dieses Eposs)

**blocked_by:** T3

- [ ] **Step 1: Demo im lean-stack-Repo.**
  ```sh
  cd /Users/erik/Obsidian/tools/lean-stack
  /Users/erik/Obsidian/tools/beans-tui/beans-tui-repository/bin/bt
  ```
  `.beans.yml` existiert dort bereits (verifiziert) → Direktstart ohne
  Lobby (US-01/US-14-Beleg im Fremd-Repo-Kontext, nicht nur im eigenen
  Dogfooding-Repo). tmux-Beleg: Tree zeigt die lean-stack-Beans (Stand
  dieser Recherche: 43). capture-pane-Auszug in den Commit-Body dieses
  Tasks.
- [ ] **Step 2: CHANGELOG — Entscheidung dokumentieren, nicht stillschweigend
  auslassen.** `ls CHANGELOG.md` — kein bestehendes CHANGELOG in diesem
  Repo, kein Nutzer-Wunsch bekannt → YAGNI, ausgelassen (im Task-Body kurz
  begründen statt kommentarlos zu überspringen).
- [ ] **Step 3: Milestone-/Epic-Ritual.** `beans show bt-zk9p --json |
  jq .tags` prüfen (aktuell kein `to-review`-Tag laut Recherche) →
  `beans update bt-zk9p --tag to-review`. `beans update bt-apmy --tag
  to-review` (Milestone, PO-Gate — Agent setzt NIE `completed`).
- [ ] **Step 4: Task-beans dieses Eposs (T1–T4) auf `completed`**
  (agent-abschließbar) — NICHT Epic `bt-zk9p`, NICHT Milestone `bt-apmy`.
- [ ] **Step 5: Supervisor-Handover-Notiz (NICHT selbst ausführen)** — im
  Bean-Body/Commit-Body dieses Tasks explizit vermerken:
  > Offen für den Supervisor nach PO-Freigabe von E6: (a) KC-Konzept
  > `po-immersion-beans-via-obsidian-bases-no-custom-tui` via `/okf` um
  > einen Supersede-Hinweis ergänzen (D07, design-spec.md §14); (b)
  > lean-stack-Verweis-bean `lean-stack-5jqa` schließen (dessen eigene
  > Akzeptanz: "v1 abgenommen (alle US validiert) → diesen bean
  > schließen"). Beide Punkte sind bewusst NICHT Teil von E6 Task 4 —
  > Skill-gebundene bzw. bewusst dem Supervisor vorbehaltene Arbeit.
- [ ] **Step 6:** `docs/SSTD.md` — Pointer-Update NUR falls sich
  Worktree-Weiche/Referenzen geändert haben (voraussichtlich unverändert:
  main-direkt bleibt, Referenzquellen bleiben).
- [ ] **Step 7:** Commit `docs(release): E6-Abschluss — Demo, Milestone
  to-review, Handover-Hinweise`.

**Akzeptanz:**
- [ ] `bt` nachweislich im lean-stack-Repo startbar (tmux-Beleg im
  Commit-Body)
- [ ] `bt-zk9p` UND `bt-apmy` tragen Tag `to-review`, keins von beiden
  `completed`
- [ ] Alle E6-Task-beans (T1–T4) `completed`
- [ ] Handover-Notiz vermerkt (D07 + lean-stack-5jqa), NICHT ausgeführt

---

## Selbst-Review (Plan gegen Auftrag + design-spec + Matrix-Entwurf)

- **Rahmenbedingungen eingehalten:** kein `git push`, keine Tags — nur
  lokale Commits (kein Tag-Schritt in T4) ✓ · offene PO-Entscheide (esc,
  Sort-Indikator, Upstream-ETag, VQA-I01/I02) landen als D01–D05 in
  `validation.md`, NICHT implementiert ✓ · D07/KC-Update UND
  lean-stack-Verweis-bean explizit NICHT als E6-Task geplant, nur als
  Handover-Notiz in T4 Step 5 vermerkt ✓ · US-01-Performance ist
  messbar (zwei unabhängige Belege, design decision b), nicht behauptet ✓.
- **Matrix-Entwurf-Coverage:** alle 8 "Empfohlene E6-Validierungsschritte"
  des Entwurfs sind einem Task/Step zugeordnet — 1 (Blocker) bereits
  erledigt vor E6-Start (bt-zk9p ist in-progress) · 2 (US-01-Performance)
  → T1 Step 2/3 · 3 (Cross-View) → T1 Step 5 · 4 (ETag live) → T1 Step 10 ·
  5 (OSC52 real) → T2 Step 4 · 6 (Repo-Wechsel) → T2 Step 7 · 7 (devd-
  Vergleich) → T2 Step 5 (VQA-Übernahme statt Doppelarbeit) · 8
  (Sort-Indikator-PO-Entscheid) → D02.
- **Bündelungs-Empfehlung des Entwurfs übernommen:** Schritte 2/3/5/6 des
  Entwurfs sind in T1 Step 5 (US-02-Cross-View) bzw. T2 Step 7
  (US-14-Repo-Wechsel) zu je EINEM großen tmux-Durchlauf gebündelt, nicht
  einzeln wiederholt — deckt sich mit "Bündelung"-Absatz im Entwurf.
- **Task-Schnitt-Alternative geprüft:** der Auftrags-Vorschlag (T1
  US-01..07 / T2 US-08..14 / T3 validation.md+README / T4
  Release-Hygiene) wurde 1:1 übernommen — Alternative (ein einziger
  Mega-Validierungs-Task) wurde verworfen, weil T1/T2 unabhängig
  parallelisierbar bleiben und die Task-Granularität den anderen Epen
  (T-Anzahl 5-8) entspricht.
- **Kein Placeholder:** jeder Task-Step hat einen konkreten Befehl, Pfad
  oder Testnamen — keine Stelle mit "prüfen ob nötig" ohne definiertes
  Prüfkriterium (Ausnahme bewusst: T4 Step 2 CHANGELOG-Entscheidung ist
  als Ja/Nein-Check mit klarer Default-Antwort formuliert, keine echte
  Lücke).
