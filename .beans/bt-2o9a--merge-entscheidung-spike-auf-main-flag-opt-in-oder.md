---
# bt-2o9a
title: 'Merge-Entscheidung: Spike auf main (Flag opt-in) oder verwerfen'
status: completed
type: task
priority: normal
created_at: 2026-07-20T07:26:50Z
updated_at: 2026-07-21T11:57:57Z
parent: bt-vy1q
---

**Freigabe-Gate.** Der Spike liegt vollstaendig auf `experiment/jira-style-ui` (Default AUS, opt-in per `BT_BOXFORM=1`). Entscheidung liegt beim PO — Ausfuehrung ⊥ Freigabe, der Agent merged NICHT selbst.

## Entscheidungsgrundlage
- GIFs: `beans-tui-boxform.gif` (breit), `beans-tui-boxform-narrow.gif` (Clamping ~80 Spalten) in `~/Obsidian/Vault/lean-stack/beans-tui/`
- 80-Spalten-tmux-Smoke bestanden (kein Overflow/Wrap-Bug)
- Voller Testlauf durchgehend gruen, Bestandsgolden byte-identisch bei Flag AUS

## Optionen
- **Mergen als opt-in** (Flag bleibt, Default AUS) — geringes Risiko, Feature verfuegbar
- **Erst Polish** (bt-fy5d Footer, bt-pl5p ID-Slug, bt-oqsv Leerzeile, bt-ze10 Scroll), dann mergen
- **Verwerfen/parken** — Branch bleibt, kein Merge

## Akzeptanz
- [ ] PO-Entscheidung dokumentiert
- [ ] Bei Merge: Reviewer-Durchlauf ueber den gesamten Branch-Diff, dann Merge durch PO
- [ ] Bei Merge: `BT_BOXFORM` in README/CLAUDE.md dokumentieren


## PO-Signal 2026-07-20: Merge ist beschlossen

Nach der Validierung gegen sproutling (VHS-GIF echte Daten + 80-Spalten-tmux-Smoke) hat der
PO den Spike als **erfolgreich** eingestuft: `experiment/jira-style-ui` wird **vollstaendig**
auf `main` gemerged (inkl. der ueber d4a5367 mitgezogenen fremden `.beans`-Aenderungen —
siehe bt-ce7i, dort Option B genau darauf gestuetzt).

Damit ist die Richtungsfrage ("besser oder verwerfen?") beantwortet. Offen bleibt nur noch
das **Timing**: der Merge erfolgt, wenn die restlichen Kinder von bt-vy1q durch sind.
Der Merge selbst bleibt PO-Autoritaet (Merge-Gate) — der Agent merged nicht eigenmaechtig.

## PO-Entscheidung 2026-07-21: JETZT mergen
PO-Freigabe 'bt-2o9a >> jetzt mergen'. Ausgefuehrt als opt-in (BT_BOXFORM bleibt, Default AUS).

### Akzeptanz erfuellt
- [x] PO-Entscheidung dokumentiert (hier)
- [x] Branch-Diff-Review: inkrementell ueber den Sprint (alle bt-vy1q-Epics/Kinder accepted, zuletzt bt-adkn 3/3); Merge-Gate zusaetzlich: voller 'command go test ./...' gruen (EXIT=0, internal/tui 153s), Bestandsgolden byte-identisch bei Flag AUS.
- [x] BT_BOXFORM dokumentiert: README §Experimental + CLAUDE.md (opt-in, Default AUS).

### Merge
experiment/jira-style-ui (112 Commits, 141 Dateien) --no-ff auf main. Inkl. der ueber d4a5367 mitgezogenen fremden .beans-Aenderungen (bt-ce7i Option B, PO-bestaetigt). Push/Tag bleibt separates PO-Gate (nicht ausgefuehrt).

## Merge ausgefuehrt 2026-07-21 (main ae2efe3)
experiment/jira-style-ui --no-ff auf main. GLOSSARY.md add/add-Konflikt (Obsidian-uid-Frontmatter + Superset) zugunsten der experiment-Version geloest (inhaltlicher Superset, kein main-only-Inhalt verloren). Post-Merge Build + voller 'command go test ./...' auf main GRUEN (EXIT=0). Push + Release-Tag bleibt separates PO-Gate (nicht ausgefuehrt).
