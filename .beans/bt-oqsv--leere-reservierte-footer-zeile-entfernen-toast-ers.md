---
# bt-oqsv
title: Leere reservierte Footer-Zeile entfernen (Toast ersetzt sie)
status: completed
type: task
priority: normal
created_at: 2026-07-20T07:26:22Z
updated_at: 2026-07-20T08:14:09Z
parent: bt-vy1q
---

**Nebenbefund N6 (PO, 2026-07-20).** Der Footer haelt eine leere Zeile vor — historisch reservierter Platz fuer Notifications. Das Toast-System (`overlay_show_toast.go`) uebernimmt das laengst, die Reservierung ist tot und kostet eine Bildschirmzeile.

Beleg: in `internal/tui/testdata/tree.golden` und `browse_boxform.golden` ist die vorletzte Zeile innerhalb des Rahmens leer.

## Ziel
Die reservierte Leerzeile entfernen, die Zeile geht an den Body/die Panes zurueck.

## Betroffen
- `internal/tui/view.go` — Footer-/Chrome-Komposition (wo die Zeile reserviert wird)
- ggf. die Hoehen-Rechnung (`bodyH`), damit der Rahmen NICHT waechst, sondern die Panes eine Zeile mehr bekommen

## Vorsicht
- Beruehrt praktisch ALLE Goldens (`tree`, `chrome`, `backlog`, `browse_boxform`, `browse_flat`) → bewusst regenerieren und im Commit begruenden
- Pruefen, dass Toasts weiterhin korrekt ueberlagern und nichts verdecken/springt
- Vorher verifizieren, dass die Zeile wirklich Notification-Reservierung ist und nicht Wrap-Reserve fuer den zweizeiligen Footer bei schmalen Terminals (bei 80 Spalten ist der Footer 2-3 Zeilen!)

## Akzeptanz
- [ ] Ursprung der Leerzeile im Code belegt (Datei:Zeile) vor der Aenderung
- [ ] Zeile entfernt, Panes gewinnen eine Zeile, Rahmen bleibt gleich hoch
- [ ] Toast-Verhalten unveraendert (Test/Smoke)
- [ ] 80-Spalten-Smoke: mehrzeiliger Footer bricht nichts
- [ ] Goldens regeneriert, voller `command go test ./...` gruen


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

Commit 3e73363 `refactor(footer): drop the reserved blank status line`
(+ Golden-Commit 2f6fe9c).

**Ursprung belegt** (Akzeptanz 1): `statusBar()` (view.go) liefert `""`, wenn
kein Indicator anliegt; die vier View-Kompositionen haengten `+ "\n" + status`
UNBEDINGT an (view_browse_repo.go, view_browse_backlog.go,
view_tag_management.go, `Chrome()` in view.go). Es ist **keine** Wrap-Reserve
fuer den mehrzeiligen Footer — `footer()` wrappt in `localKeys` selbst, dessen
Hoehe ueber `lipgloss.Height(localKeys)` separat verrechnet wird (im
80-Spalten-Smoke belegt: 2-zeiliger Footer + separat die Leerzeile).

Aenderung:
- `m.statusLine(width)` in view.go ist jetzt die EINE Quelle des Indicators
  (vorher dreimal derselbe `if m.watchUnavailable`-Block).
- `appendStatusLine()` haengt die Zeile nur an, wenn sie etwas traegt.
- `clickPaneGeometry` nimmt `status` als Parameter:
  `footH := Height(localKeys) + 1 + statusLineHeight(status)`.

## Entscheidungen

- **Status-Zeile bleibt, nur die RESERVIERUNG faellt.** Der
  watch-unavailable-Indicator braucht weiter einen Platz; ihn in die
  Footer-Zeile zu mischen haette den rechtsbuendigen Optik-Slot zerstoert.
- **`status` wird Parameter von `clickPaneGeometry`** statt intern geraten —
  das erzwingt an jeder Aufrufstelle eine explizite Entscheidung und macht
  den Grounding-Fallstrick unmoeglich zu vergessen (Compiler-Fehler statt
  stiller Offset).
- **`Chrome()` loest ein Henne-Ei-Problem mit 2-Pass**: `scrollView` erzeugt
  den Indicator, braucht dafuer aber `avail`, das wissen muss, ob der
  Indicator eine Zeile belegt. Pass 1 rechnet ohne Status-Zeile; ist eine
  noetig, re-fenstert Pass 2 eine Zeile kuerzer. Terminiert beweisbar: ein
  kleineres Fenster kann Ueberlauf nur erhoehen, nie aufheben — ein nicht-
  leerer Indicator kann im zweiten Pass nicht leer werden.

## Test-Output

Neue Tests in internal/tui/footer_status_line_test.go:

    --- PASS: TestBrowseRepoFooterHasNoReservedBlankLine
    --- PASS: TestBrowseRepoWatchIndicatorKeepsItsOwnLine
    --- PASS: TestBrowseRepoPanesGainTheFreedLine
    --- PASS: TestClickPaneGeometryBodyHMatchesRenderedPane
    --- PASS: TestTreeClickRowSurvivesStatusLineRemoval

`TestClickPaneGeometryMultiLineFooterHeightPin` (mouse_test.go) pinnt jetzt
BEIDE Zustaende auf Literale: ohne Status-Zeile bodyH=30/footerY=36, mit
Status-Zeile bodyH=29/footerY=35 (die Vor-bt-oqsv-Zahlen).

**Mutationsprobe** (Nachweis, dass der Regressionstest den Fallstrick wirklich
faengt): `footH` testweise auf das alte `Height(localKeys) + 2`
zurueckgedreht ->

    --- FAIL: TestClickPaneGeometryBodyHMatchesRenderedPane
        footer_status_line_test.go:120: watchUnavailable=false: frame height
        = 29, want 30 -- footH no longer matches the composed footer

Erster Entwurf des Tests (kurze Liste + reine Klick-Zuordnung) hat die
Mutation NICHT gefangen — deshalb zusaetzlich die Rahmenhoehen-Assertion und
ein 60-Beans-Fixture (`longTreeModel`), das laenger als die Pane ist und damit
`windowStart(bodyH)` echt durchlaeuft.

Voller Lauf (uncached, nach allen drei Commits):

    ok  github.com/xRiErOS/beans-tui/cmd            0.441s
    ok  github.com/xRiErOS/beans-tui/internal/config 1.121s
    ok  github.com/xRiErOS/beans-tui/internal/data   4.700s
    ok  github.com/xRiErOS/beans-tui/internal/theme  0.781s
    ok  github.com/xRiErOS/beans-tui/internal/tui  151.398s

**80-Spalten-tmux-Smoke** (Akzeptanz 4), beide Flag-Zustaende: Rahmen fuellt
24 Zeilen, mehrzeiliger Footer bricht nichts, keine Leerzeile mehr ueber dem
unteren Rahmen.

Goldens (Akzeptanz 5): tree/backlog/browse_flat/chrome — je eine Leerzeile
raus, je eine Pane-/Body-Zeile rein, Rahmenhoehe konstant.

## Deviations

Keine. Toast-Verhalten unveraendert (`renderToast` legt sich unabhaengig von
der Footer-Komposition ueber die fertige View; alle Toast-Tests gruen).
