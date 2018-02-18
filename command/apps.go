package command

import (
	"github.com/urfave/cli"
	"github.com/wata727/herogate/herogate"
)

// AppsCreateCommand is a command for creating a new app.
func AppsCreateCommand() cli.Command {
	return cli.Command{
		Name:      "apps:create",
		ShortName: "create",
		Usage:     "creates a new app",
		Action:    herogate.AppsCreate,
	}
}

// AppsOpenCommand is a command for opening the app in a web browser.
func AppsOpenCommand() cli.Command {
	return cli.Command{
		Name:      "apps:open",
		ShortName: "open",
		Usage:     "open the app in a web browser",
		Flags:     sharedFlags(),
		Action:    herogate.AppsOpen,
	}
}
