---
# bt-kyj5
title: E7 T4 — Meta-Layout + Kopfblock + Gutter-Stabilitaet (PF-1, PF-3, PF-4, PF-12)
status: completed
type: task
priority: high
created_at: 2026-07-15T14:26:51Z
updated_at: 2026-07-15T16:43:47Z
parent: bt-heg9
blocked_by:
    - bt-2af1
    - bt-w9o8
---

Details/Steps/Akzeptanz: docs/plans/v1-port/epic-E7-plan.md Task 4. Detail-Pane-Kopfblock (ID/Titel/type-status-prio) + Meta-Sektion wird 6-zeilige editierbare Feldliste, nicht-kollabierbar, Gutter-Stabilitaet (PF-12).


## Hinweis aus T2-Review (I01, für diesen Task relevant)

Die 3 Priority-Glyphen ·/↓/→ sind Unicode-Ambiguous-width (EAW-Klasse A; ‼/! sind safe). T4 schafft die ERSTEN Render-Stellen für theme.Priority() (Kopfblock prio:-Zeile + Meta-Feldliste) und regeneriert Detail-Goldens. Auf Terminals mit ambiguous=wide droht dort Layout-Shift. Umgang: PO-Schema ist verbindlich (keine Glyph-Abweichung) — aber beim Layout die Prio-Spalte so bauen, dass 1-vs-2-Zellen-Rendering keinen Umbruch erzeugt (Padding NACH dem Glyph, Glyph am Feldende oder feste Breite 2), und den Umstand im Golden-Diff-Report dokumentieren.


## Prelude aus T3-Review (I01, medium — ZUERST, eigener Commit)

Plan-Step 4 des Englisch-Tasks versprach Fuzzy-Regressionstests und wurde als [x] abgehakt, aber NICHT geliefert: TestPalFilteredActionsFuzzyFiltered testet weiter nur 'bckl'→backlog (Reihenfolge unverändert = risikoärmster Fall). Die 5 wortgedrehten Labels (set status/tags/parent/blocking/title) sind ungetestet. Nachziehen: Subtests 'stat'→'set status' (+ nur die set-Einträge?) und 'go'→genau die 4 'go to'-Einträge (backlog/browse/repo picker/settings). Commit test(tui): Fuzzy-Regression verb-entity-Labels (T3-Review I01), Refs: bt-kyj5.

## Summary

Prelude: 2 neue Fuzzy-Regressionstests in overlay_palette_test.go nachgezogen. 'stat' matcht empirisch (fuzzyMatch ist Rune-Subsequence, nicht Substring) NICHT nur 'set status' sondern auch 'set parent' (s-t-a-...-t via "parent"s trailing t) — der Bean-Hinweis hedgte selbst mit "(+nur die set-Einträge?)"; verifiziert statt geraten, Test spiegelt die reale 2-Treffer-Menge. 'go' matcht exakt die 4 'go to ...'-Einträge wie erwartet.

Hauptaufgabe (PF-1/PF-3/PF-4/PF-12): Meta-Sektion ist nicht mehr kollabierbar (`isOpen := n == open || n == 1`) und wird eine 6-zeilige cursor-navigierbare Feldliste (title/status/type/priority/created_at/updated_at) statt der bisherigen 4-zeiligen Status/Type/Priority/Tags-Zusammenfassung — Tags ist NICHT mehr Teil von Meta (PF-4 legt das Feld-Set fest, Tags war dort nicht vorgesehen). Neuer Kopfblock (ID/NAME/Leerzeile/"type: X    status: Y    prio: Z", PF-3+PF-4 laut PO-Antwort Q01 verschmolzen) rendert oberhalb der Accordion in Tree UND Backlog (renderAccordionPane, EINZIGER verbleibender Aufrufer seit T1). Sektionstitel jetzt META/BODY/RELATIONS/HISTORY (PF-7-Rest, von T3 bewusst hierher verschoben).

Kopfblock-Formel exakt nach epic-E7-plan.md Task 4 Signatur-Block: type/status als WORT (theme.TypeStyle/StatusStyle + Rohwort), prio als GLYPH (theme.Priority — es gibt seit PF-6/T2 keine wortproduzierende PriorityStyle-Gegenstelle mehr). Meta-Feldliste dagegen nutzt für status/type/priority durchgängig T2s GLYPH-Output (StatusIcon/TypeIcon/Priority) — design-spec.md §15 PF-4 explizit "nutzt T2s Glyph-Output". Beide Stellen bewusst UNTERSCHIEDLICH, nicht symmetrisch — verifiziert gegen Plan-Formel statt angenommen (Auftrags-Hinweis "type/status/prio als WORT — prüfe design-spec, ob Wort oder Glyph" gezielt beachtet).

PF-12-Gutter-Fix: accordion.go reservierte die "▌"-Aktiv-Markierung bisher NUR bedingt (inaktive Header nutzten die VOLLE Breite ohne Prefix, aktive truncateten auf w-1) — eine Zeile verschob ihren eigenen Inhalt um 1 Spalte, sobald IHR EIGENER Auswahlzustand wechselte. Fix: beide Zweige reservieren jetzt IMMER exakt 1 Gutter-Spalte (" " inaktiv, "▌" aktiv), truncate konsistent auf w-1 — betrifft renderAccordion (Sektionsköpfe) UND fieldStrip (Relations-Feldliste, "alle markierbaren Zeilen" laut PF-12). Meta-Feldliste selbst ist von Anfang an mit fixem ▷/▶-Prefix gebaut (kein bedingtes Weglassen), erfüllt PF-12 strukturell ohne Retrofit.

## Test-Output

Prelude: `command go test ./internal/tui/... -run 'TestPalFilteredActionsFuzzy' -v` → PASS (3 Tests: bestehend + 2 neu). Voller Kurzlauf danach grün (`command go test ./... -short`).

RED→GREEN Hauptaufgabe: `command go build ./...` sauber (keine Compile-Fehler, Go-Testdateien separat). `command go test ./internal/tui/...` initial → FAIL (10 Compile-Fehler: beanSections/metaSectionBody Signatur-Mismatch, metaFields/detailHeaderBlock undefined) — erwartetes RED. Nach Implementierung: `command go test ./internal/tui/... -short` → 2 unerwartete FAILs in EIGENEN neuen PF-12-Tests (Spalten-Vergleich per `strings.Index` nutzte Byte-Offset statt Zell-Breite — "▌" ist 3 UTF-8-Bytes vs. 1 Byte für " ", falscher Vergleich). Gefixt (`lipgloss.Width` auf den Präfix-Substring statt rohem Byte-Index, neuer Helper `cellCol`). Danach `command go test ./internal/tui/... -short` → PASS, NUR die 2 erwarteten Golden-Diffs übrig (TestTreeGolden/TestBacklogGolden — Detail-Rendering ändert sich massiv, TestChromeGolden unberührt). Nach Golden-Regen: `command go test ./internal/tui/... -short` → PASS vollständig. Voller Lauf OHNE -short: `command go test ./...` → PASS, 135.8s (internal/tui, inkl. der 7 langsamen huh-Drive-Tests). `command go test ./... -race -short` → PASS. `command go vet ./...` → leer. `command gofmt -l .` → leer (1 Formatierungsfehler in accordion.go durch Kommentar-Ausrichtung gefunden+gefixt). `command go test ./internal/tui/... -run "TestTreeGolden|TestBacklogGolden|TestChromeGolden" -count=2` → PASS, beide Durchläufe stabil.

Gezielter Nachweis aller 22 T4-relevanten Tests (`-run 'TestBeanSections|TestMetaFields|TestMetaSectionBody|TestDetailHeaderBlock|TestRenderAccordion|TestFieldStrip|TestDetailFocusRightEntersFieldLevelOnlyForSectionsWithFields'`) → PASS, alle 22 grün.

## Golden-Diffs

`tree.golden` (28 Zeilen geändert): Detail-Pane zeigt jetzt Kopfblock ("gld-tsk1" / "First golden task" / Leerzeile / "type: task    status: todo    prio: ·") statt direkt der Accordion; [1] META rendert IMMER seinen Body (6 Feldzeilen mit ▷-Prefix: title/status/type/priority/created_at/updated_at) statt geschlossen zu sein (vorher: alle 4 Sektionen nur Header, "[1] Meta  ▸" geschlossen — Default-Ruhezustand accOpen=0 zeigte VORHER gar nichts, JETZT PF-1-Meta trotzdem offen); Sektionstitel Meta/Body/Beziehungen/Historie → META/BODY/RELATIONS/HISTORY. `backlog.golden` (30 Zeilen geändert): identisches Muster (Kopfblock "gbk-tsk2"/"Second backlog task"/"type: bug    status: draft    prio: ‼" + offene Meta-Feldliste + Sektionstitel-Großschreibung). `chrome.golden`: UNVERÄNDERT (Chrome() rendert keine Bean-Zeilen/Accordion, bestätigt per `git diff --stat` — kein Eintrag).

## Smoke

tmux (dieses Repo als Datenquelle, `bin/bt .`), Sequenz:
- bt-apmy (Milestone, alle Felder gesetzt inkl. Parent, Body, Children): Kopfblock korrekt ("bt-apmy"/"beans-tui v1 — devd-TUI-Port auf beans"/"type: milestone    status: todo    prio: !"), Meta offen by default OHNE tab (PF-1 bestätigt sichtbar auch ohne Fokus). tab → Detail-Fokus, Meta-Sektion zeigt sofort "▶ title:" (Feld-Marker unabhängig von detailLevel sichtbar, wie Plan-Formel `focused && activeIdx==0` vorschreibt — kein enter/right nötig). 3x Down (Sektionsebene) bewegt secCursor korrekt 0→1→2→3 (Body→Relations→History), Meta bleibt sichtbar offen (▷ überall, kein ▶, "▌"-Bar korrekt weg), History-Body erscheint. Ziffer '1' + right → Feldebene, 3x Down bewegt ▶-Marker title→status→type→priority, alle Zeilen bleiben spaltengleich ausgerichtet (visuell verifiziert). Ziffer '2' → Body öffnet zusätzlich, Meta bleibt WEITERHIN offen (PF-1 bestätigt bei Body aktiv).
- Schmale Fenster (90 dann 60 Spalten): Meta-Feldliste WRAPPED den langen Titel über 2 Zeilen (kein Datenverlust), Kopfblock-Titelzeile TRUNCATED mit "…" (Ein-Zeilen-Clip) — bewusst unterschiedliches Verhalten (Kopfblock truncate, Meta-Liste wrapText), beides ohne Crash/Überlauf.
- bt-6oyy (Feature, KEINE Relations — kein Parent/Children/Blocking/BlockedBy): Kopfblock korrekt ("type: feature    status: todo    prio: ·"), RELATIONS-Sektion zeigt "(no relations)"-Platzhalter korrekt.
- Explizites Vorher/Nachher-Shift-Assertion (PF-12, per capture-pane-Diff, nicht nur Unit-Test): capture-pane VOR (Tree fokussiert, kein Feld selektiert) vs. NACH (Detail fokussiert, title-Feld selektiert) — Spaltenposition aller 6 Feld-Label ("title:"/"status:"/"type:"/"priority:"/"created_at:"/"updated_at:") per Python-Diff verglichen: alle 6 IDENTISCH (z.B. "title:" beide Male Spalte 74) — PF-12 live im Terminal bestätigt, nicht nur in Unit-Tests.

PASS auf jedem Teilschritt, kein Crash, kein unerwarteter Shift.

## Deviations

- **update.go Compiler-Fix (nicht in Files-Liste, aber Signatur-Zwang):** `keyDetailFocus`s `beanSections(m.idx, b, 40)`-Aufruf (Zeile ~897) brauchte die 3 neuen Parameter, um zu kompilieren. Übergeben: `m.detailFocus, m.secCursor, m.fieldCursor` (semantisch neutral — diese 3 Werte beeinflussen nur `secs[0].body`s String-Rendering, NICHT `secs[...].fields`, was die einzige hier gelesene Struktur ist). Analog zu T1s "Compiler-gesteuerte Vollständigkeit"-Prinzip.
- **update_test.go Testfix (nicht in Files-Liste, aber Verhaltens-Konsequenz):** `TestDetailFocusRightEntersFieldLevelOnlyForBeziehungenSection` prüfte "right auf Meta (keine Felder) bleibt No-Op" — diese Prämisse ist durch PF-4 (Meta hat jetzt 6 Felder) hinfällig. Umbenannt zu `TestDetailFocusRightEntersFieldLevelOnlyForSectionsWithFields`, Assertion umgedreht (right auf Meta betritt jetzt Feldebene), Body bleibt als feldlose Negativ-Kontrolle. Dies ist eine BEABSICHTIGTE Konsequenz aus "hier nur RENDERING + Fokus-Navigation" (Auftrag) — die generische Section/Feld-Navigation (unverändert seit E2) funktioniert jetzt automatisch auch für Meta, ohne dass T4 selbst etwas an update.go ändern musste.
- **TestBeanSectionsMetaWrapsLongContentToBodyWidth umgewidmet:** B01 (E2-T1-quality-review) testete ursprünglich lange Tag-Namen — Tags ist seit PF-4 nicht mehr Teil von Meta. Test auf `TestBeanSectionsMetaWrapsLongTitleToBodyWidth` umgestellt (langer TITEL statt langer Tags), gleicher Zweck (Wrap-Regression-Schutz) auf das neue Feld-Set übertragen.
- **PF-12-Test-Interpretation (eigene Analyse, nicht 1:1 Plan-Wortlaut):** Plan-Text beschreibt den Test als "zwei Renders mit unterschiedlichem activeIdx — eine NICHT-aktive Zeile hat identische Breite in beiden". Wörtlich genommen (verschiedene ANDERE Sektion aktiv, Zielzeile bleibt in beiden inaktiv) hätte der Test schon VOR dem Fix bestanden (mathematisch nachgewiesen: `headerStyle.Width(w)` normalisiert die GESAMTBREITE einer Zeile immer auf exakt w, unabhängig vom Branch — der eigentliche Bug ist ein SPALTEN-Shift des Inhalts, keine Gesamtbreiten-Differenz). Implementiert stattdessen: Vergleich derselben Zeile EINMAL aktiv, EINMAL inaktiv (Spaltenposition des Titel-Texts) — das demonstrierbar VOR dem Fix rot ist (Spalte 6 vs. 7) und NACH dem Fix grün (beide 7). Bewusste Abweichung vom Plan-Wortlaut zugunsten eines Tests, der den beschriebenen Bug tatsächlich RED→GREEN zeigt.
- **EAW-Hinweis aus T2-Review (I01) — geprüft, kein Zusatzaufwand nötig:** Der Priority-Glyph (·/↓/→, EAW-Ambiguous) steht sowohl im Kopfblock ("...prio: X", Zeilenende) als auch in der Meta-Feldliste ("▷ priority:   X", Zeilenende) IMMER als letztes Zeichen der jeweiligen Zeile — es folgt kein weiterer Inhalt, der bei 1-vs-2-Zellen-Rendering verschoben werden könnte. Die im Auftrag genannten Absicherungen (Glyph ans Feldende ODER feste Breite mit Padding danach) sind durch die Layout-Struktur selbst bereits erfüllt ("Glyph ans Feldende"), kein zusätzlicher Code nötig. Im Golden-Diff-Report (oben) dokumentiert wie gefordert.
- **Prelude-Fuzzy-Test-Ergebnis weicht vom Bean-Hinweis ab:** Hinweis vermutete 'stat'→'set status' als (vermutlich) einzigen Treffer — empirisch verifiziert: 'stat' matcht AUCH 'set parent' (Rune-Subsequence s-t-a-...-t über "parent"s Endbuchstaben). Test spiegelt die reale 2-Treffer-Menge, nicht die im Hinweis vermutete 1-Treffer-Menge (Hinweis selbst hedgte das mit "(+nur die set-Einträge?)").

## Notes for T5 (Pane-Titel-Entfernung)

- `detailHeaderBlock`s 5 Zeilen (ID/Title/Leerzeile/type-status-prio/Leerzeile) werden in `renderAccordionPane` (view_browse_repo.go) VOR den Accordion-Zeilen in `rows` eingefügt — das ist reiner ROW-INHALT innerhalb des Panes, NICHT Teil von `renderPane`s eigener Titel+Trennlinien-Geometrie (`pane.title`/das `"Detail"`-Feld bleibt unverändert bestehen, T5 entfernt das). T5s `clickPaneGeometry`-`originY`-Kürzung (die 2 Zeilen Titel+Trennlinie) ist DAHER von T4s Änderungen komplett entkoppelt — der Kopfblock verschiebt nichts an der Pane-Chrome-Formel, er lebt ausschließlich INNERHALB des Content-Bereichs, den `clickPaneGeometry` laut Plan für Detail ohnehin nicht klickverfolgt ("Detail hat keinen Click-Row-Konsumenten").
- Falls ein künftiger Task (T6+) Klick-Handling für die Meta-Feldliste nachrüstet: `detailHeaderBlock` liefert IMMER exakt 5 Zeilen (unabhängig vom Bean-Inhalt — 2 feste Leerzeilen mit drin), danach folgt `renderAccordion`s Output beginnend mit `[1] META`s Header auf Zeile 6 (0-indiziert 5) relativ zum Pane-Content-Start. Diese Fixe-5-Zeilen-Konstante ist die relevante Zeilen-Mathe für spätere Geometrie-Arbeit am Detail-Pane.
- Bean bt-w9o8 (T3) hatte den Sektionstitel-Umbau bewusst an T4 abgegeben — das ist jetzt erledigt (META/BODY/RELATIONS/HISTORY), keine offene Abhängigkeit mehr für Folge-Tasks in dieser Richtung.

Refs: bt-kyj5, Commits fccb4b4 (Prelude), a9fcd64 (Hauptaufgabe)
