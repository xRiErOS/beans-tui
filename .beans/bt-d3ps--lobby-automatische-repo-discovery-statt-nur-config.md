---
# bt-d3ps
title: 'Lobby: automatische Repo-Discovery statt nur config-Registrierung'
status: todo
type: feature
created_at: 2026-07-17T09:48:21Z
updated_at: 2026-07-17T09:48:21Z
parent: bt-5uzr
---

NB aus PO-Review E12 Runde 2 (2026-07-17, US-03-Kontext). PO wörtlich: „im lean-stack repo sind keine Projekte registriert. Daher muss hier eine automatische discovery ergänzt werden, damit ich von jeder beliebigen Stelle alle Projekte aufrufen kann."

Ist-Stand: Lobby-Repo-Liste speist sich ausschließlich aus registrierten Repos (config.yaml) — startet man bt in einem Repo ohne Registrierung, ist die Lobby leer/unvollständig.

Soll: Automatische Discovery von beans-Repos (Kandidaten-Mechanik zu klären: Scan definierter Wurzelverzeichnisse? `.beans.yml`-Suche? Merge mit config-Registrierung + Persistenz gefundener Repos?). Von jeder beliebigen Stelle aus müssen alle Projekte aufrufbar sein.

Offene Designfragen für Planner (vor Umsetzung PO klären):
- Discovery-Wurzel(n): fest (~/Obsidian/tools?), konfigurierbar, oder $HOME-weit mit Tiefenlimit?
- Gefundene Repos automatisch in config.yaml persistieren oder nur Session-flüchtig anzeigen?
- Performance: Scan bei Lobby-Öffnen (Latenz!) vs. Hintergrund/Cache.
