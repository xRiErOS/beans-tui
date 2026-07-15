---
# bt-2af1
title: E7 T2 — Theme/Glyphen-Umstellung (PF-6)
status: completed
type: task
priority: high
created_at: 2026-07-15T14:23:51Z
updated_at: 2026-07-15T15:31:45Z
parent: bt-heg9
blocked_by:
    - bt-wmtb
---

Details/Steps/Akzeptanz: docs/plans/v1-port/epic-E7-plan.md Task 2. Typ/Status/Prioritaet: Buchstaben+Glyphen statt Icons/Woerter, neue Farbtabelle (design-spec.md Section 15 PF-6).


## Prelude aus T1-Review (2 Mini-Punkte, zuerst, Reviewer 2026-07-15)

- **I01 (low):** internal/data/mutations.go RejectReview (YAGNI-behalten, kein Caller) nutzt noch das alte Tag-Literal 'rework' — Doc-Kommentar ergänzen: Literal ist stale ggü. Tag-Trio (design-spec §5, to-review/accepted/rejected), NICHT reaktivieren ohne Update. Nur Kommentar, kein Code.
- **I02 (low):** docs/plans/v1-port/epic-E7-plan.md Task-1-Abschnitt 'bewusst NICHT löschen' um EpicAncestor ergänzen (1 Zeile — war Supervisor-Vorgabe, fehlt im Plan-Text).


## Summary

PF-6 (design-spec.md §15) umgesetzt: Typ/Status/Priorität werden nicht mehr über
geometrische Unicode-Icons bzw. den gemeinsamen `◉`-Glyph unterschieden, sondern über
Buchstaben + Farbe. `internal/theme/theme.go`: `statusColor`-Map auf neue Farben
(todo Green, completed/scrapped Subtext); NEUE `statusLetter`-Map (d/t/i/c/s) ersetzt
`statusGlyph`/`statusGlyphASCII` (entfernt, DD2-176-Doc-Kommentar durch PF-6-Begründung
ersetzt); `priorityColor` umgeschrieben (high Yellow, low/deferred Subtext); NEUE
`priorityGlyph`/`priorityGlyphASCII`-Maps, `Priority()` liefert jetzt den Glyph statt
des Worts (Bold für critical/high bleibt); `priorityIconGlyph` mit demselben
Fallback-Mechanismus wie Status/Type (`fallbackGlyph()` für unbekannte Werte).
`internal/theme/icons.go`: `typeIcon`-Map → Buchstaben M/E/F/T/B, `typeColor` auf neue
Farben (milestone Blue, feature Mauve, task Sky), `typeIconASCII` komplett entfernt
(Buchstaben bereits ASCII/EAW-Narrow). `TypeIcon`/`TypeStyle`-Signaturen unverändert
(Drop-in an allen 3 Call-Sites: view_browse_repo.go, view_browse_backlog.go,
view_detail_bean.go). Grep-Sweep (alte Glyphen ◉⬢✦⯅⯁‼↓→· repo-weit) bestätigte: die
einzigen echten Konsumenten waren genau die von T1s Notes-for-T2 genannten 3 Dateien —
alle übrigen Treffer (↑↓←→ in keymap.go/view_lobby.go/view.go, `·` als Bullet-Separator
in forms_shared.go) sind Keybinding-Pfeile/Hint-Bullets, kein Type/Status/Priority-Bezug,
unangetastet.

Collateral (compiler-/test-getrieben, außerhalb der T2-Dateiliste):
`internal/tui/view_detail_bean_test.go`s `TestBeanSectionsMetaRendersStatusTypePriorityTags`
prüfte bisher `strings.Contains(meta, "critical")` gegen `theme.Priority("critical")` —
das Wort existiert seit PF-6 nicht mehr im Output (nur noch der Glyph `‼`), Assertion auf
den Glyph umgestellt, Kommentar ergänzt (Verweis auf theme_test.go/PF-6). `StatusStyle`/
`TypeStyle`-Zeilen in derselben Funktion (rendern die rohen Wörter, nicht über
StatusIcon/TypeIcon) sind unbetroffen, unverändert grün.

## Test-Output

RED (vor Implementierung, gegen die neuen theme_test.go-Erwartungen, Implementierung
temporär via `git stash` zurückgesetzt zur Gegenprobe): `TestStatusColorMapping`,
`TestAsciiFallback`, `TestTypeIconAllTypes`, `TestPriorityColorMapping` alle FAIL
(Farb-/Glyph-Diffs exakt wie erwartet, z.B. `StatusIcon("draft") = "◉", want contains
letter "d"`, `typeColor["milestone"] = "#f5a97f", want "#8aadf4"`,
`Priority("critical") = "critical", want contains glyph "‼"`). Nach `git stash pop`
(Implementierung wiederhergestellt): `command go test ./internal/theme/... -v` → alle 5
Funktionen PASS (inkl. `TestSetAccentOverridesThenNoOpOnEmptyOrInvalid`, unverändert).

Nach Golden-Regen + Collateral-Fix: `command go test ./... -short` → PASS (alle
Pakete). `command go test ./...` (voller Lauf, ohne -short) → PASS, 136.4s
(`internal/tui`, inkl. der 7 langsamen huh-drive-Tests + Golden-Suite). `command go vet
./...` → leer. `gofmt -l internal cmd .` → leer. Goldens `-count=2` nach Update:
`TestChromeGolden`/`TestTreeGolden`/`TestTreeGoldenDeterministic`/`TestBacklogGolden`/
`TestBacklogGoldenDeterministic` alle 2x PASS, stabil.

## Golden-Diffs

- **tree.golden** (SHA1 vorher `da65ba397ee75360635b253e6ae3000e6ef1d836` → nachher
  `728f4ec0cdcba56299858c17cf760c81d267fab6`): 4 Bean-Zeilen geändert. Zeile 8
  (gld-mlst, in-progress/milestone/high): `◉ ⬢` → `i M` (Status-Farbe Yellow
  unverändert, Type-Farbe Peach→Blue). Zeile 9 (gld-epic, todo/epic): `◉ ✦` → `t E`
  (Status Text→Green, Type Mauve unverändert). Zeile 10 (gld-tsk1, todo/task,
  CURSOR-Zeile `▌`): `◉ ⯅` → `t T`, aber die Zeile ist per D08-Cursor-Konvention
  (view_browse_repo.go, unverändert seit vor PF-6) komplett Accent-eingefärbt — die
  Buchstaben sind sichtbar, ihre Einzel-Farben sind wie zuvor durch den Cursor-Akzent
  überschrieben (kein PF-6-Verhalten, pure Cursor-Darstellung). Zeile 12 (gld-orph,
  draft/task): `◉ ⯅` → `d T` (Status Blue unverändert, Type Blue→Sky). Nur
  Glyph-/Farbspalten betroffen, keine Spaltenbreiten-Verschiebung (siehe EAW-Analyse
  unten) — Box-Ränder/Trennlinien byte-identisch zur alten Version.
- **backlog.golden** (SHA1 vorher `628c413e922cbacf4359ac7734fe1a691a53a235` →
  nachher `b7d25fb10cdda31bc75cdd50796e7548e9d9c304`): 2 Bean-Zeilen geändert. Zeile 8
  (gbk-tsk1, todo/task): `◉ ⯅` → `t T` (Status Text→Green, Type Blue→Sky). Zeile 9
  (gbk-tsk2, draft/bug, CURSOR-Zeile `▌`): `◉ ⯁` → `d B` (Status Blue unverändert,
  Type Red unverändert; wieder Cursor-Accent-Zeile, D08 unverändert). Keine
  Spaltenbreiten-Verschiebung.
- **chrome.golden**: UNVERÄNDERT (SHA1 `db2d3398633bc47fcaa753e438e79ff37815430c`
  identisch vorher/nachher, byte-für-byte diff leer) — bestätigt die Plan-Erwartung
  (Step 7): Chrome rendert keine Bean-Zeilen, PF-6 betrifft es nicht.

**EAW-Breiten-Analyse (Pflicht, Unicode-East-Asian-Width via Python `unicodedata`,
identisch zur Datenquelle, die `clipperhouse/displaywidth` nutzt):**

| Alt-Glyph | EAW | Neu (Type M/E/F/T/B bzw. Status d/t/i/c/s) | EAW | Breiten-Drift? |
|---|---|---|---|---|
| `◉` (Status, alle) | Neutral | Buchstabe (Status) | Narrow | Nein — beide Breite 1 |
| `⬢`/`✦`/`⯅`/`⯁` (Type) | Neutral | Buchstabe (Type) | Narrow | Nein — beide Breite 1 |

Alle alten Type-/Status-Icons waren bereits EAW=Neutral (nicht Ambiguous, entgegen der
Auftrags-Vermutung „⬢/◉ sind ambiguous-width" — verifiziert via `unicodedata.
east_asian_width()`: `◉` U+25C9, `⬢` U+2B22, `✦` U+2726, `⯅` U+2BC5, `⯁` U+2BC1 sind
alle `N` (Neutral), nicht `A` (Ambiguous); die bestehenden Code-Kommentare vor T2 hatten
das bereits korrekt dokumentiert). Neutral- wie Narrow-Zeichen werden von
displaywidth/uniseg-artigen Libraries durchgängig als Breite 1 gerechnet — kein
Ambiguous-Drift-Risiko, unabhängig vom Terminal. Beleg: Box-Ränder in beiden Goldens
byte-identisch an derselben Spalte, nur der Zeicheninhalt zwischen den Rändern ändert
sich (s. Diffs oben) — bestätigt empirisch keine Layout-Verschiebung.

**Priority-Glyphen (NEU, nicht in den 3 Goldens exerciert — `theme.Priority` wird dort
nicht gerendert, nur im Detail-Meta-Bereich, s. Smoke unten):** 3 der 5 neuen
Priority-Glyphen sind laut `unicodedata.east_asian_width()` tatsächlich **Ambiguous**:
`·` (normal, U+00B7, `A`), `↓` (low, U+2193, `A`), `→` (deferred, U+2192, `A`) — nur
`‼` (critical, U+203C, `N`) und `!` (high, U+0021, `Na`) sind spaltensicher. Das ist ein
reales, vom Auftrag korrekt antizipiertes Drift-Risiko in CJK/Ambiguous=Wide-Terminals —
aber design-spec.md §15 PF-6 legt exakt diese Glyphen verbindlich fest (PO-Direktive,
Q02 bestätigt), keine Abweichung vorgenommen. Da kein bestehender Golden `Priority()`
rendert, hat dieses Risiko hier KEINEN Golden-Impact; dokumentiert als Kenntnisnahme für
künftige Priority-Goldens (z.B. falls T4/T6 Detail-Goldens mit offener Meta-Sektion
hinzufügen).

## Smoke

tmux (Scratch-Repo `/private/tmp/.../bt-2af1-smoke`, 6 Beans: je 1 pro Typ, je 1 pro
Status, volles Prioritätsspektrum — milestone/todo/normal, epic/in-progress/high,
feature/completed/low, task/scrapped/deferred, bug/draft/critical, task/todo/normal).
Tree-View mit „Archivierte einblenden" aktiviert (sonst completed/scrapped
standardmäßig ausgeblendet — Archiv-Sicht-Verhalten unverändert, kein PF-6-Bezug) zeigt
alle 6 Zeilen. Farb-Verifikation via `tmux capture-pane -e` (RGB-Hex aus den
TrueColor-ANSI-Codes extrahiert, jeweils Buchstabe außerhalb der Cursor-Zeile geprüft):

| Status | Buchstabe | Farbe (Hex) | Erwartung | Match |
|---|---|---|---|---|
| draft | d | #8aadf4 | Blue | ✓ |
| todo | t | #a6da95 | Green | ✓ (2x, verschiedene Zeilen) |
| in-progress | i | #eed49f | Yellow | ✓ |
| completed | c | #a5adcb | Subtext | ✓ |
| scrapped | s | #a5adcb | Subtext | ✓ |

| Typ | Buchstabe | Farbe (Hex) | Erwartung | Match |
|---|---|---|---|---|
| milestone | M | #8aadf4 | Blue | ✓ |
| epic | E | #c6a0f6 | Mauve | ✓ |
| feature | F | #c6a0f6 | Mauve | ✓ |
| task | T | #91d7e3 | Sky | ✓ (2x) |
| bug | B | #ed8796 | Red | ✓ |

Priorität via Detail-Panel (`tab` → `1` Meta-Sektion, je Bean angesteuert):

| Priorität | Glyph | Farbe (Hex) | Bold | Erwartung | Match |
|---|---|---|---|---|---|
| critical | ‼ | #ed8796 | ja | Red, Bold | ✓ |
| high | ! | #eed49f | ja | Yellow, Bold | ✓ |
| normal | · | #cad3f5 | nein | Text | ✓ |
| low | ↓ | #a5adcb | nein | Subtext | ✓ |
| deferred | → | #a5adcb | nein | Subtext | ✓ |

Cursor-Zeilen (D08-Accent-Darstellung, z.B. Epic beim ersten Laden) zeigen erwartungsgemäß
KEINE Einzel-Buchstaben-Farbe (ganze Zeile Mauve-Akzent) — bestätigt als bestehendes,
von PF-6 unberührtes Verhalten (Doku-Kommentar `treeRows`, view_browse_repo.go).

## Deviations

- Kein inhaltlicher Deviation von design-spec.md §15 PF-6 oder epic-E7-plan.md Task 2 —
  alle 5 Status-, 5 Typ- und 5 Prioritäts-Zuordnungen 1:1 wie in den Tabellen.
- **Collateral-Fix außerhalb der T2-Dateiliste** (nicht in „Files" des Plans genannt,
  aber compiler-/test-getrieben notwendig für ein grünes `go test ./...`):
  `internal/tui/view_detail_bean_test.go` — Priority-Assertion in
  `TestBeanSectionsMetaRendersStatusTypePriorityTags` von `"critical"` (Wort) auf `"‼"`
  (Glyph) umgestellt, da `theme.Priority()` seit PF-6 keinen Wort-Output mehr liefert.
  Kein Scope-Creep in der Produktionslogik — reiner Test-Nachzug, analog zu T1s
  compiler-getriebenen Sweeps.
- `priorityColor`/`priorityGlyph` bekamen für unbekannte Prioritäten denselben
  generischen Fallback-Mechanismus (`fallbackGlyph()`) wie Status/Type — im Plan nicht
  explizit gefordert (nur implizit über „unverändert" für den Fallback-Pfad), aber
  konsistent mit dem bestehenden Muster und zusätzlich getestet (`TestPriorityColorMapping`
  hat jetzt einen unknown-priority-Case, den die alte Version nicht hatte). Trivial, kein
  Risiko (fällt auf denselben Text/·-Fallback wie zuvor implizit über den `default`-Zweig).

## Notes for T3 (Englisch+Palette)

Beim Durcharbeiten (Theme-Grep-Sweep + tmux-Smoke) begegnete deutsche Strings — bereits
in epic-E7-plan.md Task 3s PF-7/PF-8-Tabellen erfasst, hier nur als Bestätigung/
Cross-Check ohne neue Funde:
- Filter-Overlay (`box_filter_facets.go`): Abschnittskopf „Archiv" + Checkbox
  „Archivierte einblenden" — live im Smoke gesehen, exakt wie in Task 3s Tabelle
  (`box_filter_facets.go:39`/`:79`) bereits katalogisiert.
- Accordion-Sektionstitel „Beziehungen"/„Historie" (im Smoke live sichtbar als `[3]
  Beziehungen`/`[4] Historie`) — bereits als T4-Scope vermerkt (PF-7-Anteil,
  `beanSections`-Rewrite), nicht T3.
Keine WEITEREN, bisher unkatalogisierten deutschen UI-Strings während T2 entdeckt —
T2s Scope (theme-Paket) hat keine eigenen UI-String-Literale (nur Doc-Kommentare,
nicht Teil des PF-7-Sweeps). Der komplette Sweep bleibt T3s eigene Aufgabe (Step 1,
Vollständigkeits-Grep); dies ist nur ein Nebenbefund aus der Coverage des T2-Smoke-Pfads
(Filter-Menü + Detail-Meta), nicht erschöpfend.

Refs: bt-2af1
