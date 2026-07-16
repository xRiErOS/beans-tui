---
# bt-ct3k
title: 'Tag-Management: kein Feedback bei e/d auf freier Tag-Zeile'
status: completed
type: bug
priority: normal
created_at: 2026-07-16T20:31:51Z
updated_at: 2026-07-16T21:15:18Z
parent: bt-362n
---

PO-Nebenbefund, US-Review Runde 7 (2026-07-16): e (Rename) und d (Delete)
wirken auf einer FREIEN (unregistrierten) Tag-Zeile als stiller No-Op (laut
Spec D12/D13 by-design), aber der Footer zeigt beide Keybinds unbedingt an,
ohne dass die Zeile als "frei" erkennbar signalisiert wird -- wirkt fuer den
Nutzer wie ein kaputtes Keybind (live reproduziert mit Tag 'smoke', leere
Registry .beans-tags.yml).

## Akzeptanzkriterium (Entwurf)

- Tastendruck e/d auf freier Zeile zeigt einen Toast/Hint, z.B. "Nur definierte
  Tags koennen umbenannt/geloescht werden -- n zum Definieren" -- statt
  stillem No-Op.
- Alternative (Planner entscheidet): Footer-Hint fuer e/d nur zeigen, wenn die
  aktuelle Zeile definiert ist (kontextsensitiver Footer, mirrort andere
  View-lokale Hints).

Quelle: bt-362n US-Review Runde 7.


## PO-Praezisierung (2026-07-16, Runde 8)

PO-Wortlaut fuer den Notification-Text: 'unregistred tag - modification not
possible' -- als Vorlage fuer den Toast/Hint-Text uebernehmen (Planner
final formuliert, Tippfehler 'unregistred' im Original-Zitat).



## Planner-Konkretisierung (2026-07-16)

**Betroffene Stellen:** `openTagMgmtRename` (view_tag_management.go:412-421)
und `openTagMgmtDeleteConfirm` (Zeile 537-548) — beide geben heute bei
`!row.defined` einen stillen `return m, nil` zurück (No-Op, D12/D13-konform,
aber ohne Nutzer-Feedback).

**Fix:** beide No-Op-Zweige rufen stattdessen `m.showToast(toastWarn,
"<Text>", "", nil, false)` (Signatur `overlay_show_toast.go:107`, mirrort
bestehende Warn-Toasts z. B. `update.go:633`/`732`).

**Text (PO-Wortlaut 'unregistred tag - modification not possible'
geglättet, Runde 8):** Titel `"Unregistered tag — modification not
possible"` (Tippfehler korrigiert, Em-Dash statt Bindestrich, mirrort
andere Toast-Titel-Formulierungen wie "Conflict: bean changed
externally"), Context leer ODER `"n to define first"` als Zweitzeile
(Implementer-Entscheidung, beide Varianten toastWarn-konform).

**Akzeptanzkriterium (final):**
- `e` auf freier Tag-Zeile: Toast (toastWarn) statt stillem No-Op, Registry
  unverändert.
- `d` auf freier Tag-Zeile: identischer Toast, Delete-Confirm-Modal öffnet
  NICHT.
- `e`/`d` auf DEFINIERTER Zeile: unverändertes Verhalten (kein Toast,
  normaler Rename-/Delete-Confirm-Flow).
- Test: `keyTagManagement`-Tabellentest erweitert um Free-Row-Cases für
  beide Tasten, Toast-Zustand (`m.toast != nil`, `m.toast.kind ==
  toastWarn`) als Assertion.


## Summary

`openTagMgmtRename` und `openTagMgmtDeleteConfirm` (`internal/tui/view_tag_management.go`)
zeigen bei `!row.defined` jetzt `m.showToast(toastWarn, "Unregistered tag — modification not
possible", "n to define first", nil, false)` statt eines stillen No-Op. Registry/Bean-State
bleibt unangetastet (kein Cmd, kein Save). Verhalten auf definierten Zeilen unverändert.

## Test-Output

RED (vor Fix, 4 Assertions):
```
--- FAIL: TestOpenTagMgmtDeleteConfirmNoOpOnFreeTag (0.00s)
    want a non-nil Cmd (the warn Toast's auto-dismiss timeout) instead of a silent no-op
--- FAIL: TestKeyTagManagementDeleteNoOpOnFreeRowViaFullDispatch (0.00s)
    want a non-nil Cmd (the warn Toast), not a silent no-op
--- FAIL: TestOpenTagMgmtRenameNoOpOnFreeTag (0.00s)
    want a non-nil Cmd (the warn Toast's auto-dismiss timeout) instead of a silent no-op
--- FAIL: TestKeyTagManagementRenameNoOpOnFreeRowViaFullDispatch (0.00s)
    want a non-nil Cmd (the warn Toast), not a silent no-op
```

GREEN (nach Fix): `go test ./... -short` und voller `go test ./...` (ohne `-short`) beide grün,
`go vet ./...` clean, `gofmt -l .` clean.

## Smoke

Echter tmux-Smoke (`bin/bt` im Worktree, Bean bt-6oyy testweise mit freiem Tag "smoke" getaggt,
danach Mutation zurückgesetzt): Tag-Management-Page geöffnet, Cursor auf "smoke" (frei), `e`
gedrückt → Toast "Unregistered tag — modification not possible" / "n to define first" erscheint.
`d` auf derselben Zeile → identischer Toast. `.beans`-Testmutation vor Commit via
`git checkout --` zurückgesetzt.

## Deviations/ERRATA

Keine — Zeilenreferenzen aus Plan/bean (412-421 bzw. 537-548) stimmten mit Ist-Code überein
(minimal verschoben durch Docstrings, funktional identisch).
