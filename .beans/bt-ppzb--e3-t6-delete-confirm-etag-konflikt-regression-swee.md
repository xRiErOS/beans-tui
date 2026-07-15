---
# bt-ppzb
title: E3 T6 — Delete-Confirm + ETag-Konflikt-Regression-Sweep
status: todo
type: task
priority: high
created_at: 2026-07-15T00:28:07Z
updated_at: 2026-07-15T00:28:07Z
parent: bt-gzcu
blocked_by:
    - bt-sl45
---

Ziel: Delete-Confirm (`d`) mit Kinder-Count-Preview + ETag-Konflikt-Regression-
Sweep (Cross-Cutting-Verifikation über ALLE Mutations-Stellen aus Task 1-6, nicht
nur diese Task).

Kinder-Count-Preview ist SYNCHRON (idx.Children[id] ist schon im Speicher, KEIN
async Load wie devd box_confirm_delete.go loadDeletePreview). Semantik-Abweichung
von devd bewusst dokumentieren: `beans delete` kaskadiert NICHT (mutations.go
Delete-Doc-Kommentar: "--json skips ... any reference/child warnings", KEINE
Cascade-Delete-Funktion vorhanden) -- Kinder werden beim Löschen des Parents
NICHT mitgelöscht, sondern zu Waisen (dangling Parent), die die bestehende
"(verwaist)"-Root-Mechanik bereits korrekt rendert. Die Preview ist also eine
WARNUNG ("N Kinder werden verwaist"), kein "wird mitgelöscht"-Cascade-Count wie
bei devd.

ETag-Konflikt-Sweep: für JEDE in Task 1-6 gebaute Mutations-Stelle (Status/Type/
Prio-Menü, Tag-Picker, Parent-Picker, Blocking-Picker, Create, Title-Edit, Body-
$EDITOR, Delete) ein Regressionstest, der einen veralteten ETag simuliert (Reload
zwischen Öffnen und Submit) und verifiziert: ErrConflict -> Statuszeile +
unconditional Reload (Task 1s applyMutationResult), NIE ein Absturz/Silent-Drop.

Plan: docs/plans/v1-port/epic-E3-plan.md »Task 6«.

## Akzeptanz
- [ ] box_confirm_delete.go: deleteBox zeigt Typ+Titel+Kinder-Count (idx.Children,
      synchron, KEIN Loading-State), Rot=destruktiv (Port modalBox theme.Red wie
      quitBox mit theme.Mauve)
- [ ] enter löscht (Client.Delete, KEIN --if-match auf delete -- CLI-Signatur hat
      kein --if-match für delete, siehe mutations.go Delete()), esc/n bricht ab
- [ ] Nach Löschen: Cursor-Klemmung wie applyLoaded's bestehende oldPos-Fallback-
      Logik (E1 Task 8/E2) -- KEINE neue Cursor-Logik, Wiederverwendung verifiziert
- [ ] ETag-Konflikt-Regressionstest je Mutations-Stelle aus Task 1-6 (Tabelle im
      Plan, mind. 8 Fälle), alle grün
- [ ] go test ./... -count=1 (2x) grün, gofmt/vet leer, go build -o bin/bt . ok
