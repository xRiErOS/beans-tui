# jira-Style Flat/Salient UI — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** beans-tui im Spike auf ein flacheres, salienteres jira-artiges Modell umbauen — getitelte Dropdown-Boxen mit Hotkey-im-Rahmen, persistente Filter-Leiste, Nested/Flat-Toggle, huh→Inline-Box-Editing — und prüfen, ob das eine Verbesserung ist.

**Architecture:** Neues Render-Primitiv `dropdownBox` (manuell komponierte Rahmenzeilen, da lipgloss keinen Text-im-Rahmen kann). Detail-Pane und Filter-Leiste teilen dieses eine Widget (D01). State (Facetten) liegt bereits model-weit. Keymap bleibt Single-Source (`keymap.go`). Farbe nur über `internal/theme`-Tokens.

**Tech Stack:** Go, bubbletea (Elm), lipgloss, `charmbracelet/x/ansi` (Width/Strip/Wrap), VHS (`.tape`→GIF für visuellen Smoke), Golden-Snapshots + `tea.KeyMsg`-Update-Tests.

**Spec:** `docs/plans/jira-style-experiment/design-spec.md` (D01–D09, Farbkarte, Mockups, I01–I04).

**Branch:** `experiment/jira-style-ui` · **Merge:** nur nach PO-Abnahme gegen Erfolgskriterium (Spec §1). Kein Merge-Zwang.

---

## Slice-Roadmap (je Slice = lauffähige, testbare Software)

Nach writing-plans-Scope-Check: mehrere Subsysteme → Slice-weise. Reihenfolge = Abhängigkeit. Slice 1 ist unten voll TDD-ausgeführt; die weiteren Slices werden beim Erreichen voll detailliert (oder vorab auf Wunsch).

| Slice | Ziel | Kern-Dateien | Test | Abnahme |
|-------|------|--------------|------|---------|
| **S1** | `dropdownBox`-Primitiv (Label-im-Rahmen, `▾`, Hotkey-im-Rahmen, Fokus-Farbe, Breiten-Clamp) | `box_dropdown.go` (neu) | Unit: 3 Zeilen, Breite exakt, Label/Value/Hotkey/`▾` an richtiger Stelle, Fokus=Mauve | Grüne Unit-Tests |
| **S2** | Detail-Pane als gestapelte Box-Form (D03/D04), read-only. Accordion-Render ersetzt | `view_detail_bean.go`, `render_shared.go` | Golden-Snapshot Detail; responsive 3/2/1-up | Golden + VHS 80/100 |
| **S3** | Persistente Filter-Leiste (D02), `f` fokussiert Leiste, aktiver Chip=Peach | `box_filter_facets.go`, `view_browse_repo.go` | Golden Browse mit Leiste | Golden + VHS |
| **S4** | Keymap + Picker (D03/D07): `o` Type, `u` Priority, `G` View; Type-/Priority-Picker | `keymap.go`, `box_picker_type.go` (neu), `box_picker_priority.go` (neu), `update.go` | Update-Tests je Key; Help-Drift-Guard | Grüne Tests |
| **S5** | Nested/Flat-Switcher (D05): `G` toggelt; Flat-Tabelle (Default Hierarchie, `S`→flach) | `view_browse_repo.go`, `view_flat_list.go` (neu) | Golden Nested + Flat; Sort-Test | Golden + VHS |
| **S6** | Maus (D06): Klick öffnet Dropdown/Toggle, Filter-Chip, View-Segment | `mouse.go` | `tea.MouseMsg`-Tests je Zone | Grüne Tests |
| **S7** | huh→Inline-Box-Editing (D09): Create/Edit in Boxen, eigene maus-native Popups; huh + langsame Tests raus | `form_create_bean.go`→Ersatz, `box_menu_value.go`, `update.go` | Update-Tests Inline-Edit; Entfall `skipSlowHuhDriveInShortMode` | Grüne Tests, huh weg |
| **S8** | Politur + VHS-Smoke 80/100 + PO-Abnahme; Merge/Verwerfen entscheiden | `.tape`, `docs/…` | VHS-GIF beide Breiten | PO-Verdikt |

Risiken je Slice adressieren: **I01** (Box-in-Box-Dichte) in S2 prüfen — ggf. Pane-Rahmen weglassen. **I02** (80-Spalten) in S2/S3/S5 per VHS. **I03** (Höhe) in S2. **I04** (D09-Größe) durch S7-Isolierung.

---

## Slice 1: `dropdownBox`-Primitiv

Das eine wiederverwendete Widget (D01). Rendert drei Zeilen fester Breite:

```
╭─ Status ───────────────────╮   <- Label im oberen Rahmen (Muted)
│ ● todo                 ▾   │   <- Wert + ▾-Affordance (Chevron/Peach)
╰──────────────────── (s) ───╯   <- Hotkey unten-rechts im Rahmen (BindingKey/Teal)
```

lipgloss kann Text-im-Rahmen nicht → Rahmenzeilen werden manuell aus Box-Zeichen komponiert. Breiten-/ANSI-korrekt über `lipgloss.Width` (Werte tragen Farbe).

**Files:**
- Create: `internal/tui/box_dropdown.go`
- Test: `internal/tui/box_dropdown_test.go`

- [ ] **Step 1: Failing Test schreiben**

```go
package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// go test rendert unter der NoColor-Profil-Variante (kein TTY) -> theme.*.Render
// gibt Klartext ohne ANSI zurück, Assertions laufen auf sichtbaren Zeichen.
func TestDropdownBoxLayout(t *testing.T) {
	const w = 30
	out := dropdownBox("Status", "todo", "s", w, false)
	lines := strings.Split(out, "\n")
	if len(lines) != 3 {
		t.Fatalf("want 3 lines, got %d: %q", len(lines), out)
	}
	for i, ln := range lines {
		if got := lipgloss.Width(ln); got != w {
			t.Errorf("line %d width = %d, want %d: %q", i, got, w, ln)
		}
	}
	top, mid, bot := ansi.Strip(lines[0]), ansi.Strip(lines[1]), ansi.Strip(lines[2])
	if !strings.HasPrefix(top, "╭─ Status ") {
		t.Errorf("top border missing label: %q", top)
	}
	if !strings.Contains(mid, "todo") || !strings.Contains(mid, "▾") {
		t.Errorf("mid missing value/▾: %q", mid)
	}
	if !strings.Contains(bot, "(s)") || !strings.HasPrefix(bot, "╰") {
		t.Errorf("bottom border missing hotkey: %q", bot)
	}
}

// Ohne Hotkey: unterer Rahmen ist reine Rahmenlinie, kein "()".
func TestDropdownBoxNoHotkey(t *testing.T) {
	out := dropdownBox("Type", "task", "", 24, false)
	bot := ansi.Strip(strings.Split(out, "\n")[2])
	if strings.Contains(bot, "(") {
		t.Errorf("empty hotkey must not render parens: %q", bot)
	}
	if lipgloss.Width(strings.Split(out, "\n")[2]) != 24 {
		t.Errorf("bottom width drift without hotkey")
	}
}
```

- [ ] **Step 2: Test laufen lassen, Fehlschlag prüfen**

Run: `command go test ./internal/tui/ -run TestDropdownBox -v`
Expected: FAIL — `undefined: dropdownBox`

- [ ] **Step 3: Minimale Implementierung**

```go
package tui

// box_dropdown.go — das eine wiederverwendete jira-Style Dropdown-Widget
// (design-spec.md Experiment D01): Label im oberen Rahmen, Wert + ▾ innen,
// Hotkey unten-rechts im Rahmen. lipgloss kann keinen Text-im-Rahmen -> die
// Rahmenzeilen werden hier manuell aus Box-Zeichen komponiert. Breiten- und
// ANSI-korrekt über lipgloss.Width (Werte tragen Theme-Farben). Wird von der
// Detail-Pane (S2) UND der Filter-Leiste (S3) genutzt.

import (
	"strings"

	"github.com/xRiErOS/beans-tui/internal/theme"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// clampVisible kürzt s auf max n sichtbare Zellen (ANSI-agnostisch über
// ansi.Truncate, das den Reset am Ende selbst schließt).
func clampVisible(s string, n int) string {
	if n < 0 {
		n = 0
	}
	return ansi.Truncate(s, n, "")
}

// repeat baut eine Rahmen-Dashline gegebener Länge (nie negativ).
func borderDashes(n int) string {
	if n < 1 {
		return ""
	}
	return strings.Repeat("─", n)
}

// dropdownBox rendert das 3-Zeilen-Widget in exakt width Zellen Breite.
// label sitzt im oberen Rahmen, value + ▾ innen, hotkey (falls != "") unten-
// rechts im Rahmen. focused = Mauve-Rahmen, sonst Overlay.
func dropdownBox(label, value, hotkey string, width int, focused bool) string {
	if width < 8 {
		width = 8 // absoluter Boden, sonst kollabiert die Komposition
	}
	borderColor := theme.Overlay
	if focused {
		borderColor = theme.Mauve
	}
	frame := lipgloss.NewStyle().Foreground(borderColor)

	// --- Oben: ╭─ <label> <dashes>╮
	labelSeg := "─ " + clampVisible(label, width-6) + " "
	topFill := width - 2 - lipgloss.Width(labelSeg) // -2 für ╭ und ╮
	top := frame.Render("╭") + frame.Render(labelSeg) + frame.Render(borderDashes(topFill)) + frame.Render("╮")

	// --- Mitte: │ <value> ... ▾ │
	arrow := theme.Chevron.Render("▾")
	// innerer Platz zwischen "│ " und " ▾ │" = width - 6
	inner := width - 6
	val := clampVisible(value, inner)
	pad := inner - lipgloss.Width(val)
	if pad < 0 {
		pad = 0
	}
	mid := frame.Render("│") + " " + val + strings.Repeat(" ", pad) + " " + arrow + " " + frame.Render("│")

	// --- Unten: ╰<dashes> (hotkey) <dashes>╯  bzw. reine Linie
	var bot string
	if hotkey == "" {
		bot = frame.Render("╰") + frame.Render(borderDashes(width-2)) + frame.Render("╯")
	} else {
		badge := theme.BindingKey.Render("(" + hotkey + ")")
		badgeSeg := " " + badge + " "        // ein Space je Seite
		right := frame.Render(borderDashes(3)) // 3 Dashes rechts vom Badge
		fill := width - 2 - lipgloss.Width(badgeSeg) - lipgloss.Width("───")
		if fill < 1 {
			fill = 1
		}
		bot = frame.Render("╰") + frame.Render(borderDashes(fill)) + badgeSeg + right + frame.Render("╯")
	}

	return top + "\n" + mid + "\n" + bot
}
```

- [ ] **Step 4: Test laufen lassen, grün prüfen**

Run: `command go test ./internal/tui/ -run TestDropdownBox -v`
Expected: PASS (beide Tests). Falls Breiten-Assertion bricht: `lipgloss.Width` der drei Zeilen ausgeben und die `topFill`/`fill`/`pad`-Rechnung nachziehen (Off-by-one an `─ label ` ist der wahrscheinliche Punkt).

- [ ] **Step 5: Voller Testlauf (kein `-short` vor Commit, Projekt-Regel)**

Run: `command go test ./... `
Expected: PASS — keine bestehenden Snapshots berührt (neues Primitiv hat noch keinen Aufrufer).

- [ ] **Step 6: Commit**

```bash
git add internal/tui/box_dropdown.go internal/tui/box_dropdown_test.go
git commit -m "feat(experiment): dropdownBox primitive (label/hotkey in border)" \
  -m "S1 der jira-Style-Spike. lipgloss kann keinen Text-im-Rahmen -> Rahmenzeilen manuell komponiert. Reused von Detail-Pane (S2) + Filter-Leiste (S3)." \
  -m "Refs: experiment/jira-style-ui"
```

---

## Nach Slice 1

Slices S2–S8 werden beim Erreichen voll TDD-detailliert (writing-plans-Scope-Check: Subsystem-weise Pläne). Jede Slice endet mit grünem vollem Testlauf + (wo Layout betroffen) VHS-Smoke bei 80 UND 100 Spalten, bevor die nächste beginnt.
