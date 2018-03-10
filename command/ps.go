package command

import (
	"github.com/urfave/cli"
	"github.com/wata727/herogate/herogate"
)

// PsCommand is a command for listing containers.
func PsCommand() cli.Command {
	return cli.Command{
		Name:   "ps",
		Usage:  "list containers for an app",
		Action: herogate.Ps,
	}
}
