# Epos E8 — PO-Feedback R2: Tags-Zeile + Detail-UX-Feinschliff

Liefert: die D01-Entscheidung (Tags als 7. Meta-Feld, löst US-08s letzte
Lücke/bean `bt-gdkx`), 14 Nebenfunde B01-B14 (Detail-Pane-Politur, Maus,
Quit-Flow, Feld-Edit-Dispatch, Command-Center-Bereinigung) und sechs weitere
Grilling-Entscheidungen D02-D06/D08 (Backlog-Sort-Indikator,
Header/Footer-Neuspezifikation R2 inkl. Blocking-Remap `B`→`r`,
esc-Rückwärts-Kaskade, Tag-Page-Scope-Bestätigung) als Code — aus der
PO-Feedback-Runde 2 (2026-07-15, während D-Grilling + Live-Nutzung von
`bt`) entstanden. Details/Zitate/Herleitung: `design-spec.md` §15 PF-15/
PF-16, Epic-bean `bt-ntoz` (vollständige Auftragsquelle, D01-D08 + B01-B14
verbatim strukturiert).

Quellen: `design-spec.md` §15 PF-15 (D01, Tags-Zeile) + PF-16 (B01-B14 +
D02-D06/D08 kompakt) · Epic `bt-ntoz` (alle Funde + Entscheidungen im
Body) · `validation.md` §5 (D01-D08-Tabelle, durch dieses Epos entschieden)
· `epic-E7-plan.md` (Muster für Aufbau/Detailtiefe) · Code-Stand
2026-07-15 (Dateien einzeln je Task referenziert).

Nicht in diesem Epos: D07 (Upstream-ETag-Issue bei `hmans/beans`) — explizit
NACH v1-Abnahme, Entwurf durch Agent aber POST nur mit PO-Freigabe (bean
`bt-ntoz` PO-Freigabe-Abschnitt, T03). Q03/Tag-Management-**Page**
(`bt-6oyy`) bleibt v1.1 (D08) — B14 ist die v1-Minimal-Lösung, `bt-6oyy`
selbst NICHT anfassen. D08s beans-Nachtrag auf `bt-6oyy` ist bereits vom
PO/einer Vorsession erledigt (verifiziert 2026-07-15, kein T04 mehr nötig).

## Task-Übersicht

| Task | bean | Inhalt | Codes | blocked_by |
|---|---|---|---|---|
| T1 | `bt-e6q9` | Detail-Pane Kopfblock + Meta-Feldliste | D01, B02, B04, B09, B03(verify) | — |
| T2 | `bt-czpf` | Accordion-Header: Chevron + Teal-Experiment | B05, B06 | — |
| T3 | `bt-qbyq` | Detail-Focus: Pfeil-Symmetrie + esc-Kaskade | B01, D03 | T1 |
| T4 | `bt-duz7` | Maus im Detail-Pane | B07 | T1, T3, T6 |
| T5 | `bt-1u0t` | Quit-Flow zweistufig | B08 | — |
| T6 | `bt-y2iw` | Feld-Edit-Dispatch (BODY-Editor + Einzel-Gruppen-Value-Menüs) | B10, B11, B12 | T3 |
| T7 | `bt-yqdy` | Command-Center: Bean-Suche raus + create-tag | B13, B14 | T6 |
| T8 | `bt-d8kc` | Header/Footer R2 + Backlog-Sort-Indikator | D02, D04, D06, Q06 | — |
| T9 | `bt-6ppq` | Abschluss (Voll-Validierung, Epic to-review) | — | T2, T4, T5, T7, T8 |

**Reihenfolge-Begründung:** vier Tasks (T1/T2/T5/T8) haben keine
Abhängigkeiten — sie berühren disjunkte Dateibereiche (T1: `view_detail_
bean.go`+`view_browse_repo.go`-Signaturen; T2: `accordion.go`+`theme.go`;
T5: `box_confirm_quit.go`; T8: `keymap.go`+beide `*LocalBindings()`-
Funktionen) und können in beliebiger Reihenfolge landen. T3 folgt auf T1,
weil beide `keyDetailFocus` (`update.go`) anfassen — T1 ändert dort nur die
`beanSections()`-Aufrufzeile (neuer `detailLevel`-Parameter), T3 den
`left`-Case + einen neuen `esc`-Handler; Reihenfolge vermeidet
Merge-Überraschungen in derselben Funktion, kein technischer Zwang. T6
folgt auf T3 aus demselben Grund (B10 ergänzt `keyDetailFocus`s
`detailLevel==0`-Enter-Block, direkt neben T3s neuem `esc`-Handler). T4
(Maus) braucht T1 (7-Zeilen-Meta-Liste inkl. Tags MUSS stehen, bevor die
Klick-zu-Feldindex-Abbildung gebaut wird), T3 (derselbe `keyDetailFocus`-
Bereich) UND T6 (die Feld-Kind-Switch-Logik, die T4 in einen
`activateDetailField`-Helfer extrahiert und per Doppelklick wiederverwendet,
liegt erst nach T6s BODY-Editor-Ergänzung in ihrer finalen Form vor). T7
folgt auf T6, weil beide `overlay_palette.go`s `paletteActions()`/
`dispatchPalette()` anfassen (T6 ergänzt `set type`/`set priority`, T7
entfernt den Bean-Suche-Unterbau und ergänzt `create tag` — dieselbe
Funktion). T9 (Abschluss) `blocked_by` die fünf Blatt-Tasks des
Abhängigkeitsgraphen (T2/T4/T5/T7/T8) — deckt transitiv auch T1/T3/T6 ab,
da T4 und T7 bereits an ihnen hängen.

**Golden-Strategie (gilt für T1/T2/T8, teils T-abhängig):** T1/T2/T8
ändern sichtbaren Render-Output (Meta-Feldliste, Accordion-Header,
Header/Footer) — nach JEDEM: `command go build -o bin/bt .`, dann `command
go test ./internal/tui/ -run "TestTreeGolden|TestBacklogGolden|
TestChromeGolden" -update`. `git diff --stat internal/tui/testdata/`
danach ansehen, JEDE geänderte Datei bekommt eine Vorher/Nachher-
Beschreibung im Commit-Body (Pflicht) — auch „unverändert" ist eine
gültige, explizit zu nennende Aussage. T3/T4/T6/T5/T7 ändern reine
Tastatur-/Maus-/Overlay-Logik OHNE Render-Output-Unterschied in den 3
Basis-Goldens (Overlays sind nicht Teil von Tree/Backlog/Chrome-Goldens) —
ihr jeweiliger Golden-Schritt ist ein GEGENBELEG (`command go test
./internal/tui/ -run "..."` OHNE `-update`, MUSS grün bleiben), nicht ein
Regen.

---

## Task 1: Detail-Pane Kopfblock + Meta-Feldliste (`bt-e6q9`)

Deckt **D01** (Tags als 7. Meta-Feld), **B02** (Kopfblock-Spaltenbreiten),
**B04** (▶-Marker-Gating auf `detailLevel==1`), **B09** (▷-Marker-Farbe),
**B03** (Tree-Leaf-Marker — Verifikation, kein Fix). Vollständige
Spezifikation, Code-Sketches und TDD-Schritte: bean `bt-e6q9` (self-
contained). Kurzfassung:

- `metaFields()`/`metaFieldLabels` (`view_detail_bean.go`) wachsen von 6
  auf 7 Einträge — `tags:` NACH `priority`, VOR `created_at`. Wert via die
  bislang tote `tagsInline()` (`render_shared.go:103-112`, wiederbelebt).
  Neuer `relationField.kind == "tags"` — Enter-Kaskade öffnet
  `m.openTagPicker()`, `m.detailFocus` bleibt `true` (wie `status`/`type`/
  `priority`).
- `detailHeaderBlock()`: `type`/`status` fest auf 9/11 Zeichen gepaddet
  (Wortlänge `milestone`/`in-progress`) — keine Zeilensprünge mehr beim
  Bean-Wechsel.
- Signatur-Kette `beanSections`/`renderAccordionPane`/
  `renderBeanAccordionPane` (+ `keyDetailFocus`s eigener Navigations-
  Aufruf) wächst um einen `detailLevel int`-Parameter — `metaSectionBody`s
  ▶-Marker erscheint erst wenn `detailLevel==1` (nicht mehr automatisch
  bei `activeIdx==0`).
- `metaSectionBody`s inaktiver `▷`-Marker wird `theme.Muted`-gefärbt statt
  unstilisiert.
- `treeNodeMarker` (`view_browse_repo.go:401-409`) zeigt bei genauer
  Prüfung bereits KEIN Dreieck für kinderlose Beans — B03 reproduziert
  NICHT gegen den aktuellen Code (ERRATUM: bereits korrekt). Kein
  Code-Fix, nur ein Regressionstest (`TestTreeNodeMarkerBlankForLeaf`).
- Neu: benannte Section-Index-Konstanten (`metaSectionIdx`=0,
  `bodySectionIdx`=1, `relationsSectionIdx`=2, `historySectionIdx`=3,
  neben `beanSectionCount`) — bereitet T6/B10 vor (vermeidet dort eine
  Magic-Number `1`).

Schließt bean `bt-gdkx` (US-08-Redefinition) inhaltlich — referenziert,
NICHT geschlossen (PO-Gate, Review-Flow §5).

**Akzeptanz-Checkliste:** 7-zeilige Meta-Liste inkl. Tags · Enter auf
`tags:` öffnet Tag-Picker · Kopfblock springt nicht mehr · ▶ erscheint erst
nach explizitem Feld-Einstieg · inaktive Marker subtext-grau ·
`TestTreeNodeMarkerBlankForLeaf` grün · 3 Goldens regeneriert · voller Lauf
grün.

---

## Task 2: Accordion-Header — Chevron entfernen + Teal-Experiment (`bt-czpf`)

Deckt **B05** (redundantes `▾`/`▸`-Chevron entfernen) und **B06**
(EXPERIMENT: inaktive Header-Farbe Grau→Teal) aus bean `bt-ntoz`.
Vollständige Spezifikation: bean `bt-czpf`. Kurzfassung: `renderAccordion`
(`accordion.go`) verliert das `hint`-Suffix (Zustand ist am Inhalt selbst
sichtbar); neuer Theme-Token `theme.HeaderInactive` (Teal, `theme.go` —
Repo-Regel „keine Hex-Literale in Views" verbietet einen Inline-Style in
`accordion.go`) ersetzt `theme.Muted` NUR für den inaktiven
Sektions-Header-Titel (die Meta-Label-Spalte selbst bleibt `theme.Muted`
unverändert). B06 ist EXPLIZIT als Experiment markiert — PFLICHT vor
Abnahme: ein Vorher/Nachher-Beleg (Screenshot oder Golden-Diff-Ausschnitt)
im Commit-Body, UND ein Verweis darauf im Epic-Bean `bt-ntoz` (T9/Step 4
verlinkt ihn dorthin) für den PO-Sign-off. Unabhängig von allen anderen
E8-Tasks (kein `blocked_by`) — reiner `accordion.go`+`theme.go`-Scope.

**Akzeptanz-Checkliste:** kein `▾`/`▸`-Suffix mehr · inaktive Header nutzen
`theme.HeaderInactive` · Meta-Label-Spalte unverändert `theme.Muted` ·
Vorher/Nachher-Beleg dokumentiert · 3 Goldens regeneriert · voller Lauf
grün.

---

## Task 3: Detail-Focus — Pfeil-Symmetrie + esc-Rückwärts-Kaskade (`bt-qbyq`)

Deckt **B01** (Pfeil-links verlässt Detail-Fokus nicht mehr) und **D03**
(esc als universelles „eine Ebene zurück", inkl. Detail-Kaskade + Audit
aller esc-Sites) aus bean `bt-ntoz`. `blocked_by` `bt-e6q9` (T1). Voll-
ständige Spezifikation: bean `bt-qbyq`. Kurzfassung:

`keyDetailFocus`s `left`-Case (`update.go`) verliert seinen `else`-Zweig
(`m.detailFocus = false` bei `detailLevel==0`) — Pfeiltasten sind danach
REIN Navigation, nie Fokus-Wechsel (revidiert einen Satz aus design-
spec.md §15 PF-13, s. dortige „PF-13-Pfeil-Revision"). Im GEGENZUG bekommt
`keyDetailFocus` einen NEUEN `esc`-Handler (die Zwei-Stufen-Kaskade, die B01
aus den Pfeiltasten entfernt, wandert strukturell dorthin): erstes `esc`
Feld→Sektion, zweites `esc` verlässt den Detail-Fokus. Ergebnis: EIN
mentales Modell — `esc` geht immer genau eine Ebene zurück. Der
Audit-Nachweis (Suche/Filter/Picker/Lobby/Quit — 5 von 6 Bereichen bereits
konform, nur die Detail-Kaskade hatte eine Lücke) ist Pflichtbestandteil
des Commit-Bodys (File:Line-Belege je Standort, s. bean `bt-qbyq`).

**Akzeptanz-Checkliste:** Pfeil-links bei `detailLevel==0` ist No-Op ·
`esc` kaskadiert Feld→Sektion→Fokus-Exit · Audit-Tabelle im Commit-Body ·
kein Golden ändert sich · voller Lauf grün.

---

## Task 4: Maus im Detail-Pane (`bt-duz7`)

Deckt **B07** aus bean `bt-ntoz`. `blocked_by` `bt-e6q9` (T1, finale
7-Zeilen-Meta-Form), `bt-qbyq` (T3, derselbe `keyDetailFocus`-Bereich),
`bt-y2iw` (T6, die zu extrahierende Feld-Kind-Switch-Logik liegt erst nach
T6 in finaler Form vor). Vollständige Spezifikation: bean `bt-duz7`.
Kurzfassung: Klicks im Detail-Pane wurden bislang stillschweigend
verworfen (`treeClickRow`/`backlogClickRow` begrenzen sich auf die
linke Tree-/Backlog-Spalte). Neu:

1. `keyDetailFocus`s bestehender Feld-Kind-`switch` (status/type/priority/
   tags/title/readonly/jump) wird in einen geteilten Helfer
   `activateDetailField(b, f)` extrahiert — EINZIGE Stelle, die
   Feld-Kind auf Overlay/Form/Jump dispatcht.
2. Neue Geometrie-Funktion `detailClickRow` (`mouse.go`, analog
   `treeClickRow`) bildet Klick-Koordinaten auf Sektions- oder
   Meta-Feld-Index ab — MUSS den 5-zeiligen Kopfblock-Offset
   (`detailHeaderBlock`) berücksichtigen. Relations-Feld-Klicks werden für
   v1 auf Sektions-Ebene vereinfacht (kein Feld-genaues Adressieren über
   Gruppen-Header hinweg) — dokumentierte Scope-Entscheidung, kein
   PO-Wortlaut verlangt mehr.
3. Neuer Dispatch `mouseDetailClick`: Einzelklick selektiert
   Sektion/Feld, Doppelklick (bzw. Zweitklick auf bereits selektiertes
   Feld, mirrort `mouseTreeClick`s `lastClickIdx`/`lastClickAt`-Muster)
   ruft `activateDetailField`. Toast-Dismiss-Vorrang + Overlay-Guard
   (`handleMouse`) wirken unverändert — nur verifiziert, nicht neu gebaut.

**Akzeptanz-Checkliste:** Sektions-Klick aktiviert/expandiert · Feld-Klick
selektiert ohne Overlay zu öffnen · Doppelklick öffnet das Overlay ·
Kopfblock-Offset korrekt · `activateDetailField` ist Single-Source für
Tastatur UND Maus · kein Golden ändert sich · voller Lauf grün.

---

## Task 5: Quit-Flow zweistufig (`bt-1u0t`)

Deckt **B08** aus bean `bt-ntoz`. Unabhängig (kein `blocked_by`).
Vollständige Spezifikation: bean `bt-1u0t`. Kurzfassung: `quitBox()`-Text
„Really quit bt." → „Really quit bt?" (Frage). `keyConfirmQuit`s
`enter`-Zweig wird zweistufig:

```go
case keybind.Matches(msg, keys.Enter):
    m.confirmQuit = false
    if m.view != viewLobby && len(m.settings.Repos) > 0 {
        return m.openLobby()
    }
    return m, tea.Quit
```

Deckt alle drei PO-Fälle: Browse/Backlog + Repos konfiguriert → Lobby
(Stufe 1) · bereits in der Lobby → Exit (Stufe 2) · RANDFALL (Planner-
Entscheidung, konservativ): Browse/Backlog OHNE konfigurierte Repos → Exit
direkt wie bisher (eine leere Lobby wäre sinnlos). `keyLobby`s eigener
`esc`/`q`-Case (`view_lobby.go`) bleibt UNVERÄNDERT — separater,
bestehender Pfad, von B08 nicht betroffen. Zusatzverbesserung (Planner,
über den Wortlaut hinaus): `quitBox()`s Hint-Zeile wird kontextsensitiv
(„enter: go to lobby" vs. „enter: quit"), damit die Modal-Hint-Zeile selbst
nicht zur Überraschung wird.

**Akzeptanz-Checkliste:** Text ist Frage · zweistufige Kaskade inkl.
Randfall · `keyLobby` unverändert · 3 Goldens unverändert (verifiziert) ·
voller Lauf grün.

---

## Task 6: Feld-Edit-Dispatch — BODY-Editor + Einzel-Gruppen-Value-Menüs (`bt-y2iw`)

Deckt **B10** (BODY-Sektion `e`/`enter` kontextsensitiv), **B11**+**B12**
(Value-Menü auf EINE Gruppe splitten + Palette-Ergänzung) aus bean
`bt-ntoz`. `blocked_by` `bt-qbyq` (T3, derselbe `keyDetailFocus`-Bereich).
Vollständige Spezifikation: bean `bt-y2iw`. Kurzfassung:

Neuer geteilter Helfer `openBodyEditor(b)` (ersetzt die bisherige
Ad-hoc-Logik in `keyNodeAction`s `ctrl+e`-Zweig) — `keyNodeAction`s
`e`-Zweig ruft ihn zusätzlich, wenn `m.detailFocus && m.secCursor ==
bodySectionIdx`; `keyDetailFocus`s `detailLevel==0`-Enter-Block bekommt
denselben Fall (`enter` auf BODY war zuvor No-Op). Alle anderen Sektionen
(META/RELATIONS/HISTORY) bleiben beim bisherigen Titel-Edit-Fallback.

`buildValueMenuItems(group string)` (`box_menu_value.go`) liefert künftig
NUR die angefragte Gruppe (5 statt 15 Zeilen) — `openValueMenu(group)`
reicht `group` durch statt es nur zum Cursor-Seeding zu nutzen. `s`
bleibt Status-only (design-spec §7 reserviert bewusst EINE Taste fürs
Cluster, B12 revidiert das NICHT — Klarstellung im Commit-Body Pflicht,
falls ein Implementer versucht ist, eine neue Taste zu ergänzen). Type/
Priority bleiben über die PF-5-Meta-Kaskade erreichbar (bereits korrekt
group-scoped seit E7/T6, jetzt zusätzlich korrekt GEFILTERT). Palette
gewinnt `set type`/`set priority` (bislang nur `set status`).

**Akzeptanz-Checkliste:** `e`/`enter` auf BODY öffnen `$EDITOR` · andere
Sektionen unverändert · Value-Menü zeigt je Gruppe nur 5 Zeilen · keine
neue Type-/Priority-Taste · Palette bietet `set type`/`set priority` ·
kein Golden ändert sich · voller Lauf grün.

---

## Task 7: Command-Center — Bean-Suche raus + create-tag-Command (`bt-yqdy`)

Deckt **B13** (Bean-Suche aus dem Command-Center entfernen) und **B14**
(Tag-Neuanlage entdeckbar machen) aus bean `bt-ntoz`. `blocked_by`
`bt-y2iw` (T6, dieselbe `paletteActions()`/`dispatchPalette()`-Funktion).
Vollständige Spezifikation: bean `bt-yqdy`. Kurzfassung:

**B13** — Compiler-gesteuerte Entfernung (Muster: `epic-E7-plan.md` Task 1/
PF-14): `paletteKindBean`, `palFilteredBeans`/`palBeanMatches`/
`paletteBeanResultCap`, `dispatchPalette`s Bean-Case, `paletteBox`s
Bean-Rendering, der komplette Palette-eigene Bleve-Unterbau
(`palBleveIDs`/`palBleveFor`/`palBleveLoading`, `paletteBleveResultMsg`,
`paletteSearchCmd`, `applyPaletteBleveResult`, `maybePaletteBleveCmd`)
entfallen vollständig — `/`s eigene, komplett separate Bleve-Suche bleibt
unangetastet. **Revidiert US-04** (design-spec §10, bereits per PF-16
dokumentiert) — Command-Center zeigt danach AUSSCHLIESSLICH Commands.

**B14** — der „Neuer Tag"-Modus in `box_picker_tag.go` ist NICHT kaputt
(verifiziert: `n` ist bereits verdrahtet, die Inline-Hint-Zeile zeigt es
bereits) — nur die ÄUSSERE Footer-Zeile (Zone 3) zeigte es nicht. Neue
`keys.NewTag`-Bindung (`n`) macht es footer-fähig
(`tagPickerLocalBindings()` ergänzt). Neuer Palette-Command `create tag`
öffnet den Tag-Picker UND direkt dessen Neuanlage-Input
(`m.openTagPicker().openTagInput()`) — mit PFLICHT-Guard
`focusedBean()==nil` (sonst latenter `tagInputActive`-State ohne
offenen Picker). Tag-Management-**Page** (`bt-6oyy`) bleibt v1.1 (D08).

**Akzeptanz-Checkliste:** keine Bean-Treffer mehr im Command-Center · Bleve-
Unterbau vollständig entfernt, `/`s Suche unverändert funktional · Footer
zeigt `n:New tag` · Palette bietet `create tag` mit Guard · 3 Goldens
unverändert (verifiziert) · voller Lauf grün.

---

## Task 8: Header/Footer-Neuspezifikation R2 + Backlog-Sort-Indikator (`bt-d8kc`)

Deckt **D02** (Backlog-Sort-Suffix), **D04** (Header auf 4 Globals
gekürzt), **D06**+**Q06** (Footer-Neuspezifikation inkl. Blocking-Remap
`B`→`r`) aus bean `bt-ntoz`. Unabhängig (kein `blocked_by`) — reiner
`keymap.go`+`view_browse_repo.go`+`view_browse_backlog.go`-Scope, kein
Überlapp mit `view_detail_bean.go`/`keyDetailFocus`. Vollständige
Spezifikation: bean `bt-d8kc`. Kurzfassung:

`globalBindings()` (`keymap.go`) schrumpft von 7 auf exakt 4:
`{Palette, Picker, Help, Quit}` — `ctrl+r`/`esc`/`enter` fliegen aus dem
Header, bleiben im Help-Overlay dokumentiert (`helpGroups()` unverändert).
`browseRepoLocalBindings()`/`backlogLocalBindings()` werden komplett neu
zusammengestellt (Navigations-Keys komplett raus, finale Reihenfolge PO-
verbatim: `tab focus in · shift+tab focus out · / search · f Filter ·
s Status · c Create · d Delete · e Edit · b Backlog · t Tags · y Yank ·
a Parent · r Blocking`) — Backlog behält zusätzlich seinen bestehenden
`Sort`-Eintrag (Planner-Entscheidung, ERRATUM ggü. der wörtlichen Q06-
Liste: Sort ist Backlog-exklusiv und von der gemeinsamen Liste nicht
berührt, kein Entzug ohne PO-Anweisung). `keys.Blocking` wechselt von `B`
auf `r` (verifiziert kollisionsfrei — `r` seit PF-14 frei, `B` wird durch
den Remap frei). `renderBindings()` (`view.go`) stellt Taste/Aktionswort
künftig farblich getrennt dar (kein `:` mehr) — gilt einheitlich für Header
UND Footer (eine gemeinsame Funktion). `treeSearchLine` (`view_browse_
repo.go`) bekommt einen `sortSuffix string`-Parameter — Tree übergibt `""`
(unverändert), Backlog übergibt `"sort " + backlogSortDisplayLabel(...)`
(z. B. `⌕ / search · sort prio`, PO-Beispiel verbatim).

**Akzeptanz-Checkliste:** Header exakt 4 Bindings · Footer beider Views
zeigt die Q06-Liste in vorgegebener Reihenfolge · Backlog behält Sort
zusätzlich (dokumentiert) · `Blocking` an `r`, `B` frei · Taste/Wort
farblich getrennt (Header+Footer) · Backlog-Suchzeile zeigt Sort-Suffix,
Tree unverändert · Drift-Guard grün · Goldens regeneriert · voller Lauf
grün.

---

## Task 9: Abschluss (`bt-6ppq`)

**blocked_by:** T2, T4, T5, T7, T8 (deckt transitiv T1/T3/T6 ab). Keine
Code-Änderungen erwartet — reine Validierung + Doku + beans-Pflege
(Muster: `epic-E7-plan.md` Task 8). Vollständige Spezifikation: bean
`bt-6ppq`. Kurzfassung:

1. Voller Regressionslauf (Build, `-race`, `-short` 2×, VOLL 2×, alle
   Golden-Funktionen mit `-count=2`, gofmt/vet leer) — Beleg im bean-Body
   unter „Voll-Gate-Beleg".
2. `bt-gdkx` (US-08) bekommt Tag `to-review` + Body-Verweis auf den T1-Fix
   — bleibt selbst NICHT `completed` (PO-Gate).
3. `validation.md`-Konsistenz gegen den tatsächlichen Code-Stand
   verifizieren (Planner hat §5 bereits aktualisiert — hier nur
   Gegenprüfung nach T1-T8).
4. B06-Experiment-Sign-off-Verweis im Epic-Bean `bt-ntoz` verlinken (PFLICHT
   vor `to-review` — PO muss den T2-Beleg finden, ohne danach suchen zu
   müssen).
5. Epic-Ritual: `bt-ntoz` bekommt Tag `to-review` (Agent setzt NIE
   `completed`); T1-T8 auf `completed` verifizieren.
6. `docs/SSTD.md` — Pointer-Update nur falls nötig (prüfen + dokumentieren).
7. Commit `docs(release): E8-Abschluss — Epic to-review, US-08/bt-gdkx-
   Status, B06-Sign-off-Verweis`.

**Akzeptanz-Checkliste:** voller Lauf grün · `bt-gdkx` trägt `to-review` +
Verweis · `bt-ntoz` trägt `to-review`, nicht `completed` · T1-T8 alle
`completed` · B06-Sign-off auffindbar · `validation.md`-Konsistenz
verifiziert.

---

## Selbst-Review (Plan gegen alle D/B-Punkte aus `bt-ntoz`)

- **Jeder D/B-Punkt genau einem Task zugeordnet:** D01→T1 · B02→T1 · B03→T1
  (Verifikation) · B04→T1 · B09→T1 · B05→T2 · B06→T2 · B01→T3 · D03→T3 ·
  B07→T4 · B08→T5 · B10→T6 · B11→T6 · B12→T6 · B13→T7 · B14→T7 · D02→T8 ·
  D04→T8 · D06→T8 · Q06→T8 · D05→ Sign-off/No-Op, verifiziert in T8s
  Kontext (Redundanz durch D04 bereits aufgelöst, kein Code-Fix nötig,
  aber im PF-16-Text dokumentiert) · D08→ bereits erledigt (bt-6oyy-
  Body-Nachtrag existiert schon, verifiziert bei der Recherche zu diesem
  Plan, kein T04-Task nötig) · D07→ EXPLIZIT ausgeschlossen (nach
  v1-Abnahme, POST nur mit PO-Freigabe). Kein Punkt doppelt vergeben,
  keiner vergessen.
- **PF-13-Pfeil-Revision (B01) dokumentiert:** design-spec.md §15 trägt
  einen eigenen Abschnitt „PF-13-Pfeil-Revision" — T3s Task-bean zitiert
  ihn.
- **US-04-Revision (B13) dokumentiert:** design-spec.md §15 trägt einen
  eigenen Abschnitt „US-04-Revision" — T7s Task-bean zitiert ihn, §10 selbst
  bleibt unverändert (Revision lebt in §15, wie bei PF-14/US-08 vorgemacht).
- **Nicht-Ziele ausgeschlossen:** T03 (Upstream-Issue, D07) nirgends
  referenziert außer als expliziter Ausschluss oben. `bt-6oyy` (Q03/Tag-
  Page) nirgends angefasst, nur als Kontext für B14 zitiert.
- **Reihenfolge als `blocked_by` kodiert, nicht als Prosa:** alle fünf
  echten Abhängigkeiten (T3←T1, T6←T3, T4←T1+T3+T6, T7←T6, T9←T2+T4+T5+
  T7+T8) sind beans-`blocked_by`-Kanten, verifiziert via `beans list
  --parent bt-ntoz --json`.
- **Jedes bean self-contained:** jede Task-bean trägt File-Referenzen,
  exakte Code-Sketches für mechanische Änderungen, TDD-Schritt-Listen mit
  konkreten Testnamen, Golden-Erwartung, Commit-Vorschlag und eine eigene
  Akzeptanz-Checkliste — ein frischer Agent kann ohne Rückfrage an das
  Epic-bean `bt-ntoz` starten (dessen vollständigen Body als Detailquelle
  zitiert, nicht dupliziert).
- **Golden-Update-Pflicht erfüllt:** T1/T2/T8 haben je einen expliziten
  Golden-Regen-Schritt mit Vorher/Nachher-Pflicht; T3/T4/T5/T6/T7 haben je
  einen expliziten Gegenbeleg-Schritt (Goldens bleiben unverändert,
  verifiziert statt angenommen).
- **Keymap-Drift-Guard-Pflicht erfüllt:** T7s neues `keys.NewTag`-Feld UND
  T8s Blocking-Remap sind in `helpGroups()` verdrahtet — bestehender Guard
  (`TestHelpGroupsCoverEveryBindingExactlyOnce`) deckt sie automatisch ab.
  T8 aktualisiert `TestGlobalBindingsExactSet`/
  `TestNoDuplicateBindingBetweenGlobalAndAnyLocalHintList` auf die neue
  4-Item-Global-Menge + die neuen Local-Listen.
