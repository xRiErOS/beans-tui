---
# bt-ce7i
title: Fremde .beans-Aenderungen aus Experiment-Commit d4a5367 auf main bringen
status: todo
type: task
priority: high
created_at: 2026-07-20T07:32:50Z
updated_at: 2026-07-20T07:32:50Z
parent: bt-vy1q
---

**Fehler des Agenten, 2026-07-20.** Beim Anlegen von Bean `bt-1o4g` wurde `git add .beans/bt-*.md` (Glob) benutzt. Der Commit `d4a5367` (Message: "add bean for box-form field navigation (N8)") enthaelt dadurch **36 Dateien statt 1** — ~35 fremde, vor dieser Session bereits geaenderte bean-Dateien (tui-spec-Slices `bt-3cpw`/`bt-3km1`/`bt-57ka`/`bt-8o8v`/`bt-cuv7`/`bt-ikwb`/`bt-j3gd`/`bt-k7np`/`bt-kl0q`/`bt-nehl`/`bt-nwe7`/`bt-pz*`/`bt-vmea`/… sowie E-Epics `bt-1coz`/`bt-362n`/`bt-395t`/`bt-5h4d`/…).

## Warum das zaehlt
Diese fremden Arbeits-States liegen jetzt **ausschliesslich auf `experiment/jira-style-ui`**. Wird der Spike verworfen statt gemerged, gehen sie fuer `main` verloren — Arbeit anderer Straenge haengt am Schicksal dieses Experiments.

Zusaetzlich wurde `bt-ze10` mitcommittet, waehrend ein Implementer-Agent daran arbeitete (kein Datenverlust, nur unsauber einsortiert).

## Bewusst NICHT getan
History-Rewrite (`reset`/`amend`) waehrend ein Agent lief — haette dessen Commit verwaisen lassen koennen. Nichts ist verloren, nur falsch einsortiert.

## Optionen (PO-Entscheidung offen)
- **A (empfohlen):** die ~35 fremden `.beans`-Aenderungen zusaetzlich auf `main` bringen (dort separat committen / cherry-pick des `.beans`-Anteils), damit sie unabhaengig vom Spike-Schicksal sind
- **B:** so lassen — nur vertretbar, wenn der Spike sicher gemerged wird
- **C:** Commit nachtraeglich aufteilen (History-Rewrite) — nur wenn kein Agent laeuft

## Lehre (gilt ab sofort)
In `.beans/` **nur explizite Einzelpfade** stagen, nie ein Glob — das Repo traegt dauerhaft fremde uncommittete bean-Aenderungen.

## Akzeptanz
- [ ] PO waehlt A/B/C
- [ ] Gewaehlte Option ausgefuehrt
- [ ] `git log main --oneline -- .beans/` belegt, dass die fremden Aenderungen dort ankommen (bei A/C)
