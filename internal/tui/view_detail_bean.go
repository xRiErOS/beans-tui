package tui

// view_detail_bean.go — the 4 fixed accordion sections for a bean (design-
// spec §6 V4): Meta / Body (glamour) / Beziehungen / Historie. Always all 4
// (unlike devd's content-gated issueSections) -- E2 has no edit-field
// concept yet (E3), so there's no reason to hide an empty section; digit
// jump 1..4 stays meaningful for every bean.

import (
	"strings"
	"time"

	"beans-tui/internal/data"
	"beans-tui/internal/theme"
)

// beanSections builds the 4 fixed sections of the Detail-Accordion for b.
// bodyW is the inner body width (renderAccordion's box, already accounting
// for its own PaddingLeft/border -- callers pass the same width they'd hand
// renderAccordion).
func beanSections(idx *data.Index, b *data.Bean, bodyW int) []accordionSection {
	var secs []accordionSection
	secs = append(secs, accordionSection{title: "Meta", body: metaSectionBody(b, bodyW)})
	secs = append(secs, accordionSection{title: "Body", body: bodySectionBody(b, bodyW)})
	rel, fields := relationsSectionBody(idx, b, bodyW)
	secs = append(secs, accordionSection{title: "Beziehungen", body: rel, fields: fields})
	secs = append(secs, accordionSection{title: "Historie", body: historieSectionBody(b, bodyW)})
	return secs
}

// metaSectionBody renders Status/Type/Priority/Tags via the existing themed
// helpers (theme.StatusStyle/TypeStyle/Priority, render_shared.go's
// tagsInline) -- no new theme code, mirrors renderDetailPane's meta line
// (view_browse_repo.go) but as a full, labeled section. B01 (E2-T1-quality-
// review, MANDATORY carry-over into Task 2): wrapped via wrapText like
// relationsSectionBody already does -- a bean with many/long tags must wrap
// inside the Detail pane's bordered width instead of overflowing it.
func metaSectionBody(b *data.Bean, bodyW int) string {
	var lines []string
	lines = append(lines, theme.Muted.Render("Status: ")+theme.StatusStyle(b.Status).Render(b.Status))
	lines = append(lines, theme.Muted.Render("Type: ")+theme.TypeStyle(b.Type).Render(b.Type))
	priority := b.Priority
	if priority == "" {
		priority = "normal"
	}
	lines = append(lines, theme.Muted.Render("Priority: ")+theme.Priority(priority))
	if t := tagsInline(b.Tags); t != "" {
		lines = append(lines, theme.Muted.Render("Tags: ")+t)
	} else {
		lines = append(lines, theme.Muted.Render("Tags: ")+theme.Dim.Render("(none)"))
	}
	return wrapText(strings.Join(lines, "\n"), bodyW)
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

// historieSectionBody renders CreatedAt/UpdatedAt/ETag. nil timestamps show a
// muted placeholder instead of Go's zero-time string. B01 (E2-T1-quality-
// review, MANDATORY carry-over into Task 2): wrapped via wrapText -- a long
// ETag must wrap instead of overflowing the Detail pane's bordered width.
func historieSectionBody(b *data.Bean, bodyW int) string {
	fmtTime := func(t *time.Time) string {
		if t == nil {
			return theme.Dim.Render("(unknown)")
		}
		return t.Format("2006-01-02 15:04")
	}
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
