package data

// tags.go — the tag-name grammar the beans CLI accepts (E3 Task 2, bean
// bt-8v69, epic body bt-gzcu): lowercase alnum segments, single-hyphen-
// separated, leading letter. Own file (not bean.go, which stays a pure JSON
// contract) -- ENTSCHEIDUNG plan »Task 2« Files list.

import "regexp"

// tagNameRe mirrors the tag grammar the beans CLI accepts (epic bt-gzcu):
// lowercase alnum segments, single-hyphen-separated, leading letter.
var tagNameRe = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)

// ValidTagName reports whether s is a syntactically valid beans tag name.
// The picker (box_picker_tag.go) validates free-text tag entry against this
// BEFORE ever dispatching a mutation, so an invalid name never reaches the
// CLI (which would reject it server-side anyway, but with a less specific
// error).
func ValidTagName(s string) bool { return tagNameRe.MatchString(s) }
