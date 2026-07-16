---
# bt-y9my
title: T5 — Tag umbenennen + Propagation über alle Beans
status: completed
type: task
priority: normal
created_at: 2026-07-16T15:44:35Z
updated_at: 2026-07-16T18:57:23Z
parent: bt-362n
blocked_by:
    - bt-604w
---

T5 — Tag-Definition umbenennen + Propagation über alle Beans (Epic
`bt-362n`, D13, D14). `blocked_by` T3 (reuse des dort eingeführten
Freitext-Input-Submodus `tagMgmtInputActive`/`tagMgmtInputMode`/
`tagMgmtInput`). Größtes Einzelstück dieser Kette (mirrort E9s F01 als
„größtes Einzelstück"-Präzedenz für die Sorgfaltsstufe).

## Ziel

Einen definierten Tag umbenennen: Registry-Rename sofort (D13, praktisch
nie fehlschlagend), DANACH ein Best-Effort-Sweep über alle Beans mit dem
alten Tag (`idx.WithTag(alt)`), je EIN kombinierter `SetTags`-Call pro
Bean, continue-on-error, Ergebnis als EIN Toast.

## Betroffene Dateien/Symbole

- `internal/tui/messages.go`:
  - `type tagRenameDoneMsg struct { oldTag, newTag string; renamed int;
    failed []tagRenameFailure }` — `type tagRenameFailure struct { id
    string; err error }` (die ZWEITE bewusste Ausnahme vom geteilten
    `mutationDoneMsg`-Tail, D13 — mirrort `createDoneMsg`s eigene
    Doc-Stamp-Begründung „braucht mehr als einen bloßen Error zurück").
  - `func renameTagCmd(c *data.Client, idx *data.Index, oldTag, newTag
    string) tea.Cmd` — EIN `tea.Cmd`, das SYNCHRON (innerhalb der
    Goroutine, wie jeder andere Cmd) über `idx.WithTag(oldTag)` iteriert,
    je Bean `etag, ok := ...` (aus `idx` direkt, da dies EIN
    In-Memory-Snapshot-Sweep ist, kein `m.beanETag`-Redirect nötig — der
    Sweep arbeitet auf dem `idx`, der ihm beim Dispatch übergeben wurde)
    und `c.SetTags(b.ID, []string{newTag}, []string{oldTag}, etag)`
    aufruft; sammelt `renamed`/`failed` OHNE bei einem Fehler
    abzubrechen (D13 continue-on-error — KEIN `return` im Fehlerfall der
    Schleife); gibt am Ende EINE `tagRenameDoneMsg` zurück.
- `internal/tui/update.go`:
  - `func (m model) applyTagRenameDone(msg tagRenameDoneMsg) (tea.Model,
    tea.Cmd)` — baut EINEN Toast-Text („Renamed <oldTag> → <newTag>:
    <renamed> bean(s) updated" + bei `len(failed) > 0`: „, <N> failed
    (first: <failed[0].err>)" angehängt, mirrort
    `applyMutationResult`s Toast-Text-Zusammensetzung), Toast-Kind
    `toastError`, wenn `len(failed) > 0`, sonst normaler Erfolgs-Toast
    (welcher Toast-Kind-Wert „Erfolg" ist: gegen `overlay_show_toast.go`
    verifizieren, gleiches Muster wie jeder bestehende Erfolgspfad
    nutzen, NICHT neu erfinden); dispatcht `loadCmd(m.client)` IMMER am
    Ende (D13 — mirrort `applyMutationResult`s „always reload after"),
    damit `m.idx` den neuen Stand zeigt, BEVOR `tagMgmtRows` das nächste
    Mal aus ihm gebaut wird.
  - `handleKey`/`Update`-Dispatcher: neuer `case tagRenameDoneMsg:
    return m.applyTagRenameDone(msg)` (mirrort wie `createDoneMsg`/
    `mutationDoneMsg` bereits im zentralen `Update`-Switch verdrahtet
    sind).
- `internal/tui/view_tag_management.go`:
  - `func (m model) openTagMgmtRename() (tea.Model, tea.Cmd)` — liest
    die selektierte Zeile, No-Op wenn `!defined` (nur definierte Tags
    lassen sich umbenennen — ein freier Tag hat keinen Registry-Eintrag,
    den man „umbenennen" könnte; PO kann ihn stattdessen erst per `n`
    definieren), ruft `m.openTagMgmtInput("rename", row.name)` (T3-Infra,
    D14) MIT `tagMgmtInputTarget = row.name` (Dedupe-Check in
    `keyTagMgmtInput` muss den EIGENEN alten Namen aus dem
    Duplikat-Check ausschließen — T3s Dedupe-Logik bekommt hier ihren
    zweiten, in T3 bereits vorgesehenen Verzweigungspunkt, s. T3-Body
    „EXCL. tagMgmtInputTarget bei Rename").
  - `keyTagMgmtInput` (aus T3) bekommt den zweiten `switch`-Zweig:
    `case "rename": altName := m.tagMgmtInputTarget; neueDefs :=
    data.RenameTagDefName(currentDefs, altName, neuerName); return m,
    tea.Batch(saveTagDefsCmd(m.client, neueDefs), renameTagCmd(m.client,
    m.idx, altName, neuerName))` — ZWEI Cmds im selben `tea.Batch` (D13:
    Registry-Save und Bean-Sweep sind UNABHÄNGIG, keiner wartet auf den
    anderen — der Registry-Save ist synchron-schnell und wird i. d. R.
    zuerst ankommen, aber `tea.Batch` garantiert KEINE Reihenfolge; das
    ist bewusst folgenlos, da beide Cmds disjunkte State-Teile schreiben:
    `applyTagDefsSaved` rührt nur `tagMgmtRows`/`tagMgmtInputActive` an,
    `applyTagRenameDone` nur `m.idx`/den Toast — kein Write-Write-Konflikt).
- `internal/tui/view_tag_management.go` (`tagManagementLocalBindings`):
  neue `keys.RenameTag`-Bindung (NEUES `keyMap`-Feld, `keymap.go`) —
  gebunden auf `"e"` (D-Rationale, Epic-Body-Stil: mnemonisch „edit
  name", gleiche Taste wie das GLOBALE `keys.Editor` in einem
  komplett anderen, sich gegenseitig ausschließenden Anzeigekontext —
  mirrort den bereits etablierten Backlog-`S`-Präzedenzfall
  „gleiche Taste, unterschiedliche View-lokale Bedeutung, disjunkte
  Footer-Sichtbarkeit", KEIN funktionaler Konflikt da `viewTagManagement`
  full-capture ist, D06, `keyNodeAction`s `e`-Case wird nie erreicht
  während dieser View aktiv ist).
- `internal/tui/keymap.go`:
  - `keyMap` bekommt Feld `RenameTag keybind.Binding` (Help-Text
    „Rename").
  - `newKeyMap()`: `RenameTag: keybind.NewBinding(keybind.WithKeys("e"),
    keybind.WithHelp("e", "Rename"))`.
  - `helpGroups()`: `RenameTag` in die „Actions"-Gruppe aufnehmen (NEBEN
    `NewTag`, gleiche Sektion — Drift-Guard `TestHelpGroupsCoverEvery
    BindingExactlyOnce` verlangt genau das).
- `internal/tui/update.go` (`keyTagManagement`): neuer Case
  `keybind.Matches(msg, keys.RenameTag): return m.openTagMgmtRename()`
  (NACH dem `tagMgmtInputActive`/`tagMgmtDeleteConfirm`-Vorrang-Check aus
  T3/T4).

## TDD (RED zuerst)

```go
func TestRenameTagCmdContinuesPastOneFailure(t *testing.T) {
    // Fake-Client mit einer Bean-ID, deren SetTags absichtlich fehlschlägt
    // (stale-etag-artiger Fehler), UND zwei weiteren, die erfolgreich sind.
    idx := data.NewIndex([]data.Bean{
        {ID: "b1", Tags: []string{"old"}, ETag: "e1"},
        {ID: "b2", Tags: []string{"old"}, ETag: "stale"},
        {ID: "b3", Tags: []string{"old"}, ETag: "e3"},
    })
    fc := &fakeSetTagsClient{failIDs: map[string]bool{"b2": true}}
    msg := renameTagCmd(fc, idx, "old", "new")().(tagRenameDoneMsg)
    if msg.renamed != 2 || len(msg.failed) != 1 || msg.failed[0].id != "b2" {
        t.Fatalf("want 2 renamed, 1 failed(b2), got %+v", msg)
    }
    if !fc.calledFor["b1"] || !fc.calledFor["b3"] {
        t.Fatal("want b1 AND b3 still attempted despite b2's failure")
    }
}

func TestApplyTagRenameDoneAlwaysReloads(t *testing.T) {
    m := newModel(nil, "")
    _, cmd := m.applyTagRenameDone(tagRenameDoneMsg{oldTag: "old", newTag: "new", renamed: 1})
    if cmd == nil {
        t.Fatal("want a non-nil Cmd (toast + loadCmd batch)")
    }
}
```

(`fakeSetTagsClient` — kleines Test-Double, das `data.Client`s
`SetTags`-Signatur erfüllt, mirrort bestehende Test-Doubles im Repo für
Mutation-Cmds, falls vorhanden; sonst neu anlegen, minimal.)

Weitere Pflicht-Tests: `openTagMgmtRename` No-Op auf freier Zeile;
Dedupe-Check im Rename-Modus lässt den EIGENEN alten Namen durch (kein
„Duplikat"-Fehler, wenn der PO die Eingabe unverändert lässt und erneut
enter drückt — Randfall, EXPLIZIT zu testen); `RenameTagDefName`-Aufruf
mit dem korrekten alten/neuen Namen; `tea.Batch` liefert BEIDE Cmds
(Cmd-Liste-Länge prüfen, kein Test gegen Ausführungsreihenfolge, s.
Rationale oben).

## Golden-Strategie

GEGENBELEG (Tree/Backlog/Chrome unverändert — der Sweep ändert
Bean-Daten, aber über den BESTEHENDEN `SetTags`-Pfad, der schon heute
keine Golden-relevante Render-Änderung auslöst). Eigene Golden-Suite
(falls vorhanden): optional ein „Rename-Input vorbefüllt"-Snapshot.

## tmux-Smoke (120 UND 80 Spalten)

In einem SEPARATEN Scratch-Repo (NICHT diesem `beans-tui-repository` —
der Sweep mutiert echte Beans, ein Live-Smoke gegen die eigene
Dogfooding-`.beans/` wäre destruktiv/schwer rückgängig zu machen; Scratch
gemäß bestehender Konvention für mutations-nahe manuelle Tests, mirrort
`beans-src/CLAUDE.md`s eigene „für alles was Daten ändert: separates
Test-Projekt"-Regel): 3-4 Beans mit Tag „alt" anlegen, Page öffnen,
definierten Tag „alt" per `e` umbenennen zu „neu", Toast-Text „3
bean(s) updated" (o. ä.) verifizieren, `beans list --tag neu --json`
extern gegenprüfen, dass alle 3 den neuen Tag tragen und „alt" weg ist.
Wiederholen bei 80 Spalten (Toast-Text darf nicht abgeschnitten werden
ohne `…`). Scratch-Repo danach löschen (nie Teil des Commits).

## Akzeptanz-Checkliste

- [x] `e` auf definierter Zeile öffnet Rename-Input vorbefüllt mit dem alten Namen
- [x] `e` auf freier Zeile No-Op
- [x] Submit benennt Registry SOFORT um
- [x] Sweep läuft asynchron, continue-on-error (Regressionstest mit simuliertem Teilfehlschlag)
- [x] KEINE Cross-Bean-Transaktion vorgetäuscht (D13 dokumentiert)
- [x] Toast zeigt renamed/failed-Zahlen
- [x] `m.idx` lädt danach neu
- [x] eigener alter Name im Dedupe-Check durchgelassen
- [x] `RenameTag`-Binding in `helpGroups()`, Drift-Guard grün
- [x] Goldens Gegenbeleg grün
- [x] tmux-Smoke 120+80 in Scratch-Repo belegt, Scratch danach gelöscht
- [x] voller Lauf grün
- [x] Commit `feat(tui): E10 Tag-Definition umbenennen + Bean-Propagation`

## Summary

Liefert D13 (Rename-Semantik & Propagation) + reused D14 (Input-Submodus,
T3-Infra wörtlich wiederverwendet, kein zweiter Mechanismus). Neu in
`internal/tui/messages.go`: `tagRenameFailure{id, err}` +
`tagRenameDoneMsg{oldTag, newTag, renamed, failed}` (die ZWEITE bewusste
Ausnahme vom geteilten `mutationDoneMsg`-Tail nach `tagDefsSavedMsg`) +
`renameTagCmd(c *data.Client, idx *data.Index, oldTag, newTag string)
tea.Cmd` — iteriert `idx.WithTag(oldTag)` (in-memory Snapshot, Etag direkt
vom `idx`-Bean-Pointer, kein `m.beanETag`-Redirect nötig), feuert je Bean
GENAU EINEN `c.SetTags(b.ID, []string{newTag}, []string{oldTag}, b.ETag)`
und sammelt `renamed`/`failed` OHNE bei einem Fehler abzubrechen
(continue-on-error, D13 — kein `return` in der Fehler-Schleife).
Zwei defensive Guards, beide über das bean-eigene TDD hinaus ergänzt und im
Deviations-Abschnitt begründet: `idx == nil` degradiert zu einem
Zero-Value-No-Op-Sweep (mirrort `collectTagCounts`' eigenen
`if idx != nil`-Guard), `oldTag == newTag` degradiert ebenfalls zu einem
No-Op (kein `SetTags`-Aufruf) statt jeden betroffenen Bean über
`SetTags`' dokumentiertes „gleicher Tag in add UND remove → remove
gewinnt"-Verhalten (mutations.go I2) versehentlich zu entfernen.

`internal/tui/update.go`: `applyTagRenameDone(msg tagRenameDoneMsg) (tea.Model,
tea.Cmd)` — baut den Toast-Text („Renamed <old> → <new>: <N> bean(s)
updated" + bei Fehlern „, <N> failed (first: <err>)"), Toast-Kind
`toastInfo` bei Erfolg (verifiziert gegen `overlay_show_toast.go`s eigenen
„Blue/Mauve — Hinweis/Erfolg"-Doc-Stamp, dieselbe Kind wie der bestehende
Yank-Erfolgstoast) / `toastError` bei mind. einem Fehlschlag, dispatcht
`tea.Batch(toastCmd, loadCmd(m.client))` IMMER (mirrort `applyMutationResult`s
„always reload after"). Zentraler `Update()`-Switch bekommt `case
tagRenameDoneMsg: return m.applyTagRenameDone(msg)`.

`internal/tui/view_tag_management.go`: `openTagMgmtRename()` — liest die
selektierte Zeile, No-Op wenn außerhalb Range oder `!defined` (mirrort
`openTagMgmtDeleteConfirm`s eigenes No-Op-Muster 1:1), sonst
`openTagMgmtInput("rename", row.name)` (T3-Infra, D14 — befüllt Feld UND
`tagMgmtInputTarget` in einem Aufruf). `keyTagMgmtInput`s
`switch m.tagMgmtInputMode` bekommt `case "rename":` — dispatcht ZWEI
UNABHÄNGIGE Cmds im selben `tea.Batch`: `saveTagDefsCmd(client,
data.RenameTagDefName(defs, old, name), name)` (Registry, T1-Helfer, bereits
vorhanden) und `renameTagCmd(client, m.idx, old, name)` (Sweep) — KEINE
Ausführungsreihenfolge garantiert/benötigt (D13: beide Cmds schreiben
disjunkte State-Teile, kein Write-Write-Konflikt, `tagRenameDoneMsg`s eigener
Doc-Stamp begründet das ausführlich). `keyTagManagement` bekommt einen
neuen Case `keybind.Matches(msg, keys.RenameTag): return
m.openTagMgmtRename()` NACH den bestehenden `tagMgmtInputActive`/
`tagMgmtDeleteConfirm`-Vorrang-Checks aus T3/T4. `tagManagementLocalBindings()`
wächst um `keys.RenameTag` (Position: nach `Delete`, vor `Back`).

`internal/tui/keymap.go`: neues Feld `RenameTag keybind.Binding`
(`keybind.NewBinding(keybind.WithKeys("e"), keybind.WithHelp("e",
"Rename"))`), in `helpGroups()`s „Actions"-Gruppe neben `NewTag`
aufgenommen (Drift-Guard `TestHelpGroupsCoverEveryBindingExactlyOnce` bleibt
grün — der Guard identifiziert Bindings über ihr eigenes `Keys()`-Set, „e"
allein vs. „e,ctrl+e" (`Editor`) sind zwei unterschiedliche Identitäten,
keine Kollision trotz derselben physischen Taste, exakt der bereits
etablierte Backlog-„S"-Präzedenzfall).

Die Dedupe-Exklusion des eigenen alten Namens war bereits VOR diesem Task
in `keyTagMgmtInput` vorhanden (T3s `name != m.tagMgmtInputTarget`-Check,
für Create immer wirkungslos da `tagMgmtInputTarget == ""`) — T5 musste dort
NICHTS ändern, nur explizit regressionsgesichert (s. Test-Output).

## Test-Output (RED→GREEN wörtlich)

RED (`command go vet ./internal/tui/`, Implementierung per `git stash`
temporär zurückgenommen, NUR die Testdatei-Änderung aktiv — Nachweis, dass
die neuen Tests wirklich gegen fehlenden Code kompilieren):

```
# beans-tui/internal/tui
# [beans-tui/internal/tui]
vet: internal/tui/view_tag_management_test.go:1475:15: m.openTagMgmtRename undefined (type model has no field or method openTagMgmtRename)
```

RED (`command go test ./internal/tui/ -run "TestOpenTagMgmtRename|TestKeyTagManagementRename|TestKeyTagMgmtInputRename|TestKeyTagMgmtInputEscInRenameMode|TestKeyTagMgmtInputCapturesEveryKeyNoLeakInRenameMode|TestTagManagementLocalBindingsIncludesRenameTag|TestRenameTagCmd|TestApplyTagRenameDone|TestUpdateDispatchesTagRenameDoneMsg|TestFullRenameFlow" -v -count=1`, gleicher Zwischenstand):

```
# beans-tui/internal/tui [beans-tui/internal/tui.test]
internal/tui/view_tag_management_test.go:1475:15: m.openTagMgmtRename undefined (type model has no field or method openTagMgmtRename)
internal/tui/view_tag_management_test.go:1493:13: m.openTagMgmtRename undefined (type model has no field or method openTagMgmtRename)
internal/tui/view_tag_management_test.go:1509:15: m.openTagMgmtRename undefined (type model has no field or method openTagMgmtRename)
internal/tui/view_tag_management_test.go:1741:27: keys.RenameTag undefined (type keyMap has no field or method RenameTag)
internal/tui/view_tag_management_test.go:1774:9: undefined: renameTagCmd
internal/tui/view_tag_management_test.go:1801:9: undefined: renameTagCmd
internal/tui/view_tag_management_test.go:1818:9: undefined: renameTagCmd
internal/tui/view_tag_management_test.go:1829:9: undefined: renameTagCmd
internal/tui/view_tag_management_test.go:1851:9: undefined: renameTagCmd
internal/tui/view_tag_management_test.go:1863:14: m.applyTagRenameDone undefined (type model has no field or method applyTagRenameDone)
internal/tui/view_tag_management_test.go:1863:14: too many errors
FAIL	beans-tui/internal/tui [build failed]
FAIL
```

Implementierung per `git stash pop` zurückgeholt. GREEN (gleiches
`-run`-Filter, 23/23 grün):

```
=== RUN   TestOpenTagMgmtRenameNoOpOnFreeTag
--- PASS: TestOpenTagMgmtRenameNoOpOnFreeTag (0.00s)
=== RUN   TestOpenTagMgmtRenameNoOpWhenCursorOutOfRange
--- PASS: TestOpenTagMgmtRenameNoOpWhenCursorOutOfRange (0.00s)
=== RUN   TestOpenTagMgmtRenameOnDefinedRowOpensPrefilledInput
--- PASS: TestOpenTagMgmtRenameOnDefinedRowOpensPrefilledInput (0.00s)
=== RUN   TestKeyTagManagementRenameDispatchesToOpenInput
--- PASS: TestKeyTagManagementRenameDispatchesToOpenInput (0.00s)
=== RUN   TestKeyTagManagementRenameNoOpOnFreeRowViaFullDispatch
--- PASS: TestKeyTagManagementRenameNoOpOnFreeRowViaFullDispatch (0.00s)
=== RUN   TestKeyTagMgmtInputRenameDedupeAllowsOwnOldNameUnchanged
--- PASS: TestKeyTagMgmtInputRenameDedupeAllowsOwnOldNameUnchanged (0.00s)
=== RUN   TestKeyTagMgmtInputRenameRejectsDuplicateAgainstOtherExistingName
--- PASS: TestKeyTagMgmtInputRenameRejectsDuplicateAgainstOtherExistingName (0.00s)
=== RUN   TestKeyTagMgmtInputEscInRenameModeDiscardsOnlySubmodeNoSave
--- PASS: TestKeyTagMgmtInputEscInRenameModeDiscardsOnlySubmodeNoSave (0.00s)
=== RUN   TestKeyTagMgmtInputCapturesEveryKeyNoLeakInRenameMode
--- PASS: TestKeyTagMgmtInputCapturesEveryKeyNoLeakInRenameMode (0.00s)
=== RUN   TestKeyTagMgmtInputRenameValidSubmitFiresBatchOfTwoCmds
--- PASS: TestKeyTagMgmtInputRenameValidSubmitFiresBatchOfTwoCmds (0.00s)
=== RUN   TestKeyTagMgmtInputRenameRegistryPersistsIndependentlyOfSweepCmd
--- PASS: TestKeyTagMgmtInputRenameRegistryPersistsIndependentlyOfSweepCmd (0.00s)
=== RUN   TestTagManagementLocalBindingsIncludesRenameTag
--- PASS: TestTagManagementLocalBindingsIncludesRenameTag (0.00s)
=== RUN   TestRenameTagCmdContinuesPastOneFailure
--- PASS: TestRenameTagCmdContinuesPastOneFailure (0.28s)
=== RUN   TestRenameTagCmdDispatchesTagAndRemoveTagFlagsPerBean
--- PASS: TestRenameTagCmdDispatchesTagAndRemoveTagFlagsPerBean (0.13s)
=== RUN   TestRenameTagCmdNoBeansWithOldTagReturnsZeroRenamedNoFailures
--- PASS: TestRenameTagCmdNoBeansWithOldTagReturnsZeroRenamedNoFailures (0.00s)
=== RUN   TestRenameTagCmdNilIndexNoPanicZeroRenamed
--- PASS: TestRenameTagCmdNilIndexNoPanicZeroRenamed (0.00s)
=== RUN   TestRenameTagCmdSameOldNewNameSkipsSweepNoDestructiveSetTags
--- PASS: TestRenameTagCmdSameOldNewNameSkipsSweepNoDestructiveSetTags (0.00s)
=== RUN   TestApplyTagRenameDoneAlwaysReloads
--- PASS: TestApplyTagRenameDoneAlwaysReloads (0.00s)
=== RUN   TestApplyTagRenameDoneSuccessTextAndToastKind
--- PASS: TestApplyTagRenameDoneSuccessTextAndToastKind (0.00s)
=== RUN   TestApplyTagRenameDoneFailureTextAndToastKindError
--- PASS: TestApplyTagRenameDoneFailureTextAndToastKindError (0.00s)
=== RUN   TestUpdateDispatchesTagRenameDoneMsgToApplyTagRenameDone
--- PASS: TestUpdateDispatchesTagRenameDoneMsgToApplyTagRenameDone (0.00s)
=== RUN   TestFullRenameFlowRenamesRegistryAndSweepsBeansViaRealClient
--- PASS: TestFullRenameFlowRenamesRegistryAndSweepsBeansViaRealClient (0.13s)
PASS
ok  	beans-tui/internal/tui	1.037s
```

Voller T2-T6-Regressionskorpus (`-run
"TagManagement|TagPicker|TagRegistryRows|GoTags|CollectTagCounts"`) grün,
keine bestehende Assertion angepasst. Drift-Guard-Sweep
(`TestHelpGroupsCoverEveryBindingExactlyOnce|
TestNoDuplicateBindingBetweenGlobalAndAnyLocalHintList|
TestHistoryBindingsUnbelegtElsewhere|TestGlobalBindings*`) grün.

Gates: `command gofmt -l .` leer · `command go vet ./...` leer ·
`command go test ./... -short -count=1` grün (alle Pakete ok) ·
`command go test ./... -count=1` grün (Voll-Lauf, `internal/tui` 139.2s,
exit 0).

## Golden-Beleg

`command go build -o bin/bt .` erfolgreich, danach `command go test
./internal/tui/ -run "TestTreeGolden|TestBacklogGolden|TestChromeGolden"
-count=2 -v` → alle 5 Testfunktionen (inkl. `*Deterministic`-Varianten)
PASS, beide Wiederholungen (10/10). `git diff --stat --
internal/tui/testdata/` leer — Basis-Goldens unverändert (mirrort T2-T4s
eigene Entscheidung: keine eigene Golden-Suite für `viewTagManagement()`,
stattdessen Render-/Verhaltenstests, hier keine neuen nötig da Rename
denselben `tagMgmtInputRows()`-Render-Pfad wie Create nutzt, bereits von
T3s `TestViewTagManagementInputSubmodeNoWrapAt80` abgedeckt).

## Smoke

tmux, `TERM=xterm-256color`, SEPARATES Scratch-Repo (NICHT dieses
Dogfooding-Repo — der Sweep mutiert echte Beans), Breite 120 UND 80
(Sessions `bt120t5`/`bt80t5`).

**Setup (empirische Abweichung von der ursprünglichen Smoke-Vorgabe, s.
Deviations):** 3 Wegwerf-Beans OHNE Tag angelegt, dann JEWEILS per
separatem `beans update --tag smoke-rename-src --if-match <create-etag>
--json` getaggt (NICHT `--tag` direkt bei `beans create`) — Grund: eine
echte, upstream-dokumentierte beans-0.4.2-ETag-Inkonsistenz (bereits in
`internal/data/testrepo_test.go`s eigenem Doc-Stamp beschrieben: „für einen
Bean, dessen On-Disk-Frontmatter einen tags:-Block trägt, stimmt der von
list/show berichtete ETag nicht mit dem, was update --if-match intern
prüft, überein — konsistent erst nach dem ersten erfolgreichen update").
Empirisch reproduziert (s. Deviations) — Beans, die den Tag SCHON BEI
`create` bekommen, lassen JEDEN nachfolgenden `SetTags`-Aufruf des Sweeps
mit einem echten CONFLICT fehlschlagen; ein Tag, das per NACHTRÄGLICHEM
`update` gesetzt wird, „repariert" den ETag für alle folgenden Aufrufe.
`.beans-tags.yml` manuell mit `tags:\n    - smoke-rename-src\n`
vorbefüllt (Registry-Vorbedingung — der bestehende `n`-Create-Pfad dedupet
bewusst gegen ALLE Zeilennamen inkl. freier, T3s eigene Deviation, kann
einen bereits benutzten freien Namen also NICHT „befördern").

**120 Spalten:** `ctrl+k` → „go to tags" → `enter` → Page zeigt
`smoke-rename-src` als definierte Zeile, Count 3. Cursor darauf, `e` →
Rename-Input öffnet vorbefüllt mit „smoke-rename-src" (Hint „enter:rename
esc:cancel"). Feld geleert (End + Backspaces), „smoke-rename-dst"
getippt, `enter` → Toast „● Renamed smoke-rename-src → sm…" (36-Spalten-Box,
korrekt mit `…` gekürzt) → Zeilenliste zeigt SOFORT „smoke-rename-dst"
Count 0 / „smoke-rename-src" Count 3 (D13: Registry-Zeile ist frisch,
`m.idx` des bereits offenen Panes noch NICHT — `applyLoaded` baut
`tagMgmtRows` nicht reaktiv neu, nur `openTagManagementPage` tut das, D03 —
dokumentierter, architekturkonformer Zwischenzustand, kein Bug). Externe
Gegenprobe: `beans list --tag smoke-rename-dst --json` → alle 3 Beans,
`beans list --tag smoke-rename-src --json` → `null`. `.beans-tags.yml` →
`tags:\n    - smoke-rename-dst`. Page verlassen (`esc`) + neu geöffnet
(D03-Reload) → zeigt jetzt korrekt „smoke-rename-dst" Count 3. Ein Bean
per `cat .beans/<id>--*.md` zitiert: `tags:\n    - smoke-rename-dst`
bestätigt. `e` auf einer FREIEN Zeile (`loose-tag`, separat per CLI
angelegt) → No-Op verifiziert (kein Input öffnet, Cursor bleibt auf der
Zeile). Sitzung sauber mit `q`→`enter` beendet, `pgrep -fl "bin/bt"`
leer danach.

**80 Spalten (verkürzter Happy-Path):** Session-Neustart gegen denselben
Scratch-Repo-Stand, `smoke-rename-dst`→`smoke-rename-final` umbenannt →
Toast korrekt mit „…" gekürzt, kein rohes Abschneiden. Externe Gegenprobe:
`beans list --tag smoke-rename-final --json` → alle 3 Beans,
`smoke-rename-dst` → `null`, `.beans-tags.yml` → `smoke-rename-final`.
Sitzung sauber mit `q`→`enter` beendet.

**Teardown:** `rm -rf` auf das komplette Scratch-Repo (lag außerhalb
dieses Repos, `/private/tmp/.../scratchpad/bt-y9my-smoke-scratch`) —
niemals Teil dieses Commits. `git status --porcelain` (dieses Repo) danach
LEER; `git status --porcelain .beans/` separat verifiziert LEER (kein
echtes Bean berührt). `bin/bt` (git-ignored) ebenfalls entfernt.

## Deviations/ERRATA

- **Kein `fakeSetTagsClient`-Test-Double (Abweichung vom bean-eigenen
  RED-Test-Sketch):** `data.Client` ist ein KONKRETER Struct (kein
  Interface-Seam, shellt über `exec.Command` zur echten `beans`-CLI aus) —
  `renameTagCmd(fc, idx, ...)` mit `fc := &fakeSetTagsClient{...}` würde
  nicht kompilieren (dieselbe Assignability-Klasse von Sketch-Bug, die
  T3/T4 bereits für ihre eigenen RED-Vorlagen dokumentiert haben, ERRATA
  dort). Stattdessen die BEREITS im Repo etablierte Technik verwendet
  (`fakeBeansOnPath`/`fakeBeansConflict`/`fakeBeansEchoIfMatch`,
  `editor_test.go`): ein echtes ausführbares Fake-`beans`-Skript vorn auf
  `PATH`, das je nach Test gezielt scheitert/erfolgreich ist/Argumente
  spiegelt. `TestRenameTagCmdContinuesPastOneFailure` geht dabei über den
  bean-eigenen Sketch hinaus: statt nur die finalen Zahlen zu prüfen,
  protokolliert das Fake-Skript JEDE versuchte Bean-ID in eine Log-Datei —
  belegt explizit, dass b1 UND b3 trotz b2s Fehlschlag versucht wurden
  (nicht nur aus der Endsumme gefolgert).
- **Zwei zusätzliche Guards in `renameTagCmd`, über den bean-eigenen TDD
  hinaus (Implementer-Entscheidung, hart aus D13/D14 hergeleitet, kein
  Wortlaut-Widerspruch):**
  1. `idx == nil` → Zero-Value-No-Op (mirrort `collectTagCounts`s eigenen
     `if idx != nil`-Guard) — ohne diesen Guard hätte ein
     Pre-Load/Test-Fixture-Modell (`m.idx` unset) einen Nil-Pointer-Crash
     beim Rename-Submit ausgelöst (`WithTag` dereferenziert `idx.beans`).
  2. `oldTag == newTag` (D14s Dedupe-Exklusion des eigenen alten Namens
     lässt genau das zu — ein PO, der den Rename-Input unverändert
     bestätigt) → Zero-Value-No-Op OHNE jeden `SetTags`-Aufruf. Begründung:
     `SetTags`s eigener dokumentierter Resolver („derselbe Tag in add UND
     remove → remove gewinnt", `mutations.go` I2) hätte sonst bei einem
     harmlosen Re-Confirm-Tastendruck den Tag STILLSCHWEIGEND von JEDEM
     betroffenen Bean ENTFERNT — ein echter, vermeidbarer Datenverlust-Bug
     für den riskantesten Task dieser Kette. Eigener Regressionstest
     (`TestRenameTagCmdSameOldNewNameSkipsSweepNoDestructiveSetTags`, Fake
     scheitert bei JEDEM Aufruf — ein Bug hätte das sofort sichtbar
     gemacht).
- **Smoke-Setup abweichend von der wörtlichen bean-Vorgabe („3-4 Beans mit
  Tag „alt" anlegen"):** ein direkter `beans create --tag ...` hätte JEDEN
  Sweep-Aufruf in einen echten CONFLICT laufen lassen (upstream-ETag-Quirk,
  s. Smoke-Sektion oben, bereits vor diesem Task in
  `internal/data/testrepo_test.go` dokumentiert) — kein T5-Bug, aber ohne
  Umgehung hätte der Happy-Path-Smoke NIE einen sauberen Erfolg zeigen
  können. Erst separat getaggt (create ohne Tag, dann EIN `update --tag`),
  danach lief der Sweep sauber durch (extern per `beans list`/Frontmatter
  verifiziert). Der continue-on-error-Pfad selbst wurde in der allerersten
  (verworfenen) Runde live beobachtet: alle 3 SetTags-Aufrufe schlugen
  gleichmäßig mit CONFLICT fehl, der Sweep brach NICHT ab (Registry hatte
  bereits umbenannt, `.beans-tags.yml` zeigte den neuen Namen, obwohl kein
  einziger Bean-Sweep-Call griff) — ein zusätzlicher, ungeplanter
  Live-Beleg für D13s „Registry zuerst, Sweep unabhängig/best-effort".
- **`applyLoaded` baut `tagMgmtRows` nicht reaktiv neu (dokumentiertes,
  kein-Bug-Verhalten, live im Smoke beobachtet):** direkt nach einem
  erfolgreichen Rename zeigt die BEREITS OFFENE Tag-Page kurzzeitig noch
  die ALTEN Counts (der Registry-Save-Teil des Batches rebuildet
  `tagMgmtRows` aus dem zu diesem Zeitpunkt noch VOR-Sweep-`m.idx`) — erst
  ein erneutes Page-Open (D03) zeigt den frischen Stand. Kein
  Akzeptanzkriterium verlangt einen Live-Refresh der bereits offenen Page;
  dokumentiert für T7/eine mögliche Fast-Follow-Notiz.
- Sonst keine Abweichungen vom Bean-Body.

## Notes for T7

- **`applyLoaded` aktualisiert `m.idx`, aber NICHT `m.tagMgmtRows` reaktiv**
  (s. Deviations oben) — eine bereits offene Tag-Page zeigt nach einem
  Rename (oder jedem anderen externen Reload, z. B. `ctrl+r`/Watcher)
  veraltete Counts, bis sie verlassen und neu geöffnet wird. Kein Bug (D03
  deckt nur „frisch bei JEDEM Page-Open"), aber ein dokumentierter,
  günstig nachrüstbarer Fast-Follow-Kandidat (T7 kann das an den PO
  weiterreichen, ähnlich Q01-Q03 im Epic-Body).
- **Upstream-ETag-Quirk bei Tag-bei-Create (nicht T5-spezifisch, aber
  hier live reproduziert):** ein Bean, dessen `tags:`-Block SCHON BEI
  `beans create --tag ...` gesetzt wird, hat einen `list`-berichteten ETag,
  der nicht mit dem übereinstimmt, was `update --if-match` intern prüft —
  jeder nachfolgende `SetTags`/`update`-Aufruf (nicht nur Rename) schlägt
  dadurch mit einem echten CONFLICT fehl, bis EIN erfolgreiches `update`
  den Bean „repariert". Bereits vor T5 in
  `internal/data/testrepo_test.go`s Doc-Stamp dokumentiert, hier zusätzlich
  empirisch am Rename-Sweep reproduziert — relevant für JEDEN Task, der
  künftig echte Beans für einen Smoke-Test per `--tag` bei `create`
  anlegt (T3/T4/T6 tggten bestehende, bereits „reparierte" Repo-Beans, nie
  frisch-mit-Tag-erzeugte — daher nie aufgefallen).
- **`renameTagCmd`s beide Guards (`idx == nil`, `oldTag == newTag`) sind
  generisch genug, falls ein künftiger Task denselben Sweep-Mechanismus
  wiederverwenden wollte** (z. B. ein Bulk-Retag-Feature) — keine
  T5-spezifische Kopplung.
- E10 ist mit T5 inhaltlich vollständig für D01-D15 (T1-T5 completed, T6
  bereits completed laut vorherigem Stand, s. `bt-pqq3`) — T7 sollte den
  vollen Voll-Gate-Beleg + design-spec.md §16 + Q01-Q03-Weiterreichung wie
  in `epic-E10-plan.md` Task 7 beschrieben durchführen.
