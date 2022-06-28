package shipper

import (
	"fmt"
	"path/filepath"
	"strings"
)

// getParentPaths builds all the possible directories to look for a config file
func getParentPaths(dir string) []string {
	paths := []string{}

	for {
		paths = append(paths, dir)

		if dir == "/" {
			break
		}

		dir = filepath.Dir(dir)
	}

	return paths
}

// expandWildcardPaths expands any paths that are globs into a list of paths
func expandWildcardPaths(dir string, paths []string) ([]string, error) {
	expandedPaths := []string{}

	for _, p := range paths {
		files, err := expandWildcardPath(dir, p)
		if err != nil {
			return nil, err
		}

		expandedPaths = append(expandedPaths, files...)
	}

	return expandedPaths, nil
}

func expandWildcardPath(dir string, p string) ([]string, error) {
	// convert the relative paths into absolute paths
	absP := filepath.Join(dir, p)

	// if there is an asterisk in the path, expand it
	if strings.Contains(p, "*") {
		files, err := filepath.Glob(absP)
		if err != nil {
			return nil, fmt.Errorf("failed to expand path %s", p)
		}

		out := []string{}

		for _, f := range files {
			if !strings.Contains(f, ".DS_Store") {
				out = append(out, f)
			}
		}

		return out, nil
	}

	// otherwise just add the path
	return []string{absP}, nil
}
