---
# bt-oqsv
title: Leere reservierte Footer-Zeile entfernen (Toast ersetzt sie)
status: todo
type: task
priority: normal
created_at: 2026-07-20T07:26:22Z
updated_at: 2026-07-20T07:49:33Z
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
