package find

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/spf13/afero"
)

func findBackendConfig(afs afero.Fs, dir string, bucket *Bucket) error {
	// parse HCLs in dir and extract backend configurations
	matchFiles, err := afero.Glob(afs, filepath.Join(dir, "*.tf"))
	if err != nil {
		return fmt.Errorf("glob *.tf to get Backend configuration: %w", err)
	}
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
	return nil
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

func handleTerraformBlock(block *hclsyntax.Block, bucket *Bucket) (bool, error) {
	/*
		terraform {
		  backend "s3" {
		    bucket = "terraform-state-prod"
		    key    = "network/terraform.tfstate"
		  }
		}
	*/
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