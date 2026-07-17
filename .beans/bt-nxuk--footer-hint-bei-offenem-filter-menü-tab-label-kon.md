---
# bt-nxuk
title: 'Footer-Hint bei offenem Filter-Menü: tab-Label kontextfalsch'
status: todo
type: task
priority: low
created_at: 2026-07-17T08:44:12Z
updated_at: 2026-07-17T08:44:12Z
parent: bt-5uzr
---

Reviewer-Finding B04 aus bt-2p9m-Review (2026-07-17, non-blocking): Bei offenem Filter-Menü rendert der Footer-Kontext-Hint weiterhin das wörtliche `keys.FocusIn`/`FocusOut`-Help („tab focus in · shift+tab focus out"), obwohl tab/shift+tab dort die Facet-Kategorie wechseln. Die Modal-interne Hint-Zeile („tab/shift+tab:category") ist korrekt — Fußzeile inkonsistent für Nutzer, die nur unten lesen.

Fix-Optionen (Reviewer): eigene `keybind.Binding`-Instanzen mit kontextspezifischem Help-Text nur für `filterMenuLocalBindings`, ODER `contextualLocalHint` überschreibt Label bei `m.filterOpen`. Drift-Guards (`TestNoDuplicateBindingBetweenGlobalAndAnyLocalHintList`, `TestHelpGroupsCoverEveryBindingExactlyOnce`) beachten.

Akzeptanz:
- [ ] Footer-Hint zeigt bei offenem Filter-Menü kategoriewechsel-korrektes Label
- [ ] Drift-Guard-Tests grün
- [ ] Test-Suite grün
