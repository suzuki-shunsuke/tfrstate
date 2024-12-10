package cli

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/tfrstate/pkg/controller/find"
	"github.com/suzuki-shunsuke/tfrstate/pkg/log"
	"github.com/urfave/cli/v2"
)

type findCommand struct {
	logE *logrus.Entry
}

func (rc *findCommand) command() *cli.Command {
	return &cli.Command{
		Name:   "find",
		Usage:  "Find directories where a given terraform_remote_state data source is used",
		Action: rc.action,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "output-format",
				Usage: "Output format. One of 'json' (default), 'markdown'",
				Value: "json",
			},
			&cli.StringFlag{
				Name:  "plan-json",
				Usage: "The file path to the plan file in JSON format",
			},
			&cli.StringFlag{
				Name:  "base-dir",
				Usage: "The file path to the directory where Terraform configuration files are located",
			},
			&cli.StringFlag{
				Name:  "backend-dir",
				Usage: "The file path to the given Terraform Root Module",
			},
			&cli.StringFlag{
				Name:  "s3-bucket",
				Usage: "S3 Bucket Name of terraform_remote_state data source",
			},
			&cli.StringFlag{
				Name:  "s3-key",
				Usage: "S3 Bucket Key of terraform_remote_state data source",
			},
			&cli.StringFlag{
				Name:  "gcs-bucket",
				Usage: "GCS Bucket Name of terraform_remote_state data source",
			},
			&cli.StringFlag{
				Name:  "gcs-prefix",
				Usage: "GCS Bucket Prefix of terraform_remote_state data source",
			},
			&cli.StringSliceFlag{
				Name:    "output",
				Usage:   "Output name of terraform_remote_state data source",
				Aliases: []string{"o"},
			},
		},
	}
}

func (rc *findCommand) action(c *cli.Context) error {
	fs := afero.NewOsFs()
	logE := rc.logE
	log.SetLevel(c.String("log-level"), logE)
	log.SetColor(c.String("log-color"), logE)
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get the current directory: %w", err)
	}
	return find.Find(c.Context, logE, fs, &find.Param{ //nolint:wrapcheck
		Format:    c.String("output-format"),
		PlanFile:  c.String("plan-json"),
		Root:      c.String("base-dir"),
		Dir:       c.String("backend-dir"),
		Key:       c.String("s3-key"),
		Bucket:    c.String("s3-bucket"),
		GCSPrefix: c.String("gcs-prefix"),
		GCSBucket: c.String("gcs-bucket"),
		Outputs:   c.StringSlice("output"),
		Stdout:    os.Stdout,
		PWD:       pwd,
	})
}
