package tui

// footer_context.go — the context-sensitive Footer Zone 3 (PF-11
// Q04-Antwort, design-spec.md §15, epic-E7-plan.md Task 7, bean bt-m6at):
// while a Filter-Menu/node-action-Overlay/Search/Command-Center/Help fully
// captures input (handleKey's own capture-order precedent, update.go), the
// view-local bindings underneath (browseRepoLocalBindings/
// backlogLocalBindings) are non-functional -- showing them in the footer
// would actively mislead. contextualLocalHint swaps the footer to the
// ACTIVE capture state's own bindings instead.
//
// Two full-capture states are deliberately NOT cases below: an open huh Form
// (m.form != nil, forms_shared.go's formChrome) and m.confirmQuit (quitBox,
// box_confirm_quit.go) each already bake their own COMPLETE hint straight
// into their own modalPanel footer argument ("enter next/save · esc cancel"
// / "enter: quit   esc: cancel") -- there is no base-view fallback to build
// for either. The overlays covered below (Filter-Menu/Value-Menu/Tag-/
// Parent-/Blocking-Picker) also already render their OWN inline hint line
// at the top of their body (e.g. treeFilterBox's "space/x:toggle X:clear
// enter/esc/f:done") -- that is a separate, pre-existing surface; this
// file's job is strictly the OUTER footer (Zone 3, visible around/below the
// centered modal), which epic-E7-plan.md Task 7 Step 6 explicitly wants
// synced to the same active context too (Q04's underlying complaint: the
// outer footer kept showing STALE Tree/Backlog hints while a totally
// different overlay had full input capture).
import keybind "github.com/charmbracelet/bubbles/key"

// filterMenuCategoryHint is the Filter-Menu's OWN local tab/shift+tab
// binding (bt-nxuk, Reviewer-Finding B04 aus bt-2p9m-Review 2026-07-17):
// inside the open filter menu, tab/shift+tab switch the active facet
// category (keyFilterMenu, box_filter_facets.go) -- a DIFFERENT,
// filter-menu-local meaning from keys.FocusIn/keys.FocusOut's global
// Tree<->Detail focus-swap (keymap.go). filterMenuLocalBindings USED TO
// reuse keys.FocusIn/keys.FocusOut directly here, which meant Footer Zone 3
// rendered the stale global label ("tab focus in · shift+tab focus out")
// even though the Filter-Menu's own inline hint (treeFilterBox,
// box_filter_facets.go) already said "tab/shift+tab:category". This is a
// deliberately standalone keybind.Binding (NOT a keyMap struct field --
// see keymap_test.go's TestHelpGroupsCoverEveryBindingExactlyOnce, which
// reflects ONLY over keyMap fields and would flag an unreferenced field;
// staying local here keeps that guard, and
// TestNoDuplicateBindingBetweenGlobalAndAnyLocalHintList (scoped to
// browseRepoLocalBindings/backlogLocalBindings only), both untouched)
// mirroring treeFilterBox's own wording exactly.
var filterMenuCategoryHint = keybind.NewBinding(keybind.WithKeys("tab", "shift+tab"), keybind.WithHelp("tab/shift+tab", "category"))

// filterMenuLocalBindings is the Facet-Filter-Menu's own footer set
// (epic-E7-plan.md Task 7 Step 6, literal): keys.Toggle is exactly the
// "space: select/toggle" hint Q04 asked for, at the concrete overlay (the
// Filter-Menu) whose absence the PO actually noticed.
func filterMenuLocalBindings() []keybind.Binding {
	return []keybind.Binding{keys.Up, keys.Down, filterMenuCategoryHint, keys.Toggle, keys.FilterClear, keys.Enter, keys.Back}
}

// valueMenuLocalBindings is the Value-Menu overlay's own footer set
// (epic-E7-plan.md Task 7 Step 6, literal). keys.Status doubles as a close
// alias here (keyValueMenu, box_menu_value.go: opened by `s`, ALSO closes
// on a second `s`, same as Back) -- a genuine local binding of this
// overlay, not a stray global leaking through.
func valueMenuLocalBindings() []keybind.Binding {
	return []keybind.Binding{keys.Up, keys.Down, keys.Enter, keys.Status, keys.Back}
}

// tagPickerLocalBindings is the Tag-Picker overlay's own footer set.
//
// DEVIATION from epic-E7-plan.md Task 7 Step 6's literal text (which lumps
// Tag-/Parent-/Blocking-Picker into ONE Toggle-free {Up,Down,Enter,Back}
// set): keyTagPicker (box_picker_tag.go) actually wires a toggle. Omitting
// it here would silently hide a real, working key and leave Q04's own
// general wording ("wenn ein Form/Overlay aktiv ist ... inkl. 'space:
// select/toggle'") only half addressed -- Q04's PO example (the Filter-Menu)
// was illustrative, not exhaustive.
//
// keys.TagToggle, NOT keys.Toggle (ERRATUM/D01-Nachtrag, bean bt-9ipw
// Review-R1 B01): inside the picker's always-focused search field, "x" is a
// literal, typeable character -- only space toggles there. Advertising the
// shared "space/x" Toggle label here would mislead exactly the way the
// stale NewTag hint (below) would have.
//
// keys.NewTag REMOVED (bean bt-9ipw, US-07-Reopen 2026-07-17, D01): the
// former separate `n`-gated free-text new-tag sub-mode this hint used to
// point at is GONE -- D01 consolidated the Tag-Picker into ONE always-
// focused search field, so "n" is now just a literal, typeable character
// (e.g. filtering for a tag containing "n") rather than a picker command.
// Advertising it here would be actively misleading post-consolidation.
func tagPickerLocalBindings() []keybind.Binding {
	return []keybind.Binding{keys.Up, keys.Down, keys.TagToggle, keys.Enter, keys.Back}
}

// parentPickerLocalBindings is the Parent-Picker overlay's own footer set
// (epic-E7-plan.md Task 7 Step 6, literal) -- genuinely Toggle-free:
// keyParentPicker (box_picker_parent.go) is a single-select list, no
// space/x case at all.
func parentPickerLocalBindings() []keybind.Binding {
	return []keybind.Binding{keys.Up, keys.Down, keys.Enter, keys.Back}
}

// blockingPickerLocalBindings mirrors tagPickerLocalBindings' own Toggle
// deviation -- keyBlockingPicker (box_picker_blocking.go) also wires
// keys.Toggle (multi-select blocking-relation membership,
// toggleBlockPending).
func blockingPickerLocalBindings() []keybind.Binding {
	return []keybind.Binding{keys.Up, keys.Down, keys.Toggle, keys.Enter, keys.Back}
}

// confirmGateLocalBindings is the shared footer set for the two Confirm-Gate
// overlays (Create/Delete).
//
// GAP-FILL, not literally named by epic-E7-plan.md Task 7 Step 6: that
// step's overlay-specific-set enumeration only names Value-Menu and
// Tag-/Parent-/Blocking-Picker, leaving overlayCreateConfirm/
// overlayDeleteConfirm -- two more real `m.overlay != overlayNone` values
// (types.go) -- without an assigned set. Both keyCreateConfirm
// (box_confirm_create.go) and keyDeleteConfirm (box_confirm_delete.go)
// really only answer to Enter/Back (their own modal body already bakes in
// the destructive/constructive verb, e.g. "enter: delete permanently" --
// this base-view fallback intentionally stays generic).
func confirmGateLocalBindings() []keybind.Binding {
	return []keybind.Binding{keys.Enter, keys.Back}
}

// searchLocalBindings is the inline Tree/Backlog search input's own footer
// set while active (m.searchActive, keySearchInput, update.go) --
// epic-E7-plan.md Task 7 Step 6, literal.
func searchLocalBindings() []keybind.Binding {
	return []keybind.Binding{keys.Enter, keys.Back}
}

// paletteLocalBindings/helpLocalBindings are deliberately minimal base-view
// fallbacks (epic-E7-plan.md Task 7 Step 6, literal): paletteBox/helpBox
// (overlay_palette.go/overlay_shortcuts.go) already bake their own complete
// hint into their OWN modalPanel footer argument (non-empty, unlike the
// overlays above) -- the plan still names a base-view fallback for both, so
// one exists here too, kept intentionally short since the modal's own hint
// is the primary surface.
func paletteLocalBindings() []keybind.Binding { return []keybind.Binding{keys.Enter, keys.Back} }
func helpLocalBindings() []keybind.Binding    { return []keybind.Binding{keys.Back} }

// fullscreenDetailLocalBindings is the Detail-Vollbild's own Footer Zone 3
// set (F01 History-Stack, E9 Task 8, bean bt-1vbp, design-spec.md §15): the
// PO-Implementierungshinweis "im Footer/Help ausweisen" made concrete --
// ctrl+left/[ and ctrl+right/] (History Back/Forward) are wirksam ONLY
// while m.fullscreen == fullscreenDetail, shown here so the PO discovers
// them. keys.Back is repeated too (same convention as every other
// capture-state local set above, e.g. valueMenuLocalBindings) since esc's
// Vollbild-exit meaning here is a NEW, non-obvious D03 rung
// (keyDetailFocus's Back-case, update.go) worth reinforcing at the point
// the PO is actually looking.
func fullscreenDetailLocalBindings() []keybind.Binding {
	return []keybind.Binding{keys.HistoryBack, keys.HistoryForward, keys.Back}
}

// fullscreenListLocalBindings is the Listen-Vollbild's own Footer Zone 3
// set -- History-Keys are DELIBERATELY omitted (wirkungslos here: the
// History-Stack tracks Relations-Sprünge inside fullscreenDetail only,
// design-spec.md §15 Scope-Entscheidung). enter is the meaningful LOCAL key
// this mode adds (Listen-Vollbild -> Detail-Vollbild jump, keyFullscreen).
func fullscreenListLocalBindings() []keybind.Binding {
	return []keybind.Binding{keys.Back, keys.Enter}
}

// overlayLocalBindings dispatches m.overlay to its own footer set --
// extracted helper for contextualLocalHint's overlay case, below.
func overlayLocalBindings(o overlayID) []keybind.Binding {
	switch o {
	case overlayValueMenu:
		return valueMenuLocalBindings()
	case overlayTagPicker:
		return tagPickerLocalBindings()
	case overlayParentPicker:
		return parentPickerLocalBindings()
	case overlayBlockingPicker:
		return blockingPickerLocalBindings()
	case overlayCreateConfirm, overlayDeleteConfirm:
		return confirmGateLocalBindings()
	}
	return nil
}

// contextualLocalHint is Footer Zone 3's single source for BOTH
// Chrome-calling views (browseRepoChrome/view_browse_repo.go,
// backlogChrome/view_browse_backlog.go): view-local (viewLocal) in the
// normal state, but swaps to the active capture state's OWN bindings the
// instant one fully captures input (Q04-Antwort, PO-Nachtrag 5). Priority
// order (epic-E7-plan.md Task 7 Step 6): Filter-Menü > Overlay > Suche >
// Palette > Help > view-local default -- this is its OWN independently
// chosen order for footer display, NOT a mirror of handleKey's own
// full-capture dispatch order (update.go), which checks m.searchActive
// BEFORE m.filterOpen (the reverse of the order here). Harmless in
// practice since the two states are mutually exclusive (only one capture
// state is ever active at a time), but the two orderings must not be
// conflated (T7-Review I03, bean bt-dsog).
func (m model) contextualLocalHint(viewLocal []keybind.Binding) string {
	switch {
	case m.filterOpen:
		return renderBindings(filterMenuLocalBindings())
	case m.overlay != overlayNone:
		return renderBindings(overlayLocalBindings(m.overlay))
	case m.searchActive:
		return renderBindings(searchLocalBindings())
	case m.paletteOpen:
		return renderBindings(paletteLocalBindings())
	case m.helpOpen:
		return renderBindings(helpLocalBindings())
	case m.fullscreen == fullscreenDetail:
		// F01 (E9 Task 8, bean bt-1vbp): a further Capture-artiger Zustand,
		// slotted in AFTER Help/Palette/Suche/Overlay/Filter (none of those
		// can be active WHILE fullscreen != fullscreenNone anyway -- every
		// full-capture state routes handleKey's dispatch back to Split-Modus
		// first) but BEFORE the final viewLocal fallback, whose Tree/Backlog-
		// specific hints (tab/shift+tab/search/…) are non-functional here.
		return renderBindings(fullscreenDetailLocalBindings())
	case m.fullscreen == fullscreenList:
		return renderBindings(fullscreenListLocalBindings())
	}
	return renderBindings(viewLocal)
}
