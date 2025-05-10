package find

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/spf13/afero"
)

func ignoreDirs() []string {
	return []string{".terraform", ".git", ".github", "vendor", "node_modules"}
}

func findFiles(afs afero.Fs, baseDir string, pattern string) ([]string, error) {
	tfFiles := []string{}
	ignorePatterns := ignoreDirs()
	if err := doublestar.GlobWalk(afero.NewIOFS(afs), filepath.Join(baseDir, pattern), func(path string, d fs.DirEntry) error {
		if err := ignorePath(path, ignorePatterns); err != nil {
			return err
		}
		tfFiles = append(tfFiles, path)
		return nil
	}, doublestar.WithNoFollow(), doublestar.WithFilesOnly()); err != nil {
		return nil, fmt.Errorf("search files: %w", err)
	}
	return tfFiles, nil
}

func findTFFiles(afs afero.Fs, baseDir string) ([]string, error) {
	return findFiles(afs, baseDir, "**/*.tf")
}

func findTFJSONFiles(afs afero.Fs, baseDir string) ([]string, error) {
	return findFiles(afs, baseDir, "**/*.tf.json")
}

func ignorePath(path string, ignorePatterns []string) error {
	for _, ignoredDir := range ignorePatterns {
		f, err := doublestar.PathMatch("**/"+ignoredDir+"/**/*", path)
		if err != nil {
			return fmt.Errorf("check if a path is included in a ignored directory: %w", err)
		}
		if f {
			return fs.SkipDir
		}
	}
	return nil
}
