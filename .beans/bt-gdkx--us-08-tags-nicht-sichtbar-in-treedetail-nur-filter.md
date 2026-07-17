---
# bt-gdkx
title: 'US-08: Tags nicht sichtbar in Tree/Detail (nur Filter-Facette)'
status: in-progress
type: bug
priority: normal
tags:
    - rejected
created_at: 2026-07-15T19:21:19Z
updated_at: 2026-07-17T08:12:21Z
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


## Klarstellung (2026-07-16, PO)

US-08 im Kern ERFÜLLT — PO sieht Tags in [1] META. Die Rejection bezog sich auf den
fehlenden tags-Eintrag im Kopfblock-Meta-Strip → getrackt als E9-B05 (bt-tct9).
Dieses bean wird mit E9-B05 gemeinsam geschlossen.


## Plan-Konkretisierung E12 (2026-07-17)

Plan: `docs/plans/v1-port/epic-E12-plan.md` §„Item 5: US-08 Tags-Sichtbarkeit
— Verifikation, kein erwarteter Fix". Reihenfolge: Rang 5 (Cluster mit
`bt-se4q`, gemeinsame Datei `view_detail_bean.go`, kein hartes blocked_by).

**Kein Fix ohne neuen Befund.** Historie zeigt die Kernanforderung bereits
ZWEIFACH gelöst: (1) D01 (Epic `bt-ntoz`, Commit 397a70f) — Tags als
Meta-Feld (`metaFields`, `view_detail_bean.go`); (2) B05 (Epic `bt-tct9`,
redefiniert 2026-07-16) — Tags zusätzlich im Kopfblock-Meta-Strip
(`detailHeaderBlock`, `view_detail_bean.go:193-227`, Zeile 219). US-04
(Kopfblock-Tags) wurde vom PO in Runde 2 explizit AKZEPTIERT (`bt-tct9`
Zeile 216). Einziger offener Rest: Q02 (`bt-tct9` Zeile 165-178,
Reviewer-Finding `bt-mtig`) — bei ~100 Spalten Split-View (`accW≈61`) füllt
die Kopfblock-Zeile bereits 59/61 Zeichen OHNE Tag-Inhalt, `truncate(
typeStatusPrio, w)` (`view_detail_bean.go:224`) schneidet die Tag-Spalte
dann hart ab. PO-Entscheid in `bt-tct9`: „bis zur Antwort keine Nacharbeit".

**Vorgehen:** NUR Bestandsaufnahme, kein Code-Change. Live-Verifikation
(tmux, `bin/bt`, Standard-Split-View-Breite) bestätigt Tags sichtbar in
META (5. Feld) UND im Kopfblock. Q02-Zusammenhang explizit referenzieren.
KEINE Status-/Tag-Änderung an diesem bean — Empfehlung an den PO: `bt-gdkx`
gemeinsam mit `bt-tct9`s Q02 schließen, sobald diese beantwortet ist (bean
sagt das bereits selbst: „wird mit E9-B05 gemeinsam geschlossen").

**Akzeptanz:**
- [ ] Live-Verifikation bestätigt: Tags sichtbar in META UND Kopfblock
      (Label mindestens, Wert bei ausreichender Breite)
- [ ] Q02-Zusammenhang (`bt-tct9` Zeile 165-178) im Verifikations-Ergebnis
      explizit referenziert, kein drittes Parallel-Tracking
- [ ] KEIN Code-Change, KEIN Status-/Tag-Wechsel an diesem bean

## Verifikation E12 Item 5 (2026-07-17)

Reine Bestandsaufnahme (kein Code-Change), per docs/plans/v1-port/epic-E12-plan.md
Item 5. Build: `command go build -o bin/bt .` (frisch, exit 0). tmux-Smoke
(Session `btgdkx$$`, echtes Binary gegen dieses Repo, Guards docs/
LESSONS-LEARNED.md E11-Sektion Nr. 1/5 beachtet). Filter (f) auf Tag
`rejected` → Tree bis bt-gdkx aufgeklappt (bt-apmy > bt-tct9 > bt-gdkx),
Detail-Pane von bt-gdkx inspiziert.

(a) META als Feld — BESTÄTIGT. Bei 100 Spalten Terminalbreite, META-Sektion
    (`[1]`) aufgeklappt zeigt 5. Feld:
    `▷ tags:     ● rejected`
    (metaFields, view_detail_bean.go:135-151 — Code-Lesung deckungsgleich
    mit Live-Render).

(b) Kopfblock-Meta-Strip-Label — BESTÄTIGT sichtbar bei ausreichender
    Breite. Bei 180 Spalten Terminalbreite (rechtes Pane ~140 Spalten):
    `type: bug          status: in-progress    prio: ·    tags: ● rejected`
    — Label UND Wert vollständig sichtbar (detailHeaderBlock,
    view_detail_bean.go:193-228, Zeile 216-219 Zusammenbau, Zeile 224
    `truncate(typeStatusPrio, w)`).

(c) Kopfblock bei Standard-/Split-Breite (100 Spalten gesamt, rechtes
    Pane ~54-61 Spalten) — Q02-Effekt BESTÄTIGT, sogar sichtbar
    verschärft: Zeile zeigt
    `type: bug          status: in-progress    prio: ·   …`
    — der `truncate()`-Schnitt (Zeile 224) trifft hier VOR dem
    `tags:`-Label selbst (nicht nur vor dem Wert wie im bt-tct9-Reviewer-
    Befund bei accW≈61 beschrieben) — je nach genauer Pane-Breite kappt
    es mal das Label, mal nur den Wert, in jedem Fall bleibt der Tag-
    Inhalt im Kopfblock bei Normalbreite nicht lesbar. META (a) bleibt
    davon unberührt — dort ist der Tag-Wert unabhängig von der
    Terminalbreite immer vollständig sichtbar (wrapText statt truncate).

## Q02-Verweis

bt-tct9, Abschnitt "Review-Finding T1/F01 (2026-07-16, Reviewer bt-mtig)
— offene PO-Frage Q02" (Zeile ~165-178 im bean-Body): Reviewer-Befund zu
genau diesem Kopfblock-Truncate-Verhalten bei ~100 Spalten Gesamtbreite,
PO-Entscheid dort "bis zur Antwort keine Nacharbeit" (Sperre). Diese
Live-Verifikation bestätigt den dort dokumentierten Zustand unverändert
und reproduziert ihn zusätzlich bei etwas schmalerer Pane-Breite (Label
selbst betroffen). Kein neues/drittes Tracking — Q02 bleibt die alleinige
offene Frage.

## Empfehlung an PO

Kernanforderung US-08 bleibt laut Live-Befund zweifach erfüllt (D01 META-
Feld + B05 Kopfblock-Label/Wert bei ausreichender Breite). Kein Fix in
diesem bean nötig oder empfohlen — bt-gdkx gemeinsam mit bt-tct9s Q02
schließen, sobald der PO dort entschieden hat (Tag-Vorrang vor Ellipse
vs. zweizeiliger Umbruch vs. akzeptierte Grenze).
