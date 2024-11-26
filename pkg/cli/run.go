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
		Flags:  []cli.Flag{},
	}
}

func (rc *runCommand) action(c *cli.Context) error {
	fs := afero.NewOsFs()
	logE := rc.logE
	log.SetLevel(c.String("log-level"), logE)
	log.SetColor(c.String("log-color"), logE)
	return run.Run(c.Context, logE, fs, &run.Param{ //nolint:wrapcheck
		PlanFile: c.Args().First(),
		Dir:      c.Args().Get(2), //nolint:mnd
	})
}
