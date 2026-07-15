---
# bt-8v69
title: E3 T2 — Tag-Picker
status: completed
type: task
priority: high
created_at: 2026-07-15T00:26:06Z
updated_at: 2026-07-15T01:31:28Z
parent: bt-gzcu
blocked_by:
    - bt-dlgk
---

Ziel: Tag-Picker (`t`) — Toggle-Multi-Select über vorhandene Tags (Nutzungszähler,
sortiert Anzahl desc dann alpha, Port beans-src tagpicker.go-Sortierung) + Freitext-
Neuanlage-Zeile ("+ new tag…", eigener Text-Capture-Sub-Modus wie m.searchInput).
Enter bestätigt (Diff gegen Original-Tags), esc verwirft (Port beans-src
blockingpicker.go Enter=confirm/esc=cancel-Konvention, NICHT box_filter_facets.go's
enter=close-ohne-Wirkung — hier wird echt mutiert).

DESIGN-ENTSCHEIDUNG (Plan »Task 2«, verbindlich): der Diff wird als EIN
`beans update` mit kombinierten --tag/--remove-tag-Flags ausgeführt — neue
mutations.go-Funktion SetTags(id string, add, remove []string, etag string).
N Einzel-Mutationen (AddTag/RemoveTag nacheinander) auf EINEM ETag wären eine
Konflikt-Kaskade (die erste gewinnt, jede weitere läuft in ErrConflict).
AddTag/RemoveTag bleiben für Einzelfälle bestehen.

Plan: docs/plans/v1-port/epic-E3-plan.md »Task 2«.

## Akzeptanz
- [x] internal/data/mutations.go: SetTags (EIN update, --tag/--remove-tag kombiniert),
      Test TestSetTagsAddsAndRemovesInOneCall (client_mut_test.go-Muster)
- [x] internal/data/tags.go: ValidTagName gegen `^[a-z][a-z0-9]*(-[a-z0-9]+)*$`
      (Tag-Regex aus Epic-Body bt-gzcu), Tests inkl. Negativ-Fälle
- [x] box_picker_tag.go: Menü mit Nutzungszählern (collectTagCounts über idx.ByID,
      deterministisch sortiert count desc/alpha — KEINE Map-Walk-Ordnung, Lehre aus
      tagFilterOptions-ERRATUM), space/x toggelt Pending-State, n öffnet
      Neuanlage-Input (invalider Name -> Inline-Fehler, kein Submit)
- [x] enter diff't Pending gegen Original -> EIN mutateCmd(SetTags), ETag frisch via
      beanETag; keine Änderungen -> kein Cmd; esc verwirft alles
- [x] go test ./... grün, gofmt/vet leer

## Findings / Umsetzung

data/tags.go (neu): `ValidTagName`/`tagNameRe` (`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`),
`TestValidTagName` deckt die exakte Plan-Tabelle ab (valid/invalid). Der Regex allein
löst alle Negativ-Fälle korrekt auf (Doppel-Hyphen, führende Ziffer/Hyphen, Großbuchstabe,
Unicode, Leerzeichen) — kein Zusatzcode nötig.

mutations.go: `SetTags(id, add, remove []string, etag string)` baut EIN `update`-
Aufruf mit wiederholten `--tag`/`--remove-tag` (beide `stringArray`, gegen
`beans update --help` verifiziert). `TestSetTagsAddsAndRemovesInOneCall`
(client_mut_test.go) nutzt `tt-task` (nicht `tt-epic`) als Ziel: `testrepo_test.go`s
eigener Kommentar dokumentiert eine echte beans-0.4.2-Divergenz zwischen dem von
`list`/`show` gemeldeten ETag und dem intern von `update --if-match` erwarteten Wert
für Beans mit einem `tags:`-Frontmatter-Block, solange die Datei noch nie von beans
selbst neu geschrieben wurde — ein AddTag-„Seed"-Aufruf VOR dem eigentlichen
SetTags-Test etabliert den `tags:`-Block einmalig sauber (dieselbe Umgehung, die die
Fixture-Doku bereits für andere Mutationstests vorschreibt).

box_picker_tag.go (neu): `tagCount{tag,count}`, `collectTagCounts(idx)` (Map-Walk über
`idx.ByID` liefert die Zwischen-Counts in undefinierter Reihenfolge, aber der
End-Sortierschlüssel count-desc/tag-alpha ist auf einzigartigen Tag-Namen VOLLSTÄNDIG
deterministisch — anders als `tagFilterOptions` (die ganze `*Bean`-Werte sortiert, die
auf ALLEN Kriterien binden können) braucht dieser Fall keinen ID-sortierten
Baseline-Pass). `openTagPicker` seedet `tagOriginal`/`tagPending` als zwei
UNABHÄNGIGE, frisch gebaute Maps aus dem fokussierten Bean; `toggleTagPending` klont
trotzdem per `cloneBoolMap` vor jedem Schreiben (I01) — die Karte lebt über die ganze,
potenziell mehrschrittige Overlay-Session, nicht nur einen Tastendruck, dieselbe
Aliasing-Gefahr wie `toggleFacet`. Freitext-Neuanlage (`n`) spiegelt `keySearchInput`s
Capture-Konvention (ein persistentes `textinput.Model`, bei Öffnen zurückgesetzt +
fokussiert); ungültiger Name setzt `tagInputErr` und hält den Input offen (kein
Submit), gültiger Name dedupt gegen `tagItems` (kein Doppel-Eintrag bei erneuter
Eingabe eines schon gelisteten Tags), hängt sonst eine neue count=0-Zeile an und
sortiert neu ein, setzt `tagPending[name]=true`, schließt den Input.

`applyTagPickerDiff`: Diff gegen `tagOriginal` (add = pending\original, remove =
original\pending), `sort.Strings` auf beiden Listen (Determinismus gegen
Map-Iterations-Reihenfolge in Test/CLI-Argumenten), keine Änderungen -> kein Cmd,
`beanETag`-Vanished-Guard analog `applyValueMenuSelection` (Design d) -> Overlay zu +
Statuszeile, kein Cmd. GENAU EIN `mutateCmd(SetTags(...))` (ERRATUM-Fix ggü. der
Plan-eigenen frühen Step-1-Skizze "2 Mutationen via tea.Batch" — durch den
"Design-Nachtrag" im Plan selbst superseded; per Test verifiziert, dass `cmd()` direkt
zu `mutationDoneMsg` auflöst, nicht zu `tea.BatchMsg`).

Wiring: `keyNodeAction` (update.go) — `keys.TagAssign` -> `openTagPicker()`, Guard
(kein fokussiertes Bean) bereits von T1 vorhanden. `keyOverlay` bekommt den Case
`overlayTagPicker -> keyTagPicker`. `composeOverlays` (view_browse_repo.go) bekommt
den Case `overlayTagPicker -> tagPickerBox()`. Keine neuen keymap.go-Bindings (`t`/
Toggle space·x existierten bereits, Plan-Vorgabe eingehalten) — `TestHelpGroupsCover
EveryBindingExactlyOnce` bleibt unberührt grün.

Cheap Follow-up (T1-Review-Nachtrag, im Task-Auftrag benannt):
`TestOverlayCaptureSwallowsQuitKeysWhileTagPickerOpen` pinnt handleKeys Capture-
Reihenfolge (`m.overlay != overlayNone` VOR dem globalen ctrl+c/q-Switch) — q/ctrl+c
werden bei offenem Overlay verschluckt, kein `tea.Quit`. Overlay-agnostisch geprüft
(gegen den Tag-Picker, da der hier gebaut wird), deckt aber die generische
handleKey-Reihenfolge ab.

Tests (neu): `data/tags_test.go` (1), `data/client_mut_test.go`
(TestSetTagsAddsAndRemovesInOneCall), `tui/box_picker_tag_test.go` (11 Tests:
collectTagCounts-Sortierung, openTagPicker-Seed+Unabhängigkeit, Toggle-Original-
unberührt, Enter-Diff-EIN-SetTags-Call, Enter-ohne-Änderung-kein-Cmd, Esc-verwirft,
Neuanlage-Regex-Fehler, Neuanlage-Pending-Item, Vanished-Target-Guard,
Capture-Order-Guard). `go test ./... -race -count=1` grün (alle Pakete), `gofmt -l .`
leer, `go vet ./...` leer, `go build ./...` ok, Golden-Tests 2x grün (deterministisch).

Manueller Smoke (tmux, SCRATCH-Repo `bt-e3t2-smoke4-*` — NICHT dieses Repos `.beans/`):
zwei Beans angelegt, `t` auf Bean A geöffnet, `space` togglet `urgent` an, `n` +
Freitext `greenfield-smoke` + enter legt Pending-Zeile an, äußeres `enter` bestätigt
-> `.md` zeigt BEIDE Tags in EINEM Schreibvorgang (ein `updated_at`-Bump). Picker
erneut geöffnet: Counts korrekt aktualisiert (urgent jetzt 2), `urgent` togglet aus,
dritter Tag `final-check` per Freitext hinzu, enter -> `.md` zeigt exakt den
kombinierten Diff (`urgent` entfernt, `greenfield-smoke` behalten, `final-check` neu)
in EINEM weiteren Schreibvorgang — belegt die ONE-`SetTags`-Call-Design-Entscheidung
Ende-zu-Ende gegen den echten `beans`-Binary.

Finding (B, medium, außerhalb dieses Tasks' Scope): der lokal installierte
`beans 0.4.2`-Client hat ein reproduzierbares ETag-Divergenz-Bug, das über die
bereits in `testrepo_test.go` dokumentierte "hand-authored tags:"-Variante
hinausgeht — JEDER frisch per `beans create` angelegte Bean meldet über `list`/`show`
ein ETag, das vom intern durch `update --if-match` erwarteten Wert abweicht, bis EIN
erfolgreiches Update die Datei neu geschrieben hat (danach bleibt es konsistent,
verifiziert per direktem CLI-Reproduce: create-Response-ETag ≠ sofortiges
show/list-ETag ≠ was update intern verlangt — dieselbe "current"-ETag-Konstante taucht
über mehrere Fehlversuche stabil wieder auf). Betrifft JEDE bt-Mutation (Status/Type/
Priority/Tags/…) auf einem frisch angelegten, noch nie mutierten Bean — nicht
Tag-Picker-spezifisch. Der Smoke-Workaround (ein Warm-up-Update über das rohe CLI vor
dem ersten TUI-Zugriff) umgeht es zuverlässig; die automatisierten Go-Tests sind nicht
betroffen, da `newTestRepo`s Fixture-Beans direkt als Dateien geschrieben werden (nie
über `beans create`) bzw. `TestSetTagsAddsAndRemovesInOneCall` denselben Seed-Workaround
bereits nutzt. Kein Fix in diesem Repo möglich (Upstream-Binary) — Empfehlung: Notiz im
Epic-Body bt-gzcu oder eigenes Ticket für ein Upstream-Issue, damit T3/T4/T5/T6 nicht
erneut denselben Smoke-Fallstrick treten.

Notizen für T3 (Parent-/Blocking-Picker, bt-p1uz): das Pending-Diff-Muster
(tagOriginal/tagPending, cloneBoolMap-Toggle, Enter-Diff-gegen-Original, Esc-verwirft)
ist jetzt 1x real gebaut (nicht nur im Plan skizziert) — T3s Blocking-Picker kann
`box_picker_tag.go`s `toggleTagPending`/`applyTagPickerDiff`-Form direkt als Vorlage
kopieren (gleiche Struktur, andere Datenquelle/Mutation: `SetBlocking` statt
`SetTags`). Der Parent-Picker (Single-Select, kein Pending-Diff) braucht das Muster
nicht. WICHTIG: beim eigenen Smoke-Test IMMER erst ein Warm-up-Update über das rohe
CLI fahren (s. Finding oben), sonst schlägt der allererste TUI-Mutationsversuch auf
einem frisch angelegten Bean scheinbar grundlos fehl (Konflikt-Overlay schließt sich,
Statuszeile zeigt "Konflikt" nur für einen einzigen Render-Frame, der unmittelbar
folgende `applyMutationResult`-Reload überschreibt `m.err` sofort wieder über
`applyLoaded`s unconditional `m.err = ""` bei Erfolg des Reloads selbst — separates,
hier NICHT behobenes Findings-Detail, siehe Bug-Hinweis oben zur Statuszeilen-
Sichtbarkeit von Konflikt-Meldungen).

Commit: siehe Refs unten (main, direkt — Worktree-Weiche main-direkt für dieses Repo).
