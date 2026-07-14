package data

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// requireBeansBinary skips the test if the beans CLI is not available in
// PATH — the data layer is a thin wrapper around the real binary, so tests
// exercise it directly rather than mocking exec.Command.
func requireBeansBinary(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("beans"); err != nil {
		t.Skip("beans binary not found in PATH, skipping")
	}
}

const testBeansYML = `project:
    name: test-repo

beans:
    path: .beans
    prefix: tt-
    id_length: 4
    default_status: todo
    default_type: task
`

// newTestRepo creates a temporary beans repo (tmp dir with .beans.yml,
// prefix "tt-") containing 3 fixture beans — a milestone, an epic (child of
// the milestone), and a task (child of the epic) — written as real .md
// files with YAML frontmatter, matching the on-disk format beans itself
// produces. The milestone carries "blocking: [tt-task]", the epic carries
// "tags", and the task carries "blocked_by: [tt-mlst]", so List() callers
// exercise the full JSON contract (slice fields), not just the scalar ones.
//
// Deliberate placement quirk: "tags" specifically sits on the epic, not the
// task, because of a real beans 0.4.2 CLI bug found while building this
// fixture -- for a bean whose on-disk frontmatter includes a hand-authored
// "tags:" block, the ETag `list`/`show` report does not match the ETag
// `update --if-match`'s conflict check computes internally (verified in
// isolation: "blocked_by"/"blocking"/"parent" do NOT have this divergence,
// only "tags" does; a bean's ETag becomes consistent again once beans
// itself has rewritten the file once, e.g. via any successful update). The
// task fixture is what client_mut_test.go drives through SetStatus/
// SetPriority/AppendBody/Delete round-trips using a List()-obtained ETag,
// so it must stay free of hand-authored "tags" to keep those tests
// deterministic. See bt-tknb concerns for the upstream bug report.
//
// It returns the repo root directory.
func newTestRepo(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".beans.yml"), []byte(testBeansYML), 0o644); err != nil {
		t.Fatalf("write .beans.yml: %v", err)
	}

	beansDir := filepath.Join(dir, ".beans")
	if err := os.MkdirAll(beansDir, 0o755); err != nil {
		t.Fatalf("mkdir .beans: %v", err)
	}

	fixtures := map[string]string{
		"tt-mlst--test-milestone.md": `---
# tt-mlst
title: Test Milestone
status: todo
type: milestone
priority: high
created_at: 2026-01-01T00:00:00Z
updated_at: 2026-01-01T00:00:00Z
blocking:
    - tt-task
---

Milestone fixture body.
`,
		"tt-epic--test-epic.md": `---
# tt-epic
title: Test Epic
status: in-progress
type: epic
priority: normal
created_at: 2026-01-01T00:00:00Z
updated_at: 2026-01-01T00:00:00Z
parent: tt-mlst
tags:
    - urgent
    - backend
---

Epic fixture body.
`,
		"tt-task--test-task.md": `---
# tt-task
title: Test Task
status: todo
type: task
priority: normal
created_at: 2026-01-01T00:00:00Z
updated_at: 2026-01-01T00:00:00Z
parent: tt-epic
blocked_by:
    - tt-mlst
---

Task fixture body.
`,
	}

	for name, content := range fixtures {
		if err := os.WriteFile(filepath.Join(beansDir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("write fixture %s: %v", name, err)
		}
	}

	return dir
}
