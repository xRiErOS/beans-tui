---
# bt-p1uz
title: E3 T3 — Parent-/Blocking-Picker mit Zyklen-Ausschluss
status: in-progress
type: task
priority: high
created_at: 2026-07-15T00:26:21Z
updated_at: 2026-07-15T01:40:30Z
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
- [ ] internal/data/mutations.go: AddBlocking/RemoveBlocking/SetBlocking ergänzt,
      Tests analog client_mut_test.go (Roundtrip + Ein-Kommando-Diff)
- [ ] internal/data/hierarchy.go: validParentTypes + CollectDescendants (BFS über
      idx.Children -- KEIN Map-Neuaufbau wie beans-src, idx.Children ist diese Map
      schon; visited-Guard gegen Hand-Edit-Zyklen) + EligibleParents-Fassade
- [ ] box_picker_parent.go: Single-Select, ausgeschlossen: self, Nachkommen,
      Typen außerhalb validParentTypes; "(Kein Parent)"-Zeile ganz oben (Port
      beans-src clearParentItem); enter wendet SOFORT an (SetParent/RemoveParent)
      + schließt; Cursor startet auf aktuellem Parent
- [ ] box_picker_blocking.go: Toggle-Multi-Select über alle Beans außer sich selbst
      (Pending-State + ●/○-Indikator wie beans-src blockingpicker.go), enter
      diff't -> EIN mutateCmd(SetBlocking), esc verwirft
- [ ] go test ./... grün, gofmt/vet leer


## Übernommene Findings aus E3-T2-Review (PFLICHT)
- [ ] I1a: Test esc-aus-Tag-Input → Picker bleibt offen, tagPending intakt
- [ ] I1b: Test toggle-off-dann-on → leerer Diff → kein Cmd
- [ ] I2: SetTags-Doc-Zeile: gleicher Tag in add+remove → remove gewinnt (Upstream-Resolver-Reihenfolge); gleiche Zeile für neues SetBlocking
