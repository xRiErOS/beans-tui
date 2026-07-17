---
# bt-l8e7
title: 'Lobby-Suche: navKey verschluckt i/k als Tipp-Eingabe'
status: completed
type: bug
priority: low
created_at: 2026-07-16T21:29:12Z
updated_at: 2026-07-17T09:48:59Z
---

Bestehender Bug, entdeckt im bt-9ipw-Review (2026-07-16): `keyLobby`
(`view_lobby.go:351`) routet `navKey()` VOR dem Textinput-Update — ein
Repo-Query, der mit "i" oder "k" beginnt, wird als Navigation verschluckt statt
ins Suchfeld zu gehen. Nicht durch bt-9ipw eingeführt (dessen Typeahead fängt
bewusst nur rohe KeyUp/KeyDown ab — richtige Lösung, taugt als Vorbild).

Fix-Skizze: Lobby-Query-Submodus wie `keyTagInput` (bt-9ipw) auf rohe
`tea.KeyUp`/`tea.KeyDown` umstellen, i/k literal durchlassen. Regressionstest:
Query "ide" landet vollständig im Input.

Quelle: Reviewer-Finding non-blocking #2 aus bt-9ipw-Review.


## Plan-Konkretisierung E12 (2026-07-17)

Plan: `docs/plans/v1-port/epic-E12-plan.md` §„Item 3: Lobby-Suche
verschluckt i/k". Reihenfolge: Rang 3 (nach `bt-9ipw`/`bt-81f0`, Kontext von
`bt-9ipw`s identischem Fix-Pattern noch frisch).

**Root Cause:** `keyLobby` (`view_lobby.go:350-359`) prüft
`navKey(msg.String())` (Zeile 351) VOR der Textinput-Weiterleitung (Zeile
400-408). `navKey` (`keymap.go:232-244`) aliast `keys.Up`/`keys.Down` auf
Buchstaben inkl. `i`/`k` — ein Query, der mit "i"/"k" beginnt, wird als
Navigation verschluckt.

**Vorgehen:** exaktes Vorbild bereits im Repo: `keyTagInput`
(`box_picker_tag.go:313-335`, aus `bt-9ipw`) fängt die ROHEN
`tea.KeyUp`/`tea.KeyDown`-`KeyType`s ab statt über `navKey()`s
Buchstaben-Alias zu gehen. `keyLobby`s `switch navKey(msg.String())`-Block
(Zeile 351-359) durch `switch msg.Type`-Prüfung auf `tea.KeyUp`/
`tea.KeyDown` ersetzen (mirrort `keyTagInput` 1:1) — jede andere Taste
fällt weiter zur bestehenden `m.repoSearch.Update(msg)`-Zeile durch.

**Akzeptanz:**
- [ ] Query "ide" (oder jedes mit i/k beginnende Wort) landet vollständig im
      Suchfeld, kein verschlucktes Zeichen
- [ ] Pfeiltasten navigieren die Repo-Liste weiterhin unverändert
- [ ] Regressionstest analog `TestTagInputArrowKeysDoNotLeakIntoTypedText`
      (z. B. `TestLobbyQueryDoesNotSwallowArrowAliasLetters`)
- [ ] Test-Suite grün

## Summary

Root cause bestätigt (`view_lobby.go:350-359`, Zeilen wie im Plan-Erratum
erwartet leicht gedriftet, aber identisch adressiert): `keyLobby` prüfte
`navKey(msg.String())` VOR der Textinput-Weiterleitung. `navKey` (keymap.go)
aliast `keys.Up`/`keys.Down` auf `i`/`k` (vim-Stil) -- ein Repo-Query, der
mit `i`/`k` beginnt, wurde als Navigation verschluckt statt ins Suchfeld zu
gehen.

Fix: `keyLobby`s `switch navKey(msg.String())`-Block durch `switch msg.Type`
auf `tea.KeyUp`/`tea.KeyDown` ersetzt -- exaktes 1:1-Mirror von
`keyTagPicker` (`box_picker_tag.go`, bt-9ipw). Jede andere Taste (inkl.
literal getippter "i"/"k") fällt weiter zur bestehenden
`m.repoSearch.Update(msg)`-Zeile durch.

## Test-Output

RED (vor Fix, `TestLobbyQueryDoesNotSwallowArrowAliasLetters`):

```
=== RUN   TestLobbyQueryDoesNotSwallowArrowAliasLetters
    view_lobby_test.go:90: repoQuery = "de", want "ide" -- i/k must stay literal, not be swallowed as up/down
--- FAIL: TestLobbyQueryDoesNotSwallowArrowAliasLetters (0.00s)
```

GREEN (nach Fix, plus Gegen-Test für unveränderte Pfeiltasten-Navigation):

```
=== RUN   TestLobbyQueryDoesNotSwallowArrowAliasLetters
--- PASS: TestLobbyQueryDoesNotSwallowArrowAliasLetters (0.00s)
=== RUN   TestLobbyArrowKeysStillNavigateRepoList
--- PASS: TestLobbyArrowKeysStillNavigateRepoList (0.00s)
```

Voller Lauf (ohne -short, Pflicht vor Commit):

```
ok  	beans-tui	[no test files]
ok  	beans-tui/cmd	0.725s
ok  	beans-tui/internal/clip	[no test files]
ok  	beans-tui/internal/config	0.392s
ok  	beans-tui/internal/data	3.346s
ok  	beans-tui/internal/theme	1.469s
ok  	beans-tui/internal/tui	154.505s
```
Gesamtdauer 2:34.97 (real), exit 0. `go vet ./...` clean, `gofmt -l` clean
(beide berührten Dateien).

## Smoke

tmux (`bin/bt` aus diesem Worktree, eindeutiger Session-Name `btl8e729168`):
Lobby (`p`) geöffnet, "ide" getippt -- Filterfeld zeigt vollständig `⌕ ide`,
kein verschlucktes Zeichen:

```
│                                              ⌕ ide                                               │
```

Pfeiltasten-Navigation nicht separat live gesmoked (keine Repos im
Test-config.yaml), aber durch `TestLobbyArrowKeysStillNavigateRepoList`
(neu) UND die bereits bestehende Test-Suite (`TestLobbyFilterNarrowsBySearch`
etc.) strukturell abgedeckt -- gleiche Argumentationslinie wie bt-9ipws
eigener Test-vs-Smoke-Split.

## Deviations/ERRATA

- Plan nennt `view_lobby.go:350-359` als Fix-Zeilen -- Ist-Code stimmte
  exakt überein, keine Drift.
- Plan verweist auf `keyTagInput`/`TestTagInputArrowKeysDoNotLeakIntoTypedText`
  als Vorbild -- durch bt-9ipws eigene Konsolidierung (D01, "EIN
  konsolidierter Modus statt zwei getrennter") sind Funktion und Test
  inzwischen umbenannt zu `keyTagPicker`/
  `TestTagPickerArrowKeysDoNotLeakIntoTypedText` (box_picker_tag.go bzw.
  box_picker_tag_test.go). Fix-Pattern identisch übernommen, nur die
  Referenz-Namen sind gedriftet -- dokumentiert wie vom Auftrag verlangt
  (ERRATUM-Kultur), kein Scope-Abweichung.
- `box_picker_tag.go`s eigener Doc-Kommentar (Zeile ~248) referenziert
  bt-l8e7 noch als "EXISTING BUG, not a precedent" -- dieser Kommentar wird
  durch diesen Fix technisch überholt, aber NICHT angepasst (out of scope:
  Plan grenzt Item 3 explizit auf EINE Datei, `view_lobby.go`, ein; die
  Nachbardatei gehört zu bt-9ipws Scope).


## Review 2026-07-17 (E12, US-03)

US-03 · Lobby-Suche verschluckt i/k nicht mehr · r — PO-Begründung betrifft nicht diesen Fix, sondern fehlende Repo-Discovery in der Lobby (PO wörtlich in bt-5uzr Review-Sektion). Fix selbst unstrittig, bean bleibt completed; Nacharbeit → bt-d3ps.
