---
uid: 35ff28b5-4577-40af-8da8-751a122abf32
---
# Glossar — beans-tui

Gemeinsame Sprache für PO und Agenten. **Prosa-Begriff und Code-Bezeichner stehen
nebeneinander**, damit nicht zwei Vokabulare nebeneinanderher laufen.

Wer ein neues UI-Element baut, trägt es hier nach. Weicht ein Code-Bezeichner vom
gesprochenen Begriff ab, wird das hier vermerkt statt stillschweigend umbenannt.

## Box-Darstellung (jira-Style)

| Begriff          | Bedeutung                                                                                                                | Code                                       |
| ---------------- | ------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------ |
| **boxed field**  | Ein Feld, jira-artig als Box dargestellt. **Box-Titel = Feld-Titel**, **Box-Badge = Keybind**.                           | `dropdownBox()`                            |
| **Box-Titel**    | Das Feld-Label, im Rahmen sitzend (heute oben).                                                                          | `boxTopBorder(label, …)`                   |
| **Box-Badge**    | Der Keybind des Feldes, im Rahmen sitzend (heute unten, in Klammern: `(s)`).                                             | `boxBottomBorder(hotkey, …)`               |
| **Box-Form**     | Die gesamte Detail-Ansicht aus boxed fields: Title / Status\|Type\|Priority / Parent\|Tags / Body / Relations / History. | `detailBoxForm()`, Flag `BT_BOXFORM`       |
| **Panel**        | Ein mehrzeiliges boxed field (Body, Relations, History). Gleiche Anatomie, nur hoch.                                     | `panelBox()`                               |
| **Filter-Strip** | Die persistente Zeile boxed fields oben (Type/Status/Priority/Tags).                                                     | `filterBar()` — Bezeichner weicht ab, s.u. |
| **Zelle**        | Ein boxed field innerhalb einer mehrspaltigen Zeile.                                                                     | `scalarCell`, angeordnet von `gridRow()`   |
| **Inline-Dropdown** (Anker-Popup) | Das Wertmenü eines endlichen Feldes (Status/Type/Priority), das **feld-verankert** direkt unter/über der Box aufklappt statt zentriert — maus-nativ wählbar. D09 REVIDIERT. | `placeValueMenuOverlay`, `boxFormFieldRect`, `placeOverlayAt` |

### Anmerkungen
- **Die Badge-Position ist nicht Teil der Definition.** Für den Body wandert sie in den
  oberen Rahmen (bean `bt-oox1`), weil der untere bei langem Body wegscrollt. Es bleibt
  ein Box-Badge.
- **`filterBar` vs. „Filter-Strip":** der PO sagt Filter-Strip, der Code heißt `filterBar`.
  Bewusst nicht umbenannt — eine Umbenennung quer durch Renderer, Tests und Golden wäre
  Risiko ohne Nutzen. Hier vermerkt, damit niemand zwei verschiedene Dinge vermutet.
- **„Chip"** wurde früher für die Filter-Felder benutzt. **Aufgegeben** — es sind boxed
  fields wie alle anderen. In `box_filter_bar.go` steht der Begriff noch im Kopfkommentar.

## Views

**Top-Level-Views** — je eine Konstante in `viewID` (`internal/tui/types.go`). Das sind die
Bildschirme, zwischen denen umgeschaltet wird.

| Name               | Datei                    | Kommentar                                                                                                      |
| ------------------ | ------------------------ | -------------------------------------------------------------------------------------------------------------- |
| **Lobby**          | `view_lobby.go`          | Repo-Picker mit ASCII-Logo. Entfällt beim Start in einem einzelnen Repo. Konstante `viewLobby`.                |
| **Browse**         | `view_browse_repo.go`    | Der Primat-View: Master-Detail, links Tree, rechts Detail. Konstante `viewBrowseRepo`.                         |
| **Backlog**        | `view_browse_backlog.go` | Flache Liste parentloser/ready beans, ebenfalls Master-Detail, sortierbar. Taste `b`. Konstante `viewBacklog`. |
| **Tag-Management** | `view_tag_management.go` | Zentrale Tag-Verwaltung. Konstante `viewTagManagement`.                                                        |

**Darstellungen innerhalb eines Views** — eigene Dateien, aber **keine** eigene `viewID`.
Sie sind Zustände von Browse, nicht Geschwister davon:

| Name | Datei | Kommentar |
|---|---|---|
| **Flat** | `view_browse_flat.go` | Die linke Pane als flache Liste statt als Tree. Umschalter `G`, Zustand `flatView`. |
| **Vollbild** | `view_fullscreen.go` | Detail über die ganze Breite. Taste `v`, Zustand `fullscreenDetail`. |
| **Detail-Inhalt** | `view_detail_bean.go` | Was in der Detail-Pane steht — als Accordion oder, mit `BT_BOXFORM`, als Box-Form. |

### Achtung: die Spec-Tabelle ist veraltet
`docs/plans/v1-port/design-spec.md` §6 führt die Views als V1–V5 mit Bezeichnern, die es
**im Code nicht gibt**: `viewBrowseProject`, `viewBrowseBacklog`, `viewDetailIssue`,
`viewCommandCenter` und `viewHome` existieren allesamt nicht (`viewHome` nur als
Kommentar-Verweis auf die devd-Vorlage). Ebenso verweisen Kommentare in
`view_browse_repo.go` auf `renderReviewDetailPane` in `view_review_cockpit.go` — **beides
existiert nicht**. Beim Lesen der Spec also die Namen hier gegenprüfen; die V-Nummern
(V1 Lobby, V2 Browse, V3 Backlog, V4 Detail-Accordion, V5 Command-Center) sind weiterhin
brauchbar als Referenz, die Bezeichner nicht.

## Layout

| Begriff | Bedeutung | Code |
|---|---|---|
| **Master-Detail** | Die geteilte Ansicht: Liste links, Detail rechts. | `viewBrowseRepo` |
| **Pane** | Eine der beiden Hälften. „linke Pane" = Liste, „Detail-Pane" = rechts. | `renderPane`, `clickPaneGeometry` |
| **Tree** | Die hierarchische Liste links (Milestone → Epic → Blatt). | `treeRows`, `treeRowText` |
| **Accordion** | Die klassische, aufklappbare Detail-Darstellung — das, was die Box-Form ablöst, wenn das Flag an ist. | `renderAccordionPane` |
| **Region** | Fokussierbarer Bereich: Tree · Detail · Filter-Strip. `tab` bewegt INNERHALB, `esc` verlässt. | — |

## Overlays

| Begriff | Bedeutung | Code |
|---|---|---|
| **Overlay** | Ein modaler Aufsatz über der Hauptansicht. | `overlay*`-Konstanten |
| **Picker** | Overlay zur Auswahl eines Bezugs: Relations, Parent, Tags. | `box_picker_*.go` |
| **Value-Menü** | Overlay zur Auswahl eines Skalarwerts: Status, Type, Priority. | `box_menu_value.go` |
| **Facetten-Overlay** | Das Filter-Overlay hinter `f`. | `box_filter_facets.go` |
| **Command-Palette** | Der globale Befehlsaufruf, Taste `K`. | `overlay_palette.go` |
| **Toast** | Kurze Einblendung für Rückmeldungen; hat die frühere reservierte Statuszeile abgelöst. | — |

## Layout und Scrollen

- **Hängender Einzug** — Umbruch einer Listenzeile, bei dem die Folgezeilen auf der
  Spalte des Titels beginnen statt am Zeilenanfang.
- **Zeilen-Budget** — eine Höhenbegrenzung, die gerenderte Terminal**zeilen** zählt; das
  Gegenstück ist der **Element-Cap**, der Listen**einträge** zählt. Nur das Zeilen-Budget
  begrenzt Höhe verlässlich — sobald Titel umbrechen, erzeugt ein Eintrag mehrere Zeilen.
- **Scroll-Mitnahme** — Regel, dass der Viewport dem Fokus folgt: ein fokussiertes Feld
  darf nie außerhalb des sichtbaren Bereichs liegen. Beim Wandern über ein Feld, das höher
  als der Ausschnitt ist, wird es erst vollständig gezeigt und dann weitergesprungen
  (*reveal-then-move*).

## Arbeitsweise

| Begriff | Bedeutung |
|---|---|
| **Golden** | Ein festgeschriebener Render-Schnappschuss in `internal/tui/testdata/*.golden`. Regeneriert mit `-update`, danach Zeile für Zeile per `git diff` prüfen. |
| **Smoke** | Ein echter Lauf in tmux bei **80 Spalten** gegen ein reales beans-Repo. Pflicht bei Footer- und Umbruch-Änderungen — Unit-Tests bei 100/120 Spalten sehen Umbruch-Bugs strukturell nicht. |
| **Prelude** | Ein Befund, der per `--body-append` in das bean der nächsten passenden Aufgabe wandert, statt im Chat zu verdunsten. Der Folge-Agent erledigt ihn zuerst. |
| **Grounding** | Read-only-Vorarbeit, die Fundorte und Fallstricke in ein bean schreibt, bevor jemand implementiert. |

## Tastatur-Konventionen (D07)

- **klein = Feld-Aktion:** `s` Status · `o` Type · `u` Priority · `a` Parent · `r` Relations ·
  `t` Tags · `e` Body
- **groß = View/Global:** `S` Sort · `X` Filter zurücksetzen · `K` Palette · `G` Nested/Flat
- **`tab`/`shift+tab`** bewegen innerhalb der fokussierten Region, **`esc`** verlässt sie
- **`enter`** öffnet — überall. **`space`** schaltet einen Wert in einer offenen
  Mehrfachauswahl um.
- **`pgup`/`pgdn`** blättern im Body.

### Terminal-Fallen, die Bindungen ausschließen
- **`ctrl+i` ist `tab`** (beides 0x09) — nicht unterscheidbar, als eigene Bindung tot.
- **`ctrl+s`/`ctrl+q`** sind XOFF/XON — durch `TestKeymapNoCtrlSQ` hart verboten.
- **Buchstaben-Hotkeys sterben neben Eingabefeldern.** Wo getippt wird, braucht es
  Ctrl-Akkorde (so gelöst im Picker-Filter: `^t`/`^n`/`^p`/`^g`).
