---
# bt-l8e7
title: 'Lobby-Suche: navKey verschluckt i/k als Tipp-Eingabe'
status: todo
type: bug
priority: low
created_at: 2026-07-16T21:29:12Z
updated_at: 2026-07-16T21:29:12Z
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
