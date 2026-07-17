---
# bt-nxuk
title: 'Footer-Hint bei offenem Filter-Menü: tab-Label kontextfalsch'
status: completed
type: task
priority: low
created_at: 2026-07-17T08:44:12Z
updated_at: 2026-07-17T11:36:38Z
parent: bt-5uzr
---

Reviewer-Finding B04 aus bt-2p9m-Review (2026-07-17, non-blocking): Bei offenem Filter-Menü rendert der Footer-Kontext-Hint weiterhin das wörtliche `keys.FocusIn`/`FocusOut`-Help („tab focus in · shift+tab focus out"), obwohl tab/shift+tab dort die Facet-Kategorie wechseln. Die Modal-interne Hint-Zeile („tab/shift+tab:category") ist korrekt — Fußzeile inkonsistent für Nutzer, die nur unten lesen.

Fix-Optionen (Reviewer): eigene `keybind.Binding`-Instanzen mit kontextspezifischem Help-Text nur für `filterMenuLocalBindings`, ODER `contextualLocalHint` überschreibt Label bei `m.filterOpen`. Drift-Guards (`TestNoDuplicateBindingBetweenGlobalAndAnyLocalHintList`, `TestHelpGroupsCoverEveryBindingExactlyOnce`) beachten.

Akzeptanz:
- [ ] Footer-Hint zeigt bei offenem Filter-Menü kategoriewechsel-korrektes Label
- [ ] Drift-Guard-Tests grün
- [ ] Test-Suite grün


## Plan-Konkretisierung E13 (2026-07-17)

Plan: `docs/plans/v1-port/epic-E13-plan.md` §„Item 5: Footer-Hint bei
offenem Filter-Menü zeigt falsches Label". Reihenfolge: Parallel-Welle (mit
`bt-2kfl`/`bt-d3ps`, disjunkte Dateien), NACH der Toast-Familie.

**Root Cause (file:line, verifiziert gegen Ist-Code 2026-07-17):**
`footer_context.go:43-45` (`filterMenuLocalBindings`) gibt `keys.FocusIn`/
`keys.FocusOut` DIREKT zurück — dieselben globalen `keybind.Binding`-Werte
mit `WithHelp("tab", "focus in")`/`WithHelp("shift+tab", "focus out")`
(`keymap.go:136-137`). `renderBindings` (`view.go:128-137`) rendert das
globale Label unverändert, obwohl `keyFilterMenu`
(`box_filter_facets.go:380-397`) `tab`/`shift+tab` tatsächlich an
`filterMenuSwitchTab` (Kategorie-Wechsel) bindet — das Modal-interne Hint
(`treeFilterBox`, `box_filter_facets.go:476`) zeigt bereits korrekt
„tab/shift+tab:category".

**Vorgehen:** EIN neues, eigenständiges `keybind.Binding` (NICHT als
`keyMap`-Struct-Feld) mit `WithKeys("tab", "shift+tab")`,
`WithHelp("tab/shift+tab", "category")` — ersetzt die zwei
`keys.FocusIn, keys.FocusOut`-Einträge in `filterMenuLocalBindings()`
(`footer_context.go:44`).

**Drift-Guards verifiziert (kein Konflikt):**
`TestNoDuplicateBindingBetweenGlobalAndAnyLocalHintList`
(`keymap_test.go:208-225`) scoped nur `browseRepoLocalBindings`/
`backlogLocalBindings` — `filterMenuLocalBindings` außerhalb des Scans.
`TestHelpGroupsCoverEveryBindingExactlyOnce` (`keymap_test.go:233-274`)
reflektiert nur über `keyMap`-Struct-Felder — ein lokales, nicht-`keyMap`-
Binding ist unsichtbar für diesen Test.

**Akzeptanz:**
- [ ] Footer Zone 3 zeigt bei offenem Filter-Menü „tab/shift+tab category"
      statt „tab focus in · shift+tab focus out"
- [ ] Globales `tab`/`shift+tab` außerhalb des Filter-Menüs unverändert
- [ ] `TestNoDuplicateBindingBetweenGlobalAndAnyLocalHintList` grün
- [ ] `TestHelpGroupsCoverEveryBindingExactlyOnce` grün
- [ ] Test-Suite grün, neuer Test
      `TestFilterMenuFooterHintShowsCategoryLabel`

## Summary

Root Cause bestätigt (Plan Item 5, footer_context.go:43-45): `filterMenuLocalBindings()`
gab `keys.FocusIn`/`keys.FocusOut` (globale Bindings, `WithHelp("tab","focus in")`/
`WithHelp("shift+tab","focus out")`) direkt zurück -- Footer Zone 3 rendert deren
Help-Label unverändert, obwohl `keyFilterMenu` tab/shift+tab tatsächlich an die
Facet-Kategorie-Umschaltung bindet und das Modal-eigene Hint (`treeFilterBox`)
bereits korrekt "tab/shift+tab:category" zeigt.

Fix: neues, eigenständiges package-level `keybind.Binding` `filterMenuCategoryHint`
(`WithKeys("tab","shift+tab")`, `WithHelp("tab/shift+tab","category")`) in
footer_context.go, NICHT als keyMap-Struct-Feld -- ersetzt die zwei
`keys.FocusIn, keys.FocusOut`-Einträge in `filterMenuLocalBindings()`. Globales
tab/shift+tab-Verhalten außerhalb des Filter-Menüs unverändert (reine Footer-
Anzeigeänderung, Dispatch bleibt in keyFilterMenu).

## Test-Output

RED (vor Fix, `TestFilterMenuFooterHintShowsCategoryLabel`):
```
footer_context_test.go:61: filterOpen: contextualLocalHint = "↑/i up · ↓/k down · tab focus in · shift+tab focus out · space/x Toggle facet · X Clear filters · enter open/confirm · esc back", want it to contain the filter-menu-local "tab/shift+tab category" hint (bt-nxuk)
footer_context_test.go:64: filterOpen: contextualLocalHint = "...", must NOT contain the global FocusIn/FocusOut "focus in"/"focus out" label (bt-nxuk)
--- FAIL: TestFilterMenuFooterHintShowsCategoryLabel (0.00s)
```

GREEN (nach Fix):
```
--- PASS: TestFilterMenuFooterHintShowsCategoryLabel (0.00s)
ok  	beans-tui/internal/tui	0.449s
```

Voller Lauf (ohne `-short`), Dauer 2:31.51 total (internal/tui 150.21s):
```
?   	beans-tui	[no test files]
ok  	beans-tui/cmd	0.970s
?   	beans-tui/internal/clip	[no test files]
ok  	beans-tui/internal/config	0.624s
ok  	beans-tui/internal/data	2.492s
ok  	beans-tui/internal/theme	1.230s
ok  	beans-tui/internal/tui	150.210s
```

Drift-Guards (gezielt, beide grün):
```
--- PASS: TestNoDuplicateBindingBetweenGlobalAndAnyLocalHintList (0.00s)
--- PASS: TestHelpGroupsCoverEveryBindingExactlyOnce (0.00s)
```

Golden ×2 (byte-identisch, filterOpen ist NICHT Teil der Golden-Fixtures --
chrome.golden/tree.golden/backlog.golden decken nur den Basis-Zustand ab):
```
--- PASS: TestChromeGolden / TestTreeGolden / TestTreeGoldenDeterministic / TestBacklogGolden / TestBacklogGoldenDeterministic  (beide Läufe identisch, 2. Lauf cached)
```

`go vet ./...` clean, `gofmt -l` clean (keine Ausgabe).

## Smoke

tmux 80x24, Session btnxuk57820:
- Baseline: `tab focus in · shift+tab focus out · / search · f Filter · s Status · c Create · d Delete · e Edit · b Backlog · t Tags · y Yank · a Parent · r Blocking`
- `f` (Filter-Menü offen): `↑/i up · ↓/k down · tab/shift+tab category · space/x Toggle facet · X Clear filters · enter open/confirm · esc back`
- `esc` (Menü zu): Footer wieder `tab focus in · shift+tab focus out · ...` (global unverändert)

## Deviations/ERRATA

Keine. Vorgehen exakt wie im Plan (Item 5) konkretisiert: EIN neues,
eigenständiges Binding statt keyMap-Struct-Feld, beide Drift-Guards
verifiziert unberührt (Scope-Analyse des Planners bestätigt).
