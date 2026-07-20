---
# bt-pl5p
title: Projekt-Slug aus Bean-IDs in der linken Pane weglassen
status: todo
type: task
priority: normal
created_at: 2026-07-20T07:26:22Z
updated_at: 2026-07-20T07:26:22Z
parent: bt-vy1q
---

**Nebenbefund N5 (PO, 2026-07-20).** Die linke Pane zeigt volle Bean-IDs wie `sproutling-btv7`. Der Projekt-Slug ist redundant — welches Repo offen ist, steht bereits im Header (`> sproutling: Browse`). Weglassen ⇒ deutlich mehr Platz fuer den Bean-Titel, der aktuell hart clamped (siehe `beans-tui-boxform-narrow.gif`: Titel auf `Go…`/`Mi…` gekuerzt).

## Ziel
In der Listen-/Tree-Zeile nur das ID-Suffix rendern (`btv7` statt `sproutling-btv7`), gewonnene Breite geht an den Titel.

## Betroffen
- `internal/tui/view_browse_repo.go` — `treeRowText`
- `internal/tui/view_browse_backlog.go` — `backlogRowText` (wird auch von der Flat-View `view_browse_flat.go` genutzt)
- ggf. `render_shared.go` (`theme.Key.Render(b.ID)`-Aufrufe)

## Zu entscheiden (im Bean festhalten, bevor umgesetzt wird)
- **Wie kuerzen?** Sicher: den Prefix des AKTUELLEN Repos abschneiden (aus `.beans.yml`/Config ableitbar); wenn die ID diesen Prefix nicht traegt (Fremd-/Dangling-Referenz), volle ID zeigen. NICHT blind bis zum letzten `-` schneiden.
- **Geltungsbereich?** Nur im Box-Modus hinter `BT_BOXFORM` (Spike-Disziplin, Goldens byte-identisch) ODER global mit Golden-Regeneration. Beruehrt `tree.golden`/`backlog.golden`/`chrome.golden`.
- **Detail-Pane?** Dort ggf. die VOLLE ID behalten (Yank/Referenz-Nutzen).

## Akzeptanz
- [ ] Entscheidung Geltungsbereich dokumentiert
- [ ] Linke Pane zeigt gekuerzte ID, Titel bekommt die Breite
- [ ] Fremd-IDs ohne Repo-Prefix bleiben vollstaendig
- [ ] Betroffene Goldens bewusst regeneriert (oder Flag-gated unveraendert)
- [ ] Voller `command go test ./...` gruen
