---
# bt-y9my
title: T5 — Tag umbenennen + Propagation über alle Beans
status: in-progress
type: task
priority: normal
created_at: 2026-07-16T15:44:35Z
updated_at: 2026-07-16T18:35:14Z
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

- `e` auf definierter Zeile öffnet Rename-Input vorbefüllt mit dem alten
  Namen · `e` auf freier Zeile No-Op · Submit benennt Registry SOFORT um
  · Sweep läuft asynchron, continue-on-error (Regressionstest mit
  simuliertem Teilfehlschlag) · KEINE Cross-Bean-Transaktion vorgetäuscht
  (D13 dokumentiert) · Toast zeigt renamed/failed-Zahlen · `m.idx` lädt
  danach neu · eigener alter Name im Dedupe-Check durchgelassen ·
  `RenameTag`-Binding in `helpGroups()`, Drift-Guard grün · Goldens
  Gegenbeleg grün · tmux-Smoke 120+80 in Scratch-Repo belegt, Scratch
  danach gelöscht · voller Lauf grün · Commit `feat(tui): E10
  Tag-Definition umbenennen + Bean-Propagation`.
