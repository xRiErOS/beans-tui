---
# bt-7dfj
title: E5 T8 — E5-Abschluss
status: completed
type: task
priority: normal
created_at: 2026-07-15T09:04:38Z
updated_at: 2026-07-15T13:49:34Z
parent: bt-5h4d
---

Ziel: E5-Abschluss-Ritual (implementation-plan.md »Epos-Abschluss«): Tests+
Build grün, beans-Pflege, README-Ergänzung (Settings/Lobby/Yank/Maus kurz
dokumentiert), Commit, SSTD-Pointer falls nötig, NSP-Auto-Handover für E6.

Plan: docs/plans/v1-port/epic-E5-plan.md »Task 8«.

## Akzeptanz
- [x] `command go test ./...` grün (ohne -short, 2x hintereinander), `command
      go test ./... -race` grün, `command go build -o bin/bt .` ok, `command
      gofmt -l .` leer, `command go vet ./...` leer
- [x] Alle Goldens (Chrome/Tree/TreeDeterministic/Backlog/
      BacklogDeterministic + etwaige neue) 2x grün
- [x] tmux-Smoke im Scratch-Repo: Toast (Konflikt sticky bis Klick), Help-
      Overlay, Yank (Clipboard-Inhalt via `pbpaste`/OSC52-Capture
      verifiziert), Maus (Klick+Wheel+Doppelklick), Settings-Form (Editor/
      Accent live), Lobby+Repo-Wechsel (zwei Scratch-Repos), Archiv-Toggle
      -- als Beleg im Commit-Body/Bean-Body dokumentiert
- [x] beans-Pflege: T1-T7-Task-beans auf `completed` (agent-abschließbar),
      bt-5h4d (Epic) bekommt Tag `to-review` (NICHT completed -- PO-Gate,
      implementation-plan.md »Epos-Abschluss«)
- [x] README.md: Kurzabschnitt Settings-Pfad + Keymap-Ergänzung (p/y/?/
      Maus) falls README eine Keymap-Referenz führt (prüfen)
- [x] docs/SSTD.md: Pointer-Update falls Worktree-Weiche/Referenzen sich
      geändert haben (voraussichtlich unverändert)
- [x] Commit `docs: README + E5-Abschluss`
- [x] Skill `ce-nsp-auto`: Handover-Prompt für E6 (Validierung & Release,
      bean bt-zk9p) erzeugen -- Skill ist disabled (Auftrag), stattdessen
      `## E6-Handover-Hinweise` unten im Body (Auftrags-Vorgabe)


## Review-Finding aus T4 (I01, low — Reviewer 2026-07-15) — GESCHLOSSEN

reviewClickRow (mouse.go) dupliziert die Zeilenstruktur von reviewQueueRows als privaten Zähl-Walk (bewusster Trade-off: Golden-Risiko null). Walk- und Render-Reihenfolge sind nur per Kommentar gekoppelt, nicht compile-time. Empfehlung für T8: Table-Test ergänzen, der beide Walks gegen dieselbe Fixture vergleicht (Zeilenindex je Bean identisch), statt Kommentar-Vertrauen.

**Geschlossen in T8:** `TestReviewClickRowMatchesQueueRowsRenderOrder` (`internal/tui/mouse_test.go`) — render-gegründeter Test über `reviewFixtureBeans()` (2 Epic-Gruppen + "(kein Epic)" + Rework), klickt für JEDEN Bean die reale Bildschirmzeile (ID-Substring) an und prüft `reviewClickRow`s zurückgegebenen `flatIdx` gegen `rs.flat()`s Index. RED war nicht möglich (beide Walks aktuell konsistent) — Mutations-Beweis stattdessen durchgeführt: Gruppen-Walk-Reihenfolge in `reviewClickRow` lokal umgekehrt → Test schlug fehl (`ok=false` für tk-a1) → Revert → `git diff` leer verifiziert VOR dem Commit. Commit `test(tui): reviewClickRow-Walk gegen reviewQueueRows gelockt (T4-I01)`.


## Notiz aus T6-Review (I02, low, kein v1-Handlungsbedarf) — DOKUMENTIERT

In-flight repoMetricsCmd-Aufrufe eines früheren Lobby-Open laufen bei erneutem Öffnen weiter (kein Cancel-Guard) — redundante beans-list-Subprozesse, pfadgekeyt, keine Datenverwechslung. Bewusst akzeptiert für v1; bei vielen Repos später Kontext-Cancel. Im README (Known Issues) verankert.


## Notiz aus T7-Review (I01, low, vorbestehend seit E3) — ENTSCHIEDEN

Parent-Picker (EligibleParents, hierarchy.go) + Blocking-Picker (buildBlockingItems) filtern nicht nach showArchived/Status — archivierte/completed/scrapped Beans bleiben als Relationsziele wählbar. Keine Regression, Plan schweigt.

**Entscheidung (T8):** dokumentiert als bewusste Design-Entscheidung — "Picker zeigen bewusst alle gültigen Relationsziele" (README Known Issues). Kein Fix in v1: Beziehungen zu bereits abgeschlossenen Beans (z.B. "blocked by" ein fertiges Bean) bleiben legitim, ein Status-Filter im Picker wäre reine v1-Scope-Erweiterung ohne im Plan/design-spec verankerten Bedarf.


## Summary

E5 (Polish) ist abgeschlossen: Toast-System (inkl. Konflikt-sticky-Regel aus
E3-Übernahme-Pflicht), Help-Overlay `?` (aus der zentralen Keymap generiert),
Yank `y` (OSC52+nativ, Bean-/Epic-Kontext + eigener Review-Stand-Override im
Cockpit), Maus (Wheel/Klick/Doppelklick, devd-D03-Semantik, Toast-Klick-
Vorrang), Settings (`~/.config/beans-tui/config.yaml`+`state.json`,
Editor-Präzedenz `configuredEditor` > `$VISUAL` > `$EDITOR` > `vi`, Accent
live via `theme.SetAccent`), Lobby V1 + Repo-Picker `p` (Watcher-Lifecycle-
Switch: alter Watcher stirbt, neuer startet) und Archiv-Sicht
(`completed`/`scrapped` default-aus, togglebar über die bestehende Facetten-
Infrastruktur). T4-I01 (einziger Code-Zusatz dieses Tasks) geschlossen,
T6-I02/T7-I01 dokumentiert (kein Fix, siehe oben). 384 Testfunktionen
gesamt (`internal/tui`+`internal/data`+`internal/config`+`cmd`). Epic
bt-5h4d trägt jetzt Tag `to-review` (PO-Gate, NICHT completed).

## Smoke-Matrix

tmux, 2 Scratch-Repos (`repo-a` 6→8 Beans über den Durchlauf, `repo-b` 2→3
Beans), Scratch-`HOME` mit eigenem `~/.config/beans-tui/config.yaml`
(`repos:` beide Scratch-Repos, `editor: "true"`, `accent: "#f5a97f"`,
`tree_width: 32`), reales `bin/bt` (Build dieser Session).

| Feature | Ergebnis | Beleg |
|---|---|---|
| Help-Overlay `?` | PASS | Overlay zeigt alle 3 Gruppen (Navigation/Views & Global/Actions) inkl. neuer E5-Bindings (`p`/`y`), Footer "esc/?/q: close"; `esc` schließt |
| Yank `y` (Bean-Kontext) | PASS | `y` auf Milestone-Knoten → Toast "Kopiert: repo-a-y111" → `pbpaste` liefert Markdown-Header+Meta+Children-Tabelle |
| Maus: Klick | PASS | SGR-Klick auf "Standalone Feature"-Zeile setzt Cursor dorthin (Cursor-Bar `▌` wandert) |
| Maus: Wheel | PASS | Wheel-up/-down bewegt Tree-Cursor zwischen Milestone/Standalone Feature |
| Maus: Einzelklick auf geschlossenem Knoten | PASS | expandiert direkt (Chevron `▸`→`▾`, Kind sichtbar) |
| Maus: Einzelklick auf offenem Knoten | PASS | KEIN Collapse, nur Cursor (D03) |
| Maus: Doppelklick auf offenem Knoten (<500ms) | PASS | kollabiert (Chevron zurück auf `▸`) |
| Settings-Form: Editor live | PASS | Form-Submit `editor: vim` → `config.yaml` sofort geschrieben → `ctrl+e` (ohne Neustart) öffnet tatsächlich `vim` statt `$VISUAL`/`$EDITOR`/`vi` |
| Settings-Form: Accent live | PASS | Form-Submit `accent: #89b4fa` → Cursor-Bar-Farbe wechselt sofort von Default-Mauve (`38;2;198;160;246`) auf Blau (`38;2;137;179;250`), kein Neustart |
| Settings-Form: tree_width live | PASS | Form-Submit `tree_width: 40` → bei Terminalbreite 100 (Fenster-Resize zwecks Sichtbarkeit, da Default-220-Breite den Floor durch die 1fr-Formel maskiert) liegt die linke Pane-Grenze auf Spalte 41 — deckt sich mit `clickPaneGeometry(100,30,tw=40)`→lw=39 (Unit-Beleg, gegen `clickPaneGeometry(100,30,tw=0)`→lw=32 als Default-Referenz), ohne Neustart des Prozesses |
| Lobby öffnen (`p`) | PASS | ASCII-Banner "beans", Repo-Liste mit "offen/gesamt"-Metrik (repo-a 5/6, repo-b 2/2 zum Zeitpunkt des Öffnens) |
| Lobby: Repo-Wechsel | PASS | `enter` auf repo-b → Browse zeigt repo-b's eigene Beans, Chrome-Titel "repo-b: Browse" |
| Watcher-Lifecycle-Switch (alt) | PASS | `beans create` extern in repo-a (jetzt alt) NACH dem Wechsel → TUI (auf repo-b) zeigt KEINE Änderung |
| Watcher-Lifecycle-Switch (neu) | PASS | `beans create` extern in repo-b (jetzt aktiv) → TUI reloaded automatisch, neuer Task sichtbar |
| Archiv-Toggle EIN | PASS | `f` → "Archivierte einblenden" → `space` → vorher verstecktes `completed`-Bean "Task Three Done" wird im Tree sichtbar |
| Archiv-Toggle AUS | PASS | erneut `space` → Bean verschwindet wieder (beide Richtungen belegt) |
| Toast: Konflikt sticky | PASS | `ctrl+e` auf "Task One" (Etag beim Öffnen eingefroren) → währenddessen extern `beans update repo-a-ce71 -s in-progress` (bumpt Etag) → Speichern in `$EDITOR` → echter `ErrConflict`, Toast "Konflikt: Bean extern geändert" + Recovery-Tempfile-Pfad |
| Toast: sticky übersteht `ctrl+r` | PASS | manuelles Reload während Toast sichtbar → Toast bleibt |
| Toast: sticky übersteht Watcher-Reload | PASS | externer `beans create` (Watcher-Trigger) während Toast sichtbar → Toast bleibt |
| Toast: Klick dismissed | PASS | SGR-Klick auf Toast-Dot → Toast verschwindet |
| Yank `y` (Review-Stand-Kontext) | PASS | `R` → Cockpit zeigt "Task Two Review" unter "Epic Alpha" → `y` → Toast "Review-Stand kopiert" → `pbpaste` liefert Kopf+Gruppen-Tabelle (NICHT der einzelne Bean-Kontext) |

Alle 18 Punkte PASS, keine Deviationen vom erwarteten Verhalten.

## Known Issues

Im README (`## Known Issues`) verankert:
- Picker (Parent/Blocking) zeigen bewusst alle gültigen Relationsziele
  unabhängig von Archiv-/Status-Sichtbarkeit (T7-I01, entschieden, kein Fix).
- Lobby-Repo-Metriken laufen nicht kontext-gecancelt bei mehrfachem Öffnen
  (T6-I02, akzeptiert für v1).

## E6-Handover-Hinweise

Quelle: `docs/_free-notes/e6-validation-matrix-draft.md` (Entwurf, jetzt mit
einem T8-Update-Absatz ergänzt) — der E6-Planner sollte NICHT bei null
anfangen, sondern diesen Entwurf als Basis nehmen. Wichtigste Punkte aus
T6/T7/T8-Sicht, die der Entwurf (Stand a13c59a, vor T6/T6b/T7/T8) noch nicht
kannte:

1. **US-10 (Live-Reload) und US-14 (Repo-Wechsel) sind jetzt vollständig
   validierbar**, nicht mehr "in Arbeit"/"teilweise" — der Watcher-Lifecycle-
   Switch (`switchRepoCmd`) und die Lobby (`view_lobby.go`) existieren und
   sind sowohl unit-getestet (`TestSwitchRepoCmdStopsOldWatcherStartsNew`
   u.a.) als auch tmux-smoke-belegt (siehe Smoke-Matrix oben, letzte 4
   Watcher-Zeilen).
2. **Blocker aus Schritt 1 des Entwurfs sind ALLE weg:** T6/T6b/T7/T8 fertig,
   E5-Epic (bt-5h4d) trägt `to-review`. E6 (bt-zk9p, `blocked_by: bt-5h4d`)
   kann formal starten, sobald der PO die E5-Review abschließt.
3. **Zwei offene PO-Entscheide bleiben unverändert vom E3/E2-Erbe** (Entwurf
   »Offene PO-Entscheide«-Tabelle): I01 (`esc` in Detail-Fokus No-Op) und I02
   (kein Sort-Modus-Indikator im Backlog) — E5 hat beide NICHT berührt, E6
   sollte sie dem PO explizit vorlegen statt sie stillschweigend fallen zu
   lassen (Entwurfs-eigene Warnung).
4. **Zwei zusätzliche v1-Design-Entscheidungen aus T8 sind jetzt Known
   Issues, nicht offene Fragen:** Picker-Archiv-Filter (T7-I01) und Lobby-
   Metrik-Cancel (T6-I02) — beide im README verankert, E6 muss sie nicht
   mehr triagieren, nur ggf. gegen den PO verifizieren, dass die Doku-Only-
   Entscheidung akzeptiert wird.
5. **384 Testfunktionen gesamt** (Entwurf nannte 364) — die Differenz sind
   T6/T6b/T7/T8-Tests (Lobby/Watcher-Switch/Archiv/T4-I01-Drift-Test), kein
   Kürzen an anderer Stelle.
