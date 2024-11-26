package cli

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/tf-remote-state-find/pkg/controller/run"
	"github.com/suzuki-shunsuke/tf-remote-state-find/pkg/log"
	"github.com/urfave/cli/v2"
)

type runCommand struct {
	logE *logrus.Entry
}

func (rc *runCommand) command() *cli.Command {
	return &cli.Command{
		Name:   "run",
		Usage:  "Parse Terraform Plan file and list directories where changed outputs are used",
		Action: rc.action,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "plan-json",
			},
			&cli.StringFlag{
				Name: "root-dir",
			},
			&cli.StringFlag{
				Name: "backend-dir",
			},
			&cli.StringFlag{
				Name: "s3-bucket",
			},
			&cli.StringFlag{
				Name: "s3-key",
			},
			&cli.StringSliceFlag{
				Name:    "output",
				Aliases: []string{"o"},
			},
		},
	}
}

func (rc *runCommand) action(c *cli.Context) error {
	fs := afero.NewOsFs()
	logE := rc.logE
	log.SetLevel(c.String("log-level"), logE)
	log.SetColor(c.String("log-color"), logE)
	return run.Run(c.Context, logE, fs, &run.Param{ //nolint:wrapcheck
		PlanFile: c.String("plan-json"),
		Root:     c.String("root-dir"),
		Dir:      c.String("backend-dir"),
		Key:      c.String("s3-key"),
		Bucket:   c.String("s3-bucket"),
		Outputs:  c.StringSlice("output"),
	})
}
