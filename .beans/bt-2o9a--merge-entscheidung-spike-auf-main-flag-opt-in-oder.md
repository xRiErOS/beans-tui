---
# bt-2o9a
title: 'Merge-Entscheidung: Spike auf main (Flag opt-in) oder verwerfen'
status: draft
type: task
priority: normal
created_at: 2026-07-20T07:26:50Z
updated_at: 2026-07-20T07:26:50Z
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
