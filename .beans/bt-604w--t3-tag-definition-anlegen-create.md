---
# bt-604w
title: T3 — Tag-Definition anlegen (Create)
status: in-progress
type: task
priority: normal
created_at: 2026-07-16T15:44:30Z
updated_at: 2026-07-16T17:18:22Z
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

- `n` öffnet Freitext-Input · gültiger Name legt Definition an + Liste
  aktualisiert sofort · ungültiger Name → Inline-Fehler, kein Submit ·
  Duplikat gegen bestehende Zeilen abgewiesen · Create berührt KEIN Bean
  (D11, Regressionstest: `m.idx` unverändert nach Create) · `esc` im
  Input verwirft nur den Submodus · Goldens Gegenbeleg grün · tmux-Smoke
  120+80 belegt, Testdatei danach entfernt · voller Lauf grün · Commit
  `feat(tui): E10 Tag-Definition anlegen (Create)`.

## PRELUDE aus T6-Review (2026-07-16, F01, low — Refactor vor Task-Start)

Als erster eigener Commit: `tagRegistryRows` (view_tag_management.go:64-106) unterhält eine eigene, zu `collectTagCounts` (box_picker_tag.go, seit T6 mit defined-Primärschlüssel) strukturell fast identische Zähl-Schleife — echte Duplikation mit Drift-Risiko. Auf `collectTagCounts` umstellen (war in bt-r92i 'D10-ERRATUM-Notiz für T6' und bt-pqq3 'Notes for T3' bereits vorgezeichnet; T2/T6 liefen mit disjunktem Datei-Scope). Verhalten der Page darf sich NICHT ändern (bestehende T2-Tests bleiben grün = Beleg); D09-Sortierung der Page (definiert alpha, frei count-desc) vs. Picker-Sortierung (defined-first global) beachten — die Page-Gruppierung bleibt Page-Logik, nur die ZÄHLUNG wird geteilt.

(T6-F02, nur Doku-Notiz: T6-bean-Zielabschnitt nennt Glyph '●', implementiert ist '✓' aus T2 — sachlich gewollt, Summary korrekt.)
