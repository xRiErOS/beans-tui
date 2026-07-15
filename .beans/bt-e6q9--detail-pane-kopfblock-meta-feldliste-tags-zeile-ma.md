---
# bt-e6q9
title: 'Detail-Pane Kopfblock + Meta-Feldliste: Tags-Zeile + Marker-/Spalten-Fixes'
status: completed
type: task
priority: high
created_at: 2026-07-15T21:05:09Z
updated_at: 2026-07-15T21:37:57Z
parent: bt-ntoz
---

E8 Task 1 — deckt D01 (Tags-Zeile), B02 (Kopfblock-Spaltenbreiten), B04 (▶-Marker-Gating), B09 (▷-Marker-Farbe) aus bean bt-ntoz. Quelle der Wahrheit: design-spec.md §15 PF-15 (D01) + PF-16 (B02/B04/B09). Ist-Code: internal/tui/view_detail_bean.go, internal/tui/accordion.go, internal/tui/view_browse_repo.go (renderAccordionPane).

## D01 — Tags als 7. Meta-Feld

metaFieldLabels/metaFields (view_detail_bean.go) wachsen von 6 auf 7 Einträge: title/status/type/priority/**tags**/created_at/updated_at (tags NACH priority, VOR created_at). Neuer relationField.kind == "tags":
- metaFields(): `{kind: "tags", label: <Wert>}` — Wert via die BEREITS VORHANDENE, aktuell aufruferlose Helferfunktion `tagsInline(b.Tags)` (render_shared.go:103-112, nutzt tagSwatch/tagChipPalette — Hash-basierte Farbe). Leeres Tags-Slice -> `theme.Dim.Render("(none)")` statt Leerstring.
- metaFieldLabels: `"tags:"` einfügen, Padding-Breite 12 bleibt (bestehende Konvention, "created_at: " ist mit 12 Zeichen das längste Label).
- Enter-Kaskade (keyDetailFocus, update.go, PF-5-Case-Switch): NEUER case `"tags": return m.openTagPicker(), nil` — analog zu status/type/priority, m.detailFocus bleibt true (gleiches Verhalten wie die drei bestehenden constrained-Felder, NICHT wie der jump-Case).
- relationField-Doc-Kommentar (accordion.go) um "tags" als gültigen kind-Wert ergänzen.

Schliesst bean bt-gdkx (US-08-Redefinition, design-spec §10) INHALTLICH — referenziere bt-gdkx im Commit, schliesse es aber NICHT selbst (Review-Flow §5, PO-Gate).

## B02 — Kopfblock feste Spaltenbreiten

detailHeaderBlock (view_detail_bean.go): type/status als Wort gerendert, aktuell OHNE Padding -> Zeile springt je nach Wortlänge (epic vs. milestone, todo vs. in-progress). Fix: fmt.Sprintf("%-9s", b.Type) (9 = len("milestone")) und fmt.Sprintf("%-11s", b.Status) (11 = len("in-progress")) VOR dem jeweiligen TypeStyle/StatusStyle-Render. prio bleibt unverändert (Glyph, keine variable Wortlänge).

## B04 — ▶-Marker erst nach explizitem Feld-Einstieg

Heute markiert metaSectionBody automatisch title: mit ▶, sobald META die aktive Sektion ist (activeIdx==0) -- UNABHÄNGIG davon ob die Feld-Ebene (m.detailLevel==1) überhaupt betreten wurde. Fix: Signatur-Kette um detailLevel int erweitern:
- beanSections(idx, b, bodyW, focused bool, activeIdx, fieldIdx, detailLevel int) []accordionSection -- metaSectionBody-Call wird zu metaSectionBody(b, bodyW, focused && activeIdx==0 && detailLevel==1, fieldIdx).
- renderAccordionPane(idx, b, w, h, open, secCursor, fieldCursor, detailLevel int, focused bool) string (view_browse_repo.go) -- neuer Parameter, reicht an beanSections durch.
- renderBeanAccordionPane (view_browse_repo.go) ruft renderAccordionPane mit zusaetzlichem m.detailLevel.
- keyDetailFocus's eigener (nicht-render-)Aufruf beanSections(m.idx, b, 40, m.detailFocus, m.secCursor, m.fieldCursor) (update.go, reine Navigations-Nutzung, nur .fields wird gelesen) MUSS ebenfalls um m.detailLevel erweitert werden (Compiler erzwingt es).
- WICHTIG: renderAccordion() selbst (accordion.go) braucht KEINE Signaturaenderung -- activeSec (Header-Highlighting) ist unabhaengig von detailLevel, nur metaSectionBody() braucht den neuen Gate.
- Kleine Zusatzarbeit (fuer T6/B10 vorbereitet, gleiche Datei/Bereich): benannte Section-Index-Konstanten metaSectionIdx=0, bodySectionIdx=1, relationsSectionIdx=2, historySectionIdx=3 in view_detail_bean.go neben beanSectionCount ergaenzen -- T6 (Task, blocked_by dieser Kette) nutzt bodySectionIdx statt Magic-Number 1.

## B09 — Inaktive ▷-Marker Farbe

metaSectionBody: marker := "▷ " ist heute UNSTILISIERT (rendert in Terminal-Default-Vordergrund, "weiss"). Fix: marker := theme.Muted.Render("▷ ") (nur der AKTIVE ▶-Marker bleibt Accent/Mauve). Breite bleibt identisch (ANSI-Codes zaehlen nicht in lipgloss.Width) -- die bestehenden PF-12-Gutter-Tests bleiben unberuehrt/gruen, aber ein NEUER expliziter Farbtest fuer B09 selbst ist Pflicht.

## B03 — Tree-Leaf-Marker (Verifikation, KEIN Fix)

treeNodeMarker (view_browse_repo.go:401-409) prueft bereits "if !n.hasKids { return '  ' }" -- kinderlose Beans zeigen BEREITS keinen Expand-Marker. Das von PO gemeldete Symptom reproduziert NICHT gegen den aktuellen Code-Stand (ERRATUM: Bug bereits gefixt oder nie im aktuellen Code vorhanden). KEINE Code-Aenderung -- stattdessen EINEN Regressionstest ergaenzen (z.B. TestTreeNodeMarkerBlankForLeaf in view_browse_repo_test.go), der das bereits-korrekte Verhalten festschreibt. Im Commit-Body EXPLIZIT als "B03: bereits korrekt, verifiziert + regressionsgesichert" vermerken.

## TDD-Schritte (Reihenfolge)

1. Failing tests zuerst: view_detail_bean_test.go (TestMetaFieldsSixEntriesWithKinds -> auf 7 Eintraege inkl. tags-kind erweitern; TestMetaSectionBodyShowsSelectedFieldMarker -> Assertion um detailLevel-Parameter erweitern + NEUER Test dass ▶ NICHT erscheint wenn detailLevel==0 trotz activeIdx==0; TestDetailHeaderBlockShowsIDTitleTypeStatusPrio -> Spaltenbreiten-Assertion), view_browse_repo_test.go (NEU TestTreeNodeMarkerBlankForLeaf), update_test.go (NEU: Enter auf tags-Feld -> overlayTagPicker, m.detailFocus bleibt true).
2. command go test ./internal/tui/... -> FAIL.
3. Implementieren (Reihenfolge: view_detail_bean.go zuerst, dann view_browse_repo.go Signaturen, dann update.go Aufruf + neuer tags-case).
4. command go test ./internal/tui/... -> PASS.
5. Golden-Regen: command go build -o bin/bt ., dann command go test ./internal/tui/ -run "TestTreeGolden|TestBacklogGolden|TestChromeGolden" -update. Erwartung: tree.golden UND backlog.golden aendern sich; chrome.golden voraussichtlich unveraendert -- trotzdem Regressionslauf, im Commit-Body vermerken.
6. command go test ./... -short gruen (2x), command go test ./... -race gruen, command gofmt -l . leer, command go vet ./... leer.
7. Commit feat(tui): PF-15/PF-16 Tags-Meta-Zeile + Kopfblock/Marker-Fixes -- Body referenziert bt-gdkx (schliesst NICHT), zitiert design-spec.md §15 PF-15/PF-16, vermerkt B03 als "bereits korrekt, verifiziert".

## Akzeptanz-Checkliste

- [x] Meta-Feldliste zeigt 7 Zeilen inkl. tags: (Hash-Swatches oder "(none)")
- [x] enter auf tags-Feld oeffnet Tag-Picker, m.detailFocus bleibt true
- [x] Kopfblock type:/status: springen NICHT mehr bei Bean-Wechsel (feste Breiten 9/11)
- [x] ▶ auf title: erscheint ERST nach explizitem Feld-Einstieg (enter auf Sektion), NICHT sofort nach tab
- [x] Inaktive ▷-Marker sind theme.Muted (subtext), nicht mehr unstilisiert
- [x] TestTreeNodeMarkerBlankForLeaf gruen (B03-Regressionslock, kein Verhaltensfix)
- [x] Alle 3 Goldens regeneriert + Vorher/Nachher je Datei im Commit-Body
- [x] Voller Testlauf (inkl. -race) gruen, gofmt/vet leer


## Summary

D01 (Tags als 7. Meta-Feld), B02 (Kopfblock-Spaltenbreiten), B04
(▶-Marker-Gating auf detailLevel==1), B09 (▷-Marker-Farbe) implementiert;
B03 verifiziert bereits korrekt + regressionsgesichert (kein Fix).
metaFields wachst von 6 auf 7 Eintraege (tags NACH priority, VOR
created_at), Enter-Kaskade bekommt neuen "tags"-Case ->
m.openTagPicker(). beanSections/renderAccordionPane/
renderBeanAccordionPane wachsen um detailLevel int. Section-Index-
Konstanten (metaSectionIdx/bodySectionIdx/relationsSectionIdx/
historySectionIdx) ergaenzt fuer T3 (bt-qbyq).

## Test-Output

RED (Compile-Fail, beanSections-Signatur):
```
internal/tui/view_detail_bean_test.go:40:64: too many arguments in call to beanSections
	have (*data.Index, *data.Bean, number, bool, number, number, number)
	want (*data.Index, *data.Bean, int, bool, int, int)
```
GREEN nach Implementierung: `go build -o bin/bt .` clean, `go vet ./...` leer.

Voll-Lauf (ohne -short): `ok beans-tui/internal/tui 135.759s` (336 Test-
Funktionen im tui-Paket, alle gruen; cmd/config/data/theme ebenfalls ok).
Zwei FRISCHE Laeufe (`-count=1`, kein Cache) hintereinander: beide
`EXIT=0`, testdata/*.golden byte-identisch zwischen beiden Laeufen
(diff -q bestaetigt IDENTICAL fuer tree/backlog/chrome.golden).
`-race -short`: ok (7.3s tui). `-race` VOLL (ohne -short): `ok
beans-tui/internal/tui 139.274s`, keine DATA RACE. `gofmt -l .` leer.

Eigene TDD-Iteration (Test-Bug, kein Prod-Bug): TestTreeNodeMarkerBlankForLeaf
scheiterte initial an `len()` (Byte-Laenge) statt `lipgloss.Width()`
(Zell-Breite) fuer den Multi-Byte-UTF-8-Marker "▾ " -- Testkorrektur,
kein ERRATUM gegen bean/Plan (siehe unten).

## Smoke

Real in tmux gegen dieses Repo (`.beans/` echte Daten, `./bin/bt`),
`tmux capture-pane -p`/`-e` als Beleg:
- D01 nicht-leer: bt-apmy (Milestone) zeigt `▷ tags:       ● to-review`
  (Hash-Swatch aus echten Repo-Daten).
- D01 leer: bt-6oyy (Feature) zeigt `▷ tags:       (none)`.
- D01 Enter-Kaskade: enter auf tags-Zeile oeffnet Tag-Picker-Overlay
  ("Tags" / "▸ [x] to-review (8)"), esc x2 schliesst sauber zurueck.
- B02 live an 2 echten Beans mit unterschiedlicher Wortlaenge:
  "type: milestone    status: todo           prio: !" (bt-apmy) vs.
  "type: feature      status: todo           prio: ·" (bt-6oyy) --
  "status:" beginnt in BEIDEN Zeilen an identischer Spalte (Index 20).
- B04 live: nach `tab` zeigt [1] META die Accent-Bar (Sektion aktiv),
  title: bleibt `▷` (KEIN ▶) bis `enter` gedrueckt wird -- danach `▶`.
- B09 live (ANSI-erhaltender Capture, `-e`): `▷`-Marker jetzt in
  `\x1b[38;2;124;124;124m` (theme.Muted-RGB, identisch zur Label-Farbe)
  gewrappt -- vorher unstilisiert.
Nicht smoke-getestet (nur Unit/Golden-Ebene): B03 (kein UI-Pfad zu
demonstrieren, da kein Verhaltensaenderung -- Regressionstest ist der
Beleg).

## Deviations/ERRATA

Keine ERRATA gegen bean/Plan -- Code-Richtungen im bean stimmten exakt
mit dem tatsaechlichen Ist-Code ueberein (Signaturen, Zeilennummern,
Call-Sites). Einzige Abweichung: Commit-Granularitaet -- D01/B02/B04/B09
in EINEM Commit statt vier, da B04s detailLevel-Parameter durch
dieselbe beanSections/metaSectionBody-Signaturkette laeuft, die D01/B09
bereits aendern (echte Code-Kopplung, kein Bequemlichkeits-Split).
Commit-Body zaehlt die vier Codes einzeln auf. B03 blieb wie geplant
ein eigener, unabhaengiger Commit (kein Code-Fix, nur Testdatei).

## Notes for T3 (bt-qbyq)

- beanSections neue Signatur: `beanSections(idx *data.Index, b *data.Bean, bodyW int, focused bool, activeIdx, fieldIdx, detailLevel int) []accordionSection`.
- renderAccordionPane neue Signatur: `renderAccordionPane(idx *data.Index, b *data.Bean, w, h, open, secCursor, fieldCursor, detailLevel int, focused bool) string`.
- renderBeanAccordionPane UNVERAENDERT (b, w, h, focused) -- reicht m.detailLevel intern durch.
- Section-Index-Konstanten jetzt in view_detail_bean.go verfuegbar:
  metaSectionIdx=0, bodySectionIdx=1, relationsSectionIdx=2,
  historySectionIdx=3 (statt Magic-Number 1 fuer BODY in B10).
- metaFields()-Reihenfolge jetzt: title(0)/status(1)/type(2)/priority(3)/
  tags(4)/created_at(5)/updated_at(6) -- jeder Code der Feld-Indizes
  hart verdrahtet hat (z.B. Enter-Kaskade-Tests), MUSS diese Verschiebung
  beachten (created_at/updated_at sind NICHT mehr 4/5, sondern 5/6).
