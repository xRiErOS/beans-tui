---
# bt-nxuk
title: 'Footer-Hint bei offenem Filter-Menü: tab-Label kontextfalsch'
status: todo
type: task
priority: low
created_at: 2026-07-17T08:44:12Z
updated_at: 2026-07-17T10:08:09Z
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
