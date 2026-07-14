---
# bt-tknb
title: 'T4 Datenlayer: Mutationen mit --if-match (ErrConflict)'
status: completed
type: task
priority: high
created_at: 2026-07-14T18:34:04Z
updated_at: 2026-07-14T19:23:00Z
parent: bt-blsy
blocked_by:
    - bt-xmrs
---

Plan: implementation-plan.md »E1 Task 4«.

## Akzeptanz
- [x] Create/SetStatus/SetPriority/SetType/SetTitle/AddTag/RemoveTag/SetParent/RemoveParent/AddBlockedBy/RemoveBlockedBy/AppendBody/Delete
- [x] Updates senden --if-match; typed ErrConflict bei Stale-ETag
- [x] Tests inkl. TestConflictOnStaleETag grün


## Übernommene Review-Findings aus T2-Quality-Review (PFLICHT in diesem Task)
- [x] B01: TestListErrorIncludesStderr muss stderr-Inhalt wirklich asserten (strings.Contains auf stabilen Substring)
- [x] I01: Fixture-bean mit tags/blocking/blocked_by in newTestRepo + Assertion der geparsten Slices (JSON-Vertrag als Regression-Guard)
- [x] I02: client.go List: Unmarshal-Fehler mit Kommando-Kontext wrappen
- [x] I03: client.go run: trailing ": " bei leerem stderr vermeiden
- [x] I04: discover.go: Fehlermeldung mit resolved-Pfad statt raw start-Arg

## Summary of Changes

- `internal/data/mutations.go` (neu): `CreateOpts` + `Client.Create`, generisches
  `Client.update(id, etag, args...)` als gemeinsame Basis für alle Setter
  (`SetStatus/SetPriority/SetType/SetTitle/AddTag/RemoveTag/SetParent/RemoveParent/
  AddBlockedBy/RemoveBlockedBy/AppendBody`), `Client.Delete` (`--force`, kein
  stdin-Pipe nötig — `beans delete` hat einen echten Force-Flag). Typed Sentinel
  `ErrConflict`, erkannt über stabilen stderr-Substring `"etag mismatch"`
  (siehe unten), gewrapped via `%w` für `errors.Is`.
- `internal/data/client.go` (I02/I03): `List()` wrapped jetzt Unmarshal-Fehler mit
  Kommando-Kontext; `run()` hängt `: <stderr>` nur noch an wenn stderr
  tatsächlich Inhalt hat (kein trailing `": "` mehr bei leerem stderr).
- `internal/data/discover.go` (I04): `FindRepo` merkt sich den einmal aufgelösten
  Absolutpfad (`resolved`) und meldet den in der Fehlermeldung statt des rohen
  `start`-Arguments (relevant wenn `start` relativ war, z.B. `"."`).
- `internal/data/client_test.go` (B01 + I01): `TestListErrorIncludesStderr`
  asserted jetzt `strings.Contains` auf `"no .beans directory found"`;
  `TestListReturnsAllBeansWithBody` bekam zusätzliche Assertions auf
  `epic.Tags`, `task.BlockedBy`, `milestone.Blocking` (JSON-Vertrag-Regression-Guard).
- `internal/data/testrepo_test.go` (I01 + Bugfix-Anpassung): Fixtures um
  `blocking`/`tags`/`blocked_by` erweitert — **tags bewusst auf `tt-epic`
  platziert, nicht `tt-task`** (siehe Concern zum beans-0.4.2-ETag-Bug unten).
- `internal/data/client_mut_test.go` (neu): `TestSetStatusRoundtrip`,
  `TestCreateReturnsNewBean`, `TestConflictOnStaleETag`,
  `TestAppendBodyAddsSection`, plus Bonus `TestDeleteRemovesBean`
  (Delete ist Teil der Pflicht-API, aber im Plan nicht explizit als
  TDD-Test verlangt — trotzdem getestet, da der Force-Flag-Pfad sonst
  unbelegt bliebe).
- Alle Tests laufen gegen die echte `beans`-Binary (0.4.2, `/opt/homebrew/bin/beans`),
  kein Mock. `command gofmt -l .` leer, `command go vet ./...` clean,
  `command go build -o bin/bt .` ok.

### Concern — beans 0.4.2 CLI-Bug: tags-Feld bricht ETag-Konsistenz bei Hand-Fixtures

Manuell isoliert (Scratch-Repos, siehe Session): Für eine **hand-geschriebene**
`.md`-Datei, deren Frontmatter ein `tags:`-Listenfeld enthält, liefert
`beans list`/`beans show` ein ANDERES ETag als das, was `beans update
--if-match` intern für den Konfliktcheck berechnet — bis die Datei einmal
von beans selbst neu geschrieben wurde (z.B. durch ein erfolgreiches
Update), danach sind beide Pfade wieder konsistent. `parent`/`blocking`/
`blocked_by` als Hand-Fixture-Felder zeigen dieses Verhalten NICHT — nur
`tags`. Betrifft nur hand-authored Fixtures (nicht reale, CLI-geschriebene
Beans), aber genau das ist unser Testfixture-Muster (`newTestRepo`).
Workaround hier: `tags` liegt auf `tt-epic` (nicht mutations-getestet),
`tt-task` (Ziel der SetStatus/Conflict/AppendBody-Roundtrips) bleibt frei
von hand-geschriebenem `tags`. Kein Fix im bt-Code nötig/möglich (Upstream-
Verhalten); zur Kenntnis für künftige Fixture-Arbeit dokumentiert.
