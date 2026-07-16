---
# bt-gdkx
title: 'US-08: Tags nicht sichtbar in Tree/Detail (nur Filter-Facette)'
status: todo
type: bug
priority: normal
tags:
    - rejected
created_at: 2026-07-15T19:21:19Z
updated_at: 2026-07-16T06:21:24Z
parent: bt-tct9
---

## Befund (E6 T2, US-08-Validierung, 2026-07-15)

design-spec.md §10 US-08 (redefiniert PF-14): "beans mit Tag `to-review`/
`accepted`/`rejected` sind über Tree/Detail/Filter/Suche auffindbar wie jeder
andere Tag." Live geprüft gegen alle vier genannten Flächen — nur EINE davon
liefert tatsächlich Sichtbarkeit:

| Fläche | Ergebnis | Beleg |
|---|---|---|
| Tree | FEHLT | `treeRowText` (`internal/tui/view_browse_repo.go:413-421`) rendert nur Indent+Marker+StatusIcon+TypeIcon+ID+Title — kein Tag-Indikator, auch kein generischer Punkt |
| Detail (META) | FEHLT (by design) | `internal/tui/view_detail_bean.go:7-11` Doc-Stamp: "Tags is no longer part of Meta at all" (PF-1/PF-4, fixe 6-Feld-Liste title/status/type/priority/created_at/updated_at) |
| Filter (`f`) | FUNKTIONIERT | Live verifiziert: Tags-Facette mit Nutzungszählern, Toggle `to-review` → Tree filtert auf exakt den 1 passenden Bean, Suchkopf zeigt `⌕ Tags:to-review` |
| Suche (`/`) | FEHLT | `tag:to-review` Bleve-Feldsyntax liefert 0 Treffer — beans-CLI-Suche unterstützt nur slug/title/body (deckt sich mit T1s `status:`-Scope-Fund, kein neuer Bug, gleiche Wurzelursache) |

Tags sind aktuell NUR sichtbar über eine Mutations-nahe Overlay (Tag-Picker
`t`, zeigt aktuelle Tags mit [x]/[ ]) oder als Yank-Klartext (`context.go:43-44`,
"- Tags: ..."-Zeile im OSC52-Copy) — beides kein passiver Lese-Pfad.

**Toter Code als Nebenbefund:** `tagsInline`/`tagSwatch`
(`internal/tui/render_shared.go:91-110`) sind definiert, haben aber in der
gesamten Produktion KEINEN Aufrufer mehr (vermutlich Rest der alten
4-Zeilen-Summary, die PF-1/PF-4 laut eigenem Doc-Stamp ersetzt hat) — Hinweis
darauf, dass die Tag-Sichtbarkeit beim Meta-Redesign ersatzlos verschwunden
ist, nicht bewusst weggelassen wurde.

## Auswirkung

Praktisch nutzbar bleibt der Weg über `f` (Filter) — ein PO kann gezielt nach
`to-review` filtern und sieht dann exakt die wartenden Beans. Aber "auf einen
Blick sehe ich den Review-Stand" (US-08-Kernaussage) ist beim normalen
Tree-Browsen NICHT gegeben: nichts unterscheidet optisch einen Bean mit
`to-review`-Tag von einem ohne. PO muss aktiv wissen, danach zu filtern.

## Empfehlung (nicht umgesetzt, nur dokumentiert)

Kein Fix in diesem Validierungs-Task (Validierung sammelt Evidenz, ändert
keinen Code). Kandidaten für einen Fix-Task: (a) Tag-Chip-Zeile im Detail
wieder sichtbar machen (z.B. eigene Zeile unter dem Kopfblock, nutzt bereits
vorhandenes `tagsInline`), und/oder (b) ein Tree-Row-Indikator für
Beans mit Tags (z.B. reserviertes Glyph-Suffix, PF-12-Gutter-Konvention
beachten). PO-Entscheid nötig: ob/wie viel UI-Fläche das rechtfertigt (lean-
stack-Prinzip: nicht automatisch nachrüsten ohne Anfrage).

Repro: `/tmp/bt-scratch-a` (E6 T2 Scratch-Repo, git-ignored, ggf. bereits
aufgeräumt) — Bean mit `--tag to-review` anlegen, `bt` starten, Tree
inspizieren (kein Indikator), Detail öffnen (META hat kein Tags-Feld), `f` →
Tags-Facette togglen (funktioniert).

Quelle: bean bt-9yvh (E6 T2), docs/_free-notes/e6-t2-evidence.md.



## Auflösung (2026-07-16)

Gelöst durch **D01** (Grilling-Entscheidung, Epic `bt-ntoz`) — Umsetzung
E8 Task 1 / bean `bt-e6q9`, Commit `397a70f` (`feat(tui): Meta tags row +
Kopfblock/marker fixes`): Tags sind wieder sichtbar als **7. Meta-Feld**
(`tags:` nach `priority`, vor `created_at`; Hash-Swatch via reaktiviertem
`tagsInline`, leere Tags → `(none)`), und `enter` auf der tags-Zeile
öffnet den Tag-Picker (Kaskaden-Verhalten analog status/type/priority).
KEIN Tree-Suffix — per D01 bewusst: der Überblick läuft über die
f-Filter-Tags-Facette.

Live-Verifikation beim E8-Abschluss (bt-6ppq, 2026-07-16, echtes Binary
gegen dieses Repo in tmux): Detail-Pane von `bt-apmy` zeigt
`▷ tags:       ● to-review` als 7. Meta-Zeile — der in diesem bean
dokumentierte Fehlzustand („META hat kein Tags-Feld") besteht nicht mehr.
validation.md: US-08-Zeile auf „PASS (pending PO-Sichtprüfung)" gehoben,
Detail-Tabelle in §7 „E8-Umsetzung".

Status bleibt bewusst UNVERÄNDERT (PO-Gate, Review-Flow §5) — Tag
`to-review` gesetzt, der PO schließt nach eigener Sichtprüfung.


## Review 2026-07-16 (PO): REJECTED — Nacharbeit

PO-Sichtprüfung: "Kein tag sichtbar -> Nacharbeit erforderlich." Widerspricht dem
E8-Abschluss-Smoke (live verifiziert '▷ tags: ● to-review'). Ursachen-Analyse läuft
(Investigator). Reopened als E9-B05, umgehängt unter bt-tct9.
