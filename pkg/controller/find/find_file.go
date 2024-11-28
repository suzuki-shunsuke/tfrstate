package find

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/spf13/afero"
)

func findTFFiles(afs afero.Fs, baseDir string) ([]string, error) {
	// Find **/*.tf
	tfFiles := []string{}
	ignorePatterns := []string{".terraform", ".git", ".github", "vendor", "node_modules"}
	if err := doublestar.GlobWalk(afero.NewIOFS(afs), filepath.Join(baseDir, "**/*.tf"), func(path string, d fs.DirEntry) error {
		if err := ignorePath(path, ignorePatterns); err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		tfFiles = append(tfFiles, path)
		return nil
	}, doublestar.WithNoFollow()); err != nil {
		return nil, fmt.Errorf("search files: %w", err)
	}
	return tfFiles, nil
}

func ignorePath(path string, ignorePatterns []string) error {
	for _, ignoredDir := range ignorePatterns {
		f, err := doublestar.PathMatch(ignoredDir, path)
		if err != nil {
			return fmt.Errorf("check if a path is included in a ignored directory: %w", err)
		}
		if f {
			return fs.SkipDir
		}
	}
	return nil
}
