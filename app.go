package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/wata727/herogate/command"
)

// NewApp return new application
func NewApp() *cli.App {
	if os.Getenv("HEROGATE_DEBUG") == "1" || os.Getenv("HEROGATE_DEBUG") == "true" {
		logrus.SetLevel(logrus.DebugLevel)
	}

	app := cli.NewApp()
	app.Name = Name
	app.Usage = "Deploy and manage containerized applications like Heroku on AWS"
	app.Version = Version

	app.Commands = []cli.Command{
		command.AppsCommand(),
		command.AppsCreateCommand(),
		command.AppsInfoCommand(),
		command.AppsOpenCommand(),
		command.AppsDestroyCommand(),
		command.ConfigCommand(),
		command.ConfigGetCommand(),
		command.ConfigSetCommand(),
		command.ConfigUnsetCommand(),
		command.LogsCommand(),
		command.InternalCommand(),
	}

	return app
}
