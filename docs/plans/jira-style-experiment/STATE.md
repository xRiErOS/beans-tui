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
| S2e | Feste 3|2-Gruppierung (D12: Title full · Status\|Type\|Priority · Parent\|Tags · Body/Relations/History full, KEIN 1-up-Collapse) + Reviewer-Fixes B1–B5 (exakte Spaltenbreite, Hotkey-Overflow-clamp, gemeinsame boxTop/BottomBorder, exakte-Breite-Tests). Beide box-form-Golden regeneriert. **Nebeneffekt: kompakte Form passt in Pane → F1 für Normalfälle gelöst.** | 🟢 DONE | Commit `e27c5f2`, `gridRow`/`scalarCell` |
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

## Nutzer-Urteil am Checkpoint (2026-07-20)
Richtung ok, ABER Layout anpassen: **1-up-Stapeln = Platzverschwendung.** Gewünschte FIXE Gruppierung im Detail (NICHT auf 1-up einklappen, egal wie schmal):
- **D12:** Title full-width · `Status | Type | Priority` in EINER Zeile (3 Spalten) · `Parent | Tags` in einer Zeile (2 Spalten) · Body full-width · (Relations/History full-width darunter).
- Also: responsive perRow-Schwellen (3/2/1 nach Breite) RAUS → feste Zeilen-Gruppen; Spalten schrumpfen mit, klappen nicht auf 1-up.

## Reviewer-Findings (S1–S2b, umsetzen in S2e)
- **B1 (high):** `detailBoxForm` Grid `colW = (w-(n-1)*gap)/n` verliert Rest → Zeilen zu schmal. Fix: Rest auf die ersten `rem` Spalten verteilen, Zeile summiert exakt auf `width`.
- **B2/B3 (med):** `dropdownBox`/`panelBox` Hotkey-Badge: wenn `width(badgeSeg)+5 > width` → untere Zeile zu breit. Fix: fill-clamp so, dass Zeilenbreite nie `width` übersteigt (Badge ggf. kürzen/weglassen).
- **B4 (low):** `box_panel` dupliziert Border-Logik von `box_dropdown` → gemeinsame `boxTopBorder(label,w,frame)`/`boxBottomBorder(hotkey,w,frame)` extrahieren, beide nutzen sie.
- **B5 (low):** Tests prüfen nur `>width` → exakte `==width`-Assertion je Zeile ergänzen (fängt B1).

| S3 | Persistente Filter-Leiste oben im Browse (Type/Status/Priority/Tags-Chips via `gridRow`, aktiv=Peach, leer=(any)=Hint), additiv+gated in `viewBrowseRepo`. 3 Zeilen aus `bodyH` zurückgewonnen (Frame bleibt 30). `f` öffnet weiter das bestehende Overlay (keine Facetten-Keys). | 🟢 DONE | Commit `db52457`, `box_filter_bar.go`(+_test) |

| S4 | Editierbare Dropdowns: Keys `o`=Type, `u`=Priority (D07 lowercase) auf bestehendes Value-Menü verdrahtet (mutiert via beans-CLI). helpGroups ergänzt (drift-guard grün). Nicht gated. Type/Priority-Edit existierte schon → reine Verdrahtung. | 🟢 DONE | Commit `5dc3d01`, `keymap.go`/`update.go` |

## Restrisiken (offen, nicht blockierend)
- **B7 (kosmetisch, für S6/später):** Value-Menü-Schließen-Alias + Footer hardcoden `s` unabhängig von der Gruppe (o/u-geöffnetes Menü schließt auch mit `s`; Footer zeigt `s`). esc schließt immer → nichts kaputt, nur Label-Mismatch.
- **B6 (Maus, für S6):** `treeClickRow`/`clickPaneGeometry` (mouse.go) NICHT um die +3 Filter-Bar-Höhe korrigiert, wenn `BT_BOXFORM` an → Tree-Klicks 3 Zeilen versetzt. Nur im Box-Modus + Maus relevant. In S6 (Maus) mitfixen.
- **F1-Rest:** sehr langer Body / viele Relations überläuft die Pane weiterhin (kein Scroll im Box-Modus). Normalfall passt. Scroll-Strategie D10 (Viewport vs. Fullscreen-only) noch offen — bei Bedarf mit Nutzer klären.
- **Narrow-width:** unter ~30 Zellen Detail-Breite bricht die exakte-Breite-Garantie der 3-Spalten-Zeile (dropdownBox-Floor 8/Spalte). Im Split erst bei Terminal <~68 Spalten relevant. VHS-80-Check in S2b-Nachlauf offen.

## MEILENSTEIN (2026-07-20): Detail-Box-Form + Filter-Leiste live via `BT_BOXFORM=1`
Headline-Features (die zwei vom Nutzer gelobten) stehen: salientes Box-Detail (3|2-Grid, Hotkeys im Rahmen) + persistente Filter-Leiste. Default aus = alles wie bisher. Demobar: `BT_BOXFORM=1 bt`.

## Nächste Aktion (für Resume)
S1–S4 🟢 (zwei Headline-Features + editierbare Dropdowns live). **CHECKPOINT: wartet auf Nutzer-Steuerung für S5+.**

Offene Weichen für den Nutzer:
- **S5 (Nested/Flat):** überschneidet sich mit bestehendem Backlog-View (`b` = bereits flache Liste). Weiche: „Flat" = Backlog-Zeilen im Browse-Master-Detail wiederverwenden ODER eigenständiger Flat-Renderer? `G` (D07 uppercase) + `flatView`-State, default aus, neue Golden.
- **S7 (huh→Inline-Box-Editing):** großer/riskanter Umbau (D09). Timing bewusst offen gelassen — Nutzer wollte steuern.
- **S6 (Maus):** B6 (Klick-Offset +3 durch Filter-Bar) + B7 (Value-Menü-Label) mitfixen.

Bei „einfach weiter/entscheide selbst": S5 als eigenständiger Flat-Renderer (Backlog-Row-Rendering wiederverwenden, Master-Detail behalten), dann S6, S7 zuletzt. Alles weiter additiv + gated, bis Spike als „besser" abgenommen.
