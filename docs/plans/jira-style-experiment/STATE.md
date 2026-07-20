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

| S5 | Nested/Flat-Toggle `G` (Views&Global): linke Browse-Pane rendert flache Liste (`view_browse_flat.go`, reuse `backlogRowText`) statt Tree; Master-Detail bleibt, Tree-State bleibt bei Rückschaltung. `flatView`+`flatList` in types.go, `focusedBean()`-Flat-Zweig. Default aus → Bestandsgolden unverändert, neue `browse_flat.golden`. | 🟢 DONE | Commit `bad6c18` |

| S6 | Maus: B6-Offset (+3 Filter-Bar unter `BT_BOXFORM`) in `treeClickRow`/`flatClickRow`/`detailBoxFormClickRow`; Klick auf Detail-Feld-Box öffnet Editor (`detailBoxFormClickRow`+`gridColAt`, `gridColWidths` aus `gridRow` geteilt); Flat-Zeilen-Klick (`flatClickRow`). Render-gegroundete `tea.MouseMsg`-Tests. Golden unverändert. | 🟢 DONE | Commit `cf00b72`, `mouse_boxform_test.go` |

## Restrisiken (offen, nicht blockierend)
- **B9 (Validierung vor Merge):** Klick-Spalten-Hit-Test (`gridColAt`) nur bei 100 Spalten getestet, nicht am 80-Spalten-Rand. tmux/VHS-80-Smoke vor Merge (I02/S8).
- **B8 (für später):** Fullscreen `v` rendert immer Tree, ignoriert `flatView`. Klein.
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

S1–S6 🟢 — **das ganze jira-Modell steht außer S7.** Detail-Box-Form, Filter-Leiste, editierbare Dropdowns (Tastatur+Maus), Nested/Flat-Toggle, Klick-öffnet-Editor. Alles hinter `BT_BOXFORM` (außer additive Keys G/o/u), reversibel, getestet, Bestandsgolden intakt. 22 Commits.

## CHECKPOINT vor S7 (2026-07-20) — Nutzer-Steuerung
- **S7 (huh→Inline-Box-Editing, D09):** der letzte + größte/riskanteste Umbau. Nutzer wollte Timing steuern. Erst jetzt entscheiden: jetzt bauen / nach Live-Test / erstmal so lassen.
- **S8 Validierung:** VHS-Smoke bei 80 + 100 Spalten (`BT_BOXFORM=1`), deckt B9/I02 ab. Braucht vhs/tmux (Nutzer-Tool).
- Merge-Frage: ist der Spike „besser" → auf main mergen (Flag default aus, opt-in) oder verwerfen/anpassen.

## Validierung (2026-07-20)
- **VHS-GIF gerendert** gegen sproutling (114 beans, echte Daten): Box-Form voll bestätigt — Title/Status(s)/Type(o)/Priority(u)/Parent(a)/Tags(t)/Body, 3|2-Grid, Filter-Leiste, salient. Datei: `~/Obsidian/Vault/lean-stack/beans-tui/beans-tui-boxform.gif` (+ `boxform-demo.tape`). PATH-Fix nötig (vhs-zsh sourced ~/.zshrc nicht → `export PATH=/opt/homebrew/bin:$PATH`).
- **80-Spalten-Smoke bestanden** (tmux -x 80 gegen sproutling): kein Overflow, kein Wrap-Bug, feste 3|2 hält, Labels/Werte clampen graziös (Priority→Priorit etc.). **B9/I02 GELÖST.** Capture: scratchpad/cap_80.txt. Damit bei 80 + ~130 Spalten validiert.
- Restkosmetik (kein Bug): bei ≤80 Spalten clampen Feld-Labels/Werte (bewusster D12-Tradeoff). Optionaler Polish später: Kurz-Labels oder 2-up bei extremer Enge.

## Operationalisiert in beans (2026-07-20, N1)
Arbeit wird ab jetzt über **beans** gesteuert, nicht mehr über diese Datei. Epic: **`bt-vy1q`** (trägt den gemeinsamen Kontext + Constraints + erledigte Slices mit SHAs).
Offene Kinder: `bt-ze10` Detail-Scroll (F1, high) · `bt-fy5d` Footer entschlacken (N2) · `bt-pl5p` Projekt-Slug aus IDs (N5) · `bt-oqsv` leere Footer-Zeile (N6) · `bt-ty48` GIF Body-Scroll (N3, blocked-by ze10) · `bt-z4w7` Value-Menü-Alias (B7) · `bt-s90e` Fullscreen ignoriert flatView (B8) · `bt-dovm` S7 huh-Ersatz (draft) · `bt-2o9a` Merge-Entscheidung (draft).
`beans list --ready` zeigt den nächsten Schritt. Diese STATE.md bleibt Kontext-/Historien-Anker.

## Laufender Stand bei Kontext-Kompaktierung (2026-07-20)
- **In Arbeit:** `bt-ze10` (Detail-Scroll) — ein Implementer-Subagent wurde dafür dispatcht und lief zum Zeitpunkt der Kompaktierung. Bei Wiederaufnahme ZUERST prüfen: `git log --oneline -5` + `beans show bt-ze10` — hat er committet und das bean auf `completed` gesetzt? Falls das bean auf `in-progress` steht und kein Scroll-Commit existiert, ist die Arbeit NICHT erledigt → neu dispatchen.
- **Offener Fehler:** `bt-ce7i` — Commit `d4a5367` hat versehentlich ~35 fremde `.beans`-Dateien mitgenommen (Glob statt Einzelpfade). PO muss A/B/C wählen. Kein Datenverlust, nur falsch einsortiert.
- **Regel ab sofort:** in `.beans/` nur explizite Einzelpfade stagen, nie `git add .beans/*` — das Repo trägt dauerhaft fremde uncommittete bean-Änderungen.

## Nächste Aktion (für Resume)
1. `beans list --ready` unter Epic `bt-vy1q` (das Epos trägt allen gemeinsamen Kontext + Constraints).
2. Reihenfolge-Empfehlung: `bt-ze10` (falls offen) → `bt-ty48` (GIF, blocked-by ze10) → `bt-1o4g` (Feld-Nav, blocked-by ze10) → `bt-a3a8` (Picker-Suche, high) → Platz-Trilogie `bt-fy5d`/`bt-pl5p`/`bt-oqsv` → `bt-z4w7`/`bt-s90e`.
3. `bt-dovm` (S7 huh-Ersatz) + `bt-2o9a` (Merge) stehen auf **draft** — brauchen PO-Freigabe, nicht eigenmächtig starten.
4. Subagenten-Dispatch-Muster: voller Testlauf im VORDERGRUND anweisen (Agenten detachen sonst vor Testende und müssen per SendMessage resumed werden). Bei „S7 jetzt": großer Umbau, eigene Slice-Kette planen (Create-Form inline, Picker→eigene maus-native Popups, huh + langsame huh-Drive-Tests entfernen). Bei „erst validieren": VHS-80/100-Smoke + Live-Test, dann entscheiden. Alles weiter additiv + gated, bis Spike als „besser" abgenommen.

## Stand 2026-07-20 (nach Kompaktierung) — PO-Abnahme + Parallelbetrieb

### PO-Entscheidungen
- **Spike ist abgenommen.** Nach der Validierung gegen sproutling (VHS-GIF + 80-Spalten-Smoke)
  gilt das Experiment als erfolgreich. `experiment/jira-style-ui` wird **vollstaendig** auf
  `main` gemerged. Damit ist die Richtungsfrage beantwortet, offen bleibt nur das Timing
  (bt-2o9a, draft -> todo).
- **bt-ce7i geschlossen (Option B):** die ~35 fremden `.beans`-Aenderungen aus `d4a5367`
  bleiben, wo sie sind — sie kommen mit dem Merge ohnehin auf main. Kein History-Rewrite.
  Die Lehre bleibt in Kraft: in `.beans/` nur explizite Einzelpfade stagen, nie ein Glob.

### Erledigt
- **bt-ze10 (F1, Detail-Scroll) 🟢** — Commit `8e5a869`. Geclampter Scroll-Viewport fuer die
  Box-Form-Detail-Pane, komplett hinter `boxFormEnabled()`. Neue Felder `boxFormScroll` +
  `boxFormScrollBean` (abgeleiteter Reset per Bean-ID-Vergleich statt Nachziehen an jeder
  Cursor-Bewegung). Wiederverwendet das vorhandene `scrollView`-Primitiv. Antrieb: up/down in
  `keyDetailFocus` + Mausrad ueber der Pane (`boxFormWheelHit`), beides ueber die eine
  Funktion `adjustBoxFormScroll`. Neue Tests `box_form_scroll_test.go` (7). Golden unveraendert.
  **Scope-Cut:** Fullscreen scrollt NICHT — als Prelude in bt-s90e protokolliert (Commit `7cd3042`).

### Grounding (Investigatoren, read-only) — in die beans gegossen
- **bt-a3a8** (Commit `2efddb5`): Kernbefund — **die Suche existiert schon** im Tag-Picker
  (`box_picker_tag.go`, `textinput` + `filterTagItems`). Die Aufgabe ist ein Port, kein Neubau.
  Filter-Strip aus `gridRow`/`scalarCell`/`dropdownBox` wiederverwendbar.
- **Trilogie bt-fy5d/bt-oqsv/bt-pl5p** (Commit `b72ad49`): **muss als EIN Paket laufen** —
  die drei teilen sechs Golden-Dateien. Fallstricke: `mouse.go:112` (`footH ... + 2`) haengt an
  der Footer-Hoehe (sonst Klick-Offset); ID wird an ZWEI Stellen gerendert (`treeRowText` +
  `backlogRowText`); Footer-Ausduennung darf nur die Anzeige treffen, nicht die
  keymap-Registrierung (Drift-Guard).

### Laufend (2 Implementer parallel)
- **Trilogie** (`bt-fy5d`+`bt-oqsv`+`bt-pl5p`) im **Haupt-Tree** auf `experiment/jira-style-ui`.
- **bt-a3a8** (Picker-Suche) in einem **isolierten Worktree**, eigener Branch.
  Muss beim Zusammenfuehren zurueckgemerged werden; Golden-Konflikte moeglich, deshalb wurde
  ihm untersagt, die geteilten Golden zu regenerieren.

### Parallelisierungs-Regel (gelernt)
Der Flaschenhals ist der EINE Working Tree, nicht die Agentenzahl. `types.go`/`update.go`/
`view_browse_repo.go`/`mouse.go` werden von fast jedem bean beruehrt. Read-only-Arbeit
(Grounding, Scope, Review) laeuft gefahrlos parallel; Implementierung entweder sequenziell
im Haupt-Tree oder in Worktrees mit disjunkten Dateimengen.

### Offen danach
`bt-ty48` (GIF Body-Scroll — jetzt entblockt, ze10 fertig) · `bt-1o4g` (Feld-Navigation,
entblockt) · `bt-s90e` (Fullscreen: flatView + Scroll, zwei Luecken zusammen) · `bt-z4w7`
(Value-Menue-Alias) · `bt-dovm` (S7 huh-Ersatz, draft, braucht PO-Freigabe) · `bt-2o9a` (Merge).

### Offene Frage an den PO
Wenn der Branch komplett auf main geht: soll `BT_BOXFORM` weiter default-aus sein? Ein
abgenommenes Feature hinter einem Env-Flag zu verstecken ist fragwuerdig. Optionen: Default
an (Flag als Notausgang) / Flag ganz raus / so lassen. Noch nicht entschieden, gehoert vor
den Merge.
