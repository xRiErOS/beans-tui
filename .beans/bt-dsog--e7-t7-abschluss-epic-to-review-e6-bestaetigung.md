---
# bt-dsog
title: E7 T8 — Abschluss (Epic to-review, E6-Bestaetigung)
status: completed
type: task
priority: normal
created_at: 2026-07-15T14:26:51Z
updated_at: 2026-07-15T18:50:06Z
parent: bt-heg9
blocked_by:
    - bt-uyzf
    - bt-m6at
    - bt-t1uy
---

Details/Steps/Akzeptanz: docs/plans/v1-port/epic-E7-plan.md Task 8. Epic to-review setzen, E6-Blocking (bereits gesetzt) bestaetigen, Task-beans T1-T7 auf completed.


## Punkt aus T3-Review (Q01, low)

README.md-Keybinding-Tabelle + Fließtexte sind noch deutsch; PF-7-Scope war explizit nur internal/tui-Produktionscode. Bei der T8/README-Finalisierung entscheiden+umsetzen: Empfehlung KONSISTENT ENGLISCH (Tool-UI ist englisch; README beschreibt UI-Strings — gemischte Sprache wirkt ungepflegt). PO kann widersprechen (Eriks Repo-Doku ist sonst deutsch).


## Prelude aus T7-Review (Mini-Fixes ZUERST, eigener Commit)

- **B01 (docs-only):** epic-E7-plan.md Zeile ~750-752 — doppelte Golden-Checkbox-Zeile (eine [x] mit Beleg, identische [ ] als Artefakt) — Duplikat-Zeile entfernen.
- **I03 (low):** footer_context.go-Kommentar (+ design-spec-Parenthese) behauptet 'mirrors handleKey's dispatch order' — stimmt nicht (handleKey: searchActive VOR filterOpen; Footer: Filter vor Suche). Folgenlos (Staaten nie gleichzeitig), aber Claim korrigieren ODER Invarianten-Test ergänzen.

## Offene PO-Punkte für Epic-Review (mit to-review-Tag hochreichen)

Aus T7 (Reviewer-Einstufung): **I01 medium** — 80-Spalten-Header truncatet q:quit weg (gängiges Terminalmaß; Empfehlung: Header wrappen wie Footer oder Prioritäts-Truncation). **I02 low** — Overlay-Footer restaten Enter/Back trotz sichtbarem Header (gewollte Verstärkung? Sign-off oder Invarianten-Test). Plus aus T7-Implementer: **D01** — Footer-Umfang eng nach Plan (f/X/b/t/a/B/y bewusst raus; PO kann erweitern). Alle in der Epic-Review-Zusammenfassung listen, nicht still passieren lassen.


## E6-Freigabe-Notiz

`bt-wm4w`/`bt-9yvh` sind `blocked_by: [bt-heg9]` (bereits vom PO/D03
gesetzt, per `beans show --json | jq .blocked_by` bestätigt: beide Beans
enthalten exakt `["bt-heg9"]`). Mit dem `to-review`-Tag am Epic (T8 Step 4)
gilt die Kette als implementiert — **E6 kann formal starten.**

`blocked_by` bleibt DENNOCH bis zum PO-Accept bestehen (implementation-plan.md
Zeile 38: „NUR Epic-/Milestone-beans bekommen Tag `to-review` statt
completed — PO-Gate"). Die Kette läuft weiter (Agent kann E6-Tasks
ausführen), das eigentliche Gate sitzt am Milestone/PO-Review, nicht an
`blocked_by` selbst — `blocked_by` wird erst bei PO-Accept des Epics manuell
entfernt (kein Agent-Auto-Unblock).


## Summary

E7-Abschluss vollständig durchgeführt: Prelude (2 Mini-Fixes, B01/I03) im
eigenen Commit, Voll-Validierung (2× voller Testlauf ohne `-short`, `-race`,
Goldens `-count=2`, Build/gofmt/vet), ein tmux-Gesamt-Smoke aller E7-
Features in einem Durchlauf (Scratch-Repo + Temp-HOME, `capture-pane -e`
für Farben), README auf den E7-Stand gebracht UND vollständig ins Englische
übersetzt (Q01 aus T3-Review umgesetzt: Empfehlung „konsistent Englisch"
angenommen, PO kann widersprechen), beans-Pflege (T1-T7 verifiziert
`completed`, Epic `bt-heg9` Tag `to-review` — NICHT `completed`),
Epic-Review-Zusammenfassung (PF-1..14, PF-9 existiert nicht) in `bt-heg9`
angehängt, E6-Freigabe-Notiz in diesem bean angehängt (`bt-wm4w`/`bt-9yvh`
`blocked_by: [bt-heg9]` bestätigt, Kette gilt als implementiert, E6 kann
formal starten — `blocked_by` bleibt bis PO-Accept bestehen).

## Smoke-Matrix

tmux-Gesamt-Smoke, EIN Durchlauf, Scratch-Repo `/tmp/bt-smoke-e7` (Fixtures:
je 1 Milestone/Epic/Feature/Task/Bug, alle 5 Status, alle 5 Prioritäten
abgedeckt, ein Tag `to-review`), `HOME=/tmp/bt-smoke-home` (Repo-Config
isoliert), 220×50 tmux-Pane, `capture-pane -e` für Farbwerte.

| Feature | Ergebnis | Beleg |
|---|---|---|
| Glyphen M/E/F/T/B + Status + Prio, Farben (capture -e) | 🟢 PASS | Alle 5 Typ-Buchstaben (M blue/E mauve/F mauve/T sky/B red), 5 Status-Buchstaben (d blue/t green/i yellow/c+s subtext) und 5 Prio-Glyphen (‼ red bold/! yellow bold/· text/↓ subtext/→ subtext) einzeln mit RGB-Farbcode aus `capture -e` verifiziert |
| Englische UI überall | 🟢 PASS | Filter-Menü, Command-Center, Create-Form, Delete-Confirm, Tag-/Parent-/Blocking-Picker, Settings, Help-Overlay, Lobby — kein deutscher String mehr gefunden |
| Verb-Entity-Palette | 🟢 PASS | „set status/tags/parent/blocking/title", „delete bean", „create bean", „go to backlog/browse/repo picker/settings", „filter facets", „search beans", „reload data" — exakt PF-8-Schema, Fuzzy-Filter „settings" traf korrekt |
| Detail-Kopfblock + Meta-Feldliste (▷/▶, kein Shift) | 🟢 PASS | Kopfblock (bean-id/title/type-status-prio) über Accordion; ▷/▶-Marker an fester Spaltenposition (Gutter-Code: IMMER `▷ ` oder `▶ `, nie ausgelassen — PF-12) |
| Enter-Kaskade + Ziffern + shift+tab | 🟢 PASS | `1`/`3` springen direkt zu Sektionen (auto-expand); `enter` auf Sektion → Feld-Ebene; `enter` auf status-Feld → Value-Menu-Overlay; `enter` auf Relations-Feld → Sprung zur verlinkten Bean (Tree-Cursor folgt); `shift+tab` (BTab) springt deterministisch zurück zu Tree (Border-Fokusfarbe 198;160;246→Grau bestätigt) |
| Pane-Titel weg | 🟢 PASS | Tree/Backlog: erste Zeile ist direkt `⌕ / search`, kein „Tree"/„Backlog"-Titel; Detail: kein „Detail"-Titel, nur Kopfblock |
| Header 7 Globals + kontextsensitiver Footer | 🟢 PASS | Header exakt `ctrl+r:reload ctrl+k:commands p:repos ?:help esc:back enter:open/confirm q:quit`; Footer wechselt kontextsensitiv (Browse/Backlog-lokal → Filter-Menü-lokal → Value-Menu-lokal, je eigene Bindings, keine Globals dupliziert) |
| R = no-op (Cockpit weg) | 🟢 PASS | `R` in Browse: keine Sichtänderung, kein Crash, Help-Overlay zeigt keine Reviews-Bindings mehr |
| Backlog-enter = no-op | 🟢 PASS | `enter` auf Backlog-Zeile: keine Sichtänderung, Fokus bleibt auf Backlog-Liste (Border-Fokusfarbe unverändert) |

## Voll-Gate-Beleg

| Prüfung | Ergebnis | Dauer |
|---|---|---|
| `command go test ./...` Lauf 1 (frisch, kein Cache) | 🟢 PASS | ~136.4s (`internal/tui`) |
| `command go test ./... -count=1` Lauf 2 (Cache explizit umgangen) | 🟢 PASS | ~136.7s (`internal/tui`) |
| `command go test ./... -race` | 🟢 PASS | ~140.2s (`internal/tui`) |
| Goldens `-run Golden -count=2` (5 Funktionen: Chrome/Tree/TreeDeterministic/Backlog/BacklogDeterministic) | 🟢 PASS | 0.37s (je 2×) |
| `command go build -o bin/bt .` | 🟢 PASS | — |
| `command gofmt -l .` | 🟢 leer | — |
| `command go vet ./...` | 🟢 leer | — |

Alle 3 vollen Suite-Läufe + Race-Lauf unabhängig grün (Lauf 2 mit
`-count=1` erzwungen, um Go-Test-Caching auszuschließen — sonst wäre der
„2. Lauf" nur ein Cache-Hit ohne Aussagekraft).
