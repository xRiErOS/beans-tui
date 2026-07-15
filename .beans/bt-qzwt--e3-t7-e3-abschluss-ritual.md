---
# bt-qzwt
title: E3 T7 — E3-Abschluss-Ritual
status: completed
type: task
priority: high
created_at: 2026-07-15T00:28:17Z
updated_at: 2026-07-15T05:24:57Z
parent: bt-gzcu
blocked_by:
    - bt-ppzb
---

Ziel: E3-Abschluss-Ritual (implementation-plan.md »Epos-Rituale« -> Epos-
Abschluss), mirrors E2 Task 6 (bt-enrd).

Plan: docs/plans/v1-port/epic-E3-plan.md »Task 7«.

## Akzeptanz
- [x] go test ./... -count=1 (2x hintereinander) grün, go build -o bin/bt . ok,
      gofmt -l . leer, go vet ./... leer
- [x] Manueller Dogfooding-Smoke (tmux): s/t/a/B/c/e/ctrl+e/d auf realen Beans im
      Scratch-Repo durchgespielt, Terminal-Ausschnitt als Beleg im Commit-Body
- [ ] beans pflegen: Task-beans (T1-T6) auf completed, Epic bt-gzcu Tag
      `to-review` (NICHT -s completed -- PO-Gate, lean-stack-Prinzip "der
      ausführende Agent schließt NICHT")
      -- T1-T6 VERIFIZIERT bereits completed (kein Agent-Handeln nötig). Epic-Tag
      bewusst NICHT gesetzt: laut T7-Auftrag explizit "NOT in scope: Epic bt-gzcu
      status/tag (supervisor ritual)" -- Box bleibt offen für den Supervisor-Schritt.
- [x] Selbst-Review im Commit-Body (Spec-Coverage V7/V8, Konsolidierung ggü. devd,
      bewusste Scope-Cuts: kein Toast/E5, kein Type-Hierarchie-Client-Check bei
      Create, kein Blocking-Zyklen-Ausschluss)
- [x] Commit `docs: E3-Abschluss` (Refs: bt-qzwt)
- [ ] Skill `ce-nsp-auto` -> Handover-Prompt für E4 (Command-Center & Review-
      Cockpit, bean-Suche via `beans list --json`)
      -- bewusst NICHT ausgeführt: laut T7-Auftrag explizit out of scope.


## Übernommene Findings aus E3-T6-Review (PFLICHT)
- [x] I01: epic-E3-plan.md Task-6-Sektion: ERRATUM-Pointer oben ergänzt (Kinder→Roots,
      nicht verwaist) -- surgical, Zeilen ~1039-1124 selbst unverändert gelassen
      (Plan-Historie bleibt nachvollziehbar), nur ein Blockquote-Pointer direkt nach
      der Task-6-Überschrift.
- [x] Q01: Empirisch geprüft (scratch repo, zwei Beans A/B, A `--blocked-by` B,
      `beans delete` B, `cat` A's Frontmatter -- beide Richtungen blocked_by UND
      blocking). BEFUND: JA, `beans delete` löscht Blocking/BlockedBy-Referenzen
      ANDERER Beans still (kein CLI-Warnhinweis), exakt dasselbe Verhalten wie beim
      bereits bekannten Parent-Feld-ERRATUM. deleteBox-Warntext erweitert (neue
      `delLinks`-Zeile, singular/plural) + Regressionstests (data-layer:
      TestDeleteClearsOtherBeansBlockedByReference/...BlockingReference; TUI-layer:
      6 neue Tests für countLinkedBeans/deleteBox-Rendering).
- [x] I02: Singular-Fall implementiert (switch-Branch delChildren 0/1/>1, "1 Kind
      verliert den Parent — wird zur eigenen Wurzel" vs. "N Kinder verlieren...").
      Test TestDeleteBoxSingularChildWording (ms-1, genau 1 Kind).
- [x] I03: TestDeleteLastBeanInRepoClearsCursorGracefully -- Repo mit genau einem
      Bean, delete, Reload auf leeren Index, cursorID klemmt auf "", View() rendert
      ohne Panic.

## Zusätzlicher Befund aus dem Smoke-Test (nicht Teil der T6-Findings)
Während des E2E-Smokes leer gelaufener `s`-Mutationsversuch auf frisch per
`beans create` angelegten Scratch-Beans: der von `beans show`/`list` gemeldete
ETag wich vom internen `update --if-match`-Vergleichswert ab (CONFLICT), obwohl
die Datei seit der Erstellung unangetastet war -- reproduzierbar über mehrere
Beans/Versuche hinweg, geheilt durch EINEN unconditional `beans update` (ohne
--if-match) pro Bean. Das ist eine BREITERE Ausprägung des bereits in
internal/data/testrepo_test.go dokumentierten "tags"-ETag-Bugs (dort nur für
Beans mit handgeschriebenem tags-Block beobachtet) -- hier trat er auch bei
taglosen, batch-erzeugten Beans auf. Erklärt retroaktiv, warum der T7-Auftrag
explizit ein "Scratch-Repo MIT WARM-UP" verlangte. m.err zeigte den Konflikt in
diesem Fall nicht sichtbar an: `applyMutationResult` setzt die Konflikt-Meldung,
aber der SOFORT ausgelöste unconditional Reload (`applyLoaded`) setzt `m.err`
bei Erfolg unmittelbar wieder auf "" zurück -- die Statuszeile blitzt daher nur
für die Dauer des Reload-Roundtrips auf (lokal <100ms), praktisch unsichtbar für
eine reale PO-Session. Nicht als I0x-Finding dieses Reports geführt (keine
T6-Review-Herkunft), sondern hier dokumentiert für die Epic-Abschlussmeldung --
Empfehlung: eigenes Bean für E4/E5 (Toast-Einführung würde dies vermutlich lösen,
da Toasts i.d.R. eine Mindest-Anzeigedauer statt "bis zum nächsten Render"
haben).

## Verifikation (T7)
- `command go test ./... -count=1` grün; zusätzlich `command go test ./...
  -race -count=1 -timeout 600s` 2x hintereinander grün (127.3s, 126.7s).
- `command gofmt -l .` leer, `command go vet ./...` leer, `command make build`
  ok (bin/bt).
- Goldens (`-run Golden`) 2x hintereinander byte-identisch grün.
- `-short` Lauf: ~4.9s gesamt (überspringt die 7 huh-Drive-Tests aus T4/T6,
  ~119s Ersparnis, siehe skipSlowHuhDriveInShortMode-Doc-Stamp).
- E2E-Smoke (tmux, Scratch-Repo `e3-smoke-repo`, geheilt): volle
  Mutationssequenz s→t→a→B→c→e→ctrl+e→d in EINER Session, jeder Schritt auf
  Disk verifiziert (status→in-progress, tags→[smoke], parent-Reassignment,
  blocking→[sm-uu05], Bean-Create unter Epic mit Cursor-Sprung, Title-Edit,
  Body via Fake-$EDITOR, Delete mit Cursor-Klemmung auf Nachbarn).
