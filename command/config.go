package command

import (
	"github.com/urfave/cli"
	"github.com/wata727/herogate/herogate"
)

// ConfigCommand is a command for listing environment variables.
func ConfigCommand() cli.Command {
	return cli.Command{
		Name:   "config",
		Usage:  "display the config vars for an app",
		Flags:  sharedFlags(),
		Action: herogate.Config,
	}
}

// ConfigGetCommand is a command for getting an environment variable.
func ConfigGetCommand() cli.Command {
	return cli.Command{
		Name:   "config:get",
		Usage:  "display a config value for an app",
		Flags:  sharedFlags(),
		Action: herogate.ConfigGet,
	}
}
