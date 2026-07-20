---
# bt-pl5p
title: Projekt-Slug aus Bean-IDs in der linken Pane weglassen
status: completed
type: task
priority: normal
created_at: 2026-07-20T07:26:22Z
updated_at: 2026-07-20T08:13:38Z
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


## Grounding 2026-07-20 (Investigator, read-only)

**Diese Aufgabe wird als Paket mit bt-fy5d + bt-oqsv + bt-pl5p umgesetzt** — die drei teilen
sechs Golden-Dateien (`tree.golden`, `backlog.golden`, `browse_flat.golden`,
`browse_boxform.golden`, `chrome.golden` + boxform-Detail). Einzeln bearbeitet wuerde jede
Aenderung die Golden der anderen ueberschreiben. Ein Implementer, drei Commits, Golden am
Ende EINMAL regeneriert.

### Fundorte je Punkt

**Footer-Keys (bt-fy5d)**
| Symbol | Ort | Rolle |
|---|---|---|
| `renderBindings()` | `view.go:128` | Bindings -> kolorierte Key/Desc-Paare, " · "-getrennt |
| `footer()` | `view.go:152` | wrapped den Hint auf Breite |
| `contextualLocalHint()` | `footer_context.go:193` | tauscht Footer-Bindings bei offenem Overlay |
| `*LocalBindings()` | `footer_context.go:45-170` | 7 Funktionen (filter/value/tag/parent/blocking/tag-mgmt) |
| `browseRepoChrome()` | `view_browse_repo.go:945` | baut head + localKeys |
| Footer-Render | `view_browse_repo.go:1091` | `content += "\n" + localKeys + "\n" + status` |
| `helpGroups()` | `keymap.go:196` | Drift-Guard: jede Binding genau EINE Gruppe |

**Leere Zeile (bt-oqsv)**
| Symbol | Ort | Rolle |
|---|---|---|
| `statusBar()` | `view.go:79-95` | liefert "" oder rechtsbuendigen Indicator — Quelle der Leerzeile |
| Hoehen-Konstante | `mouse.go:112` | `footH := lipgloss.Height(localKeys) + 2` — **muss mitgezogen werden, sonst Klick-Offset** |
| View-Komposition | `view_browse_repo.go:1087-1092` | haengt `"\n" + status` an, auch wenn leer |
| Beleg | `testdata/tree.golden:29` | die sichtbare Leerzeile |

**Projekt-Slug (bt-pl5p)**
| Symbol | Ort | Rolle |
|---|---|---|
| `treeRowText()` | `view_browse_repo.go:494-510` | Zeile 509 `theme.Key.Render(b.ID)` |
| `backlogRowText()` | `view_browse_backlog.go:169` | Zeile 169 `theme.Key.Render(b.ID)` |
| geteilt von | `view_browse_flat.go:82` (`flatRows`), `view_browse_backlog.go:181` (`backlogRows`) | **zwei** Render-Stellen, nicht eine |

### Fallstricke
- **`mouse.go:112`** (`+2`) haengt an der Footer-Hoehe. Wer die Leerzeile entfernt und die
  Konstante vergisst, versetzt jeden Maus-Klick um eine Zeile. Regressionstest dafuer.
- **Zwei ID-Render-Stellen** (`treeRowText` + `backlogRowText`) — die Flat-Ansicht nutzt die
  zweite. Nur eine zu aendern laesst Tree und Flat auseinanderlaufen.
- **`helpGroups()`-Drift-Guard** (`TestHelpGroupsCoverEveryBindingExactlyOnce`) schlaegt fehl,
  wenn eine Binding aus dem Footer faellt, ohne dass die Gruppen stimmen. Footer-Ausduennung
  darf die keymap-Registrierung NICHT anfassen — nur die Anzeige.
- Betroffene Tests: `footer_context_test.go`, `keymap_test.go`, `chrome_test.go`,
  `mouse_test.go` (:485, :1012, :1039), `view_browse_{repo,backlog,flat}_test.go`,
  `overlay_shortcuts_test.go`, `primitives_test.go:395`, `archive_placeholder_test.go`,
  `update_test.go`, `view_tag_management_test.go:275`.

### Kollisions-Hinweis fuer den Dispatch
Beruehrt `view_browse_repo.go` + `mouse.go` — dieselben Dateien wie bt-ze10 (Detail-Scroll).
**Erst nach dessen Commit starten.**


## Summary

Commit 2f531b5 `feat(browse): drop repo prefix from bean IDs in rows`.

`shortBeanID(id, slug)` (view_browse_repo.go) schneidet exakt den Prefix
`<slug>-` des AKTUELL geoeffneten Repos ab; `slug` kommt aus
`m.beanIDPrefix()` -> `data.RepoSlug(m.repoDir)` (.beans.yml `beans.prefix`,
also dieselbe Quelle, aus der die IDs selbst gemuenzt werden). NICHT bis zum
letzten "-" geschnitten: Fremd-/Dangling-IDs (`lean-stack-58j0` in einem
`sproutling`-Repo) und mehrsegmentige Prefixe bleiben eindeutig. Leerer Slug
oder leerer Rest -> ID unveraendert.

Beide Render-Stellen geaendert (Grounding-Fallstrick): `treeRowText`
(view_browse_repo.go) UND `backlogRowText` (view_browse_backlog.go, auch von
`flatRows` in view_browse_flat.go genutzt). Der Slug wird einmal pro Render in
`treeRows`/`backlogRows`/`flatRows` aufgeloest, nicht pro Zeile (RepoSlug macht
File-IO).

## Entscheidungen

- **Geltungsbereich: global**, nicht hinter BT_BOXFORM. Begruendung: Das Paket
  regeneriert die Goldens ohnehin (bt-oqsv), der Platzgewinn gilt in beiden
  Modi, und eine Flag-Weiche haette Tree/Flat/Backlog uneinheitlich gemacht.
- **Detail-Pane behaelt die VOLLE ID** (Yank-/Referenz-Nutzen) — dort wurde
  nichts angefasst. Im 80-Spalten-Smoke belegt: links `vy1q`, rechts `bt-vy1q`.

## Test-Output

Neue Tests in internal/tui/bean_id_short_test.go:

    --- PASS: TestShortBeanID (7 Faelle: eigener Prefix, kurzer Prefix,
              Fremd-Prefix, leerer Slug, Prefix-ohne-Separator, leerer Rest,
              mehrsegmentiger Slug)
    --- PASS: TestTreeRowTextDropsRepoPrefix
    --- PASS: TestBacklogRowTextDropsRepoPrefix
    --- PASS: TestBacklogRowTextKeepsForeignID

Voller Lauf vor dem Commit:

    ok  github.com/xRiErOS/beans-tui/cmd            0.366s
    ok  github.com/xRiErOS/beans-tui/internal/config
    ok  github.com/xRiErOS/beans-tui/internal/data
    ok  github.com/xRiErOS/beans-tui/internal/theme
    ok  github.com/xRiErOS/beans-tui/internal/tui   149.422s

## Deviations

- **Goldens unveraendert durch diese Aenderung.** Das Golden-Fixture nutzt
  `repoDir=/tmp/bt-golden-repo` (Slug `bt-golden-repo`) mit IDs `gld-*` — kein
  Prefix-Treffer, also greift die Kuerzung dort nicht. Der Nachweis laeuft
  ueber die Unit-Tests oben plus den 80-Spalten-tmux-Smoke (linke Pane zeigt
  `vy1q` statt `bt-vy1q`).
