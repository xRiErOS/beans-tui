---
# bt-enrd
title: 'T6 E2-Abschluss: Tests+Build, beans-Pflege, NSP-Handover'
status: completed
type: task
priority: high
created_at: 2026-07-14T21:57:40Z
updated_at: 2026-07-15T00:11:57Z
parent: bt-aq5s
blocked_by:
    - bt-2jve
    - bt-gzu6
---

Ziel: E2-Abschluss-Ritual (implementation-plan.md »Epos-Rituale«): Tests+Build grün,
beans-Pflege, Commit, ce-nsp-auto-Handover für E3.

Plan: docs/plans/v1-port/epic-E2-plan.md »Task 6«.

## Akzeptanz
- [x] command go test ./... grün, command go build -o bin/bt . ok, gofmt/vet leer
- [x] Alle E2-Task-beans (T1-T5) auf completed (verifiziert, keine Re-Issue nötig)
- [ ] Epic bt-aq5s Tag to-review (NICHT completed — PO-Gate) — AUSSER Scope für
      diesen Agent-Lauf (Dispatch-Vorgabe: "bt-aq5s Status-/Tag-Änderungen =
      Supervisor-Ritual"); T6-Code/Tests/beans-Pflege der Task-Ebene sind fertig,
      das Epic-Tagging bleibt dem Supervisor/PO überlassen
- [x] Selbst-Review dokumentiert: Maus bewusst NICHT in E2 (E5-Scope), Windowing
      wiederverwendet (kein Zweitbau), I01/I03/Q01 als geschlossen referenziert
      (Task 2/4)
- [x] Commit docs: E2-Abschluss
- [ ] ce-nsp-auto Handover-Prompt für E3 (Mutationen) erzeugt (separater Schritt nach
      diesem Commit, nicht Teil der Code-Änderung)


## Übernommene Fast-Follows aus E2-T5-Review (PFLICHT)
- [x] Regressionstest backlogList-Staleness: Cursor tief parken, backlogVisible via Suche/Filter shrinken (view==viewBacklog), nächster Backlog-Key → Recovery asserten; Kommentar in view_browse_backlog.go präzisieren (Render-Artefakt möglich, nicht nur move-Bound)
- [x] Refactor: renderBeanAccordionPane(bean,w,h,focused) extrahieren — Dedup renderDetailPane/renderBacklogDetailPane (~20 Zeilen Hand-Sync-Kopie)
- [x] I01 aus T5-Smoke: esc in Detail-Fokus ist No-Op (nur j/tab verlassen) — als PO-Hinweis in Epic-Abschlussmeldung aufnehmen, kein Fix (T2-Verhalten)
- [x] I02 aus T5: kein Sort-Mode-Indikator sichtbar — PO-Entscheidungspunkt in Abschlussmeldung

## Summary

**Fast-Follow 1 (backlogList-Staleness-Regressionstest):** neuer Test
`TestKeyBacklogResyncsStaleCursorBeforeMove` (view_browse_backlog_test.go) —
10-Bean-Fixture (8 task/2 bug, alle todo/parentless), Cursor auf Index 9 geparkt,
`m.filterType = {"bug": true}` DIREKT gesetzt (exakt das Feld, das `f`+space über
den GETEILTEN keyFilterMenu-Handler mutiert hätte — backlogList bleibt dabei
unberührt, siehe Datei-Kommentar) → backlogVisible() schrumpft von 10 auf 2,
Cursor 9 ist jetzt out-of-range. Nächster Backlog-Key (down) → asserted: Cursor in
[0,2), backlogSelected() != nil und vom richtigen (gefilterten) Typ, UND die
recoverte Selektion erscheint im gerenderten viewBacklog()-Output (Beweis: kein
Render-Artefakt nach Recovery). RED-Beweis erbracht: `m.backlogList.setLen(len(vis))`
in keyBacklog (view_browse_backlog.go) temporär deaktiviert (`_ = vis` statt
setLen-Aufruf) → Test schlägt exakt an der erwarteten Stelle fehl
(`cursor not recovered: cursor=9, len(vis)=2`) → Fix zurückgesetzt → GREEN. Damit
ist bewiesen, dass der Test den bestehenden Resync-Fix wirklich bewacht, nicht nur
zufällig grün ist.

Kommentar in view_browse_backlog.go (Datei-Header) präzisiert: die alte Aussage
"rendering itself never depends on backlogList.length (only .cursor), so this is
purely a keep future move() bounds correct concern, not a render bug" war
untertrieben. Tatsächlich: JEDER Render zwischen Schrumpfen und nächstem
Backlog-Tastendruck (Live-Tippen im Suchfeld, Facetten-Toggle, jeweils während
m.view==viewBacklog) ruft backlogRows/renderBeanAccordionPane mit dem VERALTETEN
Cursor gegen die BEREITS geschrumpfte backlogVisible() auf — backlogRows
highlightet nur bei i==pos, ein Cursor jenseits der neuen (kürzeren) vis
highlightet also GAR NICHTS (der `▌`-Balken verschwindet); renderBeanAccordionPane
fällt dann fälschlich auf "(no selection)" zurück, obwohl eine echte Selektion
existiert. Beides kosmetisch (kein Panic, kein Index-Out-of-Range — windowStart/
windowAround clampen das Fenster selbst weiterhin), aber ein SICHTBARES
Render-Artefakt beim Tippen, nicht nur eine latente move()-Bound-Gefahr.

**Fast-Follow 2 (renderBeanAccordionPane-Extraktion):** neue Funktion
`renderBeanAccordionPane(b *data.Bean, w, h int, focused bool) string`
(view_browse_repo.go, direkt bei renderDetailPane) trägt jetzt den gemeinsamen
Body (bodyW/accW-Berechnung, beanSections+renderAccordion+strings.Split,
"(no selection)"-Placeholder, renderPane-Wrap) aus renderDetailPane UND
renderBacklogDetailPane. Beide Aufrufer reduzieren sich auf reines
Bean-Resolving aus ihrem jeweiligen Cursor-Typ (treeNode-Slice+orphan-Guard bzw.
backlogList-Index in []*data.Bean) + einen Aufruf. T5s Scope-Entscheidung
("kein view_browse_repo.go-Edit, Duplikat akzeptiert") ist damit explizit durch
T6 überschrieben, mit Begründung in beiden Datei-Kommentaren dokumentiert.
Reiner Refactor: TestTreeGolden/TestTreeGoldenDeterministic UND
TestBacklogGolden/TestBacklogGoldenDeterministic 2x ohne -update grün, `git status`
auf testdata/ zeigt keine Änderung — Verhalten byte-identisch bestätigt.

**Verifikation:** `go test ./... -count=1` 2x grün (138 Einzeltests via `-v`-Zählung),
`go test ./internal/tui/ -race -count=1` grün, `gofmt -l .` leer, `go vet ./...`
leer, `go build -o bin/bt .` UND `make build` ok. Beide Goldens (Tree+Backlog)
je 2x ohne -update grün, kein Drift.

**beans-Pflege:** T1-T5 (bt-ms0k, bt-2jve, bt-4ep2, bt-9ldr, bt-gzu6) waren bereits
alle `completed` (durch die jeweiligen Task-Abschlüsse selbst) — nur verifiziert,
keine Re-Issue nötig. Epic bt-aq5s Tag `to-review` (nicht `-s completed` —
PO-Gate, lean-stack-Prinzip "der ausführende Agent schließt NICHT") wurde in
diesem Lauf BEWUSST NICHT gesetzt — Dispatch-Vorgabe schließt bt-aq5s
Status-/Tag-Änderungen als "Supervisor-Ritual" explizit aus dem Scope dieses
Agent-Laufs aus; bleibt dem Supervisor/PO überlassen.

**README.md:** Status-Abschnitt auf "E1+E2 fertig" aktualisiert (E3-E6 offen).
Keybinding-Tabelle "Stand E1" → "Stand E2", neue Zeilen für `/` (Suche),
`f` (Facetten-Filter), `X` (Filter-Reset), `b` (Backlog-View), `S` (Sort-Toggle),
`1`-`4` (Accordion-Section-Sprung), `enter` (Beziehungs-Sprung) ergänzt; `tab`
Beschreibung präzisiert (jetzt echter Accordion-Fokus statt E1-Platzhalter).

**tmux-Dogfooding-Smoke (bin/bt im eigenen Repo, 220x50):** Tree zeigt
bt-apmy-Milestone, expandiert zu allen Epics inkl. E2 (bt-aq5s) mit allen 6
Task-Kindern. tab auf bt-2jve (hat Parent bt-aq5s + Blocked-By bt-ms0k) öffnet
Detail-Fokus; Ziffer `3` öffnet Beziehungen (zeigt Parent+Blocked-By-Gruppen
korrekt); l+k navigiert zum Blocked-By-Feld (bt-ms0k), enter springt — Tree-Cursor
folgt zu bt-ms0k, Detail-Fokus verlässt sich automatisch (Task-2-Verhalten
bestätigt). `/` mit "T4" filtert den Tree live auf bt-9ldr (Live-Substring-Match
bestätigt). `f` öffnet das Facetten-Menü, Toggle auf Status "todo" narrowed den
Tree sichtbar (nur noch todo-Milestones/Epics sichtbar). `b` wechselt zu Backlog
MIT demselben aktiven Filter (geteilter Zustand bestätigt) — Backlog selbst ist
in diesem Dogfooding-Repo leer (`beans list --ready` bestätigt: alles ist unter
dem einen Milestone geparkt, kein echtes Backlog-Bean vorhanden — erwartetes
Verhalten, kein Bug). `S` auf leerem Backlog: kein Crash (2x gedrückt, sauber).
`q`+`y` beendet sauber über Quit-Confirm.

**NEUER Fund beim Smoke (nicht Teil der PFLICHT-Fast-Follows, hier nur
dokumentiert):** `X` (FilterClear) hat in der Backlog-View KEINE Wirkung —
keyBacklog() (view_browse_backlog.go) hat keinen `case keybind.Matches(msg,
keys.FilterClear)`-Zweig (anders als keyTree in update.go, das X explizit als
Top-Level-Reset verdrahtet, Zeilen ~350-356). Verifiziert im Smoke: Filter aus
Tree gesetzt, `b` zu Backlog, `X` gedrückt → `b` zurück zu Tree zeigt den Filter
weiterhin aktiv (kein Reset). Kein Panic, kein Datenverlust — rein ein fehlender
Tasten-Case. Nicht in T6s Fast-Follow-Liste enthalten, daher hier NUR
dokumentiert (kein Fix) — siehe PO-Hinweis-Liste unten für Bug-Code B01.

## Selbst-Review (E2-Gesamt, implementation-plan.md »Epos-Rituale« Muster)
- Spec-Coverage: V2 Master-Detail+Fokus+Beziehungs-Nav ✓ (T2) · V3 Backlog ✓ (T5)
  · V4 Accordion Meta/Body/Beziehungen/Historie ✓ (T1) · US-02/US-03/US-05/US-09
  (design-spec §12) ✓.
- I01 (expanded-Map Copy-on-Write) geschlossen Task 2 (cloneBoolMap,
  TestSetExpandedDoesNotMutateSharedMapAcrossModelCopies) · I03 (kanonischer
  Sort statt Zweit-Logik) geschlossen Task 1 (Export SortBeans/StatusRank/
  PriorityRank) + Task 4 (Orphan-Bucket-Anwendung) · Q01 (Init() nil-Guard)
  geschlossen Task 2 (TestInitNilClientReturnsErrorMsgInsteadOfPanicking) — alle
  drei PFLICHT-Items aus dem T8-Opus-Review (bean bt-7jr8) erfüllt.
- I02 (flattenTree/collectOrphans-Cache) bewusst NICHT umgesetzt — Bean-Body
  bt-aq5s sagt explizit "Notiz, kein Auto-Fix" (YAGNI, kein Profiling-Anlass).
- Maus bewusst NICHT E2 (design-spec §12 E5) — nirgends angefasst. Windowing
  (windowStart/windowAround, E1 Task 8) in JEDER neuen Liste wiederverwendet
  (Tree-Filter Task 3/4, Backlog Task 5) — kein Neubau.
- Konsolidierung ggü. devd bestätigt: EIN Filter-Zustand/-Menü (Task 4,
  box_filter_facets.go) statt devds zwei parallelen Implementierungen, EIN
  focusedBean()-Dispatcher (Task 2/5) statt separater Backlog-/Tree-Fokus-Logik,
  UND (neu, T6) EIN renderBeanAccordionPane statt zwei hand-synchronisierter
  Pane-Renderer.
- Golden-Determinismus: tree.golden UND backlog.golden je 2x ohne -update grün
  nach dem T6-Refactor — kein stiller Drift.

## PO-Hinweis-Liste für Epic-Abschlussmeldung

| Code | Schwere/Prio | Beschreibung | Empfehlung | Status |
|------|------|--------------|------------|--------|
| I01 | low | esc in Detail-Fokus ist No-Op (nur j/left oder tab verlassen ihn) — bestehendes T2-Verhalten, gilt identisch Tree+Backlog | So lassen (kein Blocker) ODER esc als dritten Exit-Weg in keyDetailFocus ergänzen — PO-Entscheidung, kein E2-Scope-Fehler | 🟡 Unklar (PO-Entscheidung offen) |
| I02 | low | Kein sichtbarer Sort-Modus-Indikator im Backlog-Pane-Titel (devd hat "Backlog [sort: priority]", design-spec V3/US-09 verlangt es nicht explizit) | Bewusst nicht nachgerüstet (Scope-Treue) — PO kann für E5/Polish vormerken, falls beim Dogfooding vermisst | 🟡 Unklar (PO-Entscheidungspunkt) |
| B01 | low | `X` (Filter-Reset) wirkt in der Backlog-View NICHT — keyBacklog() hat keinen FilterClear-Case (anders als keyTree). Verifiziert im T6-Smoke: Filter bleibt beim Wechsel Backlog→Tree aktiv trotz X-Druck in Backlog. Kein Crash/Datenverlust, rein fehlender Tasten-Case | Kleiner Fix (ein `case keybind.Matches(msg, keys.FilterClear): m = m.clearFacets(); ...` in keyBacklog, analog zu keyTree/update.go:350-356) — Vorschlag: E3-Start oder eigener Fast-Follow-Task, nicht rückwirkend in T6 gefixt (außerhalb der PFLICHT-Fast-Follow-Liste dieses Tasks) | 🟣 Offen |
