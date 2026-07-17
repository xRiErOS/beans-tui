---
# bt-se4q
title: 'Relations-Fenster: Cursor-Zeile numerisch statt Glyph-Rescan'
status: todo
type: bug
priority: normal
created_at: 2026-07-16T22:21:31Z
updated_at: 2026-07-17T06:46:09Z
parent: bt-tct9
---

Follow-up aus bt-b0w0/NB-2-Review (2026-07-17, APPROVED mit latentem Befund).

B01 (medium, latent): `activeRelationLine` (view_browse_repo.go:640-647) sucht
die aktive Zeile per `strings.Contains(l, "▶")` — unverankert, erste Fundstelle.
Reviewer-Beleg (temporärer Testfall): eine Zeile mit `▷ … Titel mit ▶ …` VOR der
echten aktiven Zeile zentriert das Fenster falsch; im Extremfall schiebt es den
echten Cursor aus dem Sichtbereich. Aktuell NICHT scharf: kein Bean-Titel im
Repo enthält ▶/▷ (grep-verifiziert).

Fix-Skizze (Reviewer): Cursor-Zeilen-Index numerisch aus der Body-Konstruktion
(`relationsSectionBody`) zurückgeben statt Glyph-Rescan des gerenderten Strings.

Sekundär mitnehmen (low, gleiche Funktion):
- I01: Fenster zählt Display-Zeilen (post-Wrap) — nicht-aktive gewrappte Zeile
  kann am Fensterrand ohne Fortsetzung/Hinweis abgeschnitten werden.
- I02: winH==1: "Children"-Subheader fällt aus dem Fenster, ↑-Indikator trotz
  gefühlter Top-Position (mathematisch korrekt, verwirrend).

Akzeptanz:
- [ ] activeRelationLine durch numerischen Index ersetzt (kein Glyph-Rescan)
- [ ] Regressionstest: Relations-Titel mit ▶/▷-Zeichen zentriert korrekt
- [ ] I01/I02 bewertet — fixen oder begründet als Won't-fix dokumentieren


## Plan-Konkretisierung E12 (2026-07-17)

Plan: `docs/plans/v1-port/epic-E12-plan.md` §„Item 4: Relations-Fenster —
numerischer Cursor-Index statt Glyph-Rescan". Reihenfolge: Rang 4 (Cluster
mit `bt-gdkx`, gemeinsame Datei `view_detail_bean.go`, unterschiedliche
Funktionen, kein hartes blocked_by).

**Root Cause (bereits im Reviewer-Finding lokalisiert):**
`activeRelationLine` (`view_browse_repo.go:721-728`) sucht die aktive Zeile
per `strings.Contains(l, "▶")` im BEREITS gerenderten String —
unverankert. Die Information ist zum Konstruktionszeitpunkt in
`relationsSectionBody` (`view_detail_bean.go:399-460`) bereits als exakter
Integer bekannt: laufender Zeilenzähler `gi` (Parent=0, Children=1..n,
Blocking=n+1..m, Blocked By=m+1..k).

**Vorgehen:** `relationsSectionBody` um dritten Rückgabewert erweitern
(z. B. `activeLine int`), der die Display-Zeilen-Position der aktiven Zeile
mitzählt (inkl. Subheader-Zeilen und ggf. mehrzeiliger
`hangingIndentWrap`-Ausgabe). Läuft durch `beanSections`
(`view_detail_bean.go:71-89`) bis zu `renderAccordionPane`/
`windowRelationsSection` (`view_browse_repo.go:629-668`), wo
`activeRelationLine(lines)` durch den durchgereichten Wert ersetzt wird.
Implementer wählt Durchreichungs-Mechanik (Struct-Feld vs. Parallel-Slice).

Sekundär (Reviewer I01/I02, gleiche Funktion): I01 (Fenster zählt
Display-Zeilen post-Wrap, gewrappte Zeile kann ohne Fortsetzungshinweis
abgeschnitten werden), I02 (`winH==1`: "Children"-Subheader fällt trotz
korrekter Top-Position aus dem Fenster) — je bewerten: fixen ODER begründet
als Won't-fix dokumentieren.

**Akzeptanz:**
- [ ] Glyph-Rescan durch numerischen, aus der Body-Konstruktion
      durchgereichten Index ersetzt
- [ ] Regressionstest: Relations-Eintrag mit ▶/▷ IM TITEL zentriert das
      Fenster weiterhin korrekt
- [ ] I01/I02 bewertet, Ergebnis dokumentiert (Fix oder Won't-fix +
      Begründung)
- [ ] Bestehende NB-2/`bt-b0w0`-Akzeptanz bleibt unverändert grün
- [ ] tmux-Smoke mit relations-reichem Bean (`bt-apmy`)
