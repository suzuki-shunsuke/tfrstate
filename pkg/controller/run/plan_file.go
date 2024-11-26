package run

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"

	"github.com/spf13/afero"
)

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

func extractChangedOutputs(afs afero.Fs, path string) ([]string, error) {
	planFile := &PlanFile{}
	if err := readPlanFile(afs, path, planFile); err != nil {
		return nil, fmt.Errorf("read a plan file: %w", err)
	}
	excludeCreatedOutputs(planFile)
	return slices.Sorted(maps.Keys(planFile.OutputChanges)), nil
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
