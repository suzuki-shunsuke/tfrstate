package cli

import (
	"context"

	"github.com/suzuki-shunsuke/slog-util/slogutil"
	"github.com/suzuki-shunsuke/urfave-cli-v3-util/urfave"
	"github.com/urfave/cli/v3"
)

func Run(ctx context.Context, logger *slogutil.Logger, env *urfave.Env) error {
	globalArgs := &GlobalArgs{}
	return urfave.Command(env, &cli.Command{ //nolint:wrapcheck
		Name:  "tfrstate",
		Usage: "Find directories where a given terraform_remote_state data source is used",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "log-level",
				Usage:       "log level",
				Destination: &globalArgs.LogLevel,
			},
		},
		Commands: []*cli.Command{
			(&findCommand{
				Stdout: env.Stdout,
			}).command(logger, globalArgs),
		},
	}).Run(ctx, env.Args)
}
