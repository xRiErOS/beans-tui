---
# bt-pt1r
title: Projekt-Slug auch in Picker-/Overlay-Listen weglassen
status: todo
type: task
priority: low
created_at: 2026-07-20T08:19:27Z
updated_at: 2026-07-20T08:19:27Z
parent: bt-vy1q
---

Nachlauf zu bt-pl5p (Commit `2f531b5`). Dort wurde der Repo-Prefix aus den Bean-IDs
der **linken Pane** entfernt (`sproutling-eq67` -> `eq67`, ~11 Zeichen mehr Titel).

Die **Picker-/Overlay-Listen zeigen den Slug weiterhin voll**: im Blocking-Picker steht
`sproutling-xglu`, obwohl daneben die Titel umbrechen und Platz knapp ist. Beleg: tmux
80x30 gegen sproutling nach dem Merge von bt-a3a8.

Damit ist die Darstellung jetzt **innerhalb der App inkonsistent** — kurz in der Liste,
lang im Picker.

## Ziel
Dieselbe Kuerzung in den Picker-/Overlay-Zeilen. Die Logik existiert bereits aus bt-pl5p
(schneidet exakt `<slug>-` anhand `data.RepoSlug` aus `.beans.yml beans.prefix`; fremde
und mehrteilige IDs bleiben unangetastet) — **wiederverwenden, nicht neu bauen**.

## Betroffen
- `internal/tui/box_picker_blocking.go` — Zeilen-Render
- `internal/tui/box_picker_parent.go` — Zeilen-Render
- ggf. `internal/tui/box_picker_filter.go` (gemeinsame Komponente aus bt-a3a8)
- pruefen: Command-Palette (`overlay_palette.go`) zeigt sie ebenfalls?

## Bewusst NICHT aendern
Die **Detail-Pane** zeigt weiterhin die volle ID — explizite Entscheidung in bt-pl5p
(irgendwo muss die vollstaendige, kopierbare ID stehen). Nicht versehentlich mitkuerzen.

## Akzeptanz
- [ ] Picker-Zeilen zeigen die gekuerzte ID
- [ ] Detail-Pane zeigt weiterhin die volle ID
- [ ] Fremde/mehrteilige IDs bleiben unveraendert (Test)
- [ ] voller Testlauf gruen
