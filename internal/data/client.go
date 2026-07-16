package data

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// Client wraps the beans CLI as a subprocess, scoped to a single repo
// directory (the dir containing .beans.yml, as found by FindRepo).
type Client struct {
	RepoDir string
}

// List returns all beans in the repo, including their body text.
func (c *Client) List() ([]Bean, error) {
	out, err := c.run("list", "--json", "--full")
	if err != nil {
		return nil, err
	}

	var beans []Bean
	if err := json.Unmarshal(out, &beans); err != nil {
		return nil, fmt.Errorf("beans list --json --full: parse output: %w", err)
	}
	return beans, nil
}

// Search runs a Bleve full-text query (title+body) via `beans list -S`, with
// the same --full/--json contract as List (E2 Task 3, bean bt-4ep2,
// design-spec.md §6 V2: "Bleve-Modus ab 3 Zeichen"). The query itself is not
// validated/escaped here -- the beans CLI's Bleve query-string syntax
// (fuzzy/wildcard/field-scoped) passes straight through, same as List's own
// thin-wrapper contract.
func (c *Client) Search(query string) ([]Bean, error) {
	out, err := c.run("list", "--json", "--full", "--search", query)
	if err != nil {
		return nil, err
	}

	var beans []Bean
	if err := json.Unmarshal(out, &beans); err != nil {
		return nil, fmt.Errorf("beans list --search: parse output: %w", err)
	}
	return beans, nil
}

// ShowRaw returns id's full markdown representation exactly as
// `beans show <id> --raw` prints it -- verified byte-identical to the
// on-disk .beans/*.md file (client_test.go's TestShowRawReturnsFileFormat).
// This is the seed text for the whole-bean $EDITOR (D01, design-spec.md §15
// PF-17, bean bt-z4b1): no self-built markdown templating, the CLI stays
// the ONE authority for the file's canonical serialization (design-spec
// §3.1 D02). A pure read: --raw prints no JSON envelope, so there is no
// classifyError path here, unlike every mutations.go call.
func (c *Client) ShowRaw(id string) (string, error) {
	out, err := c.run("show", id, "--raw")
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// run executes `beans <args>` with RepoDir as the working directory and
// returns stdout. On failure, stdout is still returned alongside the error
// (rather than nil) -- mutations.go's classifyError parses it as a JSON
// error envelope when present. The error itself wraps the first line of
// stderr so callers (and tests) see the actual beans-CLI diagnostic without
// cobra's ~25-line usage dump (B03). stderr is only appended when
// non-empty, to avoid a dangling ": " on failures where the CLI wrote
// nothing to stderr.
func (c *Client) run(args ...string) ([]byte, error) {
	cmd := exec.Command("beans", args...)
	cmd.Dir = c.RepoDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr

	if err := cmd.Run(); err != nil {
		if msg := firstLine(stderr.String()); msg != "" {
			return stdout.Bytes(), fmt.Errorf("beans %s: %w: %s", args[0], err, msg)
		}
		return stdout.Bytes(), fmt.Errorf("beans %s: %w", args[0], err)
	}

	return stdout.Bytes(), nil
}

// firstLine trims cobra's error output down to just the diagnostic line,
// dropping the "\nUsage:\n..." flag-help dump that cobra appends after
// every CLI error (B03). Without this, every wrapped error embedded ~25
// lines of usage text, which isn't toast-suitable.
func firstLine(stderr string) string {
	s := strings.TrimSpace(stderr)
	if i := strings.Index(s, "\nUsage:"); i >= 0 {
		s = s[:i]
	}
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSpace(s)
}
