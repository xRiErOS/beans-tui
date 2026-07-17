---
# bt-se4q
title: 'Relations-Fenster: Cursor-Zeile numerisch statt Glyph-Rescan'
status: completed
type: bug
priority: normal
created_at: 2026-07-16T22:21:31Z
updated_at: 2026-07-17T08:22:00Z
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

## Summary

Root cause (bt-b0w0-Review, B01): `activeRelationLine` (view_browse_repo.go)
centered the RELATIONS scroll window by rescanning the ALREADY-RENDERED body
for the first `strings.Contains(l, "▶")` line -- unanchored, first match
wins. A relation TITLE containing "▶"/"▷" rendered before the real active
row would win the scan and center the window on the wrong line.

Fix: `relationsSectionBody` (view_detail_bean.go) already threads a running
row counter (`gi`) while it builds the groups (Parent/Children/Blocking/
Blocked By) -- it now ALSO accumulates the DISPLAY-LINE offset of each row
(accounting for subheader lines and any hangingIndentWrap continuation
lines) and returns the active row's display-line index as a 3rd return
value. `resolveSorted`/`beanListRow` grew a matching `activeOffset` return
(the offset of the active row within their own emitted lines, -1 if absent).
`accordionSection` grew a new `activeLine int` field (RELATIONS-only, zero
for the other 3 sections); `beanSections` populates it from
`relationsSectionBody`'s 3rd value. `windowRelationsSection`
(view_browse_repo.go) now takes `activeLine int` as an explicit 3rd
parameter instead of calling the removed `activeRelationLine` on the
rendered body -- no string-rescan of rendered output remains anywhere in
this path.

## Test-Output (RED -> GREEN)

RED (added `TestRelationsSectionBodyActiveLineIgnoresGlyphEmbeddedInTitle`,
which calls `relationsSectionBody` expecting 3 return values -- fails to
even build against the pre-fix 2-value signature, proving the numeric
contract does not exist yet):

    internal/tui/view_detail_bean_test.go:882:30: assignment mismatch: 3 variables but relationsSectionBody returns 2 values
    FAIL    beans-tui/internal/tui [build failed]

GREEN (after implementing the 3rd return value + full threading):

    === RUN   TestRelationsSectionBodyActiveLineIgnoresGlyphEmbeddedInTitle
    --- PASS: TestRelationsSectionBodyActiveLineIgnoresGlyphEmbeddedInTitle (0.00s)

The test builds a Children group where an EARLIER row's TITLE contains "▶"
(rendered as an inactive ▷ row) and asserts `activeLine` still points to the
LATER, actually-active row's own marker line -- and asserts the earlier
"trap" line exists and precedes it, so the test cannot pass vacuously.

Also added `TestWindowRelationsSectionIgnoresGlyphInsideUnrelatedLine`
(same trap pattern one layer up, at `windowRelationsSection`'s new 3-arg
signature) and updated the two pre-existing `windowRelationsSection` tests
(`TestWindowRelationsSectionWindowsAroundActiveLine`,
`TestWindowRelationsSectionDefaultsTopWhenNoActiveMarker`) to pass the
numeric index explicitly instead of encoding "▶"/"▷" into the synthetic
body (the old glyph-rescan contract no longer exists to test).

Full suite (`go test ./... -count=1`), all green:

    ok  	beans-tui	0.965s (cmd)
    ok  	beans-tui/internal/config	0.331s
    ok  	beans-tui/internal/data	3.435s
    ok  	beans-tui/internal/theme	0.600s
    ok  	beans-tui/internal/tui	155.261s

Total wall time ~2m36s. `go vet ./...` clean, `gofmt -l` clean on all
touched files.

## I01/I02-Bewertung

Both evaluated, both Won't-fix (no obligation to fix either, per plan):

- **I01** (a non-active wrapped row can be cut at the window's edge with no
  continuation hint): `windowStart`/`scrollView` operate on raw display
  lines, not row boundaries -- this is a pre-existing property of the
  windowing primitives themselves (NB-2/bt-b0w0), not something this bean's
  fix introduced or worsened. A real fix needs row-boundary-aware windowing
  (knowing where each row's own line span starts/ends, not just a flat line
  count) threaded through the same mechanism this bean just built -- a
  bigger structural change than a numeric-index bugfix warrants. No PO
  complaint on record; cosmetic-only, and only reachable with long titles at
  an already-constrained pane height.
- **I02** (`winH==1`: the group subheader, e.g. "Children", drops out of the
  window even though the top position is mathematically correct):
  confirmed live in the tmux smoke below -- at `winH==1` the window shows
  ONLY the active row (`▶ i E bt-blsy ...`) with an `↑`/`↓` indicator, no
  "Children" label. This is an inherent trade-off of a 1-line budget:
  showing the subheader instead would mean NEVER showing the actual cursor
  row alone, a worse outcome. No functional loss -- every relation stays
  fully reachable via ↑/↓, the data is never hidden, only the group label at
  this one extreme pane-height edge case.

## Smoke

tmux session `btse4q30406` (unique per-task name, `bin/bt` built from THIS
worktree), bean `bt-apmy` (milestone, 10 Children -- E1..E10 epics/tasks).

- 220x50: `3` opens RELATIONS, all 10 children visible simultaneously
  (fits budget), `▶` marker on `bt-blsy` (fieldCursor default 0).
- Resized to 220x20 (forces windowing, `avail`/`winH` shrink to 1): window
  correctly shows ONLY the active row + `L n-n/10` indicator (I02 above).
  `Right` (enter field level) + 3x `Down`:

      ▶ i E bt-tfqi E4 Command-Center & Review-Cockpit
      ↑ L 5–5/10 ↓

  matches fieldCursor=3 (Parent absent here, Children start at global index
  0) -- window followed the cursor exactly 1 line per keypress, both
  indicators present (not at either end). 6 more `Down` presses reached the
  LAST child:

      ▶ t E bt-heg9 E7 — PO-Feedback R1: Detail-UX + Typ/Status/Prio-Glyphen
      ↑ L 10–10/10

  only the `↑` indicator shown (no more entries below) -- correct
  end-of-list behavior. Resized back to 220x45 (all 10 fit again): `▶`
  marker correctly still on `bt-heg9` (the same cursor position, unchanged
  across the resize), NB-2/bt-b0w0's auto-scroll + indicator acceptance
  intact end-to-end.

`git status --short .beans/` clean after the smoke session (`q` to quit,
tmux session killed).

## Deviations/ERRATA

- Plan cites `activeRelationLine` at `view_browse_repo.go:721-728` and
  `relationsSectionBody` at `view_detail_bean.go:399-460` -- Ist-Code had
  drifted slightly (bt-81f0's Toast-Vereinheitlichung landed first per the
  epic-E12-plan's own Item 2-before-Item 4 ordering), actual line numbers
  differed but the functions/logic described matched exactly; no other
  drift found.
- Design-Rahmen said "Struct-Feld an `accordionSection` ODER Parallel-
  Return, deine Wahl" -- chose the struct-field (`accordionSection.
  activeLine`), since `beanSections` already threads all 4 sections through
  one `[]accordionSection` slice and `renderAccordionPane` already indexes
  into it (`secs[relationsSectionIdx]`) to rewrite `.body` in place for the
  windowing call -- adding `.activeLine` there needed no new plumbing
  layer, whereas a parallel return would have required a second slice
  threaded alongside `secs` through the same call chain for no added
  clarity.
