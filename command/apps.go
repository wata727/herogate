package command

import (
	"github.com/urfave/cli"
	"github.com/wata727/herogate/herogate"
)

// AppsCreateCommand is command for creating a new app.
func AppsCreateCommand() cli.Command {
	return cli.Command{
		Name:      "apps:create",
		ShortName: "create",
		Usage:     "creates a new app",
		Action:    herogate.AppsCreate,
	}
}
