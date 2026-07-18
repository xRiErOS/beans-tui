# TUI-Spec-Extraktion — Design-Spezifikation

- **Epic:** bt-ctox · **Meilenstein:** bt-3cpw
- **Status:** Design (Review-Gate offen)
- **Senke:** dev-wiki-Bundle (OKF Knowledge Catalogue), Cluster `tui-architecture-design`
- **Quelle:** beans-tui (`internal/`, `docs/plans/v1-port/design-spec.md`, `docs/LESSONS-LEARNED.md`)

## 1. Zweck

Die beans-tui ist die reifste TUI dieser Codebasis-Lineage. Ihre bewährten Prinzipien
sollen so exakt dokumentiert werden, dass **andere TUI-Projekte sie replizieren** können —
framework-agnostisch verstehbar und in Bubbletea/Go konkret nachbaubar.

## 2. Grounding-Befund (maßgeblich für den Ansatz)

Das dev-wiki ist **kein leeres Blatt**. Der Cluster `tui-architecture-design` enthält bereits
**32 volle Konzepte**, generalisiert aus der devd-cli-go-TUI — der Lineage, aus der beans-tui
portiert wurde. beans-tui ist die **reifere Iteration** derselben Familie.

Daraus folgt: Die Arbeit ist **Delta-Extraktion**, nicht greenfield. Bestehende Konzepte werden
gegen die reifere beans-tui geschärft/korrigiert; identifizierte Lücken werden gefüllt.

**Lücken (weiße Flecken) im Bestand:**
- Theme-/Color-Token-System (kein dediziertes Konzept; nur `color-profile-and-dark-bg`)
- Accordion als eigenständige Komponente (nur implizit in `two-level-focus-machine`)
- Command-Palette als Muster (nur als Screen-Mockup erwähnt)
- Dünn: Testing, Performance/Rendering, a11y, i18n/Umlaut-Breiten

**Template-Referenz** für Spec-Aufbau: `vik-okf/architecture/design-spec` (15 strukturierte Sektionen).

## 3. Grundsatz-Entscheidungen

| Code | Entscheidung |
|------|--------------|
| D01 | **Senke: dev-wiki erweitern.** Schreiben ausschließlich via `/okf` (nie Hand-.md). |
| D02 | **Zweischichtig.** Jede Konzept-Datei: (1) agnostisches Prinzip + (2) Bubbletea/Go-Referenz. |
| D03 | **Tiefe: Prinzip + Referenz-Snippet.** Muster/Do-Don't/Rationale + `datei:zeile`-Verankerung mit kurzem Snippet. |
| D04 | **Delta-Modus: Hybrid.** Bestehende Konzepte in-place aktualisieren (beans-tui als Referenz-Quelle) + 3 Lücken additiv anlegen. Ein Cluster, sauber evolviert. |
| D05 | **Tempo: alle Slices parallel**, ein PO-Review am Ende vor `/okf`-Ingest. |

## 4. Zweischichtige Konzept-Struktur (Vorlage je Konzept)

```
# <Konzept-Titel>

## Prinzip (agnostisch)
- Muster: <was>
- Warum: <Entscheidungsrationale>
- Do / Don't: <Leitplanken>
- Portierung: <Hinweis ratatui/textual, wo Bubbletea-spezifisch>

## Referenz (Bubbletea/Go)
- Quelle: `internal/.../datei.go:zeile`
- Snippet: <kurz, das Muster tragend>
- Fallstricke: <LESSONS-LEARNED-Bezug, falls vorhanden>
```

## 5. Vertical Slices (13)

Jede Slice ist unabhängig extrahierbar. `Delta-Typ`: `schärfen` = bestehendes Konzept
in-place aktualisieren · `NEU` = additives Konzept für eine Lücke · `synthese` = Cluster-Meta.

| # | Slice | Quell-Dateien (beans-tui) | dev-wiki-Baseline | Delta-Typ |
|---|-------|---------------------------|-------------------|-----------|
| 1 | App-Shell & Elm-Loop | `tui/{app,types,update,messages,view}.go` | value-receiver-model, message-command-separation, error-check-before-mutation | schärfen |
| 2 | Chrome: Header/Footer/Breadcrumb | `tui/{view,footer_context}.go` | layout-primitives-chrome, footer-status-vs-warning, ansi-safe-width | schärfen |
| 3 | Theme-/Color-Token-System | `theme/{theme,icons}.go`, design-spec §8 | — (nur color-profile-and-dark-bg) | **NEU** |
| 4 | Master-Detail + Tree + Accordion | `tui/{view_browse_repo,view_browse_backlog,accordion,view_detail_bean,list}.go` | masterdetail-responsive, two-level-focus | schärfen + **Accordion NEU** |
| 5 | Overlay/Modal-Primitive | `tui/{overlay,modal,box_confirm_*}.go` | overlay-compositing, modal-overlay-layer, transient-overlays | schärfen |
| 6 | Navigation & Fokus-Modell + Vollbild `v` | `tui/view_fullscreen.go`, design-spec §3.3 | two-level-focus, input-routing-precedence | schärfen |
| 7 | Mutation: Forms/Picker/Editor/Menu | `tui/{forms_shared,form_*,box_picker_*,box_menu_value,editor}.go` | huh-form-as-submodel, huh-form-modal-archetype, external-editor-suspend | schärfen |
| 8 | Keymap & Help-Generierung | `tui/{keymap,overlay_shortcuts}.go`, design-spec §7 | central-keymap-single-source | schärfen |
| 9 | Datenlayer via CLI-Subprocess | `data/{bean,client,index,mutations,watcher,discover,repo_slug}.go` | lazy-cache-optimistic-merge, message-command-separation | schärfen |
| 10 | Notifications/Toast | `tui/overlay_show_toast.go` | notification-toast-overlay | schärfen |
| 11 | Search/Filter + Command-Palette | `tui/{fuzzy,search_prefix,box_filter_facets,overlay_palette}.go` | client-side-filter-search | schärfen + **Palette NEU** |
| 12 | Config & Yank/Clipboard | `config/{settings,state}.go`, `tui/box_form_settings.go`, `clip/clip.go`, `tui/context.go` | config-and-state-persistence, clipboard-osc52 | schärfen |
| 14 | Visuelle Prüfung: VHS + tmux | `*.tape`, `Makefile`, `*_golden_test.go`, `docs/screenshots/` | hybrid-test-strategy, build-tooling-pitfalls | schärfen + **VHS/tmux NEU** |
| 15 | In-Action: PO-Feedback- & Evolutions-Mining | git-Historie, `.beans/` (PO-Feedback-Runden, Bug-Beans), LESSONS-LEARNED | design-guidelines-lessons-learned | mutiert in andere Slices |
| 13 | Synthese: Cluster-Index + Querschnitts-Leitplanken (zuletzt) | `docs/LESSONS-LEARNED.md`, Cluster-`index.md` | tui-specification, index | synthese |

**Optional (projektspezifisch, nur bei Bedarf):** Tag-Management & Review-Flow
(`data/{tagdefs,tags,hierarchy,review}.go`, `tui/view_tag_management.go`) — an beans-Domänenmodell
gekoppelt, weniger generisch replizierbar. Nicht Teil des Kern-Scopes.

## 6. Querschnitts-Leitplanken (Slice 13, aus LESSONS-LEARNED)

Nicht als eigene Slice-Konzepte, sondern als Extraktions-Leitplanken, die in mehreren Konzepten auftauchen:

- ANSI-Styles nesten nicht → Pre-Style-Konvention, kein äußerer Wash.
- Auto-Wrap im echten TrueColor-TTY unzuverlässig → explizite `\n` + Zeichenbudget.
- Neuer Modus → vollständige Exit-Pfad-Inventur + FitsOuterFrame-Tests.
- Positions-/Kontext-Info strukturell durchreichen, nie aus gerendertem Text zurückparsen.
- Geteilte Ergebnis-Handler tragen Kontext explizit, nie aus Seiteneffekt lesen.
- Golden-File + Behavioral-Tests je Bruch-Szenario.

## 7. Ausführungs-/Orchestrierungs-Modell

1. **15 Subagenten parallel** (D05, alle Opus — starkes Reasoning für Delta-Entscheidungen), einer je Slice. Jeder Subagent:
   - liest bestehende dev-wiki-Konzepte der Slice (Baseline),
   - liest die beans-tui-Quell-Dateien der Slice,
   - erzeugt OKF-fertige Delta-Drafts nach Vorlage §4 in `docs/tui-spec-okf-inbox/`.
2. **Slice 15 (In-Action)** erzeugt eine Mutations-Landkarte: jede Nutzungs-Erkenntnis → Ziel-Slice/Konzept.
   Diese wird beim Review in die betroffenen Slice-Drafts eingearbeitet (Slice 15 editiert nicht selbst, da parallel).
3. **Ein PO-Review** aller Drafts am Ende (D05).
4. **`/okf` ingest** der freigegebenen Drafts ins dev-wiki-Bundle — Hybrid: `migrate`/Update für
   bestehende Konzepte, `ingest` für die neuen Lücken-Konzepte (D01, D04).
5. **Cluster-Index/Landkarte** (Slice 13) zuletzt, nachdem alle Konzepte stehen.

**Subagent-Output-Format (Pflicht im Dispatch):** je Konzept eine Draft-Datei nach Vorlage §4;
Rückgabe = Markdown-Tabelle `konzept | pfad | delta-typ | baseline-referenz | verankerte-quellen(datei:zeile)`.

## 8. Definition of Done

- [ ] Alle 15 Slices als Drafts in `docs/tui-spec-okf-inbox/` erzeugt.
- [ ] Jedes Konzept zweischichtig (Prinzip + verankerte Bubbletea/Go-Referenz mit `datei:zeile`).
- [ ] Die Lücken (Theme-Token, Accordion, Command-Palette, VHS/tmux-Visualprüfung) als neue Konzepte abgedeckt.
- [ ] In-Action-Erkenntnisse (Slice 15) in die betroffenen Slice-Konzepte eingearbeitet.
- [ ] PO-Review bestanden.
- [ ] `/okf`-Ingest ins dev-wiki abgeschlossen; Cluster-Index aktualisiert.
- [ ] Findbar via `okf-cli --root ~/Obsidian/Knowledge-Catalogue find "<query>"`.
