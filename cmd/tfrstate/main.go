package main

import (
	"github.com/suzuki-shunsuke/tfrstate/pkg/cli"
	"github.com/suzuki-shunsuke/urfave-cli-v3-util/urfave"
)

var version = ""

func main() {
	urfave.Main("tfrstate", version, cli.Run)
}
