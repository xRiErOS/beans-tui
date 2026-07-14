package data

import (
	"fmt"
	"os"
	"path/filepath"
)

// configFileName is the beans repo marker: FindRepo walks upward looking
// for a directory containing this file.
const configFileName = ".beans.yml"

// FindRepo walks upward from start until it finds a directory containing
// .beans.yml (the beans repo config) and returns that directory's absolute
// path. It returns an error if no such directory exists between start and
// the filesystem root.
func FindRepo(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", fmt.Errorf("resolve start dir %q: %w", start, err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, configFileName)); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no %s found above %s", configFileName, start)
		}
		dir = parent
	}
}
