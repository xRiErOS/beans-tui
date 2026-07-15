package tui

// accordion.go — exclusive-open, digit-addressed detail sections (design-spec
// §6 V4). Port of devd accordion.go's render algebra (renderAccordion,
// fieldStrip), stripped of every edit-field/detailField-editor concept (E3
// scope) -- E2's "fields" are pure navigation targets inside the
// Beziehungen section (relationField), never edit targets.
//
// Port references (~/Obsidian/tools/DeveloperDashboard/apps/cli-go/internal/
// tui/): accordion.go:309-355 (renderAccordion), accordion.go:360-373
// (fieldStrip), editor.go:102-126 (glowRender).

import (
	"fmt"
	"strings"

	"beans-tui/internal/theme"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
)

// relationField is a navigation target inside a section's field list --
// EITHER a jump target (Relations section, kind == "") OR a Meta field (PF-4,
// E7 T4, bean bt-kyj5, design-spec.md §15): kind drives T6's future
// enter-dispatch (status/type/priority open the seeded Value-Menu, title
// opens the Title-Edit-Form, readonly is a No-Op for created_at/updated_at).
// beanID == "" marks either an unresolved/dangling relation reference (not
// jumpable) or any Meta field (Meta fields are never jump targets, beanID is
// always "" there).
type relationField struct {
	beanID string // target bean ID; "" for an unresolved/dangling reference (not jumpable) or any Meta field
	label  string // pre-rendered row text (theme colors already applied)
	kind   string // "" = jump (Relations, E2 behavior unchanged) | "status"|"type"|"priority" = Value-Menu seeded on group | "tags" = Tag-Picker (PF-15/D01, E8 Task 1, bean bt-e6q9) | "title" = Title-Edit-Form | "readonly" = No-Op (created_at/updated_at)
}

// accordionSection mirrors devd's shape minus the editor-specific parts.
// fields == nil for every section except Beziehungen (Meta/Body/Historie are
// pure display, no field-level navigation in E2).
type accordionSection struct {
	title  string
	body   string
	fields []relationField
}

// renderAccordion rendert die Sektions-Header (`> [n] Title`) und klappt die
// offene Sektion (1-basiert, exklusiv) als eingerückten Body darunter auf
// (port devd accordion.go:309-355). active/activeIdx/fieldIdx replace devd's
// detailFocusView{sec,field}. PF-1 (design-spec.md §15, E7 T4, bean bt-kyj5):
// section 1 (Meta) is never collapsible -- its body renders regardless of
// `open`, sections 2-4 stay exclusive-open. fieldStrip renders for ANY
// active section carrying fields EXCEPT Meta itself (`i != 0`): Meta's own
// ▷/▶ field cursor lives INSIDE metaSectionBody's per-row markers (PF-4), a
// second "Fields: ..." strip above it would be a duplicate cursor.
func renderAccordion(secs []accordionSection, open, w int, active bool, activeIdx, fieldIdx int) string {
	if len(secs) == 0 {
		return theme.Dim.Render("(no detail fields)")
	}
	headerStyle := lipgloss.NewStyle().Width(w)
	// DD2-186 (devd note, carried over): lipgloss.Width INCLUDES PaddingLeft
	// -> Width(w).PaddingLeft(2) keeps the effective content width == the
	// caller's own bodyW (w-2), indent stays 2.
	boxStyle := lipgloss.NewStyle().Width(w).PaddingLeft(2)

	var b strings.Builder
	for i, s := range secs {
		n := i + 1
		isOpen := n == open || n == 1         // PF-1: Meta (section 1) always shows its body
		activeSec := active && activeIdx == i // D08: focus is on this section
		marker := theme.Chevron.Render("> ") + theme.Key.Render(fmt.Sprintf("[%d]", n)) + " "
		var title, hint string
		if isOpen {
			title = theme.Header.Render(s.title)
			hint = theme.Muted.Render("  ▾")
		} else {
			title = theme.Muted.Render(s.title)
			hint = theme.Muted.Render("  ▸")
		}
		// PF-12 (design-spec.md §15, E7 T4): BOTH branches now reserve
		// exactly 1 gutter column (" " inactive, "▌" active) and truncate
		// content to w-1 -- previously only the active branch reserved a
		// column (truncate to w-1), the inactive branch used the FULL w with
		// no prefix, so a header's own title text shifted 1 column depending
		// on whether IT ITSELF was active (layout shift on selection).
		var header string
		if activeSec {
			// D08: active section = ▌ bar + whole line accent-tinted (own
			// per-cell colors stripped first, mirrors the tree cursor).
			header = theme.Accent.Render("▌" + truncate(ansi.Strip(marker+title+hint), w-1))
		} else {
			header = " " + truncate(marker+title+hint, w-1)
		}
		b.WriteString(headerStyle.Render(header) + "\n")
		if isOpen {
			// Field-level focus shows the field strip (which field a jump
			// would target) before the body -- skipped for Meta (i == 0):
			// its field cursor renders per-row inside the body itself
			// (metaSectionBody's ▷/▶ markers, PF-4), not as a strip above.
			if activeSec && i != 0 && len(s.fields) > 0 {
				b.WriteString(fieldStrip(s.fields, fieldIdx, w) + "\n")
			}
			b.WriteString(boxStyle.Render(s.body) + "\n")
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

// fieldStrip rendert die Beziehungen-Felder einer Section als kompakte Zeile
// mit dem aktiven Feld (D08-Balken + Accent), die übrigen muted (port devd
// accordion.go:360-373, detailField.label -> relationField.label). PF-12
// (design-spec.md §15, E7 T4): the inactive branch now reserves the same 1
// gutter column (" " prefix) the active branch's "▌" occupies -- previously
// bare (no prefix), so a field's own label shifted 1 column depending on
// whether IT ITSELF was active.
func fieldStrip(fields []relationField, active, w int) string {
	if len(fields) == 0 {
		return theme.Muted.Render("Fields: (none)")
	}
	parts := make([]string, len(fields))
	for i, f := range fields {
		if i == active {
			parts[i] = theme.Accent.Render("▌" + f.label)
		} else {
			parts[i] = " " + theme.Muted.Render(f.label)
		}
	}
	return truncate(theme.Muted.Render("Fields: ")+strings.Join(parts, "  "), w)
}

// glowRender rendert Markdown für die Body-Section via glamour. Fällt auf den
// rohen Text zurück, wenn glamour fehlschlägt. Im Ascii-Profil (Tests/kein
// TTY -> an dasselbe Profil gekoppelt wie die Goldens) wird der plain
// "notty"-Style genutzt, damit Golden-Snapshots ANSI-frei und stabil bleiben
// (port devd editor.go:102-126, verbatim).
//
// I02 (E2-T1-quality-review, decision made at Task 2 wiring time): this
// constructs a brand-new glamour.TermRenderer on EVERY call -- and, now that
// view_browse_repo.go's renderDetailPane wires beanSections into the live
// View() path, that means once per render tick while the Body section is
// open (every keypress, not just on bean-change). Decision: ACCEPT the cost
// rather than add a cache -- this mirrors devd's own upstream editor.go
// behavior 1:1 (same per-call construction there), bean bodies are typically
// a few KB of markdown (well under interactive-latency budgets), and E2 has
// no profiling evidence of this being a hot path. If profiling ever shows
// otherwise, the fix is a cache keyed by (bean.ID, bean.ETag, width) on
// model -- NOT implemented here (YAGNI until proven necessary).
//
// I03 (optional, E2-T1-quality-review): the two error-fallback paths below
// return the raw, UNWRAPPED markdown source verbatim. Accepted residual
// risk, not fixed here: both errors are effectively unreachable in practice
// (NewTermRenderer fails only on invalid style/option construction -- both
// hardcoded/valid here -- and Render() failing on well-formed markdown input
// is exceedingly rare), and the raw-text fallback already goes through
// bodySectionBody's caller-side wrapText-wrapped siblings (Meta/Historie) --
// only the Body section itself could, in the extremely unlikely error case,
// render an unwrapped line. Wrapping the fallback (wrapText(md, width)) would
// be the fix if this is ever observed in practice.
func glowRender(md string, width int) string {
	if strings.TrimSpace(md) == "" {
		return ""
	}
	w := width
	if w <= 0 {
		w = 80
	}
	style := "dark"
	if lipgloss.ColorProfile() == termenv.Ascii {
		style = "notty"
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(style),
		glamour.WithWordWrap(w),
	)
	if err != nil {
		return md
	}
	out, err := r.Render(md)
	if err != nil {
		return md
	}
	return strings.TrimRight(out, "\n")
}
