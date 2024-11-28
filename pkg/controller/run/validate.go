package run

import (
	"errors"
)

func validateParam(param *Param) error {
	if param.PlanFile == "" {
		if param.Bucket == "" || param.Key == "" {
			return errors.New("plan-json or s3-bucket and s3-key must be set")
		}
	} else {
		if param.Bucket != "" {
			return errors.New("plan-json and s3-bucket can't be used at the same time")
		}
		if param.Key != "" {
			return errors.New("plan-json and s3-key can't be used at the same time")
		}
	}
	return nil
}
