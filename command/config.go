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

// ConfigSetCommand is a command for setting environment variables.
func ConfigSetCommand() cli.Command {
	return cli.Command{
		Name:   "config:set",
		Usage:  "set one or more config vars",
		Flags:  sharedFlags(),
		Action: herogate.ConfigSet,
	}
}

// ConfigUnsetCommand is a command for unsetting environment variables.
func ConfigUnsetCommand() cli.Command {
	return cli.Command{
		Name:   "config:unset",
		Usage:  "unset one or more config vars",
		Flags:  sharedFlags(),
		Action: herogate.ConfigUnset,
	}
}
