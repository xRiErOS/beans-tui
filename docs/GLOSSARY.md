# Glossar — beans-tui

Gemeinsame Sprache für PO und Agenten. **Prosa-Begriff und Code-Bezeichner stehen
nebeneinander**, damit nicht zwei Vokabulare nebeneinanderher laufen.

Wer ein neues UI-Element baut, trägt es hier nach. Weicht ein Code-Bezeichner vom
gesprochenen Begriff ab, wird das hier vermerkt statt stillschweigend umbenannt.

## Box-Darstellung (jira-Style)

| Begriff | Bedeutung | Code |
|---|---|---|
| **boxed field** | Ein Feld, jira-artig als Box dargestellt. **Box-Titel = Feld-Titel**, **Box-Badge = Keybind**. | `dropdownBox()` |
| **Box-Titel** | Das Feld-Label, im Rahmen sitzend (heute oben). | `boxTopBorder(label, …)` |
| **Box-Badge** | Der Keybind des Feldes, im Rahmen sitzend (heute unten, in Klammern: `(s)`). | `boxBottomBorder(hotkey, …)` |
| **Box-Form** | Die gesamte Detail-Ansicht aus boxed fields: Title / Status\|Type\|Priority / Parent\|Tags / Body / Relations / History. | `detailBoxForm()`, Flag `BT_BOXFORM` |
| **Panel** | Ein mehrzeiliges boxed field (Body, Relations, History). Gleiche Anatomie, nur hoch. | `panelBox()` |
| **Filter-Strip** | Die persistente Zeile boxed fields oben (Type/Status/Priority/Tags). | `filterBar()` — Bezeichner weicht ab, s.u. |
| **Zelle** | Ein boxed field innerhalb einer mehrspaltigen Zeile. | `scalarCell`, angeordnet von `gridRow()` |

### Anmerkungen
- **Die Badge-Position ist nicht Teil der Definition.** Für den Body wandert sie in den
  oberen Rahmen (bean `bt-oox1`), weil der untere bei langem Body wegscrollt. Es bleibt
  ein Box-Badge.
- **`filterBar` vs. „Filter-Strip":** der PO sagt Filter-Strip, der Code heißt `filterBar`.
  Bewusst nicht umbenannt — eine Umbenennung quer durch Renderer, Tests und Golden wäre
  Risiko ohne Nutzen. Hier vermerkt, damit niemand zwei verschiedene Dinge vermutet.
- **„Chip"** wurde früher für die Filter-Felder benutzt. **Aufgegeben** — es sind boxed
  fields wie alle anderen. In `box_filter_bar.go` steht der Begriff noch im Kopfkommentar.

## Layout und Ansichten

| Begriff | Bedeutung | Code |
|---|---|---|
| **Master-Detail** | Die geteilte Ansicht: Liste links, Detail rechts. | `viewBrowseRepo` |
| **Pane** | Eine der beiden Hälften. „linke Pane" = Liste, „Detail-Pane" = rechts. | `renderPane`, `clickPaneGeometry` |
| **Tree** | Die hierarchische Liste links (Milestone → Epic → Blatt). | `treeRows`, `treeRowText` |
| **Flat** | Die flache Liste als Alternative zum Tree, Umschalter `G`. | `flatView`, `view_browse_flat.go` |
| **Vollbild** | Detail über die ganze Breite, Taste `v`. | `fullscreenDetail` |
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
