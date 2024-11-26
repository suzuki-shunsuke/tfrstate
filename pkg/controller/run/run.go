package run

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

type Param struct {
	PlanFile string
	Dir      string
	Root     string
}

type FileWithBackend struct {
	Terraform []TerraformBlock `hcl:"terraform,block"`
}

type TerraformBlock struct{}

func Run(_ context.Context, logE *logrus.Entry, afs afero.Fs, param *Param) error { //nolint:funlen,cyclop
	// parse plan file and extract changed outputs
	planFile := &PlanFile{}
	if err := readPlanFile(afs, param.PlanFile, planFile); err != nil {
		return fmt.Errorf("read a plan file: %w", err)
	}
	excludeCreatedOutputs(planFile)
	if len(planFile.OutputChanges) == 0 {
		logE.Info("no output changes")
		return nil
	}

	// parse HCLs in dir and extract backend configurations
	matchFiles, err := afero.Glob(afs, filepath.Join(param.Dir, "*.tf"))
	if err != nil {
		return fmt.Errorf("glob *.tf to get Backend configuration: %w", err)
	}
	bucket := &Bucket{}
	for _, matchFile := range matchFiles {
		b, err := afero.ReadFile(afs, matchFile)
		if err != nil {
			return fmt.Errorf("read a file: %w", err)
		}
		s := string(b)
		if !strings.Contains(s, "backend") {
			continue
		}
		if f, err := extractBackend(b, matchFile, bucket); err != nil {
			return fmt.Errorf("get backend configuration: %w", err)
		} else if f {
			break
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
	states := []*RemoteState{}
	for _, matchFile := range tfFiles {
		logE := logE.WithField("file", matchFile)
		b, err := afero.ReadFile(afs, matchFile)
		if err != nil {
			return fmt.Errorf("read a file: %w", err)
		}
		s := string(b)
		if !strings.Contains(s, "terraform_remote_state") {
			continue
		}
		logE.Debug("terraform_remote_state is found")
		remoteStates, err := extractRemoteStates(logE, b, matchFile, bucket)
		if err != nil {
			return fmt.Errorf("get terraform_remote_state: %w", err)
		}
		states = append(states, remoteStates...)
	}
	for _, state := range states {
		fmt.Println(state.Name, state.File) //nolint:forbidigo
	}
	return nil
}

func findTFFiles(afs afero.Fs) ([]string, error) {
	tfFiles := []string{}
	ignorePatterns := []string{".terraform", ".git", ".github", "vendor", "node_modules"}
	if err := doublestar.GlobWalk(afero.NewIOFS(afs), "**/*.tf", func(path string, d fs.DirEntry) error {
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

type PlanFile struct {
	OutputChanges map[string]*OutputChange `json:"output_changes"`
}

type OutputChange struct {
	Actions         []string `json:"actions"`
	Before          any      `json:"before"`
	After           any      `json:"after"`
	AfterUnknown    bool     `json:"after_unknown"`
	BeforeSensitive bool     `json:"before_sensitive"`
	AfterSensitive  bool     `json:"after_sensitive"`
}

func excludeCreatedOutputs(file *PlanFile) {
	for name, change := range file.OutputChanges {
		if len(change.Actions) == 1 && change.Actions[0] == "create" {
			delete(file.OutputChanges, name)
		}
	}
}

func readPlanFile(fs afero.Fs, path string, file *PlanFile) error {
	f, err := fs.Open(path)
	if err != nil {
		return fmt.Errorf("open a file file: %w", err)
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(file); err != nil {
		return fmt.Errorf("read a plan file as JSON: %w", err)
	}
	return nil
}

type Bucket struct {
	Bucket string
	Key    string
}

func handleTerraformBlock(block *hclsyntax.Block, bucket *Bucket) (bool, error) {
	if block.Type != "terraform" {
		return false, nil
	}
	for _, backend := range block.Body.Blocks {
		if backend.Type != "backend" {
			continue
		}
		if len(backend.Labels) != 1 || backend.Labels[0] != "s3" {
			return false, nil
		}
		if key, ok := backend.Body.Attributes["key"]; ok {
			val, diag := key.Expr.Value(nil)
			if diag.HasErrors() {
				return false, diag
			}
			bucket.Key = val.AsString()
		}
		if b, ok := backend.Body.Attributes["bucket"]; ok {
			val, diag := b.Expr.Value(nil)
			if diag.HasErrors() {
				return false, diag
			}
			bucket.Bucket = val.AsString()
		}
		return true, nil
	}
	return false, nil
}

func extractBackend(src []byte, filePath string, bucket *Bucket) (bool, error) {
	file, diags := hclsyntax.ParseConfig(src, filePath, hcl.Pos{Byte: 0, Line: 1, Column: 1})
	if diags.HasErrors() {
		return false, diags
	}
	body, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		return false, errors.New("convert file body to body type")
	}
	for _, block := range body.Blocks {
		if f, err := handleTerraformBlock(block, bucket); err != nil || f {
			return f, err
		}
	}
	return false, nil
}

type RemoteState struct {
	Name string
	File string
}

func handleDataBlock(logE *logrus.Entry, block *hclsyntax.Block) (*Bucket, error) { //nolint:cyclop
	if block.Type != "data" {
		return nil, nil //nolint:nilnil
	}
	if len(block.Labels) != 2 || block.Labels[0] != "terraform_remote_state" {
		return nil, nil //nolint:nilnil
	}
	logE.Info("terraform_remote_state is found")
	backendAttr, ok := block.Body.Attributes["backend"]
	if !ok {
		logE.Warn("backend attribute is not found")
		return nil, nil //nolint:nilnil
	}
	val, diag := backendAttr.Expr.Value(nil)
	if diag.HasErrors() {
		return nil, diag
	}
	backendType := val.AsString()
	if backendType != "s3" {
		logE.Debug("backend type isn't s3")
		return nil, nil //nolint:nilnil
	}
	bucket := &Bucket{}
	configAttr, ok := block.Body.Attributes["config"]
	if !ok {
		logE.Warn("config attribute is not found")
		return nil, nil //nolint:nilnil
	}
	logE.Debug("config attribute is found")

	configVal, diag := configAttr.Expr.Value(nil)
	if diag.HasErrors() {
		return nil, diag
	}

	sv := ctyjson.SimpleJSONValue{Value: configVal}
	b, err := json.Marshal(sv)
	if err != nil {
		return nil, fmt.Errorf("marshal config attribute as JSON: %w", err)
	}
	if err := json.Unmarshal(b, bucket); err != nil {
		return nil, fmt.Errorf("unmarshal config attribute as JSON: %w", err)
	}
	return bucket, nil
}

func extractRemoteStates(logE *logrus.Entry, src []byte, filePath string, backend *Bucket) ([]*RemoteState, error) {
	file, diags := hclsyntax.ParseConfig(src, filePath, hcl.Pos{Byte: 0, Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, diags
	}
	body, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		return nil, errors.New("convert file body to body type")
	}
	states := []*RemoteState{}
	for _, block := range body.Blocks {
		bucket, err := handleDataBlock(logE, block)
		if err != nil {
			return nil, err
		}
		if bucket == nil {
			continue
		}
		name := block.Labels[1]
		if bucket.Key != backend.Key || bucket.Bucket != backend.Bucket {
			break
		}
		states = append(states, &RemoteState{
			Name: name,
			File: filePath,
		})
		break
	}
	return states, nil
}
