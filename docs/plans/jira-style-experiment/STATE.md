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
| S2b | Live-Wiring per **Env-Flag `BT_BOXFORM=1`**: Detail-Pane rendert `detailBoxForm` statt Accordion, wenn gesetzt. Default aus → alle Bestandsgolden/Tests unverändert. EIN neuer Golden-Test für den an-Zustand. Max reversibel, direkt vergleichbar. Danach I01-Dichte + VHS/`BT_BOXFORM=1 bt` bei 80/100. | 🟣 NÄCHSTES | — |
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

## Nächste Aktion (für Resume)
S2b: Implementer findet die EINE Stelle, wo die Live-Detail-Pane komponiert wird (Accordion-Body), und gated sie hinter `BT_BOXFORM=1` → `detailBoxForm(idx,bean,paneW)`. Default aus = Bestandsgolden unberührt. EIN neuer Golden-Test für den an-Zustand. sonnet, explore+wire+golden in einem. Danach visuell prüfen (`BT_BOXFORM=1 bt` bzw. VHS 80/100), I01-Dichte bewerten. Reviewer-Checkpoint nach S2b.
