# jira-Style-UI Experiment — STATE (Resume-Anker)

Kompaktierungs-fester Fortschritts-Anker. Bei Resume ZUERST lesen. Ergänzt Spec + Plan (unten), ersetzt sie nicht.

- **Branch:** `experiment/jira-style-ui` (main-direct Repo; Spike, kein Merge-Zwang)
- **Spec (Wahrheit):** `docs/plans/jira-style-experiment/design-spec.md` — Entscheidungen D01–D09, Farbkarte D08, Mockups, Risiken I01–I04
- **Plan:** `docs/plans/jira-style-experiment/implementation-plan.md` — Slice-Roadmap S1–S8 + S1 voll TDD
- **Frage des Spikes:** Ist die flachere/salientere jira-Darstellung eine Verbesserung? Abnahme via VHS-Smoke 80/100 + PO-Urteil.

## Modus (User-Direktive 2026-07-19)
Autonom weiterarbeiten, Entscheidungen selbst treffen, später gemeinsam prüfen/validieren. Persistenz muss Kontext-Kompaktierung überleben → dieser Anker + git.
- **Review-Anpassung (proportional):** statt 2 Review-Subagenten je Mini-Slice → Self-Review je Slice + EIN kombinierter Reviewer-Subagent an Checkpoints (nach mehreren Slices / am Ende). Finaler Code-Review vor „fertig".

## Slice-Status
| Slice | Ziel | Status | Beleg |
|-------|------|--------|-------|
| S1 | `dropdownBox`-Primitiv (Label+Hotkey im Rahmen, ▾, Fokus-Farbe) | 🟢 DONE | Commit `2b08977`, `internal/tui/box_dropdown.go`(+_test), voller Testlauf `ok 149.966s` |
| S2 | **Additiv**: `detailBoxForm(bean,width)` Scalars (Title/Status/Type/Priority/Parent/Tags) via `dropdownBox`, responsive 3/2/1-up + R1 (Label=Subtext). Golden visuell verifiziert (Farben/Salienz/Ausrichtung gut). Kein Live-Wiring. | 🟢 DONE | Commit `f76b6e9`, `box_detail_form.go`(+_test+golden) |
| S2c | Multi-line `panelBox`-Primitiv + `detailBoxForm` um Body/Relations/History erweitert (full-width Panels), `detailBoxForm(idx,b,width)`. Additiv, Golden regeneriert. | 🟢 DONE | Commit `d62edf0`, `box_panel.go`(+_test), `box_detail_form.go` |
| S2b | Live-Wiring per **Env-Flag `BT_BOXFORM=1`** in `renderAccordionPane` (view_browse_repo.go:640). Default aus → Bestandsgolden byte-identisch. Golden `browse_boxform.golden` für an-Zustand. | 🟢 DONE | Commit `987ce06`, `box_form_flag.go`, `box_form_golden_test.go` |
| S3 | Persistente Filter-Leiste (D02), `f` fokussiert; aktiver Chip=Peach | 🟣 offen | — |
| S4 | Keymap + Picker: `o` Type, `u` Priority, `G` View; Type-/Priority-Picker | 🟣 offen | — |
| S5 | Nested/Flat-Switcher (`G`); Flat-Tabelle (Default Hierarchie, `S`→flach) | 🟣 offen | — |
| S6 | Maus: Klick öffnet Dropdown/Toggle/Chip/Segment | 🟣 offen | — |
| S7 | huh→Inline-Box-Editing (D09); huh + langsame Tests raus | 🟣 offen | — |
| S8 | Politur + VHS-Smoke 80/100 + PO-Abnahme; Merge/Verwerfen | 🟣 offen | — |

## Keymap-Beschluss (D07)
klein = Feld-Aktion: `s` Status · `t` Tags · `a` Parent · `r` Blocking · `e` Body · **`o` Type (NEU)** · **`u` Priority (NEU)**.
groß = View/Global: `S` Sort · `X` Clear · `K` Palette · **`G` View-Toggle Nested/Flat (NEU)**.
Filter: `f`-Einstieg, KEINE Facetten-Einzelkeys.

## Offene Verfeinerungen (Backlog)
- **R1** Label-Farbe in `dropdownBox`: aktuell Rahmen-Farbe (Overlay/Mauve). Farbkarte D08 will `Subtext`/`Hint` fürs Label. In S2 beim VHS-Check angleichen (eigener Label-Style statt frame-Farbe).
- **I01** Box-in-Box-Dichte in S2 prüfen (ggf. Pane-Rahmen weglassen).
- **I02** 80-Spalten: responsive 3/2/1-up in S2/S3/S5, VHS Pflicht.

## CHECKPOINT (2026-07-19) — Validierung durch Nutzer
Box-Form live-sichtbar: `BT_BOXFORM=1 bt` (bzw. `BT_BOXFORM=1` + bt-test). Gate in `renderAccordionPane` wirkt auf Browse-Tree + Backlog + Fullscreen + Review (shared body).

### Befunde aus browse_boxform.golden (100x30)
- **F1 (offen, wichtig):** Box-Form höher als Detail-Pane → Body abgeschnitten, Relations/History unsichtbar. **Kein Scrolling im Box-Modus.** Muss vor produktivem Einsatz gelöst werden (Scroll-Viewport für die Detail-Pane, ODER Box-Form primär im Fullscreen `v`).
- **F2:** I01 (Box-in-Box) tolerabel, kein Pane-Rahmen-Weglassen nötig.
- **F3:** Split-Detail ~52 breit → perRow=1; 2/3-up erst im Fullscreen.

### Offene Entscheidungen für den Nutzer (D-Codes)
- **D10?** Scroll-Strategie für die Box-Form-Detail-Pane (Viewport-Scroll vs. Fullscreen-only vs. kollabierbare Panels). → treibt die nächste Slice.
- **D11?** Gilt der Box-Modus global (auch Backlog/Fullscreen/Review) oder nur Browse-Detail? (Gate sitzt aktuell im shared body → global.)
- Richtungs-Urteil: ist das eine Verbesserung → S3+ (Filter-Leiste, Nested/Flat, Picker, huh-Ersatz) weiterbauen, oder anpassen/verwerfen?

## Nächste Aktion (für Resume)
WARTET auf Nutzer-Validierung am Checkpoint oben. Bei „weiter": zuerst F1 (Scroll) lösen — hängt an D10. Danach S3 (persistente Filter-Leiste, `dropdownBox`-Reuse), dann S4/S5/S7. Kombinierter Code-Reviewer über S1–S2b läuft am Checkpoint.
