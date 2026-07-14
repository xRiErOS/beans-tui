---
# bt-ms0k
title: 'T1 Detail-Accordion-Port: Sections Meta/Body/Beziehungen/Historie + kanonischer Sort (I03-Export)'
status: completed
type: task
priority: high
created_at: 2026-07-14T21:57:13Z
updated_at: 2026-07-14T22:08:15Z
parent: bt-aq5s
---

Ziel: Detail-Accordion Meta/Body/Beziehungen/Historie portieren (devd accordion.go +
view_detail_issue.go als Muster), read-only (Edit-Felder sind E3-Scope). Body-Section
rendert via glamour (Port devd editor.go:102-126 glowRender). Beziehungen-Section
(Parent/Children/Blocking/BlockedBy) braucht kanonische Sortierung für Blocking/
BlockedBy (Children ist über idx.Children bereits sortiert) -> I03-Vorbereitung:
data.SortBeans/StatusRank/PriorityRank exportieren.

Plan: docs/plans/v1-port/epic-E2-plan.md »Task 1«.

## Akzeptanz
- [x] data.SortBeans/StatusRank/PriorityRank exportiert + getestet (index_test.go)
- [x] accordion.go: renderAccordion (Port devd accordion.go:309-373, ohne Edit-Felder),
      exklusiv offene Section, Ziffern-Header [1]..[4]
- [x] view_detail_bean.go: beanSections(idx, bean, bodyW) — IMMER 4 Sections
      (Meta/Body/Beziehungen/Historie), Body glamour-gerendert (leer -> Platzhalter),
      Beziehungen mit relationField (jump-only, kein Edit), dangling Blocking/
      BlockedBy-IDs als "(unresolved: id)" nicht sprungfähig
- [x] go test ./... grün, gofmt/vet leer



## Summary of Changes

- `internal/data/index.go`: exportiert `SortBeans`/`StatusRank`/`PriorityRank` (dünne Wrapper über die bestehende unexported `sortBeans`/`rank`-Logik, Single-Source-Sort bleibt gewahrt) — I03-Vorbereitung. Tests in `internal/data/index_test.go`.
- `internal/tui/accordion.go` (neu): `relationField`, `accordionSection`, `renderAccordion` (Port devd `accordion.go:309-355`, ohne Edit-Feld-Konzept), `fieldStrip` (Port `accordion.go:360-373`), `glowRender` (Port `editor.go:102-126`, glamour, notty-Style unter Ascii-Profil für Golden-Determinismus). Tests in `internal/tui/accordion_test.go`.
- `internal/tui/view_detail_bean.go` (neu): `beanSections(idx, bean, bodyW)` — immer 4 Sections (Meta/Body/Beziehungen/Historie). Meta nutzt bestehende Theme-Helfer (StatusStyle/TypeStyle/Priority/tagsInline), Body rendert via glowRender (Platzhalter bei leer), Beziehungen gruppiert Parent/Children/Blocking/Blocked-By (Blocking/BlockedBy via `data.SortBeans` kanonisch sortiert nach Resolve, Children bereits über idx.Children sortiert), dangling IDs als `(unresolved: id)` mit `relationField.beanID == ""` (nicht sprungfähig). Tests in `internal/tui/view_detail_bean_test.go`.
- `go.mod`/`go.sum`: glamour v1.0.0 als direkte Dependency (`go get` + `go mod tidy`, kein Netzwerkzugriff nötig — Modul-Cache).
- Keine Verdrahtung (Task 2 Scope): `view_browse_repo.go`s Platzhalter-Detail-Pane bleibt unverändert, keine neuen model-Felder in `types.go`.
- `go test ./... -count=1`, `gofmt -l .`, `go vet ./...` alle clean.
