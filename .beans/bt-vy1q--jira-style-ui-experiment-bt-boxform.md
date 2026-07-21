---
# bt-vy1q
title: jira-Style-UI Experiment (BT_BOXFORM)
status: in-progress
type: epic
priority: high
created_at: 2026-07-20T07:24:38Z
updated_at: 2026-07-21T11:46:14Z
---

Spike: ist eine flachere, salientere jira-artige Darstellung eine Verbesserung fuer beans-tui?
Vorbild: jiratui.sh (Python/Textual — nur Design-Vorbild, kein Code-Transfer).

## Kontext fuer alle Kinder (DRY)
- **Branch:** `experiment/jira-style-ui` (main-direkt-Repo; Spike, kein Merge-Zwang)
- **Feature-Flag:** `BT_BOXFORM=1` — Default AUS. Alles additiv + gated; Bestandsgolden muessen byte-identisch bleiben.
- **Quelle der Wahrheit:** `docs/plans/jira-style-experiment/design-spec.md` (Entscheidungen D01-D12, Catppuccin-Farb-Salienzkarte, Mockups, Risiken I01-I04)
- **Plan:** `docs/plans/jira-style-experiment/implementation-plan.md` (Slice-Roadmap S1-S8)
- **Resume-Anker/Status:** `docs/plans/jira-style-experiment/STATE.md` — bei Wiederaufnahme ZUERST lesen

## Harte Constraints
- Voller Testlauf `command go test ./...` (OHNE `-short`) gruen vor jedem Commit
- Keine Hex-Literale in Views — nur Tokens aus `internal/theme/`
- Keymap Single Source: `internal/tui/keymap.go`; jede neue Bindung MUSS in `helpGroups()` (Drift-Guard-Test)
- Case-Konvention D07: **klein** = Feld-Aktion (`s t a r e o u`), **gross** = View/Global (`S X K G`)
- Build: `command go build -o bin/bt .`

## Erledigt (S1-S6, Belege in git)
- S1 `2b08977` dropdownBox-Primitiv (Label+Hotkey IM Rahmen, lipgloss kann das nicht nativ)
- S2 `f76b6e9` detailBoxForm Scalars + R1 (Label=Subtext)
- S2c `d62edf0` panelBox multi-line + Body/Relations/History
- S2b `987ce06` Live-Wiring gated in `renderAccordionPane`
- S2e `e27c5f2` feste 3|2-Gruppierung (D12) + Geometrie-Fixes B1-B5
- S3 `db52457` persistente Filter-Leiste (Chips, aktiv=Peach)
- S4 `5dc3d01` editierbare Dropdowns: Keys `o`=Type, `u`=Priority
- S5 `bad6c18` Nested/Flat-Toggle `G`
- S6 `cf00b72` Maus: B6-Offset, Klick-oeffnet-Feld-Editor, Flat-Klick

## Validiert
- VHS-GIF gegen sproutling: `~/Obsidian/Vault/lean-stack/beans-tui/beans-tui-boxform.gif`
- Clamping-GIF schmal: `beans-tui-boxform-narrow.gif`
- 80-Spalten-tmux-Smoke bestanden: kein Overflow, kein Wrap-Bug, Labels clampen graziös

## Offen
Siehe Kinder-beans. Merge auf main erst nach PO-Freigabe.


## Sitzungs-Abschluss 2026-07-20 — Runde 2 + 3 erledigt

`experiment/jira-style-ui` @ `a276121`, **80 Commits vor `main`**, Tree sauber, voller
Testlauf gruen, 80-Spalten-Smoke gegen sproutling bestanden. Alle Agent-Worktrees entfernt.

**17 Kinder completed.** In dieser Sitzung dazugekommen: bt-ze10 (Detail-Scroll, F1) ·
bt-1o4g (Feld-Cursor) · bt-s90e (Fullscreen: flatView + Scroll) · bt-a3a8 (Picker-Suche) ·
bt-z4w7 (Footer-Labels abgeleitet) · bt-pt1r + bt-pl5p (Slug-Kuerzung) · bt-fy5d + bt-oqsv
(Footer-Platz) · bt-8d35 (Fokus-Modell) · bt-hd42 (Klick-Offset) · bt-f68z + bt-vpvu (Tree
+ Maus) · bt-mx4k (Palette `K`) · bt-6nuz (Relations-Overlay) · bt-oox1 (Detail-Politur) ·
bt-ce7i (Fehl-Commit, Option B).

**Der PO hat den Spike abgenommen** — der Branch wird vollstaendig auf `main` gemerged
(bt-2o9a, Timing offen). Damit ist die Richtungsfrage des Experiments beantwortet.

### Neue Entscheidungen (design-spec.md)
D03 REVIDIERT (Klick: einfach waehlt, doppelt klappt) · D13 (`v` nur im Box-Modus im
Footer) · D14 (History bleibt im tab-Zyklus). Dazu das Fokus-Modell: **tab bewegt INNERHALB
der Region, esc verlaesst sie**; `enter` oeffnet ueberall; `X` leert Filter;
`pgup`/`pgdn` blaettern; Palette nur noch `K`.

### Verbindliche Sprache
Neu: `docs/GLOSSARY.md` (auch auf `main`, Commit `ef12bf5`) — boxed field, Box-Titel,
Box-Badge, Panel, Region, Views mit Datei-Zuordnung, Terminal-Fallen. Pointer in CLAUDE.md.

### Offen
bt-adkn (Body blaettern) · bt-p78f (Anker-Leiste + Pencil) · bt-ty48 (GIF, bewusst zuletzt)
· bt-dovm (draft, PO-Freigabe) · bt-2o9a (Merge auf main).

## Review 2026-07-21 (PO, heutige Arbeit)

- US-01 · Body seitenweise blaettern (pgup/pgdn) · r: PageDown/PageUp scrollt die Tree-Pane statt Body → bt-adkn (rework, B1)
- US-02 · Seiten-Indikator Punkte · r: statisch + Platzierung falsch (soll an Body-Box) → bt-adkn (rework, B2/B3)
- US-03 · Ankerleiste Chips oben in der Body-Box · a
- US-04 · Klick auf Chip springt zur Ueberschrift · r: Klick scrollt real nicht an die Stelle → siehe Scrap-Entscheidung
- US-05 · Ankerleiste bei 80 Spalten gekuerzt mit Ellipsis · Befund PO: 'abgeschnitten hilft als Inhaltsverzeichnis nicht' → Scrap-Entscheidung

### PO-Entscheidung D01: Inhaltsverzeichnis-Experiment (bt-p78f) SCRAPPED
Woertlich: 'Wir bauen das Experiment mit dem Inhaltsverzeichnis zurueck. Die Relevanz ist gering und die Implementierung zu aufwaendig. Daher scrapped.' Code (Parser + Ankerleiste) wird per git revert zurueckgebaut; bt-adkn (Paging) bleibt separat in rework.

## Review 2026-07-21 (Rework-Abnahme bt-adkn)
US-01 · Body-Blaettern mit pgup/pgdn ohne Tab, Tree-Cursor bleibt · a
US-02 · Sticky Seiten-Indikator in der Body-Titelzeile · r: 'Die Position passt noch nicht sauber. das (e) immer an der gleichen Stelle belassen und dafuer ── ●○ verschieben. Das fuehrt zu einer stabilen praesentation' → bt-adkn

## Review 2026-07-21 — Nachtrag (bt-adkn Rework final)
US-02 · Seiten-Indikator-Position (Body-Header) · a (nach 3 Nacharbeits-Runden: links verankert + fixe Zahlbreite + (e) an/aus stabil)
US-03 · n/N-Fallback bei sehr langem Body · a
Bilanz bt-adkn Rework: 3/3 accepted (US-01/US-02/US-03). bt-adkn completed. Offen im Epic: bt-ty48 (GIF), bt-2o9a (Merge-Gate PO).
