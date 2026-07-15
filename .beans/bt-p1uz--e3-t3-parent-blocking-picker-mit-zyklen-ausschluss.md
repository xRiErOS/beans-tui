---
# bt-p1uz
title: E3 T3 — Parent-/Blocking-Picker mit Zyklen-Ausschluss
status: completed
type: task
priority: high
created_at: 2026-07-15T00:26:21Z
updated_at: 2026-07-15T01:55:21Z
parent: bt-gzcu
blocked_by:
    - bt-8v69
---

Ziel: Parent-Picker (`a`) + Blocking-Picker (`B`) mit Zyklen-Ausschluss (Port
beans-src parentpicker.go collectDescendants-BFS + eigene validParentTypes,
gespiegelt von beancore.ValidParentTypes: milestone->kein Parent, epic->[milestone],
feature->[milestone,epic], task/bug->[milestone,epic,feature]).

KORREKTUR ggü. erster Annahme (verifiziert via `beans update --help`): die CLI
kennt `--blocking`/`--remove-blocking` ALS EIGENE Flags -- Blocking ist KEIN rein
server-berechneter Rückwärts-Index. `B` editiert das fokussierte Bean's EIGENE
Blocking-Liste -- EIN Bean/ETag (das fokussierte selbst), keine Multi-Bean-Batch-
Mutation. Diff-Ausführung als EIN Kommando via neuer SetBlocking(id, add, remove,
etag) (gleiche Konflikt-Kaskaden-Vermeidung wie T2s SetTags); AddBlocking/
RemoveBlocking als Einzel-Wrapper für Symmetrie zur BlockedBy-Familie trotzdem
ergänzt.

Zyklen-Ausschluss gilt NUR für den Parent-Picker (Baum-Struktur, ein neuer Parent-
Zyklus würde die Tree-Flatten-Logik/Render brechen). Blocking-Picker: KEIN
Zyklen-Ausschluss (Port-Parität zu beans-src blockingpicker.go, das ebenfalls
keinen hat; ein Blocking-Zyklus ist ein logischer PO-Fehler, kein Render-Risiko —
explizit YAGNI, dokumentiert nicht stillschweigend übergangen).

Plan: docs/plans/v1-port/epic-E3-plan.md »Task 3«.

## Akzeptanz
- [x] internal/data/mutations.go: AddBlocking/RemoveBlocking/SetBlocking ergänzt,
      Tests analog client_mut_test.go (Roundtrip + Ein-Kommando-Diff)
- [x] internal/data/hierarchy.go: validParentTypes + CollectDescendants (BFS über
      idx.Children -- KEIN Map-Neuaufbau wie beans-src, idx.Children ist diese Map
      schon; visited-Guard gegen Hand-Edit-Zyklen) + EligibleParents-Fassade
- [x] box_picker_parent.go: Single-Select, ausgeschlossen: self, Nachkommen,
      Typen außerhalb validParentTypes; "(Kein Parent)"-Zeile ganz oben (Port
      beans-src clearParentItem); enter wendet SOFORT an (SetParent/RemoveParent)
      + schließt; Cursor startet auf aktuellem Parent
- [x] box_picker_blocking.go: Toggle-Multi-Select über alle Beans außer sich selbst
      (Pending-State + ●/○-Indikator wie beans-src blockingpicker.go), enter
      diff't -> EIN mutateCmd(SetBlocking), esc verwirft
- [x] go test ./... grün, gofmt/vet leer


## Übernommene Findings aus E3-T2-Review (PFLICHT)
- [x] I1a: Test esc-aus-Tag-Input → Picker bleibt offen, tagPending intakt
- [x] I1b: Test toggle-off-dann-on → leerer Diff → kein Cmd
- [x] I2: SetTags-Doc-Zeile: gleicher Tag in add+remove → remove gewinnt (Upstream-Resolver-Reihenfolge); gleiche Zeile für neues SetBlocking



## Summary

Parent-Picker (`a`) + Blocking-Picker (`B`) implementiert. data/hierarchy.go:
validParentTypes (Spiegel beancore.ValidParentTypes) + CollectDescendants
(BFS über idx.Children, visited-Guard) + EligibleParents-Fassade (self +
Nachkommen + Typ-Filter, SortBeans). mutations.go: AddBlocking/
RemoveBlocking/SetBlocking (Ein-Kommando-Diff wie SetTags). box_picker_
parent.go: Single-Select, "(Kein Parent)"-Zeile zuerst, Cursor auf
aktuellem Parent, Enter wendet sofort an (keine Pending-Diff -- ein Bean hat
genau einen Parent). box_picker_blocking.go: Toggle-Multi-Select ohne
Zyklen-Ausschluss (Design-Entscheidung g, Port-Parität beans-src), ●/○-
Indikator, Pending-Diff wie T2. Beide über overlayID-Cases (keyNodeAction/
keyOverlay/composeOverlays) verdrahtet, keine neuen Keybindings nötig
(keys.Assign/keys.Blocking existierten schon aus T1).

Finding (Smoke/client_mut_test.go): die beans-CLI validiert Blocking-Zyklen
server-seitig (VALIDATION_ERROR bei Zwei-Bean-Zyklus) -- die Design-Aussage
"kein Render-Risiko" bezieht sich NUR auf die TUI (kein Client-Pre-Filter),
nicht auf Server-Validierung; TestSetBlockingAddsAndRemovesInOneCall nutzt
daher einen frisch erzeugten dritten Bean statt der Fixture-Beans, um einen
Zyklus zu vermeiden.

Übernommene PFLICHT-Findings aus T2-Review (I1a/I1b/I2) geschlossen: I1a/I1b
waren im T2-Code bereits korrekt, nur Regressionstests gefehlt (jetzt in
box_picker_tag_test.go). I2: SetTags/SetBlocking-Doc-Zeile ergänzt
(gleiches Ziel in add+remove -> remove gewinnt, empirisch gegen beans 0.4.2
verifiziert).

Tests: go test ./... -race 2x grün, gofmt/vet leer, Goldens grün. tmux-Smoke
in Scratch-Repo (mit Warm-up-Updates) verifiziert: Zyklen-Ausschluss (Epic-
Parent-Picker zeigt nur Milestone, nicht die eigenen Feature-/Task-
Nachkommen), Blocking-Picker bietet den eigenen Nachkommen an (kein
Ausschluss) und schreibt `blocking:` korrekt in EINEM Aufruf.

Commits: 1793934 (feat), 1cf0b22 (test/PFLICHT-Nachtrag).

Notes für T4 (Create-Form, bt-y4ly): der Parent-Feld-Validator im Create-
Form (design decision e, "validiert NUR auf Existenz in idx.ByID oder leer")
ist bewusst SCHWÄCHER als EligibleParents hier -- T4 dupliziert die
Typ-Hierarchie-Regel NICHT client-seitig (Server-VALIDATION_ERROR-Pfad).
Falls T4 stattdessen EligibleParents als Picker-Feld wiederverwenden will
(statt freiem Text-Input), ist data.EligibleParents(idx, b) bereits die
fertige Fassade -- braucht nur einen *data.Bean-Entwurf mit gesetztem .Type,
um für den NEUEN (noch nicht existenten) Bean-Typ zu filtern.
