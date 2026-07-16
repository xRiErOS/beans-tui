---
# bt-1lsu
title: T4 — Tag-Definition entfernen (Delete)
status: in-progress
type: task
priority: normal
created_at: 2026-07-16T15:44:30Z
updated_at: 2026-07-16T17:47:14Z
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

- `d` auf definierter Zeile öffnet Confirm mit korrektem Live-Count · `d`
  auf freier Zeile No-Op · `enter` entfernt NUR die Definition, Beans
  behalten den Tag (D12-Regressionstest) · `esc`/`n` bricht ohne Save ab
  · Confirm-Modal zentriert über der Liste sichtbar · Goldens Gegenbeleg
  grün · tmux-Smoke 120+80 belegt, Testdatei-Reste entfernt · voller Lauf
  grün · Commit `feat(tui): E10 Tag-Definition entfernen (Delete)`.
