---
# bt-0xrb
title: 'Tag-Picker: ETag-Konflikt-Toast abgeschnitten, Ursache unklar'
status: completed
type: bug
priority: normal
created_at: 2026-07-17T09:43:51Z
updated_at: 2026-07-17T11:05:56Z
parent: bt-5uzr
---

NB aus PO-Review E12 Runde 1 (2026-07-17), US-01-Einschränkung. PO wörtlich: siehe bt-362n `## Review 2026-07-17`.

Symptome:
1. Beim Taggen von `lean-stack-n0ly` (Tags smoke, smoke3) via `t`-Picker: Toast `Conflict: bean changed extern...` — Meldung abgeschnitten, erklärt nicht, WARUM das Tagging nicht klappt.
2. `lean-stack-o4c4` lässt sich mit `t` normal taggen — Konflikt ist bean-spezifisch, nicht global.

Zwei Aspekte:
- B (UX): Konflikt-Toast muss vollständig sichtbar sein (Wrap/Mehrzeiler statt Truncate) und handlungsleitend erklären (z. B. „bean wurde extern geändert — Liste aktualisiert, bitte erneut versuchen").
- B (Ursache): Warum hält bt für n0ly einen stalen ETag? Kandidaten: Upstream-Quirk beans 0.4.2 (create --tag liefert stalen ETag, LESSONS-LEARNED E10/4), Watcher-Refresh-Lücke, tatsächliche externe Änderung (lean-stack-Repo hatte uncommittete .beans-Änderungen).

Diagnose-Ergebnis wird hier angehängt (ce-diagnose, 2026-07-17).


## PO-Bestätigung Review R2 (2026-07-17, US-02)

PO wörtlich: „NB: Auch hier: Die Fehlermeldung sollte vollständig angezeigt werden >> Toast muss dynamisch größer werden. Haben wir bei US-01 schon angemerkt."

Konkretisierte Anforderung an den UX-Teil dieses beans: Toast wächst dynamisch (Breite/Mehrzeiler) bis die Meldung VOLLSTÄNDIG lesbar ist — kein Truncate mit `...`. Gilt für alle Toast-Severities, nicht nur Conflict.


## Diagnose-Ergebnis (ce-diagnose, 2026-07-17)

**Root Cause: Upstream-Bug in beans/beancore 0.4.2 — NICHT beans-tui.** Zwei inkonsistente ETag-Quellen:
- `beans list --json` liefert `hash(Bean.Render())` der IN-MEMORY-Repräsentation, in die der Loader beim Laden stille Defaults füllt (`Priority "" → "normal"`, beans-src/pkg/beancore/core.go:205-219).
- `beans update --if-match` validiert gegen `hash(rohe Datei-Bytes)` (core.go:560-576).

Divergiert deterministisch bei jedem bean, dessen Datei kein `priority:`-Feld trägt (bulk-importiert, nie durch beans-Writer gelaufen) — z. B. `lean-stack-n0ly` (list-etag 54ff1e8f… vs. raw 4c728326…). Kein Watcher-Refresh kann das heilen (content-basiert, kein Timing). Self-Heal: erste erfolgreiche Mutation schreibt die Datei kanonisch → deshalb funktionierte `lean-stack-o4c4` (schon einmal mutiert). 9 weitere beans im lean-stack-Repo tragen dieselbe latente Klasse. beans-tui-Mutationspfad (box_picker_tag.go:401-431, update.go:681-698, data/mutations.go) verifiziert korrekt.

**B01a (beans-tui, unabhängig fixbar):** overlay_show_toast.go:146/156 — toastBoxWidth=36 → Titel-Budget 30 Zeichen, „Conflict: bean changed externally" (34) wird geclippt. Zusätzlich übergibt der plain-ErrConflict-Zweig (update.go:290/303) `toastCtx=""` — keine bean-ID/ETag-Details im Toast, anders als der Editor-Conflict-Zweig.

**Fix-Optionen:**
1. B01a in beans-tui: toastCtx für plain-ErrConflict befüllen (mirrort update.go:290-303) + PO-Anforderung „Toast wächst dynamisch" umsetzen; Regressionstests overlay_show_toast_test.go/etag_conflict_test.go.
2. Upstream-Issue bei hmans/beans filen (Update-Validierung muss dieselbe Repräsentation hashen wie list/ETag()).
3. Optional Mitigation in bt (Konflikt-Sonderfall erkennen, handlungsleitende Meldung) — Diagnose-Empfehlung: eher nicht (Komplexität vs. self-healing Quirk).
4. Sofort-Heilung Daten-Seite: betroffene lean-stack-beans einmal kanonisch durchschreiben (trivial, außerhalb bt).


## PO-Entscheid Grilling 2026-07-17 (D04)

1. Upstream-Issue bei hmans/beans wird gefiled (Entwurf → PO-Freigabe → Absenden).
2. Die 10 betroffenen lean-stack-beans werden sofort geheilt (einmalige kanonische Mutation; außerhalb dieses beans).
3. KEINE Konflikt-Sonderfall-Mitigation in beans-tui. Scope dieses beans damit final: B01a — Toast wächst dynamisch bis Meldung vollständig (alle Severities), plain-ErrConflict-Zweig befüllt toastCtx mit handlungsleitenden Details (mirrort update.go:290-303-Zweig). Regressionstests overlay_show_toast_test.go + etag_conflict_test.go.


## Plan-Konkretisierung E13 (2026-07-17)

Plan: `docs/plans/v1-port/epic-E13-plan.md` §„Item 1: Toast wächst dynamisch
+ toastCtx im plain-ErrConflict-Zweig". Reihenfolge: Rang 1 (Toast-Familie,
vor `bt-tm4a`, sequenziell in EINEM Worktree).

**Root Cause (file:line, verifiziert gegen Ist-Code 2026-07-17):**
- `overlay_show_toast.go:146` `const toastBoxWidth = 36` — fixe Breite.
- `overlay_show_toast.go:156` `ansi.Truncate(t.title, innerW-2, "…")` —
  innerW-2=30, „Conflict: bean changed externally" (34 Zeichen) wird
  abgeschnitten.
- `overlay_show_toast.go:160` dieselbe Truncate-Logik für `context`.
- `update.go:275-327` (`applyMutationResult`): plain-ErrConflict-Zweig
  (Zeile 279-303, kein `*conflictWithRecovery`-Match) lässt `toastCtx = ""`
  (Zeile 290) — `err.Error()` (Bean-ID + CLI-Detail aus
  `internal/data/mutations.go:63/75`) wird verworfen.

**Vorgehen (Kurzfassung, Details im Plan):** `toastBoxWidth` durch
content-getriebene Breite ersetzen (geklemmt `[32, min(m.width-4, 70)]`,
Cap=70 Planner-Entscheidung), `ansi.Truncate("…")` durch Wordwrap ersetzen
(alle drei `toastKind`-Severities), plain-ErrConflict-Zweig `toastCtx` mit
`err.Error()` vorbelegen (vor dem `errors.As(&cr)`-Override).

**Akzeptanz (abhakbar, siehe Plan für Volltext):**
- [ ] `toastBoxWidth` content-getrieben, geklemmt `[32, min(m.width-4, 70)]`
- [ ] Kein `ansi.Truncate("…")` mehr für Titel/Context — Wordwrap statt
      Abschneiden
- [ ] Gilt für alle drei `toastKind`-Severities
- [ ] Plain-ErrConflict-Zweig setzt nicht-leeren `toastCtx` aus `err.Error()`
- [ ] Recovery-Pfad-Verhalten (`toastCtx = "Version saved: …"`) unverändert
- [ ] Test-Suite grün, neue Tests in `overlay_show_toast_test.go` +
      `etag_conflict_test.go`
- [ ] tmux-Smoke: Conflict-Repro (`lean-stack-n0ly` via `t`-Picker) zeigt
      vollständige Meldung + Detailzeile, kein Abschneiden

Scope bleibt final D04-gebunden (siehe eigene „PO-Entscheid Grilling
2026-07-17"-Sektion oben): NUR B01a, keine Konflikt-Sonderfall-Mitigation.

## Summary (bt-0xrb, 2026-07-17)

Scope final laut D04 (PO-Entscheid Grilling 2026-07-17): NUR B01a. Toast
wächst jetzt content-getrieben (Breite geklemmt `[32, min(m.width-4, 70)]`),
Titel/Context werden nicht mehr per `ansi.Truncate("…")` abgeschnitten,
sondern mit `wrapText()` (bestehender `view.go`-Baustein, ansi.Wordwrap +
ansi.Hardwrap) auf mehrere Zeilen umgebrochen — `toastGeometry()`s
`h = len(lines)` adaptiert sich dabei automatisch, keine separate
Höhen-Verdrahtung nötig. Gilt für ALLE drei `toastKind`-Severities (Fix im
gemeinsamen `toastBox()`/neue `toastBoxWidth()`-Rendering-Pfad, kein
Conflict-Sonderfall). `update.go`s plain-ErrConflict-Zweig
(`applyMutationResult`, Zeile ~290) setzt `toastCtx` jetzt auf `err.Error()`
vor statt `""` — der `errors.As(&cr)`-Treffer (Editor-Recovery-Pfad)
überschreibt das weiterhin mit `"Version saved: …"`, UNVERÄNDERT. Keine
Konflikt-Sonderfall-Mitigation (D04-Scope eingehalten).

## Test-Output RED→GREEN

RED (vor Implementierung, `go test ./internal/tui/ -run
'TestToastBoxGrowsForLongTitleAcrossAllSeverities|TestPlainConflictSetsNonEmptyToastCtx'
-v`):
- `TestToastBoxGrowsForLongTitleAcrossAllSeverities`: FAIL — `toastBox
  truncated with an ellipsis`, `toastGeometry width = 36, want > 36`
  (kind=1 Warn, kind=2 Error; kind=0 Info lief zufällig durch, da dessen
  Titel/Context in diesem Testlauf knapp unter dem alten Budget blieb —
  Tabellen-Test deckt trotzdem alle drei Severities ab).
- `TestPlainConflictSetsNonEmptyToastCtx`: FAIL — `toast.context is empty,
  want it pre-filled from err.Error()`.

GREEN (nach Implementierung, derselbe Befehl + volle Suite):
- Beide obigen Tests PASS, ebenso die 3 neuen Begleit-Tests
  (`TestToastBoxWidthClampedToCapOnWideContent`,
  `TestToastBoxWidthFloorOnNarrowTerminal`, unverändert
  `TestToastGeometryTopRight`/`TestToastHitTest`/
  `TestConflictToastIsStickyAndSurvivesReload`/`TestEtagConflictSweep`
  — alle PASS, keine Regression).
- `command go build -o bin/bt .` — clean.
- `command go vet ./...` — clean.
- `command gofmt -l` auf allen 4 geänderten Dateien — clean (keine Ausgabe).
- Voller Lauf ohne `-short`: `command go test ./... -count=1` — PASS,
  `internal/tui` 155.144s, Gesamtlauf 2:35.60 total. Keine Golden-Regression
  (`TestTreeGolden` unberührt — dessen Fixture (`goldenTreeModel`,
  tree_golden_test.go) setzt `m.toast` nie, Toast ist NICHT Teil der
  Basis-Goldens — Gegenbeleg wie im Plan gefordert; kein `-update`-Lauf
  nötig, kein golden-Diff).

## Smoke (tmux, echtes TrueColor-Terminal, Session `bt0xrb47423`)

Isolierte Test-Repo-Kopie in Scratch (`lean-stack`-Kopie, NIE die
Produktiv-Repo mutiert — Kopie danach gelöscht). Zwei Live-Repros:

1. **ETag-Konflikt-Repro (wie im Plan vorgesehen) — Gegenbefund
   dokumentiert:** `t`-Picker auf `lean-stack-n0ly` geöffnet, Tag toggled,
   Bean-Datei EXTERN editiert (updated_at geändert, echter Raw-Byte-Diff),
   dann Enter. Ergebnis: KEIN Konflikt — die Mutation ging durch. Ursache
   (Code gelesen, nicht nur vermutet): `bt`s fsnotify-Watcher hat die
   externe Änderung VOR dem Enter bereits eingelesen und den ETag im Model
   aufgefrischt (design decision d, exakt der Fall, den
   `TestConflictAfterWatchReloadUsesFreshETagNoConflict`,
   etag_conflict_test.go, für den value-menu-Pfad bereits gegen einen
   fake-etag-Echo absichert) — ein External-Edit-vor-Enter reicht bei
   laufendem Watcher NICHT für einen reproduzierbaren Live-Konflikt; nur ein
   echtes Timing-Race (Edit exakt zwischen bt's Lese- und Schreib-Aufruf)
   würde greifen, nicht deterministisch skriptbar. Aufgabenstellung erlaubt
   für genau diesen Fall explizit den Alternativ-Pfad (siehe 2).
2. **Alternativ-Pfad (Aufgabenstellung-sanktioniert): ungültige Bleve-Query
   mit langem Text.** `/title:/[a-z/` im Suchfeld getippt → echter
   `beans list --search`-Fehler (`error parsing regexp: missing closing ]`)
   via `applyBleveResult` (update.go) als `toastError` (Titel = volle
   Fehlermeldung) gerendert. Pane-Capture zeigt den Toast über ZWEI Zeilen
   umgebrochen, KEIN „…", Box sichtbar breiter als die alte feste 36-Spalten-
   Breite:

   ```
   ╭────────────────────────────────────────────────────────────────────╮
   │ ● beans list: exit status 1: Error: querying beans: error parsing  │
   │   regexp: missing closing ]: `[a-z`                                │
   ╰────────────────────────────────────────────────────────────────────╯
   ```

   Beweist live: Wachstum (Breite + Mehrzeiler) UND kein Truncate, für
   `toastError` — dieselbe Rendering-Pfad-Logik (`toastBox`/`toastBoxWidth`)
   gilt unit-getestet auch für `toastInfo`/`toastWarn`
   (`TestToastBoxGrowsForLongTitleAcrossAllSeverities`).

## Deviations/ERRATA

- Akzeptanz-Punkt „tmux-Smoke: lean-stack-n0ly ... via t-Picker taggen —
  vollständige Conflict-Meldung sichtbar" konnte NICHT wie im Plan
  formuliert reproduziert werden — s. Smoke-Sektion Punkt 1 für die
  Root-Cause (Watcher-Refresh self-heals den ETag vor Enter, verifiziert
  gegen bestehenden Test `TestConflictAfterWatchReloadUsesFreshETagNoConflict`).
  Der plain-ErrConflict-Zweig selbst ist trotzdem vollständig abgedeckt:
  unit-seitig (`TestPlainConflictSetsNonEmptyToastCtx`, treibt exakt diesen
  Zweig über `applyMutationResult` mit `fakeBeansConflict`) UND live
  (Bleve-Fehlerpfad beweist denselben Rendering-Code). Aufgabenstellung
  sanktioniert diesen Alternativ-Pfad explizit ("Alternativ reicht ein
  langer Titel via anderem Fehlerpfad ... für den Wachstums-Beweis +
  Unit-Test für den Conflict-Kontext").
- `update.go`-Zeilenangaben aus Plan/bean (275-327, Conflict-Zweig 279-303)
  stimmten exakt mit dem Ist-Code überein — kein Drift, keine
  Zeilenkorrektur nötig.
- Beim Live-Repro-Versuch 1 wurde beiläufig festgestellt, dass die im
  Diagnose-Ergebnis erwähnte Upstream-ETag-Divergenz für `lean-stack-n0ly`
  in der Produktiv-Repo bereits geheilt ist (Datei trägt `priority: normal`
  + aktualisiertes `updated_at`, vermutlich der D04-Sofort-Heilungsschritt)
  — separat von dieser Session, nur zur Kenntnisnahme.

## Notes for bt-tm4a

`bt-tm4a` ändert laut Plan `update.go:735`, `box_confirm_create.go:56`,
`overlay_palette.go:240` (alle DREI `createInFlightNote`-Toast-Severity-
Stellen `toastError`→`toastWarn`) — KEINE dieser drei Stellen liegt im
Diff dieses Commits (`bt-0xrb` änderte in `update.go` NUR die Zeilen
~279-303 im Conflict-Zweig von `applyMutationResult`, nicht Zeile 735).
Zeilennummern-Verschiebung durch diesen Commit in den betroffenen Dateien:
- `internal/tui/overlay_show_toast.go`: +62 Zeilen (36→~98 durch die neue
  `toastBoxWidth()`-Funktion + erweiterte Doc-Kommentare) — betrifft
  `bt-tm4a` nicht direkt (andere Datei-Familie), aber falls du zufällig
  `overlay_show_toast.go` mit-liest: die alte Konstante `toastBoxWidth`
  (Zeile 146) ist jetzt eine FUNKTION gleichen Namens, kein Const mehr.
- `internal/tui/update.go`: `applyMutationResult` (vorher Zeile 275-328,
  jetzt 275-335, +7 Zeilen durch den neuen Doc-Kommentar am plain-
  ErrConflict-Zweig) — `update.go:735`s `createInFlightNote`-Guard
  verschiebt sich dadurch um +7 auf **~742**. Bitte gegen Ist-Code
  verifizieren (ERRATUM-Kultur), nicht blind übernehmen.
- `box_confirm_create.go`/`overlay_palette.go`: von diesem Commit NICHT
  angefasst, keine Verschiebung.
