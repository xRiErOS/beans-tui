# B06-Beleg — Accordion-Header Chevron entfernen (B05) + Teal-Experiment (B06)

bean `bt-czpf` (E8 Task 2), Epic `bt-ntoz` (B05/B06). Vorher/Nachher-Beleg für den
PO-Sign-off vor v1-Freigabe (B06 ist EXPERIMENT, kann per Ein-Zeilen-Rollback auf
`theme.Muted` zurückgerollt werden, falls der PO das Ergebnis ablehnt).

## Erzeugung

`tmux capture-pane -e -p` (ANSI-erhaltend), Terminalgröße 100x30, `./bin/bt` im
eigenen Repo (`beans-tui-repository`, dogfooding — reale bean-Daten, aber
NUR angesehen, nichts mutiert):

- `before.txt` — Build VOR dem Fix (via `git stash` auf einen sauberen Vorher-Stand
  gebracht, Binary in `/tmp/bt-before` gebaut, dann `git stash pop` zum Wiederherstellen).
- `after.txt` — Build NACH B05+B06 (`bin/bt`, aktueller Repo-Stand).

Beide Captures zeigen die Browse-Ansicht direkt nach Start (Detail-Pane rechts,
META immer offen [PF-1], BODY/RELATIONS/HISTORY geschlossen — der Standard-
Ruhezustand `accOpen==0`).

## Token/RGB

| Zustand | Vorher | Nachher |
|---|---|---|
| Geschlossener Section-Header-Titel (BODY/RELATIONS/HISTORY) | `theme.Muted` — Hint-Grau `#7c7c7c` (ANSI `38;2;124;124;124`) | `theme.HeaderInactive` (NEU, theme.go) — Teal `#8bd5ca` (ANSI `38;2;139;213;202`) |
| Offener Section-Header-Titel (META, PF-1 immer offen) | `theme.Header` — Mauve `#c6a0f6`, bold | unverändert |
| Meta-Label-Spalte (z.B. `status:`, view_detail_bean.go) | `theme.Muted` — Hint-Grau `#7c7c7c` | unverändert (B06 betrifft NUR den Accordion-Section-Header, nicht die Feldliste) |
| Chevron-Suffix (`"  ▾"` offen / `"  ▸"` zu) | vorhanden nach jedem Header | ENTFERNT (B05, beide Zustände) |

## Objektiver Diff-Beleg (grep-Zählung, nicht interpretiert)

```
                              before.txt   after.txt
"▾" oder "▸" im Screen-Dump        5            1     (verbleibender 1x = Tree-Pane-
                                                        eigener Node-Expand-Marker,
                                                        NICHT der Accordion — B05 betraf
                                                        nur accordion.go)
ANSI "38;2;139;213;202" (Teal)     0            3     (exakt BODY+RELATIONS+HISTORY,
                                                        die 3 geschlossenen Sektionen)
```

## Eigenbewertung (Implementer, nicht PO-Freigabe)

Die Teal-Färbung hebt die geschlossenen Section-Header jetzt sichtbar von der
Meta-Label-Spalte ab (beide vorher identisch Hint-Grau) — der von B06 benannte
Verwechslungs-Fall ist im Rendering messbar behoben. Ob der Farbton selbst
(Teal vs. z.B. Sky/Sapphire) gefällt, ist der eigentliche Experiment-Charakter
und PO-Entscheidung — Rollback auf `theme.Muted` ist eine Ein-Zeilen-Änderung in
`accordion.go` (`theme.HeaderInactive.Render(s.title)` → `theme.Muted.Render(s.title)`).

## Status

PO-Sign-off AUSSTEHEND (Epic `bt-ntoz` nicht auf `completed` gesetzt — PO-Gate).
