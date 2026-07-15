---
# bt-dlgk
title: E3 T1 — Status/Type/Prio-Menü + B01-Fix + Enum-Single-Source + Mutation-Infra
status: todo
type: task
priority: high
created_at: 2026-07-15T00:25:55Z
updated_at: 2026-07-15T00:25:55Z
parent: bt-gzcu
---

Ziel: Status-/Type-/Prio-Menü (`s`, EIN kombiniertes Value-Menü mit 3 Gruppen —
design-spec §7 reserviert nur EINEN Key für dieses Cluster, nicht drei) + B01-Fix
(keys.FilterClear-Case fehlt in keyBacklog, aus E2-Abschluss übernommen) +
Enum-Single-Source (data.StatusValues/TypeValues/PriorityValues, ersetzt
box_filter_facets.go's eigenes Duplikat) + geteilte Mutation-Cmd/ETag-Konflikt-Infra
(erste Mutation Ende-zu-Ende verdrahtet; jede Folge-Task nutzt dieselbe Infra statt
eigener Cmd/Msg-Typen).

Plan: docs/plans/v1-port/epic-E3-plan.md »Task 1«.

## Akzeptanz
- [ ] B01: X (FilterClear) wirkt in keyBacklog wie in keyTree (Regressionstest)
- [ ] data.StatusValues()/TypeValues()/PriorityValues() exportiert (kanonische
      Tier-Reihenfolge aus statusOrder/typeOrder/priorityOrder); box_filter_facets.go
      buildFilterItems auf diese Single Source umgestellt (Duplikat-Fix)
- [ ] box_menu_value.go: kombiniertes Status/Type/Priority-Menü (modalBox/menuList,
      3 Gruppen wie box_filter_facets.go facetHead-Muster), enter wendet SOFORT an
      + schließt (Port beans-src statuspicker.go Enter-Semantik)
- [ ] keyNodeAction-Scaffold in update.go: zentraler Dispatch für s/t/a/B/c/d/e direkt
      neben dem bestehenden Refresh-Check in handleKey, guarded auf
      focusedBean()!=nil (außer c/Create)
- [ ] messages.go: mutationDoneMsg + mutationCmd-Producer, applyMutationResult
      (jede Mutation, Erfolg wie Fehler, reloaded unconditional via loadCmd;
      ErrConflict zusätzlich Statuszeilen-Hinweis, kein Toast — E5)
- [ ] go test ./... grün, gofmt/vet leer
