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

// run executes `beans <args>` with RepoDir as the working directory and
// returns stdout. On failure, the returned error wraps stderr so callers
// (and tests) see the actual beans-CLI diagnostic, not just an exit code.
// stderr is only appended when non-empty, to avoid a dangling ": " on
// failures where the CLI wrote nothing to stderr.
func (c *Client) run(args ...string) ([]byte, error) {
	cmd := exec.Command("beans", args...)
	cmd.Dir = c.RepoDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr

	if err := cmd.Run(); err != nil {
		if msg := strings.TrimSpace(stderr.String()); msg != "" {
			return nil, fmt.Errorf("beans %s: %w: %s", args[0], err, msg)
		}
		return nil, fmt.Errorf("beans %s: %w", args[0], err)
	}

	return stdout.Bytes(), nil
}
