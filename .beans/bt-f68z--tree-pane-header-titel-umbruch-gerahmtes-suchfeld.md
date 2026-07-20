---
# bt-f68z
title: 'Tree-Pane: Header, Titel-Umbruch, gerahmtes Suchfeld'
status: completed
type: feature
priority: high
created_at: 2026-07-20T09:22:58Z
updated_at: 2026-07-20T10:16:37Z
parent: bt-vy1q
---

PO-Befunde aus dem Live-Test 2026-07-20 (#11-#13). Drei Punkte an der linken Pane —
zusammen erledigen, sie teilen sich Renderer und Golden.

## #11 — Suchfeld zu stark gemutet
`⌕ / search` verschwindet optisch. Soll ebenfalls **boxed** dargestellt werden, wie die
uebrigen Felder. Primitive: `dropdownBox`/`boxTopBorder`/`boxBottomBorder`.

## #12 — Tree-View hat keine Spaltenueberschriften
- **Master-Detail (schmal):** Header auf die Keybindings kuerzen — `S` = Status, `Title`.
- **Vollbild (`v`, breit):** Header ausschreiben.
Also breitenabhaengig, nicht zwei getrennte Implementierungen.

## #13 — Titel brechen nicht um, sondern werden abgeschnitten
Gewuenscht ist Umbruch mit **haengendem Einzug**, ausgerichtet auf den Titelbeginn:

```
  i B eq67 God-Delete/Purge crasht: ORM nullt
           child-FK statt Cascade (alert_configs
           NOT NULL)
▶ t M elir Bug Fixing - iOS-App
▶ t M 4x21 Canary-Instanz sproutling-test —
           Staged-Rollout & Dogfood auf NAS
```

**Achtung, Folgewirkung:** Zeilen sind dann nicht mehr 1:1 zu beans. Das betrifft
- `treeClickRow` / `flatClickRow` (Maus-Trefferzonen) — muessen Zeilen auf beans abbilden
- `windowStart`/Scroll-Rechnung — zaehlt Zeilen, nicht Elemente
- Cursor-Bewegung up/down — soll ueber beans springen, nicht ueber Zeilen

Praezedenzfall im Repo: `parentPickerRowBudget` hatte exakt diesen Fehler (Elemente statt
Zeilen gezaehlt), gefunden im 80x24-Smoke von bt-a3a8. Nicht wiederholen.

## Betroffen
`internal/tui/view_browse_repo.go` (`treeRowText`, Tree-Render), `view_browse_backlog.go`
(`backlogRowText`), `view_browse_flat.go`, `mouse.go`. Golden: `tree`, `backlog`,
`browse_flat`, `browse_boxform`, `chrome`.

## Akzeptanz
- [ ] Suchfeld gerahmt
- [ ] Header im Split gekuerzt, im Vollbild ausgeschrieben
- [ ] Titel brechen mit haengendem Einzug um
- [ ] Klick trifft nach Umbruch weiterhin das richtige bean (Test mit umbrechenden Titeln)
- [ ] up/down springen ueber beans, nicht ueber Zeilen
- [ ] tmux-Smoke bei 80 Spalten
- [ ] voller Testlauf gruen


## Summary

Alle drei PO-Befunde umgesetzt, Commit `a3fdb9e`.

- **#11 Suchfeld gerahmt** — `searchHeadBox` (view_browse_repo.go) baut auf den
  vorhandenen `boxTopBorder`/`boxBottomBorder`-Primitiven auf: Label `Search` im
  oberen Rahmen, Hotkey-Badge `(/)` im unteren, Rahmen wird bei Fokus Mauve.
  Gated auf `boxFormEnabled()` — im Nicht-Box-Modus bleibt die Ein-Zeilen-Variante.
  Damit `treeSearchLine` byte-identisch bleibt, wurde nur der TEXT in
  `searchHeadText()` extrahiert, nicht die Komposition.
- **#12 Spaltenkopf** — `treeColumnHeader(width)`, EINE Funktion, Schwelle
  `treeHeaderWideMin = 48`: schmal `S T ID   Title` (spaltengenau ausgerichtet),
  breit `Status Type ID   Title`. Bewusste Entscheidung: die ausgeschriebene
  Variante gibt die Ausrichtung der beiden Glyphen-Spalten auf — die sind
  konstruktionsbedingt 1 Zelle breit, "Status" kann dort nicht stehen.
- **#13 Titel-Umbruch mit haengendem Einzug** — `hangingWrap` + `blockWindow`
  in der neuen `line_window.go`.

**Der Folgewirkungs-Teil (der eigentliche Aufwand):** Zeilen sind nicht mehr
1:1 zu beans. Statt Elemente zu fenstern, liefert `blockWindow` jetzt die
gefensterten ZEILEN **und** ein Zeile→bean-Mapping zurueck. Renderer und
Maus-Hit-Test rufen dieselbe Funktion mit denselben Argumenten auf — sie
koennen nicht auseinanderlaufen. Betrifft Tree, Browse-Flat und Backlog
gleichermassen (`treeBlocks`/`flatBlocks`/`backlogBlocks`, alle drei teilen
sich `decorateRowBlock`, das vorher dreimal kopiert war).

- **Zeilen-Cap `maxRowLines = 3`** (Implementer-Entscheidung): ohne Deckel
  haette ein 300-Zeichen-Titel in einer schmalen Pane ein Dutzend Zeilen
  erzeugt und alle anderen beans verdraengt. Letzte Zeile bleibt sichtbar
  mit `…` abgeschnitten.
- **Cursor unveraendert** — `treeCursorMove`/`listState.move` laufen weiter
  ueber bean-Indizes. Nur Rendern und Treffer-Test sind zeilen-bewusst.
- Der `parentPickerRowBudget`-Fehler wird durch
  `TestBlockWindowNeverExceedsHeight` (40 beans, gemischte Hoehen, Hoehen
  1..23, Cursor an allen Raendern) direkt abgedeckt.

**Nebenbefund, im Zuge der Umbauten mitgefixt:** `treeClickRow`/`flatClickRow`
zogen die Filter-Leisten-Hoehe nur von `originY` ab, nicht von `bodyH` — das
ist die Ursache von bt-vpvu, siehe dort.

## Test-Output

Voller Lauf, ohne `-short`, im Vordergrund:

```
?   	github.com/xRiErOS/beans-tui	[no test files]
ok  	github.com/xRiErOS/beans-tui/cmd	1.106s
?   	github.com/xRiErOS/beans-tui/internal/clip	[no test files]
ok  	github.com/xRiErOS/beans-tui/internal/config	0.358s
ok  	github.com/xRiErOS/beans-tui/internal/data	4.076s
ok  	github.com/xRiErOS/beans-tui/internal/theme	1.409s
ok  	github.com/xRiErOS/beans-tui/internal/tui	152.966s
```

Neue Tests: `line_window_test.go` (9), `view_browse_tree_head_test.go` (13).

**tmux-Smoke 80 Spalten** gegen sproutling, frisches Binary, frische Session,
beide Modi (`BT_BOXFORM=1` und Default): kein Overflow, Rahmenhoehe exakt 24,
Umbruch mit haengendem Einzug sichtbar, Suchbox und Spaltenkopf korrekt.

**Regenerierte Golden:** `tree`, `backlog`, `browse_flat`, `browse_boxform`.
Jeder Diff Zeile fuer Zeile geprueft — enthaelt ausschliesslich Spaltenkopf,
Umbruch-Zeilen mit haengendem Einzug, den `▌`-Balken auf allen Zeilen des
Cursor-Blocks und (nur boxform) die Such-Box. `chrome.golden` blieb
unveraendert und musste NICHT regeneriert werden.

## Deviations

- **Spaltenkopf nur in der Browse-Pane, nicht im Backlog.** Der bean nennt
  unter #12 ausdruecklich die Tree-View; der Backlog hat eine eigene Kopfzeile
  mit Sort-Suffix. Umbruch (#13) wirkt dagegen sehr wohl auch im Backlog, weil
  der bean `backlogRowText` unter "Betroffen" fuehrt.
- **#11 ist gated, #12/#13 sind es nicht.** Die Rahmen-Optik gehoert zur
  Box-Form (Epos-Constraint "additiv + gated"); Spaltenkopf und Umbruch sind
  allgemeine Tree-Verbesserungen und aendern deshalb auch die
  Nicht-Box-Golden — genau die Golden-Liste, die der bean vorgibt.
- **`maxRowLines = 3`** war nicht vorgegeben, siehe Begruendung oben.
- **`▌`-Cursorbalken auf JEDER Zeile des Cursor-blocks** statt nur der ersten:
  sonst liest sich ein zweizeiliger bean als zwei halb-selektierte.
