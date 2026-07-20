---
# bt-fy5d
title: 'Footer entschlacken: inline gezeigte Keys nicht doppeln'
status: completed
type: task
priority: normal
created_at: 2026-07-20T07:25:11Z
updated_at: 2026-07-20T08:14:32Z
parent: bt-vy1q
---

**Nebenbefund N2 (PO, 2026-07-20).** Im Box-Modus stehen die Feld-Hotkeys bereits salient IM Box-Rahmen (`(e) (s) (o) (u) (a) (t)`). Der Footer wiederholt sie trotzdem (`s Status · e Edit · t Tags · a Parent · r Blocking`). Redundanz — bei 80 Spalten kostet der Footer dadurch 3 Zeilen.

Beleg: `~/Obsidian/Vault/lean-stack/beans-tui/beans-tui-boxform-narrow.gif` (Footer 3-zeilig, alle Keys doppelt).

## Ziel
Wenn `boxFormEnabled()`: die Keys, die das Detail bereits inline als Badge zeigt, aus der Footer-Liste entfernen. Alles andere (tab/shift+tab, `/`, `f`, `c`, `d`, `b`, `y`, `G`) bleibt.

## Betroffen
- `internal/tui/footer_context.go` — die `*LocalBindings()`-Sets (Single Source der Footer-Zeile)
- ggf. `internal/tui/view.go` `renderBindings`
- `internal/tui/box_form_flag.go` — `boxFormEnabled()`

## Akzeptanz
- [ ] Bei `BT_BOXFORM=1` zeigt der Footer NICHT mehr `s Status`, `e Edit`, `t Tags`, `a Parent`, `r Blocking` (die inline sichtbaren)
- [ ] Bei Flag AUS ist der Footer unveraendert (Bestandsgolden byte-identisch!)
- [ ] `browse_boxform.golden` regeneriert, Footer dort kuerzer
- [ ] Test: Footer-Inhalt unter Flag AN vs AUS (assert die entfernten Keys fehlen/erscheinen)
- [ ] Kein Drift-Guard-Bruch (Keys bleiben in `helpGroups()`, nur die FOOTER-Anzeige aendert sich)
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

Commit d81e583 `feat(footer): hide box-form inline keys from the footer`
(+ Golden-Commit 2f6fe9c).

`browseRepoLocalBindings()` (view_browse_repo.go) filtert bei
`boxFormEnabled()` die Bindings heraus, deren Help-Key in `boxFormInlineKeys`
steht: `e`, `s`, `a`, `t`, `r`. Bei Flag AUS gibt die Funktion die Liste
unveraendert zurueck (frueher Return vor dem Filter) — byte-identisch.

`backlogLocalBindings()` delegiert weiterhin an `browseRepoLocalBindings()`
und duennt damit automatisch mit aus; das ist korrekt, weil der Backlog
dieselbe box-form-Detailpane rendert (`renderBeanAccordionPane`).

Footer im Box-Modus jetzt: `tab · shift+tab · / · f · c · d · b · y`.

## Entscheidungen

- **Welche Keys?** `e`/`s`/`a`/`t` sind woertliche `(x)`-Badges aus
  `detailBoxForm` (Title/Status/Parent/Tags) bzw. `panelBox("Body")`.
  `o` (Type) und `u` (Priority) sind ebenfalls Badges, standen aber nie im
  Footer-Set — nichts zu tun. `r` (Blocking) hat kein eigenes Badge, aber sein
  Gegenstand IST das Relations-Panel derselben Ansicht, und der PO hat es in
  der Akzeptanz mit aufgelistet. Deshalb mit raus.
- **Filter ueber `Help().Key`, nicht ueber Binding-Identitaet.** Das ist genau
  der String, den `renderBindings` in den Footer schreibt — die Regel liest
  sich damit als das, was sie ist: eine Anzeige-Regel.
- **Keine Aenderung an keymap.go.** Die Ausduennung ist rein visuell.

## Test-Output

Neue Tests in internal/tui/footer_boxform_test.go:

    --- PASS: TestBrowseRepoLocalBindingsDropInlineKeysWhileBoxForm
    --- PASS: TestBrowseRepoLocalBindingsUnchangedWithoutBoxForm
    --- PASS: TestBacklogFooterFollowsBrowseRepo
    --- PASS: TestBoxFormFooterKeysStayFunctional

Der Flag-AUS-Test pinnt die volle Reihenfolge als Literal
(`tab|shift+tab|/|f|s|c|d|e|b|t|y|a|r`), der letzte prueft von der anderen
Seite, dass alle fuenf entfernten Keys in `keys.helpGroups()` registriert
bleiben.

Drift-Guards gruen (Akzeptanz 5):

    --- PASS: TestHelpGroupsCoverEveryBindingExactlyOnce
    --- PASS: TestNoDuplicateBindingBetweenGlobalAndAnyLocalHintList

Voller Lauf (uncached, nach allen drei Commits):

    ok  github.com/xRiErOS/beans-tui/cmd            0.441s
    ok  github.com/xRiErOS/beans-tui/internal/config 1.121s
    ok  github.com/xRiErOS/beans-tui/internal/data   4.700s
    ok  github.com/xRiErOS/beans-tui/internal/theme  0.781s
    ok  github.com/xRiErOS/beans-tui/internal/tui  151.398s

**80-Spalten-tmux-Smoke**, BT_BOXFORM=1 — Footer 2 Zeilen statt 3:

    tab focus in · shift+tab focus out · / search · f Filter · c Create · d Delete
    · b Backlog · y Yank

Flag AUS, 80 Spalten — unveraendert:

    tab focus in · shift+tab focus out · / search · f Filter · s Status · c Create
    · d Delete · e Edit · b Backlog · t Tags · y Yank · a Parent · r Blocking

Golden `browse_boxform.golden` regeneriert: Footer ohne s/e/t/a/r.
Die Flag-AUS-Goldens (tree/backlog/browse_flat/chrome) zeigen KEINE
Footer-Aenderung — nur die von bt-oqsv entfernte Leerzeile.

## Deviations

Keine.
