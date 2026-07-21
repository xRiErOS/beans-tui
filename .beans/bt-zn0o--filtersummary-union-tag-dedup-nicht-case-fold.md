---
# bt-zn0o
title: 'filterSummary-Union: Tag-Dedup nicht case-fold'
status: completed
type: task
priority: low
created_at: 2026-07-17T11:58:00Z
updated_at: 2026-07-18T15:05:07Z
parent: bt-5uzr
---

Reviewer-Finding I01 aus bt-2kfl-Review (2026-07-17, low, kosmetisch): Die UNION-Anzeige in `filterSummary` dedupliziert Tags nach exaktem Map-Key, nicht case-fold. Menü-Filter `Bad_Tag` + getipptes `tag:Bad_Tag` (Werte werden lowercased gespeichert) → Kopf zeigt `Tags:Bad_Tag,bad_tag` — zwei Einträge für denselben Tag. Matching selbst korrekt (EqualFold), nur Anzeige betroffen.

Fix: union()-Merge in filterSummary case-insensitiv abgleichen. Test mit gemischt-case Tag (Fixture-Präzedenz: `Bad_Tag` in bt-49hh-Tests).

Akzeptanz:
- [ ] Gleicher Tag via Menü+Präfix in unterschiedlicher Schreibung → EIN Kopfzeilen-Eintrag
- [ ] Test-Suite grün
