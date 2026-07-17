# Contributing to beans-tui

Thanks for looking into this. `beans-tui` (`bt`) is a PO-cockpit TUI for
[beans](https://github.com/hmans/beans) repos — an independent client, not
affiliated with or endorsed by the beans project. It's maintained by one
person in spare time, so please keep PRs focused and expect review to take a
few days.

## Prerequisites

- Go 1.26+
- [beans CLI](https://github.com/hmans/beans) ≥ 0.4.2 on `PATH` (bt shells
  out to it — see `docs/plans/v1-port/design-spec.md` for why: beans stays
  the single authority over `.beans/` data, bt never parses/writes it
  directly)

## Build & test

```sh
git clone https://github.com/xRiErOS/beans-tui.git
cd beans-tui
command go build -o bin/bt .        # always `command go`, not `go` — a
                                     # shadowed `go` alias/function has bitten
                                     # this project before, see CLAUDE.md
command go test ./...               # full suite — required before any PR
command go test ./... -short        # fast local loop, skips the 7 slowest
                                     # huh-drive tests (~121s -> ~3-5s);
                                     # -short is NEVER a substitute for the
                                     # full run before you open a PR
```

Run `bt` against this repo's own `.beans/` to try it live — the project
dogfoods itself:

```sh
./bin/bt
```

## Before you open a PR

- **Tests first.** This project follows strict TDD (see `CLAUDE.md` →
  `superpowers:test-driven-development` if you use Claude Code, otherwise:
  write the failing test, then the code that makes it pass). Update-tests
  drive `tea.Update` via `tea.KeyMsg`; view changes get golden-file snapshots
  under `testdata/*.golden`.
- **Full test run is mandatory**, not `-short`, before every commit/PR — the
  skipped huh-drive tests catch real regressions in the Create-Form flow.
- **Footer/wrap changes need a manual check at 80 columns.** Unit tests at
  100/120 columns structurally can't see wrap bugs that only appear at the
  narrow end (NBSP-wordwrap traps etc. — see `docs/LESSONS-LEARNED.md`).
  Resize your terminal to 80 cols (or use `tmux` / `vhs`, see
  `docs/screenshots/screenshots.tape` for a working example) and eyeball it.
- **File naming** in `internal/tui/`: `<kind>_<verb>_<entity>.go`, prefixes
  `view_` / `form_` / `box_` / `picker_` / `overlay_`; infrastructure files
  carry no prefix. Match the existing pattern for whatever you're adding.
- **Keybindings** are single-source in `internal/tui/keymap.go` — the
  Help-Overlay (`?`) is generated from it. Don't hardcode a key check
  anywhere else.
- **Theme tokens** come only from `internal/theme/` (Catppuccin Macchiato,
  TrueColor). No hex literals in view code.
- `go vet ./...` clean.

## Commit style

Conventional-ish, lowercase type, imperative subject, ≤50 chars:

```
feat(tui): add X
fix(data): correct Y
docs: update Z
```

Body explains *why*, not *what* — the diff already shows what changed.

## Opening a PR

1. Fork, branch off `main` (`feat/short-description` or
   `fix/short-description`).
2. Keep the PR scoped to one change — smaller is faster to review.
3. Describe what changed and why in the PR body; link an issue if one
   exists.
4. CI (once set up) and a manual review by the maintainer both need to pass
   before merge. Force-pushes to your own branch during review are fine;
   don't rebase away requested-change history mid-review without a heads-up.

## Reporting bugs / requesting features

Use the GitHub issue templates (`.github/ISSUE_TEMPLATE/`) — they ask for
just enough to act on the report without a back-and-forth (repro steps,
`bt`/`beans`/Go/terminal versions for bugs; problem statement + proposed
behavior for features). Freeform issues are fine too if a template doesn't
fit.

## License

By contributing, you agree your contribution is licensed under this
project's [Apache-2.0 license](LICENSE).
