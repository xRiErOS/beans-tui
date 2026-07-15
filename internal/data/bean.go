// Package data is the single touchpoint between bt and the beans CLI. Every
// read/write/watch of bean data goes through this package, which shells out
// to the `beans` binary rather than reimplementing its storage format
// (design decision D02: the beans binary stays the single authority).
package data

import (
	"strings"
	"time"
)

// Bean mirrors a single issue as reported by the beans CLI. Field names and
// JSON tags are verified against real `beans list --json --full` and
// `beans show <id> --json` output (beans 0.4.2): Tags/Parent/Blocking/
// BlockedBy are omitted by beans when unset, which json.Unmarshal handles
// as zero values (nil slice / empty string) — no deviation from the
// contract found.
type Bean struct {
	ID        string     `json:"id"`
	Slug      string     `json:"slug"`
	Path      string     `json:"path"`
	Title     string     `json:"title"`
	Status    string     `json:"status"`   // draft|todo|in-progress|completed|scrapped
	Type      string     `json:"type"`     // milestone|epic|feature|task|bug
	Priority  string     `json:"priority"` // critical|high|normal|low|deferred
	Tags      []string   `json:"tags"`
	Parent    string     `json:"parent"`
	Blocking  []string   `json:"blocking"`
	BlockedBy []string   `json:"blocked_by"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
	Body      string     `json:"body"` // only populated with --full
	ETag      string     `json:"etag"`
}

// IsArchived reports whether b currently lives under .beans/archive/ (E5
// Task 7, bean bt-ggt2, design decision e, epic-E5-plan.md »Task 7«):
// Core.isArchivedPath (beans-src/pkg/beancore/core.go:826) stamps every
// archived bean's Path with a leading "archive/" segment -- a cheap,
// already-loaded-field derivation, no extra CLI round-trip needed. Verified
// empirically against a real `beans archive` run (archive_test.go,
// TestListIncludesArchivedBeans): Path goes from e.g.
// "tt-nhra--task-a.md" to "archive/tt-nhra--task-a.md", nothing else
// changes (Status/Tags/relationships are all preserved by `beans archive`).
func (b Bean) IsArchived() bool {
	return strings.HasPrefix(b.Path, "archive/")
}
