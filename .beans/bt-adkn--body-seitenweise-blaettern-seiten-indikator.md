---
# bt-adkn
title: Body seitenweise blaettern + Seiten-Indikator
status: completed
type: feature
priority: normal
created_at: 2026-07-20T09:23:37Z
updated_at: 2026-07-21T11:19:18Z
parent: bt-vy1q
---

PO-Befund 2026-07-20 (#5): Der Body soll **seitenweise blaetterbar** sein, mit einem
Seiten-Indikator aus hell-/dunkelgrauen Punkten unten rechts im Rahmen.

## ACHTUNG — die vorgeschlagenen Keys sind belegt
Der PO schlug `ctrl+k` / `ctrl+i` vor. Beides geht nicht:
- **`ctrl+i` IST Tab.** Im Terminal identisch (0x09) — nicht unterscheidbar. Kollidiert
  frontal mit dem gewuenschten Tab-Feldwechsel (#2, bean fuer das Fokus-Modell).
- **`ctrl+k` ist die Command-Palette.** Steht so im Header (`ctrl+k commands`).

Ausserdem verbietet `TestKeymapNoCtrlSQ` bereits `ctrl+s`/`ctrl+q` (XOFF/XON) — es gibt
also Praezedenz dafuer, Terminal-Kollisionen hart auszuschliessen.

**Ersatz muss mit dem PO abgestimmt werden.** Kandidaten: `pgup`/`pgdn` (semantisch exakt
"blaettern", kollisionsfrei), `ctrl+f`/`ctrl+b` (vi/less-Konvention), oder `[`/`]`.
Empfehlung: `pgup`/`pgdn` als Primaerbindung, weil Blaettern genau ihre Bedeutung ist.

## Seiten-Indikator
Punkte unten rechts im Body-Rahmen, gefuellt = aktuelle Seite. Rahmen-Render:
`boxBottomBorder` (box_dropdown.go/box_detail_form.go) — dort sitzt schon das Hotkey-Badge
rechts, der Indikator muss sich den Platz mit ihm teilen.
**Beachte #4:** dort wird der Body-Hotkey moeglicherweise nach OBEN verlegt, weil der
untere Rahmen bei langem Body wegscrollt. Dann gilt dasselbe fuer den Indikator — ein
Seiten-Anzeiger, den man beim Blaettern nicht sieht, ist sinnlos. **Beide zusammen denken.**
Theme-Token nur aus `internal/theme/` (hell/dunkel = vorhandene Grau-Token, keine Hex).

## Verhaeltnis zum vorhandenen Scroll
bt-ze10 hat zeilenweises Scrollen gebaut (`adjustBoxFormScroll`, geclamped, plus Mausrad);
bt-1o4g laesst die Pfeiltasten Felder ansteuern und den Scroll mitziehen. Seitenweises
Blaettern kommt HINZU und muss durch denselben Mutationspunkt laufen
(`adjustBoxFormScroll`), sonst driften die Pfade auseinander.

## Akzeptanz
- [ ] Ersatz-Keybinding mit dem PO abgestimmt (ctrl+i/ctrl+k sind ausgeschlossen)
- [ ] Blaettern laeuft ueber `adjustBoxFormScroll`, kein zweiter Scroll-Pfad
- [ ] Seiten-Indikator sichtbar, auch waehrend des Blaetterns
- [ ] voller Testlauf gruen


## PO-Entscheidung 2026-07-20: `pgup` / `pgdn`

Blaettern liegt auf `pgup`/`pgdn` — bedeutet genau das, kollidiert mit nichts.
`ctrl+i` (== Tab) und `ctrl+k` (Palette) sind ausgeschlossen; siehe bt-mx4k, wo `ctrl+k`
ganz entfaellt, und bt-8d35 zur Tab-Belegung.

## Summary (2026-07-21)

Seitenweises Body-Blaettern (pgup/pgdn) + Seiten-Indikator gebaut, beide hinter BT_BOXFORM.

**Teil a — Paging:** pgup/pgdn im `keyDetailFocus` (update.go), Delta = sichtbares Zeilen-Budget (`boxFormScrollBounds` height), durch DENSELBEN `adjustBoxFormScroll` wie up/down + Mausrad — ein Mutationspunkt, kein zweiter Scroll-Pfad (Load-bearing C erfuellt). Bindings `boxFormPageUp/Down` als standalone vars in box_nav_field.go.

**Teil b — Indikator:** `●○`-Punktreihe unten rechts im **aeusseren** renderPane-Rahmen (`overlayPaneBottomBadge`, render_shared.go), NICHT im Body-Panel-Rahmen. `boxFormPageIndex`/`boxFormPageBadge` (box_form_page.go): count=ceil(total/h), Clamp-Ceiling→letzte Seite. Grau-Token Subtext(hell)/Surface(dunkel), keine Hex. >count Seiten als Cells → numerischer `n/N`-Fallback.

## Akzeptanz
- [x] Ersatz-Keybinding PO-abgestimmt: pgup/pgdn (ctrl+i==Tab / ctrl+k==Palette ausgeschlossen)
- [x] Blaettern laeuft ueber adjustBoxFormScroll, kein zweiter Scroll-Pfad
- [x] Seiten-Indikator sichtbar, auch waehrend des Blaetterns (fixe Aussen-Chrome)
- [x] voller Testlauf gruen

## Test-Output (RED→GREEN)
RED: `boxFormScroll after one pgdn = 0, want exactly one page (18)` + `undefined: boxFormPageIndex/pageDotFilled`.
GREEN: `ok internal/tui 151.296s`, gesamter `go test ./...` exit 0. Neue Tests: TestBoxFormPageDownUpScrollsByPage, PageDownClampsAtEnd, PageInertWithoutFlag, PageIndexMapsOffsetToPage, PageIndicatorTracksPage, PageIndicatorAbsentWhenFits.

## Smoke (80 Spalten, tmux, reales Repo)
Initial `●○○…` (Seite 0), nach 2×PageDown `○○●○…` (Seite 2) — Indikator auf fixem Rahmen, beim Blaettern sichtbar, kein Wrap, Rahmen schliesst sauber.

## Deviations/ERRATA
- **Binding-Ort:** SSTD sagte 'keymap.go Single Source', doch der verifizierte Code-Praezedenzfall (Sibling boxFormFieldNext/Prev) legt box-form-REGION-lokale Bindings bewusst als standalone vars in box_nav_field.go ab — sonst schlaegt TestHelpGroupsCoverEveryBindingExactlyOnce an (reflektiert nur keyMap-Felder). Dem Code gefolgt (CLAUDE.md: bei Konflikt dem Beobachteten trauen).
- Goldens browse_boxform + value_menu_anchored regeneriert: einzige Aenderung = `●○` im Detail-Pane-Unterrahmen (stripped diff verifiziert, keine Struktur-Drift). Flag-AUS-Goldens byte-identisch.

## Notes for T(n+1) (bt-p78f #4)
Body-Hotkey liegt bereits oben (bt-oox1, panelBoxTopHotkey). Der Seiten-Indikator sitzt auf dem AEUSSEREN renderPane-Rahmen (fix), NICHT auf boxBottomBorder — bt-p78f muss ihn daher NICHT mit verschieben. Falls bt-p78f am aeusseren Unterrahmen etwas aendert: overlayPaneBottomBadge teilt sich diese Zeile.

## Review 2026-07-21 (PO, rejected)

**US-01 (seitenweises Blaettern) · r:** woertlich: "wenn ich PageDown / PageUp verwende, dann scrollt das gesamte linke pane."

**US-02 (Seiten-Indikator) · r:** "siehe US-01" + NB unten.

**NB-A (PO):** "der [Indikator] wechselt nicht die Darstellung, wenn ich mit PageDown 'umblaettere'."
**NB-B (PO):** "er sitzt falsch im linken Pane und muesste an der Box fuer 'Body' visualisiert werden."

### Fix-Preludes

**B1 · critical · PageDown/PageUp scrollen die Tree-Pane statt des Box-Form-Body.**
- Fundort: Key-Routing in `handleKey`/`update.go` — die pgup/pgdn-Cases liegen in `keyDetailFocus`, werden real aber NICHT dorthin geroutet; die Tree/List-Viewport-Scroll-Ebene konsumiert pgup/pgdn zuerst.
- Test-Luecke (Ursache, warum gruen trotz Bug): `TestBoxFormPageDownUpScrollsByPage` setzt `m.detailFocus=true` und schickt `keyMsg(tea.KeyPgDown)` durch `step`→`Update` — umgeht damit die reale `handleKey`-Routing-Reihenfolge. Es fehlt ein Integrationstest, der pgup/pgdn ueber den vollen handleKey-Pfad bei in-Detail-Fokus fuehrt (analog boxFormClickAt-Muster) UND einen Gegentest, dass die Tree-Pane pgup/pgdn NICHT mehr frisst, solange Detail fokussiert ist.
- Fix-Rezept: pgup/pgdn im Detail-Fokus VOR dem Tree/List-Scroll abfangen (Checkpoint-Reihenfolge in handleKey), durch denselben `adjustBoxFormScroll`. Reales 80c-tmux-Smoke mit echtem PageDown ist Pflicht (Unit-Test hat den Routing-Bug strukturell nicht gesehen).

**B2 · high · Indikator aktualisiert sich beim Blaettern nicht** — direkte Folge von B1 (kein Body-Scroll → Offset konstant → gleicher Punkt). Faellt mit B1, aber im Smoke separat verifizieren (Punkt wandert bei PageDown).

**B3 · medium · Indikator-Platzierung (Design-Revision D02→?).** PO will den Indikator AN DER BODY-BOX, nicht am aeusseren Pane-Rahmen. Konflikt mit dem urspruenglichen Akzeptanzkriterium 'sichtbar auch waehrend des Blaetterns' (die Body-Box-Unterkante scrollt bei langem Body weg — genau die Kopplung zu bt-p78f #4). PO-Wunsch hat Vorrang; vor Reimplementierung klaeren, WIE der Indikator an der Body-Box sichtbar bleibt (z.B. an der Body-TOP-Border, wo schon das (e)-Badge sitzt, statt am Bottom). Zusaetzlich PO-Beobachtung 'im linken Pane' pruefen — overlayPaneBottomBadge soll nur die rechte Detail-Pane treffen; im realen Zwei-Pane-Layout gegenchecken.

## Rework-Plan (2026-07-21, PO-Freigabe, empirisch gegroundet)

**Smoke-Repro (BT_BOXFORM=1, 80x24, langer Body):**
- Tree-Fokus (Default, kein Tab) + PageDown = NO-OP: Body pagt nicht, Punkt bleibt Seite 0.
- Nach Tab (Detail-Fokus) + PageDown = pagt korrekt: Body scrollt, Punkt wandert (●→Seite2), Indikator am rechten Detail-Aussenrahmen, aktualisiert sich.

**Revision der Review-Befunde:**
- **B1 (Kern, high):** Paging haengt an detailFocus (keyDetailFocus, update.go:1191/1364). PO erwartet pgup/pgdn auf dem SICHTBAREN Body OHNE Tab — analog Mausrad (positions-/fokus-unabhaengig, boxFormWheelHit/wheelMove, mouse.go). Routing-Ursache: pgup/pgdn erreichen keyDetailFocus nur bei detailFocus; bei Tree-Fokus geht es an keyTree (navKey-Switch matcht pgup/pgdn nicht → no-op).
- **B2:** KEIN eigener Bug — Symptom von B1 (kein Body-Scroll → Punkt statisch). Faellt mit B1 (im Smoke bewiesen: bei Detail-Fokus wandert der Punkt).
- **B3 (medium, Design-Entscheidung offen):** PO will Indikator AN DER BODY-BOX statt am aeusseren Pane-Rahmen. Spannung zu 'sichtbar beim Blaettern' (Body-Box-Rahmen scrollt weg). PO-Beobachtung 'linkes Pane' im Repro NICHT bestaetigt (Indikator sitzt rechts) — vermutlich Fehldeutung; im Zwei-Pane-Layout final gegenchecken.

**Task-Plan:**
- T1 (B1+B2): pgup/pgdn zu GLOBALEM handleKey-Checkpoint (gated boxFormEnabled() + focusedBean()!=nil, ausser overlay/form/fullscreenList), route zu adjustBoxFormScroll um eine Seite — VOR keyNodeAction/detailFocus/keyTree-Dispatch. Aus keyDetailFocus entfernen (oder global faengt zuerst). TDD: Integrationstest ueber VOLLEN handleKey bei TREE-Fokus (nicht nur detailFocus) — pagt Body + Tree-Cursor bleibt. Gegentest: keyTree bewegt Cursor bei pgup/pgdn NICHT. Reales 80c-Smoke Pflicht (Unit-Test sah Routing strukturell nicht).
- T2 (B3): Indikator-Platzierung nach PO-Entscheidung (siehe Chat). Zwei-Pane-Layout gegenchecken.

Verifikation: command go test ./... (voll) + 80c-tmux-Smoke mit echtem PageDown bei Tree-Fokus.

## B3 Design-Entscheidung (PO, 2026-07-21): Sticky Body-Kopfzeile
Indikator sitzt rechts in der Body-Titelzeile ('Body ... (e)'), die als FIXE Zeile oben im Detail-Viewport gerendert wird; der Body-Text scrollt darunter. Ergebnis: an der Body-Box UND immer sichtbar. Umsetzung T2: Body-Panel in sticky-Header (Titel + (e)-Badge + Indikator rechts) + scrollenden Inhalt trennen; overlayPaneBottomBadge am Aussenrahmen entfaellt.

## Rework umgesetzt (2026-07-21, agent-abgeschlossen)

### Summary
- **T1 (B1+B2):** pgup/pgdn zu FOKUS-UNABHAENGIGEM globalem handleKey-Checkpoint (update.go, nach Refresh/vor keyNodeAction; gated boxFormEnabled() + focusedBean()!=nil + fullscreen!=fullscreenList). Blaettert den Body OHNE Tab, analog Mausrad, durch dieselbe adjustBoxFormScroll-Mutation. Aus keyDetailFocus entfernt.
- **T2 (B3):** Seiten-Indikator sitzt jetzt in der BODY-Titelzeile (boxTopBorderBadges, box_dropdown.go), rechts vor dem (e)-Badge; overlayPaneBottomBadge (Aussenrahmen) entfernt. Body-Titelzeile ist STICKY: sobald in den Body geblaettert wird, wird sie an Viewport-Zeile 0 geheftet, Body-Text scrollt darunter (renderAccordionPane). Bei mehr Seiten als Punkte passen: kompakte n/N-Form.

### Test-Output (RED->GREEN)
- RED: TestBoxFormPageKeysPageBodyAtTreeFocus — 'boxFormScroll after tree-focused pgdn = 0, want one page (18)' (B1 reproduziert: no-op bei Tree-Fokus).
- RED: TestBoxFormBodyHeaderCarriesPageDots / ...StickyWhenPaged — Dots am Aussenrahmen statt Body-Header / kein Pin.
- GREEN nach Fix: alle 3 + TestBoxFormBodyIndicatorRendersWhenPagesExceedDotBudget (Regression fuer den Smoke-Bug). Voller Lauf 'command go test ./...' gruen (internal/tui 152s).
- Mutations-Stichprobe: dotsBudget accW-18 -> accW zuruck => TestBoxFormBodyIndicatorRendersWhenPagesExceedDotBudget FAILT (Indikator verworfen). Fix wiederhergestellt.

### Smoke (80x24, BT_BOXFORM=1, echtes Repo)
PageDown bei TREE-Fokus (kein Tab): Body pagt, Tree-Cursor bleibt auf vy1q (linkes Pane scrollt NICHT mehr — B1/PO-Reject behoben). Body-Header sticky oben, zeigt '2/25' (25-Seiten-Body -> n/N-Form). PageUp zurueck -> '1/25'. 80c-Rahmen sauber, kein Overflow.

### Deviations/ERRATA
- Smoke deckte einen Bug auf, den die Unit-Tests strukturell nicht sahen: bei VIELEN Seiten (count > Header-Dot-Budget) verwarf boxTopBorderBadges die zu breite Punktreihe KOMPLETT -> gar kein Indikator. Fix: Dot-Budget = tatsaechlicher Header-Freiraum (accW-18), dann greift die kompakte n/N-Form. Neuer Regressionstest schliesst die Luecke (mutations-verifiziert).
- B3-Grenze: der Indikator ist an die Body-Titelzeile gekoppelt; blaettert man ganz PAST den Body in Relations/History (nur bei kurzem Body relevant), entpinnt der Header und der Indikator verschwindet. Bewusst so (PO-Entscheidung: Indikator = Body-Titelzeile). Beim langen Body (dem Blaetter-Anwendungsfall) ist man stets im Body -> immer sichtbar.

### Notes
GLOSSARY 'Seiten-Indikator' aktualisiert (Body-Header + sticky + fokus-unabhaengig). Kopplung bt-p78f (#4) obsolet — TOC-Experiment scrapped, Body-Hotkey bleibt oben.

## Review 2026-07-21 (US-02 rejected)
**PO woertlich:** 'Die Position passt noch nicht sauber. das (e) immer an der gleichen Stelle belassen und dafuer ── ●○ verschieben. Das fuehrt zu einer stabilen praesentation.'

**Fix-Prelude · medium · Layout-Stabilitaet des Body-Header-Indikators**
- **Fundort:** boxTopBorderBadges (internal/tui/box_dropdown.go) + Dots-Platzierung in renderAccordionPane (internal/tui/view_browse_repo.go).
- **Ist:** Indikator (●○ / n-N) ist RIGHT-geparkt direkt LINKS vom (e)-Badge. Dadurch schwankt die linke Kante des Indikators mit seiner Breite (Punktzahl bzw. 1/25 vs 10/25) — wirkt unruhig, obwohl (e) rechts fixiert ist.
- **Soll:** (e) bleibt fest ganz rechts (unveraendert). Indikator NACH LINKS ankern — direkt hinter das 'Body'-Label: '╭─ Body ── ●○ ─────────── (e) ─╮'. Dashes fuellen die Mitte. Beim Blaettern aendert sich nur, WELCHER Punkt hell ist, nicht die Position → stabile Praesentation. Gilt auch fuer die n/N-Fallback-Form (ebenfalls links).
- **Verifikation:** Golden neu (browse_boxform / value_menu_anchored) + 80c-tmux-Smoke: Indikator-Position konstant beim Durchblaettern, (e) unbewegt.
- **Quelle:** PO-Review 2026-07-21.

US-01 (Blaettern ohne Tab) + US-03 (n/N-Fallback bei vielen Seiten) im selben Review ACCEPTED — nur die Indikator-Position (US-02) ist Nacharbeit.

## US-02-Nacharbeit umgesetzt (2026-07-21, agent-abgeschlossen)

### Summary
Seiten-Indikator in der Body-Titelzeile LINKS verankert (direkt hinter 'Body ─'), (e) fest ganz rechts: '╭─ Body ─ 1/27 ─────── (e) ──╮'. Beim Blaettern aendert sich nur die Zahl/der helle Punkt, nicht die Position — (e) und Indikator-Linkskante bleiben fix, nur die mittleren Dashes absorbieren Breitenaenderungen. boxTopBorderBadges (box_dropdown.go) von right-parked auf left-anchored umgebaut; dotsBudget accW-20 (neue Reservierung).

### Test-Output (RED->GREEN)
- RED: TestBoxTopBorderBadgesStableIndicatorPosition — 'indicator not left-anchored: starts at col 133 (narrow) vs 127 (wide)'.
- GREEN nach Fix + bestehende Header-/Golden-Tests. Voller Lauf 'command go test ./...' gruen (internal/tui 153s). Golden neu (browse_boxform / value_menu_anchored).

### Smoke (80x24, echtes Repo, 27-Seiten-Body)
PageDown x2: '1/27' -> '2/27' -> '3/27'. Indikator startet konstant an derselben Spalte, (e) unbewegt, mittlere Dashes konstant (Zahlbreite gleich) — stabile Praesentation wie vom PO gefordert.

### Notes
US-01 + US-03 im Review bereits accepted. US-02 (Position) mit dieser Nacharbeit erledigt; PO-Endabnahme im Chat. GLOSSARY-Skizze weiterhin gueltig (Indikator in Body-Header).

## US-02-Nacharbeit v2 (2026-07-21) — (e)-Sprung an Stellen-Grenze

### Root-Cause (2. Reject 'das (e) springt weiterhin', schmale Geometrie)
Der Indikator war zwar links verankert, aber die 'n/N'-Zahl waechst beim Blaettern ueber Stellen-Grenzen (9→10→100). Bei schmaler Breite nahe der Fit-Schwelle kippt sie dabei ueber das Budget → Indikator faellt auf den wideren Seiten weg → (e) springt auf die badge-lose Fallback-Position (boxTopBorderHotkey). An breiten Terminals unsichtbar (Dashes absorbieren), an schmalen (bt-sprout) sichtbar.

### Fix
boxFormPageBadge: 'n/N'-Zahl auf die Stellenzahl von count gepaddet (' 9/91' / '13/91' / '100/102') → Badge-Breite pro Bean KONSTANT ueber alle Seiten. Damit weder (e)-Verschiebung noch Wegfall mitten im Blaettern: der Badge passt (oder nicht) auf jeder Seite identisch.

### Test-Output (RED->GREEN)
- RED: TestBoxFormPageBadgeConstantWidthAcrossPages — 'badge width not constant: 1->5 50->6 100->7'.
- GREEN nach Fix. Voller Lauf gruen (internal/tui 151s), kein Golden-Drift.

### Smoke (schmal, kritischer Fall)
- W=58, Seite 9→13: '─  9/91 ──── (e) ──╮' → '─ 13/91 ──── (e) ──╮' — (e) identische Spalte, Zahl rechtsbuendig.
- W=54 (zu schmal): Indikator konsistent weggelassen, (e) konstant — kein Flackern.
Reproduziert mit vorhandenen Langtext-Beans (bt-vy1q, 27–102 Seiten je nach Breite); kein Wegwerf-Test-Bean noetig.
