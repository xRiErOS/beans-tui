package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// ErrConflict is returned by every mutation below when the beans CLI
// rejects an update because the caller's ETag no longer matches the bean
// on disk (optimistic locking via --if-match). Callers should
// errors.Is(err, ErrConflict) and reload rather than retry blindly -- the
// TUI surfaces this as a toast + reload (see design context, E3.6).
var ErrConflict = errors.New("beans: stale etag (conflict)")

// conflictSubstring is the stable fragment of the beans CLI's stderr output
// on an --if-match mismatch, captured against beans 0.4.2:
//
//	Error: etag mismatch: provided <old>, current is <new>
//
// Matching on this substring (rather than exit code alone) is what lets
// update() distinguish a conflict from any other update failure.
const conflictSubstring = "etag mismatch"

// apiResponse mirrors the `{"success":true,"bean":{...}}` shape returned by
// both `beans create --json` and `beans update --json` on success.
type apiResponse struct {
	Success bool `json:"success"`
	Bean    Bean `json:"bean"`
}

// CreateOpts are the fields accepted by Create. Title is required; every
// other field is optional and simply omitted from the `beans create`
// invocation when left at its zero value.
type CreateOpts struct {
	Title     string
	Type      string
	Status    string
	Priority  string
	Parent    string
	Tags      []string
	BlockedBy []string
	Body      string
}

// Create creates a new bean via `beans create <title> ... --json` and
// returns the bean as reported by the CLI (including its freshly minted ID
// and ETag).
func (c *Client) Create(opts CreateOpts) (Bean, error) {
	args := []string{"create", opts.Title, "--json"}
	if opts.Type != "" {
		args = append(args, "--type", opts.Type)
	}
	if opts.Status != "" {
		args = append(args, "--status", opts.Status)
	}
	if opts.Priority != "" {
		args = append(args, "--priority", opts.Priority)
	}
	if opts.Parent != "" {
		args = append(args, "--parent", opts.Parent)
	}
	for _, tag := range opts.Tags {
		args = append(args, "--tag", tag)
	}
	for _, id := range opts.BlockedBy {
		args = append(args, "--blocked-by", id)
	}
	if opts.Body != "" {
		args = append(args, "--body", opts.Body)
	}

	out, err := c.run(args...)
	if err != nil {
		return Bean{}, err
	}

	var resp apiResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		return Bean{}, fmt.Errorf("beans create: parse output: %w", err)
	}
	return resp.Bean, nil
}

// update runs `beans update <id> --if-match <etag> <args...> --json`, the
// shared plumbing behind every setter/toggle below. On any failure it
// checks stderr for the ETag-conflict fragment and, if found, wraps
// ErrConflict so callers can errors.Is against it; any other failure is
// returned as-is (already carrying stderr context from run()).
func (c *Client) update(id, etag string, args ...string) error {
	fullArgs := append([]string{"update", id, "--if-match", etag}, args...)
	fullArgs = append(fullArgs, "--json")

	_, err := c.run(fullArgs...)
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), conflictSubstring) {
		return fmt.Errorf("%w: bean %s: %s", ErrConflict, id, err)
	}
	return err
}

// SetStatus sets a bean's status (see Bean.Status for valid values).
func (c *Client) SetStatus(id, status, etag string) error {
	return c.update(id, etag, "--status", status)
}

// SetPriority sets a bean's priority (see Bean.Priority for valid values).
func (c *Client) SetPriority(id, priority, etag string) error {
	return c.update(id, etag, "--priority", priority)
}

// SetType sets a bean's type (see Bean.Type for valid values).
func (c *Client) SetType(id, typ, etag string) error {
	return c.update(id, etag, "--type", typ)
}

// SetTitle sets a bean's title.
func (c *Client) SetTitle(id, title, etag string) error {
	return c.update(id, etag, "--title", title)
}

// AddTag adds a single tag to a bean.
func (c *Client) AddTag(id, tag, etag string) error {
	return c.update(id, etag, "--tag", tag)
}

// RemoveTag removes a single tag from a bean.
func (c *Client) RemoveTag(id, tag, etag string) error {
	return c.update(id, etag, "--remove-tag", tag)
}

// SetParent sets a bean's parent.
func (c *Client) SetParent(id, parent, etag string) error {
	return c.update(id, etag, "--parent", parent)
}

// RemoveParent clears a bean's parent.
func (c *Client) RemoveParent(id, etag string) error {
	return c.update(id, etag, "--remove-parent")
}

// AddBlockedBy adds a blocker relationship: id is blocked by target.
func (c *Client) AddBlockedBy(id, target, etag string) error {
	return c.update(id, etag, "--blocked-by", target)
}

// RemoveBlockedBy removes a blocker relationship: id is no longer blocked
// by target.
func (c *Client) RemoveBlockedBy(id, target, etag string) error {
	return c.update(id, etag, "--remove-blocked-by", target)
}

// AppendBody appends text to a bean's body (`--body-append`); it does not
// replace the existing body.
func (c *Client) AppendBody(id, text, etag string) error {
	return c.update(id, etag, "--body-append", text)
}

// Delete deletes a bean outright. `beans delete` normally prompts for
// confirmation on the CLI; --force skips that prompt (and any
// reference/child warnings) since bt drives this non-interactively.
func (c *Client) Delete(id string) error {
	_, err := c.run("delete", id, "--force")
	return err
}
