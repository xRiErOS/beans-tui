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
// This is now only a FALLBACK for failures that never print a JSON error
// envelope to stdout (e.g. pre-flight errors like "no .beans directory
// found" -- the command aborts before beans ever marshals a response).
// Whenever an envelope IS present, classifyError below parses its "code"
// field instead: matching on this substring alone false-positives when a
// user-supplied value (e.g. --type "etag mismatch") echoes into an
// unrelated error message (B02).
const conflictSubstring = "etag mismatch"

// apiResponse mirrors the `{"success":true,"bean":{...}}` shape returned by
// both `beans create --json` and `beans update --json` on success.
type apiResponse struct {
	Success bool `json:"success"`
	Bean    Bean `json:"bean"`
}

// errorEnvelope mirrors the `{"success":false,"error":"...","code":"..."}`
// shape the beans CLI prints to STDOUT (not stderr) when `create`/`update`
// fail after starting to process the command -- verified empirically
// against beans 0.4.2 (e.g. code "CONFLICT" on --if-match mismatch, code
// "VALIDATION_ERROR" on an invalid --type/--status/etc value). Pre-flight
// failures (no .beans directory, unknown flag) never reach this point, so
// stdout is empty and there's nothing to unmarshal -- see classifyError.
type errorEnvelope struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    string `json:"code"`
}

// classifyError turns a failed `create`/`update` invocation into the error
// callers should see. It prefers the JSON error envelope on stdout (B02):
// code "CONFLICT" becomes an ErrConflict-wrapped error; any other code
// becomes an error built from the envelope's own message text. Only when
// stdout carries no such envelope (pre-flight failures that abort before
// beans prints anything) does it fall back to sniffing cmdErr's stderr-derived
// message for conflictSubstring.
func classifyError(id string, stdout []byte, cmdErr error) error {
	var env errorEnvelope
	if json.Unmarshal(stdout, &env) == nil && env.Code != "" {
		if env.Code == "CONFLICT" {
			if id != "" {
				return fmt.Errorf("%w: bean %s: %s", ErrConflict, id, env.Error)
			}
			return fmt.Errorf("%w: %s", ErrConflict, env.Error)
		}
		if id != "" {
			return fmt.Errorf("beans: %s: bean %s: %s", env.Code, id, env.Error)
		}
		return fmt.Errorf("beans: %s: %s", env.Code, env.Error)
	}

	if strings.Contains(cmdErr.Error(), conflictSubstring) {
		if id != "" {
			return fmt.Errorf("%w: bean %s: %s", ErrConflict, id, cmdErr)
		}
		return fmt.Errorf("%w: %s", ErrConflict, cmdErr)
	}
	return cmdErr
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

// Create creates a new bean via `beans create ... --json -- <title>` and
// returns the bean as reported by the CLI (including its freshly minted ID
// and ETag).
//
// The title is placed AFTER a `--` separator, following every flag (B01):
// passed as a bare positional argument, a title starting with `-` (e.g.
// "--force", "- fix bug") is misparsed by cobra as an unknown flag. `--`
// tells cobra "everything after this is positional", which is the
// documented, verified fix (`beans create --type task --json -- --fix login
// bug` creates a bean titled "--fix login bug").
func (c *Client) Create(opts CreateOpts) (Bean, error) {
	args := []string{"create"}
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
	args = append(args, "--json", "--", opts.Title)

	out, err := c.run(args...)
	if err != nil {
		return Bean{}, classifyError("", out, err)
	}

	var resp apiResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		return Bean{}, fmt.Errorf("beans create: parse output: %w", err)
	}
	return resp.Bean, nil
}

// update runs `beans update <id> --if-match <etag> <args...> --json`, the
// shared plumbing behind every setter/toggle below. On any failure it
// delegates to classifyError, which parses the CLI's JSON error envelope
// (B02) to distinguish a genuine ETag conflict (code "CONFLICT", wrapped as
// ErrConflict so callers can errors.Is against it) from any other failure
// (e.g. a VALIDATION_ERROR whose message happens to contain the word
// "etag").
func (c *Client) update(id, etag string, args ...string) error {
	fullArgs := append([]string{"update", id, "--if-match", etag}, args...)
	fullArgs = append(fullArgs, "--json")

	out, err := c.run(fullArgs...)
	if err == nil {
		return nil
	}
	return classifyError(id, out, err)
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

// SetTags applies a combined tag diff in ONE `beans update` call (E3 Task 2,
// bean bt-8v69, design decision recorded in epic-E3-plan.md »Task 2«):
// `--tag`/`--remove-tag` are both `stringArray` flags (verified against
// `beans update --help`) and combine freely in a single invocation. This
// matters because the tag picker can toggle MULTIPLE tags in one session
// before confirming -- N sequential AddTag/RemoveTag calls against the SAME
// etag would be a conflict cascade (the first call wins and rotates the
// etag on disk, every subsequent call then sees a stale etag and fails
// ErrConflict). SetTags instead builds ONE `update` invocation carrying
// every added tag as a repeated `--tag` and every removed tag as a repeated
// `--remove-tag`, so the whole diff lands atomically against a single etag.
// AddTag/RemoveTag remain for genuine single-tag callers.
func (c *Client) SetTags(id string, add, remove []string, etag string) error {
	var args []string
	for _, tag := range add {
		args = append(args, "--tag", tag)
	}
	for _, tag := range remove {
		args = append(args, "--remove-tag", tag)
	}
	return c.update(id, etag, args...)
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
// confirmation on the CLI; --json skips that prompt (and any
// reference/child warnings) since bt drives this non-interactively --
// --json implies --force on the real binary (verified against beans
// 0.4.2's `beans delete --help`), and passing it also gets Delete the same
// JSON-envelope error reporting as every other mutation (I02), even though
// the parsed body isn't currently surfaced to callers.
func (c *Client) Delete(id string) error {
	_, err := c.run("delete", id, "--json")
	return err
}
