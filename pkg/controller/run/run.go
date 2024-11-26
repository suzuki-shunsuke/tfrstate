package run

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
	PlanFile string
	Dir      string
	Root     string
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

func Run(_ context.Context, logE *logrus.Entry, afs afero.Fs, param *Param) error { //nolint:funlen,cyclop
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

	if bucket.Key == "" || bucket.Bucket == "" {
		logE.Info("no s3 backend configuration")
		return nil
	}
	logE.WithFields(logrus.Fields{
		"bucket": bucket.Bucket,
		"key":    bucket.Key,
	}).Info("S3 buckend configuration")

	// find HCLs in root directories and list directories where changed outputs are used
	tfFiles, err := findTFFiles(afs)
	if err != nil {
		return err
	}
	logE.WithFields(logrus.Fields{
		"num_of_files": len(tfFiles),
	}).Info("Found *.tf files")
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
	changes := toChanges(changed)
	// Output the result
	if err := output(changes, param.Stdout); err != nil {
		return err
	}
	return nil
}

func toChanges(changed map[string]map[string]map[string]struct{}) []*Change {
	changes := make([]*Change, 0, len(changed))
	for dir, m := range changed {
		files := make([]*ChangedFile, 0, len(m))
		for file, outputs := range m {
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
	return changes
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
	Bucket string
	Key    string
}

type RemoteState struct {
	Name string
	File string
}
