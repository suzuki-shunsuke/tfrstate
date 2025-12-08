package cli

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/slog-util/slogutil"
	"github.com/suzuki-shunsuke/tfrstate/pkg/controller/find"
	"github.com/urfave/cli/v3"
)

type GlobalArgs struct {
	LogLevel string
}

type FindArgs struct {
	*GlobalArgs

	OutputFormat string
	PlanFile     string
	BaseDir      string
	BackendDir   string
	S3Bucket     string
	S3Key        string
	GCSBucket    string
	GCSPrefix    string
	Outputs      []string
}

type findCommand struct {
	Stdout io.Writer
}

func (rc *findCommand) command(logger *slogutil.Logger, globalArgs *GlobalArgs) *cli.Command {
	args := &FindArgs{
		GlobalArgs: globalArgs,
	}
	return &cli.Command{
		Name:  "find",
		Usage: "Find directories where a given terraform_remote_state data source is used",
		Action: func(ctx context.Context, _ *cli.Command) error {
			return rc.action(ctx, logger, args)
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "output-format",
				Usage:       "Output format. One of 'json' (default), 'markdown'",
				Value:       "json",
				Destination: &args.OutputFormat,
			},
			&cli.StringFlag{
				Name:        "plan-json",
				Usage:       "The file path to the plan file in JSON format",
				Destination: &args.PlanFile,
			},
			&cli.StringFlag{
				Name:        "base-dir",
				Usage:       "The file path to the directory where Terraform configuration files are located",
				Destination: &args.BaseDir,
			},
			&cli.StringFlag{
				Name:        "backend-dir",
				Usage:       "The file path to the given Terraform Root Module",
				Destination: &args.BackendDir,
			},
			&cli.StringFlag{
				Name:        "s3-bucket",
				Usage:       "S3 Bucket Name of terraform_remote_state data source",
				Destination: &args.S3Bucket,
			},
			&cli.StringFlag{
				Name:        "s3-key",
				Usage:       "S3 Bucket Key of terraform_remote_state data source",
				Destination: &args.S3Key,
			},
			&cli.StringFlag{
				Name:        "gcs-bucket",
				Usage:       "GCS Bucket Name of terraform_remote_state data source",
				Destination: &args.GCSBucket,
			},
			&cli.StringFlag{
				Name:        "gcs-prefix",
				Usage:       "GCS Bucket Prefix of terraform_remote_state data source",
				Destination: &args.GCSPrefix,
			},
			&cli.StringSliceFlag{
				Name:        "output",
				Usage:       "Output name of terraform_remote_state data source",
				Aliases:     []string{"o"},
				Destination: &args.Outputs,
			},
		},
	}
}

func (rc *findCommand) action(ctx context.Context, logger *slogutil.Logger, args *FindArgs) error {
	fs := afero.NewOsFs()
	if err := logger.SetLevel(args.LogLevel); err != nil {
		return fmt.Errorf("set log level: %w", err)
	}
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get the current directory: %w", err)
	}
	return find.Find(ctx, logger.Logger, fs, &find.Param{ //nolint:wrapcheck
		Format:    args.OutputFormat,
		PlanFile:  args.PlanFile,
		Root:      args.BaseDir,
		Dir:       args.BackendDir,
		Key:       args.S3Key,
		Bucket:    args.S3Bucket,
		GCSPrefix: args.GCSPrefix,
		GCSBucket: args.GCSBucket,
		Outputs:   args.Outputs,
		Stdout:    rc.Stdout,
		PWD:       pwd,
	})
}
