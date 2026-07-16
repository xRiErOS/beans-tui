---
# bt-1lsu
title: T4 — Tag-Definition entfernen (Delete)
status: in-progress
type: task
priority: normal
created_at: 2026-07-16T15:44:30Z
updated_at: 2026-07-16T18:19:13Z
parent: bt-362n
blocked_by:
    - bt-r92i
---

T4 — Tag-Definition entfernen (Delete, registry-only) (Epic `bt-362n`,
D12, D15). `blocked_by` T2 (Page-Grundgerüst). Parallel zu T3 (disjunkte
Funktionen im selben neuen File — mirrort E9s T1/T4-Präzedenzfall
„disjunkte Funktionen, keine Abhängigkeit trotz derselben Datei").

## Ziel

Eine Tag-Definition aus der Registry entfernen — REGISTRY-ONLY (D12):
Beans, die den Tag tragen, behalten ihn (wird wieder „frei"). Confirm
zeigt den LIVE-Verwendungszähler.

## Betroffene Dateien/Symbole

- `internal/tui/types.go` (`model`): neue Felder `tagMgmtDeleteConfirm
  bool`, `tagMgmtDeleteTarget string` (D15 — page-lokales Bool+Ziel-Paar,
  KEIN neuer `overlayID`-Case, mirrort `m.confirmQuit`).
- `internal/tui/view_tag_management.go`:
  - `func (m model) openTagMgmtDeleteConfirm() (tea.Model, tea.Cmd)` —
    liest die aktuell selektierte Zeile (`m.tagMgmtRows[m.tagMgmtCursor.cursor]`),
    No-Op wenn außerhalb Range oder `!defined` (nur definierte Tags
    lassen sich „entfernen" — ein freier Tag hat keine Definition, die
    gelöscht werden könnte; No-Op statt Fehler, mirrort das
    `focusedBean()==nil`-No-Op-Muster quer durchs Repo), setzt
    `tagMgmtDeleteConfirm = true`, `tagMgmtDeleteTarget = row.name`.
  - `func (m model) tagMgmtDeleteConfirmBox() string` — `modalPanel`
    (mirrort `box_confirm_delete.go`s `deleteBox`-Aufbau): Text
    „Delete tag definition '<name>'? Still used by <count> bean(s) —
    they keep the tag, it just won't be prioritized anymore." (Count
    0 → „Not currently used by any bean." — kürzerer Text, kein
    Widerspruch), Footer „enter: delete   esc/n: cancel".
  - `func (m model) keyTagMgmtDeleteConfirm(msg tea.KeyMsg) (tea.Model,
    tea.Cmd)` — `enter`: dispatcht `saveTagDefsCmd(m.client,
    data.RemoveTagDefName(currentDefs, target))` (reuse T3s
    `saveTagDefsCmd`/`tagDefsSavedMsg`-Infra — KEIN zweiter Save-Pfad);
    `esc`/`n`: `tagMgmtDeleteConfirm = false`, kein Save-Call.
  - `composeOverlays`-Äquivalent für die Page: da `viewTagManagement`
    NICHT über den generischen `m.overlay`-Mechanismus läuft (D06), malt
    `viewTagManagement()` selbst das Confirm-Modal zentriert über die
    Liste, WENN `m.tagMgmtDeleteConfirm` (mirrort wie `viewLobby`/
    `viewBacklog` `composeOverlays(out, w, h)` am Ende aufrufen — hier
    reicht ein einfacher `if m.tagMgmtDeleteConfirm { return
    m.composeOverlays(...) }`-artiger Zusatz, DENN `composeOverlays`
    selbst kennt `tagMgmtDeleteConfirm` noch nicht — Planner-Entscheidung:
    `composeOverlays` (`view_browse_repo.go` oder wo es lebt) bekommt
    einen NEUEN Zweig für dieses Feld, exakt wie es bereits `m.confirmQuit`
    kennt — konsistent mit D15s „mirrort confirmQuit"-Rationale bis in
    die Compositing-Ebene).
- `internal/tui/view_tag_management.go` (`tagManagementLocalBindings`):
  T4 hängt `keys.Delete` an (reuse, gleiche Bedeutung „löschen", neuer
  Ort — mirrort T3s `keys.NewTag`-Reuse).
- `internal/tui/update.go` (`keyTagManagement`): neuer Case (NACH dem
  `tagMgmtInputActive`-Vorrang-Check aus T3, da Input und Delete-Confirm
  sich gegenseitig ausschließen — Planner-Entscheidung: Delete-Confirm
  kann nur aus dem Grundzustand geöffnet werden, nicht während der Input
  aktiv ist) `if m.tagMgmtDeleteConfirm { return
  m.keyTagMgmtDeleteConfirm(msg) }`, dann `keybind.Matches(msg,
  keys.Delete): return m.openTagMgmtDeleteConfirm()`.

## TDD (RED zuerst)

```go
func TestOpenTagMgmtDeleteConfirmNoOpOnFreeTag(t *testing.T) {
    m := newModel(nil, "")
    m.view = viewTagManagement
    m.tagMgmtRows = []tagRegistryRow{{name: "free-tag", defined: false, count: 3}}
    nm, _ := m.openTagMgmtDeleteConfirm()
    if nm.(model).tagMgmtDeleteConfirm {
        t.Fatal("want no-op for a free (undefined) tag row")
    }
}

func TestKeyTagMgmtDeleteConfirmEscCancelsWithoutSaving(t *testing.T) {
    m := newModel(nil, "")
    m.tagMgmtDeleteConfirm = true
    m.tagMgmtDeleteTarget = "to-review"
    nm, cmd := m.keyTagMgmtDeleteConfirm(tea.KeyMsg{Type: tea.KeyEsc})
    if nm.(model).tagMgmtDeleteConfirm || cmd != nil {
        t.Fatalf("want cancel with no Cmd, got confirm=%v cmd=%v",
            nm.(model).tagMgmtDeleteConfirm, cmd)
    }
}
```

Weitere Pflicht-Tests: `enter` dispatcht genau EINEN `saveTagDefsCmd` mit
dem Ziel entfernt aus den Defs; ein Regressionstest, der belegt, dass
Beans mit dem gelöschten Tag ihre `Tags`-Liste NICHT ändern (D12 — kein
`SetTags`/`data.Client`-Bean-Mutation-Call in diesem Pfad, nur
`SaveTagDefs`).

## Golden-Strategie

GEGENBELEG (Tree/Backlog/Chrome unverändert). Falls T2 eine eigene
Golden-Suite für die Page angelegt hat: T4 ergänzt optional einen
„Delete-Confirm offen"-Snapshot.

## tmux-Smoke (120 UND 80 Spalten)

Page öffnen → Cursor auf eine DEFINIERTE Zeile (ggf. vorher via T3s `n`
eine anlegen) → `d` → Confirm-Modal zeigt korrekten Count → `enter` →
Zeile verschwindet aus der „Definiert"-Gruppe (taucht ggf. in
„Undefiniert-in-Verwendung" wieder auf, falls Count > 0) → `d` auf einer
FREIEN Zeile → No-Op verifizieren (kein Modal öffnet). Bei 80 Spalten:
Confirm-Text darf umbrechen, aber nicht abgeschnitten werden. Danach
`.beans-tags.yml`/`git status` auf Ausgangsstand zurücksetzen.

## Akzeptanz-Checkliste

- [x] `d` auf definierter Zeile öffnet Confirm mit korrektem Live-Count
- [x] `d` auf freier Zeile No-Op
- [x] `enter` entfernt NUR die Definition, Beans behalten den Tag (D12-Regressionstest)
- [x] `esc`/`n` bricht ohne Save ab
- [x] Confirm-Modal zentriert über der Liste sichtbar
- [x] Goldens Gegenbeleg grün
- [x] tmux-Smoke 120+80 belegt, Testdatei-Reste entfernt
- [x] voller Lauf grün
- [x] Commit `feat(tui): E10 Tag-Definition entfernen (Delete)`

## Summary

Liefert D12 (Delete, registry-only) + D15 (page-lokaler Confirm, kein
`overlayID`-Case). Neue `model`-Felder (`types.go`): `tagMgmtDeleteConfirm
bool`/`tagMgmtDeleteTarget string` -- mirrort `m.confirmQuit`s Bool-Paar
1:1, keine dritte Count-Feld (der Live-Count wird bei jedem Render aus
`m.tagMgmtRows` nachgeschlagen, nicht am Open-Zeitpunkt eingefroren).

`internal/tui/view_tag_management.go` (neu):
- `openTagMgmtDeleteConfirm()` -- liest `m.tagMgmtRows[m.tagMgmtCursor.cursor]`;
  außerhalb Range ODER `!defined` → No-Op (kein `tagMgmtDeleteConfirm`
  gesetzt); sonst `tagMgmtDeleteConfirm=true`, `tagMgmtDeleteTarget=row.name`.
  Kein Cmd hier -- der eigentliche Save feuert erst auf ein bestätigtes
  `enter`.
- `keyTagMgmtDeleteConfirm(msg)` -- `enter`: dispatcht `saveTagDefsCmd(m.client,
  data.RemoveTagDefName(definedTagNames(m.tagMgmtRows), target))` (T3s
  Save-Infra 1:1 wiederverwendet, kein zweiter Save-Pfad, kein
  `SetTags`/Bean-Mutations-Call irgendwo in diesem Pfad -- D12
  regressionsgesichert); `esc`/`n`: schließt Confirm ohne Save, räumt
  `tagMgmtDeleteTarget` auf; jede andere Taste: HANDLED No-Op
  (Full-Capture-Disziplin, exakt das Muster `keyTagMgmtInput` bereits
  etabliert).
- `tagMgmtDeleteConfirmBox()` -- `modalPanel` (Rot, mirrort `deleteBox`s
  Singular/Plural-Disziplin): Titel „Delete tag definition '<name>'?",
  Body „Still used by N beans -- they keep the tag, it just won't be
  prioritized anymore." (Count 0 → „Not currently used by any bean.",
  Count 1 → Singular „1 bean"), Footer „enter: delete   esc/n: cancel".
  Count wird LIVE aus `m.tagMgmtRows` per Namenssuche zum Render-Zeitpunkt
  aufgelöst (mirrort `deleteBox`s eigenes „typ resolves from the LIVE
  index at render time" Doc-Stamp) -- kein separates, beim Open
  eingefrorenes Count-Feld.

`tagManagementLocalBindings()` wächst um `keys.Delete` (Position: vor
`Back`, nach `NewTag`, mirrort T3s eigene Positions-Entscheidung).
`keyTagManagement` prüft `m.tagMgmtDeleteConfirm` NACH dem
`tagMgmtInputActive`-Vorrang-Check (Input und Delete-Confirm schließen
sich gegenseitig aus -- Delete-Confirm kann nur aus dem Grundzustand
geöffnet werden) und bekommt einen neuen `keys.Delete`-Case, der
`openTagMgmtDeleteConfirm()` aufruft.

`composeOverlays` (`view_browse_repo.go`) bekommt einen NEUEN, eigenen
`if m.tagMgmtDeleteConfirm { ... }`-Zweig (NICHT im `switch m.overlay`,
da `viewTagManagement` D06-bedingt nicht über den generischen
`overlayID`-Mechanismus läuft) -- exakt wie es bereits `m.confirmQuit`
kennt, direkt VOR diesem platziert (rein defensive Reihenfolge, die
beiden können praktisch nie gleichzeitig aktiv sein, da `q` während der
Full-Capture-Page nie `requestQuit` erreicht). `viewTagManagement()`
selbst brauchte KEINE Änderung -- es endete bereits in
`m.composeOverlays(out, w, h)` (aus T2).

## Test-Output (RED→GREEN wörtlich)

RED (`command go vet ./internal/tui/`, vor der Implementierung):

```
# beans-tui/internal/tui
# [beans-tui/internal/tui]
vet: internal/tui/view_tag_management_test.go:856:13: m.openTagMgmtDeleteConfirm undefined (type model has no field or method openTagMgmtDeleteConfirm)
```

RED (`command go test ./internal/tui/ -run "TestOpenTagMgmtDeleteConfirm|TestKeyTagMgmtDeleteConfirm|TestKeyTagManagementDelete|TestFullDeleteFlow|TestTagMgmtDeleteConfirmBox|TestViewTagManagementRendersDeleteConfirmCentered|TestTagManagementLocalBindingsIncludesDelete" -v -count=1`, vor der Implementierung):

```
# beans-tui/internal/tui [beans-tui/internal/tui.test]
internal/tui/view_tag_management_test.go:856:13: m.openTagMgmtDeleteConfirm undefined (type model has no field or method openTagMgmtDeleteConfirm)
internal/tui/view_tag_management_test.go:871:15: m.openTagMgmtDeleteConfirm undefined (type model has no field or method openTagMgmtDeleteConfirm)
internal/tui/view_tag_management_test.go:893:13: m.openTagMgmtDeleteConfirm undefined (type model has no field or method openTagMgmtDeleteConfirm)
internal/tui/view_tag_management_test.go:910:9: mm.tagMgmtDeleteConfirm undefined (type model has no field or method tagMgmtDeleteConfirm)
internal/tui/view_tag_management_test.go:910:36: mm.tagMgmtDeleteTarget undefined (type model has no field or method tagMgmtDeleteTarget)
internal/tui/view_tag_management_test.go:911:86: mm.tagMgmtDeleteConfirm undefined (type model has no field or method tagMgmtDeleteConfirm)
internal/tui/view_tag_management_test.go:911:111: mm.tagMgmtDeleteTarget undefined (type model has no field or method tagMgmtDeleteTarget)
internal/tui/view_tag_management_test.go:926:8: mm.tagMgmtDeleteConfirm undefined (type model has no field or method tagMgmtDeleteConfirm)
internal/tui/view_tag_management_test.go:938:4: m.tagMgmtDeleteConfirm undefined (type model has no field or method tagMgmtDeleteConfirm)
internal/tui/view_tag_management_test.go:939:4: m.tagMgmtDeleteTarget undefined (type model has no field or method tagMgmtDeleteTarget)
internal/tui/view_tag_management_test.go:939:4: too many errors
FAIL	beans-tui/internal/tui [build failed]
FAIL
```

GREEN (gleiches `-run`-Filter, nach Implementierung, alle 16/16 grün):

```
=== RUN   TestOpenTagMgmtDeleteConfirmNoOpOnFreeTag
--- PASS: TestOpenTagMgmtDeleteConfirmNoOpOnFreeTag (0.00s)
=== RUN   TestOpenTagMgmtDeleteConfirmOnDefinedRowSetsConfirmAndTarget
--- PASS: TestOpenTagMgmtDeleteConfirmOnDefinedRowSetsConfirmAndTarget (0.00s)
=== RUN   TestOpenTagMgmtDeleteConfirmNoOpWhenCursorOutOfRange
--- PASS: TestOpenTagMgmtDeleteConfirmNoOpWhenCursorOutOfRange (0.00s)
=== RUN   TestKeyTagManagementDeleteDispatchesToOpenConfirm
--- PASS: TestKeyTagManagementDeleteDispatchesToOpenConfirm (0.00s)
=== RUN   TestKeyTagManagementDeleteNoOpOnFreeRowViaFullDispatch
--- PASS: TestKeyTagManagementDeleteNoOpOnFreeRowViaFullDispatch (0.00s)
=== RUN   TestKeyTagMgmtDeleteConfirmEscCancelsWithoutSaving
--- PASS: TestKeyTagMgmtDeleteConfirmEscCancelsWithoutSaving (0.00s)
=== RUN   TestKeyTagMgmtDeleteConfirmNCancelsWithoutSaving
--- PASS: TestKeyTagMgmtDeleteConfirmNCancelsWithoutSaving (0.00s)
=== RUN   TestKeyTagMgmtDeleteConfirmOtherKeysDuringConfirmAreNoOp
--- PASS: TestKeyTagMgmtDeleteConfirmOtherKeysDuringConfirmAreNoOp (0.00s)
=== RUN   TestKeyTagManagementDeleteConfirmCapturesFullDispatchNoLeak
--- PASS: TestKeyTagManagementDeleteConfirmCapturesFullDispatchNoLeak (0.00s)
=== RUN   TestKeyTagMgmtDeleteConfirmEnterFiresSaveTagDefsCmdWithTargetRemoved
--- PASS: TestKeyTagMgmtDeleteConfirmEnterFiresSaveTagDefsCmdWithTargetRemoved (0.00s)
=== RUN   TestFullDeleteFlowUsedTagFallsBackToFreeGroupCountPreserved
--- PASS: TestFullDeleteFlowUsedTagFallsBackToFreeGroupCountPreserved (0.00s)
=== RUN   TestFullDeleteFlowUnusedTagDisappearsEntirely
--- PASS: TestFullDeleteFlowUnusedTagDisappearsEntirely (0.00s)
=== RUN   TestTagMgmtDeleteConfirmBoxShowsLiveCountAndName
--- PASS: TestTagMgmtDeleteConfirmBoxShowsLiveCountAndName (0.00s)
=== RUN   TestTagMgmtDeleteConfirmBoxZeroCountShorterText
--- PASS: TestTagMgmtDeleteConfirmBoxZeroCountShorterText (0.00s)
=== RUN   TestViewTagManagementRendersDeleteConfirmCentered
--- PASS: TestViewTagManagementRendersDeleteConfirmCentered (0.00s)
=== RUN   TestTagManagementLocalBindingsIncludesDelete
--- PASS: TestTagManagementLocalBindingsIncludesDelete (0.00s)
PASS
ok  	beans-tui/internal/tui	0.470s
```

Voller T2/T3/T4/T6-Regressionskorpus (`-run "TagManagement|TagPicker|TagRegistryRows|GoTags|CollectTagCounts"`, 60 Tests inkl. der 16 neuen) grün, keine bestehende Assertion angepasst.

Gates: `command gofmt -l .` leer · `command go vet ./...` leer · `command go test ./... -short -count=1` grün (alle Pakete ok) · `command go test ./... -count=1` grün (Voll-Lauf, `internal/tui` 138.9s, exit 0).

## Golden-Beleg

`command go build -o bin/bt .` erfolgreich, danach `command go test ./internal/tui/ -run "TestTreeGolden|TestBacklogGolden|TestChromeGolden" -count=2 -v` → alle 5 Testfunktionen (inkl. `*Deterministic`-Varianten) PASS, beide Wiederholungen (10/10). `git diff --stat -- internal/tui/testdata/` leer -- Basis-Goldens unverändert (mirrort T2/T3/T6s eigene Entscheidung: keine eigene Golden-Suite für `viewTagManagement()`, stattdessen Render-/Verhaltenstests wie `TestViewTagManagementRendersDeleteConfirmCentered`).

## Smoke

tmux, `TERM=xterm-256color`, Breite 120 UND 80 (Sessions `bt120t4`/`bt80t4`,
`bin/bt` in diesem Repo). Temporäre `.beans-tags.yml` mit 1 benutztem
(`to-review`, real auf 10 Beans) + 1 unbenutztem definierten Tag
(`smoke-test-unused`) angelegt.

**120 Spalten:** `ctrl+k` → „tags" → „go to tags" → `enter` → Page öffnet:

```
▌✓ smoke-test-unused                                                              0
 ✓ to-review                                                                     10
   rejected                                                                       1
   smoke                                                                          1
```

Cursor auf `to-review` (benutztes, definiertes Tag) → `d` → Confirm zeigt
korrekten Live-Count:

```
Delete tag definition 'to-review'?
Still used by 10 beans — they keep the tag, it
 just won't be prioritized anymore.

enter: delete   esc/n: cancel
```

`enter` → Zeile fällt in die Frei-Gruppe zurück, Count erhalten:

```
 ✓ smoke-test-unused                                                              0
▌  to-review                                                                     10
```

`.beans-tags.yml` danach: `tags:\n    - smoke-test-unused` (nur noch das
unbenutzte Tag definiert). `git status --porcelain .beans/` leer (D12
bestätigt -- kein Bean angefasst).

Cursor auf `smoke-test-unused` (unbenutzt, definiert) → `d` → Confirm zeigt
die kürzere Zero-Count-Formulierung:

```
Delete tag definition 'smoke-test-unused'?
Not currently used by any bean.
```

`enter` → Zeile verschwindet GANZ aus der Liste (kein Free-Row-Rest, da
Count 0). `.beans-tags.yml` danach: `{}` (leere Registry).

`d` auf einer freien Zeile (`to-review`, jetzt frei) → No-Op verifiziert
(kein Modal öffnet, Liste unverändert).

Neues Tag `escape-tag` angelegt (`n`), dann `d` → Confirm öffnet → `esc`
→ Confirm bricht ab, KEIN Save (`.beans-tags.yml` unverändert, `escape-tag`
bleibt definiert). Erneut `d` → Confirm öffnet → `n` → identisches
Cancel-Verhalten bestätigt (esc/n-Dual). Erneut `d` → Confirm offen →
`x`/`?`/`ctrl+k` nacheinander gedrückt → alle drei Full-Capture-verschluckt
(Confirm bleibt offen und UNVERÄNDERT, kein Help/Command-Center öffnet) →
`enter` → `escape-tag` (unbenutzt) verschwindet ganz, `.beans-tags.yml`
danach `{}`. Sitzung sauber mit `q`→`enter` beendet (kein `bt`-Prozess mehr
aktiv danach, `pgrep -fl "bin/bt"` leer).

**80 Spalten:** identischer Ablauf (Session-Neustart, `.beans-tags.yml` aus
dem 120-Spalten-Lauf frisch neu angelegt, D03 bestätigt: Page lädt sofort
den aktuellen Stand). `to-review` löschen → Confirm-Text UMBRICHT korrekt
(nicht abgeschnitten, Vorgabe „darf umbrechen, aber nicht abgeschnitten
werden" erfüllt):

```
Delete tag definition 'to-review'?
Still used by 10 beans — they keep the tag, it
 just won't be prioritized anymore.
```

`d` auf der jetzt freien `to-review`-Zeile → No-Op bestätigt.
`smoke-test-unused` löschen → Zero-Count-Text bestätigt, Zeile verschwindet
ganz nach `enter`. Sitzung sauber mit `q`→`enter` beendet, `pgrep -fl
"bin/bt"` leer.

Nach BEIDEN Läufen: `.beans-tags.yml` gelöscht (`rm .beans-tags.yml`),
`bin/bt`-Binary gelöscht (git-ignored ohnehin, `bin/` in `.gitignore`),
`git status --porcelain` zeigt ausschließlich die erwarteten Quelldateien
(kein `.beans/`, keine Registry-Datei-Reste).

## Deviations/ERRATA

- **`tagMgmtDeleteConfirmBox` nutzt `modalPanel`, nicht `modalBox` direkt**
  (wie der Bean-Body explizit vorgibt: „modalPanel (mirrort
  box_confirm_delete.go's deleteBox-Aufbau)") -- `deleteBox` selbst baut
  seinen String manuell und ruft `modalBox` direkt auf; hier stattdessen
  `modalPanel(title, body, footer, ...)` verwendet (mirrort `quitBox`s
  eigenen `modalPanel`-Call), da der Bean-Text `modalPanel` wörtlich nennt.
  Nur der „Aufbau" (Singular/Plural-Disziplin, Rot-Border für eine
  „Delete"-benannte Aktion) ist von `deleteBox` übernommen, nicht der
  konkrete Funktions-Call.
- **Kein separates Count-Feld am `model`** -- der Bean-Body listet nur
  `tagMgmtDeleteConfirm`/`tagMgmtDeleteTarget` als neue Felder; der
  Live-Count wird deshalb bei JEDEM Render von `tagMgmtDeleteConfirmBox`
  aus `m.tagMgmtRows` per Namenssuche aufgelöst (mirrort `deleteBox`s
  eigenes „typ resolves from the LIVE index at render time"-Muster) --
  keine Abweichung vom Bean-Text, aber ein Implementer-Entscheid, der
  explizit dokumentiert wird, da die Alternative (ein drittes Feld) naheläge.
- **`composeOverlays`-Zweig-Position:** direkt VOR `m.confirmQuit`
  eingefügt (nach `m.helpOpen`) -- der Bean-Body/Epic-Text gibt keine
  exakte Position vor, nur „exakt wie es bereits m.confirmQuit kennt".
  Rein defensive Reihenfolge (die beiden Zustände können praktisch nie
  gleichzeitig aktiv sein, D06 verhindert `q` während der Full-Capture-Page).
- Zusätzlich zu den 2 im Bean-Body zitierten RED-Test-Vorlagen (beide
  UNVERÄNDERT übernommen, kompilierten diesmal ohne Assignability-Fix
  nötig) wurden 14 weitere Tests ergänzt (No-Op-Randfälle, volle
  Dispatch-Pfade, Full-Capture-Leak x2, Registry-Roundtrip,
  D12-Regressionsflows für „benutzt→frei" UND „unbenutzt→verschwindet",
  Confirm-Box-Text x2, Compositing, Footer-Binding) -- reine Ergänzung,
  keine Abweichung vom Spezifizierten.

Sonst keine Abweichungen vom Bean-Body.

## Notes for T5

- **`keyTagManagement`s Reihenfolge ist jetzt: `tagMgmtInputActive` →
  `tagMgmtDeleteConfirm` → navKey → der Haupt-switch.** T5 (Rename) nutzt
  DENSELBEN `tagMgmtInputActive`-Pfad wie T3/T4 (kein dritter Vorrang-Check
  nötig) -- Rename öffnet über `openTagMgmtInput("rename", altName)`
  genau wie Create, der bereits bestehende `tagMgmtInputActive`-Check
  bleibt an erster Stelle.
- **`tagManagementLocalBindings()` ist jetzt `{Up, Down, NewTag, Delete,
  Back}`.** T5 muss dort `keys.RenameTag` (laut Plan-Dokument NEUES
  Binding auf „e") ergänzen -- Position analog „vor Back" (Implementer-
  Entscheidung wie bei T3/T4).
- **`saveTagDefsCmd`/`tagDefsSavedMsg`/`applyTagDefsSaved` sind bereits
  generisch genug für T5** -- T4 hat NICHTS daran geändert, nur einen
  dritten Aufrufer (`RemoveTagDefName` statt `AddTagDefName`) hinzugefügt.
  `applyTagDefsSaved` bleibt weiterhin Mode-agnostisch (liest immer frisch
  von Disk, sucht den Cursor per Name) -- T5 profitiert davon exakt wie T4.
- **Merge-Konflikt-Hinweis aus T3s eigenen Notes bestätigt sich NICHT als
  Problem:** T3 landete bereits vollständig (Commit `d2f07ba`) BEVOR T4
  begann -- `tagManagementLocalBindings` wurde linear erweitert (`NewTag`
  dann `Delete`), kein manueller Merge nötig. T5 sollte trotzdem den
  AKTUELLEN Stand der Funktion lesen, nicht den in T3s eigenen Notes
  zitierten Zwischenstand.
- **Confirm-Modal-Rendering-Pfad (`composeOverlays`) ist jetzt der EINZIGE
  Ort, der `m.tagMgmtDeleteConfirm` kennt** -- falls T5 einen eigenen
  Rename-Bestätigungsschritt bräuchte (laut Plan-Dokument NICHT vorgesehen,
  Rename nutzt den Input-Submodus direkt ohne Extra-Confirm), wäre das ein
  eigenes, neues Bool-Paar nach demselben Muster, nicht eine Erweiterung
  dieses Felds.

## Review-Findings Runde 1 (2026-07-16, T4-Review, Verdict CHANGES_REQUIRED)

- **B01 (medium, update.go:459-484 applyTagDefsSaved / view_tag_management.go:512-526 keyTagMgmtDeleteConfirm):** applyTagDefsSaved re-findet den Cursor nach JEDEM tagDefsSavedMsg via strings.TrimSpace(m.tagMgmtInput.Value()). Vor T4 sicher (einziger Aufrufer = Create-Submit). T4s Delete-Confirm ist ein ZWEITER saveTagDefsCmd-Aufrufer, der das Input-Feld nie anfasst — und T3s esc-Abbruch löscht Value() NICHT. Repro: Registry {alpha,bravo} → Create-Input 'bravo' tippen → esc → Cursor auf alpha → d→enter (Delete alpha) → Cursor springt auf bravo statt Default. FIX (Empfehlung Reviewer, sauberer Weg): tagDefsSavedMsg um explizites refindName-Feld erweitern — Aufrufer geben den Namen explizit mit (Create: der neue Name; Delete: Ziel-Default), statt implizit m.tagMgmtInput.Value() zu lesen. Macht 'Notes for T5: mode-agnostisch' real. Minimal-Alternative: SetValue("") im Delete-Enter-Case. + Regressionstest: 'Create abgebrochen → unabhängiges Delete → Cursor darf nicht auf stale Text springen' (RED mit altem Code belegen).
- **I01 (low):** Singular-Zweig count==1 ('Still used by 1 bean') in tagMgmtDeleteConfirmBox ohne dedizierten Test — Bug-Klasse hat im Repo Präzedenz (box_confirm_delete.go I02-Doc-Stamp). Test analog zu count==0/count==7 ergänzen.
- **I02 (low, view_tag_management_test.go:553-580):** TestKeyTagMgmtInputCapturesEveryKeyNoLeak prüft laut Kommentar 'd darf Delete-Confirm nicht öffnen' via m.overlay — die D15-Implementierung berührt overlay nie: Assertion tot (Mutations-Probe belegt). m.tagMgmtDeleteConfirm-Check ergänzen/ersetzen.

Fix-Runde beim selben Implementer; Re-Review beim selben Reviewer.
