---
# bt-e4a6
title: E5 T3 — Yank y (Bean-/Epic-Kontext + Review-Stand, OSC52+nativ)
status: todo
type: task
created_at: 2026-07-15T09:04:24Z
updated_at: 2026-07-15T09:04:24Z
parent: bt-5h4d
---

Ziel: Yank `y` (OSC52 + nativ, Port devd `internal/clip` VERBATIM). Design
decision b (Deviation von devd, begründet): EIN view-agnostischer
`beanContext(idx, b)`-Markdown-Generator (ID/Titel/Status/Type/Priority/
Tags/Parent/Blocking/BlockedBy aufgelöst + Body, PLUS Children-Tabelle wenn
vorhanden -- deckt sowohl "Issue" als auch "Epic/Milestone" ab, da beans
keine devd-Dreiteilung Milestone/Sprint/Issue kennt, design-spec §4) statt
devds drei separaten milestoneClip/sprintClip/backlogClip-Funktionen. `y`
wirkt IMMER auf `m.focusedBean()` (Tree/Backlog identisch) -- EINZIGE
View-lokale Ausnahme: Review-Cockpit (design-spec §7 "Review-Stand
yanken") bekommt eine eigene `reviewStandMarkdown(idx)` (Gruppen-Tabelle
to-review je Epic + Rework-Sektion, spiegelt reviewQueue/reviewGroup aus
E4 T3, bt-hxyo).

Plan: docs/plans/v1-port/epic-E5-plan.md »Task 3«.

## Akzeptanz
- [ ] internal/clip/clip.go (NEU, Package): Copy(s string) error -- Port
      devd VERBATIM (OSC52 + tmux/screen-Passthrough + best-effort nativ
      pbcopy/wl-copy/xclip/xsel), `command go get
      github.com/aymanbagabas/go-osc52/v2` promoten (bereits indirect in
      go.mod/go.sum -- kein neues Supply-Chain-Risiko)
- [ ] internal/tui/context.go (NEU): beanContext(idx *data.Index, b
      *data.Bean) string -- Header + optionale Children-Tabelle (ID|Type|
      Status|Prio|Title, direkte Kinder, sortBeans-Reihenfolge)
- [ ] internal/tui/view_review_cockpit.go (oder eigene review_context.go):
      reviewStandMarkdown(idx *data.Index) string -- Kopf + je reviewGroup
      eine Tabelle (to-review) + Rework-Sektion, reuse reviewQueue/
      reviewRework (KEIN neuer Gruppierungs-Code)
- [ ] internal/tui/update.go keyNodeAction (oder keyTree/keyBacklog direkt):
      `y` (keys.Yank) -- clip.Copy(beanContext(...)) im Tree/Backlog,
      Erfolg -> showToast(toastInfo, "Kopiert", ...) (US-11: "Toast
      bestätigt"), Fehler -> showToast(toastWarn, ...)
- [ ] internal/tui/view_review_cockpit.go keyReviewCockpit: `y`-Override
      (Review-Stand statt Einzel-Bean, design-spec §7)
- [ ] context_test.go: TestBeanContextLeafHasNoChildrenTable,
      TestBeanContextParentHasChildrenTable, TestBeanContextResolvesParent
      TitleAndRelations, TestReviewStandMarkdownGroupsByEpic,
      TestYankShowsConfirmationToast (US-11), TestYankOnOrphanRootNoop
- [ ] `command go test ./... -short` grün, gofmt/vet leer
- [ ] Commit `feat(tui): Yank y (OSC52+nativ, Bean-/Epic-Kontext + Review-Stand)`
