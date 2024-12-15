package find

import (
	"strings"
)

func findCaller(dirs map[string]*Dir, changedOutputs []string, changed map[string]map[string]map[string]struct{}) {
	// Find files referring terraform_remote_state
	for _, dir := range dirs {
		for _, file := range dir.Files {
			if !strings.Contains(file.Content, "data.terraform_remote_state.") {
				continue
			}
			for _, state := range dir.States {
				findCallerCore(dir, file, state, changedOutputs, changed)
			}
		}
	}
}

func findCallerCore(dir *Dir, file *File, state *RemoteState, changedOutputs []string, changed map[string]map[string]map[string]struct{}) {
	if !strings.Contains(file.Content, "data.terraform_remote_state."+state.Name+".outputs.") {
		return
	}
	if len(changedOutputs) == 0 {
		m, ok := changed[dir.Path]
		if !ok {
			m = map[string]map[string]struct{}{}
		}
		m2, ok := m[file.Path]
		if !ok {
			m2 = map[string]struct{}{}
		}
		m[file.Path] = m2
		changed[dir.Path] = m
		return
	}
	for _, outputName := range changedOutputs {
		if !strings.Contains(file.Content, "data.terraform_remote_state."+state.Name+".outputs."+outputName) {
			continue
		}
		m, ok := changed[dir.Path]
		if !ok {
			m = map[string]map[string]struct{}{}
		}
		m2, ok := m[file.Path]
		if !ok {
			m2 = map[string]struct{}{}
		}
		m2[outputName] = struct{}{}
		m[file.Path] = m2
		changed[dir.Path] = m
	}
}
