package tui

// view_detail_bean.go — the 4 fixed accordion sections for a bean (design-
// spec §6 V4, §15 PF-1/PF-3/PF-4/PF-7): META / BODY (glamour) / RELATIONS /
// HISTORY. Always all 4 (unlike devd's content-gated issueSections) -- there
// is no reason to hide an empty section; digit jump 1..4 stays meaningful for
// every bean. Section titles are uppercase English (PF-7, E7 T3 deliberately
// left these 4 to this task -- bt-w9o8 "Notes for T4"). META additionally
// carries a NEW, non-collapsible (PF-1) 6-entry field list (PF-4) instead of
// its old 4-line status/type/priority/tags summary -- Tags is no longer part
// of Meta at all under the new fixed field set.

import (
	"fmt"
	"strings"
	"time"

	"beans-tui/internal/data"
	"beans-tui/internal/theme"
)

// beanSectionCount is the Single Source for the fixed section count (T6,
// design-spec.md §15 PF-2: replaces two independently hardcoded '4' literals
// in update.go::keyDetailFocus and the former view_review_cockpit.go --
// removed by PF-14 -- with one shared constant).
const beanSectionCount = 4

// Section titles (PF-7, uppercase English).
const (
	sectionTitleMeta      = "META"
	sectionTitleBody      = "BODY"
	sectionTitleRelations = "RELATIONS"
	sectionTitleHistory   = "HISTORY"
)

// beanSections builds the 4 fixed sections of the Detail-Accordion for b.
// bodyW is the inner body width (renderAccordion's box, already accounting
// for its own PaddingLeft/border -- callers pass the same width they'd hand
// renderAccordion). focused/activeIdx/fieldIdx are the SAME render-state
// triple renderAccordion itself takes (E7 T4, bean bt-kyj5): META's body
// needs to know whether IT is the currently active section to place its own
// ▶ field marker (PF-4) -- previously META's body was static text, agnostic
// of focus state.
func beanSections(idx *data.Index, b *data.Bean, bodyW int, focused bool, activeIdx, fieldIdx int) []accordionSection {
	var secs []accordionSection
	secs = append(secs, accordionSection{
		title:  sectionTitleMeta,
		body:   metaSectionBody(b, bodyW, focused && activeIdx == 0, fieldIdx),
		fields: metaFields(b),
	})
	secs = append(secs, accordionSection{title: sectionTitleBody, body: bodySectionBody(b, bodyW)})
	rel, fields := relationsSectionBody(idx, b, bodyW)
	secs = append(secs, accordionSection{title: sectionTitleRelations, body: rel, fields: fields})
	secs = append(secs, accordionSection{title: sectionTitleHistory, body: historieSectionBody(b, bodyW)})
	return secs
}

// metaFieldLabels are the 6 fixed row labels of the Meta field list (PF-4),
// left-padded to a common width (12) so every value column aligns -- exact
// spacing verified against design-spec.md §15's mockup (every label pads to
// 12 cells: "title:      ", "created_at: ", etc.).
var metaFieldLabels = [...]string{"title:", "status:", "type:", "priority:", "created_at:", "updated_at:"}

// metaFields builds the 6 kind-tagged Meta field entries (PF-4, design-
// spec.md §15): title / status / type / priority / created_at / updated_at,
// in that fixed order. status/type/priority reuse T2's GLYPH-producing theme
// helpers (StatusIcon/TypeIcon/Priority) -- NOT the word, unlike the
// Kopfblock (detailHeaderBlock, below), which uses the word for type/status.
// beanID is always "" (Meta fields are never Beziehungen-style jump
// targets); kind drives T6's future enter-dispatch (status/type/priority ->
// seeded Value-Menu, title -> Title-Edit-Form, readonly -> No-Op).
func metaFields(b *data.Bean) []relationField {
	priority := b.Priority
	if priority == "" {
		priority = "normal"
	}
	return []relationField{
		{kind: "title", label: b.Title},
		{kind: "status", label: theme.StatusIcon(b.Status)},
		{kind: "type", label: theme.TypeIcon(b.Type)},
		{kind: "priority", label: theme.Priority(priority)},
		{kind: "readonly", label: fmtTime(b.CreatedAt)},
		{kind: "readonly", label: fmtTime(b.UpdatedAt)},
	}
}

// metaSectionBody renders the 6-entry Meta field list (PF-4): a ▷/▶ cursor
// column (PF-12: ALWAYS one or the other, never omitted -- the gutter is
// inherently stable by construction here, unlike the two accordion.go call
// sites PF-12 had to retrofit) followed by the field's label (padded to
// metaFieldLabels' common width) and its pre-rendered value (metaFields).
// ▶ (Accent) marks the row only when active is true AND fieldIdx matches
// that row's index -- mirrors render_shared.go's renderPane list-cursor
// convention (only the marker glyph is accent-tinted, not the whole row).
// Wrapped via wrapText like the section body always has been (B01, E2-T1-
// quality-review, MANDATORY carry-over) -- an overlong title must wrap
// instead of overflowing the Detail pane's bordered width.
func metaSectionBody(b *data.Bean, bodyW int, active bool, fieldIdx int) string {
	fields := metaFields(b)
	var lines []string
	for i, f := range fields {
		marker := "▷ "
		if active && fieldIdx == i {
			marker = theme.Accent.Render("▶ ")
		}
		label := theme.Muted.Render(fmt.Sprintf("%-12s", metaFieldLabels[i]))
		lines = append(lines, marker+label+f.label)
	}
	return wrapText(strings.Join(lines, "\n"), bodyW)
}

// detailHeaderBlock renders the NEW Detail-Pane Kopfblock (PF-3+PF-4,
// verschmolzen laut PO-Antwort Q01, design-spec.md §15): bean-id / NAME /
// blank / "type: X    status: Y    prio: Z" / trailing blank, above the
// Accordion. type/status render as the colored WORD (theme.TypeStyle/
// StatusStyle applied to b.Type/b.Status) -- prio renders as the colored
// GLYPH (theme.Priority; PF-6 removed the word-producing counterpart, there
// is no PriorityStyle to apply to a word). Not the App-Chrome header (PO-
// Antwort Q01 explicit: "Kein separater App-Header-Umbau") -- Detail-Pane
// only.
func detailHeaderBlock(b *data.Bean, w int) string {
	priority := b.Priority
	if priority == "" {
		priority = "normal"
	}
	typeStatusPrio := theme.Muted.Render("type: ") + theme.TypeStyle(b.Type).Render(b.Type) +
		theme.Muted.Render("    status: ") + theme.StatusStyle(b.Status).Render(b.Status) +
		theme.Muted.Render("    prio: ") + theme.Priority(priority)
	lines := []string{
		truncate(theme.Key.Render(b.ID), w),
		truncate(theme.Header.Render(b.Title), w),
		"",
		truncate(typeStatusPrio, w),
		"",
	}
	return strings.Join(lines, "\n")
}

// bodySectionBody renders b.Body via glowRender (glamour), with a muted
// placeholder for an empty body -- digit-jump 2 must always show something.
func bodySectionBody(b *data.Bean, bodyW int) string {
	if strings.TrimSpace(b.Body) == "" {
		return theme.Dim.Render("(no body)")
	}
	return glowRender(b.Body, bodyW)
}

// relationRow renders one resolved relation as a status-icon+type-icon+ID+
// title row (mirrors treeRowText's glyph order, view_browse_repo.go).
func relationRow(rel *data.Bean) string {
	return theme.StatusIcon(rel.Status) + " " + theme.TypeIcon(rel.Type) + " " + theme.Key.Render(rel.ID) + " " + rel.Title
}

// resolveSorted resolves a slice of bean IDs against idx.ByID, canonically
// sorting the resolved beans (data.SortBeans -- I03) and appending unresolved
// IDs (dangling references, beans-legal) as their own relationField with
// beanID == "" (not jumpable). Returns the rendered body lines plus the
// relationFields in the same order (parallel to the rendered rows).
func resolveSorted(idx *data.Index, ids []string) (lines []string, fields []relationField) {
	var resolved []*data.Bean
	var unresolved []string
	for _, id := range ids {
		if rel, ok := idx.ByID[id]; ok {
			resolved = append(resolved, rel)
		} else {
			unresolved = append(unresolved, id)
		}
	}
	data.SortBeans(resolved)
	for _, rel := range resolved {
		lines = append(lines, relationRow(rel))
		fields = append(fields, relationField{beanID: rel.ID, label: theme.Key.Render(rel.ID) + " " + rel.Title})
	}
	for _, id := range unresolved {
		label := theme.Dim.Render("(unresolved: " + id + ")")
		lines = append(lines, label)
		fields = append(fields, relationField{beanID: "", label: label})
	}
	return lines, fields
}

// beanListRow renders an already-resolved bean list (e.g. idx.Children,
// pre-sorted -- no dangling references possible there) the same way
// resolveSorted renders its resolved half.
func beanListRow(beans []*data.Bean) (lines []string, fields []relationField) {
	for _, rel := range beans {
		lines = append(lines, relationRow(rel))
		fields = append(fields, relationField{beanID: rel.ID, label: theme.Key.Render(rel.ID) + " " + rel.Title})
	}
	return lines, fields
}

// relationsSectionBody builds the Beziehungen section body + its jump-only
// relationFields: Parent (0 or 1, via idx.ByID) / Children (idx.Children[b.ID],
// already sorted) / Blocking / Blocked By (each resolved via idx.ByID, then
// data.SortBeans -- I03), grouped under muted sub-headers. Groups with no
// entries are omitted entirely.
func relationsSectionBody(idx *data.Index, b *data.Bean, bodyW int) (string, []relationField) {
	var groups []string
	var fields []relationField

	appendGroup := func(title string, lines []string, fs []relationField) {
		if len(lines) == 0 {
			return
		}
		groups = append(groups, theme.Muted.Render(title)+"\n"+strings.Join(lines, "\n"))
		fields = append(fields, fs...)
	}

	if b.Parent != "" {
		if parent, ok := idx.ByID[b.Parent]; ok {
			appendGroup("Parent", []string{relationRow(parent)}, []relationField{{beanID: parent.ID, label: theme.Key.Render(parent.ID) + " " + parent.Title}})
		} else {
			label := theme.Dim.Render("(unresolved: " + b.Parent + ")")
			appendGroup("Parent", []string{label}, []relationField{{beanID: "", label: label}})
		}
	}

	if children := idx.Children[b.ID]; len(children) > 0 {
		lines, fs := beanListRow(children)
		appendGroup("Children", lines, fs)
	}

	if len(b.Blocking) > 0 {
		lines, fs := resolveSorted(idx, b.Blocking)
		appendGroup("Blocking", lines, fs)
	}

	if len(b.BlockedBy) > 0 {
		lines, fs := resolveSorted(idx, b.BlockedBy)
		appendGroup("Blocked By", lines, fs)
	}

	if len(groups) == 0 {
		return theme.Dim.Render("(no relations)"), nil
	}
	return wrapText(strings.Join(groups, "\n\n"), bodyW), fields
}

// fmtTime formats a nullable timestamp for display: a muted placeholder for
// nil instead of Go's zero-time string, "2006-01-02 15:04" otherwise. Shared
// by historieSectionBody (Created/Updated) and metaFields (created_at/
// updated_at, PF-4, E7 T4) -- extracted from historieSectionBody's former
// local closure so both call sites stay byte-identical, not two
// re-implementations that could drift.
func fmtTime(t *time.Time) string {
	if t == nil {
		return theme.Dim.Render("(unknown)")
	}
	return t.Format("2006-01-02 15:04")
}

// historieSectionBody renders CreatedAt/UpdatedAt/ETag. nil timestamps show a
// muted placeholder instead of Go's zero-time string. B01 (E2-T1-quality-
// review, MANDATORY carry-over into Task 2): wrapped via wrapText -- a long
// ETag must wrap instead of overflowing the Detail pane's bordered width.
func historieSectionBody(b *data.Bean, bodyW int) string {
	var lines []string
	lines = append(lines, theme.Muted.Render("Created: ")+fmtTime(b.CreatedAt))
	lines = append(lines, theme.Muted.Render("Updated: ")+fmtTime(b.UpdatedAt))
	etag := b.ETag
	if etag == "" {
		etag = theme.Dim.Render("(none)")
	}
	lines = append(lines, theme.Muted.Render("ETag: ")+etag)
	return wrapText(strings.Join(lines, "\n"), bodyW)
}
