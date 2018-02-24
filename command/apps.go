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

// AppsDestroyCommand is a command for detroying the app.
func AppsDestroyCommand() cli.Command {
	return cli.Command{
		Name:      "apps:destroy",
		ShortName: "destroy",
		Usage:     "permanently destroy an app",
		Flags:     append(sharedFlags(), appsDestroyFlags()...),
		Action:    herogate.AppsDestroy,
	}
}

func appsDestroyFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "confirm",
			Usage: "destroy an app without the app name re-typing",
		},
	}
}
