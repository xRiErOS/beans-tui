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

// relationField is a jump-only navigation target inside the Beziehungen
// section (never an edit target -- edit fields are E3 scope). beanID == ""
// marks an unresolved/dangling reference: not jumpable, Task 2's enter
// handler must guard on this.
type relationField struct {
	beanID string // target bean ID; "" for an unresolved/dangling reference (not jumpable)
	label  string // pre-rendered row text (theme colors already applied)
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
// detailFocusView{sec,field} -- fieldStrip only ever renders for the
// Beziehungen section (the only section carrying fields in E2).
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
		isOpen := n == open
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
		header := truncate(marker+title+hint, w)
		if activeSec {
			// D08: active section = ▌ bar + whole line accent-tinted (own
			// per-cell colors stripped first, mirrors the tree cursor).
			header = theme.Accent.Render("▌" + truncate(ansi.Strip(marker+title+hint), w-1))
		}
		b.WriteString(headerStyle.Render(header) + "\n")
		if isOpen {
			// Field-level focus shows the field strip (which field a jump
			// would target) before the body.
			if activeSec && len(s.fields) > 0 {
				b.WriteString(fieldStrip(s.fields, fieldIdx, w) + "\n")
			}
			b.WriteString(boxStyle.Render(s.body) + "\n")
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

// fieldStrip rendert die Beziehungen-Felder einer Section als kompakte Zeile
// mit dem aktiven Feld (D08-Balken + Accent), die übrigen muted (port devd
// accordion.go:360-373, detailField.label -> relationField.label).
func fieldStrip(fields []relationField, active, w int) string {
	if len(fields) == 0 {
		return theme.Muted.Render("Fields: (none)")
	}
	parts := make([]string, len(fields))
	for i, f := range fields {
		if i == active {
			parts[i] = theme.Accent.Render("▌" + f.label)
		} else {
			parts[i] = theme.Muted.Render(f.label)
		}
	}
	return truncate(theme.Muted.Render("Fields: ")+strings.Join(parts, "  "), w)
}

// glowRender rendert Markdown für die Body-Section via glamour. Fällt auf den
// rohen Text zurück, wenn glamour fehlschlägt. Im Ascii-Profil (Tests/kein
// TTY -> an dasselbe Profil gekoppelt wie die Goldens) wird der plain
// "notty"-Style genutzt, damit Golden-Snapshots ANSI-frei und stabil bleiben
// (port devd editor.go:102-126, verbatim).
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
