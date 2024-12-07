package find

import (
	"context"
	"fmt"
	"io"
	"maps"
	"path/filepath"
	"slices"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

type Param struct {
	Format   string
	PlanFile string
	Dir      string
	Root     string
	PWD      string
	Bucket   string
	Key      string
	Outputs  []string
	Stdout   io.Writer
}

type FileWithBackend struct {
	Terraform []TerraformBlock `hcl:"terraform,block"`
}

type TerraformBlock struct{}

type File struct {
	Path    string
	Content string
	Byte    []byte
}

type Dir struct {
	Path   string
	Files  []*File
	States []*RemoteState
}

func Find(_ context.Context, logE *logrus.Entry, afs afero.Fs, param *Param) error { //nolint:funlen,cyclop
	// parse plan file and extract changed outputs
	if err := validateParam(param); err != nil {
		return err
	}
	changedOutputs := param.Outputs
	if param.PlanFile != "" {
		arr, err := extractChangedOutputs(afs, param.PlanFile)
		if err != nil {
			return err
		}
		if len(arr) == 0 {
			logE.Info("no output changes")
			return nil
		}
		changedOutputs = arr
	}

	bucket := &Bucket{
		Bucket: param.Bucket,
		Key:    param.Key,
	}

	if param.PlanFile != "" {
		// parse HCLs in dir and extract backend configurations
		if err := findBackendConfig(afs, param.Dir, bucket); err != nil {
			return err
		}
	}

	if bucket.Type == "" {
		logE.Info("no backend configuration")
		return nil
	}
	logE.WithFields(logrus.Fields{
		"bucket": bucket.Bucket,
		"key":    bucket.Key,
		"prefix": bucket.Prefix,
		"type":   bucket.Type,
	}).Debug("backend configuration")

	// find HCLs in base directories and list directories where changed outputs are used
	tfFiles, err := findTFFiles(afs, param.Root)
	if err != nil {
		return err
	}
	logE.WithFields(logrus.Fields{
		"num_of_files": len(tfFiles),
	}).Debug("Found *.tf files")
	dirs := map[string]*Dir{}
	// Find files including a string "terraform_remote_state"
	if err := filterFilesWithRemoteState(afs, tfFiles, dirs); err != nil {
		return err
	}

	// Find terraform_remote_state data sources.
	for _, dir := range dirs {
		for _, file := range dir.Files {
			logE := logE.WithField("file", file.Path)
			logE.Debug("terraform_remote_state is found")
			remoteStates, err := extractRemoteStates(logE, file.Byte, file.Path, bucket)
			if err != nil {
				return fmt.Errorf("get terraform_remote_state: %w", logerr.WithFields(err, logrus.Fields{
					"file": file.Path,
				}))
			}
			dir.States = append(dir.States, remoteStates...)
		}
	}

	// Find attributes where changed outputs are used
	// data.terraform_remote_state.<name>.outputs.<output_name>
	// directory -> file -> changed outputs
	changed := map[string]map[string]map[string]struct{}{}
	findCaller(dirs, changedOutputs, changed)
	// Format the result to output as JSON
	changes, err := toChanges(param.PWD, param.Root, changed)
	if err != nil {
		return err
	}
	// Output the result
	if err := output(changes, param.Stdout, param.Format); err != nil {
		return err
	}
	return nil
}

func toChanges(pwd, baseDir string, changed map[string]map[string]map[string]struct{}) ([]*Change, error) {
	changes := make([]*Change, 0, len(changed))
	// baseDir is an absolute path or a relative path from the current directory
	if !filepath.IsAbs(baseDir) {
		baseDir = filepath.Join(pwd, baseDir)
	}
	for dir, m := range changed {
		// convert dir to the relative path from the base directory
		// dir is an absolute path or a relative path from the current directory
		absDir := dir
		if !filepath.IsAbs(dir) {
			absDir = filepath.Join(pwd, dir)
		}
		dir, err := filepath.Rel(baseDir, absDir)
		if err != nil {
			return nil, fmt.Errorf("get a relative path from baseDir to dir: %w", err)
		}
		files := make([]*ChangedFile, 0, len(m))
		for file, outputs := range m {
			// convert file to the relative path from dir
			// file is an absolute path or a relative path from the current directory
			if !filepath.IsAbs(file) {
				file = filepath.Join(pwd, file)
			}
			file, err := filepath.Rel(absDir, file)
			if err != nil {
				return nil, fmt.Errorf("get a relative path from baseDir to file: %w", err)
			}
			files = append(files, &ChangedFile{
				Path:    file,
				Outputs: slices.Sorted(maps.Keys(outputs)),
			})
		}
		changes = append(changes, &Change{
			Dir:   dir,
			Files: files,
		})
	}
	return changes, nil
}

func filterFilesWithRemoteState(afs afero.Fs, tfFiles []string, dirs map[string]*Dir) error {
	for _, matchFile := range tfFiles {
		// Find files including a string "terraform_remote_state"
		b, err := afero.ReadFile(afs, matchFile)
		if err != nil {
			return fmt.Errorf("read a file: %w", logerr.WithFields(err, logrus.Fields{
				"file": matchFile,
			}))
		}
		s := string(b)
		if !strings.Contains(s, "terraform_remote_state") {
			continue
		}
		dirPath := filepath.Dir(matchFile)
		dir, ok := dirs[dirPath]
		if !ok {
			dir = &Dir{
				Path: dirPath,
			}
		}
		dir.Files = append(dir.Files, &File{
			Path:    matchFile,
			Content: s,
			Byte:    b,
		})
		dirs[dirPath] = dir
	}
	return nil
}

type Change struct {
	Dir   string         `json:"dir"`
	Files []*ChangedFile `json:"files"`
}

type ChangedFile struct {
	Path    string   `json:"path"`
	Outputs []string `json:"outputs"`
}

type Bucket struct {
	Type   string `json:"type"`
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
	Prefix string `json:"prefix"`
}

func (b *Bucket) Compare(bucket *Bucket) bool {
	return b.Type == bucket.Type && b.Bucket == bucket.Bucket && b.Key == bucket.Key && b.Prefix == bucket.Prefix
}

type RemoteState struct {
	Name string
	File string
}
