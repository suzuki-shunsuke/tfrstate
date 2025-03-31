package cli

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/urfave/cli/v3"
)

type versionCommand struct {
	stdout  io.Writer
	version string
	commit  string
}

func (vc *versionCommand) command() *cli.Command {
	return &cli.Command{
		Name:   "version",
		Usage:  "Show version",
		Action: vc.action,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name: "json",
			},
		},
	}
}

func (vc *versionCommand) action(c *cli.Context) error {
	if c.Bool("json") {
		encoder := json.NewEncoder(vc.stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(map[string]string{
			"version": vc.version,
			"commit":  vc.commit,
		}); err != nil {
			return fmt.Errorf("encode the version as JSON: %w", err)
		}
		return nil
	}
	cli.ShowVersion(c)
	return nil
}
