---
# bt-pt1r
title: Projekt-Slug auch in Picker-/Overlay-Listen weglassen
status: completed
type: task
priority: low
created_at: 2026-07-20T08:19:27Z
updated_at: 2026-07-20T09:14:56Z
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

## Summary

Die Kuerzungslogik aus bt-pl5p (`shortBeanID` + `m.beanIDPrefix()`) wird
wiederverwendet, nicht neu gebaut: `relationRowPrefix` bekommt denselben
`slug`-Parameter, den `treeRowText`/`backlogRowText` schon tragen. Damit
kuerzen alle drei Zeilen-Renderer ueber DIE EINE Regel.

- `relationRowPrefix(rel, slug)` — neu parametrisiert (view_detail_bean.go)
- `buildParentItems(idx, b, slug)` / `buildBlockingItems(idx, selfID, slug)`
  reichen `m.beanIDPrefix()` durch
- Detail-Pane: `relationRow` reicht bewusst `""` durch → volle, kopierbare
  ID bleibt dort stehen (Entscheidung aus bt-pl5p), abgesichert durch
  `TestRelationRowKeepsFullIDInDetail`
- `pickerItem.id` bleibt die VOLLE ID — Mutations-Ziel, kein Anzeigetext
  (`TestPickerItemIDStaysFullForMutation`)
- Command-Palette geprueft: zeigt keine Bean-IDs, kein Handlungsbedarf.
  Tag-Picker ebenso (zeigt Tags, keine IDs).

## Test-Output

Voller Lauf `command go test ./...` gruen:

    ok  github.com/xRiErOS/beans-tui/cmd            0.827s
    ok  github.com/xRiErOS/beans-tui/internal/config 0.267s
    ok  github.com/xRiErOS/beans-tui/internal/data   4.124s
    ok  github.com/xRiErOS/beans-tui/internal/theme  1.024s
    ok  github.com/xRiErOS/beans-tui/internal/tui  150.521s

tmux 80x30 gegen sproutling: Blocking-Picker zeigt `xglu`, `s9d0`, `eq67`
statt `sproutling-xglu` — kein Overflow, kein Wrap-Bug.

Commit de46018.

## Deviations

Keine. Alle vier Akzeptanzpunkte erfuellt. Keine Golden-Datei beruehrt.
