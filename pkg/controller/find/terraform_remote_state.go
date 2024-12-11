package find

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/sirupsen/logrus"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

func extractRemoteStates(logE *logrus.Entry, src []byte, filePath string, backend *Bucket) ([]*RemoteState, error) {
	// Extract terraform_remote_state data sources matching with a given backend from a file.
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
		if !bucket.Compare(backend) {
			continue
		}
		states = append(states, &RemoteState{
			Name: name,
			File: filePath,
		})
		break
	}
	return states, nil
}

func extractRemoteStatesFromJSON(src []byte, filePath string) ([]*RemoteState, error) {
	file := struct {
		Data struct {
			TerraformRemoteState map[string]struct {
				Backend string `json:"backend"`
				Config  Bucket `json:"config"`
			} `json:"terraform_remote_state"`
		} `json:"data"`
	}{}
	if err := json.Unmarshal(src, &file); err != nil {
		return nil, fmt.Errorf("unmarshal JSON: %w", err)
	}
	if len(file.Data.TerraformRemoteState) == 0 {
		return nil, nil
	}
	states := []*RemoteState{}
	for name := range file.Data.TerraformRemoteState {
		states = append(states, &RemoteState{
			Name: name,
			File: filePath,
		})
	}
	return states, nil
}

func handleDataBlock(logE *logrus.Entry, block *hclsyntax.Block) (*Bucket, error) {
	/*
		data "terraform_remote_state" "vpc" {
		  backend = "s3"
		  config = {
		    bucket = "terraform-state-XXXXXXXXXXXX"
		    key    = "production/vpc/terraform.tfstate"
		  }
		}
	*/
	if block.Type != "data" {
		return nil, nil //nolint:nilnil
	}
	if len(block.Labels) != 2 || block.Labels[0] != "terraform_remote_state" {
		return nil, nil //nolint:nilnil
	}
	logE.Debug("terraform_remote_state is found")
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
	bucket.Type = backendType
	return bucket, nil
}
