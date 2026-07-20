---
# bt-ty48
title: 'GIF: Body-Scrollen veranschaulichen'
status: todo
type: task
priority: normal
created_at: 2026-07-20T07:25:37Z
updated_at: 2026-07-20T07:25:37Z
parent: bt-vy1q
blocked_by:
    - bt-ze10
---

**Nebenbefund N3 (PO, 2026-07-20).** Ein GIF soll das Scrollen des Body zeigen. Setzt den Scroll aus `bt-ze10` voraus.

## Vorgehen
VHS-Tape analog zu den bestehenden in `~/Obsidian/Vault/lean-stack/beans-tui/`:
- `boxform-demo.tape` (breit, ~130 Spalten) und `boxform-narrow.tape` (~80 Spalten, Clamping) als Vorlage
- **PATH-Falle:** vhs-zsh sourced `~/.zshrc` NICHT → im `Hide`-Block `export PATH=/opt/homebrew/bin:$PATH` (sonst findet bt die beans-CLI nicht und der Tree bleibt leer)
- **Kurzer sichtbarer Launch (N4):** im `Hide`-Block `alias bts='BT_BOXFORM=1 /Users/erik/Obsidian/tools/beans-tui/beans-tui-repository/bin/bt'` definieren, dann sichtbar nur `bts` tippen
- Demo-Repo: `/Users/erik/Obsidian/tools/sproutling` (114 beans, echte lange Bodies)
- Alle Pfade in der Tape MUESSEN gequotet sein (`Output "..."`, `Screenshot "..."`), sonst Parser-Fehler
- Bean mit langem Body waehlen, Detail fokussieren (tab), scrollen (up/down bzw. Mausrad)

## Akzeptanz
- [ ] GIF zeigt sichtbar, wie der Body im Detail scrollt und Relations/History erreichbar werden
- [ ] Gespeichert unter `~/Obsidian/Vault/lean-stack/beans-tui/` (Tape daneben ablegen)
- [ ] Vorher per `Screenshot` verifiziert, dass Daten geladen sind (nicht leerer Tree)
