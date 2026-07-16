---
# bt-z4b1
title: 'Edit-Modell: ''e'' wird Ganz-Bean-$EDITOR, ''enter'' bleibt Feld-Kaskade (D01)'
status: completed
type: task
priority: normal
created_at: 2026-07-16T06:45:42Z
updated_at: 2026-07-16T08:33:00Z
parent: bt-tct9
---

E9 Task 2 — deckt D01 aus bean bt-tct9 (inkl. B01, dessen Ist-Beschreibung zu D01). Quelle:
design-spec.md §15 PF-17 (Abschnitt D01, vollständige Herleitung dort — dieser bean-Body
ist die verkürzte Umsetzungs-Anleitung). Ist-Code: internal/tui/update.go (keyNodeAction's
Editor-Case, keyDetailFocus's enter-auf-BODY-Sonderfall = E8-B10), internal/tui/editor.go
(openBodyEditor, editInEditor/prepareEditor/readEditorResult), internal/data/client.go
(run/List/Search als Vorbild für einen neuen Read), internal/data/mutations.go (update/
SetTags/SetBlocking als Vorbild für den kombinierten Diff-Call), internal/tui/types.go
(model-Felder editorTarget/editorETag). Kein blocked_by — unabhängig von jedem anderen
E9-Task (eigene Dateien/Funktionen, keine Überschneidung mit B03/B04/B05/B06/F01).

## D01 — Edit-Modell: `e` wird der Ganz-Bean-`$EDITOR`, `enter` bleibt reine Feld-Kaskade

PO verbatim: "'enter' öffnet im details-view die forms für das edit eines Feldes, 'e'
öffnet egal an welcher Stelle das gesamte bean in $EDITOR." Supersedet E8-B10
(design-spec.md §15 PF-16 Tabelle, B10-Zeile: "e/enter auf BODY kontextsensitiv").
B01 (Ist-Beschreibung): "Ich kann weiterhin nur mit enter status,type,priority,tags
bearbeiten" — Soll-Verhalten ist genau diese D01-Umsetzung, kein separater Fix.

Zwei getrennte Zuständigkeiten, ab jetzt ohne Überschneidung:

1. `enter` bleibt AUSSCHLIESSLICH die PF-5-Feld-Kaskade (Section→Feld→Overlay/Form/Jump),
   unverändert — AUSSER: der E8-B10-Sonderfall "enter auf [2] BODY öffnet $EDITOR" entfällt
   ersatzlos (B10-Revision). `keyDetailFocus`s `if m.secCursor == bodySectionIdx { ... }`-
   Zweig (update.go, direkt vor dem generischen `len(secs[m.secCursor].fields) > 0`-Guard)
   wird ENTFERNT — BODY hat keine `.fields`, der generische Guard fällt danach automatisch
   auf den No-Op zurück (identisch zum Vor-E8-Zustand).
2. `e`/`ctrl+e` (EINE `keys.Editor`-Bindung, unverändert im Keymap: `"e","ctrl+e"`) öffnen
   ab sofort UNBEDINGT denselben neuen `openBeanEditor(b)`-Pfad — egal ob Detail-Fokus
   aktiv ist, egal welche Sektion/welches Feld gerade selektiert ist, egal ob aus Tree oder
   Backlog gedrückt. `keyNodeAction`s bisherige Verzweigung (`if msg.String() == "ctrl+e" ||
   (m.detailFocus && m.secCursor == bodySectionIdx) { openBodyEditor(b) } else {
   openEditTitleForm(b) }`) wird ERSETZT durch einen unconditional Call — es gibt nach
   dieser Revision nur noch EINEN Editor-Pfad, nicht mehr zwei konvergente ("ctrl+e-
   Sonderpfad vereinheitlichen", PO/D01-Wortlaut). `e`/`ctrl+e` öffnen NIE MEHR das
   Titel-Edit-Form (`openEditTitleForm` bleibt nur noch über `enter` auf dem `title:`-Feld
   erreichbar, `activateDetailField`s `"title"`-Case UNVERÄNDERT).
   Dispatch-Ort bleibt UNVERÄNDERT (`keyNodeAction`, VOR dem `detailFocus`/`keyBacklog`/
   `keyTree`-Dispatch geprüft, wirkt auf `m.focusedBean()` view-agnostisch) — das erfüllt
   "egal an welcher Stelle" bereits strukturell, keine Dispatch-Order-Änderung nötig.

## Architektur-Vorgabe: Ganz-Bean-Editor (neu zu bauendes Feature, kein bestehender CLI-Weg)

`beans show <id> --raw` liefert bereits EXAKT das reale Datei-Format (verifiziert:
`beans show bt-tct9 --raw` == on-disk `.beans/bt-tct9--*.md` Byte für Byte): YAML-
Frontmatter zwischen zwei `---`-Zeilen (mit `# <id>`-Kommentarzeile, danach `title/
status/type/priority/tags?/created_at/updated_at/parent?/blocking?/blocked_by?`, jeweils
NUR wenn gesetzt) gefolgt vom Markdown-Body. Seed-Text für `$EDITOR` ist dieser Raw-String
UNVERÄNDERT — kein selbstgebautes Markdown-Templating ("beans bleibt die eine Autorität",
design-spec.md §3.1 D02 — die kanonische Serialisierung kommt aus dem CLI).

### 1. Neuer Read: `data.Client.ShowRaw`

```go
// ShowRaw returns id's full markdown representation exactly as `beans show
// <id> --raw` prints it (verified byte-identical to the on-disk .beans/*.md
// file) -- the seed text for the whole-bean $EDITOR (D01, design-spec.md
// §15 PF-17).
func (c *Client) ShowRaw(id string) (string, error) {
    out, err := c.run("show", id, "--raw")
    if err != nil {
        return "", err
    }
    return string(out), nil
}
```

### 2. Neue Frontmatter-Parse-Struktur + kombinierter Update-Call

`gopkg.in/yaml.v3` ist bereits Projekt-Dependency (`internal/config/settings.go`).

```go
// rawBeanFrontmatter is the parse target for a whole-bean $EDITOR round-trip
// (D01) -- the "# <id>" comment line is a YAML comment, yaml.v3 skips it
// automatically. created_at/updated_at/the ID are deliberately NOT fields
// here: beans update has no flag for them (see "Bekannte Grenze" below).
type rawBeanFrontmatter struct {
    Title     string   `yaml:"title"`
    Status    string   `yaml:"status"`
    Type      string   `yaml:"type"`
    Priority  string   `yaml:"priority"`
    Tags      []string `yaml:"tags"`
    Parent    string   `yaml:"parent"`
    Blocking  []string `yaml:"blocking"`
    BlockedBy []string `yaml:"blocked_by"`
}

// parseRawBean splits raw (the $EDITOR's returned content, same shape as
// ShowRaw's seed) at the SECOND "---" delimiter into frontmatter + body,
// yaml-unmarshals the frontmatter. Returns an error for malformed input
// (missing/single "---", invalid YAML) -- caller surfaces it via the same
// recovery-tempfile convention as a CLI VALIDATION_ERROR (see below).
func parseRawBean(raw string) (rawBeanFrontmatter, string, error)
```

`data.Client.UpdateWhole` (mutations.go, mirrort SetTags'/SetBlocking's Single-Etag-No-
Cascade-Konvention — EIN `update`-Call, nicht N sequentielle):

```go
// WholeEditDiff is the field-level diff between a whole-bean $EDITOR
// session's ORIGINAL snapshot and its edited return (D01). nil/false means
// "unchanged, omit this flag" -- UpdateWhole only sends flags for fields
// that actually differ, mirroring update()'s own minimal-args convention.
type WholeEditDiff struct {
    Title, Status, Type, Priority *string
    TagsAdd, TagsRemove           []string
    BlockingAdd, BlockingRemove   []string
    BlockedByAdd, BlockedByRemove []string
    ParentChanged                 bool
    Parent                        string // valid only if ParentChanged; "" = --remove-parent
    Body                          *string
}

// UpdateWhole applies every changed field from a whole-bean $EDITOR
// round-trip in ONE beans-update call (D01) -- same single-etag-no-cascade
// rationale as SetTags/SetBlocking (mutations.go doc-stamps).
func (c *Client) UpdateWhole(id string, diff WholeEditDiff, etag string) error {
    var args []string
    if diff.Title != nil { args = append(args, "--title", *diff.Title) }
    if diff.Status != nil { args = append(args, "--status", *diff.Status) }
    if diff.Type != nil { args = append(args, "--type", *diff.Type) }
    if diff.Priority != nil { args = append(args, "--priority", *diff.Priority) }
    for _, t := range diff.TagsAdd { args = append(args, "--tag", t) }
    for _, t := range diff.TagsRemove { args = append(args, "--remove-tag", t) }
    for _, b := range diff.BlockingAdd { args = append(args, "--blocking", b) }
    for _, b := range diff.BlockingRemove { args = append(args, "--remove-blocking", b) }
    for _, b := range diff.BlockedByAdd { args = append(args, "--blocked-by", b) }
    for _, b := range diff.BlockedByRemove { args = append(args, "--remove-blocked-by", b) }
    if diff.ParentChanged {
        if diff.Parent == "" { args = append(args, "--remove-parent") } else { args = append(args, "--parent", diff.Parent) }
    }
    if diff.Body != nil { args = append(args, "--body", *diff.Body) }
    if len(args) == 0 { return nil }
    return c.update(id, etag, args...)
}
```

Diff-Bau (Tags/Blocking/BlockedBy sind Set-Diffs, mirrort `applyTagPickerDiff`s add/remove-
Berechnung, box_picker_tag.go): vergleiche `rawBeanFrontmatter` gegen den bei `$EDITOR`-Open
eingefrorenen `*data.Bean`-Snapshot (Title/Status/Type/Priority: String-Ungleichheit → Pointer
setzen; Tags/Blocking/BlockedBy: Mengendifferenz; Parent: String-Ungleichheit → `ParentChanged
=true`; Body: String-Ungleichheit → Pointer setzen).

### 3. Editor-Suspend-Wiring (`editor.go`)

`openBodyEditor` wird durch `openBeanEditor` ersetzt (EIN Editor-Helfer, kein Nebeneinander):
da `ShowRaw` ein Subprocess-Call ist (~20-50ms, design-spec §3.1), läuft der Read als
EIGENER `tea.Cmd` VOR dem `tea.ExecProcess`-Suspend (zwei Cmd-Hops, kein synchroner Call im
Update-Pfad) — neue `beanRawLoadedMsg{id, etag string; raw string; err error}` (messages.go),
neuer `showRawCmd(client *data.Client, id string) tea.Cmd`. `keyNodeAction`s Editor-Case ruft
`m.openBeanEditor(b)` (setzt `m.editorTarget = b.ID`, `m.editorETag = b.ETag`, NEU
`m.editorSnapshot = b` [Modelfeld `editorSnapshot *data.Bean`, types.go, neben editorTarget/
editorETag — der VOLLE Bean-Wert zum Öffnungszeitpunkt, nicht nur ID+ETag, weil der Diff jedes
Feld einzeln braucht] und gibt `showRawCmd(...)` zurück); `Update()`s Msg-Switch (update.go)
bekommt einen neuen `case beanRawLoadedMsg:`-Zweig, der bei `err == nil` den `editInEditor`-
Suspend feuert (gleiche `tea.ExecProcess`-Mechanik wie bisher, Suffix weiterhin ".md"), bei
`err != nil` einen Toast zeigt und den Editor-State zurücksetzt.

### 4. Rückweg (`applyEditorFinished`, update.go — umgebaut)

`editorFinishedMsg{content,changed,err}` bleibt der TYP unverändert. Bei `changed==true`:
`content` wird NICHT mehr direkt als `SetBody`-Argument verwendet, sondern:

1. `parseRawBean(content)` — Fehler → Fehlerfall (Schritt 3 unten).
2. Diff gegen `m.editorSnapshot` bauen (s.o.).
3. `client.UpdateWhole(id, diff, etag)` via `mutateCmd` (bestehendes Muster).
4. Fehlerfall (Parse-Fehler ODER CLI-`VALIDATION_ERROR`, z. B. ein von Hand getippter
   ungültiger `status:`-Wert): der Ganz-Bean-Editor ist BEWUSST unconstrained (Freitext-
   YAML statt der constrained Overlays) — Rejection wird NICHT verhindert, sondern
   RECOVERABLE gemacht, mirrort `writeConflictTempFile`s bestehende Konvention: der
   editierte Rohtext wird in eine GEHALTENE Tempdatei geschrieben, ihr Pfad reitet im
   Toast/Status-Zeilen-Text mit. Ein echter ETag-Konflikt läuft weiterhin durch den
   bestehenden `conflictWithRecovery`-Pfad (unverändert, `applyMutationResult`).

### Bekannte Grenze (Dokumentationspflicht, kein Bug)

`created_at`/`updated_at`/die `# <id>`-Kopfzeile sind im Editor-Text SICHTBAR (Teil von
`--raw`), aber NICHT wirksam editierbar — `beans update` kennt dafür keine Flags. Änderungen
daran werden beim Diff-Bau erkannt, aber beim Bauen der `WholeEditDiff` stillschweigend
verworfen (kein Flag existiert, kein `case` dafür in `UpdateWhole`). Im Commit-Body als
ERRATUM/Deviation dokumentieren, kein Implementierungsauftrag — ein CLI-seitiges Feature-Gap.

## TDD-Schritte

1. Failing tests: `internal/data/client_test.go` `TestShowRawReturnsFileFormat` (gegen
   Test-Repo-Fixture, vergleicht `ShowRaw`-Output strukturell gegen erwartete Frontmatter-
   Felder); `internal/data/client_mut_test.go` `TestUpdateWholeSendsOnlyChangedFields` (+
   Fälle: Title-only, Tags-add+remove kombiniert, Parent-clear, Body-only, No-op bei leerem
   Diff → kein CLI-Call); `internal/tui` neue Datei `editor_test.go`-Ergänzungen:
   `TestParseRawBeanRoundTrip` (ShowRaw-Format rein → gleiche Felder raus),
   `TestParseRawBeanRejectsMalformedFrontmatter`; `update_test.go`:
   `TestKeyNodeActionEditorAlwaysOpensBeanEditor` (e UND ctrl+e, aus Tree/Backlog/Detail,
   jede Sektions-/Feld-Ebene → identischer `openBeanEditor`-Aufruf, KEIN Titel-Form mehr),
   `TestKeyDetailFocusEnterOnBodyIsNoOpAgain` (B10-Revision-Regressionstest),
   `TestApplyEditorFinishedBuildsCombinedUpdateWholeCall`,
   `TestApplyEditorFinishedRecoversTempfileOnValidationError` (mirrort
   `conflictWithRecovery`-Test-Stil).
2. `command go test ./... -run "ShowRaw|UpdateWhole|ParseRawBean|EditorFinished|EditorAlwaysOpens|EnterOnBodyIsNoOp"` → FAIL.
3. Implementieren (Reihenfolge: `data`-Paket zuerst [ShowRaw, WholeEditDiff/UpdateWhole,
   parseRawBean — eigenständig testbar ohne TUI], dann `editor.go`
   [openBeanEditor/showRawCmd/beanRawLoadedMsg], dann `update.go`
   [keyNodeAction-Vereinfachung, keyDetailFocus-B10-Revert, applyEditorFinished-Umbau]).
4. Tests grün.
5. Golden-Check: reine Tastatur-/Datenlogik, KEIN Render-Output-Unterschied erwartet —
   Gegenbeleg-Lauf `command go test ./internal/tui/ -run "TestTreeGolden|TestBacklogGolden|TestChromeGolden"` OHNE -update MUSS grün bleiben.
6. `command go test ./... -short` grün (2x), voller Lauf grün, `-race` grün, gofmt/vet leer.
7. Commit `feat(tui): Ganz-Bean-$EDITOR via 'e'/'ctrl+e', enter bleibt Feld-Kaskade (D01)`
   — Body dokumentiert die "Bekannte Grenze" (created_at/updated_at nicht editierbar) als
   ERRATUM/Deviation. Footer `Refs: bt-tct9`.

## Akzeptanz-Checkliste

- [ ] `e`/`ctrl+e` öffnen von JEDER Stelle (Tree, Backlog, Detail — jede Sektion/Feld-Ebene,
      auch ohne aktiven Detail-Fokus) denselben Ganz-Bean-Editor
- [ ] `e`/`ctrl+e` öffnen NIE MEHR das Titel-Edit-Form
- [ ] `enter` auf `[2] BODY` ist wieder No-Op (B10-Revision)
- [ ] `enter` auf title/status/type/priority/tags-Feldern unverändert (PF-5-Kaskade intakt)
- [ ] Seed-Text ist `beans show <id> --raw`, byte-identisch zum on-disk Format
- [ ] Nur GEÄNDERTE Felder landen im kombinierten `beans update`-Call (verifiziert: No-op-
      Save → kein CLI-Call)
- [ ] Validation-Fehler/Parse-Fehler verlieren die PO-Edits nicht (Recovery-Tempfile)
- [ ] ETag-Konflikt läuft weiterhin über den bestehenden `conflictWithRecovery`-Pfad
- [ ] "Bekannte Grenze" (created_at/updated_at nicht editierbar) im Commit-Body dokumentiert
- [ ] Kein Golden ändert sich (Gegenbeleg grün)
- [ ] Voller Testlauf (inkl. -race) grün, gofmt/vet leer


## Akzeptanz-Checkliste (final)

- [x] e/ctrl+e öffnen von JEDER Stelle (Tree, Backlog, Detail — jede Sektion/Feld-Ebene, auch ohne aktiven Detail-Fokus) denselben Ganz-Bean-Editor
- [x] e/ctrl+e öffnen NIE MEHR das Titel-Edit-Form
- [x] enter auf [2] BODY ist wieder No-Op (B10-Revision)
- [x] enter auf title/status/type/priority/tags-Feldern unverändert (PF-5-Kaskade intakt)
- [x] Seed-Text ist beans show <id> --raw, byte-identisch zum on-disk Format
- [x] Nur GEÄNDERTE Felder landen im kombinierten beans update-Call (verifiziert: No-op-Save → kein CLI-Call)
- [x] Validation-Fehler/Parse-Fehler verlieren die PO-Edits nicht (Recovery-Tempfile)
- [x] ETag-Konflikt läuft weiterhin über den bestehenden conflictWithRecovery-Pfad
- [x] "Bekannte Grenze" (created_at/updated_at nicht editierbar) im Commit-Body dokumentiert
- [x] Kein Golden ändert sich (Gegenbeleg grün)
- [x] Voller Testlauf (inkl. -race) grün, gofmt/vet leer

## Summary

D01 umgesetzt: e/ctrl+e öffnen unbedingt das Ganze Bean im $EDITOR — von jeder Stelle (Tree/Backlog/Detail, jede Sektion/Feld-Ebene, auch ohne Detail-Fokus). enter bleibt ausschließlich die PF-5-Feld-Kaskade; der E8-B10-Sonderfall "enter auf [2] BODY öffnet $EDITOR" ist ersatzlos entfernt (Regressionstest). e/ctrl+e öffnen nie mehr das Titel-Edit-Form (nur noch via enter auf title:-Feld erreichbar).

Data-Layer (internal/data): neu `Client.ShowRaw` (`beans show <id> --raw`, reiner Read) und `Client.UpdateWhole` (kombinierter `beans update`-Call, nur geänderte Felder, No-op-Diff → kein CLI-Call).

TUI-Layer (internal/tui): `openBodyEditor` durch `openBeanEditor` ersetzt (EIN Editor-Helfer). Zwei-Cmd-Hop-Design, da ShowRaw ein Subprocess-Read ist: `showRawCmd` liest, `beanRawLoadedMsg` triggert danach den `tea.ExecProcess`-Suspend. Rückweg: `parseRawBean` (gopkg.in/yaml.v3) splittet Frontmatter/Body, `buildWholeEditDiff` vergleicht gegen den bei Öffnen eingefrorenen `editorSnapshot` (neues Modelfeld), `applyEditorFinished` feuert EINEN `UpdateWhole`-Call. Parse-Fehler UND ETag-Konflikte werden beide über Recovery-Tempfiles abgefangen (voller Rohtext, nicht nur Body).

Zusätzlich revidiert: mouseDetailClick's Doppelklick-auf-BODY-Header-Sonderfall (spiegelte denselben jetzt entfernten enter-auf-BODY-Zweig) — ebenfalls zurück auf reinen No-Op, aus demselben "$EDITOR exklusiv über e/ctrl+e"-Grund. Nicht explizit im bean-Text, aber technisch erzwungen (openBodyEditor entfällt) und durch D01s eigene Begründung ("ausschließlich für e") gedeckt.

## Test-Output

RED (data-layer): `client.UpdateWhole undefined (type *Client has no field or method UpdateWhole)`, `undefined: WholeEditDiff` (client_mut_test.go).
RED (tui-layer): `internal/tui/editor_test.go:106:37: undefined: rawBeanFrontmatter`.

GREEN: voller Lauf `command go test ./...` grün (`ok beans-tui/internal/tui 136.418s`), `command go test ./internal/tui/ -race` grün (`ok beans-tui/internal/tui 141.622s`), `gofmt -l .` leer, `command go vet ./...` leer.

Golden-Gegenbeleg: `command go test ./internal/tui/ -run "TestTreeGolden|TestBacklogGolden|TestChromeGolden"` grün ohne -update.

## Smoke

Kein manueller tmux-Smoke durchgeführt (reine Tastatur-/Datenlogik-Änderung, keine Render-/Layout-Änderung — durch Golden-Gegenbeleg abgedeckt). Verhalten end-to-end über echtes model.Update()/Cmd-Ketten getestet (TestKeyNodeActionEditorAlwaysOpensBeanEditor deckt Tree/Backlog/jede Detail-Sektion/-Feld-Ebene für e UND ctrl+e ab, TestApplyEditorFinishedBuildsCombinedUpdateWholeCall verifiziert die exakte Flag-Liste des dispatchten CLI-Calls über einen echten Fake-beans-Subprozess).

## Deviations/ERRATA

1. **Bekannte Grenze (PFLICHT-Dokumentation, kein Bug):** created_at/updated_at/die "# <id>"-Kopfzeile sind im $EDITOR-Text sichtbar (Teil von ShowRaw), aber nicht wirksam editierbar — `beans update` hat dafür keine Flags. Eine Änderung daran wird beim Diff-Bau stillschweigend verworfen (rawBeanFrontmatter parst diese Felder gar nicht erst). CLI-seitiges Feature-Gap, kein Implementierungsauftrag dieses Tasks.
2. **beanRawLoadedMsg ohne etag-Feld:** der bean-Sketch nannte `beanRawLoadedMsg{id, etag string; raw string; err error}`, `showRawCmd(client, id)` selbst kann aber gar kein etag liefern (nicht im Funktions-Sketch) — das Feld wäre also permanent leer geblieben. Bewusst weggelassen (YAGNI): m.editorETag ist bereits bei openBeanEditor eingefroren, ein zweites (leeres) etag-Feld auf der Msg hätte keinen Zweck erfüllt.
3. **mouseDetailClick Doppelklick-auf-BODY-Revert:** s. Summary oben — technisch erzwungen durch die openBodyEditor-Entfernung, inhaltlich durch D01s "ausschließlich für e"-Prinzip gedeckt. Test TestMouseDetailClickDoubleClickOnBodySectionIsNoOpAgain pinnt den No-Op.
4. **ShowRaw "byte-identisch"-Testfixture:** newTestRepo's handgeschriebene Fixture-Dateien haben eine NICHT-kanonische Frontmatter-Feldreihenfolge (tags nach parent statt davor) — `beans show --raw` normalisiert beim Serialisieren, cat't die Datei nicht wörtlich. Der Byte-Identität-Test verwendet daher einen frisch via `client.Create` angelegten Bean (dessen On-Disk-Datei bereits kanonisch ist), nicht die Hand-Fixture — spiegelt den echten Produktions-Fall (Dateien werden immer von `beans` selbst geschrieben).

## Notes for T(n+1)

- Kein bekannter Folgeauftrag aus diesem Task. Falls ein späterer Task created_at/updated_at editierbar machen soll, ist das ein CLI-seitiger Vorlauf (beans-CLI müsste neue --created-at/--updated-at-Flags bekommen) — nicht in diesem Repo lösbar.
- SetBody (data.Client) bleibt bestehen (nicht entfernt) — kein Aufrufer mehr innerhalb beans-tui, aber Teil der generischen Client-API-Oberfläche, YAGNI-Removal nicht Teil dieses Tasks.
