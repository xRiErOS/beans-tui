# Experiment: jira-Style Flat/Salient UI für beans-tui

Status: DESIGN (genehmigt) · Branch: `experiment/jira-style-ui` · Datum: 2026-07-19

## 1. Warum / Frage des Experiments

Tipp von außen: die [jira-TUI](https://jiratui.sh) wird gut angenommen, weil sie
**nicht verschachtelt** wirkt. Referenz-Screenshot: `app-homepage.png` (jiratui.sh
gallery). Das Experiment prüft die eine Frage:

> Ist eine flachere, salientere Darstellung im jira-Stil eine **Verbesserung** für
> beans-tui gegenüber dem heutigen doppelt-verschachtelten Modell?

Es ist ein **Spike** auf eigenem Branch — nicht garantiert zum Merge. Erfolgskriterium:
das neue Modell fühlt sich im tmux-Smoke (80 + 100 Spalten) klarer/schneller an als der
Ist-Zustand, ohne Funktionsverlust.

## 2. Ausgangslage (Ist) vs. jira

**beans-tui heute = doppelt verschachtelt:**
- Links: Baum-Accordion (Milestone ▾ Epic ▾ Task, eingerückt, klappbar)
- Rechts: Detail mit Accordion-Sektionen `[1] META [2] BODY [3] RELATIONS [4] HISTORY`
- Aktionen nur im Footer

**jira-TUI = flach + salient:**
- Oben: persistente Filter-Leiste (Dropdowns)
- Links: flache Tabelle, eine Ebene
- Rechts: Detail immer voll sichtbar, Felder als Boxen, Hotkeys direkt am Element
- Farbe zur Salienz-Hebung

**Was der PO explizit gut findet:** Shortcuts direkt & salient am Element · die
Filter-Leiste oben · Felder als Dropdowns (maus-zugänglich) · Farbeinsatz zur Salienz.
**Was bleibt:** der Baum ist ok — kein Zwangs-Ersatz, sondern umschaltbar (siehe D05).

## 3. Entscheidungen (verriegelt)

| Code | Entscheidung |
|------|-------------|
| D01 | **Ein wiederverwendbares Dropdown-Widget**: Label oben im Rahmen, `▾`-Affordance innen, **Hotkey-Badge unten-rechts im Rahmen**. Filter-Leiste UND Detail nutzen dasselbe Widget → UI-Konsistenz. |
| D02 | **Filter-Leiste oben persistent**, Dropdown-Look konsistent zum Detail. |
| D03 | Detail-Scalars als Dropdowns. **Type + Priority werden editierbar** (heute nicht) — neue Picker, neue Keys `o` / `u`. |
| D04 | RELATIONS/HISTORY als **gestapelte Boxen** in einer scrollenden Pane — kein Accordion mehr im Detail. |
| D05 | Links **View-Switcher `Nested ▏ Flat`** (Key `G`, klickbar). Flat-Default = Hierarchie-geordnet `M > E/F > T/B`. Eigener Sort (`S`) → Hierarchie fällt weg (flache Sortier-Liste). |
| D06 | **Maus**: Klick öffnet Dropdown/Toggle. Infra vorhanden (`app.go:84` MouseCellMotion, `mouse.go` `detailClickRow`/`detailClickKey`, Double-Click). |
| D07 | **Case-Konvention**: **klein** = Feld-Aktion (`s t a r e o u c d y`), **groß** = View/Global (`S X K G`). Zwei intern konsistente Gruppen, kein Mix innerhalb einer Gruppe. |
| D08 | **Filter über `f`-Einstieg**, KEINE Facetten-Einzelkeys (hält D07, spart Buchstaben; Facetten via Tab/Pfeile/Maus). |
| D09 | **huh ersetzen durch Inline-Box-Editing**: Detail *ist* das Edit-Formular, Create = leere Boxen gleicher Layout. huh-Forms + die langsamen huh-Drive-Tests fallen weg. Eigene Inline-Popups (jira-treu). **Synergie:** eigene Popups sind maus-nativ → löst den huh-MouseMsg-Blocker (D06) ohne Workaround. |

## 4. Farb-Salienz-Karte (D08-Farbe) — nur `internal/theme`-Tokens

Catppuccin Macchiato, TrueColor. Keine Hex-Literale in Views.

| Element | Token | Effekt |
|---------|-------|--------|
| Dropdown-Rahmen, unfokussiert | `Overlay` #8087a2 | ruhig (bestehender Feld-Border) |
| Dropdown-Rahmen, **fokussiert** | `Mauve`/`Accent` | aktives Feld poppt |
| Feld-Label (oben im Rahmen) | `Subtext` / `Hint` | muted |
| **Hotkey-Badge** (unten im Rahmen) | `BindingKey` = `Teal` | salienter Key |
| `▾` Dropdown-Affordance | `Chevron` = `Peach` | Interaktions-Marker |
| Status-Wert | `StatusColor` (todo=Green, in-progress=Yellow, draft=Blue, completed/scrapped=Subtext) | bestehend |
| Priority-Wert | `priorityColor` (critical=Red, high=Yellow, normal=Text, low/deferred=Subtext) | bestehend |
| Type-Wert | `TypeStyle` (Blue/Mauve/Sky/Red) | bestehend |
| **Aktiver Filter-Chip** (Wert gesetzt) | Label/Rahmen `Peach` | hebt sich von leeren `(any)` ab |
| Leerer Filter-Chip `(any)` | `Hint` | zurückgenommen |
| View-Toggle aktives Segment | `Select` #fe640b (invers) | laut |
| Titel/Body `✎` + `(e)` | `Teal` | Edit-Affordance |

## 5. Keymap-Änderungen (Single Source: `internal/tui/keymap.go`)

| Key | Aktion | Neu/geändert |
|-----|--------|--------------|
| `o` | Type-Picker (Feld-Edit) | NEU |
| `u` | Priority-Picker (Feld-Edit) | NEU |
| `G` | View-Toggle Nested ▏ Flat | NEU |
| `f` | fokussiert die persistente Filter-Leiste (statt Overlay) | GEÄNDERT |
| `s t a r e` | Status/Tags/Parent/Blocking/Body — unverändert, jetzt als Feld-Badges sichtbar | Darstellung |
| `S X K` | Sort/Clear/Palette — unverändert (View/Global-Gruppe) | — |

Help-Overlay generiert weiterhin aus der Keymap (kein Drift).

### 5.1 Nachtrag zu Entscheidung a3 — Schließen-Alias ist gruppen-gebunden (bean bt-z4w7)

**Revidiert, nicht stillschweigend gepatcht.** Entscheidung a3 (epic-E3-plan.md »(a3)«)
formulierte wörtlich „**esc/`s`** schließt" — korrekt zu ihrer Zeit, denn `s` war der
einzige Key, der dieses Menü öffnen konnte. S4 hat mit `o` (Type) und `u` (Priority)
zwei weitere Öffner ergänzt, ohne den Schließen-Zweig mitzuziehen. Folge: ein per `o`
geöffnetes Menü schloss auf `s`, und beide Hint-Flächen (Footer Zone 3 + der modal-
interne `esc/s:cancel`) nannten `s` unabhängig von der offenen Gruppe.

**Neu:** Es schließt **`esc` plus der Key, der das Menü geöffnet hat** (`s`/`o`/`u`).
Ein fremder Gruppen-Key schließt nicht mehr. `esc` bleibt in allen Gruppen der
universelle Ausstieg.

**Warum das die Ursache trifft und nicht das Symptom:** Der Fehler war nicht „das
Label steht falsch da", sondern dass Label und Handler **zwei getrennte Quellen**
hatten — der Handler matchte `keys.Status`, der Footer listete `keys.Status`, beide
per Hand nebeneinander gepflegt. Jetzt liest jede der drei Flächen (Handler-Match,
Footer Zone 3, Inline-Hint) dieselbe Funktion `valueMenuGroupKey(group)`
(`internal/tui/footer_context.go`). Divergenz ist damit **konstruktiv ausgeschlossen**,
nicht nur aktuell korrigiert.

Dieselbe Klasse traf zwei weitere Stellen, mitbehoben:

| Fläche | Label sagte | Real gebunden | Ursache |
|---|---|---|---|
| Blocking-Picker Toggle | `space/x Toggle facet` | nur `space` | bt-a3a8 (D6) verengte den Handler wegen des neuen Suchfelds, der Footer blieb auf dem geteilten `keys.Toggle` stehen |
| Tag-/Parent-/Blocking-Picker Navigation | `↑/i up`, `↓/k down` | nur `↑`/`↓` | die Picker schalten bewusst auf rohe `tea.KeyUp`/`KeyDown`, damit `i`/`k` im Suchfeld tippbar bleiben |

Beide nutzen jetzt overlay-lokale Bindings (`blockingPickerToggleHint`,
`pickerNavUpHint`/`pickerNavDownHint`), die der jeweilige Handler **selbst** matcht.

**Drift-Guard:** `TestPickerFooterKeysAreReservedNotTyped`
(`internal/tui/footer_binding_source_test.go`) drückt für jedes Picker-Overlay jeden
im Footer beworbenen Ein-Zeichen-Key und lässt den Build scheitern, sobald einer davon
bloß in die Suchzeile getippt wird statt zu handeln. Genau dieser Test hat den dritten
Fall (`↑/i`) überhaupt erst gefunden — er ist die eigentliche Absicherung gegen die
Wiederkehr dieser Fehlerklasse.

## 6. Layout-Mockups (Referenz)

### 6.1 Dropdown-Widget (D01) — Label im Rahmen, Hotkey im Rahmen
```
╭─ Status ───────────────────╮
│ ● todo                 ▾   │
╰──────────────────── (s) ───╯
```

### 6.2 Persistente Filter-Leiste (D02, D08)
```
  ╭─ Type ───────╮  ╭─ Status ─────╮  ╭─ Priority ───╮  ╭─ Tags ───────╮   ╭─ Search ───────────╮
  │ (any)     ▾  │  │ (any)     ▾  │  │ (any)     ▾  │  │ (any)     ▾  │   │ /                  │
  ╰──────────────╯  ╰──────────────╯  ╰──────────────╯  ╰──────────────╯   ╰────────────────────╯
     f fokussiert die Leiste · Tab/Pfeile/Maus zwischen Facetten · gesetzter Filter = Peach
```

### 6.3 Detail-Pane (D03, D04) — Boxen gestapelt, scrollend
```
╭─ Title ───────────────────────────────────────────────────────────── (e) ──╮
│ First golden task                                                        ✎  │
╰──────────────────────────────────────────────────────────────────────────────╯
╭─ Status ─────────╮ ╭─ Type ───────────╮ ╭─ Priority ───────╮
│ ● todo        ▾  │ │ task          ▾  │ │ ·             ▾  │
╰───────────(s)────╯ ╰───────────(o)────╯ ╰───────────(u)────╯
╭─ Parent ─────────────────────╮ ╭─ Tags ───────────────────────╮
│ gld-epic  Golden Epic     ▾  │ │ —                         ▾  │
╰───────────────────────(a)────╯ ╰───────────────────────(t)────╯
╭─ Body ────────────────────────────────────────────────────────────── (e) ──╮
│ First golden task. Acceptance: …                                         ✎  │
╰──────────────────────────────────────────────────────────────────────────────╯
╭─ Relations ─────────────────────────────────────────────────────────────────╮
│ blocks: — · blocked_by: — · parent: gld-epic                                │
╰──────────────────────────────────────────────────────────────────────────────╯
╭─ History ───────────────────────────────────────────────────────────────────╮
│ 2026-07-05 created · 2026-07-14 updated                                     │
╰──────────────────────────────────────────────────────────────────────────────╯
```
(Bei ≤80 Spalten: Scalars 1-up gestapelt statt 3-up — siehe I02.)

### 6.4 Links: View-Switcher Nested ▏ Flat (D05)
```
╭─ View: ‹ Nested ▏ Flat › ────────────────────╮   ╭─ View: ‹ Nested ▏[Flat]› ────────────────────╮
│ ▾ M gld-mlst Golden Milestone                │   │ Key       St     Ty    Summary               │
│   ▾ E gld-epic Golden Epic                   │   │ gld-mlst  todo   M     Golden Milestone       │
│ ▌   T gld-tsk1 First golden task             │   │ gld-epic  todo   E     Golden Epic            │
│ ▾ (orphaned)                                 │   │ gld-tsk1  todo   T     First golden task      │
│     T gld-orph Golden orphan                 │   │ gld-orph  todo   T     Golden orphan          │
╰──────────────────────────────────────────────╯   ╰──────────────────────────────────────────────╯
   Nested = Baum (heute)                              Flat = flache Liste, Default Hierarchie-Ordnung
```

## 7. Risiken / offene Spannungen

| Code | Beschreibung | Umgang |
|------|-------------|--------|
| I01 | **Box-in-Box-in-Box**: App-Rahmen → Pane-Rahmen → Feld-Rahmen = 3 Ebenen, könnte unruhig wirken (jira lässt den App-Rahmen weg). | Prototyp prüfen: ggf. Pane-Rahmen weglassen, wenn Feld-Boxen kommen. |
| I02 | **80-Spalten-Constraint** (Projekt-Regel): rechte Pane ~40 breit → Scalars 3-up passen nicht. | Responsive: 3-up ab ~110, 2-up ab ~90, 1-up bei 80. tmux-Smoke Pflicht. |
| I03 | jira-Feld ist höher (Rahmen+Wert) → ~6 Scalars kosten viel Vertikal. | Kompakte 1-Zeilen-Box (Label im Rahmen löst die separate Label-Zeile). Scroll in Detail-Pane. |
| I04 | D09 ist ein **großer** Umbau (huh raus, eigene Popups, Create neu). | Phasieren im Plan; Prototyp zuerst read-only Darstellung, dann Edit. |

## 8. Aus dem Scope / Später

- In-Picker-Maus war bei huh der Blocker — mit D09 (eigene Popups) von Anfang an nativ; kein separater Phase-2-Workaround nötig.
- Merge auf `main` erst nach PO-Abnahme des Spikes (Erfolgskriterium §1).

## 9. Machbarkeit (verifiziert)

- **Maus** aktiv: `internal/tui/app.go:84` (`tea.WithMouseCellMotion`), `internal/tui/mouse.go` mappt bereits Tree-/Backlog-/Detail-Feld-Klicks + Double-Click.
- **Theme-Tokens** alle vorhanden: `internal/theme/theme.go` (Overlay/Mauve/Teal/Peach/Hint/Select + StatusColor/priorityColor/TypeStyle).
- **Keymap** zentral: `internal/tui/keymap.go` — neue Keys dort, Help generiert daraus.

## Glossar (verbindlich ab 2026-07-20)

Gemeinsame Sprache fuer PO und Agenten. **Prosa-Begriff und Code-Bezeichner stehen
nebeneinander**, damit nicht zwei Vokabulare entstehen. Wer ein neues Element baut, traegt
es hier nach.

| Begriff | Bedeutung | Code |
|---|---|---|
| **boxed field** | Ein Feld, jira-artig als Box dargestellt. **Box-Titel = Feld-Titel**, **Box-Badge = Keybind**. | `dropdownBox()` |
| **Box-Titel** | Das Feld-Label, im Rahmen sitzend (heute oben). | `boxTopBorder(label, …)` |
| **Box-Badge** | Der Keybind des Feldes, im Rahmen sitzend (heute unten, in Klammern: `(s)`). | `boxBottomBorder(hotkey, …)` |
| **Box-Form** | Die gesamte Detail-Ansicht aus boxed fields (Title / Status\|Type\|Priority / Parent\|Tags / Body / Relations / History). | `detailBoxForm()`, Flag `BT_BOXFORM` |
| **Panel** | Ein mehrzeiliges boxed field (Body, Relations, History). Gleiche Anatomie, nur hoch. | `panelBox()` |
| **Filter-Strip** | Die persistente Zeile boxed fields oben (Type/Status/Priority/Tags). | `filterBar()` — **Bezeichner weicht ab**, s.u. |
| **Zelle** | Ein boxed field innerhalb einer mehrspaltigen Zeile. | `scalarCell`, angeordnet von `gridRow()` |
| **Region** | Fokussierbarer Bereich: Tree · Detail · Filter-Strip. `tab` bewegt INNERHALB, `esc` verlaesst. | — |

### Anmerkungen
- **Die Badge-Position ist nicht Teil der Definition.** Fuer den Body wird sie in den OBEREN
  Rahmen wandern (bean bt-oox1), weil der untere bei langem Body wegscrollt. Es bleibt ein
  Box-Badge.
- **`filterBar` vs. „Filter-Strip":** der PO sagt Filter-Strip, der Code heisst `filterBar`.
  Nicht umbenennen, solange kein anderer Grund dafuer spricht — eine reine
  Umbenennung durch den halben Renderer waere Risiko ohne Nutzen. Hier vermerkt, damit
  niemand zwei verschiedene Dinge vermutet.
- **„Chip"** wurde frueher fuer die Filter-Felder benutzt. Aufgegeben — es sind boxed fields
  wie alle anderen. In `box_filter_bar.go` steht der Begriff noch im Kopfkommentar.
