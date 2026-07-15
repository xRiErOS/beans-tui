---
# bt-8v69
title: E3 T2 — Tag-Picker
status: todo
type: task
priority: high
created_at: 2026-07-15T00:26:06Z
updated_at: 2026-07-15T00:39:30Z
parent: bt-gzcu
blocked_by:
    - bt-dlgk
---

Ziel: Tag-Picker (`t`) — Toggle-Multi-Select über vorhandene Tags (Nutzungszähler,
sortiert Anzahl desc dann alpha, Port beans-src tagpicker.go-Sortierung) + Freitext-
Neuanlage-Zeile ("+ new tag…", eigener Text-Capture-Sub-Modus wie m.searchInput).
Enter bestätigt (Diff gegen Original-Tags), esc verwirft (Port beans-src
blockingpicker.go Enter=confirm/esc=cancel-Konvention, NICHT box_filter_facets.go's
enter=close-ohne-Wirkung — hier wird echt mutiert).

DESIGN-ENTSCHEIDUNG (Plan »Task 2«, verbindlich): der Diff wird als EIN
`beans update` mit kombinierten --tag/--remove-tag-Flags ausgeführt — neue
mutations.go-Funktion SetTags(id string, add, remove []string, etag string).
N Einzel-Mutationen (AddTag/RemoveTag nacheinander) auf EINEM ETag wären eine
Konflikt-Kaskade (die erste gewinnt, jede weitere läuft in ErrConflict).
AddTag/RemoveTag bleiben für Einzelfälle bestehen.

Plan: docs/plans/v1-port/epic-E3-plan.md »Task 2«.

## Akzeptanz
- [ ] internal/data/mutations.go: SetTags (EIN update, --tag/--remove-tag kombiniert),
      Test TestSetTagsAddsAndRemovesInOneCall (client_mut_test.go-Muster)
- [ ] internal/data/tags.go: ValidTagName gegen `^[a-z][a-z0-9]*(-[a-z0-9]+)*$`
      (Tag-Regex aus Epic-Body bt-gzcu), Tests inkl. Negativ-Fälle
- [ ] box_picker_tag.go: Menü mit Nutzungszählern (collectTagCounts über idx.ByID,
      deterministisch sortiert count desc/alpha — KEINE Map-Walk-Ordnung, Lehre aus
      tagFilterOptions-ERRATUM), space/x toggelt Pending-State, n öffnet
      Neuanlage-Input (invalider Name -> Inline-Fehler, kein Submit)
- [ ] enter diff't Pending gegen Original -> EIN mutateCmd(SetTags), ETag frisch via
      beanETag; keine Änderungen -> kein Cmd; esc verwirft alles
- [ ] go test ./... grün, gofmt/vet leer
