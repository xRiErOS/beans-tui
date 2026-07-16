---
# bt-604w
title: T3 — Tag-Definition anlegen (Create)
status: completed
type: task
priority: normal
created_at: 2026-07-16T15:44:30Z
updated_at: 2026-07-16T18:40:00Z
parent: bt-362n
blocked_by:
    - bt-r92i
---

T3 — Tag-Definition anlegen (Create + geteilter Input-Submodus) (Epic
`bt-362n`, D11, D14). `blocked_by` T2 (`view_tag_management.go`/
`viewTagManagement` müssen existieren). Führt den page-lokalen
Freitext-Input-Submodus EIN — T5 (Rename) baut darauf auf.

## Ziel

Auf der Tag-Management-Page eine neue Tag-Definition anlegen: Freitext-
Eingabe, validiert gegen `data.ValidTagName`, Dedupe gegen die Registry,
speichert via `SaveTagDefs` (T1), aktualisiert die Zeilenliste sofort.
Berührt KEIN Bean (D11).

## Betroffene Dateien/Symbole

- `internal/tui/types.go` (`model`): neue Felder `tagMgmtInputActive
  bool`, `tagMgmtInputMode string` (`"create"` in diesem Task, `"rename"`
  kommt in T5), `tagMgmtInput textinput.Model`, `tagMgmtInputTarget
  string` (leer bei Create, T5 befüllt es mit dem alten Namen).
  `newModel()` konstruiert `tagMgmtInput` EINMAL (mirrort `tagInput`s
  eigene Konstruktion, Placeholder „new tag (a-z0-9,
  hyphen-separated)").
- `internal/tui/view_tag_management.go`:
  - `func (m model) openTagMgmtInput(mode, prefill string) (tea.Model,
    tea.Cmd)` — setzt `tagMgmtInputActive = true`, `tagMgmtInputMode =
    mode`, `tagMgmtInput.SetValue(prefill)` + `.Focus()`, gibt
    `textinput.Blink` zurück (mirrort `openTagInput`,
    `box_picker_tag.go`).
  - `func (m model) keyTagMgmtInput(msg tea.KeyMsg) (tea.Model, tea.Cmd)`
    — mirrort `keyTagInput`s Struktur 1:1: `esc` verwirft (nur den
    Input-Submodus, nicht die Page); `enter` validiert
    (`data.ValidTagName` + Dedupe-Check gegen die aktuelle
    `m.tagMgmtRows`-Namensmenge, EXCL. `tagMgmtInputTarget` bei Rename —
    in DIESEM Task, Create, ist `tagMgmtInputTarget == ""`, Dedupe prüft
    also gegen ALLE vorhandenen Namen); bei Fehler: `tagMgmtInputErr`
    gesetzt (neues `model`-Feld `tagMgmtInputErr string`), Input bleibt
    offen; bei Erfolg: `switch tagMgmtInputMode { case "create": …
    dispatcht `saveTagDefsCmd` (neu, s.u.) mit den ERWEITERTEN Defs
    (`data.AddTagDefName`) }`.
  - `func saveTagDefsCmd(c *data.Client, defs []string) tea.Cmd` — EIN
    async Cmd (`c.SaveTagDefs` ist ein lokaler Datei-Write, aber
    KONSISTENZ mit jedem anderen state-ändernden Aufruf verlangt einen
    Cmd, kein direkter Call im Update-Pfad — mirrort `mutateCmd`s
    Bauform, eigener Msg-Typ `tagDefsSavedMsg{err error}` in
    `messages.go`).
  - `func (m model) applyTagDefsSaved(msg tagDefsSavedMsg) (tea.Model,
    tea.Cmd)` (`update.go`): bei Fehler Toast (mirrort
    `applyMutationResult`s Fehler-Zweig, ohne den unbedingten
    `loadCmd`-Reload — hier gibt es kein `m.idx` zu invalidieren, nur
    die Registry); bei Erfolg: `tagMgmtInputActive = false`,
    `tagMgmtRows` neu aufgebaut (`tagRegistryRows(m.idx, neueDefs)`),
    Cursor auf die neue/umbenannte Zeile gesetzt (Name-basierte Suche in
    `tagMgmtRows`, mirrort `applyLoaded`s Cursor-Wiederfindungs-Muster).
- `internal/tui/view_tag_management.go` (`tagManagementLocalBindings`):
  T3 hängt `keys.NewTag` an (reuse der bestehenden Bindung — gleiche
  Bedeutung „neuer Tag", nur neuer Ort, mirrort wie `keys.Enter`/
  `keys.Back` bereits quer durch viele `*LocalBindings()`-Funktionen
  wiederverwendet werden, D14).
- `internal/tui/keymap.go` (`helpGroups`): KEINE Änderung nötig
  (`keys.NewTag` ist bereits gelistet, T3 fügt keine neue
  `keybind.Binding` hinzu).
- `internal/tui/update.go` (`keyTagManagement`, aus T2): am Anfang
  `if m.tagMgmtInputActive { return m.keyTagMgmtInput(msg) }` (full
  capture innerhalb des full-capture Views, mirrort `keyTagPicker`s
  eigenen `if m.tagInputActive` Vorrang-Check), dann neuer Case
  `keybind.Matches(msg, keys.NewTag): return m.openTagMgmtInput("create", "")`.

## TDD (RED zuerst)

```go
func TestKeyTagMgmtInputRejectsInvalidName(t *testing.T) {
    m := newModel(nil, "")
    m.view = viewTagManagement
    m, _ = m.openTagMgmtInput("create", "")
    m.tagMgmtInput.SetValue("Not Valid!")
    nm, _ := m.keyTagMgmtInput(tea.KeyMsg{Type: tea.KeyEnter})
    got := nm.(model)
    if !got.tagMgmtInputActive || got.tagMgmtInputErr == "" {
        t.Fatalf("want input to stay open with an error, got active=%v err=%q",
            got.tagMgmtInputActive, got.tagMgmtInputErr)
    }
}

func TestKeyTagMgmtInputRejectsDuplicateAgainstExistingRows(t *testing.T) {
    m := newModel(nil, "")
    m.view = viewTagManagement
    m.tagMgmtRows = []tagRegistryRow{{name: "already-there", defined: true}}
    m, _ = m.openTagMgmtInput("create", "")
    m.tagMgmtInput.SetValue("already-there")
    nm, _ := m.keyTagMgmtInput(tea.KeyMsg{Type: tea.KeyEnter})
    if got := nm.(model); !got.tagMgmtInputActive || got.tagMgmtInputErr == "" {
        t.Fatalf("want rejected duplicate, got %+v", got)
    }
}
```

Weitere Pflicht-Tests: `esc` im Input verwirft nur den Submodus (Page
bleibt offen, `tagMgmtRows` unverändert); erfolgreicher Submit ruft
`saveTagDefsCmd` mit `AddTagDefName`-erweiterten Defs auf (Cmd-Erzeugung
per Interception/Fake-Client testen, mirrort bestehende `mutateCmd`-Tests
im Repo als Vorlage).

## Golden-Strategie

GEGENBELEG (Tree/Backlog/Chrome unverändert — reine Overlay-artige
Submodus-Logik innerhalb der neuen Page, keine der drei Basis-Views
berührt). Eigene Golden-Suite (falls T2 eine angelegt hat): T3 erweitert
sie optional um einen „Input aktiv"-Snapshot.

## tmux-Smoke (120 UND 80 Spalten)

Page öffnen → `n` → Freitext „my-new-tag" → enter → Zeile erscheint
sofort in der Liste (Gruppe „Definiert", Count 0) → `git status` zeigt
`.beans-tags.yml` als neue/geänderte Datei (im laufenden Repo — NACH dem
Smoke-Test wieder auf den Ausgangsstand zurücksetzen, `git checkout --
.beans-tags.yml` bzw. Datei löschen falls sie neu war, s. globale Regel
„Temporäre Testdateien restlos entfernen"). Ungültigen Namen probieren
(z. B. „Not Valid") → Inline-Fehler, Input bleibt offen. Bei 80 Spalten:
Inline-Fehlertext darf nicht truncaten ohne `…`.

## Akzeptanz-Checkliste

- [x] `n` öffnet Freitext-Input
- [x] gültiger Name legt Definition an + Liste aktualisiert sofort
- [x] ungültiger Name → Inline-Fehler, kein Submit
- [x] Duplikat gegen bestehende Zeilen abgewiesen (defined UND frei)
- [x] Create berührt KEIN Bean (D11, Regressionstest: `m.idx` unverändert nach Create)
- [x] `esc` im Input verwirft nur den Submodus
- [x] Goldens Gegenbeleg grün
- [x] tmux-Smoke 120+80 belegt, Testdatei danach entfernt
- [x] voller Lauf grün
- [x] Commit `feat(tui): E10 Tag-Definition anlegen (Create)`

## PRELUDE aus T6-Review (2026-07-16, F01, low — Refactor vor Task-Start)

Als erster eigener Commit: `tagRegistryRows` (view_tag_management.go:64-106) unterhält eine eigene, zu `collectTagCounts` (box_picker_tag.go, seit T6 mit defined-Primärschlüssel) strukturell fast identische Zähl-Schleife — echte Duplikation mit Drift-Risiko. Auf `collectTagCounts` umstellen (war in bt-r92i 'D10-ERRATUM-Notiz für T6' und bt-pqq3 'Notes for T3' bereits vorgezeichnet; T2/T6 liefen mit disjunktem Datei-Scope). Verhalten der Page darf sich NICHT ändern (bestehende T2-Tests bleiben grün = Beleg); D09-Sortierung der Page (definiert alpha, frei count-desc) vs. Picker-Sortierung (defined-first global) beachten — die Page-Gruppierung bleibt Page-Logik, nur die ZÄHLUNG wird geteilt.

(T6-F02, nur Doku-Notiz: T6-bean-Zielabschnitt nennt Glyph '●', implementiert ist '✓' aus T2 — sachlich gewollt, Summary korrekt.)

DONE (2026-07-16): erledigt als eigener Commit `refactor(tui): tagRegistryRows->collectTagCounts` (48 Zeichen, VOR dem T3-Feature-Commit) — `tagRegistryRows` baut jetzt einen `defSet map[string]bool` aus `defs`, ruft `collectTagCounts(idx, defSet)` (T6, geteilte Zähl-/Union-Logik) und partitioniert das bereits sortierte Ergebnis in `definedRows`/`freeRows`; NUR `definedRows` wird lokal nach Name re-sortiert (D09 verlangt für die Definiert-Gruppe reine Alpha-Sortierung, `collectTagCounts` sortiert global defined-first/count-desc/alpha — genau an dieser einen Stelle divergiert die Page-Gruppierung vom geteilten Picker-Sort). `freeRows` übernimmt die von `collectTagCounts` bereits count-desc/alpha sortierte Relativ-Reihenfolge unverändert (stabile Partition). Beleg: alle 8 bestehenden `TagRegistryRows*`-Tests sowie der komplette T2/T6-Korpus (46 Tests) blieben GRÜN, KEINE Assertion musste angepasst werden.

## Summary

Liefert D11 (Create) + D14 (geteilter Freitext-Input-Submodus, hier EINGEFÜHRT, T5/Rename wird ihn wiederverwenden). Neue `model`-Felder (`types.go`): `tagMgmtInputActive bool`/`tagMgmtInputMode string`/`tagMgmtInput textinput.Model`/`tagMgmtInputTarget string`/`tagMgmtInputErr string` — `newModel()` konstruiert `tagMgmtInput` einmal (Placeholder „new tag (a-z0-9, hyphen-separated)", mirrort `tagInput`). `internal/tui/view_tag_management.go` (neu, D14): `openTagMgmtInput(mode, prefill string)` (setzt Active/Mode/Target/Focus, mirrort `openTagInput`), `keyTagMgmtInput(msg)` (esc verwirft NUR den Submodus; enter validiert `data.ValidTagName` + dedupet gegen ALLE aktuellen `m.tagMgmtRows`-Namen — defined UND frei, EXCL. `tagMgmtInputTarget`, bean-Wortlaut „Dedupe prüft also gegen ALLE vorhandenen Namen" wörtlich umgesetzt; bei Erfolg im `"create"`-Modus dispatcht `saveTagDefsCmd` mit `data.AddTagDefName(definedTagNames(m.tagMgmtRows), name)` — Sub-Modus bleibt bis zur BESTÄTIGTEN Ersparnis offen, schließt NICHT optimistisch), `definedTagNames(rows)` (Helfer: extrahiert die defined-Teilmenge als `[]string` für die T1-Helfer), `tagMgmtInputRows()` (rendert Hint+Feld+Fehler INLINE in derselben Pane — D06 „lebt INNERHALB der Full-Capture-Page", KEIN floating Overlay wie `tagInputBox`). `keyTagManagement` (D06) prüft `tagMgmtInputActive` ZUERST (mirrort `keyTagPicker`s `tagInputActive`-Vorrang) und bekommt einen neuen `keys.NewTag`-Case; `tagManagementLocalBindings` wächst um `keys.NewTag` (Position: vor `Back`, Implementer-Entscheidung, mirrort `tagPickerLocalBindings`). `viewTagManagement` rendert bei aktivem Submodus `tagMgmtInputRows()` statt der Zeilenliste in DERSELBEN `renderPane`-Zelle (Fehlertext profitiert automatisch von `renderPane`s `truncate(w-2)`-Budget — kein Extra-Code für die 80-Spalten-„…"-Vorgabe nötig). Neu in `messages.go`: `tagDefsSavedMsg{err error}` + `saveTagDefsCmd(c, defs)` (EIGENER Msg-Typ, NICHT `mutationDoneMsg` — Registry-Write berührt kein Bean, kein `m.idx`-Reload nötig, mirrort `createDoneMsg`s Präzedenzfall „eigener Msg-Typ wenn der Tail wirklich divergiert"). Neu in `update.go`: `case tagDefsSavedMsg` im `Update()`-Dispatcher + `applyTagDefsSaved(msg)` (Fehler: Toast, Submodus bleibt offen für Retry, KEIN `loadCmd`; Erfolg: Submodus schließt, `tagMgmtRows` wird aus einem FRISCHEN `LoadTagDefs()` neu aufgebaut [D02/D03 — Disk ist Source of Truth, nicht der lokal berechnete `defs`-Slice], Cursor per Namenssuche auf die neue Zeile gesetzt).

## Test-Output (RED→GREEN wörtlich)

RED (`command go vet ./internal/tui/`, vor der Implementierung):

```
# beans-tui/internal/tui
# [beans-tui/internal/tui]
vet: internal/tui/view_tag_management_test.go:440:13: m.openTagMgmtInput undefined (type model has no field or method openTagMgmtInput)
```

RED (`command go test ./internal/tui/ -run "TestKeyTagMgmt|TestKeyTagManagementNewTag|TestApplyTagDefsSaved|TestFullCreateFlow|TestTagManagementLocalBindingsIncludesNewTag|TestViewTagManagementRendersInputSubmode|TestViewTagManagementInputSubmodeNoWrapAt80" -v -count=1`, vor der Implementierung):

```
# beans-tui/internal/tui [beans-tui/internal/tui.test]
internal/tui/view_tag_management_test.go:440:13: m.openTagMgmtInput undefined (type model has no field or method openTagMgmtInput)
internal/tui/view_tag_management_test.go:442:4: m.tagMgmtInput undefined (type model has no field or method tagMgmtInput)
internal/tui/view_tag_management_test.go:443:14: m.keyTagMgmtInput undefined (type model has no field or method keyTagMgmtInput)
internal/tui/view_tag_management_test.go:458:13: m.openTagMgmtInput undefined (type model has no field or method openTagMgmtInput)
internal/tui/view_tag_management_test.go:460:4: m.tagMgmtInput undefined (type model has no field or method tagMgmtInput)
internal/tui/view_tag_management_test.go:461:14: m.keyTagMgmtInput undefined (type model has no field or method keyTagMgmtInput)
internal/tui/view_tag_management_test.go:480:13: m.openTagMgmtInput undefined (type model has no field or method openTagMgmtInput)
internal/tui/view_tag_management_test.go:482:4: m.tagMgmtInput undefined (type model has no field or method tagMgmtInput)
internal/tui/view_tag_management_test.go:483:14: m.keyTagMgmtInput undefined (type model has no field or method keyTagMgmtInput)
internal/tui/view_tag_management_test.go:500:9: mm.tagMgmtInputActive undefined (type model has no field or method tagMgmtInputActive)
internal/tui/view_tag_management_test.go:500:9: too many errors
FAIL	beans-tui/internal/tui [build failed]
FAIL
```

GREEN (gleiches `-run`-Filter, nach Implementierung, alle 14/14 grün):

```
=== RUN   TestKeyTagMgmtInputRejectsInvalidName
--- PASS: TestKeyTagMgmtInputRejectsInvalidName (0.00s)
=== RUN   TestKeyTagMgmtInputRejectsDuplicateAgainstExistingRows
--- PASS: TestKeyTagMgmtInputRejectsDuplicateAgainstExistingRows (0.00s)
=== RUN   TestKeyTagMgmtInputRejectsDuplicateAgainstFreeRowToo
--- PASS: TestKeyTagMgmtInputRejectsDuplicateAgainstFreeRowToo (0.00s)
=== RUN   TestKeyTagManagementNewTagOpensInput
--- PASS: TestKeyTagManagementNewTagOpensInput (0.00s)
=== RUN   TestKeyTagMgmtInputEscDiscardsOnlySubmode
--- PASS: TestKeyTagMgmtInputEscDiscardsOnlySubmode (0.00s)
=== RUN   TestKeyTagMgmtInputCapturesEveryKeyNoLeak
--- PASS: TestKeyTagMgmtInputCapturesEveryKeyNoLeak (0.00s)
=== RUN   TestKeyTagMgmtInputRetypingNDoesNotReopen
--- PASS: TestKeyTagMgmtInputRetypingNDoesNotReopen (0.00s)
=== RUN   TestKeyTagMgmtInputValidSubmitFiresSaveTagDefsCmdWithAddedName
--- PASS: TestKeyTagMgmtInputValidSubmitFiresSaveTagDefsCmdWithAddedName (0.00s)
=== RUN   TestApplyTagDefsSavedSuccessRefreshesRowsAndMovesCursor
--- PASS: TestApplyTagDefsSavedSuccessRefreshesRowsAndMovesCursor (0.00s)
=== RUN   TestApplyTagDefsSavedErrorKeepsInputOpenShowsToast
--- PASS: TestApplyTagDefsSavedErrorKeepsInputOpenShowsToast (0.00s)
=== RUN   TestFullCreateFlowRefreshesPageAndTouchesNoBean
--- PASS: TestFullCreateFlowRefreshesPageAndTouchesNoBean (0.00s)
=== RUN   TestTagManagementLocalBindingsIncludesNewTag
--- PASS: TestTagManagementLocalBindingsIncludesNewTag (0.00s)
=== RUN   TestViewTagManagementRendersInputSubmode
--- PASS: TestViewTagManagementRendersInputSubmode (0.00s)
=== RUN   TestViewTagManagementInputSubmodeNoWrapAt80
--- PASS: TestViewTagManagementInputSubmodeNoWrapAt80 (0.00s)
PASS
ok  	beans-tui/internal/tui	0.531s
```

Voller T2/T3/T6-Regressionskorpus (`-run "TagManagement|TagPicker|TagRegistryRows|GoTags|CollectTagCounts"`, 51 Tests) grün, keine bestehende Assertion angepasst.

Gates: `command gofmt -l .` leer · `command go vet ./...` leer · `command go test ./... -short -count=1` grün (alle Pakete ok) · `command go test ./... -count=1` grün (Voll-Lauf, `internal/tui` 138.5s, exit 0).

## Golden-Beleg

`command go build -o bin/bt .` erfolgreich, danach `command go test ./internal/tui/ -run "TestTreeGolden|TestBacklogGolden|TestChromeGolden" -count=2 -v` → alle 5 Testfunktionen (inkl. `*Deterministic`-Varianten) PASS, beide Wiederholungen. `git diff --stat -- internal/tui/testdata/` leer — Basis-Goldens unverändert (mirrort T2/T6s eigene Entscheidung: keine eigene Golden-Suite für `viewTagManagement()`, stattdessen Render-/Verhaltenstests wie `TestViewTagManagementRendersInputSubmode`/`TestViewTagManagementInputSubmodeNoWrapAt80`).

## Smoke

tmux, `TERM=xterm-256color`, Breite 120 UND 80 (Sessions `bt120t3`/`bt80t3`, `bin/bt` in diesem Repo). `ctrl+k` → „tags"/„go to tags" → `enter` → Page öffnet (3 freie Tags aus diesem Repo: `to-review`/`rejected`/`smoke`). `n` öffnet den Input:

```
enter:create  esc:cancel

new tag (a-z0-9, hyphen-separated)
```

Ungültigen Namen „Not Valid" probiert → Inline-Fehler, Input bleibt offen:

```
enter:create  esc:cancel

Not Valid

invalid tag name (a-z0-9, hyphen-separated, lowercase)
```

Duplikat „to-review" probiert (bereits als freie Zeile sichtbar) → ebenfalls abgewiesen:

```
tag already defined: to-review
```

Gültigen Namen „my-new-tag" angelegt → sofort in der Liste, Gruppe „Definiert", Count 0, Cursor auf der neuen Zeile:

```
▌✓ my-new-tag                                                            0
   to-review                                                            10
   rejected                                                              1
   smoke                                                                 1
```

`.beans-tags.yml` (Repo-Root) danach:

```
tags:
    - my-new-tag
```

`git status --porcelain` nach diesem Schritt: NUR `?? .beans-tags.yml` — keine `.beans/`-Änderung (D11 bestätigt). Zweiter Create-Versuch mit anschließendem `esc` (Freitext „abort-me" getippt) → Submodus verwirft, KEINE neue Zeile, Page unverändert. Full-Capture-Beleg (Submodus aktiv): `d`, `?`, `ctrl+k` nacheinander gedrückt → alle drei landen als Zeichen im Eingabefeld (`d?`), KEIN Delete-Confirm/Help/Command-Center öffnet, Submodus bleibt aktiv. `esc` → zurück zur Page; zweites `esc` → zurück zu Browse (Breadcrumb wechselt). Sitzung sauber mit `q`→`enter` beendet (kein `bt`-Prozess mehr aktiv danach, `pgrep -fl "bin/bt"` leer).

Wiederholt bei 80 Spalten (Session `bt80t3`, App-Neustart — `.beans-tags.yml` aus dem 120-Spalten-Lauf lädt frisch beim Page-Open, D03 bestätigt: `my-new-tag` erscheint sofort). Zweiten Tag „second-tag" angelegt (Definiert-Gruppe wächst auf 2). Ungültiger Name UND Duplikat („my-new-tag" erneut) beide korrekt abgewiesen, Fehlertext passt bei 80 Spalten ohne Umbruch (kurze Meldungen brauchten in dieser Live-Session kein „…" — der ⁠abschneidende Pfad selbst ist zusätzlich per `TestViewTagManagementInputSubmodeNoWrapAt80` mit einer synthetisch langen Fehlermeldung abgedeckt, mirrort T2s eigene Truncation-Test-Konvention). Session sauber beendet.

Nach BEIDEN Läufen: `.beans-tags.yml` gelöscht (`rm .beans-tags.yml`), `git status --porcelain` zeigt ausschließlich die erwarteten Quelldateien (kein `.beans/`, keine Registry-Datei).

## Deviations/ERRATA

- **Test-Sketch-Bug in den beiden bean-eigenen RED-Test-Vorlagen** (mirrort bt-pqq3s eigenen B-ERRATUM-Präzedenzfall): `m, _ = m.openTagMgmtInput(...)` kompiliert nicht — `openTagMgmtInput` gibt `(tea.Model, tea.Cmd)` zurück (Interface), `m` ist die konkrete `model`-Struktur; dieselbe Assignability-Regel, die JEDE andere `tea.Model`-Methode in diesem Repo bereits über eine frische Variable + Typ-Assertion löst (z. B. `TestOpenTagManagementPageNilClientBuildsFromIdxOnly`: `nm, cmd := m.openTagManagementPage(); mm := nm.(model)`). Gefixt durch `nm, _ := m.openTagMgmtInput(...); m = nm.(model)` — Assertions UNVERÄNDERT, nur die Zuweisungs-Mechanik korrigiert.
- **Dedupe-Scope explizit gegen ALLE Zeilennamen (defined UND frei), nicht nur Registry-Einträge:** der Epic-Body (D11) spricht von „Dedupe gegen bestehende Registry-Einträge", der T3-Bean-Body selbst aber explizit „Dedupe-Check gegen die aktuelle `m.tagMgmtRows`-Namensmenge ... Dedupe prüft also gegen ALLE vorhandenen Namen" — dem literaleren, spezifischeren T3-Wortlaut gefolgt (ALLE Zeilen, nicht nur `defined==true`). Konsequenz: ein exaktes Retippen eines bereits sichtbaren FREIEN Tag-Namens wird ebenfalls als Duplikat abgewiesen, nicht stillschweigend als „diesen freien Tag befördern" interpretiert (die D09-Beförderungs-Erzählung bleibt ein dokumentierter, nicht gebauter Fast-Follow, Epic-Q-Sektion) — eigener Regressionstest `TestKeyTagMgmtInputRejectsDuplicateAgainstFreeRowToo` pinnt das Verhalten explizit.
- **`saveTagDefsCmd` liegt in `messages.go`, nicht in `view_tag_management.go`:** der Bean-Body listet die Funktion unter der `view_tag_management.go`-Aufzählung, aber JEDER andere `tea.Cmd`-Produzent im gesamten Repo (`mutateCmd`/`createCmd`/`searchCmd`/`showRawCmd`/`loadCmd`/`switchRepoCmd`/`repoMetricsCmd`/`repoMetricsBatchCmd`, 8/8) liegt in `messages.go`, dessen eigener Doc-Kommentar sich explizit als „tea.Msg types + tea.Cmd producers ONLY" deklariert — der etablierten, repo-weiten Konvention gefolgt statt der vermutlich rein redaktionellen Bean-Gliederung; `tagDefsSavedMsg` liegt (wie im Bean-Body explizit gefordert) ebenfalls dort.
- **Render-Funktion `tagMgmtInputRows()` nicht im Bean-Body benannt, aber notwendig:** der Bean-Body listet keine neue View-Funktion für den Submodus, die tmux-Smoke-Vorgabe („Inline-Fehler", „Bei 80 Spalten: Inline-Fehlertext darf nicht truncaten ohne …") verlangt aber sichtbaren Render-Output. Ergänzt: `tagMgmtInputRows()` (Hint+Feld+Fehler, INLINE in derselben Pane, D06/D14) + der `viewTagManagement`-Zweig, der bei `tagMgmtInputActive` diese Rows statt der normalen Liste rendert — nutzt `renderPane`s bestehendes `truncate(w-2)`-Zeilenbudget, kein neuer Truncation-Code.
- **Bindungs-Position `keys.NewTag` in `tagManagementLocalBindings`:** Bean-Wortlaut „hängt an" ist positionsoffen — vor `Back` gewählt (mirrort `tagPickerLocalBindings`s Action-vor-Back-Konvention, `footer_context.go`), nicht ans Ende.
- Zusätzlich zu den 2 im Bean-Body zitierten RED-Test-Vorlagen (mit Fix, s.o.) wurden 12 weitere Tests ergänzt (Full-Capture-Leak, Duplikat-gegen-frei, Persistenz-Roundtrip, Erfolgs-/Fehler-Tail von `applyTagDefsSaved`, Ende-zu-Ende-Flow inkl. D11-Regressionstest, Footer-Binding, Render-Submodus, 80-Spalten-Truncation) — reine Ergänzung, keine Abweichung vom Spezifizierten.

Sonst keine Abweichungen vom Bean-Body.

## Notes for T4+T5

- **T5 (Rename) nutzt den geteilten Submodus wörtlich wie hier vorbereitet:** `openTagMgmtInput("rename", altName)` befüllt Feld UND `tagMgmtInputTarget` mit dem alten Namen in EINEM Aufruf; `keyTagMgmtInput`s Dedupe-Schleife exkludiert `tagMgmtInputTarget` bereits (`r.name == name && name != m.tagMgmtInputTarget`) — T5 muss dort NICHTS ändern, nur den `switch m.tagMgmtInputMode`-Zweig um `case "rename":` ergänzen (dispatcht `saveTagDefsCmd` mit `data.RenameTagDefName(definedTagNames(m.tagMgmtRows), tagMgmtInputTarget, name)`, T1-Helfer bereits vorhanden). `tagMgmtInputRows()` hat bereits einen `"rename"`-Hint-Zweig vorgesehen („enter:rename esc:cancel") — T5 muss auch dort nichts mehr ergänzen.
- **`applyTagDefsSaved` ist bereits Mode-agnostisch:** liest IMMER frisch von Disk (`LoadTagDefs`) und sucht den Cursor per Name — funktioniert für Rename identisch zu Create, sofern T5 vor dem Submit `tagMgmtInputTarget`/`tagMgmtInput.Value()` korrekt setzt. Keine T5-spezifische Änderung an `applyTagDefsSaved` erwartet.
- **`definedTagNames(m.tagMgmtRows)`** ist der einzige Weg, an die aktuelle Registry-Teilmenge zu kommen (kein separates `m.tagMgmtDefs`-Feld existiert) — T4 (Delete) sollte denselben Helfer für `RemoveTagDefName` wiederverwenden statt eine eigene Filter-Schleife zu bauen.
- **T4 (Delete, PARALLEL zu T3 laut Epic-Tabelle) ist NICHT von diesem Task betroffen** — disjunkter Verhaltens-Scope (Confirm-Dialog, `tagMgmtDeleteConfirm`/`tagMgmtDeleteTarget`, D15), aber gleiche Datei (`view_tag_management.go`) — bei einem Merge-Konflikt: `tagManagementLocalBindings` wird von BEIDEN Tasks erweitert (T3 fügt `keys.NewTag` VOR `Back` ein, T4 vermutlich `keys.Delete` an ähnlicher Stelle) — manuell zusammenführen, nicht einfach überschreiben.
- **`saveTagDefsCmd`/`tagDefsSavedMsg` sind bereits generisch genug für T4/T5** (nehmen nur `[]string defs` entgegen) — kein Task-3-spezifisches Detail muss dafür angepasst werden, nur der jeweilige Aufrufer (`RemoveTagDefName`/`RenameTagDefName` statt `AddTagDefName`).
