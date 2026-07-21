---
# bt-bszz
title: 'Relations-Panel: Box-Badge (r) statt nur Footer-Keybind'
status: completed
type: task
priority: normal
created_at: 2026-07-20T17:50:57Z
updated_at: 2026-07-20T17:56:15Z
parent: bt-vy1q
---

# Kontext (Epos bt-vy1q — Rest siehe Epos, DRY)

PO-Befund (2026-07-20): Das **Relations-Panel** in der Detail-Pane (Box-Form,
`BT_BOXFORM=1`) zeigt seinen Keybind `r` **nicht als Box-Badge** im Rahmen,
sondern nur im Footer. Abweichung vom Design: **Box-Badge = Keybind** (Glossar,
D07). Jedes boxed field trägt seinen Keybind im Rahmen — Relations bricht das.

PO nannte `view_browse_project.go`; die Datei existiert nicht (Glossar §Views:
real `view_browse_repo.go`). Der eigentliche Fundort ist der Box-Form-Renderer.

## Fundort (verifiziert)

`internal/tui/box_detail_form.go:179`:
```go
panelBox("Relations", relationsBody, "", width, on(boxFormRowRelations))
```
Dritter Arg (`hotkey`) ist leer → `panelBoxWith` rendert leeren Bottom-Badge.
Vergleich Body-Panel (Zeile 178) trägt `"e"` (im Top-Border, weil Body scrollt).

Keybind-Quelle: `internal/tui/keymap.go:200`
`Blocking: WithKeys("r"), WithHelp("r","Relations")` → Badge muss `(r)` sein.

## Nicht-Ziel / bewusste Abgrenzung

- **History-Panel bleibt ohne Badge.** History hat KEINEN Aktivierungs-Keybind;
  `[`/`]` (HistoryBack/Forward, keymap.go:219f) sind Stack-Navigation, nur im
  Fullscreen wirksam, kein Panel-Hotkey. Badge nur wo Keybind existiert.
- Body-Badge (`e`, Top-Border) NICHT anfassen (bt-oox1, PO #4).

## Akzeptanz (abhakbar)

- [ ] `box_detail_form.go:179`: dritter Arg `""` → `"r"` (Relations-Panel).
- [ ] Golden-Test schreiben/erweitern, der den `(r)`-Badge im Relations-Bottom-Border
      pinnt (Box-Form). Test MUSS vor dem Fix RED sein (RED→GREEN im bean zitieren).
- [ ] Goldens `detail_boxform.golden` + `browse_boxform.golden` regenerieren:
      `command go test ./internal/tui -run Golden -update`, danach `git diff` je Golden
      prüfen — Diff darf NUR den neuen `(r)` im Relations-Bottom-Border zeigen,
      sonst nichts.
- [ ] Voller Lauf grün: `command go test ./...` (nicht `-short`, vor Commit Pflicht).
- [ ] `command go build -o bin/bt .` grün.
- [ ] Commit: `fix(boxform): Relations-Panel zeigt Box-Badge (r)` · Footer `Refs: <diese-id>`.

## Hinweise

- Golden-Default byte-identisch ×2 — nach `-update` bewusst-Diff, sonst kein Update.
- Kein Footer-/Wrap-Change → kein tmux-Smoke Pflicht (CLAUDE.md). Unit-Golden deckt es.
- ERRATUM-Kultur: Snippet oben prüfen, Abweichung als ERRATUM ins bean.


## Summary

Fundort bestätigt (kein ERRATUM — Zeile 179 in box_detail_form.go matchte das
Bean-Snippet exakt). Fix: dritter Arg von panelBox("Relations", ...) von ""
auf "r" geändert (box_detail_form.go:179). Neuer RED-Test
TestDetailBoxFormRelationsHotkeyIsInBottomBorder in box_panel_test.go pinnt
das (r)-Badge im Relations-Bottom-Border am Call-Site (mirror von
TestDetailBoxFormBodyHotkeyIsInTopBorder). Beide Box-Form-Goldens
(detail_boxform.golden, browse_boxform.golden) regeneriert; Diff je Datei
zeigt exakt eine geänderte Zeile (Relations-Bottom-Border).

## Test-Output

RED (vor Fix):
```
=== RUN   TestDetailBoxFormRelationsHotkeyIsInBottomBorder
    box_panel_test.go:203: Relations panel bottom border carries no (r) badge: "╰──────────────────────────────────────────────────────────╯"
--- FAIL: TestDetailBoxFormRelationsHotkeyIsInBottomBorder (0.00s)
FAIL
```

GREEN (nach Fix):
```
=== RUN   TestDetailBoxFormRelationsHotkeyIsInBottomBorder
--- PASS: TestDetailBoxFormRelationsHotkeyIsInBottomBorder (0.00s)
PASS
ok  	github.com/xRiErOS/beans-tui/internal/tui	0.542s
```

Golden-Ebene (zusätzlicher RED-Beleg vor -update): TestDetailBoxFormGolden
und TestBrowseBoxFormGolden schlugen nach dem Fix fehl (alte Goldens ohne
Badge), TestBrowseBoxFormDefaultOffMatchesTree blieb grün (Flag-OFF-Pfad
unberührt). Nach `go test ./internal/tui -run Golden -update` alle drei
grün, zweiter Lauf (ohne -update) ebenfalls grün (Stabilität ×2 bestätigt).

Voller Lauf: `go test ./...` — alle Pakete ok (internal/tui 152.441s).
`go build -o bin/bt .` und `go vet ./...` beide grün.

## Golden-Diff

detail_boxform.golden (1 Zeile geändert):
```
-╰──────────────────────────────────────────────────────────────────────────────────────────────────╯
+╰──────────────────────────────────────────────────────────────────────────────────────────╯ (r) ───╯
```
(ANSI-Farbcodes im echten Diff um die Segmente, hier zur Lesbarkeit entfernt)

browse_boxform.golden (1 Zeile geändert):
```
-│╰──────────────────────────────────────────────────────────╯│
+│╰──────────────────────────────────────────────────────────╯ (r) ───╯│
```
(vereinfachte Darstellung; siehe git-Commit 4aa4ae0 für den vollen
ANSI-Diff — beide Dateien je genau eine geänderte Zeile, sonst
byte-identisch)

## Deviations/ERRATA

Keine. Fundort/Zeile/Fix stimmten exakt mit dem Bean-Snippet überein.

## Notes for Reviewer

- Neuer Test liegt in box_panel_test.go (nicht box_detail_form_test.go),
  weil er den bereits dort etablierten Call-Site-Test
  TestDetailBoxFormBodyHotkeyIsInTopBorder spiegelt (gleiche Datei, gleiches
  Muster: Body oben/Top-Border vs. Relations unten/Bottom-Border).
- History-Panel und Body-Panel (Top-Border-Badge) bewusst unangetastet,
  wie im Bean gefordert.
- Golden-Diffs oben sind zur Lesbarkeit ANSI-bereinigt; der tatsächliche
  git diff (siehe Commit 4aa4ae0) zeigt die vollen Farbcodes, aber ebenfalls
  nur je eine geänderte Zeile pro Golden-Datei.
