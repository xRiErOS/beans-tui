---
# bt-sohl
title: T7 — Abschluss (Voll-Validierung, Spec, Epic to-review)
status: completed
type: task
priority: normal
created_at: 2026-07-16T15:44:39Z
updated_at: 2026-07-16T19:36:46Z
parent: bt-362n
blocked_by:
    - bt-r92i
    - bt-604w
    - bt-1lsu
    - bt-y9my
    - bt-pqq3
---

T7 — Abschluss (Epic `bt-362n`). `blocked_by` T2, T3, T4, T5, T6 (deckt
T1 transitiv ab, da T2 UND T6 beide an T1 hängen — mirrort E9s T9-Muster
„Blatt-Tasks decken innere Tasks transitiv ab"). Keine Feature-Code-
Änderungen erwartet — reine Validierung + Doku + beans-Pflege (Muster
`epic-E9-plan.md` Task 9).

## Schritte

1. **Voller Regressionslauf:** `command go build -o bin/bt .` ·
   `command go vet ./...` · `gofmt -l .` (leer) · `command go test
   ./... -race` · `command go test ./... -short` (2×) · voller Lauf OHNE
   `-short` (2×) · alle Golden-Testfunktionen (`TestTreeGolden|
   TestBacklogGolden|TestChromeGolden` + jede in T2/T6 evtl. neu
   angelegte Suite) mit `-count=2`. Beleg im bean-Body unter
   „Voll-Gate-Beleg" (Laufzeiten, Zusammenfassung, `git status
   testdata/` leer).
2. **Design-Spec-Konsistenz:** `docs/plans/v1-port/design-spec.md`
   bekommt einen NEUEN Abschnitt „§16 — Tag-Management-Page (v1.1, bean
   `bt-6oyy`, Epic `bt-362n`)" — dokumentiert D01-D15 aus dem Epic-Body
   in Spec-Form (Persistenzformat, Page-Layout, Suggest-Mode-Sortierung,
   Keymap-Ergänzung `RenameTag`/`e`), UND die Entity-Mapping-Tabelle in
   §4 (Zeile „Tags: … Tag-Manager-CRUD entfällt (Tags entstehen/sterben
   implizit)") wird korrigiert — diese Aussage ist ab jetzt SUPERSEDED
   (Tag-Manager-CRUD existiert jetzt, s. §16) — mirrort wie PF-14 eine
   ähnliche „Superseded"-Markierung für die entfernte Review-Cockpit-Zeile
   gesetzt hat (§5), NICHT die alte Zeile stillschweigend löschen.
   Gegenprüfen: `hangingIndentWrap`-artige geteilte Helfer (hier:
   `collectTagCounts`s erweiterte Signatur, `saveTagDefsCmd`/
   `tagRenameDoneMsg`s Ein-Task-Ursprung T3/T5) stimmen mit dem
   TATSÄCHLICHEN Code-Stand nach T1-T6 überein, nicht nur mit diesem
   Plan-Dokument.
3. **§9 Scope-Tabelle:** „OUT (v1, bewusst): Tag-Manager-CRUD (Tags
   implizit)" wird ebenfalls als superseded markiert/verschoben (gleiche
   Begründung wie Schritt 2).
4. **Q01-Q03 aus dem Epic-Body:** im Abschluss-Kommentar (bean-Body)
   explizit als weiterhin offen, nicht blockierend, an den PO
   weiterreichen (Epic-Review-Ritual, mirrort E9s eigene offene
   PO-Punkte-Sektion in `bt-ntoz`).
5. **Feature-bean `bt-6oyy`:** Status-Konsistenz prüfen — der Planner
   hat es bereits zu Planungsbeginn auf `in-progress` gesetzt und auf das
   Epic verwiesen (Body-Append); T7 verifiziert NUR, dass dieser Verweis
   noch stimmt (kein PO-Gate hier, `bt-6oyy` bleibt `in-progress` bis der
   PO die Epic-Abnahme separat gibt — Analogie zu E9s eigenem Epic-Status
   `to-review`, NIE `completed` durch einen Agenten).
6. **Epic-Ritual:** `bt-362n` bekommt Tag `to-review` (Agent setzt NIE
   `completed`); T1-T6 auf `completed` verifizieren (Lücken explizit im
   Report benennen, falls eine Task-Kette an dieser Stelle noch offen
   ist).
7. **`docs/SSTD.md`:** Pointer-Update nur falls nötig (prüfen +
   dokumentieren, mirrort E9 T9 Schritt 5).
8. Commit `docs(release): E10-Abschluss — Epic to-review, T1-T6-Status,
   Design-Spec §16`.

## Akzeptanz-Checkliste

- voller Lauf grün (Build/vet/gofmt/`-race`/`-short`×2/voll×2/Goldens
  `-count=2`) · `bt-362n` trägt `to-review`, nicht `completed` · T1-T6
  alle `completed` (Lücken explizit benannt) · design-spec.md §16 neu +
  §4/§9 als superseded markiert (nicht stillschweigend gelöscht) · Q01-Q03
  im Report an den PO weitergereicht · `bt-6oyy`-Body-Verweis auf das
  Epic verifiziert konsistent · `docs/SSTD.md`-Konsistenz geprüft · kein
  unentdeckter Golden-Drift · `git status --porcelain` am Ende leer
  (keine Scratch-/Test-Artefakte aus T1-T6 übrig).

## PRELUDE aus T5-Review (2026-07-16, F01/F02, low)

Als erster eigener Commit (string+test-only):
- **T5-F01:** Dedupe-Fehlermeldung in keyTagMgmtInput (view_tag_management.go:452-453) sagt 'tag already defined: X' auch wenn X nur eine FREIE Zeile ist — faktisch falsch. Auf neutrales 'name already in use: X' präzisieren (beide Fälle korrekt); betroffene Tests nachziehen. Das implizite Verbot 'Rename auf existierenden freien Namen' selbst NICHT ändern (= Merge-Semantik, PO-Frage Q04 im Epic).
- **T5-F02:** bt-y9my Deviations-Abschnitt um eine Zeile ergänzen: Commit-Titel wurde gegenüber Checklisten-Wortlaut auf ≤50 gekürzt ('feat(tui): E10 Tag-Definition umbenennen (Rename)', 49).

## Summary (T7-Abschluss, 2026-07-16)

Preludes als Commit `85f158f` (`fix(tui): T5-F01 Dedupe-Fehlertext
präzisiert` — Fehlertext neutral `name already in use: X`, drei Tests
pinnen den exakten Wortlaut; T5-F02 bt-y9my-Deviations-Nachtrag).
design-spec.md §16 „Tag-Management (E10)" + §4/§9-Superseded-Markierungen
(PF-14-Muster, nichts gelöscht) als Commit `5ab40e1` — alle
§16-Behauptungen gegen den tatsächlichen Code-Stand gegengeprüft
(`collectTagCounts(idx, defined)`, `saveTagDefsCmd(c, defs, refindName)`
in messages.go, `tagRenameDoneMsg`, `keys.RenameTag` auf `e`,
`tagManagementLocalBindings` = Up/Down/NewTag/Delete/RenameTag/Back,
Marker-Glyph `✓`). Voll-Gate grün (s. u.), T1-T6 alle `completed`
verifiziert, Epic `bt-362n` per Tag `to-review` an den PO übergeben
(Status NICHT verändert), `bt-6oyy` unangetastet (`in-progress`,
Epic-Verweis konsistent), SSTD-Pointer gültig (keine Änderung nötig).

## Voll-Gate-Beleg / Test-Output

- `command go build -o bin/bt .` → OK
- `command gofmt -l .` → leer
- `command go vet ./...` → leer
- `command go test ./... -short -count=2` → alle Pakete ok (tui 18.7s)
- `command go test ./... -count=2` (voll ×2) → alle Pakete ok (tui 278.0s)
- `command go test ./... -race -count=1` → alle Pakete ok (tui 143.3s), Exit 0
- Goldens `-count=2` OHNE `-update`: `TestChromeGolden`/`TestTreeGolden`/
  `TestTreeGoldenDeterministic`/`TestBacklogGolden`/
  `TestBacklogGoldenDeterministic` je 2× PASS ·
  `git diff --stat -- internal/tui/testdata/` leer
- `beans check` → All checks passed
- `git status --porcelain` → leer (bin/bt git-ignored, kein
  `.beans-tags.yml`-Rest aus T1-T6)

RED/GREEN-Beleg T5-F01 (Test-Schärfung zuerst, gegen unveränderten Code):

```
--- FAIL: TestKeyTagMgmtInputRejectsDuplicateAgainstExistingRows
    ... got active=true err="tag already defined: already-there"
--- FAIL: TestKeyTagMgmtInputRejectsDuplicateAgainstFreeRowToo
    ... got active=true err="tag already defined: free-tag"
--- FAIL: TestKeyTagMgmtInputRenameRejectsDuplicateAgainstOtherExistingName
    ... got active=true err="tag already defined: bravo"
```

Nach dem Fix: alle drei PASS (plus voller Lauf, s. o.).

## Smoke

n/a — T7 trägt keinen Feature-Code: die Preludes sind string+test+doku-only
(der geänderte Fehlertext ist durch drei Unit-Tests exakt gepinnt; der
Render-Pfad des Inline-Fehlers war bereits durch T3s
80-Spalten-Truncation-Test abgedeckt); jede E10-Funktion hatte ihren
eigenen tmux-Smoke (120+80) in T1-T6. Der Voll-Gate-Lauf (voll ×2, -race,
Goldens ×2) ist der End-to-End-Nachweis dieser Runde.

## Akzeptanz-Checkliste (abgehakt)

- [x] voller Lauf grün (Build/vet/gofmt/`-race`/`-short`×2/voll×2/Goldens `-count=2`)
- [x] `bt-362n` trägt `to-review`, nicht `completed`
- [x] T1-T6 alle `completed` (keine Lücken)
- [x] design-spec.md §16 neu + §4/§9 als superseded markiert (nicht gelöscht)
- [x] Q01-Q03 an den PO weitergereicht (+ Q04, s. Notes)
- [x] `bt-6oyy`-Body-Verweis aufs Epic verifiziert konsistent (bleibt `in-progress`)
- [x] `docs/SSTD.md`-Konsistenz geprüft (Pointer gültig, keine Änderung nötig)
- [x] kein unentdeckter Golden-Drift
- [x] `git status --porcelain` am Ende leer

## Deviations/ERRATA

- Die drei Dedupe-Tests asserten jetzt den EXAKTEN Fehlertext statt nur
  `err != ""` — bewusste Test-Schärfung als RED-Träger für T5-F01, keine
  Verhaltensänderung (die Ablehnung selbst bleibt, Q04 = PO-Frage).
- `-short`×2/voll×2 als `-count=2` in je EINEM Lauf ausgeführt (umgeht
  zusätzlich den Test-Cache, strenger als zwei getrennte Läufe).
- Sonst keine Abweichungen vom Bean-Body.

## Notes (an den PO — Epic-Review-Ritual, offen + nicht blockierend)

- **Q01:** Destruktiver Delete-Modus (Tag zusätzlich von jedem Bean
  strippen, GitHub-Label-Semantik)? v1.1-Stand: Registry-only (D12).
- **Q02:** Definierte Tags mit Count 0 als „Aufräum-Kandidat" markieren?
  v1.1-Stand: schlichte Count-0-Anzeige (D09).
- **Q03:** Picker-`n` (B14/E8) künftig auch die Registry befüllen?
  v1.1-Stand: B14 unverändert (T6).
- **Q04** (aus T5-Review): Rename auf existierenden freien Namen = Merge
  erlauben? v1.1-Stand: per Dedupe abgelehnt; Empfehlung nächste Iteration
  (eigene Confirm-UX, Populations-Vereinigung irreversibel).
- **Fast-Follow-Kandidat** (T5-Deviations, kein Bug): eine bereits offene
  Tag-Page zeigt nach Rename/externem Reload alte Counts bis zum
  nächsten Page-Open (D03 deckt nur „frisch bei jedem Open").
- `bt-6oyy` bleibt `in-progress` — Abnahme-Entscheid liegt beim PO.
