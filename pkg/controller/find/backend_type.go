package find

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/sirupsen/logrus"
)

const (
	backendTypeGCS = "gcs"
	backendTypeS3  = "s3"
)

type Bucket struct {
	Type   string `json:"type"`
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
	Prefix string `json:"prefix"`
}

func (b *Bucket) LogE() logrus.Fields {
	fields := logrus.Fields{
		"type": b.Type,
	}
	if b.Bucket != "" {
		fields["bucket"] = b.Bucket
	}
	if b.Key != "" {
		fields["key"] = b.Key
	}
	if b.Prefix != "" {
		fields["prefix"] = b.Prefix
	}
	return fields
}

func (b *Bucket) Compare(bucket *Bucket) bool {
	return b.Type == bucket.Type && b.Bucket == bucket.Bucket && b.Key == bucket.Key && b.Prefix == bucket.Prefix
}

func (b *Bucket) Copy(bucket *Bucket) {
	bucket.Type = b.Type
	bucket.Bucket = b.Bucket
	bucket.Key = b.Key
	bucket.Prefix = b.Prefix
}

// BackendJSON represents the structure of a Terraform backend defined in a JSON file.
//
//	{
//	    "terraform": {
//	        "backend": {
//	            "s3": {
//	                "bucket": "",
//	                "key": ""
//	            }
//	        }
//	    }
//	}
type BackendJSON struct {
	Terraform struct {
		Backend struct {
			S3  *Bucket `json:"s3"`
			GCS *Bucket `json:"gcs"`
		} `json:"backend"`
	} `json:"terraform"`
}

func extractBackendFromJSON(src []byte, bucket *Bucket) (bool, error) {
	bj := &BackendJSON{}
	if err := json.Unmarshal(src, bj); err != nil {
		return false, fmt.Errorf("unmarshal *.tf.json as JSON: %w", err)
	}
	if bj.Terraform.Backend.S3 != nil {
		bj.Terraform.Backend.S3.Type = backendTypeS3
		bj.Terraform.Backend.S3.Copy(bucket)
		return true, nil
	}
	if bj.Terraform.Backend.GCS != nil {
		bj.Terraform.Backend.S3.Type = backendTypeGCS
		bj.Terraform.Backend.S3.Copy(bucket)
		return true, nil
	}
	return false, nil
}

func getHandlers() map[string]handleBackend {
	return map[string]handleBackend{
		backendTypeS3:  handleS3Backend,
		backendTypeGCS: handleGCSBackend,
	}
}

func handleS3Backend(backend *hclsyntax.Block, bucket *Bucket) error {
	bucket.Type = backendTypeS3
	if key, ok := backend.Body.Attributes["key"]; ok {
		val, diag := key.Expr.Value(nil)
		if diag.HasErrors() {
			return diag
		}
		bucket.Key = val.AsString()
	}
	if b, ok := backend.Body.Attributes["bucket"]; ok {
		val, diag := b.Expr.Value(nil)
		if diag.HasErrors() {
			return diag
		}
		bucket.Bucket = val.AsString()
	}
	return nil
}

func handleGCSBackend(backend *hclsyntax.Block, bucket *Bucket) error {
	/*
		terraform {
		  backend "gcs" {
		    bucket  = "tf-state-prod"
		    prefix  = "terraform/state"
		  }
		}
	*/
	bucket.Type = backendTypeGCS
	if prefix, ok := backend.Body.Attributes["prefix"]; ok {
		val, diag := prefix.Expr.Value(nil)
		if diag.HasErrors() {
			return diag
		}
		bucket.Prefix = val.AsString()
	}
	if b, ok := backend.Body.Attributes["bucket"]; ok {
		val, diag := b.Expr.Value(nil)
		if diag.HasErrors() {
			return diag
		}
		bucket.Bucket = val.AsString()
	}
	return nil
}
