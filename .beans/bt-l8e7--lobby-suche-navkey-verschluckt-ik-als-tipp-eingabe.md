---
# bt-l8e7
title: 'Lobby-Suche: navKey verschluckt i/k als Tipp-Eingabe'
status: in-progress
type: bug
priority: low
created_at: 2026-07-16T21:29:12Z
updated_at: 2026-07-17T08:08:31Z
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
