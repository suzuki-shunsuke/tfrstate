package find

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/sirupsen/logrus"
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

func getHandlers() map[string]handleBackend {
	return map[string]handleBackend{
		"s3":  handleS3Backend,
		"gcs": handleGCSBackend,
	}
}

func handleS3Backend(backend *hclsyntax.Block, bucket *Bucket) error {
	bucket.Type = "s3"
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
	bucket.Type = "gcs"
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
