---
# bt-w9o8
title: E7 T3 — UI-Sprache Englisch + Command-Center-Schema (PF-7, PF-8)
status: completed
type: task
priority: high
created_at: 2026-07-15T14:26:51Z
updated_at: 2026-07-15T16:07:32Z
parent: bt-heg9
blocked_by:
    - bt-2af1
---

Details/Steps/Akzeptanz: docs/plans/v1-port/epic-E7-plan.md Task 3. Vollstaendiger Deutsch-zu-Englisch-Sweep (13 Dateien, Tabelle im Plan) + Command-Center verb-entity-Schema. 4 Accordion-Titel bewusst an T4 abgegeben.

## Summary

PF-7 (UI-Sprache Englisch) + PF-8 (Command-Center `verb entity`-Schema, ohne
Doppelpunkt) vollstaendig umgesetzt. `overlay_palette.go`s 14 Aktions-Labels
exakt nach der PF-8-Tabelle (design-spec.md §15 / epic-E7-plan.md) ersetzt:
`set status` · `set tags` · `set parent` · `set blocking` · `set title` ·
`delete bean` · `create bean` · `go to backlog` · `go to browse` ·
`filter facets` · `search beans` · `reload data` · `go to repo picker` ·
`go to settings` (Cockpit-Eintrag bereits von T1 entfernt). Alle
nutzerseitigen deutschen Strings in `internal/`/`cmd/` (Toasts, Fehlertexte,
Overlay-/Picker-Titel, Footer-Hints, Placeholder-Texte, Cobra
`--help`-Ausgabe) auf Englisch umgestellt — 22 Dateien Produktionscode
(21 `internal/tui`, 1 `internal/data`, 2 `cmd`) + 1 Golden.

Accordion-Sektionstitel `Meta`/`Body`/`Beziehungen`/`Historie`
(`view_detail_bean.go`) BEWUSST unveraendert — Plan weist das explizit T4
zu (baut `beanSections`/`view_detail_bean.go` ohnehin um zu
`META`/`BODY`/`RELATIONS`/`HISTORY`, s. Notes for T4 unten).

Eigener Grep-Sweep (Pflicht, "Planner-Tabelle misstrauen") ueber die
13-Datei-Tabelle hinaus fand 9 zusaetzliche Fundstellen, die die
Planner-Tabelle nicht listete — s. Deviations.

## Test-Output

RED→GREEN: Test-Assertions in 17 `_test.go`-Dateien auf die neuen
englischen Strings umgeschrieben (Palette-Labels, Delete-Confirm-Grammatik
Singular/Plural, Konflikt-/Copied-/Fassung-Toasts, orphan-Bucket-Referenzen,
Cobra-`Use`-String) BEVOR die Produktionsstrings geaendert wurden.

`command go build -o bin/bt .` → sauber.
`command go vet ./...` → leer.
`gofmt -l .` → leer.
`command go test ./...` (voller Lauf, OHNE `-short`) → PASS, 136.4s
(`internal/tui`, inkl. der 7 langsamen huh-Drive-Tests). Zweiter voller
Lauf nach den 3 Nachzueglern (Yank-/Repo-Wechsel-Toasts, Fassung-Text,
offen/gesamt) erneut PASS, 136.1s.
`command go test ./internal/tui/... -run 'TestTreeGolden|TestBacklogGolden|
TestChromeGolden' -count=2` → PASS (beide Durchlaeufe stabil).
`command go test ./internal/tui/... -run 'TestPalFilteredActionsFuzzyFiltered|
TestPaletteActions|TestDispatchPalette'` → PASS, alle 7 Tests (Fuzzy-Match-
Regression PF-8 explizit gruen, Matching-Logik unveraendert, nur Label-Text
getauscht).

## Golden-Diffs

Nur `internal/tui/testdata/tree.golden` geaendert (1 Zeile,
byte-identische Spaltenbreite): `(verwaist)` → `(orphaned)` — beide 8
Zeichen lang, kein Layout-Shift, nichts geclippt. `backlog.golden` und
`chrome.golden` unveraendert (enthielten keine der betroffenen Strings).
Regeneriert via `command go test ./internal/tui/... -run
'TestTreeGolden|TestBacklogGolden|TestChromeGolden' -update`, danach
`-count=2` zur Stabilitaetsprobe (s. Test-Output).

## Sweep-Beleg

Abschluss-Sweep (case-insensitive, Umlaute + erweiterte Stichwortliste,
automatisiert nach Kommentar-/Code-Zeile klassifiziert):
verbleibende deutsche Treffer in USER-FACING Code-Strings (nicht-Test,
nicht-Kommentar) = **0**. 20 verbleibende Treffer sind ausschliesslich
`//`-Kommentare (laut Vorgabe zulaessig — "Code-Kommentare duerfen deutsch
bleiben, nur USER-FACING zaehlt"), u.a. Dateikopf-/Design-Notiz-Zitate wie
"Kosten/Latenz-Abwaegung dokumentieren", "windowAround bei Ueberlaenge",
"kein PO-Facet" — keiner davon ist ein tatsaechliches Go-String-Literal im
Ausfuehrungspfad.

Zusaetzlich vollstaendige Extraktion JEDES nicht-trivialen String-Literals
(Leerzeichen/Satzzeichen enthaltend) aus allen Nicht-Test-`.go`-Dateien in
`internal/`+`cmd/` manuell durchgesehen (nicht nur Keyword-gefiltert) —
bestaetigt: alle verbleibenden Literale sind Englisch, Formatverben (`%w`,
`%d/%d`), Glyphen, Farbcodes oder reine Bezeichner.

## Smoke

tmux (dieses Repo als Test-Datenquelle), Sequenz vollstaendig durchlaufen:
Tree (Breadcrumb "> repo: Browse" englisch) → `tab` Detail-Focus (Meta-
Feldliste "Status:"/"Type:"/"Priority:"/"Tags:" englisch) → `esc` → `b`
Backlog (Breadcrumb "Backlog") → `f` Filter-Menue ("Filter"-Titel,
"space/x:toggle X:clear enter/esc/f:done", "Archive"-Sektion + "Show
archived"-Zeile, kein Clipping) → `esc` → `ctrl+k` Command-Center (14
Eintraege exakt nach PF-8-Schema gelistet, s. Summary) → `esc` → `c`
Create-Form ("New bean"-Titel) → `esc` → `d` Delete-Confirm ("Delete
feature", "Irreversible.", "enter: delete permanently esc/n: cancel",
NICHT bestaetigt) → `esc` → `a` Parent-Picker ("Assign parent"-Titel,
"enter:set esc:cancel", "(No parent)"-Zeile) → `esc` → `t` Tag-Picker
("Tags"-Titel, "space/x:toggle n:new tag enter:save esc:discard") → `n`
neuer-Tag-Prompt ("New tag"-Titel, "enter:create esc:cancel", Placeholder
"new tag (a-z0-9, hyphen-separated)") → 2x `esc` → `B` Blocking-Picker
("Blocking"-Titel, "space/x:toggle enter:save esc:discard") → `esc` → `p`
Lobby ("Repo-Picker"-Untertitel [Eigenname, unveraendert], "Filter repos
(path)"-Placeholder, Spaltenkopf "open/total", Leerzustand "(no repos in
config.yaml -- ctrl+k -> settings)") → `esc` zurueck zu Browse → `ctrl+k` →
Query "settings" getippt (Fuzzy-Match traf korrekt NUR "go to settings") →
`enter` → Settings-Form ("Settings"-Titel, Feld "repos", Beschreibung "one
beans-repo path per line"). PASS auf jedem der 12 Screens/Overlays — kein
verbleibendes Deutsch gesichtet (bewusste Ausnahme: Accordion-Titel
"Beziehungen"/"Historie", T4-Scope).

## Deviations

- **9 Fundstellen jenseits der Planner-13-Datei-Tabelle** (eigener Grep-Sweep
  Pflicht lt. Auftrag — "Planner fand selbst 2 Luecken in seinem ersten
  Sweep"): `update.go:443` Editor-Aenderung-verworfen-Toast; `update.go:189`
  "Repo-Wechsel fehlgeschlagen" (Toast); `update.go:585` "Yank
  fehlgeschlagen" (Toast) — beide nur ueber case-insensitive Sweep
  gefunden, `fehlgeschlagen` (klein-f) war in der ersten Stichwortliste nur
  als Grossbuchstaben-"Fehler"-Substring gelistet; `update.go:272-273`
  ETag-Konflikt-Recovery-Text "deine Fassung"/"Fassung gesichert" (+
  zugehoerige `editor_test.go`-Marker-Konstante); `view_browse_repo.go:417`
  Tree-Orphan-Bucket-Label `"(verwaist)"` → `"(orphaned)"` (+ 7
  Kommentar-Querverweise in `types.go`/`box_confirm_delete.go`/
  `data/mutations.go`/3x `update_test.go`/`search_test.go`/
  `client_mut_test.go` mitgezogen); `types.go:530` Tag-Picker-Freitext-
  Placeholder "neuer Tag (a-z0-9, Bindestrich-getrennt)" (kein Umlaut, kein
  Stichwort-Treffer — reiner Vollextraktions-Fund);
  `box_picker_blocking.go:187`/`box_picker_tag.go:305` Footer-Hints
  "enter:speichern esc:verwerfen"; `box_picker_tag.go:334` Modal-Titel
  "Neuer Tag"; `view_lobby.go:170` Lobby-Spaltenkopf "offen/gesamt" (spaet
  im tmux-Smoke selbst entdeckt, nicht per Grep — zeigt Grenzen von
  Text-Sweeps bei kurzen Fachbegriffen ohne Umlaut/Stichwort-Ueberschneidung).
  Alle 9 uebersetzt, Kommentar-Querverweise wo direkt zitierend mitgezogen,
  betroffene Tests nachgezogen.
- `cmd/root.go`/`cmd/tui.go` (Cobra `Use`/`Short`) waren NICHT Teil der
  Planner-Tabelle (die deckte nur `internal/tui/*.go` ab) — laut
  Sweep-Anweisung im Auftrag ("über internal/ cmd/") trotzdem
  mitgenommen: `bt [repo-pfad]`→`bt [repo-path]`,
  `bt — PO-Cockpit-TUI für beans-Repos`→`bt — PO cockpit TUI for beans
  repos`, `tui [repo-pfad]`→`tui [repo-path]`, `Startet die TUI
  explizit`→`Starts the TUI explicitly`.
- `design-spec.md` §15 PF-8-Tabelle enthaelt noch eine stale Zeile
  `go to: review cockpit` → `go to review cockpit` (Zitat der URSPRUENGLICHEN
  Remap-Absicht vor T1s Removal) — NICHT angefasst: Dokument-Historie
  (Nachtrag-Log), kein Code, kein Akzeptanzkriterium dieses Tasks.
- `overlay_palette.go`-Dateikopf-Kommentar leicht praezisiert
  ("verb: label"→"verb entity"-Beschreibung des NEUEN Schemas), da er die
  eigene Namenskonvention dokumentiert, die dieser Task aendert — sonst
  ausschliesslich literale Cross-Referenz-Zitate in Kommentaren korrigiert
  (z.B. `box_form_settings.go`s Verweis auf die umbenannte Palette-Aktion),
  keine freie Uebersetzung deutscher Prosa-Kommentare (bleibt lt. Vorgabe
  erlaubt).
- `fuzzy_test.go`s generische `fuzzyMatch`-Unit-Test-Faelle (`"create:
  bean"` als Beispieltext) bewusst NICHT angefasst — reine Algorithmus-
  Testdaten ohne Bezug zu den tatsaechlichen Palette-Labels, kein
  UI-String.

## Notes for T4 (Meta-Layout)

- `view_detail_bean.go:23-27` (`beanSections`) traegt aktuell noch
  `"Meta"`/`"Body"`/`"Beziehungen"`/`"Historie"` — von diesem Task bewusst
  unangetastet gelassen (design-spec.md §15 PF-7 weist das explizit T4 zu,
  da T4 dieselbe Funktion ohnehin fuer PF-1/PF-3/PF-4/PF-12 umbaut). Ziel
  laut design-spec: `"META"`/`"BODY"`/`"RELATIONS"`/`"HISTORY"`
  (Grossschreibung + Uebersetzung in einem Schritt).
- `internal/tui/testdata/tree.golden` UND jeder andere Golden, der die
  Detail-Accordion rendert, enthaelt aktuell noch `Beziehungen`/`Historie`
  (ANSI-verpackt, z.B. `...mBeziehungen[0m`) — T4s eigener Golden-Regen
  wird diese Zeilen mitziehen; keine Ueberschneidung mit den in diesem
  Task bereits geaenderten Zeilen.
- Keine weiteren Cockpit-Reste im Weg (T1 bereits vollstaendig).

Refs: bt-w9o8, Commit 6dbba35
