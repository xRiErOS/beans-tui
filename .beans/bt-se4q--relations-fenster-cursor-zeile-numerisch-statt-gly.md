---
# bt-se4q
title: 'Relations-Fenster: Cursor-Zeile numerisch statt Glyph-Rescan'
status: todo
type: bug
priority: normal
created_at: 2026-07-16T22:21:31Z
updated_at: 2026-07-16T22:21:31Z
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
